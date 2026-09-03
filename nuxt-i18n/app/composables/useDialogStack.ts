import { shallowRef } from 'vue'

type DialogEscapeHandler = (event: KeyboardEvent) => boolean | void

interface DialogStackLayer {
  id: string
  onEscape: DialogEscapeHandler
  order: number
  priority: number
}

interface DialogStackRegisterOptions {
  priority?: number
}

const activeDialogLayers = shallowRef<DialogStackLayer[]>([])
let keyboardListenerInstalled = false
let dialogInstanceSequence = 0
let dialogLayerOrder = 0

const isClient = () => typeof document !== 'undefined'

export const createDialogStackId = (prefix: string) => {
  dialogInstanceSequence += 1
  return `${prefix}:${dialogInstanceSequence}`
}

const unregisterDialogLayer = (id: string) => {
  activeDialogLayers.value = activeDialogLayers.value.filter(layer => layer.id !== id)
}

const topDialogLayer = () => {
  return activeDialogLayers.value.reduce<DialogStackLayer | null>((topLayer, layer) => {
    if (!topLayer) return layer
    if (layer.priority > topLayer.priority) return layer
    if (layer.priority === topLayer.priority && layer.order > topLayer.order) return layer
    return topLayer
  }, null)
}

const handleDialogStackKeydown = (event: KeyboardEvent) => {
  if (event.key !== 'Escape' || event.defaultPrevented || event.isComposing) return

  const topLayer = topDialogLayer()
  if (!topLayer) return

  const handled = topLayer.onEscape(event)
  if (handled === false) return

  event.preventDefault()
  event.stopPropagation()
  event.stopImmediatePropagation()
}

const ensureKeyboardListener = () => {
  if (!isClient() || keyboardListenerInstalled) return
  keyboardListenerInstalled = true
  document.addEventListener('keydown', handleDialogStackKeydown)
}

export const useDialogStack = () => {
  const register = (
    id: string,
    onEscape: DialogEscapeHandler,
    options: DialogStackRegisterOptions = {},
  ) => {
    if (!isClient()) return () => {}

    ensureKeyboardListener()
    unregisterDialogLayer(id)
    dialogLayerOrder += 1
    activeDialogLayers.value = [
      ...activeDialogLayers.value,
      {
        id,
        onEscape,
        order: dialogLayerOrder,
        priority: typeof options.priority === 'number' && Number.isFinite(options.priority) ? options.priority : 0,
      },
    ]

    return () => unregisterDialogLayer(id)
  }

  const unregister = (id: string) => {
    unregisterDialogLayer(id)
  }

  const isTop = (id: string) => topDialogLayer()?.id === id

  return {
    activeDialogLayers,
    isTop,
    register,
    unregister,
  }
}
