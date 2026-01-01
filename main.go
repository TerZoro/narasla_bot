package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgClient "narasla_bot/clients/telegram"
	"narasla_bot/consumers/event_consumer"
	"narasla_bot/events/telegram"
	"narasla_bot/scheduler"
	"narasla_bot/sqlite"

	"github.com/joho/godotenv"
)

const (
	tgBotHost = "api.telegram.org"
	batchSize = 100
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: %w", err)
	}

	storagePath := getStoragePath()
	s, err := sqlite.New(storagePath)
	if err != nil {
		log.Fatalf("can't connect to the storage: %v", err)
	}

	if err := s.Init(ctx); err != nil {
		log.Fatalf("can't init sql storage: %v", err)
	}

	tgCl := tgClient.New(tgBotHost, getTgToken())
	botUsername := getBotUsername()

	eventsProcessor := telegram.New(
		tgCl,
		s,
		botUsername,
	)

	sch := scheduler.New(s, tgCl, 1*time.Minute)
	go func() {
		if err := sch.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("scheduler stopped: %v", err)
		}
	}()

	log.Print("Server is running")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal("servise is stopped", err)
	}
}

func getTgToken() string {
	return os.Getenv("TG_BOT_TOKEN")
}

func getStoragePath() string {
	return os.Getenv("STORAGE_PATH")
}

func getBotUsername() string {
	return os.Getenv("BOT_USERNAME")
}
