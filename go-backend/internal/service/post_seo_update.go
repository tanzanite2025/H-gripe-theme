package service

import "tanzanite/internal/domain/post"

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
	if input.MetaTitle != nil {
		existingPost.MetaTitle = *input.MetaTitle
	}
	if input.MetaDescription != nil {
		existingPost.MetaDesc = *input.MetaDescription
	}
	if input.CanonicalURL != nil {
		existingPost.CanonicalURL = *input.CanonicalURL
	}
	if err := s.postRepo.Update(existingPost); err != nil {
		return nil, err
	}

	s.clearPostCache(existingPost)
	s.invalidateStorefrontHTMLCache("post SEO update")
	return s.findPost(id)
}
