<?php
/**
 * WordPress Gallery Export Script
 * 
 * 导出图片库和图片数据到 JSON 文件
 * 
 * 使用方法:
 * 1. 上传此文件到 WordPress 根目录
 * 2. 运行: php export-galleries.php
 * 3. 输出文件: scripts/wordpress-export/export/galleries.json
 */

// 加载 WordPress
require_once('wp-load.php');

// 确保输出目录存在
$export_dir = __DIR__ . '/export';
if (!file_exists($export_dir)) {
    mkdir($export_dir, 0755, true);
}

/**
 * 导出图片库数据
 */
function export_galleries() {
    global $wpdb;
    
    echo "开始导出图片库...\n";
    
    // 查询所有图片库
    $galleries = $wpdb->get_results("
        SELECT 
            p.ID as id,
            p.post_title as name,
            p.post_name as slug,
            p.post_content as description,
            p.post_status as status,
            p.post_date as created_at,
            p.post_modified as updated_at
        FROM {$wpdb->posts} p
        WHERE p.post_type = 'gallery'
        AND p.post_status IN ('publish', 'draft')
        ORDER BY p.ID ASC
    ");
    
    $export_data = [
        'galleries' => [],
        'images' => [],
        'stats' => [
            'total_galleries' => 0,
            'total_images' => 0,
            'export_date' => date('Y-m-d H:i:s')
        ]
    ];
    
    foreach ($galleries as $gallery) {
        // 获取封面图片
        $cover_image = get_post_meta($gallery->id, 'cover_image', true);
        if (empty($cover_image)) {
            $thumbnail_id = get_post_thumbnail_id($gallery->id);
            if ($thumbnail_id) {
                $cover_image = wp_get_attachment_url($thumbnail_id);
            }
        }
        
        // 获取浏览次数
        $view_count = (int) get_post_meta($gallery->id, 'view_count', true);
        
        // 转换状态
        $status = $gallery->status === 'publish' ? 'published' : 'draft';
        
        $gallery_data = [
            'id' => (int) $gallery->id,
            'name' => $gallery->name,
            'slug' => $gallery->slug,
            'description' => $gallery->description,
            'cover_image' => $cover_image ?: '',
            'view_count' => $view_count,
            'status' => $status,
            'created_at' => $gallery->created_at,
            'updated_at' => $gallery->updated_at
        ];
        
        $export_data['galleries'][] = $gallery_data;
        
        // 导出图片库的图片
        $images = export_gallery_images($gallery->id);
        $export_data['images'] = array_merge($export_data['images'], $images);
        
        echo "  导出图片库: {$gallery->name} ({$gallery->id}) - " . count($images) . " 张图片\n";
    }
    
    $export_data['stats']['total_galleries'] = count($export_data['galleries']);
    $export_data['stats']['total_images'] = count($export_data['images']);
    
    return $export_data;
}

/**
 * 导出图片库的图片
 */
function export_gallery_images($gallery_id) {
    global $wpdb;
    
    $images = [];
    
    // 方法1: 从 post meta 获取图片（如果使用自定义字段存储）
    $gallery_images = get_post_meta($gallery_id, 'gallery_images', true);
    
    if (!empty($gallery_images) && is_array($gallery_images)) {
        $order = 1;
        foreach ($gallery_images as $image_data) {
            if (is_array($image_data)) {
                $images[] = format_image_data($gallery_id, $image_data, $order++);
            } elseif (is_numeric($image_data)) {
                // 如果只存储了附件ID
                $attachment_data = get_attachment_data($image_data);
                if ($attachment_data) {
                    $images[] = format_image_data($gallery_id, $attachment_data, $order++);
                }
            }
        }
    }
    
    // 方法2: 从附件关系获取（如果图片作为附件关联到图片库）
    if (empty($images)) {
        $attachments = get_posts([
            'post_type' => 'attachment',
            'post_parent' => $gallery_id,
            'post_mime_type' => 'image',
            'numberposts' => -1,
            'orderby' => 'menu_order',
            'order' => 'ASC'
        ]);
        
        $order = 1;
        foreach ($attachments as $attachment) {
            $attachment_data = get_attachment_data($attachment->ID);
            if ($attachment_data) {
                $images[] = format_image_data($gallery_id, $attachment_data, $order++);
            }
        }
    }
    
    // 方法3: 从自定义表获取（如果使用自定义表存储）
    if (empty($images)) {
        $custom_images = $wpdb->get_results($wpdb->prepare("
            SELECT *
            FROM {$wpdb->prefix}gallery_images
            WHERE gallery_id = %d
            ORDER BY display_order ASC, id ASC
        ", $gallery_id));
        
        foreach ($custom_images as $img) {
            $images[] = [
                'gallery_id' => (int) $gallery_id,
                'url' => $img->url,
                'thumbnail' => $img->thumbnail ?: '',
                'title' => $img->title ?: '',
                'description' => $img->description ?: '',
                'alt' => $img->alt ?: '',
                'width' => (int) $img->width,
                'height' => (int) $img->height,
                'size' => (int) $img->size,
                'tags' => $img->tags ?: '',
                'order' => (int) $img->display_order
            ];
        }
    }
    
    return $images;
}

/**
 * 从附件ID获取数据
 */
function get_attachment_data($attachment_id) {
    $url = wp_get_attachment_url($attachment_id);
    if (!$url) {
        return null;
    }
    
    $metadata = wp_get_attachment_metadata($attachment_id);
    $attachment = get_post($attachment_id);
    
    // 获取缩略图
    $thumbnail = wp_get_attachment_image_url($attachment_id, 'thumbnail');
    
    // 获取标签
    $tags = wp_get_post_terms($attachment_id, 'attachment_tag', ['fields' => 'names']);
    $tags_str = !empty($tags) ? implode(',', $tags) : '';
    
    return [
        'url' => $url,
        'thumbnail' => $thumbnail ?: '',
        'title' => $attachment->post_title,
        'description' => $attachment->post_content,
        'alt' => get_post_meta($attachment_id, '_wp_attachment_image_alt', true),
        'width' => isset($metadata['width']) ? $metadata['width'] : 0,
        'height' => isset($metadata['height']) ? $metadata['height'] : 0,
        'size' => filesize(get_attached_file($attachment_id)) ?: 0,
        'tags' => $tags_str,
        'order' => $attachment->menu_order
    ];
}

/**
 * 格式化图片数据
 */
function format_image_data($gallery_id, $data, $order) {
    return [
        'gallery_id' => (int) $gallery_id,
        'url' => $data['url'] ?? '',
        'thumbnail' => $data['thumbnail'] ?? '',
        'title' => $data['title'] ?? '',
        'description' => $data['description'] ?? '',
        'alt' => $data['alt'] ?? '',
        'width' => (int) ($data['width'] ?? 0),
        'height' => (int) ($data['height'] ?? 0),
        'size' => (int) ($data['size'] ?? 0),
        'tags' => $data['tags'] ?? '',
        'order' => (int) ($data['order'] ?? $order)
    ];
}

// 执行导出
try {
    $data = export_galleries();
    
    // 保存到文件
    $output_file = $export_dir . '/galleries.json';
    file_put_contents($output_file, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    
    echo "\n导出完成!\n";
    echo "输出文件: {$output_file}\n";
    echo "图片库总数: {$data['stats']['total_galleries']}\n";
    echo "图片总数: {$data['stats']['total_images']}\n";
    echo "导出时间: {$data['stats']['export_date']}\n";
    
} catch (Exception $e) {
    echo "导出失败: " . $e->getMessage() . "\n";
    exit(1);
}
