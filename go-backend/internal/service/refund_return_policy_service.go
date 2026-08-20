package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	refundreturndomain "commerce-platform/internal/domain/refundreturn"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
	"gorm.io/gorm"
)

type RefundReturnPolicyResult struct {
	Policy          refundreturndomain.Policy `json:"policy"`
	Locale          string                    `json:"locale"`
	RequestedLocale string                    `json:"requested_locale"`
	Fallback        bool                      `json:"fallback"`
}

type RefundReturnPolicyService struct {
	settings *repository.SettingRepository
}

func NewRefundReturnPolicyService(settings *repository.SettingRepository) *RefundReturnPolicyService {
	return &RefundReturnPolicyService{settings: settings}
}

func (s *RefundReturnPolicyService) GetPublic(locale string) (RefundReturnPolicyResult, error) {
	return s.get(locale, true)
}

func (s *RefundReturnPolicyService) GetAdmin(locale string) (RefundReturnPolicyResult, error) {
	return s.get(locale, false)
}

func (s *RefundReturnPolicyService) Update(locale string, policy refundreturndomain.Policy) (RefundReturnPolicyResult, error) {
	if s == nil || s.settings == nil {
		return RefundReturnPolicyResult{}, errors.New("refund and return policy service is unavailable")
	}

	locale = normalizePolicyLocale(locale)
	normalized, err := refundreturndomain.Normalize(policy)
	if err != nil {
		return RefundReturnPolicyResult{}, err
	}
	normalized.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	payload, err := json.Marshal(normalized)
	if err != nil {
		return RefundReturnPolicyResult{}, err
	}
	if err := s.settings.BatchSet([]setting.Setting{{
		Key:         refundreturndomain.Key,
		Value:       string(payload),
		Type:        "json",
		Locale:      locale,
		Group:       refundreturndomain.Group,
		IsPublic:    true,
		Description: "Editable refund and return policy content",
	}}); err != nil {
		return RefundReturnPolicyResult{}, err
	}

	return RefundReturnPolicyResult{
		Policy:          normalized,
		Locale:          locale,
		RequestedLocale: locale,
		Fallback:        false,
	}, nil
}

func (s *RefundReturnPolicyService) BuildOrderDisclosure(
	settings *repository.SettingRepository,
	orderID uint,
	locale string,
	policyURL string,
	source string,
	consentedAt *time.Time,
) (*orderdomain.PolicyDisclosure, error) {
	if orderID == 0 {
		return nil, errors.New("order id is required for policy disclosure")
	}
	if settings == nil {
		return nil, errors.New("refund and return policy settings repository is unavailable")
	}

	// The order snapshot must match the public policy rendered to the buyer.
	// Never capture an admin-only draft as historical customer evidence.
	result, err := s.getWithSettings(settings, locale, true)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(result.Policy)
	if err != nil {
		return nil, fmt.Errorf("marshal refund and return policy snapshot: %w", err)
	}
	hash := sha256.Sum256(payload)
	hashValue := hex.EncodeToString(hash[:])
	if strings.TrimSpace(policyURL) == "" {
		policyURL = "/policies/refund-return"
	}
	if strings.TrimSpace(source) == "" {
		source = "checkout_order_creation"
	}
	disclosedAt := time.Now().UTC()
	var normalizedConsent *time.Time
	if consentedAt != nil && !consentedAt.IsZero() {
		value := consentedAt.UTC()
		normalizedConsent = &value
	}

	return &orderdomain.PolicyDisclosure{
		OrderID:         orderID,
		PolicyKey:       refundreturndomain.Key,
		Locale:          result.Locale,
		RequestedLocale: result.RequestedLocale,
		Fallback:        result.Fallback,
		PolicyVersion:   "sha256:" + hashValue,
		PolicyHash:      hashValue,
		PolicyJSON:      string(payload),
		PolicyURL:       strings.TrimSpace(policyURL),
		DisclosedAt:     disclosedAt,
		ConsentedAt:     normalizedConsent,
		Source:          source,
	}, nil
}

func (s *RefundReturnPolicyService) get(locale string, publicOnly bool) (RefundReturnPolicyResult, error) {
	if s == nil || s.settings == nil {
		return RefundReturnPolicyResult{}, errors.New("refund and return policy service is unavailable")
	}
	return s.getWithSettings(s.settings, locale, publicOnly)
}

func (s *RefundReturnPolicyService) getWithSettings(settings *repository.SettingRepository, locale string, publicOnly bool) (RefundReturnPolicyResult, error) {
	requestedLocale := normalizePolicyLocale(locale)
	for _, candidateLocale := range policyLocaleFallbacks(requestedLocale) {
		var record *setting.Setting
		var err error
		if publicOnly {
			record, err = settings.GetPublic(refundreturndomain.Key, candidateLocale)
		} else {
			record, err = settings.Get(refundreturndomain.Key, candidateLocale)
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return RefundReturnPolicyResult{}, fmt.Errorf("load refund and return policy for locale %q: %w", candidateLocale, err)
		}

		policy, err := decodePolicy(record.Value)
		if err != nil {
			return RefundReturnPolicyResult{}, err
		}
		return RefundReturnPolicyResult{
			Policy:          policy,
			Locale:          candidateLocale,
			RequestedLocale: requestedLocale,
			Fallback:        candidateLocale != requestedLocale,
		}, nil
	}

	policy, err := refundreturndomain.Normalize(refundreturndomain.DefaultPolicy())
	if err != nil {
		return RefundReturnPolicyResult{}, err
	}
	return RefundReturnPolicyResult{
		Policy:          policy,
		Locale:          "en",
		RequestedLocale: requestedLocale,
		Fallback:        true,
	}, nil
}

func decodePolicy(value string) (refundreturndomain.Policy, error) {
	var policy refundreturndomain.Policy
	if err := json.Unmarshal([]byte(value), &policy); err != nil {
		return refundreturndomain.Policy{}, errors.New("stored refund and return policy is invalid JSON")
	}
	return refundreturndomain.Normalize(policy)
}

func normalizePolicyLocale(locale string) string {
	locale = strings.TrimSpace(strings.ToLower(strings.ReplaceAll(locale, "-", "_")))
	if locale == "" {
		return "en"
	}
	if locale == "zh" || locale == "zh_cn" || locale == "zh_hans" {
		return "zh_cn"
	}
	if index := strings.Index(locale, "_"); index > 0 {
		return locale[:index]
	}
	return locale
}

func policyLocaleFallbacks(locale string) []string {
	locales := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, candidate := range []string{locale, "en"} {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		locales = append(locales, candidate)
	}
	return locales
}
