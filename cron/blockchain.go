package cron

import (
	log "github.com/sirupsen/logrus"
	"goad/database"
	"goad/pkg/blockchain"
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
			_, err = database.CreateUniqueWallet(transaction.From.Hex(), "DEX", database.Binance)
			if err != nil {
				existsWallet++
			}
		}
		log.Debugf("[BSC] adding %d dex wallet from block %d, existed wallets: %d", len(transactions), currentBlock, existsWallet)
		time.Sleep(3 * time.Second)
	}
}

func getAllSolDexTransactionCron() {

	var lastBlock *int64
	for {
		existsWallet := 0
		currentBlock, err := solClient.GetCurrentBlock()
		if err != nil {
			log.Errorf("fetching current block error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if lastBlock != nil && currentBlock.BlockHeight == lastBlock {
			time.Sleep(5 * time.Second)
			continue
		}
		transactions, err := blockchain.GetBlockTransactions(currentBlock)
		if err != nil {
			log.Errorf("fetching transactions error: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}
		for _, transaction := range transactions {
			balance, err := blockchain.GetSolBalance(transaction.From)
			if err != nil || balance == 0 {
				time.Sleep(1 * time.Second)
				continue
			}
			_, err = database.CreateUniqueWallet(transaction.From.String(), "DEX", database.Solana)
			if err != nil {
				existsWallet++
			}
			time.Sleep(1 * time.Second)
		}
		log.Debugf("[SOL] adding %d dex wallet from block %d, existed wallets: %d", len(transactions), currentBlock.BlockHeight, existsWallet)
		time.Sleep(3 * time.Second)
	}
}
