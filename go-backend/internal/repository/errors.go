package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsDuplicatedKey(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
