/**
 * Tanzanite Settings - 订单列表页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzOrdersListConfig === 'undefined') {
            console.error('Orders List config not found');
            return;
        }

        const config = TzOrdersListConfig;
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest, formatDate } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            notice: document.getElementById('tz-orders-notice'),
            filtersForm: document.getElementById('tz-orders-filters'),
            filterSubmit: document.getElementById('tz-orders-filter-submit'),
            filterReset: document.getElementById('tz-orders-filter-reset'),
            tableBody: document.querySelector('#tz-orders-table tbody'),
            pagination: document.querySelector('.tz-orders-pagination')
        };

        // 检查必需元素
        if (!elements.tableBody) {
            console.error('Required elements not found');
            return;
        }

        let currentPage = 1;
        let totalPages = 1;

        /**
         * 渲染订单列表
         */
        function renderOrders(orders) {
            elements.tableBody.innerHTML = '';

            if (!orders || orders.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="8" style="text-align:center;">暂无订单</td>';
                elements.tableBody.appendChild(tr);
                return;
            }

            orders.forEach(order => {
                const tr = document.createElement('tr');
                
                const statusColor = {
                    'pending': '#f59e0b',
                    'paid': '#3b82f6',
                    'processing': '#8b5cf6',
                    'shipped': '#06b6d4',
                    'completed': '#10b981',
                    'cancelled': '#ef4444'
                }[order.status] || '#6b7280';

                tr.innerHTML = `
                    <td>
                        <strong>${order.order_number}</strong><br>
                        <small>ID: ${order.id}</small>
                    </td>
                    <td>
                        ${order.customer_name || '-'}<br>
                        <small>${order.customer_phone || '-'}</small>
                    </td>
                    <td>
                        <strong>¥${order.total_amount || 0}</strong><br>
                        <small>商品: ¥${order.subtotal || 0}</small>
                    </td>
                    <td>
                        <span style="color:${statusColor};font-weight:600;">
                            ${config.statusLabels[order.status] || order.status}
                        </span>
                    </td>
                    <td>
                        ${order.channel || '-'}<br>
                        <small>${order.payment_method || '-'}</small>
                    </td>
                    <td>${formatDate(order.created_at)}</td>
                    <td>
                        ${order.tracking_provider || '-'}<br>
                        <small>${order.tracking_number || '-'}</small>
                    </td>
                    <td>
                        <a href="${config.detailBase}${order.id}" class="button-link">查看</a>
                        ${config.canManage ? ` | <button class="button-link sync-tracking" data-id="${order.id}">刷新物流</button>` : ''}
                    </td>
                `;
                
                elements.tableBody.appendChild(tr);
            });
        }

        /**
         * 渲染分页
         */
        function renderPagination(page, total) {
            if (!elements.pagination) return;

            const prevDisabled = page <= 1 ? 'disabled' : '';
            const nextDisabled = page >= total ? 'disabled' : '';

            const html = `
                <button class="button" id="tz-orders-prev" ${prevDisabled}>上一页</button>
                <span>第 ${page} / ${total} 页</span>
                <button class="button" id="tz-orders-next" ${nextDisabled}>下一页</button>
            `;

            elements.pagination.innerHTML = html;

            // 绑定事件
            const prevBtn = document.getElementById('tz-orders-prev');
            const nextBtn = document.getElementById('tz-orders-next');

            if (prevBtn) {
                prevBtn.addEventListener('click', () => loadOrders(currentPage - 1));
            }

            if (nextBtn) {
                nextBtn.addEventListener('click', () => loadOrders(currentPage + 1));
            }
        }

        /**
         * 获取筛选参数
         */
        function getFilters() {
            if (!elements.filtersForm) return {};

            const formData = new FormData(elements.filtersForm);
            const filters = {};

            for (const [key, value] of formData.entries()) {
                if (value) {
                    filters[key] = value;
                }
            }

            return filters;
        }

        /**
         * 加载订单列表
         */
        async function loadOrders(page = 1) {
            try {
                showNotice(elements.notice, 'info', '加载中...');

                const filters = getFilters();
                filters.page = page;
                filters.per_page = filters.per_page || 20;

                const queryString = new URLSearchParams(filters).toString();
                const url = config.listUrl + '?' + queryString;

                const result = await apiRequest(url);

                if (!result.ok) {
                    renderOrders([]);
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                currentPage = page;
                totalPages = result.data.total_pages || 1;

                renderOrders(result.data.items || []);
                renderPagination(currentPage, totalPages);
                showNotice(elements.notice, null);

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 刷新单个订单的物流
         */
        async function syncOrderTracking(orderId) {
            if (!config.canManage) return;

            try {
                const result = await apiRequest(config.syncBase + orderId + '/tracking', {
                    method: 'POST'
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '刷新失败');
                    return;
                }

                showNotice(elements.notice, 'success', '物流状态已刷新');
                loadOrders(currentPage);

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 重置筛选
         */
        function resetFilters() {
            if (elements.filtersForm) {
                elements.filtersForm.reset();
            }
            loadOrders(1);
        }

        // 事件监听
        if (elements.filterSubmit) {
            elements.filterSubmit.addEventListener('click', () => loadOrders(1));
        }

        if (elements.filterReset) {
            elements.filterReset.addEventListener('click', resetFilters);
        }

        if (elements.tableBody) {
            elements.tableBody.addEventListener('click', function(e) {
                if (e.target.classList.contains('sync-tracking')) {
                    e.preventDefault();
                    const orderId = parseInt(e.target.dataset.id);
                    syncOrderTracking(orderId);
                }
            });
        }

        // 初始化
        loadOrders(1);
    });
})();
