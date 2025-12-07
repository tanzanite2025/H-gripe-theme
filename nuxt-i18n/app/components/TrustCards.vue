<template>
  <div>
    <!-- 标题（可选） -->
    <div v-if="showTitle" class="text-white/50 text-[10px] uppercase tracking-wider mb-2 px-1">
      {{ title }}
    </div>
    
    <!-- 卡片网格 -->
    <div 
      class="grid gap-2"
      :class="layoutClass"
    >
      <NuxtLink
        v-for="card in cards"
        :key="card.id"
        :to="localePath(card.to)"
        class="group relative flex items-center gap-3 border rounded-xl transition-all hover:-translate-y-0.5 overflow-hidden"
        :class="[sizeClass, cardColorClass(card.color)]"
      >
        <!-- 背景图片 -->
        <div 
          class="absolute inset-0 bg-contain bg-right bg-no-repeat transition-opacity"
          :style="{ backgroundImage: `url(${card.image})` }"
        ></div>
        
        <!-- 文字 -->
        <div v-if="layout === 'row' || size === 'lg'" class="relative z-10 flex flex-col">
          <span class="text-white font-medium" :class="titleSizeClass">{{ card.label }}</span>
          <span v-if="card.desc && size === 'lg'" class="text-white/70 text-xs">{{ card.desc }}</span>
        </div>
        <span v-else class="relative z-10 text-white text-sm font-semibold">{{ card.label }}</span>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useLocalePath } from '#imports'

const localePath = useLocalePath()

// Props
const props = withDefaults(defineProps<{
  layout?: 'grid' | 'row'
  size?: 'sm' | 'md' | 'lg'
  showTitle?: boolean
  title?: string
}>(), {
  layout: 'grid',
  size: 'sm',
  showTitle: true,
  title: 'Quick Access'
})

// 图标组件（需要在 cards 之前定义）
const PaymentIcon = {
  render() {
    return h('svg', { 
      fill: 'none', 
      stroke: 'currentColor', 
      viewBox: '0 0 24 24',
      class: 'w-full h-full'
    }, [
      h('path', { 
        'stroke-linecap': 'round', 
        'stroke-linejoin': 'round', 
        'stroke-width': '1.5',
        d: 'M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z'
      })
    ])
  }
}

const ShippingIcon = {
  render() {
    return h('svg', { 
      fill: 'none', 
      stroke: 'currentColor', 
      viewBox: '0 0 24 24',
      class: 'w-full h-full'
    }, [
      h('path', { 
        'stroke-linecap': 'round', 
        'stroke-linejoin': 'round', 
        'stroke-width': '1.5',
        d: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4'
      })
    ])
  }
}

const WarrantyIcon = {
  render() {
    return h('svg', { 
      fill: 'none', 
      stroke: 'currentColor', 
      viewBox: '0 0 24 24',
      class: 'w-full h-full'
    }, [
      h('path', { 
        'stroke-linecap': 'round', 
        'stroke-linejoin': 'round', 
        'stroke-width': '1.5',
        d: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z'
      })
    ])
  }
}

const AfterSalesIcon = {
  render() {
    return h('svg', { 
      fill: 'none', 
      stroke: 'currentColor', 
      viewBox: '0 0 24 24',
      class: 'w-full h-full'
    }, [
      h('path', { 
        'stroke-linecap': 'round', 
        'stroke-linejoin': 'round', 
        'stroke-width': '1.5',
        d: 'M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z'
      })
    ])
  }
}

// 卡片数据
const cards = [
  {
    id: 'payment',
    label: 'Payment',
    desc: 'Secure payment methods',
    to: '/support/payment',
    color: 'purple',
    icon: PaymentIcon,
    image: '/images/trust/payment.webp'
  },
  {
    id: 'shipping',
    label: 'Shipping',
    desc: 'Worldwide delivery',
    to: '/support/shipping',
    color: 'cyan',
    icon: ShippingIcon,
    image: '/images/trust/shipping.webp'
  },
  {
    id: 'warranty',
    label: 'Warranty',
    desc: 'Quality guarantee',
    to: '/support/warranty-check',
    color: 'violet',
    icon: WarrantyIcon,
    image: '/images/trust/warranty.webp'
  },
  {
    id: 'after-sales',
    label: 'After Sales',
    desc: 'Customer support',
    to: '/support/after-sales',
    color: 'indigo',
    icon: AfterSalesIcon,
    image: '/images/trust/after-sales.webp'
  }
]

// 布局类
const layoutClass = computed(() => {
  if (props.layout === 'row') {
    return 'grid-cols-2 lg:grid-cols-4'
  }
  return 'grid-cols-2'
})

// 尺寸类
const sizeClass = computed(() => {
  switch (props.size) {
    case 'lg':
      return 'p-4'
    case 'md':
      return 'p-3'
    default:
      return props.layout === 'grid' ? 'flex-col p-3' : 'p-2.5'
  }
})

// 图标容器尺寸
const iconSizeClass = computed(() => {
  switch (props.size) {
    case 'lg':
      return 'w-12 h-12'
    case 'md':
      return 'w-10 h-10'
    default:
      return 'w-10 h-10'
  }
})

// 标题尺寸
const titleSizeClass = computed(() => {
  switch (props.size) {
    case 'lg':
      return 'text-base'
    default:
      return 'text-sm'
  }
})

// 颜色相关类
const cardColorClass = (color: string) => {
  const colors: Record<string, string> = {
    purple: 'bg-transparent border-cyan-500 shadow-[0_0_8px_rgba(6,182,212,0.4)] hover:border-cyan-400 hover:shadow-[0_0_12px_rgba(6,182,212,0.6)]',
    cyan: 'bg-transparent border-cyan-500 shadow-[0_0_8px_rgba(6,182,212,0.4)] hover:border-cyan-400 hover:shadow-[0_0_12px_rgba(6,182,212,0.6)]',
    violet: 'bg-transparent border-cyan-500 shadow-[0_0_8px_rgba(6,182,212,0.4)] hover:border-cyan-400 hover:shadow-[0_0_12px_rgba(6,182,212,0.6)]',
    indigo: 'bg-transparent border-cyan-500 shadow-[0_0_8px_rgba(6,182,212,0.4)] hover:border-cyan-400 hover:shadow-[0_0_12px_rgba(6,182,212,0.6)]'
  }
  return colors[color] || colors.purple
}

const iconColorClass = (color: string) => {
  const colors: Record<string, string> = {
    purple: 'bg-purple-500/20 group-hover:bg-purple-500/30',
    cyan: 'bg-cyan-500/20 group-hover:bg-cyan-500/30',
    violet: 'bg-violet-500/20 group-hover:bg-violet-500/30',
    indigo: 'bg-indigo-500/20 group-hover:bg-indigo-500/30'
  }
  return colors[color] || colors.purple
}

const iconInnerClass = (color: string) => {
  const size = props.size === 'lg' ? 'w-6 h-6' : 'w-5 h-5'
  const colors: Record<string, string> = {
    purple: `${size} text-purple-400`,
    cyan: `${size} text-cyan-400`,
    violet: `${size} text-violet-400`,
    indigo: `${size} text-indigo-400`
  }
  return colors[color] || colors.purple
}
</script>
