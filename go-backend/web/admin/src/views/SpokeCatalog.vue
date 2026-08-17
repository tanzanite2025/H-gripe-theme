<template>
  <div class="space-y-4">
    <AdminPageHeader title="辐条计算器数据" description="维护计算器和 Smart Search 共用的轮圈、花鼓与推荐装配基准库">
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="loadCatalog">
          <RefreshCw class="size-4" />
          刷新
        </Button>
        <Button variant="outline" @click="downloadJson">
          <Download class="size-4" />
          导出
        </Button>
        <Button :disabled="saving || loading" @click="saveCatalog">
          <Save class="size-4" />
          保存
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <Tabs v-model="activeTab" class="space-y-4">
      <TabsList>
        <TabsTrigger value="rims">Rims</TabsTrigger>
        <TabsTrigger value="hubs">Hubs</TabsTrigger>
        <TabsTrigger value="builds">Builds</TabsTrigger>
        <TabsTrigger value="json">Import</TabsTrigger>
      </TabsList>

      <TabsContent value="rims">
        <AdminTablePanel :loading="loading">
          <template #header>
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-sm font-black uppercase tracking-tight">轮圈数据库</h2>
                <p class="text-[11px] font-bold text-muted-foreground">ERD 允许为空，前台会保留型号但不自动计算。</p>
              </div>
              <Button size="sm" variant="outline" @click="addRimBrand">
                <Plus class="size-3.5" />
                品牌
              </Button>
            </div>
          </template>

          <table class="w-full min-w-[860px] text-xs">
            <thead class="border-b border-dashed border-border/70 text-[10px] uppercase tracking-widest text-muted-foreground">
              <tr>
                <th class="px-3 py-2 text-left">Code</th>
                <th class="px-3 py-2 text-left">Name</th>
                <th class="px-3 py-2 text-left">ERD mm</th>
                <th class="px-3 py-2 text-left">Weight g</th>
                <th class="w-28 px-3 py-2 text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="(brand, brandIndex) in catalog.rims" :key="brand.id || brandIndex">
                <tr class="border-b border-dashed border-border/60 bg-muted/20">
                  <td class="px-3 py-2">
                    <input v-model="brand.id" :class="inputClass" placeholder="dt_swiss">
                  </td>
                  <td class="px-3 py-2">
                    <input v-model="brand.name" :class="inputClass" placeholder="DT Swiss">
                  </td>
                  <td class="px-3 py-2 text-muted-foreground" colspan="2">{{ brand.items.length }} models</td>
                  <td class="px-3 py-2">
                    <div class="flex justify-end gap-1">
                      <Button size="icon-xs" variant="ghost" @click="addRimModel(brand)">
                        <Plus class="size-3" />
                      </Button>
                      <Button size="icon-xs" variant="destructive" @click="removeItem(catalog.rims, brandIndex)">
                        <Trash2 class="size-3" />
                      </Button>
                    </div>
                  </td>
                </tr>
                <tr v-for="(model, modelIndex) in brand.items" :key="`${brand.id}-${model.id}-${modelIndex}`" class="border-b border-dashed border-border/40">
                  <td class="px-3 py-2 pl-8">
                    <input v-model="model.id" :class="inputClass" placeholder="rr411_db">
                  </td>
                  <td class="px-3 py-2">
                    <input v-model="model.name" :class="inputClass" placeholder="RR 411 db">
                  </td>
                  <td class="px-3 py-2">
                    <input :value="numberInputValue(model.erd)" :class="inputClass" inputmode="decimal" placeholder="598" @input="model.erd = nullableNumberFromEvent($event)">
                  </td>
                  <td class="px-3 py-2">
                    <input :value="numberInputValue(model.weight ?? null)" :class="inputClass" inputmode="decimal" placeholder="-"
                      @input="model.weight = nullableNumberFromEvent($event)">
                  </td>
                  <td class="px-3 py-2 text-right">
                    <Button size="icon-xs" variant="destructive" @click="removeItem(brand.items, modelIndex)">
                      <Trash2 class="size-3" />
                    </Button>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="hubs">
        <AdminTablePanel :loading="loading">
          <template #header>
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-sm font-black uppercase tracking-tight">花鼓几何</h2>
                <p class="text-[11px] font-bold text-muted-foreground">前/后几何可部分为空，完整四项才会参与自动计算。</p>
              </div>
              <Button size="sm" variant="outline" @click="addHubBrand">
                <Plus class="size-3.5" />
                品牌
              </Button>
            </div>
          </template>

          <table class="w-full min-w-[1360px] text-xs">
            <thead class="border-b border-dashed border-border/70 text-[10px] uppercase tracking-widest text-muted-foreground">
              <tr>
                <th class="px-3 py-2 text-left">Code</th>
                <th class="px-3 py-2 text-left">Name</th>
                <th class="px-3 py-2 text-left">F L flange</th>
                <th class="px-3 py-2 text-left">F R flange</th>
                <th class="px-3 py-2 text-left">F L PCD</th>
                <th class="px-3 py-2 text-left">F R PCD</th>
                <th class="px-3 py-2 text-left">R L flange</th>
                <th class="px-3 py-2 text-left">R R flange</th>
                <th class="px-3 py-2 text-left">R L PCD</th>
                <th class="px-3 py-2 text-left">R R PCD</th>
                <th class="w-28 px-3 py-2 text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="(brand, brandIndex) in catalog.hubs" :key="brand.id || brandIndex">
                <tr class="border-b border-dashed border-border/60 bg-muted/20">
                  <td class="px-3 py-2">
                    <input v-model="brand.id" :class="inputClass" placeholder="dt_swiss">
                  </td>
                  <td class="px-3 py-2">
                    <input v-model="brand.name" :class="inputClass" placeholder="DT Swiss">
                  </td>
                  <td class="px-3 py-2 text-muted-foreground" colspan="8">{{ brand.items.length }} models</td>
                  <td class="px-3 py-2">
                    <div class="flex justify-end gap-1">
                      <Button size="icon-xs" variant="ghost" @click="addHubModel(brand)">
                        <Plus class="size-3" />
                      </Button>
                      <Button size="icon-xs" variant="destructive" @click="removeItem(catalog.hubs, brandIndex)">
                        <Trash2 class="size-3" />
                      </Button>
                    </div>
                  </td>
                </tr>
                <tr v-for="(model, modelIndex) in brand.items" :key="`${brand.id}-${model.id}-${modelIndex}`" class="border-b border-dashed border-border/40">
                  <td class="px-3 py-2 pl-8">
                    <input v-model="model.id" :class="inputClass" placeholder="350_road_db_cl">
                  </td>
                  <td class="px-3 py-2">
                    <input v-model="model.name" :class="inputClass" placeholder="350 Road db CL">
                  </td>
                  <td v-for="field in hubNumberFields" :key="field.key" class="px-3 py-2">
                    <input
                      :value="numberInputValue(geometryValue(model, field.side, field.prop))"
                      :class="inputClass"
                      inputmode="decimal"
                      placeholder="-"
                      @input="setGeometryValue(model, field.side, field.prop, nullableNumberFromEvent($event))"
                    >
                  </td>
                  <td class="px-3 py-2 text-right">
                    <Button size="icon-xs" variant="destructive" @click="removeItem(brand.items, modelIndex)">
                      <Trash2 class="size-3" />
                    </Button>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="builds">
        <AdminTablePanel :loading="loading">
          <template #header>
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-sm font-black uppercase tracking-tight">Smart Search 装配预设</h2>
                <p class="text-[11px] font-bold text-muted-foreground">列表只展示关键配置；完整预设统一在弹窗中添加和编辑。</p>
              </div>
              <Button size="sm" variant="outline" @click="addPreset">
                <Plus class="size-3.5" />
                预设
              </Button>
            </div>
          </template>

          <div class="space-y-4">
            <div class="grid gap-3 lg:grid-cols-[minmax(260px,1fr)_auto_auto]">
              <div class="relative">
                <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  v-model="presetSearch"
                  :class="[inputClass, 'h-9 pl-9']"
                  placeholder="搜索 code、名称、花鼓、轮圈、关键词..."
                >
              </div>
              <div class="flex items-center gap-2 text-[11px] font-bold text-muted-foreground">
                <span>每页</span>
                <select v-model.number="presetPageSize" :class="[selectClass, 'w-20']">
                  <option :value="25">25</option>
                  <option :value="50">50</option>
                  <option :value="100">100</option>
                </select>
                <span>条</span>
              </div>
              <div class="flex items-center justify-end text-[11px] font-bold text-muted-foreground">
                {{ presetRangeLabel }}
              </div>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full min-w-[980px] text-xs">
                <thead class="border-b border-dashed border-border/70 text-[10px] uppercase tracking-widest text-muted-foreground">
                  <tr>
                    <th class="px-3 py-2 text-left">Preset</th>
                    <th class="px-3 py-2 text-left">Hub</th>
                    <th class="px-3 py-2 text-left">Rim</th>
                    <th class="px-3 py-2 text-left">Spokes</th>
                    <th class="px-3 py-2 text-left">Cross</th>
                    <th class="px-3 py-2 text-left">Actual length</th>
                    <th class="w-24 px-3 py-2 text-right">Action</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="paginatedPresets.length === 0">
                    <td colspan="7" class="px-4 py-10 text-center text-xs font-bold text-muted-foreground">
                      {{ presetSearch ? '没有匹配的装配预设' : '暂无装配预设' }}
                    </td>
                  </tr>
                  <tr
                    v-for="{ preset, index } in paginatedPresets"
                    :key="preset.id || index"
                    class="border-b border-dashed border-border/60 align-top transition-colors hover:bg-muted/20"
                  >
                    <td class="px-3 py-3">
 <div class="max-w-[240px] truncate font-black" :title="preset.name">{{ preset.name || '未命名预设'}}</div>
 <div class="mt-1 max-w-[240px] truncate font-mono text-[10px] text-muted-foreground" :title="preset.id">{{ preset.id || '未设置 code'}}</div>
                    </td>
                    <td class="px-3 py-3">
                      <div class="max-w-[220px] truncate font-bold" :title="presetHubLabel(preset)">{{ presetHubLabel(preset) }}</div>
                    </td>
                    <td class="px-3 py-3">
                      <div class="max-w-[220px] truncate font-bold" :title="presetRimLabel(preset)">{{ presetRimLabel(preset) }}</div>
                    </td>
 <td class="px-3 py-3 font-black">{{ preset.spokeCount || '—'}}</td>
                    <td class="px-3 py-3">
                      <div class="font-black">{{ preset.crossing }}</div>
                      <div class="mt-1 text-[10px] text-muted-foreground">{{ optionLabel(catalog.options.crossings, preset.crossing) }}</div>
                    </td>
                    <td class="px-3 py-3 font-mono text-[10px] leading-relaxed">
                      <div>F {{ lengthPair(preset.actualLengths?.frontLeft, preset.actualLengths?.frontRight) }}</div>
                      <div>R {{ lengthPair(preset.actualLengths?.rearLeft, preset.actualLengths?.rearRight) }}</div>
                    </td>
                    <td class="px-3 py-3">
                      <div class="flex justify-end gap-1">
                        <Button size="icon-xs" variant="ghost" title="编辑预设" @click="editPreset(index)">
                          <Pencil class="size-3.5" />
                        </Button>
                        <Button size="icon-xs" variant="destructive" title="删除预设" @click="removeItem(catalog.presets, index)">
                          <Trash2 class="size-3.5" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="filteredPresets.length > 0" class="flex flex-wrap items-center justify-between gap-2 border-t border-dashed border-border/60 pt-3">
              <span class="text-[11px] font-bold text-muted-foreground">第 {{ presetPage }} / {{ presetPageCount }} 页</span>
              <div class="flex items-center gap-1">
                <Button size="icon-xs" variant="outline" :disabled="presetPage <= 1" title="上一页" @click="presetPage -= 1">
                  <ChevronLeft class="size-3.5" />
                </Button>
                <Button size="icon-xs" variant="outline" :disabled="presetPage >= presetPageCount" title="下一页" @click="presetPage += 1">
                  <ChevronRight class="size-3.5" />
                </Button>
              </div>
            </div>
          </div>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="json">
        <AdminTablePanel :loading="loading">
          <template #header>
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-sm font-black uppercase tracking-tight">模板导入 / JSON 高级</h2>
                <p class="text-[11px] font-bold text-muted-foreground">优先使用系统模板；JSON 仅保留给维护人员做完整 catalog 导入。</p>
              </div>
              <div class="flex flex-wrap gap-2">
                <Button size="sm" variant="outline" :disabled="saving" @click="downloadPresetTemplate">
                  <Download class="size-3.5" />
                  下载预设模板
                </Button>
                <Button size="sm" variant="outline" :disabled="saving" @click="triggerPresetTemplateFileInput">
                  <Upload class="size-3.5" />
                  导入预设模板
                </Button>
                <Button size="sm" variant="outline" @click="refreshJsonText">
                  <RefreshCw class="size-3.5" />
                  同步 JSON
                </Button>
                <Button size="sm" variant="outline" @click="triggerFileInput">
                  <Upload class="size-3.5" />
                  上传 JSON
                </Button>
                <Button size="sm" :disabled="saving" @click="saveJsonText">
                  <Save class="size-3.5" />
                  保存 JSON
                </Button>
              </div>
            </div>
          </template>

          <input ref="presetTemplateFileInputRef" class="hidden" type="file" accept=".xlsx,.xlsm,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" @change="handlePresetTemplateFileChange">
          <input ref="fileInputRef" class="hidden" type="file" accept="application/json,.json" @change="handleFileChange">
          <div class="border-b border-dashed border-border/70 p-4">
            <div class="grid gap-3 lg:grid-cols-3">
              <div>
                <div class="text-xs font-black">标准预设模板</div>
                <p class="mt-1 text-[11px] font-bold leading-relaxed text-muted-foreground">
                  表头已锁定；轮圈、花鼓、轮位、辐条数、交叉数和辐条帽类型只能从系统字典下拉选择。
                </p>
              </div>
              <div>
                <div class="text-xs font-black">导入范围</div>
                <p class="mt-1 text-[11px] font-bold leading-relaxed text-muted-foreground">
                  只新增或更新装配预设，不覆盖轮圈、花鼓和计算器选项基础数据。
                </p>
              </div>
              <div>
                <div class="text-xs font-black">清洗规则</div>
                <p class="mt-1 text-[11px] font-bold leading-relaxed text-muted-foreground">
                  服务端会清理首尾空格、统一 ID 大小写、去重关键词，并再次校验系统关联关系。
                </p>
              </div>
            </div>
          </div>
          <div class="p-4">
            <textarea
              v-model="jsonText"
              class="min-h-[460px] w-full rounded-2xl border border-dashed border-border/80 bg-muted/40 p-4 font-mono text-[11px] leading-relaxed outline-none focus:ring-2 focus:ring-ring/40"
              spellcheck="false"
            />
          </div>
        </AdminTablePanel>
      </TabsContent>
    </Tabs>

    <SpokeBuildPresetDialog
      v-model:open="presetDialogOpen"
      :mode="presetDialogMode"
      :form="presetDialogForm"
      :rims="catalog.rims"
      :hubs="catalog.hubs"
      :options="catalog.options"
      @submit="savePresetFromDialog"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Calculator, ChevronLeft, ChevronRight, Database, Download, Pencil, Plus, RefreshCw, Save, Search, Trash2, Upload } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import SpokeBuildPresetDialog from '@/components/admin/spoke/SpokeBuildPresetDialog.vue'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import spokeCatalogApi, {
  type SpokeBrand,
  type SpokeBuildActualLengths,
  type SpokeBuildPreset,
  type SpokeCatalog,
  type SpokeCatalogOptions,
  type SpokeHubGeometry,
  type SpokeHubModel,
  type SpokeRimModel,
} from '@/api/spokeCatalog'

