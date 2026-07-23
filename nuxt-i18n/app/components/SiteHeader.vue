<template>
	<div ref="headerRootRef" class="fixed top-0 md:top-1.5 left-0 md:left-1/2 md:-translate-x-1/2 w-full md:w-[95vw] md:max-w-[1200px] z-[110] site-header-root">
		<div
			class="relative w-full rounded-none md:rounded-[30px] bg-[radial-gradient(circle_at_top_left,rgba(15,23,42,0.96),rgba(15,23,42,1))] backdrop-blur-md shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)] px-4 py-2 md:py-2"
		>
			<!-- 桌面端：宽版极简胶囊布局 (Option B Wide) -->
			<div class="hidden md:flex flex-col items-center gap-1">
				<!-- 主胶囊 Header -->
				<div class="w-full grid grid-cols-[160px_1fr_auto] lg:grid-cols-[200px_1fr_200px] items-center gap-2 lg:gap-4 px-2 lg:px-6 py-3 rounded-[99px] bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] backdrop-blur-md shadow-[0_18px_45px_-18px_rgba(0,0,0,1)]">
					
					<!-- Logo -->
					<div class="flex items-center justify-start">
						<div class="m-0 text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)] leading-none italic cursor-default">
							{{ titleText }}
						</div>
					</div>

					<!-- Nav (Centered) -->
					<nav class="flex items-center justify-center gap-2 lg:gap-3 relative" data-header-mega-nav>
						<!-- Vertical Divider Left -->
						<div class="w-px h-6 bg-white/10 absolute left-0 hidden xl:block"></div>
						
						<button
							v-for="section in primaryMegaNavSections"
							:key="section.id"
							type="button"
							class="relative inline-flex items-center gap-1.5 rounded-full px-3 py-2 text-[15px] font-medium transition-all duration-200"
							:class="currentMegaNavId === section.id ? 'tz-text-primary' : 'tz-text-secondary hover:text-white'"
							:aria-controls="megaPanelId"
							:aria-expanded="activeMegaNavId === section.id"
							aria-haspopup="dialog"
							@click.stop="toggleMegaNav(section.id)"
							@keydown.enter.prevent="toggleMegaNav(section.id)"
							@keydown.space.prevent="toggleMegaNav(section.id)"
							@keydown.down.prevent="openMegaNav(section.id)"
						>
							<span>{{ t(section.labelKey, section.labelFallback) }}</span>
							<Icon
								name="lucide:chevron-down"
								class="h-3.5 w-3.5 transition-transform duration-200"
								:class="{ 'rotate-180 text-[#40ffaa]': activeMegaNavId === section.id }"
							/>
							<span 
								class="absolute bottom-[-4px] left-3 right-3 h-[2px] rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] transition-opacity"
								:class="currentMegaNavId === section.id ? 'opacity-100' : 'opacity-0'"
							></span>
						</button>

						<!-- Vertical Divider Right -->
						<div class="w-px h-6 bg-white/10 absolute right-0 hidden xl:block"></div>
					</nav>

					<!-- Right Actions -->
					<div class="flex items-center justify-end gap-4">
						<!-- Search -->
						<button
							class="w-9 h-9 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] flex items-center justify-center text-[#0b1020] shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:text-[#0b1020] hover:shadow-[0_8px_24px_-6px_rgba(0,0,0,1)] transition-all"
							@click="openSidebar"
							aria-label="Search"
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<circle cx="11" cy="11" r="8"></circle>
								<path d="m21 21-4.3-4.3"></path>
							</svg>
						</button>

						<!-- Language -->
						<div class="relative" data-lang-wrapper>
							<button
								class="w-9 h-9 rounded-full bg-slate-900/70 flex items-center justify-center text-[11px] font-bold tz-text-secondary shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:text-sky-300 hover:shadow-[0_8px_24px_-6px_rgba(0,0,0,1)] transition-all"
								@click.stop="toggleDropdown"
								@keydown="onButtonKeydown"
								:id="buttonId"
								aria-haspopup="listbox"
								:aria-expanded="isOpen"
								:aria-label="'Switch language'"
							>
								<span class="text-xs font-bold uppercase">{{ currentLocale.iso?.split('-')[0] || 'EN' }}</span>
							</button>
							
							<!-- Dropdown Teleport Logic (Reused) -->
							<teleport to="body">
								<transition
									enter-active-class="transition-all duration-200 ease-in-out"
									leave-active-class="transition-all duration-200 ease-in-out"
									enter-from-class="opacity-0 -translate-y-2.5"
									leave-to-class="opacity-0 -translate-y-2.5"
								>
									<div
										v-if="isOpen"
										class="fixed inset-0 z-[1200] flex items-start justify-center pt-[80px]"
									>
										<div class="absolute inset-0 bg-black/80 backdrop-blur-sm md:hidden"></div>
										<div
											class="relative w-full md:w-[90vw] md:max-w-[1600px] bg-slate-950/80 backdrop-blur-xl border-2 border-[#6b73ff]/40 rounded-2xl overflow-auto max-h-[70vh] shadow-[0_0_30px_rgba(107,115,255,0.6)] grid grid-cols-[repeat(auto-fit,minmax(160px,1fr))] gap-1.5 justify-items-center"
											role="listbox"
											:id="dropdownId"
											:aria-labelledby="buttonId"
											tabindex="0"
											@keydown="onListKeydown"
										>
											<button
												v-for="(locale, index) in availableLocales"
												:key="locale.code"
												class="w-full py-2.5 px-3 bg-transparent border-none text-white text-sm text-center cursor-pointer transition-all duration-200 inline-flex items-center justify-center gap-2 hover:bg-[#2aa3ff40]"
												:class="{ 'bg-[#2aa3ff40] font-medium': locale.code === currentLocale.code }"
												role="option"
												:aria-selected="locale.code === currentLocale.code"
												:tabindex="-1"
												:ref="setOptionRefAt(index)"
												@click="switchLanguage(locale.code)"
											>
												<span class="w-[1.2em] inline-block" aria-hidden="true">
													<img :src="flagSrc(locale)" alt="" class="w-[1.2em] h-[1.2em] block" />
												</span>
												<span>{{ locale.name }}</span>
											</button>
										</div>
									</div>
								</transition>
							</teleport>
						</div>

						<!-- Share/Points -->
						<button
							class="w-10 h-10 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] flex items-center justify-center text-[#0b1020] shadow-[0_10px_26px_-6px_rgba(0,0,0,1)] hover:shadow-[0_14px_32px_-8px_rgba(0,0,0,1)] transition-all transform hover:scale-105"
							@click.stop="toggleShare()"
							:aria-expanded="shareOpen"
							aria-label="Open membership panel"
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 3h12l4 6-10 13L2 9Z"/></svg>
						</button>
					</div>
				</div>

				<!-- 面包屑 (移至胶囊下方，极简风格 - 无背景) -->
				<nav
					v-if="breadcrumbs.length"
					aria-label="Breadcrumb"
					class="flex justify-center mt-1"
				>
					<ol class="flex items-center gap-1.5 text-[11px] tz-text-muted leading-tight transition-colors hover:text-slate-300">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="index"
							class="flex items-center gap-1"
						>
							<NuxtLink
								v-if="crumb.to && index < breadcrumbs.length - 1"
								:to="crumb.to"
								class="tz-text-secondary hover:text-white transition-colors"
							>
								{{ crumb.label }}
							</NuxtLink>
							<span v-else class="tz-text-secondary font-medium">
								{{ crumb.label }}
							</span>
							<span v-if="index < breadcrumbs.length - 1" class="tz-text-disabled">/</span>
						</li>
					</ol>
				</nav>
			</div>

			<HeaderMegaMenu
				:section="activeMegaNavSection"
				:panel-id="megaPanelId"
				@navigate="closeMegaNav"
			/>

			<!-- 移动端：新版极简双行布局 -->
			<div class="md:hidden flex flex-col gap-3">
				
				<!-- 第一行：Logo (左) + 工具图标 (右) -->
				<div class="flex items-center justify-between px-1">
					<!-- Logo -->
					<div class="m-0 text-2xl phone-390:text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)] leading-none">
						{{ titleText }}
					</div>

					<!-- 右侧工具图标组 -->
					<div class="flex items-center gap-3 phone-390:gap-4">
						<!-- Search (Icon) -->
						<button
							class="w-9 h-9 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] flex items-center justify-center text-[#0b1020] shadow-[0_4px_14px_rgba(0,0,0,0.9)]"
							@click="openSidebar"
							aria-label="Search"
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<circle cx="11" cy="11" r="8"></circle>
								<path d="m21 21-4.3-4.3"></path>
							</svg>
						</button>

						<!-- Guides (Icon) -->
						<NuxtLink
							:to="localePath('/guides/tireguides')"
							class="tz-text-secondary hover:text-white transition-colors p-1"
							aria-label="Guides"
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
						</NuxtLink>

						<!-- Language Switcher (Text + Icon) -->
						<div class="relative" data-lang-wrapper>
							<button
								class="tz-text-secondary hover:text-white transition-colors flex items-center gap-1 p-1"
								@click.stop="toggleDropdown"
								@keydown="onButtonKeydown"
								:id="buttonId"
								aria-haspopup="listbox"
								:aria-expanded="isOpen"
								:aria-label="'Switch language'"
							>
								<span class="text-xs font-bold uppercase">{{ currentLocale.iso?.split('-')[0] || currentLocale.code }}</span>
								<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="{ 'rotate-180': isOpen }" class="transition-transform"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
							</button>
						</div>

						<!-- Share/Points (Icon) -->
						<button
							class="tz-text-secondary hover:text-[#40ffaa] transition-colors p-1"
							@click.stop="toggleShare()"
							:aria-expanded="shareOpen"
							aria-label="Open membership panel"
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 3h12l4 6-10 13L2 9Z"/></svg>
						</button>
					</div>
				</div>

				<!-- 第二行：主要导航 (Segmented Control Style) -->
				<nav class="bg-white/5 rounded-xl p-1 flex items-center justify-between relative" aria-label="Mobile primary navigation">
					<NuxtLink
						:to="localePath('/shop')"
						class="flex-1 py-2 rounded-lg text-sm phone-390:text-[15px] font-semibold text-center transition-all"
						:class="currentMegaNavId === 'products' ? 'bg-white/10 tz-text-primary shadow-sm border border-white/10' : 'tz-text-secondary hover:text-white'"
					>
						{{ $t('footer.menus.products', 'Products') }}
					</NuxtLink>
					
					<NuxtLink
						:to="localePath('/support/faqs')"
						class="flex-1 py-2 rounded-lg text-sm phone-390:text-[15px] font-semibold text-center transition-all"
						:class="currentMegaNavId === 'support' ? 'bg-white/10 tz-text-primary shadow-sm border border-white/10' : 'tz-text-secondary hover:text-white'"
					>
						{{ $t('footer.menus.support', 'Support') }}
					</NuxtLink>
					
					<NuxtLink
						:to="localePath('/company/about')"
						class="flex-1 py-2 rounded-lg text-sm phone-390:text-[15px] font-semibold text-center transition-all"
						:class="currentMegaNavId === 'company' ? 'bg-white/10 tz-text-primary shadow-sm border border-white/10' : 'tz-text-secondary hover:text-white'"
					>
						{{ $t('footer.menus.company', 'Company') }}
					</NuxtLink>
				</nav>

				<!-- 第三行：面包屑 (恢复移动端显示) -->
				<nav
					v-if="breadcrumbs.length"
					aria-label="Breadcrumb"
					class="px-2 pb-1 -mt-1"
				>
					<ol class="flex items-center gap-1.5 flex-wrap justify-center text-[10px] tz-text-muted leading-tight">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="index"
							class="flex items-center gap-1"
						>
							<NuxtLink
								v-if="crumb.to && index < breadcrumbs.length - 1"
								:to="crumb.to"
								class="tz-text-secondary hover:text-white transition-colors truncate max-w-[100px]"
							>
								{{ crumb.label }}
							</NuxtLink>
							<span v-else class="tz-text-secondary font-medium truncate max-w-[120px]">
								{{ crumb.label }}
							</span>
							<span v-if="index < breadcrumbs.length - 1" class="tz-text-disabled">/</span>
						</li>
					</ol>
				</nav>

			</div>
		</div>

		<!-- LeverAndPoint 弹窗 -->
		<teleport to="body">
			<transition
				enter-active-class="transition-opacity duration-300 ease-out"
				leave-active-class="transition-opacity duration-200 ease-in"
				enter-from-class="opacity-0"
				leave-to-class="opacity-0"
			>
				<div
					v-if="shareOpen"
					class="fixed inset-0 z-[9999] flex items-center justify-center p-0 md:p-4 pointer-events-none"
				>
					<!-- 不透明背景遮罩 -->
					<div
						class="absolute inset-0 bg-black/80 backdrop-blur-sm pointer-events-auto"
						@click="shareOpen = false"
					></div>
					<!-- 弹窗内容：自下而上的 slide-up 动画，与其它弹窗保持一致 -->
					<Transition name="slide-up" appear>
						<div
							class="relative w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] flex pointer-events-auto leverandpoint-modal-shell"
							aria-modal="true"
							role="dialog"
							aria-label="Membership"
						>
							<LeverAndPoint @close="shareOpen = false" />
						</div>
					</Transition>
				</div>
			</transition>
		</teleport>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, unref, watch, type ComponentPublicInstance } from 'vue'
