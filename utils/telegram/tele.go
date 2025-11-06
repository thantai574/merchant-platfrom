package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func SendTelegram(message string, channelId int64) {
	bot, err := tgbotapi.NewBotAPI("967965284:AAFx77szTpDfQgYbZiXheuDDltnpeQKpwTY")
	if err != nil {
		fmt.Println(err)
	}

	_, err = bot.Send(tgbotapi.NewMessage(channelId, message))
	if err != nil {
		panic(err)
	}
}
