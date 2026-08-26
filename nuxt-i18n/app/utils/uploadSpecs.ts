export type StorefrontUploadSpecCode =
  | 'warranty_evidence'
  | 'after_sales_evidence'
  | 'suggestion_attachment'
  | 'customer_service_attachment'
  | 'user_showcase_image'

interface StorefrontUploadSpec {
  code: StorefrontUploadSpecCode
  label: string
  acceptedExtensions: string[]
  acceptedContentTypes: string[]
  maxFileSizeBytes: number
  maxFiles?: number
  maxTotalSizeBytes?: number
  recommendedLongEdge?: number
  maxWidth?: number
  maxHeight?: number
  maxPixels?: number
}

const rasterExtensions = ['.jpg', '.jpeg', '.png', '.webp']
const rasterTypes = ['image/jpeg', 'image/png', 'image/webp']

export const STOREFRONT_UPLOAD_SPECS: Record<StorefrontUploadSpecCode, StorefrontUploadSpec> = {
  warranty_evidence: {
    code: 'warranty_evidence',
    label: 'Warranty evidence',
    acceptedExtensions: [...rasterExtensions, '.gif'],
    acceptedContentTypes: [...rasterTypes, 'image/gif'],
    maxFileSizeBytes: 8 * 1024 * 1024,
    maxFiles: 10,
    maxTotalSizeBytes: 80 * 1024 * 1024,
    recommendedLongEdge: 1200,
    maxWidth: 8000,
    maxHeight: 8000,
    maxPixels: 24_000_000,
  },
  after_sales_evidence: {
    code: 'after_sales_evidence',
    label: 'After-sales evidence',
    acceptedExtensions: [...rasterExtensions, '.gif'],
    acceptedContentTypes: [...rasterTypes, 'image/gif'],
    maxFileSizeBytes: 8 * 1024 * 1024,
    maxFiles: 10,
    maxTotalSizeBytes: 80 * 1024 * 1024,
    recommendedLongEdge: 1200,
    maxWidth: 8000,
    maxHeight: 8000,
    maxPixels: 24_000_000,
  },
  suggestion_attachment: {
    code: 'suggestion_attachment',
    label: 'Feedback image',
    acceptedExtensions: [...rasterExtensions, '.gif'],
    acceptedContentTypes: [...rasterTypes, 'image/gif'],
    maxFileSizeBytes: 5 * 1024 * 1024,
    recommendedLongEdge: 1200,
    maxWidth: 6000,
    maxHeight: 6000,
    maxPixels: 16_000_000,
  },
  customer_service_attachment: {
    code: 'customer_service_attachment',
    label: 'Customer service image',
    acceptedExtensions: [...rasterExtensions, '.gif'],
    acceptedContentTypes: [...rasterTypes, 'image/gif'],
    maxFileSizeBytes: 5 * 1024 * 1024,
    recommendedLongEdge: 1200,
    maxWidth: 6000,
    maxHeight: 6000,
    maxPixels: 16_000_000,
  },
  user_showcase_image: {
    code: 'user_showcase_image',
    label: 'Showcase image',
    acceptedExtensions: ['.webp'],
    acceptedContentTypes: ['image/webp'],
    maxFileSizeBytes: 5 * 1024 * 1024,
    maxFiles: 10,
    maxTotalSizeBytes: 50 * 1024 * 1024,
    recommendedLongEdge: 1600,
    maxWidth: 6000,
    maxHeight: 6000,
    maxPixels: 16_000_000,
  },
}

export const uploadSpecAccept = (code: StorefrontUploadSpecCode): string => {
  const spec = STOREFRONT_UPLOAD_SPECS[code]
  return [...spec.acceptedExtensions, ...spec.acceptedContentTypes].join(',')
}

