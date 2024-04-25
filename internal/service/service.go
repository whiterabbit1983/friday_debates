package service

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	Config Env
}

func New(config Env) *Service {
	return &Service{
		Config: config,
	}
}

func (s *Service) Run() {
	bot, err := tgbotapi.NewBotAPI(s.Config.BotToken)
	if err != nil {
		log.Fatalln(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Example:
		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		// bot.Send(msg)
	}
}
