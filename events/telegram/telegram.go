package telegram

import (
	"context"
	"errors"
	"narasla_bot/clients/telegram"
	"narasla_bot/events"
	"narasla_bot/lib/e"

	"narasla_bot/storage"
)

var (
	ErrorUnknownEventType = errors.New("events: unknown event type")
	ErrorUnknownMetaType  = errors.New("events: unknown meta type")
)

type Processor struct {
	tg          *telegram.Client
	offset      int
	storage     storage.Storage // interface
	botUsername string
	handlers    map[string]handler
}

// now we implement Meta interface exclusively for telegram
type Meta struct {
	Chat     Chat
	UserID   int64
	Username string
}

type Chat struct {
	ID   int64
	Type string // "private", "group", "supergroup", or "channel"
}

type handler func(ctx context.Context, arg string, m Meta) error

func New(tg *telegram.Client, st storage.Storage, botUsername string) *Processor {
	p := &Processor{
		tg:          tg,
		storage:     st,
		botUsername: botUsername,
	}
	p.initHandlers()
	return p
}

func (p *Processor) Fetch(ctx context.Context, limit int) ([]events.Event, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	updates, err := p.tg.Updates(ctx, p.offset, limit)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, e.Wrap("Events: telegram Fetch failed to get context", err)
	}

	if err != nil {
		return nil, e.Wrap("Events: telegram Fetch failed to get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(ctx context.Context, event events.Event) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	switch event.Type {
	case events.Message:
		return p.processMessage(ctx, event)
	case events.Unknown:
		return nil
	default:
		return e.Wrap("Events: Process failed to process message", ErrorUnknownEventType)
	}
}

func (p *Processor) processMessage(ctx context.Context, event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("Events: processMessage failed to process message", err)
	}

	if err := p.doCmd(ctx, event.Text, meta); err != nil {
		return e.Wrap("Events: processMessage failed to process message", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta) // this call is type assertion
	if !ok {
		return Meta{},
			e.Wrap("Events: processMessage failed to get meta", ErrorUnknownMetaType)
	}

	return res, nil
}

func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)
	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			Chat:     getChatData(upd),
			UserID:   upd.Message.From.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res
}

func getChatData(upd telegram.Update) Chat {
	return Chat{
		ID:   upd.Message.Chat.ID,
		Type: upd.Message.Chat.Type,
	}
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}

	return events.Message
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}

	return upd.Message.Text
}
