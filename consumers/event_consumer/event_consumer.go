package event_consumer

import (
	"errors"
	"log"
	"narasla_bot/events"
	tgEvents "narasla_bot/events/telegram"
	"narasla_bot/storage"
	"sync"
	"time"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	bathSize  int
}

const (
	trials    = 10
	baseDelay = 100 * time.Millisecond
)

func New(fetcher events.Fetcher, processor events.Processor, bathsize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		bathSize:  bathsize,
	}
}

func (c Consumer) Start() error {
	for {
		gotEvents, err := fetchWithRetry(c.fetcher, c.bathSize, trials)
		if err != nil {
			log.Printf("consumer: fetch failed after retries: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		if err := c.handleEvents(gotEvents); err != nil {
			log.Printf("consumer: handleEvents: %v", err)
			continue
		}

	}
}

func (c *Consumer) handleEvents(batch []events.Event) error {
	var wg sync.WaitGroup

	for _, ev := range batch {
		wg.Add(1)
		go func(ev events.Event) {
			defer wg.Done()

			log.Printf("got new event: %s", ev.Text)

			if err := processWithRetry(c.processor, ev, trials); err != nil {
				log.Printf("event failed: %v", err)
			}
		}(ev)
	}

	wg.Wait()

	return nil
}

func fetchWithRetry(fetcher events.Fetcher, batchSize, attempts int) ([]events.Event, error) {
	var lastErr error
	delay := baseDelay

	for i := 0; i < attempts; i++ {
		events, err := fetcher.Fetch(batchSize)
		if err == nil {
			return events, nil
		}

		lastErr = err

		if i < attempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}

	return nil, lastErr
}

func processWithRetry(p events.Processor, ev events.Event, attempts int) error {
	var lastErr error
	delay := baseDelay

	for i := 0; i < attempts; i++ {
		err := p.Process(ev)
		if err == nil {
			return nil
		}

		// if it's not retriable, we will stop.
		if !shouldRetry(err) {
			return err
		}

		lastErr = err
		log.Printf("retrying process (%d/%d): %v", i+1, attempts, err)

		if i < attempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}

	return lastErr
}

func shouldRetry(err error) bool {
	// here: retrying is pointless
	if errors.Is(err, storage.ErrNoSavedPages) {
		return false
	}

	// if it is message update, just skip it
	if errors.Is(err, tgEvents.ErrorUnknownEventType) {
		return false
	}

	return true
}
