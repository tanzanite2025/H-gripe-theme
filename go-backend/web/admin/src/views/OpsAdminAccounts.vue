<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 管理员账号"
      description="在已登录的环境中创建、激活或轮换后台管理员账号。"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="loadAccounts">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新账号
        </Button>
      </template>
    </AdminPageHeader>

    <section class="flex items-start gap-3 rounded-2xl border border-dashed border-amber-500/30 bg-amber-500/5 p-4">
      <ShieldAlert class="mt-0.5 size-5 shrink-0 text-amber-600" />
      <div class="space-y-1 text-xs">
        <p class="font-black">首次发布且数据库没有后台账号时，页面还不能登录。</p>
        <p class="text-muted-foreground">
          先在目标环境执行一次
          <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]">/app/adminctl ensure-admin</code>
          创建首个管理员；之后就可以从这里维护账号。
        </p>
      </div>
    </section>

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)]">
      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>创建或重置管理员</CardTitle>
          <CardDescription>邮箱已存在时会重置密码并重新激活该账号；不会显示或保存明文密码。</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4 pt-5">
          <div
            v-if="feedback.message"
            class="flex items-start gap-2 rounded-lg border px-3 py-2 text-xs font-semibold leading-5"
            :class="feedback.type === 'success'
              ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700'
              : 'border-destructive/40 bg-destructive/10 text-destructive'"
            role="alert"
          >
            <CircleCheck v-if="feedback.type === 'success'" class="mt-0.5 size-4 shrink-0" />
            <ShieldAlert v-else class="mt-0.5 size-4 shrink-0" />
            <span>{{ feedback.message }}</span>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <label class="space-y-1.5">
              <span class="text-xs font-bold">管理员邮箱</span>
              <Input v-model="form.email" type="email" autocomplete="email" placeholder="admin@example.com" />
            </label>

            <label class="space-y-1.5">
              <span class="text-xs font-bold">用户名（可选）</span>
              <Input v-model="form.username" autocomplete="username" placeholder="留空按邮箱生成" />
            </label>

            <label class="space-y-1.5 sm:col-span-2">
              <span class="text-xs font-bold">新密码</span>
              <div class="relative">
                <Input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  autocomplete="new-password"
                  placeholder="至少 12 位，包含三类字符"
                  class="pr-10"
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
            </label>

            <label class="space-y-1.5">
              <span class="text-xs font-bold">角色</span>
              <Select v-model="form.role">
                <SelectTrigger class="w-full"><SelectValue placeholder="选择角色" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">超级管理员</SelectItem>
                  <SelectItem value="manager">经理</SelectItem>
                  <SelectItem value="editor">编辑</SelectItem>
                  <SelectItem value="support">客服</SelectItem>
                </SelectContent>
              </Select>
            </label>

            <label class="space-y-1.5">
              <span class="text-xs font-bold">后台语言</span>
              <Select v-model="form.locale">
                <SelectTrigger class="w-full"><SelectValue placeholder="选择语言" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="en">English</SelectItem>
                  <SelectItem value="zh_cn">简体中文</SelectItem>
                </SelectContent>
              </Select>
            </label>

            <label class="space-y-1.5">
              <span class="text-xs font-bold">名字</span>
              <Input v-model="form.first_name" autocomplete="given-name" />
            </label>

            <label class="space-y-1.5">
              <span class="text-xs font-bold">姓氏</span>
              <Input v-model="form.last_name" autocomplete="family-name" />
            </label>
          </div>

          <label class="flex items-start gap-2 rounded-lg border border-border/70 bg-muted/30 px-3 py-2 text-xs leading-5">
            <input v-model="confirmed" type="checkbox" class="mt-1 size-3.5 accent-primary" />
            <span>我确认这是管理员账号维护操作，并已通过安全渠道确认新密码。</span>
          </label>

          <div class="flex justify-end">
            <Button :disabled="submitting || !confirmed" @click="submit">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              <KeyRound v-else class="size-4" />
              {{ submitting ? '正在保存' : '创建 / 重置账号' }}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>发布前流程</CardTitle>
          <CardDescription>DEV 和生产使用同一套账号维护逻辑。</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4 pt-5 text-xs leading-5">
          <div class="space-y-1">
            <p class="font-black">1. DEV</p>
            <p class="text-muted-foreground">先用 `adminctl ensure-admin` 创建 DEV 首个账号，再登录本页维护其他账号。</p>
          </div>
          <div class="space-y-1">
            <p class="font-black">2. 发布</p>
            <p class="text-muted-foreground">镜像只负责提供 `/app/adminctl`，不会在服务启动时偷偷重置密码。</p>
          </div>
          <div class="space-y-1">
            <p class="font-black">3. 生产</p>
            <p class="text-muted-foreground">生产数据库没有账号时，先在生产容器执行一次 bootstrap；之后直接从本页轮换密码。</p>
          </div>
          <div class="rounded-lg border border-border/70 bg-muted/30 p-3 font-mono text-[11px] leading-5">
            ADMIN_EMAIL=admin@example.com<br />
            ADMIN_PASSWORD_FILE=/run/secrets/admin-password<br />
            ADMINCTL_CONFIRM=reset-production-admin
          </div>
        </CardContent>
      </Card>
    </div>

    <Card size="sm">
      <CardHeader class="flex flex-col gap-3 border-b border-dashed border-border/70 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <CardTitle>现有后台账号</CardTitle>
          <CardDescription>只显示 admin、manager、editor 和 support 账号。</CardDescription>
        </div>
        <div class="flex w-full gap-2 sm:w-auto">
          <Input v-model="search" class="sm:w-64" placeholder="搜索邮箱或用户名" @keyup.enter="loadAccounts" />
          <Button variant="outline" size="icon" aria-label="搜索账号" title="搜索账号" @click="loadAccounts">
            <Search class="size-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent class="pt-1">
        <div v-if="loading" class="flex items-center justify-center gap-2 py-8 text-xs text-muted-foreground">
          <LoaderCircle class="size-4 animate-spin" />
          正在读取账号
        </div>
        <div v-else-if="accounts.length === 0" class="py-8 text-center text-xs text-muted-foreground">
          当前环境没有后台账号，请先使用上方的发布前 CLI 创建首个账号。
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[680px] text-left text-xs">
            <thead class="border-b border-dashed border-border/70 text-[10px] uppercase tracking-widest text-muted-foreground/70">
              <tr>
                <th class="px-3 py-3 font-black">账号</th>
                <th class="px-3 py-3 font-black">角色</th>
                <th class="px-3 py-3 font-black">状态</th>
                <th class="px-3 py-3 font-black">更新时间</th>
                <th class="px-3 py-3 text-right font-black">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="account in accounts" :key="account.id" class="border-b border-dashed border-border/60 last:border-0">
                <td class="px-3 py-3">
                  <p class="font-black">{{ account.email }}</p>
                  <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ account.username }}</p>
                </td>
                <td class="px-3 py-3">{{ roleLabel(account.role) }}</td>
                <td class="px-3 py-3">
                  <span class="rounded-full px-2 py-1 text-[10px] font-black" :class="statusClass(account.status)">
                    {{ statusLabel(account.status) }}
                  </span>
                </td>
                <td class="px-3 py-3 text-muted-foreground">{{ formatDate(account.updated_at) }}</td>
                <td class="px-3 py-3 text-right">
                  <Button variant="ghost" size="sm" @click="loadAccountIntoForm(account)">
                    <Settings2 class="size-4" />
                    载入
                  </Button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  CircleCheck,
  Eye,
  EyeOff,
  KeyRound,
  LoaderCircle,
  RefreshCw,
  Search,
  Settings2,
  ShieldAlert,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import opsApi, { type OpsAdminAccount } from '@/api/ops'

