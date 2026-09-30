import { expect, test, type Page } from '@playwright/test'

const selectorBaseURL = process.env.SCHWALBE_SELECTOR_BASE_URL ?? 'http://localhost:9100'
const selectorPath = '/guides/tireguides/schwalbe-tire-selector'

const selectorURL = (query = '') => `${selectorBaseURL}${selectorPath}${query}`

const waitForNuxtMount = (page: Page) => page.waitForFunction(() => {
  const root = document.querySelector('#__nuxt')
  return Boolean(root && Reflect.get(root, '__vue_app__'))
})

const openCatalogFilters = async (page: Page) => {
  const filterButton = page.getByRole('button', { name: 'Filters' })
  const dialog = page.getByRole('dialog', { name: 'Filter catalog' })
  await expect(filterButton).toBeVisible()
  // The page can still be finishing async Nuxt hydration after the root app
  // exists. Retry the idempotent open click until the dialog is mounted.
  await expect.poll(async () => {
    if (await dialog.isVisible()) return true
    await filterButton.click()
    return dialog.isVisible()
  }, { timeout: 10_000 }).toBe(true)
}

const openFilterGroup = async (page: Page, label: string) => {
  const summary = page.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: label })
  await expect(summary).toBeVisible()
  const accordion = summary.locator('..')
  if ((await accordion.getAttribute('open')) === null) await summary.click()
}

