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
            ruleEditor: document.getElementById('tz-shipping-rule-editor'),
            ruleInputs: {
                type: document.getElementById('tz-shipping-rule-type'),
                service: document.getElementById('tz-shipping-rule-service'),
                serviceLabel: document.getElementById('tz-shipping-rule-service-label'),
                regions: document.getElementById('tz-shipping-rule-regions'),
                zipRanges: document.getElementById('tz-shipping-rule-zip-ranges'),
                min: document.getElementById('tz-shipping-rule-min'),
                max: document.getElementById('tz-shipping-rule-max'),
                fee: document.getElementById('tz-shipping-rule-fee'),
                freeOver: document.getElementById('tz-shipping-rule-free-over'),
                etaMin: document.getElementById('tz-shipping-rule-eta-min'),
                etaMax: document.getElementById('tz-shipping-rule-eta-max')
            },
            ruleButtons: {
                save: document.getElementById('tz-shipping-rule-save'),
                reset: document.getElementById('tz-shipping-rule-reset')
            },
            inputs: {
                id: document.getElementById('tz-shipping-id'),
                name: document.getElementById('tz-shipping-name'),
                description: document.getElementById('tz-shipping-description'),
                active: document.getElementById('tz-shipping-active'),
                carrier: document.getElementById('tz-shipping-carrier')
            }
        };

        // 检查必需元素
        if (!elements.tableBody || !elements.form) {
            console.error('Required elements not found');
            return;
        }

        let rules = [];
        let editingRuleIndex = null;

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
            if (elements.inputs.carrier) {
                elements.inputs.carrier.value = '';
            }
            rules = [];
            editingRuleIndex = null;
            resetRuleEditor();
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
            if (elements.inputs.carrier) {
                const meta = data.meta || {};
                elements.inputs.carrier.value = meta.carrier || '';
            }
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
                const serviceText = rule.service_label || rule.service || '默认方式';
                const regionsText = Array.isArray(rule.regions) && rule.regions.length
                    ? rule.regions.join(', ')
                    : '未设置国家';
                const zipRangesText = Array.isArray(rule.zip_ranges) && rule.zip_ranges.length
                    ? ' [邮编: ' + rule.zip_ranges.join(', ') + ']'
                    : '';
                const etaText = (rule.eta_min_days != null || rule.eta_max_days != null)
                    ? ' · 时效: ' + (rule.eta_min_days != null ? rule.eta_min_days : '?') + '-' + (rule.eta_max_days != null ? rule.eta_max_days : '?') + ' 天'
                    : '';
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
                            <strong>${typeText}</strong> · ${serviceText} · ${regionsText}${zipRangesText} ${rangeText} → 运费: ¥${rule.fee}${freeOverText}${etaText}
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
         * 重置规则编辑表单
         */
        function resetRuleEditor() {
            const ri = elements.ruleInputs || {};

            if (ri.type && ri.type.options && ri.type.options.length) {
                ri.type.value = ri.type.value || 'weight';
            }
            if (ri.service) {
                ri.service.value = '';
            }
            if (ri.serviceLabel) {
                ri.serviceLabel.value = '';
            }
            if (ri.regions) {
                ri.regions.value = '';
            }
            if (ri.zipRanges) {
                ri.zipRanges.value = '';
            }
            if (ri.min) {
                ri.min.value = '';
            }
            if (ri.max) {
                ri.max.value = '';
            }
            if (ri.fee) {
                ri.fee.value = '';
            }
            if (ri.freeOver) {
                ri.freeOver.value = '';
            }
            if (ri.etaMin) {
                ri.etaMin.value = '';
            }
            if (ri.etaMax) {
                ri.etaMax.value = '';
            }

            editingRuleIndex = null;
        }

        /**
         * 将已有规则填充到编辑表单
         */
        function fillRuleEditor(rule, idx) {
            const ri = elements.ruleInputs || {};
            editingRuleIndex = typeof idx === 'number' ? idx : null;

            if (!rule) {
                resetRuleEditor();
                return;
            }

            if (ri.type && rule.type) {
                ri.type.value = rule.type;
            }
            if (ri.service) {
                ri.service.value = rule.service || '';
            }
            if (ri.serviceLabel) {
                ri.serviceLabel.value = rule.service_label || '';
            }
            if (ri.regions) {
                ri.regions.value = Array.isArray(rule.regions) && rule.regions.length
                    ? rule.regions.join(',')
                    : '';
            }
            if (ri.zipRanges) {
                ri.zipRanges.value = Array.isArray(rule.zip_ranges) && rule.zip_ranges.length
                    ? rule.zip_ranges.join(',')
                    : '';
            }
            if (ri.min) {
                ri.min.value = rule.min != null ? rule.min : '';
            }
            if (ri.max) {
                ri.max.value = rule.max != null ? rule.max : '';
            }
            if (ri.fee) {
                ri.fee.value = rule.fee != null ? rule.fee : '';
            }
            if (ri.freeOver) {
                ri.freeOver.value = rule.free_over != null ? rule.free_over : '';
            }
            if (ri.etaMin) {
                ri.etaMin.value = rule.eta_min_days != null ? rule.eta_min_days : '';
            }
            if (ri.etaMax) {
                ri.etaMax.value = rule.eta_max_days != null ? rule.eta_max_days : '';
            }
        }

        /**
         * 从规则编辑表单构建规则对象
         */
        function buildRuleFromEditor() {
            const ri = elements.ruleInputs || {};

            const type = ri.type && ri.type.value ? ri.type.value : 'weight';
            const service = ri.service ? ri.service.value.trim() : '';
            const serviceLabel = ri.serviceLabel ? ri.serviceLabel.value.trim() : '';

            const regionsInput = ri.regions ? ri.regions.value.trim() : '';
            let regions = [];
            if (regionsInput) {
                regions = regionsInput.split(',').map(function(s) {
                    return s.trim().toUpperCase();
                }).filter(function(s) {
                    return s.length > 0;
                });
            }

            const zipRangesInput = ri.zipRanges ? ri.zipRanges.value.trim() : '';
            let zipRanges = [];
            if (zipRangesInput) {
                zipRanges = zipRangesInput.split(',').map(function(s) {
                    return s.trim();
                }).filter(function(s) {
                    return s.length > 0;
                });
            }

            const minStr = ri.min ? ri.min.value : '';
            const maxStr = ri.max ? ri.max.value : '';
            const feeStr = ri.fee ? ri.fee.value : '';
            const freeOverStr = ri.freeOver ? ri.freeOver.value : '';
            const etaMinStr = ri.etaMin ? ri.etaMin.value : '';
            const etaMaxStr = ri.etaMax ? ri.etaMax.value : '';

            const fee = feeStr !== '' ? parseFloat(feeStr) : NaN;
            if (isNaN(fee)) {
                alert('请填写有效的运费金额');
                return null;
            }

            const min = minStr !== '' ? parseFloat(minStr) : null;
            const max = maxStr !== '' ? parseFloat(maxStr) : null;
            const freeOver = freeOverStr !== '' ? parseFloat(freeOverStr) : null;
            const etaMin = etaMinStr !== '' ? parseInt(etaMinStr, 10) : null;
            const etaMax = etaMaxStr !== '' ? parseInt(etaMaxStr, 10) : null;

            return {
                type: type,
                min: min,
                max: max,
                fee: fee,
                priority: 0,
                free_over: freeOver,
                service: service || undefined,
                service_label: serviceLabel || undefined,
                regions: regions,
                zip_ranges: zipRanges,
                eta_min_days: etaMin,
                eta_max_days: etaMax
            };
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
            // 现在使用下方的小表单来添加规则
            editingRuleIndex = null;
            resetRuleEditor();
            if (elements.ruleEditor) {
                scrollToElement(elements.ruleEditor);
            }
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

            // 	00ecfdbf	7ed7ccdffcfacfdbf	fe	0e	c3	d	f
            if (elements.inputs.carrier) {
                const carrier = elements.inputs.carrier.value.trim();
                if (carrier) {
                    payload.meta = {
                        carrier: carrier
                    };
                }
            }

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
                if (!rule) {
                    return;
                }
                fillRuleEditor(rule, idx);
                if (elements.ruleEditor) {
                    scrollToElement(elements.ruleEditor);
                }
            }
        });

        elements.createBtn.addEventListener('click', function() {
            resetForm();
            scrollToElement(elements.form);
        });

        elements.addRuleBtn.addEventListener('click', promptAddRule);
        elements.exportBtn.addEventListener('click', exportJSON);

        if (elements.ruleButtons.save) {
            elements.ruleButtons.save.addEventListener('click', function() {
                const rule = buildRuleFromEditor();
                if (!rule) {
                    return;
                }
                addOrUpdateRule(rule, editingRuleIndex);
                editingRuleIndex = null;
                resetRuleEditor();
            });
        }

        if (elements.ruleButtons.reset) {
            elements.ruleButtons.reset.addEventListener('click', function() {
                editingRuleIndex = null;
                resetRuleEditor();
            });
        }

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