type HubSide = 'front' | 'rear'
type HubGeometryProp = 'leftFlange' | 'rightFlange' | 'leftFlangePcd' | 'rightFlangePcd'

const defaultOptions: SpokeCatalogOptions = {
  spokeCounts: [16, 18, 20, 24, 28, 32, 36].map((value) => ({ value, label: String(value) })),
  crossings: [
    { value: 0, label: '0-cross (Radial)' },
    { value: 1, label: '1-cross' },
    { value: 2, label: '2-cross' },
    { value: 3, label: '3-cross' },
    { value: 4, label: '4-cross' },
  ],
  nippleTypes: [
    { value: 'standard', label: 'Standard external' },
    { value: 'hidden', label: 'Hidden / aero' },
  ],
  wheelPositions: [
    { value: 'auto', label: 'Auto' },
    { value: 'front', label: 'Front' },
    { value: 'rear', label: 'Rear' },
  ],
}

const createEmptyCatalog = (): SpokeCatalog => ({
  options: defaultOptions,
  rims: [],
  hubs: [],
  presets: [],
})

const catalog = ref<SpokeCatalog>(createEmptyCatalog())
const activeTab = ref('rims')
const loading = ref(false)
const saving = ref(false)
const jsonText = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const presetTemplateFileInputRef = ref<HTMLInputElement | null>(null)
const presetSearch = ref('')
const presetPage = ref(1)
const presetPageSize = ref(25)
const presetDialogOpen = ref(false)
const presetDialogMode = ref<'create' | 'edit'>('create')
const presetDialogIndex = ref(-1)

