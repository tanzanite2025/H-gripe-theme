import { mkdir, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'

import { brandWheelsetSpokeCatalogs } from '../app/data/brand-wheelset-spoke-specs/brand-wheelset-spoke-specs-catalog'

const projectRoot = resolve(process.cwd(), '..')
const backendCatalogPath = resolve(
  projectRoot,
  'go-backend/internal/api/v1/brandwheelsetspoke/brand-wheelset-spoke-specs-catalog.json',
)
const publicCatalogPath = resolve(
  process.cwd(),
  'app/data/brand-wheelset-spoke-specs/brand-wheelset-spoke-specs-public-catalog.ts',
)

const publicCatalogs = brandWheelsetSpokeCatalogs
  .map((catalog) => ({
    brandSlug: catalog.brandSlug,
    brandName: catalog.brandName,
    publicationStatus: catalog.publicationStatus,
    sourceCheckedAt: catalog.sourceCheckedAt,
    wheelsets: catalog.wheelsets
      .filter((wheelset) => wheelset.verificationStatus === 'verified')
      .map((wheelset) => ({
        slug: wheelset.slug,
        model: wheelset.model,
        lifecycleStatus: wheelset.lifecycleStatus,
        rim: { depthMm: wheelset.rim.depthMm },
        wheels: wheelset.wheels.map((wheel) => ({
          position: wheel.position,
          spokeCount: wheel.spokeCount,
          lacingPattern: wheel.lacingPattern,
          sides: wheel.sides.map((side) => ({
            side: side.side,
            spokeModel: side.spokeModel,
            headType: side.headType,
          })),
        })),
      })),
  }))
  .filter((catalog) => catalog.wheelsets.length > 0)

const publicCatalogHeader = `export type PublicBrandWheelsetPublicationStatus = 'draft' | 'published'
export type PublicBrandWheelsetLifecycleStatus = 'current' | 'legacy'
export type PublicBrandWheelPosition = 'front' | 'rear'
export type PublicBrandWheelSpokeSide = 'left' | 'right'
export type PublicBrandSpokeHeadType = 'straight-pull' | 'j-bend' | 'unknown'

export interface PublicBrandWheelSideSpokeSpec {
  side: PublicBrandWheelSpokeSide
  spokeModel: string
  headType: PublicBrandSpokeHeadType
}

export interface PublicBrandWheelPositionSpokeSpec {
  position: PublicBrandWheelPosition
  spokeCount: number
  lacingPattern: string
  sides: PublicBrandWheelSideSpokeSpec[]
}

export interface PublicBrandWheelsetRecord {
  slug: string
  model: string
  lifecycleStatus: PublicBrandWheelsetLifecycleStatus
  rim: { depthMm: number }
  wheels: PublicBrandWheelPositionSpokeSpec[]
}

export interface PublicBrandWheelsetCatalog {
  brandSlug: string
  brandName: string
  publicationStatus: PublicBrandWheelsetPublicationStatus
  sourceCheckedAt: string | null
  wheelsets: PublicBrandWheelsetRecord[]
}

`

await mkdir(dirname(backendCatalogPath), { recursive: true })
await writeFile(backendCatalogPath, `${JSON.stringify(brandWheelsetSpokeCatalogs, null, 2)}\n`, 'utf8')
await writeFile(
  publicCatalogPath,
  `${publicCatalogHeader}export const publicBrandWheelsetSpokeCatalogs: PublicBrandWheelsetCatalog[] = ${JSON.stringify(publicCatalogs, null, 2)}\n`,
  'utf8',
)

const wheelsets = brandWheelsetSpokeCatalogs.flatMap(catalog => catalog.wheelsets)
console.log(
  `Synchronized ${wheelsets.length} wheelset repair-kit records (${wheelsets.filter(item => item.lifecycleStatus === 'legacy').length} legacy).`,
)
