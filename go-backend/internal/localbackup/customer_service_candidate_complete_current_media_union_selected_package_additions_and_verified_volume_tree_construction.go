package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const customerServiceCandidateMediaInventoryFilename = "customer-service-candidate-complete-media-union-inventory-v1.json"
const customerServiceCandidateMediaVolumeIdentityFilename = "customer-service-candidate-media-volume-identity-v1.json"
const customerServiceCandidateMediaMaximumInventoryBytes = 64 << 20
const customerServiceCandidateMediaMaximumEntries = 100000

type customerServiceCandidateMediaFile struct {
	RootName string       `json:"root_name"`
	Key      string       `json:"key"`
	Origin   string       `json:"origin"`
	Identity FileIdentity `json:"identity"`
}

type customerServiceCandidateMediaDirectory struct {
	RootName string `json:"root_name"`
	Key      string `json:"key"`
}

type customerServiceCandidateCompleteMediaUnionInventory struct {
	Version                   int                                      `json:"version"`
	PlanSHA256                string                                   `json:"plan_sha256"`
	SelectionSHA256           string                                   `json:"selection_sha256"`
	BaselineMediaFingerprint  string                                   `json:"baseline_media_fingerprint"`
	BaselineTreeSHA256        string                                   `json:"baseline_tree_sha256"`
	CandidateMediaFingerprint string                                   `json:"candidate_media_fingerprint"`
	Files                     []customerServiceCandidateMediaFile      `json:"files"`
	Directories               []customerServiceCandidateMediaDirectory `json:"directories"`
	CurrentBytes              int64                                    `json:"current_bytes"`
	TotalBytes                int64                                    `json:"total_bytes"`
	AddedPackageFiles         int                                      `json:"added_package_files"`
	EncodedBytes              int                                      `json:"-"`
	SHA256                    string                                   `json:"sha256"`
}

type customerServiceCandidateMediaVolumeState struct {
	RootName    string `json:"root_name"`
	StorageRoot string `json:"storage_root"`
	State       string `json:"state"`
}

type customerServiceCandidateCompleteMediaUnionLifecycle struct {
	Version                   int                                        `json:"version"`
	State                     string                                     `json:"state"`
	InventorySHA256           string                                     `json:"inventory_sha256"`
	BaselineMediaFingerprint  string                                     `json:"baseline_media_fingerprint"`
	CandidateMediaFingerprint string                                     `json:"candidate_media_fingerprint"`
	CurrentFiles              int                                        `json:"current_files"`
	AddedPackageFiles         int                                        `json:"added_package_files"`
	TotalBytes                int64                                      `json:"total_bytes"`
	Volumes                   []customerServiceCandidateMediaVolumeState `json:"volumes"`
}

type customerServiceCandidateMediaVolumeIdentity struct {
	Version     int    `json:"version"`
	TaskID      string `json:"task_id"`
	AttemptID   string `json:"attempt_id"`
	PlanSHA256  string `json:"plan_sha256"`
	RootName    string `json:"root_name"`
	StorageRoot string `json:"storage_root"`
}

func customerServiceCandidateMediaPathComponents(rootName, key string) (string, error) {
	validRoot := false
	for _, name := range mediaRoots {
		if rootName == name {
			validRoot = true
		}
	}
	path := "media/" + rootName + "/" + key
	if !validRoot || key == "" || !safePackagePath(path) {
		return "", errors.New("candidate media path is unsafe")
	}
	for _, component := range strings.Split(key, "/") {
		if strings.HasPrefix(component, ".local-backup-") {
			return "", errors.New("candidate media references a reserved maintenance path")
		}
	}
	return path, nil
}

// Bound both the byte count and all entries, including empty directories. Only
// root-level maintenance artifacts are excluded; nested reserved names refuse
// construction rather than disappearing from the complete current baseline.
func (cfg Configuration) inventoryCustomerServiceCompleteCurrentMedia(ctx context.Context) (*customerServiceCandidateCompleteMediaUnionInventory, error) {
	return cfg.inventoryCustomerServiceMediaTreesWithMaintenanceArtifactPolicy(ctx, "exclude_live_root_maintenance_artifacts")
}

