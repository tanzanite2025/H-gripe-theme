<template>
  <div class="space-y-4">
    <AdminPageHeader title="用户管理" description="管理后台账号、角色和账户状态">
      <template #actions>
        <Button v-if="hasPermission('user:create')" @click="openCreateDialog">
          <Plus class="size-4" />
          添加用户
        </Button>
      </template>
    </AdminPageHeader>

    <UserFilterPanel
      :filters="filters"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <UsersTablePanel
      :loading="loading"
      :users="users"
      :selected-users="selectedUsers"
      :pagination="pagination"
      :selection-state="selectionState"
      :current-user-id="currentUser?.id"
      :can-edit="hasPermission('user:edit')"
      :can-delete="hasPermission('user:delete')"
      :get-role-name="getRoleName"
      :role-tone="roleTone"
      :get-status-name="getStatusName"
      :status-tone="statusTone"
      :format-date="formatDate"
      :format-full-name="formatFullName"
      @batch-delete="requestBatchDelete"
      @toggle-all-users="toggleAllUsers"
      @toggle-user="toggleUser"
      @edit="openEditDialog"
      @toggle-status="requestToggleStatus"
      @delete="requestDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <UserEditorDialog
      v-model:open="dialogVisible"
      v-model:show-password="showPassword"
      :mode="dialogMode"
      :submitting="submitting"
      @submit="submitUserForm"
    />

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :destructive="confirmation.destructive"
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { toTypedSchema } from '@vee-validate/zod'
import { useForm } from 'vee-validate'
import { z } from 'zod'
import { toast } from 'vue-sonner'
import { Plus } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import UserEditorDialog from '@/components/admin/user/UserEditorDialog.vue'
import UserFilterPanel from '@/components/admin/user/UserFilterPanel.vue'
import UsersTablePanel from '@/components/admin/user/UsersTablePanel.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const currentUser = computed(() => authStore.user)

const loading = ref(false)
const users = ref([])
const selectedUsers = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const editingUserId = ref(null)
const submitting = ref(false)
const showPassword = ref(false)

const filters = reactive({
  search: '',
  role: 'all',
  status: 'all'
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const confirmation = reactive({
  open: false,
  type: '',
  target: null,
  title: '',
  description: '',
  confirmLabel: '确定',
  destructive: false
})

const userSchema = toTypedSchema(
  z.object({
    email: z.string().min(1, '请输入邮箱').email('请输入正确的邮箱格式'),
    username: z.string().min(3, '用户名至少 3 个字符').max(50, '用户名最多 50 个字符'),
    password: z.string().refine((value) => !value || value.length >= 6, '密码长度至少 6 位'),
    first_name: z.string().max(100, '名字过长'),
    last_name: z.string().max(100, '姓氏过长'),
    role: z.enum(['admin', 'manager', 'editor', 'support', 'viewer']),
    locale: z.enum(['zh', 'en']),
    status: z.enum(['active', 'inactive', 'suspended'])
  })
)

const {
  handleSubmit,
  resetForm: resetUserForm,
  setFieldError,
  setValues
} = useForm({
  validationSchema: userSchema,
  initialValues: defaultUserValues()
})

const selectionCandidates = computed(() => users.value.filter((user) => user.id !== currentUser.value?.id))
const selectionState = computed(() => {
  if (selectionCandidates.value.length === 0 || selectedUsers.value.length === 0) return false
  if (selectedUsers.value.length === selectionCandidates.value.length) return true
  return 'indeterminate'
})

function defaultUserValues() {
  return {
    email: '',
    username: '',
    password: '',
    first_name: '',
    last_name: '',
    role: 'viewer',
    locale: 'zh',
    status: 'active'
  }
}

const hasPermission = (permission) => authStore.hasPermission(permission)

const getRoleName = (role) => ({
  admin: '超级管理员',
  manager: '经理',
  editor: '编辑',
  support: '客服',
  viewer: '查看者'
})[role] || role

const roleTone = (role) => ({
  admin: 'coral',
  manager: 'amber',
  editor: 'green',
  support: 'blue',
  viewer: 'gray'
})[role] || 'gray'

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

const formatFullName = (user) => {
  const name = [user.first_name, user.last_name].filter(Boolean).join(' ')
  return name || '-'
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
      ...(filters.role !== 'all' ? { role: filters.role } : {}),
      ...(filters.status !== 'all' ? { status: filters.status } : {})
    }
    const response = await axios.get('/api/admin/users', { params })
    users.value = response.data.users || []
    pagination.total = response.data.pagination?.total || 0
    selectedUsers.value = []
  } catch (error) {
    console.error('Failed to fetch users:', error)
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  fetchUsers()
}

const resetFilters = () => {
  filters.search = ''
  filters.role = 'all'
  filters.status = 'all'
  pagination.page = 1
  fetchUsers()
}

const updatePage = (page) => {
  pagination.page = page
  fetchUsers()
}

const updatePageSize = (pageSize) => {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchUsers()
}

const openCreateDialog = () => {
  dialogMode.value = 'create'
  editingUserId.value = null
  showPassword.value = false
  resetUserForm({ values: defaultUserValues() })
  dialogVisible.value = true
}

const openEditDialog = (user) => {
  dialogMode.value = 'edit'
  editingUserId.value = user.id
  showPassword.value = false
  setValues({
    email: user.email || '',
    username: user.username || '',
    password: '',
    first_name: user.first_name || '',
    last_name: user.last_name || '',
    role: user.role || 'viewer',
    locale: user.locale || 'zh',
    status: user.status || 'active'
  })
  dialogVisible.value = true
}

const submitUserForm = handleSubmit(async (values) => {
  if (dialogMode.value === 'create' && !values.password) {
    setFieldError('password', '请输入密码')
    return
  }

  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/users', values)
      toast.success('用户创建成功')
    } else {
      const data = { ...values }
      if (!data.password) delete data.password
      await axios.put(`/api/admin/users/${editingUserId.value}`, data)
      toast.success('用户更新成功')
    }
    dialogVisible.value = false
    await fetchUsers()
  } catch (error) {
    console.error('Failed to save user:', error)
  } finally {
    submitting.value = false
  }
})

