package service

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	Config      Env
	UserInfoBox *UserInfoBox
	FlowStepBox *FlowStepBox
	Api         *Api
}

func New(config Env, ub *UserInfoBox, fb *FlowStepBox, api *Api) *Service {
	return &Service{
		Config:      config,
		UserInfoBox: ub,
		FlowStepBox: fb,
		Api:         api,
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

		if update.Message.IsCommand() {
			var err error

			cmd := update.Message.Command()

			switch cmd {
			case "start":
				err = s.OnStart(bot, update.Message)
			case "chars":
				err = s.OnChars(bot, update.Message)
			case "set_info":
				err = s.OnSetInfo(bot, update.Message)
			case "info":
				err = s.OnGetInfo(bot, update.Message)
			}

			if err != nil {
				log.Printf("error handling '/%s' command: %v", cmd, err)
			}

			continue
		}

		if update.CallbackQuery != nil {
			continue
		}

		if err := s.OnUserInput(bot, update.Message); err != nil {
			log.Println("error handling user message:", err)
		}
	}
}
