<?php
/**
 * 数据库管理类
 * 
 * @package Tanzanite_Customer_Service
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class TZ_CS_Database {
    
    /**
     * 创建数据库表
     */
    public static function create_tables(): void {
        global $wpdb;
        
        // 记录激活时间用于调试
        update_option( 'tz_cs_activation_time', current_time( 'mysql' ) );
        
        $charset_collate = $wpdb->get_charset_collate();
        
        // 聊天记录表 - 只存储消息，不跟踪已读状态
        $table_messages = $wpdb->prefix . 'tz_cs_messages';
        $sql_messages = "CREATE TABLE $table_messages (
            id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
            conversation_id varchar(100) NOT NULL,
            sender_type varchar(20) NOT NULL,
            sender_id bigint(20) unsigned DEFAULT 0,
            sender_name varchar(255) NOT NULL,
            sender_email varchar(255) DEFAULT '',
            agent_id varchar(100) DEFAULT '',
            message_type varchar(20) DEFAULT 'text',
            message text NOT NULL,
            metadata longtext DEFAULT NULL,
            created_at datetime NOT NULL,
            PRIMARY KEY  (id),
            KEY conversation_id (conversation_id),
            KEY sender_type (sender_type),
            KEY agent_id (agent_id),
            KEY created_at (created_at)
        ) $charset_collate;";
        
        // 会话列表表 - 只记录基本信息，不跟踪未读消息
        $table_conversations = $wpdb->prefix . 'tz_cs_conversations';
        $sql_conversations = "CREATE TABLE $table_conversations (
            id varchar(100) NOT NULL,
            visitor_id varchar(100) DEFAULT '',
            user_id bigint(20) unsigned DEFAULT 0,
            agent_id varchar(100) DEFAULT '',
            visitor_name varchar(255) DEFAULT '',
            visitor_email varchar(255) DEFAULT '',
            last_message text DEFAULT NULL,
            last_message_time datetime DEFAULT NULL,
            status varchar(20) DEFAULT 'active',
            has_auto_reply tinyint(1) DEFAULT 0,
            auto_reply_count int DEFAULT 0,
            needs_human tinyint(1) DEFAULT 0,
            last_welcome_sent datetime DEFAULT NULL,
            transferred_from varchar(50) DEFAULT '',
            transferred_at datetime DEFAULT NULL,
            created_at datetime NOT NULL,
            updated_at datetime NOT NULL,
            PRIMARY KEY  (id),
            KEY visitor_id (visitor_id),
            KEY user_id (user_id),
            KEY agent_id (agent_id),
            KEY status (status),
            KEY has_auto_reply (has_auto_reply),
            KEY needs_human (needs_human),
            KEY updated_at (updated_at)
        ) $charset_collate;";
        
        // 自动回复规则表
        $table_auto_replies = $wpdb->prefix . 'tz_cs_auto_replies';
        $sql_auto_replies = "CREATE TABLE $table_auto_replies (
            id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
            type varchar(20) NOT NULL,
            trigger_keyword varchar(255) DEFAULT NULL,
            reply_message text NOT NULL,
            agent_id varchar(100) DEFAULT '',
            is_active tinyint(1) DEFAULT 1,
            priority int DEFAULT 0,
            match_type varchar(20) DEFAULT 'exact',
            created_at datetime NOT NULL,
            updated_at datetime NOT NULL,
            PRIMARY KEY  (id),
            KEY type (type),
            KEY agent_id (agent_id),
            KEY is_active (is_active),
            KEY priority (priority)
        ) $charset_collate;";
        
        // 客服账号表
        $table_agents = $wpdb->prefix . 'tz_cs_agents';
        $sql_agents = "CREATE TABLE $table_agents (
            id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
            agent_id varchar(50) NOT NULL,
            name varchar(100) NOT NULL,
            email varchar(100) NOT NULL,
            password varchar(255) NOT NULL,
            avatar varchar(500) DEFAULT '',
            whatsapp varchar(50) DEFAULT '',
            status varchar(20) DEFAULT 'active',
            last_login datetime,
            created_at datetime NOT NULL,
            updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
            PRIMARY KEY  (id),
            UNIQUE KEY agent_id (agent_id),
            KEY status (status),
            KEY email (email)
        ) $charset_collate;";
        
        // 客服Token表
        $table_tokens = $wpdb->prefix . 'tz_cs_tokens';
        $sql_tokens = "CREATE TABLE $table_tokens (
            id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
            agent_id varchar(50) NOT NULL,
            token varchar(255) NOT NULL,
            device_info text,
            expires_at datetime NOT NULL,
            created_at datetime NOT NULL,
            PRIMARY KEY  (id),
            UNIQUE KEY token (token),
            KEY agent_id (agent_id),
            KEY expires_at (expires_at)
        ) $charset_collate;";
        
        // 转接历史记录表
        $table_transfers = $wpdb->prefix . 'tz_cs_transfers';
        $sql_transfers = "CREATE TABLE $table_transfers (
            id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
            conversation_id varchar(100) NOT NULL,
            from_agent_id varchar(50) NOT NULL,
            from_agent_name varchar(100) NOT NULL,
            to_agent_id varchar(50) NOT NULL,
            to_agent_name varchar(100) NOT NULL,
            reason text,
            created_at datetime NOT NULL,
            PRIMARY KEY  (id),
            KEY conversation_id (conversation_id),
            KEY from_agent_id (from_agent_id),
            KEY to_agent_id (to_agent_id),
            KEY created_at (created_at)
        ) $charset_collate;";
        
        // 通知表
        $table_notifications = $wpdb->prefix . 'tz_cs_notifications';
        $sql_notifications = "CREATE TABLE $table_notifications (
            id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
            agent_id varchar(50) NOT NULL,
            type varchar(50) NOT NULL,
            title varchar(255) NOT NULL,
            message text NOT NULL,
            data text,
            is_read tinyint(1) NOT NULL DEFAULT 0,
            created_at datetime NOT NULL,
            read_at datetime DEFAULT NULL,
            PRIMARY KEY  (id),
            KEY agent_id (agent_id),
            KEY type (type),
            KEY is_read (is_read),
            KEY created_at (created_at)
        ) $charset_collate;";
        
        require_once ABSPATH . 'wp-admin/includes/upgrade.php';
        
        // 创建或更新旧表
        $result_messages = dbDelta( $sql_messages );
        $result_conversations = dbDelta( $sql_conversations );
        $result_auto_replies = dbDelta( $sql_auto_replies );
        
        // 检查客服表是否存在
        $agents_table_exists = $wpdb->get_var( "SHOW TABLES LIKE '$table_agents'" ) === $table_agents;
        
        // 如果客服表存在，检查并添加状态字段
        if ( $agents_table_exists ) {
            // 检查 status 字段是否存在
            $status_column_exists = $wpdb->get_results( 
                "SHOW COLUMNS FROM $table_agents LIKE 'status'" 
            );
            
            if ( empty( $status_column_exists ) ) {
                $wpdb->query( 
                    "ALTER TABLE $table_agents ADD COLUMN status VARCHAR(20) DEFAULT 'offline' AFTER role" 
                );
            }
            
            // 检查 last_active_at 字段是否存在
            $last_active_column_exists = $wpdb->get_results( 
                "SHOW COLUMNS FROM $table_agents LIKE 'last_active_at'" 
            );
            
            if ( empty( $last_active_column_exists ) ) {
                $wpdb->query( 
                    "ALTER TABLE $table_agents ADD COLUMN last_active_at DATETIME DEFAULT NULL AFTER status" 
                );
            }
        }
        
        if ( ! $agents_table_exists ) {
            // 表不存在，创建新表
            $wpdb->query( $sql_agents );
            $result_agents = [ 'created' => $table_agents ];
        } else {
            // 表已存在，检查并添加缺失的字段
            $columns = $wpdb->get_col( "DESCRIBE $table_agents", 0 );
            
            if ( ! in_array( 'whatsapp', $columns ) ) {
                $wpdb->query( "ALTER TABLE $table_agents ADD COLUMN whatsapp varchar(50) DEFAULT '' AFTER avatar" );
            }
            
            $result_agents = [ 'updated' => $table_agents ];
        }
        
        // 检查 Token 表是否存在
        $tokens_table_exists = $wpdb->get_var( "SHOW TABLES LIKE '$table_tokens'" ) === $table_tokens;
        
        if ( ! $tokens_table_exists ) {
            // 表不存在，创建新表
            $wpdb->query( $sql_tokens );
            $result_tokens = [ 'created' => $table_tokens ];
        } else {
            $result_tokens = [ 'exists' => $table_tokens ];
        }
        
        // 检查转接历史表是否存在
        $transfers_table_exists = $wpdb->get_var( "SHOW TABLES LIKE '$table_transfers'" ) === $table_transfers;
        
        if ( ! $transfers_table_exists ) {
            // 表不存在，创建新表
            $wpdb->query( $sql_transfers );
            $result_transfers = [ 'created' => $table_transfers ];
        } else {
            $result_transfers = [ 'exists' => $table_transfers ];
        }
        
        // 检查通知表是否存在
        $notifications_table_exists = $wpdb->get_var( "SHOW TABLES LIKE '$table_notifications'" ) === $table_notifications;
        
        if ( ! $notifications_table_exists ) {
            // 表不存在，创建新表
            $wpdb->query( $sql_notifications );
            $result_notifications = [ 'created' => $table_notifications ];
        } else {
            $result_notifications = [ 'exists' => $table_notifications ];
        }
        
        // 记录创建结果用于调试
        update_option( 'tz_cs_db_creation_result', [
            'messages' => $result_messages,
            'conversations' => $result_conversations,
            'auto_replies' => $result_auto_replies,
            'time' => current_time( 'mysql' )
        ] );
        
        // 记录版本号
        update_option( 'tz_cs_db_version', '1.2.0' );
    }
    
    /**
     * 删除数据库表
     */
    public static function drop_tables(): void {
        global $wpdb;
        
        $table_messages = $wpdb->prefix . 'tz_cs_messages';
        $table_conversations = $wpdb->prefix . 'tz_cs_conversations';
        $table_auto_replies = $wpdb->prefix . 'tz_cs_auto_replies';
        
        $wpdb->query( "DROP TABLE IF EXISTS $table_messages" );
        $wpdb->query( "DROP TABLE IF EXISTS $table_conversations" );
        $wpdb->query( "DROP TABLE IF EXISTS $table_auto_replies" );
    }
}
