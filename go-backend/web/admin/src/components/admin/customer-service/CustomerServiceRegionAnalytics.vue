<template>
  <section class="rounded-3xl border bg-card p-4 shadow-sm">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="flex items-center gap-2 text-sm font-black uppercase tracking-tight">
          <BarChart3 class="size-4 text-primary" />
          今日聊天地区分布
        </h2>
        <p class="mt-1 text-xs text-muted-foreground">
          只统计粗地区，用于运营盘点；不展示 IP 或精确定位
        </p>
      </div>
      <div class="flex items-center gap-2 text-[11px] font-bold text-muted-foreground">
        <span>{{ analytics?.date || todayLocalDate() }}</span>
        <span>·</span>
        <span>共 {{ analytics?.total_conversations || 0 }} 个会话</span>
      </div>
    </div>

    <div v-if="loading" class="flex h-20 items-center justify-center text-muted-foreground">
      <LoaderCircle class="size-5 animate-spin" />
    </div>
    <div v-else-if="!regionRows.length" class="rounded-2xl border border-dashed p-4 text-xs text-muted-foreground">
      今天暂无可统计的客服聊天地区数据。
    </div>
    <div v-else class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <article
        v-for="region in regionRows"
        :key="region.region_label"
        class="rounded-2xl border bg-muted/25 p-3"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="truncate text-sm font-black text-foreground">{{ region.region_label }}</p>
            <p class="mt-1 text-[11px] text-muted-foreground">
              会员 {{ region.member_count || 0 }} · 游客 {{ region.visitor_count || 0 }}
            </p>
          </div>
          <span class="rounded-full bg-primary/10 px-2 py-1 text-xs font-black text-primary">
            {{ region.count }}
          </span>
        </div>
        <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-background">
          <div
            class="h-full rounded-full bg-primary"
            :style="{ width: `${Math.max(4, Number(region.percent || 0))}%` }"
          />
        </div>
        <p class="mt-2 text-[11px] font-bold text-muted-foreground">
          {{ Number(region.percent || 0).toFixed(1) }}%
        </p>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { BarChart3, LoaderCircle } from '@lucide/vue'

const props = defineProps({
  analytics: { type: Object, default: null },
  loading: { type: Boolean, default: false },
})

const regionRows = computed(() => props.analytics?.regions || [])

const todayLocalDate = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
</script>
