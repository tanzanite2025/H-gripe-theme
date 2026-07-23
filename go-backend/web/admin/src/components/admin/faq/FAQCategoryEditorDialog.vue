<template>
  <Dialog v-model:open="openModel">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="$emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '添加 FAQ 分类' : '编辑 FAQ 分类' }}</DialogTitle>
          <DialogDescription>分类标识是 FAQ 内容归属字段；分类名称和图标用于 Nuxt 前端展示。</DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="页面" required>
            <Select v-model="categoryForm.page_id">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="page in structurePageOptions" :key="page.value" :value="page.value">
                  {{ page.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <AdminFormField label="语言" required>
            <Select v-model="categoryForm.locale">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="zh">中文</SelectItem>
                <SelectItem value="en">English</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <AdminFormField label="分类名称" required>
            <Input v-model="categoryForm.name" placeholder="Payment Methods" />
          </AdminFormField>
          <AdminFormField label="分类标识" required>
            <Input v-model="categoryForm.category_key" class="font-mono" placeholder="payment-methods" />
          </AdminFormField>
          <AdminFormField label="图标">
            <Input v-model="categoryForm.icon" placeholder="例如 💳，可为空" />
          </AdminFormField>
          <AdminFormField label="排序">
            <Input v-model.number="categoryForm.sort_order" type="number" min="0" step="1" />
          </AdminFormField>
          <AdminFormField label="状态" required>
            <Select v-model="categoryForm.status">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="active">启用</SelectItem>
                <SelectItem value="hidden">隐藏</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="openModel = false">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存分类' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { computed } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const props = defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, required: true },
  categoryForm: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
  structurePageOptions: { type: Array, required: true }
})

const emit = defineEmits(['update:open', 'submit'])

const openModel = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
})
</script>
