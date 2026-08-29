<template>
  <div class="match-guide space-y-8">
    <section class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-md md:p-6">
      <h2 class="mb-4 flex items-center justify-center gap-2 text-xl font-bold tz-text-secondary">
        Tire and Frame Clearance Guide
      </h2>
      <p class="mx-auto mb-6 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        With our particularly wide tires, the question often arises as to whether the tires will still fit in the frame.
        Please understand that due to the large number of bicycle models, we cannot check all frames for compatibility with
        the various tires. The tables below provide reference tire widths and diameters shown in the Schwalbe charts.
        Compare these measurements with the available clearance in your frame, fork and stays before installation.
        Actual dimensions can vary with rim inner width, tire pressure, load and the tire/rim combination.
      </p>

      <section class="match-guide__measurement-reference" aria-labelledby="measurement-reference-title">
        <div class="mx-auto w-full max-w-3xl">
          <GuideImage
            class="overflow-hidden rounded-xl shadow-md"
            src="/public/tiresizecharts/match/schwalbe-tire-fit-frame3.webp"
            alt="Diagram showing tire frame clearance measurements A maximum width, B maximum diameter and C shoulder diameter"
            caption="Measurement diagram for dimensions A, B and C"
            :zoomOnClick="true"
          />
        </div>
        <div class="mt-5">
          <h3 id="measurement-reference-title" class="match-guide__measurement-title">
            Understand dimensions A, B and C before checking frame clearance
          </h3>
          <p class="mx-auto mt-2 max-w-3xl text-sm leading-relaxed tz-text-secondary">
            A is the horizontal maximum width. Schwalbe reports B and C as diameter values: maximum diameter and shoulder
            diameter at maximum width. The diagram uses vertical arrows to show the corresponding upper measurement
            points from the rim shoulder reference.
          </p>
          <p class="mx-auto mt-3 max-w-3xl text-sm leading-relaxed tz-text-secondary">
            Important: B and C are not the one-sided height from the rim edge. To estimate that height above the rim
            shoulder, subtract the rim bead seat diameter from the reported diameter and divide the result by two.
          </p>
        </div>
        <div class="match-guide__measurement-grid">
          <div class="match-guide__measurement-item">
            <span class="match-guide__measurement-badge">A</span>
            <div>
              <h4>Maximum tire width</h4>
              <p>
                The widest outside-to-outside measurement across the mounted tire, measured from outside lug to outside lug
                at maximum air pressure. Use A to check side-to-side clearance between the tire and the frame, fork or stays.
              </p>
            </div>
          </div>
          <div class="match-guide__measurement-item">
            <span class="match-guide__measurement-badge">B</span>
            <div>
              <h4>Maximum outer diameter</h4>
              <p>
                Schwalbe's maximum-diameter value for the mounted tire and wheel. It is the full outside diameter across
                the highest point of the inflated tread or tread lug. In the diagram, the upper arrow endpoint identifies
                that highest point. Use B to check the largest overall radial envelope. For the one-sided height above
                the rim shoulder, use (B - bead seat diameter) / 2.
              </p>
            </div>
          </div>
          <div class="match-guide__measurement-item">
            <span class="match-guide__measurement-badge">C</span>
            <div>
              <h4>Shoulder diameter at maximum width</h4>
              <p>
                Schwalbe's shoulder diameter measured at the point where the tire reaches its maximum width. It is the
                full diameter across that shoulder level, usually near the outer shoulder lugs on a knobby tire, not the
                height to the lug's starting edge. For the one-sided height above the rim shoulder, use (C - bead seat
                diameter) / 2.
              </p>
            </div>
          </div>
        </div>
      </section>

      <div class="match-guide__table-list">
        <section
          v-for="chart in tireFrameClearanceTables"
          :key="chart.key"
          class="match-guide__table-section"
          :aria-labelledby="`${chart.key}-title`"
        >
          <h3 :id="`${chart.key}-title`" class="match-guide__table-title">
            {{ chart.title }}
          </h3>
          <div class="match-guide__table-scroll">
            <table class="match-guide__table">
              <caption>
                {{ chart.title }} with maximum tire width, maximum tire diameter and shoulder diameter at maximum width,
                measured in millimeters
              </caption>
              <thead>
                <tr>
                  <th scope="col">Wheel size</th>
                  <th scope="col">ETRTO size</th>
                  <th scope="col">Tire model</th>
                  <th scope="col">Maximum width (A)</th>
                  <th scope="col">Maximum diameter (B)</th>
                  <th scope="col">Shoulder diameter at maximum width (C)</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in chart.rows" :key="`${chart.key}-${row.inch}-${row.etrto}-${row.tire}`">
                  <td class="match-guide__dimension-cell">{{ row.inch }}</td>
                  <td class="match-guide__dimension-cell">{{ row.etrto }}</td>
                  <th scope="row">{{ row.tire }}</th>
                  <td class="match-guide__dimension-cell">{{ row.maxWidthMm }} mm</td>
                  <td class="match-guide__dimension-cell">{{ row.maxDiameterMm }} mm</td>
                  <td class="match-guide__dimension-cell">{{ row.shoulderDiameterMm }} mm</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <div class="mt-8 border-t tz-border-subtle pt-6">
        <h3 class="match-guide__table-title">Source chart images</h3>
        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
          <GuideImage
            class="overflow-hidden rounded-xl shadow-md"
            src="/public/tiresizecharts/match/schwalbe-tire-fit-frame1.webp"
            alt="Schwalbe tire and frame clearance chart for 27.5, 28 and 29 inch tires"
            caption="Reference chart: 27.5, 28 and 29 inch tires"
            :zoomOnClick="true"
          />
          <GuideImage
            class="overflow-hidden rounded-xl shadow-md"
            src="/public/tiresizecharts/match/schwalbe-tire-fit-frame2.webp"
            alt="Schwalbe tire and frame clearance chart for 24 and 26 inch tires"
            caption="Reference chart: 24 and 26 inch tires"
            :zoomOnClick="true"
          />
        </div>
      </div>
    </section>

    <section class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-md md:p-6">
      <h2 class="mb-4 flex items-center justify-center gap-2 text-lg font-bold uppercase tracking-wide tz-text-secondary">
        Exact Circumference Guide
      </h2>
      <p class="mx-auto mb-6 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        Exact tire circumferences are often required for precise programming of the bike computer. The wheel circumference
        varies depending on the inner rim width, puncture protection in the tire, air pressure and weight load. The values
        below are approximate reference values, not a substitute for measuring the mounted tire.
      </p>
      <p class="mx-auto mb-8 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        For precise programming of a wheel computer, we recommend a simple rolling test with the rider in the saddle:
        align the valve from the front wheel at the bottom 6 o'clock position, mark the floor, roll the bike forward in
        as straight a line as possible until the valve returns to the 6 o'clock position, and measure the distance between
        the two marks.
      </p>

      <div class="match-guide__table-scroll text-left">
        <table class="match-guide__table match-guide__table--circumference">
          <caption>
            Approximate bicycle tire wheel circumference by inch size and ETRTO size, measured in millimeters
          </caption>
          <thead>
            <tr>
              <th scope="col">Wheel size</th>
              <th scope="col">ETRTO size</th>
              <th scope="col">Approximate wheel circumference</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in tireCircumferenceRows" :key="`${row.inch}-${row.etrto}`">
              <td class="match-guide__dimension-cell">{{ row.inch }}</td>
              <th scope="row">{{ row.etrto }}</th>
              <td class="match-guide__dimension-cell">{{ row.circumferenceMm }} mm</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mx-auto mt-8 w-full max-w-4xl">
        <h3 class="match-guide__table-title">Visual reference chart</h3>
        <GuideImage
          class="mt-4 overflow-hidden rounded-xl shadow-md"
          src="/public/tiresizecharts/match/exact-circumference-of-tire.webp"
          alt="Schwalbe approximate bicycle tire wheel circumference chart by inch and ETRTO size"
          caption="Reference chart for approximate tire circumferences"
          :zoomOnClick="true"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
