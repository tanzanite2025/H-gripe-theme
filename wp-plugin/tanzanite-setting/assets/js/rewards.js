/**
 * Tanzanite Settings - 礼品卡和优惠券管理
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        if (typeof TzRewardsConfig === 'undefined') {
            console.error('Rewards config not found');
            return;
        }

        const config = TzRewardsConfig;
        const i18n = config.i18n || {};
        
        window.TanzaniteAdmin.nonce = config.nonce;
        const { showNotice, apiRequest } = window.TanzaniteAdmin;
        
        // 辅助函数：显示通知
        function showRewardsNotice(message, type) {
            const noticeEl = document.getElementById('tz-rewards-notice');
            if (noticeEl) {
                showNotice(noticeEl, type, message);
            }
        }

        // 标签页切换
        const tabBtns = document.querySelectorAll('.tz-tab-btn');
        const tabContents = document.querySelectorAll('.tz-tab-content');

        tabBtns.forEach(btn => {
            btn.addEventListener('click', function() {
                const tab = this.dataset.tab;
                
                tabBtns.forEach(b => b.classList.remove('active'));
                tabContents.forEach(c => c.style.display = 'none');
                
                this.classList.add('active');
                document.getElementById('tz-tab-' + tab).style.display = 'block';
                
                // 加载对应数据
                if (tab === 'coupons') loadCoupons();
                else if (tab === 'giftcards') loadGiftcards();
                else if (tab === 'transactions') loadTransactions();
            });
        });

        // 优惠券管理
        document.getElementById('tz-coupon-refresh')?.addEventListener('click', loadCoupons);
        document.getElementById('tz-coupon-add')?.addEventListener('click', () => {
            const code = prompt('输入优惠券代码:');
            if (!code) return;
            const title = prompt('输入标题:');
            if (!title) return;
            const amount = parseFloat(prompt('输入金额:') || '0');
            
            createCoupon({ code, title, amount, status: 'active' });
        });

        async function loadCoupons() {
            try {
                const response = await apiRequest(config.couponsListUrl);
                if (!response.ok) throw new Error('Load failed');
                
                let data;
                if (typeof response.json === 'function') {
                    data = await response.json();
                } else {
                    data = response.data || response;
                }
                renderCoupons(data.items || []);
            } catch (error) {
                console.error('Load coupons error:', error);
                const noticeEl = document.getElementById('tz-rewards-notice');
                if (noticeEl) {
                    showNotice(noticeEl, 'error', i18n.loadFailed || 'Load failed');
                }
            }
        }

        function renderCoupons(items) {
            const tbody = document.querySelector('#tz-coupon-table tbody');
            if (!tbody) return;

            if (items.length === 0) {
                tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:40px;">' + (i18n.noData || 'No data') + '</td></tr>';
                return;
            }

            tbody.innerHTML = items.map(item => `
                <tr>
                    <td>${item.id}</td>
                    <td>${escapeHtml(item.code)}</td>
                    <td>${escapeHtml(item.title)}</td>
                    <td>${item.amount}</td>
                    <td>${item.points_required || 0}</td>
                    <td>${item.usage_count || 0}/${item.usage_limit || 1}</td>
                    <td>${escapeHtml(item.status)}</td>
                    <td>
                        <button class="button button-small" onclick="window.tzDeleteCoupon(${item.id})">删除</button>
                    </td>
                </tr>
            `).join('');
        }

        async function createCoupon(data) {
            try {
                const response = await apiRequest(config.couponsListUrl, {
                    method: 'POST',
                    body: JSON.stringify(data)
                });
                
                if (!response.ok) throw new Error('Create failed');
                showRewardsNotice(i18n.saveSuccess || 'Success', 'success');
                loadCoupons();
            } catch (error) {
                console.error('Create coupon error:', error);
                showRewardsNotice(i18n.saveFailed || 'Failed', 'error');
            }
        }

        window.tzDeleteCoupon = async function(id) {
            if (!confirm(i18n.deleteConfirm || 'Delete?')) return;
            
            try {
                const response = await apiRequest(config.couponsSingleUrl + id, { method: 'DELETE' });
                if (!response.ok) throw new Error('Delete failed');
                
                showRewardsNotice(i18n.deleteSuccess || 'Deleted', 'success');
                loadCoupons();
            } catch (error) {
                console.error('Delete coupon error:', error);
                showRewardsNotice('删除失败', 'error');
            }
        };

        // 礼品卡管理
        document.getElementById('tz-giftcard-refresh')?.addEventListener('click', loadGiftcards);
        document.getElementById('tz-giftcard-add')?.addEventListener('click', () => {
            const code = prompt('输入礼品卡号:');
            if (!code) return;
            const amount = parseFloat(prompt('输入金额:') || '0');
            const pointsSpent = parseInt(prompt('输入消耗的积分 (可选，默认0):') || '0');
            
            createGiftcard({ 
                card_code: code, 
                balance: amount, 
                original_value: amount, 
                points_spent: pointsSpent,
                status: 'active' 
            });
        });

        async function loadGiftcards() {
            try {
                const response = await apiRequest(config.giftcardsListUrl);
                if (!response.ok) throw new Error('Load failed');
                
                let data;
                if (typeof response.json === 'function') {
                    data = await response.json();
                } else {
                    data = response.data || response;
                }
                renderGiftcards(data.items || []);
            } catch (error) {
                console.error('Load giftcards error:', error);
                showRewardsNotice(i18n.loadFailed || 'Load failed', 'error');
            }
        }

        function renderGiftcards(items) {
            const tbody = document.querySelector('#tz-giftcard-table tbody');
            if (!tbody) return;

            if (items.length === 0) {
                tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:40px;">' + (i18n.noData || 'No data') + '</td></tr>';
                return;
            }

            tbody.innerHTML = items.map(item => `
                <tr>
                    <td>${item.id}</td>
                    <td>${escapeHtml(item.card_code)}</td>
                    <td>$${item.balance}</td>
                    <td>$${item.original_value}</td>
                    <td>${item.points_spent || 0}</td>
                    <td>${item.owner_user_id || '-'}</td>
                    <td>${escapeHtml(item.status)}</td>
                    <td>
                        <button class="button button-small" onclick="window.tzEditGiftcard(${item.id}, ${item.points_spent || 0})">编辑</button>
                        <button class="button button-small" onclick="window.tzDeleteGiftcard(${item.id})">删除</button>
                    </td>
                </tr>
            `).join('');
        }

        async function createGiftcard(data) {
            try {
                const response = await apiRequest(config.giftcardsListUrl, {
                    method: 'POST',
                    body: JSON.stringify(data)
                });
                
                if (!response.ok) {
                    const errorMsg = response.data?.message || response.data?.code || 'Create failed';
                    console.error('Server error:', response);
                    throw new Error(errorMsg);
                }
                showRewardsNotice(i18n.saveSuccess || 'Success', 'success');
                loadGiftcards();
            } catch (error) {
                console.error('Create giftcard error:', error);
                showRewardsNotice(error.message || i18n.saveFailed || 'Failed', 'error');
            }
        }

        window.tzEditGiftcard = async function(id, currentPoints) {
            const newPoints = parseInt(prompt('输入新的积分价格 (当前: ' + currentPoints + '):', currentPoints) || '0');
            if (newPoints === currentPoints) return;
            
            try {
                const response = await apiRequest(config.giftcardsSingleUrl + id, {
                    method: 'PUT',
                    body: JSON.stringify({ points_spent: newPoints })
                });
                
                if (!response.ok) {
                    const errorMsg = response.data?.message || response.data?.code || 'Update failed';
                    console.error('Server error:', response);
                    throw new Error(errorMsg);
                }
                
                showRewardsNotice('更新成功', 'success');
                loadGiftcards();
            } catch (error) {
                console.error('Edit giftcard error:', error);
                showRewardsNotice(error.message || '更新失败', 'error');
            }
        };

        window.tzDeleteGiftcard = async function(id) {
            if (!confirm(i18n.deleteConfirm || 'Delete?')) return;
            
            try {
                const response = await apiRequest(config.giftcardsSingleUrl + id, { method: 'DELETE' });
                
                if (!response.ok) {
                    const errorMsg = response.data?.message || response.data?.code || 'Delete failed';
                    console.error('Server error:', response);
                    throw new Error(errorMsg);
                }
                
                showRewardsNotice(i18n.deleteSuccess || 'Deleted', 'success');
                loadGiftcards();
            } catch (error) {
                console.error('Delete giftcard error:', error);
                showRewardsNotice(error.message || '删除失败', 'error');
            }
        };

        // 交易记录
        document.getElementById('tz-transaction-refresh')?.addEventListener('click', loadTransactions);

        async function loadTransactions() {
            try {
                const response = await apiRequest(config.transactionsUrl);
                if (!response.ok) throw new Error('Load failed');
                
                let data;
                if (typeof response.json === 'function') {
                    data = await response.json();
                } else {
                    data = response.data || response;
                }
                renderTransactions(data.items || []);
            } catch (error) {
                console.error('Load transactions error:', error);
                const noticeEl = document.getElementById('tz-rewards-notice');
                if (noticeEl) {
                    showNotice(noticeEl, 'error', i18n.loadFailed || 'Load failed');
                }
            }
        }

        function renderTransactions(items) {
            const tbody = document.querySelector('#tz-transaction-table tbody');
            if (!tbody) return;

            if (items.length === 0) {
                tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:40px;">' + (i18n.noData || 'No data') + '</td></tr>';
                return;
            }

            tbody.innerHTML = items.map(item => `
                <tr>
                    <td>${item.id}</td>
                    <td>${item.user_id || '-'}</td>
                    <td>${escapeHtml(item.related_type)}</td>
                    <td>${escapeHtml(item.action)}</td>
                    <td>${item.points_delta > 0 ? '+' : ''}${item.points_delta}</td>
                    <td>${item.amount_delta > 0 ? '+' : ''}${item.amount_delta}</td>
                    <td>${escapeHtml(item.notes || '-')}</td>
                    <td>${item.created_at}</td>
                </tr>
            `).join('');
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // 初始加载优惠券
        loadCoupons();
    });
})();
