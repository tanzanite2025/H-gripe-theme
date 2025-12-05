<?php
/**
 * Feedback REST API Controller
 *
 * 用户反馈 / 留言相关的 REST API
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 留言 REST API 控制器
 */
class Tanzanite_REST_Feedback_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'feedback';

	/**
	 * 留言表名
	 *
	 * @var string
	 */
	private $feedback_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->feedback_table = $wpdb->prefix . 'tanz_feedback';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 列表 & 创建
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => '__return_true', // 公开读取
					'args'                => $this->get_collection_params(),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => 'is_user_logged_in', // 必须登录
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 更新状态（后台审核）
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\\d+)/status',
			array(
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_status' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
					'args'                => array(
						'id'     => array(
							'validate_callback' => array( $this, 'validate_numeric_param' ),
						),
						'status' => array(
							'type'     => 'string',
							'required' => true,
						),
					),
				),
			)
		);

		// 可选：资格检查（用于前端决定是否显示表单）
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/eligibility',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_eligibility' ),
					'permission_callback' => '__return_true',
				),
			)
		);
	}

	/**
	 * 获取留言列表（仅返回 approved，除非显式传其它状态）
	 */
	public function get_items( $request ) {
		global $wpdb;

		$pagination = $this->get_pagination_params( $request );
		$thread     = sanitize_text_field( (string) $request->get_param( 'thread' ) );
		$status     = $request->get_param( 'status' );
		$search     = $request->get_param( 'search' );

		if ( '' === $thread ) {
			return $this->respond_error( 'missing_thread', __( '缺少 thread 参数。', 'tanzanite-settings' ), 400 );
		}

		$where  = array( 'thread_key = %s' );
		$params = array( $thread );

		// 状态：前台默认只看 approved
		if ( $status ) {
			$where[]  = 'status = %s';
			$params[] = sanitize_key( $status );
		} else {
			$where[]  = "status = 'approved'";
		}

		if ( $search ) {
			$like      = '%' . $wpdb->esc_like( $search ) . '%';
			$where[]   = 'content LIKE %s';
			$params[]  = $like;
		}

		$where_sql = 'WHERE ' . implode( ' AND ', $where );

		$query    = "SELECT * FROM {$this->feedback_table} {$where_sql} ORDER BY created_at DESC LIMIT %d OFFSET %d";
		$params[] = (int) $pagination['per_page'];
		$params[] = (int) $pagination['offset'];

		$rows = $wpdb->get_results( $wpdb->prepare( $query, $params ), ARRAY_A );

		// 统计总数（去掉最后两个 limit/offset）
		$count_params = array_slice( $params, 0, count( $params ) - 2 );
		$count_query  = "SELECT COUNT(*) FROM {$this->feedback_table} {$where_sql}";
		$total        = (int) $wpdb->get_var( $wpdb->prepare( $count_query, $count_params ) );

		$db_error = $this->check_db_error( 'feedback_get_items' );
		if ( $db_error ) {
			return $db_error;
		}

		$items = array_map( array( $this, 'format_feedback_row' ), $rows );

		return $this->respond_success(
			array(
				'data'       => $items,
				'pagination' => $this->build_pagination_meta( $total, $pagination['page'], $pagination['per_page'] ),
			)
		);
	}

	/**
	 * 创建留言（登录 + 审核制，默认 pending）
	 */
	public function create_item( $request ) {
		global $wpdb;

		$required = $this->validate_required_params( $request, array( 'thread', 'content' ) );
		if ( is_wp_error( $required ) ) {
			return $this->respond_error( 'missing_parameter', $required->get_error_message(), 400 );
		}

		$user_id = get_current_user_id();
		if ( ! $user_id ) {
			return $this->respond_error( 'not_logged_in', __( '请先登录后再留言。', 'tanzanite-settings' ), 401 );
		}

		$raw = array(
			'thread' => $request->get_param( 'thread' ),
			'name'   => $request->get_param( 'name' ),
			'email'  => $request->get_param( 'email' ),
			'content'=> $request->get_param( 'content' ),
			'locale' => $request->get_param( 'locale' ),
		);

		$data = $this->sanitize_data(
			$raw,
			array(
				'thread'  => 'text',
				'name'    => 'text',
				'email'   => 'email',
				'content' => 'textarea',
				'locale'  => 'text',
			)
		);

		if ( empty( $data['thread'] ) || empty( $data['content'] ) ) {
			return $this->respond_error( 'invalid_payload', __( 'thread 和 content 不能为空。', 'tanzanite-settings' ), 400 );
		}

		// 如果未传 name/email，则从当前用户资料填充
		$user = get_userdata( $user_id );
		if ( $user ) {
			if ( empty( $data['name'] ) ) {
				$data['name'] = $user->display_name ?: $user->user_login;
			}
			if ( empty( $data['email'] ) ) {
				$data['email'] = $user->user_email;
			}
		}

		$inserted = $wpdb->insert(
			$this->feedback_table,
			array(
				'thread_key' => $data['thread'],
				'user_id'    => $user_id,
				'name'       => $data['name'] ?? null,
				'email'      => $data['email'] ?? null,
				'content'    => $data['content'],
				'status'     => 'pending',
				'locale'     => $data['locale'] ?? null,
				'created_at' => current_time( 'mysql' ),
			),
			array( '%s', '%d', '%s', '%s', '%s', '%s', '%s', '%s' )
		);

		if ( false === $inserted ) {
			$db_error = $this->check_db_error( 'feedback_create' );
			if ( $db_error ) {
				return $db_error;
			}
			return $this->respond_error( 'feedback_create_failed', __( '创建留言失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$feedback_id = $wpdb->insert_id;

		$this->log_audit(
			'create',
			'feedback',
			$feedback_id,
			array(
				'thread' => $data['thread'],
			),
			$request
		);

		return $this->respond_success(
			array(
				'message' => __( '留言已提交，等待审核。', 'tanzanite-settings' ),
				'id'      => $feedback_id,
				'status'  => 'pending',
			),
			201
		);
	}

	/**
	 * 后台更新留言状态（审核/隐藏）
	 */
	public function update_status( $request ) {
		global $wpdb;

		$id     = (int) $request['id'];
		$status = sanitize_key( (string) $request->get_param( 'status' ) );

		if ( ! $id ) {
			return $this->respond_error( 'invalid_id', __( '无效的留言 ID。', 'tanzanite-settings' ), 400 );
		}

		$allowed_statuses = array( 'pending', 'approved', 'rejected', 'hidden' );
		if ( ! in_array( $status, $allowed_statuses, true ) ) {
			return $this->respond_error( 'invalid_status', __( '无效的状态值。', 'tanzanite-settings' ), 400 );
		}

		$updated = $wpdb->update(
			$this->feedback_table,
			array( 'status' => $status ),
			array( 'id' => $id ),
			array( '%s' ),
			array( '%d' )
		);

		if ( false === $updated ) {
			$db_error = $this->check_db_error( 'feedback_update_status' );
			if ( $db_error ) {
				return $db_error;
			}
			return $this->respond_error( 'feedback_update_failed', __( '更新留言状态失败。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit(
			'update_status',
			'feedback',
			$id,
			array( 'status' => $status ),
			$request
		);

		return $this->respond_success(
			array(
				'id'     => $id,
				'status' => $status,
			)
		);
	}

	/**
	 * 检查当前用户是否可以留言（前端用来显示/隐藏表单）
	 */
	public function get_eligibility( $request ) {
		$logged_in = is_user_logged_in();

		return $this->respond_success(
			array(
				'can_post' => $logged_in,
				'logged_in' => $logged_in,
				'reason'   => $logged_in ? null : __( '请登录后再留言。', 'tanzanite-settings' ),
			)
		);
	}

	/**
	 * 集合参数定义
	 */
	private function get_collection_params() {
		return array(
			'thread'   => array(
				'type'     => 'string',
				'required' => true,
			),
			'page'     => array(
				'type'    => 'integer',
				'default' => 1,
			),
			'per_page' => array(
				'type'    => 'integer',
				'default' => 20,
			),
			'status'   => array(
				'type' => 'string',
			),
			'search'   => array(
				'type' => 'string',
			),
		);
	}

	/**
	 * 创建参数定义
	 */
	private function get_create_params() {
		return array(
			'thread' => array(
				'type'     => 'string',
				'required' => true,
			),
			'content' => array(
				'type'     => 'string',
				'required' => true,
			),
			'name'    => array(
				'type' => 'string',
			),
			'email'   => array(
				'type' => 'string',
			),
			'locale'  => array(
				'type' => 'string',
			),
		);
	}

	/**
	 * 格式化留言行
	 */
	private function format_feedback_row( $row ) {
		return array(
			'id'         => (int) $row['id'],
			'thread_key' => $row['thread_key'],
			'user_id'    => (int) $row['user_id'],
			'name'       => $row['name'],
			'content'    => $row['content'],
			'status'     => $row['status'],
			'locale'     => $row['locale'],
			'created_at' => $row['created_at'],
		);
	}
}
