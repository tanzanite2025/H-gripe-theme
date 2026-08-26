<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="t('fitmentCatalog.hub.title')"
      :description="t('fitmentCatalog.hub.description')"
    >
      <template #actions>
        <Button v-if="canCreate" @click="openCreate">
          <Plus class="size-4" />
          {{ t('fitmentCatalog.hub.create') }}
        </Button>
      </template>
    </AdminPageHeader>

    <AdminTabBar
      :tabs="fitmentTabs"
      :active-path="route.path"
      :label="t('fitmentCatalog.tabsLabel')"
    />

    <AdminFilterPanel>
      <form class="grid gap-3 md:grid-cols-[minmax(220px,1.35fr)_minmax(140px,0.55fr)_minmax(140px,0.55fr)_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            {{ t('fitmentCatalog.searchLabel') }}
          </span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" :placeholder="t('fitmentCatalog.hub.searchPlaceholder')" />
          </div>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            {{ t('fitmentCatalog.positionLabel') }}
          </span>
          <Select v-model="filters.position">
            <SelectTrigger class="h-9"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{{ t('fitmentCatalog.hubPosition.all') }}</SelectItem>
              <SelectItem value="front">{{ t('fitmentCatalog.hubPosition.front') }}</SelectItem>
              <SelectItem value="rear">{{ t('fitmentCatalog.hubPosition.rear') }}</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            {{ t('fitmentCatalog.statusLabel') }}
          </span>
          <Select v-model="filters.status">
            <SelectTrigger class="h-9"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{{ t('fitmentCatalog.allStatuses') }}</SelectItem>
              <SelectItem value="enabled">{{ t('fitmentCatalog.enabled') }}</SelectItem>
              <SelectItem value="disabled">{{ t('fitmentCatalog.disabled') }}</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 text-xs font-black uppercase tracking-wider">
            <Search class="size-3.5" />
            {{ t('fitmentCatalog.search') }}
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetFilters">
            <RotateCcw class="size-3.5" />
            {{ t('fitmentCatalog.reset') }}
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1180px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-64">{{ t('fitmentCatalog.columns.hubNameCode') }}</TableHead>
            <TableHead class="w-28">{{ t('fitmentCatalog.columns.position') }}</TableHead>
            <TableHead class="w-36">{{ t('fitmentCatalog.columns.axleType') }}</TableHead>
            <TableHead class="w-28">{{ t('fitmentCatalog.columns.axleSpacing') }}</TableHead>
            <TableHead class="w-28">{{ t('fitmentCatalog.columns.fitmentReferences') }}</TableHead>
            <TableHead class="w-24">{{ t('fitmentCatalog.columns.status') }}</TableHead>
            <TableHead class="w-36">{{ t('fitmentCatalog.columns.updatedAt') }}</TableHead>
            <TableHead class="w-20 text-right">{{ t('fitmentCatalog.columns.actions') }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="specifications.length === 0" :colspan="8">
            <div class="flex flex-col items-center text-muted-foreground">
              <Database class="mb-2 size-7 opacity-55" />
              <span class="text-xs">{{ filters.search ? t('fitmentCatalog.hub.emptySearch') : t('fitmentCatalog.hub.empty') }}</span>
            </div>
          </TableEmpty>

          <TableRow v-for="specification in specifications" :key="specification.id">
            <TableCell>
              <div class="min-w-0">
                <p class="truncate text-xs font-black">{{ specification.display_name }}</p>
                <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">{{ specification.spec_code }}</p>
              </div>
            </TableCell>
            <TableCell>
              <span class="inline-flex rounded-full bg-muted px-2 py-1 text-[10px] font-black">
                {{ formatPosition(specification.position) }}
              </span>
            </TableCell>
            <TableCell class="text-xs font-bold">
              {{ formatAxleType(specification.axle_type) }}
            </TableCell>
            <TableCell class="font-mono text-xs font-bold">
              {{ specification.axle_spacing_mm }} mm
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">
              {{ referenceCount(specification) }}
            </TableCell>
            <TableCell>
              <span
                class="inline-flex rounded-full px-2 py-1 text-[10px] font-black"
                :class="specification.is_enabled
                  ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                  : 'bg-muted text-muted-foreground'"
              >
                {{ specification.is_enabled ? t('fitmentCatalog.enabled') : t('fitmentCatalog.disabled') }}
              </span>
            </TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground/80">
              {{ formatDate(specification.updated_at || specification.created_at) }}
            </TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="t('fitmentCatalog.hub.manageAria', { name: specification.display_name })">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem v-if="canEdit" @select="openEdit(specification)">
                    <Pencil class="size-4" />
                    {{ t('fitmentCatalog.edit') }}
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="canEdit" @select="toggleStatus(specification)">
                    <Power class="size-4" />
                    {{ specification.is_enabled ? t('fitmentCatalog.disable') : t('fitmentCatalog.enable') }}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="canDelete" />
                  <DropdownMenuItem v-if="canDelete" class="text-destructive focus:text-destructive" @select="removeSpecification(specification)">
                    <Trash2 class="size-4" />
                    {{ t('fitmentCatalog.delete') }}
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
          :page-size="pagination.page_size"
          :total="pagination.total"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="dialogOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ form.id ? t('fitmentCatalog.hub.editTitle') : t('fitmentCatalog.hub.createTitle') }}</DialogTitle>
          <DialogDescription>
            {{ t('fitmentCatalog.hub.dialogDescription') }}
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-5" @submit.prevent="save">
          <div class="grid gap-3 md:grid-cols-2">
            <AdminFormField :label="t('fitmentCatalog.fields.specCode')" required :error="formErrors.spec_code" :description="t('fitmentCatalog.hub.specCodeDescription')">
              <Input v-model="form.spec_code" :disabled="saving" :placeholder="t('fitmentCatalog.placeholders.specCode')" />
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.displayName')" required :error="formErrors.display_name">
              <Input v-model="form.display_name" :disabled="saving" :placeholder="t('fitmentCatalog.placeholders.displayName')" />
            </AdminFormField>
          </div>

          <div class="grid gap-3 md:grid-cols-3">
            <AdminFormField :label="t('fitmentCatalog.fields.position')" required :error="formErrors.position">
              <Select v-model="form.position" :disabled="saving">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="front">{{ t('fitmentCatalog.hubPosition.front') }}</SelectItem>
                  <SelectItem value="rear">{{ t('fitmentCatalog.hubPosition.rear') }}</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.axleType')" required :error="formErrors.axle_type">
              <Select v-model="form.axle_type" :disabled="saving">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="quick_release">{{ t('fitmentCatalog.axleType.quick_release') }}</SelectItem>
                  <SelectItem value="thru_axle">{{ t('fitmentCatalog.axleType.thru_axle') }}</SelectItem>
                  <SelectItem value="bolt_on">{{ t('fitmentCatalog.axleType.bolt_on') }}</SelectItem>
                  <SelectItem value="other">{{ t('fitmentCatalog.axleType.other') }}</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.axleSpacing')" required :error="formErrors.axle_spacing_mm">
              <Input v-model="form.axle_spacing_mm" :disabled="saving" type="number" min="1" step="1" :placeholder="t('fitmentCatalog.placeholders.axleSpacing')" />
            </AdminFormField>
          </div>

          <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
            <AdminFormField :label="t('fitmentCatalog.fields.notes')">
              <Textarea v-model="form.notes" :disabled="saving" class="min-h-24" :placeholder="t('fitmentCatalog.placeholders.notes')" />
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.sortOrder')">
              <Input v-model="form.sort_order" :disabled="saving" type="number" min="0" step="1" class="w-32" />
            </AdminFormField>
          </div>

          <div class="flex items-center justify-between gap-4 rounded-lg border bg-muted/20 px-3 py-3">
            <div>
              <p class="text-xs font-black">{{ t('fitmentCatalog.hub.enabledTitle') }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ t('fitmentCatalog.hub.enabledDescription') }}</p>
            </div>
            <Switch v-model:checked="form.is_enabled" :disabled="saving || !form.id" :aria-label="t('fitmentCatalog.hub.enabledTitle')" />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">{{ t('fitmentCatalog.cancel') }}</Button>
            <Button type="submit" :disabled="saving || (form.id ? !canEdit : !canCreate)">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              {{ t('fitmentCatalog.hub.save') }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Database,
  LoaderCircle,
  MoreHorizontal,
  Pencil,
  Plus,
  Power,
  RotateCcw,
  Search,
  Trash2,
} from '@lucide/vue'
import { type HubSpecification, type HubSpecificationAxleType, type HubSpecificationPayload, type HubSpecificationPosition } from '@/api/fitmentCatalog'
import { fitmentHubSpecificationsApi } from '@/api/fitmentCatalog/hubSpecifications'
import { useAdminI18n } from '@/i18n'
import { buildFitmentCatalogTabs } from '@/lib/fitmentCatalogTabs'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import AdminTabBar from '@/components/admin/AdminTabBar.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'

