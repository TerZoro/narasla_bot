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

func (s *Storage) ListEnabledUsers(ctx context.Context) ([]storage.User, error) {
	rows, err := s.db.QueryContext(ctx, qListEnabledUsers)
	if err != nil {
		return nil, fmt.Errorf("can't find enabled users: %w", err)
	}
	defer rows.Close()

	enabledUsers := make([]storage.User, 0, 256)
	for rows.Next() {
		var user storage.User
		err := rows.Scan(
			&user.OwnerID,
			&user.ChatID,
			&user.Username,
			&user.Timezone,
			&user.SendHour,
			&user.SendMinute,
			&user.LastSendAt,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan enabled users: %w", err)
		}
		if user.ChatID == 0 {
			continue
		}
		enabledUsers = append(enabledUsers, user)
	}

	if err = rows.Err(); err != nil {
		return enabledUsers, fmt.Errorf("can't get rows: %w", err)
	}

	return enabledUsers, nil
}

func (s *Storage) UpdateLastSendAt(ctx context.Context, ownerID, newTime int64) error {
	if _, err := s.db.ExecContext(ctx, qUpdateLastSendAt, newTime, ownerID); err != nil {
		return fmt.Errorf("can't update last send at for user: %w", err)
	}

	return nil
}

func (s *Storage) UpdateUserInfo(ctx context.Context, ownerID, chatID int64, username string) error {
	if _, err := s.db.ExecContext(ctx, qUpdateUserInfo, ownerID, chatID, username); err != nil {
		return fmt.Errorf("can't update user info: %w", err)
	}

	return nil
}

func (s *Storage) SwitchEnable(ctx context.Context, ownerID int64, enabled bool) error {
	enabledForm := 0
	if enabled {
		enabledForm = 1
	}
	if _, err := s.db.ExecContext(ctx, qUpdateEnabled, enabledForm, ownerID); err != nil {
		return fmt.Errorf("can't change enabled for user: %w", err)
	}

	return nil
}

func (s *Storage) GetUserInfo(ctx context.Context, ownerID int64) (*storage.User, error) {
	var (
		timezone    string
		enabledForm int
		sendHour    int
		sendMinute  int
		lastSendAt  sql.NullInt64
	)

	err := s.db.QueryRowContext(ctx, qGetUserInfo, ownerID).Scan(
		&timezone,
		&enabledForm,
		&sendHour,
		&sendMinute,
		&lastSendAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("can't get user info: %w", err)
	}

	enabled := enabledForm == 1

	return &storage.User{
		OwnerID:    ownerID,
		Timezone:   timezone,
		Enabled:    enabled,
		SendHour:   sendHour,
		SendMinute: sendMinute,
		LastSendAt: lastSendAt,
	}, nil
}
