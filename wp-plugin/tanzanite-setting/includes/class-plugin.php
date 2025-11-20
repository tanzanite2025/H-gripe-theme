<?php
/**
 * Tanzanite Settings Plugin Core Class
 *
 * 插件核心类，负责初始化和协调所有模块
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 插件核心类
 *
 * 负责插件的初始化、钩子注册和模块协调
 */
class Tanzanite_Plugin {

	/**
	 * 插件版本
	 *
	 * @var string
	 */
	const VERSION = '0.2.0';

	/**
	 * 数据库版本
	 *
	 * @var string
	 */
	const DB_VERSION = '0.1.9';

	/**
	 * 单例实例
	 *
	 * @var Tanzanite_Plugin
	 */
	private static $instance = null;

	/**
	 * REST API 控制器列表
	 *
	 * @var array
	 */
	private $rest_controllers = array();

	/**
	 * 后台页面列表
	 *
	 * @var array
	 */
	private $admin_pages = array();

	/**
	 * Legacy plugin instance
	 *
	 * @var Tanzanite_Settings_Plugin
	 */
	private $legacy_plugin = null;

	/**
	 * 获取单例实例
	 *
	 * @since 0.2.0
	 * @return Tanzanite_Plugin
	 */
	public static function get_instance() {
		if ( null === self::$instance ) {
			self::$instance = new self();
		}
		return self::$instance;
	}

	/**
	 * 构造函数（私有，防止直接实例化）
	 *
	 * @since 0.2.0
	 */
	private function __construct() {
		// 私有构造函数
	}

	/**
	 * 运行插件
	 *
	 * @since 0.2.0
	 */
	public function run() {
		$this->define_constants();
		$this->load_dependencies();
		$this->load_legacy_pages();
		$this->init_hooks();
	}

	/**
	 * 定义常量
	 *
	 * @since 0.2.0
	 */
	private function define_constants() {
		if ( ! defined( 'TANZANITE_VERSION' ) ) {
			define( 'TANZANITE_VERSION', self::VERSION );
		}
		if ( ! defined( 'TANZANITE_DB_VERSION' ) ) {
			define( 'TANZANITE_DB_VERSION', self::DB_VERSION );
		}
	}

	/**
	 * 加载依赖
	 *
	 * @since 0.2.0
	 */
	private function load_dependencies() {
		// 加载 URLLink 模块
		$this->load_urllink();
	}

	/**
	 * 加载 URLLink 模块
	 *
	 * @since 0.2.0
	 */
	private function load_urllink() {
		try {
			// 定义 URLLink 常量
			if ( ! defined( 'URLLINK_VERSION' ) ) {
				define( 'URLLINK_VERSION', '0.1.0' );
			}
			if ( ! defined( 'URLLINK_DIR' ) ) {
				define( 'URLLINK_DIR', TANZANITE_PLUGIN_DIR . 'includes/urllink/' );
			}
			if ( ! defined( 'URLLINK_URL' ) ) {
				define( 'URLLINK_URL', TANZANITE_PLUGIN_URL . 'includes/urllink/' );
			}
			
			// 检查文件是否存在
			$files = array(
				URLLINK_DIR . 'meta.php',
				URLLINK_DIR . 'rewrite.php',
				URLLINK_DIR . 'rest.php',
				URLLINK_DIR . 'admin.php',
				URLLINK_DIR . 'class-urllink-plugin.php',
			);
			
			foreach ( $files as $file ) {
				if ( ! file_exists( $file ) ) {
					error_log( 'URLLink file not found: ' . $file );
					return;
				}
			}
			
			// 加载 URLLink 文件（meta.php 必须最先加载，因为包含 urllink_normalize_path 函数）
			require_once URLLINK_DIR . 'meta.php';
			require_once URLLINK_DIR . 'rewrite.php';
			require_once URLLINK_DIR . 'rest.php';
			require_once URLLINK_DIR . 'admin.php';
			require_once URLLINK_DIR . 'class-urllink-plugin.php';
			
			// 初始化 URLLink
			if ( class_exists( 'URLLink_Plugin' ) ) {
				URLLink_Plugin::instance();
			}
		} catch ( Exception $e ) {
			error_log( 'URLLink load error: ' . $e->getMessage() );
		}
	}

