<?php
/**
 * Cart Admin Page
 * 
 * @package Tanzanite_Setting
 */

if (!defined('ABSPATH')) {
    exit;
}

class Tanzanite_Cart_Admin {
    
    /**
     * 渲染购物车列表页面
     */
    public static function render_cart_list() {
        global $wpdb;
        
        // 处理操作
        if (isset($_POST['action']) && isset($_POST['cart_nonce'])) {
            if (!wp_verify_nonce($_POST['cart_nonce'], 'cart_action')) {
                wp_die('Security check failed');
            }
            
            $action = sanitize_text_field($_POST['action']);
            $cart_id = isset($_POST['cart_id']) ? intval($_POST['cart_id']) : 0;
            
            if ($action === 'delete' && $cart_id > 0) {
                $table = $wpdb->prefix . 'tanzanite_cart';
                $wpdb->delete($table, ['id' => $cart_id]);
                echo '<div class="notice notice-success"><p>购物车项已删除</p></div>';
            }
        }
        
        // 获取搜索和筛选参数
        $search = isset($_GET['s']) ? sanitize_text_field($_GET['s']) : '';
        $user_type = isset($_GET['user_type']) ? sanitize_text_field($_GET['user_type']) : 'all';
        
        // 分页参数
        $per_page = 20;
        $current_page = isset($_GET['paged']) ? max(1, intval($_GET['paged'])) : 1;
        $offset = ($current_page - 1) * $per_page;
        
        // 构建查询
        $table = $wpdb->prefix . 'tanzanite_cart';
        $where = ['1=1'];
        
        // 搜索条件
        if (!empty($search)) {
            $search_like = '%' . $wpdb->esc_like($search) . '%';
            $where[] = $wpdb->prepare(
                "(c.product_title LIKE %s OR c.sku LIKE %s OR u.user_login LIKE %s OR u.user_email LIKE %s)",
                $search_like, $search_like, $search_like, $search_like
            );
        }
        
        // 用户类型筛选
        if ($user_type === 'member') {
            $where[] = "c.user_id > 0";
        } elseif ($user_type === 'guest') {
            $where[] = "c.user_id = 0";
        }
        
        $where_sql = implode(' AND ', $where);
        
        // 获取总数
        $total = $wpdb->get_var("
            SELECT COUNT(*) 
            FROM $table c 
            LEFT JOIN {$wpdb->users} u ON c.user_id = u.ID 
            WHERE $where_sql
        ");
        
        // 获取购物车数据
        $carts = $wpdb->get_results("
            SELECT c.*, u.user_login, u.user_email 
            FROM $table c 
            LEFT JOIN {$wpdb->users} u ON c.user_id = u.ID 
            WHERE $where_sql
            ORDER BY c.updated_at DESC 
            LIMIT $per_page OFFSET $offset
        ");
        
        // 计算总页数
        $total_pages = ceil($total / $per_page);
        
        ?>
        <style>
            body.tz-settings-admin .tz-settings-wrapper {
                box-sizing: border-box;
                max-width: 1500px;
                min-height: 1500px;
                margin: 40px auto;
                padding: 32px;
                background: #fff;
                border-radius: 8px;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
                display: flex;
                flex-direction: column;
                gap: 24px;
                font-size: 14px;
                color: #1f2937;
            }
            body.tz-settings-admin .tz-settings-header {
                display: flex;
                flex-direction: column;
                gap: 12px;
                align-items: center;
                text-align: center;
            }
            body.tz-settings-admin .tz-settings-header h1 {
                margin: 0;
                font-size: 14px;
                font-weight: 700;
                letter-spacing: 0.02em;
                text-transform: uppercase;
                color: #111827;
            }
            body.tz-settings-admin .tz-settings-header p {
                margin: 0;
                max-width: 960px;
                color: #4b5563;
            }
            body.tz-settings-admin .tz-settings-section {
                background: #f8fafc;
                border-radius: 8px;
                padding: 16px;
                display: flex;
                flex-direction: column;
                gap: 12px;
                border: 1px solid #e5e7eb;
            }
            body.tz-settings-admin .button-primary {
                background: #1f2937;
                border-color: #1f2937;
                color: #fff;
                height: 30px;
                padding: 0 12px;
                font-size: 14px;
                font-weight: 600;
                border-radius: 8px;
                cursor: pointer;
            }
            body.tz-settings-admin .button-primary:hover {
                background: #111827;
                border-color: #111827;
            }
            body.tz-settings-admin .button {
                background: #f8fafc;
                color: #1d1d1f;
                border: 1px solid #e5e7eb;
                height: 30px;
                padding: 0 12px;
                font-size: 14px;
                font-weight: 600;
                border-radius: 8px;
                cursor: pointer;
            }
            body.tz-settings-admin .button:hover {
                background: #ffffff;
                border-color: #cfe1f9;
            }
        </style>
        <div class="tz-settings-wrapper">
            <div class="tz-settings-header">
                <h1>🛒 购物车管理</h1>
                <p>查看和管理所有用户的购物车数据</p>
            </div>
            
            <!-- 搜索和筛选表单 -->
            <div class="tz-settings-section">
                <form method="get" style="display: flex; gap: 10px; align-items: center; flex-wrap: wrap;">
                    <input type="hidden" name="page" value="tanzanite-cart-list">
                    <input type="search" name="s" value="<?php echo esc_attr($search); ?>" 
                           placeholder="搜索商品、SKU、用户名或邮箱..." 
                           style="width: 300px;">
                    
                    <select name="user_type">
                        <option value="all" <?php selected($user_type, 'all'); ?>>全部用户</option>
                        <option value="member" <?php selected($user_type, 'member'); ?>>会员</option>
                        <option value="guest" <?php selected($user_type, 'guest'); ?>>游客</option>
                    </select>
                    
                    <button type="submit" class="button-primary">搜索</button>
                    
                    <?php if (!empty($search) || $user_type !== 'all'): ?>
                        <a href="<?php echo admin_url('admin.php?page=tanzanite-cart-list'); ?>" class="button">
                            清除筛选
                        </a>
                    <?php endif; ?>
                </form>
            </div>
            
            <div class="tablenav top">
                <div class="alignleft actions">
                    <span class="displaying-num">共 <?php echo number_format($total); ?> 项，显示第 <?php echo number_format($offset + 1); ?>-<?php echo number_format(min($offset + $per_page, $total)); ?> 项</span>
                </div>
                
                <?php if ($total_pages > 1): ?>
                    <div class="tablenav-pages">
                        <?php
                        $page_links = paginate_links([
                            'base' => add_query_arg('paged', '%#%'),
                            'format' => '',
                            'prev_text' => '&laquo; 上一页',
                            'next_text' => '下一页 &raquo;',
                            'total' => $total_pages,
                            'current' => $current_page,
                            'add_args' => [
                                's' => $search,
                                'user_type' => $user_type
                            ]
                        ]);
                        echo $page_links;
                        ?>
                    </div>
                <?php endif; ?>
            </div>
            
            <table class="wp-list-table widefat fixed striped">
                <thead>
                    <tr>
                        <th style="width: 50px;">ID</th>
                        <th style="width: 120px;">用户</th>
                        <th>商品</th>
                        <th style="width: 80px;">价格</th>
                        <th style="width: 60px;">数量</th>
                        <th style="width: 80px;">小计</th>
                        <th style="width: 150px;">更新时间</th>
                        <th style="width: 100px;">操作</th>
                    </tr>
                </thead>
                <tbody>
                    <?php if (empty($carts)): ?>
                        <tr>
                            <td colspan="8" style="text-align: center; padding: 40px;">
                                <p style="color: #999;">暂无购物车数据</p>
                            </td>
                        </tr>
                    <?php else: ?>
                        <?php foreach ($carts as $cart): ?>
                            <tr>
                                <td><?php echo esc_html($cart->id); ?></td>
                                <td>
                                    <?php if ($cart->user_id > 0): ?>
                                        <strong><?php echo esc_html($cart->user_login); ?></strong><br>
                                        <small><?php echo esc_html($cart->user_email); ?></small>
                                    <?php else: ?>
                                        <span style="color: #999;">游客</span><br>
                                        <small><?php echo esc_html(substr($cart->session_id, 0, 8)); ?>...</small>
                                    <?php endif; ?>
                                </td>
                                <td>
                                    <div style="display: flex; align-items: center; gap: 10px;">
                                        <?php if ($cart->thumbnail): ?>
                                            <img src="<?php echo esc_url($cart->thumbnail); ?>" 
                                                 style="width: 40px; height: 40px; object-fit: cover; border-radius: 4px;">
                                        <?php endif; ?>
                                        <div>
                                            <strong><?php echo esc_html($cart->product_title); ?></strong>
                                            <?php if ($cart->sku): ?>
                                                <br><small>SKU: <?php echo esc_html($cart->sku); ?></small>
                                            <?php endif; ?>
                                        </div>
                                    </div>
                                </td>
                                <td>$<?php echo number_format($cart->price, 2); ?></td>
                                <td><?php echo esc_html($cart->quantity); ?></td>
                                <td><strong>$<?php echo number_format($cart->price * $cart->quantity, 2); ?></strong></td>
                                <td><?php echo esc_html($cart->updated_at); ?></td>
                                <td>
                                    <form method="post" style="display: inline;">
                                        <?php wp_nonce_field('cart_action', 'cart_nonce'); ?>
                                        <input type="hidden" name="action" value="delete">
                                        <input type="hidden" name="cart_id" value="<?php echo esc_attr($cart->id); ?>">
                                        <button type="submit" class="button button-small" 
                                                onclick="return confirm('确定删除此购物车项？')">
                                            删除
                                        </button>
                                    </form>
                                </td>
                            </tr>
                        <?php endforeach; ?>
                    <?php endif; ?>
                </tbody>
            </table>
        </div>
        <?php
    }
}
