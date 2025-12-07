<?php
/**
 * Packaging Rules Admin Page
 *
 * 负责渲染包装规则管理页面
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.3.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 包装规则管理类
 */
class Tanzanite_Packaging_Admin {

	/**
	 * 渲染包装规则管理页面
	 */
	public static function render_page() {
		$nonce = wp_create_nonce( 'wp_rest' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.3.0';

		// 加载包装规则管理 JS
		wp_enqueue_script(
			'tz-packaging-rules',
			TANZANITE_PLUGIN_URL . 'assets/js/packaging-rules.js',
			array( 'tz-admin-common' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-packaging-rules',
			'TzPackagingConfig',
			array(
				'listUrl'    => esc_url_raw( rest_url( 'tanzanite/v1/packaging-rules' ) ),
				'singleUrl'  => esc_url_raw( rest_url( 'tanzanite/v1/packaging-rules/' ) ),
				'installUrl' => esc_url_raw( rest_url( 'tanzanite/v1/packaging-rules/install' ) ),
				'nonce'      => $nonce,
			)
		);

		echo '<div class="tz-settings-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Packaging Rules', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '定义商品包装规则，用于计算运费时的包装重量和体积。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-packaging-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		// 数据库安装提示
		echo '  <div id="tz-packaging-install-section" class="tz-settings-section" style="display:none;background:#fff8e1;border-left:4px solid #ffc107;">';
		echo '      <div class="tz-section-title">' . esc_html__( '数据库初始化', 'tanzanite-settings' ) . '</div>';
		echo '      <p>' . esc_html__( '首次使用需要创建数据库表。点击下方按钮进行初始化。', 'tanzanite-settings' ) . '</p>';
		echo '      <button class="button button-primary" id="tz-packaging-install">' . esc_html__( '安装数据库表', 'tanzanite-settings' ) . '</button>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '包装规则列表', 'tanzanite-settings' ) . '</div>';
		echo '      <button class="button button-primary" id="tz-packaging-create">' . esc_html__( '新增规则', 'tanzanite-settings' ) . '</button>';
		echo '      <div style="overflow:auto;margin-top:16px;">';
		echo '          <table class="widefat fixed striped" id="tz-packaging-table">';
		echo '              <thead><tr>';
		echo '                  <th style="width:20%;">' . esc_html__( '规则名称', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:10%;">' . esc_html__( '包装重量', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:15%;">' . esc_html__( '包装尺寸', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:10%;">' . esc_html__( '最大件数', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:15%;">' . esc_html__( '适用范围', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:10%;">' . esc_html__( '优先级', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:8%;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:12%;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增包装规则', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-packaging-form">';
		echo '          <input type="hidden" id="tz-packaging-id" />';

		// 基本信息
		echo '          <div class="tz-form-grid">';
		echo '              <label>' . esc_html__( '规则名称', 'tanzanite-settings' ) . '<input type="text" id="tz-packaging-name" required placeholder="如：轮组专用包装" /></label>';
		echo '              <label>' . esc_html__( '优先级', 'tanzanite-settings' ) . '<input type="number" id="tz-packaging-priority" value="0" placeholder="数字越大优先级越高" /></label>';
		echo '              <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select id="tz-packaging-active"><option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option><option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option></select></label>';
		echo '          </div>';

		echo '          <label>' . esc_html__( '描述', 'tanzanite-settings' ) . '<textarea id="tz-packaging-description" rows="2" placeholder="规则的详细说明"></textarea></label>';

		// 包装尺寸
		echo '          <div style="margin-top:20px;">';
		echo '              <strong>' . esc_html__( '包装尺寸', 'tanzanite-settings' ) . '</strong>';
		echo '              <p class="description">' . esc_html__( '定义包装盒的重量和尺寸，用于计算运费。', 'tanzanite-settings' ) . '</p>';
		echo '          </div>';
		echo '          <div class="tz-form-grid">';
		echo '              <label>' . esc_html__( '包装重量 (kg)', 'tanzanite-settings' ) . '<input type="number" step="0.001" id="tz-packaging-box-weight" required placeholder="如 0.5" /></label>';
		echo '              <label>' . esc_html__( '长度 (cm)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-packaging-box-length" placeholder="可选" /></label>';
		echo '              <label>' . esc_html__( '宽度 (cm)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-packaging-box-width" placeholder="可选" /></label>';
		echo '              <label>' . esc_html__( '高度 (cm)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-packaging-box-height" placeholder="可选" /></label>';
		echo '          </div>';

		// 限制条件
		echo '          <div style="margin-top:20px;">';
		echo '              <strong>' . esc_html__( '包装限制', 'tanzanite-settings' ) . '</strong>';
		echo '              <p class="description">' . esc_html__( '定义每个包装的最大容量，超出后会自动拆分为多个包裹。', 'tanzanite-settings' ) . '</p>';
		echo '          </div>';
		echo '          <div class="tz-form-grid">';
		echo '              <label>' . esc_html__( '最大件数', 'tanzanite-settings' ) . '<input type="number" id="tz-packaging-max-items" placeholder="留空表示无限制" /></label>';
		echo '              <label>' . esc_html__( '最大承重 (kg)', 'tanzanite-settings' ) . '<input type="number" step="0.001" id="tz-packaging-max-weight" placeholder="留空表示无限制" /></label>';
		echo '          </div>';

		// 适用范围
		echo '          <div style="margin-top:20px;">';
		echo '              <strong>' . esc_html__( '适用范围', 'tanzanite-settings' ) . '</strong>';
		echo '              <p class="description">' . esc_html__( '指定该规则适用于哪些商品。可以按分类、标签或具体商品ID指定。', 'tanzanite-settings' ) . '</p>';
		echo '          </div>';
		echo '          <div id="tz-packaging-applies-list" style="margin-top:12px;"></div>';
		echo '          <div class="tz-form-grid" style="margin-top:12px;">';
		echo '              <label>' . esc_html__( '类型', 'tanzanite-settings' ) . '<select id="tz-packaging-apply-type">';
		echo '                  <option value="category">' . esc_html__( '商品分类', 'tanzanite-settings' ) . '</option>';
		echo '                  <option value="tag">' . esc_html__( '商品标签', 'tanzanite-settings' ) . '</option>';
		echo '                  <option value="product">' . esc_html__( '商品ID', 'tanzanite-settings' ) . '</option>';
		echo '                  <option value="all">' . esc_html__( '所有商品', 'tanzanite-settings' ) . '</option>';
		echo '              </select></label>';
		echo '              <label>' . esc_html__( '值', 'tanzanite-settings' ) . '<input type="text" id="tz-packaging-apply-value" placeholder="分类slug、标签slug或商品ID" /></label>';
		echo '              <label style="align-self:end;"><button type="button" class="button" id="tz-packaging-add-apply">' . esc_html__( '添加', 'tanzanite-settings' ) . '</button></label>';
		echo '          </div>';

		echo '          <div style="margin-top:16px;display:flex;gap:12px;">';
		echo '              <button class="button button-primary" id="tz-packaging-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button" id="tz-packaging-reset" type="button">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '      </form>';
		echo '  </div>';

		echo '</div>';
	}
}