test.describe('Schwalbe selector URL state', () => {
  test('renders catalog results outside an accessible filter dialog and returns focus on close', async ({ page }) => {
    await page.goto(selectorURL())
    await waitForNuxtMount(page)

    const filterButton = page.getByRole('button', { name: 'Filters' })
    await expect(filterButton).toBeVisible()
    await expect(filterButton.locator('svg')).toBeVisible()
    await expect(filterButton).toHaveCSS('background-color', 'rgb(11, 11, 11)')
    await expect(filterButton).toHaveText('')
    await expect(page.getByRole('dialog')).toHaveCount(0)
    await expect(page.locator('.schwalbe-selector__grid')).toBeVisible()

    await filterButton.click()
    const dialog = page.getByRole('dialog', { name: 'Filter catalog' })
    await expect(dialog).toBeVisible()
    await page.keyboard.press('Escape')

    await expect(dialog).toHaveCount(0)
    await expect(filterButton).toBeFocused()
  })

  test('uses a compact weight sort toggle', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto(selectorURL())
    await waitForNuxtMount(page)

    const sortButton = page.getByRole('button', { name: 'Sort by weight, light to heavy' })
    await expect(sortButton).toBeVisible()
    await expect(sortButton.locator('svg')).toBeVisible()
    await expect(sortButton).not.toContainText('Weight: light → heavy')
    await expect(sortButton).not.toContainText('Weight: heavy → light')
    const sortButtonBox = await sortButton.boundingBox()
    expect(sortButtonBox).not.toBeNull()
    expect(sortButtonBox?.width).toBeLessThanOrEqual(48)
    await expect(page).not.toHaveURL(/sort=/)

    await sortButton.click()
    await expect(page).toHaveURL(url => url.searchParams.get('sort') === 'weight_desc')
    await expect(page.getByRole('button', { name: 'Sort by weight, heavy to light' })).toBeVisible()
  })

  test('restores search and multi-select facet values from a direct link', async ({ page }) => {
    await page.goto(selectorURL('?search=Kojak&tire_width_min_mm=32&tire_width_max_mm=35&bead=WIRED'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    await expect(page.getByRole('searchbox')).toHaveValue('Kojak')
    await expect(page.getByRole('spinbutton', { name: 'Tire width Minimum', exact: true })).toHaveValue('32')
    await expect(page.getByRole('spinbutton', { name: 'Tire width Maximum', exact: true })).toHaveValue('35')
    await expect(page.getByRole('checkbox', { name: 'WIRED', exact: true })).toBeChecked()
  })

  test('keeps legacy exact tire width links working', async ({ page }) => {
    await page.goto(selectorURL('?search=Kojak&tire_width_mm=35'))
    await waitForNuxtMount(page)

    await expect(page.locator('.schwalbe-selector__grid')).toBeVisible()
    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('search') === 'Kojak'
      && url.searchParams.get('tire_width_mm') === '35'
      && !url.searchParams.has('tire_width_min_mm')
      && !url.searchParams.has('tire_width_max_mm')
    ))

    await openCatalogFilters(page)
    await expect(page.getByRole('spinbutton', { name: 'Tire width Minimum', exact: true })).toHaveValue('35')
    await expect(page.getByRole('spinbutton', { name: 'Tire width Maximum', exact: true })).toHaveValue('35')
  })

  test('restores the wheel diameter and BSD pair as one facet', async ({ page }) => {
    await page.goto(selectorURL('?wheel_size=26-559'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)
    await openFilterGroup(page, 'Wheel size (BSD)')

    const selectedOption = page.getByRole('checkbox', { name: '26" (BSD 559 mm)', exact: true })
    await expect(selectedOption).toBeChecked()
    await expect(page.getByRole('checkbox', { name: '26" (BSD 590 mm)', exact: true })).not.toBeChecked()
    await expect(page).toHaveURL(url => url.searchParams.get('wheel_size') === '26-559')
  })

  test('keeps wheel size and rim inner width in one drawer flow', async ({ page }) => {
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await expect(page.locator('.schwalbe-selector__rim-match')).toHaveCount(0)
    await openCatalogFilters(page)

    const innerWidthInput = page.getByRole('spinbutton', { name: 'Rim inner width' })
    await expect(innerWidthInput).toBeVisible()
    await expect(innerWidthInput.locator('xpath=ancestor::details')).toHaveCount(0)
    await innerWidthInput.fill('23.5')
    await expect(page.getByRole('alert')).toContainText('Choose exactly one wheel diameter and BSD pair')
    await expect(page.getByRole('button', { name: 'Show results' })).toBeDisabled()

    await openFilterGroup(page, 'Wheel size (BSD)')
    const wheel28 = page.getByRole('checkbox', { name: '28" (BSD 622 mm)', exact: true })
    const wheel29 = page.getByRole('checkbox', { name: '29" (BSD 622 mm)', exact: true })
    await wheel28.check()
    await expect(page.getByRole('button', { name: 'Show results' })).toBeEnabled()
    await wheel29.check()
    await expect(page.getByRole('button', { name: 'Show results' })).toBeDisabled()
    await wheel29.uncheck()
    await page.getByRole('button', { name: 'Show results' }).click()

    await expect(page).toHaveURL(url => (
      url.searchParams.get('wheel_size') === '28-622'
      && url.searchParams.get('inner_rim_width_mm') === '23.5'
    ))

    await openCatalogFilters(page)
    await openFilterGroup(page, 'Wheel size (BSD)')
    await page.getByRole('button', { name: 'Clear rim match' }).click()
    await openFilterGroup(page, 'Tire width')
    const minimumWidth = page.getByRole('spinbutton', { name: 'Tire width Minimum', exact: true })
    await minimumWidth.fill('32')
    await minimumWidth.press('Tab')
    await openFilterGroup(page, 'Wheel size (BSD)')
    await expect(page.getByRole('spinbutton', { name: 'Rim inner width' })).toHaveValue('')
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL(url => (
      url.searchParams.get('tire_width_min_mm') === '32'
      && !url.searchParams.has('inner_rim_width_mm')
      && url.searchParams.get('wheel_size') === '28-622'
    ))

    await openCatalogFilters(page)
    await page.getByRole('button', { name: 'Clear filters', exact: true }).click()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).not.toHaveURL(/wheel_size=|inner_rim_width_mm=|tire_width_min_mm=/)
  })

  test('sets a tire width range through numeric inputs and range sliders', async ({ page }) => {
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const minimumInput = page.getByRole('spinbutton', { name: 'Tire width Minimum', exact: true })
    const maximumInput = page.getByRole('spinbutton', { name: 'Tire width Maximum', exact: true })
    const draftURL = page.url()
    await minimumInput.fill('32')
    await minimumInput.press('Tab')
    await maximumInput.fill('35')
    await maximumInput.press('Tab')

    // Editing the drawer is a draft operation. It must not replace the URL or
    // restart the catalog request until the user confirms it.
    expect(page.url()).toBe(draftURL)

    const minimumSlider = page.getByRole('slider', { name: 'Tire width Minimum', exact: true })
    const maximumSlider = page.getByRole('slider', { name: 'Tire width Maximum', exact: true })
    await expect(minimumSlider).toBeEnabled()
    await expect(maximumSlider).toBeEnabled()

    const previousMinimum = await minimumSlider.inputValue()
    await minimumSlider.press('ArrowRight')
    await expect(minimumSlider).not.toHaveValue(previousMinimum)
    expect(page.url()).toBe(draftURL)

    const minimumValueAfterSlider = await minimumInput.inputValue()
    await expect(minimumValueAfterSlider).not.toBe('')

    const previousMaximum = await maximumSlider.inputValue()
    await maximumSlider.press('ArrowLeft')
    await expect(maximumSlider).not.toHaveValue(previousMaximum)
    const maximumValueAfterSlider = await maximumInput.inputValue()
    await expect(maximumValueAfterSlider).not.toBe('')
    expect(page.url()).toBe(draftURL)

    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('tire_width_min_mm') === minimumValueAfterSlider
      && url.searchParams.get('tire_width_max_mm') === maximumValueAfterSlider
      && !url.searchParams.has('tire_width_mm')
      && !url.searchParams.has('page')
    ))
  })

  test('discards an unsubmitted drawer draft when it is closed', async ({ page }) => {
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const minimumInput = page.getByRole('spinbutton', { name: 'Tire width Minimum', exact: true })
    const initialURL = page.url()
    await minimumInput.fill('32')
    await minimumInput.press('Tab')
    expect(page.url()).toBe(initialURL)

    await page.keyboard.press('Escape')
    await expect(page.getByRole('dialog')).toHaveCount(0)
    expect(page.url()).toBe(initialURL)

    await openCatalogFilters(page)
    await expect(page.getByRole('spinbutton', { name: 'Tire width Minimum', exact: true })).toHaveValue('')
  })

  test('keeps the drawer compact and usable on a phone viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const dialog = page.getByRole('dialog', { name: 'Filter catalog' })
    await expect(dialog).toBeVisible()
    await expect(dialog.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Tire width' })).toBeVisible()
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet')).toHaveCount(4)
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet').filter({ hasText: 'Casing orientation' })).toBeVisible()
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet').filter({ hasText: 'E-Bike rating' })).toBeVisible()
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet').filter({ hasText: 'Seal' })).toBeVisible()
    await expect(dialog.getByRole('checkbox', { name: 'Radial', exact: true })).toBeVisible()
    await expect(dialog.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Casing orientation' })).toHaveCount(0)
    await expect(dialog.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'E-Bike rating' })).toHaveCount(0)
    await expect(dialog.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Seal' })).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Show results' })).toBeVisible()
    await expect(dialog).toHaveJSProperty('open', true)
  })

  test('keeps desktop dialog dimensions stable when one accordion expands', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const dialog = page.getByRole('dialog', { name: 'Filter catalog' })
    const before = await dialog.boundingBox()
    expect(before).not.toBeNull()

    await openFilterGroup(page, 'Casing construction')

    const after = await dialog.boundingBox()
    expect(after).not.toBeNull()
    expect(after?.width).toBe(before?.width)
    expect(after?.height).toBe(before?.height)

    const panel = page.locator('.schwalbe-filter-panel')
    const accordions = panel.locator(':scope > .schwalbe-filter-panel__accordion')
    await expect(accordions).toHaveCount(3)
    await expect(panel.locator(':scope > .schwalbe-filter-panel__accordion[open]')).toHaveCount(1)
    const accordionBoxes = await Promise.all(
      Array.from({ length: 3 }, (_, index) => accordions.nth(index).boundingBox()),
    )
    expect(accordionBoxes.every(box => box !== null)).toBe(true)
    expect(new Set(accordionBoxes.map(box => Math.round(box?.x ?? 0))).size).toBe(1)
  })

  test('keeps search and clears the old page when a facet changes', async ({ page }) => {
    await page.goto(selectorURL('?search=Marathon&page=2'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const minimumInput = page.getByRole('spinbutton', { name: 'Tire width Minimum', exact: true })
    const initialURL = page.url()
    await minimumInput.fill('32')
    await minimumInput.press('Tab')

    expect(page.url()).toBe(initialURL)
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('search') === 'Marathon'
      && url.searchParams.get('tire_width_min_mm') === '32'
      && !url.searchParams.has('tire_width_max_mm')
      && !url.searchParams.has('tire_width_mm')
      && !url.searchParams.has('page')
    ))
  })

  test('hides the retired minimum tire load filter and clears its legacy condition', async ({ page }) => {
    await page.goto(selectorURL('?min_load_kg=90.5&page=2'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)
    await expect(page.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Minimum tire load capacity' })).toHaveCount(0)
    await expect(page.getByRole('spinbutton', { name: 'Minimum tire load capacity' })).toHaveCount(0)
    await page.getByRole('checkbox', { name: 'Folding', exact: true }).check()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('bead') === 'Folding'
      && !url.searchParams.has('min_load_kg')
      && !url.searchParams.has('page')
    ))
  })

  test('hides color and compound facets and clears their legacy conditions', async ({ page }) => {
    await page.goto(selectorURL('?color=Black&compound=ADDIX&page=2'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    await expect(page.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Color' })).toHaveCount(0)
    await expect(page.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Compound' })).toHaveCount(0)
    await expect(page.getByRole('checkbox', { name: 'Black', exact: true })).toHaveCount(0)
    await expect(page.getByRole('checkbox', { name: 'ADDIX', exact: true })).toHaveCount(0)

    await page.getByRole('checkbox', { name: 'Folding', exact: true }).check()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('bead') === 'Folding'
      && !url.searchParams.has('color')
      && !url.searchParams.has('compound')
      && !url.searchParams.has('page')
    ))
  })

  test('clears facet query values without clearing the search term', async ({ page }) => {
    await page.goto(selectorURL('?search=Marathon&bead=Folding&min_load_kg=90'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)
    await expect(page.getByRole('checkbox', { name: 'Folding', exact: true })).toBeChecked()

    await page.getByRole('button', { name: 'Show results' }).click()
    await openCatalogFilters(page)
    await page.getByRole('button', { name: 'Clear filters', exact: true }).click()
    await page.getByRole('button', { name: 'Show results' }).click()

    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('search') === 'Marathon'
      && !url.searchParams.has('bead')
      && !url.searchParams.has('min_load_kg')
      && !url.searchParams.has('page')
    ))
    await openCatalogFilters(page)
    await expect(page.getByRole('checkbox', { name: 'Folding', exact: true })).not.toBeChecked()
  })

  test('restores the selected facet across pagination back and forward navigation', async ({ page }) => {
    await page.goto(selectorURL('?bead=Folding'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)
    await expect(page.getByRole('checkbox', { name: 'Folding', exact: true })).toBeChecked()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)

    const pagination = page.getByRole('navigation', { name: 'Catalog pages' })
    await pagination.getByRole('link', { name: '2', exact: true }).click()
    await expect(page).toHaveURL(url => url.searchParams.get('page') === '2')
    await openCatalogFilters(page)
    await expect(page.getByRole('checkbox', { name: 'Folding', exact: true })).toBeChecked()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)

    await page.goBack()
    await expect(page).toHaveURL(url => url.searchParams.get('bead') === 'Folding' && !url.searchParams.has('page'))
    await openCatalogFilters(page)
    await expect(page.getByRole('checkbox', { name: 'Folding', exact: true })).toBeChecked()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)

    await page.goForward()
    await expect(page).toHaveURL(url => url.searchParams.get('bead') === 'Folding' && url.searchParams.get('page') === '2')
    await openCatalogFilters(page)
    await expect(page.getByRole('checkbox', { name: 'Folding', exact: true })).toBeChecked()
  })
})
