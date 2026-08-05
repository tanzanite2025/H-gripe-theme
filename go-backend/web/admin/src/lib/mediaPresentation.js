export const assetTitle = (asset) => (
  asset?.alt
  || asset?.original_filename
  || asset?.filename
  || `媒体 #${asset?.id || '-'}`
)

export const assetAccessURL = (asset) => asset?.access_url || asset?.url || ''

export const statusLabel = (status) => status === 'archived' ? '归档' : '启用'

export const mediaTypeLabel = (type) => type === 'video' ? '视频' : '图片'

export const formatMediaDate = (date) => date ? new Date(date).toLocaleDateString('zh-CN') : '-'

export const formatMediaSize = (bytes) => {
  const value = Number(bytes || 0)
  if (value >= 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB`
  if (value >= 1024) return `${Math.round(value / 1024)} KB`
  return `${value} B`
}
