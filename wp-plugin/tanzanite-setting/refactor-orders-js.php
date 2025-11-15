<?php
/**
 * 自动重构订单模块 JS
 * 
 * 这个脚本会自动删除 render_orders_list() 和 render_order_detail_page() 中的内联 JS
 * 
 * 使用方法：
 * php refactor-orders-js.php
 */

$file = __DIR__ . '/tanzanite-setting.php';

if (!file_exists($file)) {
    die("错误：找不到文件 tanzanite-setting.php\n");
}

echo "开始重构订单模块 JS...\n\n";

// 备份原文件
$backup = $file . '.backup.' . date('YmdHis');
copy($file, $backup);
echo "✓ 已创建备份：" . basename($backup) . "\n";

// 读取文件
$content = file_get_contents($file);
$lines = explode("\n", $content);
$total_lines = count($lines);

echo "✓ 文件总行数：{$total_lines}\n\n";

// ========================================
// 修改 1: render_orders_list()
// ========================================
echo "修改 1: render_orders_list()\n";
echo "----------------------------------------\n";

// 查找起始标记
$start_marker = '            echo \'  </div\';';
$end_marker = '        public function render_reviews(): void {';

$start_line = -1;
$end_line = -1;

// 找到 render_orders_list 中最后的 echo '</div>';
for ($i = 12100; $i < 12150; $i++) {
    if (isset($lines[$i]) && trim($lines[$i]) === "echo '</div>';") {
        // 检查下一行是否为空行
        if (isset($lines[$i+1]) && trim($lines[$i+1]) === '') {
            $start_line = $i + 1; // 从空行开始删除
            break;
        }
    }
}

// 找到 render_reviews 函数
for ($i = 12400; $i < 12500; $i++) {
    if (isset($lines[$i]) && strpos($lines[$i], 'public function render_reviews') !== false) {
        $end_line = $i - 1; // 删除到 render_reviews 之前
        // 往回找，跳过空行和右花括号
        while ($end_line > $start_line && (trim($lines[$end_line]) === '' || trim($lines[$end_line]) === '}')) {
            $end_line--;
        }
        $end_line++; // 包含最后一个非空行
        break;
    }
}

if ($start_line === -1 || $end_line === -1) {
    die("错误：无法找到需要删除的代码范围\n");
}

$delete_count_1 = $end_line - $start_line + 1;
echo "  起始行：{$start_line}\n";
echo "  结束行：{$end_line}\n";
echo "  删除行数：{$delete_count_1}\n";

// 删除这些行
array_splice($lines, $start_line, $delete_count_1);

// 在删除位置添加函数结束
$lines[$start_line] = '        }';
$lines[$start_line + 1] = '';

echo "✓ 已删除 render_orders_list() 的内联 JS\n\n";

// ========================================
// 修改 2: render_order_detail_page()  
// ========================================
echo "修改 2: render_order_detail_page()\n";
echo "----------------------------------------\n";

// 重新计算行号（因为已经删除了一部分）
$lines = array_values($lines); // 重新索引

// 查找 render_order_detail_page 中的内联 JS
$start_line_2 = -1;
$end_line_2 = -1;

// 查找 $status_labels 变量定义
for ($i = 1500; $i < 1700; $i++) {
    if (isset($lines[$i]) && strpos($lines[$i], '$status_labels = [') !== false) {
        $start_line_2 = $i;
        break;
    }
}

// 查找 echo '<script>' 
for ($i = $start_line_2; $i < $start_line_2 + 200; $i++) {
    if (isset($lines[$i]) && strpos($lines[$i], "echo '<script>'") !== false) {
        $end_line_2 = $i;
        break;
    }
}

if ($start_line_2 !== -1 && $end_line_2 !== -1) {
    $delete_count_2 = $end_line_2 - $start_line_2 + 1;
    echo "  起始行：{$start_line_2}\n";
    echo "  结束行：{$end_line_2}\n";
    echo "  删除行数：{$delete_count_2}\n";
    
    // 删除这些行
    array_splice($lines, $start_line_2, $delete_count_2);
    
    echo "✓ 已删除 render_order_detail_page() 的内联 JS\n\n";
} else {
    echo "⚠ 未找到 render_order_detail_page() 的内联 JS（可能已经删除）\n\n";
}

// ========================================
// 保存文件
// ========================================
$new_content = implode("\n", $lines);
file_put_contents($file, $new_content);

$new_total = count($lines);
$deleted = $total_lines - $new_total;

echo "========================================\n";
echo "重构完成！\n";
echo "========================================\n";
echo "原文件行数：{$total_lines}\n";
echo "新文件行数：{$new_total}\n";
echo "删除行数：{$deleted}\n";
echo "\n";
echo "备份文件：{$backup}\n";
echo "\n";
echo "请测试功能是否正常。如果有问题，可以从备份恢复：\n";
echo "cp " . basename($backup) . " tanzanite-setting.php\n";
echo "\n";
?>
