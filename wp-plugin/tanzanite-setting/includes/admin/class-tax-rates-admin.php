<?php
/**
 * Tax Rates Admin Page
 * 
 * 负责渲染税率管理页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.7
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 税率管理类
 */
class Tanzanite_Tax_Rates_Admin {

	/**
	 * 渲染税率管理页面
	 */
	public static function render_page() {
		$nonce = wp_create_nonce( 'wp_rest' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.7';

		// 加载税率管理 JS
		wp_enqueue_script(
			'tz-tax-rates',
			TANZANITE_PLUGIN_URL . 'assets/js/tax-rates.js',
			array( 'jquery' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-tax-rates',
			'TzTaxRatesConfig',
			array(
				'listUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/tax-rates' ) ),
				'singleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/tax-rates/' ) ),
				'nonce'     => $nonce,
			)
		);

		echo '<div class="tz-settings-wrapper tz-tax-rates-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Tax Rates', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '配置税率模板，商品可关联一个或多个税率，前端下单时自动计算税费。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-tax-rate-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '税率模板列表', 'tanzanite-settings' ) . '</div>';
		echo '      <button class="button button-primary" id="tz-tax-rate-create">' . esc_html__( '新增税率', 'tanzanite-settings' ) . '</button>';
		echo '      <div style="overflow:auto;margin-top:16px;">';
		echo '          <table class="widefat fixed striped" id="tz-tax-rate-table" style="min-width:800px;">';
		echo '              <thead><tr>';
		foreach ( [ 'Name', 'Rate (%)', 'Region', 'Active', 'Actions' ] as $column ) {
			echo '<th>' . esc_html( $column ) . '</th>';
		}
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增税率', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-tax-rate-form" class="tz-tax-rate-form">';
		echo '          <input type="hidden" id="tz-tax-rate-id" />';
		echo '          <div class="tz-form-grid">';
		echo '              <label>' . esc_html__( '名称', 'tanzanite-settings' ) . '<input type="text" id="tz-tax-rate-name" required /></label>';
		echo '              <label>' . esc_html__( '税率 (%)', 'tanzanite-settings' ) . '<input type="number" step="0.0001" id="tz-tax-rate-rate" min="0" required /></label>';
		echo '              <label>' . esc_html__( '地区', 'tanzanite-settings' ) . '<input type="text" id="tz-tax-rate-region" /></label>';
		echo '              <label>' . esc_html__( '排序', 'tanzanite-settings' ) . '<input type="number" id="tz-tax-rate-sort" value="0" /></label>';
		echo '              <label>' . esc_html__( '启用', 'tanzanite-settings' ) . '<select id="tz-tax-rate-active"><option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option><option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option></select></label>';
		echo '          </div>';
		echo '          <label>' . esc_html__( '描述', 'tanzanite-settings' ) . '<textarea id="tz-tax-rate-description" rows="3"></textarea></label>';
		echo '          <label>' . esc_html__( 'Meta 信息 (JSON，可选)', 'tanzanite-settings' ) . '<textarea id="tz-tax-rate-meta" rows="3"></textarea></label>';
		echo '          <div style="margin-top:16px;display:flex;gap:12px;">';
		echo '              <button class="button button-primary" id="tz-tax-rate-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button" id="tz-tax-rate-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '      </form>';
		echo '  </div>';

		echo '</div>';
	}
}
