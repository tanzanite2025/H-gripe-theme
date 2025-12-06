<?php
/**
 * 数据库管理类
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 数据库操作类
 */
class Tanzanite_PR_Database {

	/**
	 * 表前缀
	 *
	 * @var string
	 */
	private $prefix;

	/**
	 * 产品类型表名
	 *
	 * @var string
	 */
	public $product_types_table;

	/**
	 * 产品登记表名
	 *
	 * @var string
	 */
	public $products_table;

	/**
	 * 保修记录表名
	 *
	 * @var string
	 */
	public $warranty_records_table;

	/**
	 * 构造函数
	 */
	public function __construct() {
		global $wpdb;
		$this->prefix                 = $wpdb->prefix . 'tanz_pr_';
		$this->product_types_table    = $this->prefix . 'product_types';
		$this->products_table         = $this->prefix . 'products';
		$this->warranty_records_table = $this->prefix . 'warranty_records';
	}

	/**
	 * 创建数据库表
	 */
	public function create_tables() {
		global $wpdb;
		$charset_collate = $wpdb->get_charset_collate();

		// 产品类型表
		$sql_product_types = "CREATE TABLE IF NOT EXISTS {$this->product_types_table} (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			type_code VARCHAR(50) NOT NULL,
			type_name VARCHAR(100) NOT NULL,
			type_name_en VARCHAR(100) NOT NULL,
			default_warranty_months INT UNSIGNED NOT NULL DEFAULT 36,
			sort_order INT UNSIGNED NOT NULL DEFAULT 0,
			is_active TINYINT(1) NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY type_code (type_code)
		) $charset_collate;";

