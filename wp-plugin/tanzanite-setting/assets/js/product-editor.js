/**
 * Tanzanite Settings - 商品编辑页面
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 * 
 * 注意：这是一个大型文件，包含商品编辑的所有功能
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzProductEditorConfig === 'undefined') {
            console.error('Product Editor config not found');
            return;
        }

        // 检查 TanzaniteAdmin 是否存在
        if (typeof window.TanzaniteAdmin === 'undefined') {
            console.error('TanzaniteAdmin not found. Make sure admin-common.js is loaded.');
            return;
        }

        const config = TzProductEditorConfig;
        const strings = config.strings || {};
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest, scrollToElement } = window.TanzaniteAdmin;

        // 全局变量
        let skuRows = config.initialSkuRows || [];
        let tierRows = [];
        let currentProduct = null;

        // DOM 元素引用
        const elements = {
            notice: document.getElementById('tz-product-notice'),
            form: document.getElementById('tz-product-editor-form'),
            saveBtn: document.getElementById('tz-product-save'),
            resetBtn: document.getElementById('tz-product-reset'),
            skuTable: document.querySelector('#tz-product-sku-table tbody'),
            skuForm: document.getElementById('tz-product-sku-form'),
            skuEmpty: document.getElementById('tz-product-sku-empty'),
            featured: {
                preview: document.getElementById('tz-product-featured-preview'),
                idInput: document.getElementById('tz-product-featured-id'),
                urlInput: document.getElementById('tz-product-featured-url'),
                selectBtn: document.getElementById('tz-product-featured-select'),
                clearBtn: document.getElementById('tz-product-featured-clear')
            },
            gallery: {
                preview: document.getElementById('tz-product-gallery-preview'),
                idsInput: document.getElementById('tz-product-gallery-ids'),
                externalsTextarea: document.getElementById('tz-product-gallery-externals'),
                selectBtn: document.getElementById('tz-product-gallery-select'),
                clearBtn: document.getElementById('tz-product-gallery-clear')
            },
            video: {
                preview: document.getElementById('tz-product-video-preview'),
                idInput: document.getElementById('tz-product-video-id'),
                urlInput: document.getElementById('tz-product-video-url'),
                selectBtn: document.getElementById('tz-product-video-select'),
                clearBtn: document.getElementById('tz-product-video-clear')
            },
            taxRatesList: document.getElementById('tz-product-tax-rates-list')
        };

        let mediaFrameFeatured = null;
        let mediaFrameGallery = null;
        let mediaFrameVideo = null;

        /**
         * ========== 工具函数 ==========
         */

        function getInputValue(id) {
            const el = document.getElementById(id);
            return el ? el.value : '';
        }

        function fillInputValue(id, value) {
            const el = document.getElementById(id);
            if (el) el.value = value || '';
        }

        function getCheckboxValue(id) {
            const el = document.getElementById(id);
            return el ? el.checked : false;
        }

        function fillCheckboxValue(id, checked) {
            const el = document.getElementById(id);
            if (el) el.checked = !!checked;
        }

        function parseTextareaLines(textarea) {
            if (!textarea || !textarea.value) return [];
            return textarea.value
                .split(/\r?\n/)
                .map((line) => line.trim())
                .filter(Boolean);
        }

        function parseGalleryIds(idsInput) {
            if (!idsInput || !idsInput.value) return [];
            return idsInput.value
                .split(',')
                .map((item) => parseInt(item, 10))
                .filter((item) => Number.isFinite(item) && item > 0);
        }

        /**
         * ========== 媒体资源渲染 ==========
         */

        function renderFeaturedMedia() {
            const { preview, idInput, urlInput } = elements.featured;
            if (!preview || !idInput || !urlInput) return;

            preview.innerHTML = '';
            const mediaId = parseInt(idInput.value, 10) || 0;
            const mediaUrl = (urlInput.value || '').trim();

            if (!mediaId && !mediaUrl) return;

            const img = document.createElement('img');
            img.style.maxWidth = '180px';
            img.style.borderRadius = '6px';
            img.style.border = '1px solid #d1d5db';
            img.src = mediaUrl || config.mediaEndpoint.replace('/wp/v2/media', '/wp/v2/media/') + mediaId;
            preview.appendChild(img);
        }

        function renderGalleryMedia() {
            const { preview, idsInput, externalsTextarea } = elements.gallery;
            if (!preview) return;

            preview.innerHTML = '';
            const ids = parseGalleryIds(idsInput);
            const externals = parseTextareaLines(externalsTextarea);

            const fragment = document.createDocumentFragment();

            ids.forEach((id) => {
                const img = document.createElement('img');
                img.src = config.mediaEndpoint.replace('/wp/v2/media', '/wp/v2/media/') + id;
                img.style.width = '80px';
                img.style.height = '80px';
                img.style.objectFit = 'cover';
                img.style.borderRadius = '6px';
                img.style.border = '1px solid #d1d5db';
                fragment.appendChild(img);
            });

            externals.forEach((url) => {
                const img = document.createElement('img');
                img.src = url;
                img.style.width = '80px';
                img.style.height = '80px';
                img.style.objectFit = 'cover';
                img.style.borderRadius = '6px';
                img.style.border = '1px solid #d1d5db';
                fragment.appendChild(img);
            });

            preview.appendChild(fragment);
        }

        function renderVideoMedia() {
            const { preview, idInput, urlInput } = elements.video;
            if (!preview || !idInput || !urlInput) return;

            preview.innerHTML = '';
            const videoId = parseInt(idInput.value, 10) || 0;
            const videoUrl = (urlInput.value || '').trim();

            if (!videoId && !videoUrl) return;

            const container = document.createElement('div');
            container.style.display = 'flex';
            container.style.flexDirection = 'column';
            container.style.gap = '4px';

            if (videoUrl) {
                const link = document.createElement('a');
                link.href = videoUrl;
                link.textContent = videoUrl;
                link.target = '_blank';
                container.appendChild(link);
            } else {
                const text = document.createElement('code');
                text.textContent = '媒体库视频 ID：' + videoId;
                container.appendChild(text);
            }

            preview.appendChild(container);
        }

        function initFeaturedSelectors() {
            const { selectBtn, clearBtn, idInput, urlInput } = elements.featured;
            if (!selectBtn) return;

            selectBtn.addEventListener('click', (event) => {
                event.preventDefault();

                if (!window.wp || !wp.media) {
                    console.error('WordPress media library not available');
                    return;
                }

                if (mediaFrameFeatured) {
                    mediaFrameFeatured.open();
                    return;
                }

                mediaFrameFeatured = wp.media({
                    title: strings.mediaSelectTitle || '选择图片',
                    button: { text: strings.mediaSelectButton || '使用所选' },
                    multiple: false,
                    library: { type: ['image'] }
                });

                mediaFrameFeatured.on('select', () => {
                    const attachment = mediaFrameFeatured.state().get('selection').first();
                    if (!attachment) return;
                    const data = attachment.toJSON();
                    fillInputValue('tz-product-featured-id', data.id);
                    fillInputValue('tz-product-featured-url', data.url || '');
                    renderFeaturedMedia();
                });

                mediaFrameFeatured.open();
            });

            if (clearBtn) {
                clearBtn.addEventListener('click', (event) => {
                    event.preventDefault();
                    if (!window.confirm(strings.mediaClearConfirm || '确定要清除该媒体吗？')) return;
                    if (idInput) idInput.value = '';
                    if (urlInput) urlInput.value = '';
                    renderFeaturedMedia();
                });
            }

            if (urlInput) {
                urlInput.addEventListener('blur', renderFeaturedMedia);
            }
        }

        function initGallerySelectors() {
            const { selectBtn, clearBtn, idsInput, externalsTextarea } = elements.gallery;
            if (selectBtn) {
                selectBtn.addEventListener('click', (event) => {
                    event.preventDefault();

                    if (!window.wp || !wp.media) {
                        console.error('WordPress media library not available');
                        return;
                    }

                    if (mediaFrameGallery) {
                        mediaFrameGallery.open();
                        return;
                    }

                    mediaFrameGallery = wp.media({
                        title: strings.mediaSelectTitle || '选择图片',
                        button: { text: strings.mediaSelectButton || '使用所选' },
                        multiple: true,
                        library: { type: ['image'] }
                    });

                    mediaFrameGallery.on('select', () => {
                        const selection = mediaFrameGallery.state().get('selection');
                        const ids = selection.map((attachment) => attachment.id);
                        if (idsInput) {
                            const existing = parseGalleryIds(idsInput);
                            const combined = Array.from(new Set([...existing, ...ids]));
                            idsInput.value = combined.join(',');
                        }
                        renderGalleryMedia();
                    });

                    mediaFrameGallery.open();
                });
            }

            if (clearBtn) {
                clearBtn.addEventListener('click', (event) => {
                    event.preventDefault();
                    if (!window.confirm(strings.mediaClearConfirm || '确定要清除该媒体吗？')) return;
                    if (idsInput) idsInput.value = '';
                    renderGalleryMedia();
                });
            }

            if (externalsTextarea) {
                externalsTextarea.addEventListener('blur', renderGalleryMedia);
            }
        }

        function initVideoSelectors() {
            const { selectBtn, clearBtn, idInput, urlInput } = elements.video;
            if (selectBtn) {
                selectBtn.addEventListener('click', (event) => {
                    event.preventDefault();

                    if (!window.wp || !wp.media) {
                        console.error('WordPress media library not available');
                        return;
                    }

                    if (mediaFrameVideo) {
                        mediaFrameVideo.open();
                        return;
                    }

                    mediaFrameVideo = wp.media({
                        title: strings.mediaSelectTitle || '选择媒体',
                        button: { text: strings.mediaSelectButton || '使用所选' },
                        multiple: false
                    });

                    mediaFrameVideo.on('select', () => {
                        const attachment = mediaFrameVideo.state().get('selection').first();
                        if (!attachment) return;
                        const data = attachment.toJSON();
                        fillInputValue('tz-product-video-id', data.id);
                        fillInputValue('tz-product-video-url', data.url || '');
                        renderVideoMedia();
                    });

                    mediaFrameVideo.open();
                });
            }

            if (clearBtn) {
                clearBtn.addEventListener('click', (event) => {
                    event.preventDefault();
                    if (!window.confirm(strings.mediaClearConfirm || '确定要清除该媒体吗？')) return;
                    if (idInput) idInput.value = '';
                    if (urlInput) urlInput.value = '';
                    renderVideoMedia();
                });
            }

            if (urlInput) {
                urlInput.addEventListener('blur', renderVideoMedia);
            }
        }

        /**
         * ========== SKU 管理 ==========
         */

        function renderSkuTable() {
            if (!elements.skuTable) return;

            elements.skuTable.innerHTML = '';

            if (!skuRows || skuRows.length === 0) {
                if (elements.skuEmpty) {
                    elements.skuEmpty.style.display = 'block';
                }
                return;
            }

            if (elements.skuEmpty) {
                elements.skuEmpty.style.display = 'none';
            }

            skuRows.forEach(function(row, index) {
                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>${row.code || ''}</td>
                    <td>${row.attributes_input || ''}</td>
                    <td>${row.price_regular || ''}</td>
                    <td>${row.price_sale || ''}</td>
                    <td>${row.stock || ''}</td>
                    <td>${row.barcode || ''}</td>
                    <td>${row.weight || ''}</td>
                    <td>
                        <button type="button" class="button-link sku-edit" data-index="${index}">${strings.skuEdit || '编辑'}</button> | 
                        <button type="button" class="button-link-delete sku-delete" data-index="${index}">${strings.skuDelete || '删除'}</button>
                    </td>
                `;
                elements.skuTable.appendChild(tr);
            });
        }

        function addOrUpdateSku(skuData, editIndex) {
            if (editIndex !== null && editIndex >= 0) {
                skuRows[editIndex] = skuData;
            } else {
                // 检查重复
                const exists = skuRows.some(row => row.code === skuData.code);
                if (exists) {
                    showNotice(elements.notice, 'error', strings.skuDuplicate || 'SKU 编码已存在');
                    return false;
                }
                skuRows.push(skuData);
            }
            renderSkuTable();
            return true;
        }

        function deleteSku(index) {
            if (!confirm(strings.skuDeleteConfirm || '确定删除？')) {
                return;
            }
            skuRows.splice(index, 1);
            renderSkuTable();
            showNotice(elements.notice, 'success', strings.skuDeleted || 'SKU 已删除');
        }

        /**
         * ========== 税率加载 ==========
         */

        async function loadTaxRates() {
            if (!config.taxRatesEndpoint) return;

            const container = document.getElementById('tz-product-tax-rates-list');
            if (!container) return;

            try {
                const result = await apiRequest(config.taxRatesEndpoint);
                
                if (result.ok && result.data.items) {
                    container.innerHTML = '';
                    result.data.items.filter(r => r.is_active).forEach(function(rate) {
                        const label = document.createElement('label');
                        label.style.cssText = 'display:flex;align-items:center;gap:8px;padding:8px;background:#fff;border:1px solid #d1d5db;border-radius:6px;';
                        
                        const checkbox = document.createElement('input');
                        checkbox.type = 'checkbox';
                        checkbox.value = rate.id;
                        checkbox.className = 'tz-tax-rate-checkbox';
                        
                        label.appendChild(checkbox);
                        label.appendChild(document.createTextNode(rate.name + ' (' + rate.rate + '% - ' + rate.region + ')'));
                        container.appendChild(label);
                    });
                }
            } catch (error) {
                console.error('Failed to load tax rates:', error);
            }
        }

        /**
         * ========== 会员等级渲染 ==========
         */

        function renderMembershipTiers() {
            const container = document.getElementById('tz-product-membership-list');
            if (!container) return;

            const tiers = config.membershipTiers || [];
            
            if (!tiers.length) {
                container.innerHTML = '<p style="color:#6b7280;margin:0;">暂无会员等级，请先在 Loyalty & Points 页面配置。</p>';
                return;
            }
            
            // 获取已保存的会员等级
            const savedLevels = (config.initialPayload && config.initialPayload.preview && config.initialPayload.preview.membership && config.initialPayload.preview.membership.levels) || [];
            
            container.innerHTML = '';
            tiers.forEach(function(tier) {
                const label = document.createElement('label');
                label.style.cssText = 'display:inline-flex;align-items:center;gap:6px;padding:6px 12px;background:#fff;border:1px solid #e5e7eb;border-radius:6px;cursor:pointer;transition:all 0.2s;';
                label.onmouseover = function() { this.style.borderColor = '#3b82f6'; };
                label.onmouseout = function() { this.style.borderColor = '#e5e7eb'; };
                
                const checkbox = document.createElement('input');
                checkbox.type = 'checkbox';
                checkbox.value = tier.code;
                checkbox.className = 'tz-membership-checkbox';
                
                // 如果该等级已保存，则勾选
                if (savedLevels.includes(tier.code)) {
                    checkbox.checked = true;
                }
                
                const span = document.createElement('span');
                span.textContent = tier.label;
                
                label.appendChild(checkbox);
                label.appendChild(span);
                container.appendChild(label);
            });
        }

        // 会员等级全选按钮
        const memberSelectAllBtn = document.getElementById('tz-member-select-all');
        if (memberSelectAllBtn) {
            memberSelectAllBtn.addEventListener('click', function(e) {
                e.preventDefault();
                const checkboxes = document.querySelectorAll('.tz-membership-checkbox');
                checkboxes.forEach(function(cb) {
                    cb.checked = true;
                });
            });
        }

        // 会员等级清除按钮
        const memberSelectClearBtn = document.getElementById('tz-member-select-clear');
        if (memberSelectClearBtn) {
            memberSelectClearBtn.addEventListener('click', function(e) {
                e.preventDefault();
                const checkboxes = document.querySelectorAll('.tz-membership-checkbox');
                checkboxes.forEach(function(cb) {
                    cb.checked = false;
                });
            });
        }

        /**
         * ========== 配送模板加载 ==========
         */

        async function loadShippingTemplates() {
            if (!config.shippingTemplatesEndpoint) return;

            const select = document.getElementById('tz-product-shipping-template');
            if (!select) return;

            try {
                const result = await apiRequest(config.shippingTemplatesEndpoint);
                
                if (result.ok && result.data.items) {
                    result.data.items.filter(t => t.is_active).forEach(function(template) {
                        const option = document.createElement('option');
                        option.value = template.id;
                        option.textContent = template.template_name + (template.description ? ' - ' + template.description : '');
                        select.appendChild(option);
                    });
                }
            } catch (error) {
                console.error('Failed to load shipping templates:', error);
            }
        }

        /**
         * ========== 属性选择器与 SKU 自动生成 ==========
         */

        const attributesList = document.getElementById('tz-attributes-list');
        const attributesLoading = document.getElementById('tz-attributes-loading');
        const attributesEmpty = document.getElementById('tz-attributes-empty');
        const generateSkusBtn = document.getElementById('tz-generate-skus');
        let loadedAttributes = [];

        async function loadProductAttributes() {
            if (!config.attributesEndpoint || !attributesList) return;

            try {
                const result = await apiRequest(config.attributesEndpoint + '?per_page=100');
                
                if (!result.ok) {
                    if (attributesLoading) attributesLoading.style.display = 'none';
                    if (attributesEmpty) attributesEmpty.style.display = 'block';
                    return;
                }

                const skuAttributes = (result.data.items || []).filter(
                    attr => attr.is_enabled && attr.affects_sku
                );

                if (skuAttributes.length === 0) {
                    if (attributesLoading) attributesLoading.style.display = 'none';
                    if (attributesEmpty) attributesEmpty.style.display = 'block';
                    return;
                }

                // 加载属性值
                for (const attr of skuAttributes) {
                    const valuesResult = await apiRequest(
                        config.attributesEndpoint + '/' + attr.id + '/values'
                    );
                    if (valuesResult.ok) {
                        attr.values = (valuesResult.data.items || []).filter(v => v.is_enabled);
                    }
                }

                loadedAttributes = skuAttributes;
                renderAttributeSelectors(skuAttributes);
                
                if (attributesLoading) attributesLoading.style.display = 'none';
                if (generateSkusBtn) generateSkusBtn.style.display = 'inline-block';

            } catch (error) {
                console.error('Failed to load attributes:', error);
            }
        }

        function renderAttributeSelectors(attributes) {
            if (!attributesList) return;
            
            attributesList.innerHTML = '';

            attributes.forEach(function(attr) {
                const attrDiv = document.createElement('div');
                attrDiv.className = 'tz-attribute-item';
                attrDiv.style.cssText = 'padding:12px;background:#fff;border:1px solid #d1d5db;border-radius:6px;';

                const header = document.createElement('div');
                header.style.cssText = 'display:flex;align-items:center;gap:8px;margin-bottom:8px;';
                
                const toggle = document.createElement('input');
                toggle.type = 'checkbox';
                toggle.className = 'attr-toggle';
                toggle.dataset.attrId = attr.id;
                toggle.dataset.attrName = attr.slug;
                toggle.id = 'attr-toggle-' + attr.id;

                const label = document.createElement('label');
                label.htmlFor = 'attr-toggle-' + attr.id;
                label.style.cssText = 'font-weight:600;cursor:pointer;';
                label.textContent = attr.name + ' (' + attr.slug + ')';

                header.appendChild(toggle);
                header.appendChild(label);

                const valuesDiv = document.createElement('div');
                valuesDiv.className = 'attr-values';
                valuesDiv.dataset.attrId = attr.id;
                valuesDiv.style.cssText = 'display:none;margin-top:8px;padding-left:24px;flex-direction:column;gap:6px;';

                (attr.values || []).forEach(function(val) {
                    const valueLabel = document.createElement('label');
                    valueLabel.style.cssText = 'display:flex;align-items:center;gap:6px;';
                    
                    const checkbox = document.createElement('input');
                    checkbox.type = 'checkbox';
                    checkbox.value = val.slug;
                    checkbox.dataset.name = val.name;
                    
                    valueLabel.appendChild(checkbox);
                    valueLabel.appendChild(document.createTextNode(val.name));
                    valuesDiv.appendChild(valueLabel);
                });

                toggle.addEventListener('change', function() {
                    valuesDiv.style.display = this.checked ? 'flex' : 'none';
                });

                attrDiv.appendChild(header);
                attrDiv.appendChild(valuesDiv);
                attributesList.appendChild(attrDiv);
            });
        }

        function generateSKUCombinations() {
            const selectedAttrs = {};

            document.querySelectorAll('.attr-toggle:checked').forEach(function(toggle) {
                const attrId = toggle.dataset.attrId;
                const attrName = toggle.dataset.attrName;
                const values = [];

                document.querySelectorAll(`.attr-values[data-attr-id="${attrId}"] input[type="checkbox"]:checked`).forEach(function(checkbox) {
                    values.push({
                        slug: checkbox.value,
                        name: checkbox.dataset.name
                    });
                });

                if (values.length > 0) {
                    selectedAttrs[attrName] = values;
                }
            });

            const keys = Object.keys(selectedAttrs);
            if (keys.length === 0) {
                showNotice(elements.notice, 'warning', '请至少选择一个属性及其值');
                return;
            }

            const combinations = cartesianProduct(selectedAttrs);
            
            if (combinations.length > 100) {
                if (!confirm('将生成 ' + combinations.length + ' 个 SKU 组合，是否继续？')) {
                    return;
                }
            }

            skuRows = [];

            combinations.forEach(function(combo) {
                const attrParts = [];
                const codeParts = [];
                
                Object.entries(combo).forEach(function([key, value]) {
                    attrParts.push(key + '=' + value);
                    codeParts.push(value.toUpperCase());
                });

                const skuCode = 'SKU-' + codeParts.join('-');
                const attrString = attrParts.join(';');

                skuRows.push({
                    code: skuCode,
                    attributes_input: attrString,
                    attributes: combo,
                    price_regular: '',
                    price_sale: '',
                    stock: 100,
                    barcode: '',
                    weight: ''
                });
            });

            renderSkuTable();
            showNotice(elements.notice, 'success', '已生成 ' + combinations.length + ' 个 SKU 组合');
        }

        function cartesianProduct(attrs) {
            const keys = Object.keys(attrs);
            if (keys.length === 0) return [];

            let result = [{}];

            keys.forEach(function(key) {
                const temp = [];
                result.forEach(function(item) {
                    attrs[key].forEach(function(value) {
                        const newItem = Object.assign({}, item);
                        newItem[key] = value.slug;
                        temp.push(newItem);
                    });
                });
                result = temp;
            });

            return result;
        }

        /**
         * ========== 表单收集与保存 ==========
         */

        function collectProductPayload() {
            const payload = {
                title: getInputValue('tz-product-title').trim(),
                excerpt: getInputValue('tz-product-excerpt').trim(),
                slug: getInputValue('tz-product-slug').trim(),
                urllink_custom_path: getInputValue('tz-product-urllink-path').trim(),
                status: getInputValue('tz-product-status') || 'draft',
                content_markdown: getInputValue('tz-product-content-markdown'),
                price_regular: parseFloat(getInputValue('tz-product-price-regular')) || 0,
                price_sale: parseFloat(getInputValue('tz-product-price-sale')) || 0,
                stock: parseInt(getInputValue('tz-product-stock')) || 0,
                sku_data: skuRows
            };

            // 收集税率
            const taxRateIds = [];
            document.querySelectorAll('.tz-tax-rate-checkbox:checked').forEach(function(cb) {
                taxRateIds.push(parseInt(cb.value));
            });
            payload.tax_rate_ids = taxRateIds;

            // 收集会员等级
            const membershipLevels = [];
            document.querySelectorAll('.tz-membership-checkbox:checked').forEach(function(cb) {
                membershipLevels.push(cb.value);
            });
            payload.membership_levels = membershipLevels;

            // 配送模板
            const shippingTemplate = getInputValue('tz-product-shipping-template');
            if (shippingTemplate) {
                payload.shipping_template_id = parseInt(shippingTemplate);
            }

            return payload;
        }

        async function saveProduct(e) {
            if (e) e.preventDefault();

            const payload = collectProductPayload();
            const url = config.productId ? config.singleEndpoint : config.createEndpoint;
            const method = config.productId ? 'PUT' : 'POST';

            try {
                showNotice(elements.notice, 'info', strings.saving || '正在保存...');

                const result = await apiRequest(url, {
                    method: method,
                    body: JSON.stringify(payload)
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || strings.saveFailed || '保存失败');
                    return;
                }

                showNotice(elements.notice, 'success', strings.saveSuccess || '保存成功');
                
                // 如果是新建，跳转到编辑页
                if (!config.productId && result.data.id) {
                    setTimeout(function() {
                        window.location.href = '?page=tanzanite-settings-add-product&id=' + result.data.id;
                    }, 1000);
                }

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * ========== 事件监听 ==========
         */

        // SKU 表格事件
        if (elements.skuTable) {
            elements.skuTable.addEventListener('click', function(e) {
                const target = e.target;
                
                if (target.classList.contains('sku-edit')) {
                    const index = parseInt(target.dataset.index);
                    // TODO: 填充 SKU 表单进行编辑
                }
                
                if (target.classList.contains('sku-delete')) {
                    const index = parseInt(target.dataset.index);
                    deleteSku(index);
                }
            });
        }

        // 保存按钮
        if (elements.saveBtn) {
            elements.saveBtn.addEventListener('click', saveProduct);
        }

        // 重置按钮
        if (elements.resetBtn) {
            elements.resetBtn.addEventListener('click', function(e) {
                e.preventDefault();
                if (confirm(strings.resetConfirm || '确定重置？')) {
                    location.reload();
                }
            });
        }

        // 生成 SKU 按钮
        if (generateSkusBtn) {
            generateSkusBtn.addEventListener('click', generateSKUCombinations);
        }

        /**
         * ========== 初始化 ==========
         */

        // 修复 WordPress 编辑器的 "Add Media" 按钮
        function initMediaButtons() {
            console.log('Initializing media buttons...');
            
            if (typeof wp === 'undefined' || !wp.media) {
                console.error('WordPress media library not loaded');
                return;
            }

            // 创建媒体 frame
            let mediaFrame;
            
            // 监听所有 "Add Media" 按钮点击
            document.addEventListener('click', function(e) {
                const target = e.target;
                
                // 检查是否是 "Add Media" 按钮
                if (target.classList.contains('insert-media') || 
                    target.closest('.insert-media')) {
                    
                    e.preventDefault();
                    e.stopPropagation();
                    
                    console.log('Add Media button clicked');
                    
                    // 如果 frame 已存在，直接打开
                    if (mediaFrame) {
                        mediaFrame.open();
                        return;
                    }
                    
                    // 创建新的 media frame
                    mediaFrame = wp.media({
                        title: '选择或上传图片',
                        button: {
                            text: '插入图片'
                        },
                        multiple: false
                    });
                    
                    // 选择图片后的回调
                    mediaFrame.on('select', function() {
                        const attachment = mediaFrame.state().get('selection').first().toJSON();
                        console.log('Image selected:', attachment);
                        
                        // 获取当前编辑器
                        const editorId = 'tz-product-content';
                        const editor = window.tinymce ? window.tinymce.get(editorId) : null;
                        
                        if (editor && !editor.isHidden()) {
                            // 可视化模式
                            const imgHtml = '<img src="' + attachment.url + '" alt="' + (attachment.alt || attachment.title || '') + '" />';
                            editor.insertContent(imgHtml);
                            console.log('Image inserted into visual editor');
                        } else {
                            // 文本模式
                            const textarea = document.getElementById(editorId);
                            if (textarea) {
                                const imgHtml = '<img src="' + attachment.url + '" alt="' + (attachment.alt || attachment.title || '') + '" />';
                                textarea.value += imgHtml;
                                console.log('Image inserted into text editor');
                            }
                        }
                    });
                    
                    // 打开媒体库
                    mediaFrame.open();
                }
            }, true); // 使用捕获阶段
            
            console.log('Media buttons initialized');
        }

        // 延迟初始化，确保 WordPress 媒体库已加载
        setTimeout(initMediaButtons, 500);

        /**
         * ========== Markdown 模板按钮 ==========
         */
        function insertMarkdownTemplate(templateKey) {
            console.log('insertMarkdownTemplate called with:', templateKey);
            console.log('config.markdownTemplates:', config.markdownTemplates);
            
            const templates = config.markdownTemplates || {};
            const templateContent = templates[templateKey];
            
            if (!templateContent) {
                console.error('Template not found:', templateKey);
                console.error('Available templates:', Object.keys(templates));
                alert('未找到模板：' + templateKey);
                return;
            }
            
            console.log('Template content:', templateContent);
            
            // 获取编辑器
            const editorId = 'tz-product-content';
            const editor = window.tinymce ? window.tinymce.get(editorId) : null;
            
            console.log('Editor:', editor);
            console.log('Editor hidden:', editor ? editor.isHidden() : 'N/A');
            
            if (editor && !editor.isHidden()) {
                // 可视化模式 - 将 Markdown 转换为 HTML
                const htmlContent = convertMarkdownToHtml(templateContent);
                console.log('HTML content:', htmlContent);
                editor.insertContent(htmlContent);
                console.log('Template inserted into visual editor');
                alert('模板已插入到可视化编辑器');
            } else {
                // 文本模式 - 直接插入 Markdown
                const textarea = document.getElementById(editorId);
                console.log('Textarea:', textarea);
                if (textarea) {
                    const cursorPos = textarea.selectionStart || 0;
                    const textBefore = textarea.value.substring(0, cursorPos);
                    const textAfter = textarea.value.substring(cursorPos);
                    textarea.value = textBefore + '\n\n' + templateContent + '\n\n' + textAfter;
                    console.log('Template inserted into text editor');
                    alert('模板已插入到文本编辑器');
                } else {
                    console.error('Textarea not found');
                    alert('未找到编辑器');
                }
            }
        }
        
        /**
         * 简单的 Markdown 转 HTML
         */
        function convertMarkdownToHtml(markdown) {
            let html = markdown;
            
            // 标题
            html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
            html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
            html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');
            
            // 列表
            html = html.replace(/^\- (.*$)/gim, '<li>$1</li>');
            html = html.replace(/^\d+\. (.*$)/gim, '<li>$1</li>');
            
            // 包裹列表项
            html = html.replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>');
            
            // 引用
            html = html.replace(/^> (.*$)/gim, '<blockquote>$1</blockquote>');
            
            // 表格（简单处理）
            html = html.replace(/\|/g, '</td><td>');
            html = html.replace(/^(.+)$/gim, function(match) {
                if (match.includes('</td>')) {
                    return '<tr><td>' + match + '</tr>';
                }
                return match;
            });
            
            // 段落
            html = html.replace(/\n\n/g, '</p><p>');
            html = '<p>' + html + '</p>';
            
            // 清理
            html = html.replace(/<p><\/p>/g, '');
            html = html.replace(/<p>(<h[1-6]>)/g, '$1');
            html = html.replace(/(<\/h[1-6]>)<\/p>/g, '$1');
            html = html.replace(/<p>(<ul>)/g, '$1');
            html = html.replace(/(<\/ul>)<\/p>/g, '$1');
            html = html.replace(/<p>(<blockquote>)/g, '$1');
            html = html.replace(/(<\/blockquote>)<\/p>/g, '$1');
            
            return html;
        }
        
        // 绑定模板按钮事件
        document.querySelectorAll('[data-markdown-template]').forEach(function(btn) {
            btn.addEventListener('click', function(e) {
                e.preventDefault();
                const templateKey = this.getAttribute('data-markdown-template');
                insertMarkdownTemplate(templateKey);
            });
        });

        /**
         * ========== 阶梯价格管理 ==========
         */
        const tierTable = document.getElementById('tz-tier-pricing-table');
        const tierTableBody = tierTable ? tierTable.querySelector('tbody') : null;
        const tierEmptyNotice = document.getElementById('tz-tier-empty');
        const tierForm = document.getElementById('tz-tier-form');
        const tierIndexInput = document.getElementById('tz-tier-index');
        const tierMinInput = document.getElementById('tz-tier-min');
        const tierMaxInput = document.getElementById('tz-tier-max');
        const tierPriceInput = document.getElementById('tz-tier-price');
        const tierNoteInput = document.getElementById('tz-tier-note');
        const tierSubmitBtn = document.getElementById('tz-tier-submit');
        const tierResetBtn = document.getElementById('tz-tier-reset');
        const tierTemplateBtn = document.getElementById('tz-tier-template');
        const tierHiddenInput = document.getElementById('tz-product-tier-pricing');

        function clearTierForm() {
            if (!tierForm) return;
            tierForm.reset();
            if (tierIndexInput) tierIndexInput.value = '';
        }

        function syncTierHiddenInput() {
            if (!tierHiddenInput) return;
            const payload = tierRows.map(row => {
                const entry = { min_qty: row.min_qty, price: row.price };
                if (row.max_qty !== null && row.max_qty !== undefined) {
                    entry.max_qty = row.max_qty;
                }
                if (row.note) entry.note = row.note;
                return entry;
            });
            tierHiddenInput.value = payload.length ? JSON.stringify(payload) : '';
        }

        function validateTierRows(rows) {
            if (!Array.isArray(rows) || !rows.length) {
                return { valid: true, message: '' };
            }

            let previousMax = null;
            for (let i = 0; i < rows.length; i++) {
                const row = rows[i];
                const min = Number(row.min_qty);
                const max = row.max_qty === null || row.max_qty === undefined ? null : Number(row.max_qty);
                const price = Number(row.price);

                if (!Number.isFinite(min) || min <= 0 || !Number.isFinite(price) || price < 0) {
                    return { valid: false, message: strings.tierErrorRequired || '请填写有效的最小数量和单价。' };
                }

                if (max !== null && (!Number.isFinite(max) || max < min)) {
                    return { valid: false, message: strings.tierErrorRange || '最大数量必须大于或等于最小数量。' };
                }

                if (i > 0) {
                    if (previousMax === null) {
                        return { valid: false, message: strings.tierErrorLastUnlimited || '只有最后一个阶梯可以不设置最大数量。' };
                    }
                    if (min <= previousMax) {
                        return { valid: false, message: strings.tierErrorOverlap || '阶梯区间存在重叠或顺序错误。' };
                    }
                }
                previousMax = max;
            }
            return { valid: true, message: '' };
        }

        function renderTierTable() {
            if (!tierTableBody || !tierEmptyNotice) return;

            tierTableBody.innerHTML = '';

            if (!tierRows.length) {
                tierEmptyNotice.style.display = 'block';
                syncTierHiddenInput();
                return;
            }

            tierEmptyNotice.style.display = 'none';

            tierRows.forEach((row, index) => {
                const tr = document.createElement('tr');

                const minCell = document.createElement('td');
                minCell.textContent = String(row.min_qty);
                tr.appendChild(minCell);

                const maxCell = document.createElement('td');
                maxCell.textContent = row.max_qty === null || row.max_qty === undefined ? '∞' : String(row.max_qty);
                tr.appendChild(maxCell);

                const priceCell = document.createElement('td');
                priceCell.textContent = Number(row.price).toFixed(2);
                tr.appendChild(priceCell);

                const noteCell = document.createElement('td');
                noteCell.textContent = row.note || '';
                tr.appendChild(noteCell);

                const actionCell = document.createElement('td');
                actionCell.style.display = 'flex';
                actionCell.style.gap = '6px';

                const editBtn = document.createElement('button');
                editBtn.type = 'button';
                editBtn.className = 'button';
                editBtn.textContent = strings.skuEdit || '编辑';
                editBtn.addEventListener('click', () => {
                    tierIndexInput.value = String(index);
                    tierMinInput.value = row.min_qty;
                    tierMaxInput.value = row.max_qty === null || row.max_qty === undefined ? '' : row.max_qty;
                    tierPriceInput.value = row.price;
                    tierNoteInput.value = row.note || '';
                    tierMinInput.focus();
                });

                const deleteBtn = document.createElement('button');
                deleteBtn.type = 'button';
                deleteBtn.className = 'button button-secondary';
                deleteBtn.textContent = strings.skuDelete || '删除';
                deleteBtn.addEventListener('click', () => {
                    tierRows.splice(index, 1);
                    renderTierTable();
                    clearTierForm();
                });

                actionCell.appendChild(editBtn);
                actionCell.appendChild(deleteBtn);
                tr.appendChild(actionCell);

                tierTableBody.appendChild(tr);
            });

            syncTierHiddenInput();
        }

        // 阶梯价表单提交
        if (tierSubmitBtn) {
            tierSubmitBtn.addEventListener('click', function(e) {
                e.preventDefault();

                const min = parseInt(tierMinInput.value, 10);
                const maxRaw = tierMaxInput.value;
                const max = maxRaw === '' ? null : parseInt(maxRaw, 10);
                const price = parseFloat(tierPriceInput.value);
                const note = tierNoteInput.value.trim();

                if (!Number.isFinite(min) || min <= 0 || !Number.isFinite(price) || price < 0) {
                    showNotice(elements.notice, 'error', strings.tierErrorRequired || '请填写有效的最小数量和单价。');
                    return;
                }

                if (max !== null && (!Number.isFinite(max) || max < min)) {
                    showNotice(elements.notice, 'error', strings.tierErrorRange || '最大数量必须大于或等于最小数量。');
                    return;
                }

                const editingIndex = tierIndexInput.value ? parseInt(tierIndexInput.value, 10) : -1;
                const nextRows = tierRows.filter((_, idx) => idx !== editingIndex);

                nextRows.push({
                    min_qty: min,
                    max_qty: max,
                    price: parseFloat(price.toFixed(2)),
                    note
                });

                nextRows.sort((a, b) => a.min_qty - b.min_qty);
                const validation = validateTierRows(nextRows);
                if (!validation.valid) {
                    showNotice(elements.notice, 'error', validation.message);
                    return;
                }

                tierRows = nextRows;
                renderTierTable();
                clearTierForm();
                showNotice(elements.notice, 'success', strings.tierSaved || '阶梯价配置已更新。');
            });
        }

        // 清空阶梯价表单
        if (tierResetBtn) {
            tierResetBtn.addEventListener('click', function(e) {
                e.preventDefault();
                clearTierForm();
            });
        }

        // 应用阶梯价模板
        if (tierTemplateBtn) {
            tierTemplateBtn.addEventListener('click', function(e) {
                e.preventDefault();
                if (tierRows.length && !window.confirm(strings.tierTemplateConfirm || '此操作将覆盖现有阶梯价配置，是否继续？')) {
                    return;
                }

                const basePrice = parseFloat(getInputValue('tz-product-price-sale')) || parseFloat(getInputValue('tz-product-price-regular')) || 0;
                const tierOnePrice = basePrice > 0 ? parseFloat((basePrice * 0.95).toFixed(2)) : 0;
                const tierTwoPrice = basePrice > 0 ? parseFloat((basePrice * 0.9).toFixed(2)) : 0;

                tierRows = [
                    { min_qty: 1, max_qty: 9, price: basePrice || 0, note: '' },
                    { min_qty: 10, max_qty: 49, price: tierOnePrice, note: '95%' },
                    { min_qty: 50, max_qty: null, price: tierTwoPrice, note: '90%' }
                ];

                renderTierTable();
                clearTierForm();
                showNotice(elements.notice, 'info', strings.tierTemplateApplied || '示例阶梯价已插入。');
            });
        }

        // 初始化阶梯价表格
        renderTierTable();

        /**
         * ========== SKU 表单按钮 ==========
         */
        const skuFormSubmitBtn = document.getElementById('tz-product-sku-form-submit');
        const skuFormResetBtn = document.getElementById('tz-product-sku-form-reset');
        const skuFormIndexInput = document.getElementById('tz-product-sku-form-index');
        const skuFormCodeInput = document.getElementById('tz-product-sku-form-code');
        const skuFormAttrsInput = document.getElementById('tz-product-sku-form-attrs');
        const skuFormPriceRegularInput = document.getElementById('tz-product-sku-form-price-regular');
        const skuFormPriceSaleInput = document.getElementById('tz-product-sku-form-price-sale');
        const skuFormStockInput = document.getElementById('tz-product-sku-form-stock');
        const skuFormBarcodeInput = document.getElementById('tz-product-sku-form-barcode');
        const skuFormWeightInput = document.getElementById('tz-product-sku-form-weight');

        // SKU 表单提交
        if (skuFormSubmitBtn) {
            skuFormSubmitBtn.addEventListener('click', function(e) {
                e.preventDefault();

                const code = skuFormCodeInput ? skuFormCodeInput.value.trim() : '';
                if (!code) {
                    showNotice(elements.notice, 'error', strings.skuCodeRequired || '请填写 SKU 编码。');
                    return;
                }

                const editIndex = skuFormIndexInput && skuFormIndexInput.value ? parseInt(skuFormIndexInput.value, 10) : -1;

                // 检查重复（排除正在编辑的项）
                const isDuplicate = skuRows.some((row, idx) => {
                    return idx !== editIndex && row.sku_code === code;
                });

                if (isDuplicate) {
                    showNotice(elements.notice, 'error', strings.skuDuplicate || '该 SKU 编码已存在。');
                    return;
                }

                const skuData = {
                    sku_code: code,
                    attributes_input: skuFormAttrsInput ? skuFormAttrsInput.value.trim() : '',
                    price_regular: skuFormPriceRegularInput ? parseFloat(skuFormPriceRegularInput.value) || 0 : 0,
                    price_sale: skuFormPriceSaleInput ? parseFloat(skuFormPriceSaleInput.value) || 0 : 0,
                    stock_qty: skuFormStockInput ? parseInt(skuFormStockInput.value, 10) || 0 : 0,
                    barcode: skuFormBarcodeInput ? skuFormBarcodeInput.value.trim() : '',
                    weight: skuFormWeightInput ? parseFloat(skuFormWeightInput.value) || 0 : 0
                };

                if (addOrUpdateSku(skuData, editIndex)) {
                    showNotice(elements.notice, 'success', strings.skuSaved || 'SKU 已保存。');
                    // 清空表单
                    if (elements.skuForm) {
                        elements.skuForm.reset();
                        if (skuFormIndexInput) skuFormIndexInput.value = '';
                    }
                }
            });
        }

        // SKU 表单重置
        if (skuFormResetBtn) {
            skuFormResetBtn.addEventListener('click', function(e) {
                e.preventDefault();
                if (elements.skuForm) {
                    elements.skuForm.reset();
                    if (skuFormIndexInput) skuFormIndexInput.value = '';
                }
            });
        }

        loadTaxRates();
        loadShippingTemplates();
        loadProductAttributes();
        renderMembershipTiers();
        renderSkuTable();
        renderFeaturedMedia();
        renderGalleryMedia();
        renderVideoMedia();
        initFeaturedSelectors();
        initGallerySelectors();
        initVideoSelectors();

        console.log('Product Editor initialized');
    });
})();