type StatusFilter = 'all' | 'enabled' | 'disabled'
type PositionFilter = 'all' | HubSpecificationPosition

interface HubSpecificationForm {
  id?: number
  spec_code: string
  display_name: string
  position: HubSpecificationPosition
  axle_type: HubSpecificationAxleType
  axle_spacing_mm: string | number
  notes: string
  is_enabled: boolean
  sort_order: string | number
}

const authStore = useAuthStore()
const route = useRoute()
const { t, locale } = useAdminI18n()
const fitmentTabs = computed(() => buildFitmentCatalogTabs(t))
const canCreate = computed(() => authStore.hasPermission('fitment_catalog:create'))
const canEdit = computed(() => authStore.hasPermission('fitment_catalog:edit'))
const canDelete = computed(() => authStore.hasPermission('fitment_catalog:delete'))

const specifications = ref<HubSpecification[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const filters = reactive<{ search: string; position: PositionFilter; status: StatusFilter }>({
  search: '',
  position: 'all',
  status: 'all',
})
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 1,
})
const form = reactive<HubSpecificationForm>(emptyForm())
const formErrors = reactive<Record<string, string>>({})

function emptyForm(): HubSpecificationForm {
  return {
    spec_code: '',
    display_name: '',
    position: 'rear',
    axle_type: 'thru_axle',
    axle_spacing_mm: 142,
    notes: '',
    is_enabled: false,
    sort_order: 0,
  }
}

