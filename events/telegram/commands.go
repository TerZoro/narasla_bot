package telegram

import (
	"log"
	"narasla_bot/lib/e"
	"narasla_bot/storage"
	"net/url"
	"strings"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new message '%s' from '%s'", text, username)

	if isAddCmd(text) {
		// TODO: add page
	}

	switch text {
	case RndCmd:
	case HelpCmd:
	case StartCmd:
	default:

	}
}

func (p *Processor) SavePage(chatID int, pageUrl string, username string) (err error) {
	defer func() { err = e.Wrap("can't do command: sage page", err) }()

	page := &storage.Page{
		URL:      pageUrl,
		UserName: username,
	}
	isExists, err := p.storage.IsExists(page)
	if err != nil {
		return err
	}
	// TODO: Complete this one

}

func isAddCmd(text string) bool {
	u, err := url.Parse(text)
	return err != nil && u.Host != ""
}
