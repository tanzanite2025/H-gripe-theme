# Gallery API 测试脚本
# 测试所有 Gallery 相关的 API 端点

$baseUrl = "http://localhost:9000"
$apiUrl = "$baseUrl/api/v1"

Write-Host "=== Gallery API 测试 ===" -ForegroundColor Cyan
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

# 1. 获取图片库列表
Test-API -Name "获取图片库列表" `
    -Url "$apiUrl/galleries?page=1&page_size=10"

# 2. 获取单个图片库
Test-API -Name "获取单个图片库" `
    -Url "$apiUrl/galleries/1"

# 3. 获取图片库的所有图片
Test-API -Name "获取图片库的所有图片" `
    -Url "$apiUrl/galleries/1/images"

# 4. 搜索图片
Test-API -Name "搜索图片" `
    -Url "$apiUrl/galleries/images/search?q=product&page=1&page_size=10"

# 5. 根据标签获取图片
Test-API -Name "根据标签获取图片" `
    -Url "$apiUrl/galleries/images/tags?tags=product,premium&page=1&page_size=10"

# ========== 管理员 API 测试 ==========

Write-Host "`n=== 管理员 API 测试 ===" -ForegroundColor Cyan
Write-Host "注意: 这些测试需要认证，当前会返回 401" -ForegroundColor Yellow

# 6. 创建图片库（需要认证）
$newGallery = @{
    name = "Test Gallery"
    slug = "test-gallery-$(Get-Date -Format 'yyyyMMddHHmmss')"
    description = "This is a test gallery"
    status = "published"
}

Test-API -Name "创建图片库 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/admin/galleries" `
    -Body $newGallery `
    -ExpectedStatus 401

# 7. 更新图片库（需要认证）
$updateGallery = @{
    name = "Updated Gallery Name"
    description = "Updated description"
}

Test-API -Name "更新图片库 (需要认证)" `
    -Method "PUT" `
    -Url "$apiUrl/admin/galleries/1" `
    -Body $updateGallery `
    -ExpectedStatus 401

# 8. 创建图片（需要认证）
$newImage = @{
    url = "https://example.com/test-image.jpg"
    thumbnail = "https://example.com/test-image-thumb.jpg"
    title = "Test Image"
    description = "This is a test image"
    alt = "Test Image Alt"
    width = 1200
    height = 800
    size = 245760
    tags = "test,sample"
    order = 1
}

Test-API -Name "创建图片 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/admin/galleries/1/images" `
    -Body $newImage `
    -ExpectedStatus 401

# 9. 批量创建图片（需要认证）
$batchImages = @{
    images = @(
        @{
            url = "https://example.com/batch1.jpg"
            title = "Batch Image 1"
            order = 1
        },
        @{
            url = "https://example.com/batch2.jpg"
            title = "Batch Image 2"
            order = 2
        }
    )
}

Test-API -Name "批量创建图片 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/admin/galleries/1/images/batch" `
    -Body $batchImages `
    -ExpectedStatus 401

# 10. 更新图片（需要认证）
$updateImage = @{
    title = "Updated Image Title"
    description = "Updated description"
}

Test-API -Name "更新图片 (需要认证)" `
    -Method "PUT" `
    -Url "$apiUrl/admin/galleries/images/1" `
    -Body $updateImage `
    -ExpectedStatus 401

# 11. 批量更新排序（需要认证）
$batchOrder = @{
    orders = @{
        "1" = 10
        "2" = 20
        "3" = 30
    }
}

Test-API -Name "批量更新排序 (需要认证)" `
    -Method "POST" `
    -Url "$apiUrl/admin/galleries/images/batch-order" `
    -Body $batchOrder `
    -ExpectedStatus 401

# 12. 删除图片（需要认证）
Test-API -Name "删除图片 (需要认证)" `
    -Method "DELETE" `
    -Url "$apiUrl/admin/galleries/images/999" `
    -ExpectedStatus 401

# 13. 批量删除图片（需要认证）
$batchDelete = @{
    ids = @(998, 999)
}

Test-API -Name "批量删除图片 (需要认证)" `
    -Method "DELETE" `
    -Url "$apiUrl/admin/galleries/images/batch" `
    -Body $batchDelete `
    -ExpectedStatus 401

# 14. 删除图片库（需要认证）
Test-API -Name "删除图片库 (需要认证)" `
    -Method "DELETE" `
    -Url "$apiUrl/admin/galleries/999" `
    -ExpectedStatus 401

# ========== 边界测试 ==========

Write-Host "`n=== 边界测试 ===" -ForegroundColor Cyan

# 15. 获取不存在的图片库
Test-API -Name "获取不存在的图片库" `
    -Url "$apiUrl/galleries/99999" `
    -ExpectedStatus 404

# 16. 搜索空关键词
Test-API -Name "搜索空关键词" `
    -Url "$apiUrl/galleries/images/search?q=" `
    -ExpectedStatus 400

# 17. 标签参数为空
Test-API -Name "标签参数为空" `
    -Url "$apiUrl/galleries/images/tags?tags=" `
    -ExpectedStatus 400

# 18. 无效的图片库 ID
Test-API -Name "无效的图片库 ID" `
    -Url "$apiUrl/galleries/invalid" `
    -ExpectedStatus 400

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
