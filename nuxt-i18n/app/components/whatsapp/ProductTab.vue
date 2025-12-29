<template>
  <div class="flex flex-col h-full overflow-hidden">
    <!-- 搜索栏 -->
    <div class="flex-none p-3 md:p-6 pb-0 md:pb-0">
      <div class="flex gap-2 mb-3 items-center">
        <input
          :value="searchQuery"
          @input="$emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
          type="text"
          placeholder="Search products..."
          class="flex-1 h-10 md:h-[42px] px-3 rounded-xl md:rounded-lg text-white text-sm focus:outline-none transition-colors"
          :class="[
            'bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))]',
            'shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)]',
          ]"
          @keydown.enter.prevent="$emit('search')"
        />
        <button
          @click="$emit('search')"
          :disabled="isSearching"
          class="h-10 md:h-[42px] px-3 md:px-4 rounded-xl md:rounded-lg text-sm font-semibold disabled:opacity-50 transition-colors whitespace-nowrap"
          :class="isSearching
            ? 'bg-[rgba(15,23,42,0.98)] text-white/70 shadow-[0_2px_6px_-4px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)]'
            : 'bg-[linear-gradient(135deg,rgba(45,212,191,0.7),rgba(59,130,246,0.85))] text-white shadow-[0_4px_14px_-8px_rgba(59,130,246,0.8),0_0_14px_rgba(45,212,191,0.55)]'"
        >
          {{ isSearching ? 'Searching...' : 'Search' }}
        </button>
      </div>

      <!-- Actions Loop: History / Cart / Wishlist -->
      <div class="flex justify-center gap-1.5 md:gap-3 mb-3 md:mb-4">
        <button
          type="button"
          @click="$emit('openHistory')"
          class="flex-1 md:flex-none md:px-4 h-10 md:h-[34px] rounded-full md:rounded-full text-[11px] md:text-sm font-semibold md:font-medium tracking-wide flex items-center justify-center gap-1.5 transition-all bg-[rgba(31,41,55,0.9)] text-white shadow-[0_3px_9px_rgba(0,0,0,0.9)] hover:bg-[rgba(51,65,85,0.95)]"
          :style="{ borderColor: currentThemeColor }"
        >
          <svg class="w-3.5 h-3.5 md:w-4 md:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="8" stroke-width="1.7" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12 8v4l2.5 2.5" />
          </svg>
          <span>History</span>
        </button>
        <button
          type="button"
          @click="$emit('openCart')"
          class="flex-1 md:flex-none md:px-4 h-10 md:h-[34px] rounded-full text-[11px] md:text-sm font-semibold md:font-medium tracking-wide flex items-center justify-center gap-1.5 transition-all bg-[rgba(31,41,55,0.9)] text-white shadow-[0_3px_9px_rgba(0,0,0,0.9)] hover:bg-[rgba(51,65,85,0.95)]"
          :style="{ borderColor: currentThemeColor }"
        >
          <svg class="w-3.5 h-3.5 md:w-4 md:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M3 3h2l2 13h12l2-9H6" />
            <circle cx="9" cy="19" r="1.4" />
            <circle cx="17" cy="19" r="1.4" />
          </svg>
          <span>Cart</span>
        </button>
        <button
          type="button"
          @click="$emit('openWishlist')"
          class="flex-1 md:flex-none md:px-4 h-10 md:h-[34px] rounded-full text-[11px] md:text-sm font-semibold md:font-medium tracking-wide flex items-center justify-center gap-1.5 transition-all bg-[rgba(31,41,55,0.9)] text-white shadow-[0_3px_9px_rgba(0,0,0,0.9)] hover:bg-[rgba(51,65,85,0.95)]"
          :style="{ borderColor: currentThemeColor }"
        >
          <svg class="w-3.5 h-3.5 md:w-4 md:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z" />
          </svg>
          <span>Wishlist</span>
        </button>
      </div>
    </div>

    <!-- 结果列表 -->
    <div v-if="!productDrawerVisible" class="flex-1 overflow-y-auto px-1 md:p-6 md:pt-0 pb-3">
      <div v-if="searchResults.length > 0" class="space-y-3 md:space-y-0 md:grid md:grid-cols-2 md:gap-3">
        <div
          v-for="product in searchResults"
          :key="product.id"
          @click="$emit('shareProduct', product)"
          class="border border-white/10 rounded-2xl md:rounded-lg p-3 bg-black/30 hover:bg-white/[0.05] cursor-pointer transition-colors"
        >
          <img
            v-if="product.thumbnail"
            :src="product.thumbnail"
            alt="Product"
            class="w-full h-28 md:h-32 object-cover rounded-xl md:rounded-lg mb-2"
          />
          <h4 class="text-white text-sm font-semibold md:font-medium truncate">{{ product.title }}</h4>
          <p v-if="product.price" class="text-white/70 text-xs mt-1">{{ product.price }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  searchQuery: string
  isSearching: boolean
  searchResults: any[]
  productDrawerVisible: boolean
  currentThemeColor: string
}>()

defineEmits<{
  'update:searchQuery': [value: string]
  'search': []
  'shareProduct': [product: any]
  'openHistory': []
  'openCart': []
  'openWishlist': []
}>()
</script>
