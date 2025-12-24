package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"narasla_bot/storage"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed queries/save.sql
var qSave string

//go:embed queries/pick_random.sql
var qPickRandom string

//go:embed queries/remove.sql
var qRemove string

//go:embed queries/is_exists.sql
var qIsExists string

//go:embed queries/init.sql
var qInit string

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, page *storage.Page) error {
	if _, err := s.db.ExecContext(ctx, qSave, page.URL, page.UserName); err != nil {
		return fmt.Errorf("can't save page: %w", err)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, userName string) (*storage.Page, error) {
	var url string

	err := s.db.QueryRowContext(ctx, qPickRandom, userName).Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, fmt.Errorf("can't get random page: %w", err)
	}

	return &storage.Page{
		URL:      url,
		UserName: userName,
	}, nil
}

func (s *Storage) Remove(ctx context.Context, page *storage.Page) error {
	if _, err := s.db.ExecContext(ctx, qRemove, page.URL, page.UserName); err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	return nil
}

// IsExists checks if page exists in storage.
func (s *Storage) IsExists(ctx context.Context, page *storage.Page) (bool, error) {
	var count int

	if err := s.db.QueryRowContext(ctx, qIsExists, page.URL, page.UserName).Scan(&count); err != nil {
		return false, fmt.Errorf("can't check page exists: %w", err)
	}

	return count > 0, nil
}

func (s *Storage) Init(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, qInit); err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}

	return nil
}
