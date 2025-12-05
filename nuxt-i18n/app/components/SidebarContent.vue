<template>
  <div class="w-full h-full p-[5px] box-border">
    <div class="w-full h-full flex flex-col">
      <!-- 左侧列：Product related -->
      <!-- 右侧列：Quick entry -->
      <div class="w-full pb-2 mb-2 border-b border-white/10 flex items-center justify-center">
        <div class="inline-flex rounded-full bg-white/5 p-[2px]">
          <button
            type="button"
            class="px-4 py-1.5 text-xs font-semibold rounded-full transition-colors"
            :class="activeTab === 'product' ? 'bg-[#6b73ff] text-black' : 'text-white/70 hover:text-white'"
            @click="activeTab = 'product'"
          >
            Product related
          </button>
          <button
            type="button"
            class="px-4 py-1.5 text-xs font-semibold rounded-full transition-colors"
            :class="activeTab === 'quick' ? 'bg-[#6b73ff] text-black' : 'text-white/70 hover:text-white'"
            @click="activeTab = 'quick'"
          >
            Quick entry
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto">
        <div v-if="activeTab === 'product'" class="pr-1 h-full">
          <!-- 复用 /shop 页完整的搜索 + 高级筛选容器 -->
          <ProductSearchPanel />
        </div>
        <div v-else class="h-full">
          <div class="w-full h-full flex items-center justify-center text-xs text-white/50">
            Quick entry content will be added here.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// 空侧边栏内容组件，后续由你自定义内部布局
import { ref } from 'vue'
import ProductSearchPanel from '~/components/ProductSearchPanel.vue'

const activeTab = ref<'product' | 'quick'>('product')
</script>

<style scoped>
</style>
