<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="后台账号管理"
      description="创建、激活或轮换后台身份账号；账号目录和维护操作仅对系统管理员开放。"
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
      <OpsAdminAccountMaintenanceForm
        ref="formPanel"
        :submitting="submitting"
        :feedback="feedback"
        @submit="submit"
        @validation-error="showValidationError"
      />
      <OpsAdminAccountBootstrapGuide />
    </div>

    <OpsAdminAccountTablePanel
      :accounts="accounts"
      :loading="loading"
      :search="search"
      @update:search="search = $event"
      @refresh="loadAccounts"
      @select-account="loadAccountIntoForm"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RefreshCw, ShieldAlert } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import OpsAdminAccountBootstrapGuide from '@/components/admin/ops/OpsAdminAccountBootstrapGuide.vue'
import OpsAdminAccountMaintenanceForm from '@/components/admin/ops/OpsAdminAccountMaintenanceForm.vue'
import OpsAdminAccountTablePanel from '@/components/admin/ops/OpsAdminAccountTablePanel.vue'
import { Button } from '@/components/ui/button'
import opsApi, { type OpsAdminAccount, type OpsAdminAccountInput } from '@/api/ops'

type FeedbackType = 'success' | 'error' | ''

const accounts = ref<OpsAdminAccount[]>([])
const loading = ref(false)
const submitting = ref(false)
const search = ref('')
const requestSequence = ref(0)
const formPanel = ref<{
  loadAccount: (account: OpsAdminAccount) => void
  clearPassword: () => void
} | null>(null)
const feedback = reactive<{ type: FeedbackType; message: string }>({ type: '', message: '' })

async function loadAccounts(): Promise<void> {
  const requestID = ++requestSequence.value
  loading.value = true
  try {
    const nextAccounts = await opsApi.listAdminAccounts(search.value)
    if (requestID !== requestSequence.value) return
    accounts.value = nextAccounts
  } catch (error: any) {
    if (requestID !== requestSequence.value) return
    feedback.type = 'error'
    feedback.message = error?.response?.data?.message || error?.response?.data?.error || '管理员账号读取失败'
  } finally {
    if (requestID === requestSequence.value) loading.value = false
  }
}

async function submit(payload: OpsAdminAccountInput): Promise<void> {
  feedback.message = ''
  submitting.value = true
  try {
    const result = await opsApi.ensureAdminAccount(payload)
    feedback.type = 'success'
    feedback.message = result.created
      ? `管理员账号 ${result.email} 已创建。`
      : `管理员账号 ${result.email} 已重置并激活。`
    formPanel.value?.clearPassword()
    await loadAccounts()
  } catch (error: any) {
    feedback.type = 'error'
    feedback.message = error?.response?.data?.message || error?.response?.data?.error || '管理员账号保存失败'
  } finally {
    submitting.value = false
  }
}

function showValidationError(message: string): void {
  feedback.type = 'error'
  feedback.message = message
}

function loadAccountIntoForm(account: OpsAdminAccount): void {
  formPanel.value?.loadAccount(account)
  feedback.type = ''
  feedback.message = ''
}

onMounted(loadAccounts)
</script>
