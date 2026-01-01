package scheduler

import (
	"context"
	"errors"
	"fmt"
	"narasla_bot/storage"
	"time"
)

func (s *Scheduler) step(ctx context.Context) error {
	users, err := s.st.ListEnabledUsers(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	for _, u := range users {
		ok, err := s.shouldSendNow(u, now)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		if err := s.sendOne(ctx, u, now); err != nil {
			if !errors.Is(err, storage.ErrNoSavedPages) {
				return fmt.Errorf("scheduler: sendOne failed owner=%d: %w", u.OwnerID, err)
			}
		}
	}

	return nil
}

func (s *Scheduler) shouldSendNow(u storage.User, now time.Time) (bool, error) {
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil {
		loc = time.UTC
	}

	nowLocal := now.In(loc)

	timeToSend := time.Date(
		nowLocal.Year(), nowLocal.Month(), nowLocal.Day(),
		u.SendHour, u.SendMinute, 0, 0,
		loc,
	)

	if nowLocal.Before(timeToSend) {
		return false, nil
	}

	if !u.LastSendAt.Valid {
		return true, nil
	}

	last := time.Unix(u.LastSendAt.Int64, 0).In(loc)
	if alrSendToday(last, nowLocal) {
		return false, nil
	}

	return true, nil
}

func (s *Scheduler) sendOne(ctx context.Context, u storage.User, now time.Time) error {
	page, err := s.st.PickRandom(ctx, u.OwnerID)
	if err != nil {
		return err
	}

	if err := s.tg.SendMessage(ctx, u.ChatID, page.URL); err != nil {
		return err
	}

	if err := s.st.Remove(ctx, page); err != nil {
		return err
	}

	return s.st.UpdateLastSendAt(ctx, u.OwnerID, now.Unix())
}

func alrSendToday(first, last time.Time) bool {
	firstY, firstM, firstD := first.Date()
	lastY, lastM, lastD := last.Date()

	return firstY == lastY && firstM == lastM && firstD == lastD
}
