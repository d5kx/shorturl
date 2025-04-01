package e

import (
	"errors"
	"fmt"
)

var (
	ErrSaveUniqueViolation = errors.New("index uniqueness conflict")
	ErrAuthTokenNotValid   = errors.New("auth token not valid")
	ErrUnexpSigningMethod  = errors.New("unexpected signing method")
	ErrAuthTokenParse      = errors.New("auth token parsing error")
)

func WrapError(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func WrapIfError(msg string, err error) error {
	if err == nil {
		return nil
	}
	return WrapError(msg, err)
}
