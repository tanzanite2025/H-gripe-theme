import { onBeforeUnmount, onMounted, ref } from 'vue'

/**
 * Shared visibility state for fixed navigation chrome.
 *
 * Navigation stays mounted and fixed so it never participates in document
 * scrolling. While the user is actively scrolling it moves out of the way;
 * after a short idle period it returns automatically.
 */
export const useScrollIdleVisibility = (idleDelay = 180) => {
  const isScrolling = ref(false)
  let idleTimer: ReturnType<typeof setTimeout> | null = null

  const clearIdleTimer = () => {
    if (!idleTimer) return
    clearTimeout(idleTimer)
    idleTimer = null
  }

  const handleScroll = () => {
    isScrolling.value = true
    clearIdleTimer()
    idleTimer = setTimeout(() => {
      isScrolling.value = false
      idleTimer = null
    }, idleDelay)
  }

  onMounted(() => {
    window.addEventListener('scroll', handleScroll, { passive: true })
  })

  onBeforeUnmount(() => {
    window.removeEventListener('scroll', handleScroll)
    clearIdleTimer()
  })

  return { isScrolling }
}
