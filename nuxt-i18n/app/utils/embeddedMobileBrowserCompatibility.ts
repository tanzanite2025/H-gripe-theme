export type EmbeddedMobileBrowserName =
  | 'wechat'
  | 'facebook'
  | 'instagram'
  | 'whatsapp'
  | 'messenger'
  | 'line'
  | 'telegram'
  | 'tiktok'
  | 'snapchat'
  | 'google-app'
  | 'generic-webview'
  | 'unknown'

export interface EmbeddedMobileBrowserDetectionResult {
  isEmbeddedMobileBrowser: boolean
  browserName: EmbeddedMobileBrowserName
}

const embeddedMobileBrowserMatchers: Array<{
  browserName: Exclude<EmbeddedMobileBrowserName, 'unknown' | 'generic-webview'>
  pattern: RegExp
}> = [
  { browserName: 'wechat', pattern: /micromessenger|wechat/i },
  {
    browserName: 'messenger',
    pattern: /messenger|fban\/messenger|fb_iab\/orca/i,
  },
  { browserName: 'facebook', pattern: /fban|fbav|fb_iab|fb4a|fbios/i },
  { browserName: 'instagram', pattern: /instagram/i },
  { browserName: 'whatsapp', pattern: /whatsapp/i },
  { browserName: 'line', pattern: /line\//i },
  { browserName: 'telegram', pattern: /telegram/i },
  { browserName: 'tiktok', pattern: /tiktok|musical_ly|bytedancewebview/i },
  { browserName: 'snapchat', pattern: /snapchat/i },
  { browserName: 'google-app', pattern: /\bGSA\//i },
]

const isMobileDeviceUserAgent = (userAgent: string): boolean =>
  /android|iphone|ipad|ipod|iemobile|opera mini|mobile/i.test(userAgent)

const isGenericAndroidWebViewUserAgent = (userAgent: string): boolean =>
  /android/i.test(userAgent) && /;\s*wv(?:\)|;|\s)/i.test(userAgent)

export const detectEmbeddedMobileBrowser = (
  userAgent: unknown,
): EmbeddedMobileBrowserDetectionResult => {
  const normalizedUserAgent = String(userAgent || '').trim()
  if (!normalizedUserAgent || !isMobileDeviceUserAgent(normalizedUserAgent)) {
    return { isEmbeddedMobileBrowser: false, browserName: 'unknown' }
  }

  const matchedBrowser = embeddedMobileBrowserMatchers.find(({ pattern }) =>
    pattern.test(normalizedUserAgent),
  )
  if (matchedBrowser) {
    return {
      isEmbeddedMobileBrowser: true,
      browserName: matchedBrowser.browserName,
    }
  }

  if (isGenericAndroidWebViewUserAgent(normalizedUserAgent)) {
    return {
      isEmbeddedMobileBrowser: true,
      browserName: 'generic-webview',
    }
  }

  return { isEmbeddedMobileBrowser: false, browserName: 'unknown' }
}

export const buildSecureCurrentPageUrl = (currentPageUrl: string): string => {
  try {
    const secureUrl = new URL(currentPageUrl)
    if (!['http:', 'https:'].includes(secureUrl.protocol)) return ''
    secureUrl.username = ''
    secureUrl.password = ''
    if (secureUrl.protocol === 'http:') secureUrl.protocol = 'https:'
    return secureUrl.toString()
  } catch {
    return ''
  }
}