import {
  tireCircumferenceRows,
  tireFrameClearanceTables,
} from '~/data/tireguides/match'
</script>

<style scoped>
.match-guide__measurement-reference {
  margin: 0 auto;
  max-width: 72rem;
  border: 1px solid rgba(5, 150, 105, 0.24);
  border-radius: 1rem;
  background: var(--tz-surface-subtle);
  padding: 1rem;
  text-align: left;
}

.match-guide__measurement-title {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.4;
  text-align: center;
  text-transform: uppercase;
}

.match-guide__measurement-grid {
  display: grid;
  gap: 1rem;
  margin-top: 1.25rem;
}

.match-guide__measurement-item {
  display: grid;
  grid-template-columns: 2rem minmax(0, 1fr);
  gap: 0.7rem;
  border-top: 1px solid rgba(5, 150, 105, 0.2);
  padding-top: 0.85rem;
}

.match-guide__measurement-badge {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-radius: 0.5rem;
  background: rgba(5, 150, 105, 0.12);
  color: var(--tz-text-accent);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1;
}

.match-guide__measurement-item h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1.35;
}

.match-guide__measurement-item p {
  margin: 0.35rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.55;
}

.match-guide__table-list {
  display: grid;
  gap: 1.5rem;
  text-align: left;
}

.match-guide__table-section {
  min-width: 0;
}

