package repository

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/orderevidence"

	"gorm.io/gorm"
)

func (r *OrderEvidenceRepository) FindPackageByID(id uint) (*orderevidence.OrderEvidencePackage, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var pkg orderevidence.OrderEvidencePackage
	err := r.db.Preload("Items.Attachments").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		First(&pkg, id).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *OrderEvidenceRepository) FindItemAndPackageByID(
	itemID uint,
) (*orderevidence.OrderEvidenceItem, *orderevidence.OrderEvidencePackage, error) {
	if r == nil || r.db == nil {
		return nil, nil, gorm.ErrInvalidDB
	}
	if itemID == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}
	var item orderevidence.OrderEvidenceItem
	if err := r.db.First(&item, itemID).Error; err != nil {
		return nil, nil, err
	}
	pkg, err := r.FindPackageByID(item.PackageID)
	if err != nil {
		return nil, nil, err
	}
	if pkg.OrderID != item.OrderID {
		return nil, nil, errors.New("order evidence item package order scope mismatch")
	}
	return &item, pkg, nil
}

func (r *OrderEvidenceRepository) UpdateItemAndRefreshPackage(
	item *orderevidence.OrderEvidenceItem,
) (*orderevidence.OrderEvidencePackage, orderevidence.EvidenceCompleteness, error) {
	if r == nil || r.db == nil {
		return nil, orderevidence.EvidenceCompleteness{}, gorm.ErrInvalidDB
	}
	if item == nil || item.ID == 0 {
		return nil, orderevidence.EvidenceCompleteness{}, errors.New("order evidence item is required")
	}

	var updatedPackage orderevidence.OrderEvidencePackage
	var completeness orderevidence.EvidenceCompleteness
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var storedItem orderevidence.OrderEvidenceItem
		if err := tx.First(&storedItem, item.ID).Error; err != nil {
			return err
		}
		if storedItem.PackageID != item.PackageID || storedItem.OrderID != item.OrderID ||
			storedItem.ItemType != item.ItemType || storedItem.RequiredReason != item.RequiredReason ||
			storedItem.SnapshotID == nil != (item.SnapshotID == nil) {
			return errors.New("order evidence item immutable scope changed")
		}

		var pkg orderevidence.OrderEvidencePackage
		if err := tx.First(&pkg, item.PackageID).Error; err != nil {
			return err
		}
		if err := pkg.EnsureMutable(); err != nil {
			return err
		}

		itemToSave := *item
		itemToSave.Attachments = nil
		if err := tx.Save(&itemToSave).Error; err != nil {
			return err
		}

		var items []orderevidence.OrderEvidenceItem
		if err := tx.Where("package_id = ?", pkg.ID).Order("id ASC").Find(&items).Error; err != nil {
			return err
		}
		completeness = orderevidence.EvaluateEvidenceCompleteness(items)
		if completeness.Ready {
			pkg.Status = orderevidence.PackageStatusReady
		} else {
			pkg.Status = orderevidence.PackageStatusIncomplete
		}
		pkg.LockedAt = nil
		pkgToSave := pkg
		pkgToSave.Items = nil
		if err := tx.Save(&pkgToSave).Error; err != nil {
			return err
		}
		updatedPackage = pkg
		return nil
	})
	if err != nil {
		return nil, orderevidence.EvidenceCompleteness{}, err
	}
	return &updatedPackage, completeness, nil
}

func (r *OrderEvidenceRepository) LockPackage(
	packageID uint,
	lockedAt time.Time,
) (*orderevidence.OrderEvidencePackage, orderevidence.EvidenceCompleteness, error) {
	if r == nil || r.db == nil {
		return nil, orderevidence.EvidenceCompleteness{}, gorm.ErrInvalidDB
	}
	if packageID == 0 {
		return nil, orderevidence.EvidenceCompleteness{}, gorm.ErrRecordNotFound
	}

	var lockedPackage orderevidence.OrderEvidencePackage
	var completeness orderevidence.EvidenceCompleteness
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var pkg orderevidence.OrderEvidencePackage
		if err := tx.First(&pkg, packageID).Error; err != nil {
			return err
		}
		if pkg.IsLocked() {
			lockedPackage = pkg
			items, err := loadEvidenceItems(tx, packageID)
			if err != nil {
				return err
			}
			completeness = orderevidence.EvaluateEvidenceCompleteness(items)
			return orderevidence.ErrOrderEvidencePackageLocked
		}
		if pkg.IsSuperseded() {
			return orderevidence.ErrOrderEvidencePackageSuperseded
		}
		items, err := loadEvidenceItems(tx, packageID)
		if err != nil {
			return err
		}
		completeness = orderevidence.EvaluateEvidenceCompleteness(items)
		if !completeness.Ready {
			return orderevidence.ErrOrderEvidencePackageIncomplete
		}
		pkg.Status = orderevidence.PackageStatusLocked
		pkg.LockedAt = &lockedAt
		pkgToSave := pkg
		pkgToSave.Items = nil
		if err := tx.Save(&pkgToSave).Error; err != nil {
			return err
		}
		lockedPackage = pkg
		return nil
	})
	if err != nil {
		return nil, completeness, err
	}
	return &lockedPackage, completeness, nil
}

