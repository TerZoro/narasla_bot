package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgClient "narasla_bot/clients/telegram"
	"narasla_bot/consumers/event_consumer"
	"narasla_bot/events/telegram"
	"narasla_bot/sqlite"

	"github.com/joho/godotenv"
)

// temporary
const (
	tgBotHost         = "api.telegram.org"
	sqliteStoragePath = "data/sqlite/storage.db"
	batchSize         = 100
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: %w", err)
	}

	s, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatalf("can't connect to the storage: %v", err)
	}

	if err := s.Init(ctx); err != nil {
		log.Fatalf("can't init sql storage: %v", err)
	}

	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, getTgToken()),
		s,
	)

	log.Print("Server is running")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal("servise is stopped", err)
	}
}

func getTgToken() string {
	return os.Getenv("TG_BOT_TOKEN")
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
