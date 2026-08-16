import { shallowRef } from 'vue'
import type { Router } from 'vue-router'

export type OverlayCloseReason = 'back' | 'user' | 'navigate' | 'replace'
export type OverlayOpenMode = 'replace' | 'push'

interface OverlayHistoryState {
  version: 1
  stack: string[]
}

interface OverlayLayer {
  id: string
  close: (reason: OverlayCloseReason) => void
  resume?: () => void
}

interface OverlayOpenOptions {
  mode?: OverlayOpenMode
  resume?: () => void
}

const OVERLAY_HISTORY_KEY = '__tanzanite_overlay_history__'
const activeLayers = shallowRef<OverlayLayer[]>([])

let browserListenerInstalled = false
let routerListenerInstalled = false
let pendingHistoryResolve: (() => void) | null = null
let overlayInstanceSequence = 0

export const createOverlayInstanceId = (prefix: string) => {
  overlayInstanceSequence += 1
  return `${prefix}:${overlayInstanceSequence}`
}

const isClient = () => typeof window !== 'undefined'

const readOverlayHistoryState = (): OverlayHistoryState | null => {
  if (!isClient()) return null
  const raw = window.history.state?.[OVERLAY_HISTORY_KEY]
  if (!raw || raw.version !== 1 || !Array.isArray(raw.stack)) return null

  return {
    version: 1,
    stack: raw.stack.filter((id: unknown): id is string => typeof id === 'string'),
  }
}

const withOverlayHistoryState = (state: Record<string, unknown>, stack: string[]) => {
  const nextState = { ...state }

  if (stack.length) {
    nextState[OVERLAY_HISTORY_KEY] = {
      version: 1,
      stack,
    } satisfies OverlayHistoryState
  } else {
    delete nextState[OVERLAY_HISTORY_KEY]
  }

  return nextState
}

const replaceCurrentHistoryState = (stack: string[]) => {
  if (!isClient()) return

  window.history.replaceState(
    withOverlayHistoryState(window.history.state || {}, stack),
    '',
    window.location.href,
  )
}

const pushOverlayHistoryState = (stack: string[]) => {
  if (!isClient()) return

  window.history.pushState(
    withOverlayHistoryState(window.history.state || {}, stack),
    '',
    window.location.href,
  )
}

const resolvePendingHistoryOperation = () => {
  const resolve = pendingHistoryResolve
  pendingHistoryResolve = null
  resolve?.()
}

const closeLayerState = (layer: OverlayLayer, reason: OverlayCloseReason) => {
  try {
    layer.close(reason)
  } catch (error) {
    console.error(`[overlay-back-stack] Failed to close "${layer.id}" (${reason})`, error)
  }
}

const resumeTopLayer = () => {
  activeLayers.value.at(-1)?.resume?.()
}

const handlePopState = (event: PopStateEvent) => {
  const historyState = readOverlayHistoryState()

  if (activeLayers.value.length) {
    const closingLayer = activeLayers.value.at(-1)
    if (closingLayer) {
      activeLayers.value = activeLayers.value.slice(0, -1)
      closeLayerState(closingLayer, 'back')
    }

    if (activeLayers.value.length) {
      // The browser has already moved back to the page entry. Recreate one
      // same-URL overlay entry so another back gesture closes the next layer.
      pushOverlayHistoryState(activeLayers.value.map(layer => layer.id))
      resumeTopLayer()
    }

    resolvePendingHistoryOperation()
    return
  }

  // A forward gesture can land on a stale overlay entry after the user closed
  // the overlay. Keep the page usable instead of leaving a ghost history layer.
  if (historyState?.stack.length) {
    window.history.replaceState(
      withOverlayHistoryState(event.state || {}, []),
      '',
      window.location.href,
    )
  }

  resolvePendingHistoryOperation()
}

const ensureBrowserListener = () => {
  if (!isClient() || browserListenerInstalled) return
  browserListenerInstalled = true
  window.addEventListener('popstate', handlePopState)
}

