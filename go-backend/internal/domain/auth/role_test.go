package auth

import "testing"

func TestNormalizeRole(t *testing.T) {
	tests := map[string]Role{
		"admin":                      RoleAdmin,
		"ADMIN":                      RoleAdmin,
		"manager":                    RoleManager,
		"support":                    RoleSupport,
		"editor":                     RoleEditor,
		"viewer":                     RoleViewer,
		"administrator":              RoleUser,
		"shop_manager":               RoleUser,
		"agent":                      RoleUser,
		"customer_service":           RoleUser,
		"administrator,shop_manager": RoleUser,
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
	for _, raw := range []string{"admin", "manager", "support"} {
		if !IsCustomerServiceAgentRole(raw) {
			t.Fatalf("expected %q to be a customer service agent role", raw)
		}
	}

	for _, raw := range []string{"agent", "administrator", "shop_manager", "customer_service", "editor", "viewer", "subscriber", "customer"} {
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

func TestMerchantPermissionsAreSeparateFromProductPermissions(t *testing.T) {
	if PermMerchantView == Permission("") || PermMerchantEdit == Permission("") || PermMerchantSync == Permission("") {
		t.Fatal("expected merchant permissions to be explicit")
	}
	if PermMerchantView == PermProductView || PermMerchantEdit == PermProductEdit {
		t.Fatal("merchant permissions must not alias product permissions")
	}
	if !RoleAdmin.HasPermission(PermMerchantView) || !RoleAdmin.HasPermission(PermMerchantEdit) || !RoleAdmin.HasPermission(PermMerchantSync) {
		t.Fatal("admin should have full merchant permissions")
	}
	if !RoleEditor.HasPermission(PermMerchantSync) {
		t.Fatal("editor should keep day-to-day merchant sync access")
	}
	if RoleViewer.HasPermission(PermMerchantView) || RoleSupport.HasPermission(PermMerchantView) {
		t.Fatal("viewer and support roles should not enter merchant operations by default")
	}
}