func (cfg Configuration) inventoryCustomerServiceMediaTreesWithMaintenanceArtifactPolicy(ctx context.Context, policy string) (*customerServiceCandidateCompleteMediaUnionInventory, error) {
	if policy != "exclude_live_root_maintenance_artifacts" && policy != "reject_all_maintenance_artifacts" {
		return nil, errors.New("candidate media inventory policy is invalid")
	}
	inventory := &customerServiceCandidateCompleteMediaUnionInventory{Version: 1, Files: []customerServiceCandidateMediaFile{}, Directories: []customerServiceCandidateMediaDirectory{}}
	for _, rootName := range mediaRoots {
		directory := cfg.StorageRoots[rootName]
		info, err := os.Lstat(directory)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("candidate media storage root is unsafe")
		}
		root, err := os.OpenRoot(directory)
		if err != nil {
			return nil, err
		}
		err = fs.WalkDir(root.FS(), ".", func(key string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if key == "." {
				return nil
			}
			if policy == "exclude_live_root_maintenance_artifacts" && !strings.Contains(key, "/") && strings.HasPrefix(entry.Name(), ".local-backup-") {
				if entry.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			path, err := customerServiceCandidateMediaPathComponents(rootName, key)
			if err != nil {
				return err
			}
			if err := accountCustomerServiceCandidateMediaInventoryMetadata(inventory, path); err != nil {
				return err
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return errors.New("candidate media inventory refuses symlinks")
			}
			if len(inventory.Files)+len(inventory.Directories) >= min(cfg.MaxFiles, customerServiceCandidateMediaMaximumEntries) {
				return errors.New("candidate complete media entry limit exceeded")
			}
			if entry.IsDir() {
				inventory.Directories = append(inventory.Directories, customerServiceCandidateMediaDirectory{rootName, key})
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return errors.New("candidate media inventory refuses non-regular files")
			}
			identity, err := inspectBoundedModuleRecoveryRegularFileIdentity(ctx, directory, key, cfg.MaxExpandedBytes-inventory.TotalBytes)
			if err != nil {
				return err
			}
			identity.Path = path
			inventory.TotalBytes += identity.Size
			inventory.Files = append(inventory.Files, customerServiceCandidateMediaFile{rootName, key, "current_target", identity})
			return nil
		})
		root.Close()
		if err != nil {
			return nil, err
		}
	}
	inventory.CurrentBytes = inventory.TotalBytes
	sortCustomerServiceCandidateMediaInventory(inventory)
	return inventory, nil
}

func sortCustomerServiceCandidateMediaInventory(inventory *customerServiceCandidateCompleteMediaUnionInventory) {
	sort.Slice(inventory.Files, func(i, j int) bool { return inventory.Files[i].Identity.Path < inventory.Files[j].Identity.Path })
	sort.Slice(inventory.Directories, func(i, j int) bool {
		a, b := inventory.Directories[i], inventory.Directories[j]
		return a.RootName+"/"+a.Key < b.RootName+"/"+b.Key
	})
}

func accountCustomerServiceCandidateMediaInventoryMetadata(inventory *customerServiceCandidateCompleteMediaUnionInventory, path string) error {
	if len(path) > 4096 {
		return errors.New("candidate media portable path length limit exceeded")
	}
	estimate := len(path)*12 + 1024
	if inventory.EncodedBytes > customerServiceCandidateMediaMaximumInventoryBytes-estimate {
		return errors.New("candidate media inventory exceeds its metadata limit")
	}
	inventory.EncodedBytes += estimate
	return nil
}

func fingerprintCustomerServiceCandidateMediaInventory(inventory *customerServiceCandidateCompleteMediaUnionInventory) string {
	manifest := Manifest{}
	for _, entry := range inventory.Files {
		manifest.Files = append(manifest.Files, entry.Identity)
	}
	return fingerprintManifestMedia(manifest)
}

