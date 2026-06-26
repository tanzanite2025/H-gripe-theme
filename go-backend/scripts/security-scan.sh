#!/bin/bash

#############################################
# 安全扫描脚本
# 用于检测代码和依赖中的安全漏洞
#############################################

set -euo pipefail

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

echo ""
echo "=========================================="
echo "  安全扫描工具"
echo "=========================================="
echo ""

# 检查工具是否安装
check_tool() {
    local tool=$1
    local install_info=$2
    
    if ! command -v "${tool}" &> /dev/null; then
        warn "${tool} 未安装"
        info "  ${install_info}"
        return 1
    fi
    return 0
}

SCAN_RESULTS=0

#############################################
# 1. gosec - Go代码安全检查
#############################################
log "1. 运行 gosec 安全检查..."

if check_tool "gosec" "go install github.com/securego/gosec/v2/cmd/gosec@latest"; then
    if gosec -fmt=json -out=gosec-report.json ./...; then
        log "✓ gosec 检查通过"
    else
        error "✗ gosec 发现安全问题"
        SCAN_RESULTS=$((SCAN_RESULTS + 1))
    fi
    
    # 显示简要结果
    if [ -f "gosec-report.json" ]; then
        log "详细报告: gosec-report.json"
    fi
else
    warn "跳过 gosec 检查"
fi

echo ""

#############################################
# 2. govulncheck - Go依赖漏洞检查
#############################################
log "2. 运行 govulncheck 依赖漏洞检查..."

if check_tool "govulncheck" "go install golang.org/x/vuln/cmd/govulncheck@latest"; then
    if govulncheck ./...; then
        log "✓ govulncheck 未发现已知漏洞"
    else
        error "✗ govulncheck 发现已知漏洞"
        SCAN_RESULTS=$((SCAN_RESULTS + 1))
    fi
else
    warn "跳过 govulncheck 检查"
fi

echo ""

#############################################
# 3. nancy - 依赖安全审计
#############################################
log "3. 运行 nancy 依赖安全审计..."

if check_tool "nancy" "go install github.com/sonatype-nexus-community/nancy@latest"; then
    if go list -json -m all | nancy sleuth; then
        log "✓ nancy 未发现问题"
    else
        error "✗ nancy 发现安全问题"
        SCAN_RESULTS=$((SCAN_RESULTS + 1))
    fi
else
    warn "跳过 nancy 检查"
fi

echo ""

#############################################
# 4. Trivy - 容器镜像扫描
#############################################
log "4. 运行 Trivy 容器镜像扫描..."

if check_tool "trivy" "https://aquasecurity.github.io/trivy/latest/getting-started/installation/"; then
    # 扫描Docker镜像
    if [ -f "Dockerfile" ]; then
        log "扫描 Dockerfile..."
        if trivy config Dockerfile; then
            log "✓ Dockerfile 检查通过"
        else
            warn "Dockerfile 存在配置问题"
        fi
    fi
    
    # 扫描文件系统
    log "扫描项目目录..."
    if trivy fs --security-checks vuln,config .; then
        log "✓ Trivy 扫描通过"
    else
        error "✗ Trivy 发现问题"
        SCAN_RESULTS=$((SCAN_RESULTS + 1))
    fi
else
    warn "跳过 Trivy 扫描"
fi

echo ""

#############################################
# 5. 检查敏感信息泄露
#############################################
log "5. 检查敏感信息泄露..."

# 检查.env文件是否被跟踪
if git ls-files --error-unmatch .env 2>/dev/null; then
    error "✗ .env 文件被Git跟踪，可能泄露敏感信息"
    SCAN_RESULTS=$((SCAN_RESULTS + 1))
else
    log "✓ .env 文件未被跟踪"
fi

# 检查常见的敏感信息模式
log "搜索可能的API密钥和密码..."
if grep -r -i -E "(password|secret|api_key|token)\s*=\s*['\"][^'\"]+['\"]" --exclude-dir=vendor --exclude-dir=.git --exclude="*.md" .; then
    warn "发现可能的硬编码密钥"
    info "  请检查上述文件，确保不包含真实密钥"
fi

echo ""

#############################################
# 6. 依赖许可证检查
#############################################
log "6. 检查依赖许可证..."

if check_tool "go-licenses" "go install github.com/google/go-licenses@latest"; then
    log "生成依赖许可证报告..."
    if go-licenses report ./... > licenses-report.txt 2>/dev/null; then
        log "✓ 许可证报告已生成: licenses-report.txt"
    else
        warn "许可证报告生成失败"
    fi
else
    warn "跳过许可证检查"
fi

echo ""

#############################################
# 7. 生成安全报告
#############################################
log "7. 生成安全扫描报告..."

cat > security-report.md <<EOF
# 安全扫描报告

**生成时间**: $(date)

## 扫描结果摘要

- **gosec**: 代码安全检查
- **govulncheck**: 依赖漏洞检查
- **nancy**: 依赖安全审计
- **Trivy**: 容器镜像扫描
- **敏感信息**: 泄露检查
- **许可证**: 依赖许可证检查

## 详细报告文件

- gosec结果: \`gosec-report.json\`
- 许可证报告: \`licenses-report.txt\`

## 建议

EOF

if [ ${SCAN_RESULTS} -eq 0 ]; then
    cat >> security-report.md <<EOF
✅ **所有安全检查通过！**

建议继续保持：
1. 定期更新依赖包
2. 定期运行安全扫描
3. 审查代码变更
4. 使用环境变量管理敏感信息
EOF
else
    cat >> security-report.md <<EOF
⚠️ **发现 ${SCAN_RESULTS} 个安全问题**

请采取以下行动：
1. 查看详细报告文件
2. 修复发现的安全问题
3. 更新存在漏洞的依赖
4. 重新运行安全扫描验证
EOF
fi

log "安全报告已生成: security-report.md"

echo ""
echo "=========================================="
if [ ${SCAN_RESULTS} -eq 0 ]; then
    log "✓ 安全扫描完成，未发现严重问题"
else
    error "✗ 安全扫描完成，发现 ${SCAN_RESULTS} 个问题"
fi
echo "=========================================="
echo ""

exit ${SCAN_RESULTS}
