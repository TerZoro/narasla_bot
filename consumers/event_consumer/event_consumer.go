package event_consumer

import (
	"context"
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
	trials    = 8
	baseDelay = 100 * time.Millisecond
)

func New(fetcher events.Fetcher, processor events.Processor, bathsize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		bathSize:  bathsize,
	}
}

func (c Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		gotEvents, err := fetchWithRetry(ctx, c.fetcher, c.bathSize, trials)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			log.Printf("consumer: fetch failed after retries: %v", err)
			if sleepCtx(ctx, 1*time.Second) != nil {
				return ctx.Err()
			}
			continue
		}

		if len(gotEvents) == 0 {
			if sleepCtx(ctx, 1*time.Second) != nil {
				return ctx.Err()
			}
			continue
		}

		if err := c.handleEvents(ctx, gotEvents); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			log.Printf("consumer: handleEvents: %v", err)
			continue
		}

	}
}

func (c *Consumer) handleEvents(ctx context.Context, batch []events.Event) error {
	var wg sync.WaitGroup

	for _, ev := range batch {
		wg.Add(1)
		go func(ev events.Event) {
			defer wg.Done()

			// if we already stopped, don't continue
			if ctx.Err() != nil {
				return
			}

			log.Printf("got new event: %s", ev.Text)

			if err := processWithRetry(ctx, c.processor, ev, trials); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					log.Printf("event failed: %v", err)
				}
			}
		}(ev)
	}

	wg.Wait()

	return nil
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func fetchWithRetry(ctx context.Context, fetcher events.Fetcher, batchSize, attempts int) ([]events.Event, error) {
	var lastErr error
	delay := baseDelay

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		events, err := fetcher.Fetch(ctx, batchSize)
		if err == nil {
			return events, nil
		}

		lastErr = err

		if i < attempts-1 {
			if err := sleepCtx(ctx, delay); err != nil {
				return nil, err
			}
			delay *= 2
		}
	}

	return nil, lastErr
}

func processWithRetry(ctx context.Context, p events.Processor, ev events.Event, attempts int) error {
	var lastErr error
	delay := baseDelay

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := p.Process(ctx, ev)
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
			if err := sleepCtx(ctx, delay); err != nil {
				return err
			}
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
