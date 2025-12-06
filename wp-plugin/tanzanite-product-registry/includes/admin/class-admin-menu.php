<?php
/**
 * 后台菜单管理
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 后台菜单类
 */
class Tanzanite_PR_Admin_Menu {

	/**
	 * 构造函数
	 */
	public function __construct() {
		add_action( 'admin_menu', array( $this, 'add_menu' ) );
	}

	/**
	 * 添加菜单
	 */
	public function add_menu() {
		// 主菜单
		add_menu_page(
			'产品管理',
			'产品管理',
			'manage_options',
			'tanzanite-pr',
			array( $this, 'render_products_page' ),
			'dashicons-archive',
			30
		);

		// 产品列表（子菜单）
		add_submenu_page(
			'tanzanite-pr',
			'所有产品',
			'所有产品',
			'manage_options',
			'tanzanite-pr',
			array( $this, 'render_products_page' )
		);

		// 添加产品
		add_submenu_page(
			'tanzanite-pr',
			'添加产品',
			'添加产品',
			'manage_options',
			'tanzanite-pr-add',
			array( $this, 'render_add_product_page' )
		);

		// 产品类型
		add_submenu_page(
			'tanzanite-pr',
			'产品类型',
			'产品类型',
			'manage_options',
			'tanzanite-pr-types',
			array( $this, 'render_types_page' )
		);

		// 保修记录
		add_submenu_page(
			'tanzanite-pr',
			'保修记录',
			'保修记录',
			'manage_options',
			'tanzanite-pr-records',
			array( $this, 'render_records_page' )
		);

		// 批量导入/导出
		add_submenu_page(
			'tanzanite-pr',
			'导入/导出',
			'导入/导出',
			'manage_options',
			'tanzanite-pr-import',
			array( $this, 'render_import_page' )
		);
	}

	/**
	 * 渲染产品列表页面
	 */
	public function render_products_page() {
		include TANZANITE_PR_PLUGIN_DIR . 'includes/admin/views/products-list.php';
	}

	/**
	 * 渲染添加/编辑产品页面
	 */
	public function render_add_product_page() {
		include TANZANITE_PR_PLUGIN_DIR . 'includes/admin/views/product-edit.php';
	}

	/**
	 * 渲染产品类型页面
	 */
	public function render_types_page() {
		include TANZANITE_PR_PLUGIN_DIR . 'includes/admin/views/product-types.php';
	}

	/**
	 * 渲染保修记录页面
	 */
	public function render_records_page() {
		include TANZANITE_PR_PLUGIN_DIR . 'includes/admin/views/warranty-records.php';
	}

	/**
	 * 渲染导入/导出页面
	 */
	public function render_import_page() {
		include TANZANITE_PR_PLUGIN_DIR . 'includes/admin/views/import-export.php';
	}
}
