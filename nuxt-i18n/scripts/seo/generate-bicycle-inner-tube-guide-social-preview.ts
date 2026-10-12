import { fileURLToPath } from 'node:url'
import sharp from 'sharp'

const innerTubeGuideSocialPreviewSourcePath = fileURLToPath(new URL(
  '../../public/images/guides/inner-tube/bicycle-inner-tube-size-valve-and-rim-depth-guide-social-preview.svg',
  import.meta.url,
))
const innerTubeGuideSocialPreviewOutputPath = innerTubeGuideSocialPreviewSourcePath.replace(/\.svg$/, '.png')

await sharp(innerTubeGuideSocialPreviewSourcePath)
  .resize(1200, 630)
  .png({ compressionLevel: 9, palette: true, quality: 90, colours: 128, effort: 10 })
  .toFile(innerTubeGuideSocialPreviewOutputPath)

console.log(`Generated 1200 × 630 inner-tube guide social preview: ${innerTubeGuideSocialPreviewOutputPath}`)
