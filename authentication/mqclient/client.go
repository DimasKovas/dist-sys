package mqclient

import (
	"encoding/json"
	"os"

	"github.com/streadway/amqp"
)

type Client struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func CreateMqClient() (Client, error) {
	var client Client
	var err error
	client.connection, err = amqp.Dial(os.Getenv("MESSAGE_QUEUE_URL"))
	if err != nil {
		return client, err
	}
	client.channel, err = client.connection.Channel()
	if err != nil {
		return client, err
	}
	_, err = client.channel.QueueDeclare(
		"sms_messages", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	return client, err
}

func (c *Client) Destroy() {
	c.connection.Close()
}

type Message struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

func (c *Client) SendMessage(message Message) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = c.channel.Publish(
		"",             // exchange
		"sms_messages", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bytes,
		})
	return err
}
