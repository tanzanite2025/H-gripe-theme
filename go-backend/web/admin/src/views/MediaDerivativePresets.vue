<template>
  <div class="space-y-4">
    <AdminPageHeader title="图片尺寸转换" description="管理图片派生尺寸定义">
      <template #actions>
        <Button
          size="icon"
          variant="outline"
          title="刷新转换定义"
          :disabled="loading"
          @click="loadPresets"
        >
          <RefreshCw :class="['size-4', { 'animate-spin': loading }]" />
        </Button>
        <Button v-if="canConfigure" @click="openCreate">
          <Plus class="size-4" />
          添加转换
        </Button>
        <Button
          v-if="canConfigure"
          variant="outline"
          :disabled="requestingRebuild"
          @click="requestRebuild"
        >
          <LoaderCircle v-if="requestingRebuild" class="size-4 animate-spin" />
          <RotateCw v-else class="size-4" />
          重建图片
        </Button>
      </template>
    </AdminPageHeader>

    <section class="overflow-hidden border bg-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
        <h2 class="text-sm font-black">转换定义</h2>
        <span class="font-mono text-xs font-bold text-muted-foreground">{{ presets.length }} 项</span>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[900px] text-sm">
          <thead class="border-b bg-muted/30 text-left text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            <tr>
              <th class="px-4 py-3">转换</th>
              <th class="px-4 py-3">最大宽度</th>
              <th class="px-4 py-3">版本</th>
              <th class="px-4 py-3">已生成</th>
              <th class="px-4 py-3">类型</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <tr v-if="loading">
              <td colspan="7" class="px-4 py-12 text-center text-sm text-muted-foreground">正在加载转换定义</td>
            </tr>
            <tr v-else-if="presets.length === 0">
              <td colspan="7" class="px-4 py-12 text-center text-sm text-muted-foreground">暂未定义图片尺寸转换</td>
            </tr>
            <tr v-for="preset in presets" :key="preset.id">
              <td class="px-4 py-3">
                <p class="font-bold">{{ preset.label }}</p>
                <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ preset.code }}</p>
              </td>
              <td class="px-4 py-3 font-mono text-xs font-bold">{{ preset.max_width }} px</td>
              <td class="px-4 py-3 font-mono text-xs">v{{ preset.generation_version }}</td>
              <td class="px-4 py-3 font-mono text-xs">{{ formatCount(preset.generated_derivatives) }}</td>
              <td class="px-4 py-3">
                <AdminStatusBadge :tone="preset.is_system ? 'blue' : 'gray'">
                  {{ preset.is_system ? '系统' : '自定义' }}
                </AdminStatusBadge>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <Switch
                    size="sm"
                    :model-value="preset.enabled"
                    :disabled="!canConfigure || updatingEnabledID === preset.id"
                    :aria-label="`${preset.label}启用状态`"
                    @update:model-value="updateEnabled(preset, Boolean($event))"
                  />
                  <AdminStatusBadge :tone="preset.enabled ? 'green' : 'gray'">
                    {{ preset.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </div>
              </td>
              <td class="px-4 py-3 text-right">
                <div class="flex justify-end gap-1">
                  <Button
                    v-if="canConfigure"
                    size="icon"
                    variant="ghost"
                    :title="`编辑${preset.label}`"
                    :aria-label="`编辑${preset.label}`"
                    @click="openEdit(preset)"
                  >
                    <Pencil class="size-4" />
                  </Button>
                  <Button
                    v-if="canConfigure && !preset.is_system"
                    size="icon"
                    variant="ghost"
                    :disabled="preset.generated_derivatives > 0 || deletingID === preset.id"
                    :title="preset.generated_derivatives > 0 ? '已有派生文件，只能停用' : `删除${preset.label}`"
                    :aria-label="`删除${preset.label}`"
                    @click="removePreset(preset)"
                  >
                    <LoaderCircle v-if="deletingID === preset.id" class="size-4 animate-spin" />
                    <Trash2 v-else class="size-4 text-destructive" />
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="overflow-hidden border bg-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
        <h2 class="text-sm font-black">重建队列</h2>
        <span class="font-mono text-xs font-bold text-muted-foreground">{{ rebuildJobs.length }} 条</span>
      </div>
      <div v-if="rebuildJobs.length === 0" class="px-4 py-8 text-sm text-muted-foreground">暂无图片重建任务</div>
      <div v-else class="divide-y">
        <div v-for="job in rebuildJobs" :key="job.id" class="grid gap-2 px-4 py-3 text-sm md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <AdminStatusBadge :tone="rebuildJobTone(job.status)">{{ rebuildJobLabel(job.status) }}</AdminStatusBadge>
              <span class="font-mono text-xs text-muted-foreground">#{{ job.id }}</span>
              <span class="text-xs text-muted-foreground">{{ job.scanned_assets }} 已扫描 / {{ job.generated_derivatives }} 已生成 / {{ job.failed_assets }} 失败</span>
            </div>
            <p v-if="job.last_error" class="mt-1 truncate text-xs text-destructive" :title="job.last_error">{{ job.last_error }}</p>
          </div>
          <span class="font-mono text-xs text-muted-foreground">{{ formatJobTime(job.updated_at || job.created_at) }}</span>
        </div>
      </div>
    </section>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="max-w-xl">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑尺寸转换' : '添加尺寸转换' }}</DialogTitle>
          <DialogDescription>尺寸调整会创建新的派生版本，等待上线前检查补齐。</DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="savePreset">
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminFormField label="名称" required>
              <Input v-model="form.label" placeholder="例如 商品横幅" />
            </AdminFormField>
            <AdminFormField label="代码" required>
              <Input
                v-model="form.code"
                class="font-mono"
                :disabled="Boolean(form.id)"
                placeholder="例如 product-banner"
              />
            </AdminFormField>
            <AdminFormField label="最大宽度" required>
              <Input v-model.number="form.max_width" type="number" min="1" max="8000" />
            </AdminFormField>
            <AdminFormField label="排序">
              <Input v-model.number="form.sort_order" type="number" min="-100000" max="100000" />
            </AdminFormField>
            <div class="flex items-center justify-between gap-4 border px-3 py-2.5 sm:col-span-2">
              <span class="text-sm">启用转换</span>
              <Switch v-model="form.enabled" />
            </div>
          </div>
          <p
            v-if="form.id && form.max_width !== originalMaxWidth"
            class="text-xs font-medium text-amber-700"
          >
            保存后将从 v{{ form.generation_version }} 更新到 v{{ form.generation_version + 1 }}。
          </p>
          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { LoaderCircle, Pencil, Plus, RefreshCw, RotateCw, Trash2 } from '@lucide/vue'
import { toast } from 'vue-sonner'
import { mediaApi, type MediaDerivativePreset, type MediaDerivativePresetInput, type MediaDerivativeRebuildJob } from '@/api/media'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { useAuthStore } from '@/stores/auth'

interface DerivativePresetForm {
  id?: number
  code: string
  label: string
  max_width: number
  sort_order: number
  enabled: boolean
  generation_version: number
}

const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const presets = ref<MediaDerivativePreset[]>([])
const rebuildJobs = ref<MediaDerivativeRebuildJob[]>([])
const updatingEnabledID = ref<number | null>(null)
const deletingID = ref<number | null>(null)
const requestingRebuild = ref(false)
const originalMaxWidth = ref(0)
const form = reactive<DerivativePresetForm>({
  id: undefined,
  code: '',
  label: '',
  max_width: 640,
  sort_order: 100,
  enabled: true,
  generation_version: 1,
})

const canConfigure = computed(() => authStore.hasPermission('media:configure'))

const resetForm = (): void => {
  Object.assign(form, {
    id: undefined,
    code: '',
    label: '',
    max_width: 640,
    sort_order: 100,
    enabled: true,
    generation_version: 1,
  })
  originalMaxWidth.value = 0
}

const openCreate = (): void => {
  resetForm()
  dialogOpen.value = true
}

const openEdit = (preset: MediaDerivativePreset): void => {
  Object.assign(form, {
    id: preset.id,
    code: preset.code,
    label: preset.label,
    max_width: preset.max_width,
    sort_order: preset.sort_order,
    enabled: preset.enabled,
    generation_version: preset.generation_version,
  })
  originalMaxWidth.value = preset.max_width
  dialogOpen.value = true
}

const loadPresets = async (): Promise<void> => {
  loading.value = true
  try {
    const [nextPresets, nextJobs] = await Promise.all([
      mediaApi.listDerivativePresets(),
      mediaApi.listDerivativeRebuildJobs(),
    ])
    presets.value = nextPresets
    rebuildJobs.value = nextJobs
  } catch (error) {
    toast.error(errorMessage(error, '图片尺寸转换加载失败'))
  } finally {
    loading.value = false
  }
}

const requestRebuild = async (): Promise<void> => {
  requestingRebuild.value = true
  try {
    const job = await mediaApi.requestDerivativeRebuild()
    toast.success(`图片重建任务 #${job.id} 已进入队列`)
    await loadPresets()
  } catch (error) {
    toast.error(errorMessage(error, '图片重建任务创建失败'))
  } finally {
    requestingRebuild.value = false
  }
}

const savePreset = async (): Promise<void> => {
  const code = form.code.trim()
  const label = form.label.trim()
  const maxWidth = Number(form.max_width)
  const sortOrder = Number(form.sort_order)
  if (!label || !Number.isInteger(maxWidth) || maxWidth < 1 || maxWidth > 8000 || !Number.isInteger(sortOrder)) {
    toast.error('请填写有效的转换名称、最大宽度和排序')
    return
  }
  if (!form.id && !code) {
    toast.error('请填写转换代码')
    return
  }

  saving.value = true
  try {
    if (form.id) {
      await mediaApi.updateDerivativePreset(form.id, {
        label,
        max_width: maxWidth,
        sort_order: sortOrder,
        enabled: form.enabled,
      })
    } else {
      const payload: MediaDerivativePresetInput = {
        code,
        label,
        max_width: maxWidth,
        sort_order: sortOrder,
        enabled: form.enabled,
      }
      await mediaApi.createDerivativePreset(payload)
    }
    dialogOpen.value = false
    toast.success('尺寸转换已保存')
    await loadPresets()
  } catch (error) {
    toast.error(errorMessage(error, '尺寸转换保存失败'))
  } finally {
    saving.value = false
  }
}

const updateEnabled = async (preset: MediaDerivativePreset, enabled: boolean): Promise<void> => {
  if (!canConfigure.value || updatingEnabledID.value || preset.enabled === enabled) return
  updatingEnabledID.value = preset.id
  try {
    await mediaApi.updateDerivativePresetEnabled(preset.id, enabled)
    toast.success(enabled ? '尺寸转换已启用' : '尺寸转换已停用')
    await loadPresets()
  } catch (error) {
    toast.error(errorMessage(error, '尺寸转换状态更新失败'))
  } finally {
    updatingEnabledID.value = null
  }
}

const removePreset = async (preset: MediaDerivativePreset): Promise<void> => {
  if (preset.generated_derivatives > 0) {
    toast.error('已有派生文件的转换不能删除，请先停用')
    return
  }
  if (!window.confirm(`确定删除尺寸转换“${preset.label}”？`)) return

  deletingID.value = preset.id
  try {
    await mediaApi.deleteDerivativePreset(preset.id)
    toast.success('尺寸转换已删除')
    await loadPresets()
  } catch (error) {
    toast.error(errorMessage(error, '尺寸转换删除失败'))
  } finally {
    deletingID.value = null
  }
}

const formatCount = (value: number): string => Number(value || 0).toLocaleString()

const rebuildJobLabel = (status: MediaDerivativeRebuildJob['status']): string => {
  if (status === 'pending') return '等待中'
  if (status === 'running') return '重建中'
  return '已完成'
}

const rebuildJobTone = (status: MediaDerivativeRebuildJob['status']): 'gray' | 'blue' | 'green' => {
  if (status === 'pending') return 'gray'
  if (status === 'running') return 'blue'
  return 'green'
}

const formatJobTime = (value?: string): string => {
  if (!value) return '刚刚'
  const timestamp = new Date(value)
  return Number.isNaN(timestamp.getTime()) ? value : timestamp.toLocaleString()
}

const errorMessage = (error: unknown, fallback: string): string => {
  const candidate = error as { response?: { data?: { error?: unknown } } }
  return typeof candidate.response?.data?.error === 'string'
    ? candidate.response.data.error
    : fallback
}

onMounted(() => {
  void loadPresets()
})
</script>
