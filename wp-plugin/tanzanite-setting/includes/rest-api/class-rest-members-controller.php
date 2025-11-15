<?php
/**
 * Members REST API Controller
 *
 * 处理会员档案相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 会员档案 REST API 控制器
 *
 * 提供会员档案的 CRUD 操作
 */
class Tanzanite_REST_Members_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'members';

	/**
	 * 会员档案表名
	 *
	 * @var string
	 */
	private $member_profiles_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->member_profiles_table = $wpdb->prefix . 'tanz_member_profiles';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 列表
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => $this->get_collection_params(),
				),
			)
		);

		// 获取、更新、删除单个会员
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => array(
						'id' => array(
							'validate_callback' => 'is_numeric',
						),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => array(
						'id' => array(
							'validate_callback' => 'is_numeric',
						),
					),
				),
			)
		);
	}

	/**
	 * 获取会员列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		$per_page   = (int) max( 1, min( 200, $request->get_param( 'per_page' ) ?: 20 ) );
		$page       = (int) max( 1, $request->get_param( 'page' ) ?: 1 );
		$search     = trim( (string) ( $request->get_param( 'search' ) ?: '' ) );
		$min_points = (int) max( 0, $request->get_param( 'min_points' ) ?: 0 );
		$offset     = ( $page - 1 ) * $per_page;

		// 构建查询
		$where_clauses = array( '1=1' );
		$where_values  = array();

		if ( ! empty( $search ) ) {
			$like_search     = '%' . $wpdb->esc_like( $search ) . '%';
			$where_clauses[] = '(u.user_login LIKE %s OR u.user_email LIKE %s OR u.display_name LIKE %s OR p.full_name LIKE %s OR p.phone LIKE %s)';
			$where_values    = array_merge( $where_values, array( $like_search, $like_search, $like_search, $like_search, $like_search ) );
		}

		if ( $min_points > 0 ) {
			$where_clauses[] = 'COALESCE(p.points, 0) >= %d';
			$where_values[]  = $min_points;
		}

		$where_sql = implode( ' AND ', $where_clauses );

		// 获取总数
		$count_sql = "SELECT COUNT(DISTINCT u.ID) FROM {$wpdb->users} u 
					 LEFT JOIN {$this->member_profiles_table} p ON u.ID = p.user_id 
					 WHERE {$where_sql}";

		if ( ! empty( $where_values ) ) {
			$count_sql = $wpdb->prepare( $count_sql, $where_values );
		}

		$total = (int) $wpdb->get_var( $count_sql );

		// 获取数据
		$query_sql = "SELECT u.ID as user_id, u.user_login as username, u.user_email as email, 
					 u.user_registered as registered_at,
					 p.full_name, p.phone, p.country, p.address, p.brand, 
					 COALESCE(p.points, 0) as points, p.marketing_optin, p.notes
					 FROM {$wpdb->users} u 
					 LEFT JOIN {$this->member_profiles_table} p ON u.ID = p.user_id 
					 WHERE {$where_sql}
					 ORDER BY u.ID DESC
					 LIMIT %d OFFSET %d";

		$query_values = array_merge( $where_values, array( $per_page, $offset ) );
		$query_sql    = $wpdb->prepare( $query_sql, $query_values );

		$results = $wpdb->get_results( $query_sql, ARRAY_A );

		return $this->respond_success(
			array(
				'items'    => $results ?: array(),
				'total'    => $total,
				'page'     => $page,
				'per_page' => $per_page,
			)
		);
	}

	/**
	 * 获取单个会员信息
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		global $wpdb;

		$user_id = (int) $request->get_param( 'id' );
		$user    = get_userdata( $user_id );

		if ( ! $user ) {
			return $this->respond_error( 'member_not_found', __( '会员不存在。', 'tanzanite-settings' ), 404 );
		}

		$profile_sql = $wpdb->prepare(
			"SELECT * FROM {$this->member_profiles_table} WHERE user_id = %d",
			$user_id
		);
		$profile     = $wpdb->get_row( $profile_sql, ARRAY_A );

		$data = array(
			'user_id'         => $user->ID,
			'username'        => $user->user_login,
			'email'           => $user->user_email,
			'registered_at'   => $user->user_registered,
			'full_name'       => $profile['full_name'] ?? '',
			'phone'           => $profile['phone'] ?? '',
			'country'         => $profile['country'] ?? '',
			'address'         => $profile['address'] ?? '',
			'brand'           => $profile['brand'] ?? '',
			'points'          => isset( $profile['points'] ) ? (int) $profile['points'] : 0,
			'marketing_optin' => isset( $profile['marketing_optin'] ) ? (int) $profile['marketing_optin'] : 0,
			'notes'           => $profile['notes'] ?? '',
		);

		return $this->respond_success( $data );
	}

	/**
	 * 更新会员信息
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$user_id = (int) $request->get_param( 'id' );
		$user    = get_userdata( $user_id );

		if ( ! $user ) {
			return $this->respond_error( 'member_not_found', __( '会员不存在。', 'tanzanite-settings' ), 404 );
		}

		$data = array(
			'full_name'       => sanitize_text_field( $request->get_param( 'full_name' ) ?? '' ),
			'phone'           => sanitize_text_field( $request->get_param( 'phone' ) ?? '' ),
			'country'         => sanitize_text_field( $request->get_param( 'country' ) ?? '' ),
			'address'         => sanitize_text_field( $request->get_param( 'address' ) ?? '' ),
			'brand'           => sanitize_text_field( $request->get_param( 'brand' ) ?? '' ),
			'points'          => max( 0, (int) ( $request->get_param( 'points' ) ?? 0 ) ),
			'marketing_optin' => (int) ( $request->get_param( 'marketing_optin' ) ?? 0 ),
			'notes'           => sanitize_textarea_field( $request->get_param( 'notes' ) ?? '' ),
			'updated_at'      => current_time( 'mysql' ),
		);

		// 检查档案是否存在
		$exists = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->member_profiles_table} WHERE user_id = %d",
				$user_id
			)
		);

		if ( $exists ) {
			// 更新
			$updated = $wpdb->update(
				$this->member_profiles_table,
				$data,
				array( 'user_id' => $user_id ),
				array( '%s', '%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s' ),
				array( '%d' )
			);

			if ( false === $updated ) {
				return $this->respond_error( 'failed_update_member', __( '更新会员信息失败。', 'tanzanite-settings' ), 500 );
			}
		} else {
			// 插入
			$data['user_id']    = $user_id;
			$data['created_at'] = current_time( 'mysql' );
			$inserted           = $wpdb->insert(
				$this->member_profiles_table,
				$data,
				array( '%d', '%s', '%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s', '%s' )
			);

			if ( false === $inserted ) {
				return $this->respond_error( 'failed_create_member_profile', __( '创建会员档案失败。', 'tanzanite-settings' ), 500 );
			}
		}

		$this->log_audit( 'update', 'member', $user_id, array( 'username' => $user->user_login ), $request );

		return $this->respond_success( array( 'message' => __( '会员信息已更新。', 'tanzanite-settings' ) ) );
	}

	/**
	 * 删除会员档案
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$user_id = (int) $request->get_param( 'id' );

		$deleted = $wpdb->delete(
			$this->member_profiles_table,
			array( 'user_id' => $user_id ),
			array( '%d' )
		);

		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_member', __( '删除失败。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'member', $user_id, array(), $request );

		return $this->respond_success( array( 'message' => __( '会员档案已删除。', 'tanzanite-settings' ) ) );
	}

	/**
	 * 获取集合参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_collection_params() {
		return array(
			'page'       => array(
				'type'    => 'integer',
				'default' => 1,
			),
			'per_page'   => array(
				'type'    => 'integer',
				'default' => 20,
			),
			'search'     => array(
				'type'    => 'string',
				'default' => '',
			),
			'min_points' => array(
				'type'    => 'integer',
				'default' => 0,
			),
		);
	}

	/**
	 * 获取更新参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_update_params() {
		return array(
			'id'              => array(
				'validate_callback' => 'is_numeric',
			),
			'full_name'       => array(
				'type' => 'string',
			),
			'phone'           => array(
				'type' => 'string',
			),
			'country'         => array(
				'type' => 'string',
			),
			'address'         => array(
				'type' => 'string',
			),
			'brand'           => array(
				'type' => 'string',
			),
			'points'          => array(
				'type' => 'integer',
			),
			'marketing_optin' => array(
				'type' => 'integer',
			),
			'notes'           => array(
				'type' => 'string',
			),
		);
	}
}
