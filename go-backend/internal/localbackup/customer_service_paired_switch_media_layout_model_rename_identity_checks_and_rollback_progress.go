package localbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func customerServicePairedSwitchMediaMoveLogicalPaths(move customerServiceIndependentPairedSwitchMediaEntryMove) (string, string) {
	key := move.RootName + "/" + move.EntryName
	if move.Direction == "retain_current" {
		return "live/" + key, "previous/" + key
	}
	return "candidate/" + key, "live/" + key
}

func (manager *Manager) customerServicePairedSwitchLogicalMediaPath(record *customerServicePersistentCandidateLifecycle, logical string) (string, error) {
	parts := strings.SplitN(logical, "/", 3)
	if len(parts) != 3 || strings.ContainsAny(parts[2], "/\\") {
		return "", errors.New("paired media logical path is invalid")
	}
	if _, err := customerServiceCandidateMediaPathComponents(parts[1], parts[2]); err != nil {
		return "", err
	}
	for _, volume := range record.MediaUnion.Volumes {
		if volume.RootName != parts[1] {
			continue
		}
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
		if err != nil {
			return "", err
		}
		switch parts[0] {
		case "live":
			return filepath.Join(volume.StorageRoot, parts[2]), nil
		case "candidate":
			return filepath.Join(directory, "tree", parts[2]), nil
		case "previous":
			return filepath.Join(directory, "previous", parts[2]), nil
		}
	}
	return "", errors.New("paired media logical volume is invalid")
}

func (manager *Manager) validateCustomerServicePairedSwitchOwnedMediaVolumes(record *customerServicePersistentCandidateLifecycle) error {
	for _, volume := range record.MediaUnion.Volumes {
		if volume.StorageRoot != manager.Configuration.StorageRoots[volume.RootName] {
			return errors.New("paired media storage mapping changed")
		}
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
		if err != nil {
			return err
		}
		for _, parent := range []string{volume.StorageRoot, directory} {
			info, err := os.Lstat(parent)
			if err != nil {
				return err
			}
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return errors.New("paired media volume directory is unsafe")
			}
		}
		data, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename), 4096)
		if err != nil {
			return err
		}
		var identity customerServiceCandidateMediaVolumeIdentity
		if decodeStrictJSON(data, &identity) != nil || identity != customerServiceCandidateExpectedMediaVolumeIdentity(record, volume) {
			return errors.New("paired media volume identity changed")
		}
		entries, err := os.ReadDir(directory)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Name() != customerServiceCandidateMediaVolumeIdentityFilename && entry.Name() != "tree" && entry.Name() != "previous" {
				return errors.New("paired media volume contains an unregistered entry")
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return errors.New("paired media volume contains a symbolic link")
			}
		}
	}
	return nil
}

func (manager *Manager) inspectCustomerServicePairedSwitchCompletePhysicalMediaLayout(ctx context.Context, record *customerServicePersistentCandidateLifecycle) (map[string]string, error) {
	if err := manager.validateCustomerServicePairedSwitchOwnedMediaVolumes(record); err != nil {
		return nil, err
	}
	layout := map[string]string{}
	for _, volume := range record.MediaUnion.Volumes {
		directory, _ := customerServiceCandidateMediaVolumeDirectory(record, volume)
		for _, location := range []struct{ label, path string }{{"live", volume.StorageRoot}, {"candidate", filepath.Join(directory, "tree")}, {"previous", filepath.Join(directory, "previous")}} {
			info, err := os.Lstat(location.path)
			if errors.Is(err, os.ErrNotExist) && location.label == "previous" {
				continue
			}
			if err != nil {
				return nil, err
			}
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil, errors.New("paired media tree parent is unsafe")
			}
			entries, err := os.ReadDir(location.path)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if location.label == "live" && strings.HasPrefix(entry.Name(), ".local-backup-") {
					continue
				}
				digest, err := manager.Configuration.inspectCustomerServicePairedSwitchActualMediaEntryDigest(ctx, location.path, volume.RootName, entry.Name())
				if err != nil {
					return nil, err
				}
				layout[location.label+"/"+volume.RootName+"/"+entry.Name()] = digest
			}
		}
	}
	return layout, nil
}

func simulateCustomerServicePairedSwitchMediaRename(layout map[string]string, source, destination, digest string, allowUnperformedForwardMove bool) bool {
	sourceDigest, sourceExists := layout[source]
	destinationDigest, destinationExists := layout[destination]
	if allowUnperformedForwardMove && !sourceExists && destinationExists && destinationDigest == digest {
		return true
	}
	if !sourceExists || sourceDigest != digest || destinationExists {
		return false
	}
	delete(layout, source)
	layout[destination] = digest
	return true
}

