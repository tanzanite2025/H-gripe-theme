import { computed, watch } from 'vue'
import { useState } from 'nuxt/app'
import { useAuth } from '~/composables/useAuth'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'

export interface WishlistItem {
  id: number
  product_id: number
  product?: {
    id: number
    name: string
    slug: string
    price_decimal: string
    sale_price_decimal?: string | null
    availability: 'in_stock' | 'out_of_stock'
    thumbnail?: string
  }
}

const wishlistMessage = (err: unknown, fallback: string) => {
  if (err instanceof Error && err.message) return err.message
  return fallback
}

export const useWishlist = () => {
  const auth = useAuth()
  const { track: trackBehaviorEvent } = useBehaviorEvents()
  const fallbackOwnerKeys = new WeakMap<object, string>()
  let nextFallbackOwnerId = 0
  const items = useState<WishlistItem[]>('wishlist-items', () => [])
  const loading = useState<boolean>('wishlist-loading', () => false)
  const error = useState<string | null>('wishlist-error', () => null)
  const loadedOnce = useState<boolean>('wishlist-loaded-once', () => false)
  const ownerKey = useState<string | null | undefined>('wishlist-owner-key', () => undefined)
  const stateVersion = useState<number>('wishlist-state-version', () => 0)
  const getOwnerKey = (user: { id?: number; email?: string; username?: string } | null | undefined) => {
    if (!user) return null
    if (user.id !== undefined && user.id !== null) return `id:${user.id}`
    if (user.email) return `email:${user.email}`
    if (user.username) return `username:${user.username}`
    let key = fallbackOwnerKeys.get(user)
    if (!key) {
      key = `object:${++nextFallbackOwnerId}`
      fallbackOwnerKeys.set(user, key)
    }
    return key
  }
  const resetWishlistState = (nextOwnerKey: string | null) => {
    if (ownerKey.value === nextOwnerKey) return
    ownerKey.value = nextOwnerKey
    items.value = []
    loading.value = false
    error.value = null
    loadedOnce.value = false
    stateVersion.value += 1
  }

  watch(() => getOwnerKey(auth.user.value), resetWishlistState, { immediate: true, flush: 'sync' })

  const ensureAuthenticated = async () => {
    if (!auth.initialized.value) {
      await auth.ensureSession()
    }

    return auth.isAuthenticated.value
  }

  const loadWishlist = async () => {
    if (!(await ensureAuthenticated())) {
      if (ownerKey.value === null) {
        error.value = 'Please log in to view your wishlist.'
        loadedOnce.value = true
      }
      return
    }
    resetWishlistState(getOwnerKey(auth.user.value))
    if (loading.value) return
    const requestVersion = stateVersion.value
    const requestOwnerKey = ownerKey.value
    loading.value = true
    error.value = null
    try {
      const response = await auth.request<{ items: WishlistItem[] }>(
        '/wishlist',
        {
          headers: { accept: 'application/json' },
        },
        'Please log in to view your wishlist.'
      )
      if (stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        items.value = Array.isArray(response?.items) ? response.items : []
        loadedOnce.value = true
      }
    } catch (e: unknown) {
      console.error('Failed to load wishlist:', e)
      if (stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        error.value = wishlistMessage(e, 'Failed to load wishlist.')
      }
    } finally {
      if (stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        loading.value = false
      }
    }
  }

  const addToWishlist = async (productId: number) => {
    if (!productId) return { success: false, message: 'Invalid product id' }
    if (!(await ensureAuthenticated())) {
      const message = 'Please log in to use wishlist.'
      error.value = message
      return { success: false, message }
    }
    resetWishlistState(getOwnerKey(auth.user.value))
    const requestVersion = stateVersion.value
    const requestOwnerKey = ownerKey.value
    error.value = null
    try {
      const response = await auth.request<{ item: WishlistItem }>(
        '/wishlist',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ product_id: productId }),
        },
        'Please log in to use wishlist.'
      )
      const item = response?.item
      if (item && stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        const exists = items.value.find((x) => x.id === item.id)
        if (!exists) {
          items.value.unshift(item)
        }
      }
      if (stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        trackBehaviorEvent({
          eventType: 'wishlist_add',
          productId,
          metadata: {
            source: 'wishlist_action',
          },
        })
      }
      return { success: true, item }
    } catch (e: unknown) {
      console.error('Failed to add to wishlist:', e)
      const message = wishlistMessage(e, 'Failed to add to wishlist.')
      if (stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        error.value = message
      }
      return { success: false, message }
    }
  }

  const removeFromWishlist = async (wishlistId: number) => {
    if (!wishlistId) return { success: false, message: 'Invalid wishlist id' }
    if (!(await ensureAuthenticated())) {
      const message = 'Please log in to use wishlist.'
      error.value = message
      return { success: false, message }
    }
    resetWishlistState(getOwnerKey(auth.user.value))
    const requestVersion = stateVersion.value
    const requestOwnerKey = ownerKey.value
    error.value = null
    try {
      await auth.request(
        `/wishlist/${wishlistId}`,
        {
          method: 'DELETE',
          headers: { accept: 'application/json' },
        },
        'Please log in to use wishlist.'
      )
      if (stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        items.value = items.value.filter((item) => item.id !== wishlistId)
      }
      return { success: true }
    } catch (e: unknown) {
      console.error('Failed to remove from wishlist:', e)
      const message = wishlistMessage(e, 'Failed to remove from wishlist.')
      if (stateVersion.value === requestVersion && ownerKey.value === requestOwnerKey) {
        error.value = message
      }
      return { success: false, message }
    }
  }

  return {
    // state
    items,
    loading,
    error,
    loadedOnce: computed(() => loadedOnce.value),

    // actions
    loadWishlist,
    addToWishlist,
    removeFromWishlist,
  }
}
