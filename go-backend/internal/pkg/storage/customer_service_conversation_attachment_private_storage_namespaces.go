package storage

import "strings"

const (
	CustomerServiceConversationAttachmentOriginalsPrefix   = "customer-service/attachments/originals"
	CustomerServiceConversationAttachmentDerivativesPrefix = "customer-service/attachments/derivatives"
)

func IsCustomerServiceConversationAttachmentOriginalObjectKey(key string) bool {
	normalized, ok := NormalizeObjectKey(key)
	return ok && strings.HasPrefix(normalized, CustomerServiceConversationAttachmentOriginalsPrefix+"/")
}

func IsCustomerServiceConversationAttachmentDerivativeObjectKey(key string) bool {
	normalized, ok := NormalizeObjectKey(key)
	return ok && strings.HasPrefix(normalized, CustomerServiceConversationAttachmentDerivativesPrefix+"/")
}
