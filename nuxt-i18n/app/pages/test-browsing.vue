<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-6xl mx-auto px-4">
      <!-- 页面标题 -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-2">浏览历史组件测试</h1>
        <p class="text-gray-600">测试浏览记录的添加、显示和删除功能</p>
      </div>

      <!-- 测试控制面板 -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">测试控制</h2>
        
        <div class="space-y-4">
          <!-- 添加测试商品 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              添加测试商品到浏览历史
            </label>
            <div class="flex gap-2">
              <button
                @click="addTestProduct(1)"
                class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm"
              >
                添加商品 1
              </button>
              <button
                @click="addTestProduct(2)"
                class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm"
              >
                添加商品 2
              </button>
              <button
                @click="addTestProduct(3)"
                class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors text-sm"
              >
                添加商品 3
              </button>
              <button
                @click="addMultipleProducts"
                class="px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors text-sm"
              >
                批量添加 10 个
              </button>
            </div>
          </div>

          <!-- 统计信息 -->
          <div class="flex gap-4 text-sm">
            <div class="px-4 py-2 bg-gray-100 rounded-lg">
              <span class="text-gray-600">历史记录数量：</span>
              <span class="font-semibold text-gray-900">{{ historyCount }}</span>
            </div>
            <div class="px-4 py-2 bg-gray-100 rounded-lg">
              <span class="text-gray-600">是否有记录：</span>
              <span class="font-semibold text-gray-900">{{ hasHistory ? '是' : '否' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态提示 -->
      <div v-if="!hasHistory" class="bg-yellow-50 border border-yellow-200 rounded-lg p-6 text-center">
        <svg class="w-12 h-12 text-yellow-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p class="text-yellow-800 font-medium mb-2">暂无浏览历史</p>
        <p class="text-yellow-600 text-sm">点击上方按钮添加测试商品</p>
      </div>

      <!-- 使用说明 -->
      <div class="bg-blue-50 border border-blue-200 rounded-lg p-6 mt-6">
        <h3 class="text-lg font-semibold text-blue-900 mb-3">使用说明</h3>
        <ul class="space-y-2 text-sm text-blue-800">
          <li class="flex items-start gap-2">
            <svg class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>点击上方按钮添加测试商品到浏览历史</span>
          </li>
          <li class="flex items-start gap-2">
            <svg class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>鼠标悬停在商品卡片上可以看到删除按钮</span>
          </li>
          <li class="flex items-start gap-2">
            <svg class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>点击"清空历史"可以清除所有浏览记录</span>
          </li>
          <li class="flex items-start gap-2">
            <svg class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>桌面端可以使用左右箭头按钮滚动，移动端可以左右滑动</span>
          </li>
          <li class="flex items-start gap-2">
            <svg class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>数据保存在 localStorage 中，刷新页面不会丢失</span>
          </li>
        </ul>
      </div>

      <!-- 返回按钮 -->
      <div class="mt-8 text-center">
        <NuxtLink
          to="/"
          class="inline-flex items-center gap-2 px-6 py-3 bg-gray-800 text-white rounded-lg hover:bg-gray-900 transition-colors"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          返回首页
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useBrowsingHistory } from '~/composables/useBrowsingHistory'

const { historyCount, hasHistory, addToHistory } = useBrowsingHistory()

// 测试商品数据
const testProducts = [
  {
    id: 1,
    title: 'iPhone 15 Pro Max 256GB 深空黑色',
    thumbnail: 'https://images.unsplash.com/photo-1592286927505-2fd0f3a3b8d4?w=400',
    price: '$1,199.00',
    url: '/product/1'
  },
  {
    id: 2,
    title: 'MacBook Pro 14" M3 芯片 16GB 512GB',
    thumbnail: 'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=400',
    price: '$1,999.00',
    url: '/product/2'
  },
  {
    id: 3,
    title: 'AirPods Pro 第二代 主动降噪无线耳机',
    thumbnail: 'https://images.unsplash.com/photo-1606841837239-c5a1a4a07af7?w=400',
    price: '$249.00',
    url: '/product/3'
  },
  {
    id: 4,
    title: 'Apple Watch Series 9 GPS 45mm',
    thumbnail: 'https://images.unsplash.com/photo-1579586337278-3befd40fd17a?w=400',
    price: '$429.00',
    url: '/product/4'
  },
  {
    id: 5,
    title: 'iPad Air 第五代 10.9英寸 256GB',
    thumbnail: 'https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?w=400',
    price: '$749.00',
    url: '/product/5'
  },
  {
    id: 6,
    title: 'Magic Keyboard 妙控键盘 触控板版',
    thumbnail: 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=400',
    price: '$149.00',
    url: '/product/6'
  },
  {
    id: 7,
    title: 'HomePod mini 智能音箱 白色',
    thumbnail: 'https://images.unsplash.com/photo-1589003077984-894e133dabab?w=400',
    price: '$99.00',
    url: '/product/7'
  },
  {
    id: 8,
    title: 'AirTag 4件装 防丢追踪器',
    thumbnail: 'https://images.unsplash.com/photo-1621768216002-5ac171876625?w=400',
    price: '$99.00',
    url: '/product/8'
  },
  {
    id: 9,
    title: 'MagSafe 充电器 15W 无线快充',
    thumbnail: 'https://images.unsplash.com/photo-1591290619762-c588f0e8e23f?w=400',
    price: '$39.00',
    url: '/product/9'
  },
  {
    id: 10,
    title: 'Apple Pencil 第二代 触控笔',
    thumbnail: 'https://images.unsplash.com/photo-1611532736597-de2d4265fba3?w=400',
    price: '$129.00',
    url: '/product/10'
  }
]

// 添加单个测试商品
const addTestProduct = (index: number) => {
  const product = testProducts[index - 1]
  if (product) {
    addToHistory(product)
  }
}

// 批量添加测试商品
const addMultipleProducts = () => {
  testProducts.forEach((product, index) => {
    setTimeout(() => {
      addToHistory(product)
    }, index * 100) // 每个商品间隔100ms添加，模拟真实浏览
  })
}
</script>
