<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="t('fitmentCatalog.fork.title')"
      :description="t('fitmentCatalog.fork.description')"
    >
      <template #actions>
        <Button v-if="canCreate" @click="openCreate">
          <Plus class="size-4" />
          {{ t('fitmentCatalog.fork.create') }}
        </Button>
      </template>
    </AdminPageHeader>

    <AdminTabBar
      :tabs="fitmentTabs"
      :active-path="route.path"
      :label="t('fitmentCatalog.tabsLabel')"
    />

    <AdminFilterPanel>
      <form class="grid gap-3 md:grid-cols-[minmax(260px,1.5fr)_minmax(150px,0.55fr)_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            {{ t('fitmentCatalog.searchLabel') }}
          </span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" :placeholder="t('fitmentCatalog.fork.searchPlaceholder')" />
          </div>
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
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-56">{{ t('fitmentCatalog.columns.forkBrandModel') }}</TableHead>
            <TableHead class="w-48">{{ t('fitmentCatalog.columns.seriesGeneration') }}</TableHead>
            <TableHead class="w-36">{{ t('fitmentCatalog.columns.year') }}</TableHead>
            <TableHead class="w-28">{{ t('fitmentCatalog.columns.market') }}</TableHead>
            <TableHead class="w-28">{{ t('fitmentCatalog.columns.frontHubSpecs') }}</TableHead>
            <TableHead class="w-24">{{ t('fitmentCatalog.columns.status') }}</TableHead>
            <TableHead class="w-36">{{ t('fitmentCatalog.columns.updatedAt') }}</TableHead>
            <TableHead class="w-20 text-right">{{ t('fitmentCatalog.columns.actions') }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="entries.length === 0" :colspan="8">
            <div class="flex flex-col items-center text-muted-foreground">
              <Waypoints class="mb-2 size-7 opacity-55" />
              <span class="text-xs">{{ filters.search ? t('fitmentCatalog.fork.emptySearch') : t('fitmentCatalog.fork.empty') }}</span>
            </div>
          </TableEmpty>

          <TableRow v-for="entry in entries" :key="entry.id">
            <TableCell>
              <div class="min-w-0">
                <p class="truncate text-xs font-black">{{ entry.brand_name }}</p>
                <p class="mt-1 truncate text-[11px] text-muted-foreground">{{ entry.model_name }}</p>
              </div>
            </TableCell>
            <TableCell>
              <div class="min-w-0">
                <p class="truncate text-xs font-bold">{{ entry.series_name || t('fitmentCatalog.fork.unfilledSeries') }}</p>
                <p v-if="entry.generation_name" class="mt-1 truncate text-[10px] text-muted-foreground">
                  {{ entry.generation_name }}
                </p>
              </div>
            </TableCell>
            <TableCell class="font-mono text-xs font-bold">
              {{ formatYear(entry) }}
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">
              {{ entry.market_code || t('fitmentCatalog.fork.globalMarket') }}
            </TableCell>
            <TableCell class="font-mono text-xs font-bold">
              {{ entry.hub_specification_count }}
            </TableCell>
            <TableCell>
              <span
                class="inline-flex rounded-full px-2 py-1 text-[10px] font-black"
                :class="entry.is_enabled
                  ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                  : 'bg-muted text-muted-foreground'"
              >
                {{ entry.is_enabled ? t('fitmentCatalog.enabled') : t('fitmentCatalog.disabled') }}
              </span>
            </TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground/80">
              {{ formatDate(entry.updated_at || entry.created_at) }}
            </TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="t('fitmentCatalog.fork.manageAria', { name: forkDisplayName(entry) })">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem v-if="canEdit" @select="openEdit(entry)">
                    <Pencil class="size-4" />
                    {{ t('fitmentCatalog.edit') }}
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="canEdit" @select="toggleStatus(entry)">
                    <Power class="size-4" />
                    {{ entry.is_enabled ? t('fitmentCatalog.disable') : t('fitmentCatalog.enable') }}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="canDelete" />
                  <DropdownMenuItem v-if="canDelete" class="text-destructive focus:text-destructive" @select="removeEntry(entry)">
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
          <DialogTitle>{{ form.id ? t('fitmentCatalog.fork.editTitle') : t('fitmentCatalog.fork.createTitle') }}</DialogTitle>
          <DialogDescription>
            {{ t('fitmentCatalog.fork.dialogDescription') }}
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-5" @submit.prevent="save">
          <div class="grid gap-3 md:grid-cols-2">
            <AdminFormField :label="t('fitmentCatalog.fields.forkBrand')" required :error="formErrors.brand_name">
              <Input v-model="form.brand_name" :disabled="saving" :placeholder="t('fitmentCatalog.placeholders.forkBrand')" />
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.forkModel')" required :error="formErrors.model_name">
              <Input v-model="form.model_name" :disabled="saving" :placeholder="t('fitmentCatalog.placeholders.forkModel')" />
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.series')">
              <Input v-model="form.series_name" :disabled="saving" :placeholder="t('fitmentCatalog.placeholders.optional')" />
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.generation')">
              <Input v-model="form.generation_name" :disabled="saving" :placeholder="t('fitmentCatalog.placeholders.optional')" />
            </AdminFormField>
          </div>

          <div class="grid gap-3 md:grid-cols-3">
            <AdminFormField :label="t('fitmentCatalog.fields.yearMode')" required :error="formErrors.year_mode">
              <Select v-model="form.year_mode" :disabled="saving" @update:model-value="handleYearModeChange">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="single">{{ t('fitmentCatalog.yearMode.single') }}</SelectItem>
                  <SelectItem value="range">{{ t('fitmentCatalog.yearMode.range') }}</SelectItem>
                  <SelectItem value="all">{{ t('fitmentCatalog.yearMode.all') }}</SelectItem>
                  <SelectItem value="unknown">{{ t('fitmentCatalog.yearMode.unknown') }}</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>
            <AdminFormField v-if="needsYearFrom" :label="t('fitmentCatalog.fields.yearFrom')" required :error="formErrors.year_from">
              <Input v-model="form.year_from" :disabled="saving" type="number" min="1800" max="2200" step="1" :placeholder="t('fitmentCatalog.placeholders.yearFrom')" />
            </AdminFormField>
            <AdminFormField v-if="needsYearTo" :label="t('fitmentCatalog.fields.yearTo')" required :error="formErrors.year_to">
              <Input v-model="form.year_to" :disabled="saving" type="number" min="1800" max="2200" step="1" :placeholder="t('fitmentCatalog.placeholders.yearTo')" />
            </AdminFormField>
          </div>

          <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
            <AdminFormField :label="t('fitmentCatalog.fields.market')" :description="t('fitmentCatalog.descriptions.market')">
              <Input v-model="form.market_code" :disabled="saving" class="uppercase" :placeholder="t('fitmentCatalog.placeholders.market')" />
            </AdminFormField>
            <AdminFormField :label="t('fitmentCatalog.fields.sortOrder')">
              <Input v-model="form.sort_order" :disabled="saving" type="number" min="0" step="1" class="w-32" />
            </AdminFormField>
          </div>

          <AdminFormField
            :label="t('fitmentCatalog.fields.frontHubSpecs')"
            required
            :error="formErrors.hub_specification_ids"
            :description="t('fitmentCatalog.fork.frontHubDescription')"
          >
            <div v-if="hubOptionsLoading" class="rounded-xl border border-dashed px-3 py-4 text-xs text-muted-foreground">
              {{ t('fitmentCatalog.fork.loadingHubOptions') }}
            </div>
            <div v-else-if="hubOptions.length === 0" class="rounded-xl border border-dashed px-3 py-4 text-xs text-muted-foreground">
              {{ t('fitmentCatalog.fork.noHubOptions') }}
            </div>
            <div v-else class="grid gap-2 md:grid-cols-2">
              <label
                v-for="specification in hubOptions"
                :key="specification.id"
                class="flex min-w-0 items-start gap-3 rounded-xl border border-dashed px-3 py-2.5 transition-colors"
                :class="specification.is_enabled ? 'hover:border-primary/40 hover:bg-primary/5' : 'border-amber-500/40 bg-amber-500/5'"
              >
                <Checkbox
                  class="mt-0.5"
                  :model-value="form.hub_specification_ids.includes(specification.id)"
                  :disabled="saving || (!specification.is_enabled && !form.hub_specification_ids.includes(specification.id))"
                  @update:model-value="toggleHubSpecification(specification.id, $event)"
                />
                <span class="min-w-0">
                  <span class="block truncate text-xs font-black">{{ specification.display_name }}</span>
                  <span class="mt-1 block truncate font-mono text-[10px] text-muted-foreground">
                    {{ specification.spec_code }} · {{ formatAxleType(specification.axle_type) }} · {{ specification.axle_spacing_mm }} mm
                  </span>
                  <span v-if="!specification.is_enabled" class="mt-1 block text-[10px] font-bold text-amber-700 dark:text-amber-300">
                    {{ t('fitmentCatalog.fork.disabledHubNote') }}
                  </span>
                </span>
              </label>
            </div>
          </AdminFormField>

          <AdminFormField :label="t('fitmentCatalog.fields.notes')">
            <Textarea v-model="form.notes" :disabled="saving" class="min-h-24" :placeholder="t('fitmentCatalog.placeholders.notes')" />
          </AdminFormField>

          <div class="flex items-center justify-between gap-4 rounded-lg border bg-muted/20 px-3 py-3">
            <div>
              <p class="text-xs font-black">{{ t('fitmentCatalog.fork.enabledTitle') }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ t('fitmentCatalog.fork.enabledDescription') }}</p>
            </div>
            <Switch v-model:checked="form.is_enabled" :disabled="saving || !form.id" :aria-label="t('fitmentCatalog.fork.enabledTitle')" />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">{{ t('fitmentCatalog.cancel') }}</Button>
            <Button type="submit" :disabled="saving || (form.id ? !canEdit : !canCreate)">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              {{ t('fitmentCatalog.fork.save') }}
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
  LoaderCircle,
  MoreHorizontal,
  Pencil,
  Plus,
  Power,
  RotateCcw,
  Search,
  Trash2,
  Waypoints,
} from '@lucide/vue'
import { type FitmentYearMode, type ForkFitmentEntry, type ForkFitmentEntryPayload, type HubSpecification } from '@/api/fitmentCatalog'
import { forkFitmentEntriesApi } from '@/api/fitmentCatalog/forkEntries'
import { fitmentHubSpecificationsApi } from '@/api/fitmentCatalog/hubSpecifications'
import AdminTabBar from '@/components/admin/AdminTabBar.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
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
import { useAdminI18n } from '@/i18n'
import { buildFitmentCatalogTabs } from '@/lib/fitmentCatalogTabs'

