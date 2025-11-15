/**
 * Tanzanite Settings - Products List Render
 * 渲染功能模块，负责表格和卡片视图的渲染
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    window.TzProductsList = window.TzProductsList || {};
    
    /**
     * 渲染商品列表
     */
    window.TzProductsList.render = function(items, view) {
        view = view || 'table';
        
        if (view === 'table') {
            renderTable(items);
        } else {
            renderCards(items);
        }
    };

    /**
     * 构建商品信息单元格
     */
    function buildProductInfoCell(item) {
        const wrapper = document.createElement('div');
        wrapper.style.display = 'flex';
        wrapper.style.gap = '12px';
        wrapper.style.alignItems = 'flex-start';

        if (item.thumbnail) {
            const thumb = document.createElement('img');
            thumb.src = item.thumbnail;
            thumb.alt = item.title || '#' + item.id;
            thumb.style.width = '60px';
            thumb.style.height = '60px';
            thumb.style.objectFit = 'cover';
            thumb.style.borderRadius = '6px';
            wrapper.appendChild(thumb);
        }

        const info = document.createElement('div');
        info.innerHTML = '<strong>' + (item.title || '-') + '</strong><br><small>ID: ' + item.id + ' · SKU: ' + (item.sku_count || 0) + '</small>';
        if (item.slug) {
            info.innerHTML += '<br><small>Slug: ' + item.slug + '</small>';
        }
        wrapper.appendChild(info);

        return wrapper;
    }

    /**
     * 渲染表格视图
     */
    function renderTable(items) {
        const config = window.TzProductsList.config;
        const i18n = window.TzProductsList.i18n;
        const elements = window.TzProductsList.elements;
        const formatAmount = window.TzProductsList.formatAmount;

        elements.tableBody.innerHTML = '';

        if (!Array.isArray(items) || !items.length) {
            const tr = document.createElement('tr');
            const td = document.createElement('td');
            td.colSpan = 8;
            td.textContent = i18n.noData || '暂无数据';
            td.style.textAlign = 'center';
            tr.appendChild(td);
            elements.tableBody.appendChild(tr);
            return;
        }

        items.forEach(function(item) {
            const tr = document.createElement('tr');
            tr.dataset.id = item.id;

            // 商品信息
            const productCell = document.createElement('td');
            productCell.appendChild(buildProductInfoCell(item));
            tr.appendChild(productCell);

            // 价格
            const prices = item.prices || {};
            const priceCell = document.createElement('td');
            priceCell.innerHTML = '<strong>' + formatAmount(prices.sale || prices.regular || 0) + '</strong>';
            if (prices.sale && prices.sale !== prices.regular) {
                priceCell.innerHTML += '<br><small>' + formatAmount(prices.regular) + ' → ' + formatAmount(prices.sale) + '</small>';
            }
            if (prices.member) {
                priceCell.innerHTML += '<br><small>' + (i18n.memberPriceLabel || '会员价') + ': ' + formatAmount(prices.member) + '</small>';
            }
            tr.appendChild(priceCell);

            // 库存
            const stock = item.stock || {};
            const stockCell = document.createElement('td');
            stockCell.innerHTML = (stock.quantity || 0) + '<br><small>' + (i18n.stockAlertLabel || '警戒') + ': ' + (stock.alert || 0) + '</small>';
            tr.appendChild(stockCell);

            // 积分
            const points = item.points || {};
            const pointsCell = document.createElement('td');
            pointsCell.innerHTML = (points.reward || 0) + '<br><small>' + (i18n.pointsLimitLabel || '限制') + ': ' + (points.limit || 0) + '</small>';
            tr.appendChild(pointsCell);

            // 分类
            const categoryCell = document.createElement('td');
            const categories = Array.isArray(item.categories) ? item.categories.map(function(cat) { return cat.name; }).join(', ') : '-';
            categoryCell.textContent = categories || '-';
            tr.appendChild(categoryCell);

            // 状态
            const statusCell = document.createElement('td');
            statusCell.innerHTML = '<strong>' + (item.status || '-') + '</strong>';
            if (item.sticky) {
                statusCell.innerHTML += '<br><span class="tz-badge">' + (i18n.stickyBadge || '置顶') + '</span>';
            }
            tr.appendChild(statusCell);

            // 更新时间
            const updatedAtCell = document.createElement('td');
            updatedAtCell.textContent = item.updated_at || '-';
            tr.appendChild(updatedAtCell);

            // 操作
            const actionCell = document.createElement('td');
            actionCell.style.display = 'flex';
            actionCell.style.gap = '8px';
            actionCell.style.flexWrap = 'wrap';

            const editBtn = document.createElement('a');
            editBtn.className = 'button button-small';
            editBtn.href = config.editUrl + item.id;
            editBtn.textContent = i18n.editLabel || '编辑';
            actionCell.appendChild(editBtn);

            const seoBtn = document.createElement('a');
            seoBtn.className = 'button button-small';
            seoBtn.href = config.seoUrl + item.id;
            seoBtn.textContent = i18n.seoLabel || 'SEO';
            actionCell.appendChild(seoBtn);

            if (item.preview_url) {
                const previewBtn = document.createElement('a');
                previewBtn.className = 'button button-small';
                previewBtn.href = item.preview_url;
                previewBtn.target = '_blank';
                previewBtn.rel = 'noopener noreferrer';
                previewBtn.textContent = i18n.previewLabel || '预览';
                actionCell.appendChild(previewBtn);
            }

            const copyBtn = document.createElement('button');
            copyBtn.type = 'button';
            copyBtn.className = 'button button-small tz-copy-payload';
            copyBtn.dataset.id = item.id;
            copyBtn.textContent = i18n.copyPayloadLabel || '复制';
            actionCell.appendChild(copyBtn);

            if (config.canManage) {
                const stickyBtn = document.createElement('button');
                stickyBtn.type = 'button';
                stickyBtn.className = 'button button-small tz-toggle-sticky';
                stickyBtn.dataset.id = item.id;
                stickyBtn.dataset.sticky = item.sticky ? '1' : '0';
                stickyBtn.textContent = item.sticky ? (i18n.unstickLabel || '取消置顶') : (i18n.stickLabel || '置顶');
                actionCell.appendChild(stickyBtn);

                const deleteBtn = document.createElement('button');
                deleteBtn.type = 'button';
                deleteBtn.className = 'button button-small button-link-delete tz-product-delete';
                deleteBtn.dataset.id = item.id;
                deleteBtn.textContent = i18n.deleteLabel || '删除';
                actionCell.appendChild(deleteBtn);
            }

            tr.appendChild(actionCell);
            elements.tableBody.appendChild(tr);
        });
    }

    /**
     * 渲染卡片视图
     */
    function renderCards(items) {
        const config = window.TzProductsList.config;
        const i18n = window.TzProductsList.i18n;
        const elements = window.TzProductsList.elements;
        const formatAmount = window.TzProductsList.formatAmount;

        elements.cardsWrapper.innerHTML = '';

        if (!Array.isArray(items) || !items.length) {
            const card = document.createElement('div');
            card.style.textAlign = 'center';
            card.textContent = i18n.noData || '暂无数据';
            elements.cardsWrapper.appendChild(card);
            return;
        }

        items.forEach(function(item) {
            const card = document.createElement('div');
            card.className = 'tz-product-card';
            card.dataset.id = item.id;

            const header = document.createElement('div');
            header.className = 'tz-card-header';
            header.appendChild(buildProductInfoCell(item));
            card.appendChild(header);

            const body = document.createElement('div');
            body.className = 'tz-card-body';

            const prices = item.prices || {};
            const stock = item.stock || {};
            const points = item.points || {};
            const categories = Array.isArray(item.categories) ? item.categories.map(function(cat) { return cat.name; }).join(', ') : '-';

            body.innerHTML = '<p>' + (i18n.priceLabel || '价格') + ': <strong>' + formatAmount(prices.sale || prices.regular || 0) + '</strong></p>' +
                '<p>' + (i18n.stockLabel || '库存') + ': ' + (stock.quantity || 0) + ' · ' + (i18n.stockAlertLabel || '警戒') + ': ' + (stock.alert || 0) + '</p>' +
                '<p>' + (i18n.pointsRewardLabel || '奖励积分') + ': ' + (points.reward || 0) + ' · ' + (i18n.pointsLimitLabel || '限制') + ': ' + (points.limit || 0) + '</p>' +
                '<p>' + (i18n.categoryLabel || '分类') + ': ' + (categories || '-') + '</p>' +
                '<p>' + (i18n.statusLabel || '状态') + ': ' + (item.status || '-') + '</p>';
            card.appendChild(body);

            const footer = document.createElement('div');
            footer.className = 'tz-card-footer';

            const editBtn = document.createElement('a');
            editBtn.href = config.editUrl + item.id;
            editBtn.textContent = i18n.editLabel || '编辑';
            footer.appendChild(editBtn);

            if (config.canManage) {
                const deleteBtn = document.createElement('button');
                deleteBtn.type = 'button';
                deleteBtn.className = 'button button-small button-link-delete tz-product-delete';
                deleteBtn.dataset.id = item.id;
                deleteBtn.textContent = i18n.deleteLabel || '删除';
                footer.appendChild(deleteBtn);
            }

            card.appendChild(footer);
            elements.cardsWrapper.appendChild(card);
        });
    }
})();
