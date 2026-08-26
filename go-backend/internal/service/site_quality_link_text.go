package service

import (
	"strings"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
)

const (
	siteQualityLinkDescriptiveTextRuleID = sitequalitydomain.SiteQualityRuleIDDescriptiveLinkText
	siteQualityLinkTextAuditID           = sitequalitydomain.SiteQualityProviderAuditIDLinkText
)

// officialNonDescriptiveLinkTexts mirrors Lighthouse's link-text audit
// vocabulary. The provider only flags exact, trimmed link text matches.
var officialNonDescriptiveLinkTexts = map[string]map[string]struct{}{
	"en": {
		"click here":       {},
		"click this":       {},
		"go":               {},
		"here":             {},
		"information":      {},
		"learn more":       {},
		"more":             {},
		"more info":        {},
		"more information": {},
		"right here":       {},
		"read more":        {},
		"see more":         {},
		"start":            {},
		"this":             {},
	},
	"ja": {
		"ここをクリック":  {},
		"こちらをクリック": {},
		"リンク":      {},
		"続きを読む":    {},
		"続く":       {},
		"全文表示":     {},
	},
	"es": {
		"click aquí":      {},
		"click aqui":      {},
		"clicka aquí":     {},
		"clicka aqui":     {},
		"pincha aquí":     {},
		"pincha aqui":     {},
		"aquí":            {},
		"aqui":            {},
		"más":             {},
		"mas":             {},
		"más información": {},
		"más informacion": {},
		"mas información": {},
		"mas informacion": {},
		"este":            {},
		"enlace":          {},
		"este enlace":     {},
		"empezar":         {},
	},
	"pt": {
		"clique aqui":      {},
		"ir":               {},
		"mais informação":  {},
		"mais informações": {},
		"mais":             {},
		"veja mais":        {},
	},
	"ko": {
		"여기":     {},
		"여기를 클릭": {},
		"클릭":     {},
		"링크":     {},
		"자세히":    {},
		"자세히 보기": {},
		"계속":     {},
		"이동":     {},
		"전체 보기":  {},
	},
	"sv": {
		"här":             {},
		"klicka här":      {},
		"läs mer":         {},
		"mer":             {},
		"mer info":        {},
		"mer information": {},
	},
	"de": {
		"klicke hier":  {},
		"hier klicken": {},
		"hier":         {},
		"mehr":         {},
		"siehe":        {},
		"dies":         {},
		"das":          {},
		"weiterlesen":  {},
	},
	"ta": {
		"அடுத்த பக்கம்":               {},
		"மறுபக்கம்":                   {},
		"முந்தைய பக்கம்":              {},
		"முன்பக்கம்":                  {},
		"மேலும் அறிக":                 {},
		"மேலும் தகவலுக்கு":            {},
		"மேலும் தரவுகளுக்கு":          {},
		"தயவுசெய்து இங்கே அழுத்தவும்": {},
		"இங்கே கிளிக் செய்யவும்":      {},
	},
	"fa": {
		"اطلاعات بیشتر":   {},
		"اطلاعات":         {},
		"این":             {},
		"اینجا بزنید":     {},
		"اینجا کلیک کنید": {},
		"اینجا":           {},
		"برو":             {},
		"بیشتر بخوانید":   {},
		"بیشتر بدانید":    {},
		"بیشتر":           {},
		"شروع":            {},
	},
}

func normalizeOfficialLinkText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func officialLinkTextMatches(value string, textLang string) bool {
	normalized := normalizeOfficialLinkText(value)
	if normalized == "" {
		return false
	}

	language := strings.ToLower(strings.TrimSpace(textLang))
	if language != "" {
		language = strings.Split(language, "-")[0]
		if texts, ok := officialNonDescriptiveLinkTexts[language]; ok {
			_, found := texts[normalized]
			return found
		}
		return false
	}

	for _, texts := range officialNonDescriptiveLinkTexts {
		if _, found := texts[normalized]; found {
			return true
		}
	}
	return false
}

func officialLinkTextKind(value string) string {
	normalized := normalizeOfficialLinkText(value)
	switch normalized {
	case "learn more", "more info", "more information", "information":
		return "learn"
	default:
		return "view"
	}
}