type StatusFilter = 'all' | 'enabled' | 'disabled'

interface ForkForm {
  id?: number
  brand_name: string
  model_name: string
  series_name: string
  generation_name: string
  year_mode: FitmentYearMode
  year_from: string | number | null
  year_to: string | number | null
  market_code: string
  notes: string
  is_enabled: boolean
  sort_order: string | number
  hub_specification_ids: number[]
}

const authStore = useAuthStore()
const route = useRoute()
const { t, locale } = useAdminI18n()
const fitmentTabs = computed(() => buildFitmentCatalogTabs(t))
const canCreate = computed(() => authStore.hasPermission('fitment_catalog:create'))
const canEdit = computed(() => authStore.hasPermission('fitment_catalog:edit'))
const canDelete = computed(() => authStore.hasPermission('fitment_catalog:delete'))

const entries = ref<ForkFitmentEntry[]>([])
const loading = ref(false)
const hubOptionsLoading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const hubOptions = ref<HubSpecification[]>([])
const filters = reactive<{ search: string; status: StatusFilter }>({
  search: '',
  status: 'all',
})
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 1,
})
const form = reactive<ForkForm>(emptyForm())
const formErrors = reactive<Record<string, string>>({})

const needsYearFrom = computed(() => form.year_mode === 'single' || form.year_mode === 'range')
const needsYearTo = computed(() => form.year_mode === 'range')

