// 启动 Nuxt 开发服务后，通过 playwright-cli run-code --filename=<本文件> 运行。
async function verifyIndependentTestReportNetworkRecoveryAndDialogLifecycle(page) {
  const origin = await page.evaluate(() => location.origin)
  const results = []
  let currentScenario = 'initialization'
  const uncaughtBrowserErrors = []
  const recordUncaughtBrowserError = error => uncaughtBrowserErrors.push(error.message)
  page.on('pageerror', recordUncaughtBrowserError)

  function assertTestReportRegressionCondition(condition, description, details) {
    if (!condition) throw new Error(`${description}: ${JSON.stringify(details)}`)
  }

  async function mountIndependentTestReportControlsOnColdPage() {
    await page.goto(`${origin}/zh_cn/technology/wheelset-test-report`)
    await page.waitForFunction(() => {
      const nuxtApp = document.querySelector('#__nuxt')?.__vue_app__?.$nuxt
      return nuxtApp && !nuxtApp.isHydrating
    })
    await page.evaluate(async () => {
      const vueModuleUrl = performance.getEntriesByType('resource')
        .find(entry => entry.name.includes('vue.runtime.esm-bundler'))?.name
      if (!vueModuleUrl) throw new Error('The Nuxt development page must finish loading before this check runs.')
      const { createVNode, defineComponent, h, nextTick, ref, render } = await import(vueModuleUrl)
      const Trigger = (await import('/_nuxt/components/technology/test-reports/TestReportDialogTriggerButton.vue')).default
      const Dialog = (await import('/_nuxt/components/technology/test-reports/TestReportBlockDialog.vue')).default
      const nuxtApp = document.querySelector('#__nuxt').__vue_app__.$nuxt
      const host = document.createElement('div')
      host.style.cssText = 'padding:180px 16px 16px'
      document.body.prepend(host)
      const primaryDialogOpen = ref(false)
      const secondaryDialogOpen = ref(false)
      const triggerVisible = ref(true)
      const selectedTestId = ref('rim-impact-test')
      const root = defineComponent({
        setup() {
          return () => h('div', [
            triggerVisible.value ? h(Trigger, {
              testId: selectedTestId.value,
              label: '独立测试详情验证',
              'data-regression-test-report-trigger': '',
            }) : null,
            primaryDialogOpen.value ? h(Dialog, {
              key: 'controlled-primary-test-dialog',
              testId: 'rim-spoke-hole-tension-test',
              onClose: () => { primaryDialogOpen.value = false },
            }) : null,
            secondaryDialogOpen.value ? h(Dialog, {
              key: 'controlled-secondary-test-dialog',
              testId: 'rim-impact-test',
              onClose: () => { secondaryDialogOpen.value = false },
            }) : null,
          ])
        },
      })
      const vnode = createVNode(root)
      vnode.appContext = nuxtApp.vueApp._context
      render(vnode, host)
      window.testReportBrowserRegression = {
        host, nextTick, primaryDialogOpen, secondaryDialogOpen, triggerVisible, selectedTestId,
        destroy: () => { render(null, host); host.remove() },
      }
    })
  }

  async function readTestReportOverlayState() {
    return page.evaluate(() => ({
      dialogs: document.querySelectorAll('dialog[open]').length,
      overflow: document.body.style.overflow,
      layers: history.state.__tanzanite_overlay_history__?.stack || [],
      path: location.pathname,
    }))
  }

  async function assertTestReportOverlayResourcesReleased(description) {
    currentScenario = description
    await page.waitForFunction(() => !document.querySelector('dialog[open]')
      && !(history.state.__tanzanite_overlay_history__?.stack?.length))
    const state = await readTestReportOverlayState()
    assertTestReportRegressionCondition(!state.dialogs && !state.overflow && !state.layers.length, description, state)
  }

  try {
    // 测试组件首次加载失败；外壳保持可关闭，恢复网络后可以在弹窗内重试。
    currentScenario = 'component failure: cold page and hydration'
    await mountIndependentTestReportControlsOnColdPage()
    let failTestComponentRequest = true
    await page.route('**/RimImpactTestReportBlock.vue*', route => failTestComponentRequest ? route.abort('failed') : route.continue())
    await page.locator('[data-regression-test-report-trigger]').click()
    currentScenario = 'component failure: error state'
    await page.locator('dialog[open] [data-test-report-loading-state=error]').waitFor()
    const failedComponentState = await readTestReportOverlayState()
    assertTestReportRegressionCondition(failedComponentState.dialogs === 1 && failedComponentState.layers.length === 1, 'Failed content must retain one usable dialog', failedComponentState)
    failTestComponentRequest = false
    await page.locator('dialog[open] [data-test-report-loading-state=error] button').click()
    currentScenario = 'component failure: recovery'
    await page.locator('dialog[open] [data-test-report-id=rim-impact-test]').waitFor()
    assertTestReportRegressionCondition(await page.evaluate(() => document.querySelector('dialog[open]').contains(document.activeElement)), 'Component retry must keep focus in the dialog')
    await page.keyboard.press('Escape')
    await assertTestReportOverlayResourcesReleased('Retry then close must release resources')
    await page.unroute('**/RimImpactTestReportBlock.vue*')
    results.push('component request failure, retry and close')

    // 正文翻译首次加载失败，不得永久停留在加载中。
    let failTestMessageRequests = true
    currentScenario = 'translation failure: cold page and hydration'
    await page.route('**/testReportRim/*.json*', route => failTestMessageRequests ? route.abort('failed') : route.continue())
    await mountIndependentTestReportControlsOnColdPage()
    await page.locator('[data-regression-test-report-trigger]').click()
    currentScenario = 'translation failure: error state'
    await page.locator('dialog[open] [data-test-report-loading-state=error]').waitFor()
    failTestMessageRequests = false
    await page.locator('dialog[open] [data-test-report-loading-state=error] button').click()
    currentScenario = 'translation failure: recovery'
    await page.locator('dialog[open] [data-test-report-id=rim-impact-test]').waitFor()
    assertTestReportRegressionCondition(await page.evaluate(() => document.querySelector('dialog[open]').contains(document.activeElement)), 'Translation retry must keep focus in the dialog')
    await page.keyboard.press('Escape')
    await assertTestReportOverlayResourcesReleased('Translation retry then close must release resources')
    await page.unroute('**/testReportRim/*.json*')
    results.push('translation request failure, retry and close')

    // 测试内容仍在加载时关闭，延迟响应完成后不得再次打开弹窗。
    await mountIndependentTestReportControlsOnColdPage()
    currentScenario = 'delayed content'
    let releaseDelayedTestComponent
    const delayedTestComponent = new Promise(resolve => { releaseDelayedTestComponent = resolve })
    await page.route('**/RimImpactTestReportBlock.vue*', async route => {
      await delayedTestComponent
      await route.continue()
    })
    await page.locator('[data-regression-test-report-trigger]').click()
    await page.locator('dialog[open] [data-test-report-loading-state=loading]').waitFor()
    await page.keyboard.press('Escape')
    await assertTestReportOverlayResourcesReleased('Closing during loading must release resources')
    releaseDelayedTestComponent()
    await page.waitForLoadState('networkidle')
    await page.unroute('**/RimImpactTestReportBlock.vue*')
    await assertTestReportOverlayResourcesReleased('A late response must not reopen the dialog')
    results.push('close while loading and ignore late response')

    // 主动在 history.back 的 popstate 之前重新打开。
    for (let iteration = 0; iteration < 20; iteration += 1) {
      currentScenario = `rapid reopen ${iteration + 1}`
      const rapidReopenState = await page.evaluate(async () => {
        const { nextTick } = window.testReportBrowserRegression
        const trigger = document.querySelector('[data-regression-test-report-trigger]')
        trigger.click()
        await nextTick()
        const historyTraversalFinished = new Promise(resolve => {
          window.addEventListener('popstate', () => setTimeout(resolve, 0), { once: true })
        })
        document.querySelector('dialog[open] .test-report-block-dialog__close').click()
        await nextTick()
        trigger.click()
        await nextTick()
        await historyTraversalFinished
        return {
          dialogs: document.querySelectorAll('dialog[open]').length,
          layers: history.state.__tanzanite_overlay_history__?.stack || [],
          overflow: document.body.style.overflow,
        }
      })
      assertTestReportRegressionCondition(rapidReopenState.dialogs === 1 && rapidReopenState.layers.length === 1 && rapidReopenState.overflow === 'hidden', 'Rapid reopen must retain exactly one dialog', rapidReopenState)
      await page.keyboard.press('Escape')
      await assertTestReportOverlayResourcesReleased('Rapid reopen final close must release resources')
    }
    results.push('20 close and immediate reopen cycles before popstate')

    currentScenario = 'inflation test from an independent button'
    await page.evaluate(() => { window.testReportBrowserRegression.selectedTestId.value = 'rim-inflation-test' })
    await page.locator('[data-regression-test-report-trigger]').click()
    await page.locator('dialog[open] [data-test-report-id=rim-inflation-test]').waitFor()
    const inflationTestState = await page.evaluate(() => {
      const dialog = document.querySelector('dialog[open]')
      return {
        selectedTests: dialog.querySelectorAll('[data-test-report-id]').length,
        title: dialog.querySelector('.rim-test-media-layout__title').textContent.trim(),
        photoSlots: dialog.querySelectorAll('.rim-test-media-layout__photo-frame').length,
      }
    })
    assertTestReportRegressionCondition(inflationTestState.selectedTests === 1 && inflationTestState.title === '打气测试' && inflationTestState.photoSlots === 4, 'An independent button must open only the inflation test', inflationTestState)
    await page.keyboard.press('Escape')
    await assertTestReportOverlayResourcesReleased('The independent inflation dialog must release resources')
    results.push('inflation test from an independent button')

    // 两个原生测试弹窗同时存在时，卸载底层必须按栈顺序清理上层。
    await page.evaluate(() => { window.testReportBrowserRegression.primaryDialogOpen.value = true })
    await page.locator('dialog[open][data-test-report-dialog=rim-spoke-hole-tension-test]').waitFor()
    await page.evaluate(() => { window.testReportBrowserRegression.secondaryDialogOpen.value = true })
    await page.waitForFunction(() => document.querySelectorAll('dialog[open]').length === 2)
    await page.evaluate(() => { window.testReportBrowserRegression.primaryDialogOpen.value = false })
    await assertTestReportOverlayResourcesReleased('Unmounting the lower dialog must also release the upper dialog')
    results.push('two native dialogs and lower dialog unmount')

    // 独立根节点不随 NuxtPage 卸载；加载中的弹窗也必须随路由清理。
    currentScenario = 'route navigation while content is loading'
    await mountIndependentTestReportControlsOnColdPage()
    let releaseNavigatingTestComponent
    const navigatingTestComponent = new Promise(resolve => { releaseNavigatingTestComponent = resolve })
    await page.route('**/RimImpactTestReportBlock.vue*', async route => {
      await navigatingTestComponent
      await route.continue()
    })
    await page.locator('[data-regression-test-report-trigger]').click()
    await page.locator('dialog[open] [data-test-report-loading-state=loading]').waitFor()
    await page.evaluate(async () => {
      await document.querySelector('#__nuxt').__vue_app__.$nuxt.$router.push('/zh_cn/technology/wheel-building-quality-checks')
    })
    await page.waitForURL('**/zh_cn/technology/wheel-building-quality-checks')
    await assertTestReportOverlayResourcesReleased('Route navigation must clean up independent dialogs')
    releaseNavigatingTestComponent()
    await page.waitForLoadState('networkidle')
    await page.unroute('**/RimImpactTestReportBlock.vue*')
    const stateAfterNavigatingResponse = await readTestReportOverlayState()
    assertTestReportRegressionCondition(stateAfterNavigatingResponse.path === '/zh_cn/technology/wheel-building-quality-checks' && !stateAfterNavigatingResponse.dialogs && !stateAfterNavigatingResponse.layers.length && !stateAfterNavigatingResponse.overflow, 'A late response after navigation must not reopen or redirect the page', stateAfterNavigatingResponse)
    results.push('route navigation while loading from an independent root')

    assertTestReportRegressionCondition(!uncaughtBrowserErrors.length, 'No uncaught browser exceptions are permitted', uncaughtBrowserErrors)
    return { passed: results, uncaughtBrowserErrors }
  } catch (error) {
    const state = await readTestReportOverlayState()
    const visibleTests = await page.locator('dialog[open]').evaluateAll(dialogs => dialogs.map(dialog => dialog.dataset.testReportDialog))
    throw new Error(`${currentScenario}; completed: ${results.join(', ')}; state: ${JSON.stringify({ ...state, visibleTests })}; ${error.message}`)
  } finally {
    await page.unrouteAll({ behavior: 'ignoreErrors' })
    await page.evaluate(() => { window.testReportBrowserRegression?.destroy() })
    page.off('pageerror', recordUncaughtBrowserError)
  }
}
