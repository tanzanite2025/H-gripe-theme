<?php
/**
 * WordPress 设置导出脚本
 * 
 * 用途: 从 WordPress 导出站点设置到 JSON 文件
 * 运行: php export-settings.php
 * 输出: export/settings.json
 */

// 加载 WordPress
require_once('wp-load.php');

// 确保导出目录存在
$exportDir = __DIR__ . '/export';
if (!file_exists($exportDir)) {
    mkdir($exportDir, 0755, true);
}

// 定义要导出的设置
$settingsToExport = [
    // 站点设置
    'site' => [
        'site_name' => 'blogname',
        'site_description' => 'blogdescription',
        'site_logo' => 'custom_logo',
        'contact_email' => 'admin_email',
        'contact_phone' => 'contact_phone',
        'social_links' => 'social_links',
    ],
    
    // 快速购买设置
    'quick-buy' => [
        'enabled' => 'quick_buy_enabled',
        'button_text' => 'quick_buy_button_text',
        'success_message' => 'quick_buy_success_message',
        'require_login' => 'quick_buy_require_login',
    ],
    
    // SEO 设置
    'seo' => [
        'meta_title' => 'seo_meta_title',
        'meta_description' => 'seo_meta_description',
        'meta_keywords' => 'seo_meta_keywords',
        'google_analytics' => 'google_analytics_id',
        'google_tag_manager' => 'google_tag_manager_id',
    ],
    
    // 社交媒体设置
    'social' => [
        'facebook' => 'social_facebook',
        'twitter' => 'social_twitter',
        'instagram' => 'social_instagram',
        'linkedin' => 'social_linkedin',
        'youtube' => 'social_youtube',
        'wechat' => 'social_wechat',
    ],
    
    // 邮件设置
    'email' => [
        'smtp_host' => 'smtp_host',
        'smtp_port' => 'smtp_port',
        'smtp_username' => 'smtp_username',
        'smtp_password' => 'smtp_password',
        'from_email' => 'from_email',
        'from_name' => 'from_name',
    ],
];

// 导出设置
$exportData = [];
$stats = [
    'total' => 0,
    'by_group' => [],
];

foreach ($settingsToExport as $group => $settings) {
    $stats['by_group'][$group] = 0;
    
    foreach ($settings as $key => $wpOptionName) {
        $value = get_option($wpOptionName, '');
        
        // 特殊处理
        if ($key === 'site_logo' && is_numeric($value)) {
            // 获取 logo URL
            $value = wp_get_attachment_url($value);
        }
        
        // 确定类型
        $type = 'string';
        if (is_bool($value) || $value === 'true' || $value === 'false') {
            $type = 'boolean';
            $value = $value ? 'true' : 'false';
        } elseif (is_numeric($value)) {
            $type = 'number';
        } elseif (is_array($value) || is_object($value)) {
            $type = 'json';
            $value = json_encode($value);
        }
        
        // 确定是否公开
        $isPublic = true;
        if ($group === 'email' && in_array($key, ['smtp_password', 'smtp_username'])) {
            $isPublic = false;
        }
        
        $exportData[] = [
            'key' => $key,
            'value' => (string)$value,
            'type' => $type,
            'group' => $group,
            'locale' => 'en', // 默认语言
            'is_public' => $isPublic,
            'description' => ucwords(str_replace('_', ' ', $key)),
        ];
        
        $stats['total']++;
        $stats['by_group'][$group]++;
    }
}

// 保存到文件
$outputFile = $exportDir . '/settings.json';
file_put_contents($outputFile, json_encode($exportData, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));

// 输出统计信息
echo "========================================\n";
echo "WordPress 设置导出完成\n";
echo "========================================\n";
echo "导出文件: $outputFile\n";
echo "总设置数: {$stats['total']}\n";
echo "\n按分组统计:\n";
foreach ($stats['by_group'] as $group => $count) {
    echo "  - $group: $count\n";
}
echo "\n";

// 输出前 5 条示例
echo "示例数据（前 5 条）:\n";
foreach (array_slice($exportData, 0, 5) as $setting) {
    echo "  - {$setting['group']}.{$setting['key']}: {$setting['value']}\n";
}
echo "\n";

echo "下一步:\n";
echo "1. 检查导出的 JSON 文件\n";
echo "2. 运行 Go 导入工具: go run cmd/import/settings.go\n";
echo "========================================\n";
