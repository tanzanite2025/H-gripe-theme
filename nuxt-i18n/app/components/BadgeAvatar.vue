<template>
  <div class="badge" :class="tierClass" :style="badgeStyle">
    <div class="badge-core">
      <div class="badge-inner">
        <span v-if="showStar" class="icon">★</span>
        <span v-else class="icon">?</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const props = defineProps({
  logged: { type: Boolean, default: false },
  level: { type: [String, Number], default: '' },
  topTierImageUrl: { type: String, default: '' }
})

const accent = ref('')

const tierClass = computed(() => {
  const lv = (props.level || '').toString().toLowerCase()
  if (!props.logged || !lv) return 'tier-guest'
  if (['vip','v1','level1','bronze'].includes(lv)) return 'tier-1'
  if (['v2','level2','silver'].includes(lv)) return 'tier-2'
  if (['v3','level3','gold'].includes(lv)) return 'tier-3'
  if (['v4','level4','diamond','top','pro','elite'].includes(lv)) return 'tier-top'
  return 'tier-1'
})

const showStar = computed(() => {
  return props.logged && tierClass.value !== 'tier-guest'
})

const badgeStyle = computed(() => {
  if (tierClass.value === 'tier-top' && props.topTierImageUrl) {
    return { 
      backgroundImage: `url(${props.topTierImageUrl})`, 
      backgroundSize: 'cover', 
      backgroundPosition: 'center' 
    }
  }
  if (tierClass.value === 'tier-top' && accent.value) {
    const c = accent.value
    return { background: `linear-gradient(180deg, ${c}, rgba(0,0,0,.4))` }
  }
  return {}
})

const pickFirstColor = (bg) => {
  // 粗略解析 linear-gradient 中的第一个 rgb/rgba 颜色
  const m = bg.match(/(rgba?\([^\)]+\))/)
  return m ? m[1] : ''
}

onMounted(() => {
  if (process.client) {
    try {
      // 1) 优先读取全局 CSS 变量
      const root = getComputedStyle(document.documentElement)
      const vars = ['--accent-color','--pricebar-accent','--brand-primary','--mytheme-accent']
      for (const v of vars) {
        const val = root.getPropertyValue(v).trim()
        if (val) { 
          accent.value = val
          break 
        }
      }
      if (accent.value) return
      
      // 2) 从底部价格重量块读取颜色
      const candidates = [
        '#price-weight-bar', '#cart-summary', '#checkout-summary',
        '.price-weight-bar', '.cart-summary', '.checkout-summary',
        '[data-role="price-bar"]', '[data-component="price-bar"]'
      ]
      let el = null
      for (const sel of candidates) { 
        el = document.querySelector(sel)
        if (el) break 
      }
      if (el) {
        const cs = getComputedStyle(el)
        const bg = cs.backgroundImage && cs.backgroundImage !== 'none' ? cs.backgroundImage : ''
        const bc = cs.borderColor && cs.borderColor !== 'rgba(0, 0, 0, 0)' ? cs.borderColor : ''
        const fc = cs.color
        accent.value = (bg && bg.includes('rgb')) ? pickFirstColor(bg) : (bc || fc || '')
      }
    } catch (_) { /* noop */ }
  }
})
</script>

<style scoped>
.badge { position: relative; width: 96px; height: 96px; border-radius: 9999px; display: inline-flex; align-items: center; justify-content: center; background: rgba(15,23,42,0.98); box-shadow: 0 4px 10px -4px rgba(0,0,0,0.95); }
.badge::before { content: none; }
.badge::after { content: ""; position: absolute; inset: 0; background-image: var(--badge-frame-url, none); background-repeat: no-repeat; background-position: center; background-size: contain; pointer-events: none; }
.badge-core { width: 86px; height: 86px; border-radius: 9999px; display: flex; align-items: center; justify-content: center; background: rgba(255,255,255,.06); }
</style>
