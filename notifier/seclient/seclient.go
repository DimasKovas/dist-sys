package seclient

import (
	"errors"
	"net/http"
	"os"
)

type Client struct {
	address string
	apiId   string
	client  http.Client
}

func CreateSeClient() (Client, error) {
	var client Client
	client.address = os.Getenv("EMAIL_PROVIDER_ADDRESS")
	client.apiId = os.Getenv("PROVIDER_API_ID")
	if len(client.address) == 0 {
		return client, errors.New("Email provider address is not specified")
	}
	if len(client.apiId) == 0 {
		return client, errors.New("Provider api_id is not specified")
	}
	return client, nil
}

func (c *Client) Send(to string, msg string) error {
	req, err := http.NewRequest("GET", c.address, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("api_id", c.apiId)
	q.Add("to", to)
	q.Add("msg", msg)
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Send request failed")
	}
	return nil
}
