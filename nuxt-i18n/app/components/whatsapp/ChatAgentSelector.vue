<template>
  <div>
    <div v-if="displayAgents.length === 3" class="relative h-[230px]">
      <button
        v-for="agentEntry in displayAgentEntries"
        :key="agentEntry.agent.id"
        type="button"
        class="agent-row group absolute left-1/2 -translate-x-1/2 text-left rounded-full shadow-md transition-[transform,width,background-color,filter] duration-300 ease-out will-change-transform"
        :class="getRowClass(agentEntry.agent)"
        @click="handleSelect(agentEntry.agent)"
      >
        <div class="h-full flex items-center gap-3 px-3" :class="getInnerClass(agentEntry.agent)">
          <span class="relative shrink-0 w-14 h-14">
            <span
              class="w-full h-full rounded-full tz-surface-muted flex items-center justify-center text-xs font-semibold overflow-hidden"
                :class="isSelected(agentEntry.agent) ? 'text-white/80' : 'tz-text-primary'"
            >
              <template v-if="getAvatarSrc(agentEntry.agent)">
                <StorefrontImage :src="getAvatarSrc(agentEntry.agent)" :alt="agentEntry.agent.name" class="w-full h-full rounded-full object-cover" preset="avatar" />
              </template>
              <template v-else>
                {{ agentEntry.presentation.initials }}
              </template>
            </span>

            <span
              class="absolute -right-0.5 -bottom-0.5 w-3.5 h-3.5 rounded-full"
              :class="getStatusDotClass(agentEntry.agent)"
              :title="getStatusLabel(agentEntry.agent)"
              aria-hidden="true"
            ></span>
          </span>

          <span class="min-w-0 flex-1">
            <span class="flex min-w-0 items-center gap-2">
              <span
                class="text-sm font-semibold truncate"
                :class="isSelected(agentEntry.agent) ? 'text-white' : 'tz-text-primary'"
              >
                {{ agentEntry.agent.name }}
              </span>
              <span
                class="inline-flex max-w-[8.5rem] shrink-0 items-center rounded-full border px-2 py-0.5 text-[10px] font-semibold leading-none truncate"
                :class="getGroupBadgeClass(agentEntry.agent)"
                :title="agentEntry.presentation.groupLabel"
              >
                {{ agentEntry.presentation.groupLabel }}
              </span>
            </span>
            <span
              class="mt-0.5 block text-xs truncate"
              :class="isSelected(agentEntry.agent) ? 'text-white/70' : 'tz-text-secondary'"
            >
              {{ agentEntry.presentation.contactLabel || t('chatModal.agentSelector.descriptions.default') }}
            </span>
          </span>

          <span class="shrink-0" :class="isSelected(agentEntry.agent) ? 'text-white/70' : 'tz-text-muted'">→</span>
        </div>
      </button>
    </div>

    <div v-else class="space-y-3">
      <button
        v-for="agentEntry in displayAgentEntries"
        :key="agentEntry.agent.id"
        type="button"
        class="w-full text-left rounded-full tz-surface-subtle hover:tz-surface-muted shadow-[0_10px_18px_-14px_rgba(20,32,43,0.14)] transition-colors"
        @click="handleSelect(agentEntry.agent)"
      >
        <div class="flex items-center gap-3 px-3 py-2">
          <span class="relative shrink-0 w-14 h-14">
            <span class="w-full h-full rounded-full tz-surface-muted flex items-center justify-center text-xs font-semibold overflow-hidden tz-text-primary">
              <template v-if="getAvatarSrc(agentEntry.agent)">
                <StorefrontImage :src="getAvatarSrc(agentEntry.agent)" :alt="agentEntry.agent.name" class="w-full h-full rounded-full object-cover" preset="avatar" />
              </template>
              <template v-else>
                {{ agentEntry.presentation.initials }}
              </template>
            </span>
            <span
              class="absolute -right-0.5 -bottom-0.5 w-3.5 h-3.5 rounded-full"
              :class="getStatusDotClass(agentEntry.agent)"
              :title="getStatusLabel(agentEntry.agent)"
              aria-hidden="true"
            ></span>
          </span>

          <span class="min-w-0 flex-1">
            <span class="flex min-w-0 items-center gap-2">
              <span class="text-sm font-semibold tz-text-primary truncate">{{ agentEntry.agent.name }}</span>
              <span
                class="inline-flex max-w-[8.5rem] shrink-0 items-center rounded-full border border-border/60 bg-white/70 px-2 py-0.5 text-[10px] font-semibold leading-none tz-text-secondary truncate"
                :title="agentEntry.presentation.groupLabel"
              >
                {{ agentEntry.presentation.groupLabel }}
              </span>
            </span>
            <span class="mt-0.5 block text-xs tz-text-secondary truncate">{{ agentEntry.presentation.contactLabel || t('chatModal.agentSelector.descriptions.default') }}</span>
          </span>

          <span class="shrink-0 tz-text-muted">→</span>
        </div>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import { buildChatAgentPresentationList } from '~/lib/chatAgentPresentation'

