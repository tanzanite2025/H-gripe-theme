<?php
/**
 * Plugin Name: Tanzanite Customer Service
 * Plugin URI: https://tanzanite.site
 * Description: 客服管理插件 - 管理客服信息并提供 REST API
 * Version: 1.2.0
 * Author: Tanzanite
 * Text Domain: tanzanite-cs
 * Domain Path: /languages
 * Requires at least: 5.8
 * Requires PHP: 8.0
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

// 定义插件常量
define( 'TZ_CS_VERSION', '1.2.0' );
define( 'TZ_CS_DB_VERSION', '1.2.0' );
define( 'TZ_CS_PLUGIN_FILE', __FILE__ );
define( 'TZ_CS_PLUGIN_DIR', plugin_dir_path( __FILE__ ) );
define( 'TZ_CS_PLUGIN_URL', plugin_dir_url( __FILE__ ) );

// 引入数据库类
require_once TZ_CS_PLUGIN_DIR . 'includes/class-database.php';

// 引入客服认证类
require_once TZ_CS_PLUGIN_DIR . 'includes/class-agent-auth.php';

// 引入客服端API类
require_once TZ_CS_PLUGIN_DIR . 'api/class-agent-api.php';

// 引入自动回复API类
require_once TZ_CS_PLUGIN_DIR . 'api/class-auto-reply-api.php';

/**
 * 客服管理插件主类
 */
class Tanzanite_Customer_Service_Plugin {
    
    /**
     * 单例实例
     */
    private static $instance = null;
    
    /**
     * 获取单例实例
     */
    public static function instance() {
        if ( null === self::$instance ) {
            self::$instance = new self();
        }
        return self::$instance;
    }
    
    /**
     * 构造函数
     */
    private function __construct() {
        $this->init_hooks();
    }
    
    /**
     * 初始化钩子
     */
    private function init_hooks(): void {
        add_action( 'admin_menu', [ $this, 'register_admin_menu' ] );
        add_action( 'rest_api_init', [ $this, 'register_rest_routes' ] );
        add_action( 'admin_enqueue_scripts', [ $this, 'enqueue_admin_scripts' ] );
    }
    
    /**
     * 加载管理页面脚本和样式
     */
    public function enqueue_admin_scripts( $hook ): void {
        // 在客服管理、聊天记录和自动回复设置页面加载
        if ( 'toplevel_page_tanzanite-customer-service' !== $hook && 
             'customer-service_page_tanzanite-cs-chat-records' !== $hook &&
             'customer-service_page_tanzanite-cs-auto-reply' !== $hook ) {
            return;
        }
        
        // 加载媒体库
        wp_enqueue_media();
        
        // 内联样式
        wp_add_inline_style( 'wp-admin', $this->get_admin_css() );
    }
    
