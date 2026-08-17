<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <DialogTitle>联系拒付客户</DialogTitle>
        <DialogDescription>{{ form.to || '客户邮箱缺失' }}</DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <label class="block space-y-1.5">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">TO / 收件人</span>
          <Input :model-value="form.to" disabled />
        </label>

        <label class="block space-y-1.5">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SUBJECT / 主题</span>
          <Input :model-value="form.subject" :disabled="sending" @update:model-value="emit('update:subject', String($event))" />
        </label>

        <label class="block space-y-1.5">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">BODY / 正文</span>
          <Textarea
            :model-value="form.body"
            rows="12"
            :disabled="sending"
            @update:model-value="emit('update:body', String($event))"
          />
        </label>

        <div class="flex flex-col gap-2 sm:flex-row sm:justify-end">
          <Button
            v-if="mailtoUrl"
            as="a"
            variant="outline"
            class="rounded-full font-black uppercase tracking-wider"
            :href="mailtoUrl"
          >
            <ExternalLink class="size-3.5" />
            本机邮件
          </Button>
          <Button
            class="rounded-full font-black uppercase tracking-wider"
            :disabled="sending || !form.to || !form.subject || !form.body"
            @click="emit('submit')"
          >
 <Send :class="['size-3.5', sending ? 'animate-pulse': '']" />
            {{ sending ? '发送中' : '发送邮件' }}
          </Button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ExternalLink, Send } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type { OrderDisputeEmailForm } from './orderTypes'

withDefaults(defineProps<{
  open?: boolean
  form: OrderDisputeEmailForm
  sending?: boolean
  mailtoUrl?: string
}>(), {
  open: false,
  sending: false,
  mailtoUrl: ''
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:subject', value: string): void
  (event: 'update:body', value: string): void
  (event: 'submit'): void
}>()
</script>
