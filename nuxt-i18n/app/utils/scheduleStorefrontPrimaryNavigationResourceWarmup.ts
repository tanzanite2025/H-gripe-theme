import { STOREFRONT_PRIMARY_NAVIGATION_RESOURCE_WARMUP } from './storefrontLoadingPolicy'

// The deadline starts when the header DOM is ready, independently of Nuxt's
// root Suspense, the page component's mounted hook, and the window load event.
// A hidden tab waits until visible; timers cannot guarantee wall-clock execution
// while a browser throttles a tab or its main thread is blocked.
export function scheduleStorefrontPrimaryNavigationResourceWarmup(startNavigationResourceWarmup: () => void): () => void {
  if (typeof window === 'undefined' || typeof document === 'undefined') return () => {}

  const policy = STOREFRONT_PRIMARY_NAVIGATION_RESOURCE_WARMUP
  const navigationReadyAt = performance.now()
  let latestLargestContentfulPaintAt = navigationReadyAt
  let finished = false
  let stabilityTimer: number | undefined
  let deadlineTimer: number | undefined
  let idleCallbackHandle: number | undefined
  let largestContentfulPaintObserver: PerformanceObserver | undefined

  function cancelNavigationWarmupScheduling(): void {
    finished = true
    window.clearTimeout(stabilityTimer)
    window.clearTimeout(deadlineTimer)
    if (idleCallbackHandle !== undefined) window.cancelIdleCallback?.(idleCallbackHandle)
    largestContentfulPaintObserver?.disconnect()
    document.removeEventListener('visibilitychange', resumeNavigationWarmupWhenVisible)
  }

  function startNavigationWarmupOnce(): void {
    if (finished || document.visibilityState === 'hidden') return
    cancelNavigationWarmupScheduling()
    startNavigationResourceWarmup()
  }

  function scheduleNavigationWarmupAfterStablePaint(): void {
    if (finished || document.visibilityState === 'hidden') return
    window.clearTimeout(stabilityTimer)
    if (idleCallbackHandle !== undefined) {
      window.cancelIdleCallback?.(idleCallbackHandle)
      idleCallbackHandle = undefined
    }
    const stablePaintDelay = Math.max(
      navigationReadyAt + policy.minimumDelayMs,
      latestLargestContentfulPaintAt + policy.paintQuietWindowMs,
    ) - performance.now()
    stabilityTimer = window.setTimeout(() => {
      if (typeof window.requestIdleCallback === 'function') {
        idleCallbackHandle = window.requestIdleCallback(() => {
          idleCallbackHandle = undefined
          // A newer LCP can arrive while waiting for an idle callback.
          if (performance.now() - latestLargestContentfulPaintAt < policy.paintQuietWindowMs) {
            scheduleNavigationWarmupAfterStablePaint()
          } else {
            startNavigationWarmupOnce()
          }
        })
      } else {
        startNavigationWarmupOnce()
      }
    }, Math.max(0, stablePaintDelay))
  }

  function resumeNavigationWarmupWhenVisible(): void {
    if (document.visibilityState === 'hidden') return
    if (performance.now() - navigationReadyAt >= policy.maximumStartDelayMs) startNavigationWarmupOnce()
    else scheduleNavigationWarmupAfterStablePaint()
  }

  if (typeof PerformanceObserver !== 'undefined') {
    try {
      largestContentfulPaintObserver = new PerformanceObserver(entries => {
        for (const entry of entries.getEntries()) {
          latestLargestContentfulPaintAt = Math.max(latestLargestContentfulPaintAt, entry.startTime)
        }
        scheduleNavigationWarmupAfterStablePaint()
      })
      largestContentfulPaintObserver.observe({ type: 'largest-contentful-paint', buffered: true })
    } catch {
      largestContentfulPaintObserver?.disconnect()
    }
  }

  document.addEventListener('visibilitychange', resumeNavigationWarmupWhenVisible)
  deadlineTimer = window.setTimeout(startNavigationWarmupOnce, policy.maximumStartDelayMs)
  scheduleNavigationWarmupAfterStablePaint()
  return cancelNavigationWarmupScheduling
}