const isSelected = (userId) => selectedUsers.value.some((user) => user.id === userId)

const toggleAllUsers = (checked) => {
  selectedUsers.value = checked === true ? [...selectionCandidates.value] : []
}

const toggleUser = (user, checked) => {
  if (user.id === currentUser.value?.id) return
  if (checked === true && !isSelected(user.id)) {
    selectedUsers.value = [...selectedUsers.value, user]
    return
  }
  selectedUsers.value = selectedUsers.value.filter((selected) => selected.id !== user.id)
}

const setConfirmation = (values) => {
  Object.assign(confirmation, { open: true, destructive: false, confirmLabel: '确定', ...values })
}

const requestToggleStatus = (user) => {
  const enabling = user.status !== 'active'
  setConfirmation({
    type: 'toggle-status',
    target: user,
    title: enabling ? '启用用户？' : '停用用户？',
    description: `${enabling ? '启用' : '停用'}用户 ${user.username}。`,
    confirmLabel: enabling ? '启用' : '停用',
    destructive: !enabling
  })
}

const requestDelete = (user) => {
  setConfirmation({
    type: 'delete',
    target: user,
    title: '删除用户？',
    description: `用户 ${user.username} 将被永久删除，此操作不可恢复。`,
    confirmLabel: '删除',
    destructive: true
  })
}

const requestBatchDelete = () => {
  setConfirmation({
    type: 'batch-delete',
    target: [...selectedUsers.value],
    title: '批量删除用户？',
    description: `已选择的 ${selectedUsers.value.length} 个用户将被永久删除，此操作不可恢复。`,
    confirmLabel: '批量删除',
    destructive: true
  })
}

const executeConfirmedAction = async () => {
  const type = confirmation.type
  const target = confirmation.target
  confirmation.open = false

  try {
    if (type === 'toggle-status') {
      const status = target.status === 'active' ? 'suspended' : 'active'
      await axios.patch(`/api/admin/users/${target.id}/status`, { status })
      toast.success(status === 'active' ? '用户已启用' : '用户已停用')
    } else if (type === 'delete') {
      await axios.delete(`/api/admin/users/${target.id}`)
      toast.success('用户已删除')
    } else if (type === 'batch-delete') {
      await axios.post('/api/admin/users/batch-delete', { user_ids: target.map((user) => user.id) })
      toast.success('批量删除成功')
    }
    await fetchUsers()
  } catch (error) {
    console.error('Failed to update users:', error)
  }
}

onMounted(fetchUsers)
</script>
