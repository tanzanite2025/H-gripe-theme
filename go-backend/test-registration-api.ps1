# Product Registration API 测试脚本
# 测试所有产品注册和保修相关的 API 端点

$baseUrl = "http://localhost:9000"
$apiUrl = "$baseUrl/api/v1"

Write-Host "=== Product Registration API 测试 ===" -ForegroundColor Cyan
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

# ========== 公开 API 测试 ==========

Write-Host "`n=== 公开 API 测试 ===" -ForegroundColor Cyan

# 1. 验证序列号
$verifyRequest = @{
    serial_number = "SN2024001001"
}

Test-API -Name "验证序列号" `
    -Method "POST" `
    -Url "$apiUrl/registrations/verify" `
    -Body $verifyRequest

# ========== 需要认证的 API 测试 ==========

Write-Host "`n=== 需要认证的 API 测试 ===" -ForegroundColor Cyan
Write-Host "注意: 这些测试需要认证，当前会返回 401" -ForegroundColor Yellow

# 2. 创建产品注册（需要认证）
$newRegistration = @{
    product_id = 1
    serial_number = "SN2024TEST001"
    purchase_date = "2024-05-01T10:00:00Z"
    retailer = "Test Store"
    warranty_period = 12
    notes = "Test registration"
}

Test-API -Name "创建产品注册 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/registrations" `
    -Body $newRegistration `
    -ExpectedStatus 401

# 3. 获取用户注册列表（需要认证）
Test-API -Name "获取用户注册列表 (需要认证)" `
    -Url "$apiUrl/registrations?page=1&page_size=10" `
    -ExpectedStatus 401

# 4. 获取注册详情（需要认证）
Test-API -Name "获取注册详情 (需要认证)" `
    -Url "$apiUrl/registrations/1" `
    -ExpectedStatus 401

# 5. 更新注册信息（需要认证）
$updateRegistration = @{
    notes = "Updated notes"
    retailer = "Updated Store"
}

Test-API -Name "更新注册信息 (需要认证)" `
    -Method "PUT" `
    -Url "$apiUrl/registrations/1" `
    -Body $updateRegistration `
    -ExpectedStatus 401

# 6. 创建保修申请（需要认证）
$newClaim = @{
    registration_id = 1
    issue_type = "defect"
    description = "Test warranty claim"
}

Test-API -Name "创建保修申请 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/registrations/warranty-claims" `
    -Body $newClaim `
    -ExpectedStatus 401

# 7. 获取保修申请详情（需要认证）
Test-API -Name "获取保修申请详情 (需要认证)" `
    -Url "$apiUrl/registrations/warranty-claims/1" `
    -ExpectedStatus 401

# 8. 获取注册的保修申请列表（需要认证）
Test-API -Name "获取注册的保修申请列表 (需要认证)" `
    -Url "$apiUrl/registrations/1/warranty-claims" `
    -ExpectedStatus 401

# ========== 管理员 API 测试 ==========

Write-Host "`n=== 管理员 API 测试 ===" -ForegroundColor Cyan
Write-Host "注意: 这些测试需要管理员权限，当前会返回 401" -ForegroundColor Yellow

# 9. 获取所有注册（管理员）
Test-API -Name "获取所有注册 (管理员)" `
    -Url "$apiUrl/admin/registrations?page=1&page_size=10" `
    -ExpectedStatus 401

# 10. 获取即将过期的保修（管理员）
Test-API -Name "获取即将过期的保修 (管理员)" `
    -Url "$apiUrl/admin/registrations/expiring?days=30" `
    -ExpectedStatus 401

# 11. 获取注册统计（管理员）
Test-API -Name "获取注册统计 (管理员)" `
    -Url "$apiUrl/admin/registrations/stats" `
    -ExpectedStatus 401

# 12. 更新注册状态（管理员）
$updateStatus = @{
    status = "expired"
}

Test-API -Name "更新注册状态 (管理员)" `
    -Method "PUT" `
    -Url "$apiUrl/admin/registrations/1/status" `
    -Body $updateStatus `
    -ExpectedStatus 401

# 13. 获取所有保修申请（管理员）
Test-API -Name "获取所有保修申请 (管理员)" `
    -Url "$apiUrl/admin/registrations/warranty-claims?page=1&page_size=10" `
    -ExpectedStatus 401

# 14. 更新保修申请状态（管理员）
$updateClaimStatus = @{
    status = "approved"
}

Test-API -Name "更新保修申请状态 (管理员)" `
    -Method "PUT" `
    -Url "$apiUrl/admin/registrations/warranty-claims/1/status" `
    -Body $updateClaimStatus `
    -ExpectedStatus 401

# ========== 边界测试 ==========

Write-Host "`n=== 边界测试 ===" -ForegroundColor Cyan

# 15. 验证不存在的序列号
$invalidSerial = @{
    serial_number = "INVALID999999"
}

Test-API -Name "验证不存在的序列号" `
    -Method "POST" `
    -Url "$apiUrl/registrations/verify" `
    -Body $invalidSerial `
    -ExpectedStatus 404

# 16. 获取不存在的注册
Test-API -Name "获取不存在的注册 (需要认证)" `
    -Url "$apiUrl/registrations/99999" `
    -ExpectedStatus 401

# 17. 无效的注册ID
Test-API -Name "无效的注册ID (需要认证)" `
    -Url "$apiUrl/registrations/invalid" `
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
