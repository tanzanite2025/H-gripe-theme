<template>
  <div>
    <p class="sizecharts-section__intro">
      We are preparing a step-by-step guide to help you choose the right freehub body for your drivetrain.
      This section will explain compatibility between HG, MS, XD, XDR and different cassette setups.
    </p>

    <FreehubGroupsetHelper />

    <div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <GuideImage
        src="/public/wheelsetbuyersguide/choose%20freehub/shimano-micro-spline-11-12-speed-mountain-freehub.webp"
        alt="Shimano Micro Spline 11/12-speed mountain freehub body"
        :zoomOnClick="true"
        caption="Shimano Micro Spline 11/12-speed MTB freehub"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/choose%20freehub/shimano-8-9-10-11-speed-road-hyper-freehub.webp"
        alt="Shimano 8/9/10/11-speed road HG freehub body"
        :zoomOnClick="true"
        caption="Shimano 8/9/10/11-speed road HG freehub"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/choose%20freehub/shimano-8-9-10-11-speed-mountain-hyper-freehub.webp"
        alt="Shimano 8/9/10/11-speed mountain HG freehub body"
        :zoomOnClick="true"
        caption="Shimano 8/9/10/11-speed MTB HG freehub"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/choose%20freehub/sram-xd-11-12-speed-mountain-freehub.webp"
        alt="SRAM XD 11/12-speed mountain freehub body"
        :zoomOnClick="true"
        caption="SRAM XD 11/12-speed MTB freehub"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/choose%20freehub/sram-xdr-road-11-12-speed-freehub.webp"
        alt="SRAM XDR 11/12-speed road freehub body"
        :zoomOnClick="true"
        caption="SRAM XDR 11/12-speed road freehub"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/choose%20freehub/Campagnolo-8-9-10-11-N3W-freehub.webp"
        alt="Campagnolo N3W freehub body for 13-speed and 9–12-speed cassettes"
        :zoomOnClick="true"
        caption="Campagnolo N3W 13-speed and 9–12-speed freehub"
      />
      <GuideImage
        src="/public/wheelsetbuyersguide/choose%20freehub/Campagnolo-8-9-10-11-spd-freehub.webp"
        alt="Campagnolo 8/9/10/11-speed classic freehub body"
        :zoomOnClick="true"
        caption="Campagnolo classic 8/9/10/11-speed freehub"
      />
    </div>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-3"
    >
      Freehub types and matching groupsets
    </h3>
    <p class="mt-2 text-center text-slate-100 text-sm">
      Different freehub bodies correspond to specific groupsets with clear compatibility ranges. For example:
    </p>
    <ul class="mt-2 space-y-1 text-center text-slate-100 text-xs sm:text-sm">
      <li>
        MS (Micro Spline 12-speed) &rarr; Shimano MTB 12-speed groupsets (such as Deore M6100, SLX M7100, XT M8100,
        XTR M9100).
      </li>
      <li>
        HG 8-speed &rarr; Shimano Claris, Altus, Acera and other entry-level road / MTB groupsets.
      </li>
      <li>
        HG 9&ndash;11 speed &rarr; Shimano Sora, Tiagra, 105, Ultegra, Dura-Ace, as well as MTB groupsets like Deore,
        SLX, XT.
      </li>
      <li>
        XD &rarr; SRAM MTB 11/12-speed groupsets (such as GX Eagle, X01 Eagle, XX1 Eagle).
      </li>
      <li>
        XDR &rarr; SRAM road 12-speed groupsets (such as Force eTap AXS, Red eTap AXS).
      </li>
      <li>
        N3W &rarr; Campagnolo Ekar 13-speed; with the appropriate adapter it can also run Record, Super Record and
        Chorus 10&ndash;12-speed cassettes (new standard with backward compatibility).
      </li>
      <li>
        CP (Campagnolo classic freehub) &rarr; supports only Campagnolo 11-speed road groupsets (such as Super Record 11,
        Record 11, Chorus 11) and does not support 12/13-speed.
      </li>
    </ul>

    <h3
      class="sizecharts-section__subheading text-sky-300 font-semibold text-center mt-4"
    >
      Notes
    </h3>
    <ul class="mt-2 space-y-1 text-center text-slate-100 text-xs sm:text-sm">
      <li>
        MS freehub is Shimano&rsquo;s newer MTB spline standard designed around 12-speed cassettes and is not
        backward-compatible with HG freehubs.
      </li>
      <li>
        HG freehub is the most widely used standard across 8&ndash;11 speed road and MTB (some combinations may require
        spacers).
      </li>
      <li>
        XD / XDR freehubs are mainly used for SRAM; XD is MTB-oriented, while XDR is road-oriented.
      </li>
      <li>
        N3W is Campagnolo&rsquo;s latest standard, offering strong compatibility and good backward support.
      </li>
      <li>
        CP freehub is Campagnolo&rsquo;s traditional design, supporting only 11-speed road.
      </li>
    </ul>

    <!-- Freehub type vs groupset overview (translated from Chinese chart) -->
    <div class="mt-4 overflow-x-auto">
      <table
        class="min-w-full border-separate border-spacing-0 rounded-xl overflow-hidden bg-slate-900/70 text-slate-100 text-xs sm:text-sm"
      >
        <thead>
          <tr class="bg-slate-800/90 font-semibold text-slate-50">
            <th class="px-3 py-2 text-left border-b border-slate-700">Freehub type</th>
            <th class="px-3 py-2 text-left border-b border-slate-700">Example compatible groupsets</th>
            <th class="px-3 py-2 text-left border-b border-slate-700">Key features</th>
          </tr>
        </thead>
        <tbody>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">MS (Micro Spline 12-speed)</td>
            <td class="px-3 py-2 align-top">
              Shimano Deore M6100, SLX M7100, XT M8100, XTR M9100
            </td>
            <td class="px-3 py-2 align-top">
              Designed specifically for MTB 12-speed; supports 10T smallest cog.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">HG 8-speed</td>
            <td class="px-3 py-2 align-top">
              Shimano Claris R2000, Altus M310, Acera M360
            </td>
            <td class="px-3 py-2 align-top">
              Entry-level road / MTB setups, typically with 11&ndash;32T cassettes.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">HG 9&ndash;11 speed</td>
            <td class="px-3 py-2 align-top">
              Shimano Sora R3000, Tiagra 4700, 105 R7000, Ultegra R8000, Dura-Ace R9100;
              MTB Deore M6000, SLX M7000, XT M8000
            </td>
            <td class="px-3 py-2 align-top">
              Most widely used standard with strong cross-compatibility.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">XD</td>
            <td class="px-3 py-2 align-top">
              SRAM GX Eagle, X01 Eagle, XX1 Eagle (MTB 12-speed)
            </td>
            <td class="px-3 py-2 align-top">
              Supports 10T smallest cog; focused on lightweight MTB performance.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">XDR</td>
            <td class="px-3 py-2 align-top">
              SRAM Force eTap AXS, Red eTap AXS (road 12-speed)
            </td>
            <td class="px-3 py-2 align-top">
              Road-specific version of XD, 1.85 mm longer than XD.
            </td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">N3W</td>
            <td class="px-3 py-2 align-top">
              Campagnolo Ekar 13-speed; with adapter also compatible with Record, Super Record and Chorus
              10&ndash;12-speed cassettes
            </td>
            <td class="px-3 py-2 align-top">
              New 13-speed freehub platform designed for Ekar with backward compatibility to selected 10&ndash;12-speed
              Campagnolo road cassettes via adapter.
            </td>
          </tr>
          <tr>
            <td class="px-3 py-2 align-top">CP (Campagnolo classic)</td>
            <td class="px-3 py-2 align-top">
              Campagnolo Super Record 11, Record 11, Chorus 11
            </td>
            <td class="px-3 py-2 align-top">
              Traditional Campagnolo 11-speed-only road freehub standard (not compatible with 12/13-speed cassettes).
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="mt-4 overflow-x-auto">
      <table
        class="min-w-full border-separate border-spacing-0 rounded-xl overflow-hidden bg-slate-900/70 text-slate-100 text-xs sm:text-sm"
      >
        <thead>
          <tr class="bg-slate-800/90 font-semibold text-slate-50">
            <th class="px-3 py-2 text-left border-b border-slate-700">Brand</th>
            <th class="px-3 py-2 text-left border-b border-slate-700">Freehub</th>
            <th class="px-3 py-2 text-left border-b border-slate-700">Cassette</th>
            <th class="px-3 py-2 text-left border-b border-slate-700">Spacer(s) needed</th>
          </tr>
        </thead>
        <tbody>
          <!-- Shimano -->
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">8/9/11 speed mountain cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">10 speed cassette</td>
            <td class="px-3 py-2">1 mm spacer</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">11-speed road cassette</td>
            <td class="px-3 py-2">not compatible</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 10-speed ONLY</td>
            <td class="px-3 py-2">10 speed cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 10-speed ONLY</td>
            <td class="px-3 py-2">other-speed cassettes</td>
            <td class="px-3 py-2">not compatible</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 11/12 speed road freehub</td>
            <td class="px-3 py-2">8/9/11 speed mountain cassette</td>
            <td class="px-3 py-2">1.85 mm spacer</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 11/12 speed road freehub</td>
            <td class="px-3 py-2">10 speed cassette</td>
            <td class="px-3 py-2">1.85 mm + 1 mm spacers</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 11/12 speed road freehub</td>
            <td class="px-3 py-2">11-speed road cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG 11/12 speed road freehub</td>
            <td class="px-3 py-2">12-speed road cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">MS 11/12 speed mountain freehub</td>
            <td class="px-3 py-2">11/12 speed cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG L2 12-speed road only freehub</td>
            <td class="px-3 py-2">11-speed road cassette</td>
            <td class="px-3 py-2">not compatible</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Shimano</td>
            <td class="px-3 py-2 align-top">HG L2 12-speed road only freehub</td>
            <td class="px-3 py-2">12-speed road cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>

          <!-- SRAM -->
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">HG-style 9/10 speed freehub</td>
            <td class="px-3 py-2">SRAM 9 speed</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">HG-style 9/10 speed freehub</td>
            <td class="px-3 py-2">SRAM 10 speed</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">HG-style 9/10 speed freehub</td>
            <td class="px-3 py-2">NX Eagle 12 speed cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">HG-style 11-speed road freehub</td>
            <td class="px-3 py-2">SRAM 9 speed</td>
            <td class="px-3 py-2">1.85 mm spacer</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">HG-style 11-speed road freehub</td>
            <td class="px-3 py-2">SRAM 10 speed</td>
            <td class="px-3 py-2">1.85 mm spacer</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">HG-style 11-speed road freehub</td>
            <td class="px-3 py-2">SRAM 11 speed</td>
            <td class="px-3 py-2">no</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">XD 11 &amp; 12 speed freehub</td>
            <td class="px-3 py-2">
              SRAM 11 &amp; 12 speed mountain XD cassettes
              <span class="block text-[0.7rem] text-slate-300">
                (SRAM XX1, XX1 Eagle, X01, X01 Eagle, X1, GX, GX Eagle)
              </span>
            </td>
            <td class="px-3 py-2">no</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">XDR 11 &amp; 12 speed freehub</td>
            <td class="px-3 py-2">SRAM 11/12 speed road XDR cassette</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">SRAM</td>
            <td class="px-3 py-2 align-top">XDR 11 &amp; 12 speed freehub</td>
            <td class="px-3 py-2">SRAM XD cassette</td>
            <td class="px-3 py-2">1.85 mm spacer</td>
          </tr>

          <!-- Campagnolo -->
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Campagnolo</td>
            <td class="px-3 py-2 align-top">9/10/11/12 speed freehub</td>
            <td class="px-3 py-2">9/10/11/12 speed Ultra-Drive cassettes</td>
            <td class="px-3 py-2">no</td>
          </tr>

          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Campagnolo</td>
            <td class="px-3 py-2 align-top">N3W 13 speed freehub</td>
            <td class="px-3 py-2">12/13 speed cassettes (9t/10t cog)</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Campagnolo</td>
            <td class="px-3 py-2 align-top">N3W 13 speed freehub</td>
            <td class="px-3 py-2">9/10/11/12 speed cassettes (11t starting cog)</td>
            <td class="px-3 py-2">AC21-N3W adapter</td>
          </tr>

          <!-- Mavic -->
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Mavic</td>
            <td class="px-3 py-2 align-top">Shimano HG-style 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">Shimano/SRAM 9-speed cassette</td>
            <td class="px-3 py-2">1.75 mm Mavic adapter</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Mavic</td>
            <td class="px-3 py-2 align-top">Shimano HG-style 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">Shimano/SRAM 10-speed cassette</td>
            <td class="px-3 py-2">1.75 mm Mavic spacer</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Mavic</td>
            <td class="px-3 py-2 align-top">Shimano HG-style 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">SRAM 9 speed</td>
            <td class="px-3 py-2">1.75 mm Mavic spacer</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Mavic</td>
            <td class="px-3 py-2 align-top">Shimano HG-style 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">Shimano 10 speed</td>
            <td class="px-3 py-2">1 mm + 1.75 mm Mavic spacer</td>
          </tr>
          <tr class="border-b border-slate-800/70">
            <td class="px-3 py-2 align-top">Mavic</td>
            <td class="px-3 py-2 align-top">Shimano HG-style 8&ndash;11 speed mountain freehub</td>
            <td class="px-3 py-2">Shimano/SRAM 11 speed</td>
            <td class="px-3 py-2">no</td>
          </tr>
          <tr>
            <td class="px-3 py-2 align-top">Mavic</td>
            <td class="px-3 py-2 align-top">Campy-style freehub</td>
            <td class="px-3 py-2">Campy 9/10/11/12 speed</td>
            <td class="px-3 py-2">no</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
import FreehubGroupsetHelper from '~/components/FreehubGroupsetHelper.vue'
</script>
