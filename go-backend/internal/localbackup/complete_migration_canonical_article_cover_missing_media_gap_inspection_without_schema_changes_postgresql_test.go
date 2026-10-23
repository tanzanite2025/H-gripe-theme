package localbackup

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func verifyCompleteMigrationCanonicalArticleCoverMediaGapsWithoutSchemaChanges(t *testing.T, cfg Configuration, db *sql.DB) {
	t.Helper()
	ctx := t.Context()
	const key = "catalog/synthetic-formal-article-cover.png"
	var postID int64
	require.NoError(t, db.QueryRowContext(ctx, `INSERT INTO posts(title,slug,author_id,featured_image,status,deleted_at) SELECT 'Synthetic retained article','synthetic-canonical-cover-gap',id,$1,'draft',now() FROM users WHERE username='synthetic-row-check' RETURNING id`, "/uploads/"+key).Scan(&postID))
	defer func() {
		_, err := db.ExecContext(ctx, `DELETE FROM posts WHERE id=$1`, postID)
		require.NoError(t, err)
	}()
	for _, included := range []bool{false, true} {
		name := "canonical_article_cover_missing_bytes_are_reported"
		if included {
			name = "canonical_article_cover_included_bytes_resolve_gap"
		}
		t.Run(name, func(t *testing.T) {
			manifest := Manifest{Tables: map[string]int64{"posts": 1}}
			files := map[string]bool{}
			if included {
				files["media/uploads/"+key] = true
			}
			require.NoError(t, cfg.findAdditionalDeclaredMediaReferenceGaps(t.Context(), db, &manifest, files))
			if included {
				require.Empty(t, manifest.Gaps)
			} else {
				require.Equal(t, []string{fmt.Sprintf("unresolved_business_media_reference:posts.featured_image:%d", postID)}, manifest.Gaps)
			}
			var legacyColumnExists bool
			require.NoError(t, db.QueryRowContext(t.Context(), `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='posts' AND column_name='featured_img')`).Scan(&legacyColumnExists))
			require.False(t, legacyColumnExists, "gap inspection must not create a legacy compatibility column")
		})
	}
}
