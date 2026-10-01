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

test.describe('Schwalbe selector URL state', () => {
  test('keeps the page H1 semantic while hiding only its visual presentation', async ({ page }) => {
    await page.goto(selectorURL())
    await waitForNuxtMount(page)

    const title = page.locator('#schwalbe-selector-title')
    await expect(title).toHaveText('Schwalbe tire selector')
    await expect(title).toHaveAttribute('class', /schwalbe-selector__title--sr-only/)
    await expect(title).toHaveCSS('position', 'absolute')
    await expect(title).toHaveCSS('overflow', 'hidden')

    const titleBox = await title.boundingBox()
    expect(titleBox).not.toBeNull()
    expect(titleBox?.width).toBeLessThanOrEqual(1)
    expect(titleBox?.height).toBeLessThanOrEqual(1)
  })

  test('separates the introduction and full catalog search into page tabs', async ({ page }) => {
    await page.goto(selectorURL('?search=Kojak&wheel_size=26-559'))
    await waitForNuxtMount(page)

    const introTab = page.getByRole('tab', { name: 'Introduction', exact: true })
    const searchTab = page.getByRole('tab', { name: 'Search', exact: true })
    const introPanel = page.locator('.schwalbe-selector__section-panel--intro')
    const searchPanel = page.locator('.schwalbe-selector__section-panel--search')
    const searchbox = page.getByRole('searchbox')

    await expect(searchTab).toHaveAttribute('aria-selected', 'true')
    await expect(searchPanel).toBeVisible()
    await expect(introPanel).toBeHidden()
    await expect(searchbox).toHaveValue('Kojak')

    await introTab.click()
    await expect(introTab).toHaveAttribute('aria-selected', 'true')
    await expect(introPanel).toBeVisible()
    await expect(searchPanel).toBeHidden()
    await expect(page).toHaveURL(/search=Kojak/)
    await expect(page).toHaveURL(/wheel_size=26-559/)

    await searchTab.click()
    await expect(searchPanel).toBeVisible()
    await expect(searchbox).toHaveValue('Kojak')
    await expect(page.getByRole('button', { name: '26" · BSD 559 mm', exact: true })).toHaveAttribute('aria-pressed', 'true')
  })

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

  test('uses the storefront font for telemetry metadata', async ({ page }) => {
    await page.goto(selectorURL())
    await waitForNuxtMount(page)

    const fontFamilies = await page.locator([
      '.schwalbe-telemetry__year',
      '.schwalbe-telemetry__kicker',
      '.schwalbe-telemetry__badge',
      '.schwalbe-telemetry__tab',
    ].join(',')).evaluateAll((elements) => elements.map((element) => getComputedStyle(element).fontFamily))

    expect(fontFamilies.length).toBeGreaterThan(0)
    expect(fontFamilies.every((fontFamily) => fontFamily.includes('MapleUI'))).toBe(true)
  })

  test('fills the shared content shell on a wide viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 900 })
    await page.goto(selectorURL())
    await waitForNuxtMount(page)

    const widths = await page.evaluate(() => {
      const shell = document.querySelector('.page-content-shell')?.getBoundingClientRect()
      const pageRoot = document.querySelector('.schwalbe-page')?.getBoundingClientRect()
      return {
        shell: shell?.width ?? 0,
        pageRoot: pageRoot?.width ?? 0,
      }
    })

    expect(widths.shell).toBeGreaterThan(1700)
    expect(widths.pageRoot).toBeCloseTo(widths.shell, 0)
  })

  test('explains casing, color, and compound values in separate telemetry tabs', async ({ page }) => {
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await page.getByRole('tab', { name: 'Introduction', exact: true }).click()

    const colorValues = [
      'Black',
      'Black/Coffee+Reflex',
      'Black+BlackReflex',
      'Black+Reflex',
      'Blue Stripes',
      'Bronze',
      'Bronze Sidewall',
      'Bronze+Reflex',
      'Brown/Whitewall+Reflex',
      'Brown+Reflex',
      'Classic',
      'Creme+Reflex',
      'Grey Stripes',
      'Grey/Black',
      'Gumwall',
      'Red Stripes',
      'Transparent Sidewall',
      'White Stripes',
      'White/Bordeaux',
      'Whitewall',
      'Whitewall+Reflex',
    ]
    const compoundValues = [
      'ADDIX',
      'ADDIX 365',
      'ADDIX 4-Season',
      'ADDIX E',
      'ADDIX Eco',
      'ADDIX Green',
      'ADDIX Race',
      'ADDIX Soft',
      'ADDIX Speed',
      'ADDIX SpeedGrip',
      'ADDIX Ultra Soft',
      "Black'n'Roll",
      'Endurance',
      'GRC',
      'Green Compound',
      'MID',
      'SBC',
      'Silica',
      'SOFT',
      'SPEED',
      'ULTRA SOFT',
      'WheelStar',
      'Winter',
    ]
    const casingValues = [
      'Super Race',
      'Super Ground',
      'Super Trail',
      'Super Downhill',
      'TRAIL',
      'TRAIL PRO',
      'GRAVITY',
      'GRAVITY PRO',
    ]

    const casingTab = page.getByRole('tab', { name: 'Casing construction', exact: true })
    await expect(casingTab).toBeVisible()
    await casingTab.click()

    const casingTopic = page.locator('.schwalbe-telemetry__topic--casing')
    await expect(casingTopic).toBeVisible()
    await expect(casingTopic).toContainText('What the Schwalbe casing construction values mean')
    await expect(casingTopic.locator('.schwalbe-telemetry__catalog-value')).toHaveText(casingValues)
    await expect(casingTopic.locator('.schwalbe-telemetry__catalog-card')).toHaveCount(casingValues.length)

    const colorTab = page.getByRole('tab', { name: /Color$/ })
    await expect(colorTab).toBeVisible()
    await colorTab.click()

    const colorTopic = page.locator('.schwalbe-telemetry__topic--color')
    await expect(colorTopic).toBeVisible()
    await expect(colorTopic).toContainText('Every color value in this catalog')
    await expect(colorTopic.locator('.schwalbe-telemetry__catalog-section')).toHaveCount(1)
    await expect(colorTopic.locator('.schwalbe-telemetry__catalog-value')).toHaveText(colorValues)
    await expect(colorTopic.locator('.schwalbe-telemetry__catalog-card')).toHaveCount(colorValues.length)

    const compoundTab = page.getByRole('tab', { name: '🧪 Compound', exact: true })
    await expect(compoundTab).toBeVisible()
    await compoundTab.click()

    const compoundTopic = page.locator('.schwalbe-telemetry__topic--compound')
    await expect(compoundTopic).toBeVisible()
    await expect(compoundTopic).toContainText('Every compound value in this catalog')
    await expect(compoundTopic.locator('.schwalbe-telemetry__catalog-section')).toHaveCount(1)
    await expect(compoundTopic.locator('.schwalbe-telemetry__catalog-value')).toHaveText(compoundValues)
    await expect(compoundTopic.locator('.schwalbe-telemetry__catalog-card')).toHaveCount(compoundValues.length)
  })

  test('groups telemetry tabs into swipeable pairs on a phone viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await page.getByRole('tab', { name: 'Introduction', exact: true }).click()

    const desktopTabs = page.locator('.schwalbe-telemetry__tabs--desktop')
    const mobileTabs = page.locator('.schwalbe-telemetry__mobile-tabs')
    const mobileRail = page.locator('.schwalbe-telemetry__mobile-tab-rail')
    const mobileDots = page.locator('.schwalbe-telemetry__mobile-pagination .tz-carousel-pagination__dot')

    await expect(desktopTabs).toBeHidden()
    await expect(mobileTabs).toBeVisible()
    await expect(mobileRail.locator('[data-mobile-tab-group]')).toHaveCount(4)
    await expect(mobileDots).toHaveCount(4)
    await expect(mobileTabs.getByRole('tab', { name: 'Radial casing', exact: true })).toHaveAttribute('aria-selected', 'true')

    await mobileDots.nth(1).click()
    await expect(mobileDots.nth(1)).toHaveAttribute('aria-selected', 'true')
    await expect(mobileTabs.getByRole('tab', { name: 'Green Marathon', exact: true })).toHaveAttribute('aria-selected', 'true')
    await expect(page.locator('.schwalbe-telemetry__topic--green')).toBeVisible()

    await mobileRail.locator('[data-mobile-tab-group="2"]').scrollIntoViewIfNeeded()
    await expect.poll(async () => mobileDots.nth(2).getAttribute('aria-selected')).toBe('true')
    await expect(mobileTabs.getByRole('tab', { name: 'Color', exact: true })).toHaveAttribute('aria-selected', 'true')

    await mobileDots.nth(3).click()
    await expect(mobileTabs.getByRole('tab', { name: 'ADDIX / Compound', exact: true })).toHaveAttribute('aria-selected', 'true')
    await expect(mobileRail.locator('[data-mobile-tab-group="3"] [role="tab"]')).toHaveCount(1)
    await expect(page.locator('.schwalbe-telemetry__topic--addix')).toBeVisible()
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

  test('hides the tire width facet and clears legacy width conditions', async ({ page }) => {
    await page.goto(selectorURL('?search=Kojak&tire_width_min_mm=32&tire_width_max_mm=35&page=2'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    await expect(page.getByRole('searchbox')).toHaveValue('Kojak')
    await expect(page.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Tire width' })).toHaveCount(0)
    await expect(page.getByRole('spinbutton', { name: /Tire width/ })).toHaveCount(0)
    await expect(page.getByRole('slider', { name: /Tire width/ })).toHaveCount(0)

    await page.getByRole('radio', { name: 'Radial', exact: true }).check()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('search') === 'Kojak'
      && !url.searchParams.has('tire_width_mm')
      && !url.searchParams.has('tire_width_min_mm')
      && !url.searchParams.has('tire_width_max_mm')
      && !url.searchParams.has('page')
    ))
  })

  test('shows the wheel diameter and BSD pair as a page-level size button', async ({ page }) => {
    await page.goto(selectorURL('?wheel_size=26-559'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    await expect(page.getByRole('button', { name: '26" · BSD 559 mm', exact: true })).toHaveAttribute('aria-pressed', 'true')
    await expect(page.getByRole('button', { name: '26" · BSD 590 mm', exact: true })).toHaveAttribute('aria-pressed', 'false')
    await expect(page.getByRole('dialog').getByText('Wheel size', { exact: true })).toHaveCount(0)
    await page.getByRole('button', { name: 'Close filters', exact: true }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)
    await expect(page).toHaveURL(url => url.searchParams.get('wheel_size') === '26-559')
  })

  test('uses a bounded wheel size picker dialog on a phone viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto(selectorURL('?wheel_size=26-559'))
    await waitForNuxtMount(page)

    const wheelSizeTrigger = page.getByRole('button', { name: /Wheel size navigation/ })
    await expect(wheelSizeTrigger).toBeVisible()
    await expect(page.locator('.schwalbe-wheel-size-tabs__desktop')).toBeHidden()
    await expect(wheelSizeTrigger).toContainText('26" · BSD 559 mm')
    await expect.poll(async () => page.locator('.schwalbe-wheel-size-tabs__mobile-trigger').evaluate((element) => {
      const rect = element.getBoundingClientRect()
      return rect.left >= 0 && rect.right <= window.innerWidth && rect.width > 0
    })).toBe(true)
    await expect.poll(async () => page.evaluate(() => ({
      viewport: window.innerWidth,
      scrollWidth: document.documentElement.scrollWidth,
    }))).toEqual({ viewport: 390, scrollWidth: 390 })

    await wheelSizeTrigger.click()
    const wheelSizeDialog = page.getByRole('dialog', { name: 'Select wheel size' })
    await expect(wheelSizeDialog).toBeVisible()
    await expect.poll(async () => wheelSizeDialog.evaluate((element) => {
      const rect = element.getBoundingClientRect()
      return rect.left >= 0
        && rect.top >= 0
        && rect.right <= window.innerWidth
        && rect.bottom <= window.innerHeight
        && element.scrollWidth <= window.innerWidth
    })).toBe(true)
    await expect(wheelSizeDialog.getByRole('option', { name: '26" · BSD 559 mm', exact: true })).toHaveAttribute('aria-selected', 'true')

    await wheelSizeDialog.getByRole('option', { name: '29" · BSD 622 mm', exact: true }).click()
    await expect(page).toHaveURL(url => url.searchParams.get('wheel_size') === '29-622')
    await expect(wheelSizeDialog).toHaveCount(0)
    await expect(wheelSizeTrigger).toContainText('29" · BSD 622 mm')

    await wheelSizeTrigger.click()
    await page.getByRole('dialog', { name: 'Select wheel size' }).getByRole('option', { name: 'All wheel sizes', exact: true }).click()
    await expect(page).not.toHaveURL(/wheel_size=/)
  })

  test('hides casing construction filtering and keeps Radial in the URL', async ({ page }) => {
    await page.goto(selectorURL('?casing=Super%20Race&radial=1'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const radialOption = page.getByRole('radio', { name: 'Radial', exact: true })
    await expect(radialOption).toBeChecked()
    await expect(page.locator('fieldset.schwalbe-filter-panel__inline-facet').filter({ hasText: 'Casing orientation' })).toHaveCount(0)
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL(url => (
      url.searchParams.get('radial') === '1'
      && !url.searchParams.has('casing')
    ))
  })

  test('shows rim inner-width guidance on cards instead of filtering by entered width', async ({ page }) => {
    await page.goto(selectorURL('?inner_rim_width_mm=23.5'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    await expect(page.getByRole('spinbutton', { name: 'Rim inner width' })).toHaveCount(0)
    await expect(page.locator('.schwalbe-filter-panel__rim-match')).toHaveCount(0)
    await page.getByRole('button', { name: 'Close filters', exact: true }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)

    const wheel28 = page.getByRole('button', { name: '28" · BSD 622 mm', exact: true })
    const wheel29 = page.getByRole('button', { name: '29" · BSD 622 mm', exact: true })
    await wheel28.click()
    await expect(page).toHaveURL(url => (
      url.searchParams.get('wheel_size') === '28-622'
      && !url.searchParams.has('inner_rim_width_mm')
    ))

    await wheel29.click()
    await expect(page).toHaveURL(url => (
      url.searchParams.get('wheel_size') === '29-622'
      && !url.searchParams.has('inner_rim_width_mm')
    ))

    const guidance = page.locator('.schwalbe-tire-card__rim-guidance').first()
    await expect(guidance).toBeVisible()
    await expect(guidance).toContainText('Rim inner-width reference')
    await expect(guidance).toContainText('17–27 mm')

    await openCatalogFilters(page)
    await page.getByRole('button', { name: 'Clear filters', exact: true }).click()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL(url => (
      url.searchParams.get('wheel_size') === '29-622'
      && !url.searchParams.has('inner_rim_width_mm')
    ))
    await page.getByRole('button', { name: 'All wheel sizes', exact: true }).click()
    await expect(page).not.toHaveURL(/wheel_size=|inner_rim_width_mm=/)
  })

  test('keeps the drawer compact and usable on a phone viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const dialog = page.getByRole('dialog', { name: 'Filter catalog' })
    await expect(dialog).toBeVisible()
    await expect(dialog.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Tire width' })).toHaveCount(0)
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet')).toHaveCount(4)
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet').filter({ hasText: 'Casing orientation' })).toHaveCount(0)
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet').filter({ hasText: 'E-Bike marking' })).toBeVisible()
    await expect(dialog.locator('fieldset.schwalbe-filter-panel__inline-facet').filter({ hasText: 'Seal' })).toBeVisible()
    await expect(dialog.getByRole('radio', { name: 'ALL', exact: true })).toBeChecked()
    await expect(dialog.getByRole('radio', { name: 'Radial', exact: true })).not.toBeChecked()
    await dialog.getByRole('radio', { name: 'Radial', exact: true }).check()
    await expect(dialog.getByRole('radio', { name: 'Radial', exact: true })).toBeChecked()
    await expect(dialog.getByRole('radio', { name: 'ALL', exact: true })).not.toBeChecked()
    await dialog.getByRole('radio', { name: 'ALL', exact: true }).check()
    await expect(dialog.getByRole('radio', { name: 'ALL', exact: true })).toBeChecked()
    await expect(dialog.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'E-Bike marking' })).toHaveCount(0)
    await expect(dialog.locator('summary.schwalbe-filter-panel__accordion-title').filter({ hasText: 'Seal' })).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Show results' })).toBeVisible()
    await expect(dialog).toHaveJSProperty('open', true)
  })

  test('keeps desktop dialog dimensions stable with inline facets', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto(selectorURL())
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    const dialog = page.getByRole('dialog', { name: 'Filter catalog' })
    const before = await dialog.boundingBox()
    expect(before).not.toBeNull()

    const after = await dialog.boundingBox()
    expect(after).not.toBeNull()
    expect(after?.width).toBe(before?.width)
    expect(after?.height).toBe(before?.height)

    const panel = page.locator('.schwalbe-filter-panel')
    await expect(panel.locator(':scope > .schwalbe-filter-panel__inline-facet')).toHaveCount(4)
    await expect(panel.locator('summary.schwalbe-filter-panel__accordion-title')).toHaveCount(0)
  })

  test('keeps search and clears the old page when a facet changes', async ({ page }) => {
    await page.goto(selectorURL('?search=Marathon&page=2'))
    await waitForNuxtMount(page)
    await openCatalogFilters(page)

    await page.getByRole('radio', { name: 'Radial', exact: true }).check()
    await page.getByRole('button', { name: 'Show results' }).click()
    await expect(page).toHaveURL((url) => (
      url.pathname === selectorPath
      && url.searchParams.get('search') === 'Marathon'
      && url.searchParams.get('radial') === '1'
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
