package setting

// Keys, together with locale, identify settings independently of mutable group
// metadata. A generic writer cannot disable media validation by changing group.
func IsWebsiteProfileSharedMediaImageSettingKey(key string) bool {
	return key == WebsiteProfileKeyAvatarURL || key == WebsiteProfileKeyFactoryImageURL
}

func IsWebsiteProfileSharedMediaRichTextSettingKey(key string) bool {
	return key == WebsiteProfileKeyStatementBody || key == WebsiteProfileKeyStatementParagraph1 || key == WebsiteProfileKeyStatementParagraph2
}