		// 产品登记表
		$sql_products = "CREATE TABLE IF NOT EXISTS {$this->products_table} (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			product_code VARCHAR(100) NOT NULL,
			product_type_id INT UNSIGNED NOT NULL,
			product_name VARCHAR(200) DEFAULT '',
			ship_date DATE NOT NULL,
			warranty_months INT UNSIGNED NOT NULL DEFAULT 36,
			order_id VARCHAR(100) DEFAULT '',
			customer_name VARCHAR(100) DEFAULT '',
			customer_email VARCHAR(100) DEFAULT '',
			customer_phone VARCHAR(50) DEFAULT '',
			notes TEXT,
			created_by INT UNSIGNED DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY product_code (product_code),
			KEY product_type_id (product_type_id),
			KEY order_id (order_id),
			KEY customer_email (customer_email),
			KEY ship_date (ship_date)
		) $charset_collate;";

		// 保修记录表
		$sql_warranty_records = "CREATE TABLE IF NOT EXISTS {$this->warranty_records_table} (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			product_id INT UNSIGNED NOT NULL,
			record_type ENUM('repair', 'extend', 'replace') NOT NULL,
			record_date DATE NOT NULL,
			description TEXT,
			extend_months INT UNSIGNED DEFAULT 0,
			operator VARCHAR(100) DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY product_id (product_id),
			KEY record_type (record_type),
			KEY record_date (record_date)
		) $charset_collate;";

		require_once ABSPATH . 'wp-admin/includes/upgrade.php';
		dbDelta( $sql_product_types );
		dbDelta( $sql_products );
		dbDelta( $sql_warranty_records );
	}

	/**
	 * 插入默认数据
	 */
	public function insert_default_data() {
		global $wpdb;

		// 检查是否已有数据
		$count = $wpdb->get_var( "SELECT COUNT(*) FROM {$this->product_types_table}" );
		if ( $count > 0 ) {
			return;
		}

		// 默认产品类型
		$default_types = array(
			array(
				'type_code'               => 'hub',
				'type_name'               => '花鼓',
				'type_name_en'            => 'Hub',
				'default_warranty_months' => 36,
				'sort_order'              => 1,
			),
			array(
				'type_code'               => 'rim',
				'type_name'               => '轮圈',
				'type_name_en'            => 'Rim',
				'default_warranty_months' => 36,
				'sort_order'              => 2,
			),
			array(
				'type_code'               => 'wheelset',
				'type_name'               => '轮组',
				'type_name_en'            => 'Wheelset',
				'default_warranty_months' => 36,
				'sort_order'              => 3,
			),
			array(
				'type_code'               => 'spoke',
				'type_name'               => '辐条',
				'type_name_en'            => 'Spoke',
				'default_warranty_months' => 36,
				'sort_order'              => 4,
			),
			array(
				'type_code'               => 'other',
				'type_name'               => '其他',
				'type_name_en'            => 'Other',
				'default_warranty_months' => 36,
				'sort_order'              => 99,
			),
		);

		foreach ( $default_types as $type ) {
			$wpdb->insert( $this->product_types_table, $type );
		}
	}

	/**
	 * 获取所有产品类型
	 *
	 * @param bool $active_only 是否只获取启用的
	 * @return array
	 */
	public function get_product_types( $active_only = false ) {
		global $wpdb;
		$where = $active_only ? 'WHERE is_active = 1' : '';
		return $wpdb->get_results(
			"SELECT * FROM {$this->product_types_table} {$where} ORDER BY sort_order ASC, id ASC",
			ARRAY_A
		);
	}

	/**
	 * 获取单个产品类型
	 *
	 * @param int $id 类型ID
	 * @return array|null
	 */
	public function get_product_type( $id ) {
		global $wpdb;
		return $wpdb->get_row(
			$wpdb->prepare( "SELECT * FROM {$this->product_types_table} WHERE id = %d", $id ),
			ARRAY_A
		);
	}

	/**
	 * 保存产品类型
	 *
	 * @param array $data 类型数据
	 * @return int|false 插入的ID或false
	 */
	public function save_product_type( $data ) {
		global $wpdb;

		$id = isset( $data['id'] ) ? intval( $data['id'] ) : 0;
		unset( $data['id'] );

		if ( $id > 0 ) {
			$wpdb->update( $this->product_types_table, $data, array( 'id' => $id ) );
			return $id;
		} else {
			$wpdb->insert( $this->product_types_table, $data );
			return $wpdb->insert_id;
		}
	}

	/**
	 * 删除产品类型
	 *
	 * @param int $id 类型ID
	 * @return bool
	 */
	public function delete_product_type( $id ) {
		global $wpdb;
		return $wpdb->delete( $this->product_types_table, array( 'id' => $id ) ) !== false;
	}

	/**
	 * 获取产品列表
	 *
	 * @param array $args 查询参数
	 * @return array
	 */
	public function get_products( $args = array() ) {
		global $wpdb;

		$defaults = array(
			'search'    => '',
			'type_id'   => 0,
			'page'      => 1,
			'per_page'  => 20,
			'orderby'   => 'created_at',
			'order'     => 'DESC',
		);
		$args = wp_parse_args( $args, $defaults );

		$where = array( '1=1' );
		$values = array();

		if ( ! empty( $args['search'] ) ) {
			$where[] = '(p.product_code LIKE %s OR p.product_name LIKE %s OR p.customer_name LIKE %s OR p.order_id LIKE %s)';
			$search = '%' . $wpdb->esc_like( $args['search'] ) . '%';
			$values = array_merge( $values, array( $search, $search, $search, $search ) );
		}

		if ( $args['type_id'] > 0 ) {
			$where[] = 'p.product_type_id = %d';
			$values[] = $args['type_id'];
		}

		$where_sql = implode( ' AND ', $where );
		$offset = ( $args['page'] - 1 ) * $args['per_page'];

		$orderby = sanitize_sql_orderby( $args['orderby'] . ' ' . $args['order'] );
		if ( ! $orderby ) {
			$orderby = 'p.created_at DESC';
		} else {
			$orderby = 'p.' . $orderby;
		}

		$sql = "SELECT p.*, t.type_name, t.type_name_en, t.type_code
				FROM {$this->products_table} p
				LEFT JOIN {$this->product_types_table} t ON p.product_type_id = t.id
				WHERE {$where_sql}
				ORDER BY {$orderby}
				LIMIT %d OFFSET %d";

		$values[] = $args['per_page'];
		$values[] = $offset;

		$items = $wpdb->get_results( $wpdb->prepare( $sql, $values ), ARRAY_A );

		// 获取总数
		$count_sql = "SELECT COUNT(*) FROM {$this->products_table} p WHERE {$where_sql}";
		$total = $wpdb->get_var( $wpdb->prepare( $count_sql, array_slice( $values, 0, -2 ) ) );

		return array(
			'items' => $items,
			'total' => intval( $total ),
			'pages' => ceil( $total / $args['per_page'] ),
		);
	}

	/**
	 * 获取单个产品
	 *
	 * @param int $id 产品ID
	 * @return array|null
	 */
	public function get_product( $id ) {
		global $wpdb;
		return $wpdb->get_row(
			$wpdb->prepare(
				"SELECT p.*, t.type_name, t.type_name_en, t.type_code
				FROM {$this->products_table} p
				LEFT JOIN {$this->product_types_table} t ON p.product_type_id = t.id
				WHERE p.id = %d",
				$id
			),
			ARRAY_A
		);
	}

	/**
	 * 通过编码获取产品
	 *
	 * @param string $code 产品编码
	 * @return array|null
	 */
	public function get_product_by_code( $code ) {
		global $wpdb;
		return $wpdb->get_row(
			$wpdb->prepare(
				"SELECT p.*, t.type_name, t.type_name_en, t.type_code
				FROM {$this->products_table} p
				LEFT JOIN {$this->product_types_table} t ON p.product_type_id = t.id
				WHERE p.product_code = %s",
				$code
			),
			ARRAY_A
		);
	}

	/**
	 * 保存产品
	 *
	 * @param array $data 产品数据
	 * @return int|false
	 */
	public function save_product( $data ) {
		global $wpdb;

		$id = isset( $data['id'] ) ? intval( $data['id'] ) : 0;
		unset( $data['id'] );

		if ( $id > 0 ) {
			$wpdb->update( $this->products_table, $data, array( 'id' => $id ) );
			return $id;
		} else {
			$data['created_by'] = get_current_user_id();
			$wpdb->insert( $this->products_table, $data );
			return $wpdb->insert_id;
		}
	}

	/**
	 * 删除产品
	 *
	 * @param int $id 产品ID
	 * @return bool
	 */
	public function delete_product( $id ) {
		global $wpdb;
		// 同时删除相关的保修记录
		$wpdb->delete( $this->warranty_records_table, array( 'product_id' => $id ) );
		return $wpdb->delete( $this->products_table, array( 'id' => $id ) ) !== false;
	}

	/**
	 * 获取产品的保修记录
	 *
	 * @param int $product_id 产品ID
	 * @return array
	 */
	public function get_warranty_records( $product_id ) {
		global $wpdb;
		return $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$this->warranty_records_table} WHERE product_id = %d ORDER BY record_date DESC, id DESC",
				$product_id
			),
			ARRAY_A
		);
	}

	/**
	 * 保存保修记录
	 *
	 * @param array $data 记录数据
	 * @return int|false
	 */
	public function save_warranty_record( $data ) {
		global $wpdb;

		$id = isset( $data['id'] ) ? intval( $data['id'] ) : 0;
		unset( $data['id'] );

		if ( $id > 0 ) {
			$wpdb->update( $this->warranty_records_table, $data, array( 'id' => $id ) );
			return $id;
		} else {
			$wpdb->insert( $this->warranty_records_table, $data );
			return $wpdb->insert_id;
		}
	}

	/**
	 * 删除保修记录
	 *
	 * @param int $id 记录ID
	 * @return bool
	 */
	public function delete_warranty_record( $id ) {
		global $wpdb;
		return $wpdb->delete( $this->warranty_records_table, array( 'id' => $id ) ) !== false;
	}

	/**
	 * 计算产品的总延保月数
	 *
	 * @param int $product_id 产品ID
	 * @return int
	 */
	public function get_total_extend_months( $product_id ) {
		global $wpdb;
		$total = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT SUM(extend_months) FROM {$this->warranty_records_table} WHERE product_id = %d AND record_type = 'extend'",
				$product_id
			)
		);
		return intval( $total );
	}
}
