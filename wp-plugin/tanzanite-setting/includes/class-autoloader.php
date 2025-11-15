<?php
/**
 * Tanzanite Settings Autoloader
 *
 * 自动加载类文件，将类名转换为文件路径
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 自动加载器类
 *
 * 负责自动加载所有 Tanzanite_ 开头的类
 */
class Tanzanite_Autoloader {

	/**
	 * 类文件映射缓存
	 *
	 * @var array
	 */
	private static $class_map = array();

	/**
	 * 注册自动加载器
	 *
	 * @since 0.2.0
	 */
	public static function register() {
		spl_autoload_register( array( __CLASS__, 'autoload' ) );
	}

	/**
	 * 自动加载类文件
	 *
	 * @since 0.2.0
	 * @param string $class 类名
	 */
	public static function autoload( $class ) {
		// 只处理 Tanzanite_ 开头的类
		if ( strpos( $class, 'Tanzanite_' ) !== 0 ) {
			return;
		}

		// 跳过 Tanzanite_Settings_Plugin，它在 legacy-pages.php 中定义
		if ( $class === 'Tanzanite_Settings_Plugin' ) {
			return;
		}

		// 检查缓存
		if ( isset( self::$class_map[ $class ] ) ) {
			require_once self::$class_map[ $class ];
			return;
		}

		// 获取文件路径
		$file_path = self::get_file_path( $class );

		if ( file_exists( $file_path ) ) {
			self::$class_map[ $class ] = $file_path;
			require_once $file_path;
		} elseif ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
			error_log( "Tanzanite Autoloader: Class file not found for {$class} at {$file_path}" );
		}
	}

	/**
	 * 将类名转换为文件路径
	 *
	 * 转换规则:
	 * - Tanzanite_Plugin -> includes/class-plugin.php
	 * - Tanzanite_REST_Products_Controller -> includes/rest-api/class-products-controller.php
	 * - Tanzanite_Admin_Products_Page -> includes/admin/pages/class-products-page.php
	 * - Tanzanite_Product_Model -> includes/models/class-product-model.php
	 *
	 * @since 0.2.0
	 * @param string $class 类名
	 * @return string 文件路径
	 */
	private static function get_file_path( $class ) {
		// 移除 Tanzanite_ 前缀
		$class_name = substr( $class, 10 ); // strlen('Tanzanite_') = 10

		// 转换为小写并用连字符分隔
		$class_name = strtolower( str_replace( '_', '-', $class_name ) );

		// 确定子目录
		$subdirectory = self::get_subdirectory( $class );

		// 构建文件路径
		$file_name = 'class-' . $class_name . '.php';
		$file_path = TANZANITE_PLUGIN_DIR . 'includes/' . $subdirectory . $file_name;

		return $file_path;
	}

	/**
	 * 根据类名确定子目录
	 *
	 * @since 0.2.0
	 * @param string $class 类名
	 * @return string 子目录路径（带尾部斜杠）
	 */
	private static function get_subdirectory( $class ) {
		// REST API 控制器
		if ( strpos( $class, 'Tanzanite_REST_' ) === 0 ) {
			return 'rest-api/';
		}

		// 后台页面
		if ( strpos( $class, 'Tanzanite_Admin_' ) === 0 ) {
			// 检查是否是页面类
			if ( strpos( $class, '_Page' ) !== false ) {
				return 'admin/pages/';
			}
			return 'admin/';
		}

		// 数据模型
		if ( strpos( $class, '_Model' ) !== false ) {
			return 'models/';
		}

		// 服务类
		if ( strpos( $class, '_Service' ) !== false ) {
			return 'services/';
		}

		// 辅助类
		if ( strpos( $class, 'Tanzanite_Helper_' ) === 0 || 
		     strpos( $class, 'Tanzanite_Validator' ) === 0 || 
		     strpos( $class, 'Tanzanite_Sanitizer' ) === 0 || 
		     strpos( $class, 'Tanzanite_Formatter' ) === 0 ) {
			return 'helpers/';
		}

		// 默认在 includes 根目录
		return '';
	}

	/**
	 * 获取已加载的类列表（用于调试）
	 *
	 * @since 0.2.0
	 * @return array
	 */
	public static function get_loaded_classes() {
		return array_keys( self::$class_map );
	}

	/**
	 * 清空类映射缓存（用于测试）
	 *
	 * @since 0.2.0
	 */
	public static function clear_cache() {
		self::$class_map = array();
	}
}
