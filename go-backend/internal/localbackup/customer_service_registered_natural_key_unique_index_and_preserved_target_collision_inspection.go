package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"github.com/lib/pq"
)

type moduleRecoveryIndexDefinition struct {
	Name, Definition, Predicate                                                            string
	Columns                                                                                []string
	Primary, Unique, Exclusion, Valid, Ready, Immediate, NullsNotDistinct, DefaultEquality bool
}

// Observe all index definitions for snapshot binding; only explicitly registered
// simple equality keys may participate in customer-service collision checks.
func inspectModuleRecoveryDatabaseIndexDefinitions(ctx context.Context, transaction *sql.Tx) (map[string][]moduleRecoveryIndexDefinition, error) {
	rows, err := transaction.QueryContext(ctx, `SELECT relation.relname,index_relation.relname,pg_get_indexdef(index_record.indexrelid,0,true),coalesce(pg_get_expr(index_record.indpred,index_record.indrelid,true),''),ARRAY(SELECT coalesce(attribute.attname,'') FROM unnest(index_record.indkey) WITH ORDINALITY key_number(number,position) LEFT JOIN pg_attribute attribute ON attribute.attrelid=relation.oid AND attribute.attnum=key_number.number WHERE key_number.position<=index_record.indnkeyatts ORDER BY key_number.position),index_record.indisprimary,index_record.indisunique,index_record.indisexclusion,index_record.indisvalid,index_record.indisready,index_record.indimmediate,index_record.indnullsnotdistinct,(access_method.amname='btree' AND index_record.indexprs IS NULL AND NOT EXISTS(SELECT 1 FROM unnest(index_record.indkey,index_record.indclass,index_record.indcollation) WITH ORDINALITY key_spec(attribute_number,operator_class,collation_oid,position) LEFT JOIN pg_attribute attribute ON attribute.attrelid=relation.oid AND attribute.attnum=key_spec.attribute_number LEFT JOIN pg_opclass operator_class ON operator_class.oid=key_spec.operator_class LEFT JOIN pg_collation comparison_collation ON comparison_collation.oid=key_spec.collation_oid WHERE key_spec.position<=index_record.indnkeyatts AND (NOT coalesce(operator_class.opcdefault,false) OR NOT coalesce(comparison_collation.collisdeterministic,true) OR attribute.atttypid NOT IN ('int2'::regtype,'int4'::regtype,'int8'::regtype,'text'::regtype,'varchar'::regtype) OR attribute.attcollation IS DISTINCT FROM key_spec.collation_oid))) FROM pg_index index_record JOIN pg_class relation ON relation.oid=index_record.indrelid JOIN pg_class index_relation ON index_relation.oid=index_record.indexrelid JOIN pg_am access_method ON access_method.oid=index_relation.relam WHERE relation.relnamespace='public'::regnamespace ORDER BY relation.relname,index_relation.relname`)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
	definitions := map[string][]moduleRecoveryIndexDefinition{}
	for rows.Next() {
		var table string
		var definition moduleRecoveryIndexDefinition
		if err := rows.Scan(&table, &definition.Name, &definition.Definition, &definition.Predicate, pq.Array(&definition.Columns), &definition.Primary, &definition.Unique, &definition.Exclusion, &definition.Valid, &definition.Ready, &definition.Immediate, &definition.NullsNotDistinct, &definition.DefaultEquality); err != nil {
			return nil, err
		}
		definitions[table] = append(definitions[table], definition)
	}
	return definitions, rows.Err()
}

type customerServiceNaturalKeyContract struct {
	Table            string
	Columns          []string
	Nullable         bool
	NonNullPredicate bool
}

func registeredCustomerServiceConversationNaturalKeys() []customerServiceNaturalKeyContract {
	return []customerServiceNaturalKeyContract{
		{"tickets", []string{"ticket_number"}, false, false},
		{"tickets", []string{"conversation_id"}, true, true},
		{"customer_service_inbox_states", []string{"recipient_user_id", "ticket_id"}, false, false},
		{"ticket_message_media_attachments", []string{"ticket_message_id", "attachment_index"}, false, false},
		{"ticket_message_media_attachments", []string{"ticket_message_id", "media_asset_id"}, false, false},
		{"ticket_message_media_attachments", []string{"upload_session_id"}, true, false},
		{"media_upload_sessions", []string{"object_key"}, false, false},
		{"media_upload_sessions", []string{"media_asset_id"}, true, false},
	}
}

func customerServiceNaturalKeyColumnSignature(columns []string) string {
	copyOfColumns := append([]string{}, columns...)
	sort.Strings(copyOfColumns)
	return strings.Join(copyOfColumns, ",")
}

func customerServiceNaturalKeyIndexMatchesRegisteredContract(index moduleRecoveryIndexDefinition, contract customerServiceNaturalKeyContract) bool {
	if !index.Unique || index.Exclusion || !index.Valid || !index.Ready || !index.Immediate || index.NullsNotDistinct || !index.DefaultEquality || customerServiceNaturalKeyColumnSignature(index.Columns) != customerServiceNaturalKeyColumnSignature(contract.Columns) {
		return false
	}
	predicate := strings.Trim(strings.TrimSpace(index.Predicate), "()")
	return predicate == "" || contract.NonNullPredicate && predicate == contract.Columns[0]+" IS NOT NULL"
}

