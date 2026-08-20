package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsNetworkSummaryHandlerFiltersByEnvironment(t *testing.T) {
	handler := newOpsNetworkSummaryTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/network/summary?environment=staging", nil)

	handler.Get(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Environment string `json:"environment"`
			Summary     struct {
				ExplicitRuleCount int `json:"explicit_rule_count"`
			} `json:"summary"`
			Items []struct {
				Name string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, ops.DomainEnvironmentStaging, body.Data.Environment)
	require.Equal(t, 1, body.Data.Summary.ExplicitRuleCount)
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, "Staging ingress", body.Data.Items[0].Name)
}

func TestOpsNetworkSummaryHandlerRejectsInvalidEnvironment(t *testing.T) {
	handler := newOpsNetworkSummaryTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/network/summary?environment=qa", nil)

	handler.Get(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func newOpsNetworkSummaryTestHandler(t *testing.T) *OpsNetworkSummaryHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&ops.Connector{},
		&ops.VPSBinding{},
		&ops.ProjectBinding{},
		&ops.DomainBinding{},
		&ops.NetworkRule{},
	))
	require.NoError(t, db.Create(&ops.NetworkRule{
		Name:           "Production ingress",
		Environment:    ops.DomainEnvironmentProduction,
		OwnerKind:      ops.NetworkOwnerManual,
		ManagedBy:      ops.NetworkManagedByManual,
		Scope:          ops.NetworkScopeGateway,
		Direction:      ops.NetworkDirectionIngress,
		Protocol:       ops.NetworkProtocolTCP,
		DesiredState:   ops.NetworkStateOpen,
		ObservedState:  ops.NetworkStateUnknown,
		EffectiveState: ops.NetworkStateUnknown,
		Status:         ops.NetworkStatusPending,
		Enabled:        true,
	}).Error)
	require.NoError(t, db.Create(&ops.NetworkRule{
		Name:           "Staging ingress",
		Environment:    ops.DomainEnvironmentStaging,
		OwnerKind:      ops.NetworkOwnerManual,
		ManagedBy:      ops.NetworkManagedByManual,
		Scope:          ops.NetworkScopeGateway,
		Direction:      ops.NetworkDirectionIngress,
		Protocol:       ops.NetworkProtocolTCP,
		DesiredState:   ops.NetworkStateOpen,
		ObservedState:  ops.NetworkStateUnknown,
		EffectiveState: ops.NetworkStateUnknown,
		Status:         ops.NetworkStatusPending,
		Enabled:        true,
	}).Error)

	return NewOpsNetworkSummaryHandler(service.NewOpsNetworkSummaryService(
		repository.NewOpsNetworkRuleRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		repository.NewOpsDomainBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
	))
}