const inputClass = 'h-8 w-full min-w-0 rounded-xl border-none bg-muted/50 px-2.5 text-xs font-bold outline-none transition focus:ring-2 focus:ring-ring/40'
const selectClass = 'h-8 w-full min-w-0 rounded-xl border-none bg-muted/50 px-2.5 text-xs font-bold outline-none transition focus:ring-2 focus:ring-ring/40'

const hubNumberFields: Array<{ key: string; side: HubSide; prop: HubGeometryProp }> = [
  { key: 'front-left-flange', side: 'front', prop: 'leftFlange' },
  { key: 'front-right-flange', side: 'front', prop: 'rightFlange' },
  { key: 'front-left-pcd', side: 'front', prop: 'leftFlangePcd' },
  { key: 'front-right-pcd', side: 'front', prop: 'rightFlangePcd' },
  { key: 'rear-left-flange', side: 'rear', prop: 'leftFlange' },
  { key: 'rear-right-flange', side: 'rear', prop: 'rightFlange' },
  { key: 'rear-left-pcd', side: 'rear', prop: 'leftFlangePcd' },
  { key: 'rear-right-pcd', side: 'rear', prop: 'rightFlangePcd' },
]

const statItems = computed(() => [
  { key: 'rim-brands', label: 'Rim brands', value: catalog.value.rims.length, icon: Database, tone: 'blue' },
  { key: 'rim-models', label: 'Rim models', value: catalog.value.rims.reduce((total, brand) => total + brand.items.length, 0), icon: Calculator, tone: 'green' },
  { key: 'hub-models', label: 'Hub models', value: catalog.value.hubs.reduce((total, brand) => total + brand.items.length, 0), icon: Database, tone: 'amber' },
  { key: 'presets', label: 'Search builds', value: catalog.value.presets.length, icon: Search, tone: 'gray' },
])

