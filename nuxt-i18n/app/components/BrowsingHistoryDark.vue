<template>
  <div
    class="browsing-history-dark w-full rounded-2xl border backdrop-blur-md shadow-md overflow-hidden"
    :class="[
      density === 'cart'
        ? 'browsing-history-dark--cart tz-border-subtle tz-surface-subtle'
        : 'tz-border-subtle tz-surface-card',
      {
        'browsing-history-dark--cart-empty': density === 'cart' && !hasHistory,
      },
    ]"
  >
    <!-- 标题栏 -->
    <div class="browsing-history-dark__header flex items-center justify-between px-4 py-3 border-b tz-border-subtle tz-surface-subtle">
      <div class="flex items-center gap-2">
      <svg class="w-5 h-5 tz-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 class="text-sm font-semibold tz-text-primary">Recently Viewed</h3>
      <span class="text-xs tz-text-muted">({{ historyCount }})</span>
      </div>
      <button
        v-if="hasHistory"
        @click="handleClearHistory"
        class="text-xs tz-text-muted hover:text-red-400 transition-colors"
      >
        Clear History
      </button>
    </div>

    <!-- 商品列表 - 横向滚动 -->
    <div class="browsing-history-dark__body relative">
      <!-- 空状态 -->
      <div v-if="!hasHistory" class="browsing-history-dark__empty flex flex-col items-center justify-center py-8 px-4">
        <svg class="w-16 h-16 tz-text-primary/20 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p class="tz-text-secondary text-sm">No browsing history</p>
        <p class="tz-text-muted text-xs mt-1">Products you view will appear here</p>
      </div>
      
      <!-- 商品列表 -->
      <div
        v-else
        ref="scrollContainer"
        class="browsing-history-dark__list flex gap-3 p-4 overflow-x-auto scrollbar-hide scroll-smooth"
        style="scrollbar-width: none; -ms-overflow-style: none;"
      >
        <div
          v-for="item in history"
          :key="item.id"
          class="browsing-history-dark__item flex-shrink-0 w-40 group"
        >
          <!-- 商品卡片：暗色玻璃 + 纯黑阴影，无边框 -->
          <div
            class="browsing-history-dark__product-card relative rounded-[14px]
                   tz-surface-card
                   shadow-[0_4px_12px_-4px_rgba(15,23,42,0.12)] overflow-hidden
                   transition-all duration-200"
          >
            <!-- 删除按钮 -->
            <button
              @click="handleRemoveItem(item.id)"
               class="absolute top-1 right-1 z-10 w-5 h-5 tz-surface-muted hover:bg-red-100 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
              title="Remove"
            >
              <svg class="w-3 h-3 tz-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <!-- 商品图片 -->
            <div class="browsing-history-dark__image relative w-full h-32 tz-surface-subtle">
              <StorefrontImage
                v-if="item.thumbnail"
                :src="item.thumbnail"
                :alt="item.title"
                class="w-full h-full object-cover"
                preset="history"
              />
              <div v-else class="w-full h-full flex items-center justify-center tz-text-muted">
                <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
            </div>

            <!-- 商品信息 -->
            <div class="p-2">
              <h4 class="text-xs font-medium tz-text-primary line-clamp-2 mb-1">
                {{ item.title }}
              </h4>
              <p
                class="browsing-history-dark__price text-sm font-semibold mb-2"
                :class="density === 'cart' ? 'tz-text-secondary' : 'text-[#059669]'"
              >
                {{ item.price }}
              </p>
              
              <!-- 操作按钮 -->
              <div class="flex gap-1.5 items-center">
                <!-- 加入心愿单按钮 -->
                <button
                  @click="handleAddToWishlist(item)"
          class="w-8 h-8 flex items-center justify-center rounded-full border tz-border-strong/25 tz-text-secondary hover:tz-surface-subtle transition-colors"
                  title="Add to wishlist"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z" />
                  </svg>
                </button>

                <!-- 查看详情按钮 -->
                <NuxtLink
                  :to="item.url"
                  class="flex-1 px-2 py-1.5 tz-surface-subtle hover:tz-surface-subtle border tz-border-subtle hover:tz-border-strong/40 rounded text-xs tz-text-primary text-center transition-all"
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
                  class="browsing-history-dark__share-btn flex-1 px-2 py-1.5 rounded text-xs tz-text-primary transition-all"
                  :class="density === 'cart'
                    ? 'border tz-border-subtle tz-surface-subtle hover:tz-border-strong/40 hover:tz-surface-subtle'
                    : 'bg-emerald-600 text-white hover:bg-emerald-700 shadow-lg'"
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
        class="hidden md:flex absolute left-2 top-1/2 -translate-y-1/2 w-8 h-8 tz-surface-subtle hover:tz-surface-subtle backdrop-blur-sm shadow-lg rounded-full items-center justify-center z-10 border tz-border-subtle"
      >
        <svg class="w-4 h-4 tz-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <button
        v-if="hasHistory && showRightArrow"
        @click="scrollRight"
        class="hidden md:flex absolute right-2 top-1/2 -translate-y-1/2 w-8 h-8 tz-surface-subtle hover:tz-surface-subtle backdrop-blur-sm shadow-lg rounded-full items-center justify-center z-10 border tz-border-subtle"
      >
        <svg class="w-4 h-4 tz-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <!-- 移动端滑动提示 -->
    <div v-if="hasHistory" class="browsing-history-dark__mobile-hint md:hidden px-4 pb-3 text-center">
      <p class="text-xs tz-text-muted">← Swipe to see more →</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useThrottleFn } from '@vueuse/core'