import { useThrottleFn } from '@vueuse/core'
import { useI18n, useLocalePath, useRoute, useRouter, useState } from '#imports'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'
import HeaderMegaMenu from '~/components/HeaderMegaMenu.vue'
import LeverAndPoint from '~/components/LeverAndPoint.vue'
import {
  findPrimaryMegaNavSectionByPath,
  normalizePrimaryMegaNavPath,
  primaryMegaNavSections,
  primaryMegaNavPathMatches,
  type PrimaryMegaNavCard,
  type PrimaryMegaNavId,
  type PrimaryMegaNavSection,
} from '~/utils/primaryMegaNav'

// Site Title
const props = defineProps<{ title?: string }>()
const { siteTitle } = useSiteTitle()
const titleText = computed(() => {
  const fromProp = (props.title ?? '').toString().trim()
  return fromProp.length ? fromProp : siteTitle.value
})

const headerRootRef = ref<HTMLElement | null>(null)
let headerResizeObserver: ResizeObserver | null = null

const megaPanelId = 'header-primary-mega-menu'
const activeMegaNavId = ref<PrimaryMegaNavId | null>(null)

const closeMegaNav = () => {
  activeMegaNavId.value = null
}

const activeMegaNavSection = computed<PrimaryMegaNavSection | null>(() => {
  if (!activeMegaNavId.value) return null
  return primaryMegaNavSections.find(section => section.id === activeMegaNavId.value) || null
})

