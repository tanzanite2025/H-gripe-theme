package localbackup

import (
	"errors"
	"sort"
	"strings"
)

type CustomerServiceCandidateColumnContractCoverage struct {
	Version                         int                                          `json:"version"`
	Side                            string                                       `json:"side"`
	Complete                        bool                                         `json:"complete"`
	ObservedColumns                 int                                          `json:"observed_columns"`
	RegisteredColumns               int                                          `json:"registered_columns"`
	ForeignKeyColumns               int                                          `json:"foreign_key_columns"`
	UnregisteredColumns             int                                          `json:"unregistered_columns"`
	UnregisteredColumnCountsByTable map[string]int                               `json:"unregistered_column_counts_by_table"`
	MissingCandidateTables          []string                                     `json:"missing_candidate_tables"`
	ColumnContractsSHA256           string                                       `json:"column_contracts_sha256"`
	UnregisteredColumnSamples       []CustomerServiceCandidateUnregisteredColumn `json:"unregistered_column_samples"`
	SamplesTruncated                bool                                         `json:"samples_truncated"`
	SampleLimit                     int                                          `json:"sample_limit"`
}

type CustomerServiceCandidateUnregisteredColumn struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

type customerServiceCandidateObservedColumnContract struct {
	Table, Column, Semantics string
}

// This registry approves the meaning of PRESENT columns, not a migration
// baseline or permission to restore. Required columns, types, constraints and
// natural keys have separate checks. A matching source/target schema alone
// cannot approve an additional JSON, HTML, storage or other opaque column.
func registeredCustomerServiceCandidateColumnReferenceSemantics() map[string]map[string]string {
	registry := map[string]map[string]string{}
	registerCustomerServiceCandidateColumnSemanticsGroup := func(table, semantics, columns string) {
		if registry[table] == nil {
			registry[table] = map[string]string{}
		}
		for _, column := range strings.Fields(columns) {
			registry[table][column] = semantics
		}
	}
	registerCustomerServiceCandidateColumnSemanticsGroup("tickets", "row_and_conversation_identity", "id ticket_number conversation_id category")
	registerCustomerServiceCandidateColumnSemanticsGroup("tickets", "registered_account_visitor_and_assignment_reference", "user_id customer_user_id visitor_session_hash assigned_to group_id")
	registerCustomerServiceCandidateColumnSemanticsGroup("tickets", "plain_text_display_no_html_interpretation", "subject tags")
	registerCustomerServiceCandidateColumnSemanticsGroup("tickets", "lifecycle_and_realtime_version", "priority status status_version created_at updated_at resolved_at closed_at deleted_at")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_messages", "row_identity", "id")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_messages", "registered_conversation_and_account_reference", "ticket_id user_id")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_messages", "plain_text_display_no_html_interpretation", "content")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_messages", "registered_structured_message_rendering_and_known_snapshot_paths", "message_type metadata")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_messages", "registered_legacy_and_bound_media_membership", "attachments attachment_reference_format")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_messages", "message_visibility_and_lifecycle", "is_staff is_internal is_read created_at deleted_at")
	registerCustomerServiceCandidateColumnSemanticsGroup("customer_service_inbox_states", "row_identity", "id")
	registerCustomerServiceCandidateColumnSemanticsGroup("customer_service_inbox_states", "registered_conversation_account_and_read_cursor_reference", "ticket_id recipient_user_id last_read_message_id")
	registerCustomerServiceCandidateColumnSemanticsGroup("customer_service_inbox_states", "inbox_realtime_version_and_lifecycle", "unread_count assignment_version last_read_at created_at updated_at deleted_at archived_at")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_message_media_attachments", "row_identity", "id")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_message_media_attachments", "registered_message_media_upload_and_membership_reference", "ticket_message_id media_asset_id upload_session_id attachment_index")
	registerCustomerServiceCandidateColumnSemanticsGroup("ticket_message_media_attachments", "binding_lifecycle", "created_at")
	return registry
}

func inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation *moduleRecoveryDatabaseObservation, side string) (*CustomerServiceCandidateColumnContractCoverage, error) {
	coverage := &CustomerServiceCandidateColumnContractCoverage{Version: 1, Side: side, Complete: true, MissingCandidateTables: []string{}, UnregisteredColumnCountsByTable: map[string]int{}, UnregisteredColumnSamples: []CustomerServiceCandidateUnregisteredColumn{}, SampleLimit: 200}
	registry := registeredCustomerServiceCandidateColumnReferenceSemantics()
	var contracts []customerServiceCandidateObservedColumnContract
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"} {
		if len(observation.Columns[table]) == 0 {
			coverage.MissingCandidateTables = append(coverage.MissingCandidateTables, table)
			coverage.Complete = false
			continue
		}
		var columns []string
		for column, exists := range observation.Columns[table] {
			if exists {
				columns = append(columns, column)
			}
		}
		sort.Strings(columns)
		for _, column := range columns {
			coverage.ObservedColumns++
			semantics := registry[table][column]
			switch {
			case semantics != "":
				coverage.RegisteredColumns++
			case customerServiceCandidateColumnHasObservedForeignKeyContract(observation, table, column):
				semantics = "observed_foreign_key_reference_existing_dependency_closure"
				coverage.ForeignKeyColumns++
			default:
				semantics = "unregistered_column_reference_contract_required"
				coverage.UnregisteredColumns++
				coverage.UnregisteredColumnCountsByTable[table]++
				coverage.Complete = false
				if len(coverage.UnregisteredColumnSamples) < coverage.SampleLimit {
					coverage.UnregisteredColumnSamples = append(coverage.UnregisteredColumnSamples, CustomerServiceCandidateUnregisteredColumn{table, column})
				}
			}
			contracts = append(contracts, customerServiceCandidateObservedColumnContract{table, column, semantics})
		}
	}
	coverage.SamplesTruncated = coverage.UnregisteredColumns > len(coverage.UnregisteredColumnSamples)
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(contracts)
	coverage.ColumnContractsSHA256 = digest
	return coverage, err
}

func (graph *moduleRecoveryRowDependencyGraph) inspectCustomerServiceCandidateColumnReferenceSemantics() error {
	coverage, err := inspectCustomerServiceCandidateColumnReferenceContractCoverage(graph.Observation, graph.Side)
	if err != nil {
		return err
	}
	if graph.Report.ConversationReferenceChecks != nil {
		graph.Report.ConversationReferenceChecks.ColumnContracts = append(graph.Report.ConversationReferenceChecks.ColumnContracts, *coverage)
	}
	for _, table := range coverage.MissingCandidateTables {
		graph.markIncomplete("conversation_column_contract_table_missing", table, "客服候选表不存在，逐列引用契约检查未完成。")
	}
	// Count every unregistered column even after the public sample limit. Values
	// and bodies never enter the report, or the private relationship projection.
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"} {
		for count := 0; count < coverage.UnregisteredColumnCountsByTable[table]; count++ {
			graph.markIncomplete("conversation_column_reference_contract_unregistered", table, "候选完整行包含未登记引用语义的字段；自定义 JSON、HTML 或其他字段须先登记契约，不能由内容选择豁免。")
		}
	}
	return nil
}

func customerServiceCandidateColumnHasObservedForeignKeyContract(observation *moduleRecoveryDatabaseObservation, table, column string) bool {
	for _, reference := range observation.References {
		if reference.Kind != "foreign_key" || reference.SourceTable != table || len(reference.SourceColumns) == 0 || len(reference.SourceColumns) != len(reference.TargetColumns) {
			continue
		}
		complete, matches := true, false
		for index, sourceColumn := range reference.SourceColumns {
			complete = complete && observation.Columns[table][sourceColumn] && observation.Columns[reference.TargetTable][reference.TargetColumns[index]]
			matches = matches || sourceColumn == column
		}
		if complete && matches {
			return true
		}
	}
	return false
}

func assertCustomerServiceCandidateColumnReferenceContractsRegistered(observation *moduleRecoveryDatabaseObservation) error {
	coverage, err := inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation, "candidate_write_guard")
	if err != nil {
		return err
	}
	if !coverage.Complete {
		return errors.New("candidate full-row columns contain an unregistered reference contract or missing table")
	}
	return nil
}
