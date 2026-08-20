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
		"test_user":                  RoleTestUser,
		"TEST_USER":                  RoleTestUser,
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

func TestTestUserIsStorefrontOnly(t *testing.T) {
	if !RoleTestUser.IsValid() {
		t.Fatal("expected test user role to be valid")
	}
	if !IsStorefrontRole(RoleTestUser.String()) {
		t.Fatal("expected test user role to be a storefront role")
	}
	if IsBackofficeRole(RoleTestUser.String()) {
		t.Fatal("test user role must not be a backoffice role")
	}
	if len(RoleTestUser.GetPermissions()) != 0 {
		t.Fatal("test user role must not receive backoffice permissions")
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

func TestReviewModerationPermissionsAreSeparated(t *testing.T) {
	if !RoleAdmin.HasPermission(PermReviewView) || !RoleAdmin.HasPermission(PermReviewModerate) {
		t.Fatal("admin should have review view and moderation permissions")
	}
	if !RoleManager.HasPermission(PermReviewView) || !RoleManager.HasPermission(PermReviewModerate) {
		t.Fatal("manager should have review view and moderation permissions")
	}
	if !RoleEditor.HasPermission(PermReviewView) || !RoleEditor.HasPermission(PermReviewModerate) {
		t.Fatal("editor should have review view and moderation permissions")
	}
	if !RoleViewer.HasPermission(PermReviewView) || RoleViewer.HasPermission(PermReviewModerate) {
		t.Fatal("viewer should only have review view permission")
	}
}
