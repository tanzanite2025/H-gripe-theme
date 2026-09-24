export const warrantyResultFromResponse = <T>(response: {
  code: number
  data?: { success: boolean; data?: T }
}): T | null => {
  if (response.code !== 0 || response.data?.success !== true) return null
  return response.data.data ?? null
}
