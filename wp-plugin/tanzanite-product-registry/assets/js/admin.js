/**
 * Tanzanite Product Registry - Admin Scripts
 */

(function($) {
	'use strict';

	// 通用工具函数
	window.tanzanitePRUtils = {
		// 转义 HTML
		escapeHtml: function(text) {
			if (!text) return '';
			const div = document.createElement('div');
			div.textContent = text;
			return div.innerHTML;
		},

		// 格式化日期
		formatDate: function(dateStr) {
			if (!dateStr) return '-';
			return dateStr.substring(0, 10);
		},

		// 格式化年月
		formatYearMonth: function(dateStr) {
			if (!dateStr) return '-';
			return dateStr.substring(0, 7);
		}
	};

})(jQuery);
