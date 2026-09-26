<script setup lang="ts">
import { computed, ref } from 'vue'
import { CheckCircle2, PackageCheck, Plus, Search } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:fpx:manage'))

const collectionDialogOpen = ref(false)
const collectionSearch = ref('')
const curatedChannels = ref<Array<{
  id: string
  displayName: string
  serviceCode: string
  recommendation: string
  enabled: boolean
}>>([])
const collectionForm = ref({
  serviceCode: '',
  displayName: '',
  maxLength: '',
  maxWidth: '',
  maxHeight: '',
  volumeDivisor: '6000',
  maxWeight: '',
  notes: '',
})
const filteredCuratedChannels = computed(() => {
  const query = collectionSearch.value.trim().toLowerCase()
  if (!query) return curatedChannels.value
  return curatedChannels.value.filter((channel) =>
    [channel.displayName, channel.serviceCode, channel.recommendation]
      .some((value) => value.toLowerCase().includes(query)),
  )
})
const canSaveCollectionChannel = computed(() =>
  canManage.value
  && collectionForm.value.serviceCode.trim().length > 0
  && collectionForm.value.displayName.trim().length > 0,
)

const openCollectionDialog = () => {
  collectionDialogOpen.value = true
}

const saveCollectionChannel = () => {
  if (!canSaveCollectionChannel.value) return
  const form = collectionForm.value
  curatedChannels.value.push({
    id: `${Date.now()}`,
    displayName: form.displayName.trim(),
    serviceCode: form.serviceCode.trim(),
    recommendation: form.notes.trim() || '待补充人工测算依据',
    enabled: false,
  })
  collectionForm.value = {
    serviceCode: '',
    displayName: '',
    maxLength: '',
    maxWidth: '',
    maxHeight: '',
    volumeDivisor: '6000',
    maxWeight: '',
    notes: '',
  }
  collectionDialogOpen.value = false
}

</script>

<template>
      <section class="rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-violet-600">Single Outbound Port</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">4PX 精选服务集合</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">这是 4PX 独立域对主运费模板的唯一公开数据出口。仅发布经过人工比价复核的 3 至 5 条大件直发渠道，其他 TAB 数据保持域内私有。</p>
          </div>
          <Button size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:fpx:manage 权限'" @click="openCollectionDialog">
            <Plus class="size-3.5" />
            手动添加精选渠道
          </Button>
        </div>

        <div class="mt-5 grid gap-3 md:grid-cols-3">
          <article class="rounded-[22px] border border-dashed border-border/80 bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">已发布渠道</p>
            <p class="mt-2 text-2xl font-black">{{ curatedChannels.filter((channel) => channel.enabled).length }}</p>
            <p class="mt-1 text-[11px] text-muted-foreground">供 ShippingTemplate 只读使用</p>
          </article>
          <article class="rounded-[22px] border border-dashed border-border/80 bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">集合总量</p>
            <p class="mt-2 text-2xl font-black">{{ curatedChannels.length }}</p>
            <p class="mt-1 text-[11px] text-muted-foreground">建议控制在 3 至 5 条</p>
          </article>
          <article class="rounded-[22px] border border-emerald-500/30 bg-emerald-500/10 p-4">
            <p class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-emerald-700 dark:text-emerald-300"><CheckCircle2 class="size-3.5" /> 边界状态</p>
            <p class="mt-2 text-sm font-black">独立域出口已隔离</p>
            <p class="mt-1 text-[11px] text-emerald-700/80 dark:text-emerald-300/80">不暴露运单、轨迹与 RMA 私有数据</p>
          </article>
        </div>
      </section>

      <section class="overflow-hidden rounded-[28px] border border-dashed border-border/80 bg-card shadow-sm">
        <div class="flex flex-col gap-3 border-b border-dashed border-border/80 px-5 py-4 sm:flex-row sm:items-end sm:justify-between sm:px-6">
          <div>
            <h2 class="text-base font-black">精选渠道发布台账</h2>
            <p class="mt-1 text-xs text-muted-foreground">人工测算、人工维护、受控发布</p>
          </div>
          <label class="w-full space-y-1.5 sm:max-w-sm">
            <span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><Search class="size-3" /> 搜索渠道</span>
            <Input v-model="collectionSearch" placeholder="业务别名 / 官方服务代码 / 履约类型" />
          </label>
        </div>

        <div v-if="filteredCuratedChannels.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
          <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><PackageCheck class="size-5" /></span>
          <div>
            <p class="text-sm font-black">尚未发布 4PX 精选渠道</p>
            <p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">请根据官方大件报价单完成人工测算后添加。当前没有使用示例渠道冒充真实配置。</p>
          </div>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[980px] text-left text-sm">
            <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              <tr><th class="px-5 py-3">业务别名</th><th class="px-4 py-3">官方服务代码</th><th class="px-4 py-3">人工测算说明</th><th class="px-4 py-3">发布状态</th><th class="px-5 py-3 text-right">操作</th></tr>
            </thead>
            <tbody class="divide-y divide-dashed divide-border/70">
              <tr v-for="channel in filteredCuratedChannels" :key="channel.id">
                <td class="px-5 py-4 font-semibold">{{ channel.displayName }}</td>
                <td class="px-4 py-4 font-mono text-xs font-bold">{{ channel.serviceCode }}</td>
                <td class="max-w-sm px-4 py-4 text-xs leading-5 text-muted-foreground">{{ channel.recommendation }}</td>
                <td class="px-4 py-4"><span :class="channel.enabled ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'bg-muted text-muted-foreground'" class="rounded-full px-2.5 py-1 text-[10px] font-bold">{{ channel.enabled ? '已启用' : '临时停用' }}</span></td>
                <td class="px-5 py-4 text-right"><Button variant="ghost" size="sm" :disabled="!canManage">编辑</Button></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <Dialog v-model:open="collectionDialogOpen">
        <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>手动添加 4PX 精选渠道</DialogTitle>
            <DialogDescription>录入前请先完成官方报价的人工比价复核。新渠道默认不发布。</DialogDescription>
          </DialogHeader>
          <div class="grid gap-4 py-2 sm:grid-cols-2">
            <label class="space-y-1.5"><span class="text-xs font-bold">官方服务代码</span><Input v-model="collectionForm.serviceCode" placeholder="输入 4PX 返回的 service_code" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">业务别名</span><Input v-model="collectionForm.displayName" placeholder="如 4PX-美西大件专线" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">最大长度（cm）</span><Input v-model="collectionForm.maxLength" inputmode="decimal" placeholder="例如 102" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">最大宽度（cm）</span><Input v-model="collectionForm.maxWidth" inputmode="decimal" placeholder="例如 22" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">最大高度（cm）</span><Input v-model="collectionForm.maxHeight" inputmode="decimal" placeholder="例如 58" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">材积除数</span><Input v-model="collectionForm.volumeDivisor" inputmode="numeric" placeholder="6000" /></label>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">最大毛重（kg）</span><Input v-model="collectionForm.maxWeight" inputmode="decimal" placeholder="例如 30" /></label>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">测算说明与备注</span><textarea v-model="collectionForm.notes" rows="4" class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" placeholder="记录报价版本、目标市场、时效、费用对比和复核依据" /></label>
          </div>
          <DialogFooter>
            <Button variant="outline" @click="collectionDialogOpen = false">取消</Button>
            <Button :disabled="!canSaveCollectionChannel" @click="saveCollectionChannel">保存为未发布渠道</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
</template>
