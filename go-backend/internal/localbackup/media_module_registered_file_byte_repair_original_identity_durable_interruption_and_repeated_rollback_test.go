package localbackup

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestMediaModuleRegisteredFileByteRepairRetainsOriginalAcrossReplacementAndRepeatedRollback(t *testing.T) {
	for _, rootName := range []string{"uploads", "uploads_private"} {
		for _, scenario := range []string{"intent_only", "original_linked", "replacement_linked", "installed", "changed_original_backup", "missing_original_backup", "changed_current_bytes", "same_bytes_different_inode"} {
			t.Run(rootName+"/"+scenario, func(t *testing.T) {
				cfg := createIndependentRecoveryTestConfiguration(t)
				manager, err := NewManager(cfg)
				require.NoError(t, err)
				defer manager.Close()
				task, err := manager.createTask("upload", "synthetic-media-byte-repair")
				require.NoError(t, err)
				key := "registered/nested/file.png"
				packagePath := "media/" + rootName + "/" + key
				filename := filepath.Join(cfg.StorageRoots[rootName], filepath.FromSlash(key))
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
				require.NoError(t, os.WriteFile(filename, []byte("original damaged bytes"), 0600))
				original, err := fileIdentity(filename, packagePath)
				require.NoError(t, err)
				extracted := t.TempDir()
				source := filepath.Join(extracted, filepath.FromSlash(packagePath))
				require.NoError(t, os.MkdirAll(filepath.Dir(source), 0700))
				require.NoError(t, os.WriteFile(source, []byte("authenticated backup image"), 0600))
				wanted, err := fileIdentity(source, packagePath)
				require.NoError(t, err)
				task.Manifest = &Manifest{Scope: selectedBusinessModulesBackupScope, SelectedModules: []string{"media"}, Files: []FileIdentity{wanted}}
				task.MediaModuleRestore = &BusinessModuleTransactionalRestoreProgress{Files: []BusinessModuleTransactionalRestoreFile{{Identity: wanted, OriginalIdentity: &original}}}
				file := &task.MediaModuleRestore.Files[0]
				if scenario == "intent_only" {
					file.Started = true
				} else {
					checkpoint := func(point string) error {
						if scenario == "original_linked" && point == "media_original_backup_linked_before_acknowledgement" || scenario == "replacement_linked" && point == "media_file_replaced_before_acknowledgement" {
							return errors.New("synthetic interruption")
						}
						return nil
					}
					err := manager.installTransactionalBusinessModuleRestoreFiles(context.Background(), task, extracted, checkpoint)
					if scenario == "original_linked" || scenario == "replacement_linked" {
						require.ErrorContains(t, err, "synthetic interruption")
						require.True(t, file.Started)
						require.False(t, file.Installed)
					} else {
						require.NoError(t, err)
						require.True(t, file.Installed)
					}
				}
				backup := filepath.Join(cfg.StorageRoots[rootName], mediaModuleFileRepairStageRelativePath(task, "originals", key))
				switch scenario {
				case "changed_original_backup":
					require.NoError(t, os.WriteFile(backup, []byte("unrelated old file"), 0600))
				case "missing_original_backup":
					require.NoError(t, os.Remove(backup))
				case "changed_current_bytes":
					require.NoError(t, os.WriteFile(filename, []byte("unexpected current writer"), 0600))
				case "same_bytes_different_inode":
					require.NoError(t, os.Remove(filename))
					require.NoError(t, os.WriteFile(filename, []byte("authenticated backup image"), 0600))
				}
				err = manager.removeUncommittedTransactionalBusinessModuleRestoreFileAdditions(task)
				if strings.HasPrefix(scenario, "changed_") || scenario == "missing_original_backup" || scenario == "same_bytes_different_inode" {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.NoError(t, manager.removeUncommittedTransactionalBusinessModuleRestoreFileAdditions(task))
				actual, err := fileIdentity(filename, packagePath)
				require.NoError(t, err)
				require.Equal(t, original, actual)
				require.NoError(t, manager.removeOwnedTransactionalBusinessModuleAttachmentStaging(task))
			})
		}
	}
}

func TestMediaModuleRegisteredFileByteRepairRequiresTheSameCompleteCurrentRegistration(t *testing.T) {
	for _, scenario := range []string{"matching_registration", "changed_registration", "missing_registration", "duplicate_registration", "foreign_module", "unregistered_file", "company_static", "deleted_registration"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := createIndependentRecoveryTestConfiguration(t)
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer database.Close()
			mock.ExpectBegin()
			tx, err := database.BeginTx(context.Background(), nil)
			require.NoError(t, err)
			row := json.RawMessage(`{"id":42,"status":"active","visibility":"public","storage_key":"products/example.png"}`)
			if scenario == "deleted_registration" {
				row = json.RawMessage(`{"id":42,"status":"active","visibility":"public","storage_key":"products/example.png","deleted_at":"2026-01-01"}`)
			}
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"media"}, Tables: map[string]businessModuleCompleteTableSnapshot{"media_assets": {Rows: []json.RawMessage{row}}}}
			wanted := FileIdentity{Path: "media/uploads/products/example.png", Size: 10, SHA256: strings.Repeat("a", 64)}
			original := FileIdentity{Path: wanted.Path, Size: 8, SHA256: strings.Repeat("b", 64)}
			switch scenario {
			case "foreign_module":
				snapshot.Modules = []string{"orders"}
			case "unregistered_file":
				snapshot.Tables = nil
			case "company_static":
				wanted.Path, original.Path = "product-static/company/logo.png", "product-static/company/logo.png"
			case "duplicate_registration":
				snapshot.Tables["media_assets"] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{row, row}}
			}
			if scenario == "matching_registration" || scenario == "changed_registration" || scenario == "missing_registration" || scenario == "duplicate_registration" {
				count := 1
				if scenario == "duplicate_registration" {
					count = 2
				}
				for i := 0; i < count; i++ {
					mock.ExpectQuery(`SELECT EXISTS`).WithArgs(string(row)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(scenario != "changed_registration" && scenario != "missing_registration"))
				}
			}
			repair, err := cfg.inspectMediaModuleExistingRegisteredFileByteRepair(context.Background(), tx, snapshot, wanted, original)
			require.NoError(t, err)
			require.Equal(t, scenario == "matching_registration", repair != nil)
			mock.ExpectRollback()
			require.NoError(t, tx.Rollback())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMediaModuleRegisteredFileByteRepairOptionalJournalFieldKeepsOldJSONAndRejectsWrongModule(t *testing.T) {
	file := BusinessModuleTransactionalRestoreFile{Identity: FileIdentity{Path: "media/uploads/original.png", Size: 1, SHA256: strings.Repeat("a", 64)}}
	encoded, err := json.Marshal(file)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "original_identity")
	file.OriginalIdentity = &FileIdentity{Path: file.Identity.Path, Size: 2, SHA256: strings.Repeat("b", 64)}
	for _, scenario := range []string{"media", "order", "different_path", "same_identity", "negative_size", "invalid_hash"} {
		t.Run(scenario, func(t *testing.T) {
			copyFile := file
			original := *file.OriginalIdentity
			copyFile.OriginalIdentity = &original
			task := &Task{MediaModuleRestore: &BusinessModuleTransactionalRestoreProgress{}}
			switch scenario {
			case "order":
				task.MediaModuleRestore, task.OrderModuleRestore = nil, &BusinessModuleTransactionalRestoreProgress{}
			case "different_path":
				original.Path = "media/uploads/another.png"
			case "same_identity":
				original = file.Identity
			case "negative_size":
				original.Size = -1
			case "invalid_hash":
				original.SHA256 = "invalid"
			}
			require.Equal(t, scenario == "media", validateMediaModuleByteRepairJournalOriginalIdentity(task, copyFile) == nil)
		})
	}
}
