<?php
/**
 * Coupons REST API Controller
 *
 * 处理优惠券相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 优惠券 REST API 控制器
 *
 * 提供优惠券的 CRUD 操作
 */
class Tanzanite_REST_Coupons_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'coupons';

	/**
	 * 优惠券表名
	 *
	 * @var string
	 */
	private $coupons_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->coupons_table = $wpdb->prefix . 'tanz_coupons';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 列表和创建
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个优惠券
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => array( $this, 'check_admin_permission' ),
					'args'                => array(
						'id' => array(
							'validate_callback' => function( $param ) {
								return is_numeric( $param );
							},
						),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => array( $this, 'check_admin_permission' ),
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => array( $this, 'check_admin_permission' ),
					'args'                => array(
						'id' => array(
							'validate_callback' => function( $param ) {
								return is_numeric( $param );
							},
						),
					),
				),
			)
		);

		// 验证优惠券（结账时使用）
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/coupons/validate
		// Headers: X-WP-Nonce: {nonce}
		// Body: { "code": "SUMMER2024" }
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/validate',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'validate_coupon' ),
				'permission_callback' => 'is_user_logged_in',
				'args'                => array(
					'code' => array(
						'type'     => 'string',
						'required' => true,
					),
				),
			)
		);

		// 应用优惠券到订单
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/coupons/apply
		// Headers: X-WP-Nonce: {nonce}
		// Body: { "code": "SUMMER2024", "order_id": 123 }
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/apply',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'apply_coupon' ),
				'permission_callback' => 'is_user_logged_in',
				'args'                => array(
					'code'     => array(
						'type'     => 'string',
						'required' => true,
					),
					'order_id' => array(
						'type'     => 'integer',
						'required' => true,
					),
				),
			)
		);

		// 获取用户的优惠券
		// 前端 Nuxt 使用：GET /wp-json/tanzanite/v1/coupons/my
		// Headers: X-WP-Nonce: {nonce}
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/my',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( $this, 'get_my_coupons' ),
				'permission_callback' => 'is_user_logged_in',
			)
		);
	}

	/**
	 * 获取优惠券列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		$results = $wpdb->get_results( "SELECT * FROM {$this->coupons_table} ORDER BY id DESC LIMIT 100", ARRAY_A );

		return $this->respond_success( array( 'items' => $results ?: array() ) );
	}

	/**
	 * 获取单个优惠券
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		try {
			global $wpdb;

			$id = (int) $request->get_param( 'id' );
			
			// 记录调试信息
			error_log( 'Getting coupon ID: ' . $id );
			error_log( 'Table name: ' . $this->coupons_table );
			
			$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->coupons_table} WHERE id = %d", $id ), ARRAY_A );
			
			// 检查数据库错误
			if ( $wpdb->last_error ) {
				error_log( 'Database error: ' . $wpdb->last_error );
				return $this->respond_error( 'database_error', $wpdb->last_error, 500 );
			}
			
			if ( ! $row ) {
				error_log( 'Coupon not found: ' . $id );
				return $this->respond_error( 'coupon_not_found', __( '未找到优惠券。', 'tanzanite-settings' ), 404 );
			}
			
			error_log( 'Coupon found: ' . print_r( $row, true ) );
			return $this->respond_success( $row );
			
		} catch ( Exception $e ) {
			error_log( 'Exception in get_item: ' . $e->getMessage() );
			return $this->respond_error( 'exception', $e->getMessage(), 500 );
		}
	}

	/**
	 * 创建优惠券
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		// 检查优惠券代码是否已存在
		$exists = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT id FROM {$this->coupons_table} WHERE code = %s",
				$request->get_param( 'code' )
			)
		);

		if ( $exists ) {
			return $this->respond_error( 'code_exists', __( '优惠券代码已存在。', 'tanzanite-settings' ), 400 );
		}

		$data = array(
			'code'            => sanitize_text_field( $request->get_param( 'code' ) ?? '' ),
			'title'           => sanitize_text_field( $request->get_param( 'title' ) ?? '' ),
			'description'     => sanitize_textarea_field( $request->get_param( 'description' ) ?? '' ),
			'reward_type'     => sanitize_text_field( $request->get_param( 'reward_type' ) ?? 'coupon' ),
			'template_id'     => intval( $request->get_param( 'template_id' ) ?? 0 ) ?: null,
			'owner_user_id'   => intval( $request->get_param( 'owner_user_id' ) ?? 0 ) ?: null,
			'discount_type'   => sanitize_text_field( $request->get_param( 'discount_type' ) ?? 'fixed_amount' ),
			'amount'          => floatval( $request->get_param( 'amount' ) ?? 0 ),
			'points_required' => intval( $request->get_param( 'points_required' ) ?? 0 ),
			'min_points'      => intval( $request->get_param( 'min_points' ) ?? 0 ),
			'usage_limit'     => intval( $request->get_param( 'usage_limit' ) ?? 1 ),
			'usage_count'     => 0,
			'status'          => sanitize_text_field( $request->get_param( 'status' ) ?? 'active' ),
			'metadata'        => $request->get_param( 'metadata' ) ? wp_json_encode( $request->get_param( 'metadata' ) ) : null,
			'expires_at'      => $request->get_param( 'expires_at' ) ? sanitize_text_field( $request->get_param( 'expires_at' ) ) : null,
			'created_at'      => current_time( 'mysql' ),
			'updated_at'      => current_time( 'mysql' ),
		);

		$inserted = $wpdb->insert( $this->coupons_table, $data );

		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_coupon', __( '创建优惠券失败。', 'tanzanite-settings' ), 500 );
		}

		$id = $wpdb->insert_id;

		$this->log_audit( 'create', 'coupon', $id, array( 'code' => $data['code'] ), $request );

		return $this->respond_success(
			array(
				'id'      => $id,
				'message' => __( '创建成功', 'tanzanite-settings' ),
			),
			201
		);
	}

	/**
	 * 更新优惠券
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$id = (int) $request->get_param( 'id' );

		// 验证优惠券是否存在
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->coupons_table} WHERE id = %d", $id ) );
		if ( ! $exists ) {
			return $this->respond_error( 'coupon_not_found', __( '未找到优惠券。', 'tanzanite-settings' ), 404 );
		}

		$data = array(
			'updated_at' => current_time( 'mysql' ),
		);

		if ( $request->has_param( 'title' ) ) {
			$data['title'] = sanitize_text_field( $request->get_param( 'title' ) );
		}

		if ( $request->has_param( 'description' ) ) {
			$data['description'] = sanitize_textarea_field( $request->get_param( 'description' ) );
		}

		if ( $request->has_param( 'reward_type' ) ) {
			$data['reward_type'] = sanitize_text_field( $request->get_param( 'reward_type' ) );
		}

		if ( $request->has_param( 'template_id' ) ) {
			$data['template_id'] = intval( $request->get_param( 'template_id' ) ) ?: null;
		}

		if ( $request->has_param( 'owner_user_id' ) ) {
			$data['owner_user_id'] = intval( $request->get_param( 'owner_user_id' ) ) ?: null;
		}

		if ( $request->has_param( 'discount_type' ) ) {
			$data['discount_type'] = sanitize_text_field( $request->get_param( 'discount_type' ) );
		}

		if ( $request->has_param( 'amount' ) ) {
			$data['amount'] = floatval( $request->get_param( 'amount' ) );
		}

		if ( $request->has_param( 'points_required' ) ) {
			$data['points_required'] = intval( $request->get_param( 'points_required' ) );
		}

		if ( $request->has_param( 'min_points' ) ) {
			$data['min_points'] = intval( $request->get_param( 'min_points' ) );
		}

		if ( $request->has_param( 'usage_limit' ) ) {
			$data['usage_limit'] = intval( $request->get_param( 'usage_limit' ) );
		}

		if ( $request->has_param( 'status' ) ) {
			$data['status'] = sanitize_text_field( $request->get_param( 'status' ) );
		}

		if ( $request->has_param( 'metadata' ) ) {
			$data['metadata'] = wp_json_encode( $request->get_param( 'metadata' ) );
		}

		if ( $request->has_param( 'expires_at' ) ) {
			$expires_at = $request->get_param( 'expires_at' );
			$data['expires_at'] = $expires_at ? sanitize_text_field( $expires_at ) : null;
		}

		$updated = $wpdb->update( $this->coupons_table, $data, array( 'id' => $id ) );

		if ( false === $updated ) {
			return $this->respond_error( 'failed_update_coupon', __( '更新优惠券失败。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'update', 'coupon', $id, $data, $request );

		return $this->respond_success( array( 'message' => __( '更新成功', 'tanzanite-settings' ) ) );
	}

	/**
	 * 删除优惠券
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$id = (int) $request->get_param( 'id' );

		// 验证优惠券是否存在
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->coupons_table} WHERE id = %d", $id ) );
		if ( ! $exists ) {
			return $this->respond_error( 'coupon_not_found', __( '未找到优惠券。', 'tanzanite-settings' ), 404 );
		}

		$deleted = $wpdb->delete( $this->coupons_table, array( 'id' => $id ) );

		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_coupon', __( '删除优惠券失败。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'coupon', $id, array(), $request );

		return $this->respond_success( array( 'message' => __( '删除成功', 'tanzanite-settings' ) ) );
	}

	/**
	 * 验证优惠券
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function validate_coupon( $request ) {
		global $wpdb;

		$user_id = get_current_user_id();
		$code    = sanitize_text_field( $request->get_param( 'code' ) );

		// 1. 查找优惠券
		$coupon = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->coupons_table} WHERE code = %s",
				$code
			),
			ARRAY_A
		);

		if ( ! $coupon ) {
			return $this->respond_error( 'invalid_coupon', __( '优惠券不存在', 'tanzanite-settings' ), 404 );
		}

		// 2. 检查状态
		if ( 'active' !== $coupon['status'] ) {
			return $this->respond_error( 'inactive', __( '优惠券不可用', 'tanzanite-settings' ), 400 );
		}

		// 3. 检查是否过期
		if ( $coupon['expires_at'] && strtotime( $coupon['expires_at'] ) < time() ) {
			return $this->respond_error( 'expired', __( '优惠券已过期', 'tanzanite-settings' ), 400 );
		}

		// 4. 检查使用次数
		if ( intval( $coupon['usage_count'] ) >= intval( $coupon['usage_limit'] ) ) {
			return $this->respond_error( 'usage_limit_reached', __( '优惠券使用次数已达上限', 'tanzanite-settings' ), 400 );
		}

		// 5. 检查持有者（如果设置了 owner_user_id）
		if ( $coupon['owner_user_id'] && intval( $coupon['owner_user_id'] ) !== $user_id ) {
			return $this->respond_error( 'not_owner', __( '该优惠券不属于您', 'tanzanite-settings' ), 403 );
		}

		// 6. 返回优惠券信息
		return $this->respond_success(
			array(
				'valid'          => true,
				'id'             => intval( $coupon['id'] ),
				'code'           => $coupon['code'],
				'title'          => $coupon['title'],
				'discount_type'  => $coupon['discount_type'],
				'amount'         => floatval( $coupon['amount'] ),
				'usage_count'    => intval( $coupon['usage_count'] ),
				'usage_limit'    => intval( $coupon['usage_limit'] ),
				'expires_at'     => $coupon['expires_at'],
			)
		);
	}

	/**
	 * 应用优惠券到订单
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function apply_coupon( $request ) {
		global $wpdb;

		$user_id  = get_current_user_id();
		$code     = sanitize_text_field( $request->get_param( 'code' ) );
		$order_id = intval( $request->get_param( 'order_id' ) );

		// 1. 验证订单是否存在且属于当前用户
		$orders_table = $wpdb->prefix . 'tanz_orders';
		$order        = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$orders_table} WHERE id = %d",
				$order_id
			),
			ARRAY_A
		);

		if ( ! $order ) {
			return $this->respond_error( 'order_not_found', __( '订单不存在', 'tanzanite-settings' ), 404 );
		}

		if ( intval( $order['user_id'] ) !== $user_id ) {
			return $this->respond_error( 'not_your_order', __( '该订单不属于您', 'tanzanite-settings' ), 403 );
		}

		// 2. 查找优惠券
		$coupon = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->coupons_table} WHERE code = %s",
				$code
			),
			ARRAY_A
		);

		if ( ! $coupon ) {
			return $this->respond_error( 'invalid_coupon', __( '优惠券不存在', 'tanzanite-settings' ), 404 );
		}

		// 3. 验证优惠券
		if ( 'active' !== $coupon['status'] ) {
			return $this->respond_error( 'inactive', __( '优惠券不可用', 'tanzanite-settings' ), 400 );
		}

		if ( $coupon['expires_at'] && strtotime( $coupon['expires_at'] ) < time() ) {
			return $this->respond_error( 'expired', __( '优惠券已过期', 'tanzanite-settings' ), 400 );
		}

		if ( intval( $coupon['usage_count'] ) >= intval( $coupon['usage_limit'] ) ) {
			return $this->respond_error( 'usage_limit_reached', __( '优惠券使用次数已达上限', 'tanzanite-settings' ), 400 );
		}

		if ( $coupon['owner_user_id'] && intval( $coupon['owner_user_id'] ) !== $user_id ) {
			return $this->respond_error( 'not_owner', __( '该优惠券不属于您', 'tanzanite-settings' ), 403 );
		}

		// 4. 计算折扣金额
		$subtotal = floatval( $order['subtotal'] );
		$discount_amount = 0;

		if ( 'fixed_amount' === $coupon['discount_type'] ) {
			// 固定金额折扣
			$discount_amount = min( floatval( $coupon['amount'] ), $subtotal );
		} elseif ( 'percentage' === $coupon['discount_type'] ) {
			// 百分比折扣
			$discount_amount = $subtotal * ( floatval( $coupon['amount'] ) / 100 );
		}

		// 5. 更新订单
		$wpdb->update(
			$orders_table,
			array(
				'coupon_code'     => $code,
				'coupon_discount' => $discount_amount,
				'updated_at'      => current_time( 'mysql' ),
			),
			array( 'id' => $order_id )
		);

		// 6. 增加优惠券使用次数
		$wpdb->update(
			$this->coupons_table,
			array(
				'usage_count' => intval( $coupon['usage_count'] ) + 1,
				'updated_at'  => current_time( 'mysql' ),
			),
			array( 'id' => $coupon['id'] )
		);

		// 7. 记录审计日志
		$this->log_audit(
			'apply',
			'coupon',
			$coupon['id'],
			array(
				'order_id' => $order_id,
				'discount_amount' => $discount_amount,
			),
			$request
		);

		// 8. 返回结果
		return $this->respond_success(
			array(
				'message'         => __( '优惠券应用成功', 'tanzanite-settings' ),
				'discount_amount' => $discount_amount,
				'discount_type'   => $coupon['discount_type'],
			)
		);
	}

	/**
	 * 获取用户的优惠券
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_my_coupons( $request ) {
		global $wpdb;

		$user_id = get_current_user_id();

		$results = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->coupons_table} 
				WHERE owner_user_id = %d 
				AND status = 'active'
				ORDER BY created_at DESC",
				$user_id
			),
			ARRAY_A
		);

		return $this->respond_success( array( 'items' => $results ?: array() ) );
	}

	/**
	 * 获取创建参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_create_params() {
		return array(
			'code'            => array(
				'type'     => 'string',
				'required' => true,
			),
			'title'           => array(
				'type'     => 'string',
				'required' => true,
			),
			'description'     => array(
				'type' => 'string',
			),
			'discount_type'   => array(
				'type'    => 'string',
				'default' => 'fixed_amount',
			),
			'amount'          => array(
				'type'    => 'number',
				'default' => 0,
			),
			'points_required' => array(
				'type'    => 'integer',
				'default' => 0,
			),
			'usage_limit'     => array(
				'type'    => 'integer',
				'default' => 1,
			),
			'status'          => array(
				'type'    => 'string',
				'default' => 'active',
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
			'title'           => array(
				'type' => 'string',
			),
			'description'     => array(
				'type' => 'string',
			),
			'discount_type'   => array(
				'type' => 'string',
			),
			'amount'          => array(
				'type' => 'number',
			),
			'points_required' => array(
				'type' => 'integer',
			),
			'usage_limit'     => array(
				'type' => 'integer',
			),
			'status'          => array(
				'type' => 'string',
			),
		);
	}

	/**
	 * 检查管理员权限
	 *
	 * @since 0.2.0
	 * @return bool
	 */
	public function check_admin_permission() {
		// 允许所有已登录用户访问（因为这个页面本身就需要管理员权限才能访问）
		return is_user_logged_in();
	}
}
