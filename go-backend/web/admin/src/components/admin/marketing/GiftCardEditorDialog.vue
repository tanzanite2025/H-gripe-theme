<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-2xl" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>创建礼品卡</DialogTitle>
          <DialogDescription>发行礼品卡并设置收件人与到期信息。</DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="卡号" required :error="errors.code">
            <Input v-model="form.code" class="font-mono uppercase" placeholder="请输入卡号" @input="emit('clear-error', 'code')" />
          </AdminFormField>
          <AdminFormField label="币种">
            <Input v-model="form.currency" class="uppercase" placeholder="USD" />
          </AdminFormField>
          <AdminFormField label="初始金额" required :error="errors.initial_value">
            <Input v-model.number="form.initial_value" type="number" min="0" step="0.01" @input="emit('clear-error', 'initial_value')" />
          </AdminFormField>
          <AdminFormField label="到期时间">
            <Input v-model="form.expires_at" type="datetime-local" />
          </AdminFormField>
          <AdminFormField label="收件人邮箱">
            <Input v-model="form.recipient_email" type="email" placeholder="name@example.com" />
          </AdminFormField>
          <AdminFormField label="收件人姓名">
            <Input v-model="form.recipient_name" />
          </AdminFormField>
          <AdminFormField label="发送人姓名">
            <Input v-model="form.sender_name" />
          </AdminFormField>
          <AdminFormField label="封面图片">
            <Input v-model="form.cover_image" type="url" placeholder="https://..." />
          </AdminFormField>
          <AdminFormField label="祝福语" class="sm:col-span-2">
            <Textarea v-model="form.message" class="min-h-24" />
          </AdminFormField>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '创建中' : '创建礼品卡' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false }
})
const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
