<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { PackageCheck, Pencil, Plus, Power, Search, Trash2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAuthStore } from '@/stores/auth'
import yanwenLogisticsAdminApi, { type YanwenProductCatalogEntry, type YanwenPublishedChannel } from '@/api/yanwenLogisticsAdminApi'

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:yanwen:manage'))

const selectedEnvironment = ref<'fat' | 'production'>('production')
const collectionSearch = ref('')
const collectionStatus = ref('')
const collectionDialogOpen = ref(false)
const collectionActionMessage = ref('')
const collectionErrorMessage = ref('')
const collectionLoading = ref(false)
const editingCollectionChannelId = ref<number | null>(null)
const collectionForm = ref({
  productCode: '',
  displayName: '',
  countries: '',
  packageType: '普货配件',
  maxWeight: '2000',
  volumeDivisor: '8000',
  requireReceiverTaxNumber: false,
  requireIoss: false,
  requireEori: false,
  notes: '',
})
type YanwenCollectionChannel = {
  id: number
  environment: 'fat' | 'production'
  productCode: string
  displayName: string
  countries: string
  packageType: string
  notes: string
  maxWeight: number
  volumeDivisor: number
  requireReceiverTaxNumber: boolean
  requireIoss: boolean
  requireEori: boolean
  enabled: boolean
}
const curatedChannels = ref<YanwenCollectionChannel[]>([])
const officialProducts = ref<YanwenProductCatalogEntry[]>([])
const officialProductsLoading = ref(false)
const normalizeYanwenCollectionChannel = (channel: YanwenPublishedChannel): YanwenCollectionChannel => ({
  id: Number(channel.id),
  environment: channel.environment === 'fat' ? 'fat' : 'production',
  productCode: String(channel.product_code || '').trim(),
  displayName: String(channel.display_name || '').trim(),
  countries: String(channel.countries || '[]'),
  packageType: String(channel.package_type || '普货配件').trim(),
  notes: String(channel.notes || '').trim(),
  maxWeight: Number(channel.max_weight_grams || 0),
  volumeDivisor: Number(channel.volumetric_divisor || 8000),
  requireReceiverTaxNumber: channel.require_receiver_tax_number === true,
  requireIoss: channel.require_ioss === true,
  requireEori: channel.require_eori === true,
  enabled: channel.enabled === true,
})
const filteredCuratedChannels = computed(() => {
  const query = collectionSearch.value.trim().toLowerCase()
  return curatedChannels.value.filter((channel) => {
    const matchesQuery = !query || [channel.productCode, channel.displayName, channel.countries, channel.packageType, channel.notes]
      .some((value) => value.toLowerCase().includes(query))
    const matchesStatus = !collectionStatus.value
      || (collectionStatus.value === 'enabled' ? channel.enabled : !channel.enabled)
    return matchesQuery && matchesStatus
  })
})
const channelCountriesLabel = (value: unknown) => {
  const raw = String(value || '').trim()
  if (!raw) return '未限制'
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed) && parsed.length) return parsed.join(', ')
  } catch {
    // Keep compatibility with older comma-separated collection records.
  }
  return raw.replace(/[\[\]"']/g, '').replace(/[,，;|]+/g, ', ')
}
const selectedProductIsOfficial = computed(() => editingCollectionChannelId.value !== null
  || officialProducts.value.some((product) => product.product_id === collectionForm.value.productCode.trim()))
const canSaveCollection = computed(() => canManage.value
  && selectedProductIsOfficial.value
  && collectionForm.value.displayName.trim().length > 0)
const collectionDialogTitle = computed(() => editingCollectionChannelId.value === null ? '手动添加燕文精选渠道' : '编辑燕文精选渠道')
const collectionDialogDescription = computed(() => editingCollectionChannelId.value === null
  ? selectedEnvironment.value === 'production'
    ? '从生产环境官方产品目录选择产品。保存后默认停用，复核并启用后才会进入物流管理运费模板。'
    : '从 FAT 官方产品目录选择产品。该渠道仅供 FAT 环境真实运单使用，不会进入物流管理运费模板。'
  : '更新当前环境的官方产品代码、配送地区、包裹限制和人工复核备注；渠道来源环境与发布状态保持不变。')
const resetCollectionForm = () => {
  editingCollectionChannelId.value = null
  collectionForm.value = { productCode: '', displayName: '', countries: '', packageType: '普货配件', maxWeight: '2000', volumeDivisor: '8000', requireReceiverTaxNumber: false, requireIoss: false, requireEori: false, notes: '' }
}
const openCollectionDialog = () => {
  resetCollectionForm()
  collectionDialogOpen.value = true
  collectionActionMessage.value = ''
  collectionErrorMessage.value = ''
}
const openEditCollectionChannelDialog = (channel: YanwenCollectionChannel) => {
  editingCollectionChannelId.value = channel.id
  collectionForm.value = {
    productCode: channel.productCode,
    displayName: channel.displayName,
    countries: channelCountriesLabel(channel.countries),
    packageType: channel.packageType,
    maxWeight: String(channel.maxWeight || 0),
    volumeDivisor: String(channel.volumeDivisor || 8000),
    requireReceiverTaxNumber: channel.requireReceiverTaxNumber,
    requireIoss: channel.requireIoss,
    requireEori: channel.requireEori,
    notes: channel.notes,
  }
  collectionDialogOpen.value = true
  collectionActionMessage.value = ''
  collectionErrorMessage.value = ''
}
const loadCollectionChannels = async (environment: 'fat' | 'production' = selectedEnvironment.value) => {
  collectionLoading.value = true
  collectionErrorMessage.value = ''
  try {
    const channels = await yanwenLogisticsAdminApi.listYanwenChannels({ environment })
    if (environment === selectedEnvironment.value) {
      curatedChannels.value = channels.map(normalizeYanwenCollectionChannel)
    }
  } catch (error) {
    if (environment === selectedEnvironment.value) {
      collectionErrorMessage.value = error instanceof Error ? error.message : '加载燕文精选渠道失败'
    }
  } finally {
    if (environment === selectedEnvironment.value) collectionLoading.value = false
  }
}
const loadOfficialProducts = async (environment: 'fat' | 'production' = selectedEnvironment.value) => {
  officialProductsLoading.value = true
  try {
    const products = await yanwenLogisticsAdminApi.listYanwenProducts(environment)
    if (environment === selectedEnvironment.value) officialProducts.value = products
  } catch (error) {
    if (environment === selectedEnvironment.value) {
      collectionErrorMessage.value = error instanceof Error ? error.message : '加载燕文官方产品目录失败'
    }
  } finally {
    if (environment === selectedEnvironment.value) officialProductsLoading.value = false
  }
}
const saveCollectionChannel = async () => {
  if (!canSaveCollection.value) return
  const form = collectionForm.value
  const editingChannelId = editingCollectionChannelId.value
  const channelPayload = {
    environment: selectedEnvironment.value,
    product_code: form.productCode.trim(),
    display_name: form.displayName.trim(),
    countries: form.countries.trim(),
    package_type: form.packageType,
    max_weight_grams: Number(form.maxWeight || 0),
    volumetric_divisor: Number(form.volumeDivisor || 8000),
    require_receiver_tax_number: form.requireReceiverTaxNumber,
    require_ioss: form.requireIoss,
    require_eori: form.requireEori,
    notes: form.notes.trim(),
  }
  collectionErrorMessage.value = ''
  try {
    if (editingChannelId === null) {
      await yanwenLogisticsAdminApi.createYanwenChannel({ ...channelPayload, enabled: false })
      collectionActionMessage.value = '渠道已保存为停用草稿，完成官方产品确认和人工复核后再发布给主运费模板。'
    } else {
      const updated = await yanwenLogisticsAdminApi.updateYanwenChannel(editingChannelId, channelPayload)
      const normalizedUpdatedChannel = normalizeYanwenCollectionChannel(updated)
      const channelIndex = curatedChannels.value.findIndex((channel) => channel.id === editingChannelId)
      if (channelIndex >= 0) curatedChannels.value[channelIndex] = normalizedUpdatedChannel
      collectionActionMessage.value = '渠道信息已更新，当前发布状态保持不变。'
    }
    resetCollectionForm()
    collectionDialogOpen.value = false
    if (editingChannelId === null) await loadCollectionChannels()
  } catch (error) {
    collectionErrorMessage.value = error instanceof Error ? error.message : editingChannelId === null ? '保存燕文精选渠道失败' : '更新燕文精选渠道失败'
  }
}
const toggleCuratedChannel = async (channel: (typeof curatedChannels.value)[number]) => {
  if (!canManage.value) return
  collectionErrorMessage.value = ''
  try {
    const updated = await yanwenLogisticsAdminApi.updateYanwenChannel(channel.id, { enabled: !channel.enabled })
    Object.assign(channel, normalizeYanwenCollectionChannel(updated))
    collectionActionMessage.value = channel.enabled
      ? channel.environment === 'production' ? '生产渠道已启用并纳入主运费模板可读取的精选集合。' : 'FAT 渠道已启用，仅供 FAT 环境真实运单使用。'
      : channel.environment === 'production' ? '生产渠道已停用，主运费模板将不再读取该渠道。' : 'FAT 渠道已停用，FAT 环境真实建单将不再读取该渠道。'
  } catch (error) {
    collectionErrorMessage.value = error instanceof Error ? error.message : '更新燕文渠道状态失败'
  }
}
const removeCuratedChannel = async (id: number) => {
  if (!canManage.value) return
  collectionErrorMessage.value = ''
  try {
    await yanwenLogisticsAdminApi.deleteYanwenChannel(id)
    curatedChannels.value = curatedChannels.value.filter((channel) => channel.id !== id)
    collectionActionMessage.value = '渠道已从精选集合移除。'
  } catch (error) {
    collectionErrorMessage.value = error instanceof Error ? error.message : '移除燕文渠道失败'
  }
}

const loadCollectionPage = async (environment: 'fat' | 'production' = selectedEnvironment.value) => {
  await Promise.all([loadCollectionChannels(environment), loadOfficialProducts(environment)])
}

onMounted(loadCollectionPage)
watch(selectedEnvironment, (environment) => { void loadCollectionPage(environment) })
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Yanwen Curated Channels Roster</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">精选开通渠道白名单</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">按环境维护燕文精选渠道。只有生产环境已启用渠道进入物流管理运费模板与前台线路；FAT 渠道仅供 FAT 环境内真实运单使用。</p>
          </div>
          <Button size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:yanwen:manage 权限'" @click="openCollectionDialog"><Plus class="size-3.5" />手动添加渠道</Button>
        </div>

        <div class="grid gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 md:grid-cols-[minmax(0,1fr)_180px_180px_auto]">
            <label class="space-y-1.5"><span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground/80"><Search class="size-3" />搜索渠道</span><Input v-model="collectionSearch" placeholder="业务名称 / 官方产品代码 / 官方依据备注" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">来源环境</span><select v-model="selectedEnvironment" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="production">生产环境 PRD</option><option value="fat">测试环境 FAT</option></select></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">发布状态</span><select v-model="collectionStatus" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="">全部状态</option><option value="enabled">已启用</option><option value="disabled">已停用</option></select></label>
          <div class="flex items-end"><span class="rounded-full border border-border bg-background px-3 py-1.5 text-[11px] font-bold text-muted-foreground">{{ curatedChannels.filter((channel) => channel.enabled).length }} 条{{ selectedEnvironment === 'production' ? '对外发布' : 'FAT 可用' }}</span></div>
        </div>

        <p v-if="collectionActionMessage" class="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-xs leading-5 text-emerald-700 dark:text-emerald-300">{{ collectionActionMessage }}</p>
        <p v-if="collectionErrorMessage" class="rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-xs leading-5 text-red-700 dark:text-red-300">{{ collectionErrorMessage }}</p>

        <section class="overflow-hidden rounded-[24px] border border-dashed border-border/80">
          <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/80 px-5 py-4">
            <div><h3 class="text-sm font-black">燕文精选渠道清单</h3><p class="mt-1 text-[11px] text-muted-foreground">{{ filteredCuratedChannels.length }} 条 {{ selectedEnvironment === 'production' ? '生产环境' : 'FAT 环境' }}记录 · 新增渠道默认停用</p></div>
            <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">按 environment + product_code 隔离</span>
          </div>
          <div v-if="collectionLoading" class="flex min-h-64 items-center justify-center px-6 py-12 text-sm text-muted-foreground">正在加载燕文精选渠道…</div>
          <div v-else-if="filteredCuratedChannels.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
            <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><PackageCheck class="size-5" /></span>
            <div><p class="text-sm font-black">暂无精选燕文渠道</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">请根据燕文官网报价单和人工复核依据手动添加 3-5 条优势渠道。未通过人工复核前，不会向主运费模板发布。</p></div>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[1320px] text-left text-sm">
               <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><tr><th class="px-5 py-3">业务别名 / 官方代码</th><th class="px-4 py-3">配送地区</th><th class="px-4 py-3">适用类型</th><th class="px-4 py-3">建单必填字段</th><th class="px-4 py-3">官方依据与复核备注</th><th class="px-4 py-3">状态</th><th class="px-5 py-3 text-right">操作</th></tr></thead>
               <tbody class="divide-y divide-dashed divide-border/70"><tr v-for="channel in filteredCuratedChannels" :key="channel.id"><td class="px-5 py-4"><p class="font-semibold">{{ channel.displayName }}</p><p class="mt-1 font-mono text-[11px] text-muted-foreground">{{ channel.productCode }}</p><p class="mt-1 text-[10px] font-bold text-muted-foreground">{{ channel.environment === 'production' ? '生产 PRD' : '测试 FAT' }}</p></td><td class="px-4 py-4 text-xs text-muted-foreground">{{ channelCountriesLabel(channel.countries) }}</td><td class="px-4 py-4">{{ channel.packageType }}</td><td class="px-4 py-4 text-xs text-muted-foreground">{{ [channel.requireReceiverTaxNumber ? '收件人税号' : '', channel.requireIoss ? 'IOSS' : '', channel.requireEori ? 'EORI' : ''].filter(Boolean).join('、') || '无额外必填' }}</td><td class="max-w-sm px-4 py-4 text-xs leading-5 text-muted-foreground">{{ channel.notes || '未填写官方依据备注' }}</td><td class="px-4 py-4"><span :class="channel.enabled ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'bg-muted text-muted-foreground'" class="rounded-full px-2.5 py-1 text-[10px] font-bold">{{ channel.enabled ? '已启用' : '已停用' }}</span></td><td class="px-5 py-4 text-right"><div class="flex justify-end gap-1"><Button variant="ghost" size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:yanwen:manage 权限'" @click="openEditCollectionChannelDialog(channel)"><Pencil class="size-3.5" />编辑</Button><Button variant="ghost" size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:yanwen:manage 权限'" @click="toggleCuratedChannel(channel)"><Power class="size-3.5" />{{ channel.enabled ? '停用' : '启用' }}</Button><Button variant="ghost" size="sm" :disabled="!canManage" :title="canManage ? undefined : '需要 logistics:yanwen:manage 权限'" @click="removeCuratedChannel(channel.id)"><Trash2 class="size-3.5" />移除</Button></div></td></tr></tbody>
            </table>
          </div>
        </section>
      </section>

      <Dialog v-model:open="collectionDialogOpen">
        <DialogContent size="lg">
          <DialogHeader><DialogTitle>{{ collectionDialogTitle }}</DialogTitle><DialogDescription>{{ collectionDialogDescription }}</DialogDescription></DialogHeader>
          <div class="grid gap-4 py-2 sm:grid-cols-2">
            <label class="space-y-1.5"><span class="text-xs font-bold">官方产品 ID</span><select v-model="collectionForm.productCode" :disabled="officialProductsLoading" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">选择已同步的官方产品</option><option v-if="editingCollectionChannelId !== null && collectionForm.productCode && !officialProducts.some((product) => product.product_id === collectionForm.productCode)" :value="collectionForm.productCode">当前记录：{{ collectionForm.productCode }}</option><option v-for="product in officialProducts" :key="product.product_id" :value="product.product_id">{{ product.product_id }} · {{ product.name_ch || product.name_en }}</option></select><p v-if="officialProducts.length === 0" class="text-[11px] leading-4 text-muted-foreground">请先在网关配置中同步 `express.channel.getlist`，此处不接受手工填写产品代码。</p></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">业务别名</span><Input v-model="collectionForm.displayName" placeholder="例如 燕文-欧美配件专线" /></label>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">配送国家/地区</span><Input v-model="collectionForm.countries" placeholder="例如 US, CA, GB；留空表示服务集合未提供地区限制" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">允许货品属性</span><select v-model="collectionForm.packageType" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="普货配件">普货配件</option><option value="特货小包">特货小包</option><option value="带电货品">带电货品</option><option value="敏感货">敏感货</option></select></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">最大毛重（g）</span><Input v-model="collectionForm.maxWeight" inputmode="numeric" placeholder="2000" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">体积重系数</span><Input v-model="collectionForm.volumeDivisor" inputmode="numeric" placeholder="8000" /></label>
            <fieldset class="space-y-2 rounded-xl border border-border p-3 sm:col-span-2"><legend class="px-1 text-xs font-bold">燕文建单字段必填规则</legend><p class="text-[11px] leading-5 text-muted-foreground">只按此渠道人工确认的规则阻断缺失字段；不会按国家推断，也不代表官方号码有效性校验。</p><div class="grid gap-2 sm:grid-cols-3"><label class="flex items-center gap-2 text-xs"><input v-model="collectionForm.requireReceiverTaxNumber" type="checkbox" class="size-4 rounded border-border" />收件人税号必填</label><label class="flex items-center gap-2 text-xs"><input v-model="collectionForm.requireIoss" type="checkbox" class="size-4 rounded border-border" />IOSS 必填</label><label class="flex items-center gap-2 text-xs"><input v-model="collectionForm.requireEori" type="checkbox" class="size-4 rounded border-border" />EORI 必填</label></div></fieldset>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">官方依据与复核备注</span><textarea v-model="collectionForm.notes" rows="4" class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" placeholder="记录官网产品说明、目的国和人工复核依据" /></label>
          </div>
          <DialogFooter><Button variant="outline" @click="collectionDialogOpen = false">取消</Button><Button :disabled="!canSaveCollection" @click="saveCollectionChannel">{{ editingCollectionChannelId === null ? '保存停用草稿' : '保存更新' }}</Button></DialogFooter>
        </DialogContent>
      </Dialog>
</template>
