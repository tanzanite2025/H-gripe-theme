# 博客多语言功能测试脚本
# 此脚本会启动必要的服务并测试 i18n 功能

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "博客多语言功能测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Docker 是否安装
Write-Host "检查 Docker 环境..." -ForegroundColor Yellow
$dockerVersion = docker --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker 未安装或未在 PATH 中" -ForegroundColor Red
    Write-Host "请先安装 Docker Desktop: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "或者，如果你已经有 PostgreSQL 和 Redis 运行在本地，请按任意键继续..." -ForegroundColor Yellow
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
} else {
    Write-Host "✓ Docker 已安装: $dockerVersion" -ForegroundColor Green
    
    # 启动 PostgreSQL 和 Redis
    Write-Host ""
    Write-Host "启动数据库服务 (PostgreSQL + Redis)..." -ForegroundColor Yellow
    Write-Host "提示: 首次运行可能需要下载镜像，请耐心等待" -ForegroundColor Gray
    
    Set-Location "go-backend"
    docker-compose up -d postgres redis
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 数据库服务启动成功" -ForegroundColor Green
        
        # 等待数据库就绪
        Write-Host ""
        Write-Host "等待数据库就绪..." -ForegroundColor Yellow
        $maxRetries = 30
        $retries = 0
        
        while ($retries -lt $maxRetries) {
            $pgReady = docker-compose exec -T postgres pg_isready -U tanzanite 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✓ PostgreSQL 已就绪" -ForegroundColor Green
                break
            }
            $retries++
            Write-Host "  等待中... ($retries/$maxRetries)" -ForegroundColor Gray
            Start-Sleep -Seconds 2
        }
        
        if ($retries -eq $maxRetries) {
            Write-Host "❌ PostgreSQL 启动超时" -ForegroundColor Red
            exit 1
        }
        
        # 检查 Redis
        $redisReady = docker-compose exec -T redis redis-cli ping 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Redis 已就绪" -ForegroundColor Green
        }
    } else {
        Write-Host "❌ 数据库服务启动失败" -ForegroundColor Red
        Write-Host "请检查 Docker 是否正常运行" -ForegroundColor Yellow
        Set-Location ..
        exit 1
    }
    
    Set-Location ..
}

# 检查 Go 是否安装
Write-Host ""
Write-Host "检查 Go 环境..." -ForegroundColor Yellow
$goVersion = go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Go 未安装或未在 PATH 中" -ForegroundColor Red
    Write-Host "请先安装 Go: https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}
Write-Host "✓ Go 已安装: $goVersion" -ForegroundColor Green

# 检查配置文件
Write-Host ""
Write-Host "检查配置文件..." -ForegroundColor Yellow
if (Test-Path "go-backend\config\config.yaml") {
    Write-Host "✓ 配置文件存在" -ForegroundColor Green
} else {
    Write-Host "❌ 配置文件不存在" -ForegroundColor Red
    Write-Host "请确保 go-backend/config/config.yaml 存在" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "启动 Go 后端服务器" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 启动 Go 后端（在后台）
Write-Host "启动 Go 后端 (http://localhost:9000)..." -ForegroundColor Yellow
$goProcess = Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD\go-backend'; Write-Host 'Go 后端服务器' -ForegroundColor Cyan; go run cmd/server/main.go" -PassThru

# 等待服务器启动
Write-Host "等待服务器启动..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# 检查服务器是否启动
$maxRetries = 10
$retries = 0
$serverReady = $false

while ($retries -lt $maxRetries) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:9000/health" -Method GET -TimeoutSec 2 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            Write-Host "✓ Go 后端服务器已启动" -ForegroundColor Green
            $serverReady = $true
            break
        }
    } catch {
        # 忽略错误，继续重试
    }
    $retries++
    Write-Host "  等待中... ($retries/$maxRetries)" -ForegroundColor Gray
    Start-Sleep -Seconds 2
}

if (-not $serverReady) {
    Write-Host "❌ Go 后端服务器启动失败或超时" -ForegroundColor Red
    Write-Host "请检查 PowerShell 窗口中的错误信息" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "开始测试 i18n API" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# 测试 1: 健康检查
Write-Host "测试 1: 健康检查" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "http://localhost:9000/health" -Method GET
    Write-Host "✓ 健康检查通过" -ForegroundColor Green
    Write-Host "  响应: $($response | ConvertTo-Json -Compress)" -ForegroundColor Gray
} catch {
    Write-Host "❌ 健康检查失败: $_" -ForegroundColor Red
}

Write-Host ""

