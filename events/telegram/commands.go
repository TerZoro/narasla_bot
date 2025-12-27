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
	SaveCmd   = "/save"
	RndCmd    = "/rnd"
	HelpCmd   = "/help"
	StartCmd  = "/start"
	DeleteCmd = "/del"
	ListCmd   = "/list"
)

func (p *Processor) doCmd(ctx context.Context, text string, m Meta) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	log.Printf("Chat Type: %s", m.Chat.Type)

	if m.Chat.Type == "private" && isAddCmd(text) {
		return p.savePage(ctx, m.Chat.ID, m.UserID, text, m.Username)
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil
	}
	cmdRaw := parts[0]

	if !strings.HasPrefix(cmdRaw, "/") {
		return nil
	}

	arg := ""
	if len(parts) > 1 {
		arg = strings.Join(parts[1:], " ")
	}

	cmd, ok := p.resolveCmd(cmdRaw, m.Chat.Type)
	if !ok {
		return nil
	}

	log.Printf("Commands: got new message '{\n%s\n}' from '%s'", text, m.Username)

	return p.middleHandler(ctx, cmd, arg, m)
}

func (p *Processor) resolveCmd(cmdRaw, chatType string) (string, bool) {
	isPrivate := chatType == "private"
	mentionBot := strings.Contains(cmdRaw, "@")

	if !isPrivate && !mentionBot {
		return "", false
	}

	if !mentionBot {
		return cmdRaw, true
	}

	return p.normalizeCmd(cmdRaw)
}

func (p *Processor) normalizeCmd(cmdRaw string) (string, bool) {
	cmd := cmdRaw
	if i := strings.IndexByte(cmdRaw, '@'); i != -1 {
		base := cmdRaw[:i]
		bot := cmdRaw[i+1:]

		if p.botUsername != "" && !strings.EqualFold(bot, p.botUsername) {
			return "", false
		}
		cmd = base
	}

	return cmd, true
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
