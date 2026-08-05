<template>
  <AdminFormSection title="商品媒体" description="这里只维护商品主图、图库图片和商品视频；详情内容中的图片或视频请放在详细描述里。">
    <div class="space-y-4">
      <div class="grid gap-3 md:grid-cols-2">
        <label class="flex min-h-24 cursor-pointer flex-col items-center justify-center gap-2 rounded-xl border border-dashed bg-muted/20 px-4 py-5 text-center transition hover:border-primary/60 hover:bg-primary/5">
          <input
            class="sr-only"
            type="file"
            accept="image/jpeg,image/png,image/webp"
            multiple
            :disabled="uploading"
            @change="emit('upload', $event, 'image')"
          />
          <ImageIcon class="size-5 text-muted-foreground" />
          <span class="text-sm font-medium">上传商品图片</span>
          <span class="text-xs text-muted-foreground">主图或图库图片，按排序展示在商品页媒体区</span>
        </label>

        <label class="flex min-h-24 cursor-pointer flex-col items-center justify-center gap-2 rounded-xl border border-dashed bg-muted/20 px-4 py-5 text-center transition hover:border-primary/60 hover:bg-primary/5">
          <input
            class="sr-only"
            type="file"
            accept="video/mp4,video/quicktime,video/webm"
            multiple
            :disabled="uploading"
            @change="emit('upload', $event, 'video')"
          />
          <Video class="size-5 text-muted-foreground" />
          <span class="text-sm font-medium">上传商品视频</span>
          <span class="text-xs text-muted-foreground">支持 MP4 / MOV / WEBM，视频可配置封面图</span>
        </label>
      </div>

      <div class="flex flex-wrap gap-2">
        <Button type="button" variant="outline" size="sm" @click="emit('add-url', 'image')">
          <Plus class="size-3.5" />
          添加图片 URL
        </Button>
        <Button type="button" variant="outline" size="sm" @click="emit('add-url', 'video')">
          <Plus class="size-3.5" />
          添加视频 URL
        </Button>
        <span v-if="uploading" class="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
          <LoaderCircle class="size-3.5 animate-spin" />
          媒体上传中…
        </span>
      </div>

      <p v-if="error" class="text-xs font-medium text-destructive">{{ error }}</p>

      <div v-if="mediaItems.length" class="grid gap-3 lg:grid-cols-2">
        <div
          v-for="(mediaItem, index) in mediaItems"
          :key="mediaItem.local_key || mediaItem.id || `${mediaItem.media_type}-${index}`"
          class="min-w-0 rounded-xl border bg-background/80 p-3"
        >
          <div class="grid gap-3 md:grid-cols-[9rem_minmax(0,1fr)]">
            <div class="relative aspect-square overflow-hidden rounded-lg border bg-muted/40">
              <img
                v-if="mediaItem.media_type === 'image' && mediaItem.url"
                :src="mediaItem.url"
                :alt="mediaItem.alt || mediaItem.title || '商品图片'"
                class="h-full w-full object-contain"
              />
              <video
                v-else-if="mediaItem.media_type === 'video' && mediaItem.url"
                :src="mediaItem.url"
                :poster="mediaItem.poster_url || mediaItem.thumbnail_url"
                class="h-full w-full bg-slate-950 object-contain"
                controls
                preload="metadata"
              />
              <div v-else class="flex h-full w-full items-center justify-center text-muted-foreground">
                <ImageIcon v-if="mediaItem.media_type === 'image'" class="size-7" />
                <Video v-else class="size-7" />
              </div>
              <span class="absolute left-2 top-2 rounded-full bg-background/90 px-2 py-0.5 text-[11px] font-medium shadow-sm">
                {{ getProductMediaTypeLabel(mediaItem.media_type) }}
              </span>
              <span
                v-if="mediaItem.is_primary"
                class="absolute right-2 top-2 rounded-full bg-amber-500 px-2 py-0.5 text-[11px] font-bold text-white shadow-sm"
              >
                主图
              </span>
            </div>

            <div class="min-w-0 space-y-3">
              <div class="grid gap-3 sm:grid-cols-2">
                <AdminFormField :label="mediaItem.media_type === 'video' ? '媒体类型' : '图片位置'">
                  <Select v-if="mediaItem.media_type === 'image'" v-model="mediaItem.role">
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem
                        v-for="option in getProductMediaRoleOptions(mediaItem.media_type)"
                        :key="option.value"
                        :value="option.value"
                      >
                        {{ option.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <div
                    v-else
                    class="flex h-9 items-center rounded-md border bg-muted/30 px-3 text-sm font-medium text-muted-foreground"
                  >
                    商品视频
                  </div>
                </AdminFormField>
                <AdminFormField label="排序">
                  <Input v-model.number="mediaItem.sort_order" type="number" min="0" />
                </AdminFormField>
              </div>

              <AdminFormField label="媒体 URL" required>
                <Input v-model="mediaItem.url" placeholder="上传后自动填充，也可粘贴外部 CDN URL" @input="emit('clear-error')" />
              </AdminFormField>

              <div class="grid gap-3 sm:grid-cols-2">
                <AdminFormField label="标题">
                  <Input v-model="mediaItem.title" placeholder="内部识别标题" />
                </AdminFormField>
                <AdminFormField label="Alt 文案">
                  <Input v-model="mediaItem.alt" placeholder="图片替代文本 / 视频说明" />
                </AdminFormField>
              </div>

              <AdminFormField v-if="mediaItem.media_type === 'video'" label="视频封面 URL">
                <Input v-model="mediaItem.poster_url" placeholder="视频封面图 URL，可后续上传图片后粘贴" />
              </AdminFormField>

              <div class="flex flex-wrap justify-between gap-2">
                <div class="flex flex-wrap gap-2">
                  <Button
                    v-if="mediaItem.media_type === 'image'"
                    type="button"
                    variant="outline"
                    size="sm"
                    @click="emit('set-primary', index)"
                  >
                    <Star class="size-3.5" />
                    设为主图
                  </Button>
                  <Button type="button" variant="outline" size="sm" :disabled="index === 0" @click="emit('move', index, -1)">
                    上移
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    :disabled="index === mediaItems.length - 1"
                    @click="emit('move', index, 1)"
                  >
                    下移
                  </Button>
                </div>
                <Button type="button" variant="destructive" size="sm" @click="emit('remove', index)">
                  <Trash2 class="size-3.5" />
                  删除
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="rounded-xl border bg-muted/20 px-4 py-6 text-center text-sm text-muted-foreground">
        暂未添加商品媒体。新商品上线前建议至少上传一张商品主图。
      </div>
    </div>
  </AdminFormSection>
</template>

<script setup>
import { ImageIcon, LoaderCircle, Plus, Star, Trash2, Video } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  getProductMediaRoleOptions,
  getProductMediaTypeLabel,
} from '@/lib/productMedia'

defineProps({
  mediaItems: { type: Array, default: () => [] },
  uploading: { type: Boolean, default: false },
  error: { type: String, default: '' },
})

const emit = defineEmits([
  'upload',
  'add-url',
  'clear-error',
  'set-primary',
  'move',
  'remove',
])
</script>
