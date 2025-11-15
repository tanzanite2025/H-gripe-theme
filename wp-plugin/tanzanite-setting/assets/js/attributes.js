/**
 * Tanzanite Settings - 商品属性管理页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    function init() {
        // 检查配置是否存在
        if (typeof TzAttributesConfig === 'undefined') {
            console.error('Attributes config not found');
            return;
        }

        const config = TzAttributesConfig;
        
        // 检查 TanzaniteAdmin 是否存在，提供回退
        if (typeof window.TanzaniteAdmin === 'undefined') {
            console.warn('TanzaniteAdmin not found! Using fallback functions.');
            window.TanzaniteAdmin = {
                showNotice: function(el, type, msg) {
                    if (!el) return;
                    if (!type) { el.style.display = 'none'; return; }
                    el.className = 'notice notice-' + type;
                    el.textContent = msg;
                    el.style.display = 'block';
                },
                apiRequest: async function(url, options = {}) {
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
                },
                scrollToElement: function(el) {
                    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }
            };
        } else {
            window.TanzaniteAdmin.nonce = config.nonce;
        }
        
        const { showNotice, apiRequest, scrollToElement } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            attrTable: document.querySelector('#tz-attr-table tbody'),
            valuesTable: document.querySelector('#tz-values-table tbody'),
            notice: document.getElementById('tz-attr-notice'),
            form: document.getElementById('tz-attr-form'),
            valuesSection: document.getElementById('tz-attr-values-section'),
            currentAttrName: document.getElementById('tz-current-attr-name'),
            createBtn: document.getElementById('tz-attr-create'),
            saveBtn: document.getElementById('tz-attr-save'),
            resetBtn: document.getElementById('tz-attr-reset'),
            valueAddBtn: document.getElementById('tz-value-add'),
            valueFieldLabel: document.getElementById('tz-value-field-label'),
            inputs: {
                id: document.getElementById('tz-attr-id'),
                name: document.getElementById('tz-attr-name'),
                slug: document.getElementById('tz-attr-slug'),
                type: document.getElementById('tz-attr-type'),
                sort: document.getElementById('tz-attr-sort'),
                filterable: document.getElementById('tz-attr-filterable'),
                sku: document.getElementById('tz-attr-sku'),
                stock: document.getElementById('tz-attr-stock'),
                enabled: document.getElementById('tz-attr-enabled')
            },
            valueInputs: {
                name: document.getElementById('tz-value-name'),
                slug: document.getElementById('tz-value-slug'),
                value: document.getElementById('tz-value-value')
            }
        };

        // 检查必需元素
        if (!elements.attrTable || !elements.form) {
            console.error('Required elements not found');
            return;
        }

        let currentAttrId = null;

        /**
         * 重置表单
         */
        function resetForm() {
            elements.form.reset();
            elements.inputs.id.value = '';
            elements.valuesSection.style.display = 'none';
            currentAttrId = null;
            showNotice(elements.notice, null);
        }

        /**
         * 填充表单
         */
        function fillForm(data) {
            resetForm();
            elements.inputs.id.value = data.id || '';
            elements.inputs.name.value = data.name || '';
            elements.inputs.slug.value = data.slug || '';
            elements.inputs.type.value = data.type || 'select';
            elements.inputs.sort.value = data.sort_order || 0;
            elements.inputs.filterable.checked = data.is_filterable;
            elements.inputs.sku.checked = data.affects_sku;
            elements.inputs.stock.checked = data.affects_stock;
            elements.inputs.enabled.checked = data.is_enabled;
            
            currentAttrId = data.id;
            elements.currentAttrName.textContent = data.name;
            elements.valuesSection.style.display = 'block';
            
            fetchValues(data.id);
        }

        /**
         * 渲染属性列表
         */
        function renderAttrs(items) {
            elements.attrTable.innerHTML = '';

            if (!items || items.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="7" style="text-align:center;">暂无属性</td>';
                elements.attrTable.appendChild(tr);
                return;
            }

            const typeMap = { select: '下拉', color: '色块', image: '图标' };

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                
                tr.innerHTML = `
                    <td>${item.name}</td>
                    <td>${typeMap[item.type] || item.type}</td>
                    <td>${item.values_count || 0}</td>
                    <td>${item.is_filterable ? '✓' : '-'}</td>
                    <td>${item.affects_sku ? '✓' : '-'}</td>
                    <td>${item.is_enabled ? '启用' : '禁用'}</td>
                    <td>
                        <button class="button-link edit-attr" data-id="${item.id}">编辑</button> | 
                        <button class="button-link-delete del-attr" data-id="${item.id}">删除</button>
                    </td>
                `;
                
                elements.attrTable.appendChild(tr);
            });
        }

        /**
         * 渲染属性值列表
         */
        function renderValues(items) {
            elements.valuesTable.innerHTML = '';

            if (!items || items.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="6" style="text-align:center;">暂无属性值</td>';
                elements.valuesTable.appendChild(tr);
                return;
            }

            const type = elements.inputs.type.value;

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                
                let preview = '-';
                if (item.value) {
                    if (type === 'color') {
                        preview = '<span style="display:inline-block;width:24px;height:24px;background:' + item.value + ';border:1px solid #ddd;border-radius:4px;"></span>';
                    } else if (type === 'image') {
                        preview = '<img src="' + item.value + '" style="max-width:40px;max-height:24px;" />';
                    } else {
                        preview = item.value;
                    }
                }
                
                tr.innerHTML = `
                    <td>${item.name}</td>
                    <td>${item.slug}</td>
                    <td>${preview}</td>
                    <td>${item.sort_order}</td>
                    <td>${item.is_enabled ? '启用' : '禁用'}</td>
                    <td>
                        <button class="button-link-delete del-value" data-id="${item.id}">删除</button>
                    </td>
                `;
                
                elements.valuesTable.appendChild(tr);
            });
        }

        /**
         * 获取属性列表
         */
        async function fetchAttrs() {
            try {
                const result = await apiRequest(config.attrUrl);
                
                if (!result.ok) {
                    renderAttrs([]);
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                renderAttrs(result.data.items || []);
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 获取单个属性
         */
        async function fetchAttr(id) {
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
         * 保存属性
         */
        async function saveAttr(e) {
            e.preventDefault();

            const id = elements.inputs.id.value;
            
            const payload = {
                name: elements.inputs.name.value,
                slug: elements.inputs.slug.value,
                type: elements.inputs.type.value,
                sort_order: parseInt(elements.inputs.sort.value) || 0,
                is_filterable: elements.inputs.filterable.checked,
                affects_sku: elements.inputs.sku.checked,
                affects_stock: elements.inputs.stock.checked,
                is_enabled: elements.inputs.enabled.checked
            };

            const url = id ? config.singleUrl + id : config.attrUrl;
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
                
                if (!id) {
                    fillForm(result.data);
                }
                
                fetchAttrs();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 删除属性
         */
        async function deleteAttr(id) {
            if (!confirm('确定删除？这将同时删除所有属性值。')) {
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
                resetForm();
                fetchAttrs();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 获取属性值列表
         */
        async function fetchValues(attrId) {
            try {
                const result = await apiRequest(config.singleUrl + attrId + '/values');
                
                if (!result.ok) {
                    renderValues([]);
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                renderValues(result.data.items || []);
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 添加属性值
         */
        async function addValue() {
            if (!currentAttrId) {
                showNotice(elements.notice, 'error', '请先选择属性组');
                return;
            }

            const payload = {
                name: elements.valueInputs.name.value,
                slug: elements.valueInputs.slug.value,
                value: elements.valueInputs.value.value,
                is_enabled: true,
                sort_order: 0
            };

            try {
                const result = await apiRequest(config.singleUrl + currentAttrId + '/values', {
                    method: 'POST',
                    body: JSON.stringify(payload)
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '添加失败');
                    return;
                }

                showNotice(elements.notice, 'success', '已添加');
                elements.valueInputs.name.value = '';
                elements.valueInputs.slug.value = '';
                elements.valueInputs.value.value = '';
                
                fetchValues(currentAttrId);
                fetchAttrs();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 删除属性值
         */
        async function deleteValue(id) {
            if (!confirm('确定删除？')) {
                return;
            }

            try {
                const result = await apiRequest(config.singleUrl + currentAttrId + '/values/' + id, {
                    method: 'DELETE'
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '删除失败');
                    return;
                }

                showNotice(elements.notice, 'success', '已删除');
                fetchValues(currentAttrId);
                fetchAttrs();
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 更新值字段标签
         */
        function updateValueFieldLabel() {
            const type = elements.inputs.type.value;
            let labelHtml = '';
            
            if (type === 'color') {
                labelHtml = '色值<input type="text" id="tz-value-value" placeholder="#FF0000" />';
            } else if (type === 'image') {
                labelHtml = '图标URL<input type="text" id="tz-value-value" placeholder="https://..." />';
            } else {
                labelHtml = '值<input type="text" id="tz-value-value" />';
            }
            
            elements.valueFieldLabel.innerHTML = labelHtml;
            elements.valueInputs.value = document.getElementById('tz-value-value');
        }

        // 事件监听
        elements.attrTable.addEventListener('click', function(e) {
            const target = e.target;
            
            if (target.classList.contains('edit-attr')) {
                fetchAttr(target.dataset.id);
            }
            
            if (target.classList.contains('del-attr')) {
                deleteAttr(target.dataset.id);
            }
        });

        elements.valuesTable.addEventListener('click', function(e) {
            if (e.target.classList.contains('del-value')) {
                deleteValue(e.target.dataset.id);
            }
        });

        elements.createBtn.addEventListener('click', function() {
            resetForm();
            scrollToElement(elements.form);
        });

        elements.form.addEventListener('submit', saveAttr);
        elements.saveBtn.addEventListener('click', saveAttr);
        elements.resetBtn.addEventListener('click', function(e) {
            e.preventDefault();
            resetForm();
        });

        elements.valueAddBtn.addEventListener('click', addValue);
        elements.inputs.type.addEventListener('change', updateValueFieldLabel);

        // 初始化
        fetchAttrs();
    }

    // 如果 DOM 已经加载完成，立即执行；否则等待 DOMContentLoaded
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
