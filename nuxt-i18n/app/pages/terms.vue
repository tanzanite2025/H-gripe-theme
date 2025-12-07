<template>
  <div class="min-h-screen bg-slate-950 text-slate-50 pt-24 pb-16 md:pt-28">
    <div class="max-w-4xl mx-auto px-4">
      <!-- H1 隐藏，SEO 保留 -->
      <h1 class="sr-only">Terms of Service</h1>
      <p class="text-sm text-slate-400 mb-6">
        Please read these terms carefully before using our website or making a purchase.
      </p>

      <div class="mt-4 space-y-8 text-slate-300">
        <!-- Agreement -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">1. Agreement to Terms</h2>
          <p class="text-sm leading-relaxed">
            By accessing or using the Tanzanite website, you agree to be bound by these Terms of Service 
            and all applicable laws and regulations. If you do not agree with any of these terms, you 
            are prohibited from using or accessing this site.
          </p>
        </section>

        <!-- Use License -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">2. Use License</h2>
          <p class="text-sm leading-relaxed mb-3">
            Permission is granted to temporarily view the materials on Tanzanite's website for personal, 
            non-commercial use only. This license does not include:
          </p>
          <ul class="list-disc list-inside text-sm space-y-2 text-slate-400">
            <li>Modifying or copying the materials</li>
            <li>Using the materials for any commercial purpose</li>
            <li>Attempting to reverse engineer any software on the website</li>
            <li>Removing any copyright or proprietary notations</li>
            <li>Transferring the materials to another person</li>
          </ul>
        </section>

        <!-- Products and Pricing -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">3. Products and Pricing</h2>
          <p class="text-sm leading-relaxed mb-3">
            All products displayed on our website are subject to availability. We reserve the right to:
          </p>
          <ul class="list-disc list-inside text-sm space-y-2 text-slate-400">
            <li>Limit quantities available for purchase</li>
            <li>Discontinue any product at any time</li>
            <li>Refuse any order you place with us</li>
            <li>Correct any errors in pricing or product descriptions</li>
          </ul>
          <p class="text-sm leading-relaxed mt-3">
            Prices are subject to change without notice. We make every effort to display accurate pricing, 
            but errors may occur. If we discover a pricing error, we will notify you and give you the 
            option to proceed or cancel your order.
          </p>
        </section>

        <!-- Orders and Payment -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">4. Orders and Payment</h2>
          <p class="text-sm leading-relaxed">
            By placing an order, you warrant that you are legally capable of entering into binding contracts. 
            All orders are subject to acceptance and availability. We accept major credit cards and other 
            payment methods as displayed at checkout. Payment must be received in full before order processing.
          </p>
        </section>

        <!-- Shipping and Delivery -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">5. Shipping and Delivery</h2>
          <p class="text-sm leading-relaxed">
            Shipping times and costs vary depending on your location and chosen shipping method. 
            Delivery times are estimates and not guaranteed. We are not responsible for delays caused 
            by customs, weather, or other factors beyond our control. Risk of loss transfers to you 
            upon delivery to the carrier.
          </p>
        </section>

        <!-- Returns and Refunds -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">6. Returns and Refunds</h2>
          <p class="text-sm leading-relaxed">
            We accept returns within 30 days of delivery for unused items in original packaging. 
            Custom or personalized items are non-refundable. Refunds will be processed to the original 
            payment method within 5-10 business days after we receive the returned item. Shipping costs 
            for returns are the customer's responsibility unless the item is defective.
          </p>
        </section>

        <!-- Warranty -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">7. Warranty</h2>
          <p class="text-sm leading-relaxed">
            Products are covered by manufacturer warranty as specified in product descriptions. 
            Warranty does not cover damage from misuse, accidents, or normal wear and tear. 
            For warranty claims, please contact our customer support with proof of purchase.
          </p>
        </section>

        <!-- Limitation of Liability -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">8. Limitation of Liability</h2>
          <p class="text-sm leading-relaxed">
            Tanzanite shall not be liable for any indirect, incidental, special, consequential, or 
            punitive damages arising from your use of our website or products. Our total liability 
            shall not exceed the amount you paid for the product in question.
          </p>
        </section>

        <!-- Intellectual Property -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">9. Intellectual Property</h2>
          <p class="text-sm leading-relaxed">
            All content on this website, including text, graphics, logos, images, and software, 
            is the property of Tanzanite or its content suppliers and is protected by international 
            copyright laws. The Tanzanite name and logo are registered trademarks.
          </p>
        </section>

        <!-- Governing Law -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">10. Governing Law</h2>
          <p class="text-sm leading-relaxed">
            These terms shall be governed by and construed in accordance with applicable laws, 
            without regard to conflict of law principles. Any disputes arising from these terms 
            shall be resolved in the appropriate courts.
          </p>
        </section>

        <!-- Changes to Terms -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">11. Changes to Terms</h2>
          <p class="text-sm leading-relaxed">
            We reserve the right to modify these terms at any time. Changes will be effective 
            immediately upon posting to the website. Your continued use of the website after 
            changes constitutes acceptance of the modified terms.
          </p>
        </section>

        <!-- Contact -->
        <section>
          <h2 class="text-lg font-semibold text-white mb-3">12. Contact Us</h2>
          <p class="text-sm leading-relaxed">
            If you have any questions about these Terms of Service, please contact us at 
            <a href="mailto:support@tanzanite.com" class="text-cyan-400 hover:text-cyan-300 underline">
              support@tanzanite.com
            </a>
          </p>
        </section>

        <!-- Last Updated -->
        <div class="pt-6 border-t border-slate-800">
          <p class="text-xs text-slate-500">Last updated: December 2024</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
useHead({
  title: 'Terms of Service',
})
</script>
