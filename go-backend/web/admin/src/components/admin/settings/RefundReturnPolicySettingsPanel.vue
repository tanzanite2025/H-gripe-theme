<template>
  <section class="w-full max-w-none space-y-3">
    <div class="max-w-3xl">
      <h2 class="text-sm font-black uppercase text-foreground">退货退款内容</h2>
      <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">前台政策页与 Warranty / Returns TAB 共用这份内容。</p>
    </div>
    <div class="space-y-5">
      <div class="flex flex-wrap items-end gap-3 border-b border-dashed border-border pb-4">
        <AdminFormField label="编辑语言" class="min-w-56">
          <StorefrontLocaleSelect
            :model-value="locale"
            :language-options="languageOptions"
            :disabled="loading || saving"
            @update:model-value="emit('locale-change', $event)"
          />
        </AdminFormField>
        <div class="pb-1 text-xs text-muted-foreground">
          <span v-if="fallback">当前语言尚未单独配置，正在编辑英文默认内容的副本。</span>
          <span v-else>当前语言已配置独立内容。</span>
        </div>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <AdminFormField label="页面标题" required>
          <Input v-model="policy.title" :disabled="!canEdit || saving" />
        </AdminFormField>
        <AdminFormField label="更新时间" description="保存时由后台自动更新。">
          <Input :model-value="policy.updated_at || '-'" disabled />
        </AdminFormField>
        <AdminFormField label="引导说明" class="md:col-span-2">
          <Textarea v-model="policy.intro" class="min-h-24" :disabled="!canEdit || saving" />
        </AdminFormField>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h3 class="text-sm font-black">政策段落</h3>
            <p class="mt-1 text-xs text-muted-foreground">每段可以有正文、要点和一张说明图片。</p>
          </div>
          <Button v-if="canEdit" type="button" variant="outline" size="sm" :disabled="saving" @click="addSection">
            <Plus class="size-4" />
            添加段落
          </Button>
        </div>

        <div v-if="policy.sections.length === 0" class="rounded-lg border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
          还没有政策段落。
        </div>

        <article
          v-for="(section, index) in policy.sections"
          :key="section.id"
          class="space-y-4 rounded-lg border border-border bg-card/60 p-4"
        >
          <div class="flex items-start justify-between gap-3">
            <div>
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">Section {{ index + 1 }}</span>
              <h4 class="mt-1 text-sm font-black">{{ section.title || '未命名段落' }}</h4>
            </div>
            <div class="flex items-center gap-1">
              <Button type="button" variant="ghost" size="icon" :disabled="!canEdit || saving || index === 0" title="上移" @click="moveSection(index, -1)">
                <ArrowUp class="size-4" />
              </Button>
              <Button type="button" variant="ghost" size="icon" :disabled="!canEdit || saving || index === policy.sections.length - 1" title="下移" @click="moveSection(index, 1)">
                <ArrowDown class="size-4" />
              </Button>
              <Button type="button" variant="ghost" size="icon" :disabled="!canEdit || saving" title="删除段落" @click="removeSection(index)">
                <Trash2 class="size-4" />
              </Button>
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="段落标题" required>
              <Input v-model="section.title" :disabled="!canEdit || saving" />
            </AdminFormField>
            <AdminFormField label="段落 ID" description="用于页面锚点，保持唯一。">
              <Input v-model="section.id" :disabled="!canEdit || saving" class="font-mono" />
            </AdminFormField>
            <AdminFormField label="正文" class="md:col-span-2">
              <Textarea v-model="section.body" class="min-h-24" :disabled="!canEdit || saving" />
            </AdminFormField>
            <AdminFormField label="要点" description="每行一个要点；留空则不显示列表。" class="md:col-span-2">
              <Textarea v-model="section.bulletsText" class="min-h-24" :disabled="!canEdit || saving" />
            </AdminFormField>
          </div>

          <div class="grid gap-4 border-t border-dashed border-border pt-4 md:grid-cols-2 xl:grid-cols-[minmax(0,2.2fr)_minmax(0,170px)_minmax(0,1fr)_minmax(0,1fr)] xl:items-end">
            <AdminFormField label="说明图片 URL" description="也可以使用右侧按钮上传到媒体库。" class="min-w-0">
              <Input v-model="section.image.url" type="url" :disabled="!canEdit || saving" placeholder="https://..." />
            </AdminFormField>
            <div class="flex h-full items-end">
              <Button
                v-if="canEdit"
                type="button"
                variant="outline"
                class="h-9 w-full justify-center"
                :disabled="saving || uploadingSection === index"
                @click="chooseImage(index)"
              >
                <LoaderCircle v-if="uploadingSection === index" class="size-4 animate-spin" />
                <ImagePlus v-else class="size-4" />
                上传图片
              </Button>
              <input
                :ref="(element) => setImageInputRef(index, element)"
                type="file"
                class="sr-only"
                :accept="uploadSpecAccept('refund_return_image')"
                :disabled="!canEdit || saving"
                @change="handleImageChange($event, index)"
              />
              <UploadSpecHint code="refund_return_image" />
            </div>
            <AdminFormField label="图片替代文字" class="min-w-0">
              <Input v-model="section.image.alt" :disabled="!canEdit || saving" />
            </AdminFormField>
            <AdminFormField label="图片说明" class="min-w-0">
              <Input v-model="section.image.caption" :disabled="!canEdit || saving" />
            </AdminFormField>
            <div v-if="section.image.url" class="overflow-hidden rounded-lg border border-border bg-muted md:col-span-2 xl:col-span-4">
              <img :src="section.image.url" :alt="section.image.alt || section.title" class="max-h-64 w-full object-contain" />
            </div>
          </div>
        </article>
      </div>

      <div class="grid gap-4 border-t border-dashed border-border pt-5 md:grid-cols-2">
        <AdminFormField label="联系说明">
          <Textarea v-model="policy.contact_label" class="min-h-20" :disabled="!canEdit || saving" />
        </AdminFormField>
        <AdminFormField label="联系页面 URL">
          <Input v-model="policy.contact_url" :disabled="!canEdit || saving" placeholder="/company/contact" />
        </AdminFormField>
      </div>

      <div class="flex justify-end border-t border-dashed border-border pt-5">
        <Button type="button" :disabled="!canEdit || saving || loading" @click="emit('save')">
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? '保存中' : '保存退货退款内容' }}
        </Button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import {
  ArrowDown,
  ArrowUp,
  ImagePlus,
  LoaderCircle,
  Plus,
  Save,
  Trash2,
} from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { STOREFRONT_SUPPORTED_LANGUAGES, buildLanguageOptions } from '@/lib/languages'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import { uploadSpecAccept } from '@/lib/uploadSpecs'
import type { RefundReturnPolicyEditor, RefundReturnPolicyEditorSection } from '@/api/refundReturnPolicy'

