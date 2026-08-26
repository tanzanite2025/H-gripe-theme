import type { Component } from 'vue'
import { AtSign, Share2, Video } from '@lucide/vue'

export type SocialProvider = 'meta' | 'x' | 'youtube' | 'reddit'

export type SocialFieldKey = 'facebook' | 'instagram' | 'x' | 'youtube' | 'reddit'

export interface SocialPlatformDefinition {
  provider: SocialProvider
  label: string
  overviewDescription: string
  description: string
  capability: string
  icon: Component
  routeName: string
}

export const socialPlatforms: Record<SocialProvider, SocialPlatformDefinition> = {
  youtube: {
    provider: 'youtube',
    label: 'YouTube',
    overviewDescription: '绑定频道并发布视频',
    description: '绑定 Google 账号和 YouTube Channel，为后续视频发布提供账号上下文。',
    capability: '视频发布',
    icon: Video,
    routeName: 'SocialYouTube',
  },
  meta: {
    provider: 'meta',
    label: 'Facebook / Instagram',
    overviewDescription: '管理 Meta 平台账号',
    description: '统一维护 Meta 账号连接，并显示 Facebook Page 与 Instagram Professional Account。',
    capability: '图文 / 视频',
    icon: Share2,
    routeName: 'SocialMeta',
  },
  x: {
    provider: 'x',
    label: 'X',
    overviewDescription: '管理品牌内容分发',
    description: '为品牌内容、媒体资源和发布记录预留平台连接能力。',
    capability: '内容发布',
    icon: AtSign,
    routeName: 'SocialX',
  },
  reddit: {
    provider: 'reddit',
    label: 'Reddit',
    overviewDescription: '管理 Reddit 账号连接',
    description: '绑定 Reddit 账号，为社区内容发布和账号状态管理提供连接。',
    capability: '社区内容',
    icon: Share2,
    routeName: 'SocialReddit',
  },
}

export const socialProviderList = Object.keys(socialPlatforms) as SocialProvider[]

export interface SocialFieldDefinition {
  key: SocialFieldKey
  label: string
  placeholder: string
}

export const socialFields: readonly SocialFieldDefinition[] = [
  { key: 'facebook', label: 'Facebook', placeholder: 'Facebook 页面 URL' },
  { key: 'instagram', label: 'Instagram', placeholder: '账号 URL' },
  { key: 'x', label: 'X', placeholder: '账号 URL' },
  { key: 'youtube', label: 'YouTube', placeholder: '频道 URL' },
  { key: 'reddit', label: 'Reddit', placeholder: '社区或账号 URL' },
]
