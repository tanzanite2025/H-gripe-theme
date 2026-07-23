<template>
  <Dialog v-model:open="openModel">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="$emit('submit')">
        <DialogHeader>
          <DialogTitle>编辑 FAQ 页面</DialogTitle>
          <DialogDescription>页面标识与 Nuxt 的 PageFaq pageId 对应，标题/副标题用于前端 FAQ 区块展示。</DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="页面标识" required>
            <Input v-model="pageForm.page_id" disabled class="font-mono" />
          </AdminFormField>
          <AdminFormField label="语言" required>
            <Input :model-value="localeName(pageForm.locale)" disabled />
          </AdminFormField>
          <AdminFormField label="页面标题" required>
            <Input v-model="pageForm.title" placeholder="Payment FAQs" />
          </AdminFormField>
          <AdminFormField label="路由路径">
            <Input v-model="pageForm.route_path" placeholder="/support/payment" />
          </AdminFormField>
          <AdminFormField label="所属域">
            <Input v-model="pageForm.domain" placeholder="support / products / guides / company" />
          </AdminFormField>
          <AdminFormField label="排序">
            <Input v-model.number="pageForm.sort_order" type="number" min="0" step="1" />
          </AdminFormField>
          <AdminFormField label="状态" required>
            <Select v-model="pageForm.status">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="active">启用</SelectItem>
                <SelectItem value="hidden">隐藏</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
        </div>

        <AdminFormField label="副标题">
          <Textarea v-model="pageForm.subtitle" class="min-h-20" placeholder="Common questions about..." />
        </AdminFormField>

        <DialogFooter>
          <Button type="button" variant="outline" @click="openModel = false">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存页面' }}
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
import { Textarea } from '@/components/ui/textarea'

const props = defineProps({
  open: { type: Boolean, default: false },
  pageForm: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
  localeName: { type: Function, required: true }
})

const emit = defineEmits(['update:open', 'submit'])

const openModel = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
})
</script>
