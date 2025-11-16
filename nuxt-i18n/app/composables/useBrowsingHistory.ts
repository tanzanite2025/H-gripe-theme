import { ref, computed } from 'vue'

interface BrowsingHistoryItem {
  id: number
  title: string
  thumbnail: string
  price: string
  url: string
  viewedAt: string
}

const STORAGE_KEY = 'tz_browsing_history'
const MAX_ITEMS = 20

export const useBrowsingHistory = () => {
  const history = ref<BrowsingHistoryItem[]>([])

  // 从 localStorage 加载历史记录
  const loadHistory = () => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        const parsed = JSON.parse(stored)
        history.value = Array.isArray(parsed) ? parsed : []
      }
    } catch (error) {
      console.error('加载浏览历史失败:', error)
      history.value = []
    }
  }

  // 保存历史记录到 localStorage
  const saveHistory = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(history.value))
    } catch (error) {
      console.error('保存浏览历史失败:', error)
    }
  }

  // 添加商品到浏览历史
  const addToHistory = (item: Omit<BrowsingHistoryItem, 'viewedAt'>) => {
    try {
      // 移除已存在的相同商品
      history.value = history.value.filter(h => h.id !== item.id)
      
      // 添加到开头
      history.value.unshift({
        ...item,
        viewedAt: new Date().toISOString()
      })
      
      // 限制数量
      if (history.value.length > MAX_ITEMS) {
        history.value = history.value.slice(0, MAX_ITEMS)
      }
      
      saveHistory()
    } catch (error) {
      console.error('添加浏览历史失败:', error)
    }
  }

  // 清空浏览历史
  const clearHistory = () => {
    try {
      history.value = []
      localStorage.removeItem(STORAGE_KEY)
    } catch (error) {
      console.error('清空浏览历史失败:', error)
    }
  }

  // 移除单个商品
  const removeItem = (id: number) => {
    try {
      history.value = history.value.filter(h => h.id !== id)
      saveHistory()
    } catch (error) {
      console.error('移除浏览历史失败:', error)
    }
  }

  // 获取历史记录数量
  const historyCount = computed(() => history.value.length)

  // 检查是否有历史记录
  const hasHistory = computed(() => history.value.length > 0)

  // 初始化时加载历史记录
  if (typeof window !== 'undefined') {
    loadHistory()
  }

  return {
    history,
    historyCount,
    hasHistory,
    addToHistory,
    clearHistory,
    removeItem,
    loadHistory
  }
}
