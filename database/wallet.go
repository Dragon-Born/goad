package database

import (
	"database/sql"
	"errors"
	"fmt"
	common2 "github.com/blocto/solana-go-sdk/common"
	"github.com/fatih/color"
	"goad/pkg/airdrop"
	"gorm.io/gorm"
	"time"
)

type Wallet struct {
	ID              uint           `gorm:"primarykey"`
	Address         string         `gorm:"index;unique" json:"user_id"`
	TotalBalance    sql.NullInt64  `token:"total_balance"`
	Chain           string         `json:"chain" gorm:"index;type:varchar(10)"`
	Token           sql.NullString `json:"token" gorm:"index;type:varchar(10)"`
	Type            string         `json:"type" gorm:"index;type:varchar(10)"`
	TX              sql.NullString `gorm:"type:varchar(100)" json:"tx"`
	TransactionDate *time.Time     `json:"transaction_date"`
	Amount          sql.NullInt64
	Enabled         bool
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (w *Wallet) SolAddress() common2.PublicKey {
	return common2.PublicKeyFromString(w.Address)

}

func (w *Wallet) AddressMask(noColor ...bool) string {
	maskAddress := w.Address
	if noColor != nil && noColor[0] {
		return fmt.Sprintf("%s...%s", maskAddress[:6], maskAddress[len(maskAddress)-4:])
	}
	return color.CyanString(fmt.Sprintf("%s...%s", maskAddress[:6], maskAddress[len(maskAddress)-4:]))
}

type BlockChains string

const (
	Ethereum BlockChains = "ethereum"
	Solana   BlockChains = "solana"
	Binance  BlockChains = "binance"
)

func CreateUniqueWallet(address string, _type string, blockChain BlockChains) (wal *Wallet, err error) {
	var count int64
	DB.Model(&Wallet{}).Where("address = ?", address).Count(&count)
	if count > 0 {
		return nil, errors.New("a wallet with this address already exists")
	}
	wal = &Wallet{Address: address, Type: _type, Enabled: true, Chain: string(blockChain)}
	if err = DB.Create(wal).Error; err != nil {
		return nil, err
	}
	return
}

func getRandF() string {
	rand := "RAND()"
	if Config.Database.DBs[Config.Database.CurrentMode].Type == "sqlite" {
		rand = "RANDOM()"
	}
	return rand
}

func GetWalletsWithoutTX() ([]Wallet, error) {
	var wallets []Wallet

	err := DB.Where("(TX IS NULL OR TX = ?) AND enabled = ?", "", true).Order(getRandF()).Find(&wallets).Error
	if err != nil {
		return nil, err
	}
	return wallets, nil
}
func GetOneWalletWithoutTX(chain BlockChains) (*Wallet, error) {
	var wallet Wallet
	err := DB.Where("(TX IS NULL OR TX = ?) AND enabled = ? AND chain = ?", "", true, chain).Order(getRandF()).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (w *Wallet) GetBalance() (amount float64, err error) {
	if !w.TotalBalance.Valid {
		return 0, errors.New("balance is not set")
	}
	amount = float64(w.TotalBalance.Int64) / CoinRate
	return
}

func (w *Wallet) GenerateAmount() (value float64, err error) {
	balance, err := w.GetBalance()
	if err != nil {
		return 0, err
	}
	thresholds := []airdrop.Threshold{
		{Balance: 0.0001, AirdropAmt: 0.05},
		{Balance: 0.2, AirdropAmt: 0.3},
		{Balance: 0.3, AirdropAmt: 0.4},
		{Balance: 1, AirdropAmt: 0.5},
		{Balance: 5, AirdropAmt: 1},
		{Balance: 9, AirdropAmt: 1.5},
		{Balance: 10, AirdropAmt: 2},
		{Balance: 20, AirdropAmt: 3},
		{Balance: 30, AirdropAmt: 4},
		{Balance: 40, AirdropAmt: 5},
		{Balance: 50, AirdropAmt: 6},
		{Balance: 100, AirdropAmt: 10},
	}
	highestAmount := 0.0
	for _, threshold := range thresholds {
		if highestAmount < threshold.AirdropAmt {
			highestAmount = threshold.AirdropAmt
		}
	}
	value = airdrop.GenerateAirdropAmount(balance, thresholds)
	if value > highestAmount {
		return 0, errors.New(fmt.Sprintf("generated amount is higher than maximum airdrop amount, generated: %f, max: %f", value, highestAmount))
	}
	return
}
func (w *Wallet) SetBalance(amount float64) (err error) {
	err = w.TotalBalance.Scan(int64(amount * CoinRate))
	if err != nil {
		return err
	}
	err = DB.Save(w).Error
	return
}

func (w *Wallet) SetAmount(amount float64) (err error) {
	err = w.Amount.Scan(int64(amount * CoinRate))
	if err != nil {
		return err
	}
	err = DB.Save(w).Error
	return
}

func (w *Wallet) GetAmount() (amount float64, err error) {
	if !w.Amount.Valid {
		return 0, errors.New("balance is not set")
	}
	amount = float64(w.Amount.Int64) / CoinRate
	return
}

func (w *Wallet) Enable() (err error) {
	if w.Enabled {
		return errors.New("wallet is already enabled")
	}
	w.Enabled = true
	err = DB.Save(w).Error
	return
}

func (w *Wallet) Disable() (err error) {
	if !w.Enabled {
		return errors.New("wallet is already disabled")
	}
	w.Enabled = false
	err = DB.Save(w).Error
	return
}

func (w *Wallet) AddTX(token, tx string, amount float64) (err error) {
	err = w.TX.Scan(tx)
	if err != nil {
		return err
	}
	err = w.Amount.Scan(int64(amount * CoinRate))
	if err != nil {
		return err
	}
	err = w.Token.Scan(token)
	if err != nil {
		return err
	}
	now := time.Now()
	w.TransactionDate = &now
	err = DB.Save(w).Error
	return
}