const openMegaNav = (id: PrimaryMegaNavId) => {
  activeMegaNavId.value = id
  isOpen.value = false

  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'header-mega-nav' } }))
  }
}

const toggleMegaNav = (id: PrimaryMegaNavId) => {
  if (activeMegaNavId.value === id) {
    closeMegaNav()
    return
  }

  openMegaNav(id)
}

const updateHeaderOffset = () => {
  if (typeof window === 'undefined') return
  const el = headerRootRef.value
  if (!el) return

  const rect = el.getBoundingClientRect()
  const offset = Math.max(0, Math.ceil(rect.bottom))
  document.documentElement.style.setProperty('--site-header-offset', `${offset}px`)
}

const throttledUpdateHeaderOffset = useThrottleFn(updateHeaderOffset, 150)

// Share button (Membership panel)
const shareOpen = ref(false)

const toggleShare = () => {
  closeMegaNav()
  isOpen.value = false
  shareOpen.value = !shareOpen.value
  if (shareOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'header-share' } }))
  }
}

// Open Sidebar (Search)
const openSidebar = () => {
  closeMegaNav()
  isOpen.value = false
  openShopSearch()
}

const { open: openShopSearch } = useShopSearchSheet()

// Language Switcher
const { locale, locales, setLocale, t } = useI18n() as any
const localePath = useLocalePath()
const router = useRouter()
const route = useRoute()

