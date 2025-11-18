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
      <nav class="justify-self-center w-full max-w-[600px] h-10 rounded-full bg-black border border-black pointer-events-none" aria-label="Breadcrumbs">
        <div class="w-full h-full flex items-center justify-center px-4">
          <span class="text-white/60 font-semibold text-[13px]">Breadcrumbs</span>
        </div>
      </nav>
      
      <!-- 右侧：FAQ + 分享 + 语言切换器 -->
      <div class="justify-self-end flex items-center gap-3">
        <!-- FAQ 按钮 -->
        <button 
          class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-12 h-12 rounded-full inline-flex items-center justify-center bg-black border-2 border-[#6b73ff]" 
          @click.stop="toggleFaq()" 
          :aria-expanded="faqOpen" 
          aria-haspopup="dialog" 
          aria-label="Open FAQ"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="10"/>
            <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
        </button>
        
        <!-- 分享按钮（会员积分） - 改为圆形 -->
        <button 
          class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-12 h-12 rounded-full inline-flex items-center justify-center bg-black border-2 border-[#6b73ff]" 
          @click.stop="toggleShare()" 
          :aria-expanded="shareOpen" 
          aria-haspopup="dialog" 
          aria-label="Open membership panel"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 16 16" aria-hidden="true" class="fill-current"><g fill="none" stroke="#ed8796" stroke-linecap="round" stroke-linejoin="round" stroke-width="1"><path d="m8 12.5l4.5-5l-2-2h-5l-2 2z"/><path d="M14.5 12L8 15.5L1.5 12V4L8 .5L14.5 4z"/></g></svg>
        </button>
        
        <!-- 翻译转换器 -->
        <div class="relative">
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
    
    <!-- 移动端：三排垂直布局 -->
    <div class="md:hidden grid gap-0.5 justify-items-center">
      <!-- 第一排：站点标题 -->
      <div class="flex justify-center items-center">
        <h1 class="m-0 text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] [font-family:'AerialFaster',sans-serif] tracking-wide drop-shadow-[0_2px_8px_rgba(64,255,170,0.3)]">
          {{ titleText }}
        </h1>
      </div>
      
      <!-- 第二排：FAQ + 分享 + 翻译转换器 -->
      <div class="flex justify-center items-center gap-2">
        <!-- FAQ 按钮 -->
        <button 
          class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-[37px] h-[37px] rounded-full inline-flex items-center justify-center bg-black border-2 border-[#6b73ff]" 
          @click.stop="toggleFaq()" 
          :aria-expanded="faqOpen" 
          aria-haspopup="dialog" 
          aria-label="Open FAQ"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="10"/>
            <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
        </button>
        
        <!-- 分享按钮 - 改为圆形 -->
        <button 
          class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-[37px] h-[37px] rounded-full inline-flex items-center justify-center bg-black border-2 border-[#6b73ff]" 
          @click.stop="toggleShare()" 
          :aria-expanded="shareOpen" 
          aria-haspopup="dialog" 
          aria-label="Open membership panel"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 16 16" aria-hidden="true" class="fill-current"><g fill="none" stroke="#ed8796" stroke-linecap="round" stroke-linejoin="round" stroke-width="1"><path d="m8 12.5l4.5-5l-2-2h-5l-2 2z"/><path d="M14.5 12L8 15.5L1.5 12V4L8 .5L14.5 4z"/></g></svg>
        </button>
        
        <!-- 翻译转换器 -->
        <div class="relative min-w-[140px]">
        <button 
          class="flex items-center justify-between gap-3 px-4 py-2.5 rounded-full text-white text-sm font-medium cursor-pointer transition-all duration-200 w-[125px] h-[37px] shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] bg-black border-2 border-[#6b73ff]" 
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
      <nav class="w-[85vw] max-w-[600px] h-[30px] rounded-[30px] bg-black border border-black pointer-events-none" aria-label="Breadcrumbs">
        <div class="w-full h-full flex items-center justify-center px-3">
          <span class="text-white/60 font-semibold text-[13px]">Breadcrumbs</span>
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
      <div v-if="faqOpen" class="fixed inset-0 z-[9999] flex items-center justify-center p-4" @click.self="faqOpen = false">
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
      <div v-if="shareOpen" class="fixed inset-0 z-[9999] flex items-center justify-center" @click.self="shareOpen = false">
        <!-- 不透明背景遮罩 -->
        <div class="absolute inset-0 bg-black"></div>
        <!-- 弹窗内容 -->
        <div class="relative w-[min(95vw,1650px)] max-h-[90vh] overflow-auto" aria-modal="true" role="dialog" aria-label="Membership">
          <LeverAndPoint @close="shareOpen = false" />
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useSiteTitle } from '~/composables/useSiteTitle'
import LeverAndPoint from '~/components/LeverAndPoint.vue'
import FaqModal from '~/components/FaqModal.vue'
import WhatsAppChatModal from '~/components/WhatsAppChatModal.vue'

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
const switchLocalePath = useSwitchLocalePath()
const router = useRouter()

const isOpen = ref(false)

const currentLocale = computed(() => locales.value.find(l => l.code === locale.value) || locales.value[0])
const availableLocales = computed(() => locales.value.filter(l => l.code !== locale.value))

const buttonId = 'lang-switcher-button'
const dropdownId = 'lang-switcher-dropdown'

const optionRefs = ref([])
const setOptionRef = (el, index) => {
  if (!el) return
  optionRefs.value[index] = el
}

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'language' } }))
  }
}

const onButtonKeydown = (e) => {
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

const onListKeydown = (e) => {
  const refs = optionRefs.value
  if (!Array.isArray(refs) || !refs.length) return
  const idx = refs.findIndex(el => el === document.activeElement)
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    const n = refs[(idx + 1 + refs.length) % refs.length]
    n?.focus()
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    const n = refs[(idx - 1 + refs.length) % refs.length]
    n?.focus()
  } else if (e.key === 'Escape') {
    isOpen.value = false
    document.getElementById(buttonId)?.focus()
  }
}

const switchLanguage = async (code) => {
  try {
    if (!code || code === locale.value) { isOpen.value = false; return }
    locale.value = code
    await nextTick()
    try { await setLocale(code) } catch {}
    const targetPath = switchLocalePath(code)
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

const handleClickOutside = (event) => {
  if (!(event.target instanceof Element)) return
  if (!event.target.closest('.relative.min-w-\\[140px\\]') && !event.target.closest('#' + dropdownId)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  const onGlobalPopup = (e) => {
    try {
      const id = e?.detail?.id
      if (id !== 'language') {
        isOpen.value = false
      }
    } catch {}
  }
  window.addEventListener('ui:popup-open', onGlobalPopup)
  onBeforeUnmount(() => {
    window.removeEventListener('ui:popup-open', onGlobalPopup)
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

const flagFilenameFromISO = (entry) => {
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

const flagSrc = (entry) => {
  const file = flagFilenameFromISO(entry)
  if (!file) return ''
  return `/twemoji/svg/${file}`
}
</script>
