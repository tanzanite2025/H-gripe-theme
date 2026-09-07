package admin

import (
	"strings"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/service"
)

const adminAuditResourceOrderOutboundEvidence = "order_outbound_evidence"

func orderEvidenceAuditItemValue(item *orderevidence.OrderEvidenceItem) map[string]interface{} {
	if item == nil {
		return nil
	}
	return map[string]interface{}{
		"item_id":             item.ID,
		"package_id":          item.PackageID,
		"order_id":            item.OrderID,
		"item_type":           strings.TrimSpace(item.ItemType),
		"required_reason":     strings.TrimSpace(item.RequiredReason),
		"status":              strings.TrimSpace(item.Status),
		"order_item_id":       auditUintPointerValue(item.OrderItemID),
		"attachment_count":    len(item.Attachments),
		"captured_by":         item.CapturedBy,
		"captured_at_present": item.CapturedAt != nil && !item.CapturedAt.IsZero(),
	}
}

func orderEvidenceAuditPackageValue(pkg *orderevidence.OrderEvidencePackage) map[string]interface{} {
	if pkg == nil {
		return nil
	}
	return map[string]interface{}{
		"package_id":           pkg.ID,
		"order_id":             pkg.OrderID,
		"package_version":      pkg.PackageVersion,
		"status":               strings.TrimSpace(pkg.Status),
		"is_high_value":        pkg.IsHighValue,
		"has_spoke_tension_qc": pkg.HasSpokeTensionQC,
		"item_count":           len(pkg.Items),
	}
}

func orderEvidenceUpdateAuditChanges(
	itemID uint,
	request orderEvidenceItemUpdateRequest,
	result *service.OrderEvidenceAdminPackageResult,
) map[string]interface{} {
	changes := map[string]interface{}{
		"item_id":             itemID,
		"requested_status":    strings.TrimSpace(request.Status),
		"data_json_present":   len(request.DataJSON) > 0,
		"captured_at_present": request.CapturedAt != nil && !request.CapturedAt.IsZero(),
	}
	if item := orderEvidenceResultItem(result, itemID); item != nil {
		changes["attachment_count"] = len(item.Attachments)
		changes["item_type"] = strings.TrimSpace(item.ItemType)
		changes["package_version"] = orderEvidenceResultPackageVersion(result)
	}
	return changes
}

func orderEvidenceResultItem(
	result *service.OrderEvidenceAdminPackageResult,
	itemID uint,
) *orderevidence.OrderEvidenceItem {
	if result == nil || result.Package == nil {
		return nil
	}
	for index := range result.Package.Items {
		if result.Package.Items[index].ID == itemID {
			return &result.Package.Items[index]
		}
	}
	return nil
}

func orderEvidenceResultPackageVersion(result *service.OrderEvidenceAdminPackageResult) int {
	if result == nil || result.Package == nil {
		return 0
	}
	return result.Package.PackageVersion
}

func evidencePackageFromResult(
	result *service.OrderEvidenceAdminPackageResult,
) *orderevidence.OrderEvidencePackage {
	if result == nil {
		return nil
	}
	return result.Package
}

func orderEvidenceAttachmentAuditValue(
	attachment *orderevidence.OrderEvidenceAttachment,
	attachmentCount int,
) map[string]interface{} {
	if attachment == nil {
		return nil
	}
	return map[string]interface{}{
		"attachment_id":    attachment.ID,
		"evidence_item_id": attachment.EvidenceItemID,
		"mime_type":        strings.TrimSpace(attachment.MimeType),
		"size_bytes":       attachment.SizeBytes,
		"sha256":           strings.TrimSpace(attachment.SHA256),
		"attachment_count": attachmentCount,
	}
}

func orderEvidenceExportSnapshotAuditValue(
	snapshot *orderevidence.OrderEvidenceExportSnapshot,
) map[string]interface{} {
	if snapshot == nil {
		return nil
	}
	return map[string]interface{}{
		"snapshot_id":             snapshot.ID,
		"order_id":                snapshot.OrderID,
		"evidence_package_id":     snapshot.EvidencePackageID,
		"evidence_package_version": snapshot.EvidencePackageVersion,
		"snapshot_version":        snapshot.Version,
		"snapshot_sha256":         strings.TrimSpace(snapshot.SnapshotSHA256),
		"locked_at":               snapshot.LockedAt.UTC(),
	}
}
