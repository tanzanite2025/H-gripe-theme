package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

func createCompletelyMigratedDisposableTemplateCategoryMediaBindingTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || directory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and complete migrations")
	}
	control, err := sql.Open("postgres", connectionURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, control.Close()) })
	databaseName := fmt.Sprintf("local_backup_template_category_writes_%d", time.Now().UnixNano())
	_, err = control.ExecContext(t.Context(), `CREATE DATABASE `+pq.QuoteIdentifier(databaseName))
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, err := control.ExecContext(ctx, `DROP DATABASE `+pq.QuoteIdentifier(databaseName)+` WITH (FORCE)`)
		require.NoError(t, err)
	})
	parsed, err := url.Parse(connectionURL)
	require.NoError(t, err)
	parsed.Path = "/" + databaseName
	migrationDB, err := sql.Open("postgres", parsed.String())
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(migrationDB, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(directory), "postgres", driver)
	require.NoError(t, err)
	err = migrator.Up()
	if err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	version, dirty, err := migrator.Version()
	require.NoError(t, err)
	require.Equal(t, uint(436), version)
	require.False(t, dirty)
	sourceError, databaseError := migrator.Close()
	require.NoError(t, sourceError)
	require.NoError(t, databaseError)
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(8)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	return db
}

func waitForTemplateCategoryConcurrentOperationToBlockOnTransaction(t *testing.T, db *gorm.DB, blockerPID int, table string) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw(`SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE ?)`, blockerPID, "%"+table+"%").Scan(&waiting).Error
		return err == nil && waiting
	}, 5*time.Second, 20*time.Millisecond, "the second writer must wait on the actual PostgreSQL row lock")
}

func TestPostgreSQLCompletelyMigratedTemplateAndCategoryIntroSerializeNormalBindingRetirementAndCleanup(t *testing.T) {
	db := createCompletelyMigratedDisposableTemplateCategoryMediaBindingTestDatabase(t)
	for _, field := range []string{"template_create", "template_update", "category_default", "category_translation", "category_translation_first_write"} {
		for _, placement := range []string{"image", "video", "poster", "file_link"} {
			for _, firstOperation := range []string{"bind", "retire"} {
				t.Run(field+"/"+placement+"/"+firstOperation+"_first", func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
					defer cancel()
					env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, db.WithContext(ctx))
					target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
					asset := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/concurrent-"+uuid.NewString()+".png")
					markup := `<img src="/uploads/` + asset.StorageKey + `">`
					switch placement {
					case "video":
						require.NoError(t, env.db.Model(asset).Update("media_type", "video").Error)
						markup = `<video src="/uploads/` + asset.StorageKey + `"></video>`
					case "poster":
						markup = `<video src="https://ordinary.invalid/video.mp4" poster="/uploads/` + asset.StorageKey + `"></video>`
					case "file_link":
						markup = `<a href="/uploads/` + asset.StorageKey + `#page=2">File</a>`
					}
					locked, releaseChannel := make(chan int, 1), make(chan struct{})
					var pauseOnce, releaseOnce sync.Once
					release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
					defer release()
					callbackName := "pause_template_category_asset_binding_" + uuid.NewString()
					pause := func(tx *gorm.DB) {
						if tx.Statement.Table == target.table {
							pauseOnce.Do(func() {
								pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
							})
						}
					}
					if firstOperation == "bind" {
						require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callbackName, pause))
						require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callbackName, pause))
						defer env.db.Callback().Create().Remove(callbackName)
						defer env.db.Callback().Update().Remove(callbackName)
					} else {
						require.NoError(t, env.db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "media_assets" {
								pauseOnce.Do(func() {
									pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
								})
							}
						}))
						defer env.db.Callback().Delete().Remove(callbackName)
					}
					bound, retired := make(chan error, 1), make(chan error, 1)
					bind := func() { bound <- target.write(markup) }
					retire := func() { retired <- env.assets.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }
					if firstOperation == "bind" {
						go bind()
					} else {
						go retire()
					}
					var blockerPID int
					select {
					case blockerPID = <-locked:
					case err := <-bound:
						t.Fatalf("first binding ended before reaching its locked write: %v", err)
					case err := <-retired:
						t.Fatalf("first retirement ended before reaching its locked write: %v", err)
					case <-ctx.Done():
						t.Fatal("first operation did not reach its locked business/retirement write")
					}
					if firstOperation == "bind" {
						go retire()
					} else {
						go bind()
					}
					waitForTemplateCategoryConcurrentOperationToBlockOnTransaction(t, env.db, blockerPID, "media_assets")
					release()
					var bindError, retirementError error
					select {
					case bindError = <-bound:
					case <-ctx.Done():
						t.Fatal("binding failed to finish")
					}
					select {
					case retirementError = <-retired:
					case <-ctx.Done():
						t.Fatal("retirement failed to finish")
					}
					var cleanupEvents int64
					require.NoError(t, env.db.Model(&outbox.Event{}).Where("event_type=? AND aggregate_id=?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).Count(&cleanupEvents).Error)
					if firstOperation == "bind" {
						require.NoError(t, bindError)
						require.ErrorIs(t, retirementError, ErrMediaAssetInUse)
						require.Zero(t, cleanupEvents)
						require.Contains(t, readTemplateOrCategoryIntroPersistedMediaHTML(t, env.db, target), asset.URL)
					} else {
						require.NoError(t, retirementError)
						require.ErrorIs(t, bindError, target.invalidError)
						require.EqualValues(t, 1, cleanupEvents)
						require.ErrorIs(t, target.write(markup), target.invalidError, "a physically removed registration cannot resurrect its old managed path")
						if target.id != 0 {
							require.Empty(t, readTemplateOrCategoryIntroPersistedMediaHTML(t, env.db, target))
						}
					}
				})
			}
		}
	}
	verifyTemplateCategoryOmittedFieldsAndTranslationFirstWritesUnderPostgreSQLLocks(t, db)
	verifyTemplateCategoryCombinedAssetLockOrderUnderPostgreSQLConcurrency(t, db)
	verifyTemplateCategoryHistoricalDedicatedObjectDetachmentAndCleanupConcurrency(t, db)
}

