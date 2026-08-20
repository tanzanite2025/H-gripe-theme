<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="max-h-[calc(100dvh-1rem)] overflow-y-auto p-4 sm:p-5" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增选型配置 Key' : '编辑选型配置 Key' }}</DialogTitle>
          <DialogDescription>Key 只允许从这里维护，已使用的 Code 不建议修改。</DialogDescription>
        </DialogHeader>

        <section class="grid gap-3 border-t border-dashed border-border/70 pt-4">
          <AdminFormField label="类型" required description="决定这是问题 KEY 还是回答 KEY">
            <Select v-model="form.kind" :disabled="disabled || mode === 'edit'">
              <SelectTrigger class="w-full">
                <SelectValue placeholder="选择类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="item in kindOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField
            label="Code"
            required
            :description="mode === 'create' ? '例如 riding_category，创建后不建议修改' : '创建后锁定，避免引用失真'"
          >
            <Input v-model="form.code" class="font-mono" placeholder="riding_category" :disabled="disabled || mode === 'edit'" />
          </AdminFormField>

          <AdminFormField label="显示名称" required description="后台列表里看到的名字">
            <Input v-model="form.display_label" placeholder="骑行分类" :disabled="disabled" />
          </AdminFormField>

          <AdminFormField label="说明" description="给运营和编辑看的补充说明">
            <Textarea v-model="form.description" class="min-h-24 resize-y" placeholder="可选" :disabled="disabled" />
          </AdminFormField>

          <div class="grid gap-3 sm:grid-cols-2">
            <AdminFormField label="排序" description="数值越小越靠前">
              <Input v-model.number="form.sort_order" type="number" min="1" step="1" :disabled="disabled" />
            </AdminFormField>
            <label class="inline-flex items-center gap-2 pt-6 text-sm font-bold">
              <Switch v-model="form.is_enabled" :disabled="disabled" />
              启用
            </label>
          </div>
        </section>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="disabled || saving">
            <Save class="size-4" />
            {{ saving ? '保存中' : '保存' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Save } from '@lucide/vue'
import type { SelectionConfigurationKeyKind } from '@/modules/selection-configuration/selectionConfigurationKeys'

export interface SelectionConfigurationKeyEditorForm {
  id?: number
  kind: SelectionConfigurationKeyKind
  code: string
  display_label: string
  description: string
  is_enabled: boolean
  sort_order: number
}

interface KindOption {
  label: string
  value: SelectionConfigurationKeyKind
}

withDefaults(defineProps<{
  open?: boolean
  mode?: 'create' | 'edit'
  form: SelectionConfigurationKeyEditorForm
  kindOptions: KindOption[]
  disabled?: boolean
  saving?: boolean
}>(), {
  open: false,
  mode: 'create',
  disabled: false,
  saving: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
}>()
</script>
