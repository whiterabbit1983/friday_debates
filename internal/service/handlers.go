package service

import (
	"errors"
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

var (
	ErrUserNotCreated = errors.New("user not created")
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
	greets := tgbotapi.NewMessage(message.Chat.ID, "Here is the list of characters:")

	if _, err := bot.Send(greets); err != nil {
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
	q := s.UserInfoBox.Query(UserInfo_.ChatID.Equals(message.Chat.ID))
	users, err := q.Limit(1).Find()

	if err != nil {
		return fmt.Errorf("error getting user info: %v", err)
	}

	if len(users) == 0 {
		respMsg := tgbotapi.NewMessage(message.Chat.ID, "You have not set your user info yet, use /set_info")
		if _, err := bot.Send(respMsg); err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}

		return ErrUserNotCreated
	}

	user := users[0]

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

func (s *Service) OnUserInput(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	q := s.FlowStepBox.Query(FlowStep_.ChatID.Equals(message.Chat.ID))

	steps, err := q.Limit(1).Find()

	if err != nil {
		return fmt.Errorf("error handling user input: %v", err)
	}

	if len(steps) > 0 {
		step := steps[0]

		if step == nil {
			return errors.New("step is nil")
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
	}

	return nil
}

func (s *Service) createJulepUser(chatId int64) error {
	q := s.UserInfoBox.Query(UserInfo_.ChatID.Equals(chatId))
	users, err := q.Limit(1).Find()

	if err != nil {
		return fmt.Errorf("error getting user info: %v", err)
	}

	user := users[0]

	if user == nil {
		return errors.New("user is nil")
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
	q := s.FlowStepBox.Query(FlowStep_.ChatID.Equals(chatId))
	if _, err := q.Remove(); err != nil {
		return fmt.Errorf("error updating step: %v", err)
	}

	if _, err := s.FlowStepBox.Put(&FlowStep{ChatID: chatId, Step: newStep}); err != nil {
		return fmt.Errorf("error putting new step: %v", err)
	}

	return nil
}

func (s *Service) removeStep(chatId int64) error {
	q := s.FlowStepBox.Query(FlowStep_.ChatID.Equals(chatId))
	if _, err := q.Remove(); err != nil {
		return fmt.Errorf("error updating step: %v", err)
	}

	return nil
}

func (s *Service) updateUser(chatId int64, name, about, userId, sessionId string) error {
	q := s.UserInfoBox.Query(UserInfo_.ChatID.Equals(chatId))
	users, err := q.Find()

	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no users found for chat ID: %d", chatId)
	}

	newName := users[0].Name
	newAbout := users[0].About
	newUserId := users[0].UserID
	newSessionId := users[0].SessionID

	if len(name) > 0 {
		newName = name
	}

	if len(about) > 0 {
		newAbout = about
	}

	if len(userId) > 0 {
		newUserId = userId
	}

	if len(sessionId) > 0 {
		newSessionId = sessionId
	}

	if _, err := q.Remove(); err != nil {
		return fmt.Errorf("error removing old user: %v", err)
	}

	newUser := &UserInfo{
		ChatID:    chatId,
		Name:      newName,
		About:     newAbout,
		UserID:    newUserId,
		SessionID: newSessionId,
	}

	if _, err := s.UserInfoBox.Put(newUser); err != nil {
		return fmt.Errorf("error putting user: %v", err)
	}

	return nil
}
