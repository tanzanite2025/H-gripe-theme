/**
 * Tanzanite Settings - 订单打印发货单
 * 
 * @package TanzaniteSettings
 * @version 0.1.7
 */

(function() {
    'use strict';

    // 全局命名空间
    window.TanzaniteOrderPrint = window.TanzaniteOrderPrint || {};

    /**
     * 打印发货单
     * @param {Object} order - 订单数据
     */
    window.TanzaniteOrderPrint.printShippingLabel = function(order) {
        if (!order) {
            alert('订单数据不存在');
            return;
        }

        // 创建打印窗口
        const printWindow = window.open('', '_blank', 'width=800,height=600');
        
        if (!printWindow) {
            alert('无法打开打印窗口，请检查浏览器弹窗设置');
            return;
        }

        // 生成打印内容
        const printContent = generateShippingLabelHTML(order);
        
        printWindow.document.write(printContent);
        printWindow.document.close();
        
        // 等待内容加载完成后打印
        printWindow.onload = function() {
            printWindow.print();
            // 打印完成后关闭窗口
            printWindow.onafterprint = function() {
                printWindow.close();
            };
        };
    };

    /**
     * 生成发货单 HTML
     * @param {Object} order - 订单数据
     * @returns {string}
     */
    function generateShippingLabelHTML(order) {
        const now = new Date().toLocaleString('zh-CN');
        
        return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>发货单 - ${order.order_number}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: "Microsoft YaHei", Arial, sans-serif;
            font-size: 14px;
            line-height: 1.6;
            color: #333;
            padding: 20px;
        }
        
        .shipping-label {
            max-width: 800px;
            margin: 0 auto;
            border: 2px solid #000;
            padding: 20px;
        }
        
        .header {
            text-align: center;
            border-bottom: 2px solid #000;
            padding-bottom: 15px;
            margin-bottom: 20px;
        }
        
        .header h1 {
            font-size: 28px;
            margin-bottom: 10px;
        }
        
        .header .order-number {
            font-size: 18px;
            font-weight: bold;
        }
        
        .section {
            margin-bottom: 20px;
            padding: 15px;
            border: 1px solid #ddd;
        }
        
        .section-title {
            font-size: 16px;
            font-weight: bold;
            margin-bottom: 10px;
            padding-bottom: 5px;
            border-bottom: 1px solid #ddd;
        }
        
        .info-row {
            display: flex;
            padding: 5px 0;
        }
        
        .info-label {
            width: 120px;
            font-weight: bold;
        }
        
        .info-value {
            flex: 1;
        }
        
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 10px;
        }
        
        .items-table th,
        .items-table td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        
        .items-table th {
            background-color: #f5f5f5;
            font-weight: bold;
        }
        
        .items-table td {
            text-align: center;
        }
        
        .items-table td:first-child {
            text-align: left;
        }
        
        .total-row {
            font-weight: bold;
            background-color: #f9f9f9;
        }
        
        .footer {
            margin-top: 30px;
            padding-top: 15px;
            border-top: 1px solid #ddd;
            display: flex;
            justify-content: space-between;
        }
        
        .signature-box {
            width: 45%;
        }
        
        .signature-line {
            margin-top: 40px;
            border-top: 1px solid #000;
            padding-top: 5px;
            text-align: center;
        }
        
        .print-time {
            text-align: right;
            color: #666;
            font-size: 12px;
            margin-top: 20px;
        }
        
        @media print {
            body {
                padding: 0;
            }
            
            .shipping-label {
                border: none;
            }
            
            .print-time {
                display: none;
            }
        }
    </style>
</head>
<body>
    <div class="shipping-label">
        <div class="header">
            <h1>发货单</h1>
            <div class="order-number">订单号: ${order.order_number}</div>
        </div>
        
        <div class="section">
            <div class="section-title">收货信息</div>
            <div class="info-row">
                <div class="info-label">收货人:</div>
                <div class="info-value">${order.customer_name || '-'}</div>
            </div>
            <div class="info-row">
                <div class="info-label">联系电话:</div>
                <div class="info-value">${order.customer_phone || '-'}</div>
            </div>
            <div class="info-row">
                <div class="info-label">收货地址:</div>
                <div class="info-value">${order.shipping_address || '-'}</div>
            </div>
        </div>
        
        <div class="section">
            <div class="section-title">物流信息</div>
            <div class="info-row">
                <div class="info-label">物流公司:</div>
                <div class="info-value">${order.tracking_provider || '-'}</div>
            </div>
            <div class="info-row">
                <div class="info-label">运单号:</div>
                <div class="info-value">${order.tracking_number || '-'}</div>
            </div>
        </div>
        
        <div class="section">
            <div class="section-title">商品明细</div>
            <table class="items-table">
                <thead>
                    <tr>
                        <th>商品名称</th>
                        <th>SKU</th>
                        <th>单价</th>
                        <th>数量</th>
                        <th>小计</th>
                    </tr>
                </thead>
                <tbody>
                    ${generateItemsRows(order.items || [])}
                    <tr class="total-row">
                        <td colspan="4" style="text-align:right;">商品小计:</td>
                        <td>¥${order.subtotal || 0}</td>
                    </tr>
                    <tr>
                        <td colspan="4" style="text-align:right;">运费:</td>
                        <td>¥${order.shipping_fee || 0}</td>
                    </tr>
                    <tr>
                        <td colspan="4" style="text-align:right;">优惠:</td>
                        <td>-¥${order.discount_amount || 0}</td>
                    </tr>
                    <tr class="total-row">
                        <td colspan="4" style="text-align:right;">实付总额:</td>
                        <td>¥${order.total_amount || 0}</td>
                    </tr>
                </tbody>
            </table>
        </div>
        
        <div class="section">
            <div class="section-title">备注信息</div>
            <div class="info-row">
                <div class="info-label">客户备注:</div>
                <div class="info-value">${order.customer_note || '无'}</div>
            </div>
            <div class="info-row">
                <div class="info-label">客服备注:</div>
                <div class="info-value">${order.admin_note || '无'}</div>
            </div>
        </div>
        
        <div class="footer">
            <div class="signature-box">
                <div>发货人签名:</div>
                <div class="signature-line"></div>
            </div>
            <div class="signature-box">
                <div>收货人签名:</div>
                <div class="signature-line"></div>
            </div>
        </div>
        
        <div class="print-time">打印时间: ${now}</div>
    </div>
</body>
</html>
        `;
    }

    /**
     * 生成商品行 HTML
     * @param {Array} items - 商品列表
     * @returns {string}
     */
    function generateItemsRows(items) {
        if (!items || items.length === 0) {
            return '<tr><td colspan="5" style="text-align:center;">暂无商品</td></tr>';
        }

        return items.map(item => `
            <tr>
                <td>${item.product_title || '-'}</td>
                <td>${item.sku_code || '-'}</td>
                <td>¥${item.unit_price || 0}</td>
                <td>${item.quantity || 0}</td>
                <td>¥${item.subtotal || 0}</td>
            </tr>
        `).join('');
    }

    /**
     * 批量打印发货单
     * @param {Array} orders - 订单数组
     */
    window.TanzaniteOrderPrint.printMultipleShippingLabels = function(orders) {
        if (!orders || orders.length === 0) {
            alert('没有可打印的订单');
            return;
        }

        // 为每个订单打开打印窗口
        orders.forEach((order, index) => {
            setTimeout(() => {
                window.TanzaniteOrderPrint.printShippingLabel(order);
            }, index * 500); // 延迟打开，避免浏览器阻止
        });
    };

    console.log('Order Print JS loaded');
})();
