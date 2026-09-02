import { defineNuxtPlugin } from '#imports'
import { attachDeviceFingerprintHeader, resolveDeviceFingerprint } from '~/utils/deviceFingerprint'
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'
import { STOREFRONT_DEVICE_FINGERPRINT_WARMUP } from '~/utils/storefrontLoadingPolicy'

export default defineNuxtPlugin(() => {
  scheduleDeferredClientWork(() => {
    void resolveDeviceFingerprint()
  }, STOREFRONT_DEVICE_FINGERPRINT_WARMUP)

  const currentFetch = (globalThis as { $fetch?: any }).$fetch
  if (!currentFetch?.create) {
    return
  }

  ;(globalThis as { $fetch?: any }).$fetch = currentFetch.create({
    async onRequest({ options }: { options: { headers?: HeadersInit } }) {
      options.headers = await attachDeviceFingerprintHeader(options.headers)
    },
  })
})
