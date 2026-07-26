#!/usr/bin/env bash

#############################################
# Tanzanite security scan
#
# The release image must be supplied explicitly:
#   REQUIRE_IMAGE_SCAN=true IMAGE_REF=ghcr.io/example/tanzanite-theme-api:sha-... ./scripts/security-scan.sh
#
# Critical/high image findings and confirmed historical secrets fail the scan.
#############################################

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_DIR="${REPORT_DIR:-${ROOT_DIR}/security-reports}"
IMAGE_REF="${IMAGE_REF:-}"
REQUIRE_IMAGE_SCAN="${REQUIRE_IMAGE_SCAN:-false}"
GITLEAKS_LOG_OPTS="${GITLEAKS_LOG_OPTS:---all --tags}"
SCAN_RESULTS=0

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

check_tool() {
    local tool="$1"
    local install_info="$2"

    if ! command -v "${tool}" >/dev/null 2>&1; then
        warn "${tool} 未安装"
        info "  ${install_info}"
        return 1
    fi
    return 0
}

record_failure() {
    SCAN_RESULTS=$((SCAN_RESULTS + 1))
}

mkdir -p "${REPORT_DIR}"

echo ""
echo "=========================================="
echo "  Tanzanite 安全扫描"
echo "=========================================="
echo ""

#############################################
# 1. Go source and dependency checks
#############################################
cd "${ROOT_DIR}"

log "1. 运行 gosec..."
if check_tool "gosec" "go install github.com/securego/gosec/v2/cmd/gosec@latest"; then
    if gosec -fmt=json -out="${REPORT_DIR}/gosec-report.json" ./...; then
        log "✓ gosec 检查通过"
    else
        error "✗ gosec 发现安全问题"
        record_failure
    fi
else
    warn "跳过 gosec"
fi

log "运行 govulncheck..."
if check_tool "govulncheck" "go install golang.org/x/vuln/cmd/govulncheck@latest"; then
    if govulncheck ./... 2>&1 | tee "${REPORT_DIR}/govulncheck.txt"; then
        log "✓ govulncheck 检查通过"
    else
        error "✗ govulncheck 发现已知漏洞"
        record_failure
    fi
else
    warn "跳过 govulncheck"
fi

log "运行 nancy..."
if check_tool "nancy" "go install github.com/sonatype-nexus-community/nancy@latest"; then
    if go list -json -m all | nancy sleuth 2>&1 | tee "${REPORT_DIR}/nancy.txt"; then
        log "✓ nancy 检查通过"
    else
        error "✗ nancy 发现依赖问题"
        record_failure
    fi
else
    warn "跳过 nancy"
fi

#############################################
# 2. Final image scanning
#############################################
log "2. 扫描最终运行镜像..."
if [[ -z "${IMAGE_REF}" ]]; then
    warn "未设置 IMAGE_REF，跳过 Trivy/Grype 镜像扫描"
    info "发布前必须使用精确 tag 或 digest 设置 IMAGE_REF"
    if [[ "${REQUIRE_IMAGE_SCAN}" == "true" ]]; then
        error "REQUIRE_IMAGE_SCAN=true 但未提供 IMAGE_REF"
        record_failure
    fi
else
    info "镜像: ${IMAGE_REF}"

    if check_tool "trivy" "https://aquasecurity.github.io/trivy/latest/getting-started/installation/"; then
        if trivy image \
            --scanners vuln \
            --severity HIGH,CRITICAL \
            --exit-code 1 \
            --format json \
            --output "${REPORT_DIR}/trivy-image.json" \
            "${IMAGE_REF}"; then
            log "✓ Trivy 镜像扫描通过"
        else
            error "✗ Trivy 发现 High/Critical 镜像漏洞"
            record_failure
        fi
    else
        warn "跳过 Trivy 镜像扫描"
        record_failure
    fi

    if check_tool "grype" "https://github.com/anchore/grype#installation"; then
        if grype "${IMAGE_REF}" \
            --fail-on high \
            --output json > "${REPORT_DIR}/grype-image.json"; then
            log "✓ Grype 镜像扫描通过"
        else
            error "✗ Grype 发现 High/Critical 镜像漏洞"
            record_failure
        fi
    else
        warn "跳过 Grype 镜像扫描"
        record_failure
    fi
fi

#############################################
# 3. Full Git history secrets scan
#############################################
log "3. 运行 gitleaks 全历史 secrets 扫描..."
if check_tool "gitleaks" "https://github.com/gitleaks/gitleaks#installation"; then
    if gitleaks git \
        --source "${ROOT_DIR}" \
        --log-opts="${GITLEAKS_LOG_OPTS}" \
        --redact \
        --report-format sarif \
        --report-path "${REPORT_DIR}/gitleaks-history.sarif"; then
        log "✓ gitleaks 历史扫描通过"
    else
        error "✗ gitleaks 发现历史 secrets"
        record_failure
    fi
else
    warn "跳过 gitleaks 历史扫描"
    record_failure
fi

#############################################
# 4. Repository hygiene
#############################################
log "4. 检查未跟踪的环境文件..."
if git -C "${ROOT_DIR}" ls-files --error-unmatch .env >/dev/null 2>&1; then
    error "✗ .env 文件被 Git 跟踪"
    record_failure
else
    log "✓ .env 文件未被 Git 跟踪"
fi

cat > "${REPORT_DIR}/security-report.md" <<EOF
# Tanzanite Security Scan

Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")

Image: \`${IMAGE_REF:-not supplied}\`

Scans:

- gosec
- govulncheck
- nancy
- Trivy final image scan
- Grype final image scan
- gitleaks full Git history scan

High/Critical image findings and confirmed historical secrets are release blockers.
EOF

echo ""
echo "=========================================="
if [[ ${SCAN_RESULTS} -eq 0 ]]; then
    log "✓ 安全扫描完成"
else
    error "✗ 安全扫描完成，发现 ${SCAN_RESULTS} 个阻断项"
fi
echo "报告目录: ${REPORT_DIR}"
echo "=========================================="
echo ""

exit "${SCAN_RESULTS}"
