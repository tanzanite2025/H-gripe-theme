package service

import (
	"commerce-platform/internal/repository"
	"errors"
)

var (
	ErrUnsupportedLocale            = errors.New("unsupported locale")
	ErrPostLocaleImmutable          = errors.New("post locale cannot be changed after creation")
	ErrFAQNotFound                  = errors.New("faq not found")
	ErrFAQLocaleImmutable           = errors.New("FAQ locale cannot be changed after creation")
	ErrFAQCategoryIdentityImmutable = errors.New("FAQ category page, locale, and key cannot be changed after creation")
	ErrGalleryNotFound              = errors.New("gallery not found")
	ErrPaymentNotFound              = errors.New("payment resource not found")
	ErrShippingNotFound             = errors.New("shipping resource not found")
)

func IsRecordNotFound(err error) bool {
	return repository.IsRecordNotFound(err)
}
