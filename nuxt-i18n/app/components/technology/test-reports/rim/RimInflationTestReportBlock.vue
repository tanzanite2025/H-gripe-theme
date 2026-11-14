<template>
  <RimTestDescriptionVideoAndFourPhotoGridLayout v-bind="props" :test-definition="inflationTestDefinition">
    <template #header-actions><slot name="header-actions" /></template>
  </RimTestDescriptionVideoAndFourPhotoGridLayout>
</template>

<script setup lang="ts">
import RimTestDescriptionVideoAndFourPhotoGridLayout from '~/components/RimTestDescriptionVideoAndFourPhotoGridLayout.vue'
import { getRimTestReportTestDefinition } from '~/data/rimTestReportTestDefinitionsAndMediaConfiguration'
import type { TestReportBlockPresentationProps } from '~/types/testReportBlock'

const props = defineProps<TestReportBlockPresentationProps>()
const inflationTestDefinition = getRimTestReportTestDefinition('rim-inflation-test')
</script>
