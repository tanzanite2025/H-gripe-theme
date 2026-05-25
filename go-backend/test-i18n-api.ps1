# 博客多语言功能 API 测试脚本 (PowerShell)
# 使用方法: .\test-i18n-api.ps1 [-BaseUrl "http://localhost:9000"]

param(
    [string]$BaseUrl = "http://localhost:9000"
)

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "博客多语言功能 API 测试" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# 测试函数
function Test-API {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Endpoint,
        [string]$Body = $null
    )
    
    Write-Host "测试: $Name" -ForegroundColor Yellow
    Write-Host "请求: $Method $Endpoint"
    
    try {
        $url = "$BaseUrl$Endpoint"
        
        if ($Body) {
            $response = Invoke-RestMethod -Uri $url -Method $Method -Body $Body -ContentType "application/json" -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $url -Method $Method -ErrorAction Stop
        }
        
        Write-Host "✓ 成功" -ForegroundColor Green
        $response | ConvertTo-Json -Depth 10
    }
    catch {
        Write-Host "✗ 失败" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "------------------------------------------"
    Write-Host ""
}

# 1. 测试健康检查
Test-API -Name "健康检查" -Method "GET" -Endpoint "/health"

# 2. 测试获取语言列表
Test-API -Name "获取支持的语言列表" -Method "GET" -Endpoint "/api/v1/i18n/languages"

# 3. 测试语言检测
Test-API -Name "检测用户语言偏好" -Method "GET" -Endpoint "/api/v1/i18n/detect"

# 4. 测试设置语言
Test-API -Name "设置用户语言偏好为中文" -Method "POST" -Endpoint "/api/v1/i18n/set-language" -Body '{"locale":"zh"}'

# 5. 测试获取文章翻译（假设文章ID为1）
Test-API -Name "获取文章翻译版本 (ID=1)" -Method "GET" -Endpoint "/api/v1/i18n/translations/1"

# 6. 测试 Sitemap 索引
Write-Host "测试: Sitemap 索引" -ForegroundColor Yellow
Write-Host "请求: GET /sitemap.xml"
try {
    $sitemap = Invoke-WebRequest -Uri "$BaseUrl/sitemap.xml" -ErrorAction Stop
    $outputPath = "$env:TEMP\sitemap-index.xml"
    $sitemap.Content | Out-File -FilePath $outputPath -Encoding UTF8
    Write-Host "✓ 成功" -ForegroundColor Green
    Write-Host "文件已保存到: $outputPath"
    Write-Host ($sitemap.Content -split "`n" | Select-Object -First 20 | Out-String)
}
catch {
    Write-Host "✗ 失败" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}
Write-Host ""
Write-Host "------------------------------------------"
Write-Host ""

# 7. 测试 Hreflang Sitemap
Write-Host "测试: Hreflang Sitemap" -ForegroundColor Yellow
Write-Host "请求: GET /sitemap-hreflang.xml"
try {
    $sitemap = Invoke-WebRequest -Uri "$BaseUrl/sitemap-hreflang.xml" -ErrorAction Stop
    $outputPath = "$env:TEMP\sitemap-hreflang.xml"
    $sitemap.Content | Out-File -FilePath $outputPath -Encoding UTF8
    Write-Host "✓ 成功" -ForegroundColor Green
    Write-Host "文件已保存到: $outputPath"
    Write-Host ($sitemap.Content -split "`n" | Select-Object -First 30 | Out-String)
}
catch {
    Write-Host "✗ 失败" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}
Write-Host ""
Write-Host "------------------------------------------"
Write-Host ""

# 8. 测试单语言 Sitemap
Write-Host "测试: 英文 Sitemap" -ForegroundColor Yellow
Write-Host "请求: GET /sitemap-en.xml"
try {
    $sitemap = Invoke-WebRequest -Uri "$BaseUrl/sitemap-en.xml" -ErrorAction Stop
    $outputPath = "$env:TEMP\sitemap-en.xml"
    $sitemap.Content | Out-File -FilePath $outputPath -Encoding UTF8
    Write-Host "✓ 成功" -ForegroundColor Green
    Write-Host "文件已保存到: $outputPath"
    Write-Host ($sitemap.Content -split "`n" | Select-Object -First 20 | Out-String)
}
catch {
    Write-Host "✗ 失败" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}
Write-Host ""
Write-Host "------------------------------------------"
Write-Host ""

# 9. 测试中文 Sitemap
Write-Host "测试: 中文 Sitemap" -ForegroundColor Yellow
Write-Host "请求: GET /sitemap-zh.xml"
try {
    $sitemap = Invoke-WebRequest -Uri "$BaseUrl/sitemap-zh.xml" -ErrorAction Stop
    $outputPath = "$env:TEMP\sitemap-zh.xml"
    $sitemap.Content | Out-File -FilePath $outputPath -Encoding UTF8
    Write-Host "✓ 成功" -ForegroundColor Green
    Write-Host "文件已保存到: $outputPath"
    Write-Host ($sitemap.Content -split "`n" | Select-Object -First 20 | Out-String)
}
catch {
    Write-Host "✗ 失败" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}
Write-Host ""
Write-Host "------------------------------------------"
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "生成的文件位于: $env:TEMP" -ForegroundColor Green
Write-Host "  - sitemap-index.xml"
Write-Host "  - sitemap-hreflang.xml"
Write-Host "  - sitemap-en.xml"
Write-Host "  - sitemap-zh.xml"
Write-Host ""
Write-Host "提示: 使用 'Get-Content `$env:TEMP\sitemap-*.xml' 查看完整内容" -ForegroundColor Yellow
