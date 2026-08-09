package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

var (
	ErrProductMediaReferenceInvalid              = errors.New("product media reference invalid")
	ErrProductVariantOptionValueReferenceInvalid = errors.New("product variant option value reference invalid")
)

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsDuplicatedKey(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
