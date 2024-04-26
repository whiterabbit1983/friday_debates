package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	askName         = "What is your name?"
	askAboutInfo    = "Tell me a bit about yourself"
	userInfoUpdated = "Your profile info has been updated"
	stepAskName     = "askName"
	stepAskAbout    = "askAbout"
	situation       = "You are on a talk show"
	sessionReset    = "Your debates have been stopped"
	noActiveSession = "no any active debates found, choose another character using /chars"
	letsGo          = "Let the battle begin 🔫"
)

func (s *Service) CreateSession(bot *tgbotapi.BotAPI, chatId int64, agentId string) error {
	// user, err := s.getUserInfo(chatId)

	// if err != nil {
	// 	return fmt.Errorf("error getting user info for create session: %v", err)
	// }

	//TODO: the user ID needs to comr form the DB
	id, err := s.createAPISession(s.Config.UserID, agentId, situation)

	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}

	//FIXME: needs to be removed
	id = s.Config.SessionID

	if err = s.newSession(chatId, id); err != nil {
		return fmt.Errorf("error updating user with new session id: %v", err)
	}

	log.Printf("session created: %s\n", id)

	doneMsg := tgbotapi.NewMessage(chatId, letsGo)

	if _, err := bot.Send(doneMsg); err != nil {
		return fmt.Errorf("error sending done message: %v", err)
	}

	return nil
}

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
	text := "⚔️ Choose your opponent ⚔️\n"

	for _, a := range agents.Items {
		b := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⚔️ %s", a.Name), a.ID)
		buttons = append(buttons, b)
		text += fmt.Sprintf("• %s: %s", a.Name, a.About)
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

func (s *Service) OnResetSession(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	if err := s.resetSession(message.Chat.ID); err != nil {
		return fmt.Errorf("error resetting session: %v", err)
	}

	nameMsg := tgbotapi.NewMessage(message.Chat.ID, sessionReset)

	if _, err := bot.Send(nameMsg); err != nil {
		return fmt.Errorf("error sending session reset message: %v", err)
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
	row, err := s.DB.QueryRow(fmt.Sprintf("select chat_id, name, about, user_id, session_id from user_info where chat_id = %d", chatId))

	if err != nil {
		return nil, fmt.Errorf("error getting user info: %v", err)
	}

	var cid int64
	var n string
	var a string
	var uid string
	var sid string

	err = row.Scan(&cid, &n, &a, &uid, &sid)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling user info: %v", err)
	}

	user := UserInfo{
		ChatID:    cid,
		Name:      n,
		About:     a,
		UserID:    uid,
		SessionID: sid,
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
	var freeInput bool
	step, err := s.getStep(message.Chat.ID)

	if err != nil && strings.Contains(err.Error(), "not found") {
		freeInput = true
	} else if err != nil {
		return fmt.Errorf("error getting on input flow: %v", err)
	}

	if !freeInput {
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
	} else {
		sessionID, err := s.getActiveSession(message.Chat.ID)

		if err != nil && strings.Contains(err.Error(), "not found") {
			nameMsg := tgbotapi.NewMessage(message.Chat.ID, noActiveSession)

			if _, err := bot.Send(nameMsg); err != nil {
				return fmt.Errorf("error sending message: %v", err)
			}

			return nil
		} else if err != nil {
			return fmt.Errorf("error getting session info: %v", err)
		}

		user, err := s.getUserInfo(message.Chat.ID)

		if err != nil {
			return fmt.Errorf("error getting user info: %v", err)
		}

		msg := Message{
			Role:    "user",
			Content: message.Text,
			Name:    user.Name,
		}

		resp, err := s.chatJulep(sessionID, msg)

		if err != nil {
			return fmt.Errorf("error chatting: %v", err)
		}

		if len(resp.Response) == 0 {
			return errors.New("empty response")
		}

		reply := tgbotapi.NewMessage(
			message.Chat.ID,
			strings.Trim(resp.Response[0][0].Content, " \n"),
		)

		if _, err := bot.Send(reply); err != nil {
			return fmt.Errorf("error sending reply message: %v", err)
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
		err = s.DB.Exec("delete from session_info where chat_id = ?", chatId)

		if err != nil {
			return fmt.Errorf("error updating user's session_id: %v", err)
		}

		err = s.DB.Exec("update session_info set session_id = ? where chat_id = ?", sessionId, chatId)

		if err != nil {
			return fmt.Errorf("error updating user's session_id: %v", err)
		}
	}

	return nil
}

func (s *Service) newSession(chatId int64, sessionId string) error {
	err := s.DB.Exec("delete from session_info where chat_id = ?", chatId)

	if err != nil {
		return fmt.Errorf("error setting wen session_id: %v", err)
	}

	err = s.DB.Exec("insert into session_info (session_id, chat_id) values (?, ?)", sessionId, chatId)

	if err != nil {
		return fmt.Errorf("error setting new session_id: %v", err)
	}

	return nil
}

func (s *Service) resetSession(chatId int64) error {
	err := s.DB.Exec("delete from session_info where chat_id = ?", chatId)

	if err != nil {
		return fmt.Errorf("error resetting session_id: %v", err)
	}

	return nil
}

func (s *Service) getActiveSession(chatId int64) (string, error) {
	row, err := s.DB.QueryRow("select * from session_info where chat_id = ?", chatId)

	if err != nil {
		return "", fmt.Errorf("error getting user info: %v", err)
	}

	m := make(map[string]any)

	err = row.MapScan(m)

	if err != nil {
		return "", fmt.Errorf("error unmarshalling user info: %v", err)
	}

	sid, ok := m["session_id"]

	if !ok {
		return "", nil
	}

	newSid := sid.(string)

	return newSid, nil
}
