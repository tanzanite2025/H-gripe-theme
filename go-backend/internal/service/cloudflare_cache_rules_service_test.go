package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeCloudflareCacheRulesClient struct {
	ruleEnabled bool
	writeCalls  []fakeCloudflareCacheRulesWriteCall
}

type fakeCloudflareCacheRulesWriteCall struct {
	Method string
	Path   string
	Body   map[string]bool
}

func (f *fakeCloudflareCacheRulesClient) CloudflareRead(
	_ context.Context,
	_ uint,
	path string,
	_ url.Values,
	target interface{},
) (int, error) {
	switch {
	case path == "/zones":
		assignCloudflareCacheRulesJSON(target, map[string]interface{}{
			"result": []map[string]string{{"id": "zone-1", "name": "example.com"}},
		})
	case strings.Contains(path, "/rulesets/phases/"+cloudflareCacheRulesPhase+"/entrypoint"):
		assignCloudflareCacheRulesJSON(target, map[string]interface{}{
			"id":    "ruleset-1",
			"name":  "default",
			"phase": cloudflareCacheRulesPhase,
			"rules": []map[string]interface{}{
				{
					"id":          "rule-1",
					"action":      "set_cache_settings",
					"description": "Static assets",
					"expression":  `starts_with(http.request.uri.path, "/uploads/")`,
					"enabled":     f.ruleEnabled,
					"action_parameters": map[string]interface{}{
						"edge_ttl": map[string]interface{}{"mode": "override_origin", "default": 31536000},
					},
				},
			},
		})
	default:
		return http.StatusNotFound, nil
	}
	return http.StatusOK, nil
}

func (f *fakeCloudflareCacheRulesClient) CloudflareWrite(
	_ context.Context,
	_ uint,
	method string,
	path string,
	body []byte,
	target interface{},
) (int, error) {
	payload := map[string]bool{}
	if err := json.Unmarshal(body, &payload); err != nil {
		panic(err)
	}
	f.writeCalls = append(f.writeCalls, fakeCloudflareCacheRulesWriteCall{
		Method: method,
		Path:   path,
		Body:   payload,
	})
	f.ruleEnabled = payload["enabled"]
	assignCloudflareCacheRulesJSON(target, map[string]interface{}{
		"id":      "rule-1",
		"enabled": f.ruleEnabled,
	})
	return http.StatusOK, nil
}

func TestCloudflareCacheRulesServiceReadsOriginOverrides(t *testing.T) {
	client := &fakeCloudflareCacheRulesClient{ruleEnabled: true}
	service := &CloudflareCacheRulesService{connectorService: client}

	result, err := service.Get(context.Background(), 4, "example.com")

	require.NoError(t, err)
	require.True(t, result.RulesetConfigured)
	require.Equal(t, "zone-1", result.ZoneID)
	require.Equal(t, "overridden", result.OriginCacheControlStatus)
	require.Len(t, result.Rules, 1)
	require.Equal(t, "overridden", result.Rules[0].OriginCacheControlStatus)
	require.Equal(t, "override_origin", result.Rules[0].EdgeTTLMode)
}

func TestCloudflareCacheRulesServiceUpdatesRuleEnabledState(t *testing.T) {
	client := &fakeCloudflareCacheRulesClient{ruleEnabled: true}
	service := &CloudflareCacheRulesService{connectorService: client}

	result, err := service.SetRuleEnabled(context.Background(), 4, "example.com", "rule-1", false)

	require.NoError(t, err)
	require.Len(t, client.writeCalls, 1)
	require.Equal(t, http.MethodPatch, client.writeCalls[0].Method)
	require.Equal(t, "/zones/zone-1/rulesets/ruleset-1/rules/rule-1", client.writeCalls[0].Path)
	require.Equal(t, false, client.writeCalls[0].Body["enabled"])
	require.False(t, result.Rules[0].Enabled)
	require.Equal(t, "no_rules", result.OriginCacheControlStatus)
}

func TestCloudflareWriteRulesetPathOnlyAllowsRulePatch(t *testing.T) {
	path := "/zones/zone-1/rulesets/ruleset-1/rules/rule-1"

	require.True(t, isAllowedCloudflareWritePath(http.MethodPatch, path))
	require.False(t, isAllowedCloudflareWritePath(http.MethodPost, path))
	require.False(t, isAllowedCloudflareWritePath(http.MethodDelete, path))
}

func assignCloudflareCacheRulesJSON(target interface{}, value interface{}) {
	if target == nil {
		return
	}
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		panic(err)
	}
}
