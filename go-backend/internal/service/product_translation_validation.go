package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/repository"
)

func (s *ProductService) validateProductTranslationParent(parentID *uint, productID uint, locale string) error {
	if parentID == nil || *parentID == 0 {
		return nil
	}
	if *parentID == productID {
		return fmt.Errorf("%w: a product cannot translate itself", ErrProductTranslationInvalid)
	}

	parent, err := s.productRepo.FindTranslationParent(*parentID)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return fmt.Errorf("%w: translation parent not found", ErrProductTranslationInvalid)
	}
	if err != nil {
		return err
	}
	if parent.ParentID != nil && *parent.ParentID > 0 {
		return fmt.Errorf("%w: translations must point to a root product", ErrProductTranslationInvalid)
	}
	if parent.Locale == locale {
		return fmt.Errorf("%w: translation locale must differ from the root product locale", ErrProductTranslationInvalid)
	}

	exists, err := s.productRepo.TranslationLocaleExists(*parentID, locale, productID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: translation locale already exists in this product group", ErrProductTranslationInvalid)
	}
	return nil
}
