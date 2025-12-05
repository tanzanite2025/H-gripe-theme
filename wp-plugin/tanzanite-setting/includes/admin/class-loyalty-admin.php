<?php
/**
 * Loyalty Admin Page
 * 
 * 负责渲染积分/等级设置页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.6
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 积分设置管理类
 */
class Tanzanite_Loyalty_Admin {

	/**
	 * 渲染积分/等级设置页面
	 */
	public static function render_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		$settings = self::get_loyalty_settings();

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.6';

		// 加载 WordPress 组件样式
		wp_enqueue_style( 'wp-components' );
		
		// 加载 Loyalty Settings JS
		wp_enqueue_script(
			'tz-loyalty-settings',
			TANZANITE_PLUGIN_URL . 'assets/js/loyalty-settings.js',
			array( 'tz-admin-common', 'wp-element', 'wp-i18n', 'wp-components', 'wp-url', 'wp-api-fetch' ),
			$version . '.sync.' . time(),
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-loyalty-settings',
			'TzLoyaltyConfig',
			array(
				'nonce'    => wp_create_nonce( 'wp_rest' ),
				'settings' => $settings,
			)
		);

		echo '<div class="tz-settings-wrapper tz-loyalty-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Loyalty & Points Settings', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '配置会员等级、积分规则、推荐奖励等忠诚度系统设置。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-loyalty-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <form id="tz-loyalty-form" method="post" action="options.php">';
		settings_fields( 'tanzanite_loyalty' );
		echo '      <input type="hidden" id="tz-loyalty-config" name="tanzanite_loyalty_config" value="' . esc_attr( wp_json_encode( $settings, JSON_UNESCAPED_UNICODE ) ) . '" />';
		echo '      <div id="tz-loyalty-app"></div>';
		echo '      <noscript><p class="notice notice-error">' . esc_html__( '需要启用 JavaScript 才能管理积分等级。', 'tanzanite-settings' ) . '</p></noscript>';
		echo '      <div style="display:flex;gap:12px;align-items:center;margin-top:16px;">';
		submit_button( __( '保存设置', 'tanzanite-settings' ), 'primary', 'submit', false, [ 'id' => 'tz-loyalty-submit' ] );
		echo '          <button type="button" class="button" id="tz-loyalty-reset" onclick="resetLoyaltyToEnglish()">' . esc_html__( '重置会员等级为英文', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </form>';
		
		// 添加重置脚本
		echo '<script type="text/javascript">';
		echo 'function resetLoyaltyToEnglish() {';
		echo '    if (!confirm("确定要重置会员等级名称为英文吗？这将覆盖当前的自定义设置。")) return;';
		echo '    ';
		echo '    fetch("' . esc_url( admin_url( 'admin-ajax.php' ) ) . '", {';
		echo '        method: "POST",';
		echo '        headers: { "Content-Type": "application/x-www-form-urlencoded" },';
		echo '        body: "action=reset_loyalty_to_english&nonce=' . wp_create_nonce( 'reset_loyalty_english' ) . '"';
		echo '    })';
		echo '    .then(response => response.json())';
		echo '    .then(data => {';
		echo '        if (data.success) {';
		echo '            alert("会员等级已重置为英文，请刷新页面查看效果。");';
		echo '            location.reload();';
		echo '        } else {';
		echo '            alert("重置失败：" + (data.data || "未知错误"));';
		echo '        }';
		echo '    })';
		echo '    .catch(error => {';
		echo '        console.error("Reset error:", error);';
		echo '        alert("重置失败，请稍后重试。");';
		echo '    });';
		echo '}';
		echo '</script>';

		echo '</div>';
	}

	/**
	 * 获取积分/等级设置
	 */
	private static function get_loyalty_settings(): array {
		$default = self::get_default_loyalty_config();
		$saved_json = get_option( 'tanzanite_loyalty_config', '' );

		// 数据库中存储的是 JSON 字符串，需要解码
		if ( empty( $saved_json ) ) {
			return $default;
		}

		$saved = json_decode( $saved_json, true );
		if ( ! is_array( $saved ) || empty( $saved ) ) {
			return $default;
		}

		return array_replace_recursive( $default, $saved );
	}

