package main

import (
	"encoding/json"
	"log"
	"notifier/mqclient"
	"notifier/seclient"
)

type notifyMessage struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

func main() {
	sc, err := seclient.CreateSeClient()
	if err != nil {
		log.Panic(err)
	}
	cl, err := mqclient.CreateMqClient()
	if err != nil {
		log.Panic(err)
	}
	msgs, err := cl.GetMessages()
	if err != nil {
		log.Panic(err)
	}
	log.Println("Notifier service started")
	for m := range msgs {
		var nm notifyMessage
		json.Unmarshal(m.Body, &nm)
		err := sc.Send(nm.To, nm.Message)
		if err != nil {
			log.Panic(err)
		}
		err = m.Ack(false)
		if err != nil {
			log.Panic(err)
		}
	}
}