function emptyForm(): ForkForm {
  return {
    brand_name: '',
    model_name: '',
    series_name: '',
    generation_name: '',
    year_mode: 'unknown',
    year_from: null,
    year_to: null,
    market_code: '',
    notes: '',
    is_enabled: false,
    sort_order: 0,
    hub_specification_ids: [],
  }
}

const clearFormErrors = () => {
  Object.keys(formErrors).forEach((key) => delete formErrors[key])
}

const assignForm = (entry?: ForkFitmentEntry) => {
  Object.assign(form, entry
    ? {
        id: entry.id,
        brand_name: entry.brand_name,
        model_name: entry.model_name,
        series_name: entry.series_name,
        generation_name: entry.generation_name,
        year_mode: entry.year_mode,
        year_from: entry.year_from,
        year_to: entry.year_to,
        market_code: entry.market_code,
        notes: entry.notes,
        is_enabled: entry.is_enabled,
        sort_order: entry.sort_order,
        hub_specification_ids: entry.hub_specifications.map((specification) => specification.id),
      }
    : { id: undefined, ...emptyForm() })
  clearFormErrors()
}

const listParams = () => ({
  page: pagination.page,
  page_size: pagination.page_size,
  search: filters.search.trim() || undefined,
  is_enabled: filters.status === 'all' ? undefined : filters.status === 'enabled',
})

