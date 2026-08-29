<template>
	<div
		ref="headerRootRef"
		class="fixed top-0 left-0 w-full z-[900] site-header-root"
	>
		<div
			class="site-header-surface relative w-full rounded-none px-4 py-2 md:px-0 md:py-0"
		>
			<!-- 桌面端：全宽单层横向导航（1280px+） -->
			<div class="hidden xl:flex flex-col items-stretch">
				<div class="site-header-mainbar desktop-header-grid w-full grid grid-cols-[220px_1fr_220px] xl:grid-cols-[280px_1fr_280px] items-center gap-4 px-4 lg:px-8 py-0 min-h-[64px]">

					<!-- Logo -->
					<div class="flex items-center justify-start">
						<NuxtLink v-if="hasSiteLogo" :to="localePath('/')" class="site-header-brand site-header-brand--desktop site-header-brand--image" :aria-label="brandHomeLabel">
							<!-- Original asset dimensions are metadata; the header uses a fixed display box. -->
							<StorefrontImage
								v-if="hasSiteLogo"
								:src="siteLogo"
								alt=""
								class="site-header-brand__image"
								preset="logo"
								loading="eager"
								decoding="async"
								@error="handleSiteLogoError"
							/>
						</NuxtLink>
					</div>

					<!-- Nav (Centered) -->
					<nav class="flex items-center justify-center gap-5 lg:gap-8 xl:gap-10 relative" data-header-mega-nav>
						<!-- Vertical Divider Left -->
						<div class="hidden"></div>

						<button
							v-for="section in primaryMegaNavSections"
							:key="section.id"
							type="button"
							class="site-header-menu-laser"
							:class="{ 'site-header-menu-laser--active': currentMegaNavId === section.id }"
							:aria-controls="megaPanelId"
							:aria-expanded="activeMegaNavId === section.id"
							aria-haspopup="dialog"
							@click.stop="toggleMegaNav(section.id)"
							@keydown.enter.prevent="toggleMegaNav(section.id)"
							@keydown.space.prevent="toggleMegaNav(section.id)"
							@keydown.down.prevent="openMegaNav(section.id)"
						>
							<span class="site-header-menu-laser__text">{{ t(section.labelKey, section.labelFallback) }}</span>
							<Icon
								name="lucide:chevron-down"
								class="site-header-menu-laser__icon"
								:class="{ 'rotate-180 text-[#059669]': activeMegaNavId === section.id }"
							/>
						</button>

						<!-- Vertical Divider Right -->
						<div class="hidden"></div>
					</nav>

					<!-- Right Actions -->
					<div class="site-header-actions flex items-center justify-end">
						<!-- Search -->
						<div class="site-header-action-cell site-header-action-cell--search">
							<button
								ref="desktopContentNavigationTriggerRef"
								class="site-header-action-button site-header-search-trigger"
								@click="toggleContentNavigationTransition"
								:aria-label="contentNavigationTriggerLabel"
							>
								<Icon name="lucide:search" class="site-header-search-trigger__icon" />
							</button>
							<div
								v-if="!contentNavigationTransitionOpen"
								class="site-header-search-hint"
								role="tooltip"
							>
								<span class="site-header-search-hint__title">{{ contentNavigationTriggerLabel }}</span>
								<span class="site-header-search-hint__body">{{ contentNavigationTriggerBody }}</span>
							</div>
						</div>

						<!-- Language -->
						<div class="site-header-action-cell site-header-language-wrapper relative" data-lang-wrapper>
							<button
								class="site-header-action-button site-header-language-trigger tz-text-secondary transition-colors"
								@click.stop="toggleDropdown"
								@keydown="onButtonKeydown"
								:id="buttonId"
								aria-haspopup="listbox"
								:aria-expanded="isOpen"
								:aria-label="'Switch language'"
							>
								<span v-if="currentLocaleFlagSrc" class="inline-flex h-5 w-5 items-center justify-center" aria-hidden="true">
										<img :src="currentLocaleFlagSrc" alt="" width="20" height="20" class="block h-5 w-5" />
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
										class="language-dropdown-layer fixed z-[1200] flex items-stretch justify-center md:inset-0 md:items-start md:pt-[calc(var(--site-header-overlay-offset,80px)+18px)]"
									>
										<div
											class="language-dropdown-surface tz-mobile-dialog-surface relative w-full md:w-[88vw] md:max-w-[1500px] backdrop-blur-xl border tz-border-subtle rounded-2xl overflow-auto h-[90vh] max-h-[90vh] md:h-auto md:max-h-[70vh] py-3 md:py-3.5 shadow-md grid grid-cols-[repeat(auto-fit,minmax(160px,1fr))] gap-1.5 justify-items-center"
											role="listbox"
											:id="dropdownId"
											:aria-labelledby="buttonId"
											tabindex="0"
											@keydown="onListKeydown"
										>
											<button
												v-for="(locale, index) in availableLocales"
												:key="locale.code"
												class="w-full py-2.5 px-3 bg-transparent border-none tz-text-primary text-sm text-center cursor-pointer transition-all duration-200 inline-flex items-center justify-center gap-2 hover:tz-surface-subtle"
												:class="{ 'tz-surface-subtle font-medium': locale.code === currentLocale.code }"
												role="option"
												:aria-selected="locale.code === currentLocale.code"
												:tabindex="-1"
												:ref="setOptionRefAt(index)"
												@click="switchLanguage(locale.code)"
											>
												<span class="w-[1.2em] inline-block" aria-hidden="true">
																	<img :src="flagSrc(locale)" alt="" width="20" height="20" class="w-[1.2em] h-[1.2em] block" />
												</span>
												<span :lang="locale.iso || locale.code.replace('_', '-')">{{ locale.name }}</span>
											</button>
										</div>
									</div>
								</transition>
							</teleport>
						</div>

					</div>

					<teleport to="body" :disabled="!isMobileViewport">
						<LazyHeaderMegaMenu
							v-if="activeMegaNavSection"
							:section="activeMegaNavSection"
							:panel-id="megaPanelId"
							@navigate="handleMegaNavNavigate"
						/>
					</teleport>
				</div>

				<!-- 面包屑：点击当前层级箭头弹出该层级的同级路由 -->
				<nav
					v-if="breadcrumbs.length && !isStorefrontCatalogDetailBreadcrumbRoute"
					aria-label="Breadcrumb"
					class="site-header-breadcrumb-row"
				>
					<ol class="site-header-breadcrumb-list flex items-center gap-1.5 text-sm tz-text-secondary leading-tight transition-colors">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="crumb.id"
							class="site-header-breadcrumb-item relative flex items-center gap-1"
							:data-breadcrumb-subnav="crumb.subNavigation ? crumb.id : undefined"
						>
							<template v-if="crumb.subNavigation">
								<button
									type="button"
									class="breadcrumb-subnav-trigger"
									:class="{ 'breadcrumb-subnav-trigger--open': activeBreadcrumbSubNavId === crumb.id }"
									:aria-expanded="activeBreadcrumbSubNavId === crumb.id"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop="toggleBreadcrumbSubNav(crumb.id, $event)"
								>
									<span>{{ crumb.label }}</span>
									<Icon name="lucide:chevron-down" class="breadcrumb-subnav-trigger__icon" />
									<span
										v-if="crumb.id === lastExpandableBreadcrumbId"
										class="breadcrumb-subnav-pulse-dot"
										aria-hidden="true"
									></span>
								</button>

								<div
									v-if="activeBreadcrumbSubNavId === crumb.id"
									class="breadcrumb-subnav-menu"
									:style="breadcrumbSubNavMenuStyle"
									role="menu"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop
								>
									<a
										v-for="tab in crumb.subNavigation.tabs"
										:key="tab.id"
										class="breadcrumb-subnav-link"
										:class="{ 'breadcrumb-subnav-link--active': tab.active }"
										:href="tab.to"
										role="menuitem"
										:aria-current="tab.active ? 'page' : undefined"
										@click.prevent="navigateBreadcrumbSubNav(tab.to)"
									>
										{{ tab.label }}
									</a>
								</div>
							</template>
							<NuxtLink
								v-else-if="crumb.id === 'home'"
								:to="crumb.to || localePath('/')"
								class="tz-text-secondary transition-colors inline-flex items-center justify-center"
								:aria-label="crumb.label"
								:title="crumb.label"
							>
								<Icon name="lucide:house" class="h-4 w-4 text-[var(--tz-site-accent)]" aria-hidden="true" />
							</NuxtLink>
							<NuxtLink
								v-else-if="crumb.to && index < breadcrumbs.length - 1"
								:to="crumb.to"
								class="tz-text-secondary transition-colors"
							>
								{{ crumb.label }}
							</NuxtLink>
							<span v-else class="tz-text-secondary font-medium">
								{{ crumb.label }}
							</span>
							<span v-if="index < breadcrumbs.length - 1" class="site-header-breadcrumb-separator tz-text-muted">/</span>
						</li>
					</ol>
				</nav>
			</div>

			<!-- 移动端和平板：品牌工具栏、主导航、面包屑三行独立布局 -->
			<div class="xl:hidden flex flex-col gap-0">
				<div class="site-header-mobile-surface -mx-4 -mt-2 flex flex-col px-4 pt-2 pb-0">

				<!-- 第一行：左侧 Shop 图标、居中 Logo、右侧工具图标区 -->
				<div ref="mobileTopbarRef" class="site-header-mobile-topbar grid grid-cols-3 items-center px-1">
					<div class="site-header-mobile-shop-slot flex items-center justify-start">
						<button
							v-if="mobileShopSection"
							type="button"
							class="site-header-mobile-shop-trigger"
							:class="{ 'site-header-mobile-shop-trigger--active': currentMegaNavId === mobileShopSection.id }"
							:aria-controls="megaPanelId"
							:aria-expanded="activeMegaNavId === mobileShopSection.id"
							aria-haspopup="dialog"
							:aria-label="mobileShopLabel"
							:title="mobileShopLabel"
							@click.prevent.stop="toggleMegaNav(mobileShopSection.id)"
						>
							<Icon name="lucide:shopping-bag" class="h-5 w-5" aria-hidden="true" />
						</button>
					</div>

					<div class="site-header-mobile-brand-slot flex min-w-0 items-center justify-center">
						<NuxtLink v-if="hasSiteLogo" :to="localePath('/')" class="site-header-brand site-header-brand--mobile site-header-brand--image" :aria-label="brandHomeLabel">
							<StorefrontImage
								v-if="hasSiteLogo"
								:src="siteLogo"
								alt=""
								width="48"
								height="48"
								class="site-header-brand__image"
								preset="logo"
								sizes="48px"
								densities="1x 2x"
								loading="eager"
								decoding="async"
								@error="handleSiteLogoError"
							/>
						</NuxtLink>
					</div>

					<!-- 工具图标固定在右侧三分之一区域内右对齐。 -->
					<div class="site-header-actions site-header-actions--mobile">
						<!-- Search (Icon) -->
						<div class="site-header-action-cell site-header-action-cell--search">
							<button
								ref="mobileContentNavigationTriggerRef"
								class="site-header-action-button site-header-search-trigger site-header-search-trigger--mobile"
								@click="toggleContentNavigationTransition"
								:aria-label="contentNavigationTriggerLabel"
							>
								<Icon name="lucide:search" class="site-header-search-trigger__icon" />
							</button>
						</div>

						<!-- Language Switcher (Text + Icon) -->
						<div class="site-header-action-cell site-header-language-wrapper relative" data-lang-wrapper>
							<button
								class="site-header-action-button site-header-language-trigger tz-text-secondary transition-colors"
								@click.stop="toggleDropdown"
								@keydown="onButtonKeydown"
								:id="buttonId"
								aria-haspopup="listbox"
								:aria-expanded="isOpen"
								:aria-label="'Switch language'"
							>
								<span v-if="currentLocaleFlagSrc" class="inline-flex h-5 w-5 items-center justify-center" aria-hidden="true">
										<img :src="currentLocaleFlagSrc" alt="" width="20" height="20" class="block h-5 w-5" />
								</span>
								<Icon v-else name="lucide:languages" class="h-5 w-5" aria-hidden="true" />
								<span class="site-header-language-label--mobile text-[13px] font-bold uppercase leading-none">{{ currentLocaleLabel }}</span>
							</button>
						</div>

					</div>
				</div>

				</div>

				<!-- 第二行：其余四个大类独立于品牌工具栏，保持固定可控的行高。 -->
				<div class="site-header-mobile-nav-row -mx-4 px-0">
					<nav ref="mobilePrimaryNavRef" class="site-header-mobile-nav w-full rounded-none px-0 py-0 grid grid-cols-4 gap-0 relative shadow-md" aria-label="Mobile primary navigation">
					<button
						v-for="section in mobileSecondaryNavSections"
						:key="section.id"
						type="button"
						class="site-header-mobile-nav__button min-w-0 w-full py-2 px-0.5 rounded-none text-[13px] font-semibold uppercase leading-none text-center tz-text-primary transition-all inline-flex items-center justify-center gap-1"
						:class="{ 'site-header-mobile-nav__button--active': currentMegaNavId === section.id }"
						:aria-controls="megaPanelId"
						:aria-expanded="activeMegaNavId === section.id"
						aria-haspopup="dialog"
						@click.stop="toggleMegaNav(section.id)"
					>
						<span class="min-w-0 truncate">{{ t(section.labelKey, section.labelFallback) }}</span>
						<Icon
							name="lucide:chevron-down"
							class="h-3.5 w-3.5 shrink-0 transition-transform duration-200"
							:class="{ 'rotate-180 text-[#059669]': activeMegaNavId === section.id }"
						/>
					</button>
					</nav>
				</div>

				<!-- 第三行：面包屑 -->
				<nav
					v-if="breadcrumbs.length && !isStorefrontCatalogDetailBreadcrumbRoute"
					aria-label="Breadcrumb"
					class="site-header-breadcrumb-row site-header-breadcrumb-row--mobile"
				>
					<ol class="site-header-breadcrumb-list flex items-center gap-1.5 text-sm tz-text-secondary leading-tight">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="crumb.id"
							class="site-header-breadcrumb-item relative flex items-center gap-1"
							:data-breadcrumb-subnav="crumb.subNavigation ? crumb.id : undefined"
						>
							<template v-if="crumb.subNavigation">
								<button
									type="button"
									class="breadcrumb-subnav-trigger breadcrumb-subnav-trigger--mobile"
									:class="{ 'breadcrumb-subnav-trigger--open': activeBreadcrumbSubNavId === crumb.id }"
									:aria-expanded="activeBreadcrumbSubNavId === crumb.id"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop="toggleBreadcrumbSubNav(crumb.id, $event)"
								>
									<span>{{ crumb.label }}</span>
									<Icon name="lucide:chevron-down" class="breadcrumb-subnav-trigger__icon" />
									<span
										v-if="crumb.id === lastExpandableBreadcrumbId"
										class="breadcrumb-subnav-pulse-dot"
										aria-hidden="true"
									></span>
								</button>

								<div
									v-if="activeBreadcrumbSubNavId === crumb.id"
									class="breadcrumb-subnav-menu breadcrumb-subnav-menu--mobile"
									:style="breadcrumbSubNavMenuStyle"
									role="menu"
									:aria-label="crumb.subNavigation.ariaLabel"
									@click.stop
								>
									<a
										v-for="tab in crumb.subNavigation.tabs"
										:key="tab.id"
										class="breadcrumb-subnav-link"
										:class="{ 'breadcrumb-subnav-link--active': tab.active }"
										:href="tab.to"
										role="menuitem"
										:aria-current="tab.active ? 'page' : undefined"
										@click.prevent="navigateBreadcrumbSubNav(tab.to)"
									>
										{{ tab.label }}
									</a>
								</div>
							</template>
							<NuxtLink
								v-else-if="crumb.id === 'home'"
								:to="crumb.to || localePath('/')"
								class="tz-text-secondary transition-colors inline-flex items-center justify-center"
								:aria-label="crumb.label"
								:title="crumb.label"
							>
								<Icon name="lucide:house" class="h-4 w-4 text-[var(--tz-site-accent)]" aria-hidden="true" />
							</NuxtLink>
							<NuxtLink
								v-else-if="crumb.to && index < breadcrumbs.length - 1"
								:to="crumb.to"
								class="tz-text-secondary transition-colors truncate max-w-[100px]"
							>
								{{ crumb.label }}
							</NuxtLink>
							<span v-else class="tz-text-secondary font-medium truncate max-w-[120px]">
								{{ crumb.label }}
							</span>
							<span v-if="index < breadcrumbs.length - 1" class="site-header-breadcrumb-separator tz-text-muted">/</span>
						</li>
					</ol>
				</nav>

			</div>
		</div>

		<GlobalContentNavigationTransitionOverlay
			:open="contentNavigationTransitionOpen"
			:desktop-anchor="desktopContentNavigationTriggerRef"
			:mobile-anchor="mobileContentNavigationTriggerRef"
			@select="handleContentNavigationOption"
		/>
		<GlobalAllFaqsSearchOverlay
			v-if="globalFaqSearchMounted"
			:open="globalFaqSearchOpen"
			@close="closeGlobalFaqSearch"
		/>

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
					class="fixed inset-0 z-[9999] flex items-center justify-center p-0 md:p-4 pointer-events-none tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
				>
					<!-- 不透明背景遮罩 -->
					<div
						class="absolute inset-0 bg-slate-900/20 backdrop-blur-sm pointer-events-auto"
						@click="closeShare"
					></div>
					<!-- 弹窗内容：自下而上的 slide-up 动画，与其它弹窗保持一致 -->
					<Transition name="slide-up" appear>
						<div
							class="relative w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] flex pointer-events-auto leverandpoint-modal-shell tz-mobile-dialog-surface"
							aria-modal="true"
							role="dialog"
							aria-label="Membership"
						>
							<LazyLeverAndPoint
								v-if="shareOpen"
								@close="closeShare"
							/>
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
import { useSiteSettings } from '~/composables/usePublicSettings'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'
import GlobalContentNavigationTransitionOverlay from '~/components/GlobalContentNavigationTransitionOverlay.vue'
import GlobalAllFaqsSearchOverlay from '~/components/faq/GlobalAllFaqsSearchOverlay.vue'
import {
  findPrimaryMegaNavSectionByPath,
  normalizePrimaryMegaNavPath,
  primaryMegaNavSections,
  type PrimaryMegaNavCard,
  type PrimaryMegaNavId,
  type PrimaryMegaNavSection,
} from '~/utils/primaryMegaNav'
import {
  getPageSubNavigationTabFromPath,
  pageSubNavigationChildPath,
  pageSubNavigationEntries,
  type PageSubNavigationEntry,
  type PageSubNavigationTab,
} from '~/utils/pageSubNavigation'
import localeManifest from '~/i18n/locales.manifest'

