import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import vuePlugin from '@vitejs/plugin-vue'
import { renderToString } from '@vue/server-renderer'
import { parseFragment, type DefaultTreeAdapterMap } from 'parse5'
import { createServer } from 'vite'
import { createSSRApp, defineComponent, h, type Component } from 'vue'
import { createI18n } from 'vue-i18n'
import type {
  TireRimReverseSuggestion,
  TireRimSuggestion,
} from '../app/composables/useTireRimWidthReferenceRecommendation.js'
import {
  mapTireRimBackendBestFitmentForPresentation,
  type TireRimApplicationProfile,
  type TireRimStandardReference,
} from '../app/data/tireguides/tireRimWidthReferencePresentation.js'

const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const pageMessages = JSON.parse(await readFile(
  path.join(frontendDirectory, 'app/i18n/page-messages/guidesTireRimHelper/zh_cn.json'),
  'utf8',
))

// Use the real Vue component and translations. Replace Nuxt's routing and
// page-message context because this SSR app runs outside the Nuxt runtime.
const componentCompilationServer = await createServer({
  root: frontendDirectory,
  configFile: false,
  plugins: [
    vuePlugin(),
    {
      name: 'provide-nuxt-translation-context-for-tire-rim-rendering-tests',
      enforce: 'pre',
      resolveId(source) {
        if (source === '#imports') return '\0tire-rim-test-i18n-imports'
        const normalizedSource = source.replaceAll('\\', '/')
        if (normalizedSource === '~/composables/usePageMessages'
          || /\/app\/composables\/usePageMessages(?:\.ts)?$/.test(normalizedSource)) {
          return '\0tire-rim-test-page-message-loader'
        }
      },
      load(moduleId) {
        if (moduleId === '\0tire-rim-test-i18n-imports') {
          return 'export { useI18n } from "vue-i18n"; export const useLocalePath = () => path => "/zh_cn" + path'
        }
        if (moduleId === '\0tire-rim-test-page-message-loader') {
          return 'export const usePageMessages = () => ({ loadPageMessages: async () => {} })'
        }
      },
    },
  ],
  resolve: { alias: { '~': path.join(frontendDirectory, 'app') } },
  server: { middlewareMode: true, hmr: false, ws: false, watch: null },
})

type ParsedHtmlNode = DefaultTreeAdapterMap['node']
type ParsedHtmlElement = DefaultTreeAdapterMap['element']

function findRenderedElementsWithClass(node: ParsedHtmlNode, className: string): ParsedHtmlElement[] {
  const matches: ParsedHtmlElement[] = []
  if ('tagName' in node && node.attrs.some(attribute => (
    attribute.name === 'class' && attribute.value.split(/\s+/).includes(className)
  ))) matches.push(node)
  if ('childNodes' in node) {
    for (const child of node.childNodes) matches.push(...findRenderedElementsWithClass(child, className))
  }
  return matches
}

function readRenderedElementText(node: ParsedHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  if ('childNodes' in node) return node.childNodes.map(readRenderedElementText).join('')
  return ''
}

function readSingleRenderedSectionText(fragment: ParsedHtmlNode, className: string): string {
  const sections = findRenderedElementsWithClass(fragment, className)
  assert.equal(sections.length, 1, `Expected exactly one ${className} section`)
  return readRenderedElementText(sections[0]!).replace(/\s+/g, ' ').trim()
}

function createStandardReferenceFixture(): TireRimStandardReference {
  return {
    iso_5775_part_1: 'ISO 5775-1:2023',
    iso_5775_part_2: 'ISO 5775-2:2021',
    etrto_edition: '',
    dt_swiss_used_by_calculation: false,
    dt_swiss_reference_role: 'display_only',
  }
}

function createAppliedISONormalFitmentReferenceFixture() {
  return {
    standard: 'ISO 5775-1:2023', clause: '4.4', table: '7',
    source_kind: 'user_provided_standard_width_matrix_image',
    scope: 'nominal_width_pairing_only',
  }
}

function createAppliedISONormalRuleReferenceFixture() {
  return {
    normal_range_basis: 'iso_5775_1_2023_clause_4_4_table_7',
    recommended_range_basis: '', manufacturer_source_urls: [] as string[], interpolated: false,
  }
}

