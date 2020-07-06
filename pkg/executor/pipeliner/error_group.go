package pipeliner

import (
	"strings"

	"github.com/pkg/errors"
)

type ErrorGroup []error

func (eg ErrorGroup) Error() error {
	if len(eg) == 0 {
		return nil
	}

	var errs []string
	for _, err := range eg {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	return errors.New(strings.Join(errs, ", "))
}
