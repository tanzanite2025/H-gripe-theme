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
              <div v-if="query" class="text-[11px] tz-text-muted truncate">
                Keyword: <span class="tz-text-primary">{{ query }}</span>
              </div>
            </div>
            <button
              type="button"
              class="wa-drawer-close-btn"
              @click="handleClose"
            >
              <span class="text-lg leading-none">x</span>
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
                        <div v-if="product.price" class="text-xs text-[#40ffaa] mt-1">
                          {{ product.price }}
                        </div>
                        <div class="mt-3 flex flex-wrap gap-2">
                          <button
                            type="button"
                            class="flex-1 min-w-[120px] px-3 py-1.5 rounded-full bg-[#40ffaa]/90 text-[11px] text-[#07120b] border border-[#86efac]/70 hover:bg-[#86efac] transition-colors"
                            @click="handleAddToCart(product)"
                          >
                            加入购物车
                          </button>
                          <button
                            type="button"
                            class="flex-1 min-w-[120px] px-3 py-1.5 rounded-full bg-white/10 text-[11px] tz-text-primary border border-white/30 hover:bg-white/20 transition-colors"
                            @click="handleShareToChat(product)"
                          >
                            分享到聊天
                          </button>
                          <button
                            type="button"
                            class="flex-1 min-w-[140px] px-3 py-1.5 rounded-full bg-[#6b73ff]/90 text-[11px] text-white border border-[#a5b4fc]/60 hover:bg-[#818cf8] transition-colors"
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
                  class="inline-flex items-center gap-1 px-3 py-1.5 rounded-full border border-white/30 text-[11px] tz-text-secondary hover:bg-white/10 transition-colors"
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
                        v-if="selectedConfigProduct.price"
                        class="text-xs text-[#40ffaa] mt-1"
                      >
                        {{ selectedConfigProduct.price }}
                      </div>
                      <div class="mt-2 text-[11px] tz-text-secondary line-clamp-2">
                        待完善的配置详情占位文案，后续将展示戒托、主石、预算等具体字段。
                      </div>
                    </div>
                  </div>

                  <div
                    class="border border-dashed border-white/20 rounded-xl bg-white/[0.02] p-3 md:p-4 text-[12px] tz-text-secondary"
                  >
                    配置选项区域占位：这里将展示可选参数（材质、主石大小、预算等），仅为视觉占位，不会影响当前聊天。
                  </div>
                </div>

                <div class="border border-white/10 rounded-xl bg-white/[0.04] p-3 md:p-4 flex flex-col gap-3">
                  <div class="text-xs tz-text-secondary">
                    当前为占位体验，暂不发送真实配置到聊天。后续版本会生成一条结构化的配置卡片消息，方便客服快速了解你的需求。
                  </div>
                  <button
                    type="button"
                    class="mt-1 inline-flex items-center justify-center px-4 py-2.5 rounded-full bg-white/10 text-xs font-medium tz-text-disabled cursor-not-allowed border border-white/30"
                    disabled
                  >
                    发送配置给客服（即将上线）
                  </button>
                  <div class="text-[11px] tz-text-muted">
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
import { ref, watch } from 'vue'

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
}>()

const viewMode = ref<'list' | 'configConfirm'>('list')
const selectedConfigProduct = ref<any | null>(null)

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
  viewMode.value = 'configConfirm'
}

const backToList = () => {
  viewMode.value = 'list'
  selectedConfigProduct.value = null
}

watch(
  () => props.modelValue,
  value => {
    if (!value) {
      viewMode.value = 'list'
      selectedConfigProduct.value = null
    }
  }
)
</script>
