<template>
  <Teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        v-if="conversation"
        class="fixed inset-0 z-[9999] flex items-center justify-center p-0 md:p-4"
        @click.self="handleClose"
      >
        <!-- 半透明背景遮罩 -->
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>
        <!-- 聊天窗口 - 三栏布局 -->
        <Transition name="slide-up">
          <div
            v-if="conversation"
            class="relative border-2 border-[#6b73ff] rounded-2xl shadow-[0_0_30px_rgba(107,115,255,0.3)] max-w-[1400px] w-full h-[90vh] md:h-[700px] max-h-[85vh] overflow-hidden flex flex-row transition-colors duration-300 bg-black"
          >
            <!-- 左侧：客服列表(桌面 400px) - 移动端隐藏 -->
            <div class="hidden md:flex w-[400px] min-w-[400px] max-w-[400px] border-r border-white/10 flex-col" style="background-color: rgba(0, 0, 0, 0.5) !important;">
              <!-- 客服列表标题 -->
              <div class="px-4 py-4 border-b border-white/10">
                <h3 class="font-semibold text-sm bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] bg-clip-text text-transparent">Agents</h3>
              </div>

              <!-- 桌面搜索框 -->
              <div class="px-4 pt-3 pb-4 border-b border-white/5 bg-white/5">
                <div class="relative">
                  <span class="absolute left-3 top-1/2 -translate-y-1/2 text-white/50 text-sm">🔍</span>
                  <input
                    v-model="desktopSearchQuery"
                    type="text"
                    placeholder="Search chats..."
                    class="w-full h-10 pl-9 pr-3 rounded-full bg-black/50 text-white text-sm border border-white/20 focus:outline-none focus:border-[#6b73ff] placeholder:text-white/40"
                  />
                </div>
                <p class="text-[11px] text-white/40 mt-2">Search results will appear here (UI only).</p>
              </div>

              <!-- 客服列表 -->
              <div class="flex-1 overflow-y-auto">
                <div v-if="isLoadingAgents" class="flex items-center justify-center py-8">
                  <svg class="animate-spin h-6 w-6 text-white/50" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </div>
                
                <div
                  v-for="agent in agents"
                  :key="agent.id"
                  @click="selectAgent(agent)"
                  class="px-3 py-3 cursor-pointer transition-colors border-b border-white/5"
                  :class="selectedAgent?.id === agent.id 
                    ? 'bg-[#6b73ff]/20 border-l-2 border-l-[#6b73ff]' 
                    : 'hover:bg-white/5'"
                >
                  <div class="flex items-center gap-2">
                    <!-- 头像 -->
                    <div class="w-10 h-10 rounded-full bg-gradient-to-br from-[#40ffaa] to-[#6b73ff] flex items-center justify-center text-white font-semibold text-xs flex-shrink-0">
                      <img
                        v-if="agent.avatar"
                        :src="agent.avatar"
                        :alt="agent.name"
                        class="w-full h-full rounded-full object-cover"
                      />
                      <span v-else>{{ getInitials(agent.name) }}</span>
                    </div>
                    
                    <!-- 信息 -->
                    <div class="flex-1 min-w-0">
                      <div class="text-white text-sm font-medium truncate">
                        {{ agent.name }}
                      </div>
                      <div class="text-white/50 text-xs truncate">
                        {{ agent.email }}
                      </div>
                    </div>
                  </div>
                  <div v-if="selectedAgent?.id === agent.id" class="mt-3 flex flex-col gap-2">
                    <div class="flex gap-2">
                      <a
                        v-if="agent.whatsapp"
                        :href="`https://wa.me/${agent.whatsapp.replace('+', '')}`"
                        target="_blank"
                        class="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-xs font-semibold text-black bg-[#25D366] hover:bg-[#20BA5A] transition-colors"
                        @click.stop
                      >
                        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                          <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884" />
                        </svg>
                        WhatsApp
                      </a>
                      <button
                        v-if="conversation"
                        @click.stop="showTransferModal = true"
                        class="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-xs font-semibold text-white bg-blue-600 hover:bg-blue-700 transition-colors"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                        </svg>
                        转接
                      </button>
                    </div>
                  </div>
                </div>
                <div v-if="!isLoadingAgents && agents.length === 0" class="text-center text-white/50 py-8 text-sm">
                  No agents available
                </div>
              </div>
              
              <!-- FAQ + 邮箱按钮区域 - 固定在底部 -->
              <div class="border-t border-white/10 p-3 space-y-2">
                <button
                  @click="showFAQ = true"
                  class="w-full px-3 py-2 rounded-lg text-xs font-semibold border border-white/20 text-white hover:bg-white/10 transition-colors"
                >
                  FAQ
                </button>
                <!-- Pre-sales 邮箱按钮 -->
                <a
                  :href="emailSettings.preSalesEmail ? `mailto:${emailSettings.preSalesEmail}?subject=Pre-sales Inquiry` : 'javascript:void(0)'"
                  :class="[
                    'w-full px-3 py-2 rounded-lg text-xs transition-all inline-flex items-center justify-center gap-2',
                    emailSettings.preSalesEmail 
                      ? 'text-white cursor-pointer shadow-lg' 
                      : 'text-gray-400 cursor-not-allowed opacity-50'
                  ]"
                  :style="emailSettings.preSalesEmail ? 'background: linear-gradient(to right, #60D5FF, #4A90E2) !important;' : 'background-color: #4b5563 !important;'"
                  :title="emailSettings.preSalesEmail ? 'Pre-sales Email' : 'No email configured'"
                  @click="!emailSettings.preSalesEmail && $event.preventDefault()"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                  </svg>
                  <span>Pre-sales</span>
                </a>
                
                <!-- After-sales 邮箱按钮 -->
                <a
                  :href="emailSettings.afterSalesEmail ? `mailto:${emailSettings.afterSalesEmail}?subject=After-sales Support` : 'javascript:void(0)'"
                  :class="[
                    'w-full px-3 py-2 rounded-lg text-xs transition-all inline-flex items-center justify-center gap-2',
                    emailSettings.afterSalesEmail 
                      ? 'text-white cursor-pointer shadow-lg' 
                      : 'text-gray-400 cursor-not-allowed opacity-50'
                  ]"
                  :style="emailSettings.afterSalesEmail ? 'background: linear-gradient(to right, #C77DFF, #9B59B6) !important;' : 'background-color: #4b5563 !important;'"
                  :title="emailSettings.afterSalesEmail ? 'After-sales Email' : 'No email configured'"
                  @click="!emailSettings.afterSalesEmail && $event.preventDefault()"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                  </svg>
                  <span>After-sales</span>
                </a>
              </div>
            </div>

            <!-- 中间：聊天区域(主栏) -->
            <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
              <!-- 头部 - 固定高度避免跳动 -->
              <div class="border-b border-white/10">
                <!-- 移动端：操作按钮 + 客服标签 -->
                <div class="md:hidden bg-black/70 backdrop-blur-md shadow-[0_4px_16px_rgba(0,0,0,0.35)]">
                  <!-- 第一排：邮箱图标 + FAQ + 关闭 -->
                  <div class="px-3 pt-2 pb-2 border-b border-white/10">
                    <div class="flex items-center gap-1.5 h-[42px]">
                      <div class="flex-1 flex gap-1.5 h-full">
                        <a
                          v-if="emailSettings.preSalesEmail"
                          :href="`mailto:${emailSettings.preSalesEmail}?subject=Pre-sales Inquiry`"
                          class="flex-1 h-full rounded-full flex items-center justify-center shadow-sm transition-all"
                          style="background: linear-gradient(120deg, #60D5FF, #4A90E2);"
                          aria-label="Pre-sales email"
                        >
                          <svg class="w-[70%] h-[70%]" fill="none" stroke="black" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                          </svg>
                        </a>
                        <a
                          v-if="emailSettings.afterSalesEmail"
                          :href="`mailto:${emailSettings.afterSalesEmail}?subject=After-sales Support`"
                          class="flex-1 h-full rounded-full flex items-center justify-center shadow-sm transition-all"
                          style="background: linear-gradient(120deg, #C77DFF, #9B59B6);"
                          aria-label="After-sales email"
                        >
                          <svg class="w-4.5 h-4.5" fill="none" stroke="black" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                          </svg>
                        </a>
                        <button
                          class="flex-1 h-full rounded-full bg-white/15 text-white flex items-center justify-center transition-colors hover:bg-white/25"
                          @click="showFAQ = true"
                          aria-label="Open FAQ"
                        >
                          <svg class="w-[70%] h-[70%]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.09 9a3 3 0 015.83 1c0 2-3 3-3 3m.01 4h.01M21 12c0 4.418-4.03 8-9 8s-9-3.582-9-8 4.03-8 9-8 9 3.582 9 8z" />
                          </svg>
                        </button>
                      </div>
                      <button
                        @click="handleClose"
                        class="w-[42px] h-[42px] rounded-full border-2 border-red-500 text-red-400 flex items-center justify-center hover:bg-red-500/15 transition-colors"
                        aria-label="Close"
                      >
                        <svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                  </div>

                  <!-- 第二排：客服标签 -->
                  <div class="px-3 pb-3 pt-2">
                    <div v-if="agents.length > 0" class="grid grid-cols-3 gap-1.5">
                      <button
                        v-for="agent in agents"
                        :key="agent.id"
                        @click="selectAgent(agent)"
                        class="h-[42px] rounded-full text-xs font-semibold uppercase tracking-wide transition-all border"
                        :class="selectedAgent?.id === agent.id
                          ? 'text-black shadow-[0_4px_12px_rgba(0,0,0,0.35)]'
                          : 'bg-black/40 text-white/70'"
                        :style="selectedAgent?.id === agent.id
                          ? { backgroundColor: getAgentThemeColor(agent.id), borderColor: getAgentThemeColor(agent.id) }
                          : { borderColor: getAgentThemeColor(agent.id), color: getAgentThemeColor(agent.id) }"
                      >
                        <span class="truncate px-2">{{ agent.name }}</span>
                      </button>
                    </div>
                    <div v-else class="text-center text-white/60 text-sm py-2">
                      Loading agents...
                    </div>
                  </div>
                </div>

                <!-- 桌面端：仅保留提示 + 关闭按钮 -->
                <div class="hidden md:flex items-center justify-end px-6 py-4">
                  <div class="text-white/50 text-sm mr-auto" v-if="!selectedAgent">
                    Select an agent to start chat
                  </div>
                  <button
                    @click="handleClose"
                    class="w-10 h-10 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors"
                    aria-label="Close"
                  >
                    <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>

              <!-- 移动端：Chrome 样式主题容器 -->
              <div v-if="selectedAgent" class="md:hidden flex-1 min-h-0 px-3 pb-4">
                <div class="flex flex-col h-full rounded-[28px] border-2 overflow-hidden" :style="mobilePanelStyle">
                  <!-- 第三排：功能按钮 -->
                  <div class="flex gap-1.5 px-3 pt-4 pb-2">
                    <button
                      @click="activeTab = 'chat'"
                      class="flex-1 h-10 rounded-full text-xs font-semibold tracking-wide transition-all"
                      :style="activeTab === 'chat'
                        ? { backgroundColor: '#000', color: currentThemeColor }
                        : { backgroundColor: 'rgba(0,0,0,0.35)', color: '#fff' }"
                    >
                      Chat
                    </button>
                    <button
                      @click="activeTab = 'share'"
                      class="flex-1 h-10 rounded-full text-xs font-semibold tracking-wide transition-all"
                      :style="activeTab === 'share'
                        ? { backgroundColor: '#000', color: currentThemeColor }
                        : { backgroundColor: 'rgba(0,0,0,0.35)', color: '#fff' }"
                    >
                      Products
                    </button>
                    <button
                      @click="activeTab = 'orders'"
                      class="flex-1 h-10 rounded-full text-xs font-semibold tracking-wide transition-all"
                      :style="activeTab === 'orders'
                        ? { backgroundColor: '#000', color: currentThemeColor }
                        : { backgroundColor: 'rgba(0,0,0,0.35)', color: '#fff' }"
                    >
                      Orders
                    </button>
                  </div>

                  <!-- 第四排：WhatsApp -->
                  <div class="px-3 pb-3">
                    <button
                      v-if="selectedAgent?.whatsapp"
                      @click.prevent="handleWhatsAppClick(selectedAgent)"
                      @touchstart="handleWhatsAppTouchStart(selectedAgent)"
                      @touchend="handleWhatsAppTouchEnd"
                      @touchcancel="handleWhatsAppTouchEnd"
                      class="w-full flex items-center justify-center gap-2 rounded-2xl py-2.5 text-sm font-semibold text-black"
                      :style="{ backgroundColor: currentThemeColor }"
                    >
                      <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                        <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                      </svg>
                      WhatsApp
                    </button>
                    <div v-else class="w-full rounded-2xl py-2.5 text-center text-sm text-white/60 border border-white/30">
                      WhatsApp unavailable
                    </div>
                  </div>

                  <!-- 内容区域 -->
                  <div class="flex-1 min-h-0 overflow-hidden px-2 pb-3">
                    <div
                      v-if="activeTab === 'chat'"
                      ref="messagesContainerMobile"
                      class="h-full overflow-y-auto space-y-3 px-1"
                    >
                      <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-white/70 text-sm">
                        <svg class="w-12 h-12 mb-2 text-white/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                        </svg>
                        No messages yet
                      </div>
                      <div
                        v-for="message in messages"
                        :key="message.id"
                        class="flex"
                        :class="message.is_agent ? 'justify-end' : 'justify-start'"
                      >
                        <a
                          v-if="message.type === 'card'"
                          :href="message.url || '#'"
                          target="_blank"
                          rel="noopener"
                          class="flex gap-2.5 p-2 border border-white/20 rounded-2xl bg-black/40 max-w-[75%]"
                        >
                          <img
                            v-if="message.thumbnail"
                            :src="message.thumbnail"
                            alt="thumbnail"
                            class="w-14 h-14 object-cover rounded-xl"
                          />
                          <div class="text-xs text-white">{{ message.title || message.message }}</div>
                        </a>
                        <div
                          v-else
                          class="max-w-[75%] rounded-2xl px-3 py-2 text-white shadow-lg"
                          :style="message.is_agent
                            ? { backgroundColor: 'rgba(0,0,0,0.4)', border: `1px solid ${currentThemeColor}` }
                            : { backgroundColor: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.2)' }"
                          @touchstart="handleMessageTouchStart(message)"
                          @touchend="handleMessageTouchEnd"
                          @touchcancel="handleMessageTouchEnd"
                          @mousedown="handleMessageMouseDown(message)"
                          @mouseup="handleMessageMouseUp"
                          @mouseleave="handleMessageMouseUp"
                          @contextmenu.prevent="handleMessageContextMenu(message)"
                        >
                          <div class="text-[11px] mb-1 opacity-70">
                            {{ message.is_agent ? 'Agent' : message.sender_name }}
                          </div>
                          <div class="text-sm whitespace-pre-wrap break-words">
                            {{ message.message }}
                          </div>
                          <div class="text-[10px] opacity-50 mt-1">
                            {{ formatMessageTime(message.created_at) }}
                          </div>
                          <div v-if="message.attachment_url" class="mt-2">
                            <img :src="message.attachment_url" alt="附件" class="max-w-full rounded-xl" />
                          </div>
                        </div>
                      </div>
                    </div>

                    <div v-else-if="activeTab === 'share'" class="h-full flex flex-col">
                      <div class="flex gap-2 mb-3 items-center">
                        <input
                          v-model="searchQuery"
                          type="text"
                          placeholder="Search products..."
                          class="flex-1 h-10 px-3 rounded-xl bg-black/40 text-white text-sm border border-white/30 focus:outline-none"
                          @keydown.enter.prevent="searchProducts"
                        />
                        <button
                          @click="searchProducts"
                          :disabled="isSearching"
                          class="px-3 h-10 rounded-xl text-sm font-semibold text-white border border-white/40 disabled:opacity-50"
                        >
                          {{ isSearching ? 'Searching...' : 'Search' }}
                        </button>
                      </div>
                      <div v-if="!productDrawerVisible" class="flex-1 overflow-y-auto space-y-3 pr-1">
                        <div
                          v-for="product in searchResults"
                          :key="product.id"
                          @click="shareProductToChat(product)"
                          class="border border-white/10 rounded-2xl p-3 bg-black/30"
                        >
                          <img
                            v-if="product.thumbnail"
                            :src="product.thumbnail"
                            alt="商品图片"
                            class="w-full h-28 object-cover rounded-xl mb-2"
                          />
                          <h4 class="text-white text-sm font-semibold truncate">{{ product.title }}</h4>
                          <p v-if="product.price" class="text-white/70 text-xs mt-1">{{ product.price }}</p>
                        </div>
                        <div v-if="!isSearching && searchResults.length === 0" class="text-center text-white/60 text-sm py-8">
                          {{ searchQuery ? 'No products found' : 'Search products to share' }}
                        </div>
                      </div>
                    </div>

                    <div v-else class="h-full overflow-y-auto space-y-3 px-1">
                      <div v-if="isLoadingOrders" class="text-center text-white/60 py-10 text-sm">
                        Loading orders...
                      </div>
                      <div v-else-if="ordersList.length > 0" class="space-y-2">
                        <div
                          v-for="order in ordersList"
                          :key="order.id"
                          @click="shareOrderToChat(order)"
                          class="border border-white/15 rounded-2xl p-3 bg-black/35"
                        >
                          <div class="flex items-center justify-between mb-1">
                            <span class="text-white text-sm font-semibold">Order #{{ order.id }}</span>
                            <span class="text-[10px] px-2 py-0.5 rounded-full bg-white/15 text-white/70">{{ order.status || 'Processing' }}</span>
                          </div>
                          <p class="text-white/70 text-xs">{{ order.total }} {{ order.currency || '' }}</p>
                          <p class="text-white/50 text-[11px] mt-1">{{ order.date }}</p>
                        </div>
                      </div>
                      <div v-else class="text-center text-white/60 text-sm py-10">
                        No orders yet
                      </div>
                    </div>
                  </div>

                  <!-- 输入区 -->
                  <div v-if="activeTab === 'chat'" class="px-3 pb-4 border-t border-white/15">
                    <form @submit.prevent="handleSendMessage" class="flex items-center gap-2">
                      <input
                        v-model="newMessage"
                        type="text"
                        placeholder="Type a message..."
                        class="flex-1 h-11 px-4 rounded-full text-sm text-white bg-black/40 border"
                        :style="{ borderColor: currentThemeColor }"
                        :disabled="isSending"
                      />
                      <input
                        ref="imageInput"
                        type="file"
                        accept="image/*"
                        class="hidden"
                        @change="handleImageUpload"
                      />
                      <button
                        type="button"
                        @click="imageInput?.click()"
                        :disabled="isUploadingImage"
                        class="w-10 h-10 rounded-full border border-white/40 text-white flex items-center justify-center disabled:opacity-50"
                      >
                        <svg v-if="!isUploadingImage" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                        </svg>
                        <svg v-else class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                      </button>
                      <button
                        type="submit"
                        :disabled="!newMessage.trim() || isSending"
                        class="px-4 h-11 rounded-full font-semibold text-sm text-black"
                        :style="{ backgroundColor: currentThemeColor }"
                      >
                        <span v-if="!isSending">Send</span>
                        <span v-else class="flex items-center gap-2">
                          <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                          </svg>
                          Sending...
                        </span>
                      </button>
                    </form>
                  </div>
                </div>
              </div>
              <div v-else class="md:hidden text-center text-white/60 py-10">
                Select an agent to start chat
              </div>

              <!-- 桌面端内容保持不变 -->
              <div class="hidden md:flex flex-col flex-1 min-h-0">
                <div class="flex gap-2 justify-center py-3 border-b border-white/10 px-2">
                  <button
                    @click="activeTab = 'chat'"
                    class="px-4 py-1.5 rounded-full text-sm transition-all"
                    :class="activeTab === 'chat' 
                      ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white' 
                      : 'bg-white/[0.08] text-white/70 border border-white hover:bg-white/[0.15]'"
                  >
                    Chat
                  </button>
                  <button
                    @click="activeTab = 'share'"
                    class="px-4 py-1.5 rounded-full text-sm transition-all whitespace-nowrap"
                    :class="activeTab === 'share' 
                      ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white' 
                      : 'bg-white/[0.08] text-white/70 border border-white hover:bg-white/[0.15]'"
                  >
                    Share Products
                  </button>
                  <button
                    @click="activeTab = 'orders'"
                    class="px-4 py-1.5 rounded-full text-sm transition-all whitespace-nowrap"
                    :class="activeTab === 'orders' 
                      ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white' 
                      : 'bg-white/[0.08] text-white/70 border border-white hover:bg-white/[0.15]'"
                  >
                    My Orders
                  </button>
                </div>

                <div v-if="activeTab === 'chat'" ref="messagesContainerDesktop" class="flex-1 overflow-y-auto p-6 space-y-4">
                  <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full">
                    <svg class="w-16 h-16 text-white/30 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                    </svg>
                    <p class="text-white/50">No messages yet</p>
                  </div>
                  <div
                    v-for="message in messages"
                    :key="message.id"
                    class="flex"
                    :class="message.is_agent ? 'justify-end' : 'justify-start'"
                  >
                    <a
                      v-if="message.type === 'card'"
                      :href="message.url || '#'"
                      target="_blank"
                      rel="noopener"
                      class="flex gap-2.5 p-2 border border-white/[0.18] rounded-[10px] bg-white/[0.06] hover:bg-white/[0.10] transition-colors max-w-[70%]"
                    >
                      <img
                        v-if="message.thumbnail"
                        :src="message.thumbnail"
                        alt="thumbnail"
                        class="w-14 h-14 object-cover rounded-lg"
                      />
                      <div class="text-sm text-white">{{ message.title || message.message }}</div>
                    </a>
                    <div
                      v-else
                      class="max-w-[70%] rounded-xl px-3 py-2 text-white shadow-lg"
                      :class="message.is_agent 
                        ? 'bg-[rgba(64,255,170,0.35)] border border-[rgba(64,255,170,0.6)]' 
                        : 'bg-[rgba(64,122,255,0.35)] border border-[rgba(64,122,255,0.6)]'"
                      @touchstart="handleMessageTouchStart(message)"
                      @touchend="handleMessageTouchEnd"
                      @touchcancel="handleMessageTouchEnd"
                      @mousedown="handleMessageMouseDown(message)"
                      @mouseup="handleMessageMouseUp"
                      @mouseleave="handleMessageMouseUp"
                      @contextmenu.prevent="handleMessageContextMenu(message)"
                    >
                      <div class="text-xs mb-1 opacity-70">
                        {{ message.is_agent ? 'Agent' : message.sender_name }}
                      </div>
                      <div class="flex items-end gap-2">
                        <div class="text-sm whitespace-pre-wrap break-words flex-1">
                          {{ message.message }}
                        </div>
                        <div class="text-[10px] opacity-60 whitespace-nowrap flex-shrink-0">
                          {{ formatMessageTime(message.created_at) }}
                        </div>
                      </div>
                      <div v-if="message.attachment_url" class="mt-2">
                        <img
                          :src="message.attachment_url"
                          alt="附件"
                          class="max-w-full rounded-lg"
                        />
                      </div>
                    </div>
                  </div>
                </div>

                <div v-if="activeTab === 'share'" class="flex-1 flex flex-col overflow-hidden">
                  <div class="flex-1 overflow-y-auto p-6">
                    <div class="flex gap-2 mb-4 items-center">
                      <input
                        v-model="searchQuery"
                        type="text"
                        placeholder="Search products..."
                        class="flex-1 h-[42px] px-3 rounded-lg bg-white/[0.06] text-white border border-white focus:outline-none focus:border-[#6b73ff] transition-colors text-sm"
                        @keydown.enter.prevent="searchProducts"
                      />
                      <button
                        @click="searchProducts"
                        :disabled="isSearching"
                        class="h-[42px] px-4 bg-white/[0.08] hover:bg-white/[0.15] text-white border border-white rounded-lg transition-colors disabled:opacity-50 whitespace-nowrap text-sm"
                      >
                        {{ isSearching ? 'Searching...' : 'Search' }}
                      </button>
                    </div>
                    <div v-if="searchResults.length > 0 && !productDrawerVisible" class="grid grid-cols-2 gap-3">
                      <div
                        v-for="product in searchResults"
                        :key="product.id"
                        @click="shareProductToChat(product)"
                        class="border border-white/10 rounded-lg p-3 hover:bg-white/[0.05] cursor-pointer transition-colors"
                      >
                        <img
                          v-if="product.thumbnail"
                          :src="product.thumbnail"
                          alt="商品图片"
                          class="w-full h-32 object-cover rounded-lg mb-2"
                        />
                        <h4 class="text-white text-sm font-medium truncate">{{ product.title }}</h4>
                        <p v-if="product.price" class="text-white/70 text-xs mt-1">{{ product.price }}</p>
                      </div>
                    </div>
                    <div v-else-if="!productDrawerVisible && !isSearching && searchQuery" class="text-center text-white/50 py-12">
                      No products found
                    </div>
                    <div v-else-if="!productDrawerVisible && !isSearching" class="text-center text-white/50 py-12">
                      Search products to share in chat
                    </div>
                  </div>
                  <div class="border-t border-white/10 p-4 bg-black/20">
                    <BrowsingHistoryDark @share-to-chat="handleShareProductFromHistory" />
                  </div>
                </div>

                <div v-if="activeTab === 'orders'" class="flex-1 overflow-y-auto p-6">
                  <div v-if="isLoadingOrders" class="text-center text-white/50 py-12">
                    Loading orders...
                  </div>
                  <div v-else-if="ordersList.length > 0" class="grid grid-cols-2 gap-3">
                    <div
                      v-for="order in ordersList"
                      :key="order.id"
                      @click="shareOrderToChat(order)"
                      class="border border-white/10 rounded-lg p-3 hover:bg-white/[0.05] cursor-pointer transition-colors"
                    >
                      <div class="flex items-center justify-between mb-2">
                        <span class="text-white text-sm font-medium">Order #{{ order.id }}</span>
                        <span class="text-xs px-2 py-0.5 rounded-full bg-white/10 text-white/70">
                          {{ order.status || 'Processing' }}
                        </span>
                      </div>
                      <p class="text-white/70 text-xs">{{ order.total }} {{ order.currency || '' }}</p>
                      <p class="text-white/50 text-xs mt-1">{{ order.date }}</p>
                    </div>
                  </div>
                  <div v-else class="text-center text-white/50 py-12">
                    No orders yet
                  </div>
                </div>

                <div v-if="activeTab === 'chat'" class="border-t border-white p-4">
                  <form @submit.prevent="handleSendMessage" class="flex gap-2">
                    <input
                      v-model="newMessage"
                      type="text"
                      placeholder="Type a message..."
                      class="flex-1 px-4 py-2.5 bg-white/[0.06] text-white border border-white rounded-full focus:outline-none focus:border-[#6b73ff] transition-colors text-base"
                      :disabled="isSending"
                    />
                    <input
                      ref="imageInput"
                      type="file"
                      accept="image/*"
                      class="hidden"
                      @change="handleImageUpload"
                    />
                    <button
                      type="button"
                      @click="imageInput?.click()"
                      :disabled="isUploadingImage"
                      class="w-11 h-11 bg-white/[0.08] hover:bg-white/[0.15] text-white border border-white rounded-full transition-colors disabled:opacity-50 flex items-center justify-center"
                      title="Upload image"
                    >
                      <svg v-if="!isUploadingImage" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                      </svg>
                      <svg v-else class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                    </button>
                    <button
                      type="submit"
                      :disabled="!newMessage.trim() || isSending"
                      class="px-6 py-2.5 bg-[#6b73ff] hover:bg-[#5d65e8] text-white rounded-full transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed text-base"
                    >
                      <span v-if="!isSending">Send</span>
                      <span v-else class="flex items-center gap-2">
                        <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Sending...
                      </span>
                    </button>
                  </form>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
    
    <!-- 转接弹窗 -->
    <Transition name="fade">
      <div
        v-if="showTransferModal"
        class="fixed inset-0 bg-black/50 z-[10000] flex items-center justify-center p-4"
        @click.self="showTransferModal = false"
      >
        <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-2xl">
          <h3 class="text-xl font-bold text-gray-900 mb-4">转接会话</h3>
          
          <div class="space-y-4">
            <!-- 选择客服 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                选择目标客服 *
              </label>
              <select
                v-model="transferToAgent"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">请选择客服</option>
                <option
                  v-for="agent in agents.filter(a => a.id !== selectedAgent?.id)"
                  :key="agent.id"
                  :value="agent.id"
                >
                  {{ agent.name }} ({{ agent.email }})
                </option>
              </select>
            </div>
            
            <!-- 转接备注 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                转接备注（可选）
              </label>
              <textarea
                v-model="transferNote"
                rows="3"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                placeholder="例如：客户需要技术支持..."
              ></textarea>
            </div>
          </div>
          
          <!-- 按钮 -->
          <div class="flex gap-3 mt-6">
            <button
              @click="showTransferModal = false"
              :disabled="isTransferring"
              class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
            >
              取消
            </button>
            <button
              @click="handleTransfer"
              :disabled="isTransferring || !transferToAgent"
              class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ isTransferring ? '转接中...' : '确认转接' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
    <!-- Toast 提示 -->
    <Transition name="fade">
      <div
        v-if="showToast"
        class="fixed bottom-20 left-1/2 -translate-x-1/2 z-[10001] px-4 py-2 bg-black/90 text-white text-sm rounded-lg shadow-lg backdrop-blur-sm"
      >
        {{ toastMessage }}
      </div>
    </Transition>
    
    <!-- FAQ 弹窗 - 移动端从底部滑出，桌面端居中 -->
    <Transition name="slide-up">
      <div
        v-if="showFAQ"
        class="fixed inset-0 z-[10001] flex items-end md:items-center justify-center p-0 md:p-4"
      >
        <FaqModal @close="showFAQ = false" />
      </div>
    </Transition>

    <WhatsAppProductSearchResultDrawer
      v-model="productDrawerVisible"
      :loading="isSearching"
      :results="searchResults"
      :error="productDrawerError"
      :agent="selectedAgent"
      :query="productDrawerQuery"
      @close="handleProductDrawerClose"
      @select="shareProductToChat"
    />
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'
import FaqModal from '~/components/FaqModal.vue'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'