// These are representative backend-contract fixtures, not calculation rules.
// The render tests deliberately keep legacy ranges and targets distinct from
// best_fitment, so a future display regression cannot mistake them for an answer.
function createKnownHookedMountainTireDisplayFixture(): TireRimSuggestion {
  return {
    isEngineeringRecommended: false,
    isEngineeringReference: false,
    aeroRuleApplies: false,
    rimWidthRanges: [{ min: 21, max: 35 }],
    recommendedRimWidthRanges: [{ min: 28, max: 33 }],
    normalFitmentRimWidthRanges: [{ min: 21, max: 35 }],
    bestFitment: mapTireRimBackendBestFitmentForPresentation({
      rim_inner_width_mm: 31,
      tire_width_mm: 60,
      application_profile: 'mtb_xc_trail',
      selection_method: 'recommended_window_center',
      reason_code: 'nearest_recommended_window_center',
      model_status: 'pending_profile_calibration',
      target_tire_width_mm: 61.5,
      target_tire_width_difference_mm: 1.5,
      fitment_rule_reference: {
        normal_range_basis: 'project_envelope',
        recommended_range_basis: 'reviewed_manufacturer_optimal_bands_project_envelope',
        manufacturer_source_urls: ['https://www.wtb.com/pages/tire-rim-fit-chart'],
        interpolated: false,
      },
    }),
    applicationProfile: 'mtb_xc_trail',
    applicationProfileSelection: 'automatic_by_tire_width',
    standardReference: createStandardReferenceFixture(),
    normalFitmentReference: createAppliedISONormalFitmentReferenceFixture(),
    physical: {
      tireWidth: { min: 60, max: 60 },
      aeroTargetOuterWidth: null,
      rimOuterWidthEstimates: [],
    },
    modelVersion: 'tire-rim-engineering-reference-v11',
    knowledgeAsOf: '2026-10-11',
    engineering: null,
  }
}

function createKnownHooklessRimDisplayFixture(
  profile: TireRimApplicationProfile = 'mtb_xc_trail',
): TireRimReverseSuggestion {
  const plusPurpose = profile === 'mtb_plus'
  const tireWidth = plusPurpose ? 71 : 67
  const ruleReference = {
    normal_range_basis: 'project_envelope_including_reviewed_model_guidance',
    recommended_range_basis: 'reviewed_manufacturer_guidance_project_window',
    manufacturer_source_urls: plusPurpose
      ? ['https://www.stans.com/']
      : ['https://enve.com/pages/tire-compatibility'],
    interpolated: false,
  }
  return {
    rimInnerWidthMm: 35,
    rimSystem: 'hookless',
    bestFitment: mapTireRimBackendBestFitmentForPresentation({
      rim_inner_width_mm: 35,
      tire_width_mm: tireWidth,
      application_profile: profile,
      selection_method: 'recommended_window_center',
      reason_code: 'nearest_integer_recommended_tire_window_center',
      model_status: 'pending_profile_calibration',
      target_tire_width_mm: plusPurpose ? 71.12 : 67.31,
      target_tire_width_difference_mm: plusPurpose ? 0.12 : 0.31,
      fitment_rule_reference: ruleReference,
    }),
    applicationProfile: profile,
    resolvedApplicationProfile: profile,
    applicationProfileSelection: 'explicit',
    standardReference: createStandardReferenceFixture(),
    normalFitmentReference: createAppliedISONormalFitmentReferenceFixture(),
    fitmentRuleReference: createAppliedISONormalRuleReferenceFixture(),
    aeroRuleApplies: false,
    aeroTargetTireWidthMm: null,
    bestMatch: null,
    normalFitmentTireWidthRange: { min: 58, max: 83 },
    recommendedFitmentTireWidthRange: plusPurpose ? { min: 66, max: 76 } : { min: 64, max: 71 },
    rimOuterWidthEstimate: {
      rimInnerWidthMm: 35,
      wallThicknessPerSideEstimateMm: 3,
      rimOuterWidthEstimateMm: 41,
    },
    modelVersion: 'tire-rim-engineering-reference-v11',
    knowledgeAsOf: '2026-10-11',
  }
}

