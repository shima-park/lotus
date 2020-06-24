package pipeline

import (
	"strings"

	"github.com/pkg/errors"
)

type ErrorGroup []string

func (eg ErrorGroup) Error() error {
	if len(eg) == 0 {
		return nil
	}

	return errors.New(strings.Join(eg, ""))
}
