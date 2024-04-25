package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type UserDto struct {
	Name  string `json:"name" validate:"required"`
	About string `json:"about" validate:"required"`
}

type CreateUserResponseDto struct {
	ID string `json:"id" validate:"required"`
}

func (s *Service) createAPIUser(name, about string) (string, error) {
	b, err := json.Marshal(UserDto{
		Name:  name,
		About: about,
	})

	if err != nil {
		return "", fmt.Errorf("error marshalling user: %v", err)
	}

	url := fmt.Sprintf("%s/%s", s.Config.JulepBaseUrl, "users")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Config.JulepApiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending create user Julep request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %v", err)
	}

	var user CreateUserResponseDto

	if err = json.Unmarshal(body, &user); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	return user.ID, nil
}

func (s *Service) getAPIUser(userId string) (*UserDto, error) {
	url := fmt.Sprintf("%s/%s/%s", s.Config.JulepBaseUrl, "users", userId)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Config.JulepApiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending create user Julep request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	var user UserDto

	if err = json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return &user, nil
}
