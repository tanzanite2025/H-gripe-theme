package setting

type WebsiteNameSettings struct {
	Locale  string `json:"locale"`
	Status  string `json:"status"`
	Intro   string `json:"intro"`
	Eyebrow string `json:"eyebrow"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Note    string `json:"note"`
}

type WebsiteNameUpdateRequest struct {
	Locale  string `json:"locale"`
	Status  string `json:"status"`
	Intro   string `json:"intro"`
	Eyebrow string `json:"eyebrow"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Note    string `json:"note"`
}

func (request WebsiteNameUpdateRequest) Settings() WebsiteNameSettings {
	return WebsiteNameSettings{
		Locale:  request.Locale,
		Status:  request.Status,
		Intro:   request.Intro,
		Eyebrow: request.Eyebrow,
		Title:   request.Title,
		Body:    request.Body,
		Note:    request.Note,
	}
}

func DefaultWebsiteNameSettings(locale string) WebsiteNameSettings {
	if defaults, ok := websiteNameDefaultSettingsByLocale[locale]; ok {
		return defaults
	}

	return websiteNameDefaultSettingsByLocale["en"]
}
