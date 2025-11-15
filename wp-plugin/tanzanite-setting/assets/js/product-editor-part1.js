/**
 * Tanzanite Settings - 商品编辑页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 * 
 * 注意：这是一个大型文件，包含商品编辑的所有功能
 */

(function() {
    ''use strict'';

    document.addEventListener(''DOMContentLoaded'', function() {
        // 检查配置是否存在
        if (typeof TzProductEditorConfig === ''undefined'') {
            console.error(''Product Editor config not found'');
            return;
        }

        const config = TzProductEditorConfig;
        const strings = config.strings || {};
        const { showNotice, apiRequest, scrollToElement } = window.TanzaniteAdmin;

        // 全局变量
        let skuRows = config.initialSkuRows || [];
        let currentProduct = null;

        // DOM 元素引用
        const elements = {
            notice: document.getElementById(''tz-product-notice''),
            form: document.getElementById(''tz-product-editor-form''),
            saveBtn: document.getElementById(''tz-product-save''),
            resetBtn: document.getElementById(''tz-product-reset''),
            skuTable: document.querySelector(''#tz-product-sku-table tbody''),
            skuForm: document.getElementById(''tz-product-sku-form''),
            skuEmpty: document.getElementById(''tz-product-sku-empty''),
            featured: {
                preview: document.getElementById(''tz-product-featured-preview''),
                idInput: document.getElementById(''tz-product-featured-id''),
                urlInput: document.getElementById(''tz-product-featured-url''),
                selectBtn: document.getElementById(''tz-product-featured-select''),
                clearBtn: document.getElementById(''tz-product-featured-clear'')
            },
            gallery: {
                preview: document.getElementById(''tz-product-gallery-preview''),
                idsInput: document.getElementById(''tz-product-gallery-ids''),
                externalsTextarea: document.getElementById(''tz-product-gallery-externals''),
                selectBtn: document.getElementById(''tz-product-gallery-select''),
                clearBtn: document.getElementById(''tz-product-gallery-clear'')
            },
            video: {
                preview: document.getElementById(''tz-product-video-preview''),
                idInput: document.getElementById(''tz-product-video-id''),
                urlInput: document.getElementById(''tz-product-video-url''),
                selectBtn: document.getElementById(''tz-product-video-select''),
                clearBtn: document.getElementById(''tz-product-video-clear'')
            },
            taxRatesList: document.getElementById(''tz-product-tax-rates-list'')
        };

        let mediaFrameFeatured = null;
        let mediaFrameGallery = null;
        let mediaFrameVideo = null;
