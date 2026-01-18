package scheduler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
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

	if !u.LastSendAt.Valid {
		return true, nil
	}

	last := time.Unix(u.LastSendAt.Int64, 0).In(loc)
	if alrSendToday(last, nowLocal) {
		return false, nil
	}

	timeToSend := time.Date(
		nowLocal.Year(), nowLocal.Month(), nowLocal.Day(),
		u.SendHour, u.SendMinute, 0, 0,
		loc,
	)

	return nowLocal.After(timeToSend), nil
}

func (s *Scheduler) sendOne(ctx context.Context, u storage.User, now time.Time) error {
	page, err := s.st.PickRandom(ctx, u.OwnerID)
	if err != nil {
		return err
	}

	// hardcoded: u.ChatID if you want scheduler to send only in private.
	// rn, it will send to the last chatID whether it is Group of Private.
	if err := s.tg.SendMessage(ctx, page.ChatID, page.URL); err != nil {
		return err
	}

	if err := s.st.Remove(ctx, page); err != nil {
		return err
	}

	newHour := 9 + rand.Intn(15) // from 9 to 23
	newMinute := rand.Intn(60)

	return s.st.UpdateLastSendAt(ctx, u.OwnerID, now.Unix(), newHour, newMinute)
}

func alrSendToday(first, last time.Time) bool {
	firstY, firstM, firstD := first.Date()
	lastY, lastM, lastD := last.Date()

	return firstY == lastY && firstM == lastM && firstD == lastD
}
