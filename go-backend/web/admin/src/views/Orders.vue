<template>
  <div class="orders-page">
    <div class="page-header">
      <h2>订单管理</h2>
      <el-button type="success" :icon="Download" @click="exportOrders">
        导出订单
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">总订单数</div>
            <div class="stat-value">{{ stats.total || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">今日订单</div>
            <div class="stat-value text-primary">{{ stats.today || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">总销售额</div>
            <div class="stat-value text-success">¥{{ formatMoney(stats.total_revenue || 0) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">今日销售额</div>
            <div class="stat-value text-warning">¥{{ formatMoney(stats.today_revenue || 0) }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="搜索">
          <el-input
            v-model="filters.search"
            placeholder="订单号/客户姓名/邮箱"
            clearable
            @clear="fetchOrders"
            @keyup.enter="fetchOrders"
          />
        </el-form-item>

        <el-form-item label="订单状态">
          <el-select v-model="filters.status" placeholder="全部" clearable @change="fetchOrders">
            <el-option label="待支付" value="pending" />
            <el-option label="已支付" value="paid" />
            <el-option label="处理中" value="processing" />
            <el-option label="已发货" value="shipped" />
            <el-option label="已完成" value="completed" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="已退款" value="refunded" />
          </el-select>
        </el-form-item>

        <el-form-item label="支付状态">
          <el-select v-model="filters.payment_status" placeholder="全部" clearable @change="fetchOrders">
            <el-option label="未支付" value="unpaid" />
            <el-option label="已支付" value="paid" />
            <el-option label="已退款" value="refunded" />
          </el-select>
        </el-form-item>

        <el-form-item label="物流状态">
          <el-select v-model="filters.shipping_status" placeholder="全部" clearable @change="fetchOrders">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已发货" value="shipped" />
            <el-option label="已送达" value="delivered" />
          </el-select>
        </el-form-item>

        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            @change="handleDateChange"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchOrders">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 订单列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="orders"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column prop="id" label="ID" width="80" />
        
        <el-table-column prop="order_number" label="订单号" width="180" />
        
        <el-table-column label="客户" min-width="150">
          <template #default="{ row }">
            {{ row.shipping_address.first_name }} {{ row.shipping_address.last_name }}
          </template>
        </el-table-column>
        
        <el-table-column label="订单状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getOrderStatusType(row.status)">
              {{ getOrderStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="支付状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getPaymentStatusType(row.payment_status)">
              {{ getPaymentStatusName(row.payment_status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="物流状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getShippingStatusType(row.shipping_status)">
              {{ getShippingStatusName(row.shipping_status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="total_amount" label="总金额" width="120">
          <template #default="{ row }">
            ¥{{ formatMoney(row.total_amount) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="showOrderDetail(row)"
            >
              详情
            </el-button>
            <el-button
              v-if="hasPermission('order:edit')"
              type="warning"
              size="small"
              link
              @click="showStatusDialog(row)"
            >
              状态
            </el-button>
            <el-button
              v-if="hasPermission('order:delete')"
              type="danger"
              size="small"
              link
              @click="deleteOrder(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedOrders.length > 0" class="batch-actions">
        <span>已选择 {{ selectedOrders.length }} 项</span>
        <div>
          <el-button
            v-if="hasPermission('order:edit')"
            type="success"
            size="small"
            @click="batchUpdateStatus('completed')"
          >
            批量完成
          </el-button>
          <el-button
            v-if="hasPermission('order:edit')"
            type="warning"
            size="small"
            @click="batchUpdateStatus('cancelled')"
          >
            批量取消
          </el-button>
        </div>
      </div>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchOrders"
        @current-change="fetchOrders"
      />
    </el-card>

    <!-- 订单详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="订单详情"
      width="900px"
    >
      <div v-if="currentOrder" class="order-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="订单号">{{ currentOrder.order_number }}</el-descriptions-item>
          <el-descriptions-item label="订单状态">
            <el-tag :type="getOrderStatusType(currentOrder.status)">
              {{ getOrderStatusName(currentOrder.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="支付状态">
            <el-tag :type="getPaymentStatusType(currentOrder.payment_status)">
              {{ getPaymentStatusName(currentOrder.payment_status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="物流状态">
            <el-tag :type="getShippingStatusType(currentOrder.shipping_status)">
              {{ getShippingStatusName(currentOrder.shipping_status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="支付方式">{{ currentOrder.payment_method || '-' }}</el-descriptions-item>
          <el-descriptions-item label="物流方式">{{ currentOrder.shipping_method || '-' }}</el-descriptions-item>
          <el-descriptions-item label="物流单号">{{ currentOrder.tracking_number || '-' }}</el-descriptions-item>
          <el-descriptions-item label="物流公司">{{ currentOrder.carrier_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(currentOrder.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="支付时间">{{ currentOrder.paid_at ? formatDate(currentOrder.paid_at) : '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">收货地址</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="姓名">
            {{ currentOrder.shipping_address.first_name }} {{ currentOrder.shipping_address.last_name }}
          </el-descriptions-item>
          <el-descriptions-item label="电话">{{ currentOrder.shipping_address.phone }}</el-descriptions-item>
          <el-descriptions-item label="邮箱" :span="2">{{ currentOrder.shipping_address.email }}</el-descriptions-item>
          <el-descriptions-item label="地址" :span="2">
            {{ currentOrder.shipping_address.address_1 }}
            {{ currentOrder.shipping_address.address_2 }}
          </el-descriptions-item>
          <el-descriptions-item label="城市">{{ currentOrder.shipping_address.city }}</el-descriptions-item>
          <el-descriptions-item label="省/州">{{ currentOrder.shipping_address.state }}</el-descriptions-item>
          <el-descriptions-item label="邮编">{{ currentOrder.shipping_address.postal_code }}</el-descriptions-item>
          <el-descriptions-item label="国家">{{ currentOrder.shipping_address.country }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">订单商品</el-divider>
        <el-table :data="currentOrder.items" border>
          <el-table-column prop="product_name" label="商品名称" />
          <el-table-column prop="sku" label="SKU" width="120" />
          <el-table-column prop="price" label="单价" width="100">
            <template #default="{ row }">¥{{ formatMoney(row.price) }}</template>
          </el-table-column>
          <el-table-column prop="quantity" label="数量" width="80" />
          <el-table-column prop="total" label="小计" width="120">
            <template #default="{ row }">¥{{ formatMoney(row.total) }}</template>
          </el-table-column>
        </el-table>

        <el-divider content-position="left">金额明细</el-divider>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="商品小计">¥{{ formatMoney(currentOrder.subtotal_amount) }}</el-descriptions-item>
          <el-descriptions-item label="运费">¥{{ formatMoney(currentOrder.shipping_fee) }}</el-descriptions-item>
          <el-descriptions-item label="税费">¥{{ formatMoney(currentOrder.tax_amount) }}</el-descriptions-item>
          <el-descriptions-item label="优惠">-¥{{ formatMoney(currentOrder.discount_amount) }}</el-descriptions-item>
          <el-descriptions-item label="订单总额">
            <span class="total-amount">¥{{ formatMoney(currentOrder.total_amount) }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">备注</el-divider>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="客户备注">{{ currentOrder.customer_note || '-' }}</el-descriptions-item>
          <el-descriptions-item label="管理员备注">
            <el-input
              v-model="adminNoteForm.admin_note"
              type="textarea"
              :rows="3"
              placeholder="请输入管理员备注"
            />
            <el-button
              v-if="hasPermission('order:edit')"
              type="primary"
              size="small"
              style="margin-top: 10px"
              @click="updateAdminNote"
            >
              保存备注
            </el-button>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 状态管理对话框 -->
    <el-dialog
      v-model="statusDialogVisible"
      title="状态管理"
      width="500px"
    >
      <el-form :model="statusForm" label-width="100px">
        <el-form-item label="订单号">
          <span>{{ statusForm.order_number }}</span>
        </el-form-item>
        <el-form-item label="订单状态">
          <el-select v-model="statusForm.status" style="width: 100%">
            <el-option label="待支付" value="pending" />
            <el-option label="已支付" value="paid" />
            <el-option label="处理中" value="processing" />
            <el-option label="已发货" value="shipped" />
            <el-option label="已完成" value="completed" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="已退款" value="refunded" />
          </el-select>
        </el-form-item>
        <el-form-item label="支付状态">
          <el-select v-model="statusForm.payment_status" style="width: 100%">
            <el-option label="未支付" value="unpaid" />
            <el-option label="已支付" value="paid" />
            <el-option label="已退款" value="refunded" />
          </el-select>
        </el-form-item>
        <el-form-item label="物流状态">
          <el-select v-model="statusForm.shipping_status" style="width: 100%">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已发货" value="shipped" />
            <el-option label="已送达" value="delivered" />
          </el-select>
        </el-form-item>
        <el-form-item label="物流单号">
          <el-input v-model="statusForm.tracking_number" placeholder="请输入物流单号" />
        </el-form-item>
        <el-form-item label="物流公司">
          <el-input v-model="statusForm.carrier_code" placeholder="请输入物流公司代码" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="statusDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitStatus">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Download } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const loading = ref(false)
const orders = ref([])
const selectedOrders = ref([])
const detailDialogVisible = ref(false)
const statusDialogVisible = ref(false)
const submitting = ref(false)
const currentOrder = ref(null)
const dateRange = ref([])

const stats = ref({})

const filters = reactive({
  search: '',
  status: '',
  payment_status: '',
  shipping_status: '',
  start_date: '',
  end_date: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const statusForm = reactive({
  id: null,
  order_number: '',
  status: '',
  payment_status: '',
  shipping_status: '',
  tracking_number: '',
  carrier_code: ''
})

const adminNoteForm = reactive({
  admin_note: ''
})

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const getOrderStatusName = (status) => {
  const statusMap = {
    pending: '待支付',
    paid: '已支付',
    processing: '处理中',
    shipped: '已发货',
    completed: '已完成',
    cancelled: '已取消',
    refunded: '已退款'
  }
  return statusMap[status] || status
}

const getOrderStatusType = (status) => {
  const typeMap = {
    pending: 'info',
    paid: 'success',
    processing: 'warning',
    shipped: 'primary',
    completed: 'success',
    cancelled: 'danger',
    refunded: 'warning'
  }
  return typeMap[status] || ''
}

const getPaymentStatusName = (status) => {
  const statusMap = {
    unpaid: '未支付',
    paid: '已支付',
    refunded: '已退款'
  }
  return statusMap[status] || status
}

const getPaymentStatusType = (status) => {
  const typeMap = {
    unpaid: 'info',
    paid: 'success',
    refunded: 'warning'
  }
  return typeMap[status] || ''
}

const getShippingStatusName = (status) => {
  const statusMap = {
    pending: '待处理',
    processing: '处理中',
    shipped: '已发货',
    delivered: '已送达'
  }
  return statusMap[status] || status
}

const getShippingStatusType = (status) => {
  const typeMap = {
    pending: 'info',
    processing: 'warning',
    shipped: 'primary',
    delivered: 'success'
  }
  return typeMap[status] || ''
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const formatMoney = (amount) => {
  return Number(amount).toFixed(2)
}

const handleDateChange = (dates) => {
  if (dates && dates.length === 2) {
    filters.start_date = dates[0].toISOString().split('T')[0]
    filters.end_date = dates[1].toISOString().split('T')[0]
  } else {
    filters.start_date = ''
    filters.end_date = ''
  }
  fetchOrders()
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/orders/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计数据失败', error)
  }
}

const fetchOrders = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    const response = await axios.get('/api/admin/orders', { params })
    orders.value = response.data.orders
    pagination.total = response.data.pagination.total
  } catch (error) {
    ElMessage.error('获取订单列表失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.status = ''
  filters.payment_status = ''
  filters.shipping_status = ''
  filters.start_date = ''
  filters.end_date = ''
  dateRange.value = []
  pagination.page = 1
  fetchOrders()
}

const showOrderDetail = async (order) => {
  try {
    const response = await axios.get(`/api/admin/orders/${order.id}`)
    currentOrder.value = response.data.order
    adminNoteForm.admin_note = currentOrder.value.admin_note || ''
    detailDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取订单详情失败')
  }
}

const showStatusDialog = (order) => {
  Object.assign(statusForm, {
    id: order.id,
    order_number: order.order_number,
    status: order.status,
    payment_status: order.payment_status,
    shipping_status: order.shipping_status,
    tracking_number: order.tracking_number || '',
    carrier_code: order.carrier_code || ''
  })
  statusDialogVisible.value = true
}

const submitStatus = async () => {
  submitting.value = true

  try {
    // 更新订单状态
    await axios.patch(`/api/admin/orders/${statusForm.id}/status`, {
      status: statusForm.status
    })

    // 更新支付状态
    await axios.patch(`/api/admin/orders/${statusForm.id}/payment-status`, {
      payment_status: statusForm.payment_status
    })

    // 更新物流状态
    await axios.patch(`/api/admin/orders/${statusForm.id}/shipping-status`, {
      shipping_status: statusForm.shipping_status
    })

    // 更新物流信息
    if (statusForm.tracking_number) {
      await axios.patch(`/api/admin/orders/${statusForm.id}/tracking`, {
        tracking_number: statusForm.tracking_number,
        carrier_code: statusForm.carrier_code
      })
    }

    ElMessage.success('状态更新成功')
    statusDialogVisible.value = false
    fetchOrders()
    fetchStats()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '更新失败')
  } finally {
    submitting.value = false
  }
}

const updateAdminNote = async () => {
  try {
    await axios.patch(`/api/admin/orders/${currentOrder.value.id}/admin-note`, {
      admin_note: adminNoteForm.admin_note
    })
    ElMessage.success('备注保存成功')
    currentOrder.value.admin_note = adminNoteForm.admin_note
  } catch (error) {
    ElMessage.error('保存备注失败')
  }
}

const deleteOrder = async (order) => {
  try {
    await ElMessageBox.confirm(`确定要删除订单 ${order.order_number} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/orders/${order.id}`)
    ElMessage.success('删除成功')
    fetchOrders()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSelectionChange = (selection) => {
  selectedOrders.value = selection
}

const batchUpdateStatus = async (status) => {
  const action = status === 'completed' ? '完成' : '取消'
  
  try {
    await ElMessageBox.confirm(`确定要${action}选中的 ${selectedOrders.value.length} 个订单吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const orderIds = selectedOrders.value.map(o => o.id)
    const response = await axios.post('/api/admin/orders/batch-status', {
      order_ids: orderIds,
      status: status
    })

    ElMessage.success(`批量${action}成功：${response.data.updated}/${response.data.total}`)
    fetchOrders()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`批量${action}失败`)
    }
  }
}

const exportOrders = async () => {
  try {
    const params = new URLSearchParams(filters).toString()
    const url = `/api/admin/orders/export?${params}`
    
    const response = await axios.get(url, {
      responseType: 'blob'
    })
    
    const blob = new Blob([response.data], { type: 'text/csv' })
    const link = document.createElement('a')
    link.href = window.URL.createObjectURL(blob)
    link.download = `orders_${new Date().getTime()}.csv`
    link.click()
    
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

onMounted(() => {
  fetchStats()
  fetchOrders()
})
</script>

<style scoped>
.orders-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.text-primary {
  color: #409eff;
}

.text-success {
  color: #67c23a;
}

.text-warning {
  color: #e6a23c;
}

.filter-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.batch-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-top: 1px solid #ebeef5;
  margin-top: 12px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}

.order-detail {
  max-height: 70vh;
  overflow-y: auto;
}

.total-amount {
  font-size: 18px;
  font-weight: bold;
  color: #f56c6c;
}

.el-divider {
  margin: 20px 0;
}
</style>
