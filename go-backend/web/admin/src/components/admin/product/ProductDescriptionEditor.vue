<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import {
  Bold,
  Heading2,
  Heading3,
  ImagePlus,
  Italic,
  Link2,
  List,
  ListOrdered,
  Pilcrow,
  Quote,
  Redo2,
  Undo2,
  Video,
} from '@lucide/vue'
import { toast } from 'vue-sonner'
import mediaApi from '@/api/media'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

const props = withDefaults(defineProps<{
  modelValue: string
}>(), {
  modelValue: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editor = ref<HTMLElement | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)
const videoInput = ref<HTMLInputElement | null>(null)
const linkURL = ref('')
const mediaURL = ref('')
const uploading = ref(false)
const syncing = ref(false)
let savedRange: Range | null = null

const isSelectionInsideEditor = () => {
  if (!editor.value) return false
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return false
  const node = selection.anchorNode
  return Boolean(node && editor.value.contains(node))
}

const saveSelection = () => {
  if (!isSelectionInsideEditor()) return
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) return
  savedRange = selection.getRangeAt(0).cloneRange()
}

const restoreSelection = () => {
  editor.value?.focus()
  if (!savedRange) return
  const selection = window.getSelection()
  selection?.removeAllRanges()
  selection?.addRange(savedRange)
}

const syncEditor = async () => {
  if (!editor.value || editor.value === document.activeElement) return
  syncing.value = true
  await nextTick()
  editor.value.innerHTML = props.modelValue || ''
  syncing.value = false
}

onMounted(syncEditor)
watch(() => props.modelValue, syncEditor)

const emitContent = () => {
  if (!editor.value || syncing.value) return
  emit('update:modelValue', editor.value.innerHTML)
  saveSelection()
}

const exec = (command: string, value?: string) => {
  restoreSelection()
  document.execCommand(command, false, value)
  emitContent()
}

const setBlock = (tag: 'P' | 'H2' | 'H3' | 'BLOCKQUOTE') => {
  exec('formatBlock', tag)
}

