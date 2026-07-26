package service

import (
	"errors"
	"tanzanite/internal/repository"
)

var ErrUnsupportedLocale = errors.New("unsupported locale")

func IsRecordNotFound(err error) bool {
	return repository.IsRecordNotFound(err)
}
