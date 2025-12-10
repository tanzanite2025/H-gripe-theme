<template>
  <Teleport to="body">
    <Transition name="whatsapp-product-drawer">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[10001] flex items-end justify-center p-0 md:p-4 pointer-events-none"
      >
        <div
          class="pointer-events-auto w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]
                 rounded-2xl border-2 border-[#6b73ff]/40
                 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))]
                 backdrop-blur-md shadow-[0_0_30px_rgba(107,115,255,0.6)]
                 flex flex-col overflow-hidden"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-4 py-3 border-b border-white/10">
            <div class="flex flex-col gap-1 min-w-0">
              <div class="text-sm font-semibold text-white/90 truncate">
                Search results
                <span v-if="agent" class="text-xs text-white/60 ml-1">({{ agent.name }})</span>
              </div>
              <div v-if="query" class="text-[11px] text-white/50 truncate">
                Keyword: <span class="text-white/80">{{ query }}</span>
              </div>
            </div>
            <button
              type="button"
              class="w-8 h-8 rounded-full border border-white/40 text-white flex items-center justify-center hover:bg-white/10 transition-colors"
              @click="handleClose"
            >
              <span class="text-lg leading-none">x</span>
            </button>
          </div>

          <!-- Content -->
          <div class="flex-1 min-h-0 overflow-y-auto p-4 md:p-6">
            <div v-if="viewMode === 'list'" class="h-full">
              <div
                v-if="loading"
                class="flex flex-col items-center justify-center h-full text-white/70 text-sm gap-3"
              >
                <svg class="animate-spin h-6 w-6 text-white/60" fill="none" viewBox="0 0 24 24">
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
                  class="flex-1 flex items-start justify-center text-white/60 text-sm text-center px-4 pt-4"
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
                            class="flex-1 min-w-[120px] px-3 py-1.5 rounded-full bg-white/10 text-[11px] text-white/90 border border-white/30 hover:bg-white/20 transition-colors"
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
                  class="inline-flex items-center gap-1 px-3 py-1.5 rounded-full border border-white/30 text-[11px] text-white/80 hover:bg-white/10 transition-colors"
                  @click="backToList"
                >
                  <span class="text-xs">←</span>
                  <span>返回商品列表</span>
                </button>
                <div class="text-sm font-semibold text-white/90 truncate">
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
                      <div class="mt-2 text-[11px] text-white/60 line-clamp-2">
                        待完善的配置详情占位文案，后续将展示戒托、主石、预算等具体字段。
                      </div>
                    </div>
                  </div>

                  <div
                    class="border border-dashed border-white/20 rounded-xl bg-white/[0.02] p-3 md:p-4 text-[12px] text-white/70"
                  >
                    配置选项区域占位：这里将展示可选参数（材质、主石大小、预算等），仅为视觉占位，不会影响当前聊天。
                  </div>
                </div>

                <div class="border border-white/10 rounded-xl bg-white/[0.04] p-3 md:p-4 flex flex-col gap-3">
                  <div class="text-xs text-white/70">
                    当前为占位体验，暂不发送真实配置到聊天。后续版本会生成一条结构化的配置卡片消息，方便客服快速了解你的需求。
                  </div>
                  <button
                    type="button"
                    class="mt-1 inline-flex items-center justify-center px-4 py-2.5 rounded-full bg-white/10 text-xs font-medium text-white/70 cursor-not-allowed border border-white/30"
                    disabled
                  >
                    发送配置给客服（即将上线）
                  </button>
                  <div class="text-[11px] text-white/50">
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

<style scoped>
.whatsapp-product-drawer-enter-active,
.whatsapp-product-drawer-leave-active {
  transition: transform 0.3s ease-out, opacity 0.3s ease-out;
}

.whatsapp-product-drawer-enter-from,
.whatsapp-product-drawer-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.whatsapp-product-drawer-enter-to,
.whatsapp-product-drawer-leave-from {
  transform: translateY(0%);
  opacity: 1;
}
</style>
