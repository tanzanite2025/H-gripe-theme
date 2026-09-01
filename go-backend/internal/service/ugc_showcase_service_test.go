package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShowcaseListPublicOnlyReturnsApprovedItems(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)

	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 1,
		Kind:   ugcshowcase.KindUser,
		Title:  "approved",
		Status: ugcshowcase.StatusApproved,
	}).Error)
	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 2,
		Kind:   ugcshowcase.KindUser,
		Title:  "pending",
		Status: ugcshowcase.StatusPending,
	}).Error)

	items, err := UGCShowcaseService.ListPublic(ugcshowcase.KindUser, 1, 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "approved", items[0].Title)
}

func TestShowcaseUploadPhotosDeletesUploadedImagesWhenCreateFails(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{
		uploadURLs: []string{
			"http://localhost:9200/uploads/showcase/pending/2026/08/13/first.webp",
			"http://localhost:9200/uploads/showcase/pending/2026/08/13/second.webp",
		},
	}
	UGCShowcaseService.storage = storage
	require.NoError(t, db.Migrator().DropTable(&ugcshowcase.UGCShowcase{}))
	completedAt := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&order.Order{
		ID:          91,
		UserID:      7,
		OrderNumber: "SHOWCASE-91",
		Status:      "completed",
		CompletedAt: &completedAt,
		TotalAmount: 100,
		Currency:    "USD",
	}).Error)
	UGCShowcaseService.ConfigureUploadEligibility(NewUGCShowcaseUploadEligibilityService(repository.NewOrderRepository(db)))

	_, err := UGCShowcaseService.UploadPhotos(context.Background(), 7, 91, []*multipart.FileHeader{
		{Filename: "first.webp", Size: 100},
		{Filename: "second.webp", Size: 100},
	}, map[string]string{"region": "US"})
	require.Error(t, err)
	assert.Equal(t, []string{
		"showcase/pending/2026/08/13/first.webp",
		"showcase/pending/2026/08/13/second.webp",
	}, storage.deletedURLs)
	assert.Equal(t, []string{"showcase/pending", "showcase/pending"}, storage.uploadPrefixes)
}

func TestShowcaseUploadPhotosDeletesPreviousImagesWhenLaterUploadFails(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{
		uploadURLs: []string{"http://localhost:9200/uploads/showcase/pending/2026/08/13/first.webp"},
		uploadErr:  errors.New("storage write failed"),
		errAt:      2,
	}
	UGCShowcaseService.storage = storage
	completedAt := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&order.Order{
		ID:          92,
		UserID:      7,
		OrderNumber: "SHOWCASE-92",
		Status:      "completed",
		CompletedAt: &completedAt,
		TotalAmount: 100,
		Currency:    "USD",
	}).Error)
	UGCShowcaseService.ConfigureUploadEligibility(NewUGCShowcaseUploadEligibilityService(repository.NewOrderRepository(db)))

	_, err := UGCShowcaseService.UploadPhotos(context.Background(), 7, 92, []*multipart.FileHeader{
		{Filename: "first.webp", Size: 100},
		{Filename: "second.webp", Size: 100},
	}, map[string]string{"region": "US"})
	require.Error(t, err)
	assert.Equal(t, []string{"showcase/pending/2026/08/13/first.webp"}, storage.deletedURLs)
	assert.Equal(t, []string{"showcase/pending", "showcase/pending"}, storage.uploadPrefixes)
}

