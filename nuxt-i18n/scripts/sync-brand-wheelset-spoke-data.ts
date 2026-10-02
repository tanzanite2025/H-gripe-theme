import { mkdir, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'

import { brandWheelsetSpokeCatalogs } from '../app/data/brand-wheelset-spoke-specs/brand-wheelset-spoke-specs-catalog.js'
import type {
  BrandWheelPositionSpokeSpec,
  BrandWheelSideSpokeSpec,
  BrandWheelsetSpokeCatalog,
  BrandWheelsetSpokeRecord,
} from '../app/data/brand-wheelset-spoke-specs/brand-wheelset-spoke-specs-types.js'

const projectRoot = resolve(process.cwd(), '..')
const backendCatalogPath = resolve(
  projectRoot,
  'go-backend/internal/domain/wheelsetcatalog/brand-wheelset-spoke-specs-catalog.json',
)
const publicCatalogPath = resolve(
  process.cwd(),
  'app/data/brand-wheelset-spoke-specs/brand-wheelset-spoke-specs-public-catalog.ts',
)

const typedCatalogs = brandWheelsetSpokeCatalogs as BrandWheelsetSpokeCatalog[]
const publicCatalogs = typedCatalogs
  .filter((catalog: BrandWheelsetSpokeCatalog) => catalog.publicationStatus === 'published')
  .map((catalog: BrandWheelsetSpokeCatalog) => ({
    brandSlug: catalog.brandSlug,
    brandName: catalog.brandName,
    publicationStatus: catalog.publicationStatus,
    sourceCheckedAt: catalog.sourceCheckedAt,
    wheelsets: catalog.wheelsets
      .filter((wheelset: BrandWheelsetSpokeRecord) => wheelset.verificationStatus === 'verified')
      .map((wheelset: BrandWheelsetSpokeRecord) => ({
        slug: wheelset.slug,
        model: wheelset.model,
        lifecycleStatus: wheelset.lifecycleStatus,
        nippleModel: wheelset.nippleModel,
        nippleLengthMm: wheelset.nippleLengthMm,
        rim: {
          depthMm: wheelset.rim.depthMm,
          depthFrontMm: wheelset.rim.depthFrontMm,
          depthRearMm: wheelset.rim.depthRearMm,
        },
        wheels: wheelset.wheels.map((wheel: BrandWheelPositionSpokeSpec) => ({
          position: wheel.position,
          spokeCount: wheel.spokeCount,
          lacingPattern: wheel.lacingPattern,
          sides: wheel.sides.map((side: BrandWheelSideSpokeSpec) => ({
            side: side.side,
            lengthMm: side.lengthMm,
            spokeModel: side.spokeModel,
            headType: side.headType,
          })),
        })),
      })),
  }))
  .filter((catalog: { wheelsets: unknown[] }) => catalog.wheelsets.length > 0)

const publicCatalogHeader = `export type PublicBrandWheelsetPublicationStatus = 'draft' | 'published'
export type PublicBrandWheelsetLifecycleStatus = 'current' | 'legacy'
export type PublicBrandWheelPosition = 'front' | 'rear'
export type PublicBrandWheelSpokeSide = 'left' | 'right'
export type PublicBrandSpokeHeadType = 'straight-pull' | 'j-bend' | 'unknown'

export interface PublicBrandWheelSideSpokeSpec {
  side: PublicBrandWheelSpokeSide
  lengthMm: number | null
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
  nippleModel: string
  nippleLengthMm: number | null
  rim: {
    depthMm: number
    depthFrontMm?: number
    depthRearMm?: number
  }
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

const wheelsets = typedCatalogs.flatMap((catalog: BrandWheelsetSpokeCatalog) => catalog.wheelsets)
console.log(
  `Synchronized ${wheelsets.length} wheelset repair-kit records (${wheelsets.filter((item: BrandWheelsetSpokeRecord) => item.lifecycleStatus === 'legacy').length} legacy).`,
)
