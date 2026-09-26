<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, AlertTriangle, Calculator, CheckCircle2, ClipboardCheck, Download, Globe2, PackageCheck, Plus, Power, Printer, RadioTower, RefreshCw, Search, Send, Settings2, Trash2, Truck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:yanwen:manage'))

const collectionSearch = ref('')
const collectionStatus = ref('')
const collectionDialogOpen = ref(false)
const collectionActionMessage = ref('')
const collectionForm = ref({
  productCode: '',
  displayName: '',
  packageType: '普货配件',
  maxWeight: '2000',
  volumeDivisor: '8000',
  notes: '',
})
const curatedChannels = ref<Array<{
  id: string
  productCode: string
  displayName: string
  packageType: string
  validity: string
  costHint: string
  enabled: boolean
}>>([])
const filteredCuratedChannels = computed(() => {
  const query = collectionSearch.value.trim().toLowerCase()
  return curatedChannels.value.filter((channel) => {
    const matchesQuery = !query || [channel.productCode, channel.displayName, channel.packageType, channel.costHint]
      .some((value) => value.toLowerCase().includes(query))
    const matchesStatus = !collectionStatus.value
      || (collectionStatus.value === 'enabled' ? channel.enabled : !channel.enabled)
    return matchesQuery && matchesStatus
  })
})
const canSaveCollection = computed(() => canManage.value && collectionForm.value.productCode.trim().length > 0 && collectionForm.value.displayName.trim().length > 0)
const openCollectionDialog = () => {
  collectionDialogOpen.value = true
  collectionActionMessage.value = ''
}
const saveCollectionChannel = () => {
  if (!canSaveCollection.value) return
  const form = collectionForm.value
  curatedChannels.value.unshift({
    id: `YW-${Date.now()}`,
    productCode: form.productCode.trim(),
    displayName: form.displayName.trim(),
    packageType: form.packageType,
    validity: '待人工测算',
    costHint: form.notes.trim() || '待补充官方报价与时效依据',
    enabled: false,
  })
  collectionForm.value = { productCode: '', displayName: '', packageType: '普货配件', maxWeight: '2000', volumeDivisor: '8000', notes: '' }
  collectionDialogOpen.value = false
  collectionActionMessage.value = '渠道已保存为停用草稿，完成人工测算和复核后再发布给主运费模板。'
}
const toggleCuratedChannel = (channel: (typeof curatedChannels.value)[number]) => {
  if (!canManage.value) return
  channel.enabled = !channel.enabled
  collectionActionMessage.value = channel.enabled ? '渠道已启用并纳入精选集合。' : '渠道已停用，主运费模板将不再读取该渠道。'
}
const removeCuratedChannel = (id: string) => {
  if (!canManage.value) return
  curatedChannels.value = curatedChannels.value.filter((channel) => channel.id !== id)
  collectionActionMessage.value = '渠道已从精选集合移除。'
}
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Yanwen Curated Channels Roster</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">精选开通渠道白名单</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">燕文域唯一对外公开数据出口。主运费模板只能读取这里已启用的精选渠道，燕文其他 TAB 的内部数据不会旁路暴露。</p>
          </div>
          <Button size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:yanwen:manage 权限'" @click="openCollectionDialog"><Plus class="size-3.5" />手动添加渠道</Button>
        </div>

        <div class="grid gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 md:grid-cols-[minmax(0,1fr)_220px_auto]">
          <label class="space-y-1.5"><span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground/80"><Search class="size-3" />搜索渠道</span><Input v-model="collectionSearch" placeholder="业务别名 / 官方产品代码 / 测算备注" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">发布状态</span><select v-model="collectionStatus" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="">全部状态</option><option value="enabled">已启用</option><option value="disabled">已停用</option></select></label>
          <div class="flex items-end"><span class="rounded-full border border-border bg-background px-3 py-1.5 text-[11px] font-bold text-muted-foreground">{{ curatedChannels.filter((channel) => channel.enabled).length }} 条对外发布</span></div>
        </div>

        <p v-if="collectionActionMessage" class="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-xs leading-5 text-emerald-700 dark:text-emerald-300">{{ collectionActionMessage }}</p>

        <section class="overflow-hidden rounded-[24px] border border-dashed border-border/80">
          <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/80 px-5 py-4">
            <div><h3 class="text-sm font-black">燕文精选渠道清单</h3><p class="mt-1 text-[11px] text-muted-foreground">{{ filteredCuratedChannels.length }} 条记录 · 新增渠道默认停用，需人工复核后发布</p></div>
            <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">只读对外契约：YanwenCuratedCollection</span>
          </div>
          <div v-if="filteredCuratedChannels.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
            <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><PackageCheck class="size-5" /></span>
            <div><p class="text-sm font-black">暂无精选燕文渠道</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">请根据官方报价单或运价试算结果手动添加 3-5 条优势渠道。未通过人工复核前，不会向主运费模板发布。</p></div>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[980px] text-left text-sm">
              <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><tr><th class="px-5 py-3">业务别名 / 官方代码</th><th class="px-4 py-3">适用类型</th><th class="px-4 py-3">测算参考时效</th><th class="px-4 py-3">人工测算说明</th><th class="px-4 py-3">状态</th><th class="px-5 py-3 text-right">操作</th></tr></thead>
              <tbody class="divide-y divide-dashed divide-border/70"><tr v-for="channel in filteredCuratedChannels" :key="channel.id"><td class="px-5 py-4"><p class="font-semibold">{{ channel.displayName }}</p><p class="mt-1 font-mono text-[11px] text-muted-foreground">{{ channel.productCode }}</p></td><td class="px-4 py-4">{{ channel.packageType }}</td><td class="px-4 py-4 text-muted-foreground">{{ channel.validity }}</td><td class="max-w-sm px-4 py-4 text-xs leading-5 text-muted-foreground">{{ channel.costHint }}</td><td class="px-4 py-4"><span :class="channel.enabled ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'bg-muted text-muted-foreground'" class="rounded-full px-2.5 py-1 text-[10px] font-bold">{{ channel.enabled ? '已启用' : '已停用' }}</span></td><td class="px-5 py-4 text-right"><div class="flex justify-end gap-1"><Button variant="ghost" size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:yanwen:manage 权限'" @click="toggleCuratedChannel(channel)"><Power class="size-3.5" />{{ channel.enabled ? '停用' : '启用' }}</Button><Button variant="ghost" size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:yanwen:manage 权限'" @click="removeCuratedChannel(channel.id)"><Trash2 class="size-3.5" />移除</Button></div></td></tr></tbody>
            </table>
          </div>
        </section>
      </section>

      <Dialog v-model:open="collectionDialogOpen">
        <DialogContent size="lg">
          <DialogHeader><DialogTitle>手动添加燕文精选渠道</DialogTitle><DialogDescription>录入人工比价后的官方产品代码与业务别名。保存后默认停用，避免未经复核的渠道进入主运费模板。</DialogDescription></DialogHeader>
          <div class="grid gap-4 py-2 sm:grid-cols-2">
            <label class="space-y-1.5"><span class="text-xs font-bold">官方产品 ID</span><Input v-model="collectionForm.productCode" placeholder="来自 express.channel.getlist" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">业务别名</span><Input v-model="collectionForm.displayName" placeholder="例如 燕文-欧美配件专线" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">允许货品属性</span><select v-model="collectionForm.packageType" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="普货配件">普货配件</option><option value="特货小包">特货小包</option><option value="带电货品">带电货品</option><option value="敏感货">敏感货</option></select></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">最大毛重（g）</span><Input v-model="collectionForm.maxWeight" inputmode="numeric" placeholder="2000" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">体积重系数</span><Input v-model="collectionForm.volumeDivisor" inputmode="numeric" placeholder="8000" /></label>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">测算说明与备注</span><textarea v-model="collectionForm.notes" rows="4" class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" placeholder="记录官方报价、预计时效、目的国和人工比价依据" /></label>
          </div>
          <DialogFooter><Button variant="outline" @click="collectionDialogOpen = false">取消</Button><Button :disabled="!canSaveCollection" @click="saveCollectionChannel">保存停用草稿</Button></DialogFooter>
        </DialogContent>
      </Dialog>
</template>
