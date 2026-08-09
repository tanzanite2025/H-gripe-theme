import { computed, unref } from 'vue'
import type { MaybeRef, WritableComputedRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

type RouteTabEnabled = MaybeRef<boolean> | (() => boolean)

interface UseRouteTabOptions<T extends string> {
  defaultValue: T
  values: T[]
  routes?: Partial<Record<T, string | string[]>>
  enabled?: RouteTabEnabled
}

const routeNamesFor = (value?: string | string[] | null): string[] => Array.isArray(value) ? value : value ? [value] : []

export function useRouteTab<T extends string>({
  defaultValue,
  values,
  routes,
  enabled = true,
}: UseRouteTabOptions<T>): WritableComputedRef<T> {
  const route = useRoute()
  const router = useRouter()
  const allowedValues = new Set<T>(values)

  const isEnabled = (): boolean => {
    if (typeof enabled === 'function') return Boolean(enabled())
    return Boolean(unref(enabled))
  }

  const normalize = (value: T): T => allowedValues.has(value) ? value : defaultValue
  const routeNameFor = (value: T): string | undefined => routeNamesFor(routes?.[value])[0]

  return computed({
    get() {
      if (!isEnabled()) return defaultValue

      const routeName = String(route.name || '')
      return values.find((value) => routeNamesFor(routes?.[value]).includes(routeName)) || defaultValue
    },
    set(value) {
      if (!isEnabled()) return

      const targetRouteName = routeNameFor(normalize(value))
      if (!targetRouteName || route.name === targetRouteName) return

      router.push({
        name: targetRouteName,
        query: route.query,
        hash: route.hash,
      })
    }
  })
}
