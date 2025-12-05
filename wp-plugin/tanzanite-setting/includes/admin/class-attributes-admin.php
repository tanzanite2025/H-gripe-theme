<?php
/**
 * Attributes Admin Page
 * 
 * 负责渲染属性管理页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.13
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 属性管理类
 */
class Tanzanite_Attributes_Admin {

	/**
	 * 渲染属性管理页面
	 */
	public static function render_page() {
		// 直接输出配置和脚本到页面
		$nonce = wp_create_nonce( 'wp_rest' );
		
		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.13';

		// 加载属性管理 JS
		wp_enqueue_media();
		wp_enqueue_script(
			'tz-attributes',
			TANZANITE_PLUGIN_URL . 'assets/js/attributes.js',
			array( 'jquery', 'wp-media' ),
			$version,
			true
		);

		?>
		<script type="text/javascript">
		var TzAttributesConfig = {
			attrUrl: <?php echo wp_json_encode( rest_url( 'tanzanite/v1/attributes' ) ); ?>,
			singleUrl: <?php echo wp_json_encode( rest_url( 'tanzanite/v1/attributes/' ) ); ?>,
			nonce: <?php echo wp_json_encode( $nonce ); ?>
		};
		</script>
		<?php
		echo '<div class="tz-settings-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>Attributes</h1>';
		echo '      <p>' . esc_html__( '管理商品属性组与属性值，支持颜色、图标等多种类型', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';
		echo '  <div id="tz-attr-notice" class="notice" style="display:none;"></div>';
		
		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '属性组列表', 'tanzanite-settings' ) . '</div>';
		echo '      <button type="button" class="button button-primary" id="tz-attr-create">' . esc_html__( '新增属性组', 'tanzanite-settings' ) . '</button>';
		echo '      <div style="overflow:auto;margin-top:16px;"><table class="widefat fixed striped" id="tz-attr-table"><thead><tr><th>' . esc_html__( '名称', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '类型', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '属性值数', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '筛选', 'tanzanite-settings' ) . '</th><th>' . esc_html__( 'SKU', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '状态', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '操作', 'tanzanite-settings' ) . '</th></tr></thead><tbody></tbody></table></div>';
		echo '  </div>';
		
		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增属性组', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-attr-form" onsubmit="return false;"><input type="hidden" id="tz-attr-id" />';
		echo '          <div class="tz-form-grid">';
		echo '              <label>' . esc_html__( '名称', 'tanzanite-settings' ) . '<input type="text" id="tz-attr-name" required /></label>';
		echo '              <label>' . esc_html__( 'Slug', 'tanzanite-settings' ) . '<input type="text" id="tz-attr-slug" placeholder="' . esc_attr__( '自动生成', 'tanzanite-settings' ) . '" /></label>';
		echo '              <label>' . esc_html__( '类型', 'tanzanite-settings' ) . '<select id="tz-attr-type"><option value="select">' . esc_html__( '下拉选择', 'tanzanite-settings' ) . '</option><option value="color">' . esc_html__( '色块', 'tanzanite-settings' ) . '</option><option value="image">' . esc_html__( '图标', 'tanzanite-settings' ) . '</option></select></label>';
		echo '              <label>' . esc_html__( '排序', 'tanzanite-settings' ) . '<input type="number" id="tz-attr-sort" value="0" /></label>';
		echo '          </div>';
		echo '          <div style="margin-top:12px;display:flex;gap:16px;flex-wrap:wrap;">';
		echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-filterable" checked /> ' . esc_html__( '参与前端筛选', 'tanzanite-settings' ) . '</label>';
		echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-sku" checked /> ' . esc_html__( '影响 SKU 组合', 'tanzanite-settings' ) . '</label>';
		echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-stock" /> ' . esc_html__( '影响库存', 'tanzanite-settings' ) . '</label>';
		echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-enabled" checked /> ' . esc_html__( '启用', 'tanzanite-settings' ) . '</label>';
		echo '          </div>';
		echo '          <div style="margin-top:16px;display:flex;gap:12px;"><button class="button button-primary" id="tz-attr-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button><button class="button" id="tz-attr-reset" type="button">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button></div>';
		echo '      </form>';
		echo '  </div>';
		
		echo '  <div class="tz-settings-section" id="tz-attr-values-section" style="display:none;">';
		echo '      <div class="tz-section-title">' . esc_html__( '属性值管理', 'tanzanite-settings' ) . ' - <span id="tz-current-attr-name"></span></div>';
		echo '      <div style="margin-bottom:16px;display:flex;gap:12px;align-items:end;">';
		echo '          <label style="flex:1;">' . esc_html__( '名称', 'tanzanite-settings' ) . '<input type="text" id="tz-value-name" /></label>';
		echo '          <label style="flex:1;">' . esc_html__( 'Slug', 'tanzanite-settings' ) . '<input type="text" id="tz-value-slug" placeholder="' . esc_attr__( '自动生成', 'tanzanite-settings' ) . '" /></label>';
		echo '          <label style="flex:1;" id="tz-value-field-label">' . esc_html__( '值', 'tanzanite-settings' ) . '<input type="text" id="tz-value-value" placeholder="' . esc_attr__( '如: #FF0000', 'tanzanite-settings' ) . '" /></label>';
		echo '          <button class="button button-primary" id="tz-value-add">' . esc_html__( '添加', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '      <table class="widefat fixed striped" id="tz-values-table"><thead><tr><th>' . esc_html__( '名称', 'tanzanite-settings' ) . '</th><th>' . esc_html__( 'Slug', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '预览', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '排序', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '状态', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '操作', 'tanzanite-settings' ) . '</th></tr></thead><tbody></tbody></table>';
		echo '  </div>';
		echo '</div>';
	}
}