func TestShowcaseUploadPhotosStoresPendingImageKeysOnly(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{
		uploadURLs: []string{
			"http://localhost:9200/uploads/showcase/pending/2026/08/13/first.webp",
			"http://localhost:9200/uploads/showcase/pending/2026/08/13/second.webp",
		},
	}
	UGCShowcaseService.storage = storage
	completedAt := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&order.Order{
		ID:          93,
		UserID:      7,
		OrderNumber: "SHOWCASE-93",
		Status:      "completed",
		CompletedAt: &completedAt,
		TotalAmount: 100,
		Currency:    "USD",
	}).Error)
	UGCShowcaseService.ConfigureUploadEligibility(NewUGCShowcaseUploadEligibilityService(repository.NewOrderRepository(db)))

	item, err := UGCShowcaseService.UploadPhotos(context.Background(), 7, 93, []*multipart.FileHeader{
		{Filename: "first.webp", Size: 100},
		{Filename: "second.webp", Size: 100},
	}, map[string]string{"region": "US"})
	require.NoError(t, err)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)

	var savedImages []string
	require.NoError(t, json.Unmarshal(saved.Images, &savedImages))
	require.Equal(t, []string{
		"showcase/pending/2026/08/13/first.webp",
		"showcase/pending/2026/08/13/second.webp",
	}, savedImages)
	for _, imageReference := range savedImages {
		assert.NotContains(t, imageReference, "://")
		assert.True(t, strings.HasPrefix(imageReference, "showcase/pending/"))
	}
	assert.NotNil(t, saved.OrderID)
	assert.Equal(t, uint(93), *saved.OrderID)
}

func TestShowcaseUploadPhotosRejectsIneligibleOrderBeforeStorageWrite(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{
		uploadURLs: []string{"http://localhost:9200/uploads/showcase/pending/2026/08/13/should-not-exist.webp"},
	}
	UGCShowcaseService.storage = storage
	UGCShowcaseService.ConfigureUploadEligibility(NewUGCShowcaseUploadEligibilityService(repository.NewOrderRepository(db)))

	_, err := UGCShowcaseService.UploadPhotos(context.Background(), 7, 404, []*multipart.FileHeader{
		{Filename: "should-not-exist.webp", Size: 100},
	}, map[string]string{"region": "US"})
	require.ErrorIs(t, err, ErrShowcaseUploadOrderNotEligible)
	assert.Empty(t, storage.uploadPrefixes)
	assert.Empty(t, storage.deletedURLs)
}

func TestShowcaseUploadPhotosEnforcesPendingLimitAtCreateTime(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{
		uploadURLs: []string{"http://localhost:9200/uploads/showcase/pending/2026/08/13/blocked.webp"},
	}
	UGCShowcaseService.storage = storage
	UGCShowcaseService.ConfigurePendingSubmissionLimit(1)
	completedAt := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&user.User{
		ID:       7,
		Email:    "showcase-limit@example.com",
		Username: "showcase-limit",
		Password: "password",
	}).Error)
	require.NoError(t, db.Create(&order.Order{
		ID:          94,
		UserID:      7,
		OrderNumber: "SHOWCASE-94",
		Status:      "completed",
		CompletedAt: &completedAt,
		TotalAmount: 100,
		Currency:    "USD",
	}).Error)
	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 7,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
	}).Error)
	UGCShowcaseService.ConfigureUploadEligibility(NewUGCShowcaseUploadEligibilityService(repository.NewOrderRepository(db)))

	_, err := UGCShowcaseService.UploadPhotos(context.Background(), 7, 94, []*multipart.FileHeader{
		{Filename: "blocked.webp", Size: 100},
	}, map[string]string{"region": "US"})
	require.ErrorIs(t, err, ErrShowcaseUploadPendingLimitExceeded)
	assert.Equal(t, []string{"showcase/pending/2026/08/13/blocked.webp"}, storage.deletedURLs)

	count, err := UGCShowcaseService.CountPendingSubmissions(7)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestShowcaseCountPendingSubmissionsUsesPendingStatus(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)

	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 7,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
	}).Error)
	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 7,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusApproved,
	}).Error)
	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 8,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
	}).Error)

	count, err := UGCShowcaseService.CountPendingSubmissions(7)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestShowcasePublicImageAccessBlocksPendingAndAllowsApprovedRecords(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	UGCShowcaseService.storage = &fakeShowcaseStorage{}

	pendingImage := "http://localhost:9200/uploads/showcase/pending/2026/08/13/pending.webp"
	approvedImage := "http://localhost:9200/uploads/showcase/approved/2026/08/13/approved.webp"
	pendingJSON, err := json.Marshal([]string{pendingImage})
	require.NoError(t, err)
	approvedJSON, err := json.Marshal([]string{approvedImage})
	require.NoError(t, err)

	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 1,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: pendingJSON,
	}).Error)
	require.NoError(t, db.Create(&ugcshowcase.UGCShowcase{
		UserID: 2,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusApproved,
		Images: approvedJSON,
	}).Error)

	pendingAccess, err := UGCShowcaseService.PublicImageAccess(context.Background(), "showcase/pending/2026/08/13/pending.webp")
	require.NoError(t, err)
	assert.True(t, pendingAccess.Found)
	assert.False(t, pendingAccess.Allowed)

	approvedAccess, err := UGCShowcaseService.PublicImageAccess(context.Background(), "showcase/approved/2026/08/13/approved.webp")
	require.NoError(t, err)
	assert.True(t, approvedAccess.Found)
	assert.True(t, approvedAccess.Allowed)
}

