package admin

import (
	"encoding/json"
	"testing"
)

func TestOptionalUintRequestDistinguishesOmittedAndExplicitNull(t *testing.T) {
	var omitted opsDomainBindingRequest
	if err := json.Unmarshal([]byte(`{"domain":"example.com","role":"internal","environment":"production","provider":"other"}`), &omitted); err != nil {
		t.Fatalf("unmarshal omitted request: %v", err)
	}
	if omitted.ConnectorID.Set || omitted.ProjectBindingID.Set {
		t.Fatalf("omitted references marked as set: connector=%#v project=%#v", omitted.ConnectorID, omitted.ProjectBindingID)
	}

	var explicitNull opsDomainBindingRequest
	if err := json.Unmarshal([]byte(`{"domain":"example.com","role":"internal","environment":"production","provider":"other","connector_id":null,"project_binding_id":null}`), &explicitNull); err != nil {
		t.Fatalf("unmarshal explicit null request: %v", err)
	}
	if !explicitNull.ConnectorID.Set || !explicitNull.ProjectBindingID.Set {
		t.Fatalf("explicit null references not marked as set: connector=%#v project=%#v", explicitNull.ConnectorID, explicitNull.ProjectBindingID)
	}
	if explicitNull.ConnectorID.Value != nil || explicitNull.ProjectBindingID.Value != nil {
		t.Fatalf("explicit null references have values: connector=%#v project=%#v", explicitNull.ConnectorID, explicitNull.ProjectBindingID)
	}
}