	/**
	 * 加载旧的后台页面
	 *
	 * @since 0.2.0
	 */
	private function load_legacy_pages() {
		$legacy_file = TANZANITE_PLUGIN_DIR . 'includes/legacy-pages.php';
		
		if ( file_exists( $legacy_file ) ) {
			// 强制加载 legacy-pages.php，即使自动加载器尝试过
			if ( ! class_exists( 'Tanzanite_Settings_Plugin' ) ) {
				require_once $legacy_file;
			}
			
			if ( class_exists( 'Tanzanite_Settings_Plugin' ) ) {
				$this->legacy_plugin = Tanzanite_Settings_Plugin::instance();
				error_log( 'Tanzanite Plugin: Legacy plugin instance created' );
				
				// 移除 legacy plugin 的菜单注册，避免重复
				remove_action( 'admin_menu', array( $this->legacy_plugin, 'register_admin_menu' ) );
				
				// 保留 legacy plugin 的 REST API 注册
				// 保留 legacy plugin 的样式和脚本加载（enqueue_admin_assets）
				// 保留 legacy plugin 的 body class 过滤器（filter_admin_body_class）
			} else {
				error_log( 'Tanzanite Plugin: Failed to load Tanzanite_Settings_Plugin class' );
			}
		} else {
			error_log( 'Tanzanite Plugin: legacy-pages.php not found at ' . $legacy_file );
		}
	}

	/**
	 * 初始化钩子
	 *
	 * @since 0.2.0
	 */
	private function init_hooks() {
		// REST API 路由 - 直接注册，不依赖 legacy plugin
		add_action( 'rest_api_init', array( $this, 'register_rest_routes' ), 5 );

		// 后台菜单
		add_action( 'admin_menu', array( $this, 'register_admin_menu' ) );
		
		// 后台脚本和样式 - 调用 legacy plugin 的方法
		add_action( 'admin_enqueue_scripts', array( $this, 'enqueue_admin_assets' ) );
		
		error_log( 'Tanzanite Plugin: init_hooks() called, rest_api_init hook registered' );
	}

	/**
	 * 注册 REST API 路由
	 *
	 * @since 0.2.0
	 */
	public function register_rest_routes() {
		error_log( '=== Tanzanite Plugin: register_rest_routes() called in main plugin ===' );
		
		// 注册所有 REST API 控制器
		$controller_classes = array(
			'Tanzanite_REST_Orders_Controller',
			'Tanzanite_REST_Products_Controller',
			'Tanzanite_REST_Payments_Controller',
			'Tanzanite_REST_TaxRates_Controller',
			'Tanzanite_REST_Reviews_Controller',
			'Tanzanite_REST_Members_Controller',
			'Tanzanite_REST_Carriers_Controller',
			'Tanzanite_REST_Coupons_Controller',
			'Tanzanite_REST_Giftcards_Controller',
			'Tanzanite_REST_Redeem_Controller',
			'Tanzanite_REST_Loyalty_Controller',
			'Tanzanite_REST_Attributes_Controller',
			'Tanzanite_REST_Audit_Controller',
			'Tanzanite_REST_ShippingTemplates_Controller',
			'Tanzanite_REST_User_Assets_Controller',
			'Tanzanite_REST_Wishlist_Controller',
		);
		
		foreach ( $controller_classes as $class_name ) {
			try {
				if ( ! class_exists( $class_name ) ) {
					error_log( "Tanzanite Plugin: Controller class not found: {$class_name}" );
					continue;
				}
				
				$controller = new $class_name();
				$controller->register_routes();
				
				error_log( "Tanzanite Plugin: Registered routes for {$class_name}" );
			} catch ( Exception $e ) {
				error_log( "Tanzanite Plugin: Failed to register {$class_name}: " . $e->getMessage() );
			}
		}
	}

