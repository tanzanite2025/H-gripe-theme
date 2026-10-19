package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/repository"
)

var ErrPostMediaReferenceInvalid = repository.ErrPostMediaReferenceInvalid

func (s *PostService) ConfigureMediaService(assets *MediaService) {
	if s != nil && s.postRepo != nil {
		s.postRepo.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	}
}

func mapPostRepositoryWriteErrorToServiceError(err error) error {
	if repository.IsRecordNotFound(err) {
		return ErrPostNotFound
	}
	if errors.Is(err, repository.ErrPostCategoryRelationshipInvalid) {
		return fmt.Errorf("%w: %v", ErrBlogCategoryInvalid, err)
	}
	return err
}
