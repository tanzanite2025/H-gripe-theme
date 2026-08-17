package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const cloudflareCacheRulesPhase = "http_request_cache_settings"

var (
	ErrCloudflareCacheRules        = errors.New("Cloudflare cache rules request failed")
	ErrInvalidCloudflareCacheRule  = errors.New("invalid Cloudflare cache rule request")
	ErrCloudflareCacheRuleNotFound = errors.New("Cloudflare cache rule was not found")
)

type cloudflareCacheRulesClient interface {
	CloudflareRead(context.Context, uint, string, url.Values, interface{}) (int, error)
	CloudflareWrite(context.Context, uint, string, string, []byte, interface{}) (int, error)
}

type CloudflareCacheRulesService struct {
	connectorService cloudflareCacheRulesClient
}

type CloudflareCacheRule struct {
	ID                       string `json:"id"`
	Version                  string `json:"version,omitempty"`
	Action                   string `json:"action"`
	Description              string `json:"description,omitempty"`
	Expression               string `json:"expression"`
	Enabled                  bool   `json:"enabled"`
	LastUpdated              string `json:"last_updated,omitempty"`
	EdgeTTLMode              string `json:"edge_ttl_mode,omitempty"`
	BrowserTTL               string `json:"browser_ttl,omitempty"`
	OriginCacheControlStatus string `json:"origin_cache_control_status"`
}

type CloudflareCacheRulesResult struct {
	ConnectorID              uint                  `json:"connector_id"`
	Zone                     string                `json:"zone"`
	ZoneID                   string                `json:"zone_id"`
	RulesetID                string                `json:"ruleset_id,omitempty"`
	RulesetName              string                `json:"ruleset_name,omitempty"`
	RulesetConfigured        bool                  `json:"ruleset_configured"`
	OriginCacheControlStatus string                `json:"origin_cache_control_status"`
	Rules                    []CloudflareCacheRule `json:"rules"`
}

type cloudflareCacheRulesetResponse struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Phase       string                      `json:"phase"`
	Rules       []cloudflareCacheRuleRemote `json:"rules"`
}

type cloudflareCacheRuleRemote struct {
	ID               string                 `json:"id"`
	Version          string                 `json:"version"`
	Action           string                 `json:"action"`
	ActionParameters map[string]interface{} `json:"action_parameters"`
	Description      string                 `json:"description"`
	Expression       string                 `json:"expression"`
	Enabled          bool                   `json:"enabled"`
	LastUpdated      string                 `json:"last_updated"`
}

func NewCloudflareCacheRulesService(connectorService *OpsConnectorService) *CloudflareCacheRulesService {
	return &CloudflareCacheRulesService{connectorService: connectorService}
}

func (s *CloudflareCacheRulesService) Get(
	ctx context.Context,
	connectorID uint,
	zone string,
) (*CloudflareCacheRulesResult, error) {
	if s == nil || s.connectorService == nil {
		return nil, errors.New("Cloudflare cache rules service is not configured")
	}
	if connectorID == 0 {
		return nil, fmt.Errorf("%w: connector id is required", ErrInvalidCloudflareCacheRule)
	}
	zone, err := normalizeCloudflareCacheRuleZone(zone)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	zoneID, err := s.findZoneID(ctx, connectorID, zone)
	if err != nil {
		return nil, err
	}
	result := &CloudflareCacheRulesResult{
		ConnectorID:              connectorID,
		Zone:                     zone,
		ZoneID:                   zoneID,
		Rules:                    []CloudflareCacheRule{},
		OriginCacheControlStatus: "no_rules",
	}

	var ruleset cloudflareCacheRulesetResponse
	statusCode, err := s.connectorService.CloudflareRead(
		ctx,
		connectorID,
		"/zones/"+url.PathEscape(zoneID)+"/rulesets/phases/"+cloudflareCacheRulesPhase+"/entrypoint",
		nil,
		&ruleset,
	)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return result, nil
		}
		return nil, fmt.Errorf("%w: read cache rules for %s: %v", ErrCloudflareCacheRules, zone, err)
	}
	if strings.TrimSpace(ruleset.ID) == "" {
		return result, nil
	}

	result.RulesetConfigured = true
	result.RulesetID = strings.TrimSpace(ruleset.ID)
	result.RulesetName = strings.TrimSpace(ruleset.Name)
	result.Rules = make([]CloudflareCacheRule, 0, len(ruleset.Rules))
	for _, rule := range ruleset.Rules {
		result.Rules = append(result.Rules, newCloudflareCacheRule(rule))
	}
	sort.Slice(result.Rules, func(i, j int) bool {
		return result.Rules[i].ID < result.Rules[j].ID
	})
	result.OriginCacheControlStatus = cloudflareOriginCacheControlStatus(result.Rules)
	return result, nil
}

