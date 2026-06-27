package auth

import "testing"

func TestNormalizeRole(t *testing.T) {
	tests := map[string]Role{
		"administrator":              RoleAdmin,
		"shop_manager":               RoleManager,
		"agent":                      RoleSupport,
		"customer_service":           RoleSupport,
		"administrator,shop_manager": RoleAdmin,
		"subscriber":                 RoleUser,
		"":                           RoleUser,
	}

	for raw, expected := range tests {
		if actual := NormalizeRole(raw); actual != expected {
			t.Fatalf("NormalizeRole(%q) = %q, want %q", raw, actual, expected)
		}
	}
}

func TestIsCustomerServiceAgentRole(t *testing.T) {
	for _, raw := range []string{"admin", "manager", "support", "administrator", "shop_manager", "customer_service"} {
		if !IsCustomerServiceAgentRole(raw) {
			t.Fatalf("expected %q to be a customer service agent role", raw)
		}
	}

	for _, raw := range []string{"editor", "viewer", "subscriber", "customer"} {
		if IsCustomerServiceAgentRole(raw) {
			t.Fatalf("expected %q to not be a customer service agent role", raw)
		}
	}
}

func TestRoleUserIsValid(t *testing.T) {
	if !RoleUser.IsValid() {
		t.Fatal("expected user role to be valid")
	}
}
