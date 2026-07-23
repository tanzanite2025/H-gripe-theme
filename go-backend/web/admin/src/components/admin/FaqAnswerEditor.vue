<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { ImagePlus, Link2, LoaderCircle, Minus, Bold, Italic } from '@lucide/vue'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import axios from '@/utils/axios'

const props = withDefaults(defineProps<{
  modelValue: string
  imageUrl?: string
  imageAlt?: string
}>(), {
  imageUrl: '',
  imageAlt: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:imageUrl': [value: string]
  'update:imageAlt': [value: string]
}>()

const editor = ref<HTMLElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const linkURL = ref('')
const uploading = ref(false)
const syncing = ref(false)

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
}

const exec = (command: string, value?: string) => {
  editor.value?.focus()
  document.execCommand(command, false, value)
  emitContent()
}

const insertLink = () => {
  const value = linkURL.value.trim()
  if (!value) {
    toast.error('请输入链接地址')
    return
  }
  if (!/^(https?:\/\/|mailto:|\/)/i.test(value)) {
    toast.error('链接只支持 http、https、mailto 或站内路径')
    return
  }
  exec('createLink', value)
  linkURL.value = ''
}

const validateImage = (file: File) => new Promise<boolean>((resolve) => {
  if (file.type !== 'image/webp' || !file.name.toLowerCase().endsWith('.webp')) {
    toast.error('FAQ 图片必须是 WebP 格式')
    resolve(false)
    return
  }
  if (file.size > 3 * 1024 * 1024) {
    toast.error('FAQ 图片不能超过 3MB')
    resolve(false)
    return
  }
  const url = URL.createObjectURL(file)
  const image = new Image()
  image.onload = () => {
    URL.revokeObjectURL(url)
    if (image.width !== 800 || image.height !== 800) {
      toast.error(`FAQ 图片必须是 800×800 像素，当前为 ${image.width}×${image.height}`)
      resolve(false)
      return
    }
    resolve(true)
  }
  image.onerror = () => {
    URL.revokeObjectURL(url)
    toast.error('无法读取图片尺寸')
    resolve(false)
  }
  image.src = url
})

const uploadImage = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !(await validateImage(file))) return

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const response = await axios.post('/api/admin/faqs/answer-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    const image = response.data.image
    emit('update:imageUrl', image.url)
    emit('update:imageAlt', props.imageAlt || file.name.replace(/\.webp$/i, ''))
    toast.success('FAQ 图片上传成功')
  } catch (error) {
    console.error('Failed to upload FAQ answer image:', error)
    toast.error('FAQ 图片上传失败')
  } finally {
    uploading.value = false
  }
}

const chooseImage = () => {
  fileInput.value?.click()
}

const removeImage = () => {
  emit('update:imageUrl', '')
  emit('update:imageAlt', '')
}
</script>

<template>
  <div class="space-y-3">
    <div class="rounded-2xl border bg-muted/20 p-2">
      <div class="mb-2 flex flex-wrap items-center gap-1 border-b pb-2">
        <Button type="button" variant="ghost" size="icon-sm" title="加粗" @click="exec('bold')">
          <Bold class="size-4" />
        </Button>
        <Button type="button" variant="ghost" size="icon-sm" title="斜体" @click="exec('italic')">
          <Italic class="size-4" />
        </Button>
        <div class="flex min-w-[16rem] flex-1 items-center gap-1">
          <Input v-model="linkURL" class="h-8 text-xs" placeholder="https://... 或 /support/..." @keydown.enter.prevent="insertLink" />
          <Button type="button" variant="outline" size="sm" class="h-8 px-2 text-xs" @click="insertLink">
            <Link2 class="size-3.5" />
            插入链接
          </Button>
        </div>
      </div>
      <div
        ref="editor"
        contenteditable="true"
        role="textbox"
        aria-multiline="true"
        class="min-h-32 rounded-xl bg-background px-3 py-2 text-sm leading-6 outline-none focus:ring-2 focus:ring-ring"
        data-placeholder="请输入 FAQ 答案，可使用段落、加粗、斜体和链接"
        @input="emitContent"
      />
    </div>
    <p class="text-[11px] text-muted-foreground">
      仅支持轻量文本和链接；图片不放入答案正文，单独上传一张 800×800 WebP。
    </p>

    <div class="rounded-2xl border border-dashed bg-muted/15 p-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <p class="text-xs font-black">FAQ 答案图片</p>
          <p class="mt-0.5 text-[11px] text-muted-foreground">固定 800×800 px · WebP · 最大 3MB · 每条 FAQ 最多一张</p>
        </div>
        <Button type="button" variant="outline" size="sm" class="h-8 px-3" :disabled="uploading" @click="chooseImage">
          <LoaderCircle v-if="uploading" class="size-3.5 animate-spin" />
          <ImagePlus v-else class="size-3.5" />
          {{ uploading ? '上传中' : imageUrl ? '更换图片' : '上传图片' }}
        </Button>
        <input ref="fileInput" type="file" class="sr-only" accept=".webp,image/webp" :disabled="uploading" @change="uploadImage" />
      </div>

      <div v-if="imageUrl" class="mt-3 grid gap-3 sm:grid-cols-[7rem_minmax(0,1fr)_auto] sm:items-center">
        <img :src="imageUrl" :alt="imageAlt || 'FAQ 图片预览'" class="aspect-square w-28 rounded-xl border object-cover" />
        <Input
          :model-value="imageAlt"
          class="h-9"
          placeholder="图片替代文本"
          @update:model-value="emit('update:imageAlt', String($event))"
        />
        <Button type="button" variant="ghost" size="icon" class="text-destructive" title="移除图片" @click="removeImage">
          <Minus class="size-4" />
        </Button>
      </div>
    </div>
  </div>
</template>

<style scoped>
[contenteditable='true']:empty::before {
  color: hsl(var(--muted-foreground));
  content: attr(data-placeholder);
  pointer-events: none;
}
</style>
