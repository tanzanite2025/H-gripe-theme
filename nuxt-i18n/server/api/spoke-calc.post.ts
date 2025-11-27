import { defineEventHandler, readBody, createError } from 'h3'
import type { RimGeometry, HubGeometry, SpokeCalcInput, SpokeCalcResult } from '~/types/spoke'

interface SpokeCalcDebug {
  rim: RimGeometry
  hub: HubGeometry
  formulaVersion: string
}

// API response shape used by the UI. You can extend this later.
export type SpokeCalcApiResponse = SpokeCalcResult & {
  debug: SpokeCalcDebug
}

type ComputeInput = {
  rim: RimGeometry
  hub: HubGeometry
  crossing: number
  spokeCount: number
}

export default defineEventHandler(async (event) => {
  const body = await readBody<SpokeCalcInput>(event)

  if (!body?.rimId || !body?.hubId) {
    throw createError({ statusCode: 400, statusMessage: 'rimId and hubId are required' })
  }

  const rim = await getRimGeometry(body.rimId)
  const hub = await getHubGeometry(body.hubId, body.wheelPosition)

  if (!rim || !hub) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Unknown rim or hub geometry',
    })
  }

  // Optional: warn if requested spoke count does not match geometry.
  if (rim.spokeHoles !== body.spokeCount || hub.spokeHoles !== body.spokeCount) {
    // For now just log; later you can decide whether to treat this as an error.
    // eslint-disable-next-line no-console
    console.warn('Spoke count mismatch between selection and geometry', {
      requested: body.spokeCount,
      rim: rim.spokeHoles,
      hub: hub.spokeHoles,
    })
  }

  const { left, right } = computeSpokeLengths({
    rim,
    hub,
    crossing: body.crossing,
    spokeCount: body.spokeCount,
  })

  const response: SpokeCalcApiResponse = {
    leftLengthMm: left,
    rightLengthMm: right,
    debug: {
      rim,
      hub,
      formulaVersion: 'v0.1-mock',
    },
  }

  return response
})

// Placeholder geometry-based formula.
// TODO: replace with the verified formula from your Excel sheet.
function computeSpokeLengths(input: ComputeInput): { left: number; right: number } {
  const { rim, hub, crossing } = input

  const baseRadius = rim.erd / 2
  const crossingFactor = 1 + crossing * 0.06

  const left = baseRadius - hub.leftFlangeToCenterMm * 0.6
  const right = baseRadius - hub.rightFlangeToCenterMm * 0.6

  return {
    left: Number((left * crossingFactor).toFixed(1)),
    right: Number((right * crossingFactor * 1.01).toFixed(1)),
  }
}

// Temporary mock geometry lookup. Later this should read from
// tanzanite-setting / WordPress (meta fields) or a dedicated geometry table.
async function getRimGeometry(id: string): Promise<RimGeometry | null> {
  const mockRims: RimGeometry[] = [
    {
      id: 'demo-rim-700c-32h',
      brand: 'Demo',
      model: '700C Alloy Rim',
      erd: 602,
      spokeHoles: 32,
    },
  ]

  return mockRims.find((r) => r.id === id) ?? null
}

async function getHubGeometry(id: string, _position: 'front' | 'rear'): Promise<HubGeometry | null> {
  const mockHubs: HubGeometry[] = [
    {
      id: 'demo-hub-front-32h',
      brand: 'Demo',
      model: 'Front Disc Hub',
      spokeHoles: 32,
      type: 'front',
      brakeType: 'disc',
      axleWidthMm: 100,
      leftFlangePcdMm: 45,
      rightFlangePcdMm: 45,
      leftFlangeToCenterMm: 35,
      rightFlangeToCenterMm: 35,
    },
    {
      id: 'demo-hub-rear-32h',
      brand: 'Demo',
      model: 'Rear Disc Hub',
      spokeHoles: 32,
      type: 'rear',
      brakeType: 'disc',
      axleWidthMm: 142,
      leftFlangePcdMm: 55,
      rightFlangePcdMm: 55,
      leftFlangeToCenterMm: 33,
      rightFlangeToCenterMm: 20,
    },
  ]

  return mockHubs.find((h) => h.id === id) ?? null
}
