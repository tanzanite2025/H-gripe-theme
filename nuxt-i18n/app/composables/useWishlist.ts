import { ref, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'

export interface WishlistItem {
  id: number
  product_id: number
  created_at: string
  product?: unknown
}

const items = ref<WishlistItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const loadedOnce = ref(false)

const wishlistMessage = (err: unknown, fallback: string) => {
  if (err instanceof Error && err.message) return err.message
  return fallback
}

export const useWishlist = () => {
  const auth = useAuth()

  const ensureAuthenticated = async () => {
    if (!auth.initialized.value) {
      await auth.ensureSession()
    }

    return auth.isAuthenticated.value
  }

  const loadWishlist = async () => {
    if (!(await ensureAuthenticated())) {
      items.value = []
      error.value = 'Please log in to view your wishlist.'
      loadedOnce.value = true
      return
    }
    if (loading.value) return
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
      items.value = Array.isArray(response?.items) ? response.items : []
      loadedOnce.value = true
    } catch (e: unknown) {
      console.error('Failed to load wishlist:', e)
      error.value = wishlistMessage(e, 'Failed to load wishlist.')
    } finally {
      loading.value = false
    }
  }

  const addToWishlist = async (productId: number) => {
    if (!productId) return { success: false, message: 'Invalid product id' }
    if (!(await ensureAuthenticated())) {
      const message = 'Please log in to use wishlist.'
      error.value = message
      return { success: false, message }
    }
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
      if (item) {
        const exists = items.value.find((x) => x.id === item.id)
        if (!exists) {
          items.value.unshift(item)
        }
      }
      return { success: true, item }
    } catch (e: unknown) {
      console.error('Failed to add to wishlist:', e)
      const message = wishlistMessage(e, 'Failed to add to wishlist.')
      error.value = message
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
      items.value = items.value.filter((item) => item.id !== wishlistId)
      return { success: true }
    } catch (e: unknown) {
      console.error('Failed to remove from wishlist:', e)
      const message = wishlistMessage(e, 'Failed to remove from wishlist.')
      error.value = message
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
