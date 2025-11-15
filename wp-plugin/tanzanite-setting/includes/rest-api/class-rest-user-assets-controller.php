<?php
/**
 * User Assets REST API Controller
 *
 * 处理用户资产相关的 REST API 请求
 * - 获取用户优惠券数量
 * - 获取用户积分卡数量
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.3.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 用户资产 REST API 控制器
 */
class Tanzanite_REST_User_Assets_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'user/assets';

	/**
	 * 构造函数
	 *
	 * @since 0.3.0
	 */
	public function __construct() {
		parent::__construct();
	}

	/**
	 * 注册路由
	 *
	 * @since 0.3.0
	 */
	public function register_routes() {
		// 获取用户资产（优惠券和积分卡）
		// 前端 Nuxt 使用：GET /wp-json/mytheme/v1/user/assets
		// Headers: Cookie (自动携带)
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( $this, 'get_user_assets' ),
				'permission_callback' => 'is_user_logged_in',
			)
		);
	}

	/**
	 * 获取用户资产
	 *
	 * @param WP_REST_Request $request 请求对象
	 * @return WP_REST_Response 响应对象
	 * @since 0.3.0
	 */
	public function get_user_assets( $request ) {
		$user_id = get_current_user_id();

		if ( ! $user_id ) {
			return new WP_REST_Response(
				array(
					'success' => false,
					'message' => __( 'User not logged in', 'tanzanite' ),
				),
				401
			);
		}

		try {
			global $wpdb;

			// 获取优惠券数量
			$coupons_count = $this->get_user_coupons_count( $user_id );

			// 获取积分卡数量
			$point_cards_count = $this->get_user_point_cards_count( $user_id );

			return new WP_REST_Response(
				array(
					'success' => true,
					'data'    => array(
						'coupons'     => $coupons_count,
						'point_cards' => $point_cards_count,
					),
				),
				200
			);

		} catch ( Exception $e ) {
			return new WP_REST_Response(
				array(
					'success' => false,
					'message' => $e->getMessage(),
				),
				500
			);
		}
	}

	/**
	 * 获取用户优惠券数量
	 *
	 * @param int $user_id 用户ID
	 * @return int 优惠券数量
	 * @since 0.3.0
	 */
	private function get_user_coupons_count( $user_id ) {
		global $wpdb;

		// 查询用户拥有的优惠券数量
		// 假设优惠券存储在 user_meta 或专门的优惠券表中
		$coupons_table = $wpdb->prefix . 'tanz_user_coupons';

		// 检查表是否存在
		$table_exists = $wpdb->get_var( "SHOW TABLES LIKE '{$coupons_table}'" ) === $coupons_table;

		if ( ! $table_exists ) {
			// 如果表不存在，返回 0 或从 user_meta 获取
			$coupons = get_user_meta( $user_id, 'tanz_coupons_count', true );
			return $coupons ? intval( $coupons ) : 0;
		}

		// 从优惠券表查询有效的优惠券数量
		$count = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$coupons_table} 
				WHERE user_id = %d 
				AND status = 'active' 
				AND (expiry_date IS NULL OR expiry_date > NOW())",
				$user_id
			)
		);

		return $count ? intval( $count ) : 0;
	}

	/**
	 * 获取用户积分卡数量
	 *
	 * @param int $user_id 用户ID
	 * @return int 积分卡数量
	 * @since 0.3.0
	 */
	private function get_user_point_cards_count( $user_id ) {
		global $wpdb;

		// 查询用户拥有的积分卡数量
		// 假设积分卡存储在专门的积分卡表中
		$point_cards_table = $wpdb->prefix . 'tanz_user_point_cards';

		// 检查表是否存在
		$table_exists = $wpdb->get_var( "SHOW TABLES LIKE '{$point_cards_table}'" ) === $point_cards_table;

		if ( ! $table_exists ) {
			// 如果表不存在，返回 0 或从 user_meta 获取
			$point_cards = get_user_meta( $user_id, 'tanz_point_cards_count', true );
			return $point_cards ? intval( $point_cards ) : 0;
		}

		// 从积分卡表查询有效的积分卡数量
		$count = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM {$point_cards_table} 
				WHERE user_id = %d 
				AND status = 'active' 
				AND (expiry_date IS NULL OR expiry_date > NOW())",
				$user_id
			)
		);

		return $count ? intval( $count ) : 0;
	}
}