type FeedbackType = 'success' | 'error' | ''

interface AdminAccountForm {
  email: string
  username: string
  password: string
  role: 'admin' | 'manager' | 'editor' | 'support'
  first_name: string
  last_name: string
  locale: string
}

const accounts = ref<OpsAdminAccount[]>([])
const loading = ref(false)
const submitting = ref(false)
const showPassword = ref(false)
const confirmed = ref(false)
const search = ref('')
const feedback = reactive<{ type: FeedbackType; message: string }>({ type: '', message: '' })
const form = reactive<AdminAccountForm>(defaultForm())

function defaultForm(): AdminAccountForm {
  return {
    email: '',
    username: '',
    password: '',
    role: 'admin',
    first_name: '',
    last_name: '',
    locale: 'en',
  }
}

const loadAccounts = async (): Promise<void> => {
  loading.value = true
  try {
    accounts.value = await opsApi.listAdminAccounts(search.value)
  } catch (error: any) {
    feedback.type = 'error'
    feedback.message = error?.response?.data?.message || error?.response?.data?.error || '管理员账号读取失败'
  } finally {
    loading.value = false
  }
}

const validateForm = (): string => {
  if (!form.email.trim() || !form.email.includes('@')) return '请输入有效的管理员邮箱。'
  if (form.username.trim() && (form.username.trim().length < 3 || form.username.trim().length > 50)) {
    return '用户名长度必须在 3 到 50 个字符之间。'
  }
  if (form.password.length < 12) return '密码至少需要 12 位。'
  const classes = [
    /[a-z]/.test(form.password),
    /[A-Z]/.test(form.password),
    /\d/.test(form.password),
    /[^A-Za-z0-9]/.test(form.password),
  ].filter(Boolean).length
  if (classes < 3) return '密码需要至少包含小写、大写、数字、符号中的三类。'
  return ''
}

