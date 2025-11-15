<?php
/**
 * Plugin Name:       Tanzanite Settings
 * Plugin URI:        https://tanzanite.site
 * Description:       提供 Tanzanite 前端商城所需的后台设置与数据管理入口（模块化重构版）
 * Version:           0.2.0
 * Author:            Tanzanite Team
 * Author URI:        https://tanzanite.site
 * Requires at least: 6.0
 * Requires PHP:      7.4
 * Text Domain:       tanzanite-settings
 * Domain Path:       /languages
 * License:           GPL v2 or later
 * License URI:       https://www.gnu.org/licenses/gpl-2.0.html
 *
 * @package Tanzanite_Settings
 * @since   0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit; // 防止直接访问
}

// ============================================================================
// 定义插件常量
// ============================================================================

if ( ! defined( 'TANZANITE_PLUGIN_FILE' ) ) {
	define( 'TANZANITE_PLUGIN_FILE', __FILE__ );
}

if ( ! defined( 'TANZANITE_PLUGIN_DIR' ) ) {
	define( 'TANZANITE_PLUGIN_DIR', plugin_dir_path( __FILE__ ) );
}

if ( ! defined( 'TANZANITE_PLUGIN_URL' ) ) {
	define( 'TANZANITE_PLUGIN_URL', plugin_dir_url( __FILE__ ) );
}

if ( ! defined( 'TANZANITE_PLUGIN_BASENAME' ) ) {
	define( 'TANZANITE_PLUGIN_BASENAME', plugin_basename( __FILE__ ) );
}

// ============================================================================
// 加载自动加载器
// ============================================================================

require_once TANZANITE_PLUGIN_DIR . 'includes/class-autoloader.php';

// 注册自动加载器
Tanzanite_Autoloader::register();

// ============================================================================
// 加载 MyTheme SEO 模块
// ============================================================================

require_once TANZANITE_PLUGIN_DIR . 'includes/class-mytheme-seo.php';

// ============================================================================
// 插件初始化
// ============================================================================

/**
 * 初始化插件
 *
 * @since 0.2.0
 */
function tanzanite_settings_init() {
	// 获取插件实例
	$plugin = Tanzanite_Plugin::get_instance();
	
	// 运行插件
	$plugin->run();
	
	// 初始化 MyTheme SEO
	if ( class_exists( 'MyTheme_SEO_Plugin' ) ) {
		error_log('Tanzanite Settings: MyTheme_SEO_Plugin class found, initializing...');
		MyTheme_SEO_Plugin::instance();
		error_log('Tanzanite Settings: MyTheme_SEO_Plugin initialized');
	} else {
		error_log('Tanzanite Settings: MyTheme_SEO_Plugin class NOT found!');
	}
}

// 在 WordPress 加载完成后初始化插件
add_action( 'plugins_loaded', 'tanzanite_settings_init' );

// ============================================================================
// 激活和停用钩子
// ============================================================================

/**
 * 插件激活钩子
 *
 * @since 0.2.0
 */
function tanzanite_settings_activate() {
	// 确保自动加载器已加载
	if ( ! class_exists( 'Tanzanite_Plugin' ) ) {
		require_once TANZANITE_PLUGIN_DIR . 'includes/class-autoloader.php';
		Tanzanite_Autoloader::register();
	}
	
	// 调用激活方法
	Tanzanite_Plugin::activate();
}

/**
 * 插件停用钩子
 *
 * @since 0.2.0
 */
function tanzanite_settings_deactivate() {
	// 调用停用方法
	Tanzanite_Plugin::deactivate();
}

// 注册激活和停用钩子
register_activation_hook( __FILE__, 'tanzanite_settings_activate' );
register_deactivation_hook( __FILE__, 'tanzanite_settings_deactivate' );

// ============================================================================
// 调试信息（仅在 WP_DEBUG 模式下）
// ============================================================================

if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
	add_action( 'admin_notices', function() {
		if ( current_user_can( 'manage_options' ) ) {
			$loaded_classes = Tanzanite_Autoloader::get_loaded_classes();
			if ( ! empty( $loaded_classes ) ) {
				echo '<div class="notice notice-info is-dismissible">';
				echo '<p><strong>Tanzanite Settings v0.2.0 (重构版)</strong></p>';
				echo '<p>已加载 ' . count( $loaded_classes ) . ' 个类文件</p>';
				echo '</div>';
			}
		}
	} );
}

// ============================================================================
// 兼容性检查
// ============================================================================

/**
 * 检查 PHP 和 WordPress 版本
 *
 * @since 0.2.0
 */
function tanzanite_settings_check_requirements() {
	$php_version = '7.4';
	$wp_version  = '6.0';

	if ( version_compare( PHP_VERSION, $php_version, '<' ) ) {
		deactivate_plugins( plugin_basename( __FILE__ ) );
		wp_die(
			sprintf(
				__( 'Tanzanite Settings 需要 PHP %s 或更高版本。当前版本: %s', 'tanzanite-settings' ),
				$php_version,
				PHP_VERSION
			)
		);
	}

	global $wp_version;
	if ( version_compare( $wp_version, $wp_version, '<' ) ) {
		deactivate_plugins( plugin_basename( __FILE__ ) );
		wp_die(
			sprintf(
				__( 'Tanzanite Settings 需要 WordPress %s 或更高版本。当前版本: %s', 'tanzanite-settings' ),
				$wp_version,
				$wp_version
			)
		);
	}
}

// 在插件加载时检查要求
add_action( 'plugins_loaded', 'tanzanite_settings_check_requirements', 1 );

// ============================================================================
// 就这样！主文件只有 ~180 行
// 所有功能都在 includes/ 目录的模块化文件中
// ============================================================================