func TestShowcaseOpenPublicImageFileRejectsPendingRecord(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	UGCShowcaseService.storage = &fakeShowcaseStorage{}

	imagesJSON, err := json.Marshal([]string{
		"showcase/pending/2026/08/13/pending.webp",
	})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 1,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	_, err = UGCShowcaseService.OpenPublicImageFile(context.Background(), item.ID, 0)
	require.ErrorIs(t, err, ErrShowcaseImageNotFound)
}

func TestShowcaseOpenPublicImageFileRequiresApprovedRecord(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	UGCShowcaseService.storage = &fakeShowcaseStorage{}

	imagesJSON, err := json.Marshal([]string{
		"showcase/approved/2026/08/13/approved.webp",
	})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 1,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusApproved,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	_, err = UGCShowcaseService.OpenPublicImageFile(context.Background(), item.ID, 0)
	require.Error(t, err)
}

func TestShowcaseOpenPublicImageFileRejectsApprovedRecordOutsideApprovedNamespace(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	UGCShowcaseService.storage = &fakeShowcaseStorage{}

	imagesJSON, err := json.Marshal([]string{
		"media/private/2026/08/13/private.webp",
	})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 1,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusApproved,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	_, err = UGCShowcaseService.OpenPublicImageFile(context.Background(), item.ID, 0)
	require.ErrorIs(t, err, ErrShowcaseImageNotFound)
}

func TestShowcaseApprovePublishesPendingImagesAndDeletesPendingSources(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{}
	UGCShowcaseService.storage = storage

	imageKey := "showcase/pending/2026/08/13/source.webp"
	imagesJSON, err := json.Marshal([]string{imageKey})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 3,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	require.NoError(t, UGCShowcaseService.Approve(context.Background(), item.ID))
	require.Len(t, storage.copiedKeys, 1)
	assert.Equal(t, "showcase/pending/2026/08/13/source.webp", storage.copiedKeys[0][0])
	assert.True(t, strings.HasPrefix(storage.copiedKeys[0][1], "showcase/approved/"))
	assert.True(t, strings.HasSuffix(storage.copiedKeys[0][1], "/2026/08/13/source.webp"))
	assert.Equal(t, []string{imageKey}, storage.deletedURLs)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, ugcshowcase.StatusApproved, saved.Status)
	var savedImages []string
	require.NoError(t, json.Unmarshal(saved.Images, &savedImages))
	assert.Equal(t, []string{storage.copiedKeys[0][1]}, savedImages)
}

func TestShowcaseApproveReturnsStorageUnavailableWhenStorageIsMissing(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)

	imagesJSON, err := json.Marshal([]string{"showcase/pending/2026/08/13/source.webp"})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 3,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	err = UGCShowcaseService.Approve(context.Background(), item.ID)
	require.ErrorIs(t, err, ErrShowcaseStorageUnavailable)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, ugcshowcase.StatusPending, saved.Status)
}