const props = withDefaults(defineProps<{
  policy: RefundReturnPolicyEditor
  locale: string
  fallback?: boolean
  loading?: boolean
  saving?: boolean
  canEdit?: boolean
  uploadingSection?: number | null
}>(), {
  fallback: false,
  loading: false,
  saving: false,
  canEdit: false,
  uploadingSection: null,
})

const emit = defineEmits<{
  (event: 'locale-change', value: string): void
  (event: 'save'): void
  (event: 'upload-image', payload: { index: number; file: File }): void
}>()

const languageOptions = buildLanguageOptions(STOREFRONT_SUPPORTED_LANGUAGES)
const imageInputs = ref<Record<number, HTMLInputElement | null>>({})

const createSection = (): RefundReturnPolicyEditorSection => ({
  id: `section-${props.policy.sections.length + 1}`,
  title: '',
  body: '',
  bullets: [],
  bulletsText: '',
  image: { url: '', alt: '', caption: '' },
})

const addSection = () => {
  props.policy.sections.push(createSection())
}

const removeSection = (index: number) => {
  props.policy.sections.splice(index, 1)
}

const moveSection = (index: number, direction: -1 | 1) => {
  const target = index + direction
  if (target < 0 || target >= props.policy.sections.length) return
  const [section] = props.policy.sections.splice(index, 1)
  props.policy.sections.splice(target, 0, section)
}

const chooseImage = (index: number) => {
  imageInputs.value[index]?.click()
}

const setImageInputRef = (index: number, element: Element | ComponentPublicInstance | null) => {
  imageInputs.value[index] = element instanceof HTMLInputElement ? element : null
}

const handleImageChange = (event: Event, index: number) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (file) emit('upload-image', { index, file })
}
</script>
