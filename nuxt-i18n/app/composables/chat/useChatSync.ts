/**
 * 聊天消息同步服务
 * 负责将聊天消息同步到后端，支持离线缓存和自动重试
 */

import { ref } from 'vue'

export interface ChatMessage {
  id: string
  content: string
  role: 'user' | 'agent' | 'system'
  timestamp: number
  agentId?: string
  metadata?: Record<string, any>
}

export interface SyncResult {
  success: boolean
  error?: string
  synced?: number
  failed?: number
}

const pendingSyncQueue = ref<Array<{ sessionId: string; message: ChatMessage }>>([])
const isSyncing = ref(false)

export const useChatSync = () => {
  const { request } = useApiRequest()

  /**
   * 保存消息到本地存储（快速响应）
   */
  const saveMessageLocally = (sessionId: string, message: ChatMessage) => {
    if (!import.meta.client) return

    try {
      const key = `tz_chat_${sessionId}`
      const data = localStorage.getItem(key)
      const chat = data ? JSON.parse(data) : { messages: [] }

      // 避免重复消息
      const exists = chat.messages.some((m: ChatMessage) => m.id === message.id)
      if (!exists) {
        chat.messages.push(message)
        localStorage.setItem(key, JSON.stringify(chat))
      }
    } catch (e) {
      console.error('[ChatSync] Failed to save message locally:', e)
    }
  }

  /**
   * 同步单条消息到后端
   */
  const syncMessageToBackend = async (
    sessionId: string,
    message: ChatMessage
  ): Promise<{ success: boolean; error?: string }> => {
    try {
      await request('/chat/messages', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          accept: 'application/json',
        },
        body: JSON.stringify({
          session_id: sessionId,
          message: {
            id: message.id,
            content: message.content,
            role: message.role,
            timestamp: message.timestamp,
            agent_id: message.agentId,
            metadata: message.metadata
          },
        }),
      }, 'Failed to sync chat message')

      return { success: true }
    } catch (e: any) {
      console.error('[ChatSync] Failed to sync message:', e)
      return {
        success: false,
        error: e.message || 'Unknown error'
      }
    }
  }

  /**
   * 保存消息（先本地，后异步同步到后端）
   */
  const saveMessage = async (sessionId: string, message: ChatMessage) => {
    // 1. 立即保存到本地（快速响应）
    saveMessageLocally(sessionId, message)

    // 2. 异步同步到后端
    if (import.meta.client) {
      // 添加到同步队列
      pendingSyncQueue.value.push({ sessionId, message })

      // 触发同步（防抖，避免频繁请求）
      setTimeout(() => {
        processSyncQueue()
      }, 500)
    }
  }

  /**
   * 处理同步队列
   */
  const processSyncQueue = async () => {
    if (isSyncing.value || pendingSyncQueue.value.length === 0) {
      return
    }

    isSyncing.value = true

    try {
      // 批量同步
      const queue = [...pendingSyncQueue.value]
      pendingSyncQueue.value = []

      const results = await Promise.allSettled(
        queue.map(({ sessionId, message }) =>
          syncMessageToBackend(sessionId, message)
        )
      )

      // 检查失败的消息，重新加入队列
      const failedItems = queue.filter((_, index) => {
        const result = results[index]
        return result.status === 'rejected' || 
               (result.status === 'fulfilled' && !result.value.success)
      })

      if (failedItems.length > 0) {
        console.warn(`[ChatSync] ${failedItems.length} messages failed, will retry later`)
        // 重新加入队列（但限制最多重试5次）
        failedItems.forEach(item => {
          const retryCount = (item.message.metadata?.retryCount || 0) + 1
          if (retryCount <= 5) {
            pendingSyncQueue.value.push({
              sessionId: item.sessionId,
              message: {
                ...item.message,
                metadata: {
                  ...item.message.metadata,
                  retryCount
                }
              }
            })
          }
        })
      }
    } finally {
      isSyncing.value = false
    }
  }

  /**
   * 从后端加载聊天历史
   */
  const loadChatHistory = async (
    sessionId: string
  ): Promise<{ messages: ChatMessage[]; fromCache: boolean }> => {
    try {
      // 优先从后端加载
      const response = await request<{ messages: ChatMessage[]; count: number }>(
        `/chat/messages?session_id=${encodeURIComponent(sessionId)}&limit=100`,
        {
          headers: { accept: 'application/json' },
        },
        'Failed to load chat history'
      )

      if (response.messages && response.messages.length > 0) {
        console.log(`[ChatSync] Loaded ${response.count} messages from backend`)
        return { messages: response.messages, fromCache: false }
      }
    } catch (e) {
      console.warn('[ChatSync] Failed to load from backend, using local cache:', e)
    }

    // 降级到本地缓存
    if (import.meta.client) {
      try {
        const key = `tz_chat_${sessionId}`
        const data = localStorage.getItem(key)
        if (data) {
          const chat = JSON.parse(data)
          console.log(`[ChatSync] Loaded ${chat.messages?.length || 0} messages from cache`)
          return { messages: chat.messages || [], fromCache: true }
        }
      } catch (e) {
        console.error('[ChatSync] Failed to load from cache:', e)
      }
    }

    return { messages: [], fromCache: true }
  }

  /**
   * 手动触发全部同步（用于网络恢复后）
   */
  const syncAll = async (): Promise<SyncResult> => {
    if (!import.meta.client) {
      return { success: false, error: 'Not in client' }
    }

    try {
      // 获取所有本地聊天数据
      const keys = Object.keys(localStorage).filter(key =>
        key.startsWith('tz_chat_')
      )

      let synced = 0
      let failed = 0

      for (const key of keys) {
        const sessionId = key.replace('tz_chat_', '')
        const data = localStorage.getItem(key)
        if (!data) continue

        try {
          const chat = JSON.parse(data)
          const messages = chat.messages || []

          for (const message of messages) {
            const result = await syncMessageToBackend(sessionId, message)
            if (result.success) {
              synced++
            } else {
              failed++
            }
          }
        } catch (e) {
          console.error(`[ChatSync] Failed to sync session ${sessionId}:`, e)
          failed++
        }
      }

      return { success: true, synced, failed }
    } catch (e: any) {
      return { success: false, error: e.message }
    }
  }

  /**
   * 清理过期的本地聊天数据（超过30天）
   */
  const cleanupOldChats = () => {
    if (!import.meta.client) return

    try {
      const keys = Object.keys(localStorage).filter(key =>
        key.startsWith('tz_chat_')
      )

      const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000

      keys.forEach(key => {
        const data = localStorage.getItem(key)
        if (!data) return

        try {
          const chat = JSON.parse(data)
          const messages = chat.messages || []

          // 检查最后一条消息的时间
          const lastMessage = messages[messages.length - 1]
          if (lastMessage && lastMessage.timestamp < thirtyDaysAgo) {
            localStorage.removeItem(key)
            console.log(`[ChatSync] Cleaned up old chat: ${key}`)
          }
        } catch (e) {
          console.error(`[ChatSync] Failed to parse chat ${key}:`, e)
        }
      })
    } catch (e) {
      console.error('[ChatSync] Failed to cleanup old chats:', e)
    }
  }

  return {
    saveMessage,
    loadChatHistory,
    syncAll,
    cleanupOldChats,
    pendingSyncQueue,
    isSyncing
  }
}
