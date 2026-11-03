package localbackup

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
)

// Authentication and bounded extraction are shared; the caller still owns its
// domain eligibility, write set and restore transaction. Existing package and
// task formats, extraction limits and diagnostic strings remain unchanged.
func (manager *Manager) readAuthenticatedTransactionalBusinessModuleRestorePackage(task *Task, module string) (selectedBusinessModuleCompleteSnapshot, string, func(), error) {
	var empty selectedBusinessModuleCompleteSnapshot
	if task == nil || task.Manifest == nil || (module != "orders" && module != "customer_service" && module != "customers" && module != "media" && module != "configuration") || task.Manifest.Scope != selectedBusinessModulesBackupScope || !reflect.DeepEqual(task.Manifest.SelectedModules, []string{module}) {
		return empty, "", nil, errors.New("a standalone authenticated business module package is required")
	}
	directory, err := manager.taskDirectory(task.ID)
	if err != nil {
		return empty, "", nil, err
	}
	workspace, err := os.MkdirTemp(directory, "authenticated-business-module-restore-")
	if err != nil {
		return empty, "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(workspace) }
	packagePath, _ := manager.taskPackagePath(task.ID)
	identity, err := fileIdentity(packagePath, "")
	if err != nil || identity.Size != task.Size || identity.SHA256 != task.CiphertextSHA256 {
		cleanup()
		return empty, "", nil, errors.New("order package bytes changed")
	}
	passphrase, err := readProtectedCredential(manager.Configuration.PassphraseFile)
	if err != nil {
		cleanup()
		return empty, "", nil, err
	}
	extracted := filepath.Join(workspace, "verified")
	manifest, err := extractAndVerifyEncryptedPackageWithAllowedScopes(packagePath, extracted, passphrase, manager.Configuration.MaxExpandedBytes, manager.Configuration.MaxFiles, true, manager.Configuration.checkWorkspaceCapacity)
	if err != nil {
		cleanup()
		return empty, "", nil, err
	}
	if !reflect.DeepEqual(manifest, *task.Manifest) || len(manifest.Gaps) > 0 {
		cleanup()
		return empty, "", nil, errors.New("order package manifest changed or has gaps")
	}
	snapshot, err := manager.readValidatedSelectedBusinessModuleSnapshot(task, extracted)
	if err != nil {
		cleanup()
		return empty, "", nil, err
	}
	return snapshot, extracted, cleanup, nil
}
