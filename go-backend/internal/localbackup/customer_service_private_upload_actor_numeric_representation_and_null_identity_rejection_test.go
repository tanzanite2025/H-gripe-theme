package localbackup

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomerServiceRecoveryUploadActorRequiresCanonicalUnsignedAccountAndNonNullIdentityFields(t *testing.T) {
	for _, scenario := range []struct {
		name, account, actor, visitor string
		valid                         bool
	}{
		{name: "canonical_customer", account: "1", valid: true},
		{name: "negative_account", account: "-1"},
		{name: "numeric_overflow", account: "18446744073709551616"},
		{name: "exponent_account", account: "1e0"},
		{name: "decimal_account", account: "1.0"},
		{name: "quoted_account", account: `"1"`},
		{name: "noncanonical_leading_zero", account: "01"},
		{name: "noncanonical_plus", account: "+1"},
		{name: "missing_account"},
		{name: "null_account", account: "null"},
		{name: "null_actor", account: "1", actor: "null"},
		{name: "null_visitor_hash", account: "1", visitor: "null"},
		{name: "nonstrings_cannot_be_coerced", account: "1", actor: "123"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			actor, visitor := scenario.actor, scenario.visitor
			if actor == "" {
				actor = `"customer"`
			}
			if visitor == "" {
				visitor = `""`
			}
			row := &moduleRecoveryDependencyRow{Values: map[string]json.RawMessage{"actor": json.RawMessage(actor), "account": json.RawMessage(scenario.account), "visitor": json.RawMessage(visitor)}}
			_, valid := decodeCustomerServiceRecoveryMediaUploadActorIdentity(row, "actor", "account", "visitor")
			require.Equal(t, scenario.valid, valid)
		})
	}
}
