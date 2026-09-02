import { defineNuxtPlugin } from '#imports'
import { attachDeviceFingerprintHeader, resolveDeviceFingerprint } from '~/utils/deviceFingerprint'
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'

export default defineNuxtPlugin(() => {
  scheduleDeferredClientWork(() => {
    void resolveDeviceFingerprint()
  }, { delayMs: 7000, idleTimeoutMs: 3000 })

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
