import { mkdir, writeFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import {
  resolveWheelsetLacingDisplayGeometryTopologySelection,
  validateWheelsetLacingDisplayGeometryResponse,
} from '../../app/utils/wheelsetLacingDisplayGeometryContract'

// Regenerate from the running simulator backend with:
// npx tsx scripts/company/generate-static-rim-drilling-direction-illustrations.ts
// RIM_DRILLING_ILLUSTRATION_BACKEND_ORIGIN overrides the local API origin.

interface StaticDrillingIllustrationGeometryPoint {
  id: number
  side: 'A' | 'B'
  x: number
  y: number
}

interface StaticDrillingIllustrationGeometrySpoke {
  hub: StaticDrillingIllustrationGeometryPoint
  rim: StaticDrillingIllustrationGeometryPoint
  side: 'A' | 'B'
}

interface StaticLacingReferenceSpokeCrossing {
  distanceFractionFromReferenceHub: number
  x: number
  y: number
}

interface StaticDrillingIllustrationBackendGeometry {
  spokes: StaticDrillingIllustrationGeometrySpoke[]
  rim_holes: StaticDrillingIllustrationGeometryPoint[]
  hub_holes_a: StaticDrillingIllustrationGeometryPoint[]
  hub_holes_b: StaticDrillingIllustrationGeometryPoint[]
  flange_profile: {
    flange_offset_a_mm: number
    flange_offset_b_mm: number
  }
}

const staticDrillingIllustrationBackendOrigin = process.env.RIM_DRILLING_ILLUSTRATION_BACKEND_ORIGIN ?? 'http://127.0.0.1:10241'
const staticDrillingIllustrationOutputDirectory = new URL('../../public/company/ourstory/holepatterns/', import.meta.url)
const staticDrillingIllustrationHighlightedSpokeColors = ['#047857', '#0369a1'] as const
const staticDrillingIllustrationExamples = [
  { crossCount: 0, filenameLabel: 'radial' },
  { crossCount: 2, filenameLabel: 'two-cross' },
  { crossCount: 3, filenameLabel: 'three-cross' },
] as const

function formatStaticDrillingIllustrationCoordinate(coordinate: number): string {
  return coordinate.toFixed(2)
}

function createStaticDrillingIllustrationSvgDocument(width: number, height: number, content: string): string {
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
  <!-- Schematic direction projection, not production geometry. Generated from the lacing simulator backend. -->
  <defs>
    <pattern id="rim-wall-hatching" width="8" height="8" patternUnits="userSpaceOnUse"><path d="M-2 8 8-2M6 10 10 6" stroke="#cbd5e1" stroke-width="1"/></pattern>
  </defs>
  <rect width="${width}" height="${height}" fill="#f8fafc"/>
${content.replace(/[ \t]+$/gm, '')}
</svg>
`
}

function getStaticLongitudinalDrillingExampleSpokesFromOppositeFlanges(geometry: StaticDrillingIllustrationBackendGeometry): StaticDrillingIllustrationGeometrySpoke[] {
  // Local drilling details compare one seat from each flange, using the axial view's side colors.
  return [0, 3].map((rimHoleId) => {
    const spoke = geometry.spokes.find((candidate) => candidate.rim.id === rimHoleId)
    if (!spoke) throw new Error(`Missing reference spoke for rim hole ${rimHoleId}`)
    const expectedFlangeSide = rimHoleId === 0 ? 'A' : 'B'
    if (spoke.side !== expectedFlangeSide) throw new Error(`Reference rim hole ${rimHoleId} must belong to flange ${expectedFlangeSide}`)
    return spoke
  })
}

function findSameFlangeCrossingsAlongStaticLacingReferenceSpoke(
  sameFlangeSpokes: StaticDrillingIllustrationGeometrySpoke[],
  referenceSpoke: StaticDrillingIllustrationGeometrySpoke,
): StaticLacingReferenceSpokeCrossing[] {
  const referenceDirectionX = referenceSpoke.rim.x - referenceSpoke.hub.x
  const referenceDirectionY = referenceSpoke.rim.y - referenceSpoke.hub.y
  const crossings: StaticLacingReferenceSpokeCrossing[] = []

  for (const otherSpoke of sameFlangeSpokes) {
    if (otherSpoke === referenceSpoke) continue
    const otherDirectionX = otherSpoke.rim.x - otherSpoke.hub.x
    const otherDirectionY = otherSpoke.rim.y - otherSpoke.hub.y
    const intersectionDenominator = referenceDirectionX * otherDirectionY - referenceDirectionY * otherDirectionX
    if (Math.abs(intersectionDenominator) < 1e-8) continue

    const hubSeparationX = otherSpoke.hub.x - referenceSpoke.hub.x
    const hubSeparationY = otherSpoke.hub.y - referenceSpoke.hub.y
    const distanceFractionFromReferenceHub = (hubSeparationX * otherDirectionY - hubSeparationY * otherDirectionX) / intersectionDenominator
    const distanceFractionFromOtherHub = (hubSeparationX * referenceDirectionY - hubSeparationY * referenceDirectionX) / intersectionDenominator
    // Count intersections inside both spoke segments, excluding shared endpoints and line extensions.
    if (distanceFractionFromReferenceHub <= 1e-6 || distanceFractionFromReferenceHub >= 1 - 1e-6
      || distanceFractionFromOtherHub <= 1e-6 || distanceFractionFromOtherHub >= 1 - 1e-6) continue

    crossings.push({
      distanceFractionFromReferenceHub,
      x: referenceSpoke.hub.x + distanceFractionFromReferenceHub * referenceDirectionX,
      y: referenceSpoke.hub.y + distanceFractionFromReferenceHub * referenceDirectionY,
    })
  }

  return crossings.sort((firstCrossing, secondCrossing) => firstCrossing.distanceFractionFromReferenceHub - secondCrossing.distanceFractionFromReferenceHub)
}

function createStaticSingleFlangeLacingWithCountedCrossingsSvg(geometry: StaticDrillingIllustrationBackendGeometry, expectedCrossCount: number): string {
  const sameFlangeSpokes = geometry.spokes.filter((spoke) => spoke.side === 'A')
  const referenceSpoke = sameFlangeSpokes.find((spoke) => spoke.rim.id === 0)
  if (sameFlangeSpokes.length !== 16 || !referenceSpoke) throw new Error('The 32-hole lacing illustration requires 16 spokes on flange A and a reference at rim hole 0')
  const crossings = findSameFlangeCrossingsAlongStaticLacingReferenceSpoke(sameFlangeSpokes, referenceSpoke)
  if (crossings.length !== expectedCrossCount) throw new Error(`Expected ${expectedCrossCount} same-flange crossings, found ${crossings.length}`)

  const backgroundSpokes = sameFlangeSpokes.filter((spoke) => spoke !== referenceSpoke).map((spoke) => (
    `<line data-lacing-spoke="same-flange" x1="${spoke.hub.x}" y1="${spoke.hub.y}" x2="${spoke.rim.x}" y2="${spoke.rim.y}" stroke="#94a3b8" stroke-width="2"/>`
  )).join('\n')
  const rimHoles = geometry.rim_holes.map((hole) => (
    `<circle cx="${hole.x}" cy="${hole.y}" r="2.6" fill="${hole.side === 'A' ? '#64748b' : '#cbd5e1'}"/>`
  )).join('\n')
  const flangeHoles = geometry.hub_holes_a.map((hole) => (
    `<circle cx="${hole.x}" cy="${hole.y}" r="2.5" fill="#64748b"/>`
  )).join('\n')
  const referenceSpokeLength = Math.hypot(referenceSpoke.rim.x - referenceSpoke.hub.x, referenceSpoke.rim.y - referenceSpoke.hub.y)
  const labelOffsetDirectionX = -(referenceSpoke.rim.y - referenceSpoke.hub.y) / referenceSpokeLength
  const labelOffsetDirectionY = (referenceSpoke.rim.x - referenceSpoke.hub.x) / referenceSpokeLength
  const numberedCrossings = crossings.map((crossing, index) => {
    // Stagger the first two labels because their actual intersections are close to the flange.
    const labelOffset = index === 1 ? 72 : 36
    const labelCenterX = formatStaticDrillingIllustrationCoordinate(crossing.x + labelOffsetDirectionX * labelOffset)
    const labelCenterY = formatStaticDrillingIllustrationCoordinate(crossing.y + labelOffsetDirectionY * labelOffset)
    return `<g data-reference-spoke-crossing="${index + 1}">
      <line x1="${crossing.x}" y1="${crossing.y}" x2="${labelCenterX}" y2="${labelCenterY}" stroke="#047857" stroke-width="1.5"/>
      <circle cx="${crossing.x}" cy="${crossing.y}" r="4.5" fill="#f8fafc" stroke="#047857" stroke-width="2"/>
      <circle cx="${labelCenterX}" cy="${labelCenterY}" r="18" fill="#ffffff" stroke="#047857" stroke-width="1.5"/>
      <text x="${labelCenterX}" y="${labelCenterY}" dy="0.35em" text-anchor="middle" font-family="Arial, sans-serif" font-size="24" font-weight="600" fill="#047857">${index + 1}</text>
    </g>`
  }).join('\n')
  const flangeRadius = Math.hypot(geometry.hub_holes_a[0]!.x, geometry.hub_holes_a[0]!.y)
  return createStaticDrillingIllustrationSvgDocument(560, 560, `  <g data-single-flange-spoke-count="16" data-reference-cross-count="${expectedCrossCount}" transform="translate(280 280) scale(0.98)">
    <circle r="246" fill="#e2e8f0" stroke="#64748b" stroke-width="1.5"/>
    <circle r="220" fill="#f8fafc" stroke="#94a3b8" stroke-width="1.5"/>
    <circle r="${formatStaticDrillingIllustrationCoordinate(flangeRadius)}" fill="#e2e8f0" stroke="#64748b" stroke-width="1.5"/>
    ${backgroundSpokes}
    ${flangeHoles}
    ${rimHoles}
    <line data-lacing-spoke="reference" x1="${referenceSpoke.hub.x}" y1="${referenceSpoke.hub.y}" x2="${referenceSpoke.rim.x}" y2="${referenceSpoke.rim.y}" stroke="#047857" stroke-width="4.5"/>
    <circle cx="${referenceSpoke.hub.x}" cy="${referenceSpoke.hub.y}" r="4" fill="#047857"/>
    <circle cx="${referenceSpoke.rim.x}" cy="${referenceSpoke.rim.y}" r="4.5" fill="#047857"/>
    <circle r="13" fill="#f8fafc" stroke="#64748b" stroke-width="2"/>
    <path d="M-7 0H7M0-7V7" stroke="#94a3b8" stroke-width="1"/>
    ${numberedCrossings}
  </g>`)
}

function getStaticLongitudinalDrillingAxisSlope(spoke: StaticDrillingIllustrationGeometrySpoke): number {
  const rimRadius = Math.hypot(spoke.rim.x, spoke.rim.y)
  const radialUnitX = spoke.rim.x / rimRadius
  const radialUnitY = spoke.rim.y / rimRadius
  const inwardSpokeX = spoke.hub.x - spoke.rim.x
  const inwardSpokeY = spoke.hub.y - spoke.rim.y
  const tangentialComponent = -inwardSpokeX * radialUnitY + inwardSpokeY * radialUnitX
  const inwardRadialComponent = -inwardSpokeX * radialUnitX - inwardSpokeY * radialUnitY
  return tangentialComponent / inwardRadialComponent
}

function createStaticLongitudinalRimHoleDrillingDetailSvg(geometry: StaticDrillingIllustrationBackendGeometry): string {
  const holeDetails = getStaticLongitudinalDrillingExampleSpokesFromOppositeFlanges(geometry).map((spoke, index) => {
    const holeCenterX = index === 0 ? 110 : 310
    const holeCenterY = 80
    const longitudinalSlope = getStaticLongitudinalDrillingAxisSlope(spoke)
    const directionColor = staticDrillingIllustrationHighlightedSpokeColors[index]
    const getLongitudinalHoleCenterXAtHeight = (height: number) => Number(formatStaticDrillingIllustrationCoordinate(holeCenterX + longitudinalSlope * (height - holeCenterY)))
    return `<path d="M${holeCenterX - 75} 60h150v38h-150Z" fill="#e2e8f0"/>
    <path d="M${holeCenterX - 75} 60h150v38h-150Z" fill="url(#rim-wall-hatching)" stroke="#94a3b8" stroke-width="1.5"/>
    <line x1="${holeCenterX}" y1="30" x2="${holeCenterX}" y2="177" stroke="#cbd5e1" stroke-width="1" stroke-dasharray="3 4"/>
    <path d="M${getLongitudinalHoleCenterXAtHeight(60)} 60 ${getLongitudinalHoleCenterXAtHeight(98)} 98" stroke="#f8fafc" stroke-width="13"/>
    <path d="M${getLongitudinalHoleCenterXAtHeight(60)} 60 ${getLongitudinalHoleCenterXAtHeight(98)} 98" stroke="${directionColor}" stroke-opacity="0.2" stroke-width="13"/>
    <path d="M${getLongitudinalHoleCenterXAtHeight(60) - 6} 60 ${getLongitudinalHoleCenterXAtHeight(98) - 6} 98M${getLongitudinalHoleCenterXAtHeight(60) + 6} 60 ${getLongitudinalHoleCenterXAtHeight(98) + 6} 98" stroke="${directionColor}" stroke-width="1.5"/>
    <line x1="${getLongitudinalHoleCenterXAtHeight(98)}" y1="98" x2="${getLongitudinalHoleCenterXAtHeight(176)}" y2="176" stroke="${directionColor}" stroke-width="3"/>
    <line x1="${getLongitudinalHoleCenterXAtHeight(25)}" y1="25" x2="${getLongitudinalHoleCenterXAtHeight(114)}" y2="114" stroke="${directionColor}" stroke-width="2" stroke-dasharray="5 4"/>`
  }).join('\n')
  return createStaticDrillingIllustrationSvgDocument(420, 200, `  ${holeDetails}`)
}

function createStaticLeftRightOffsetRimHolesAndHubFlangeAxialDrillingProjectionSvg(geometry: StaticDrillingIllustrationBackendGeometry): string {
  // Both sides are projected together; these are separate rim-hole seats, not a single physical cross-section.
  const hubCenterX = 210
  const flangeOffsetScale = 1.2
  const leftFlangeX = hubCenterX - geometry.flange_profile.flange_offset_b_mm * flangeOffsetScale
  const rightFlangeX = hubCenterX + geometry.flange_profile.flange_offset_a_mm * flangeOffsetScale
  const axialHoleProjections = [
    { holeX: 199, flangeX: leftFlangeX, color: staticDrillingIllustrationHighlightedSpokeColors[1] },
    { holeX: 221, flangeX: rightFlangeX, color: staticDrillingIllustrationHighlightedSpokeColors[0] },
  ].map(({ holeX, flangeX, color }) => {
    const getAxialHoleProjectionXAtHeight = (height: number) => formatStaticDrillingIllustrationCoordinate(holeX + (flangeX - holeX) * (height - 70) / 65)
    return `<path d="M${getAxialHoleProjectionXAtHeight(59)} 59 ${getAxialHoleProjectionXAtHeight(81)} 81" stroke="#f8fafc" stroke-width="10"/>
    <path d="M${getAxialHoleProjectionXAtHeight(59)} 59 ${getAxialHoleProjectionXAtHeight(81)} 81" stroke="${color}" stroke-opacity="0.25" stroke-width="10"/>
    <line x1="${getAxialHoleProjectionXAtHeight(81)}" y1="81" x2="${flangeX}" y2="135" stroke="${color}" stroke-width="2.5"/>
    <line x1="${getAxialHoleProjectionXAtHeight(42)}" y1="42" x2="${getAxialHoleProjectionXAtHeight(97)}" y2="97" stroke="${color}" stroke-width="2" stroke-dasharray="5 4"/>
    <path d="M${flangeX} 112v44" stroke="#64748b" stroke-width="5" stroke-linecap="round"/>
    <circle cx="${flangeX}" cy="135" r="3.5" fill="${color}"/>`
  }).join('\n')
  // Center the projection in a square viewport instead of retaining empty space on both sides.
  return createStaticDrillingIllustrationSvgDocument(200, 200, `  <g transform="translate(-110 0)">
  <path d="M178 24h64v57h-64Z" fill="#e2e8f0" stroke="#94a3b8" stroke-width="1.5"/>
  <path d="M178 24h64v57h-64Z" fill="url(#rim-wall-hatching)"/>
  <path d="M190 36h40v23h-40Z" fill="#f8fafc" stroke="#94a3b8" stroke-width="1"/>
  <line x1="${hubCenterX}" y1="14" x2="${hubCenterX}" y2="181" stroke="#94a3b8" stroke-width="1" stroke-dasharray="4 4"/>
  <path d="M${leftFlangeX - 26} 143H${rightFlangeX + 26}" stroke="#94a3b8" stroke-width="6" stroke-linecap="round"/>
  ${axialHoleProjections}
  <g data-rim-hole-lateral-offset="left" stroke="#0369a1" stroke-width="1.2" fill="none">
    <path d="M160 48h30l9 22"/>
    <circle cx="199" cy="70" r="3" fill="#f8fafc"/>
  </g>
  <g data-rim-hole-lateral-offset="right" stroke="#047857" stroke-width="1.2" fill="none">
    <path d="M260 48h-30l-9 22"/>
    <circle cx="221" cy="70" r="3" fill="#f8fafc"/>
  </g>
  <g font-family="Arial, sans-serif" font-size="12" font-weight="600" text-anchor="middle">
    <text x="151" y="52" fill="#0369a1">L</text>
    <text x="269" y="52" fill="#047857">R</text>
  </g>
  <g stroke="#64748b" stroke-width="1" fill="none">
    <path d="M${leftFlangeX} 162v17m${hubCenterX - leftFlangeX} -17v17m${rightFlangeX - hubCenterX} -17v17M${leftFlangeX} 170H${rightFlangeX}"/>
  </g>
  <g fill="#475569" font-family="Arial, sans-serif" font-size="12" text-anchor="middle">
    <text x="${(leftFlangeX + hubCenterX) / 2}" y="190">WL</text>
    <text x="${(rightFlangeX + hubCenterX) / 2}" y="190">WR</text>
  </g>
  </g>`)
}

async function generateStaticRimDrillingDirectionIllustrationsFromSimulatorGeometry(): Promise<void> {
  const referenceGeometries = await Promise.all(staticDrillingIllustrationExamples.map(async (example) => {
    const topologySelection = resolveWheelsetLacingDisplayGeometryTopologySelection(32, example.crossCount)
    const geometryEndpoint = new URL('/api/v1/wheelset-lacing/display-geometry', staticDrillingIllustrationBackendOrigin)
    geometryEndpoint.searchParams.set('topology_id', topologySelection.topologyId)
    const geometryResponse = await fetch(geometryEndpoint, { signal: AbortSignal.timeout(15000) })
    if (!geometryResponse.ok) throw new Error(`Geometry request failed with HTTP ${geometryResponse.status}`)
    const geometryEnvelope = await geometryResponse.json() as { code: number; data: unknown }
    if (geometryEnvelope.code !== 0) throw new Error(`Geometry API returned code ${geometryEnvelope.code}`)
    validateWheelsetLacingDisplayGeometryResponse(geometryEnvelope.data, topologySelection)
    return { example, geometry: geometryEnvelope.data as StaticDrillingIllustrationBackendGeometry }
  }))

  await mkdir(staticDrillingIllustrationOutputDirectory, { recursive: true })
  for (const { example, geometry } of referenceGeometries) {
    await writeFile(new URL(`32-hole-${example.filenameLabel}-single-flange-lacing-with-counted-crossings.svg`, staticDrillingIllustrationOutputDirectory), createStaticSingleFlangeLacingWithCountedCrossingsSvg(geometry, example.crossCount))
    await writeFile(new URL(`${example.filenameLabel}-lacing-longitudinal-rim-hole-drilling-detail.svg`, staticDrillingIllustrationOutputDirectory), createStaticLongitudinalRimHoleDrillingDetailSvg(geometry))
  }
  await writeFile(new URL('left-right-offset-rim-holes-and-hub-flange-axial-drilling-projection.svg', staticDrillingIllustrationOutputDirectory), createStaticLeftRightOffsetRimHolesAndHubFlangeAxialDrillingProjectionSvg(referenceGeometries[1]!.geometry))
  console.log(`Generated seven static drilling illustrations in ${fileURLToPath(staticDrillingIllustrationOutputDirectory)}`)
}

await generateStaticRimDrillingDirectionIllustrationsFromSimulatorGeometry()
