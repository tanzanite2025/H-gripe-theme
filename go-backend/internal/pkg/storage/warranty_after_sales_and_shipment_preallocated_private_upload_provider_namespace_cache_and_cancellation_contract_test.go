package storage

import "testing"

func TestPrivateOrderBusinessAttachmentPreallocatedProviderNamespacesPrivateBucketNoStoreAndCancellation(t *testing.T) {
	for _, prefix := range []string{"warranty/claims", "after-sales/pending", "warranty/shipment/41"} {
		t.Run(prefix, func(t *testing.T) {
			requirePreallocatedPrivateObjectProviderExactKeyPrivateBucketNoStoreAndCancellationContract(t, prefix+"/owned-upload-11111111-1111-4111-8111-111111111111")
		})
	}
}