// Header brand logo is controlled only by the public site settings API.
const { siteSettings } = useSiteSettings()
const siteLogo = computed(() => (siteSettings.value.siteLogo || '').toString().trim())
const siteLogoFailed = ref(false)
const hasSiteLogo = computed(() => Boolean(siteLogo.value && !siteLogoFailed.value))
const brandHomeLabel = computed(() => 'Home')

const handleSiteLogoError = () => {
	siteLogoFailed.value = true
}

watch(siteLogo, () => {
	siteLogoFailed.value = false
})

const headerRootRef = ref<HTMLElement | null>(null)
const mobileTopbarRef = ref<HTMLElement | null>(null)
const mobilePrimaryNavRef = ref<HTMLElement | null>(null)
const desktopContentNavigationTriggerRef = ref<HTMLElement | null>(null)
const mobileContentNavigationTriggerRef = ref<HTMLElement | null>(null)
const isMobileViewport = ref(false)
let headerResizeObserver: ResizeObserver | null = null
let headerMetricsFrame: number | null = null
const appliedHeaderMetrics = new Map<string, string>()

const megaPanelId = 'header-primary-mega-menu'
const activeMegaNavId = ref<PrimaryMegaNavId | null>(null)
const mobileShopSection = primaryMegaNavSections.find((section) => section.id === 'products') || null
const mobileSecondaryNavSections = primaryMegaNavSections.filter((section) => section.id !== 'products')
const activeBreadcrumbSubNavId = ref<string | null>(null)
const breadcrumbSubNavMobileTop = ref('8.5rem')
const contentNavigationTransitionOpen = ref(false)
const globalFaqSearchMounted = ref(false)
const globalFaqSearchOpen = ref(false)
const overlayBackStack = useOverlayBackStack()
const breadcrumbSubNavMenuStyle = computed(() => ({
  '--breadcrumb-subnav-mobile-top': breadcrumbSubNavMobileTop.value,
}))

const closeMegaNavState = () => {
  activeMegaNavId.value = null
}

const closeMegaNav = () => {
  void overlayBackStack.close('header-mega-nav')
  closeMegaNavState()
}

const handleMegaNavNavigate = () => {
  // Clear the menu history state without navigating back before NuxtLink pushes its target.
  void overlayBackStack.close('header-mega-nav', 'navigate')
}

const activeMegaNavSection = computed<PrimaryMegaNavSection | null>(() => {
  if (!activeMegaNavId.value) return null
  return primaryMegaNavSections.find(section => section.id === activeMegaNavId.value) || null
})

