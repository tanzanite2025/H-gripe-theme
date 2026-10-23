package localbackup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/require"
)

func createEncryptedPackageFixture(t *testing.T) (string, string, Manifest) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(source, 0700))
	manifest := Manifest{FormatVersion: PackageVersion, PackageID: newTaskID(), Scope: "whole_application", Files: []FileIdentity{}}
	for name, content := range map[string]string{"database.dump": "synthetic database only", "recovery-secrets.json": `{"JWT_SECRET":"synthetic-signing-key"}`, "media/uploads_private/conversations/customer-receipt.txt": "private attachment"} {
		filename := filepath.Join(source, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, []byte(content), 0600))
		entry, err := fileIdentity(filename, name)
		require.NoError(t, err)
		manifest.Files = append(manifest.Files, entry)
	}
	packagePath := filepath.Join(root, "backup.age")
	passphrase := strings.Repeat("independent-passphrase-", 3)
	require.NoError(t, WriteEncryptedPackage(packagePath, source, manifest, passphrase, 8<<20))
	return packagePath, passphrase, manifest
}

func TestEncryptedPackageRejectsCorruptionAndIncompleteTransfersWithoutLeavingExtractedSecrets(t *testing.T) {
	filename, passphrase, manifest := createEncryptedPackageFixture(t)
	encrypted, err := os.ReadFile(filename)
	require.NoError(t, err)
	destination := filepath.Join(t.TempDir(), "verified")
	restored, err := ExtractAndVerifyEncryptedPackage(filename, destination, passphrase, 1<<20, 100)
	require.NoError(t, err)
	require.Equal(t, manifest.Files, restored.Files)
	contents, err := os.ReadFile(filepath.Join(destination, "media", "uploads_private", "conversations", "customer-receipt.txt"))
	require.NoError(t, err)
	require.Equal(t, "private attachment", string(contents))
	for _, scenario := range []string{"wrong_password", "truncated", "changed_ciphertext", "expanded_capacity", "file_count"} {
		t.Run(scenario, func(t *testing.T) {
			data := append([]byte{}, encrypted...)
			key := passphrase
			maxBytes, maxFiles := int64(1<<20), 100
			switch scenario {
			case "wrong_password":
				key = strings.Repeat("wrong-password", 4)
			case "truncated":
				data = data[:len(data)-16]
			case "changed_ciphertext":
				data[len(data)-35] ^= 0x20
			case "expanded_capacity":
				maxBytes = 1
			case "file_count":
				maxFiles = 1
			}
			bad := filepath.Join(t.TempDir(), "bad.age")
			require.NoError(t, os.WriteFile(bad, data, 0600))
			output := filepath.Join(t.TempDir(), "rejected")
			_, err := ExtractAndVerifyEncryptedPackage(bad, output, key, maxBytes, maxFiles)
			require.Error(t, err)
			_, statErr := os.Stat(output)
			require.True(t, os.IsNotExist(statErr), "partial secrets must not survive failed authentication")
		})
	}
}

func TestAuthenticatedArchiveCannotEscapeStagingOrSmuggleLinksDuplicatesOrTrailingPayloads(t *testing.T) {
	_, passphrase, manifest := createEncryptedPackageFixture(t)
	for _, scenario := range []string{"parent_path", "absolute_path", "symlink", "duplicate", "missing_file", "trailing_plaintext", "previous_package_version"} {
		t.Run(scenario, func(t *testing.T) {
			var encrypted bytes.Buffer
			recipient, err := age.NewScryptRecipient(passphrase)
			require.NoError(t, err)
			writer, err := age.Encrypt(&encrypted, recipient)
			require.NoError(t, err)
			compressed := gzip.NewWriter(writer)
			archive := tar.NewWriter(compressed)
			modified := manifest
			modified.Files = append([]FileIdentity{}, manifest.Files...)
			if scenario == "previous_package_version" {
				modified.FormatVersion = 1
			}
			if scenario == "parent_path" {
				modified.Files[0].Path = "media/uploads/../../../outside"
			}
			if scenario == "absolute_path" {
				modified.Files[0].Path = "C:/Windows/escaped"
			}
			data, err := json.Marshal(modified)
			require.NoError(t, err)
			require.NoError(t, archive.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(data))}))
			_, err = archive.Write(data)
			require.NoError(t, err)
			if scenario == "symlink" {
				require.NoError(t, archive.WriteHeader(&tar.Header{Name: manifest.Files[0].Path, Typeflag: tar.TypeSymlink, Linkname: "../../../outside"}))
			}
			if scenario == "duplicate" || scenario == "trailing_plaintext" || scenario == "missing_file" {
				contents := map[string]string{"database.dump": "synthetic database only", "recovery-secrets.json": `{"JWT_SECRET":"synthetic-signing-key"}`, "media/uploads_private/conversations/customer-receipt.txt": "private attachment"}
				for index, entry := range manifest.Files {
					if scenario == "missing_file" && index == len(manifest.Files)-1 {
						continue
					}
					require.NoError(t, archive.WriteHeader(&tar.Header{Name: entry.Path, Mode: 0600, Size: entry.Size}))
					_, err = io.WriteString(archive, contents[entry.Path])
					require.NoError(t, err)
				}
				if scenario == "duplicate" {
					entry := manifest.Files[0]
					require.NoError(t, archive.WriteHeader(&tar.Header{Name: entry.Path, Mode: 0600, Size: entry.Size}))
					_, err = io.WriteString(archive, contents[entry.Path])
					require.NoError(t, err)
				}
			}
			require.NoError(t, archive.Close())
			require.NoError(t, compressed.Close())
			if scenario == "trailing_plaintext" {
				_, err = io.WriteString(writer, "unexpected second payload")
				require.NoError(t, err)
			}
			require.NoError(t, writer.Close())
			filename := filepath.Join(t.TempDir(), "hostile.age")
			require.NoError(t, os.WriteFile(filename, encrypted.Bytes(), 0600))
			_, err = ExtractAndVerifyEncryptedPackage(filename, filepath.Join(t.TempDir(), "staging"), passphrase, 1<<20, 100)
			require.Error(t, err)
			if scenario == "previous_package_version" {
				require.Contains(t, err.Error(), "requires version 2")
			}
		})
	}
}