const emptyGeometry = (): SpokeHubGeometry => ({
  leftFlange: null,
  rightFlange: null,
  leftFlangePcd: null,
  rightFlangePcd: null,
})

const emptyActualLengths = (): SpokeBuildActualLengths => ({
  frontLeft: null,
  frontRight: null,
  rearLeft: null,
  rearRight: null,
  notes: '',
})

const clonePreset = (preset: SpokeBuildPreset): SpokeBuildPreset => ({
  ...preset,
  description: preset.description || '',
  keywords: Array.isArray(preset.keywords) ? [...preset.keywords] : [],
  wheelPosition: preset.wheelPosition || 'auto',
  actualLengths: {
    ...emptyActualLengths(),
    ...(preset.actualLengths || {}),
    notes: preset.actualLengths?.notes || '',
  },
})

const createEmptyPreset = (): SpokeBuildPreset => {
  const rimBrand = catalog.value.rims[0]
  const hubBrand = catalog.value.hubs[0]
  return {
    id: `build_${catalog.value.presets.length + 1}`,
    name: 'New verified build',
    description: '',
    keywords: [],
    rimBrandId: rimBrand?.id || '',
    rimModelId: rimBrand?.items[0]?.id || '',
    hubBrandId: hubBrand?.id || '',
    hubModelId: hubBrand?.items[0]?.id || '',
    wheelPosition: catalog.value.options.wheelPositions[0]?.value || 'auto',
    spokeCount: catalog.value.options.spokeCounts[0]?.value || 24,
    crossing: catalog.value.options.crossings[0]?.value || 0,
    nippleType: catalog.value.options.nippleTypes[0]?.value || 'standard',
    nippleLength: null,
    actualLengths: emptyActualLengths(),
  }
}