const getLocaleCodes = () => {
  return (unref(locales) || [])
    .map((item: any) => (typeof item === 'string' ? item : item?.code))
    .filter(Boolean)
}

const normalizeNavPath = (path: string) => normalizePrimaryMegaNavPath(path, getLocaleCodes())

const currentMegaNavId = computed<PrimaryMegaNavId | null>(() => {
  const section = findPrimaryMegaNavSectionByPath(route.path || '/', primaryMegaNavSections, getLocaleCodes())

  return section?.id || null
})

const alternateLinksOverride = useState<{ code: string; path: string }[] | null>(
  'alternateLinksOverride',
  () => null
)

interface BreadcrumbItem {
  label: string
  to?: string
}

const routePathFromTo = (to: string) => {
  return to.split('#')[0]?.split('?')[0] || '/'
}

const isSameOrNestedPath = (currentPath: string, targetPath: string) => {
  return primaryMegaNavPathMatches(currentPath, targetPath, getLocaleCodes())
}

const cardDisplayLabel = (card: PrimaryMegaNavCard) => {
  return card.title || (t(card.labelKey, card.labelFallback) as string)
}

const findCurrentMegaCard = (): { section: PrimaryMegaNavSection; card: PrimaryMegaNavCard } | null => {
  const currentPath = normalizeNavPath(route.path || '/')
  const currentSection = currentMegaNavId.value
    ? primaryMegaNavSections.find(section => section.id === currentMegaNavId.value)
    : null

  if (!currentSection) return null

  for (const card of currentSection.cards) {
    const targetPath = normalizeNavPath(routePathFromTo(card.to))
    if (isSameOrNestedPath(currentPath, targetPath)) {
      return { section: currentSection, card }
    }
  }

  return null
}

