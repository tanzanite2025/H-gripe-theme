/**
 * Tanzanite Settings - 物流公司管理
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        // 检查配置是否存在
        if (typeof TzCarriersConfig === 'undefined') {
            console.error('Carriers config not found');
            return;
        }

        const config = TzCarriersConfig;
        
        // 设置全局 nonce
        window.TanzaniteAdmin.nonce = config.nonce;
        
        const { showNotice, apiRequest, confirmAction } = window.TanzaniteAdmin;

        // DOM 元素
        const elements = {
            notice: document.getElementById('tz-carriers-notice'),
            tableBody: document.querySelector('#tz-carriers-table tbody'),
            form: document.getElementById('tz-carriers-form'),
            createBtn: document.getElementById('tz-carriers-create'),
            saveBtn: document.getElementById('tz-carriers-save'),
            resetBtn: document.getElementById('tz-carriers-reset'),
            inputs: {
                id: document.getElementById('tz-carrier-id'),
                code: document.getElementById('tz-carrier-code'),
                name: document.getElementById('tz-carrier-name'),
                contactPerson: document.getElementById('tz-carrier-contact-person'),
                contactPhone: document.getElementById('tz-carrier-contact-phone'),
                trackingUrl: document.getElementById('tz-carrier-tracking-url'),
                serviceRegions: document.getElementById('tz-carrier-service-regions'),
                isActive: document.getElementById('tz-carrier-is-active'),
                sortOrder: document.getElementById('tz-carrier-sort-order'),
                meta: document.getElementById('tz-carrier-meta')
            }
        };

        // 检查必需元素
        if (!elements.tableBody || !elements.form) {
            console.error('Required elements not found');
            return;
        }

        /**
         * 渲染物流公司列表
         */
        function renderCarriers(carriers) {
            elements.tableBody.innerHTML = '';

            if (!carriers || carriers.length === 0) {
                const tr = document.createElement('tr');
                tr.innerHTML = '<td colspan="7" style="text-align:center;">暂无物流公司</td>';
                elements.tableBody.appendChild(tr);
                return;
            }

            carriers.forEach(carrier => {
                const tr = document.createElement('tr');
                
                const statusColor = carrier.is_active ? '#10b981' : '#9ca3af';
                const statusText = carrier.is_active ? '启用' : '禁用';

                tr.innerHTML = `
                    <td><strong>${carrier.code}</strong></td>
                    <td>${carrier.name}</td>
                    <td>${carrier.contact_person || '-'}</td>
                    <td>${carrier.contact_phone || '-'}</td>
                    <td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;" title="${carrier.tracking_url || ''}">${carrier.tracking_url || '-'}</td>
                    <td><span style="color:${statusColor};font-weight:600;">${statusText}</span></td>
                    <td>
                        <button class="button-link edit-carrier" data-id="${carrier.id}">编辑</button> |
                        <button class="button-link-delete delete-carrier" data-id="${carrier.id}">删除</button>
                    </td>
                `;
                
                elements.tableBody.appendChild(tr);
            });
        }

        /**
         * 重置表单
         */
        function resetForm() {
            elements.form.reset();
            elements.inputs.id.value = '';
            showNotice(elements.notice, null);
        }

        /**
         * 填充表单
         */
        function fillForm(carrier) {
            resetForm();
            
            elements.inputs.id.value = carrier.id || '';
            elements.inputs.code.value = carrier.code || '';
            elements.inputs.name.value = carrier.name || '';
            elements.inputs.contactPerson.value = carrier.contact_person || '';
            elements.inputs.contactPhone.value = carrier.contact_phone || '';
            elements.inputs.trackingUrl.value = carrier.tracking_url || '';
            elements.inputs.serviceRegions.value = carrier.service_regions ? carrier.service_regions.join(', ') : '';
            elements.inputs.isActive.value = carrier.is_active ? '1' : '0';
            elements.inputs.sortOrder.value = carrier.sort_order || 0;
            elements.inputs.meta.value = carrier.meta ? JSON.stringify(carrier.meta, null, 2) : '';

            // 滚动到表单
            window.scrollTo({
                top: elements.form.offsetTop - 80,
                behavior: 'smooth'
            });
        }

        /**
         * 加载物流公司列表
         */
        async function loadCarriers() {
            try {
                const result = await apiRequest(config.listUrl);

                if (!result.ok) {
                    renderCarriers([]);
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                renderCarriers(result.data.items || []);
                showNotice(elements.notice, null);

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 加载单个物流公司
         */
        async function loadCarrier(id) {
            try {
                const result = await apiRequest(config.singleUrl + id);

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '加载失败');
                    return;
                }

                fillForm(result.data);

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 保存物流公司
         */
        async function saveCarrier(e) {
            e.preventDefault();

            const id = elements.inputs.id.value;
            const payload = {
                code: elements.inputs.code.value.trim(),
                name: elements.inputs.name.value.trim(),
                contact_person: elements.inputs.contactPerson.value.trim(),
                contact_phone: elements.inputs.contactPhone.value.trim(),
                tracking_url: elements.inputs.trackingUrl.value.trim(),
                service_regions: elements.inputs.serviceRegions.value.split(',').map(r => r.trim()).filter(Boolean),
                is_active: parseInt(elements.inputs.isActive.value) === 1,
                sort_order: parseInt(elements.inputs.sortOrder.value) || 0
            };

            // 解析 meta JSON
            if (elements.inputs.meta.value.trim()) {
                try {
                    payload.meta = JSON.parse(elements.inputs.meta.value);
                } catch (e) {
                    showNotice(elements.notice, 'error', 'Meta 字段 JSON 格式错误');
                    return;
                }
            }

            const url = id ? config.singleUrl + id : config.listUrl;
            const method = id ? 'PUT' : 'POST';

            try {
                const result = await apiRequest(url, {
                    method,
                    body: JSON.stringify(payload)
                });

                if (!result.ok) {
                    showNotice(elements.notice, 'error', result.data.message || '保存失败');
                    return;
                }

                showNotice(elements.notice, 'success', '保存成功');
                resetForm();
                loadCarriers();

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        /**
         * 删除物流公司
         */
        async function deleteCarrier(id) {
            if (!await confirmAction('确定要删除这个物流公司吗？')) {
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

                showNotice(elements.notice, 'success', '删除成功');
                loadCarriers();

            } catch (error) {
                showNotice(elements.notice, 'error', error.message);
            }
        }

        // 事件监听
        if (elements.createBtn) {
            elements.createBtn.addEventListener('click', resetForm);
        }

        if (elements.form) {
            elements.form.addEventListener('submit', saveCarrier);
        }

        if (elements.saveBtn) {
            elements.saveBtn.addEventListener('click', saveCarrier);
        }

        if (elements.resetBtn) {
            elements.resetBtn.addEventListener('click', resetForm);
        }

        if (elements.tableBody) {
            elements.tableBody.addEventListener('click', function(e) {
                if (e.target.classList.contains('edit-carrier')) {
                    e.preventDefault();
                    const id = parseInt(e.target.dataset.id);
                    loadCarrier(id);
                } else if (e.target.classList.contains('delete-carrier')) {
                    e.preventDefault();
                    const id = parseInt(e.target.dataset.id);
                    deleteCarrier(id);
                }
            });
        }

        // 初始化
        loadCarriers();
    });
})();
