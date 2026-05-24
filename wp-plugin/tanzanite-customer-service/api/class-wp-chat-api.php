<?php
/**
 * WordPress-authenticated customer service chat API.
 *
 * This owns the website agent chat surface and keeps legacy /chat/* aliases
 * working while the frontend migrates to /customer-service/agent/*.
 *
 * @package Tanzanite_Customer_Service
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class TZ_CS_WP_Chat_API {

	public static function register_routes(): void {
		register_rest_route(
			'tanzanite/v1',
			'/customer-service/agent/conversations',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( __CLASS__, 'get_conversations' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/customer-service/agent/conversations/(?P<conversation_id>[a-zA-Z0-9_-]+)/messages',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( __CLASS__, 'get_messages' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/customer-service/agent/messages',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( __CLASS__, 'send_message' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/customer-service/agent/messages/read',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( __CLASS__, 'mark_as_read' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/customer-service/agent/conversations/(?P<conversation_id>[a-zA-Z0-9_-]+)/transfer',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( __CLASS__, 'transfer_conversation' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/customer-service/agent/status',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( __CLASS__, 'get_agent_status' ),
					'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( __CLASS__, 'update_agent_status' ),
					'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
				),
			)
		);

		self::register_legacy_chat_aliases();
	}

	private static function register_legacy_chat_aliases(): void {
		register_rest_route(
			'tanzanite/v1',
			'/chat/conversations',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( __CLASS__, 'get_conversations' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/chat/messages/(?P<conversation_id>[a-zA-Z0-9_-]+)',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( __CLASS__, 'get_messages' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/chat/send',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( __CLASS__, 'send_message' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/chat/mark-read/(?P<conversation_id>[a-zA-Z0-9_-]+)',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( __CLASS__, 'mark_as_read' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/chat/unread-count',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( __CLASS__, 'get_unread_count' ),
				'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/chat/agent-status',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( __CLASS__, 'get_agent_status' ),
					'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( __CLASS__, 'update_agent_status' ),
					'permission_callback' => array( __CLASS__, 'check_wp_agent_permission' ),
				),
			)
		);
	}

	public static function check_wp_agent_permission( WP_REST_Request $request ) {
		if ( ! is_user_logged_in() ) {
			return new WP_Error( 'not_logged_in', __( 'Not logged in.', 'tanzanite-cs' ), array( 'status' => 401 ) );
		}

		$agent = self::get_current_wp_agent();
		if ( ! $agent ) {
			return new WP_Error( 'not_agent', __( 'Current user is not an active customer service agent.', 'tanzanite-cs' ), array( 'status' => 403 ) );
		}

		$request->set_param( '_wp_agent', $agent );
		return true;
	}

	public static function get_conversations( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$agent               = $request->get_param( '_wp_agent' );
		$table_conversation  = self::conversations_table();
		$table_messages      = self::messages_table();

		if ( ! self::table_exists( $table_conversation ) ) {
			return self::ok( array( 'items' => array(), 'meta' => self::meta( 0, 1, 20 ) ), 200, array( 'conversations' => array() ) );
		}

		$page     = max( 1, (int) ( $request->get_param( 'page' ) ?: 1 ) );
		$per_page = max( 1, min( 100, (int) ( $request->get_param( 'per_page' ) ?: $request->get_param( 'limit' ) ?: 20 ) ) );
		$offset   = ( $page - 1 ) * $per_page;
		$status   = sanitize_text_field( (string) ( $request->get_param( 'status' ) ?: 'active' ) );

		$where  = 'WHERE agent_id = %s';
		$params = array( $agent['agent_id'] );

		if ( 'all' !== $status ) {
			$where   .= ' AND status = %s';
			$params[] = $status;
		}

		$total = (int) $wpdb->get_var( $wpdb->prepare( "SELECT COUNT(*) FROM {$table_conversation} {$where}", $params ) );

		$query_params = array_merge( $params, array( $per_page, $offset ) );
		$rows         = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$table_conversation} {$where} ORDER BY updated_at DESC LIMIT %d OFFSET %d",
				$query_params
			),
			ARRAY_A
		);

		$items = array();
		foreach ( $rows as $row ) {
			$items[] = self::format_conversation( $row, $table_messages );
		}

		return self::ok(
			array(
				'items' => $items,
				'meta'  => self::meta( $total, $page, $per_page ),
			),
			200,
			array(
				'conversations' => $items,
			)
		);
	}

	public static function get_messages( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$conversation_id = sanitize_text_field( (string) $request->get_param( 'conversation_id' ) );
		$agent           = $request->get_param( '_wp_agent' );
		$conversation    = self::get_agent_conversation( $conversation_id, $agent['agent_id'] );

		if ( ! $conversation ) {
			return self::error( 'conversation_not_found', __( 'Conversation not found.', 'tanzanite-cs' ), 404 );
		}

		$page     = max( 1, (int) ( $request->get_param( 'page' ) ?: 1 ) );
		$per_page = max( 1, min( 200, (int) ( $request->get_param( 'per_page' ) ?: $request->get_param( 'limit' ) ?: 50 ) ) );
		$offset   = ( $page - 1 ) * $per_page;

		$table_messages = self::messages_table();
		$total          = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$table_messages} WHERE conversation_id = %s",
				$conversation_id
			)
		);

		$rows = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$table_messages} WHERE conversation_id = %s ORDER BY created_at ASC LIMIT %d OFFSET %d",
				$conversation_id,
				$per_page,
				$offset
			),
			ARRAY_A
		);

		$items = array_map( array( __CLASS__, 'format_message' ), $rows ?: array() );

		return self::ok(
			array(
				'items' => $items,
				'meta'  => self::meta( $total, $page, $per_page ),
			),
			200,
			array(
				'messages' => $items,
			)
		);
	}

	public static function send_message( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$agent           = $request->get_param( '_wp_agent' );
		$conversation_id = sanitize_text_field( (string) $request->get_param( 'conversation_id' ) );
		$message         = sanitize_textarea_field( (string) $request->get_param( 'message' ) );
		$message_type    = sanitize_text_field( (string) ( $request->get_param( 'message_type' ) ?: 'text' ) );

		if ( '' === $conversation_id || '' === $message ) {
			return self::error( 'missing_params', __( 'Conversation id and message are required.', 'tanzanite-cs' ), 400 );
		}

		$conversation = self::get_agent_conversation( $conversation_id, $agent['agent_id'], true );
		if ( ! $conversation ) {
			return self::error( 'conversation_not_found', __( 'Conversation not found.', 'tanzanite-cs' ), 404 );
		}

		$table_messages = self::messages_table();
		$now            = current_time( 'mysql' );

		$result = $wpdb->insert(
			$table_messages,
			array(
				'conversation_id' => $conversation_id,
				'sender_type'     => 'agent',
				'sender_id'       => get_current_user_id(),
				'sender_name'     => $agent['name'],
				'sender_email'    => $agent['email'],
				'agent_id'        => $agent['agent_id'],
				'message_type'    => $message_type,
				'message'         => $message,
				'is_read'         => 0,
				'created_at'      => $now,
			),
			array( '%s', '%s', '%d', '%s', '%s', '%s', '%s', '%s', '%d', '%s' )
		);

		if ( false === $result ) {
			return self::error( 'send_failed', __( 'Message could not be saved.', 'tanzanite-cs' ), 500 );
		}

		self::touch_conversation( $conversation_id, $agent['agent_id'], $message, $now );

		$row = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$table_messages} WHERE id = %d",
				$wpdb->insert_id
			),
			ARRAY_A
		);

		$formatted = self::format_message( $row );

		return self::ok(
			array(
				'message' => $formatted,
			),
			200,
			array(
				'message' => $formatted,
			)
		);
	}

	public static function mark_as_read( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$agent           = $request->get_param( '_wp_agent' );
		$conversation_id = sanitize_text_field( (string) ( $request->get_param( 'conversation_id' ) ?: $request['conversation_id'] ) );
		$conversation    = self::get_agent_conversation( $conversation_id, $agent['agent_id'] );

		if ( ! $conversation ) {
			return self::error( 'conversation_not_found', __( 'Conversation not found.', 'tanzanite-cs' ), 404 );
		}

		$table_messages = self::messages_table();
		$wpdb->query(
			$wpdb->prepare(
				"UPDATE {$table_messages}
				 SET is_read = 1
				 WHERE conversation_id = %s
				 AND sender_type IN ('visitor', 'user')
				 AND is_read = 0",
				$conversation_id
			)
		);

		return self::ok(
			array(
				'success'      => true,
				'unread_count' => 0,
			)
		);
	}

	public static function transfer_conversation( WP_REST_Request $request ): WP_REST_Response {
		$agent           = $request->get_param( '_wp_agent' );
		$conversation_id = sanitize_text_field( (string) $request->get_param( 'conversation_id' ) );

		if ( ! self::get_agent_conversation( $conversation_id, $agent['agent_id'] ) ) {
			return self::error( 'conversation_not_found', __( 'Conversation not found.', 'tanzanite-cs' ), 404 );
		}

		$request->set_param( '_agent', $agent );
		$response = TZ_CS_Agent_API::transfer_conversation( $request );
		$data     = $response->get_data();

		if ( ! empty( $data['success'] ) ) {
			$data['ok'] = true;
		}

		return new WP_REST_Response( $data, $response->get_status() );
	}

	public static function get_agent_status( WP_REST_Request $request ): WP_REST_Response {
		$agent  = $request->get_param( '_wp_agent' );
		$status = $agent['online_status'] ?: 'offline';

		return self::ok(
			array(
				'agent_id'       => $agent['agent_id'],
				'name'           => $agent['name'],
				'status'         => $status,
				'last_active_at' => $agent['last_active_at'],
			)
		);
	}

	public static function update_agent_status( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$agent  = $request->get_param( '_wp_agent' );
		$status = sanitize_text_field( (string) $request->get_param( 'status' ) );

		if ( ! in_array( $status, array( 'online', 'busy', 'away', 'offline' ), true ) ) {
			return self::error( 'invalid_status', __( 'Invalid status.', 'tanzanite-cs' ), 400 );
		}

		$now = current_time( 'mysql' );
		$wpdb->update(
			self::agents_table(),
			array(
				'online_status'  => $status,
				'last_active_at' => $now,
			),
			array( 'agent_id' => $agent['agent_id'] ),
			array( '%s', '%s' ),
			array( '%s' )
		);

		return self::ok(
			array(
				'status'         => $status,
				'last_active_at' => $now,
			)
		);
	}

	public static function get_unread_count( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$agent              = $request->get_param( '_wp_agent' );
		$table_conversation = self::conversations_table();
		$table_messages     = self::messages_table();
		$count              = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*)
				 FROM {$table_messages} m
				 INNER JOIN {$table_conversation} c ON m.conversation_id = c.id
				 WHERE c.agent_id = %s
				 AND m.is_read = 0
				 AND m.sender_type IN ('visitor', 'user')",
				$agent['agent_id']
			)
		);

		return self::ok(
			array(
				'count'        => $count,
				'totalUnread'  => $count,
				'total_unread' => $count,
			)
		);
	}

	private static function get_current_wp_agent() {
		global $wpdb;

		$table = self::agents_table();
		if ( ! self::table_exists( $table ) ) {
			return null;
		}

		$user_id = get_current_user_id();
		if ( ! $user_id ) {
			return null;
		}

		return $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$table} WHERE wp_user_id = %d AND status = 'active' LIMIT 1",
				$user_id
			),
			ARRAY_A
		);
	}

	private static function get_agent_conversation( $conversation_id, $agent_id, $claim_unassigned = false ) {
		global $wpdb;

		if ( '' === $conversation_id ) {
			return null;
		}

		$table = self::conversations_table();
		if ( ! self::table_exists( $table ) ) {
			return null;
		}

		$conversation = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$table} WHERE id = %s AND (agent_id = %s OR agent_id = '') LIMIT 1",
				$conversation_id,
				$agent_id
			),
			ARRAY_A
		);

		if ( $conversation && '' === (string) $conversation['agent_id'] && $claim_unassigned ) {
			$wpdb->update(
				$table,
				array(
					'agent_id'   => $agent_id,
					'updated_at' => current_time( 'mysql' ),
				),
				array( 'id' => $conversation_id ),
				array( '%s', '%s' ),
				array( '%s' )
			);
			$conversation['agent_id'] = $agent_id;
		}

		return $conversation;
	}

	private static function touch_conversation( $conversation_id, $agent_id, $message, $time ) {
		global $wpdb;

		$wpdb->update(
			self::conversations_table(),
			array(
				'agent_id'          => $agent_id,
				'last_message'      => $message,
				'last_message_time' => $time,
				'updated_at'        => $time,
			),
			array( 'id' => $conversation_id ),
			array( '%s', '%s', '%s', '%s' ),
			array( '%s' )
		);
	}

	private static function format_conversation( array $row, $table_messages ) {
		global $wpdb;

		$user_id        = isset( $row['user_id'] ) ? (int) $row['user_id'] : 0;
		$wp_user        = $user_id ? get_userdata( $user_id ) : null;
		$customer_name  = $row['visitor_name'] ?: ( $wp_user ? $wp_user->display_name : __( 'Visitor', 'tanzanite-cs' ) );
		$customer_email = $row['visitor_email'] ?: ( $wp_user ? $wp_user->user_email : '' );
		$unread_count   = self::table_exists( $table_messages ) ? (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$table_messages}
				 WHERE conversation_id = %s
				 AND is_read = 0
				 AND sender_type IN ('visitor', 'user')",
				$row['id']
			)
		) : 0;

		return array(
			'id'                => $row['id'],
			'customer_id'       => $user_id,
			'customer_name'     => $customer_name,
			'customer_email'    => $customer_email,
			'customer_avatar'   => $user_id ? get_avatar_url( $user_id ) : '',
			'agent_id'          => $row['agent_id'],
			'status'            => $row['status'] ?: 'active',
			'last_message'      => $row['last_message'] ?: '',
			'last_message_time' => $row['last_message_time'] ?: $row['updated_at'],
			'unread_count'      => $unread_count,
			'created_at'        => $row['created_at'],
			'updated_at'        => $row['updated_at'],
		);
	}

	private static function format_message( array $row ) {
		$is_agent = ( $row['sender_type'] ?? '' ) === 'agent';
		$metadata = array();

		if ( ! empty( $row['metadata'] ) ) {
			$decoded  = json_decode( $row['metadata'], true );
			$metadata = is_array( $decoded ) ? $decoded : array();
		}

		return array(
			'id'              => isset( $row['id'] ) ? (int) $row['id'] : 0,
			'conversation_id' => $row['conversation_id'] ?? '',
			'sender_id'       => isset( $row['sender_id'] ) ? (int) $row['sender_id'] : 0,
			'sender_name'     => $row['sender_name'] ?? '',
			'sender_email'    => $row['sender_email'] ?? '',
			'sender_type'     => $row['sender_type'] ?? '',
			'is_agent'        => $is_agent,
			'agent_id'        => $row['agent_id'] ?? '',
			'message'         => $row['message'] ?? '',
			'message_type'    => $row['message_type'] ?? 'text',
			'type'            => $row['message_type'] ?? 'text',
			'is_read'         => ! empty( $row['is_read'] ),
			'metadata'        => $metadata,
			'created_at'      => $row['created_at'] ?? '',
		);
	}

	private static function ok( array $data, $status = 200, array $root_extra = array() ): WP_REST_Response {
		return new WP_REST_Response(
			array_merge(
				array(
					'ok'      => true,
					'success' => true,
					'data'    => $data,
				),
				$root_extra
			),
			$status
		);
	}

	private static function error( $code, $message, $status = 400 ): WP_REST_Response {
		return new WP_REST_Response(
			array(
				'ok'      => false,
				'success' => false,
				'code'    => $code,
				'message' => $message,
			),
			$status
		);
	}

	private static function meta( $total, $page, $per_page ) {
		return array(
			'total'       => (int) $total,
			'page'        => (int) $page,
			'per_page'    => (int) $per_page,
			'total_pages' => $per_page ? (int) ceil( $total / $per_page ) : 0,
		);
	}

	private static function table_exists( $table ) {
		global $wpdb;

		return $wpdb->get_var( $wpdb->prepare( 'SHOW TABLES LIKE %s', $table ) ) === $table;
	}

	private static function conversations_table() {
		global $wpdb;
		return $wpdb->prefix . 'tz_cs_conversations';
	}

	private static function messages_table() {
		global $wpdb;
		return $wpdb->prefix . 'tz_cs_messages';
	}

	private static function agents_table() {
		global $wpdb;
		return $wpdb->prefix . 'tz_cs_agents';
	}
}
