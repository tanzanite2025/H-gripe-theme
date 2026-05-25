<template>
  <div class="content-page">
    <div class="page-header">
      <h2>内容管理</h2>
      <el-button
        v-if="hasPermission('content:create')"
        type="primary"
        :icon="Plus"
        @click="showCreateDialog"
      >
        添加文章
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">总文章数</div>
            <div class="stat-value">{{ stats.total || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">已发布</div>
            <div class="stat-value text-success">{{ stats.published || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">草稿</div>
            <div class="stat-value text-warning">{{ stats.draft || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">总浏览量</div>
            <div class="stat-value text-primary">{{ stats.total_views || 0 }}</div>
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
            placeholder="标题/内容"
            clearable
            @clear="fetchPosts"
            @keyup.enter="fetchPosts"
          />
        </el-form-item>

        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable @change="fetchPosts">
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
            <el-option label="已归档" value="archived" />
          </el-select>
        </el-form-item>

        <el-form-item label="语言">
          <el-select v-model="filters.locale" placeholder="全部" clearable @change="fetchPosts">
            <el-option label="中文" value="zh" />
            <el-option label="English" value="en" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchPosts">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 文章列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="posts"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column prop="id" label="ID" width="80" />
        
        <el-table-column prop="title" label="标题" min-width="250" />
        
        <el-table-column prop="slug" label="Slug" min-width="180" />
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="locale" label="语言" width="80" />
        
        <el-table-column prop="view_count" label="浏览量" width="100" />
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="showEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="hasPermission('content:edit')"
              type="info"
              size="small"
              link
              @click="showTranslationsDialog(row)"
            >
              翻译
            </el-button>
            <el-button
              v-if="hasPermission('content:edit')"
              type="warning"
              size="small"
              link
              @click="toggleStatus(row)"
            >
              {{ row.status === 'published' ? '下线' : '发布' }}
            </el-button>
            <el-button
              v-if="hasPermission('content:delete')"
              type="danger"
              size="small"
              link
              @click="deletePost(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedPosts.length > 0" class="batch-actions">
        <span>已选择 {{ selectedPosts.length }} 项</span>
        <div>
          <el-button
            v-if="hasPermission('content:edit')"
            type="success"
            size="small"
            @click="batchUpdateStatus('published')"
          >
            批量发布
          </el-button>
          <el-button
            v-if="hasPermission('content:edit')"
            type="warning"
            size="small"
            @click="batchUpdateStatus('draft')"
          >
            批量转草稿
          </el-button>
          <el-button
            v-if="hasPermission('content:delete')"
            type="danger"
            size="small"
            @click="batchDelete"
          >
            批量删除
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
        @size-change="fetchPosts"
        @current-change="fetchPosts"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加文章' : '编辑文章'"
      width="900px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="postFormRef"
        :model="postForm"
        :rules="postFormRules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="16">
            <el-form-item label="标题" prop="title">
              <el-input v-model="postForm.title" placeholder="请输入文章标题" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="语言" prop="locale">
              <el-select v-model="postForm.locale" placeholder="请选择语言" style="width: 100%">
                <el-option label="中文" value="zh" />
                <el-option label="English" value="en" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="Slug" prop="slug">
          <el-input v-model="postForm.slug" placeholder="请输入 URL slug" />
        </el-form-item>

        <el-form-item label="摘要">
          <el-input
            v-model="postForm.excerpt"
            type="textarea"
            :rows="3"
            placeholder="请输入文章摘要"
          />
        </el-form-item>

        <el-form-item label="内容">
          <el-input
            v-model="postForm.content"
            type="textarea"
            :rows="10"
            placeholder="请输入文章内容（支持 Markdown）"
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-select v-model="postForm.status" placeholder="请选择状态" style="width: 100%">
                <el-option label="草稿" value="draft" />
                <el-option label="已发布" value="published" />
                <el-option label="已归档" value="archived" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="特色图片">
              <el-input v-model="postForm.featured_image" placeholder="图片 URL" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="标签">
          <el-input v-model="postForm.tags" placeholder="多个标签用逗号分隔" />
        </el-form-item>

        <el-divider content-position="left">SEO 设置</el-divider>

        <el-form-item label="SEO 标题">
          <el-input v-model="postForm.meta_title" placeholder="SEO 标题" />
        </el-form-item>

        <el-form-item label="SEO 描述">
          <el-input
            v-model="postForm.meta_description"
            type="textarea"
            :rows="2"
            placeholder="SEO 描述"
          />
        </el-form-item>

        <el-form-item label="SEO 关键词">
          <el-input v-model="postForm.meta_keywords" placeholder="多个关键词用逗号分隔" />
        </el-form-item>

        <el-form-item label="规范 URL">
          <el-input v-model="postForm.canonical_url" placeholder="规范 URL" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 翻译管理对话框 -->
    <el-dialog
      v-model="translationsDialogVisible"
      title="翻译管理"
      width="700px"
    >
      <div v-if="currentPost">
        <p><strong>原文章：</strong>{{ currentPost.title }} ({{ currentPost.locale }})</p>
        
        <el-divider />
        
        <h4>现有翻译版本</h4>
        <el-table :data="translations" border>
          <el-table-column prop="locale" label="语言" width="100" />
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)">
                {{ getStatusName(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button
                type="primary"
                size="small"
                link
                @click="editTranslation(row)"
              >
                编辑
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
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
const posts = ref([])
const selectedPosts = ref([])
const dialogVisible = ref(false)
const translationsDialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const postFormRef = ref(null)
const currentPost = ref(null)
const translations = ref([])

const stats = ref({})

const filters = reactive({
  search: '',
  status: '',
  locale: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const postForm = reactive({
  title: '',
  slug: '',
  content: '',
  excerpt: '',
  status: 'draft',
  locale: 'zh',
  featured_image: '',
  tags: '',
  meta_title: '',
  meta_description: '',
  meta_keywords: '',
  canonical_url: '',
  translation_group_id: null
})

const postFormRules = {
  title: [
    { required: true, message: '请输入文章标题', trigger: 'blur' }
  ],
  slug: [
    { required: true, message: '请输入 URL slug', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ],
  locale: [
    { required: true, message: '请选择语言', trigger: 'change' }
  ]
}

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const getStatusName = (status) => {
  const statusMap = {
    draft: '草稿',
    published: '已发布',
    archived: '已归档'
  }
  return statusMap[status] || status
}

const getStatusType = (status) => {
  const typeMap = {
    draft: 'info',
    published: 'success',
    archived: 'warning'
  }
  return typeMap[status] || ''
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/content/posts/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计数据失败', error)
  }
}

const fetchPosts = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    const response = await axios.get('/api/admin/content/posts', { params })
    posts.value = response.data.posts
    pagination.total = response.data.pagination.total
  } catch (error) {
    ElMessage.error('获取文章列表失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.status = ''
  filters.locale = ''
  pagination.page = 1
  fetchPosts()
}

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}

const showEditDialog = (post) => {
  dialogMode.value = 'edit'
  Object.assign(postForm, {
    id: post.id,
    title: post.title,
    slug: post.slug,
    content: post.content || '',
    excerpt: post.excerpt || '',
    status: post.status,
    locale: post.locale,
    featured_image: post.featured_image || '',
    tags: post.tags || '',
    meta_title: post.meta_title || '',
    meta_description: post.meta_desc || '',
    meta_keywords: post.meta_keywords || '',
    canonical_url: post.canonical_url || '',
    translation_group_id: post.translation_group_id
  })
  dialogVisible.value = true
}

const resetForm = () => {
  Object.assign(postForm, {
    title: '',
    slug: '',
    content: '',
    excerpt: '',
    status: 'draft',
    locale: 'zh',
    featured_image: '',
    tags: '',
    meta_title: '',
    meta_description: '',
    meta_keywords: '',
    canonical_url: '',
    translation_group_id: null
  })
  if (postFormRef.value) {
    postFormRef.value.clearValidate()
  }
}

const submitForm = async () => {
  if (!postFormRef.value) return

  await postFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true

    try {
      if (dialogMode.value === 'create') {
        await axios.post('/api/admin/content/posts', postForm)
        ElMessage.success('文章创建成功')
      } else {
        const { id, ...data } = postForm
        await axios.put(`/api/admin/content/posts/${id}`, data)
        ElMessage.success('文章更新成功')
      }

      dialogVisible.value = false
      fetchPosts()
      fetchStats()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const toggleStatus = async (post) => {
  const newStatus = post.status === 'published' ? 'draft' : 'published'
  const action = newStatus === 'published' ? '发布' : '下线'

  try {
    await ElMessageBox.confirm(`确定要${action}文章 ${post.title} 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await axios.patch(`/api/admin/content/posts/${post.id}/status`, { status: newStatus })
    ElMessage.success(`${action}成功`)
    fetchPosts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  }
}

const deletePost = async (post) => {
  try {
    await ElMessageBox.confirm(`确定要删除文章 ${post.title} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/content/posts/${post.id}`)
    ElMessage.success('删除成功')
    fetchPosts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const showTranslationsDialog = async (post) => {
  currentPost.value = post
  
  try {
    const response = await axios.get(`/api/admin/content/posts/${post.id}/translations`)
    translations.value = response.data.translations || []
    translationsDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取翻译版本失败')
  }
}

const editTranslation = (translation) => {
  translationsDialogVisible.value = false
  showEditDialog(translation)
}

const handleSelectionChange = (selection) => {
  selectedPosts.value = selection
}

const batchUpdateStatus = async (status) => {
  const action = status === 'published' ? '发布' : '转为草稿'
  
  try {
    await ElMessageBox.confirm(`确定要${action}选中的 ${selectedPosts.value.length} 篇文章吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const postIds = selectedPosts.value.map(p => p.id)
    await axios.post('/api/admin/content/posts/batch-status', {
      post_ids: postIds,
      status: status
    })
    ElMessage.success(`批量${action}成功`)
    fetchPosts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`批量${action}失败`)
    }
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedPosts.value.length} 篇文章吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    const postIds = selectedPosts.value.map(p => p.id)
    await axios.post('/api/admin/content/posts/batch-delete', { post_ids: postIds })
    ElMessage.success('批量删除成功')
    fetchPosts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

onMounted(() => {
  fetchStats()
  fetchPosts()
})
</script>

<style scoped>
.content-page {
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

.el-divider {
  margin: 20px 0;
}
</style>