const presetDialogForm = ref<SpokeBuildPreset>(createEmptyPreset())

const normalizeLoadedCatalog = (payload: SpokeCatalog): SpokeCatalog => ({
  options: payload.options || defaultOptions,
  rims: Array.isArray(payload.rims) ? payload.rims : [],
  hubs: Array.isArray(payload.hubs) ? payload.hubs : [],
  presets: Array.isArray(payload.presets) ? payload.presets.map(clonePreset) : [],
})

const loadCatalog = async () => {
  loading.value = true
  try {
    catalog.value = normalizeLoadedCatalog(await spokeCatalogApi.get())
    refreshJsonText()
  } finally {
    loading.value = false
  }
}

const saveCatalog = async () => {
  saving.value = true
  try {
    catalog.value = normalizeLoadedCatalog(await spokeCatalogApi.save(catalog.value))
    refreshJsonText()
    toast.success('辐条计算器数据已保存')
  } finally {
    saving.value = false
  }
}

const saveJsonText = async () => {
  let parsed: SpokeCatalog
  try {
    parsed = JSON.parse(jsonText.value)
  } catch (error: any) {
    toast.error(error?.message || 'JSON 格式错误')
    return
  }

  saving.value = true
  try {
    catalog.value = normalizeLoadedCatalog(await spokeCatalogApi.save(parsed))
    refreshJsonText()
    toast.success('JSON 已保存')
  } finally {
    saving.value = false
  }
}

