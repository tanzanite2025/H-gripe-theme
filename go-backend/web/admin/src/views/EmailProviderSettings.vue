<template>
  <div class="space-y-5">
    <AdminPageHeader
      title="邮件发件通道"
      description="管理事务邮件使用的 SMTP Provider、默认通道和真实连通性测试。"
    >
      <template #actions>
        <Button variant="outline" :disabled="loadingEmailProviderSettings" @click="loadEmailProviderSettings">
          <RefreshCw :class="['size-4', loadingEmailProviderSettings ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEditEmailProviderSettings" @click="openCreateEmailProviderDialog">
          <Plus class="size-4" />
          添加发件通道
        </Button>
      </template>
    </AdminPageHeader>

    <section class="rounded-2xl border border-dashed border-border/80 bg-muted/10 p-4 sm:p-5">
      <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div class="flex min-w-0 items-start gap-3">
          <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Mail class="size-4" />
          </span>
          <div class="min-w-0">
            <h2 class="text-sm font-black">邮件域配置中心</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">
              设置默认且启用的 Provider 后，新产生的事务邮件会动态使用它。没有数据库默认通道时，系统才会回退到部署环境中的 SMTP 配置。
            </p>
          </div>
        </div>
        <RouterLink
          to="/email/templates"
          class="inline-flex shrink-0 items-center justify-center rounded-full border border-border/80 border-dashed px-3 py-2 text-xs font-black tracking-tight transition-colors hover:bg-muted"
        >
          管理事务邮件模板
        </RouterLink>
      </div>
    </section>

    <section class="overflow-hidden border bg-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
        <div>
          <h2 class="text-sm font-black">发件通道</h2>
          <p class="mt-1 text-xs text-muted-foreground">密码只保存为加密密文，列表和编辑表单不会回显原密码。</p>
        </div>
        <AdminStatusBadge :tone="defaultEmailProvider ? 'green' : 'amber'">
          {{ defaultEmailProvider ? '默认通道已配置' : '当前使用环境变量回退' }}
        </AdminStatusBadge>
      </div>

      <div v-if="loadingEmailProviderSettings" class="px-4 py-12 text-center text-sm text-muted-foreground">
        正在加载邮件通道
      </div>
      <div v-else-if="emailProviderSettingsRecords.length === 0" class="px-4 py-12 text-center">
        <Mail class="mx-auto size-8 text-muted-foreground/50" />
        <p class="mt-3 text-sm font-bold">还没有数据库邮件通道</p>
        <p class="mt-1 text-xs text-muted-foreground">可以继续使用环境变量 SMTP，也可以添加一个可动态切换的 Provider。</p>
        <Button v-if="canEditEmailProviderSettings" class="mt-4" @click="openCreateEmailProviderDialog">
          <Plus class="size-4" />
          添加第一个通道
        </Button>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[1080px] text-sm">
          <thead class="border-b bg-muted/30 text-left text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            <tr>
              <th class="px-4 py-3">通道</th>
              <th class="px-4 py-3">SMTP</th>
              <th class="px-4 py-3">发件身份</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">最近测试</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <tr v-for="provider in emailProviderSettingsRecords" :key="provider.id" class="align-top">
              <td class="px-4 py-4">
                <div class="flex items-start gap-2">
                  <span class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                    <Mail class="size-3.5" />
                  </span>
                  <div class="min-w-0">
                    <p class="font-bold">{{ provider.name }}</p>
                    <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ provider.code }}</p>
                  </div>
                </div>
              </td>
              <td class="px-4 py-4">
                <p class="font-mono text-xs font-bold">{{ provider.host }}:{{ provider.port }}</p>
                <p class="mt-1 text-[10px] uppercase text-muted-foreground">{{ provider.encryption_type }} · {{ provider.username || '无用户名' }}</p>
              </td>
              <td class="px-4 py-4">
                <p class="font-bold">{{ provider.from_name || '-' }}</p>
                <p class="mt-1 max-w-56 truncate font-mono text-[10px] text-muted-foreground" :title="provider.from_email">{{ provider.from_email }}</p>
                <p v-if="provider.reply_to" class="mt-1 max-w-56 truncate text-[10px] text-muted-foreground" :title="provider.reply_to">回复至 {{ provider.reply_to }}</p>
              </td>
              <td class="px-4 py-4">
                <div class="flex flex-wrap items-center gap-1.5">
                  <AdminStatusBadge :tone="provider.is_default ? 'green' : 'gray'">
                    {{ provider.is_default ? '默认' : '备用' }}
                  </AdminStatusBadge>
                  <AdminStatusBadge :tone="provider.is_active ? 'blue' : 'gray'">
                    {{ provider.is_active ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </div>
                <div v-if="canEditEmailProviderSettings" class="mt-2 flex items-center gap-2">
                  <Switch
                    size="sm"
                    :model-value="provider.is_active"
                    :disabled="busyEmailProviderID === provider.id"
                    :aria-label="`${provider.name}启用状态`"
                    @update:model-value="toggleEmailProviderActive(provider, Boolean($event))"
                  />
                  <span class="text-[10px] text-muted-foreground">切换启用</span>
                </div>
              </td>
              <td class="px-4 py-4">
                <div class="flex items-center gap-2">
                  <AdminStatusBadge :tone="emailProviderTestTone(provider)">
                    {{ emailProviderTestLabel(provider) }}
                  </AdminStatusBadge>
                </div>
                <p class="mt-1 text-[10px] text-muted-foreground">{{ formatEmailProviderTestTime(provider.last_tested_at) }}</p>
                <p v-if="provider.last_test_error" class="mt-1 max-w-56 truncate text-[10px] text-destructive" :title="provider.last_test_error">{{ provider.last_test_error }}</p>
              </td>
              <td class="px-4 py-4 text-right">
                <div class="flex flex-wrap justify-end gap-1">
                  <Button v-if="canEditEmailProviderSettings && !provider.is_default" size="sm" variant="outline" :disabled="busyEmailProviderID === provider.id" @click="setEmailProviderAsDefault(provider)">
                    <Star class="size-3.5" />
                    设为默认
                  </Button>
                  <Button v-if="canEditEmailProviderSettings" size="sm" variant="outline" :disabled="testingEmailProviderID === provider.id" @click="openEmailProviderTestDialog(provider)">
                    <LoaderCircle v-if="testingEmailProviderID === provider.id" class="size-3.5 animate-spin" />
                    <Send v-else class="size-3.5" />
                    测试发送
                  </Button>
                  <Button v-if="canEditEmailProviderSettings" size="icon" variant="ghost" :aria-label="`编辑${provider.name}`" :title="`编辑${provider.name}`" @click="openEditEmailProviderDialog(provider)">
                    <Pencil class="size-4" />
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <Dialog v-model:open="emailProviderSettingsDialogOpen">
      <DialogContent class="max-h-[calc(100dvh-1.5rem)] max-w-3xl overflow-y-auto" @open-auto-focus.prevent>
        <DialogHeader>
          <DialogTitle>{{ emailProviderSettingsForm.id ? '编辑邮件发件通道' : '添加邮件发件通道' }}</DialogTitle>
          <DialogDescription>只填写 SMTP 传输和发件身份信息；保存后可单独发送测试邮件验证连接。</DialogDescription>
        </DialogHeader>

        <form class="space-y-5" @submit.prevent="saveEmailProviderSettings">
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminFormField label="通道编码" required>
              <Input v-model="emailProviderSettingsForm.code" class="font-mono" :disabled="Boolean(emailProviderSettingsForm.id)" placeholder="例如 primary_smtp" />
            </AdminFormField>
            <AdminFormField label="通道名称" required>
              <Input v-model="emailProviderSettingsForm.name" placeholder="例如 官方事务邮件 SMTP" />
            </AdminFormField>
            <AdminFormField label="SMTP 主机" required>
              <Input v-model="emailProviderSettingsForm.host" class="font-mono" placeholder="smtp.example.com" />
            </AdminFormField>
            <AdminFormField label="SMTP 端口" required>
              <Input v-model.number="emailProviderSettingsForm.port" type="number" min="1" max="65535" />
            </AdminFormField>
            <AdminFormField label="SMTP 用户名">
              <Input v-model="emailProviderSettingsForm.username" autocomplete="off" />
            </AdminFormField>
            <AdminFormField label="加密方式" required>
              <Select v-model="emailProviderSettingsForm.encryption_type">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="starttls">STARTTLS</SelectItem>
                  <SelectItem value="tls">TLS</SelectItem>
                  <SelectItem value="none">无加密</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>
            <AdminFormField label="发件人名称" required>
              <Input v-model="emailProviderSettingsForm.from_name" placeholder="官方客户支持" />
            </AdminFormField>
            <AdminFormField label="发件人邮箱" required>
              <Input v-model="emailProviderSettingsForm.from_email" type="email" placeholder="support@example.com" />
            </AdminFormField>
            <AdminFormField label="回复地址">
              <Input v-model="emailProviderSettingsForm.reply_to" type="email" placeholder="可选，默认使用发件人邮箱" />
            </AdminFormField>
            <AdminFormField label="SMTP 密码">
              <Input v-model="emailProviderSettingsForm.password" type="password" autocomplete="new-password" :placeholder="emailProviderSettingsForm.hasExistingPassword ? '已保存密码，留空保持不变' : '输入 SMTP 授权码或密码'" />
            </AdminFormField>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="flex items-center justify-between gap-4 rounded-xl border border-dashed px-3 py-3 text-sm">
              <span>
                <span class="block font-bold">启用通道</span>
                <span class="mt-1 block text-[10px] text-muted-foreground">停用后不能作为默认发送通道。</span>
              </span>
              <Switch v-model="emailProviderSettingsForm.is_active" />
            </label>
            <label class="flex items-center justify-between gap-4 rounded-xl border border-dashed px-3 py-3 text-sm">
              <span>
                <span class="block font-bold">设为默认</span>
                <span class="mt-1 block text-[10px] text-muted-foreground">同一时间只保留一个默认通道。</span>
              </span>
              <Switch v-model="emailProviderSettingsForm.is_default" />
            </label>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="emailProviderSettingsDialogOpen = false">取消</Button>
            <Button type="submit" :disabled="savingEmailProviderSettings">
              <LoaderCircle v-if="savingEmailProviderSettings" class="size-4 animate-spin" />
              保存通道
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="emailProviderTestDialogOpen">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>测试邮件通道</DialogTitle>
          <DialogDescription>
            将通过 {{ selectedEmailProviderForTest?.name || '当前通道' }} 发送一封真实测试邮件，收件地址必须明确填写。
          </DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="sendEmailProviderTestMessage">
          <AdminFormField label="测试收件地址" required>
            <Input v-model="emailProviderTestTargetEmail" type="email" autocomplete="email" placeholder="your-email@example.com" />
          </AdminFormField>
          <p class="text-xs leading-5 text-muted-foreground">测试结果会保存在通道记录中；失败时不会回退到其他 Provider。</p>
          <DialogFooter>
            <Button type="button" variant="outline" @click="emailProviderTestDialogOpen = false">取消</Button>
            <Button type="submit" :disabled="testingEmailProviderID !== null">
              <LoaderCircle v-if="testingEmailProviderID !== null" class="size-4 animate-spin" />
              <Send v-else class="size-4" />
              发送测试邮件
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Mail, Pencil, Plus, RefreshCw, Send, Star } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { useAuthStore } from '@/stores/auth'
import emailProviderSettingsApi, {
  getEmailProviderSettingsErrorMessage,
  type EmailProviderSettingsRecord,
  type SaveEmailProviderSettingsInput,
} from '@/api/emailProviderSettingsApi'

interface EmailProviderSettingsFormState extends SaveEmailProviderSettingsInput {
  id?: number
  hasExistingPassword: boolean
}

const authStore = useAuthStore()
const canEditEmailProviderSettings = computed(() => authStore.hasPermission('settings:edit'))
const loadingEmailProviderSettings = ref(false)
const savingEmailProviderSettings = ref(false)
const emailProviderSettingsRecords = ref<EmailProviderSettingsRecord[]>([])
const busyEmailProviderID = ref<number | null>(null)
const testingEmailProviderID = ref<number | null>(null)
const emailProviderSettingsDialogOpen = ref(false)
const emailProviderTestDialogOpen = ref(false)
const emailProviderTestTargetEmail = ref('')
const selectedEmailProviderForTest = ref<EmailProviderSettingsRecord | null>(null)

const createEmptyEmailProviderSettingsForm = (): EmailProviderSettingsFormState => ({
  id: undefined,
  code: 'primary_smtp',
  name: '',
  driver: 'smtp',
  host: '',
  port: 587,
  username: '',
  password: '',
  from_name: '',
  from_email: '',
  reply_to: '',
  encryption_type: 'starttls',
  is_active: true,
  is_default: false,
  hasExistingPassword: false,
})

const emailProviderSettingsForm = reactive<EmailProviderSettingsFormState>(createEmptyEmailProviderSettingsForm())

const defaultEmailProvider = computed(() => emailProviderSettingsRecords.value.find((provider) => provider.is_default && provider.is_active) || null)

const replaceEmailProviderSettingsRecord = (updatedProvider: EmailProviderSettingsRecord): void => {
  emailProviderSettingsRecords.value = emailProviderSettingsRecords.value.map((provider) => (
    provider.id === updatedProvider.id ? updatedProvider : provider
  ))
}

const upsertEmailProviderSettingsRecord = (updatedProvider: EmailProviderSettingsRecord): void => {
  const hasUpdatedProvider = emailProviderSettingsRecords.value.some((provider) => provider.id === updatedProvider.id)
  const records = hasUpdatedProvider
    ? emailProviderSettingsRecords.value.map((provider) => provider.id === updatedProvider.id ? updatedProvider : provider)
    : [...emailProviderSettingsRecords.value, updatedProvider]

  emailProviderSettingsRecords.value = records.map((provider) => (
    updatedProvider.is_default && provider.id !== updatedProvider.id
      ? { ...provider, is_default: false }
      : provider
  ))
}

const loadEmailProviderSettings = async (): Promise<void> => {
  loadingEmailProviderSettings.value = true
  try {
    emailProviderSettingsRecords.value = await emailProviderSettingsApi.listEmailProviderSettings()
  } catch (error) {
    toast.error(getEmailProviderSettingsErrorMessage(error))
  } finally {
    loadingEmailProviderSettings.value = false
  }
}

const openCreateEmailProviderDialog = (): void => {
  Object.assign(emailProviderSettingsForm, createEmptyEmailProviderSettingsForm())
  emailProviderSettingsDialogOpen.value = true
}

const openEditEmailProviderDialog = (provider: EmailProviderSettingsRecord): void => {
  Object.assign(emailProviderSettingsForm, {
    id: provider.id,
    code: provider.code,
    name: provider.name,
    driver: provider.driver || 'smtp',
    host: provider.host,
    port: provider.port,
    username: provider.username,
    password: '',
    from_name: provider.from_name,
    from_email: provider.from_email,
    reply_to: provider.reply_to || '',
    encryption_type: provider.encryption_type,
    is_active: provider.is_active,
    is_default: provider.is_default,
    hasExistingPassword: provider.has_password,
  })
  emailProviderSettingsDialogOpen.value = true
}

const buildEmailProviderSettingsPayload = (): SaveEmailProviderSettingsInput => {
  const payload: SaveEmailProviderSettingsInput = {
    code: emailProviderSettingsForm.code.trim(),
    name: emailProviderSettingsForm.name.trim(),
    driver: 'smtp',
    host: emailProviderSettingsForm.host.trim(),
    port: Number(emailProviderSettingsForm.port),
    username: emailProviderSettingsForm.username.trim(),
    from_name: emailProviderSettingsForm.from_name.trim(),
    from_email: emailProviderSettingsForm.from_email.trim(),
    reply_to: emailProviderSettingsForm.reply_to?.trim() || '',
    encryption_type: emailProviderSettingsForm.encryption_type,
    is_active: emailProviderSettingsForm.is_active,
    is_default: emailProviderSettingsForm.is_default,
  }
  const password = emailProviderSettingsForm.password?.trim()
  if (password) payload.password = password
  return payload
}

const saveEmailProviderSettings = async (): Promise<void> => {
  if (!canEditEmailProviderSettings.value) return
  const payload = buildEmailProviderSettingsPayload()
  if (!payload.code || !payload.name || !payload.host || !payload.from_name || !payload.from_email) {
    toast.error('请填写通道编码、名称、SMTP 主机和发件身份')
    return
  }
  if (!Number.isInteger(payload.port) || payload.port < 1 || payload.port > 65535) {
    toast.error('SMTP 端口必须是 1 到 65535 之间的整数')
    return
  }
  savingEmailProviderSettings.value = true
  try {
    const savedProvider = emailProviderSettingsForm.id
      ? await emailProviderSettingsApi.updateEmailProviderSettings(emailProviderSettingsForm.id, payload)
      : await emailProviderSettingsApi.createEmailProviderSettings(payload)
    upsertEmailProviderSettingsRecord(savedProvider)
    emailProviderSettingsDialogOpen.value = false
    toast.success('邮件发件通道已保存')
  } catch (error) {
    toast.error(getEmailProviderSettingsErrorMessage(error))
  } finally {
    savingEmailProviderSettings.value = false
  }
}

const toggleEmailProviderActive = async (provider: EmailProviderSettingsRecord, active: boolean): Promise<void> => {
  if (!canEditEmailProviderSettings.value || busyEmailProviderID.value !== null) return
  busyEmailProviderID.value = provider.id
  try {
    const updatedProvider = await emailProviderSettingsApi.setEmailProviderSettingsActive(provider.id, active)
    replaceEmailProviderSettingsRecord(updatedProvider)
    toast.success(active ? '邮件通道已启用' : '邮件通道已停用')
  } catch (error) {
    toast.error(getEmailProviderSettingsErrorMessage(error))
  } finally {
    busyEmailProviderID.value = null
  }
}

const setEmailProviderAsDefault = async (provider: EmailProviderSettingsRecord): Promise<void> => {
  if (!canEditEmailProviderSettings.value || busyEmailProviderID.value !== null) return
  busyEmailProviderID.value = provider.id
  try {
    const updatedProvider = await emailProviderSettingsApi.setDefaultEmailProviderSettings(provider.id)
    emailProviderSettingsRecords.value = emailProviderSettingsRecords.value.map((item) => ({
      ...item,
      is_default: item.id === updatedProvider.id,
      is_active: item.id === updatedProvider.id ? updatedProvider.is_active : item.is_active,
    }))
    toast.success('默认邮件通道已切换')
  } catch (error) {
    toast.error(getEmailProviderSettingsErrorMessage(error))
  } finally {
    busyEmailProviderID.value = null
  }
}

const openEmailProviderTestDialog = (provider: EmailProviderSettingsRecord): void => {
  selectedEmailProviderForTest.value = provider
  emailProviderTestTargetEmail.value = authStore.user?.email || ''
  emailProviderTestDialogOpen.value = true
}

const sendEmailProviderTestMessage = async (): Promise<void> => {
  const provider = selectedEmailProviderForTest.value
  const targetEmail = emailProviderTestTargetEmail.value.trim()
  if (!provider || !targetEmail) {
    toast.error('请输入测试收件地址')
    return
  }
  testingEmailProviderID.value = provider.id
  try {
    const updatedProvider = await emailProviderSettingsApi.sendEmailProviderSettingsTestEmail(provider.id, targetEmail)
    replaceEmailProviderSettingsRecord(updatedProvider)
    emailProviderTestDialogOpen.value = false
    toast.success('测试邮件已发送，请检查收件箱和垃圾邮件文件夹')
  } catch (error) {
    toast.error(getEmailProviderSettingsErrorMessage(error))
    await loadEmailProviderSettings()
  } finally {
    testingEmailProviderID.value = null
  }
}

const emailProviderTestTone = (provider: EmailProviderSettingsRecord): AdminStatusTone => {
  if (provider.last_test_status === 'healthy') return 'green'
  if (provider.last_test_status === 'failed') return 'coral'
  return 'amber'
}

const emailProviderTestLabel = (provider: EmailProviderSettingsRecord): string => {
  if (provider.last_test_status === 'healthy') return '测试正常'
  if (provider.last_test_status === 'failed') return '测试失败'
  return '未测试'
}

const formatEmailProviderTestTime = (value?: string): string => {
  if (!value) return '尚未执行连通性测试'
  const parsedDate = new Date(value)
  return Number.isNaN(parsedDate.getTime()) ? '测试时间不可用' : parsedDate.toLocaleString('zh-CN')
}

onMounted(() => {
  void loadEmailProviderSettings()
})
</script>
