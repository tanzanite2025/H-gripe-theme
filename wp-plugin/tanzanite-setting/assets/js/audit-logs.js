/**
 * Tanzanite Settings - Audit Logs
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzAuditLogsConfig === 'undefined') {
            console.error('Audit Logs config not found');
            return;
        }

        const config = TzAuditLogsConfig;

        // DOM 元素
        const tblBody = document.querySelector('#tz-audit-table tbody');
        const summary = document.getElementById('tz-audit-summary');
        const btnPrev = document.getElementById('tz-audit-prev');
        const btnNext = document.getElementById('tz-audit-next');
        const pageInfo = document.getElementById('tz-audit-page-info');
        const applyBtn = document.getElementById('tz-audit-apply');
        const resetBtn = document.getElementById('tz-audit-reset');
        const exportBtn = document.getElementById('tz-audit-export');
        const noticeBox = document.getElementById('tz-audit-notice');

        const filters = {
            action: document.getElementById('tz-audit-action'),
            target: document.getElementById('tz-audit-target'),
            actor: document.getElementById('tz-audit-actor'),
            search: document.getElementById('tz-audit-search'),
            start: document.getElementById('tz-audit-start'),
            end: document.getElementById('tz-audit-end'),
            perPage: document.getElementById('tz-audit-per-page')
        };

        if (!tblBody || !summary || !applyBtn) {
            return;
        }

        let state = {
            page: 1,
            totalPages: 1,
            total: 0
        };

        /**
         * 显示通知
         */
        function showNotice(type, message, hint, actions) {
            if (!message) {
                noticeBox.style.display = 'none';
                noticeBox.className = 'notice';
                noticeBox.innerHTML = '';
                return;
            }

            const typeClass = type ? ' notice-' + type : ' notice-info';
            noticeBox.className = 'notice' + typeClass;
            
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
            
            noticeBox.innerHTML = html;
            noticeBox.style.display = 'block';
        }

        /**
         * 构建查询参数
         */
        function buildParams(extra) {
            const params = new URLSearchParams();
            params.set('page', state.page);
            params.set('per_page', filters.perPage.value || 20);
            
            if (filters.action.value) {
                params.set('action', filters.action.value);
            }
            if (filters.target.value) {
                params.set('target_type', filters.target.value);
            }
            if (filters.actor.value) {
                params.set('actor', filters.actor.value);
            }
            if (filters.search.value) {
                params.set('search', filters.search.value);
            }
            if (filters.start.value) {
                params.set('start', filters.start.value);
            }
            if (filters.end.value) {
                params.set('end', filters.end.value);
            }
            
            if (extra) {
                Object.entries(extra).forEach(function([key, val]) {
                    if (val != null) {
                        params.set(key, val);
                    }
                });
            }
            
            return params;
        }

        /**
         * 渲染表格行
         */
        function renderRows(items) {
            tblBody.innerHTML = '';
            
            if (!items.length) {
                const tr = document.createElement('tr');
                const td = document.createElement('td');
                td.colSpan = 7;
                td.textContent = config.strings.noRecords || '暂无记录';
                td.style.textAlign = 'center';
                tr.appendChild(td);
                tblBody.appendChild(tr);
                return;
            }

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                const payload = JSON.stringify(item.payload || {}, null, 2);
                
                const fields = [
                    item.id,
                    item.created_at,
                    item.action,
                    item.target_type + ' #' + item.target_id,
                    item.actor_name || item.actor_id,
                    item.ip_address || '-',
                    payload
                ];
                
                fields.forEach(function(value, index) {
                    const td = document.createElement('td');
                    
                    if (index === 6) {
                        const pre = document.createElement('code');
                        pre.textContent = value;
                        td.appendChild(pre);
                    } else {
                        td.textContent = value;
                    }
                    
                    tr.appendChild(td);
                });
                
                tblBody.appendChild(tr);
            });
        }

        /**
         * 渲染摘要信息
         */
        function renderSummary(meta) {
            const template = config.strings.summaryTemplate || '共 {total} 条记录，当前第 {page}/{pages} 页。';
            const text = template
                .replace('{total}', meta.total)
                .replace('{page}', state.page)
                .replace('{pages}', meta.total_pages || 1);
            
            summary.innerHTML = '<p>' + text + '</p>';
        }

        /**
         * 加载日志数据
         */
        async function fetchLogs() {
            const params = buildParams();
            showNotice(null);

            try {
                const resp = await fetch(config.restUrl + '?' + params.toString(), {
                    headers: {
                        'X-WP-Nonce': config.nonce
                    }
                });

                const data = await resp.json();

                if (!resp.ok) {
                    renderRows([]);
                    renderSummary({ total: 0, total_pages: 0 });
                    pageInfo.textContent = '';
                    btnPrev.disabled = true;
                    btnNext.disabled = true;
                    showNotice('error', data.message || config.strings.loadFailed || '加载失败。', data.hint, data.actions);
                    return;
                }

                renderRows(data.items || []);
                state.totalPages = data.meta?.total_pages || 1;
                state.total = data.meta?.total || 0;
                renderSummary(data.meta || {});
                
                const pageTemplate = config.strings.pageTemplate || '第 {page}/{pages} 页';
                pageInfo.textContent = pageTemplate
                    .replace('{page}', state.page)
                    .replace('{pages}', state.totalPages);
                
                btnPrev.disabled = state.page <= 1;
                btnNext.disabled = state.page >= state.totalPages;

            } catch (err) {
                renderRows([]);
                renderSummary({ total: 0, total_pages: 0 });
                pageInfo.textContent = '';
                btnPrev.disabled = true;
                btnNext.disabled = true;
                showNotice('error', err.message);
            }
        }

        /**
         * 应用筛选
         */
        applyBtn.addEventListener('click', function() {
            state.page = 1;
            fetchLogs();
        });

        /**
         * 重置筛选
         */
        resetBtn.addEventListener('click', function() {
            Object.values(filters).forEach(function(input) {
                if (input.tagName === 'SELECT' || input.tagName === 'INPUT') {
                    input.value = '';
                }
            });
            filters.perPage.value = '20';
            state.page = 1;
            fetchLogs();
        });

        /**
         * 上一页
         */
        btnPrev.addEventListener('click', function() {
            if (state.page > 1) {
                state.page--;
                fetchLogs();
            }
        });

        /**
         * 下一页
         */
        btnNext.addEventListener('click', function() {
            if (state.page < state.totalPages) {
                state.page++;
                fetchLogs();
            }
        });

        /**
         * 导出 CSV
         */
        exportBtn.addEventListener('click', function() {
            const params = buildParams({ format: 'csv', per_page: 5000 });
            showNotice('success', config.strings.exportStarted || '已开始导出，CSV 将在新标签页中打开。');
            window.open(config.restUrl + '?' + params.toString(), '_blank');
        });

        // 初始加载
        fetchLogs();
    });
})();