const { t } = useI18n()

const props = defineProps<{
  agents: any[]
  selectedAgent: any | null
}>()

const emit = defineEmits<{
  (e: 'select', agent: any): void
}>()

const displayAgents = computed(() => {
  return Array.isArray(props.agents) ? props.agents.slice(0, 3) : []
})

const displayAgentEntries = computed(() => buildChatAgentPresentationList(displayAgents.value))

const selectedId = computed(() => {
  const ids = displayAgents.value.map(agent => String(agent?.id ?? ''))
  const current = props.selectedAgent?.id != null ? String(props.selectedAgent.id) : ''
  if (current && ids.includes(current)) return current
  return ids[1] || ids[0] || ''
})

const getOrderIds = computed(() => {
  const base = displayAgents.value.map(agent => String(agent?.id ?? ''))
  const sel = selectedId.value
  const others = base.filter(id => id !== sel)
  return [others[0], sel, others[1]]
})

const getSlotIndex = (agent: any) => {
  const id = String(agent?.id ?? '')
  return getOrderIds.value.indexOf(id)
}

const isSelected = (agent: any) => {
  return String(agent?.id ?? '') === selectedId.value
}

const handleSelect = (agent: any) => {
  emit('select', agent)
}

const translateClasses = ['translate-y-0', 'translate-y-[76px]', 'translate-y-[156px]'] as const

const getRowClass = (agent: any) => {
  const selected = isSelected(agent)
  const slotIndex = getSlotIndex(agent)
  const translate = translateClasses[Math.max(0, slotIndex)] || 'translate-y-0'

  return [
    translate,
    selected ? 'w-full h-[76px]' : 'w-[92%] h-[66px]',
    selected
       ? 'tz-surface-muted tz-text-primary hover:tz-surface-subtle'
      : 'tz-surface-subtle hover:tz-surface-muted',
  ]
}

const getInnerClass = (agent: any) => {
  return isSelected(agent) ? 'py-2' : 'py-1'
}

const getAvatarSrc = (agent: any): string => {
  return String(agent?.avatar || '').trim()
}

const getAgentStatus = (agent: any): string => {
  const status = String(agent?.online_status ?? agent?.status ?? '').trim().toLowerCase()
  return ['online', 'busy', 'away', 'offline'].includes(status) ? status : 'offline'
}

const getStatusDotClass = (agent: any): string => {
  switch (getAgentStatus(agent)) {
    case 'online':
      return 'bg-[#059669] shadow-[0_0_0_2px_rgba(5, 150, 105,0.25),0_0_10px_rgba(5, 150, 105,0.7)]'
    case 'busy':
      return 'bg-amber-300 shadow-[0_0_0_2px_rgba(252,211,77,0.22),0_0_10px_rgba(252,211,77,0.5)]'
    case 'away':
      return 'bg-yellow-500 shadow-[0_0_0_2px_rgba(234,179,8,0.2),0_0_10px_rgba(234,179,8,0.42)]'
    default:
      return 'bg-slate-400 shadow-[0_0_0_2px_rgba(100,116,139,0.18)]'
  }
}

const getStatusLabel = (agent: any): string => {
  return t(`chatModal.agentPanel.status.${getAgentStatus(agent)}`)
}

const getGroupBadgeClass = (agent: any) => {
  return isSelected(agent)
    ? 'border-white/20 bg-white/10 text-white/90'
    : 'border-border/60 bg-white/70 tz-text-secondary'
}
</script>

