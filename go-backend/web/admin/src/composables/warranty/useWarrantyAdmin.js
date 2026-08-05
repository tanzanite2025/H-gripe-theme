import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import registrationApi from '@/api/registrations'
import { useRouteTab } from '@/composables/useRouteTab'
import { useAuthStore } from '@/stores/auth'
import {
  CLAIM_STATUS_OPTIONS,
  REGISTRATION_STATUS_OPTIONS,
  WARRANTY_SERVICE_STATUS_OPTIONS,
  WARRANTY_SERVICE_TYPE_OPTIONS
} from '@/lib/warrantyPresentation'
import {
  Clock3,
  FileWarning,
  ShieldCheck
} from '@lucide/vue'

export function useWarrantyAdmin() {
  const authStore = useAuthStore()
  const activeTab = useRouteTab({
    defaultValue: 'registrations',
    values: ['registrations', 'claims', 'expiring', 'boundary'],
    routes: {
      registrations: 'WarrantyRegistrations',
      claims: 'WarrantyClaims',
      expiring: 'WarrantyExpiring',
      boundary: 'WarrantyBoundary',
    },
  })
  const refreshing = ref(false)
  const stats = ref({})
  const registrations = ref([])
  const claims = ref([])
  const expiring = ref([])
  const selectedClaim = ref(null)
  const claimResolutionDraft = ref('')
  const claimOrderItems = ref([])
  const serviceRecords = ref([])
  const orderItemSelection = ref('none')
  const serviceRecordForm = reactive({
    serviceType: 'inspection',
    status: 'open',
    summary: '',
    costAmount: '',
    currency: '',
    performedAt: ''
  })

  const loading = reactive({
    stats: false,
    registrations: false,
    claims: false,
    expiring: false,
    claimDetail: false,
    claimResolution: false,
    claimOrderItems: false,
    claimOrderItemBinding: false,
    serviceRecords: false,
    serviceRecordCreating: false
  })

  const statusUpdating = reactive({
    registration: null,
    claim: null
  })

  const registrationFilters = reactive({
    status: 'all'
  })

  const claimFilters = reactive({
    status: 'all'
  })

  const registrationPagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const claimPagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const canEdit = computed(() => authStore.hasPermission('product:edit'))

  const statItems = computed(() => [
    {
      key: 'total',
      label: '总注册',
      value: stats.value.total_count ?? registrations.value.length,
      icon: ShieldCheck,
      tone: 'blue'
    },
    {
      key: 'active',
      label: '有效保修',
      value: stats.value.active_count ?? 0,
      icon: ShieldCheck,
      tone: 'green'
    },
    {
      key: 'expired',
      label: '已过期',
      value: stats.value.expired_count ?? 0,
      icon: Clock3,
      tone: 'amber'
    },
    {
      key: 'claims',
      label: '当前申请',
      value: claimPagination.total,
      icon: FileWarning,
      tone: 'coral'
    }
  ])

  const fetchStats = async () => {
    loading.stats = true
    try {
      stats.value = await registrationApi.getStats()
    } finally {
      loading.stats = false
    }
  }

  const fetchRegistrations = async () => {
    loading.registrations = true
    try {
      const response = await registrationApi.listRegistrations({
        page: registrationPagination.page,
        page_size: registrationPagination.pageSize,
        status: registrationFilters.status === 'all' ? undefined : registrationFilters.status
      })
      registrations.value = response.data
      registrationPagination.total = response.pagination.total || 0
    } finally {
      loading.registrations = false
    }
  }

  const fetchClaims = async () => {
    loading.claims = true
    try {
      const response = await registrationApi.listWarrantyClaims({
        page: claimPagination.page,
        page_size: claimPagination.pageSize,
        status: claimFilters.status === 'all' ? undefined : claimFilters.status
      })
      claims.value = response.data
      claimPagination.total = response.pagination.total || 0

      if (selectedClaim.value) {
        const refreshed = claims.value.find((claim) => Number(claim.id) === Number(selectedClaim.value.id))
        if (refreshed) selectedClaim.value = { ...selectedClaim.value, ...refreshed }
      }
    } finally {
      loading.claims = false
    }
  }

  const fetchExpiring = async () => {
    loading.expiring = true
    try {
      expiring.value = await registrationApi.listExpiringWarranties(30)
    } finally {
      loading.expiring = false
    }
  }

  const refreshCurrent = async () => {
    refreshing.value = true
    try {
      await fetchStats()
      if (activeTab.value === 'claims') await fetchClaims()
      else if (activeTab.value === 'expiring') await fetchExpiring()
      else await fetchRegistrations()
    } finally {
      refreshing.value = false
    }
  }

  const updateRegistrationStatus = async (registration, status) => {
    if (!registration?.id || registration.status === status) return
    statusUpdating.registration = registration.id
    try {
      await registrationApi.updateRegistrationStatus(registration.id, status)
      registration.status = status
      toast.success('注册状态已更新')
      await fetchStats()
    } finally {
      statusUpdating.registration = null
    }
  }

  const updateClaimStatus = async (claim, status) => {
    if (!claim?.id || claim.status === status) return
    statusUpdating.claim = claim.id
    try {
      await registrationApi.updateWarrantyClaimStatus(claim.id, status)
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

  const selectClaim = async (claim) => {
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
      const detail = await registrationApi.getWarrantyClaim(claim.id)
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

  const fetchClaimOrderItems = async (claimID, orderNumber) => {
    if (!claimID || !orderNumber) {
      claimOrderItems.value = []
      return
    }

    loading.claimOrderItems = true
    try {
      claimOrderItems.value = await registrationApi.listWarrantyClaimOrderItems(claimID)
    } finally {
      loading.claimOrderItems = false
    }
  }

  const fetchServiceRecords = async (claimID) => {
    if (!claimID) {
      serviceRecords.value = []
      return
    }

    loading.serviceRecords = true
    try {
      serviceRecords.value = await registrationApi.listWarrantyServiceRecords(claimID)
    } finally {
      loading.serviceRecords = false
    }
  }

  const updateOrderItemSelection = (value) => {
    orderItemSelection.value = value
  }

  const bindClaimOrderItem = async () => {
    if (!selectedClaim.value?.id) return
    const nextOrderItemID = orderItemSelection.value === 'none' ? null : Number(orderItemSelection.value)
    loading.claimOrderItemBinding = true
    try {
      await registrationApi.bindWarrantyClaimOrderItem(selectedClaim.value.id, nextOrderItemID)
      const boundItem = claimOrderItems.value.find((item) => Number(item.id) === Number(nextOrderItemID)) || null
      selectedClaim.value = {
        ...selectedClaim.value,
        order_item_id: nextOrderItemID,
        order_item: boundItem
      }
      const index = claims.value.findIndex((claim) => Number(claim.id) === Number(selectedClaim.value.id))
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

  const updateClaimResolutionDraft = (resolution) => {
    claimResolutionDraft.value = resolution
  }

  const saveClaimResolution = async () => {
    if (!selectedClaim.value?.id) return
    loading.claimResolution = true
    try {
      await registrationApi.updateWarrantyClaimResolution(selectedClaim.value.id, claimResolutionDraft.value)
      selectedClaim.value = {
        ...selectedClaim.value,
        resolution: claimResolutionDraft.value,
        processed_at: new Date().toISOString()
      }
      const index = claims.value.findIndex((claim) => Number(claim.id) === Number(selectedClaim.value.id))
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

  const resetServiceRecordForm = () => {
    Object.assign(serviceRecordForm, {
      serviceType: 'inspection',
      status: 'open',
      summary: '',
      costAmount: '',
      currency: '',
      performedAt: ''
    })
  }

  const updateServiceRecordForm = (patch = {}) => {
    Object.assign(serviceRecordForm, patch)
  }

  const createServiceRecord = async () => {
    if (!selectedClaim.value?.id) return
    loading.serviceRecordCreating = true
    try {
      const record = await registrationApi.createWarrantyServiceRecord(selectedClaim.value.id, {
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

  const updateRegistrationPage = (page) => {
    registrationPagination.page = page
  }

  const updateRegistrationPageSize = (pageSize) => {
    registrationPagination.pageSize = pageSize
    registrationPagination.page = 1
  }

  const updateClaimPage = (page) => {
    claimPagination.page = page
  }

  const updateClaimPageSize = (pageSize) => {
    claimPagination.pageSize = pageSize
    claimPagination.page = 1
  }

  watch(
    () => registrationFilters.status,
    () => {
      registrationPagination.page = 1
    }
  )

  watch(
    () => [
      registrationFilters.status,
      registrationPagination.page,
      registrationPagination.pageSize
    ],
    fetchRegistrations
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
    if (tab === 'claims' && claims.value.length === 0) await fetchClaims()
    if (tab === 'expiring' && expiring.value.length === 0) await fetchExpiring()
  })

  onMounted(async () => {
    await Promise.all([
      fetchStats(),
      fetchRegistrations(),
      fetchClaims(),
      fetchExpiring()
    ])
  })

  return {
    activeTab,
    refreshing,
    registrations,
    claims,
    expiring,
    selectedClaim,
    claimResolutionDraft,
    claimOrderItems,
    serviceRecords,
    orderItemSelection,
    serviceRecordForm,
    loading,
    statusUpdating,
    registrationFilters,
    claimFilters,
    registrationPagination,
    claimPagination,
    canEdit,
    statItems,
    registrationStatusOptions: REGISTRATION_STATUS_OPTIONS,
    claimStatusOptions: CLAIM_STATUS_OPTIONS,
    serviceTypeOptions: WARRANTY_SERVICE_TYPE_OPTIONS,
    serviceStatusOptions: WARRANTY_SERVICE_STATUS_OPTIONS,
    refreshCurrent,
    updateRegistrationStatus,
    updateClaimStatus,
    selectClaim,
    updateOrderItemSelection,
    bindClaimOrderItem,
    updateClaimResolutionDraft,
    saveClaimResolution,
    updateServiceRecordForm,
    createServiceRecord,
    updateRegistrationPage,
    updateRegistrationPageSize,
    updateClaimPage,
    updateClaimPageSize
  }
}