const breadcrumbs = computed<BreadcrumbItem[]>(() => {
  const items: BreadcrumbItem[] = []
  const homeTo = localePath('/')

  // Home
  items.push({ label: t('breadcrumbs.home', 'Home') as string, to: homeTo })

  const currentPath = route.path || ''

  if (currentPath === homeTo) {
    return items
  }

  // Blog hub: Home / Wheelsbuild blog
  const blogHub = localePath('/blog')
  if (currentPath === blogHub) {
    items.push({ label: t('breadcrumbs.blog', 'Blog') as string })
    return items
  }

  // Blog 子页面：Home / Wheelsbuild blog / {具体页面}
  if (currentPath.startsWith(blogHub + '/')) {
    items.push({
      label: t('breadcrumbs.blog', 'Blog') as string,
      to: blogHub,
    })

    if (currentPath === localePath('/blog/news')) {
      items.push({ label: t('blog.nav.news', 'News') as string })
    } else if (currentPath === localePath('/blog/wheelsbuild')) {
      items.push({ label: t('blog.nav.wheelsbuild', 'Wheelbuild') as string })
    } else {
      const segments = currentPath.split('/').filter(Boolean)
      const last = segments[segments.length - 1] || ''
      items.push({ label: last })
    }

    return items
  }

  // Guides category: Home / Guides / {具体页面}
  const tireGuidesPath = localePath('/guides/tireguides')
  const wheelsetGuidePath = localePath('/guides/wheelset-buyers')
  const guidesPrefix = tireGuidesPath.replace(/\/tireguides\/?$/, '')
  if (currentPath.startsWith(`${guidesPrefix}/`)) {
    items.push({ label: t('breadcrumbs.guides', 'Guides') as string })

    // 根据具体路径映射更友好的标题

    if (currentPath === tireGuidesPath) {
      items.push({ label: t('products.nav.tireSizeCharts', 'Tire Guides') as string })
    } else if (currentPath === wheelsetGuidePath) {
      items.push({ label: t('products.nav.wheelsetBuyersGuide', 'Wheelset Buyers Guide') as string })
    } else {
      // 其它 /guides/* 页面，使用最后一段路径作为标题占位
      const segments = currentPath.split('/').filter(Boolean)
      const last = segments[segments.length - 1] || ''
      items.push({ label: last })
    }

    return items
  }

  // Privacy Policy 页面
  const privacyPath = localePath('/privacy')
  if (currentPath === privacyPath) {
    items.push({ label: 'Privacy Policy' })
    return items
  }

  // Cookie Policy 页面
  const cookiePolicyPath = localePath('/cookie-policy')
  if (currentPath === cookiePolicyPath) {
    items.push({ label: 'Cookie Policy' })
    return items
  }

  // Terms of Service 页面
  const termsPath = localePath('/terms')
  if (currentPath === termsPath) {
    items.push({ label: 'Terms of Service' })
    return items
  }

  // Policies 页面：Home / Policies (/ + 子页)
  const policiesHub = localePath('/policies')
  if (currentPath === policiesHub) {
    items.push({ label: 'Policies' })
    return items
  }

  if (currentPath.startsWith(policiesHub + '/')) {
    items.push({ label: 'Policies', to: policiesHub })

    const segments = currentPath.split('/').filter(Boolean)
    const last = segments[segments.length - 1] || ''
    const policiesLabels: Record<string, string> = {
      privacy: 'Privacy Policy',
      cookie: 'Cookie Policy',
      'refund-return': 'Refund & Return',
      terms: 'Terms of Service',
    }

    items.push({ label: policiesLabels[last] || last })
    return items
  }

  // Header mega menu categories are the single source of truth for section breadcrumbs.
  const megaMatch = findCurrentMegaCard()
  if (megaMatch) {
    items.push({ label: t(megaMatch.section.labelKey, megaMatch.section.labelFallback) as string })
    items.push({ label: cardDisplayLabel(megaMatch.card) })
    return items
  }

  return items
})