func (cfg Configuration) planCustomerServiceCompleteCurrentMediaUnion(ctx context.Context, selection *CustomerServiceEffectiveSelectionInspection, manifest *Manifest, verifiedDirectory string) (*customerServiceCandidateCompleteMediaUnionInventory, []byte, error) {
	if selection == nil || !selection.SelectionDependenciesSatisfied || manifest == nil {
		return nil, nil, errors.New("candidate media union requires complete saved selections")
	}
	inventory, err := cfg.inventoryCustomerServiceCompleteCurrentMedia(ctx)
	if err != nil {
		return nil, nil, err
	}
	baselineDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(inventory)
	if err != nil {
		return nil, nil, err
	}
	inventory.BaselineTreeSHA256 = baselineDigest
	inventory.BaselineMediaFingerprint = fingerprintCustomerServiceCandidateMediaInventory(inventory)
	if inventory.BaselineMediaFingerprint != selection.Binding.TargetMediaFingerprint {
		return nil, nil, errors.New("complete current media changed since the stopped selection")
	}
	inventory.PlanSHA256, inventory.SelectionSHA256 = selection.PlanSHA256, selection.SHA256
	paths := map[string]FileIdentity{}
	hashPaths := map[string]string{}
	aliases := map[string]string{}
	directories := map[string]bool{}
	registerPath := func(path string, isDirectory bool) error {
		alias := strings.ToLower(path)
		if existing, ok := aliases[alias]; ok && existing != path {
			return errors.New("candidate media paths have a portable case collision")
		}
		aliases[alias] = path
		if isDirectory {
			if _, exists := paths[path]; exists {
				return errors.New("candidate media file and directory paths collide")
			}
			directories[path] = true
		} else if directories[path] {
			return errors.New("candidate media file and directory paths collide")
		}
		return nil
	}
	for _, directory := range inventory.Directories {
		path, _ := customerServiceCandidateMediaPathComponents(directory.RootName, directory.Key)
		if err := registerPath(path, true); err != nil {
			return nil, nil, err
		}
	}
	for _, entry := range inventory.Files {
		if err := registerPath(entry.Identity.Path, false); err != nil {
			return nil, nil, err
		}
		paths[entry.Identity.Path] = entry.Identity
		digest := sha256.Sum256([]byte(entry.Identity.Path))
		hashPaths[hex.EncodeToString(digest[:])] = entry.Identity.Path
	}
	packageFiles := map[string]FileIdentity{}
	for _, entry := range manifest.Files {
		if !strings.HasPrefix(entry.Path, "media/") {
			continue
		}
		if _, duplicate := packageFiles[entry.Path]; duplicate {
			return nil, nil, errors.New("candidate media package has duplicate file identities")
		}
		packageFiles[entry.Path] = entry
		digest := sha256.Sum256([]byte(entry.Path))
		hashPaths[hex.EncodeToString(digest[:])] = entry.Path
	}
	seenActions := map[string]bool{}
	for _, action := range selection.RequiredFiles {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		path, ok := hashPaths[action.PathSHA256]
		if !ok || seenActions[action.PathSHA256] {
			return nil, nil, errors.New("candidate required media identity is missing or duplicated")
		}
		seenActions[action.PathSHA256] = true
		parts := strings.SplitN(path, "/", 3)
		if len(parts) != 3 {
			return nil, nil, errors.New("candidate required media path is invalid")
		}
		if _, err := customerServiceCandidateMediaPathComponents(parts[1], parts[2]); err != nil {
			return nil, nil, err
		}
		switch action.Action {
		case "preserve_current_required_file":
			identity, exists := paths[path]
			if !exists || identity.SHA256 != action.TargetSHA256 || identity.Size != action.TargetSize {
				return nil, nil, errors.New("candidate required current file changed")
			}
		case "selected_add_package_file":
			if _, exists := paths[path]; exists {
				return nil, nil, errors.New("candidate package addition would overwrite a current file")
			}
			expected, exists := packageFiles[path]
			if !exists || expected.Size < 0 || expected.SHA256 != action.SourceSHA256 || expected.Size != action.SourceSize {
				return nil, nil, errors.New("candidate package addition is not bound to its verified identity")
			}
			actual, err := inspectBoundedModuleRecoveryRegularFileIdentity(ctx, verifiedDirectory, path, expected.Size)
			if err != nil || actual.SHA256 != expected.SHA256 || actual.Size != expected.Size {
				return nil, nil, errors.New("selected candidate package file changed")
			}
			if err := registerPath(path, false); err != nil {
				return nil, nil, err
			}
			for parent := filepath.ToSlash(filepath.Dir(parts[2])); parent != "."; parent = filepath.ToSlash(filepath.Dir(parent)) {
				parentPath := "media/" + parts[1] + "/" + parent
				if !directories[parentPath] {
					if err := registerPath(parentPath, true); err != nil {
						return nil, nil, err
					}
					inventory.Directories = append(inventory.Directories, customerServiceCandidateMediaDirectory{parts[1], parent})
					if err := accountCustomerServiceCandidateMediaInventoryMetadata(inventory, parentPath); err != nil {
						return nil, nil, err
					}
				}
			}
			if expected.Size > cfg.MaxExpandedBytes-inventory.TotalBytes {
				return nil, nil, errors.New("candidate complete media byte limit exceeded")
			}
			inventory.TotalBytes += expected.Size
			if err := accountCustomerServiceCandidateMediaInventoryMetadata(inventory, path); err != nil {
				return nil, nil, err
			}
			inventory.AddedPackageFiles++
			paths[path] = expected
			inventory.Files = append(inventory.Files, customerServiceCandidateMediaFile{parts[1], parts[2], "selected_package", expected})
		default:
			return nil, nil, errors.New("candidate required media action is blocked or unsupported")
		}
	}
	if len(inventory.Files)+len(inventory.Directories) > min(cfg.MaxFiles, customerServiceCandidateMediaMaximumEntries) {
		return nil, nil, errors.New("candidate complete media entry limit exceeded")
	}
	sortCustomerServiceCandidateMediaInventory(inventory)
	inventory.CandidateMediaFingerprint = fingerprintCustomerServiceCandidateMediaInventory(inventory)
	inventory.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(inventory)
	if err != nil {
		return nil, nil, err
	}
	encoded, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	if len(encoded) > customerServiceCandidateMediaMaximumInventoryBytes {
		return nil, nil, errors.New("candidate media inventory exceeds its metadata limit")
	}
	inventory.EncodedBytes = len(encoded)
	return inventory, encoded, nil
}