const mobileShopLabel = computed(() => (
  mobileShopSection
    ? t(mobileShopSection.labelKey, mobileShopSection.labelFallback) as string
    : 'Shop'
))

const openMegaNav = (id: PrimaryMegaNavId) => {
  scheduleHeaderOffsetUpdate()
  activeMegaNavId.value = id
  isOpen.value = false
  nextTick(scheduleHeaderOffsetUpdate)
  overlayBackStack.open('header-mega-nav', closeMegaNavState)
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

  // Read all geometry before changing root styles, keeping this work to one frame.
  const rect = el.getBoundingClientRect()
  const offset = Math.max(0, Math.ceil(rect.bottom))
  const mobileTopbarRect = mobileTopbarRef.value?.getBoundingClientRect()
  const mobileLanguageTop = Math.max(0, Math.ceil(mobileTopbarRect?.bottom ?? rect.bottom))
  const mobileNavRect = mobilePrimaryNavRef.value?.getBoundingClientRect()
  const mobileMegaTop = Math.max(0, Math.ceil((mobileNavRect?.bottom ?? rect.bottom) + 2))

  isMobileViewport.value = window.innerWidth < 1280

  const metrics = [
    // Overlay geometry may be measured after mount; document-flow spacers must not use it.
    ['--site-header-overlay-offset', `${offset}px`],
    ['--site-header-mobile-topbar-bottom', `${mobileLanguageTop}px`],
    ['--header-mega-mobile-top', `${mobileMegaTop}px`],
  ] as const

  for (const [name, value] of metrics) {
    if (appliedHeaderMetrics.get(name) === value) continue
    document.documentElement.style.setProperty(name, value)
    appliedHeaderMetrics.set(name, value)
  }
}

const scheduleHeaderOffsetUpdate = () => {
  if (typeof window === 'undefined' || headerMetricsFrame !== null) return

  headerMetricsFrame = window.requestAnimationFrame(() => {
    headerMetricsFrame = null
    updateHeaderOffset()
  })
}

const throttledUpdateHeaderOffset = useThrottleFn(scheduleHeaderOffsetUpdate, 150)

// Share button (Membership panel)
const shareOpen = ref(false)

const closeShareState = () => {
  shareOpen.value = false
}

const closeShare = () => {
  void overlayBackStack.close('header-share')
  closeShareState()
}

const toggleShare = () => {
	isOpen.value = false
  if (shareOpen.value) {
    closeShare()
    return
  }

  shareOpen.value = true
  overlayBackStack.open('header-share', closeShareState)
}

type ContentNavigationOptionId = 'products' | 'faq' | 'pages' | 'blog'

const closeContentNavigationTransitionState = () => {
	contentNavigationTransitionOpen.value = false
}

const closeContentNavigationTransition = (
	reason: 'user' | 'navigate' | 'replace' = 'user',
) => {
	void overlayBackStack.close('global-content-navigation-transition', reason)
	closeContentNavigationTransitionState()
}

const openContentNavigationTransition = () => {
	isOpen.value = false
	contentNavigationTransitionOpen.value = true
	overlayBackStack.open(
		'global-content-navigation-transition',
		closeContentNavigationTransitionState,
	)
}

const toggleContentNavigationTransition = () => {
	if (contentNavigationTransitionOpen.value) {
		closeContentNavigationTransition()
		return
	}

	openContentNavigationTransition()
}

const closeGlobalFaqSearchState = () => {
	globalFaqSearchOpen.value = false
}

const closeGlobalFaqSearch = (
	reason: 'user' | 'navigate' | 'replace' = 'user',
) => {
	void overlayBackStack.close('global-all-faqs-search', reason)
	closeGlobalFaqSearchState()
}

const openGlobalFaqSearch = () => {
	globalFaqSearchMounted.value = true
	globalFaqSearchOpen.value = true
	overlayBackStack.open(
		'global-all-faqs-search',
		closeGlobalFaqSearchState,
	)
}

// Language Switcher
const { locale, locales, setLocale, t } = useI18n() as any
const localePath = useLocalePath()
const router = useRouter()
const route = useRoute()

const contentNavigationTriggerLabel = computed(() => (
	t(
		'header.globalNavigationTransition.trigger',
		'Open content navigation',
	)
))
const contentNavigationTriggerBody = computed(() => (
	t(
		'header.globalNavigationTransition.triggerBody',
		'Choose Products, FAQ, Pages, or Blog.',
	)
))

const handleContentNavigationOption = (option: ContentNavigationOptionId) => {
	if (option === 'faq') {
		closeContentNavigationTransition('navigate')
		openGlobalFaqSearch()
		return
	}

	if (option !== 'products') return

	closeContentNavigationTransition('navigate')

	if (isMobileViewport.value) {
		if (typeof window !== 'undefined') {
			window.dispatchEvent(new CustomEvent('ui:product-category-sidebar-open'))
		}
		return
	}

	void router.push(localePath('/shop'))
}

const getLocaleCodes = () => {
  return (unref(locales) || [])
    .map((item: any) => (typeof item === 'string' ? item : item?.code))
    .filter(Boolean)
}

const getAllLocaleCodes = () => {
  return Array.from(new Set([
    ...localeManifest.map(entry => entry.code),
    ...getLocaleCodes(),
  ]))
}

const normalizeNavPath = (path: string) => normalizePrimaryMegaNavPath(path, getAllLocaleCodes())

const currentMegaNavId = computed<PrimaryMegaNavId | null>(() => {
  const section = findPrimaryMegaNavSectionByPath(route.path || '/', primaryMegaNavSections, getAllLocaleCodes())

  return section?.id || null
})

const alternateLinksOverride = useState<{ code: string; path: string }[] | null>(
  'alternateLinksOverride',
  () => null
)

interface BreadcrumbItem {
  id: string
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
  return to.split('?')[0] || '/'
}

const cardDisplayLabel = (card: PrimaryMegaNavCard) => {
  return card.title || card.labelFallback
}

