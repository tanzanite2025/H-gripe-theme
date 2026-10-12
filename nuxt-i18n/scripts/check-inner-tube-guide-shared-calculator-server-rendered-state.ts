import assert from 'node:assert/strict'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import { parseQuery } from 'vue-router'
import { hasInnerTubeValveFitmentShareQueryParameters } from '../app/utils/innerTubeValveFitmentShareUrlState.ts'

type InnerTubeGuideHtmlNode = DefaultTreeAdapterMap['node']
type InnerTubeGuideHtmlElement = DefaultTreeAdapterMap['element']

const storefrontBaseUrl = process.env.SEO_STOREFRONT_BASE_URL
if (!storefrontBaseUrl) throw new Error('Set SEO_STOREFRONT_BASE_URL to a running storefront origin.')

function collectServerRenderedInnerTubeGuideElements(node: InnerTubeGuideHtmlNode): InnerTubeGuideHtmlElement[] {
  const elements: InnerTubeGuideHtmlElement[] = 'tagName' in node ? [node] : []
  if ('childNodes' in node) {
    for (const child of node.childNodes) elements.push(...collectServerRenderedInnerTubeGuideElements(child))
  }
  return elements
}

function readServerRenderedInnerTubeGuideAttribute(element: InnerTubeGuideHtmlElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value ?? ''
}

function readServerRenderedInnerTubeGuideText(node: InnerTubeGuideHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readServerRenderedInnerTubeGuideText).join('') : ''
}

const manualShareParameters = 'rim_depth=80&fitment_mode=manual&valve_length=80&extender_length=30&pump_head_grip_depth=22&rim_depth_uncertainty=4'
const routeChecks = [
  { routePath: '/tools/valve-length-fitment-calculator', calculatorSelected: false, rimDepth: 50, gripDepth: 15, uncertainty: 2, manual: false },
  { routePath: '/tools/valve-length-fitment-calculator?utm_source=forum', calculatorSelected: false, rimDepth: 50, gripDepth: 15, uncertainty: 2, manual: false },
  { routePath: '/tools/valve-length-fitment-calculator?tab=calculator', calculatorSelected: true, rimDepth: 50, gripDepth: 15, uncertainty: 2, manual: false },
  { routePath: '/zh_cn/tools/valve-length-fitment-calculator?tab=calculator', calculatorSelected: true, rimDepth: 50, gripDepth: 15, uncertainty: 2, manual: false },
  { routePath: '/tools/valve-length-fitment-calculator?tab=calculator#introduction', calculatorSelected: true, rimDepth: 50, gripDepth: 15, uncertainty: 2, manual: false },
  { routePath: '/tools/valve-length-fitment-calculator?tab=invalid', calculatorSelected: false, rimDepth: 50, gripDepth: 15, uncertainty: 2, manual: false },
  { routePath: `/tools/valve-length-fitment-calculator?tab=introduction&${manualShareParameters}`, calculatorSelected: false, rimDepth: 80, gripDepth: 22, uncertainty: 4, manual: true },
  { routePath: `/tools/valve-length-fitment-calculator?tab=calculator&${manualShareParameters}`, calculatorSelected: true, rimDepth: 80, gripDepth: 22, uncertainty: 4, manual: true },
  { routePath: `/tools/valve-length-fitment-calculator?${manualShareParameters}`, calculatorSelected: true, rimDepth: 80, gripDepth: 22, uncertainty: 4, manual: true },
  { routePath: `/zh_cn/tools/valve-length-fitment-calculator?${manualShareParameters}`, calculatorSelected: true, rimDepth: 80, gripDepth: 22, uncertainty: 4, manual: true },
  { routePath: '/tools/valve-length-fitment-calculator?rim_depth=60', calculatorSelected: true, rimDepth: 60, gripDepth: 15, uncertainty: 2, manual: false },
  { routePath: '/tools/valve-length-fitment-calculator?rim_depth=%20&fitment_mode=invalid&pump_head_grip_depth=&rim_depth_uncertainty=NaN', calculatorSelected: true, rimDepth: 50, gripDepth: 15, uncertainty: 2, manual: false },
]

for (const check of routeChecks) {
  const response: Response = await fetch(new URL(check.routePath, storefrontBaseUrl), { signal: AbortSignal.timeout(60_000) })
  assert.equal(response.status, 200)
  const elements = collectServerRenderedInnerTubeGuideElements(parse(await response.text()))
  const readUniqueServerRenderedInnerTubeGuideElementById = (elementId: string): InnerTubeGuideHtmlElement => {
    const matchingElements = elements.filter(element => readServerRenderedInnerTubeGuideAttribute(element, 'id') === elementId)
    assert.equal(matchingElements.length, 1, `Expected one ${elementId}`)
    return matchingElements[0]!
  }

  assert.equal(readServerRenderedInnerTubeGuideAttribute(readUniqueServerRenderedInnerTubeGuideElementById('inner-tube-calculator-tab'), 'aria-selected'), String(check.calculatorSelected))
  assert.equal(readServerRenderedInnerTubeGuideAttribute(readUniqueServerRenderedInnerTubeGuideElementById('inner-tube-introduction-tab'), 'aria-selected'), String(!check.calculatorSelected))
  assert.equal(/display:\s*none/.test(readServerRenderedInnerTubeGuideAttribute(readUniqueServerRenderedInnerTubeGuideElementById('inner-tube-calculator-panel'), 'style')), !check.calculatorSelected)
  assert.equal(/display:\s*none/.test(readServerRenderedInnerTubeGuideAttribute(readUniqueServerRenderedInnerTubeGuideElementById('inner-tube-introduction-panel'), 'style')), check.calculatorSelected)

  for (const [elementId, expectedValue] of [
    ['inner-tube-rim-depth-input', check.rimDepth],
    ['inner-tube-pump-head-grip-depth-input', check.gripDepth],
    ['inner-tube-rim-depth-uncertainty-input', check.uncertainty],
  ] as const) {
    assert.equal(readServerRenderedInnerTubeGuideAttribute(readUniqueServerRenderedInnerTubeGuideElementById(elementId), 'value'), String(expectedValue))
  }

  const pressedButtons = elements.filter(element => element.tagName === 'button'
    && readServerRenderedInnerTubeGuideAttribute(element, 'aria-pressed') === 'true')
    .map(element => readServerRenderedInnerTubeGuideText(element).trim())
  const isChineseGuide = check.routePath.startsWith('/zh_cn/')
  assert.ok(pressedButtons.includes(check.manual
    ? (isChineseGuide ? '手动试配' : 'Manual fit check')
    : (isChineseGuide ? '智能推荐' : 'Smart recommendation')))
  if (check.manual) {
    assert.ok(pressedButtons.includes('80 mm'))
    assert.ok(pressedButtons.includes('+30 mm'))
  }
  if (hasInnerTubeValveFitmentShareQueryParameters(parseQuery(new URL(check.routePath, storefrontBaseUrl).search))) {
    assert.equal(elements.some(element => readServerRenderedInnerTubeGuideAttribute(element, 'class').split(/\s+/).includes('inner-tube-fitment-status')), false,
      'A shared URL must not render the unrelated default calculation as its result')
  }
  console.log(`Shared calculator SSR state passed: ${check.routePath}`)
}