    /**
     * 获取管理页面 CSS
     */
    private function get_admin_css(): string {
        return '
            .tz-cs-admin .tz-settings-wrapper {
                box-sizing: border-box;
                max-width: 1500px;
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
            
            .tz-cs-admin .tz-settings-header {
                display: flex;
                flex-direction: column;
                gap: 12px;
                align-items: center;
                text-align: center;
            }
            
            .tz-cs-admin .tz-settings-header h1 {
                margin: 0;
                font-size: 14px;
                font-weight: 700;
                letter-spacing: 0.02em;
                text-transform: uppercase;
                color: #111827;
            }
            
            .tz-cs-admin .tz-settings-header p {
                margin: 0;
                max-width: 960px;
                color: #4b5563;
            }
            
            .tz-cs-admin .tz-settings-section {
                background: #f8fafc;
                border-radius: 8px;
                padding: 16px;
                display: flex;
                flex-direction: column;
                gap: 12px;
                border: 1px solid #e5e7eb;
            }
            
            .tz-cs-admin .wp-list-table {
                border: 1px solid #e5e7eb;
                border-radius: 8px;
                background: #fff;
            }
            
            .tz-cs-admin .wp-list-table thead th {
                background: #f3f4f6;
                color: #111827;
                font-weight: 600;
                font-size: 14px;
                border-bottom: 1px solid #e5e7eb;
            }
            
            .tz-cs-admin .wp-list-table tbody tr {
                border-bottom: 1px solid #f3f4f6;
            }
            
            .tz-cs-admin .wp-list-table tbody tr:last-child {
                border-bottom: none;
            }
            
            .tz-cs-admin .button {
                background: #1f2937;
                border-color: #1f2937;
                color: #fff;
                border-radius: 6px;
                padding: 8px 16px;
                font-weight: 500;
                transition: all 0.2s;
            }
            
            .tz-cs-admin .button:hover {
                background: #111827;
                border-color: #111827;
            }
            
            .tz-cs-admin .button-primary {
                background: #1f2937;
                border-color: #1f2937;
            }
            
            .tz-cs-admin .button-primary:hover {
                background: #111827;
                border-color: #111827;
            }
            
            .tz-cs-admin input[type="text"],
            .tz-cs-admin input[type="email"],
            .tz-cs-admin input[type="url"],
            .tz-cs-admin input[type="number"],
            .tz-cs-admin select {
                border: 1px solid #e5e7eb;
                border-radius: 6px;
                padding: 8px 12px;
                font-size: 14px;
            }
            
            .tz-cs-admin input[type="text"]:focus,
            .tz-cs-admin input[type="email"]:focus,
            .tz-cs-admin input[type="url"]:focus,
            .tz-cs-admin input[type="number"]:focus,
            .tz-cs-admin select:focus {
                border-color: #1f2937;
                box-shadow: 0 0 0 1px #1f2937;
            }
            
            .tz-cs-admin .notice {
                border-radius: 6px;
                border-left-width: 4px;
            }
            
            .tz-cs-admin .card {
                border: 1px solid #e5e7eb;
                border-radius: 8px;
                background: #f8fafc;
                padding: 16px;
            }
            
            .tz-cs-admin .card h2 {
                margin-top: 0;
                font-size: 16px;
                font-weight: 600;
                color: #111827;
            }
            
            .tz-cs-admin .card code {
                display: block;
                padding: 12px;
                background: #fff;
                border: 1px solid #e5e7eb;
                border-radius: 6px;
                font-family: "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas, "Courier New", monospace;
                font-size: 13px;
                color: #1f2937;
                overflow-x: auto;
            }
            
            .tz-cs-admin .avatar-upload-wrapper {
                display: flex;
                align-items: center;
                gap: 12px;
            }
            
            .tz-cs-admin .avatar-preview {
                width: 50px;
                height: 50px;
                border-radius: 50%;
                object-fit: cover;
                border: 2px solid #e5e7eb;
                background: #f3f4f6;
            }
            
            .tz-cs-admin .avatar-preview.placeholder {
                display: flex;
                align-items: center;
                justify-content: center;
                color: #9ca3af;
                font-size: 12px;
            }
            
            .tz-cs-admin .upload-avatar-btn {
                padding: 6px 12px;
                font-size: 13px;
            }
            
            @media (max-width: 1600px) {
                .tz-cs-admin .tz-settings-wrapper {
                    width: 90%;
                }
            }
            
            @media (max-width: 1024px) {
                .tz-cs-admin .tz-settings-wrapper {
                    width: calc(100% - 48px);
                    margin: 24px auto;
                    padding: 24px;
                }
            }
        ';
    }
    
    /**
     * 注册管理菜单
     */
    public function register_admin_menu(): void {
        add_menu_page(
            __( 'Customer Service', 'tanzanite-cs' ),
            __( 'Customer Service', 'tanzanite-cs' ),
            'manage_options',
            'tanzanite-customer-service',
            [ $this, 'render_admin_page' ],
            'dashicons-businessman',
            30
        );
        
        // 添加子菜单：聊天记录
        add_submenu_page(
            'tanzanite-customer-service',
            __( 'Chat Records', 'tanzanite-cs' ),
            __( '聊天记录', 'tanzanite-cs' ),
            'manage_options',
            'tanzanite-cs-chat-records',
            [ $this, 'render_chat_records_page' ]
        );
        
        // 添加子菜单：自动回复设置
        add_submenu_page(
            'tanzanite-customer-service',
            __( 'Auto Reply Settings', 'tanzanite-cs' ),
            __( '自动回复设置', 'tanzanite-cs' ),
            'manage_options',
            'tanzanite-cs-auto-reply',
            [ $this, 'render_auto_reply_page' ]
        );
    }
    
    /**
     * 注册 REST API 路由
     */
    public function register_rest_routes(): void {
        // 获取客服列表
        register_rest_route( 'tanzanite/v1', '/customer-service/agents', [
            'methods'             => 'GET',
            'callback'            => [ $this, 'rest_get_agents' ],
            'permission_callback' => '__return_true',
        ] );
        
        // 发送消息
        register_rest_route( 'tanzanite/v1', '/customer-service/messages', [
            'methods'             => 'POST',
            'callback'            => [ $this, 'rest_send_message' ],
            'permission_callback' => '__return_true',
        ] );
        
        // 获取消息列表
        register_rest_route( 'tanzanite/v1', '/customer-service/messages/(?P<conversation_id>[a-zA-Z0-9_-]+)', [
            'methods'             => 'GET',
            'callback'            => [ $this, 'rest_get_messages' ],
            'permission_callback' => '__return_true',
        ] );
        
        // 获取会话列表（客服端）
        register_rest_route( 'tanzanite/v1', '/customer-service/conversations', [
            'methods'             => 'GET',
            'callback'            => [ $this, 'rest_get_conversations' ],
            'permission_callback' => [ $this, 'check_agent_permission' ],
        ] );
        
        // 注册客服端API路由
        TZ_CS_Agent_API::register_routes();
        
        // 注册自动回复API路由
        TZ_CS_Auto_Reply_API::register_routes();
    }
    
    /**
     * REST API: 获取客服列表（访客端）
     */
    public function rest_get_agents( \WP_REST_Request $request ): \WP_REST_Response {
        global $wpdb;
        $table = $wpdb->prefix . 'tz_cs_agents';
        
        // 只返回启用的客服
        $agents = $wpdb->get_results(
            "SELECT agent_id, name, email, avatar, whatsapp FROM $table WHERE status = 'active' ORDER BY created_at ASC"
        );
        
        // 格式化输出
        $formatted = array_map( fn( $agent ) => [
            'id'       => $agent->agent_id,
            'name'     => $agent->name,
            'email'    => $agent->email,
            'avatar'   => $agent->avatar,
            'whatsapp' => $agent->whatsapp,
        ], $agents );
        
        return new \WP_REST_Response( [
            'success' => true,
            'data'    => $formatted,
        ], 200 );
    }
    
    /**
     * REST API: 发送消息
     */
    public function rest_send_message( \WP_REST_Request $request ): \WP_REST_Response {
        global $wpdb;
        
        $conversation_id = sanitize_text_field( $request->get_param( 'conversation_id' ) );
        $message = sanitize_textarea_field( $request->get_param( 'message' ) );
        $sender_type = sanitize_text_field( $request->get_param( 'sender_type' ) ); // visitor, user, agent
        $sender_name = sanitize_text_field( $request->get_param( 'sender_name' ) );
        $sender_email = sanitize_email( $request->get_param( 'sender_email' ) );
        $agent_id = sanitize_text_field( $request->get_param( 'agent_id' ) );
        $message_type = sanitize_text_field( $request->get_param( 'message_type' ) ?: 'text' );
        $metadata = $request->get_param( 'metadata' );
        
        if ( empty( $conversation_id ) || empty( $message ) || empty( $sender_name ) ) {
            return new \WP_REST_Response( [
                'success' => false,
                'message' => '缺少必需参数',
            ], 400 );
        }
        
        $table_messages = $wpdb->prefix . 'tz_cs_messages';
        $table_conversations = $wpdb->prefix . 'tz_cs_conversations';
        
        // 插入消息
        $result = $wpdb->insert(
            $table_messages,
            [
                'conversation_id' => $conversation_id,
                'sender_type' => $sender_type,
                'sender_id' => get_current_user_id(),
                'sender_name' => $sender_name,
                'sender_email' => $sender_email,
                'agent_id' => $agent_id,
                'message_type' => $message_type,
                'message' => $message,
                'metadata' => $metadata ? wp_json_encode( $metadata ) : null,
                'created_at' => current_time( 'mysql' ),
            ],
            [ '%s', '%s', '%d', '%s', '%s', '%s', '%s', '%s', '%s', '%s' ]
        );
        
        if ( false === $result ) {
            return new \WP_REST_Response( [
                'success' => false,
                'message' => '消息保存失败',
            ], 500 );
        }
        
        $message_id = $wpdb->insert_id;
        
        // 更新或创建会话记录
        $wpdb->replace(
            $table_conversations,
            [
                'id' => $conversation_id,
                'agent_id' => $agent_id,
                'visitor_name' => $sender_type !== 'agent' ? $sender_name : '',
                'visitor_email' => $sender_type !== 'agent' ? $sender_email : '',
                'user_id' => $sender_type === 'user' ? get_current_user_id() : 0,
                'last_message' => $message,
                'last_message_time' => current_time( 'mysql' ),
                'status' => 'active',
                'updated_at' => current_time( 'mysql' ),
            ],
            [ '%s', '%s', '%s', '%s', '%d', '%s', '%s', '%s', '%s' ]
        );
        
        return new \WP_REST_Response( [
            'success' => true,
            'message_id' => $message_id,
            'data' => [
                'id' => $message_id,
                'conversation_id' => $conversation_id,
                'sender_name' => $sender_name,
                'message' => $message,
                'created_at' => current_time( 'mysql' ),
            ],
        ], 200 );
    }
    
    /**
     * REST API: 获取消息列表
     */
    public function rest_get_messages( \WP_REST_Request $request ): \WP_REST_Response {
        global $wpdb;
        
        $conversation_id = $request->get_param( 'conversation_id' );
        $limit = intval( $request->get_param( 'limit' ) ?: 50 );
        $offset = intval( $request->get_param( 'offset' ) ?: 0 );
        
        if ( empty( $conversation_id ) ) {
            return new \WP_REST_Response( [
                'success' => false,
                'message' => '缺少会话ID',
            ], 400 );
        }
        
        $table_messages = $wpdb->prefix . 'tz_cs_messages';
        
        $messages = $wpdb->get_results( $wpdb->prepare(
            "SELECT * FROM $table_messages 
            WHERE conversation_id = %s 
            ORDER BY created_at ASC 
            LIMIT %d OFFSET %d",
            $conversation_id,
            $limit,
            $offset
        ) );
        
        // 解析 metadata JSON
        foreach ( $messages as $message ) {
            if ( ! empty( $message->metadata ) ) {
                $message->metadata = json_decode( $message->metadata, true );
            }
        }
        
        return new \WP_REST_Response( [
            'success' => true,
            'data' => $messages,
            'total' => count( $messages ),
        ], 200 );
    }
    
    /**
     * REST API: 获取会话列表（客服端）
     */
    public function rest_get_conversations( \WP_REST_Request $request ): \WP_REST_Response {
        global $wpdb;
        
        $status = sanitize_text_field( $request->get_param( 'status' ) ?: 'active' );
        $limit = intval( $request->get_param( 'limit' ) ?: 100 );
        
        $table_conversations = $wpdb->prefix . 'tz_cs_conversations';
        
        $conversations = $wpdb->get_results( $wpdb->prepare(
            "SELECT * FROM $table_conversations 
            WHERE status = %s 
            ORDER BY updated_at DESC 
            LIMIT %d",
            $status,
            $limit
        ) );
        
        return new \WP_REST_Response( [
            'success' => true,
            'data' => $conversations,
            'total' => count( $conversations ),
        ], 200 );
    }
    
    /**
     * 检查客服权限
     */
    public function check_agent_permission(): bool {
        return current_user_can( 'manage_options' );
    }
    
    /**
     * 渲染管理页面
     */
    public function render_admin_page(): void {
        // 处理手动创建数据库表
        if ( isset( $_GET['create_tables'] ) && $_GET['create_tables'] === '1' ) {
            TZ_CS_Database::create_tables();
            wp_redirect( admin_url( 'admin.php?page=tanzanite-customer-service&tables_created=1' ) );
            exit;
        }
        
        // 处理手动创建数据库表
        if ( isset( $_POST['tz_cs_create_tables'] ) && check_admin_referer( 'tz_cs_create_tables' ) ) {
            global $wpdb;
            
            // 显示 SQL 错误
            $wpdb->show_errors();
            
            TZ_CS_Database::create_tables();
            update_option( 'tz_cs_db_version', TZ_CS_DB_VERSION );
            
            // 检查表是否真的创建成功
            $table = $wpdb->prefix . 'tz_cs_agents';
            $table_exists = $wpdb->get_var( "SHOW TABLES LIKE '$table'" ) === $table;
            
            if ( $table_exists ) {
                echo '<div class="notice notice-success"><p>✅ 数据库表创建成功！请刷新页面。</p></div>';
            } else {
                echo '<div class="notice notice-error"><p>❌ 表创建失败！</p>';
                if ( $wpdb->last_error ) {
                    echo '<p>错误信息：' . esc_html( $wpdb->last_error ) . '</p>';
                }
                echo '</div>';
            }
            
            $wpdb->hide_errors();
        }
        
        // 处理添加新客服
        if ( isset( $_POST['tz_cs_add_agent'] ) && check_admin_referer( 'tz_cs_add_agent' ) ) {
            global $wpdb;
            $table = $wpdb->prefix . 'tz_cs_agents';
            
            $agent_id = sanitize_text_field( $_POST['agent_id'] );
            $name     = sanitize_text_field( $_POST['name'] );
            $email    = sanitize_email( $_POST['email'] );
            $password = $_POST['password'];
            $avatar   = esc_url_raw( $_POST['avatar'] );
            $whatsapp = sanitize_text_field( $_POST['whatsapp'] );
            
            // 检查工号是否已存在
            $exists = $wpdb->get_var( $wpdb->prepare(
                "SELECT COUNT(*) FROM $table WHERE agent_id = %s",
                $agent_id
            ) );
            
            if ( $exists ) {
                echo '<div class="notice notice-error"><p>客服工号已存在！</p></div>';
            } else {
                // 插入新客服
                $result = $wpdb->insert(
                    $table,
                    [
                        'agent_id'   => $agent_id,
                        'name'       => $name,
                        'email'      => $email,
                        'password'   => password_hash( $password, PASSWORD_BCRYPT ),
                        'avatar'     => $avatar,
                        'whatsapp'   => $whatsapp,
                        'status'     => 'active',
                        'created_at' => current_time( 'mysql' ),
                    ],
                    [ '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s' ]
                );
                
                if ( $result ) {
                    echo '<div class="notice notice-success"><p>客服创建成功！工号：' . esc_html( $agent_id ) . '</p></div>';
                } else {
                    echo '<div class="notice notice-error"><p>创建失败：' . esc_html( $wpdb->last_error ) . '</p></div>';
                }
            }
        }
        
        // 处理更新客服状态
        if ( isset( $_POST['tz_cs_update_status'] ) && check_admin_referer( 'tz_cs_update_status' ) ) {
            global $wpdb;
            $table = $wpdb->prefix . 'tz_cs_agents';
            
            $agent_id = sanitize_text_field( $_POST['agent_id'] );
            $status   = sanitize_text_field( $_POST['status'] );
            
            $wpdb->update(
                $table,
                [ 'status' => $status ],
                [ 'agent_id' => $agent_id ],
                [ '%s' ],
                [ '%s' ]
            );
            
            echo '<div class="notice notice-success"><p>客服状态已更新！</p></div>';
        }
        
        // 处理重置密码
        if ( isset( $_POST['tz_cs_reset_password'] ) && check_admin_referer( 'tz_cs_reset_password' ) ) {
            global $wpdb;
            $table = $wpdb->prefix . 'tz_cs_agents';
            
            $agent_id     = sanitize_text_field( $_POST['agent_id'] );
            $new_password = $_POST['new_password'];
            
            $wpdb->update(
                $table,
                [ 'password' => password_hash( $new_password, PASSWORD_BCRYPT ) ],
                [ 'agent_id' => $agent_id ],
                [ '%s' ],
                [ '%s' ]
            );
            
            echo '<div class="notice notice-success"><p>密码已重置！</p></div>';
        }
        
        // 检查表是否存在
        global $wpdb;
        $table = $wpdb->prefix . 'tz_cs_agents';
        $table_exists = $wpdb->get_var( "SHOW TABLES LIKE '$table'" ) === $table;
        
        // 如果表不存在，显示创建按钮
        if ( ! $table_exists ) {
            ?>
            <div class="wrap">
                <h1>Customer Service Management</h1>
                <div class="notice notice-warning" style="padding: 20px; margin: 20px 0;">
                    <h2 style="margin-top: 0;">⚠️ 数据库表未创建</h2>
                    <p>检测到客服管理所需的数据库表尚未创建。请点击下方按钮创建数据库表。</p>
                    <form method="post" style="margin-top: 16px;">
                        <?php wp_nonce_field( 'tz_cs_create_tables' ); ?>
                        <button type="submit" name="tz_cs_create_tables" class="button button-primary button-large">
                            🔧 立即创建数据库表
                        </button>
                    </form>
                </div>
            </div>
            <?php
            return;
        }
        
        // 获取数据库中的客服列表
        $agents = $wpdb->get_results( "SELECT * FROM $table ORDER BY created_at DESC" );
        
        ?>
        <div class="wrap tz-cs-admin">
            <div class="tz-settings-wrapper">
                <div class="tz-settings-header">
                    <h1><?php _e( 'Customer Service Management', 'tanzanite-cs' ); ?></h1>
                    <p><?php _e( '管理客服账号，客服可使用工号和密码登录移动端 App。', 'tanzanite-cs' ); ?></p>
                </div>
                
                <!-- 添加新客服表单 -->
                <div class="tz-settings-section" style="background: #f0f9ff; border: 1px solid #0ea5e9; border-radius: 8px; padding: 20px; margin-bottom: 24px;">
                    <h2 style="margin-top: 0;">添加新客服</h2>
                    <form method="post" id="tz-add-agent-form">
                        <?php wp_nonce_field( 'tz_cs_add_agent' ); ?>
                        
                        <table class="form-table">
                            <tr>
                                <th><label for="agent_id">客服工号 *</label></th>
                                <td><input type="text" name="agent_id" id="agent_id" class="regular-text" required placeholder="例如：CS001"></td>
                            </tr>
                            <tr>
                                <th><label for="name">客服名称 *</label></th>
                                <td><input type="text" name="name" id="name" class="regular-text" required placeholder="例如：张三"></td>
                            </tr>
                            <tr>
                                <th><label for="email">邮箱 *</label></th>
                                <td><input type="email" name="email" id="email" class="regular-text" required placeholder="agent@example.com"></td>
                            </tr>
                            <tr>
                                <th><label for="password">密码 *</label></th>
                                <td><input type="password" name="password" id="password" class="regular-text" required placeholder="至少 8 位" minlength="8"></td>
                            </tr>
                            <tr>
                                <th><label for="whatsapp">WhatsApp 号码</label></th>
                                <td>
                                    <input type="text" name="whatsapp" id="whatsapp" class="regular-text" placeholder="例如：+8613800138000">
                                    <p class="description">用于前端显示 WhatsApp 联系按钮，格式：+国家码+号码（如 +8613800138000）</p>
                                </td>
                            </tr>
                            <tr>
                                <th><label for="avatar">头像 URL</label></th>
                                <td>
                                    <input type="url" name="avatar" id="avatar" class="regular-text" placeholder="https://...">
                                    <button type="button" class="button" id="upload-avatar-btn">上传头像</button>
                                </td>
                            </tr>
                        </table>
                        
                        <p class="submit">
                            <button type="submit" name="tz_cs_add_agent" class="button button-primary">创建客服</button>
                        </p>
                    </form>
                </div>
                
                <!-- 现有客服列表 -->
                <div class="tz-settings-section">
                    <h2>现有客服</h2>
                
                <table class="wp-list-table widefat fixed striped">
                    <thead>
                        <tr>
                            <th>工号</th>
                            <th>名称</th>
                            <th>邮箱</th>
                            <th>WhatsApp</th>
                            <th>头像</th>
                            <th>状态</th>
                            <th>最后登录</th>
                            <th>创建时间</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if ( empty( $agents ) ) : ?>
                            <tr>
                                <td colspan="9" style="text-align: center; padding: 40px; color: #6b7280;">暂无客服，请使用上方表单添加新客服。</td>
                            </tr>
                        <?php else : ?>
                            <?php foreach ( $agents as $agent ) : ?>
                                <tr>
                                    <td><strong><?php echo esc_html( $agent->agent_id ); ?></strong></td>
                                    <td><?php echo esc_html( $agent->name ); ?></td>
                                    <td><?php echo esc_html( $agent->email ); ?></td>
                                    <td>
                                        <?php if ( ! empty( $agent->whatsapp ) ) : ?>
                                            <a href="https://wa.me/<?php echo esc_attr( str_replace( '+', '', $agent->whatsapp ) ); ?>" target="_blank" style="color: #25D366; text-decoration: none;">
                                                📱 <?php echo esc_html( $agent->whatsapp ); ?>
                                            </a>
                                        <?php else : ?>
                                            <span style="color: #9ca3af;">未设置</span>
                                        <?php endif; ?>
                                    </td>
                                    <td>
                                        <?php if ( ! empty( $agent->avatar ) ) : ?>
                                            <img src="<?php echo esc_url( $agent->avatar ); ?>" style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;" alt="Avatar">
                                        <?php else : ?>
                                            <span style="color: #9ca3af;">无</span>
                                        <?php endif; ?>
                                    </td>
                                    <td>
                                        <span style="display: inline-block; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 500; background: <?php echo $agent->status === 'active' ? '#10b981' : '#ef4444'; ?>; color: white;">
                                            <?php echo $agent->status === 'active' ? '启用' : '禁用'; ?>
                                        </span>
                                    </td>
                                    <td><?php echo $agent->last_login ? esc_html( $agent->last_login ) : '<span style="color: #9ca3af;">从未登录</span>'; ?></td>
                                    <td><?php echo esc_html( $agent->created_at ); ?></td>
                                    <td>
                                        <form method="post" style="display: inline-block; margin-right: 8px;">
                                            <?php wp_nonce_field( 'tz_cs_update_status' ); ?>
                                            <input type="hidden" name="agent_id" value="<?php echo esc_attr( $agent->agent_id ); ?>">
                                            <input type="hidden" name="status" value="<?php echo $agent->status === 'active' ? 'inactive' : 'active'; ?>">
                                            <button type="submit" name="tz_cs_update_status" class="button button-small">
                                                <?php echo $agent->status === 'active' ? '禁用' : '启用'; ?>
                                            </button>
                                        </form>
                                        <button type="button" class="button button-small reset-password-btn" data-agent-id="<?php echo esc_attr( $agent->agent_id ); ?>" data-agent-name="<?php echo esc_attr( $agent->name ); ?>">重置密码</button>
                                    </td>
                                </tr>
                            <?php endforeach; ?>
                        <?php endif; ?>
                    </tbody>
                </table>
                </div>
            </div>
        </div>
        
        <script>
        jQuery(document).ready(function($) {
            // 头像上传（添加新客服表单）
            $('#upload-avatar-btn').on('click', function(e) {
                e.preventDefault();
                
                const mediaUploader = wp.media({
                    title: '选择头像',
                    button: { text: '使用此图片' },
                    multiple: false,
                    library: { type: 'image' }
                });
                
                mediaUploader.on('select', function() {
                    const attachment = mediaUploader.state().get('selection').first().toJSON();
                    $('#avatar').val(attachment.url);
                });
                
                mediaUploader.open();
            });
            
            // 重置密码
            $('.reset-password-btn').on('click', function() {
                const agentId = $(this).data('agent-id');
                const agentName = $(this).data('agent-name');
                
                const newPassword = prompt('请输入 ' + agentName + ' 的新密码（至少 8 位）：');
                
                if (newPassword && newPassword.length >= 8) {
                    const form = $('<form method="post">' +
                        '<?php wp_nonce_field( "tz_cs_reset_password", "_wpnonce", true, false ); ?>' +
                        '<input type="hidden" name="agent_id" value="' + agentId + '">' +
                        '<input type="hidden" name="new_password" value="' + newPassword + '">' +
                        '<input type="hidden" name="tz_cs_reset_password" value="1">' +
                        '</form>');
                    $('body').append(form);
                    form.submit();
                } else if (newPassword !== null) {
                    alert('密码至少需要 8 位！');
                }
            });
        });
        </script>
        <?php
    }
    
    /**
     * 渲染聊天记录页面
     */
    public function render_chat_records_page(): void {
        global $wpdb;
        
        $table_conversations = $wpdb->prefix . 'tz_cs_conversations';
        $table_messages = $wpdb->prefix . 'tz_cs_messages';
        
        // 获取筛选参数
        $status = isset( $_GET['status'] ) ? sanitize_text_field( $_GET['status'] ) : 'all';
        $search = isset( $_GET['s'] ) ? sanitize_text_field( $_GET['s'] ) : '';
        
        // 构建查询
        $where = "WHERE 1=1";
        if ( $status !== 'all' ) {
            $where .= $wpdb->prepare( " AND status = %s", $status );
        }
        if ( ! empty( $search ) ) {
            $where .= $wpdb->prepare( " AND (visitor_name LIKE %s OR visitor_email LIKE %s OR id LIKE %s)", 
                '%' . $wpdb->esc_like( $search ) . '%',
                '%' . $wpdb->esc_like( $search ) . '%',
                '%' . $wpdb->esc_like( $search ) . '%'
            );
        }
        
        // 获取会话列表
        $conversations = $wpdb->get_results( 
            "SELECT * FROM $table_conversations $where ORDER BY updated_at DESC LIMIT 100"
        );
        
        // 获取每个会话的消息数量
        foreach ( $conversations as $conversation ) {
            $conversation->message_count = $wpdb->get_var( $wpdb->prepare(
                "SELECT COUNT(*) FROM $table_messages WHERE conversation_id = %s",
                $conversation->id
            ) );
        }
        
        ?>
        <div class="wrap tz-cs-admin">
            <div class="tz-settings-wrapper">
                <div class="tz-settings-header">
                    <h1><?php _e( '聊天记录管理', 'tanzanite-cs' ); ?></h1>
                    <p><?php _e( '查看和管理所有客服聊天记录。', 'tanzanite-cs' ); ?></p>
                </div>
                
                <!-- 筛选和搜索 -->
                <div class="tz-settings-section" style="margin-bottom: 20px;">
                    <form method="get" style="display: flex; gap: 12px; align-items: center;">
                        <input type="hidden" name="page" value="tanzanite-cs-chat-records">
                        
                        <select name="status" style="height: 32px; border-radius: 6px; border: 1px solid #e5e7eb;">
                            <option value="all" <?php selected( $status, 'all' ); ?>>全部状态</option>
                            <option value="active" <?php selected( $status, 'active' ); ?>>进行中</option>
                            <option value="closed" <?php selected( $status, 'closed' ); ?>>已关闭</option>
                        </select>
                        
                        <input type="search" name="s" value="<?php echo esc_attr( $search ); ?>" 
                               placeholder="搜索会话ID、访客姓名或邮箱..." 
                               style="width: 300px; height: 32px; border-radius: 6px; border: 1px solid #e5e7eb; padding: 0 10px;">
                        
                        <button type="submit" class="button button-primary">筛选</button>
                        <?php if ( $status !== 'all' || ! empty( $search ) ) : ?>
                            <a href="<?php echo admin_url( 'admin.php?page=tanzanite-cs-chat-records' ); ?>" class="button">重置</a>
                        <?php endif; ?>
                    </form>
                </div>
                
                <!-- 会话列表 -->
                <div class="tz-settings-section">
                    <table class="wp-list-table widefat fixed striped">
                        <thead>
                            <tr>
                                <th style="width: 40px;"></th>
                                <th style="width: 150px;">会话ID</th>
                                <th>访客信息</th>
                                <th>客服</th>
                                <th style="width: 80px;">消息数</th>
                                <th>最后消息</th>
                                <th style="width: 150px;">更新时间</th>
                                <th style="width: 80px;">状态</th>
                                <th style="width: 100px;">操作</th>
                            </tr>
                        </thead>
                        <tbody>
                            <?php if ( empty( $conversations ) ) : ?>
                                <tr>
                                    <td colspan="9" style="text-align: center; padding: 40px; color: #6b7280;">
                                        暂无聊天记录
                                    </td>
                                </tr>
                            <?php else : ?>
                                <?php foreach ( $conversations as $conv ) : ?>
                                    <tr>
                                        <td style="text-align: center;">
                                            <?php if ( $conv->has_auto_reply ) : ?>
                                                <span style="font-size: 18px;" title="触发了 <?php echo intval( $conv->auto_reply_count ); ?> 次自动回复<?php echo $conv->needs_human ? '，需要人工介入' : ''; ?>">🤖</span>
                                            <?php endif; ?>
                                        </td>
                                        <td><code><?php echo esc_html( $conv->id ); ?></code></td>
                                        <td>
                                            <strong><?php echo esc_html( $conv->visitor_name ?: '匿名' ); ?></strong><br>
                                            <small style="color: #6b7280;"><?php echo esc_html( $conv->visitor_email ); ?></small>
                                        </td>
                                        <td><?php echo esc_html( $conv->agent_id ?: '-' ); ?></td>
                                        <td>
                                            <?php echo intval( $conv->message_count ); ?>
                                            <?php if ( $conv->has_auto_reply ) : ?>
                                                <br><small style="color: #6b7280;">自动: <?php echo intval( $conv->auto_reply_count ); ?></small>
                                            <?php endif; ?>
                                        </td>
                                        <td style="max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                                            <?php echo esc_html( $conv->last_message ); ?>
                                        </td>
                                        <td><?php echo esc_html( $conv->updated_at ); ?></td>
                                        <td>
                                            <?php if ( $conv->needs_human ) : ?>
                                                <span style="color: #ef4444;">⚠️ 需人工</span>
                                            <?php elseif ( $conv->status === 'active' ) : ?>
                                                <span style="color: #10b981;">● 进行中</span>
                                            <?php else : ?>
                                                <span style="color: #6b7280;">● 已关闭</span>
                                            <?php endif; ?>
                                        </td>
                                        <td>
                                            <a href="<?php echo admin_url( 'admin.php?page=tanzanite-cs-chat-records&action=view&conversation_id=' . urlencode( $conv->id ) ); ?>" 
                                               class="button button-small">查看</a>
                                        </td>
                                    </tr>
                                <?php endforeach; ?>
                            <?php endif; ?>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
        <?php
    }
    
    /**
     * 渲染自动回复设置页面
     */
    public function render_auto_reply_page(): void {
        global $wpdb;
        
        $table_auto_replies = $wpdb->prefix . 'tz_cs_auto_replies';
        
        // 处理表单提交
        if ( isset( $_POST['tz_cs_save_auto_reply'] ) && check_admin_referer( 'tz_auto_reply_save' ) ) {
            $type = sanitize_text_field( $_POST['type'] ?? '' );
            $message = sanitize_textarea_field( $_POST['message'] ?? '' );
            $keyword = sanitize_text_field( $_POST['keyword'] ?? '' );
            $id = intval( $_POST['id'] ?? 0 );
            
            $data = [
                'type' => $type,
                'reply_message' => $message,
                'trigger_keyword' => $keyword,
                'updated_at' => current_time( 'mysql' ),
            ];
            
            if ( $id > 0 ) {
                // 更新
                $wpdb->update(
                    $table_auto_replies,
                    $data,
                    [ 'id' => $id ],
                    [ '%s', '%s', '%s', '%s' ],
                    [ '%d' ]
                );
                echo '<div class="notice notice-success"><p>自动回复规则已更新</p></div>';
            } else {
                // 新建
                $data['created_at'] = current_time( 'mysql' );
                $wpdb->insert( $table_auto_replies, $data );
                echo '<div class="notice notice-success"><p>自动回复规则已创建</p></div>';
            }
        }
        
        // 处理删除
        if ( isset( $_GET['action'] ) && $_GET['action'] === 'delete' && isset( $_GET['id'] ) ) {
            $id = intval( $_GET['id'] );
            $wpdb->delete( $table_auto_replies, [ 'id' => $id ], [ '%d' ] );
            echo '<div class="notice notice-success"><p>规则已删除</p></div>';
        }
        
        // 获取所有规则
        $welcome_rules = $wpdb->get_results( $wpdb->prepare(
            "SELECT * FROM $table_auto_replies WHERE type = %s ORDER BY created_at DESC",
            'welcome'
        ) );
        
        $keyword_rules = $wpdb->get_results( $wpdb->prepare(
            "SELECT * FROM $table_auto_replies WHERE type = %s ORDER BY priority DESC, created_at DESC",
            'keyword'
        ) );
        
        ?>
        <div class="wrap tz-cs-admin">
            <div class="tz-settings-wrapper">
                <div class="tz-settings-header">
                    <h1><?php _e( '自动回复设置', 'tanzanite-cs' ); ?></h1>
                    <p><?php _e( '配置欢迎语和关键词自动回复规则', 'tanzanite-cs' ); ?></p>
                </div>
                
                <!-- 欢迎语设置 -->
                <div class="tz-settings-section">
                    <h2 style="margin: 0 0 16px 0; font-size: 16px; font-weight: 600;">👋 欢迎语设置</h2>
                    <p style="margin: 0 0 12px 0; color: #6b7280; font-size: 13px;">
                        用户打开聊天窗口时自动发送（24小时内只发送一次）
                    </p>
                    
                    <?php if ( empty( $welcome_rules ) ) : ?>
                        <form method="post" style="background: #fff; padding: 16px; border-radius: 8px; border: 1px solid #e5e7eb;">
                            <?php wp_nonce_field( 'tz_auto_reply_save' ); ?>
                            <input type="hidden" name="type" value="welcome">
                            
                            <textarea name="message" rows="4" class="large-text" 
                                      placeholder="👋 您好！欢迎来到客服中心..." 
                                      required style="width: 100%; border-radius: 6px; border: 1px solid #e5e7eb; padding: 8px 12px;"></textarea>
                            
                            <p style="margin-top: 12px;">
                                <button type="submit" name="tz_cs_save_auto_reply" class="button button-primary">保存欢迎语</button>
                            </p>
                        </form>
                    <?php else : ?>
                        <?php foreach ( $welcome_rules as $rule ) : ?>
                            <div style="background: #fff; padding: 16px; border-radius: 8px; border: 1px solid #e5e7eb; margin-bottom: 12px;">
                                <div style="display: flex; justify-content: space-between; align-items: start; gap: 12px;">
                                    <div style="flex: 1; white-space: pre-wrap; font-size: 14px; color: #1f2937;">
                                        <?php echo esc_html( $rule->reply_message ); ?>
                                    </div>
                                    <a href="?page=tanzanite-cs-auto-reply&action=delete&id=<?php echo $rule->id; ?>" 
                                       class="button button-small"
                                       onclick="return confirm('确定要删除此欢迎语吗？')">删除</a>
                                </div>
                            </div>
                        <?php endforeach; ?>
                    <?php endif; ?>
                </div>
                
                <!-- 关键词自动回复 -->
                <div class="tz-settings-section">
                    <h2 style="margin: 0 0 16px 0; font-size: 16px; font-weight: 600;">🤖 关键词自动回复</h2>
                    <p style="margin: 0 0 12px 0; color: #6b7280; font-size: 13px;">
                        用户发送特定关键词时自动回复（精确匹配）
                    </p>
                    
                    <!-- 添加新规则表单 -->
                    <form method="post" style="background: #fff; padding: 16px; border-radius: 8px; border: 1px solid #e5e7eb; margin-bottom: 16px;">
                        <?php wp_nonce_field( 'tz_auto_reply_save' ); ?>
                        <input type="hidden" name="type" value="keyword">
                        
                        <div style="display: grid; grid-template-columns: 200px 1fr; gap: 12px; margin-bottom: 12px;">
                            <div>
                                <label style="display: block; margin-bottom: 4px; font-size: 13px; font-weight: 500;">触发关键词</label>
                                <input type="text" name="keyword" class="regular-text" 
                                       placeholder="订单" required 
                                       style="width: 100%; border-radius: 6px; border: 1px solid #e5e7eb; padding: 6px 10px;">
                            </div>
                            <div>
                                <label style="display: block; margin-bottom: 4px; font-size: 13px; font-weight: 500;">自动回复内容</label>
                                <textarea name="message" rows="2" class="large-text" 
                                          placeholder="📦 请点击右侧 'My Orders' 标签查看您的订单..." 
                                          required style="width: 100%; border-radius: 6px; border: 1px solid #e5e7eb; padding: 6px 10px;"></textarea>
                            </div>
                        </div>
                        
                        <button type="submit" name="tz_cs_save_auto_reply" class="button button-primary">添加规则</button>
                    </form>
                    
                    <!-- 现有规则列表 -->
                    <?php if ( ! empty( $keyword_rules ) ) : ?>
                        <div style="background: #fff; border-radius: 8px; border: 1px solid #e5e7eb;">
                            <table class="wp-list-table widefat fixed striped" style="border: none;">
                                <thead>
                                    <tr>
                                        <th style="width: 150px;">关键词</th>
                                        <th>回复内容</th>
                                        <th style="width: 100px;">操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <?php foreach ( $keyword_rules as $rule ) : ?>
                                        <tr>
                                            <td><strong><?php echo esc_html( $rule->trigger_keyword ); ?></strong></td>
                                            <td style="white-space: pre-wrap; font-size: 13px;"><?php echo esc_html( $rule->reply_message ); ?></td>
                                            <td>
                                                <a href="?page=tanzanite-cs-auto-reply&action=delete&id=<?php echo $rule->id; ?>" 
                                                   class="button button-small"
                                                   onclick="return confirm('确定要删除此规则吗？')">删除</a>
                                            </td>
                                        </tr>
                                    <?php endforeach; ?>
                                </tbody>
                            </table>
                        </div>
                    <?php else : ?>
                        <p style="text-align: center; color: #9ca3af; padding: 24px;">暂无关键词规则</p>
                    <?php endif; ?>
                </div>
            </div>
        </div>
        <?php
    }
}

// 插件激活钩子 - 创建数据库表
register_activation_hook( __FILE__, [ 'TZ_CS_Database', 'create_tables' ] );

// 初始化插件
add_action( 'plugins_loaded', function() {
    // 检查数据库版本，如果不匹配则更新表结构
    $installed_version = get_option( 'tz_cs_db_version', '0' );
    if ( version_compare( $installed_version, TZ_CS_DB_VERSION, '<' ) ) {
        TZ_CS_Database::create_tables();
        update_option( 'tz_cs_db_version', TZ_CS_DB_VERSION );
    }
    
    Tanzanite_Customer_Service_Plugin::instance();
} );
