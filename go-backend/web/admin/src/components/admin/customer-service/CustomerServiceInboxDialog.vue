<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent
      size="full"
      class="h-[calc(100dvh-1rem)] max-h-[calc(100dvh-1rem)] grid-rows-[auto_minmax(0,1fr)] gap-0 rounded-2xl p-0 sm:h-[min(88dvh,60rem)] sm:max-h-[calc(100dvh-2rem)]"
      @open-auto-focus.prevent
    >
      <DialogHeader class="shrink-0 border-b px-5 py-4 pr-14">
        <DialogTitle>{{ title }}</DialogTitle>
        <DialogDescription class="sr-only">{{ description }}</DialogDescription>
      </DialogHeader>

      <div class="min-h-0 overflow-hidden p-3 sm:p-4">
        <CustomerServiceInboxWorkbench
          v-if="open"
          :show-header="false"
          :show-stats="false"
        />
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import CustomerServiceInboxWorkbench from '@/components/admin/customer-service/CustomerServiceInboxWorkbench.vue'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'

interface CustomerServiceInboxDialogProps {
  open?: boolean
  title?: string
  description?: string
}

const props = withDefaults(defineProps<CustomerServiceInboxDialogProps>(), {
  open: false,
  title: '客服工作台',
  description: '搜索会话、回复客户，并查看客户上下文。',
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
}>()
</script>
