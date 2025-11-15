/**
 * Tanzanite Settings - Products List Core
 * 主入口文件，协调所有模块
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    // 等待 DOM 加载完成
    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzProductsListConfig === 'undefined') {
            console.error('Products List config not found');
            return;
        }

        const config = TzProductsListConfig;
        const i18n = config.strings || {};

        // DOM 元素
        const elements = {
            notice: document.getElementById('tz-products-notice'),
            filtersForm: document.getElementById('tz-products-filters'),
            filtersToggle: document.getElementById('tz-products-filters-toggle'),
            tableBody: document.querySelector('#tz-products-table tbody'),
            submitBtn: document.getElementById('tz-products-filter-submit'),
            resetBtn: document.getElementById('tz-products-filter-reset'),
            prevBtn: document.getElementById('tz-products-prev'),
            nextBtn: document.getElementById('tz-products-next'),
            pageInfo: document.getElementById('tz-products-page-info'),
            viewToggle: document.querySelectorAll('.tz-view-toggle'),
            tableWrapper: document.querySelector('.tz-products-table-wrapper'),
            cardsWrapper: document.querySelector('.tz-products-cards'),
            summaryCards: document.querySelectorAll('.tz-dashboard-card')
        };

        // 检查必需元素
        if (!elements.filtersForm || !elements.tableBody) {
            console.error('Required elements not found');
            return;
        }

        // 状态管理
        const state = {
            page: 1,
            perPage: parseInt(elements.filtersForm.elements.per_page.value, 10) || 20,
            totalPages: 1,
            loading: false,
            currentView: 'table' // 'table' or 'cards'
        };

        // 通知显示函数
        function showNotice(type, message, hint, actions) {
            if (!elements.notice) return;
            
            if (!message) {
                elements.notice.style.display = 'none';
                elements.notice.className = 'notice';
                elements.notice.innerHTML = '';
                return;
            }

            const typeClass = type ? ' notice-' + type : ' notice-info';
            elements.notice.className = 'notice' + typeClass;
            
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
            
            elements.notice.innerHTML = html;
            elements.notice.style.display = 'block';
        }

        // 格式化金额
        function formatAmount(value) {
            if (typeof value === 'number') {
                return value.toFixed(2);
            }
            const numberValue = parseFloat(value);
            if (!Number.isFinite(numberValue)) {
                return '0.00';
            }
            return numberValue.toFixed(2);
        }

        // 构建查询参数
        function buildQuery() {
            const params = new URLSearchParams();
            const formData = new FormData(elements.filtersForm);

            formData.forEach(function(value, key) {
                if (!value) return;
                
                if (key === 'tags[]') {
                    params.append('tags', value);
                    return;
                }
                
                params.append(key, value);
            });

            params.set('page', state.page);
            params.set('per_page', state.perPage);

            return params.toString();
        }

        // 加载商品列表
        async function loadProducts() {
            if (state.loading) return;
            
            state.loading = true;
            showNotice(null);

            try {
                const query = buildQuery();
                const resp = await fetch(config.listUrl + '?' + query, {
                    headers: {
                        'X-WP-Nonce': config.nonce
                    }
                });

                const data = await resp.json();

                if (!resp.ok) {
                    showNotice('error', data.message || i18n.loadFailed, data.hint, data.actions);
                    return;
                }

                // 更新摘要卡片
                if (data.summary && elements.summaryCards) {
                    updateSummaryCards(data.summary);
                }

                // 渲染商品列表
                if (window.TzProductsList && window.TzProductsList.render) {
                    window.TzProductsList.render(data.items || [], state.currentView);
                }

                // 更新分页
                state.totalPages = data.meta?.total_pages || 1;
                updatePagination();

            } catch (err) {
                showNotice('error', err.message);
            } finally {
                state.loading = false;
            }
        }

        // 更新摘要卡片
        function updateSummaryCards(summary) {
            elements.summaryCards.forEach(function(card) {
                const metric = card.dataset.metric;
                if (metric && summary[metric] !== undefined) {
                    const valueEl = card.querySelector('.tz-card-value');
                    if (valueEl) {
                        valueEl.textContent = summary[metric];
                    }
                }
            });
        }

        // 更新分页
        function updatePagination() {
            if (elements.pageInfo) {
                const template = i18n.pageTemplate || '第 {page}/{pages} 页';
                elements.pageInfo.textContent = template
                    .replace('{page}', state.page)
                    .replace('{pages}', state.totalPages);
            }

            if (elements.prevBtn) {
                elements.prevBtn.disabled = state.page <= 1;
            }
            if (elements.nextBtn) {
                elements.nextBtn.disabled = state.page >= state.totalPages;
            }
        }

        // 筛选器折叠切换
        if (elements.filtersToggle) {
            elements.filtersToggle.addEventListener('click', function() {
                const isCollapsed = elements.filtersForm.classList.toggle('tz-filters-collapsed');
                const label = this.querySelector('.tz-toggle-label');
                const icon = this.querySelector('.dashicons');
                
                if (label) {
                    label.textContent = isCollapsed ? i18n.expandFilters : i18n.collapseFilters;
                }
                if (icon) {
                    icon.classList.toggle('dashicons-arrow-down', !isCollapsed);
                    icon.classList.toggle('dashicons-arrow-up', isCollapsed);
                }
                
                this.setAttribute('aria-expanded', !isCollapsed);
            });
        }

        // 视图切换
        elements.viewToggle.forEach(function(btn) {
            btn.addEventListener('click', function() {
                const view = this.dataset.view;
                if (view === state.currentView) return;
                
                state.currentView = view;
                
                // 更新按钮状态
                elements.viewToggle.forEach(function(b) {
                    b.classList.toggle('active', b.dataset.view === view);
                });
                
                // 切换显示
                if (elements.tableWrapper) {
                    elements.tableWrapper.style.display = view === 'table' ? '' : 'none';
                }
                if (elements.cardsWrapper) {
                    elements.cardsWrapper.style.display = view === 'cards' ? '' : 'none';
                }
                
                // 重新渲染
                loadProducts();
            });
        });

        // 筛选提交
        if (elements.submitBtn) {
            elements.submitBtn.addEventListener('click', function(e) {
                e.preventDefault();
                state.page = 1;
                loadProducts();
            });
        }

        // 重置筛选
        if (elements.resetBtn) {
            elements.resetBtn.addEventListener('click', function(e) {
                e.preventDefault();
                elements.filtersForm.reset();
                state.page = 1;
                loadProducts();
            });
        }

        // 分页按钮
        if (elements.prevBtn) {
            elements.prevBtn.addEventListener('click', function() {
                if (state.page > 1) {
                    state.page--;
                    loadProducts();
                }
            });
        }

        if (elements.nextBtn) {
            elements.nextBtn.addEventListener('click', function() {
                if (state.page < state.totalPages) {
                    state.page++;
                    loadProducts();
                }
            });
        }

        // 暴露公共 API
        window.TzProductsList = window.TzProductsList || {};
        window.TzProductsList.config = config;
        window.TzProductsList.i18n = i18n;
        window.TzProductsList.elements = elements;
        window.TzProductsList.state = state;
        window.TzProductsList.showNotice = showNotice;
        window.TzProductsList.formatAmount = formatAmount;
        window.TzProductsList.loadProducts = loadProducts;

        // 初始化其他模块
        if (window.TzProductsList.initFilters) {
            window.TzProductsList.initFilters();
        }
        if (window.TzProductsList.initBulk) {
            window.TzProductsList.initBulk();
        }
        if (window.TzProductsList.initActions) {
            window.TzProductsList.initActions();
        }

        // 初始加载
        loadProducts();
    });
})();
