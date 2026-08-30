<template>
  <p
    class="mt-1 text-[11px] leading-4 text-muted-foreground"
    :class="compact ? 'line-clamp-2 break-words text-[10px]' : ''"
    :title="fullHint"
  >
    {{ compact ? compactHint : fullHint }}
  </p>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  UPLOAD_SPECS,
  uploadSpecHint,
  type UploadSpecCode,
} from '@/lib/uploadSpecs'

const props = withDefaults(defineProps<{
  code: UploadSpecCode
  compact?: boolean
}>(), {
  compact: false,
})

const compactHint = computed(() => uploadSpecHint(props.code))

const fullHint = computed(() => {
  const spec = UPLOAD_SPECS[props.code]
  const qualityNote = spec.qualityNote ? ` · ${spec.qualityNote}` : ''
  return `${uploadSpecHint(props.code)}${qualityNote} · 前台显示尺寸由响应式图片系统按展示槽位和 DPR 自动生成`
})
</script>
