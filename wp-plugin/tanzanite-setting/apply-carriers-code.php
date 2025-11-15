<?php
/**
 * 自动应用物流公司模块代码
 * 
 * 使用方法：
 * php apply-carriers-code.php
 */

$file = __DIR__ . '/tanzanite-setting.php';

if (!file_exists($file)) {
    die("错误：找不到文件 tanzanite-setting.php\n");
}

echo "开始应用物流公司模块代码...\n\n";

// 备份原文件
$backup = $file . '.backup.' . date('YmdHis');
copy($file, $backup);
echo "✓ 已创建备份：" . basename($backup) . "\n\n";

// 读取文件
$content = file_get_contents($file);

// ========================================
// 步骤 1: 添加 REST API 处理函数
// ========================================
echo "步骤 1: 添加 REST API 处理函数\n";
echo "----------------------------------------\n";

$rest_api_code = file_get_contents(__DIR__ . '/CARRIERS-REST-API-CODE.txt');

// 找到合适的插入位置（在 rest_list_shipping_templates 函数之后）
$search_pattern = '/public function rest_list_shipping_templates\([^}]+\{[^}]+\}/s';
if (preg_match($search_pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
    $insert_pos = $matches[0][1] + strlen($matches[0][0]);
    
    // 插入代码
    $content = substr_replace($content, "\n\n" . $rest_api_code, $insert_pos, 0);
    echo "✓ REST API 处理函数已添加\n\n";
} else {
    echo "⚠ 未找到插入位置，请手动添加 REST API 代码\n\n";
}

// ========================================
// 步骤 2: 替换 render_carriers() 函数
// ========================================
echo "步骤 2: 替换 render_carriers() 函数\n";
echo "----------------------------------------\n";

$admin_page_code = file_get_contents(__DIR__ . '/CARRIERS-ADMIN-PAGE-CODE.txt');

// 找到并替换 render_carriers 函数
$pattern = '/public function render_carriers\(\): void \{[^}]*\$this->render_placeholder_page\([^;]+\);[^}]*\}/s';
if (preg_match($pattern, $content)) {
    $content = preg_replace($pattern, $admin_page_code, $content);
    echo "✓ render_carriers() 函数已替换\n\n";
} else {
    echo "⚠ 未找到 render_carriers() 函数，请手动替换\n\n";
}

// ========================================
// 步骤 3: 添加发货推送 Hook
// ========================================
echo "步骤 3: 添加发货推送 Hook\n";
echo "----------------------------------------\n";

$hook_code = <<<'PHP'

    /**
     * 发货推送 Hook
     */
    public function trigger_order_shipped_hook( int $order_id, array $tracking_data ): void {
        /**
         * 订单发货时触发
         *
         * @param int   $order_id      订单 ID
         * @param array $tracking_data 物流数据
         */
        do_action( 'tanzanite_order_shipped', $order_id, $tracking_data );
    }

    /**
     * 物流状态更新 Hook
     */
    public function trigger_tracking_updated_hook( int $order_id, array $tracking_events ): void {
        /**
         * 物流状态更新时触发
         *
         * @param int   $order_id         订单 ID
         * @param array $tracking_events  物流事件列表
         */
        do_action( 'tanzanite_tracking_updated', $order_id, $tracking_events );
    }
PHP;

// 在类的末尾添加（在最后一个函数之后）
$pattern = '/(public function render_carriers\(\): void \{[^}]+\})\s*\}\s*\}/s';
if (preg_match($pattern, $content)) {
    $content = preg_replace($pattern, "$1\n" . $hook_code . "\n    }\n}", $content);
    echo "✓ 发货推送 Hook 已添加\n\n";
} else {
    echo "⚠ 未找到插入位置，请手动添加 Hook 代码\n\n";
}

// ========================================
// 保存文件
// ========================================
file_put_contents($file, $content);

echo "========================================\n";
echo "代码应用完成！\n";
echo "========================================\n";
echo "\n";
echo "备份文件：{$backup}\n";
echo "\n";
echo "下一步：\n";
echo "1. 刷新 WordPress 后台\n";
echo "2. 访问 Logistics → Carriers & Tracking\n";
echo "3. 测试物流公司管理功能\n";
echo "4. 测试 API 配置功能\n";
echo "\n";
echo "如果有问题，可以从备份恢复：\n";
echo "cp " . basename($backup) . " tanzanite-setting.php\n";
echo "\n";
?>
