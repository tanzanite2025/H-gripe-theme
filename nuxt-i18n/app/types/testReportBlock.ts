export type RimTestReportBlockId =
  | 'rim-strength-test'
  | 'rim-impact-test'
  | 'rim-inflation-test'
  | 'rim-radial-load-test'
  | 'rim-spoke-hole-tension-test'
  | 'rim-cross-section-inspection'
  | 'rim-xray-flatness-inspection'

export type TestReportBlockId = RimTestReportBlockId

export interface TestReportBlockPresentationProps {
  headingTag?: 'h2' | 'h3' | 'h4'
  usePageAnchor?: boolean
  presentation?: 'standalone' | 'dialog'
}
