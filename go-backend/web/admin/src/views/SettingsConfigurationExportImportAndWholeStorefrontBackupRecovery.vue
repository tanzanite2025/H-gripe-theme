<template>
  <div class="min-w-0 space-y-5">
    <AdminPageHeader title="系统设置 / 备份与恢复" description="将业务数据和图片保存到电脑，上传备份进行恢复验证；普通配置迁移单独管理。">
      <template #actions>
        <Button variant="outline" :disabled="loadingOverview" @click="refreshConfigurationBackupOverview">刷新状态</Button>
      </template>
    </AdminPageHeader>

    <p v-if="overviewError" role="alert" class="rounded-lg border border-destructive/30 p-4 text-sm text-destructive">{{ overviewError }}</p>

    <Tabs v-model="activeBackupTab" class="gap-5">
      <TabsList class="h-auto max-w-full flex-wrap justify-start">
        <TabsTrigger value="whole-storefront">业务数据备份与恢复</TabsTrigger>
        <TabsTrigger value="configuration">普通配置迁移</TabsTrigger>
      </TabsList>

      <TabsContent value="configuration" class="space-y-5">
        <section class="rounded-xl border bg-card p-5 sm:p-6">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div class="space-y-2">
              <h2 class="text-base font-semibold">导出当前配置</h2>
              <p class="max-w-2xl text-sm leading-6 text-muted-foreground">下载已保存配置的 JSON 文件，包含各语言版本。可以上传这份文件，恢复其中的配置项。</p>
            </div>
            <Button :disabled="exportingConfiguration || !overview?.configuration_export_available || overview.configuration_count === 0" @click="downloadCurrentSettingsConfiguration">
              {{ exportingConfiguration ? '正在导出…' : '导出配置到电脑' }}
            </Button>
          </div>
          <div v-if="overview" class="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <div v-for="scope in overview.scopes" :key="scope.group" class="rounded-lg border px-4 py-3">
              <p class="text-sm">{{ scope.label }}</p>
              <p class="mt-1 text-xl font-semibold tabular-nums">{{ scope.count }} <span class="text-xs font-normal text-muted-foreground">项</span></p>
            </div>
          </div>
          <p class="mt-4 text-sm leading-6 text-muted-foreground">此文件只包含上方列出的普通配置。付款渠道、市场与税率规则、积分规则、独立邮件通道、密码和服务器密钥需通过整站加密备份及独立恢复材料保存。图片只保存引用，不包含图片文件。</p>
          <p v-if="overview" class="mt-2 text-xs text-muted-foreground">可导出 {{ overview.configuration_count }} 项；另有 {{ overview.excluded_setting_count }} 项保存记录不在此配置文件范围内。</p>
          <p v-if="configurationDownloadMessage" role="status" class="mt-3 text-sm text-primary">{{ configurationDownloadMessage }}</p>
        </section>

        <section class="space-y-4 rounded-xl border bg-card p-5 sm:p-6">
          <div class="space-y-2">
            <h2 class="text-base font-semibold">从电脑导入配置</h2>
            <p class="text-sm leading-6 text-muted-foreground">选择此前导出的 JSON 文件。先查看新增、更新和不变的项目，确认后再保存；文件中没有的配置保持当前值。</p>
          </div>
          <input ref="configurationFileInput" type="file" accept=".json,application/json" aria-label="选择配置 JSON 文件" class="block w-full max-w-full rounded-lg border p-3 text-sm" :disabled="inspectingConfiguration || importingConfiguration || !overview" @change="selectAndPreviewSettingsConfigurationFile" />
          <p class="text-xs text-muted-foreground">文件最大 8 MiB；仅导入本站支持的配置项。引用的图片必须已在当前商城存在。</p>
          <p v-if="selectedConfigurationFilename" class="break-all text-sm text-muted-foreground">已选择：{{ selectedConfigurationFilename }}</p>
          <p v-if="inspectingConfiguration" role="status" class="text-sm">正在校验文件并读取当前配置…</p>
          <p v-if="configurationImportError" role="alert" class="break-words text-sm text-destructive">{{ configurationImportError }}</p>
          <p v-if="configurationImportMessage" role="status" class="text-sm text-primary">{{ configurationImportMessage }}</p>

          <div v-if="configurationPreview" class="space-y-4">
            <div class="flex flex-wrap gap-3 text-sm">
              <span>新增 {{ configurationPreview.create_count }} 项</span>
              <span>更新 {{ configurationPreview.update_count }} 项</span>
              <span>不变 {{ configurationPreview.unchanged_count }} 项</span>
            </div>
            <div class="max-w-full overflow-x-auto rounded-lg border">
              <table class="w-full min-w-[600px] table-fixed text-left text-sm">
                <thead class="bg-muted/40"><tr><th class="w-1/4 p-3">配置 / 语言</th><th class="w-20 p-3">变更</th><th class="p-3">当前值</th><th class="p-3">导入值</th></tr></thead>
                <tbody><tr v-for="change in visibleConfigurationChanges" :key="`${change.key}:${change.locale}`" class="border-t align-top">
                  <td class="break-all p-3">{{ change.key }}<span class="mt-1 block text-xs text-muted-foreground">{{ change.group }} · {{ change.locale }}</span></td>
                  <td class="p-3">{{ configurationChangeLabel(change.action) }}</td>
                  <td class="p-3"><p class="line-clamp-4 break-all whitespace-pre-wrap">{{ change.current_value || '（空）' }}</p><p v-if="change.action !== 'create'" class="mt-2 break-all text-xs text-muted-foreground">{{ change.current_type }} · {{ change.current_is_public ? '公开' : '非公开' }}<br />说明：{{ change.current_description || '（无）' }}</p></td>
                  <td class="p-3"><p class="line-clamp-4 break-all whitespace-pre-wrap">{{ change.imported_value || '（空）' }}</p><p class="mt-2 break-all text-xs text-muted-foreground">{{ change.imported_type }} · {{ change.imported_is_public ? '公开' : '非公开' }}<br />说明：{{ change.imported_description || '（无）' }}</p></td>
                </tr></tbody>
              </table>
            </div>
            <div v-if="configurationPreview.changes.length > configurationPreviewPageSize" class="flex flex-wrap items-center gap-3 text-sm">
              <Button variant="outline" :disabled="configurationPreviewPage === 0" @click="configurationPreviewPage--">上一页</Button>
              <span>第 {{ configurationPreviewPage + 1 }} / {{ Math.ceil(configurationPreview.changes.length / configurationPreviewPageSize) }} 页</span>
              <Button variant="outline" :disabled="(configurationPreviewPage + 1) * configurationPreviewPageSize >= configurationPreview.changes.length" @click="configurationPreviewPage++">下一页</Button>
            </div>
            <div v-if="canImportSettingsConfiguration" class="space-y-3 rounded-lg border p-4">
              <label class="flex items-start gap-2 text-sm leading-6"><input v-model="configurationOverwriteConfirmed" type="checkbox" class="mt-1" :disabled="importingConfiguration" />我已核对以上变更，确认用文件中的值覆盖对应配置。</label>
              <Button :disabled="!configurationOverwriteConfirmed || importingConfiguration || inspectingConfiguration || configurationPreview.create_count + configurationPreview.update_count === 0" @click="applyPreviewedSettingsConfiguration">
                {{ importingConfiguration ? '正在导入…' : '确认导入配置' }}
              </Button>
            </div>
            <p v-else class="text-sm text-muted-foreground">当前账号可查看预览，导入配置需要系统设置编辑权限。</p>
          </div>
        </section>
      </TabsContent>

      <TabsContent value="whole-storefront" class="space-y-4">
        <p v-if="loadingOverview" role="status" class="text-sm text-muted-foreground">正在读取业务备份状态…</p>
        <section v-else-if="!authStore.hasPermission('ops:backup:view')" class="rounded-xl border p-5 text-sm">当前账号没有业务备份查看权限，可以使用普通配置迁移功能。</section>
        <section v-else-if="overview && !overview.whole_storefront_maintenance_configured" class="space-y-3 rounded-xl border bg-card p-5 sm:p-6">
          <h2 class="text-base font-semibold">业务备份服务尚未启用</h2>
          <p class="text-sm leading-6 text-muted-foreground">整站包保存订单、客户、客服历史、业务配置和附件；商品与实际图片可以独立导出并回传隔离验证。部署独立服务后，这里会显示对应操作和备份记录。</p>
          <p class="text-sm leading-6 text-muted-foreground">需要先完成独立备份服务、恢复凭据和维护窗口的部署。配置导出与导入可以独立使用。</p>
          <p class="text-sm leading-6 text-muted-foreground">Hostinger 周备份仍由提供商单独管理，不会混入电脑上的应用备份。</p>
        </section>
        <LocalComputerBackupAndConsistentRecovery v-else-if="overview?.whole_storefront_maintenance_configured" />
      </TabsContent>
    </Tabs>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import LocalComputerBackupAndConsistentRecovery from '@/views/LocalComputerBackupAndConsistentRecovery.vue'
