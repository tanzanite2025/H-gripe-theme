<template>
  <p class="mt-1 text-[11px] leading-4 text-muted-foreground">
    {{ uploadSpecHint(code) }}
    <span v-if="UPLOAD_SPECS[code].qualityNote"> · {{ UPLOAD_SPECS[code].qualityNote }}</span>
    <span> · 前台显示尺寸由响应式图片系统按展示槽位和 DPR 自动生成</span>
  </p>
</template>

<script setup lang="ts">
import {
  UPLOAD_SPECS,
  uploadSpecHint,
  type UploadSpecCode,
} from '@/lib/uploadSpecs'

defineProps<{
  code: UploadSpecCode
}>()
</script>
