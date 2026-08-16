<template>
  <section class="contact-service-entry" aria-labelledby="contact-service-entry-title">
    <div class="contact-service-entry__copy">
      <p class="contact-service-entry__eyebrow">
        <Icon name="lucide:sparkles" class="h-4 w-4" aria-hidden="true" />
        Connect with us
      </p>
      <h2 id="contact-service-entry-title" class="contact-service-entry__title">
        Start a conversation with our team.
      </h2>
      <p class="contact-service-entry__description">
        For product advice, dealer enquiries, OEM/ODM projects, or order support, contact us by email or continue in the service chat. Chat messages are handled in our existing support inbox.
      </p>

      <div class="contact-service-entry__actions">
        <a
          v-if="contactEmail"
          :href="emailHref"
          class="contact-service-entry__action"
        >
          <Icon name="lucide:mail" class="h-5 w-5 shrink-0" aria-hidden="true" />
          <span class="min-w-0 text-left">
            <span class="contact-service-entry__action-label">Email</span>
            <span class="contact-service-entry__action-value">{{ contactEmail }}</span>
          </span>
          <Icon name="lucide:arrow-up-right" class="h-4 w-4 shrink-0" aria-hidden="true" />
        </a>
        <span v-else class="contact-service-entry__action contact-service-entry__action--unavailable" aria-disabled="true">
          <Icon name="lucide:mail" class="h-5 w-5 shrink-0" aria-hidden="true" />
          <span class="min-w-0 text-left">
            <span class="contact-service-entry__action-label">Email</span>
            <span class="contact-service-entry__action-value">Email support</span>
          </span>
        </span>

        <button
          type="button"
          class="contact-service-entry__action contact-service-entry__action--chat"
          @click="openSupportChat"
        >
          <Icon name="lucide:messages-square" class="h-5 w-5 shrink-0" aria-hidden="true" />
          <span class="min-w-0 text-left">
            <span class="contact-service-entry__action-label">Online consultation</span>
            <span class="contact-service-entry__action-value">Open service chat</span>
          </span>
          <Icon name="lucide:arrow-up-right" class="h-4 w-4 shrink-0" aria-hidden="true" />
        </button>
      </div>
    </div>

    <aside class="contact-service-entry__panel">
      <div class="contact-service-entry__status">
        <span class="contact-service-entry__status-dot" aria-hidden="true"></span>
        Support channel available
      </div>
      <div class="contact-service-entry__panel-icon" aria-hidden="true">
        <Icon name="lucide:headset" class="h-8 w-8" />
      </div>
      <h3>Talk to our support team</h3>
      <p>Use service chat for a direct conversation with the team, without creating a separate contact form.</p>
      <button type="button" class="contact-service-entry__chat-button" @click="openSupportChat">
        <Icon name="lucide:message-circle" class="h-4 w-4" aria-hidden="true" />
        Open service chat
      </button>
    </aside>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { useSiteSettings } from '~/composables/usePublicSettings'

const { openChat } = useChatWidget()
const { siteSettings } = useSiteSettings()

const contactEmail = computed(() => siteSettings.value.contactEmail?.trim() || '')
const emailHref = computed(() => `mailto:${contactEmail.value}`)

const openSupportChat = () => {
  openChat({
    showAgentList: true,
    source: 'company-contact',
  })
}
</script>

<style scoped>
.contact-service-entry {
  display: grid;
  gap: 1.5rem;
  padding: 1.5rem;
  border-radius: 1rem;
  background: var(--tz-card-surface);
  box-shadow: 0 18px 38px rgba(0, 0, 0, 0.3);
}

.contact-service-entry__copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
}

.contact-service-entry__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  color: #b5ff6d;
  font-size: var(--tz-type-caption);
  font-weight: 700;
  line-height: 1.3;
  text-transform: uppercase;
}

.contact-service-entry__title {
  max-width: 14ch;
  margin: 0.75rem 0 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.75rem, 2.8vw, 3.25rem);
  font-weight: 700;
  line-height: 1.08;
}

