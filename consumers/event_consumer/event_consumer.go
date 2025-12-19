package event_consumer

import "narasla_bot/events"

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	bathSize  int
}
