<template>
  <Teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        v-if="conversation"
        class="fixed inset-0 bg-black z-[9999] flex items-center justify-center p-4"
        @click.self="handleClose"
      >
        <!-- 聊天窗口 - 三栏布局 -->
        <Transition name="slide-up">
          <div
            v-if="conversation"
            class="border border-[#6e6ee9] rounded-2xl max-w-[1400px] w-full h-[90vh] md:h-[700px] max-h-[85vh] overflow-hidden shadow-2xl flex flex-row transition-colors duration-300"
            :style="{ backgroundColor: selectedAgent ? getAgentBgColorValue(selectedAgent.id) : '#000000' }"
          >
            <!-- 左侧：客服列表(窄栏 200px) - 移动端隐藏 -->
            <div class="hidden md:flex w-[200px] min-w-[200px] max-w-[200px] border-r border-white/10 flex-col" style="background-color: rgba(0, 0, 0, 0.5) !important;">
              <!-- 客服列表标题 -->
              <div class="px-4 py-4 border-b border-white/10">
                <h3 class="font-semibold text-sm bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] bg-clip-text text-transparent">Agents</h3>
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
                </div>
                
                <div v-if="!isLoadingAgents && agents.length === 0" class="text-center text-white/50 py-8 text-sm">
                  No agents available
                </div>
              </div>
              
              <!-- 邮箱按钮区域 - 固定在底部 -->
              <div class="border-t border-white/10 p-3 space-y-2">
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
                <!-- 移动端：邮箱按钮和关闭按钮同一行 -->
                <div class="md:hidden flex items-center justify-between px-2 pt-3 gap-2">
                  <!-- 左侧：邮箱按钮 -->
                  <div class="flex gap-1.5 flex-1">
                    <!-- Pre-sales 邮箱按钮 -->
                    <a
                      :href="emailSettings.preSalesEmail ? `mailto:${emailSettings.preSalesEmail}?subject=Pre-sales Inquiry` : 'javascript:void(0)'"
                      :class="[
                        'px-3 py-1.5 rounded-full text-xs transition-all inline-flex items-center justify-center gap-1.5 flex-1',
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
                        'px-3 py-1.5 rounded-full text-xs transition-all inline-flex items-center justify-center gap-1.5 flex-1',
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
                  
                  <!-- 右侧：关闭按钮 -->
                  <button
                    @click="handleClose"
                    class="w-10 h-10 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors flex-shrink-0"
                    aria-label="Close"
                  >
                    <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                
                <!-- 移动端：客服选择按钮 - 固定高度容器 -->
                <div v-if="selectedAgent" class="md:hidden pb-2 border-b border-white/10">
                  <div class="flex flex-col gap-2">
                    <!-- 客服选择按钮 - 圆形头像样式 -->
                    <div class="flex gap-2 justify-center px-4 pt-2 pb-1">
                      <button
                        v-for="agent in agents"
                        :key="agent.id"
                        @click="selectAgent(agent)"
                        class="flex flex-col items-center gap-1.5 transition-all flex-1"
                      >
                        <!-- 圆形头像 -->
                        <div 
                          class="w-[35px] h-[35px] rounded-full flex items-center justify-center text-white font-bold text-sm shadow-lg transition-all overflow-hidden"
                          :class="selectedAgent?.id === agent.id
                            ? 'ring-4 ring-[#ff6b6b] ring-offset-2 ring-offset-black scale-110'
                            : 'ring-2 ring-white/20'"
                          :style="agent.avatar 
                            ? '' 
                            : (selectedAgent?.id === agent.id
                              ? 'background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);'
                              : 'background: linear-gradient(135deg, #40ffaa 0%, #6b73ff 100%);')"
                        >
                          <img 
                            v-if="agent.avatar" 
                            :src="agent.avatar" 
                            :alt="agent.name"
                            class="w-full h-full object-cover"
                          />
                          <span v-else>{{ agent.name.charAt(0).toUpperCase() }}</span>
                        </div>
                        <!-- 客服名称 -->
                        <span 
                          class="text-[10px] font-medium transition-colors"
                          :class="selectedAgent?.id === agent.id ? 'text-[#ff6b6b]' : 'text-white/70'"
                        >
                          {{ agent.name }}
                        </span>
                      </button>
                    </div>
                    
                    <!-- WhatsApp 按钮 -->
                    <div class="flex gap-1.5 justify-center px-2">
                      <a
                        :href="selectedAgent?.whatsapp ? `https://wa.me/${selectedAgent.whatsapp.replace('+', '')}` : '#'"
                        :target="selectedAgent?.whatsapp ? '_blank' : '_self'"
                        :class="[
                          'px-4 py-2 rounded-full text-xs transition-all flex-1 inline-flex items-center justify-center gap-1.5',
                          selectedAgent?.whatsapp 
                            ? 'bg-[#25D366] hover:bg-[#20BA5A] text-white cursor-pointer shadow-lg' 
                            : 'bg-gray-600 text-gray-400 cursor-not-allowed opacity-50'
                        ]"
                        :title="selectedAgent?.whatsapp ? 'WhatsApp' : 'No WhatsApp configured'"
                        @click.prevent="selectedAgent?.whatsapp && window.open(`https://wa.me/${selectedAgent.whatsapp.replace('+', '')}`, '_blank')"
                      >
                        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                          <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                        </svg>
                        <span>WhatsApp</span>
                      </a>
                    </div>
                  </div>
                </div>
                
                <!-- 桌面端：客服信息 + 关闭按钮 -->
                <div class="hidden md:flex items-center justify-between px-6 py-4">
                  <div class="flex items-center gap-3 flex-1 min-w-0">
                    <!-- 桌面端：当前选中的客服信息 -->
                    <div v-if="selectedAgent" class="flex items-center gap-3">
                      <div class="w-10 h-10 rounded-full bg-gradient-to-br from-[#40ffaa] to-[#6b73ff] flex items-center justify-center text-white font-semibold">
                        <img
                          v-if="selectedAgent.avatar"
                          :src="selectedAgent.avatar"
                          :alt="selectedAgent.name"
                          class="w-full h-full rounded-full object-cover"
                        />
                        <span v-else>{{ getInitials(selectedAgent.name) }}</span>
                      </div>
                      
                      <div>
                        <h2 class="text-lg font-bold text-white">
                          {{ selectedAgent.name }}
                        </h2>
                        <p class="text-xs text-white/70">
                          {{ selectedAgent.email }}
                        </p>
                      </div>
                    </div>
                    
                    <div v-else class="text-white/50 text-sm">
                      Select an agent to start chat
                    </div>
                  </div>
                  
                  <!-- WhatsApp 按钮 -->
                  <a
                    v-if="selectedAgent?.whatsapp"
                    :href="`https://wa.me/${selectedAgent.whatsapp.replace('+', '')}`"
                    target="_blank"
                    class="flex items-center gap-2 px-4 py-2 bg-[#25D366] hover:bg-[#20BA5A] text-white rounded-lg transition-colors text-sm font-medium"
                    title="通过 WhatsApp 联系此客服"
                  >
                    <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                    </svg>
                    <span>WhatsApp</span>
                  </a>
                  
                  <!-- 转接按钮 -->
                  <button
                    v-if="selectedAgent && conversation"
                    @click="showTransferModal = true"
                    class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm font-medium"
                    title="转接会话"
                  >
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"/>
                    </svg>
                    <span>转接</span>
                  </button>
                  
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

              <!-- 标签切换 -->
              <div class="flex gap-1.5 md:gap-2 justify-center py-2 md:py-3 border-b border-white/10 px-2">
              <button
                @click="activeTab = 'chat'"
                class="px-3 md:px-4 py-1.5 rounded-full text-xs md:text-sm transition-all flex-1 md:flex-none"
                :class="activeTab === 'chat' 
                  ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white' 
                  : 'bg-white/[0.08] text-white/70 border border-white hover:bg-white/[0.15]'"
              >
                Chat
              </button>
              <button
                @click="activeTab = 'share'"
                class="px-3 md:px-4 py-1.5 rounded-full text-xs md:text-sm transition-all flex-1 md:flex-none whitespace-nowrap"
                :class="activeTab === 'share' 
                  ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white' 
                  : 'bg-white/[0.08] text-white/70 border border-white hover:bg-white/[0.15]'"
              >
                <span class="hidden md:inline">Share Products</span>
                <span class="md:hidden">Products</span>
              </button>
              <button
                @click="activeTab = 'orders'"
                class="px-3 md:px-4 py-1.5 rounded-full text-xs md:text-sm transition-all flex-1 md:flex-none whitespace-nowrap"
                :class="activeTab === 'orders' 
                  ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white' 
                  : 'bg-white/[0.08] text-white/70 border border-white hover:bg-white/[0.15]'"
              >
                <span class="hidden md:inline">My Orders</span>
                <span class="md:hidden">Orders</span>
              </button>
              </div>

              <!-- 即时聊天标签 -->
              <div v-if="activeTab === 'chat'" ref="messagesContainer" class="flex-1 overflow-y-auto p-3 md:p-6 space-y-3 md:space-y-4">
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
                <!-- 卡片消息（商品/订单）-->
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
                    alt="缩略图"
                    class="w-14 h-14 object-cover rounded-lg"
                  />
                  <div class="text-sm text-white">{{ message.title || message.message }}</div>
                </a>
                
                <!-- 普通文本消息 -->
                <div
                  v-else
                  class="max-w-[70%] rounded-xl px-3 py-2 text-white shadow-lg"
                  :class="message.is_agent 
                    ? 'bg-[rgba(64,255,170,0.35)] border border-[rgba(64,255,170,0.6)]' 
                    : 'bg-[rgba(64,122,255,0.35)] border border-[rgba(64,122,255,0.6)]'"
                >
                  <!-- 发送者名称 -->
                  <div class="text-xs mb-1 opacity-70">
                    {{ message.is_agent ? 'Agent' : message.sender_name }}
                  </div>
                  
                  <!-- 消息内容 -->
                  <div class="text-sm whitespace-pre-wrap break-words">
                    {{ message.message }}
                  </div>
                  
                  <!-- 附件 -->
                  <div v-if="message.attachment_url" class="mt-2">
                    <img
                      :src="message.attachment_url"
                      alt="附件"
                      class="max-w-full rounded-lg"
                    />
                  </div>
                  
                  <!-- 时间 -->
                  <div class="text-xs mt-1 opacity-60">
                    {{ formatMessageTime(message.created_at) }}
                  </div>
                </div>
              </div>
              </div>

              <!-- 商品分享标签 -->
              <div v-if="activeTab === 'share'" class="flex-1 flex flex-col overflow-hidden">
                <!-- 搜索和商品列表区域 - 可滚动 -->
                <div class="flex-1 overflow-y-auto p-3 md:p-6">
                  <!-- 搜索框 -->
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
                      class="h-[42px] px-3 md:px-4 bg-white/[0.08] hover:bg-white/[0.15] text-white border border-white rounded-lg transition-colors disabled:opacity-50 whitespace-nowrap text-sm"
                    >
                      {{ isSearching ? 'Searching...' : 'Search' }}
                    </button>
                  </div>

                  <!-- 商品列表 -->
                  <div v-if="searchResults.length > 0" class="grid grid-cols-1 md:grid-cols-2 gap-3">
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
                  <div v-else-if="!isSearching && searchQuery" class="text-center text-white/50 py-12">
                    No products found
                  </div>
                  <div v-else-if="!isSearching" class="text-center text-white/50 py-12">
                    Search products to share in chat
                  </div>
                </div>
                
                <!-- 浏览历史组件 - 固定在底部 -->
                <div class="border-t border-white/10 p-3 md:p-4 bg-black/20">
                  <BrowsingHistoryDark @share-to-chat="handleShareProductFromHistory" />
                </div>
              </div>

              <!-- 我的订单标签 -->
              <div v-if="activeTab === 'orders'" class="flex-1 overflow-y-auto p-3 md:p-6">
              <div v-if="isLoadingOrders" class="text-center text-white/50 py-12">
                Loading orders...
              </div>
              <div v-else-if="ordersList.length > 0" class="grid grid-cols-1 md:grid-cols-2 gap-3">
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

              <!-- 输入框（仅在聊天标签显示）-->
              <div v-if="activeTab === 'chat'" class="border-t border-white/10 p-2 md:p-4">
                <form @submit.prevent="handleSendMessage" class="flex gap-1.5 md:gap-2">
                  <input
                    v-model="newMessage"
                    type="text"
                    placeholder="Type a message..."
                    class="flex-1 px-3 py-2 md:py-2.5 bg-white/[0.06] text-white border border-white rounded-lg focus:outline-none focus:border-[#6b73ff] transition-colors text-sm md:text-base"
                    :disabled="isSending"
                  />
                  
                  <!-- 图片上传按钮 -->
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
                    class="px-2.5 py-2 md:px-3 md:py-2.5 bg-white/[0.08] hover:bg-white/[0.15] text-white border border-white rounded-lg transition-colors disabled:opacity-50 flex-shrink-0"
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
                    class="px-3 py-2 md:px-6 md:py-2.5 bg-[#6b73ff] hover:bg-[#5d65e8] text-white rounded-lg transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed text-sm md:text-base flex-shrink-0"
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
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'

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

