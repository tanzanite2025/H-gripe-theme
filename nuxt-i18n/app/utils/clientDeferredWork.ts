type WindowWithIdleCallback = Window & {
  requestIdleCallback?: (
    callback: IdleRequestCallback,
    options?: IdleRequestOptions,
  ) => number
  cancelIdleCallback?: (handle: number) => void
}

interface DeferredClientWorkOptions {
  delayMs?: number
  idleTimeoutMs?: number
}

export const scheduleDeferredClientWork = (
  task: () => void,
  options: DeferredClientWorkOptions = {},
) => {
  if (!import.meta.client || typeof window === 'undefined') {
    return () => {}
  }

  const delayMs = Math.max(0, options.delayMs ?? 0)
  const idleTimeoutMs = Math.max(0, options.idleTimeoutMs ?? 3000)
  const browserWindow = window as WindowWithIdleCallback
  let cancelled = false
  let idleHandle: number | null = null

  const run = () => {
    if (cancelled) return

    if (typeof browserWindow.requestIdleCallback === 'function') {
      idleHandle = browserWindow.requestIdleCallback(() => {
        idleHandle = null
        if (!cancelled) task()
      }, { timeout: idleTimeoutMs })
      return
    }

    task()
  }

  const timeoutHandle = window.setTimeout(run, delayMs)

  return () => {
    cancelled = true
    window.clearTimeout(timeoutHandle)
    if (idleHandle !== null && typeof browserWindow.cancelIdleCallback === 'function') {
      browserWindow.cancelIdleCallback(idleHandle)
    }
  }
}
