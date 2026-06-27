package auth

import "strings"

// Role 用户角色
type Role string

const (
	RoleAdmin   Role = "admin"   // 超级管理员 - 所有权限
	RoleManager Role = "manager" // 经理 - 大部分权限
	RoleEditor  Role = "editor"  // 编辑 - 内容管理
	RoleSupport Role = "support" // 客服 - 工单和客户
	RoleViewer  Role = "viewer"  // 查看者 - 只读
	RoleUser    Role = "user"    // 普通用户
)

func NormalizeRole(raw string) Role {
	roles := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(raw)), func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	if len(roles) == 0 {
		return RoleUser
	}

	priority := []struct {
		role    Role
		aliases map[string]bool
	}{
		{RoleAdmin, map[string]bool{"admin": true, "administrator": true}},
		{RoleManager, map[string]bool{"manager": true, "shop_manager": true}},
		{RoleSupport, map[string]bool{"support": true, "agent": true, "customer_service": true, "customer_support": true}},
		{RoleEditor, map[string]bool{"editor": true}},
		{RoleViewer, map[string]bool{"viewer": true}},
	}

	for _, item := range priority {
		for _, role := range roles {
			if item.aliases[role] {
				return item.role
			}
		}
	}

	return RoleUser
}

func IsCustomerServiceAgentRole(raw string) bool {
	role := NormalizeRole(raw)
	return role == RoleAdmin || role == RoleManager || role == RoleSupport
}

// Permission 权限
type Permission string

const (
	// 商品权限
	PermProductView   Permission = "product:view"
	PermProductCreate Permission = "product:create"
	PermProductEdit   Permission = "product:edit"
	PermProductDelete Permission = "product:delete"

	// 订单权限
	PermOrderView   Permission = "order:view"
	PermOrderEdit   Permission = "order:edit"
	PermOrderRefund Permission = "order:refund"
	PermOrderDelete Permission = "order:delete"

	// 用户权限
	PermUserView   Permission = "user:view"
	PermUserCreate Permission = "user:create"
	PermUserEdit   Permission = "user:edit"
	PermUserDelete Permission = "user:delete"

	// 内容权限
	PermContentView   Permission = "content:view"
	PermContentCreate Permission = "content:create"
	PermContentEdit   Permission = "content:edit"
	PermContentDelete Permission = "content:delete"

	// FAQ 权限
	PermFAQView   Permission = "faq:view"
	PermFAQCreate Permission = "faq:create"
	PermFAQEdit   Permission = "faq:edit"
	PermFAQDelete Permission = "faq:delete"

	// 图库权限
	PermGalleryView   Permission = "gallery:view"
	PermGalleryCreate Permission = "gallery:create"
	PermGalleryEdit   Permission = "gallery:edit"
	PermGalleryDelete Permission = "gallery:delete"

	// 订阅权限
	PermSubscriptionView   Permission = "subscription:view"
	PermSubscriptionEdit   Permission = "subscription:edit"
	PermSubscriptionDelete Permission = "subscription:delete"
	PermSubscriptionExport Permission = "subscription:export"

	// 工单权限
	PermTicketView   Permission = "ticket:view"
	PermTicketCreate Permission = "ticket:create"
	PermTicketEdit   Permission = "ticket:edit"
	PermTicketAssign Permission = "ticket:assign"
	PermTicketClose  Permission = "ticket:close"
	PermTicketDelete Permission = "ticket:delete"

	// 营销权限
	PermMarketingView   Permission = "marketing:view"
	PermMarketingCreate Permission = "marketing:create"
	PermMarketingEdit   Permission = "marketing:edit"
	PermMarketingDelete Permission = "marketing:delete"

	// 设置权限
	PermSettingsView Permission = "settings:view"
	PermSettingsEdit Permission = "settings:edit"

	// 审计日志权限
	PermLogsView Permission = "logs:view"

	// 系统权限
	PermSystemManage Permission = "system:manage"
)

// RolePermissions 角色权限映射
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// 超级管理员拥有所有权限
		PermProductView, PermProductCreate, PermProductEdit, PermProductDelete,
		PermOrderView, PermOrderEdit, PermOrderRefund, PermOrderDelete,
		PermUserView, PermUserCreate, PermUserEdit, PermUserDelete,
		PermContentView, PermContentCreate, PermContentEdit, PermContentDelete,
		PermFAQView, PermFAQCreate, PermFAQEdit, PermFAQDelete,
		PermGalleryView, PermGalleryCreate, PermGalleryEdit, PermGalleryDelete,
		PermSubscriptionView, PermSubscriptionEdit, PermSubscriptionDelete, PermSubscriptionExport,
		PermTicketView, PermTicketCreate, PermTicketEdit, PermTicketAssign, PermTicketClose, PermTicketDelete,
		PermMarketingView, PermMarketingCreate, PermMarketingEdit, PermMarketingDelete,
		PermSettingsView, PermSettingsEdit,
		PermLogsView,
		PermSystemManage,
	},
	RoleManager: {
		// 经理拥有大部分权限，除了系统管理
		PermProductView, PermProductCreate, PermProductEdit, PermProductDelete,
		PermOrderView, PermOrderEdit, PermOrderRefund,
		PermUserView, PermUserCreate, PermUserEdit,
		PermContentView, PermContentCreate, PermContentEdit, PermContentDelete,
		PermFAQView, PermFAQCreate, PermFAQEdit, PermFAQDelete,
		PermGalleryView, PermGalleryCreate, PermGalleryEdit, PermGalleryDelete,
		PermSubscriptionView, PermSubscriptionEdit, PermSubscriptionExport,
		PermTicketView, PermTicketCreate, PermTicketEdit, PermTicketAssign, PermTicketClose,
		PermMarketingView, PermMarketingCreate, PermMarketingEdit, PermMarketingDelete,
		PermSettingsView,
		PermLogsView,
	},
	RoleEditor: {
		// 编辑只有内容相关权限
		PermProductView, PermProductCreate, PermProductEdit,
		PermContentView, PermContentCreate, PermContentEdit,
		PermFAQView, PermFAQCreate, PermFAQEdit,
		PermGalleryView, PermGalleryCreate, PermGalleryEdit,
	},
	RoleSupport: {
		// 客服只有工单和客户相关权限
		PermOrderView,
		PermUserView,
		PermTicketView, PermTicketCreate, PermTicketEdit, PermTicketClose,
		PermSubscriptionView,
	},
	RoleViewer: {
		// 查看者只有查看权限
		PermProductView,
		PermOrderView,
		PermUserView,
		PermContentView,
		PermFAQView,
		PermGalleryView,
		PermSubscriptionView,
		PermTicketView,
		PermMarketingView,
		PermSettingsView,
	},
}

// HasPermission 检查角色是否拥有指定权限
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

// GetPermissions 获取角色的所有权限
func (r Role) GetPermissions() []Permission {
	return RolePermissions[r]
}

// IsValid 检查角色是否有效
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleManager, RoleEditor, RoleSupport, RoleViewer, RoleUser:
		return true
	default:
		return false
	}
}

// String 返回角色的字符串表示
func (r Role) String() string {
	return string(r)
}
