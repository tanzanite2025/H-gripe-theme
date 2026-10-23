package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
)

type customerServiceCandidateFilesystemCapacityRequirement struct {
	FilesystemIdentity string   `json:"filesystem_identity"`
	Components         []string `json:"components"`
	AdditionalBytes    int64    `json:"additional_bytes"`
	AvailableBytes     uint64   `json:"available_bytes"`
}

type customerServiceCandidateCombinedPeakCapacityBudget struct {
	CurrentDatabaseBytes       int64                                                   `json:"current_database_bytes_already_allocated"`
	OriginalStageDatabaseBytes int64                                                   `json:"original_stage_database_bytes_already_allocated"`
	CurrentMediaBytes          int64                                                   `json:"current_media_bytes_already_allocated"`
	CandidateMediaBytes        int64                                                   `json:"candidate_media_logical_bytes"`
	TargetDumpByteLimit        int64                                                   `json:"target_dump_byte_limit"`
	Requirements               []customerServiceCandidateFilesystemCapacityRequirement `json:"combined_filesystem_requirements"`
}

func addCustomerServiceCandidateCapacityBytesWithoutOverflow(total, additional int64) (int64, error) {
	if total < 0 || additional < 0 || total > math.MaxInt64-additional {
		return 0, errors.New("candidate capacity estimate overflow")
	}
	return total + additional, nil
}

// Each component is charged to the underlying filesystem, including bind mounts.
// Current/stage databases and original media already occupy disk. A later paired
// rollback renames those originals; it does not require a second byte copy.
func (cfg Configuration) inspectCustomerServiceCandidateCombinedPeakCapacity(ctx context.Context, source, target *sql.Tx, inventory *customerServiceCandidateCompleteMediaUnionInventory) (*customerServiceCandidateCombinedPeakCapacityBudget, error) {
	budget := &customerServiceCandidateCombinedPeakCapacityBudget{CurrentMediaBytes: inventory.CurrentBytes, CandidateMediaBytes: inventory.TotalBytes}
	if err := target.QueryRowContext(ctx, `SELECT pg_database_size(current_database())`).Scan(&budget.CurrentDatabaseBytes); err != nil {
		return nil, err
	}
	if err := source.QueryRowContext(ctx, `SELECT pg_database_size(current_database())`).Scan(&budget.OriginalStageDatabaseBytes); err != nil {
		return nil, err
	}
	if budget.CurrentDatabaseBytes <= 0 || budget.OriginalStageDatabaseBytes <= 0 || budget.CurrentDatabaseBytes > (math.MaxInt64-(256<<20))/3 {
		return nil, errors.New("candidate database capacity baseline is invalid")
	}
	budget.TargetDumpByteLimit = budget.CurrentDatabaseBytes*2 + (16 << 20)
	workspaceNeed, err := addCustomerServiceCandidateCapacityBytesWithoutOverflow(budget.TargetDumpByteLimit, int64(inventory.EncodedBytes)+(16<<20))
	if err != nil {
		return nil, err
	}
	// The full cleanup ledger and its atomic replacement coexist during a write.
	workspaceNeed, err = addCustomerServiceCandidateCapacityBytesWithoutOverflow(workspaceNeed, 2*customerServiceCleanupDispositionMaximumAuditBytes)
	if err != nil {
		return nil, err
	}
	if err := cfg.checkWorkspaceCapacity(workspaceNeed); err != nil {
		return nil, err
	}
	type component struct {
		name, path string
		bytes      int64
	}
	components := []component{{"workspace_dump_inventory_and_reserve", cfg.Workspace, workspaceNeed + (64 << 20)}, {"database_clone_wal_and_reserve", cfg.PostgreSQLCapacityDirectory, budget.CurrentDatabaseBytes*3 + (256 << 20)}}
	for _, rootName := range mediaRoots {
		var bytes, entries int64
		for _, file := range inventory.Files {
			if file.RootName == rootName {
				bytes += file.Identity.Size
				entries++
			}
		}
		for _, directory := range inventory.Directories {
			if directory.RootName == rootName {
				entries++
			}
		}
		// Include directory/allocation overhead and a separate 64 MiB reserve.
		need, err := addCustomerServiceCandidateCapacityBytesWithoutOverflow(bytes, entries*(64<<10)+(64<<20))
		if err != nil {
			return nil, err
		}
		components = append(components, component{"media_" + rootName, cfg.StorageRoots[rootName], need})
	}
	groups := map[string]*customerServiceCandidateFilesystemCapacityRequirement{}
	for _, component := range components {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		identity, err := customerServiceCandidateCapacityFilesystemIdentity(component.path)
		if err != nil {
			return nil, errors.New("cannot identify candidate capacity filesystem")
		}
		free, err := availableDiskBytes(component.path)
		if err != nil {
			return nil, errors.New("cannot inspect candidate capacity filesystem")
		}
		group := groups[identity]
		if group == nil {
			group = &customerServiceCandidateFilesystemCapacityRequirement{FilesystemIdentity: identity, AvailableBytes: free}
			groups[identity] = group
		}
		if free < group.AvailableBytes {
			group.AvailableBytes = free
		}
		group.AdditionalBytes, err = addCustomerServiceCandidateCapacityBytesWithoutOverflow(group.AdditionalBytes, component.bytes)
		if err != nil {
			return nil, err
		}
		group.Components = append(group.Components, component.name)
	}
	for _, group := range groups {
		sort.Strings(group.Components)
		budget.Requirements = append(budget.Requirements, *group)
	}
	sort.Slice(budget.Requirements, func(i, j int) bool {
		return budget.Requirements[i].FilesystemIdentity < budget.Requirements[j].FilesystemIdentity
	})
	if err := validateCustomerServiceCandidateCombinedFilesystemCapacity(budget.Requirements); err != nil {
		return nil, err
	}
	return budget, nil
}

func validateCustomerServiceCandidateCombinedFilesystemCapacity(requirements []customerServiceCandidateFilesystemCapacityRequirement) error {
	for _, requirement := range requirements {
		if requirement.AdditionalBytes < 0 || uint64(requirement.AdditionalBytes) > requirement.AvailableBytes {
			return fmt.Errorf("insufficient combined candidate filesystem capacity for %d components", len(requirement.Components))
		}
	}
	return nil
}
