<?php
/**
 * WordPress Ticket Export Script
 * 
 * 导出客服工单数据到 JSON 文件
 * 
 * 使用方法:
 * 1. 上传此文件到 WordPress 根目录
 * 2. 运行: php export-tickets.php
 * 3. 输出文件: scripts/wordpress-export/export/tickets.json
 */

// 加载 WordPress
require_once('wp-load.php');

// 确保输出目录存在
$export_dir = __DIR__ . '/export';
if (!file_exists($export_dir)) {
    mkdir($export_dir, 0755, true);
}

/**
 * 导出工单数据
 */
function export_tickets() {
    global $wpdb;
    
    echo "开始导出工单...\n";
    
    // 查询所有工单
    $tickets = $wpdb->get_results("
        SELECT *
        FROM {$wpdb->prefix}support_tickets
        ORDER BY id ASC
    ");
    
    $export_data = [
        'tickets' => [],
        'messages' => [],
        'stats' => [
            'total_tickets' => 0,
            'total_messages' => 0,
            'export_date' => date('Y-m-d H:i:s')
        ]
    ];
    
    foreach ($tickets as $ticket) {
        $ticket_data = [
            'id' => (int) $ticket->id,
            'ticket_number' => $ticket->ticket_number,
            'user_id' => (int) $ticket->user_id,
            'subject' => $ticket->subject,
            'category' => $ticket->category ?: 'other',
            'priority' => $ticket->priority ?: 'medium',
            'status' => $ticket->status,
            'assigned_to' => $ticket->assigned_to ? (int) $ticket->assigned_to : null,
            'tags' => $ticket->tags ?: '',
            'created_at' => $ticket->created_at,
            'updated_at' => $ticket->updated_at,
            'resolved_at' => $ticket->resolved_at ?: null,
            'closed_at' => $ticket->closed_at ?: null
        ];
        
        $export_data['tickets'][] = $ticket_data;
        
        // 导出工单消息
        $messages = export_ticket_messages($ticket->id);
        $export_data['messages'] = array_merge($export_data['messages'], $messages);
        
        echo "  导出工单: {$ticket->ticket_number} - " . count($messages) . " 条消息\n";
    }
    
    $export_data['stats']['total_tickets'] = count($export_data['tickets']);
    $export_data['stats']['total_messages'] = count($export_data['messages']);
    
    return $export_data;
}

/**
 * 导出工单消息
 */
function export_ticket_messages($ticket_id) {
    global $wpdb;
    
    $messages = $wpdb->get_results($wpdb->prepare("
        SELECT *
        FROM {$wpdb->prefix}support_ticket_messages
        WHERE ticket_id = %d
        ORDER BY created_at ASC
    ", $ticket_id));
    
    $message_data = [];
    foreach ($messages as $msg) {
        $message_data[] = [
            'ticket_id' => (int) $ticket_id,
            'user_id' => (int) $msg->user_id,
            'is_staff' => (bool) $msg->is_staff,
            'content' => $msg->content,
            'attachments' => $msg->attachments ?: '',
            'is_internal' => isset($msg->is_internal) ? (bool) $msg->is_internal : false,
            'created_at' => $msg->created_at
        ];
    }
    
    return $message_data;
}

// 执行导出
try {
    $data = export_tickets();
    
    // 保存到文件
    $output_file = $export_dir . '/tickets.json';
    file_put_contents($output_file, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    
    echo "\n导出完成!\n";
    echo "输出文件: {$output_file}\n";
    echo "工单总数: {$data['stats']['total_tickets']}\n";
    echo "消息总数: {$data['stats']['total_messages']}\n";
    echo "导出时间: {$data['stats']['export_date']}\n";
    
} catch (Exception $e) {
    echo "导出失败: " . $e->getMessage() . "\n";
    exit(1);
}
