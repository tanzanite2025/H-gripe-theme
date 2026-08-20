import { expect, test, type Page } from '@playwright/test'

const storefrontUrl = process.env.STOREFRONT_URL || 'http://127.0.0.1:9199/'
const latinFontPath = '/fonts/MapleUI-Latin.00af3fec5b34.woff2'
const cjkFontPath = '/fonts/MapleUI-CJK.f8ce6d72e8cb.woff2'
const latinAccentFontPath = '/fonts/MapleUI-Coverage-NotoSans-Latin-Accents.e645edc952b6.woff2'
const arabicFontPath = '/fonts/MapleUI-Coverage-NotoSans-Arabic.ce85091f0209.woff2'
const devanagariFontPath = '/fonts/MapleUI-Coverage-NotoSans-Devanagari.3b3cae4d2600.woff2'
const thaiFontPath = '/fonts/MapleUI-Coverage-NotoSans-Thai.1f5a173641bb.woff2'
const approvedStorefrontFontPaths = [
  latinFontPath,
  cjkFontPath,
  latinAccentFontPath,
  arabicFontPath,
  devanagariFontPath,
  thaiFontPath,
]
const nonDefaultEnglishHomepageFontPaths = [
  cjkFontPath,
  latinAccentFontPath,
  arabicFontPath,
  devanagariFontPath,
  thaiFontPath,
]
const optionalHomepageFontRangeSpecs = [
  {
    label: 'CJK',
    ranges: [
      [0x0250, 0x02ff], [0x0370, 0x052f], [0x0530, 0x058f], [0x1f00, 0x1fff],
      [0x2070, 0x209f], [0x2150, 0x218f], [0x2c00, 0x2dff], [0x2e80, 0x2eff],
      [0x3000, 0x303f], [0x3040, 0x30ff], [0x3100, 0x31ff], [0x3200, 0x33ff],
      [0x3400, 0x4dbf], [0x4e00, 0x9fff], [0xac00, 0xd7af], [0xf900, 0xfaff],
      [0xfe30, 0xfe4f], [0xff00, 0xffef], [0x20000, 0x2fa1f],
    ],
  },
  { label: 'LatinAccents', ranges: [[0x00c0, 0x00c1], [0x00c4, 0x00c5], [0x00c8, 0x00c9], [0x00cb, 0x00cb], [0x00cc, 0x00cd], [0x00cf, 0x00cf], [0x00d2, 0x00d3], [0x00d6, 0x00d6], [0x00d9, 0x00da], [0x00dc, 0x00dc], [0x0116, 0x0116], [0x0130, 0x0130], [0x016e, 0x016e], [0x01d7, 0x01d7], [0x01db, 0x01db], [0x01fa, 0x01fa], [0x0226, 0x0226], [0x022e, 0x022e], [0x0300, 0x0301], [0x0307, 0x0308], [0x030a, 0x030a], [0x1e2e, 0x1e2e], [0xfe00, 0xfe00]] },
  { label: 'Arabic', ranges: [[0x0600, 0x06ff], [0x0750, 0x077f], [0x0870, 0x08ff], [0xfb50, 0xfdff], [0xfe70, 0xfefc], [0x102e0, 0x102fb], [0x10e60, 0x10e7e], [0x10ec2, 0x10eff], [0x1ee00, 0x1eef1]] },
  { label: 'Devanagari', ranges: [[0x0900, 0x097f], [0x1cd0, 0x1cf9], [0x20a8, 0x20a8], [0x20b9, 0x20b9], [0xa830, 0xa839], [0xa8e0, 0xa8ff], [0x11b00, 0x11b09]] },
  { label: 'Thai', ranges: [[0x02d7, 0x02d7], [0x0303, 0x0303], [0x0331, 0x0331], [0x0e01, 0x0e5b], [0x200c, 0x200d], [0x25cc, 0x25cc]] },
]
const expectedFontDisplayByFamily = {
  MapleUILatin: 'block',
  MapleUICJK: 'block',
  MapleUICoverageNotoSansDevanagari: 'block',
  MapleUICoverageNotoSansLatinAccents: 'block',
  MapleUICoverageNotoSansArabic: 'block',
  MapleUICoverageNotoSansThai: 'block',
} as const

const fontPathname = (url: string): string => {
  try {
    return new URL(url).pathname
  } catch {
    return url
  }
}

