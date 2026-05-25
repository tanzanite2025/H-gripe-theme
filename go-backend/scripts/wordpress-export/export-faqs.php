<?php
/**
 * WordPress FAQ 导出脚本
 * 
 * 用途: 从 WordPress 导出 FAQ 数据到 JSON 文件
 * 运行: php export-faqs.php
 * 输出: export/faqs.json
 */

// 加载 WordPress
require_once('wp-load.php');

// 确保导出目录存在
$exportDir = __DIR__ . '/export';
if (!file_exists($exportDir)) {
    mkdir($exportDir, 0755, true);
}

// 获取所有 FAQ
$args = [
    'post_type' => 'faq',
    'posts_per_page' => -1,
    'post_status' => ['publish', 'draft'],
    'orderby' => 'menu_order',
    'order' => 'ASC',
];

$faqs = get_posts($args);

// 导出数据
$exportData = [];
$stats = [
    'total' => 0,
    'by_category' => [],
    'by_locale' => [],
    'by_status' => [],
];

foreach ($faqs as $faq) {
    // 获取元数据
    $category = get_post_meta($faq->ID, 'faq_category', true) ?: 'General';
    $locale = get_post_meta($faq->ID, 'locale', true) ?: 'en';
    $parentID = get_post_meta($faq->ID, 'translation_parent_id', true);
    $order = get_post_meta($faq->ID, 'menu_order', true) ?: 0;
    $viewCount = get_post_meta($faq->ID, 'view_count', true) ?: 0;
    
    // 确定状态
    $status = $faq->post_status === 'publish' ? 'published' : 'draft';
    
    $exportData[] = [
        'id' => $faq->ID,
        'question' => $faq->post_title,
        'answer' => $faq->post_content,
        'category' => $category,
        'locale' => $locale,
        'parent_id' => $parentID ? (int)$parentID : null,
        'order' => (int)$order,
        'status' => $status,
        'view_count' => (int)$viewCount,
        'created_at' => $faq->post_date,
        'updated_at' => $faq->post_modified,
    ];
    
    // 统计
    $stats['total']++;
    
    if (!isset($stats['by_category'][$category])) {
        $stats['by_category'][$category] = 0;
    }
    $stats['by_category'][$category]++;
    
    if (!isset($stats['by_locale'][$locale])) {
        $stats['by_locale'][$locale] = 0;
    }
    $stats['by_locale'][$locale]++;
    
    if (!isset($stats['by_status'][$status])) {
        $stats['by_status'][$status] = 0;
    }
    $stats['by_status'][$status]++;
}

// 保存到文件
$outputFile = $exportDir . '/faqs.json';
file_put_contents($outputFile, json_encode($exportData, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));

// 输出统计信息
echo "========================================\n";
echo "WordPress FAQ 导出完成\n";
echo "========================================\n";
echo "导出文件: $outputFile\n";
echo "总 FAQ 数: {$stats['total']}\n";
echo "\n按分类统计:\n";
foreach ($stats['by_category'] as $category => $count) {
    echo "  - $category: $count\n";
}
echo "\n按语言统计:\n";
foreach ($stats['by_locale'] as $locale => $count) {
    echo "  - $locale: $count\n";
}
echo "\n按状态统计:\n";
foreach ($stats['by_status'] as $status => $count) {
    echo "  - $status: $count\n";
}
echo "\n";

// 输出前 3 条示例
echo "示例数据（前 3 条）:\n";
foreach (array_slice($exportData, 0, 3) as $faq) {
    echo "  - [{$faq['category']}] {$faq['question']}\n";
}
echo "\n";

echo "下一步:\n";
echo "1. 检查导出的 JSON 文件\n";
echo "2. 运行 Go 导入工具: go run cmd/import/faqs.go\n";
echo "========================================\n";
