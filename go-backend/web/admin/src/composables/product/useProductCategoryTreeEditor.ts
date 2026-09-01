import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import productCategoryApi from '@/api/productCategories'
import {
  flattenProductCategoryTree,
  type DraftCategoryRow,
  type ProductCategoryParentOption,
  type ProductCategoryStats,
} from '@/modules/product/productCategoryTypes'

export const rootProductCategoryParentValue = '__root__'

const slugify = (value: string): string => String(value || '')
  .trim()
  .toLowerCase()
  .replace(/[^a-z0-9_-]+/g, '-')
  .replace(/^-+|-+$/g, '')
  .replace(/-{2,}/g, '-')

export const useProductCategoryTreeEditor = () => {
  const loading = ref(false)
  const savingAll = ref(false)
  const rows = ref<DraftCategoryRow[]>([])
  const maxDepth = ref(5)
  const nextDraftID = ref(1)

  const rowMap = computed(() => new Map(rows.value.map((row) => [row.key, row])))
  const changedCount = computed(() => rows.value.filter((row) => row.is_new || row.dirty).length)
  const hasChanges = computed(() => changedCount.value > 0)
  const stats = computed<ProductCategoryStats>(() => ({
    total: rows.value.length,
    enabled: rows.value.filter((row) => row.is_enabled).length,
    changed: changedCount.value,
    deepestLevel: rows.value.reduce((max, row) => Math.max(max, row.depth), 0),
    maxDepth: maxDepth.value,
    translationIncomplete: rows.value.filter((row) => (
      row.translation_total > 0
      && row.translation_completed < row.translation_total
    )).length,
  }))

  const markDirty = (row: DraftCategoryRow) => {
    if (!row.is_new) row.dirty = true
  }

  const normalizeSlug = (row: DraftCategoryRow) => {
    const next = slugify(row.slug)
    if (next === row.slug) return
    row.slug = next
    markDirty(row)
  }

  const childrenOf = (parentKey: string): DraftCategoryRow[] => rows.value.filter((row) => row.parent_key === parentKey)

  const subtreeKeys = (rootKey: string): Set<string> => {
    const keys = new Set<string>()
    const walk = (key: string) => {
      if (keys.has(key)) return
      keys.add(key)
      childrenOf(key).forEach((child) => walk(child.key))
    }
    walk(rootKey)
    return keys
  }

  const subtreeHeight = (rootKey: string): number => {
    const root = rowMap.value.get(rootKey)
    if (!root) return 1
    let deepest = root.depth
    subtreeKeys(rootKey).forEach((key) => {
      const row = rowMap.value.get(key)
      if (row) deepest = Math.max(deepest, row.depth)
    })
    return Math.max(1, deepest - root.depth + 1)
  }

  const lastSubtreeIndex = (rootKey: string): number => {
    const keys = subtreeKeys(rootKey)
    let index = -1
    rows.value.forEach((row, rowIndex) => {
      if (keys.has(row.key)) index = rowIndex
    })
    return index
  }

  const assignDepths = (rootKey: string, depth: number) => {
    const row = rowMap.value.get(rootKey)
    if (!row) return
    row.depth = depth
    childrenOf(rootKey).forEach((child) => assignDepths(child.key, depth + 1))
  }

  const parentOptions = (row: DraftCategoryRow): ProductCategoryParentOption[] => {
    const blocked = subtreeKeys(row.key)
    const height = subtreeHeight(row.key)
    return rows.value
      .filter((candidate) => candidate.key !== row.key)
      .map((candidate) => ({
        ...candidate,
        disabled: blocked.has(candidate.key) || candidate.depth + height > maxDepth.value,
      }))
  }

  const createDraftRow = (parentKey: string | null): DraftCategoryRow | null => {
    const parent = parentKey ? rowMap.value.get(parentKey) : null
    const depth = parent ? parent.depth + 1 : 1
    if (depth > maxDepth.value) {
      toast.error(`分类最多只能到第 ${maxDepth.value} 级`)
      return null
    }
    return {
      key: `new:${nextDraftID.value++}`,
      id: null,
      parent_key: parentKey,
      name: '',
      slug: '',
      description: '',
      image_media_asset_id: null,
      image_url: '',
      depth,
      sort_order: 0,
      is_enabled: true,
      translation_completed: 0,
      translation_total: 0,
      translation_missing_locales: [],
      is_new: true,
      dirty: true,
    }
  }

  const addRootCategory = () => {
    const row = createDraftRow(null)
    if (row) rows.value.push(row)
  }

  const addSibling = (current: DraftCategoryRow) => {
    const row = createDraftRow(current.parent_key)
    if (!row) return
    const index = lastSubtreeIndex(current.key)
    rows.value.splice(index + 1, 0, row)
  }

  const addChild = (current: DraftCategoryRow) => {
    const row = createDraftRow(current.key)
    if (!row) return
    const index = lastSubtreeIndex(current.key)
    rows.value.splice(index + 1, 0, row)
  }

  const changeParent = (row: DraftCategoryRow, rawValue: string) => {
    const nextParentKey = rawValue === rootProductCategoryParentValue ? null : rawValue
    if (nextParentKey === row.parent_key) return
    if (nextParentKey && !rowMap.value.has(nextParentKey)) return

    const blocked = subtreeKeys(row.key)
    if (nextParentKey && blocked.has(nextParentKey)) {
      toast.error('父级不能选自己或自己的下级')
      return
    }

    const parent = nextParentKey ? rowMap.value.get(nextParentKey) : null
    const newDepth = parent ? parent.depth + 1 : 1
    if (newDepth + subtreeHeight(row.key) - 1 > maxDepth.value) {
      toast.error(`移动后会超过 ${maxDepth.value} 级`)
      return
    }

    const movingKeys = subtreeKeys(row.key)
    const movingRows = rows.value.filter((item) => movingKeys.has(item.key))
    rows.value = rows.value.filter((item) => !movingKeys.has(item.key))

    row.parent_key = nextParentKey
    markDirty(row)

    const insertIndex = nextParentKey ? lastSubtreeIndex(nextParentKey) + 1 : rows.value.length
    rows.value.splice(insertIndex, 0, ...movingRows)
    assignDepths(row.key, newDepth)
  }

  const removeUnsavedSubtree = (row: DraftCategoryRow): boolean => {
    const keys = subtreeKeys(row.key)
    if (keys.size <= 1 || !rows.value.filter((item) => keys.has(item.key)).every((item) => item.is_new)) {
      return false
    }
    rows.value = rows.value.filter((item) => !keys.has(item.key))
    toast.success('未保存的分类分支已移除')
    return true
  }

  const removeUnsavedRow = (row: DraftCategoryRow): boolean => {
    if (!row.is_new) return false
    rows.value = rows.value.filter((item) => item.key !== row.key)
    return true
  }

  const validateRows = (): boolean => {
    const slugPattern = /^[a-z0-9]+(?:[_-][a-z0-9]+)*$/
    const seenSlugs = new Set<string>()
    for (const row of rows.value) {
      row.name = row.name.trim()
      if (!row.slug.trim()) row.slug = slugify(row.name)
      else row.slug = slugify(row.slug)
      row.description = row.description.trim()
      if (!row.name) {
        toast.error('有分类缺少名称')
        return false
      }
      if (!slugPattern.test(row.slug)) {
        toast.error(`分类“${row.name}”的 slug 无效`)
        return false
      }
      if (seenSlugs.has(row.slug)) {
        toast.error(`slug “${row.slug}” 重复`)
        return false
      }
      seenSlugs.add(row.slug)
      if (row.depth < 1 || row.depth > maxDepth.value) {
        toast.error(`分类“${row.name}”超过 ${maxDepth.value} 级限制`)
        return false
      }
    }
    return true
  }

  const assignAutomaticSortOrders = () => {
    const siblingIndexByParent = new Map<string, number>()
    rows.value.forEach((row) => {
      const parentKey = row.parent_key || rootProductCategoryParentValue
      const siblingIndex = siblingIndexByParent.get(parentKey) || 0
      siblingIndexByParent.set(parentKey, siblingIndex + 1)
      const nextSortOrder = siblingIndex * 10
      if (row.sort_order === nextSortOrder) return
      row.sort_order = nextSortOrder
      markDirty(row)
    })
  }

  const payloadFor = (row: DraftCategoryRow, keyToID: Map<string, number>) => {
    let parentID: number | null = null
    if (row.parent_key) {
      const resolvedParentID = keyToID.get(row.parent_key)
      if (!resolvedParentID) throw new Error(`missing parent id for ${row.name}`)
      parentID = resolvedParentID
    }
    return {
      parent_id: parentID,
      name: row.name.trim(),
      slug: row.slug.trim().toLowerCase(),
      description: row.description.trim(),
      image_media_asset_id: row.image_media_asset_id,
      sort_order: Number(row.sort_order || 0),
      is_enabled: row.is_enabled !== false,
    }
  }

  const fetchCategories = async () => {
    loading.value = true
    try {
      const payload = await productCategoryApi.list({ include_disabled: true })
      rows.value = flattenProductCategoryTree(payload.tree)
      maxDepth.value = Number(payload.max_depth || 5)
    } catch (error) {
      console.error('Failed to fetch product categories:', error)
      toast.error('商品分类加载失败')
    } finally {
      loading.value = false
    }
  }

  const saveAll = async () => {
    if (!hasChanges.value) return
    assignAutomaticSortOrders()
    if (!validateRows()) return

    savingAll.value = true
    try {
      const ordered = [...rows.value].sort((a, b) => a.depth - b.depth)
      const keyToID = new Map<string, number>()
      rows.value.forEach((row) => {
        if (row.id) keyToID.set(row.key, row.id)
      })

      for (const row of ordered) {
        if (!row.is_new) continue
        const saved = await productCategoryApi.create(payloadFor(row, keyToID))
        keyToID.set(row.key, saved.id)
      }

      for (const row of ordered) {
        if (row.is_new || !row.dirty || !row.id) continue
        await productCategoryApi.update(row.id, payloadFor(row, keyToID))
      }

      toast.success('商品分类已保存')
      await fetchCategories()
    } catch (error) {
      console.error('Failed to save product categories:', error)
      toast.error('商品分类保存失败，请检查层级、slug 和名称')
    } finally {
      savingAll.value = false
    }
  }

  return {
    loading,
    savingAll,
    rows,
    maxDepth,
    stats,
    hasChanges,
    addRootCategory,
    addSibling,
    addChild,
    changeParent,
    fetchCategories,
    markDirty,
    normalizeSlug,
    parentOptions,
    removeUnsavedRow,
    removeUnsavedSubtree,
    saveAll,
    subtreeKeys,
  }
}