const collectLoadedFontPaths = async (
  page: Page,
  sample: string,
  fontFamily: string,
): Promise<string[]> => {
  const blankUrl = new URL(`/__font-contract-blank?sample=${encodeURIComponent(sample)}`, storefrontUrl).toString()
  await page.route(blankUrl, route => route.fulfill({
    status: 200,
    contentType: 'text/html; charset=utf-8',
    body: '<!doctype html><html><head><title>font contract</title></head><body></body></html>',
  }), { times: 1 })
  await page.goto(blankUrl)

  return page.evaluate(async ({ cssHref, sample, fontFamily }) => {
    performance.clearResourceTimings()

    await new Promise<void>((resolve, reject) => {
      const existing = document.querySelector<HTMLLinkElement>('link[data-font-contract-stylesheet]')
      existing?.remove()

      const link = document.createElement('link')
      link.rel = 'stylesheet'
      link.href = cssHref
      link.dataset.fontContractStylesheet = 'true'
      link.addEventListener('load', () => resolve(), { once: true })
      link.addEventListener('error', () => reject(new Error(`Unable to load ${cssHref}`)), { once: true })
      document.head.append(link)
    })

    const sampleElement = document.createElement('p')
    sampleElement.textContent = sample
    Object.assign(sampleElement.style, {
      fontFamily,
      fontSize: '64px',
      fontWeight: '400',
      lineHeight: '1.2',
      margin: '0',
      position: 'absolute',
      visibility: 'hidden',
      whiteSpace: 'pre',
    })
    document.body.replaceChildren(sampleElement)
    sampleElement.getBoundingClientRect()

    await document.fonts.load(`400 64px ${fontFamily}`, sample)
    await document.fonts.ready
    await new Promise<void>((resolve) => {
      requestAnimationFrame(() => requestAnimationFrame(() => resolve()))
    })

    return performance.getEntriesByType('resource')
      .map(entry => new URL(entry.name).pathname)
      .filter(path => path.startsWith('/fonts/MapleUI-') && path.endsWith('.woff2'))
      .sort()
  }, { cssHref: new URL('/fonts/maple-ui.css', storefrontUrl).toString(), sample, fontFamily })
}

test('the English storefront home page only requests the default Latin shard', async ({ page }) => {
  const requestedFontPaths: string[] = []
  page.on('request', (request) => {
    const pathname = fontPathname(request.url())
    if (pathname.startsWith('/fonts/MapleUI-') && pathname.endsWith('.woff2')) {
      requestedFontPaths.push(pathname)
    }
  })

  await page.goto(storefrontUrl, { waitUntil: 'domcontentloaded' })
  await page.waitForLoadState('networkidle', { timeout: 15_000 }).catch(() => undefined)

  const report = await page.evaluate(async ({ rangeSpecs }) => {
    await document.fonts.ready

    const pathFor = (element: Element): string => {
      const parts: string[] = []
      let node: Element | null = element

      while (node && parts.length < 6) {
        let part = node.tagName.toLowerCase()
        if (node.id) part += `#${node.id}`

        const classes = String(node.className || '')
          .split(/\s+/)
          .filter(Boolean)
          .slice(0, 4)
        if (classes.length) part += `.${classes.join('.')}`

        parts.unshift(part)
        node = node.parentElement
      }

      return parts.join(' > ')
    }

    const codePointLabel = (character: string): string => {
      const codePoint = character.codePointAt(0) ?? 0
      return `U+${codePoint.toString(16).toUpperCase().padStart(4, '0')}`
    }

    const matchingRangeLabels = (text: string): string[] => {
      const labels = new Set<string>()

      for (const character of text) {
        const codePoint = character.codePointAt(0)
        if (codePoint === undefined) continue

        for (const spec of rangeSpecs) {
          if (spec.ranges.some(([start, end]) => codePoint >= start && codePoint <= end)) {
            labels.add(spec.label)
          }
        }
      }

      return [...labels].sort()
    }

    const renderedOptionalShardText: Array<{
      labels: string[]
      text: string
      characters: string[]
      path: string
      computedFamily: string
      lang: string
      rectCount: number
    }> = []
    const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT, {
      acceptNode(node) {
        const parent = node.parentElement
        if (!parent || ['SCRIPT', 'STYLE', 'NOSCRIPT', 'TEMPLATE'].includes(parent.tagName)) {
          return NodeFilter.FILTER_REJECT
        }

        return matchingRangeLabels(node.nodeValue || '').length > 0
          ? NodeFilter.FILTER_ACCEPT
          : NodeFilter.FILTER_SKIP
      },
    })

    while (walker.nextNode()) {
      const node = walker.currentNode
      const parent = node.parentElement
      if (!parent) continue

      const text = (node.nodeValue || '').replace(/\s+/g, ' ').trim()
      if (!text) continue

      const range = document.createRange()
      range.selectNodeContents(node)
      const rectCount = range.getClientRects().length
      range.detach()
      if (rectCount === 0) continue

      const labels = matchingRangeLabels(text)
      const characters = [...new Set([...text]
        .filter(character => matchingRangeLabels(character).length > 0)
        .map(character => `${character} ${codePointLabel(character)}`))]

      renderedOptionalShardText.push({
        labels,
        text: text.slice(0, 160),
        characters,
        path: pathFor(parent),
        computedFamily: getComputedStyle(parent).fontFamily,
        lang: parent.closest('[lang]')?.getAttribute('lang') || document.documentElement.lang || '',
        rectCount,
      })
    }

    return {
      htmlLang: document.documentElement.lang,
      bodyFont: getComputedStyle(document.body).fontFamily,
      resourceFontPaths: performance.getEntriesByType('resource')
        .map(entry => new URL(entry.name).pathname)
        .filter(path => path.startsWith('/fonts/MapleUI-') && path.endsWith('.woff2')),
      renderedOptionalShardText,
    }
  }, { rangeSpecs: optionalHomepageFontRangeSpecs })

  const loadedFontPaths = [...new Set([...requestedFontPaths, ...report.resourceFontPaths])].sort()

  expect(report.htmlLang).toBe('en-US')
  expect(report.bodyFont.replace(/\s+/g, '')).toBe('MapleUILatin,MapleUICJK')
  expect(report.renderedOptionalShardText).toEqual([])
  expect(loadedFontPaths).toContain(latinFontPath)
  expect(loadedFontPaths.filter(path => !approvedStorefrontFontPaths.includes(path))).toEqual([])
  expect(loadedFontPaths.filter(path => nonDefaultEnglishHomepageFontPaths.includes(path))).toEqual([])
})

