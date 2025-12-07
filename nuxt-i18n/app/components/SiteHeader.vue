<template>
	<div class="fixed top-1.5 left-1/2 -translate-x-1/2 w-[95vw] max-w-[1200px] z-[110] site-header-root">
		<div
			class="relative w-full rounded-[30px] bg-[#0b1020]/70 backdrop-blur-md border border-white/10 shadow-[0_18px_45px_rgba(15,23,42,0.9)] px-4 py-2 md:py-2"
		>
			<!-- 桌面端：宽版极简胶囊布局 (Option B Wide) -->
			<div class="hidden md:flex flex-col items-center gap-1">
				<!-- 主胶囊 Header -->
				<div class="w-full grid grid-cols-[200px_1fr_200px] items-center gap-4 px-6 py-3 rounded-[99px] bg-[#0b1020]/80 backdrop-blur-md border border-white/10 shadow-lg">
					
					<!-- Logo -->
					<div class="flex items-center justify-start">
						<h1 class="m-0 text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)] leading-none italic cursor-default">
							{{ titleText }}
						</h1>
					</div>

					<!-- Nav (Centered) -->
					<nav class="flex items-center justify-center gap-8 relative">
						<!-- Vertical Divider Left -->
						<div class="w-px h-6 bg-white/10 absolute left-0 hidden xl:block"></div>
						
						<NuxtLink
							:to="localePath('/products')"
							class="text-[15px] font-medium text-white/60 hover:text-white transition-colors relative group"
							active-class="!text-white"
						>
							{{ $t('footer.menus.products', 'Products') }}
							<span 
								class="absolute bottom-[-4px] left-0 right-0 h-[2px] bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] rounded-full transition-opacity"
								:class="route.path.startsWith(localePath('/products')) ? 'opacity-100' : 'opacity-0'"
							></span>
						</NuxtLink>
						<NuxtLink
							:to="localePath('/support')"
							class="text-[15px] font-medium text-white/60 hover:text-white transition-colors relative group"
							active-class="!text-white"
						>
							{{ $t('footer.menus.support', 'Support') }}
							<span 
								class="absolute bottom-[-4px] left-0 right-0 h-[2px] bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] rounded-full transition-opacity"
								:class="route.path.startsWith(localePath('/support')) ? 'opacity-100' : 'opacity-0'"
							></span>
						</NuxtLink>
						<NuxtLink
							:to="localePath('/company')"
							class="text-[15px] font-medium text-white/60 hover:text-white transition-colors relative group"
							active-class="!text-white"
						>
							{{ $t('footer.menus.company', 'Company') }}
							<span 
								class="absolute bottom-[-4px] left-0 right-0 h-[2px] bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] rounded-full transition-opacity"
								:class="route.path.startsWith(localePath('/company')) ? 'opacity-100' : 'opacity-0'"
							></span>
						</NuxtLink>
						<NuxtLink
							:to="localePath('/guides')"
							class="text-[15px] font-medium text-white/60 hover:text-white transition-colors relative group"
							active-class="!text-white"
						>
							Guides
							<span 
								class="absolute bottom-[-4px] left-0 right-0 h-[2px] bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] rounded-full transition-opacity"
								:class="route.path.startsWith(localePath('/guides')) ? 'opacity-100' : 'opacity-0'"
							></span>
						</NuxtLink>

						<!-- Vertical Divider Right -->
						<div class="w-px h-6 bg-white/10 absolute right-0 hidden xl:block"></div>
					</nav>

					<!-- Right Actions -->
					<div class="flex items-center justify-end gap-4">
						<!-- Search -->
						<button
							class="w-9 h-9 rounded-full border border-white/5 bg-white/5 flex items-center justify-center text-white/80 hover:text-white hover:bg-white/10 transition-all"
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
								class="w-9 h-9 rounded-full border border-white/5 bg-white/5 flex items-center justify-center text-white/80 hover:text-white hover:bg-white/10 transition-all"
								@click.stop="toggleDropdown"
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
										class="fixed top-[80px] left-1/2 -translate-x-1/2 w-[90vw] max-w-[1600px] bg-[#0b1020] border border-[#6b79ff] rounded-md overflow-auto max-h-[70vh] shadow-none grid grid-cols-[repeat(auto-fit,minmax(160px,1fr))] gap-1.5 justify-items-center z-[1200]"
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
											:ref="el => setOptionRef(el, index)"
											@click="switchLanguage(locale.code)"
										>
											<span class="w-[1.2em] inline-block" aria-hidden="true">
												<img :src="flagSrc(locale)" alt="" class="w-[1.2em] h-[1.2em] block" />
											</span>
											<span>{{ locale.name }}</span>
										</button>
									</div>
								</transition>
							</teleport>
						</div>

						<!-- Share/Points -->
						<button
							class="w-10 h-10 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] flex items-center justify-center text-[#0b1020] shadow-lg hover:shadow-[#40ffaa]/30 transition-all transform hover:scale-105"
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
					<ol class="flex items-center gap-1.5 text-[11px] text-slate-500 leading-tight transition-colors hover:text-slate-400">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="index"
							class="flex items-center gap-1"
						>
							<NuxtLink
								v-if="crumb.to && index < breadcrumbs.length - 1"
								:to="crumb.to"
								class="hover:text-white transition-colors"
							>
								{{ crumb.label }}
							</NuxtLink>
							<span v-else class="text-slate-300 font-medium">
								{{ crumb.label }}
							</span>
							<span v-if="index < breadcrumbs.length - 1" class="text-slate-700">/</span>
						</li>
					</ol>
				</nav>
			</div>

			<!-- 移动端：新版极简双行布局 -->
			<div class="md:hidden flex flex-col gap-3">
				
				<!-- 第一行：Logo (左) + 工具图标 (右) -->
				<div class="flex items-center justify-between px-1">
					<!-- Logo -->
					<h1 class="m-0 text-2xl phone-390:text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)] leading-none">
						{{ titleText }}
					</h1>

					<!-- 右侧工具图标组 -->
					<div class="flex items-center gap-3 phone-390:gap-4">
						<!-- Search (Icon) -->
						<button
							class="text-white/70 hover:text-white transition-colors p-1"
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
							:to="localePath('/guides')"
							class="text-white/70 hover:text-white transition-colors p-1"
							aria-label="Guides"
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
						</NuxtLink>

						<!-- Language Switcher (Text + Icon) -->
						<div class="relative" data-lang-wrapper>
							<button
								class="text-white/70 hover:text-white transition-colors flex items-center gap-1 p-1"
								@click.stop="toggleDropdown"
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
							class="text-white/70 hover:text-[#40ffaa] transition-colors p-1"
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
						:to="localePath('/products')"
						class="flex-1 py-2 rounded-lg text-xs phone-390:text-[13px] font-medium text-center transition-all"
						:class="route.path.startsWith(localePath('/products')) ? 'bg-white/10 text-white shadow-sm border border-white/10' : 'text-white/60 hover:text-white'"
					>
						{{ $t('footer.menus.products', 'Products') }}
					</NuxtLink>
					
					<NuxtLink
						:to="localePath('/support')"
						class="flex-1 py-2 rounded-lg text-xs phone-390:text-[13px] font-medium text-center transition-all"
						:class="route.path.startsWith(localePath('/support')) ? 'bg-white/10 text-white shadow-sm border border-white/10' : 'text-white/60 hover:text-white'"
					>
						{{ $t('footer.menus.support', 'Support') }}
					</NuxtLink>
					
					<NuxtLink
						:to="localePath('/company')"
						class="flex-1 py-2 rounded-lg text-xs phone-390:text-[13px] font-medium text-center transition-all"
						:class="route.path.startsWith(localePath('/company')) ? 'bg-white/10 text-white shadow-sm border border-white/10' : 'text-white/60 hover:text-white'"
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
					<ol class="flex items-center gap-1.5 flex-wrap justify-center text-[10px] text-slate-400 leading-tight">
						<li
							v-for="(crumb, index) in breadcrumbs"
							:key="index"
							class="flex items-center gap-1"
						>
							<NuxtLink
								v-if="crumb.to && index < breadcrumbs.length - 1"
								:to="crumb.to"
								class="hover:text-white transition-colors truncate max-w-[100px]"
							>
								{{ crumb.label }}
							</NuxtLink>
							<span v-else class="text-slate-200 font-medium truncate max-w-[120px]">
								{{ crumb.label }}
							</span>
							<span v-if="index < breadcrumbs.length - 1" class="text-slate-600">/</span>
						</li>
					</ol>
				</nav>

			</div>
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
				<!-- 弹窗内容 -->
				<div
					class="relative w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] flex pointer-events-auto"
					aria-modal="true"
					role="dialog"
					aria-label="Membership"
				>
					<LeverAndPoint @close="shareOpen = false" />
				</div>
			</div>
		</transition>
	</teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch, type ComponentPublicInstance } from 'vue'
