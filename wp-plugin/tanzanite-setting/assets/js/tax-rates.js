/**
 * Tanzanite Settings - 税率管理页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzTaxRatesConfig === 'undefined') {
            console.error('Tax Rates config not found');
            return;
        }

        const config = TzTaxRatesConfig;
        
        // 检查 TanzaniteAdmin 是否存在
        let showNotice, apiRequest, scrollToElement, confirm;
        
        if (typeof window.TanzaniteAdmin === 'undefined') {
            console.warn('TanzaniteAdmin not found! Using fallback functions.');
            
            // 提供基本的回退函数
            showNotice = function(el, type, msg) {
                if (!el) return;
                if (!type) {
                    el.style.display = 'none';
                    return;
                }
                el.className = 'notice notice-' + type;
                el.textContent = msg;
                el.style.display = 'block';
            };
            
            apiRequest = async function(url, options = {}) {
                const response = await fetch(url, {
                    ...options,
                    headers: {
                        'Content-Type': 'application/json',
                        'X-WP-Nonce': config.nonce,
                        ...options.headers
                    }
                });
                const data = await response.json();
                return { ok: response.ok, data };
            };
            
            scrollToElement = function(el) {
                if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
            };
            
            confirm = function(msg) {
                return window.confirm(msg);
            };
        } else {
            window.TanzaniteAdmin.nonce = config.nonce;
            ({ showNotice, apiRequest, scrollToElement, confirm } = window.TanzaniteAdmin);
        }

        // DOM 元素
        const elements = {
            tableBody: document.querySelector('#tz-tax-rate-table tbody'),
            notice: document.getElementById('tz-tax-rate-notice'),
            form: document.getElementById('tz-tax-rate-form'),
            createBtn: document.getElementById('tz-tax-rate-create'),
            saveBtn: document.getElementById('tz-tax-rate-save'),
            resetBtn: document.getElementById('tz-tax-rate-reset'),
            inputs: {
                id: document.getElementById('tz-tax-rate-id'),
                name: document.getElementById('tz-tax-rate-name'),
                rate: document.getElementById('tz-tax-rate-rate'),
                region: document.getElementById('tz-tax-rate-region'),
                description: document.getElementById('tz-tax-rate-description'),
                sort: document.getElementById('tz-tax-rate-sort'),
                active: document.getElementById('tz-tax-rate-active'),
                meta: document.getElementById('tz-tax-rate-meta')
            }
        };

        // 检查必需元素
        if (!elements.tableBody || !elements.form) {
            console.error('Required elements not found');
            return;
        }

        /**
         * 重置表单
         */
        function resetForm() {
            elements.form.reset();
            elements.inputs.id.value = '';
            elements.inputs.sort.value = '0';
            elements.inputs.active.value = '1';
            showNotice(elements.notice, null);
        }

        /**
         * 填充表单
         * @param {Object} data - 税率数据
         */
        function fillForm(data) {
            resetForm();
            elements.inputs.id.value = data.id || '';
            elements.inputs.name.value = data.name || '';
            elements.inputs.rate.value = data.rate || 0;
            elements.inputs.region.value = data.region || '';
            elements.inputs.description.value = data.description || '';
            elements.inputs.sort.value = data.sort_order || 0;
            elements.inputs.active.value = data.is_active ? '1' : '0';
            elements.inputs.meta.value = data.meta ? JSON.stringify(data.meta, null, 2) : '';
        }

        /**
         * 渲染表格行
         * @param {Array} items - 税率列表
         */
        function renderRows(items) {
            elements.tableBody.innerHTML = '';

            if (!items || items.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="5" style="text-align:center;">暂无税率</td>';
                elements.tableBody.appendChild(tr);
                return;
            }

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                
                const statusText = item.is_active 
                    ? '<span style="color:#16a34a;">启用</span>' 
                    : '<span style="color:#9ca3af;">禁用</span>';
                
                tr.innerHTML = `
                    <td><strong>${item.name}</strong></td>
                    <td>${item.rate}%</td>
                    <td>${item.region || '-'}</td>
                    <td>${statusText}</td>
                    <td>
                        <button class="button-link tz-tax-rate-edit" data-id="${item.id}">编辑</button> | 
                        <button class="button-link-delete tz-tax-rate-delete" data-id="${item.id}">删除</button>
                    </td>
                `;
                
                elements.tableBody.appendChild(tr);
            });
        }

        /**
         * 获取税率列表
         */
        async function fetchList() {
            showNotice(elements.notice, null);

            try {
                const result = await apiRequest(config.listUrl);
                
                if (!result.ok) {
                    renderRows([]);
                    showNotice(elements.notice, 'error', result.data.message || '加载税率失败');
                    return;
                }

                renderRows(result.data.items || []);
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 获取单个税率
         * @param {number} id - 税率 ID
         */
        async function fetchSingle(id) {
            try {
                const result = await apiRequest(config.singleUrl + id);
                
                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '加载税率失败');
                    return;
                }

                fillForm(result.data);
                scrollToElement(elements.form);
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 保存表单
         * @param {Event} e - 事件对象
         */
        async function saveForm(e) {
            e.preventDefault();

            const id = elements.inputs.id.value;
            let meta = null;

            // 解析 meta JSON
            if (elements.inputs.meta.value.trim()) {
                try {
                    meta = JSON.parse(elements.inputs.meta.value);
                } catch (err) {
                    showNotice(elements.notice, 'error', 'Meta 信息格式错误，请输入有效的 JSON');
                    return;
                }
            }

            const payload = {
                name: elements.inputs.name.value,
                rate: parseFloat(elements.inputs.rate.value),
                region: elements.inputs.region.value,
                description: elements.inputs.description.value,
                sort_order: parseInt(elements.inputs.sort.value) || 0,
                is_active: elements.inputs.active.value === '1',
                meta: meta
            };

            const url = id ? config.singleUrl + id : config.listUrl;
            const method = id ? 'PUT' : 'POST';

            try {
                const result = await apiRequest(url, {
                    method: method,
                    body: JSON.stringify(payload)
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '保存失败');
                    return;
                }

                showNotice(elements.notice, 'success', '保存成功');
                resetForm();
                fetchList();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 删除税率
         * @param {number} id - 税率 ID
         */
        async function deleteTaxRate(id) {
            if (!confirm('确定要删除该税率吗？')) {
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

                showNotice(elements.notice, 'success', '已删除');
                fetchList();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        // 事件监听
        elements.tableBody.addEventListener('click', function(e) {
            const target = e.target;
            
            if (target.classList.contains('tz-tax-rate-edit')) {
                e.preventDefault();
                fetchSingle(target.dataset.id);
            }
            
            if (target.classList.contains('tz-tax-rate-delete')) {
                e.preventDefault();
                deleteTaxRate(target.dataset.id);
            }
        });

        elements.createBtn.addEventListener('click', function() {
            resetForm();
            scrollToElement(elements.form);
        });

        elements.form.addEventListener('submit', saveForm);
        elements.saveBtn.addEventListener('click', saveForm);
        elements.resetBtn.addEventListener('click', function(e) {
            e.preventDefault();
            resetForm();
        });

        // 初始化
        fetchList();
    });
})();