test('the Latin subset cannot shift layout against the complete Maple UI face', async ({ page }) => {
  await page.goto(storefrontUrl)

  const result = await page.evaluate(async ({
    latinFontPath,
    cjkFontPath,
    expectedFontDisplayByFamily,
  }) => {
    const sample = 'Factory-Direct Carbon Rims & Wheelsets 2026 $99.95'

    const loadFont = async (family: string, sourcePath: string) => {
      const response = await fetch(sourcePath)
      if (!response.ok) {
        throw new Error(`Unable to load ${sourcePath}: ${response.status}`)
      }

      const face = new FontFace(family, await response.arrayBuffer(), {
        display: 'block',
        style: 'normal',
        weight: '100 900',
      })

      document.fonts.add(await face.load())
    }

    const measure = (family: string) => {
      const element = document.createElement('span')
      element.textContent = sample
      Object.assign(element.style, {
        display: 'inline-block',
        fontFamily: family,
        fontSize: '48px',
        fontStyle: 'normal',
        fontWeight: '400',
        lineHeight: 'normal',
        position: 'absolute',
        visibility: 'hidden',
        whiteSpace: 'pre',
      })
      document.body.append(element)

      const bounds = element.getBoundingClientRect()
      element.remove()

      const canvas = document.createElement('canvas')
      canvas.width = 1800
      canvas.height = 128
      const context = canvas.getContext('2d')
      if (!context) throw new Error('Canvas 2D context is unavailable.')

      context.font = `400 48px "${family}"`
      context.textBaseline = 'alphabetic'
      context.fillText(sample, 8, 72)

      const textMetrics = context.measureText(sample)
      const pixels = context.getImageData(0, 0, canvas.width, canvas.height).data
      let pixelHash = 2166136261

      for (const pixel of pixels) {
        pixelHash = Math.imul(pixelHash ^ pixel, 16777619)
      }

      return {
        width: bounds.width,
        height: bounds.height,
        textWidth: textMetrics.width,
        fontBoundingBoxAscent: textMetrics.fontBoundingBoxAscent,
        fontBoundingBoxDescent: textMetrics.fontBoundingBoxDescent,
        actualBoundingBoxAscent: textMetrics.actualBoundingBoxAscent,
        actualBoundingBoxDescent: textMetrics.actualBoundingBoxDescent,
        pixelHash: pixelHash >>> 0,
      }
    }

    await Promise.all([
      loadFont('StorefrontFontContractLatin', latinFontPath),
      loadFont('StorefrontFontContractCjk', cjkFontPath),
    ])
    await document.fonts.ready

    const fontDisplayByFamily: Record<string, string> = {}
    for (const stylesheet of document.styleSheets) {
      let rules: CSSRuleList

      try {
        rules = stylesheet.cssRules
      } catch {
        continue
      }

      for (const rule of rules) {
        if (rule.type !== CSSRule.FONT_FACE_RULE) continue

        const fontFaceRule = rule as CSSFontFaceRule
        const family = fontFaceRule.style.fontFamily.replaceAll('"', '').replaceAll("'", '')
        if (family in expectedFontDisplayByFamily) {
          fontDisplayByFamily[family] = fontFaceRule.style.fontDisplay
        }
      }
    }

    return {
      bodyFont: getComputedStyle(document.body).fontFamily,
      fontDisplayByFamily,
      latin: measure('StorefrontFontContractLatin'),
      cjk: measure('StorefrontFontContractCjk'),
    }
  }, { latinFontPath, cjkFontPath, expectedFontDisplayByFamily })

  expect(result.bodyFont).toBe('MapleUILatin, MapleUICJK')
  expect(result.fontDisplayByFamily).toEqual(expectedFontDisplayByFamily)
  expect(result.latin).toEqual(result.cjk)
})

