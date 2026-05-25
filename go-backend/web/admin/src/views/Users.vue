<template>
  <div class="users-page">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button
        v-if="hasPermission('user:create')"
        type="primary"
        :icon="Plus"
        @click="showCreateDialog"
      >
        添加用户
      </el-button>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="搜索">
          <el-input
            v-model="filters.search"
            placeholder="邮箱/用户名/姓名"
            clearable
            @clear="fetchUsers"
            @keyup.enter="fetchUsers"
          />
        </el-form-item>

        <el-form-item label="角色">
          <el-select v-model="filters.role" placeholder="全部" clearable @change="fetchUsers">
            <el-option label="超级管理员" value="admin" />
            <el-option label="经理" value="manager" />
            <el-option label="编辑" value="editor" />
            <el-option label="客服" value="support" />
            <el-option label="查看者" value="viewer" />
          </el-select>
        </el-form-item>

        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable @change="fetchUsers">
            <el-option label="活跃" value="active" />
            <el-option label="未激活" value="inactive" />
            <el-option label="已停用" value="suspended" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchUsers">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 用户列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="users"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column prop="id" label="ID" width="80" />
        
        <el-table-column prop="username" label="用户名" min-width="120" />
        
        <el-table-column prop="email" label="邮箱" min-width="180" />
        
        <el-table-column label="姓名" min-width="120">
          <template #default="{ row }">
            {{ row.first_name }} {{ row.last_name }}
          </template>
        </el-table-column>
        
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="getRoleType(row.role)">
              {{ getRoleName(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusName(row.status) }}
            </el-tag>
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
              v-if="hasPermission('user:edit')"
              type="primary"
              size="small"
              link
              @click="showEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="hasPermission('user:edit')"
              type="warning"
              size="small"
              link
              @click="toggleStatus(row)"
            >
              {{ row.status === 'active' ? '停用' : '启用' }}
            </el-button>
            <el-button
              v-if="hasPermission('user:delete') && row.id !== currentUser.id"
              type="danger"
              size="small"
              link
              @click="deleteUser(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedUsers.length > 0" class="batch-actions">
        <span>已选择 {{ selectedUsers.length }} 项</span>
        <el-button
          v-if="hasPermission('user:delete')"
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
        @size-change="fetchUsers"
        @current-change="fetchUsers"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加用户' : '编辑用户'"
      width="600px"
    >
      <el-form
        ref="userFormRef"
        :model="userForm"
        :rules="userFormRules"
        label-width="100px"
      >
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>

        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>

        <el-form-item label="密码" :prop="dialogMode === 'create' ? 'password' : ''">
          <el-input
            v-model="userForm.password"
            type="password"
            :placeholder="dialogMode === 'create' ? '请输入密码' : '留空则不修改'"
            show-password
          />
        </el-form-item>

        <el-form-item label="名字" prop="first_name">
          <el-input v-model="userForm.first_name" placeholder="请输入名字" />
        </el-form-item>

        <el-form-item label="姓氏" prop="last_name">
          <el-input v-model="userForm.last_name" placeholder="请输入姓氏" />
        </el-form-item>

        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="超级管理员" value="admin" />
            <el-option label="经理" value="manager" />
            <el-option label="编辑" value="editor" />
            <el-option label="客服" value="support" />
            <el-option label="查看者" value="viewer" />
          </el-select>
        </el-form-item>

        <el-form-item label="语言" prop="locale">
          <el-select v-model="userForm.locale" placeholder="请选择语言" style="width: 100%">
            <el-option label="中文" value="zh" />
            <el-option label="English" value="en" />
          </el-select>
        </el-form-item>

        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="userForm.status">
            <el-radio label="active">活跃</el-radio>
            <el-radio label="inactive">未激活</el-radio>
            <el-radio label="suspended">已停用</el-radio>
          </el-radio-group>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const currentUser = computed(() => authStore.user)

const loading = ref(false)
const users = ref([])
const selectedUsers = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const userFormRef = ref(null)

const filters = reactive({
  search: '',
  role: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const userForm = reactive({
  email: '',
  username: '',
  password: '',
  first_name: '',
  last_name: '',
  role: 'viewer',
  locale: 'zh',
  status: 'active'
})

const userFormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 个字符', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ]
}

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const getRoleName = (role) => {
  const roleMap = {
    admin: '超级管理员',
    manager: '经理',
    editor: '编辑',
    support: '客服',
    viewer: '查看者'
  }
  return roleMap[role] || role
}

const getRoleType = (role) => {
  const typeMap = {
    admin: 'danger',
    manager: 'warning',
    editor: 'success',
    support: 'info',
    viewer: ''
  }
  return typeMap[role] || ''
}

const getStatusName = (status) => {
  const statusMap = {
    active: '活跃',
    inactive: '未激活',
    suspended: '已停用'
  }
  return statusMap[status] || status
}

const getStatusType = (status) => {
  const typeMap = {
    active: 'success',
    inactive: 'info',
    suspended: 'danger'
  }
  return typeMap[status] || ''
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    const response = await axios.get('/api/admin/users', { params })
    users.value = response.data.users
    pagination.total = response.data.pagination.total
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.role = ''
  filters.status = ''
  pagination.page = 1
  fetchUsers()
}

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}

const showEditDialog = (user) => {
  dialogMode.value = 'edit'
  Object.assign(userForm, {
    id: user.id,
    email: user.email,
    username: user.username,
    password: '',
    first_name: user.first_name,
    last_name: user.last_name,
    role: user.role,
    locale: user.locale,
    status: user.status
  })
  dialogVisible.value = true
}

const resetForm = () => {
  Object.assign(userForm, {
    email: '',
    username: '',
    password: '',
    first_name: '',
    last_name: '',
    role: 'viewer',
    locale: 'zh',
    status: 'active'
  })
  if (userFormRef.value) {
    userFormRef.value.clearValidate()
  }
}

const submitForm = async () => {
  if (!userFormRef.value) return

  await userFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true

    try {
      if (dialogMode.value === 'create') {
        await axios.post('/api/admin/users', userForm)
        ElMessage.success('用户创建成功')
      } else {
        const { id, ...data } = userForm
        if (!data.password) {
          delete data.password
        }
        await axios.put(`/api/admin/users/${id}`, data)
        ElMessage.success('用户更新成功')
      }

      dialogVisible.value = false
      fetchUsers()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const toggleStatus = async (user) => {
  const newStatus = user.status === 'active' ? 'suspended' : 'active'
  const action = newStatus === 'active' ? '启用' : '停用'

  try {
    await ElMessageBox.confirm(`确定要${action}用户 ${user.username} 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await axios.patch(`/api/admin/users/${user.id}/status`, { status: newStatus })
    ElMessage.success(`${action}成功`)
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  }
}

const deleteUser = async (user) => {
  try {
    await ElMessageBox.confirm(`确定要删除用户 ${user.username} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/users/${user.id}`)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSelectionChange = (selection) => {
  selectedUsers.value = selection
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedUsers.value.length} 个用户吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    const userIds = selectedUsers.value.map(u => u.id)
    await axios.post('/api/admin/users/batch-delete', { user_ids: userIds })
    ElMessage.success('批量删除成功')
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.users-page {
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
