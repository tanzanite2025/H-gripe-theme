<template>
  <div class="faqs-page">
    <div class="page-header">
      <h2>FAQ 管理</h2>
      <el-button
        v-if="hasPermission('faq:create')"
        type="primary"
        :icon="Plus"
        @click="showCreateDialog"
      >
        添加 FAQ
      </el-button>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="搜索">
          <el-input
            v-model="filters.search"
            placeholder="问题/答案"
            clearable
            @clear="fetchFAQs"
            @keyup.enter="fetchFAQs"
          />
        </el-form-item>

        <el-form-item label="分类">
          <el-select v-model="filters.category" placeholder="全部" clearable @change="fetchFAQs">
            <el-option
              v-for="cat in categories"
              :key="cat"
              :label="cat"
              :value="cat"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="语言">
          <el-select v-model="filters.locale" placeholder="全部" clearable @change="fetchFAQs">
            <el-option label="中文" value="zh" />
            <el-option label="English" value="en" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchFAQs">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- FAQ 列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="faqs"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column prop="id" label="ID" width="80" />
        
        <el-table-column prop="question" label="问题" min-width="200" show-overflow-tooltip />
        
        <el-table-column prop="answer" label="答案" min-width="250" show-overflow-tooltip />
        
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="locale" label="语言" width="80">
          <template #default="{ row }">
            {{ row.locale === 'zh' ? '中文' : 'EN' }}
          </template>
        </el-table-column>
        
        <el-table-column prop="sort_order" label="排序" width="80" />
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="hasPermission('faq:edit')"
              type="primary"
              size="small"
              link
              @click="showEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="hasPermission('faq:delete')"
              type="danger"
              size="small"
              link
              @click="deleteFAQ(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedFAQs.length > 0" class="batch-actions">
        <span>已选择 {{ selectedFAQs.length }} 项</span>
        <el-button
          v-if="hasPermission('faq:delete')"
          type="danger"
          size="small"
          @click="batchDelete"
        >
          批量删除
        </el-button>
      </div>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchFAQs"
        @current-change="fetchFAQs"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加 FAQ' : '编辑 FAQ'"
      width="700px"
    >
      <el-form
        ref="faqFormRef"
        :model="faqForm"
        :rules="faqFormRules"
        label-width="100px"
      >
        <el-form-item label="问题" prop="question">
          <el-input
            v-model="faqForm.question"
            type="textarea"
            :rows="2"
            placeholder="请输入问题"
          />
        </el-form-item>

        <el-form-item label="答案" prop="answer">
          <el-input
            v-model="faqForm.answer"
            type="textarea"
            :rows="4"
            placeholder="请输入答案"
          />
        </el-form-item>

        <el-form-item label="分类" prop="category">
          <el-input v-model="faqForm.category" placeholder="请输入分类" />
        </el-form-item>

        <el-form-item label="语言" prop="locale">
          <el-select v-model="faqForm.locale" placeholder="请选择语言" style="width: 100%">
            <el-option label="中文" value="zh" />
            <el-option label="English" value="en" />
          </el-select>
        </el-form-item>

        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="faqForm.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const loading = ref(false)
const faqs = ref([])
const categories = ref([])
const selectedFAQs = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const faqFormRef = ref(null)

const filters = reactive({
  search: '',
  category: '',
  locale: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const faqForm = reactive({
  question: '',
  answer: '',
  category: '',
  locale: 'zh',
  sort_order: 0
})

const faqFormRules = {
  question: [
    { required: true, message: '请输入问题', trigger: 'blur' }
  ],
  answer: [
    { required: true, message: '请输入答案', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请输入分类', trigger: 'blur' }
  ],
  locale: [
    { required: true, message: '请选择语言', trigger: 'change' }
  ]
}

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const fetchFAQs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    const response = await axios.get('/api/admin/faqs', { params })
    faqs.value = response.data.faqs
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('获取 FAQ 列表失败')
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const response = await axios.get('/api/admin/faqs/categories')
    categories.value = response.data.categories
  } catch (error) {
    console.error('获取分类失败', error)
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.category = ''
  filters.locale = ''
  pagination.page = 1
  fetchFAQs()
}

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}

const showEditDialog = async (faq) => {
  dialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/faqs/${faq.id}`)
    Object.assign(faqForm, response.data.faq)
    dialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取 FAQ 详情失败')
  }
}

const resetForm = () => {
  Object.assign(faqForm, {
    question: '',
    answer: '',
    category: '',
    locale: 'zh',
    sort_order: 0
  })
  if (faqFormRef.value) {
    faqFormRef.value.clearValidate()
  }
}

const submitForm = async () => {
  if (!faqFormRef.value) return

  await faqFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true

    try {
      if (dialogMode.value === 'create') {
        await axios.post('/api/admin/faqs', faqForm)
        ElMessage.success('FAQ 创建成功')
      } else {
        const { id, ...data } = faqForm
        await axios.put(`/api/admin/faqs/${id}`, data)
        ElMessage.success('FAQ 更新成功')
      }

      dialogVisible.value = false
      fetchFAQs()
      fetchCategories()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const deleteFAQ = async (faq) => {
  try {
    await ElMessageBox.confirm(`确定要删除此 FAQ 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/faqs/${faq.id}`)
    ElMessage.success('删除成功')
    fetchFAQs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSelectionChange = (selection) => {
  selectedFAQs.value = selection
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedFAQs.value.length} 个 FAQ 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    const ids = selectedFAQs.value.map(f => f.id)
    await axios.post('/api/admin/faqs/batch-delete', { ids })
    ElMessage.success('批量删除成功')
    fetchFAQs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

onMounted(() => {
  fetchFAQs()
  fetchCategories()
})
</script>

<style scoped>
.faqs-page {
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
</style>
