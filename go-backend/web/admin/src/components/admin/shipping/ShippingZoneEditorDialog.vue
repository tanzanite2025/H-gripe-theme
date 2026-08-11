<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[calc(100dvh-1rem)] sm:max-h-[90dvh]" @open-auto-focus.prevent>
      <form class="space-y-4" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增配送区域' : '编辑配送区域' }}</DialogTitle>
          <DialogDescription>
            从地址库搜索并选择国家/地区，系统会自动保存标准区域代码。
          </DialogDescription>
        </DialogHeader>

        <div class="grid items-end gap-4 sm:grid-cols-[minmax(0,1fr)_auto]">
          <AdminFormField label="区域名称" required :error="errors.name">
            <Input v-model.trim="form.name" placeholder="例如 北美、欧盟、东南亚" @input="emit('clear-error', 'name')" />
          </AdminFormField>

          <div class="flex min-h-9 items-center justify-between gap-3 rounded-xl bg-muted/45 px-3 py-2">
            <div>
              <span class="text-[10px] font-black uppercase tracking-wider">启用区域</span>
              <p class="mt-0.5 text-[9px] font-bold text-muted-foreground">停用后不参与匹配</p>
            </div>
            <Switch v-model="form.enabled" size="sm" aria-label="启用配送区域" />
          </div>
        </div>

        <AdminFormField label="配送国家/地区" required :error="errors.countries">
          <div class="rounded-2xl bg-muted/35 p-3">
            <div class="relative">
              <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/65" />
              <Input
                v-model="searchKeyword"
                class="bg-background/85 pl-9"
                placeholder="搜索中文名、英文名或二位码"
                autocomplete="off"
              />
            </div>

            <div class="mt-3 flex items-center justify-between gap-3">
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
                已选 {{ selectedCodes.length }} 个
              </span>
              <Button
                v-if="selectedCodes.length"
                type="button"
                variant="ghost"
                size="xs"
                @click="clearSelected"
              >
                清空
              </Button>
            </div>

            <div v-if="selectedCodes.length" class="mt-2 flex flex-wrap gap-1.5">
              <button
                v-for="code in selectedCodes"
                :key="`selected-${code}`"
                type="button"
                class="inline-flex max-w-full items-center gap-1.5 rounded-full bg-primary px-2.5 py-1 text-[10px] font-black text-primary-foreground transition-opacity hover:opacity-85"
                :aria-label="`移除${optionForCode(code)?.name || code}`"
                @click="toggleRegion(code)"
              >
                <span class="max-w-44 truncate">{{ optionForCode(code)?.name || code }}</span>
                <span class="font-mono text-[9px] opacity-75">{{ code }}</span>
                <X class="size-3 shrink-0" />
              </button>
            </div>

            <div class="mt-3 grid max-h-64 gap-1 overflow-y-auto pr-1 sm:grid-cols-2">
              <button
                v-for="option in filteredRegionOptions"
                :key="option.code"
                type="button"
                class="flex min-w-0 items-center gap-2 rounded-xl px-2.5 py-2 text-left text-[11px] font-bold transition-colors"
                :class="isSelected(option.code) ? 'bg-primary text-primary-foreground' : 'bg-background/75 text-foreground hover:bg-background'"
                :aria-pressed="isSelected(option.code)"
                @click="toggleRegion(option.code)"
              >
                <span class="flex size-4 shrink-0 items-center justify-center rounded-full" :class="isSelected(option.code) ? 'bg-primary-foreground/20' : 'bg-muted'">
                  <Check v-if="isSelected(option.code)" class="size-3" />
                </span>
                <span class="min-w-0 flex-1 truncate">{{ option.name }}</span>
                <span class="shrink-0 font-mono text-[9px]" :class="isSelected(option.code) ? 'text-primary-foreground/70' : 'text-muted-foreground/70'">
                  {{ option.code }}
                </span>
              </button>
            </div>

            <p v-if="filteredRegionOptions.length === 0" class="py-8 text-center text-[11px] font-bold text-muted-foreground">
              没有找到对应的国家/地区
            </p>
          </div>
        </AdminFormField>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存区域' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref, toRef, watch } from 'vue'
import { Check, LoaderCircle, Search, X } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  addressRegionOptions,
  parseAddressRegionCodes,
  serializeAddressRegionCodes,
} from '@/lib/addressRegions'
import type { ShippingDialogMode, ShippingErrorMap, ShippingZoneForm } from './shippingTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: ShippingDialogMode
  form: ShippingZoneForm
  errors: ShippingErrorMap
  submitting?: boolean
}>(), {
  open: false,
  mode: 'create',
  submitting: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', field: string): void
}>()

const form = toRef(props, 'form')
const searchKeyword = ref('')
const selectedCodes = ref<string[]>([])
const regionByCode = new Map(addressRegionOptions.map((option) => [option.code, option]))

const filteredRegionOptions = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return addressRegionOptions
  return addressRegionOptions.filter((option) => option.keywords.includes(keyword))
})

const optionForCode = (code: string) => regionByCode.get(code)
const isSelected = (code: string) => selectedCodes.value.includes(code)

const syncSelectedCodes = () => {
  selectedCodes.value = parseAddressRegionCodes(form.value.countries)
}

const updateSelectedCodes = (codes: string[]) => {
  selectedCodes.value = codes
  form.value.countries = serializeAddressRegionCodes(codes)
  emit('clear-error', 'countries')
}

const toggleRegion = (code: string) => {
  const nextCodes = isSelected(code)
    ? selectedCodes.value.filter((selectedCode) => selectedCode !== code)
    : [...selectedCodes.value, code]

  updateSelectedCodes(nextCodes)
}

const clearSelected = () => {
  updateSelectedCodes([])
}

watch(() => form.value.countries, syncSelectedCodes, { immediate: true })
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    searchKeyword.value = ''
    syncSelectedCodes()
  }
})
</script>
