<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="sm" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>分配工单</DialogTitle>
          <DialogDescription>{{ currentTicket?.ticket_number }} · {{ currentTicket?.subject }}</DialogDescription>
        </DialogHeader>
        <label class="block space-y-1.5">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">负责人</span>
          <Select :model-value="assignTo" @update:model-value="(value) => emit('update:assignTo', String(value))">
            <SelectTrigger class="w-full"><SelectValue placeholder="请选择负责人" /></SelectTrigger>
            <SelectContent>
              <SelectItem v-for="user in supportUsers" :key="user.id" :value="String(user.id)">
                {{ supportUserName(user) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>
        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="assigning || !assignTo">
            <LoaderCircle v-if="assigning" class="size-4 animate-spin" />
            确认分配
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { LoaderCircle } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { TicketRecord, TicketUser } from './ticketTypes'

withDefaults(defineProps<{
  open?: boolean
  currentTicket?: TicketRecord | null
  assignTo?: string
  assigning?: boolean
  supportUsers?: TicketUser[]
  supportUserName: (user: TicketUser) => string
}>(), {
  open: false,
  currentTicket: null,
  assignTo: '',
  assigning: false,
  supportUsers: () => []
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:assignTo', value: string): void
  (event: 'submit'): void
}>()
</script>