func TestShowcaseApproveKeepsPendingSourceWhenCopyFails(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{copyErr: errors.New("copy failed")}
	UGCShowcaseService.storage = storage

	imageKey := "showcase/pending/2026/08/13/source.webp"
	imagesJSON, err := json.Marshal([]string{imageKey})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 3,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	require.Error(t, UGCShowcaseService.Approve(context.Background(), item.ID))
	assert.Empty(t, storage.deletedURLs)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, ugcshowcase.StatusPending, saved.Status)
	var savedImages []string
	require.NoError(t, json.Unmarshal(saved.Images, &savedImages))
	assert.Equal(t, []string{imageKey}, savedImages)
}

func TestShowcaseApproveDeletesCopiedApprovedImagesWhenLaterCopyFails(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{
		copyErr:   errors.New("copy failed"),
		copyErrAt: 2,
	}
	UGCShowcaseService.storage = storage

	imageKeys := []string{
		"showcase/pending/2026/08/13/first.webp",
		"showcase/pending/2026/08/13/second.webp",
	}
	imagesJSON, err := json.Marshal(imageKeys)
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 3,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	require.Error(t, UGCShowcaseService.Approve(context.Background(), item.ID))
	require.Len(t, storage.copiedKeys, 1)
	assert.Equal(t, []string{storage.copiedKeys[0][1]}, storage.deletedURLs)
	assert.NotContains(t, storage.deletedURLs, imageKeys[0])
	assert.NotContains(t, storage.deletedURLs, imageKeys[1])

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, ugcshowcase.StatusPending, saved.Status)
	var savedImages []string
	require.NoError(t, json.Unmarshal(saved.Images, &savedImages))
	assert.Equal(t, imageKeys, savedImages)
}

func TestShowcaseRejectOnlyDeletesPendingImages(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{}
	UGCShowcaseService.storage = storage

	imagesJSON, err := json.Marshal([]string{
		"showcase/pending/2026/08/13/pending.webp",
		"http://localhost:9200/uploads/showcase/approved/2026/08/13/approved.webp",
	})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID: 4,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: imagesJSON,
	}
	require.NoError(t, db.Create(item).Error)

	require.NoError(t, UGCShowcaseService.Reject(context.Background(), item.ID, "not suitable"))
	assert.Equal(t, []string{
		"showcase/pending/2026/08/13/pending.webp",
	}, storage.deletedURLs)
}

func TestShowcaseRejectDoesNotChangeStatusWhenImagesAreInvalid(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{}
	UGCShowcaseService.storage = storage

	item := &ugcshowcase.UGCShowcase{
		UserID: 10,
		Kind:   ugcshowcase.KindUser,
		Status: ugcshowcase.StatusPending,
		Images: []byte(`{"invalid":true}`),
	}
	require.NoError(t, db.Create(item).Error)

	err := UGCShowcaseService.Reject(context.Background(), item.ID, "invalid image record")
	require.ErrorIs(t, err, ErrShowcaseImagesInvalid)
	assert.Empty(t, storage.deletedURLs)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, ugcshowcase.StatusPending, saved.Status)
}

func TestShowcaseApproveRejectsNonPendingRecords(t *testing.T) {
	for _, status := range []string{ugcshowcase.StatusApproved, ugcshowcase.StatusRejected} {
		t.Run(status, func(t *testing.T) {
			db, UGCShowcaseService := newTestShowcaseService(t)
			storage := &fakeShowcaseStorage{}
			UGCShowcaseService.storage = storage

			imagesJSON, err := json.Marshal([]string{
				"showcase/pending/2026/08/13/image.webp",
			})
			require.NoError(t, err)
			item := &ugcshowcase.UGCShowcase{
				UserID: 8,
				Kind:   ugcshowcase.KindUser,
				Status: status,
				Images: imagesJSON,
				Title:  "already moderated",
			}
			require.NoError(t, db.Create(item).Error)

			err = UGCShowcaseService.Approve(context.Background(), item.ID)
			require.ErrorIs(t, err, ErrShowcaseInvalidTransition)
			assert.Empty(t, storage.copiedKeys)
			assert.Empty(t, storage.deletedURLs)
		})
	}
}

