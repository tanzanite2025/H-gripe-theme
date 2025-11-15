<?php
/**
 * Loyalty REST API Controller
 *
 * 处理积分获取相关的 REST API 请求
 * - 每日签到
 * - 推荐奖励
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 积分获取 REST API 控制器
 */
class Tanzanite_REST_Loyalty_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'loyalty';

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
		$this->rewards_table = $wpdb->prefix . 'tanz_rewards_transactions';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 每日签到
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/loyalty/checkin
		// Headers: X-WP-Nonce: {nonce}
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/checkin',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'checkin' ),
				'permission_callback' => 'is_user_logged_in',
			)
		);

		// 生成推荐码
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/loyalty/referral/generate
		// Headers: X-WP-Nonce: {nonce}
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/referral/generate',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'generate_referral' ),
				'permission_callback' => 'is_user_logged_in',
			)
		);

		// 应用推荐码
		// 前端 Nuxt 使用：POST /wp-json/tanzanite/v1/loyalty/referral/apply
		// Headers: X-WP-Nonce: {nonce}
		// Body: { "code": "REF12345678" }
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/referral/apply',
			array(
				'methods'             => WP_REST_Server::CREATABLE,
				'callback'            => array( $this, 'apply_referral' ),
				'permission_callback' => 'is_user_logged_in',
				'args'                => array(
					'code' => array(
						'type'     => 'string',
						'required' => true,
					),
				),
			)
		);

		// 获取积分配置（公开接口，无需登录）
		// 前端 Nuxt 使用：GET /wp-json/tanzanite/v1/loyalty/config
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/config',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( $this, 'get_config' ),
				'permission_callback' => '__return_true', // 公开接口
			)
		);

		// 获取推荐统计
		// 前端 Nuxt 使用：GET /wp-json/tanzanite/v1/loyalty/referral/stats
		// Headers: X-WP-Nonce: {nonce}
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/referral/stats',
			array(
				'methods'             => WP_REST_Server::READABLE,
				'callback'            => array( $this, 'get_referral_stats' ),
				'permission_callback' => 'is_user_logged_in',
			)
		);
	}

	/**
	 * 每日签到
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function checkin( $request ) {
		global $wpdb;

		$user_id = get_current_user_id();
		$today   = gmdate( 'Y-m-d' );

		// 1. 检查今天是否已签到
		$last_checkin = get_user_meta( $user_id, 'last_checkin_date', true );
		if ( $last_checkin === $today ) {
			return $this->respond_error( 'already_checked_in', __( '今天已经签到过了', 'tanzanite-settings' ), 400 );
		}

		// 2. 获取签到积分配置
		$config   = get_option( 'tanzanite_loyalty_config', '' );
		$settings = json_decode( $config, true );
		if ( ! is_array( $settings ) ) {
			$settings = array();
		}

		$checkin_points = isset( $settings['daily_checkin_points'] ) ? absint( $settings['daily_checkin_points'] ) : 0;

		if ( $checkin_points <= 0 ) {
			return $this->respond_error( 'checkin_disabled', __( '签到功能未启用', 'tanzanite-settings' ), 400 );
		}

		// 3. 增加用户积分
		$current_points = absint( get_user_meta( $user_id, 'loyalty_points', true ) );
		$new_points     = $current_points + $checkin_points;
		update_user_meta( $user_id, 'loyalty_points', $new_points );

		// 4. 更新签到日期
		update_user_meta( $user_id, 'last_checkin_date', $today );

		// 5. 增加连续签到天数
		$last_date = get_user_meta( $user_id, 'last_checkin_date', true );
		$yesterday = gmdate( 'Y-m-d', strtotime( '-1 day' ) );
		
		if ( $last_date === $yesterday ) {
			// 连续签到
			$streak = absint( get_user_meta( $user_id, 'checkin_streak', true ) );
			update_user_meta( $user_id, 'checkin_streak', $streak + 1 );
		} else {
			// 重新开始
			update_user_meta( $user_id, 'checkin_streak', 1 );
		}

		// 6. 记录积分交易
		$wpdb->insert(
			$this->rewards_table,
			array(
				'user_id'      => $user_id,
				'related_type' => 'checkin',
				'related_id'   => null,
				'action'       => 'earn',
				'points_delta' => $checkin_points,
				'amount_delta' => 0,
				'notes'        => __( '每日签到获得积分', 'tanzanite-settings' ),
				'created_at'   => current_time( 'mysql' ),
			)
		);

		// 7. 记录审计日志
		$this->log_audit( 'checkin', 'loyalty', $user_id, array( 'points' => $checkin_points ), $request );

		return $this->respond_success(
			array(
				'message'        => __( '签到成功', 'tanzanite-settings' ),
				'points_earned'  => $checkin_points,
				'total_points'   => $new_points,
				'checkin_streak' => absint( get_user_meta( $user_id, 'checkin_streak', true ) ),
			)
		);
	}

	/**
	 * 生成推荐码
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function generate_referral( $request ) {
		$user_id = get_current_user_id();

		// 检查是否已有推荐码
		$existing_code = get_user_meta( $user_id, 'referral_code', true );
		if ( $existing_code ) {
			return $this->respond_success(
				array(
					'code'    => $existing_code,
					'url'     => home_url( '?ref=' . $existing_code ),
					'message' => __( '已有推荐码', 'tanzanite-settings' ),
				)
			);
		}

		// 生成唯一推荐码
		$code = 'REF' . strtoupper( wp_generate_password( 8, false ) );

		// 确保唯一性
		global $wpdb;
		$exists = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT user_id FROM {$wpdb->usermeta} 
				WHERE meta_key = 'referral_code' AND meta_value = %s",
				$code
			)
		);

		// 如果重复，重新生成
		while ( $exists ) {
			$code   = 'REF' . strtoupper( wp_generate_password( 8, false ) );
			$exists = $wpdb->get_var(
				$wpdb->prepare(
					"SELECT user_id FROM {$wpdb->usermeta} 
					WHERE meta_key = 'referral_code' AND meta_value = %s",
					$code
				)
			);
		}

		// 存储推荐码
		update_user_meta( $user_id, 'referral_code', $code );
		update_user_meta( $user_id, 'referral_code_created', time() );

		// 记录审计日志
		$this->log_audit( 'generate_referral', 'loyalty', $user_id, array( 'code' => $code ), $request );

		return $this->respond_success(
			array(
				'code'    => $code,
				'url'     => home_url( '?ref=' . $code ),
				'message' => __( '推荐码生成成功', 'tanzanite-settings' ),
			)
		);
	}

	/**
	 * 应用推荐码
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function apply_referral( $request ) {
		global $wpdb;

		$user_id = get_current_user_id();
		$code    = sanitize_text_field( $request->get_param( 'code' ) );

		// 1. 查找推荐者
		$inviter_id = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT user_id FROM {$wpdb->usermeta} 
				WHERE meta_key = 'referral_code' AND meta_value = %s",
				$code
			)
		);

		if ( ! $inviter_id ) {
			return $this->respond_error( 'invalid_code', __( '推荐码无效', 'tanzanite-settings' ), 404 );
		}

		if ( intval( $inviter_id ) === $user_id ) {
			return $this->respond_error( 'self_referral', __( '不能使用自己的推荐码', 'tanzanite-settings' ), 400 );
		}

		// 2. 检查是否已使用过推荐码
		$used = get_user_meta( $user_id, 'referral_used', true );
		if ( $used ) {
			return $this->respond_error( 'already_used', __( '已经使用过推荐码', 'tanzanite-settings' ), 400 );
		}

		// 3. 获取奖励配置
		$config   = get_option( 'tanzanite_loyalty_config', '' );
		$settings = json_decode( $config, true );
		if ( ! is_array( $settings ) ) {
			$settings = array();
		}

		$referral = isset( $settings['referral'] ) && is_array( $settings['referral'] ) ? $settings['referral'] : array();

		if ( empty( $referral['enabled'] ) ) {
			return $this->respond_error( 'referral_disabled', __( '推荐功能未启用', 'tanzanite-settings' ), 400 );
		}

		$bonus_inviter = isset( $referral['bonus_inviter'] ) ? absint( $referral['bonus_inviter'] ) : 50;
		$bonus_invitee = isset( $referral['bonus_invitee'] ) ? absint( $referral['bonus_invitee'] ) : 30;

		// 4. 给邀请者增加积分
		$inviter_points = absint( get_user_meta( $inviter_id, 'loyalty_points', true ) );
		update_user_meta( $inviter_id, 'loyalty_points', $inviter_points + $bonus_inviter );

		// 5. 给被邀请者增加积分
		$invitee_points = absint( get_user_meta( $user_id, 'loyalty_points', true ) );
		update_user_meta( $user_id, 'loyalty_points', $invitee_points + $bonus_invitee );

		// 6. 标记已使用
		update_user_meta( $user_id, 'referral_used', true );
		update_user_meta( $user_id, 'referred_by', $inviter_id );

		// 7. 增加邀请者的推荐计数
		$inviter_count = absint( get_user_meta( $inviter_id, 'referral_count', true ) );
		update_user_meta( $inviter_id, 'referral_count', $inviter_count + 1 );

		// 8. 记录积分交易 - 邀请者
		$wpdb->insert(
			$this->rewards_table,
			array(
				'user_id'      => $inviter_id,
				'related_type' => 'referral',
				'related_id'   => $user_id,
				'action'       => 'earn',
				'points_delta' => $bonus_inviter,
				'amount_delta' => 0,
				'notes'        => sprintf( __( '推荐用户 %d 获得奖励', 'tanzanite-settings' ), $user_id ),
				'created_at'   => current_time( 'mysql' ),
			)
		);

		// 9. 记录积分交易 - 被邀请者
		$wpdb->insert(
			$this->rewards_table,
			array(
				'user_id'      => $user_id,
				'related_type' => 'referral',
				'related_id'   => $inviter_id,
				'action'       => 'earn',
				'points_delta' => $bonus_invitee,
				'amount_delta' => 0,
				'notes'        => sprintf( __( '使用推荐码 %s 获得奖励', 'tanzanite-settings' ), $code ),
				'created_at'   => current_time( 'mysql' ),
			)
		);

		// 10. 记录审计日志
		$this->log_audit(
			'apply_referral',
			'loyalty',
			$user_id,
			array(
				'code'          => $code,
				'inviter_id'    => $inviter_id,
				'inviter_bonus' => $bonus_inviter,
				'invitee_bonus' => $bonus_invitee,
			),
			$request
		);

		return $this->respond_success(
			array(
				'message'       => __( '推荐码应用成功', 'tanzanite-settings' ),
				'points_earned' => $bonus_invitee,
				'total_points'  => $invitee_points + $bonus_invitee,
			)
		);
	}

	/**
	 * 获取推荐统计
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_referral_stats( $request ) {
		$user_id = get_current_user_id();

		$referral_code  = get_user_meta( $user_id, 'referral_code', true );
		$referral_count = absint( get_user_meta( $user_id, 'referral_count', true ) );
		$referred_by    = get_user_meta( $user_id, 'referred_by', true );

		$stats = array(
			'referral_code'  => $referral_code ?: null,
			'referral_url'   => $referral_code ? home_url( '?ref=' . $referral_code ) : null,
			'referral_count' => $referral_count,
			'referred_by'    => $referred_by ? intval( $referred_by ) : null,
		);

		return $this->respond_success( $stats );
	}

	/**
	 * 获取积分配置（公开接口）
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_config( $request ) {
		// 获取积分配置
		$config_json = get_option( 'tanzanite_loyalty_config', '' );
		$config      = json_decode( $config_json, true );

		// 如果配置为空，使用默认配置
		if ( ! is_array( $config ) || empty( $config ) ) {
			$config = $this->get_default_config();
		}

		// 格式化会员等级数据，供前端使用
		$tiers = array();
		if ( isset( $config['tiers'] ) && is_array( $config['tiers'] ) ) {
			foreach ( $config['tiers'] as $key => $tier ) {
				$tiers[] = array(
					'key'      => $key,
					'name'     => $tier['name'] ?? $tier['label'] ?? ucfirst( $key ),
					'label'    => $tier['label'] ?? $tier['name'] ?? ucfirst( $key ),
					'min'      => (int) ( $tier['min'] ?? 0 ),
					'max'      => $tier['max'] === null ? -1 : (int) $tier['max'],
					'discount' => (float) ( $tier['discount'] ?? 0 ),
					'redeem'   => array(
						'enabled'              => (bool) ( $tier['redeem']['enabled'] ?? false ),
						'percent_of_total'     => (float) ( $tier['redeem']['percent_of_total'] ?? 5 ),
						'value_per_point_base' => (float) ( $tier['redeem']['value_per_point_base'] ?? 0.01 ),
						'min_points'           => (int) ( $tier['redeem']['min_points'] ?? 0 ),
						'stack_with_percent'   => (bool) ( $tier['redeem']['stack_with_percent'] ?? true ),
					),
				);
			}
		}

		// 格式化推荐奖励配置
		$referral = array(
			'enabled'        => (bool) ( $config['referral']['enabled'] ?? true ),
			'bonus_inviter'  => (int) ( $config['referral']['bonus_inviter'] ?? 50 ),
			'bonus_invitee'  => (int) ( $config['referral']['bonus_invitee'] ?? 30 ),
			'token_ttl_days' => (int) ( $config['referral']['token_ttl_days'] ?? 7 ),
			'token_max_uses' => (int) ( $config['referral']['token_max_uses'] ?? 50 ),
		);

		// 返回格式化后的配置
		return rest_ensure_response(
			array(
				'enabled'              => (bool) ( $config['enabled'] ?? true ),
				'apply_cart_discount'  => (bool) ( $config['apply_cart_discount'] ?? true ),
				'points_per_unit'      => (float) ( $config['points_per_unit'] ?? 1 ),
				'daily_checkin_points' => (int) ( $config['daily_checkin_points'] ?? 0 ),
				'tiers'                => $tiers,
				'referral'             => $referral,
			)
		);
	}

	/**
	 * 获取默认积分配置
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_default_config() {
		return array(
			'enabled'              => true,
			'apply_cart_discount'  => true,
			'points_per_unit'      => 1,
			'daily_checkin_points' => 0,
			'referral'             => array(
				'enabled'        => true,
				'bonus_inviter'  => 50,
				'bonus_invitee'  => 30,
				'token_ttl_days' => 7,
				'token_max_uses' => 50,
			),
			'tiers'                => array(
				'ordinary' => array(
					'label'    => 'Ordinary',
					'name'     => 'Ordinary',
					'min'      => 0,
					'max'      => 499,
					'discount' => 0,
					'redeem'   => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'bronze'   => array(
					'label'    => 'Bronze',
					'name'     => 'Bronze',
					'min'      => 500,
					'max'      => 1999,
					'discount' => 5,
					'redeem'   => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'silver'   => array(
					'label'    => 'Silver',
					'name'     => 'Silver',
					'min'      => 2000,
					'max'      => 4999,
					'discount' => 10,
					'redeem'   => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'gold'     => array(
					'label'    => 'Gold',
					'name'     => 'Gold',
					'min'      => 5000,
					'max'      => 9999,
					'discount' => 15,
					'redeem'   => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'platinum' => array(
					'label'    => 'Platinum',
					'name'     => 'Platinum',
					'min'      => 10000,
					'max'      => null,
					'discount' => 20,
					'redeem'   => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
			),
		);
	}
}
