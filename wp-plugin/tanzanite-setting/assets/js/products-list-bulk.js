/**
 * Tanzanite Settings - Products List Bulk Operations
 * 批量操作模块，包括批量删除和价格调整
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    window.TzProductsList = window.TzProductsList || {};
    
    window.TzProductsList.initBulk = function() {
        const config = window.TzProductsList.config;
        const i18n = window.TzProductsList.i18n;
        const showNotice = window.TzProductsList.showNotice;
        const loadProducts = window.TzProductsList.loadProducts;

        // 批量操作按钮
        const bulkDeleteBtn = document.getElementById('tz-bulk-delete');
        const bulkPriceBtn = document.getElementById('tz-bulk-price');

        if (!config.canBulk) {
            if (bulkDeleteBtn) bulkDeleteBtn.style.display = 'none';
            if (bulkPriceBtn) bulkPriceBtn.style.display = 'none';
            return;
        }

        /**
         * 获取选中的商品 ID
         */
        function getSelectedIds() {
            const checkboxes = document.querySelectorAll('.tz-product-checkbox:checked');
            return Array.from(checkboxes).map(function(cb) {
                return parseInt(cb.value, 10);
            }).filter(function(id) {
                return id > 0;
            });
        }

        /**
         * 批量删除
         */
        if (bulkDeleteBtn) {
            bulkDeleteBtn.addEventListener('click', async function() {
                const ids = getSelectedIds();
                
                if (!ids.length) {
                    showNotice('warning', i18n.bulkNoSelection || '请先选择商品');
                    return;
                }

                const mode = confirm(i18n.bulkDeleteConfirm || '确认删除选中的商品吗？') ? 'trash' : null;
                if (!mode) return;

                try {
                    const resp = await fetch(config.bulkUrl, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-WP-Nonce': config.nonce
                        },
                        body: JSON.stringify({
                            action: 'delete',
                            ids: ids,
                            mode: mode
                        })
                    });

                    const data = await resp.json();

                    if (!resp.ok) {
                        showNotice('error', data.message || i18n.bulkDeleteFailed, data.hint, data.actions);
                        return;
                    }

                    showNotice('success', data.message || i18n.bulkDeleteSuccess);
                    loadProducts();

                } catch (err) {
                    showNotice('error', err.message);
                }
            });
        }

        /**
         * 批量价格调整
         */
        if (bulkPriceBtn) {
            bulkPriceBtn.addEventListener('click', async function() {
                const ids = getSelectedIds();
                
                if (!ids.length) {
                    showNotice('warning', i18n.bulkNoSelection || '请先选择商品');
                    return;
                }

                // 这里可以显示一个模态框让用户输入价格调整参数
                // 简化版本：直接提示用户
                showNotice('info', i18n.bulkPriceNotImplemented || '批量价格调整功能开发中');
            });
        }

        /**
         * 全选/取消全选
         */
        const selectAllCheckbox = document.getElementById('tz-select-all');
        if (selectAllCheckbox) {
            selectAllCheckbox.addEventListener('change', function() {
                const checkboxes = document.querySelectorAll('.tz-product-checkbox');
                checkboxes.forEach(function(cb) {
                    cb.checked = selectAllCheckbox.checked;
                });
            });
        }
    };
})();
