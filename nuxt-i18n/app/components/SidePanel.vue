<template>
  <teleport to="body">
    <!-- 左侧面板 (Sidebar) -->
    <aside 
      class="fixed left-0 top-1/2 -translate-y-1/2 pointer-events-none z-[9999]"
      aria-label="Sidebar"
    >
      <section
        class="relative w-[78vw] max-md:w-[80vw] h-[95vh] max-md:h-[90vh] flex justify-center border-2 border-[#6b73ff] rounded-2xl bg-black shadow-[0_0_30px_rgba(107,115,255,0.3)] pointer-events-auto transition-transform duration-[280ms] ease-in-out"
        :class="{
          'translate-x-0': leftOpen,
          '-translate-x-full': !leftOpen,
          'translate-x-[-60vw]': leftOpen && rightOpen && rightPriority
        }"
      >
        <!-- 左侧关闭按钮 -->
        <button 
          class="absolute top-2 right-2 w-7 h-7 inline-flex items-center justify-center border border-[rgba(124,117,255,0.6)] rounded-md bg-[rgba(30,27,75,0.6)] text-[#e8e9ff] cursor-pointer hover:brightness-110 transition-all z-10 pointer-events-auto" 
          type="button" 
          @click="closeLeft" 
          aria-label="Close sidebar"
        >×</button>
        
        <!-- 左侧把手按钮 -->
        <button 
          class="w-[26px] h-[120px] rounded-r-[26px] box-border inline-flex items-center justify-center absolute -right-[26px] top-1/2 -translate-y-1/2 bg-gradient-to-br from-purple-500 to-indigo-500 border-2 border-[rgba(124,117,255,0.85)] shadow-[0_0_0_3px_rgba(124,117,255,0.18)] text-[#e8e9ff] cursor-pointer pointer-events-auto hover:brightness-110 hover:shadow-[0_0_0_4px_rgba(124,117,255,0.22),0_8px_22px_rgba(0,0,0,0.42)] focus-visible:brightness-110 focus-visible:shadow-[0_0_0_4px_rgba(124,117,255,0.22),0_8px_22px_rgba(0,0,0,0.42)] transition-all" 
          type="button" 
          @click="toggleLeft" 
          :aria-expanded="leftOpen"
        >
          <span class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-xs leading-none">{{ leftArrow }}</span>
        </button>
        
        <!-- 左侧内容 -->
        <div class="w-full h-full box-border m-0 relative overflow-y-auto bg-black p-4 rounded-2xl">
          <slot name="left" />
        </div>
      </section>
    </aside>

    <!-- 右侧面板 (RightBar/Member) -->
    <div 
      class="fixed right-0 top-1/2 -translate-y-1/2 w-[80vw] max-md:w-[80vw] h-[95vh] max-md:h-[90vh] bg-black border-2 border-[#6b73ff] rounded-2xl shadow-[0_0_30px_rgba(107,115,255,0.3)] overflow-y-auto flex flex-col items-stretch pointer-events-none transition-transform duration-[280ms] ease-in-out z-[9999]"
      :class="{
        'translate-x-0': rightOpen,
        'translate-x-full': !rightOpen,
        'translate-x-[60vw]': rightOpen && leftOpen && !rightPriority
      }"
    >
      <!-- 右侧关闭按钮 -->
      <button 
        class="absolute left-2 top-2 w-7 h-7 inline-flex items-center justify-center border border-[rgba(124,117,255,0.6)] rounded-md bg-[rgba(30,27,75,0.6)] text-[#e8e9ff] cursor-pointer p-0 transition-[filter,box-shadow] duration-200 z-[5] pointer-events-auto hover:brightness-[1.08] hover:shadow-[0_0_0_4px_rgba(124,117,255,0.22),0_8px_22px_rgba(0,0,0,0.42)]" 
        @click="closeRight"
      >×</button>
      
      <!-- 右侧把手按钮 -->
      <button 
        class="absolute left-0 top-1/2 -translate-y-1/2 w-[26px] h-[120px] max-md:h-[100px] rounded-r-[26px] box-border inline-flex items-center justify-center bg-gradient-to-br from-[#8b5cf6] to-[#6366f1] border-2 border-[rgba(124,117,255,0.85)] shadow-[0_0_0_3px_rgba(124,117,255,0.18)] text-[#e8e9ff] cursor-pointer p-0 z-[999] pointer-events-auto hover:brightness-110 transition-all" 
        type="button" 
        :aria-label="rightArrow === '◀' ? 'expand member center' : 'retract member center'" 
        @click="toggleRight"
      >
        <span class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-base leading-none">{{ rightArrow }}</span>
      </button>
      
      <!-- 右侧内容 -->
      <div class="w-full box-border m-0 relative px-5 py-4 max-md:px-[15px] max-md:py-[15px] overflow-y-auto h-full flex flex-col pb-5 pointer-events-auto max-md:w-[calc(90vw-32px)] max-md:max-w-[calc(90vw-32px)] max-md:max-h-[calc(80vh-32px)] bg-black rounded-2xl">
        <div class="flex-1 grid grid-cols-1 auto-rows-max gap-4 items-start">
          <slot name="right" />
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { ref, computed, provide } from 'vue'

