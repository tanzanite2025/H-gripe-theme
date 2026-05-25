# 订阅系统 API 测试脚本
# 测试所有订阅相关的 API 端点

$baseUrl = "http://localhost:8080/api/v1"
$adminToken = ""  # 需要先登录获取管理员 token

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "订阅系统 API 测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 测试邮箱
$testEmail = "test-$(Get-Random)@example.com"
$testEmail2 = "test-$(Get-Random)@example.com"

# ========== 公开 API 测试 ==========

Write-Host "1. 测试订阅 (POST /subscriptions)" -ForegroundColor Yellow
$subscribeBody = @{
    email = $testEmail
    source = "website"
    locale = "zh"
    tags = @("newsletter", "promotions")
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/subscriptions" -Method Post -Body $subscribeBody -ContentType "application/json"
    Write-Host "✓ 订阅成功" -ForegroundColor Green
    Write-Host "  Email: $($response.data.email)" -ForegroundColor Gray
    Write-Host "  Status: $($response.data.status)" -ForegroundColor Gray
    Write-Host "  Token: $($response.data.unsub_token)" -ForegroundColor Gray
    $unsubToken = $response.data.unsub_token
} catch {
    Write-Host "✗ 订阅失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "2. 测试重复订阅 (应该返回 409)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/subscriptions" -Method Post -Body $subscribeBody -ContentType "application/json"
    Write-Host "✗ 应该返回错误但成功了" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 409) {
        Write-Host "✓ 正确返回 409 Conflict" -ForegroundColor Green
    } else {
        Write-Host "✗ 返回了错误的状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "3. 测试获取订阅状态 (GET /subscriptions/status/:email)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/subscriptions/status/$testEmail" -Method Get
    Write-Host "✓ 获取订阅状态成功" -ForegroundColor Green
    Write-Host "  Email: $($response.data.email)" -ForegroundColor Gray
    Write-Host "  Status: $($response.data.status)" -ForegroundColor Gray
    Write-Host "  Locale: $($response.data.locale)" -ForegroundColor Gray
} catch {
    Write-Host "✗ 获取订阅状态失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "4. 测试通过令牌退订 (GET /subscriptions/unsubscribe/:token)" -ForegroundColor Yellow
if ($unsubToken) {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/subscriptions/unsubscribe/$unsubToken" -Method Get
        Write-Host "✓ 退订成功" -ForegroundColor Green
    } catch {
        Write-Host "✗ 退订失败: $($_.Exception.Message)" -ForegroundColor Red
    }
} else {
    Write-Host "⊘ 跳过（没有退订令牌）" -ForegroundColor Gray
}
Write-Host ""

Write-Host "5. 测试重新订阅 (POST /subscriptions/resubscribe)" -ForegroundColor Yellow
$resubscribeBody = @{
    email = $testEmail
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/subscriptions/resubscribe" -Method Post -Body $resubscribeBody -ContentType "application/json"
    Write-Host "✓ 重新订阅成功" -ForegroundColor Green
} catch {
    Write-Host "✗ 重新订阅失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "6. 测试通过邮箱退订 (POST /subscriptions/unsubscribe)" -ForegroundColor Yellow
$unsubscribeBody = @{
    email = $testEmail
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/subscriptions/unsubscribe" -Method Post -Body $unsubscribeBody -ContentType "application/json"
    Write-Host "✓ 通过邮箱退订成功" -ForegroundColor Green
} catch {
    Write-Host "✗ 通过邮箱退订失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 创建第二个测试订阅
Write-Host "7. 创建第二个测试订阅" -ForegroundColor Yellow
$subscribe2Body = @{
    email = $testEmail2
    source = "popup"
    locale = "en"
    tags = @("newsletter")
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/subscriptions" -Method Post -Body $subscribe2Body -ContentType "application/json"
    Write-Host "✓ 第二个订阅创建成功" -ForegroundColor Green
} catch {
    Write-Host "✗ 创建失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# ========== 管理员 API 测试 ==========

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "管理员 API 测试（需要认证）" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($adminToken -eq "") {
    Write-Host "⚠ 未设置管理员 token，跳过管理员 API 测试" -ForegroundColor Yellow
    Write-Host "请先登录获取 token，然后设置 `$adminToken 变量" -ForegroundColor Yellow
} else {
    $headers = @{
        "Authorization" = "Bearer $adminToken"
    }

    Write-Host "8. 测试获取所有订阅 (GET /admin/subscriptions)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/subscriptions?page=1&page_size=10" -Method Get -Headers $headers
        Write-Host "✓ 获取订阅列表成功" -ForegroundColor Green
        Write-Host "  总数: $($response.pagination.total)" -ForegroundColor Gray
        Write-Host "  当前页: $($response.pagination.page)" -ForegroundColor Gray
        Write-Host "  每页数量: $($response.pagination.page_size)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 获取订阅列表失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""

    Write-Host "9. 测试按状态筛选订阅 (GET /admin/subscriptions?status=active)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/subscriptions?status=active&page=1&page_size=10" -Method Get -Headers $headers
        Write-Host "✓ 获取活跃订阅成功" -ForegroundColor Green
        Write-Host "  活跃订阅数: $($response.pagination.total)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 获取活跃订阅失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""

    Write-Host "10. 测试按标签获取订阅 (GET /admin/subscriptions/tags)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/subscriptions/tags?tags=newsletter" -Method Get -Headers $headers
        Write-Host "✓ 按标签获取订阅成功" -ForegroundColor Green
        Write-Host "  匹配数量: $($response.pagination.total)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 按标签获取订阅失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""

    Write-Host "11. 测试获取订阅统计 (GET /admin/subscriptions/stats)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/subscriptions/stats" -Method Get -Headers $headers
        Write-Host "✓ 获取统计信息成功" -ForegroundColor Green
        Write-Host "  总订阅数: $($response.total)" -ForegroundColor Gray
        Write-Host "  活跃: $($response.active)" -ForegroundColor Gray
        Write-Host "  已退订: $($response.unsubscribed)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 获取统计信息失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""

    Write-Host "12. 测试导出邮箱列表 (GET /admin/subscriptions/export)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/subscriptions/export" -Method Get -Headers $headers
        Write-Host "✓ 导出邮箱列表成功" -ForegroundColor Green
        Write-Host "  邮箱数量: $($response.count)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 导出邮箱列表失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""

    Write-Host "13. 测试按标签导出邮箱 (GET /admin/subscriptions/export?tags=newsletter)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/subscriptions/export?tags=newsletter" -Method Get -Headers $headers
        Write-Host "✓ 按标签导出成功" -ForegroundColor Green
        Write-Host "  邮箱数量: $($response.count)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 按标签导出失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""

    Write-Host "14. 测试删除订阅 (DELETE /admin/subscriptions/:email)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/subscriptions/$testEmail2" -Method Delete -Headers $headers
        Write-Host "✓ 删除订阅成功" -ForegroundColor Green
    } catch {
        Write-Host "✗ 删除订阅失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "测试完成" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