const switchLocalePath = (targetLocale: string) => {
	const currentFullPath = router.currentRoute.value?.fullPath || '/'
	// 宽松断言交给 vue-i18n 处理具体的 locale 类型，避免 TS 联合类型报错
	return localePath({ path: currentFullPath }, targetLocale as any)
}

const isOpen = ref(false)

type LocaleOption = { code: string; name?: string; iso?: string }

const normalizedLocales = computed<LocaleOption[]>(() => {
  const list = locales.value
  if (Array.isArray(list)) {
    return list.map((entry: any) => ({
      code: entry.code,
      name: entry.name,
      iso: entry.iso,
    }))
  }
  return []
})

type LocaleCode = typeof locale.value
const isLocaleCode = (value: string): value is LocaleCode => {
  return normalizedLocales.value.some((item: LocaleOption) => item.code === value)
}

const currentLocale = computed<LocaleOption>(() => {
  return (
    normalizedLocales.value.find((l: LocaleOption) => l.code === locale.value) ||
    normalizedLocales.value[0] ||
    { code: locale.value }
  )
})

const availableLocales = computed<LocaleOption[]>(() => {
  return normalizedLocales.value.filter((l: LocaleOption) => l.code !== locale.value)
})

const buttonId = 'lang-switcher-button'
const dropdownId = 'lang-switcher-dropdown'

const optionRefs = ref<HTMLElement[]>([])
const setOptionRef = (el: Element | ComponentPublicInstance | null, index: number) => {
  const target = (el && '$el' in el)
    ? ((el as ComponentPublicInstance).$el as HTMLElement | null)
    : (el as HTMLElement | null)
  if (!target) return
  optionRefs.value[index] = target
}

const setOptionRefAt = (index: number) => {
  return (el: Element | ComponentPublicInstance | null) => setOptionRef(el, index)
}

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    closeMegaNav()
  }
  if (isOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'language' } }))
  }
}

const onButtonKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    isOpen.value = !isOpen.value
    if (isOpen.value) {
      closeMegaNav()
    }
    if (isOpen.value) {
      nextTick(() => optionRefs.value[0]?.focus())
    }
  } else if (e.key === 'Escape') {
    isOpen.value = false
    closeMegaNav()
  }
}

const onListKeydown = (e: KeyboardEvent) => {
  const refs = optionRefs.value
  if (!Array.isArray(refs) || !refs.length) return
  const idx = refs.findIndex(el => el === document.activeElement)
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    const nextIndex = idx >= 0 ? (idx + 1) % refs.length : 0
    refs[nextIndex]?.focus()
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    const prevIndex = idx >= 0 ? (idx - 1 + refs.length) % refs.length : refs.length - 1
    refs[prevIndex]?.focus()
  } else if (e.key === 'Escape') {
    isOpen.value = false
    closeMegaNav()
    document.getElementById(buttonId)?.focus()
  }
}

const switchLanguage = async (code: string) => {
  try {
    if (!code || !isLocaleCode(code) || code === locale.value) { isOpen.value = false; return }

    const overrideTargetPath = alternateLinksOverride.value?.find((entry: { code: string; path: string }) => entry.code === code)?.path
    const currentFullPath = router.currentRoute.value?.fullPath || ''
    const fallbackTargetPath = switchLocalePath(code as any)
    const targetPath = overrideTargetPath || fallbackTargetPath

    locale.value = code
    await nextTick()
    try { await setLocale(code) } catch {}
    if (targetPath && targetPath !== currentFullPath) {
      try {
        await router.push(targetPath)
      } catch {
        window.location.assign(targetPath)
      }
    }
  } finally {
    isOpen.value = false
    closeMegaNav()
  }
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target
  if (!(target instanceof Element)) return
  if (!target.closest('[data-lang-wrapper]') && !target.closest('#' + dropdownId)) {
    isOpen.value = false
  }
  if (!target.closest('.site-header-root')) {
    closeMegaNav()
  }
}

