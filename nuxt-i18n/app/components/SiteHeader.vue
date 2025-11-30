<template>
	<div class="fixed top-1.5 left-1/2 -translate-x-1/2 w-[95vw] max-w-[1200px] z-[110]">
		<div
			class="relative w-full rounded-[30px] bg-[#0b1020]/70 backdrop-blur-md border border-white/10 shadow-[0_18px_45px_rgba(15,23,42,0.9)] px-4 py-0.5 md:py-2"
		>
			<!-- 桌面端：第一行 标题 + 主导航 + 右侧控件，第二行 面包屑 -->
			<div class="hidden md:flex flex-col gap-1">
				<!-- 第一行：标题 + 主导航 + 右侧控件 -->
				<div
					class="grid grid-cols-[230px_1fr_240px] items-center gap-3 lg:grid-cols-[250px_1fr_250px] lg:gap-4"
				>
					<!-- 左侧：站点标题 -->
					<div class="flex justify-start items-center">
						<h1 class="m-0 text-4xl lg:text-5xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)] whitespace-nowrap">
							{{ titleText }}
						</h1>
					</div>

					<!-- 中间：主导航（与站点标题同一行） -->
					<nav
						class="justify-self-center flex items-center justify-center gap-3 lg:gap-4"
						aria-label="Primary navigation"
					>
						<NuxtLink
							:to="localePath('/products')"
							class="px-3 py-1 lg:px-4 lg:py-1.5 rounded-full border border-white/15 bg-white/5 text-[12px] lg:text-[13px] font-medium text-white/80 hover:text-white hover:bg-white/10 hover:border-white/30 transition-colors"
						>
							{{ $t('footer.menus.products', 'Products') }}
						</NuxtLink>
						<NuxtLink
							:to="localePath('/support')"
							class="px-3 py-1 lg:px-4 lg:py-1.5 rounded-full border border-white/15 bg-white/5 text-[12px] lg:text-[13px] font-medium text-white/80 hover:text-white hover:bg-white/10 hover:border-white/30 transition-colors"
						>
							{{ $t('footer.menus.support', 'Support') }}
						</NuxtLink>
						<NuxtLink
							:to="localePath('/company')"
							class="px-3 py-1 lg:px-4 lg:py-1.5 rounded-full border border-white/15 bg-white/5 text-[12px] lg:text-[13px] font-medium text-white/80 hover:text-white hover:bg-white/10 hover:border-white/30 transition-colors"
						>
							{{ $t('footer.menus.company', 'Company') }}
						</NuxtLink>
					</nav>

					<!-- 右侧：All Guides + 分享 + 语言切换器 -->
					<div
						class="justify-self-end flex items-center gap-2 lg:gap-3"
					>
						<!-- All Guides 文本按钮（桌面端） -->
						<NuxtLink
							:to="localePath('/guides')"
							class="pointer-events-auto text-slate-950 shadow-[0_0_16px_rgba(64,115,255,0.65)] hover:shadow-[0_0_22px_rgba(64,115,255,0.9)] transition-all duration-200 h-10 lg:h-12 px-3.5 lg:px-4 rounded-full hidden lg:inline-flex items-center justify-center bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-[11px] lg:text-[12px] font-semibold hover:-translate-y-[1px] hover:scale-[1.02]"
							aria-label="All Guides"
						>
							All Guides
						</NuxtLink>

						<!-- 分享按钮（会员积分） - 改为圆形 -->
						<button
							class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-10 h-10 lg:w-12 lg:h-12 rounded-full hidden lg:inline-flex items-center justify-center bg-[#0b1020]"
							@click.stop="toggleShare()"
							:aria-expanded="shareOpen"
							aria-haspopup="dialog"
							aria-label="Open membership panel"
						>
							<img src="/icons/token-branded--looks.svg" alt="" class="w-full h-full" />
						</button>

						<!-- 翻译转换器 -->
						<div class="relative" data-lang-wrapper>
							<button
								class="flex items-center justify-between gap-2 lg:gap-3 px-3 py-2 lg:px-4 lg:py-2.5 rounded-full text-white text-xs lg:text-sm font-medium cursor-pointer transition-all duration-200 w-[110px] h-10 lg:w-[125px] lg:h-12 shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] bg-black border-2 border-[#6b73ff]"
								@click.stop="toggleDropdown"
								@keydown="onButtonKeydown"
								:id="buttonId"
								aria-haspopup="listbox"
								:aria-expanded="isOpen"
								:aria-controls="dropdownId"
								:aria-label="'Switch language'"
							>
								<span class="font-medium flex items-center gap-2">
									<span class="w-[1.2em] inline-block" aria-hidden="true">
										<img :src="flagSrc(currentLocale)" alt="" class="w-[1.2em] h-[1.2em] block" />
									</span>
									{{ currentLocale.name }}
								</span>
								<span class="text-[10px] transition-transform duration-200" :class="{ 'rotate-180': isOpen }">▼</span>
							</button>
						</div>

						<teleport to="body">
							<transition
								enter-active-class="transition-all duration-200 ease-in-out"
								leave-active-class="transition-all duration-200 ease-in-out"
								enter-from-class="opacity-0 -translate-y-2.5"
								leave-to-class="opacity-0 -translate-y-2.5"
							>
								<div
									v-if="isOpen"
									class="fixed top-[100px] max-md:top-[83px] left-1/2 -translate-x-1/2 w-[90vw] max-md:w-[70vw] max-w-[1600px] bg-[#0b1020] border border-[#6b79ff] rounded-md overflow-auto [-webkit-overflow-scrolling:touch] [overscroll-behavior:contain] [touch-action:pan-y] max-h-[70vh] max-md:max-h-[45vh] shadow-none grid grid-cols-[repeat(auto-fit,minmax(160px,1fr))] max-md:grid-cols-2 gap-1.5 justify-items-center z-[1200]"
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
				</div>

				<!-- 桌面端：第二行 面包屑 -->
				<div class="w-full flex items-center">
					<div class="hidden md:flex lg:hidden w-[44px] justify-start mr-2">
						<NuxtLink
							:to="localePath('/guides')"
							class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-[44px] h-[44px] rounded-full inline-flex items-center justify-center bg-[#0b1020]"
							aria-label="All Guides"
						>
							<img src="/icons/token-branded--ionx.svg" alt="" class="w-full h-full" />
						</NuxtLink>
					</div>
					<div class="flex-1 flex justify-center">
						<nav
							v-if="breadcrumbs.length"
							aria-label="Breadcrumb"
							class="text-[12px] leading-tight text-slate-300"
						>
							<ol class="flex items-center gap-1.5">
								<li
									v-for="(crumb, index) in breadcrumbs"
									:key="index"
									class="flex items-center gap-1"
								>
									<NuxtLink
										v-if="crumb.to && index < breadcrumbs.length - 1"
										:to="crumb.to"
										class="hover:text-white"
									>
										{{ crumb.label }}
									</NuxtLink>
									<span v-else class="text-white font-semibold">
										{{ crumb.label }}
									</span>
									<span v-if="index < breadcrumbs.length - 1">/</span>
								</li>
							</ol>
						</nav>
					</div>
					<div class="hidden md:flex lg:hidden w-[44px] justify-end ml-2">
						<button
							class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-[44px] h-[44px] rounded-full inline-flex items-center justify-center bg-[#0b1020]"
							@click.stop="toggleShare()"
							:aria-expanded="shareOpen"
							aria-haspopup="dialog"
							aria-label="Open membership panel"
						>
							<img src="/icons/token-branded--looks.svg" alt="" class="w-full h-full" />
						</button>
					</div>
				</div>
			</div>

			<!-- 移动端：三排垂直布局（更紧凑的间距） -->
			<div class="md:hidden grid gap-0 justify-items-center">
				<!-- 第一排：站点标题 + All Guides + 分享（整体向上贴近容器顶部，414 宽单独微调） -->
				<div class="w-[90vw] max-w-[600px] flex items-center justify-between -mt-1 phone-414:mt-0">
					<!-- All Guides 文本按钮（移动端） -->
					<NuxtLink
						:to="localePath('/guides')"
						class="pointer-events-auto text-slate-950 transition-all duration-200 h-[52px] px-3.5 rounded-full inline-flex items-center justify-center bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-[11px] font-semibold shadow-[0_0_14px_rgba(64,115,255,0.65)] hover:shadow-[0_0_20px_rgba(64,115,255,0.9)] hover:-translate-y-[1px]"
						aria-label="All Guides"
					>
						All Guides
					</NuxtLink>

					<!-- 站点标题（压缩上下留白） -->
					<div class="flex-1 flex justify-center px-2 py-0">
						<h1 class="m-0 text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)] leading-none text-center">
							{{ titleText }}
						</h1>
					</div>

					<!-- 分享按钮 - 圆形 -->
					<button
						class="pointer-events-auto text-white transition-all duration-200 w-[52px] h-[52px] rounded-full inline-flex items-center justify-center bg-[#0b1020]"
						@click.stop="toggleShare()"
						:aria-expanded="shareOpen"
						aria-haspopup="dialog"
						aria-label="Open membership panel"
					>
						<img src="/icons/token-branded--looks.svg" alt="" class="w-full h-full" />
					</button>
				</div>

				<!-- 第二排：翻译转换器（单独一排，居中，414 宽下与上下行间距更均衡） -->
				<div class="flex justify-center items-center -mt-1.5 phone-414:-mt-1">
					<div class="relative min-w-[150px]" data-lang-wrapper>
						<button
							class="flex items-center justify-between gap-3 px-4 py-1.5 rounded-full text-white text-sm font-medium cursor-pointer transition-all duration-200 w-[150px] h-[36px] shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] bg-black border-2 border-[#6b73ff]"
							@click.stop="toggleDropdown"
							@keydown="onButtonKeydown"
							:id="buttonId"
							aria-haspopup="listbox"
							:aria-expanded="isOpen"
							:aria-controls="dropdownId"
							:aria-label="'Switch language'"
						>
							<span class="font-medium flex items-center gap-2">
								<span class="w-[1.2em] inline-block" aria-hidden="true">
									<img :src="flagSrc(currentLocale)" alt="" class="w-[1.2em] h-[1.2em] block" />
								</span>
								{{ currentLocale.name }}
							</span>
							<span class="text-[10px] transition-transform duration-200" :class="{ 'rotate-180': isOpen }">▼</span>
						</button>
					</div>
				</div>

				<!-- 第三排：主导航 + 面包屑导航（414 宽下略微下移） -->
				<nav
					class="w-[85vw] max-w-[600px] rounded-[30px] bg-transparent border border-transparent pointer-events-auto mt-0 phone-414:mt-1"
					aria-label="Primary navigation and breadcrumbs"
				>
					<div
						class="w-full flex flex-col items-center justify-center px-3 py-0.5 gap-1
							phone-375:py-0.5 phone-390:py-0.5 phone-414:py-0.5 phone-430:py-1
							tablet-768:py-1 tablet-820:py-1.5"
					>
						<!-- 第一行：主导航（Products / Support / Company） -->
						<div
							class="w-full flex items-center justify-between gap-2 text-[10px]
								phone-375:text-[10px]
								phone-390:text-[10px]
								phone-414:text-[10px]
								phone-430:text-[11px]
								tablet-768:text-[11px]
								tablet-820:text-[12px]"
						>
							<NuxtLink
								:to="localePath('/products')"
								class="flex-1 text-center px-3 py-1 rounded-full border border-white/15 bg-white/5 text-white/80 hover:text-white hover:bg-white/10 hover:border-white/30 transition-colors"
							>
								{{ $t('footer.menus.products', 'Products') }}
							</NuxtLink>
							<NuxtLink
								:to="localePath('/support')"
								class="flex-1 text-center px-3 py-1 rounded-full border border-white/15 bg-white/5 text-white/80 hover:text-white hover:bg-white/10 hover:border-white/30 transition-colors"
							>
								{{ $t('footer.menus.support', 'Support') }}
							</NuxtLink>
							<NuxtLink
								:to="localePath('/company')"
								class="flex-1 text-center px-3 py-1 rounded-full border border-white/15 bg-white/5 text-white/80 hover:text-white hover:bg-white/10 hover:border-white/30 transition-colors"
							>
								{{ $t('footer.menus.company', 'Company') }}
							</NuxtLink>
						</div>

						<!-- 第二行：面包屑 -->
						<nav
							v-if="breadcrumbs.length"
							aria-label="Breadcrumb"
							class="text-[11px] text-slate-300"
						>
							<ol class="flex items-center gap-1.5">
								<li
									v-for="(crumb, index) in breadcrumbs"
									:key="index"
									class="flex items-center gap-1"
								>
									<NuxtLink
										v-if="crumb.to && index < breadcrumbs.length - 1"
										:to="crumb.to"
										class="hover:text-white"
									>
										{{ crumb.label }}
									</NuxtLink>
									<span v-else class="text-white font-semibold">
										{{ crumb.label }}
									</span>
									<span v-if="index < breadcrumbs.length - 1">/</span>
								</li>
							</ol>
						</nav>
					</div>
				</nav>
			</div>
		</div>
	</div>

	<!-- FAQ 弹窗 -->
	<teleport to="body">
		<transition
			enter-active-class="transition-opacity duration-300 ease-out"
			leave-active-class="transition-opacity duration-200 ease-in"
			enter-from-class="opacity-0"
			leave-to-class="opacity-0"
		>
			<div
				v-if="faqOpen"
				class="fixed inset-0 z-[9999] flex items-center justify-center p-0 md:p-4"
				@click.self="faqOpen = false"
			>
				<!-- 半透明背景遮罩 -->
				<div class="absolute inset-0 bg-black/80 backdrop-blur-sm -z-10"></div>
				<!-- 弹窗内容 -->
				<FaqModal @close="faqOpen = false" @openWhatsApp="handleOpenWhatsApp" class="relative z-10" />
			</div>
		</transition>
	</teleport>

	<!-- WhatsApp Chat 弹窗 -->
	<WhatsAppChatModal
		v-if="whatsappOpen"
		:conversation="{ showAgentList: true }"
		@close="whatsappOpen = false"
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
import FaqModal from '~/components/FaqModal.vue'
import WhatsAppChatModal from '~/components/WhatsAppChatModal.vue'
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

// FAQ button
const faqOpen = ref(false)

const toggleFaq = () => {
  faqOpen.value = !faqOpen.value
  if (faqOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'header-faq' } }))
  }
}

const SIDEBAR_TOKEN_HEADER_FAQ = 'header-faq'

watch(faqOpen, (open) => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_HEADER_FAQ, open)
}, { immediate: true })

onBeforeUnmount(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_HEADER_FAQ, false)
})

// WhatsApp Chat Modal
const whatsappOpen = ref(false)

const handleOpenWhatsApp = () => {
  // 先关闭 FAQ 弹窗
  faqOpen.value = false
  // 打开 WhatsApp 弹窗
  whatsappOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}

// Share button (Membership panel)
const shareOpen = ref(false)

const toggleShare = () => {
  shareOpen.value = !shareOpen.value
  if (shareOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'header-share' } }))
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
