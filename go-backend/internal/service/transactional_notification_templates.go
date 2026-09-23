package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"commerce-platform/internal/pkg/locales"
)

var (
	ErrNotificationTemplateUnknown            = errors.New("notification template is not registered")
	ErrNotificationTemplateLocaleUnavailable  = errors.New("notification template locale is unavailable")
	ErrNotificationTemplateVariableUnknown    = errors.New("notification template variable is not allowed")
	ErrNotificationTemplateVariableMissing    = errors.New("notification template required variable is missing")
	ErrNotificationTemplateDefinitionMismatch = errors.New("notification template persisted variable definition does not match contract")
)

// NotificationTemplateDefinition is the stable contract consumed by a future
// renderer and template repository. It intentionally contains metadata only;
// template bodies and provider configuration remain outside this phase.
type NotificationTemplateDefinition struct {
	Code              string
	Category          string
	RequiredVariables []string
	AllowedVariables  []string
	Locale            string
}

var notificationTemplateDefinitions = map[string]NotificationTemplateDefinition{
	NotificationTemplateOrderConfirmation: {
		Code:              NotificationTemplateOrderConfirmation,
		Category:          "order",
		RequiredVariables: []string{"order_number", "order_amount", "paid_at"},
		AllowedVariables:  []string{"order_number", "order_amount", "paid_at", "customer_name", "currency", "items"},
	},
	NotificationTemplateOrderPaymentExpired: {
		Code:              NotificationTemplateOrderPaymentExpired,
		Category:          "order",
		RequiredVariables: []string{"order_number"},
		AllowedVariables:  []string{"order_number", "customer_name", "expired_at"},
	},
	NotificationTemplateOrderCancelled: {
		Code:              NotificationTemplateOrderCancelled,
		Category:          "order",
		RequiredVariables: []string{"order_number"},
		AllowedVariables:  []string{"order_number", "customer_name", "cancelled_at", "cancellation_reason"},
	},
	NotificationTemplateOrderShippingNotification: {
		Code:              NotificationTemplateOrderShippingNotification,
		Category:          "order",
		RequiredVariables: []string{"order_number", "carrier_name", "tracking_number", "tracking_url"},
		AllowedVariables:  []string{"order_number", "carrier_name", "tracking_number", "tracking_url", "customer_name", "shipped_at", "items"},
	},
	NotificationTemplateOrderDelivered: {
		Code:              NotificationTemplateOrderDelivered,
		Category:          "order",
		RequiredVariables: []string{"order_number", "delivered_at"},
		AllowedVariables:  []string{"order_number", "delivered_at", "customer_name", "tracking_number"},
	},
	NotificationTemplateOrderCompleted: {
		Code:              NotificationTemplateOrderCompleted,
		Category:          "order",
		RequiredVariables: []string{"order_number", "completed_at"},
		AllowedVariables:  []string{"order_number", "completed_at", "customer_name"},
	},
	NotificationTemplateOrderRefunded: {
		Code:              NotificationTemplateOrderRefunded,
		Category:          "order",
		RequiredVariables: []string{"order_number", "refund_amount", "currency"},
		AllowedVariables:  []string{"order_number", "refund_amount", "currency", "customer_name", "refunded_at", "refund_reason"},
	},
	NotificationTemplateAfterSalesRequested: {
		Code:              NotificationTemplateAfterSalesRequested,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "order_number", "after_sales_type", "after_sales_status"},
		AllowedVariables:  []string{"after_sales_case_number", "order_number", "after_sales_type", "after_sales_status", "customer_name", "requested_at"},
	},
	NotificationTemplateAfterSalesApproved: {
		Code:              NotificationTemplateAfterSalesApproved,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "after_sales_type", "after_sales_status"},
		AllowedVariables:  []string{"after_sales_case_number", "after_sales_type", "after_sales_status", "customer_name", "approved_at", "next_step"},
	},
	NotificationTemplateAfterSalesAwaitingReturn: {
		Code:              NotificationTemplateAfterSalesAwaitingReturn,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "return_warehouse_name", "return_warehouse_address"},
		AllowedVariables:  []string{"after_sales_case_number", "return_warehouse_name", "return_warehouse_address", "return_label_url", "customer_name", "return_deadline"},
	},
	NotificationTemplateAfterSalesReturnInTransit: {
		Code:              NotificationTemplateAfterSalesReturnInTransit,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "carrier_name", "tracking_number", "tracking_url"},
		AllowedVariables:  []string{"after_sales_case_number", "carrier_name", "tracking_number", "tracking_url", "customer_name", "shipped_at"},
	},
	NotificationTemplateAfterSalesReceived: {
		Code:              NotificationTemplateAfterSalesReceived,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "received_at"},
		AllowedVariables:  []string{"after_sales_case_number", "received_at", "customer_name", "carrier_name", "tracking_number"},
	},
	NotificationTemplateAfterSalesResolving: {
		Code:              NotificationTemplateAfterSalesResolving,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "after_sales_status"},
		AllowedVariables:  []string{"after_sales_case_number", "after_sales_status", "customer_name", "resolution_note"},
	},
	NotificationTemplateAfterSalesCompleted: {
		Code:              NotificationTemplateAfterSalesCompleted,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "after_sales_status", "refund_amount"},
		AllowedVariables:  []string{"after_sales_case_number", "after_sales_status", "refund_amount", "customer_name", "completed_at", "resolution"},
	},
	NotificationTemplateAfterSalesRejected: {
		Code:              NotificationTemplateAfterSalesRejected,
		Category:          "after_sales",
		RequiredVariables: []string{"after_sales_case_number", "rejection_reason"},
		AllowedVariables:  []string{"after_sales_case_number", "rejection_reason", "customer_name", "rejected_at"},
	},
}

