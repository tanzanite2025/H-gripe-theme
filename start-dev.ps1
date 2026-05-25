# 开发环境快速启动脚本
# 同时启动 Go 后端和 Nuxt 前端

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Tanzanite 开发环境启动" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Go 是否安装
Write-Host "检查 Go 环境..." -ForegroundColor Yellow
$goVersion = go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Go 未安装或未在 PATH 中" -ForegroundColor Red
    Write-Host "请先安装 Go: https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}
Write-Host "✓ Go 已安装: $goVersion" -ForegroundColor Green

# 检查 Node.js 是否安装
Write-Host "检查 Node.js 环境..." -ForegroundColor Yellow
$nodeVersion = node --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Node.js 未安装或未在 PATH 中" -ForegroundColor Red
    Write-Host "请先安装 Node.js: https://nodejs.org/" -ForegroundColor Yellow
    exit 1
}
Write-Host "✓ Node.js 已安装: $nodeVersion" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "启动服务..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 启动 Go 后端
Write-Host "启动 Go 后端 (http://localhost:9000)..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd go-backend; Write-Host 'Go 后端服务器' -ForegroundColor Cyan; go run cmd/server/main.go"

# 等待 2 秒
Start-Sleep -Seconds 2

# 启动 Nuxt 前端
Write-Host "启动 Nuxt 前端 (http://localhost:3000)..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd nuxt-i18n; Write-Host 'Nuxt 前端服务器' -ForegroundColor Cyan; npm run dev"

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "服务启动完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "访问地址:" -ForegroundColor Cyan
Write-Host "  - 前端: http://localhost:3000" -ForegroundColor White
Write-Host "  - 后端 API: http://localhost:9000" -ForegroundColor White
Write-Host "  - 健康检查: http://localhost:9000/health" -ForegroundColor White
Write-Host ""
Write-Host "API 端点:" -ForegroundColor Cyan
Write-Host "  - 语言列表: http://localhost:9000/api/v1/i18n/languages" -ForegroundColor White
Write-Host "  - Sitemap: http://localhost:9000/sitemap.xml" -ForegroundColor White
Write-Host ""
Write-Host "提示: 关闭此窗口不会停止服务，请手动关闭打开的 PowerShell 窗口" -ForegroundColor Yellow