func TestShowcaseRejectRejectsNonPendingRecords(t *testing.T) {
	for _, status := range []string{ugcshowcase.StatusApproved, ugcshowcase.StatusRejected} {
		t.Run(status, func(t *testing.T) {
			db, UGCShowcaseService := newTestShowcaseService(t)
			storage := &fakeShowcaseStorage{}
			UGCShowcaseService.storage = storage

			imagesJSON, err := json.Marshal([]string{
				"showcase/pending/2026/08/13/image.webp",
			})
			require.NoError(t, err)
			item := &ugcshowcase.UGCShowcase{
				UserID: 9,
				Kind:   ugcshowcase.KindUser,
				Status: status,
				Images: imagesJSON,
				Title:  "already moderated",
			}
			require.NoError(t, db.Create(item).Error)

			err = UGCShowcaseService.Reject(context.Background(), item.ID, "duplicate moderation")
			require.ErrorIs(t, err, ErrShowcaseInvalidTransition)
			assert.Empty(t, storage.deletedURLs)
		})
	}
}

func TestShowcaseCleanupExpiresOldPendingRecordsAndRemovesDeletedImageReferences(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{}
	UGCShowcaseService.storage = storage

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	pendingKey := "showcase/pending/2026/08/01/pending.webp"
	approvedURL := "http://localhost:9200/uploads/showcase/approved/existing/approved.webp"
	imagesJSON, err := json.Marshal([]string{pendingKey, approvedURL})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID:    5,
		Kind:      ugcshowcase.KindUser,
		Status:    ugcshowcase.StatusPending,
		Images:    imagesJSON,
		CreatedAt: now.Add(-8 * 24 * time.Hour),
		UpdatedAt: now.Add(-8 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(item).Error)

	result, err := UGCShowcaseService.CleanupExpiredPendingImages(
		context.Background(),
		now,
		7*24*time.Hour,
		100,
	)
	require.NoError(t, err)
	assert.Equal(t, 1, result.ScannedCandidates)
	assert.Equal(t, 1, result.ExpiredPendingRecords)
	assert.Equal(t, 1, result.DeletedPendingImages)
	assert.Equal(t, 1, result.UpdatedImageReferences)
	assert.Equal(t, []string{pendingKey}, storage.deletedURLs)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, ugcshowcase.StatusRejected, saved.Status)
	assert.Equal(t, showcaseExpiredPendingReason, saved.RejectedReason)
	var savedImages []string
	require.NoError(t, json.Unmarshal(saved.Images, &savedImages))
	assert.Equal(t, []string{approvedURL}, savedImages)
}

func TestShowcaseCleanupRetainsFailedPendingImageReferencesForRetry(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	pendingKey := "showcase/pending/2026/08/01/pending.webp"
	storage := &fakeShowcaseStorage{
		deleteErrors: map[string]error{pendingKey: errors.New("delete failed")},
	}
	UGCShowcaseService.storage = storage

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	imagesJSON, err := json.Marshal([]string{pendingKey})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID:    6,
		Kind:      ugcshowcase.KindUser,
		Status:    ugcshowcase.StatusRejected,
		Images:    imagesJSON,
		CreatedAt: now.Add(-8 * 24 * time.Hour),
		UpdatedAt: now.Add(-8 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(item).Error)

	result, err := UGCShowcaseService.CleanupExpiredPendingImages(
		context.Background(),
		now,
		7*24*time.Hour,
		100,
	)
	require.NoError(t, err)
	assert.Equal(t, 1, result.ScannedCandidates)
	assert.Equal(t, 0, result.DeletedPendingImages)
	assert.Equal(t, 1, result.RetainedFailedImages)
	assert.Equal(t, 0, result.UpdatedImageReferences)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.JSONEq(t, string(imagesJSON), string(saved.Images))
}

