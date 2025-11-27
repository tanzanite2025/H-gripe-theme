<template>
  <div class="layout layout--support">
    <main class="layout-main">
      <!-- Support hero / banner -->
      <section class="support-hero">
        <div class="support-hero__inner">
          <h1 class="support-hero__title">Support</h1>
        </div>
      </section>

      <!-- Top horizontal Support navigation -->
      <SupportTopNav />

      <!-- Support page content -->
      <section class="support-content">
        <div class="support-content__inner">
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
import SupportTopNav from '~/components/SupportTopNav.vue'
</script>

<style scoped>
.layout--support {
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

.support-hero {
  margin-top: 2.75rem;
  /* 桌面端：略微压缩底部 padding，让下方 Support 导航更靠上 */
  padding: 3rem 1.5rem 0.75rem;
  /* 去掉单独的 hero 渐变背景，直接使用整体布局背景，避免形成一条额外的色带 */
  background: transparent;
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
  display: none;
}

.support-hero__subtitle {
  margin: 0;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-content {
  padding: 2rem 1.5rem 3rem;
}

.support-content__inner {
  max-width: 960px;
  margin: 0 auto;
}

@media (max-width: 768px) {
  .support-hero {
    /* 为顶部固定 SiteHeader 预留更多空间，避免 Support 标题和顶部导航被遮挡 */
    margin-top: 4.5rem;
    padding: 2rem 1.25rem 0.75rem;
  }

  .support-hero__title {
    font-size: 1.5rem;
    /* 保持隐藏：移动端同样不展示 Support 标题行，避免占用垂直空间 */
    display: none;
  }

  .support-content {
    padding-inline: 1.25rem;
  }
}
</style>
