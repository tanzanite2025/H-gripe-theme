# FAQ API 测试脚本

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "FAQ API 测试" -ForegroundColor Cyan
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
        
        # 显示部分响应
        if ($response.data) {
            Write-Host "  数据条数: $($response.data.Count)" -ForegroundColor Gray
        } elseif ($response.total) {
            Write-Host "  总数: $($response.total)" -ForegroundColor Gray
        }
        
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

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "公开 FAQ API 测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 获取 FAQ 列表
Test-API -Name "获取 FAQ 列表" `
    -Method GET `
    -Url "$baseUrl/api/v1/content/faqs?page=1&page_size=10"

# 2. 获取 FAQ 分类
Test-API -Name "获取 FAQ 分类" `
    -Method GET `
    -Url "$baseUrl/api/v1/content/faq-categories"

# 3. 按分类获取 FAQ
Test-API -Name "按分类获取 FAQ (General)" `
    -Method GET `
    -Url "$baseUrl/api/v1/content/faqs?category=General"

# 4. 搜索 FAQ
Test-API -Name "搜索 FAQ (shipping)" `
    -Method GET `
    -Url "$baseUrl/api/v1/content/faqs/search?q=shipping"

# 5. 获取分类下的 FAQ
Test-API -Name "获取 Shipping 分类的 FAQ" `
    -Method GET `
    -Url "$baseUrl/api/v1/content/faqs/category/Shipping"

# 6. 获取热门 FAQ
Test-API -Name "获取热门 FAQ" `
    -Method GET `
    -Url "$baseUrl/api/v1/content/faqs/popular?limit=5"

# 7. 获取单个 FAQ
Write-Host "测试: 获取单个 FAQ" -ForegroundColor Yellow
try {
    # 先获取列表，取第一个 FAQ 的 ID
    $faqs = Invoke-RestMethod -Uri "$baseUrl/api/v1/content/faqs?page=1&page_size=1"
    if ($faqs.data -and $faqs.data.Count -gt 0) {
        $faqId = $faqs.data[0].id
        $faq = Invoke-RestMethod -Uri "$baseUrl/api/v1/content/faqs/$faqId"
        Write-Host "  OK 测试通过" -ForegroundColor Green
        Write-Host "  FAQ: $($faq.question)" -ForegroundColor Gray
        $script:testsPassed++
    } else {
        Write-Host "  - 跳过（没有 FAQ 数据）" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  X 测试失败: $_" -ForegroundColor Red
    $script:testsFailed++
}
Write-Host ""

# 8. 增加浏览次数
Write-Host "测试: 增加 FAQ 浏览次数" -ForegroundColor Yellow
try {
    $faqs = Invoke-RestMethod -Uri "$baseUrl/api/v1/content/faqs?page=1&page_size=1"
    if ($faqs.data -and $faqs.data.Count -gt 0) {
        $faqId = $faqs.data[0].id
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/content/faqs/$faqId/view" -Method POST
        Write-Host "  OK 测试通过" -ForegroundColor Green
        $script:testsPassed++
    } else {
        Write-Host "  - 跳过（没有 FAQ 数据）" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  X 测试失败: $_" -ForegroundColor Red
    $script:testsFailed++
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "管理员 FAQ API 测试（需要认证）" -ForegroundColor Cyan
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

# 9. 创建 FAQ（管理员）
$createBody = @{
    question = "Test FAQ Question"
    answer = "This is a test FAQ answer."
    category = "Test"
    locale = "en"
    order = 100
    status = "published"
} | ConvertTo-Json

Test-API -Name "创建 FAQ（管理员）" `
    -Method POST `
    -Url "$baseUrl/api/v1/admin/faqs" `
    -Headers $adminHeaders `
    -Body $createBody

# 10. 更新 FAQ（管理员）
$updateBody = @{
    question = "Updated Test FAQ Question"
    answer = "This is an updated test FAQ answer."
    category = "Test"
    status = "published"
} | ConvertTo-Json

Test-API -Name "更新 FAQ（管理员）" `
    -Method PUT `
    -Url "$baseUrl/api/v1/admin/faqs/1" `
    -Headers $adminHeaders `
    -Body $updateBody

# 11. 更新 FAQ 排序（管理员）
$orderBody = @{
    order = 50
} | ConvertTo-Json

Test-API -Name "更新 FAQ 排序（管理员）" `
    -Method PUT `
    -Url "$baseUrl/api/v1/admin/faqs/1/order" `
    -Headers $adminHeaders `
    -Body $orderBody

# 12. 批量更新排序（管理员）
$batchOrderBody = @{
    orders = @{
        1 = 10
        2 = 20
        3 = 30
    }
} | ConvertTo-Json

Test-API -Name "批量更新 FAQ 排序（管理员）" `
    -Method POST `
    -Url "$baseUrl/api/v1/admin/faqs/batch-order" `
    -Headers $adminHeaders `
    -Body $batchOrderBody

# 13. 删除 FAQ（管理员）
Test-API -Name "删除 FAQ（管理员）" `
    -Method DELETE `
    -Url "$baseUrl/api/v1/admin/faqs/999" `
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
