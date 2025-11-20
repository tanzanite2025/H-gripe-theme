<template>
  <Teleport to="body">
    <Transition name="whatsapp-product-drawer">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[10001] flex items-end justify-center p-0 md:p-4 pointer-events-none"
      >
        <div
          class="pointer-events-auto w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]
                 rounded-2xl border-2 border-[#6b73ff] bg-black shadow-[0_0_30px_rgba(107,115,255,0.6)]
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

            <div
              v-else-if="!results || results.length === 0"
              class="flex items-center justify-center h-full text-white/60 text-sm text-center px-4"
            >
              <span>
                {{ query ? 'No products found' : 'Search products to share in chat' }}
              </span>
            </div>

            <div
              v-else
              class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 md:gap-4"
            >
              <button
                v-for="product in results"
                :key="product.id"
                type="button"
                class="border border-white/10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08]
                       cursor-pointer transition-colors overflow-hidden text-left"
                @click="$emit('select', product)"
              >
                <img
                  v-if="product.thumbnail"
                  :src="product.thumbnail"
                  alt="Product image"
                  class="w-full h-32 object-cover rounded-t-xl"
                />
                <div class="px-3 pt-2 pb-3">
                  <div class="text-sm font-semibold text-white truncate">
                    {{ product.title }}
                  </div>
                  <div v-if="product.price" class="text-xs text-[#40ffaa] mt-1">
                    {{ product.price }}
                  </div>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
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

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}
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
