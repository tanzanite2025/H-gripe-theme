<template>
  <article id="freehub-cassette-fitment-tool" class="freehub-cassette-fitment-page">
    <InteractiveToolPageHeader
      :title="t('guidesWheelsetBuyersChooseFreehub.seo.title')"
    />

    <section class="wheelset-section sizecharts-section" aria-labelledby="freehub-cassette-fitment-section-title">
      <h2 id="freehub-cassette-fitment-section-title" class="sizecharts-section__title">
        {{ t('guidesWheelsetBuyersChooseFreehub.intro.title') }}
      </h2>
      <WheelsetChooseFreehubSection>
        <template #after-helper>
          <section class="wheelset-freehub-guide-intro" aria-labelledby="freehub-guide-intro-title">
            <div class="wheelset-freehub-guide-intro__copy">
              <h3 id="freehub-guide-intro-title" class="wheelset-freehub-guide-intro__title">
                {{ t('guidesWheelsetBuyersChooseFreehub.guideIntro.title') }}
              </h3>
              <p class="wheelset-guide-step__body">
                {{ t('guidesWheelsetBuyersChooseFreehub.guideIntro.body') }}
              </p>
            </div>

            <div class="wheelset-freehub-guide-intro__steps wheelset-freehub-guide-intro__steps--desktop">
              <article
                v-for="(step, index) in freehubGuideSteps"
                :key="step.title"
                class="wheelset-guide-step"
              >
                <div class="wheelset-guide-step__number">{{ index + 1 }}</div>
                <div class="wheelset-guide-step__content">
                  <h4 class="wheelset-guide-step__title">{{ step.title }}</h4>
                  <p class="wheelset-guide-step__body">{{ step.body }}</p>
                </div>
              </article>
            </div>

            <div class="wheelset-freehub-guide-intro__steps wheelset-freehub-guide-intro__steps--mobile">
              <SmartAccordion>
                <AccordionItem
                  v-for="(step, index) in freehubGuideSteps"
                  :id="`freehub-guide-step-${index + 1}`"
                  :key="step.title"
                  :title="`${index + 1}. ${step.title}`"
                >
                  <p class="wheelset-guide-step__body">{{ step.body }}</p>
                </AccordionItem>
              </SmartAccordion>
            </div>

            <div class="wheelset-guide-panel-grid wheelset-freehub-guide-intro__help">
              <div class="wheelset-guide-panel">
                <strong>{{ t('guidesWheelsetBuyersChooseFreehub.guideIntro.identifyTitle') }}</strong>
                <p>{{ t('guidesWheelsetBuyersChooseFreehub.guideIntro.identifyBody') }}</p>
              </div>
              <div class="wheelset-guide-panel">
                <strong>{{ t('guidesWheelsetBuyersChooseFreehub.guideIntro.photoTitle') }}</strong>
                <p>{{ t('guidesWheelsetBuyersChooseFreehub.guideIntro.photoBody') }}</p>
              </div>
            </div>

            <div class="wheelset-guide-actions wheelset-freehub-guide-intro__actions">
              <button
                type="button"
                class="wheelset-guide-button wheelset-guide-button--solid"
                :aria-label="t('guidesWheelsetBuyersChooseFreehub.guideIntro.chatButton')"
                @click="openFreehubSupportChat"
              >
                <Icon name="lucide:message-circle" class="h-4 w-4" aria-hidden="true" />
                {{ t('guidesWheelsetBuyersChooseFreehub.guideIntro.chatButton') }}
              </button>
            </div>
          </section>
        </template>
      </WheelsetChooseFreehubSection>
    </section>
  </article>
</template>

<script setup lang="ts">
import InteractiveToolPageHeader from '~/components/InteractiveToolPageHeader.vue'
import { computed } from 'vue'
import { useI18n } from '#imports'
import AccordionItem from '~/components/ui/AccordionItem.vue'
import SmartAccordion from '~/components/ui/SmartAccordion.vue'
import WheelsetChooseFreehubSection from '~/components/WheelsetChooseFreehubSection.vue'
import { useChatWidget } from '~/composables/useChatWidget'

const { t } = useI18n()
const { openChat } = useChatWidget()

