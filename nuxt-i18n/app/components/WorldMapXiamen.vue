<template>
  <section
    class="mt-4 rounded-2xl bg-slate-950/60 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
  >
    <div class="mb-3 flex flex-col gap-1">
      <h3 class="text-sm font-semibold text-slate-100">
        Global partnership network  Xiamen origin
      </h3>
      <p class="text-xs tz-text-secondary">
        All routes start from Xiamen, Fujian, China and connect to current demo
        cities.
      </p>
    </div>

    <svg
      ref="svgRef"
      id="worldMap"
      class="h-auto w-full rounded-2xl bg-gradient-to-b from-slate-950 via-slate-950 to-black shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
      viewBox="0 0 800 400"
      xmlns="http://www.w3.org/2000/svg"
      aria-labelledby="title desc"
      role="img"
    >
      <title id="title">Global partnership routes from Xiamen</title>
      <desc id="desc">
        Animated arcs showing connections from Xiamen, Fujian, China to several
        global cities.
      </desc>

      <defs>
        <!-- glow for path strokes and points -->
        <filter id="glow">
          <feGaussianBlur stdDeviation="2" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        <!-- soft radial highlight behind Asia region as a hint -->
        <radialGradient id="asiaHighlight" cx="65%" cy="50" r="0.35">
          <stop offset="0%" stop-color="#22c55e" stop-opacity="0.25" />
          <stop offset="100%" stop-color="#22c55e" stop-opacity="0" />
        </radialGradient>
      </defs>

      <!-- base dark background -->
      <rect width="800" height="400" fill="#020617" />
      <!-- world map background image from public/ -->
      <image
        href="/company/globalpartners/tanzanite-mapchart.webp"
        x="0"
        y="0"
        width="800"
        height="400"
        preserveAspectRatio="none"
        opacity="0.9"
      />
      <!-- dark overlay to blend image into card background -->
      <rect width="800" height="400" fill="#020617" opacity="0.35" />
      <!-- asia highlight overlay -->
      <rect width="800" height="400" fill="url(#asiaHighlight)" opacity="0.9" />

      <g ref="pathsLayerRef"></g>
      <g ref="pointsLayerRef"></g>
    </svg>

    <p class="mt-2 text-[11px] tz-text-muted">
      Visualization is illustrative only. Routes are stylized connections from
      Xiamen to selected partner cities.
    </p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

const svgRef = ref<SVGSVGElement | null>(null)
const pathsLayerRef = ref<SVGGElement | null>(null)
const pointsLayerRef = ref<SVGGElement | null>(null)

