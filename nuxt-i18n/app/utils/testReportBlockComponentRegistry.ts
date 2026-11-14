import { defineAsyncComponent, type Component } from 'vue'
import type { TestReportBlockId } from '~/types/testReportBlock'

interface TestReportBlockRegistration {
  component: Component
  messageNamespace: string
  titleMessageKey: string
}

// 按钮和弹窗只依赖注册信息，不读取某一类测试的素材或业务定义。
export const testReportBlockRegistrationsById = {
  'rim-strength-test': {
    component: defineAsyncComponent(() => import('~/components/technology/test-reports/rim/RimStrengthTestReportBlock.vue')),
    messageNamespace: 'testReportRim',
    titleMessageKey: 'testReportRim.tests.strength.title',
  },
  'rim-impact-test': {
    component: defineAsyncComponent(() => import('~/components/technology/test-reports/rim/RimImpactTestReportBlock.vue')),
    messageNamespace: 'testReportRim',
    titleMessageKey: 'testReportRim.tests.impact.title',
  },
  'rim-inflation-test': {
    component: defineAsyncComponent(() => import('~/components/technology/test-reports/rim/RimInflationTestReportBlock.vue')),
    messageNamespace: 'testReportRim',
    titleMessageKey: 'testReportRim.tests.inflation.title',
  },
  'rim-radial-load-test': {
    component: defineAsyncComponent(() => import('~/components/technology/test-reports/rim/RimRadialLoadTestReportBlock.vue')),
    messageNamespace: 'testReportRim',
    titleMessageKey: 'testReportRim.tests.radial.title',
  },
  'rim-spoke-hole-tension-test': {
    component: defineAsyncComponent(() => import('~/components/technology/test-reports/rim/RimSpokeHoleTensionTestReportBlock.vue')),
    messageNamespace: 'testReportRim',
    titleMessageKey: 'testReportRim.tests.holeTension.title',
  },
  'rim-cross-section-inspection': {
    component: defineAsyncComponent(() => import('~/components/technology/test-reports/rim/RimCrossSectionInspectionReportBlock.vue')),
    messageNamespace: 'testReportRim',
    titleMessageKey: 'testReportRim.tests.cuttingPieces.title',
  },
  'rim-xray-flatness-inspection': {
    component: defineAsyncComponent(() => import('~/components/technology/test-reports/rim/RimXRayFlatnessInspectionReportBlock.vue')),
    messageNamespace: 'testReportRim',
    titleMessageKey: 'testReportRim.tests.xray.title',
  },
} satisfies Record<TestReportBlockId, TestReportBlockRegistration>
