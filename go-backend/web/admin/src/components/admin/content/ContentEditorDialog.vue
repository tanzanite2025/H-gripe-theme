<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="max-h-[92dvh] overflow-y-auto p-0" @open-auto-focus.prevent>
      <form @submit.prevent="emit('submit')">
        <DialogHeader class="border-b px-5 py-4 pr-12">
          <DialogTitle>{{ mode === 'create' ? '添加文章' : '编辑文章' }}</DialogTitle>
          <DialogDescription>正文支持 Markdown；SEO 元数据请在 SEO / 文章中维护。</DialogDescription>
        </DialogHeader>

        <div class="space-y-7 px-5 py-5">
          <section class="grid gap-4 border-t border-dashed pt-6 first:border-t-0 first:pt-0 lg:grid-cols-[170px_minmax(0,1fr)]">
            <div>
              <h3 class="text-sm font-black tracking-tighter uppercase text-foreground">正文内容</h3>
              <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
                文章标题、摘要与 Markdown 正文。
              </p>
            </div>
            <div class="min-w-0">
              <div class="grid gap-4 md:grid-cols-[minmax(0,2fr)_minmax(150px,0.7fr)]">
                <AdminFormField label="标题" required :error="errors.title">
                  <Input v-model="form.title" placeholder="请输入文章标题" @input="emit('clear-error', 'title')" />
                </AdminFormField>
                <AdminFormField
                  label="语言"
                  required
                  :error="errors.locale"
                  :description="mode === 'edit' ? '编辑文章时语言已锁定；如需其他语言，请新建对应语种文章。' : ''"
                >
                  <StorefrontLocaleSelect
                    v-model="form.locale"
                    :language-options="languageOptions"
                    :disabled="mode === 'edit'"
                    :locked="mode === 'edit'"
                    locked-title="文章语言已锁定"
                  />
                </AdminFormField>
                <AdminFormField label="Slug" required :error="errors.slug" class="md:col-span-2">
                  <Input v-model="form.slug" placeholder="例如 crystal-care-guide" @input="emit('clear-error', 'slug')" />
                </AdminFormField>
                <AdminFormField label="摘要" class="md:col-span-2">
                  <Textarea v-model="form.excerpt" class="min-h-24" placeholder="请输入文章摘要" />
                </AdminFormField>
                <AdminFormField label="内容（Markdown）" class="md:col-span-2">
                  <Textarea
                    v-model="form.content"
                    class="min-h-80 resize-y font-mono text-[13px] leading-6"
                    placeholder="请输入文章内容"
                  />
                </AdminFormField>
              </div>
            </div>
          </section>

          <section class="grid gap-4 border-t border-dashed pt-6 first:border-t-0 first:pt-0 lg:grid-cols-[170px_minmax(0,1fr)]">
            <div>
              <h3 class="text-sm font-black tracking-tighter uppercase text-foreground">发布信息</h3>
              <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
                控制文章状态、封面图和内容标签。
              </p>
            </div>
            <div class="min-w-0">
              <div class="grid gap-4 md:grid-cols-2">
                <AdminFormField label="状态" required>
                  <Select v-model="form.status">
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="draft">草稿</SelectItem>
                      <SelectItem value="published">已发布</SelectItem>
                      <SelectItem value="archived">已归档</SelectItem>
                    </SelectContent>
                  </Select>
                </AdminFormField>
                <AdminFormField label="特色图片">
                  <div class="relative">
                    <ImageIcon class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input v-model="form.featured_image" class="pl-9" placeholder="图片 URL" />
                  </div>
                </AdminFormField>
                <AdminFormField label="标签" class="md:col-span-2">
                  <Input v-model="form.tags" placeholder="多个标签用逗号分隔" />
                </AdminFormField>
              </div>
            </div>
          </section>

        </div>

        <DialogFooter class="sticky bottom-0 mx-0 mb-0 rounded-b-lg border-t bg-background px-5 py-4">
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存文章' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Image as ImageIcon, LoaderCircle } from '@lucide/vue'
import type { LanguageOption } from '@/lib/languages'
import AdminFormField from '@/components/admin/AdminFormField.vue'
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
import type { ContentDialogMode, ContentFormErrors, ContentPostForm } from '@/modules/content/contentTypes'

withDefaults(defineProps<{
  open?: boolean
  mode?: ContentDialogMode
  form: ContentPostForm
  errors?: ContentFormErrors
  submitting?: boolean
  languageOptions?: LanguageOption[]
}>(), {
  open: false,
  mode: 'create',
  errors: () => ({}),
  submitting: false,
  languageOptions: () => []
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', key: string): void
}>()
</script>

