<template>
	<div ref="headerRootRef" class="fixed top-0 md:top-1.5 left-0 md:left-1/2 md:-translate-x-1/2 w-full md:w-[95vw] md:max-w-[1200px] z-[900] site-header-root">
		<div
			class="site-header-surface relative w-full rounded-none md:rounded-[30px] bg-[radial-gradient(circle_at_top_left,rgba(15,23,42,0.96),rgba(15,23,42,1))] backdrop-blur-md shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)] px-4 py-2 md:py-2"
		>
			<!-- 桌面端：宽版极简胶囊布局 (Option B Wide) -->
			<div class="hidden md:flex flex-col items-center gap-1">
				<!-- 主胶囊 Header -->
				<div class="w-full grid grid-cols-[160px_1fr_auto] lg:grid-cols-[200px_1fr_200px] items-center gap-2 lg:gap-4 px-2 lg:px-6 py-3 rounded-[99px] bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] backdrop-blur-md shadow-[0_18px_45px_-18px_rgba(0,0,0,1)]">
					
					<!-- Logo -->
					<div class="flex items-center justify-start">
						<div class="m-0 text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#ffffff] via-[#e5e7eb] to-[#94a3b8] tracking-wide drop-shadow-[0_2px_8px_rgba(226,232,240,0.22)] leading-none cursor-default">
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
							class="relative inline-flex items-center gap-1.5 rounded-full px-3 py-2 text-[16px] font-medium text-white transition-all duration-200"
							:class="currentMegaNavId === section.id ? 'font-semibold' : 'font-medium'"
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
							class="site-header-top-icon-button--plain w-9 h-9 rounded-full bg-transparent flex items-center justify-center tz-text-secondary hover:text-white transition-colors"
							@click="openSidebar"
							aria-label="Search"
						>
							<svg xmlns="http://www.w3.org/2000/svg" class="w-[22px] h-[22px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<circle cx="11" cy="11" r="8"></circle>
								<path d="m21 21-4.3-4.3"></path>
							</svg>
						</button>

						<!-- Language -->
						<div class="relative" data-lang-wrapper>
							<button
								class="site-header-top-icon-button--plain h-9 min-w-[3.5rem] rounded-full bg-transparent inline-flex items-center justify-center gap-1.5 px-2 tz-text-secondary hover:text-white transition-colors"
								@click.stop="toggleDropdown"
								@keydown="onButtonKeydown"
								:id="buttonId"
								aria-haspopup="listbox"
								:aria-expanded="isOpen"
								:aria-label="'Switch language'"
							>
								<span v-if="currentLocaleFlagSrc" class="inline-flex h-5 w-5 items-center justify-center" aria-hidden="true">
									<img :src="currentLocaleFlagSrc" alt="" class="block h-5 w-5" />
								</span>
								<Icon v-else name="lucide:languages" class="h-5 w-5" aria-hidden="true" />
								<span class="text-[13px] font-bold uppercase leading-none">{{ currentLocaleLabel }}</span>
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
										class="fixed inset-0 z-[1200] flex items-center justify-center md:items-start md:pt-[calc(var(--site-header-offset,80px)+18px)]"
									>
										<div class="absolute inset-0 bg-black/80 backdrop-blur-sm md:hidden"></div>
										<div
											class="language-dropdown-surface relative w-full md:w-[88vw] md:max-w-[1500px] backdrop-blur-xl border border-white/15 rounded-2xl overflow-auto h-[90vh] max-h-[90vh] md:h-auto md:max-h-[70vh] py-3 md:py-3.5 shadow-[0_18px_56px_rgba(255,255,255,0.10),0_28px_80px_rgba(0,0,0,0.55)] grid grid-cols-[repeat(auto-fit,minmax(160px,1fr))] gap-1.5 justify-items-center"
											role="listbox"
											:id="dropdownId"
											:aria-labelledby="buttonId"
											tabindex="0"
											@keydown="onListKeydown"
										>
											<button
												v-for="(locale, index) in availableLocales"
												:key="locale.code"
												class="w-full py-2.5 px-3 bg-transparent border-none text-white text-sm text-center cursor-pointer transition-all duration-200 inline-flex items-center justify-center gap-2 hover:bg-white/10"
												:class="{ 'bg-white/10 font-medium': locale.code === currentLocale.code }"
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
							class="site-header-top-icon-button--plain w-10 h-10 rounded-full bg-transparent flex items-center justify-center tz-text-secondary hover:text-[#40ffaa] transition-colors"
							@click.stop="toggleShare()"
							:aria-expanded="shareOpen"
							aria-label="Open membership panel"
						>
							<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 3h12l4 6-10 13L2 9Z"/></svg>
						</button>
					</div>
				</div>

				<!-- 面包屑 (移至胶囊下方，极简风格 - 无背景) -->
				<nav
					v-if="breadcrumbs.length"
					aria-label="Breadcrumb"
					class="flex justify-center mt-1"
				>
					<ol class="flex items-center gap-1.5 text-[13px] tz-text-muted leading-tight transition-colors hover:text-slate-300">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="index"
							class="relative flex items-center gap-1"
							:data-breadcrumb-subnav="crumb.subNavigation ? '' : undefined"
						>
							<template v-if="crumb.subNavigation">
								<button
									type="button"
									class="breadcrumb-subnav-trigger"
									:class="{ 'breadcrumb-subnav-trigger--open': breadcrumbSubNavOpen }"
									:aria-expanded="breadcrumbSubNavOpen"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop="toggleBreadcrumbSubNav($event)"
								>
									<span>{{ crumb.label }}</span>
									<Icon name="lucide:chevron-down" class="breadcrumb-subnav-trigger__icon" />
								</button>

								<div
									v-if="breadcrumbSubNavOpen"
									class="breadcrumb-subnav-menu"
									:style="breadcrumbSubNavMenuStyle"
									role="menu"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop
								>
									<NuxtLink
										v-for="tab in crumb.subNavigation.tabs"
										:key="tab.id"
										class="breadcrumb-subnav-link"
										:class="{ 'breadcrumb-subnav-link--active': tab.active }"
										:to="tab.to"
										role="menuitem"
										:aria-current="tab.active ? 'page' : undefined"
										@click="scheduleBreadcrumbSubNavClose"
									>
										{{ tab.label }}
									</NuxtLink>
								</div>
							</template>
							<NuxtLink
								v-else-if="crumb.to && index < breadcrumbs.length - 1"
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

			<teleport to="body" :disabled="!isMobileViewport">
				<HeaderMegaMenu
					:section="activeMegaNavSection"
					:panel-id="megaPanelId"
					@navigate="closeMegaNav"
				/>
			</teleport>

			<!-- 移动端：新版极简双行布局 -->
			<div class="md:hidden flex flex-col gap-3">
				
				<!-- 第一行：Logo (左) + 工具图标 (右) -->
				<div class="flex items-center justify-between px-1">
					<!-- Logo -->
					<div class="m-0 text-2xl phone-390:text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#ffffff] via-[#e5e7eb] to-[#94a3b8] tracking-wide drop-shadow-[0_2px_8px_rgba(226,232,240,0.22)] leading-none">
						{{ titleText }}
					</div>

					<!-- 右侧工具图标组 -->
					<div class="flex items-center gap-3 phone-390:gap-4">
						<!-- Search (Icon) -->
						<button
							class="site-header-top-icon-button--plain w-9 h-9 rounded-full bg-transparent flex items-center justify-center tz-text-secondary hover:text-white transition-colors"
							@click="openSidebar"
							aria-label="Search"
						>
							<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<circle cx="11" cy="11" r="8"></circle>
								<path d="m21 21-4.3-4.3"></path>
							</svg>
						</button>

						<!-- Language Switcher (Text + Icon) -->
						<div class="relative" data-lang-wrapper>
							<button
								class="site-header-top-icon-button--plain h-9 min-w-[3.5rem] rounded-full bg-transparent inline-flex items-center justify-center gap-1.5 px-1.5 tz-text-secondary hover:text-white transition-colors"
								@click.stop="toggleDropdown"
								@keydown="onButtonKeydown"
								:id="buttonId"
								aria-haspopup="listbox"
								:aria-expanded="isOpen"
								:aria-label="'Switch language'"
							>
								<span v-if="currentLocaleFlagSrc" class="inline-flex h-5 w-5 items-center justify-center" aria-hidden="true">
									<img :src="currentLocaleFlagSrc" alt="" class="block h-5 w-5" />
								</span>
								<Icon v-else name="lucide:languages" class="h-5 w-5" aria-hidden="true" />
								<span class="text-[13px] font-bold uppercase leading-none">{{ currentLocaleLabel }}</span>
							</button>
						</div>

						<!-- Share/Points (Icon) -->
						<button
							class="site-header-top-icon-button--plain tz-text-secondary hover:text-[#40ffaa] transition-colors p-1"
							@click.stop="toggleShare()"
							:aria-expanded="shareOpen"
							aria-label="Open membership panel"
						>
							<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 3h12l4 6-10 13L2 9Z"/></svg>
						</button>
					</div>
				</div>

				<!-- 第二行：主要导航。移动端同样打开卡片菜单，避免三级 TAB 漏在页面横条里。 -->
				<nav ref="mobilePrimaryNavRef" class="site-header-mobile-nav rounded-xl p-1 grid grid-cols-4 gap-1 relative shadow-[0_8px_18px_-14px_rgba(0,0,0,0.9)]" aria-label="Mobile primary navigation">
					<button
						v-for="section in primaryMegaNavSections"
						:key="section.id"
						type="button"
						class="min-w-0 py-2 rounded-lg text-[13px] phone-390:text-[14px] font-semibold text-center text-white transition-all inline-flex items-center justify-center gap-0.5"
						:class="currentMegaNavId === section.id ? 'bg-[#2d3238] shadow-sm' : 'bg-[#101318]'"
						:aria-controls="megaPanelId"
						:aria-expanded="activeMegaNavId === section.id"
						aria-haspopup="dialog"
						@click.stop="toggleMegaNav(section.id)"
					>
						<span class="truncate">{{ t(section.labelKey, section.labelFallback) }}</span>
						<Icon
							name="lucide:chevron-down"
							class="h-3 w-3 shrink-0 transition-transform duration-200"
							:class="{ 'rotate-180 text-[#40ffaa]': activeMegaNavId === section.id }"
						/>
					</button>
				</nav>

				<!-- 第三行：面包屑 (恢复移动端显示) -->
				<nav
					v-if="breadcrumbs.length"
					aria-label="Breadcrumb"
					class="px-2 pb-1 -mt-1"
				>
					<ol class="flex items-center gap-1.5 flex-wrap justify-center text-[13px] tz-text-muted leading-tight">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="index"
							class="relative flex items-center gap-1"
							:data-breadcrumb-subnav="crumb.subNavigation ? '' : undefined"
						>
							<template v-if="crumb.subNavigation">
								<button
									type="button"
									class="breadcrumb-subnav-trigger breadcrumb-subnav-trigger--mobile"
									:class="{ 'breadcrumb-subnav-trigger--open': breadcrumbSubNavOpen }"
									:aria-expanded="breadcrumbSubNavOpen"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop="toggleBreadcrumbSubNav($event)"
								>
									<span>{{ crumb.label }}</span>
									<Icon name="lucide:chevron-down" class="breadcrumb-subnav-trigger__icon" />
								</button>

								<div
									v-if="breadcrumbSubNavOpen"
									class="breadcrumb-subnav-menu breadcrumb-subnav-menu--mobile"
									:style="breadcrumbSubNavMenuStyle"
									role="menu"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop
								>
									<NuxtLink
										v-for="tab in crumb.subNavigation.tabs"
										:key="tab.id"
										class="breadcrumb-subnav-link"
										:class="{ 'breadcrumb-subnav-link--active': tab.active }"
										:to="tab.to"
										role="menuitem"
										:aria-current="tab.active ? 'page' : undefined"
										@click="scheduleBreadcrumbSubNavClose"
									>
										{{ tab.label }}
									</NuxtLink>
								</div>
							</template>
							<NuxtLink
								v-else-if="crumb.to && index < breadcrumbs.length - 1"
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
import {
  getPageSubNavigationForPath,
  type PageSubNavigationTab,
} from '~/utils/pageSubNavigation'

