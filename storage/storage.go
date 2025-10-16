package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"narasla_bot/lib/e"
)

type Storage interface {
	Save(p *Page) error
	PickRandom(userName string) (*Page, error)
	Remove(p *Page)
	isExists(p *Page) (bool, error)
}

type Page struct {
	URL      string
	UserName string
	// Created time.Time
}

func (p *Page) Hash() (string, error) {
	h := sha256.New()

	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", e.Wrap("storage: Page Hash failed", err)
	}

	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", e.Wrap("storage: Page Hash failed", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
