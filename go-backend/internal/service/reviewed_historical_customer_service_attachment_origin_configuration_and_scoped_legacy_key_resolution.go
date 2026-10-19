package service

import (
	"fmt"
	"net/url"
	"strings"

	"commerce-platform/internal/pkg/storage"
)

// This allowlist applies only to reading persisted unclassified historical
// attachments through their conversation. It does not change public-business
// trust, authorize new attachments, or publish an object.
func (s *MediaService) ConfigureReviewedHistoricalCustomerServiceAttachmentOrigins(configuredOrigins string) error {
	origins := make(map[string]struct{})
	if strings.TrimSpace(configuredOrigins) != "" {
		for _, value := range strings.Split(configuredOrigins, ",") {
			parsed, err := url.Parse(strings.TrimSpace(value))
			if err != nil || parsed.Host == "" || parsed.Hostname() == "" || strings.ContainsAny(parsed.Host, "*\\") || parsed.User != nil || parsed.Opaque != "" || strings.Contains(value, "#") ||
				(parsed.Scheme != "http" && parsed.Scheme != "https") ||
				(parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
				return fmt.Errorf("reviewed historical customer-service attachment origins must be comma-separated HTTP(S) origins without paths, credentials, queries or fragments")
			}
			origins[strings.ToLower(parsed.Scheme+"://"+parsed.Host)] = struct{}{}
		}
	}
	s.reviewedHistoricalCustomerServiceAttachmentOrigins = origins
	return nil
}

func (s *MediaService) reviewedHistoricalCustomerServiceAttachmentObjectKey(reference string) string {
	if s == nil || len(s.reviewedHistoricalCustomerServiceAttachmentOrigins) == 0 {
		return ""
	}
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil || parsed.User != nil || parsed.Opaque != "" || !strings.HasPrefix(parsed.Path, "/uploads/") {
		return ""
	}
	if _, allowed := s.reviewedHistoricalCustomerServiceAttachmentOrigins[strings.ToLower(parsed.Scheme+"://"+parsed.Host)]; !allowed {
		return ""
	}
	key, valid := storage.NormalizeObjectKey(strings.TrimPrefix(parsed.Path, "/uploads/"))
	if !valid || storage.IsPrivateObjectKey(key) {
		return ""
	}
	return key
}
