package service

import (
	"testing"

	"commerce-platform/internal/domain/ops"
)

func TestHostingerVPSObservedStatus(t *testing.T) {
	tests := map[string]string{
		"running":   ops.VPSObservedHealthy,
		"stopped":   ops.VPSObservedOffline,
		"suspended": ops.VPSObservedOffline,
		"starting":  ops.VPSObservedDegraded,
		"error":     ops.VPSObservedDegraded,
		"unknown":   ops.VPSObservedUnknown,
	}

	for state, want := range tests {
		if got := hostingerVPSObservedStatus(state); got != want {
			t.Errorf("hostingerVPSObservedStatus(%q) = %q, want %q", state, got, want)
		}
	}
}

func TestHostingerProjectHealthUsesStatusFallbacks(t *testing.T) {
	containers := []hostingerDockerContainer{
		{Status: "running", Health: "healthy"},
		{State: "running", Health: "healthy"},
	}
	if got := hostingerProjectHealth("running", containers); got != ops.ProjectHealthHealthy {
		t.Fatalf("hostingerProjectHealth() = %q, want %q", got, ops.ProjectHealthHealthy)
	}

	containers[1].Status = "exited"
	containers[1].State = ""
	if got := hostingerProjectHealth("running", containers); got != ops.ProjectHealthDegraded {
		t.Fatalf("hostingerProjectHealth() = %q, want %q", got, ops.ProjectHealthDegraded)
	}

	if got := hostingerProjectHealth("stopped", containers); got != ops.ProjectHealthOffline {
		t.Fatalf("hostingerProjectHealth(stopped) = %q, want %q", got, ops.ProjectHealthOffline)
	}
}

func TestHostingerIPAddressAcceptsStringAndObjectShapes(t *testing.T) {
	var stringAddress hostingerIPAddress
	if err := stringAddress.UnmarshalJSON([]byte(`"2.25.85.201"`)); err != nil {
		t.Fatalf("string address decode failed: %v", err)
	}
	if stringAddress.Address != "2.25.85.201" {
		t.Fatalf("string address = %q, want IPv4", stringAddress.Address)
	}

	var objectAddress hostingerIPAddress
	if err := objectAddress.UnmarshalJSON([]byte(`{"ip":"2.25.85.201"}`)); err != nil {
		t.Fatalf("object address decode failed: %v", err)
	}
	if objectAddress.Address != "2.25.85.201" {
		t.Fatalf("object address = %q, want IPv4", objectAddress.Address)
	}
}

func TestHostingerNamedValueAcceptsStringAndObjectShapes(t *testing.T) {
	var stringValue hostingerNamedValue
	if err := stringValue.UnmarshalJSON([]byte(`"KVM 2"`)); err != nil {
		t.Fatalf("string value decode failed: %v", err)
	}
	if stringValue.Value != "KVM 2" {
		t.Fatalf("string value = %q, want plan name", stringValue.Value)
	}

	var objectValue hostingerNamedValue
	if err := objectValue.UnmarshalJSON([]byte(`{"name":"KVM 4","slug":"kvm-4"}`)); err != nil {
		t.Fatalf("object value decode failed: %v", err)
	}
	if objectValue.Value != "KVM 4" {
		t.Fatalf("object value = %q, want object name", objectValue.Value)
	}
}
