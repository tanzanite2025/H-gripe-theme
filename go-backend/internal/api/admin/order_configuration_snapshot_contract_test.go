package admin

import (
	"encoding/json"
	"testing"

	orderdomain "commerce-platform/internal/domain/order"
	"gorm.io/datatypes"

	"github.com/stretchr/testify/require"
)

func TestAdminOrderItemContractExposesConfigurationSnapshot(t *testing.T) {
	item := orderdomain.OrderItem{
		ID:                        7,
		ProductName:               "Configured wheelset",
		ConfigurationSnapshotData: datatypes.JSON(`{"schema_version":1,"configuration_hash":"abc"}`),
	}

	payload, err := json.Marshal(item)
	require.NoError(t, err)
	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.JSONEq(t, `{"schema_version":1,"configuration_hash":"abc"}`, string(decoded["configuration_snapshot"]))
}