const escapeAttribute = (value: string) => value
  .replace(/&/g, '&amp;')
  .replace(/"/g, '&quot;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')

const validContentURL = (value: string) => /^(https?:\/\/|\/(?!\/))/i.test(value)

const insertHTML = (html: string) => {
  restoreSelection()
  document.execCommand('insertHTML', false, html)
  emitContent()
}

const insertLink = () => {
  const value = linkURL.value.trim()
  if (!value) {
    toast.error('请输入链接地址')
    return
  }
  if (!/^(https?:\/\/|mailto:|\/(?!\/))/i.test(value)) {
    toast.error('链接只支持 http、https、mailto 或站内路径')
    return
  }

  restoreSelection()
  const selection = window.getSelection()
  if (selection && selection.rangeCount > 0 && !selection.getRangeAt(0).collapsed) {
    document.execCommand('createLink', false, value)
    emitContent()
  } else {
    const safeURL = escapeAttribute(value)
    insertHTML(`<a href="${safeURL}">${safeURL}</a>`)
  }
  linkURL.value = ''
}

const insertMediaURL = (type: 'image' | 'video') => {
  const value = mediaURL.value.trim()
  if (!value) {
    toast.error('请输入媒体地址')
    return
  }
  if (!validContentURL(value)) {
    toast.error('媒体地址只支持 http、https 或站内路径')
    return
  }
  insertMedia(type, value)
  mediaURL.value = ''
}

const insertMedia = (type: 'image' | 'video', url: string, title = '') => {
  const safeURL = escapeAttribute(url)
  const safeTitle = escapeAttribute(title)
  if (type === 'image') {
    insertHTML(`<figure><img src="${safeURL}" alt="${safeTitle}" loading="lazy"></figure><p><br></p>`)
    return
  }
  insertHTML(`<figure><video src="${safeURL}" controls preload="metadata"></video></figure><p><br></p>`)
}

const chooseUpload = (type: 'image' | 'video') => {
  if (type === 'image') {
    imageInput.value?.click()
    return
  }
  videoInput.value?.click()
}

const uploadAndInsert = async (event: Event, type: 'image' | 'video') => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('media_type', type)
    const asset = await mediaApi.uploadAsset(formData)
    const url = String(asset?.url || '').trim()
    if (!url) {
      toast.error('上传成功但没有返回媒体地址')
      return
    }
    insertMedia(type, url, asset?.alt || asset?.original_filename || file.name)
    toast.success(type === 'image' ? '详情图片已插入' : '详情视频已插入')
  } catch (error) {
    console.error('Failed to upload product description media:', error)
    toast.error(type === 'image' ? '详情图片上传失败' : '详情视频上传失败')
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div class="rounded-lg border bg-muted/15 p-2">
    <div class="flex flex-wrap items-center gap-1 border-b pb-2">
      <Button type="button" variant="ghost" size="icon-sm" title="撤销" @click="exec('undo')">
        <Undo2 class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="重做" @click="exec('redo')">
        <Redo2 class="size-4" />
      </Button>
      <span class="mx-1 h-5 w-px bg-border" aria-hidden="true" />
      <Button type="button" variant="ghost" size="icon-sm" title="段落" @click="setBlock('P')">
        <Pilcrow class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="二级标题" @click="setBlock('H2')">
        <Heading2 class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="三级标题" @click="setBlock('H3')">
        <Heading3 class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="引用" @click="setBlock('BLOCKQUOTE')">
        <Quote class="size-4" />
      </Button>
      <span class="mx-1 h-5 w-px bg-border" aria-hidden="true" />
      <Button type="button" variant="ghost" size="icon-sm" title="加粗" @click="exec('bold')">
        <Bold class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="斜体" @click="exec('italic')">
        <Italic class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="无序列表" @click="exec('insertUnorderedList')">
        <List class="size-4" />
      </Button>
      <Button type="button" variant="ghost" size="icon-sm" title="有序列表" @click="exec('insertOrderedList')">
        <ListOrdered class="size-4" />
      </Button>

      <div class="flex min-w-[12rem] flex-1 items-center gap-1">
        <Input
          v-model="linkURL"
          class="h-7 min-w-0 text-xs"
          placeholder="链接 URL"
          @keydown.enter.prevent="insertLink"
        />
        <Button type="button" variant="outline" size="icon-sm" title="插入链接" @click="insertLink">
          <Link2 class="size-4" />
        </Button>
      </div>

      <div class="flex min-w-[14rem] flex-1 items-center gap-1">
        <Input
          v-model="mediaURL"
          class="h-7 min-w-0 text-xs"
          placeholder="图片 / 视频 URL"
          @keydown.enter.prevent="insertMediaURL('image')"
        />
        <Button type="button" variant="outline" size="icon-sm" title="插入图片 URL" @click="insertMediaURL('image')">
          <ImagePlus class="size-4" />
        </Button>
        <Button type="button" variant="outline" size="icon-sm" title="插入视频 URL" @click="insertMediaURL('video')">
          <Video class="size-4" />
        </Button>
      </div>

      <Button type="button" variant="outline" size="icon-sm" title="上传详情图片" :disabled="uploading" @click="chooseUpload('image')">
        <ImagePlus class="size-4" />
      </Button>
      <Button type="button" variant="outline" size="icon-sm" title="上传详情视频" :disabled="uploading" @click="chooseUpload('video')">
        <Video class="size-4" />
      </Button>
      <input ref="imageInput" type="file" class="sr-only" accept="image/jpeg,image/png,image/webp" :disabled="uploading" @change="uploadAndInsert($event, 'image')" />
      <input ref="videoInput" type="file" class="sr-only" accept="video/mp4,video/quicktime,video/webm" :disabled="uploading" @change="uploadAndInsert($event, 'video')" />
    </div>

    <div
      ref="editor"
      contenteditable="true"
      role="textbox"
      aria-multiline="true"
      class="product-description-editor__canvas mt-2 min-h-48 rounded-md bg-background px-4 py-3 text-sm leading-6 outline-none focus:ring-2 focus:ring-ring"
      data-placeholder="请输入商品详细描述"
      @focus="saveSelection"
      @keyup="saveSelection"
      @mouseup="saveSelection"
      @input="emitContent"
      @blur="emitContent"
    />
  </div>
</template>

<style scoped>
.product-description-editor__canvas:empty::before {
  color: hsl(var(--muted-foreground));
  content: attr(data-placeholder);
  pointer-events: none;
}

.product-description-editor__canvas :deep(h2),
.product-description-editor__canvas :deep(h3) {
  margin: 0.85rem 0 0.45rem;
  font-weight: 800;
}

.product-description-editor__canvas :deep(p),
.product-description-editor__canvas :deep(ul),
.product-description-editor__canvas :deep(ol),
.product-description-editor__canvas :deep(blockquote),
.product-description-editor__canvas :deep(figure) {
  margin: 0 0 0.75rem;
}

.product-description-editor__canvas :deep(ul),
.product-description-editor__canvas :deep(ol) {
  padding-left: 1.3rem;
  list-style-position: outside;
}

.product-description-editor__canvas :deep(ul) {
  list-style-type: disc;
}

.product-description-editor__canvas :deep(ol) {
  list-style-type: decimal;
}

.product-description-editor__canvas :deep(blockquote) {
  border-left: 3px solid hsl(var(--primary));
  color: hsl(var(--muted-foreground));
  padding-left: 0.85rem;
}

.product-description-editor__canvas :deep(img),
.product-description-editor__canvas :deep(video) {
  display: block;
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
}
</style>
