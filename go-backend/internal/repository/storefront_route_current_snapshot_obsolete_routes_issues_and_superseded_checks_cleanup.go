package repository

import (
	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"

	"gorm.io/gorm"
)

// Related modules are optional in isolated catalog installations. Production
// foreign keys are supplemented by explicit cleanup for each installed module.
func deleteStorefrontRouteChecksAndIssues(tx *gorm.DB, routeEntryIDs []uint) error {
	if len(routeEntryIDs) == 0 {
		return nil
	}
	if tx.Migrator().HasTable(&urlmanagementdomain.StorefrontURLIssue{}) {
		issueIDs := tx.Model(&urlmanagementdomain.StorefrontURLIssue{}).Select("id").Where("route_entry_id IN ?", routeEntryIDs)
		if tx.Migrator().HasTable(&urlmanagementdomain.StorefrontURLIssueEvent{}) {
			if err := tx.Where("issue_id IN (?)", issueIDs).Delete(&urlmanagementdomain.StorefrontURLIssueEvent{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("route_entry_id IN ?", routeEntryIDs).Delete(&urlmanagementdomain.StorefrontURLIssue{}).Error; err != nil {
			return err
		}
	}
	if tx.Migrator().HasTable(&seodomain.StorefrontRouteCheckResult{}) {
		return tx.Where("route_entry_id IN ?", routeEntryIDs).Delete(&seodomain.StorefrontRouteCheckResult{}).Error
	}
	return nil
}

func deleteStorefrontRoutesAbsentFromCurrentSnapshot(tx *gorm.DB, routeEntryIDs []uint) error {
	if len(routeEntryIDs) == 0 {
		return nil
	}
	if err := deleteStorefrontRouteChecksAndIssues(tx, routeEntryIDs); err != nil {
		return err
	}
	if tx.Migrator().HasTable(&urlmanagementdomain.StorefrontURLSearchProfile{}) {
		if err := tx.Where("route_entry_id IN ?", routeEntryIDs).Delete(&urlmanagementdomain.StorefrontURLSearchProfile{}).Error; err != nil {
			return err
		}
	}
	// Stop future quality samples before the route foreign key is cleared.
	if tx.Migrator().HasTable("site_quality_targets") {
		if err := tx.Table("site_quality_targets").Where("route_entry_id IN ?", routeEntryIDs).
			Updates(map[string]interface{}{"enabled": false, "next_scheduled_at": nil}).Error; err != nil {
			return err
		}
	}
	return tx.Where("id IN ?", routeEntryIDs).Delete(&seodomain.StorefrontRouteCatalogEntry{}).Error
}

func retainOnlyLatestStorefrontRouteCheckResults(tx *gorm.DB, routeEntryIDs ...uint) error {
	if !tx.Migrator().HasTable(&seodomain.StorefrontRouteCheckResult{}) {
		return nil
	}
	// checked_at, then id, determines the latest sample even for imported data.
	supersededChecks := tx.Table("storefront_route_check_results AS previous_check").Select("previous_check.id").
		Where(`EXISTS (SELECT 1 FROM storefront_route_check_results current_check
			WHERE current_check.route_entry_id = previous_check.route_entry_id
			AND (current_check.checked_at > previous_check.checked_at
			OR (current_check.checked_at = previous_check.checked_at AND current_check.id > previous_check.id)))`)
	if len(routeEntryIDs) > 0 {
		supersededChecks = supersededChecks.Where("previous_check.route_entry_id IN ?", routeEntryIDs)
	}
	if tx.Migrator().HasTable(&urlmanagementdomain.StorefrontURLIssue{}) {
		if err := tx.Model(&urlmanagementdomain.StorefrontURLIssue{}).
			Where("latest_check_result_id IN (?)", supersededChecks).
			Update("latest_check_result_id", gorm.Expr(`(SELECT id FROM storefront_route_check_results
				WHERE route_entry_id = storefront_url_issues.route_entry_id ORDER BY checked_at DESC, id DESC LIMIT 1)`)).Error; err != nil {
			return err
		}
	}
	return tx.Where("id IN (?)", supersededChecks).Delete(&seodomain.StorefrontRouteCheckResult{}).Error
}
