package service

import (
	"errors"
	"tanzanite/internal/domain/product"
)

type ProductCreateInput struct {
	ProductTypeID *uint
	Name          string
	Slug          string
	Description   string
	ShortDesc     string
	Weight        int
	Status        string
	Locale        string
	ParentID      *uint
	Featured      bool
	MetaTitle     string
	MetaDesc      string
	SpecValues    map[string]string
	Variants      []ProductVariantInput
}

type ProductUpdateInput struct {
	ProductTypeID       *uint
	UpdateProductTypeID bool
	Name                *string
	Slug                *string
	Description         *string
	ShortDesc           *string
	Weight              *int
	Status              *string
	Locale              *string
	ParentID            *uint
	UpdateParentID      bool
	Featured            *bool
	MetaTitle           *string
	MetaDesc            *string
	SpecValues          map[string]string
	UpdateSpecValues    bool
	Variants            []ProductVariantInput
	UpdateVariants      bool
}

func (s *ProductService) ListAdmin(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	return s.productRepo.FindAllWithFilters(page, pageSize, status, locale, search, featured)
}

func (s *ProductService) GetAdminProduct(id uint) (*product.Product, error) {
	return s.findProduct(id)
}

func (s *ProductService) GetStats() (map[string]interface{}, error) {
	return s.productRepo.GetStats()
}

func (s *ProductService) CreateAdminProduct(input ProductCreateInput) (*product.Product, error) {
	specValues, err := s.buildSpecValues(input.ProductTypeID, input.SpecValues)
	if err != nil {
		return nil, err
	}

	variants, err := s.buildVariants(input.ProductTypeID, input.Variants)
	if err != nil {
		return nil, err
	}
	if err := s.ensureVariantSKUsAvailable(variants, 0); err != nil {
		return nil, err
	}

	newProduct := &product.Product{
		ProductTypeID: input.ProductTypeID,
		SKU:           defaultVariantSKU(variants),
		Name:          input.Name,
		Slug:          input.Slug,
		Description:   input.Description,
		ShortDesc:     input.ShortDesc,
		Weight:        input.Weight,
		Status:        input.Status,
		Locale:        input.Locale,
		ParentID:      input.ParentID,
		Featured:      input.Featured,
		MetaTitle:     input.MetaTitle,
		MetaDesc:      input.MetaDesc,
	}

	if err := s.productRepo.CreateWithSpecValuesAndVariants(newProduct, specValues, variants); err != nil {
		return nil, err
	}

	return s.findProduct(newProduct.ID)
}

func (s *ProductService) UpdateAdminProduct(id uint, input ProductUpdateInput) (*product.Product, error) {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}

	previousProduct := *existingProduct

	if input.UpdateProductTypeID {
		existingProduct.ProductTypeID = input.ProductTypeID
	}
	if input.Name != nil {
		existingProduct.Name = *input.Name
	}
	if input.Slug != nil {
		existingProduct.Slug = *input.Slug
	}
	if input.Description != nil {
		existingProduct.Description = *input.Description
	}
	if input.ShortDesc != nil {
		existingProduct.ShortDesc = *input.ShortDesc
	}
	if input.Weight != nil {
		existingProduct.Weight = *input.Weight
	}
	if input.Status != nil {
		existingProduct.Status = *input.Status
	}
	if input.Locale != nil {
		existingProduct.Locale = *input.Locale
	}
	if input.UpdateParentID {
		existingProduct.ParentID = input.ParentID
	}
	if input.Featured != nil {
		existingProduct.Featured = *input.Featured
	}
	if input.MetaTitle != nil {
		existingProduct.MetaTitle = *input.MetaTitle
	}
	if input.MetaDesc != nil {
		existingProduct.MetaDesc = *input.MetaDesc
	}

	var specValues []product.ProductSpecValue
	if input.UpdateSpecValues {
		specValues, err = s.buildSpecValues(existingProduct.ProductTypeID, input.SpecValues)
		if err != nil {
			return nil, err
		}
	}

	var variants []product.ProductVariant
	if input.UpdateVariants {
		variants, err = s.buildVariants(existingProduct.ProductTypeID, input.Variants)
		if err != nil {
			return nil, err
		}
		if err := s.ensureVariantSKUsAvailable(variants, existingProduct.ID); err != nil {
			return nil, err
		}
	}

	if err := s.productRepo.UpdateWithSpecValuesAndVariants(existingProduct, specValues, input.UpdateSpecValues, variants, input.UpdateVariants); err != nil {
		return nil, err
	}

	s.clearProductCache(&previousProduct)
	s.clearProductCache(existingProduct)

	return s.findProduct(existingProduct.ID)
}

func (s *ProductService) Delete(id uint) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)

	return nil
}

func (s *ProductService) UpdateStatus(id uint, status string) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)

	return nil
}

func (s *ProductService) BatchUpdateStatus(ids []uint, status string) (int, error) {
	updated := 0
	for _, id := range ids {
		if err := s.UpdateStatus(id, status); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return updated, err
		}
		updated++
	}

	return updated, nil
}

func (s *ProductService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return deleted, err
		}
		deleted++
	}

	return deleted, nil
}