const loadEntries = async () => {
  loading.value = true
  try {
    const result = await forkFitmentEntriesApi.listForkEntries(listParams())
    entries.value = result.entries
    Object.assign(pagination, result.pagination)
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.fork.toast.loadFailed'))
  } finally {
    loading.value = false
  }
}

const mergeHubOptions = (specifications: HubSpecification[]) => {
  const byId = new Map(hubOptions.value.map((specification) => [specification.id, specification]))
  specifications.forEach((specification) => byId.set(specification.id, specification))
  hubOptions.value = Array.from(byId.values()).sort((left, right) => {
    if (left.is_enabled !== right.is_enabled) return left.is_enabled ? -1 : 1
    if (left.sort_order !== right.sort_order) return left.sort_order - right.sort_order
    return left.display_name.localeCompare(right.display_name)
  })
}

const loadHubOptions = async () => {
  hubOptionsLoading.value = true
  try {
    const result = await fitmentHubSpecificationsApi.listHubSpecifications({
      page: 1,
      page_size: 100,
      position: 'front',
      is_enabled: true,
    })
    mergeHubOptions(result.specifications)
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.fork.toast.hubLoadFailed'))
  } finally {
    hubOptionsLoading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  void loadEntries()
}

const resetFilters = () => {
  filters.search = ''
  filters.status = 'all'
  pagination.page = 1
  void loadEntries()
}

const updatePage = (page: number) => {
  pagination.page = page
  void loadEntries()
}

const updatePageSize = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadEntries()
}

const openCreate = () => {
  assignForm()
  dialogOpen.value = true
}

const openEdit = async (entry: ForkFitmentEntry) => {
  try {
    const fullEntry = await forkFitmentEntriesApi.getForkEntry(entry.id)
    mergeHubOptions(fullEntry.hub_specifications)
    assignForm(fullEntry)
    dialogOpen.value = true
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.fork.toast.loadFailed'))
  }
}

const handleYearModeChange = (value: string) => {
  form.year_mode = value as FitmentYearMode
  if (form.year_mode === 'single') form.year_to = null
  if (form.year_mode === 'all' || form.year_mode === 'unknown') {
    form.year_from = null
    form.year_to = null
  }
}

const toNullableInt = (value: string | number | null): number | null => {
  if (value === null || value === '') return null
  const parsed = Number(value)
  return Number.isInteger(parsed) ? parsed : null
}

const toPayload = (): ForkFitmentEntryPayload => ({
  brand_name: form.brand_name.trim(),
  model_name: form.model_name.trim(),
  series_name: form.series_name.trim(),
  generation_name: form.generation_name.trim(),
  year_mode: form.year_mode,
  year_from: needsYearFrom.value ? toNullableInt(form.year_from) : null,
  year_to: needsYearTo.value ? toNullableInt(form.year_to) : null,
  market_code: form.market_code.trim().toUpperCase(),
  notes: form.notes.trim(),
  is_enabled: form.is_enabled,
  sort_order: Math.max(0, Number(form.sort_order) || 0),
  hub_specification_ids: [...form.hub_specification_ids],
})

