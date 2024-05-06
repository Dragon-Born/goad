package cron

import (
	log "github.com/sirupsen/logrus"
	"goad/database"
	"goad/pkg/blockchain"
	tele "gopkg.in/telebot.v3"
)

var client *blockchain.Client
var solClient *blockchain.SolClient
var bot *tele.Bot

func RunCron() (err error) {
	client, err = blockchain.NewClient(database.Config.DataSeed)
	if err != nil {
		return err
	}
	solClient = blockchain.NewSolanaClient(
		"https://api.mainnet-beta.solana.com",
	)
	bot, err = tele.NewBot(tele.Settings{Token: database.Config.TelegramBot.Token})
	if err != nil {
		return err
	}
	log.Infof("Telegram bot %s @%s (%v) connected", bot.Me.FirstName, bot.Me.Username, bot.Me.ID)
	for _, token := range database.Config.Tokens {
		if token.Active {
			if token.Chain == "bsc" {
				go SendToken(token)
			} else if token.Chain == "sol" {
				go SendTokenSOL(token)
			} else {
				log.Fatalf("Unsupported chain %v", token.Chain)
			}
		} else {
			log.Infof("Token %v is disabled", token.Address)
		}
	}
	go getAllDexTransactionCron()
	getAllSolDexTransactionCron()

	return
}
