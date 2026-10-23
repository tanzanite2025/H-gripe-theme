package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestProductCanonicalDescriptionsPersistSearchTranslateAndProtectSummaryOnlyMedia(t *testing.T) {
	for _, useManager := range []bool{false, true} {
		t.Run(fmt.Sprintf("application_transaction_manager_%t", useManager), func(t *testing.T) {
			db, products, assets := createProductMediaTransactionalBindingTestServices(t, useManager)
			for _, column := range []string{"short_description", "meta_description"} {
				require.True(t, db.Migrator().HasColumn("products", column))
			}
			for _, column := range []string{"short_desc", "meta_desc"} {
				require.False(t, db.Migrator().HasColumn("products", column), "AutoMigrate must match the original SQL schema")
			}
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/summary-only.png")
			input := createProductInputForTransactionalMediaBindingTest()
			input.Status = "active"
			input.ShortDesc = `<p>SummariesOnlyTerm</p><img src="` + asset.URL + `">`
			created, err := products.CreateAdminProduct(input)
			require.NoError(t, err)
			var persistedSummary string
			require.NoError(t, db.Raw(`SELECT short_description FROM products WHERE id=?`, created.ID).Scan(&persistedSummary).Error)
			require.Equal(t, created.ShortDesc, persistedSummary)
			require.Contains(t, persistedSummary, "SummariesOnlyTerm")
			require.NoError(t, products.productRepo.UpdateSEO(created.ID, "Canonical SEO title", "Canonical SEO description"))
			loaded, err := products.GetAdminProduct(created.ID)
			require.NoError(t, err)
			require.Equal(t, created.ShortDesc, loaded.ShortDesc)
			require.Equal(t, "Canonical SEO description", loaded.MetaDesc)
			name := "Ordinary product name edit"
			loaded, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &name})
			require.NoError(t, err)
			require.Equal(t, created.ShortDesc, loaded.ShortDesc)
			require.Equal(t, "Canonical SEO description", loaded.MetaDesc)
			encoded, err := json.Marshal(loaded)
			require.NoError(t, err)
			var payload map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(encoded, &payload))
			require.Contains(t, payload, "short_description")
			require.Contains(t, payload, "meta_description")
			require.NotContains(t, payload, "short_desc")
			require.NotContains(t, payload, "meta_desc")
			for _, search := range []struct {
				name string
				run  func() ([]product.Product, error)
			}{
				{"public", func() ([]product.Product, error) {
					rows, _, err := products.productRepo.SearchPublic(repository.ProductSearchQuery{Locale: "en", Status: "active", Keyword: "summariesonlyterm", Limit: 10})
					return rows, err
				}},
				{"public_compact", func() ([]product.Product, error) {
					return products.productRepo.SearchPublicCompact(repository.ProductSearchQuery{Locale: "en", Status: "active", Keyword: "summariesonlyterm", Limit: 10})
				}},
				{"admin", func() ([]product.Product, error) {
					// Admin search intentionally searches names/descriptions/SKUs; summaries must
					// still be loaded without extending that established contract.
					rows, _, err := products.ListAdmin(1, 10, "active", "en", name, "", "", "")
					return rows, err
				}},
				{"quick_buy", func() ([]product.Product, error) {
					rows, _, err := products.productRepo.ListQuickBuyCandidates(repository.ProductQuickBuyCandidateQuery{Locale: "en", Keyword: "summariesonlyterm", Limit: 10})
					return rows, err
				}},
				{"recommendation", func() ([]product.Product, error) {
					rows, _, err := products.productRepo.ListRecommendationCandidates(repository.ProductRecommendationQuery{Locale: "en", Keyword: "summariesonlyterm", Limit: 10})
					return rows, err
				}},
			} {
				t.Run(search.name, func(t *testing.T) {
					rows, err := search.run()
					require.NoError(t, err)
					require.Len(t, rows, 1)
					require.Equal(t, created.ID, rows[0].ID)
					require.Equal(t, created.ShortDesc, rows[0].ShortDesc)
				})
			}
			translated, _, err := products.CopyAdminProductTranslation(created.ID, "fr")
			require.NoError(t, err)
			require.Equal(t, created.ShortDesc, translated.ShortDesc)
			report, err := assets.GetAssetReferences(asset.ID)
			require.NoError(t, err)
			require.Len(t, report.References, 2)
			for _, reference := range report.References {
				require.Equal(t, "short_description", reference.Field)
			}
			require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
			audit := NewHistoricalMediaReadOnlyInventoryService(db,
				&storage.Config{Type: storage.StorageTypeLocal, LocalPath: "unused-summary-audit", BaseURL: "https://media.example.test"},
				&storage.Config{Type: storage.StorageTypeLocal, LocalPath: "unused-separate-logo-audit", BaseURL: "https://logos.example.test"}, "https://shop.example.test")
			inventory, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
			require.NoError(t, err)
			found := false
			for _, source := range inventory.Sources {
				if source.Table == "products" && source.Field == "short_description" {
					require.False(t, source.Unavailable)
					require.True(t, source.Complete)
					require.Equal(t, 2, source.RowsScanned)
					found = true
				}
			}
			require.True(t, found)
			empty := ""
			cleared, err := products.UpdateAdminProduct(created.ID, ProductUpdateInput{ShortDesc: &empty})
			require.NoError(t, err)
			require.Empty(t, cleared.ShortDesc)
			require.NoError(t, db.Raw(`SELECT short_description FROM products WHERE id=?`, created.ID).Scan(&persistedSummary).Error)
			require.Empty(t, persistedSummary)
			require.Equal(t, "Canonical SEO description", cleared.MetaDesc)
		})
	}
}
