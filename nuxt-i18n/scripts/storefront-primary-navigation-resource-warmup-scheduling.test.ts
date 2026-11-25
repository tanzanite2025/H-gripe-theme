import assert from 'node:assert/strict'
import { scheduleStorefrontPrimaryNavigationResourceWarmup } from '../app/utils/scheduleStorefrontPrimaryNavigationResourceWarmup.js'

const originalGlobalDescriptors = new Map(
  ['window', 'document', 'performance', 'PerformanceObserver'].map(name => [
    name, Object.getOwnPropertyDescriptor(globalThis, name),
  ]),
)

function createNavigationWarmupBrowserHarness(options: { idleCallbacksSupported?: boolean; lcpSupported?: boolean } = {}) {
  let now = 0
  let nextHandle = 0
  let visibilityState = 'visible'
  let paintObserverCallback: PerformanceObserverCallback | undefined
  const timers = new Map<number, { at: number; callback: () => void }>()
  const idleCallbacks = new Map<number, () => void>()
  const documentEvents = new EventTarget()
  const warmupStarts: number[] = []
  const browserWindow = {
    setTimeout(callback: () => void, delay: number) {
      const handle = ++nextHandle
      timers.set(handle, { at: now + delay, callback })
      return handle
    },
    clearTimeout(handle: number) { timers.delete(handle) },
    ...(options.idleCallbacksSupported === false ? {} : {
      requestIdleCallback(callback: () => void) {
        const handle = ++nextHandle
        idleCallbacks.set(handle, callback)
        return handle
      },
      cancelIdleCallback(handle: number) { idleCallbacks.delete(handle) },
    }),
  }

  class NavigationWarmupTestPerformanceObserver {
    constructor(callback: PerformanceObserverCallback) { paintObserverCallback = callback }
    observe() { if (options.lcpSupported === false) throw new Error('LCP not supported') }
    disconnect() { paintObserverCallback = undefined }
  }

  Object.defineProperties(globalThis, {
    window: { configurable: true, value: browserWindow },
    document: { configurable: true, value: {
      get visibilityState() { return visibilityState },
      addEventListener: documentEvents.addEventListener.bind(documentEvents),
      removeEventListener: documentEvents.removeEventListener.bind(documentEvents),
    } },
    performance: { configurable: true, value: { now: () => now } },
    PerformanceObserver: { configurable: true, value: NavigationWarmupTestPerformanceObserver },
  })

  const cancelWarmup = scheduleStorefrontPrimaryNavigationResourceWarmup(() => warmupStarts.push(now))

  return {
    warmupStarts,
    cancelWarmup,
    advanceBrowserTimeTo(targetTime: number) {
      while (true) {
        const nextTimer = [...timers.entries()].filter(([, timer]) => timer.at <= targetTime)
          .sort((left, right) => left[1].at - right[1].at)[0]
        if (!nextTimer) break
        now = nextTimer[1].at
        timers.delete(nextTimer[0])
        nextTimer[1].callback()
      }
      now = targetTime
    },
    runBrowserIdleCallbacks() {
      const callbacks = [...idleCallbacks.values()]
      idleCallbacks.clear()
      for (const callback of callbacks) callback()
    },
    emitLargestContentfulPaintAt(startTime: number) {
      paintObserverCallback?.({ getEntries: () => [{ startTime }] } as PerformanceObserverEntryList,
        {} as PerformanceObserver)
    },
    setBrowserVisibility(nextVisibility: 'visible' | 'hidden') {
      visibilityState = nextVisibility
      documentEvents.dispatchEvent(new Event('visibilitychange'))
    },
  }
}

try {
  const stablePage = createNavigationWarmupBrowserHarness()
  stablePage.advanceBrowserTimeTo(2499)
  stablePage.runBrowserIdleCallbacks()
  assert.deepEqual(stablePage.warmupStarts, [], 'Do not compete with initial hydration')
  stablePage.advanceBrowserTimeTo(2500)
  stablePage.runBrowserIdleCallbacks()
  stablePage.advanceBrowserTimeTo(12000)
  assert.deepEqual(stablePage.warmupStarts, [2500], 'Stable page warms once, without waiting for window.load')

  const changingPaint = createNavigationWarmupBrowserHarness()
  changingPaint.advanceBrowserTimeTo(2400)
  changingPaint.emitLargestContentfulPaintAt(2400)
  changingPaint.advanceBrowserTimeTo(3199)
  changingPaint.runBrowserIdleCallbacks()
  assert.deepEqual(changingPaint.warmupStarts, [])
  changingPaint.advanceBrowserTimeTo(3200)
  changingPaint.runBrowserIdleCallbacks()
  assert.deepEqual(changingPaint.warmupStarts, [3200], 'Wait for a quiet paint window')

  const paintWhileAwaitingIdle = createNavigationWarmupBrowserHarness()
  paintWhileAwaitingIdle.advanceBrowserTimeTo(2600)
  paintWhileAwaitingIdle.emitLargestContentfulPaintAt(2600)
  paintWhileAwaitingIdle.runBrowserIdleCallbacks()
  assert.deepEqual(paintWhileAwaitingIdle.warmupStarts, [], 'Recheck paint stability inside the idle callback')
  paintWhileAwaitingIdle.advanceBrowserTimeTo(3400)
  paintWhileAwaitingIdle.runBrowserIdleCallbacks()
  assert.deepEqual(paintWhileAwaitingIdle.warmupStarts, [3400])

  const neverIdle = createNavigationWarmupBrowserHarness()
  neverIdle.advanceBrowserTimeTo(10000)
  neverIdle.runBrowserIdleCallbacks()
  assert.deepEqual(neverIdle.warmupStarts, [10000], 'Busy or still-loading page reaches the start deadline')

  const hiddenPage = createNavigationWarmupBrowserHarness()
  hiddenPage.setBrowserVisibility('hidden')
  hiddenPage.advanceBrowserTimeTo(15000)
  assert.deepEqual(hiddenPage.warmupStarts, [], 'Avoid downloading in a hidden tab')
  hiddenPage.setBrowserVisibility('visible')
  assert.deepEqual(hiddenPage.warmupStarts, [15000], 'Resume overdue warmup immediately when visible')

  const unmountedHeader = createNavigationWarmupBrowserHarness()
  unmountedHeader.advanceBrowserTimeTo(2500)
  unmountedHeader.cancelWarmup()
  unmountedHeader.advanceBrowserTimeTo(12000)
  unmountedHeader.runBrowserIdleCallbacks()
  assert.deepEqual(unmountedHeader.warmupStarts, [], 'Unmount cancels timers and idle work')

  const unsupportedBrowserApis = createNavigationWarmupBrowserHarness({ idleCallbacksSupported: false, lcpSupported: false })
  unsupportedBrowserApis.advanceBrowserTimeTo(2500)
  assert.deepEqual(unsupportedBrowserApis.warmupStarts, [2500], 'Unsupported APIs retain a timed fallback')
} finally {
  for (const [name, descriptor] of originalGlobalDescriptors) {
    if (descriptor) Object.defineProperty(globalThis, name, descriptor)
    else Reflect.deleteProperty(globalThis, name)
  }
}

console.log('Primary navigation warmup scheduling regression checks passed.')
