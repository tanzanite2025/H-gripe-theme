import { expect, test } from '@playwright/test'

const storefrontUrl = process.env.STOREFRONT_URL || 'http://127.0.0.1:4317/'

test.use({ locale: 'zh-CN' })

test('anonymous initialization stays quiet and avoids optional account probes', async ({ page }) => {
  const apiRequests: string[] = []
  const consoleErrors: string[] = []
  const pageErrors: string[] = []

  page.on('request', (request) => {
    const url = new URL(request.url())
    if (url.pathname.startsWith('/api/v1/')) apiRequests.push(url.pathname)
  })

  page.on('console', (message) => {
    if (message.type() === 'error') consoleErrors.push(message.text())
  })

  page.on('pageerror', (error) => {
    pageErrors.push(error.message)
  })

  await page.route('**/api/v1/**', async (route) => {
    const pathname = new URL(route.request().url()).pathname

    if (pathname === '/api/v1/auth/profile') {
      await route.fulfill({
        status: 204,
        body: '',
      })
      return
    }

    const data = pathname === '/api/v1/shipping/templates' || pathname === '/api/v1/payment/tax-rates'
      ? { code: 0, data: { data: [] } }
      : { code: 0, data: { data: [], items: [] }, items: [] }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(data),
    })
  })

  await page.goto(storefrontUrl)

  await page.waitForTimeout(500)

  expect(apiRequests).not.toContain('/api/v1/auth/profile')
  expect(apiRequests).not.toContain('/api/v1/cart/summary')
  expect(apiRequests).not.toContain('/api/v1/wishlist')
  expect(apiRequests).not.toContain('/api/v1/marketing/loyalty/points')
  expect(apiRequests).not.toContain('/api/v1/marketing/loyalty/levels')
  expect(apiRequests).not.toContain('/api/v1/marketing/loyalty/config')
  expect(apiRequests).not.toContain('/api/v1/marketing/loyalty/rules')
  expect(pageErrors).toEqual([])
  expect(consoleErrors).toEqual([])
})

test('session hints restore profile and keep account data lazy', async ({ page, context }) => {
  const apiRequests: string[] = []

  page.on('request', (request) => {
    const url = new URL(request.url())
    if (url.pathname.startsWith('/api/v1/')) apiRequests.push(url.pathname)
  })

  await page.route('**/api/v1/**', async (route) => {
    const pathname = new URL(route.request().url()).pathname
    let data: unknown

    if (pathname === '/api/v1/auth/profile') {
      data = { code: 0, data: { id: 42, username: 'returning-customer' } }
    } else if (pathname === '/api/v1/marketing/loyalty/points') {
      data = { code: 0, data: { total: 100, available: 80, tier: 'member' } }
    } else {
      data = { code: 0, data: { data: [], items: [] }, items: [] }
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(data),
    })
  })

  await context.addCookies([
    {
      name: 'csrf_token',
      value: 'session-hint',
      url: storefrontUrl,
    },
  ])

  await page.goto(storefrontUrl)

  await expect.poll(() => apiRequests).toContain('/api/v1/auth/profile')
  await page.waitForTimeout(500)
  expect(apiRequests).not.toContain('/api/v1/wishlist')
  expect(apiRequests).not.toContain('/api/v1/marketing/loyalty/assets')

  await page.evaluate(() => {
    window.dispatchEvent(new CustomEvent('ui:sidebar-open', { detail: { side: 'left' } }))
  })

  await expect.poll(() => apiRequests).toContain('/api/v1/wishlist')
  await expect.poll(() => apiRequests).toContain('/api/v1/marketing/loyalty/assets')
  expect(apiRequests.filter((pathname) => pathname === '/api/v1/auth/profile')).toHaveLength(1)
})
