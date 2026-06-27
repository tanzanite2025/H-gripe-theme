import { ref, computed } from 'vue'

interface BrowsingHistoryItem {
  id: number
  title: string
  thumbnail: string
  price: string
  url: string
  viewedAt: string
}

interface BrowsingHistoryBackend {
  id: number
  user_id: number
  product_id: number
  view_count: number
  last_viewed_at: string
  created_at: string
  updated_at: string
}

const STORAGE_KEY = 'tz_browsing_history'
const MAX_ITEMS = 20
const BACKEND_API = '/api/v1/user/browsing-history'

export const useBrowsingHistory = () => {
  const history = ref<BrowsingHistoryItem[]>([])
  const { isAuthenticated } = useAuth()
  
  let syncTimeout: NodeJS.Timeout | null = null
  let syncQueue: Set<number> = new Set()

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

  // 从后端加载浏览历史（登录用户）
  const loadFromBackend = async () => {
    if (!isAuthenticated.value) return
    
    try {
      const response = await fetch(`${BACKEND_API}?limit=${MAX_ITEMS}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      
      if (data.history && Array.isArray(data.history)) {
        // 后端返回的是 product_id，需要转换为前端格式
        // 这里只保存 product_id，前端需要根据 product_id 获取产品详情
        // 或者后端可以扩展返回产品详情
        const backendHistory: BrowsingHistoryBackend[] = data.history
        
        // 合并后端数据到本地
        // 注意：这里需要产品详情，暂时保持本地优先
        // 实际使用时应该让后端返回产品详情或前端批量查询
      }
    } catch (error) {
      console.error('从后端加载浏览历史失败:', error)
    }
  }

  // 同步单个产品到后端（去重、批量）
  const syncToBackend = async (productID: number) => {
    if (!isAuthenticated.value) return
    
    // 添加到同步队列
    syncQueue.add(productID)
    
    // 清除之前的定时器
    if (syncTimeout) {
      clearTimeout(syncTimeout)
    }
    
    // 500ms 后批量同步
    syncTimeout = setTimeout(async () => {
      const idsToSync = Array.from(syncQueue)
      syncQueue.clear()
      
      for (const id of idsToSync) {
        try {
          const response = await fetch(BACKEND_API, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({ product_id: id })
          })
          
          if (!response.ok) {
            console.error(`同步产品 ${id} 到后端失败:`, response.status)
          }
        } catch (error) {
          console.error(`同步产品 ${id} 到后端失败:`, error)
        }
      }
    }, 500)
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
      
      // 异步同步到后端（登录用户）
      if (isAuthenticated.value) {
        syncToBackend(item.id)
      }
    } catch (error) {
      console.error('添加浏览历史失败:', error)
    }
  }

  // 清空浏览历史
  const clearHistory = async () => {
    try {
      history.value = []
      localStorage.removeItem(STORAGE_KEY)
      
      // 同步清空到后端
      if (isAuthenticated.value) {
        await fetch(BACKEND_API, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
            'Content-Type': 'application/json'
          }
        })
      }
    } catch (error) {
      console.error('清空浏览历史失败:', error)
    }
  }

  // 移除单个商品
  const removeItem = async (id: number) => {
    try {
      history.value = history.value.filter(h => h.id !== id)
      saveHistory()
      
      // 同步删除到后端
      if (isAuthenticated.value) {
        await fetch(`${BACKEND_API}/${id}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
            'Content-Type': 'application/json'
          }
        })
      }
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
    
    // 登录用户从后端加载
    if (isAuthenticated.value) {
      loadFromBackend()
    }
  }

  return {
    history,
    historyCount,
    hasHistory,
    addToHistory,
    clearHistory,
    removeItem,
    loadHistory,
    loadFromBackend
  }
}
