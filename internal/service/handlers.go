package service

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	askName         = "What is your name?"
	askAboutInfo    = "Tell me a bit about yourself"
	userInfoUpdated = "Your profile info has been updated"
	stepAskName     = "askName"
	stepAskAbout    = "askAbout"
)

func (s *Service) OnStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	greets := tgbotapi.NewMessage(
		message.Chat.ID,
		"Welcome to Friday Debates arena! Are you ready?\n"+
			"- Use /chars command to see all available characters to talk to.\n"+
			"- Use /set_info command to set your personal information.\n"+
			"- Use /info command to view your profile info.",
	)

	if _, err := bot.Send(greets); err != nil {
		return err
	}

	return nil
}

func (s *Service) OnChars(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	agents, err := s.listAPIAgents()

	if err != nil {
		return fmt.Errorf("error getting agents list: %v", err)
	}

	buttons := make([]tgbotapi.InlineKeyboardButton, 0, len(agents.Items))
	text := "Here is the list of characters:\n"

	for _, a := range agents.Items {
		b := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⚔️ %s", a.Name), a.ID)
		buttons = append(buttons, b)
		text += fmt.Sprintf("- %s: %s", a.Name, a.About)
	}

	agentsList := tgbotapi.NewInlineKeyboardMarkup(buttons)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = agentsList

	if _, err := bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func (s *Service) OnSetInfo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	nameMsg := tgbotapi.NewMessage(message.Chat.ID, askName)

	if _, err := bot.Send(nameMsg); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	if err := s.updateStep(message.Chat.ID, stepAskName); err != nil {
		return fmt.Errorf("error adding flow step: %v", err)
	}

	return nil
}

func (s *Service) OnGetInfo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	user, err := s.getUserInfo(message.Chat.ID)

	if err != nil {
		respMsg := tgbotapi.NewMessage(message.Chat.ID, "You haven't created your user yet, please run /set_info")
		if _, err := bot.Send(respMsg); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}

		return nil
	}

	info, err := s.getAPIUser(user.UserID)

	if err != nil {
		return fmt.Errorf("error getting user info from API: %v", err)
	}

	respMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("your name: %s, about info: %s", info.Name, info.About))
	if _, err := bot.Send(respMsg); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}

func (s *Service) getUserInfo(chatId int64) (*UserInfo, error) {
	row, err := s.DB.QueryRow(fmt.Sprintf("select * from user_info where chat_id = %d", chatId))

	if err != nil {
		return nil, fmt.Errorf("error getting user info: %v", err)
	}

	var user UserInfo

	err = row.StructScan(&user)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling user info: %v", err)
	}

	return &user, nil
}

func (s *Service) getStep(chatId int64) (*FlowStep, error) {
	row, err := s.DB.QueryRow(fmt.Sprintf("select * from flow_steps where chat_id = %d", chatId))

	if err != nil {
		return nil, fmt.Errorf("error getting flow info: %v", err)
	}

	var step FlowStep

	err = row.StructScan(&step)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling flow info: %v", err)
	}

	return &step, nil
}

func (s *Service) OnUserInput(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	step, err := s.getStep(message.Chat.ID)

	if err != nil {
		return fmt.Errorf("error getting on input flow: %v", err)
	}

	switch step.Step {
	case stepAskName:
		if err := s.updateStep(message.Chat.ID, stepAskAbout); err != nil {
			return fmt.Errorf("error updating flow step: %v", err)
		}

		if err := s.updateUser(message.Chat.ID, message.Text, "", "", ""); err != nil {
			return fmt.Errorf("error setting user name: %v", err)
		}

		aboutMsg := tgbotapi.NewMessage(message.Chat.ID, askAboutInfo)

		if _, err := bot.Send(aboutMsg); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}
	case stepAskAbout:
		if err = s.removeStep(message.Chat.ID); err != nil {
			return fmt.Errorf("error removing flow step: %v", err)
		}

		if err = s.updateUser(message.Chat.ID, "", message.Text, "", ""); err != nil {
			return fmt.Errorf("error setting user about: %v", err)
		}

		if err = s.createJulepUser(message.Chat.ID); err != nil {
			return fmt.Errorf("error creating Julep user: %v", err)
		}

		aboutMsg := tgbotapi.NewMessage(message.Chat.ID, userInfoUpdated)

		if _, err := bot.Send(aboutMsg); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}

