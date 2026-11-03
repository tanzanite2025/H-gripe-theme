// Package mapleuifont makes the existing approved admin font available to the
// independent recovery page without relying on the admin web server.
package mapleuifont

import "embed"

const ApprovedMapleUICJKFontFilename = "MapleUI-CJK.f8ce6d72e8cb.woff2"

//go:embed MapleUI-CJK.f8ce6d72e8cb.woff2
var ApprovedMapleUICJKFontFiles embed.FS