const leftOpen = ref(false)
const rightOpen = ref(false)
const rightPriority = ref(false) // true = 右侧优先, false = 左侧优先

// 左侧箭头
const leftArrow = computed(() => {
  if (!leftOpen.value) return '▶'
  // 打开状态：如果右侧优先（左侧被压缩），显示向右箭头；否则显示向左箭头
  return rightPriority.value ? '▶' : '◀'
})

// 右侧箭头
const rightArrow = computed(() => {
  // 如果左侧优先（右侧被压缩），显示向左箭头；否则显示向右箭头
  return !rightPriority.value && leftOpen.value ? '◀' : '▶'
})

// 切换左侧面板
const toggleLeft = () => {
  if (!leftOpen.value) {
    // 打开左侧，设置左侧优先
    leftOpen.value = true
    rightPriority.value = false
    return
  }

  // 已打开：如果右侧优先，切换为左侧优先
  if (rightPriority.value && rightOpen.value) {
    rightPriority.value = false
    return
  }

  // 已打开且左侧优先
  if (!rightOpen.value) {
    // 如果右侧未打开，完全关闭左侧
    leftOpen.value = false
  } else {
    // 如果右侧打开，切换为右侧优先
    rightPriority.value = true
  }
}

// 切换右侧面板
const toggleRight = () => {
  if (!rightOpen.value) {
    // 打开右侧，设置右侧优先
    rightOpen.value = true
    rightPriority.value = true
    return
  }

  // 已打开：如果左侧优先，切换为右侧优先
  if (!rightPriority.value && leftOpen.value) {
    rightPriority.value = true
    return
  }

  // 已打开且右侧优先
  if (!leftOpen.value) {
    // 如果左侧未打开，完全关闭右侧
    rightOpen.value = false
  } else {
    // 如果左侧打开，切换为左侧优先
    rightPriority.value = false
  }
}

// 关闭左侧
const closeLeft = () => {
  leftOpen.value = false
}

// 关闭右侧
const closeRight = () => {
  rightOpen.value = false
}

// 暴露方法供外部调用
const openLeft = () => {
  leftOpen.value = true
  rightPriority.value = false
}

const openRight = () => {
  rightOpen.value = true
  rightPriority.value = true
}

// 提供给子组件使用
provide('sidePanel', {
  openLeft,
  openRight,
  closeLeft,
  closeRight,
  toggleLeft,
  toggleRight
})

// 暴露给父组件使用
defineExpose({
  openLeft,
  openRight,
  closeLeft,
  closeRight,
  toggleLeft,
  toggleRight
})
</script>
