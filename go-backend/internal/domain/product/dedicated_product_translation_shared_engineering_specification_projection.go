package product

import "gorm.io/gorm"

// Product translations keep localized copy, while dedicated engineering facts
// are projected from the source-language product as their canonical record.
func (p *Product) AfterFind(tx *gorm.DB) error {
	if p.ParentID == nil || *p.ParentID == 0 {
		return nil
	}

	switch p.ProductType {
	case ProductTypeCarbonRim:
		var source Product
		if err := tx.Session(&gorm.Session{NewDB: true}).Select("carbon_rim_specification").Where("id = ? AND product_type = ?", *p.ParentID, ProductTypeCarbonRim).First(&source).Error; err != nil {
			return err
		}
		p.CarbonRimSpecification = source.CarbonRimSpecification
	case ProductTypeSpoke:
		if err := hydrateTranslatedSpokesProductSpecificationFromRootProduct(tx, p); err != nil {
			return err
		}
	}
	return nil
}

func hydrateTranslatedSpokesProductSpecificationFromRootProduct(tx *gorm.DB, translatedProduct *Product) error {
	var source Product
	if err := tx.Session(&gorm.Session{NewDB: true}).Select("spoke_specification").Where("id = ? AND product_type = ?", *translatedProduct.ParentID, ProductTypeSpoke).First(&source).Error; err != nil {
		return err
	}
	translatedProduct.SpokeSpecification = source.SpokeSpecification
	return nil
}
