<?php
/**
 * Gift Cards REST API Controller
 *
 * 处理礼品卡相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 礼品卡 REST API 控制器
 *
 * 提供礼品卡的 CRUD 操作
 */
class Tanzanite_REST_Giftcards_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'giftcards';

	/**
	 * 礼品卡表名
	 *
	 * @var string
	 */
	private $giftcards_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->giftcards_table = $wpdb->prefix . 'tanz_giftcards';
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
					'permission_callback' => '__return_true',
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => '__return_true',
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个礼品卡
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => '__return_true',
					'args'                => array(
						'id' => array(
							'validate_callback' => function( $param ) { return is_numeric( $param ); },
						),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => '__return_true',
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => function() { return true; },
					'args'                => array(
						'id' => array(
							'validate_callback' => function( $param ) { return is_numeric( $param ); },
						),
					),
				),
			)
		);

		// 积分兑换礼品卡
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/giftcards/redeem
		// Headers: X-WP-Nonce: {nonce}
		// Body: { "points": 1000, "giftcard_value": 10 }
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/redeem',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'redeem_with_points' ),
				'permission_callback' => 'is_user_logged_in',
				'args'                => array(
					'points'         => array(
						'type'     => 'integer',
						'required' => true,
						'minimum'  => 1,
					),
					'giftcard_value' => array(
						'type'     => 'number',
						'required' => true,
						'minimum'  => 0.01,
					),
				),
			)
		);

		// 查询当前用户的礼品卡列表
		// 前端 Nuxt 使用：GET /wp-json/tanzanite/v1/giftcards/my
		// Headers: X-WP-Nonce: {nonce}
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/my',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( $this, 'get_my_giftcards' ),
				'permission_callback' => 'is_user_logged_in',
			)
		);

		// 赠送礼品卡给其他用户
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/giftcards/{id}/transfer
		// Headers: X-WP-Nonce: {nonce}
		// Body: { "recipient_email": "user@example.com" }
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)/transfer',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'transfer_giftcard' ),
				'permission_callback' => 'is_user_logged_in',
				'args'                => array(
					'id'              => array(
						'validate_callback' => 'is_numeric',
					),
					'recipient_email' => array(
						'type'     => 'string',
						'required' => true,
						'format'   => 'email',
					),
				),
			)
		);

		// 验证礼品卡（用于结账页面）
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/giftcards/validate
		// Headers: X-WP-Nonce: {nonce}
		// Body: { "card_code": "REDEEM-ABC123XYZ" }
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/validate',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'validate_giftcard' ),
				'permission_callback' => 'is_user_logged_in',
				'args'                => array(
					'card_code' => array(
						'type'     => 'string',
						'required' => true,
					),
				),
			)
		);

		// 应用礼品卡到订单（在订单创建时调用）
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/giftcards/apply
		// Headers: X-WP-Nonce: {nonce}
		// Body: { "card_code": "REDEEM-ABC123XYZ", "order_id": 123, "amount": 50 }
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/apply',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'apply_giftcard' ),
				'permission_callback' => 'is_user_logged_in',
				'args'                => array(
					'card_code' => array(
						'type'     => 'string',
						'required' => true,
					),
					'order_id'  => array(
						'type'     => 'integer',
						'required' => true,
					),
					'amount'    => array(
						'type'     => 'number',
						'required' => true,
						'minimum'  => 0.01,
					),
				),
			)
		);
	}

	/**
	 * 获取礼品卡列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		$results = $wpdb->get_results( "SELECT * FROM {$this->giftcards_table} ORDER BY id DESC LIMIT 100", ARRAY_A );

		return $this->respond_success( array( 'items' => $results ?: array() ) );
	}

	/**
	 * 获取单个礼品卡
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		global $wpdb;

		$id  = (int) $request->get_param( 'id' );
		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->giftcards_table} WHERE id = %d", $id ), ARRAY_A );

		if ( ! $row ) {
			return $this->respond_error( 'giftcard_not_found', __( '未找到礼品卡。', 'tanzanite-settings' ), 404 );
		}

		return $this->respond_success( $row );
	}

	/**
	 * 创建礼品卡
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$data = array(
			'card_code'      => sanitize_text_field( $request->get_param( 'card_code' ) ?? '' ),
			'balance'        => floatval( $request->get_param( 'balance' ) ?? 0 ),
			'original_value' => floatval( $request->get_param( 'original_value' ) ?? 0 ),
			'currency'       => sanitize_text_field( $request->get_param( 'currency' ) ?? 'USD' ),
			'owner_user_id'  => intval( $request->get_param( 'owner_user_id' ) ?? 0 ),
			'points_spent'   => intval( $request->get_param( 'points_spent' ) ?? 0 ),
			'status'         => sanitize_text_field( $request->get_param( 'status' ) ?? 'active' ),
			'created_at'     => current_time( 'mysql' ),
			'updated_at'     => current_time( 'mysql' ),
		);
		
		// 只在字段存在时才添加 cover_image
		if ( $request->has_param( 'cover_image' ) ) {
			global $wpdb;
			$columns = $wpdb->get_col( "DESCRIBE {$this->giftcards_table}" );
			if ( in_array( 'cover_image', $columns, true ) ) {
				$data['cover_image'] = esc_url_raw( $request->get_param( 'cover_image' ) );
			}
		}

		$inserted = $wpdb->insert( $this->giftcards_table, $data );

		if ( false === $inserted ) {
			// 记录详细的数据库错误
			error_log( 'Giftcard creation failed. Table: ' . $this->giftcards_table );
			error_log( 'Data: ' . print_r( $data, true ) );
			error_log( 'DB Error: ' . $wpdb->last_error );
			
			return $this->respond_error( 
				'failed_create_giftcard', 
				__( '创建礼品卡失败。', 'tanzanite-settings' ) . ' DB Error: ' . $wpdb->last_error, 
				500 
			);
		}

		$id = $wpdb->insert_id;

		$this->log_audit( 'create', 'giftcard', $id, array( 'card_code' => $data['card_code'] ), $request );

		return $this->respond_success(
			array(
				'id'      => $id,
				'message' => __( '创建成功', 'tanzanite-settings' ),
			),
			201
		);
	}

	/**
	 * 更新礼品卡
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$id = (int) $request->get_param( 'id' );

		// 验证礼品卡是否存在
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->giftcards_table} WHERE id = %d", $id ) );
		if ( ! $exists ) {
			return $this->respond_error( 'giftcard_not_found', __( '未找到礼品卡。', 'tanzanite-settings' ), 404 );
		}

		$data = array(
			'updated_at' => current_time( 'mysql' ),
		);

		if ( $request->has_param( 'balance' ) ) {
			$data['balance'] = floatval( $request->get_param( 'balance' ) );
		}

		if ( $request->has_param( 'status' ) ) {
			$data['status'] = sanitize_text_field( $request->get_param( 'status' ) );
		}

		if ( $request->has_param( 'points_spent' ) ) {
			$data['points_spent'] = intval( $request->get_param( 'points_spent' ) );
		}

		if ( $request->has_param( 'cover_image' ) ) {
			$data['cover_image'] = esc_url_raw( $request->get_param( 'cover_image' ) );
		}

		$updated = $wpdb->update( $this->giftcards_table, $data, array( 'id' => $id ) );

		if ( false === $updated ) {
			return $this->respond_error( 'failed_update_giftcard', __( '更新礼品卡失败。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'update', 'giftcard', $id, $data, $request );

		return $this->respond_success( array( 'message' => __( '更新成功', 'tanzanite-settings' ) ) );
	}

	/**
	 * 删除礼品卡
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		error_log( 'Delete giftcard called. ID: ' . $request->get_param( 'id' ) );
		global $wpdb;

		$id = (int) $request->get_param( 'id' );

		// 验证礼品卡是否存在
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->giftcards_table} WHERE id = %d", $id ) );
		if ( ! $exists ) {
			return $this->respond_error( 'giftcard_not_found', __( '未找到礼品卡。', 'tanzanite-settings' ), 404 );
		}

		$deleted = $wpdb->delete( $this->giftcards_table, array( 'id' => $id ) );

		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_giftcard', __( '删除礼品卡失败。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'giftcard', $id, array(), $request );

		return $this->respond_success( array( 'message' => __( '删除成功', 'tanzanite-settings' ) ) );
	}

	/**
	 * 获取创建参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_create_params() {
		return array(
			'card_code'      => array(
				'type'     => 'string',
				'required' => true,
			),
			'balance'        => array(
				'type'    => 'number',
				'default' => 0,
			),
			'original_value' => array(
				'type'    => 'number',
				'default' => 0,
			),
			'currency'       => array(
				'type'    => 'string',
				'default' => 'CNY',
			),
			'owner_user_id'  => array(
				'type'    => 'integer',
				'default' => 0,
			),
			'points_spent'   => array(
				'type'    => 'integer',
				'default' => 0,
			),
			'status'         => array(
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
			'id'      => array(
				'validate_callback' => 'is_numeric',
			),
			'balance' => array(
				'type' => 'number',
			),
			'status'  => array(
				'type' => 'string',
			),
		);
	}

	/**
	 * 积分兑换礼品卡
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function redeem_with_points( $request ) {
		global $wpdb;

		$user_id         = get_current_user_id();
		$points_to_spend = intval( $request->get_param( 'points' ) );
		$giftcard_value  = floatval( $request->get_param( 'giftcard_value' ) );

		// 1. 检查功能是否启用
		if ( '1' !== get_option( 'tz_redeem_enabled', '1' ) ) {
			return $this->respond_error( 'redeem_disabled', __( '积分兑换功能已关闭', 'tanzanite-settings' ), 403 );
		}

		// 2. 检查用户积分余额
		$current_points = absint( get_user_meta( $user_id, 'loyalty_points', true ) );
		if ( $current_points < $points_to_spend ) {
			return $this->respond_error( 'insufficient_points', __( '积分不足', 'tanzanite-settings' ), 400 );
		}

		// 3. 生成礼品卡卡号
		$card_code = 'REDEEM-' . strtoupper( wp_generate_password( 12, false ) );

		// 7. 计算过期时间
		$expiry_days = absint( get_option( 'tz_redeem_card_expiry_days', '365' ) );
		$expires_at  = $expiry_days > 0 ? gmdate( 'Y-m-d H:i:s', strtotime( "+{$expiry_days} days" ) ) : null;

		// 8. 创建礼品卡
		$inserted = $wpdb->insert(
			$this->giftcards_table,
			array(
				'card_code'      => $card_code,
				'balance'        => $giftcard_value,
				'original_value' => $giftcard_value,
				'currency'       => 'USD',
				'owner_user_id'  => $user_id,
				'points_spent'   => $points_to_spend,
				'status'         => 'active',
				'expires_at'     => $expires_at,
				'created_at'     => current_time( 'mysql' ),
				'updated_at'     => current_time( 'mysql' ),
			)
		);

		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_giftcard', __( '创建礼品卡失败', 'tanzanite-settings' ), 500 );
		}

		$giftcard_id = $wpdb->insert_id;

		// 9. 扣除用户积分
		$new_points = $current_points - $points_to_spend;
		update_user_meta( $user_id, 'loyalty_points', $new_points );

		// 10. 记录积分交易
		$rewards_table = $wpdb->prefix . 'tanz_rewards_transactions';
		$wpdb->insert(
			$rewards_table,
			array(
				'user_id'      => $user_id,
				'related_type' => 'giftcard',
				'related_id'   => $giftcard_id,
				'action'       => 'redeem',
				'points_delta' => -$points_to_spend,
				'amount_delta' => $giftcard_value,
				'notes'        => sprintf( '兑换礼品卡 %s', $card_code ),
				'created_at'   => current_time( 'mysql' ),
			)
		);

		// 11. 记录审计日志
		$this->log_audit( 'redeem', 'giftcard', $giftcard_id, array( 'card_code' => $card_code, 'points_spent' => $points_to_spend ), $request );

		// 12. 返回结果
		return $this->respond_success(
			array(
				'giftcard_id'       => $giftcard_id,
				'card_code'         => $card_code,
				'balance'           => $giftcard_value,
				'points_spent'      => $points_to_spend,
				'points_remaining'  => $new_points,
				'expires_at'        => $expires_at,
				'message'           => __( '兑换成功', 'tanzanite-settings' ),
			),
			201
		);
	}

	/**
	 * 获取当前用户的礼品卡列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_my_giftcards( $request ) {
		global $wpdb;

		$user_id = get_current_user_id();

		$results = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->giftcards_table} 
				WHERE owner_user_id = %d 
				ORDER BY created_at DESC",
				$user_id
			),
			ARRAY_A
		);

		return $this->respond_success( array( 'items' => $results ?: array() ) );
	}

	/**
	 * 赠送礼品卡给其他用户
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function transfer_giftcard( $request ) {
		global $wpdb;

		$user_id         = get_current_user_id();
		$giftcard_id     = intval( $request->get_param( 'id' ) );
		$recipient_email = sanitize_email( $request->get_param( 'recipient_email' ) );

		// 1. 验证礼品卡是否存在且属于当前用户
		$giftcard = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->giftcards_table} WHERE id = %d",
				$giftcard_id
			),
			ARRAY_A
		);

		if ( ! $giftcard ) {
			return $this->respond_error( 'giftcard_not_found', __( '礼品卡不存在', 'tanzanite-settings' ), 404 );
		}

		if ( intval( $giftcard['owner_user_id'] ) !== $user_id ) {
			return $this->respond_error( 'not_owner', __( '您不是该礼品卡的持有者', 'tanzanite-settings' ), 403 );
		}

		// 2. 检查礼品卡状态
		if ( 'active' !== $giftcard['status'] ) {
			return $this->respond_error( 'invalid_status', __( '该礼品卡不可转赠', 'tanzanite-settings' ), 400 );
		}

		if ( floatval( $giftcard['balance'] ) <= 0 ) {
			return $this->respond_error( 'no_balance', __( '礼品卡余额为零，无法转赠', 'tanzanite-settings' ), 400 );
		}

		// 3. 检查是否过期
		if ( $giftcard['expires_at'] && strtotime( $giftcard['expires_at'] ) < time() ) {
			return $this->respond_error( 'expired', __( '礼品卡已过期', 'tanzanite-settings' ), 400 );
		}

		// 4. 查找接收者用户
		$recipient = get_user_by( 'email', $recipient_email );
		if ( ! $recipient ) {
			return $this->respond_error( 'recipient_not_found', __( '接收者不存在，请确认邮箱地址正确', 'tanzanite-settings' ), 404 );
		}

		if ( $recipient->ID === $user_id ) {
			return $this->respond_error( 'self_transfer', __( '不能转赠给自己', 'tanzanite-settings' ), 400 );
		}

		// 5. 更新礼品卡持有者
		$updated = $wpdb->update(
			$this->giftcards_table,
			array(
				'owner_user_id' => $recipient->ID,
				'updated_at'    => current_time( 'mysql' ),
			),
			array( 'id' => $giftcard_id )
		);

		if ( false === $updated ) {
			return $this->respond_error( 'transfer_failed', __( '转赠失败', 'tanzanite-settings' ), 500 );
		}

		// 6. 记录审计日志
		$this->log_audit(
			'transfer',
			'giftcard',
			$giftcard_id,
			array(
				'from_user_id' => $user_id,
				'to_user_id'   => $recipient->ID,
				'to_email'     => $recipient_email,
			),
			$request
		);

		// 7. 发送通知邮件给接收者（可选）
		// TODO: 实现邮件通知功能

		return $this->respond_success(
			array(
				'message'        => __( '转赠成功', 'tanzanite-settings' ),
				'recipient_name' => $recipient->display_name,
				'recipient_email' => $recipient_email,
			)
		);
	}

	/**
	 * 验证礼品卡（用于结账页面）
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function validate_giftcard( $request ) {
		global $wpdb;

		$user_id   = get_current_user_id();
		$card_code = sanitize_text_field( $request->get_param( 'card_code' ) );

		// 1. 查找礼品卡
		$giftcard = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->giftcards_table} WHERE card_code = %s",
				$card_code
			),
			ARRAY_A
		);

		if ( ! $giftcard ) {
			return $this->respond_error( 'invalid_card', __( '礼品卡不存在', 'tanzanite-settings' ), 404 );
		}

		// 2. 检查持有者
		if ( intval( $giftcard['owner_user_id'] ) !== $user_id ) {
			return $this->respond_error( 'not_owner', __( '该礼品卡不属于您', 'tanzanite-settings' ), 403 );
		}

		// 3. 检查状态
		if ( 'active' !== $giftcard['status'] ) {
			return $this->respond_error( 'inactive', __( '礼品卡不可用', 'tanzanite-settings' ), 400 );
		}

		// 4. 检查余额
		$balance = floatval( $giftcard['balance'] );
		if ( $balance <= 0 ) {
			return $this->respond_error( 'no_balance', __( '礼品卡余额不足', 'tanzanite-settings' ), 400 );
		}

		// 5. 检查是否过期
		if ( $giftcard['expires_at'] && strtotime( $giftcard['expires_at'] ) < time() ) {
			return $this->respond_error( 'expired', __( '礼品卡已过期', 'tanzanite-settings' ), 400 );
		}

		// 6. 返回礼品卡信息
		return $this->respond_success(
			array(
				'valid'          => true,
				'id'             => intval( $giftcard['id'] ),
				'card_code'      => $giftcard['card_code'],
				'balance'        => $balance,
				'original_value' => floatval( $giftcard['original_value'] ),
				'currency'       => $giftcard['currency'],
				'expires_at'     => $giftcard['expires_at'],
			)
		);
	}

	/**
	 * 应用礼品卡到订单
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function apply_giftcard( $request ) {
		global $wpdb;

		$user_id   = get_current_user_id();
		$card_code = sanitize_text_field( $request->get_param( 'card_code' ) );
		$order_id  = intval( $request->get_param( 'order_id' ) );
		$amount    = floatval( $request->get_param( 'amount' ) );

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

		// 2. 查找礼品卡
		$giftcard = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->giftcards_table} WHERE card_code = %s",
				$card_code
			),
			ARRAY_A
		);

		if ( ! $giftcard ) {
			return $this->respond_error( 'invalid_card', __( '礼品卡不存在', 'tanzanite-settings' ), 404 );
		}

		// 3. 验证礼品卡
		if ( intval( $giftcard['owner_user_id'] ) !== $user_id ) {
			return $this->respond_error( 'not_owner', __( '该礼品卡不属于您', 'tanzanite-settings' ), 403 );
		}

		if ( 'active' !== $giftcard['status'] ) {
			return $this->respond_error( 'inactive', __( '礼品卡不可用', 'tanzanite-settings' ), 400 );
		}

		$balance = floatval( $giftcard['balance'] );
		if ( $balance <= 0 ) {
			return $this->respond_error( 'no_balance', __( '礼品卡余额不足', 'tanzanite-settings' ), 400 );
		}

		if ( $giftcard['expires_at'] && strtotime( $giftcard['expires_at'] ) < time() ) {
			return $this->respond_error( 'expired', __( '礼品卡已过期', 'tanzanite-settings' ), 400 );
		}

		// 4. 验证使用金额
		if ( $amount > $balance ) {
			return $this->respond_error( 'insufficient_balance', __( '礼品卡余额不足', 'tanzanite-settings' ), 400 );
		}

		// 5. 扣除礼品卡余额
		$new_balance = $balance - $amount;
		$new_status  = $new_balance <= 0 ? 'used' : 'active';

		$updated = $wpdb->update(
			$this->giftcards_table,
			array(
				'balance'    => $new_balance,
				'status'     => $new_status,
				'updated_at' => current_time( 'mysql' ),
			),
			array( 'id' => $giftcard['id'] )
		);

		if ( false === $updated ) {
			return $this->respond_error( 'apply_failed', __( '应用礼品卡失败', 'tanzanite-settings' ), 500 );
		}

		// 6. 更新订单的礼品卡抵扣金额
		// 假设订单表有 giftcard_discount 字段
		$wpdb->update(
			$orders_table,
			array(
				'giftcard_discount' => $amount,
				'updated_at'        => current_time( 'mysql' ),
			),
			array( 'id' => $order_id )
		);

		// 7. 记录审计日志
		$this->log_audit(
			'apply',
			'giftcard',
			$giftcard['id'],
			array(
				'order_id' => $order_id,
				'amount'   => $amount,
			),
			$request
		);

		// 8. 返回结果
		return $this->respond_success(
			array(
				'message'         => __( '礼品卡应用成功', 'tanzanite-settings' ),
				'amount_applied'  => $amount,
				'remaining_balance' => $new_balance,
				'card_status'     => $new_status,
			)
		);
	}
}
