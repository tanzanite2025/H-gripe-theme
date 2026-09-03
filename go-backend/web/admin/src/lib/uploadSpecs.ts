export type UploadSpecCode =
  | 'product_image'
  | 'product_description_image'
  | 'product_variant_swatch'
  | 'media_library_image'
  | 'faq_answer_image'
  | 'visual_showcase_home_categories'
  | 'visual_showcase_editorial'
  | 'site_logo'
  | 'site_favicon'
  | 'customer_service_avatar'
  | 'website_profile_avatar'
  | 'website_profile_image'
  | 'refund_cancellation_image'
  | 'warranty_evidence'
  | 'after_sales_evidence'
  | 'suggestion_attachment'
  | 'customer_service_attachment'
  | 'user_showcase_image'

export interface UploadSpec {
  code: UploadSpecCode
  kind: 'image' | 'svg'
  label: string
  description: string
  acceptedExtensions: string[]
  acceptedContentTypes: string[]
  maxFileSizeBytes: number
  maxFiles?: number
  maxTotalSizeBytes?: number
  exactWidth?: number
  exactHeight?: number
  recommendedWidth?: number
  recommendedHeight?: number
  recommendedLongEdge?: number
  maxWidth?: number
  maxHeight?: number
  maxPixels?: number
  aspectRatioWidth?: number
  aspectRatioHeight?: number
  aspectRatioLabel?: string
  qualityNote: string
}

const imageTypes = ['image/jpeg', 'image/png', 'image/webp']
const imageExtensions = ['.jpg', '.jpeg', '.png', '.webp']
const productLimits = {
  maxFileSizeBytes: 12 * 1024 * 1024,
  maxWidth: 8000,
  maxHeight: 8000,
  maxPixels: 24_000_000,
}

