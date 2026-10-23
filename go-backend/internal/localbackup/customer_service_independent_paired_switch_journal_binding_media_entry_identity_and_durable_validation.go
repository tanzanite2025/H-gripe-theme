package localbackup

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

const customerServiceIndependentPairedSwitchJournalFilename = "customer-service-independent-paired-switch-journal-v1.json"
const customerServiceIndependentPairedSwitchMaximumJournalBytes = 64 << 20

type customerServiceIndependentPairedSwitchMediaEntryMove struct {
	RootName        string `json:"root_name"`
	EntryName       string `json:"entry_name"`
	Direction       string `json:"direction"`
	EntrySHA256     string `json:"entry_sha256"`
	Started         bool   `json:"started"`
	Completed       bool   `json:"completed"`
	RollbackStarted bool   `json:"rollback_started"`
	RolledBack      bool   `json:"rolled_back"`
}

// This journal permits only an internal isolated rehearsal and its rollback.
// It never changes Task.StageDatabase or grants production execution authority.
type customerServiceIndependentPairedSwitchJournal struct {
	Version                      int                                                    `json:"version"`
	State                        string                                                 `json:"state"`
	DatabasePhase                string                                                 `json:"database_phase"`
	ExecutionAvailable           bool                                                   `json:"execution_available"`
	TaskID                       string                                                 `json:"task_id"`
	AttemptID                    string                                                 `json:"attempt_id"`
	CreatedAt                    time.Time                                              `json:"created_at"`
	UpdatedAt                    time.Time                                              `json:"updated_at"`
	ConfigurationSHA256          string                                                 `json:"configuration_sha256"`
	OriginalTaskSHA256           string                                                 `json:"original_task_sha256"`
	CandidateLifecycleSHA256     string                                                 `json:"candidate_lifecycle_sha256"`
	InventorySHA256              string                                                 `json:"inventory_sha256"`
	PlanSHA256                   string                                                 `json:"plan_sha256"`
	SelectionSHA256              string                                                 `json:"selection_sha256"`
	TargetDatabase               string                                                 `json:"target_database"`
	CandidateDatabase            string                                                 `json:"candidate_database"`
	RollbackDatabase             string                                                 `json:"rollback_database"`
	OriginalDatabaseOID          int64                                                  `json:"original_database_oid"`
	CandidateDatabaseOID         int64                                                  `json:"candidate_database_oid"`
	OriginalDatabaseOwner        string                                                 `json:"original_database_owner"`
	BaselineDatabaseFingerprint  string                                                 `json:"baseline_database_fingerprint"`
	CandidateDatabaseFingerprint string                                                 `json:"candidate_database_fingerprint"`
	BaselineDirectories          []customerServiceCandidateMediaDirectory               `json:"baseline_directories"`
	Moves                        []customerServiceIndependentPairedSwitchMediaEntryMove `json:"moves"`
	ErrorCode                    string                                                 `json:"error_code,omitempty"`
	SHA256                       string                                                 `json:"sha256"`
}

func isCustomerServiceIndependentPairedSwitchSHA256(digest string) bool {
	decoded, err := hex.DecodeString(digest)
	return err == nil && len(decoded) == 32
}

func (manager *Manager) saveCustomerServiceIndependentPairedSwitchJournal(journal *customerServiceIndependentPairedSwitchJournal) error {
	directory, err := manager.taskDirectory(journal.TaskID)
	if err != nil {
		return err
	}
	journal.UpdatedAt = time.Now().UTC()
	journal.SHA256 = ""
	journal.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(journal)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return err
	}
	if len(encoded) > customerServiceIndependentPairedSwitchMaximumJournalBytes {
		return errors.New("independent paired journal exceeds its size limit")
	}
	if err := manager.Configuration.checkWorkspaceCapacity(int64(len(encoded)) * 2); err != nil {
		return err
	}
	return writeDurableFile(filepath.Join(directory, customerServiceIndependentPairedSwitchJournalFilename), encoded)
}

func readCustomerServicePairedRecoveryBoundedRegularFile(filename string, maximum int64) ([]byte, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maximum {
		return nil, errors.New("paired recovery requires a bounded regular record")
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(data)) > maximum {
		return nil, errors.New("paired recovery record exceeds its size limit")
	}
	return data, nil
}