test('unicode-range segmentation loads only the font shards required by rendered text', async ({ page }) => {
  const asciiFonts = await collectLoadedFontPaths(
    page,
    'Factory Direct Carbon Wheelsets 2026 USD 99.95',
    '"MapleUILatin", "MapleUICJK"',
  )
  expect(asciiFonts).toContain(latinFontPath)
  expect(asciiFonts).not.toContain(cjkFontPath)
  expect(asciiFonts).not.toContain(latinAccentFontPath)

  const cjkFonts = await collectLoadedFontPaths(
    page,
    '碳纤维轮组',
    '"MapleUICJK"',
  )
  expect(cjkFonts).toContain(cjkFontPath)

  const accentFonts = await collectLoadedFontPaths(
    page,
    'Français Español Österreich',
    '"MapleUICoverageNotoSansLatinAccents"',
  )
  expect(accentFonts).toContain(latinAccentFontPath)
  expect(accentFonts).not.toContain(cjkFontPath)
})

test('the deployed font preflight manifest enforces the built-in font shard release gate', async ({ request }) => {
  const response = await request.get(new URL('/_internal/font-preflight.json', storefrontUrl).toString())
  expect(response.ok()).toBeTruthy()

  const report = await response.json()
  expect(report).toMatchObject({
    schema_version: 1,
    project: 'tanzanite-theme storefront',
    overall_status: 'pass',
    baseline: {
      id: 'storefront-built-in-font-shards-v1',
      font_display: 'block',
    },
    strategy: {
      status: 'pass',
      default_stack: ['MapleUILatin', 'MapleUICJK'],
      maple_ui_cjk_family: 'MapleUICJK',
      coverage_source_faces: ['Noto Sans Devanagari', 'Noto Sans', 'Noto Sans Arabic', 'Noto Sans Thai'],
      layout_parity_verified: true,
    },
    coverage: {
      missing_characters: 0,
    },
  })
  expect(report.strategy.latin_bytes).toBeLessThanOrEqual(report.strategy.latin_budget_bytes)
  expect(report.checks.map((check: { key: string; status: string }) => [check.key, check.status])).toEqual([
    ['no-external-system-fonts', 'pass'],
    ['font-face-contract', 'pass'],
    ['multilingual-split', 'pass'],
    ['layout-parity', 'pass'],
    ['subset-completeness', 'pass'],
  ])
  expect(report.faces).toHaveLength(6)
  expect(report.faces.every((face: { self_hosted: boolean; font_display: string }) => (
    face.self_hosted && face.font_display === 'block'
  ))).toBeTruthy()
  expect(report.coverage.locales.length).toBe(report.coverage.locale_count)
  expect(report.coverage.locales.every((locale: { status: string; missing_characters: number }) => (
    locale.status === 'pass' && locale.missing_characters === 0
  ))).toBeTruthy()
})