const clearFormErrors = () => {
  Object.keys(formErrors).forEach((key) => delete formErrors[key])
}

const assignForm = (specification?: HubSpecification) => {
  Object.assign(form, specification
    ? {
        id: specification.id,
        spec_code: specification.spec_code,
        display_name: specification.display_name,
        position: specification.position,
        axle_type: specification.axle_type,
        axle_spacing_mm: specification.axle_spacing_mm,
        notes: specification.notes,
        is_enabled: specification.is_enabled,
        sort_order: specification.sort_order,
      }
    : { id: undefined, ...emptyForm() })
  clearFormErrors()
}

const listParams = () => ({
  page: pagination.page,
  page_size: pagination.page_size,
  search: filters.search.trim() || undefined,
  position: filters.position === 'all' ? undefined : filters.position,
  is_enabled: filters.status === 'all' ? undefined : filters.status === 'enabled',
})

const loadSpecifications = async () => {
  loading.value = true
  try {
    const result = await fitmentHubSpecificationsApi.listHubSpecifications(listParams())
    specifications.value = result.specifications
    Object.assign(pagination, result.pagination)
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.hub.toast.loadFailed'))
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  void loadSpecifications()
}

const resetFilters = () => {
  filters.search = ''
  filters.position = 'all'
  filters.status = 'all'
  pagination.page = 1
  void loadSpecifications()
}

const updatePage = (page: number) => {
  pagination.page = page
  void loadSpecifications()
}

const updatePageSize = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadSpecifications()
}

