import { expect, test, type Page } from '@playwright/test'
import fs from 'node:fs'
import { expectedStorefrontFontFaceFiles, originalStorefrontCjkFontSourcePath, parseStorefrontFontUnicodeRangeCodePoints, storefrontCjkCharacterShardManifest } from '../scripts/storefront-font-character-shard-contract'

const storefrontUrl = process.env.STOREFRONT_URL || 'http://127.0.0.1:9199/'
const latinFontPath = '/fonts/MapleUI-Latin.00af3fec5b34.woff2'
const cjkFontPaths = storefrontCjkCharacterShardManifest.shards.map(shard => `/fonts/${shard.filename}`)
const originalFontReferenceTestRoute = '/__font-contract-original-reference.woff2'
const latinAccentFontPath = '/fonts/MapleUI-Coverage-NotoSans-Latin-Accents.e645edc952b6.woff2'
const arabicFontPath = '/fonts/MapleUI-Coverage-NotoSans-Arabic.ce85091f0209.woff2'
const devanagariFontPath = '/fonts/MapleUI-Coverage-NotoSans-Devanagari.3b3cae4d2600.woff2'
const thaiFontPath = '/fonts/MapleUI-Coverage-NotoSans-Thai.1f5a173641bb.woff2'
const firstViewportCjkFontPreloadPathsByLocale = {
  zh_cn: [
    '/fonts/MapleUI-CJK-Han-Ideographs-Common-Storefront-Interface-Character-Group-001.b9c988b34e22.woff2',
    '/fonts/MapleUI-CJK-Han-Ideographs-Common-Storefront-Interface-Character-Group-002.d78176f77521.woff2',
    '/fonts/MapleUI-CJK-Kana-And-CJK-Punctuation-Common-Storefront-Interface-Character-Group-001.809b9143c728.woff2',
  ],
  ja: [
    '/fonts/MapleUI-CJK-Han-Ideographs-Common-Storefront-Interface-Character-Group-001.b9c988b34e22.woff2',
    '/fonts/MapleUI-CJK-Kana-And-CJK-Punctuation-Common-Storefront-Interface-Character-Group-001.809b9143c728.woff2',
  ],
  ko: [
    '/fonts/MapleUI-CJK-Korean-Hangul-Common-Storefront-Interface-Character-Group-001.cb1c1a092210.woff2',
  ],
  ru: [
    '/fonts/MapleUI-CJK-Greek-And-Cyrillic-Common-Storefront-Interface-Character-Group-001.c9869ab43163.woff2',
  ],
} as const
const approvedStorefrontFontPaths = [
  latinFontPath,
  ...cjkFontPaths,
  latinAccentFontPath,
  arabicFontPath,
  devanagariFontPath,
  thaiFontPath,
]
const nonDefaultEnglishHomepageFontPaths = [
  ...cjkFontPaths,
  arabicFontPath,
  devanagariFontPath,
  thaiFontPath,
]
const completeMapleUiCjkUnicodeRanges = storefrontCjkCharacterShardManifest.shards.flatMap(shard => (
  shard.unicodeRange.split(',').map((declaration) => {
    const match = declaration.trim().match(/^U\+([0-9A-F]+)(?:-([0-9A-F]+))?$/i)
    if (!match) throw new Error(`Invalid generated CJK unicode range: ${declaration}`)
    return [Number.parseInt(match[1], 16), Number.parseInt(match[2] || match[1], 16)] as [number, number]
  })
))
const optionalHomepageFontRangeSpecs = [
  {
    label: 'CJK',
    ranges: completeMapleUiCjkUnicodeRanges,
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

test('the English storefront home page only requests its approved Latin shards', async ({ page }) => {
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
  expect(report.bodyFont.replace(/\s+/g, '')).toBe('MapleUICoverageNotoSansLatinAccents,MapleUILatin,MapleUICJK')
  expect(report.renderedOptionalShardText.filter(entry => entry.labels.some(label => label !== 'LatinAccents'))).toEqual([])
  expect(loadedFontPaths).toContain(latinFontPath)
  expect(loadedFontPaths.filter(path => !approvedStorefrontFontPaths.includes(path))).toEqual([])
  expect(loadedFontPaths.filter(path => nonDefaultEnglishHomepageFontPaths.includes(path))).toEqual([])
})

test('the homepage preloads the active first-paint shard', async ({ page }) => {
  await page.goto(storefrontUrl, { waitUntil: 'domcontentloaded' })

  const fontPreloads = await page.evaluate(() => (
    [...document.querySelectorAll<HTMLLinkElement>('link[rel="preload"][as="font"]')].map(link => ({
      pathname: new URL(link.href).pathname,
      rel: link.rel,
      as: link.getAttribute('as'),
      type: link.getAttribute('type'),
      crossOrigin: link.getAttribute('crossorigin'),
    }))
  ))

  expect(fontPreloads).toEqual([
    {
      pathname: latinFontPath,
      rel: 'preload',
      as: 'font',
      type: 'font/woff2',
      crossOrigin: 'anonymous',
    },
    { pathname: latinAccentFontPath, rel: 'preload', as: 'font', type: 'font/woff2', crossOrigin: 'anonymous' },
  ])
})

test('CJK storefront locales preload only their first-viewport common character shards', async ({ page }) => {
  const latinAccentFontPath = '/fonts/MapleUI-Coverage-NotoSans-Latin-Accents.e645edc952b6.woff2'

  for (const [locale, expectedCjkPaths] of Object.entries(firstViewportCjkFontPreloadPathsByLocale)) {
    await page.goto(new URL(`/${locale}`, storefrontUrl).toString(), { waitUntil: 'domcontentloaded' })

    const preloadedPaths = await page.evaluate(() => (
      [...document.querySelectorAll<HTMLLinkElement>('link[rel="preload"][as="font"]')]
        .map(link => new URL(link.href).pathname)
        .sort()
    ))
    const expectedPaths = [latinFontPath, ...expectedCjkPaths]
    if (locale === 'zh_cn') expectedPaths.push(latinAccentFontPath)

    expect(preloadedPaths, locale).toEqual(expectedPaths.sort())
  }
})

test('the Latin subset cannot shift layout against the complete Maple UI face', async ({ page }) => {
  await page.route(`**${originalFontReferenceTestRoute}`, route => route.fulfill({ contentType: 'font/woff2', body: fs.readFileSync(originalStorefrontCjkFontSourcePath) }))
  await page.goto(storefrontUrl)

  const result = await page.evaluate(async ({
    latinFontPath,
    originalFontReferenceTestRoute,
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

    const measure = (family: string, text = sample) => {
      const element = document.createElement('span')
      element.textContent = text
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

      context.font = `400 48px ${family}`
      context.textBaseline = 'alphabetic'
      context.fillText(text, 8, 72)

      const textMetrics = context.measureText(text)
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
      loadFont('StorefrontFontContractCjk', originalFontReferenceTestRoute),
    ])
    await document.fonts.ready

    const characterShardSamples = []
    for (const text of ['碳纤维轮组，技术与售后。', '日本語ホイール・技術', '한국어 휠셋 기술', 'Русский Карбон', '齉龘', '─⿰𝔸𝛂\uE000']) {
      await document.fonts.load('400 48px MapleUILatin, MapleUICJK', text)
      characterShardSamples.push({
        text,
        original: measure('StorefrontFontContractCjk', text),
        shards: measure('MapleUILatin, MapleUICJK', text),
      })
    }

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
      characterShardSamples,
    }
  }, { latinFontPath, originalFontReferenceTestRoute, expectedFontDisplayByFamily })

  expect(result.bodyFont).toBe('MapleUICoverageNotoSansLatinAccents, MapleUILatin, MapleUICJK')
  expect(result.fontDisplayByFamily).toEqual(expectedFontDisplayByFamily)
  expect(result.latin).toEqual(result.cjk)
  for (const sample of result.characterShardSamples) expect(sample.shards, sample.text).toEqual(sample.original)
})

test('unicode-range segmentation loads only the font shards required by rendered text', async ({ page }) => {
  const asciiFonts = await collectLoadedFontPaths(
    page,
    'Factory Direct Carbon Wheelsets 2026 USD 99.95',
    '"MapleUILatin", "MapleUICJK"',
  )
  expect(asciiFonts).toContain(latinFontPath)
  expect(asciiFonts.filter(fontPath => cjkFontPaths.includes(fontPath))).toEqual([])
  expect(asciiFonts).not.toContain(latinAccentFontPath)

  for (const sample of ['碳纤维轮组', '快速查看常见问题答案', '日本語ホイール', '한국어 휠셋', 'Русский', '齉龘']) {
    const cjkFonts = await collectLoadedFontPaths(page, sample, '"MapleUICJK"')
    const expectedPaths = storefrontCjkCharacterShardManifest.shards.filter(shard => {
      const covered = parseStorefrontFontUnicodeRangeCodePoints(shard.unicodeRange)
      return [...sample].some(character => covered.has(character.codePointAt(0)!))
    }).map(shard => `/fonts/${shard.filename}`).sort()
    expect(cjkFonts, sample).toEqual(expectedPaths)
    if (sample === '快速查看常见问题答案') {
      expect(cjkFonts.every(fontPath => fontPath.includes('Common-Storefront-Interface'))).toBe(true)
    }
    expect(cjkFonts.some(fontPath => /MapleUI-CJK\.[a-f0-9]+\.woff2$/.test(fontPath))).toBe(false)
  }

  const accentFonts = await collectLoadedFontPaths(
    page,
    'Français Español Österreich',
    '"MapleUICoverageNotoSansLatinAccents"',
  )
  expect(accentFonts).toContain(latinAccentFontPath)
  expect(accentFonts.filter(fontPath => cjkFontPaths.includes(fontPath))).toEqual([])
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
  expect(report.faces).toHaveLength(expectedStorefrontFontFaceFiles.length)
  expect(report.strategy.cjk_character_shard_count).toBe(storefrontCjkCharacterShardManifest.shards.length)
  expect(report.faces.every((face: { self_hosted: boolean; font_display: string }) => (
    face.self_hosted && face.font_display === 'block'
  ))).toBeTruthy()
  expect(report.coverage.locales.length).toBe(report.coverage.locale_count)
  expect(report.coverage.locales.every((locale: { status: string; missing_characters: number }) => (
    locale.status === 'pass' && locale.missing_characters === 0
  ))).toBeTruthy()
})
