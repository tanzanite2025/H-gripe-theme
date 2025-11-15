<?php
/**
 * Shipping Templates REST API Controller
 *
 * 处理配送模板相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 配送模板 REST API 控制器
 *
 * 提供配送模板的 CRUD 操作
 */
class Tanzanite_REST_ShippingTemplates_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'shipping-templates';

	/**
	 * 配送模板表名
	 *
	 * @var string
	 */
	private $shipping_templates_table;

	/**
	 * 配送规则类型
	 *
	 * @var array
	 */
	private $shipping_rule_types = array( 'weight', 'quantity', 'volume', 'amount', 'items' );

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->shipping_templates_table = $wpdb->prefix . 'tanz_shipping_templates';
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
					'permission_callback' => $this->permission_callback( 'tanz_manage_shipping', true ),
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个配送模板
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
	 * 获取配送模板列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		$rows = $wpdb->get_results( "SELECT * FROM {$this->shipping_templates_table} ORDER BY id DESC LIMIT 100", ARRAY_A );

		return $this->respond_success(
			array(
				'items' => array_map( array( $this, 'format_shipping_template_row' ), $rows ),
				'meta'  => array(
					'rule_types' => $this->shipping_rule_types,
					'total'      => count( $rows ),
				),
			)
		);
	}

	/**
	 * 获取单个配送模板
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		global $wpdb;

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->shipping_templates_table} WHERE id = %d", (int) $request['id'] ), ARRAY_A );

		if ( ! $row ) {
			return $this->respond_error( 'shipping_template_not_found', __( 'Shipping template not found.', 'tanzanite-settings' ), 404 );
		}

		return $this->respond_success( $this->format_shipping_template_row( $row ) );
	}

	/**
	 * 创建配送模板
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$rules = $request->get_param( 'rules' );
		if ( ! is_array( $rules ) ) {
			$rules = array();
		}

		$rules = $this->sanitize_shipping_rules( $rules );
		if ( is_wp_error( $rules ) ) {
			return $this->respond_error( $rules->get_error_code(), $rules->get_error_message(), 400 );
		}

		$data = array(
			'template_name' => $request->get_param( 'template_name' ),
			'description'   => $request->get_param( 'description' ),
			'rules'         => wp_json_encode( $rules ),
			'is_active'     => 1,
			'meta'          => wp_json_encode( array( 'stage' => 'placeholder' ) ),
		);

		$inserted = $wpdb->insert( $this->shipping_templates_table, $data, array( '%s', '%s', '%s', '%d', '%s' ) );

		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_shipping_template', __( 'Failed to create shipping template.', 'tanzanite-settings' ), 500 );
		}

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->shipping_templates_table} WHERE id = %d", $wpdb->insert_id ), ARRAY_A );

		$this->log_audit( 'create', 'shipping_template', $wpdb->insert_id, array( 'template_name' => $data['template_name'] ), $request );

		return $this->respond_success( $this->format_shipping_template_row( $row ), 201 );
	}

	/**
	 * 更新配送模板
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->shipping_templates_table} WHERE id = %d", (int) $request['id'] ), ARRAY_A );
		if ( ! $row ) {
			return $this->respond_error( 'shipping_template_not_found', __( 'Shipping template not found.', 'tanzanite-settings' ), 404 );
		}

		$data  = array();
		$types = array();

		if ( $request->has_param( 'template_name' ) ) {
			$data['template_name'] = sanitize_text_field( $request->get_param( 'template_name' ) );
			$types[]               = '%s';
		}
		if ( $request->has_param( 'description' ) ) {
			$data['description'] = sanitize_textarea_field( $request->get_param( 'description' ) );
			$types[]             = '%s';
		}
		if ( $request->has_param( 'rules' ) ) {
			$rules = $request->get_param( 'rules' );
			if ( ! is_array( $rules ) ) {
				return $this->respond_error( 'invalid_rules', __( 'Rules must be an array.', 'tanzanite-settings' ), 400 );
			}

			$rules = $this->sanitize_shipping_rules( $rules );
			if ( is_wp_error( $rules ) ) {
				return $this->respond_error( $rules->get_error_code(), $rules->get_error_message(), 400 );
			}

			$data['rules'] = wp_json_encode( $rules );
			$types[]       = '%s';
		}
		if ( $request->has_param( 'is_active' ) ) {
			$data['is_active'] = $request->get_param( 'is_active' ) ? 1 : 0;
			$types[]           = '%d';
		}

		if ( ! empty( $data ) ) {
			$updated = $wpdb->update( $this->shipping_templates_table, $data, array( 'id' => (int) $request['id'] ), $types, array( '%d' ) );
			if ( false === $updated ) {
				return $this->respond_error( 'failed_update_shipping_template', __( 'Failed to update shipping template.', 'tanzanite-settings' ), 500 );
			}
		}

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->shipping_templates_table} WHERE id = %d", (int) $request['id'] ), ARRAY_A );

		$this->log_audit( 'update', 'shipping_template', (int) $request['id'], $data, $request );

		return $this->respond_success( $this->format_shipping_template_row( $row ) );
	}

	/**
	 * 删除配送模板
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT id FROM {$this->shipping_templates_table} WHERE id = %d", (int) $request['id'] ), ARRAY_A );
		if ( ! $row ) {
			return $this->respond_error( 'shipping_template_not_found', __( 'Shipping template not found.', 'tanzanite-settings' ), 404 );
		}

		$deleted = $wpdb->delete( $this->shipping_templates_table, array( 'id' => (int) $request['id'] ), array( '%d' ) );
		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_shipping_template', __( 'Failed to delete shipping template.', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'shipping_template', (int) $request['id'], array(), $request );

		return $this->respond_success( array( 'deleted' => true ) );
	}

	/**
	 * 格式化配送模板数据
	 *
	 * @since 0.2.0
	 * @param array $row 数据库行
	 * @return array
	 */
	private function format_shipping_template_row( $row ) {
		return array(
			'id'            => (int) $row['id'],
			'template_name' => $row['template_name'],
			'description'   => $row['description'],
			'is_active'     => (bool) $row['is_active'],
			'updated_at'    => $row['updated_at'],
			'rules'         => $this->decode_shipping_rules( $row['rules'] ),
		);
	}

	/**
	 * 解码配送规则
	 *
	 * @since 0.2.0
	 * @param string|null $json JSON 字符串
	 * @return array
	 */
	private function decode_shipping_rules( $json ) {
		if ( null === $json || '' === $json ) {
			return array();
		}

		$decoded = json_decode( $json, true );
		if ( ! is_array( $decoded ) ) {
			return array();
		}

		$result = $this->sanitize_shipping_rules( $decoded );
		if ( is_wp_error( $result ) ) {
			return array();
		}

		return $result;
	}

	/**
	 * 清理配送规则
	 *
	 * @since 0.2.0
	 * @param array $rules 规则数组
	 * @return array|WP_Error
	 */
	private function sanitize_shipping_rules( $rules ) {
		$sanitized = array();

		foreach ( $rules as $index => $rule ) {
			if ( ! is_array( $rule ) ) {
				return new WP_Error( 'invalid_rule', sprintf( __( 'Rule %d must be an object.', 'tanzanite-settings' ), $index + 1 ) );
			}

			$type = isset( $rule['type'] ) ? sanitize_key( $rule['type'] ) : '';
			if ( ! in_array( $type, $this->shipping_rule_types, true ) ) {
				return new WP_Error( 'invalid_rule_type', sprintf( __( 'Rule %d type is invalid.', 'tanzanite-settings' ), $index + 1 ) );
			}

			$min = isset( $rule['min'] ) && '' !== $rule['min'] ? (float) $rule['min'] : null;
			$max = isset( $rule['max'] ) && '' !== $rule['max'] ? (float) $rule['max'] : null;
			if ( null !== $min && null !== $max && $min > $max ) {
				return new WP_Error( 'invalid_rule_range', sprintf( __( 'Rule %1$d min cannot exceed max.', 'tanzanite-settings' ), $index + 1 ) );
			}

			$fee       = isset( $rule['fee'] ) ? (float) $rule['fee'] : 0.0;
			$priority  = isset( $rule['priority'] ) ? (int) $rule['priority'] : 0;
			$free_over = isset( $rule['free_over'] ) ? (float) $rule['free_over'] : null;

			$regions = array();
			if ( isset( $rule['regions'] ) && is_array( $rule['regions'] ) ) {
				foreach ( $rule['regions'] as $region ) {
					$regions[] = sanitize_text_field( $region );
				}
			}

			$sanitized[] = array_filter(
				array(
					'type'      => $type,
					'min'       => $min,
					'max'       => $max,
					'fee'       => $fee,
					'priority'  => $priority,
					'free_over' => $free_over,
					'regions'   => $regions,
				),
				function ( $value ) {
					if ( is_array( $value ) ) {
						return ! empty( $value );
					}

					return null !== $value;
				}
			);
		}

		return $sanitized;
	}

	/**
	 * 获取创建参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_create_params() {
		return array(
			'template_name' => array(
				'type'              => 'string',
				'required'          => true,
				'sanitize_callback' => 'sanitize_text_field',
			),
			'rules'         => array(
				'type'    => 'array',
				'default' => array(),
			),
			'description'   => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_textarea_field',
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
			'id'            => array(
				'validate_callback' => 'is_numeric',
			),
			'template_name' => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_text_field',
			),
			'description'   => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_textarea_field',
			),
			'rules'         => array(
				'type' => 'array',
			),
			'is_active'     => array(
				'type' => 'boolean',
			),
		);
	}
}
