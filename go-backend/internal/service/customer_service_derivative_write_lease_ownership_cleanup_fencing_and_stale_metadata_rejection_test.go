package service

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

func TestCustomerServicePrivateDerivativeWriteLeasePreventsCleanupAndCannotBeReleasedByAnotherWriter(t *testing.T) {
	env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
	writeContext, release, err := env.mediaService.acquireCustomerServicePrivateDerivativeWriteLease(t.Context(), asset)
	require.NoError(t, err)
	t.Cleanup(release)
	deadline, bounded := writeContext.Deadline()
	require.True(t, bounded)
	require.LessOrEqual(t, time.Until(deadline), media.CustomerServiceAttachmentUploadMaximumWriteDuration)
	leased, err := env.mediaService.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.NotEmpty(t, leased.CustomerServiceDerivativeWriteToken)
	_, _, err = env.mediaService.acquireCustomerServicePrivateDerivativeWriteLease(t.Context(), asset)
	require.ErrorIs(t, err, ErrCustomerServiceAttachmentWritesInProgress)
	require.NoError(t, env.mediaService.repo.ReleaseCustomerServiceDerivativeWriteLease(t.Context(), asset.ID, uuid.NewString()))
	stillLeased, err := env.mediaService.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, leased.CustomerServiceDerivativeWriteToken, stillLeased.CustomerServiceDerivativeWriteToken)
	result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Blocked)
	require.Zero(t, result.Retired)
	release()
	result, err = retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	// Late metadata cannot resurrect a retired source or introduce orphan rows.
	require.Error(t, env.mediaService.repo.CreateAssetDerivatives(asset.Derivatives))
	_, err = env.mediaService.repo.ReplaceAssetDerivatives(asset.Derivatives)
	require.Error(t, err)
	_, _, err = env.mediaService.acquireCustomerServicePrivateDerivativeWriteLease(t.Context(), asset)
	require.Error(t, err)
}

type customerServicePrivateDerivativeUploadStorageWithLeaseObservation struct {
	storage.StorageService
	observeLease func()
}

func (s *customerServicePrivateDerivativeUploadStorageWithLeaseObservation) UploadPrivateObjectFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename, prefix, cacheControl string) (string, error) {
	s.observeLease()
	return s.StorageService.(storage.PrivateCacheControlledReaderUploader).UploadPrivateObjectFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, cacheControl)
}

func TestCustomerServicePrivateDerivativeGenerationOwnsPersistentLeaseDuringEveryObjectWrite(t *testing.T) {
	env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	writes := 0
	env.mediaService.storage = &customerServicePrivateDerivativeUploadStorageWithLeaseObservation{StorageService: env.mediaService.storage,
		observeLease: func() {
			current, err := env.mediaService.repo.FindAssetByID(asset.ID)
			require.NoError(t, err)
			require.NotEmpty(t, current.CustomerServiceDerivativeWriteToken)
			require.NotNil(t, current.CustomerServiceDerivativeWriteExpiresAt)
			require.True(t, current.CustomerServiceDerivativeWriteExpiresAt.After(time.Now()))
			writes++
		}}
	generated, err := env.mediaService.generateAssetDerivatives(t.Context(), asset, multipartFileHeader(t, "lease-observed.png", "image/png", testOpaquePNG(t, 64, 32)))
	require.NoError(t, err)
	require.Len(t, generated, 3)
	require.Equal(t, 3, writes)
	current, err := env.mediaService.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.Empty(t, current.CustomerServiceDerivativeWriteToken)
	require.Nil(t, current.CustomerServiceDerivativeWriteExpiresAt)
}
