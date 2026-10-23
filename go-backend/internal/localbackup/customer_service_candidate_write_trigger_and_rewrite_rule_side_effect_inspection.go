package localbackup

import (
	"context"
	"database/sql"
)

type moduleRecoveryDatabaseWriteSideEffectDefinition struct {
	Kind             string `json:"kind"`
	Name             string `json:"name"`
	Enabled          string `json:"enabled"`
	DefinitionSHA256 string `json:"definition_sha256"`
	FunctionSHA256   string `json:"function_sha256,omitempty"`
}

// Definitions and trigger bodies stay in PostgreSQL. No custom write trigger or
// table rewrite rule has an approved customer-service candidate contract yet.
func inspectModuleRecoveryDatabaseWriteSideEffectDefinitions(ctx context.Context, transaction *sql.Tx) (map[string][]moduleRecoveryDatabaseWriteSideEffectDefinition, error) {
	rows, err := transaction.QueryContext(ctx, `SELECT relation.relname,'trigger',trigger_record.tgname,trigger_record.tgenabled::text,encode(sha256(convert_to(pg_get_triggerdef(trigger_record.oid,true),'UTF8')),'hex'),encode(sha256(convert_to(pg_get_functiondef(trigger_record.tgfoid),'UTF8')),'hex') FROM pg_trigger trigger_record JOIN pg_class relation ON relation.oid=trigger_record.tgrelid WHERE relation.relnamespace='public'::regnamespace AND NOT trigger_record.tgisinternal UNION ALL SELECT relation.relname,'rewrite_rule',rewrite_record.rulename,rewrite_record.ev_enabled::text,encode(sha256(convert_to(pg_get_ruledef(rewrite_record.oid,true),'UTF8')),'hex'),'' FROM pg_rewrite rewrite_record JOIN pg_class relation ON relation.oid=rewrite_record.ev_class WHERE relation.relnamespace='public'::regnamespace AND relation.relkind IN ('r','p') ORDER BY 1,2,3`)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
	definitions := map[string][]moduleRecoveryDatabaseWriteSideEffectDefinition{}
	for rows.Next() {
		var table string
		var definition moduleRecoveryDatabaseWriteSideEffectDefinition
		if err := rows.Scan(&table, &definition.Kind, &definition.Name, &definition.Enabled, &definition.DefinitionSHA256, &definition.FunctionSHA256); err != nil {
			return nil, err
		}
		definitions[table] = append(definitions[table], definition)
	}
	return definitions, rows.Err()
}

func (graph *moduleRecoveryRowDependencyGraph) inspectCustomerServiceCandidateWriteSideEffectContracts() {
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"} {
		for range graph.Observation.WriteSideEffects[table] {
			graph.markIncomplete("conversation_write_side_effect_contract_unsupported", table, "客服候选表存在尚未登记的写入触发器或重写规则，包括禁用规则；须核对副作用，不能由内容选择豁免。")
		}
	}
}
