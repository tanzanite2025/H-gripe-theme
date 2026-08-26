<template>
  <section class="footer-site-overview" aria-labelledby="footer-site-overview-title">
    <h2 id="footer-site-overview-title" class="footer-site-overview__title">
      {{ t('footer.menus.siteOverview', 'Site Overview') }}
    </h2>

    <div v-if="canRenderSiteSettings" class="footer-site-overview__content">
      <p v-if="siteName" class="footer-site-overview__name">
        {{ siteName }}
      </p>
      <p v-if="siteDescription" class="footer-site-overview__paragraph">
        {{ siteDescription }}
      </p>
      <p v-if="contactEmail" class="footer-site-overview__paragraph">
        <a
          :href="`mailto:${contactEmail}`"
          class="footer-site-overview__link"
        >
          {{ contactEmail }}
        </a>
      </p>
      <p v-if="contactPhone" class="footer-site-overview__paragraph">
        <a
          :href="`tel:${contactPhone}`"
          class="footer-site-overview__link"
        >
          {{ contactPhone }}
        </a>
      </p>
    </div>

    <ul v-if="links.length" class="footer-site-overview__links">
      <li v-for="link in links" :key="link.to" class="footer-site-overview__link-item">
        <NuxtLink
          v-if="!link.external"
          class="footer-site-overview__link footer-site-overview__route-link"
          :to="localePath(link.to)"
        >
          {{ link.labelKey ? t(link.labelKey, link.fallback || link.labelKey) : link.fallback }}
          <Icon
            name="lucide:arrow-up-right"
            class="footer-site-overview__link-arrow"
            aria-hidden="true"
          />
        </NuxtLink>
        <a
          v-else
          class="footer-site-overview__link footer-site-overview__route-link"
          :href="link.to"
          target="_blank"
          rel="noopener noreferrer"
        >
          {{ link.labelKey ? t(link.labelKey, link.fallback || link.labelKey) : link.fallback }}
          <Icon
            name="lucide:arrow-up-right"
            class="footer-site-overview__link-arrow"
            aria-hidden="true"
          />
        </a>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useSiteSettings } from '~/composables/usePublicSettings'
import type { FooterLink } from '~/utils/footerMenus'

const props = defineProps<{
  links?: FooterLink[]
}>()

const { t } = useI18n()
const localePath = useLocalePath()
const { siteSettings } = useSiteSettings()
const links = computed(() => props.links || [])

const siteName = computed(() => siteSettings.value.siteTitle?.trim() || '')
const siteDescription = computed(() => siteSettings.value.siteDescription?.trim() || '')
const contactEmail = computed(() => siteSettings.value.contactEmail?.trim() || '')
const contactPhone = computed(() => siteSettings.value.contactPhone?.trim() || '')
const canRenderSiteSettings = ref(false)

onMounted(() => {
  canRenderSiteSettings.value = true
})
</script>

<style scoped>
.footer-site-overview {
  min-width: 0;
  text-align: left;
}

.footer-site-overview__title {
  margin: 0 0 1.25rem;
  font-size: 0.85rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  line-height: 1.3;
  text-transform: uppercase;
  color: var(--tz-text-primary);
}

.footer-site-overview__content {
  min-width: 0;
  font-size: 0.85rem;
  line-height: 1.7;
  color: var(--tz-text-secondary);
}

.footer-site-overview__name {
  margin: 0 0 0.85rem;
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.45;
  color: var(--tz-text-primary);
}

.footer-site-overview__paragraph {
  margin: 0 0 1rem;
  min-width: 0;
}

.footer-site-overview__link {
  color: inherit;
  overflow-wrap: anywhere;
  word-break: break-word;
  text-decoration: none;
  transition: color 0.2s ease;
}

.footer-site-overview__route-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.footer-site-overview__link-arrow {
  width: 0.95rem;
  height: 0.95rem;
  flex: 0 0 auto;
  transition: transform 0.2s ease;
}

.footer-site-overview__links {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  margin: 1.25rem 0 0;
  padding: 1rem 0 0;
  border-top: 1px solid var(--tz-border-subtle);
  list-style: none;
}

.footer-site-overview__link-item {
  margin: 0;
}

.footer-site-overview__link:hover,
.footer-site-overview__link:focus-visible {
  color: var(--tz-text-primary);
}

.footer-site-overview__route-link:hover .footer-site-overview__link-arrow,
.footer-site-overview__route-link:focus-visible .footer-site-overview__link-arrow {
  transform: translate(0.12rem, -0.12rem);
}

@media (max-width: 767px) {
  .footer-site-overview {
    text-align: center;
  }
}
</style>
