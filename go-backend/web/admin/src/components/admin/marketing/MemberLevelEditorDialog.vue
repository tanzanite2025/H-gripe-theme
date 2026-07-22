<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '创建会员等级' : '编辑会员等级' }}</DialogTitle>
          <DialogDescription>设置积分区间、会员折扣和积分倍率。</DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="等级名称" required :error="errors.name" class="sm:col-span-2">
            <Input v-model="form.name" placeholder="请输入等级名称" @input="emit('clear-error', 'name')" />
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
          <AdminFormField label="积分倍数">
            <Input v-model.number="form.points_multiplier" type="number" min="0.01" step="0.01" />
          </AdminFormField>
          <AdminFormField label="排序">
            <Input v-model.number="form.sort_order" type="number" min="0" step="1" />
          </AdminFormField>
          <AdminFormField label="图标">
            <Input v-model="form.icon" placeholder="图标名称或 URL" />
          </AdminFormField>
          <AdminFormField label="颜色" class="sm:col-span-2">
            <div class="flex gap-2">
              <input v-model="form.color" type="color" class="h-9 w-12 cursor-pointer rounded-xl border border-dashed bg-transparent p-1" aria-label="选择等级颜色" />
              <Input v-model="form.color" class="font-mono text-xs uppercase" placeholder="#2563eb" />
            </div>
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
</script>