export const UPLOAD_SPECS: Record<UploadSpecCode, UploadSpec> = {
  product_image: {
    code: 'product_image',
    kind: 'image',
    label: '商品图片',
    description: '商品主图和图库图片',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    ...productLimits,
    recommendedWidth: 1600,
    recommendedHeight: 1600,
    qualityNote: '建议至少 1600×1600 px；非正方形图片保持原比例。',
  },
  product_description_image: {
    code: 'product_description_image',
    kind: 'image',
    label: '商品详情图片',
    description: '插入商品详情正文的图片',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    ...productLimits,
    recommendedLongEdge: 1600,
    qualityNote: '建议长边至少 1600 px，桌面端放大查看时更清晰。',
  },
  product_variant_swatch: {
    code: 'product_variant_swatch',
    kind: 'image',
    label: 'SKU 色板图片',
    description: '颜色、表面处理等选项展示图',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    maxFileSizeBytes: 2 * 1024 * 1024,
    maxWidth: 2048,
    maxHeight: 2048,
    maxPixels: 4_194_304,
    recommendedWidth: 512,
    recommendedHeight: 512,
    qualityNote: '建议使用主体居中的 512×512 px 正方形图片。',
  },
  media_library_image: {
    code: 'media_library_image',
    kind: 'image',
    label: '媒体库图片',
    description: '可复用的通用媒体资源',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    ...productLimits,
    recommendedLongEdge: 1600,
    qualityNote: '建议长边至少 1600 px，方便多个前台位置复用。',
  },
  faq_answer_image: {
    code: 'faq_answer_image',
    kind: 'image',
    label: 'FAQ 答案图片',
    description: 'FAQ 答案附图',
    acceptedExtensions: ['.webp'],
    acceptedContentTypes: ['image/webp'],
    maxFileSizeBytes: 3 * 1024 * 1024,
    exactWidth: 800,
    exactHeight: 800,
    maxWidth: 800,
    maxHeight: 800,
    maxPixels: 640_000,
    recommendedWidth: 800,
    recommendedHeight: 800,
    qualityNote: '必须是 800×800 px WebP，每条 FAQ 最多一张。',
  },
  visual_showcase_home_categories: {
    code: 'visual_showcase_home_categories',
    kind: 'image',
    label: '首页视觉展示图片',
    description: '首页产品分类视觉展示',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    ...productLimits,
    recommendedWidth: 1920,
    recommendedHeight: 1080,
    aspectRatioWidth: 16,
    aspectRatioHeight: 9,
    aspectRatioLabel: '16:9',
    qualityNote: '必须保持 16:9，建议 1920×1080 px。',
  },
  visual_showcase_editorial: {
    code: 'visual_showcase_editorial',
    kind: 'image',
    label: '竖版视觉展示图片',
    description: '编辑型竖版视觉展示',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    ...productLimits,
    recommendedWidth: 1200,
    recommendedHeight: 1600,
    aspectRatioWidth: 3,
    aspectRatioHeight: 4,
    aspectRatioLabel: '3:4',
    qualityNote: '必须保持 3:4，建议 1200×1600 px。',
  },
  site_logo: {
    code: 'site_logo',
    kind: 'image',
    label: '站点 Logo',
    description: '站点标识',
    acceptedExtensions: ['.webp'],
    acceptedContentTypes: ['image/webp'],
    maxFileSizeBytes: 1 * 1024 * 1024,
    exactWidth: 512,
    exactHeight: 512,
    recommendedWidth: 512,
    recommendedHeight: 512,
    qualityNote: '必须是 512×512 px 正方形 WebP；前台会按展示位自动缩放。',
  },
  site_favicon: {
    code: 'site_favicon',
    kind: 'image',
    label: '站点 Favicon',
    description: '浏览器标签、收藏夹和 PWA 图标',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    maxFileSizeBytes: 1 * 1024 * 1024,
    maxWidth: 1024,
    maxHeight: 1024,
    maxPixels: 1_048_576,
    recommendedWidth: 512,
    recommendedHeight: 512,
    qualityNote: '建议使用 512×512 px 正方形 PNG 或 WebP。',
  },
  customer_service_avatar: {
    code: 'customer_service_avatar',
    kind: 'image',
    label: '客服头像',
    description: '客服人员头像',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    maxFileSizeBytes: 2 * 1024 * 1024,
    maxWidth: 2048,
    maxHeight: 2048,
    maxPixels: 4_194_304,
    recommendedWidth: 512,
    recommendedHeight: 512,
    qualityNote: '建议使用主体居中的 512×512 px 正方形图片。',
  },
  website_profile_avatar: {
    code: 'website_profile_avatar',
    kind: 'image',
    label: '网站资料头像',
    description: '网站资料页面的头像',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    maxFileSizeBytes: 2 * 1024 * 1024,
    maxWidth: 2048,
    maxHeight: 2048,
    maxPixels: 4_194_304,
    recommendedWidth: 512,
    recommendedHeight: 512,
    qualityNote: '建议使用主体居中的 512×512 px 正方形图片。',
  },
  website_profile_image: {
    code: 'website_profile_image',
    kind: 'image',
    label: '网站资料图片',
    description: '工厂、公司或网站资料图片',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    ...productLimits,
    recommendedLongEdge: 1600,
    qualityNote: '建议长边至少 1600 px。',
  },
  refund_cancellation_image: {
    code: 'refund_cancellation_image',
    kind: 'image',
    label: '退款取消说明图片',
    description: '退款取消政策中的说明图片',
    acceptedExtensions: imageExtensions,
    acceptedContentTypes: imageTypes,
    ...productLimits,
    recommendedLongEdge: 1600,
    qualityNote: '建议长边至少 1600 px。',
  },
  warranty_evidence: {
    code: 'warranty_evidence',
    kind: 'image',
    label: '保修凭据图片',
    description: '后台保修或发货记录图片',
    acceptedExtensions: [...imageExtensions, '.gif'],
    acceptedContentTypes: [...imageTypes, 'image/gif'],
    maxFileSizeBytes: 8 * 1024 * 1024,
    maxFiles: 10,
    maxTotalSizeBytes: 80 * 1024 * 1024,
    maxWidth: 8000,
    maxHeight: 8000,
    maxPixels: 24_000_000,
    recommendedLongEdge: 1200,
    qualityNote: '建议长边至少 1200 px，确保标签和损伤细节可读。',
  },
  after_sales_evidence: {
    code: 'after_sales_evidence',
    kind: 'image',
    label: '售后凭据图片',
    description: '客户售后申请中的证据图片',
    acceptedExtensions: [...imageExtensions, '.gif'],
    acceptedContentTypes: [...imageTypes, 'image/gif'],
    maxFileSizeBytes: 8 * 1024 * 1024,
    maxFiles: 10,
    maxTotalSizeBytes: 80 * 1024 * 1024,
    maxWidth: 8000,
    maxHeight: 8000,
    maxPixels: 24_000_000,
    recommendedLongEdge: 1200,
    qualityNote: '建议长边至少 1200 px，确保标签和损伤细节可读。',
  },
  suggestion_attachment: {
    code: 'suggestion_attachment',
    kind: 'image',
    label: '意见反馈图片',
    description: '客户意见反馈中的图片',
    acceptedExtensions: [...imageExtensions, '.gif'],
    acceptedContentTypes: [...imageTypes, 'image/gif'],
    maxFileSizeBytes: 5 * 1024 * 1024,
    maxWidth: 6000,
    maxHeight: 6000,
    maxPixels: 16_000_000,
    recommendedLongEdge: 1200,
    qualityNote: '包含文字或细节时，建议长边至少 1200 px。',
  },
  customer_service_attachment: {
    code: 'customer_service_attachment',
    kind: 'image',
    label: '客服对话附件',
    description: '客服对话中的图片附件',
    acceptedExtensions: [...imageExtensions, '.gif'],
    acceptedContentTypes: [...imageTypes, 'image/gif'],
    maxFileSizeBytes: 5 * 1024 * 1024,
    maxWidth: 6000,
    maxHeight: 6000,
    maxPixels: 16_000_000,
    recommendedLongEdge: 1200,
    qualityNote: '包含文字或细节时，建议长边至少 1200 px。',
  },
  user_showcase_image: {
    code: 'user_showcase_image',
    kind: 'image',
    label: '用户晒单图片',
    description: '公开图片仓库中的用户晒单',
    acceptedExtensions: ['.webp'],
    acceptedContentTypes: ['image/webp'],
    maxFileSizeBytes: 5 * 1024 * 1024,
    maxFiles: 10,
    maxTotalSizeBytes: 50 * 1024 * 1024,
    maxWidth: 6000,
    maxHeight: 6000,
    maxPixels: 16_000_000,
    recommendedLongEdge: 1600,
    qualityNote: '只支持 WebP；建议长边至少 1600 px。',
  },
}

