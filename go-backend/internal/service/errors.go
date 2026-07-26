package service

import (
	"errors"
	"tanzanite/internal/repository"
)

var (
	ErrUnsupportedLocale            = errors.New("unsupported locale")
	ErrFAQLocaleImmutable           = errors.New("FAQ locale cannot be changed after creation")
	ErrFAQCategoryIdentityImmutable = errors.New("FAQ category page, locale, and key cannot be changed after creation")
)

func IsRecordNotFound(err error) bool {
	return repository.IsRecordNotFound(err)
}
