<?php
/**
 * WordPress 订阅数据导出脚本
 * 
 * 用途：从 WordPress 数据库导出订阅数据到 JSON 文件
 * 使用方法：php export-subscriptions.php
 */

// WordPress 配置
define('DB_HOST', 'localhost');
define('DB_NAME', 'wordpress_db');
define('DB_USER', 'root');
define('DB_PASSWORD', '');
define('TABLE_PREFIX', 'wp_');

// 输出文件
$outputFile = __DIR__ . '/../../data/subscriptions.json';

try {
    // 连接数据库
    $pdo = new PDO(
        "mysql:host=" . DB_HOST . ";dbname=" . DB_NAME . ";charset=utf8mb4",
        DB_USER,
        DB_PASSWORD,
        [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
    );

    echo "正在导出订阅数据...\n";

    // 查询订阅数据
    // 注意：根据实际的 WordPress 订阅插件表结构调整查询
    $sql = "
        SELECT 
            id,
            email,
            status,
            locale,
            source,
            created_at,
            updated_at
        FROM " . TABLE_PREFIX . "subscriptions
        ORDER BY id ASC
    ";

    $stmt = $pdo->query($sql);
    $subscriptions = $stmt->fetchAll(PDO::FETCH_ASSOC);

    // 处理数据
    $exportData = [];
    foreach ($subscriptions as $sub) {
        // 生成退订令牌
        $unsubToken = bin2hex(random_bytes(16));
        
        // 根据来源推断标签
        $tags = [];
        if (!empty($sub['source'])) {
            $tags[] = $sub['source'];
        }
        
        // 如果有语言信息，添加语言标签
        if (!empty($sub['locale']) && $sub['locale'] !== 'en') {
            $tags[] = 'locale_' . $sub['locale'];
        }

        $exportData[] = [
            'id' => (int)$sub['id'],
            'email' => $sub['email'],
            'status' => $sub['status'] ?: 'active',
            'locale' => $sub['locale'] ?: 'en',
            'source' => $sub['source'] ?: 'website',
            'tags' => implode(',', $tags),
            'unsub_token' => $unsubToken,
            'subscribed_at' => $sub['created_at'],
            'unsubscribed_at' => ($sub['status'] === 'unsubscribed') ? $sub['updated_at'] : null,
            'created_at' => $sub['created_at'],
            'updated_at' => $sub['updated_at'],
        ];
    }

    // 确保输出目录存在
    $outputDir = dirname($outputFile);
    if (!is_dir($outputDir)) {
        mkdir($outputDir, 0755, true);
    }

    // 写入 JSON 文件
    $json = json_encode($exportData, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
    file_put_contents($outputFile, $json);

    echo "✓ 成功导出 " . count($exportData) . " 条订阅记录\n";
    echo "✓ 输出文件: $outputFile\n";

    // 显示统计信息
    $stats = [
        'total' => count($exportData),
        'active' => 0,
        'unsubscribed' => 0,
        'bounced' => 0,
    ];

    foreach ($exportData as $sub) {
        if (isset($stats[$sub['status']])) {
            $stats[$sub['status']]++;
        }
    }

    echo "\n统计信息:\n";
    echo "  总计: {$stats['total']}\n";
    echo "  活跃: {$stats['active']}\n";
    echo "  已退订: {$stats['unsubscribed']}\n";
    echo "  退回: {$stats['bounced']}\n";

} catch (PDOException $e) {
    echo "错误: " . $e->getMessage() . "\n";
    exit(1);
}
