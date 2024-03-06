package cron

import (
	log "github.com/sirupsen/logrus"
	"goad/database"
	"goad/pkg/blockchain"
	tele "gopkg.in/telebot.v3"
)

var client *blockchain.Client
var bot *tele.Bot

func RunCron() (err error) {
	client, err = blockchain.NewClient(database.Config.DataSeed)
	if err != nil {
		return err
	}
	bot, err = tele.NewBot(tele.Settings{Token: database.Config.TelegramBot.Token})
	if err != nil {
		return err
	}
	log.Infof("Telegram bot %s @%s (%v) connected", bot.Me.FirstName, bot.Me.Username, bot.Me.ID)
	for _, token := range database.Config.Tokens {
		go SendToken(token)
	}
	getAllDexTransactionCron()
	return
}
