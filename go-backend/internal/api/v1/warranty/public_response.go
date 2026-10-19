package warranty

import (
	"encoding/json"
	"fmt"
	"strings"

	"commerce-platform/internal/api/v1/publicmedia"
	domainwarranty "commerce-platform/internal/domain/warranty"
)

func publicWarrantyClaimFromDomain(
	item domainwarranty.WarrantyClaim,
	resolver publicmedia.Resolver,
) domainwarranty.WarrantyClaim {
	result := item
	var images []string
	_ = json.Unmarshal([]byte(item.Images), &images)
	paths := make([]string, 0, len(images))
	for index, reference := range images {
		if strings.TrimSpace(reference) != "" {
			paths = append(paths, fmt.Sprintf("/warranty/claims/%d/attachments/images/%d", item.ID, index))
		}
	}
	encodedPaths, _ := json.Marshal(paths)
	result.Images = string(encodedPaths)
	result.VideoURL = ""
	if strings.TrimSpace(item.VideoURL) != "" {
		result.VideoURL = fmt.Sprintf("/warranty/claims/%d/attachments/video/0", item.ID)
	}

	return result
}

func publicWarrantyClaimsFromDomain(
	items []domainwarranty.WarrantyClaim,
	resolver publicmedia.Resolver,
) []domainwarranty.WarrantyClaim {
	result := make([]domainwarranty.WarrantyClaim, 0, len(items))
	for _, item := range items {
		result = append(result, publicWarrantyClaimFromDomain(item, resolver))
	}
	return result
}