import { useAuthStore } from '@/stores/auth'
import { settingsConfigurationBackupAndRecoveryApi } from '@/api/settingsConfigurationBackupAndRecovery'
import type { PortableSettingsConfigurationFile, SettingsConfigurationBackupOverview, SettingsConfigurationImportPreview } from '@/api/settingsConfigurationBackupAndRecovery'

const authStore = useAuthStore()
const route = useRoute()
const activeBackupTab = ref(route.query.tab === 'configuration' ? 'configuration' : 'whole-storefront')
const overview = ref<SettingsConfigurationBackupOverview | null>(null)
const loadingOverview = ref(false)
const overviewError = ref('')
const exportingConfiguration = ref(false)
const configurationDownloadMessage = ref('')
const configurationFileInput = ref<HTMLInputElement | null>(null)
const selectedConfigurationFilename = ref('')
const selectedConfigurationBackup = ref<PortableSettingsConfigurationFile | null>(null)
const configurationPreview = ref<SettingsConfigurationImportPreview | null>(null)
const configurationPreviewPage = ref(0)
const configurationPreviewPageSize = 50
const inspectingConfiguration = ref(false)
const importingConfiguration = ref(false)
const configurationOverwriteConfirmed = ref(false)
const configurationImportError = ref('')
const configurationImportMessage = ref('')
const canImportSettingsConfiguration = computed(() => authStore.hasPermission('settings:edit'))
const visibleConfigurationChanges = computed(() => configurationPreview.value?.changes.slice(configurationPreviewPage.value * configurationPreviewPageSize, (configurationPreviewPage.value + 1) * configurationPreviewPageSize) || [])

