/**
 * 包装计算 Composable
 * 用于计算商品的包装重量和体积
 */

import { ref } from 'vue'
import type { CartItem } from '~/types/cart'

/**
 * 包装规则接口
 */
export interface PackagingRule {
  id: number
  rule_name: string
  description?: string
  box_weight: number
  box_length?: number | null
  box_width?: number | null
  box_height?: number | null
  max_items?: number | null
  max_weight?: number | null
  priority: number
  is_active: boolean
  applies_to: Array<{
    type: 'category' | 'tag' | 'product' | 'all'
    value: string | null
  }>
}

/**
 * 包裹信息
 */
export interface PackageInfo {
  items: CartItem[]
  productWeight: number
  boxWeight: number
  totalWeight: number
  rule: PackagingRule | null
}

/**
 * 包装计算结果
 */
export interface PackagingResult {
  packages: PackageInfo[]
  totalProductWeight: number
  totalBoxWeight: number
  totalShippingWeight: number
  packageCount: number
}

/**
 * 包装计算 Composable
 */
export function usePackagingCalculation() {
  const packagingRules = ref<PackagingRule[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  /**
   * 加载包装规则
   */
  async function loadPackagingRules(): Promise<void> {
    if (packagingRules.value.length > 0) {
      return // 已加载，不重复请求
    }

    isLoading.value = true
    error.value = null

    try {
      const config = useRuntimeConfig()
      const base = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
      const response = await $fetch<{ data: PackagingRule[] }>(
        `${base}/shipping/packaging-rules`
      )

      if (Array.isArray(response?.data)) {
        packagingRules.value = response.data.filter(r => r.is_active)
      }
    } catch (e) {
      console.error('Failed to load packaging rules:', e)
      error.value = 'Failed to load packaging information'
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 查找适用于商品的包装规则
   */
  function findMatchingRule(item: CartItem): PackagingRule | null {
    // 按优先级排序（高优先级在前）
    const sortedRules = [...packagingRules.value].sort((a, b) => b.priority - a.priority)

    for (const rule of sortedRules) {
      if (!rule.is_active) continue

      for (const apply of rule.applies_to) {
        // 匹配所有商品
        if (apply.type === 'all') {
          return rule
        }

        // 匹配商品ID
        if (apply.type === 'product' && apply.value === String(item.id)) {
          return rule
        }

        // 匹配分类
        if (apply.type === 'category' && item.category && apply.value === item.category) {
          return rule
        }

        // 匹配标签
        if (apply.type === 'tag' && item.tags && apply.value && item.tags.includes(apply.value)) {
          return rule
        }
      }
    }

    return null
  }

  /**
   * 计算商品的包装
   */
  function calculatePackaging(items: CartItem[]): PackagingResult {
    const packages: PackageInfo[] = []
    let totalProductWeight = 0
    let totalBoxWeight = 0

    // 按包装规则分组
    const groupedItems = new Map<number | 'default', { rule: PackagingRule | null; items: CartItem[] }>()

    for (const item of items) {
      const rule = findMatchingRule(item)
      const key = rule ? rule.id : 'default'

      if (!groupedItems.has(key)) {
        groupedItems.set(key, { rule, items: [] })
      }

      // 展开数量
      for (let i = 0; i < item.quantity; i++) {
        groupedItems.get(key)!.items.push({ ...item, quantity: 1 })
      }
    }

    // 为每组创建包裹
    for (const [_, group] of groupedItems) {
      const { rule, items: groupItems } = group

      if (groupItems.length === 0) continue

      // 如果没有规则，每个商品单独一个包裹
      if (!rule) {
        for (const item of groupItems) {
          const productWeight = item.weight || 0
          totalProductWeight += productWeight

          packages.push({
            items: [item],
            productWeight,
            boxWeight: 0, // 没有规则时不计算包装重量
            totalWeight: productWeight,
            rule: null
          })
        }
        continue
      }

      // 根据规则限制分包
      let currentPackage: CartItem[] = []
      let currentWeight = 0
      let currentCount = 0

      for (const item of groupItems) {
        const itemWeight = item.weight || 0
        const wouldExceedItems = rule.max_items && currentCount + 1 > rule.max_items
        const wouldExceedWeight = rule.max_weight && currentWeight + itemWeight > rule.max_weight

        // 需要开新包裹
        if (currentPackage.length > 0 && (wouldExceedItems || wouldExceedWeight)) {
          const productWeight = currentWeight
          totalProductWeight += productWeight
          totalBoxWeight += rule.box_weight

          packages.push({
            items: [...currentPackage],
            productWeight,
            boxWeight: rule.box_weight,
            totalWeight: productWeight + rule.box_weight,
            rule
          })

          currentPackage = []
          currentWeight = 0
          currentCount = 0
        }

        currentPackage.push(item)
        currentWeight += itemWeight
        currentCount++
      }

      // 处理最后一个包裹
      if (currentPackage.length > 0) {
        const productWeight = currentWeight
        totalProductWeight += productWeight
        totalBoxWeight += rule.box_weight

        packages.push({
          items: [...currentPackage],
          productWeight,
          boxWeight: rule.box_weight,
          totalWeight: productWeight + rule.box_weight,
          rule
        })
      }
    }

    return {
      packages,
      totalProductWeight,
      totalBoxWeight,
      totalShippingWeight: totalProductWeight + totalBoxWeight,
      packageCount: packages.length
    }
  }

  /**
   * 获取默认包装规则（适用于所有商品的规则）
   */
  function getDefaultRule(): PackagingRule | null {
    return packagingRules.value.find(rule => 
      rule.is_active && rule.applies_to.some(a => a.type === 'all')
    ) || null
  }

  return {
    // State
    packagingRules,
    isLoading,
    error,

    // Methods
    loadPackagingRules,
    findMatchingRule,
    calculatePackaging,
    getDefaultRule
  }
}
