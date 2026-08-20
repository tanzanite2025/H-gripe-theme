export interface GalleryLightboxLabels {
  close: string
  previousImage: string
  nextImage: string
  imageThumbnails: string
  loadingDetails: string
  noImage: string
  relatedProducts: string
  noRelatedProducts: string
}

export const defaultGalleryLightboxLabels: GalleryLightboxLabels = {
  close: 'Close gallery',
  previousImage: 'Previous image',
  nextImage: 'Next image',
  imageThumbnails: 'Gallery image thumbnails',
  loadingDetails: 'Loading gallery details...',
  noImage: 'No image available',
  relatedProducts: 'Related products',
  noRelatedProducts: 'No related products configured.',
}
