<script setup lang="ts">
import { Activity, CheckCircle2, PackageCheck } from '@lucide/vue'
import { computed, onMounted, ref } from 'vue'
import fpxLogisticsAdminApi, { type FpxOverview } from '@/api/fpxLogisticsAdminApi'

const overview = ref<FpxOverview | null>(null)
const loading = ref(true)
const loadError = ref('')

const metrics = computed(() => {
  const channels = overview.value?.channels
  return [
    { label: '生产服务目录', value: channels ? String(channels.total) : '—', detail: '生产环境同步的本地引用', icon: PackageCheck },
    { label: '已启用生产服务', value: channels ? String(channels.enabled) : '—', detail: '提供给物流管理模板选择', icon: CheckCircle2 },
  ]
})

const loadOverview = async () => {
  loading.value = true
  loadError.value = ''
  try {
    overview.value = await fpxLogisticsAdminApi.getFpxOverview()
  } catch (error) {
    overview.value = null
    loadError.value = error instanceof Error ? error.message : '无法读取 4PX 服务目录概览'
  } finally {
    loading.value = false
  }
}

onMounted(loadOverview)
</script>

<template>
  <div class="space-y-5">
    <section class="grid gap-3 sm:grid-cols-2">
      <article v-for="metric in metrics" :key="metric.label" class="group relative min-h-36 overflow-hidden rounded-[24px] border border-dashed border-border/80 bg-card p-4">
        <div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-transparent" />
        <div class="relative flex h-full flex-col justify-between gap-4">
          <div class="flex items-center justify-between gap-2"><p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">{{ metric.label }}</p><component :is="metric.icon" class="size-4 text-muted-foreground/70" aria-hidden="true" /></div>
          <div><p class="font-mono text-3xl font-black tracking-tight text-foreground">{{ metric.value }}</p><p class="mt-1 text-[11px] leading-4 text-muted-foreground">{{ metric.detail }}</p></div>
        </div>
      </article>
    </section>

    <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-muted/5 p-5 sm:p-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div><p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground/70">4PX Production Service Directory</p><h2 class="mt-1 text-lg font-black tracking-tight">生产渠道目录监控</h2><p class="mt-1 text-xs text-muted-foreground">此概览只统计生产环境；物流管理模板只读取生产环境中已启用的 4PX 服务。</p></div>
        <span class="inline-flex w-fit items-center gap-2 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1.5 text-[11px] font-bold text-emerald-700 dark:text-emerald-300"><span class="size-1.5 rounded-full bg-emerald-500" />只读目录</span>
      </div>
      <div v-if="loading" class="flex items-start gap-2 rounded-2xl border border-border/70 bg-card/70 px-4 py-3 text-xs leading-5 text-muted-foreground"><Activity class="mt-0.5 size-4 shrink-0" aria-hidden="true" /><p>正在读取 4PX 服务目录。</p></div>
      <div v-else-if="loadError" class="flex items-start gap-2 rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs leading-5 text-rose-700 dark:text-rose-300"><Activity class="mt-0.5 size-4 shrink-0" aria-hidden="true" /><p>目录读取失败：{{ loadError }}</p></div>
      <div v-else class="flex items-start gap-2 rounded-2xl border border-border/70 bg-card/70 px-4 py-3 text-xs leading-5 text-muted-foreground"><Activity class="mt-0.5 size-4 shrink-0" aria-hidden="true" /><p>数据来自 4PX 服务目录缓存，不执行下单、报价、面单或轨迹请求。</p></div>
    </section>
  </div>
</template>
