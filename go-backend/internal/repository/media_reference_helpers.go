package repository

import (
	"commerce-platform/internal/domain/media"
	"fmt"
	"strings"
)

func newMediaAssetReferenceQuery(asset *media.MediaAsset) mediaAssetReferenceQuery {
	if asset == nil || asset.ID == 0 {
		return mediaAssetReferenceQuery{}
	}

	values := []string{asset.URL}
	if key := strings.Trim(strings.ReplaceAll(asset.StorageKey, "\\", "/"), "/"); key != "" {
		values = append(values, "/uploads/"+key, "uploads/"+key)
	}

	seen := make(map[string]struct{}, len(values))
	urls := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		urls = append(urls, value)
	}
	return mediaAssetReferenceQuery{AssetID: asset.ID, URLs: urls}
}

func (r *MediaRepository) hasTable(table string) bool {
	return r.db.Migrator().HasTable(table)
}

func mediaReferenceContainsCondition(columns, urls []string) (string, []interface{}) {
	conditions := make([]string, 0, len(columns)*len(urls))
	args := make([]interface{}, 0, len(columns)*len(urls))
	for _, column := range columns {
		for _, value := range urls {
			conditions = append(conditions, column+" LIKE ?")
			args = append(args, "%"+value+"%")
		}
	}
	if len(conditions) == 0 {
		return "1 = 0", []interface{}{}
	}
	return strings.Join(conditions, " OR "), args
}

func containsMediaReferenceURL(urls []string, value string) bool {
	value = strings.TrimSpace(value)
	for _, candidate := range urls {
		if value == candidate {
			return true
		}
	}
	return false
}

func containsMediaReferenceURLInText(urls []string, value string) bool {
	for _, candidate := range urls {
		if candidate != "" && strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func newMediaReference(
	category media.ReferenceCategory,
	resourceType string,
	resourceID uint,
	parentResourceID uint,
	label string,
	field string,
) media.AssetReference {
	return media.AssetReference{
		Category:         category,
		ResourceType:     resourceType,
		ResourceID:       resourceID,
		ParentResourceID: parentResourceID,
		Label:            label,
		Field:            field,
	}
}

func namedMediaReferenceLabel(resource string, id uint, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Sprintf("%s #%d", resource, id)
	}
	return fmt.Sprintf("%s #%d：%s", resource, id, truncateMediaReferenceLabel(name))
}

func truncateMediaReferenceLabel(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= 80 {
		return string(runes)
	}
	return string(runes[:80]) + "..."
}
