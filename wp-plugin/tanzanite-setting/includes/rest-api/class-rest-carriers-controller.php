<?php
/**
 * Carriers REST API Controller
 *
 * 处理物流公司相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 物流公司 REST API 控制器
 *
 * 提供物流公司的 CRUD 操作
 */
class Tanzanite_REST_Carriers_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'carriers';

	/**
	 * 物流公司表名
	 *
	 * @var string
	 */
	private $carriers_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->carriers_table = $wpdb->prefix . 'tanz_carriers';
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
					'permission_callback' => $this->permission_callback( 'tanz_view_shipping', true ),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_shipping', true ),
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个物流公司
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_view_shipping', true ),
					'args'                => array(
						'id' => array(
							'validate_callback' => 'is_numeric',
						),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_shipping', true ),
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_shipping', true ),
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
	 * 获取物流公司列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		$pagination = $this->get_pagination_params( $request );

		$where = array( '1=1' );
		$args  = array();

		// 状态筛选
		if ( $request->has_param( 'is_active' ) ) {
			$where[] = 'is_active = %d';
			$args[]  = (int) $request->get_param( 'is_active' );
		}

		// 搜索
		if ( $search = $request->get_param( 's' ) ) {
			$where[] = '(code LIKE %s OR name LIKE %s)';
			$like    = '%' . $wpdb->esc_like( $search ) . '%';
			$args[]  = $like;
			$args[]  = $like;
		}

		$where_sql = implode( ' AND ', $where );

		// 获取总数
		$total_sql = "SELECT COUNT(*) FROM {$this->carriers_table} WHERE {$where_sql}";
		$total     = (int) $wpdb->get_var( $args ? $wpdb->prepare( $total_sql, $args ) : $total_sql );

		// 获取数据
		$list_sql = "SELECT * FROM {$this->carriers_table} WHERE {$where_sql} ORDER BY sort_order ASC, id DESC LIMIT %d OFFSET %d";
		$rows     = $wpdb->get_results( $wpdb->prepare( $list_sql, array_merge( $args, array( $pagination['per_page'], $pagination['offset'] ) ) ), ARRAY_A );

		$items = array_map( array( $this, 'format_carrier_row' ), $rows );

		return rest_ensure_response(
			array(
				'ok'   => true,
				'data' => array(
					'items'      => $items,
					'pagination' => array(
						'page'        => $pagination['page'],
						'per_page'    => $pagination['per_page'],
						'total'       => $total,
						'total_pages' => ceil( $total / $pagination['per_page'] ),
					),
				),
			)
		);
	}

	/**
	 * 获取单个物流公司
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		global $wpdb;

		$id  = (int) $request['id'];
		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->carriers_table} WHERE id = %d", $id ), ARRAY_A );

		if ( ! $row ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => '物流公司不存在' ),
				),
				404
			);
		}

		return rest_ensure_response(
			array(
				'ok'   => true,
				'data' => $this->format_carrier_row( $row ),
			)
		);
	}

	/**
	 * 创建物流公司
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$data = $this->sanitize_carrier_request( $request, true );

		if ( is_wp_error( $data ) ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => $data->get_error_message() ),
				),
				400
			);
		}

		// 检查编码是否已存在
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->carriers_table} WHERE code = %s", $data['code'] ) );
		if ( $exists ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => '物流公司编码已存在' ),
				),
				400
			);
		}

		$data['created_at'] = current_time( 'mysql' );
		$data['updated_at'] = current_time( 'mysql' );

		$inserted = $wpdb->insert( $this->carriers_table, $data );

		if ( ! $inserted ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => '创建失败' ),
				),
				500
			);
		}

		$carrier_id = $wpdb->insert_id;
		$this->log_audit( 'create', 'carrier', $carrier_id, $data, $request );

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->carriers_table} WHERE id = %d", $carrier_id ), ARRAY_A );

		return rest_ensure_response(
			array(
				'ok'   => true,
				'data' => $this->format_carrier_row( $row ),
			)
		);
	}

	/**
	 * 更新物流公司
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$id     = (int) $request['id'];
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->carriers_table} WHERE id = %d", $id ) );

		if ( ! $exists ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => '物流公司不存在' ),
				),
				404
			);
		}

		$data = $this->sanitize_carrier_request( $request, false );

		if ( is_wp_error( $data ) ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => $data->get_error_message() ),
				),
				400
			);
		}

		// 检查编码冲突
		if ( isset( $data['code'] ) ) {
			$conflict = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->carriers_table} WHERE code = %s AND id != %d", $data['code'], $id ) );
			if ( $conflict ) {
				return new WP_REST_Response(
					array(
						'ok'   => false,
						'data' => array( 'message' => '物流公司编码已存在' ),
					),
					400
				);
			}
		}

		$data['updated_at'] = current_time( 'mysql' );
		$updated            = $wpdb->update( $this->carriers_table, $data, array( 'id' => $id ) );

		if ( false === $updated ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => '更新失败' ),
				),
				500
			);
		}

		$this->log_audit( 'update', 'carrier', $id, $data, $request );

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->carriers_table} WHERE id = %d", $id ), ARRAY_A );

		return rest_ensure_response(
			array(
				'ok'   => true,
				'data' => $this->format_carrier_row( $row ),
			)
		);
	}

	/**
	 * 删除物流公司
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$id     = (int) $request['id'];
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->carriers_table} WHERE id = %d", $id ) );

		if ( ! $exists ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => '物流公司不存在' ),
				),
				404
			);
		}

		$deleted = $wpdb->delete( $this->carriers_table, array( 'id' => $id ), array( '%d' ) );

		if ( ! $deleted ) {
			return new WP_REST_Response(
				array(
					'ok'   => false,
					'data' => array( 'message' => '删除失败' ),
				),
				500
			);
		}

		$this->log_audit( 'delete', 'carrier', $id, array(), $request );

		return rest_ensure_response(
			array(
				'ok'   => true,
				'data' => array( 'deleted' => true, 'id' => $id ),
			)
		);
	}

	/**
	 * 格式化物流公司数据行
	 *
	 * @since 0.2.0
	 * @param array|null $row 数据库行
	 * @return array|null
	 */
	private function format_carrier_row( $row ) {
		if ( ! $row ) {
			return null;
		}

		return array(
			'id'              => (int) $row['id'],
			'code'            => $row['code'],
			'name'            => $row['name'],
			'contact_person'  => $row['contact_person'],
			'contact_phone'   => $row['contact_phone'],
			'tracking_url'    => $row['tracking_url'],
			'service_regions' => $row['service_regions'] ? json_decode( $row['service_regions'], true ) : array(),
			'is_active'       => (bool) $row['is_active'],
			'sort_order'      => (int) $row['sort_order'],
			'meta'            => $row['meta'] ? json_decode( $row['meta'], true ) : null,
			'created_at'      => $row['created_at'],
			'updated_at'      => $row['updated_at'],
		);
	}

	/**
	 * 验证和清理物流公司请求数据
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request     REST 请求对象
	 * @param bool            $require_all 是否要求所有必填字段
	 * @return array|WP_Error
	 */
	private function sanitize_carrier_request( $request, $require_all = true ) {
		$data = array();

		if ( $request->has_param( 'code' ) ) {
			$data['code'] = sanitize_key( $request->get_param( 'code' ) );
		} elseif ( $require_all ) {
			return new WP_Error( 'missing_code', '物流公司编码不能为空' );
		}

		if ( $request->has_param( 'name' ) ) {
			$data['name'] = sanitize_text_field( $request->get_param( 'name' ) );
		} elseif ( $require_all ) {
			return new WP_Error( 'missing_name', '物流公司名称不能为空' );
		}

		if ( $request->has_param( 'contact_person' ) ) {
			$data['contact_person'] = sanitize_text_field( $request->get_param( 'contact_person' ) );
		}

		if ( $request->has_param( 'contact_phone' ) ) {
			$data['contact_phone'] = sanitize_text_field( $request->get_param( 'contact_phone' ) );
		}

		if ( $request->has_param( 'tracking_url' ) ) {
			$data['tracking_url'] = esc_url_raw( $request->get_param( 'tracking_url' ) );
		}

		if ( $request->has_param( 'service_regions' ) ) {
			$regions = $request->get_param( 'service_regions' );
			if ( is_array( $regions ) ) {
				$data['service_regions'] = wp_json_encode( array_map( 'sanitize_text_field', $regions ) );
			}
		}

		if ( $request->has_param( 'is_active' ) ) {
			$data['is_active'] = (int) (bool) $request->get_param( 'is_active' );
		} elseif ( $require_all ) {
			$data['is_active'] = 1;
		}

		if ( $request->has_param( 'sort_order' ) ) {
			$data['sort_order'] = (int) $request->get_param( 'sort_order' );
		} elseif ( $require_all ) {
			$data['sort_order'] = 0;
		}

		if ( $request->has_param( 'meta' ) ) {
			$meta = $request->get_param( 'meta' );
			if ( is_array( $meta ) || is_object( $meta ) ) {
				$data['meta'] = wp_json_encode( $meta );
			}
		}

		return $data;
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
				'type'              => 'string',
				'required'          => true,
				'sanitize_callback' => 'sanitize_key',
			),
			'name'            => array(
				'type'              => 'string',
				'required'          => true,
				'sanitize_callback' => 'sanitize_text_field',
			),
			'contact_person'  => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_text_field',
			),
			'contact_phone'   => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_text_field',
			),
			'tracking_url'    => array(
				'type'              => 'string',
				'sanitize_callback' => 'esc_url_raw',
			),
			'service_regions' => array(
				'type' => 'array',
			),
			'is_active'       => array(
				'type' => 'boolean',
			),
			'sort_order'      => array(
				'type' => 'integer',
			),
			'meta'            => array(
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
			'id'              => array(
				'validate_callback' => 'is_numeric',
			),
			'code'            => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_key',
			),
			'name'            => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_text_field',
			),
			'contact_person'  => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_text_field',
			),
			'contact_phone'   => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_text_field',
			),
			'tracking_url'    => array(
				'type'              => 'string',
				'sanitize_callback' => 'esc_url_raw',
			),
			'service_regions' => array(
				'type' => 'array',
			),
			'is_active'       => array(
				'type' => 'boolean',
			),
			'sort_order'      => array(
				'type' => 'integer',
			),
			'meta'            => array(
				'type' => 'object',
			),
		);
	}
}
