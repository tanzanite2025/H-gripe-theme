<template>
  <Card class="shrink-0 gap-0 py-0">
    <CardContent class="flex flex-wrap items-center gap-2.5 p-3">
      <div class="relative min-w-[14rem] flex-1 basis-80">
        <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input :model-value="search" class="pl-8" placeholder="搜索文件名、ALT、说明或 URL" @update:model-value="(value) => emit('update:search', String(value))" @keyup.enter="emit('apply')" />
      </div>
      <div class="w-36 shrink-0">
        <Select :model-value="mediaType" @update:model-value="(value) => emit('update:media-type', String(value))">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部类型</SelectItem>
            <SelectItem value="image">图片</SelectItem>
            <SelectItem value="video">视频</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="w-36 shrink-0">
        <Select :model-value="status" @update:model-value="(value) => emit('update:status', String(value))">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="active">启用</SelectItem>
            <SelectItem value="archived">归档</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="flex shrink-0 items-center gap-2 text-[10px] font-bold text-muted-foreground">
        <Badge variant="outline" class="whitespace-nowrap">TOTAL {{ total }}</Badge>
        <Badge variant="outline" class="whitespace-nowrap">媒体文件与引用关系独立管理</Badge>
      </div>
      <div class="ml-auto flex shrink-0 items-center gap-2">
        <Button variant="outline" size="sm" @click="emit('reset')">重置</Button>
        <Button size="sm" @click="emit('apply')">应用筛选</Button>
        <Button variant="outline" size="icon-sm" :disabled="loading" title="刷新媒体库" aria-label="刷新媒体库" @click="emit('refresh')">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
        </Button>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { RefreshCw, Search } from '@lucide/vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

withDefaults(defineProps<{
  search?: string
  mediaType?: string
  status?: string
  total?: number
  loading?: boolean
}>(), {
  search: '',
  mediaType: 'all',
  status: 'all',
  total: 0,
  loading: false
})

const emit = defineEmits<{
  (event: 'update:search', value: string): void
  (event: 'update:media-type', value: string): void
  (event: 'update:status', value: string): void
  (event: 'apply'): void
  (event: 'reset'): void
  (event: 'refresh'): void
}>()
</script>
