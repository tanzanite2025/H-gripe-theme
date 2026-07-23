<template>
  <Transition name="fade">
    <div
      v-if="modelValue"
      class="fixed inset-0 bg-black/50 z-[10000] flex items-center justify-center p-4"
      @click.self="emit('update:modelValue', false)"
    >
      <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-2xl">
        <h3 class="text-xl font-bold text-gray-900 mb-4">{{ t('chatModal.transfer.title') }}</h3>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('chatModal.transfer.targetAgentLabel') }}
            </label>
            <select
              :value="transferToAgent"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="emit('update:transferToAgent', ($event.target as HTMLSelectElement).value)"
            >
              <option value="">{{ t('chatModal.transfer.targetAgentPlaceholder') }}</option>
              <option
                v-for="agent in availableAgents"
                :key="agent.id"
                :value="agent.id"
              >
                {{ agent.name }} ({{ agent.email }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('chatModal.transfer.noteLabel') }}
            </label>
            <textarea
              :value="transferNote"
              rows="3"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              :placeholder="t('chatModal.transfer.notePlaceholder')"
              @input="emit('update:transferNote', ($event.target as HTMLTextAreaElement).value)"
            ></textarea>
          </div>
        </div>

        <div class="flex gap-3 mt-6">
          <button
            @click="emit('update:modelValue', false)"
            :disabled="isTransferring"
            class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
          >
            {{ t('chatModal.transfer.cancel') }}
          </button>
          <button
            @click="emit('submit')"
            :disabled="isTransferring || !transferToAgent"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ isTransferring ? t('chatModal.transfer.transferring') : t('chatModal.transfer.confirm') }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
  agents: any[]
  selectedAgent: any | null
  transferToAgent: string
  transferNote: string
  isTransferring: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'update:transferToAgent', value: string): void
  (e: 'update:transferNote', value: string): void
  (e: 'submit'): void
}>()

const availableAgents = computed(() => {
  return props.agents.filter(agent => agent.id !== props.selectedAgent?.id)
})
</script>
