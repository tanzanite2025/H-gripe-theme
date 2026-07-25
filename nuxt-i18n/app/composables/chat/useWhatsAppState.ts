import { ref, watch, nextTick, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from '#imports'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { useMembership } from '~/composables/useMembership'
import { normalizeShopProduct } from '~/composables/useShopProducts'
import { loadChatAgentDirectory } from '~/composables/chat/useChatAgentDirectory'
import { useCustomerServiceChatSync } from '~/composables/chat/useCustomerServiceChatSync'
import { buildProductConfigConfirmMetadata } from '~/composables/chat/useProductConfigConfirmPayload'
import { buildOrderChatMetadata } from '~/composables/chat/useOrderChatPayload'
import {
  CHAT_STORAGE_EXPIRY_DAYS,
  createEmptyChatRoom,
  getChatStorageKey,
  getLastSelectedAgentId,
  hasLocalChatHistory,
  loadChatRoomFromStorage,
  saveChatRoomToStorage,
  saveLastSelectedAgentId
} from '~/composables/chat/useChatStorage'
import type { ChatRoomState, ChatTab } from '~/composables/chat/useChatStorage'

export const useWhatsAppState = (emit: any) => {
  const { t } = useI18n()
  const { user, request: authRequest } = useAuth()
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
  
  // 欢迎页状态。Nuxt 前台只承接客户侧聊天，不承接客服工作台。
  const showWelcomeScreen = ref(true)
  
  // 是否有历史对话（用于显示 "Continue" 或 "Start"）
  const hasHistoryChat = ref(false)
  
  // 检查本地是否有历史对话（同步，立即返回）
  const checkLocalHistoryChat = (): boolean => {
    if (typeof window === 'undefined') return false
    try {
      return hasLocalChatHistory()
    } catch (error) {
      console.error('检查本地历史对话失败:', error)
    }
    return false
  }
  
  // 从后端 API 检查是否有历史对话（异步校验）
  const checkApiHistoryChat = async (): Promise<boolean> => {
    try {
      // 获取访客ID
      const response = await $fetch<{ hasConversation: boolean; conversation_id?: string }>(`${publicApiBase.value}/customer-service/has-conversation`, {
        credentials: 'include'
      })
      if (response?.conversation_id) {
        conversationId.value = response.conversation_id
      }
      
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
    if (apiResult) {
      hasHistoryChat.value = true
    }
  }
  
  // 客服列表和选中状态
  const agents = ref<any[]>([])
  const selectedAgent = ref<any>(null)
  const isLoadingAgents = ref(false)
  
  const welcomeAgents = computed(() => agents.value.slice(0, 3))
  
  // 在线客服数量
  const onlineAgentsCount = computed(() => agents.value.length)
  
  watch([showWelcomeScreen, welcomeAgents], () => {
    if (!showWelcomeScreen.value) return
    if (!welcomeAgents.value.length) return
  
    const ids = welcomeAgents.value.map(agent => String(agent?.id ?? ''))
    const currentId = selectedAgent.value?.id != null ? String(selectedAgent.value.id) : ''
    if (!currentId || !ids.includes(currentId)) {
      selectedAgent.value = welcomeAgents.value[1] || welcomeAgents.value[0]
    }
  }, { immediate: true })
  
  // 全局邮箱设置
  const emailSettings = ref({
    preSalesEmail: '',
    afterSalesEmail: ''
  })
  const visitorEmail = ref('')
  const showVisitorEmailCapture = computed(() => !user.value)

  if (import.meta.client) {
    visitorEmail.value = localStorage.getItem('tanzanite_chat_visitor_email') || ''
  }

  watch(visitorEmail, (value) => {
    if (!import.meta.client) return
    const normalized = value.trim().toLowerCase()
    if (normalized) {
      localStorage.setItem('tanzanite_chat_visitor_email', normalized)
    } else {
      localStorage.removeItem('tanzanite_chat_visitor_email')
    }
  })
  
  const chatRooms = ref<Record<number, ChatRoomState>>({})
  
  const isSending = ref(false)
  
  const ensureChatRoom = (agentId: number): ChatRoomState => {
    if (!chatRooms.value[agentId]) {
      chatRooms.value[agentId] = createEmptyChatRoom()
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
  
  // 图片上传
  const isUploadingImage = ref(false)
  
  // 生成会话ID（基于访客标识）
  const conversationId = ref('')
  const STORAGE_KEY = computed(() => {
    return getChatStorageKey(conversationId.value, selectedAgent.value?.id)
  })
  const STORAGE_EXPIRY_DAYS = CHAT_STORAGE_EXPIRY_DAYS
  // Toast 提示
  const showToast = ref(false)
  const toastMessage = ref('')
  let toastTimer: number | null = null
  
  const messagePressTimer = ref<number | null>(null)
  const pressedMessage = ref<any | null>(null)
  
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
  
  // 显示 Toast 提示
  const displayToast = (message: string, duration = 2000) => {
    toastMessage.value = message
    showToast.value = true
    
    if (toastTimer) clearTimeout(toastTimer)
    toastTimer = setTimeout(() => {
      showToast.value = false
    }, duration)
  }
  
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
  
  // 滚动到底部
  const scrollToBottom = () => {
    nextTick()
  }
  
  // 监听消息变化，自动滚动到底部
  watch(messages, () => {
    scrollToBottom()
  }, { deep: true })
  
  // 监听客服切换，加载对应的聊天记录
  watch(() => selectedAgent.value?.id, (newId, oldId) => {
    if (newId && newId !== oldId) {
      saveLastSelectedAgentId(newId)
      loadMessagesFromStorage()
      if (conversationId.value) {
        loadMessagesFromAPI()
        connectCustomerServiceRealtime()
      }
      scrollToBottom()
    }
  })

  watch(conversationId, (newId, oldId) => {
    if (!import.meta.client) return
    if (newId && newId !== oldId) {
      loadMessagesFromAPI()
      connectCustomerServiceRealtime()
      return
    }
    if (!newId) {
      closeCustomerServiceRealtime()
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
      const storedRoom = loadChatRoomFromStorage(STORAGE_KEY.value, STORAGE_EXPIRY_DAYS)
      if (storedRoom) {
        Object.assign(currentRoom, storedRoom.room)

        if (storedRoom.hasExpiredMessages) {
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
      saveChatRoomToStorage(STORAGE_KEY.value, currentRoom)
    } catch (error) {
      console.error('保存消息失败:', error)
    }
  }

  const {
    currentSenderEmail,
    loadMessagesFromAPI,
    sendMessageToAPI,
    sendTypingIndicator,
    replaceLocalMessageWithServerMessage,
    markLocalMessageFailed,
    checkAutoReply,
    sendWelcomeMessage,
    connectCustomerServiceRealtime,
    closeCustomerServiceRealtime,
    agentTyping,
    clearAgentTyping
  } = useCustomerServiceChatSync({
    publicApiBase,
    conversationId,
    selectedAgent,
    messages,
    user,
    visitorEmail,
    authRequest,
    saveMessagesToStorage,
    scrollToBottom
  })

  const customerTypingSignalGapMs = 2500
  let lastCustomerTypingSignalAt = 0
  let customerTypingIdleTimer: number | null = null

  const notifyCustomerTyping = (isTyping = true) => {
    if (!import.meta.client || !selectedAgent.value) return

    if (isTyping) {
      const now = Date.now()
      if (now - lastCustomerTypingSignalAt < customerTypingSignalGapMs) return
      lastCustomerTypingSignalAt = now
    } else {
      lastCustomerTypingSignalAt = 0
    }

    sendTypingIndicator(isTyping)
  }

  watch(newMessage, (value) => {
    if (!import.meta.client || activeTab.value !== 'chat') return

    if (!String(value || '').trim()) {
      if (customerTypingIdleTimer) {
        window.clearTimeout(customerTypingIdleTimer)
        customerTypingIdleTimer = null
      }
      notifyCustomerTyping(false)
      return
    }

    notifyCustomerTyping(true)
    if (customerTypingIdleTimer) {
      window.clearTimeout(customerTypingIdleTimer)
    }
    customerTypingIdleTimer = window.setTimeout(() => {
      customerTypingIdleTimer = null
      notifyCustomerTyping(false)
    }, 3500)
  })

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
      sender_email: currentSenderEmail(),
      message: messageText,
      message_type: 'text',
      created_at: new Date().toISOString(),
      is_agent: false,
      sync_state: 'sending'
    }
  
    try {
      // 1. 先添加到本地显示
      messages.value.push(messageData)
      scrollToBottom()
      
      // 2. 保存到 localStorage
      saveMessagesToStorage()
      
      // 3. 发送到后端 API（实时存储）
      const response = await sendMessageToAPI(messageData)
      replaceLocalMessageWithServerMessage(messageData.id, response)
      
      // 4. 检查关键词自动回复
      await checkAutoReply(messageText)
    } catch (error) {
      markLocalMessageFailed(messageData.id)
      console.error('发送失败', error)
      // 可以添加重试逻辑或提示用户
    } finally {
      isSending.value = false
    }
  }
  
  // 搜索商品
  const searchProducts = async () => {
    const trimmedQuery = searchQuery.value.trim()
  
    // 如果关键字为空：仍然打开抽屉，只显示空状态，方便确认组件是否挂载
    if (!trimmedQuery) {
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
      const response = await $fetch<any>(`${publicApiBase.value}/customer-service/products`, {
        params: {
          keyword: trimmedQuery,
          per_page: 20
        }
      })
      
      // 转换数据格式以适配前端显示
      if (!response || !Array.isArray(response.items)) { throw new Error('[CRITICAL] Invalid response format for products API'); }
      
      searchResults.value = response.items.map((item: any) => {
        const normalized = normalizeShopProduct({
          ...item,
          url: item.preview_url || item.url,
        })
        return {
          ...normalized,
          name: normalized.title,
          url: item.preview_url || normalized.url,
          priceValue: normalized.priceNumber,
          price: normalized.priceLabel,
          maxStock: normalized.stockQuantity || 0,
        }
      })
      
    } catch (error) {
      console.error('搜索失败:', error)
      productDrawerError.value = 'Search failed, please try again.'
      searchResults.value = []
    } finally {
      isSearching.value = false
    }
  }
  
  const handleAddProductToCart = (product: any) => {
    const result = addToCart({
      id: product.id,
      product_id: product.id,
      variant_id: product.defaultVariantId || null,
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
  
  const shareProductMessageToChat = async (product: any, errorLabel: string) => {
    if (!selectedAgent.value || isSending.value) return
    
    isSending.value = true
    
    const messageData = {
      id: Date.now(),
      conversation_id: conversationId.value,
      sender_id: user.value?.id || 0,
      sender_name: user.value?.display_name || '访客',
      sender_email: currentSenderEmail(),
      message: product.title || '商品',
      message_type: 'product',
      metadata: {
        title: product.title,
        url: product.url,
        thumbnail: product.thumbnail,
        price: product.price
      },
      created_at: new Date().toISOString(),
      is_agent: false,
      sync_state: 'sending'
    }
    
    try {
      messages.value.push(messageData)
      saveMessagesToStorage()
      const response = await sendMessageToAPI(messageData)
      replaceLocalMessageWithServerMessage(messageData.id, response)
      activeTab.value = 'chat'
      scrollToBottom()
    } catch (error) {
      markLocalMessageFailed(messageData.id)
      console.error(errorLabel, error)
    } finally {
      isSending.value = false
    }
  }

  // 分享商品到聊天
  const shareProductToChat = (product: any) => {
    return shareProductMessageToChat(product, '分享商品失败:')
  }

  const shareProductConfigConfirmToChat = async (payload: any) => {
    const product = payload?.product || payload
    const variant = payload?.variant || payload?.selectedVariant || null
    if (!selectedAgent.value || isSending.value || !product) return

    isSending.value = true

    const metadata = buildProductConfigConfirmMetadata(product, variant)
    const productTitle = metadata.product.title || product.title || product.name || 'Product'
    const messageData = {
      id: Date.now(),
      conversation_id: conversationId.value,
      sender_id: user.value?.id || 0,
      sender_name: user.value?.display_name || '访客',
      sender_email: currentSenderEmail(),
      message: `Configuration confirmation request: ${productTitle}`,
      message_type: 'config_confirm',
      metadata,
      created_at: new Date().toISOString(),
      is_agent: false,
      sync_state: 'sending'
    }

    try {
      messages.value.push(messageData)
      saveMessagesToStorage()
      const response = await sendMessageToAPI(messageData)
      replaceLocalMessageWithServerMessage(messageData.id, response)
      activeTab.value = 'chat'
      handleProductDrawerClose()
      displayToast('Configuration request sent', 1800)
      scrollToBottom()
    } catch (error) {
      markLocalMessageFailed(messageData.id)
      console.error('发送配置确认失败:', error)
    } finally {
      isSending.value = false
    }
  }
  
  // 从浏览历史分享商品到聊天
  const handleShareProductFromHistory = (product: any) => {
    return shareProductMessageToChat(product, '从浏览历史分享商品失败:')
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

    const metadata = buildOrderChatMetadata(order)
    
    const messageData = {
      id: Date.now(),
      conversation_id: conversationId.value,
      sender_id: user.value?.id || 0,
      sender_name: user.value?.display_name || '访客',
      sender_email: currentSenderEmail(),
      message: `Order confirmation request: ${metadata.order_number || metadata.order_id}`,
      message_type: 'order',
      metadata,
      created_at: new Date().toISOString(),
      is_agent: false,
      sync_state: 'sending'
    }
    
    try {
      messages.value.push(messageData)
      saveMessagesToStorage()
      const response = await sendMessageToAPI(messageData)
      replaceLocalMessageWithServerMessage(messageData.id, response)
      activeTab.value = 'chat'
      scrollToBottom()
    } catch (error) {
      markLocalMessageFailed(messageData.id)
      console.error('分享订单失败:', error)
    } finally {
      isSending.value = false
    }
  }
  
  // 获取客服列表（带缓存）
  const fetchAgents = async () => {
    isLoadingAgents.value = true
    try {
      const directory = await loadChatAgentDirectory({
        apiBase: publicApiBase.value,
        currentUserId: user.value?.id,
        allowDevFallback: import.meta.dev
      })

      if (directory.emailSettings) {
        emailSettings.value = directory.emailSettings
      }

      if (directory.agents.length > 0) {
        agents.value = directory.agents
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
    const storedId = getLastSelectedAgentId()
    if (storedId) {
      const matched = agents.value.find(agent => String(agent.id) === storedId)
      if (matched) {
        defaultAgent = matched
      }
    }
  
    if (!selectedAgent.value || selectedAgent.value.id !== defaultAgent.id) {
      selectedAgent.value = defaultAgent
      ensureChatRoom(defaultAgent.id)
      loadMessagesFromStorage()
      await sendWelcomeMessage()
    }
  }
  
  // 选择客服
  const selectAgent = (agent: any) => {
    if (selectedAgent.value?.id === agent.id) return
    selectedAgent.value = agent
    ensureChatRoom(agent.id)
    loadMessagesFromStorage()
  }
  
  const agentThemePalette = ['#6b73ff', '#40ffaa', '#C77DFF']
  const getAgentThemeColor = (agentId: number) => {
    return agentThemePalette[(agentId - 1) % agentThemePalette.length] || agentThemePalette[0]
  }
  
  const currentThemeColor = computed(() => {
    if (!selectedAgent.value?.id) return agentThemePalette[0]
    return getAgentThemeColor(selectedAgent.value.id)
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
          sender_email: currentSenderEmail(),
          message: '[图片]',
          message_type: 'image',
          attachment_url: imageUrl,
          created_at: new Date().toISOString(),
          is_agent: false,
          sync_state: 'sending'
        }
        
        // 添加到消息列表
        messages.value.push(messageData)
        saveMessagesToStorage()
        scrollToBottom()
        
        // 发送到后端
        try {
          const response = await sendMessageToAPI(messageData)
          replaceLocalMessageWithServerMessage(messageData.id, response)
        } catch (error) {
          markLocalMessageFailed(messageData.id)
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
  
  // 组件挂载时获取客服列表、会员数据和检查历史对话
  onMounted(async () => {
    await initMembership()
    await fetchAgents()
    initHistoryChatCheck()
    scrollToBottom()
  })

  onBeforeUnmount(() => {
    if (customerTypingIdleTimer) {
      window.clearTimeout(customerTypingIdleTimer)
      customerTypingIdleTimer = null
    }
    notifyCustomerTyping(false)
    clearAgentTyping()
    closeCustomerServiceRealtime()
  })
  
  return {
    user,
    showWelcomeScreen,
    hasHistoryChat,
    agents,
    selectedAgent,
    welcomeAgents,
    onlineAgentsCount,
    emailSettings,
    visitorEmail,
    showVisitorEmailCapture,
    isSending,
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
    isUploadingImage,
    agentTyping,
    showToast,
    toastMessage,
    isMemberLogged,
    levelName,
    points,
    tierInfo,
    levelDiscounts,
    userCoupons,
    userPointCards,
    isLoggedInForWarranty,
    showAuthModal,
    authMode,
    currentThemeColor,
    openMemberAuth,
    handleWarrantyLoginRequest,
    handleChatAuthSuccess,
    handleClose,
    enterChat,
    selectAgentFromWelcome,
    handleMessageContextMenu,
    handleSendMessage,
    searchProducts,
    handleAddProductToCart,
    handleProductDrawerClose,
    handleHistoryDrawerClose,
    shareProductToChat,
    shareProductConfigConfirmToChat,
    handleShareProductFromHistory,
    shareOrderToChat,
    openCartFromChat,
    getInitials,
    handleImageUpload
  }
}
