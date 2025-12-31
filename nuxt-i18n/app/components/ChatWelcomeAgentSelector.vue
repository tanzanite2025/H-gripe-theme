<template>
  <div>
    <div v-if="displayAgents.length === 3" class="relative h-[230px]">
      <button
        v-for="agent in displayAgents"
        :key="agent.id"
        type="button"
        class="agent-row group absolute left-1/2 -translate-x-1/2 text-left rounded-full shadow-[0_10px_18px_-14px_rgba(0,0,0,1)] transition-[transform,width,background-color,filter] duration-300 ease-out will-change-transform"
        :class="getRowClass(agent)"
        @click="handleSelect(agent)"
      >
        <div class="h-full flex items-center gap-3 px-3" :class="getInnerClass(agent)">
          <span class="relative shrink-0 w-14 h-14">
<!-- ... (middle content remains same) ... -->

            <span
              class="w-full h-full rounded-full bg-white/[0.16] flex items-center justify-center text-xs font-semibold overflow-hidden"
              :class="isSelected(agent) ? 'text-black/80' : 'text-white/85'"
            >
              <template v-if="getAvatarSrc(agent)">
                <img :src="getAvatarSrc(agent)" :alt="agent.name" class="w-full h-full rounded-full" :class="getAvatarImgFitClass(agent)" />
              </template>
              <template v-else>
                {{ getInitials(agent.name) }}
              </template>
            </span>

            <span
              class="absolute -right-0.5 -bottom-0.5 w-3.5 h-3.5 rounded-full bg-emerald-400 shadow-[0_0_0_2px_rgba(52,211,153,0.25),0_0_10px_rgba(52,211,153,0.7)]"
              aria-hidden="true"
            ></span>
          </span>

          <span class="min-w-0 flex-1">
            <span class="flex items-center gap-2">
              <span
                class="text-sm font-semibold truncate"
                :class="isSelected(agent) ? 'text-black' : 'text-white/95'"
              >
                {{ agent.name }}
              </span>
              <span class="text-[11px] text-emerald-300/95">Online</span>
            </span>
            <span
              class="mt-0.5 block text-xs truncate"
              :class="isSelected(agent) ? 'text-black/70' : 'text-white/70'"
            >
              {{ getAgentDescription(agent.name) }}
            </span>
          </span>

          <span class="shrink-0" :class="isSelected(agent) ? 'text-black/70' : 'text-white/55'">→</span>
        </div>
      </button>
    </div>

    <div v-else class="space-y-3">
      <button
        v-for="agent in displayAgents"
        :key="agent.id"
        type="button"
        class="w-full text-left rounded-full bg-white/[0.10] hover:bg-white/[0.14] shadow-[0_10px_18px_-14px_rgba(0,0,0,1)] transition-colors"
        @click="handleSelect(agent)"
      >
        <div class="flex items-center gap-3 px-3 py-2">
          <span class="relative shrink-0 w-14 h-14">
            <span class="w-full h-full rounded-full bg-white/[0.16] flex items-center justify-center text-xs font-semibold overflow-hidden text-white/85">
              <template v-if="getAvatarSrc(agent)">
                <img :src="getAvatarSrc(agent)" :alt="agent.name" class="w-full h-full rounded-full" :class="getAvatarImgFitClass(agent)" />
              </template>
              <template v-else>
                {{ getInitials(agent.name) }}
              </template>
            </span>
            <span
              class="absolute -right-0.5 -bottom-0.5 w-3.5 h-3.5 rounded-full bg-emerald-400 shadow-[0_0_0_2px_rgba(52,211,153,0.25),0_0_10px_rgba(52,211,153,0.7)]"
              aria-hidden="true"
            ></span>
          </span>

          <span class="min-w-0 flex-1">
            <span class="flex items-center gap-2">
              <span class="text-sm font-semibold text-white/95 truncate">{{ agent.name }}</span>
              <span class="text-[11px] text-emerald-300/95">Online</span>
            </span>
            <span class="mt-0.5 block text-xs text-white/70 truncate">{{ getAgentDescription(agent.name) }}</span>
          </span>

          <span class="shrink-0 text-white/55">→</span>
        </div>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

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
      ? 'bg-[linear-gradient(135deg,rgba(203,213,225,0.95),rgba(148,163,184,0.95))] hover:brightness-95'
      : 'bg-white/[0.10] hover:bg-white/[0.14]',
  ]
}

const getInnerClass = (agent: any) => {
  return isSelected(agent) ? 'py-2' : 'py-1'
}

const getInitials = (name: string) => {
  const parts = String(name || '').trim().split(/\s+/).filter(Boolean)
  if (!parts.length) return '?'
  const first = parts[0]?.[0] || ''
  const second = parts.length > 1 ? parts[1]?.[0] || '' : ''
  return `${first}${second}`.toUpperCase()
}

const getAvatarSrc = (agent: any): string => {
  const name = String(agent?.name ?? '').toLowerCase()
  const id = String(agent?.id ?? '')
  const isTech = name.includes('tech') || id === 'CS002'
  if (isTech) return '/whatsappmodel/welcome/techsupport.webp'
  return String(agent?.avatar || '')
}

const getAvatarImgFitClass = (agent: any): string => {
  const name = String(agent?.name ?? '').toLowerCase()
  const id = String(agent?.id ?? '')
  const isTech = name.includes('tech') || id === 'CS002'
  if (isTech) return 'object-contain'
  return 'object-cover'
}

const getAgentDescription = (name: string) => {
  const key = String(name || '').toLowerCase()
  if (key.includes('tech')) return 'Compatibility, specs, setup, and troubleshooting'
  if (key.includes('after')) return 'Order tracking, warranty, returns, and post-purchase help'
  if (key.includes('sale')) return 'Pre-sales questions, pricing, and quotes'
  return 'Chat with our team for help and guidance'
}
</script>
