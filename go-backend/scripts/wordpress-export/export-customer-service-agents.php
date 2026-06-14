<?php
/**
 * WordPress customer service agent export script.
 *
 * Run from the WordPress root: php go-backend/scripts/wordpress-export/export-customer-service-agents.php
 * Output: go-backend/scripts/wordpress-export/export/customer-service-agents.json
 */

require_once('wp-load.php');

global $wpdb;

$exportDir = __DIR__ . '/export';
if (!file_exists($exportDir)) {
    mkdir($exportDir, 0755, true);
}

$table = $wpdb->prefix . 'tz_cs_agents';
$tableExists = $wpdb->get_var($wpdb->prepare('SHOW TABLES LIKE %s', $table)) === $table;

$agents = [];
if ($tableExists) {
    $agents = $wpdb->get_results(
        "SELECT id, agent_id, wp_user_id, name, email, avatar, whatsapp, pre_sales_email, after_sales_email, status, online_status, last_active_at, last_login, created_at, updated_at FROM {$table} ORDER BY created_at ASC",
        ARRAY_A
    );
}

$exportData = array_map(static function ($agent) {
    return [
        'id' => isset($agent['id']) ? (int) $agent['id'] : null,
        'agent_id' => $agent['agent_id'] ?? '',
        'wp_user_id' => !empty($agent['wp_user_id']) ? (int) $agent['wp_user_id'] : null,
        'name' => $agent['name'] ?? '',
        'email' => $agent['email'] ?? '',
        'avatar' => $agent['avatar'] ?? '',
        'whatsapp' => $agent['whatsapp'] ?? '',
        'pre_sales_email' => $agent['pre_sales_email'] ?? '',
        'after_sales_email' => $agent['after_sales_email'] ?? '',
        'status' => $agent['status'] ?? 'active',
        'online_status' => $agent['online_status'] ?? 'offline',
        'last_active_at' => $agent['last_active_at'] ?? null,
        'last_login' => $agent['last_login'] ?? null,
        'created_at' => $agent['created_at'] ?? null,
        'updated_at' => $agent['updated_at'] ?? null,
    ];
}, $agents);

$outputFile = $exportDir . '/customer-service-agents.json';
file_put_contents($outputFile, json_encode($exportData, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));

echo "========================================\n";
echo "Customer service agents export complete\n";
echo "========================================\n";
echo "Export file: {$outputFile}\n";
echo "Total agents: " . count($exportData) . "\n";
echo "Next: copy the JSON into go-backend/scripts/export/customer-service-agents.json and run go run scripts/import-data.go\n";
