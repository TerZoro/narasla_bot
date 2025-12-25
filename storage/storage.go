package storage

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"narasla_bot/lib/e"
	"time"
)

// TODO: implement new fields
type Storage interface {
	Save(ctx context.Context, p *Page) error
	PickRandom(ctx context.Context, userName string) (*Page, error)
	Remove(ctx context.Context, p *Page) error
	RemoveByURL(ctx context.Context, username, url string) error
	List(ctx context.Context, username string, limit, offset int) ([]Page, error)
	Count(ctx context.Context, username string) (int, error)
	IsExists(ctx context.Context, p *Page) (bool, error)
}

var (
	ErrNoSavedPages = errors.New("Storage: no saved pages")
	ErrNotFound     = errors.New("Storage: page not found")
)

type Page struct {
	ID        int
	URL       string
	UserName  string
	CreatedAt time.Time
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
