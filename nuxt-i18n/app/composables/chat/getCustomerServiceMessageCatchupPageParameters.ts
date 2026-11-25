export interface CustomerServiceMessageCatchupCursor {
  created_at: string
  id: number | string
}

export const getCustomerServiceMessageCatchupPageParameters = (
  limit: number,
  cursor: CustomerServiceMessageCatchupCursor | null,
): Record<string, string | number> => {
  if (!cursor) {
    return { limit, latest: 'true' }
  }

  return {
    limit,
    after_created_at: cursor.created_at,
    after_id: cursor.id,
  }
}
