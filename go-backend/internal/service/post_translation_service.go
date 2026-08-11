package service

import (
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/repository"
)

func (s *PostService) GetTranslations(postID uint) ([]post.Post, error) {
	groupID, err := s.postRepo.GetTranslationGroupID(postID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	if groupID == nil {
		return []post.Post{}, nil
	}

	return s.postRepo.FindByTranslationGroup(*groupID)
}

func (s *PostService) GetPublicTranslations(postID uint) ([]post.Post, error) {
	sourcePost, err := s.postRepo.FindByID(postID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	if sourcePost.Status != "published" {
		return nil, ErrPostNotFound
	}
	if sourcePost.TranslationGroupID == nil {
		return []post.Post{}, nil
	}

	return s.postRepo.FindPublishedByTranslationGroup(*sourcePost.TranslationGroupID)
}

func (s *PostService) GetTranslationsByGroup(groupID uint) ([]post.Post, error) {
	return s.postRepo.FindByTranslationGroup(groupID)
}
