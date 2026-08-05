<template>
  <Teleport to="body">
    <Transition name="wa-drawer">
      <div
        v-if="modelValue"
        class="wa-drawer-mask"
      >
        <!-- Backdrop -->
        <div 
          class="wa-drawer-backdrop md:hidden"
          @click="handleClose"
        />

        <div class="wa-drawer-shell">
          <!-- Header -->
          <div class="wa-drawer-header">
            <div class="flex flex-col gap-1 min-w-0">
              <div class="wa-drawer-title">
                Search results
                <span v-if="agent" class="text-xs tz-text-secondary ml-1">({{ agent.name }})</span>
              </div>
              <div v-if="query" class="tz-caption tz-text-muted truncate">
                Keyword: <span class="tz-text-primary">{{ query }}</span>
              </div>
            </div>
            <button
              type="button"
              class="wa-drawer-close-btn"
              @click="handleClose"
            >
              <Icon name="lucide:x" class="h-3.5 w-3.5" />
            </button>
          </div>

          <!-- Content -->
          <div class="wa-drawer-content">
            <div v-if="viewMode === 'list'" class="h-full">
              <div
                v-if="loading"
                class="flex flex-col items-center justify-center h-full tz-text-secondary text-sm gap-3"
              >
                <svg class="animate-spin h-6 w-6 tz-text-secondary" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                <span>Searching products...</span>
              </div>

              <div
                v-else-if="error"
                class="flex items-center justify-center h-full text-red-300 text-sm text-center px-4"
              >
                {{ error }}
              </div>

              <div v-else class="h-full flex flex-col">
                <div
                  v-if="!results || results.length === 0"
                  class="flex-1 flex items-start justify-center tz-text-secondary text-sm text-center px-4 pt-4"
                >
                  <span>
                    {{ query ? 'No products found' : '' }}
                  </span>
                </div>

                <div v-else class="flex-1">
                  <div
                    class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 md:gap-4"
                  >
                    <div
                      v-for="product in results"
                      :key="product.id"
                      class="border border-white/10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08]
                             transition-colors overflow-hidden text-left flex flex-col"
                    >
                      <img
                        v-if="product.thumbnail"
                        :src="product.thumbnail"
                        alt="Product image"
                        class="w-full h-32 object-cover rounded-t-xl"
                      />
                      <div class="px-3 pt-2 pb-3 flex-1 flex flex-col">
                        <div class="text-sm font-semibold text-white truncate">
                          {{ product.title }}
                        </div>
                        <div v-if="product.price" class="text-xs text-[#B5FF6D] mt-1">
                          {{ product.price }}
                        </div>
                        <div v-if="product.variants?.length" class="mt-1 tz-caption tz-text-muted">
                          {{ product.variants.length }} SKU option{{ product.variants.length > 1 ? 's' : '' }}
                        </div>
                        <div class="mt-3 flex flex-wrap gap-2">
                          <button
                            type="button"
                            class="flex-1 min-w-[120px] px-3 py-1.5 rounded-full bg-[#B5FF6D]/90 tz-caption text-[#07120b] border border-[#D3FFA8]/70 hover:bg-[#D3FFA8] transition-colors"
                            @click="handleAddToCart(product)"
                          >
                            加入购物车
                          </button>
                          <button
                            type="button"
                            class="flex-1 min-w-[120px] px-3 py-1.5 rounded-full bg-white/10 tz-caption tz-text-primary border border-white/30 hover:bg-white/20 transition-colors"
                            @click="handleShareToChat(product)"
                          >
                            分享到聊天
                          </button>
                          <button
                            type="button"
                            class="flex-1 min-w-[140px] px-3 py-1.5 rounded-full bg-white tz-caption text-slate-950 border border-white/70 hover:bg-white/90 transition-colors"
                            @click="openConfigConfirm(product)"
                          >
                            和客服确认配置
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div
              v-else-if="viewMode === 'configConfirm' && selectedConfigProduct"
              class="h-full flex flex-col gap-4 md:gap-6"
            >
              <div class="flex items-center justify-between">
                <button
                  type="button"
                  class="inline-flex items-center gap-1 px-3 py-1.5 rounded-full border border-white/30 tz-caption tz-text-secondary hover:bg-white/10 transition-colors"
                  @click="backToList"
                >
                  <span class="text-xs">←</span>
                  <span>返回商品列表</span>
                </button>
                <div class="text-sm font-semibold tz-text-primary truncate">
                  和客服确认配置
                </div>
                <div class="w-8" />
              </div>

              <div class="grid grid-cols-1 md:grid-cols-[minmax(0,1.3fr)_minmax(0,1fr)] gap-4 md:gap-6">
                <div class="space-y-4">
                  <div class="border border-white/10 rounded-xl bg-white/[0.04] p-3 md:p-4 flex gap-3">
                    <img
                      v-if="selectedConfigProduct.thumbnail"
                      :src="selectedConfigProduct.thumbnail"
                      alt="Product image"
                      class="w-20 h-20 object-cover rounded-lg flex-shrink-0"
                    />
                    <div class="flex-1 min-w-0">
                      <div class="text-sm font-semibold text-white truncate">
                        {{ selectedConfigProduct.title }}
                      </div>
                      <div
                        v-if="selectedConfigPriceLabel"
                        class="text-xs text-[#B5FF6D] mt-1"
                      >
                        {{ selectedConfigPriceLabel }}
                      </div>
                      <div v-if="selectedConfigVariant?.sku || selectedConfigProduct.sku" class="mt-2 tz-caption tz-text-secondary">
                        SKU: {{ selectedConfigVariant?.sku || selectedConfigProduct.sku }}
                      </div>
                      <div v-if="selectedConfigVariant?.weightGrams" class="mt-1 tz-caption tz-text-secondary">
                        Weight: {{ selectedConfigVariant.weightGrams }}g
                      </div>
                      <div v-if="selectedConfigVariant" class="mt-1 tz-caption tz-text-secondary">
                        {{ selectedConfigVariant.availability === 'in_stock' ? 'Available' : 'Out of stock' }}
                      </div>
                    </div>
                  </div>

                  <div
                    class="border border-dashed border-white/20 rounded-xl bg-white/[0.02] p-3 md:p-4"
                  >
                    <label
                      v-if="configVariants.length > 1"
                      class="block tz-compact-label font-semibold tz-text-muted"
                      for="wa-config-variant"
                    >
                      Choose SKU
                    </label>
                    <select
                      v-if="configVariants.length > 1"
                      id="wa-config-variant"
                      v-model.number="selectedConfigVariantId"
                      class="mt-2 w-full rounded-xl border border-white/20 bg-slate-950/90 px-3 py-2 text-xs text-white focus:border-[#6b73ff] focus:outline-none"
                    >
                      <option
                        v-for="variant in configVariants"
                        :key="variant.id"
                        :value="variant.id"
                      >
                        {{ variantLabel(variant) }}
                      </option>
                    </select>

                    <div class="mt-3 tz-caption tz-text-secondary">
                      <div class="mb-2 tz-compact-label font-semibold tz-text-muted">
                        Configuration facts
                      </div>
                      <dl v-if="selectedConfigOptionRows.length" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
                        <div
                          v-for="option in selectedConfigOptionRows"
                          :key="option.key"
                          class="rounded-xl border border-white/10 bg-white/[0.04] px-3 py-2"
                        >
                          <dt class="tz-compact-label tz-text-muted">
                            {{ option.label }}
                          </dt>
                          <dd class="mt-1 text-xs font-semibold text-white">
                            {{ option.value }}<span v-if="option.unit"> {{ option.unit }}</span>
                          </dd>
                        </div>
                      </dl>
                      <p v-else class="rounded-xl border border-white/10 bg-white/[0.04] px-3 py-2 tz-caption">
                        This product currently has no variant option fields. The selected SKU will still be sent as the configuration fact.
                      </p>
                    </div>
                  </div>
                </div>

                <div class="border border-white/10 rounded-xl bg-white/[0.04] p-3 md:p-4 flex flex-col gap-3">
                  <div class="tz-caption tz-text-secondary">
                    The selected product and SKU facts will be sent as a structured card. Staff can read them in Admin without editing or duplicating product data.
                  </div>
                  <button
                    type="button"
                    class="mt-1 inline-flex items-center justify-center px-4 py-2.5 rounded-full bg-white text-xs font-medium text-slate-950 border border-white/70 hover:bg-white/90 transition-colors"
                    @click="handleConfirmConfig"
                  >
                    发送配置给客服
                  </button>
                  <div class="tz-caption tz-text-muted">
                    你仍然可以在聊天中手动描述戒指款式和预算，我们的客服会协助推荐合适的产品。
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ShopProductSpecDefinition, ShopProductVariant } from '~/composables/useShopProducts'