function createEnteredHooklessMountainPairingDisplayFixture(): TireRimSuggestion {
  const response = createKnownHooklessRimDisplayFixture()
  return {
    ...createKnownHookedMountainTireDisplayFixture(),
    applicationProfileSelection: 'explicit',
    bestFitment: response.bestFitment,
    // This range describes the entered 84 mm tire; the exact-pairing normal
    // section must instead use the fixed 35 mm rim's allowed tire-width range.
    normalFitmentRimWidthRanges: [{ min: 41, max: 64 }],
    physical: {
      tireWidth: { min: 84, max: 84 },
      aeroTargetOuterWidth: null,
      rimOuterWidthEstimates: [response.rimOuterWidthEstimate],
    },
    engineering: {
      modelVersion: response.modelVersion,
      verdict: 'critical',
      verdictReason: 'above_maximum',
      rimSystem: 'hookless',
      rimInnerWidthMm: 35,
      tireWidthMm: 84,
      applicationProfile: 'mtb_xc_trail',
      applicationProfileSelection: 'explicit',
      standardReference: response.standardReference,
      fitmentRuleReference: response.fitmentRuleReference,
      normalFitmentReference: response.normalFitmentReference,
      allowedTireWidthRange: response.normalFitmentTireWidthRange,
      recommendedTireWidthRange: response.recommendedFitmentTireWidthRange,
      aeroRuleApplies: false,
      aeroTargetOuterWidthMm: null,
      aeroTargetTireWidthMm: null,
      maximumPressure: { bar: 3, psi: 43.5 },
      tubeRecommendation: { band: 'wide', widthMinMm: 60, widthMaxMm: 80 },
    },
  }
}

let renderedRegressionCaseCount = 0

