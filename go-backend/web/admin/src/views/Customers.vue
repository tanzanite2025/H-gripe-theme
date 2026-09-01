<template>
  <div class="space-y-4">
    <AdminPageHeader title="客户账户" description="前台注册客户">
      <template #actions>
        <Button variant="outline" size="icon" aria-label="刷新客户账户" @click="fetchCustomers">
          <RefreshCw class="size-4" />
        </Button>
      </template>
    </AdminPageHeader>

    <CustomerAccountsFilterPanel
      :filters="filters"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <CustomerAccountsTablePanel
      :loading="loading"
      :customers="customers"
      :pagination="pagination"
      :get-status-name="getStatusName"
      :status-tone="statusTone"
      :format-date="formatDate"
      :format-full-name="formatFullName"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RefreshCw } from '@lucide/vue'
import CustomerAccountsFilterPanel from '@/components/admin/customer/CustomerAccountsFilterPanel.vue'
import CustomerAccountsTablePanel from '@/components/admin/customer/CustomerAccountsTablePanel.vue'
import type {
  CustomerAccount,
  CustomerFilters,
  CustomerListResponse,
  CustomerPagination,
  CustomerStatusTone
} from '@/modules/customer/customerTypes'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'

const loading = ref(false)
const customers = ref<CustomerAccount[]>([])

const filters = reactive<CustomerFilters>({
  search: '',
  status: 'all'
})

const pagination = reactive<CustomerPagination>({
  page: 1,
  pageSize: 20,
  total: 0
})

const statusNames: Record<string, string> = {
  active: '活跃',
  inactive: '未激活',
  suspended: '已停用'
}

const statusTones: Record<string, CustomerStatusTone> = {
  active: 'green',
  inactive: 'gray',
  suspended: 'coral'
}

const getStatusName = (status?: string | null): string => statusNames[status || ''] || status || '-'
const statusTone = (status?: string | null): CustomerStatusTone => statusTones[status || ''] || 'gray'

const formatDate = (dateString?: string | null): string => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

const formatFullName = (customer: CustomerAccount): string => {
  const name = [customer.first_name, customer.last_name].filter(Boolean).join(' ')
  return name || customer.display_name || '-'
}

const fetchCustomers = async (): Promise<void> => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
      ...(filters.status !== 'all' ? { status: filters.status } : {})
    }
    const response = await axios.get<CustomerListResponse>('/api/admin/customers', { params })
    customers.value = response.data.customers || []
    pagination.total = response.data.pagination?.total || 0
  } catch (error) {
    console.error('Failed to fetch customers:', error)
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchCustomers()
}

const resetFilters = (): void => {
  filters.search = ''
  filters.status = 'all'
  pagination.page = 1
  void fetchCustomers()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchCustomers()
}

const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchCustomers()
}

onMounted(() => {
  void fetchCustomers()
})
</script>

