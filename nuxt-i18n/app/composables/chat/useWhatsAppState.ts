import { ref, watch, nextTick, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { useMembership } from '~/composables/useMembership'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import AuthModal from '~/components/AuthModal.vue'
import AgentChatPanel from '~/components/whatsapp/AgentChatPanel.vue'
import ChatTransferModal from '~/components/whatsapp/ChatTransferModal.vue'
import ChatWelcomePanel from '~/components/whatsapp/ChatWelcomePanel.vue'
import UserChatBody from '~/components/whatsapp/UserChatBody.vue'
import TestReportDrawer from '~/components/whatsapp/TestReportDrawer.vue'

export const useWhatsAppState = (emit: any) => {
  const { user, isAgent, agentId, request: authRequest } = useAuth()
  const { addToCart, openCartFromChat } = useCart()
  const {
    isLogged: isMemberLogged,
    levelName,
    points,
    tierInfo,
    levelDiscounts,
    userCoupons,
    userPointCards,
    initMembership,
    refreshData: refreshMembershipData,
  } = useMembership()
  const config = useRuntimeConfig()
  const publicApiBase = computed(() => {
    const base = (config.public as { apiBase?: string }).apiBase || '/api/v1'
    return base.replace(/\/$/, '')
  })
  
  // Test Report Drawer
  const testReportDrawerVisible = ref(false)
  const handleOpenTestReport = () => {
    testReportDrawerVisible.value = true
  }
  
  // 客服模式状态
  const agentMode = computed(() => isAgent.value)
  
  // 客服会话列表
  const agentConversations = ref<any[]>([])
  const isLoadingConversations = ref(false)
  const selectedConversation = ref<any>(null)
  
  // 客服状态管理
  const currentAgentStatus = ref<string>('offline')
  const showStatusDropdown = ref(false)
  
  // 状态颜色配置
  const agentStatusColors: Record<string, { dot: string; text: string }> = {
    online: { dot: 'bg-emerald-500', text: 'text-emerald-400' },
    busy: { dot: 'bg-amber-500', text: 'text-amber-400' },
    away: { dot: 'bg-orange-500', text: 'text-orange-400' },
    offline: { dot: 'bg-gray-500', text: 'text-gray-400' }
  }
  
  // 状态标签
  const agentStatusLabels: Record<string, string> = {
    online: 'Online',
    busy: 'Busy',
    away: 'Away',
    offline: 'Offline'
  }
  
  // 欢迎页状态（客服模式下不显示欢迎页）
  const showWelcomeScreen = ref(true)
  
  // 是否有历史对话（用于显示 "Continue" 或 "Start"）
  const hasHistoryChat = ref(false)
  
  // 检查本地是否有历史对话（同步，立即返回）
  const checkLocalHistoryChat = (): boolean => {
    if (typeof window === 'undefined') return false
    try {
      const keys = Object.keys(localStorage)
      const chatKeys = keys.filter(key => key.startsWith('tz_chat_'))
      
      for (const key of chatKeys) {
        const data = localStorage.getItem(key)
        if (data) {
          const parsed = JSON.parse(data)
          if (parsed.messages && parsed.messages.length > 0) {
            return true
          }
        }
      }
    } catch (error) {
      console.error('检查本地历史对话失败:', error)
    }
    return false
  }
  
  // 从后端 API 检查是否有历史对话（异步校验）
  const checkApiHistoryChat = async (): Promise<boolean> => {
    try {
      // 获取访客ID
      let visitorId = localStorage.getItem('tz_visitor_id')
      if (!visitorId && !user.value) return false
      
      const identifier = user.value ? `user_${user.value.id}` : visitorId
      
      const response = await $fetch<{ hasConversation: boolean }>(`${publicApiBase.value}/customer-service/has-conversation`, {
        params: { visitor_id: identifier }
      })
      
      return response?.hasConversation || false
    } catch (error) {
      // API 失败时不影响用户体验，保持 localStorage 的结果
      console.error('检查后端历史对话失败:', error)
      return hasHistoryChat.value
    }
  }
  
  // 初始化历史对话检查
  const initHistoryChatCheck = async () => {
    // 1. 先从 localStorage 同步读取（立即显示）
    hasHistoryChat.value = checkLocalHistoryChat()
    
    // 2. 后台 API 校验（如果结果不同则更新）
    const apiResult = await checkApiHistoryChat()
    if (apiResult !== hasHistoryChat.value) {
      hasHistoryChat.value = apiResult
    }
  }
  
  // 客服列表和选中状态
  const agents = ref<any[]>([])
  const selectedAgent = ref<any>(null)
  const isLoadingAgents = ref(false)
  
  const welcomeAgents = computed(() => agents.value.slice(0, 3))
  
  // 在线客服数量
  const onlineAgentsCount = computed(() => agents.value.length)
  
  const isDesktopSearchFocused = ref(false)
  
  watch([showWelcomeScreen, welcomeAgents], () => {
    if (!showWelcomeScreen.value) return
    if (!welcomeAgents.value.length) return
  
    const ids = welcomeAgents.value.map(agent => String(agent?.id ?? ''))
    const currentId = selectedAgent.value?.id != null ? String(selectedAgent.value.id) : ''
    if (!currentId || !ids.includes(currentId)) {
      selectedAgent.value = welcomeAgents.value[1] || welcomeAgents.value[0]
    }
  }, { immediate: true })
  // Desktop-only搜索占位
  const desktopSearchQuery = ref('')
  
  const matchingAgents = computed<any[]>(() => {
    const query = desktopSearchQuery.value.trim().toLowerCase()
    if (!query) return []
  
    return agents.value.filter((agent) => {
      const name = (agent?.name || '').toLowerCase()
      const email = (agent?.email || '').toLowerCase()
      const rawTags = Array.isArray((agent as any).tags)
        ? (agent as any).tags.join(' ')
        : (agent as any).tags || ''
      const tags = String(rawTags).toLowerCase()
  
      return name.includes(query) || email.includes(query) || tags.includes(query)
    })
  })
  
  const shouldShowDesktopSearchResults = computed(() => {
    return isDesktopSearchFocused.value && !!desktopSearchQuery.value.trim()
  })
  
  // 全局邮箱设置
  const emailSettings = ref({
    preSalesEmail: '',
    afterSalesEmail: ''
  })
  
  type ChatTab = 'chat' | 'share' | 'orders' | 'faq' | 'warranty' | 'member' | 'test' | 'tire'
  interface ChatRoomState {
    messages: any[]
    activeTab: ChatTab
    newMessage: string
    searchQuery: string
    searchResults: any[]
    ordersList: any[]
    isLoadingOrders: boolean
    isSearching: boolean
  }
  
  const chatRooms = ref<Record<number, ChatRoomState>>({})
  const LAST_AGENT_STORAGE_KEY = 'tz_last_selected_agent'
  
  const messagesContainerMobile = ref<HTMLElement | null>(null)
  const isSending = ref(false)
  
  const ensureChatRoom = (agentId: number): ChatRoomState => {
    if (!chatRooms.value[agentId]) {
      chatRooms.value[agentId] = {
        messages: [],
        activeTab: 'chat',
        newMessage: '',
        searchQuery: '',
        searchResults: [],
        ordersList: [],
        isLoadingOrders: false,
        isSearching: false
      }
    }
    return chatRooms.value[agentId]
  }
  
  const currentChatRoom = computed<ChatRoomState | null>(() => {
    const agentId = selectedAgent.value?.id
    if (!agentId) return null
    return ensureChatRoom(agentId)
  })
  
  const messages = computed<any[]>(
    {
      get: () => { if (!currentChatRoom.value) return []; if (!currentChatRoom.value.messages) throw new Error('[CRITICAL] messages array missing in current chat room'); return currentChatRoom.value.messages; },
      set: (val) => {
        if (currentChatRoom.value) currentChatRoom.value.messages = val
      }
    }
  )
  
  const activeTab = computed({
    get: () => { if (!currentChatRoom.value) return 'chat'; if (!currentChatRoom.value.activeTab) throw new Error('[CRITICAL] activeTab missing in current chat room'); return currentChatRoom.value.activeTab; },
    set: (val: ChatTab) => {
      if (currentChatRoom.value) {
        currentChatRoom.value.activeTab = val
      }
    }
  })
  
  // 监听弹窗关闭，自动切回聊天 Tab
  watch(testReportDrawerVisible, (visible) => {
    if (!visible && activeTab.value === 'test') {
      activeTab.value = 'chat'
    }
  })
  
  // 监听 Tab 切换，如果切走则关闭弹窗
  watch(activeTab, (newTab) => {
    if (newTab !== 'test' && testReportDrawerVisible.value) {
      testReportDrawerVisible.value = false
    }
  })
  
  const newMessage = computed({
    get: () => { if (!currentChatRoom.value) return ''; if (currentChatRoom.value.newMessage === undefined) throw new Error('[CRITICAL] newMessage missing in current chat room'); return currentChatRoom.value.newMessage; },
    set: (val) => {
      if (currentChatRoom.value) currentChatRoom.value.newMessage = val
    }
  })
  
  const searchQuery = computed({
    get: () => { if (!currentChatRoom.value) return ''; if (currentChatRoom.value.searchQuery === undefined) throw new Error('[CRITICAL] searchQuery missing in current chat room'); return currentChatRoom.value.searchQuery; },
    set: (val) => {
      if (currentChatRoom.value) currentChatRoom.value.searchQuery = val
    }
  })
  
  const searchResults = computed<any[]>({
    get: () => { if (!currentChatRoom.value) return []; if (!currentChatRoom.value.searchResults) throw new Error('[CRITICAL] searchResults missing in current chat room'); return currentChatRoom.value.searchResults; },
    set: (val) => {
      if (currentChatRoom.value) currentChatRoom.value.searchResults = val
    }
  })
  
  const isSearching = computed({
    get: () => currentChatRoom.value?.isSearching || false,
    set: (val: boolean) => {
      if (currentChatRoom.value) currentChatRoom.value.isSearching = val
    }
  })
  
  const ordersList = computed<any[]>({
    get: () => { if (!currentChatRoom.value) return []; if (!currentChatRoom.value.ordersList) throw new Error('[CRITICAL] ordersList missing in current chat room'); return currentChatRoom.value.ordersList; },
    set: (val) => {
      if (currentChatRoom.value) currentChatRoom.value.ordersList = val
    }
  })
  
  const isLoadingOrders = computed({
    get: () => currentChatRoom.value?.isLoadingOrders || false,
    set: (val: boolean) => {
      if (currentChatRoom.value) currentChatRoom.value.isLoadingOrders = val
    }
  })
  
  const productDrawerVisible = ref(false)
  const productDrawerError = ref<string | null>(null)
  const productDrawerQuery = ref('')
  const historyDrawerVisible = ref(false)
  const wishlistDrawerVisible = ref(false)
  
  // 转接功能
  const showTransferModal = ref(false)
  const transferToAgent = ref('')
  const transferNote = ref('')
  const isTransferring = ref(false)
  
  // 图片上传
  const imageInput = ref<HTMLInputElement | null>(null)
  const isUploadingImage = ref(false)
  
  // 生成会话ID（基于访客标识）
  const conversationId = computed(() => {
    if (user.value) {
      return `user_${user.value.id}`
    }
    // 访客使用 localStorage 中的唯一ID
    let visitorId = localStorage.getItem('tz_visitor_id')
    if (!visitorId) {
      visitorId = `visitor_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      localStorage.setItem('tz_visitor_id', visitorId)
    }
    return visitorId
  })
  
  // LocalStorage 键名（包含客服ID，确保每个客服的聊天记录独立）
  const STORAGE_KEY = computed(() => {
    const agentId = selectedAgent.value?.id || 'default'
    return `tz_chat_${conversationId.value}_agent_${agentId}`
  })
  const STORAGE_EXPIRY_DAYS = 5
  
  // Toast 提示
  const showToast = ref(false)
  const toastMessage = ref('')
  let toastTimer: number | null = null
  
  const messagePressTimer = ref<number | null>(null)
  const pressedMessage = ref<any | null>(null)
  
  // WhatsApp 长按相关
  let longPressTimer: number | null = null
  const longPressDuration = 500 // 长按时长（毫秒）
  let isLongPress = ref(false)
  
  // 是否显示"我的订单"标签
  const shouldShowOrders = computed(() => !!user.value)
  
  // 保修查询登录状态
  const isLoggedInForWarranty = computed(() => !!user.value)
  
  // 聊天内登录弹窗状态
  const showAuthModal = ref(false)
  const authMode = ref<'login' | 'register'>('login')
  
  // 打开聊天内 AuthModal（用于会员 / 保修登录）
  const openMemberAuth = (mode: 'login' | 'register') => {
    authMode.value = mode
    showAuthModal.value = true
  }
  
  // 从聊天中的保修查询触发登录：打开 AuthModal
  const handleWarrantyLoginRequest = () => {
    openMemberAuth('login')
  }
  
  const handleChatAuthSuccess = async () => {
    showAuthModal.value = false
    await refreshMembershipData()
  }
  
  // 关闭弹窗
  const handleClose = () => {
    emit('close')
  }
  
  // 进入聊天（从欢迎页）
  const enterChat = () => {
    if (selectedAgent.value) {
      showWelcomeScreen.value = false
    }
  }
  
  // 在欢迎页选择客服
  const selectAgentFromWelcome = (agent: any) => {
    selectedAgent.value = agent
    ensureChatRoom(agent.id)
    loadMessagesFromStorage()
  }
  
  // FAQ 数据
  const faqItems = [
    { id: 'wheelset', text: 'How to choose the right wheelset?', url: '/guides/wheelset-buyers' },
    { id: 'warranty', text: "What's the warranty policy?", url: '/support/warranty-check' },
    { id: 'shipping', text: 'Shipping & delivery times', url: '/support/shipping' },
  ]
  
  // 显示 Toast 提示
  const displayToast = (message: string, duration = 2000) => {
    toastMessage.value = message
    showToast.value = true
    
    if (toastTimer) clearTimeout(toastTimer)
    toastTimer = setTimeout(() => {
      showToast.value = false
    }, duration)
  }
  
  // WhatsApp 触摸开始（长按检测）
  const handleWhatsAppTouchStart = (agent: any) => {
    if (!agent.whatsapp) return
    
    isLongPress.value = false
    longPressTimer = setTimeout(() => {
      isLongPress.value = true
      // 长按触发，打开 WhatsApp
      if (confirm(`Open WhatsApp to contact ${agent.name}?`)) {
        window.open(`https://wa.me/${agent.whatsapp.replace('+', '')}`, '_blank')
      }
    }, longPressDuration)
  }
  
  // WhatsApp 触摸结束
  const handleWhatsAppTouchEnd = () => {
    if (longPressTimer) {
      clearTimeout(longPressTimer)
      longPressTimer = null
    }
  }
  
  // WhatsApp 点击（桌面端或短按）
  const handleWhatsAppClick = (agent: any) => {
    if (!agent.whatsapp) return
    
    // 如果是长按触发的，不执行点击逻辑
    if (isLongPress.value) {
      isLongPress.value = false
      return
    }
    
    // 短按显示提示
    displayToast('Long press to open WhatsApp', 2000)
  }
  
  // WhatsApp 链接
  
  const whatsappLink = computed(() => {
    if (!selectedAgent.value?.whatsapp) return ''
    return `https://wa.me/${selectedAgent.value.whatsapp.replace('+', '')}`
  })
  
  const canDeleteMessage = (message: any) => !message.is_agent
  
  const confirmDeleteMessage = (message: any) => {
    if (!canDeleteMessage(message)) return
    const ok = confirm('Delete this message from your local history?')
    if (ok) {
      deleteMessage(message)
    }
  }
  
  const deleteMessage = (message: any) => {
    if (!currentChatRoom.value) return
    currentChatRoom.value.messages = currentChatRoom.value.messages.filter((msg) => msg.id !== message.id)
    saveMessagesToStorage()
    displayToast('Message deleted', 1800)
  }
  
  const clearMessagePressTimer = () => {
    if (messagePressTimer.value) {
      clearTimeout(messagePressTimer.value)
      messagePressTimer.value = null
    }
    pressedMessage.value = null
  }
  
  const startMessagePress = (message: any) => {
    if (!canDeleteMessage(message)) return
    pressedMessage.value = message
    clearMessagePressTimer()
    messagePressTimer.value = window.setTimeout(() => {
      messagePressTimer.value = null
      if (pressedMessage.value) {
        confirmDeleteMessage(pressedMessage.value)
        pressedMessage.value = null
      }
    }, 600)
  }
  
  const handleMessageTouchStart = (message: any) => {
    startMessagePress(message)
  }
  
  const handleMessageTouchEnd = () => {
    clearMessagePressTimer()
  }
  
  const handleMessageMouseDown = (message: any) => {
    // Only handle long press for non-touch devices when mouse button held
    if ((window as any)?.ontouchstart !== undefined) return
    startMessagePress(message)
  }
  
  const handleMessageMouseUp = () => {
    clearMessagePressTimer()
  }
  
  const handleMessageContextMenu = (message: any) => {
    confirmDeleteMessage(message)
  }
  
  // 获取状态文本
  const getStatusText = (status: string) => {
    const statusMap: Record<string, string> = {
      active: '在线',
      closed: '已关闭',
      pending: '待处理'
    }
    return statusMap[status] || status
  }
  
  // 格式化消息时间
  const formatMessageTime = (time: string) => {
    const date = new Date(time)
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  
  // 滚动到底部
  const scrollToBottom = () => {
    nextTick(() => {
      const containers = [messagesContainerMobile.value]
      containers.forEach((container) => {
        if (container) {
          container.scrollTop = container.scrollHeight
        }
      })
    })
  }
  
  // 监听消息变化，自动滚动到底部
  watch(messages, () => {
    scrollToBottom()
  }, { deep: true })
  
  // 监听客服切换，加载对应的聊天记录
  watch(() => selectedAgent.value?.id, (newId, oldId) => {
    if (newId && newId !== oldId) {
      localStorage.setItem(LAST_AGENT_STORAGE_KEY, String(newId))
      loadMessagesFromStorage()
      scrollToBottom()
    }
  })
  
  // 监听标签切换，按需加载订单
  watch(activeTab, (tab) => {
    if (tab === 'orders' && !ordersList.value.length && !isLoadingOrders.value) {
      loadOrders()
    }
  })
  
  // 从 localStorage 加载消息
  const loadMessagesFromStorage = () => {
    if (!selectedAgent.value) return
    const currentRoom = ensureChatRoom(selectedAgent.value.id)
  
    try {
      const stored = localStorage.getItem(STORAGE_KEY.value)
      if (stored) {
        const data = JSON.parse(stored)
        const now = Date.now()
        const expiryTime = STORAGE_EXPIRY_DAYS * 24 * 60 * 60 * 1000
        
        if (!data.messages) throw new Error('[CRITICAL] Messages array missing from local storage data');
        const validMessages = data.messages.filter((msg: any) => {
          const msgTime = new Date(msg.created_at).getTime()
          return (now - msgTime) < expiryTime
        })
  
        currentRoom.messages = validMessages
        currentRoom.activeTab = (data.activeTab as ChatTab) || 'chat'
        currentRoom.newMessage = data.newMessage || ''
        currentRoom.searchQuery = data.searchQuery || ''
        if (!data.searchResults) throw new Error('[CRITICAL] searchResults missing in storage data');
        currentRoom.searchResults = Array.isArray(data.searchResults) ? data.searchResults : []
        if (!data.ordersList) throw new Error('[CRITICAL] ordersList missing in storage data');
        currentRoom.ordersList = Array.isArray(data.ordersList) ? data.ordersList : []
        currentRoom.isSearching = !!data.isSearching
        currentRoom.isLoadingOrders = !!data.isLoadingOrders
  
        if (validMessages.length !== data.messages.length) {
          saveMessagesToStorage()
        }
      } else {
        currentRoom.messages = []
      }
    } catch (error) {
      console.error('加载消息失败:', error)
    }
  }
  
  // 保存消息到 localStorage
  const saveMessagesToStorage = () => {
    if (!selectedAgent.value) return
    const currentRoom = ensureChatRoom(selectedAgent.value.id)
    try {
      localStorage.setItem(STORAGE_KEY.value, JSON.stringify({
        messages: currentRoom.messages,
        activeTab: currentRoom.activeTab,
        newMessage: currentRoom.newMessage,
        searchQuery: currentRoom.searchQuery,
        searchResults: currentRoom.searchResults,
        ordersList: currentRoom.ordersList,
        isSearching: currentRoom.isSearching,
        isLoadingOrders: currentRoom.isLoadingOrders,
        lastUpdated: new Date().toISOString()
      }))
    } catch (error) {
      console.error('保存消息失败:', error)
    }
  }
  
  // 发送消息到后端 API
  const sendMessageToAPI = async (messageData: any) => {
    try {
      const response = await authRequest<any>(
        '/customer-service/messages',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            conversation_id: conversationId.value,
            message: messageData.message,
            sender_type: user.value ? 'user' : 'visitor',
            sender_name: user.value?.display_name || '访客',
            sender_email: user.value?.email || '',
            agent_id: selectedAgent.value?.id || '',
            message_type: messageData.message_type || 'text',
            metadata: messageData.metadata || null
          })
        },
        'Failed to send customer-service message'
      )
      return response
    } catch (error) {
      console.error('发送消息到API失败:', error)
      throw error
    }
  }
  
  // 发送消息
  const handleSendMessage = async () => {
    if (!newMessage.value.trim() || !selectedAgent.value || isSending.value) {
      return
    }
  
    isSending.value = true
    const messageText = newMessage.value
    newMessage.value = ''
  
    const messageData = {
      id: Date.now(),
      conversation_id: conversationId.value,
      sender_id: user.value?.id || 0,
      sender_name: user.value?.display_name || '访客',
      sender_email: user.value?.email || '',
      message: messageText,
      message_type: 'text',
      created_at: new Date().toISOString(),
      is_agent: false
    }
  
    try {
      // 1. 先添加到本地显示
      messages.value.push(messageData)
      scrollToBottom()
      
      // 2. 保存到 localStorage
      saveMessagesToStorage()
      
      // 3. 发送到后端 API（实时存储）
      await sendMessageToAPI(messageData)
      
      // 4. 检查关键词自动回复
      await checkAutoReply(messageText)
    } catch (error) {
      // 如果 API 失败，消息仍然保存在 localStorage 中
      console.error('发送失败', error)
      // 可以添加重试逻辑或提示用户
    } finally {
      isSending.value = false
    }
  }
  
  // 检查关键词自动回复
  const checkAutoReply = async (userMessage: string) => {
    try {
      const response = await $fetch<any>(`${publicApiBase.value}/customer-service/auto-reply/match`, {
        method: 'POST',
        body: {
          message: userMessage,
          conversation_id: conversationId.value
        }
      })
      
      if (response.success && response.data.reply) {
        // 延迟 500ms 模拟真实回复
        setTimeout(() => {
          messages.value.push({
            id: Date.now(),
            conversation_id: conversationId.value,
            sender_id: 0,
            sender_name: 'Auto Reply',
            sender_email: '',
            message: response.data.reply,
            message_type: 'text',
            created_at: new Date().toISOString(),
            is_agent: true
          })
          
          saveMessagesToStorage()
          scrollToBottom()
        }, 500)
      }
    } catch (error) {
      console.error('自动回复检查失败', error)
    }
  }
  
  // 搜索商品
  const searchProducts = async () => {
    console.log('[WhatsAppChatModal] searchProducts clicked, query =', searchQuery.value)
  
    const trimmedQuery = searchQuery.value.trim()
  
    // 如果关键字为空：仍然打开抽屉，只显示空状态，方便确认组件是否挂载
    if (!trimmedQuery) {
      console.log('[WhatsAppChatModal] empty search query, open drawer with empty state')
      productDrawerQuery.value = ''
      productDrawerError.value = null
      productDrawerVisible.value = true
      searchResults.value = []
      isSearching.value = false
      return
    }
  
    productDrawerQuery.value = trimmedQuery
    productDrawerError.value = null
    productDrawerVisible.value = true
  
    isSearching.value = true
    try {
      console.log('[WhatsAppChatModal] fetching products...')
      const response = await $fetch<any>(`${publicApiBase.value}/customer-service/products`, {
        params: {
          keyword: trimmedQuery,
          per_page: 20
        }
      })
      
      // 转换数据格式以适配前端显示
      if (!response || !Array.isArray(response.items)) { throw new Error('[CRITICAL] Invalid response format for products API'); }
      
        searchResults.value = response.items.map((item: any) => ({
          id: item.id,
          title: item.title,
          name: item.title,
          slug: item.slug,
          sku: item.sku,
          url: item.preview_url || `/shop/${item.slug || item.id}`,
          thumbnail: item.thumbnail,
          priceValue: item.prices?.sale > 0 ? item.prices.sale : (item.prices?.regular || 0),
          price: item.prices?.sale > 0 
            ? `$${item.prices.sale}` 
            : (item.prices?.regular > 0 ? `$${item.prices.regular}` : ''),
          maxStock: item.stock?.quantity || 0
        }))
        console.log('[WhatsAppChatModal] products loaded:', searchResults.value.length)
      
    } catch (error) {
      console.error('搜索失败:', error)
      productDrawerError.value = 'Search failed, please try again.'
      searchResults.value = []
    } finally {
      isSearching.value = false
      console.log('[WhatsAppChatModal] search finished')
    }
  }
  
  const handleAddProductToCart = (product: any) => {
    const result = addToCart({
      id: product.id,
      title: product.title || product.name || 'Product',
      name: product.name || product.title || 'Product',
      slug: product.slug,
      sku: product.sku,
      thumbnail: product.thumbnail,
      price: Number(product.priceValue || 0),
      maxStock: Number(product.maxStock || 0)
    })
  
    if (result.success) {
      openCartFromChat()
    } else {
      productDrawerError.value = result.message || 'Unable to add this product to cart.'
    }
  }
  
  const handleProductDrawerClose = () => {
    productDrawerVisible.value = false
    productDrawerError.value = null
    productDrawerQuery.value = ''
    searchQuery.value = ''
    searchResults.value = []
    isSearching.value = false
  }
  
  const handleHistoryDrawerClose = () => {
    historyDrawerVisible.value = false
  }
  
  // 分享商品到聊天
  const shareProductToChat = async (product: any) => {
    if (!selectedAgent.value || isSending.value) return
    
    isSending.value = true
    
    const messageData = {
      id: Date.now(),
      conversation_id: conversationId.value,
      sender_id: user.value?.id || 0,
      sender_name: user.value?.display_name || '访客',
      sender_email: user.value?.email || '',
      message: product.title || '商品',
      message_type: 'product',
      metadata: {
        title: product.title,
        url: product.url,
        thumbnail: product.thumbnail,
        price: product.price
      },
      created_at: new Date().toISOString(),
      is_agent: false
    }
    
    try {
      messages.value.push(messageData)
      saveMessagesToStorage()
      await sendMessageToAPI(messageData)
      activeTab.value = 'chat'
      scrollToBottom()
    } catch (error) {
      console.error('分享商品失败:', error)
    } finally {
      isSending.value = false
    }
  }
  
  // 从浏览历史分享商品到聊天
  const handleShareProductFromHistory = async (product: any) => {
    if (!selectedAgent.value || isSending.value) return
    
    isSending.value = true
    
    const messageData = {
      id: Date.now(),
      conversation_id: conversationId.value,
      sender_id: user.value?.id || 0,
      sender_name: user.value?.display_name || '访客',
      sender_email: user.value?.email || '',
      message: product.title || '商品',
      message_type: 'product',
      metadata: {
        title: product.title,
        url: product.url,
        thumbnail: product.thumbnail,
        price: product.price
      },
      created_at: new Date().toISOString(),
      is_agent: false
    }
    
    try {
      messages.value.push(messageData)
      saveMessagesToStorage()
      await sendMessageToAPI(messageData)
      activeTab.value = 'chat'
      scrollToBottom()
    } catch (error) {
      console.error('从浏览历史分享商品失败:', error)
    } finally {
      isSending.value = false
    }
  }
  
  // 加载订单列表
  const loadOrders = async () => {
    isLoadingOrders.value = true
    try {
      const response = await authRequest<any[]>('/customer-service/orders?limit=10', {
        headers: { accept: 'application/json' }
      }, 'Failed to load customer-service orders')
      if (!Array.isArray(response)) throw new Error('[CRITICAL] Invalid response format for orders API');
      ordersList.value = response
    } catch (error) {
      console.error('加载订单失败:', error)
      ordersList.value = []
    } finally {
      isLoadingOrders.value = false
    }
  }
  
  // 分享订单到聊天
  const shareOrderToChat = async (order: any) => {
    if (!selectedAgent.value || isSending.value) return
    
    isSending.value = true
    
    const messageData = {
      id: Date.now(),
      conversation_id: conversationId.value,
      sender_id: user.value?.id || 0,
      sender_name: user.value?.display_name || '访客',
      sender_email: user.value?.email || '',
      message: `订单 #${order.id}`,
      message_type: 'order',
      metadata: {
        order_id: order.id,
        title: `订单 #${order.id}`,
        total: order.total,
        currency: order.currency,
        url: order.url,
        thumbnail: order.thumbnail
      },
      created_at: new Date().toISOString(),
      is_agent: false
    }
    
    try {
      messages.value.push(messageData)
      saveMessagesToStorage()
      await sendMessageToAPI(messageData)
      activeTab.value = 'chat'
      scrollToBottom()
    } catch (error) {
      console.error('分享订单失败:', error)
    } finally {
      isSending.value = false
    }
  }
  
  // 获取客服列表（带缓存）
  const fetchAgents = async () => {
    isLoadingAgents.value = true
    try {
      // 1. 先尝试从 localStorage 读取缓存
      if (typeof window !== 'undefined') {
        const cached = localStorage.getItem('whatsapp_agents_cache')
        if (cached) {
          try {
            const { data, timestamp } = JSON.parse(cached)
            // 缓存有效期：30分钟
            if (Date.now() - timestamp < 30 * 60 * 1000) {
              // 过滤掉当前登录用户关联的客服
              const currentUserId = user.value?.id
              const filteredAgents = data.agents.filter((agent: any) => {
                return !agent.wp_user_id || agent.wp_user_id !== currentUserId
              })
              
              agents.value = filteredAgents
              if (data.emailSettings) {
                emailSettings.value = data.emailSettings
              }
  
              await initializeSelectedAgent()
              isLoadingAgents.value = false
              return
            }
          } catch (e) {
            // 缓存解析失败，继续请求
          }
        }
      }
      
      // 2. 缓存不存在或过期，从 API 获取
      let agentsData: any[] = []
      
      try {
        const response = await $fetch<any>(`${publicApiBase.value}/customer-service/agents`)
        if (response.success && response.data) {
          agentsData = response.data
        }
      } catch (error) {
        console.warn('Failed to fetch agents from API, using mock data for dev')
      }
      
      // 用于缓存的原始数据
      let cacheData = { agents: agentsData, emailSettings: null as any }
      
      // 开发环境：如果 API 没有返回数据，使用模拟数据
      if (agentsData.length === 0 && import.meta.dev) {
        agentsData = [
          { id: 'CS001', name: 'Sales', email: 'sales@tanzanite.site', avatar: '', whatsapp: '+8613800138001', wp_user_id: null },
          { id: 'CS002', name: 'Tech Support', email: 'tech@tanzanite.site', avatar: '', whatsapp: '+8613800138002', wp_user_id: null },
          { id: 'CS003', name: 'After Sales', email: 'support@tanzanite.site', avatar: '', whatsapp: '+8613800138003', wp_user_id: null },
        ]
        cacheData.agents = agentsData
        // 开发环境设置默认邮箱
        emailSettings.value = {
          preSalesEmail: 'sales@tanzanite.site',
          afterSalesEmail: 'support@tanzanite.site'
        }
        cacheData.emailSettings = emailSettings.value
      }
      
      if (agentsData.length > 0) {
        // 过滤掉当前登录用户关联的客服
        const currentUserId = user.value?.id
        const filteredAgents = agentsData.filter((agent: any) => {
          // 如果客服没有关联 wp_user_id，或者不是当前用户，则显示
          return !agent.wp_user_id || agent.wp_user_id !== currentUserId
        })
        
        agents.value = filteredAgents
        
        // 3. 保存到 localStorage（保存原始数据，过滤在读取时进行）
        if (typeof window !== 'undefined' && cacheData.agents.length > 0) {
          localStorage.setItem('whatsapp_agents_cache', JSON.stringify({
            data: cacheData,
            timestamp: Date.now()
          }))
        }
        
        await initializeSelectedAgent()
      }
    } catch (error) {
      console.error('获取客服列表失败:', error)
    } finally {
      isLoadingAgents.value = false
    }
  }
  
  const initializeSelectedAgent = async () => {
    if (!agents.value.length) {
      selectedAgent.value = null
      return
    }
  
    let defaultAgent = agents.value[0]
    if (typeof window !== 'undefined') {
      const storedId = localStorage.getItem(LAST_AGENT_STORAGE_KEY)
      if (storedId) {
        const matched = agents.value.find(agent => String(agent.id) === storedId)
        if (matched) {
          defaultAgent = matched
        }
      }
    }
  
    if (!selectedAgent.value || selectedAgent.value.id !== defaultAgent.id) {
      selectedAgent.value = defaultAgent
      ensureChatRoom(defaultAgent.id)
      loadMessagesFromStorage()
      await sendWelcomeMessage()
    }
  }
  
  // 发送欢迎语
  const sendWelcomeMessage = async () => {
    try {
      const response = await $fetch<any>(`${publicApiBase.value}/customer-service/auto-reply/welcome`, {
        params: {
          conversation_id: conversationId.value
        }
      })
      
      if (response.success && response.data.message && !response.data.already_sent) {
        // 添加欢迎消息到消息列表
        messages.value.push({
          id: Date.now(),
          conversation_id: conversationId.value,
          sender_id: 0,
          sender_name: 'System',
          sender_email: '',
          message: response.data.message,
          message_type: 'text',
          created_at: new Date().toISOString(),
          is_agent: true
        })
        
        saveMessagesToStorage()
        scrollToBottom()
      }
    } catch (error) {
      console.error('发送欢迎语失败:', error)
    }
  }
  
  // 选择客服
  const selectAgent = (agent: any) => {
    if (selectedAgent.value?.id === agent.id) return
    selectedAgent.value = agent
    ensureChatRoom(agent.id)
    loadMessagesFromStorage()
  }
  
  const selectAgentFromSearch = (agent: any) => {
    selectAgent(agent)
    isDesktopSearchFocused.value = false
  }
  
  const handleDesktopSearchBlur = () => {
    setTimeout(() => {
      isDesktopSearchFocused.value = false
    }, 100)
  }
  
  // 根据客服ID获取背景颜色值（深色系）
  const getAgentBgColorValue = (agentId: number) => {
    const colors = [
      '#0a0a0a',      // 深黑（默认）
      '#0d1117',      // 深蓝黑
      '#0f0a14',      // 深紫黑
      '#0a1410',      // 深绿黑
      '#14100a',      // 深橙黑
      '#100a14',      // 深紫红黑
    ]
    return colors[agentId % colors.length] || colors[0]
  }
  
  // 根据客服ID获取背景颜色类名（深色系）- 保留用于其他地方
  const getAgentBgColor = (agentId: number) => {
    const colors = [
      'bg-[#0a0a0a]',      // 深黑（默认）
      'bg-[#0d1117]',      // 深蓝黑
      'bg-[#0f0a14]',      // 深紫黑
      'bg-[#0a1410]',      // 深绿黑
      'bg-[#14100a]',      // 深橙黑
      'bg-[#100a14]',      // 深紫红黑
    ]
    return colors[agentId % colors.length] || colors[0]
  }
  
  const agentThemePalette = ['#6b73ff', '#40ffaa', '#C77DFF']
  const getAgentThemeColor = (agentId: number) => {
    return agentThemePalette[(agentId - 1) % agentThemePalette.length] || agentThemePalette[0]
  }
  
  const currentThemeColor = computed(() => {
    if (!selectedAgent.value?.id) return agentThemePalette[0]
    return getAgentThemeColor(selectedAgent.value.id)
  })
  
  const mobilePanelStyle = computed(() => {
    const color = currentThemeColor.value
    return {
      borderColor: color,
      background: `linear-gradient(180deg, ${color}33 0%, rgba(0,0,0,0.85) 100%)`,
      boxShadow: `0 15px 40px ${color}40`
    }
  })
  
  // 获取首字母
  const getInitials = (name: string) => {
    if (!name) return '?'
    const parts = name.split(' ')
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase()
    }
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
  }
  
  // 转接会话
  async function handleTransfer() {
    if (!transferToAgent.value) {
      alert('请选择要转接的客服')
      return
    }
    
    if (transferToAgent.value === selectedAgent.value?.id) {
      alert('不能转接给当前客服')
      return
    }
    
    isTransferring.value = true
    
    try {
      const data = await authRequest<any>(
        `/customer-service/agent/conversations/${selectedConversation.value.id}/transfer`,
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            to_agent_id: String(transferToAgent.value),
            note: transferNote.value,
          }),
        },
        'Transfer failed'
      )
      
      if (data.success) {
        alert(`转接成功！会话已转接给 ${data.data.to_agent}`)
        showTransferModal.value = false
        transferToAgent.value = ''
        transferNote.value = ''
        
        // 刷新消息列表以显示系统消息
        loadMessagesFromStorage()
      } else {
        alert(data.message || '转接失败')
      }
    } catch (error) {
      console.error('转接失败:', error)
      alert('转接失败，请稍后重试')
    } finally {
      isTransferring.value = false
    }
  }
  
  // ...
  // 图片上传处理
  const handleImageUpload = async (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    
    if (!file) return
    
    // 检查文件大小（限制5MB）
    if (file.size > 5 * 1024 * 1024) {
      alert('图片大小不能超过 5MB')
      return
    }
    
    isUploadingImage.value = true
    
    try {
      // TODO: 实现图片上传到服务器
      // 这里暂时使用 FileReader 转为 base64
      const reader = new FileReader()
      reader.onload = async (e) => {
        const imageUrl = e.target?.result as string
        
        // 创建图片消息
        const messageData = {
          id: Date.now(),
          conversation_id: conversationId.value,
          sender_id: user.value?.id || 0,
          sender_name: user.value?.display_name || '访客',
          sender_email: user.value?.email || '',
          message: '[图片]',
          message_type: 'image',
          attachment_url: imageUrl,
          created_at: new Date().toISOString(),
          is_agent: false
        }
        
        // 添加到消息列表
        messages.value.push(messageData)
        saveMessagesToStorage()
        scrollToBottom()
        
        // 发送到后端
        try {
          await sendMessageToAPI(messageData)
        } catch (error) {
          console.error('发送图片失败', error)
        }
      }
      
      reader.readAsDataURL(file)
    } catch (error) {
      console.error('上传图片失败:', error)
      alert('上传失败，请重试')
    } finally {
      isUploadingImage.value = false
      // 清空文件选择
      if (target) {
        target.value = ''
      }
    }
  }
  
  // 获取客服会话列表（客服模式）
  const fetchAgentConversations = async () => {
    if (!agentMode.value) return
    
    isLoadingConversations.value = true
    try {
      const response = await authRequest<any>(
        '/customer-service/agent/conversations',
        { headers: { accept: 'application/json' } },
        'Failed to load agent conversations'
      )
      
      if (response?.ok && response?.data) {
        if (!response.data.items) throw new Error('[CRITICAL] Items array missing in agent conversations response');
        agentConversations.value = response.data.items
      }
    } catch (error) {
      console.error('获取客服会话列表失败:', error)
    } finally {
      isLoadingConversations.value = false
    }
  }
  
  // 选择会话（客服模式）
  const selectConversation = (conversation: any) => {
    selectedConversation.value = conversation
    // 加载该会话的消息
    loadConversationMessages(conversation.id)
  }
  
  // 加载会话消息
  const loadConversationMessages = async (conversationId: string) => {
    try {
      const response = await authRequest<any>(
        `/customer-service/agent/conversations/${conversationId}/messages`,
        { headers: { accept: 'application/json' } },
        'Failed to load conversation messages'
      )
      
      if (response?.ok && response?.data) {
        if (!response.data.items) throw new Error('[CRITICAL] Items array missing in conversation messages response');
        messages.value = response.data.items
        scrollToBottom()
      }
    } catch (error) {
      console.error('加载会话消息失败:', error)
    }
  }
  
  // 返回会话列表（客服模式）
  const backToConversationList = () => {
    selectedConversation.value = null
  }
  
  // 发送消息（客服模式）
  const sendMessage = async () => {
    if (!newMessage.value.trim() || !selectedConversation.value) return
    
    isSending.value = true
    const messageText = newMessage.value.trim()
    newMessage.value = ''
    
    try {
      const response = await authRequest<any>(
        '/customer-service/agent/messages',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            conversation_id: selectedConversation.value.id,
            message: messageText
          })
        },
        'Failed to send agent message'
      )
      
      if (response?.ok) {
        // 添加消息到列表
        messages.value.push({
          id: Date.now(),
          message: messageText,
          sender_type: 'agent',
          created_at: new Date().toISOString()
        })
        scrollToBottom()
      }
    } catch (error) {
      console.error('发送消息失败:', error)
      // 恢复消息
      newMessage.value = messageText
    } finally {
      isSending.value = false
    }
  }
  
  // 获取客服状态
  const fetchAgentStatus = async () => {
    if (!agentMode.value) return
    
    try {
      const response = await authRequest<any>(
        '/customer-service/agent/status',
        { headers: { accept: 'application/json' } },
        'Failed to load agent status'
      )
      
      if (response?.ok && response?.data?.status) {
        currentAgentStatus.value = response.data.status
      }
    } catch (error) {
      console.error('获取客服状态失败:', error)
    }
  }
  
  // 更新客服状态
  const changeAgentStatus = async (status: string) => {
    showStatusDropdown.value = false
    
    const previousStatus = currentAgentStatus.value
    currentAgentStatus.value = status // 乐观更新
    
    try {
      const response = await authRequest<any>(
        '/customer-service/agent/status',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ status })
        },
        'Failed to update agent status'
      )
      
      if (!response?.ok) {
        // 回滚
        currentAgentStatus.value = previousStatus
      }
    } catch (error) {
      console.error('更新客服状态失败:', error)
      // 回滚
      currentAgentStatus.value = previousStatus
    }
  }
  
  // 组件挂载时获取客服列表、会员数据和检查历史对话
  onMounted(async () => {
    await initMembership()
  
    if (agentMode.value) {
      // 客服模式：获取会话列表和状态，跳过欢迎页
      showWelcomeScreen.value = false
      await Promise.all([
        fetchAgentConversations(),
        fetchAgentStatus()
      ])
    } else {
      // 访客模式：获取客服列表
      await fetchAgents()
      initHistoryChatCheck()
    }
    scrollToBottom()
  })
  
  return {
    config,
    publicApiBase,
    base,
    testReportDrawerVisible,
    handleOpenTestReport,
    agentMode,
    agentConversations,
    isLoadingConversations,
    selectedConversation,
    currentAgentStatus,
    showStatusDropdown,
    showWelcomeScreen,
    hasHistoryChat,
    checkLocalHistoryChat,
    keys,
    chatKeys,
    data,
    parsed,
    checkApiHistoryChat,
    visitorId,
    identifier,
    response,
    initHistoryChatCheck,
    apiResult,
    agents,
    selectedAgent,
    isLoadingAgents,
    welcomeAgents,
    onlineAgentsCount,
    isDesktopSearchFocused,
    ids,
    currentId,
    desktopSearchQuery,
    matchingAgents,
    query,
    name,
    email,
    rawTags,
    tags,
    shouldShowDesktopSearchResults,
    emailSettings,
    chatRooms,
    LAST_AGENT_STORAGE_KEY,
    messagesContainerMobile,
    isSending,
    ensureChatRoom,
    currentChatRoom,
    agentId,
    messages,
    activeTab,
    newMessage,
    searchQuery,
    searchResults,
    isSearching,
    ordersList,
    isLoadingOrders,
    productDrawerVisible,
    productDrawerError,
    productDrawerQuery,
    historyDrawerVisible,
    wishlistDrawerVisible,
    showTransferModal,
    transferToAgent,
    transferNote,
    isTransferring,
    imageInput,
    isUploadingImage,
    conversationId,
    STORAGE_KEY,
    STORAGE_EXPIRY_DAYS,
    showToast,
    toastMessage,
    messagePressTimer,
    pressedMessage,
    longPressDuration,
    isLongPress,
    shouldShowOrders,
    isLoggedInForWarranty,
    showAuthModal,
    authMode,
    openMemberAuth,
    handleWarrantyLoginRequest,
    handleChatAuthSuccess,
    handleClose,
    enterChat,
    selectAgentFromWelcome,
    faqItems,
    displayToast,
    handleWhatsAppTouchStart,
    handleWhatsAppTouchEnd,
    handleWhatsAppClick,
    whatsappLink,
    canDeleteMessage,
    confirmDeleteMessage,
    ok,
    deleteMessage,
    clearMessagePressTimer,
    startMessagePress,
    handleMessageTouchStart,
    handleMessageTouchEnd,
    handleMessageMouseDown,
    handleMessageMouseUp,
    handleMessageContextMenu,
    getStatusText,
    formatMessageTime,
    date,
    scrollToBottom,
    containers,
    loadMessagesFromStorage,
    currentRoom,
    stored,
    now,
    expiryTime,
    validMessages,
    msgTime,
    saveMessagesToStorage,
    sendMessageToAPI,
    handleSendMessage,
    messageText,
    messageData,
    checkAutoReply,
    searchProducts,
    trimmedQuery,
    handleAddProductToCart,
    result,
    handleProductDrawerClose,
    handleHistoryDrawerClose,
    shareProductToChat,
    handleShareProductFromHistory,
    loadOrders,
    shareOrderToChat,
    fetchAgents,
    cached,
    currentUserId,
    filteredAgents,
    cacheData,
    initializeSelectedAgent,
    defaultAgent,
    storedId,
    matched,
    sendWelcomeMessage,
    selectAgent,
    selectAgentFromSearch,
    handleDesktopSearchBlur,
    getAgentBgColorValue,
    colors,
    getAgentBgColor,
    agentThemePalette,
    getAgentThemeColor,
    currentThemeColor,
    mobilePanelStyle,
    color,
    getInitials,
    parts,
    handleImageUpload,
    target,
    file,
    reader,
    imageUrl,
    fetchAgentConversations,
    selectConversation,
    loadConversationMessages,
    backToConversationList,
    sendMessage,
    fetchAgentStatus,
    changeAgentStatus,
    previousStatus
  }
}
