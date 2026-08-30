import { beforeEach, describe, expect, it, vi } from 'vitest'

const axiosMock = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
}))

vi.mock('@/utils/axios', () => ({
  default: axiosMock,
}))

import selectionConfigurationKeyApi from './selectionConfigurationKeys'
import { selectionConfigurationKeyKindAnswerKey, selectionConfigurationKeyKindQuestionKey } from '@/modules/selection-configuration/selectionConfigurationKeys'

describe('selectionConfigurationKeyApi', () => {
  beforeEach(() => {
    axiosMock.get.mockReset()
    axiosMock.post.mockReset()
    axiosMock.put.mockReset()
  })

  it('reads key lists directly from the api data array', async () => {
    axiosMock.get.mockResolvedValueOnce({
      data: {
        code: 0,
        data: [
          {
            id: 1,
            kind: selectionConfigurationKeyKindQuestionKey,
            code: 'rear_axle',
            display_label: '后轴规格',
            description: '后轴相关问题',
            is_enabled: true,
            sort_order: 10,
          },
        ],
      },
    })

    await expect(selectionConfigurationKeyApi.listKeys(selectionConfigurationKeyKindQuestionKey, false)).resolves.toEqual([
      {
        id: 1,
        kind: selectionConfigurationKeyKindQuestionKey,
        code: 'rear_axle',
        display_label: '后轴规格',
        description: '后轴相关问题',
        is_enabled: true,
        sort_order: 10,
      },
    ])

    expect(axiosMock.get).toHaveBeenCalledWith('/api/admin/selection-configuration/keys?kind=question_key&include_disabled=false')
  })

  it('reads enabled key options directly from the api data array', async () => {
    axiosMock.get.mockResolvedValueOnce({
      data: {
        code: 0,
        data: [
          {
            id: 2,
            code: 'rear_axle',
            display_label: '后轴规格',
          },
        ],
      },
    })

    await expect(selectionConfigurationKeyApi.listEnabledKeyOptions(selectionConfigurationKeyKindAnswerKey)).resolves.toEqual([
      {
        id: 2,
        code: 'rear_axle',
        display_label: '后轴规格',
      },
    ])

    expect(axiosMock.get).toHaveBeenCalledWith('/api/admin/selection-configuration/keys/options?kind=answer_key')
  })

  it('reads created keys from the api data object', async () => {
    const payload = {
      kind: selectionConfigurationKeyKindQuestionKey,
      code: 'rear_axle',
      display_label: '后轴规格',
      description: '后轴相关问题',
      is_enabled: true,
      sort_order: 10,
    }

    axiosMock.post.mockResolvedValueOnce({
      data: {
        code: 0,
        data: {
          id: 3,
          ...payload,
        },
      },
    })

    await expect(selectionConfigurationKeyApi.createKey(payload)).resolves.toEqual({
      id: 3,
      ...payload,
    })

    expect(axiosMock.post).toHaveBeenCalledWith('/api/admin/selection-configuration/keys', payload)
  })

  it('reads updated keys from the api data object', async () => {
    const payload = {
      kind: selectionConfigurationKeyKindAnswerKey,
      code: 'front_axle',
      display_label: '前轴规格',
      description: '前轴相关问题',
      is_enabled: false,
      sort_order: 20,
    }

    axiosMock.put.mockResolvedValueOnce({
      data: {
        code: 0,
        data: {
          id: 4,
          ...payload,
        },
      },
    })

    await expect(selectionConfigurationKeyApi.updateKey(4, payload)).resolves.toEqual({
      id: 4,
      ...payload,
    })

    expect(axiosMock.put).toHaveBeenCalledWith('/api/admin/selection-configuration/keys/4', payload)
  })
})
