import { computed, onMounted, ref, watch } from 'vue'
import { yanwenLogisticsAdminApi, type YanwenPublishedChannel, type YanwenWaybill, type YanwenWaybillCreationRequest, type YanwenWarehouseCatalogEntry } from '@/api/yanwenLogisticsAdminApi'
import { useAuthStore } from '@/stores/auth'

export const useYanwenWaybillOperations = () => {
const authStore = useAuthStore()
const waybillSearch = ref('')
const waybillEnvironment = ref<'fat' | 'production'>('fat')
const waybillChannel = ref('')
const waybillWarehouse = ref('')
const waybillStatus = ref('')
const waybillRefreshing = ref(false)
const waybillSyncingId = ref<string | null>(null)
const waybillBatchSyncing = ref(false)
const waybillBatchCancelling = ref(false)
const waybillBatchLabelDownloading = ref(false)
const waybillLabelDownloadingId = ref<string | null>(null)
const waybillCancellingId = ref<string | null>(null)
const waybillSelectedIds = ref<string[]>([])
const waybillActionMessage = ref('')
const createWaybillDialogOpen = ref(false)
const createWaybillSaving = ref(false)
const createWaybillForm = ref({ orderId: '', channelId: '', warehouseCode: '', hasBattery: '', receiverTaxNumber: '', ioss: '', eori: '' })
const batchCreateWaybillDialogOpen = ref(false)
const batchCreateWaybillSaving = ref(false)
type YanwenBatchWaybillCreationFormRow = {
  orderId: string
  hasBattery: '' | 'true' | 'false'
  receiverTaxNumber: string
  ioss: string
  eori: string
}
const batchCreateWaybillForm = ref({ channelId: '', warehouseCode: '' })
const batchCreateWaybillRows = ref<YanwenBatchWaybillCreationFormRow[]>([])
const curatedChannels = ref<YanwenPublishedChannel[]>([])
const warehouses = ref<YanwenWarehouseCatalogEntry[]>([])
const selectedYanwenChannel = computed(() => curatedChannels.value.find((channel) => String(channel.id) === createWaybillForm.value.channelId) || null)
const selectedBatchYanwenChannel = computed(() => curatedChannels.value.find((channel) => String(channel.id) === batchCreateWaybillForm.value.channelId) || null)
const selectedYanwenChannelRequiredFields = computed(() => {
  const channel = selectedYanwenChannel.value
  if (!channel) return []
  return [
    channel.require_receiver_tax_number ? '收件人税号' : '',
    channel.require_ioss ? 'IOSS' : '',
    channel.require_eori ? 'EORI' : '',
  ].filter(Boolean)
})
const selectedBatchYanwenChannelRequiredFields = computed(() => {
  const channel = selectedBatchYanwenChannel.value
  if (!channel) return []
  return [
    channel.require_receiver_tax_number ? '收件人税号' : '',
    channel.require_ioss ? 'IOSS' : '',
    channel.require_eori ? 'EORI' : '',
  ].filter(Boolean)
})
const canShip = computed(() => authStore.hasPermission('logistics:yanwen:ship'))
const canCancel = computed(() => authStore.hasPermission('logistics:yanwen:cancel'))
const yanwenOfficialStatusLabels: Record<number, string> = {
  0: '已制单',
  1: '已确认发货',
  2: '已收货',
  3: '运输中',
  4: '已妥投',
  5: '已取消',
  6: '已截留',
  7: '投递失败',
  8: '仓内异常',
  9: '仓内退件',
  10: '退件签收',
  11: '转运异常',
  12: '派送中',
  13: '待提取',
  15: '追踪结束',
}
type WaybillRow = {
  id: string
  orderNo: string
  createdAt: string
  waybill: string
  lastMile: string
  channel: string
  warehouse: string
  destination: string
  consignee: string
  description: string
  weight: string
  status: string
  officialStatusLabel: string
  referenceNumber: string
  isPrinted: boolean
  officialSyncedAt: string
}
const waybillRecords = ref<WaybillRow[]>([])
const filteredWaybills = computed(() => {
  const query = waybillSearch.value.trim().toLowerCase()
  return waybillRecords.value.filter((record) => {
    const matchesSearch = !query || [record.orderNo, record.waybill, record.referenceNumber, record.lastMile, record.consignee, record.description]
      .some((value) => value.toLowerCase().includes(query))
    const matchesChannel = !waybillChannel.value || record.channel === waybillChannel.value
    const matchesWarehouse = !waybillWarehouse.value || record.warehouse === waybillWarehouse.value
    const matchesStatus = !waybillStatus.value || record.status === waybillStatus.value
    return matchesSearch && matchesChannel && matchesWarehouse && matchesStatus
  })
})
const waybillChannelOptions = computed(() => Array.from(new Set(waybillRecords.value.map((record) => record.channel).filter(Boolean))))
const hasSelectedWaybills = computed(() => waybillSelectedIds.value.length > 0)
const refreshWaybills = async () => {
  waybillRefreshing.value = true
  waybillActionMessage.value = ''
  try {
    const records = await yanwenLogisticsAdminApi.listYanwenWaybills({
      environment: waybillEnvironment.value,
      keyword: waybillSearch.value.trim() || undefined,
      status: waybillStatus.value === '已建单' ? 'created' : undefined,
      warehouse_code: waybillWarehouse.value.trim() || undefined,
    })
    waybillRecords.value = records.map((record: YanwenWaybill) => ({
      id: String(record.id),
      orderNo: record.order_number,
      createdAt: record.created_at,
      waybill: record.waybill_number,
      lastMile: record.yanwen_order_number,
      referenceNumber: record.reference_number,
      channel: record.channel_name || record.product_code,
      warehouse: record.warehouse_code,
      destination: record.destination_country,
      consignee: record.consignee_name,
      description: record.declared_description,
      weight: `${record.total_weight_grams} g · ${record.total_quantity} 件`,
      status: record.status === 'created' ? '已建单' : record.status,
      officialStatusLabel: record.last_official_synced_at
        ? (yanwenOfficialStatusLabels[record.official_status] || `未知状态（${record.official_status}）`)
        : '未同步官方状态',
      isPrinted: record.is_printed,
      officialSyncedAt: record.last_official_synced_at || '',
    }))
    waybillSelectedIds.value = []
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '读取燕文真实运单失败。'
  } finally {
    waybillRefreshing.value = false
  }
}
const resetWaybillFilters = async () => {
  waybillSearch.value = ''
  waybillChannel.value = ''
  waybillWarehouse.value = ''
  waybillStatus.value = ''
  await refreshWaybills()
}
const toggleWaybill = (id: string) => {
  waybillSelectedIds.value = waybillSelectedIds.value.includes(id)
    ? waybillSelectedIds.value.filter((selectedId) => selectedId !== id)
    : [...waybillSelectedIds.value, id]
}
const downloadSelectedWaybillLabels = async () => {
  if (!canShip.value || !hasSelectedWaybills.value || waybillBatchLabelDownloading.value) return
  waybillBatchLabelDownloading.value = true
  waybillActionMessage.value = ''
  try {
    const result = await yanwenLogisticsAdminApi.downloadYanwenWaybillLabels(waybillSelectedIds.value.map((id) => Number(id)))
    const downloadUrl = URL.createObjectURL(result.blob)
    const downloadLink = document.createElement('a')
    downloadLink.href = downloadUrl
    downloadLink.download = 'yanwen-waybill-labels.zip'
    document.body.appendChild(downloadLink)
    downloadLink.click()
    downloadLink.remove()
    window.setTimeout(() => URL.revokeObjectURL(downloadUrl), 0)
    await refreshWaybills()
    waybillActionMessage.value = result.failed > 0
      ? `批量面单已下载：成功 ${result.succeeded} 份，失败 ${result.failed} 份。失败明细已写入压缩包 manifest。`
      : `已下载 ${result.succeeded} 份燕文官方 PDF 面单。`
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '下载燕文批量官方面单失败。'
  } finally {
    waybillBatchLabelDownloading.value = false
  }
}
const cancelWaybill = async (id: string) => {
  if (!canCancel.value || waybillCancellingId.value || waybillBatchCancelling.value) return
  waybillCancellingId.value = id
  waybillActionMessage.value = ''
  try {
    const result = await yanwenLogisticsAdminApi.cancelYanwenWaybill(id)
    await refreshWaybills()
    waybillActionMessage.value = result.message || `运单 ${id} 的燕文取消请求已提交。`
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '提交燕文官方取消请求失败。'
  } finally {
    waybillCancellingId.value = null
  }
}
const cancelSelectedWaybills = async () => {
  if (!canCancel.value || !hasSelectedWaybills.value || waybillBatchCancelling.value) return
  waybillBatchCancelling.value = true
  waybillActionMessage.value = ''
  try {
    const result = await yanwenLogisticsAdminApi.cancelYanwenWaybills(waybillSelectedIds.value.map((id) => Number(id)))
    await refreshWaybills()
    const firstFailure = result.items.find((item) => item.error)?.error
    waybillActionMessage.value = firstFailure
      ? `批量取消完成：成功 ${result.succeeded} 条，失败 ${result.failed} 条。${firstFailure}`
      : `已提交 ${result.succeeded} 条燕文官方取消请求，并逐票回查官方状态。`
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '批量提交燕文官方取消请求失败。'
  } finally {
    waybillBatchCancelling.value = false
  }
}
const syncWaybillOfficialDetails = async (id: string) => {
  if (!canShip.value || waybillSyncingId.value) return
  waybillSyncingId.value = id
  waybillActionMessage.value = ''
  try {
    await yanwenLogisticsAdminApi.syncYanwenWaybillOfficialDetails(id)
    await refreshWaybills()
    waybillActionMessage.value = `运单 ${id} 已同步燕文官方状态。`
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '同步燕文官方运单详情失败。'
  } finally {
    waybillSyncingId.value = null
  }
}

const syncSelectedWaybillsOfficialDetails = async () => {
  if (!canShip.value || !hasSelectedWaybills.value || waybillBatchSyncing.value) return
  waybillBatchSyncing.value = true
  waybillActionMessage.value = ''
  try {
    const syncedWaybills = await yanwenLogisticsAdminApi.syncYanwenWaybillsOfficialDetails(waybillSelectedIds.value.map((id) => Number(id)))
    await refreshWaybills()
    waybillActionMessage.value = `已通过 express.order.getlist 同步 ${syncedWaybills.length} 条燕文官方详情。`
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '批量同步燕文官方运单详情失败。'
  } finally {
    waybillBatchSyncing.value = false
  }
}

const downloadWaybillLabel = async (id: string) => {
  if (!canShip.value || waybillLabelDownloadingId.value || waybillBatchLabelDownloading.value) return
  waybillLabelDownloadingId.value = id
  waybillActionMessage.value = ''
  try {
    const label = await yanwenLogisticsAdminApi.getYanwenWaybillLabel(id)
    const binaryContent = window.atob(label.base64_string)
    const pdfBytes = Uint8Array.from(binaryContent, (character) => character.charCodeAt(0))
    const pdfBlob = new Blob([pdfBytes], { type: label.content_type || 'application/pdf' })
    const downloadUrl = URL.createObjectURL(pdfBlob)
    const downloadLink = document.createElement('a')
    downloadLink.href = downloadUrl
    downloadLink.download = label.file_name || `yanwen-${label.waybill_number}.pdf`
    document.body.appendChild(downloadLink)
    downloadLink.click()
    downloadLink.remove()
    window.setTimeout(() => URL.revokeObjectURL(downloadUrl), 0)
    waybillActionMessage.value = `运单 ${label.waybill_number} 的真实面单已下载。`
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '下载燕文官方面单失败。'
  } finally {
    waybillLabelDownloadingId.value = null
  }
}

const loadYanwenWaybillCreationReferences = async () => {
  const [channels, officialWarehouses] = await Promise.all([
    yanwenLogisticsAdminApi.listYanwenChannels({ environment: waybillEnvironment.value, enabled: true }),
    yanwenLogisticsAdminApi.listYanwenWarehouses(waybillEnvironment.value),
  ])
  curatedChannels.value = channels as YanwenPublishedChannel[]
  warehouses.value = officialWarehouses
}

const createEmptyYanwenBatchWaybillCreationFormRow = (): YanwenBatchWaybillCreationFormRow => ({
  orderId: '',
  hasBattery: '',
  receiverTaxNumber: '',
  ioss: '',
  eori: '',
})

const openCreateWaybillDialog = async () => {
  if (!canShip.value) return
  waybillActionMessage.value = ''
  try {
    await loadYanwenWaybillCreationReferences()
    createWaybillDialogOpen.value = true
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '读取燕文渠道或交货仓失败。'
  }
}

const openBatchCreateWaybillDialog = async () => {
  if (!canShip.value) return
  waybillActionMessage.value = ''
  try {
    await loadYanwenWaybillCreationReferences()
    batchCreateWaybillForm.value = { channelId: '', warehouseCode: '' }
    batchCreateWaybillRows.value = [createEmptyYanwenBatchWaybillCreationFormRow()]
    batchCreateWaybillDialogOpen.value = true
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '读取燕文渠道或交货仓失败。'
  }
}

const addYanwenBatchWaybillCreationFormRow = () => {
  if (batchCreateWaybillRows.value.length >= 50) return
  batchCreateWaybillRows.value.push(createEmptyYanwenBatchWaybillCreationFormRow())
}

const removeYanwenBatchWaybillCreationFormRow = (rowIndex: number) => {
  if (batchCreateWaybillRows.value.length <= 1) return
  batchCreateWaybillRows.value.splice(rowIndex, 1)
}

const createYanwenBatchWaybills = async () => {
  if (!canShip.value || batchCreateWaybillSaving.value) return
  const channel = selectedBatchYanwenChannel.value
  if (!batchCreateWaybillForm.value.channelId || !batchCreateWaybillForm.value.warehouseCode || !channel) {
    waybillActionMessage.value = '批量推单必须选择已启用燕文渠道和官方交货仓。'
    return
  }
  if (batchCreateWaybillRows.value.length === 0 || batchCreateWaybillRows.value.length > 50) {
    waybillActionMessage.value = '批量推单每次需要提交 1 至 50 条记录。'
    return
  }
  const requests: YanwenWaybillCreationRequest[] = []
  for (const [rowIndex, row] of batchCreateWaybillRows.value.entries()) {
    if (!row.orderId.trim() || row.hasBattery === '') {
      waybillActionMessage.value = `第 ${rowIndex + 1} 行必须填写订单 ID 和带电标记。`
      return
    }
    const orderID = Number(row.orderId)
    if (!Number.isInteger(orderID) || orderID <= 0) {
      waybillActionMessage.value = `第 ${rowIndex + 1} 行的订单 ID 必须是正整数。`
      return
    }
    const missingRequiredFields = [
      channel.require_receiver_tax_number && !row.receiverTaxNumber.trim() ? '收件人税号' : '',
      channel.require_ioss && !row.ioss.trim() ? 'IOSS' : '',
      channel.require_eori && !row.eori.trim() ? 'EORI' : '',
    ].filter(Boolean)
    if (missingRequiredFields.length > 0) {
      waybillActionMessage.value = `第 ${rowIndex + 1} 行需要填写：${missingRequiredFields.join('、')}。`
      return
    }
    requests.push({
      environment: waybillEnvironment.value,
      order_id: orderID,
      channel_id: Number(batchCreateWaybillForm.value.channelId),
      warehouse_code: batchCreateWaybillForm.value.warehouseCode,
      has_battery: row.hasBattery === 'true',
      receiver_tax_number: row.receiverTaxNumber.trim(),
      ioss: row.ioss.trim(),
      eori: row.eori.trim(),
    })
  }
  batchCreateWaybillSaving.value = true
  waybillActionMessage.value = ''
  try {
    const result = await yanwenLogisticsAdminApi.createYanwenWaybills(requests)
    batchCreateWaybillDialogOpen.value = false
    batchCreateWaybillRows.value = []
    await refreshWaybills()
    const firstFailure = result.items.find((item) => item.error)?.error
    waybillActionMessage.value = firstFailure
      ? `批量推单完成：成功 ${result.succeeded} 条，失败 ${result.failed} 条。${firstFailure}`
      : `已完成 ${result.succeeded} 条燕文批量推单。`
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '燕文批量推单失败。'
  } finally {
    batchCreateWaybillSaving.value = false
  }
}

const createWaybill = async () => {
  if (!canShip.value || createWaybillSaving.value) return
  if (!createWaybillForm.value.orderId || !createWaybillForm.value.channelId || !createWaybillForm.value.warehouseCode || createWaybillForm.value.hasBattery === '') {
    waybillActionMessage.value = '订单、已启用渠道、官方交货仓和带电标记均为必填。'
    return
  }
  const channel = selectedYanwenChannel.value
  const missingRequiredDeclarationFields = [
    channel?.require_receiver_tax_number && !createWaybillForm.value.receiverTaxNumber.trim() ? '收件人税号' : '',
    channel?.require_ioss && !createWaybillForm.value.ioss.trim() ? 'IOSS' : '',
    channel?.require_eori && !createWaybillForm.value.eori.trim() ? 'EORI' : '',
  ].filter(Boolean)
  if (missingRequiredDeclarationFields.length > 0) {
    waybillActionMessage.value = `当前燕文渠道要求填写：${missingRequiredDeclarationFields.join('、')}。`
    return
  }
  createWaybillSaving.value = true
  waybillActionMessage.value = ''
  try {
    await yanwenLogisticsAdminApi.createYanwenWaybill({
      environment: waybillEnvironment.value,
      order_id: Number(createWaybillForm.value.orderId),
      channel_id: Number(createWaybillForm.value.channelId),
      warehouse_code: createWaybillForm.value.warehouseCode,
      has_battery: createWaybillForm.value.hasBattery === 'true',
      receiver_tax_number: createWaybillForm.value.receiverTaxNumber.trim(),
      ioss: createWaybillForm.value.ioss.trim(),
      eori: createWaybillForm.value.eori.trim(),
    })
    createWaybillDialogOpen.value = false
    createWaybillForm.value = { orderId: '', channelId: '', warehouseCode: '', hasBattery: '', receiverTaxNumber: '', ioss: '', eori: '' }
    await refreshWaybills()
  } catch (error: any) {
    waybillActionMessage.value = error?.message || '燕文真实建单失败。'
  } finally {
    createWaybillSaving.value = false
  }
}

onMounted(refreshWaybills)
watch(waybillEnvironment, () => {
  waybillChannel.value = ''
  void refreshWaybills()
})

  return {
    waybillSearch,
    waybillEnvironment,
    waybillChannel,
    waybillWarehouse,
    waybillStatus,
    waybillRefreshing,
    waybillSyncingId,
    waybillBatchSyncing,
    waybillBatchCancelling,
    waybillBatchLabelDownloading,
    waybillLabelDownloadingId,
    waybillCancellingId,
    waybillSelectedIds,
    waybillActionMessage,
    createWaybillDialogOpen,
    createWaybillSaving,
    createWaybillForm,
    batchCreateWaybillDialogOpen,
    batchCreateWaybillSaving,
    batchCreateWaybillForm,
    batchCreateWaybillRows,
    curatedChannels,
    warehouses,
    selectedYanwenChannel,
    selectedBatchYanwenChannel,
    selectedYanwenChannelRequiredFields,
    selectedBatchYanwenChannelRequiredFields,
    canShip,
    canCancel,
    waybillRecords,
    filteredWaybills,
    waybillChannelOptions,
    hasSelectedWaybills,
    refreshWaybills,
    resetWaybillFilters,
    toggleWaybill,
    downloadSelectedWaybillLabels,
    cancelWaybill,
    cancelSelectedWaybills,
    syncWaybillOfficialDetails,
    syncSelectedWaybillsOfficialDetails,
    downloadWaybillLabel,
    openCreateWaybillDialog,
    openBatchCreateWaybillDialog,
    addYanwenBatchWaybillCreationFormRow,
    removeYanwenBatchWaybillCreationFormRow,
    createYanwenBatchWaybills,
    createWaybill,
  }
}
