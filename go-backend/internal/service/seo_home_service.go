package service

import (
	seodomain "tanzanite/internal/domain/seo"
)

func (s *SEOService) GetHome(locale string) (*seodomain.Settings, error) {
	return s.getHome(locale)
}

func (s *SEOService) UpdateHome(request seodomain.UpdateRequest) (*seodomain.Settings, error) {
	return s.updateHome(request)
}
