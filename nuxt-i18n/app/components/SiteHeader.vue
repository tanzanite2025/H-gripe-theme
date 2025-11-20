<template>
  <div class="fixed top-1.5 left-1/2 -translate-x-1/2 w-[95vw] max-w-[1200px] z-[110]">
    <!-- 桌面端：一排三元素 -->
    <div class="hidden md:grid grid-cols-[250px_1fr_250px] items-center gap-4">
      <!-- 左侧：站点标题 -->
      <div class="flex justify-start items-center">
        <h1 class="m-0 text-5xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)] whitespace-nowrap">
          {{ titleText }}
        </h1>
      </div>
      
      <!-- 中间：面包屑导航 -->
      <nav class="justify-self-center w-full max-w-[600px] h-10 rounded-full bg-transparent border border-transparent pointer-events-none" aria-label="Breadcrumbs">
        <div class="w-full h-full flex items-center justify-center px-4">
          <span class="text-white font-semibold text-[13px]">Breadcrumbs</span>
        </div>
      </nav>
      
      <!-- 右侧：FAQ + 分享 + 语言切换器 -->
      <div class="justify-self-end flex items-center gap-3">
        <!-- FAQ 按钮 -->
        <button 
          class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-[52px] h-[52px] rounded-full inline-flex items-center justify-center bg-[#0b1020]" 
          @click.stop="toggleFaq()" 
          :aria-expanded="faqOpen" 
          aria-haspopup="dialog" 
          aria-label="Open FAQ"
        >
          <img src="/icons/token-branded--ionx.svg" alt="" class="w-full h-full" />
        </button>
        
        <!-- 分享按钮（会员积分） - 改为圆形 -->
        <button 
          class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-[52px] h-[52px] rounded-full inline-flex items-center justify-center bg-[#0b1020]" 
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
          class="flex items-center justify-between gap-3 px-4 py-2.5 rounded-full text-white text-sm font-medium cursor-pointer transition-all duration-200 w-[125px] h-12 shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] bg-black border-2 border-[#6b73ff]" 
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
    
    <!-- 移动端：三排垂直布局（紧凑间距） -->
    <div class="md:hidden grid gap-0.5 justify-items-center">
      <!-- 第一排：站点标题 + FAQ + 分享 -->
      <div class="w-[90vw] max-w-[600px] flex items-center justify-between">
        <!-- FAQ 按钮 -->
        <button 
          class="pointer-events-auto text-white transition-all duration-200 w-[52px] h-[52px] rounded-full inline-flex items-center justify-center bg-[#0b1020]" 
          @click.stop="toggleFaq()" 
          :aria-expanded="faqOpen" 
          aria-haspopup="dialog" 
          aria-label="Open FAQ"
        >
          <img src="/icons/token-branded--ionx.svg" alt="" class="w-full h-full" />
        </button>

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

      <!-- 第二排：翻译转换器（单独一排，居中，稍微压缩高度） -->
      <div class="flex justify-center items-center">
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
      
      <!-- 第三排：面包屑导航 -->
      <nav class="w-[85vw] max-w-[600px] h-[30px] rounded-[30px] bg-transparent border border-transparent pointer-events-none" aria-label="Breadcrumbs">
        <div class="w-full h-full flex items-center justify-center px-3">
          <span class="text-white font-semibold text-[13px]">Breadcrumbs</span>
        </div>
      </nav>
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
  <WhatsAppChatModal v-if="whatsappOpen" :conversation="{ showAgentList: true }" @close="whatsappOpen = false" />
  
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
        class="fixed inset-0 z-[9999] flex items-center justify-center p-0 md:p-4"
        @click.self="shareOpen = false"
      >
        <!-- 不透明背景遮罩 -->
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>
        <!-- 弹窗内容 -->
        <div class="relative w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] flex" aria-modal="true" role="dialog" aria-label="Membership">
          <LeverAndPoint @close="shareOpen = false" />
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch, type ComponentPublicInstance } from 'vue'
import { useLocalePath } from '#imports'
import { useSiteTitle } from '~/composables/useSiteTitle'
import LeverAndPoint from '~/components/LeverAndPoint.vue'
import FaqModal from '~/components/FaqModal.vue'
import WhatsAppChatModal from '~/components/WhatsAppChatModal.vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'

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
const { locale, locales, setLocale } = useI18n()
const localePath = useLocalePath()
const router = useRouter()

const switchLocalePath = (targetLocale: string) => {
  const currentFullPath = router.currentRoute.value?.fullPath || '/'
  return localePath({ path: currentFullPath }, targetLocale)
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
    const targetPath = switchLocalePath(code as Parameters<typeof switchLocalePath>[0])
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