// Props - 现在不需要预先传入conversation
const props = defineProps<{
  conversation?: {
    showAgentList?: boolean
  }
}>()

// Emits
const emit = defineEmits<{
  close: []
}>()

const { user } = useAuth()
const config = useRuntimeConfig()

// Desktop-only搜索占位
const desktopSearchQuery = ref('')

// 客服列表和选中状态
const agents = ref<any[]>([])
const selectedAgent = ref<any>(null)
const isLoadingAgents = ref(false)

// 全局邮箱设置
const emailSettings = ref({
  preSalesEmail: '',
  afterSalesEmail: ''
})

type ChatTab = 'chat' | 'share' | 'orders'
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
const messagesContainerDesktop = ref<HTMLElement | null>(null)
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
    get: () => currentChatRoom.value?.messages || [],
    set: (val) => {
      if (currentChatRoom.value) currentChatRoom.value.messages = val
    }
  }
)

const activeTab = computed<ChatTab>({
  get: () => currentChatRoom.value?.activeTab || 'chat',
  set: (val) => {
    if (currentChatRoom.value) currentChatRoom.value.activeTab = val
  }
})

const newMessage = computed({
  get: () => currentChatRoom.value?.newMessage || '',
  set: (val) => {
    if (currentChatRoom.value) currentChatRoom.value.newMessage = val
  }
})

