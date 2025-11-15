/**
 * Tanzanite Settings - 订单详情页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzOrderDetailConfig === 'undefined') {
            console.error('Order Detail config not found');
            return;
        }

        const config = TzOrderDetailConfig;
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest, formatDate } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            notice: document.getElementById('tz-order-detail-notice'),
            summaryMeta: document.getElementById('tz-order-summary-meta'),
            timelineList: document.getElementById('tz-order-timeline'),
            customerBox: document.getElementById('tz-order-customer'),
            amountsBox: document.getElementById('tz-order-amounts'),
            invoiceBox: document.getElementById('tz-order-invoice'),
            notesBox: document.getElementById('tz-order-notes'),
            itemsTable: document.querySelector('#tz-order-items tbody'),
            actionsBox: document.getElementById('tz-order-status-actions'),
            trackingSummary: document.getElementById('tz-order-tracking-summary'),
            trackingHint: document.getElementById('tz-order-tracking-hint'),
            trackingTable: document.querySelector('#tz-order-tracking-events tbody'),
            auditTable: document.querySelector('#tz-order-audit tbody'),
            refreshBtn: document.getElementById('tz-order-refresh-tracking'),
            shippingForm: document.getElementById('tz-order-shipping-form'),
            shippingProvider: document.getElementById('tz-order-shipping-provider'),
            shippingNumber: document.getElementById('tz-order-shipping-number'),
            shippingSubmit: document.getElementById('tz-order-shipping-submit'),
            shippingHint: document.getElementById('tz-order-shipping-hint')
        };

        let currentOrder = null;

        /**
         * 渲染订单摘要
         */
        function renderSummary(order) {
            if (!elements.summaryMeta) return;

            const html = `
                <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:16px;">
                    <div><strong>订单号:</strong> ${order.order_number}</div>
                    <div><strong>状态:</strong> ${config.statusLabels[order.status] || order.status}</div>
                    <div><strong>渠道:</strong> ${order.channel || '-'}</div>
                    <div><strong>支付方式:</strong> ${order.payment_method || '-'}</div>
                    <div><strong>创建时间:</strong> ${formatDate(order.created_at)}</div>
                </div>
            `;
            elements.summaryMeta.innerHTML = html;
        }

        /**
         * 渲染状态时间线
         */
        function renderTimeline(order) {
            if (!elements.timelineList) return;

            elements.timelineList.innerHTML = '';

            const events = [
                { label: '创建', time: order.created_at },
                { label: '支付', time: order.paid_at },
                { label: '发货', time: order.shipped_at },
                { label: '完成', time: order.completed_at }
            ];

            events.forEach(event => {
                if (event.time) {
                    const li = document.createElement('li');
                    li.innerHTML = `<strong>${event.label}:</strong> ${formatDate(event.time)}`;
                    elements.timelineList.appendChild(li);
                }
            });
        }

        /**
         * 渲染客户信息
         */
        function renderCustomer(order) {
            if (!elements.customerBox) return;

            const html = `
                <div style="display:grid;gap:8px;">
                    <div><strong>姓名:</strong> ${order.customer_name || '-'}</div>
                    <div><strong>电话:</strong> ${order.customer_phone || '-'}</div>
                    <div><strong>邮箱:</strong> ${order.customer_email || '-'}</div>
                    <div><strong>地址:</strong> ${order.shipping_address || '-'}</div>
                </div>
            `;
            elements.customerBox.innerHTML = html;
        }

        /**
         * 渲染金额明细
         */
        function renderAmounts(order) {
            if (!elements.amountsBox) return;

            const html = `
                <div style="display:grid;gap:8px;">
                    <div><strong>商品小计:</strong> ¥${order.subtotal || 0}</div>
                    <div><strong>运费:</strong> ¥${order.shipping_fee || 0}</div>
                    <div><strong>优惠:</strong> -¥${order.discount_amount || 0}</div>
                    <div><strong>积分抵扣:</strong> -¥${order.points_discount || 0}</div>
                    <div style="color:#16a34a;font-size:18px;"><strong>实付总额:</strong> ¥${order.total_amount || 0}</div>
                </div>
            `;
            elements.amountsBox.innerHTML = html;
        }

        /**
         * 渲染发票和备注
         */
        function renderInvoiceAndNotes(order) {
            if (elements.invoiceBox) {
                if (order.invoice_data) {
                    const invoice = order.invoice_data;
                    const html = `
                        <div style="display:grid;gap:8px;">
                            <div><strong>类型:</strong> ${invoice.type || '-'}</div>
                            <div><strong>抬头:</strong> ${invoice.title || '-'}</div>
                            <div><strong>税号:</strong> ${invoice.tax_number || '-'}</div>
                        </div>
                    `;
                    elements.invoiceBox.innerHTML = html;
                } else {
                    elements.invoiceBox.innerHTML = '<p class="description">无发票信息</p>';
                }
            }

            if (elements.notesBox) {
                const html = `
                    <div style="display:grid;gap:8px;">
                        <div><strong>客户备注:</strong> ${order.customer_note || '-'}</div>
                        <div><strong>客服备注:</strong> ${order.admin_note || '-'}</div>
                    </div>
                `;
                elements.notesBox.innerHTML = html;
            }
        }

        /**
         * 渲染商品明细
         */
        function renderItems(items) {
            if (!elements.itemsTable) return;

            elements.itemsTable.innerHTML = '';

            if (!items || items.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="7" style="text-align:center;">暂无商品</td>';
                elements.itemsTable.appendChild(tr);
                return;
            }

            items.forEach(item => {
                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>${item.product_title || '-'}</td>
                    <td>${item.sku_code || '-'}</td>
                    <td>¥${item.unit_price || 0}</td>
                    <td>${item.quantity || 0}</td>
                    <td>¥${item.discount_amount || 0}</td>
                    <td>¥${item.subtotal || 0}</td>
                    <td>${item.note || '-'}</td>
                `;
                elements.itemsTable.appendChild(tr);
            });
        }

        /**
         * 渲染状态操作按钮
         */
        function renderStatusActions(order) {
            if (!elements.actionsBox) return;

            elements.actionsBox.innerHTML = '';

            if (!config.canManage) {
                elements.actionsBox.innerHTML = '<p class="description">无权限操作</p>';
                return;
            }

            const currentStatus = order.status;
            const allowedTransitions = config.statusTransitions[currentStatus] || [];

            if (allowedTransitions.length === 0) {
                elements.actionsBox.innerHTML = '<p class="description">当前状态无可用操作</p>';
                return;
            }

            allowedTransitions.forEach(targetStatus => {
                const button = document.createElement('button');
                button.className = 'button';
                button.textContent = '切换至 ' + (config.statusLabels[targetStatus] || targetStatus);
                button.dataset.status = targetStatus;
                button.addEventListener('click', () => updateOrderStatus(targetStatus));
                elements.actionsBox.appendChild(button);
            });
        }

        /**
         * 渲染物流追踪
         */
        function renderTracking(order) {
            if (elements.trackingSummary) {
                if (order.tracking_provider && order.tracking_number) {
                    const html = `
                        <div style="display:grid;gap:8px;">
                            <div><strong>物流公司:</strong> ${order.tracking_provider}</div>
                            <div><strong>运单号:</strong> ${order.tracking_number}</div>
                            <div><strong>最近同步:</strong> ${order.tracking_synced_at ? formatDate(order.tracking_synced_at) : '-'}</div>
                        </div>
                    `;
                    elements.trackingSummary.innerHTML = html;
                } else {
                    elements.trackingSummary.innerHTML = '<p class="description">暂无物流信息</p>';
                }
            }

            if (elements.trackingTable) {
                elements.trackingTable.innerHTML = '';
                // 物流事件需要从另一个接口获取
            }
        }

        /**
         * 更新订单状态
         */
        async function updateOrderStatus(targetStatus) {
            if (!confirm('确定要更新订单状态吗？')) {
                return;
            }

            try {
                const result = await apiRequest(config.updateUrl, {
                    method: 'PUT',
                    body: JSON.stringify({ status: targetStatus })
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '更新失败');
                    return;
                }

                showNotice(elements.notice, 'success', '状态已更新');
                loadOrderDetail();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 提交发货信息
         */
        async function submitShipping(e) {
            e.preventDefault();

            const provider = elements.shippingProvider.value.trim();
            const number = elements.shippingNumber.value.trim();

            if (!provider || !number) {
                showNotice(elements.notice, 'warning', '请填写物流公司和运单号');
                return;
            }

            try {
                const result = await apiRequest(config.updateUrl, {
                    method: 'PUT',
                    body: JSON.stringify({
                        tracking_provider: provider,
                        tracking_number: number,
                        status: 'shipped'
                    })
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '发货失败');
                    return;
                }

                showNotice(elements.notice, 'success', '发货信息已更新');
                elements.shippingForm.reset();
                loadOrderDetail();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 刷新物流信息
         */
        async function refreshTracking() {
            if (!config.canManage) return;

            try {
                showNotice(elements.notice, 'info', '正在刷新物流...');

                const result = await apiRequest(config.syncUrl, {
                    method: 'POST'
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '刷新失败');
                    return;
                }

                showNotice(elements.notice, 'success', '物流状态已刷新');
                loadOrderDetail();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 加载订单详情
         */
        async function loadOrderDetail() {
            try {
                const result = await apiRequest(config.detailUrl);

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                currentOrder = result.data;

                // 保存到全局变量供打印功能使用
                window.currentOrderData = currentOrder;

                // 渲染各个部分
                renderSummary(currentOrder);
                renderTimeline(currentOrder);
                renderCustomer(currentOrder);
                renderAmounts(currentOrder);
                renderInvoiceAndNotes(currentOrder);
                renderItems(currentOrder.items || []);
                renderStatusActions(currentOrder);
                renderTracking(currentOrder);

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        // 事件监听
        if (elements.shippingForm) {
            elements.shippingForm.addEventListener('submit', submitShipping);
        }

        if (elements.refreshBtn) {
            elements.refreshBtn.addEventListener('click', refreshTracking);
        }

        // 初始化
        loadOrderDetail();
    });
})();
