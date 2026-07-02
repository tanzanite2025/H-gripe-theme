package service

import "tanzanite/internal/repository"

func IsRecordNotFound(err error) bool {
	return repository.IsRecordNotFound(err)
}