onMounted(() => {
  const svgNS = 'http://www.w3.org/2000/svg'

  const origin = {
    name: 'Xiamen',
    lat: 24.4798,
    lng: 118.0895,
  }

  const destinations = [
    { name: 'Los Angeles', lat: 34.0522, lng: -118.2437 },
    { name: 'Brasilia', lat: -15.7975, lng: -47.8919 },
    // For Lisbon & London we pin them directly to screen coordinates that
    // visually align with the labels on this static map image.
    // (These numbers were chosen by trial to sit on land, not ocean.)
    { name: 'Lisbon', lat: 38.7223, lng: -9.1393, screen: { x: 440, y: 185 } },
    { name: 'London', lat: 51.5074, lng: -0.1278, screen: { x: 465, y: 145 } },
    { name: 'New Delhi', lat: 28.6139, lng: 77.209 },
    { name: 'Vladivostok', lat: 43.1332, lng: 131.9113 },
    { name: 'Nairobi', lat: -1.2921, lng: 36.8219 },
  ]

  const viewBox = { width: 800, height: 400 }

  // Approximate geographic bounds of the background map image.
  // Many world-map images crop off the extreme poles, so we use
  // slightly narrower latitude range instead of [-90, 90].
  const mapBounds = {
    minLat: -60, // bottom of map image ~60°S
    maxLat: 85, // top of map image ~85°N
    minLng: -180,
    maxLng: 180,
  }

  function project(lat: number, lng: number) {
    const xNorm = (lng - mapBounds.minLng) / (mapBounds.maxLng - mapBounds.minLng)
    const yNorm = (mapBounds.maxLat - lat) / (mapBounds.maxLat - mapBounds.minLat)
    const x = xNorm * viewBox.width
    const y = yNorm * viewBox.height
    return { x, y }
  }

  function createCurvedPath(
    start: { x: number; y: number },
    end: { x: number; y: number },
  ) {
    const midX = (start.x + end.x) / 2
    const midY = Math.min(start.y, end.y) - 60 // raise a bit for nicer arc
    return `M ${start.x} ${start.y} Q ${midX} ${midY} ${end.x} ${end.y}`
  }

  function createCircleGroup(x: number, y: number, isOrigin: boolean) {
    const g = document.createElementNS(svgNS, 'g')
    g.setAttribute('transform', `translate(${x}, ${y})`)

    const core = document.createElementNS(svgNS, 'circle')
    core.setAttribute('class', 'point-core')
    core.setAttribute('r', isOrigin ? '4' : '3')
    core.setAttribute('cx', '0')
    core.setAttribute('cy', '0')

    const pulse = document.createElementNS(svgNS, 'circle')
    pulse.setAttribute('class', isOrigin ? 'point-pulse point-pulse--slow' : 'point-pulse')
    pulse.setAttribute('r', isOrigin ? '4' : '3')
    pulse.setAttribute('cx', '0')
    pulse.setAttribute('cy', '0')

    g.appendChild(pulse)
    g.appendChild(core)

    return g
  }

  function createLabel(
    x: number,
    y: number,
    text: string,
    options?: { origin?: boolean; align?: 'left' | 'right' },
  ) {
    const label = document.createElementNS(svgNS, 'text')
    label.textContent = text
    label.setAttribute(
      'class',
      `map-label ${options && options.origin ? 'map-label--origin' : ''} ${
        options && options.align === 'left' ? 'map-label--left' : 'map-label--right'
      }`,
    )

    const dx = options && options.align === 'left' ? -10 : 10
    const dy = options && options.origin ? -10 : -6

    label.setAttribute('x', String(x + dx))
    label.setAttribute('y', String(y + dy))

    return label
  }

  const svg = svgRef.value
  const pathsLayer = pathsLayerRef.value
  const pointsLayer = pointsLayerRef.value
  if (!svg || !pathsLayer || !pointsLayer) return

  const originPt = project(origin.lat, origin.lng)

  // Draw origin point + label
  const originGroup = createCircleGroup(originPt.x, originPt.y, true)
  const originLabel = createLabel(originPt.x, originPt.y, origin.name, {
    origin: true,
    align: 'left',
  })

  pointsLayer.appendChild(originGroup)
  pointsLayer.appendChild(originLabel)

  const baseDuration = 10 // seconds, synced with CSS keyframes

  destinations.forEach((dest, index) => {
    const projected = project(dest.lat, dest.lng)

    const destPt = (dest as any).screen
      ? { x: (dest as any).screen.x, y: (dest as any).screen.y }
      : projected

    // Path
    const path = document.createElementNS(svgNS, 'path')
    path.setAttribute('class', 'path-arc')
    path.setAttribute('d', createCurvedPath(originPt, destPt))

    pathsLayer.appendChild(path)

    // compute length for dash animation
    const length = path.getTotalLength()
    path.style.setProperty('--path-len', String(length))
    path.style.strokeDasharray = String(length)
    path.style.strokeDashoffset = String(length)

    // fine-tune stagger based on index
    const delay = index * 0.4
    path.style.animationDelay = `${delay}s`
    path.style.animationDuration = `${baseDuration}s`

    // Destination points and labels
    const destGroup = createCircleGroup(destPt.x, destPt.y, false)
    const align = dest.lng < 0 ? 'left' : 'right'
    const destLabel = createLabel(destPt.x, destPt.y, dest.name, { align })

    pointsLayer.appendChild(destGroup)
    pointsLayer.appendChild(destLabel)
  })
})
</script>

<style>
.map-label {
  font-size: 10px;
  fill: #e5e7eb;
  paint-order: stroke;
  stroke: #020617;
  stroke-width: 2;
}

.map-label--origin {
  font-weight: 700;
  fill: #a5b4fc;
}

.map-label--right {
  text-anchor: start;
}

.map-label--left {
  text-anchor: end;
}

.path-arc {
  fill: none;
  stroke: #38bdf8;
  stroke-width: 1.4;
  stroke-linecap: round;
  filter: url(#glow);
  opacity: 0.2;
  stroke-dasharray: var(--path-len, 1);
  stroke-dashoffset: var(--path-len, 1);
  animation: drawLine 10s ease-in-out infinite;
}

.path-arc:nth-of-type(1) {
  animation-delay: 0s;
}
.path-arc:nth-of-type(2) {
  animation-delay: 0.4s;
}
.path-arc:nth-of-type(3) {
  animation-delay: 0.8s;
}
.path-arc:nth-of-type(4) {
  animation-delay: 1.2s;
}
.path-arc:nth-of-type(5) {
  animation-delay: 1.6s;
}
.path-arc:nth-of-type(6) {
  animation-delay: 2s;
}
.path-arc:nth-of-type(7) {
  animation-delay: 2.4s;
}

@keyframes drawLine {
  0% {
    stroke-dashoffset: var(--path-len);
    opacity: 0;
  }
  18% {
    opacity: 1;
  }
  55% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
  100% {
    stroke-dashoffset: 0;
    opacity: 0.35;
  }
}

.point-core {
  fill: #38bdf8;
}

.point-pulse {
  fill: #38bdf8;
  opacity: 0.55;
  animation: pulse 2.4s ease-out infinite;
}

.point-pulse--slow {
  animation-duration: 3s;
  animation-delay: 0.6s;
}

@keyframes pulse {
  0% {
    r: 3;
    opacity: 0.6;
  }
  100% {
    r: 14;
    opacity: 0;
  }
}
</style>
