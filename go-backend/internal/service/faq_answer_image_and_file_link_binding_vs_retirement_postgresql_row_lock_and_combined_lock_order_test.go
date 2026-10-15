package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func TestPostgreSQLFAQAnswerImageAndFileLinkCreateAndUpdateSerializeBindingAndRetirement(t *testing.T) {
	for _, field := range []string{"answer image", "file link"} {
		for _, mutation := range []string{"create", "update"} {
			for _, firstOperation := range []string{"bind", "retire"} {
				t.Run(field+"/"+mutation+"/"+firstOperation+" first", func(t *testing.T) {
					env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
					require.NoError(t, env.db.AutoMigrate(&faq.FAQ{}))
					ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
					defer cancel()
					db := env.db.WithContext(ctx)
					assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
					assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
					faqs := repository.NewFAQRepository(db)
					faqs.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
					asset := createRegistered800PixelWebPAssetForFAQMediaBindingTest(t, db, "content/racing-faq.webp")
					var existingID uint
					if mutation == "update" {
						item := &faq.FAQ{Question: "Existing", Answer: "<p>Before</p>", Locale: "en"}
						require.NoError(t, faqs.Create(item))
						existingID = item.ID
					}
					firstLocked := make(chan int, 1)
					releaseFirst := make(chan struct{})
					var releaseOnce, pauseOnce sync.Once
					release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
					defer release()
					pauseAfterAssetLock := func(tx *gorm.DB) {
						pauseOnce.Do(func() {
							var pid int
							if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
								tx.AddError(err)
								return
							}
							firstLocked <- pid
							select {
							case <-releaseFirst:
							case <-ctx.Done():
								tx.AddError(ctx.Err())
							}
						})
					}
					callbackName := "pause_faq_binding_retirement_" + uuid.NewString()
					if firstOperation == "bind" {
						callback := func(tx *gorm.DB) {
							if tx.Statement.Table == "faqs" {
								pauseAfterAssetLock(tx)
							}
						}
						if mutation == "create" {
							require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, callback))
							defer db.Callback().Create().Remove(callbackName)
						} else {
							require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, callback))
							defer db.Callback().Update().Remove(callbackName)
						}
					} else {
						require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "media_assets" {
								pauseAfterAssetLock(tx)
							}
						}))
						defer db.Callback().Delete().Remove(callbackName)
					}
					bindFAQReference := func() error {
						item := &faq.FAQ{ID: existingID, Question: "Concurrent FAQ", Answer: "<p>Answer</p>", Locale: "en"}
						if field == "answer image" {
							item.AnswerImageURL = "/uploads/" + asset.StorageKey
						} else {
							item.Answer = `<a href="/uploads/` + asset.StorageKey + `">Download</a>`
						}
						if mutation == "create" {
							return faqs.Create(item)
						}
						return faqs.Update(item)
					}
					bindResult, retireResult := make(chan error, 1), make(chan error, 1)
					bind := func() { bindResult <- bindFAQReference() }
					retire := func() { retireResult <- assets.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }
					if firstOperation == "bind" {
						go bind()
					} else {
						go retire()
					}
					var blockerPID int
					select {
					case blockerPID = <-firstLocked:
					case <-ctx.Done():
						t.Fatal("first operation did not acquire its media lock")
					}
					if firstOperation == "bind" {
						go retire()
					} else {
						go bind()
					}
					require.Eventually(t, func() bool {
						var waiting bool
						err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE '%media_assets%')", blockerPID).Scan(&waiting).Error
						return err == nil && waiting
					}, 4*time.Second, 20*time.Millisecond)
					release()
					var bindError, retireError error
					select {
					case bindError = <-bindResult:
					case <-ctx.Done():
						t.Fatal("FAQ binding timed out")
					}
					select {
					case retireError = <-retireResult:
					case <-ctx.Done():
						t.Fatal("FAQ retirement timed out")
					}
					var cleanupEvents int64
					require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).Count(&cleanupEvents).Error)
					if firstOperation == "bind" {
						require.NoError(t, bindError)
						require.ErrorIs(t, retireError, ErrMediaAssetInUse)
						require.Zero(t, cleanupEvents)
						report, err := assets.GetAssetReferences(asset.ID)
						require.NoError(t, err)
						require.Equal(t, 1, report.Total)
					} else {
						require.NoError(t, retireError)
						require.ErrorIs(t, bindError, repository.ErrFAQMediaReferenceInvalid)
						require.EqualValues(t, 1, cleanupEvents)
						var registrations int64
						require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Where("id = ?", asset.ID).Count(&registrations).Error)
						require.Zero(t, registrations)
						require.ErrorIs(t, bindFAQReference(), repository.ErrFAQMediaReferenceInvalid, "cleanup must not permit URL-only rebinding")
						if existingID != 0 {
							current, err := faqs.FindByID(existingID)
							require.NoError(t, err)
							require.Equal(t, "<p>Before</p>", current.Answer)
							require.Empty(t, current.AnswerImageURL)
						}
					}
				})
			}
		}
	}
}

func TestPostgreSQLFAQAnswerLinksAndDedicatedImagesAcquireOneConsistentAssetLockOrder(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&faq.FAQ{}))
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	db := env.db.WithContext(ctx)
	assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	faqs := repository.NewFAQRepository(db)
	faqs.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	low := createRegistered800PixelWebPAssetForFAQMediaBindingTest(t, db, "content/low.webp")
	high := createRegistered800PixelWebPAssetForFAQMediaBindingTest(t, db, "content/high.webp")
	first := &faq.FAQ{Question: "First", Answer: "<p>Before</p>"}
	second := &faq.FAQ{Question: "Second", Answer: "<p>Before</p>"}
	require.NoError(t, faqs.Create(first))
	require.NoError(t, faqs.Create(second))
	first.Answer, first.AnswerImageURL = `<a href="`+high.URL+`">Download</a>`, low.URL
	second.Answer, second.AnswerImageURL = `<a href="`+low.URL+`">Download</a>`, high.URL
	firstLocked, releaseFirst := make(chan int, 1), make(chan struct{})
	var pauseOnce, releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
	defer release()
	callbackName := "pause_faq_combined_asset_order_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table != "media_assets" || !strings.Contains(tx.Statement.SQL.String(), "FOR UPDATE") {
			return
		}
		pauseOnce.Do(func() {
			asset, ok := tx.Statement.Dest.(*media.MediaAsset)
			if !ok || asset.ID != low.ID {
				tx.AddError(fmt.Errorf("combined FAQ plan must lock the lowest asset ID first"))
				return
			}
			var pid int
			if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
				tx.AddError(err)
				return
			}
			firstLocked <- pid
			select {
			case <-releaseFirst:
			case <-ctx.Done():
				tx.AddError(ctx.Err())
			}
		})
	}))
	defer db.Callback().Query().Remove(callbackName)
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go func() { firstResult <- faqs.Update(first) }()
	var blockerPID int
	select {
	case blockerPID = <-firstLocked:
	case <-ctx.Done():
		t.Fatal("first FAQ did not acquire the lowest asset lock")
	}
	go func() { secondResult <- faqs.Update(second) }()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE '%media_assets%')", blockerPID).Scan(&waiting).Error
		return err == nil && waiting
	}, 4*time.Second, 20*time.Millisecond)
	release()
	for _, result := range []chan error{firstResult, secondResult} {
		select {
		case err := <-result:
			require.NoError(t, err)
		case <-ctx.Done():
			t.Fatal("combined FAQ asset locks deadlocked")
		}
	}
}