// Header brand title is controlled only by the public site settings API.
const { brandTitle } = useSiteTitle()
const titleText = computed(() => brandTitle.value)

const headerRootRef = ref<HTMLElement | null>(null)
const mobilePrimaryNavRef = ref<HTMLElement | null>(null)
const isMobileViewport = ref(false)
let headerResizeObserver: ResizeObserver | null = null

const megaPanelId = 'header-primary-mega-menu'
const activeMegaNavId = ref<PrimaryMegaNavId | null>(null)
const breadcrumbSubNavOpen = ref(false)
const breadcrumbSubNavMobileTop = ref('8.5rem')
const breadcrumbSubNavMenuStyle = computed(() => ({
  '--breadcrumb-subnav-mobile-top': breadcrumbSubNavMobileTop.value,
}))

const closeMegaNav = () => {
  activeMegaNavId.value = null
}

const activeMegaNavSection = computed<PrimaryMegaNavSection | null>(() => {
  if (!activeMegaNavId.value) return null
  return primaryMegaNavSections.find(section => section.id === activeMegaNavId.value) || null
})

const openMegaNav = (id: PrimaryMegaNavId) => {
  updateHeaderOffset()
  activeMegaNavId.value = id
  isOpen.value = false
  breadcrumbSubNavOpen.value = false
  nextTick(updateHeaderOffset)

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
  isMobileViewport.value = window.innerWidth < 768

  const el = headerRootRef.value
  if (!el) return

  const rect = el.getBoundingClientRect()
  const offset = Math.max(0, Math.ceil(rect.bottom))
  document.documentElement.style.setProperty('--site-header-offset', `${offset}px`)

  const mobileNavRect = mobilePrimaryNavRef.value?.getBoundingClientRect()
  const mobileMegaTop = Math.max(0, Math.ceil((mobileNavRect?.bottom ?? rect.bottom) + 2))
  document.documentElement.style.setProperty('--header-mega-mobile-top', `${mobileMegaTop}px`)
}

