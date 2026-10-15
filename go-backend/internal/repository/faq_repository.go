package repository

import "gorm.io/gorm"

type FAQRepository struct {
	db                             *gorm.DB
	businessMediaReferenceResolver PublicBusinessMediaReferenceObjectKeyResolver
}

func NewFAQRepository(db *gorm.DB) *FAQRepository {
	return &FAQRepository{db: db}
}

func (r *FAQRepository) ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver PublicBusinessMediaReferenceObjectKeyResolver) {
	r.businessMediaReferenceResolver = resolver
}