func (graph *moduleRecoveryRowDependencyGraph) inspectRegisteredCustomerServiceNaturalKeyIndexCoverage() {
	contracts := registeredCustomerServiceConversationNaturalKeys()
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments", "media_upload_sessions"} {
		for _, index := range graph.Observation.Indexes[table] {
			if !index.Unique && !index.Exclusion {
				continue
			}
			registered := index.Primary && len(index.Columns) == 1 && index.Columns[0] == "id" && index.DefaultEquality && index.Valid && index.Ready && index.Immediate
			for _, contract := range contracts {
				registered = registered || contract.Table == table && customerServiceNaturalKeyIndexMatchesRegisteredContract(index, contract)
			}
			if !registered {
				graph.markIncomplete("conversation_unique_index_contract_unsupported", table, "存在未登记、失效或非默认比较语义的唯一/排他索引，不能推断候选写入是否安全。")
			}
		}
	}
	for _, contract := range contracts {
		covered := false
		for _, index := range graph.Observation.Indexes[contract.Table] {
			covered = covered || customerServiceNaturalKeyIndexMatchesRegisteredContract(index, contract)
		}
		for _, column := range contract.Columns {
			if !graph.Observation.Columns[contract.Table][column] {
				covered = false
			}
		}
		if !covered {
			graph.markIncomplete("conversation_natural_key_index_missing", contract.Table, "缺少客服自然键字段或受支持的唯一索引，逐键检查未覆盖该约束。")
		}
	}
}

func (graph *moduleRecoveryRowDependencyGraph) blockCustomerServiceCandidateRowForContractIssue(table string, row *moduleRecoveryDependencyRow, code, message string) {
	if graph.BlockedCandidateRows[table] == nil {
		graph.BlockedCandidateRows[table] = map[string]bool{}
	}
	graph.BlockedCandidateRows[table][row.Identity] = true
	graph.Report.recordIssue(code, "blocking", graph.Side, table, moduleRecoveryDependencyRowPublicIdentity(row, graph.Observation.PrimaryKeys[table]), message)
}

func inspectCustomerServiceConversationNaturalKeyCollisions(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph) error {
	for _, graph := range []*moduleRecoveryRowDependencyGraph{source, target} {
		if graph != nil {
			graph.inspectRegisteredCustomerServiceNaturalKeyIndexCoverage()
		}
	}
	for _, contract := range registeredCustomerServiceConversationNaturalKeys() {
		var sourceRows []*moduleRecoveryDependencyRow
		for _, row := range source.Rows[contract.Table] {
			sourceRows = append(sourceRows, row)
		}
		sort.Slice(sourceRows, func(i, j int) bool { return sourceRows[i].Identity < sourceRows[j].Identity })
		for start := 0; start < len(sourceRows) && !source.Exhausted; start += 200 {
			batch := sourceRows[start:min(start+200, len(sourceRows))]
			keys := []map[string]json.RawMessage{}
			ownersByKey := map[string][]*moduleRecoveryDependencyRow{}
			for _, row := range batch {
				key := map[string]json.RawMessage{}
				missing := false
				for _, column := range contract.Columns {
					value := row.Values[column]
					if len(value) == 0 || string(value) == "null" {
						missing = true
						break
					}
					if (column == "ticket_number" || column == "object_key") && strings.TrimSpace(row.referenceColumnText(column)) == "" {
						missing = true
						break
					}
					key[column] = value
				}
				if missing {
					if !contract.Nullable {
						source.blockCustomerServiceCandidateRowForContractIssue(contract.Table, row, "conversation_required_natural_key_missing", "客服行缺少必需的自然键，不能构造候选写入。")
					}
					continue
				}
				keys = append(keys, key)
				identity := moduleRecoveryDependencyRowKey(key, contract.Columns)
				ownersByKey[identity] = append(ownersByKey[identity], row)
			}
			for _, graph := range []*moduleRecoveryRowDependencyGraph{source, target} {
				if graph == nil {
					continue
				}
				matches, err := graph.readRowsMatchingReferenceKeys(ctx, contract.Table, contract.Columns, keys, graph.Limits.Rows)
				if err != nil {
					return err
				}
				for _, match := range matches {
					for _, owner := range ownersByKey[moduleRecoveryDependencyRowKey(match.Values, contract.Columns)] {
						source.Report.ConversationReferenceChecks.NaturalKeyComparisons++
						if match.Identity != owner.Identity {
							source.blockCustomerServiceCandidateRowForContractIssue(contract.Table, owner, "conversation_natural_key_owned_by_different_row", "客服自然键已由包内或目标的另一主键占用；保留当前行，禁止自动合并、改号或覆盖。")
						}
					}
				}
			}
		}
	}
	return nil
}
