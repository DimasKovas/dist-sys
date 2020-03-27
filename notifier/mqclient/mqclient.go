package mqclient

import (
	"os"

	"github.com/streadway/amqp"
)

type Client struct {
	connection *amqp.Connection
}

func CreateMqClient() (Client, error) {
	var client Client
	var err error
	client.connection, err = amqp.Dial(os.Getenv("MESSAGE_QUEUE_URL"))
	return client, err
}

func (c *Client) Destroy() {
	c.connection.Close()
}

func (c *Client) GetMessages() (<-chan amqp.Delivery, error) {
	ch, err := c.connection.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(
		"email_messages", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	return msgs, err
}
