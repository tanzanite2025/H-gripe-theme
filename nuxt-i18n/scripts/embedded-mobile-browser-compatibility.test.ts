import assert from 'node:assert/strict'
import {
  buildSecureCurrentPageUrl,
  detectEmbeddedMobileBrowser,
  type EmbeddedMobileBrowserName,
} from '../app/utils/embeddedMobileBrowserCompatibility.js'

const androidBrowserPrefix =
  'Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36'
const iphoneBrowserPrefix =
  'Mozilla/5.0 (iPhone; CPU iPhone OS 17_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148'
const embeddedBrowserCases: Array<[string, EmbeddedMobileBrowserName]> = [
  [`${androidBrowserPrefix} MicroMessenger/8.0.40`, 'wechat'],
  [`${iphoneBrowserPrefix} MicroMessenger/8.0.40`, 'wechat'],
  [`${androidBrowserPrefix} [FB_IAB/FB4A;FBAV/492.0.0.0.60;]`, 'facebook'],
  [`${iphoneBrowserPrefix} [FBAN/FBIOS;FBAV/492.0.0.0.60;]`, 'facebook'],
  [`${iphoneBrowserPrefix} Instagram 340.0`, 'instagram'],
  [`${androidBrowserPrefix} WhatsApp/2.24.20.10`, 'whatsapp'],
  [`${iphoneBrowserPrefix} [FBAN/MessengerForiOS;FBAV/492.0;]`, 'messenger'],
  [`${androidBrowserPrefix} [FB_IAB/Orca-Android;FBAV/492.0;]`, 'messenger'],
  [`${iphoneBrowserPrefix} Line/14.0.0`, 'line'],
  [`${iphoneBrowserPrefix} Telegram/10.0`, 'telegram'],
  [
    `${androidBrowserPrefix} musical_ly_2024001 BytedanceWebview/d8a21c6`,
    'tiktok',
  ],
  [`${iphoneBrowserPrefix} TikTok/35.0`, 'tiktok'],
  [`${iphoneBrowserPrefix} Snapchat/13.0`, 'snapchat'],
  [
    `${iphoneBrowserPrefix} GSA/350.0.712345678 Mobile/15E148 Safari/604.1`,
    'google-app',
  ],
  [
    'Mozilla/5.0 (Linux; Android 13; Pixel 7 Build/ABC; wv) AppleWebKit/537.36 Version/4.0 Chrome/120.0.0.0 Mobile Safari/537.36',
    'generic-webview',
  ],
]

for (const [userAgent, browserName] of embeddedBrowserCases) {
  assert.deepEqual(
    detectEmbeddedMobileBrowser(userAgent),
    {
      isEmbeddedMobileBrowser: true,
      browserName,
    },
    userAgent,
  )
}

const ordinaryBrowserCases: unknown[] = [
  androidBrowserPrefix,
  `${iphoneBrowserPrefix} Version/17.6 Safari/604.1`,
  `${iphoneBrowserPrefix} CriOS/120.0.0.0 Safari/604.1`,
  `${androidBrowserPrefix} SamsungBrowser/24.0`,
  'Mozilla/5.0 (Linux; Android 4.4; Nexus 5) AppleWebKit/537.36 Version/4.0 Mobile Safari/537.36',
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0',
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) MicroMessenger/3.9',
  'facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)',
  '',
  undefined,
  null,
]

for (const userAgent of ordinaryBrowserCases) {
  assert.deepEqual(
    detectEmbeddedMobileBrowser(userAgent),
    {
      isEmbeddedMobileBrowser: false,
      browserName: 'unknown',
    },
    String(userAgent),
  )
}

for (const hostname of [
  'learn.gripe',
  'replacement-domain.example',
  'shop.new-brand.example',
]) {
  const pagePath =
    '/zh_cn/tools/a%2Fb?size=700&return=https%3A%2F%2Fexample.com%2Fa#results'
  for (const scheme of ['http', 'https']) {
    assert.equal(
      buildSecureCurrentPageUrl(`${scheme}://${hostname}${pagePath}`),
      `https://${hostname}${pagePath}`,
    )
  }
}
assert.equal(
  buildSecureCurrentPageUrl(
    'https://user:password@new-brand.example:8443/path',
  ),
  'https://new-brand.example:8443/path',
)
assert.equal(
  buildSecureCurrentPageUrl('http://new-brand.example:8080/path'),
  'https://new-brand.example:8080/path',
)
for (const invalidUrl of [
  'javascript:alert(1)',
  'data:text/html,test',
  'file:///etc/hosts',
  '//new-brand.example/path',
  '/relative',
  'not a URL',
  'https://',
]) {
  assert.equal(buildSecureCurrentPageUrl(invalidUrl), '', invalidUrl)
}

console.log(
  `Embedded browser detection: ${embeddedBrowserCases.length} embedded and ${ordinaryBrowserCases.length} ordinary/empty agents passed; domain-independent URL safety passed.`,
)
