# 订单管理 API 测试脚本
# 使用方法: .\test-order-api.ps1

$baseUrl = "http://localhost:8080"
$adminEmail = "admin@tanzanite.com"
$adminPassword = "admin123"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "订单管理 API 测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 管理员登录
Write-Host "1. 管理员登录..." -ForegroundColor Yellow
$loginBody = @{
    email = $adminEmail
    password = $adminPassword
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "✓ 登录成功" -ForegroundColor Green
    Write-Host "Token: $($token.Substring(0, 20))..." -ForegroundColor Gray
} catch {
    Write-Host "✗ 登录失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host ""

# 2. 获取订单统计
Write-Host "2. 获取订单统计..." -ForegroundColor Yellow
try {
    $stats = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders/stats" -Method Get -Headers $headers
    Write-Host "✓ 统计数据获取成功" -ForegroundColor Green
    Write-Host "  总订单数: $($stats.total)" -ForegroundColor Gray
    Write-Host "  今日订单: $($stats.today)" -ForegroundColor Gray
    Write-Host "  待支付: $($stats.pending)" -ForegroundColor Gray
    Write-Host "  已支付: $($stats.paid)" -ForegroundColor Gray
    Write-Host "  已完成: $($stats.completed)" -ForegroundColor Gray
    Write-Host "  总销售额: ¥$($stats.total_revenue)" -ForegroundColor Gray
    Write-Host "  今日销售额: ¥$($stats.today_revenue)" -ForegroundColor Gray
} catch {
    Write-Host "✗ 获取统计失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 3. 获取订单列表
Write-Host "3. 获取订单列表..." -ForegroundColor Yellow
try {
    $orders = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders?page=1&page_size=10" -Method Get -Headers $headers
    Write-Host "✓ 订单列表获取成功" -ForegroundColor Green
    Write-Host "  总数: $($orders.pagination.total)" -ForegroundColor Gray
    Write-Host "  当前页: $($orders.pagination.page)" -ForegroundColor Gray
    Write-Host "  每页数量: $($orders.pagination.page_size)" -ForegroundColor Gray
    Write-Host "  订单数量: $($orders.orders.Count)" -ForegroundColor Gray
    
    if ($orders.orders.Count -gt 0) {
        Write-Host "  前3个订单:" -ForegroundColor Gray
        $orders.orders | Select-Object -First 3 | ForEach-Object {
            Write-Host "    - [$($_.id)] $($_.order_number) - ¥$($_.total_amount) ($($_.status))" -ForegroundColor Gray
        }
        
        # 保存第一个订单ID用于后续测试
        $testOrderId = $orders.orders[0].id
        $testOrderNumber = $orders.orders[0].order_number
    }
} catch {
    Write-Host "✗ 获取订单列表失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

if ($testOrderId) {
    # 4. 获取订单详情
    Write-Host "4. 获取订单详情 (ID: $testOrderId)..." -ForegroundColor Yellow
    try {
        $order = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders/$testOrderId" -Method Get -Headers $headers
        Write-Host "✓ 订单详情获取成功" -ForegroundColor Green
        Write-Host "  订单号: $($order.order.order_number)" -ForegroundColor Gray
        Write-Host "  订单状态: $($order.order.status)" -ForegroundColor Gray
        Write-Host "  支付状态: $($order.order.payment_status)" -ForegroundColor Gray
        Write-Host "  物流状态: $($order.order.shipping_status)" -ForegroundColor Gray
        Write-Host "  总金额: ¥$($order.order.total_amount)" -ForegroundColor Gray
        Write-Host "  商品数量: $($order.order.items.Count)" -ForegroundColor Gray
        Write-Host "  客户: $($order.order.shipping_address.first_name) $($order.order.shipping_address.last_name)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 获取订单详情失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""

    # 5. 更新管理员备注
    Write-Host "5. 更新管理员备注..." -ForegroundColor Yellow
    $noteBody = @{
        admin_note = "测试备注 - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    } | ConvertTo-Json

    try {
        $noteResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders/$testOrderId/admin-note" -Method Patch -Body $noteBody -Headers $headers
        Write-Host "✓ 备注更新成功" -ForegroundColor Green
    } catch {
        Write-Host "✗ 更新备注失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""
}

# 6. 筛选订单（按状态）
Write-Host "6. 筛选订单（已支付）..." -ForegroundColor Yellow
try {
    $paidOrders = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders?status=paid&page=1&page_size=5" -Method Get -Headers $headers
    Write-Host "✓ 筛选成功" -ForegroundColor Green
    Write-Host "  找到 $($paidOrders.pagination.total) 个已支付订单" -ForegroundColor Gray
} catch {
    Write-Host "✗ 筛选失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 7. 筛选订单（按支付状态）
Write-Host "7. 筛选订单（未支付）..." -ForegroundColor Yellow
try {
    $unpaidOrders = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders?payment_status=unpaid&page=1&page_size=5" -Method Get -Headers $headers
    Write-Host "✓ 筛选成功" -ForegroundColor Green
    Write-Host "  找到 $($unpaidOrders.pagination.total) 个未支付订单" -ForegroundColor Gray
} catch {
    Write-Host "✗ 筛选失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 8. 获取销售图表数据
Write-Host "8. 获取销售图表数据（最近30天）..." -ForegroundColor Yellow
try {
    $salesChart = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders/sales-chart?days=30" -Method Get -Headers $headers
    Write-Host "✓ 图表数据获取成功" -ForegroundColor Green
    Write-Host "  开始日期: $($salesChart.start_date)" -ForegroundColor Gray
    Write-Host "  结束日期: $($salesChart.end_date)" -ForegroundColor Gray
    Write-Host "  数据点数: $($salesChart.data.Count)" -ForegroundColor Gray
    
    if ($salesChart.data.Count -gt 0) {
        $totalSales = ($salesChart.data | Measure-Object -Property amount -Sum).Sum
        $totalOrders = ($salesChart.data | Measure-Object -Property count -Sum).Sum
        Write-Host "  期间总订单: $totalOrders" -ForegroundColor Gray
        Write-Host "  期间总销售额: ¥$totalSales" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ 获取图表数据失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 9. 搜索订单
Write-Host "9. 搜索订单（关键词：TZ）..." -ForegroundColor Yellow
try {
    $searchOrders = Invoke-RestMethod -Uri "$baseUrl/api/admin/orders?search=TZ&page=1&page_size=5" -Method Get -Headers $headers
    Write-Host "✓ 搜索成功" -ForegroundColor Green
    Write-Host "  找到 $($searchOrders.pagination.total) 个订单" -ForegroundColor Gray
    
    if ($searchOrders.orders.Count -gt 0) {
        $searchOrders.orders | ForEach-Object {
            Write-Host "    - $($_.order_number) (¥$($_.total_amount))" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "✗ 搜索失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "提示：" -ForegroundColor Yellow
Write-Host "1. 确保服务器正在运行: go run cmd/server/main.go" -ForegroundColor Gray
Write-Host "2. 确保数据库已初始化并有订单数据" -ForegroundColor Gray
Write-Host "3. 可以访问前端管理界面测试: http://localhost:8080/admin" -ForegroundColor Gray
Write-Host "4. 状态更新和批量操作需要在前端界面测试" -ForegroundColor Gray
