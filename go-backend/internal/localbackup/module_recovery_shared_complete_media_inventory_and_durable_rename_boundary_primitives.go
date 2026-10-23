package localbackup

import "context"

// These filesystem primitives originated in customer-service recovery. Their
// contracts contain no conversation selections, permissions or execution state;
// share them without granting another module customer-service recovery authority.
type moduleRecoveryCompleteMediaInventory = customerServiceCandidateCompleteMediaUnionInventory
type moduleRecoveryMediaEntryMove = customerServiceIndependentPairedSwitchMediaEntryMove

func (cfg Configuration) inventoryModuleRecoveryCompleteMedia(ctx context.Context) (*moduleRecoveryCompleteMediaInventory, error) {
	inventory, err := cfg.inventoryCustomerServiceCompleteCurrentMedia(ctx)
	if err != nil {
		return nil, err
	}
	// This construction-only capacity counter is deliberately not serialized.
	// Normalize it before binding an inventory that will survive a task reload.
	inventory.EncodedBytes = 0
	return inventory, nil
}

func buildModuleRecoveryPairedMediaMoves(baseline, candidate *moduleRecoveryCompleteMediaInventory) ([]moduleRecoveryMediaEntryMove, error) {
	return buildCustomerServicePairedSwitchExpectedMediaMoves(baseline, candidate)
}

func validateModuleRecoveryPairedMediaLayout(moves []moduleRecoveryMediaEntryMove, layout map[string]string) error {
	return validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(&customerServiceIndependentPairedSwitchJournal{Moves: moves}, layout)
}

func (cfg Configuration) inspectModuleRecoveryMediaEntryDigest(ctx context.Context, directory, rootName, entryName string) (string, error) {
	return cfg.inspectCustomerServicePairedSwitchActualMediaEntryDigest(ctx, directory, rootName, entryName)
}
