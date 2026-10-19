import { describe, expect, it } from 'vitest'
import { claimImageAttachmentIndicesFromOriginalStoredReferences } from './warrantyPresentation'

describe('private warranty image selectors for historical references', () => {
  it('preserves original indices when historical references contain empty entries', () => {
    expect(claimImageAttachmentIndicesFromOriginalStoredReferences({ images: '["", "warranty/first.png", "  ", "warranty/second.png"]' })).toEqual([1, 3])
  })
  it('does not invent selectors for malformed or missing historical image lists', () => {
    expect(claimImageAttachmentIndicesFromOriginalStoredReferences({ images: 'broken JSON' })).toEqual([])
    expect(claimImageAttachmentIndicesFromOriginalStoredReferences()).toEqual([])
  })
})
