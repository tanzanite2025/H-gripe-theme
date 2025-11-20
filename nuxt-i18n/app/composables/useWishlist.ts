import { ref, computed } from 'vue'

export interface WishlistItem {
  id: number
  product_id: number
  created_at: string
  product: any
}

const items = ref<WishlistItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
let loadedOnce = false

export const useWishlist = () => {
  const config = useRuntimeConfig()
  const apiBase = computed(() => {
    const base = (config.public as { wpApiBase?: string }).wpApiBase || '/wp-json'
    return base.replace(/\/$/, '')
  })

  const loadWishlist = async () => {
    if (loading.value) return
    loading.value = true
    error.value = null
    try {
      const response = await $fetch<{ items: WishlistItem[] }>(
        `${apiBase.value}/tanzanite/v1/wishlist`,
        {
          credentials: 'include',
          headers: { accept: 'application/json' },
        },
      )
      items.value = Array.isArray(response?.items) ? response.items : []
      loadedOnce = true
    } catch (e: any) {
      console.error('Failed to load wishlist:', e)
      error.value = e?.data?.message || 'Failed to load wishlist.'
    } finally {
      loading.value = false
    }
  }

  const addToWishlist = async (productId: number) => {
    if (!productId) return { success: false, message: 'Invalid product id' }
    error.value = null
    try {
      const response = await $fetch<{ item: WishlistItem }>(
        `${apiBase.value}/tanzanite/v1/wishlist`,
        {
          method: 'POST',
          credentials: 'include',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: { product_id: productId },
        },
      )
      const item = response?.item
      if (item) {
        const exists = items.value.find((x) => x.id === item.id)
        if (!exists) {
          items.value.unshift(item)
        }
      }
      return { success: true, item }
    } catch (e: any) {
      console.error('Failed to add to wishlist:', e)
      const message = e?.data?.message || 'Failed to add to wishlist.'
      error.value = message
      return { success: false, message }
    }
  }

  const removeFromWishlist = async (wishlistId: number) => {
    if (!wishlistId) return { success: false, message: 'Invalid wishlist id' }
    error.value = null
    try {
      await $fetch(
        `${apiBase.value}/tanzanite/v1/wishlist/${wishlistId}`,
        {
          method: 'DELETE',
          credentials: 'include',
          headers: { accept: 'application/json' },
        },
      )
      items.value = items.value.filter((item) => item.id !== wishlistId)
      return { success: true }
    } catch (e: any) {
      console.error('Failed to remove from wishlist:', e)
      const message = e?.data?.message || 'Failed to remove from wishlist.'
      error.value = message
      return { success: false, message }
    }
  }

  return {
    // state
    items,
    loading,
    error,
    loadedOnce: computed(() => loadedOnce),

    // actions
    loadWishlist,
    addToWishlist,
    removeFromWishlist,
  }
}
