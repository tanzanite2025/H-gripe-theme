package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createCompletelyMigratedDisposableProductDescriptionRepositoryTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || directory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and complete migrations")
	}
	control, err := sql.Open("postgres", connectionURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, control.Close()) })
	databaseName := fmt.Sprintf("local_backup_product_description_%d", time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	t.Cleanup(cancel)
	_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(databaseName))
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		_, err := control.ExecContext(cleanupContext, `DROP DATABASE `+pq.QuoteIdentifier(databaseName)+` WITH (FORCE)`)
		require.NoError(t, err)
	})
	parsed, err := url.Parse(connectionURL)
	require.NoError(t, err)
	parsed.Path = "/" + databaseName
	migrationDatabase, err := sql.Open("postgres", parsed.String())
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(migrationDatabase, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(directory), "postgres", driver)
	require.NoError(t, err)
	defer func() {
		sourceError, databaseError := migrator.Close()
		require.NoError(t, sourceError)
		require.NoError(t, databaseError)
	}()
	err = migrator.Up()
	if err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	version, dirty, err := migrator.Version()
	require.NoError(t, err)
	require.Equal(t, uint(435), version)
	require.False(t, dirty)
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(4)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	return db
}

func TestPostgreSQLCompletelyMigratedProductDescriptionsPersistSearchTranslateAndRetainSummaryMedia(t *testing.T) {
	db := createCompletelyMigratedDisposableProductDescriptionRepositoryTestDatabase(t)
	for _, column := range []string{"short_description", "meta_description"} {
		require.True(t, db.Migrator().HasColumn("products", column))
	}
	for _, column := range []string{"short_desc", "meta_desc"} {
		require.False(t, db.Migrator().HasColumn("products", column))
	}
	asset := media.MediaAsset{Filename: "summary-only.png", URL: "/uploads/catalog/summary-only.png", StorageKey: "catalog/summary-only.png", Size: 8, MediaType: "image", Visibility: "public", Status: "active"}
	require.NoError(t, db.Create(&asset).Error)
	summary := `<p>CanonicalSummaryOnlyNeedle &amp; exact</p><img src="` + asset.URL + `">`
	item := &product.Product{Name: "Canonical product", Slug: "canonical-product-description-test", Locale: "en", Status: "active", Currency: "USD", ShortDesc: summary, MetaDesc: "Initial exact SEO", PriceMinor: 1200}
	repo := NewProductRepository(db)
	variants := []product.ProductVariant{{SKU: "CANONICAL-DESCRIPTION-TEST", Currency: "USD", PriceMinor: 1200, Stock: 7, IsActive: true, IsDefault: true}}
	require.NoError(t, repo.CreateWithSpecValuesAndVariants(item, nil, variants))
	var actualSummary, actualSEO string
	require.NoError(t, db.Raw(`SELECT short_description,meta_description FROM products WHERE id=?`, item.ID).Row().Scan(&actualSummary, &actualSEO))
	require.Equal(t, summary, actualSummary)
	require.Equal(t, "Initial exact SEO", actualSEO)
	require.NoError(t, repo.UpdateSEO(item.ID, "Updated title", "Updated exact SEO"))
	loaded, err := repo.FindByID(item.ID)
	require.NoError(t, err)
	loaded.Name = "Ordinary canonical name edit"
	require.NoError(t, repo.Update(loaded))
	loaded, err = repo.FindByID(item.ID)
	require.NoError(t, err)
	require.Equal(t, summary, loaded.ShortDesc)
	require.Equal(t, "Updated exact SEO", loaded.MetaDesc)
	for _, search := range []struct {
		name string
		run  func() ([]product.Product, error)
	}{
		{"public", func() ([]product.Product, error) {
			rows, _, err := repo.SearchPublic(ProductSearchQuery{Locale: "en", Status: "active", Keyword: "canonicalsummaryonlyneedle", Limit: 10})
			return rows, err
		}},
		{"public_compact", func() ([]product.Product, error) {
			return repo.SearchPublicCompact(ProductSearchQuery{Locale: "en", Status: "active", Keyword: "canonicalsummaryonlyneedle", Limit: 10})
		}},
		{"admin_existing_name_search", func() ([]product.Product, error) {
			rows, _, err := repo.FindAllWithFilters(1, 10, "active", "en", loaded.Name, "", "", "")
			return rows, err
		}},
		{"quick_buy", func() ([]product.Product, error) {
			rows, _, err := repo.ListQuickBuyCandidates(ProductQuickBuyCandidateQuery{Locale: "en", Keyword: "canonicalsummaryonlyneedle", Limit: 10})
			return rows, err
		}},
		{"recommendation", func() ([]product.Product, error) {
			rows, _, err := repo.ListRecommendationCandidates(ProductRecommendationQuery{Locale: "en", Keyword: "canonicalsummaryonlyneedle", Limit: 10})
			return rows, err
		}},
	} {
		t.Run(search.name, func(t *testing.T) {
			rows, err := search.run()
			require.NoError(t, err)
			require.Len(t, rows, 1)
			require.Equal(t, item.ID, rows[0].ID)
			require.Equal(t, summary, rows[0].ShortDesc)
			require.Equal(t, "Updated exact SEO", rows[0].MetaDesc)
		})
	}
	translation := &product.Product{Name: "French canonical product", Slug: item.Slug, Locale: "fr", ParentID: &item.ID, Status: "active", Currency: "USD", ShortDesc: loaded.ShortDesc, MetaDesc: loaded.MetaDesc, PriceMinor: 1200}
	require.Len(t, loaded.Variants, 1)
	require.NoError(t, repo.CreateTranslatedCopy(loaded, translation, map[uint]string{loaded.Variants[0].ID: "CANONICAL-DESCRIPTION-FR-TEST"}))
	translated, err := repo.FindByID(translation.ID)
	require.NoError(t, err)
	require.Equal(t, summary, translated.ShortDesc)
	require.Equal(t, loaded.MetaDesc, translated.MetaDesc)
	require.NoError(t, db.Delete(&product.Product{}, translation.ID).Error)
	references, err := NewMediaRepository(db).productRichTextMediaAttributeReferences(newMediaAssetReferenceQuery(&asset))
	require.NoError(t, err)
	require.Len(t, references, 2, "soft-deleted translations still retain summary images")
	states := map[uint]string{}
	for _, reference := range references {
		require.Equal(t, "short_description", reference.Field)
		states[reference.ResourceID] = reference.UsageState
	}
	require.Equal(t, "active", states[item.ID])
	require.Equal(t, "retained_deleted_parent", states[translation.ID])
	loaded.ShortDesc = ""
	require.NoError(t, repo.Update(loaded))
	require.NoError(t, db.Raw(`SELECT short_description,meta_description FROM products WHERE id=?`, item.ID).Row().Scan(&actualSummary, &actualSEO))
	require.Empty(t, actualSummary)
	require.Equal(t, "Updated exact SEO", actualSEO)
	t.Log("complete formal migration 435: canonical persistence, SEO update, five reads, translation and retained summary media verified without AutoMigrate")
}
