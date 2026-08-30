import { ref } from 'vue'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'

const isPagesSearchOpen = ref(false)
const pagesSearchQuery = ref('')
const overlayBackStack = useOverlayBackStack()

const closePagesSearchState = () => {
  isPagesSearchOpen.value = false
  pagesSearchQuery.value = ''
}

export const usePagesSearchOverlayState = () => {
  const openPagesSearch = (initialQuery = '') => {
    pagesSearchQuery.value = initialQuery.trim()
    isPagesSearchOpen.value = true
    overlayBackStack.open('pages-search', closePagesSearchState)
  }

  const closePagesSearch = (reason: 'user' | 'navigate' | 'replace' = 'user') => {
    void overlayBackStack.close('pages-search', reason)
    closePagesSearchState()
  }

  return {
    isPagesSearchOpen,
    pagesSearchQuery,
    openPagesSearch,
    closePagesSearch,
  }
}
