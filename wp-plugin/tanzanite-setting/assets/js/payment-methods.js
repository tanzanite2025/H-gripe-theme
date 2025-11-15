/**
 * Tanzanite Settings - 支付方式管理页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.8
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        console.log('Payment Methods JS: DOMContentLoaded');
        
        // 检查配置是否存在
        if (typeof TzPaymentMethodsConfig === 'undefined') {
            console.error('Payment Methods config not found');
            return;
        }

        const config = TzPaymentMethodsConfig;
        console.log('Payment Methods config loaded');
        
        // 检查 TanzaniteAdmin 是否存在
        let showNotice, apiRequest, scrollToElement, confirm;
        
        if (typeof window.TanzaniteAdmin === 'undefined') {
            console.warn('TanzaniteAdmin not found! Some features may not work.');
            
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
            console.log('TanzaniteAdmin found');
            window.TanzaniteAdmin.nonce = config.nonce;
            ({ showNotice, apiRequest, scrollToElement, confirm } = window.TanzaniteAdmin);
        }

        // DOM 元素
        const elements = {
            tableBody: document.querySelector('#tz-payment-table tbody'),
            notice: document.getElementById('tz-payment-notice'),
            form: document.getElementById('tz-payment-form'),
            createBtn: document.getElementById('tz-payment-create'),
            saveBtn: document.getElementById('tz-payment-save'),
            resetBtn: document.getElementById('tz-payment-reset'),
            iconPreview: document.getElementById('tz-payment-icon-preview'),
            iconUploadBtn: document.getElementById('tz-payment-icon-upload'),
            inputs: {
                id: document.getElementById('tz-payment-id'),
                code: document.getElementById('tz-payment-code'),
                name: document.getElementById('tz-payment-name'),
                description: document.getElementById('tz-payment-description'),
                iconUrl: document.getElementById('tz-payment-icon-url'),
                feeType: document.getElementById('tz-payment-fee-type'),
                feeValue: document.getElementById('tz-payment-fee-value'),
                terminalsContainer: document.getElementById('tz-payment-terminals'),
                levels: document.getElementById('tz-payment-levels'),
                currencies: document.getElementById('tz-payment-currencies'),
                defaultCurrency: document.getElementById('tz-payment-default-currency'),
                enabled: document.getElementById('tz-payment-enabled'),
                sort: document.getElementById('tz-payment-sort')
            }
        };

        // 检查必需元素
        if (!elements.tableBody || !elements.form) {
            console.error('Required elements not found');
            return;
        }

        // 终端选项
        const terminalOptions = [
            { value: 'web', label: '网页端' },
            { value: 'mobile', label: '移动端' },
            { value: 'app', label: 'APP' },
            { value: 'wechat', label: '微信小程序' }
        ];

        // WordPress 媒体选择器
        let mediaUploader;

        /**
         * 初始化终端复选框
         */
        function initTerminalsCheckboxes() {
            if (!elements.inputs.terminalsContainer) return;

            elements.inputs.terminalsContainer.innerHTML = '';
            terminalOptions.forEach(function(option) {
                const label = document.createElement('label');
                label.style.display = 'flex';
                label.style.alignItems = 'center';
                label.style.gap = '8px';
                label.style.marginBottom = '8px';

                const checkbox = document.createElement('input');
                checkbox.type = 'checkbox';
                checkbox.value = option.value;
                checkbox.className = 'terminal-checkbox';

                const span = document.createElement('span');
                span.textContent = option.label;

                label.appendChild(checkbox);
                label.appendChild(span);
                elements.inputs.terminalsContainer.appendChild(label);
            });
        }

        /**
         * 获取选中的终端
         */
        function getSelectedTerminals() {
            const checkboxes = elements.inputs.terminalsContainer.querySelectorAll('.terminal-checkbox:checked');
            return Array.from(checkboxes).map(cb => cb.value);
        }

        /**
         * 设置选中的终端
         */
        function setSelectedTerminals(terminals) {
            const checkboxes = elements.inputs.terminalsContainer.querySelectorAll('.terminal-checkbox');
            checkboxes.forEach(function(cb) {
                cb.checked = terminals.includes(cb.value);
            });
        }

        /**
         * 更新图标预览
         */
        function updateIconPreview() {
            const url = elements.inputs.iconUrl.value;
            if (url) {
                elements.iconPreview.innerHTML = '<img src="' + url + '" style="max-width:120px;max-height:60px;" />';
            } else {
                elements.iconPreview.innerHTML = '<span style="color:#9ca3af;">无图标</span>';
            }
        }

        /**
         * 打开媒体选择器
         */
        function openMediaUploader(e) {
            if (e) e.preventDefault();
            
            // 检查 wp.media 是否存在
            if (typeof wp === 'undefined' || typeof wp.media === 'undefined') {
                alert('媒体库未加载，请刷新页面重试');
                console.error('wp.media is not available');
                return;
            }
            
            if (mediaUploader) {
                mediaUploader.open();
                return;
            }

            try {
                mediaUploader = wp.media({
                    title: '选择支付图标',
                    button: { text: '使用此图片' },
                    multiple: false
                });

                mediaUploader.on('select', function() {
                    const attachment = mediaUploader.state().get('selection').first().toJSON();
                    elements.inputs.iconUrl.value = attachment.url;
                    updateIconPreview();
                });

                mediaUploader.open();
            } catch (error) {
                console.error('Failed to open media uploader:', error);
                alert('打开媒体库失败: ' + error.message);
            }
        }

        /**
         * 重置表单
         */
        function resetForm() {
            elements.form.reset();
            elements.inputs.id.value = '';
            elements.inputs.feeType.value = 'fixed';
            elements.inputs.enabled.value = '1';
            elements.inputs.sort.value = '0';
            setSelectedTerminals([]);
            
            // 重置支付平台配置
            loadGatewayConfig(null);
            
            updateIconPreview();
            showNotice(elements.notice, null);
        }

        /**
         * 填充表单
         */
        function fillForm(data) {
            resetForm();
            elements.inputs.id.value = data.id || '';
            elements.inputs.code.value = data.code || '';
            elements.inputs.name.value = data.name || '';
            elements.inputs.description.value = data.description || '';
            elements.inputs.iconUrl.value = data.icon_url || '';
            elements.inputs.feeType.value = data.fee_type || 'fixed';
            elements.inputs.feeValue.value = data.fee_value || 0;
            setSelectedTerminals(data.terminals || []);
            elements.inputs.levels.value = (data.membership_levels || []).join(', ');
            elements.inputs.currencies.value = (data.currencies || []).join(', ');
            elements.inputs.defaultCurrency.value = data.default_currency || '';
            elements.inputs.enabled.value = data.is_enabled ? '1' : '0';
            elements.inputs.sort.value = data.sort_order || 0;
            
            // 加载支付平台配置
            loadGatewayConfig(data.gateway_config);
            
            updateIconPreview();
        }

        /**
         * 渲染表格行
         */
        function renderRows(items) {
            elements.tableBody.innerHTML = '';

            if (!items || items.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="6" style="text-align:center;">暂无支付方式</td>';
                elements.tableBody.appendChild(tr);
                return;
            }

            items.forEach(function(item) {
                const tr = document.createElement('tr');
                
                const iconHtml = item.icon_url 
                    ? '<img src="' + item.icon_url + '" style="max-width:60px;max-height:30px;" />'
                    : '-';
                
                const statusText = item.is_active 
                    ? '<span style="color:#16a34a;">启用</span>' 
                    : '<span style="color:#9ca3af;">禁用</span>';
                
                const feeText = item.fee_type === 'fixed' 
                    ? '¥' + item.fee_value 
                    : item.fee_value + '%';
                
                tr.innerHTML = `
                    <td>${iconHtml}</td>
                    <td><strong>${item.name}</strong><br><small>${item.code}</small></td>
                    <td>${feeText}</td>
                    <td>${(item.currencies || []).join(', ') || '-'}</td>
                    <td>${statusText}</td>
                    <td>
                        <button class="button-link tz-payment-edit" data-id="${item.id}">编辑</button> | 
                        <button class="button-link-delete tz-payment-delete" data-id="${item.id}">删除</button>
                    </td>
                `;
                
                elements.tableBody.appendChild(tr);
            });
        }

        /**
         * 获取列表
         */
        async function fetchList() {
            showNotice(elements.notice, null);

            try {
                const result = await apiRequest(config.listUrl);
                
                if (!result.ok) {
                    renderRows([]);
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                renderRows(result.data.items || []);
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
            
            // 处理货币列表
            const currencies = elements.inputs.currencies.value
                .split(',')
                .map(c => c.trim().toUpperCase())
                .filter(c => c);

            // 收集支付平台配置
            const gatewayConfig = collectGatewayConfig();

            const payload = {
                code: elements.inputs.code.value,
                name: elements.inputs.name.value,
                description: elements.inputs.description.value,
                icon_url: elements.inputs.iconUrl.value,
                fee_type: elements.inputs.feeType.value,
                fee_value: parseFloat(elements.inputs.feeValue.value) || 0,
                terminals: getSelectedTerminals(),
                membership_levels: elements.inputs.levels.value.split(',').map(m => m.trim()).filter(m => m),
                currencies: currencies,
                default_currency: elements.inputs.defaultCurrency.value.toUpperCase(),
                is_enabled: elements.inputs.enabled.value === '1',
                sort_order: parseInt(elements.inputs.sort.value) || 0,
                gateway_config: gatewayConfig
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
         * 删除
         */
        async function deleteItem(id) {
            if (!confirm('确定要删除该支付方式吗？')) {
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
         * 收集支付平台配置数据
         */
        function collectGatewayConfig() {
            const gatewayTypeSelect = document.getElementById('tz-payment-gateway-type');
            if (!gatewayTypeSelect) return null;
            
            const gatewayType = gatewayTypeSelect.value;
            if (!gatewayType || !config.gatewayFields || !config.gatewayFields[gatewayType]) {
                return null;
            }

            const gatewayConfig = { type: gatewayType };
            const fields = config.gatewayFields[gatewayType];
            
            fields.forEach(function(field) {
                const inputId = 'tz-' + gatewayType + '-' + field.replace(/_/g, '-');
                const input = document.getElementById(inputId);
                if (input) {
                    gatewayConfig[field] = input.value;
                }
            });

            return gatewayConfig;
        }

        /**
         * 加载支付平台配置数据
         */
        function loadGatewayConfig(gatewayConfig) {
            const gatewayTypeSelect = document.getElementById('tz-payment-gateway-type');
            if (!gatewayTypeSelect) return;
            
            // 重置平台类型选择
            gatewayTypeSelect.value = '';

            if (!gatewayConfig || !gatewayConfig.type) {
                return;
            }

            const gatewayType = gatewayConfig.type;
            gatewayTypeSelect.value = gatewayType;
            
            // 触发 change 事件以显示配置区域（由内联脚本处理）
            const event = new Event('change');
            gatewayTypeSelect.dispatchEvent(event);

            // 填充字段值
            if (config.gatewayFields && config.gatewayFields[gatewayType]) {
                // 延迟一下确保 DOM 已更新
                setTimeout(function() {
                    config.gatewayFields[gatewayType].forEach(function(field) {
                        const inputId = 'tz-' + gatewayType + '-' + field.replace(/_/g, '-');
                        const input = document.getElementById(inputId);
                        if (input && gatewayConfig[field] !== undefined) {
                            input.value = gatewayConfig[field];
                        }
                    });
                }, 50);
            }
        }

        // 事件监听
        elements.tableBody.addEventListener('click', function(e) {
            const target = e.target;
            
            if (target.classList.contains('tz-payment-edit')) {
                e.preventDefault();
                fetchSingle(target.dataset.id);
            }
            
            if (target.classList.contains('tz-payment-delete')) {
                e.preventDefault();
                deleteItem(target.dataset.id);
            }
        });

        elements.createBtn.addEventListener('click', function() {
            resetForm();
            scrollToElement(elements.form);
        });

        if (elements.iconUploadBtn) {
            console.log('Icon upload button found, binding event');
            elements.iconUploadBtn.addEventListener('click', openMediaUploader);
        } else {
            console.warn('Icon upload button not found');
        }
        
        if (elements.inputs.iconUrl) {
            elements.inputs.iconUrl.addEventListener('input', updateIconPreview);
        }

        elements.form.addEventListener('submit', saveForm);
        elements.saveBtn.addEventListener('click', saveForm);
        elements.resetBtn.addEventListener('click', function(e) {
            e.preventDefault();
            resetForm();
        });

        // 初始化
        initTerminalsCheckboxes();
        updateIconPreview();
        fetchList();
    });
})();
