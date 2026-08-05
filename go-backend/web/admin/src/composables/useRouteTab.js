import { computed, unref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const routeNamesFor = (value) => Array.isArray(value) ? value : value ? [value] : []

export function useRouteTab({
  defaultValue,
  values,
  routes,
  enabled = true,
}) {
  const route = useRoute()
  const router = useRouter()
  const allowedValues = new Set(values)

  const isEnabled = () => {
    if (typeof enabled === 'function') return Boolean(enabled())
    return Boolean(unref(enabled))
  }

  const normalize = (value) => allowedValues.has(value) ? value : defaultValue
  const routeNameFor = (value) => routeNamesFor(routes?.[value])[0]

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
