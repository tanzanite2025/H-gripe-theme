<template>
  <Card size="sm">
    <CardHeader class="border-b border-dashed border-border/70">
      <CardTitle>创建或重置后台账号</CardTitle>
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
              :title="showPassword ? '隐藏密码' : '显示密码'"
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
        <span>我确认这是后台账号维护操作，并已通过安全渠道确认新密码。</span>
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
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { CircleCheck, Eye, EyeOff, KeyRound, LoaderCircle, ShieldAlert } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { OpsAdminAccount, OpsAdminAccountInput } from '@/api/ops'

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

defineProps<{
  submitting: boolean
  feedback: { type: FeedbackType; message: string }
}>()

const emit = defineEmits<{
  (event: 'submit', payload: OpsAdminAccountInput): void
  (event: 'validation-error', message: string): void
}>()

const showPassword = ref(false)
const confirmed = ref(false)
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

function submit(): void {
  const validationError = validateForm()
  if (validationError) {
    emit('validation-error', validationError)
    return
  }

  emit('submit', {
    email: form.email.trim(),
    username: form.username.trim() || undefined,
    password: form.password,
    role: form.role,
    first_name: form.first_name.trim() || undefined,
    last_name: form.last_name.trim() || undefined,
    locale: form.locale,
  })
}

function validateForm(): string {
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

function loadAccount(account: OpsAdminAccount): void {
  form.email = account.email
  form.username = account.username
  form.role = account.role
  form.first_name = account.first_name || ''
  form.last_name = account.last_name || ''
  form.locale = account.locale || 'en'
  form.password = ''
  confirmed.value = false
}

function clearPassword(): void {
  form.password = ''
  confirmed.value = false
  showPassword.value = false
}

defineExpose({
  loadAccount,
  clearPassword,
  validateForm,
})
</script>
