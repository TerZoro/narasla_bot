package e

import "fmt"

func Wrap(msg string, error error) error {
	if error == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", msg, error)
}