const refreshJsonText = () => {
  jsonText.value = JSON.stringify(catalog.value, null, 2)
}

const downloadJson = () => {
  const blob = new Blob([JSON.stringify(catalog.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'spoke-calculator-catalog.json'
  link.click()
  URL.revokeObjectURL(url)
}

const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const triggerPresetTemplateFileInput = () => {
  presetTemplateFileInputRef.value?.click()
}

const downloadPresetTemplate = async () => {
  saving.value = true
  try {
    const blob = await spokeCatalogApi.downloadPresetTemplate()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'spoke-preset-template.xlsx'
    link.click()
    URL.revokeObjectURL(url)
    toast.success('预设模板已下载')
  } finally {
    saving.value = false
  }
}

const handlePresetTemplateFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  saving.value = true
  try {
    catalog.value = normalizeLoadedCatalog(await spokeCatalogApi.importPresetTemplate(file))
    refreshJsonText()
    toast.success('预设模板已导入')
  } finally {
    saving.value = false
  }
}

const handleFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  saving.value = true
  try {
    catalog.value = normalizeLoadedCatalog(await spokeCatalogApi.importFile(file))
    refreshJsonText()
    toast.success('JSON 文件已导入')
  } finally {
    saving.value = false
  }
}

const numberInputValue = (value: number | null | undefined) => value == null ? '' : String(value)
const nullableNumberFromEvent = (event: Event): number | null => {
  const value = (event.target as HTMLInputElement).value.trim()
  if (!value) return null
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : null
}

const removeItem = <T>(items: T[], index: number) => {
  items.splice(index, 1)
}

const addRimBrand = () => {
  catalog.value.rims.push({ id: `rim_brand_${catalog.value.rims.length + 1}`, name: 'New rim brand', items: [] })
}

const addRimModel = (brand: SpokeBrand<SpokeRimModel>) => {
  brand.items.push({ id: `rim_model_${brand.items.length + 1}`, name: 'New rim model', erd: null, weight: null })
}

const addHubBrand = () => {
  catalog.value.hubs.push({ id: `hub_brand_${catalog.value.hubs.length + 1}`, name: 'New hub brand', items: [] })
}

const addHubModel = (brand: SpokeBrand<SpokeHubModel>) => {
  brand.items.push({ id: `hub_model_${brand.items.length + 1}`, name: 'New hub model', front: emptyGeometry(), rear: emptyGeometry() })
}

const ensureGeometry = (model: SpokeHubModel, side: HubSide): SpokeHubGeometry => {
  if (!model[side]) model[side] = emptyGeometry()
  return model[side]!
}

const geometryValue = (model: SpokeHubModel, side: HubSide, prop: HubGeometryProp): number | null => (
  ensureGeometry(model, side)[prop]
)

const setGeometryValue = (model: SpokeHubModel, side: HubSide, prop: HubGeometryProp, value: number | null) => {
  ensureGeometry(model, side)[prop] = value
}

const addPreset = () => {
  presetDialogMode.value = 'create'
  presetDialogIndex.value = -1
  presetDialogForm.value = createEmptyPreset()
  presetDialogOpen.value = true
}

const editPreset = (index: number) => {
  const preset = catalog.value.presets[index]
  if (!preset) return
  presetDialogMode.value = 'edit'
  presetDialogIndex.value = index
  presetDialogForm.value = clonePreset(preset)
  presetDialogOpen.value = true
}

const savePresetFromDialog = () => {
  const nextPreset = clonePreset(presetDialogForm.value)
  if (!nextPreset.id.trim()) {
    toast.error('Code 不能为空')
    return
  }
  if (!nextPreset.name.trim()) {
    toast.error('名称不能为空')
    return
  }
  const duplicateIndex = catalog.value.presets.findIndex((preset, index) => (
    index !== presetDialogIndex.value && preset.id.trim() === nextPreset.id.trim()
  ))
  if (duplicateIndex >= 0) {
    toast.error('Code 已存在，请换一个唯一标识')
    return
  }

  if (presetDialogIndex.value >= 0) {
    catalog.value.presets.splice(presetDialogIndex.value, 1, nextPreset)
    toast.success('预设已更新，点击顶部保存后提交')
  } else {
    catalog.value.presets.push(nextPreset)
    toast.success('预设已加入列表，点击顶部保存后提交')
  }
  presetDialogOpen.value = false
}

const brandName = <T extends { id: string; name: string }>(brands: SpokeBrand<T>[], id: string) => (
  brands.find((brand) => brand.id === id)?.name || id || '未选择'
)

const modelName = <T extends { id: string; name: string }>(brands: SpokeBrand<T>[], brandId: string, modelId: string) => (
  brands.find((brand) => brand.id === brandId)?.items.find((model) => model.id === modelId)?.name || modelId || '未选择'
)

const presetHubLabel = (preset: SpokeBuildPreset) => (
  `${brandName(catalog.value.hubs, preset.hubBrandId)} / ${modelName(catalog.value.hubs, preset.hubBrandId, preset.hubModelId)}`
)

const presetRimLabel = (preset: SpokeBuildPreset) => (
  `${brandName(catalog.value.rims, preset.rimBrandId)} / ${modelName(catalog.value.rims, preset.rimBrandId, preset.rimModelId)}`
)

const optionLabel = (options: Array<{ value: number; label: string }>, value: number) => (
  options.find((option) => option.value === value)?.label || String(value)
)

const lengthValue = (value: number | null | undefined) => value == null ? '—' : String(value)
const lengthPair = (left: number | null | undefined, right: number | null | undefined) => `${lengthValue(left)} / ${lengthValue(right)}`

const filteredPresets = computed(() => {
  const query = presetSearch.value.trim().toLowerCase()
  return catalog.value.presets
    .map((preset, index) => ({ preset, index }))
    .filter(({ preset }) => {
      if (!query) return true
      const haystack = [
        preset.id,
        preset.name,
        preset.description,
        preset.keywords.join(' '),
        presetHubLabel(preset),
        presetRimLabel(preset),
        String(preset.spokeCount),
        String(preset.crossing),
        optionLabel(catalog.value.options.crossings, preset.crossing),
      ].join(' ').toLowerCase()
      return haystack.includes(query)
    })
})

const presetPageCount = computed(() => Math.max(1, Math.ceil(filteredPresets.value.length / presetPageSize.value)))
const paginatedPresets = computed(() => {
  const start = (presetPage.value - 1) * presetPageSize.value
  return filteredPresets.value.slice(start, start + presetPageSize.value)
})
const presetRangeLabel = computed(() => {
  if (filteredPresets.value.length === 0) return '0 条'
  const start = (presetPage.value - 1) * presetPageSize.value + 1
  const end = Math.min(presetPage.value * presetPageSize.value, filteredPresets.value.length)
  return `${start}-${end} / ${filteredPresets.value.length} 条`
})

watch([presetSearch, presetPageSize], () => {
  presetPage.value = 1
})

watch(presetPageCount, (pageCount) => {
  if (presetPage.value > pageCount) presetPage.value = pageCount
})

onMounted(() => {
  void loadCatalog()
})
</script>
