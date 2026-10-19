package workbenchfeed

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) findActiveEntryWithMediaForUpdate(id uint) (*FeedEntry, error) {
	var entry FeedEntry
	if err := r.db.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&entry, id).Error; err != nil {
		return nil, err
	}
	// Associations are read only after the authority row lock has been acquired.
	return r.FindByID(id)
}

func (r *Repository) updateOnlyEntryStatusWithinParentLockedTransaction(id uint, status string, actorID uint) (*FeedEntry, error) {
	var result *FeedEntry
	err := r.db.Transaction(func(tx *gorm.DB) error {
		repo := r.WithTx(tx)
		entry, err := repo.findActiveEntryWithMediaForUpdate(id)
		if err != nil {
			return err
		}
		values := map[string]any{"status": status, "updated_by": actorID}
		if status == StatusPublished && entry.PublishedAt.IsZero() {
			values["published_at"] = time.Now().UTC()
		}
		if err := tx.Model(&FeedEntry{}).Where("id = ?", id).Updates(values).Error; err != nil {
			return err
		}
		result, err = repo.FindByID(id)
		return err
	})
	return result, err
}
