package service

import (
	"fmt"
)

type UserDto struct {
	Name  string `json:"name" validate:"required"`
	About string `json:"about" validate:"required"`
}

type CreateUserResponseDto struct {
	ID string `json:"id" validate:"required"`
}

type CreateSessionResponseDto struct {
	ID string `json:"id" validate:"required"`
}

type SessionDto struct {
	UserID    string `json:"user_id" validate:"required"`
	AgentID   string `json:"agent_id" validate:"required"`
	Situation string `json:"situation" validate:"required"`
}

type AgentDto struct {
	ID    string `json:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	About string `json:"about" validate:"required"`
}

type ListAgentsResponseDto struct {
	Items []AgentDto `json:"items" validate:"required"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name"`
}

type ChatRequest struct {
	Messages []Message `json:"messages" validate:"required"`
}

type ChatResponseDto struct {
	Response [][]Message `json:"response" validate:"required"`
}

func (s *Service) createAPIUser(name, about string) (string, error) {
	data := &UserDto{
		Name:  name,
		About: about,
	}

	user, err := PostCall[CreateUserResponseDto](s.Api, data, s.Config.JulepBaseUrl, "users")

	if err != nil {
		return "", fmt.Errorf("error calling POST: %v", err)
	}

	return user.ID, nil
}

func (s *Service) getAPIUser(userId string) (*UserDto, error) {
	user, err := GetCall[UserDto](s.Api, s.Config.JulepBaseUrl, "users", userId)

	if err != nil {
		return nil, fmt.Errorf("error calling GET: %v", err)
	}

	return user, nil
}

func (s *Service) createAPISession(userId, agentId, situation string) (string, error) {
	data := &SessionDto{
		UserID:    userId,
		AgentID:   agentId,
		Situation: situation,
	}

	user, err := PostCall[CreateSessionResponseDto](s.Api, data, s.Config.JulepBaseUrl, "users")

	if err != nil {
		return "", fmt.Errorf("error calling POST: %v", err)
	}

	return user.ID, nil
}

func (s *Service) listAPIAgents() (*ListAgentsResponseDto, error) {
	agents, err := GetCall[ListAgentsResponseDto](s.Api, s.Config.JulepBaseUrl, "agents")

	if err != nil {
		return nil, fmt.Errorf("error calling agents GET: %v", err)
	}

	return agents, nil
}

func (s *Service) chatJulep(sessionId string, message Message) (*ChatResponseDto, error) {
	req := &ChatRequest{
		Messages: []Message{message},
	}
	r, err := PostCall[ChatResponseDto](s.Api, req, s.Config.JulepBaseUrl, "sessions", sessionId, "chat")

	if err != nil {
		return nil, fmt.Errorf("error calling session chat: %v", err)
	}

	return r, nil
}