// 客服列表和选中状态
const agents = ref<any[]>([])
const selectedAgent = ref<any>(null)
const isLoadingAgents = ref(false)

// 全局邮箱设置
const emailSettings = ref({
  preSalesEmail: '',
  afterSalesEmail: ''
})

const activeTab = ref<'chat' | 'share' | 'orders'>('chat')
const newMessage = ref('')
const isSending = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)

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

// 消息列表
const messages = ref<any[]>([])

// LocalStorage 键名（包含客服ID，确保每个客服的聊天记录独立）
const STORAGE_KEY = computed(() => {
  const agentId = selectedAgent.value?.id || 'default'
  return `tz_chat_${conversationId.value}_agent_${agentId}`
})
const STORAGE_EXPIRY_DAYS = 5

// 商品搜索
const searchQuery = ref('')
const isSearching = ref(false)
const searchResults = ref<any[]>([])

// 订单列表
const isLoadingOrders = ref(false)
const ordersList = ref<any[]>([])

// 是否显示"我的订单"标签
// 访客不显示，已登录用户显示
const shouldShowOrders = computed(() => {
  return !!user.value
})

// 关闭弹窗
const handleClose = () => {
  emit('close')
}

// WhatsApp 链接
const whatsappLink = computed(() => {
  if (!props.conversation?.email) return ''
  // 可以根据 email 生成 WhatsApp 链接，或者从客服信息中获取电话号码
  return ''
})

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
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 监听消息变化，自动滚动到底部
watch(messages, () => {
  scrollToBottom()
}, { deep: true })

