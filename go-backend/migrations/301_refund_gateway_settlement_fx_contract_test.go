package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestRefundGatewaySettlementFXMigrationContract(t *testing.T) {
	up, err := os.ReadFile("301_refund_gateway_settlement_fx.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile("301_refund_gateway_settlement_fx.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	upSQL := strings.ToLower(string(up))
	downSQL := strings.ToLower(string(down))
	for _, fragment := range []string{
		"alter table refunds",
		"alter table payment_refund_executions",
		"settlement_amount_minor",
		"settlement_currency",
		"settlement_balance_transaction_id",
		"fx_gain_loss_minor",
		"fx_gain_loss_currency",
		"idx_refunds_settlement_balance_transaction",
		"idx_payment_refund_executions_settlement_balance_transaction",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	for _, fragment := range []string{
		"drop index if exists idx_refunds_settlement_balance_transaction",
		"drop index if exists idx_payment_refund_executions_settlement_balance_transaction",
		"drop column if exists settlement_amount_minor",
		"drop column if exists fx_gain_loss_currency",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}
