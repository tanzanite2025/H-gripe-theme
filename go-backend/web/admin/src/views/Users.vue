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

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.4fr)_minmax(150px,0.7fr)_minmax(150px,0.7fr)_auto]" @submit.prevent="applyFilters">
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="邮箱、用户名或姓名" />
          </div>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">ROLE / 角色</span>
          <Select v-model="filters.role">
            <SelectTrigger class="h-9 w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部角色</SelectItem>
              <SelectItem value="admin">超级管理员</SelectItem>
              <SelectItem value="manager">经理</SelectItem>
              <SelectItem value="editor">编辑</SelectItem>
              <SelectItem value="support">客服</SelectItem>
              <SelectItem value="viewer">查看者</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">STATUS / 状态</span>
          <Select v-model="filters.status">
            <SelectTrigger class="h-9 w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="active">活跃</SelectItem>
              <SelectItem value="inactive">未激活</SelectItem>
              <SelectItem value="suspended">已停用</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading" :batch-visible="selectedUsers.length > 0">
      <template #batch>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs font-bold text-muted-foreground/80">已选择 {{ selectedUsers.length }} 个用户</span>
          <Button
            v-if="hasPermission('user:delete')"
            variant="destructive"
            size="sm"
            @click="requestBatchDelete"
          >
            <Trash2 class="size-3.5" />
            批量删除
          </Button>
        </div>
      </template>

      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-11">
              <Checkbox
                :model-value="selectionState"
                aria-label="选择当前页用户"
                @update:model-value="toggleAllUsers"
              />
            </TableHead>
            <TableHead class="w-20">ID</TableHead>
            <TableHead>用户名</TableHead>
            <TableHead>邮箱</TableHead>
            <TableHead>姓名</TableHead>
            <TableHead class="w-28">角色</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-44">创建时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="users.length === 0" :colspan="9">
            <div class="flex flex-col items-center text-muted-foreground">
              <UsersRound class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无用户</span>
            </div>
          </TableEmpty>

          <TableRow v-for="user in users" :key="user.id">
            <TableCell>
              <Checkbox
                :model-value="isSelected(user.id)"
                :disabled="user.id === currentUser?.id"
                :aria-label="`选择用户 ${user.username}`"
                @update:model-value="toggleUser(user, $event)"
              />
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">{{ user.id }}</TableCell>
            <TableCell class="font-bold text-xs">{{ user.username }}</TableCell>
            <TableCell class="max-w-64 truncate text-muted-foreground">{{ user.email }}</TableCell>
            <TableCell>{{ formatFullName(user) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="roleTone(user.role)">{{ getRoleName(user.role) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(user.status)">{{ getStatusName(user.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(user.created_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理用户 ${user.username}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem v-if="hasPermission('user:edit')" @select="openEditDialog(user)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    v-if="hasPermission('user:edit') && user.id !== currentUser?.id"
                    @select="requestToggleStatus(user)"
                  >
                    <UserRoundCheck v-if="user.status !== 'active'" class="size-4" />
                    <UserRoundX v-else class="size-4" />
                    {{ user.status === 'active' ? '停用' : '启用' }}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('user:delete') && user.id !== currentUser?.id" />
                  <DropdownMenuItem
                    v-if="hasPermission('user:delete') && user.id !== currentUser?.id"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(user)"
                  >
                    <Trash2 class="size-4" />
                    删除
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="dialogVisible">
      <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto">
        <form @submit="submitUserForm">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'create' ? '添加用户' : '编辑用户' }}</DialogTitle>
            <DialogDescription>
              {{ dialogMode === 'create' ? '创建新的后台用户并分配角色。' : '更新账号资料、角色和状态。' }}
            </DialogDescription>
          </DialogHeader>

          <div class="grid grid-cols-1 gap-4 py-5 sm:grid-cols-2">
            <FormField v-slot="{ componentField }" name="email">
              <FormItem>
                <FormLabel>邮箱</FormLabel>
                <FormControl><Input v-bind="componentField" type="email" autocomplete="email" /></FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="username">
              <FormItem>
                <FormLabel>用户名</FormLabel>
                <FormControl><Input v-bind="componentField" autocomplete="username" /></FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="password">
              <FormItem class="sm:col-span-2">
                <FormLabel>密码</FormLabel>
                <FormControl>
                  <div class="relative">
                    <Input
                      v-bind="componentField"
                      :type="showPassword ? 'text' : 'password'"
                      :placeholder="dialogMode === 'create' ? '至少 6 位' : '留空则不修改'"
                      autocomplete="new-password"
                      class="pr-9"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      class="absolute right-1 top-1/2 -translate-y-1/2"
                      :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                      @click="showPassword = !showPassword"
                    >
                      <EyeOff v-if="showPassword" class="size-4" />
                      <Eye v-else class="size-4" />
                    </Button>
                  </div>
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="first_name">
              <FormItem>
                <FormLabel>名字</FormLabel>
                <FormControl><Input v-bind="componentField" /></FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="last_name">
              <FormItem>
                <FormLabel>姓氏</FormLabel>
                <FormControl><Input v-bind="componentField" /></FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="role">
              <FormItem>
                <FormLabel>角色</FormLabel>
                <Select v-bind="componentField">
                  <FormControl>
                    <SelectTrigger class="w-full"><SelectValue placeholder="请选择角色" /></SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="admin">超级管理员</SelectItem>
                    <SelectItem value="manager">经理</SelectItem>
                    <SelectItem value="editor">编辑</SelectItem>
                    <SelectItem value="support">客服</SelectItem>
                    <SelectItem value="viewer">查看者</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="locale">
              <FormItem>
                <FormLabel>语言</FormLabel>
                <Select v-bind="componentField">
                  <FormControl>
                    <SelectTrigger class="w-full"><SelectValue placeholder="请选择语言" /></SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="zh">中文</SelectItem>
                    <SelectItem value="en">English</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="status">
              <FormItem class="sm:col-span-2">
                <FormLabel>状态</FormLabel>
                <FormControl>
                  <RadioGroup v-bind="componentField" class="grid grid-cols-1 gap-2 sm:grid-cols-3">
                    <label class="flex h-9 items-center gap-2 rounded-lg border px-3 text-sm">
                      <RadioGroupItem value="active" />活跃
                    </label>
                    <label class="flex h-9 items-center gap-2 rounded-lg border px-3 text-sm">
                      <RadioGroupItem value="inactive" />未激活
                    </label>
                    <label class="flex h-9 items-center gap-2 rounded-lg border px-3 text-sm">
                      <RadioGroupItem value="suspended" />已停用
                    </label>
                  </RadioGroup>
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogVisible = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '正在保存' : '保存' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

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
import {
  Eye,
  EyeOff,
  LoaderCircle,
  MoreHorizontal,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  Trash2,
  UserRoundCheck,
  UserRoundX,
  UsersRound
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table'
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
