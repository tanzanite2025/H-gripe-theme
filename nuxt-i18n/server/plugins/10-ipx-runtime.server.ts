import sharp from 'sharp'
import { defineNitroPlugin } from 'nitropack/runtime'

const positiveInteger = (value: string | undefined, fallback: number) => {
  const parsed = Number.parseInt(String(value || ''), 10)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback
}

export default defineNitroPlugin(() => {
  const concurrency = positiveInteger(process.env.STOREFRONT_IMAGE_SHARP_CONCURRENCY, 1)
  const cacheMemory = positiveInteger(process.env.STOREFRONT_IMAGE_SHARP_CACHE_MEMORY_MB, 32)

  sharp.concurrency(concurrency)
  sharp.cache({
    memory: cacheMemory,
    files: 0,
    items: 100,
  })
})