const submit = async (): Promise<void> => {
  feedback.message = ''
  const validationError = validateForm()
  if (validationError) {
    feedback.type = 'error'
    feedback.message = validationError
    return
  }

  submitting.value = true
  try {
    const result = await opsApi.ensureAdminAccount({
      email: form.email.trim(),
      username: form.username.trim() || undefined,
      password: form.password,
      role: form.role,
      first_name: form.first_name.trim() || undefined,
      last_name: form.last_name.trim() || undefined,
      locale: form.locale,
    })
    feedback.type = 'success'
    feedback.message = result.created
      ? `管理员账号 ${result.email} 已创建。`
      : `管理员账号 ${result.email} 已重置并激活。`
    form.password = ''
    confirmed.value = false
    await loadAccounts()
  } catch (error: any) {
    feedback.type = 'error'
    feedback.message = error?.response?.data?.message || error?.response?.data?.error || '管理员账号保存失败'
  } finally {
    submitting.value = false
  }
}

const loadAccountIntoForm = (account: OpsAdminAccount): void => {
  form.email = account.email
  form.username = account.username
  form.role = account.role
  form.first_name = account.first_name || ''
  form.last_name = account.last_name || ''
  form.locale = account.locale || 'en'
  form.password = ''
  confirmed.value = false
  feedback.type = ''
  feedback.message = ''
}

const roleLabel = (role: string): string => ({
  admin: '超级管理员',
  manager: '经理',
  editor: '编辑',
  support: '客服',
}[role] || role)

const statusLabel = (status: string): string => ({
  active: '活跃',
  inactive: '未激活',
  suspended: '已停用',
}[status] || status)

const statusClass = (status: string): string => ({
  active: 'bg-emerald-500/10 text-emerald-700',
  inactive: 'bg-muted text-muted-foreground',
  suspended: 'bg-destructive/10 text-destructive',
}[status] || 'bg-muted text-muted-foreground')

const formatDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '-'

onMounted(loadAccounts)
</script>
