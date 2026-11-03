package localbackup

import (
	"strings"
)

// Preserve legacy permission precedence for unregistered paths. Registered task actions carry their permission beside their handler.
func requiredMaintenancePermissionForUnregisteredRequestPath(path string) string {
	path = strings.TrimRight(path, "/")
	permission := PermissionView
	if strings.HasSuffix(path, "/customer-profile-module-formal-restore") || strings.HasSuffix(path, "/customer-profile-module-interrupted-restore-recovery") || strings.HasSuffix(path, "/order-module-formal-history-restore") || strings.HasSuffix(path, "/order-module-interrupted-restore-recovery") || strings.HasSuffix(path, "/customer-service-module-formal-restore") || strings.HasSuffix(path, "/customer-service-module-interrupted-restore-recovery") {
		permission = PermissionRestore
	}
	if strings.HasSuffix(path, "/product-catalog-stopped-read-only-public-image-http-query") {
		permission = PermissionRestore
	}
	if strings.HasPrefix(path, "/downloads/") || strings.HasSuffix(path, "/download-ticket") || path == "/api/backups" || path == "/api/product-catalog-backups" || path == "/api/module-backup-preview" || path == "/api/module-backups" {
		permission = PermissionExport
	}
	if strings.HasPrefix(path, "/api/uploads") || strings.HasSuffix(path, "/verify") {
		permission = PermissionUpload
	}
	if strings.HasSuffix(path, "/apply") || strings.HasSuffix(path, "/rollback") || strings.HasSuffix(path, "/complete") || strings.HasSuffix(path, "/release") || strings.HasSuffix(path, "/selected-module-isolated-restore-preview") || strings.HasSuffix(path, "/selected-customer-service-restore-choice") || strings.HasSuffix(path, "/selected-customer-service-candidate-rehearsal") {
		permission = PermissionRestore
	}
	if strings.HasSuffix(path, "/reconciliation-report") || strings.HasSuffix(path, "/reconciliation-connections") || strings.HasSuffix(path, "/module-dependency-report") || strings.HasSuffix(path, "/customer-service-conflict-review") || strings.HasSuffix(path, "/customer-service-effective-selection-inspection") || strings.HasSuffix(path, "/product-catalog-handoff-inspection") || strings.HasSuffix(path, "/product-view-stopped-drain-and-rebuild") {
		permission = PermissionRestore
	}
	if strings.HasSuffix(path, "/product-catalog-stopped-handoff-preparation") || strings.HasSuffix(path, "/product-catalog-stopped-handoff-cancellation") || strings.HasSuffix(path, "/product-catalog-stopped-isolated-switch") || strings.HasSuffix(path, "/product-catalog-stopped-post-switch-inspection") || strings.HasSuffix(path, "/product-catalog-stopped-switch-rollback") || strings.HasSuffix(path, "/product-catalog-stopped-detail-cache-invalidation") || strings.HasSuffix(path, "/product-catalog-formal-restore-acceptance") {
		permission = PermissionRestore
	}
	if strings.HasSuffix(path, "/product-catalog-stopped-projection-inspection") || strings.HasSuffix(path, "/product-catalog-stopped-projection-evidence") || strings.HasSuffix(path, "/product-catalog-stopped-read-only-product-http-query") || strings.HasSuffix(path, "/product-catalog-stopped-read-only-media-http-query") || strings.HasSuffix(path, "/product-catalog-stopped-read-only-storefront-http-query") {
		permission = PermissionRestore
	}
	return permission
}
