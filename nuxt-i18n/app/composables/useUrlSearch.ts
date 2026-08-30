import { computed, type Ref } from 'vue'
import type { StorefrontURLSearchProfile } from '~/data/url-search/types'

const normalizeSearchText = (value: string): string => value.toLowerCase().replace(/\s+/g, ' ').trim()

const profileSearchText = (profile: StorefrontURLSearchProfile): string => {
  const route = profile.route_entry
  return [
    profile.display_title,
    profile.display_summary,
    ...(Array.isArray(profile.keywords) ? profile.keywords : []),
    route?.path || '',
    route?.title || '',
    route?.summary || '',
    route?.source_key || '',
  ]
    .map(normalizeSearchText)
    .filter(Boolean)
    .join(' ')
}

const tokenizeSearchQuery = (query: string): string[] => {
  const normalized = normalizeSearchText(query)
  if (!normalized) return []
  return normalized.split(' ').filter(Boolean)
}

const rankSearchProfile = (profile: StorefrontURLSearchProfile, queryTokens: string[], queryText: string): number => {
  const haystack = profileSearchText(profile)
  if (!haystack) return -1

  let score = profile.search_weight || 0
  for (const token of queryTokens) {
    if (!haystack.includes(token)) return -1
    score += 10
  }

  if (queryText && haystack.includes(queryText)) {
    score += 15
  }

  const route = profile.route_entry
  if (route?.path && normalizeSearchText(route.path) === queryText) {
    score += 40
  }
  if (profile.display_title && normalizeSearchText(profile.display_title) === queryText) {
    score += 20
  }

  return score
}

export function useUrlSearch(
  urlSearchProfiles: Readonly<Ref<StorefrontURLSearchProfile[]>>,
  searchQuery: Readonly<Ref<string>>,
) {
  const filteredProfiles = computed(() => {
    const queryText = normalizeSearchText(searchQuery.value)
    if (!queryText) return []

    const queryTokens = tokenizeSearchQuery(searchQuery.value)
    return [...urlSearchProfiles.value]
      .map((profile) => ({
        profile,
        score: rankSearchProfile(profile, queryTokens, queryText),
      }))
      .filter((entry) => entry.score >= 0)
      .sort((left, right) => {
        if (right.score !== left.score) return right.score - left.score
        const leftPath = left.profile.route_entry?.path || ''
        const rightPath = right.profile.route_entry?.path || ''
        return leftPath.localeCompare(rightPath)
      })
      .map((entry) => entry.profile)
  })

  const searchResults = computed(() => filteredProfiles.value.slice(0, 6))
  const searchResultCount = computed(() => filteredProfiles.value.length)

  return {
    filteredProfiles,
    searchResults,
    searchResultCount,
  }
}
