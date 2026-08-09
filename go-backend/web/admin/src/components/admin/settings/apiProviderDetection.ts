interface KnownAPIProvider {
  fragments: string[]
  label: string
}

const knownAPIProviders: KnownAPIProvider[] = [
  { fragments: ['frankfurter'], label: 'Frankfurter' },
  { fragments: ['openexchangerates'], label: 'Open Exchange Rates' },
  { fragments: ['exchangerate-api'], label: 'ExchangeRate-API' },
  { fragments: ['exchangerate'], label: 'ExchangeRate' },
  { fragments: ['currencyapi'], label: 'Currency API' },
  { fragments: ['freecurrencyapi'], label: 'FreeCurrencyAPI' },
  { fragments: ['worldtimeapi'], label: 'WorldTimeAPI' },
  { fragments: ['timeapi'], label: 'TimeAPI' },
]

export const inferAPIProviderFromEndpoint = (endpoint: unknown): string => {
  const raw = String(endpoint || '').trim()
  if (!raw) return ''

  try {
    const url = new URL(raw.includes('://') ? raw : `https://${raw}`)
    const host = url.hostname.replace(/^www\./, '').toLowerCase()
    const matched = knownAPIProviders.find((provider) =>
      provider.fragments.some((fragment) => host.includes(fragment)),
    )
    return matched?.label || host
  } catch {
    return ''
  }
}
