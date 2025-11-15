<?php
/**
 * 客服端 API 类
 * 
 * @package Tanzanite_Customer_Service
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class TZ_CS_Agent_API {
    
    /**
     * 注册 API 路由
     */
    public static function register_routes(): void {
        // 客服登录
        register_rest_route( 'tanzanite/v1', '/agent/login', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'login' ],
            'permission_callback' => '__return_true',
        ] );
        
        // 客服登出
        register_rest_route( 'tanzanite/v1', '/agent/logout', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'logout' ],
            'permission_callback' => '__return_true',
        ] );
        
        // 获取客服信息
        register_rest_route( 'tanzanite/v1', '/agent/me', [
            'methods'             => 'GET',
            'callback'            => [ __CLASS__, 'get_agent_info' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 获取会话列表
        register_rest_route( 'tanzanite/v1', '/agent/conversations', [
            'methods'             => 'GET',
            'callback'            => [ __CLASS__, 'get_conversations' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 获取会话消息
        register_rest_route( 'tanzanite/v1', '/agent/conversations/(?P<conversation_id>[a-zA-Z0-9_-]+)/messages', [
            'methods'             => 'GET',
            'callback'            => [ __CLASS__, 'get_messages' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 发送消息
        register_rest_route( 'tanzanite/v1', '/agent/messages', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'send_message' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 标记消息已读
        register_rest_route( 'tanzanite/v1', '/agent/messages/read', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'mark_as_read' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 转接会话
        register_rest_route( 'tanzanite/v1', '/agent/conversations/(?P<conversation_id>[a-zA-Z0-9_-]+)/transfer', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'transfer_conversation' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 获取转接历史
        register_rest_route( 'tanzanite/v1', '/agent/conversations/(?P<conversation_id>[a-zA-Z0-9_-]+)/transfer-history', [
            'methods'             => 'GET',
            'callback'            => [ __CLASS__, 'get_transfer_history' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 获取通知列表
        register_rest_route( 'tanzanite/v1', '/agent/notifications', [
            'methods'             => 'GET',
            'callback'            => [ __CLASS__, 'get_notifications' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 标记通知已读
        register_rest_route( 'tanzanite/v1', '/agent/notifications/(?P<notification_id>\d+)/read', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'mark_notification_read' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 标记所有通知已读
        register_rest_route( 'tanzanite/v1', '/agent/notifications/read-all', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'mark_all_notifications_read' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 更新客服状态
        register_rest_route( 'tanzanite/v1', '/agent/status', [
            'methods'             => 'POST',
            'callback'            => [ __CLASS__, 'update_agent_status' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
        
        // 获取在线客服列表
        register_rest_route( 'tanzanite/v1', '/agent/online-agents', [
            'methods'             => 'GET',
            'callback'            => [ __CLASS__, 'get_online_agents' ],
            'permission_callback' => [ __CLASS__, 'verify_agent_token' ],
        ] );
    }
    
    /**
     * 验证客服 Token
     */
    public static function verify_agent_token( WP_REST_Request $request ): bool {
        $token = $request->get_header( 'Authorization' );
        
        if ( ! $token ) {
            return false;
        }
        
        // 移除 "Bearer " 前缀
        $token = str_replace( 'Bearer ', '', $token );
        
        $agent = TZ_CS_Agent_Auth::verify_token( $token );
        
        if ( ! $agent ) {
            return false;
        }
        
        // 将客服信息存储到请求中
        $request->set_param( '_agent', $agent );
        
        return true;
    }
    
    /**
     * 客服登录
     */
    public static function login( WP_REST_Request $request ): WP_REST_Response {
        $agent_id    = $request->get_param( 'agent_id' );
        $password    = $request->get_param( 'password' );
        $device_info = $request->get_param( 'device_info' ) ?? '';
        
        if ( ! $agent_id || ! $password ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '客服工号和密码不能为空',
            ], 400 );
        }
        
        $result = TZ_CS_Agent_Auth::login( $agent_id, $password, $device_info );
        
        if ( is_wp_error( $result ) ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => $result->get_error_message(),
            ], 401 );
        }
        
        return new WP_REST_Response( $result, 200 );
    }
    
    /**
     * 客服登出
     */
    public static function logout( WP_REST_Request $request ): WP_REST_Response {
        $token = $request->get_header( 'Authorization' );
        $token = str_replace( 'Bearer ', '', $token );
        
        if ( $token ) {
            TZ_CS_Agent_Auth::logout( $token );
        }
        
        return new WP_REST_Response( [
            'success' => true,
            'message' => '登出成功',
        ], 200 );
    }
    
    /**
     * 获取客服信息
     */
    public static function get_agent_info( WP_REST_Request $request ): WP_REST_Response {
        $agent = $request->get_param( '_agent' );
        
        return new WP_REST_Response( [
            'success' => true,
            'agent'   => $agent,
        ], 200 );
    }
    
    /**
     * 获取会话列表
     */
    public static function get_conversations( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table = $wpdb->prefix . 'tz_cs_conversations';
        
        $status = $request->get_param( 'status' ) ?? 'all';
        $page   = max( 1, intval( $request->get_param( 'page' ) ?? 1 ) );
        $limit  = 20;
        $offset = ( $page - 1 ) * $limit;
        
        // 构建查询
        $where = "WHERE 1=1";
        if ( $status !== 'all' ) {
            $where .= $wpdb->prepare( " AND status = %s", $status );
        }
        
        // 获取会话列表
        $conversations = $wpdb->get_results( $wpdb->prepare(
            "SELECT * FROM $table $where ORDER BY updated_at DESC LIMIT %d OFFSET %d",
            $limit,
            $offset
        ) );
        
        // 获取总数
        $total = $wpdb->get_var( "SELECT COUNT(*) FROM $table $where" );
        
        return new WP_REST_Response( [
            'success'       => true,
            'conversations' => $conversations,
            'pagination'    => [
                'page'       => $page,
                'limit'      => $limit,
                'total'      => intval( $total ),
                'total_pages' => ceil( $total / $limit ),
            ],
        ], 200 );
    }
    
    /**
     * 获取会话消息
     */
    public static function get_messages( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table = $wpdb->prefix . 'tz_cs_messages';
        
        $conversation_id = $request->get_param( 'conversation_id' );
        $page            = max( 1, intval( $request->get_param( 'page' ) ?? 1 ) );
        $limit           = 50;
        $offset          = ( $page - 1 ) * $limit;
        
        // 获取消息列表
        $messages = $wpdb->get_results( $wpdb->prepare(
            "SELECT * FROM $table WHERE conversation_id = %s ORDER BY created_at ASC LIMIT %d OFFSET %d",
            $conversation_id,
            $limit,
            $offset
        ) );
        
        // 获取总数
        $total = $wpdb->get_var( $wpdb->prepare(
            "SELECT COUNT(*) FROM $table WHERE conversation_id = %s",
            $conversation_id
        ) );
        
        return new WP_REST_Response( [
            'success'    => true,
            'messages'   => $messages,
            'pagination' => [
                'page'        => $page,
                'limit'       => $limit,
                'total'       => intval( $total ),
                'total_pages' => ceil( $total / $limit ),
            ],
        ], 200 );
    }
    
    /**
     * 发送消息
     */
    public static function send_message( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_messages      = $wpdb->prefix . 'tz_cs_messages';
        $table_conversations = $wpdb->prefix . 'tz_cs_conversations';
        
        $agent           = $request->get_param( '_agent' );
        $conversation_id = $request->get_param( 'conversation_id' );
        $message         = $request->get_param( 'message' );
        $message_type    = $request->get_param( 'message_type' ) ?? 'text';
        
        if ( ! $conversation_id || ! $message ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '会话ID和消息内容不能为空',
            ], 400 );
        }
        
        // 插入消息
        $result = $wpdb->insert(
            $table_messages,
            [
                'conversation_id' => $conversation_id,
                'sender_type'     => 'agent',
                'sender_id'       => $agent['agent_id'],
                'message'         => $message,
                'message_type'    => $message_type,
                'is_read'         => 0,
                'created_at'      => current_time( 'mysql' ),
            ],
            [ '%s', '%s', '%s', '%s', '%s', '%d', '%s' ]
        );
        
        if ( ! $result ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '发送失败',
            ], 500 );
        }
        
        $message_id = $wpdb->insert_id;
        
        // 更新会话最后消息时间
        $wpdb->update(
            $table_conversations,
            [
                'last_message'    => $message,
                'last_message_at' => current_time( 'mysql' ),
                'updated_at'      => current_time( 'mysql' ),
            ],
            [ 'id' => $conversation_id ],
            [ '%s', '%s', '%s' ],
            [ '%s' ]
        );
        
        // 获取完整消息
        $full_message = $wpdb->get_row( $wpdb->prepare(
            "SELECT * FROM $table_messages WHERE id = %d",
            $message_id
        ) );
        
        return new WP_REST_Response( [
            'success' => true,
            'message' => $full_message,
        ], 200 );
    }
    
    /**
     * 标记消息已读
     */
    public static function mark_as_read( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table = $wpdb->prefix . 'tz_cs_messages';
        
        $conversation_id = $request->get_param( 'conversation_id' );
        
        if ( ! $conversation_id ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '会话ID不能为空',
            ], 400 );
        }
        
        // 标记访客发送的消息为已读
        $wpdb->update(
            $table,
            [ 'is_read' => 1 ],
            [
                'conversation_id' => $conversation_id,
                'sender_type'     => 'visitor',
            ],
            [ '%d' ],
            [ '%s', '%s' ]
        );
        
        return new WP_REST_Response( [
            'success' => true,
            'message' => '已标记为已读',
        ], 200 );
    }
    
    /**
     * 转接会话
     */
    public static function transfer_conversation( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_conversations  = $wpdb->prefix . 'tz_cs_conversations';
        $table_messages       = $wpdb->prefix . 'tz_cs_messages';
        $table_agents         = $wpdb->prefix . 'tz_cs_agents';
        $table_transfers      = $wpdb->prefix . 'tz_cs_transfers';
        $table_notifications  = $wpdb->prefix . 'tz_cs_notifications';
        
        $agent           = $request->get_param( '_agent' );
        $conversation_id = $request->get_param( 'conversation_id' );
        $to_agent_id     = $request->get_param( 'to_agent_id' );
        $note            = $request->get_param( 'note' ) ?? '';
        
        if ( ! $conversation_id || ! $to_agent_id ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '会话ID和目标客服ID不能为空',
            ], 400 );
        }
        
        // 检查目标客服是否存在且启用
        $target_agent = $wpdb->get_row( $wpdb->prepare(
            "SELECT * FROM $table_agents WHERE agent_id = %s AND status = 'active'",
            $to_agent_id
        ) );
        
        if ( ! $target_agent ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '目标客服不存在或已禁用',
            ], 400 );
        }
        
        // 检查是否转接给自己
        if ( $to_agent_id === $agent['agent_id'] ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '不能转接给自己',
            ], 400 );
        }
        
        // 更新会话
        $result = $wpdb->update(
            $table_conversations,
            [
                'agent_id'         => $to_agent_id,
                'transferred_from' => $agent['agent_id'],
                'transferred_at'   => current_time( 'mysql' ),
                'updated_at'       => current_time( 'mysql' ),
            ],
            [ 'id' => $conversation_id ],
            [ '%s', '%s', '%s', '%s' ],
            [ '%s' ]
        );
        
        if ( ! $result ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '转接失败',
            ], 500 );
        }
        
        // 插入系统消息
        $system_message = sprintf(
            '会话已从客服 %s 转接给客服 %s',
            $agent['name'],
            $target_agent->name
        );
        
        if ( ! empty( $note ) ) {
            $system_message .= "\n转接备注：" . $note;
        }
        
        $wpdb->insert(
            $table_messages,
            [
                'conversation_id' => $conversation_id,
                'sender_type'     => 'system',
                'sender_id'       => 0,
                'sender_name'     => 'System',
                'sender_email'    => '',
                'agent_id'        => '',
                'message_type'    => 'text',
                'message'         => $system_message,
                'created_at'      => current_time( 'mysql' ),
            ],
            [ '%s', '%s', '%d', '%s', '%s', '%s', '%s', '%s', '%s' ]
        );
        
        // 记录转接历史
        $wpdb->insert(
            $table_transfers,
            [
                'conversation_id'  => $conversation_id,
                'from_agent_id'    => $agent['agent_id'],
                'from_agent_name'  => $agent['name'],
                'to_agent_id'      => $to_agent_id,
                'to_agent_name'    => $target_agent->name,
                'reason'           => $note,
                'created_at'       => current_time( 'mysql' ),
            ],
            [ '%s', '%s', '%s', '%s', '%s', '%s', '%s' ]
        );
        
        // 创建通知给目标客服
        $notification_title = '新的会话转接';
        $notification_message = sprintf(
            '%s 将会话转接给了你',
            $agent['name']
        );
        if ( ! empty( $note ) ) {
            $notification_message .= "\n备注：" . $note;
        }
        
        $wpdb->insert(
            $table_notifications,
            [
                'agent_id'    => $to_agent_id,
                'type'        => 'transfer',
                'title'       => $notification_title,
                'message'     => $notification_message,
                'data'        => json_encode( [
                    'conversation_id' => $conversation_id,
                    'from_agent_id'   => $agent['agent_id'],
                    'from_agent_name' => $agent['name'],
                ] ),
                'is_read'     => 0,
                'created_at'  => current_time( 'mysql' ),
            ],
            [ '%s', '%s', '%s', '%s', '%s', '%d', '%s' ]
        );
        
        return new WP_REST_Response( [
            'success' => true,
            'message' => '转接成功',
            'data'    => [
                'from_agent' => $agent['name'],
                'to_agent'   => $target_agent->name,
            ],
        ], 200 );
    }
    
    /**
     * 获取转接历史
     */
    public static function get_transfer_history( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_transfers = $wpdb->prefix . 'tz_cs_transfers';
        
        $conversation_id = $request->get_param( 'conversation_id' );
        
        if ( ! $conversation_id ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '会话ID不能为空',
            ], 400 );
        }
        
        // 获取转接历史
        $history = $wpdb->get_results( $wpdb->prepare(
            "SELECT * FROM $table_transfers WHERE conversation_id = %s ORDER BY created_at DESC",
            $conversation_id
        ) );
        
        return new WP_REST_Response( [
            'success' => true,
            'data'    => $history,
        ], 200 );
    }
    
    /**
     * 获取通知列表
     */
    public static function get_notifications( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_notifications = $wpdb->prefix . 'tz_cs_notifications';
        
        $agent = $request->get_param( '_agent' );
        $unread_only = $request->get_param( 'unread_only' ) === 'true';
        
        // 构建查询
        $where = $wpdb->prepare( "agent_id = %s", $agent['agent_id'] );
        if ( $unread_only ) {
            $where .= " AND is_read = 0";
        }
        
        // 获取通知列表
        $notifications = $wpdb->get_results(
            "SELECT * FROM $table_notifications WHERE $where ORDER BY created_at DESC LIMIT 50"
        );
        
        // 解析 data 字段
        foreach ( $notifications as $notification ) {
            if ( ! empty( $notification->data ) ) {
                $notification->data = json_decode( $notification->data );
            }
        }
        
        // 获取未读数量
        $unread_count = $wpdb->get_var( $wpdb->prepare(
            "SELECT COUNT(*) FROM $table_notifications WHERE agent_id = %s AND is_read = 0",
            $agent['agent_id']
        ) );
        
        return new WP_REST_Response( [
            'success' => true,
            'data'    => [
                'notifications' => $notifications,
                'unread_count'  => (int) $unread_count,
            ],
        ], 200 );
    }
    
    /**
     * 标记通知已读
     */
    public static function mark_notification_read( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_notifications = $wpdb->prefix . 'tz_cs_notifications';
        
        $agent = $request->get_param( '_agent' );
        $notification_id = $request->get_param( 'notification_id' );
        
        // 验证通知是否属于当前客服
        $notification = $wpdb->get_row( $wpdb->prepare(
            "SELECT * FROM $table_notifications WHERE id = %d AND agent_id = %s",
            $notification_id,
            $agent['agent_id']
        ) );
        
        if ( ! $notification ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '通知不存在',
            ], 404 );
        }
        
        // 标记已读
        $wpdb->update(
            $table_notifications,
            [
                'is_read'  => 1,
                'read_at'  => current_time( 'mysql' ),
            ],
            [ 'id' => $notification_id ],
            [ '%d', '%s' ],
            [ '%d' ]
        );
        
        return new WP_REST_Response( [
            'success' => true,
            'message' => '已标记为已读',
        ], 200 );
    }
    
    /**
     * 标记所有通知已读
     */
    public static function mark_all_notifications_read( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_notifications = $wpdb->prefix . 'tz_cs_notifications';
        
        $agent = $request->get_param( '_agent' );
        
        // 标记所有未读通知为已读
        $wpdb->query( $wpdb->prepare(
            "UPDATE $table_notifications SET is_read = 1, read_at = %s WHERE agent_id = %s AND is_read = 0",
            current_time( 'mysql' ),
            $agent['agent_id']
        ) );
        
        return new WP_REST_Response( [
            'success' => true,
            'message' => '所有通知已标记为已读',
        ], 200 );
    }
    
    /**
     * 更新客服状态
     */
    public static function update_agent_status( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_agents = $wpdb->prefix . 'tz_cs_agents';
        
        $agent = $request->get_param( '_agent' );
        $status = $request->get_param( 'status' );
        
        // 验证状态值
        $valid_statuses = [ 'online', 'busy', 'away', 'offline' ];
        if ( ! in_array( $status, $valid_statuses ) ) {
            return new WP_REST_Response( [
                'success' => false,
                'message' => '无效的状态值',
            ], 400 );
        }
        
        // 更新状态
        $wpdb->update(
            $table_agents,
            [
                'status'          => $status,
                'last_active_at'  => current_time( 'mysql' ),
            ],
            [ 'agent_id' => $agent['agent_id'] ],
            [ '%s', '%s' ],
            [ '%s' ]
        );
        
        return new WP_REST_Response( [
            'success' => true,
            'message' => '状态更新成功',
            'data'    => [
                'status'         => $status,
                'last_active_at' => current_time( 'mysql' ),
            ],
        ], 200 );
    }
    
    /**
     * 获取在线客服列表
     */
    public static function get_online_agents( WP_REST_Request $request ): WP_REST_Response {
        global $wpdb;
        $table_agents = $wpdb->prefix . 'tz_cs_agents';
        
        // 获取所有在线客服（online, busy, away）
        $agents = $wpdb->get_results(
            "SELECT agent_id, name, email, avatar, status, last_active_at 
             FROM $table_agents 
             WHERE status IN ('online', 'busy', 'away') 
             ORDER BY status ASC, last_active_at DESC"
        );
        
        return new WP_REST_Response( [
            'success' => true,
            'data'    => $agents,
        ], 200 );
    }
}
