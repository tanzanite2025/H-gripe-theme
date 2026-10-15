import type { MediaAsset, MediaLibraryPartition } from '@/api/media'

export const mediaLibraryBusinessPartitions = [
  { value: 'operational', label: '运营素材' },
  { value: 'catalog', label: '商品素材' },
  { value: 'content', label: '页面素材' },
  { value: 'shared', label: '公共素材' },
  { value: 'customer_service', label: '客服资料' },
  { value: 'customer_submissions', label: '客户提交' },
  { value: 'unclassified', label: '待整理' },
  { value: 'all', label: '全部资源' },
] as const

export type MediaLibraryPartitionView = typeof mediaLibraryBusinessPartitions[number]['value']

export const mediaLibraryUploadOriginOptions = [
  { value: 'legacy_unknown', label: '历史来源未知' },
  { value: 'administrator_general_upload', label: '后台通用上传' },
  { value: 'administrator_catalog_upload', label: '商品上传入口' },
  { value: 'administrator_content_upload', label: '页面上传入口' },
  { value: 'administrator_website_profile_upload', label: '网站资料上传' },
  { value: 'administrator_shared_upload', label: '公共素材上传' },
  { value: 'administrator_customer_service_reusable_upload', label: '客服常用素材上传' },
  { value: 'customer_service_conversation_upload', label: '会话附件上传' },
  { value: 'administrator_workbench_feed_upload', label: '工作台动态上传' },
] as const

export function normalizeMediaLibraryPartitionView(value: unknown): MediaLibraryPartitionView {
  return mediaLibraryBusinessPartitions.some((partition) => partition.value === value)
    ? value as MediaLibraryPartitionView : 'operational'
}

export function mediaLibraryPartitionLabel(partition?: MediaLibraryPartition | null): string {
  return mediaLibraryBusinessPartitions.find((item) => item.value === partition)?.label || '待整理'
}

export function mediaAssetUploadOriginLabel(origin?: string | null): string {
  return mediaLibraryUploadOriginOptions.find((item) => item.value === origin)?.label || '历史来源未知'
}

export function mediaAssetLibraryOwnershipLabel(asset: MediaAsset): string {
  if (asset.customer_service_media_purpose === 'conversation_attachment' || asset.requires_conversation_access) return '客服会话附件'
  if (asset.customer_service_media_purpose === 'reusable_reply') return '客服常用素材'
  return mediaLibraryPartitionLabel(asset.library_partition)
}