const openCreate = () => {
  assignForm()
  dialogOpen.value = true
}

const openEdit = (specification: HubSpecification) => {
  assignForm(specification)
  dialogOpen.value = true
}

const toPayload = (): HubSpecificationPayload => ({
  spec_code: form.spec_code.trim().toUpperCase(),
  display_name: form.display_name.trim(),
  position: form.position,
  axle_type: form.axle_type,
  axle_spacing_mm: Math.floor(Number(form.axle_spacing_mm) || 0),
  notes: form.notes.trim(),
  is_enabled: form.is_enabled,
  sort_order: Math.max(0, Math.floor(Number(form.sort_order) || 0)),
})

const validateForm = (payload: HubSpecificationPayload): boolean => {
  clearFormErrors()
  if (!payload.spec_code) formErrors.spec_code = t('fitmentCatalog.hub.validation.specCode')
  if (!payload.display_name) formErrors.display_name = t('fitmentCatalog.hub.validation.displayName')
  if (!payload.position) formErrors.position = t('fitmentCatalog.hub.validation.position')
  if (!payload.axle_type) formErrors.axle_type = t('fitmentCatalog.hub.validation.axleType')
  if (!Number.isInteger(payload.axle_spacing_mm) || payload.axle_spacing_mm <= 0) {
    formErrors.axle_spacing_mm = t('fitmentCatalog.hub.validation.axleSpacing')
  }
  return Object.keys(formErrors).length === 0
}

const save = async () => {
  const payload = toPayload()
  if (!validateForm(payload)) return

  saving.value = true
  try {
    if (form.id) {
      await fitmentHubSpecificationsApi.updateHubSpecification(form.id, payload)
      toast.success(t('fitmentCatalog.hub.toast.saved'))
    } else {
      await fitmentHubSpecificationsApi.createHubSpecification(payload)
      toast.success(t('fitmentCatalog.hub.toast.created'))
    }
    dialogOpen.value = false
    await loadSpecifications()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.hub.toast.saveFailed'))
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (specification: HubSpecification) => {
  try {
    await fitmentHubSpecificationsApi.updateHubSpecificationStatus(specification.id, !specification.is_enabled)
    toast.success(specification.is_enabled ? t('fitmentCatalog.hub.toast.disabled') : t('fitmentCatalog.hub.toast.enabled'))
    await loadSpecifications()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.hub.toast.statusFailed'))
  }
}

const removeSpecification = async (specification: HubSpecification) => {
  if (!window.confirm(t('fitmentCatalog.hub.confirmDelete', { name: specification.display_name }))) return
  try {
    await fitmentHubSpecificationsApi.removeHubSpecification(specification.id)
    toast.success(t('fitmentCatalog.hub.toast.deleted'))
    if (specifications.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadSpecifications()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.hub.toast.deleteFailed'))
  }
}

const formatPosition = (value: HubSpecificationPosition): string => (
  value === 'front' ? t('fitmentCatalog.hubPosition.front') : t('fitmentCatalog.hubPosition.rear')
)

const formatAxleType = (value: HubSpecificationAxleType): string => {
  if (value === 'quick_release') return t('fitmentCatalog.axleType.quick_release')
  if (value === 'thru_axle') return t('fitmentCatalog.axleType.thru_axle')
  if (value === 'bolt_on') return t('fitmentCatalog.axleType.bolt_on')
  return t('fitmentCatalog.axleType.other')
}

const referenceCount = (specification: HubSpecification): number => (
  specification.frame_reference_count + specification.fork_reference_count
)

const formatDate = (value?: string): string => {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleString(locale.value, { dateStyle: 'medium', timeStyle: 'short' })
}

onMounted(() => {
  void loadSpecifications()
})
</script>
