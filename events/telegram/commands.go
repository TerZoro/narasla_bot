package telegram

import (
	"context"
	"errors"
	"log"
	"narasla_bot/clients/telegram"
	"narasla_bot/lib/e"
	"narasla_bot/storage"
	"net/url"
	"strings"
	"time"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (p *Processor) doCmd(ctx context.Context, text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Commands: got new message '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(ctx, chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(ctx, chatID, username)
	case HelpCmd:
		return p.sendHelp(ctx, chatID)
	case StartCmd:
		return p.sendHello(ctx, chatID)
	default:
		return p.tg.SendMessage(ctx, chatID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(ctx context.Context, chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.Wrap("Commands: can't do savePage", err) }()

	sendMsg := newMessageSender(ctx, chatID, p.tg)

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(ctx, page)
	if err != nil {
		return err
	}

	if isExists {
		return sendMsg(msgAlreadyExists)
	}

	if err := p.storage.Save(ctx, page); err != nil {
		return err
	}

	if err := sendMsg(msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(ctx context.Context, chatID int, username string) (err error) {
	defer func() { err = e.Wrap("Commands: can't do sendRandom", err) }()

	sendMsg := newMessageSender(ctx, chatID, p.tg)

	randPage, err := p.storage.PickRandom(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedPages) {
			return sendMsg(msgNoSavedPages)
		}

		return err
	}

	if randPage == nil {
		return sendMsg(msgNoSavedPages)
	}

	if err := p.tg.SendMessage(ctx, chatID, randPage.URL); err != nil {
		return err
	}

	return p.storage.Remove(ctx, randPage)
}

func (p *Processor) sendHello(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHello)
}

func (p *Processor) sendHelp(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHelp)
}

// using wrapper reduces redundant usage of chatID in savePage func, and makes code more readable.
func newMessageSender(ctx context.Context, chatID int, tgClient *telegram.Client) func(string) error {
	return func(msg string) error {
		return tgClient.SendMessage(ctx, chatID, msg)
	}
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
