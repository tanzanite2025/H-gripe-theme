/**
 * Tanzanite Settings - SKU Importer
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzSkuImporterConfig === 'undefined') {
            console.error('SKU Importer config not found');
            return;
        }

        const config = TzSkuImporterConfig;

        // DOM 元素
        const productInput = document.getElementById('tz-sku-import-product');
        const textArea = document.getElementById('tz-sku-import-text');
        const previewBtn = document.getElementById('tz-sku-import-preview');
        const applyBtn = document.getElementById('tz-sku-import-apply');
        const downloadBtn = document.getElementById('tz-sku-import-download-template');
        const previewPanel = document.getElementById('tz-sku-import-preview-panel');
        const previewTable = document.querySelector('#tz-sku-import-table tbody');
        const summary = document.getElementById('tz-sku-import-summary');

        if (!productInput || !textArea || !previewBtn || !applyBtn || !downloadBtn) {
            return;
        }

        let lastPreview = null;

        // CSV 模板内容
        const templateCSV = 'sku_code,price_regular,price_sale,stock_qty,attributes\nSKU-001,199,179,50,{"color":"blue"}';

        /**
         * 下载 CSV 模板
         */
        downloadBtn.addEventListener('click', function(e) {
            e.preventDefault();
            
            const blob = new Blob([templateCSV], { type: 'text/csv' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'sku-template.csv';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        });

        /**
         * 渲染预检结果
         */
        function renderPreview(data) {
            previewTable.innerHTML = '';
            
            if (!data.skus || !data.skus.length) {
                return;
            }

            data.skus.forEach(function(item) {
                const tr = document.createElement('tr');
                
                ['sku_code', 'price_regular', 'price_sale', 'stock_qty', 'tier_prices', 'sort_order'].forEach(function(key) {
                    const td = document.createElement('td');
                    let value = item[key];
                    
                    if (key === 'tier_prices') {
                        value = JSON.stringify(value || []);
                    }
                    
                    td.textContent = value ?? '';
                    tr.appendChild(td);
                });
                
                previewTable.appendChild(tr);
            });

            // 渲染摘要信息
            let summaryHtml = '';
            summaryHtml += '<p><strong>' + data.summary.product_title + '</strong></p>';
            summaryHtml += '<p>' + data.summary.message + '</p>';
            
            if (data.summary.duplicates && data.summary.duplicates.length) {
                summaryHtml += '<p style="color:#d97706">' + data.summary.duplicates_message + ': ' + data.summary.duplicates.join(', ') + '</p>';
            }
            
            summary.innerHTML = summaryHtml;
            previewPanel.style.display = '';
            applyBtn.disabled = false;
        }

        /**
         * 显示错误信息
         */
        function showError(message, hint) {
            previewPanel.style.display = '';
            previewTable.innerHTML = '';
            
            let errorHtml = '<p style="color:#b91c1c">' + message + '</p>';
            if (hint) {
                errorHtml += '<p>' + hint + '</p>';
            }
            
            summary.innerHTML = errorHtml;
            applyBtn.disabled = true;
        }

        /**
         * 预检数据
         */
        previewBtn.addEventListener('click', async function() {
            applyBtn.disabled = true;
            previewPanel.style.display = 'none';
            summary.innerHTML = '';
            previewTable.innerHTML = '';

            const productId = parseInt(productInput.value, 10);
            if (!productId) {
                showError(config.strings.errorNoProductId || '请先输入商品 ID。', '');
                return;
            }

            const payload = {
                product_id: productId,
                skus_bulk: textArea.value
            };

            try {
                const resp = await fetch(config.previewUrl, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-WP-Nonce': config.nonce
                    },
                    body: JSON.stringify(payload)
                });

                const data = await resp.json();

                if (!resp.ok) {
                    showError(
                        data.message || config.strings.errorPreviewFailed || '预检失败。',
                        data.hint
                    );
                    return;
                }

                lastPreview = data;
                renderPreview(data);

            } catch (err) {
                showError(
                    err.message,
                    config.strings.errorNetworkRetry || '请检查网络后重试。'
                );
            }
        });

        /**
         * 确认导入
         */
        applyBtn.addEventListener('click', async function() {
            if (!lastPreview) {
                return;
            }

            applyBtn.disabled = true;
            const productId = parseInt(productInput.value, 10);

            try {
                const resp = await fetch(config.applyUrl, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-WP-Nonce': config.nonce
                    },
                    body: JSON.stringify({
                        product_id: productId,
                        skus: lastPreview.skus
                    })
                });

                const data = await resp.json();

                if (!resp.ok) {
                    showError(
                        data.message || config.strings.errorImportFailed || '导入失败。',
                        data.hint
                    );
                    return;
                }

                summary.innerHTML = '<p style="color:#047857">' + (config.strings.successImported || '导入成功，SKU 已更新。') + '</p>';
                previewTable.innerHTML = '';
                lastPreview = null;
                textArea.value = '';

            } catch (err) {
                showError(
                    err.message,
                    config.strings.errorRetryLater || '请稍后重试。'
                );
            }
        });
    });
})();
