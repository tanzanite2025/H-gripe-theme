package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lib/pq"
)

func (cfg Configuration) copyMediaSnapshot(ctx context.Context, directory string, manifest *Manifest) error {
	var total int64
	for _, name := range mediaRoots {
		root := cfg.StorageRoots[name]
		err := filepath.WalkDir(root, func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if filename == root {
				return nil
			}
			if strings.HasPrefix(entry.Name(), ".local-backup-") {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return errors.New("media snapshot refuses symbolic links")
			}
			if entry.IsDir() {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return errors.New("media snapshot refuses non-regular files")
			}
			relative, err := filepath.Rel(root, filename)
			if err != nil {
				return err
			}
			packagePath := "media/" + name + "/" + filepath.ToSlash(relative)
			if !safePackagePath(packagePath) {
				return errors.New("media filename is unsafe for portable restore")
			}
			if info.Size() > cfg.MaxExpandedBytes-total || len(manifest.Files) >= cfg.MaxFiles-2 {
				return errors.New("snapshot exceeds configured expanded capacity")
			}
			total += info.Size()
			identity, err := copyVerifiedRegularFile(filename, filepath.Join(directory, filepath.FromSlash(packagePath)), packagePath)
			if err != nil {
				return err
			}
			manifest.Files = append(manifest.Files, identity)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func copyVerifiedRegularFile(source, destination, packagePath string) (FileIdentity, error) {
	input, err := os.Open(source)
	if err != nil {
		return FileIdentity{}, err
	}
	defer input.Close()
	before, err := input.Stat()
	if err != nil || !before.Mode().IsRegular() {
		return FileIdentity{}, errors.New("source is not a regular file")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return FileIdentity{}, err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return FileIdentity{}, err
	}
	hash := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(output, hash), io.LimitReader(input, before.Size()+1))
	syncErr, closeErr := output.Sync(), output.Close()
	if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
		return FileIdentity{}, err
	}
	after, err := input.Stat()
	if err != nil || size != before.Size() || !after.ModTime().Equal(before.ModTime()) || after.Size() != before.Size() {
		return FileIdentity{}, errors.New("file changed while taking snapshot")
	}
	return FileIdentity{Path: packagePath, Size: size, SHA256: hex.EncodeToString(hash.Sum(nil))}, nil
}

// Database registrations and historical attachment JSON are inspected while
// writers are fenced. Unknown URLs and malformed JSON are explicit gaps, never
// silently accepted as a complete recoverable snapshot.
func (cfg Configuration) findMediaReferenceGaps(ctx context.Context, db *sql.DB, manifest *Manifest) error {
	files := make(map[string]bool, len(manifest.Files))
	identities := make(map[string]FileIdentity, len(manifest.Files))
	for _, entry := range manifest.Files {
		files[entry.Path] = true
		identities[entry.Path] = entry
	}
	if err := cfg.findAdditionalDeclaredMediaReferenceGaps(ctx, db, manifest, files); err != nil {
		return err
	}
	recordGap := func(code, table string, id int64) {
		if len(manifest.Gaps) < 10000 {
			manifest.Gaps = append(manifest.Gaps, fmt.Sprintf("%s:%s:%d", code, table, id))
		}
	}
	if _, exists := manifest.Tables["media_assets"]; exists {
		rows, err := db.QueryContext(ctx, `SELECT id,coalesce(storage_key,''),coalesce(visibility,'public'),coalesce(url,''),coalesce(to_jsonb(asset)->>'content_sha256',''),coalesce((to_jsonb(asset)->>'size')::bigint,0) FROM media_assets asset`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id int64
			var key, visibility, reference, expectedHash string
			var expectedSize int64
			if err := rows.Scan(&id, &key, &visibility, &reference, &expectedHash, &expectedSize); err != nil {
				rows.Close()
				return err
			}
			root := registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility)
			if key == "" {
				root, key = cfg.resolveLocalMediaReference(reference)
			}
			if key == "" || !files["media/"+root+"/"+key] {
				recordGap("missing_or_unresolved_media_object", "media_assets", id)
			} else if expectedHash != "" || expectedSize > 0 {
				entry := identities["media/"+root+"/"+key]
				if (expectedHash != "" && entry.SHA256 != strings.ToLower(expectedHash)) || (expectedSize > 0 && entry.Size != expectedSize) {
					recordGap("registered_original_content_mismatch", "media_assets", id)
				}
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	for _, source := range []struct{ table, root string }{{"order_evidence_attachments", "uploads_private"}, {"site_logo_assets", "site_logo_uploads"}, {"site_favicon_assets", "uploads"}, {"visual_showcase_items", "uploads"}} {
		if _, exists := manifest.Tables[source.table]; !exists {
			continue
		}
		rows, err := db.QueryContext(ctx, `SELECT id,coalesce(storage_key,''),coalesce(to_jsonb(asset)->>'sha256',to_jsonb(asset)->>'content_sha256',''),coalesce((to_jsonb(asset)->>'size_bytes')::bigint,(to_jsonb(asset)->>'size')::bigint,0) FROM public.`+pq.QuoteIdentifier(source.table)+` asset`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id, size int64
			var key, digest string
			if err := rows.Scan(&id, &key, &digest, &size); err != nil {
				rows.Close()
				return err
			}
			objectPath := "media/" + source.root + "/" + key
			if key == "" || !files[objectPath] {
				recordGap("missing_registered_business_attachment", source.table, id)
				continue
			}
			entry := identities[objectPath]
			if (digest != "" && entry.SHA256 != strings.ToLower(digest)) || (size > 0 && entry.Size != size) {
				recordGap("business_attachment_content_mismatch", source.table, id)
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	if _, exists := manifest.Tables["media_asset_derivatives"]; exists {
		rows, err := db.QueryContext(ctx, `SELECT d.id,coalesce(d.storage_key,''),coalesce(a.visibility,'public') FROM media_asset_derivatives d LEFT JOIN media_assets a ON a.id=d.media_asset_id`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id int64
			var key, visibility string
			if err := rows.Scan(&id, &key, &visibility); err != nil {
				rows.Close()
				return err
			}
			root := registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility)
			if key == "" || !files["media/"+root+"/"+key] {
				recordGap("missing_derivative", "media_asset_derivatives", id)
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	// All columns named attachments are included; this covers retained/deleted
	// messages as well. No application default scopes filter historical rows.
	columns, err := db.QueryContext(ctx, `SELECT table_name FROM information_schema.columns WHERE table_schema='public' AND column_name='attachments' AND data_type IN ('text','character varying','json','jsonb') ORDER BY table_name`)
	if err != nil {
		return err
	}
	var tables []string
	for columns.Next() {
		var table string
		if err := columns.Scan(&table); err != nil {
			columns.Close()
			return err
		}
		tables = append(tables, table)
	}
	err = columns.Err()
	columns.Close()
	if err != nil {
		return err
	}
	for _, table := range tables {
		rows, err := db.QueryContext(ctx, `SELECT id,attachments::text FROM public.`+pq.QuoteIdentifier(table)+` WHERE attachments IS NOT NULL AND btrim(attachments::text) NOT IN ('','null','[]')`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id int64
			var raw string
			if err := rows.Scan(&id, &raw); err != nil {
				rows.Close()
				return err
			}
			var refs []string
			if json.Unmarshal([]byte(raw), &refs) != nil {
				if table == "suggestion_feedback" {
					var attachments []struct {
						URL string `json:"url"`
					}
					if json.Unmarshal([]byte(raw), &attachments) == nil {
						for _, attachment := range attachments {
							refs = append(refs, attachment.URL)
						}
					} else {
						recordGap("malformed_attachment_json", table, id)
						continue
					}
				} else {
					recordGap("malformed_attachment_json", table, id)
					continue
				}
			}
			for _, reference := range refs {
				packagePath := cfg.wholeBackupBusinessReferencePackagePath(reference)
				if packagePath == "" || !files[packagePath] {
					recordGap("unresolved_attachment_reference", table, id)
				}
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Logo uploads have their own storage binding. Favicon uploads are owned by
// SiteFaviconService but stored in shared uploads, just like the live API.
func registeredMediaStorageRootForObjectKeyAndVisibility(key, visibility string) string {
	if strings.HasPrefix(key, "site-logo/") {
		return "site_logo_uploads"
	}
	if visibility == "private" {
		return "uploads_private"
	}
	return "uploads"
}

func (cfg Configuration) resolveLocalMediaReference(reference string) (string, string) {
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil {
		return "", ""
	}
	if parsed.Host != "" {
		recognized := false
		for _, origin := range cfg.PublicOrigins {
			configured, err := url.Parse(origin)
			if err == nil && strings.EqualFold(configured.Host, parsed.Host) && configured.Scheme == parsed.Scheme {
				recognized = true
				break
			}
		}
		if !recognized {
			return "", ""
		}
	}
	for _, prefix := range [][2]string{{"/uploads/site-logo/", "site_logo_uploads"}, {"/uploads/", "uploads"}, {"/uploads-private/", "uploads_private"}, {"/site-logo-uploads/", "site_logo_uploads"}} {
		if strings.HasPrefix(parsed.Path, prefix[0]) {
			key := strings.TrimPrefix(parsed.Path, prefix[0])
			if prefix[0] == "/uploads/site-logo/" {
				key = "site-logo/" + key
			}
			if safePackagePath("media/" + prefix[1] + "/" + key) {
				return prefix[1], key
			}
		}
	}
	return "", ""
}

func (cfg Configuration) mediaFingerprint(ctx context.Context) (string, error) {
	manifest := Manifest{}
	for _, name := range mediaRoots {
		err := filepath.WalkDir(cfg.StorageRoots[name], func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if filename == cfg.StorageRoots[name] {
				return nil
			}
			if strings.HasPrefix(entry.Name(), ".local-backup-") {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return errors.New("unexpected media symlink")
			}
			relative, err := filepath.Rel(cfg.StorageRoots[name], filename)
			if err != nil {
				return err
			}
			identity, err := fileIdentity(filename, "media/"+name+"/"+filepath.ToSlash(relative))
			if err != nil {
				return err
			}
			manifest.Files = append(manifest.Files, identity)
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	// WalkDir sorts each directory separately. Its depth-first order differs
	// from full-path order when both "name.ext" and "name/child.ext" exist.
	// Use the package's canonical ordering for every live/candidate comparison.
	return fingerprintManifestMedia(manifest), nil
}

func (cfg Configuration) checkWorkspaceCapacity(required int64) error {
	var used int64
	err := filepath.WalkDir(cfg.Workspace, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("workspace contains a symlink")
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		used += info.Size()
		return nil
	})
	if err != nil {
		return err
	}
	if required < 0 || used > cfg.MaxWorkspaceBytes-required {
		return errors.New("private backup workspace capacity exceeded")
	}
	free, err := availableDiskBytes(cfg.Workspace)
	if err != nil {
		return err
	}
	if uint64(required) > free || free-uint64(required) < 64<<20 {
		return errors.New("insufficient free workspace disk space")
	}
	return nil
}

func (cfg Configuration) sortedManifestFiles(manifest *Manifest) {
	sort.Slice(manifest.Files, func(i, j int) bool { return manifest.Files[i].Path < manifest.Files[j].Path })
}

func (cfg Configuration) estimateBackupWorkspaceBytes(ctx context.Context) (int64, error) {
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var total int64
	if err := db.QueryRowContext(ctx, `SELECT pg_database_size(current_database())`).Scan(&total); err != nil {
		return 0, err
	}
	for _, root := range cfg.wholeApplicationRecoveryStorageRoots(cfg.ProductStaticAssetsDirectory != "") {
		err := filepath.WalkDir(root.Directory, func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if filename == root.Directory {
				return nil
			}
			if strings.HasPrefix(entry.Name(), ".local-backup-") {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if info.Size() > cfg.MaxExpandedBytes-total {
				return errors.New("database and media exceed configured snapshot capacity")
			}
			total += info.Size()
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	if total > cfg.MaxExpandedBytes {
		return 0, errors.New("database and media exceed configured snapshot capacity")
	}
	return total*2 + (64 << 20), nil
}

func fingerprintManifestMedia(manifest Manifest) string {
	hash := sha256.New()
	for _, root := range mediaRoots {
		prefix := "media/" + root + "/"
		var entries []FileIdentity
		for _, entry := range manifest.Files {
			if strings.HasPrefix(entry.Path, prefix) {
				entries = append(entries, entry)
			}
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
		for _, entry := range entries {
			fmt.Fprintf(hash, "%q:%d:%s\n", strings.TrimPrefix(entry.Path, "media/"), entry.Size, entry.SHA256)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (cfg Configuration) findAdditionalDeclaredMediaReferenceGaps(ctx context.Context, db *sql.DB, manifest *Manifest, files map[string]bool) error {
	knownStorageTables := map[string]bool{"media_assets": true, "media_asset_derivatives": true, "order_evidence_attachments": true, "site_logo_assets": true, "site_favicon_assets": true, "visual_showcase_items": true}
	columns, err := db.QueryContext(ctx, `SELECT table_name FROM information_schema.columns WHERE table_schema='public' AND column_name='storage_key' ORDER BY table_name`)
	if err != nil {
		return err
	}
	for columns.Next() {
		var table string
		if err := columns.Scan(&table); err != nil {
			columns.Close()
			return err
		}
		if !knownStorageTables[table] {
			manifest.Gaps = append(manifest.Gaps, "unregistered_storage_key_source:"+table)
		}
	}
	readErr := columns.Err()
	columns.Close()
	if readErr != nil {
		return readErr
	}
	for _, source := range [][2]string{{"after_sales_case_attachments", "storage_url"}, {"product_media", "url"}, {"product_media", "thumbnail_url"}, {"product_spec_options", "image_url"}, {"product_spec_option_items", "swatch_url"}, {"posts", "featured_image"}, {"posts", "featured_img"}, {"faqs", "answer_image_url"}, {"galleries", "cover_image"}, {"gallery_images", "url"}, {"gallery_images", "thumbnail"}, {"workbench_feed_media", "file_path"}, {"visual_showcase_items", "thumbnail_url"}} {
		if _, exists := manifest.Tables[source[0]]; !exists {
			continue
		}
		rows, err := db.QueryContext(ctx, `SELECT id,to_jsonb(record)->>$1 FROM public.`+pq.QuoteIdentifier(source[0])+` record WHERE coalesce(to_jsonb(record)->>$1,'')<>''`, source[1])
		if err != nil {
			return err
		}
		for rows.Next() {
			var id int64
			var reference string
			if err := rows.Scan(&id, &reference); err != nil {
				rows.Close()
				return err
			}
			packagePath := cfg.wholeBackupBusinessReferencePackagePath(reference)
			if packagePath == "" || !files[packagePath] {
				if len(manifest.Gaps) < 10000 {
					manifest.Gaps = append(manifest.Gaps, fmt.Sprintf("unresolved_business_media_reference:%s.%s:%d", source[0], source[1], id))
				}
			}
		}
		readErr := rows.Err()
		rows.Close()
		if readErr != nil {
			return readErr
		}
	}
	return nil
}
