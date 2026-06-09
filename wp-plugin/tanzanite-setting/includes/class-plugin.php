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
		if ( class_exists( 'Tanzanite_Tube_Specs_Database' ) ) {
			Tanzanite_Tube_Specs_Database::maybe_upgrade();
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

		// 鍔犺浇 Rewards Admin - [迁移至 Go 后端，停止加载]
		/* $rewards_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-rewards-admin.php';
		if ( file_exists( $rewards_admin ) ) {
			require_once $rewards_admin;
		} */

		// 鍔犺浇 Loyalty Admin - [迁移至 Go 后端，停止加载]
		/* $loyalty_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-loyalty-admin.php';
		if ( file_exists( $loyalty_admin ) ) {
			require_once $loyalty_admin;
		} */

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

		// Tube specs 数据库与后台页面
		$tube_specs_db = TANZANITE_PLUGIN_DIR . 'includes/database/class-tube-specs-database.php';
		if ( file_exists( $tube_specs_db ) ) {
			require_once $tube_specs_db;
		}

		$tube_specs_admin = TANZANITE_PLUGIN_DIR . 'includes/admin/class-tube-specs-admin.php';
		if ( file_exists( $tube_specs_admin ) ) {
			require_once $tube_specs_admin;
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



		add_action( 'admin_post_tanz_save_tracking_settings', array( $this, 'handle_save_tracking_settings' ) );
		add_action( 'admin_post_tanz_test_tracking', array( $this, 'handle_test_tracking' ) );

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

	public function handle_save_tracking_settings() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限执行此操作。', 'tanzanite-settings' ) );
		}

		$nonce = isset( $_POST['_wpnonce'] ) ? sanitize_text_field( wp_unslash( $_POST['_wpnonce'] ) ) : '';
		if ( ! $nonce || ! wp_verify_nonce( $nonce, 'tanz_tracking_settings' ) ) {
			wp_die( __( '安全验证失败。', 'tanzanite-settings' ) );
		}

		$provider = isset( $_POST['provider'] ) ? sanitize_key( wp_unslash( $_POST['provider'] ) ) : '17track';
		$settings = isset( $_POST['settings'] ) && is_array( $_POST['settings'] ) ? (array) wp_unslash( $_POST['settings'] ) : array();

		$providers = array( '17track' );
		if ( class_exists( 'Tanzanite_Carriers_Admin' ) && defined( 'Tanzanite_Carriers_Admin::TRACKING_PROVIDERS' ) ) {
			$providers = array_keys( Tanzanite_Carriers_Admin::TRACKING_PROVIDERS );
		}

		if ( ! in_array( $provider, $providers, true ) ) {
			$provider = $providers[0] ?? '17track';
		}

		$allowed_fields = array();
		if ( class_exists( 'Tanzanite_Carriers_Admin' ) && defined( 'Tanzanite_Carriers_Admin::TRACKING_PROVIDERS' ) ) {
			$allowed_fields = Tanzanite_Carriers_Admin::TRACKING_PROVIDERS[ $provider ]['fields'] ?? array();
		}

		$sanitized_settings = array();
		if ( is_array( $allowed_fields ) && ! empty( $allowed_fields ) ) {
			foreach ( $allowed_fields as $field_key => $field_config ) {
				if ( ! array_key_exists( $field_key, $settings ) ) {
					continue;
				}

				$value = is_scalar( $settings[ $field_key ] ) ? (string) $settings[ $field_key ] : '';
				$value = trim( $value );
				if ( '' === $value ) {
					continue;
				}

				if ( 'endpoint' === $field_key ) {
					$sanitized_settings[ $field_key ] = esc_url_raw( $value );
				} else {
					$sanitized_settings[ $field_key ] = sanitize_text_field( $value );
				}
			}
		} else {
			foreach ( $settings as $field_key => $value ) {
				if ( ! is_scalar( $value ) ) {
					continue;
				}
				$sanitized_settings[ sanitize_key( $field_key ) ] = sanitize_text_field( (string) $value );
			}
		}

		update_option(
			'tanzanite_tracking_settings',
			array(
				'provider'  => $provider,
				'settings'  => array(
					$provider => $sanitized_settings,
				),
				'updated_at' => current_time( 'mysql', true ),
			),
			false
		);

		wp_safe_redirect( admin_url( 'admin.php?page=tanzanite-settings-carriers&tab=config&updated=1' ) );
		exit;
	}

	public function handle_test_tracking() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_send_json_error( __( '无权限执行此操作。', 'tanzanite-settings' ), 403 );
		}

		$nonce = isset( $_POST['_wpnonce'] ) ? sanitize_text_field( wp_unslash( $_POST['_wpnonce'] ) ) : '';
		if ( ! $nonce || ! wp_verify_nonce( $nonce, 'tanz_tracking_settings' ) ) {
			wp_send_json_error( __( '安全验证失败。', 'tanzanite-settings' ), 403 );
		}

		$provider = isset( $_POST['provider'] ) ? sanitize_key( wp_unslash( $_POST['provider'] ) ) : '17track';
		$settings = isset( $_POST['settings'] ) && is_array( $_POST['settings'] ) ? (array) wp_unslash( $_POST['settings'] ) : array();

		$providers = array( '17track' );
		if ( class_exists( 'Tanzanite_Carriers_Admin' ) && defined( 'Tanzanite_Carriers_Admin::TRACKING_PROVIDERS' ) ) {
			$providers = array_keys( Tanzanite_Carriers_Admin::TRACKING_PROVIDERS );
		}

		if ( ! in_array( $provider, $providers, true ) ) {
			wp_send_json_error( __( '追踪服务商无效。', 'tanzanite-settings' ), 400 );
		}

		$required = array();
		if ( class_exists( 'Tanzanite_Carriers_Admin' ) && defined( 'Tanzanite_Carriers_Admin::TRACKING_PROVIDERS' ) ) {
			$fields = Tanzanite_Carriers_Admin::TRACKING_PROVIDERS[ $provider ]['fields'] ?? array();
			foreach ( $fields as $field_key => $field_config ) {
				$required[] = $field_key;
			}
		}

		$missing = array();
		foreach ( $required as $field_key ) {
			$value = isset( $settings[ $field_key ] ) && is_scalar( $settings[ $field_key ] ) ? trim( (string) $settings[ $field_key ] ) : '';
			if ( '' === $value ) {
				$missing[] = $field_key;
			}
		}

		if ( ! empty( $missing ) ) {
			wp_send_json_error(
				sprintf(
					__( '缺少配置字段：%s', 'tanzanite-settings' ),
					implode( ', ', array_map( 'sanitize_key', $missing ) )
				),
				400
			);
		}

		wp_send_json_success(
			array(
				'provider' => $provider,
				'message'  => __( '配置已通过校验（未执行外部网络请求）。', 'tanzanite-settings' ),
			)
		);
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
			'Tanzanite_REST_Auth_Controller',
			'Tanzanite_REST_Orders_Controller',
			'Tanzanite_REST_Products_Controller',
			'Tanzanite_REST_Payments_Controller',
			'Tanzanite_REST_TaxRates_Controller',
			'Tanzanite_REST_Reviews_Controller',
			'Tanzanite_REST_Members_Controller',
			'Tanzanite_REST_Carriers_Controller',
			// 'Tanzanite_REST_Coupons_Controller',   // 迁移至 Go 后端
			// 'Tanzanite_REST_Giftcards_Controller', // 迁移至 Go 后端
			// 'Tanzanite_REST_Redeem_Controller',    // 迁移至 Go 后端
			// 'Tanzanite_REST_Loyalty_Controller',   // 迁移至 Go 后端
			'Tanzanite_REST_Attributes_Controller',
			// 'Tanzanite_REST_Audit_Controller',
			'Tanzanite_REST_ShippingTemplates_Controller',
			'Tanzanite_REST_Packaging_Controller',
			'Tanzanite_REST_User_Assets_Controller',
			'Tanzanite_REST_Wishlist_Controller',
			// 新增：用户反馈 / 留言控制器
			'Tanzanite_REST_Feedback_Controller',
			'Tanzanite_REST_Suggestion_Feedback_Controller',
			'Tanzanite_REST_SpokeExport_Controller', // New Export Controller

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

		// Rewards - [迁移至 Go 后端，隐藏菜单]
		/* add_submenu_page(
			$root_slug,
			__( 'Gift Cards & Coupons', 'tanzanite-settings' ),
			__( 'Gift Cards & Coupons', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-rewards',
			array( $this, 'render_rewards' )
		); */

		// Loyalty - [迁移至 Go 后端，隐藏菜单]
		/* add_submenu_page(
			$root_slug,
			__( 'Loyalty Settings', 'tanzanite-settings' ),
			__( 'Loyalty & Points', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-loyalty',
			array( $this, 'render_loyalty_settings' )
		); */

		// Cart
		add_submenu_page(
			$root_slug,
			__( 'Cart Management', 'tanzanite-settings' ),
			__( 'Cart & Orders', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-cart-list',
			array( $this, 'render_cart_list' )
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

		// Tube Specs
		add_submenu_page(
			$root_slug,
			__( 'Tube Specs', 'tanzanite-settings' ),
			__( 'Tube Specs', 'tanzanite-settings' ),
			$root_capability,
			'tanzanite-settings-tube-specs',
			array( $this, 'render_tube_specs' )
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
	 * 渲染 Tube Specs 后台页面
	 *
	 * @since 0.3.0
	 */
	public function render_tube_specs() {
		if ( class_exists( 'Tanzanite_Tube_Specs_Admin' ) ) {
			Tanzanite_Tube_Specs_Admin::render_page();
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
