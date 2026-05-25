# Settings API 测试脚本

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Settings API 测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:9000"
$testsPassed = 0
$testsFailed = 0

# 测试函数
function Test-API {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [hashtable]$Headers = @{},
        [string]$Body = $null
    )
    
    Write-Host "测试: $Name" -ForegroundColor Yellow
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            Headers = $Headers
            ErrorAction = 'Stop'
        }
        
        if ($Body) {
            $params.Body = $Body
            $params.ContentType = 'application/json'
        }
        
        $response = Invoke-RestMethod @params
        Write-Host "  OK 测试通过" -ForegroundColor Green
        Write-Host "  响应: $($response | ConvertTo-Json -Compress -Depth 3)" -ForegroundColor Gray
        $script:testsPassed++
        return $response
    }
    catch {
        Write-Host "  X 测试失败: $_" -ForegroundColor Red
        $script:testsFailed++
        return $null
    }
    
    Write-Host ""
}

# 1. 获取站点设置
Test-API -Name "获取站点设置" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/site"

# 2. 获取快速购买设置
Test-API -Name "获取快速购买设置" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/quick-buy"

# 3. 获取 SEO 设置
Test-API -Name "获取 SEO 设置" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/seo"

# 4. 获取社交媒体设置
Test-API -Name "获取社交媒体设置" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/social"

# 5. 获取所有公开设置
Test-API -Name "获取所有公开设置" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/public"

# 6. 获取所有分组
Test-API -Name "获取所有分组" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/groups"

# 7. 获取分组设置 (site)
Test-API -Name "获取 site 分组设置" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/group/site"

# 8. 获取单个设置
Test-API -Name "获取单个设置 (site_name)" `
    -Method GET `
    -Url "$baseUrl/api/v1/settings/site_name?locale=en"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "管理员 API 测试（需要认证）" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "提示: 以下测试需要管理员认证，如果没有 token 会失败" -ForegroundColor Yellow
Write-Host ""

# 注意: 这些测试需要有效的 JWT token
# 如果你有 token，可以取消注释以下代码

<#
$adminToken = "your-admin-jwt-token-here"
$adminHeaders = @{
    "Authorization" = "Bearer $adminToken"
}

# 9. 获取所有设置（管理员）
Test-API -Name "获取所有设置（管理员）" `
    -Method GET `
    -Url "$baseUrl/api/v1/admin/settings?locale=en" `
    -Headers $adminHeaders

# 10. 更新设置（管理员）
$updateBody = @{
    key = "test_setting"
    value = "test_value"
    type = "string"
    group = "test"
    locale = "en"
    is_public = $true
    description = "Test setting"
} | ConvertTo-Json

Test-API -Name "更新设置（管理员）" `
    -Method POST `
    -Url "$baseUrl/api/v1/admin/settings" `
    -Headers $adminHeaders `
    -Body $updateBody

# 11. 批量更新设置（管理员）
$batchBody = @{
    settings = @(
        @{
            key = "test_setting_1"
            value = "value_1"
            type = "string"
            group = "test"
            locale = "en"
            is_public = $true
        },
        @{
            key = "test_setting_2"
            value = "value_2"
            type = "string"
            group = "test"
            locale = "en"
            is_public = $true
        }
    )
} | ConvertTo-Json -Depth 3

Test-API -Name "批量更新设置（管理员）" `
    -Method POST `
    -Url "$baseUrl/api/v1/admin/settings/batch" `
    -Headers $adminHeaders `
    -Body $batchBody

# 12. 删除设置（管理员）
Test-API -Name "删除设置（管理员）" `
    -Method DELETE `
    -Url "$baseUrl/api/v1/admin/settings/test_setting?locale=en" `
    -Headers $adminHeaders
#>

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "测试完成" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "通过: $testsPassed" -ForegroundColor Green
Write-Host "失败: $testsFailed" -ForegroundColor Red
Write-Host ""

if ($testsFailed -eq 0) {
    Write-Host "OK 所有测试通过！" -ForegroundColor Green
} else {
    Write-Host "X 部分测试失败" -ForegroundColor Red
}
