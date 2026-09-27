<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CheckCircle2, PackageCheck, Power, RefreshCw, Search } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuthStore } from '@/stores/auth'
import shippingApi from '@/api/shipping'

type FpxChannel = {
  id: number
  service_code: string
  display_name: string
  enabled: boolean
}

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:fpx:manage'))
const collectionSearch = ref('')
const loading = ref(false)
const errorMessage = ref('')
const curatedChannels = ref<FpxChannel[]>([])

const filteredChannels = computed(() => {
  const query = collectionSearch.value.trim().toLowerCase()
  if (!query) return curatedChannels.value
  return curatedChannels.value.filter((channel) => (
    channel.display_name.toLowerCase().includes(query) || channel.service_code.toLowerCase().includes(query)
  ))
})

const enabledCount = computed(() => curatedChannels.value.filter((channel) => channel.enabled).length)

const loadChannels = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    curatedChannels.value = await shippingApi.listFpxChannels() as FpxChannel[]
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载 4PX 渠道失败'
  } finally {
    loading.value = false
  }
}

const toggleChannel = async (channel: FpxChannel) => {
  if (!canManage.value) return
  errorMessage.value = ''
  try {
    const updated = await shippingApi.confirmFpxChannel(channel.id, { enabled: !channel.enabled })
    const index = curatedChannels.value.findIndex((item) => item.id === channel.id)
    if (index >= 0) curatedChannels.value[index] = updated as FpxChannel
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '更新 4PX 渠道状态失败'
  }
}

onMounted(loadChannels)
</script>

<template>
  <div class="space-y-5">
    <section class="rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-violet-600">4PX Channel References</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">4PX 服务集合</h2>
          <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">展示已同步的官方服务代码，并管理下游可引用的启用状态。</p>
        </div>
        <Button variant="outline" size="sm" :disabled="loading" title="刷新 4PX 服务" @click="loadChannels">
          <RefreshCw class="size-3.5" />刷新服务
        </Button>
      </div>

      <p v-if="errorMessage" class="mt-4 rounded-xl border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-700 dark:text-red-300">{{ errorMessage }}</p>

      <div class="mt-5 grid gap-3 md:grid-cols-3">
        <article class="rounded-[22px] border border-dashed border-border/80 bg-muted/20 p-4">
          <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">已启用渠道</p>
          <p class="mt-2 text-2xl font-black">{{ enabledCount }}</p>
          <p class="mt-1 text-[11px] text-muted-foreground">允许下游业务引用</p>
        </article>
        <article class="rounded-[22px] border border-dashed border-border/80 bg-muted/20 p-4">
          <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">目录总量</p>
          <p class="mt-2 text-2xl font-black">{{ curatedChannels.length }}</p>
          <p class="mt-1 text-[11px] text-muted-foreground">来自 4PX 官方服务同步</p>
        </article>
        <article class="rounded-[22px] border border-emerald-500/30 bg-emerald-500/10 p-4">
          <p class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-emerald-700 dark:text-emerald-300"><CheckCircle2 class="size-3.5" />数据边界</p>
          <p class="mt-2 text-sm font-black">仅维护服务引用</p>
          <p class="mt-1 text-[11px] text-emerald-700/80 dark:text-emerald-300/80">价格与试算以 4PX 官方接口为准</p>
        </article>
      </div>
    </section>

    <section class="overflow-hidden rounded-[28px] border border-dashed border-border/80 bg-card shadow-sm">
      <div class="flex flex-col gap-3 border-b border-dashed border-border/80 px-5 py-4 sm:flex-row sm:items-end sm:justify-between sm:px-6">
        <div>
          <h2 class="text-base font-black">渠道引用</h2>
          <p class="mt-1 text-xs text-muted-foreground">只有已确认启用的服务才允许后续物流管理引用。</p>
        </div>
        <label class="w-full space-y-1.5 sm:max-w-sm">
          <span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><Search class="size-3" />搜索渠道</span>
          <Input v-model="collectionSearch" placeholder="业务名称 / 官方服务代码" />
        </label>
      </div>

      <div v-if="loading" class="flex min-h-64 items-center justify-center px-6 py-12 text-sm text-muted-foreground">正在加载 4PX 渠道目录…</div>
      <div v-else-if="filteredChannels.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
        <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><PackageCheck class="size-5" /></span>
        <div><p class="text-sm font-black">暂无 4PX 渠道</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">请先在网关配置页保存凭据并同步官方服务代码。</p></div>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[720px] text-left text-sm">
          <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><tr><th class="px-5 py-3">服务名称</th><th class="px-4 py-3">官方服务代码</th><th class="px-4 py-3">确认状态</th><th class="px-5 py-3 text-right">操作</th></tr></thead>
          <tbody class="divide-y divide-dashed divide-border/70">
            <tr v-for="channel in filteredChannels" :key="channel.id">
              <td class="px-5 py-4 font-semibold">{{ channel.display_name }}</td>
              <td class="px-4 py-4 font-mono text-xs font-bold">{{ channel.service_code }}</td>
              <td class="px-4 py-4"><span :class="channel.enabled ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'bg-muted text-muted-foreground'" class="rounded-full px-2.5 py-1 text-[10px] font-bold">{{ channel.enabled ? '已确认' : '待确认' }}</span></td>
              <td class="px-5 py-4 text-right"><Button variant="outline" size="sm" :disabled="!canManage" @click="toggleChannel(channel)"><Power class="size-3.5" />{{ channel.enabled ? '取消启用' : '确认并启用' }}</Button></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