export const installOverlayBackStack = (router?: Router) => {
  if (!isClient()) return

  ensureBrowserListener()

  if (router && !routerListenerInstalled) {
    routerListenerInstalled = true
    router.afterEach(() => {
      const closingLayers = [...activeLayers.value].reverse()
      const hasOverlayHistory = Boolean(readOverlayHistoryState()?.stack.length)

      activeLayers.value = []
      closingLayers.forEach(layer => closeLayerState(layer, 'navigate'))
      if (hasOverlayHistory) {
        replaceCurrentHistoryState([])
      }
      resolvePendingHistoryOperation()
    })
  }
}

export const useOverlayBackStack = () => {
  installOverlayBackStack()

  const isActive = (id: string) => activeLayers.value.some(layer => layer.id === id)

  const open = (
    id: string,
    close: (reason: OverlayCloseReason) => void,
    options: OverlayOpenOptions = {},
  ) => {
    if (!isClient()) return

    ensureBrowserListener()

    const mode = options.mode || 'replace'
    const existingIndex = activeLayers.value.findIndex(layer => layer.id === id)

    if (existingIndex >= 0) {
      const discardedLayers = activeLayers.value.slice(existingIndex + 1).reverse()
      const nextLayers = activeLayers.value.slice(0, existingIndex + 1)
      nextLayers[existingIndex] = {
        id,
        close,
        resume: options.resume,
      }
      activeLayers.value = nextLayers
      discardedLayers.forEach(layer => closeLayerState(layer, 'replace'))
      replaceCurrentHistoryState(nextLayers.map(layer => layer.id))
      return
    }

    const nextLayer: OverlayLayer = {
      id,
      close,
      resume: options.resume,
    }

    if (!activeLayers.value.length) {
      activeLayers.value = [nextLayer]
      pushOverlayHistoryState([id])
      return
    }

    if (mode === 'push') {
      activeLayers.value = [...activeLayers.value, nextLayer]
    } else {
      const previousLayer = activeLayers.value.at(-1)
      if (previousLayer) closeLayerState(previousLayer, 'replace')
      activeLayers.value = [...activeLayers.value.slice(0, -1), nextLayer]
    }

    replaceCurrentHistoryState(activeLayers.value.map(layer => layer.id))
  }

  const setResume = (id: string, resume?: () => void) => {
    const index = activeLayers.value.findIndex(layer => layer.id === id)
    if (index < 0) return

    const nextLayers = [...activeLayers.value]
    const current = nextLayers[index]
    if (!current) return

    nextLayers[index] = {
      ...current,
      resume,
    }
    activeLayers.value = nextLayers
  }

  const close = async (
    id: string,
    reason: OverlayCloseReason = 'user',
  ): Promise<void> => {
    if (!isClient()) return

    const index = activeLayers.value.findIndex(layer => layer.id === id)
    if (index < 0) return

    const closingLayers = activeLayers.value.slice(index).reverse()
    activeLayers.value = activeLayers.value.slice(0, index)
    closingLayers.forEach(layer => closeLayerState(layer, reason))

    if (activeLayers.value.length) {
      replaceCurrentHistoryState(activeLayers.value.map(layer => layer.id))
      resumeTopLayer()
      return
    }

    if (reason === 'navigate') {
      replaceCurrentHistoryState([])
      return
    }

    const hasHistoryEntry = Boolean(readOverlayHistoryState()?.stack.length)
    if (!hasHistoryEntry) return

    await new Promise<void>((resolve) => {
      pendingHistoryResolve = resolve
      window.history.back()
    })
  }

  const closeAll = async (reason: OverlayCloseReason = 'user') => {
    const firstLayer = activeLayers.value[0]
    if (!firstLayer) return
    await close(firstLayer.id, reason)
  }

  return {
    activeLayers,
    isActive,
    open,
    close,
    closeAll,
    setResume,
  }
}
