<template>
  <div
    class="sidebar-panel chat-modal-shell relative w-full md:w-[420px] max-w-full md:max-w-[calc(100vw-2rem)] h-[95vh] md:h-[85vh] max-h-[800px] rounded-2xl overflow-hidden flex flex-col border-2 border-[#6b73ff]/40 ring-1 ring-white/10 bg-slate-950/80 backdrop-blur-xl shadow-[0_0_30px_rgba(107,115,255,0.6)] pointer-events-auto"
  >
    <div class="absolute inset-x-0 top-0 h-[200px] bg-gradient-to-br from-indigo-600/20 to-teal-600/20 blur-3xl pointer-events-none z-0"></div>
    <button
      type="button"
      class="absolute top-4 right-4 z-20 w-9 h-9 rounded-full border-2 border-white/20 bg-black/50 backdrop-blur-sm tz-text-secondary flex items-center justify-center hover:border-red-500 hover:text-red-500 transition-colors pointer-events-auto"
      :aria-label="t('chatModal.actions.close')"
      @click="emit('close')"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M18 6L6 18M6 6l12 12"/>
      </svg>
    </button>

    <div class="flex-1 overflow-y-auto p-3 md:p-4 relative z-10 welcome-scroll-area">
      <div class="w-full">
        <div class="mb-2 md:mb-3 welcome-header">
          <div class="flex items-center gap-3 mb-2 welcome-logo-row">
            <img
              src="/images/chat-logo.webp"
              :alt="t('chatModal.welcome.logoAlt')"
              class="w-12 h-12 rounded-xl object-cover shrink-0 welcome-logo"
            />

            <h1 class="text-2xl md:text-3xl font-bold text-white welcome-title">
              {{ t('chatModal.welcome.title') }} <span class="inline-block animate-wave">👋</span>
            </h1>
          </div>

          <p class="text-sm md:text-base tz-text-secondary leading-relaxed welcome-desc">
            {{ t('chatModal.welcome.description') }}
          </p>
        </div>

        <div class="space-y-2 mb-0">
          <div class="flex items-center gap-2 mb-2 pl-1">
            <div class="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.8)] animate-pulse"></div>
            <div class="text-xs font-bold text-emerald-400 uppercase tracking-wider">
              {{ t('chatModal.welcome.onlineTeam') }}
            </div>
          </div>

          <ChatWelcomeAgentSelector
            :agents="welcomeAgents"
            :selected-agent="selectedAgent"
            @select="emit('selectAgent', $event)"
          />

          <p class="text-[10px] tz-text-muted leading-relaxed text-center px-2">
            {{ t('chatModal.welcome.onlineSummary', { count: onlineAgentsCount, agentLabel: onlineAgentsLabel }) }}
          </p>
        </div>
      </div>
    </div>

    <div class="p-2 md:px-5 md:pb-3 shrink-0 z-20 bg-white/[0.02] border-t border-white/[0.08]">
      <ChatStartButton
        class="w-full text-sm"
        :label="startButtonLabel"
        :disabled="!selectedAgent"
        @click="emit('enterChat')"
      />

      <div class="flex gap-2.5 mt-2">
        <a
          v-if="selectedAgent?.whatsapp"
          :href="`https://wa.me/${selectedAgent.whatsapp.replace('+', '')}`"
          target="_blank"
          class="flex-1 py-2.5 rounded-full bg-[#25D366] text-white text-sm font-medium flex items-center justify-center gap-1.5 shadow-[0_4px_12px_rgba(0,0,0,0.9)] hover:-translate-y-0.5 transition-transform"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
            <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347"/>
          </svg>
          {{ t('chatModal.actions.whatsapp') }}
        </a>
        <a
          v-if="emailSettings.preSalesEmail"
          :href="`mailto:${emailSettings.preSalesEmail}`"
          class="flex-1 py-2.5 rounded-full bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white text-sm font-medium flex items-center justify-center gap-1.5 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] hover:-translate-y-0.5 transition-transform"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
          </svg>
          {{ t('chatModal.actions.email') }}
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const props = defineProps<{
  welcomeAgents: any[]
  selectedAgent: any | null
  onlineAgentsCount: number
  hasHistoryChat: boolean
  emailSettings: {
    preSalesEmail: string
    afterSalesEmail: string
  }
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'selectAgent', agent: any): void
  (e: 'enterChat'): void
}>()

const { t } = useI18n()

const onlineAgentsLabel = computed(() => {
  return props.onlineAgentsCount === 1 ? t('chatModal.welcome.agentSingular') : t('chatModal.welcome.agentPlural')
})

const startButtonLabel = computed(() => {
  return props.hasHistoryChat ? t('chatModal.welcome.continueCta') : t('chatModal.welcome.startCta')
})
</script>
