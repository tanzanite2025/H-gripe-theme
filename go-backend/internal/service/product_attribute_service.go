package service

import "tanzanite/internal/domain/product"

func (s *ProductService) GetAttributeByID(id uint) (*product.ProductAttribute, error) {
	return s.productRepo.FindAttributeByID(id)
}

func (s *ProductService) GetAttributeBySlug(slug string) (*product.ProductAttribute, error) {
	return s.productRepo.FindAttributeBySlug(slug)
}

func (s *ProductService) ListAttributes(page, pageSize int) ([]product.ProductAttribute, int64, error) {
	return s.productRepo.FindAllAttributes(page, pageSize)
}

func (s *ProductService) CreateAttribute(attr *product.ProductAttribute) error {
	return s.productRepo.CreateAttribute(attr)
}

func (s *ProductService) UpdateAttribute(attr *product.ProductAttribute) error {
	return s.productRepo.UpdateAttribute(attr)
}

func (s *ProductService) DeleteAttribute(id uint) error {
	return s.productRepo.DeleteAttribute(id)
}

func (s *ProductService) GetFilterableAttributes() ([]product.ProductAttribute, error) {
	return s.productRepo.FindFilterableAttributes()
}

func (s *ProductService) GetAttributeValueByID(id uint) (*product.AttributeValue, error) {
	return s.productRepo.FindAttributeValueByID(id)
}

func (s *ProductService) CreateAttributeValue(val *product.AttributeValue) error {
	return s.productRepo.CreateAttributeValue(val)
}

func (s *ProductService) UpdateAttributeValue(val *product.AttributeValue) error {
	return s.productRepo.UpdateAttributeValue(val)
}

func (s *ProductService) DeleteAttributeValue(id uint) error {
	return s.productRepo.DeleteAttributeValue(id)
}

func (s *ProductService) GetValuesByAttributeID(attrID uint) ([]product.AttributeValue, error) {
	return s.productRepo.FindValuesByAttributeID(attrID)
}

func (s *ProductService) ListProductTypes(includeDisabled bool) ([]product.ProductType, error) {
	return s.productRepo.FindAllProductTypes(includeDisabled)
}
