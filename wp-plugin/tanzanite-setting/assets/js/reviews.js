/**
 * Tanzanite Settings - 商品评价管理
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzReviewsConfig === 'undefined') {
            console.error('Reviews config not found');
            return;
        }

        const config = TzReviewsConfig;
        const i18n = config.i18n || {};
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            notice: document.getElementById('tz-review-notice'),
            filtersForm: document.getElementById('tz-review-filters'),
            refreshBtn: document.getElementById('tz-review-refresh'),
            resetBtn: document.getElementById('tz-review-reset'),
            tableBody: document.querySelector('#tz-review-table tbody'),
            prevBtn: document.getElementById('tz-review-prev'),
            nextBtn: document.getElementById('tz-review-next'),
            pageInfo: document.getElementById('tz-review-page-info'),
            detailForm: document.getElementById('tz-review-detail'),
            fieldId: document.getElementById('tz-review-id'),
            fieldStatus: document.getElementById('tz-review-status'),
            fieldFeatured: document.getElementById('tz-review-featured'),
            fieldRating: document.getElementById('tz-review-rating'),
            fieldAuthor: document.getElementById('tz-review-author'),
            fieldCreated: document.getElementById('tz-review-created'),
            fieldContent: document.getElementById('tz-review-content'),
            fieldImages: document.getElementById('tz-review-images'),
            fieldReply: document.getElementById('tz-review-reply'),
            saveBtn: document.getElementById('tz-review-save'),
            cancelBtn: document.getElementById('tz-review-cancel'),
            deleteBtn: document.getElementById('tz-review-delete')
        };

        // 状态
        const state = {
            page: 1,
            totalPages: 1,
            perPage: parseInt(elements.filtersForm.elements.per_page.value, 10) || 20,
            selected: null,
            items: []
        };

        /**
         * 构建查询参数
         */
        function buildParams(extra) {
            const params = new URLSearchParams();
            const formData = new FormData(elements.filtersForm);

            formData.forEach(function(value, key) {
                if (value) {
                    params.append(key, value);
                }
            });

            params.set('page', state.page);
            params.set('per_page', state.perPage);

            if (extra) {
                Object.keys(extra).forEach(function(key) {
                    if (extra[key] !== null && extra[key] !== undefined) {
                        params.set(key, extra[key]);
                    }
                });
            }

            return params;
        }

        /**
         * 渲染图片列表
         */
        function renderImages(images) {
            elements.fieldImages.innerHTML = '';
            
            if (!Array.isArray(images) || !images.length) {
                elements.fieldImages.textContent = i18n.contentPlaceholder || '暂无内容';
                return;
            }

            images.forEach(function(img) {
                const link = document.createElement('a');
                link.href = img.url || img;
                link.target = '_blank';
                link.rel = 'noopener noreferrer';
                link.textContent = img.filename || img.url || img;
                link.style.display = 'block';
                elements.fieldImages.appendChild(link);
            });
        }

        /**
         * 渲染评价列表
         */
        function renderRows(items) {
            elements.tableBody.innerHTML = '';
            state.items = items;

            if (!items.length) {
                const tr = document.createElement('tr');
                const td = document.createElement('td');
                td.colSpan = 9;
                td.textContent = '-';
                td.style.textAlign = 'center';
                tr.appendChild(td);
                elements.tableBody.appendChild(tr);
                return;
            }

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                tr.dataset.id = item.id;
                
                const content = (item.content || '').slice(0, 90);
                const contentSuffix = item.content && item.content.length > 90 ? '…' : '';
                
                tr.innerHTML = `
                    <td>${item.id}</td>
                    <td>${item.product_id}</td>
                    <td>${item.user_id ? item.user_id : (item.author_name || '-')}</td>
                    <td>${item.rating}</td>
                    <td>${content}${contentSuffix}</td>
                    <td>${item.status}</td>
                    <td>${item.is_featured ? (i18n.yes || '是') : (i18n.no || '否')}</td>
                    <td>${item.created_at || '-'}</td>
                    <td class="tz-review-actions"></td>
                `;

                const actionsTd = tr.querySelector('.tz-review-actions');

                if (config.canManage) {
                    const featuredText = item.is_featured ? (i18n.unmarkFeatured || '取消精华') : (i18n.markFeatured || '标记精华');
                    actionsTd.innerHTML = `
                        <button class="button-link tz-review-view" data-id="${item.id}">${i18n.view || '查看'}</button>
                        | <button class="button-link tz-review-approve" data-id="${item.id}">${i18n.approve || '通过'}</button>
                        | <button class="button-link tz-review-reject" data-id="${item.id}">${i18n.reject || '拒绝'}</button>
                        | <button class="button-link tz-review-hide" data-id="${item.id}">${i18n.hide || '隐藏'}</button>
                        | <button class="button-link tz-review-feature" data-id="${item.id}" data-featured="${item.is_featured ? 1 : 0}">${featuredText}</button>
                    `;
                } else {
                    actionsTd.textContent = '-';
                }

                elements.tableBody.appendChild(tr);
            });
        }

        /**
         * 渲染分页
         */
        function renderPagination(meta) {
            state.totalPages = meta.total_pages || 1;
            state.perPage = meta.per_page || state.perPage;

            elements.prevBtn.disabled = state.page <= 1;
            elements.nextBtn.disabled = state.page >= state.totalPages;
            elements.pageInfo.textContent = `第 ${state.page} / ${state.totalPages} 页，共 ${meta.total || 0} ${i18n.itemsLabel || '条评价'}`;
        }

        /**
         * 加载评价列表
         */
        async function loadReviews() {
            try {
                const params = buildParams();
                const url = config.listUrl + '?' + params.toString();
                const result = await apiRequest(url);

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || i18n.loadFailed || '加载失败');
                    renderRows([]);
                    return;
                }

                renderRows(result.data.items || []);
                renderPagination(result.data.pagination || {});
                showNotice(elements.notice, null);

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
                renderRows([]);
            }
        }

        /**
         * 加载单个评价
         */
        async function loadReview(id) {
            try {
                const result = await apiRequest(config.singleUrl + id);

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                const review = result.data;
                state.selected = review;

                elements.fieldId.value = review.id;
                elements.fieldStatus.value = review.status || '';
                elements.fieldFeatured.checked = !!review.is_featured;
                elements.fieldRating.textContent = review.rating || '-';
                elements.fieldAuthor.textContent = review.author_name || review.user_id || '-';
                elements.fieldCreated.textContent = review.created_at || '-';
                elements.fieldContent.value = review.content || '';
                elements.fieldReply.value = review.reply_text || '';

                renderImages(review.images || []);

                // 滚动到详情表单
                window.scrollTo({
                    top: elements.detailForm.offsetTop - 80,
                    behavior: 'smooth'
                });

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 保存评价
         */
        async function saveReview(e) {
            e.preventDefault();

            const id = elements.fieldId.value;
            if (!id) {
                showNotice(elements.notice, 'warning', i18n.selectReview || '请先选择要操作的评价');
                return;
            }

            const payload = {};

            if (elements.fieldStatus.value) {
                payload.status = elements.fieldStatus.value;
            }

            payload.is_featured = elements.fieldFeatured.checked;

            if (elements.fieldReply.value.trim()) {
                payload.reply_text = elements.fieldReply.value.trim();
            } else {
                payload.reply_text = null;
            }

            try {
                const result = await apiRequest(config.singleUrl + id, {
                    method: 'PUT',
                    body: JSON.stringify(payload)
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '保存失败');
                    return;
                }

                showNotice(elements.notice, 'success', i18n.saveSuccess || '评价已更新');
                loadReviews();
                resetDetail();

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 删除评价
         */
        async function deleteReview() {
            const id = elements.fieldId.value;
            if (!id) {
                showNotice(elements.notice, 'warning', i18n.selectReview || '请先选择要操作的评价');
                return;
            }

            if (!confirm(i18n.deleteConfirm || '确定删除该评价？此操作不可撤销。')) {
                return;
            }

            try {
                const result = await apiRequest(config.singleUrl + id, {
                    method: 'DELETE'
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '删除失败');
                    return;
                }

                showNotice(elements.notice, 'success', i18n.deleteSuccess || '评价已删除');
                loadReviews();
                resetDetail();

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 快速操作
         */
        async function quickAction(id, action, value) {
            const payload = {};

            switch (action) {
                case 'approve':
                    payload.status = 'approved';
                    break;
                case 'reject':
                    payload.status = 'rejected';
                    break;
                case 'hide':
                    payload.status = 'hidden';
                    break;
                case 'feature':
                    payload.is_featured = value;
                    break;
                default:
                    return;
            }

            try {
                const result = await apiRequest(config.singleUrl + id, {
                    method: 'PUT',
                    body: JSON.stringify(payload)
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '操作失败');
                    return;
                }

                showNotice(elements.notice, 'success', '操作成功');
                loadReviews();

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 重置详情表单
         */
        function resetDetail() {
            elements.detailForm.reset();
            elements.fieldId.value = '';
            elements.fieldRating.textContent = '-';
            elements.fieldAuthor.textContent = '-';
            elements.fieldCreated.textContent = '-';
            elements.fieldImages.innerHTML = '';
            state.selected = null;
        }

        /**
         * 重置筛选
         */
        function resetFilters() {
            elements.filtersForm.reset();
            state.page = 1;
            loadReviews();
        }

        // 事件监听
        if (elements.refreshBtn) {
            elements.refreshBtn.addEventListener('click', function() {
                state.page = 1;
                loadReviews();
            });
        }

        if (elements.resetBtn) {
            elements.resetBtn.addEventListener('click', resetFilters);
        }

        if (elements.prevBtn) {
            elements.prevBtn.addEventListener('click', function() {
                if (state.page > 1) {
                    state.page--;
                    loadReviews();
                }
            });
        }

        if (elements.nextBtn) {
            elements.nextBtn.addEventListener('click', function() {
                if (state.page < state.totalPages) {
                    state.page++;
                    loadReviews();
                }
            });
        }

        if (elements.detailForm) {
            elements.detailForm.addEventListener('submit', saveReview);
        }

        if (elements.cancelBtn) {
            elements.cancelBtn.addEventListener('click', resetDetail);
        }

        if (elements.deleteBtn) {
            elements.deleteBtn.addEventListener('click', deleteReview);
        }

        // 表格行点击事件
        if (elements.tableBody) {
            elements.tableBody.addEventListener('click', function(e) {
                const target = e.target;
                const id = parseInt(target.dataset.id);

                if (!id) return;

                if (target.classList.contains('tz-review-view')) {
                    e.preventDefault();
                    loadReview(id);
                } else if (target.classList.contains('tz-review-approve')) {
                    e.preventDefault();
                    quickAction(id, 'approve');
                } else if (target.classList.contains('tz-review-reject')) {
                    e.preventDefault();
                    quickAction(id, 'reject');
                } else if (target.classList.contains('tz-review-hide')) {
                    e.preventDefault();
                    quickAction(id, 'hide');
                } else if (target.classList.contains('tz-review-feature')) {
                    e.preventDefault();
                    const featured = parseInt(target.dataset.featured);
                    quickAction(id, 'feature', !featured);
                }
            });
        }

        // 每页条数变化
        if (elements.filtersForm) {
            elements.filtersForm.elements.per_page.addEventListener('change', function() {
                state.perPage = parseInt(this.value, 10);
                state.page = 1;
                loadReviews();
            });
        }

        // 权限提示
        if (!config.canManage) {
            showNotice(elements.notice, 'warning', i18n.noPermission || '当前账号仅具备查看权限', i18n.noPermissionHint || '如需执行审核或回复，请联系管理员授予权限');
        }

        // 初始化
        loadReviews();
    });
})();