func (s *Service) createJulepUser(chatId int64) error {
	user, err := s.getUserInfo(chatId)

	if err != nil {
		return fmt.Errorf("error getting user info: %v", err)
	}

	newUserId, err := s.createAPIUser(user.Name, user.About)

	if err != nil {
		return fmt.Errorf("error accessing Julep user API: %v", err)
	}

	if err = s.updateUser(chatId, "", "", newUserId, ""); err != nil {
		return fmt.Errorf("error updating new user ID: %v", err)
	}

	return nil
}

func (s *Service) updateStep(chatId int64, newStep string) error {
	var upsert bool

	_, err := s.DB.QueryRow(fmt.Sprintf("select * from flow_steps where chat_id = %d limit 1", chatId))

	if err != nil {
		upsert = true
	}

	if upsert {
		if err := s.DB.Exec(fmt.Sprintf("insert into flow_steps (chat_id, step) values (%d, '%s')", chatId, newStep)); err != nil {
			return fmt.Errorf("error inserting step: %v", err)
		}
	} else {
		if err := s.DB.Exec(fmt.Sprintf("update flow_steps set step = '%s' where chat_id = %d", newStep, chatId)); err != nil {
			return fmt.Errorf("error updating step: %v", err)
		}
	}

	return nil
}

func (s *Service) removeStep(chatId int64) error {
	if err := s.DB.Exec(fmt.Sprintf("delete from flow_steps where chat_id = %d", chatId)); err != nil {
		return fmt.Errorf("error updating step: %v", err)
	}

	return nil
}

func (s *Service) updateUser(chatId int64, name, about, userId, sessionId string) error {
	var upsert bool

	_, err := s.DB.QueryRow(fmt.Sprintf("select * from flow_steps where chat_id = %d limit 1", chatId))

	if err != nil {
		upsert = true
	}

	if len(name) > 0 {
		if upsert {
			err = s.DB.Exec(fmt.Sprintf("insert into user_info (name, chat_id) values ('%s', %d)", name, chatId))
			upsert = false
		} else {
			err = s.DB.Exec(fmt.Sprintf("update user_info set name = '%s' where chat_id = %d", name, chatId))
		}

		if err != nil {
			return fmt.Errorf("error updating user name: %v", err)
		}
	}

	if len(about) > 0 {
		if upsert {
			err = s.DB.Exec(fmt.Sprintf("insert into user_info (about, chat_id) values ('%s', %d)", about, chatId))
			upsert = false
		} else {
			err = s.DB.Exec(fmt.Sprintf("update user_info set about = '%s' where chat_id = %d", about, chatId))
		}

		if err != nil {
			return fmt.Errorf("error updating user about info: %v", err)
		}
	}

	if len(userId) > 0 {
		if upsert {
			err = s.DB.Exec(fmt.Sprintf("insert into user_info (user_id, chat_id) values ('%s', %d)", userId, chatId))
			upsert = false
		} else {
			err = s.DB.Exec(fmt.Sprintf("update user_info set user_id = '%s' where chat_id = %d", userId, chatId))
		}

		if err != nil {
			return fmt.Errorf("error updating user's user_id: %v", err)
		}
	}

	if len(sessionId) > 0 {
		if upsert {
			err = s.DB.Exec(fmt.Sprintf("insert into user_info (session_id, chat_id) values ('%s', %d)", sessionId, chatId))
			upsert = false
		} else {
			err = s.DB.Exec(fmt.Sprintf("update user_info set session_id = '%s' where chat_id = %d", sessionId, chatId))
		}

		if err != nil {
			return fmt.Errorf("error updating user's session_id: %v", err)
		}
	}

	return nil
}
