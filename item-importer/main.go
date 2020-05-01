package main

import (
	"common/dbclient"
	"encoding/json"
	"item-importer/mqclient"
	"log"
)

func main() {
	db, err := dbclient.CreateDbClient()
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
	log.Println("Item-importer service started")
	for m := range msgs {
		var batch []dbclient.Item
		json.Unmarshal(m.Body, &batch)
		err := db.ImportItemBatch(batch)
		if err != nil {
			log.Panic(err)
		}
		err = m.Ack(false)
		if err != nil {
			log.Panic(err)
		}
	}
}
