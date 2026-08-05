<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="gap-0 p-0" @open-auto-focus.prevent>
      <DialogHeader class="shrink-0 border-b px-5 py-4 pr-14">
        <DialogTitle>{{ title }}</DialogTitle>
        <DialogDescription>{{ description }}</DialogDescription>
      </DialogHeader>

      <div class="h-[min(82dvh,58rem)] min-h-[34rem] overflow-hidden p-4">
        <CustomerServiceInboxWorkbench
          v-if="open"
          :show-header="false"
          :show-stats="false"
          :show-region-analytics="false"
        />
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import CustomerServiceInboxWorkbench from '@/components/admin/customer-service/CustomerServiceInboxWorkbench.vue'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'

defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '客服工作台' },
  description: { type: String, default: '搜索会话、回复客户，并查看客户上下文。' },
})

const emit = defineEmits(['update:open'])
</script>