// LookupTransactionalNotificationTemplate returns a defensive copy so callers
// cannot mutate the process-wide contract registry.
func LookupTransactionalNotificationTemplate(code string) (NotificationTemplateDefinition, error) {
	code = strings.TrimSpace(code)
	definition, ok := notificationTemplateDefinitions[code]
	if !ok {
		return NotificationTemplateDefinition{}, fmt.Errorf("%w: %s", ErrNotificationTemplateUnknown, code)
	}
	definition.RequiredVariables = append([]string(nil), definition.RequiredVariables...)
	definition.AllowedVariables = append([]string(nil), definition.AllowedVariables...)
	return definition, nil
}

// ValidateTransactionalNotificationTemplateResolution ensures the event rule
// and the template contract agree on the required variable set. Keeping this
// check at the Outbox boundary catches drift before a future renderer sends an
// incomplete message.
func ValidateTransactionalNotificationTemplateResolution(resolution NotificationTemplateResolution) error {
	if strings.TrimSpace(resolution.TemplateCode) == "" {
		return nil
	}
	definition, err := LookupTransactionalNotificationTemplate(resolution.TemplateCode)
	if err != nil {
		return err
	}
	if len(definition.RequiredVariables) != len(resolution.RequiredVariables) {
		return fmt.Errorf("notification template %q required variable contract mismatch", resolution.TemplateCode)
	}
	for index, variable := range definition.RequiredVariables {
		if variable != resolution.RequiredVariables[index] {
			return fmt.Errorf("notification template %q required variable contract mismatch", resolution.TemplateCode)
		}
	}
	return nil
}

// ResolveTransactionalNotificationTemplateDefinition applies the locale
// fallback contract to a persisted template locale list. Exact locale is
// preferred, then the default English locale. No template content is loaded.
func ResolveTransactionalNotificationTemplateDefinition(code, requestedLocale string, availableLocales []string) (NotificationTemplateDefinition, error) {
	definition, err := LookupTransactionalNotificationTemplate(code)
	if err != nil {
		return NotificationTemplateDefinition{}, err
	}

	available := make(map[string]struct{}, len(availableLocales))
	for _, candidate := range availableLocales {
		if normalized := locales.ResolveSupported(candidate); normalized != "" {
			available[normalized] = struct{}{}
		}
	}
	if len(available) == 0 {
		return NotificationTemplateDefinition{}, fmt.Errorf("%w: %s has no enabled locale", ErrNotificationTemplateLocaleUnavailable, code)
	}

	wanted := locales.ResolveSupported(requestedLocale)
	if wanted == "" {
		wanted = "en"
	}
	if _, ok := available[wanted]; !ok {
		if _, ok := available["en"]; !ok {
			return NotificationTemplateDefinition{}, fmt.Errorf("%w: %s", ErrNotificationTemplateLocaleUnavailable, code)
		}
		wanted = "en"
	}
	definition.Locale = wanted
	return definition, nil
}

// ValidateTransactionalNotificationVariables enforces the template contract
// before a future renderer is allowed to produce a message. Unknown variables
// are rejected to prevent accidental leakage of unrelated payload fields.
func ValidateTransactionalNotificationVariables(templateCode string, variables map[string]string) error {
	definition, err := LookupTransactionalNotificationTemplate(templateCode)
	if err != nil {
		return err
	}
	allowed := make(map[string]struct{}, len(definition.AllowedVariables))
	for _, variable := range definition.AllowedVariables {
		allowed[variable] = struct{}{}
	}
	for variable := range variables {
		if _, ok := allowed[strings.TrimSpace(variable)]; !ok {
			return fmt.Errorf("%w: %s", ErrNotificationTemplateVariableUnknown, variable)
		}
	}
	for _, variable := range definition.RequiredVariables {
		if strings.TrimSpace(variables[variable]) == "" {
			return fmt.Errorf("%w: %s", ErrNotificationTemplateVariableMissing, variable)
		}
	}
	return nil
}

// RegisteredTransactionalNotificationTemplateCodes returns stable codes in a
// deterministic order for diagnostics and future admin/API discovery.
func RegisteredTransactionalNotificationTemplateCodes() []string {
	codes := make([]string, 0, len(notificationTemplateDefinitions))
	for code := range notificationTemplateDefinitions {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}