func (manager *Manager) loadCustomerServiceIndependentPairedSwitchJournal(id string) (*customerServiceIndependentPairedSwitchJournal, error) {
	directory, err := manager.taskDirectory(id)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("independent paired journal task directory is unsafe")
	}
	data, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(directory, customerServiceIndependentPairedSwitchJournalFilename), customerServiceIndependentPairedSwitchMaximumJournalBytes)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var journal customerServiceIndependentPairedSwitchJournal
	if decodeStrictJSON(data, &journal) != nil {
		return nil, errors.New("independent paired journal is invalid; inspection required")
	}
	copyOfJournal := journal
	copyOfJournal.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyOfJournal)
	if err != nil || digest != journal.SHA256 || journal.Version != 1 || journal.ExecutionAvailable || journal.TaskID != id || !safeTaskID.MatchString(journal.AttemptID) {
		return nil, errors.New("independent paired journal digest or identity changed")
	}
	if journal.TargetDatabase != manager.Configuration.DatabaseName || journal.CandidateDatabase != "local_backup_candidate_"+journal.AttemptID[:16] || journal.RollbackDatabase != "local_backup_cs_old_"+journal.AttemptID[:16] || journal.TargetDatabase == journal.CandidateDatabase || journal.TargetDatabase == journal.RollbackDatabase || journal.OriginalDatabaseOID <= 0 || journal.CandidateDatabaseOID <= 0 || journal.OriginalDatabaseOID == journal.CandidateDatabaseOID || journal.OriginalDatabaseOwner == "" {
		return nil, errors.New("independent paired journal database identity is invalid")
	}
	for _, digest := range []string{journal.ConfigurationSHA256, journal.OriginalTaskSHA256, journal.CandidateLifecycleSHA256, journal.InventorySHA256, journal.PlanSHA256, journal.SelectionSHA256, journal.BaselineDatabaseFingerprint, journal.CandidateDatabaseFingerprint} {
		if !isCustomerServiceIndependentPairedSwitchSHA256(digest) {
			return nil, errors.New("independent paired journal binding is invalid")
		}
	}
	switch journal.State {
	case "prepared_isolated_rehearsal", "switching_isolated_rehearsal", "switched_isolated_rollback_required", "rolling_back_isolated_rehearsal", "rolled_back_isolated", "inspection_required":
	default:
		return nil, errors.New("independent paired journal state is unsupported")
	}
	switch journal.DatabasePhase {
	case "not_started", "retain_original_database_intent", "original_database_retained", "install_candidate_database_intent", "candidate_database_installed", "return_candidate_database_intent", "rolling_back_media", "restore_original_database_intent", "original_database_and_media_verified":
	default:
		return nil, errors.New("independent paired database phase is unsupported")
	}
	if (journal.State == "prepared_isolated_rehearsal" && journal.DatabasePhase != "not_started") ||
		(journal.State == "switched_isolated_rollback_required" && journal.DatabasePhase != "candidate_database_installed") ||
		(journal.State == "rolled_back_isolated" && journal.DatabasePhase != "original_database_and_media_verified") {
		return nil, errors.New("independent paired journal database phase contradicts its state")
	}
	if len(journal.Moves) > customerServiceCandidateMediaMaximumEntries*2 || len(journal.BaselineDirectories) > customerServiceCandidateMediaMaximumEntries {
		return nil, errors.New("independent paired journal entry limit exceeded")
	}
	seen := map[string]bool{}
	for _, move := range journal.Moves {
		if strings.Contains(move.EntryName, "/") || strings.Contains(move.EntryName, "\\") {
			return nil, errors.New("independent paired journal entry is not a top-level name")
		}
		if _, err := customerServiceCandidateMediaPathComponents(move.RootName, move.EntryName); err != nil {
			return nil, err
		}
		key := move.Direction + "/" + move.RootName + "/" + move.EntryName
		if seen[key] || !isCustomerServiceIndependentPairedSwitchSHA256(move.EntrySHA256) || (move.Direction != "retain_current" && move.Direction != "install_candidate") || (move.Completed && !move.Started) || (move.RollbackStarted && !move.Started) || (move.RolledBack && !move.RollbackStarted) {
			return nil, errors.New("independent paired journal move is invalid")
		}
		seen[key] = true
		if journal.State == "prepared_isolated_rehearsal" && (move.Started || move.RollbackStarted) {
			return nil, errors.New("unstarted independent journal contains move progress")
		}
		if journal.State == "switched_isolated_rollback_required" && (!move.Completed || move.RollbackStarted) {
			return nil, errors.New("switched independent journal is incomplete")
		}
		if journal.State == "rolled_back_isolated" && move.Started && !move.RolledBack {
			return nil, errors.New("independent paired rollback evidence is incomplete")
		}
	}
	return &journal, nil
}