const localizedNavTarget = (to: string) => {
  if (/^https?:\/\//i.test(to)) return to

  const queryIndex = to.indexOf('?')
  const path = queryIndex >= 0 ? to.slice(0, queryIndex) : to
  const query = queryIndex >= 0 ? to.slice(queryIndex) : ''

  return `${localePath(path || '/')}${query}`
}

const createBreadcrumbSubNavigation = (
  ariaLabel: string,
  tabs: Array<{ id: string; label: string; to: string; active?: boolean }>
): BreadcrumbSubNavigation | undefined => {
  if (tabs.length < 2) return undefined

  return {
    ariaLabel,
    tabs: tabs.map(tab => ({
      id: tab.id,
      label: tab.label,
      to: localizedNavTarget(tab.to),
      active: Boolean(tab.active),
    })),
  }
}

type BreadcrumbRouteRecord = ReturnType<typeof router.getRoutes>[number]
type BreadcrumbLabelDefinition = { labelKey?: string; fallback: string }

interface BreadcrumbRouteCandidate {
  path: string
  segments: string[]
  depth: number
  dynamicSegmentCount: number
  order: number
  routeRecord: BreadcrumbRouteRecord
}

interface BreadcrumbRouteFamily {
  id: string
  label: string
  to: string
  rootPath: string
  depth: number
  order: number
}

interface BreadcrumbRouteLevelGroup {
  id: string
  segment: string
  path: string
  target: string
  depth: number
  order: number
}

const breadcrumbLabelDefinitions: Record<string, BreadcrumbLabelDefinition> = {
  '/resources': { labelKey: 'footer.menus.resources', fallback: 'Resources' },
  '/resources/blog': { labelKey: 'breadcrumbs.blog', fallback: 'Blog' },
  '/resources/blog/news': { labelKey: 'blog.nav.news', fallback: 'News' },
  '/resources/blog/wheelsbuild': { labelKey: 'blog.nav.wheelsbuild', fallback: 'Wheelbuild' },
  '/company': { labelKey: 'footer.menus.company', fallback: 'Company' },
  '/guides': { labelKey: 'breadcrumbs.guides', fallback: 'Guides' },
  '/resources/membershipandpoints': { labelKey: 'company.nav.membershipPoints', fallback: 'Membership & Points' },
  '/resources/picture-warehouse': { labelKey: 'company.nav.pictureWarehouse', fallback: 'Picture Warehouse' },
  '/policies': { labelKey: 'footer.menus.policies', fallback: 'Policies' },
  '/policies/cookie': { fallback: 'Cookie Policy' },
  '/policies/privacy': { fallback: 'Privacy Policy' },
  '/policies/refund-return': { fallback: 'Refund & Return Policy' },
  '/policies/terms': { fallback: 'Terms of Service' },
  '/shop': { labelKey: 'products.nav.shop', fallback: 'Shop' },
  '/resources/spoke-calculator': { labelKey: 'support.nav.spokeCalculator', fallback: 'Spoke Calculator' },
  '/support': { labelKey: 'footer.menus.support', fallback: 'Support' },
}

const excludedBreadcrumbRootSegments = new Set(['checkout', 'faq'])
const excludedBreadcrumbExactPaths = new Set([
  '/policies',
  '/sitemap.xml',
])
const technicalBreadcrumbRootSegments = new Set([
  '__nuxt_error',
  '__sitemap__',
  '__site-config__',
  '_internal',
  '_nuxt',
])

const localeCodeSet = () => new Set(getAllLocaleCodes().map(code => String(code).toLowerCase()))

const isBreadcrumbLocaleSegment = (segment: string) => {
  return localeCodeSet().has(segment.toLowerCase())
}

const normalizeBreadcrumbPath = (path: string) => {
  const normalized = normalizeNavPath(path)
  const segments = normalized.split('/').filter(Boolean)

  while (segments[0] && isBreadcrumbLocaleSegment(segments[0])) {
    segments.shift()
  }

  return segments.length ? `/${segments.join('/')}` : '/'
}

const getBreadcrumbPathSegments = (path: string) => {
  return normalizeBreadcrumbPath(path).split('/').filter(Boolean)
}

const isDynamicBreadcrumbSegment = (segment: string) => {
  return (
    segment.startsWith(':') ||
    segment.startsWith('*') ||
    segment.includes('[') ||
    /^:.+\(.+\)$/.test(segment)
  )
}

const isLocaleBreadcrumbRoute = (path: string) => {
  const firstSegment = path.split('/').filter(Boolean)[0] || ''
  return firstSegment ? isBreadcrumbLocaleSegment(firstSegment) : false
}

const isBreadcrumbExcludedPath = (path: string) => {
  const normalizedPath = normalizeBreadcrumbPath(path)
  const segments = normalizedPath.split('/').filter(Boolean)
  const rootSegment = segments[0] || ''

  return (
    !rootSegment ||
    excludedBreadcrumbExactPaths.has(normalizedPath) ||
    excludedBreadcrumbRootSegments.has(rootSegment) ||
    technicalBreadcrumbRootSegments.has(rootSegment) ||
    rootSegment.startsWith('_') ||
    rootSegment.endsWith('.xml')
  )
}

const sameBreadcrumbSegments = (left: string[], right: string[]) => {
  return left.length === right.length && left.every((segment, index) => segment === right[index])
}

const fallbackBreadcrumbRouteFamilyLabel = (segment: string) => {
  let decodedSegment = segment
  try {
    decodedSegment = decodeURIComponent(segment)
  } catch {}

  return decodedSegment
    .replace(/[_-]+/g, ' ')
    .split(' ')
    .filter(Boolean)
    .map(part => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

const resolveBreadcrumbLabelDefinition = (definition: BreadcrumbLabelDefinition) => {
  return definition.fallback
}

const getBreadcrumbMetaLabel = (meta: Record<string, unknown> | undefined) => {
  if (!meta) return ''

  const labelKey =
    meta.breadcrumbLabelKey ||
    meta.navLabelKey ||
    meta.footerLabelKey
  const labelFallback =
    meta.breadcrumbLabelFallback ||
    meta.navLabelFallback ||
    meta.footerLabelFallback

  if (typeof labelKey === 'string') {
    return typeof labelFallback === 'string' ? labelFallback.trim() : ''
  }

  const rawLabel =
    meta.breadcrumb ||
    meta.breadcrumbLabel ||
    meta.navLabel ||
    meta.label ||
    meta.title

  if (typeof rawLabel === 'string') return rawLabel.trim()
  if (rawLabel && typeof rawLabel === 'object') {
    const labels = rawLabel as Record<string, unknown>
    const english = labels.en || labels['en-US'] || labels.default
    if (typeof english === 'string') return english.trim()
  }

  return ''
}

const getBreadcrumbRouteCandidates = (includeDynamic = false): BreadcrumbRouteCandidate[] => {
  const candidates: BreadcrumbRouteCandidate[] = []
  const seen = new Set<string>()

  router.getRoutes().forEach((routeRecord, order) => {
    const rawPath = routeRecord.path || '/'
    const path = normalizeBreadcrumbPath(rawPath)
    const segments = path.split('/').filter(Boolean)
    const dynamic = segments.some(segment => isDynamicBreadcrumbSegment(segment))

    if (
      isLocaleBreadcrumbRoute(rawPath) ||
      !routeRecord.name ||
      routeRecord.redirect ||
      isBreadcrumbExcludedPath(path) ||
      (!includeDynamic && dynamic)
    ) {
      return
    }

    const key = `${path}:${dynamic ? 'dynamic' : 'static'}`
    if (seen.has(key)) return
    seen.add(key)

    candidates.push({
      path,
      segments,
      depth: segments.length,
      dynamicSegmentCount: segments.filter(segment => isDynamicBreadcrumbSegment(segment)).length,
      order,
      routeRecord,
    })
  })

  return candidates
}

const staticBreadcrumbRouteCandidates = () => getBreadcrumbRouteCandidates(false)

const getPrimaryMegaNavCardPath = (card: PrimaryMegaNavCard) => {
  return normalizeBreadcrumbPath(routePathFromTo(card.to))
}

const primaryMegaNavCardsInDisplayOrder = () => {
  return primaryMegaNavSections.flatMap(section => section.cards)
}

const primaryRootSortOrder = () => {
  const order = new Map<string, number>()
  let index = 0

  for (const card of primaryMegaNavCardsInDisplayOrder()) {
    const rootSegment = getBreadcrumbPathSegments(card.to)[0] || ''
    if (rootSegment && !order.has(rootSegment)) order.set(rootSegment, index++)
  }

  for (const section of primaryMegaNavSections) {
    for (const prefix of section.routePrefixes) {
      const rootSegment = getBreadcrumbPathSegments(prefix)[0] || ''
      if (rootSegment && !order.has(rootSegment)) order.set(rootSegment, index++)
    }
  }

  return order
}

const getPrimaryCardOrderForPath = (path: string) => {
  const normalizedPath = normalizeBreadcrumbPath(path)
  const index = primaryMegaNavCardsInDisplayOrder()
    .findIndex(card => getPrimaryMegaNavCardPath(card) === normalizedPath)

  return index >= 0 ? index : Number.MAX_SAFE_INTEGER
}

const getPrimaryMegaNavCardForPath = (path: string) => {
  const normalizedPath = normalizeBreadcrumbPath(path)

  return primaryMegaNavCardsInDisplayOrder()
    .find(card => getPrimaryMegaNavCardPath(card) === normalizedPath) || null
}

const routePatternMatchesBreadcrumbPath = (patternPath: string, targetPath: string) => {
  const patternSegments = getBreadcrumbPathSegments(patternPath)
  const targetSegments = getBreadcrumbPathSegments(targetPath)

  if (patternSegments.length !== targetSegments.length) return false

  return patternSegments.every((segment, index) => (
    isDynamicBreadcrumbSegment(segment) || segment === targetSegments[index]
  ))
}

const findBreadcrumbRouteCandidateForPath = (
  path: string,
  includeDynamic = true
) => {
  const normalizedPath = normalizeBreadcrumbPath(path)

  return getBreadcrumbRouteCandidates(includeDynamic)
    .filter(candidate => routePatternMatchesBreadcrumbPath(candidate.path, normalizedPath))
    .sort((left, right) => {
      if (left.path === normalizedPath && right.path !== normalizedPath) return -1
      if (right.path === normalizedPath && left.path !== normalizedPath) return 1
      if (left.dynamicSegmentCount !== right.dynamicSegmentCount) {
        return left.dynamicSegmentCount - right.dynamicSegmentCount
      }
      if (left.depth !== right.depth) return right.depth - left.depth
      return left.order - right.order
    })[0] || null
}

const getStaticBreadcrumbRouteCandidateForPath = (path: string) => {
  const normalizedPath = normalizeBreadcrumbPath(path)
  return staticBreadcrumbRouteCandidates().find(candidate => candidate.path === normalizedPath) || null
}

const getBreadcrumbKnownLabel = (path: string) => {
  const definition = breadcrumbLabelDefinitions[normalizeBreadcrumbPath(path)]
  return definition ? resolveBreadcrumbLabelDefinition(definition) : ''
}

const getBreadcrumbRouteFamilyLabel = (segment: string) => {
  const path = `/${segment}`
  const routeCandidate = getStaticBreadcrumbRouteCandidateForPath(path)
  const routeMetaLabel = getBreadcrumbMetaLabel(
    routeCandidate?.routeRecord.meta as Record<string, unknown> | undefined
  )

  return (
    getBreadcrumbKnownLabel(path) ||
    routeMetaLabel ||
    fallbackBreadcrumbRouteFamilyLabel(segment)
  )
}

const getBreadcrumbRouteLabel = (path: string, segment: string) => {
  const normalizedPath = normalizeBreadcrumbPath(path)
  const matchingPageTab = getBreadcrumbPageSubNavigationTab(normalizedPath)
  if (matchingPageTab) return pageSubNavigationTabLabel(matchingPageTab.tab)

  const matchingCard = getPrimaryMegaNavCardForPath(normalizedPath)
  if (matchingCard) return cardDisplayLabel(matchingCard)

  const routeCandidate = findBreadcrumbRouteCandidateForPath(normalizedPath)
  const routeMetaLabel = getBreadcrumbMetaLabel(routeCandidate?.routeRecord.meta as Record<string, unknown> | undefined)
  if (routeMetaLabel) return routeMetaLabel

  if (normalizeBreadcrumbPath(route.path || '/') === normalizedPath) {
    const currentRouteMetaLabel = getBreadcrumbMetaLabel(route.meta as Record<string, unknown> | undefined)
    if (currentRouteMetaLabel) return currentRouteMetaLabel
  }

  return getBreadcrumbKnownLabel(normalizedPath) || fallbackBreadcrumbRouteFamilyLabel(segment)
}

const sortBreadcrumbRouteCandidates = (
  left: BreadcrumbRouteCandidate,
  right: BreadcrumbRouteCandidate
) => {
  const leftCardOrder = getPrimaryCardOrderForPath(left.path)
  const rightCardOrder = getPrimaryCardOrderForPath(right.path)

  if (leftCardOrder !== rightCardOrder) return leftCardOrder - rightCardOrder
  if (left.depth !== right.depth) return left.depth - right.depth
  if (left.order !== right.order) return left.order - right.order
  return left.path.localeCompare(right.path)
}

const getPreferredBreadcrumbLevelCandidate = (
  groupPath: string,
  candidates: BreadcrumbRouteCandidate[]
) => {
  const exactCandidate = candidates.find(candidate => candidate.path === groupPath)
  if (exactCandidate && !excludedBreadcrumbExactPaths.has(groupPath)) return exactCandidate

  const primaryCardCandidate = primaryMegaNavCardsInDisplayOrder()
    .map(card => getPrimaryMegaNavCardPath(card))
    .map(cardPath => candidates.find(candidate => candidate.path === cardPath))
    .find((candidate): candidate is BreadcrumbRouteCandidate => Boolean(candidate))

  if (primaryCardCandidate) return primaryCardCandidate

  return [...candidates].sort(sortBreadcrumbRouteCandidates)[0] || null
}

const getBreadcrumbRouteLevelGroups = (
  parentSegments: string[],
  depth: number
): BreadcrumbRouteLevelGroup[] => {
  const groups = new Map<string, { segment: string; path: string; candidates: BreadcrumbRouteCandidate[] }>()

  for (const candidate of staticBreadcrumbRouteCandidates()) {
    if (
      candidate.depth < depth ||
      !sameBreadcrumbSegments(candidate.segments.slice(0, parentSegments.length), parentSegments)
    ) {
      continue
    }

    const segment = candidate.segments[depth - 1] || ''
    if (!segment) continue

    const path = `/${candidate.segments.slice(0, depth).join('/')}`
    const group = groups.get(path) || { segment, path, candidates: [] }
    group.candidates.push(candidate)
    groups.set(path, group)
  }

  return Array.from(groups.values())
    .map((group) => {
      const preferred = getPreferredBreadcrumbLevelCandidate(group.path, group.candidates)
      if (!preferred) return null

      return {
        id: group.path,
        segment: group.segment,
        path: group.path,
        target: preferred.path,
        depth,
        order: preferred.order,
      }
    })
    .filter((group): group is BreadcrumbRouteLevelGroup => Boolean(group))
    .sort((left, right) => {
      const leftCardOrder = getPrimaryCardOrderForPath(left.target)
      const rightCardOrder = getPrimaryCardOrderForPath(right.target)

      if (leftCardOrder !== rightCardOrder) return leftCardOrder - rightCardOrder
      if (left.order !== right.order) return left.order - right.order
      return left.path.localeCompare(right.path)
    })
}

const getBreadcrumbRouteFamilies = (): BreadcrumbRouteFamily[] => {
  const rootOrder = primaryRootSortOrder()

  return getBreadcrumbRouteLevelGroups([], 1)
    .map(group => ({
      id: group.segment,
      label: getBreadcrumbRouteFamilyLabel(group.segment),
      to: group.target,
      rootPath: group.path,
      depth: group.depth,
      order: rootOrder.get(group.segment) ?? Number.MAX_SAFE_INTEGER,
    }))
    .sort((left, right) => {
      if (left.order !== right.order) return left.order - right.order
      if (left.depth !== right.depth) return left.depth - right.depth
      return left.label.localeCompare(right.label)
    })
}

const isSameOrNestedBreadcrumbPath = (currentPath: string, targetPath: string) => {
  const current = normalizeBreadcrumbPath(currentPath)
  const target = normalizeBreadcrumbPath(targetPath)

  return current === target || (current.startsWith(target) && current[target.length] === '/')
}

const getBreadcrumbFamilyTarget = (rootSegment: string) => {
  return getBreadcrumbRouteFamilies().find(family => family.id === rootSegment)?.to || ''
}

const getBreadcrumbTarget = (path: string) => {
  const normalizedPath = normalizeBreadcrumbPath(path)
  const segments = getBreadcrumbPathSegments(normalizedPath)

  if (getStaticBreadcrumbRouteCandidateForPath(normalizedPath)) {
    return localizedNavTarget(normalizedPath)
  }

  if (segments.length === 1) {
    const familyTarget = getBreadcrumbFamilyTarget(segments[0] || '')
    return familyTarget ? localizedNavTarget(familyTarget) : undefined
  }

  return undefined
}

const getRouteFamilyBreadcrumbSubNavigation = (
  currentRootPath: string
): BreadcrumbSubNavigation | undefined => {
  const normalizedRootPath = normalizeBreadcrumbPath(currentRootPath)
  const families = getBreadcrumbRouteFamilies()

  return createBreadcrumbSubNavigation(
    'Site sections',
    families.map(family => ({
      id: family.id,
      label: family.label,
      to: family.to,
      active: isSameOrNestedBreadcrumbPath(normalizedRootPath, family.rootPath),
    }))
  )
}

const pageSubNavigationTabLabel = (tab: PageSubNavigationTab) => {
  return tab.label || tab.fallback || fallbackBreadcrumbRouteFamilyLabel(tab.id)
}

const getBreadcrumbPageSubNavigationTab = (
  targetPath: string
): { entry: PageSubNavigationEntry; tab: PageSubNavigationTab } | null => {
  const normalizedTargetPath = normalizeBreadcrumbPath(targetPath)

  for (const entry of pageSubNavigationEntries) {
    const tabId = getPageSubNavigationTabFromPath(entry.tabs, entry.path, normalizedTargetPath, {
      localeCodes: getAllLocaleCodes(),
      match: 'exact',
    })
    if (!tabId) continue

    const tab = entry.tabs.find(item => item.id === tabId)
    if (tab) return { entry, tab }
  }

  return null
}

const getPageSubNavigationBreadcrumbSubNavigation = (
  targetPath: string
): BreadcrumbSubNavigation | undefined => {
  const normalizedTargetPath = normalizeBreadcrumbPath(targetPath)
  const baseEntry = pageSubNavigationEntries.find(entry => (
    normalizeBreadcrumbPath(entry.path) === normalizedTargetPath
  ))
  const match = baseEntry
    ? { entry: baseEntry }
    : getBreadcrumbPageSubNavigationTab(normalizedTargetPath)
  if (!match) return undefined

  const currentPath = normalizeBreadcrumbPath(route.path || '/')
  const tabs = match.entry.tabs.map(tab => {
    const tabPath = tab.to || pageSubNavigationChildPath(match.entry.path, tab.id)

    return {
      id: `${match.entry.path}:${tab.id}`,
      label: pageSubNavigationTabLabel(tab),
      to: tabPath,
      active: normalizeBreadcrumbPath(tabPath) === currentPath,
    }
  })

  return createBreadcrumbSubNavigation(
    `${getBreadcrumbRouteLabel(match.entry.path, match.entry.path.split('/').filter(Boolean).at(-1) || '')} tabs`,
    tabs
  )
}

const getBreadcrumbSiblingSubNavigation = (
  targetPath: string
): BreadcrumbSubNavigation | undefined => {
  const normalizedTargetPath = normalizeBreadcrumbPath(targetPath)
  const targetSegments = getBreadcrumbPathSegments(normalizedTargetPath)
  const targetDepth = targetSegments.length

  if (targetDepth === 0) return undefined
  const pageSubNavigation = getPageSubNavigationBreadcrumbSubNavigation(normalizedTargetPath)
  if (pageSubNavigation) return pageSubNavigation

  if (targetDepth === 1) return getRouteFamilyBreadcrumbSubNavigation(normalizedTargetPath)

  const parentSegments = targetSegments.slice(0, -1)
  const siblingGroups = getBreadcrumbRouteLevelGroups(parentSegments, targetDepth)
  if (!siblingGroups.some(group => group.path === normalizedTargetPath)) return undefined

  const currentPath = normalizeBreadcrumbPath(route.path || '/')
  const tabs = siblingGroups.map(group => ({
    id: group.id,
    label: getBreadcrumbRouteLabel(group.path, group.segment),
    to: group.target,
    active: isSameOrNestedBreadcrumbPath(currentPath, group.path),
  }))

  return createBreadcrumbSubNavigation(
    `${getBreadcrumbRouteLabel(`/${parentSegments.join('/')}`, parentSegments[parentSegments.length - 1] || '')} pages`,
    tabs
  )
}

const breadcrumbs = computed<BreadcrumbItem[]>(() => {
  const items: BreadcrumbItem[] = [{
    id: 'home',
    label: 'Home',
    to: localePath('/'),
  }]

  const currentPath = normalizeBreadcrumbPath(route.path || '/')
  const segments = getBreadcrumbPathSegments(currentPath)
  if (!segments.length) return items

  segments.forEach((segment, index) => {
    const path = `/${segments.slice(0, index + 1).join('/')}`
    const subNavigation = getBreadcrumbSiblingSubNavigation(path)

    items.push({
      id: `route:${path}`,
      label: index === 0
        ? getBreadcrumbRouteFamilyLabel(segment)
        : getBreadcrumbRouteLabel(path, segment),
      to: getBreadcrumbTarget(path),
      subNavigation,
    })
  })

  return items
})

const isStorefrontCatalogDetailBreadcrumbRoute = computed(() => {
  const segments = getBreadcrumbPathSegments(route.path || '/')
  if (segments[0] === 'products') {
    return segments.length === 2
  }
  return segments[0] === 'shop' && segments.length >= 2
})

const lastExpandableBreadcrumbId = computed(() => {
  for (let index = breadcrumbs.value.length - 1; index >= 0; index--) {
    const crumb = breadcrumbs.value[index]
    if (crumb?.subNavigation?.tabs.length) return crumb.id
  }

  return ''
})

const closeBreadcrumbSubNavState = () => {
  activeBreadcrumbSubNavId.value = null
}

const closeBreadcrumbSubNav = async () => {
  const closePromise = overlayBackStack.close('breadcrumb-subnav')
  closeBreadcrumbSubNavState()
  await closePromise
}

const closeBreadcrumbSubNavForNavigation = async () => {
  const closePromise = overlayBackStack.close('breadcrumb-subnav', 'navigate')
  closeBreadcrumbSubNavState()
  await closePromise
}

const navigateBreadcrumbSubNav = async (to: string) => {
  const targetPath = to || localePath('/')
  const currentFullPath = router.currentRoute.value?.fullPath || route.fullPath || '/'

  try {
    await closeBreadcrumbSubNavForNavigation()
    if (targetPath !== currentFullPath) {
      await router.push(targetPath)
    }
  } catch {
    if (typeof window !== 'undefined') {
      window.location.assign(targetPath)
    }
  }
}

const updateBreadcrumbSubNavPosition = (target: EventTarget | null | undefined) => {
  if (typeof window === 'undefined' || !(target instanceof HTMLElement)) return
  const rect = target.getBoundingClientRect()
  const safeGap = 8
  const nextTop = Math.max(safeGap, Math.min(window.innerHeight - safeGap, rect.bottom + safeGap))
  breadcrumbSubNavMobileTop.value = `${Math.round(nextTop)}px`
}

const toggleBreadcrumbSubNav = (id: string, event?: MouseEvent) => {
  const nextId = activeBreadcrumbSubNavId.value === id ? null : id
  if (!nextId) {
    closeBreadcrumbSubNav()
    return
  }

  if (nextId) {
    updateBreadcrumbSubNavPosition(event?.currentTarget)
  }
  activeBreadcrumbSubNavId.value = nextId
  if (nextId) {
    isOpen.value = false
    overlayBackStack.open('breadcrumb-subnav', closeBreadcrumbSubNavState)
  }
}

const switchLocalePath = (targetLocale: string) => {
	const currentFullPath = router.currentRoute.value?.fullPath || '/'
	// 宽松断言交给 vue-i18n 处理具体的 locale 类型，避免 TS 联合类型报错
	return localePath({ path: currentFullPath }, targetLocale as any)
}

const isOpen = ref(false)
const closeLanguageState = () => {
  isOpen.value = false
}

const closeLanguage = async () => {
  const closePromise = overlayBackStack.close('language')
  closeLanguageState()
  await closePromise
}

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
  return (raw.replace('_', '-').split('-')[0] || '').slice(0, 2).toUpperCase()
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
  if (isOpen.value) {
    closeLanguage()
    return
  }

  isOpen.value = true
  overlayBackStack.open('language', closeLanguageState)
}

const onButtonKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    toggleDropdown()
    if (isOpen.value) {
      nextTick(() => optionRefs.value[0]?.focus())
    }
  } else if (e.key === 'Escape') {
    void closeLanguage()
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
    void closeLanguage()
    closeMegaNav()
    closeBreadcrumbSubNav()
    document.getElementById(buttonId)?.focus()
  }
}

