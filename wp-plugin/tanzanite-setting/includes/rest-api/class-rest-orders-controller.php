<?php
/**
 * Orders REST API Controller
 *
 * @package Tanzanite_Settings
 * @subpackage REST_API
 * @since 0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * Orders REST API 控制器
 *
 * @since 0.2.0
 */
class Tanzanite_REST_Orders_Controller extends Tanzanite_REST_Controller {

	/**
	 * 订单表名
	 *
	 * @var string
	 */
	private $orders_table;

	/**
	 * 订单项表名
	 *
	 * @var string
	 */
	private $order_items_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		global $wpdb;
		$this->orders_table      = $wpdb->prefix . 'tanz_orders';
		$this->order_items_table = $wpdb->prefix . 'tanz_order_items';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// GET /orders - 列表
		// POST /orders - 创建
		// PUT /orders - 批量操作
		register_rest_route(
			'tanzanite/v1',
			'/orders',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => array(
						'page'              => array( 'type' => 'integer', 'default' => 1 ),
						'per_page'          => array( 'type' => 'integer', 'default' => 20 ),
						'status'            => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ),
						'payment_method'    => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'channel'           => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'tracking_provider' => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ),
						'customer_keyword'  => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'date_start'        => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'date_end'          => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
					),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => array( $this, 'create_item_permissions_check' ),
					'args'                => array(
						'channel'           => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'payment_method'    => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'total'             => array( 'type' => 'number', 'default' => 0 ),
						'subtotal'          => array( 'type' => 'number', 'default' => 0 ),
						'discount_total'    => array( 'type' => 'number', 'default' => 0 ),
						'shipping_total'    => array( 'type' => 'number', 'default' => 0 ),
						'status'            => array( 'type' => 'string', 'default' => 'pending', 'sanitize_callback' => 'sanitize_key' ),
						'tracking_provider' => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ),
						'tracking_number'   => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'items'             => array( 'type' => 'array', 'required' => true ),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'bulk_action' ),
					'permission_callback' => array( $this, 'bulk_action_permissions_check' ),
					'args'                => array(
						'action'  => array( 'type' => 'string', 'required' => true ),
						'ids'     => array( 'type' => 'array', 'required' => true ),
						'payload' => array( 'type' => 'object' ),
					),
				),
			)
		);

		// GET /orders/{id} - 获取单个
		// PUT /orders/{id} - 更新
		// DELETE /orders/{id} - 删除
		register_rest_route(
			'tanzanite/v1',
			'/orders/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => array(
						'id' => array( 'validate_callback' => 'is_numeric' ),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => array( $this, 'update_item_permissions_check' ),
					'args'                => array(
						'id'                => array( 'validate_callback' => 'is_numeric' ),
						'status'            => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ),
						'payment_method'    => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'channel'           => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'total'             => array( 'type' => 'number' ),
						'subtotal'          => array( 'type' => 'number' ),
						'discount_total'    => array( 'type' => 'number' ),
						'shipping_total'    => array( 'type' => 'number' ),
						'tracking_provider' => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ),
						'tracking_number'   => array( 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ),
						'items'             => array( 'type' => 'array' ),
					),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => array( $this, 'delete_item_permissions_check' ),
					'args'                => array(
						'id' => array( 'validate_callback' => 'is_numeric' ),
					),
				),
			)
		);

		// POST /orders/{id}/tracking - 同步物流
		register_rest_route(
			'tanzanite/v1',
			'/orders/(?P<id>\d+)/tracking',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'sync_tracking' ),
				'permission_callback' => array( $this, 'update_item_permissions_check' ),
				'args'                => array(
					'id' => array( 'validate_callback' => 'is_numeric' ),
				),
			)
		);
	}

	/**
	 * 权限检查 - 创建
	 *
	 * @since 0.2.0
	 */
	public function create_item_permissions_check( $request ) {
		return current_user_can( 'tanz_manage_orders' );
	}

	/**
	 * 权限检查 - 更新
	 *
	 * @since 0.2.0
	 */
	public function update_item_permissions_check( $request ) {
		return current_user_can( 'tanz_manage_orders' );
	}

	/**
	 * 权限检查 - 删除
	 *
	 * @since 0.2.0
	 */
	public function delete_item_permissions_check( $request ) {
		return current_user_can( 'tanz_manage_orders' );
	}

	/**
	 * 权限检查 - 批量操作
	 *
	 * @since 0.2.0
	 */
	public function bulk_action_permissions_check( $request ) {
		return current_user_can( 'tanz_bulk_orders' );
	}

	/**
	 * 获取订单列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		// 分页参数
		$page     = max( 1, (int) $request->get_param( 'page' ) );
		$per_page = (int) $request->get_param( 'per_page' );
		if ( $per_page <= 0 ) {
			$per_page = 20;
		}
		$per_page = min( 100, $per_page );
		$offset   = ( $page - 1 ) * $per_page;

		// 构建查询条件
		$where  = array( '1=1' );
		$params = array();
		$joins  = '';

		// 状态筛选
		$status = $request->get_param( 'status' );
		if ( $status ) {
			$status_key = sanitize_key( $status );
			$allowed_statuses = array( 'pending', 'paid', 'shipped', 'completed', 'cancelled', 'refunded' );
			if ( in_array( $status_key, $allowed_statuses, true ) ) {
				$where[]  = 'o.status = %s';
				$params[] = $status_key;
			}
		}

		// 渠道筛选
		$channel = $request->get_param( 'channel' );
		if ( $channel ) {
			$where[]  = 'o.channel = %s';
			$params[] = sanitize_text_field( $channel );
		}

		// 支付方式筛选
		$payment_method = $request->get_param( 'payment_method' );
		if ( $payment_method ) {
			$where[]  = 'o.payment_method = %s';
			$params[] = sanitize_text_field( $payment_method );
		}

		// 物流商筛选
		$tracking_provider = $request->get_param( 'tracking_provider' );
		if ( $tracking_provider ) {
			$where[]  = 'o.tracking_provider = %s';
			$params[] = sanitize_key( $tracking_provider );
		}

		// 日期范围筛选
		$date_start = $request->get_param( 'date_start' );
		if ( $date_start ) {
			$start_time = strtotime( sanitize_text_field( $date_start ) );
			if ( $start_time ) {
				$where[]  = 'o.created_at >= %s';
				$params[] = gmdate( 'Y-m-d H:i:s', $start_time );
			}
		}

		$date_end = $request->get_param( 'date_end' );
		if ( $date_end ) {
			$end_time = strtotime( sanitize_text_field( $date_end ) );
			if ( $end_time ) {
				$end_time = strtotime( '+1 day', $end_time ) - 1;
				$where[]  = 'o.created_at <= %s';
				$params[] = gmdate( 'Y-m-d H:i:s', $end_time );
			}
		}

		// 客户关键词搜索
		$customer_keyword = $request->get_param( 'customer_keyword' );
		if ( $customer_keyword ) {
			$joins   .= " LEFT JOIN {$wpdb->users} u ON u.ID = o.user_id";
			$like     = '%' . $wpdb->esc_like( $customer_keyword ) . '%';
			$where[]  = '(u.display_name LIKE %s OR u.user_email LIKE %s OR u.user_login LIKE %s)';
			$params[] = $like;
			$params[] = $like;
			$params[] = $like;
		}

		// 构建 SQL
		$base_sql  = "FROM {$this->orders_table} o{$joins}";
		$where_sql = $where ? 'WHERE ' . implode( ' AND ', $where ) : '';

		// 获取总数
		if ( $params ) {
			$total = (int) $wpdb->get_var( $wpdb->prepare( "SELECT COUNT(*) {$base_sql} {$where_sql}", ...$params ) );
		} else {
			$total = (int) $wpdb->get_var( "SELECT COUNT(*) {$base_sql} {$where_sql}" );
		}

		// 获取数据
		$data_sql     = "SELECT o.* {$base_sql} {$where_sql} ORDER BY o.id DESC LIMIT %d OFFSET %d";
		$data_params  = array_merge( $params, array( $per_page, $offset ) );
		$prepared_sql = $wpdb->prepare( $data_sql, ...$data_params );
		$rows         = $wpdb->get_results( $prepared_sql, ARRAY_A );

		// 格式化数据
		$items = array_map( array( $this, 'prepare_item_for_response' ), $rows );

		return $this->respond_success(
			array(
				'items' => $items,
				'meta'  => array(
					'page'        => $page,
					'per_page'    => $per_page,
					'total'       => $total,
					'total_pages' => $per_page ? (int) ceil( $total / $per_page ) : 0,
				),
			)
		);
	}

	/**
	 * 获取单个订单
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		$order_id = (int) $request->get_param( 'id' );

		// 获取订单数据
		$row = $this->fetch_order_row( $order_id );

		if ( ! $row ) {
			return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
		}

		// 返回详情模式的数据
		return $this->respond_success( $this->prepare_item_for_response( $row, 'detail' ) );
	}

	/**
	 * 创建订单
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		// 验证订单项
		$items = $this->sanitize_order_items( $request->get_param( 'items' ) );
		if ( is_wp_error( $items ) ) {
			return $this->respond_error( $items->get_error_code(), $items->get_error_message(), 400 );
		}

		// 准备订单数据
		$tracking_provider = $request->get_param( 'tracking_provider' );
		$tracking_number   = $request->get_param( 'tracking_number' );

		$data = array(
			'order_number'      => $this->generate_order_number(),
			'user_id'           => get_current_user_id(),
			'status'            => sanitize_key( $request->get_param( 'status' ) ?: 'pending' ),
			'payment_method'    => sanitize_text_field( $request->get_param( 'payment_method' ) ?: '' ),
			'channel'           => sanitize_text_field( $request->get_param( 'channel' ) ?: '' ),
			'total'             => (float) $request->get_param( 'total' ),
			'subtotal'          => (float) $request->get_param( 'subtotal' ),
			'discount_total'    => (float) $request->get_param( 'discount_total' ),
			'shipping_total'    => (float) $request->get_param( 'shipping_total' ),
			'points_used'       => 0,
			'currency'          => 'CNY',
			'tracking_provider' => $tracking_provider ? sanitize_key( $tracking_provider ) : '',
			'tracking_number'   => $tracking_number ? sanitize_text_field( $tracking_number ) : '',
			'meta'              => wp_json_encode( array( 'stage' => 'placeholder' ), JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES ),
		);

		$format_map = array(
			'%s', '%d', '%s', '%s', '%s', '%f', '%f', '%f', '%f', '%d', '%s', '%s', '%s', '%s',
		);

		// 应用状态时间戳
		$this->apply_status_timestamps( $data, $format_map, array( 'status' => 'pending' ), $data['status'] );

		// 插入订单
		$inserted = $wpdb->insert( $this->orders_table, $data, $format_map );

		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_order', __( '创建订单失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$order_id = (int) $wpdb->insert_id;

		// 保存订单项
		if ( ! $this->persist_order_items( $order_id, $items ) ) {
			$wpdb->delete( $this->orders_table, array( 'id' => $order_id ), array( '%d' ) );
			return $this->respond_error( 'failed_create_order', __( '保存订单明细失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		// 如果订单状态是已发货且有物流信息，尝试同步物流
		$row = $this->fetch_order_row( $order_id );
		if ( 'shipped' === $data['status'] && $tracking_provider && $tracking_number ) {
			$this->perform_tracking_sync( $order_id, $row );
			$row = $this->fetch_order_row( $order_id );
		}

		// 审计日志
		$this->log_audit(
			'create',
			'order',
			$order_id,
			array(
				'status'      => $row['status'],
				'items_count' => count( $items ),
			),
			$request
		);

		return $this->respond_success( $this->prepare_item_for_response( $row, 'detail' ), 201 );
	}

	/**
	 * 更新订单
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$order_id = (int) $request->get_param( 'id' );

		// 检查订单是否存在
		$row = $this->fetch_order_row( $order_id );
		if ( ! $row ) {
			return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
		}

		// 保存旧状态，用于判断是否需要奖励积分
		$old_status = $row['status'];

		$data  = array();
		$types = array();

		$requested_status = $row['status'];

		// 更新状态
		if ( $request->has_param( 'status' ) ) {
			$requested_status = sanitize_key( $request->get_param( 'status' ) );
			$allowed_statuses = array( 'pending', 'paid', 'shipped', 'completed', 'cancelled', 'refunded' );

			if ( ! in_array( $requested_status, $allowed_statuses, true ) ) {
				return $this->respond_error( 'invalid_order_status', __( '无效的订单状态。', 'tanzanite-settings' ) );
			}

			$data['status'] = $requested_status;
			$types[]        = '%s';
		}

		// 更新文本字段
		foreach ( array( 'payment_method', 'channel' ) as $field ) {
			if ( $request->has_param( $field ) ) {
				$data[ $field ] = sanitize_text_field( $request->get_param( $field ) );
				$types[]        = '%s';
			}
		}

		// 更新金额字段
		foreach ( array( 'total', 'subtotal', 'discount_total', 'shipping_total' ) as $field ) {
			if ( $request->has_param( $field ) ) {
				$data[ $field ] = (float) $request->get_param( $field );
				$types[]        = '%f';
			}
		}

		// 更新物流信息
		if ( $request->has_param( 'tracking_provider' ) ) {
			$data['tracking_provider'] = sanitize_key( $request->get_param( 'tracking_provider' ) );
			$types[]                   = '%s';
		}

		if ( $request->has_param( 'tracking_number' ) ) {
			$data['tracking_number'] = sanitize_text_field( $request->get_param( 'tracking_number' ) );
			$types[]                 = '%s';
		}

		// 应用状态时间戳
		$this->apply_status_timestamps( $data, $types, $row, $requested_status );

		// 更新订单数据
		if ( ! empty( $data ) ) {
			$update_result = $wpdb->update( $this->orders_table, $data, array( 'id' => $order_id ), $types, array( '%d' ) );
			if ( false === $update_result ) {
				return $this->respond_error( 'failed_update_order', __( '更新订单失败，请稍后重试。', 'tanzanite-settings' ), 500 );
			}
		}

		// 更新订单项
		if ( $request->has_param( 'items' ) ) {
			$items = $this->sanitize_order_items( $request->get_param( 'items' ) );
			if ( is_wp_error( $items ) ) {
				return $this->respond_error( $items->get_error_code(), $items->get_error_message(), 400 );
			}

			if ( ! $this->persist_order_items( $order_id, $items ) ) {
				return $this->respond_error( 'failed_update_order', __( '更新订单明细失败，请稍后重试。', 'tanzanite-settings' ), 500 );
			}
		}

		// 重新获取订单数据
		$row = $this->fetch_order_row( $order_id );

		// 如果订单状态变更为 completed，自动增加积分
		if ( $row['status'] === 'completed' && $old_status !== 'completed' ) {
			$this->award_order_points( $order_id, $row );
		}

		// 审计日志
		$this->log_audit(
			'update',
			'order',
			$order_id,
			array(
				'status' => $row['status'],
			),
			$request
		);

		return $this->respond_success( $this->prepare_item_for_response( $row, 'detail' ) );
	}

	/**
	 * 删除订单
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$order_id = (int) $request->get_param( 'id' );

		// 检查订单是否存在
		$row = $this->fetch_order_row( $order_id );
		if ( ! $row ) {
			return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
		}

		// 删除订单
		$deleted = $wpdb->delete( $this->orders_table, array( 'id' => $order_id ), array( '%d' ) );
		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_order', __( '删除订单失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		// 级联删除订单项
		$this->delete_order_items( $order_id );

		// 审计日志
		$this->log_audit(
			'delete',
			'order',
			$order_id,
			array(
				'status' => $row['status'],
			),
			$request
		);

		return $this->respond_success( array( 'deleted' => true ) );
	}

	/**
	 * 批量操作
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function bulk_action( $request ) {
		global $wpdb;

		$action = sanitize_key( $request->get_param( 'action' ) );

		// 支持的批量操作
		$allowed_actions = array( 'set_status', 'export' );

		if ( ! in_array( $action, $allowed_actions, true ) ) {
			return $this->respond_error( 'invalid_bulk_action', __( '当前批量操作类型不受支持。', 'tanzanite-settings' ) );
		}

		// 验证 ID 列表
		$ids = $request->get_param( 'ids' );
		if ( ! is_array( $ids ) || empty( $ids ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '请选择至少一个需要处理的订单。', 'tanzanite-settings' ) );
		}

		$ids = array_map( 'absint', $ids );
		$ids = array_filter( $ids );

		if ( empty( $ids ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '无效的订单ID列表。', 'tanzanite-settings' ) );
		}

		// 获取 payload
		$payload = $request->get_param( 'payload' );
		if ( ! is_array( $payload ) ) {
			$payload = array();
		}

		// 初始化结果
		$summary = array(
			'action'    => $action,
			'total'     => count( $ids ),
			'updated'   => 0,
			'failed'    => array(),
			'details'   => array(),
			'timestamp' => current_time( 'mysql' ),
		);

		// 根据操作类型处理
		switch ( $action ) {
			case 'set_status':
				$summary = $this->bulk_set_status( $ids, $payload, $summary, $request );
				break;

			case 'export':
				return $this->bulk_export_orders( $ids, $summary );
		}

		return $this->respond_success( $summary );
	}

	/**
	 * 同步物流信息
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function sync_tracking( $request ) {
		$order_id = (int) $request->get_param( 'id' );

		// 检查订单是否存在
		$row = $this->fetch_order_row( $order_id );
		if ( ! $row ) {
			return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
		}

		// 检查是否有物流信息
		if ( empty( $row['tracking_provider'] ) || empty( $row['tracking_number'] ) ) {
			return $this->respond_error( 'missing_tracking_info', __( '订单缺少物流信息。', 'tanzanite-settings' ), 400 );
		}

		// 执行物流同步
		$result = $this->perform_tracking_sync( $order_id, $row );

		if ( is_wp_error( $result ) ) {
			return $this->respond_error( $result->get_error_code(), $result->get_error_message() );
		}

		// 重新获取订单数据
		$row = $this->fetch_order_row( $order_id );

		$response = $this->prepare_item_for_response( $row, 'detail' );
		$response['tracking_sync'] = array(
			'synced_at' => $result['synced_at'] ?? null,
			'status'    => 'success',
		);

		return $this->respond_success( $response );
	}

	/**
	 * 准备订单数据用于响应
	 *
	 * @since 0.2.0
	 * @param array  $row 订单数据行
	 * @param string $context 上下文（summary 或 detail）
	 * @return array
	 */
	private function prepare_item_for_response( $row, $context = 'summary' ) {
		$response = array(
			'id'             => (int) $row['id'],
			'order_number'   => $row['order_number'],
			'status'         => $row['status'],
			'channel'        => $row['channel'],
			'payment_method' => $row['payment_method'],
			'amounts'        => array(
				'total'          => (float) $row['total'],
				'subtotal'       => (float) $row['subtotal'],
				'discount_total' => (float) $row['discount_total'],
				'shipping_total' => (float) $row['shipping_total'],
				'points_used'    => (int) $row['points_used'],
				'currency'       => $row['currency'],
			),
			'customer'       => $this->get_order_customer_summary( $row ),
			'timestamps'     => array_filter(
				array(
					'created_at'   => $row['created_at'],
					'updated_at'   => $row['updated_at'],
					'paid_at'      => $row['paid_at'] ?? null,
					'shipped_at'   => $row['shipped_at'] ?? null,
					'completed_at' => $row['completed_at'] ?? null,
					'cancelled_at' => $row['cancelled_at'] ?? null,
				),
				function ( $value ) {
					return ! empty( $value );
				}
			),
			'tracking'       => $this->get_order_tracking_payload( $row ),
			'tracking_meta'  => $this->get_order_tracking_meta( $row ),
		);

		// 详情模式包含更多信息
		if ( 'detail' === $context ) {
			$response['items']           = $this->get_order_items_payload( (int) $row['id'] );
			$response['customer_detail'] = $this->get_order_customer_detail( $row );
		}

		return $response;
	}

	/**
	 * 获取订单客户摘要
	 *
	 * @since 0.2.0
	 * @param array $row 订单数据行
	 * @return array
	 */
	private function get_order_customer_summary( $row ) {
		$user_id = isset( $row['user_id'] ) ? (int) $row['user_id'] : 0;

		if ( $user_id <= 0 ) {
			return array();
		}

		$user = get_userdata( $user_id );
		if ( ! $user ) {
			return array();
		}

		return array(
			'id'           => $user_id,
			'display_name' => $user->display_name,
			'email'        => $user->user_email,
		);
	}

	/**
	 * 获取订单客户详情
	 *
	 * @since 0.2.0
	 * @param array $row 订单数据行
	 * @return array
	 */
	private function get_order_customer_detail( $row ) {
		$user_id = isset( $row['user_id'] ) ? (int) $row['user_id'] : 0;

		if ( $user_id <= 0 ) {
			return array();
		}

		$user = get_userdata( $user_id );
		if ( ! $user ) {
			return array();
		}

		return array(
			'id'           => $user_id,
			'display_name' => $user->display_name,
			'email'        => $user->user_email,
			'login'        => $user->user_login,
		);
	}

	/**
	 * 获取订单物流信息
	 *
	 * @since 0.2.0
	 * @param array $row 订单数据行
	 * @return array
	 */
	private function get_order_tracking_payload( $row ) {
		$provider = $row['tracking_provider'] ?? null;
		$number   = $row['tracking_number'] ?? null;
		$synced   = $row['tracking_synced_at'] ?? null;

		return array(
			'provider'   => $provider,
			'number'     => $number,
			'synced_at'  => $synced,
		);
	}

	/**
	 * 获取订单物流元数据
	 *
	 * @since 0.2.0
	 * @param array $row 订单数据行
	 * @return array
	 */
	private function get_order_tracking_meta( $row ) {
		return array_filter(
			array(
				'synced_at' => $row['tracking_synced_at'] ?? null,
			),
			function ( $value ) {
				return null !== $value && '' !== $value;
			}
		);
	}

	/**
	 * 获取订单项列表
	 *
	 * @since 0.2.0
	 * @param int $order_id 订单ID
	 * @return array
	 */
	private function get_order_items_payload( $order_id ) {
		global $wpdb;

		$items = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->order_items_table} WHERE order_id = %d ORDER BY id ASC",
				$order_id
			),
			ARRAY_A
		);

		if ( ! $items ) {
			return array();
		}

		return array_map(
			function ( $item ) {
				return array(
					'product_id'    => (int) $item['product_id'],
					'sku_id'        => (int) $item['sku_id'],
					'product_title' => $item['product_title'],
					'sku_code'      => $item['sku_code'],
					'quantity'      => (int) $item['quantity'],
					'price'         => (float) $item['price'],
					'total'         => (float) $item['total'],
					'meta'          => $item['meta'],
				);
			},
			$items
		);
	}

	/**
	 * 从数据库获取订单数据行
	 *
	 * @since 0.2.0
	 * @param int $order_id 订单ID
	 * @return array|null
	 */
	private function fetch_order_row( $order_id ) {
		global $wpdb;

		$row = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$this->orders_table} WHERE id = %d",
				$order_id
			),
			ARRAY_A
		);

		return $row ? $row : null;
	}

	/**
	 * 删除订单的所有订单项
	 *
	 * @since 0.2.0
	 * @param int $order_id 订单ID
	 * @return void
	 */
	private function delete_order_items( $order_id ) {
		global $wpdb;

		$wpdb->delete(
			$this->order_items_table,
			array( 'order_id' => $order_id ),
			array( '%d' )
		);
	}

	/**
	 * 应用状态时间戳
	 *
	 * @since 0.2.0
	 * @param array  &$data 数据数组（引用）
	 * @param array  &$types 类型数组（引用）
	 * @param array  $row 当前订单数据
	 * @param string $new_status 新状态
	 * @return void
	 */
	private function apply_status_timestamps( &$data, &$types, $row, $new_status ) {
		$old_status = $row['status'];

		// 如果状态没有变化，不更新时间戳
		if ( $old_status === $new_status ) {
			return;
		}

		$now = current_time( 'mysql', true );

		// 根据新状态设置相应的时间戳
		switch ( $new_status ) {
			case 'paid':
				if ( empty( $row['paid_at'] ) ) {
					$data['paid_at'] = $now;
					$types[]         = '%s';
				}
				break;

			case 'shipped':
				if ( empty( $row['shipped_at'] ) ) {
					$data['shipped_at'] = $now;
					$types[]            = '%s';
				}
				break;

			case 'completed':
				if ( empty( $row['completed_at'] ) ) {
					$data['completed_at'] = $now;
					$types[]              = '%s';
				}
				break;

			case 'cancelled':
				if ( empty( $row['cancelled_at'] ) ) {
					$data['cancelled_at'] = $now;
					$types[]              = '%s';
				}
				break;
		}
	}

	/**
	 * 清理和验证订单项
	 *
	 * @since 0.2.0
	 * @param mixed $items 订单项数据
	 * @return array|WP_Error
	 */
	private function sanitize_order_items( $items ) {
		if ( ! is_array( $items ) || empty( $items ) ) {
			return new WP_Error( 'invalid_order_items', __( '订单项必须是非空数组。', 'tanzanite-settings' ) );
		}

		$sanitized = array();

		foreach ( $items as $index => $item ) {
			if ( ! is_array( $item ) ) {
				return new WP_Error( 'invalid_order_item', sprintf( __( '订单项 %d 格式无效。', 'tanzanite-settings' ), $index + 1 ) );
			}

			// 必填字段验证
			if ( empty( $item['product_id'] ) ) {
				return new WP_Error( 'missing_product_id', sprintf( __( '订单项 %d 缺少 product_id。', 'tanzanite-settings' ), $index + 1 ) );
			}

			if ( empty( $item['quantity'] ) || $item['quantity'] <= 0 ) {
				return new WP_Error( 'invalid_quantity', sprintf( __( '订单项 %d 的数量无效。', 'tanzanite-settings' ), $index + 1 ) );
			}

			$sanitized[] = array(
				'product_id'    => (int) $item['product_id'],
				'sku_id'        => isset( $item['sku_id'] ) ? (int) $item['sku_id'] : 0,
				'product_title' => isset( $item['product_title'] ) ? sanitize_text_field( $item['product_title'] ) : '',
				'sku_code'      => isset( $item['sku_code'] ) ? sanitize_text_field( $item['sku_code'] ) : '',
				'quantity'      => (int) $item['quantity'],
				'price'         => isset( $item['price'] ) ? (float) $item['price'] : 0.0,
				'total'         => isset( $item['total'] ) ? (float) $item['total'] : 0.0,
				'meta'          => isset( $item['meta'] ) ? wp_json_encode( $item['meta'] ) : '{}',
			);
		}

		return $sanitized;
	}

	/**
	 * 保存订单项到数据库
	 *
	 * @since 0.2.0
	 * @param int   $order_id 订单ID
	 * @param array $items 订单项数组
	 * @return bool
	 */
	private function persist_order_items( $order_id, $items ) {
		global $wpdb;

		// 先删除旧的订单项
		$this->delete_order_items( $order_id );

		// 插入新的订单项
		foreach ( $items as $item ) {
			$result = $wpdb->insert(
				$this->order_items_table,
				array(
					'order_id'      => $order_id,
					'product_id'    => $item['product_id'],
					'sku_id'        => $item['sku_id'],
					'product_title' => $item['product_title'],
					'sku_code'      => $item['sku_code'],
					'quantity'      => $item['quantity'],
					'price'         => $item['price'],
					'total'         => $item['total'],
					'meta'          => $item['meta'],
				),
				array( '%d', '%d', '%d', '%s', '%s', '%d', '%f', '%f', '%s' )
			);

			if ( false === $result ) {
				return false;
			}
		}

		return true;
	}

	/**
	 * 批量设置订单状态
	 *
	 * @since 0.2.0
	 */
	private function bulk_set_status( $ids, $payload, $summary, $request ) {
		global $wpdb;

		if ( empty( $payload['status'] ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '批量修改状态需要指定目标状态。', 'tanzanite-settings' ) );
		}

		$target_status = sanitize_key( (string) $payload['status'] );
		$allowed_statuses = array( 'pending', 'paid', 'shipped', 'completed', 'cancelled', 'refunded' );

		if ( ! in_array( $target_status, $allowed_statuses, true ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '目标状态无效。', 'tanzanite-settings' ) );
		}

		foreach ( $ids as $order_id ) {
			$row = $this->fetch_order_row( $order_id );

			if ( ! $row ) {
				$summary['failed'][] = array(
					'id'     => $order_id,
					'reason' => __( '订单不存在。', 'tanzanite-settings' ),
				);
				continue;
			}

			$current_status = $row['status'];

			// 如果状态相同，跳过
			if ( $current_status === $target_status ) {
				$summary['updated']++;
				$summary['details'][] = array(
					'id'      => $order_id,
					'status'  => $target_status,
					'changed' => false,
				);
				continue;
			}

			// 更新状态
			$data  = array( 'status' => $target_status );
			$types = array( '%s' );

			$this->apply_status_timestamps( $data, $types, $row, $target_status );

			$updated = $wpdb->update( $this->orders_table, $data, array( 'id' => $order_id ), $types, array( '%d' ) );

			if ( false === $updated ) {
				$summary['failed'][] = array(
					'id'     => $order_id,
					'reason' => __( '数据库更新失败。', 'tanzanite-settings' ),
				);
				continue;
			}

			$summary['updated']++;
			$summary['details'][] = array(
				'id'      => $order_id,
				'from'    => $current_status,
				'status'  => $target_status,
				'changed' => true,
			);

			$this->log_audit(
				'bulk_set_status',
				'order',
				$order_id,
				array(
					'action' => 'set_status',
					'from'   => $current_status,
					'to'     => $target_status,
				),
				$request
			);
		}

		return $summary;
	}

	/**
	 * 批量导出订单
	 *
	 * @since 0.2.0
	 */
	private function bulk_export_orders( $ids, $summary ) {
		$export_rows = array();

		foreach ( $ids as $order_id ) {
			$row = $this->fetch_order_row( $order_id );

			if ( ! $row ) {
				$summary['failed'][] = array(
					'id'     => $order_id,
					'reason' => __( '订单不存在。', 'tanzanite-settings' ),
				);
				continue;
			}

			$export_rows[] = $this->prepare_item_for_response( $row );
		}

		return $this->respond_success(
			array(
				'action'    => 'export',
				'total'     => $summary['total'],
				'exported'  => count( $export_rows ),
				'failed'    => $summary['failed'],
				'timestamp' => $summary['timestamp'],
				'items'     => $export_rows,
			)
		);
	}

	/**
	 * 执行物流同步
	 *
	 * @since 0.2.0
	 * @param int   $order_id 订单ID
	 * @param array $row 订单数据
	 * @return array|WP_Error
	 */
	private function perform_tracking_sync( $order_id, $row ) {
		global $wpdb;

		$provider = $row['tracking_provider'];
		$number   = $row['tracking_number'];

		// 这里可以集成第三方物流API
		// 例如：调用快递100、快递鸟等API获取物流信息
		// 目前实现简化版本，仅更新同步时间

		/**
		 * 允许第三方插件hook物流同步
		 *
		 * @param int    $order_id 订单ID
		 * @param string $provider 物流商
		 * @param string $number   物流单号
		 */
		$tracking_data = apply_filters( 'tanzanite_sync_order_tracking', null, $order_id, $provider, $number );

		// 如果有第三方实现，使用第三方返回的数据
		if ( is_wp_error( $tracking_data ) ) {
			return $tracking_data;
		}

		// 更新同步时间
		$synced_at = current_time( 'mysql', true );
		$updated   = $wpdb->update(
			$this->orders_table,
			array( 'tracking_synced_at' => $synced_at ),
			array( 'id' => $order_id ),
			array( '%s' ),
			array( '%d' )
		);

		if ( false === $updated ) {
			return new WP_Error( 'sync_failed', __( '物流同步失败，请稍后重试。', 'tanzanite-settings' ) );
		}

		// 如果有第三方返回的物流事件，可以保存到物流事件表
		// 这里简化处理，仅返回同步时间
		return array(
			'synced_at'      => $synced_at,
			'provider'       => $provider,
			'tracking_number' => $number,
		);
	}

	/**
	 * 生成订单编号
	 *
	 * @since 0.2.0
	 * @return string
	 */
	private function generate_order_number() {
		// 生成格式：XXXX-TZXXXXXXXX
		// 例如：AB12-TZ20250110001
		return strtoupper( wp_generate_password( 4, false, false ) ) . '-' . wp_unique_id( 'TZ' );
	}

	/**
	 * 订单完成后自动增加积分
	 *
	 * @since 0.2.0
	 * @param int   $order_id 订单 ID
	 * @param array $order    订单数据
	 * @return void
	 */
	private function award_order_points( $order_id, $order ) {
		global $wpdb;

		$user_id = isset( $order['user_id'] ) ? absint( $order['user_id'] ) : 0;
		if ( $user_id <= 0 ) {
			return; // 游客订单不奖励积分
		}

		// 1. 获取积分配置
		$config   = get_option( 'tanzanite_loyalty_config', '' );
		$settings = json_decode( $config, true );
		if ( ! is_array( $settings ) ) {
			$settings = array();
		}

		// 检查积分系统是否启用
		if ( empty( $settings['enabled'] ) ) {
			return;
		}

		$points_per_unit = isset( $settings['points_per_unit'] ) ? absint( $settings['points_per_unit'] ) : 1;

		if ( $points_per_unit <= 0 ) {
			return; // 未配置积分比例
		}

		// 2. 计算积分（每消费1元获得对应积分）
		$total  = isset( $order['total'] ) ? floatval( $order['total'] ) : 0;
		$points = floor( $total * $points_per_unit );

		if ( $points <= 0 ) {
			return;
		}

		// 3. 检查是否已经奖励过积分（防止重复奖励）
		$rewards_table = $wpdb->prefix . 'tanz_rewards_transactions';
		$existing      = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT id FROM {$rewards_table} 
				WHERE user_id = %d 
				AND related_type = 'order' 
				AND related_id = %d 
				AND action = 'earn'",
				$user_id,
				$order_id
			)
		);

		if ( $existing ) {
			return; // 已经奖励过
		}

		// 4. 增加用户积分
		$current_points = absint( get_user_meta( $user_id, 'loyalty_points', true ) );
		$new_points     = $current_points + $points;
		update_user_meta( $user_id, 'loyalty_points', $new_points );

		// 5. 记录积分交易
		$wpdb->insert(
			$rewards_table,
			array(
				'user_id'      => $user_id,
				'related_type' => 'order',
				'related_id'   => $order_id,
				'action'       => 'earn',
				'points_delta' => $points,
				'amount_delta' => $total,
				'notes'        => sprintf(
					/* translators: %d: order ID */
					__( '订单完成获得积分，订单号：%d', 'tanzanite-settings' ),
					$order_id
				),
				'created_at'   => current_time( 'mysql' ),
			)
		);

		// 6. 记录日志（调试用）
		if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
			error_log(
				sprintf(
					'[Tanzanite] Order #%d completed, awarded %d points to user #%d (total: %d)',
					$order_id,
					$points,
					$user_id,
					$new_points
				)
			);
		}
	}
}
