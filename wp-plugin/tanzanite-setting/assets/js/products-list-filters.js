/**
 * Tanzanite Settings - Products List Filters
 * 筛选功能模块，包括分类和标签的动态加载
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    // 初始化筛选功能
    window.TzProductsList = window.TzProductsList || {};
    
    window.TzProductsList.initFilters = function() {
        const config = window.TzProductsList.config;
        const i18n = window.TzProductsList.i18n;

        // 分类和标签状态管理
        const taxonomyState = {
            category: { page: 1, search: '', hasMore: true, loading: false },
            tags: { page: 1, search: '', hasMore: true, loading: false }
        };

        const taxonomySelects = {
            category: document.getElementById('tz-filter-category'),
            tags: document.getElementById('tz-filter-tags')
        };

        const taxonomyStatusLabels = {
            category: document.querySelector('.tz-taxonomy-status[data-taxonomy="category"]'),
            tags: document.querySelector('.tz-taxonomy-status[data-taxonomy="tags"]')
        };

        const taxonomySearchInputs = {
            category: document.querySelector('.tz-taxonomy-search[data-taxonomy="category"]'),
            tags: document.querySelector('.tz-taxonomy-search[data-taxonomy="tags"]')
        };

        const taxonomySearchButtons = {
            category: document.querySelector('.tz-taxonomy-search-btn[data-taxonomy="category"]'),
            tags: document.querySelector('.tz-taxonomy-search-btn[data-taxonomy="tags"]')
        };

        const taxonomyLoadButtons = {
            category: document.querySelector('.tz-taxonomy-load-more[data-taxonomy="category"]'),
            tags: document.querySelector('.tz-taxonomy-load-more[data-taxonomy="tags"]')
        };

        // 保存占位符文本
        Object.keys(taxonomySelects).forEach(function(taxonomy) {
            const select = taxonomySelects[taxonomy];
            if (!select) return;
            
            if (!select.dataset.placeholder) {
                const placeholderOption = select.querySelector('option[value=""]');
                if (placeholderOption) {
                    select.dataset.placeholder = placeholderOption.textContent || '';
                }
            }
        });

        /**
         * 设置分类/标签状态提示
         */
        function setTaxonomyStatus(taxonomy, message, type) {
            const label = taxonomyStatusLabels[taxonomy];
            if (!label) return;
            
            label.textContent = message || '';
            label.dataset.type = type || '';
        }

        /**
         * 加载分类/标签数据
         */
        async function fetchTaxonomy(taxonomy, options) {
            options = options || {};
            const reset = options.reset || false;

            // 构建端点名称（category -> categoryEndpoint, tags -> tagsEndpoint）
            const endpointKey = taxonomy + 'Endpoint';
            if (!config[endpointKey]) {
                return;
            }

            const store = taxonomyState[taxonomy];
            if (!store) return;

            if (store.loading) return;

            if (reset) {
                store.page = 1;
                store.hasMore = true;
            } else if (!store.hasMore) {
                setTaxonomyStatus(taxonomy, i18n.taxonomyNoMore, 'info');
                return;
            }

            store.loading = true;
            setTaxonomyStatus(taxonomy, i18n.taxonomyLoading, 'loading');

            const params = new URLSearchParams({
                per_page: '20',
                page: String(store.page)
            });

            if (store.search) {
                params.set('search', store.search);
            }

            try {
                const resp = await fetch(config[endpointKey] + '?' + params.toString(), {
                    headers: {
                        'X-WP-Nonce': config.nonce
                    }
                });

                const data = await resp.json();

                if (!resp.ok) {
                    throw new Error(data.message || i18n.taxonomyLoadFailed);
                }

                const items = Array.isArray(data.items) ? data.items : [];
                const meta = data.meta || {};

                populateTaxonomyOptions(taxonomy, items, reset);

                const currentPage = parseInt(meta.page, 10);
                const basePage = Number.isInteger(currentPage) && currentPage > 0 ? currentPage : store.page;

                store.hasMore = Boolean(meta.has_more);

                if (store.hasMore) {
                    const nextPage = parseInt(meta.next_page, 10);
                    if (Number.isInteger(nextPage) && nextPage > basePage) {
                        store.page = nextPage;
                    } else {
                        store.page = basePage + 1;
                    }
                } else {
                    store.page = basePage;
                }

                if (!items.length) {
                    setTaxonomyStatus(taxonomy, i18n.taxonomyEmpty, 'empty');
                } else if (store.hasMore) {
                    setTaxonomyStatus(taxonomy, i18n.taxonomyHasMore, 'info');
                } else {
                    setTaxonomyStatus(taxonomy, i18n.taxonomyNoMore, 'info');
                }

            } catch (err) {
                setTaxonomyStatus(taxonomy, err.message || i18n.taxonomyLoadFailed, 'error');
            } finally {
                store.loading = false;
            }
        }

        /**
         * 填充分类/标签选项
         */
        function populateTaxonomyOptions(taxonomy, items, reset) {
            const select = taxonomySelects[taxonomy];
            if (!select) return;

            const selectedValues = Array.from(select.selectedOptions).map(function(opt) {
                return opt.value;
            });

            if (reset) {
                if (taxonomy === 'category') {
                    const first = select.querySelector('option[value=""]');
                    select.innerHTML = '';
                    if (first) {
                        select.appendChild(first);
                    } else {
                        const placeholder = document.createElement('option');
                        placeholder.value = '';
                        placeholder.textContent = select.dataset.placeholder || '';
                        select.appendChild(placeholder);
                    }
                } else {
                    select.innerHTML = '';
                }
            }

            items.forEach(function(item) {
                const value = String(
                    taxonomy === 'category'
                        ? (item.term_id || '')
                        : (item.slug || item.term_id || '')
                ).trim();
                
                if (!value) return;

                const existingOption = Array.from(select.options).find(function(opt) {
                    return opt.value === value;
                });

                if (existingOption) {
                    existingOption.textContent = item.name || value;
                    if (selectedValues.includes(existingOption.value)) {
                        existingOption.selected = true;
                    }
                    return;
                }

                const option = document.createElement('option');
                option.value = value;
                option.textContent = item.name || value;

                if (selectedValues.includes(option.value)) {
                    option.selected = true;
                }

                select.appendChild(option);
            });
        }

        /**
         * 处理分类/标签搜索
         */
        function handleTaxonomySearch(taxonomy) {
            const input = taxonomySearchInputs[taxonomy];
            const store = taxonomyState[taxonomy];
            if (!store) return;

            if (input) {
                store.search = input.value.trim();
            } else {
                store.search = '';
            }

            fetchTaxonomy(taxonomy, { reset: true });
        }

        /**
         * 绑定分类/标签控件事件
         */
        function bindTaxonomyControls() {
            Object.keys(taxonomyState).forEach(function(taxonomy) {
                const searchBtn = taxonomySearchButtons[taxonomy];
                if (searchBtn) {
                    searchBtn.addEventListener('click', function(event) {
                        event.preventDefault();
                        handleTaxonomySearch(taxonomy);
                    });
                }

                const searchInput = taxonomySearchInputs[taxonomy];
                if (searchInput) {
                    searchInput.addEventListener('keydown', function(event) {
                        if (event.key === 'Enter') {
                            event.preventDefault();
                            handleTaxonomySearch(taxonomy);
                        }
                    });
                }

                const loadBtn = taxonomyLoadButtons[taxonomy];
                if (loadBtn) {
                    loadBtn.addEventListener('click', function(event) {
                        event.preventDefault();
                        fetchTaxonomy(taxonomy);
                    });
                }
            });
        }

        // 初始化
        bindTaxonomyControls();

        // 初始加载分类和标签
        fetchTaxonomy('category');
        fetchTaxonomy('tags');
    };
})();
