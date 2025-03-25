package e

import (
	"errors"
	"fmt"
)

var ErrSaveUniqueViolation = errors.New("index uniqueness conflict")

func WrapError(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func WrapIfError(msg string, err error) error {
	if err == nil {
		return nil
	}
	return WrapError(msg, err)
}
