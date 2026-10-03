<template>
  <article class="spoke-dislocation-page">
    <header class="spoke-dislocation-page__hero">
      <div>
        <p class="spoke-dislocation-page__eyebrow">{{ t('guidesSpokeDislocationMechanics.version') }}</p>
        <h1>{{ t('guidesSpokeDislocationMechanics.title') }}</h1>
        <p class="spoke-dislocation-page__subtitle">{{ t('guidesSpokeDislocationMechanics.subtitle') }}</p>
      </div>
      <div class="spoke-dislocation-page__tags" :aria-label="t('guidesSpokeDislocationMechanics.tags.ariaLabel')">
        <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--dark">{{ t('guidesSpokeDislocationMechanics.tags.material') }}</span>
        <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--emerald">{{ t('guidesSpokeDislocationMechanics.tags.backend') }}</span>
        <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--blue">{{ t('guidesSpokeDislocationMechanics.tags.elongation') }}</span>
        <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--amber">{{ t('guidesSpokeDislocationMechanics.tags.scope') }}</span>
      </div>
    </header>

    <div class="spoke-dislocation-page__two-column">
      <section class="spoke-dislocation-page__card">
        <div class="spoke-dislocation-page__card-heading">
          <h2>{{ t('guidesSpokeDislocationMechanics.materialLadder.title') }}</h2>
          <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--emerald">{{ t('guidesSpokeDislocationMechanics.materialLadder.tag') }}</span>
        </div>
        <p class="spoke-dislocation-page__body-copy">{{ t('guidesSpokeDislocationMechanics.materialLadder.description') }}</p>
        <div class="spoke-dislocation-page__load-ladder">
          <div class="spoke-dislocation-page__load-step spoke-dislocation-page__load-step--blue">
            <strong>{{ t('guidesSpokeDislocationMechanics.materialLadder.workingValueLabel') }}</strong>
            <div>
              <b>{{ t('guidesSpokeDislocationMechanics.materialLadder.workingTitle') }}</b>
              <span>{{ t('guidesSpokeDislocationMechanics.materialLadder.workingDescription') }}</span>
            </div>
          </div>
          <div class="spoke-dislocation-page__load-step spoke-dislocation-page__load-step--emerald">
            <strong>{{ t('guidesSpokeDislocationMechanics.materialLadder.overloadValueLabel') }}</strong>
            <div>
              <b>{{ t('guidesSpokeDislocationMechanics.materialLadder.overloadTitle') }}</b>
              <span>{{ t('guidesSpokeDislocationMechanics.materialLadder.overloadDescription') }}</span>
            </div>
          </div>
          <div class="spoke-dislocation-page__load-step spoke-dislocation-page__load-step--slate">
            <strong>{{ backendReferenceTranslationValues.rimBaseline }} N</strong>
            <div>
              <b>{{ t('guidesSpokeDislocationMechanics.materialLadder.rimTitle') }}</b>
              <span>{{ t('guidesSpokeDislocationMechanics.materialLadder.rimDescription') }}</span>
            </div>
          </div>
          <div class="spoke-dislocation-page__load-step spoke-dislocation-page__load-step--amber">
            <strong>{{ backendReferenceTranslationValues.yieldReference }} MPa</strong>
            <div>
              <b>{{ t('guidesSpokeDislocationMechanics.materialLadder.yieldTitle') }}</b>
              <span>{{ t('guidesSpokeDislocationMechanics.materialLadder.yieldDescription') }}</span>
            </div>
          </div>
        </div>
        <div class="spoke-dislocation-page__callout spoke-dislocation-page__callout--emerald">
          <span aria-hidden="true">💡</span>
          <p>{{ t('guidesSpokeDislocationMechanics.materialLadder.note') }}</p>
        </div>
      </section>

      <section class="spoke-dislocation-page__card">
        <div class="spoke-dislocation-page__card-heading">
          <h2>{{ t('guidesSpokeDislocationMechanics.microstructure.title') }}</h2>
          <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--emerald">{{ t('guidesSpokeDislocationMechanics.microstructure.tag') }}</span>
        </div>
        <p class="spoke-dislocation-page__body-copy">{{ t('guidesSpokeDislocationMechanics.microstructure.description') }}</p>
        <svg
          class="spoke-dislocation-page__microstructure-diagram"
          viewBox="0 0 600 240"
          role="img"
          :aria-label="t('guidesSpokeDislocationMechanics.microstructure.ariaLabel')"
        >
          <defs>
            <marker id="spoke-dislocation-state-arrow" markerWidth="7" markerHeight="7" refX="6" refY="3.5" orient="auto">
              <path d="M0,0 L7,3.5 L0,7 Z" fill="#e11d48" />
            </marker>
          </defs>
          <rect x="20" y="20" width="260" height="200" rx="10" fill="currentColor" opacity="0.04" stroke="#cbd5e1" stroke-dasharray="4" />
          <text x="35" y="42" font-size="11" font-weight="800" fill="currentColor">{{ t('guidesSpokeDislocationMechanics.microstructure.stateA') }}</text>
          <g fill="#e11d48" font-size="16" font-weight="900" font-family="var(--tz-font-ui)">
            <text x="60" y="80">丄</text>
            <text x="140" y="110">丄</text>
            <text x="90" y="150">丄</text>
            <text x="180" y="170">丄</text>
          </g>
          <path d="M75 75 Q 110 85 130 100" fill="none" stroke="#e11d48" stroke-width="2" stroke-dasharray="2" marker-end="url(#spoke-dislocation-state-arrow)" />
          <text x="35" y="205" font-size="9.5" fill="#94a3b8" font-family="var(--tz-font-ui)">{{ t('guidesSpokeDislocationMechanics.microstructure.stateAFooter') }}</text>

          <rect x="320" y="20" width="260" height="200" rx="10" fill="currentColor" opacity="0.04" stroke="#a7f3d0" stroke-width="2" />
          <text x="335" y="42" font-size="11" font-weight="800" fill="#059669">{{ t('guidesSpokeDislocationMechanics.microstructure.stateB') }}</text>
          <line x1="530" y1="50" x2="530" y2="190" stroke="#059669" stroke-width="4" />
          <g fill="#059669" font-size="14" font-weight="900">
            <text x="515" y="80">丄丄丄</text>
            <text x="515" y="120">丄丄丄</text>
            <text x="515" y="160">丄丄丄</text>
          </g>
          <circle cx="530" cy="85" r="4" fill="#d97706" />
          <circle cx="530" cy="125" r="4" fill="#d97706" />
          <circle cx="530" cy="165" r="4" fill="#d97706" />
          <text x="335" y="185" font-size="9.5" font-weight="700" fill="#059669">{{ t('guidesSpokeDislocationMechanics.microstructure.stateBFooter') }}</text>
          <text x="335" y="205" font-size="9.5" fill="#64748b" font-family="var(--tz-font-ui)">{{ t('guidesSpokeDislocationMechanics.microstructure.stateBNote') }}</text>
        </svg>
      </section>
    </div>

    <section class="spoke-dislocation-page__card">
      <div class="spoke-dislocation-page__card-heading">
        <h2>{{ t('guidesSpokeDislocationMechanics.calculator.title') }}</h2>
        <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--emerald">{{ t('guidesSpokeDislocationMechanics.calculator.tag') }}</span>
      </div>
      <p class="spoke-dislocation-page__body-copy">{{ t('guidesSpokeDislocationMechanics.calculator.intro') }}</p>

      <div class="spoke-dislocation-page__calculator">
        <div class="spoke-dislocation-page__calculator-inputs">
          <label>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.modelLabel') }}</span>
            <select v-model="selectedSpokeModel" @change="handleStainlessSteelSpokeModelSelection">
              <option v-for="option in spokeModelOptions" :key="option.id" :value="option.id">
                {{ option.label }}
              </option>
            </select>
          </label>
          <label>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.areaLabel') }}</span>
            <input
              v-if="isStainlessSteelSpokeCustomAreaEditable"
              v-model.number="customSpokeArea"
              type="number"
              step="0.01"
              inputmode="decimal"
              :min="calculationInputBounds?.effective_spoke_area_min_exclusive_mm2"
              :max="calculationInputBounds?.effective_spoke_area_max_mm2"
              @input="synchronizeStainlessSteelSpokeCustomAreaSelection"
            >
            <output
              v-else
              class="spoke-dislocation-page__calculator-area-value"
              :title="t('guidesSpokeDislocationMechanics.calculator.areaPresetNote')"
            >
              {{ selectedBackendModelAreaDisplay }}
            </output>
          </label>
          <label>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.lengthLabel') }}</span>
            <input
              v-model.number="effectiveSpokeLength"
              type="number"
              step="1"
              inputmode="decimal"
              :min="calculationInputBounds?.effective_spoke_length_min_exclusive_mm"
              :max="calculationInputBounds?.effective_spoke_length_max_mm"
            >
          </label>
          <label>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.tensionLabel') }}</span>
            <input
              v-model.number="workingTension"
              type="number"
              step="20"
              inputmode="numeric"
              :min="calculationInputBounds?.nominal_working_tension_min_exclusive_n"
              :max="calculationInputBounds?.nominal_working_tension_max_n"
            >
          </label>
          <label>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.ratioLabel') }}</span>
            <input
              v-model.number="overloadRatio"
              type="number"
              step="5"
              inputmode="numeric"
              :min="calculationInputBounds?.overload_ratio_min_percent"
              :max="calculationInputBounds?.overload_ratio_max_percent"
            >
          </label>
        </div>
        <p v-if="!isStainlessSteelSpokeCustomAreaEditable" class="spoke-dislocation-page__calculator-readonly-note">
          {{ t('guidesSpokeDislocationMechanics.calculator.areaPresetNote') }}
        </p>
        <p class="spoke-dislocation-page__calculator-input-note">
          {{ t('guidesSpokeDislocationMechanics.calculator.lengthNote') }}
        </p>

        <div class="spoke-dislocation-page__results" aria-live="polite">
          <div>
            <strong>{{ formattedBackendWorkingStress }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.workStressLabel') }}</span>
            <small>{{ formattedBackendWorkingYieldRatio }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--emerald">{{ formattedBackendOverloadForce }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.overloadLabel') }}</span>
            <small>{{ formattedBackendOverloadDelta }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--blue">{{ formattedBackendOverloadStress }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.overloadStressLabel') }}</span>
            <small>{{ t('guidesSpokeDislocationMechanics.calculator.overloadStressNote') }}</small>
          </div>
          <div>
            <strong :class="yieldRatioToneClass">{{ formattedBackendOverloadYieldRatio }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.yieldRatioLabel') }}</span>
            <small>{{ formattedBackendYieldMargin }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--blue">{{ formattedBackendWorkElasticElongation }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.workElongationLabel') }}</span>
            <small>{{ formattedBackendElasticElongationNote }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--emerald">{{ formattedBackendOverloadElasticElongation }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.overloadElongationLabel') }}</span>
            <small>{{ formattedBackendOverloadDeltaElasticElongation }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--blue">{{ formattedBackendWorkTotalElongation }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.workTotalElongationLabel') }}</span>
            <small>{{ t('guidesSpokeDislocationMechanics.calculator.totalElongationNote') }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--rose">{{ formattedBackendWorkPermanentElongation }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.workPermanentElongationLabel') }}</span>
            <small>{{ t('guidesSpokeDislocationMechanics.calculator.permanentElongationNote') }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--emerald">{{ formattedBackendOverloadTotalElongation }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.overloadTotalElongationLabel') }}</span>
            <small>{{ formattedBackendOverloadDeltaTotalElongation }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--rose">{{ formattedBackendOverloadPermanentElongation }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.overloadPermanentElongationLabel') }}</span>
            <small>{{ formattedBackendOverloadDeltaPermanentElongation }}</small>
          </div>
          <div>
            <strong class="spoke-dislocation-page__result-value--rose">{{ formattedBackendTotalElongationToFailure }}</strong>
            <span>{{ t('guidesSpokeDislocationMechanics.calculator.totalElongationToFailureLabel') }}</span>
            <small>{{ formattedBackendTotalElongationToFailureNote }}</small>
          </div>
        </div>

        <div class="spoke-dislocation-page__callout spoke-dislocation-page__callout--rose">
          <span aria-hidden="true">⛔</span>
          <div>
            <strong>{{ t('guidesSpokeDislocationMechanics.calculator.fractureReferenceTitle') }}</strong>
            <p>{{ t('guidesSpokeDislocationMechanics.calculator.fractureReferenceDescription', {
              percent: materialReference ? formatStainlessSteelSpokeNumericValue(materialReference.total_elongation_to_failure_percent, 1) : '—',
              length: calculationResult ? formatStainlessSteelSpokeNumericValue(calculationResult.effective_spoke_length_mm, 1) : '—',
              elongation: formattedBackendTotalElongationToFailure,
            }) }}</p>
          </div>
        </div>

        <div class="spoke-dislocation-page__callout spoke-dislocation-page__callout--safety" :class="`spoke-dislocation-page__callout--${safetyGate.tone}`">
          <span aria-hidden="true">{{ safetyGate.icon }}</span>
          <div>
            <strong>{{ safetyGate.title }}</strong>
            <p>{{ safetyGate.description }}</p>
          </div>
        </div>
      </div>

      <div class="spoke-dislocation-page__callout spoke-dislocation-page__callout--amber">
        <span aria-hidden="true">📐</span>
        <div>
          <strong>{{ t('guidesSpokeDislocationMechanics.calculator.scopeTitle') }}</strong>
          <p>{{ t('guidesSpokeDislocationMechanics.calculator.scopeDescription') }}</p>
        </div>
      </div>

      <div class="spoke-dislocation-page__callout spoke-dislocation-page__callout--blue">
        <span aria-hidden="true">🧭</span>
        <div>
          <strong>{{ t('guidesSpokeDislocationMechanics.calculator.rimReferenceTitle', backendReferenceTranslationValues) }}</strong>
          <p>{{ t('guidesSpokeDislocationMechanics.calculator.rimReferenceDescription', backendReferenceTranslationValues) }}</p>
        </div>
      </div>

      <div class="spoke-dislocation-page__boundary-notes">
        <p><strong>{{ t('guidesSpokeDislocationMechanics.calculator.boundaryLabel') }}</strong> {{ t('guidesSpokeDislocationMechanics.calculator.boundary', backendReferenceTranslationValues) }}</p>
        <p><strong>{{ t('guidesSpokeDislocationMechanics.calculator.torsionLabel') }}</strong> {{ t('guidesSpokeDislocationMechanics.calculator.torsion') }}</p>
      </div>

      <div v-if="materialReference" class="spoke-dislocation-page__callout spoke-dislocation-page__callout--blue">
        <span aria-hidden="true">🧾</span>
        <div>
          <strong>{{ t('guidesSpokeDislocationMechanics.calculator.materialReferenceTitle') }}</strong>
          <p>{{ t('guidesSpokeDislocationMechanics.calculator.materialReferenceDescription', {
            name: materialReference.material_name,
          }) }}</p>
          <p class="spoke-dislocation-page__material-basis">
            {{ t('guidesSpokeDislocationMechanics.calculator.materialReferenceBasis', {
              family: materialReference.material_family,
              condition: materialReference.manufacturing_condition,
              modulus: formatStainlessSteelSpokeNumericValue(materialReference.elastic_modulus_mpa, 0),
              yieldReference: formatStainlessSteelSpokeNumericValue(materialReference.yield_strength_mpa, 0),
              ultimateStrength: formatStainlessSteelSpokeNumericValue(materialReference.ultimate_tensile_strength_mpa, 0),
              totalElongationToFailure: formatStainlessSteelSpokeNumericValue(materialReference.total_elongation_to_failure_percent, 1),
            }) }}
          </p>
          <p class="spoke-dislocation-page__material-basis">
            <strong>{{ t('guidesSpokeDislocationMechanics.calculator.coldWorkHardeningTitle') }}</strong>
            {{ t('guidesSpokeDislocationMechanics.calculator.coldWorkHardeningDescription') }}
          </p>
          <dl class="spoke-dislocation-page__material-values">
            <div>
              <dt>{{ t('guidesSpokeDislocationMechanics.calculator.materialFamilyLabel') }}</dt>
              <dd>{{ materialReference.material_family }}</dd>
            </div>
            <div>
              <dt>{{ t('guidesSpokeDislocationMechanics.calculator.manufacturingConditionLabel') }}</dt>
              <dd>{{ materialReference.manufacturing_condition }}</dd>
            </div>
            <div>
              <dt>{{ t('guidesSpokeDislocationMechanics.calculator.curveTypeLabel') }}</dt>
              <dd>{{ t('guidesSpokeDislocationMechanics.calculator.curveTypeDescription') }}</dd>
            </div>
            <div>
              <dt>{{ t('guidesSpokeDislocationMechanics.calculator.elasticModulusLabel') }}</dt>
              <dd>{{ formatStainlessSteelSpokeNumericValue(materialReference.elastic_modulus_mpa, 0) }} MPa</dd>
            </div>
            <div>
              <dt>{{ t('guidesSpokeDislocationMechanics.calculator.yieldReferenceLabel') }}</dt>
              <dd>{{ formatStainlessSteelSpokeNumericValue(materialReference.yield_strength_mpa, 0) }} MPa</dd>
            </div>
            <div>
              <dt>{{ t('guidesSpokeDislocationMechanics.calculator.ultimateStrengthLabel') }}</dt>
              <dd>{{ formatStainlessSteelSpokeNumericValue(materialReference.ultimate_tensile_strength_mpa, 0) }} MPa</dd>
            </div>
            <div>
              <dt>{{ t('guidesSpokeDislocationMechanics.calculator.totalElongationToFailureLabel') }}</dt>
              <dd>{{ formatStainlessSteelSpokeNumericValue(materialReference.total_elongation_to_failure_percent, 1) }}%</dd>
            </div>
          </dl>
        </div>
      </div>
    </section>


    <section class="spoke-dislocation-page__card">
      <div class="spoke-dislocation-page__card-heading">
        <h2>{{ t('guidesSpokeDislocationMechanics.backendCatalog.title') }}</h2>
        <span class="spoke-dislocation-page__tag spoke-dislocation-page__tag--blue">{{ t('guidesSpokeDislocationMechanics.backendCatalog.tag') }}</span>
      </div>
      <p class="spoke-dislocation-page__body-copy">{{ t('guidesSpokeDislocationMechanics.backendCatalog.description', {
        workingTensionKgf: referenceCalculation ? formatStainlessSteelSpokeNumericValue(referenceCalculation.reference_calculation_tension_kgf, 0) : '180',
        length: referenceCalculation ? formatStainlessSteelSpokeNumericValue(referenceCalculation.effective_spoke_length_mm, 0) : '270',
      }) }}</p>
      <div v-if="referenceCalculation" class="spoke-dislocation-page__callout spoke-dislocation-page__callout--blue spoke-dislocation-page__reference-baseline">
        <span aria-hidden="true">📌</span>
        <p>{{ t('guidesSpokeDislocationMechanics.backendCatalog.referenceBaseline', {
          length: formatStainlessSteelSpokeNumericValue(referenceCalculation.effective_spoke_length_mm, 0),
          referenceWorkingTensionKgf: formatStainlessSteelSpokeNumericValue(referenceCalculation.reference_calculation_tension_kgf, 0),
          referenceWorkingTensionN: formatStainlessSteelSpokeNumericValue(referenceCalculation.reference_calculation_tension_n, 0),
          yieldStrength: formatStainlessSteelSpokeNumericValue(referenceCalculation.yield_strength_mpa, 0),
          curveEndpoint: formatStainlessSteelSpokeNumericValue(referenceCalculation.curve_endpoint_stress_mpa, 0),
          curveEndpointStrain: formatStainlessSteelSpokeNumericValue(referenceCalculation.curve_endpoint_total_strain_percent, 3),
          curveEndpointElongation: formatStainlessSteelSpokeNumericValue(referenceCalculation.curve_endpoint_total_elongation_mm, 3),
          totalElongationToFailurePercent: formatStainlessSteelSpokeNumericValue(referenceCalculation.total_elongation_to_failure_percent, 1),
          totalElongationToFailure: formatStainlessSteelSpokeNumericValue(referenceCalculation.fracture_reference_total_elongation_mm, 3),
        }) }}</p>
      </div>

      <div v-if="backendModelRows.length" class="spoke-dislocation-page__table-wrapper" tabindex="0">
        <table class="spoke-dislocation-page__table spoke-dislocation-page__backend-catalog-table">
          <caption class="spoke-dislocation-page__sr-only">{{ t('guidesSpokeDislocationMechanics.backendCatalog.title') }}</caption>
          <thead>
            <tr>
              <th scope="col">{{ t('guidesSpokeDislocationMechanics.backendCatalog.headers.model') }}</th>
              <th scope="col">{{ t('guidesSpokeDislocationMechanics.backendCatalog.headers.geometry') }}</th>
              <th scope="col">{{ t('guidesSpokeDislocationMechanics.backendCatalog.headers.area') }}</th>
              <th scope="col">{{ t('guidesSpokeDislocationMechanics.backendCatalog.headers.material') }}</th>
              <th scope="col">{{ t('guidesSpokeDislocationMechanics.backendCatalog.headers.yieldReferenceForce') }}</th>
              <th scope="col">{{ t('guidesSpokeDislocationMechanics.backendCatalog.headers.curveEndpointForce') }}</th>
              <th scope="col">{{ t('guidesSpokeDislocationMechanics.backendCatalog.headers.referenceElongation', {
                workingTensionKgf: referenceCalculation ? formatStainlessSteelSpokeNumericValue(referenceCalculation.reference_calculation_tension_kgf, 0) : '180',
              }) }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="model in backendModelRows"
              :key="model.id"
              :class="model.isCurrent ? 'spoke-dislocation-page__table-row--emerald' : undefined"
              :aria-current="model.isCurrent ? 'true' : undefined"
            >
              <th scope="row">
                <span class="spoke-dislocation-page__mono">{{ model.display_name }}</span>
              </th>
              <td>{{ model.geometry_description }}</td>
              <td class="spoke-dislocation-page__mono">{{ formatStainlessSteelSpokeNumericValue(model.effective_area_mm2, 2) }} mm²</td>
              <td>
                <span
                  class="spoke-dislocation-page__material-reference"
                  :title="model.materialName"
                  :aria-label="model.materialName"
                >
                  {{ formatStainlessSteelSpokeMaterialReferenceLabel(model.materialName) }}
                </span>
              </td>
              <td class="spoke-dislocation-page__mono">
                <template v-if="model.referenceYieldForceN !== null">
                  {{ formatStainlessSteelSpokeReferenceForce(model.referenceYieldForceN) }}
                </template>
                <span v-else class="spoke-dislocation-page__catalog-pending">{{ t('guidesSpokeDislocationMechanics.backendCatalog.selectModelToViewResult') }}</span>
              </td>
              <td class="spoke-dislocation-page__mono">
                <template v-if="model.referenceCurveEndpointForceN !== null">
                  {{ formatStainlessSteelSpokeReferenceForce(model.referenceCurveEndpointForceN) }}
                </template>
                <span v-else class="spoke-dislocation-page__catalog-pending">{{ t('guidesSpokeDislocationMechanics.backendCatalog.selectModelToViewResult') }}</span>
              </td>
              <td>
                <template v-if="model.referenceWorkingTotalElongationMM !== null">
                  <span class="spoke-dislocation-page__mono">{{ formatStainlessSteelSpokeNumericValue(model.referenceWorkingTotalElongationMM, 3) }} mm</span>
                  <small class="spoke-dislocation-page__table-note">{{ t('guidesSpokeDislocationMechanics.backendCatalog.selectedResultNote', {
                    workingTensionKgf: referenceCalculation ? formatStainlessSteelSpokeNumericValue(referenceCalculation.reference_calculation_tension_kgf, 0) : '180',
                    workingTensionN: referenceCalculation ? formatStainlessSteelSpokeNumericValue(referenceCalculation.reference_calculation_tension_n, 0) : '1765',
                    permanent: formatStainlessSteelSpokeNumericValue(model.referenceWorkingPermanentElongationMM, 3),
                    failure: formatStainlessSteelSpokeNumericValue(model.referenceFractureTotalElongationMM, 3),
                  }) }}</small>
                </template>
                <span v-else class="spoke-dislocation-page__catalog-pending">{{ t('guidesSpokeDislocationMechanics.backendCatalog.selectModelToViewResult') }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else class="spoke-dislocation-page__calculator-input-note">
        {{ t('guidesSpokeDislocationMechanics.backendCatalog.metadataUnavailable') }}
      </p>

      <div class="spoke-dislocation-page__callout spoke-dislocation-page__callout--blue">
        <span aria-hidden="true">🧮</span>
        <p><strong>{{ t('guidesSpokeDislocationMechanics.backendCatalog.boundaryLabel') }}</strong> {{ t('guidesSpokeDislocationMechanics.backendCatalog.boundary') }}</p>
      </div>
    </section>

  </article>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  useStainlessSteelSpokeDislocationMechanicsCalculation,
  type StainlessSteelSpokeModelReference,
} from '~/composables/useStainlessSteelSpokeDislocationMechanicsCalculation'

definePageMeta({
  layout: 'products',
  breadcrumbLabelKey: 'guidesSpokeDislocationMechanics.title',
  breadcrumbLabelFallback: 'Stainless Steel Spoke Axial Elongation Calculator',
  footerLabelKey: 'guidesSpokeDislocationMechanics.title',
  footerLabelFallback: 'Stainless Steel Spoke Axial Elongation Calculator',
  footerGroupLabelFallback: 'Spoke Guides',
  feedbackThreadKey: 'guides-spoke-dislocation-mechanics',
  feedbackTitleKey: 'guidesSpokeDislocationMechanics.feedbackTitle',
  feedbackTitle: 'Share your feedback about this spoke elongation calculator',
})

type SafetyTone = 'emerald' | 'amber' | 'rose'

const formatStainlessSteelSpokeNumericValue = (
  value: unknown,
  fractionDigits: number,
  fallback = '—',
) => {
  if (value === null || value === undefined || value === '') return fallback
  const numericValue = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(numericValue) ? numericValue.toFixed(fractionDigits) : fallback
}

const formatStainlessSteelSpokeMaterialReferenceLabel = (value: unknown) => {
  const materialReferenceText = String(value ?? '').trim()
  if (!materialReferenceText) return '—'

  const materialReferenceCharacters = Array.from(materialReferenceText)
  return materialReferenceCharacters.length > 10
    ? `${materialReferenceCharacters.slice(0, 10).join('')}…`
    : materialReferenceText
}

const formatStainlessSteelSpokeReferenceForce = (value: unknown) => {
  const numericValue = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(numericValue)) return '—'
  return `${numericValue.toFixed(0)} N ≈ ${(numericValue / 9.80665).toFixed(1)} kgf`
}

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesSpokeDislocationMechanics')

await loadPageMessages(locale.value)
watch(locale, nextLocale => void loadPageMessages(nextLocale))

useHead(() => ({
  title: t('guidesSpokeDislocationMechanics.seo.title'),
  meta: [
    { name: 'description', content: t('guidesSpokeDislocationMechanics.seo.description') },
    { property: 'og:type', content: 'website' },
    { property: 'og:title', content: t('guidesSpokeDislocationMechanics.seo.title') },
    { property: 'og:description', content: t('guidesSpokeDislocationMechanics.seo.description') },
  ],
  script: [{
    key: 'spoke-dislocation-mechanics-web-application-jsonld',
    type: 'application/ld+json',
    children: JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'WebApplication',
      name: t('guidesSpokeDislocationMechanics.seo.title'),
      description: t('guidesSpokeDislocationMechanics.seo.description'),
      applicationCategory: 'EngineeringApplication',
      operatingSystem: 'Web',
      isAccessibleForFree: true,
      inLanguage: locale.value,
    }),
  }],
}))

const selectedSpokeModel = ref<string>('cx-ray')
const customSpokeArea = ref<number | null>(null)
const effectiveSpokeLength = ref<number | null>(290)
const workingTension = ref<number | null>(1200)
const overloadRatio = ref<number | null>(135)
const isStainlessSteelSpokeCustomAreaEditable = computed(() => selectedSpokeModel.value === 'custom')

const {
  metadata: stainlessSteelSpokeDislocationMechanicsMetadata,
  calculationResult: stainlessSteelSpokeDislocationMechanicsCalculationResult,
  calculationPending: stainlessSteelSpokeDislocationMechanicsCalculationPending,
  calculationErrorKind: stainlessSteelSpokeDislocationMechanicsCalculationErrorKind,
} = await useStainlessSteelSpokeDislocationMechanicsCalculation(
  selectedSpokeModel,
  customSpokeArea,
  effectiveSpokeLength,
  workingTension,
  overloadRatio,
)

type SpokeModelOption = {
  id: string
  effectiveAreaMM2: number | null
  label: string
}

const spokeModelTranslationKeyByModelID: Record<string, string> = {
  'cx-ray': 'guidesSpokeDislocationMechanics.calculator.models.cxRay',
  'pillar-wing-20': 'guidesSpokeDislocationMechanics.calculator.models.pillarWing20',
  laser: 'guidesSpokeDislocationMechanics.calculator.models.laser',
  'd-light': 'guidesSpokeDislocationMechanics.calculator.models.dLight',
  race: 'guidesSpokeDislocationMechanics.calculator.models.race',
  leader: 'guidesSpokeDislocationMechanics.calculator.models.leader',
  'dt-aerolite': 'guidesSpokeDislocationMechanics.calculator.models.dtAerolite',
  'dt-aero-comp': 'guidesSpokeDislocationMechanics.calculator.models.dtAeroComp',
  'dt-revolite': 'guidesSpokeDislocationMechanics.calculator.models.dtRevolite',
  'dt-revolution': 'guidesSpokeDislocationMechanics.calculator.models.dtRevolution',
  'dt-competition': 'guidesSpokeDislocationMechanics.calculator.models.dtCompetition',
  'dt-champion': 'guidesSpokeDislocationMechanics.calculator.models.dtChampion',
}

const formatStainlessSteelSpokeModelOptionLabel = (model: StainlessSteelSpokeModelReference) => {
  const translationKey = spokeModelTranslationKeyByModelID[model.id]
  const translatedName = translationKey ? t(translationKey) : model.display_name
  return `${translatedName || '—'} · ${t('guidesSpokeDislocationMechanics.calculator.catalogModelMaterialSuffix')} (${model.geometry_description || '—'} / ${formatStainlessSteelSpokeNumericValue(model.effective_area_mm2, 2)} mm²)`
}

const spokeModelOptions = computed<SpokeModelOption[]>(() => {
  const backendModels = stainlessSteelSpokeDislocationMechanicsMetadata.value?.models || []
  const backendOptions: SpokeModelOption[] = backendModels.map(model => ({
    id: model.id,
    effectiveAreaMM2: model.effective_area_mm2,
    label: formatStainlessSteelSpokeModelOptionLabel(model),
  }))
  if (backendOptions.length === 0) {
    backendOptions.push({
      id: selectedSpokeModel.value,
      effectiveAreaMM2: null,
      label: t('guidesSpokeDislocationMechanics.calculator.metadataLoading'),
    })
  }
  backendOptions.push({
    id: 'custom',
    effectiveAreaMM2: null,
    label: t('guidesSpokeDislocationMechanics.calculator.models.custom'),
  })
  return backendOptions
})

watch(stainlessSteelSpokeDislocationMechanicsMetadata, metadata => {
  if (customSpokeArea.value !== null) return
  const defaultModel = metadata?.models.find(model => model.id === selectedSpokeModel.value)
  if (defaultModel) customSpokeArea.value = defaultModel.effective_area_mm2
}, { immediate: true })

const calculationResult = computed(() => stainlessSteelSpokeDislocationMechanicsCalculationResult.value)
const materialReference = computed(() => calculationResult.value?.material_reference
  || stainlessSteelSpokeDislocationMechanicsMetadata.value?.material_reference
  || null)
const referenceCalculation = computed(() => stainlessSteelSpokeDislocationMechanicsMetadata.value?.reference_calculation || null)
const referenceCalculationResultByModelID = computed(() => new Map(
  (referenceCalculation.value?.results || []).map(result => [result.model_id, result]),
))
const selectedBackendModelReference = computed(() => {
  const selectedModelFromMetadata = stainlessSteelSpokeDislocationMechanicsMetadata.value?.models.find(model => model.id === selectedSpokeModel.value)
  if (selectedModelFromMetadata) return selectedModelFromMetadata
  const selectedModelFromCalculation = calculationResult.value?.model_reference
  return selectedModelFromCalculation?.id === selectedSpokeModel.value ? selectedModelFromCalculation : null
})
const selectedBackendModelAreaDisplay = computed(() => {
  const effectiveAreaMM2 = selectedBackendModelReference.value?.effective_area_mm2
    ?? (calculationResult.value?.model_id === selectedSpokeModel.value ? calculationResult.value.effective_area_mm2 : null)
  return `${formatStainlessSteelSpokeNumericValue(effectiveAreaMM2, 2)} mm²`
})
const backendReferenceTranslationValues = computed(() => {
  const references = stainlessSteelSpokeDislocationMechanicsMetadata.value?.physical_references
  return {
    rimBaseline: formatStainlessSteelSpokeNumericValue(references?.carbon_rim_spoke_hole_warning_force_n, 0),
    yieldReference: formatStainlessSteelSpokeNumericValue(references?.yield_strength_mpa, 0),
    uts: formatStainlessSteelSpokeNumericValue(references?.ultimate_tensile_strength_mpa, 0),
  }
})
const backendModelRows = computed(() => {
  const currentResult = calculationResult.value
  const catalogMaterialName = stainlessSteelSpokeDislocationMechanicsMetadata.value?.material_reference?.material_name || '—'

  return (stainlessSteelSpokeDislocationMechanicsMetadata.value?.models || []).map(model => {
    const isCurrent = currentResult?.model_id === model.id
    const referenceResult = referenceCalculationResultByModelID.value.get(model.id)
    return {
      ...model,
      isCurrent,
      materialName: isCurrent ? currentResult?.material_reference.material_name || catalogMaterialName : catalogMaterialName,
      referenceWorkingStressMPA: referenceResult?.reference_calculation_stress_mpa ?? null,
      referenceWorkingTotalElongationMM: referenceResult?.reference_calculation_total_elongation_mm ?? null,
      referenceWorkingPermanentElongationMM: referenceResult?.reference_calculation_permanent_elongation_mm ?? null,
      referenceYieldForceN: referenceResult?.yield_reference_force_n ?? null,
      referenceCurveEndpointForceN: referenceResult?.curve_endpoint_force_n ?? null,
      referenceFractureTotalElongationMM: referenceResult?.fracture_reference_total_elongation_mm ?? null,
    }
  })
})

const hasBackendCalculationResult = computed(() => calculationResult.value !== null)
const calculationInputBounds = computed(() => stainlessSteelSpokeDislocationMechanicsMetadata.value?.input_bounds ?? null)
const formattedBackendMissingResult = computed(() => stainlessSteelSpokeDislocationMechanicsCalculationPending.value
  ? t('guidesSpokeDislocationMechanics.calculator.pendingResult')
  : t('guidesSpokeDislocationMechanics.calculator.unavailableResult'))

const formattedBackendWorkingStress = computed(() => calculationResult.value ? `${calculationResult.value.work_stress_mpa} MPa` : '—')
const formattedBackendWorkingYieldRatio = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.workYieldRatio', {
      ratio: formatStainlessSteelSpokeNumericValue(calculationResult.value.work_yield_ratio_percent, 1),
      yieldReference: backendReferenceTranslationValues.value.yieldReference,
    })
  : formattedBackendMissingResult.value)
const formattedBackendOverloadForce = computed(() => calculationResult.value ? `${calculationResult.value.overload_force_n} N` : '—')
const formattedBackendOverloadDelta = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.overloadDelta', { delta: calculationResult.value.overload_delta_force_n })
  : formattedBackendMissingResult.value)
const formattedBackendOverloadStress = computed(() => calculationResult.value ? `${calculationResult.value.overload_stress_mpa} MPa` : '—')
const formattedBackendOverloadYieldRatio = computed(() => calculationResult.value ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_yield_ratio_percent, 1)}%` : '—')
const formattedBackendYieldMargin = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.yieldMargin', { margin: formatStainlessSteelSpokeNumericValue(calculationResult.value.yield_margin_percent, 1) })
  : formattedBackendMissingResult.value)
const formattedBackendWorkElasticElongation = computed(() => calculationResult.value
  ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.work_elastic_elongation_mm, 3)} mm`
  : '—')
const formattedBackendOverloadElasticElongation = computed(() => calculationResult.value
  ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_elastic_elongation_mm, 3)} mm`
  : '—')
const formattedBackendOverloadDeltaElasticElongation = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.overloadDeltaElongation', { delta: formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_delta_elastic_elongation_mm, 3) })
  : formattedBackendMissingResult.value)
const formattedBackendWorkTotalElongation = computed(() => calculationResult.value
  ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.work_total_elongation_mm, 3)} mm`
  : '—')
