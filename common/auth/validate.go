package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type AuthClient struct {
	client  http.Client
	address string
}

func CreateAuthClient() (AuthClient, error) {
	client := AuthClient{}
	client.address = os.Getenv("AUTH_SERVER_ADDRESS")
	if len(client.address) == 0 {
		return client, errors.New("Auth server address is not specified")
	}
	return client, nil
}

type ErrResponseWithStatus struct {
	StatusCode  int
	RemoteError error
}

func (e *ErrResponseWithStatus) Error() string {
	return e.RemoteError.Error()
}

type errorResponse struct {
	Error string `json:"error"`
}

func (c *AuthClient) Validate(token string) error {
	req, err := http.NewRequest("GET", c.address, nil)
	if err != nil {
		return err
	}
	req.Header.Set("auth", token)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		var respError errorResponse
		err := json.NewDecoder(resp.Body).Decode(&respError)
		if err == nil {
			err = errors.New(respError.Error)
		}
		return &ErrResponseWithStatus{resp.StatusCode, err}
	}
	return nil
}
