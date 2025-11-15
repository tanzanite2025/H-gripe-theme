/**
 * Tanzanite Settings - Tracking Providers Settings
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        const select = document.querySelector("select[name='provider']");
        const groups = document.querySelectorAll(".tz-provider-fields");

        if (!select || !groups.length) {
            return;
        }

        /**
         * 切换显示对应服务商的字段组
         */
        function toggleProviderFields() {
            const selectedProvider = select.value;
            
            groups.forEach(function(el) {
                if (el.dataset.provider === selectedProvider) {
                    el.style.display = 'block';
                } else {
                    el.style.display = 'none';
                }
            });
        }

        // 监听服务商选择变化
        select.addEventListener('change', toggleProviderFields);
        
        // 初始化显示
        toggleProviderFields();
    });
})();
