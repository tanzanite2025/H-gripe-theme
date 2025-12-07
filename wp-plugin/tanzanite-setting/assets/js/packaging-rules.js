/**
 * Tanzanite Settings - 包装规则管理页面
 * 
 * @package TanzaniteSettings
 * @version 0.3.0
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzPackagingConfig === 'undefined') {
            console.error('Packaging Rules config not found');
            return;
        }

        const config = TzPackagingConfig;
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest, scrollToElement, confirm } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            tableBody: document.querySelector('#tz-packaging-table tbody'),
            notice: document.getElementById('tz-packaging-notice'),
            form: document.getElementById('tz-packaging-form'),
            installSection: document.getElementById('tz-packaging-install-section'),
            installBtn: document.getElementById('tz-packaging-install'),
            createBtn: document.getElementById('tz-packaging-create'),
            saveBtn: document.getElementById('tz-packaging-save'),
            resetBtn: document.getElementById('tz-packaging-reset'),
            appliesList: document.getElementById('tz-packaging-applies-list'),
            addApplyBtn: document.getElementById('tz-packaging-add-apply'),
            inputs: {
                id: document.getElementById('tz-packaging-id'),
                name: document.getElementById('tz-packaging-name'),
                description: document.getElementById('tz-packaging-description'),
                priority: document.getElementById('tz-packaging-priority'),
                active: document.getElementById('tz-packaging-active'),
                boxWeight: document.getElementById('tz-packaging-box-weight'),
                boxLength: document.getElementById('tz-packaging-box-length'),
                boxWidth: document.getElementById('tz-packaging-box-width'),
                boxHeight: document.getElementById('tz-packaging-box-height'),
                maxItems: document.getElementById('tz-packaging-max-items'),
                maxWeight: document.getElementById('tz-packaging-max-weight'),
                applyType: document.getElementById('tz-packaging-apply-type'),
                applyValue: document.getElementById('tz-packaging-apply-value')
            }
        };

        // 检查必需元素
        if (!elements.tableBody || !elements.form) {
            console.error('Required elements not found');
            return;
        }

        // 适用范围列表
        let appliesTo = [];

        const applyTypeLabels = {
            category: '分类',
            tag: '标签',
            product: '商品ID',
            all: '所有商品'
        };

        /**
         * 重置表单
         */
        function resetForm() {
            elements.form.reset();
            elements.inputs.id.value = '';
            elements.inputs.priority.value = '0';
            elements.inputs.active.value = '1';
            appliesTo = [];
            renderAppliesTo();
            showNotice(elements.notice, null);
        }

        /**
         * 填充表单
         */
        function fillForm(data) {
            resetForm();
            elements.inputs.id.value = data.id || '';
            elements.inputs.name.value = data.rule_name || '';
            elements.inputs.description.value = data.description || '';
            elements.inputs.priority.value = data.priority || 0;
            elements.inputs.active.value = data.is_active ? '1' : '0';
            elements.inputs.boxWeight.value = data.box_weight || '';
            elements.inputs.boxLength.value = data.box_length || '';
            elements.inputs.boxWidth.value = data.box_width || '';
            elements.inputs.boxHeight.value = data.box_height || '';
            elements.inputs.maxItems.value = data.max_items || '';
            elements.inputs.maxWeight.value = data.max_weight || '';
            appliesTo = data.applies_to || [];
            renderAppliesTo();
        }

        /**
         * 渲染适用范围列表
         */
        function renderAppliesTo() {
            elements.appliesList.innerHTML = '';

            if (!appliesTo.length) {
                elements.appliesList.innerHTML = '<p class="description">暂未指定适用范围，将不会匹配任何商品。</p>';
                return;
            }

            appliesTo.forEach(function(apply, idx) {
                const div = document.createElement('div');
                div.style.cssText = 'display:inline-flex;align-items:center;gap:8px;padding:6px 12px;background:#e5e7eb;border-radius:4px;margin-right:8px;margin-bottom:8px;';

                const typeLabel = applyTypeLabels[apply.type] || apply.type;
                const valueText = apply.type === 'all' ? '' : ': ' + (apply.value || '');

                div.innerHTML = `
                    <span>${typeLabel}${valueText}</span>
                    <button type="button" class="button-link-delete del-apply" data-idx="${idx}" style="color:#dc2626;text-decoration:none;">×</button>
                `;

                elements.appliesList.appendChild(div);
            });
        }

        /**
         * 添加适用范围
         */
        function addApply() {
            const type = elements.inputs.applyType.value;
            const value = elements.inputs.applyValue.value.trim();

            if (type !== 'all' && !value) {
                alert('请输入值');
                return;
            }

            // 检查是否已存在
            const exists = appliesTo.some(function(a) {
                return a.type === type && a.value === (type === 'all' ? null : value);
            });

            if (exists) {
                alert('该适用范围已存在');
                return;
            }

            appliesTo.push({
                type: type,
                value: type === 'all' ? null : value
            });

            elements.inputs.applyValue.value = '';
            renderAppliesTo();
        }

        /**
         * 删除适用范围
         */
        function deleteApply(idx) {
            appliesTo.splice(idx, 1);
            renderAppliesTo();
        }

        /**
         * 渲染表格
         */
        function renderTable(items) {
            elements.tableBody.innerHTML = '';

            if (!items || items.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="8" style="text-align:center;">暂无包装规则</td>';
                elements.tableBody.appendChild(tr);
                return;
            }

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                
                const statusText = item.is_active 
                    ? '<span style="color:#16a34a;">启用</span>' 
                    : '<span style="color:#9ca3af;">禁用</span>';

                // 包装尺寸
                let sizeText = '-';
                if (item.box_length && item.box_width && item.box_height) {
                    sizeText = item.box_length + '×' + item.box_width + '×' + item.box_height + ' cm';
                }

                // 适用范围
                let appliesText = '-';
                if (item.applies_to && item.applies_to.length > 0) {
                    appliesText = item.applies_to.map(function(a) {
                        const label = applyTypeLabels[a.type] || a.type;
                        return a.type === 'all' ? label : label + ':' + a.value;
                    }).join(', ');
                }

                tr.innerHTML = `
                    <td><strong>${item.rule_name}</strong></td>
                    <td>${item.box_weight} kg</td>
                    <td>${sizeText}</td>
                    <td>${item.max_items || '无限制'}</td>
                    <td>${appliesText}</td>
                    <td>${item.priority}</td>
                    <td>${statusText}</td>
                    <td>
                        <button class="button-link edit-rule" data-id="${item.id}">编辑</button> | 
                        <button class="button-link-delete del-rule" data-id="${item.id}">删除</button>
                    </td>
                `;

                elements.tableBody.appendChild(tr);
            });
        }

        /**
         * 获取列表
         */
        async function fetchList() {
            try {
                const result = await apiRequest(config.listUrl);
                
                if (!result.ok) {
                    renderTable([]);
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                // 检查是否需要安装数据库
                if (result.data.table_missing) {
                    elements.installSection.style.display = 'block';
                    renderTable([]);
                    return;
                }

                elements.installSection.style.display = 'none';
                renderTable(result.data.items || []);
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 安装数据库表
         */
        async function installTables() {
            try {
                elements.installBtn.disabled = true;
                elements.installBtn.textContent = '安装中...';

                const result = await apiRequest(config.installUrl, {
                    method: 'POST'
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '安装失败');
                    elements.installBtn.disabled = false;
                    elements.installBtn.textContent = '安装数据库表';
                    return;
                }

                showNotice(elements.notice, 'success', '数据库表安装成功');
                elements.installSection.style.display = 'none';
                fetchList();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
                elements.installBtn.disabled = false;
                elements.installBtn.textContent = '安装数据库表';
            }
        }

        /**
         * 获取单个
         */
        async function fetchSingle(id) {
            try {
                const result = await apiRequest(config.singleUrl + id);
                
                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
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
         */
        async function saveForm(e) {
            e.preventDefault();

            const id = elements.inputs.id.value;

            const payload = {
                rule_name: elements.inputs.name.value,
                description: elements.inputs.description.value,
                priority: parseInt(elements.inputs.priority.value) || 0,
                is_active: elements.inputs.active.value === '1',
                box_weight: parseFloat(elements.inputs.boxWeight.value) || 0,
                box_length: elements.inputs.boxLength.value ? parseFloat(elements.inputs.boxLength.value) : null,
                box_width: elements.inputs.boxWidth.value ? parseFloat(elements.inputs.boxWidth.value) : null,
                box_height: elements.inputs.boxHeight.value ? parseFloat(elements.inputs.boxHeight.value) : null,
                max_items: elements.inputs.maxItems.value ? parseInt(elements.inputs.maxItems.value) : null,
                max_weight: elements.inputs.maxWeight.value ? parseFloat(elements.inputs.maxWeight.value) : null,
                applies_to: appliesTo
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
         * 删除规则
         */
        async function deleteRule(id) {
            if (!confirm('确定删除该包装规则？')) {
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
            
            if (target.classList.contains('edit-rule')) {
                fetchSingle(target.dataset.id);
            }
            
            if (target.classList.contains('del-rule')) {
                deleteRule(target.dataset.id);
            }
        });

        elements.appliesList.addEventListener('click', function(e) {
            const target = e.target;
            
            if (target.classList.contains('del-apply')) {
                deleteApply(parseInt(target.dataset.idx));
            }
        });

        if (elements.installBtn) {
            elements.installBtn.addEventListener('click', installTables);
        }

        elements.createBtn.addEventListener('click', function() {
            resetForm();
            scrollToElement(elements.form);
        });

        elements.addApplyBtn.addEventListener('click', addApply);

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