func customerServiceCandidateMediaVolumeDirectory(record *customerServicePersistentCandidateLifecycle, volume customerServiceCandidateMediaVolumeState) (string, error) {
	if !safeTaskID.MatchString(record.TaskID) || !safeTaskID.MatchString(record.AttemptID) || !filepath.IsAbs(volume.StorageRoot) {
		return "", errors.New("candidate media volume identity is invalid")
	}
	path := filepath.Join(volume.StorageRoot, ".local-backup-customer-service-candidate-"+record.TaskID+"-"+record.AttemptID)
	if filepath.Dir(path) != filepath.Clean(volume.StorageRoot) {
		return "", errors.New("candidate media volume escaped its storage root")
	}
	return path, nil
}

func customerServiceCandidateExpectedMediaVolumeIdentity(record *customerServicePersistentCandidateLifecycle, volume customerServiceCandidateMediaVolumeState) customerServiceCandidateMediaVolumeIdentity {
	return customerServiceCandidateMediaVolumeIdentity{1, record.TaskID, record.AttemptID, record.Request.PlanSHA256, volume.RootName, volume.StorageRoot}
}

// This copier opens both sides through filesystem roots, rejects every symlink
// component, limits reads to the expected length and checks the final digest.
func copyCustomerServiceCandidateBoundedVerifiedMediaFile(ctx context.Context, sourceDirectory, sourceKey, destinationDirectory, destinationKey string, expected FileIdentity, mode os.FileMode) error {
	source, err := os.OpenRoot(sourceDirectory)
	if err != nil {
		return err
	}
	defer source.Close()
	var componentPath string
	for _, component := range strings.Split(sourceKey, "/") {
		componentPath = filepath.Join(componentPath, component)
		info, err := source.Lstat(componentPath)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("candidate media copy refuses source symlinks")
		}
	}
	input, err := source.Open(filepath.FromSlash(sourceKey))
	if err != nil {
		return err
	}
	defer input.Close()
	before, err := input.Stat()
	if err != nil {
		return err
	}
	if !before.Mode().IsRegular() || before.Size() != expected.Size {
		return errors.New("candidate media copy source size or type changed")
	}
	destination, err := os.OpenRoot(destinationDirectory)
	if err != nil {
		return err
	}
	defer destination.Close()
	if err := destination.MkdirAll(filepath.Dir(filepath.FromSlash(destinationKey)), mode|0100); err != nil {
		return err
	}
	output, err := destination.OpenFile(filepath.FromSlash(destinationKey), os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	hash := sha256.New()
	buffer := make([]byte, 64<<10)
	var size int64
	copyErr := func() error {
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			remaining := expected.Size - size + 1
			if remaining <= 0 {
				return errors.New("candidate media copy exceeded expected size")
			}
			count, readErr := input.Read(buffer[:min(int64(len(buffer)), remaining)])
			size += int64(count)
			if size > expected.Size {
				return errors.New("candidate media copy exceeded expected size")
			}
			if _, err := io.MultiWriter(output, hash).Write(buffer[:count]); err != nil {
				return err
			}
			if readErr == io.EOF {
				return nil
			}
			if readErr != nil {
				return readErr
			}
		}
	}()
	syncErr, closeErr := output.Sync(), output.Close()
	if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
		return err
	}
	after, err := source.Lstat(filepath.FromSlash(sourceKey))
	if err != nil {
		return err
	}
	if !after.Mode().IsRegular() || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) || size != expected.Size || after.Size() != expected.Size || hex.EncodeToString(hash.Sum(nil)) != expected.SHA256 {
		return errors.New("candidate media copy content or source identity changed")
	}
	return nil
}

