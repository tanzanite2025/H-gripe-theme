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
	refundcancellationdomain "commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
	"gorm.io/gorm"
)

type RefundCancellationPolicyResult struct {
	Policy          refundcancellationdomain.Policy `json:"policy"`
	Locale          string                          `json:"locale"`
	RequestedLocale string                          `json:"requested_locale"`
	Fallback        bool                            `json:"fallback"`
}

type RefundCancellationPolicyService struct {
	settings *repository.SettingRepository
}

func NewRefundCancellationPolicyService(settings *repository.SettingRepository) *RefundCancellationPolicyService {
	return &RefundCancellationPolicyService{settings: settings}
}

func (s *RefundCancellationPolicyService) GetPublic(locale string) (RefundCancellationPolicyResult, error) {
	return s.get(locale, true)
}

func (s *RefundCancellationPolicyService) GetAdmin(locale string) (RefundCancellationPolicyResult, error) {
	return s.get(locale, false)
}

func (s *RefundCancellationPolicyService) Update(locale string, policy refundcancellationdomain.Policy) (RefundCancellationPolicyResult, error) {
	if s == nil || s.settings == nil {
		return RefundCancellationPolicyResult{}, errors.New("refund and cancellation policy service is unavailable")
	}

	locale = normalizePolicyLocale(locale)
	normalized, err := refundcancellationdomain.Normalize(policy)
	if err != nil {
		return RefundCancellationPolicyResult{}, err
	}
	normalized.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	payload, err := json.Marshal(normalized)
	if err != nil {
		return RefundCancellationPolicyResult{}, err
	}
	if err := s.settings.BatchSet([]setting.Setting{{
		Key:         refundcancellationdomain.Key,
		Value:       string(payload),
		Type:        "json",
		Locale:      locale,
		Group:       refundcancellationdomain.Group,
		IsPublic:    true,
		Description: "Editable refund and cancellation policy content",
	}}); err != nil {
		return RefundCancellationPolicyResult{}, err
	}

	return RefundCancellationPolicyResult{
		Policy:          normalized,
		Locale:          locale,
		RequestedLocale: locale,
		Fallback:        false,
	}, nil
}

func (s *RefundCancellationPolicyService) BuildOrderDisclosure(
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
		return nil, errors.New("refund and cancellation policy settings repository is unavailable")
	}

	// The order snapshot must match the public policy rendered to the buyer.
	// Never capture an admin-only draft as historical customer evidence.
	result, err := s.getWithSettings(settings, locale, true)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(result.Policy)
	if err != nil {
		return nil, fmt.Errorf("marshal refund and cancellation policy snapshot: %w", err)
	}
	hash := sha256.Sum256(payload)
	hashValue := hex.EncodeToString(hash[:])
	if strings.TrimSpace(policyURL) == "" {
		policyURL = "/policies/refund-cancellation"
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
		PolicyKey:       refundcancellationdomain.Key,
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

func (s *RefundCancellationPolicyService) get(locale string, publicOnly bool) (RefundCancellationPolicyResult, error) {
	if s == nil || s.settings == nil {
		return RefundCancellationPolicyResult{}, errors.New("refund and cancellation policy service is unavailable")
	}
	return s.getWithSettings(s.settings, locale, publicOnly)
}

func (s *RefundCancellationPolicyService) getWithSettings(settings *repository.SettingRepository, locale string, publicOnly bool) (RefundCancellationPolicyResult, error) {
	requestedLocale := normalizePolicyLocale(locale)
	for _, candidateLocale := range policyLocaleFallbacks(requestedLocale) {
		var record *setting.Setting
		var err error
		if publicOnly {
			record, err = settings.GetPublic(refundcancellationdomain.Key, candidateLocale)
		} else {
			record, err = settings.Get(refundcancellationdomain.Key, candidateLocale)
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return RefundCancellationPolicyResult{}, fmt.Errorf("load refund and cancellation policy for locale %q: %w", candidateLocale, err)
		}

		policy, err := decodePolicy(record.Value)
		if err != nil {
			return RefundCancellationPolicyResult{}, err
		}
		return RefundCancellationPolicyResult{
			Policy:          policy,
			Locale:          candidateLocale,
			RequestedLocale: requestedLocale,
			Fallback:        candidateLocale != requestedLocale,
		}, nil
	}

	policy, err := refundcancellationdomain.Normalize(refundcancellationdomain.DefaultPolicy())
	if err != nil {
		return RefundCancellationPolicyResult{}, err
	}
	return RefundCancellationPolicyResult{
		Policy:          policy,
		Locale:          "en",
		RequestedLocale: requestedLocale,
		Fallback:        true,
	}, nil
}

func decodePolicy(value string) (refundcancellationdomain.Policy, error) {
	var policy refundcancellationdomain.Policy
	if err := json.Unmarshal([]byte(value), &policy); err != nil {
		return refundcancellationdomain.Policy{}, errors.New("stored refund and cancellation policy is invalid JSON")
	}
	return refundcancellationdomain.Normalize(policy)
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
