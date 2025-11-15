<?php
/**
 * Chat REST API Controller
 *
 * 处理聊天相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.3.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 聊天 REST API 控制器
 *
 * 提供会话管理、消息收发、在线状态等功能
 */
class Tanzanite_REST_Chat_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'chat';

	/**
	 * 会话表名
	 *
	 * @var string
	 */
	private $conversations_table;

	/**
	 * 消息表名
	 *
	 * @var string
	 */
	private $messages_table;

	/**
	 * 构造函数
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->conversations_table = $wpdb->prefix . 'tanz_chat_conversations';
		$this->messages_table      = $wpdb->prefix . 'tanz_chat_messages';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.3.0
	 */
	public function register_routes() {
		// 登录
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/login',
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'login' ),
					'permission_callback' => '__return_true',
				),
			)
		);

		// 登出
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/logout',
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'logout' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 获取当前用户信息
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/me',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_current_user' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 会话列表
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/conversations',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_conversations' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 消息列表
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/messages/(?P<conversation_id>[\w-]+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_messages' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 发送消息
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/send',
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'send_message' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 标记已读
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/mark-read/(?P<conversation_id>[\w-]+)',
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'mark_as_read' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 在线状态
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/status',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_status' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 上传文件
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/upload',
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'upload_file' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		// 未读消息数
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/unread-count',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_unread_count' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);
	}

	/**
	 * 登录
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function login( $request ) {
		$username = sanitize_text_field( $request->get_param( 'username' ) );
		$password = $request->get_param( 'password' );

		if ( empty( $username ) || empty( $password ) ) {
			return $this->respond_error( 'missing_credentials', '请输入用户名和密码', 400 );
		}

		// 验证用户
		$user = wp_authenticate( $username, $password );

		if ( is_wp_error( $user ) ) {
			return $this->respond_error( 'invalid_credentials', '用户名或密码错误', 401 );
		}

		// 检查用户是否有权限（管理员或客服）
		if ( ! user_can( $user, 'manage_options' ) && ! user_can( $user, 'edit_posts' ) ) {
			return $this->respond_error( 'insufficient_permissions', '您没有权限访问客服系统', 403 );
		}

		// 设置登录状态
		wp_set_current_user( $user->ID );
		wp_set_auth_cookie( $user->ID, true );

		// 更新最后活动时间
		update_user_meta( $user->ID, 'last_activity', time() );

		return $this->respond_success(
			array(
				'user' => array(
					'id'           => $user->ID,
					'username'     => $user->user_login,
					'display_name' => $user->display_name,
					'email'        => $user->user_email,
					'avatar'       => get_avatar_url( $user->ID ),
					'roles'        => $user->roles,
				),
			)
		);
	}

	/**
	 * 登出
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function logout( $request ) {
		wp_logout();
		return $this->respond_success( array( 'success' => true ) );
	}

	/**
	 * 获取当前用户信息
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_current_user( $request ) {
		$user = wp_get_current_user();

		if ( ! $user || ! $user->ID ) {
			return $this->respond_error( 'not_logged_in', '未登录', 401 );
		}

		return $this->respond_success(
			array(
				'user' => array(
					'id'           => $user->ID,
					'username'     => $user->user_login,
					'display_name' => $user->display_name,
					'email'        => $user->user_email,
					'avatar'       => get_avatar_url( $user->ID ),
					'roles'        => $user->roles,
				),
			)
		);
	}

	/**
	 * 获取会话列表
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_conversations( $request ) {
		global $wpdb;

		$user_id  = get_current_user_id();
		$page     = max( 1, (int) $request->get_param( 'page' ) ?: 1 );
		$per_page = max( 1, min( 100, (int) $request->get_param( 'per_page' ) ?: 20 ) );
		$status   = sanitize_text_field( $request->get_param( 'status' ) ?: '' );
		$offset   = ( $page - 1 ) * $per_page;

		// 构建查询
		$where = array( '1=1' );

		// 只查询分配给当前客服的会话
		$where[] = $wpdb->prepare( 'agent_id = %d', $user_id );

		// 状态筛选
		if ( in_array( $status, array( 'active', 'closed', 'pending' ), true ) ) {
			$where[] = $wpdb->prepare( 'status = %s', $status );
		}

		$where_sql = implode( ' AND ', $where );

		// 查询总数
		$total = (int) $wpdb->get_var(
			"SELECT COUNT(*) FROM {$this->conversations_table} WHERE {$where_sql}"
		);

		// 查询会话列表
		$conversations = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->conversations_table} 
				WHERE {$where_sql} 
				ORDER BY updated_at DESC 
				LIMIT %d OFFSET %d",
				$per_page,
				$offset
			)
		);

		$items = array();
		foreach ( $conversations as $conv ) {
			$customer = get_userdata( $conv->customer_id );
			
			// 获取未读消息数
			$unread_count = (int) $wpdb->get_var(
				$wpdb->prepare(
					"SELECT COUNT(*) FROM {$this->messages_table} 
					WHERE conversation_id = %s AND is_read = 0 AND sender_type = 'customer'",
					$conv->id
				)
			);

			// 获取最后一条消息
			$last_message = $wpdb->get_row(
				$wpdb->prepare(
					"SELECT message, created_at FROM {$this->messages_table} 
					WHERE conversation_id = %s 
					ORDER BY created_at DESC 
					LIMIT 1",
					$conv->id
				)
			);

			$items[] = array(
				'id'                => $conv->id,
				'customer_id'       => (int) $conv->customer_id,
				'customer_name'     => $customer ? $customer->display_name : 'Unknown',
				'customer_avatar'   => get_avatar_url( $conv->customer_id ),
				'customer_phone'    => get_user_meta( $conv->customer_id, 'phone', true ),
				'agent_id'          => (int) $conv->agent_id,
				'status'            => $conv->status,
				'last_message'      => $last_message ? $last_message->message : '',
				'last_message_time' => $last_message ? $last_message->created_at : $conv->created_at,
				'unread_count'      => $unread_count,
				'online'            => $this->is_user_online( $conv->customer_id ),
				'created_at'        => $conv->created_at,
				'updated_at'        => $conv->updated_at,
			);
		}

		return $this->respond_success(
			array(
				'items' => $items,
				'meta'  => array(
					'page'        => $page,
					'per_page'    => $per_page,
					'total'       => $total,
					'total_pages' => ceil( $total / $per_page ),
				),
			)
		);
	}

	/**
	 * 获取消息列表
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_messages( $request ) {
		global $wpdb;

		$conversation_id = sanitize_text_field( $request['conversation_id'] );
		$page            = max( 1, (int) $request->get_param( 'page' ) ?: 1 );
		$per_page        = max( 1, min( 200, (int) $request->get_param( 'per_page' ) ?: 50 ) );
		$offset          = ( $page - 1 ) * $per_page;

		// 验证会话权限
		$conversation = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->conversations_table} WHERE id = %s",
				$conversation_id
			)
		);

		if ( ! $conversation ) {
			return $this->respond_error( 'conversation_not_found', '会话不存在', 404 );
		}

		$user_id = get_current_user_id();
		if ( (int) $conversation->agent_id !== $user_id && (int) $conversation->customer_id !== $user_id ) {
			return $this->respond_error( 'no_permission', '无权访问此会话', 403 );
		}

		// 查询总数
		$total = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->messages_table} WHERE conversation_id = %s",
				$conversation_id
			)
		);

		// 查询消息列表
		$messages = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->messages_table} 
				WHERE conversation_id = %s 
				ORDER BY created_at ASC 
				LIMIT %d OFFSET %d",
				$conversation_id,
				$per_page,
				$offset
			)
		);

		$items = array();
		foreach ( $messages as $msg ) {
			$sender = get_userdata( $msg->sender_id );
			
			$items[] = array(
				'id'             => (int) $msg->id,
				'conversation_id' => $msg->conversation_id,
				'sender_id'      => (int) $msg->sender_id,
				'sender_name'    => $sender ? $sender->display_name : 'Unknown',
				'sender_type'    => $msg->sender_type,
				'message'        => $msg->message,
				'type'           => $msg->type,
				'attachment_url' => $msg->attachment_url,
				'is_read'        => (bool) $msg->is_read,
				'created_at'     => $msg->created_at,
			);
		}

		return $this->respond_success(
			array(
				'items' => $items,
				'meta'  => array(
					'page'      => $page,
					'per_page'  => $per_page,
					'total'     => $total,
					'has_more'  => $total > ( $page * $per_page ),
				),
			)
		);
	}

	/**
	 * 发送消息
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function send_message( $request ) {
		global $wpdb;

		$conversation_id = sanitize_text_field( $request->get_param( 'conversation_id' ) );
		$message         = sanitize_textarea_field( $request->get_param( 'message' ) );
		$type            = sanitize_text_field( $request->get_param( 'type' ) ?: 'text' );
		$attachment_url  = esc_url_raw( $request->get_param( 'attachment_url' ) ?: '' );

		if ( empty( $conversation_id ) || empty( $message ) ) {
			return $this->respond_error( 'missing_params', '缺少必要参数', 400 );
		}

		// 验证会话
		$conversation = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->conversations_table} WHERE id = %s",
				$conversation_id
			)
		);

		if ( ! $conversation ) {
			return $this->respond_error( 'conversation_not_found', '会话不存在', 404 );
		}

		$user_id = get_current_user_id();
		$sender_type = ( (int) $conversation->agent_id === $user_id ) ? 'agent' : 'customer';

		// 插入消息
		$result = $wpdb->insert(
			$this->messages_table,
			array(
				'conversation_id' => $conversation_id,
				'sender_id'       => $user_id,
				'sender_type'     => $sender_type,
				'message'         => $message,
				'type'            => $type,
				'attachment_url'  => $attachment_url,
				'is_read'         => 0,
				'created_at'      => current_time( 'mysql' ),
			),
			array( '%s', '%d', '%s', '%s', '%s', '%s', '%d', '%s' )
		);

		if ( ! $result ) {
			return $this->respond_error( 'send_failed', '发送失败', 500 );
		}

		$message_id = $wpdb->insert_id;

		// 更新会话时间
		$wpdb->update(
			$this->conversations_table,
			array( 'updated_at' => current_time( 'mysql' ) ),
			array( 'id' => $conversation_id ),
			array( '%s' ),
			array( '%s' )
		);

		// 获取发送者信息
		$sender = get_userdata( $user_id );

		return $this->respond_success(
			array(
				'message' => array(
					'id'             => $message_id,
					'conversation_id' => $conversation_id,
					'sender_id'      => $user_id,
					'sender_name'    => $sender ? $sender->display_name : 'Unknown',
					'sender_type'    => $sender_type,
					'message'        => $message,
					'type'           => $type,
					'attachment_url' => $attachment_url,
					'created_at'     => current_time( 'mysql' ),
				),
			)
		);
	}

	/**
	 * 标记消息为已读
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function mark_as_read( $request ) {
		global $wpdb;

		$conversation_id = sanitize_text_field( $request['conversation_id'] );
		$user_id         = get_current_user_id();

		// 验证会话
		$conversation = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->conversations_table} WHERE id = %s",
				$conversation_id
			)
		);

		if ( ! $conversation ) {
			return $this->respond_error( 'conversation_not_found', '会话不存在', 404 );
		}

		// 标记所有未读消息为已读（对方发送的消息）
		$sender_type = ( (int) $conversation->agent_id === $user_id ) ? 'customer' : 'agent';
		
		$wpdb->update(
			$this->messages_table,
			array( 'is_read' => 1 ),
			array(
				'conversation_id' => $conversation_id,
				'sender_type'     => $sender_type,
				'is_read'         => 0,
			),
			array( '%d' ),
			array( '%s', '%s', '%d' )
		);

		// 获取剩余未读数
		$unread_count = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->messages_table} 
				WHERE conversation_id = %s AND is_read = 0 AND sender_type = %s",
				$conversation_id,
				$sender_type
			)
		);

		return $this->respond_success(
			array(
				'success'      => true,
				'unread_count' => $unread_count,
			)
		);
	}

	/**
	 * 获取在线状态
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_status( $request ) {
		$conversation_ids = $request->get_param( 'conversation_ids' );
		
		if ( empty( $conversation_ids ) ) {
			return $this->respond_error( 'missing_params', '缺少参数', 400 );
		}

		$ids = array_map( 'sanitize_text_field', explode( ',', $conversation_ids ) );
		
		global $wpdb;
		$statuses = array();

		foreach ( $ids as $conv_id ) {
			$conversation = $wpdb->get_row(
				$wpdb->prepare(
					"SELECT customer_id FROM {$this->conversations_table} WHERE id = %s",
					$conv_id
				)
			);

			if ( $conversation ) {
				$statuses[] = array(
					'conversation_id' => $conv_id,
					'customer_id'     => (int) $conversation->customer_id,
					'online'          => $this->is_user_online( $conversation->customer_id ),
					'last_seen'       => $this->get_last_seen( $conversation->customer_id ),
				);
			}
		}

		return $this->respond_success( array( 'statuses' => $statuses ) );
	}

	/**
	 * 上传文件
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function upload_file( $request ) {
		$files = $request->get_file_params();
		
		if ( empty( $files['file'] ) ) {
			return $this->respond_error( 'no_file', '未上传文件', 400 );
		}

		require_once ABSPATH . 'wp-admin/includes/file.php';
		require_once ABSPATH . 'wp-admin/includes/media.php';
		require_once ABSPATH . 'wp-admin/includes/image.php';

		$file = $files['file'];
		
		// 上传文件
		$upload = wp_handle_upload(
			$file,
			array( 'test_form' => false )
		);

		if ( isset( $upload['error'] ) ) {
			return $this->respond_error( 'upload_failed', $upload['error'], 500 );
		}

		return $this->respond_success(
			array(
				'success' => true,
				'url'     => $upload['url'],
				'type'    => wp_check_filetype( $upload['file'] )['type'],
				'size'    => filesize( $upload['file'] ),
			)
		);
	}

	/**
	 * 获取未读消息总数
	 *
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_unread_count( $request ) {
		global $wpdb;
		$user_id = get_current_user_id();

		// 获取用户所有会话的未读消息数
		$count = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->messages_table} m
				INNER JOIN {$this->conversations_table} c ON m.conversation_id = c.id
				WHERE c.agent_id = %d AND m.is_read = 0 AND m.sender_type = 'customer'",
				$user_id
			)
		);

		return $this->respond_success( array( 'count' => $count ) );
	}

	/**
	 * 检查用户是否在线
	 *
	 * @param int $user_id 用户 ID
	 * @return bool
	 */
	private function is_user_online( $user_id ) {
		$last_activity = get_user_meta( $user_id, 'last_activity', true );
		if ( ! $last_activity ) {
			return false;
		}
		
		// 5 分钟内有活动视为在线
		return ( time() - (int) $last_activity ) < 300;
	}

	/**
	 * 获取用户最后在线时间
	 *
	 * @param int $user_id 用户 ID
	 * @return int
	 */
	private function get_last_seen( $user_id ) {
		$last_activity = get_user_meta( $user_id, 'last_activity', true );
		return $last_activity ? (int) $last_activity : 0;
	}
}