const props = defineProps<{
  modelValue: boolean
  loading: boolean
  results: any[]
  error?: string | null
  agent?: any | null
  query?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
  (e: 'select', product: any): void
  (e: 'add-to-cart', product: any): void
  (e: 'confirm-config', product: any): void
}>()

const viewMode = ref<'list' | 'configConfirm'>('list')
const selectedConfigProduct = ref<any | null>(null)
const selectedConfigVariantId = ref<number | null>(null)

const configVariants = computed<ShopProductVariant[]>(() => {
  return Array.isArray(selectedConfigProduct.value?.variants)
    ? selectedConfigProduct.value.variants
    : []
})

const selectedConfigVariant = computed<ShopProductVariant | null>(() => {
  if (configVariants.value.length === 0) return null
  const selectedId = Number(selectedConfigVariantId.value || 0)
  return configVariants.value.find(variant => Number(variant.id) === selectedId)
    || configVariants.value.find(variant => variant.isDefault)
    || configVariants.value[0]
    || null
})

const specDefinitions = computed<ShopProductSpecDefinition[]>(() => {
  return Array.isArray(selectedConfigProduct.value?.productType?.specDefinitions)
    ? selectedConfigProduct.value.productType.specDefinitions
    : []
})

const specLabel = (key: string) => {
  const definition = specDefinitions.value.find(item => item.slug === key)
  if (definition?.name) return definition.name
  return key
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, char => char.toUpperCase())
}

