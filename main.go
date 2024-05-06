package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"goad/cron"
	"goad/database"
	"goad/pkg/encryption"
	"goad/watcher"
)

// GetBlockTransactions connects to an Ethereum node, fetches a block by its number, and prints its transactions.

func main() {
	//blockNumber := int64(36703824) // Example block number, change it to the block number you're interested in
	//spew.Dump(blockchain.GetBlockTransactions(blockNumber))

	var wal string
	var oldDb string
	flag.StringVar(&wal, "w", "", "encrypt wallet address private key")
	flag.StringVar(&oldDb, "import", "", "import from old file")
	var flagNoColor = flag.Bool("no-color", false, "Disable color output")

	if *flagNoColor {
		color.NoColor = true // disables colorized output
	}
	flag.Parse()
	if wal != "" {
		//_, err := wallet.FromPrivateKey(wal)
		//if err != nil {
		//	log.Error("error encrypting wallet private key: invalid wallet address\n")
		//	return
		//}
		hex, err := encryption.EncryptHex(wal, cron.Password)
		if err != nil {
			log.Errorf("[Error] error encrypting wallet private key: %v\n", err)
			return
		}
		fmt.Printf("Encrypted wallet address: %v\n", hex)
		return
	}

	err := database.InitDB()
	if err != nil {
		log.Errorf("database failed to run: %v", err)
		return
	}

	if oldDb != "" {
		err = database.InitOldDB(oldDb)
		if err != nil {
			log.Errorf("old database failed to run: %v", err)
			return
		}
		count, err := database.ImportOldDB()
		if err != nil {
			log.Errorf("error importing database: %v", err)
			return
		}
		log.Infof("Finished importing %v wallets", count)
		return
	}
	go watcher.WatchConfig()

	err = cron.RunCron()
	if err != nil {
		log.Errorf("cron jobs failed to run: %v", err)
		return
	}
}