const switchLanguage = async (code: string) => {
  try {
    if (!code || !isLocaleCode(code) || code === locale.value) {
      await closeLanguage()
      return
    }

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
    await closeLanguage()
    closeMegaNav()
    closeBreadcrumbSubNav()
  }
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target
  if (!(target instanceof Element)) return
  if (!target.closest('[data-lang-wrapper]') && !target.closest('#' + dropdownId)) {
    void closeLanguage()
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
	void closeLanguage()
	closeContentNavigationTransition()
	closeGlobalFaqSearch()
	closeMegaNav()
	closeBreadcrumbSubNav()
}

watch(
  () => route.fullPath,
  () => {
    closeMegaNav()
    closeBreadcrumbSubNav()
    nextTick(scheduleHeaderOffsetUpdate)
  },
)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleHeaderKeydown)

  nextTick(() => {
    scheduleHeaderOffsetUpdate()
    window.addEventListener('resize', throttledUpdateHeaderOffset)
    if ('ResizeObserver' in window) {
      headerResizeObserver = new ResizeObserver(() => throttledUpdateHeaderOffset())
      if (headerRootRef.value) {
        headerResizeObserver.observe(headerRootRef.value)
      }
    }
  })

})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleHeaderKeydown)

  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', throttledUpdateHeaderOffset)
    if (headerMetricsFrame !== null) {
      window.cancelAnimationFrame(headerMetricsFrame)
      headerMetricsFrame = null
    }
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