export const uploadSpecHint = (code: StorefrontUploadSpecCode): string => {
  const spec = STOREFRONT_UPLOAD_SPECS[code]
  const dimensions = spec.recommendedLongEdge
    ? `Source long edge >= ${spec.recommendedLongEdge}px`
    : ''
  const formats = spec.acceptedExtensions.map((value) => value.replace('.', '').toUpperCase()).join(' / ')
  const size = `${Math.round(spec.maxFileSizeBytes / 1024 / 1024)}MB each`
  const files = spec.maxFiles ? `, up to ${spec.maxFiles} files` : ''
  const total = spec.maxTotalSizeBytes
    ? `, ${Math.round(spec.maxTotalSizeBytes / 1024 / 1024)}MB total`
    : ''
  return [dimensions, formats, size + files + total].filter(Boolean).join(' · ')
}

interface UploadValidationResult {
  ok: boolean
  error?: string
  warning?: string
  width?: number
  height?: number
}

const fileExtension = (file: File): string => {
  const index = file.name.lastIndexOf('.')
  return index >= 0 ? file.name.slice(index).toLowerCase() : ''
}

const readImageDimensions = (file: File): Promise<{ width: number; height: number }> => new Promise((resolve, reject) => {
  const objectURL = URL.createObjectURL(file)
  const image = new Image()
  image.onload = () => {
    URL.revokeObjectURL(objectURL)
    resolve({ width: image.naturalWidth || image.width, height: image.naturalHeight || image.height })
  }
  image.onerror = () => {
    URL.revokeObjectURL(objectURL)
    reject(new Error('Unable to read image dimensions'))
  }
  image.src = objectURL
})

export const validateStorefrontUploadFile = async (
  file: File,
  code: StorefrontUploadSpecCode,
): Promise<UploadValidationResult> => {
  const spec = STOREFRONT_UPLOAD_SPECS[code]
  const extension = fileExtension(file)
  if (!spec.acceptedExtensions.includes(extension) && !spec.acceptedContentTypes.includes(file.type)) {
    return { ok: false, error: `${spec.label} must use ${spec.acceptedExtensions.join(', ')}.` }
  }
  if (file.size > spec.maxFileSizeBytes) {
    return { ok: false, error: `${file.name} exceeds the ${Math.round(spec.maxFileSizeBytes / 1024 / 1024)}MB per-file limit.` }
  }

  let width = 0
  let height = 0
  try {
    const dimensions = await readImageDimensions(file)
    width = dimensions.width
    height = dimensions.height
  } catch (error) {
    return { ok: false, error: error instanceof Error ? error.message : 'Unable to read image dimensions' }
  }

  if (spec.maxWidth && width > spec.maxWidth) {
    return { ok: false, error: `${file.name} is wider than ${spec.maxWidth}px.` }
  }
  if (spec.maxHeight && height > spec.maxHeight) {
    return { ok: false, error: `${file.name} is taller than ${spec.maxHeight}px.` }
  }
  if (spec.maxPixels && width * height > spec.maxPixels) {
    return { ok: false, error: `${file.name} exceeds the ${spec.maxPixels.toLocaleString()} pixel limit.` }
  }

  const warning = spec.recommendedLongEdge && Math.max(width, height) < spec.recommendedLongEdge
    ? `${file.name} is below the recommended ${spec.recommendedLongEdge}px long edge and may look soft when enlarged.`
    : undefined

  return { ok: true, warning, width, height }
}

export const validateStorefrontUploadFiles = async (
  files: File[],
  code: StorefrontUploadSpecCode,
): Promise<UploadValidationResult> => {
  const spec = STOREFRONT_UPLOAD_SPECS[code]
  if (spec.maxFiles && files.length > spec.maxFiles) {
    return { ok: false, error: `${spec.label} allows up to ${spec.maxFiles} files.` }
  }
  const totalSize = files.reduce((sum, file) => sum + file.size, 0)
  if (spec.maxTotalSizeBytes && totalSize > spec.maxTotalSizeBytes) {
    return { ok: false, error: `${spec.label} must stay within ${Math.round(spec.maxTotalSizeBytes / 1024 / 1024)}MB total.` }
  }
  for (const file of files) {
    const result = await validateStorefrontUploadFile(file, code)
    if (!result.ok) return result
  }
  return { ok: true }
}
