export const appendNoTransformDirectiveToHtmlCacheControl = (
  cacheControl: string | number | string[] | undefined,
): string => {
  const existingCacheControl = Array.isArray(cacheControl)
    ? cacheControl.join(', ')
    : String(cacheControl ?? '').trim()

  if (existingCacheControl.split(',').some(directive => directive.trim().toLowerCase() === 'no-transform')) {
    return existingCacheControl
  }

  return existingCacheControl ? `${existingCacheControl}, no-transform` : 'no-transform'
}