.contact-service-entry__description {
  max-width: 57ch;
  margin: 0.875rem 0 0;
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-description);
  line-height: 1.65;
}

.contact-service-entry__actions {
  display: grid;
  width: 100%;
  max-width: 43rem;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.contact-service-entry__action {
  display: grid;
  width: 100%;
  min-height: 4.5rem;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0.875rem;
  border: 1px solid rgba(255, 255, 255, 0.13);
  border-radius: 6px;
  background: #121217;
  color: var(--tz-text-primary);
  text-decoration: none;
  transition: border-color 0.18s ease, background-color 0.18s ease, transform 0.18s ease;
}

.contact-service-entry__action:hover {
  border-color: rgba(181, 255, 109, 0.72);
  background: #17171d;
  transform: translateY(-1px);
}

.contact-service-entry__action--chat {
  cursor: pointer;
  font: inherit;
  text-align: inherit;
}

.contact-service-entry__action--unavailable {
  border-style: dashed;
  color: var(--tz-text-muted);
  opacity: 0.72;
}

.contact-service-entry__action-label,
.contact-service-entry__action-value {
  display: block;
  overflow-wrap: anywhere;
}

.contact-service-entry__action-label {
  color: var(--tz-text-muted);
  font-size: var(--tz-type-caption);
  font-weight: 600;
  line-height: 1.25;
}

.contact-service-entry__action-value {
  margin-top: 0.125rem;
  color: var(--tz-text-primary);
  font-size: var(--tz-type-description);
  font-weight: 600;
  line-height: 1.35;
}

.contact-service-entry__panel {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
  padding: 1.25rem;
  border-radius: 1rem;
  background: #0d0d10;
}

.contact-service-entry__status {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.625rem;
  border: 1px solid rgba(181, 255, 109, 0.18);
  border-radius: 999px;
  background: rgba(181, 255, 109, 0.09);
  color: #d9ffb2;
  font-size: var(--tz-type-caption);
  font-weight: 600;
  line-height: 1.25;
}

.contact-service-entry__status-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 999px;
  background: #b5ff6d;
  box-shadow: 0 0 0 3px rgba(181, 255, 109, 0.1);
}

.contact-service-entry__panel-icon {
  display: grid;
  width: 3.5rem;
  height: 3.5rem;
  margin-top: auto;
  place-items: center;
  border: 1px solid rgba(181, 255, 109, 0.28);
  border-radius: 999px;
  color: #b5ff6d;
}

.contact-service-entry__panel h3 {
  margin: 1rem 0 0;
  color: var(--tz-text-primary);
  font-size: var(--tz-type-section-title);
  font-weight: 700;
  line-height: 1.3;
}

.contact-service-entry__panel p {
  margin: 0.5rem 0 0;
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-description);
  line-height: 1.6;
}

.contact-service-entry__chat-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.625rem;
  margin-top: 1.25rem;
  padding: 0.625rem 0.875rem;
  border: 1px solid #b5ff6d;
  border-radius: 6px;
  background: #b5ff6d;
  color: #0b0b0e;
  cursor: pointer;
  font-size: var(--tz-type-caption);
  font-weight: 700;
  line-height: 1.2;
  transition: background-color 0.18s ease, border-color 0.18s ease, transform 0.18s ease;
}

.contact-service-entry__chat-button:hover {
  border-color: #c8ff91;
  background: #c8ff91;
  transform: translateY(-1px);
}

@media (min-width: 900px) {
  .contact-service-entry {
    grid-template-columns: minmax(0, 1.45fr) minmax(18rem, 0.55fr);
    align-items: stretch;
    padding: 2rem;
  }

  .contact-service-entry__panel {
    padding: 1.5rem;
  }
}

@media (max-width: 639px) {
  .contact-service-entry {
    padding: 1rem;
  }

  .contact-service-entry__title {
    max-width: 13ch;
  }
}
</style>