import { useLocalePath, useRoute } from '#imports'
import { useSiteTitle } from '~/composables/useSiteTitle'
import LeverAndPoint from '~/components/LeverAndPoint.vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'
import { productsNavItems } from '~/utils/productsNav'
import { supportNavItems } from '~/utils/supportNav'
import { companyNavItems } from '~/utils/companyNav'

// Site Title
const props = defineProps<{ title?: string }>()
const { siteTitle } = useSiteTitle()
const titleText = computed(() => {
  const fromProp = (props.title ?? '').toString().trim()
  return fromProp.length ? fromProp : siteTitle.value
})

// Share button (Membership panel)
const shareOpen = ref(false)

const toggleShare = () => {
  shareOpen.value = !shareOpen.value
  if (shareOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'header-share' } }))
  }
}

// Open Sidebar (Search)
const openSidebar = () => {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:sidebar-open', { detail: { side: 'left' } }))
  }
}

// Language Switcher
const { locale, locales, setLocale, t } = useI18n()
const localePath = useLocalePath()
const router = useRouter()
const route = useRoute()

interface BreadcrumbItem {
  label: string
  to?: string
}

const matchNavItemForPath = (
  items: { to: string; labelKey: string }[],
): { label: string; to: string } | null => {
  const currentPath = route.path || ''

  for (const item of items) {
    const targetPath = localePath(item.to)
    if (
      currentPath === targetPath ||
      (currentPath.startsWith(targetPath) && currentPath[targetPath.length] === '/')
    ) {
      return {
        label: t(item.labelKey) as string,
        to: targetPath,
      }
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

  // Products hub: Home / Products
  const productsHub = localePath('/products')
  if (currentPath === productsHub) {
    items.push({ label: t('footer.menus.products', 'Products') as string })
    return items
  }

  // Guides is its own top-level section (All Guides hub)
  const guidesHub = localePath('/guides')
  if (currentPath === guidesHub) {
    items.push({ label: t('breadcrumbs.guides', 'Guides') as string })
    return items
  }

  // 单个 Guides 子页面：Home / Guides / {具体页面}
  if (currentPath.startsWith(guidesHub + '/')) {
    items.push({ label: t('breadcrumbs.guides', 'Guides') as string, to: guidesHub })

    // 根据具体路径映射更友好的标题
    if (currentPath === localePath('/guides/tools')) {
      items.push({ label: t('products.nav.aboutTools', 'About Tools') as string })
    } else if (currentPath === localePath('/guides/sizecharts')) {
      items.push({ label: t('products.nav.tireSizeCharts', 'Tire Size Charts') as string })
    } else if (currentPath === localePath('/guides/technical')) {
      items.push({ label: t('products.nav.technicalDocs', 'Technical') as string })
    } else if (currentPath === localePath('/guides/wheelset-buyers')) {
      items.push({ label: t('products.nav.wheelsetBuyersGuide', 'Wheelset Buyers Guide') as string })
    } else {
      // 其它 /guides/* 页面，使用最后一段路径作为标题占位
      const segments = currentPath.split('/').filter(Boolean)
      const last = segments[segments.length - 1] || ''
      items.push({ label: last })
    }

    return items
  }

  // Support hub: Home / Support
  const supportHub = localePath('/support')
  if (currentPath === supportHub) {
    items.push({ label: t('footer.menus.support', 'Support') as string })
    return items
  }

  // Company hub: Home / Company
  const companyHub = localePath('/company')
  if (currentPath === companyHub) {
    items.push({ label: t('footer.menus.company', 'Company') as string })
    return items
  }

  // /shop 作为独立页面：直接显示 Home / Shop
  const shopPath = localePath('/shop')
  if (currentPath === shopPath) {
    items.push({ label: t('products.nav.shop', 'Shop') as string })
    return items
  }

  // Wheelsbuild blog 作为独立页面：直接显示 Home / Wheelsbuild blog
  const wheelsbuildPath = localePath('/wheelsbuild')
  if (currentPath === wheelsbuildPath) {
    items.push({ label: t('products.nav.wheelsbuildBlog', 'Wheelsbuild blog') as string })
    return items
  }

  // Products layout 子页面：Home / Products / {具体页面}
  const productMatch = matchNavItemForPath(productsNavItems)
  if (productMatch) {
    items.push({
      label: t('footer.menus.products', 'Products') as string,
      to: localePath('/products'),
    })
    const last = items[items.length - 1]
    if (!last || productMatch.to !== last.to) {
      items.push({ label: productMatch.label })
    }
    return items
  }

  // Support layout 子页面：Home / Support / {具体页面}
  const supportMatch = matchNavItemForPath(supportNavItems)
  if (supportMatch) {
    items.push({
      label: t('footer.menus.support', 'Support') as string,
      to: localePath('/support'),
    })
    const last = items[items.length - 1]
    if (!last || supportMatch.to !== last.to) {
      items.push({ label: supportMatch.label })
    }
    return items
  }

  // Company layout 子页面：Home / Company / {具体页面}
  const companyMatch = matchNavItemForPath(companyNavItems)
  if (companyMatch) {
    items.push({
      label: t('footer.menus.company', 'Company') as string,
      to: localePath('/company'),
    })
    const last = items[items.length - 1]
    if (!last || companyMatch.to !== last.to) {
      items.push({ label: companyMatch.label })
    }
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
  return normalizedLocales.value.some(item => item.code === value)
}

const currentLocale = computed<LocaleOption>(() => {
  return (
    normalizedLocales.value.find(l => l.code === locale.value) ||
    normalizedLocales.value[0] ||
    { code: locale.value }
  )
})

const availableLocales = computed<LocaleOption[]>(() => {
  return normalizedLocales.value.filter(l => l.code !== locale.value)
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

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'language' } }))
  }
}

const onButtonKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    isOpen.value = !isOpen.value
    if (isOpen.value) {
      nextTick(() => optionRefs.value[0]?.focus())
    }
  } else if (e.key === 'Escape') {
    isOpen.value = false
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
    document.getElementById(buttonId)?.focus()
  }
}

const switchLanguage = async (code: string) => {
  try {
    if (!code || !isLocaleCode(code) || code === locale.value) { isOpen.value = false; return }
    locale.value = code
    await nextTick()
    try { await setLocale(code) } catch {}
    const targetPath = switchLocalePath(code as any)
    const current = router.currentRoute.value?.fullPath || ''
    if (targetPath && targetPath !== current) {
      try {
        await router.push(targetPath)
      } catch {
        window.location.assign(targetPath)
      }
    }
  } finally {
    isOpen.value = false
  }
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target
  if (!(target instanceof Element)) return
  if (!target.closest('[data-lang-wrapper]') && !target.closest('#' + dropdownId)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  const onGlobalPopup = (event: Event) => {
    try {
      const custom = event as CustomEvent<{ id?: string }>
      const id = custom?.detail?.id
      if (id !== 'language') {
        isOpen.value = false
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

/* tablet-820: 820x1180 等宽度段，限制 SiteHeader 高度为 130px */
@media (min-width: 820px) and (max-width: 1023px) {
	.site-header-root {
		max-height: 130px;
	}
}
</style>