func (manager *Manager) constructCustomerServiceCompleteMediaUnionOnOwnedVolumes(ctx context.Context, record *customerServicePersistentCandidateLifecycle, inventory *customerServiceCandidateCompleteMediaUnionInventory, verifiedDirectory string) error {
	for index := range record.MediaUnion.Volumes {
		volume := &record.MediaUnion.Volumes[index]
		if volume.StorageRoot != manager.Configuration.StorageRoots[volume.RootName] {
			return errors.New("candidate media storage mapping changed")
		}
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, *volume)
		if err != nil {
			return err
		}
		volume.State = "allocation_intent"
		if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
			return err
		}
		if err := os.Mkdir(directory, 0700); err != nil {
			return err
		}
		if err := syncDirectory(volume.StorageRoot); err != nil {
			return err
		}
		if err := writeDurableJSON(filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename), customerServiceCandidateExpectedMediaVolumeIdentity(record, *volume)); err != nil {
			return err
		}
		volume.State = "identity_recorded"
		if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
			return err
		}
		tree := filepath.Join(directory, "tree")
		if err := os.Mkdir(tree, 0700); err != nil {
			return err
		}
		directoryMode, fileMode := os.FileMode(0755), os.FileMode(0644)
		if volume.RootName == "uploads_private" {
			directoryMode, fileMode = 0700, 0600
		}
		root, err := os.OpenRoot(tree)
		if err != nil {
			return err
		}
		for _, entry := range inventory.Directories {
			if entry.RootName == volume.RootName {
				if err := root.MkdirAll(filepath.FromSlash(entry.Key), directoryMode); err != nil {
					root.Close()
					return err
				}
			}
		}
		root.Close()
		for _, entry := range inventory.Files {
			if entry.RootName != volume.RootName {
				continue
			}
			sourceDirectory, sourceKey := volume.StorageRoot, entry.Key
			if entry.Origin == "selected_package" {
				sourceDirectory, sourceKey = verifiedDirectory, entry.Identity.Path
			}
			if err := copyCustomerServiceCandidateBoundedVerifiedMediaFile(ctx, sourceDirectory, sourceKey, tree, entry.Key, entry.Identity, fileMode); err != nil {
				return err
			}
		}
		if err := filepath.WalkDir(tree, func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if entry.IsDir() {
				if err := os.Chmod(filename, directoryMode); err != nil {
					return err
				}
				return syncDirectory(filename)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := syncDirectory(directory); err != nil {
			return err
		}
		volume.State = "tree_prepared"
		if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
			return err
		}
	}
	if err := manager.verifyCustomerServiceCompleteCandidateMediaUnion(ctx, record, inventory); err != nil {
		return err
	}
	record.MediaUnion.State = "prepared_not_switchable"
	return manager.saveCustomerServicePersistentCandidateLifecycle(record)
}

func (manager *Manager) verifyCustomerServiceCompleteCandidateMediaUnion(ctx context.Context, record *customerServicePersistentCandidateLifecycle, expected *customerServiceCandidateCompleteMediaUnionInventory) error {
	current, err := manager.Configuration.inventoryCustomerServiceCompleteCurrentMedia(ctx)
	if err != nil {
		return err
	}
	baselineDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(current)
	if err != nil {
		return err
	}
	if baselineDigest != expected.BaselineTreeSHA256 {
		return errors.New("complete current media changed during candidate construction")
	}
	candidateCfg := manager.Configuration
	candidateCfg.StorageRoots = map[string]string{}
	for _, volume := range record.MediaUnion.Volumes {
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
		if err != nil {
			return err
		}
		candidateCfg.StorageRoots[volume.RootName] = filepath.Join(directory, "tree")
	}
	actual, err := candidateCfg.inventoryCustomerServiceMediaTreesWithMaintenanceArtifactPolicy(ctx, "reject_all_maintenance_artifacts")
	if err != nil {
		return err
	}
	if len(actual.Files) != len(expected.Files) || !reflect.DeepEqual(actual.Directories, expected.Directories) {
		return errors.New("candidate complete media file or directory set changed")
	}
	for index, file := range actual.Files {
		if file.Identity != expected.Files[index].Identity {
			return errors.New("candidate complete media content differs from its expected union")
		}
	}
	if fingerprintCustomerServiceCandidateMediaInventory(actual) != expected.CandidateMediaFingerprint {
		return errors.New("candidate complete media fingerprint differs")
	}
	return nil
}

