<?php
/**
 * Redeem REST API Controller
 *
 * 处理积分兑换礼品卡相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 积分兑换 REST API 控制器
 *
 * 提供积分兑换礼品卡的功能
 * 
 * 前端 Nuxt 对接说明：
 * - 所有接口都需要用户登录
 * - 请求头需要包含 X-WP-Nonce
 * - 兑换接口会自动扣除用户积分并创建交易记录
 */
class Tanzanite_REST_Redeem_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'redeem';

	/**
	 * 礼品卡表名
	 *
	 * @var string
	 */
	private $giftcards_table;

	/**
	 * 积分交易表名
	 *
	 * @var string
	 */
	private $rewards_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->giftcards_table = $wpdb->prefix . 'tanz_giftcards';
		$this->rewards_table   = $wpdb->prefix . 'tanz_rewards_transactions';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 获取兑换配置
		// 前端 Nuxt 使用：GET /wp-json/tanzanite/v1/redeem/config
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/config',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( $this, 'get_config' ),
				'permission_callback' => '__return_true', // 公开接口
			)
		);

		// 查询用户积分余额（公开接口，未登录返回 0）
		// 前端 Nuxt 使用：GET /wp-json/tanzanite/v1/loyalty/points
		// Headers: X-WP-Nonce: {nonce} (可选)
		register_rest_route(
			$this->namespace,
			'/loyalty/points',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( $this, 'get_user_points' ),
				'permission_callback' => '__return_true',
			)
		);
	}

	/**
	 * 获取兑换配置
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_config( $request ) {
		$preset_values_str = get_option( 'tz_redeem_preset_values', '10,50,100,200,500' );
		$preset_values     = array_map( 'floatval', explode( ',', $preset_values_str ) );

		return $this->respond_success(
			array(
				'enabled'            => '1' === get_option( 'tz_redeem_enabled', '1' ),
				'exchange_rate'      => absint( get_option( 'tz_redeem_exchange_rate', '100' ) ),
				'min_points'         => absint( get_option( 'tz_redeem_min_points', '1000' ) ),
				'max_value_per_day'  => floatval( get_option( 'tz_redeem_max_value_per_day', '500' ) ),
				'card_expiry_days'   => absint( get_option( 'tz_redeem_card_expiry_days', '365' ) ),
				'preset_values'      => $preset_values,
			)
		);
	}

	/**
	 * 获取用户积分余额
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_user_points( $request ) {
		$user_id = get_current_user_id();
		
		// 未登录时返回默认值
		if ( ! $user_id ) {
			return $this->respond_success(
				array(
					'user_id'              => 0,
					'points'               => 0,
					'can_redeem'           => false,
					'max_redeemable_value' => 0,
					'today_redeemed'       => 0,
				)
			);
		}
		
		$points  = absint( get_user_meta( $user_id, 'loyalty_points', true ) );

		// 获取配置
		$exchange_rate     = absint( get_option( 'tz_redeem_exchange_rate', '100' ) );
		$min_points        = absint( get_option( 'tz_redeem_min_points', '1000' ) );
		$max_value_per_day = floatval( get_option( 'tz_redeem_max_value_per_day', '500' ) );

		// 计算今日已兑换金额
		global $wpdb;
		$today_start = gmdate( 'Y-m-d 00:00:00' );
		$today_end   = gmdate( 'Y-m-d 23:59:59' );

		$today_redeemed = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT SUM(amount_delta) FROM {$this->rewards_table} 
				WHERE user_id = %d 
				AND action = 'redeem' 
				AND related_type = 'giftcard'
				AND created_at BETWEEN %s AND %s",
				$user_id,
				$today_start,
				$today_end
			)
		);

		$today_redeemed = floatval( $today_redeemed );

		// 计算可兑换的最大金额
		$max_redeemable_by_points = floor( $points / $exchange_rate );
		$max_redeemable_by_limit  = $max_value_per_day > 0 ? $max_value_per_day - $today_redeemed : PHP_INT_MAX;
		$max_redeemable_value     = min( $max_redeemable_by_points, $max_redeemable_by_limit );

		return $this->respond_success(
			array(
				'user_id'              => $user_id,
				'points'               => $points,
				'can_redeem'           => $points >= $min_points && $max_redeemable_value > 0,
				'max_redeemable_value' => max( 0, $max_redeemable_value ),
				'today_redeemed'       => $today_redeemed,
			)
		);
	}
}
