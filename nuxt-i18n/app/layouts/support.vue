<template>
  <div class="layout layout--support">
    <main class="layout-main">
      <div class="support-header-spacer" aria-hidden="true"></div>
      <PrimarySectionTabBar />

      <!-- Support page content -->
      <section class="support-content">
        <div class="support-content__inner page-content-shell">
          <slot />
        </div>
      </section>
    </main>

    <AppFooter />
    <GradientDockMenu />
  </div>
</template>

<script setup lang="ts">
import AppFooter from '~/components/AppFooter.vue'
import GradientDockMenu from '~/components/GradientDockMenu.vue'
import PrimarySectionTabBar from '~/components/PrimarySectionTabBar.vue'
</script>

<style scoped>
.layout--support {
  --primary-section-tab-bar-height: 58px;
  --page-tab-bar-flow-offset: -1.5rem;
  --page-tab-bar-sticky-top: calc(var(--site-header-offset, 120px) + var(--primary-section-tab-bar-height, 58px));
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: radial-gradient(circle at top, #020617 0, #020617 24%, #000000 100%);
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.support-header-spacer {
  height: var(--site-header-offset, 145px);
}

.support-hero {
  margin-top: 0;
  /* 桌面端：略微压缩底部 padding，让下方 Support 导航更靠上 */
  padding: 1.5rem 1.5rem 0.75rem;
  /* 去掉单独的 hero 渐变背景，直接使用整体布局背景，避免形成一条额外的色带 */
  background: transparent !important;
}

.support-hero__inner {
  max-width: 960px;
  margin: 0 auto;
}

.support-hero__title {
  margin: 0 0 0.5rem;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #f9fafb;
  display: none !important;
}

.support-hero__subtitle {
  margin: 0;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-content {
  padding: 2rem 0 3rem;
}

.support-content__inner {
  max-width: none;
}

@media (max-width: 768px) {
  .support-hero {
    /* 顶部固定 SiteHeader 由全局 spacer 处理，这里只负责内容与导航的间距 */
    margin-top: 0;
    padding: 1.25rem 1.25rem 0.75rem;
  }

  .support-hero__title {
    font-size: 1.5rem;
    /* 保持隐藏：移动端同样不展示 Support 标题行，避免占用垂直空间 */
    display: none;
  }

  .support-content {
    padding-inline: 0;
  }

  .support-content__inner {
    max-width: none;
  }
}

@media (min-width: 768px) {
  .support-header-spacer {
    height: var(--site-header-offset, 112px);
  }
}

@media (max-width: 900px) {
  .layout--support {
    --primary-section-tab-bar-height: 56px;
  }
}
</style>
