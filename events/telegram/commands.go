package telegram

import (
	"log"
	"strings"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new message '%s' from '%s'", text, username)
}
