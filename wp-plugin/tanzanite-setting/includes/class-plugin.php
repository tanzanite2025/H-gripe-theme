<?php
/**
 * Tanzanite Settings Plugin Core Class
 *
 * Plugin core class, responsible for initialization and coordination of all modules.
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 鎻掍欢鏍稿績绫? *
 * 璐熻矗鎻掍欢鐨勫垵濮嬪寲銆侀挬瀛愭敞鍐屽拰妯″潡鍗忚皟
 */
class Tanzanite_Plugin {

	/**
	 * 鎻掍欢鐗堟湰
	 *
	 * @var string
	 */
	const VERSION = '0.2.0';

	/**
	 * 鏁版嵁搴撶増鏈?	 *
	 * @var string
	 */
	const DB_VERSION = '0.1.9';

	/**
	 * 鍗曚緥瀹炰緥
	 *
	 * @var Tanzanite_Plugin
	 */
	private static $instance = null;

	/**
	 * REST API 鎺у埗鍣ㄥ垪琛?	 *
	 * @var array
	 */
	private $rest_controllers = array();

	/**
	 * 鍚庡彴椤甸潰鍒楄〃
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
	 * 鑾峰彇鍗曚緥瀹炰緥
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
	 * 私有构造函数（防止直接实例化）
	 */
	private function __construct() {
		// 私有构造函数
	}

	/**
	 * 	权限检查
	 *
	 * @since 0.2.0
	 */
	public function run() {
		$this->define_constants();
		$this->load_dependencies();
		if ( class_exists( 'Tanzanite_Suggestion_Feedback_Database' ) ) {
			Tanzanite_Suggestion_Feedback_Database::maybe_upgrade();
		}
		$this->load_legacy_pages();
		$this->init_hooks();
	}

	/**
	 * 瀹氫箟甯搁噺
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
	 * 鍔犺浇渚濊禆
	 *
	 * @since 0.2.0
	 */
	private function load_dependencies() {
		// 鍔犺浇 URLLink 妯″潡
		$this->load_urllink();

		$geometry_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-product-geometry-admin.php';
		if ( file_exists( $geometry_admin ) ) {
			require_once $geometry_admin;
		}

		// 加载 Spoke Geometry Admin（独立后台页面用）
		$spoke_geometry_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-spoke-geometry-admin.php';
		if ( file_exists( $spoke_geometry_admin ) ) {
			require_once $spoke_geometry_admin;
		}

		// 鍔犺浇 Orders Admin (Refactored)
		$orders_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-orders-admin.php';
		if ( file_exists( $orders_admin ) ) {
			require_once $orders_admin;
		}

		// 鍔犺浇 Reviews Admin
		$reviews_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-reviews-admin.php';
		if ( file_exists( $reviews_admin ) ) {
			require_once $reviews_admin;
		}

		// 鍔犺浇 Carriers Admin
		$carriers_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-carriers-admin.php';
		if ( file_exists( $carriers_admin ) ) {
			require_once $carriers_admin;
		}

		// 鍔犺浇 Shipping Templates Admin
		$shipping_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-shipping-admin.php';
		if ( file_exists( $shipping_admin ) ) {
			require_once $shipping_admin;
		}

		// 加载 Packaging Rules Admin
		$packaging_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-packaging-admin.php';
		if ( file_exists( $packaging_admin ) ) {
			require_once $packaging_admin;
		}

		// 鍔犺浇 Members Admin
		$members_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-members-admin.php';
		if ( file_exists( $members_admin ) ) {
			require_once $members_admin;
		}

		// 鍔犺浇 Rewards Admin
		$rewards_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-rewards-admin.php';
		if ( file_exists( $rewards_admin ) ) {
			require_once $rewards_admin;
		}

		// 鍔犺浇 Loyalty Admin
		$loyalty_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-loyalty-admin.php';
		if ( file_exists( $loyalty_admin ) ) {
			require_once $loyalty_admin;
		}

		// 鍔犺浇 Payment Admin
		$payment_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-payment-admin.php';
		if ( file_exists( $payment_admin ) ) {
			require_once $payment_admin;
		}

		// 鍔犺浇 Tax Rates Admin
		$tax_rates_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-tax-rates-admin.php';
		if ( file_exists( $tax_rates_admin ) ) {
			require_once $tax_rates_admin;
		}

		// 鍔犺浇 Markdown Templates Admin
		$markdown_templates_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-markdown-templates-admin.php';
		if ( file_exists( $markdown_templates_admin ) ) {
			require_once $markdown_templates_admin;
		}

		// 鍔犺浇 Products List Admin
		$products_list_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-products-list-admin.php';
		if ( file_exists( $products_list_admin ) ) {
			require_once $products_list_admin;
		}

		// 加载 Add Product Admin
		$add_product_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-add-product-admin.php';
		if ( file_exists( $add_product_admin ) ) {
			require_once $add_product_admin;
		}

		// 加载 Attributes Admin
		$attributes_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-attributes-admin.php';
		if ( file_exists( $attributes_admin ) ) {
			require_once $attributes_admin;
		}

		$suggestion_db = TANZANITE_PLUGIN_DIR . 'includes/database/class-suggestion-feedback-database.php';
		if ( file_exists( $suggestion_db ) ) {
			require_once $suggestion_db;
		}

		$suggestion_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-suggestion-feedback-admin.php';
		if ( file_exists( $suggestion_admin ) ) {
			require_once $suggestion_admin;
		}
	}

