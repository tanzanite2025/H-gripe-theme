export * from './fitmentCatalog/types'
export { frameFitmentEntriesApi } from './fitmentCatalog/frameEntries'
export { forkFitmentEntriesApi } from './fitmentCatalog/forkEntries'
export { fitmentHubSpecificationsApi } from './fitmentCatalog/hubSpecifications'

import { frameFitmentEntriesApi } from './fitmentCatalog/frameEntries'
import { forkFitmentEntriesApi } from './fitmentCatalog/forkEntries'
import { fitmentHubSpecificationsApi } from './fitmentCatalog/hubSpecifications'

export const fitmentCatalogApi = {
  ...frameFitmentEntriesApi,
  ...forkFitmentEntriesApi,
  ...fitmentHubSpecificationsApi,
}

export default fitmentCatalogApi
