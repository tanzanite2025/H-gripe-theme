<?php
/**
 * 货币迁移脚本：从 CNY 改为 USD
 * 
 * 使用方法：
 * 1. 将此文件放在插件根目录
 * 2. 在浏览器访问：http://your-site.com/wp-content/plugins/tanzanite-setting/migrate-currency-to-usd.php
 * 3. 或者在 WordPress 后台运行
 * 
 * @package TanzaniteSettings
 * @version 1.0.0
 */

// 加载 WordPress
require_once('../../../wp-load.php');

// 检查权限
if (!current_user_can('manage_options')) {
    wp_die('您没有权限执行此操作');
}

global $wpdb;

echo '<h1>货币迁移：CNY → USD</h1>';
echo '<p>开始时间：' . date('Y-m-d H:i:s') . '</p>';

// 1. 更新订单表结构
echo '<h2>1. 更新订单表默认货币</h2>';
$orders_table = $wpdb->prefix . 'tanz_orders';
$result1 = $wpdb->query("ALTER TABLE `{$orders_table}` MODIFY COLUMN `currency` CHAR(3) NOT NULL DEFAULT 'USD'");
if ($result1 === false) {
    echo '<p style="color:red;">❌ 失败：' . $wpdb->last_error . '</p>';
} else {
    echo '<p style="color:green;">✅ 成功更新订单表默认货币为 USD</p>';
}

// 2. 统计现有订单的货币分布
echo '<h2>2. 现有订单货币统计</h2>';
$currency_stats = $wpdb->get_results("SELECT currency, COUNT(*) as count FROM `{$orders_table}` GROUP BY currency");
if ($currency_stats) {
    echo '<ul>';
    foreach ($currency_stats as $stat) {
        echo '<li>' . esc_html($stat->currency) . ': ' . esc_html($stat->count) . ' 个订单</li>';
    }
    echo '</ul>';
} else {
    echo '<p>没有找到订单</p>';
}

// 3. 更新现有 CNY 订单为 USD（可选，默认注释掉）
echo '<h2>3. 更新现有订单货币</h2>';
echo '<p><strong>注意：</strong>此操作会将所有 CNY 订单改为 USD。如需执行，请取消下面代码的注释。</p>';
/*
$updated_orders = $wpdb->update(
    $orders_table,
    array('currency' => 'USD'),
    array('currency' => 'CNY'),
    array('%s'),
    array('%s')
);
if ($updated_orders === false) {
    echo '<p style="color:red;">❌ 更新订单失败：' . $wpdb->last_error . '</p>';
} else {
    echo '<p style="color:green;">✅ 成功更新 ' . $updated_orders . ' 个订单的货币为 USD</p>';
}
*/
echo '<p style="color:orange;">⚠️ 已跳过（代码已注释）</p>';

// 4. 更新礼品卡表结构
echo '<h2>4. 更新礼品卡表默认货币</h2>';
$giftcards_table = $wpdb->prefix . 'tanz_giftcards';
$result2 = $wpdb->query("ALTER TABLE `{$giftcards_table}` MODIFY COLUMN `currency` VARCHAR(16) NOT NULL DEFAULT 'USD'");
if ($result2 === false) {
    echo '<p style="color:red;">❌ 失败：' . $wpdb->last_error . '</p>';
} else {
    echo '<p style="color:green;">✅ 成功更新礼品卡表默认货币为 USD</p>';
}

// 5. 统计现有礼品卡的货币分布
echo '<h2>5. 现有礼品卡货币统计</h2>';
$giftcard_stats = $wpdb->get_results("SELECT currency, COUNT(*) as count FROM `{$giftcards_table}` GROUP BY currency");
if ($giftcard_stats) {
    echo '<ul>';
    foreach ($giftcard_stats as $stat) {
        echo '<li>' . esc_html($stat->currency) . ': ' . esc_html($stat->count) . ' 张礼品卡</li>';
    }
    echo '</ul>';
} else {
    echo '<p>没有找到礼品卡</p>';
}

// 6. 更新现有 CNY 礼品卡为 USD（可选，默认注释掉）
echo '<h2>6. 更新现有礼品卡货币</h2>';
echo '<p><strong>注意：</strong>此操作会将所有 CNY 礼品卡改为 USD。如需执行，请取消下面代码的注释。</p>';
/*
$updated_giftcards = $wpdb->update(
    $giftcards_table,
    array('currency' => 'USD'),
    array('currency' => 'CNY'),
    array('%s'),
    array('%s')
);
if ($updated_giftcards === false) {
    echo '<p style="color:red;">❌ 更新礼品卡失败：' . $wpdb->last_error . '</p>';
} else {
    echo '<p style="color:green;">✅ 成功更新 ' . $updated_giftcards . ' 张礼品卡的货币为 USD</p>';
}
*/
echo '<p style="color:orange;">⚠️ 已跳过（代码已注释）</p>';

// 7. 优惠券说明
echo '<h2>7. 优惠券说明</h2>';
echo '<p>✅ 优惠券表 (wp_tanz_coupons) 没有货币字段</p>';
echo '<p>优惠券的金额会应用到订单上，订单的货币字段决定了实际使用的货币单位</p>';

echo '<hr>';
echo '<h2>✅ 迁移完成</h2>';
echo '<p>完成时间：' . date('Y-m-d H:i:s') . '</p>';
echo '<p><strong>重要提示：</strong></p>';
echo '<ul>';
echo '<li>新创建的订单将默认使用 USD 货币</li>';
echo '<li>新创建的礼品卡将默认使用 USD 货币</li>';
echo '<li>如需更新现有数据，请编辑此文件并取消相关代码的注释</li>';
echo '<li>建议在执行数据更新前先备份数据库</li>';
echo '</ul>';

// 安全提示
echo '<hr>';
echo '<p style="color:red;"><strong>⚠️ 安全提示：</strong>迁移完成后，请删除此文件或将其移出 web 可访问目录！</p>';
