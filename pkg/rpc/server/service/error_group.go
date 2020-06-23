package service

import (
	"errors"
	"strings"
)

type ErrorGroup []string

func (eg ErrorGroup) Error() error {
	if len(eg) == 0 {
		return nil
	}

	return errors.New(strings.Join(eg, ""))
}
