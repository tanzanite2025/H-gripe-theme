<template>
  <Transition name="fade-scale" mode="out-in">
    <div
      v-if="!selectedConversation"
      key="agent-list"
      class="relative border-2 border-emerald-500 rounded-2xl shadow-[0_0_30px_rgba(16,185,129,0.3)] w-[420px] max-w-[calc(100vw-2rem)] h-[85vh] max-h-[800px] overflow-hidden bg-gradient-to-b from-[#0d1117] to-black pointer-events-auto"
    >
      <div class="border-b border-white/10 bg-black/70 backdrop-blur-md px-4 py-3 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-full bg-gradient-to-br from-emerald-500 to-teal-400 flex items-center justify-center text-sm font-semibold text-black">
            {{ user?.display_name?.charAt(0) || 'A' }}
          </div>
          <div>
            <div class="text-white font-medium text-sm">{{ user?.display_name || 'Agent' }}</div>
            <div class="relative">
              <button
                type="button"
                class="text-xs flex items-center gap-1 hover:opacity-80 transition-opacity"
                :class="agentStatusColors[currentAgentStatus]?.text || 'text-gray-400'"
                @click="emit('update:showStatusDropdown', !showStatusDropdown)"
              >
                <span
                  class="w-2 h-2 rounded-full"
                  :class="[agentStatusColors[currentAgentStatus]?.dot || 'bg-gray-500', currentAgentStatus === 'online' ? 'animate-pulse' : '']"
                ></span>
                {{ agentStatusLabels[currentAgentStatus] || 'Offline' }}
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M6 9l6 6 6-6"/>
                </svg>
              </button>
              <div
                v-if="showStatusDropdown"
                class="absolute top-full left-0 mt-1 bg-black/95 border border-white/10 rounded-lg py-1 min-w-[120px] z-50 shadow-xl"
              >
                <button
                  v-for="status in ['online', 'busy', 'away', 'offline']"
                  :key="status"
                  type="button"
                  class="w-full px-3 py-1.5 text-left text-xs flex items-center gap-2 hover:bg-white/5 transition-colors"
                  :class="currentAgentStatus === status ? 'bg-white/10' : ''"
                  @click="emit('changeStatus', status)"
                >
                  <span
                    class="w-2 h-2 rounded-full"
                    :class="agentStatusColors[status]?.dot || 'bg-gray-500'"
                  ></span>
                  <span :class="agentStatusColors[status]?.text || 'text-gray-400'">
                    {{ agentStatusLabels[status] }}
                  </span>
                </button>
              </div>
            </div>
          </div>
        </div>
        <button
          type="button"
          class="w-9 h-9 rounded-full border-2 border-white/20 text-white/60 flex items-center justify-center hover:border-red-500 hover:text-red-500 transition-colors"
          @click="emit('close')"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-4">
        <div class="text-white/50 text-xs uppercase tracking-wider mb-3">Conversations</div>

        <div v-if="isLoadingConversations" class="flex items-center justify-center py-8">
          <div class="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin"></div>
        </div>

        <div v-else-if="agentConversations.length === 0" class="text-center py-8">
          <div class="text-white/30 text-4xl mb-2">💬</div>
          <div class="text-white/50 text-sm">No conversations yet</div>
        </div>

        <div v-else class="space-y-2">
          <button
            v-for="conv in agentConversations"
            :key="conv.id"
            type="button"
            class="w-full p-3 rounded-xl border border-white/10 hover:border-emerald-500/50 hover:bg-emerald-500/5 transition-all text-left"
            @click="emit('selectConversation', conv)"
          >
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-gradient-to-br from-[#6b73ff] to-[#40ffaa] flex items-center justify-center text-sm font-semibold text-black flex-shrink-0">
                {{ conv.customer_name?.charAt(0) || 'U' }}
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between">
                  <span class="text-white font-medium text-sm truncate">{{ conv.customer_name || 'Unknown' }}</span>
                  <span class="text-white/40 text-xs">{{ formatTime(conv.updated_at) }}</span>
                </div>
                <div class="text-white/50 text-xs truncate">{{ conv.last_message || 'No messages' }}</div>
              </div>
              <div v-if="conv.unread_count > 0" class="w-5 h-5 rounded-full bg-emerald-500 text-black text-xs font-bold flex items-center justify-center flex-shrink-0">
                {{ conv.unread_count > 9 ? '9+' : conv.unread_count }}
              </div>
            </div>
          </button>
        </div>
      </div>

      <div class="border-t border-white/10 p-3">
        <button
          type="button"
          class="w-full py-2 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-sm font-medium hover:bg-emerald-500/20 transition-colors"
          @click="emit('refreshConversations')"
        >
          Refresh Conversations
        </button>
      </div>
    </div>

    <div
      v-else
      key="agent-chat"
      class="relative border-2 border-emerald-500 rounded-2xl shadow-[0_0_30px_rgba(16,185,129,0.3)] w-[420px] max-w-[calc(100vw-2rem)] h-[85vh] max-h-[800px] overflow-hidden flex flex-col bg-black pointer-events-auto"
    >
      <div class="border-b border-white/10 bg-black/70 backdrop-blur-md px-4 py-3 flex items-center gap-3">
        <button
          type="button"
          class="w-9 h-9 rounded-full border border-white/20 text-white/60 flex items-center justify-center hover:border-white/40 hover:text-white transition-colors"
          @click="emit('backToConversationList')"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M15 18l-6-6 6-6"/>
          </svg>
        </button>
        <div
          class="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center text-xs font-semibold text-white overflow-hidden shadow-[0_0_12px_rgba(15,23,42,0.95)] flex-shrink-0"
        >
          {{ selectedConversation.customer_name?.charAt(0) || 'U' }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-white font-medium text-sm truncate">{{ selectedConversation.customer_name || 'Customer' }}</div>
          <div class="text-white/50 text-xs truncate">{{ selectedConversation.customer_email || '' }}</div>
        </div>
        <button
          type="button"
          class="w-9 h-9 rounded-full border-2 border-white/20 text-white/60 flex items-center justify-center hover:border-red-500 hover:text-red-500 transition-colors"
          @click="emit('close')"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <div ref="messagesContainer" class="flex-1 overflow-y-auto p-4 space-y-3">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="flex"
          :class="msg.sender_type === 'agent' ? 'justify-end' : 'justify-start'"
        >
          <div
            class="max-w-[80%] px-4 py-2 rounded-2xl text-sm"
            :class="msg.sender_type === 'agent' ? 'bg-emerald-500 text-black' : 'bg-white/10 text-white'"
          >
            {{ msg.message }}
          </div>
        </div>
      </div>

      <div class="border-t border-white/10 p-3">
        <div class="flex gap-2">
          <input
            :value="newMessage"
            type="text"
            placeholder="Type a message..."
            class="flex-1 px-4 py-2 rounded-full text-white text-sm placeholder-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
            @input="emit('update:newMessage', ($event.target as HTMLInputElement).value)"
            @keyup.enter="emit('sendMessage')"
          />
          <button
            type="button"
            class="w-10 h-10 rounded-full bg-emerald-500 text-black flex items-center justify-center hover:bg-emerald-400 transition-colors disabled:opacity-50"
            :disabled="!newMessage.trim() || isSending"
            @click="emit('sendMessage')"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'

type AgentStatusClasses = Record<string, { dot: string; text: string }>

const props = defineProps<{
  user: any
  selectedConversation: any | null
  isLoadingConversations: boolean
  agentConversations: any[]
  currentAgentStatus: string
  showStatusDropdown: boolean
  agentStatusColors: AgentStatusClasses
  agentStatusLabels: Record<string, string>
  messages: any[]
  newMessage: string
  isSending: boolean
}>()

const emit = defineEmits<{
  (e: 'update:showStatusDropdown', value: boolean): void
  (e: 'update:newMessage', value: string): void
  (e: 'close'): void
  (e: 'changeStatus', status: string): void
  (e: 'selectConversation', conversation: any): void
  (e: 'refreshConversations'): void
  (e: 'backToConversationList'): void
  (e: 'sendMessage'): void
}>()

const messagesContainer = ref<HTMLElement | null>(null)

watch(() => [props.messages.length, props.selectedConversation?.id], async () => {
  await nextTick()
  const container = messagesContainer.value
  if (container) {
    container.scrollTop = container.scrollHeight
  }
})

const formatTime = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60 * 1000) return 'Just now'
  if (diff < 60 * 60 * 1000) return `${Math.floor(diff / 60000)}m`
  if (diff < 24 * 60 * 60 * 1000) return `${Math.floor(diff / 3600000)}h`
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>
