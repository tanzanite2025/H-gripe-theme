import { computed } from 'vue'

/**
 * 统一的 API Base URL 获取
 * 
 * 用于替代在每个文件中重复的 API URL 构造逻辑
 * 
 * @returns computed API base URL (去除尾部斜杠)
 * 
 * @example
 * ```typescript
 * const apiBase = useApiBase()
 * const products = await $fetch(`${apiBase.value}/products`)
 * ```
 */
export const useApiBase = () => {
  const config = useRuntimeConfig()

  return computed(() => {
    const base = (config.public as { apiBase?: string }).apiBase || '/api/v1'
    // 移除尾部斜杠，确保 URL 格式一致
    return base.replace(/\/$/, '')
  })
}