func verifyTemplateCategoryOmittedFieldsAndTranslationFirstWritesUnderPostgreSQLLocks(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, field := range []string{"category_default", "category_translation"} {
		t.Run("omitted_intro_after_concurrent_detachment/"+field, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, db.WithContext(ctx))
			target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/omitted-"+uuid.NewString()+".png")
			require.NoError(t, target.write(`<img src="`+asset.URL+`">`))
			categoryID, locale := target.id, "en"
			if field == "category_translation" {
				locale = "fr"
				require.NoError(t, env.db.Table(target.table).Where("id=?", target.id).Pluck("product_category_id", &categoryID).Error)
			}
			tx := env.db.Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()
			var category product.ProductCategory
			require.NoError(t, tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&category, categoryID).Error)
			require.NoError(t, tx.Table(target.table).Where("id=?", target.id).Updates(map[string]any{target.column: "", "meta_description": "Newest committed description"}).Error)
			var blockerPID int
			require.NoError(t, tx.Raw("SELECT pg_backend_pid()").Scan(&blockerPID).Error)
			done := make(chan error, 1)
			go func() {
				title := "Changed title only"
				_, err := env.categories.UpdateSEO(categoryID, locale, &title, nil, nil)
				done <- err
			}()
			waitForTemplateCategoryConcurrentOperationToBlockOnTransaction(t, env.db, blockerPID, "product_categories")
			require.NoError(t, tx.Commit().Error)
			require.NoError(t, <-done)
			require.Empty(t, readTemplateOrCategoryIntroPersistedMediaHTML(t, env.db, target))
			view, err := env.categories.GetSEO(categoryID, locale)
			require.NoError(t, err)
			require.Equal(t, "Newest committed description", view.MetaDescription)
		})
	}
	for _, secondOperation := range []string{"seo_metadata", "translation_names"} {
		t.Run("translation_first_write_then_"+secondOperation, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, db.WithContext(ctx))
			target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, "category_translation_first_write")
			var categoryID uint
			require.NoError(t, env.db.Model(&product.ProductCategory{}).Order("id DESC").Limit(1).Pluck("id", &categoryID).Error)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/translation-first-"+uuid.NewString()+".png")
			locked, releaseChannel := make(chan int, 1), make(chan struct{})
			var pauseOnce, releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
			defer release()
			callbackName := "pause_first_category_translation_insert_" + uuid.NewString()
			require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
				if tx.Statement.Table == "product_category_translations" {
					pauseOnce.Do(func() {
						pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
					})
				}
			}))
			defer env.db.Callback().Create().Remove(callbackName)
			first, second := make(chan error, 1), make(chan error, 1)
			go func() { first <- target.write(`<img src="` + asset.URL + `">`) }()
			var blockerPID int
			select {
			case blockerPID = <-locked:
			case <-ctx.Done():
				t.Fatal("first translation write did not pause")
			}
			go func() {
				if secondOperation == "seo_metadata" {
					description := "Second writer description"
					_, err := env.categories.UpdateSEO(categoryID, "fr", nil, &description, nil)
					second <- err
				} else {
					_, err := env.categories.UpdateTranslations(categoryID, []ProductCategoryTranslationInput{{Locale: "fr", Name: "Second writer name"}})
					second <- err
				}
			}()
			waitForTemplateCategoryConcurrentOperationToBlockOnTransaction(t, env.db, blockerPID, "product_categories")
			release()
			require.NoError(t, <-first)
			require.NoError(t, <-second)
			var translations []product.ProductCategoryTranslation
			require.NoError(t, env.db.Where("product_category_id=? AND locale='fr'", categoryID).Find(&translations).Error)
			require.Len(t, translations, 1)
			require.Equal(t, target.id, translations[0].ID)
			require.Contains(t, translations[0].SEOIntro, asset.URL)
		})
	}
}

