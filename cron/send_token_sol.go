package cron

import (
	"errors"
	"fmt"
	common2 "github.com/blocto/solana-go-sdk/common"
	"github.com/savioxavier/termlink"
	tele "gopkg.in/telebot.v3"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"goad/database"
	"goad/pkg/airdrop"
	"goad/pkg/blockchain"
	"goad/pkg/yaml"
	"gorm.io/gorm"
	"math/rand"
	"sync"
	"time"
)

var sendMuSol sync.Mutex

//const Password = "ItsSomethingThatOnlyICanImagine!:)G0051ock;Ifyou'reseeingthisprobablyit'sbecauseIwroteitin5m"

func SendTokenSOL(token *yaml.TokenConfig) {
	token.Counter = 1
	cColors := color.New(color.FgHiMagenta)
	tColors := color.New(color.FgGreen)
	bColors := color.New(color.FgYellow)
	if token.Chain == "bsc" {
		log.Fatalf("Invalid chain: %v", token.Address)
	} else if token.Chain == "sol" {
		if ok := blockchain.IsSolanaWallet(token.Address); !ok {
			log.Fatalf("Invalid token address: %v", token.Address)
		}
	} else {
		log.Fatalf("Invalid chain: %v", token.Chain)
	}
	//for !token.Active {
	//	time.Sleep(1 * time.Second)
	//}
	var name string
	var symbol string
	var currentPrice float64
	var err error
	tokenInfo, err := blockchain.GetTokenInfoByMintAddress(token.Address)
	if err != nil {
		return
	}
	name = tokenInfo.Data.Name
	symbol = tokenInfo.Data.Symbol
	currentPrice = token.Price
	log.Infof("[%v] Token (%s) Loaded, current price: $%.6f", cColors.Sprint(name), cColors.Sprint(symbol), currentPrice)
	var accounts []*blockchain.SolClient
	for _, privateKey := range token.Wallets {
		var w *blockchain.SolClient
		key := privateKey
		//key, err = encryption.DecryptHex(privateKey, Password)
		//if err != nil {
		//	log.Errorf("[%v] Wallet failed to load: Private(%v), %v", cColors.Sprint(name), privateKey, err)
		//	continue
		//}
		w, err = blockchain.NewSolanaWallet(key, "")
		if err != nil {
			if len(privateKey) > 10 {
				privateKey = fmt.Sprintf("%v...%v", privateKey[:4], privateKey[len(privateKey)-4:])
			}
			log.Errorf("[%v] Wallet failed to load: Private(%v), %v", cColors.Sprint(name), privateKey, err)
			continue
		}
		var tokenBalance float64
		tokenBalance, err = blockchain.GetSolTokenBalance(w.Address(), token.Address)
		if err != nil {
			log.Errorf("Wallet (%v) failed to get %v token balance, %v", w.Address(), cColors.Sprint(name), err)
			continue
		}
		w.SetTokenBalance(symbol, tokenBalance)
		w.SOLBalance, err = blockchain.GetSolBalance(w.Address())
		if err != nil {
			log.Errorf("Wallet (%v) failed to get %v coin balance, %v", w.Address(), cColors.Sprint(name), err)
			continue
		}
		tBalance := tColors.Sprintf("%.3f", tokenBalance)
		cBalance := bColors.Sprintf("%.4f Sol", w.SOLBalance)
		log.Infof("[%v] Wallet loaded: %v, balance: %s %v, %s", cColors.Sprint(name), w.Address(), tBalance, cColors.Sprint(symbol), cBalance)
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
		wal, err = database.GetOneWalletWithoutTX(database.Solana)
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
		//if !common.IsHexAddress(wal.Address) {
		//	log.Errorf("Invalid wallet address format %v", wal.Address)
		//	wal.Disable()
		//	sendMu.Unlock()
		//	continue
		//}
		//_, err = hexutil.DecodeBig(wal.Address)
		//if err != nil {
		//	sendMu.Unlock()
		//	log.Errorf("Invalid wallet address %v", wal.Address)
		//	wal.Disable()
		//	sleep = 10
		//	continue
		//}
		var balance float64
		balance, err = blockchain.GetSolBalance(wal.SolAddress())
		if err != nil {
			log.Errorf("[%s] Could not get ballance of wallet %v, %s, wait for 180 seconds", cColors.Sprint(name), wal.Address, err)
			wal.Disable()
			sendMu.Unlock()
			sleep = 180
			continue
		}
		bnbBalance := balance
		if bnbBalance == 0 {
			log.Warnf("[%s] Wallet (%v) balance is 0: %v", cColors.Sprint(name), wal.Address, bnbBalance)
			wal.Disable()
			sendMu.Unlock()
			sleep = 0
			continue
		}
		err = wal.SetBalance(bnbBalance)
		if err != nil {
			log.Errorf("[%s] Could not set wallet Sol Balance: %v, balance: %v, %s", cColors.Sprint(name), wal.Address, bnbBalance, err)
			sendMu.Unlock()
			continue
		}
		var airdropAmount float64
		airdropAmount, err = wal.GenerateAmount()
		if err != nil {
			log.Errorf("[%s] Could not generate amount Balance: %v, balance: %v, %s", cColors.Sprint(name), wal.Address, bnbBalance, err)
			sendMu.Unlock()
			continue
		}
		airdropAmount = airdropAmount * token.Ratio
		currentPrice = token.Price
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
		//var transCost *big.Int
		//transCost, err = BSCContract.CalculateCost(common.HexToAddress(token.Address))
		//if err != nil {
		//	sendMu.Unlock()
		//	log.Errorf("Error getting preTransaction")
		//	continue
		//}
		//log.Errorf("cost %v", transCost)

		tokenWallet := getRandomWalletWithBalanceSol(accounts, name, symbol, tokenAmount, 0.0002)
		if tokenWallet == nil {
			sendMu.Unlock()
			log.Errorf("[%s] All wallets are empty. wait 600 seconds", cColors.Sprint(name))
			bot.Send(&tele.Chat{ID: database.Config.TelegramBot.AnnounceChannel},
				fmt.Sprintf("❌ [%s] All wallets are empty ❌", cColors.Sprint(name)))
			err = UpdateWalletsBalanceSol(name, symbol, token.Address, accounts)

			sleep = 600
			continue
		}
		amountBigInt := blockchain.AmountToBigInt(tokenAmount, 18)
		amountBigFloat := blockchain.BigIntToBigFloat(amountBigInt, 18)
		AmountString := blockchain.BigFloatToString(amountBigFloat)
		var transferToken string
		transferToken, err = tokenWallet.SendSLPToken(token.Address, wal.Address, tokenAmount)
		if err != nil {
			sendMu.Unlock()
			log.Errorf("[%s] Wallet (%v) failed to send %s to address %s", cColors.Sprint(name), tokenWallet.Address(), tColors.Sprintf("%s %s", AmountString, wal.Address), err)
			err = UpdateWalletsBalanceSol(name, symbol, token.Address, accounts)
			continue
		}
		err = wal.AddTX(symbol, transferToken, tokenAmount)
		if err != nil {
			log.Fatalf("[%s] Wallet (%v) failed to add TX \"%s\" with amount %s, %s", cColors.Sprint(name), tokenWallet.Address(), transferToken, tColors.Sprintf("%s %s", AmountString, wal.Address), err)
			sendMu.Unlock()
			continue
		}
		tx := transferToken
		//tx := "non"

		var tokenBalance float64
		tokenBalance, err = blockchain.GetSolTokenBalance(tokenWallet.Address(), token.Address)
		if err != nil {
			sendMu.Unlock()
			log.Errorf("Wallet (%v) failed to get %v token balance, %v", tokenWallet.Address(), cColors.Sprint(symbol), err)
			continue
		}
		tokenWallet.SetTokenBalance(symbol, tokenBalance)
		tokenWallet.SOLBalance, err = blockchain.GetSolBalance(tokenWallet.Address())
		if err != nil {
			sendMu.Unlock()
			log.Errorf("Wallet (%v) failed to get %v coin balance, %v", tokenWallet.Address(), cColors.Sprint(symbol), err)
			continue
		}
		link := termlink.Link(wal.AddressMask(), fmt.Sprintf("https://solscan.io/tx/%s", tx))
		link = color.New(color.BgBlack).Add(color.FgWhite).Add(color.Bold).Sprintf(link)
		b := fmt.Sprintf("%.4f %s", tokenBalance, symbol)
		c := fmt.Sprintf("%.4f Sol", tokenWallet.SOLBalance)
		tBalance := tColors.Sprint(b)
		cBalance := bColors.Sprint(c)
		sleep, err = token.GetSleep()
		if err != nil {
			log.Errorf("Error getting token sleep time of address: %v, %v setting sleep to 180 seconds", token.Address, err)
			sleep = 180
		}
		log.Infof("[%s] %d. %s sent to %s from %s remaining %s %v, %s, next in %ds, r: %f", cColors.Sprint(name), token.Counter, tColors.Sprintf("$%.2f", airdropAmount), link, tokenWallet.Address(), tBalance, cColors.Sprint(symbol), cBalance, sleep, token.Ratio)
		to := fmt.Sprintf("<a href='https://solscan.io/account/%s'>%s</a>", wal.Address, wal.AddressMask(true))
		_from := fmt.Sprintf("<a href='https://solscan.io/account/%s'>%s</a>", tokenWallet.Address().String(), tokenWallet.Address().String())
		_tx := fmt.Sprintf("<a href='https://solscan.io/tx/%s'>Transaction</a>", tx)
		text := "✅ %s • %d • Next in %ds\n\n🔁 From: %s\n\n➡️ To: %s\n\n💰 Amount: %.3f %s ($%.3f) • %s\n\n📉 Remaining: %s • %s"
		text = fmt.Sprintf(text, name, token.Counter, sleep, _from, to, tokenAmount, symbol, airdropAmount, _tx, b, c)
		bot.Send(&tele.Chat{ID: database.Config.TelegramBot.AnnounceChannel}, text, tele.NoPreview, tele.ModeHTML)

		// send to public tx channel
		if tokenAmount > 5000 {
			text = fmt.Sprintf("https://solscan.io/tx/%s\n↗️ Sent %f Jin 💥", tx, tokenAmount)
			var _bot *tele.Bot
			_bot, err = tele.NewBot(tele.Settings{Token: "6898034177:AAE7RL_nLJwiVjKNZEO-CE6t3zCMA62aTas"})
			if err == nil {
				_bot.Send(&tele.Chat{ID: -1002133731605}, text, tele.ModeHTML)
			}

		}
		token.Counter++
		sendMu.Unlock()
		log.Debugf("Sleep for %d seconds", sleep)
	}
}

