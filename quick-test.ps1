# 快速测试脚本 - 仅测试 Go 后端编译和基本功能

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "快速测试 - Go 后端编译检查" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Go
Write-Host "检查 Go 环境..." -ForegroundColor Yellow
$goVersion = go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "X Go 未安装" -ForegroundColor Red
    exit 1
}
Write-Host "OK Go 已安装: $goVersion" -ForegroundColor Green
Write-Host ""

# 进入 go-backend 目录
Set-Location "go-backend"

# 检查 go.mod
Write-Host "检查 Go 模块..." -ForegroundColor Yellow
if (Test-Path "go.mod") {
    Write-Host "OK go.mod 存在" -ForegroundColor Green
} else {
    Write-Host "X go.mod 不存在" -ForegroundColor Red
    Set-Location ..
    exit 1
}
Write-Host ""

# 下载依赖
Write-Host "下载 Go 依赖..." -ForegroundColor Yellow
Write-Host "提示: 首次运行可能需要几分钟..." -ForegroundColor Gray
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "OK 依赖下载完成" -ForegroundColor Green
} else {
    Write-Host "X 依赖下载失败" -ForegroundColor Red
    Set-Location ..
    exit 1
}
Write-Host ""

# 编译检查
Write-Host "编译检查..." -ForegroundColor Yellow
go build -o test-server.exe cmd/server/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "OK 编译成功" -ForegroundColor Green
    
    # 清理编译产物
    if (Test-Path "test-server.exe") {
        Remove-Item "test-server.exe"
    }
} else {
    Write-Host "X 编译失败" -ForegroundColor Red
    Set-Location ..
    exit 1
}
Write-Host ""

Set-Location ..

Write-Host "========================================" -ForegroundColor Green
Write-Host "编译检查通过！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

Write-Host "下一步:" -ForegroundColor Cyan
Write-Host "1. 启动数据库服务 (PostgreSQL + Redis)" -ForegroundColor White
Write-Host "   cd go-backend" -ForegroundColor Gray
Write-Host "   docker-compose up -d postgres redis" -ForegroundColor Gray
Write-Host ""
Write-Host "2. 启动 Go 后端" -ForegroundColor White
Write-Host "   cd go-backend" -ForegroundColor Gray
Write-Host "   go run cmd/server/main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "3. 测试 API" -ForegroundColor White
Write-Host "   访问: http://localhost:9000/health" -ForegroundColor Gray
Write-Host "   访问: http://localhost:9000/api/v1/i18n/languages" -ForegroundColor Gray
Write-Host ""
