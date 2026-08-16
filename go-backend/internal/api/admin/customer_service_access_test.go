package admin

import (
	"testing"

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