var lastMessageSol map[common2.PublicKey]*time.Time

const durationSol time.Duration = 5 * 60 * time.Second

func getRandomWalletWithBalanceSol(wallets []*blockchain.SolClient, name, symbol string, amount float64, cost float64) *blockchain.SolClient {
	if lastMessageSol == nil {
		lastMessageSol = make(map[common2.PublicKey]*time.Time)
	}
	annID := database.Config.TelegramBot.AnnounceChannel
	now := time.Now()
	copiedSlice := make([]*blockchain.SolClient, len(wallets))
	for i, account := range wallets {
		copiedAccount := *account
		copiedSlice[i] = &copiedAccount
	}
	rand.Shuffle(len(copiedSlice), func(i, j int) {
		copiedSlice[i], copiedSlice[j] = copiedSlice[j], copiedSlice[i]
	})
	var account *blockchain.SolClient
	for _, acc := range copiedSlice {
		if acc.TokenBalance(symbol) >= amount {
			if acc.SOLBalance >= cost {
				account = acc
				continue
			}
			//bnbAmount := color.YellowString("%.4f BNB", blockchain.BigIntToAmount(acc.BNBBalance, 18))
			tAmount := fmt.Sprintf("%.4f %s", acc.TokenBalance(symbol), symbol)
			log.Warnf("Wallet %s ran out of Sol, Currnet Balance: %v", acc.Address(), acc.TokenBalance(symbol))
			if _, ok := lastMessageSol[acc.Address()]; !ok || lastMessageSol[acc.Address()].Add(duration).Before(now) {
				teleText := fmt.Sprintf("⚠️ %s\n\nWallet ran out of Sol\n\nCurrnet Balance:\n%.4f Sol\n%s\n\n💼 Wallet: %s", name, acc.SOLBalance, tAmount, acc.Address().String())
				_, err := bot.Send(&tele.Chat{ID: annID}, teleText)
				if err != nil {
					log.Errorf("error sending telegram message: %v", err)
				}
				lastMessageSol[acc.Address()] = &now

			}
			continue
		}
		bnbAmount := fmt.Sprintf("%.4f Sol", acc.SOLBalance)
		tAmount := fmt.Sprintf("%.4f %s", acc.TokenBalance(symbol), symbol)
		log.Warnf("Wallet %s ran out of %s, Currnet Balance: %s", acc.Address(), symbol, tAmount)
		if _, ok := lastMessageSol[acc.Address()]; !ok || lastMessageSol[acc.Address()].Add(duration).Before(now) {
			teleText := fmt.Sprintf("⚠️ %s\n\nWallet ran out of %s\n\nCurrnet Balance:\n%s\n%s\n\n💼 Wallet: %s", name, symbol, bnbAmount, tAmount, acc.Address().String())
			_, err := bot.Send(&tele.Chat{ID: annID}, teleText)
			if err != nil {
				log.Errorf("error sending telegram message: %v", err)
			}
			lastMessageSol[acc.Address()] = &now
		}
	}
	return account
}

func UpdateWalletsBalanceSol(name, symbol string, mintAddr string, wallets []*blockchain.SolClient) (err error) {
	for _, w := range wallets {
		var tokenBalance float64
		tokenBalance, err = blockchain.GetSolTokenBalance(w.Address(), mintAddr)
		if err != nil {
			log.Errorf("Wallet (%v) failed to get %v token balance, %v", w.Address(), name, err)
			continue
		}
		w.SetTokenBalance(symbol, tokenBalance)
		w.SOLBalance, err = blockchain.GetSolBalance(w.Address())
		if err != nil {
			log.Errorf("Wallet (%v) failed to get %v coin balance, %v", w.Address(), name, err)
			continue
		}
		tBalance := fmt.Sprintf("%0.3f", tokenBalance)
		cBalance := fmt.Sprintf("%.4f Sol", w.SOLBalance)
		log.Infof("[%v] Wallet loaded: %v, balance: %s %v, %s", name, w.Address(), tBalance, symbol, cBalance)

	}
	return
}
