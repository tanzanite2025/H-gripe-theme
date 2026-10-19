package storage

import "strings"

// DedicatedPublicObjectNamespaceForStorageKey identifies objects whose bytes
// belong to a dedicated lifecycle, even after their current owner is removed.
func DedicatedPublicObjectNamespaceForStorageKey(key string) string {
	normalized, valid := NormalizeObjectKey(key)
	if !valid {
		return ""
	}
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase", "customer-service/avatars", "showcase/pending", "showcase/approved", "order-evidence", SuggestionFeedbackAttachmentNamespace, PayPalDisputeInvoiceAttachmentNamespace} {
		if normalized == namespace || strings.HasPrefix(normalized, namespace+"/") {
			return namespace
		}
	}
	return ""
}
