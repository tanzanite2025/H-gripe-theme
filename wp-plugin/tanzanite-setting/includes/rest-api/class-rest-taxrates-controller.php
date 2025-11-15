<?php
/**
 * Tax Rates REST API Controller
 *
 * 处理税率相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 税率 REST API 控制器
 *
 * 提供税率的 CRUD 操作
 */
class Tanzanite_REST_TaxRates_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'tax-rates';

	/**
	 * 税率表名
	 *
	 * @var string
	 */
	private $tax_rates_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->tax_rates_table = $wpdb->prefix . 'tanz_tax_rates';
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
					'args'                => $this->get_collection_params(),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个税率
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
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
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
	 * 获取税率列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		try {
			$pagination = $this->get_pagination_params( $request );

			$total = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$this->tax_rates_table}" );

			if ( $wpdb->last_error ) {
				if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
					error_log( 'Tax rates count error: ' . $wpdb->last_error );
				}
				return new WP_REST_Response( array( 'error' => 'Database error: ' . $wpdb->last_error ), 500 );
			}

			$rows = $wpdb->get_results( $wpdb->prepare( "SELECT * FROM {$this->tax_rates_table} ORDER BY sort_order ASC, id ASC LIMIT %d OFFSET %d", $pagination['per_page'], $pagination['offset'] ), ARRAY_A );

			if ( $wpdb->last_error ) {
				if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
					error_log( 'Tax rates query error: ' . $wpdb->last_error );
				}
				return new WP_REST_Response( array( 'error' => 'Database error: ' . $wpdb->last_error ), 500 );
			}

			$items = array();
			if ( is_array( $rows ) && ! empty( $rows ) ) {
				$items = array_map( array( $this, 'format_tax_rate_row' ), $rows );
			}

			return $this->respond_success(
				array(
					'items' => $items,
					'meta'  => $this->build_pagination_meta( $total, $pagination['page'], $pagination['per_page'] ),
				)
			);
		} catch ( Exception $e ) {
			if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
				error_log( 'Tax rates API exception: ' . $e->getMessage() );
			}
			return new WP_REST_Response( array( 'error' => $e->getMessage() ), 500 );
		}
	}

	/**
	 * 获取单个税率
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		$row = $this->fetch_tax_rate_row( (int) $request['id'] );

		if ( ! $row ) {
			return $this->respond_error( 'tax_rate_not_found', __( '指定的税率不存在。', 'tanzanite-settings' ), 404 );
		}

		return $this->respond_success( $this->format_tax_rate_row( $row ) );
	}

	/**
	 * 创建税率
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$data = array(
			'name'        => sanitize_text_field( $request->get_param( 'name' ) ),
			'rate'        => (float) $request->get_param( 'rate' ),
			'region'      => sanitize_text_field( $request->get_param( 'region' ) ?: '' ),
			'description' => sanitize_textarea_field( $request->get_param( 'description' ) ?: '' ),
			'is_active'   => (bool) $request->get_param( 'is_active' ),
			'sort_order'  => (int) $request->get_param( 'sort_order' ),
		);

		$meta = $request->get_param( 'meta' );
		if ( is_array( $meta ) ) {
			$data['meta'] = wp_json_encode( $meta );
		} else {
			$data['meta'] = wp_json_encode( array() );
		}

		$format = array( '%s', '%f', '%s', '%s', '%d', '%d', '%s' );

		$inserted = $wpdb->insert( $this->tax_rates_table, $data, $format );
		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_tax_rate', __( '创建税率失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$id   = (int) $wpdb->insert_id;
		$row  = $this->fetch_tax_rate_row( $id );
		$item = $this->format_tax_rate_row( $row );

		$this->log_audit( 'create', 'tax_rate', $id, array( 'name' => $item['name'] ), $request );

		return $this->respond_success( $item, 201 );
	}

	/**
	 * 更新税率
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$id  = (int) $request['id'];
		$row = $this->fetch_tax_rate_row( $id );
		if ( ! $row ) {
			return $this->respond_error( 'tax_rate_not_found', __( '指定的税率不存在。', 'tanzanite-settings' ), 404 );
		}

		$data   = array();
		$format = array();

		if ( $request->has_param( 'name' ) ) {
			$data['name'] = sanitize_text_field( $request->get_param( 'name' ) );
			$format[]     = '%s';
		}

		if ( $request->has_param( 'rate' ) ) {
			$data['rate'] = (float) $request->get_param( 'rate' );
			$format[]     = '%f';
		}

		if ( $request->has_param( 'region' ) ) {
			$data['region'] = sanitize_text_field( $request->get_param( 'region' ) );
			$format[]       = '%s';
		}

		if ( $request->has_param( 'description' ) ) {
			$data['description'] = sanitize_textarea_field( $request->get_param( 'description' ) );
			$format[]            = '%s';
		}

		if ( $request->has_param( 'is_active' ) ) {
			$data['is_active'] = (bool) $request->get_param( 'is_active' );
			$format[]          = '%d';
		}

		if ( $request->has_param( 'sort_order' ) ) {
			$data['sort_order'] = (int) $request->get_param( 'sort_order' );
			$format[]           = '%d';
		}

		if ( $request->has_param( 'meta' ) ) {
			$meta = $request->get_param( 'meta' );
			if ( is_array( $meta ) ) {
				$data['meta'] = wp_json_encode( $meta );
			} else {
				$data['meta'] = wp_json_encode( array() );
			}
			$format[] = '%s';
		}

		if ( empty( $data ) ) {
			return $this->respond_error( 'invalid_tax_rate_payload', __( '没有可更新的字段。', 'tanzanite-settings' ) );
		}

		$updated = $wpdb->update( $this->tax_rates_table, $data, array( 'id' => $id ), $format, array( '%d' ) );
		if ( false === $updated ) {
			return $this->respond_error( 'failed_update_tax_rate', __( '更新税率失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$updated_row = $this->fetch_tax_rate_row( $id );
		$item        = $this->format_tax_rate_row( $updated_row );

		$this->log_audit( 'update', 'tax_rate', $id, array( 'name' => $item['name'] ), $request );

		return $this->respond_success( $item );
	}

	/**
	 * 删除税率
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$id  = (int) $request['id'];
		$row = $this->fetch_tax_rate_row( $id );

		if ( ! $row ) {
			return $this->respond_error( 'tax_rate_not_found', __( '指定的税率不存在。', 'tanzanite-settings' ), 404 );
		}

		$deleted = $wpdb->delete( $this->tax_rates_table, array( 'id' => $id ), array( '%d' ) );
		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_tax_rate', __( '删除税率失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'tax_rate', $id, array( 'name' => $row['name'] ), $request );

		return $this->respond_success( array( 'deleted' => true ) );
	}

	/**
	 * 获取单个税率（内部方法）
	 *
	 * @since 0.2.0
	 * @param int $id 税率 ID
	 * @return array|null
	 */
	private function fetch_tax_rate_row( $id ) {
		global $wpdb;

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->tax_rates_table} WHERE id = %d", $id ), ARRAY_A );

		return $row ?: null;
	}

	/**
	 * 格式化税率数据
	 *
	 * @since 0.2.0
	 * @param array $row 数据库行
	 * @return array
	 */
	private function format_tax_rate_row( $row ) {
		$meta = $row['meta'] ? json_decode( $row['meta'], true ) : array();
		if ( ! is_array( $meta ) ) {
			$meta = array();
		}

		return array(
			'id'          => (int) $row['id'],
			'name'        => $row['name'],
			'rate'        => (float) $row['rate'],
			'region'      => $row['region'],
			'description' => $row['description'],
			'is_active'   => (bool) $row['is_active'],
			'sort_order'  => (int) $row['sort_order'],
			'meta'        => $meta,
			'created_at'  => $row['created_at'],
			'updated_at'  => $row['updated_at'],
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
			'name'        => array(
				'type'     => 'string',
				'required' => true,
			),
			'rate'        => array(
				'type'     => 'number',
				'required' => true,
			),
			'region'      => array(
				'type'    => 'string',
				'default' => '',
			),
			'description' => array(
				'type' => 'string',
			),
			'is_active'   => array(
				'type'    => 'boolean',
				'default' => true,
			),
			'sort_order'  => array(
				'type'    => 'integer',
				'default' => 0,
			),
			'meta'        => array(
				'type' => 'object',
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
			'name'        => array(
				'type' => 'string',
			),
			'rate'        => array(
				'type' => 'number',
			),
			'region'      => array(
				'type' => 'string',
			),
			'description' => array(
				'type' => 'string',
			),
			'is_active'   => array(
				'type' => 'boolean',
			),
			'sort_order'  => array(
				'type' => 'integer',
			),
			'meta'        => array(
				'type' => 'object',
			),
		);
	}
}
