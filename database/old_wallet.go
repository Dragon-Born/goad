package database

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"time"
)

type WalletOld struct {
	ID         uint           `gorm:"primarykey"`
	Address    string         `gorm:"index;unique" json:"user_id"`
	BscAmount  float64        `json:"bsc_amount"`
	SentAmount float64        `json:"sent_amount" gorm:"index"`
	TX         sql.NullString `gorm:"type:varchar(100)" json:"tx"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (WalletOld) TableName() string {
	return "wallet"
}

func GetOldWalletsWithNonNilTX() ([]WalletOld, error) {
	var wallets []WalletOld
	if err := OldDB.Where("tx IS NOT NULL OR tx != ?", "").Find(&wallets).Error; err != nil {
		return nil, err
	}
	return wallets, nil
}

func ImportOldDB() (count int, err error) {
	wallets, err := GetOldWalletsWithNonNilTX()
	if err != nil {
		return 0, err
	}
	for _, wallet := range wallets {
		newWallet, err := CreateUniqueWallet(wallet.Address, "DEX", "")
		if err != nil {
			log.Errorf("Could not create wallet %v, %v", wallet.Address, err)
			continue
		}
		err = newWallet.AddTX("IMPORTED", wallet.TX.String, wallet.SentAmount)
		if err != nil {
			log.Errorf("Could not add TX to wallet %v, %v", wallet.Address, err)
			continue
		}
		if wallet.BscAmount != 0 {
			err = newWallet.SetBalance(float64(wallet.BscAmount) / CoinRate)
			if err != nil {
				log.Errorf("Could not set balance to wallet %v, %v", wallet.Address, err)
				continue
			}
		}
		count++
	}
	return
}
