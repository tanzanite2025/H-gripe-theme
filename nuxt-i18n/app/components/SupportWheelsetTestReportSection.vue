<template>
  <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-3 md:p-6">
    <h3 class="support-section__title text-center">Wheelset Test Report</h3>

    <p class="support-section__body mt-4 text-center">
      This section summarises how TANZANITE wheelsets are validated in our in-house lab,
      covering lateral load, torsional stiffness, environmental durability, dynamic balance,
      fatigue, and braking performance tests.
    </p>

    <div class="mt-3 flex justify-center">
      <div
        class="inline-flex items-center rounded-full border border-sky-500/70 bg-sky-500/5 px-3 py-1 text-xs sm:text-sm text-sky-300"
      >
        <span>To learn more about wheelset shipping and assembly tests,</span>
        <button
          type="button"
          class="ml-1 underline decoration-sky-400/80 underline-offset-2 hover:text-sky-200 hover:decoration-sky-200 transition-colors duration-150"
          @click="goToWheelsetAssembly"
        >
          see Wheelset Assembly.
        </button>
      </div>
    </div>

    <div class="rim-test-card mt-6">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Lateral Load Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Tests wheelset stability and stiffness under lateral forces to ensure safe cornering and resistance to side
        impacts.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Secure the complete wheelset in the testing fixture.</li>
        <li>
          Apply a lateral static load at the rim or tire contact point, typically 600-800 N (approximately 60-80 kgf),
          and maintain for a specified duration (e.g., 1 minute).
        </li>
        <li>
          Conduct dynamic cyclic lateral loading, for example applying alternating forces of 300-500 N over
          50,000-100,000 cycles to simulate cornering stresses, crosswinds, or side impacts during riding.
        </li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>The wheelset must not exhibit cracks, fractures, or permanent structural failure.</li>
        <li>Lateral deformation must remain within specified limits (e.g., lateral deviation ≤ 1.5 mm).</li>
        <li>
          After testing, the wheelset must retain full functionality, ensuring stability and safety in real-world use.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Torsional Stiffness Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Evaluates wheelset stiffness and response when transmitting torque, ensuring efficient power transfer and
        reliability during acceleration, braking, and climbing.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Secure the complete wheelset in the testing fixture, simulating hub drive conditions.</li>
        <li>
          Apply a torque load at the hub drive side, typically 40-80 Nm, and maintain for a specified duration (e.g.,
          1 minute).
        </li>
        <li>
          Conduct dynamic cyclic torque testing, for example applying alternating torques of 20-40 Nm over
          50,000-100,000 cycles to simulate repeated acceleration and braking forces.
        </li>
        <li>Measure angular deformation of the rim and spokes to evaluate stiffness and responsiveness.</li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>The wheelset must not exhibit cracks, fractures, or permanent structural failure.</li>
        <li>Angular torsional deformation must remain within specified limits (e.g., ≤ 2.0°).</li>
        <li>
          After testing, the wheelset must retain full functionality, ensuring efficient power transfer and safe riding
          performance.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Environmental Durability Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Simulates humidity, temperature, salt spray, and other environmental conditions to assess rim and wheelset
        durability, ensuring long-term reliability in harsh environments.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 9227 & ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Place the rim or complete wheelset in an environmental chamber.</li>
        <li>
          Humidity Test: Maintain relative humidity at 95% for 48 hours to observe moisture absorption or
          corrosion.
        </li>
        <li>
          Temperature Cycling Test: Cycle between -20°C and +60°C for 10-20 cycles, including heating, holding, and
          cooling phases.
        </li>
        <li>
          Salt Spray Test: Expose to a 5% NaCl solution salt spray for 96 hours to simulate coastal or rainy
          conditions.
        </li>
        <li>Additional conditions may include UV exposure, mud spray, or chemical agent exposure.</li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>No cracks, corrosion perforation, or structural failure should occur.</li>
        <li>Surface coating or anodized layer must remain intact without significant peeling.</li>
        <li>Dimensional changes must remain within specified tolerances (e.g., diameter variation ≤ 0.5 mm).</li>
        <li>
          After testing, the rim and wheelset must retain full functionality, ensuring safe long-term use.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Dynamic Balance Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Tests wheelset balance and stability during high-speed rotation to ensure smooth performance and rider
        safety.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Mount the complete wheelset on a high-speed rotation testing rig.</li>
        <li>Gradually increase rotational speed to an equivalent of 60-80 km/h, maintaining stable rotation.</li>
        <li>Use sensors to monitor radial and lateral vibrations, as well as imbalance-induced wobble.</li>
        <li>Simulate different operating conditions such as tire installation, applied loads, or braking forces.</li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>No excessive vibration or wobble that could compromise riding stability.</li>
        <li>Dynamic imbalance must remain within specified limits (e.g., ≤ 10 g·mm).</li>
        <li>The wheelset must retain structural integrity under high-speed rotation, with no cracks or failures.</li>
        <li>
          After testing, the wheelset must remain fully functional, ensuring safe and comfortable high-speed riding.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Fatigue Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        Simulates repeated loads over long-term riding to verify rim durability and structural reliability.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Mount the complete wheelset on a fatigue testing machine.</li>
        <li>
          Apply alternating cyclic loads at the rim or tire contact point, typically 300-600 N, to replicate continuous
          stresses during riding.
        </li>
        <li>
          Conduct high-cycle testing, for example 100,000-200,000 cycles, to evaluate long-term durability.
        </li>
        <li>Monitor radial and lateral deformation of the rim as well as spoke tension variations during the test.</li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>The rim must not exhibit cracks, fractures, or permanent structural failure.</li>
        <li>Deformation must remain within specified limits (e.g., radial deviation ≤ 1.0 mm, lateral deviation ≤ 1.5 mm).</li>
        <li>
          After testing, the rim must retain full functionality, ensuring safe and reliable long-term riding
          performance.
        </li>
      </ul>
    </div>

    <div class="rim-test-card mt-4">
      <h4 class="sizecharts-section__subheading text-sky-300 font-semibold">
        Braking Performance Test
      </h4>
      <p class="support-section__body">
        <strong>Purpose:</strong>
        For V-brake and disc brake wheelsets, tests stability and heat resistance during braking to ensure safe and
        reliable performance.
      </p>
      <p class="support-section__body mt-2">
        <strong>Test Method:</strong>
        <span class="inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md bg-slate-800/80 border border-slate-700/50 text-xs font-medium text-slate-300 align-middle">
          <svg class="w-3.5 h-3.5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          ISO 4210-7:2023
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>Mount the complete wheelset on a braking performance testing rig.</li>
        <li>
          For V-brake wheelsets: Apply braking force at the rim surface, typically 300-500 N, and maintain for a
          specified duration.
        </li>
        <li>
          For disc brake wheelsets: Apply torque load at the rotor, typically 40-60 Nm, and conduct sustained braking
          cycles.
        </li>
        <li>Gradually increase rotational speed to an equivalent of 30-60 km/h, performing multiple braking cycles.</li>
        <li>Use thermocouples or infrared sensors to monitor temperature rise at the braking surface or rotor.</li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>Evaluation Criteria:</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li>No cracks, excessive wear, or structural failure should occur.</li>
        <li>The wheelset must remain stable during braking, with no significant wobble or vibration.</li>
        <li>Temperature rise must remain within specified limits (e.g., rim ≤ 200°C, rotor ≤ 250°C).</li>
        <li>
          After testing, the wheelset must retain full functionality, ensuring safe long-term braking performance.
        </li>
      </ul>
    </div>

    <div class="mt-4 flex justify-center">
      <div
        class="support-video-thumbnail"
        @click="openWheelsetVideo"
      >
        <img
          class="support-video-thumbnail__image"
          src="/testreport/wheelsettestreport/tanzanite-wheelssettestroport-video-firstpicture.webp"
          alt="Play wheelset test report video for Tanzanite wheelsets"
          loading="lazy"
        />
        <div class="support-video-thumbnail__overlay">
          <span class="support-video-thumbnail__icon">▶</span>
          <span class="support-video-thumbnail__label">Watch wheelset test report video</span>
        </div>
      </div>
    </div>

    <div
      class="mt-6 rounded-lg border border-amber-300/80 bg-slate-900/80 px-4 py-3 text-sm leading-relaxed text-amber-100"
    >
      <h4 class="mb-1 font-semibold text-amber-300">
        Disclaimer
      </h4>
      <p class="mb-2">
        We pick only one sample for every test report and the results will likely vary as the rim diameters change.
        Please note that the differences between models in test results are specially designed by our engineers for the
        intended uses.
      </p>
      <p>
        All the test results of this section are based on the lab criteria of TANZANITE and are implemented at our
        well-established testing facilities. TANZANITE is only responsible for the test results themselves which are
        not set for any comparison to other brands or such regards.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const { openWheelsetVideo, goToWheelsetAssembly } = defineProps<{
  openWheelsetVideo: () => void | Promise<void>
  goToWheelsetAssembly: () => void | Promise<void>
}>()
</script>

<style scoped>
.support-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: #e5e7eb;
}

.support-section__body {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: rgba(148, 163, 184, 0.9);
}

.rim-test-card {
  padding: 1.25rem;
  border-radius: 0.75rem;
  background: rgba(2, 6, 23, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.03);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

.support-video-thumbnail {
  margin-top: 0.75rem;
  position: relative;
  border-radius: 0.75rem;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.85);
}

.support-video-thumbnail__image {
  display: block;
  width: 100%;
  height: auto;
}

.support-video-thumbnail__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  background: linear-gradient(
    to top,
    rgba(15, 23, 42, 0.85),
    rgba(15, 23, 42, 0.45)
  );
  color: #e5e7eb;
}

.support-video-thumbnail__icon {
  font-size: 1.8rem;
}

.support-video-thumbnail__label {
  font-size: 0.9rem;
  font-weight: 500;
}
</style>
