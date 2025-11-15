/**
 * Tanzanite Settings - Products List Actions
 * 单个商品操作模块，包括删除、置顶、复制等
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    window.TzProductsList = window.TzProductsList || {};
    
    window.TzProductsList.initActions = function() {
        const config = window.TzProductsList.config;
        const i18n = window.TzProductsList.i18n;
        const showNotice = window.TzProductsList.showNotice;
        const loadProducts = window.TzProductsList.loadProducts;
        const elements = window.TzProductsList.elements;

        /**
         * 使用事件委托处理所有操作按钮
         */
        if (elements.tableBody) {
            elements.tableBody.addEventListener('click', handleTableAction);
        }

        if (elements.cardsWrapper) {
            elements.cardsWrapper.addEventListener('click', handleCardAction);
        }

        /**
         * 处理表格中的操作
         */
        function handleTableAction(e) {
            const target = e.target;

            // 删除商品
            if (target.classList.contains('tz-product-delete')) {
                e.preventDefault();
                const id = parseInt(target.dataset.id, 10);
                if (id) deleteProduct(id);
            }

            // 切换置顶
            if (target.classList.contains('tz-toggle-sticky')) {
                e.preventDefault();
                const id = parseInt(target.dataset.id, 10);
                const isSticky = target.dataset.sticky === '1';
                if (id) toggleSticky(id, !isSticky);
            }

            // 复制 Payload
            if (target.classList.contains('tz-copy-payload')) {
                e.preventDefault();
                const id = parseInt(target.dataset.id, 10);
                if (id) copyPayload(id);
            }
        }

        /**
         * 处理卡片中的操作
         */
        function handleCardAction(e) {
            const target = e.target;

            // 删除商品
            if (target.classList.contains('tz-product-delete')) {
                e.preventDefault();
                const id = parseInt(target.dataset.id, 10);
                if (id) deleteProduct(id);
            }
        }

        /**
         * 删除商品
         */
        async function deleteProduct(id) {
            if (!confirm(i18n.deleteConfirm || '确认删除此商品吗？')) {
                return;
            }

            try {
                const resp = await fetch(config.singleUrl + id, {
                    method: 'DELETE',
                    headers: {
                        'X-WP-Nonce': config.nonce
                    }
                });

                const data = await resp.json();

                if (!resp.ok) {
                    showNotice('error', data.message || i18n.deleteFailed, data.hint, data.actions);
                    return;
                }

                showNotice('success', data.message || i18n.deleteSuccess);
                loadProducts();

            } catch (err) {
                showNotice('error', err.message);
            }
        }

        /**
         * 切换置顶状态
         */
        async function toggleSticky(id, sticky) {
            try {
                const resp = await fetch(config.singleUrl + id, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-WP-Nonce': config.nonce
                    },
                    body: JSON.stringify({
                        sticky: sticky
                    })
                });

                const data = await resp.json();

                if (!resp.ok) {
                    showNotice('error', data.message || i18n.stickyFailed, data.hint, data.actions);
                    return;
                }

                showNotice('success', data.message || i18n.stickySuccess);
                loadProducts();

            } catch (err) {
                showNotice('error', err.message);
            }
        }

        /**
         * 复制商品 Payload
         */
        async function copyPayload(id) {
            try {
                const resp = await fetch(config.singleUrl + id, {
                    headers: {
                        'X-WP-Nonce': config.nonce
                    }
                });

                const data = await resp.json();

                if (!resp.ok) {
                    showNotice('error', data.message || i18n.copyFailed);
                    return;
                }

                // 复制到剪贴板
                const payload = JSON.stringify(data, null, 2);
                
                if (navigator.clipboard) {
                    await navigator.clipboard.writeText(payload);
                    showNotice('success', i18n.copySuccess || '已复制到剪贴板');
                } else {
                    // 降级方案
                    const textarea = document.createElement('textarea');
                    textarea.value = payload;
                    textarea.style.position = 'fixed';
                    textarea.style.opacity = '0';
                    document.body.appendChild(textarea);
                    textarea.select();
                    document.execCommand('copy');
                    document.body.removeChild(textarea);
                    showNotice('success', i18n.copySuccess || '已复制到剪贴板');
                }

            } catch (err) {
                showNotice('error', err.message);
            }
        }
    };
})();
