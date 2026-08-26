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
          class="flex-1 h-10 md:h-[42px] px-3 rounded-xl md:rounded-lg tz-text-primary text-sm focus:outline-none transition-colors"
          :class="[
            'tz-surface-input border tz-border-subtle',
            'shadow-[0_2px_6px_rgba(20,32,43,0.08)]',
          ]"
          @keydown.enter.prevent="$emit('search')"
        />
        <button
          @click="$emit('search')"
          :disabled="isSearching"
          class="h-10 md:h-[42px] px-3 md:px-4 rounded-xl md:rounded-lg text-sm font-semibold disabled:opacity-50 transition-colors whitespace-nowrap"
          :class="isSearching
            ? 'tz-surface-muted tz-text-disabled shadow-[0_2px_6px_rgba(20,32,43,0.08)]'
            : 'bg-white text-slate-950 shadow-[0_4px_14px_-8px_rgba(20,32,43,0.16)] hover:bg-slate-50'"
        >
          {{ isSearching ? 'Searching...' : 'Search' }}
        </button>
      </div>

      <!-- Actions Loop: History / Cart / Wishlist -->
      <div class="flex justify-center gap-1.5 md:gap-3 mb-3 md:mb-4">
        <button
          type="button"
          @click="$emit('openHistory')"
          class="flex-1 md:flex-none md:px-4 h-10 md:h-[34px] rounded-full md:rounded-full tz-caption md:text-sm font-semibold md:font-medium tracking-wide flex items-center justify-center gap-1.5 transition-all tz-surface-muted tz-text-primary shadow-[0_3px_9px_rgba(20,32,43,0.1)] hover:tz-surface-subtle"
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
          class="flex-1 md:flex-none md:px-4 h-10 md:h-[34px] rounded-full tz-caption md:text-sm font-semibold md:font-medium tracking-wide flex items-center justify-center gap-1.5 transition-all tz-surface-muted tz-text-primary shadow-[0_3px_9px_rgba(20,32,43,0.1)] hover:tz-surface-subtle"
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
          class="flex-1 md:flex-none md:px-4 h-10 md:h-[34px] rounded-full tz-caption md:text-sm font-semibold md:font-medium tracking-wide flex items-center justify-center gap-1.5 transition-all tz-surface-muted tz-text-primary shadow-[0_3px_9px_rgba(20,32,43,0.1)] hover:tz-surface-subtle"
          :style="{ borderColor: currentThemeColor }"
        >
          <svg class="w-3.5 h-3.5 md:w-4 md:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z" />
          </svg>
          <span>Wishlist</span>
        </button>
      </div>
    </div>

    <!-- 搜索结果统一由 WhatsAppProductSearchResultDrawer 展示，避免 Tab 内重复渲染 -->
    <div class="flex-1" aria-hidden="true" />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  searchQuery: string
  isSearching: boolean
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

