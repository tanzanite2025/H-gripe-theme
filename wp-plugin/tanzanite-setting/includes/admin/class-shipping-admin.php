<?php
/**
 * Shipping Templates Admin Page
 * 
 * 负责渲染运费模板管理页面
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
 * 运费模板管理类
 */
class Tanzanite_Shipping_Admin {

	/**
	 * 渲染运费模板管理页面
	 */
	public static function render_page() {
		$nonce = wp_create_nonce( 'wp_rest' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.6';

		// 加载配送模板管理 JS
		wp_enqueue_script(
			'tz-shipping-templates',
			TANZANITE_PLUGIN_URL . 'assets/js/shipping-templates.js',
			array( 'tz-admin-common' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-shipping-templates',
			'TzShippingConfig',
			array(
				'listUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/shipping-templates' ) ),
				'singleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/shipping-templates/' ) ),
				'nonce'     => $nonce,
			)
		);

		echo '<div class="tz-settings-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Shipping Templates', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '定义配送规则、包邮策略与配送时效说明。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-shipping-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '配送模板列表', 'tanzanite-settings' ) . '</div>';
		echo '      <button class="button button-primary" id="tz-shipping-create">' . esc_html__( '新增模板', 'tanzanite-settings' ) . '</button>';
		echo '      <button class="button" id="tz-shipping-export" style="margin-left:8px;">' . esc_html__( '导出 JSON', 'tanzanite-settings' ) . '</button>';
		echo '      <div style="overflow:auto;margin-top:16px;">';
		echo '          <table class="widefat fixed striped" id="tz-shipping-table">';
		echo '              <thead><tr>';
		echo '                  <th style="width:25%;">' . esc_html__( '模板名称', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:35%;">' . esc_html__( '描述', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:10%;">' . esc_html__( '规则数', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:10%;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:20%;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增配送模板', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-shipping-form">';
		echo '          <input type="hidden" id="tz-shipping-id" />';
		echo '          <div class="tz-form-grid">';
		echo '              <label>' . esc_html__( '模板名称', 'tanzanite-settings' ) . '<input type="text" id="tz-shipping-name" required /></label>';
		echo '              <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select id="tz-shipping-active"><option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option><option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option></select></label>';
		echo '              <label>' . esc_html__( '物流公司编码 (Carrier)', 'tanzanite-settings' ) . '<input type="text" id="tz-shipping-carrier" placeholder="sf_express" /></label>';
		echo '          </div>';
		echo '          <label>' . esc_html__( '描述', 'tanzanite-settings' ) . '<textarea id="tz-shipping-description" rows="2"></textarea></label>';
		
		echo '          <div style="margin-top:20px;">';
		echo '              <strong>' . esc_html__( '配送规则', 'tanzanite-settings' ) . '</strong>';
		echo '              <p class="description">' . esc_html__( '按重量、金额、件数等条件设置运费。支持多条规则，系统将按优先级匹配。', 'tanzanite-settings' ) . '</p>';
		echo '              <div id="tz-shipping-rules-list" style="margin-top:12px;"></div>';
		echo '              <div id="tz-shipping-rule-editor" style="margin-top:16px;padding:12px;border:1px solid #e5e7eb;border-radius:4px;background:#f9fafb;">';
		echo '                  <div style="margin-bottom:8px;"><strong>' . esc_html__( '规则编辑', 'tanzanite-settings' ) . '</strong><span class="description" style="margin-left:8px;">' . esc_html__( '填写国家、区间和运费，保存后将加入上方规则列表。', 'tanzanite-settings' ) . '</span></div>';
		echo '                  <div class="tz-form-grid">';
		echo '                      <label>' . esc_html__( '规则类型', 'tanzanite-settings' ) . '<select id="tz-shipping-rule-type"><option value="weight">' . esc_html__( '按重量', 'tanzanite-settings' ) . '</option><option value="amount">' . esc_html__( '按金额', 'tanzanite-settings' ) . '</option><option value="quantity">' . esc_html__( '按件数', 'tanzanite-settings' ) . '</option><option value="volume">' . esc_html__( '按体积', 'tanzanite-settings' ) . '</option><option value="items">' . esc_html__( '按商品数', 'tanzanite-settings' ) . '</option></select></label>';
		echo '                      <label>' . esc_html__( '服务编码', 'tanzanite-settings' ) . '<input type="text" id="tz-shipping-rule-service" placeholder="air, sea, express" /></label>';
		echo '                      <label>' . esc_html__( '服务名称', 'tanzanite-settings' ) . '<input type="text" id="tz-shipping-rule-service-label" placeholder="空运、海运" /></label>';
		echo '                  </div>';
		echo '                  <div class="tz-form-grid">';
		echo '                      <label>' . esc_html__( '适用国家代码', 'tanzanite-settings' ) . '<input type="text" id="tz-shipping-rule-regions" placeholder="JP 或 JP,US（ISO 代码，逗号分隔）" /></label>';
		echo '                      <label>' . esc_html__( '邮编范围', 'tanzanite-settings' ) . '<input type="text" id="tz-shipping-rule-zip-ranges" placeholder="10001-10999,11001-11999（留空=全国）" /></label>';
		echo '                  </div>';
		echo '                  <p class="description" style="margin-top:4px;color:#6b7280;">' . esc_html__( '邮编范围示例：美国纽约 10001-10999，阿拉斯加 99501-99950。留空表示该国家的兜底规则。', 'tanzanite-settings' ) . '</p>';
		echo '                  <div class="tz-form-grid">';
		echo '                      <label><span id="tz-shipping-rule-min-label">' . esc_html__( '最小重量 (KG)', 'tanzanite-settings' ) . '</span><input type="number" step="0.01" id="tz-shipping-rule-min" placeholder="' . esc_attr__( '如 0、0.5、1', 'tanzanite-settings' ) . '" /></label>';
		echo '                      <label><span id="tz-shipping-rule-max-label">' . esc_html__( '最大重量 (KG)', 'tanzanite-settings' ) . '</span><input type="number" step="0.01" id="tz-shipping-rule-max" placeholder="' . esc_attr__( '留空表示无上限', 'tanzanite-settings' ) . '" /></label>';
		echo '                  </div>';
		echo '                  <p class="description" id="tz-shipping-rule-range-hint" style="margin-top:4px;color:#6b7280;">' . esc_html__( '示例：0~0.5KG 运费15元，0.5~1KG 运费25元。请为每个重量区间创建一条规则。', 'tanzanite-settings' ) . '</p>';
		echo '                  <div class="tz-form-grid">';
		echo '                      <label>' . esc_html__( '运费金额', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-shipping-rule-fee" required /></label>';
		echo '                      <label>' . esc_html__( '满额包邮', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-shipping-rule-free-over" placeholder="' . esc_attr__( '留空表示不包邮', 'tanzanite-settings' ) . '" /></label>';
		echo '                      <label>' . esc_html__( '预计时效（天）', 'tanzanite-settings' ) . '<input type="number" id="tz-shipping-rule-eta-min" placeholder="' . esc_attr__( '最少', 'tanzanite-settings' ) . '" style="width:80px;margin-right:4px;" /><input type="number" id="tz-shipping-rule-eta-max" placeholder="' . esc_attr__( '最多', 'tanzanite-settings' ) . '" style="width:80px;" /></label>';
		echo '                  </div>';
		echo '                  <div style="margin-top:8px;">';
		echo '                      <button type="button" class="button button-primary" id="tz-shipping-rule-save">' . esc_html__( '保存规则', 'tanzanite-settings' ) . '</button>';
		echo '                      <button type="button" class="button" id="tz-shipping-rule-reset" style="margin-left:8px;">' . esc_html__( '清空', 'tanzanite-settings' ) . '</button>';
		echo '                  </div>';
		echo '              </div>';
		echo '              <button type="button" class="button" id="tz-shipping-add-rule" style="margin-top:12px;">' . esc_html__( '新增一条规则', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';

		echo '          <div style="margin-top:16px;display:flex;gap:12px;">';
		echo '              <button class="button button-primary" id="tz-shipping-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button" id="tz-shipping-reset" type="button">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '      </form>';
		echo '  </div>';

		echo '</div>';
	}
}
