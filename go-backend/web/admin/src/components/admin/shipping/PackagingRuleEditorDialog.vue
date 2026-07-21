<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增包装规则' : '编辑包装规则' }}</DialogTitle>
          <DialogDescription>
            设置包装箱重量、尺寸和最大承重。产品/SKU 绑定会在后续独立面板里做，避免这个弹窗变得臃肿。
          </DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 lg:grid-cols-4">
          <AdminFormField label="规则名称" required :error="errors.rule_name" class="lg:col-span-2">
            <Input v-model.trim="form.rule_name" placeholder="例如 小件空运包装" @input="emit('clear-error', 'rule_name')" />
          </AdminFormField>

          <AdminFormField label="包装重量 kg" required :error="errors.box_weight">
            <Input v-model.number="form.box_weight" type="number" min="0" step="0.001" @input="emit('clear-error', 'box_weight')" />
          </AdminFormField>

          <AdminFormField label="最大承重 kg" required :error="errors.max_weight">
            <Input v-model.number="form.max_weight" type="number" min="0" step="0.001" @input="emit('clear-error', 'max_weight')" />
          </AdminFormField>

          <AdminFormField label="长 cm">
            <Input v-model.number="form.box_length" type="number" min="0" step="0.01" />
          </AdminFormField>

          <AdminFormField label="宽 cm">
            <Input v-model.number="form.box_width" type="number" min="0" step="0.01" />
          </AdminFormField>

          <AdminFormField label="高 cm">
            <Input v-model.number="form.box_height" type="number" min="0" step="0.01" />
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-medium">启用规则</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后不会参与包装匹配。</p>
            </div>
            <Switch v-model="form.is_active" aria-label="启用包装规则" />
          </div>

          <AdminFormField label="说明" class="lg:col-span-4">
            <Textarea v-model="form.description" class="min-h-24" placeholder="适用场景、材料说明或内部备注" />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存包装规则' }}
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
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
