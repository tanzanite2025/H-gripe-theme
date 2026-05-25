<template>
  <div class="marketing-page">
    <div class="page-header">
      <h2>营销管理</h2>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">优惠券总数</div>
            <div class="stat-value">{{ stats.coupons?.total || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">活跃优惠券</div>
            <div class="stat-value" style="color: #67c23a">{{ stats.coupons?.active || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">已使用次数</div>
            <div class="stat-value" style="color: #409eff">{{ stats.coupons?.used || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">会员总数</div>
            <div class="stat-value" style="color: #e6a23c">{{ stats.loyalty?.total_members || 0 }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 标签页 -->
    <el-card>
      <el-tabs v-model="activeTab">
        <!-- 优惠券管理 -->
        <el-tab-pane label="优惠券管理" name="coupons">
          <div class="tab-header">
            <el-button
              v-if="hasPermission('marketing:create')"
              type="primary"
              :icon="Plus"
              @click="showCreateCouponDialog"
            >
              创建优惠券
            </el-button>
          </div>

          <el-table
            v-loading="couponsLoading"
            :data="coupons"
            stripe
            style="width: 100%"
          >
            <el-table-column prop="code" label="优惠码" width="150" />
            <el-table-column label="类型" width="100">
              <template #default="{ row }">
                {{ row.type === 'fixed' ? '固定金额' : '百分比' }}
              </template>
            </el-table-column>
            <el-table-column prop="value" label="折扣值" width="100" />
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
            <el-table-column label="使用情况" width="120">
              <template #default="{ row }">
                {{ row.used_count }} / {{ row.usage_limit || '∞' }}
              </template>
            </el-table-column>
            <el-table-column label="有效期" width="200">
              <template #default="{ row }">
                {{ formatDate(row.start_date) }} ~ {{ formatDate(row.end_date) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="hasPermission('marketing:edit')"
                  type="primary"
                  size="small"
                  link
                  @click="showEditCouponDialog(row)"
                >
                  编辑
                </el-button>
                <el-button
                  v-if="hasPermission('marketing:delete')"
                  type="danger"
                  size="small"
                  link
                  @click="deleteCoupon(row)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 礼品卡管理 -->
        <el-tab-pane label="礼品卡管理" name="giftcards">
          <div class="tab-header">
            <el-button
              v-if="hasPermission('marketing:create')"
              type="primary"
              :icon="Plus"
              @click="showCreateGiftCardDialog"
            >
              创建礼品卡
            </el-button>
          </div>

          <el-table
            v-loading="giftCardsLoading"
            :data="giftCards"
            stripe
            style="width: 100%"
          >
            <el-table-column prop="code" label="卡号" width="180" />
            <el-table-column prop="initial_value" label="初始金额" width="120" />
            <el-table-column prop="balance" label="余额" width="120" />
            <el-table-column prop="recipient_email" label="收件人邮箱" min-width="180" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getGiftCardStatusType(row.status)">
                  {{ getGiftCardStatusName(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="primary"
                  size="small"
                  link
                  @click="viewGiftCard(row)"
                >
                  查看
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 会员等级 -->
        <el-tab-pane label="会员等级" name="levels">
          <div class="tab-header">
            <el-button
              v-if="hasPermission('marketing:create')"
              type="primary"
              :icon="Plus"
              @click="showCreateLevelDialog"
            >
              创建等级
            </el-button>
          </div>

          <el-table
            v-loading="levelsLoading"
            :data="levels"
            stripe
            style="width: 100%"
          >
            <el-table-column prop="name" label="等级名称" width="150" />
            <el-table-column label="积分范围" width="200">
              <template #default="{ row }">
                {{ row.min_points }} - {{ row.max_points }}
              </template>
            </el-table-column>
            <el-table-column prop="discount_rate" label="折扣率 (%)" width="120" />
            <el-table-column prop="points_multiplier" label="积分倍数" width="120" />
            <el-table-column prop="sort_order" label="排序" width="80" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="hasPermission('marketing:edit')"
                  type="primary"
                  size="small"
                  link
                  @click="showEditLevelDialog(row)"
                >
                  编辑
                </el-button>
                <el-button
                  v-if="hasPermission('marketing:delete')"
                  type="danger"
                  size="small"
                  link
                  @click="deleteLevel(row)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 创建/编辑优惠券对话框 -->
    <el-dialog
      v-model="couponDialogVisible"
      :title="couponDialogMode === 'create' ? '创建优惠券' : '编辑优惠券'"
      width="700px"
    >
      <el-form
        ref="couponFormRef"
        :model="couponForm"
        :rules="couponFormRules"
        label-width="120px"
      >
        <el-form-item label="优惠码" prop="code">
          <el-input v-model="couponForm.code" placeholder="请输入优惠码" />
        </el-form-item>

        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="couponForm.type">
            <el-radio label="fixed">固定金额</el-radio>
            <el-radio label="percentage">百分比</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="折扣值" prop="value">
          <el-input-number v-model="couponForm.value" :min="0" :precision="2" />
          <span style="margin-left: 10px; color: #909399;">
            {{ couponForm.type === 'percentage' ? '%' : '元' }}
          </span>
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input v-model="couponForm.description" type="textarea" :rows="2" />
        </el-form-item>

        <el-form-item label="最低消费">
          <el-input-number v-model="couponForm.min_amount" :min="0" :precision="2" />
        </el-form-item>

        <el-form-item label="最大折扣">
          <el-input-number v-model="couponForm.max_discount" :min="0" :precision="2" />
        </el-form-item>

        <el-form-item label="使用次数限制">
          <el-input-number v-model="couponForm.usage_limit" :min="0" />
          <span style="margin-left: 10px; color: #909399;">0 表示无限制</span>
        </el-form-item>

        <el-form-item label="单用户限制">
          <el-input-number v-model="couponForm.usage_limit_per_user" :min="0" />
        </el-form-item>

        <el-form-item label="有效期" prop="dates">
          <el-date-picker
            v-model="couponDates"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="启用">
          <el-switch v-model="couponForm.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="couponDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="couponSubmitting" @click="submitCouponForm">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建礼品卡对话框 -->
    <el-dialog
      v-model="giftCardDialogVisible"
      title="创建礼品卡"
      width="600px"
    >
      <el-form
        ref="giftCardFormRef"
        :model="giftCardForm"
        :rules="giftCardFormRules"
        label-width="120px"
      >
        <el-form-item label="卡号" prop="code">
          <el-input v-model="giftCardForm.code" placeholder="请输入卡号" />
        </el-form-item>

        <el-form-item label="金额" prop="initial_value">
          <el-input-number v-model="giftCardForm.initial_value" :min="0" :precision="2" />
        </el-form-item>

        <el-form-item label="收件人邮箱">
          <el-input v-model="giftCardForm.recipient_email" placeholder="请输入收件人邮箱" />
        </el-form-item>

        <el-form-item label="收件人姓名">
          <el-input v-model="giftCardForm.recipient_name" placeholder="请输入收件人姓名" />
        </el-form-item>

        <el-form-item label="发送人姓名">
          <el-input v-model="giftCardForm.sender_name" placeholder="请输入发送人姓名" />
        </el-form-item>

        <el-form-item label="祝福语">
          <el-input v-model="giftCardForm.message" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="giftCardDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="giftCardSubmitting" @click="submitGiftCardForm">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑会员等级对话框 -->
    <el-dialog
      v-model="levelDialogVisible"
      :title="levelDialogMode === 'create' ? '创建等级' : '编辑等级'"
      width="600px"
    >
      <el-form
        ref="levelFormRef"
        :model="levelForm"
        :rules="levelFormRules"
        label-width="120px"
      >
        <el-form-item label="等级名称" prop="name">
          <el-input v-model="levelForm.name" placeholder="请输入等级名称" />
        </el-form-item>

        <el-form-item label="最小积分" prop="min_points">
          <el-input-number v-model="levelForm.min_points" :min="0" />
        </el-form-item>

        <el-form-item label="最大积分" prop="max_points">
          <el-input-number v-model="levelForm.max_points" :min="0" />
        </el-form-item>

        <el-form-item label="折扣率 (%)">
          <el-input-number v-model="levelForm.discount_rate" :min="0" :max="100" :precision="2" />
        </el-form-item>

        <el-form-item label="积分倍数">
          <el-input-number v-model="levelForm.points_multiplier" :min="1" :precision="2" />
        </el-form-item>

        <el-form-item label="排序">
          <el-input-number v-model="levelForm.sort_order" :min="0" />
        </el-form-item>

        <el-form-item label="权益说明">
          <el-input v-model="levelForm.benefits" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="levelDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="levelSubmitting" @click="submitLevelForm">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const activeTab = ref('coupons')
const stats = ref({})

// 优惠券
const couponsLoading = ref(false)
const coupons = ref([])
const couponDialogVisible = ref(false)
const couponDialogMode = ref('create')
const couponSubmitting = ref(false)
const couponFormRef = ref(null)
const couponDates = ref([])

const couponForm = reactive({
  code: '',
  type: 'fixed',
  value: 0,
  description: '',
  min_amount: 0,
  max_discount: 0,
  usage_limit: 0,
  usage_limit_per_user: 0,
  start_date: '',
  end_date: '',
  enabled: true
})

const couponFormRules = {
  code: [{ required: true, message: '请输入优惠码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  value: [{ required: true, message: '请输入折扣值', trigger: 'blur' }]
}

// 礼品卡
const giftCardsLoading = ref(false)
const giftCards = ref([])
const giftCardDialogVisible = ref(false)
const giftCardSubmitting = ref(false)
const giftCardFormRef = ref(null)

const giftCardForm = reactive({
  code: '',
  initial_value: 0,
  recipient_email: '',
  recipient_name: '',
  sender_name: '',
  message: ''
})

const giftCardFormRules = {
  code: [{ required: true, message: '请输入卡号', trigger: 'blur' }],
  initial_value: [{ required: true, message: '请输入金额', trigger: 'blur' }]
}

// 会员等级
const levelsLoading = ref(false)
const levels = ref([])
const levelDialogVisible = ref(false)
const levelDialogMode = ref('create')
const levelSubmitting = ref(false)
const levelFormRef = ref(null)

const levelForm = reactive({
  name: '',
  min_points: 0,
  max_points: 0,
  discount_rate: 0,
  points_multiplier: 1,
  sort_order: 0,
  benefits: ''
})

const levelFormRules = {
  name: [{ required: true, message: '请输入等级名称', trigger: 'blur' }],
  min_points: [{ required: true, message: '请输入最小积分', trigger: 'blur' }],
  max_points: [{ required: true, message: '请输入最大积分', trigger: 'blur' }]
}

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

const getGiftCardStatusName = (status) => {
  const map = {
    active: '活跃',
    used: '已使用',
    expired: '已过期',
    cancelled: '已取消'
  }
  return map[status] || status
}

const getGiftCardStatusType = (status) => {
  const map = {
    active: 'success',
    used: 'info',
    expired: 'warning',
    cancelled: 'danger'
  }
  return map[status] || ''
}

// 获取统计
const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/marketing/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计失败', error)
  }
}

// 优惠券相关
const fetchCoupons = async () => {
  couponsLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/coupons')
    coupons.value = response.data.coupons
  } catch (error) {
    ElMessage.error('获取优惠券列表失败')
  } finally {
    couponsLoading.value = false
  }
}

const showCreateCouponDialog = () => {
  couponDialogMode.value = 'create'
  resetCouponForm()
  couponDialogVisible.value = true
}

const showEditCouponDialog = async (coupon) => {
  couponDialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/marketing/coupons/${coupon.id}`)
    const data = response.data.coupon
    Object.assign(couponForm, data)
    couponDates.value = [new Date(data.start_date), new Date(data.end_date)]
    couponDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取优惠券详情失败')
  }
}

const resetCouponForm = () => {
  Object.assign(couponForm, {
    code: '',
    type: 'fixed',
    value: 0,
    description: '',
    min_amount: 0,
    max_discount: 0,
    usage_limit: 0,
    usage_limit_per_user: 0,
    enabled: true
  })
  couponDates.value = []
  if (couponFormRef.value) {
    couponFormRef.value.clearValidate()
  }
}

const submitCouponForm = async () => {
  if (!couponFormRef.value) return

  await couponFormRef.value.validate(async (valid) => {
    if (!valid) return

    if (!couponDates.value || couponDates.value.length !== 2) {
      ElMessage.warning('请选择有效期')
      return
    }

    couponSubmitting.value = true

    try {
      const data = {
        ...couponForm,
        start_date: couponDates.value[0].toISOString(),
        end_date: couponDates.value[1].toISOString()
      }

      if (couponDialogMode.value === 'create') {
        await axios.post('/api/admin/marketing/coupons', data)
        ElMessage.success('优惠券创建成功')
      } else {
        const { id, ...updateData } = data
        await axios.put(`/api/admin/marketing/coupons/${id}`, updateData)
        ElMessage.success('优惠券更新成功')
      }

      couponDialogVisible.value = false
      fetchCoupons()
      fetchStats()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      couponSubmitting.value = false
    }
  })
}

const deleteCoupon = async (coupon) => {
  try {
    await ElMessageBox.confirm(`确定要删除优惠券 ${coupon.code} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/marketing/coupons/${coupon.id}`)
    ElMessage.success('删除成功')
    fetchCoupons()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 礼品卡相关
const fetchGiftCards = async () => {
  giftCardsLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/gift-cards')
    giftCards.value = response.data.gift_cards
  } catch (error) {
    ElMessage.error('获取礼品卡列表失败')
  } finally {
    giftCardsLoading.value = false
  }
}

const showCreateGiftCardDialog = () => {
  resetGiftCardForm()
  giftCardDialogVisible.value = true
}

const resetGiftCardForm = () => {
  Object.assign(giftCardForm, {
    code: '',
    initial_value: 0,
    recipient_email: '',
    recipient_name: '',
    sender_name: '',
    message: ''
  })
  if (giftCardFormRef.value) {
    giftCardFormRef.value.clearValidate()
  }
}

const submitGiftCardForm = async () => {
  if (!giftCardFormRef.value) return

  await giftCardFormRef.value.validate(async (valid) => {
    if (!valid) return

    giftCardSubmitting.value = true

    try {
      await axios.post('/api/admin/marketing/gift-cards', giftCardForm)
      ElMessage.success('礼品卡创建成功')
      giftCardDialogVisible.value = false
      fetchGiftCards()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      giftCardSubmitting.value = false
    }
  })
}

const viewGiftCard = async (giftCard) => {
  try {
    const response = await axios.get(`/api/admin/marketing/gift-cards/${giftCard.id}`)
    ElMessageBox.alert(JSON.stringify(response.data, null, 2), '礼品卡详情', {
      confirmButtonText: '关闭'
    })
  } catch (error) {
    ElMessage.error('获取礼品卡详情失败')
  }
}

// 会员等级相关
const fetchLevels = async () => {
  levelsLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/levels')
    levels.value = response.data.levels
  } catch (error) {
    ElMessage.error('获取会员等级列表失败')
  } finally {
    levelsLoading.value = false
  }
}

const showCreateLevelDialog = () => {
  levelDialogMode.value = 'create'
  resetLevelForm()
  levelDialogVisible.value = true
}

const showEditLevelDialog = async (level) => {
  levelDialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/marketing/levels/${level.id}`)
    Object.assign(levelForm, response.data.level)
    levelDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取等级详情失败')
  }
}

const resetLevelForm = () => {
  Object.assign(levelForm, {
    name: '',
    min_points: 0,
    max_points: 0,
    discount_rate: 0,
    points_multiplier: 1,
    sort_order: 0,
    benefits: ''
  })
  if (levelFormRef.value) {
    levelFormRef.value.clearValidate()
  }
}

const submitLevelForm = async () => {
  if (!levelFormRef.value) return

  await levelFormRef.value.validate(async (valid) => {
    if (!valid) return

    levelSubmitting.value = true

    try {
      if (levelDialogMode.value === 'create') {
        await axios.post('/api/admin/marketing/levels', levelForm)
        ElMessage.success('等级创建成功')
      } else {
        const { id, ...data } = levelForm
        await axios.put(`/api/admin/marketing/levels/${id}`, data)
        ElMessage.success('等级更新成功')
      }

      levelDialogVisible.value = false
      fetchLevels()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      levelSubmitting.value = false
    }
  })
}

const deleteLevel = async (level) => {
  try {
    await ElMessageBox.confirm(`确定要删除等级 ${level.name} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/marketing/levels/${level.id}`)
    ElMessage.success('删除成功')
    fetchLevels()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  fetchStats()
  fetchCoupons()
  fetchGiftCards()
  fetchLevels()
})
</script>

<style scoped>
.marketing-page {
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

.tab-header {
  margin-bottom: 16px;
}
</style>
