export type GlobalPartnerPublicDirectoryEntry = {
  id: string
  name: string
  countryName: string
  cityName: string
  latitude: number
  longitude: number
  partnershipChannelId: 'oem-odm' | 'dealer-builder' | 'teams-ambassadors'
  logoUrl?: string
  websiteUrl?: string
}

// Add only partner identities and locations approved for public display.
// Confidential collaborations are counted separately in the page statistics;
// their identities must never be included in this client-side directory.
export const globalPartnerPublicDirectoryEntries: readonly GlobalPartnerPublicDirectoryEntry[] = []
