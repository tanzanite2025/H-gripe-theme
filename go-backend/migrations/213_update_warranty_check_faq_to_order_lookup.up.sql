-- Keep the deployed warranty FAQ aligned with the order-number lookup flow.
-- Warranty lookup is based on shipped orders; product identifiers are optional
-- staff evidence and must not be presented as a customer lookup key.

UPDATE faqs
SET question = CASE
        WHEN locale IN ('zh', 'zh_cn') THEN '在哪里可以找到我的订单号？'
        ELSE 'Where can I find my order number?'
    END,
    answer = CASE
        WHEN locale IN ('zh', 'zh_cn') THEN
            '订单号通常可以在以下位置找到：
            <ul>
              <li><strong>订单确认邮件</strong>：订单详情中会列出订单号</li>
              <li><strong>账户订单</strong>：登录后打开订单历史</li>
              <li><strong>客服沟通记录</strong>：与该订单相关的消息中会包含订单号</li>
            </ul>
            请按订单中显示的内容准确输入订单号，例如 TZ202608230001。'
        ELSE
            'Your order number can be found in several places:
            <ul>
              <li><strong>Order confirmation email</strong>: Listed in your order details</li>
              <li><strong>Account orders</strong>: Open your order history after signing in</li>
              <li><strong>Support correspondence</strong>: Included in messages about the order</li>
            </ul>
            Enter the order number exactly as shown, for example TZ202608230001.'
    END,
    updated_at = NOW()
WHERE page_id = 'support-warranty-check'
  AND category = 'how-to-check'
  AND locale IN ('en', 'zh', 'zh_cn')
  AND question = 'Where can I find my product code?'
  AND deleted_at IS NULL;

UPDATE faqs
SET question = CASE
        WHEN locale IN ('zh', 'zh_cn') THEN '如果找不到订单号怎么办？'
        ELSE 'What if my order number is not found?'
    END,
    answer = CASE
        WHEN locale IN ('zh', 'zh_cn') THEN
            '如果找不到订单号，请确认：
            <ul>
              <li>输入的订单号与订单中显示的内容完全一致</li>
              <li>订单已经发货；保修查询仅适用于已发货订单</li>
              <li>使用的是正确店铺账户中的订单号</li>
            </ul>
            如果问题仍然存在，请携带订单信息联系客服。'
        ELSE
            'If your order number is not found, please check:
            <ul>
              <li>The order number is entered exactly as shown (no typos)</li>
              <li>The order has already been shipped; warranty lookup is available for shipped orders</li>
              <li>You''re using the order number from the correct store account</li>
            </ul>
            If the issue persists, please contact our support team with your order details.'
    END,
    updated_at = NOW()
WHERE page_id = 'support-warranty-check'
  AND category = 'troubleshooting'
  AND locale IN ('en', 'zh', 'zh_cn')
  AND question = 'What if my product code is not found?'
  AND deleted_at IS NULL;
