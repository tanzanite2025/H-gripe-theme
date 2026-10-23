package localbackup

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

func WriteEncryptedPackage(destination string, sourceDirectory string, manifest Manifest, passphrase string, limit int64) (err error) {
	if err := validateWholeApplicationCompanyStaticManifestReleaseBinding(manifest); err != nil {
		return err
	}
	if err := validateExtractedWholeApplicationCompanyStaticReleaseMarker(manifest, sourceDirectory); err != nil {
		return err
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			file.Close()
			os.Remove(destination)
		}
	}()
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return err
	}
	encrypted, err := age.Encrypt(&boundedPackageWriter{destination: file, remaining: limit}, recipient)
	if err != nil {
		return err
	}
	compressed := gzip.NewWriter(encrypted)
	archive := tar.NewWriter(compressed)
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	if len(manifestBytes) > 64<<20 {
		return errors.New("manifest exceeds protocol capacity")
	}
	if err = archive.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(manifestBytes))}); err != nil {
		return err
	}
	if _, err = archive.Write(manifestBytes); err != nil {
		return err
	}
	for _, entry := range manifest.Files {
		if !isAllowedBackupFileForScope(manifest.Scope, entry.Path) {
			return errors.New("invalid manifest file path")
		}
		if err = archive.WriteHeader(&tar.Header{Name: entry.Path, Mode: 0600, Size: entry.Size}); err != nil {
			return err
		}
		source, openErr := os.Open(filepath.Join(sourceDirectory, filepath.FromSlash(entry.Path)))
		if openErr != nil {
			return openErr
		}
		hash := sha256.New()
		size, copyErr := io.Copy(io.MultiWriter(archive, hash), source)
		closeErr := source.Close()
		if copyErr != nil || closeErr != nil {
			return errors.Join(copyErr, closeErr)
		}
		if size != entry.Size || hex.EncodeToString(hash.Sum(nil)) != entry.SHA256 {
			return errors.New("source changed during package generation")
		}
	}
	if err = archive.Close(); err != nil {
		return err
	}
	if err = compressed.Close(); err != nil {
		return err
	}
	if err = encrypted.Close(); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	return file.Close()
}

func ExtractAndVerifyEncryptedPackage(source, destination, passphrase string, maxBytes int64, maxFiles int, capacityChecks ...func(int64) error) (Manifest, error) {
	return extractAndVerifyEncryptedPackageWithAllowedScopes(source, destination, passphrase, maxBytes, maxFiles, false, capacityChecks...)
}

