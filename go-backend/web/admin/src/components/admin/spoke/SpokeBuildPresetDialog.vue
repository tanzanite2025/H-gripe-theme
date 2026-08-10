<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="max-h-[calc(100dvh-1rem)] overflow-y-auto p-4 sm:p-5" @open-auto-focus.prevent>
      <form class="space-y-3" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '添加装配预设' : '编辑装配预设' }}</DialogTitle>
          <DialogDescription class="text-[11px]">完整配置只在这里维护，列表页只保留便于检索和核对的关键结果。</DialogDescription>
        </DialogHeader>

        <div class="grid gap-2 sm:grid-cols-2">
          <AdminFormField label="Code" required>
            <Input v-model="form.id" class="h-8 font-mono" placeholder="tz_ar45_dt350_fr" />
          </AdminFormField>
          <AdminFormField label="名称" required>
            <Input v-model="form.name" class="h-8" placeholder="DT Swiss 350 / AR45 / 28H / 2-cross" />
          </AdminFormField>
          <AdminFormField label="描述" class="sm:col-span-2">
            <Input v-model="descriptionModel" class="h-8" placeholder="用于搜索结果补充说明，可为空" />
          </AdminFormField>
        </div>

        <section class="space-y-3 border-t border-dashed border-border/70 pt-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <h3 class="text-sm font-black tracking-tight">装配选择</h3>
            <p class="text-[10px] font-bold text-muted-foreground">选项与前端计算器字典保持一致</p>
          </div>

          <div class="grid gap-2 lg:grid-cols-4">
            <AdminFormField label="轮圈品牌 / 型号" required class="lg:col-span-2">
              <div class="grid gap-2 sm:grid-cols-2">
                <Select :model-value="form.rimBrandId" @update:model-value="handleRimBrandChange">
                  <SelectTrigger size="sm" class="w-full min-w-0"><SelectValue placeholder="选择轮圈品牌" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="brand in rims" :key="brand.id" :value="brand.id">
                      {{ brand.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <Select v-model="form.rimModelId">
                  <SelectTrigger size="sm" class="w-full min-w-0"><SelectValue placeholder="选择轮圈型号" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="model in rimModelsFor(form.rimBrandId)" :key="model.id" :value="model.id">
                      {{ model.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </AdminFormField>

            <AdminFormField label="花鼓品牌 / 型号" required class="lg:col-span-2">
              <div class="grid gap-2 sm:grid-cols-2">
                <Select :model-value="form.hubBrandId" @update:model-value="handleHubBrandChange">
                  <SelectTrigger size="sm" class="w-full min-w-0"><SelectValue placeholder="选择花鼓品牌" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="brand in hubs" :key="brand.id" :value="brand.id">
                      {{ brand.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <Select v-model="form.hubModelId">
                  <SelectTrigger size="sm" class="w-full min-w-0"><SelectValue placeholder="选择花鼓型号" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="model in hubModelsFor(form.hubBrandId)" :key="model.id" :value="model.id">
                      {{ model.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </AdminFormField>

            <AdminFormField label="辐条数" required>
              <select v-model.number="form.spokeCount" :class="selectClass">
                <option v-for="option in options.spokeCounts" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </AdminFormField>

            <AdminFormField label="交叉数" required>
              <select v-model.number="form.crossing" :class="selectClass">
                <option v-for="option in options.crossings" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </AdminFormField>

            <AdminFormField label="辐条帽类型" required>
              <select v-model="form.nippleType" :class="selectClass">
                <option v-for="option in options.nippleTypes" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </AdminFormField>

            <AdminFormField label="辐条帽长度">
              <Input
                :model-value="numberInputValue(form.nippleLength)"
                type="number"
                inputmode="decimal"
                placeholder="可为空"
                @update:model-value="form.nippleLength = nullableNumber($event)"
              />
            </AdminFormField>

            <AdminFormField label="轮位" required>
              <select v-model="form.wheelPosition" :class="selectClass">
                <option v-for="option in options.wheelPositions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </AdminFormField>
          </div>
        </section>

        <section class="space-y-2 border-t border-dashed border-border/70 pt-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <h3 class="text-sm font-black tracking-tight">实际编制长度结果</h3>
            <p class="text-[10px] font-bold text-muted-foreground">缺失项允许为空，Smart Search 不会自行计算</p>
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">mm</span>
          </div>

          <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
            <AdminFormField v-for="field in actualLengthFields" :key="field.key" :label="field.label">
              <Input
                :model-value="numberInputValue(actualLengths[field.key])"
                class="h-8"
                type="number"
                inputmode="decimal"
                placeholder="可为空"
                @update:model-value="setActualLength(field.key, $event)"
              />
            </AdminFormField>
          </div>

          <AdminFormField label="结果备注" class="xl:col-span-1">
            <Input v-model="actualLengths.notes" class="h-8" placeholder="实测确认或来源说明" />
          </AdminFormField>
        </section>

        <AdminFormField label="搜索关键词" description="每行一个关键词，可为空">
          <Textarea v-model="keywordsModel" class="min-h-12 resize-none py-2 font-mono text-xs" placeholder="DT Swiss 350&#10;AR45&#10;28H 2-cross" />
        </AdminFormField>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit">
            {{ mode === 'create' ? '加入预设列表' : '更新预设' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, toRef } from 'vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import type {
  SpokeBrand,
  SpokeBuildActualLengths,
  SpokeBuildPreset,
  SpokeCatalogOptions,
  SpokeHubModel,
  SpokeRimModel,
} from '@/api/spokeCatalog'

type ActualLengthProp = keyof Pick<SpokeBuildActualLengths, 'frontLeft' | 'frontRight' | 'rearLeft' | 'rearRight'>

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: 'create' | 'edit'
  form: SpokeBuildPreset
  rims: SpokeBrand<SpokeRimModel>[]
  hubs: SpokeBrand<SpokeHubModel>[]
  options: SpokeCatalogOptions
}>(), {
  open: false,
  mode: 'create',
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
}>()

const form = toRef(props, 'form')
const selectClass = 'h-9 w-full min-w-0 rounded-xl border-none bg-muted/50 px-3 text-xs font-bold outline-none transition focus:ring-2 focus:ring-ring/40'
const actualLengthFields: Array<{ key: ActualLengthProp; label: string }> = [
  { key: 'frontLeft', label: '前轮左侧' },
  { key: 'frontRight', label: '前轮右侧' },
  { key: 'rearLeft', label: '后轮左侧' },
  { key: 'rearRight', label: '后轮右侧' },
]

const descriptionModel = computed({
  get: () => form.value.description || '',
  set: (value: string) => { form.value.description = value },
})

const keywordsModel = computed({
  get: () => form.value.keywords.join('\n'),
  set: (value: string) => {
    form.value.keywords = value
      .split(/\r?\n/)
      .map((item) => item.trim())
      .filter(Boolean)
  },
})

const actualLengths = computed<SpokeBuildActualLengths>(() => {
  if (!form.value.actualLengths) {
    form.value.actualLengths = {
      frontLeft: null,
      frontRight: null,
      rearLeft: null,
      rearRight: null,
      notes: '',
    }
  }
  return form.value.actualLengths
})

const rimModelsFor = (brandId: string) => rimsForBrand(brandId, props.rims)
const hubModelsFor = (brandId: string) => hubsForBrand(brandId, props.hubs)

const rimsForBrand = (brandId: string, brands: SpokeBrand<SpokeRimModel>[]) => (
  brands.find((brand) => brand.id === brandId)?.items || []
)

const hubsForBrand = (brandId: string, brands: SpokeBrand<SpokeHubModel>[]) => (
  brands.find((brand) => brand.id === brandId)?.items || []
)

const handleRimBrandChange = (value: string | undefined) => {
  form.value.rimBrandId = value || ''
  const models = rimModelsFor(form.value.rimBrandId)
  if (!models.some((model) => model.id === form.value.rimModelId)) {
    form.value.rimModelId = models[0]?.id || ''
  }
}

const handleHubBrandChange = (value: string | undefined) => {
  form.value.hubBrandId = value || ''
  const models = hubModelsFor(form.value.hubBrandId)
  if (!models.some((model) => model.id === form.value.hubModelId)) {
    form.value.hubModelId = models[0]?.id || ''
  }
}

const numberInputValue = (value: number | null | undefined) => value == null ? '' : String(value)

const nullableNumber = (value: string | number) => {
  const normalized = String(value).trim()
  if (!normalized) return null
  const parsed = Number(normalized)
  return Number.isFinite(parsed) ? parsed : null
}

const setActualLength = (field: ActualLengthProp, value: string | number) => {
  actualLengths.value[field] = nullableNumber(value)
}
</script>
