<?php
/**
 * Attributes REST API Controller
 *
 * 处理商品属性相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 商品属性 REST API 控制器
 *
 * 提供商品属性和属性值的 CRUD 操作
 */
class Tanzanite_REST_Attributes_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'attributes';

	/**
	 * 商品属性表名
	 *
	 * @var string
	 */
	private $product_attributes_table;

	/**
	 * 属性值表名
	 *
	 * @var string
	 */
	private $attribute_values_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->product_attributes_table = $wpdb->prefix . 'tanz_product_attributes';
		$this->attribute_values_table   = $wpdb->prefix . 'tanz_attribute_values';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 属性列表和创建
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => array(
						'page'     => array( 'type' => 'integer', 'default' => 1 ),
						'per_page' => array( 'type' => 'integer', 'default' => 100 ),
					),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
			)
		);

		// 属性单个操作
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => 'is_user_logged_in',
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
			)
		);

		// 属性值列表和创建
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<attribute_id>\d+)/values',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_values' ),
					'permission_callback' => 'is_user_logged_in',
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_value' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
			)
		);

		// 属性值单个操作
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<attribute_id>\d+)/values/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_value' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_value' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
			)
		);
	}

	/**
	 * 获取属性列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		$page     = max( 1, (int) $request->get_param( 'page' ) );
		$per_page = max( 1, min( 100, (int) $request->get_param( 'per_page' ) ) );
		$offset   = ( $page - 1 ) * $per_page;

		$total = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$this->product_attributes_table}" );

		$rows = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->product_attributes_table} ORDER BY sort_order ASC, id ASC LIMIT %d OFFSET %d",
				$per_page,
				$offset
			),
			ARRAY_A
		);

		$items = array();
		foreach ( $rows as $row ) {
			$item = $this->format_attribute_row( $row );

			$value_count = (int) $wpdb->get_var(
				$wpdb->prepare(
					"SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d",
					$row['id']
				)
			);
			$item['values_count'] = $value_count;

			$items[] = $item;
		}

		return $this->respond_success(
			array(
				'items'       => $items,
				'total'       => $total,
				'page'        => $page,
				'per_page'    => $per_page,
				'total_pages' => ceil( $total / $per_page ),
			)
		);
	}
	
	/**
	 * 创建属性
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$name = sanitize_text_field( (string) $request->get_param( 'name' ) );
		if ( '' === $name ) {
			return $this->respond_error( 'invalid_attribute_payload', __( '属性名称不能为空。', 'tanzanite-settings' ) );
		}

		$slug = $request->get_param( 'slug' );
		if ( empty( $slug ) ) {
			$slug = sanitize_title( $name );
		} else {
			$slug = sanitize_title( $slug );
		}

		$exists = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->product_attributes_table} WHERE slug = %s",
				$slug
			)
		);

		if ( $exists > 0 ) {
			return $this->respond_error( 'duplicate_attribute_slug', __( '属性 Slug 已存在，请更换。', 'tanzanite-settings' ) );
		}

		$type = sanitize_key( (string) $request->get_param( 'type' ) );
		if ( ! in_array( $type, array( 'select', 'color', 'image' ), true ) ) {
			$type = 'select';
		}

		$meta = $request->get_param( 'meta' );
		if ( ! is_array( $meta ) ) {
			$meta = array();
		}

		$data = array(
			'name'          => $name,
			'slug'          => $slug,
			'type'          => $type,
			'is_filterable' => (bool) $request->get_param( 'is_filterable' ),
			'affects_sku'   => (bool) $request->get_param( 'affects_sku' ),
			'affects_stock' => (bool) $request->get_param( 'affects_stock' ),
			'is_enabled'    => (bool) $request->get_param( 'is_enabled' ),
			'sort_order'    => (int) $request->get_param( 'sort_order' ),
			'meta'          => wp_json_encode( $meta ),
		);

		$inserted = $wpdb->insert( $this->product_attributes_table, $data, array( '%s', '%s', '%s', '%d', '%d', '%d', '%d', '%d', '%s' ) );
		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_attribute', __( '创建属性失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$id                   = (int) $wpdb->insert_id;
		$row                  = $this->fetch_attribute_row( $id );
		$item                 = $this->format_attribute_row( $row );
		$item['values_count'] = 0;

		$this->log_audit( 'create', 'product_attribute', $id, array( 'name' => $item['name'] ), $request );

		return $this->respond_success( $item, 201 );
	}
	
	public function get_item( $request ) {
		$id  = (int) $request['id'];
		$row = $this->fetch_attribute_row( $id );

		if ( ! $row ) {
			return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
		}

		$item = $this->format_attribute_row( $row );

		global $wpdb;
		$value_count = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d",
				$id
			)
		);
		$item['values_count'] = $value_count;

		return $this->respond_success( $item );
	}

	public function update_item( $request ) {
		global $wpdb;

		$id  = (int) $request['id'];
		$row = $this->fetch_attribute_row( $id );

		if ( ! $row ) {
			return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
		}

		$data   = array();
		$format = array();

		if ( $request->has_param( 'name' ) ) {
			$name = sanitize_text_field( (string) $request->get_param( 'name' ) );
			if ( '' === $name ) {
				return $this->respond_error( 'invalid_attribute_payload', __( '属性名称不能为空。', 'tanzanite-settings' ) );
			}
			$data['name'] = $name;
			$format[]     = '%s';
		}

		if ( $request->has_param( 'slug' ) ) {
			$slug = sanitize_title( (string) $request->get_param( 'slug' ) );

			$exists = $wpdb->get_var(
				$wpdb->prepare(
					"SELECT COUNT(*) FROM {$this->product_attributes_table} WHERE slug = %s AND id != %d",
					$slug,
					$id
				)
			);

			if ( $exists > 0 ) {
				return $this->respond_error( 'duplicate_attribute_slug', __( '属性 Slug 已存在，请更换。', 'tanzanite-settings' ) );
			}

			$data['slug'] = $slug;
			$format[]     = '%s';
		}

		if ( $request->has_param( 'type' ) ) {
			$type = sanitize_key( (string) $request->get_param( 'type' ) );
			if ( ! in_array( $type, array( 'select', 'color', 'image' ), true ) ) {
				$type = 'select';
			}
			$data['type'] = $type;
			$format[]     = '%s';
		}

		if ( $request->has_param( 'is_filterable' ) ) {
			$data['is_filterable'] = (bool) $request->get_param( 'is_filterable' );
			$format[]              = '%d';
		}

		if ( $request->has_param( 'affects_sku' ) ) {
			$data['affects_sku'] = (bool) $request->get_param( 'affects_sku' );
			$format[]            = '%d';
		}

		if ( $request->has_param( 'affects_stock' ) ) {
			$data['affects_stock'] = (bool) $request->get_param( 'affects_stock' );
			$format[]              = '%d';
		}

		if ( $request->has_param( 'is_enabled' ) ) {
			$data['is_enabled'] = (bool) $request->get_param( 'is_enabled' );
			$format[]           = '%d';
		}

		if ( $request->has_param( 'sort_order' ) ) {
			$data['sort_order'] = (int) $request->get_param( 'sort_order' );
			$format[]           = '%d';
		}

		if ( $request->has_param( 'meta' ) ) {
			$meta = $request->get_param( 'meta' );
			if ( ! is_array( $meta ) ) {
				$meta = array();
			}
			$data['meta'] = wp_json_encode( $meta );
			$format[]     = '%s';
		}

		if ( empty( $data ) ) {
			return $this->respond_error( 'no_update_data', __( '没有需要更新的数据。', 'tanzanite-settings' ) );
		}

		$updated = $wpdb->update( $this->product_attributes_table, $data, array( 'id' => $id ), $format, array( '%d' ) );
		if ( false === $updated ) {
			return $this->respond_error( 'failed_update_attribute', __( '更新属性失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$updated_row = $this->fetch_attribute_row( $id );
		$item        = $this->format_attribute_row( $updated_row );

		$value_count = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d",
				$id
			)
		);
		$item['values_count'] = $value_count;

		$this->log_audit( 'update', 'product_attribute', $id, array( 'name' => $item['name'] ), $request );

		return $this->respond_success( $item );
	}

	public function delete_item( $request ) {
		global $wpdb;

		$id  = (int) $request['id'];
		$row = $this->fetch_attribute_row( $id );

		if ( ! $row ) {
			return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
		}

		$wpdb->delete( $this->attribute_values_table, array( 'attribute_id' => $id ), array( '%d' ) );

		$deleted = $wpdb->delete( $this->product_attributes_table, array( 'id' => $id ), array( '%d' ) );
		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_attribute', __( '删除属性失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'product_attribute', $id, array( 'name' => $row['name'] ), $request );

		return $this->respond_success( array( 'deleted' => true ) );
	}

	public function get_values( $request ) {
		global $wpdb;

		$attribute_id = (int) $request['attribute_id'];

		$attribute = $this->fetch_attribute_row( $attribute_id );
		if ( ! $attribute ) {
			return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
		}

		$rows = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->attribute_values_table} WHERE attribute_id = %d ORDER BY sort_order ASC, id ASC",
				$attribute_id
			),
			ARRAY_A
		);

		$items = array();
		foreach ( $rows as $row ) {
			$items[] = $this->format_attribute_value_row( $row );
		}

		return $this->respond_success(
			array(
				'items' => $items,
				'total' => count( $items ),
			)
		);
	}

	public function create_value( $request ) {
		global $wpdb;

		$attribute_id = (int) $request['attribute_id'];

		$attribute = $this->fetch_attribute_row( $attribute_id );
		if ( ! $attribute ) {
			return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
		}

		$name = sanitize_text_field( (string) $request->get_param( 'name' ) );
		if ( '' === $name ) {
			return $this->respond_error( 'invalid_value_payload', __( '属性值名称不能为空。', 'tanzanite-settings' ) );
		}

		$slug = $request->get_param( 'slug' );
		if ( empty( $slug ) ) {
			$slug = sanitize_title( $name );
		} else {
			$slug = sanitize_title( $slug );
		}

		$exists = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d AND slug = %s",
				$attribute_id,
				$slug
			)
		);

		if ( $exists > 0 ) {
			return $this->respond_error( 'duplicate_value_slug', __( '属性值 Slug 已存在，请更换。', 'tanzanite-settings' ) );
		}

		$value = sanitize_text_field( (string) $request->get_param( 'value' ) );
		$meta  = $request->get_param( 'meta' );
		if ( ! is_array( $meta ) ) {
			$meta = array();
		}

		$data = array(
			'attribute_id' => $attribute_id,
			'name'         => $name,
			'slug'         => $slug,
			'value'        => $value,
			'is_enabled'   => (bool) $request->get_param( 'is_enabled' ),
			'sort_order'   => (int) $request->get_param( 'sort_order' ),
			'meta'         => wp_json_encode( $meta ),
		);

		$inserted = $wpdb->insert( $this->attribute_values_table, $data, array( '%d', '%s', '%s', '%s', '%d', '%d', '%s' ) );
		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_value', __( '创建属性值失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$id   = (int) $wpdb->insert_id;
		$row  = $this->fetch_attribute_value_row( $id );
		$item = $this->format_attribute_value_row( $row );

		$this->log_audit( 'create', 'attribute_value', $id, array( 'name' => $item['name'], 'attribute_id' => $attribute_id ), $request );

		return $this->respond_success( $item, 201 );
	}

	public function update_value( $request ) {
		global $wpdb;

		$attribute_id = (int) $request['attribute_id'];
		$id           = (int) $request['id'];

		$row = $this->fetch_attribute_value_row( $id );
		if ( ! $row || (int) $row['attribute_id'] !== $attribute_id ) {
			return $this->respond_error( 'value_not_found', __( '指定的属性值不存在。', 'tanzanite-settings' ), 404 );
		}

		$data   = array();
		$format = array();

		if ( $request->has_param( 'name' ) ) {
			$name = sanitize_text_field( (string) $request->get_param( 'name' ) );
			if ( '' === $name ) {
				return $this->respond_error( 'invalid_value_payload', __( '属性值名称不能为空。', 'tanzanite-settings' ) );
			}
			$data['name'] = $name;
			$format[]     = '%s';
		}

		if ( $request->has_param( 'slug' ) ) {
			$slug = sanitize_title( (string) $request->get_param( 'slug' ) );

			$exists = $wpdb->get_var(
				$wpdb->prepare(
					"SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d AND slug = %s AND id != %d",
					$attribute_id,
					$slug,
					$id
				)
			);

			if ( $exists > 0 ) {
				return $this->respond_error( 'duplicate_value_slug', __( '属性值 Slug 已存在，请更换。', 'tanzanite-settings' ) );
			}

			$data['slug'] = $slug;
			$format[]     = '%s';
		}

		if ( $request->has_param( 'value' ) ) {
			$data['value'] = sanitize_text_field( (string) $request->get_param( 'value' ) );
			$format[]      = '%s';
		}

		if ( $request->has_param( 'is_enabled' ) ) {
			$data['is_enabled'] = (bool) $request->get_param( 'is_enabled' );
			$format[]           = '%d';
		}

		if ( $request->has_param( 'sort_order' ) ) {
			$data['sort_order'] = (int) $request->get_param( 'sort_order' );
			$format[]           = '%d';
		}

		if ( $request->has_param( 'meta' ) ) {
			$meta = $request->get_param( 'meta' );
			if ( ! is_array( $meta ) ) {
				$meta = array();
			}
			$data['meta'] = wp_json_encode( $meta );
			$format[]     = '%s';
		}

		if ( empty( $data ) ) {
			return $this->respond_error( 'no_update_data', __( '没有需要更新的数据。', 'tanzanite-settings' ) );
		}

		$updated = $wpdb->update( $this->attribute_values_table, $data, array( 'id' => $id ), $format, array( '%d' ) );
		if ( false === $updated ) {
			return $this->respond_error( 'failed_update_value', __( '更新属性值失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$updated_row = $this->fetch_attribute_value_row( $id );
		$item        = $this->format_attribute_value_row( $updated_row );

		$this->log_audit( 'update', 'attribute_value', $id, array( 'name' => $item['name'] ), $request );

		return $this->respond_success( $item );
	}

	public function delete_value( $request ) {
		global $wpdb;

		$attribute_id = (int) $request['attribute_id'];
		$id           = (int) $request['id'];

		$row = $this->fetch_attribute_value_row( $id );
		if ( ! $row || (int) $row['attribute_id'] !== $attribute_id ) {
			return $this->respond_error( 'value_not_found', __( '指定的属性值不存在。', 'tanzanite-settings' ), 404 );
		}

		$deleted = $wpdb->delete( $this->attribute_values_table, array( 'id' => $id ), array( '%d' ) );
		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_value', __( '删除属性值失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'delete', 'attribute_value', $id, array( 'name' => $row['name'] ), $request );

		return $this->respond_success( array( 'deleted' => true ) );
	}

	private function fetch_attribute_value_row( $id ) {
		global $wpdb;
		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->attribute_values_table} WHERE id = %d", $id ), ARRAY_A );
		return $row ?: null;
	}

	private function format_attribute_value_row( $row ) {
		$meta = $row['meta'] ? json_decode( $row['meta'], true ) : array();
		if ( ! is_array( $meta ) ) {
			$meta = array();
		}

		return array(
			'id'           => (int) $row['id'],
			'attribute_id' => (int) $row['attribute_id'],
			'name'         => $row['name'],
			'slug'         => $row['slug'],
			'value'        => $row['value'],
			'is_enabled'   => (bool) $row['is_enabled'],
			'sort_order'   => (int) $row['sort_order'],
			'meta'         => $meta,
			'created_at'   => $row['created_at'],
			'updated_at'   => $row['updated_at'],
		);
	}

	private function fetch_attribute_row( $id ) {
		global $wpdb;
		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->product_attributes_table} WHERE id = %d", $id ), ARRAY_A );
		return $row ?: null;
	}

	private function format_attribute_row( $row ) {
		$meta = $row['meta'] ? json_decode( $row['meta'], true ) : array();
		if ( ! is_array( $meta ) ) {
			$meta = array();
		}

		return array(
			'id'            => (int) $row['id'],
			'name'          => $row['name'],
			'slug'          => $row['slug'],
			'type'          => $row['type'],
			'is_filterable' => (bool) $row['is_filterable'],
			'affects_sku'   => (bool) $row['affects_sku'],
			'affects_stock' => (bool) $row['affects_stock'],
			'is_enabled'    => (bool) $row['is_enabled'],
			'sort_order'    => (int) $row['sort_order'],
			'meta'          => $meta,
			'created_at'    => $row['created_at'],
			'updated_at'    => $row['updated_at'],
		);
	}
}
