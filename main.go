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
	if os.Getenv("TG_BOT_TOKEN") == "" &&
		os.Getenv("BOT_USERNAME") == "" &&
		os.Getenv("STORAGE_PATH") == "" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
	}

	tgToken := mustEnv("TG_BOT_TOKEN")
	storagePath := mustEnv("STORAGE_PATH")
	botUsername := mustEnv("BOT_USERNAME")

	s, err := sqlite.New(storagePath)
	if err != nil {
		log.Fatalf("can't connect to the storage: %v", err)
	}

	if err := s.Init(ctx); err != nil {
		log.Fatalf("can't init sql storage: %v", err)
	}

	tgCl := tgClient.New(tgBotHost, tgToken)

	eventsProcessor := telegram.New(
		tgCl,
		s,
		botUsername,
	)

	sch := scheduler.New(s, tgCl, 10*time.Minute)
	go func() {
		if err := sch.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("scheduler stopped: %v", err)
		}
	}()

	log.Print("Server is running")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("servise is stopped: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("no %s key found in .env file", key)
	}

	return v
}
