#!/bin/bash

#############################################
# 日志分析脚本
# 用于分析应用程序日志，提取关键信息
#############################################

set -euo pipefail

# 默认配置
LOG_FILE="${LOG_FILE:-./logs/app.log}"
TIME_RANGE="${TIME_RANGE:-1h}"

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

info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

usage() {
    cat <<EOF
日志分析工具

用法: $0 [选项]

选项:
    -f <file>       日志文件路径（默认：./logs/app.log）
    -t <time>       时间范围（例如：1h, 24h, 7d）
    -e              只显示错误日志
    -w              只显示警告日志
    -s              显示统计摘要
    -h              显示帮助信息

示例:
    $0 -f app.log -t 24h -s
    $0 -e
    $0 -w -t 1h
EOF
    exit 0
}

# 检查日志文件是否存在
check_log_file() {
    if [ ! -f "${LOG_FILE}" ]; then
        error "日志文件不存在: ${LOG_FILE}"
        exit 1
    fi
}

# 显示统计摘要
show_summary() {
    log "日志统计摘要"
    echo ""
    
    # 总行数
    local total_lines
    total_lines=$(wc -l < "${LOG_FILE}")
    info "总日志条数: ${total_lines}"
    
    # 错误数量
    local error_count
    error_count=$(grep -c "\"level\":\"error\"" "${LOG_FILE}" || echo "0")
    error "错误数量: ${error_count}"
    
    # 警告数量
    local warn_count
    warn_count=$(grep -c "\"level\":\"warn\"" "${LOG_FILE}" || echo "0")
    echo -e "${YELLOW}[WARN]${NC} 警告数量: ${warn_count}"
    
    # Info数量
    local info_count
    info_count=$(grep -c "\"level\":\"info\"" "${LOG_FILE}" || echo "0")
    info "Info数量: ${info_count}"
    
    echo ""
    
    # 最频繁的错误
    log "最频繁的错误（Top 5）:"
    grep "\"level\":\"error\"" "${LOG_FILE}" | \
        grep -o "\"msg\":\"[^\"]*\"" | \
        sort | uniq -c | sort -rn | head -5 | \
        awk '{$1=$1; print "  " $0}'
    
    echo ""
    
    # 请求统计
    if grep -q "\"method\":" "${LOG_FILE}"; then
        log "API请求统计:"
        
        # 按HTTP方法统计
        info "按方法统计:"
        grep "\"method\":" "${LOG_FILE}" | \
            grep -o "\"method\":\"[^\"]*\"" | \
            sort | uniq -c | sort -rn | \
            awk '{print "  " $2 ": " $1}'
        
        echo ""
        
        # 按状态码统计
        info "按状态码统计:"
        grep "\"status\":" "${LOG_FILE}" | \
            grep -o "\"status\":[0-9]*" | \
            sort | uniq -c | sort -rn | \
            awk '{print "  " $2 ": " $1}'
    fi
    
    echo ""
}

# 显示错误日志
show_errors() {
    log "错误日志:"
    echo ""
    grep "\"level\":\"error\"" "${LOG_FILE}" | \
        jq -r '. | "\(.time) [\(.level | ascii_upcase)] \(.msg)"' 2>/dev/null || \
        grep "\"level\":\"error\"" "${LOG_FILE}"
}

# 显示警告日志
show_warnings() {
    log "警告日志:"
    echo ""
    grep "\"level\":\"warn\"" "${LOG_FILE}" | \
        jq -r '. | "\(.time) [\(.level | ascii_upcase)] \(.msg)"' 2>/dev/null || \
        grep "\"level\":\"warn\"" "${LOG_FILE}"
}

# 实时监控日志
tail_logs() {
    log "实时监控日志（Ctrl+C 退出）..."
    echo ""
    tail -f "${LOG_FILE}" | \
        jq -r '. | "\(.time) [\(.level | ascii_upcase)] \(.msg)"' 2>/dev/null || \
        tail -f "${LOG_FILE}"
}

# 主函数
main() {
    local show_summary_flag=false
    local show_errors_flag=false
    local show_warnings_flag=false
    
    # 解析参数
    while getopts "f:t:ewsh" opt; do
        case ${opt} in
            f)
                LOG_FILE="${OPTARG}"
                ;;
            t)
                TIME_RANGE="${OPTARG}"
                ;;
            e)
                show_errors_flag=true
                ;;
            w)
                show_warnings_flag=true
                ;;
            s)
                show_summary_flag=true
                ;;
            h)
                usage
                ;;
            \?)
                error "无效选项: -${OPTARG}"
                usage
                ;;
        esac
    done
    
    check_log_file
    
    echo ""
    echo "=========================================="
    echo "  日志分析"
    echo "=========================================="
    echo ""
    info "日志文件: ${LOG_FILE}"
    info "时间范围: ${TIME_RANGE}"
    echo ""
    
    # 执行相应的操作
    if ${show_summary_flag}; then
        show_summary
    fi
    
    if ${show_errors_flag}; then
        show_errors
    fi
    
    if ${show_warnings_flag}; then
        show_warnings
    fi
    
    # 如果没有指定任何标志，显示摘要
    if ! ${show_summary_flag} && ! ${show_errors_flag} && ! ${show_warnings_flag}; then
        show_summary
    fi
    
    echo ""
    info "分析完成"
    echo ""
}

main "$@"
