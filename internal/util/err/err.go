package err

import (
	"fmt"
)

func WrapError(msg string, err error) error {
	return fmt.Errorf("s%: %w", msg, err)
}

func WrapIfError(msg string, err error) error {
	if err == nil {
		return nil
	}
	return WrapError(msg, err)
}
