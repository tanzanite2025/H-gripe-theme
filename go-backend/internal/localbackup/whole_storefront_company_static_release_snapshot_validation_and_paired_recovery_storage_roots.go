package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const companyStaticPackagePrefix = "product-static/company/"
const companyStaticReleaseVersionFilename = ".storefront-release-version"
const companyStaticRecoveryRootName = "storefront_company_static_assets"

type wholeApplicationRecoveryStorageRoot struct {
	Name, Directory, PackagePrefix string
}

func (cfg Configuration) wholeApplicationRecoveryStorageRoots(includeCompanyStatic bool) []wholeApplicationRecoveryStorageRoot {
	roots := make([]wholeApplicationRecoveryStorageRoot, 0, 4)
	for _, name := range mediaRoots {
		roots = append(roots, wholeApplicationRecoveryStorageRoot{name, cfg.StorageRoots[name], "media/" + name + "/"})
	}
	if includeCompanyStatic && cfg.ProductStaticAssetsDirectory != "" {
		roots = append(roots, wholeApplicationRecoveryStorageRoot{companyStaticRecoveryRootName, filepath.Join(cfg.ProductStaticAssetsDirectory, "company"), companyStaticPackagePrefix})
	}
	return roots
}

func wholeApplicationPackageIncludesCompanyStatic(manifest *Manifest) bool {
	return manifest != nil && manifest.Scope == "whole_application" && manifest.StorefrontCompanyStaticReleaseVersion != ""
}

func (cfg Configuration) walkCompanyStaticReleaseRegularFiles(ctx context.Context, visit func(string, string, fs.FileInfo) error) error {
	root := filepath.Join(cfg.ProductStaticAssetsDirectory, "company")
	return filepath.WalkDir(root, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if strings.HasPrefix(entry.Name(), ".local-backup-") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("company static snapshot refuses symbolic links")
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("company static snapshot requires regular files")
		}
		relative, err := filepath.Rel(root, filename)
		if err != nil {
			return err
		}
		packagePath := companyStaticPackagePrefix + filepath.ToSlash(relative)
		if !isAllowedCompanyStaticPackagePath(packagePath) {
			return errors.New("company static filename is unsafe for portable recovery")
		}
		return visit(filename, packagePath, info)
	})
}

func isAllowedCompanyStaticPackagePath(filename string) bool {
	return strings.HasPrefix(filename, companyStaticPackagePrefix) && safePackagePath("media/uploads/"+strings.TrimPrefix(filename, "product-static/"))
}

func (cfg Configuration) validateCompanyStaticSourceReleaseVersion() error {
	if cfg.ProductStaticAssetsDirectory == "" {
		return nil
	}
	filename := filepath.Join(cfg.ProductStaticAssetsDirectory, "company", companyStaticReleaseVersionFilename)
	info, err := os.Lstat(filename)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 4096 {
		return errors.New("running storefront company static release marker is missing; initialize the matching release volume before export")
	}
	version, err := os.ReadFile(filename)
	if err != nil || strings.TrimSpace(string(version)) != cfg.ApplicationVersion {
		return errors.New("company static files belong to a different storefront release; export is blocked")
	}
	return nil
}

func (cfg Configuration) copyWholeApplicationCompanyStaticSnapshot(ctx context.Context, directory string, manifest *Manifest) error {
	if cfg.ProductStaticAssetsDirectory == "" {
		return nil
	}
	if err := cfg.validateCompanyStaticSourceReleaseVersion(); err != nil {
		return err
	}
	var total int64
	for _, file := range manifest.Files {
		total += file.Size
	}
	if err := cfg.walkCompanyStaticReleaseRegularFiles(ctx, func(source, packagePath string, info fs.FileInfo) error {
		if info.Size() > cfg.MaxExpandedBytes-total || len(manifest.Files) >= cfg.MaxFiles-2 {
			return errors.New("company static snapshot exceeds configured package capacity")
		}
		identity, err := copyVerifiedRegularFile(source, filepath.Join(directory, filepath.FromSlash(packagePath)), packagePath)
		if err != nil {
			return err
		}
		total += identity.Size
		manifest.Files = append(manifest.Files, identity)
		return nil
	}); err != nil {
		return err
	}
	manifest.StorefrontCompanyStaticReleaseVersion = cfg.ApplicationVersion
	return nil
}

