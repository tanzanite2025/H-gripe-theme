/**
 * Tanzanite Settings - Order Bulk Operations
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzOrderBulkConfig === 'undefined') {
            console.error('Order Bulk config not found');
            return;
        }

        const config = TzOrderBulkConfig;
        const i18n = config.strings || {};

        // DOM 元素
        const notice = document.getElementById('tz-order-bulk-notice');
        const resultBox = document.getElementById('tz-order-bulk-result');
        const forms = {
            status: document.getElementById('tz-order-bulk-status'),
            export: document.getElementById('tz-order-bulk-export')
        };

        if (!notice || !resultBox || !forms.status || !forms.export) {
            return;
        }

        /**
         * 解析订单 ID 列表
         */
        function parseIds(raw) {
            if (!raw) {
                return [];
            }
            return raw.split(/[\s,]+/)
                .map(function(id) {
                    return parseInt(id, 10);
                })
                .filter(function(id) {
                    return id > 0;
                });
        }

        /**
         * 显示通知
         */
        function showNotice(type, message, hint, actions) {
            if (!message) {
                notice.style.display = 'none';
                notice.className = 'notice';
                notice.innerHTML = '';
                return;
            }

            const typeClass = type ? ' notice-' + type : ' notice-info';
            notice.className = 'notice' + typeClass;
            
            let html = '<p>' + message + '</p>';
            
            if (hint) {
                html += '<p>' + hint + '</p>';
            }
            
            if (Array.isArray(actions) && actions.length) {
                html += '<ul style="margin:8px 0 0 18px;">';
                actions.forEach(function(act) {
                    html += '<li>' + act + '</li>';
                });
                html += '</ul>';
            }
            
            notice.innerHTML = html;
            notice.style.display = 'block';
        }

        /**
         * 渲染结果
         */
        function renderResult(data) {
            if (!data) {
                resultBox.textContent = '';
                return;
            }
            resultBox.textContent = JSON.stringify(data, null, 2);
        }

        /**
         * 发送批量请求
         */
        async function requestBulk(body) {
            showNotice(null);
            renderResult({ loading: true, action: body.action });

            try {
                const resp = await fetch(config.url, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-WP-Nonce': config.nonce
                    },
                    body: JSON.stringify(body)
                });

                const json = await resp.json();
                const type = resp.ok ? 'success' : (resp.status === 207 ? 'warning' : 'error');
                
                showNotice(type, json.message || i18n.done, json.hint, json.actions);
                renderResult(json);

                // 处理文件下载
                if (json.file && json.file.content) {
                    const link = document.createElement('a');
                    link.href = 'data:' + (json.file.mime || 'text/csv') + ';base64,' + json.file.content;
                    link.download = json.file.name || 'orders-export.csv';
                    document.body.appendChild(link);
                    link.click();
                    link.remove();
                }

            } catch (err) {
                showNotice('error', err.message);
                renderResult({ error: err.message });
            }
        }

        /**
         * 批量更新状态表单提交
         */
        forms.status.addEventListener('submit', function(event) {
            event.preventDefault();
            
            const ids = parseIds(this.elements.ids.value);
            if (!ids.length) {
                showNotice('warning', i18n.invalidIds);
                return;
            }

            const status = this.elements.status.value;
            if (!status) {
                showNotice('warning', i18n.invalidStatus);
                return;
            }

            requestBulk({
                action: 'set_status',
                ids: ids,
                payload: { status: status }
            });
        });

        /**
         * 批量导出表单提交
         */
        forms.export.addEventListener('submit', function(event) {
            event.preventDefault();
            
            const ids = parseIds(this.elements.ids.value);
            if (!ids.length) {
                showNotice('warning', i18n.invalidIds);
                return;
            }

            requestBulk({
                action: 'export',
                ids: ids,
                payload: {}
            });
        });
    });
})();