try {
  const compiledComponent = (await componentCompilationServer.ssrLoadModule(
    '/app/components/tireguides/TireRimLayeredFitmentResults.vue',
  )).default as Component

  async function renderLayeredFitmentWithRealChineseMessages(
    suggestion: TireRimSuggestion | TireRimReverseSuggestion,
  ): Promise<ParsedHtmlNode> {
    const capturedVueErrors: unknown[] = []
    const application = createSSRApp({ render: () => h(compiledComponent, { suggestion }) })
    application.component('NuxtLink', defineComponent({
      props: { to: { type: String, required: true } },
      setup(props, { slots }) {
        return () => h('a', { href: props.to }, slots.default?.())
      },
    }))
    application.config.errorHandler = error => { capturedVueErrors.push(error) }
    application.use(createI18n({
      legacy: false,
      locale: 'zh_cn',
      messages: { zh_cn: { guidesTireRimHelper: pageMessages } },
      missingWarn: false,
      fallbackWarn: false,
    }))
    const markup = await renderToString(application)
    assert.deepEqual(capturedVueErrors, [], 'The actual result component failed to render')
    const fragment = parseFragment(markup)
    assert.equal(findRenderedElementsWithClass(fragment, 'tire-rim-layered-results').length, 1)
    const resultCards = findRenderedElementsWithClass(fragment, 'tire-rim-layered-results')[0]!.childNodes
      .filter((node): node is ParsedHtmlElement => 'tagName' in node && node.tagName === 'section')
    assert.equal(resultCards.length, 2, 'Only the best-width and normal-range cards should be separate')
    const normalRangeCard = findRenderedElementsWithClass(fragment, 'tire-rim-layered-results__normal')[0]!
    const referenceBlock = findRenderedElementsWithClass(fragment, 'tire-rim-layered-results__reference')[0]!
    assert.equal(referenceBlock.tagName, 'div', 'Sources must not create a third result card')
    assert.equal(findRenderedElementsWithClass(normalRangeCard, 'tire-rim-layered-results__reference')[0], referenceBlock,
      'The standards and normal-range sources must belong to the normal-range card')
    assert.doesNotMatch(readSingleRenderedSectionText(fragment, 'tire-rim-layered-results__best'), /ISO 5775|ETRTO/)
    const renderedText = readRenderedElementText(fragment)
    assert.doesNotMatch(renderedText, /DT Swiss|dt_swiss|wangguangen|王关根|TANZANITE/)
    assert.doesNotMatch(renderedText, /pending_profile_calibration|recommended_window_center|tire-rim-engineering-reference-v11|user_provided_standard_width_matrix_image/)
    const clearanceLinks = findRenderedElementsWithClass(fragment, 'tire-rim-layered-results__clearance-link')
    assert.equal(clearanceLinks.length, 1, 'Each result should link to the existing frame-clearance guide')
    assert.equal(clearanceLinks[0]!.tagName, 'a')
    assert.equal(clearanceLinks[0]!.attrs.find(attribute => attribute.name === 'href')?.value,
      '/zh_cn/guides/tireguides/tire-frame-clearance', 'The clearance link must stay within the selected language')
    assert.equal((markup.match(/<a\s/g) ?? []).length, 1, 'Standards references must not add external navigation')
    renderedRegressionCaseCount++
    return fragment
  }

  const knownTire = await renderLayeredFitmentWithRealChineseMessages(createKnownHookedMountainTireDisplayFixture())
  assert.equal(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__value'), '31mm')
  assert.match(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__best'), /最合适的车圈内宽/)
  assert.doesNotMatch(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__best'), /28-45|28-33|105%/)
  assert.match(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__normal'), /车圈内宽范围\s*21-35 mm/)
  assert.match(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__reference'), /ISO 5775-1:2023 · 第 4.4 节 · 表 7/)
  assert.match(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__reference'), /按表内的名义胎宽与车圈内宽查得本卡片的正常配合范围/)
  assert.doesNotMatch(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__reference'), /尚未|ETRTO/)
  assert.match(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__best'), /最佳宽度参考：WTB/)
  assert.doesNotMatch(readSingleRenderedSectionText(knownTire, 'tire-rim-layered-results__reference'), /WTB/)

  for (const [profile, expectedWidth, expectedRange, expectedSource] of [
    ['mtb_xc_trail', '67mm', '58-83 mm', 'ENVE'],
    ['mtb_plus', '71mm', '58-83 mm', 'Stan’s'],
  ] as const) {
    const knownRim = await renderLayeredFitmentWithRealChineseMessages(createKnownHooklessRimDisplayFixture(profile))
    assert.equal(readSingleRenderedSectionText(knownRim, 'tire-rim-layered-results__value'), expectedWidth)
    assert.match(readSingleRenderedSectionText(knownRim, 'tire-rim-layered-results__best'), /最合适的外胎宽度/)
    assert.ok(readSingleRenderedSectionText(knownRim, 'tire-rim-layered-results__normal').includes(expectedRange))
    assert.ok(readSingleRenderedSectionText(knownRim, 'tire-rim-layered-results__best').includes(expectedSource))
    assert.equal(readSingleRenderedSectionText(knownRim, 'tire-rim-layered-results__reference').includes(expectedSource), false)
    assert.doesNotMatch(readRenderedElementText(knownRim), /105%/)
  }

  const separateBestAndNormalSources = createKnownHooklessRimDisplayFixture()
  // Deliberately distinct source fields test renderer isolation; production
  // normal ranges cite ISO alone and do not copy best-fitment manufacturers.
  separateBestAndNormalSources.fitmentRuleReference.manufacturer_source_urls = ['https://enve.com/pages/tire-compatibility']
  separateBestAndNormalSources.bestFitment = {
    ...separateBestAndNormalSources.bestFitment!,
    fitmentRuleReference: {
      ...separateBestAndNormalSources.bestFitment!.fitmentRuleReference,
      manufacturer_source_urls: ['https://www.wtb.com/pages/tire-rim-fit-chart'],
    },
  }
  const independentSources = await renderLayeredFitmentWithRealChineseMessages(separateBestAndNormalSources)
  assert.match(readSingleRenderedSectionText(independentSources, 'tire-rim-layered-results__best'), /最佳宽度参考：WTB/)
  assert.doesNotMatch(readSingleRenderedSectionText(independentSources, 'tire-rim-layered-results__best'), /ENVE/)
  assert.match(readSingleRenderedSectionText(independentSources, 'tire-rim-layered-results__reference'), /配合尺寸资料：ENVE/)
  assert.doesNotMatch(readSingleRenderedSectionText(independentSources, 'tire-rim-layered-results__reference'), /WTB/)
  assert.match(readSingleRenderedSectionText(independentSources, 'tire-rim-layered-results__normal'), /58-83 mm/)

  const exactPairing = await renderLayeredFitmentWithRealChineseMessages(createEnteredHooklessMountainPairingDisplayFixture())
  assert.equal(readSingleRenderedSectionText(exactPairing, 'tire-rim-layered-results__value'), '67mm')
  assert.match(readSingleRenderedSectionText(exactPairing, 'tire-rim-layered-results__best'), /车圈内宽 35 mm · 胎宽 67 mm/)
  assert.match(readSingleRenderedSectionText(exactPairing, 'tire-rim-layered-results__normal'), /外胎宽度范围\s*58-83 mm/)
  assert.match(readSingleRenderedSectionText(exactPairing, 'tire-rim-layered-results__entered'), /车圈内宽 35 mm · 胎宽 84 mm/)
  assert.match(readSingleRenderedSectionText(exactPairing, 'tire-rim-layered-results__entered'), /超出正常配合范围上限 83 mm/)
  assert.equal(findRenderedElementsWithClass(exactPairing, 'tire-rim-layered-results__entered--critical').length, 1)

  const noFatRecommendation = createKnownHooklessRimDisplayFixture('fat_snow')
  noFatRecommendation.bestFitment = null
  noFatRecommendation.rimInnerWidthMm = 76
  noFatRecommendation.normalFitmentTireWidthRange = { min: 96, max: 132 }
  noFatRecommendation.recommendedFitmentTireWidthRange = null
  noFatRecommendation.rimOuterWidthEstimate = {
    rimInnerWidthMm: 76, wallThicknessPerSideEstimateMm: 3, rimOuterWidthEstimateMm: 82,
  }
  // Deliberately keep the retired display target populated. It must not turn
  // a null best_fitment into the former 115 mm "best" result.
  noFatRecommendation.bestMatch = { tireWidthMm: 115, targetTireWidthDifferenceMm: 0.5 }
  noFatRecommendation.fitmentRuleReference.manufacturer_source_urls = []
  const emptyFat = await renderLayeredFitmentWithRealChineseMessages(noFatRecommendation)
  assert.equal(findRenderedElementsWithClass(emptyFat, 'tire-rim-layered-results__value').length, 0)
  assert.match(readSingleRenderedSectionText(emptyFat, 'tire-rim-layered-results__best'), /当前暂无单一推荐/)
  assert.doesNotMatch(readSingleRenderedSectionText(emptyFat, 'tire-rim-layered-results__best'), /115|105%/)
  assert.match(readSingleRenderedSectionText(emptyFat, 'tire-rim-layered-results__normal'), /96-132 mm/)

  const emptyKnownFatTire = createKnownHookedMountainTireDisplayFixture()
  emptyKnownFatTire.applicationProfile = 'fat_snow'
  emptyKnownFatTire.bestFitment = null
  emptyKnownFatTire.recommendedRimWidthRanges = []
  emptyKnownFatTire.rimWidthRanges = [{ min: 72, max: 100 }]
  emptyKnownFatTire.normalFitmentRimWidthRanges = [{ min: 72, max: 100 }]
  emptyKnownFatTire.physical.tireWidth = { min: 115, max: 115 }
  const knownFatTireWithoutBest = await renderLayeredFitmentWithRealChineseMessages(emptyKnownFatTire)
  assert.equal(findRenderedElementsWithClass(knownFatTireWithoutBest, 'tire-rim-layered-results__value').length, 0)
  assert.match(readSingleRenderedSectionText(knownFatTireWithoutBest, 'tire-rim-layered-results__best'), /最合适的车圈内宽.*当前暂无单一推荐/)
  assert.match(readSingleRenderedSectionText(knownFatTireWithoutBest, 'tire-rim-layered-results__normal'), /车圈内宽范围\s*72-100 mm/)

  const emptyEnteredFatPairing = createEnteredHooklessMountainPairingDisplayFixture()
  emptyEnteredFatPairing.applicationProfile = 'fat_snow'
  emptyEnteredFatPairing.bestFitment = null
  emptyEnteredFatPairing.recommendedRimWidthRanges = []
  emptyEnteredFatPairing.physical = {
    tireWidth: { min: 115, max: 115 },
    aeroTargetOuterWidth: null,
    rimOuterWidthEstimates: [noFatRecommendation.rimOuterWidthEstimate],
  }
  Object.assign(emptyEnteredFatPairing.engineering!, {
    verdict: 'reference', verdictReason: 'no_recommended_window',
    rimInnerWidthMm: 76, tireWidthMm: 115, applicationProfile: 'fat_snow',
    allowedTireWidthRange: { min: 96, max: 132 },
    recommendedTireWidthRange: null, maximumPressure: null,
    fitmentRuleReference: noFatRecommendation.fitmentRuleReference,
  })
  const enteredFatPairingWithoutBest = await renderLayeredFitmentWithRealChineseMessages(emptyEnteredFatPairing)
  assert.equal(findRenderedElementsWithClass(enteredFatPairingWithoutBest, 'tire-rim-layered-results__value').length, 0)
  assert.match(readSingleRenderedSectionText(enteredFatPairingWithoutBest, 'tire-rim-layered-results__entered'), /车圈内宽 76 mm · 胎宽 115 mm/)
  assert.match(readSingleRenderedSectionText(enteredFatPairingWithoutBest, 'tire-rim-layered-results__entered'), /处于正常参考范围.*暂无尺寸优化建议/)
  assert.match(readSingleRenderedSectionText(enteredFatPairingWithoutBest, 'tire-rim-layered-results__normal'), /外胎宽度范围\s*96-132 mm/)

  const reviewedFatRim = createKnownHooklessRimDisplayFixture('fat_snow')
  reviewedFatRim.rimInnerWidthMm = 85
  reviewedFatRim.bestFitment = {
    ...reviewedFatRim.bestFitment!, rimInnerWidthMm: 85, tireWidthMm: 105,
    targetTireWidthMm: 105.41, targetTireWidthDifferenceMm: 0.41,
  }
  reviewedFatRim.normalFitmentTireWidthRange = { min: 96, max: 132 }
  reviewedFatRim.recommendedFitmentTireWidthRange = { min: 97, max: 114 }
  reviewedFatRim.rimOuterWidthEstimate = { rimInnerWidthMm: 85, wallThicknessPerSideEstimateMm: 3, rimOuterWidthEstimateMm: 91 }
  const reviewedFatRecommendation = await renderLayeredFitmentWithRealChineseMessages(reviewedFatRim)
  assert.equal(readSingleRenderedSectionText(reviewedFatRecommendation, 'tire-rim-layered-results__value'), '105mm')
  assert.match(readSingleRenderedSectionText(reviewedFatRecommendation, 'tire-rim-layered-results__normal'), /96-132 mm/)
  assert.doesNotMatch(readRenderedElementText(reviewedFatRecommendation), /105%/)

  const automaticallyResolvedRim = createKnownHooklessRimDisplayFixture()
  automaticallyResolvedRim.applicationProfile = 'automatic_by_width'
  const automaticRim = await renderLayeredFitmentWithRealChineseMessages(automaticallyResolvedRim)
  assert.match(readSingleRenderedSectionText(automaticRim, 'tire-rim-layered-results__heading'), /山地 XC \/ Trail/)
  assert.doesNotMatch(readSingleRenderedSectionText(automaticRim, 'tire-rim-layered-results__heading'), /自动/)

  const blockedLegacyAeroTarget = createKnownHooklessRimDisplayFixture('road_aero')
  blockedLegacyAeroTarget.bestFitment = null
  blockedLegacyAeroTarget.rimInnerWidthMm = 19
  blockedLegacyAeroTarget.rimOuterWidthEstimate = {
    rimInnerWidthMm: 19, wallThicknessPerSideEstimateMm: 3, rimOuterWidthEstimateMm: 25,
  }
  blockedLegacyAeroTarget.normalFitmentTireWidthRange = { min: 22, max: 57 }
  blockedLegacyAeroTarget.recommendedFitmentTireWidthRange = { min: 28, max: 28 }
  blockedLegacyAeroTarget.fitmentRuleReference.manufacturer_source_urls = []
  blockedLegacyAeroTarget.aeroRuleApplies = true
  blockedLegacyAeroTarget.bestMatch = { tireWidthMm: 26, targetTireWidthDifferenceMm: 0.25 }
  blockedLegacyAeroTarget.aeroTargetTireWidthMm = 26.25
  const blockedAero = await renderLayeredFitmentWithRealChineseMessages(blockedLegacyAeroTarget)
  assert.equal(findRenderedElementsWithClass(blockedAero, 'tire-rim-layered-results__value').length, 0)
  assert.match(readSingleRenderedSectionText(blockedAero, 'tire-rim-layered-results__best'), /当前暂无单一推荐/)
  assert.match(readSingleRenderedSectionText(blockedAero, 'tire-rim-layered-results__best'), /目标胎宽约 26\.25 mm/)

  const boundaryThirty = createKnownHookedMountainTireDisplayFixture()
  boundaryThirty.applicationProfile = 'road_aero'
  boundaryThirty.normalFitmentRimWidthRanges = [{ min: 16, max: 25 }]
  boundaryThirty.aeroRuleApplies = true
  boundaryThirty.bestFitment = {
    ...boundaryThirty.bestFitment!,
    rimInnerWidthMm: 23,
    tireWidthMm: 30,
    applicationProfile: 'road_aero',
  }
  boundaryThirty.physical = {
    tireWidth: { min: 30, max: 30 },
    aeroTargetOuterWidth: { min: 28.6, max: 28.6 },
    rimOuterWidthEstimates: [{ rimInnerWidthMm: 23, wallThicknessPerSideEstimateMm: 3, rimOuterWidthEstimateMm: 29 }],
  }
  const narrowTire = await renderLayeredFitmentWithRealChineseMessages(boundaryThirty)
  assert.equal(readSingleRenderedSectionText(narrowTire, 'tire-rim-layered-results__value'), '23mm')
  assert.match(readSingleRenderedSectionText(narrowTire, 'tire-rim-layered-results__best'), /105%/)
  assert.match(readSingleRenderedSectionText(narrowTire, 'tire-rim-layered-results__best'), /两侧各约 3 mm ≈ 29 mm/)
  assert.match(readSingleRenderedSectionText(narrowTire, 'tire-rim-layered-results__best'), /28\.6 mm/)

  const boundaryThirtyOne = createKnownHookedMountainTireDisplayFixture()
  boundaryThirtyOne.applicationProfile = 'gravel_all_road'
  boundaryThirtyOne.normalFitmentRimWidthRanges = [{ min: 16, max: 25 }]
  boundaryThirtyOne.bestFitment = { ...boundaryThirtyOne.bestFitment!, rimInnerWidthMm: 25, tireWidthMm: 31, applicationProfile: 'gravel_all_road' }
  // Stale optional geometry must not be shown when the current backend result
  // explicitly says that the aero rule does not apply.
  boundaryThirtyOne.physical.aeroTargetOuterWidth = { min: 28.6, max: 28.6 }
  const wideTire = await renderLayeredFitmentWithRealChineseMessages(boundaryThirtyOne)
  assert.equal(readSingleRenderedSectionText(wideTire, 'tire-rim-layered-results__value'), '25mm')
  assert.doesNotMatch(readRenderedElementText(wideTire), /105%|28\.6/)

  const withoutReferenceMetadata = createKnownHookedMountainTireDisplayFixture()
  Reflect.deleteProperty(withoutReferenceMetadata, 'standardReference')
  Reflect.deleteProperty(withoutReferenceMetadata, 'normalFitmentReference')
  const incompleteReference = await renderLayeredFitmentWithRealChineseMessages(withoutReferenceMetadata)
  assert.equal(readSingleRenderedSectionText(incompleteReference, 'tire-rim-layered-results__value'), '31mm')
  assert.match(readSingleRenderedSectionText(incompleteReference, 'tire-rim-layered-results__normal'), /21-35 mm/)
  assert.match(readSingleRenderedSectionText(incompleteReference, 'tire-rim-layered-results__reference'), /标准参考暂未提供/)

  console.log(`Tire/rim layered rendering regression passed: ${renderedRegressionCaseCount} cases using the actual Vue component.`)
} finally {
  await componentCompilationServer.close()
}
