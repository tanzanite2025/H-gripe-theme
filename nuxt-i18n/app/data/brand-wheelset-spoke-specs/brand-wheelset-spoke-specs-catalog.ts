import { dtSwissWheelsetSpokeCatalog } from './sources/dt-swiss-wheelset-spoke-specs.js'
import { enveWheelsetSpokeCatalog } from './sources/enve-wheelset-spoke-specs.js'
import { shimanoWheelsetSpokeCatalog } from './sources/shimano-wheelset-spoke-specs.js'
import { zippSramWheelsetSpokeCatalog } from './sources/zipp-sram-wheelset-spoke-specs.js'
import type { BrandWheelsetSpokeCatalog } from './brand-wheelset-spoke-specs-types.js'

export type * from './brand-wheelset-spoke-specs-types.js'

export const brandWheelsetSpokeCatalogs: BrandWheelsetSpokeCatalog[] = [
  dtSwissWheelsetSpokeCatalog,
  enveWheelsetSpokeCatalog,
  shimanoWheelsetSpokeCatalog,
  zippSramWheelsetSpokeCatalog,
]

const assertUniqueCatalogIdentifiers = (catalogs: BrandWheelsetSpokeCatalog[]) => {
  const brandSlugs = new Set<string>()
  const modelSlugs = new Set<string>()

  for (const catalog of catalogs) {
    if (brandSlugs.has(catalog.brandSlug)) {
      throw new Error(`Duplicate brand slug in wheelset spoke catalogs: ${catalog.brandSlug}`)
    }
    brandSlugs.add(catalog.brandSlug)

    for (const wheelset of catalog.wheelsets) {
      if (modelSlugs.has(wheelset.slug)) {
        throw new Error(`Duplicate wheelset model slug in spoke catalogs: ${wheelset.slug}`)
      }
      modelSlugs.add(wheelset.slug)
    }
  }
}

assertUniqueCatalogIdentifiers(brandWheelsetSpokeCatalogs)