const handleHeaderKeydown = (event: KeyboardEvent) => {
  if (event.key !== 'Escape') return
  isOpen.value = false
  closeMegaNav()
}

watch(
  () => route.fullPath,
  () => {
    closeMegaNav()
    nextTick(updateHeaderOffset)
  },
)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleHeaderKeydown)

  nextTick(() => {
    updateHeaderOffset()
    window.addEventListener('resize', throttledUpdateHeaderOffset)
    if ('ResizeObserver' in window) {
      headerResizeObserver = new ResizeObserver(() => throttledUpdateHeaderOffset())
      if (headerRootRef.value) {
        headerResizeObserver.observe(headerRootRef.value)
      }
    }
  })

  const onGlobalPopup = (event: Event) => {
    try {
      const custom = event as CustomEvent<{ id?: string }>
      const id = custom?.detail?.id
      if (id !== 'language') {
        isOpen.value = false
      }
      if (id !== 'header-mega-nav') {
        closeMegaNav()
      }
    } catch {}
  }
  window.addEventListener('ui:popup-open', onGlobalPopup as EventListener)
  onBeforeUnmount(() => {
    window.removeEventListener('ui:popup-open', onGlobalPopup as EventListener)
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleHeaderKeydown)

  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', throttledUpdateHeaderOffset)
  }

  if (headerResizeObserver) {
    headerResizeObserver.disconnect()
    headerResizeObserver = null
  }
})

const flagFilenameFromISO = (entry: LocaleOption | null | undefined) => {
  try {
    const iso = (entry && entry.iso) ? String(entry.iso) : ''
    const cc = (iso.split('-')[1] || '').toUpperCase()
    if (cc.length !== 2) return null
    const codepoints = [...cc]
      .map(c => 0x1F1E6 + (c.charCodeAt(0) - 65))
      .map(cp => cp.toString(16))
      .join('-')
    return `${codepoints}.svg`
  } catch {
    return null
  }
}

const flagSrc = (entry: LocaleOption | null | undefined) => {
  const file = flagFilenameFromISO(entry)
  if (!file) return ''
  return `/twemoji/svg/${file}`
}
</script>

<style scoped>
.header-mobile-nav-text {
  font-size: 12px !important;
}

@media (max-width: 767px) {
	.site-header-root {
		max-height: 150px;
	}
}

/* iPad / small tablets: prevent desktop language switcher from overflowing header pill */
@media (min-width: 768px) and (max-width: 1100px) {
	.site-header-root {
		width: 90vw;
	}

	.desktop-header-grid {
		grid-template-columns: 230px minmax(0, 1fr) max-content;
	}

	.desktop-lang-switcher {
		width: 106px;
		padding-inline: 0.5rem;
	}
}

/* tablet-768: 768x1024 等宽度段，限制 SiteHeader 高度为 130px */
@media (min-width: 768px) and (max-width: 819px) {
	.site-header-root {
		max-height: 130px;
	}
}

/* LeverAndPoint 弹窗使用的自下而上滑入动画（与 CartDrawer/QuickBuy/Wishlist 保持一致） */
.slide-up-enter-active,
.slide-up-leave-active {
	transition: transform 0.3s ease-out, opacity 0.3s ease-out;
}

.slide-up-enter-from,
.slide-up-leave-to {
	transform: translateY(100%);
	opacity: 0;
}

.slide-up-enter-to,
.slide-up-leave-from {
	transform: translateY(0%);
	opacity: 1;
}

@media (max-width: 767px) {
	.leverandpoint-modal-shell {
		height: min(95vh, calc(100vh - 16px));
		max-height: min(95vh, calc(100vh - 16px));
	}

	@supports (height: 100svh) {
		.leverandpoint-modal-shell {
			height: min(95svh, calc(100svh - 16px));
			max-height: min(95svh, calc(100svh - 16px));
		}
	}

	@supports (height: 100dvh) {
		.leverandpoint-modal-shell {
			height: min(95dvh, calc(100dvh - 16px));
			max-height: min(95dvh, calc(100dvh - 16px));
		}
	}
}

/* tablet-820: 820x1180 等宽度段，限制 SiteHeader 高度为 130px */
@media (min-width: 820px) and (max-width: 1023px) {
	.site-header-root {
		max-height: 130px;
	}
}
</style>
