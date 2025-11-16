<template>
  <div class="w-full bg-white/[0.06] border border-white/[0.18] rounded-lg overflow-hidden">
    <!-- 标题栏 -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-white/10 bg-white/[0.03]">
      <div class="flex items-center gap-2">
        <svg class="w-5 h-5 text-white/70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 class="text-sm font-semibold text-white">Recently Viewed</h3>
        <span class="text-xs text-white/50">({{ historyCount }})</span>
      </div>
      <button
        v-if="hasHistory"
        @click="handleClearHistory"
        class="text-xs text-white/50 hover:text-red-400 transition-colors"
      >
        Clear History
      </button>
    </div>

    <!-- 商品列表 - 横向滚动 -->
    <div class="relative">
      <!-- 空状态 -->
      <div v-if="!hasHistory" class="flex flex-col items-center justify-center py-8 px-4">
        <svg class="w-16 h-16 text-white/20 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p class="text-white/50 text-sm">No browsing history</p>
        <p class="text-white/30 text-xs mt-1">Products you view will appear here</p>
      </div>
      
      <!-- 商品列表 -->
      <div
        v-else
        ref="scrollContainer"
        class="flex gap-3 p-4 overflow-x-auto scrollbar-hide scroll-smooth"
        style="scrollbar-width: none; -ms-overflow-style: none;"
      >
        <div
          v-for="item in history"
          :key="item.id"
          class="flex-shrink-0 w-40 group"
        >
          <!-- 商品卡片 -->
          <div class="relative bg-white/[0.06] border border-white/[0.18] rounded-lg overflow-hidden hover:bg-white/[0.10] hover:border-white/30 transition-all duration-200">
            <!-- 删除按钮 -->
            <button
              @click="handleRemoveItem(item.id)"
              class="absolute top-1 right-1 z-10 w-5 h-5 bg-black/70 hover:bg-red-600 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
              title="Remove"
            >
              <svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <!-- 商品图片 -->
            <div class="relative w-full h-32 bg-white/[0.03]">
              <img
                v-if="item.thumbnail"
                :src="item.thumbnail"
                :alt="item.title"
                class="w-full h-full object-cover"
                loading="lazy"
              />
              <div v-else class="w-full h-full flex items-center justify-center text-white/30">
                <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
            </div>

            <!-- 商品信息 -->
            <div class="p-2">
              <h4 class="text-xs font-medium text-white line-clamp-2 mb-1">
                {{ item.title }}
              </h4>
              <p class="text-sm font-semibold text-[#40ffaa] mb-2">
                {{ item.price }}
              </p>
              
              <!-- 操作按钮 -->
              <div class="flex gap-1.5">
                <!-- 查看详情按钮 -->
                <NuxtLink
                  :to="item.url"
                  class="flex-1 px-2 py-1.5 bg-white/10 hover:bg-white/20 border border-white/20 hover:border-white/40 rounded text-xs text-white text-center transition-all"
                  title="View Product"
                >
                  <svg class="w-3.5 h-3.5 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </NuxtLink>
                
                <!-- 分享到聊天按钮 -->
                <button
                  @click="(e) => handleShareToChat(e, item)"
                  class="flex-1 px-2 py-1.5 bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] hover:from-[#35e599] hover:to-[#5a62ee] rounded text-xs text-white transition-all shadow-lg"
                  title="Share to Chat"
                >
                  <svg class="w-3.5 h-3.5 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 左右滚动按钮（桌面端） -->
      <button
        v-if="hasHistory && showLeftArrow"
        @click="scrollLeft"
        class="hidden md:flex absolute left-2 top-1/2 -translate-y-1/2 w-8 h-8 bg-white/10 hover:bg-white/20 backdrop-blur-sm shadow-lg rounded-full items-center justify-center z-10 border border-white/20"
      >
        <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <button
        v-if="hasHistory && showRightArrow"
        @click="scrollRight"
        class="hidden md:flex absolute right-2 top-1/2 -translate-y-1/2 w-8 h-8 bg-white/10 hover:bg-white/20 backdrop-blur-sm shadow-lg rounded-full items-center justify-center z-10 border border-white/20"
      >
        <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <!-- 移动端滑动提示 -->
    <div v-if="hasHistory" class="md:hidden px-4 pb-3 text-center">
      <p class="text-xs text-white/40">← Swipe to see more →</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useBrowsingHistory } from '~/composables/useBrowsingHistory'

// 定义 emit 事件
const emit = defineEmits<{
  'share-to-chat': [product: any]
}>()

const { history, historyCount, hasHistory, clearHistory, removeItem } = useBrowsingHistory()

const scrollContainer = ref<HTMLElement | null>(null)
const showLeftArrow = ref(false)
const showRightArrow = ref(false)

// 检查滚动位置，显示/隐藏箭头
const checkScroll = () => {
  if (!scrollContainer.value) return
  
  const { scrollLeft, scrollWidth, clientWidth } = scrollContainer.value
  showLeftArrow.value = scrollLeft > 0
  showRightArrow.value = scrollLeft < scrollWidth - clientWidth - 10
}

// 向左滚动
const scrollLeft = () => {
  if (!scrollContainer.value) return
  scrollContainer.value.scrollBy({ left: -300, behavior: 'smooth' })
}

// 向右滚动
const scrollRight = () => {
  if (!scrollContainer.value) return
  scrollContainer.value.scrollBy({ left: 300, behavior: 'smooth' })
}

// 清空历史记录
const handleClearHistory = () => {
  if (confirm('Are you sure you want to clear your browsing history?')) {
    clearHistory()
  }
}

// 移除单个商品
const handleRemoveItem = (id: number) => {
  removeItem(id)
}

// 分享商品到聊天
const handleShareToChat = (event: Event, item: any) => {
  event.preventDefault() // 阻止链接跳转
  emit('share-to-chat', item)
}

// 监听滚动事件
onMounted(() => {
  if (scrollContainer.value) {
    scrollContainer.value.addEventListener('scroll', checkScroll)
    checkScroll()
  }
})

onUnmounted(() => {
  if (scrollContainer.value) {
    scrollContainer.value.removeEventListener('scroll', checkScroll)
  }
})
</script>

<style scoped>
/* 隐藏滚动条 */
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}

/* 文本截断 */
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
