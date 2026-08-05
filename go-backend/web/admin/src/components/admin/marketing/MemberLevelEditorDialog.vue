<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '创建会员等级' : '编辑会员等级规则' }}</DialogTitle>
          <DialogDescription>
            {{ mode === 'create' ? '设置积分区间、会员折扣和权益说明。' : '等级名称为系统内置，只维护积分区间、折扣率和权益说明。' }}
          </DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="等级名称" required :error="errors.name" class="sm:col-span-2">
            <Input
              v-if="mode === 'create'"
              v-model="form.name"
              placeholder="请输入等级名称"
              @input="emit('clear-error', 'name')"
            />
            <div v-else class="flex h-9 items-center justify-between rounded-xl border bg-muted/30 px-3 text-sm">
              <span class="font-bold text-foreground">{{ memberLevelLabel(form.name) }}</span>
              <span class="font-mono text-[11px] uppercase text-muted-foreground">{{ form.name }}</span>
            </div>
          </AdminFormField>
          <AdminFormField label="最小积分" required :error="errors.min_points">
            <Input v-model.number="form.min_points" type="number" min="0" step="1" @input="emit('clear-error', 'min_points')" />
          </AdminFormField>
          <AdminFormField label="最大积分" required :error="errors.max_points">
            <Input v-model.number="form.max_points" type="number" min="0" step="1" @input="emit('clear-error', 'max_points')" />
          </AdminFormField>
          <AdminFormField label="折扣率（%）">
            <Input v-model.number="form.discount_rate" type="number" min="0" max="100" step="0.01" />
          </AdminFormField>
          <AdminFormField label="权益说明" class="sm:col-span-2">
            <Textarea v-model="form.benefits" class="min-h-24" />
          </AdminFormField>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存等级' }}
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
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false }
})
const emit = defineEmits(['update:open', 'submit', 'clear-error'])

const memberLevelLabel = (name) => {
  const key = String(name || '').trim().toLowerCase()
  const labels = {
    ordinary: '普通',
    bronze: '铜牌',
    silver: '银牌',
    gold: '金牌',
    platinum: '铂金',
    diamond: '钻石',
  }
  return labels[key] || name || '-'
}
</script>