func extractAndVerifyEncryptedPackageWithAllowedScopes(source, destination, passphrase string, maxBytes int64, maxFiles int, allowProductModule bool, capacityChecks ...func(int64) error) (Manifest, error) {
	var manifest Manifest
	file, err := os.Open(source)
	if err != nil {
		return manifest, err
	}
	defer file.Close()
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return manifest, err
	}
	identity.SetMaxWorkFactor(18)
	decrypted, err := age.Decrypt(file, identity)
	if err != nil {
		return manifest, errors.New("package authentication failed")
	}
	bufferedPlaintext := bufio.NewReader(decrypted)
	compressed, err := gzip.NewReader(bufferedPlaintext)
	if err != nil {
		return manifest, errors.New("invalid compressed package")
	}
	defer compressed.Close()
	compressed.Multistream(false)
	archive := tar.NewReader(compressed)
	header, err := archive.Next()
	if err != nil || header.Name != "manifest.json" || header.Typeflag != tar.TypeReg || header.Size < 1 || header.Size > 64<<20 {
		return manifest, errors.New("missing or oversized manifest")
	}
	manifestBytes, err := io.ReadAll(io.LimitReader(archive, (64<<20)+1))
	if err != nil {
		return manifest, err
	}
	if err := decodeStrictJSON(manifestBytes, &manifest); err != nil {
		return manifest, err
	}
	if manifest.FormatVersion != PackageVersion {
		return manifest, fmt.Errorf("unsupported package format version %d; this executor requires version %d", manifest.FormatVersion, PackageVersion)
	}
	if err := validateWholeApplicationCompanyStaticManifestReleaseBinding(manifest); err != nil {
		return manifest, err
	}
	if !safeTaskID.MatchString(manifest.PackageID) || (manifest.Scope != "whole_application" && !(allowProductModule && manifest.Scope == productCatalogBackupScope)) || len(manifest.Files) > maxFiles {
		return manifest, errors.New("unsupported package format or scope")
	}
	entries := make(map[string]FileIdentity, len(manifest.Files))
	var total int64
	for _, entry := range manifest.Files {
		if !isAllowedBackupFileForScope(manifest.Scope, entry.Path) || entry.Size < 0 || entry.Size > maxBytes-total || len(entry.SHA256) != 64 {
			return manifest, errors.New("invalid file identity or expanded capacity exceeded")
		}
		if entry.Path == "recovery-secrets.json" && entry.Size > 64<<10 {
			return manifest, errors.New("recovery material exceeds protocol capacity")
		}
		if _, exists := entries[entry.Path]; exists {
			return manifest, errors.New("duplicate manifest file")
		}
		entries[entry.Path] = entry
		total += entry.Size
	}
	if _, ok := entries["database.dump"]; manifest.Scope == "whole_application" && !ok {
		return manifest, errors.New("database dump missing")
	}
	if _, ok := entries["recovery-secrets.json"]; manifest.Scope == "whole_application" && !ok {
		return manifest, errors.New("recovery materials missing")
	}
	if _, ok := entries[productCatalogRowsFilename]; manifest.Scope == productCatalogBackupScope && !ok {
		return manifest, errors.New("product catalog data missing")
	}
	for _, check := range capacityChecks {
		if err := check(total); err != nil {
			return manifest, err
		}
	}
	if err := os.Mkdir(destination, 0700); err != nil {
		return manifest, err
	}
	succeeded := false
	defer func() {
		if !succeeded {
			os.RemoveAll(destination)
		}
	}()
	for {
		header, err = archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return manifest, err
		}
		entry, exists := entries[header.Name]
		if !exists || header.Typeflag != tar.TypeReg || header.Size != entry.Size {
			return manifest, errors.New("unexpected, duplicate or non-regular archive entry")
		}
		filename := filepath.Join(destination, filepath.FromSlash(entry.Path))
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return manifest, err
		}
		output, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return manifest, err
		}
		hash := sha256.New()
		size, copyErr := io.Copy(io.MultiWriter(output, hash), archive)
		syncErr, closeErr := output.Sync(), output.Close()
		if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
			return manifest, err
		}
		if size != entry.Size || hex.EncodeToString(hash.Sum(nil)) != entry.SHA256 {
			return manifest, fmt.Errorf("file verification failed: %s", entry.Path)
		}
		delete(entries, header.Name)
	}
	if len(entries) != 0 {
		return manifest, errors.New("package is missing manifest files")
	}
	// Drain both layers: tar EOF alone does not authenticate the final age frame
	// or verify gzip's checksum. Reject appended plaintext/concatenated streams.
	if remaining, err := io.Copy(io.Discard, compressed); err != nil || remaining != 0 {
		return manifest, errors.New("invalid archive trailer")
	}
	if remaining, err := io.Copy(io.Discard, bufferedPlaintext); err != nil || remaining != 0 {
		return manifest, errors.New("invalid encrypted package trailer")
	}
	if err := validateExtractedWholeApplicationCompanyStaticReleaseMarker(manifest, destination); err != nil {
		return manifest, err
	}
	succeeded = true
	return manifest, nil
}

func safePackagePath(value string) bool {
	if value == "database.dump" || value == "recovery-secrets.json" {
		return true
	}
	if strings.ContainsAny(value, "\\:\x00") || path.IsAbs(value) || path.Clean(value) != value {
		return false
	}
	for _, name := range mediaRoots {
		if strings.HasPrefix(value, "media/"+name+"/") {
			for _, component := range strings.Split(value, "/") {
				if component == "." || component == ".." || strings.HasPrefix(component, ".local-backup-") {
					return false
				}
			}
			return !strings.HasSuffix(value, "/")
		}
	}
	return false
}

type boundedPackageWriter struct {
	destination io.Writer
	remaining   int64
}

func (writer *boundedPackageWriter) Write(data []byte) (int, error) {
	if int64(len(data)) > writer.remaining {
		return 0, errors.New("encrypted package capacity exceeded")
	}
	written, err := writer.destination.Write(data)
	writer.remaining -= int64(written)
	return written, err
}
