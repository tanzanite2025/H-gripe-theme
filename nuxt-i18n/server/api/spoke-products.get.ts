import { defineEventHandler } from 'h3'

interface RimOptionDto {
  id: string
  label: string
  spokeHoles: number
}

interface HubOptionDto {
  id: string
  label: string
  position: 'front' | 'rear' | 'front-rear-compatible'
  spokeHoles: number
}

interface NippleOptionDto {
  id: string
  label: string
}

export interface SpokeProductsResponse {
  rims: RimOptionDto[]
  hubs: HubOptionDto[]
  nipples: NippleOptionDto[]
}

// Temporary mock product list for the spoke calculator dropdowns.
// Later this should be replaced with data loaded from tanzanite-setting /
// WordPress product meta and categories (e.g. rim / hub / nipple).
export default defineEventHandler((): SpokeProductsResponse => {
  const rims: RimOptionDto[] = [
    {
      id: 'demo-rim-700c-32h',
      label: 'Demo 700C Alloy Rim (32H)',
      spokeHoles: 32,
    },
  ]

  const hubs: HubOptionDto[] = [
    {
      id: 'demo-hub-front-32h',
      label: 'Demo Front Disc Hub (32H)',
      position: 'front',
      spokeHoles: 32,
    },
    {
      id: 'demo-hub-rear-32h',
      label: 'Demo Rear Disc Hub (32H)',
      position: 'rear',
      spokeHoles: 32,
    },
  ]

  const nipples: NippleOptionDto[] = []

  return { rims, hubs, nipples }
})