const throttledUpdateHeaderOffset = useThrottleFn(updateHeaderOffset, 150)

// Share button (Membership panel)
const shareOpen = ref(false)

const toggleShare = () => {
  closeMegaNav()
  isOpen.value = false
  breadcrumbSubNavOpen.value = false
  shareOpen.value = !shareOpen.value
  if (shareOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'header-share' } }))
  }
}

// Open Sidebar (Search)
const openSidebar = () => {
  closeMegaNav()
  isOpen.value = false
  breadcrumbSubNavOpen.value = false
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
  subNavigation?: BreadcrumbSubNavigation
}

interface BreadcrumbSubNavigationItem {
  id: string
  label: string
  to: string
  active: boolean
}

interface BreadcrumbSubNavigation {
  ariaLabel: string
  tabs: BreadcrumbSubNavigationItem[]
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

const localizedNavTarget = (to: string) => {
  if (/^https?:\/\//i.test(to)) return to

  const hashIndex = to.indexOf('#')
  const withoutHash = hashIndex >= 0 ? to.slice(0, hashIndex) : to
  const hash = hashIndex >= 0 ? to.slice(hashIndex) : ''

  const queryIndex = withoutHash.indexOf('?')
  const path = queryIndex >= 0 ? withoutHash.slice(0, queryIndex) : withoutHash
  const query = queryIndex >= 0 ? withoutHash.slice(queryIndex) : ''

  return `${localePath(path || '/')}${query}${hash}`
}

const getSubNavigationLabel = (tab: PageSubNavigationTab) => {
  if (tab.labelKey) return t(tab.labelKey, tab.fallback || tab.label || tab.id) as string
  return tab.label || tab.fallback || tab.id
}

const currentPageSubNavigation = computed(() => {
  return getPageSubNavigationForPath(route.path || '/', getLocaleCodes())
})

const normalizedRouteHash = computed(() => {
  const raw = String(route.hash || '').replace(/^#/, '')
  if (!raw) return ''

  try {
    return decodeURIComponent(raw)
  } catch {
    return raw
  }
})

const activePageSubNavigationTab = computed<PageSubNavigationTab | null>(() => {
  const entry = currentPageSubNavigation.value
  if (!entry || !normalizedRouteHash.value) return null

  return entry.tabs.find((tab) => tab.id === normalizedRouteHash.value) || null
})

const breadcrumbSubNavigationItems = computed<BreadcrumbSubNavigationItem[]>(() => {
  const entry = currentPageSubNavigation.value
  if (!entry) return []

  const activeId = activePageSubNavigationTab.value?.id || ''

  return entry.tabs.map((tab: PageSubNavigationTab) => {
    const target = typeof tab.to === 'string' && tab.to ? tab.to : `${entry.path}#${String(tab.id)}`

    return {
      id: String(tab.id),
      label: getSubNavigationLabel(tab),
      to: localizedNavTarget(target),
      active: tab.id === activeId,
    }
  })
})

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

const baseBreadcrumbs = computed<BreadcrumbItem[]>(() => {
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

const breadcrumbs = computed<BreadcrumbItem[]>(() => {
  const items = baseBreadcrumbs.value.map((item) => ({ ...item }))
  const entry = currentPageSubNavigation.value
  const tabs = breadcrumbSubNavigationItems.value

  if (!entry || !tabs.length || items.length < 2) return items

  const lastIndex = items.length - 1
  const pageCrumb = items[lastIndex] as BreadcrumbItem
  const activeTab = activePageSubNavigationTab.value
  const pageLabel = pageCrumb.label || (t('breadcrumbs.pageSections', 'Page sections') as string)
  const subNavigation: BreadcrumbSubNavigation = {
    ariaLabel: `${pageLabel} sections`,
    tabs,
  }

  if (activeTab) {
    items[lastIndex] = {
      ...pageCrumb,
      to: pageCrumb.to || localizedNavTarget(entry.path),
    }
    items.push({
      label: getSubNavigationLabel(activeTab),
      subNavigation,
    })
    return items
  }

  items[lastIndex] = {
    ...pageCrumb,
    subNavigation,
  }

  return items
})

const closeBreadcrumbSubNav = () => {
  breadcrumbSubNavOpen.value = false
}

const scheduleBreadcrumbSubNavClose = () => {
  if (typeof window === 'undefined') return
  window.setTimeout(closeBreadcrumbSubNav, 0)
}

const updateBreadcrumbSubNavPosition = (target: EventTarget | null | undefined) => {
  if (typeof window === 'undefined' || !(target instanceof HTMLElement)) return
  const rect = target.getBoundingClientRect()
  const safeGap = 8
  const nextTop = Math.max(safeGap, Math.min(window.innerHeight - safeGap, rect.bottom + safeGap))
  breadcrumbSubNavMobileTop.value = `${Math.round(nextTop)}px`
}

const toggleBreadcrumbSubNav = (event?: MouseEvent) => {
  const nextOpen = !breadcrumbSubNavOpen.value
  if (nextOpen) {
    updateBreadcrumbSubNavPosition(event?.currentTarget)
  }
  breadcrumbSubNavOpen.value = nextOpen
  if (breadcrumbSubNavOpen.value) {
    closeMegaNav()
    isOpen.value = false
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'breadcrumb-subnav' } }))
    }
  }
}

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

