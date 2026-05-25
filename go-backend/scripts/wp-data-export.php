<?php
/**
 * WordPress Data Export Script
 * 
 * 在WordPress环境中运行此脚本，导出数据为JSON格式
 * 用于迁移到Go后端
 * 
 * 使用方法:
 * php wp-data-export.php
 * 或在WordPress根目录运行: wp-cli eval-file scripts/wp-data-export.php
 */

// 加载WordPress
require_once(__DIR__ . '/../../../../wp-load.php');

// 创建导出目录
$export_dir = __DIR__ . '/export';
if (!file_exists($export_dir)) {
    mkdir($export_dir, 0755, true);
}

echo "Starting WordPress data export...\n";

/**
 * 导出用户
 */
function export_users($export_dir) {
    echo "Exporting users...\n";
    
    $users = get_users(['number' => -1]);
    $data = [];
    
    foreach ($users as $user) {
        $data[] = [
            'id' => $user->ID,
            'email' => $user->user_email,
            'username' => $user->user_login,
            'first_name' => get_user_meta($user->ID, 'first_name', true),
            'last_name' => get_user_meta($user->ID, 'last_name', true),
            'role' => implode(',', $user->roles),
            'locale' => get_user_meta($user->ID, 'locale', true) ?: 'en',
            'registered' => $user->user_registered,
        ];
    }
    
    file_put_contents($export_dir . '/users.json', json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    echo "Exported " . count($data) . " users\n";
}

/**
 * 导出文章
 */
function export_posts($export_dir) {
    echo "Exporting posts...\n";
    
    $posts = get_posts([
        'numberposts' => -1,
        'post_type' => 'post',
        'post_status' => 'any'
    ]);
    
    $data = [];
    
    foreach ($posts as $post) {
        $locale = get_post_meta($post->ID, 'locale', true) ?: 'en';
        $parent_id = get_post_meta($post->ID, 'translation_parent', true) ?: null;
        
        $data[] = [
            'id' => $post->ID,
            'title' => $post->post_title,
            'slug' => $post->post_name,
            'content' => $post->post_content,
            'excerpt' => $post->post_excerpt,
            'status' => $post->post_status,
            'author_id' => $post->post_author,
            'locale' => $locale,
            'parent_id' => $parent_id,
            'featured_image' => get_the_post_thumbnail_url($post->ID, 'full'),
            'meta_title' => get_post_meta($post->ID, '_yoast_wpseo_title', true),
            'meta_description' => get_post_meta($post->ID, '_yoast_wpseo_metadesc', true),
            'tags' => implode(',', wp_get_post_tags($post->ID, ['fields' => 'names'])),
            'created_at' => $post->post_date,
            'updated_at' => $post->post_modified,
            'published_at' => $post->post_status === 'publish' ? $post->post_date : null,
        ];
    }
    
    file_put_contents($export_dir . '/posts.json', json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    echo "Exported " . count($data) . " posts\n";
}

/**
 * 导出产品（WooCommerce）
 */
function export_products($export_dir) {
    echo "Exporting products...\n";
    
    if (!class_exists('WooCommerce')) {
        echo "WooCommerce not installed, skipping products\n";
        return;
    }
    
    $args = [
        'post_type' => 'product',
        'posts_per_page' => -1,
        'post_status' => 'any'
    ];
    
    $products = get_posts($args);
    $data = [];
    
    foreach ($products as $post) {
        $product = wc_get_product($post->ID);
        if (!$product) continue;
        
        $locale = get_post_meta($post->ID, 'locale', true) ?: 'en';
        $parent_id = get_post_meta($post->ID, 'translation_parent', true) ?: null;
        
        $images = [];
        $image_ids = $product->get_gallery_image_ids();
        foreach ($image_ids as $index => $image_id) {
            $images[] = [
                'url' => wp_get_attachment_url($image_id),
                'alt' => get_post_meta($image_id, '_wp_attachment_image_alt', true),
                'order' => $index
            ];
        }
        
        $data[] = [
            'id' => $product->get_id(),
            'sku' => $product->get_sku(),
            'name' => $product->get_name(),
            'slug' => $product->get_slug(),
            'description' => $product->get_description(),
            'short_description' => $product->get_short_description(),
            'price' => $product->get_regular_price(),
            'sale_price' => $product->get_sale_price() ?: null,
            'stock' => $product->get_stock_quantity() ?: 0,
            'weight_grams' => $product->get_weight() ? (float)$product->get_weight() * 1000 : 0,
            'status' => $product->get_status(),
            'locale' => $locale,
            'parent_id' => $parent_id,
            'featured' => $product->is_featured(),
            'meta_title' => get_post_meta($post->ID, '_yoast_wpseo_title', true),
            'meta_description' => get_post_meta($post->ID, '_yoast_wpseo_metadesc', true),
            'images' => $images,
            'created_at' => $post->post_date,
            'updated_at' => $post->post_modified,
        ];
    }
    
    file_put_contents($export_dir . '/products.json', json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    echo "Exported " . count($data) . " products\n";
}

/**
 * 导出设置
 */
function export_settings($export_dir) {
    echo "Exporting settings...\n";
    
    $settings = [
        [
            'key' => 'site_name',
            'value' => get_bloginfo('name'),
            'type' => 'string',
            'group' => 'site',
            'locale' => 'en'
        ],
        [
            'key' => 'site_description',
            'value' => get_bloginfo('description'),
            'type' => 'string',
            'group' => 'site',
            'locale' => 'en'
        ],
        [
            'key' => 'contact_email',
            'value' => get_option('admin_email'),
            'type' => 'string',
            'group' => 'site',
            'locale' => 'en'
        ],
    ];
    
    file_put_contents($export_dir . '/settings.json', json_encode($settings, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    echo "Exported " . count($settings) . " settings\n";
}

/**
 * 导出FAQ
 */
function export_faqs($export_dir) {
    echo "Exporting FAQs...\n";
    
    $faqs = get_posts([
        'numberposts' => -1,
        'post_type' => 'faq',
        'post_status' => 'any'
    ]);
    
    $data = [];
    
    foreach ($faqs as $post) {
        $locale = get_post_meta($post->ID, 'locale', true) ?: 'en';
        $parent_id = get_post_meta($post->ID, 'translation_parent', true) ?: null;
        
        $data[] = [
            'id' => $post->ID,
            'question' => $post->post_title,
            'answer' => $post->post_content,
            'category' => get_post_meta($post->ID, 'faq_category', true),
            'locale' => $locale,
            'parent_id' => $parent_id,
            'order' => get_post_meta($post->ID, 'faq_order', true) ?: 0,
            'status' => $post->post_status,
            'created_at' => $post->post_date,
            'updated_at' => $post->post_modified,
        ];
    }
    
    file_put_contents($export_dir . '/faqs.json', json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    echo "Exported " . count($data) . " FAQs\n";
}

// 执行导出
try {
    export_users($export_dir);
    export_posts($export_dir);
    export_products($export_dir);
    export_settings($export_dir);
    export_faqs($export_dir);
    
    echo "\n✅ Export completed successfully!\n";
    echo "Files saved to: $export_dir\n";
} catch (Exception $e) {
    echo "\n❌ Export failed: " . $e->getMessage() . "\n";
    exit(1);
}
