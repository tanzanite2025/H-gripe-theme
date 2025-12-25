<template>
  <div>
    <p class="sizecharts-section__intro">
      Basic spoke concepts: dimensions, materials and typical use cases for different riding styles.
    </p>

    <ul class="sizecharts-section__list">
      <li class="list-none pl-0">
        Common spoke diameters and shapes (straight / J-bend, round / bladed).
        <div class="mt-2 grid grid-cols-1 gap-3 md:grid-cols-2">
          <GuideImage
            src="/public/technical/j-bend-spokeandstraightpull-spoke.webp"
            alt="Comparison of J-bend and straight-pull bicycle spokes"
            :zoomOnClick="true"
            caption="This illustration compares J-bend and straight-pull spoke heads and how they seat in the hub flange."
          />
          <GuideImage
            src="/public/technical/Sapim-black-stainless-steel-straight-gauge-2.0mm-spoke-double-butted-2.0-1.8-2.0mm-spoke-aero-flat-bladed-spoke.webp"
            alt="Examples of Sapim stainless steel spokes in different shapes and profiles"
            :zoomOnClick="true"
            caption="Examples of stainless steel spokes in straight gauge, double-butted and aero bladed profiles used on performance wheelsets."
          />
        </div>
      </li>
    </ul>

    <div class="mt-4">
      <p class="sizecharts-section__intro">
        The following is a detailed introduction to different spoke materials:
      </p>
      <ul class="sizecharts-section__list mt-2">
        <li class="rounded-xl border border-slate-500/40 bg-slate-900/80 p-3 md:p-4">
          <strong class="block mb-1 text-amber-300">1. Stainless Steel</strong>
          <p>
            Stainless steel is currently the most mainstream and widely used spoke
            material on the market.
          </p>
          <p>The main brands currently on the market include:</p>
          <ul>
            <li>
              <strong>Sapim</strong> &mdash; regarded as No. 1 in spokes and focused solely on
              spokes. Several of its models are high quality, and CX-RAY is used by many
              wheel brands.
              <button type="button" class="technical-brand-button">
                Sapim
              </button>
            </li>
            <li>
              <strong>DT Swiss</strong> &mdash; globally recognised quality, producing spokes,
              wheels and hubs.
              <button type="button" class="technical-brand-button">
                DT Swiss
              </button>
            </li>
            <li>
              <strong>Pillar</strong> &mdash; currently the third-ranked brand and very
              cost-effective compared with the first two.
              <button type="button" class="technical-brand-button">
                Pillar
              </button>
            </li>
          </ul>
          <p>We also offer another cost-effective brand:</p>
          <ul>
            <li>
              <strong>YT (Yongteng)</strong> &mdash; an emerging Chinese brand producing both
              bicycle and motorcycle spokes.
              <button type="button" class="technical-brand-button">
                YT
              </button>
            </li>
          </ul>
          <p><strong>Advantages:</strong></p>
          <ul>
            <li><strong>High strength and durability</strong> &mdash; able to withstand high
            tension and long-term use.</li>
            <li><strong>Strong corrosion resistance</strong> &mdash; not easy to rust and
            suitable for various weather conditions.</li>
            <li><strong>Cost-effective</strong> &mdash; relatively low cost while balancing
            weight and performance.</li>
            <li><strong>Good toughness</strong> &mdash; steel spokes have ductility and
            elasticity, more likely to deform gradually than suddenly break under impact.</li>
          </ul>
          <p><strong>Disadvantages:</strong></p>
          <ul>
            <li><strong>Relatively heavy weight</strong> compared with carbon fiber or titanium
            alloys.</li>
          </ul>
          <p>
            <strong>Applicable scenarios:</strong> suitable for almost all types of bicycles,
            from daily commuter bikes to professional mountain and road bikes.
          </p>
        </li>
        <li class="mt-3 rounded-xl border border-slate-500/40 bg-slate-900/80 p-3 md:p-4">
          <strong class="block mb-1 text-amber-300">2. Carbon Fiber</strong>
          <p>
            Carbon fiber spokes are typically found on the most high-end,
            performance-focused wheelsets. We also offer factory-developed carbon fiber
            spokes.
          </p>
          <p><strong>Advantages:</strong></p>
          <ul>
            <li><strong>Ultra-lightweight</strong> &mdash; can significantly reduce wheelset
            weight compared with steel spokes (often saving more than 100 grams) and improve
            acceleration.</li>
            <li><strong>Excellent rigidity</strong> &mdash; provides very good lateral rigidity
            and tensile strength, giving more direct power transfer and sharper handling.</li>
            <li><strong>Good shock absorption</strong> &mdash; can absorb part of riding
            vibration and improve comfort.</li>
          </ul>
          <p><strong>Disadvantages:</strong></p>
          <ul>
            <li><strong>Expensive</strong> &mdash; much higher price than stainless steel
            spokes.</li>
            <li><strong>Poor side impact resistance</strong> &mdash; very strong in tension but
            vulnerable to side impacts or punctures; once damaged it usually cannot be
            repaired and must be replaced.</li>
            <li><strong>Higher maintenance requirements</strong> &mdash; installation and
            service require professional skills and specific tools.</li>
          </ul>
          <p>
            <strong>Applicable scenarios:</strong> high-end road racing, time trials,
            triathlons and other situations that demand extreme light weight and stiffness.
          </p>
        </li>
        <li class="mt-3 rounded-xl border border-slate-500/40 bg-slate-900/80 p-3 md:p-4">
          <strong class="block mb-1 text-amber-300">3. Titanium Alloy</strong>
          <p>
            Titanium spokes sit between steel and carbon fiber in character, but are
            relatively rare on the market. We plan to launch titanium options later.
          </p>
          <p><strong>Advantages:</strong></p>
          <ul>
            <li><strong>Lightweight and high strength</strong> &mdash; very high
            strength-to-weight ratio, lighter than stainless steel while remaining strong.</li>
            <li><strong>Excellent corrosion resistance</strong> &mdash; almost no rust.</li>
            <li><strong>Smoother ride feel</strong> &mdash; good vibration absorption for a
            smoother riding experience.</li>
          </ul>
          <p><strong>Disadvantages:</strong></p>
          <ul>
            <li><strong>Very expensive</strong> compared with stainless steel.</li>
            <li><strong>Limited availability</strong> &mdash; relatively few options on the
            market.</li>
          </ul>
          <p>
            <strong>Applicable scenarios:</strong> high-end custom wheelsets and riders who
            pursue light weight, comfort and uniqueness.
          </p>
        </li>
        <li class="mt-3 rounded-xl border border-slate-500/40 bg-slate-900/80 p-3 md:p-4">
          <strong class="block mb-1 text-amber-300">4. Aluminum Alloy</strong>
          <p>
            Aluminum spokes are relatively rare and mainly used by a few manufacturers (such
            as Campagnolo) in certain high-end wheelsets.
          </p>
          <p><strong>Advantages:</strong></p>
          <ul>
            <li><strong>Very light</strong> &mdash; one of the lightest metal spoke
            materials.</li>
          </ul>
          <p><strong>Disadvantages:</strong></p>
          <ul>
            <li><strong>Lower strength and durability</strong> than stainless steel or titanium
            alloy, more prone to fatigue fracture.</li>
            <li><strong>More prone to corrosion</strong> and requires more careful
            maintenance.</li>
          </ul>
          <p>
            <strong>Applicable scenarios:</strong> specially designed ultra-lightweight racing
            wheelsets.
          </p>
        </li>
      </ul>
      <p class="sizecharts-section__intro mt-3">
        Which material you choose ultimately depends on your budget, riding style and
        specific performance needs such as weight, stiffness and comfort.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
</script>
