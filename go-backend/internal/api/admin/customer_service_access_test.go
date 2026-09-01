package admin

import (
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdminCustomerServiceAccessUsesCanonicalRoleClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		userID      uint
		role        string
		wantViewAll bool
		wantEdit    bool
	}{
		{name: "support can edit assigned conversations", userID: 12, role: "support", wantEdit: true},
		{name: "manager can view all and edit", userID: 13, role: "manager", wantViewAll: true, wantEdit: true},
		{name: "editor cannot access ticket edits", userID: 14, role: "editor"},
		{name: "viewer cannot access ticket edits", userID: 15, role: "viewer"},
		{name: "missing role is denied", userID: 16},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(nil)
			context.Set("user_id", test.userID)
			if test.role != "" {
				context.Set("role", test.role)
			}
			// A conflicting compatibility key must not alter the access decision.
			context.Set("user_role", "admin")

			userID, canViewAll := adminCustomerServiceScope(context)
			assert.Equal(t, test.userID, userID)
			assert.Equal(t, test.wantViewAll, canViewAll)
			assert.Equal(t, test.wantEdit, adminCustomerServiceCanEdit(context))
		})
	}
}

func TestAdminCustomerServiceAccessDeniesNilContext(t *testing.T) {
	userID, canViewAll := adminCustomerServiceScope(nil)
	assert.Zero(t, userID)
	assert.False(t, canViewAll)
	assert.False(t, adminCustomerServiceCanEdit(nil))
}

func TestScopeAdminCustomerServiceConversationFiltersKeepsAdminInboxGlobal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		wantAssignedTo *uint
	}{
		{name: "admin sees all conversations", role: "admin"},
		{name: "manager sees all conversations", role: "manager"},
		{name: "support sees assigned conversations", role: "support", wantAssignedTo: uintPtr(12)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(nil)
			context.Set("user_id", uint(12))
			context.Set("role", test.role)

			filters, agentUserID, canViewAll := scopeAdminCustomerServiceConversationFilters(
				context,
				service.CustomerServiceConversationListInput{},
			)

			assert.Equal(t, uint(12), agentUserID)
			assert.Equal(t, test.role == "admin" || test.role == "manager", canViewAll)
			if test.wantAssignedTo == nil {
				assert.Nil(t, filters.AssignedTo)
			} else if assert.NotNil(t, filters.AssignedTo) {
				assert.Equal(t, *test.wantAssignedTo, *filters.AssignedTo)
			}
		})
	}
}

func uintPtr(value uint) *uint {
	return &value
}