func verifyTemplateCategoryCombinedAssetLockOrderUnderPostgreSQLConcurrency(t *testing.T, db *gorm.DB) {
	t.Helper()
	t.Run("template_and_nested_category_images_share_sorted_asset_locks", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
		defer cancel()
		env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, db.WithContext(ctx))
		target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, "template_update")
		low := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/low-"+uuid.NewString()+".png")
		high := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/high-"+uuid.NewString()+".png")
		locked, releaseChannel := make(chan int, 1), make(chan struct{})
		var pauseOnce, releaseOnce sync.Once
		release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
		defer release()
		callbackName := "pause_template_category_sorted_asset_plan_" + uuid.NewString()
		require.NoError(t, env.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
			if tx.Statement.Table != "media_assets" || !strings.Contains(tx.Statement.SQL.String(), "FOR UPDATE") {
				return
			}
			pauseOnce.Do(func() {
				asset, ok := tx.Statement.Dest.(*media.MediaAsset)
				if !ok || asset.ID != low.ID {
					tx.AddError(fmt.Errorf("combined media plan must acquire the lowest asset ID first"))
					return
				}
				pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
			})
		}))
		defer env.db.Callback().Query().Remove(callbackName)
		first, second := make(chan error, 1), make(chan error, 1)
		go func() { first <- target.write(`<img src="` + high.URL + `"><img src="` + low.URL + `">`) }()
		var blockerPID int
		select {
		case blockerPID = <-locked:
		case <-ctx.Done():
			t.Fatal("combined media plan did not acquire its first sorted asset lock")
		}
		category := &product.ProductCategory{Name: "Combined category", Slug: "combined-" + uuid.NewString(), ImageMediaAssetID: &high.ID, Translations: []product.ProductCategoryTranslation{{Locale: "fr", Name: "Combined translation", SEOIntro: `<img src="` + low.URL + `">`}}}
		go func() { second <- env.categories.repo.Create(category) }()
		waitForTemplateCategoryConcurrentOperationToBlockOnTransaction(t, env.db, blockerPID, "media_assets")
		release()
		require.NoError(t, <-first)
		require.NoError(t, <-second)
	})
}

