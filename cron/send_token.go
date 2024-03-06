package cron

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/savioxavier/termlink"
	tele "gopkg.in/telebot.v3"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"goad/database"
	"goad/pkg/airdrop"
	"goad/pkg/blockchain"
	"goad/pkg/encryption"
	"goad/pkg/wallet"
	"goad/pkg/yaml"
	"gorm.io/gorm"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

var sendMu sync.Mutex

const Password = "ItsSomethingThatOnlyICanImagine!:)G0051ock;Ifyou'reseeingthisprobablyit'sbecauseIwroteitin5m"

func SendToken(token *yaml.TokenConfig) {
	cColors := color.New(color.FgHiMagenta)
	tColors := color.New(color.FgGreen)
	bColors := color.New(color.FgYellow)
	if !common.IsHexAddress(token.Address) {
		log.Fatalf("Invalid token address: %v", token.Address)
	}
	//for !token.Active {
	//	time.Sleep(1 * time.Second)
	//}
	contract := client.NewContract(common.HexToAddress(token.Address))
	name, err := contract.GetTokenName()
	if err != nil {
		log.Fatalf("Error getting token name of address: %v", token.Address)
		return
	}
	symbol, err := contract.GetTokenSymbol()
	if err != nil {
		log.Fatalf("Error getting token symbol of address: %v", token.Address)
		return
	}
	currentPrice, err := contract.GetPrice()
	if err != nil {
		log.Fatalf("Error getting token price of address: %v, %v", token.Address, err)
		return
	}
	log.Infof("[%v] Token (%s) Loaded, current price: $%.6f", cColors.Sprint(name), cColors.Sprint(symbol), currentPrice)
	counter := 1
	var accounts []*wallet.Account
	for _, privateKey := range token.Wallets {
		var w *wallet.Account
		key := privateKey
		key, err = encryption.DecryptHex(privateKey, Password)
		if err != nil {
			log.Errorf("[%v] Wallet failed to load: Private(%v), %v", cColors.Sprint(name), privateKey, err)
			continue
		}
		w, err = wallet.FromPrivateKey(key)
		if err != nil {
			if len(privateKey) > 10 {
				privateKey = fmt.Sprintf("%v...%v", privateKey[:4], privateKey[len(privateKey)-4:])
			}
			log.Errorf("[%v] Wallet failed to load: Private(%v), %v", cColors.Sprint(name), privateKey, err)
			continue
		}
		var tokenBalance *big.Int
		tokenBalance, err = contract.GetTokenBalance(w.Address())
		if err != nil {
			log.Errorf("Wallet (%v) failed to get %v token balance, %v", w.AddressMask(), cColors.Sprint(name), err)
			continue
		}
		w.SetTokenBalance(symbol, tokenBalance)
		w.BNBBalance, err = contract.GetBalance(w.Address())
		if err != nil {
			log.Errorf("Wallet (%v) failed to get %v coin balance, %v", w.AddressMask(), cColors.Sprint(name), err)
			continue
		}
		tBalance := tColors.Sprintf("%s", blockchain.BigFloatToString(blockchain.BigIntToBigFloat(tokenBalance, 18)))
		cBalance := bColors.Sprintf("%.4f BNB", blockchain.BigIntToAmount(w.BNBBalance, 18))
		log.Infof("[%v] Wallet loaded: %v, balance: %s %v, %s", cColors.Sprint(name), w.AddressMask(), tBalance, cColors.Sprint(symbol), cBalance)
		accounts = append(accounts, w)
	}
	var sleep time.Duration
	sleep = 0
	log.Infof("[%v] Sending transaction job started with %d wallets", cColors.Sprint(name), len(accounts))

	for {
		time.Sleep(sleep * time.Second)
		sleep, err = token.GetSleep()
		if err != nil {
			log.Errorf("Error getting token sleep time of address: %v, %v setting sleep to 180 seconds", token.Address, err)
			sleep = 180
		}
		var wal *database.Wallet
		sendMu.Lock()
		wal, err = database.GetOneWalletWithoutTX()
		if err != nil {
			sendMu.Unlock()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Errorf("Wallet storage is empty wait for 30 second...")
				sleep = 30
				continue
			}
			log.Errorf("could not get wallet, %v", err)
			continue
		}
		if !common.IsHexAddress(wal.Address) {
			log.Errorf("Invalid wallet address format %v", wal.Address)
			wal.Disable()
			sendMu.Unlock()
			continue
		}
		var balance *big.Int
		balance, err = client.GetBalance(common.HexToAddress(wal.Address))
		if err != nil {
			log.Errorf("Could not get ballance of wallet %v, %s, wait for 180 seconds", wal.Address, err)
			wal.Disable()
			sendMu.Unlock()
			sleep = 180
			continue
		}
		bnbBalance := blockchain.BigIntToAmount(balance, 18)
		if bnbBalance == 0 {
			log.Warnf("Wallet (%v) balance is 0: %v", wal.Address, bnbBalance)
			wal.Disable()
			sendMu.Unlock()
			continue
		}
		err = wal.SetBalance(bnbBalance)
		if err != nil {
			log.Errorf("Could not set wallet BNB Balance: %v, balance: %v, %s", wal.Address, bnbBalance, err)
			sendMu.Unlock()
			continue
		}
		var airdropAmount float64
		airdropAmount, err = wal.GenerateAmount()
		if err != nil {
			log.Errorf("Could not generate amount Balance: %v, balance: %v, %s", wal.Address, bnbBalance, err)
			sendMu.Unlock()
			continue
		}
		airdropAmount *= token.Ratio
		currentPrice, err = contract.GetPrice()
		if err != nil {
			log.Errorf("Error getting token price of address: %v, %v wait for 180 seconds", token.Address, err)
			sendMu.Unlock()
			sleep = 180
			continue
		}
		tokenAmount := airdropAmount / currentPrice
		min := 0.8
		max := 1.2
		randomRange := min + rand.Float64()*(max-min)
		airdropAmount = randomRange * airdropAmount
		tokenAmount = airdrop.RandomizeDecimalCount(randomRange * airdropAmount / currentPrice)
		var transCost *big.Int
		transCost, err = contract.CalculateCost(common.HexToAddress(token.Address))
		if err != nil {
			sendMu.Unlock()
			log.Errorf("Error getting preTransaction")
			continue
		}
		tokenWallet := getRandomWalletWithBalance(accounts, name, symbol, blockchain.AmountToBigInt(tokenAmount, 18), transCost)
		if tokenWallet == nil {
			sendMu.Unlock()
			log.Errorf("[%s] All wallets are empty. wait 600 seconds", cColors.Sprint(name))
			bot.Send(&tele.Chat{
				ID: database.Config.TelegramBot.AnnounceChannel,
			}, fmt.Sprintf("[%s] All wallets are empty", cColors.Sprint(name)))
			sleep = 600
			continue
		}
		amountBigInt := blockchain.AmountToBigInt(tokenAmount, 18)
		amountBigFloat := blockchain.BigIntToBigFloat(amountBigInt, 18)
		AmountString := blockchain.BigFloatToString(amountBigFloat)
		var transferToken *types.Transaction
		transferToken, err = contract.TransferToken(tokenWallet, common.HexToAddress(wal.Address), tokenAmount)
		if err != nil {
			sendMu.Unlock()
			log.Errorf("[%s] Wallet (%v) failed to send %s to address %s", cColors.Sprint(name), tokenWallet.AddressMask(), tColors.Sprintf("%s %s", AmountString, wal.Address), err)
			continue
		}
		err = wal.AddTX(symbol, transferToken.Hash().String(), tokenAmount)
		if err != nil {
			log.Fatalf("[%s] Wallet (%v) failed to add TX \"%s\" with amount %s, %s", cColors.Sprint(name), tokenWallet.AddressMask(), transferToken.Hash().String(), tColors.Sprintf("%s %s", AmountString, wal.Address), err)
			sendMu.Unlock()
			continue
		}
		tx := transferToken.Hash().String()
		//tx := "non"

		var tokenBalance *big.Int
		tokenBalance, err = contract.GetTokenBalance(tokenWallet.Address())
		if err != nil {
			sendMu.Unlock()
			log.Errorf("Wallet (%v) failed to get %v token balance, %v", tokenWallet.AddressMask(), cColors.Sprint(symbol), err)
			continue
		}
		tokenWallet.SetTokenBalance(symbol, tokenBalance)
		tokenWallet.BNBBalance, err = contract.GetBalance(tokenWallet.Address())
		if err != nil {
			sendMu.Unlock()
			log.Errorf("Wallet (%v) failed to get %v coin balance, %v", tokenWallet.AddressMask(), cColors.Sprint(symbol), err)
			continue
		}
		link := termlink.Link(wal.AddressMask(), fmt.Sprintf("https://bscscan.com/tx/%s", tx))
		link = color.New(color.BgBlack).Add(color.FgWhite).Add(color.Bold).Sprintf(link)
		tBalance := tColors.Sprintf("%s", blockchain.BigFloatToString(blockchain.BigIntToBigFloat(tokenBalance, 18)))
		cBalance := bColors.Sprintf("%.4f BNB", blockchain.BigIntToAmount(tokenWallet.BNBBalance, 18))
		log.Infof("[%s] %d. %s sent to %s from %s remaining %s %v, %s, next in %ds", cColors.Sprint(name), counter, tColors.Sprintf("$%.2f", airdropAmount), link, tokenWallet.AddressMask(), tBalance, cColors.Sprint(symbol), cBalance, sleep)
		counter++
		sendMu.Unlock()
		log.Debugf("Sleep for %d seconds", sleep)
	}
}

