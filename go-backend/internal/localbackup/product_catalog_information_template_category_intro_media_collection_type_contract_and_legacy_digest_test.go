package localbackup

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductCatalogTemplateAndCategoryRichTextUseExactAttributesAndExcludePlainTextFromFileDiscovery(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	for _, source := range []struct{ table, column string }{{"product_information_templates", "content"}, {"product_categories", "seo_intro"}, {"product_category_translations", "seo_intro"}} {
		t.Run(source.table, func(t *testing.T) {
			row, err := json.Marshal(map[string]any{"id": 1, source.column: `<p>/uploads/not-a-file.png</p><img src='/uploads/photo.png?a=1&amp;b=2' alt='/uploads/alt.png'><video src='/uploads/video.mp4' poster='/uploads/photo.png'></video><a href='/uploads/manual#page=2'>Manual</a><a href='https://external.invalid/photo.jpg'>External</a><script><img src='/uploads/removed.png'></script>`, "name": "/uploads/name.png", "description": `<img src='/uploads/plain-description.png'>`, "meta_title": "/uploads/title.png", "meta_description": "/uploads/meta.png"})
			require.NoError(t, err)
			// Information templates have no description/meta columns. Only their
			// actual known text fields are part of this contract.
			if source.table == "product_information_templates" {
				row = copyProductCatalogRowWithJSONFieldsChanged(t, row, `{"description":null,"meta_title":null,"meta_description":null}`)
			}
			snapshot := productCatalogModuleSnapshot{Tables: map[string]productCatalogTableSnapshot{
				source.table:   {Rows: []json.RawMessage{row}},
				"media_assets": {Rows: []json.RawMessage{json.RawMessage(`{"id":1,"storage_key":"photo.png","media_type":"image"}`), json.RawMessage(`{"id":2,"storage_key":"video.mp4","media_type":"video"}`), json.RawMessage(`{"id":3,"storage_key":"manual","media_type":"document"}`)}},
			}}
			before, err := json.Marshal(snapshot)
			require.NoError(t, err)
			references, err := cfg.collectProductCatalogMediaReferences(snapshot)
			require.NoError(t, err)
			require.Equal(t, map[string]bool{"/uploads/photo.png?a=1&b=2": true, "/uploads/photo.png": true, "/uploads/video.mp4": true, "/uploads/manual#page=2": true}, references)
			require.NoError(t, cfg.validateProductCatalogCompleteRichTextRegisteredMediaTypes(snapshot))
			after, err := json.Marshal(snapshot)
			require.NoError(t, err)
			require.Equal(t, before, after)
			mutation, err := json.Marshal(map[string]string{source.column: `<img src='/uploads/video.mp4'>`})
			require.NoError(t, err)
			selected := snapshot.Tables[source.table]
			selected.Rows[0] = copyProductCatalogRowWithJSONFieldsChanged(t, row, string(mutation))
			snapshot.Tables[source.table] = selected
			require.ErrorContains(t, cfg.validateProductCatalogCompleteRichTextRegisteredMediaTypes(snapshot), source.table+"."+source.column+" img.src media type does not match")
			_, err = productCatalogPersistedRichTextMediaColumnsForTable(source.table, json.RawMessage(`{"id":1}`))
			require.ErrorContains(t, err, "missing its "+source.column+" column")
		})
	}
}

func TestProductCatalogPreviousRichTextPreviewPreservesExactRollbackDigestWithoutTemplateAndCategoryGuard(t *testing.T) {
	legacy := `{"version":1,"required_images":3,"inspected_registrations":2,"retained_cleanup_receipts":0,"required_paths_sha256":"paths","registration_rows_sha256":"registrations","cleanup_rows_sha256":"receipts","standard_write_guard_version":1,"media_placement_guard_version":1,"rich_text_media_guard_version":1}`
	var decoded productCatalogRequiredImageCleanupProtection
	require.NoError(t, json.Unmarshal([]byte(legacy), &decoded))
	encoded, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.Equal(t, legacy, string(encoded))
}
