package localbackup

import (
	"context"
	"encoding/json"
	"sort"
)

type moduleRecoveryRowComparisonPolicy struct {
	ScopedContentDifferenceCode, ScopedContentDifferenceMessage string
	IdentityConflictMessage, TargetOnlyCode, TargetOnlyMessage  string
	OwnershipConflicts                                          func(string, *moduleRecoveryDependencyRow, *moduleRecoveryDependencyRow) bool
}

func compareModuleRecoveryDependencyRowGraphs(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph, policy moduleRecoveryRowComparisonPolicy) error {
	tables := map[string]bool{}
	for table := range source.Rows {
		tables[table] = true
	}
	for table := range target.Rows {
		tables[table] = true
	}
	var names []string
	for table := range tables {
		names = append(names, table)
	}
	sort.Strings(names)
	for _, table := range names {
		comparison := ModuleRecoveryRowComparison{Table: table, SourceSelected: len(source.Rows[table]), TargetSelected: len(target.Rows[table]), SourceScoped: len(source.Scoped[table]), TargetScoped: len(target.Scoped[table])}
		var sourceRows []*moduleRecoveryDependencyRow
		for _, row := range source.Rows[table] {
			sourceRows = append(sourceRows, row)
		}
		sort.Slice(sourceRows, func(i, j int) bool { return sourceRows[i].Identity < sourceRows[j].Identity })
		if source.Observation.Shapes[table] != target.Observation.Shapes[table] {
			source.Report.Complete = false
			source.Report.recordIssue("row_schema_conflict", "blocking", "both", table, "", "表结构不同，逐行比较未执行。")
		} else {
			for start := 0; start < len(sourceRows) && !target.Exhausted; start += 200 {
				batch := sourceRows[start:min(start+200, len(sourceRows))]
				var keys []map[string]json.RawMessage
				for _, row := range batch {
					key := map[string]json.RawMessage{}
					for _, column := range source.Observation.PrimaryKeys[table] {
						key[column] = row.Values[column]
					}
					keys = append(keys, key)
				}
				matches, err := target.readRowsMatchingReferenceKeys(ctx, table, target.Observation.PrimaryKeys[table], keys, len(batch))
				if err != nil {
					return err
				}
				if target.Exhausted {
					break
				}
				matched := map[string]*moduleRecoveryDependencyRow{}
				if target.ComparedTargetRows[table] == nil {
					target.ComparedTargetRows[table] = map[string]*moduleRecoveryDependencyRow{}
				}
				for _, row := range matches {
					matched[row.Identity] = row
					target.ComparedTargetRows[table][row.Identity] = row
				}
				for _, row := range batch {
					current := matched[row.Identity]
					identity := moduleRecoveryDependencyRowPublicIdentity(row, source.Observation.PrimaryKeys[table])
					if current == nil {
						comparison.TargetMissing++
						source.Report.recordIssue("target_dependency_row_missing", "blocking", "target", table, identity, "包内所需行在目标中不存在，补全前须验证身份、唯一键与共享引用。")
						if policy.ScopedContentDifferenceCode == "conversation_row_content_differs" && source.Scoped[table][row.Identity] {
							source.recordCustomerServiceDecisionIssue(table, row.Identity, "target_dependency_row_missing")
						}
						continue
					}
					if policy.OwnershipConflicts(table, row, current) {
						comparison.IdentityConflicts++
						source.Report.recordIssue("row_identity_reassigned", "blocking", "target", table, identity, policy.IdentityConflictMessage)
					}
					if row.Digest == current.Digest {
						comparison.Matching++
					} else {
						comparison.ContentDiffers++
						code, message := "shared_dependency_row_differs", "所引用的共享行内容不同；不会覆盖其他客户或业务的数据。"
						if source.Scoped[table][row.Identity] {
							code, message = policy.ScopedContentDifferenceCode, policy.ScopedContentDifferenceMessage
						}
						source.Report.recordIssue(code, "blocking", "target", table, identity, message)
						if code == "conversation_row_content_differs" {
							source.recordCustomerServiceDecisionIssue(table, row.Identity, code)
						}
					}
				}
			}
		}
		for key := range target.Scoped[table] {
			if !source.Scoped[table][key] {
				comparison.TargetOnlyScoped++
				source.Report.recordIssue(policy.TargetOnlyCode, "blocking", "target", table, moduleRecoveryDependencyRowPublicIdentity(target.Rows[table][key], target.Observation.PrimaryKeys[table]), policy.TargetOnlyMessage)
				if policy.TargetOnlyCode == "target_only_conversation_row" {
					source.recordCustomerServiceDecisionIssue(table, key, policy.TargetOnlyCode)
				}
			}
		}
		source.Report.Tables = append(source.Report.Tables, comparison)
	}
	return nil
}
