package storage

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const SuggestionFeedbackAttachmentNamespace = "suggestion-feedback/attachments"
const PayPalDisputeInvoiceAttachmentNamespace = "paypal-dispute-invoices"

func AllocateSuggestionOrPayPalAttachmentObjectKey(namespace string, ownerID uint, filename string) (string, error) {
	key := fmt.Sprintf("%s/%d/%s%s", namespace, ownerID, uuid.NewString(), strings.ToLower(path.Ext(filename)))
	if _, _, _, valid := SuggestionOrPayPalAttachmentObjectIdentity(key); !valid {
		return "", fmt.Errorf("invalid suggestion or PayPal attachment object identity")
	}
	return key, nil
}

func SuggestionOrPayPalAttachmentObjectIdentity(key string) (namespace string, ownerID uint, receiptID string, valid bool) {
	normalized, ok := NormalizeObjectKey(key)
	if !ok || normalized != key {
		return
	}
	for _, candidate := range []string{SuggestionFeedbackAttachmentNamespace, PayPalDisputeInvoiceAttachmentNamespace} {
		if strings.HasPrefix(key, candidate+"/") {
			namespace = candidate
			break
		}
	}
	if namespace == "" {
		return
	}
	parts := strings.Split(strings.TrimPrefix(key, namespace+"/"), "/")
	if len(parts) != 2 {
		return namespace, 0, "", false
	}
	parsedOwner, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || parsedOwner == 0 || uint64(uint(parsedOwner)) != parsedOwner || strconv.FormatUint(parsedOwner, 10) != parts[0] {
		return namespace, 0, "", false
	}
	extension := path.Ext(parts[1])
	if namespace == PayPalDisputeInvoiceAttachmentNamespace {
		if extension != ".pdf" {
			return namespace, 0, "", false
		}
	} else {
		switch extension {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		default:
			return namespace, 0, "", false
		}
	}
	receiptID = strings.TrimSuffix(parts[1], extension)
	id, err := uuid.Parse(receiptID)
	if err != nil || id == uuid.Nil || id.String() != receiptID {
		return namespace, 0, "", false
	}
	return namespace, uint(parsedOwner), receiptID, true
}
