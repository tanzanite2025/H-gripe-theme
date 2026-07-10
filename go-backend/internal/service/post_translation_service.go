package service

import (
	"tanzanite/internal/domain/post"
	"tanzanite/internal/repository"
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

func (s *PostService) GetTranslationsByGroup(groupID uint) ([]post.Post, error) {
	return s.postRepo.FindByTranslationGroup(groupID)
}