const freehubGuideSteps = computed(() => [
  {
    title: t('guidesWheelsetBuyersChooseFreehub.guideIntro.step1.title'),
    body: t('guidesWheelsetBuyersChooseFreehub.guideIntro.step1.body'),
  },
  {
    title: t('guidesWheelsetBuyersChooseFreehub.guideIntro.step2.title'),
    body: t('guidesWheelsetBuyersChooseFreehub.guideIntro.step2.body'),
  },
  {
    title: t('guidesWheelsetBuyersChooseFreehub.guideIntro.step3.title'),
    body: t('guidesWheelsetBuyersChooseFreehub.guideIntro.step3.body'),
  },
])

const openFreehubSupportChat = () => {
  openChat({
    showAgentList: true,
    source: 'wheelset-freehub-guide',
  })
}
</script>

<style scoped>
.freehub-cassette-fitment-page {
  --wheelset-border: rgba(203, 213, 225, 0.72);
  --tz-text-accent: var(--tz-site-accent);
  width: 100%;
  margin: 0 auto;
}

.freehub-cassette-fitment-page .sizecharts-section__title {
  margin: 0 0 2rem;
  color: var(--tz-text-primary);
  font-weight: 650;
  text-align: center;
}

.wheelset-freehub-guide-intro {
  display: grid;
  gap: 1.15rem;
  width: 100%;
  margin: 0 0 2rem;
  border: 1px solid var(--wheelset-border);
  border-radius: 0.5rem;
  background: var(--tz-surface-subtle);
  padding: clamp(1rem, 2vw, 1.5rem);
}

.wheelset-freehub-guide-intro__copy {
  max-width: 72rem;
}

.wheelset-freehub-guide-intro__title {
  margin: 0 0 0.55rem;
  color: var(--tz-text-primary);
  font-size: 1.15rem;
  font-weight: 700;
  line-height: 1.35;
}

.wheelset-freehub-guide-intro__copy .wheelset-guide-step__body {
  max-width: 68rem;
  margin-bottom: 0;
}

.wheelset-freehub-guide-intro__steps {
  min-width: 0;
}

.wheelset-freehub-guide-intro__steps--desktop {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1.35rem;
}

.wheelset-freehub-guide-intro__steps--mobile {
  display: none;
}

.wheelset-freehub-guide-intro__help,
.wheelset-freehub-guide-intro__actions {
  margin-top: 0;
}

.wheelset-guide-step {
  position: relative;
  display: block;
}

.wheelset-guide-step__number {
  position: absolute;
  top: 0;
  left: 0;
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1;
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.35);
}

.wheelset-guide-step__content {
  width: 100%;
  min-width: 0;
  padding-top: 0;
}

.wheelset-guide-step__title {
  display: flex;
  min-height: 2rem;
  align-items: center;
  margin: 0 0 0.55rem;
  padding-left: 3rem;
  color: var(--tz-text-primary);
  font-size: 1.05rem;
  font-weight: 700;
  line-height: 1.35;
}

.wheelset-guide-step__body {
  margin: 0 0 1rem;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.7;
  text-align: left;
}

.wheelset-guide-panel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
  margin-top: 1rem;
}

.wheelset-guide-panel {
  border-radius: 0.5rem;
  background: var(--tz-card-surface);
  padding: 1rem;
}

.wheelset-guide-panel strong {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  max-width: 100%;
  margin-bottom: 0.65rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 9999px;
  color: var(--tz-text-primary);
  padding: 0.28rem 0.65rem;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.35;
  white-space: normal;
}

.wheelset-guide-panel p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.65;
  text-align: left;
}

.wheelset-guide-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 1rem;
}

.wheelset-guide-button {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  border: 1px solid var(--tz-action-primary);
  border-radius: 9999px;
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  padding: 0.6rem 1rem;
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1.25;
  transition: background-color 0.16s ease, border-color 0.16s ease, color 0.16s ease;
}

.wheelset-guide-button:hover {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
  color: #ffffff;
}

@media (max-width: 860px) {
  .wheelset-freehub-guide-intro__steps--desktop {
    display: none;
  }

  .wheelset-freehub-guide-intro__steps--mobile {
    display: block;
  }

  .wheelset-guide-step__number {
    width: 1.75rem;
    height: 1.75rem;
  }

  .wheelset-guide-step__title {
    min-height: 1.75rem;
    padding-left: 2.5rem;
  }

  .wheelset-guide-panel-grid {
    grid-template-columns: 1fr;
  }

  .wheelset-guide-button {
    width: 100%;
  }
}
</style>