func (s *CloudflareCacheRulesService) SetRuleEnabled(
	ctx context.Context,
	connectorID uint,
	zone string,
	ruleID string,
	enabled bool,
) (*CloudflareCacheRulesResult, error) {
	ruleID = strings.TrimSpace(ruleID)
	if ruleID == "" || strings.ContainsAny(ruleID, "/?#") {
		return nil, fmt.Errorf("%w: rule id is invalid", ErrInvalidCloudflareCacheRule)
	}

	current, err := s.Get(ctx, connectorID, zone)
	if err != nil {
		return nil, err
	}
	if !current.RulesetConfigured || current.RulesetID == "" {
		return nil, fmt.Errorf("%w: no cache ruleset is configured", ErrCloudflareCacheRuleNotFound)
	}

	found := false
	for _, rule := range current.Rules {
		if rule.ID == ruleID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrCloudflareCacheRuleNotFound, ruleID)
	}

	payload, err := json.Marshal(map[string]bool{"enabled": enabled})
	if err != nil {
		return nil, fmt.Errorf("%w: encode update: %v", ErrCloudflareCacheRules, err)
	}
	var updated cloudflareCacheRuleRemote
	_, err = s.connectorService.CloudflareWrite(
		ctx,
		connectorID,
		http.MethodPatch,
		"/zones/"+url.PathEscape(current.ZoneID)+"/rulesets/"+url.PathEscape(current.RulesetID)+"/rules/"+url.PathEscape(ruleID),
		payload,
		&updated,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: update rule %s: %v", ErrCloudflareCacheRules, ruleID, err)
	}
	return s.Get(ctx, connectorID, current.Zone)
}

func (s *CloudflareCacheRulesService) findZoneID(ctx context.Context, connectorID uint, zone string) (string, error) {
	var zones struct {
		Result []cloudflareZone `json:"result"`
	}
	_, err := s.connectorService.CloudflareRead(ctx, connectorID, "/zones", url.Values{
		"name":     []string{zone},
		"status":   []string{"active"},
		"per_page": []string{"50"},
	}, &zones)
	if err != nil {
		return "", fmt.Errorf("%w: read Cloudflare Zone %s: %v", ErrCloudflareCacheRules, zone, err)
	}
	for _, candidate := range zones.Result {
		if strings.EqualFold(strings.TrimSuffix(strings.TrimSpace(candidate.Name), "."), zone) {
			if zoneID := strings.TrimSpace(candidate.ID); zoneID != "" {
				return zoneID, nil
			}
		}
	}
	return "", fmt.Errorf("%w: Cloudflare Zone %s was not found", ErrCloudflareCacheRules, zone)
}

func normalizeCloudflareCacheRuleZone(value string) (string, error) {
	value = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	if value == "" || len(value) > 253 || strings.ContainsAny(value, " \t\r\n/") || strings.Contains(value, "://") {
		return "", fmt.Errorf("%w: zone is invalid", ErrInvalidCloudflareCacheRule)
	}
	return value, nil
}

func newCloudflareCacheRule(rule cloudflareCacheRuleRemote) CloudflareCacheRule {
	edgeTTLMode := cloudflareNestedString(rule.ActionParameters, "edge_ttl", "mode")
	browserTTL := cloudflareRuleValueString(rule.ActionParameters["browser_ttl"])
	return CloudflareCacheRule{
		ID:                       strings.TrimSpace(rule.ID),
		Version:                  strings.TrimSpace(rule.Version),
		Action:                   strings.TrimSpace(rule.Action),
		Description:              strings.TrimSpace(rule.Description),
		Expression:               strings.TrimSpace(rule.Expression),
		Enabled:                  rule.Enabled,
		LastUpdated:              strings.TrimSpace(rule.LastUpdated),
		EdgeTTLMode:              edgeTTLMode,
		BrowserTTL:               browserTTL,
		OriginCacheControlStatus: cloudflareRuleOriginCacheControlStatus(rule.Action, edgeTTLMode, browserTTL),
	}
}

func cloudflareOriginCacheControlStatus(rules []CloudflareCacheRule) string {
	hasCacheRule := false
	for _, rule := range rules {
		if !rule.Enabled || rule.Action != "set_cache_settings" {
			continue
		}
		hasCacheRule = true
		if rule.OriginCacheControlStatus == "overridden" {
			return "overridden"
		}
	}
	if hasCacheRule {
		return "respected"
	}
	return "no_rules"
}

func cloudflareRuleOriginCacheControlStatus(action, edgeTTLMode, browserTTL string) string {
	if strings.TrimSpace(action) != "set_cache_settings" {
		return "not_applicable"
	}
	if strings.EqualFold(edgeTTLMode, "override_origin") {
		return "overridden"
	}
	if browserTTL != "" &&
		!strings.EqualFold(browserTTL, "respect_origin") &&
		!strings.EqualFold(browserTTL, "respect_existing_headers") &&
		browserTTL != "0" {
		return "overridden"
	}
	return "respected"
}

func cloudflareNestedString(values map[string]interface{}, key, nestedKey string) string {
	if values == nil {
		return ""
	}
	nested, ok := values[key].(map[string]interface{})
	if !ok {
		return ""
	}
	return cloudflareRuleValueString(nested[nestedKey])
}

func cloudflareRuleValueString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", typed), "0"), ".")
	case json.Number:
		return typed.String()
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	default:
		return ""
	}
}