import { useBrowsingHistory } from '~/composables/useBrowsingHistory'
import { useWishlist } from '~/composables/useWishlist'

// 定义 emit 事件
const emit = defineEmits<{
  'share-to-chat': [product: any]
}>()

withDefaults(defineProps<{
  density?: 'default' | 'cart'
}>(), {
  density: 'default',
})

const { history, historyCount, hasHistory, clearHistory, removeItem } = useBrowsingHistory()
const { addToWishlist } = useWishlist()

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

const throttledCheckScroll = useThrottleFn(checkScroll, 150)

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

// 加入心愿单
const handleAddToWishlist = async (item: any) => {
  if (!item || !item.id) return
  try {
    await addToWishlist(item.id)
  } catch (e) {
    console.error('Failed to add to wishlist from history:', e)
  }
}

// 分享商品到聊天
const handleShareToChat = (event: Event, item: any) => {
  event.preventDefault() // 阻止链接跳转
  emit('share-to-chat', item)
}

// 监听滚动事件
onMounted(() => {
  if (scrollContainer.value) {
    scrollContainer.value.addEventListener('scroll', throttledCheckScroll)
    checkScroll()
  }
})

onUnmounted(() => {
  if (scrollContainer.value) {
    scrollContainer.value.removeEventListener('scroll', throttledCheckScroll)
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

.browsing-history-dark--cart {
  display: flex;
  height: auto;
  min-height: 0;
  flex-direction: column;
  border-color: var(--tz-border-subtle) !important;
  background: var(--tz-surface-subtle) !important;
  background-image: none !important;
}

.browsing-history-dark--cart .browsing-history-dark__header {
  flex: 0 0 auto;
  padding-block: 0.55rem;
  background: var(--tz-surface-muted) !important;
}

.browsing-history-dark--cart .browsing-history-dark__body {
  flex: 0 0 auto;
  min-height: 0;
  overflow: visible;
}

.browsing-history-dark--cart .browsing-history-dark__list {
  height: auto;
  min-height: 0;
  padding-block: 0.65rem;
  overflow-x: auto;
  overflow-y: visible;
  align-items: flex-start;
}

.browsing-history-dark--cart .browsing-history-dark__item {
  width: 8.5rem;
}

.browsing-history-dark--cart .browsing-history-dark__image {
  height: 5.75rem;
}

.browsing-history-dark--cart .browsing-history-dark__empty {
  height: auto;
  min-height: 0;
  padding-block: 0.8rem;
}

.browsing-history-dark--cart .browsing-history-dark__empty svg {
  width: 2.25rem;
  height: 2.25rem;
  margin-bottom: 0.35rem;
}

.browsing-history-dark--cart-empty {
  height: auto;
  min-height: 0;
}

.browsing-history-dark--cart-empty .browsing-history-dark__body {
  flex: 0 0 auto;
}

.browsing-history-dark--cart .browsing-history-dark__product-card {
  background: var(--tz-surface-card) !important;
  background-image: none !important;
  box-shadow: 0 4px 12px -4px rgba(15, 23, 42, 0.12);
}

.browsing-history-dark--cart .browsing-history-dark__price {
  color: var(--tz-text-primary) !important;
}

.browsing-history-dark--cart .browsing-history-dark__share-btn {
  border: 1px solid var(--tz-border-strong);
  background: var(--tz-surface-muted) !important;
  background-image: none !important;
  color: var(--tz-text-primary) !important;
  box-shadow: none !important;
}

.browsing-history-dark--cart .browsing-history-dark__share-btn:hover {
  border-color: var(--tz-border-strong);
  background: var(--tz-surface-subtle) !important;
}

.browsing-history-dark--cart .browsing-history-dark__mobile-hint {
  flex: 0 0 auto;
  padding-bottom: 0.55rem;
}

@media (min-width: 768px) {
  .browsing-history-dark--cart .browsing-history-dark__item {
    width: 9rem;
  }

  .browsing-history-dark--cart .browsing-history-dark__image {
    height: 6rem;
  }
}

@media (max-width: 767px) {
  .browsing-history-dark--cart .browsing-history-dark__image {
    height: 4.25rem;
  }

  .browsing-history-dark--cart .browsing-history-dark__product-card > .p-2 {
    padding: 0.45rem;
  }

  .browsing-history-dark--cart .browsing-history-dark__product-card .w-8.h-8 {
    width: 1.75rem;
    height: 1.75rem;
  }
}
</style>