.site-header-brand {
	display: inline-flex;
	min-width: 0;
	align-items: center;
	justify-content: flex-start;
	color: inherit;
	line-height: 1;
	text-decoration: none;
	transition:
		background-color 0.18s ease,
		border-color 0.18s ease,
		box-shadow 0.18s ease;
}

.site-header-brand--image {
	justify-content: center;
	border: 0;
	border-radius: 0;
	background: transparent;
	padding: 0;
	box-shadow: none;
}

.site-header-brand--image:hover,
.site-header-brand--image:focus-visible {
	border: 0;
	background: transparent;
	box-shadow: none;
}

.site-header-brand:focus-visible {
	outline: 2px solid rgba(5, 150, 105, 0.68);
	outline-offset: 4px;
}

.site-header-brand__image {
	display: block;
	max-width: 100%;
	object-fit: contain;
}

.site-header-brand--desktop {
	width: 17.5rem;
	height: 3.4rem;
	max-width: 17.5rem;
	--tz-image-loading-surface: transparent;
}

.site-header-brand--desktop .site-header-brand__image {
	width: 100%;
	height: 100%;
	max-width: 100%;
	max-height: 100%;
	background: transparent !important;
	object-fit: contain;
	object-position: left center;
}

.site-header-brand--mobile {
	width: 100%;
	max-width: 100%;
	justify-content: center;
}

.site-header-brand--mobile .site-header-brand__image {
	max-height: 2.9rem;
}

