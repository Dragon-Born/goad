package cron

import (
	log "github.com/sirupsen/logrus"
	"goad/database"
	"math/big"
	"time"
)

func getAllDexTransactionCron() {
	var lastBlock *big.Int
	for {
		existsWallet := 0
		currentBlock := client.GetCurrentBlock()
		if lastBlock != nil && currentBlock.Int64() == lastBlock.Int64() {
			time.Sleep(5 * time.Second)
			continue
		}
		transactions, err := client.GetBlockTransactions(currentBlock.Int64())
		if err != nil {
			log.Errorf("fetching transactions error: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}
		for _, transaction := range transactions {
			_, err = database.CreateUniqueWallet(transaction.From.Hex(), "DEX")
			if err != nil {
				existsWallet++
			}
		}
		log.Debugf("adding %d dex wallet from block %d, existed wallets: %d", len(transactions), currentBlock, existsWallet)
		time.Sleep(3 * time.Second)
	}
}
