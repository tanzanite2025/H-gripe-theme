// Package testfixtures provides backend-owned data imported only by tests.
package testfixtures

import (
	"bytes"
	_ "embed"
)

// One fixture is shared by upload validation and logo/favicon lifecycle tests.
// It needs no storefront asset, encoder or working directory at test time.
//
//go:embed testdata/site-logo-and-favicon-upload-lifecycle-512-by-512-lossless-webp-fixture.webp
var siteLogoAndFaviconBackendOwnedLosslessWebPFixture []byte

func SiteLogoAndFaviconUploadLifecycle512By512LosslessWebPFixtureBytes() []byte {
	return bytes.Clone(siteLogoAndFaviconBackendOwnedLosslessWebPFixture)
}
