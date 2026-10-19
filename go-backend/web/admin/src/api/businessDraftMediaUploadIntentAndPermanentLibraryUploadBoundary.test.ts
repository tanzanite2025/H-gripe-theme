import { beforeEach, describe, expect, it, vi } from 'vitest'
import axios from '@/utils/axios'
import { mediaApi } from './media'

vi.mock('@/utils/axios', () => ({ default: { post: vi.fn() } }))

describe('business editor and permanent media-library upload intents', () => {
  beforeEach(() => vi.mocked(axios.post).mockReset())

  it.each(['catalog', 'content', 'website-profile'] as const)('uses a server-selected draft workflow for %s', async (context) => {
    const form = new FormData()
    const asset = { id: 42, url: '/uploads/editor.png' }
    vi.mocked(axios.post).mockResolvedValue({ data: { code: 0, data: { asset } } })
    expect(await mediaApi.uploadBusinessDraftAsset(form, context)).toEqual(asset)
    expect(axios.post).toHaveBeenCalledWith(`/api/admin/media/uploads/business-drafts/${context}`, form)
  })

  it('keeps a catalog upload made in the media library permanent', async () => {
    const form = new FormData()
    vi.mocked(axios.post).mockResolvedValue({ data: { code: 0, data: { asset: { id: 42 } } } })
    await mediaApi.uploadAsset(form, 'catalog')
    expect(axios.post).toHaveBeenCalledWith('/api/admin/media/uploads/catalog', form)
  })

  it('rejects a draft upload response without its asset', async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: { code: 0, data: {} } })
    await expect(mediaApi.uploadBusinessDraftAsset(new FormData(), 'content')).rejects.toThrow()
  })
})