func verifyTemplateCategoryHistoricalDedicatedObjectDetachmentAndCleanupConcurrency(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		for _, field := range []string{"template_update", "category_default", "category_translation"} {
			for _, scenario := range []string{"uncommitted_detachment", "stale_restoration_during_cleanup"} {
				t.Run("dedicated_cleanup/"+namespace+"/"+field+"/"+scenario, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
					defer cancel()
					env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, db.WithContext(ctx))
					target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
					root, key := t.TempDir(), namespace+"/historical/"+uuid.NewString()+".png"
					provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://dedicated.example.test"})
					require.NoError(t, err)
					objectPath := filepath.Join(root, filepath.FromSlash(key))
					require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
					require.NoError(t, os.WriteFile(objectPath, []byte("synthetic historical template/category object"), 0o600))
					html := `<img src="/uploads/` + key + `">`
					require.NoError(t, env.db.Table(target.table).Where("id=?", target.id).Update(target.column, html).Error)
					logos := NewSiteLogoService(repository.NewSiteLogoRepository(env.db), provider, "https://shop.example.test")
					favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(env.db), provider, "https://shop.example.test")
					visuals := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(env.db), provider, "https://shop.example.test")
					handler := NewObjectStorageCleanupOutboxHandler(nil, logos, visuals, nil)
					handler.ConfigureSiteFaviconService(favicons)
					resource := map[string]string{"site-logo": objectCleanupResourceSiteLogo, "site-favicon": objectCleanupResourceSiteFavicon, "visual-showcase": objectCleanupResourceHomeVisualTile}[namespace]
					event, err := newObjectStorageCleanupEvent(resource, "1", resource, "1", []string{key})
					require.NoError(t, err)
					releaseChannel := make(chan struct{})
					var releaseOnce sync.Once
					release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
					defer release()
					if scenario == "uncommitted_detachment" {
						locked := make(chan int, 1)
						var pauseOnce sync.Once
						callbackName := "pause_template_category_dedicated_detachment_" + uuid.NewString()
						require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == target.table {
								pauseOnce.Do(func() {
									pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
								})
							}
						}))
						defer env.db.Callback().Update().Remove(callbackName)
						done := make(chan error, 1)
						go func() { done <- target.write(`<p>Detached</p>`) }()
						select {
						case <-locked:
						case <-ctx.Done():
							t.Fatal("dedicated detachment did not pause")
						}
						require.ErrorIs(t, handler.Handle(ctx, *event), ErrMediaAssetInUse)
						requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
						release()
						require.NoError(t, <-done)
						require.NoError(t, handler.Handle(ctx, *event))
					} else {
						require.NoError(t, target.write(`<p>Detached</p>`))
						deleting := make(chan struct{}, 1)
						paused := &historicalDedicatedObjectCleanupStorageWithDeletionPause{StorageService: provider, beforeDelete: func() error {
							deleting <- struct{}{}
							select {
							case <-releaseChannel:
								return nil
							case <-ctx.Done():
								return ctx.Err()
							}
						}}
						logos.storage, favicons.storage, visuals.storage = paused, paused, paused
						done := make(chan error, 1)
						go func() { done <- handler.Handle(ctx, *event) }()
						select {
						case <-deleting:
						case err := <-done:
							t.Fatalf("dedicated cleanup ended before physical deletion: %v", err)
						case <-ctx.Done():
							t.Fatal("dedicated cleanup did not reach physical deletion")
						}
						require.ErrorIs(t, target.write(html), target.invalidError)
						release()
						require.NoError(t, <-done)
					}
					_, err = os.Stat(objectPath)
					require.True(t, os.IsNotExist(err))
					require.Equal(t, `<p>Detached</p>`, readTemplateOrCategoryIntroPersistedMediaHTML(t, env.db, target))
				})
			}
		}
	}
}
