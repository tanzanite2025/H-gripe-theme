package localbackup

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// Observe sequence definitions, owning columns and current values without advancing counters. Live transactional advancement and disposable-candidate adjustments remain separate.
type businessModuleRestoreSequenceState struct {
	Name, Table, Column, Definition        string
	LastValue, Increment, Minimum, Maximum int64
	Called, Cycle                          bool
}

func readBusinessModuleRestoreSequenceStates(ctx context.Context, transaction *sql.Tx) ([]businessModuleRestoreSequenceState, error) {
	query := `SELECT sequence.relname,coalesce(owner.relname,''),coalesce(attribute.attname,''),jsonb_build_array(definition.seqstart,definition.seqincrement,definition.seqmax,definition.seqmin,definition.seqcache,definition.seqcycle)::text,definition.seqincrement,definition.seqmin,definition.seqmax,definition.seqcycle FROM pg_class sequence JOIN pg_sequence definition ON definition.seqrelid=sequence.oid LEFT JOIN pg_depend dependency ON dependency.classid='pg_class'::regclass AND dependency.objid=sequence.oid AND dependency.refclassid='pg_class'::regclass AND dependency.deptype IN ('a','i') LEFT JOIN pg_class owner ON owner.oid=dependency.refobjid LEFT JOIN pg_attribute attribute ON attribute.attrelid=owner.oid AND attribute.attnum=dependency.refobjsubid WHERE sequence.relnamespace='public'::regnamespace ORDER BY sequence.relname`
	result, err := transaction.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	var states []businessModuleRestoreSequenceState
	for result.Next() {
		var state businessModuleRestoreSequenceState
		if err := result.Scan(&state.Name, &state.Table, &state.Column, &state.Definition, &state.Increment, &state.Minimum, &state.Maximum, &state.Cycle); err != nil {
			_ = result.Close()
			return nil, err
		}
		states = append(states, state)
	}
	err = result.Err()
	_ = result.Close()
	if err != nil {
		return nil, err
	}
	for index := range states {
		state := &states[index]
		if err := transaction.QueryRowContext(ctx, `SELECT last_value,is_called FROM public.`+pq.QuoteIdentifier(state.Name)).Scan(&state.LastValue, &state.Called); err != nil {
			return nil, err
		}
	}
	return states, nil
}
