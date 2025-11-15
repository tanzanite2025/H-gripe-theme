/**
 * Tanzanite Settings - 配送模板管理页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzShippingConfig === 'undefined') {
            console.error('Shipping Templates config not found');
            return;
        }

        const config = TzShippingConfig;
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest, scrollToElement, confirm } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            tableBody: document.querySelector('#tz-shipping-table tbody'),
            notice: document.getElementById('tz-shipping-notice'),
            form: document.getElementById('tz-shipping-form'),
            rulesList: document.getElementById('tz-shipping-rules-list'),
            createBtn: document.getElementById('tz-shipping-create'),
            saveBtn: document.getElementById('tz-shipping-save'),
            resetBtn: document.getElementById('tz-shipping-reset'),
            exportBtn: document.getElementById('tz-shipping-export'),
            addRuleBtn: document.getElementById('tz-shipping-add-rule'),
            inputs: {
                id: document.getElementById('tz-shipping-id'),
                name: document.getElementById('tz-shipping-name'),
                description: document.getElementById('tz-shipping-description'),
                active: document.getElementById('tz-shipping-active')
            }
        };

        // 检查必需元素
        if (!elements.tableBody || !elements.form) {
            console.error('Required elements not found');
            return;
        }

        let rules = [];

        const ruleTypes = {
            weight: '按重量',
            amount: '按金额',
            quantity: '按件数',
            volume: '按体积',
            items: '按商品数'
        };

        /**
         * 重置表单
         */
        function resetForm() {
            elements.form.reset();
            elements.inputs.id.value = '';
            rules = [];
            renderRules();
            showNotice(elements.notice, null);
        }

        /**
         * 填充表单
         */
        function fillForm(data) {
            resetForm();
            elements.inputs.id.value = data.id || '';
            elements.inputs.name.value = data.template_name || '';
            elements.inputs.description.value = data.description || '';
            elements.inputs.active.value = data.is_active ? '1' : '0';
            rules = data.rules || [];
            renderRules();
        }

        /**
         * 渲染规则列表
         */
        function renderRules() {
            elements.rulesList.innerHTML = '';

            if (!rules.length) {
                elements.rulesList.innerHTML = '<p class="description">暂无规则，点击下方按钮添加。</p>';
                return;
            }

            rules.forEach(function(rule, idx) {
                const div = document.createElement('div');
                div.style.cssText = 'padding:12px;background:#f9fafb;border:1px solid #e5e7eb;border-radius:6px;margin-bottom:8px;';

                const typeText = ruleTypes[rule.type] || rule.type;
                let rangeText = '';

                if (rule.min !== null && rule.max !== null) {
                    rangeText = rule.min + ' - ' + rule.max;
                } else if (rule.min !== null) {
                    rangeText = '≥ ' + rule.min;
                } else if (rule.max !== null) {
                    rangeText = '≤ ' + rule.max;
                }

                const freeOverText = rule.free_over ? ' (满¥' + rule.free_over + '包邮)' : '';

                div.innerHTML = `
                    <div style="display:flex;justify-content:space-between;align-items:center;">
                        <div>
                            <strong>${typeText}</strong> ${rangeText} → 运费: ¥${rule.fee}${freeOverText}
                        </div>
                        <div>
                            <button type="button" class="button button-small edit-rule" data-idx="${idx}">编辑</button>
                            <button type="button" class="button-link-delete del-rule" data-idx="${idx}">删除</button>
                        </div>
                    </div>
                `;

                elements.rulesList.appendChild(div);
            });
        }

        /**
         * 添加或更新规则
         */
        function addOrUpdateRule(ruleData, editIdx) {
            if (editIdx !== null && editIdx >= 0) {
                rules[editIdx] = ruleData;
            } else {
                rules.push(ruleData);
            }
            renderRules();
        }

        /**
         * 删除规则
         */
        function deleteRule(idx) {
            if (!confirm('确定删除该规则？')) {
                return;
            }
            rules.splice(idx, 1);
            renderRules();
        }

        /**
         * 添加规则（简化版）
         */
        function promptAddRule() {
            const type = prompt('规则类型：\n1=按重量\n2=按金额\n3=按件数\n4=按体积\n5=按商品数', '1');
            if (!type) return;

            const typeMap = {
                '1': 'weight',
                '2': 'amount',
                '3': 'quantity',
                '4': 'volume',
                '5': 'items'
            };

            const selectedType = typeMap[type];
            if (!selectedType) {
                alert('无效类型');
                return;
            }

            const min = prompt('最小值（留空表示无限制）', '');
            const max = prompt('最大值（留空表示无限制）', '');
            const fee = prompt('运费（元）', '10');
            const freeOver = prompt('满多少金额包邮（留空表示不包邮）', '');

            const rule = {
                type: selectedType,
                min: min ? parseFloat(min) : null,
                max: max ? parseFloat(max) : null,
                fee: parseFloat(fee) || 0,
                priority: 0,
                free_over: freeOver ? parseFloat(freeOver) : null
            };

            addOrUpdateRule(rule, null);
        }

        /**
         * 渲染表格
         */
        function renderTable(items) {
            elements.tableBody.innerHTML = '';

            if (!items || items.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="5" style="text-align:center;">暂无模板</td>';
                elements.tableBody.appendChild(tr);
                return;
            }

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                
                const statusText = item.is_active 
                    ? '<span style="color:#16a34a;">启用</span>' 
                    : '<span style="color:#9ca3af;">禁用</span>';

                tr.innerHTML = `
                    <td><strong>${item.template_name}</strong></td>
                    <td>${item.description || '-'}</td>
                    <td>${(item.rules || []).length}</td>
                    <td>${statusText}</td>
                    <td>
                        <button class="button-link edit-tpl" data-id="${item.id}">编辑</button> | 
                        <button class="button-link copy-tpl" data-id="${item.id}">复制</button> | 
                        <button class="button-link-delete del-tpl" data-id="${item.id}">删除</button>
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

                renderTable(result.data.items || []);
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
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
                template_name: elements.inputs.name.value,
                description: elements.inputs.description.value,
                is_active: elements.inputs.active.value === '1',
                rules: rules
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
         * 删除模板
         */
        async function deleteTemplate(id) {
            if (!confirm('确定删除该模板？')) {
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

        /**
         * 复制模板
         */
        async function copyTemplate(id) {
            try {
                const result = await apiRequest(config.singleUrl + id);
                
                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                const data = result.data;
                data.template_name = data.template_name + ' (副本)';
                delete data.id;
                
                fillForm(data);
                scrollToElement(elements.form);
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 导出 JSON
         */
        async function exportJSON() {
            try {
                const result = await apiRequest(config.listUrl);
                
                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '导出失败');
                    return;
                }

                const blob = new Blob([JSON.stringify(result.data.items, null, 2)], {
                    type: 'application/json'
                });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = 'shipping-templates.json';
                a.click();
                
                showNotice(elements.notice, 'success', '已导出');
            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        // 事件监听
        elements.tableBody.addEventListener('click', function(e) {
            const target = e.target;
            
            if (target.classList.contains('edit-tpl')) {
                fetchSingle(target.dataset.id);
            }
            
            if (target.classList.contains('del-tpl')) {
                deleteTemplate(target.dataset.id);
            }
            
            if (target.classList.contains('copy-tpl')) {
                copyTemplate(target.dataset.id);
            }
        });

        elements.rulesList.addEventListener('click', function(e) {
            const target = e.target;
            
            if (target.classList.contains('del-rule')) {
                deleteRule(parseInt(target.dataset.idx));
            }
            
            if (target.classList.contains('edit-rule')) {
                const idx = parseInt(target.dataset.idx);
                const rule = rules[idx];
                const fee = prompt('运费（元）', rule.fee);
                if (fee !== null) {
                    rule.fee = parseFloat(fee) || 0;
                    renderRules();
                }
            }
        });

        elements.createBtn.addEventListener('click', function() {
            resetForm();
            scrollToElement(elements.form);
        });

        elements.addRuleBtn.addEventListener('click', promptAddRule);
        elements.exportBtn.addEventListener('click', exportJSON);

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