const validateForm = (payload: ForkFitmentEntryPayload): boolean => {
  clearFormErrors()
  if (!payload.brand_name) formErrors.brand_name = t('fitmentCatalog.fork.validation.brand')
  if (!payload.model_name) formErrors.model_name = t('fitmentCatalog.fork.validation.model')
  if (!payload.year_mode) formErrors.year_mode = t('fitmentCatalog.fork.validation.yearMode')
  if ((payload.year_mode === 'single' || payload.year_mode === 'range') && payload.year_from === null) {
    formErrors.year_from = t('fitmentCatalog.fork.validation.yearFrom')
  }
  if (payload.year_mode === 'range' && payload.year_to === null) {
    formErrors.year_to = t('fitmentCatalog.fork.validation.yearTo')
  }
  if (payload.year_mode === 'range' && payload.year_from !== null && payload.year_to !== null && payload.year_from > payload.year_to) {
    formErrors.year_to = t('fitmentCatalog.fork.validation.yearRange')
  }
  if (payload.is_enabled && payload.hub_specification_ids.length === 0) {
    formErrors.hub_specification_ids = t('fitmentCatalog.fork.validation.hubSpecs')
  }
  return Object.keys(formErrors).length === 0
}

const toggleHubSpecification = (id: number, value: boolean | 'indeterminate') => {
  if (value === 'indeterminate') return
  const selected = new Set(form.hub_specification_ids)
  if (value) selected.add(id)
  else selected.delete(id)
  form.hub_specification_ids = Array.from(selected)
}

const forkDisplayName = (entry: ForkFitmentEntry): string => `${entry.brand_name} / ${entry.model_name}`

const formatAxleType = (value: HubSpecification['axle_type']): string => {
  if (value === 'quick_release') return t('fitmentCatalog.axleType.quick_release')
  if (value === 'thru_axle') return t('fitmentCatalog.axleType.thru_axle')
  if (value === 'bolt_on') return t('fitmentCatalog.axleType.bolt_on')
  return t('fitmentCatalog.axleType.other')
}

const save = async () => {
  const payload = toPayload()
  if (!validateForm(payload)) return

  saving.value = true
  try {
    if (form.id) {
      await forkFitmentEntriesApi.updateForkEntry(form.id, payload)
      toast.success(t('fitmentCatalog.fork.toast.saved'))
    } else {
      await forkFitmentEntriesApi.createForkEntry(payload)
      toast.success(t('fitmentCatalog.fork.toast.created'))
    }
    dialogOpen.value = false
    await loadEntries()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.fork.toast.saveFailed'))
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (entry: ForkFitmentEntry) => {
  try {
    await forkFitmentEntriesApi.updateForkEntryStatus(entry.id, !entry.is_enabled)
    toast.success(entry.is_enabled ? t('fitmentCatalog.fork.toast.disabled') : t('fitmentCatalog.fork.toast.enabled'))
    await loadEntries()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.fork.toast.statusFailed'))
  }
}

const removeEntry = async (entry: ForkFitmentEntry) => {
  if (!window.confirm(t('fitmentCatalog.fork.confirmDelete', { name: forkDisplayName(entry) }))) return
  try {
    await forkFitmentEntriesApi.removeForkEntry(entry.id)
    toast.success(t('fitmentCatalog.fork.toast.deleted'))
    if (entries.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadEntries()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : t('fitmentCatalog.fork.toast.deleteFailed'))
  }
}

const formatYear = (entry: ForkFitmentEntry): string => {
  if (entry.year_mode === 'single') return String(entry.year_from || '—')
  if (entry.year_mode === 'range') return `${entry.year_from || '—'} - ${entry.year_to || '—'}`
  if (entry.year_mode === 'all') return t('fitmentCatalog.yearMode.all')
  return t('fitmentCatalog.yearMode.unknown')
}

const formatDate = (value?: string): string => {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleString(locale.value, { dateStyle: 'medium', timeStyle: 'short' })
}

onMounted(() => {
  void loadEntries()
  void loadHubOptions()
})
</script>
