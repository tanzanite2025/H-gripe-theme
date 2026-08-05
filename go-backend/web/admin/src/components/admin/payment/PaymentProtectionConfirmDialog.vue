<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="sm" class="gap-4" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>{{ title }}</DialogTitle>
        <DialogDescription>{{ description }}</DialogDescription>
      </DialogHeader>

      <div class="rounded-2xl border border-dashed border-border/80 bg-muted/35 p-3 text-xs">
        <dl class="grid grid-cols-[5rem_minmax(0,1fr)] gap-x-2 gap-y-1">
          <dt class="font-black text-muted-foreground">动作</dt>
          <dd>{{ actionLabel }}</dd>
          <dt class="font-black text-muted-foreground">范围</dt>
          <dd>{{ scopeLabel || '-' }}</dd>
          <dt v-if="controlId" class="font-black text-muted-foreground">控制 ID</dt>
          <dd v-if="controlId" class="font-mono">#{{ controlId }}</dd>
        </dl>
      </div>

      <label class="block space-y-2">
        <span class="text-[11px] font-black uppercase tracking-widest text-muted-foreground">
          输入 {{ confirmationToken }} 确认
        </span>
        <Input v-model="confirmationText" :placeholder="confirmationToken" autocomplete="off" />
      </label>

      <p class="rounded-2xl border border-amber-500/20 bg-amber-500/10 p-3 text-xs text-amber-800">
        {{ boundaryText }}
      </p>

      <DialogFooter>
        <Button variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
        <Button :variant="isPauseOrRevoke ? 'destructive' : 'default'" :disabled="saving || !canConfirm" @click="emit('confirm')">
          <AlertTriangle v-if="isPauseOrRevoke" class="size-4" />
          <ShieldCheck v-else class="size-4" />
          {{ confirmLabel }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { AlertTriangle, ShieldCheck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'

const props = defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  action: { type: String, default: 'force_3ds' },
  scopeLabel: { type: String, default: '' },
  controlId: { type: [Number, String], default: '' },
  saving: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'confirm'])

const confirmationText = ref('')

const isRevoke = computed(() => props.mode === 'revoke')
const isPause = computed(() => props.action === 'pause_payment')
const isPauseOrRevoke = computed(() => isRevoke.value || isPause.value)
const actionLabel = computed(() => {
  const labels = {
    force_3ds: '强制 3DS',
    pause_payment: '暂停新支付',
  }
  return labels[props.action] || props.action || '-'
})
const confirmationToken = computed(() => {
  if (isRevoke.value) return '撤销保护'
  return isPause.value ? '暂停新支付' : '强制 3DS'
})
const title = computed(() => (isRevoke.value ? '撤销人工保护控制' : `启用${actionLabel.value}`))
const description = computed(() => (
  isRevoke.value
    ? '撤销后，匹配范围的新订单与新支付启动将不再受这条控制影响。'
    : '这是一条带过期时间的人工控制，提交后会写入审计记录。'
))
const boundaryText = computed(() => (
  isRevoke.value
    ? '撤销只改变后续保护判断，不会补偿执行期间已经被拦截的客户操作。'
    : '该控制只影响匹配范围内的新订单与新支付启动，不会自动退款、切换支付通道或取消已创建支付。'
))
const confirmLabel = computed(() => (isRevoke.value ? '确认撤销' : '确认启用'))
const canConfirm = computed(() => confirmationText.value.trim() === confirmationToken.value)

watch(() => props.open, (open) => {
  if (!open) confirmationText.value = ''
})
</script>