# 测试 2: 获取语言列表
Write-Host "测试 2: 获取语言列表" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "http://localhost:9000/api/v1/i18n/languages" -Method GET
    $languages = $response.languages
    Write-Host "✓ 获取语言列表成功" -ForegroundColor Green
    Write-Host "  支持的语言数量: $($languages.Count)" -ForegroundColor Gray
    Write-Host "  前 5 种语言: $($languages[0..4].code -join ', ')" -ForegroundColor Gray
} catch {
    Write-Host "❌ 获取语言列表失败: $_" -ForegroundColor Red
}

Write-Host ""

# 测试 3: 语言检测
Write-Host "测试 3: 语言检测" -ForegroundColor Cyan
try {
    $headers = @{
        "Accept-Language" = "zh-CN,zh,en"
    }
    $response = Invoke-RestMethod -Uri "http://localhost:9000/api/v1/i18n/detect" -Method GET -Headers $headers
    Write-Host "✓ 语言检测成功" -ForegroundColor Green
    Write-Host "  检测到的语言: $($response.detected_locale)" -ForegroundColor Gray
} catch {
    Write-Host "❌ 语言检测失败: $_" -ForegroundColor Red
}

Write-Host ""

# 测试 4: 设置语言
Write-Host "测试 4: 设置语言" -ForegroundColor Cyan
try {
    $body = @{
        locale = "fr"
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "http://localhost:9000/api/v1/i18n/set-language" -Method POST -Body $body -ContentType "application/json"
    Write-Host "✓ 设置语言成功" -ForegroundColor Green
    Write-Host "  设置的语言: $($response.locale)" -ForegroundColor Gray
} catch {
    Write-Host "❌ 设置语言失败: $_" -ForegroundColor Red
}

Write-Host ""

# 测试 5: Sitemap 索引
Write-Host "测试 5: Sitemap 索引" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:9000/sitemap.xml" -Method GET
    if ($response.StatusCode -eq 200) {
        Write-Host "✓ Sitemap 索引生成成功" -ForegroundColor Green
        Write-Host "  内容长度: $($response.Content.Length) 字节" -ForegroundColor Gray
        
        # 检查是否包含 hreflang sitemap
        if ($response.Content -match "sitemap-hreflang.xml") {
            Write-Host "  ✓ 包含 Hreflang Sitemap" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "❌ Sitemap 索引生成失败: $_" -ForegroundColor Red
}

Write-Host ""

# 测试 6: Hreflang Sitemap
Write-Host "测试 6: Hreflang Sitemap" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:9000/sitemap-hreflang.xml" -Method GET
    if ($response.StatusCode -eq 200) {
        Write-Host "✓ Hreflang Sitemap 生成成功" -ForegroundColor Green
        Write-Host "  内容长度: $($response.Content.Length) 字节" -ForegroundColor Gray
        
        # 检查是否包含 hreflang 标签
        if ($response.Content -match 'hreflang="') {
            Write-Host "  ✓ 包含 Hreflang 标签" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "❌ Hreflang Sitemap 生成失败: $_" -ForegroundColor Red
}

Write-Host ""

# 测试 7: 单语言 Sitemap
Write-Host "测试 7: 单语言 Sitemap (en)" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:9000/sitemap-en.xml" -Method GET
    if ($response.StatusCode -eq 200) {
        Write-Host "✓ 单语言 Sitemap 生成成功" -ForegroundColor Green
        Write-Host "  内容长度: $($response.Content.Length) 字节" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ 单语言 Sitemap 生成失败: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "测试完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

Write-Host "服务访问地址:" -ForegroundColor Cyan
Write-Host "  - 后端 API: http://localhost:9000" -ForegroundColor White
Write-Host "  - 健康检查: http://localhost:9000/health" -ForegroundColor White
Write-Host "  - 语言列表: http://localhost:9000/api/v1/i18n/languages" -ForegroundColor White
Write-Host "  - Sitemap: http://localhost:9000/sitemap.xml" -ForegroundColor White
Write-Host ""

Write-Host "数据库管理工具 (可选):" -ForegroundColor Cyan
Write-Host "  启动 Adminer: docker-compose --profile dev up -d adminer" -ForegroundColor Gray
Write-Host "  访问地址: http://localhost:8081" -ForegroundColor Gray
Write-Host ""

Write-Host "停止服务:" -ForegroundColor Yellow
Write-Host "  1. 关闭 Go 后端的 PowerShell 窗口" -ForegroundColor Gray
Write-Host "  2. 停止数据库: cd go-backend; docker-compose down" -ForegroundColor Gray
Write-Host ""

Write-Host "提示: Go 后端服务器正在另一个 PowerShell 窗口中运行" -ForegroundColor Yellow
Write-Host "按任意键退出此脚本（不会停止服务器）..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