func (manager *Manager) cleanupCustomerServiceCandidateMediaVolumesByRecordedIdentity(record *customerServicePersistentCandidateLifecycle) error {
	if record.MediaUnion == nil {
		return nil
	}
	// Validate every existing volume before deleting any. Missing markers after a
	// mkdir interruption retain the scene; names alone never authorize deletion.
	for _, volume := range record.MediaUnion.Volumes {
		if volume.StorageRoot != manager.Configuration.StorageRoots[volume.RootName] {
			return errors.New("candidate media cleanup storage mapping changed")
		}
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
		if err != nil {
			return err
		}
		rootInfo, err := os.Lstat(volume.StorageRoot)
		if err != nil {
			return err
		}
		if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
			return errors.New("candidate media cleanup root is unsafe")
		}
		info, err := os.Lstat(directory)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("candidate media cleanup directory is unsafe")
		}
		filename := filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename)
		info, err = os.Lstat(filename)
		if err != nil {
			return errors.New("candidate media cleanup identity is missing; inspection required")
		}
		if !info.Mode().IsRegular() || info.Size() > 4096 {
			return errors.New("candidate media cleanup identity is unsafe")
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		var actual customerServiceCandidateMediaVolumeIdentity
		if decodeStrictJSON(data, &actual) != nil || actual != customerServiceCandidateExpectedMediaVolumeIdentity(record, volume) {
			return errors.New("candidate media cleanup identity changed; inspection required")
		}
	}
	record.MediaUnion.State = "discard_required"
	if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
		return err
	}
	for index := range record.MediaUnion.Volumes {
		volume := &record.MediaUnion.Volumes[index]
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, *volume)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(directory); err != nil {
			return fmt.Errorf("candidate media cleanup failed; lifecycle retained")
		}
		if err := syncDirectory(volume.StorageRoot); err != nil {
			return err
		}
		volume.State = "removed"
		if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
			return err
		}
	}
	record.MediaUnion.State = "discarded_not_switchable"
	return manager.saveCustomerServicePersistentCandidateLifecycle(record)
}

func (manager *Manager) validateCustomerServiceCandidateMediaLifecycle(record *customerServicePersistentCandidateLifecycle) error {
	media := record.MediaUnion
	if media == nil {
		return nil
	} // Earlier v1 records describe database-only candidates.
	if media.Version != 1 || len(media.Volumes) != len(mediaRoots) || len(media.InventorySHA256) != 64 || media.TotalBytes < 0 || media.CurrentFiles < 0 || media.AddedPackageFiles < 0 {
		return errors.New("candidate complete media lifecycle is invalid")
	}
	if media.State != "allocation_intent" && media.State != "prepared_not_switchable" && media.State != "discarded_not_switchable" && media.State != "discard_required" {
		return errors.New("candidate complete media lifecycle state is unsupported")
	}
	seen := map[string]bool{}
	for _, volume := range media.Volumes {
		if seen[volume.RootName] || volume.StorageRoot == "" || volume.StorageRoot != manager.Configuration.StorageRoots[volume.RootName] {
			return errors.New("candidate complete media volume mapping changed")
		}
		seen[volume.RootName] = true
		if _, err := customerServiceCandidateMediaVolumeDirectory(record, volume); err != nil {
			return err
		}
		switch volume.State {
		case "unallocated", "allocation_intent", "identity_recorded", "tree_prepared", "removed":
		default:
			return errors.New("candidate complete media volume state is unsupported")
		}
		if media.State == "prepared_not_switchable" && volume.State != "tree_prepared" {
			return errors.New("candidate complete media preparation evidence is incomplete")
		}
		if media.State == "discarded_not_switchable" && volume.State != "removed" {
			return errors.New("candidate complete media cleanup evidence is incomplete")
		}
	}
	if record.State == "sealed_not_switchable" && (media.State != "prepared_not_switchable" || record.PeakCapacity == nil || record.Binding == nil || media.BaselineMediaFingerprint != record.Binding.TargetMediaFingerprint) {
		return errors.New("sealed candidate complete media evidence is invalid")
	}
	if record.State == "discarded_not_switchable" && media.State != "discarded_not_switchable" {
		return errors.New("discarded candidate complete media cleanup evidence is invalid")
	}
	return nil
}
