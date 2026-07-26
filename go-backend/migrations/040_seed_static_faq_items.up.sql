-- Seed the editable FAQ table with the storefront static FAQ fallback content.
-- These rows make existing page FAQ questions visible in the admin FAQ list,
-- instead of leaving only page/category shells with 0 editable FAQ items.

WITH seed_faqs(page_id, category, question, answer, answer_image_url, answer_image_alt, answer_image_width, answer_image_height, sort_order) AS (
  VALUES
    ('company-certificates', 'general', 'Are all TANZANITE wheels UCI approved?', 'Yes, our core road and track wheelsets are UCI approved and listed on the official Union Cycliste Internationale website. This means they are certified for use in all professional and amateur racing events worldwide.', '', '', 0, 0, 110),
    ('company-certificates', 'general', 'What is ISO 4210 certification?', 'ISO 4210 is the international safety standard for bicycles and components. Meeting this standard ensures that our wheels have passed rigorous impact, fatigue, and environmental tests to guarantee rider safety.', '', '', 0, 0, 120),
    ('company-certificates', 'general', 'How does your internal testing differ from standard requirements?', 'We believe standard compliance is just the starting point. Our internal laboratory tests subjects rims to impact energies and fatigue cycles that are 120% to 150% higher than ISO requirements to ensure durability under extreme conditions.', '', '', 0, 0, 130),
    ('company-certificates', 'general', 'Can I request specific test reports?', 'Yes, we can provide detailed test reports for specific production batches upon request for our ODM/OEM partners. Please contact our support team for more information.', '', '', 0, 0, 140),
    ('company-contact', 'general', 'What is your typical response time?', 'We strive to respond to all inquiries within 24 hours during business days (Mon-Fri). For urgent technical support, please use our WhatsApp channel for faster assistance.', '', '', 0, 0, 110),
    ('company-contact', 'general', 'Can I visit the Xiamen factory?', 'Yes, we welcome visits from our OEM/ODM partners. Please contact our sales team at least 2 weeks in advance to schedule a tour and ensure the appropriate staff are available to meet you.', '', '', 0, 0, 120),
    ('company-contact', 'general', 'Do you have local distributors in my country?', 'We operate primarily through a direct-to-consumer model to offer the best prices. However, we have a growing network of service partners in key regions. Please contact us to find the nearest partner.', '', '', 0, 0, 130),
    ('company-contact', 'general', 'How can I apply for sponsorship?', 'We are always looking for passionate riders and teams. Please send your racing resume and proposal to our support email with the subject line "Sponsorship Application".', '', '', 0, 0, 140),
    ('company-global-partners', 'partnership', 'What are the criteria for becoming a distributor?', 'We look for partners with established distribution channels, technical service capabilities, and a commitment to brand building. A background in premium cycling components is preferred.', '', '', 0, 0, 110),
    ('company-global-partners', 'partnership', 'Do you offer OEM/ODM services for global brands?', 'Yes, we provide comprehensive OEM/ODM solutions, including mold design, layup optimization, and private labeling for established bicycle brands.', '', '', 0, 0, 120),
    ('company-global-partners', 'partnership', 'How do you handle global evaluations and logistics?', 'We have extensive experience in global logistics, including DDP shipping to key markets. We can also arrange sample evaluations for potential partners to verify our quality standards.', '', '', 0, 0, 130),
    ('company-global-partners', 'partnership', 'Is regional exclusivity available?', 'Regional exclusivity is negotiable based on projected volume, market coverage, and a proven track record of sales performance.', '', '', 0, 0, 140),
    ('company-membership', 'membership', 'What membership tiers are available?', 'We offer several membership tiers based on your purchase history:
            <ul>
              <li><strong>Bronze</strong> - New members (0-499 points)</li>
              <li><strong>Silver</strong> - 500-1,999 points</li>
              <li><strong>Gold</strong> - 2,000-4,999 points</li>
              <li><strong>Platinum</strong> - 5,000+ points</li>
            </ul>
            Higher tiers unlock better discounts and exclusive benefits.', '', '', 0, 0, 110),
    ('company-membership', 'membership', 'How do I upgrade my membership tier?', 'Your membership tier is automatically upgraded based on your accumulated points. Points are earned through purchases, reviews, and other activities. Once you reach the required points threshold, your tier will be upgraded immediately.', '', '', 0, 0, 120),
    ('company-membership', 'membership', 'Do membership tiers expire?', 'Membership tiers are evaluated annually. To maintain your current tier, you need to earn at least 50% of the tier''s point threshold within each 12-month period. If you don''t meet this requirement, you may be moved to a lower tier.', '', '', 0, 0, 130),
    ('company-membership', 'points', 'How do I earn points?', 'You can earn points through various activities:
            <ul>
              <li><strong>Purchases</strong> - 1 point per $1 spent</li>
              <li><strong>Product reviews</strong> - 50 points per review</li>
              <li><strong>Referrals</strong> - 200 points when a friend makes their first purchase</li>
              <li><strong>Daily login</strong> - 1 point per day (30-day validity)</li>
              <li><strong>Birthday bonus</strong> - Double points during your birthday month</li>
            </ul>', '', '', 0, 0, 210),
    ('company-membership', 'points', 'How do I redeem my points?', 'Points can be redeemed at checkout for discounts on your order. The redemption rate is typically 100 points = $1 discount. You can choose how many points to redeem, up to a maximum of 20% of your order total.', '', '', 0, 0, 220),
    ('company-membership', 'points', 'Do points expire?', 'Points earned from purchases are valid for 24 months from the date earned. Daily login points expire after 30 days. Bonus points from promotions may have different expiration periods - check the promotion details for specifics.', '', '', 0, 0, 230),
    ('company-membership', 'points', 'Can I transfer points to another account?', 'Points are non-transferable and can only be used by the account holder. Each account must earn and redeem their own points.', '', '', 0, 0, 240),
    ('company-membership', 'benefits', 'What benefits do members receive?', 'Member benefits vary by tier and include:
            <ul>
              <li>Exclusive member-only discounts</li>
              <li>Early access to new products and sales</li>
              <li>Free shipping on orders over a certain amount</li>
              <li>Birthday rewards and special promotions</li>
              <li>Priority customer support (Gold and above)</li>
            </ul>', '', '', 0, 0, 310),
    ('company-membership', 'benefits', 'How do I access member-only deals?', 'Member-only deals are automatically displayed when you''re logged into your account. You''ll also receive email notifications about exclusive offers. Make sure your email preferences are set to receive promotional emails.', '', '', 0, 0, 320),
    ('company-oem-odm', 'general', 'What is the Minimum Order Quantity (MOQ) for OEM/ODM?', 'Our MOQ is flexible to support growing brands. For standard OEM rims, MOQ can be as low as 20 pairs. For ODM (custom mold) projects, we typically require an initial commitment of 50-100 rims to amortize the mold costs effectively.', '', '', 0, 0, 110),
    ('company-oem-odm', 'general', 'Do you offer design services for custom molds?', 'Yes, our in-house R&D team provides comprehensive design services, including 3D modeling, aerodynamic layout analysis, and graphic design for decals. We can work from a simple sketch or a detailed CAD file.', '', '', 0, 0, 120),
    ('company-oem-odm', 'general', 'What is the typical lead time for production?', 'For standard OEM orders, lead time is typically 25-35 days. For ODM projects involving new mold creation, the timeline is usually: 15 days for design confirmation, 25 days for mold opening, and 10 days for prototyping/testing.', '', '', 0, 0, 130),
    ('company-oem-odm', 'general', 'Is my design confidential?', 'Absolutely. We sign strict Non-Disclosure Agreements (NDAs) with all our ODM partners. Your private molds and layup schedules are exclusive to your brand and will never be shared with or sold to other clients.', '', '', 0, 0, 140),
    ('company-oem-odm', 'general', 'Do OEM/ODM wheels come with a warranty?', 'Yes, we stand behind our manufacturing quality. We offer a standard 2-year warranty on all OEM/ODM rims against manufacturing defects, with options to extend coverage based on specific partnership agreements.', '', '', 0, 0, 150),
    ('company-ourstory', 'brand', 'Where is Tanzanite based?', 'Tanzanite is a brand of Top Sports Co., Limited, with our Global Headquarters in Hong Kong and our state-of-the-art Manufacturing & R&D Base in Xiamen, China.', '', '', 0, 0, 110),
    ('company-ourstory', 'brand', 'What is Tanzanite’s core mission?', 'Our mission is to democratize high-performance cycling components by leveraging advanced manufacturing and direct-to-consumer efficiency, delivering premium carbon wheels without the traditional markup.', '', '', 0, 0, 120),
    ('company-ourstory', 'brand', 'How does Tanzanite approach sustainability?', 'We are committed to sustainable manufacturing practices, minimizing waste in our carbon layup process, and designing durable products that stand the test of time, reducing the need for frequent replacements.', '', '', 0, 0, 130),
    ('company-ourstory', 'products', 'What drives your product design?', 'We believe in data-driven design backed by rigorous testing. Every rim profile and layup schedule is optimized for the specific demands of its discipline, whether it’s aerodynamics for road or impact resistance for MTB.', '', '', 0, 0, 210),
    ('guides-tireguides', 'sizing', 'How do I read tire size and match an inner tube?', 'Check the size printed on your tire sidewall (e.g., 700x28C or 29x2.2, plus the ETRTO number). Choose an inner tube whose width range covers your tire size and with the correct diameter.', '', '', 0, 0, 110),
    ('guides-tireguides', 'sizing', 'Which valve type and length should I pick?', 'First, match the valve type to your rim hole (Presta/SV for road/MTB, Schrader/AV for wider holes).
          <br><br>
          <strong>Rule of thumb for length:</strong> The valve should extend at least 15mm above the rim.
          <ul>
            <li><strong>Low profile (≤30mm)</strong>: Standard 40mm valves.</li>
            <li><strong>Mid profile (30-45mm)</strong>: 60mm valves.</li>
            <li><strong>Deep profile (50-65mm)</strong>: 80mm valves.</li>
            <li><strong>Ultra deep (>70mm)</strong>: Use an 80mm valve with an extender.</li>
          </ul>', '', '', 0, 0, 120),
    ('guides-tireguides', 'sizing', 'When do I need a valve extender and which type?', 'If your rims are deeper than 50-60mm, standard valves might be too short to pump.
          <ul>
            <li><strong>Internal (Core Removable)</strong>: Best choice. You remove the valve core, screw the extender into the valve shaft, and put the core back on top. Allows for easier pumping and pressure adjustments.</li>
            <li><strong>External</strong>: Screws onto the valve tip. Easier to install but requires the valve to remain open, which can sometimes leak air.</li>
          </ul>', '', '', 0, 0, 130),
    ('guides-tireguides', 'installation', 'Tips for seating difficult tubeless tires?', 'Seating tubeless tires can be tricky. Try these steps:
          <ol>
            <li><strong>Use soapy water</strong>: Apply generously to the tire beads and rim hook to help them slide into place.</li>
            <li><strong>Remove valve core</strong>: This allows air to enter much faster, creating the sudden pressure needed to "pop" the beads into place.</li>
            <li><strong>Massage the tire</strong>: Ensure the tire beads are sitting in the center channel (the lowest part) of the rim before inflating.</li>
            <li><strong>Use a booster</strong>: If a floor pump fails, use a compressor or a tubeless booster canister.</li>
          </ol>', '', '', 0, 0, 210),
    ('guides-tireguides', 'installation', 'How should I set tire pressure for road vs gravel?', 'Start from the manufacturer’s recommended range and adjust for rider weight and terrain. Road tires typically run higher pressure for speed; gravel and wider tires use lower pressure for comfort and grip. Avoid exceeding max pressure printed on the tire.', '', '', 0, 0, 220),
    ('guides-tireguides', 'installation', 'Can I run tubeless on any tire and rim?', 'Use tubeless-ready tires and rims with proper tape and valves. Non-tubeless components may not seal reliably and can burp air. Always check the rim/tire manufacturer’s tubeless compatibility.', '', '', 0, 0, 230),
    ('guides-tireguides', 'maintenance', 'How can I reduce pinch flats with tubes?', 'Ensure correct tube size, keep pressure within the recommended range, and avoid trapping the tube during installation. Check rim tape for damage and inspect the tire for debris before mounting.', '', '', 0, 0, 310),
    ('guides-tireguides', 'maintenance', 'How should I store spare tubes?', 'Keep tubes in a cool, dry place away from direct sunlight and ozone sources. Avoid sharp folds; a loose roll or small pouch helps prevent creases that can weaken rubber over time.', '', '', 0, 0, 320),
    ('guides-wheelset-buyers', 'choosing', 'How do I choose the right wheelset for my riding style?', 'Consider these factors when choosing a wheelset:
            <ul>
              <li><strong>Riding discipline</strong> - Road, gravel, MTB, or mixed use</li>
              <li><strong>Rim depth</strong> - Deeper rims for aero, shallower for climbing</li>
              <li><strong>Rim width</strong> - Wider for better tire support and comfort</li>
              <li><strong>Weight</strong> - Lighter for climbing, durability for rough terrain</li>
              <li><strong>Brake type</strong> - Rim brake or disc brake compatibility</li>
            </ul>', '', '', 0, 0, 110),
    ('guides-wheelset-buyers', 'choosing', 'What is a mullet wheelset?', 'A mullet wheelset uses different wheel sizes front and rear - typically a 29" front wheel paired with a 27.5" rear wheel. This combination offers the rolling efficiency and obstacle clearance of a larger front wheel with the agility and acceleration of a smaller rear wheel. Popular for enduro and trail riding.', '', '', 0, 0, 120),
    ('guides-wheelset-buyers', 'choosing', 'Should I choose carbon or aluminum rims?', '<strong>Carbon rims</strong> offer better stiffness-to-weight ratio, aerodynamics, and can be shaped more precisely. They''re ideal for performance-focused riders.<br><br><strong>Aluminum rims</strong> are more affordable, easier to repair, and handle impacts well. They''re great for everyday riding and rough conditions.', '', '', 0, 0, 130),
    ('guides-wheelset-buyers', 'customization', 'Can I customize the appearance of my wheelset?', 'Yes! We offer several customization options:
            <ul>
              <li><strong>Laser engraving</strong> - Eco-friendly, precise designs in light gray</li>
              <li><strong>Waterslide decals</strong> - Full color graphics</li>
              <li><strong>Vinyl stickers</strong> - Removable, laser-cut options</li>
            </ul>
            Our graphics team will work with you to design the perfect look.', '', '', 0, 0, 210),
    ('guides-wheelset-buyers', 'customization', 'What is laser engraving and why choose it?', 'Laser engraving uses precision laser technology to etch designs directly into the carbon rim surface. Benefits include:
            <ul>
              <li>Eco-friendly - reduces plastic waste from stickers</li>
              <li>Permanent - won''t peel or fade</li>
              <li>Sleek appearance - subtle light gray finish</li>
              <li>No added weight</li>
            </ul>', '', '', 0, 0, 220),
    ('guides-wheelset-buyers', 'specs', 'What spoke count should I choose?', 'Spoke count affects strength, weight, and aerodynamics:
            <ul>
              <li><strong>20-24 spokes</strong> - Lighter, more aero, best for lighter riders and smooth roads</li>
              <li><strong>28-32 spokes</strong> - Stronger, more durable, better for heavier riders and rough terrain</li>
            </ul>
            We can help you choose the right spoke count based on your weight and riding style.', '', '', 0, 0, 310),
    ('guides-wheelset-buyers', 'specs', 'What hub options are available?', 'We offer various hub options to match your needs:
            <ul>
              <li>Different engagement points (36T, 54T, etc.)</li>
              <li>Various axle standards (QR, thru-axle)</li>
              <li>Multiple freehub body options (Shimano, SRAM XD, Campagnolo)</li>
            </ul>
            Contact us for specific hub recommendations.', '', '', 0, 0, 320),
    ('products-spoke-calculator', 'usage', 'How do I use this calculator?', 'To calculate the correct spoke length:
            <ol>
              <li><strong>Select your Rim</strong>: Choose from our product list or enter ERD manually.</li>
              <li><strong>Select your Hub</strong>: Choose from our product list or enter flange dimensions manually.</li>
              <li><strong>Configure setup</strong>: Set spoke count, lacing pattern (e.g., 3-cross), and nipple type.</li>
              <li><strong>Calculate</strong>: Click the button to get the recommended spoke lengths for left and right sides.</li>
            </ol>', '', '', 0, 0, 110),
    ('products-spoke-calculator', 'usage', 'What is ERD and why is it important?', 'ERD (Effective Rim Diameter) is the diameter of the rim at the point where the spoke nipples seat. It is the most critical dimension for spoke calculation. An incorrect ERD is the most common cause of wrong spoke lengths. We recommend measuring your specific rim''s ERD yourself using two cut spokes and nipples to be absolutely sure along with the manufacturer''s spec.', '', '', 0, 0, 120),
    ('products-spoke-calculator', 'usage', 'How accurate are the results?', 'The results are mathematically precise based on the inputs provided. However, real-world variations in rim roundness, hub manufacturing tolerances, and nipple dimensions mean the calculated length is a theoretical ideal. We recommend rounding to the nearest available even millimeter length.', '', '', 0, 0, 130),
    ('products-spoke-calculator', 'troubleshooting', 'Why are my calculated lengths different from another calculator?', 'Different calculators might use slightly different formulas or assumptions about spoke stretch. Our calculator uses standard trigonometric formulas including compensation for spoke hole offset (if applicable). Small differences of +/- 1mm are normal.', '', '', 0, 0, 210),
    ('products-spoke-calculator', 'troubleshooting', 'What if my hub is not in the list?', 'If your hub is not in our dropdown list, you can manually enter the flange dimensions. You will need: Left/Right Flange Distance (center to flange), and Left/Right Flange Diameter (PCD). Consult your hub manufacturer''s technical manual for these specs.', '', '', 0, 0, 220),
    ('support-payment', 'payment-methods', 'What payment methods do you accept?', 'We accept a variety of payment methods including:
            <ul>
              <li><strong>Credit & Debit Cards</strong> - Visa, MasterCard, and other major cards</li>
              <li><strong>PayPal</strong> - For faster checkout with buyer protection</li>
              <li><strong>Stripe</strong> - Secure card processing</li>
              <li><strong>WeChat Pay & Alipay</strong> - For customers in China</li>
              <li><strong>Bank Transfer</strong> - For larger or custom orders (by request)</li>
              <li><strong>WorldFirst</strong> - For selected international orders</li>
            </ul>', '', '', 0, 0, 110),
    ('support-payment', 'payment-methods', 'Are there any payment processing fees?', 'Yes, different payment methods have different processing fees:
            <ul>
              <li><strong>PayPal & Credit Cards</strong>: 3.5% of the order amount</li>
              <li><strong>Stripe</strong>: 3.5% of the order amount</li>
              <li><strong>WeChat Pay</strong>: 0.6% of the order amount</li>
              <li><strong>Alipay</strong>: 1% of the order amount</li>
              <li><strong>WorldFirst</strong>: 1% of the order amount</li>
              <li><strong>Bank Transfer</strong>: $45 USD bank fee (your bank may charge additional fees)</li>
            </ul>', '', '', 0, 0, 120),
    ('support-payment', 'payment-methods', 'Can I pay via bank transfer?', 'Yes, for larger or custom orders we can arrange payment via bank transfer. Please contact our support team before placing the order so we can provide the correct account details and reserve your items. Note that bank transfers incur a $45 USD bank fee, and your bank may charge additional fees.', '', '', 0, 0, 130),
    ('support-payment', 'security', 'Is my payment information secure?', 'Absolutely. All pages on our site use HTTPS with SSL certificates issued by trusted certificate authorities. Your payment information is transmitted over encrypted channels. We use reputable payment providers (PayPal, Stripe, Alipay, WeChat Pay) to process payments - we only receive the payment result and <strong>never store</strong> your card number, CVV, or other sensitive data.', '', '', 0, 0, 210),
    ('support-payment', 'security', 'What is 3D Secure verification?', '3D Secure is an additional security layer for online card payments. Depending on your card issuer and region, you may be asked to verify your identity through your bank''s app or a one-time code sent to your phone. This helps protect against unauthorized use of your card.', '', '', 0, 0, 220),
    ('support-payment', 'billing', 'When will my card be charged?', 'For card and wallet payments, an authorization is created when you confirm the order. The final capture normally happens when the order is accepted and prepared for shipment. For bank transfers, we start processing your order after the funds have been received and matched to your order reference.', '', '', 0, 0, 310),
    ('support-payment', 'billing', 'Why is the charged amount different from the displayed price?', 'Product prices are usually shown in a reference currency (e.g., USD). Your bank or payment provider may convert this amount into your local currency using their own exchange rate and may add conversion fees. Additionally, depending on your shipping country, local VAT or import duties may apply.', '', '', 0, 0, 320),
    ('support-payment', 'billing', 'What about taxes and import duties?', 'Depending on your shipping country, local VAT or import duties may apply. These charges are handled according to the shipping option shown at checkout. Please review the order summary carefully before confirming payment.', '', '', 0, 0, 330),
    ('support-payment', 'troubleshooting', 'My payment was declined. What should I do?', 'If your card or wallet is declined, please:
            <ul>
              <li>Check that your billing details match your bank records</li>
              <li>Ensure sufficient funds are available</li>
              <li>Try again or use a different payment method</li>
            </ul>
            Your bank can often provide more detail via their app or customer service.', '', '', 0, 0, 410),
    ('support-payment', 'troubleshooting', 'I was charged but my order was not created. What happened?', 'On rare occasions, a network issue can interrupt the redirect back to our site after payment. If you see a charge but no order in your account, please contact our support team with your payment reference so we can investigate and resolve the issue.', '', '', 0, 0, 420),
    ('support-payment', 'troubleshooting', 'I see multiple pending charges on my card. Is this normal?', 'If you attempted payment several times, your bank may show multiple pending authorizations. Normally, any unused authorizations are released automatically after a short period (usually 3-7 business days). If this does not happen, please contact your bank or our support team.', '', '', 0, 0, 430),
    ('support-product-feedback', 'feedback', 'How can I submit product feedback?', 'You can submit feedback through several channels:
            <ul>
              <li>Use the feedback form on this page</li>
              <li>Email our support team directly</li>
              <li>Leave a review on your order</li>
              <li>Contact us via WhatsApp</li>
            </ul>
            We read and consider all feedback to improve our products.', '', '', 0, 0, 110),
    ('support-product-feedback', 'feedback', 'What kind of feedback are you looking for?', 'We welcome all types of feedback including:
            <ul>
              <li>Product performance and durability</li>
              <li>Build quality and finish</li>
              <li>Suggestions for new products or features</li>
              <li>Issues or problems you''ve encountered</li>
              <li>Comparison with other products</li>
            </ul>', '', '', 0, 0, 120),
    ('support-product-feedback', 'feedback', 'Will I receive a response to my feedback?', 'We read all feedback but may not respond to every submission individually. If your feedback requires follow-up or contains a specific question, our team will reach out to you via email.', '', '', 0, 0, 130),
    ('support-product-feedback', 'reviews', 'How can I leave a product review?', 'After receiving your order, you can leave a review by:
            <ul>
              <li>Clicking the review link in your order confirmation email</li>
              <li>Logging into your account and visiting your order history</li>
              <li>Visiting the product page and clicking "Write a Review"</li>
            </ul>', '', '', 0, 0, 210),
    ('support-product-feedback', 'reviews', 'Can I edit or delete my review?', 'Yes, you can edit or delete your review by logging into your account and visiting your order history. Find the order containing the reviewed product and click "Edit Review" or "Delete Review".', '', '', 0, 0, 220),
    ('support-shipping', 'delivery', 'How long does shipping take?', 'Shipping times vary by destination:
            <ul>
              <li><strong>Domestic (China)</strong>: 3-5 business days</li>
              <li><strong>Asia Pacific</strong>: 7-14 business days</li>
              <li><strong>Europe</strong>: 10-20 business days</li>
              <li><strong>North America</strong>: 10-20 business days</li>
              <li><strong>Other regions</strong>: 14-30 business days</li>
            </ul>
            Note: Custom-built wheelsets may require additional 3-7 days for assembly before shipping.', '', '', 0, 0, 110),
    ('support-shipping', 'delivery', 'What is the order processing time?', 'Order processing typically takes 1-3 business days for in-stock items. For custom wheel builds, please allow an additional 3-7 business days for assembly and quality inspection before shipping.', '', '', 0, 0, 120),
    ('support-shipping', 'delivery', 'Can I get express shipping?', 'Yes, we offer express shipping options for most destinations. Express shipping typically reduces delivery time by 50% or more. Please contact our support team for express shipping quotes and availability for your location.', '', '', 0, 0, 130),
    ('support-shipping', 'tracking', 'How can I track my order?', 'Once your order ships, you will receive an email with a tracking number and link. You can also log into your account on our website to view the latest shipping status. For any tracking issues, please contact our support team.', '', '', 0, 0, 210),
    ('support-shipping', 'tracking', 'My tracking hasn''t updated in several days. Is this normal?', 'Yes, this can be normal, especially for international shipments. Tracking updates may pause during customs clearance or when packages transfer between carriers. If there''s no update for more than 7 days, please contact our support team.', '', '', 0, 0, 220),
    ('support-shipping', 'international', 'Do you ship internationally?', 'Yes, we ship to most countries worldwide. Shipping costs and delivery times vary by destination. Some remote areas may have limited shipping options or longer delivery times.', '', '', 0, 0, 310),
    ('support-shipping', 'international', 'Will I have to pay customs duties or import taxes?', 'Depending on your country''s regulations, you may be required to pay customs duties, import taxes, or VAT upon delivery. These charges are determined by your local customs authority and are the responsibility of the recipient. We recommend checking with your local customs office for more information.', '', '', 0, 0, 320),
    ('support-shipping', 'international', 'What documents are included with international shipments?', 'All international shipments include a commercial invoice and packing list. For certain destinations, we may also include a certificate of origin or other required documentation. If you need specific documents for customs clearance, please let us know before shipping.', '', '', 0, 0, 330),
    ('support-shipping', 'issues', 'What if my package is damaged during shipping?', 'If your package arrives damaged, please take photos of the packaging and contents immediately. Contact our support team within 48 hours of delivery with the photos and your order number. We will work with the carrier to file a claim and arrange a replacement or refund.', '', '', 0, 0, 410),
    ('support-shipping', 'issues', 'What if my package is lost?', 'If your tracking shows no updates for an extended period or indicates the package is lost, please contact our support team. We will investigate with the carrier and, if the package cannot be located, arrange a replacement shipment or refund.', '', '', 0, 0, 420),
    ('support-test-report', 'testing', 'What tests do you perform on your products?', 'Our products undergo rigorous testing including:
            <ul>
              <li><strong>Impact testing</strong> - Simulating real-world impacts</li>
              <li><strong>Fatigue testing</strong> - Thousands of cycles to ensure durability</li>
              <li><strong>Brake heat testing</strong> - For rim brake compatibility</li>
              <li><strong>Spoke tension testing</strong> - Ensuring consistent wheel build quality</li>
              <li><strong>Weight verification</strong> - Confirming advertised specifications</li>
            </ul>', '', '', 0, 0, 110),
    ('support-test-report', 'testing', 'Are your products tested by third parties?', 'Yes, we work with independent testing laboratories to verify our products meet international standards. Third-party test reports are available for many of our products upon request.', '', '', 0, 0, 120),
    ('support-test-report', 'testing', 'What standards do your products meet?', 'Our products are designed and tested to meet or exceed relevant industry standards including UCI regulations for competitive cycling and ISO standards for bicycle components.', '', '', 0, 0, 130),
    ('support-test-report', 'reports', 'How can I get a test report for a specific product?', 'Test reports for specific products can be requested by contacting our support team. Please provide the product name or SKU, and we will share available documentation. Some reports are also available for download on individual product pages.', '', '', 0, 0, 210),
    ('support-test-report', 'reports', 'Are test reports available in multiple languages?', 'Most test reports are available in English. For major markets, we may have translated versions available. Please contact our support team to inquire about specific language availability.', '', '', 0, 0, 220),
    ('support-test-report', 'quality', 'How do you ensure consistent quality?', 'We maintain strict quality control throughout our manufacturing process:
            <ul>
              <li>Incoming material inspection</li>
              <li>In-process quality checks at each production stage</li>
              <li>Final inspection before packaging</li>
              <li>Random sampling for destructive testing</li>
            </ul>', '', '', 0, 0, 310),
    ('support-test-report', 'quality', 'What if I receive a defective product?', 'If you receive a product that doesn''t meet our quality standards, please contact our support team immediately. We will arrange for a replacement or refund. Defective products are covered under our warranty policy.', '', '', 0, 0, 320),
    ('support-test-report', 'assembly', 'What is the recommended spoke tension?', 'We recommend a spoke tension range of <strong>100–135 kgf</strong>. Tension should be as uniform as possible, with variations on the same side kept within ±5%.', '', '', 0, 0, 410),
    ('support-test-report', 'assembly', 'What tools do I need for wheel assembly?', 'Essential tools include a <strong>truing stand</strong>, a <strong>spoke tension meter</strong>, and a suitable <strong>spoke nipple wrench</strong>. We also use custom-developed tools in our factory for maximum precision.', '', '', 0, 0, 420),
    ('support-test-report', 'assembly', 'Should I use threadlocker on spoke nipples?', 'Yes, we recommend applying a <strong>low-strength threadlocker</strong> (like Loctite 222) to the spoke threads during the final truing stage. This prevents nipples from loosening over time due to vibration.', '', '', 0, 0, 430),
    ('support-test-report', 'assembly', 'What are the standard build tolerances?', 'Our standard tolerances are strict: <strong>Lateral/Radial runout ≤ 0.2mm</strong>, and <strong>Center (Dish) offset ≤ 0.5mm</strong>.', '', '', 0, 0, 440),
    ('support-warranty-check', 'how-to-check', 'Where can I find my product code?', 'Your product code can be found in several places:
            <ul>
              <li><strong>Product packaging</strong>: On the original box or packaging</li>
              <li><strong>Warranty card</strong>: Included with your purchase</li>
              <li><strong>Product label</strong>: On the rim or hub sticker</li>
              <li><strong>Order confirmation email</strong>: Listed in your order details</li>
            </ul>
            The code is typically a combination of letters and numbers (e.g., TZ-2024-ABC123).', '', '', 0, 0, 110),
    ('support-warranty-check', 'how-to-check', 'Why do I need to log in to check warranty?', 'Logging in helps us:
            <ul>
              <li>Verify your ownership of the product</li>
              <li>Link warranty information to your account</li>
              <li>Provide personalized support if needed</li>
              <li>Keep your warranty history in one place</li>
            </ul>
            Your account information is kept secure and private.', '', '', 0, 0, 120),
    ('support-warranty-check', 'how-to-check', 'What information will I see after checking?', 'After a successful warranty check, you''ll see:
            <ul>
              <li><strong>Product details</strong>: Name, type, and specifications</li>
              <li><strong>Warranty status</strong>: Valid or expired</li>
              <li><strong>Ship date</strong>: When the product was shipped</li>
              <li><strong>Warranty period</strong>: Length of coverage</li>
              <li><strong>Expiration date</strong>: When warranty ends</li>
              <li><strong>Remaining time</strong>: Days/months left on warranty</li>
            </ul>', '', '', 0, 0, 130),
    ('support-warranty-check', 'troubleshooting', 'What if my product code is not found?', 'If your product code is not found, please check:
            <ul>
              <li>The code is entered correctly (no typos)</li>
              <li>You''re using the correct format</li>
              <li>The product is a genuine Tanzanite product</li>
            </ul>
            If the issue persists, please contact our support team with your order details.', '', '', 0, 0, 210),
    ('support-warranty-check', 'troubleshooting', 'My warranty shows as expired but I just bought it. What should I do?', 'If your warranty appears expired incorrectly:
            <ol>
              <li>Check your purchase date against the displayed ship date</li>
              <li>Verify the warranty period for your product type</li>
              <li>Contact support with your order confirmation and receipt</li>
            </ol>
            We''ll investigate and correct any discrepancies in our system.', '', '', 0, 0, 220),
    ('support-warranty-check', 'troubleshooting', 'Can I transfer warranty to a new owner?', 'Warranty transfer policies:
            <ul>
              <li>Warranty is generally tied to the original purchaser</li>
              <li>For second-hand purchases, contact us with proof of original purchase</li>
              <li>Transfer may be possible with proper documentation</li>
            </ul>
            Please contact our support team for warranty transfer requests.', '', '', 0, 0, 230),
    ('support-warranty', 'policy', 'What is the warranty period for Tanzanite wheels?', 'We offer two coverage options for all Tanzanite series Wheels/Rims:
            <ul>
              <li><strong>Standard Warranty</strong>: 5 Years from the date of purchase (included by default).</li>
              <li><strong>Lifetime Warranty</strong>: Available as an upgrade for USD $100 per rim.</li>
            </ul>', '', '', 0, 0, 110),
    ('support-warranty', 'policy', 'What does the warranty cover?', 'Our warranty covers <strong>manufacturing defects</strong> in materials and workmanship. If a structural failure occurs due to a defect within the warranty period, we will replace the rim.', '', '', 0, 0, 120),
    ('support-warranty', 'policy', 'Is the warranty transferable?', 'No, the warranty applies only to the <strong>original purchaser</strong> and is non-transferable. Valid proof of purchase is required for all claims.', '', '', 0, 0, 130),
    ('support-warranty', 'policy', 'What isn''t covered by the warranty?', 'The warranty does not cover damage caused by <strong>improper assembly</strong>, use of incompatible parts, unauthorized modifications, or normal wear and tear. Accidental damage (crashes) is covered under our Crash Replacement Policy.', '', '', 0, 0, 140),
    ('support-warranty', 'policy', 'What happens if my rim is discontinued?', 'If a warranty replacement is approved but your specific rim model is discontinued, we will upgrade you to the <strong>latest equivalent model</strong> at no additional cost.', '', '', 0, 0, 150),
    ('support-warranty', 'crash', 'Do you offer crash replacement?', 'Yes, we offer a <strong>Crash Replacement Policy</strong> for accidental damage (e.g., crashes, jumps, rock impacts) that is not covered under the standard warranty.', '', '', 0, 0, 210),
    ('support-warranty', 'crash', 'What are the terms for crash replacement?', 'This coverage is valid for <strong>3 years</strong> from the date of purchase. You can receive a replacement rim at a <strong>10% discount</strong>.', '', '', 0, 0, 220),
    ('support-warranty', 'claims', 'How do I submit a warranty claim?', 'You can submit a claim directly on this page under the <strong>"Submit Warranty"</strong> tab. Please have your order number, photos, and a description of the issue ready.', '', '', 0, 0, 310),
    ('support-warranty', 'claims', 'Who covers shipping costs?', 'Shipping responsibility depends on when the claim is made:
            <ul>
              <li><strong>Within 30 days</strong> of receipt: Tanzanite covers shipping costs.</li>
              <li><strong>After 30 days</strong>: The customer is responsible for shipping costs.</li>
            </ul>', '', '', 0, 0, 320),
    ('support-warranty', 'claims', 'Do I need to return the damaged product?', 'In most cases, <strong>no</strong>. We typically require clear photos and videos of the damage. If we do need the item returned for inspection, we will issue a return authorization.', '', '', 0, 0, 330)
),
locales(locale) AS (
  VALUES ('en'), ('zh')
)
INSERT INTO faqs (
  page_id,
  category,
  locale,
  question,
  answer,
  answer_image_url,
  answer_image_alt,
  answer_image_width,
  answer_image_height,
  status,
  "order",
  created_at,
  updated_at
)
SELECT
  seed_faqs.page_id,
  seed_faqs.category,
  locales.locale,
  seed_faqs.question,
  seed_faqs.answer,
  seed_faqs.answer_image_url,
  seed_faqs.answer_image_alt,
  seed_faqs.answer_image_width,
  seed_faqs.answer_image_height,
  'published',
  seed_faqs.sort_order,
  NOW(),
  NOW()
FROM seed_faqs
CROSS JOIN locales
WHERE NOT EXISTS (
  SELECT 1
  FROM faqs existing
  WHERE existing.deleted_at IS NULL
    AND existing.page_id = seed_faqs.page_id
    AND existing.category = seed_faqs.category
    AND existing.locale = locales.locale
    AND existing.question = seed_faqs.question
);
