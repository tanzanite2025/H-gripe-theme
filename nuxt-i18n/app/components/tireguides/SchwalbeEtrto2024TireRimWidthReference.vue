<template>
  <div class="schwalbe-width-reference">
    <header class="schwalbe-width-reference__header">
      <h3>{{ t('guidesTireChoose.schwalbeReference.title') }}</h3>
      <p class="schwalbe-width-reference__source">
        {{ t('guidesTireChoose.schwalbeReference.source') }}
      </p>
      <p>{{ t('guidesTireChoose.schwalbeReference.description') }}</p>
    </header>

    <aside class="schwalbe-width-reference__notice" aria-labelledby="schwalbe-hookless-restriction-title">
      <h4 id="schwalbe-hookless-restriction-title">
        {{ t('guidesTireChoose.schwalbeReference.hooklessTitle') }}
      </h4>
      <p>{{ t('guidesTireChoose.schwalbeReference.hooklessRestriction') }}</p>
    </aside>

    <div id="schwalbe-etrto-2024-width-table" class="schwalbe-width-reference__table-container">
      <p class="schwalbe-width-reference__table-heading">
        {{ t('guidesTireChoose.schwalbeReference.tableCaption') }}
      </p>
      <div class="schwalbe-width-reference__table-columns">
        <table
          v-for="(rows, columnIndex) in schwalbeReferenceTableColumnRows"
          :key="columnIndex"
          class="schwalbe-width-reference__table"
        >
          <caption class="sr-only">{{ t('guidesTireChoose.schwalbeReference.tableCaption') }}</caption>
          <thead :class="{ 'schwalbe-width-reference__continuation-header': columnIndex > 0 }">
            <tr>
              <th scope="col">{{ t('guidesTireChoose.schwalbeReference.tireWidthHeader') }}</th>
              <th scope="col">{{ t('guidesTireChoose.schwalbeReference.rimWidthHeader') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.tireWidthRangeMm.min">
              <th scope="row">{{ formatTireRimWidthReferenceRange(row.tireWidthRangeMm) }}</th>
              <td>{{ formatTireRimWidthReferenceRange(row.possibleRimInnerWidthRangeMm) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <p class="schwalbe-width-reference__manufacturer-note">
      {{ t('guidesTireChoose.schwalbeReference.manufacturerNote') }}
    </p>

    <section aria-labelledby="schwalbe-width-reference-original-page-title">
      <h4 id="schwalbe-width-reference-original-page-title" class="schwalbe-width-reference__image-title">
        {{ t('guidesTireChoose.schwalbeReference.originalPageTitle') }}
      </h4>
      <figure class="schwalbe-width-reference__figure">
        <img
          :src="referenceData.sourceImage.path"
          :width="referenceData.sourceImage.width"
          :height="referenceData.sourceImage.height"
          :alt="t('guidesTireChoose.schwalbeReference.originalPageAlt')"
          loading="lazy"
          decoding="async"
        />
        <figcaption>{{ t('guidesTireChoose.schwalbeReference.originalPageCaption') }}</figcaption>
      </figure>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { schwalbeEtrto2024TireRimWidthReferenceData as referenceData } from '~/data/tireguides/schwalbeEtrto2024TireRimWidthReferenceData'
import { formatTireRimWidthReferenceRange } from '~/data/tireguides/tireRimWidthReferencePresentation'

const { t } = useI18n()
const schwalbeReferenceFirstColumnRowCount = Math.ceil(referenceData.rows.length / 2)
const schwalbeReferenceTableColumnRows = [
  referenceData.rows.slice(0, schwalbeReferenceFirstColumnRowCount),
  referenceData.rows.slice(schwalbeReferenceFirstColumnRowCount),
]
</script>

<style scoped>
.schwalbe-width-reference {
  width: 100%;
  min-width: 0;
  margin-top: 1.5rem;
  color: var(--tz-text-secondary);
  text-align: left;
  font-size: 0.875rem;
  line-height: 1.65;
}

.schwalbe-width-reference h3,
.schwalbe-width-reference h4 {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.schwalbe-width-reference p {
  text-align: left;
}

.schwalbe-width-reference__header h3 {
  font-size: 1.125rem;
}

.schwalbe-width-reference__header p {
  margin-top: 0.65rem;
}

.schwalbe-width-reference__source {
  color: var(--tz-text-muted);
  font-size: 0.8rem;
}

.schwalbe-width-reference__notice {
  margin: 1.25rem 0;
  border: 1px solid var(--tz-border-subtle);
  border-left: 3px solid var(--tz-site-accent);
  border-radius: 0.75rem;
  padding: 0.85rem 1rem;
  background: var(--tz-form-panel-surface);
}

.schwalbe-width-reference__notice p {
  margin-top: 0.35rem;
}

.schwalbe-width-reference__table-container {
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
}

.schwalbe-width-reference__table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-variant-numeric: tabular-nums;
}

.schwalbe-width-reference__table-heading {
  padding: 0.75rem 1rem;
  color: var(--tz-text-muted);
  font-size: 0.8rem;
  text-align: left;
}

.schwalbe-width-reference__continuation-header {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.schwalbe-width-reference__table th,
.schwalbe-width-reference__table td {
  border-top: 1px solid var(--tz-border-subtle);
  padding: 0.6rem 1rem;
  text-align: left;
  overflow-wrap: anywhere;
}

.schwalbe-width-reference__table thead {
  background: var(--tz-form-panel-surface);
  color: var(--tz-text-primary);
  font-size: 0.8rem;
}

.schwalbe-width-reference__table tbody th {
  color: var(--tz-text-primary);
  font-weight: 600;
}

.schwalbe-width-reference__table tbody tr:nth-child(even) {
  background: var(--tz-form-panel-surface);
}

@media (min-width: 1024px) {
  .schwalbe-width-reference__table-columns {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-items: start;
  }

  .schwalbe-width-reference__table + .schwalbe-width-reference__table {
    border-inline-start: 1px solid var(--tz-border-subtle);
  }

  .schwalbe-width-reference__continuation-header {
    position: static;
    width: auto;
    height: auto;
    margin: 0;
    overflow: visible;
    clip: auto;
    white-space: normal;
  }
}

.schwalbe-width-reference__manufacturer-note {
  margin-top: 0.75rem;
  color: var(--tz-text-muted);
  font-size: 0.8rem;
}

.schwalbe-width-reference__image-title {
  margin: 1.75rem 0 0.75rem;
}

.schwalbe-width-reference__figure {
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
}

.schwalbe-width-reference__figure img {
  display: block;
  width: 100%;
  height: auto;
  background: white;
}

.schwalbe-width-reference__figure figcaption {
  padding: 0.65rem 1rem;
  color: var(--tz-text-muted);
  font-size: 0.8rem;
}

@media (max-width: 640px) {
  .schwalbe-width-reference__table th,
  .schwalbe-width-reference__table td,
  .schwalbe-width-reference__table-heading {
    padding-right: 0.75rem;
    padding-left: 0.75rem;
  }
}
</style>
