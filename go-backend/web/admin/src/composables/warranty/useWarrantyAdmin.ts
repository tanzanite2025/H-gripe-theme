import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { validateUploadFile } from '@/lib/uploadSpecs'
import warrantyApi from '@/api/warranty'
import type { WarrantyStats } from '@/api/warranty'
import type {
  WarrantyClaim,
  WarrantyFilters,
  WarrantyID,
  WarrantyOrderItem,
  WarrantyPagination,
  WarrantyShipmentDraft,
  WarrantyShipmentRecord,
  WarrantyServiceRecord,
  WarrantyServiceRecordForm,
  WarrantyStatusUpdating
} from '@/modules/warranty/warrantyTypes'
import { useRouteTab } from '@/composables/useRouteTab'
import { useAuthStore } from '@/stores/auth'
import {
  CLAIM_STATUS_OPTIONS,
  WARRANTY_SERVICE_STATUS_OPTIONS,
  WARRANTY_SERVICE_TYPE_OPTIONS
} from '@/lib/warrantyPresentation'
import {
  Clock3,
  FileWarning,
  ShieldCheck
} from '@lucide/vue'

type WarrantyTab = 'shipments' | 'claims' | 'boundary'

interface WarrantyLoadingState {
  shipments: boolean
  shipmentDetail: boolean
  shipmentSaving: boolean
  shipmentUploading: boolean
  claims: boolean
  claimDetail: boolean
  claimResolution: boolean
  claimOrderItems: boolean
  claimOrderItemBinding: boolean
  serviceRecords: boolean
  serviceRecordCreating: boolean
}

