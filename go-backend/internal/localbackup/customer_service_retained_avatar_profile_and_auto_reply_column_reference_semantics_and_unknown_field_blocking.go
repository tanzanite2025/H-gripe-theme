package localbackup

import (
	"errors"
	"sort"
	"strings"
)

type CustomerServiceRetainedMediaOwnerColumnContractCoverage struct {
	Version                         int                                                   `json:"version"`
	Side                            string                                                `json:"side"`
	Complete                        bool                                                  `json:"complete"`
	ObservedColumns                 int                                                   `json:"observed_columns"`
	RegisteredColumns               int                                                   `json:"registered_columns"`
	UnregisteredColumns             int                                                   `json:"unregistered_columns"`
	UnregisteredColumnCountsByTable map[string]int                                        `json:"unregistered_column_counts_by_table"`
	MissingRetainedOwnerTables      []string                                              `json:"missing_retained_owner_tables"`
	ColumnContractsSHA256           string                                                `json:"column_contracts_sha256"`
	UnregisteredColumnSamples       []CustomerServiceRetainedMediaOwnerUnregisteredColumn `json:"unregistered_column_samples"`
	SamplesTruncated                bool                                                  `json:"samples_truncated"`
	SampleLimit                     int                                                   `json:"sample_limit"`
}

type CustomerServiceRetainedMediaOwnerUnregisteredColumn struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

type customerServiceRetainedMediaOwnerObservedColumnContract struct {
	Table, Column, Semantics string
}

// Retaining full rows does not discover references inside new fields. These
// two owner tables are preserved, including disabled or soft-deleted rows.
// An extra FK is not enough: its objects would also need to enter the retained
// media/cleanup scope, rather than only have referential integrity checked.
func registeredCustomerServiceRetainedMediaOwnerColumnReferenceSemantics() map[string]map[string]string {
	registry := map[string]map[string]string{}
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup := func(table, semantics, columns string) {
		if registry[table] == nil {
			registry[table] = map[string]string{}
		}
		for _, column := range strings.Fields(columns) {
			registry[table][column] = semantics
		}
	}
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("customer_service_agent_profiles", "preserved_profile_and_account_identity", "id agent_id user_id")
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("customer_service_agent_profiles", "registered_retained_avatar_object_identity", "avatar")
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("customer_service_agent_profiles", "preserved_plain_text_contact_and_display", "name email whats_app")
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("customer_service_agent_profiles", "preserved_profile_availability_and_lifecycle", "status online_status last_active_at last_login created_at updated_at deleted_at")
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("ticket_auto_replies", "preserved_rule_account_and_group_identity", "id agent_id group_id")
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("ticket_auto_replies", "preserved_plain_text_matching_and_reply", "trigger_keyword reply_message")
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("ticket_auto_replies", "registered_retained_message_type_metadata_and_attachment_paths", "message_type metadata attachments")
	registerCustomerServiceRetainedMediaOwnerColumnSemanticsGroup("ticket_auto_replies", "preserved_rule_activation_matching_and_lifecycle", "type is_active priority match_type locale cooldown_seconds created_at updated_at")
	return registry
}

func inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(observation *moduleRecoveryDatabaseObservation, side string) (*CustomerServiceRetainedMediaOwnerColumnContractCoverage, error) {
	coverage := &CustomerServiceRetainedMediaOwnerColumnContractCoverage{Version: 1, Side: side, Complete: true, MissingRetainedOwnerTables: []string{}, UnregisteredColumnCountsByTable: map[string]int{}, UnregisteredColumnSamples: []CustomerServiceRetainedMediaOwnerUnregisteredColumn{}, SampleLimit: 200}
	registry := registeredCustomerServiceRetainedMediaOwnerColumnReferenceSemantics()
	var contracts []customerServiceRetainedMediaOwnerObservedColumnContract
	for _, table := range []string{"customer_service_agent_profiles", "ticket_auto_replies"} {
		var columns []string
		for column, exists := range observation.Columns[table] {
			if exists {
				columns = append(columns, column)
			}
		}
		if len(columns) == 0 {
			coverage.MissingRetainedOwnerTables = append(coverage.MissingRetainedOwnerTables, table)
			coverage.Complete = false
		}
		sort.Strings(columns)
		for _, column := range columns {
			coverage.ObservedColumns++
			semantics := registry[table][column]
			if semantics != "" {
				coverage.RegisteredColumns++
			} else {
				semantics = "unregistered_retained_media_owner_reference_contract_required"
				coverage.UnregisteredColumns++
				coverage.UnregisteredColumnCountsByTable[table]++
				coverage.Complete = false
				if len(coverage.UnregisteredColumnSamples) < coverage.SampleLimit {
					coverage.UnregisteredColumnSamples = append(coverage.UnregisteredColumnSamples, CustomerServiceRetainedMediaOwnerUnregisteredColumn{table, column})
				}
			}
			contracts = append(contracts, customerServiceRetainedMediaOwnerObservedColumnContract{table, column, semantics})
		}
	}
	coverage.SamplesTruncated = coverage.UnregisteredColumns > len(coverage.UnregisteredColumnSamples)
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(contracts)
	coverage.ColumnContractsSHA256 = digest
	return coverage, err
}

func (graph *moduleRecoveryRowDependencyGraph) inspectCustomerServiceRetainedMediaOwnerColumnReferenceSemantics() error {
	coverage, err := inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(graph.Observation, graph.Side)
	if err != nil {
		return err
	}
	if graph.Report.ConversationReferenceChecks != nil {
		graph.Report.ConversationReferenceChecks.RetainedMediaOwnerColumnContracts = append(graph.Report.ConversationReferenceChecks.RetainedMediaOwnerColumnContracts, *coverage)
	}
	for _, table := range coverage.MissingRetainedOwnerTables {
		graph.markIncomplete("retained_media_owner_column_contract_table_missing", table, "缺少保留头像档案或自动回复表，不能完成保留对象字段引用契约检查。")
	}
	for _, table := range []string{"customer_service_agent_profiles", "ticket_auto_replies"} {
		for count := 0; count < coverage.UnregisteredColumnCountsByTable[table]; count++ {
			graph.markIncomplete("retained_media_owner_column_reference_contract_unregistered", table, "保留头像档案或自动回复包含未登记引用语义的字段，不能确认清理对象范围；空值、禁用或软删除均不豁免。")
		}
	}
	return nil
}

func assertCustomerServiceRetainedMediaOwnerColumnReferenceContractsRegistered(observation *moduleRecoveryDatabaseObservation) error {
	coverage, err := inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(observation, "retained_owner_candidate_guard")
	if err != nil {
		return err
	}
	if !coverage.Complete {
		return errors.New("retained avatar profile or auto reply columns contain an unregistered reference contract or missing table")
	}
	return nil
}