const specUnit = (key: string) => {
  return specDefinitions.value.find(item => item.slug === key)?.unit || ''
}

const selectedConfigOptionRows = computed(() => {
  const options = selectedConfigVariant.value?.optionValues || {}
  return Object.entries(options).map(([key, value]) => ({
    key,
    label: specLabel(key),
    value,
    unit: specUnit(key),
  }))
})

const selectedConfigPriceLabel = computed(() => {
  const price = selectedConfigVariant.value?.priceNumber
  if (price && price > 0) return `$${price}`
  return selectedConfigProduct.value?.price || selectedConfigProduct.value?.priceLabel || ''
})

const variantLabel = (variant: ShopProductVariant) => {
  const optionValues = Object.values(variant.optionValues || {}).filter(Boolean)
  const optionText = optionValues.length ? ` · ${optionValues.join(' / ')}` : ''
  const weightText = variant.weightGrams ? ` · ${variant.weightGrams}g` : ''
  return `${variant.title || variant.sku}${optionText}${weightText}`
}

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleShareToChat = (product: any) => {
  emit('select', product)
}

const handleAddToCart = (product: any) => {
  emit('add-to-cart', product)
}

const openConfigConfirm = (product: any) => {
  selectedConfigProduct.value = product
  const variants = Array.isArray(product?.variants) ? product.variants : []
  const defaultVariant = variants.find((variant: ShopProductVariant) => variant.isDefault)
    || variants.find((variant: ShopProductVariant) => Number(variant.id) === Number(product.defaultVariantId || 0))
    || variants[0]
    || null
  selectedConfigVariantId.value = defaultVariant?.id || product?.defaultVariantId || null
  viewMode.value = 'configConfirm'
}

const handleConfirmConfig = () => {
  if (!selectedConfigProduct.value) return
  emit('confirm-config', {
    product: selectedConfigProduct.value,
    variant: selectedConfigVariant.value,
  })
}

const backToList = () => {
  viewMode.value = 'list'
  selectedConfigProduct.value = null
  selectedConfigVariantId.value = null
}

watch(
  () => props.modelValue,
  value => {
    if (!value) {
      viewMode.value = 'list'
      selectedConfigProduct.value = null
      selectedConfigVariantId.value = null
    }
  }
)
</script>

