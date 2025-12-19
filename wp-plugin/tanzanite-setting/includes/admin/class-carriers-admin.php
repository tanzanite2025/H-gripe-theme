<?php
/**
 * Carriers Admin Page
 * 
 * 负责渲染物流商管理页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.3
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 物流商管理类
 */
class Tanzanite_Carriers_Admin {

	/**
	 * 预定义物流服务商配置
	 * 
	 * 从 Tanzanite_Settings_Plugin::TRACKING_PROVIDERS 复制而来
	 */
	const TRACKING_PROVIDERS = [
		'17track' => [
			'label' => '17TRACK',
			'fields' => [
				'api_key'    => [ 'label' => 'API Key',    'type' => 'password' ],
				'secret_key' => [ 'label' => 'Secret Key', 'type' => 'password' ],
				'endpoint'   => [ 'label' => 'API Endpoint', 'type' => 'text', 'default' => 'https://api.17track.net/track' ],
			],
		],
	];

	/**
	 * 渲染物流商管理页面
	 */
	public static function render_page() {
		$active_tab = isset( $_GET['tab'] ) ? sanitize_key( $_GET['tab'] ) : 'config'; // phpcs:ignore WordPress.Security.NonceVerification.Recommended
		$nonce      = wp_create_nonce( 'wp_rest' );
		$list_url   = esc_url_raw( rest_url( 'tanzanite/v1/carriers' ) );
		$single_url = esc_url_raw( rest_url( 'tanzanite/v1/carriers/' ) );
		$can_manage = current_user_can( 'manage_options' );

		if ( 'list' === $active_tab ) {
			wp_safe_redirect( admin_url( 'admin.php?page=tanzanite-settings-carriers&tab=config' ) );
			exit;
		}

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.3';

		if ( 'list' === $active_tab ) {
			wp_enqueue_script(
				'tz-carriers',
				TANZANITE_PLUGIN_URL . 'assets/js/carriers.js',
				array( 'tz-admin-common' ),
				$version,
				true
			);

			wp_localize_script(
				'tz-carriers',
				'TzCarriersConfig',
				array(
					'nonce'     => $nonce,
					'listUrl'   => $list_url,
					'singleUrl' => $single_url,
					'canManage' => $can_manage,
				)
			);

			echo '<div class="tz-settings-wrapper">';
			echo '  <div class="tz-settings-header">';
			echo '      <h1>' . esc_html__( 'Carriers & Tracking', 'tanzanite-settings' ) . '</h1>';
			echo '      <p>' . esc_html__( '管理物流公司信息和追踪配置。', 'tanzanite-settings' ) . '</p>';
			echo '  </div>';

			echo '  <nav class="nav-tab-wrapper">';
			echo '      <a href="?page=tanzanite-settings-carriers&tab=list" class="nav-tab nav-tab-active">' . esc_html__( '物流公司管理', 'tanzanite-settings' ) . '</a>';
			echo '      <a href="?page=tanzanite-settings-carriers&tab=config" class="nav-tab">' . esc_html__( 'API 配置', 'tanzanite-settings' ) . '</a>';
			echo '  </nav>';

			echo '  <div id="tz-carriers-notice" class="notice" style="display:none;margin-top:16px;"></div>';

			echo '  <div class="tz-settings-section" style="margin-top:24px;">';
			echo '      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;">';
			echo '          <h2>' . esc_html__( '物流公司列表', 'tanzanite-settings' ) . '</h2>';
			if ( $can_manage ) {
				echo '          <button class="button button-primary" id="tz-carriers-create">' . esc_html__( '新建物流公司', 'tanzanite-settings' ) . '</button>';
			}
			echo '      </div>';

			echo '      <table class="widefat fixed striped" id="tz-carriers-table">';
			echo '          <thead>';
			echo '              <tr>';
			echo '                  <th style="width:120px;">' . esc_html__( '编码', 'tanzanite-settings' ) . '</th>';
			echo '                  <th>' . esc_html__( '名称', 'tanzanite-settings' ) . '</th>';
			echo '                  <th style="width:150px;">' . esc_html__( '联系人', 'tanzanite-settings' ) . '</th>';
			echo '                  <th style="width:120px;">' . esc_html__( '电话', 'tanzanite-settings' ) . '</th>';
			echo '                  <th style="width:200px;">' . esc_html__( '追踪 URL', 'tanzanite-settings' ) . '</th>';
			echo '                  <th style="width:80px;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
			echo '                  <th style="width:120px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
			echo '              </tr>';
			echo '          </thead>';
			echo '          <tbody></tbody>';
			echo '      </table>';
			echo '  </div>';

			if ( $can_manage ) {
				echo '  <div class="tz-settings-section" style="margin-top:24px;">';
				echo '      <h2>' . esc_html__( '新建/编辑物流公司', 'tanzanite-settings' ) . '</h2>';
				echo '      <form id="tz-carriers-form" style="max-width:800px;">';
				echo '          <input type="hidden" id="tz-carrier-id" />';
				echo '          <div class="tz-form-grid">';
				echo '              <label><strong>' . esc_html__( '编码', 'tanzanite-settings' ) . ' *</strong>';
				echo '                  <input type="text" id="tz-carrier-code" class="regular-text" required placeholder="sf_express" />';
				echo '              </label>';
				echo '              <label><strong>' . esc_html__( '名称', 'tanzanite-settings' ) . ' *</strong>';
				echo '                  <input type="text" id="tz-carrier-name" class="regular-text" required placeholder="顺丰速运" />';
				echo '              </label>';
				echo '              <label><strong>' . esc_html__( '联系人', 'tanzanite-settings' ) . '</strong>';
				echo '                  <input type="text" id="tz-carrier-contact-person" class="regular-text" placeholder="张三" />';
				echo '              </label>';
				echo '              <label><strong>' . esc_html__( '联系电话', 'tanzanite-settings' ) . '</strong>';
				echo '                  <input type="text" id="tz-carrier-contact-phone" class="regular-text" placeholder="400-111-1111" />';
				echo '              </label>';
				echo '          </div>';
				echo '          <label style="display:block;margin-top:12px;"><strong>' . esc_html__( '追踪 URL 模板', 'tanzanite-settings' ) . '</strong>';
				echo '              <input type="url" id="tz-carrier-tracking-url" class="regular-text" placeholder="https://www.sf-express.com/cn/sc/dynamic_function/waybill/#search_waybill={{tracking_number}}" />';
				echo '              <p class="description">' . esc_html__( '使用 {{tracking_number}} 作为运单号占位符', 'tanzanite-settings' ) . '</p>';
				echo '          </label>';
				echo '          <label style="display:block;margin-top:12px;"><strong>' . esc_html__( '服务地区', 'tanzanite-settings' ) . '</strong>';
				echo '              <input type="text" id="tz-carrier-service-regions" class="regular-text" placeholder="中国大陆, 香港, 台湾（逗号分隔）" />';
				echo '          </label>';
				echo '          <div class="tz-form-grid" style="margin-top:12px;">';
				echo '              <label><strong>' . esc_html__( '状态', 'tanzanite-settings' ) . '</strong>';
				echo '                  <select id="tz-carrier-is-active" class="regular-text">';
				echo '                      <option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option>';
				echo '                      <option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option>';
				echo '                  </select>';
				echo '              </label>';
				echo '              <label><strong>' . esc_html__( '排序', 'tanzanite-settings' ) . '</strong>';
				echo '                  <input type="number" id="tz-carrier-sort-order" class="regular-text" value="0" />';
				echo '              </label>';
				echo '          </div>';
				echo '          <label style="display:block;margin-top:12px;"><strong>' . esc_html__( 'Meta (JSON)', 'tanzanite-settings' ) . '</strong>';
				echo '              <textarea id="tz-carrier-meta" class="large-text code" rows="4" placeholder=\'{"key": "value"}\'></textarea>';
				echo '          </label>';
				echo '          <div style="margin-top:16px;">';
				echo '              <button type="submit" class="button button-primary" id="tz-carriers-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
				echo '              <button type="button" class="button" id="tz-carriers-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
				echo '          </div>';
				echo '      </form>';
				echo '  </div>';
			}

			echo '</div>';

		} elseif ( 'config' === $active_tab ) {
			$tracking_option = get_option( 'tanzanite_tracking_settings', [] );
			$provider        = $tracking_option['provider'] ?? '17track';
			$settings        = $tracking_option['settings'][ $provider ] ?? [];

			wp_enqueue_script(
				'tz-carriers-config',
				TANZANITE_PLUGIN_URL . 'assets/js/carriers-config.js',
				array( 'tz-admin-common' ),
				$version,
				true
			);

			wp_localize_script(
				'tz-carriers-config',
				'TzCarriersConfigPage',
				array(
					'providers'       => self::TRACKING_PROVIDERS,
					'currentProvider' => $provider,
					'currentSettings' => $settings,
					'testUrl'         => admin_url( 'admin-post.php?action=tanz_test_tracking' ),
				)
			);

			echo '<div class="tz-settings-wrapper">';
			echo '  <div class="tz-settings-header">';
			echo '      <h1>' . esc_html__( 'Carriers & Tracking', 'tanzanite-settings' ) . '</h1>';
			echo '      <p>' . esc_html__( '管理物流公司信息和追踪配置。', 'tanzanite-settings' ) . '</p>';
			echo '  </div>';

			echo '  <nav class="nav-tab-wrapper">';
			echo '      <a href="?page=tanzanite-settings-carriers&tab=config" class="nav-tab nav-tab-active">' . esc_html__( 'API 配置', 'tanzanite-settings' ) . '</a>';
			echo '  </nav>';

			echo '  <div id="tz-carriers-config-notice" class="notice" style="display:none;margin-top:16px;"></div>';

			echo '  <div class="tz-settings-section" style="margin-top:24px;">';
			echo '      <h2>' . esc_html__( '物流追踪 API 配置', 'tanzanite-settings' ) . '</h2>';
			echo '      <form method="post" action="' . esc_url( admin_url( 'admin-post.php' ) ) . '" id="tz-tracking-config-form" style="max-width:600px;">';
			wp_nonce_field( 'tanz_tracking_settings' );
			echo '          <input type="hidden" name="action" value="tanz_save_tracking_settings" />';
			echo '          <label style="display:block;margin-bottom:16px;"><strong>' . esc_html__( '追踪服务商', 'tanzanite-settings' ) . '</strong>';
			echo '              <select name="provider" id="tz-tracking-provider" class="regular-text" style="display:block;margin-top:4px;">';
			foreach ( self::TRACKING_PROVIDERS as $key => $config ) {
				$selected = ( $key === $provider ) ? ' selected' : '';
				echo '                  <option value="' . esc_attr( $key ) . '"' . $selected . '>' . esc_html( $config['label'] ) . '</option>';
			}
			echo '              </select>';
			echo '          </label>';
			echo '          <div id="tz-tracking-fields"></div>';
			echo '          <div style="margin-top:16px;">';
			echo '              <button type="submit" class="button button-primary" id="tz-tracking-save">' . esc_html__( '保存配置', 'tanzanite-settings' ) . '</button>';
			echo '              <button type="button" class="button" id="tz-tracking-test">' . esc_html__( '测试连接', 'tanzanite-settings' ) . '</button>';
			echo '          </div>';
			echo '      </form>';
			echo '  </div>';

			echo '</div>';
		}
	}
}
