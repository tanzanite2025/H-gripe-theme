<?php
/**
 * WordPress 博客翻译数据导出脚本
 * 
 * 使用方法:
 * 1. 将此文件放到 WordPress 根目录
 * 2. 在浏览器访问: http://your-site.com/export-blog-translations.php
 * 或在命令行运行: php export-blog-translations.php
 */

// 加载 WordPress
require_once('wp-load.php');

// 安全检查：只允许管理员运行
if (!is_admin() && !defined('WP_CLI')) {
    if (!current_user_can('administrator')) {
        die('Unauthorized: Only administrators can run this script.');
    }
}

// 设置输出目录
$export_dir = WP_CONTENT_DIR . '/exports';
if (!file_exists($export_dir)) {
    mkdir($export_dir, 0755, true);
}

echo "========================================\n";
echo "博客翻译数据导出工具\n";
echo "========================================\n\n";

/**
 * 导出文章翻译关联数据
 */
function export_post_translations() {
    global $wpdb;
    
    echo "正在导出文章翻译关联...\n";
    
    // 获取所有文章
    $posts = get_posts([
        'post_type' => 'post',
        'post_status' => 'any',
        'numberposts' => -1,
        'orderby' => 'ID',
        'order' => 'ASC',
    ]);
    
    $translations = [];
    $groups = [];
    $post_count = 0;
    $group_count = 0;
    
    foreach ($posts as $post) {
        // 获取翻译组ID
        $translation_group = get_post_meta($post->ID, 'translation_group_id', true);
        
        // 如果没有翻译组ID，尝试从其他元数据获取
        if (empty($translation_group)) {
            // 检查是否有 parent_id（旧的翻译关联方式）
            $parent_id = get_post_meta($post->ID, 'parent_id', true);
            if (!empty($parent_id)) {
                // 使用 parent_id 作为翻译组ID
                $translation_group = $parent_id;
            } else {
                // 如果都没有，使用文章自己的ID作为翻译组ID
                $translation_group = $post->ID;
            }
        }
        
        // 获取语言代码
        $locale = get_post_meta($post->ID, 'locale', true);
        if (empty($locale)) {
            $locale = 'en'; // 默认英文
        }
        
        // 获取 SEO 元数据
        $meta_keywords = get_post_meta($post->ID, 'meta_keywords', true);
        $canonical_url = get_post_meta($post->ID, 'canonical_url', true);
        
        // 构建翻译数据
        $translation_data = [
            'wordpress_post_id' => $post->ID,
            'translation_group_id' => intval($translation_group),
            'locale' => $locale,
            'slug' => $post->post_name,
            'title' => $post->post_title,
            'status' => $post->post_status,
            'published_at' => $post->post_date,
            'modified_at' => $post->post_modified,
            'meta_keywords' => $meta_keywords,
            'canonical_url' => $canonical_url,
        ];
        
        $translations[] = $translation_data;
        
        // 按翻译组分组
        $group_id = intval($translation_group);
        if (!isset($groups[$group_id])) {
            $groups[$group_id] = [];
            $group_count++;
        }
        $groups[$group_id][] = $translation_data;
        
        $post_count++;
    }
    
    echo "  - 找到 {$post_count} 篇文章\n";
    echo "  - 找到 {$group_count} 个翻译组\n";
    
    // 统计每个翻译组的语言数量
    $multi_lang_groups = 0;
    foreach ($groups as $group_id => $group_posts) {
        if (count($group_posts) > 1) {
            $multi_lang_groups++;
        }
    }
    echo "  - 其中 {$multi_lang_groups} 个翻译组有多个语言版本\n";
    
    return [
        'translations' => $translations,
        'groups' => $groups,
        'stats' => [
            'total_posts' => $post_count,
            'total_groups' => $group_count,
            'multi_lang_groups' => $multi_lang_groups,
        ]
    ];
}

/**
 * 导出语言统计
 */
function export_language_stats($translations) {
    echo "\n正在统计语言分布...\n";
    
    $locale_stats = [];
    foreach ($translations as $trans) {
        $locale = $trans['locale'];
        if (!isset($locale_stats[$locale])) {
            $locale_stats[$locale] = 0;
        }
        $locale_stats[$locale]++;
    }
    
    // 排序
    arsort($locale_stats);
    
    echo "  语言分布:\n";
    foreach ($locale_stats as $locale => $count) {
        echo "    - {$locale}: {$count} 篇\n";
    }
    
    return $locale_stats;
}

/**
 * 导出示例翻译组
 */
function export_sample_groups($groups, $limit = 5) {
    echo "\n示例翻译组（前 {$limit} 个多语言组）:\n";
    
    $count = 0;
    foreach ($groups as $group_id => $group_posts) {
        if (count($group_posts) > 1 && $count < $limit) {
            echo "  翻译组 #{$group_id}:\n";
            foreach ($group_posts as $post) {
                echo "    - [{$post['locale']}] {$post['title']} (ID: {$post['wordpress_post_id']})\n";
            }
            $count++;
        }
    }
}

// 执行导出
try {
    // 1. 导出翻译数据
    $result = export_post_translations();
    
    // 2. 统计语言分布
    $locale_stats = export_language_stats($result['translations']);
    
    // 3. 显示示例
    export_sample_groups($result['groups']);
    
    // 4. 保存为 JSON
    $export_data = [
        'export_date' => date('Y-m-d H:i:s'),
        'wordpress_version' => get_bloginfo('version'),
        'site_url' => get_site_url(),
        'stats' => $result['stats'],
        'locale_stats' => $locale_stats,
        'translations' => $result['translations'],
        'groups' => $result['groups'],
    ];
    
    $json_file = $export_dir . '/blog-translations.json';
    $json = json_encode($export_data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
    file_put_contents($json_file, $json);
    
    echo "\n========================================\n";
    echo "导出完成！\n";
    echo "========================================\n";
    echo "文件位置: {$json_file}\n";
    echo "文件大小: " . number_format(filesize($json_file) / 1024, 2) . " KB\n";
    echo "\n下一步:\n";
    echo "1. 下载导出的 JSON 文件\n";
    echo "2. 将文件复制到 Go 项目的 exports/ 目录\n";
    echo "3. 运行 Go 导入工具: go run cmd/import/blog_translations.go\n";
    
} catch (Exception $e) {
    echo "\n错误: " . $e->getMessage() . "\n";
    exit(1);
}
