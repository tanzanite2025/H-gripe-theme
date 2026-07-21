<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增模板绑定' : '编辑模板绑定' }}</DialogTitle>
          <DialogDescription>
            将运费模板绑定到默认、产品类型、单个产品或具体 SKU。匹配时优先级越高越先命中。
          </DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 lg:grid-cols-3">
          <AdminFormField label="运费模板" required :error="errors.template_id">
            <Select v-model="form.template_id" @update:model-value="emit('clear-error', 'template_id')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择模板" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="template in templates" :key="template.id" :value="String(template.id)">
                  {{ template.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="绑定范围" required :error="errors.scope">
            <Select v-model="form.scope" @update:model-value="emit('clear-error', 'scope')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择绑定范围" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="default">默认模板</SelectItem>
                <SelectItem value="product_type">产品类型</SelectItem>
                <SelectItem value="product">单个产品</SelectItem>
                <SelectItem value="variant">SKU / 变体</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="优先级">
            <Input v-model.number="form.priority" type="number" step="1" />
          </AdminFormField>

          <AdminFormField v-if="form.scope === 'product_type'" label="产品类型 ID" required :error="errors.product_type_id">
            <Input v-model.number="form.product_type_id" type="number" min="1" step="1" @input="emit('clear-error', 'product_type_id')" />
          </AdminFormField>

          <AdminFormField v-if="form.scope === 'product'" label="产品 ID" required :error="errors.product_id">
            <Input v-model.number="form.product_id" type="number" min="1" step="1" @input="emit('clear-error', 'product_id')" />
          </AdminFormField>

          <AdminFormField v-if="form.scope === 'variant'" label="SKU / 变体 ID" required :error="errors.variant_id">
            <Input v-model.number="form.variant_id" type="number" min="1" step="1" @input="emit('clear-error', 'variant_id')" />
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-medium">启用绑定</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后保留记录但不参与匹配。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用模板绑定" />
          </div>
        </div>

        <div class="rounded-lg border bg-muted/35 p-3 text-xs text-muted-foreground">
          匹配建议：SKU / 变体 &gt; 单个产品 &gt; 产品类型 &gt; 默认模板。同一范围内按优先级从高到低匹配。
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存绑定' }}
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  templates: { type: Array, default: () => [] },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
