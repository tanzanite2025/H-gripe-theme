<template>
  <div class="company-page">
    <h1 class="sr-only">Cookie Policy</h1>

    <div class="policies-content">
      <p class="text-sm tz-text-secondary mb-6">
        This page explains how Tanzanite uses cookies and similar technologies to improve your browsing experience.
      </p>

      <div class="mt-4 space-y-8 tz-text-secondary">
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">What are Cookies?</h2>
          <p class="text-sm leading-relaxed">
            Cookies are small text files that are stored on your device when you visit a website. They help the website remember your preferences and improve your browsing experience.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold text-white mb-3">Types of Cookies We Use</h2>
          <div class="space-y-4">
            <div class="rounded-xl p-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] backdrop-blur-md shadow-[0_18px_45px_-18px_rgba(0,0,0,1)]">
              <h3 class="text-base font-medium text-cyan-400 mb-2">Essential Cookies</h3>
              <p class="text-sm leading-relaxed mb-2">
                These cookies are necessary for the website to function properly. They enable basic features like page navigation, secure areas access, and shopping cart functionality.
              </p>
              <p class="text-xs tz-text-muted">Cannot be disabled</p>
            </div>

            <div class="rounded-xl p-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] backdrop-blur-md shadow-[0_18px_45px_-18px_rgba(0,0,0,1)]">
              <h3 class="text-base font-medium text-cyan-400 mb-2">Performance Cookies</h3>
              <p class="text-sm leading-relaxed mb-2">
                These cookies help us understand how visitors interact with our website by collecting and reporting information anonymously. This helps us improve our website's performance.
              </p>
              <p class="text-xs tz-text-muted">Can be disabled</p>
            </div>

            <div class="rounded-xl p-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] backdrop-blur-md shadow-[0_18px_45px_-18px_rgba(0,0,0,1)]">
              <h3 class="text-base font-medium text-cyan-400 mb-2">Functional Cookies</h3>
              <p class="text-sm leading-relaxed mb-2">
                These cookies enable enhanced functionality and personalization, such as remembering your language preference or region. They may be set by us or by third-party providers.
              </p>
              <p class="text-xs tz-text-muted">Can be disabled</p>
            </div>

            <div class="rounded-xl p-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] backdrop-blur-md shadow-[0_18px_45px_-18px_rgba(0,0,0,1)]">
              <h3 class="text-base font-medium text-cyan-400 mb-2">Advertising Cookies</h3>
              <p class="text-sm leading-relaxed mb-2">
                These cookies are used to deliver advertisements that are more relevant to you and your interests. They also help limit the number of times you see an advertisement.
              </p>
              <p class="text-xs tz-text-muted">Can be disabled</p>
            </div>
          </div>
        </section>

        <section>
          <h2 class="text-lg font-semibold text-white mb-3">Managing Your Cookie Preferences</h2>
          <p class="text-sm leading-relaxed mb-4">
            You can manage your cookie preferences at any time. Most web browsers allow you to control cookies through their settings. You can:
          </p>
          <ul class="list-disc list-inside text-sm space-y-2 tz-text-secondary">
            <li>View what cookies are stored on your device</li>
            <li>Delete all or specific cookies</li>
            <li>Block cookies from being set</li>
            <li>Set your browser to notify you when cookies are being set</li>
          </ul>
          <p class="text-sm leading-relaxed mt-4 tz-text-secondary">
            Please note that disabling certain cookies may affect the functionality of our website.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold text-white mb-3">Third-Party Cookies</h2>
          <p class="text-sm leading-relaxed">
            Some cookies on our website are set by third-party services, including analytics providers and payment processors. These third parties have their own privacy policies and we encourage you to review them.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold text-white mb-3">Updates to This Policy</h2>
          <p class="text-sm leading-relaxed">
            We may update this Cookie Policy from time to time to reflect changes in our practices or for other operational, legal, or regulatory reasons. Please revisit this page periodically to stay informed about our use of cookies.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold text-white mb-3">Contact Us</h2>
          <p class="text-sm leading-relaxed">
            If you have any questions about our use of cookies, please contact us at
            <a href="mailto:support@tanzanite.site" class="text-cyan-400 hover:text-cyan-300 underline">
              support@tanzanite.site
            </a>
          </p>
        </section>

        <div class="pt-6 border-t border-slate-800">
        <p class="text-xs tz-text-muted">Last updated: December 2024</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useHead, definePageMeta } from '#imports'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Cookie Policy',
})
</script>

<style scoped>
.company-page {
  width: 100%;
  max-width: none;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.policies-content {
  margin-top: 0;
}

@media (max-width: 767px) {
  .company-page {
    max-width: 100%;
    padding-inline: 0;
  }
}
</style>
