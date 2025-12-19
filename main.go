package main

import (
	"flag"
	"log"

	tgClient "narasla_bot/clients/telegram"
	"narasla_bot/consumers/event_consumer"
	"narasla_bot/events/telegram"
	"narasla_bot/storage/files"
)

// temporary
const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 100
)

func main() {
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		files.New(storagePath),
	)

	log.Print("Server is running")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := consumer.Start(); err != nil {
		log.Fatal("servise is stopped", err)
	}
}

func mustToken() string {
	tok := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)

	flag.Parse()

	if *tok == "" {
		log.Fatal("token is not specified")
	}

	return *tok
}
