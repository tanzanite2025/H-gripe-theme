package service

import "commerce-platform/internal/domain/post"

// PostSEOUpdateInput is the only article Meta write contract exposed to the
// SEO control plane. Content create/update inputs deliberately omit these
// fields so article editing cannot become a second SEO entry point.
type PostSEOUpdateInput struct {
	MetaTitle       *string
	MetaDescription *string
	CanonicalURL    *string
}

func (s *PostService) UpdatePostSEO(id uint, input PostSEOUpdateInput) (*post.Post, error) {
	if s == nil {
		return nil, ErrPostNotFound
	}

	existingPost, err := s.findPost(id)
	if err != nil {
		return nil, err
	}
	if err := s.postRepo.UpdatePostSEOMetadataFields(id, input.MetaTitle, input.MetaDescription, input.CanonicalURL); err != nil {
		return nil, mapPostRepositoryWriteErrorToServiceError(err)
	}

	s.clearPostCache(existingPost)
	s.invalidateStorefrontHTMLCache("post SEO update")
	return s.findPost(id)
}
