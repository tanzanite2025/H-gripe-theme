import assert from 'node:assert/strict'
import { parseQuery } from 'vue-router'
import {
  isInnerTubeGuideContentSectionHash,
  resolveInnerTubeGuideContentPanelFromRouteLocation,
  resolveInnerTubeGuideContentSectionIdFromHash,
} from '../app/utils/innerTubeGuideContentSectionNavigation.ts'

const innerTubeGuideNavigationCases = [
  { query: '', hash: '', expectedPanel: 'introduction' },
  { query: 'utm_source=forum', hash: '', expectedPanel: 'introduction' },
  { query: 'tab=calculator', hash: '', expectedPanel: 'calculator' },
  { query: 'tab=introduction&rim_depth=80', hash: '', expectedPanel: 'introduction' },
  { query: 'rim_depth=80', hash: '', expectedPanel: 'calculator' },
  { query: 'tab=unknown', hash: '', expectedPanel: 'introduction' },
  { query: 'tab=unknown&rim_depth=80', hash: '', expectedPanel: 'calculator' },
  { query: 'tab', hash: '', expectedPanel: 'introduction' },
  { query: 'tab=calculator&tab=introduction', hash: '', expectedPanel: 'calculator' },
  { query: 'tab=introduction&tab=calculator', hash: '', expectedPanel: 'introduction' },
  { query: '', hash: '#calculator', expectedPanel: 'calculator' },
  { query: 'tab=calculator', hash: '#introduction', expectedPanel: 'introduction' },
  { query: 'tab=introduction', hash: '#inner-tube-fitment-how-to-step-3', expectedPanel: 'calculator' },
  { query: 'tab=calculator', hash: '#inner-tube-valves', expectedPanel: 'introduction' },
  { query: 'tab=calculator', hash: '#unknown', expectedPanel: 'calculator' },
  { query: 'tab=calculator', hash: '#inner-tube-guide-references', expectedPanel: 'calculator' },
]

for (const navigationCase of innerTubeGuideNavigationCases) {
  assert.equal(
    resolveInnerTubeGuideContentPanelFromRouteLocation(parseQuery(navigationCase.query), navigationCase.hash),
    navigationCase.expectedPanel,
    `Unexpected panel for ?${navigationCase.query}${navigationCase.hash}`,
  )
}

assert.equal(resolveInnerTubeGuideContentSectionIdFromHash('#calculator'), 'inner-tube-calculator-panel')
assert.equal(resolveInnerTubeGuideContentSectionIdFromHash('#introduction'), 'inner-tube-introduction-panel')
assert.equal(resolveInnerTubeGuideContentSectionIdFromHash('#inner-tube-fitment-matrix'), 'inner-tube-fitment-matrix')
assert.equal(isInnerTubeGuideContentSectionHash('#calculator'), true)
assert.equal(isInnerTubeGuideContentSectionHash('#introduction'), true)
assert.equal(isInnerTubeGuideContentSectionHash('#unknown'), false)

console.log('Inner-tube guide tab query and section hash navigation tests passed.')
