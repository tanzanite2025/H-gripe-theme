<?php
/**
 * 修复插件语法错误
 * 
 * 使用方法：
 * php fix-syntax-errors.php
 */

$file = __DIR__ . '/tanzanite-setting.php';

if (!file_exists($file)) {
    die("错误：找不到文件 tanzanite-setting.php\n");
}

echo "开始修复语法错误...\n\n";

// 备份原文件
$backup = $file . '.backup.' . date('YmdHis');
copy($file, $backup);
echo "✓ 已创建备份：" . basename($backup) . "\n\n";

// 读取文件
$content = file_get_contents($file);
$lines = explode("\n", $content);

echo "原文件行数：" . count($lines) . "\n\n";

// 删除第 12689-13100 行之间的孤立 JS 代码
echo "步骤 1: 删除孤立的 JS 代码（第 12689-13100 行）\n";
echo "----------------------------------------\n";

// 找到要删除的起始和结束行
$start_line = 12688; // 0-indexed, 所以是 12689-1
$end_line = 13099;   // 0-indexed, 所以是 13100-1

// 删除这些行
array_splice($lines, $start_line, $end_line - $start_line + 1);

echo "✓ 已删除 " . ($end_line - $start_line + 1) . " 行孤立代码\n\n";

// 重新组合文件
$content = implode("\n", $lines);

// 保存文件
file_put_contents($file, $content);

$new_line_count = count($lines);
echo "新文件行数：{$new_line_count}\n\n";

// 测试 PHP 语法
echo "步骤 2: 测试 PHP 语法\n";
echo "----------------------------------------\n";

$output = [];
$return_var = 0;
exec("php -l " . escapeshellarg($file) . " 2>&1", $output, $return_var);

if ($return_var === 0) {
    echo "✅ PHP 语法检查通过！\n\n";
} else {
    echo "❌ PHP 语法检查失败：\n";
    echo implode("\n", $output) . "\n\n";
}

echo "========================================\n";
echo "修复完成！\n";
echo "========================================\n";
echo "\n";
echo "备份文件：{$backup}\n";
echo "\n";

if ($return_var === 0) {
    echo "✅ 所有语法错误已修复！\n";
    echo "\n";
    echo "下一步：\n";
    echo "1. 刷新 WordPress 后台\n";
    echo "2. 测试所有功能\n";
} else {
    echo "⚠️ 还有语法错误需要手动修复\n";
    echo "\n";
    echo "如果需要恢复备份：\n";
    echo "cp " . basename($backup) . " tanzanite-setting.php\n";
}

echo "\n";
?>
