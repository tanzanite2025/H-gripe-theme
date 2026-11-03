// Package adminstyletokens shares existing admin theme values with the
// independent recovery page without requiring the admin server to be running.
package adminstyletokens

import "embed"

const SharedAdminThemeTokensFilename = "admin-shared-theme-colors-and-pill-button-tokens.css"

//go:embed admin-shared-theme-colors-and-pill-button-tokens.css
var SharedAdminThemeTokenFiles embed.FS
