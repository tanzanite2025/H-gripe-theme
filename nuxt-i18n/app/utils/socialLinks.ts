export type SocialNetwork = 'facebook' | 'instagram' | 'x' | 'youtube' | 'reddit'

export interface SocialLinkViewModel {
  network: SocialNetwork
  url: string
  label: string
}

export const socialNetworkLabels: Record<SocialNetwork, string> = {
  facebook: 'Facebook',
  instagram: 'Instagram',
  x: 'X',
  youtube: 'YouTube',
  reddit: 'Reddit',
}

export const socialSettingDefinitions = Object.entries(socialNetworkLabels).map(([network, label]) => ({
  network: network as SocialNetwork,
  label,
}))

const supportedSocialNetworks = new Set<SocialNetwork>(Object.keys(socialNetworkLabels) as SocialNetwork[])

export const normalizeSocialLinkItems = (value: unknown): SocialLinkViewModel[] => {
  if (!Array.isArray(value)) return []

  return value
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object')
    .map((item) => {
      const network = String(item.network ?? '').trim().toLowerCase()
      const url = String(item.url ?? '').trim()
      const customLabel = String(item.label ?? '').trim()

      return {
        network,
        url,
        label: customLabel || socialNetworkLabels[network as SocialNetwork] || network.toUpperCase(),
      }
    })
    .filter((item): item is SocialLinkViewModel => (
      Boolean(item.url) && supportedSocialNetworks.has(item.network as SocialNetwork)
    ))
    .map((item) => ({
      ...item,
      network: item.network as SocialNetwork,
    }))
}