const searchQuery = computed({
  get: () => currentChatRoom.value?.searchQuery || '',
  set: (val) => {
    if (currentChatRoom.value) currentChatRoom.value.searchQuery = val
  }
})

const searchResults = computed<any[]>({
  get: () => currentChatRoom.value?.searchResults || [],
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
  get: () => currentChatRoom.value?.ordersList || [],
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

// FAQ 弹窗
const showFAQ = ref(false)

// 是否显示"我的订单"标签
const shouldShowOrders = computed(() => !!user.value)

// 关闭弹窗
const handleClose = () => {
  emit('close')
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
    const containers = [messagesContainerMobile.value, messagesContainerDesktop.value]
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
      
      const validMessages = (data.messages || []).filter((msg: any) => {
        const msgTime = new Date(msg.created_at).getTime()
        return (now - msgTime) < expiryTime
      })

      currentRoom.messages = validMessages
      currentRoom.activeTab = (data.activeTab as ChatTab) || 'chat'
      currentRoom.newMessage = data.newMessage || ''
      currentRoom.searchQuery = data.searchQuery || ''
      currentRoom.searchResults = Array.isArray(data.searchResults) ? data.searchResults : []
      currentRoom.ordersList = Array.isArray(data.ordersList) ? data.ordersList : []
      currentRoom.isSearching = !!data.isSearching
      currentRoom.isLoadingOrders = !!data.isLoadingOrders

      if (validMessages.length !== (data.messages || []).length) {
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
    const response = await $fetch('/wp-json/tanzanite/v1/customer-service/messages', {
      method: 'POST',
      body: {
        conversation_id: conversationId.value,
        message: messageData.message,
        sender_type: user.value ? 'user' : 'visitor',
        sender_name: user.value?.display_name || '访客',
        sender_email: user.value?.email || '',
        agent_id: selectedAgent.value?.id || '',
        message_type: messageData.message_type || 'text',
        metadata: messageData.metadata || null
      }
    })
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
    const response = await $fetch<any>('/wp-json/tanzanite/v1/auto-reply/match', {
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
    const response = await $fetch<any>('/wp-json/tanzanite/v1/products', {
      params: {
        keyword: trimmedQuery,
        per_page: 20,
        status: 'publish'
      },
      credentials: 'include'
    })
    
    // 转换数据格式以适配前端显示
    if (response && Array.isArray(response.items)) {
      searchResults.value = response.items.map((item: any) => ({
        id: item.id,
        title: item.title,
        url: item.preview_url || `/product/${item.slug || item.id}`,
        thumbnail: item.thumbnail,
        price: item.prices?.sale > 0 
          ? `$${item.prices.sale}` 
          : (item.prices?.regular > 0 ? `$${item.prices.regular}` : '')
      }))
      console.log('[WhatsAppChatModal] products loaded:', searchResults.value.length)
    } else {
      searchResults.value = []
      console.log('[WhatsAppChatModal] products response empty or invalid')
    }
  } catch (error) {
    console.error('搜索失败:', error)
    productDrawerError.value = 'Search failed, please try again.'
    searchResults.value = []
  } finally {
    isSearching.value = false
    console.log('[WhatsAppChatModal] search finished')
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
    const response = await $fetch<any>('/wp-json/mytheme-vue/v1/my-orders', {
      params: { limit: 10 },
      credentials: 'include'
    })
    ordersList.value = Array.isArray(response) ? response : []
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
            agents.value = data.agents
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
    const response = await $fetch<any>('/wp-json/tanzanite/v1/customer-service/agents')
    if (response.success && response.data) {
      agents.value = response.data
      
      // 保存全局邮箱设置
      if (response.emailSettings) {
        emailSettings.value = response.emailSettings
      }
      
      // 3. 保存到 localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('whatsapp_agents_cache', JSON.stringify({
          data: {
            agents: response.data,
            emailSettings: response.emailSettings
          },
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
    const response = await $fetch<any>('/wp-json/tanzanite/v1/auto-reply/welcome', {
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
    const response = await fetch(`${config.public.apiBase}/wp-json/tanzanite/v1/agent/conversations/${conversationId.value}/transfer`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        to_agent_id: transferToAgent.value,
        note: transferNote.value,
      }),
    })
    
    const data = await response.json()
    
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

// 组件挂载时获取客服列表
onMounted(async () => {
  await fetchAgents()
  scrollToBottom()
})
</script>

<style scoped>
/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 滑入动画 - FAQ 从底部滑上来 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}

/* 渐变边框按钮 */
.gradient-border-btn {
  background: linear-gradient(black, black) padding-box,
              linear-gradient(to right, #40ffaa, #6b73ff) border-box;
  border: 2px solid transparent;
}

/* 渐变文字 */
.gradient-text {
  background: linear-gradient(to right, #40ffaa, #6b73ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 自定义滚动条 */
.overflow-y-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