@media (min-width: 1280px) {
	.site-header-root {
		width: 100%;
		max-width: none;
		transform: none;
	}

	.site-header-mainbar {
		position: relative;
		height: var(--tz-site-header-desktop-mainbar-height);
		min-height: var(--tz-site-header-desktop-mainbar-height);
		box-sizing: border-box;
    background: var(--tz-input-surface);
		border-bottom: 1px solid rgba(20, 32, 43, 0.12);
		box-shadow: 0 12px 28px -16px rgba(20, 32, 43, 0.26);
	}

	.desktop-header-grid {
		display: flex !important;
		grid-template-columns: none;
		align-items: center;
		justify-content: space-between;
	}

	.desktop-header-grid > nav[data-header-mega-nav] {
		position: absolute;
		top: 50%;
		left: 50%;
		z-index: 1;
		transform: translate(-50%, -50%);
	}

	.desktop-header-grid > :first-child,
	.desktop-header-grid > .site-header-actions {
		position: relative;
		z-index: 2;
	}

	.desktop-header-grid > .site-header-actions {
		margin-left: auto;
	}

	.site-header-breadcrumb-row {
		height: var(--tz-site-header-desktop-breadcrumb-height);
		min-height: var(--tz-site-header-desktop-breadcrumb-height);
		box-sizing: border-box;
	}

	.site-header-breadcrumb-list {
		padding: 0.16rem 1rem 0.42rem;
	}
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

.site-header-actions {
	--site-header-action-width: 4.5rem;
	--site-header-action-height: 2.25rem;
	--site-header-action-icon-size: 1.375rem;
	display: flex;
	align-items: center;
	gap: 0.38rem !important;
}

.site-header-action-cell {
	display: inline-flex;
	flex: 0 0 var(--site-header-action-width);
	width: var(--site-header-action-width);
	min-width: 0;
	height: var(--site-header-action-height);
	align-items: center;
	justify-content: center;
	margin: 0;
}

.site-header-action-cell--search {
	position: relative;
	overflow: visible;
}

@media (min-width: 1280px) {
	.site-header-action-cell--search {
		flex-basis: var(--site-header-action-width);
		width: var(--site-header-action-width);
	}
}

@media (min-width: 1024px) {
	.site-header-action-cell--search {
		flex-basis: var(--site-header-action-width);
		width: var(--site-header-action-width);
	}
}

@media (min-width: 1280px) {
	.site-header-action-cell--search {
		flex-basis: var(--site-header-action-width);
		width: var(--site-header-action-width);
	}
}

.site-header-language-wrapper {
	margin: 0;
}

.site-header-action-button {
	position: relative;
	box-sizing: border-box;
	display: inline-flex;
	overflow: hidden;
	width: 100% !important;
	min-width: 0 !important;
	height: 100% !important;
	min-height: 0 !important;
	align-items: center;
	justify-content: center;
	margin: 0 !important;
	border: 0 !important;
	border-radius: 9999px;
	background: transparent !important;
	background-image: none !important;
	color: var(--tz-text-primary);
	padding: 0 0.5rem !important;
	letter-spacing: 0;
	line-height: 1;
	box-shadow: none !important;
	transition:
		background 0.18s ease,
		border-color 0.18s ease,
		color 0.18s ease,
		transform 0.18s ease;
}

.site-header-action-button:hover,
.site-header-action-button:focus-visible,
.site-header-action-button[aria-expanded='true'] {
	border: 0 !important;
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
	color: var(--tz-text-primary);
}

.site-header-action-button:focus-visible {
	outline: none !important;
}

.site-header-language-trigger {
	gap: 0.34rem !important;
	border: 0 !important;
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
	color: var(--tz-text-secondary);
	padding-inline: 0.42rem !important;
}

.site-header-language-trigger:hover,
.site-header-language-trigger:focus-visible,
.site-header-language-trigger[aria-expanded='true'] {
	border: 0 !important;
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
	color: var(--tz-text-primary);
}

.site-header-language-trigger > span:first-child {
	width: var(--site-header-action-icon-size) !important;
	height: var(--site-header-action-icon-size) !important;
	flex: 0 0 var(--site-header-action-icon-size) !important;
}

.site-header-language-trigger img {
	display: block;
	width: var(--site-header-action-icon-size) !important;
	height: var(--site-header-action-icon-size) !important;
	flex: 0 0 var(--site-header-action-icon-size) !important;
}

.site-header-language-trigger .iconify,
.site-header-language-trigger svg {
	width: var(--site-header-action-icon-size) !important;
	height: var(--site-header-action-icon-size) !important;
	flex: 0 0 var(--site-header-action-icon-size) !important;
}

.site-header-membership-trigger {
	padding: 0 !important;
}

.site-header-membership-trigger:hover,
.site-header-membership-trigger:focus-visible,
.site-header-membership-trigger[aria-expanded='true'] {
	border: 0 !important;
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
	color: #059669;
}

.site-header-membership-trigger svg {
	width: var(--site-header-action-icon-size) !important;
	height: var(--site-header-action-icon-size) !important;
	flex: 0 0 var(--site-header-action-icon-size) !important;
}

.site-header-search-trigger {
	gap: 0 !important;
	justify-content: center;
	border: 0 !important;
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
	color: var(--tz-text-primary);
	padding: 0 !important;
	font-size: 0.82rem;
	font-weight: 720;
}

.site-header-search-trigger:hover,
.site-header-search-trigger:focus-visible {
	border: 0 !important;
	background: transparent !important;
	background-image: none !important;
	box-shadow: none !important;
	color: var(--tz-text-primary);
}

.site-header-search-trigger__icon {
	width: var(--site-header-action-icon-size);
	height: var(--site-header-action-icon-size);
	flex: 0 0 var(--site-header-action-icon-size);
	margin: 0;
	animation: site-header-search-nudge 30s ease-in-out infinite;
}

.site-header-search-trigger__label {
	display: none;
	min-width: 0;
	align-items: center;
	overflow: hidden;
	padding: 0;
	text-overflow: ellipsis;
	text-transform: none;
	white-space: nowrap;
}

.site-header-search-trigger--mobile {
	width: 100%;
	min-width: 0;
	height: var(--site-header-action-height);
	justify-content: center;
	padding: 0 !important;
	gap: 0 !important;
}

.site-header-search-trigger--mobile .site-header-search-trigger__label {
	display: none;
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.site-header-search-hint {
	position: absolute;
	top: calc(100% + 0.55rem);
	left: 50%;
	z-index: 1250;
	display: grid;
	width: max-content;
	max-width: min(18rem, 72vw);
	transform: translate(-50%, -0.2rem);
	gap: 0.25rem;
	padding: 0.72rem 0.86rem;
	border: 1px solid rgba(20, 32, 43, 0.14);
	border-radius: 10px;
	background: #ffffff;
	box-shadow: 0 12px 34px rgba(20, 32, 43, 0.16);
	opacity: 0;
	pointer-events: none;
	transition: opacity 0.18s ease, transform 0.18s ease;
	animation: site-header-search-hint-cycle 30s ease-in-out infinite;
}

.site-header-search-hint::before {
	content: '';
	position: absolute;
	top: -5px;
	left: 50%;
	width: 9px;
	height: 9px;
	transform: translateX(-50%) rotate(45deg);
	border-top: 1px solid rgba(20, 32, 43, 0.14);
	border-left: 1px solid rgba(20, 32, 43, 0.14);
	background: #ffffff;
}

.site-header-search-hint__title {
	color: var(--tz-text-primary);
	font-size: 0.78rem;
	font-weight: 720;
	line-height: 1.25;
	white-space: nowrap;
}

.site-header-search-hint__body {
	color: var(--tz-text-secondary);
	font-size: 0.72rem;
	font-weight: 500;
	line-height: 1.35;
}

.site-header-action-cell--search:hover .site-header-search-hint,
.site-header-action-cell--search:focus-within .site-header-search-hint {
	transform: translate(-50%, 0);
	animation: none;
	opacity: 1;
}

.site-header-search-hint--mobile {
	top: calc(100% + 0.42rem);
	left: 50%;
	max-width: min(16rem, 78vw);
}

.site-header-actions--mobile {
	--site-header-action-height: 2.25rem;
	display: flex;
	width: 100%;
	max-width: 100%;
	align-items: center;
	justify-content: flex-end;
	margin-left: 0;
	gap: 0.625rem !important;
}

.site-header-actions--mobile .site-header-action-cell,
.site-header-actions--mobile .site-header-language-wrapper {
	width: auto;
	min-width: 0;
	flex: 0 0 auto;
}

.site-header-actions--mobile .site-header-search-trigger--mobile {
	width: var(--site-header-action-height);
}

@keyframes site-header-search-nudge {
	0%,
	84%,
	100% {
		transform: translateY(0) scale(1);
	}

	87% {
		transform: translateY(-1px) scale(1.06);
	}

	90% {
		transform: translateY(0) scale(1);
	}

	93% {
		transform: translateY(-1px) scale(1.04);
	}
}

@keyframes site-header-search-hint-cycle {
	0%,
	72%,
	100% {
		opacity: 0;
		transform: translate(-50%, -0.2rem);
	}

	76%,
	88% {
		opacity: 1;
		transform: translate(-50%, 0);
	}
}

@media (prefers-reduced-motion: reduce) {
	.site-header-search-trigger__icon,
	.site-header-search-hint {
		animation: none;
	}
}

.site-header-menu-laser {
	position: relative;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	gap: 0.375rem;
	border: 0;
	border-radius: 9999px;
	background: transparent;
	padding: 0.5rem 0.75rem 0.75rem;
	color: var(--tz-text-primary);
	font-size: 1.125rem;
	font-weight: 800;
	letter-spacing: 0;
	line-height: 1;
	transition: color 0.3s ease;
}

.site-header-menu-laser::after {
	position: absolute;
	bottom: 0.25rem;
	left: 50%;
	width: 0;
	height: 2px;
	content: '';
	transform: translateX(-50%);
	border-radius: 9999px;
	background: #059669;
	box-shadow: 0 0 10px #059669;
	transition:
		width 0.35s cubic-bezier(0.4, 0, 0.2, 1),
		opacity 0.3s ease;
	opacity: 0;
}

.site-header-menu-laser__text {
  display: inline-block;
  transform-origin: center;
  transition:
    color 0.3s ease,
    transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  text-transform: uppercase;
  white-space: nowrap;
}

.site-header-menu-laser__icon {
	width: 0.875rem;
	height: 0.875rem;
	flex-shrink: 0;
	transition:
		color 0.3s ease,
		transform 0.2s ease;
}

.site-header-menu-laser:hover,
.site-header-menu-laser:focus-visible,
.site-header-menu-laser[aria-expanded='true'] {
	color: var(--tz-text-primary);
}

.site-header-menu-laser:hover::after,
.site-header-menu-laser:focus-visible::after,
.site-header-menu-laser[aria-expanded='true']::after,
.site-header-menu-laser--active::after {
	width: calc(100% - 1.5rem);
	opacity: 1;
}

.site-header-menu-laser:hover .site-header-menu-laser__text,
.site-header-menu-laser:focus-visible .site-header-menu-laser__text,
.site-header-menu-laser[aria-expanded='true'] .site-header-menu-laser__text {
	transform: scale(1.08);
	color: var(--tz-text-primary);
}

.site-header-menu-laser--active {
	color: #059669;
	font-weight: 900;
}

.site-header-menu-laser--active .site-header-menu-laser__text {
	color: #059669;
}

.site-header-menu-laser:focus-visible {
	outline: 1px solid rgb(16 185 129 / 0.55);
	outline-offset: 0.16rem;
}

.site-header-mobile-shop-slot {
	min-width: 0;
}

.site-header-mobile-shop-trigger {
	display: inline-flex;
	width: var(--site-header-action-height, 2.25rem);
	height: var(--site-header-action-height, 2.25rem);
	align-items: center;
	justify-content: center;
	border: 0;
	border-radius: 9999px;
	background: transparent;
	padding: 0;
	color: var(--tz-text-primary);
	cursor: pointer;
	transition:
		background-color 0.18s ease,
		color 0.18s ease,
		transform 0.18s ease;
}

.site-header-mobile-shop-trigger:hover,
.site-header-mobile-shop-trigger:focus-visible,
.site-header-mobile-shop-trigger[aria-expanded='true'] {
	background: rgba(5, 150, 105, 0.1);
	color: #059669;
	transform: translateY(-1px);
}

.site-header-mobile-shop-trigger:focus-visible {
	outline: 1px solid rgb(16 185 129 / 0.55);
	outline-offset: 0.12rem;
}

.site-header-mobile-shop-trigger--active {
	color: #059669;
}

.site-header-mobile-nav-row {
	display: flex;
	overflow: hidden;
	min-height: 2.875rem;
	align-items: stretch;
    background: var(--tz-mobile-bottom-chrome-surface);
	border-bottom: 1px solid var(--tz-mobile-chrome-edge-border);
}

.site-header-mobile-nav__button {
	position: relative;
	overflow: hidden;
	border: 1px solid transparent;
	background: transparent;
	color: var(--tz-text-secondary);
}

.site-header-mobile-nav__button::after {
	position: absolute;
	bottom: 0.18rem;
	left: 50%;
	width: 0;
	height: 2px;
	content: '';
	transform: translateX(-50%);
	border-radius: 9999px;
	background: #059669;
	box-shadow: 0 0 10px rgb(16 185 129 / 0.72);
	opacity: 0;
	transition:
		width 0.28s cubic-bezier(0.4, 0, 0.2, 1),
		opacity 0.22s ease;
}

.site-header-mobile-nav__button:hover,
.site-header-mobile-nav__button:focus-visible,
.site-header-mobile-nav__button[aria-expanded='true'] {
	background: transparent;
	color: var(--tz-text-primary);
}

.site-header-mobile-nav__button:focus-visible {
	outline: 1px solid rgb(16 185 129 / 0.55);
	outline-offset: 0.12rem;
}

.site-header-mobile-nav__button:hover::after,
.site-header-mobile-nav__button:focus-visible::after,
.site-header-mobile-nav__button[aria-expanded='true']::after,
.site-header-mobile-nav__button--active::after {
	width: calc(100% - 1.25rem);
	opacity: 1;
}

.site-header-mobile-nav__button--active {
	color: #059669;
	font-weight: 800;
}

.language-dropdown-surface {
	background: #ffffff !important;
}

@media (max-width: 1279px) {
	.language-dropdown-layer {
		top: calc(var(--site-header-mobile-topbar-bottom, 3.5rem) + 1px);
		right: 1px;
		bottom: 1px;
		left: 1px;
		background: transparent;
		pointer-events: none;
	}

	.language-dropdown-surface {
		width: 100% !important;
		height: 100% !important;
		max-height: none !important;
		row-gap: 0.25rem;
		column-gap: 0.375rem;
		align-content: start;
		padding: 0.5rem !important;
		border-radius: 1rem !important;
		background: #ffffff !important;
		box-shadow: 0 18px 42px rgba(20, 32, 43, 0.16);
		pointer-events: auto;
	}

	.language-dropdown-surface > [role='option'] {
		min-height: 2.5rem;
		padding-block: 0.5rem !important;
	}
}

.site-header-breadcrumb-row {
	display: flex;
	width: 100%;
	min-width: 0;
	align-items: center;
	background: #f7f9fa;
	border-top: 1px solid rgba(20, 32, 43, 0.1);
	border-bottom: 1px solid rgba(20, 32, 43, 0.1);
	overflow: visible;
	scrollbar-width: none;
}

.site-header-breadcrumb-row::-webkit-scrollbar {
	display: none;
}

.site-header-breadcrumb-list {
	display: flex;
	width: max-content;
	min-width: 100%;
	flex: 0 0 auto;
	flex-wrap: nowrap;
	align-items: center;
	justify-content: center;
	padding: 0.32rem 0.75rem 0.38rem;
	white-space: nowrap;
}

.site-header-breadcrumb-item {
	flex: 0 0 auto;
	min-width: 0;
	white-space: nowrap;
}

.site-header-breadcrumb-separator {
	display: inline-flex;
	flex: 0 0 auto;
	align-items: center;
	justify-content: center;
	min-width: 0.42rem;
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
	color: var(--tz-text-primary);
}

.breadcrumb-subnav-trigger:focus-visible {
	outline: 1px solid rgba(5, 150, 105, 0.72);
	outline-offset: 0.18rem;
}

.breadcrumb-subnav-trigger span {
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.breadcrumb-subnav-trigger__icon {
	width: 0.82rem;
	height: 0.82rem;
	min-width: 0.82rem;
	min-height: 0.82rem;
	display: block;
	flex: 0 0 0.82rem;
	color: #059669;
	transition: transform 0.18s ease;
}

.breadcrumb-subnav-pulse-dot {
	width: 0.44rem;
	height: 0.44rem;
	flex: 0 0 auto;
	border-radius: 9999px;
	background: #059669;
	box-shadow:
		0 0 0 0 rgba(5, 150, 105, 0.46),
		0 0 10px rgba(5, 150, 105, 0.72);
	animation: breadcrumb-pulse-dot 1.35s ease-in-out infinite;
}

@keyframes breadcrumb-pulse-dot {
	0%,
	100% {
		opacity: 0.58;
		transform: scale(0.86);
		box-shadow:
			0 0 0 0 rgba(5, 150, 105, 0.42),
			0 0 8px rgba(5, 150, 105, 0.56);
	}

	50% {
		opacity: 1;
		transform: scale(1);
		box-shadow:
			0 0 0 5px rgba(5, 150, 105, 0),
			0 0 12px rgba(5, 150, 105, 0.86);
	}
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
	border: 1px solid var(--tz-border-subtle);
	border-radius: 0.95rem;
	background: var(--tz-card-surface, #111116);
	background-image: none;
	padding: 0.42rem;
	box-shadow:
		0 24px 54px -22px rgba(20, 32, 43, 0.16),
		inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.breadcrumb-subnav-link {
	display: flex;
	align-items: center;
	justify-content: space-between;
	border-radius: 0.72rem;
	padding: 0.58rem 0.72rem;
	color: var(--tz-text-secondary);
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
	background: rgba(20, 32, 43, 0.05);
	color: var(--tz-text-primary);
	transform: translateY(-1px);
}

.breadcrumb-subnav-link:focus-visible {
	outline: 1px solid rgba(5, 150, 105, 0.72);
	outline-offset: 0.12rem;
}

.breadcrumb-subnav-link--active,
.breadcrumb-subnav-link--active:hover,
.breadcrumb-subnav-link--active:focus-visible {
	background: rgba(5, 150, 105, 0.24);
	color: var(--tz-text-primary);
}

.breadcrumb-subnav-trigger--mobile {
	max-width: 36vw;
}

.site-header-breadcrumb-row--mobile {
	width: calc(100% + 2rem);
	margin-inline: -1rem;
	overflow-x: auto;
	overflow-y: visible;
}

.site-header-breadcrumb-row--mobile .site-header-breadcrumb-list {
	min-width: max-content;
	justify-content: flex-start;
	padding-inline: max(0.75rem, env(safe-area-inset-left)) max(0.75rem, env(safe-area-inset-right));
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

@media (max-width: 1279px) {
	.site-header-root {
		height: var(--tz-site-header-mobile-stack-height);
		max-height: var(--tz-site-header-mobile-stack-height);
	}

	.site-header-surface {
		padding: 0 !important;
	}

	.site-header-language-trigger {
		height: var(--site-header-action-height);
		min-width: 0;
		gap: 0.36rem !important;
		padding-inline: 0.3rem !important;
	}

	.site-header-language-label--mobile {
		display: none;
	}

	.site-header-search-trigger {
		margin: 0 !important;
	}

	.site-header-search-trigger__icon {
		margin-right: 0;
	}

	.site-header-search-hint--mobile {
		display: none !important;
		animation: none !important;
	}

	.site-header-language-wrapper {
		margin: 0;
	}

	.site-header-membership-trigger {
		margin-left: 0;
	}

	.site-header-mobile-nav {
		background: transparent !important;
		box-shadow: none !important;
		height: 100%;
		min-width: 0;
		min-height: 0;
		align-items: stretch;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		overflow: hidden;
	}

	.site-header-mobile-surface {
		width: 100%;
		margin: 0;
		padding-right: 0;
		padding-left: 0;
		padding-top: 0;
		height: var(--tz-site-header-mobile-topbar-height);
		min-height: var(--tz-site-header-mobile-topbar-height);
		box-sizing: border-box;
		background: var(--tz-input-surface);
		border-bottom: 0;
		box-shadow: none;
	}

	.site-header-mobile-topbar {
		height: var(--tz-site-header-mobile-topbar-height);
		min-height: var(--tz-site-header-mobile-topbar-height);
		box-sizing: border-box;
		padding-right: 0;
		padding-left: 0;
		padding-top: 0;
		padding-bottom: 0;
		border-bottom: 0;
	}

	.site-header-brand--mobile {
		--tz-image-loading-surface: transparent;
		background: transparent !important;
	}

	.site-header-brand--mobile :deep(.tz-storefront-image),
	.site-header-brand--mobile .site-header-brand__image {
		width: 48px;
		height: 48px;
		max-width: 48px;
		max-height: 48px;
		background: transparent !important;
		object-fit: contain;
	}

	.site-header-actions--mobile .site-header-action-button,
	.site-header-actions--mobile .site-header-language-trigger,
	.site-header-actions--mobile .site-header-search-trigger {
		color: var(--tz-text-primary) !important;
	}

	.site-header-actions--mobile .site-header-action-button:hover,
	.site-header-actions--mobile .site-header-action-button:focus-visible,
	.site-header-actions--mobile .site-header-action-button[aria-expanded='true'],
	.site-header-actions--mobile .site-header-language-trigger:hover,
	.site-header-actions--mobile .site-header-language-trigger:focus-visible,
	.site-header-actions--mobile .site-header-language-trigger[aria-expanded='true'],
	.site-header-actions--mobile .site-header-search-trigger:hover,
	.site-header-actions--mobile .site-header-search-trigger:focus-visible {
		color: var(--tz-text-primary) !important;
	}

	.site-header-mobile-nav-row {
		--site-header-mobile-nav-inline-padding: 0px;
		width: 100%;
		margin: 0;
		padding-right: var(--site-header-mobile-nav-inline-padding);
		padding-left: var(--site-header-mobile-nav-inline-padding);
		height: var(--tz-site-header-mobile-nav-height);
		min-height: var(--tz-site-header-mobile-nav-height);
		box-sizing: border-box;
    background: var(--tz-mobile-bottom-chrome-surface);
		border-top: 0;
		border-bottom: 1px solid var(--tz-mobile-chrome-edge-border);
	}

	.site-header-mobile-nav__button {
		height: 100%;
		min-width: 0 !important;
		width: 100% !important;
		gap: 0.18rem;
		overflow: hidden;
		padding-top: 0;
		padding-bottom: 0;
		color: var(--tz-text-secondary);
		text-transform: uppercase;
	}

	.site-header-mobile-nav__button > span:first-child {
		min-width: 0;
		max-width: calc(100% - 1rem);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.site-header-mobile-nav__button > :deep(svg) {
		flex: 0 0 0.875rem;
		width: 0.875rem;
		height: 0.875rem;
	}

	.site-header-mobile-nav__button:hover,
	.site-header-mobile-nav__button:focus-visible,
	.site-header-mobile-nav__button[aria-expanded='true'] {
		color: var(--tz-text-primary);
	}

	.site-header-mobile-nav__button--active {
		color: #059669;
	}

	.site-header-breadcrumb-row--mobile {
		width: 100%;
		height: var(--tz-site-header-mobile-breadcrumb-height);
		min-height: var(--tz-site-header-mobile-breadcrumb-height);
		margin-inline: 0;
		align-items: stretch;
    background: var(--tz-mobile-top-chrome-surface);
		border-top: 0;
		border-bottom: 1px solid var(--tz-mobile-chrome-edge-border);
	}

	.site-header-breadcrumb-row--mobile .site-header-breadcrumb-list {
		--site-header-mobile-breadcrumb-inline-padding: 0.75rem;
		height: 100%;
		min-height: 0;
		padding-top: 0;
		padding-bottom: 0;
		padding-inline: max(var(--site-header-mobile-breadcrumb-inline-padding), env(safe-area-inset-left)) max(var(--site-header-mobile-breadcrumb-inline-padding), env(safe-area-inset-right));
		color: var(--tz-text-secondary);
	}

	.breadcrumb-subnav-trigger--mobile {
		color: var(--tz-text-secondary);
	}

	.breadcrumb-subnav-trigger--mobile:hover,
	.breadcrumb-subnav-trigger--mobile:focus-visible,
	.breadcrumb-subnav-trigger--mobile.breadcrumb-subnav-trigger--open {
		color: var(--tz-text-primary);
	}

}

/* iPad / small tablets: prevent desktop language switcher from overflowing header pill */
@media (min-width: 768px) and (max-width: 1100px) {
	.site-header-root {
		width: 100%;
	}

	.desktop-header-grid {
		display: grid !important;
		grid-template-columns: 168px minmax(0, 1fr) 152px;
		gap: 0.5rem;
		align-items: center;
	}

	.desktop-header-grid > :first-child {
		min-width: 0;
	}

	.desktop-header-grid > nav {
		position: static;
		min-width: 0;
		width: 100%;
		transform: none;
		justify-content: space-between;
		gap: 0;
	}

	.site-header-menu-laser {
		flex: 1 1 0;
		min-width: 0;
		padding-inline: 0.35rem;
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

@media (max-width: 1279px) {
	.leverandpoint-modal-shell {
		height: min(95vh, calc(var(--tz-mobile-safe-viewport-height, 100vh) - 16px));
		max-height: min(95vh, calc(var(--tz-mobile-safe-viewport-height, 100vh) - 16px));
	}

	@supports (height: 100svh) {
		.leverandpoint-modal-shell {
			height: min(95svh, calc(var(--tz-mobile-safe-viewport-height, 100svh) - 16px));
			max-height: min(95svh, calc(var(--tz-mobile-safe-viewport-height, 100svh) - 16px));
		}
	}

	@supports (height: 100dvh) {
		.leverandpoint-modal-shell {
			height: min(95dvh, calc(var(--tz-mobile-safe-viewport-height, 100dvh) - 16px));
			max-height: min(95dvh, calc(var(--tz-mobile-safe-viewport-height, 100dvh) - 16px));
		}
	}
}

/* tablet-820: 820x1180 等宽度段，限制 SiteHeader 高度为 130px */
@media (min-width: 820px) and (max-width: 1023px) {
	.site-header-root {
		max-height: 130px;
	}
}

/* Light storefront chrome: keep the brand green for active states and focus. */
:global(.tz-light-theme) .site-header-surface {
	background: #ffffff;
	color: var(--tz-text-primary);
}

:global(.tz-light-theme) .site-header-mainbar {
    background: var(--tz-input-surface) !important;
	border-bottom: 1px solid rgba(20, 32, 43, 0.12);
	box-shadow: 0 12px 28px -16px rgba(20, 32, 43, 0.26) !important;
}

:global(.tz-light-theme) .site-header-action-button,
:global(.tz-light-theme) .site-header-language-trigger,
:global(.tz-light-theme) .site-header-menu-laser {
	color: var(--tz-text-secondary);
}

:global(.tz-light-theme) .site-header-action-button:hover,
:global(.tz-light-theme) .site-header-action-button:focus-visible,
:global(.tz-light-theme) .site-header-action-button[aria-expanded='true'],
:global(.tz-light-theme) .site-header-language-trigger:hover,
:global(.tz-light-theme) .site-header-language-trigger:focus-visible,
:global(.tz-light-theme) .site-header-language-trigger[aria-expanded='true'],
:global(.tz-light-theme) .site-header-menu-laser:hover,
:global(.tz-light-theme) .site-header-menu-laser:focus-visible,
:global(.tz-light-theme) .site-header-menu-laser[aria-expanded='true'] {
	color: var(--tz-text-primary);
}

:global(.tz-light-theme) .site-header-mobile-nav-row {
    background: var(--tz-mobile-bottom-chrome-surface);
	border-bottom-color: var(--tz-mobile-chrome-edge-border);
}

:global(.tz-light-theme) .site-header-mobile-nav__button {
	color: var(--tz-text-secondary);
}

:global(.tz-light-theme) .site-header-mobile-nav__button:hover,
:global(.tz-light-theme) .site-header-mobile-nav__button:focus-visible,
:global(.tz-light-theme) .site-header-mobile-nav__button[aria-expanded='true'] {
	color: var(--tz-text-primary);
}

:global(.tz-light-theme) .site-header-breadcrumb-row,
:global(.tz-light-theme) .site-header-breadcrumb-row--mobile {
	background: #f7f9fa;
	border-color: rgba(20, 32, 43, 0.1);
}

:global(.tz-light-theme) .language-dropdown-surface,
:global(.tz-light-theme) .breadcrumb-subnav-menu {
	background: #ffffff !important;
	border-color: rgba(20, 32, 43, 0.12);
	box-shadow: 0 18px 42px rgba(20, 32, 43, 0.16);
}

:global(.tz-light-theme) .breadcrumb-subnav-link {
	color: var(--tz-text-secondary);
}

:global(.tz-light-theme) .breadcrumb-subnav-link:hover,
:global(.tz-light-theme) .breadcrumb-subnav-link:focus-visible {
	background: #f1f4f6;
	color: var(--tz-text-primary);
}

</style>
