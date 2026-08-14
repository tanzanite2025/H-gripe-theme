package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsDeploymentHealthCheckServiceChecksDNSHTTPAndHTTPS(t *testing.T) {
	db := newDeploymentHealthTestDB(t)
	projectID := uint(21)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:           "shop.example.com",
		ProjectBindingID: &projectID,
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Enabled:          true,
	}).Error)

	service := NewOpsDeploymentHealthCheckService(repository.NewOpsDomainBindingRepository(db))
	service.lookupHost = func(context.Context, string) ([]string, error) {
		return []string{"203.0.113.20"}, nil
	}
	service.httpClient = &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
				Request:    request,
			}, nil
		}),
	}

	report, err := service.CheckProject(context.Background(), &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ID:          projectID,
			Name:        "commerce-platform",
			Environment: ops.ProjectEnvironmentProduction,
		},
	})
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentHealthHealthy, report.Status)
	require.Len(t, report.Checks, 3)
	require.Contains(t, report.Summary, "通过")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func newDeploymentHealthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&ops.DomainBinding{}))
	return db
}