	/**
	 * 获取默认积分/等级配置
	 */
	public static function get_default_loyalty_config(): array {
		return array(
			'enabled'              => true,
			'apply_cart_discount'  => true,
			'points_per_unit'      => 1,
			'daily_checkin_points' => 0,
			'referral'             => array(
				'enabled'         => true,
				'bonus_inviter'   => 50,
				'bonus_invitee'   => 30,
				'token_ttl_days'  => 7,
				'token_max_uses'  => 50,
			),
			'tiers'                => array(
				'ordinary' => array(
					'label'      => __( 'Ordinary', 'tanzanite-settings' ),
					'name'       => __( 'Ordinary', 'tanzanite-settings' ),
					'min'        => 0,
					'max'        => 499,
					'discount'   => 0,
					'products'   => array(),
					'categories' => array(),
					'redeem'     => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'bronze'   => array(
					'label'      => __( 'Bronze', 'tanzanite-settings' ),
					'name'       => __( 'Bronze', 'tanzanite-settings' ),
					'min'        => 500,
					'max'        => 1999,
					'discount'   => 5,
					'products'   => array(),
					'categories' => array(),
					'redeem'     => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'silver'   => array(
					'label'      => __( 'Silver', 'tanzanite-settings' ),
					'name'       => __( 'Silver', 'tanzanite-settings' ),
					'min'        => 2000,
					'max'        => 4999,
					'discount'   => 10,
					'products'   => array(),
					'categories' => array(),
					'redeem'     => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'gold'     => array(
					'label'      => __( 'Gold', 'tanzanite-settings' ),
					'name'       => __( 'Gold', 'tanzanite-settings' ),
					'min'        => 5000,
					'max'        => 9999,
					'discount'   => 15,
					'products'   => array(),
					'categories' => array(),
					'redeem'     => array(
						'enabled'              => false,
						'percent_of_total'     => 5,
						'value_per_point_base' => 0.01,
						'min_points'           => 0,
						'stack_with_percent'   => true,
					),
				),
				'platinum' => array(
					'label'      => __( 'Platinum', 'tanzanite-settings' ),
					'name'       => __( 'Platinum', 'tanzanite-settings' ),
					'min'        => 10000,
					'max'        => null,
					'discount'   => 20,
					'products'   => array(),
					'categories' => array(),
					'redeem'     => array(
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

	/**
	 * 注册积分/等级设置选项
	 */
	public static function register_settings() {
		register_setting(
			'tanzanite_loyalty',
			'tanzanite_loyalty_config',
			array(
				'type'              => 'string',
				'sanitize_callback' => array( __CLASS__, 'sanitize_loyalty_config' ),
				'default'           => wp_json_encode( self::get_default_loyalty_config() ),
			)
		);
	}

	/**
	 * 消毒/验证积分配置
	 */
	public static function sanitize_loyalty_config( $input ) {
		// 简单验证：确保是合法的 JSON
		$data = json_decode( $input, true );
		if ( ! is_array( $data ) ) {
			add_settings_error( 'tanzanite_loyalty_config', 'invalid_json', __( '配置格式无效', 'tanzanite-settings' ) );
			return get_option( 'tanzanite_loyalty_config' );
		}
		// 可以在这里添加更详细的字段验证逻辑
		return wp_json_encode( $data, JSON_UNESCAPED_UNICODE );
	}

	/**
	 * 处理重置会员等级为英文的 AJAX 请求
	 */
	public static function handle_reset_to_english() {
		// 验证 nonce
		if ( ! wp_verify_nonce( $_POST['nonce'] ?? '', 'reset_loyalty_english' ) ) {
			wp_send_json_error( '安全验证失败' );
			return;
		}

		// 检查权限
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_send_json_error( '权限不足' );
			return;
		}

		try {
			// 删除现有的设置，这样会使用默认的英文配置
			delete_option( 'tanzanite_loyalty_config' );
			
			wp_send_json_success( '会员等级已重置为英文' );
		} catch ( Exception $e ) {
			wp_send_json_error( '重置失败：' . $e->getMessage() );
		}
	}
}
