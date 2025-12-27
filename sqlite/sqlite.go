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

//go:embed queries/remove_by_url.sql
var qRemoveByUrl string

//go:embed queries/list.sql
var qList string

//go:embed queries/count.sql
var qCount string

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
	if _, err := s.db.ExecContext(
		ctx,
		qSave,
		page.OwnerID,
		page.ChatID,
		page.URL,
		page.UserName,
	); err != nil {
		return fmt.Errorf("can't save page: %w", err)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, ownerID int64) (*storage.Page, error) {
	var pageID int64
	var url string

	err := s.db.QueryRowContext(ctx, qPickRandom, ownerID).Scan(&pageID, &url)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, fmt.Errorf("can't get random page: %w", err)
	}

	return &storage.Page{
		ID:      pageID,
		URL:     url,
		OwnerID: ownerID,
	}, nil
}

func (s *Storage) Remove(ctx context.Context, page *storage.Page) error {
	res, err := s.db.ExecContext(ctx, qRemove, page.OwnerID, page.ID)
	if err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (s *Storage) RemoveByURL(ctx context.Context, ownerID int64, url string) error {
	res, err := s.db.ExecContext(ctx, qRemoveByUrl, ownerID, url)
	if err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (s *Storage) List(ctx context.Context, ownerID int64, username string, limit, offset int) ([]storage.Page, error) {
	rows, err := s.db.QueryContext(ctx, qList, ownerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("can't get list: %w", err)
	}
	defer rows.Close()

	list := make([]storage.Page, 0, limit)

	for rows.Next() {
		page := storage.Page{
			OwnerID:  ownerID,
			UserName: username,
		}

		if err := rows.Scan(&page.ID, &page.URL, &page.CreatedAt); err != nil {
			return list, fmt.Errorf("can't scan page: %w", err)
		}
		list = append(list, page)
	}

	if err = rows.Err(); err != nil {
		return list, fmt.Errorf("can't get rows: %w", err)
	}

	return list, nil
}

func (s *Storage) Count(ctx context.Context, ownerID int64) (int, error) {
	var count int

	err := s.db.QueryRowContext(ctx, qCount, ownerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("can't count pages: %w", err)
	}

	return count, nil
}

// IsExists checks if page exists in storage.
func (s *Storage) IsExists(ctx context.Context, ownerID int64, url string) (bool, error) {
	var count int

	if err := s.db.QueryRowContext(ctx, qIsExists, ownerID, url).Scan(&count); err != nil {
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
