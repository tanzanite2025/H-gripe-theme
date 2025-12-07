<?php
/**
 * Packaging Rules REST API Controller
 *
 * 处理包装规则相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.3.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 包装规则 REST API 控制器
 *
 * 提供包装规则的 CRUD 操作
 */
class Tanzanite_REST_Packaging_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'packaging-rules';

	/**
	 * 包装规则表名
	 *
	 * @var string
	 */
	private $packaging_rules_table;

	/**
	 * 包装规则适用范围表名
	 *
	 * @var string
	 */
	private $packaging_applies_table;

	/**
	 * 构造函数
	 *
	 * @since 0.3.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->packaging_rules_table   = $wpdb->prefix . 'tanz_packaging_rules';
		$this->packaging_applies_table = $wpdb->prefix . 'tanz_packaging_rule_applies';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.3.0
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
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个包装规则
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
							'validate_callback' => array( $this, 'validate_numeric_param' ),
						),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
					'args'                => array(
						'id' => array(
							'validate_callback' => array( $this, 'validate_numeric_param' ),
						),
					),
				),
			)
		);

		// 安装数据库表（仅管理员）
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/install',
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'install_tables' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
				),
			)
		);
	}

	/**
	 * 安装数据库表
	 *
	 * @since 0.3.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function install_tables( $request ) {
		global $wpdb;

		require_once ABSPATH . 'wp-admin/includes/upgrade.php';

		$charset = $wpdb->get_charset_collate();

		// 包装规则表
		$rules_sql = "CREATE TABLE {$this->packaging_rules_table} (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			rule_name VARCHAR(100) NOT NULL,
			description TEXT NULL,
			box_weight DECIMAL(10,3) NOT NULL DEFAULT 0,
			box_length DECIMAL(10,2) DEFAULT NULL,
			box_width DECIMAL(10,2) DEFAULT NULL,
			box_height DECIMAL(10,2) DEFAULT NULL,
			max_items INT DEFAULT NULL,
			max_weight DECIMAL(10,3) DEFAULT NULL,
			priority INT DEFAULT 0,
			is_active TINYINT(1) DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_active_priority (is_active, priority)
		) {$charset};";

		dbDelta( $rules_sql );

		// 包装规则适用范围表
		$applies_sql = "CREATE TABLE {$this->packaging_applies_table} (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			rule_id BIGINT UNSIGNED NOT NULL,
			apply_type VARCHAR(20) NOT NULL,
			apply_value VARCHAR(100) DEFAULT NULL,
			PRIMARY KEY (id),
			KEY idx_rule_id (rule_id),
			KEY idx_rule_type (rule_id, apply_type)
		) {$charset};";

		dbDelta( $applies_sql );

		// 检查表是否创建成功
		$rules_exists   = $wpdb->get_var( "SHOW TABLES LIKE '{$this->packaging_rules_table}'" ) === $this->packaging_rules_table;
		$applies_exists = $wpdb->get_var( "SHOW TABLES LIKE '{$this->packaging_applies_table}'" ) === $this->packaging_applies_table;

		if ( $rules_exists && $applies_exists ) {
			return $this->respond_success(
				array(
					'message' => __( 'Database tables created successfully.', 'tanzanite-settings' ),
					'tables'  => array(
						$this->packaging_rules_table,
						$this->packaging_applies_table,
					),
				)
			);
		}

		return $this->respond_error(
			'table_creation_failed',
			__( 'Failed to create database tables.', 'tanzanite-settings' ),
			500
		);
	}

	/**
	 * 获取包装规则列表
	 *
	 * @since 0.3.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		// 检查表是否存在
		$table_exists = $wpdb->get_var( "SHOW TABLES LIKE '{$this->packaging_rules_table}'" ) === $this->packaging_rules_table;
		if ( ! $table_exists ) {
			return $this->respond_success(
				array(
					'items'         => array(),
					'table_missing' => true,
					'message'       => __( 'Database table not found. Please install first.', 'tanzanite-settings' ),
				)
			);
		}

		$rows = $wpdb->get_results( "SELECT * FROM {$this->packaging_rules_table} ORDER BY priority DESC, id DESC LIMIT 100", ARRAY_A );

		$items = array();
		foreach ( $rows as $row ) {
			$items[] = $this->format_rule_row( $row );
		}

		return $this->respond_success(
			array(
				'items' => $items,
				'meta'  => array(
					'total'       => count( $items ),
					'apply_types' => array( 'category', 'tag', 'product', 'all' ),
				),
			)
		);
	}

	/**
	 * 获取单个包装规则
	 *
	 * @since 0.3.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		global $wpdb;

		$row = $wpdb->get_row(
			$wpdb->prepare( "SELECT * FROM {$this->packaging_rules_table} WHERE id = %d", (int) $request['id'] ),
			ARRAY_A
		);

		if ( ! $row ) {
			return $this->respond_error( 'packaging_rule_not_found', __( 'Packaging rule not found.', 'tanzanite-settings' ), 404 );
		}

		return $this->respond_success( $this->format_rule_row( $row ) );
	}

	/**
	 * 创建包装规则
	 *
	 * @since 0.3.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$data = array(
			'rule_name'   => sanitize_text_field( $request->get_param( 'rule_name' ) ),
			'description' => sanitize_textarea_field( $request->get_param( 'description' ) ),
			'box_weight'  => (float) $request->get_param( 'box_weight' ),
			'box_length'  => $request->get_param( 'box_length' ) !== null ? (float) $request->get_param( 'box_length' ) : null,
			'box_width'   => $request->get_param( 'box_width' ) !== null ? (float) $request->get_param( 'box_width' ) : null,
			'box_height'  => $request->get_param( 'box_height' ) !== null ? (float) $request->get_param( 'box_height' ) : null,
			'max_items'   => $request->get_param( 'max_items' ) !== null ? (int) $request->get_param( 'max_items' ) : null,
			'max_weight'  => $request->get_param( 'max_weight' ) !== null ? (float) $request->get_param( 'max_weight' ) : null,
			'priority'    => (int) $request->get_param( 'priority' ),
			'is_active'   => $request->get_param( 'is_active' ) !== false ? 1 : 0,
		);

		$inserted = $wpdb->insert( $this->packaging_rules_table, $data );

		if ( false === $inserted ) {
			return $this->respond_error( 'create_failed', __( 'Failed to create packaging rule.', 'tanzanite-settings' ), 500 );
		}

		$rule_id = $wpdb->insert_id;

		// 保存适用范围
		$applies_to = $request->get_param( 'applies_to' );
		if ( is_array( $applies_to ) ) {
			$this->save_applies_to( $rule_id, $applies_to );
		}

		// 获取创建的规则
		$row = $wpdb->get_row(
			$wpdb->prepare( "SELECT * FROM {$this->packaging_rules_table} WHERE id = %d", $rule_id ),
			ARRAY_A
		);

		return $this->respond_success( $this->format_rule_row( $row ) );
	}

	/**
	 * 更新包装规则
	 *
	 * @since 0.3.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$id = (int) $request['id'];

		// 检查规则是否存在
		$exists = $wpdb->get_var( $wpdb->prepare( "SELECT id FROM {$this->packaging_rules_table} WHERE id = %d", $id ) );
		if ( ! $exists ) {
			return $this->respond_error( 'packaging_rule_not_found', __( 'Packaging rule not found.', 'tanzanite-settings' ), 404 );
		}

		$data = array();

		if ( $request->get_param( 'rule_name' ) !== null ) {
			$data['rule_name'] = sanitize_text_field( $request->get_param( 'rule_name' ) );
		}
		if ( $request->get_param( 'description' ) !== null ) {
			$data['description'] = sanitize_textarea_field( $request->get_param( 'description' ) );
		}
		if ( $request->get_param( 'box_weight' ) !== null ) {
			$data['box_weight'] = (float) $request->get_param( 'box_weight' );
		}
		if ( $request->get_param( 'box_length' ) !== null ) {
			$data['box_length'] = (float) $request->get_param( 'box_length' );
		}
		if ( $request->get_param( 'box_width' ) !== null ) {
			$data['box_width'] = (float) $request->get_param( 'box_width' );
		}
		if ( $request->get_param( 'box_height' ) !== null ) {
			$data['box_height'] = (float) $request->get_param( 'box_height' );
		}
		if ( $request->get_param( 'max_items' ) !== null ) {
			$data['max_items'] = (int) $request->get_param( 'max_items' );
		}
		if ( $request->get_param( 'max_weight' ) !== null ) {
			$data['max_weight'] = (float) $request->get_param( 'max_weight' );
		}
		if ( $request->get_param( 'priority' ) !== null ) {
			$data['priority'] = (int) $request->get_param( 'priority' );
		}
		if ( $request->get_param( 'is_active' ) !== null ) {
			$data['is_active'] = $request->get_param( 'is_active' ) ? 1 : 0;
		}

		if ( ! empty( $data ) ) {
			$wpdb->update( $this->packaging_rules_table, $data, array( 'id' => $id ) );
		}

		// 更新适用范围
		$applies_to = $request->get_param( 'applies_to' );
		if ( is_array( $applies_to ) ) {
			$this->save_applies_to( $id, $applies_to );
		}

		// 获取更新后的规则
		$row = $wpdb->get_row(
			$wpdb->prepare( "SELECT * FROM {$this->packaging_rules_table} WHERE id = %d", $id ),
			ARRAY_A
		);

		return $this->respond_success( $this->format_rule_row( $row ) );
	}

	/**
	 * 删除包装规则
	 *
	 * @since 0.3.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$id = (int) $request['id'];

		// 删除适用范围
		$wpdb->delete( $this->packaging_applies_table, array( 'rule_id' => $id ) );

		// 删除规则
		$deleted = $wpdb->delete( $this->packaging_rules_table, array( 'id' => $id ) );

		if ( false === $deleted ) {
			return $this->respond_error( 'delete_failed', __( 'Failed to delete packaging rule.', 'tanzanite-settings' ), 500 );
		}

		return $this->respond_success( array( 'deleted' => true ) );
	}

	/**
	 * 保存适用范围
	 *
	 * @since 0.3.0
	 * @param int   $rule_id    规则 ID
	 * @param array $applies_to 适用范围数组
	 */
	private function save_applies_to( $rule_id, $applies_to ) {
		global $wpdb;

		// 先删除旧的
		$wpdb->delete( $this->packaging_applies_table, array( 'rule_id' => $rule_id ) );

		// 插入新的
		foreach ( $applies_to as $apply ) {
			if ( ! isset( $apply['type'] ) ) {
				continue;
			}

			$type  = sanitize_key( $apply['type'] );
			$value = isset( $apply['value'] ) ? sanitize_text_field( $apply['value'] ) : null;

			if ( ! in_array( $type, array( 'category', 'tag', 'product', 'all' ), true ) ) {
				continue;
			}

			$wpdb->insert(
				$this->packaging_applies_table,
				array(
					'rule_id'     => $rule_id,
					'apply_type'  => $type,
					'apply_value' => $value,
				)
			);
		}
	}

	/**
	 * 获取规则的适用范围
	 *
	 * @since 0.3.0
	 * @param int $rule_id 规则 ID
	 * @return array
	 */
	private function get_applies_to( $rule_id ) {
		global $wpdb;

		$rows = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT apply_type, apply_value FROM {$this->packaging_applies_table} WHERE rule_id = %d",
				$rule_id
			),
			ARRAY_A
		);

		$result = array();
		foreach ( $rows as $row ) {
			$result[] = array(
				'type'  => $row['apply_type'],
				'value' => $row['apply_value'],
			);
		}

		return $result;
	}

	/**
	 * 格式化规则数据
	 *
	 * @since 0.3.0
	 * @param array $row 数据库行
	 * @return array
	 */
	private function format_rule_row( $row ) {
		return array(
			'id'          => (int) $row['id'],
			'rule_name'   => $row['rule_name'],
			'description' => $row['description'],
			'box_weight'  => (float) $row['box_weight'],
			'box_length'  => $row['box_length'] !== null ? (float) $row['box_length'] : null,
			'box_width'   => $row['box_width'] !== null ? (float) $row['box_width'] : null,
			'box_height'  => $row['box_height'] !== null ? (float) $row['box_height'] : null,
			'max_items'   => $row['max_items'] !== null ? (int) $row['max_items'] : null,
			'max_weight'  => $row['max_weight'] !== null ? (float) $row['max_weight'] : null,
			'priority'    => (int) $row['priority'],
			'is_active'   => (bool) $row['is_active'],
			'created_at'  => $row['created_at'],
			'updated_at'  => $row['updated_at'],
			'applies_to'  => $this->get_applies_to( (int) $row['id'] ),
		);
	}

	/**
	 * 获取创建参数定义
	 *
	 * @since 0.3.0
	 * @return array
	 */
	private function get_create_params() {
		return array(
			'rule_name'   => array(
				'type'              => 'string',
				'required'          => true,
				'sanitize_callback' => 'sanitize_text_field',
			),
			'description' => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_textarea_field',
			),
			'box_weight'  => array(
				'type'    => 'number',
				'default' => 0,
			),
			'box_length'  => array(
				'type' => 'number',
			),
			'box_width'   => array(
				'type' => 'number',
			),
			'box_height'  => array(
				'type' => 'number',
			),
			'max_items'   => array(
				'type' => 'integer',
			),
			'max_weight'  => array(
				'type' => 'number',
			),
			'priority'    => array(
				'type'    => 'integer',
				'default' => 0,
			),
			'is_active'   => array(
				'type'    => 'boolean',
				'default' => true,
			),
			'applies_to'  => array(
				'type'    => 'array',
				'default' => array(),
			),
		);
	}

	/**
	 * 获取更新参数定义
	 *
	 * @since 0.3.0
	 * @return array
	 */
	private function get_update_params() {
		return array(
			'id'          => array(
				'validate_callback' => array( $this, 'validate_numeric_param' ),
			),
			'rule_name'   => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_text_field',
			),
			'description' => array(
				'type'              => 'string',
				'sanitize_callback' => 'sanitize_textarea_field',
			),
			'box_weight'  => array(
				'type' => 'number',
			),
			'box_length'  => array(
				'type' => 'number',
			),
			'box_width'   => array(
				'type' => 'number',
			),
			'box_height'  => array(
				'type' => 'number',
			),
			'max_items'   => array(
				'type' => 'integer',
			),
			'max_weight'  => array(
				'type' => 'number',
			),
			'priority'    => array(
				'type' => 'integer',
			),
			'is_active'   => array(
				'type' => 'boolean',
			),
			'applies_to'  => array(
				'type' => 'array',
			),
		);
	}
}
