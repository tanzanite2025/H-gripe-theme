import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import { useAsyncData, useState } from '#imports'

const successfulFaqResponseCacheLifetimeMs = 5 * 60 * 1000
const maximumSuccessfulFaqResponseCacheEntries = 40

interface SuccessfulFaqResponse<Value> {
  requestKey: string
  value: Value
  fetchedAt: number
}

/** Keep public FAQ responses within this Nuxt app, isolated by locale and page. */
export async function useStorefrontFaqAsyncDataWithLastSuccessfulResponse<Value>(
  requestKey: MaybeRefOrGetter<string>,
  fetchFaqResponse: () => Promise<Value>,
) {
  // useState is request-scoped on the server and survives client navigation.
  // Never store public content in localStorage or a process-wide SSR cache.
  const successfulResponses = useState<Record<string, SuccessfulFaqResponse<unknown>>>(
    'storefront-faq-last-successful-responses',
    () => ({}),
  )

  const findFreshSuccessfulFaqResponse = (key: string): SuccessfulFaqResponse<Value> | undefined => {
    const response = successfulResponses.value[key]
    if (!response || Date.now() - response.fetchedAt >= successfulFaqResponseCacheLifetimeMs) return
    return response as SuccessfulFaqResponse<Value>
  }

  const { data: response, status, error, refresh } = await useAsyncData(
    requestKey,
    async () => {
      const key = toValue(requestKey)
      const value = await fetchFaqResponse()
      // A truthy envelope also caches a successful null/empty response in the
      // SSR payload, so hydration does not repeat a missing page request.
      const successfulResponse = { requestKey: key, value, fetchedAt: Date.now() }
      successfulResponses.value[key] = successfulResponse
      const retainedKeys = Object.keys(successfulResponses.value)
        .sort((left, right) => successfulResponses.value[right]!.fetchedAt - successfulResponses.value[left]!.fetchedAt)
      for (const expiredKey of retainedKeys.slice(maximumSuccessfulFaqResponseCacheEntries)) {
        delete successfulResponses.value[expiredKey]
      }
      return successfulResponse
    },
    {
      lazy: true,
      dedupe: 'defer',
      getCachedData(key, nuxtApp, context) {
        // Explicit retries and locale/route refreshes always reach the API.
        if (context.cause !== 'initial') return
        if (nuxtApp.isHydrating && nuxtApp.payload.data[key] !== undefined) {
          return nuxtApp.payload.data[key]
        }
        return findFreshSuccessfulFaqResponse(key)
      },
    },
  )

  const data = computed(() => {
    const key = toValue(requestKey)
    // Nuxt may temporarily copy the old key's value during navigation. Do not
    // display another route or language while its replacement is loading.
    if (response.value?.requestKey === key) return response.value.value
    return findFreshSuccessfulFaqResponse(key)?.value
  })

  return {
    data,
    pending: computed(() => status.value === 'pending'
      || (status.value === 'idle' && !error.value && data.value === undefined)),
    error,
    refresh,
  }
}
