package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/lib/pq"
)

// ALTER SEQUENCE RESTART is transactional, unlike setval. A failed restore
// rolls back both complete rows and selected counters to their original state.
// The caller supplies its domain-approved table selection and owns the commit.
func advanceBusinessModuleRestoreOwnedSequencesTransactionallyAndVerifyPreservation(ctx context.Context, tx *sql.Tx, baseline []businessModuleRestoreSequenceState, selected map[string][]json.RawMessage) error {
	current, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, baseline) {
		return errors.New("order restore sequence state changed before advancement")
	}
	expected := append([]businessModuleRestoreSequenceState(nil), baseline...)
	for index, state := range baseline {
		if len(selected[state.Table]) == 0 {
			continue
		}
		if state.Cycle || state.Increment != 1 {
			return errors.New("order history restore requires noncycling increment-one owned sequences")
		}
		var maximum sql.NullInt64
		if err := tx.QueryRowContext(ctx, `SELECT max(`+pq.QuoteIdentifier(state.Column)+`) FROM public.`+pq.QuoteIdentifier(state.Table)).Scan(&maximum); err != nil {
			return err
		}
		if !maximum.Valid || maximum.Int64 < state.LastValue || maximum.Int64 == state.LastValue && state.Called {
			continue
		}
		if maximum.Int64 >= state.Maximum {
			return errors.New("order restore would exhaust an owned sequence")
		}
		if _, err := tx.ExecContext(ctx, `ALTER SEQUENCE public.`+pq.QuoteIdentifier(state.Name)+` RESTART WITH `+fmt.Sprint(maximum.Int64+1)); err != nil {
			return err
		}
		expected[index].LastValue, expected[index].Called = maximum.Int64+1, false
	}
	actual, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, actual) {
		return errors.New("order restore sequence adjustment or preservation verification failed")
	}
	return nil
}
