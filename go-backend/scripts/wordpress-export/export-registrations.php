<?php
/**
 * WordPress Product Registration Export Script
 * 
 * 导出产品注册和保修申请数据到 JSON 文件
 * 
 * 使用方法:
 * 1. 上传此文件到 WordPress 根目录
 * 2. 运行: php export-registrations.php
 * 3. 输出文件: scripts/wordpress-export/export/registrations.json
 */

// 加载 WordPress
require_once('wp-load.php');

// 确保输出目录存在
$export_dir = __DIR__ . '/export';
if (!file_exists($export_dir)) {
    mkdir($export_dir, 0755, true);
}

/**
 * 导出产品注册数据
 */
function export_registrations() {
    global $wpdb;
    
    echo "开始导出产品注册...\n";
    
    // 查询所有产品注册
    $registrations = $wpdb->get_results("
        SELECT *
        FROM {$wpdb->prefix}product_registrations
        ORDER BY id ASC
    ");
    
    $export_data = [
        'registrations' => [],
        'warranty_claims' => [],
        'stats' => [
            'total_registrations' => 0,
            'total_claims' => 0,
            'export_date' => date('Y-m-d H:i:s')
        ]
    ];
    
    foreach ($registrations as $reg) {
        $registration_data = [
            'id' => (int) $reg->id,
            'user_id' => (int) $reg->user_id,
            'product_id' => (int) $reg->product_id,
            'serial_number' => $reg->serial_number,
            'purchase_date' => $reg->purchase_date,
            'purchase_proof' => $reg->purchase_proof ?: '',
            'retailer' => $reg->retailer ?: '',
            'warranty_period' => (int) $reg->warranty_period,
            'warranty_expires' => $reg->warranty_expires,
            'status' => $reg->status,
            'notes' => $reg->notes ?: '',
            'created_at' => $reg->created_at,
            'updated_at' => $reg->updated_at
        ];
        
        $export_data['registrations'][] = $registration_data;
        
        echo "  导出注册: {$reg->serial_number} (ID: {$reg->id})\n";
    }
    
    // 导出保修申请
    $claims = $wpdb->get_results("
        SELECT *
        FROM {$wpdb->prefix}warranty_claims
        ORDER BY id ASC
    ");
    
    foreach ($claims as $claim) {
        $claim_data = [
            'id' => (int) $claim->id,
            'registration_id' => (int) $claim->registration_id,
            'user_id' => (int) $claim->user_id,
            'issue_type' => $claim->issue_type,
            'description' => $claim->description,
            'images' => $claim->images ?: '',
            'status' => $claim->status,
            'resolution' => $claim->resolution ?: '',
            'processed_by' => $claim->processed_by ? (int) $claim->processed_by : null,
            'processed_at' => $claim->processed_at ?: null,
            'created_at' => $claim->created_at,
            'updated_at' => $claim->updated_at
        ];
        
        $export_data['warranty_claims'][] = $claim_data;
        
        echo "  导出保修申请: ID {$claim->id} (注册 ID: {$claim->registration_id})\n";
    }
    
    $export_data['stats']['total_registrations'] = count($export_data['registrations']);
    $export_data['stats']['total_claims'] = count($export_data['warranty_claims']);
    
    return $export_data;
}

// 执行导出
try {
    $data = export_registrations();
    
    // 保存到文件
    $output_file = $export_dir . '/registrations.json';
    file_put_contents($output_file, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    
    echo "\n导出完成!\n";
    echo "输出文件: {$output_file}\n";
    echo "产品注册总数: {$data['stats']['total_registrations']}\n";
    echo "保修申请总数: {$data['stats']['total_claims']}\n";
    echo "导出时间: {$data['stats']['export_date']}\n";
    
} catch (Exception $e) {
    echo "导出失败: " . $e->getMessage() . "\n";
    exit(1);
}
