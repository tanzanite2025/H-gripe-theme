<template>
  <teleport to="body">
    <!-- 背景蒙版 -->
    <transition
      enter-active-class="transition-opacity duration-300 ease-out"
      leave-active-class="transition-opacity duration-200 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div 
        v-if="leftOpen"
        class="fixed inset-0 bg-black/80 backdrop-blur-sm z-[9998]"
        @click="handleBackdropClick"
      ></div>
    </transition>

    <!-- 左侧面板 (Sidebar) -->
    <aside 
      class="fixed left-0 top-0 sidepanel-shell pointer-events-none z-[9999]"
      aria-label="Sidebar"
    >
      <section
        class="sidebar-panel relative w-[85vw] md:w-[45vw] h-full flex border border-white/20 rounded-none md:rounded-2xl shadow-[0_18px_60px_-24px_rgba(0,0,0,1)] pointer-events-auto transition-transform duration-[280ms] ease-in-out"
        :class="{
          'translate-x-0': leftOpen,
          '-translate-x-full': !leftOpen
        }"
      >
        <!-- 左侧关闭按钮 -->
        <button 
          class="absolute top-2 right-2 w-8 h-8 inline-flex items-center justify-center rounded-xl bg-gradient-to-br from-teal-400 to-blue-500 text-black font-bold text-sm cursor-pointer hover:shadow-[0_4px_12px_rgba(45,212,191,0.4)] hover:scale-105 transition-all z-10 pointer-events-auto" 
          type="button" 
          @click="closeLeft" 
          aria-label="Close sidebar"
        >×</button>
        
        <!-- 左侧把手按钮 -->
        <button 
          class="sidebar-handle sidebar-handle--left w-[26px] h-[120px] rounded-r-[26px] box-border inline-flex items-center justify-center absolute -right-[26px] top-1/2 -translate-y-1/2 bg-gradient-to-br from-purple-500 to-indigo-500 border-2 border-[rgba(124,117,255,0.85)] shadow-[0_0_0_3px_rgba(124,117,255,0.18)] text-[#e8e9ff] cursor-pointer pointer-events-auto hover:brightness-110 hover:shadow-[0_0_0_4px_rgba(124,117,255,0.22),0_8px_22px_rgba(0,0,0,0.42)] focus-visible:brightness-110 focus-visible:shadow-[0_0_0_4px_rgba(124,117,255,0.22),0_8px_22px_rgba(0,0,0,0.42)] transition-all" 
          type="button" 
          @click="toggleLeft" 
          :aria-expanded="leftOpen"
        >
          <span class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-xs leading-none">{{ leftArrow }}</span>
        </button>
        
        <!-- 左侧内容 -->
        <div class="w-full h-full box-border m-0 relative overflow-y-auto pt-10 px-2 pb-0 md:pt-12 md:px-4 md:pb-4 rounded-none md:rounded-2xl">
          <slot name="left" />
        </div>
      </section>
    </aside>
  </teleport>
</template>

<script setup>
import { ref, computed, provide, onMounted, onBeforeUnmount } from 'vue'

// 左侧 Sidebar 打开状态
const leftOpen = ref(false)

// 左侧箭头：关闭时向右，打开时向左
const leftArrow = computed(() => (leftOpen.value ? '◀' : '▶'))

// 切换左侧面板
const toggleLeft = () => {
  leftOpen.value = !leftOpen.value
}

// 关闭左侧
const closeLeft = () => {
  leftOpen.value = false
}

// 点击蒙版关闭侧边栏
const handleBackdropClick = () => {
  leftOpen.value = false
}

// 暴露方法供外部调用
const openLeft = () => {
  leftOpen.value = true
}

// 提供给子组件使用（仅左侧）
provide('sidePanel', {
  openLeft,
  closeLeft,
  toggleLeft,
})

// 暴露给父组件使用（仅左侧）
defineExpose({
  openLeft,
  closeLeft,
  toggleLeft,
})

// 监听全局事件，允许外部通过 CustomEvent 打开左侧侧边栏
const handleGlobalSidebarEvent = (event) => {
  try {
    const detail = event && event.detail ? event.detail : {}
    const side = detail.side || 'left'
    if (side === 'left') {
      openLeft()
    }
  } catch {}
}

onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener('ui:sidebar-open', handleGlobalSidebarEvent)
  }
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('ui:sidebar-open', handleGlobalSidebarEvent)
  }
})
</script>

<style>
/* 侧边栏面板背景 - Style C 渐变 */
.sidebar-panel {
  background: radial-gradient(circle at 50% 0%, #0f172a 0%, #000000 100%);
}

 .sidepanel-shell {
   height: 100vh;
 }

 @supports (height: 100svh) {
   .sidepanel-shell {
     height: 100svh;
   }
 }

 @supports (height: 100dvh) {
   .sidepanel-shell {
     height: 100dvh;
   }
 }

body.hide-sidebar-handles .sidebar-handle {
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s ease;
}

/* 桌面端和移动端统一隐藏左右侧边栏句柄按钮，
   通过底部 Dock 按钮或其它入口来打开侧边栏。 */
.sidebar-handle {
  display: none;
}
</style>