const formattedBackendWorkPermanentElongation = computed(() => calculationResult.value
  ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.work_permanent_elongation_mm, 3)} mm`
  : '—')
const formattedBackendOverloadTotalElongation = computed(() => calculationResult.value
  ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_total_elongation_mm, 3)} mm`
  : '—')
const formattedBackendOverloadPermanentElongation = computed(() => calculationResult.value
  ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_permanent_elongation_mm, 3)} mm`
  : '—')
const formattedBackendTotalElongationToFailure = computed(() => calculationResult.value
  ? `${formatStainlessSteelSpokeNumericValue(calculationResult.value.fracture_reference_total_elongation_mm, 3)} mm`
  : '—')
const formattedBackendTotalElongationToFailureNote = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.totalElongationToFailureNote', {
      percent: formatStainlessSteelSpokeNumericValue(calculationResult.value.material_reference.total_elongation_to_failure_percent, 1),
      length: formatStainlessSteelSpokeNumericValue(calculationResult.value.effective_spoke_length_mm, 1),
    })
  : formattedBackendMissingResult.value)
const formattedBackendOverloadDeltaTotalElongation = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.overloadDeltaTotalElongation', { delta: formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_delta_total_elongation_mm, 3) })
  : formattedBackendMissingResult.value)
const formattedBackendOverloadDeltaPermanentElongation = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.overloadDeltaPermanentElongation', { delta: formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_delta_permanent_elongation_mm, 3) })
  : formattedBackendMissingResult.value)
const formattedBackendElasticElongationNote = computed(() => calculationResult.value
  ? t('guidesSpokeDislocationMechanics.calculator.elasticElongationNote', {
      length: formatStainlessSteelSpokeNumericValue(calculationResult.value.effective_spoke_length_mm, 1),
      area: formatStainlessSteelSpokeNumericValue(calculationResult.value.effective_area_mm2, 2),
      modulus: formatStainlessSteelSpokeNumericValue(calculationResult.value.elastic_modulus_mpa, 0),
      micrometres: formatStainlessSteelSpokeNumericValue(Number(calculationResult.value.work_elastic_elongation_mm) * 1000, 0),
    })
  : formattedBackendMissingResult.value)
const yieldRatioToneClass = computed(() => {
  if (!hasBackendCalculationResult.value) return 'spoke-dislocation-page__result-value--rose'
  if (calculationResult.value?.safety_level === 'critical') return 'spoke-dislocation-page__result-value--rose'
  if (calculationResult.value?.safety_level === 'warning') return 'spoke-dislocation-page__result-value--amber'
  return 'spoke-dislocation-page__result-value--emerald'
})

const safetyGate = computed(() => {
  if (stainlessSteelSpokeDislocationMechanicsCalculationPending.value) {
    return {
      tone: 'amber' as SafetyTone,
      icon: '⏳',
      title: t('guidesSpokeDislocationMechanics.calculator.safety.loadingTitle'),
      description: t('guidesSpokeDislocationMechanics.calculator.safety.loadingDescription'),
    }
  }

  if (!hasBackendCalculationResult.value) {
    const isInvalidInput = stainlessSteelSpokeDislocationMechanicsCalculationErrorKind.value === 'invalid-input'
    const backendStateDescription = t(isInvalidInput
      ? 'guidesSpokeDislocationMechanics.calculator.safety.invalidDescription'
      : 'guidesSpokeDislocationMechanics.calculator.safety.unavailableDescription')
    return {
      tone: 'rose' as SafetyTone,
      icon: '🛑',
      title: t(isInvalidInput
        ? 'guidesSpokeDislocationMechanics.calculator.safety.invalidTitle'
        : 'guidesSpokeDislocationMechanics.calculator.safety.unavailableTitle'),
      description: backendStateDescription,
    }
  }

  if (calculationResult.value?.safety_level === 'critical') {
    return {
      tone: 'rose' as SafetyTone,
      icon: '🛑',
      title: t('guidesSpokeDislocationMechanics.calculator.safety.criticalTitle'),
      description: t('guidesSpokeDislocationMechanics.calculator.safety.criticalDescription', {
        overload: calculationResult.value.overload_force_n,
        rimMargin: calculationResult.value.rim_warning_margin_n,
        stress: calculationResult.value.overload_stress_mpa,
        ratio: formatStainlessSteelSpokeNumericValue(calculationResult.value.overload_yield_ratio_percent, 1),
        rimBaseline: backendReferenceTranslationValues.value.rimBaseline,
      }),
    }
  }

  if (calculationResult.value?.safety_level === 'warning') {
    return {
      tone: 'amber' as SafetyTone,
      icon: '⚠️',
      title: t('guidesSpokeDislocationMechanics.calculator.safety.warningTitle'),
      description: t('guidesSpokeDislocationMechanics.calculator.safety.warningDescription', {
        overload: calculationResult.value.overload_force_n,
        stress: calculationResult.value.overload_stress_mpa,
      }),
    }
  }

  return {
    tone: 'emerald' as SafetyTone,
    icon: '✅',
    title: t('guidesSpokeDislocationMechanics.calculator.safety.safeTitle'),
    description: t('guidesSpokeDislocationMechanics.calculator.safety.safeDescription', {
      overload: calculationResult.value?.overload_force_n,
      margin: formatStainlessSteelSpokeNumericValue(calculationResult.value?.rim_warning_margin_percent, 1),
        rimBaseline: backendReferenceTranslationValues.value.rimBaseline,
    }),
  }
})

const handleStainlessSteelSpokeModelSelection = () => {
  if (selectedSpokeModel.value === 'custom') {
    customSpokeArea.value = null
    return
  }
  const selectedOption = spokeModelOptions.value.find(option => option.id === selectedSpokeModel.value)
  if (selectedOption?.effectiveAreaMM2 !== null && selectedOption?.effectiveAreaMM2 !== undefined) {
    customSpokeArea.value = selectedOption.effectiveAreaMM2
  }
}

const synchronizeStainlessSteelSpokeCustomAreaSelection = () => {
  const area = Number(customSpokeArea.value)
  const selectedOption = spokeModelOptions.value.find(option => option.id === selectedSpokeModel.value)
  if (selectedOption?.effectiveAreaMM2 !== null
    && selectedOption?.effectiveAreaMM2 !== undefined
    && Math.abs(selectedOption.effectiveAreaMM2 - area) < 0.001) {
    return
  }
  const matchingOption = spokeModelOptions.value.find(option => option.effectiveAreaMM2 !== null && Math.abs(option.effectiveAreaMM2 - area) < 0.001)
  selectedSpokeModel.value = matchingOption?.id || 'custom'
}
</script>

<style scoped>
.spoke-dislocation-page {
  --spoke-page-surface: var(--tz-surface-page, #f8fafc);
  --spoke-card-surface: var(--tz-card-surface, #ffffff);
  --spoke-muted-surface: var(--tz-surface-muted, #f1f5f9);
  --spoke-border: var(--tz-border-subtle, #cbd5e1);
  --spoke-border-solid: var(--tz-border-strong, #e2e8f0);
  --spoke-text: var(--tz-text-primary, #0f172a);
  --spoke-secondary: var(--tz-text-secondary, #475569);
  --spoke-muted: var(--tz-text-muted, #94a3b8);
  --spoke-emerald: #059669;
  --spoke-emerald-bg: #ecfdf5;
  --spoke-emerald-border: #a7f3d0;
  --spoke-blue: #2563eb;
  --spoke-blue-bg: #eff6ff;
  --spoke-blue-border: #bfdbfe;
  --spoke-amber: #d97706;
  --spoke-amber-bg: #fffbeb;
  --spoke-amber-border: #fde68a;
  --spoke-rose: #e11d48;
  --spoke-rose-bg: #fff1f2;
  --spoke-rose-border: #fecdd3;
  width: 100%;
  max-width: 1680px;
  margin: 0 auto;
  color: var(--spoke-text);
  font-size: 0.95rem;
}

.spoke-dislocation-page__hero,
.spoke-dislocation-page__card,
.spoke-dislocation-page__notice {
  border: 1px dashed var(--spoke-border);
  border-radius: 1.25rem;
  background: var(--spoke-card-surface);
  box-shadow: 0 1px 3px rgb(15 23 42 / 0.04);
}

.spoke-dislocation-page__hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
  padding: clamp(1.1rem, 2vw, 1.75rem);
}

.spoke-dislocation-page__eyebrow {
  margin: 0 0 0.35rem;
  color: var(--spoke-muted);
  font: 700 0.68rem/1.35 var(--tz-font-ui);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.spoke-dislocation-page h1,
.spoke-dislocation-page h2,
.spoke-dislocation-page h3,
.spoke-dislocation-page p {
  margin-top: 0;
}

.spoke-dislocation-page h1 {
  margin-bottom: 0.35rem;
  font-size: clamp(1.35rem, 2.5vw, 2rem);
  font-weight: 800;
  letter-spacing: -0.035em;
  line-height: 1.18;
}

.spoke-dislocation-page__subtitle {
  max-width: 58rem;
  margin-bottom: 0;
  color: var(--spoke-muted);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.spoke-dislocation-page__tags {
  display: flex;
  gap: 0.45rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.spoke-dislocation-page__tag {
  display: inline-flex;
  align-items: center;
  min-height: 1.65rem;
  border: 1px dashed;
  border-radius: 999px;
  padding: 0.25rem 0.55rem;
  font: 800 0.66rem/1.2 var(--tz-font-ui);
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.spoke-dislocation-page__tag--dark { border-color: var(--spoke-text); background: var(--spoke-text); color: #fff; }
.spoke-dislocation-page__tag--emerald { border-color: var(--spoke-emerald-border); background: var(--spoke-emerald-bg); color: var(--spoke-emerald); }
.spoke-dislocation-page__tag--blue { border-color: var(--spoke-blue-border); background: var(--spoke-blue-bg); color: var(--spoke-blue); }
.spoke-dislocation-page__tag--amber { border-color: var(--spoke-amber-border); background: var(--spoke-amber-bg); color: var(--spoke-amber); }

.spoke-dislocation-page__body-copy {
  margin-bottom: 0;
  font-size: 0.8rem;
  line-height: 1.6;
}

.spoke-dislocation-page__two-column {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.spoke-dislocation-page__card {
  margin-bottom: 1rem;
  padding: clamp(1rem, 2vw, 1.35rem);
}

.spoke-dislocation-page__two-column .spoke-dislocation-page__card {
  margin-bottom: 0;
}

.spoke-dislocation-page__load-ladder {
  display: grid;
  gap: 0.5rem;
  margin-top: 0.85rem;
}

.spoke-dislocation-page__load-step {
  display: grid;
  grid-template-columns: 5.8rem minmax(0, 1fr);
  align-items: center;
  gap: 0.75rem;
  border: 1px solid var(--spoke-border-solid);
  border-radius: 0.85rem;
  padding: 0.65rem 0.75rem;
}

.spoke-dislocation-page__load-step > strong {
  font: 800 1rem/1.1 var(--tz-font-ui);
  white-space: nowrap;
}

.spoke-dislocation-page__load-step b,
.spoke-dislocation-page__load-step span {
  display: block;
}

.spoke-dislocation-page__load-step b {
  font-size: 0.75rem;
  line-height: 1.35;
}

.spoke-dislocation-page__load-step span {
  margin-top: 0.15rem;
  color: var(--spoke-secondary);
  font-size: 0.7rem;
  line-height: 1.45;
}

.spoke-dislocation-page__load-step--blue { border-color: var(--spoke-blue-border); background: var(--spoke-blue-bg); }
.spoke-dislocation-page__load-step--blue > strong { color: var(--spoke-blue); }
.spoke-dislocation-page__load-step--emerald { border-color: var(--spoke-emerald-border); background: var(--spoke-emerald-bg); }
.spoke-dislocation-page__load-step--emerald > strong { color: var(--spoke-emerald); }
.spoke-dislocation-page__load-step--slate { background: var(--spoke-muted-surface); }
.spoke-dislocation-page__load-step--amber { border-color: var(--spoke-amber-border); background: var(--spoke-amber-bg); }
.spoke-dislocation-page__load-step--amber > strong { color: var(--spoke-amber); }

.spoke-dislocation-page__microstructure-diagram {
  display: block;
  width: 100%;
  height: auto;
  max-height: 22rem;
  margin-top: 0.85rem;
  border: 1px solid var(--spoke-border-solid);
  border-radius: 0.85rem;
  background: var(--spoke-muted-surface);
  color: var(--spoke-secondary);
}

.spoke-dislocation-page__card-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.8rem;
  padding-bottom: 0.7rem;
  border-bottom: 1px solid var(--spoke-border-solid);
}

.spoke-dislocation-page__card-heading h2 {
  margin-bottom: 0;
  font-size: clamp(1rem, 1.5vw, 1.15rem);
  font-weight: 800;
  letter-spacing: -0.02em;
  line-height: 1.3;
}

.spoke-dislocation-page__callout {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  margin-top: 0.85rem;
  border: 1px dashed;
  border-radius: 0.85rem;
  padding: 0.75rem 0.85rem;
  font-size: 0.8rem;
  line-height: 1.55;
}

.spoke-dislocation-page__callout p { margin-bottom: 0; }
.spoke-dislocation-page__callout--emerald { border-color: var(--spoke-emerald-border); background: var(--spoke-emerald-bg); color: #065f46; }
.spoke-dislocation-page__callout--blue { border-color: var(--spoke-blue-border); background: var(--spoke-blue-bg); color: #1e40af; }
.spoke-dislocation-page__callout--amber { border-color: var(--spoke-amber-border); background: var(--spoke-amber-bg); color: #92400e; }
.spoke-dislocation-page__callout--rose { border-color: var(--spoke-rose-border); background: var(--spoke-rose-bg); color: #9f1239; }

.spoke-dislocation-page__table-wrapper {
  overflow-x: auto;
  border: 1px solid var(--spoke-border-solid);
  border-radius: 0.85rem;
}

.spoke-dislocation-page__table {
  width: 100%;
  min-width: 58rem;
  border-collapse: collapse;
  font-size: 0.78rem;
  text-align: left;
}

.spoke-dislocation-page__table th,
.spoke-dislocation-page__table td {
  border-bottom: 1px solid var(--spoke-border-solid);
  padding: 0.65rem 0.75rem;
  vertical-align: top;
}

.spoke-dislocation-page__table thead th {
  background: var(--spoke-muted-surface);
  color: var(--spoke-secondary);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  white-space: nowrap;
}

.spoke-dislocation-page__table tbody tr:last-child th,
.spoke-dislocation-page__table tbody tr:last-child td { border-bottom: 0; }
.spoke-dislocation-page__table-row--emerald { background: var(--spoke-emerald-bg); }
.spoke-dislocation-page__table-row--rose { background: var(--spoke-rose-bg); }
.spoke-dislocation-page__mono { font-family: var(--tz-font-ui); }
.spoke-dislocation-page__material-reference {
  display: inline-block;
  max-width: 11ch;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
  white-space: nowrap;
}
.spoke-dislocation-page__table-note {
  display: block;
  margin-top: 0.22rem;
  color: var(--spoke-muted);
  font: 700 0.65rem/1.35 var(--tz-font-ui);
  overflow-wrap: anywhere;
}
.spoke-dislocation-page__catalog-pending {
  color: var(--spoke-muted);
  font-size: 0.72rem;
  line-height: 1.4;
}
.spoke-dislocation-page__backend-catalog-table {
  min-width: 52rem;
}

.spoke-dislocation-page__calculator {
  border: 1px dashed var(--spoke-border);
  border-radius: 1rem;
  background: var(--spoke-muted-surface);
  padding: 1rem;
}

.spoke-dislocation-page__calculator-inputs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem 1.25rem;
}

.spoke-dislocation-page__calculator-inputs label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--spoke-secondary);
  font-size: 0.76rem;
  font-weight: 750;
}

.spoke-dislocation-page__calculator-inputs select,
.spoke-dislocation-page__calculator-inputs input {
  min-width: 7rem;
  border: 1px solid var(--spoke-border-solid);
  border-radius: 0.55rem;
  background: var(--spoke-card-surface);
  color: var(--spoke-text);
  padding: 0.45rem 0.55rem;
  font: 700 0.78rem/1.2 var(--tz-font-ui);
}

.spoke-dislocation-page__calculator-inputs select { max-width: 22rem; }
.spoke-dislocation-page__calculator-inputs input { width: 7rem; text-align: right; }
.spoke-dislocation-page__calculator-area-value {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  min-width: 7rem;
  border: 1px solid var(--spoke-border-solid);
  border-radius: 0.55rem;
  background: var(--spoke-muted-surface);
  color: var(--spoke-secondary);
  padding: 0.45rem 0.55rem;
  font: 700 0.78rem/1.2 var(--tz-font-ui);
  text-align: right;
}
.spoke-dislocation-page__calculator-readonly-note {
  margin: 0.55rem 0 0;
  color: var(--spoke-muted);
  font: 700 0.7rem/1.5 var(--tz-font-ui);
}

.spoke-dislocation-page__calculator-input-note {
  margin: 0.55rem 0 0;
  color: var(--spoke-secondary);
  font: 700 0.7rem/1.5 var(--tz-font-ui);
}

.spoke-dislocation-page__results {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 0.65rem;
  margin-top: 0.85rem;
  border: 1px solid var(--spoke-border-solid);
  border-radius: 0.85rem;
  background: var(--spoke-card-surface);
  padding: 0.85rem;
  text-align: center;
}

.spoke-dislocation-page__results > div { min-width: 0; }
.spoke-dislocation-page__results strong { display: block; font: 900 1.1rem/1.15 var(--tz-font-ui); }
.spoke-dislocation-page__results span { display: block; margin-top: 0.2rem; color: var(--spoke-muted); font-size: 0.65rem; font-weight: 800; line-height: 1.35; text-transform: uppercase; }
.spoke-dislocation-page__results small { display: block; margin-top: 0.3rem; color: var(--spoke-muted); font: 700 0.65rem/1.35 var(--tz-font-ui); }
.spoke-dislocation-page__result-value--emerald { color: var(--spoke-emerald); }
.spoke-dislocation-page__result-value--blue { color: var(--spoke-blue); }
.spoke-dislocation-page__result-value--amber { color: var(--spoke-amber); }
.spoke-dislocation-page__result-value--rose { color: var(--spoke-rose); }

.spoke-dislocation-page__callout--safety { margin-top: 0.85rem; }
.spoke-dislocation-page__callout--safety strong { display: block; margin-bottom: 0.15rem; }
.spoke-dislocation-page__callout--safety p { font-size: 0.78rem; }

.spoke-dislocation-page__boundary-notes {
  margin-top: 0.85rem;
  color: var(--spoke-muted);
  font: 700 0.7rem/1.6 var(--tz-font-ui);
}

.spoke-dislocation-page__boundary-notes p { margin-bottom: 0.35rem; }
.spoke-dislocation-page__boundary-notes p:last-child { margin-bottom: 0; }
.spoke-dislocation-page__material-basis {
  margin: 0.55rem 0 0;
  color: var(--spoke-secondary);
  font-size: 0.78rem;
  line-height: 1.65;
}
.spoke-dislocation-page__material-values {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  gap: 0.55rem;
  margin: 0.75rem 0 0;
}
.spoke-dislocation-page__material-values div {
  padding: 0.55rem 0.65rem;
  border: 1px solid var(--spoke-border-solid);
  border-radius: 0.65rem;
  background: var(--spoke-muted-surface);
}
.spoke-dislocation-page__material-values dt {
  color: var(--spoke-muted);
  font-size: 0.66rem;
  font-weight: 800;
  line-height: 1.35;
}
.spoke-dislocation-page__material-values dd {
  margin: 0.2rem 0 0;
  color: var(--spoke-text);
  font-size: 0.76rem;
  line-height: 1.45;
}
.spoke-dislocation-page__notice { margin-top: 0; border-color: var(--spoke-amber-border); background: var(--spoke-amber-bg); color: #92400e; padding: 0.9rem 1rem; }
.spoke-dislocation-page__notice strong { display: block; margin-bottom: 0.2rem; }
.spoke-dislocation-page__notice p { margin-bottom: 0; font-size: 0.8rem; line-height: 1.6; }
.spoke-dislocation-page__sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }

@media (max-width: 1050px) {
  .spoke-dislocation-page__two-column { grid-template-columns: 1fr; }
  .spoke-dislocation-page__two-column .spoke-dislocation-page__card { margin-bottom: 1rem; }
}

@media (max-width: 720px) {
  .spoke-dislocation-page__hero { align-items: flex-start; }
  .spoke-dislocation-page__tags { justify-content: flex-start; }
  .spoke-dislocation-page__calculator-inputs,
  .spoke-dislocation-page__results { grid-template-columns: 1fr 1fr; }
  .spoke-dislocation-page__calculator-inputs label { align-items: flex-start; flex-direction: column; }
  .spoke-dislocation-page__calculator-inputs select,
  .spoke-dislocation-page__calculator-inputs input { width: 100%; max-width: none; }
}

@media (max-width: 480px) {
  .spoke-dislocation-page__calculator-inputs,
  .spoke-dislocation-page__results { grid-template-columns: 1fr; }
  .spoke-dislocation-page__card-heading { align-items: flex-start; flex-direction: column; }
}
</style>
