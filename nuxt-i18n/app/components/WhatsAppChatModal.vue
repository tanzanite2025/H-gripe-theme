<template>
  <Teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        v-if="conversation"
        class="fixed inset-0 z-[9000] flex items-center justify-end p-4 md:p-6 pointer-events-none"
      >
        <!-- 聊天窗口容器 - 右下角定位 -->
        
        <!-- 欢迎页 / 聊天窗口 切换容器 -->
        <Transition name="fade-scale" mode="out-in">
          <!-- 欢迎页 -->
          <div
            v-if="showWelcomeScreen"
            key="welcome"
            class="relative border-2 border-[#6b73ff] rounded-2xl shadow-[0_0_30px_rgba(107,115,255,0.3)] w-[420px] max-w-[calc(100vw-2rem)] h-[85vh] max-h-[800px] overflow-hidden bg-gradient-to-b from-[#0d1117] to-black pointer-events-auto"
          >
            <!-- 关闭按钮 - 固定右上角 -->
            <button
              type="button"
              class="absolute top-4 right-4 z-10 w-9 h-9 rounded-full border-2 border-white/20 bg-black/50 backdrop-blur-sm text-white/60 flex items-center justify-center hover:border-red-500 hover:text-red-500 transition-colors"
              @click="handleClose"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 6L6 18M6 6l12 12"/>
              </svg>
            </button>
            
            <!-- 可滚动内容区域 -->
            <div class="h-full overflow-y-auto p-6 md:p-8">
              <div class="w-full">
                <!-- Logo -->
                <div class="mb-4">
                  <img src="/images/chat-logo.webp" alt="Tanzanite" class="w-12 h-12 rounded-xl object-cover" />
                </div>

              <!-- 欢迎语 -->
              <div class="mb-5">
                <h1 class="text-2xl md:text-3xl font-bold text-white mb-2">
                  Hi there! <span class="inline-block animate-wave">👋</span>
                </h1>
                <p class="text-sm md:text-base text-white/70 leading-relaxed">
                  Welcome to Tanzanite. We're here to help you find the perfect carbon wheels for your ride.
                </p>
              </div>

              <!-- 客服状态卡片 -->
              <div class="bg-white/[0.03] border border-white/[0.08] rounded-2xl p-4 mb-4">
                <div class="flex items-center gap-2 mb-4">
                  <div class="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse"></div>
                  <span class="text-sm text-white/80">
                    <strong class="text-emerald-500">{{ onlineAgentsCount }} agent{{ onlineAgentsCount > 1 ? 's' : '' }}</strong> online now
                  </span>
                </div>

                <!-- 客服头像列表 -->
                <div class="flex gap-3 flex-wrap">
                  <button
                    v-for="agent in agents"
                    :key="agent.id"
                    type="button"
                    class="flex flex-col items-center gap-2 p-3 rounded-xl border-2 transition-all flex-1 min-w-[80px]"
                    :class="selectedAgent?.id === agent.id ? 'bg-[#6b73ff]/15 border-[#6b73ff]' : 'border-transparent hover:bg-[#6b73ff]/10 hover:border-[#6b73ff]/30'"
                    @click="selectAgentFromWelcome(agent)"
                  >
                    <div class="relative">
                      <div class="w-12 h-12 rounded-full bg-gradient-to-br from-[#6b73ff] to-[#40ffaa] flex items-center justify-center text-base font-semibold text-black overflow-hidden">
                        <img v-if="agent.avatar" :src="agent.avatar" :alt="agent.name" class="w-full h-full object-cover" />
                        <span v-else>{{ getInitials(agent.name) }}</span>
                      </div>
                      <div class="absolute bottom-0.5 right-0.5 w-3 h-3 rounded-full bg-emerald-500 border-2 border-[#0d1117]"></div>
                    </div>
                    <span class="text-xs text-white/80 text-center max-w-[70px] truncate">{{ agent.name }}</span>
                  </button>
                </div>
              </div>

              <!-- 开始对话按钮 -->
              <button
                type="button"
                class="w-full py-4 rounded-xl bg-gradient-to-r from-[#6b73ff] to-[#40ffaa] text-black text-base font-semibold flex items-center justify-center gap-2 hover:shadow-[0_8px_24px_rgba(107,115,255,0.4)] hover:-translate-y-0.5 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
                :disabled="!selectedAgent"
                @click="enterChat"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>
                </svg>
                Start Conversation
              </button>

              <!-- 快捷联系 -->
              <div class="flex gap-2.5 mt-4">
                <a
                  v-if="selectedAgent?.whatsapp"
                  :href="`https://wa.me/${selectedAgent.whatsapp.replace('+', '')}`"
                  target="_blank"
                  class="flex-1 py-3 rounded-xl bg-[#25D366] text-white text-sm font-medium flex items-center justify-center gap-1.5 hover:-translate-y-0.5 transition-transform"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347"/>
                  </svg>
                  WhatsApp
                </a>
                <a
                  v-if="emailSettings.preSalesEmail"
                  :href="`mailto:${emailSettings.preSalesEmail}`"
                  class="flex-1 py-3 rounded-xl bg-white/10 border border-white/15 text-white/80 text-sm font-medium flex items-center justify-center gap-1.5 hover:-translate-y-0.5 transition-transform"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                  </svg>
                  Email
                </a>
              </div>
              </div>
            </div>
          </div>

          <!-- 聊天窗口 - 简化布局 -->
          <div
            v-else
            key="chat"
            class="relative border-2 border-[#6b73ff] rounded-2xl shadow-[0_0_30px_rgba(107,115,255,0.3)] w-[420px] max-w-[calc(100vw-2rem)] h-[85vh] max-h-[800px] overflow-hidden flex flex-col transition-colors duration-300 bg-black pointer-events-auto"
          >
            <!-- 聊天区域 -->
            <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
              <!-- 头部 - 当前客服信息 -->
              <div class="border-b border-white/10 bg-black/70 backdrop-blur-md">
                <div class="px-4 py-3 flex items-center gap-3">
                  <!-- 返回欢迎页按钮 -->
                  <button
                    type="button"
                    class="w-9 h-9 rounded-full border border-white/20 text-white/60 flex items-center justify-center hover:border-white/40 hover:text-white transition-colors"
                    @click="showWelcomeScreen = true"
                    title="Back to welcome"
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M15 18l-6-6 6-6"/>
                    </svg>
                  </button>
                  
                  <!-- 当前客服头像 -->
                  <div class="w-10 h-10 rounded-full bg-gradient-to-br from-[#6b73ff] to-[#40ffaa] flex items-center justify-center text-sm font-semibold text-black overflow-hidden flex-shrink-0">
                    <img v-if="selectedAgent?.avatar" :src="selectedAgent.avatar" :alt="selectedAgent.name" class="w-full h-full object-cover" />
                    <span v-else>{{ selectedAgent ? getInitials(selectedAgent.name) : '?' }}</span>
                  </div>
                  
                  <!-- 客服信息 -->
                  <div class="flex-1 min-w-0">
                    <div class="text-white font-medium text-sm truncate">{{ selectedAgent?.name || 'Agent' }}</div>
                    <div class="text-white/50 text-xs truncate">{{ selectedAgent?.email }}</div>
                  </div>
                  
                  <!-- WhatsApp 按钮 -->
                  <a
                    v-if="selectedAgent?.whatsapp"
                    :href="`https://wa.me/${selectedAgent.whatsapp.replace('+', '')}`"
                    target="_blank"
                    class="w-9 h-9 rounded-full bg-[#25D366] text-white flex items-center justify-center hover:bg-[#20BA5A] transition-colors"
                    title="WhatsApp"
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347"/>
                    </svg>
                  </a>
                  
                  <!-- 关闭按钮 -->
                  <button
                    type="button"
                    class="w-9 h-9 rounded-full border-2 border-white/20 text-white/60 flex items-center justify-center hover:border-red-500 hover:text-red-500 transition-colors"
                    @click="handleClose"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M18 6L6 18M6 6l12 12"/>
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

                      <!-- Mobile actions: History / Cart / Wishlist -->
                      <div class="flex gap-1.5 mb-3">
                        <button
                          type="button"
                          @click="historyDrawerVisible = true"
                          class="flex-1 h-10 rounded-full text-[11px] font-semibold tracking-wide border bg-black/40 flex items-center justify-center gap-1.5"
                          :style="{ borderColor: currentThemeColor, color: currentThemeColor }"
                        >
                          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <circle cx="12" cy="12" r="8" stroke-width="1.7" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12 8v4l2.5 2.5" />
                          </svg>
                          <span>History</span>
                        </button>
                        <button
                          type="button"
                          @click="openCart"
                          class="flex-1 h-10 rounded-full text-[11px] font-semibold tracking-wide border bg-black/40 flex items-center justify-center gap-1.5"
                          :style="{ borderColor: currentThemeColor, color: currentThemeColor }"
                        >
                          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M3 3h2l2 13h12l2-9H6" />
                            <circle cx="9" cy="19" r="1.4" />
                            <circle cx="17" cy="19" r="1.4" />
                          </svg>
                          <span>Cart</span>
                        </button>
                        <button
                          type="button"
                          @click="wishlistDrawerVisible = true"
                          class="flex-1 h-10 rounded-full text-[11px] font-semibold tracking-wide border bg-black/40 flex items-center justify-center gap-1.5"
                          :style="{ borderColor: currentThemeColor, color: currentThemeColor }"
                        >
                          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z" />
                          </svg>
                          <span>Wishlist</span>
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
                    <div class="flex gap-2 mb-3 items-center">
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

                    <div
                      v-if="!productDrawerVisible && !isSearching && searchResults.length === 0"
                      class="text-center text-white/50 text-sm mb-4"
                    >
                      {{ searchQuery ? 'No products found' : 'Search products to share in chat' }}
                    </div>

                    <div class="flex justify-center gap-3 mb-4">
                      <button
                        type="button"
                        @click="historyDrawerVisible = true"
                        class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full text-sm bg-white/[0.08] text-white/80 border border-white hover:bg-white/[0.15] transition-colors"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <circle cx="12" cy="12" r="8" stroke-width="1.7" />
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12 8v4l2.5 2.5" />
                        </svg>
                        <span>History</span>
                      </button>
                      <button
                        type="button"
                        @click="openCart"
                        class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full text-sm bg-white/[0.08] text-white/80 border border-white hover:bg-white/[0.15] transition-colors"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M3 3h2l2 13h12l2-9H6" />
                          <circle cx="9" cy="19" r="1.4" />
                          <circle cx="17" cy="19" r="1.4" />
                        </svg>
                        <span>Cart</span>
                      </button>
                      <button
                        type="button"
                        @click="wishlistDrawerVisible = true"
                        class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full text-sm bg-white/[0.08] text-white/80 border border-white hover:bg-white/[0.15] transition-colors"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z" />
                        </svg>
                        <span>Wishlist</span>
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

    <WishlistDrawer
      v-model="wishlistDrawerVisible"
      @share-to-chat="handleShareProductFromHistory"
    />

    <Transition name="slide-up">
      <div
        v-if="historyDrawerVisible"
        class="fixed inset-0 z-[10001] flex items-end justify-center p-0 md:p-4 pointer-events-none"
        @click.self="handleHistoryDrawerClose"
      >
        <div
          class="pointer-events-auto w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]
                 rounded-2xl border-2 border-[#6b73ff] bg-black shadow-[0_0_30px_rgba(107,115,255,0.6)]
                 flex flex-col overflow-hidden"
        >
          <div class="flex items-center justify-between px-4 py-3 border-b border-white/10">
            <div class="text-sm font-semibold text-white/90">
              Browsing history
            </div>
            <button
              type="button"
              class="w-8 h-8 rounded-full border border-white/40 text-white flex items-center justify-center hover:bg-white/10 transition-colors"
              @click="handleHistoryDrawerClose"
            >
              <span class="text-lg leading-none">x</span>
            </button>
          </div>

          <div class="flex-1 min-h-0 overflow-y-auto p-4 md:p-6">
            <BrowsingHistoryDark @share-to-chat="handleShareProductFromHistory" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'

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
const { openCart } = useCart()
const config = useRuntimeConfig()

// Desktop-only搜索占位
const desktopSearchQuery = ref('')

// 欢迎页状态
const showWelcomeScreen = ref(true)

// 客服列表和选中状态
const agents = ref<any[]>([])
const selectedAgent = ref<any>(null)
const isLoadingAgents = ref(false)

// 在线客服数量
const onlineAgentsCount = computed(() => agents.value.length)

const isDesktopSearchFocused = ref(false)

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
        url: item.preview_url || `/shop/${item.slug || item.id}`,
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
      const response = await $fetch<any>('/wp-json/tanzanite/v1/customer-service/agents')
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

/* 挥手动画 */
@keyframes wave {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(20deg); }
  75% { transform: rotate(-10deg); }
}

.animate-wave {
  animation: wave 1.5s ease-in-out infinite;
}

/* 欢迎页/聊天窗口切换动画 */
.fade-scale-enter-active,
.fade-scale-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-scale-enter-from {
  opacity: 0;
  transform: scale(0.95);
}

.fade-scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
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