func (manager *Manager) requireCustomerServiceIndependentPairedSwitchRolledBack(id string) error {
	journal, err := manager.loadCustomerServiceIndependentPairedSwitchJournal(id)
	if err != nil {
		return err
	}
	if journal != nil && journal.State != "rolled_back_isolated" {
		return errors.New("independent customer-service paired switch requires verified rollback before candidate disposal, maintenance release or whole-storefront operations")
	}
	return nil
}

func (manager *Manager) loadCustomerServicePairedSwitchBoundCompleteMediaInventory(record *customerServicePersistentCandidateLifecycle) (*customerServiceCandidateCompleteMediaUnionInventory, error) {
	workspace, err := manager.customerServicePersistentCandidateConstructionWorkspace(record)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(workspace.Directory)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("paired switch inventory directory is unsafe")
	}
	data, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(workspace.Directory, customerServiceCandidateMediaInventoryFilename), customerServiceCandidateMediaMaximumInventoryBytes)
	if err != nil {
		return nil, err
	}
	var inventory customerServiceCandidateCompleteMediaUnionInventory
	if decodeStrictJSON(data, &inventory) != nil {
		return nil, errors.New("paired switch media inventory is invalid")
	}
	copyOfInventory := inventory
	copyOfInventory.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyOfInventory)
	if err != nil || digest != inventory.SHA256 || record.MediaUnion == nil || record.Validation == nil || digest != record.MediaUnion.InventorySHA256 || inventory.Version != 1 || inventory.PlanSHA256 != record.Request.PlanSHA256 || inventory.SelectionSHA256 != record.Validation.SelectionSHA256 || inventory.CandidateMediaFingerprint != record.MediaUnion.CandidateMediaFingerprint || inventory.BaselineMediaFingerprint != record.MediaUnion.BaselineMediaFingerprint {
		return nil, errors.New("paired switch media inventory binding changed")
	}
	if len(inventory.Files)+len(inventory.Directories) > min(manager.Configuration.MaxFiles, customerServiceCandidateMediaMaximumEntries) {
		return nil, errors.New("paired switch media inventory entry limit exceeded")
	}
	return &inventory, nil
}