// 组件挂载时加载消息
onMounted(() => {
  loadMessagesFromStorage()
  scrollToBottom()
})

// 从 localStorage 加载消息
const loadMessagesFromStorage = () => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY.value)
    if (stored) {
      const data = JSON.parse(stored)
      const now = Date.now()
      const expiryTime = STORAGE_EXPIRY_DAYS * 24 * 60 * 60 * 1000
      
      // 过滤掉超过5天的消息
      const validMessages = data.messages.filter((msg: any) => {
        const msgTime = new Date(msg.created_at).getTime()
        return (now - msgTime) < expiryTime
      })
      
      messages.value = validMessages
      
      // 更新 localStorage
      if (validMessages.length !== data.messages.length) {
        saveMessagesToStorage()
      }
    }
  } catch (error) {
    console.error('加载消息失败:', error)
  }
}

// 保存消息到 localStorage
const saveMessagesToStorage = () => {
  try {
    localStorage.setItem(STORAGE_KEY.value, JSON.stringify({
      messages: messages.value,
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
        agent_id: props.conversation?.id || '',
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
  if (!newMessage.value.trim() || !props.conversation || isSending.value) {
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
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    return
  }

  isSearching.value = true
  try {
    const response = await $fetch<any>('/wp-json/tanzanite/v1/products', {
      params: {
        keyword: searchQuery.value,
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
    } else {
      searchResults.value = []
    }
  } catch (error) {
    console.error('搜索失败:', error)
    searchResults.value = []
  } finally {
    isSearching.value = false
  }
}

// 分享商品到聊天
const shareProductToChat = async (product: any) => {
  if (!props.conversation || isSending.value) return
  
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
  if (!props.conversation || isSending.value) return
  
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
  if (!props.conversation || isSending.value) return
  
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

// 获取客服列表
const fetchAgents = async () => {
  isLoadingAgents.value = true
  try {
    const response = await $fetch<any>('/wp-json/tanzanite/v1/customer-service/agents')
    if (response.success && response.data) {
      agents.value = response.data
      
      // 保存全局邮箱设置
      if (response.emailSettings) {
        emailSettings.value = response.emailSettings
      }
      
      // 默认选择第一个客服
      if (agents.value.length > 0 && !selectedAgent.value) {
        selectedAgent.value = agents.value[0]
        // 加载该客服的聊天记录
        loadMessagesFromStorage()
        // 发送欢迎语
        await sendWelcomeMessage()
      }
    }
  } catch (error) {
    console.error('获取客服列表失败:', error)
  } finally {
    isLoadingAgents.value = false
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
  selectedAgent.value = agent
  // 切换客服时加载对应的聊天记录
  loadMessagesFromStorage()
  scrollToBottom()
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

// 监听标签切换，自动加载订单
watch(activeTab, (newTab) => {
  if (newTab === 'orders' && ordersList.value.length === 0 && !isLoadingOrders.value) {
    loadOrders()
  }
})

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
onMounted(() => {
  fetchAgents()
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

/* 滑入动画 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
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
