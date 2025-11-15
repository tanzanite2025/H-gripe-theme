/**
 * Tanzanite Settings - 公共 JavaScript 函数
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    // 全局命名空间
    window.TanzaniteAdmin = window.TanzaniteAdmin || {};

    /**
     * 显示通知消息
     * @param {HTMLElement} noticeElement - 通知元素
     * @param {string} type - 类型 (success/error/warning/info)
     * @param {string} message - 消息内容
     */
    window.TanzaniteAdmin.showNotice = function(noticeElement, type, message) {
        if (!noticeElement) return;
        
        if (!message) {
            noticeElement.style.display = 'none';
            return;
        }
        
        noticeElement.className = 'notice notice-' + type;
        noticeElement.innerHTML = '<p>' + message + '</p>';
        noticeElement.style.display = 'block';
        
        // 自动隐藏成功消息
        if (type === 'success') {
            setTimeout(function() {
                noticeElement.style.display = 'none';
            }, 3000);
        }
    };

    /**
     * 获取输入框值
     * @param {string} id - 元素 ID
     * @returns {string}
     */
    window.TanzaniteAdmin.getInputValue = function(id) {
        const el = document.getElementById(id);
        return el ? el.value : '';
    };

    /**
     * 设置输入框值
     * @param {string} id - 元素 ID
     * @param {string} value - 值
     */
    window.TanzaniteAdmin.setInputValue = function(id, value) {
        const el = document.getElementById(id);
        if (el) el.value = value || '';
    };

    /**
     * 获取复选框状态
     * @param {string} id - 元素 ID
     * @returns {boolean}
     */
    window.TanzaniteAdmin.getCheckboxValue = function(id) {
        const el = document.getElementById(id);
        return el ? el.checked : false;
    };

    /**
     * 设置复选框状态
     * @param {string} id - 元素 ID
     * @param {boolean} checked - 是否选中
     */
    window.TanzaniteAdmin.setCheckboxValue = function(id, checked) {
        const el = document.getElementById(id);
        if (el) el.checked = !!checked;
    };

    /**
     * 发起 REST API 请求
     * @param {string} url - API URL
     * @param {object} options - 请求选项
     * @returns {Promise}
     */
    window.TanzaniteAdmin.apiRequest = async function(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                'X-WP-Nonce': window.TanzaniteAdmin.nonce || ''
            }
        };

        const mergedOptions = Object.assign({}, defaultOptions, options);
        
        if (options.headers) {
            mergedOptions.headers = Object.assign({}, defaultOptions.headers, options.headers);
        }

        try {
            const response = await fetch(url, mergedOptions);
            
            // 检查响应类型
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/json')) {
                const text = await response.text();
                console.error('API returned non-JSON response:', text.substring(0, 500));
                return {
                    ok: false,
                    status: response.status,
                    data: { 
                        message: '服务器返回了非 JSON 响应。可能是权限问题或 API 未正确注册。请检查浏览器控制台查看详细信息。' 
                    }
                };
            }
            
            const data = await response.json();
            
            return {
                ok: response.ok,
                status: response.status,
                data: data
            };
        } catch (error) {
            console.error('API request error:', error);
            return {
                ok: false,
                status: 0,
                data: { message: '请求失败: ' + error.message }
            };
        }
    };

    /**
     * 确认对话框
     * @param {string} message - 确认消息
     * @returns {boolean}
     */
    window.TanzaniteAdmin.confirm = function(message) {
        return confirm(message || '确定要执行此操作吗？');
    };

    /**
     * 滚动到元素
     * @param {HTMLElement} element - 目标元素
     */
    window.TanzaniteAdmin.scrollToElement = function(element) {
        if (!element) return;
        
        window.scrollTo({
            top: element.offsetTop - 80,
            behavior: 'smooth'
        });
    };

    /**
     * 格式化日期
     * @param {string} dateString - 日期字符串
     * @returns {string}
     */
    window.TanzaniteAdmin.formatDate = function(dateString) {
        if (!dateString) return '';
        
        const date = new Date(dateString);
        return date.toLocaleString('zh-CN');
    };

    /**
     * 转义 HTML
     * @param {string} text - 文本
     * @returns {string}
     */
    window.TanzaniteAdmin.escapeHtml = function(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    };

    /**
     * 防抖函数
     * @param {Function} func - 函数
     * @param {number} wait - 等待时间（毫秒）
     * @returns {Function}
     */
    window.TanzaniteAdmin.debounce = function(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    };

    /**
     * 节流函数
     * @param {Function} func - 函数
     * @param {number} limit - 限制时间（毫秒）
     * @returns {Function}
     */
    window.TanzaniteAdmin.throttle = function(func, limit) {
        let inThrottle;
        return function(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    };

    console.log('Tanzanite Admin Common JS loaded');
})();