func TestShowcaseCleanupLeavesRecentPendingRecordsUntouched(t *testing.T) {
	db, UGCShowcaseService := newTestShowcaseService(t)
	storage := &fakeShowcaseStorage{}
	UGCShowcaseService.storage = storage

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	pendingKey := "showcase/pending/2026/08/12/recent.webp"
	imagesJSON, err := json.Marshal([]string{pendingKey})
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{
		UserID:    7,
		Kind:      ugcshowcase.KindUser,
		Status:    ugcshowcase.StatusPending,
		Images:    imagesJSON,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now.Add(-24 * time.Hour),
	}
	require.NoError(t, db.Create(item).Error)

	result, err := UGCShowcaseService.CleanupExpiredPendingImages(
		context.Background(),
		now,
		7*24*time.Hour,
		100,
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ScannedCandidates)
	assert.Empty(t, storage.deletedURLs)

	var saved ugcshowcase.UGCShowcase
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, ugcshowcase.StatusPending, saved.Status)
}

func newTestShowcaseService(t *testing.T) (*gorm.DB, *UGCShowcaseService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&user.User{}, &order.Order{}, &ugcshowcase.UGCShowcase{}, &ugcshowcase.UGCShowcaseComment{}))
	return db, NewUGCShowcaseService(repository.NewUGCShowcaseRepository(db), nil)
}

type fakeShowcaseStorage struct {
	uploadURLs     []string
	deletedURLs    []string
	copiedKeys     [][2]string
	uploadPrefixes []string
	uploadErr      error
	copyErr        error
	deleteErrors   map[string]error
	copyErrAt      int
	errAt          int
	uploadCalled   int
	copyCalled     int
}

func (s *fakeShowcaseStorage) Upload(context.Context, *multipart.FileHeader) (string, error) {
	return "", errors.New("UploadWithPrefix expected")
}

func (s *fakeShowcaseStorage) UploadWithPrefix(_ context.Context, _ *multipart.FileHeader, prefix string) (string, error) {
	s.uploadCalled++
	s.uploadPrefixes = append(s.uploadPrefixes, prefix)
	if s.uploadErr != nil && s.uploadCalled == s.errAt {
		return "", s.uploadErr
	}
	if s.uploadCalled > len(s.uploadURLs) {
		return "", errors.New("unexpected upload")
	}
	return s.uploadURLs[s.uploadCalled-1], nil
}

func (s *fakeShowcaseStorage) UploadWithPrefixPrivate(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	return s.UploadWithPrefix(ctx, file, prefix)
}

func (s *fakeShowcaseStorage) UploadFromReader(context.Context, io.Reader, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *fakeShowcaseStorage) UploadFromReaderWithPrefix(context.Context, io.Reader, string, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *fakeShowcaseStorage) Delete(_ context.Context, url string) error {
	s.deletedURLs = append(s.deletedURLs, url)
	if err := s.deleteErrors[url]; err != nil {
		return err
	}
	return nil
}

func (s *fakeShowcaseStorage) GetURL(filename string) string {
	return filename
}

func (s *fakeShowcaseStorage) ObjectKey(reference string) (string, error) {
	const marker = "/uploads/"
	if index := strings.Index(reference, marker); index >= 0 {
		return strings.Trim(reference[index+len(marker):], "/"), nil
	}
	return strings.Trim(reference, "/"), nil
}

func (s *fakeShowcaseStorage) CopyObject(_ context.Context, sourceKey, destKey string) error {
	s.copyCalled++
	if s.copyErr != nil && (s.copyErrAt == 0 || s.copyCalled == s.copyErrAt) {
		return s.copyErr
	}
	s.copiedKeys = append(s.copiedKeys, [2]string{sourceKey, destKey})
	return nil
}
