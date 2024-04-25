package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Api struct {
	BaseURL       string
	Authorization string
	ContentType   string
}

func NewAPI(baseURL string, auth string, ctype string) *Api {
	return &Api{
		BaseURL:       baseURL,
		Authorization: auth,
		ContentType:   ctype,
	}
}

func GetCall[T any](api *Api, path ...string) (*T, error) {
	url := strings.Join(path, "/")

	req, err := makeRequest[[]byte]("GET", url, api.Authorization, api.ContentType, nil)

	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	result, err := doCall[T](req)

	if err != nil {
		return nil, fmt.Errorf("error making API call: %v", err)
	}

	return result, nil
}

func PostCall[T any, U any](api *Api, data *U, path ...string) (*T, error) {
	url := strings.Join(path, "/")

	req, err := makeRequest("POST", url, api.Authorization, api.ContentType, data)

	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	result, err := doCall[T](req)

	if err != nil {
		return nil, fmt.Errorf("error making API call: %v", err)
	}

	return result, nil
}

func makeRequest[T any](method, url, auth, ctype string, data *T) (*http.Request, error) {
	var body io.Reader

	if data != nil {
		b, err := json.Marshal(data)

		if err != nil {
			return nil, fmt.Errorf("error marshalling request data: %v", err)
		}

		body = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	req.Header.Set("Content-Type", ctype)

	return req, nil
}

func doCall[T any](req *http.Request) (*T, error) {
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

	var result T

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return &result, nil
}
