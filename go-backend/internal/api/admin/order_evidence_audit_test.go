package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderEvidenceAuditRecorder struct {
	logs []audit.AuditLog
}

func (r *orderEvidenceAuditRecorder) CreateAuditLog(log *audit.AuditLog) error {
	if log != nil {
		r.logs = append(r.logs, *log)
	}
	return nil
}

func TestOrderEvidenceHandlerAuditOmitsRawEvidenceDataAndStorageKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, orderID, itemID, _ := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService, attachmentService := newOrderEvidenceServicesForHandlerTest(
		db,
		evidenceRepo,
		newAdminOrderEvidenceHandlerStorage(t),
	)
	auditRecorder := &orderEvidenceAuditRecorder{}
	handler := NewOrderEvidenceHandler(adminService)
	handler.ConfigureAttachmentService(attachmentService)
	handler.ConfigureAuditService(auditRecorder)

	updateRecorder := httptest.NewRecorder()
	updateContext, _ := gin.CreateTestContext(updateRecorder)
	updateContext.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/admin/orders/"+formatUint(orderID)+"/evidence/items/"+formatUint(itemID),
		strings.NewReader(`{"status":"draft","data_json":{"private_note":"must-not-be-audit-data"}}`),
	)
	updateContext.Request.Header.Set("Content-Type", "application/json")
	updateContext.Params = gin.Params{
		{Key: "id", Value: formatUint(orderID)},
		{Key: "item_id", Value: formatUint(itemID)},
	}
	updateContext.Set("user_id", uint(7))
	updateContext.Set("username", "evidence-admin")

	handler.UpdateItem(updateContext)

	require.Equal(t, http.StatusOK, updateRecorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	updateLog := auditRecorder.logs[0]
	assert.Equal(t, "order_outbound_evidence", updateLog.Resource)
	assert.Equal(t, "update", updateLog.Action)
	assert.Contains(t, updateLog.Changes, `"data_json_present":true`)
	assert.NotContains(t, updateLog.Changes, "private_note")
	assert.NotContains(t, updateLog.OldValue, "private_note")
	assert.NotContains(t, updateLog.NewValue, "private_note")
	assert.NotContains(t, updateLog.NewValue, `"data_json"`)

	uploadRequest := newAdminEvidenceMultipartRequest(
		t,
		http.MethodPost,
		"/api/admin/orders/"+formatUint(orderID)+"/evidence/items/"+formatUint(itemID)+"/attachments",
		"manual-evidence.png",
	)
	uploadRecorder := httptest.NewRecorder()
	uploadContext, _ := gin.CreateTestContext(uploadRecorder)
	uploadContext.Request = uploadRequest
	uploadContext.Params = gin.Params{
		{Key: "id", Value: formatUint(orderID)},
		{Key: "item_id", Value: formatUint(itemID)},
	}
	uploadContext.Set("user_id", uint(7))
	uploadContext.Set("username", "evidence-admin")

	handler.UploadAttachment(uploadContext)

	require.Equal(t, http.StatusCreated, uploadRecorder.Code)
	require.Len(t, auditRecorder.logs, 2)
	uploadLog := auditRecorder.logs[1]
	assert.Equal(t, "create", uploadLog.Action)
	assert.Contains(t, uploadLog.NewValue, `"mime_type":"image/png"`)
	assert.NotContains(t, uploadLog.NewValue, "order-evidence/")
	assert.NotContains(t, uploadLog.NewValue, "manual-evidence.png")
}
