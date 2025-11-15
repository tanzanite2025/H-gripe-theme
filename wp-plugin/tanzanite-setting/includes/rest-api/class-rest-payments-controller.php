<?php
/**
 * Payment Methods REST API Controller
 *
 * 处理支付方式相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 支付方式 REST API 控制器
 *
 * 提供支付方式的 CRUD 操作
 */
class Tanzanite_REST_Payments_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'payment-methods';

	/**
	 * 支付方式表名
	 *
	 * @var string
	 */
	private $payment_methods_table;

	/**
	 * 支付终端常量
	 *
	 * @var array
	 */
	private $payment_terminals = array( 'web', 'mobile', 'app', 'pos', 'wechat', 'alipay' );

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->payment_methods_table = $wpdb->prefix . 'tanz_payment_methods';
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
					'permission_callback' => 'is_user_logged_in',
					'args'                => $this->get_collection_params(),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_payments', true ),
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个支付方式
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
					'permission_callback' => $this->permission_callback( 'tanz_manage_payments', true ),
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_payments', true ),
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
	 * 获取支付方式列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		$pagination = $this->get_pagination_params( $request );

		$total = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$this->payment_methods_table}" );
		$rows  = $wpdb->get_results( $wpdb->prepare( "SELECT * FROM {$this->payment_methods_table} ORDER BY sort_order ASC, id ASC LIMIT %d OFFSET %d", $pagination['per_page'], $pagination['offset'] ), ARRAY_A );

		$items = array_map( array( $this, 'format_payment_method_row' ), $rows );

		return $this->respond_success(
			array(
				'items' => $items,
				'meta'  => $this->build_pagination_meta( $total, $pagination['page'], $pagination['per_page'] ),
			)
		);
	}

	/**
	 * 获取单个支付方式
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		$row = $this->fetch_payment_method_row( (int) $request['id'] );

		if ( ! $row ) {
			return $this->respond_error( 'payment_method_not_found', __( '指定的支付方式不存在。', 'tanzanite-settings' ), 404 );
		}

		return $this->respond_success( $this->format_payment_method_row( $row ) );
	}

	/**
	 * 创建支付方式
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$sanitized = $this->sanitize_payment_method_request( $request, false );
		if ( is_wp_error( $sanitized ) ) {
			return $this->respond_error( $sanitized->get_error_code(), $sanitized->get_error_message(), 400 );
		}

		$data   = $sanitized['data'];
		$format = $sanitized['format'];

		$inserted = $wpdb->insert( $this->payment_methods_table, $data, $format );
		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_payment_method', __( '创建支付方式失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$id   = (int) $wpdb->insert_id;
		$row  = $this->fetch_payment_method_row( $id );
		$item = $this->format_payment_method_row( $row );

		$this->log_audit( 'create', 'payment_method', $id, array( 'code' => $item['code'] ), $request );

		return $this->respond_success( $item, 201 );
	}

	/**
	 * 更新支付方式
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$id  = (int) $request['id'];
		$row = $this->fetch_payment_method_row( $id );
		if ( ! $row ) {
			return $this->respond_error( 'payment_method_not_found', __( '指定的支付方式不存在。', 'tanzanite-settings' ), 404 );
		}

		$sanitized = $this->sanitize_payment_method_request( $request, true, $row );
		if ( is_wp_error( $sanitized ) ) {
			return $this->respond_error( $sanitized->get_error_code(), $sanitized->get_error_message(), 400 );
		}

		$data   = $sanitized['data'];
		$format = $sanitized['format'];

		$updated = $wpdb->update( $this->payment_methods_table, $data, array( 'id' => $id ), $format, array( '%d' ) );
		if ( false === $updated ) {
			return $this->respond_error( 'failed_update_payment_method', __( '更新支付方式失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$updated_row = $this->fetch_payment_method_row( $id );
		$item        = $this->format_payment_method_row( $updated_row );

		$this->log_audit( 'update', 'payment_method', $id, array( 'code' => $item['code'] ), $request );

		return $this->respond_success( $item );
	}

	/**
	 * 删除支付方式
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$id  = (int) $request['id'];
		$row = $this->fetch_payment_method_row( $id );

		if ( ! $row ) {
			return $this->respond_error( 'payment_method_not_found', __( '指定的支付方式不存在。', 'tanzanite-settings' ), 404 );
		}

		$deleted = $wpdb->delete( $this->payment_methods_table, array( 'id' => $id ), array( '%d' ) );
		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_payment_method', __( '删除支付方式失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'payment_method', $id, array( 'code' => $row['code'] ), $request );

		return $this->respond_success( array( 'deleted' => true ) );
	}

	/**
	 * 获取单个支付方式（内部方法）
	 *
	 * @since 0.2.0
	 * @param int $id 支付方式 ID
	 * @return array|null
	 */
	private function fetch_payment_method_row( $id ) {
		global $wpdb;

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->payment_methods_table} WHERE id = %d", $id ), ARRAY_A );

		return $row ?: null;
	}

	/**
	 * 格式化支付方式数据
	 *
	 * @since 0.2.0
	 * @param array $row 数据库行
	 * @return array
	 */
	private function format_payment_method_row( $row ) {
		$terminals = $row['terminals'] ? json_decode( $row['terminals'], true ) : array();
		if ( ! is_array( $terminals ) ) {
			$terminals = array();
		}
		$terminals = array_values( array_intersect( array_map( 'sanitize_key', $terminals ), $this->payment_terminals ) );

		$membership_levels = $row['membership_levels'] ? json_decode( $row['membership_levels'], true ) : array();
		if ( ! is_array( $membership_levels ) ) {
			$membership_levels = array();
		}

		$currencies = $row['currencies'] ? json_decode( $row['currencies'], true ) : array();
		if ( ! is_array( $currencies ) ) {
			$currencies = array();
		}

		$settings = $row['settings'] ? json_decode( $row['settings'], true ) : array();
		if ( ! is_array( $settings ) ) {
			$settings = array();
		}

		$meta = $row['meta'] ? json_decode( $row['meta'], true ) : array();
		if ( ! is_array( $meta ) ) {
			$meta = array();
		}

		return array(
			'id'                => (int) $row['id'],
			'code'              => $row['code'],
			'name'              => $row['name'],
			'description'       => $row['description'],
			'icon_url'          => $row['icon_url'] ? esc_url_raw( $row['icon_url'] ) : '',
			'fee_type'          => $row['fee_type'],
			'fee_value'         => (float) $row['fee_value'],
			'terminals'         => $terminals,
			'membership_levels' => $membership_levels,
			'currencies'        => $currencies,
			'default_currency'  => sanitize_text_field( (string) $row['default_currency'] ),
			'settings'          => $settings,
			'is_enabled'        => (bool) $row['is_enabled'],
			'sort_order'        => (int) $row['sort_order'],
			'meta'              => $meta,
			'created_at'        => $row['created_at'],
			'updated_at'        => $row['updated_at'],
		);
	}

	/**
	 * 验证和清理支付方式请求数据
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request     REST 请求对象
	 * @param bool            $is_update   是否为更新操作
	 * @param array           $existing    现有数据
	 * @return array|WP_Error
	 */
	private function sanitize_payment_method_request( $request, $is_update = false, $existing = array() ) {
		$data   = array();
		$format = array();

		// 验证 code
		if ( ! $is_update || $request->has_param( 'code' ) ) {
			$code = sanitize_key( (string) $request->get_param( 'code' ) );
			if ( '' === $code ) {
				return new WP_Error( 'invalid_payment_payload', __( '支付方式编码不能为空。', 'tanzanite-settings' ) );
			}

			$data['code'] = $code;
			$format[]     = '%s';
		}

		// 验证 name
		if ( ! $is_update || $request->has_param( 'name' ) ) {
			$name = sanitize_text_field( (string) $request->get_param( 'name' ) );
			if ( '' === $name ) {
				return new WP_Error( 'invalid_payment_payload', __( '支付方式名称不能为空。', 'tanzanite-settings' ) );
			}

			$data['name'] = $name;
			$format[]     = '%s';
		}

		// 其他字段
		if ( $request->has_param( 'description' ) ) {
			$data['description'] = sanitize_textarea_field( (string) $request->get_param( 'description' ) );
			$format[]            = '%s';
		}

		if ( $request->has_param( 'icon_url' ) ) {
			$icon_url = trim( (string) $request->get_param( 'icon_url' ) );
			if ( '' !== $icon_url ) {
				$icon_url = esc_url_raw( $icon_url );
			}
			$data['icon_url'] = $icon_url;
			$format[]         = '%s';
		}

		if ( ! $is_update || $request->has_param( 'fee_type' ) ) {
			$fee_type = sanitize_key( (string) $request->get_param( 'fee_type' ) );
			if ( ! in_array( $fee_type, array( 'fixed', 'percentage' ), true ) ) {
				$fee_type = 'fixed';
			}
			$data['fee_type'] = $fee_type;
			$format[]         = '%s';
		}

		if ( ! $is_update || $request->has_param( 'fee_value' ) ) {
			$data['fee_value'] = max( 0, (float) $request->get_param( 'fee_value' ) );
			$format[]          = '%f';
		}

		if ( $request->has_param( 'terminals' ) ) {
			$terminals        = (array) $request->get_param( 'terminals' );
			$data['terminals'] = wp_json_encode( array_values( array_intersect( $terminals, $this->payment_terminals ) ) );
			$format[]         = '%s';
		}

		if ( $request->has_param( 'membership_levels' ) ) {
			$levels                    = (array) $request->get_param( 'membership_levels' );
			$data['membership_levels'] = wp_json_encode( $levels );
			$format[]                  = '%s';
		}

		if ( $request->has_param( 'currencies' ) ) {
			$currencies         = (array) $request->get_param( 'currencies' );
			$data['currencies'] = wp_json_encode( $currencies );
			$format[]           = '%s';
		}

		if ( $request->has_param( 'default_currency' ) ) {
			$data['default_currency'] = strtoupper( sanitize_text_field( (string) $request->get_param( 'default_currency' ) ) );
			$format[]                 = '%s';
		}

		if ( $request->has_param( 'settings' ) ) {
			$settings         = $request->get_param( 'settings' );
			$data['settings'] = is_array( $settings ) ? wp_json_encode( $settings ) : '{}';
			$format[]         = '%s';
		}

		if ( $request->has_param( 'meta' ) ) {
			$meta         = $request->get_param( 'meta' );
			$data['meta'] = is_array( $meta ) ? wp_json_encode( $meta ) : '{}';
			$format[]     = '%s';
		}

		if ( ! $is_update || $request->has_param( 'is_enabled' ) ) {
			$data['is_enabled'] = (bool) $request->get_param( 'is_enabled' );
			$format[]           = '%d';
		}

		if ( ! $is_update || $request->has_param( 'sort_order' ) ) {
			$data['sort_order'] = (int) $request->get_param( 'sort_order' );
			$format[]           = '%d';
		}

		return array(
			'data'   => $data,
			'format' => $format,
		);
	}

	/**
	 * 获取集合参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_collection_params() {
		return array(
			'page'     => array(
				'type'    => 'integer',
				'default' => 1,
			),
			'per_page' => array(
				'type'    => 'integer',
				'default' => 50,
			),
		);
	}

	/**
	 * 获取创建参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_create_params() {
		return array(
			'code'              => array(
				'type'     => 'string',
				'required' => true,
			),
			'name'              => array(
				'type'     => 'string',
				'required' => true,
			),
			'description'       => array(
				'type' => 'string',
			),
			'fee_type'          => array(
				'type'    => 'string',
				'default' => 'fixed',
			),
			'fee_value'         => array(
				'type'    => 'number',
				'default' => 0,
			),
			'terminals'         => array(
				'type' => 'array',
			),
			'membership_levels' => array(
				'type' => 'array',
			),
			'currencies'        => array(
				'type' => 'array',
			),
			'default_currency'  => array(
				'type' => 'string',
			),
			'settings'          => array(
				'type' => 'object',
			),
			'meta'              => array(
				'type' => 'object',
			),
			'is_enabled'        => array(
				'type'    => 'boolean',
				'default' => true,
			),
			'sort_order'        => array(
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
			'code'              => array(
				'type' => 'string',
			),
			'name'              => array(
				'type' => 'string',
			),
			'description'       => array(
				'type' => 'string',
			),
			'icon_url'          => array(
				'type' => 'string',
			),
			'fee_type'          => array(
				'type' => 'string',
			),
			'fee_value'         => array(
				'type' => 'number',
			),
			'terminals'         => array(
				'type' => 'array',
			),
			'membership_levels' => array(
				'type' => 'array',
			),
			'currencies'        => array(
				'type' => 'array',
			),
			'default_currency'  => array(
				'type' => 'string',
			),
			'settings'          => array(
				'type' => 'object',
			),
			'meta'              => array(
				'type' => 'object',
			),
			'is_enabled'        => array(
				'type' => 'boolean',
			),
			'sort_order'        => array(
				'type' => 'integer',
			),
		);
	}
}
