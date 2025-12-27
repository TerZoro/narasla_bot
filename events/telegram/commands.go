package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"narasla_bot/clients/telegram"
	"narasla_bot/lib/e"
	"narasla_bot/storage"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const limit = 20

const (
	RndCmd    = "/rnd"
	HelpCmd   = "/help"
	StartCmd  = "/start"
	DeleteCmd = "/del"
	ListCmd   = "/list"
)

func (p *Processor) doCmd(ctx context.Context, text string, chatID, userID int64, username string) error {
	text = strings.TrimSpace(text)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Commands: got new message '{\n%s\n}' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(ctx, chatID, userID, text, username)
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil
	}
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.Join(parts[1:], " ")
	}

	switch cmd {
	case RndCmd:
		return p.sendRandom(ctx, chatID, userID)
	case HelpCmd:
		return p.sendHelp(ctx, chatID, userID)
	case StartCmd:
		return p.sendHello(ctx, chatID, userID)
	case DeleteCmd:
		return p.removePage(ctx, chatID, userID, username, arg)
	case ListCmd:
		return p.sendList(ctx, chatID, userID, username)
	default:
		return p.tg.SendMessage(ctx, chatID, userID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(ctx context.Context, chatID, userID int64, pageURL, username string) (err error) {
	defer func() { err = e.Wrap("Commands: can't do savePage", err) }()

	sendMsg := newMessageSender(ctx, chatID, userID, p.tg)

	page := &storage.Page{
		URL:      pageURL,
		OwnerID:  userID,
		ChatID:   chatID,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(ctx, userID, pageURL)
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

func (p *Processor) sendRandom(ctx context.Context, chatID, userID int64) (err error) {
	defer func() { err = e.Wrap("Commands: can't do sendRandom", err) }()

	sendMsg := newMessageSender(ctx, chatID, userID, p.tg)

	randPage, err := p.storage.PickRandom(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedPages) {
			return sendMsg(msgNoSavedPages)
		}

		return err
	}

	if randPage == nil {
		return sendMsg(msgNoSavedPages)
	}

	if err := p.tg.SendMessage(ctx, chatID, userID, randPage.URL); err != nil {
		return err
	}

	return p.storage.Remove(ctx, randPage)
}

func (p *Processor) sendHello(ctx context.Context, chatID, userID int64) error {
	return p.tg.SendMessage(ctx, chatID, userID, msgHello)
}

func (p *Processor) sendHelp(ctx context.Context, chatID, userID int64) error {
	return p.tg.SendMessage(ctx, chatID, userID, msgHelp)
}

func (p *Processor) removePage(ctx context.Context, chatID, userID int64, username, arg string) (err error) {
	defer func() { err = e.Wrap("Commands: can't delete page", err) }()

	sendMsg := newMessageSender(ctx, chatID, userID, p.tg)
	arg = strings.TrimSpace(arg)

	if arg == "" {
		return p.sendList(ctx, chatID, userID, username)
	}

	if isURL(arg) {
		if err := p.storage.RemoveByURL(ctx, userID, arg); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return sendMsg(msgNoSavedPages)
			}

			return err
		}
		return sendMsg(msgDeleted)
	}

	num, err := strconv.Atoi(arg)
	if err != nil || num <= 0 {
		return sendMsg(msgIncorrectDeleteArg)
	}

	list, err := p.storage.List(ctx, userID, username, limit, 0)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return sendMsg(msgNoSavedPages)
	}

	if num > len(list) {
		return sendMsg(
			fmt.Sprintf("You have only %d items in the list. Send /del to see them.", len(list)))
	}

	page := list[num-1]
	if err := p.storage.Remove(ctx, &page); err != nil {
		return err
	}

	return sendMsg(msgDeleted)
}

func (p *Processor) sendList(ctx context.Context, chatID, userID int64, username string) (err error) {
	defer func() { err = e.Wrap("Command: can't send list", err) }()

	sendMsg := newMessageSender(ctx, chatID, userID, p.tg)

	list, err := p.storage.List(ctx, userID, username, limit, 0)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return sendMsg(msgNoSavedPages)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s 's saved pages:\n\n", username))

	for i, p := range list {
		sb.WriteString(fmt.Sprintf("%d. — %s\n", i+1, p.URL))
	}

	sb.WriteString("\nDelete: /del <number> or /del <url>")

	return sendMsg(sb.String())
}

// using wrapper reduces redundant usage of chatID in savePage func, and makes code more readable.
func newMessageSender(ctx context.Context, chatID, userID int64, tgClient *telegram.Client) func(string) error {
	return func(msg string) error {
		return tgClient.SendMessage(ctx, chatID, userID, msg)
	}
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
