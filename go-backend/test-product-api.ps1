# 商品管理 API 测试脚本
# 使用方法: .\test-product-api.ps1

$baseUrl = "http://localhost:8080"
$adminEmail = "admin@tanzanite.com"
$adminPassword = "admin123"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "商品管理 API 测试" -ForegroundColor Cyan
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

# 2. 获取商品统计
Write-Host "2. 获取商品统计..." -ForegroundColor Yellow
try {
    $stats = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/stats" -Method Get -Headers $headers
    Write-Host "✓ 统计数据获取成功" -ForegroundColor Green
    Write-Host "  总商品数: $($stats.total)" -ForegroundColor Gray
    Write-Host "  在售商品: $($stats.active)" -ForegroundColor Gray
    Write-Host "  下架商品: $($stats.inactive)" -ForegroundColor Gray
    Write-Host "  缺货商品: $($stats.out_of_stock)" -ForegroundColor Gray
    Write-Host "  精选商品: $($stats.featured)" -ForegroundColor Gray
    Write-Host "  低库存商品: $($stats.low_stock)" -ForegroundColor Gray
} catch {
    Write-Host "✗ 获取统计失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 3. 创建测试商品
Write-Host "3. 创建测试商品..." -ForegroundColor Yellow
$newProduct = @{
    sku = "TEST-PRODUCT-$(Get-Random -Maximum 9999)"
    name = "测试商品 - 坦桑石项链"
    slug = "test-tanzanite-necklace-$(Get-Random -Maximum 9999)"
    description = "这是一条精美的坦桑石项链，采用优质坦桑石宝石制作。"
    short_description = "精美坦桑石项链"
    price = 1299.99
    sale_price = 999.99
    stock = 50
    weight_grams = 25
    status = "active"
    locale = "zh"
    featured = $true
    meta_title = "坦桑石项链 - 精美珠宝"
    meta_description = "购买精美的坦桑石项链，优质宝石，精湛工艺。"
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/products" -Method Post -Body $newProduct -Headers $headers
    $productId = $createResponse.product.id
    Write-Host "✓ 商品创建成功" -ForegroundColor Green
    Write-Host "  商品ID: $productId" -ForegroundColor Gray
    Write-Host "  SKU: $($createResponse.product.sku)" -ForegroundColor Gray
    Write-Host "  名称: $($createResponse.product.name)" -ForegroundColor Gray
    Write-Host "  价格: ¥$($createResponse.product.price)" -ForegroundColor Gray
    Write-Host "  促销价: ¥$($createResponse.product.sale_price)" -ForegroundColor Gray
} catch {
    Write-Host "✗ 创建商品失败: $($_.Exception.Message)" -ForegroundColor Red
    $productId = $null
}

Write-Host ""

if ($productId) {
    # 4. 获取商品详情
    Write-Host "4. 获取商品详情..." -ForegroundColor Yellow
    try {
        $product = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/$productId" -Method Get -Headers $headers
        Write-Host "✓ 商品详情获取成功" -ForegroundColor Green
        Write-Host "  ID: $($product.product.id)" -ForegroundColor Gray
        Write-Host "  SKU: $($product.product.sku)" -ForegroundColor Gray
        Write-Host "  名称: $($product.product.name)" -ForegroundColor Gray
        Write-Host "  状态: $($product.product.status)" -ForegroundColor Gray
        Write-Host "  库存: $($product.product.stock)" -ForegroundColor Gray
        Write-Host "  精选: $($product.product.featured)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 获取商品详情失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""

    # 5. 更新商品信息
    Write-Host "5. 更新商品信息..." -ForegroundColor Yellow
    $updateProduct = @{
        name = "测试商品 - 坦桑石项链（已更新）"
        price = 1399.99
        sale_price = 1099.99
        description = "这是一条精美的坦桑石项链，采用优质坦桑石宝石制作。已更新描述。"
    } | ConvertTo-Json

    try {
        $updateResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/$productId" -Method Put -Body $updateProduct -Headers $headers
        Write-Host "✓ 商品更新成功" -ForegroundColor Green
        Write-Host "  新名称: $($updateResponse.product.name)" -ForegroundColor Gray
        Write-Host "  新价格: ¥$($updateResponse.product.price)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 更新商品失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""

    # 6. 更新商品库存
    Write-Host "6. 更新商品库存..." -ForegroundColor Yellow
    $updateStock = @{
        stock = 100
    } | ConvertTo-Json

    try {
        $stockResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/$productId/stock" -Method Patch -Body $updateStock -Headers $headers
        Write-Host "✓ 库存更新成功" -ForegroundColor Green
        Write-Host "  新库存: 100" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 更新库存失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""

    # 7. 更新商品状态
    Write-Host "7. 更新商品状态（下架）..." -ForegroundColor Yellow
    $updateStatus = @{
        status = "inactive"
    } | ConvertTo-Json

    try {
        $statusResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/$productId/status" -Method Patch -Body $updateStatus -Headers $headers
        Write-Host "✓ 状态更新成功（已下架）" -ForegroundColor Green
    } catch {
        Write-Host "✗ 更新状态失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""

    # 8. 再次上架
    Write-Host "8. 更新商品状态（上架）..." -ForegroundColor Yellow
    $updateStatus = @{
        status = "active"
    } | ConvertTo-Json

    try {
        $statusResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/$productId/status" -Method Patch -Body $updateStatus -Headers $headers
        Write-Host "✓ 状态更新成功（已上架）" -ForegroundColor Green
    } catch {
        Write-Host "✗ 更新状态失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""
}

# 9. 获取商品列表（带筛选）
Write-Host "9. 获取商品列表..." -ForegroundColor Yellow
try {
    $products = Invoke-RestMethod -Uri "$baseUrl/api/admin/products?page=1&page_size=10&status=active" -Method Get -Headers $headers
    Write-Host "✓ 商品列表获取成功" -ForegroundColor Green
    Write-Host "  总数: $($products.pagination.total)" -ForegroundColor Gray
    Write-Host "  当前页: $($products.pagination.page)" -ForegroundColor Gray
    Write-Host "  每页数量: $($products.pagination.page_size)" -ForegroundColor Gray
    Write-Host "  商品数量: $($products.products.Count)" -ForegroundColor Gray
    
    if ($products.products.Count -gt 0) {
        Write-Host "  前3个商品:" -ForegroundColor Gray
        $products.products | Select-Object -First 3 | ForEach-Object {
            Write-Host "    - [$($_.id)] $($_.name) (SKU: $($_.sku), 库存: $($_.stock))" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "✗ 获取商品列表失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 10. 搜索商品
Write-Host "10. 搜索商品（关键词：坦桑石）..." -ForegroundColor Yellow
try {
    $searchProducts = Invoke-RestMethod -Uri "$baseUrl/api/admin/products?search=坦桑石&page=1&page_size=5" -Method Get -Headers $headers
    Write-Host "✓ 搜索成功" -ForegroundColor Green
    Write-Host "  找到 $($searchProducts.pagination.total) 个商品" -ForegroundColor Gray
    
    if ($searchProducts.products.Count -gt 0) {
        $searchProducts.products | ForEach-Object {
            Write-Host "    - $($_.name) (SKU: $($_.sku))" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "✗ 搜索失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 11. 筛选精选商品
Write-Host "11. 筛选精选商品..." -ForegroundColor Yellow
try {
    $featuredProducts = Invoke-RestMethod -Uri "$baseUrl/api/admin/products?featured=true&page=1&page_size=5" -Method Get -Headers $headers
    Write-Host "✓ 筛选成功" -ForegroundColor Green
    Write-Host "  找到 $($featuredProducts.pagination.total) 个精选商品" -ForegroundColor Gray
} catch {
    Write-Host "✗ 筛选失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

if ($productId) {
    # 12. 删除测试商品
    Write-Host "12. 删除测试商品..." -ForegroundColor Yellow
    try {
        $deleteResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/$productId" -Method Delete -Headers $headers
        Write-Host "✓ 商品删除成功" -ForegroundColor Green
    } catch {
        Write-Host "✗ 删除商品失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host ""
}

# 13. 再次获取统计（验证删除）
Write-Host "13. 再次获取商品统计..." -ForegroundColor Yellow
try {
    $finalStats = Invoke-RestMethod -Uri "$baseUrl/api/admin/products/stats" -Method Get -Headers $headers
    Write-Host "✓ 统计数据获取成功" -ForegroundColor Green
    Write-Host "  总商品数: $($finalStats.total)" -ForegroundColor Gray
    Write-Host "  在售商品: $($finalStats.active)" -ForegroundColor Gray
} catch {
    Write-Host "✗ 获取统计失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "提示：" -ForegroundColor Yellow
Write-Host "1. 确保服务器正在运行: go run cmd/server/main.go" -ForegroundColor Gray
Write-Host "2. 确保数据库已初始化并有管理员账户" -ForegroundColor Gray
Write-Host "3. 可以访问前端管理界面测试: http://localhost:8080/admin" -ForegroundColor Gray
