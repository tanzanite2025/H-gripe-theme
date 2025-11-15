/**
 * Tanzanite Settings - 会员档案管理
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzMemberProfilesConfig === 'undefined') {
            console.error('Member Profiles config not found');
            return;
        }

        const config = TzMemberProfilesConfig;
        const i18n = config.i18n || {};
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            notice: document.getElementById('tz-member-notice'),
            filtersForm: document.getElementById('tz-member-filters'),
            refreshBtn: document.getElementById('tz-member-refresh'),
            resetBtn: document.getElementById('tz-member-reset'),
            exportBtn: document.getElementById('tz-member-export'),
            tableBody: document.querySelector('#tz-member-table tbody'),
            prevBtn: document.getElementById('tz-member-prev'),
            nextBtn: document.getElementById('tz-member-next'),
            pageInfo: document.getElementById('tz-member-page-info'),
            detailSection: document.getElementById('tz-member-detail-section'),
            detailForm: document.getElementById('tz-member-detail'),
            fieldId: document.getElementById('tz-member-id'),
            fieldUsername: document.getElementById('tz-member-username'),
            fieldEmail: document.getElementById('tz-member-email'),
            fieldFullname: document.getElementById('tz-member-fullname'),
            fieldPhone: document.getElementById('tz-member-phone'),
            fieldCountry: document.getElementById('tz-member-country'),
            fieldAddress: document.getElementById('tz-member-address'),
            fieldBrand: document.getElementById('tz-member-brand'),
            fieldPoints: document.getElementById('tz-member-points'),
            fieldMarketing: document.getElementById('tz-member-marketing'),
            fieldNotes: document.getElementById('tz-member-notes'),
            saveBtn: document.getElementById('tz-member-save'),
            cancelBtn: document.getElementById('tz-member-cancel')
        };

        // 状态
        let currentPage = 1;
        let totalPages = 1;
        let currentFilters = {};

        // 初始化
        loadMembers();

        // 事件监听
        elements.refreshBtn?.addEventListener('click', () => {
            currentPage = 1;
            loadMembers();
        });

        elements.resetBtn?.addEventListener('click', () => {
            elements.filtersForm?.reset();
            currentPage = 1;
            currentFilters = {};
            loadMembers();
        });

        elements.exportBtn?.addEventListener('click', exportMembers);
        elements.prevBtn?.addEventListener('click', () => changePage(-1));
        elements.nextBtn?.addEventListener('click', () => changePage(1));
        elements.saveBtn?.addEventListener('click', saveMember);
        elements.cancelBtn?.addEventListener('click', hideDetail);

        // 加载会员列表
        async function loadMembers() {
            try {
                // 获取筛选条件
                const formData = new FormData(elements.filtersForm);
                const filters = {
                    page: currentPage,
                    per_page: formData.get('per_page') || 20,
                    search: formData.get('search') || '',
                    min_points: formData.get('min_points') || ''
                };

                currentFilters = filters;

                const params = new URLSearchParams(filters);
                const response = await apiRequest(config.listUrl + '?' + params.toString());

                if (!response.ok) {
                    throw new Error(i18n.loadFailed || 'Load failed');
                }

                const data = await response.json();
                renderMembers(data.items || []);
                updatePagination(data.total || 0, data.per_page || 20);

            } catch (error) {
                console.error('Load members error:', error);
                showNotice(i18n.loadFailed || 'Load failed', 'error');
            }
        }

        // 渲染会员列表
        function renderMembers(members) {
            if (!elements.tableBody) return;

            if (members.length === 0) {
                elements.tableBody.innerHTML = '<tr><td colspan="9" style="text-align:center;padding:40px;">' + (i18n.noData || 'No data') + '</td></tr>';
                return;
            }

            elements.tableBody.innerHTML = members.map(member => {
                const tier = getTierByPoints(member.points || 0);
                return `
                    <tr>
                        <td>${member.user_id || '-'}</td>
                        <td>${escapeHtml(member.username || '-')}</td>
                        <td>${escapeHtml(member.full_name || '-')}</td>
                        <td>${escapeHtml(member.email || '-')}</td>
                        <td>${escapeHtml(member.phone || '-')}</td>
                        <td>${member.points || 0}</td>
                        <td>${escapeHtml(tier.label || '-')}</td>
                        <td>${member.registered_at || '-'}</td>
                        <td>
                            <button class="button button-small" onclick="window.tzEditMember(${member.user_id})">编辑</button>
                        </td>
                    </tr>
                `;
            }).join('');
        }

        // 根据积分获取等级
        function getTierByPoints(points) {
            const loyaltySettings = window.TzLoyaltyConfig?.settings || {};
            const tiers = loyaltySettings.tiers || {};

            for (const [key, tier] of Object.entries(tiers)) {
                if (points >= tier.min && (tier.max === null || points <= tier.max)) {
                    return tier;
                }
            }

            return { label: '未知' };
        }

        // 更新分页
        function updatePagination(total, perPage) {
            totalPages = Math.ceil(total / perPage);
            
            if (elements.pageInfo) {
                elements.pageInfo.textContent = `共 ${total} 条`;
            }

            if (elements.prevBtn) {
                elements.prevBtn.disabled = currentPage <= 1;
            }

            if (elements.nextBtn) {
                elements.nextBtn.disabled = currentPage >= totalPages;
            }
        }

        // 翻页
        function changePage(delta) {
            const newPage = currentPage + delta;
            if (newPage >= 1 && newPage <= totalPages) {
                currentPage = newPage;
                loadMembers();
            }
        }

        // 编辑会员
        window.tzEditMember = async function(userId) {
            try {
                const response = await apiRequest(config.singleUrl + userId);
                if (!response.ok) throw new Error('Load failed');

                const member = await response.json();
                
                elements.fieldId.value = member.user_id || '';
                elements.fieldUsername.value = member.username || '';
                elements.fieldEmail.value = member.email || '';
                elements.fieldFullname.value = member.full_name || '';
                elements.fieldPhone.value = member.phone || '';
                elements.fieldCountry.value = member.country || '';
                elements.fieldAddress.value = member.address || '';
                elements.fieldBrand.value = member.brand || '';
                elements.fieldPoints.value = member.points || 0;
                elements.fieldMarketing.checked = !!member.marketing_optin;
                elements.fieldNotes.value = member.notes || '';

                elements.detailSection.style.display = 'block';
                elements.detailSection.scrollIntoView({ behavior: 'smooth' });

            } catch (error) {
                console.error('Load member error:', error);
                showNotice(i18n.loadFailed || 'Load failed', 'error');
            }
        };

        // 保存会员
        async function saveMember(e) {
            e.preventDefault();

            const userId = elements.fieldId.value;
            if (!userId) return;

            const data = {
                full_name: elements.fieldFullname.value,
                phone: elements.fieldPhone.value,
                country: elements.fieldCountry.value,
                address: elements.fieldAddress.value,
                brand: elements.fieldBrand.value,
                points: parseInt(elements.fieldPoints.value) || 0,
                marketing_optin: elements.fieldMarketing.checked ? 1 : 0,
                notes: elements.fieldNotes.value
            };

            try {
                const response = await apiRequest(config.singleUrl + userId, {
                    method: 'PUT',
                    body: JSON.stringify(data)
                });

                if (!response.ok) throw new Error('Save failed');

                showNotice(i18n.saveSuccess || 'Saved', 'success');
                hideDetail();
                loadMembers();

            } catch (error) {
                console.error('Save member error:', error);
                showNotice(i18n.saveFailed || 'Save failed', 'error');
            }
        }

        // 隐藏详情
        function hideDetail() {
            elements.detailSection.style.display = 'none';
            elements.detailForm?.reset();
        }

        // 导出会员
        function exportMembers() {
            window.location.href = config.exportUrl + '&' + new URLSearchParams(currentFilters).toString();
        }

        // HTML 转义
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    });
})();