.match-guide__table-title {
  margin: 0 0 0.65rem;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.35;
  text-transform: uppercase;
}

.match-guide__table-scroll {
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.match-guide__table {
  min-width: 60rem;
  width: 100%;
  border-collapse: collapse;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.4;
  table-layout: fixed;
}

.match-guide__table--circumference {
  min-width: 38rem;
}

.match-guide__table caption {
  padding: 0.75rem 0.85rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.45;
  text-align: left;
}

.match-guide__table th,
.match-guide__table td {
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  padding: 0.55rem 0.65rem;
  text-align: left;
  vertical-align: top;
  overflow-wrap: anywhere;
}

.match-guide__table thead th {
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1.35;
}

.match-guide__table tbody tr:hover {
  background: var(--tz-surface-subtle);
}

.match-guide__table tbody th {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.match-guide__dimension-cell {
  color: var(--tz-text-secondary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.match-guide__table th:nth-child(1),
.match-guide__table td:nth-child(1) {
  width: 12%;
}

.match-guide__table th:nth-child(2),
.match-guide__table td:nth-child(2) {
  width: 14%;
}

.match-guide__table th:nth-child(3),
.match-guide__table td:nth-child(3) {
  width: 22%;
}

.match-guide__table th:nth-child(4),
.match-guide__table td:nth-child(4),
.match-guide__table th:nth-child(5),
.match-guide__table td:nth-child(5) {
  width: 15%;
}

.match-guide__table th:nth-child(6),
.match-guide__table td:nth-child(6) {
  width: 22%;
}

.match-guide__table--circumference th:nth-child(1),
.match-guide__table--circumference td:nth-child(1) {
  width: 25%;
}

.match-guide__table--circumference th:nth-child(2),
.match-guide__table--circumference td:nth-child(2) {
  width: 30%;
}

.match-guide__table--circumference th:nth-child(3),
.match-guide__table--circumference td:nth-child(3) {
  width: 45%;
}

@media (min-width: 768px) {
  .match-guide__measurement-reference {
    padding: 1.25rem;
  }

  .match-guide__measurement-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 1.25rem;
  }

  .match-guide__measurement-item {
    display: block;
    border-top-width: 2px;
    padding: 0.85rem 0.25rem 0;
  }

  .match-guide__measurement-item h4 {
    margin-top: 0.7rem;
  }

  .match-guide__table {
    min-width: 0;
  }
}
</style>
