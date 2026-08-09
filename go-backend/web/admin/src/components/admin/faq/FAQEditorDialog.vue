<template>
  <Dialog v-model:open="openModel">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="$emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ dialogMode === 'create' ? '添加 FAQ' : '编辑 FAQ' }}</DialogTitle>
          <DialogDescription>填写面向访客的问题与答案，并设置其展示位置。</DialogDescription>
        </DialogHeader>

        <AdminFormField label="问题" required :error="formErrors.question">
          <Textarea
            v-model="faqForm.question"
            class="min-h-20"
            placeholder="请输入问题"
            @input="$emit('clear-error', 'question')"
          />
        </AdminFormField>

        <AdminFormField label="答案" required :error="formErrors.answer">
          <FaqAnswerEditor
            :model-value="faqForm.answer"
            v-model:image-url="faqForm.answer_image_url"
            v-model:image-alt="faqForm.answer_image_alt"
            @update:model-value="$emit('update-answer', $event)"
          />
        </AdminFormField>

        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField
            label="语言"
            required
            :error="formErrors.locale"
            :description="isPlacementReadonly ? '当前上下文已锁定' : ''"
          >
            <StorefrontLocaleSelect
              v-model="faqForm.locale"
              :language-options="languageOptions"
              :disabled="isPlacementReadonly"
              :locked="isPlacementReadonly"
              locked-title="当前上下文已锁定"
            />
          </AdminFormField>

          <AdminFormField label="页面" required :error="formErrors.page_id" :description="placementLocked ? '从分类添加时页面已锁定' : ''">
            <Select v-model="faqForm.page_id" :disabled="placementLocked">
              <SelectTrigger class="w-full"><SelectValue placeholder="选择前端页面" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="page in faqPageOptions" :key="page.value" :value="page.value">
                  {{ page.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="分类" required :error="formErrors.category" :description="placementLocked ? '从分类添加时分类已锁定' : ''">
            <Select v-model="faqForm.category" :disabled="placementLocked || availableFaqCategories.length === 0">
              <SelectTrigger class="w-full"><SelectValue placeholder="选择页面分类" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="category in availableFaqCategories" :key="category.category_key" :value="category.category_key">
                  {{ category.icon ? `${category.icon} ` : '' }}{{ category.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="状态" required>
            <Select v-model="faqForm.status">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="published">已发布</SelectItem>
                <SelectItem value="draft">草稿</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="排序">
            <Input v-model.number="faqForm.order" type="number" min="0" max="9999" step="1" />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="openModel = false">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存 FAQ' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import type { LanguageOption } from '@/lib/languages'
import type { FAQCategory } from '@/lib/faqAdminPresentation'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import FaqAnswerEditor from '@/components/admin/FaqAnswerEditor.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
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
import type { FAQDialogMode, FAQForm, FAQFormErrors } from '@/composables/faq/useFaqEditor'

const props = withDefaults(defineProps<{
  open?: boolean
  dialogMode: FAQDialogMode
  faqForm: FAQForm
  formErrors: FAQFormErrors
  submitting?: boolean
  faqPageOptions: LanguageOption[]
  availableFaqCategories: FAQCategory[]
  languageOptions: LanguageOption[]
  placementLocked?: boolean
}>(), {
  open: false,
  submitting: false,
  placementLocked: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', key: keyof FAQForm): void
  (event: 'update-answer', value: string): void
}>()

const openModel = computed<boolean>({
  get: () => props.open,
  set: (value: boolean) => emit('update:open', value)
})

const isPlacementReadonly = computed(() => props.dialogMode === 'edit' || props.placementLocked)
</script>