func fingerprintCompanyStaticManifestFiles(files []FileIdentity) string {
	entries := make([]FileIdentity, 0)
	for _, entry := range files {
		if strings.HasPrefix(entry.Path, companyStaticPackagePrefix) {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	hash := sha256.New()
	for _, entry := range entries {
		fmt.Fprintf(hash, "%q:%d:%s\n", entry.Path, entry.Size, entry.SHA256)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (cfg Configuration) companyStaticReleaseFileFingerprint(ctx context.Context) (string, error) {
	var files []FileIdentity
	var total int64
	err := cfg.walkCompanyStaticReleaseRegularFiles(ctx, func(source, packagePath string, info fs.FileInfo) error {
		if info.Size() > cfg.MaxExpandedBytes-total || len(files) >= cfg.MaxFiles {
			return errors.New("company static inventory exceeds configured capacity")
		}
		identity, err := fileIdentity(source, packagePath)
		if err != nil {
			return err
		}
		total += identity.Size
		files = append(files, identity)
		return nil
	})
	if err != nil {
		return "", err
	}
	return fingerprintCompanyStaticManifestFiles(files), nil
}

func validateWholeApplicationCompanyStaticManifestReleaseBinding(manifest Manifest) error {
	if manifest.Scope != "whole_application" {
		if manifest.StorefrontCompanyStaticReleaseVersion != "" {
			return errors.New("company static release binding is reserved for whole-storefront packages")
		}
		return nil
	}
	hasStaticFiles, hasMarker := false, false
	for _, file := range manifest.Files {
		if strings.HasPrefix(file.Path, companyStaticPackagePrefix) {
			hasStaticFiles = true
		}
		if file.Path == companyStaticPackagePrefix+companyStaticReleaseVersionFilename && file.Size > 0 && file.Size <= 4096 {
			hasMarker = true
		}
	}
	if !hasStaticFiles && manifest.StorefrontCompanyStaticReleaseVersion == "" {
		return nil
	}
	if !hasStaticFiles || !hasMarker || manifest.StorefrontCompanyStaticReleaseVersion == "" || manifest.StorefrontCompanyStaticReleaseVersion != manifest.ApplicationVersion {
		return errors.New("company static files require a complete matching whole-storefront release binding and marker")
	}
	return nil
}

func validateExtractedWholeApplicationCompanyStaticReleaseMarker(manifest Manifest, extracted string) error {
	if !wholeApplicationPackageIncludesCompanyStatic(&manifest) {
		return nil
	}
	version, err := os.ReadFile(filepath.Join(extracted, filepath.FromSlash(companyStaticPackagePrefix+companyStaticReleaseVersionFilename)))
	if err != nil || strings.TrimSpace(string(version)) != manifest.StorefrontCompanyStaticReleaseVersion {
		return errors.New("whole-storefront company static release marker does not match the authenticated package")
	}
	return nil
}

func (cfg Configuration) validateWholeApplicationCompanyStaticPackage(manifest Manifest, extracted string) error {
	if err := validateWholeApplicationCompanyStaticManifestReleaseBinding(manifest); err != nil {
		return err
	}
	if manifest.Scope != "whole_application" {
		return nil
	}
	if manifest.StorefrontCompanyStaticReleaseVersion == "" {
		if cfg.ProductStaticAssetsDirectory != "" {
			return errors.New("whole-storefront package is missing its company static release binding; incomplete legacy packages cannot overwrite this target")
		}
		return nil
	}
	if cfg.ProductStaticAssetsDirectory == "" || manifest.StorefrontCompanyStaticReleaseVersion != cfg.ApplicationVersion {
		return errors.New("whole-storefront company static package requires its matching configured release volume")
	}
	return validateExtractedWholeApplicationCompanyStaticReleaseMarker(manifest, extracted)
}

func (cfg Configuration) wholeBackupBusinessReferencePackagePath(reference string) string {
	if root, key := cfg.resolveLocalMediaReference(reference); key != "" {
		return "media/" + root + "/" + key
	}
	_, filename, err := cfg.resolveProductCatalogImageSource(reference)
	if err == nil && isAllowedCompanyStaticPackagePath(filename) {
		return filename
	}
	return ""
}
