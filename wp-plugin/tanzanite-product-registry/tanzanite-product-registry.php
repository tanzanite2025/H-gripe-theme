<?php
/**
 * Plugin Name: Tanzanite Product Registry
 * Plugin URI: https://tanzanite.cc
 * Description: 产品编码与保修查询系统 - 管理产品编码、客户信息、保修记录
 * Version: 1.0.0
 * Author: Tanzanite
 * Author URI: https://tanzanite.cc
 * Text Domain: tanzanite-product-registry
 * Domain Path: /languages
 * Requires at least: 5.8
 * Requires PHP: 7.4
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

// 插件常量
define( 'TANZANITE_PR_VERSION', '1.0.0' );
define( 'TANZANITE_PR_PLUGIN_DIR', plugin_dir_path( __FILE__ ) );
define( 'TANZANITE_PR_PLUGIN_URL', plugin_dir_url( __FILE__ ) );
define( 'TANZANITE_PR_PLUGIN_BASENAME', plugin_basename( __FILE__ ) );

/**
 * 主插件类
 */
final class Tanzanite_Product_Registry {

	/**
	 * 单例实例
	 *
	 * @var Tanzanite_Product_Registry
	 */
	private static $instance = null;

	/**
	 * 数据库类实例
	 *
	 * @var Tanzanite_PR_Database
	 */
	public $database;

	/**
	 * 获取单例实例
	 *
	 * @return Tanzanite_Product_Registry
	 */
	public static function instance() {
		if ( is_null( self::$instance ) ) {
			self::$instance = new self();
		}
		return self::$instance;
	}

	/**
	 * 构造函数
	 */
	private function __construct() {
		$this->includes();
		$this->init_hooks();
	}

	/**
	 * 加载依赖文件
	 */
	private function includes() {
		// 数据库类
		require_once TANZANITE_PR_PLUGIN_DIR . 'includes/class-database.php';
		
		// 后台管理
		if ( is_admin() ) {
			require_once TANZANITE_PR_PLUGIN_DIR . 'includes/admin/class-admin-menu.php';
			require_once TANZANITE_PR_PLUGIN_DIR . 'includes/admin/class-product-types-admin.php';
			require_once TANZANITE_PR_PLUGIN_DIR . 'includes/admin/class-products-admin.php';
			require_once TANZANITE_PR_PLUGIN_DIR . 'includes/admin/class-warranty-records-admin.php';
			require_once TANZANITE_PR_PLUGIN_DIR . 'includes/admin/class-import-export.php';
		}
		
		// REST API
		require_once TANZANITE_PR_PLUGIN_DIR . 'includes/rest-api/class-rest-warranty-controller.php';
	}

	/**
	 * 初始化钩子
	 */
	private function init_hooks() {
		// 激活/停用钩子
		register_activation_hook( __FILE__, array( $this, 'activate' ) );
		register_deactivation_hook( __FILE__, array( $this, 'deactivate' ) );

		// 初始化
		add_action( 'init', array( $this, 'init' ) );
		
		// 注册 REST API
		add_action( 'rest_api_init', array( $this, 'register_rest_routes' ) );
		
		// 加载后台资源
		add_action( 'admin_enqueue_scripts', array( $this, 'admin_enqueue_scripts' ) );
	}

	/**
	 * 插件激活
	 */
	public function activate() {
		$this->database = new Tanzanite_PR_Database();
		$this->database->create_tables();
		$this->database->insert_default_data();
		
		flush_rewrite_rules();
	}

	/**
	 * 插件停用
	 */
	public function deactivate() {
		flush_rewrite_rules();
	}

	/**
	 * 初始化
	 */
	public function init() {
		$this->database = new Tanzanite_PR_Database();
		
		// 初始化后台管理
		if ( is_admin() ) {
			new Tanzanite_PR_Admin_Menu();
			new Tanzanite_PR_Product_Types_Admin();
			new Tanzanite_PR_Products_Admin();
			new Tanzanite_PR_Warranty_Records_Admin();
			new Tanzanite_PR_Import_Export();
		}
	}

	/**
	 * 注册 REST API 路由
	 */
	public function register_rest_routes() {
		$controller = new Tanzanite_REST_Warranty_Controller();
		$controller->register_routes();
	}

	/**
	 * 加载后台资源
	 *
	 * @param string $hook 当前页面钩子
	 */
	public function admin_enqueue_scripts( $hook ) {
		// 只在插件页面加载
		if ( strpos( $hook, 'tanzanite-pr' ) === false ) {
			return;
		}

		wp_enqueue_style(
			'tanzanite-pr-admin',
			TANZANITE_PR_PLUGIN_URL . 'assets/css/admin.css',
			array(),
			TANZANITE_PR_VERSION
		);

		wp_enqueue_script(
			'tanzanite-pr-admin',
			TANZANITE_PR_PLUGIN_URL . 'assets/js/admin.js',
			array( 'jquery' ),
			TANZANITE_PR_VERSION,
			true
		);

		wp_localize_script(
			'tanzanite-pr-admin',
			'tanzanitePR',
			array(
				'ajaxUrl' => admin_url( 'admin-ajax.php' ),
				'nonce'   => wp_create_nonce( 'tanzanite_pr_nonce' ),
				'i18n'    => array(
					'confirmDelete' => '确定要删除吗？此操作不可撤销。',
					'saving'        => '保存中...',
					'saved'         => '已保存',
					'error'         => '操作失败，请重试',
				),
			)
		);
	}
}

/**
 * 获取插件实例
 *
 * @return Tanzanite_Product_Registry
 */
function tanzanite_product_registry() {
	return Tanzanite_Product_Registry::instance();
}

// 启动插件
tanzanite_product_registry();