func (r *OrderEvidenceRepository) CreateRevision(
	packageID uint,
	createdBy uint,
) (*orderevidence.OrderEvidencePackage, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if packageID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var revision orderevidence.OrderEvidencePackage
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var source orderevidence.OrderEvidencePackage
		if err := tx.Preload("Items.Attachments").
			Preload("Items", func(db *gorm.DB) *gorm.DB {
				return db.Order("id ASC")
			}).
			First(&source, packageID).Error; err != nil {
			return err
		}
		if err := source.EnsureRevisionSource(); err != nil {
			return err
		}
		if len(source.Items) == 0 {
			return errors.New("order evidence package items are required")
		}

		revision = orderevidence.OrderEvidencePackage{
			OrderID:                    source.OrderID,
			SnapshotID:                 source.SnapshotID,
			PackageVersion:             source.PackageVersion + 1,
			Status:                     orderevidence.PackageStatusIncomplete,
			OrderTotalUSDSnapshotMinor: source.OrderTotalUSDSnapshotMinor,
			IsHighValue:                source.IsHighValue,
			HasSpokeTensionQC:          source.HasSpokeTensionQC,
			SchemaVersion:              source.SchemaVersion,
			CreatedBy:                  createdBy,
		}
		copiedItems := make([]orderevidence.OrderEvidenceItem, 0, len(source.Items))
		type attachmentCopies struct {
			itemIndex   int
			attachments []orderevidence.OrderEvidenceAttachment
		}
		copiedAttachments := make([]attachmentCopies, 0, len(source.Items))
		for index, sourceItem := range source.Items {
			copiedItem := sourceItem
			copiedItem.ID = 0
			copiedItem.PackageID = 0
			copiedItem.Attachments = nil
			copiedItem.CreatedAt = time.Time{}
			copiedItem.UpdatedAt = time.Time{}
			copiedItems = append(copiedItems, copiedItem)

			attachments := make([]orderevidence.OrderEvidenceAttachment, 0, len(sourceItem.Attachments))
			for _, sourceAttachment := range sourceItem.Attachments {
				copiedAttachment := sourceAttachment
				copiedAttachment.ID = 0
				copiedAttachment.EvidenceItemID = 0
				copiedAttachment.CreatedAt = time.Time{}
				attachments = append(attachments, copiedAttachment)
			}
			copiedAttachments = append(copiedAttachments, attachmentCopies{
				itemIndex:   index,
				attachments: attachments,
			})
		}

		if completeness := orderevidence.EvaluateEvidenceCompleteness(copiedItems); completeness.Ready {
			revision.Status = orderevidence.PackageStatusReady
		}
		revisionToSave := revision
		revisionToSave.Items = nil
		if err := tx.Create(&revisionToSave).Error; err != nil {
			return err
		}
		for index := range copiedItems {
			copiedItems[index].PackageID = revisionToSave.ID
			if err := tx.Create(&copiedItems[index]).Error; err != nil {
				return err
			}
			for _, copiedAttachment := range copiedAttachments[index].attachments {
				copiedAttachment.EvidenceItemID = copiedItems[index].ID
				if err := tx.Create(&copiedAttachment).Error; err != nil {
					return err
				}
			}
		}
		revision = revisionToSave
		revision.Items = copiedItems
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &revision, nil
}

func loadEvidenceItems(tx *gorm.DB, packageID uint) ([]orderevidence.OrderEvidenceItem, error) {
	var items []orderevidence.OrderEvidenceItem
	err := tx.Where("package_id = ?", packageID).Order("id ASC").Find(&items).Error
	return items, err
}
