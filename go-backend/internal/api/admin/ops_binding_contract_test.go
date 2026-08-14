package admin

import (
	"encoding/json"
	"testing"
)

func TestOpsBindingRequestsPreserveConnectorPresence(t *testing.T) {
	var projectOmitted opsProjectBindingRequest
	if err := json.Unmarshal([]byte(`{"name":"project","vps_binding_id":1}`), &projectOmitted); err != nil {
		t.Fatalf("unmarshal project omitted connector: %v", err)
	}
	projectInput := projectOmitted.toServiceInput()
	if projectInput.ConnectorIDSet {
		t.Fatal("project omitted connector was marked as set")
	}

	var projectNull opsProjectBindingRequest
	if err := json.Unmarshal([]byte(`{"name":"project","vps_binding_id":1,"connector_id":null}`), &projectNull); err != nil {
		t.Fatalf("unmarshal project null connector: %v", err)
	}
	projectInput = projectNull.toServiceInput()
	if !projectInput.ConnectorIDSet || projectInput.ConnectorID != nil {
		t.Fatalf("project explicit null connector = set:%t value:%#v", projectInput.ConnectorIDSet, projectInput.ConnectorID)
	}

	var vpsOmitted opsVPSBindingRequest
	if err := json.Unmarshal([]byte(`{"name":"vps","provider":"hostinger"}`), &vpsOmitted); err != nil {
		t.Fatalf("unmarshal VPS omitted connector: %v", err)
	}
	vpsInput := vpsOmitted.toServiceInput()
	if vpsInput.ConnectorIDSet {
		t.Fatal("VPS omitted connector was marked as set")
	}

	var vpsNull opsVPSBindingRequest
	if err := json.Unmarshal([]byte(`{"name":"vps","provider":"hostinger","connector_id":null}`), &vpsNull); err != nil {
		t.Fatalf("unmarshal VPS null connector: %v", err)
	}
	vpsInput = vpsNull.toServiceInput()
	if !vpsInput.ConnectorIDSet || vpsInput.ConnectorID != nil {
		t.Fatalf("VPS explicit null connector = set:%t value:%#v", vpsInput.ConnectorIDSet, vpsInput.ConnectorID)
	}
}

func TestOpsBindingRequestsDoNotAcceptObservedStateAsWriteInput(t *testing.T) {
	var projectRequest opsProjectBindingRequest
	if err := json.Unmarshal([]byte(`{
		"name":"project",
		"vps_binding_id":1,
		"health_status":"offline",
		"last_checked_at":"2026-08-13T10:00:00Z",
		"last_error":"forged"
	}`), &projectRequest); err != nil {
		t.Fatalf("unmarshal project observed fields: %v", err)
	}
	projectInput := projectRequest.toServiceInput()
	projectJSON, err := json.Marshal(projectInput)
	if err != nil {
		t.Fatalf("marshal project service input: %v", err)
	}
	if string(projectJSON) == "" || string(projectJSON) == "null" {
		t.Fatalf("project service input JSON is empty: %s", projectJSON)
	}
	for _, field := range []string{"health_status", "last_checked_at", "last_error"} {
		if string(projectJSON) != "" && containsJSONField(projectJSON, field) {
			t.Fatalf("project service input still exposes observed field %q: %s", field, projectJSON)
		}
	}

	var vpsRequest opsVPSBindingRequest
	if err := json.Unmarshal([]byte(`{
		"name":"vps",
		"provider":"hostinger",
		"observed_status":"offline",
		"last_error":"forged"
	}`), &vpsRequest); err != nil {
		t.Fatalf("unmarshal VPS observed fields: %v", err)
	}
	vpsInput := vpsRequest.toServiceInput()
	vpsJSON, err := json.Marshal(vpsInput)
	if err != nil {
		t.Fatalf("marshal VPS service input: %v", err)
	}
	for _, field := range []string{"observed_status", "last_error"} {
		if containsJSONField(vpsJSON, field) {
			t.Fatalf("VPS service input still exposes observed field %q: %s", field, vpsJSON)
		}
	}
}

func TestOpsProjectBindingRequestPreservesQuickBuyRateLimitPolicy(t *testing.T) {
	var projectRequest opsProjectBindingRequest
	if err := json.Unmarshal([]byte(`{
		"name":"project",
		"vps_binding_id":1,
		"quick_buy_rate_limit_policy":"{\"enabled\":true,\"ip_requests_per_minute\":90,\"ip_burst\":30,\"session_requests_per_minute\":45,\"session_burst\":15}"
	}`), &projectRequest); err != nil {
		t.Fatalf("unmarshal project quick buy policy: %v", err)
	}

	projectInput := projectRequest.toServiceInput()
	if projectInput.QuickBuyRateLimitPolicy == "" {
		t.Fatal("quick buy rate limit policy was not preserved")
	}
	if !containsJSONField(mustMarshalJSON(t, projectInput), "quick_buy_rate_limit_policy") {
		t.Fatalf("project service input JSON does not expose quick buy policy: %#v", projectInput)
	}
}

func mustMarshalJSON(t *testing.T, value interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	return data
}

func containsJSONField(data []byte, field string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(data, &payload); err != nil {
		return false
	}
	_, ok := payload[field]
	return ok
}