// A response may be lost after the last forward or reverse rename. Compare the
// entire physical layout with each permissible durable boundary, rather than
// guessing from one path that a later move may now occupy with identical bytes.
func validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(journal *customerServiceIndependentPairedSwitchJournal, actual map[string]string) error {
	started := 0
	for index, move := range journal.Moves {
		if move.Started {
			if index != started || (index > 0 && !journal.Moves[index-1].Completed) {
				return errors.New("paired forward progress is not a contiguous durable prefix")
			}
			started++
		}
	}
	rollbackBoundarySeen := false
	for index := started - 1; index >= 0; index-- {
		move := journal.Moves[index]
		if move.RollbackStarted {
			if rollbackBoundarySeen || (index < started-1 && !journal.Moves[index+1].RolledBack) {
				return errors.New("paired reverse progress is not a contiguous durable suffix")
			}
			if !move.RolledBack {
				rollbackBoundarySeen = true
			}
		} else {
			rollbackBoundarySeen = true
		}
	}
	forwardCounts := []int{started}
	if started > 0 && !journal.Moves[started-1].Completed {
		forwardCounts = append(forwardCounts, started-1)
	}
	for _, forwardCount := range forwardCounts {
		for _, performUnacknowledgedReverse := range []bool{false, true} {
			layout := map[string]string{}
			valid := true
			for _, move := range journal.Moves {
				source, _ := customerServicePairedSwitchMediaMoveLogicalPaths(move)
				if _, duplicate := layout[source]; duplicate {
					valid = false
					break
				}
				layout[source] = move.EntrySHA256
			}
			for index := 0; valid && index < forwardCount; index++ {
				move := journal.Moves[index]
				source, destination := customerServicePairedSwitchMediaMoveLogicalPaths(move)
				valid = simulateCustomerServicePairedSwitchMediaRename(layout, source, destination, move.EntrySHA256, false)
			}
			for index := started - 1; valid && index >= 0; index-- {
				move := journal.Moves[index]
				if !move.RollbackStarted || (!move.RolledBack && !performUnacknowledgedReverse) {
					continue
				}
				source, destination := customerServicePairedSwitchMediaMoveLogicalPaths(move)
				valid = simulateCustomerServicePairedSwitchMediaRename(layout, destination, source, move.EntrySHA256, true)
			}
			if valid && reflect.DeepEqual(layout, actual) {
				return nil
			}
		}
	}
	return errors.New("paired media layout or content differs from every recorded rename boundary; inspection required")
}

func (manager *Manager) performCustomerServicePairedSwitchBoundMediaRename(ctx context.Context, record *customerServicePersistentCandidateLifecycle, move customerServiceIndependentPairedSwitchMediaEntryMove, reverse bool) error {
	sourceLogical, destinationLogical := customerServicePairedSwitchMediaMoveLogicalPaths(move)
	if reverse {
		sourceLogical, destinationLogical = destinationLogical, sourceLogical
	}
	source, err := manager.customerServicePairedSwitchLogicalMediaPath(record, sourceLogical)
	if err != nil {
		return err
	}
	destination, err := manager.customerServicePairedSwitchLogicalMediaPath(record, destinationLogical)
	if err != nil {
		return err
	}
	_, sourceErr := os.Lstat(source)
	_, destinationErr := os.Lstat(destination)
	if reverse && errors.Is(sourceErr, os.ErrNotExist) && destinationErr == nil {
		digest, err := manager.Configuration.inspectCustomerServicePairedSwitchActualMediaEntryDigest(ctx, filepath.Dir(destination), move.RootName, move.EntryName)
		if err != nil || digest != move.EntrySHA256 {
			return errors.New("paired already-reversed media entry identity changed")
		}
		return nil
	}
	if sourceErr != nil || !errors.Is(destinationErr, os.ErrNotExist) {
		return errors.New("paired media rename endpoints are ambiguous")
	}
	digest, err := manager.Configuration.inspectCustomerServicePairedSwitchActualMediaEntryDigest(ctx, filepath.Dir(source), move.RootName, move.EntryName)
	if err != nil || digest != move.EntrySHA256 {
		return errors.New("paired media rename source identity changed")
	}
	if err := os.Rename(source, destination); err != nil {
		return err
	}
	return errors.Join(syncDirectory(filepath.Dir(source)), syncDirectory(filepath.Dir(destination)))
}
