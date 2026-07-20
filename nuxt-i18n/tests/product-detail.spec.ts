import { expect, test } from '@playwright/test'

const productDetailUrl = process.env.PRODUCT_DETAIL_URL
  || '/shop/g35-carbon-rim'

test('product detail keeps readable text and persistent information tabs', async ({ page }) => {
  const consoleErrors: string[] = []
  const pageErrors: string[] = []

  page.on('console', (message) => {
    if (message.type() === 'error') consoleErrors.push(message.text())
  })

  page.on('pageerror', (error) => {
    pageErrors.push(error.message)
  })

  await page.goto(productDetailUrl)
  await page.waitForFunction(() => {
    const nuxtRoot = document.querySelector('#__nuxt') as HTMLElement & { __vue_app__?: unknown }
    return Boolean(nuxtRoot?.__vue_app__)
  })

  const title = page.getByRole('heading', { level: 1, name: 'G35 Carbon Rim' })
  await expect(title).toBeVisible()

  await expect(title).toHaveCSS('color', 'rgb(248, 250, 252)')
  await expect(page.locator('.product-price')).toHaveCSS('color', 'rgb(248, 250, 252)')
  await expect(page.locator('.product-sku').first()).toHaveCSS('color', 'rgb(203, 213, 225)')
  await expect(page.locator('.variant-stock')).toHaveCSS('color', 'rgba(255, 255, 255, 0.58)')
  await expect(page.locator('.product-gallery h2')).toHaveCSS('color', 'rgb(248, 250, 252)')

  const addButton = page.locator('.product-add-button')
  await expect(addButton).toBeDisabled()
  await expect(addButton).toHaveCSS('opacity', '1')
  await expect(addButton).toHaveCSS('color', 'rgb(203, 213, 225)')

  const tabs = page.getByRole('tab')
  await expect(tabs).toHaveCount(4)
  await expect(tabs).toHaveText(['Details', 'After-sales', 'Packaging', 'Shipping'])

  const detailsTab = page.getByRole('tab', { name: 'Details' })
  await expect(detailsTab).toHaveAttribute('aria-selected', 'true')
  await expect(page.getByRole('tabpanel', { name: 'Details' })).toBeVisible()

  await detailsTab.focus()
  await detailsTab.press('ArrowRight')
  const afterSalesTab = page.getByRole('tab', { name: 'After-sales' })
  await expect(afterSalesTab).toBeFocused()
  await expect(afterSalesTab).toHaveAttribute('aria-selected', 'true')
  await expect(afterSalesTab).toHaveAttribute('tabindex', '0')
  await expect(detailsTab).toHaveAttribute('tabindex', '-1')
  await expect(page.getByRole('tabpanel', { name: 'Details' })).toBeHidden()
  await expect(page.getByRole('tabpanel', { name: 'After-sales' })).toContainText(
    'After-sales information has not been added yet.',
  )

  await afterSalesTab.press('End')
  await expect(page.getByRole('tab', { name: 'Shipping' })).toBeFocused()
  await page.getByRole('tab', { name: 'Shipping' }).press('ArrowRight')
  await expect(detailsTab).toBeFocused()

  await page.getByRole('tab', { name: 'Packaging' }).click()
  await expect(page.getByRole('tabpanel', { name: 'Packaging' })).toContainText(
    'Packaging information has not been added yet.',
  )

  await page.getByRole('tab', { name: 'Shipping' }).click()
  await expect(page.getByRole('tabpanel', { name: 'Shipping' })).toContainText(
    'Shipping information has not been added yet.',
  )

  await page.setViewportSize({ width: 375, height: 844 })
  await expect.poll(async () => page.evaluate(() => {
    return document.documentElement.scrollWidth <= document.documentElement.clientWidth
  })).toBe(true)

  await detailsTab.focus()
  await detailsTab.press('End')
  const shippingBox = await page.getByRole('tab', { name: 'Shipping' }).boundingBox()
  expect(shippingBox).not.toBeNull()
  expect(shippingBox!.x).toBeGreaterThanOrEqual(0)
  expect(shippingBox!.x + shippingBox!.width).toBeLessThanOrEqual(375)

  expect(pageErrors).toEqual([])
  expect(consoleErrors).toEqual([])
})