// The digest describes the complete subtree including empty directories. Paths
// remain relative to the media root so identities survive same-volume renames.
func customerServicePairedSwitchMediaEntryDigest(inventory *customerServiceCandidateCompleteMediaUnionInventory, rootName, entryName string) (string, error) {
	files := []FileIdentity{}
	directories := []string{}
	for _, file := range inventory.Files {
		if file.RootName == rootName && (file.Key == entryName || strings.HasPrefix(file.Key, entryName+"/")) {
			files = append(files, file.Identity)
		}
	}
	for _, directory := range inventory.Directories {
		if directory.RootName == rootName && (directory.Key == entryName || strings.HasPrefix(directory.Key, entryName+"/")) {
			directories = append(directories, directory.Key)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	sort.Strings(directories)
	return customerServiceRecoveryPlanCanonicalJSONSHA256(struct {
		Files       []FileIdentity
		Directories []string
	}{files, directories})
}

func (cfg Configuration) inspectCustomerServicePairedSwitchActualMediaEntryDigest(ctx context.Context, directory, rootName, entryName string) (string, error) {
	rootInfo, err := os.Lstat(directory)
	if err != nil {
		return "", err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("paired media entry parent is unsafe")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return "", err
	}
	defer root.Close()
	inventory := &customerServiceCandidateCompleteMediaUnionInventory{}
	err = fs.WalkDir(root.FS(), entryName, func(key string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		path, err := customerServiceCandidateMediaPathComponents(rootName, key)
		if err != nil {
			return err
		}
		if err := accountCustomerServiceCandidateMediaInventoryMetadata(inventory, path); err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 || len(inventory.Files)+len(inventory.Directories) >= min(cfg.MaxFiles, customerServiceCandidateMediaMaximumEntries) {
			return errors.New("paired media entry is unsafe or exceeds its entry limit")
		}
		if entry.IsDir() {
			inventory.Directories = append(inventory.Directories, customerServiceCandidateMediaDirectory{rootName, key})
			return nil
		}
		identity, err := inspectBoundedModuleRecoveryRegularFileIdentity(ctx, directory, key, cfg.MaxExpandedBytes-inventory.TotalBytes)
		if err != nil {
			return err
		}
		identity.Path = path
		inventory.TotalBytes += identity.Size
		inventory.Files = append(inventory.Files, customerServiceCandidateMediaFile{rootName, key, "", identity})
		return nil
	})
	if err != nil {
		return "", err
	}
	return customerServicePairedSwitchMediaEntryDigest(inventory, rootName, entryName)
}

func buildCustomerServicePairedSwitchExpectedMediaMoves(baseline, candidate *customerServiceCandidateCompleteMediaUnionInventory) ([]customerServiceIndependentPairedSwitchMediaEntryMove, error) {
	moves := []customerServiceIndependentPairedSwitchMediaEntryMove{}
	for _, source := range []struct {
		inventory *customerServiceCandidateCompleteMediaUnionInventory
		direction string
	}{{baseline, "retain_current"}, {candidate, "install_candidate"}} {
		groups := map[string]*customerServiceCandidateCompleteMediaUnionInventory{}
		groupForEntry := func(rootName, key string) *customerServiceCandidateCompleteMediaUnionInventory {
			name := rootName + "/" + strings.Split(key, "/")[0]
			if groups[name] == nil {
				groups[name] = &customerServiceCandidateCompleteMediaUnionInventory{}
			}
			return groups[name]
		}
		for _, file := range source.inventory.Files {
			group := groupForEntry(file.RootName, file.Key)
			group.Files = append(group.Files, file)
		}
		for _, directory := range source.inventory.Directories {
			group := groupForEntry(directory.RootName, directory.Key)
			group.Directories = append(group.Directories, directory)
		}
		for _, rootName := range mediaRoots {
			entries := map[string]bool{}
			for _, file := range source.inventory.Files {
				if file.RootName == rootName {
					entries[strings.Split(file.Key, "/")[0]] = true
				}
			}
			for _, directory := range source.inventory.Directories {
				if directory.RootName == rootName {
					entries[strings.Split(directory.Key, "/")[0]] = true
				}
			}
			names := []string{}
			for name := range entries {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				if _, err := customerServiceCandidateMediaPathComponents(rootName, name); err != nil {
					return nil, err
				}
				digest, err := customerServicePairedSwitchMediaEntryDigest(groups[rootName+"/"+name], rootName, name)
				if err != nil {
					return nil, err
				}
				moves = append(moves, customerServiceIndependentPairedSwitchMediaEntryMove{RootName: rootName, EntryName: name, Direction: source.direction, EntrySHA256: digest})
			}
		}
	}
	return moves, nil
}

func reconstructCustomerServicePairedSwitchBaselineAndValidateMoves(journal *customerServiceIndependentPairedSwitchJournal, inventory *customerServiceCandidateCompleteMediaUnionInventory) (*customerServiceCandidateCompleteMediaUnionInventory, error) {
	baseline := &customerServiceCandidateCompleteMediaUnionInventory{Version: 1, Files: []customerServiceCandidateMediaFile{}, Directories: journal.BaselineDirectories, CurrentBytes: inventory.CurrentBytes, TotalBytes: inventory.CurrentBytes}
	for _, file := range inventory.Files {
		if file.Origin == "current_target" {
			baseline.Files = append(baseline.Files, file)
		} else if file.Origin != "selected_package" {
			return nil, errors.New("paired inventory file origin is unsupported")
		}
	}
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(baseline)
	if err != nil || digest != inventory.BaselineTreeSHA256 {
		return nil, errors.New("paired switch baseline directory or file identity changed")
	}
	expectedMoves, err := buildCustomerServicePairedSwitchExpectedMediaMoves(baseline, inventory)
	if err != nil || len(expectedMoves) != len(journal.Moves) {
		return nil, errors.New("paired switch media move set changed")
	}
	for index, move := range journal.Moves {
		move.Started, move.Completed, move.RollbackStarted, move.RolledBack = false, false, false, false
		if !reflect.DeepEqual(move, expectedMoves[index]) {
			return nil, errors.New("paired switch media move identity changed")
		}
	}
	return baseline, nil
}