function configurationOperationErrorMessage(error: unknown, fallback: string): string {
  const responseError = (error as { response?: { data?: { error?: unknown } } })?.response?.data?.error
  if (typeof responseError === 'string') return responseError
  return error instanceof Error ? error.message : fallback
}

async function refreshConfigurationBackupOverview() {
  loadingOverview.value = true
  overviewError.value = ''
  try { overview.value = await settingsConfigurationBackupAndRecoveryApi.getOverview() }
  catch (error) { overviewError.value = configurationOperationErrorMessage(error, '无法读取备份与恢复状态') }
  finally { loadingOverview.value = false }
}

async function downloadCurrentSettingsConfiguration() {
  exportingConfiguration.value = true
  configurationDownloadMessage.value = ''
  try {
    const file = await settingsConfigurationBackupAndRecoveryApi.exportConfiguration()
    const blob = new Blob([JSON.stringify(file)], { type: 'application/json' })
    const downloadUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = `storefront-configuration-${new Date().toISOString().replace(/[:.]/g, '-')}.json`
    document.body.append(link)
    link.click()
    link.remove()
    window.setTimeout(() => URL.revokeObjectURL(downloadUrl), 1000)
    configurationDownloadMessage.value = `已发起 ${file.settings.length} 项配置的文件下载，请确认电脑下载目录中的实际文件。`
  } catch (error) { overviewError.value = configurationOperationErrorMessage(error, '配置导出失败') }
  finally { exportingConfiguration.value = false }
}

async function selectAndPreviewSettingsConfigurationFile(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  configurationPreview.value = null
  selectedConfigurationBackup.value = null
  configurationOverwriteConfirmed.value = false
  configurationImportError.value = ''
  configurationImportMessage.value = ''
  configurationPreviewPage.value = 0
  selectedConfigurationFilename.value = file?.name || ''
  if (!file) return
  if (file.size > (overview.value?.max_configuration_file_bytes || 8 * 1024 * 1024)) { configurationImportError.value = '配置文件超过 8 MiB 上限'; return }
  inspectingConfiguration.value = true
  try {
    let backup: PortableSettingsConfigurationFile
    try { backup = JSON.parse(await file.text()) as PortableSettingsConfigurationFile }
    catch { throw new Error('文件不是有效的 JSON 配置文件') }
    const preview = await settingsConfigurationBackupAndRecoveryApi.previewConfiguration(backup)
    selectedConfigurationBackup.value = backup
    configurationPreview.value = preview
  } catch (error) { configurationImportError.value = configurationOperationErrorMessage(error, '配置文件校验失败') }
  finally { inspectingConfiguration.value = false }
}

async function applyPreviewedSettingsConfiguration() {
  if (!configurationPreview.value || !selectedConfigurationBackup.value || !configurationOverwriteConfirmed.value || !canImportSettingsConfiguration.value || importingConfiguration.value) return
  importingConfiguration.value = true
  configurationImportError.value = ''
  try {
    const result = await settingsConfigurationBackupAndRecoveryApi.importConfiguration(selectedConfigurationBackup.value, configurationPreview.value.target_sha256)
    configurationImportMessage.value = `已导入 ${result.count} 项配置。文件中没有的配置保持当前值。`
    configurationPreview.value = null
    selectedConfigurationBackup.value = null
    configurationOverwriteConfirmed.value = false
    if (configurationFileInput.value) configurationFileInput.value.value = ''
    selectedConfigurationFilename.value = ''
    await refreshConfigurationBackupOverview()
  } catch (error) {
    configurationImportError.value = configurationOperationErrorMessage(error, '导入失败，请重新选择文件并预览')
    configurationPreview.value = null
    selectedConfigurationBackup.value = null
    configurationOverwriteConfirmed.value = false
    if (configurationFileInput.value) configurationFileInput.value.value = ''
  } finally { importingConfiguration.value = false }
}

function configurationChangeLabel(action: string): string {
  return ({ create: '新增', update: '更新', unchanged: '不变' } as Record<string, string>)[action] || action
}

onMounted(refreshConfigurationBackupOverview)
</script>