	/**
	 * 鍔犺浇 URLLink 妯″潡
	 *
	 * @since 0.2.0
	 */
	private function load_urllink() {
		try {
			// 瀹氫箟 URLLink 甯搁噺
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
			
			// 鍔犺浇 URLLink 鏂囦欢锛坢eta.php 蹇呴』鏈€鍏堝姞杞斤紝鍥犱负鍖呭惈 urllink_normalize_path 鍑芥暟锛?			require_once URLLINK_DIR . 'meta.php';
			require_once URLLINK_DIR . 'rewrite.php';
			require_once URLLINK_DIR . 'rest.php';
			require_once URLLINK_DIR . 'admin.php';
			require_once URLLINK_DIR . 'class-urllink-plugin.php';
			
			// 鍒濆鍖?URLLink
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

		// Spoke Length History 导入处理
		add_action( 'admin_post_tanz_spoke_history_import', array( $this, 'handle_spoke_history_import' ) );

		// 商品几何参数 meta box（仅在后台商品编辑页使用）
		if ( is_admin() && class_exists( 'Tanzanite_Product_Geometry_Admin' ) ) {
			Tanzanite_Product_Geometry_Admin::init();
		}
		
		// 后台脚本和样式 - 调用 legacy plugin 的方法
		add_action( 'admin_enqueue_scripts', array( $this, 'enqueue_admin_assets' ) );
		
		// 添加自定义 body class
		add_filter( 'admin_body_class', array( $this, 'filter_admin_body_class' ) );

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
			// 'Tanzanite_REST_Audit_Controller',
			'Tanzanite_REST_ShippingTemplates_Controller',
			'Tanzanite_REST_Packaging_Controller',
			'Tanzanite_REST_User_Assets_Controller',
			'Tanzanite_REST_Wishlist_Controller',
			// 新增：辐条计算器专用商品列表控制器
			'Tanzanite_REST_Spoke_Products_Controller',
			// 新增：用户反馈 / 留言控制器
			'Tanzanite_REST_Feedback_Controller',
			'Tanzanite_REST_Suggestion_Feedback_Controller',
			// 新增：辐条长度历史搜索控制器
			'Tanzanite_REST_Spoke_History_Controller',
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

		// 娣诲姞鍟嗗搧
		add_submenu_page(
			$root_slug,
			__( 'Add New Product', 'tanzanite-settings' ),
			__( 'Add New Product', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-add-product',
			array( $this, 'render_add_product' )
		);

		// 鏀粯鏂瑰紡
		add_submenu_page(
			$root_slug,
			__( 'Payment Method', 'tanzanite-settings' ),
			__( 'Payment Method', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-payment-method',
			array( $this, 'render_payment_method' )
		);

		// 绋庣巼绠＄悊
		add_submenu_page(
			$root_slug,
			__( 'Tax Rates', 'tanzanite-settings' ),
			__( 'Tax Rates', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-tax-rates',
			array( $this, 'render_tax_rates' )
		);

		// 璁㈠崟鍒楄〃
		add_submenu_page(
			$root_slug,
			__( 'All Orders', 'tanzanite-settings' ),
			__( 'All Orders', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-orders',
			array( $this, 'render_orders_list' )
		);

		// Order Bulk
		add_submenu_page(
			$root_slug,
			__( 'Order Bulk', 'tanzanite-settings' ),
			__( 'Order Bulk', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-orders-bulk',
			array( $this, 'render_orders_bulk' )
		);

		// Shipping Templates
		add_submenu_page(
			$root_slug,
			__( 'Shipping Templates', 'tanzanite-settings' ),
			__( 'Shipping Templates', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-shipping-templates',
			array( $this, 'render_shipping_templates' )
		);

		// Packaging Rules
		add_submenu_page(
			$root_slug,
			__( 'Packaging Rules', 'tanzanite-settings' ),
			__( 'Packaging Rules', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-packaging-rules',
			array( $this, 'render_packaging_rules' )
		);

		// Carriers
		add_submenu_page(
			$root_slug,
			__( 'Carriers & Tracking', 'tanzanite-settings' ),
			__( 'Carriers & Tracking', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-carriers',
			array( $this, 'render_carriers' )
		);

		// Member Profiles
		add_submenu_page(
			$root_slug,
			__( 'Member Profiles', 'tanzanite-settings' ),
			__( 'Member Profiles', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-members',
			array( $this, 'render_member_profiles' )
		);

		// Rewards
		add_submenu_page(
			$root_slug,
			__( 'Gift Cards & Coupons', 'tanzanite-settings' ),
			__( 'Gift Cards & Coupons', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-rewards',
			array( $this, 'render_rewards' )
		);

		// Loyalty
		add_submenu_page(
			$root_slug,
			__( 'Loyalty Settings', 'tanzanite-settings' ),
			__( 'Loyalty & Points', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-loyalty',
			array( $this, 'render_loyalty_settings' )
		);

		// Cart
		add_submenu_page(
			$root_slug,
			__( 'Cart Management', 'tanzanite-settings' ),
			__( 'Cart & Orders', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-cart-list',
			array( $this, 'render_cart_list' )
		);

		// Spoke Geometry
		add_submenu_page(
			$root_slug,
			__( 'Spoke Geometry', 'tanzanite-settings' ),
			__( 'Spoke Geometry', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-spoke-geometry',
			array( $this, 'render_spoke_geometry_page' )
		);

		// Spoke History
		add_submenu_page(
			$root_slug,
			__( 'Spoke Length History', 'tanzanite-settings' ),
			__( 'Spoke Length History', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-spoke-history',
			array( $this, 'render_spoke_history_page' )
		);


		// URLLink
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

		add_submenu_page(
			$root_slug,
			__( 'Suggestion Feedback', 'tanzanite-settings' ),
			__( 'Suggestion Feedback', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-suggestion-feedback',
			array( $this, 'render_suggestion_feedback' )
		);
	}

	/**
	 * 加载后台资源
	 *
	 * @since 0.2.0
	 * @param string $hook 当前页面钩子
	 */
	public function enqueue_admin_assets( $hook ) {
		// 加载全局样式
		// 注意：admin.css 和 admin.min.css 内容相同，保留两个文件是为了符合 WordPress 标准
		// 生产环境自动加载 .min.css，开发环境（SCRIPT_DEBUG=true）加载 .css
		$suffix = defined( 'SCRIPT_DEBUG' ) && SCRIPT_DEBUG ? '' : '.min';
		wp_enqueue_style(
			'tanzanite-settings-admin',
			TANZANITE_PLUGIN_URL . 'assets/css/admin' . $suffix . '.css',
			[],
			self::VERSION
		);

		// 注册公共 JS 库
		wp_register_script(
			'tz-admin-common',
			TANZANITE_PLUGIN_URL . 'assets/js/admin-common.js',
			array( 'jquery' ),
			self::VERSION,
			true
		);

		// Attributes 页面 JS 加载 - 迁移后逻辑
		$screen = get_current_screen();
		if ( $screen && strpos( $screen->id, 'tanzanite-settings-attributes' ) !== false ) {
			wp_enqueue_media();
			wp_enqueue_script(
				'tz-attributes',
				TANZANITE_PLUGIN_URL . 'assets/js/attributes.js',
				array( 'jquery', 'wp-media' ),
				self::VERSION,
				true
			);
		}
	}

	/**
	 * 在插件页面添加自定义 body class 以便样式隔离。
	 */
	public function filter_admin_body_class( $classes ) {
		$screen = get_current_screen();

		// 为所有 Tanzanite 主设置页、购物车列表页、Spoke Geometry 页
		// Tanzanite Photos（tanz_photo 列表/编辑页），以及 Subscription Broadcasts
		// 页面统一应用 tz-settings-admin 样式
		if (
			$screen
			&& (
				false !== strpos( $screen->id, 'tanzanite-settings' )
				|| false !== strpos( $screen->id, 'tanzanite-cart' )
				|| false !== strpos( $screen->id, 'tanzanite-spoke-geometry' )
				|| false !== strpos( $screen->id, 'tanz_photo' )
				|| false !== strpos( $screen->id, 'tanzanite-subscription-broadcasts' )
			)
		) {
			$classes .= ' tz-settings-admin';
		}

		return $classes;
	}

	/**
	 * 检查并升级数据库	 *
	 * @since 0.2.0
	 */
	public function maybe_upgrade_database() {
		$stored_version = get_option( 'tanzanite_db_version' );
		
		if ( self::DB_VERSION !== $stored_version ) {
			// 灏嗙敱 Database 绫诲鐞?			// Tanzanite_Database::upgrade();
			update_option( 'tanzanite_db_version', self::DB_VERSION );
		}
	}

	/**
	 * 鎻掍欢婵€娲婚挬瀛?	 *
	 * @since 0.2.0
	 */
	public static function activate() {
		if ( class_exists( 'Tanzanite_Suggestion_Feedback_Database' ) ) {
			Tanzanite_Suggestion_Feedback_Database::ensure_tables();
		}
		
		// 鍒涘缓瑙掕壊鍜屾潈闄?		// Tanzanite_Permissions::create_roles();
		
		// 鍒锋柊閲嶅啓瑙勫垯
		flush_rewrite_rules();
	}

	/**
	 * 鎻掍欢鍋滅敤閽╁瓙
	 *
	 * @since 0.2.0
	 */
	public static function deactivate() {
		// 鍒锋柊閲嶅啓瑙勫垯
		flush_rewrite_rules();
	}

	/**
	 * 鑾峰彇鎻掍欢鐗堟湰
	 *
	 * @since 0.2.0
	 * @return string
	 */
	public function get_version() {
		return self::VERSION;
	}

	/**
	 * 鑾峰彇鏁版嵁搴撶増鏈?	 *
	 * @since 0.2.0
	 * @return string
	 */
	public function get_db_version() {
		return self::DB_VERSION;
	}

	/**
	 * 娓叉煋鍟嗗搧鍒楄〃椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_all_products() {
		if ( class_exists( 'Tanzanite_Products_List_Admin' ) ) {
			Tanzanite_Products_List_Admin::render_page();
		} elseif ( $this->legacy_plugin && method_exists( $this->legacy_plugin, 'render_all_products' ) ) {
			$this->legacy_plugin->render_all_products();
		}
	}

	/**
	 * 娓叉煋娣诲姞鍟嗗搧椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_add_product() {
		if ( class_exists( 'Tanzanite_Add_Product_Admin' ) ) {
			Tanzanite_Add_Product_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_add_product' );
		}
	}

	public function render_orders_list() {
		if ( class_exists( 'Tanzanite_Orders_Admin' ) ) {
			Tanzanite_Orders_Admin::render_list_page();
		} else {
			$this->call_legacy_method( 'render_orders_list' );
		}
	}

	/**
	 * 娓叉煋璁㈠崟鎵归噺鎿嶄綔椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_orders_bulk() {
		if ( class_exists( 'Tanzanite_Orders_Admin' ) ) {
			Tanzanite_Orders_Admin::render_bulk_page();
		} else {
			$this->call_legacy_method( 'render_orders_bulk' );
		}
	}

	/**
	 * 娓叉煋灞炴€х鐞嗛〉闈?	 *
	 * @since 0.2.0
	 */
	public function render_attributes() {
		if ( class_exists( 'Tanzanite_Attributes_Admin' ) ) {
			Tanzanite_Attributes_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_attributes' );
		}
	}

	/**
	 * 娓叉煋璇勮绠＄悊椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_reviews() {
		if ( class_exists( 'Tanzanite_Reviews_Admin' ) ) {
			Tanzanite_Reviews_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_reviews' );
		}
	}

	/**
	 * 娓叉煋鏀粯鏂瑰紡椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_payment_method() {
		if ( class_exists( 'Tanzanite_Payment_Admin' ) ) {
			Tanzanite_Payment_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_payment_method' );
		}
	}

	/**
	 * 娓叉煋绋庣巼绠＄悊椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_tax_rates() {
		if ( class_exists( 'Tanzanite_Tax_Rates_Admin' ) ) {
			Tanzanite_Tax_Rates_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_tax_rates' );
		}
	}

	/**
	 * 娓叉煋杩愯垂妯℃澘椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_shipping_templates() {
		if ( class_exists( 'Tanzanite_Shipping_Admin' ) ) {
			Tanzanite_Shipping_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_shipping_templates' );
		}
	}

	/**
	 * 渲染包装规则管理页面
	 *
	 * @since 0.3.0
	 */
	public function render_packaging_rules() {
		if ( class_exists( 'Tanzanite_Packaging_Admin' ) ) {
			Tanzanite_Packaging_Admin::render_page();
		}
	}

	/**
	 * 娓叉煋鐗╂祦鍟嗙鐞嗛〉闈?	 *
	 * @since 0.2.0
	 */
	public function render_carriers() {
		if ( class_exists( 'Tanzanite_Carriers_Admin' ) ) {
			Tanzanite_Carriers_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_carriers' );
		}
	}

	/**
	 * 娓叉煋浼氬憳妗ｆ椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_member_profiles() {
		if ( class_exists( 'Tanzanite_Members_Admin' ) ) {
			Tanzanite_Members_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_member_profiles' );
		}
	}

	/**
	 * 娓叉煋绀煎搧鍗″拰浼樻儬鍒搁〉闈?	 *
	 * @since 0.2.0
	 */
	public function render_rewards() {
		if ( class_exists( 'Tanzanite_Rewards_Admin' ) ) {
			Tanzanite_Rewards_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_rewards' );
		}
	}

	/**
	 * 娓叉煋绉垎璁剧疆椤甸潰
	 *
	 * @since 0.2.0
	 */
	public function render_loyalty_settings() {
		if ( class_exists( 'Tanzanite_Loyalty_Admin' ) ) {
			Tanzanite_Loyalty_Admin::render_page();
		} else {
			$this->call_legacy_method( 'render_loyalty_settings' );
		}
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
		if ( class_exists( 'Tanzanite_Markdown_Templates_Admin' ) ) {
			Tanzanite_Markdown_Templates_Admin::render_page();
		} elseif ( $this->legacy_plugin && method_exists( $this->legacy_plugin, 'render_markdown_templates_page' ) ) {
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
	 * 渲染 Spoke Geometry 管理页面
	 *
	 * @since 0.2.0
	 */
	public function render_spoke_geometry_page() {
		if ( class_exists( 'Tanzanite_Spoke_Geometry_Admin' ) ) {
			Tanzanite_Spoke_Geometry_Admin::render_page();
		} else {
			echo '<div class="wrap"><h1>错误</h1><p>Tanzanite_Spoke_Geometry_Admin 类未找到</p></div>';
		}
	}

	/**
	 * 渲染 Spoke Length History 页面
	 *
	 * @since 0.2.1
	 */
	public function render_spoke_history_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		global $wpdb;
		$table = $wpdb->prefix . 'tanz_spoke_history';

		echo '<div class="wrap tz-settings-wrapper">';

		// 统一使用后台通用的标题区域样式
		echo '<div class="tz-settings-header">';
		echo '<h1>' . esc_html__( 'Spoke Length History / 辐条长度历史', 'tanzanite-settings' ) . '</h1>';
		echo '<p>' . esc_html__( 'Import and review historical spoke length records for the spoke calculator.', 'tanzanite-settings' ) . '</p>';
		echo '</div>';

		// 导入结果提示
		if ( isset( $_GET['import'] ) && 'success' === $_GET['import'] ) {
			$inserted = isset( $_GET['inserted'] ) ? (int) $_GET['inserted'] : 0;
			$updated  = isset( $_GET['updated'] ) ? (int) $_GET['updated'] : 0;
			$skipped  = isset( $_GET['skipped'] ) ? (int) $_GET['skipped'] : 0;

			echo '<div class="notice notice-success is-dismissible"><p>';
			printf(
				esc_html__( 'Import completed: %1$d inserted, %2$d updated, %3$d skipped.', 'tanzanite-settings' ),
				$inserted,
				$updated,
				$skipped
			);
			echo '</p></div>';
		}

		$action_url = admin_url( 'admin-post.php' );

		// 导入区域 - 使用统一的 tz-settings-section 卡片布局
		echo '<div class="tz-settings-section">';
		echo '<h2>' . esc_html__( 'Import from CSV', 'tanzanite-settings' ) . '</h2>';
		echo '<form method="post" action="' . esc_url( $action_url ) . '" enctype="multipart/form-data">';
		wp_nonce_field( 'tanz_spoke_history_import', 'tanz_spoke_history_import_nonce' );
		echo '<input type="hidden" name="action" value="tanz_spoke_history_import" />';
		echo '<p>';
		echo '<input type="file" name="spoke_history_csv" accept=".csv" required /> ';
		submit_button( __( 'Import', 'tanzanite-settings' ), 'primary', 'submit', false );
		echo '</p>';
		echo '<p class="description">' . esc_html__( 'CSV header should include columns such as hub_model, spoke_count, lacing_pattern, rim_model, erd_mm, etc.', 'tanzanite-settings' ) . '</p>';
		echo '</form>';
		echo '</div>';

		// 最近记录列表，同样包裹在 section 内
		echo '<div class="tz-settings-section">';
		echo '<h2>' . esc_html__( 'Recent Records', 'tanzanite-settings' ) . '</h2>';

		$rows = $wpdb->get_results(
			"SELECT id, hub_brand, hub_model, rim_brand, rim_model, left_length_mm, right_length_mm, wheel_type, source_type, spoke_count, lacing_pattern, created_at FROM {$table} ORDER BY created_at DESC LIMIT 50",
			ARRAY_A
		);

		if ( empty( $rows ) ) {
			echo '<p>' . esc_html__( 'No records found.', 'tanzanite-settings' ) . '</p>';
		} else {
			echo '<table class="widefat fixed striped">';
			echo '<thead><tr>';
			echo '<th>' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
			echo '<th>' . esc_html__( 'Hub', 'tanzanite-settings' ) . '</th>';
			echo '<th>' . esc_html__( 'Rim', 'tanzanite-settings' ) . '</th>';
			echo '<th>' . esc_html__( 'Lengths (L/R)', 'tanzanite-settings' ) . '</th>';
			echo '<th>' . esc_html__( 'Spokes / Pattern', 'tanzanite-settings' ) . '</th>';
			echo '<th>' . esc_html__( 'Wheel / Source', 'tanzanite-settings' ) . '</th>';
			echo '<th>' . esc_html__( 'Created At', 'tanzanite-settings' ) . '</th>';
			echo '</tr></thead><tbody>';

			foreach ( $rows as $row ) {
				$hub = trim( ( $row['hub_brand'] ?? '' ) . ' ' . ( $row['hub_model'] ?? '' ) );
				$rim = trim( ( $row['rim_brand'] ?? '' ) . ' ' . ( $row['rim_model'] ?? '' ) );
				$len_l = isset( $row['left_length_mm'] ) ? (float) $row['left_length_mm'] : null;
				$len_r = isset( $row['right_length_mm'] ) ? (float) $row['right_length_mm'] : null;
				$len_display = '';
				if ( null !== $len_l || null !== $len_r ) {
					$len_display = ( null !== $len_l ? $len_l : '-' ) . ' / ' . ( null !== $len_r ? $len_r : '-' );
				}

				echo '<tr>';
				echo '<td>' . (int) $row['id'] . '</td>';
				echo '<td>' . esc_html( $hub ? $hub : '-' ) . '</td>';
				echo '<td>' . esc_html( $rim ? $rim : '-' ) . '</td>';
				echo '<td>' . esc_html( $len_display ? $len_display : '-' ) . '</td>';
				$spoke_info = '';
				if ( ! empty( $row['spoke_count'] ) ) {
					$spoke_info = (int) $row['spoke_count'] . ' spokes';
				}
				if ( ! empty( $row['lacing_pattern'] ) ) {
					$spoke_info .= $spoke_info ? ' · ' : '';
					$spoke_info .= $row['lacing_pattern'];
				}
				echo '<td>' . esc_html( $spoke_info ? $spoke_info : '-' ) . '</td>';
				$wheel_source = '';
				if ( ! empty( $row['wheel_type'] ) ) {
					$wheel_source = $row['wheel_type'];
				}
				if ( ! empty( $row['source_type'] ) ) {
					$wheel_source .= $wheel_source ? ' · ' : '';
					$wheel_source .= $row['source_type'];
				}
				echo '<td>' . esc_html( $wheel_source ? $wheel_source : '-' ) . '</td>';
				echo '<td>' . esc_html( $row['created_at'] ) . '</td>';
				echo '</tr>';
			}

			echo '</tbody></table>';
		}

		echo '</div>';
	}

	/**
	 * 处理 Spoke Length History CSV 导入
	 *
	 * @since 0.2.1
	 */
	public function handle_spoke_history_import() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限执行此操作。', 'tanzanite-settings' ) );
		}

		if ( ! isset( $_POST['tanz_spoke_history_import_nonce'] ) || ! wp_verify_nonce( $_POST['tanz_spoke_history_import_nonce'], 'tanz_spoke_history_import' ) ) {
			wp_die( __( '安全验证失败。', 'tanzanite-settings' ) );
		}

		if ( empty( $_FILES['spoke_history_csv']['tmp_name'] ) || ! is_uploaded_file( $_FILES['spoke_history_csv']['tmp_name'] ) ) {
			wp_die( __( '未上传有效的 CSV 文件。', 'tanzanite-settings' ) );
		}

		$file   = $_FILES['spoke_history_csv']['tmp_name'];
		$handle = fopen( $file, 'r' );
		if ( false === $handle ) {
			wp_die( __( '无法读取上传的文件。', 'tanzanite-settings' ) );
		}

		$header = fgetcsv( $handle );
		if ( ! $header || ! is_array( $header ) ) {
			fclose( $handle );
			wp_die( __( 'CSV 头部无效。', 'tanzanite-settings' ) );
		}

		$columns = array();
		foreach ( $header as $index => $name ) {
			$key = sanitize_key( $name );
			if ( $key ) {
				$columns[ $index ] = $key;
			}
		}

		if ( empty( $columns ) ) {
			fclose( $handle );
			wp_die( __( 'CSV 中未检测到有效字段。', 'tanzanite-settings' ) );
		}

		global $wpdb;
		$table = $wpdb->prefix . 'tanz_spoke_history';

		$inserted = 0;
		$updated  = 0;
		$skipped  = 0;

		$to_string = static function( $value ) {
			$value = trim( (string) $value );
			return '' === $value ? null : $value;
		};

		$to_float = static function( $value ) {
			$value = trim( (string) $value );
			return '' === $value ? null : (float) $value;
		};

		$to_int = static function( $value ) {
			$value = trim( (string) $value );
			return '' === $value ? null : (int) $value;
		};

		while ( ( $row = fgetcsv( $handle ) ) !== false ) {
			// 跳过空行
			if ( empty( array_filter( $row, 'strlen' ) ) ) {
				continue;
			}

			$raw = array();
			foreach ( $columns as $index => $key ) {
				$raw[ $key ] = isset( $row[ $index ] ) ? $row[ $index ] : '';
			}

			$data = array(
				'wheel_type'                => $to_string( $raw['wheel_type'] ?? '' ),
				'source_type'               => $to_string( $raw['source_type'] ?? '' ),
				'rim_brand'                 => $to_string( $raw['rim_brand'] ?? '' ),
				'rim_model'                 => $to_string( $raw['rim_model'] ?? '' ),
				'hub_brand'                 => $to_string( $raw['hub_brand'] ?? '' ),
				'hub_model'                 => $to_string( $raw['hub_model'] ?? '' ),
				'erd_mm'                    => $to_float( $raw['erd_mm'] ?? '' ),
				'left_flange_pcd_mm'        => $to_float( $raw['left_flange_pcd_mm'] ?? '' ),
				'right_flange_pcd_mm'       => $to_float( $raw['right_flange_pcd_mm'] ?? '' ),
				'left_flange_to_center_mm'  => $to_float( $raw['left_flange_to_center_mm'] ?? '' ),
				'right_flange_to_center_mm' => $to_float( $raw['right_flange_to_center_mm'] ?? '' ),
				'spoke_count'               => $to_int( $raw['spoke_count'] ?? '' ),
				'lacing_pattern'            => $to_string( $raw['lacing_pattern'] ?? '' ),
				'nipple_type'               => $to_string( $raw['nipple_type'] ?? '' ),
				'left_length_mm'            => $to_float( $raw['left_length_mm'] ?? '' ),
				'right_length_mm'           => $to_float( $raw['right_length_mm'] ?? '' ),
			);

			$hub_model      = $data['hub_model'];
			$spoke_count    = $data['spoke_count'];
			$lacing_pattern = $data['lacing_pattern'];

			$existing_id = null;
			// 按 hub_model + spoke_count + lacing_pattern 去重 / 覆盖
			if ( $hub_model && $spoke_count && $lacing_pattern ) {
				$existing_id = $wpdb->get_var(
					$wpdb->prepare(
						"SELECT id FROM {$table} WHERE hub_model = %s AND spoke_count = %d AND lacing_pattern = %s LIMIT 1",
						$hub_model,
						$spoke_count,
						$lacing_pattern
					)
				);
			}

			if ( $existing_id ) {
				// 更新已有记录（保留 created_at）
				$result = $wpdb->update(
					$table,
					$data,
					array( 'id' => (int) $existing_id ),
					null,
					array( '%d' )

				);

				if ( false === $result ) {
					$skipped++;
				} else {
					$updated++;
				}
			} else {
				// 鏂板璁板綍
				$data['created_at'] = current_time( 'mysql' );
				$result            = $wpdb->insert( $table, $data );

				if ( false === $result ) {
					$skipped++;
				} else {
					$inserted++;
				}
			}
		}

		fclose( $handle );

		$redirect_url = add_query_arg(
			array(
				'page'     => 'tanzanite-spoke-history',
				'import'   => 'success',
				'inserted' => $inserted,
				'updated'  => $updated,
				'skipped'  => $skipped,
			),
			admin_url( 'admin.php' )
		);

		wp_safe_redirect( $redirect_url );
		exit;
	}


	/**
	 * 楠岃瘉鍜屾竻鐞嗛樁姊畾浠锋暟鎹?	 *
	 * @since 0.2.0
	 * @param mixed $value 杈撳叆鏁版嵁
	 * @param bool $from_request 鏄惁鏉ヨ嚜 API 璇锋眰锛堝鏋滄槸锛岃繑鍥?WP_Error锛?	 * @return array|WP_Error
	 */
	public static function sanitize_tier_pricing( $value, bool $from_request = false ) {
		if ( empty( $value ) || ! is_array( $value ) ) {
			return [];
		}

		$sanitized = [];

		foreach ( $value as $item ) {
			if ( ! is_array( $item ) ) {
				if ( is_object( $item ) ) {
					$item = json_decode( wp_json_encode( $item ), true );
				} else {
					continue;
				}
			}

			$min_qty = isset( $item['min_qty'] ) ? (int) $item['min_qty'] : ( isset( $item['minQty'] ) ? (int) $item['minQty'] : ( isset( $item['min'] ) ? (int) $item['min'] : 0 ) );
			$max_raw = $item['max_qty'] ?? ( $item['maxQty'] ?? ( $item['max'] ?? null ) );
			$max_qty = ( '' === $max_raw || null === $max_raw ) ? null : (int) $max_raw;
			$price_raw = $item['price'] ?? ( $item['amount'] ?? ( $item['value'] ?? null ) );
			$price = is_numeric( $price_raw ) ? (float) $price_raw : null;
			$note_raw = $item['note'] ?? ( $item['label'] ?? ( $item['desc'] ?? '' ) );
			$note = $note_raw ? sanitize_text_field( (string) $note_raw ) : '';

			if ( $min_qty <= 0 || null === $price || $price < 0 ) {
				if ( $from_request ) {
					return new \WP_Error( 'invalid_tier_qty', __( '请填写有效的最小数量和单价。', 'tanzanite-settings' ) );
				}
				continue;
			}

			if ( null !== $max_qty && $max_qty < $min_qty ) {
				if ( $from_request ) {
					return new \WP_Error( 'invalid_tier_range', __( '最大数量必须大于或等于最小数量。', 'tanzanite-settings' ) );
				}
				continue;
			}

			$sanitized[] = [
				'min_qty' => $min_qty,
				'max_qty' => $max_qty,
				'price'   => (float) number_format( $price, 2, '.', '' ),
				'note'    => $note,
			];
		}

		if ( empty( $sanitized ) ) {
			return [];
		}

		usort(
			$sanitized,
			static function ( $a, $b ) {
				return $a['min_qty'] <=> $b['min_qty'];
			}
		);

		$previous_max = null;
		$previous_min = null;

		foreach ( $sanitized as $index => $row ) {
			$min = (int) $row['min_qty'];
			$max = $row['max_qty'];

			if ( 0 === $index ) {
				$previous_max = $max;
				$previous_min = $min;
				continue;
			}

			if ( null === $previous_max ) {
				$error = __( 'Only the last tier can have no maximum quantity.', 'tanzanite-settings' );

				return $from_request ? new \WP_Error( 'invalid_tier_limit', $error ) : [];
			}

			if ( $min <= $previous_max || $min <= $previous_min ) {
				$error = __( 'Tier ranges overlap or are out of order. Please check.', 'tanzanite-settings' );

				return $from_request ? new \WP_Error( 'invalid_tier_overlap', $error ) : [];
			}

			$previous_max = $max;
			$previous_min = $min;
		}

		return $sanitized;
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

	/**
	 * 渲染 Suggestion Feedback 后台页面
	 *
	 * @since 0.2.0
	 */
	public function render_suggestion_feedback() {
		if ( class_exists( 'Tanzanite_Suggestion_Feedback_Admin' ) ) {
			Tanzanite_Suggestion_Feedback_Admin::render_page();
		} else {
			wp_die( esc_html__( 'Suggestion feedback admin class not found.', 'tanzanite-settings' ) );
		}
	}

}


if ( ! function_exists( 'tanzanite_settings_sanitize_tier_pricing' ) ) {
    function tanzanite_settings_sanitize_tier_pricing( $value, bool $from_request = false ) {
        return Tanzanite_Plugin::sanitize_tier_pricing( $value, $from_request );
    }
}