const currentLocaleLabel = computed(() => {
  const entry = currentLocale.value
  const raw = entry.iso?.split('-')[0] || entry.code || ''
  return raw.replace('_', '-').split('-')[0].slice(0, 2).toUpperCase()
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
    breadcrumbSubNavOpen.value = false
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
      breadcrumbSubNavOpen.value = false
    }
    if (isOpen.value) {
      nextTick(() => optionRefs.value[0]?.focus())
    }
  } else if (e.key === 'Escape') {
    isOpen.value = false
    closeMegaNav()
    closeBreadcrumbSubNav()
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
    closeBreadcrumbSubNav()
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
    closeBreadcrumbSubNav()
  }
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target
  if (!(target instanceof Element)) return
  if (!target.closest('[data-lang-wrapper]') && !target.closest('#' + dropdownId)) {
    isOpen.value = false
  }
  if (!target.closest('[data-breadcrumb-subnav]')) {
    closeBreadcrumbSubNav()
  }
  if (!target.closest('.site-header-root')) {
    closeMegaNav()
  }
}

const handleHeaderKeydown = (event: KeyboardEvent) => {
  if (event.key !== 'Escape') return
  isOpen.value = false
  closeMegaNav()
  closeBreadcrumbSubNav()
}

watch(
  () => route.fullPath,
  () => {
    closeMegaNav()
    closeBreadcrumbSubNav()
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
      if (id !== 'breadcrumb-subnav') {
        closeBreadcrumbSubNav()
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

const currentLocaleFlagSrc = computed(() => flagSrc(currentLocale.value))
</script>

<style scoped>
.header-mobile-nav-text {
  font-size: 12px !important;
}

.site-header-root .site-header-top-icon-button--plain {
	border: 0 !important;
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
}

.site-header-root .site-header-top-icon-button--plain:hover,
.site-header-root .site-header-top-icon-button--plain:focus-visible,
.site-header-root .site-header-top-icon-button--plain[aria-expanded='true'] {
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
}

.language-dropdown-surface {
	background: linear-gradient(
		135deg,
		rgba(148, 163, 184, 0.12) 0%,
		rgba(15, 17, 21, 0.96) 38%,
		rgba(0, 0, 0, 0.98) 100%
	) !important;
}

.breadcrumb-subnav-trigger {
	display: inline-flex;
	max-width: min(46vw, 220px);
	align-items: center;
	gap: 0.2rem;
	border: 0;
	border-radius: 9999px;
	background: transparent;
	padding: 0;
	color: var(--tz-text-secondary);
	font: inherit;
	font-weight: 650;
	line-height: inherit;
	cursor: pointer;
	transition: color 0.18s ease;
}

.breadcrumb-subnav-trigger:hover,
.breadcrumb-subnav-trigger:focus-visible,
.breadcrumb-subnav-trigger--open {
	color: #ffffff;
}

.breadcrumb-subnav-trigger:focus-visible {
	outline: 1px solid rgba(64, 255, 170, 0.72);
	outline-offset: 0.18rem;
}

.breadcrumb-subnav-trigger span {
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.breadcrumb-subnav-trigger__icon {
	width: 0.72rem;
	height: 0.72rem;
	flex: 0 0 auto;
	color: #40ffaa;
	transition: transform 0.18s ease;
}

.breadcrumb-subnav-trigger--open .breadcrumb-subnav-trigger__icon {
	transform: rotate(180deg);
}

.breadcrumb-subnav-menu {
	position: absolute;
	top: calc(100% + 0.45rem);
	left: 50%;
	z-index: 180;
	display: grid;
	width: min(88vw, 22rem);
	min-width: min(78vw, 16rem);
	max-width: calc(100vw - 2rem);
	max-height: min(58vh, 390px);
	transform: translateX(-50%);
	gap: 0.25rem;
	overflow: auto;
	border: 1px solid rgba(148, 163, 184, 0.2);
	border-radius: 0.95rem;
	background:
		radial-gradient(circle at top left, rgba(64, 255, 170, 0.12), transparent 42%),
		linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(2, 6, 23, 0.98));
	padding: 0.42rem;
	box-shadow:
		0 24px 54px -22px rgba(0, 0, 0, 1),
		inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.breadcrumb-subnav-link {
	display: flex;
	align-items: center;
	justify-content: space-between;
	border-radius: 0.72rem;
	padding: 0.58rem 0.72rem;
	color: rgba(226, 232, 240, 0.86);
	font-size: 0.78rem;
	font-weight: 700;
	line-height: 1.15;
	text-decoration: none;
	overflow-wrap: anywhere;
	transition:
		background-color 0.18s ease,
		color 0.18s ease,
		transform 0.18s ease;
}

.breadcrumb-subnav-link:hover,
.breadcrumb-subnav-link:focus-visible {
	background: rgba(255, 255, 255, 0.08);
	color: #ffffff;
	transform: translateY(-1px);
}

.breadcrumb-subnav-link:focus-visible {
	outline: 1px solid rgba(64, 255, 170, 0.72);
	outline-offset: 0.12rem;
}

.breadcrumb-subnav-link--active {
	background: #ffffff;
	color: #020617;
}

.breadcrumb-subnav-link--active:hover,
.breadcrumb-subnav-link--active:focus-visible {
	background: #ffffff;
	color: #020617;
}

.breadcrumb-subnav-trigger--mobile {
	max-width: 36vw;
}

.breadcrumb-subnav-menu--mobile {
	position: fixed;
	top: var(--breadcrumb-subnav-mobile-top, 8.5rem);
	right: max(0.75rem, env(safe-area-inset-right));
	left: max(0.75rem, env(safe-area-inset-left));
	width: auto;
	min-width: 0;
	max-width: none;
	max-height: min(50vh, 360px);
	transform: none;
}

@media (max-width: 767px) {
	.site-header-root {
		max-height: 150px;
	}

	.site-header-surface {
		background: linear-gradient(180deg, #17191c 0%, #101216 52%, #0d111b 100%) !important;
		border-bottom: 0;
	}

	.site-header-mobile-nav {
		background: linear-gradient(180deg, #171a1f 0%, #101318 100%) !important;
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

