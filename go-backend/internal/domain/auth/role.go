package auth

import "strings"

// Role represents an admin/user role.
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleEditor  Role = "editor"
	RoleSupport Role = "support"
	RoleViewer  Role = "viewer"
	RoleUser    Role = "user"
)

func NormalizeRole(raw string) Role {
	switch Role(strings.ToLower(strings.TrimSpace(raw))) {
	case RoleAdmin:
		return RoleAdmin
	case RoleManager:
		return RoleManager
	case RoleSupport:
		return RoleSupport
	case RoleEditor:
		return RoleEditor
	case RoleViewer:
		return RoleViewer
	default:
		return RoleUser
	}
}

func IsCustomerServiceAgentRole(raw string) bool {
	role := NormalizeRole(raw)
	return role == RoleAdmin || role == RoleManager || role == RoleSupport
}

// Permission is a backend/admin permission code.
type Permission string

const (
	PermProductView   Permission = "product:view"
	PermProductCreate Permission = "product:create"
	PermProductEdit   Permission = "product:edit"
	PermProductDelete Permission = "product:delete"

	PermOrderView   Permission = "order:view"
	PermOrderEdit   Permission = "order:edit"
	PermOrderRefund Permission = "order:refund"
	PermOrderDelete Permission = "order:delete"

	PermUserView   Permission = "user:view"
	PermUserCreate Permission = "user:create"
	PermUserEdit   Permission = "user:edit"
	PermUserDelete Permission = "user:delete"

	PermContentView   Permission = "content:view"
	PermContentCreate Permission = "content:create"
	PermContentEdit   Permission = "content:edit"
	PermContentDelete Permission = "content:delete"

	PermFAQView   Permission = "faq:view"
	PermFAQCreate Permission = "faq:create"
	PermFAQEdit   Permission = "faq:edit"
	PermFAQDelete Permission = "faq:delete"

	PermGalleryView   Permission = "gallery:view"
	PermGalleryCreate Permission = "gallery:create"
	PermGalleryEdit   Permission = "gallery:edit"
	PermGalleryDelete Permission = "gallery:delete"

	PermSubscriptionView   Permission = "subscription:view"
	PermSubscriptionEdit   Permission = "subscription:edit"
	PermSubscriptionDelete Permission = "subscription:delete"
	PermSubscriptionExport Permission = "subscription:export"

	PermTicketView   Permission = "ticket:view"
	PermTicketCreate Permission = "ticket:create"
	PermTicketEdit   Permission = "ticket:edit"
	PermTicketAssign Permission = "ticket:assign"
	PermTicketClose  Permission = "ticket:close"
	PermTicketDelete Permission = "ticket:delete"

	PermMarketingView   Permission = "marketing:view"
	PermMarketingCreate Permission = "marketing:create"
	PermMarketingEdit   Permission = "marketing:edit"
	PermMarketingDelete Permission = "marketing:delete"

	PermShippingView     Permission = "shipping:view"
	PermShippingCreate   Permission = "shipping:create"
	PermShippingEdit     Permission = "shipping:edit"
	PermShippingDelete   Permission = "shipping:delete"
	PermShippingTracking Permission = "shipping:tracking"

	PermSettingsView Permission = "settings:view"
	PermSettingsEdit Permission = "settings:edit"

	PermLogsView Permission = "logs:view"

	PermSystemManage Permission = "system:manage"
)

// RolePermissions maps each role to the admin permissions it receives.
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermProductView, PermProductCreate, PermProductEdit, PermProductDelete,
		PermOrderView, PermOrderEdit, PermOrderRefund, PermOrderDelete,
		PermUserView, PermUserCreate, PermUserEdit, PermUserDelete,
		PermContentView, PermContentCreate, PermContentEdit, PermContentDelete,
		PermFAQView, PermFAQCreate, PermFAQEdit, PermFAQDelete,
		PermGalleryView, PermGalleryCreate, PermGalleryEdit, PermGalleryDelete,
		PermSubscriptionView, PermSubscriptionEdit, PermSubscriptionDelete, PermSubscriptionExport,
		PermTicketView, PermTicketCreate, PermTicketEdit, PermTicketAssign, PermTicketClose, PermTicketDelete,
		PermMarketingView, PermMarketingCreate, PermMarketingEdit, PermMarketingDelete,
		PermShippingView, PermShippingCreate, PermShippingEdit, PermShippingDelete, PermShippingTracking,
		PermSettingsView, PermSettingsEdit,
		PermLogsView,
		PermSystemManage,
	},
	RoleManager: {
		PermProductView, PermProductCreate, PermProductEdit, PermProductDelete,
		PermOrderView, PermOrderEdit, PermOrderRefund,
		PermUserView, PermUserCreate, PermUserEdit,
		PermContentView, PermContentCreate, PermContentEdit, PermContentDelete,
		PermFAQView, PermFAQCreate, PermFAQEdit, PermFAQDelete,
		PermGalleryView, PermGalleryCreate, PermGalleryEdit, PermGalleryDelete,
		PermSubscriptionView, PermSubscriptionEdit, PermSubscriptionExport,
		PermTicketView, PermTicketCreate, PermTicketEdit, PermTicketAssign, PermTicketClose,
		PermMarketingView, PermMarketingCreate, PermMarketingEdit, PermMarketingDelete,
		PermShippingView, PermShippingCreate, PermShippingEdit, PermShippingTracking,
		PermSettingsView,
		PermLogsView,
	},
	RoleEditor: {
		PermProductView, PermProductCreate, PermProductEdit,
		PermShippingView,
		PermContentView, PermContentCreate, PermContentEdit,
		PermFAQView, PermFAQCreate, PermFAQEdit,
		PermGalleryView, PermGalleryCreate, PermGalleryEdit,
	},
	RoleSupport: {
		PermOrderView,
		PermUserView,
		PermTicketView, PermTicketCreate, PermTicketEdit, PermTicketClose,
		PermSubscriptionView,
		PermShippingView, PermShippingTracking,
	},
	RoleViewer: {
		PermProductView,
		PermOrderView,
		PermUserView,
		PermContentView,
		PermFAQView,
		PermGalleryView,
		PermSubscriptionView,
		PermTicketView,
		PermMarketingView,
		PermShippingView,
		PermSettingsView,
	},
	RoleUser: {},
}

func (r Role) HasPermission(perm Permission) bool {
	permissions, exists := RolePermissions[r]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}

func (r Role) GetPermissions() []Permission {
	return RolePermissions[r]
}

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleManager, RoleEditor, RoleSupport, RoleViewer, RoleUser:
		return true
	default:
		return false
	}
}

func (r Role) String() string {
	return string(r)
}
