export const STOREFRONT_CLIENT_OVERLAYS_EVENT = 'storefront:activate-client-overlays'

export const activateStorefrontClientOverlays = () => {
  if (!import.meta.client || typeof window === 'undefined') return
  window.dispatchEvent(new Event(STOREFRONT_CLIENT_OVERLAYS_EVENT))
}
