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

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { RefreshCw } from '@lucide/vue'
import CustomerAccountsFilterPanel from '@/components/admin/customer/CustomerAccountsFilterPanel.vue'
import CustomerAccountsTablePanel from '@/components/admin/customer/CustomerAccountsTablePanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'

const loading = ref(false)
const customers = ref([])

const filters = reactive({
  search: '',
  status: 'all'
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const getStatusName = (status) => ({
  active: '活跃',
  inactive: '未激活',
  suspended: '已停用'
})[status] || status

const statusTone = (status) => ({
  active: 'green',
  inactive: 'gray',
  suspended: 'coral'
})[status] || 'gray'

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

const formatFullName = (customer) => {
  const name = [customer.first_name, customer.last_name].filter(Boolean).join(' ')
  return name || customer.display_name || '-'
}

const fetchCustomers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
      ...(filters.status !== 'all' ? { status: filters.status } : {})
    }
    const response = await axios.get('/api/admin/customers', { params })
    customers.value = response.data.customers || []
    pagination.total = response.data.pagination?.total || 0
  } catch (error) {
    console.error('Failed to fetch customers:', error)
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  fetchCustomers()
}

const resetFilters = () => {
  filters.search = ''
  filters.status = 'all'
  pagination.page = 1
  fetchCustomers()
}

const updatePage = (page) => {
  pagination.page = page
  fetchCustomers()
}

const updatePageSize = (pageSize) => {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchCustomers()
}

onMounted(fetchCustomers)
</script>
