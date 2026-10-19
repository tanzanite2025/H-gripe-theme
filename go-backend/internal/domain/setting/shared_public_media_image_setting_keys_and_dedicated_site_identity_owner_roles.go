package setting

const (
	SiteLogoSettingKey    = "site_logo"
	SiteFaviconSettingKey = "site_favicon"
)

func IsSharedPublicMediaImageSettingKey(key string) bool {
	return key == SiteLogoSettingKey || IsWebsiteProfileSharedMediaImageSettingKey(key)
}

func IsDedicatedSiteIdentityOwnerSettingKey(key string) bool {
	return key == SiteLogoSettingKey || key == SiteFaviconSettingKey
}
