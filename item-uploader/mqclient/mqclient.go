package mqclient

import (
	"encoding/json"
	"os"

	"common/dbclient"

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
		"batch_import", // name
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

func (c *Client) SendImportBatch(batch []dbclient.Item) error {
	bytes, err := json.Marshal(batch)
	if err != nil {
		return err
	}
	err = c.channel.Publish(
		"",             // exchange
		"batch_import", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bytes,
		})
	return err
}