	/**
	 * 注册后台菜单
	 *
	 * @since 0.2.0
	 */
	public function register_admin_menu() {
		$root_capability = 'manage_options';
		$root_slug       = 'tanzanite-settings';

		// 添加主菜单
		add_menu_page(
			__( 'Tanzanite', 'tanzanite-settings' ),
			__( 'Tanzanite', 'tanzanite-settings' ),
			$root_capability,
			$root_slug,
			array( $this, 'render_all_products' ),
			'dashicons-store',
			56
		);

		// 商品列表
		add_submenu_page(
			$root_slug,
			__( 'All Products', 'tanzanite-settings' ),
			__( 'All Products', 'tanzanite-settings' ),
			$root_capability,
			$root_slug,
			array( $this, 'render_all_products' )
		);

		// 属性管理
		add_submenu_page(
			$root_slug,
			__( 'Attributes', 'tanzanite-settings' ),
			__( 'Attributes', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-attributes',
			array( $this, 'render_attributes' )
		);

		// 评论管理
		add_submenu_page(
			$root_slug,
			__( 'Reviews', 'tanzanite-settings' ),
			__( 'Reviews', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-reviews',
			array( $this, 'render_reviews' )
		);

		// 添加商品
		add_submenu_page(
			$root_slug,
			__( 'Add New Product', 'tanzanite-settings' ),
			__( 'Add New Product', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-add-product',
			array( $this, 'render_add_product' )
		);

		// 支付方式
		add_submenu_page(
			$root_slug,
			__( 'Payment Method', 'tanzanite-settings' ),
			__( 'Payment Method', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-payment-method',
			array( $this, 'render_payment_method' )
		);

		// 税率管理
		add_submenu_page(
			$root_slug,
			__( 'Tax Rates', 'tanzanite-settings' ),
			__( 'Tax Rates', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-tax-rates',
			array( $this, 'render_tax_rates' )
		);

		// 订单列表
		add_submenu_page(
			$root_slug,
			__( 'All Orders', 'tanzanite-settings' ),
			__( 'All Orders', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-orders',
			array( $this, 'render_orders_list' )
		);

		// 订单批量操作
		add_submenu_page(
			$root_slug,
			__( 'Order Bulk', 'tanzanite-settings' ),
			__( 'Order Bulk', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-orders-bulk',
			array( $this, 'render_orders_bulk' )
		);

		// 运费模板
		add_submenu_page(
			$root_slug,
			__( 'Shipping Templates', 'tanzanite-settings' ),
			__( 'Shipping Templates', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-shipping-templates',
			array( $this, 'render_shipping_templates' )
		);

		// 物流商管理
		add_submenu_page(
			$root_slug,
			__( 'Carriers & Tracking', 'tanzanite-settings' ),
			__( 'Carriers & Tracking', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-carriers',
			array( $this, 'render_carriers' )
		);

		// 会员档案
		add_submenu_page(
			$root_slug,
			__( 'Member Profiles', 'tanzanite-settings' ),
			__( 'Member Profiles', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-members',
			array( $this, 'render_member_profiles' )
		);

		// 礼品卡和优惠券
		add_submenu_page(
			$root_slug,
			__( 'Gift Cards & Coupons', 'tanzanite-settings' ),
			__( 'Gift Cards & Coupons', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-rewards',
			array( $this, 'render_rewards' )
		);

		// 积分设置
		add_submenu_page(
			$root_slug,
			__( 'Loyalty Settings', 'tanzanite-settings' ),
			__( 'Loyalty & Points', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-loyalty',
			array( $this, 'render_loyalty_settings' )
		);

		// 购物车管理
		add_submenu_page(
			$root_slug,
			__( 'Cart Management', 'tanzanite-settings' ),
			__( '🛒 Cart & Orders', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-cart-list',
			array( $this, 'render_cart_list' )
		);

		// 审计日志
		add_submenu_page(
			$root_slug,
			__( 'Audit Logs', 'tanzanite-settings' ),
			__( 'Audit Logs', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-audit-logs',
			array( $this, 'render_audit_logs' )
		);

		// URLLink - 自定义永久链接
		add_submenu_page(
			$root_slug,
			__( 'URL Management', 'tanzanite-settings' ),
			__( 'URL Management', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-urllink',
			array( $this, 'render_urllink' )
		);

		// SEO Settings
		add_submenu_page(
			$root_slug,
			__( 'SEO Settings', 'tanzanite-settings' ),
			__( 'SEO Settings', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-seo',
			array( $this, 'render_seo_page' )
		);

		// Markdown Templates
		add_submenu_page(
			$root_slug,
			__( 'Markdown Templates', 'tanzanite-settings' ),
			__( 'Markdown Templates', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-markdown-templates',
			array( $this, 'render_markdown_templates' )
		);
	}

	/**
	 * 加载后台资源
	 *
	 * @since 0.2.0
	 * @param string $hook 当前页面钩子
	 */
	public function enqueue_admin_assets( $hook ) {
		// 调用 legacy plugin 的方法
		if ( $this->legacy_plugin && method_exists( $this->legacy_plugin, 'enqueue_admin_assets' ) ) {
			$this->legacy_plugin->enqueue_admin_assets( $hook );
		}
	}

	/**
	 * 检查并升级数据库
	 *
	 * @since 0.2.0
	 */
	public function maybe_upgrade_database() {
		$stored_version = get_option( 'tanzanite_db_version' );
		
		if ( self::DB_VERSION !== $stored_version ) {
			// 将由 Database 类处理
			// Tanzanite_Database::upgrade();
			update_option( 'tanzanite_db_version', self::DB_VERSION );
		}
	}

	/**
	 * 插件激活钩子
	 *
	 * @since 0.2.0
	 */
	public static function activate() {
		// 创建数据库表
		// Tanzanite_Database::create_tables();
		
		// 创建角色和权限
		// Tanzanite_Permissions::create_roles();
		
		// 刷新重写规则
		flush_rewrite_rules();
	}

	/**
	 * 插件停用钩子
	 *
	 * @since 0.2.0
	 */
	public static function deactivate() {
		// 刷新重写规则
		flush_rewrite_rules();
	}

	/**
	 * 获取插件版本
	 *
	 * @since 0.2.0
	 * @return string
	 */
	public function get_version() {
		return self::VERSION;
	}

	/**
	 * 获取数据库版本
	 *
	 * @since 0.2.0
	 * @return string
	 */
	public function get_db_version() {
		return self::DB_VERSION;
	}

	/**
	 * 渲染商品列表页面
	 *
	 * @since 0.2.0
	 */
	public function render_all_products() {
		if ( $this->legacy_plugin && method_exists( $this->legacy_plugin, 'render_all_products' ) ) {
			$this->legacy_plugin->render_all_products();
		}
	}

	/**
	 * 渲染添加商品页面
	 *
	 * @since 0.2.0
	 */
	public function render_add_product() {
		$this->call_legacy_method( 'render_add_product' );
	}


	/**
	 * 渲染订单批量操作页面
	 *
	 * @since 0.2.0
	 */
	public function render_orders_bulk() {
		$this->call_legacy_method( 'render_orders_bulk' );
	}

	/**
	 * 渲染属性管理页面
	 *
	 * @since 0.2.0
	 */
	public function render_attributes() {
		$this->call_legacy_method( 'render_attributes' );
	}

	/**
	 * 渲染评论管理页面
	 *
	 * @since 0.2.0
	 */
	public function render_reviews() {
		$this->call_legacy_method( 'render_reviews' );
	}

	/**
	 * 渲染支付方式页面
	 *
	 * @since 0.2.0
	 */
	public function render_payment_method() {
		$this->call_legacy_method( 'render_payment_method' );
	}

	/**
	 * 渲染税率管理页面
	 *
	 * @since 0.2.0
	 */
	public function render_tax_rates() {
		$this->call_legacy_method( 'render_tax_rates' );
	}

	/**
	 * 渲染运费模板页面
	 *
	 * @since 0.2.0
	 */
	public function render_shipping_templates() {
		$this->call_legacy_method( 'render_shipping_templates' );
	}

	/**
	 * 渲染物流商管理页面
	 *
	 * @since 0.2.0
	 */
	public function render_carriers() {
		$this->call_legacy_method( 'render_carriers' );
	}

	/**
	 * 渲染会员档案页面
	 *
	 * @since 0.2.0
	 */
	public function render_member_profiles() {
		$this->call_legacy_method( 'render_member_profiles' );
	}

	/**
	 * 渲染礼品卡和优惠券页面
	 *
	 * @since 0.2.0
	 */
	public function render_rewards() {
		$this->call_legacy_method( 'render_rewards' );
	}

	/**
	 * 渲染积分设置页面
	 *
	 * @since 0.2.0
	 */
	public function render_loyalty_settings() {
		$this->call_legacy_method( 'render_loyalty_settings' );
	}

	/**
	 * 渲染审计日志页面
	 *
	 * @since 0.2.0
	 */
	public function render_audit_logs() {
		$this->call_legacy_method( 'render_audit_logs' );
	}

	/**
	 * 渲染 URLLink 页面
	 *
	 * @since 0.2.0
	 */
	public function render_urllink() {
		// URLLink 有自己的渲染逻辑
		if ( function_exists( 'urllink_render_admin_page' ) ) {
			urllink_render_admin_page();
		} elseif ( function_exists( 'urllink_admin_page' ) ) {
			urllink_admin_page();
		} else {
			echo '<div class="wrap"><h1>URL Management</h1>';
			echo '<div class="notice notice-error"><p>URLLink 渲染函数未找到。</p></div>';
			echo '</div>';
		}
	}

	/**
	 * 渲染 SEO 设置页面
	 *
	 * @since 0.2.0
	 */
	public function render_seo_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		// 调用 MyTheme SEO 的渲染方法
		if ( class_exists( 'MyTheme_SEO_Plugin' ) ) {
			$seo_instance = MyTheme_SEO_Plugin::instance();
			$seo_instance->render_admin_page();
		} else {
			echo '<div class="wrap">';
			echo '<h1>' . esc_html__( 'SEO Settings', 'tanzanite-settings' ) . '</h1>';
			echo '<p>' . esc_html__( 'MyTheme SEO 模块未正确加载。', 'tanzanite-settings' ) . '</p>';
			echo '</div>';
		}
	}

	/**
	 * 渲染 Markdown Templates 页面
	 *
	 * @since 0.2.0
	 */
	public function render_markdown_templates() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		// 调用 legacy plugin 的方法
		if ( $this->legacy_plugin && method_exists( $this->legacy_plugin, 'render_markdown_templates_page' ) ) {
			$this->legacy_plugin->render_markdown_templates_page();
		}
	}

	/**
	 * 渲染购物车列表页面
	 *
	 * @since 0.2.0
	 */
	public function render_cart_list() {
		// 手动加载类文件
		$class_file = TANZANITE_PLUGIN_DIR . 'includes/admin/class-cart-admin.php';
		if ( file_exists( $class_file ) ) {
			require_once $class_file;
		}
		
		if ( class_exists( 'Tanzanite_Cart_Admin' ) ) {
			Tanzanite_Cart_Admin::render_cart_list();
		} else {
			echo '<div class="wrap"><h1>错误</h1><p>Tanzanite_Cart_Admin 类未找到</p></div>';
		}
	}

	/**
	 * 调用 legacy 方法的辅助函数
	 *
	 * @since 0.2.0
	 * @param string $method 方法名
	 */
	private function call_legacy_method( $method ) {
		if ( $this->legacy_plugin && method_exists( $this->legacy_plugin, $method ) ) {
			$this->legacy_plugin->$method();
		}
	}

}
