package product

import "gorm.io/gorm"

// Localized descriptions remain independent. Engineering facts are read from
// the source product, just as translated SKU inventory is read from its master.
func (p *Product) AfterFind(tx *gorm.DB) error {
	if p.ProductType != ProductTypeCarbonRim || p.ParentID == nil || *p.ParentID == 0 {
		return nil
	}
	var source Product
	if err := tx.Session(&gorm.Session{NewDB: true}).Select("carbon_rim_specification").Where("id = ? AND product_type = ?", *p.ParentID, ProductTypeCarbonRim).First(&source).Error; err != nil {
		return err
	}
	p.CarbonRimSpecification = source.CarbonRimSpecification
	return nil
}
