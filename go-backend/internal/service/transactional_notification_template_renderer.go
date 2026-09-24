package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"

	"commerce-platform/internal/domain/notification"
)

var (
	ErrNotificationTemplatePlaceholderInvalid = errors.New("notification template placeholder is invalid")
	ErrNotificationTemplateSubjectUnsafe      = errors.New("notification template subject contains unsafe line breaks")
)

var notificationTemplatePlaceholderPattern = regexp.MustCompile(`\{\{\s*([a-z][a-z0-9_]*)\s*\}\}`)

// RenderedTransactionalNotification is the provider-neutral result of
// rendering a persisted template. The SMTP/API provider is intentionally not
// involved here.
type RenderedTransactionalNotification struct {
	Subject string
	HTML    string
	Text    string
}

// RenderTransactionalNotificationTemplate validates a persisted template and
// renders only controlled scalar variables. HTML values are escaped, while
// plain-text values remain readable. Rich line-item markup must be assembled by
// the caller as a single escaped variable; arbitrary template expressions are
// never evaluated.
func RenderTransactionalNotificationTemplate(templateRecord *notification.EmailTemplate, variables map[string]string) (RenderedTransactionalNotification, error) {
	if templateRecord == nil {
		return RenderedTransactionalNotification{}, errors.New("notification template is required")
	}
	definition, err := LookupTransactionalNotificationTemplate(templateRecord.Code)
	if err != nil {
		return RenderedTransactionalNotification{}, err
	}
	if err := validatePersistedTemplateDefinition(*templateRecord, definition); err != nil {
		return RenderedTransactionalNotification{}, err
	}
	if err := ValidateTransactionalNotificationVariables(templateRecord.Code, variables); err != nil {
		return RenderedTransactionalNotification{}, err
	}

	subject, err := renderNotificationTemplateString(templateRecord.SubjectTemplate, variables, false, true)
	if err != nil {
		return RenderedTransactionalNotification{}, fmt.Errorf("render notification subject: %w", err)
	}
	htmlBody, err := renderNotificationTemplateString(templateRecord.BodyHTML, variables, true, false)
	if err != nil {
		return RenderedTransactionalNotification{}, fmt.Errorf("render notification HTML body: %w", err)
	}
	textBody, err := renderNotificationTemplateString(templateRecord.BodyText, variables, false, false)
	if err != nil {
		return RenderedTransactionalNotification{}, fmt.Errorf("render notification text body: %w", err)
	}
	return RenderedTransactionalNotification{Subject: subject, HTML: htmlBody, Text: textBody}, nil
}

func renderNotificationTemplateString(source string, variables map[string]string, escapeHTML, subject bool) (string, error) {
	if strings.TrimSpace(source) == "" {
		return "", nil
	}
	if subject && (strings.ContainsAny(source, "\r\n") || strings.ContainsAny(source, "\u2028\u2029")) {
		return "", ErrNotificationTemplateSubjectUnsafe
	}
	if strings.Contains(source, "{{{") || strings.Contains(source, "}}}") {
		return "", ErrNotificationTemplatePlaceholderInvalid
	}

	matches := notificationTemplatePlaceholderPattern.FindAllStringSubmatchIndex(source, -1)
	if strings.Contains(source, "{{") && len(matches) == 0 {
		return "", ErrNotificationTemplatePlaceholderInvalid
	}
	if strings.Contains(source, "}}") && len(matches) == 0 {
		return "", ErrNotificationTemplatePlaceholderInvalid
	}

	var builder strings.Builder
	last := 0
	for _, match := range matches {
		start, end := match[0], match[1]
		nameStart, nameEnd := match[2], match[3]
		builder.WriteString(source[last:start])
		name := source[nameStart:nameEnd]
		value, ok := variables[name]
		if !ok {
			return "", fmt.Errorf("%w: %s", ErrNotificationTemplateVariableMissing, name)
		}
		if subject && strings.ContainsAny(value, "\r\n\u2028\u2029") {
			return "", ErrNotificationTemplateSubjectUnsafe
		}
		if escapeHTML {
			value = html.EscapeString(value)
		}
		builder.WriteString(value)
		last = end
	}
	builder.WriteString(source[last:])
	rendered := builder.String()
	if strings.Contains(rendered, "{{") || strings.Contains(rendered, "}}") {
		return "", ErrNotificationTemplatePlaceholderInvalid
	}
	return rendered, nil
}

// DecodeTemplateVariableList is shared by future admin/import paths that need
// to inspect the JSON contract without exposing datatypes.JSON to callers.
func DecodeTemplateVariableList(value []byte) ([]string, error) {
	var variables []string
	if err := json.Unmarshal(value, &variables); err != nil {
		return nil, err
	}
	return variables, nil
}