export const uploadSpecAccept = (code: UploadSpecCode): string => {
  const spec = UPLOAD_SPECS[code]
  return [...spec.acceptedExtensions, ...spec.acceptedContentTypes].join(',')
}

const formatBytes = (bytes: number): string => {
  if (bytes >= 1024 * 1024) {
    const megabytes = bytes / (1024 * 1024)
    return `${Number.isInteger(megabytes) ? megabytes : megabytes.toFixed(1)} MB`
  }
  return `${Math.round(bytes / 1024)} KB`
}

export const uploadSpecHint = (code: UploadSpecCode): string => {
  const spec = UPLOAD_SPECS[code]
  const dimensionLabel = spec.kind === 'svg' ? '文件' : '源图'
  const dimensions = spec.exactWidth && spec.exactHeight
    ? `${dimensionLabel}必须 ${spec.exactWidth}×${spec.exactHeight} px`
    : spec.recommendedWidth && spec.recommendedHeight
      ? `${dimensionLabel}建议 ≥${spec.recommendedWidth}×${spec.recommendedHeight} px`
      : spec.recommendedLongEdge
        ? `${dimensionLabel}建议长边 ≥${spec.recommendedLongEdge} px`
        : ''
  const ratio = spec.aspectRatioLabel ? ` · 比例 ${spec.aspectRatioLabel}` : ''
  const files = spec.maxFiles ? ` · 最多 ${spec.maxFiles} 张` : ''
  return `${dimensions}${ratio} · ${spec.acceptedExtensions.join(' / ').replace(/\./g, '').toUpperCase()} · 单张 ≤${formatBytes(spec.maxFileSizeBytes)}${files}`
}

export interface UploadValidationResult {
  ok: boolean
  warning?: string
  error?: string
  width?: number
  height?: number
}

const extensionOf = (filename: string): string => {
  const index = filename.lastIndexOf('.')
  return index >= 0 ? filename.slice(index).toLowerCase() : ''
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
    reject(new Error('无法读取图片尺寸'))
  }
  image.src = objectURL
})

export const validateUploadFile = async (
  file: File,
  code: UploadSpecCode,
): Promise<UploadValidationResult> => {
  const spec = UPLOAD_SPECS[code]
  const extension = extensionOf(file.name)
  if (
    !spec.acceptedExtensions.includes(extension)
    && !spec.acceptedContentTypes.includes(file.type)
  ) {
    return { ok: false, error: `${spec.label}仅支持 ${spec.acceptedExtensions.join(' / ')} 格式` }
  }
  if (file.size > spec.maxFileSizeBytes) {
    return { ok: false, error: `${spec.label}单张不能超过 ${formatBytes(spec.maxFileSizeBytes)}` }
  }

  let dimensions: { width: number; height: number } | null = null
  if (spec.kind === 'image') {
    try {
      dimensions = await readImageDimensions(file)
    } catch (error) {
      return { ok: false, error: error instanceof Error ? error.message : '无法读取图片尺寸' }
    }
    const { width, height } = dimensions
    if (spec.exactWidth && spec.exactHeight && (width !== spec.exactWidth || height !== spec.exactHeight)) {
      return { ok: false, error: `${spec.label}必须是 ${spec.exactWidth}×${spec.exactHeight} px，当前为 ${width}×${height}` }
    }
    if (spec.maxWidth && width > spec.maxWidth) {
      return { ok: false, error: `${spec.label}宽度不能超过 ${spec.maxWidth} px` }
    }
    if (spec.maxHeight && height > spec.maxHeight) {
      return { ok: false, error: `${spec.label}高度不能超过 ${spec.maxHeight} px` }
    }
    if (spec.maxPixels && width * height > spec.maxPixels) {
      return { ok: false, error: `${spec.label}像素总数不能超过 ${spec.maxPixels.toLocaleString()} px` }
    }
    if (spec.aspectRatioWidth && spec.aspectRatioHeight && width * spec.aspectRatioHeight !== height * spec.aspectRatioWidth) {
      return { ok: false, error: `${spec.label}必须保持 ${spec.aspectRatioLabel} 比例，当前为 ${width}×${height}` }
    }
  }

  const warning = dimensions && (
    spec.recommendedLongEdge
      ? Math.max(dimensions.width, dimensions.height) < spec.recommendedLongEdge
      : spec.recommendedWidth && spec.recommendedHeight
        ? dimensions.width < spec.recommendedWidth || dimensions.height < spec.recommendedHeight
        : false
  )
    ? `${spec.label}低于建议尺寸，上传后在大尺寸展示时可能偏软`
    : undefined

  return {
    ok: true,
    warning,
    width: dimensions?.width,
    height: dimensions?.height,
  }
}
