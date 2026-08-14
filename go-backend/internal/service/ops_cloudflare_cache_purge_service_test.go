package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type fakeOpsCloudflareCachePurgeClient struct {
	readCalls  []string
	writeCalls []fakeOpsCloudflareWriteCall
}

type fakeOpsCloudflareWriteCall struct {
	Method string
	Path   string
	Body   map[string][]string
}

func (f *fakeOpsCloudflareCachePurgeClient) CloudflareRead(
	_ context.Context,
	_ uint,
	path string,
	_ url.Values,
	target interface{},
) (int, error) {
	f.readCalls = append(f.readCalls, path)
	if path != "/zones" {
		return http.StatusNotFound, fmt.Errorf("unexpected read path %s", path)
	}
	return http.StatusOK, json.Unmarshal(
		[]byte(`{"result":[{"id":"zone-1","name":"example.com"},{"id":"zone-2","name":"other.example"}]}`),
		target,
	)
}

func (f *fakeOpsCloudflareCachePurgeClient) CloudflareWrite(
	_ context.Context,
	_ uint,
	method string,
	path string,
	body []byte,
	target interface{},
) (int, error) {
	var payload map[string][]string
	if err := json.Unmarshal(body, &payload); err != nil {
		return http.StatusBadRequest, err
	}
	f.writeCalls = append(f.writeCalls, fakeOpsCloudflareWriteCall{
		Method: method,
		Path:   path,
		Body:   payload,
	})
	if target != nil {
		_ = json.Unmarshal([]byte(`{"id":"purge-operation"}`), target)
	}
	return http.StatusOK, nil
}

func TestOpsCloudflareCachePurgeGroupsAndBatchesHosts(t *testing.T) {
	db, projectID := newOpsCloudflareCachePurgeTestDB(t)
	domainRepo := repository.NewOpsDomainBindingRepository(db)
	client := &fakeOpsCloudflareCachePurgeClient{}
	cachePurgeService := &OpsCloudflareCachePurgeService{
		domainRepo:       domainRepo,
		connectorService: client,
	}

	for index := 0; index < 31; index++ {
		require.NoError(t, db.Create(&ops.DomainBinding{
			Domain:           fmt.Sprintf("app-%02d.example.com", index),
			ConnectorID:      opsCachePurgeUintPointer(projectID - 40),
			ProjectBindingID: opsCachePurgeUintPointer(projectID),
			Role:             ops.DomainRoleAlias,
			Environment:      ops.DomainEnvironmentProduction,
			Provider:         ops.DomainProviderCloudflare,
			Zone:             "example.com",
			Enabled:          true,
		}).Error)
	}
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:           "admin.other.example",
		ConnectorID:      opsCachePurgeUintPointer(projectID - 39),
		ProjectBindingID: opsCachePurgeUintPointer(projectID),
		Role:             ops.DomainRoleAdmin,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Zone:             "other.example",
		Enabled:          true,
	}).Error)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:           "disabled.example.com",
		ConnectorID:      opsCachePurgeUintPointer(projectID - 40),
		ProjectBindingID: opsCachePurgeUintPointer(projectID),
		Role:             ops.DomainRoleAlias,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Zone:             "example.com",
		Enabled:          false,
	}).Error)
	require.NoError(t, db.Model(&ops.DomainBinding{}).
		Where("domain = ?", "disabled.example.com").
		Update("enabled", false).Error)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:           "origin.example.com",
		ProjectBindingID: opsCachePurgeUintPointer(projectID),
		Role:             ops.DomainRoleInternal,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderOther,
		Enabled:          true,
	}).Error)

	result, err := cachePurgeService.PurgeProject(context.Background(), projectID)
	require.NoError(t, err)
	require.False(t, result.Skipped)
	require.Equal(t, 32, result.DomainCount)
	require.Equal(t, 32, result.HostCount)
	require.Equal(t, 2, result.ZoneCount)
	require.Equal(t, 3, result.RequestCount)
	require.Len(t, result.Groups, 2)
	require.Len(t, client.readCalls, 2)
	require.Len(t, client.writeCalls, 3)
	for _, call := range client.writeCalls {
		require.Equal(t, http.MethodPost, call.Method)
		require.Contains(t, call.Path, "/purge_cache")
		require.LessOrEqual(t, len(call.Body["hosts"]), opsCloudflareCachePurgeBatchSize)
	}
}

func TestOpsCloudflareCachePurgeSkipsProjectsWithoutCloudflareDomains(t *testing.T) {
	db, projectID := newOpsCloudflareCachePurgeTestDB(t)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:           "internal.example",
		ProjectBindingID: opsCachePurgeUintPointer(projectID),
		Role:             ops.DomainRoleInternal,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderOther,
		Enabled:          true,
	}).Error)
	client := &fakeOpsCloudflareCachePurgeClient{}
	cachePurgeService := &OpsCloudflareCachePurgeService{
		domainRepo:       repository.NewOpsDomainBindingRepository(db),
		connectorService: client,
	}

	result, err := cachePurgeService.PurgeProject(context.Background(), projectID)
	require.NoError(t, err)
	require.True(t, result.Skipped)
	require.Equal(t, 0, result.HostCount)
	require.Empty(t, client.readCalls)
	require.Empty(t, client.writeCalls)
}

func TestOpsCloudflareCachePurgeRejectsIncompleteCloudflareBinding(t *testing.T) {
	db, projectID := newOpsCloudflareCachePurgeTestDB(t)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:           "www.example.com",
		ProjectBindingID: opsCachePurgeUintPointer(projectID),
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Enabled:          true,
	}).Error)
	cachePurgeService := &OpsCloudflareCachePurgeService{
		domainRepo:       repository.NewOpsDomainBindingRepository(db),
		connectorService: &fakeOpsCloudflareCachePurgeClient{},
	}

	_, err := cachePurgeService.PurgeProject(context.Background(), projectID)
	require.ErrorIs(t, err, ErrOpsCloudflareCachePurge)
}

func newOpsCloudflareCachePurgeTestDB(t *testing.T) (*gorm.DB, uint) {
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
	return db, 41
}

func opsCachePurgeUintPointer(value uint) *uint {
	return &value
}