var lastMessage map[common.Address]*time.Time

const duration time.Duration = 5 * 60 * time.Second

func getRandomWalletWithBalance(wallets []*wallet.Account, name, symbol string, amount *big.Int, cost *big.Int) *wallet.Account {
	if lastMessage == nil {
		lastMessage = make(map[common.Address]*time.Time)
	}
	annID := database.Config.TelegramBot.AnnounceChannel
	now := time.Now()
	copiedSlice := make([]*wallet.Account, len(wallets))
	for i, account := range wallets {
		copiedAccount := *account
		copiedSlice[i] = &copiedAccount
	}
	rand.Shuffle(len(copiedSlice), func(i, j int) {
		copiedSlice[i], copiedSlice[j] = copiedSlice[j], copiedSlice[i]
	})
	var account *wallet.Account
	for _, acc := range copiedSlice {
		if acc.TokenBalance(symbol).Cmp(amount) >= 0 {
			if acc.BNBBalance.Cmp(cost) >= 0 {
				account = acc
				continue
			}
			bnbAmount := color.YellowString("%.4f BNB", blockchain.BigIntToAmount(acc.BNBBalance, 18))
			tAmount := fmt.Sprintf("%.4f %s", blockchain.BigIntToAmount(acc.TokenBalance(symbol), 18), symbol)
			log.Warnf("Wallet %s ran out of BNB, Currnet Balance: %v", acc.AddressMask(), bnbAmount)
			if _, ok := lastMessage[acc.Address()]; !ok || lastMessage[acc.Address()].Add(duration).Before(now) {
				teleText := fmt.Sprintf("%s\n\nWallet %s ran out of BNB\n\nCurrnet Balance:\n%.4f BNB\n%s\n\nWallet: %s", name, acc.AddressMask(true), blockchain.BigIntToAmount(acc.BNBBalance, 18), tAmount, acc.Address().String())
				_, err := bot.Send(&tele.Chat{ID: annID}, teleText)
				if err != nil {
					log.Errorf("error sending telegram message: %v", err)
				}
				lastMessage[acc.Address()] = &now
			}
			continue
		}
		bnbAmount := fmt.Sprintf("%.4f BNB", blockchain.BigIntToAmount(acc.BNBBalance, 18))
		tAmount := fmt.Sprintf("%.4f %s", blockchain.BigIntToAmount(acc.TokenBalance(symbol), 18), symbol)
		log.Warnf("Wallet %s ran out of %s, Currnet Balance: %s", acc.AddressMask(), symbol, tAmount)
		if _, ok := lastMessage[acc.Address()]; !ok || lastMessage[acc.Address()].Add(duration).Before(now) {
			teleText := fmt.Sprintf("%s\n\nWallet %s ran out of %s\n\nCurrnet Balance:\n%s\n%s\n\nWallet: %s", name, acc.AddressMask(true), symbol, bnbAmount, tAmount, acc.Address().String())
			_, err := bot.Send(&tele.Chat{ID: annID}, teleText)
			if err != nil {
				log.Errorf("error sending telegram message: %v", err)
			}
			lastMessage[acc.Address()] = &now
		}
	}
	return account
}
