# Ticket API 测试脚本
# 测试所有客服工单相关的 API 端点

$baseUrl = "http://localhost:9000"
$apiUrl = "$baseUrl/api/v1"

Write-Host "=== Ticket API 测试 ===" -ForegroundColor Cyan
Write-Host "基础 URL: $baseUrl`n" -ForegroundColor Gray

# 测试计数器
$testCount = 0
$passCount = 0
$failCount = 0

# 测试函数
function Test-API {
    param(
        [string]$Name,
        [string]$Method = "GET",
        [string]$Url,
        [object]$Body = $null,
        [int]$ExpectedStatus = 200
    )
    
    $script:testCount++
    Write-Host "[$script:testCount] 测试: $Name" -ForegroundColor Yellow
    Write-Host "    $Method $Url" -ForegroundColor Gray
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            ContentType = "application/json"
        }
        
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
            Write-Host "    Body: $($params.Body)" -ForegroundColor Gray
        }
        
        $response = Invoke-WebRequest @params -UseBasicParsing
        
        if ($response.StatusCode -eq $ExpectedStatus) {
            Write-Host "    ✓ 通过 (状态码: $($response.StatusCode))" -ForegroundColor Green
            $script:passCount++
            
            # 显示响应内容（格式化）
            if ($response.Content) {
                $json = $response.Content | ConvertFrom-Json
                Write-Host "    响应: $($json | ConvertTo-Json -Depth 3 -Compress)" -ForegroundColor Gray
            }
            
            return $response
        } else {
            Write-Host "    ✗ 失败 (期望: $ExpectedStatus, 实际: $($response.StatusCode))" -ForegroundColor Red
            $script:failCount++
            return $null
        }
    }
    catch {
        Write-Host "    ✗ 失败: $($_.Exception.Message)" -ForegroundColor Red
        $script:failCount++
        return $null
    }
    
    Write-Host ""
}

# ========== 需要认证的 API 测试 ==========

Write-Host "`n=== 需要认证的 API 测试 ===" -ForegroundColor Cyan
Write-Host "注意: 这些测试需要认证，当前会返回 401" -ForegroundColor Yellow

# 1. 创建工单（需要认证）
$newTicket = @{
    subject = "Test ticket"
    category = "product"
    priority = "medium"
    content = "This is a test ticket message"
}

Test-API -Name "创建工单 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/tickets" `
    -Body $newTicket `
    -ExpectedStatus 401

# 2. 获取工单列表（需要认证）
Test-API -Name "获取工单列表 (需要认证)" `
    -Url "$apiUrl/tickets?page=1&page_size=10" `
    -ExpectedStatus 401

# 3. 获取工单详情（需要认证）
Test-API -Name "获取工单详情 (需要认证)" `
    -Url "$apiUrl/tickets/1" `
    -ExpectedStatus 401

# 4. 获取工单统计（需要认证）
Test-API -Name "获取工单统计 (需要认证)" `
    -Url "$apiUrl/tickets/stats" `
    -ExpectedStatus 401

# 5. 更新工单状态（需要认证）
$updateStatus = @{
    status = "in_progress"
}

Test-API -Name "更新工单状态 (需要认证)" `
    -Method "PUT" `
    -Url "$apiUrl/tickets/1/status" `
    -Body $updateStatus `
    -ExpectedStatus 401

# 6. 关闭工单（需要认证）
Test-API -Name "关闭工单 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/tickets/1/close" `
    -ExpectedStatus 401

# 7. 添加消息（需要认证）
$newMessage = @{
    content = "This is a test message"
    attachments = @()
}

Test-API -Name "添加消息 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/tickets/1/messages" `
    -Body $newMessage `
    -ExpectedStatus 401

# 8. 获取工单消息列表（需要认证）
Test-API -Name "获取工单消息列表 (需要认证)" `
    -Url "$apiUrl/tickets/1/messages" `
    -ExpectedStatus 401

# ========== 管理员 API 测试 ==========

Write-Host "`n=== 管理员 API 测试 ===" -ForegroundColor Cyan
Write-Host "注意: 这些测试需要管理员权限，当前会返回 401" -ForegroundColor Yellow

# 9. 获取所有工单（管理员）
Test-API -Name "获取所有工单 (管理员)" `
    -Url "$apiUrl/admin/tickets?page=1&page_size=10" `
    -ExpectedStatus 401

# 10. 按状态筛选工单（管理员）
Test-API -Name "按状态筛选工单 (管理员)" `
    -Url "$apiUrl/admin/tickets?status=open&page=1&page_size=10" `
    -ExpectedStatus 401

# 11. 按优先级筛选工单（管理员）
Test-API -Name "按优先级筛选工单 (管理员)" `
    -Url "$apiUrl/admin/tickets?priority=high&page=1&page_size=10" `
    -ExpectedStatus 401

# 12. 分配工单（管理员）
$assignTicket = @{
    assigned_to = 2
}

Test-API -Name "分配工单 (管理员)" `
    -Method "POST" `
    -Url "$apiUrl/admin/tickets/1/assign" `
    -Body $assignTicket `
    -ExpectedStatus 401

# 13. 获取客服仪表板（管理员）
Test-API -Name "获取客服仪表板 (管理员)" `
    -Url "$apiUrl/admin/tickets/dashboard" `
    -ExpectedStatus 401

# 14. 获取最近的工单（管理员）
Test-API -Name "获取最近的工单 (管理员)" `
    -Url "$apiUrl/admin/tickets/recent?limit=10" `
    -ExpectedStatus 401

# ========== 边界测试 ==========

Write-Host "`n=== 边界测试 ===" -ForegroundColor Cyan

# 15. 获取不存在的工单
Test-API -Name "获取不存在的工单 (需要认证)" `
    -Url "$apiUrl/tickets/99999" `
    -ExpectedStatus 401

# 16. 无效的工单ID
Test-API -Name "无效的工单ID (需要认证)" `
    -Url "$apiUrl/tickets/invalid" `
    -ExpectedStatus 401

# 17. 无效的状态值
$invalidStatus = @{
    status = "invalid_status"
}

Test-API -Name "无效的状态值 (需要认证)" `
    -Method "PUT" `
    -Url "$apiUrl/tickets/1/status" `
    -Body $invalidStatus `
    -ExpectedStatus 401

# ========== 测试总结 ==========

Write-Host "`n=== 测试总结 ===" -ForegroundColor Cyan
Write-Host "总测试数: $testCount" -ForegroundColor White
Write-Host "通过: $passCount" -ForegroundColor Green
Write-Host "失败: $failCount" -ForegroundColor Red

if ($failCount -eq 0) {
    Write-Host "`n✓ 所有测试通过!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n✗ 有 $failCount 个测试失败" -ForegroundColor Red
    exit 1
}
