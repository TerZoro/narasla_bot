package main

import (
	"flag"
	"log"

	"narasla_bot/clients/telegram"
)

// temporary
const (
	tgBotHost = "api.telegram.org"
)

func main() {
	tgClient := telegram.New(tgBotHost, mustToken())
}

func mustToken() string {
	tok := flag.String(
		"telegram-bot-token",
		"",
		"token for access to telegram bot",
	)

	flag.Parse()

	if *tok == "" {
		log.Fatal("token is not specified")
	}

	return *tok
}
