package service

import (
	"errors"
	"strings"
)

type ErrorGroup []error

func (eg ErrorGroup) Error() error {
	if len(eg) == 0 {
		return nil
	}

	var errs []string
	for _, err := range eg {
		errs = append(errs, err.Error())
	}

	return errors.New(strings.Join(errs, ", "))
}