export function useWarrantyAdmin() {
  const authStore = useAuthStore()
  const activeTab = useRouteTab<WarrantyTab>({
    defaultValue: 'shipments',
    values: ['shipments', 'claims', 'boundary'],
    routes: {
      shipments: 'WarrantyShipments',
      claims: 'WarrantyClaims',
      boundary: 'WarrantyBoundary',
    },
  })
  const refreshing = ref(false)
  const shipmentStats = ref<WarrantyStats>({})
  const shipments = ref<WarrantyShipmentRecord[]>([])
  const claims = ref<WarrantyClaim[]>([])
  const selectedShipment = ref<WarrantyShipmentRecord | null>(null)
  const shipmentDraft = reactive<WarrantyShipmentDraft>({
    shippingNote: '',
    shippingImages: [],
    productCodes: [],
    warrantyMonths: 12,
    warrantyStart: ''
  })
  const selectedClaim = ref<WarrantyClaim | null>(null)
  const claimResolutionDraft = ref('')
  const claimOrderItems = ref<WarrantyOrderItem[]>([])
  const serviceRecords = ref<WarrantyServiceRecord[]>([])
  const orderItemSelection = ref('none')
  const serviceRecordForm = reactive<WarrantyServiceRecordForm>({
    serviceType: 'inspection',
    status: 'open',
    summary: '',
    costAmount: '',
    currency: '',
    performedAt: ''
  })

  const loading = reactive<WarrantyLoadingState>({
    shipments: false,
    shipmentDetail: false,
    shipmentSaving: false,
    shipmentUploading: false,
    claims: false,
    claimDetail: false,
    claimResolution: false,
    claimOrderItems: false,
    claimOrderItemBinding: false,
    serviceRecords: false,
    serviceRecordCreating: false
  })

  const statusUpdating = reactive<WarrantyStatusUpdating>({
    claim: null
  })

  const shipmentFilters = reactive<WarrantyFilters>({
    status: 'all',
    keyword: ''
  })

  const claimFilters = reactive<WarrantyFilters>({
    status: 'all'
  })

  const shipmentPagination = reactive<WarrantyPagination>({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const claimPagination = reactive<WarrantyPagination>({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const canEdit = computed(() => authStore.hasPermission('product:edit'))

  const statItems = computed(() => [
    {
      key: 'shipments',
      label: '已发货',
      value: shipmentStats.value.total_count ?? shipmentPagination.total,
      icon: ShieldCheck,
      tone: 'blue'
    },
    {
      key: 'active',
      label: '有效保修',
      value: shipmentStats.value.active_count ?? 0,
      icon: ShieldCheck,
      tone: 'green'
    },
    {
      key: 'expired',
      label: '已过期',
      value: shipmentStats.value.expired_count ?? 0,
      icon: Clock3,
      tone: 'amber'
    },
    {
      key: 'unbound',
      label: '待补充凭据',
      value: shipmentStats.value.unbound_count ?? 0,
      icon: FileWarning,
      tone: 'gray'
    },
    {
      key: 'claims',
      label: '当前申请',
      value: claimPagination.total,
      icon: FileWarning,
      tone: 'coral'
    }
  ])

  const fetchShipmentStats = async (): Promise<void> => {
    shipmentStats.value = await warrantyApi.getShipmentStats()
  }

  const fetchShipments = async (): Promise<void> => {
    loading.shipments = true
    try {
      const response = await warrantyApi.listShipmentRecords({
        page: shipmentPagination.page,
        page_size: shipmentPagination.pageSize,
        keyword: shipmentFilters.keyword?.trim() || undefined,
        status: shipmentFilters.status === 'all' ? undefined : shipmentFilters.status
      })
      shipments.value = response.data
      shipmentPagination.total = response.pagination.total || 0
      if (selectedShipment.value) {
        const refreshed = shipments.value.find((item) => Number(item.order_id) === Number(selectedShipment.value?.order_id))
        if (refreshed) selectedShipment.value = { ...selectedShipment.value, ...refreshed }
      }
    } finally {
      loading.shipments = false
    }
  }

  const fetchClaims = async (): Promise<void> => {
    loading.claims = true
    try {
      const response = await warrantyApi.listWarrantyClaims({
        page: claimPagination.page,
        page_size: claimPagination.pageSize,
        status: claimFilters.status === 'all' ? undefined : claimFilters.status
      })
      claims.value = response.data
      claimPagination.total = response.pagination.total || 0

      if (selectedClaim.value) {
        const refreshed = claims.value.find((claim) => Number(claim.id) === Number(selectedClaim.value?.id))
        if (refreshed) selectedClaim.value = { ...selectedClaim.value, ...refreshed }
      }
    } finally {
      loading.claims = false
    }
  }

  const refreshCurrent = async (): Promise<void> => {
    refreshing.value = true
    try {
      await fetchShipmentStats()
      if (activeTab.value === 'shipments') await fetchShipments()
      else if (activeTab.value === 'claims') await fetchClaims()
    } finally {
      refreshing.value = false
    }
  }

  const setShipmentDraft = (record: WarrantyShipmentRecord): void => {
    Object.assign(shipmentDraft, {
      shippingNote: record.shipping_note || '',
      shippingImages: [...(record.shipping_images || [])],
      productCodes: [...(record.product_codes || [])],
      warrantyMonths: Number(record.warranty_months || 12),
      warrantyStart: record.warranty_start_at
        ? new Date(record.warranty_start_at).toISOString().slice(0, 10)
        : record.shipped_at
          ? new Date(record.shipped_at).toISOString().slice(0, 10)
          : ''
    })
  }

  const selectShipment = async (record?: WarrantyShipmentRecord | null): Promise<void> => {
    if (!record?.id) {
      selectedShipment.value = null
      return
    }
    selectedShipment.value = record
    setShipmentDraft(record)
    loading.shipmentDetail = true
    try {
      const detail = await warrantyApi.getShipmentRecord(record.order_id || record.id)
      selectedShipment.value = detail
      setShipmentDraft(detail)
    } finally {
      loading.shipmentDetail = false
    }
  }

  const updateShipmentDraft = (patch: Partial<WarrantyShipmentDraft>): void => {
    Object.assign(shipmentDraft, patch)
  }

  const removeShipmentImage = (image: string): void => {
    shipmentDraft.shippingImages = shipmentDraft.shippingImages.filter((item) => item !== image)
  }

  const saveShipment = async (): Promise<void> => {
    if (!selectedShipment.value?.order_id) return
    loading.shipmentSaving = true
    try {
      const record = await warrantyApi.updateShipmentRecord(selectedShipment.value.order_id, {
        shipping_note: shipmentDraft.shippingNote,
        shipping_images: shipmentDraft.shippingImages,
        product_codes: shipmentDraft.productCodes,
        warranty_months: Number(shipmentDraft.warrantyMonths || 12),
        warranty_start_at: shipmentDraft.warrantyStart
      })
      selectedShipment.value = record
      setShipmentDraft(record)
      const index = shipments.value.findIndex((item) => Number(item.order_id) === Number(record.order_id))
      if (index >= 0) shipments.value[index] = record
      toast.success('发货记录已保存')
    } finally {
      loading.shipmentSaving = false
    }
  }

  const uploadShipmentImages = async (files: File[]): Promise<void> => {
    if (!selectedShipment.value?.order_id || files.length === 0) return
    if ((shipmentDraft.shippingImages.length + files.length) > 10) {
      toast.error('发货凭据图片最多 10 张')
      return
    }
    for (const file of files) {
      const validation = await validateUploadFile(file, 'warranty_evidence')
      if (!validation.ok) {
        toast.error(validation.error || '发货凭据图片不符合上传规范')
        return
      }
      if (validation.warning) toast.warning(validation.warning)
    }
    loading.shipmentUploading = true
    try {
      const record = await warrantyApi.uploadShipmentImages(selectedShipment.value.order_id, files)
      selectedShipment.value = record
      setShipmentDraft(record)
      const index = shipments.value.findIndex((item) => Number(item.order_id) === Number(record.order_id))
      if (index >= 0) shipments.value[index] = record
      toast.success('发货图片已上传')
    } finally {
      loading.shipmentUploading = false
    }
  }

  const updateClaimStatus = async (claim: WarrantyClaim, status: string): Promise<void> => {
    if (!claim?.id || claim.status === status) return
    statusUpdating.claim = claim.id
    try {
      await warrantyApi.updateWarrantyClaimStatus(claim.id, status)
      claim.status = status
      claim.processed_at = new Date().toISOString()
      if (selectedClaim.value?.id === claim.id) {
        selectedClaim.value = { ...selectedClaim.value, status: claim.status, processed_at: claim.processed_at }
      }
      toast.success('保修申请状态已更新')
    } finally {
      statusUpdating.claim = null
    }
  }

  const selectClaim = async (claim?: WarrantyClaim | null): Promise<void> => {
    if (!claim?.id) {
      selectedClaim.value = null
      claimResolutionDraft.value = ''
      claimOrderItems.value = []
      serviceRecords.value = []
      orderItemSelection.value = 'none'
      return
    }

    selectedClaim.value = claim
    claimResolutionDraft.value = claim.resolution || ''
    claimOrderItems.value = []
    serviceRecords.value = []
    orderItemSelection.value = claim.order_item_id ? String(claim.order_item_id) : 'none'
    loading.claimDetail = true
    try {
      const detail = await warrantyApi.getWarrantyClaim(claim.id)
      selectedClaim.value = detail
      claimResolutionDraft.value = detail.resolution || ''
      orderItemSelection.value = detail.order_item_id ? String(detail.order_item_id) : 'none'
      serviceRecords.value = detail.service_records || []
      await Promise.all([
        fetchClaimOrderItems(detail.id, detail.order_number),
        fetchServiceRecords(detail.id)
      ])
    } finally {
      loading.claimDetail = false
    }
  }

  const fetchClaimOrderItems = async (claimID?: WarrantyID | null, orderNumber?: string | null): Promise<void> => {
    if (!claimID || !orderNumber) {
      claimOrderItems.value = []
      return
    }

    loading.claimOrderItems = true
    try {
      claimOrderItems.value = await warrantyApi.listWarrantyClaimOrderItems(claimID)
    } finally {
      loading.claimOrderItems = false
    }
  }

  const fetchServiceRecords = async (claimID?: WarrantyID | null): Promise<void> => {
    if (!claimID) {
      serviceRecords.value = []
      return
    }

    loading.serviceRecords = true
    try {
      serviceRecords.value = await warrantyApi.listWarrantyServiceRecords(claimID)
    } finally {
      loading.serviceRecords = false
    }
  }

  const updateOrderItemSelection = (value: string): void => {
    orderItemSelection.value = value
  }

  const bindClaimOrderItem = async (): Promise<void> => {
    if (!selectedClaim.value?.id) return
    const nextOrderItemID = orderItemSelection.value === 'none' ? null : Number(orderItemSelection.value)
    loading.claimOrderItemBinding = true
    try {
      await warrantyApi.bindWarrantyClaimOrderItem(selectedClaim.value.id, nextOrderItemID)
      const boundItem = claimOrderItems.value.find((item) => Number(item.id) === Number(nextOrderItemID)) || null
      selectedClaim.value = {
        ...selectedClaim.value,
        order_item_id: nextOrderItemID,
        order_item: boundItem
      }
      const index = claims.value.findIndex((claim) => Number(claim.id) === Number(selectedClaim.value?.id))
      if (index >= 0) {
        claims.value[index] = {
          ...claims.value[index],
          order_item_id: nextOrderItemID,
          order_item: boundItem
        }
      }
      toast.success(nextOrderItemID ? '订单行已绑定' : '订单行已解绑')
    } finally {
      loading.claimOrderItemBinding = false
    }
  }

  const updateClaimResolutionDraft = (resolution: string): void => {
    claimResolutionDraft.value = resolution
  }

  const saveClaimResolution = async (): Promise<void> => {
    if (!selectedClaim.value?.id) return
    loading.claimResolution = true
    try {
      await warrantyApi.updateWarrantyClaimResolution(selectedClaim.value.id, claimResolutionDraft.value)
      selectedClaim.value = {
        ...selectedClaim.value,
        resolution: claimResolutionDraft.value,
        processed_at: new Date().toISOString()
      }
      const index = claims.value.findIndex((claim) => Number(claim.id) === Number(selectedClaim.value?.id))
      if (index >= 0) {
        claims.value[index] = {
          ...claims.value[index],
          resolution: claimResolutionDraft.value,
          processed_at: selectedClaim.value.processed_at
        }
      }
      toast.success('处理备注已保存')
    } finally {
      loading.claimResolution = false
    }
  }

  const resetServiceRecordForm = (): void => {
    Object.assign(serviceRecordForm, {
      serviceType: 'inspection',
      status: 'open',
      summary: '',
      costAmount: '',
      currency: '',
      performedAt: ''
    })
  }

  const updateServiceRecordForm = (patch: Partial<WarrantyServiceRecordForm> = {}): void => {
    Object.assign(serviceRecordForm, patch)
  }

  const createServiceRecord = async (): Promise<void> => {
    if (!selectedClaim.value?.id) return
    loading.serviceRecordCreating = true
    try {
      const record = await warrantyApi.createWarrantyServiceRecord(selectedClaim.value.id, {
        service_type: serviceRecordForm.serviceType,
        status: serviceRecordForm.status,
        summary: serviceRecordForm.summary.trim(),
        cost_amount: Number(serviceRecordForm.costAmount || 0),
        currency: serviceRecordForm.currency.trim().toUpperCase(),
        performed_at: serviceRecordForm.performedAt || undefined
      })
      serviceRecords.value = [record, ...serviceRecords.value]
      selectedClaim.value = {
        ...selectedClaim.value,
        service_records: serviceRecords.value
      }
      resetServiceRecordForm()
      toast.success('服务记录已添加')
    } finally {
      loading.serviceRecordCreating = false
    }
  }

  const updateShipmentPage = (page: number): void => {
    shipmentPagination.page = page
  }

  const updateShipmentPageSize = (pageSize: number): void => {
    shipmentPagination.pageSize = pageSize
    shipmentPagination.page = 1
  }

  const updateClaimPage = (page: number): void => {
    claimPagination.page = page
  }

  const updateClaimPageSize = (pageSize: number): void => {
    claimPagination.pageSize = pageSize
    claimPagination.page = 1
  }

  watch(
    () => [shipmentFilters.status, shipmentFilters.keyword],
    () => {
      shipmentPagination.page = 1
    }
  )

  watch(
    () => [
      shipmentFilters.status,
      shipmentFilters.keyword,
      shipmentPagination.page,
      shipmentPagination.pageSize
    ],
    fetchShipments
  )

  watch(
    () => claimFilters.status,
    () => {
      claimPagination.page = 1
    }
  )

  watch(
    () => [
      claimFilters.status,
      claimPagination.page,
      claimPagination.pageSize
    ],
    fetchClaims
  )

  watch(activeTab, async (tab) => {
    if (tab === 'shipments' && shipments.value.length === 0) await fetchShipments()
    if (tab === 'claims' && claims.value.length === 0) await fetchClaims()
  })

  onMounted(async () => {
    await Promise.all([
      fetchShipmentStats(),
      fetchShipments(),
      fetchClaims()
    ])
  })

  return {
    activeTab,
    refreshing,
    shipments,
    claims,
    selectedShipment,
    shipmentDraft,
    selectedClaim,
    claimResolutionDraft,
    claimOrderItems,
    serviceRecords,
    orderItemSelection,
    serviceRecordForm,
    loading,
    statusUpdating,
    shipmentFilters,
    claimFilters,
    shipmentPagination,
    claimPagination,
    canEdit,
    statItems,
    claimStatusOptions: CLAIM_STATUS_OPTIONS,
    serviceTypeOptions: WARRANTY_SERVICE_TYPE_OPTIONS,
    serviceStatusOptions: WARRANTY_SERVICE_STATUS_OPTIONS,
    refreshCurrent,
    selectShipment,
    updateShipmentDraft,
    removeShipmentImage,
    saveShipment,
    uploadShipmentImages,
    updateClaimStatus,
    selectClaim,
    updateOrderItemSelection,
    bindClaimOrderItem,
    updateClaimResolutionDraft,
    saveClaimResolution,
    updateServiceRecordForm,
    createServiceRecord,
    updateShipmentPage,
    updateShipmentPageSize,
    updateClaimPage,
    updateClaimPageSize
  }
}

