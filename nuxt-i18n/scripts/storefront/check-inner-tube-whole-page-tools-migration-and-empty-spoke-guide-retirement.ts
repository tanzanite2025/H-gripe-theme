import assert from 'node:assert/strict'
import storefrontLocales from '../../app/i18n/locales.manifest.ts'

const storefrontOrigin = process.env.SEO_STOREFRONT_BASE_URL
if (!storefrontOrigin) throw new Error('Set SEO_STOREFRONT_BASE_URL to a running storefront origin.')

async function assertRetiredInnerTubeAndEmptySpokeGuideResponse(routePath: string) {
  const response = await fetch(new URL(routePath, storefrontOrigin), {
    redirect: 'manual', signal: AbortSignal.timeout(60_000),
  })
  assert.equal(response.status, 404, `${routePath} must return a real 404`)
  assert.equal(response.headers.get('location'), null, `${routePath} must not redirect`)
}

for (const locale of storefrontLocales) {
  const prefix = locale.code === 'en' ? '' : `/${locale.code}`
  for (const suffix of ['', '?tab=calculator', '?rim_depth=80&fitment_mode=manual&valve_length=80']) {
    await assertRetiredInnerTubeAndEmptySpokeGuideResponse(`${prefix}/guides/tireguides/choose-inner-tube${suffix}`)
  }
  for (const suffix of ['?tab=choose-inner-tube', '?tab=choose-inner-tube&rim_depth=80', '?tab=tubeless&tab=choose-inner-tube']) {
    await assertRetiredInnerTubeAndEmptySpokeGuideResponse(`${prefix}/guides/tireguides${suffix}`)
  }
  await assertRetiredInnerTubeAndEmptySpokeGuideResponse(`${prefix}/guides/spokeguides`)
  await assertRetiredInnerTubeAndEmptySpokeGuideResponse(`${prefix}/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics`)
}

for (const prefix of ['', '/zh_cn']) {
  const toolResponse: Response = await fetch(new URL(`${prefix}/tools/valve-length-fitment-calculator`, storefrontOrigin))
  assert.equal(toolResponse.status, 200)
  const toolHtml = await toolResponse.text()
  assert.ok(toolHtml.includes('inner-tube-guide'))
  assert.ok(toolHtml.includes('inner-tube-fitment-'))
  for (const suffix of ['', '?tab=unknown']) {
    const guideResponse: Response = await fetch(new URL(`${prefix}/guides/tireguides${suffix}`, storefrontOrigin))
    assert.equal(guideResponse.status, 200, 'The remaining tire-guide category must remain available')
    assert.ok(!(await guideResponse.text()).includes(`href="${prefix}/guides/tireguides/choose-inner-tube`))
  }
}

console.log('Whole-page inner-tube migration passed: 120 retired page/tab URLs, 40 empty spoke-guide URLs, no redirects, new tools and remaining categories available.')
