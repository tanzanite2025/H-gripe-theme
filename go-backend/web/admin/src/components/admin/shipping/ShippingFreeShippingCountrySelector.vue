<template>
  <div class="space-y-3 rounded-lg border bg-background p-3">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div>
        <p class="text-xs font-bold">免邮国家范围</p>
        <p class="mt-1 text-[11px] leading-4 text-muted-foreground">只有选中的国家可以使用此系统免邮模板，其他国家无法提交订单。</p>
      </div>
      <span class="rounded-full bg-emerald-500/10 px-2 py-1 text-[11px] font-semibold text-emerald-700 dark:text-emerald-300">
        已选 {{ selectedCountryCodes.length }} 个国家
      </span>
    </div>

    <div class="flex flex-wrap gap-2">
      <Input v-model.trim="searchKeyword" class="min-w-64 flex-1" placeholder="搜索国家名称或代码" />
      <Button type="button" variant="outline" size="sm" @click="selectAllVisibleCountries">全选当前结果</Button>
      <Button type="button" variant="ghost" size="sm" @click="clearAllSelectedCountries">清空</Button>
    </div>

    <div v-if="filteredCountryOptions.length" class="grid max-h-56 gap-2 overflow-y-auto rounded-md border p-2 sm:grid-cols-2 lg:grid-cols-4">
      <label
        v-for="country in filteredCountryOptions"
        :key="country.code"
        class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-xs transition hover:bg-muted/50"
      >
        <input
          type="checkbox"
          class="size-3.5 accent-[var(--admin-selected)]"
          :checked="selectedCountryCodes.includes(country.code)"
          @change="toggleCountrySelection(country.code, $event)"
        />
        <span class="font-mono text-[10px] text-muted-foreground">{{ country.code }}</span>
        <span class="truncate">{{ country.nameZh }} / {{ country.name }}</span>
      </label>
    </div>
    <p v-else class="rounded-md border border-dashed px-3 py-4 text-center text-xs text-muted-foreground">没有匹配的国家</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { SHIPPING_COUNTRY_OPTIONS } from '@/lib/shippingCountryCatalog'

const props = defineProps<{ modelValue?: string | string[] | null }>()
const emit = defineEmits<{ (event: 'update:modelValue', value: string): void }>()
const searchKeyword = ref('')

const parseSelectedCountryCodes = (value: unknown): string[] => {
  if (Array.isArray(value)) return value.map(code => String(code || '').trim().toUpperCase()).filter(Boolean)
  const raw = String(value || '').trim()
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parseSelectedCountryCodes(parsed)
  } catch {
    // Keep compatibility with old comma-separated country scopes.
  }
  return raw.split(/[,，;|\s]+/).map(code => code.trim().toUpperCase()).filter(Boolean)
}

const selectedCountryCodes = computed(() => parseSelectedCountryCodes(props.modelValue))
const filteredCountryOptions = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return SHIPPING_COUNTRY_OPTIONS
  return SHIPPING_COUNTRY_OPTIONS.filter(country => (
    country.code.toLowerCase().includes(keyword)
    || country.name.toLowerCase().includes(keyword)
    || country.nameZh.includes(keyword)
  ))
})

const emitSelectedCountryCodes = (codes: string[]) => {
  const availableCodes = new Set(SHIPPING_COUNTRY_OPTIONS.map(country => country.code))
  const selected = Array.from(new Set(codes.map(code => code.toUpperCase())))
    .filter(code => availableCodes.has(code))
    .sort((left, right) => left.localeCompare(right))
  emit('update:modelValue', JSON.stringify(selected))
}

const toggleCountrySelection = (code: string, event: Event) => {
  const checked = (event.target as HTMLInputElement | null)?.checked === true
  const next = selectedCountryCodes.value.filter(countryCode => countryCode !== code)
  if (checked) next.push(code)
  emitSelectedCountryCodes(next)
}

const selectAllVisibleCountries = () => {
  emitSelectedCountryCodes([...selectedCountryCodes.value, ...filteredCountryOptions.value.map(country => country.code)])
}

const clearAllSelectedCountries = () => emitSelectedCountryCodes([])
</script>
