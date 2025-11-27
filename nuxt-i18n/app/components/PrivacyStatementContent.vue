<template>
  <div class="text-sm text-white/80 leading-relaxed space-y-4">
    <p class="font-semibold text-white">
      {{ $t('privacy.policyHeading', 'Privacy Policy for E-Commerce Website') }}
    </p>
    <p class="text-white/60 text-xs">
      {{ $t('privacy.effectiveDate', 'Effective Date: [2025/11/11]') }}
    </p>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section1Title', '1. Introduction') }}
    </p>
    <p>
      {{
        $t(
          'privacy.section1Body',
          'Welcome to [www.tanzanite.site]. We value your privacy and are committed to protecting your personal information. This Privacy Policy explains how we collect, use, disclose, and safeguard your data when you visit our website [www.tanzanite.site] and use our services.'
        )
      }}
    </p>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section2Title', '2. Information We Collect') }}
    </p>
    <p>
      {{
        $t(
          'privacy.section2Intro',
          'We may collect the following types of personal information:'
        )
      }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        <span class="font-semibold">
          {{ $t('privacy.identityLabel', 'Identity Data') }}:
        </span>
        {{ $t('privacy.identityDesc', 'Full name, username, password') }}
      </li>
      <li>
        <span class="font-semibold">
          {{ $t('privacy.contactLabel', 'Contact Data') }}:
        </span>
        {{
          $t(
            'privacy.contactDesc',
            'Email address, phone number, shipping/billing address'
          )
        }}
      </li>
      <li>
        <span class="font-semibold">
          {{ $t('privacy.transactionLabel', 'Transaction Data') }}:
        </span>
        {{
          $t(
            'privacy.transactionDesc',
            'Purchase history, payment details (processed via secure third-party gateways)'
          )
        }}
      </li>
      <li>
        <span class="font-semibold">
          {{ $t('privacy.technicalLabel', 'Technical Data') }}:
        </span>
        {{
          $t(
            'privacy.technicalDesc',
            'IP address, browser type, device information, cookies'
          )
        }}
      </li>
      <li>
        <span class="font-semibold">
          {{ $t('privacy.marketingLabel', 'Marketing Data') }}:
        </span>
        {{ $t('privacy.marketingDesc', 'Preferences, feedback, survey responses') }}
      </li>
    </ul>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section3Title', '3. How We Use Your Information') }}
    </p>
    <p>
      {{ $t('privacy.section3Intro', 'We use your data to:') }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{ $t('privacy.useProcessOrders', 'Process orders and deliver products') }}
      </li>
      <li>
        {{
          $t(
            'privacy.useManageAccount',
            'Manage your account and provide customer support'
          )
        }}
      </li>
      <li>
        {{
          $t(
            'privacy.useImproveSite',
            'Improve website functionality and user experience'
          )
        }}
      </li>
      <li>
        {{
          $t(
            'privacy.usePromotions',
            'Send promotional offers (with your consent)'
          )
        }}
      </li>
      <li>
        {{ $t('privacy.useLegal', 'Comply with legal obligations') }}
      </li>
    </ul>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section4Title', '4. Sharing Your Information') }}
    </p>
    <p>
      {{ $t('privacy.section4Intro', 'We may share your information with:') }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{
          $t(
            'privacy.shareServiceProviders',
            'Service Providers: Payment processors, logistics partners, analytics tools'
          )
        }}
      </li>
      <li>
        {{
          $t(
            'privacy.shareLegalAuthorities',
            'Legal Authorities: When required by law or to protect our rights'
          )
        }}
      </li>
      <li>
        {{
          $t(
            'privacy.shareBusinessTransfers',
            'Business Transfers: In case of merger, acquisition, or asset sale'
          )
        }}
      </li>
    </ul>
    <p>
      {{
        $t(
          'privacy.shareNoSale',
          'We do not sell your personal data to third parties.'
        )
      }}
    </p>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section5Title', '5. Cookies and Tracking Technologies') }}
    </p>
    <p>
      {{
        $t(
          'privacy.section5Intro',
          'We use cookies and similar technologies to:'
        )
      }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{ $t('privacy.cookiesRemember', 'Remember your preferences') }}
      </li>
      <li>
        {{ $t('privacy.cookiesAnalytics', 'Analyze site traffic and usage') }}
      </li>
      <li>
        {{ $t('privacy.cookiesPersonalize', 'Personalize content and ads') }}
      </li>
    </ul>
    <p>
      {{
        $t(
          'privacy.cookiesSettings',
          'You can manage cookie settings in your browser.'
        )
      }}
    </p>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section6Title', '6. Data Security') }}
    </p>
    <p>
      {{
        $t(
          'privacy.section6Body',
          'We implement appropriate technical and organizational measures to protect your data from unauthorized access, alteration, disclosure, or destruction.'
        )
      }}
    </p>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section7Title', '7. Your Rights') }}
    </p>
    <p>
      {{
        $t(
          'privacy.section7Intro',
          'Depending on your jurisdiction, you may have the right to:'
        )
      }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{
          $t(
            'privacy.rightsAccess',
            'Access, correct, or delete your personal data'
          )
        }}
      </li>
      <li>
        {{
          $t(
            'privacy.rightsWithdrawConsent',
            'Withdraw consent at any time'
          )
        }}
      </li>
      <li>
        {{
          $t(
            'privacy.rightsObjectProcessing',
            'Object to certain data processing'
          )
        }}
      </li>
      <li>
        {{
          $t(
            'privacy.rightsPortability',
            'Request data portability'
          )
        }}
      </li>
    </ul>
    <p>
      {{
        $t(
          'privacy.rightsContact',
          'To exercise these rights, please contact us.'
        )
      }}
    </p>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section8Title', '8. International Transfers') }}
    </p>
    <p>
      {{
        $t(
          'privacy.section8Body',
          'If you are located outside [Your Country], your data may be transferred to and processed in [Your Country], where data protection laws may differ.'
        )
      }}
    </p>

    <p class="font-semibold text-white mt-4">
      {{ $t('privacy.section9Title', '9. Changes to This Policy') }}
    </p>
    <p>
      {{
        $t(
          'privacy.section9Body',
          'We may update this Privacy Policy from time to time. Changes will be posted on this page with an updated effective date.'
        )
      }}
    </p>
  </div>
</template>

<script setup lang="ts">
// purely presentational content component
</script>
