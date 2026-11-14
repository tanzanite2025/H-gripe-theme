import type { RimTestReportBlockId } from '~/types/testReportBlock'

export interface RimTestReportVideoAndFourPhotoMediaConfiguration {
  videoSourcePath: string | null
  videoPosterPath: string | null
  photoSourcePaths: readonly [string | null, string | null, string | null, string | null]
}

export interface RimTestReportTestDefinition {
  id: 'strength' | 'impact' | 'inflation' | 'radial' | 'holeTension' | 'cuttingPieces' | 'xray'
  sectionId: RimTestReportBlockId
  hasStandard: boolean
  methodCount: number
  criteriaCount: number
  mediaConfiguration: RimTestReportVideoAndFourPhotoMediaConfiguration
}

// 素材按 public/testreport/rimtestreport/<测试项目>/videos 或 photos 分开存放。
// 填写 public 之后的访问路径；未提供的素材保留 null，页面显示占位。
// photoSourcePaths 的四个位置依次对应左上、右上、左下、右下，不要删除缺图位置。
export const rimTestReportTestDefinitions: readonly RimTestReportTestDefinition[] = [
  {
    id: 'strength',
    sectionId: 'rim-strength-test',
    hasStandard: true,
    methodCount: 3,
    criteriaCount: 3,
    mediaConfiguration: {
      videoSourcePath: null,
      videoPosterPath: null,
      photoSourcePaths: [
        '/testreport/rimtestreport/strength-test/photos/rim-strength-test-photo-1.webp',
        '/testreport/rimtestreport/strength-test/photos/rim-strength-test-photo-2.webp',
        null,
        null,
      ],
    },
  },
  {
    id: 'impact',
    sectionId: 'rim-impact-test',
    hasStandard: true,
    methodCount: 3,
    criteriaCount: 3,
    mediaConfiguration: {
      // 视频：/testreport/rimtestreport/impact-test/videos/rim-impact-test-video.webm
      // 封面：/testreport/rimtestreport/impact-test/videos/rim-impact-test-video-poster.webp
      videoSourcePath: null,
      videoPosterPath: null,
      // 照片：/testreport/rimtestreport/impact-test/photos/rim-impact-test-photo-1.webp（其余为 2、3、4）
      photoSourcePaths: [null, null, null, null],
    },
  },
  {
    id: 'inflation',
    sectionId: 'rim-inflation-test',
    hasStandard: false,
    methodCount: 4,
    criteriaCount: 4,
    mediaConfiguration: {
      // 视频：/testreport/rimtestreport/inflation-test/videos/rim-inflation-test-video.webm
      // 封面：/testreport/rimtestreport/inflation-test/videos/rim-inflation-test-video-poster.webp
      videoSourcePath: null,
      videoPosterPath: null,
      // 照片：/testreport/rimtestreport/inflation-test/photos/rim-inflation-test-photo-1.webp（其余为 2、3、4）
      photoSourcePaths: [null, null, null, null],
    },
  },
  {
    id: 'radial',
    sectionId: 'rim-radial-load-test',
    hasStandard: true,
    methodCount: 3,
    criteriaCount: 3,
    mediaConfiguration: {
      videoSourcePath: null,
      videoPosterPath: null,
      photoSourcePaths: [
        '/testreport/rimtestreport/radial-load-test/photos/rim-radial-load-test-photo-1.webp',
        null,
        null,
        null,
      ],
    },
  },
  {
    id: 'holeTension',
    sectionId: 'rim-spoke-hole-tension-test',
    hasStandard: true,
    methodCount: 3,
    criteriaCount: 3,
    mediaConfiguration: {
      videoSourcePath: '/testreport/rimtestreport/spoke-hole-tension-test/videos/rim-spoke-hole-tension-test-video.webm',
      videoPosterPath: '/testreport/rimtestreport/spoke-hole-tension-test/photos/rim-spoke-hole-tension-test-photo-1.webp',
      photoSourcePaths: [
        '/testreport/rimtestreport/spoke-hole-tension-test/photos/rim-spoke-hole-tension-test-photo-1.webp',
        null,
        null,
        null,
      ],
    },
  },
  {
    id: 'cuttingPieces',
    sectionId: 'rim-cross-section-inspection',
    hasStandard: false,
    methodCount: 4,
    criteriaCount: 4,
    mediaConfiguration: {
      videoSourcePath: null,
      videoPosterPath: null,
      photoSourcePaths: [
        '/testreport/rimtestreport/rim-cross-section-inspection/photos/rim-cross-section-inspection-photo-1.webp',
        null,
        null,
        null,
      ],
    },
  },
  {
    id: 'xray',
    sectionId: 'rim-xray-flatness-inspection',
    hasStandard: false,
    methodCount: 4,
    criteriaCount: 4,
    mediaConfiguration: {
      videoSourcePath: null,
      videoPosterPath: null,
      photoSourcePaths: [
        '/testreport/rimtestreport/xray-flatness-inspection/photos/rim-xray-flatness-inspection-photo-1.webp',
        null,
        null,
        null,
      ],
    },
  },
]

export function getRimTestReportTestDefinition(testReportBlockId: RimTestReportBlockId): RimTestReportTestDefinition {
  const testDefinition = rimTestReportTestDefinitions.find(definition => definition.sectionId === testReportBlockId)
  if (!testDefinition) throw new Error(`Rim test report definition is missing: ${testReportBlockId}`)
  return testDefinition
}
