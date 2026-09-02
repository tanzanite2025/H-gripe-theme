<template>
  <ClientOnly>
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
        class="fixed inset-0 bg-slate-900/20 backdrop-blur-sm z-[9998]"
        @click="handleBackdropClick"
      ></div>
    </transition>

    <!-- 左侧面板 (Sidebar) -->
    <aside
      class="fixed left-0 top-0 sidepanel-shell pointer-events-none z-[9999]"
      aria-label="Sidebar"
    >
      <section
        class="sidebar-panel relative w-[85vw] md:w-[45vw] h-full flex border tz-border-subtle rounded-none md:rounded-2xl shadow-md pointer-events-auto transition-transform duration-[280ms] ease-in-out"
        :class="{
          'translate-x-0': leftOpen,
          '-translate-x-full': !leftOpen
        }"
      >
        <!-- 左侧关闭按钮 -->
        <button
          class="tz-global-close-btn absolute top-2 right-2 z-10 pointer-events-auto"
          type="button"
          @click="closeLeft"
          aria-label="Close sidebar"
        >
          <span class="text-sm leading-none">×</span>
        </button>

        <!-- 左侧把手按钮 -->
        <button
          class="sidebar-handle sidebar-handle--left w-[26px] h-[120px] rounded-r-[26px] box-border inline-flex items-center justify-center absolute -right-[26px] top-1/2 -translate-y-1/2 bg-[#059669] border-2 border-[rgba(5, 150, 105,0.8)] shadow-[0_0_0_3px_rgba(5, 150, 105,0.16)] text-white cursor-pointer pointer-events-auto hover:bg-[#047857] hover:shadow-md focus-visible:bg-[#047857] focus-visible:shadow-md transition-all"
          type="button"
          @click="toggleLeft"
          :aria-expanded="leftOpen"
        >
          <span class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-xs leading-none">{{ leftArrow }}</span>
        </button>

        <!-- 左侧内容 -->
        <div class="w-full h-full box-border m-0 relative overflow-y-auto pt-10 px-2 pb-0 md:pt-12 md:px-4 md:pb-4 rounded-none md:rounded-2xl">
          <slot v-if="leftEverOpened" name="left" />
        </div>
      </section>
    </aside>
    </teleport>
  </ClientOnly>
</template>

<script setup lang="ts">
import { computed, provide } from 'vue'
import { useSidePanelState } from '~/composables/useSidePanelState'

const {
  leftOpen,
  leftEverOpened,
  openLeft,
  closeLeft,
  toggleLeft,
} = useSidePanelState()

// 左侧箭头：关闭时向右，打开时向左
const leftArrow = computed(() => (leftOpen.value ? '◀' : '▶'))

// 点击蒙版关闭侧边栏
const handleBackdropClick = () => {
  closeLeft()
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
</script>

<style>
/* 侧边栏面板背景 */
.sidebar-panel {
  background: var(--tz-card-surface);
}

 .sidepanel-shell {
   height: var(--tz-mobile-safe-viewport-height, 100vh);
 }

 @supports (height: 100svh) {
   .sidepanel-shell {
     height: var(--tz-mobile-safe-viewport-height, 100svh);
   }
 }

 @supports (height: 100dvh) {
   .sidepanel-shell {
     height: var(--tz-mobile-safe-viewport-height, 100dvh);
   }
 }

 @media (max-width: 767px) {
   .sidepanel-shell {
     top: var(--tz-safe-area-top, 0px);
   }

   .sidepanel-shell .sidebar-panel > div {
     padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 0.75rem);
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
