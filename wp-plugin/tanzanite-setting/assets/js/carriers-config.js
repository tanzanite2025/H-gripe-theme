/**
 * Tanzanite Settings - 物流追踪 API 配置
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzCarriersConfigPage === 'undefined') {
            console.error('Carriers Config Page config not found');
            return;
        }

        const config = TzCarriersConfigPage;
        const { showNotice } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            notice: document.getElementById('tz-carriers-config-notice'),
            form: document.getElementById('tz-tracking-config-form'),
            providerSelect: document.getElementById('tz-tracking-provider'),
            fieldsContainer: document.getElementById('tz-tracking-fields'),
            testBtn: document.getElementById('tz-tracking-test'),
            saveBtn: document.getElementById('tz-tracking-save')
        };

        // 检查必需元素
        if (!elements.form) {
            console.error('Required elements not found');
            return;
        }

        /**
         * 渲染配置字段
         */
        function renderFields(provider) {
            if (!elements.fieldsContainer) return;

            const providerConfig = config.providers[provider];
            if (!providerConfig) return;

            elements.fieldsContainer.innerHTML = '';

            Object.keys(providerConfig.fields).forEach(fieldKey => {
                const field = providerConfig.fields[fieldKey];
                const value = config.currentSettings[fieldKey] || field.default || '';

                const label = document.createElement('label');
                label.style.cssText = 'display:block;margin-bottom:12px;';
                
                const labelText = document.createElement('strong');
                labelText.textContent = field.label;
                label.appendChild(labelText);

                const input = document.createElement('input');
                input.type = field.type || 'text';
                input.name = `settings[${fieldKey}]`;
                input.value = value;
                input.className = 'regular-text';
                input.style.cssText = 'display:block;margin-top:4px;';
                
                label.appendChild(input);
                elements.fieldsContainer.appendChild(label);
            });
        }

        /**
         * 测试连接
         */
        async function testConnection() {
            if (!elements.testBtn) return;

            elements.testBtn.disabled = true;
            elements.testBtn.textContent = '测试中...';

            const formData = new FormData(elements.form);
            
            try {
                const response = await fetch(config.testUrl, {
                    method: 'POST',
                    body: formData
                });

                const data = await response.json();

                if (data.success) {
                    showNotice(elements.notice, 'success', '连接测试成功！');
                } else {
                    showNotice(elements.notice, 'error', data.data || '连接测试失败');
                }

            } catch (error) {
                showNotice(elements.notice, 'error', '测试失败：' + error.message);
            } finally {
                elements.testBtn.disabled = false;
                elements.testBtn.textContent = '测试连接';
            }
        }

        // 事件监听
        if (elements.providerSelect) {
            elements.providerSelect.addEventListener('change', function() {
                renderFields(this.value);
            });

            // 初始化渲染
            renderFields(elements.providerSelect.value);
        }

        if (elements.testBtn) {
            elements.testBtn.addEventListener('click', function(e) {
                e.preventDefault();
                testConnection();
            });
        }
    });
})();
