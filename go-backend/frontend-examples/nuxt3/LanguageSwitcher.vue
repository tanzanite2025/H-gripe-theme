<template>
  <div class="language-switcher">
    <!-- 下拉选择器样式 -->
    <select 
      v-if="displayMode === 'select'"
      v-model="currentLocale" 
      @change="handleLanguageChange"
      class="language-select"
    >
      <option v-for="lang in languages" :key="lang.code" :value="lang.code">
        {{ lang.native_name }}
      </option>
    </select>

    <!-- 按钮列表样式 -->
    <div v-else-if="displayMode === 'buttons'" class="language-buttons">
      <button
        v-for="lang in languages"
        :key="lang.code"
        :class="['language-button', { active: currentLocale === lang.code }]"
        @click="switchTo(lang.code)"
      >
        {{ lang.code.toUpperCase() }}
      </button>
    </div>

    <!-- 下拉菜单样式 -->
    <div v-else-if="displayMode === 'dropdown'" class="language-dropdown">
      <button @click="isOpen = !isOpen" class="dropdown-trigger">
        <span>{{ currentLanguageName }}</span>
        <svg 
          class="dropdown-icon" 
          :class="{ 'rotate-180': isOpen }"
          width="16" 
          height="16" 
          viewBox="0 0 16 16"
        >
          <path d="M4 6l4 4 4-4" stroke="currentColor" stroke-width="2" fill="none"/>
        </svg>
      </button>
      
      <transition name="dropdown">
        <ul v-if="isOpen" class="dropdown-menu">
          <li 
            v-for="lang in languages" 
            :key="lang.code"
            :class="{ active: currentLocale === lang.code }"
          >
            <button @click="switchTo(lang.code)">
              {{ lang.native_name }}
            </button>
          </li>
        </ul>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Language } from './useI18n'

interface Props {
  displayMode?: 'select' | 'buttons' | 'dropdown'
  showAllLanguages?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  displayMode: 'select',
  showAllLanguages: false
})

const { locale, getLanguages, switchLanguage, getLanguageName } = useI18n()

const languages = ref<Language[]>([])
const currentLocale = ref(locale.value)
const currentLanguageName = ref('')
const isOpen = ref(false)

// 加载语言列表
onMounted(async () => {
  const allLanguages = await getLanguages()
  
  // 根据配置决定显示哪些语言
  if (props.showAllLanguages) {
    languages.value = allLanguages
  } else {
    // 只显示启用的语言
    languages.value = allLanguages.filter(lang => lang.enabled)
  }
  
  // 获取当前语言名称
  currentLanguageName.value = await getLanguageName(currentLocale.value)
})

// 监听语言变化
watch(locale, async (newLocale) => {
  currentLocale.value = newLocale
  currentLanguageName.value = await getLanguageName(newLocale)
})

// 处理语言切换（select 模式）
const handleLanguageChange = async () => {
  await switchLanguage(currentLocale.value)
}

// 切换到指定语言（buttons/dropdown 模式）
const switchTo = async (newLocale: string) => {
  isOpen.value = false
  currentLocale.value = newLocale
  await switchLanguage(newLocale)
}

// 点击外部关闭下拉菜单
onClickOutside(
  () => document.querySelector('.language-dropdown'),
  () => { isOpen.value = false }
)
</script>

<style scoped>
.language-switcher {
  position: relative;
}

/* Select 样式 */
.language-select {
  padding: 0.5rem 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  background-color: white;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.language-select:hover {
  border-color: #3b82f6;
}

.language-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* Buttons 样式 */
.language-buttons {
  display: flex;
  gap: 0.5rem;
}

.language-button {
  padding: 0.5rem 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  background-color: white;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
}

.language-button:hover {
  background-color: #f3f4f6;
  border-color: #3b82f6;
}

.language-button.active {
  background-color: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

/* Dropdown 样式 */
.language-dropdown {
  position: relative;
}

.dropdown-trigger {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  background-color: white;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.dropdown-trigger:hover {
  border-color: #3b82f6;
}

.dropdown-icon {
  transition: transform 0.2s;
}

.dropdown-icon.rotate-180 {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  min-width: 200px;
  background-color: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  list-style: none;
  padding: 0.5rem 0;
  margin: 0;
  z-index: 50;
}

.dropdown-menu li {
  margin: 0;
}

.dropdown-menu button {
  width: 100%;
  padding: 0.5rem 1rem;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.dropdown-menu button:hover {
  background-color: #f3f4f6;
}

.dropdown-menu li.active button {
  background-color: #eff6ff;
  color: #3b82f6;
  font-weight: 500;
}

/* 下拉动画 */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
