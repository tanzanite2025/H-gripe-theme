<?php
/**
 * Add Product Admin Page
 * 
 * 负责渲染添加商品页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.12
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 添加商品页面管理类
 */
class Tanzanite_Add_Product_Admin {

	/**
	 * 渲染添加商品页面
	 */
	public static function render_page() {
		if ( ! current_user_can( 'tanz_manage_products' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		$product_id = isset( $_GET['product_id'] ) ? absint( $_GET['product_id'] ) : 0;

		wp_enqueue_media();
		wp_enqueue_editor();

		$nonce            = wp_create_nonce( 'wp_rest' );
		$create_endpoint  = esc_url_raw( rest_url( 'tanzanite/v1/products' ) );
		$single_endpoint  = esc_url_raw( rest_url( 'tanzanite/v1/products/' ) );
		$media_endpoint   = esc_url_raw( rest_url( 'wp/v2/media' ) );
		$detail_endpoint  = $product_id ? esc_url_raw( trailingslashit( $single_endpoint ) . $product_id ) : '';
		$seo_endpoint     = esc_url_raw( rest_url( 'mytheme/v1/seo/product/' ) );
		
		// 获取会员等级设置
		$member_tiers = self::get_member_tiers();

		$initial_skus = [];
		if ( $product_id ) {
			$initial_skus = array_map(
				static function ( array $sku ): array {
					$attributes_input = '';
					if ( isset( $sku['attributes'] ) && is_array( $sku['attributes'] ) && ! empty( $sku['attributes'] ) ) {
						$pairs = [];
						foreach ( $sku['attributes'] as $key => $value ) {
							if ( is_array( $value ) ) {
								foreach ( $value as $single ) {
									$pairs[] = sprintf( '%s=%s', $key, $single );
								}
							} elseif ( '' !== (string) $value ) {
								$pairs[] = sprintf( '%s=%s', $key, $value );
							}
						}
						$attributes_input = implode( '; ', $pairs );
					}

					$sku['attributes_input'] = $attributes_input;

					return $sku;
				},
				self::get_product_skus( $product_id )
			);
		}

		// 从数据库读取 Markdown 模板，如果不存在则使用默认值
		$saved_templates = get_option( 'tanzanite_markdown_templates', [] );
		$default_templates = self::get_default_markdown_templates();
		$markdown_templates = wp_parse_args( $saved_templates, $default_templates );

		$markdown_rules = [
			'requiredSections'        => [
				[
					'label'   => __( '商品亮点段落', 'tanzanite-settings' ),
					'keyword' => '商品亮点',
				],
				[
					'label'   => __( '规格参数段落', 'tanzanite-settings' ),
					'keyword' => '规格参数',
				],
				[
					'label'   => __( '售后保障段落', 'tanzanite-settings' ),
					'keyword' => '售后',
				],
			],
			'forbiddenTerms'          => [ '国家级', '顶级保证', '100%治愈', '零风险', '最高标准' ],
			'placeholderTokens'       => [ 'TODO', '待补', '待完善', '占位', '示意图', '未定稿', '[图片占位]' ],
			'requireImages'           => 1,
			'imagePlaceholderKeywords'=> [ 'placeholder', 'temp', 'todo', 'example.com', 'dummy' ],
		];

		echo '<div class="tz-settings-wrapper tz-product-editor">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Add New Product', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '按照商品规划完善基础信息、定价、库存与物流配置，保存后即可在 Nuxt 前端展示。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-product-editor-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <form id="tz-product-editor-form" class="tz-product-editor-form" autocomplete="off">';

		echo '      <section class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( '基础信息', 'tanzanite-settings' ) . '</div>';
		echo '          <div class="tz-section-body" style="display:grid;gap:12px;">';
		echo '              <label>' . esc_html__( '商品标题', 'tanzanite-settings' ) . '<input type="text" id="tz-product-title" class="regular-text" required /></label>';
		echo '              <label>' . esc_html__( '副标题 / 摘要', 'tanzanite-settings' ) . '<textarea id="tz-product-excerpt" rows="3" class="widefat"></textarea></label>';
		echo '              <label>' . esc_html__( 'Slug（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-product-slug" class="regular-text" /></label>';
		echo '              <div>';
		echo '                  <label>' . esc_html__( 'URL 自定义路径（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-product-urllink-path" class="regular-text" placeholder="例如：products/category/product-name" /></label>';
		echo '                  <p class="description" style="margin:4px 0 0 0;">' . esc_html__( '自定义此商品的完整 URL 路径。留空则使用默认的 WordPress 固定链接。示例：products/electronics/phone', 'tanzanite-settings' ) . '</p>';
		echo '              </div>';
		echo '          </div>';
		echo '      </section>';

		echo '      <section class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( '详情内容', 'tanzanite-settings' ) . '</div>';
		echo '          <p class="description">' . esc_html__( '支持标题、段落、图片与富文本排版，内容将同步输出至 Nuxt 前端。', 'tanzanite-settings' ) . '</p>';

		ob_start();
		wp_editor(
			'',
			'tz-product-content',
			[
				'textarea_name'  => 'tz-product-content',
				'media_buttons'  => true,
				'textarea_rows'  => 18,
				'editor_height'  => 360,
				'drag_drop_upload' => true,
				'tinymce'        => [
					'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
					'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
				],
				'quicktags'      => true,
			]
		);
		$editor_html = ob_get_clean();
		echo $editor_html;

		echo '          <div class="tz-markdown-toolbar" style="display:flex;flex-wrap:wrap;gap:12px;margin-top:12px;align-items:center;background:#f9f9f9;padding:12px;border:1px solid #ddd;border-radius:4px;">';
		echo '              <label style="display:flex;align-items:center;gap:6px;font-weight:500;color:#23282d;font-size:13px;">';
		echo '                  <input type="checkbox" id="tz-product-content-toggle" style="margin:0;width:16px;height:16px;flex-shrink:0;" /> ' . esc_html__( '启用 Markdown 模式（左侧编辑、右侧预览）', 'tanzanite-settings' );
		echo '              </label>';
		echo '              <span class="description" style="color:#666;font-size:12px;">' . esc_html__( '可选用下方模板快速填充内容，Markdown 将即时同步至富文本编辑器。', 'tanzanite-settings' ) . '</span>';
		echo '          </div>';

		echo '          <div class="tz-markdown-templates" style="display:flex;flex-wrap:wrap;gap:8px;margin-top:8px;">';
		echo '              <button type="button" class="button" data-markdown-template="basic">' . esc_html__( '插入「商品亮点」模板', 'tanzanite-settings' ) . '</button>';
		echo '              <button type="button" class="button" data-markdown-template="spec">' . esc_html__( '插入「规格参数」模板', 'tanzanite-settings' ) . '</button>';
		echo '              <button type="button" class="button" data-markdown-template="after_sale">' . esc_html__( '插入「售后说明」模板', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';

		echo '          <div id="tz-product-content-markdown-wrapper" class="tz-markdown-wrapper" style="display:none;grid-template-columns:1fr 1fr;gap:16px;margin-top:16px;">';
		echo '              <div class="tz-markdown-column" style="display:flex;flex-direction:column;gap:8px;">';
		echo '                  <label for="tz-product-content-markdown" style="font-weight:600;">' . esc_html__( 'Markdown 输入', 'tanzanite-settings' ) . '</label>';
		echo '                  <textarea id="tz-product-content-markdown" class="widefat" style="min-height:360px;font-family:SFMono-Regular,Menlo,Monaco,Consolas,\'Courier New\',monospace;"></textarea>';
		echo '              </div>';
		echo '              <div class="tz-markdown-column" style="display:flex;flex-direction:column;gap:8px;">';
		echo '                  <label style="font-weight:600;">' . esc_html__( '实时预览', 'tanzanite-settings' ) . '</label>';
		echo '                  <div id="tz-product-content-preview" class="tz-markdown-preview" style="min-height:360px;border:1px solid #dcdfe5;border-radius:6px;padding:12px;background:#fff;overflow:auto;"></div>';
		echo '              </div>';
		echo '          </div>';

		echo '      </section>';

		echo '      <section class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( '媒体资源', 'tanzanite-settings' ) . '</div>';
		echo '          <div class="tz-section-body" style="display:grid;gap:20px;">';
		echo '              <div class="tz-media-block">';
		echo '                  <div class="tz-media-block__header">' . esc_html__( '主图', 'tanzanite-settings' ) . '</div>';
		echo '                  <div id="tz-product-featured-preview" class="tz-media-preview"></div>';
		echo '                  <div class="tz-media-actions" style="display:flex;gap:8px;flex-wrap:wrap;margin:12px 0;">';
		echo '                      <button type="button" class="button" id="tz-product-featured-select">' . esc_html__( '选择图片', 'tanzanite-settings' ) . '</button>';
		echo '                      <button type="button" class="button-link" id="tz-product-featured-clear">' . esc_html__( '清除', 'tanzanite-settings' ) . '</button>';
		echo '                  </div>';
		echo '                  <input type="hidden" id="tz-product-featured-id" />';
		echo '                  <label>' . esc_html__( '主图 URL（可选，覆盖上传结果）', 'tanzanite-settings' ) . '<input type="url" id="tz-product-featured-url" class="regular-text" placeholder="https://" /></label>';
		echo '              </div>';
		echo '              <div class="tz-media-block">';
		echo '                  <div class="tz-media-block__header">' . esc_html__( '图库', 'tanzanite-settings' ) . '</div>';
		echo '                  <div id="tz-product-gallery-preview" class="tz-media-gallery" style="display:flex;gap:12px;flex-wrap:wrap;"></div>';
		echo '                  <div class="tz-media-actions" style="display:flex;gap:8px;flex-wrap:wrap;margin:12px 0;">';
		echo '                      <button type="button" class="button" id="tz-product-gallery-select">' . esc_html__( '选择图片', 'tanzanite-settings' ) . '</button>';
		echo '                      <button type="button" class="button-link" id="tz-product-gallery-clear">' . esc_html__( '清空', 'tanzanite-settings' ) . '</button>';
		echo '                  </div>';
		echo '                  <input type="hidden" id="tz-product-gallery-ids" />';
		echo '                  <label>' . esc_html__( '外链图片 URL（每行一条，可选）', 'tanzanite-settings' ) . '<textarea id="tz-product-gallery-externals" rows="3" class="widefat" placeholder="https://example.com/image.jpg"></textarea></label>';
		echo '              </div>';
		echo '              <div class="tz-media-block">';
		echo '                  <div class="tz-media-block__header">' . esc_html__( '主图视频', 'tanzanite-settings' ) . '</div>';
		echo '                  <div id="tz-product-video-preview" class="tz-media-preview"></div>';
		echo '                  <div class="tz-media-actions" style="display:flex;gap:8px;flex-wrap:wrap;margin:12px 0;">';
		echo '                      <button type="button" class="button" id="tz-product-video-select">' . esc_html__( '选择视频', 'tanzanite-settings' ) . '</button>';
		echo '                      <button type="button" class="button-link" id="tz-product-video-clear">' . esc_html__( '清除', 'tanzanite-settings' ) . '</button>';
		echo '                  </div>';
		echo '                  <input type="hidden" id="tz-product-video-id" />';
		echo '                  <label>' . esc_html__( '视频 URL（可选，覆盖上传结果）', 'tanzanite-settings' ) . '<input type="url" id="tz-product-video-url" class="regular-text" placeholder="https://" /></label>';
		echo '              </div>';
		echo '          </div>';
		echo '      </section>';

		echo '      <section class="tz-settings-grid" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(340px,1fr));gap:16px;">';

		echo '          <div class="tz-settings-section">';
		echo '              <div class="tz-section-title">' . esc_html__( '价格与活动', 'tanzanite-settings' ) . '</div>';
		echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
		echo '                  <label>' . esc_html__( '原价 (USD)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-price-regular" class="regular-text" placeholder="0.00" /></label>';
		echo '                  <label>' . esc_html__( '现价 (USD)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-price-sale" class="regular-text" placeholder="0.00" /></label>';
		echo '                  <label>' . esc_html__( '会员价 (USD)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-price-member" class="regular-text" placeholder="0.00" /></label>';
		echo '                  <label>' . esc_html__( '限购数量', 'tanzanite-settings' ) . '<input type="number" id="tz-product-limit" class="regular-text" /></label>';
		echo '                  <label>' . esc_html__( '起购数量', 'tanzanite-settings' ) . '<input type="number" id="tz-product-min-purchase" class="regular-text" /></label>';
		echo '                  <div class="tz-tier-pricing" style="display:flex;flex-direction:column;gap:12px;">';
		echo '                      <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:8px;">';
		echo '                          <strong>' . esc_html__( '阶梯价配置', 'tanzanite-settings' ) . '</strong>';
		echo '                          <button type="button" class="button" id="tz-tier-template">' . esc_html__( '插入示例阶梯', 'tanzanite-settings' ) . '</button>';
		echo '                      </div>';
		echo '                      <p class="description">' . esc_html__( '按购买数量区间设置折扣或特价。示例：10 件起 95 折，50 件起 9 折。可留空表示不启用阶梯价。', 'tanzanite-settings' ) . '</p>';
		echo '                      <table class="widefat striped" id="tz-tier-pricing-table" style="margin:0;">';
		echo '                          <thead><tr>';
		echo '                              <th style="width:20%;">' . esc_html__( '最小数量', 'tanzanite-settings' ) . '</th>';
		echo '                              <th style="width:20%;">' . esc_html__( '最大数量（可选）', 'tanzanite-settings' ) . '</th>';
		echo '                              <th style="width:20%;">' . esc_html__( '单价 / 折扣', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '说明', 'tanzanite-settings' ) . '</th>';
		echo '                              <th style="width:120px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
		echo '                          </tr></thead>';
		echo '                          <tbody></tbody>';
		echo '                      </table>';
		echo '                      <div id="tz-tier-empty" class="notice notice-info" style="margin:0;">' . esc_html__( '尚未配置阶梯价。', 'tanzanite-settings' ) . '</div>';
		echo '                      <input type="hidden" id="tz-product-tier-pricing" />';
		echo '                      <form id="tz-tier-form" style="display:grid;gap:12px;padding:12px;border:1px solid #dcdfe5;border-radius:8px;background:#f9fafb;">';
		echo '                          <input type="hidden" id="tz-tier-index" value="" />';
		echo '                          <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px;">';
		echo '                              <label>' . esc_html__( '最小数量', 'tanzanite-settings' ) . '<input type="number" min="1" step="1" id="tz-tier-min" class="regular-text" required /></label>';
		echo '                              <label>' . esc_html__( '最大数量（选填）', 'tanzanite-settings' ) . '<input type="number" min="1" step="1" id="tz-tier-max" class="regular-text" /></label>';
		echo '                              <label>' . esc_html__( '单价 / 折扣', 'tanzanite-settings' ) . '<input type="number" min="0" step="0.01" id="tz-tier-price" class="regular-text" required /></label>';
		echo '                              <label>' . esc_html__( '备注（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-tier-note" class="regular-text" placeholder="' . esc_attr__( '例如：95 折', 'tanzanite-settings' ) . '" /></label>';
		echo '                          </div>';
		echo '                          <div style="display:flex;gap:8px;flex-wrap:wrap;">';
		echo '                              <button type="button" class="button button-primary" id="tz-tier-submit">' . esc_html__( '保存阶梯价', 'tanzanite-settings' ) . '</button>';
		echo '                              <button type="button" class="button" id="tz-tier-reset">' . esc_html__( '清空表单', 'tanzanite-settings' ) . '</button>';
		echo '                          </div>';
		echo '                      </form>';
		echo '                  </div>';
		echo '              </div>';
		echo '          </div>';

		echo '          <div class="tz-settings-section">';
		echo '              <div class="tz-section-title">' . esc_html__( '库存与规格', 'tanzanite-settings' ) . '</div>';
		echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
		echo '                  <label>' . esc_html__( '总库存', 'tanzanite-settings' ) . '<input type="number" id="tz-product-stock" class="regular-text" /></label>';
		echo '                  <label>' . esc_html__( '库存预警值', 'tanzanite-settings' ) . '<input type="number" id="tz-product-stock-alert" class="regular-text" /></label>';
		echo '                  <label><input type="checkbox" id="tz-product-backorders" /> ' . esc_html__( '允许超卖 / 接受缺货订单', 'tanzanite-settings' ) . '</label>';
		
		echo '                  <div class="tz-product-attributes-selector" style="margin:16px 0;padding:16px;background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;">';
		echo '                      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">';
		echo '                          <strong>' . esc_html__( '商品属性选择', 'tanzanite-settings' ) . '</strong>';
		echo '                          <button type="button" class="button button-primary" id="tz-generate-skus" style="display:none;">' . esc_html__( '自动生成 SKU 组合', 'tanzanite-settings' ) . '</button>';
		echo '                      </div>';
		echo '                      <p class="description" style="margin:0 0 12px 0;">' . esc_html__( '从属性模板中选择影响 SKU 的属性，勾选需要的属性值，然后点击"自动生成 SKU 组合"。', 'tanzanite-settings' ) . '</p>';
		echo '                      <div id="tz-attributes-list" class="tz-attributes-list" style="display:flex;flex-direction:column;gap:16px;"></div>';
		echo '                      <div id="tz-attributes-loading" style="padding:20px;text-align:center;color:#6b7280;">' . esc_html__( '正在加载属性模板...', 'tanzanite-settings' ) . '</div>';
		echo '                      <div id="tz-attributes-empty" style="display:none;padding:20px;text-align:center;color:#6b7280;">' . esc_html__( '暂无可用属性。请先在「Attributes」页面创建属性并标记为 "影响 SKU"。', 'tanzanite-settings' ) . '</div>';
		echo '                  </div>';
		
		echo '                  <div class="tz-sku-editor" id="tz-product-sku-editor" style="display:flex;flex-direction:column;gap:12px;">';
		echo '                      <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:8px;">';
		echo '                          <strong>' . esc_html__( 'SKU 组合与价格', 'tanzanite-settings' ) . '</strong>';
		echo '                          <span class="description" id="tz-product-sku-hint">' . esc_html__( '可从上方属性自动生成，或手动输入。属性格式：颜色=蓝;尺寸=16', 'tanzanite-settings' ) . '</span>';
		echo '                      </div>';
		echo '                      <table class="widefat fixed striped" id="tz-product-sku-table">';
		echo '                          <thead><tr>';
		echo '                              <th>' . esc_html__( 'SKU 编码', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '属性组合', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '原价', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '现价', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '库存', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '条码', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '重量(kg)', 'tanzanite-settings' ) . '</th>';
		echo '                              <th>' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
		echo '                          </tr></thead>';
		echo '                          <tbody></tbody>';
		echo '                      </table>';
		echo '                      <div id="tz-product-sku-empty" class="notice notice-info" style="margin:0;">';
		echo '                          <p>' . esc_html__( '尚未添加 SKU，请在下方表单中录入。', 'tanzanite-settings' ) . '</p>';
		echo '                      </div>';
		echo '                      <form id="tz-product-sku-form" class="tz-sku-form" style="display:grid;gap:12px;padding:12px;border:1px solid #dcdfe5;border-radius:8px;background:#f9fafb;">';
		echo '                          <input type="hidden" id="tz-product-sku-form-index" value="" />';
		echo '                          <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:12px;">';
		echo '                              <label>' . esc_html__( 'SKU 编码', 'tanzanite-settings' ) . '<input type="text" id="tz-product-sku-form-code" class="regular-text" required /></label>';
		echo '                              <label>' . esc_html__( '属性组合', 'tanzanite-settings' ) . '<input type="text" id="tz-product-sku-form-attrs" class="regular-text" placeholder="颜色=蓝;尺寸=16" /></label>';
		echo '                              <label>' . esc_html__( '原价 (USD)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-sku-form-price-regular" class="regular-text" placeholder="0.00" /></label>';
		echo '                              <label>' . esc_html__( '现价 (USD)', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-sku-form-price-sale" class="regular-text" placeholder="0.00" /></label>';
		echo '                              <label>' . esc_html__( '库存', 'tanzanite-settings' ) . '<input type="number" id="tz-product-sku-form-stock" class="regular-text" /></label>';
		echo '                              <label>' . esc_html__( '条码', 'tanzanite-settings' ) . '<input type="text" id="tz-product-sku-form-barcode" class="regular-text" /></label>';
		echo '                              <label>' . esc_html__( '重量 (kg)', 'tanzanite-settings' ) . '<input type="number" step="0.001" id="tz-product-sku-form-weight" class="regular-text" placeholder="0.000" /></label>';
		echo '                          </div>';
		echo '                          <div style="display:flex;gap:8px;flex-wrap:wrap;">';
		echo '                              <button type="button" class="button button-primary" id="tz-product-sku-form-submit">' . esc_html__( '保存 / 新增 SKU', 'tanzanite-settings' ) . '</button>';
		echo '                              <button type="button" class="button" id="tz-product-sku-form-reset">' . esc_html__( '清空表单', 'tanzanite-settings' ) . '</button>';
		echo '                          </div>';
		echo '                      </form>';
		echo '                  </div>';
		echo '              </div>';
		echo '          </div>';
		echo '      </section>';
		echo '          <div class="tz-settings-section">';
		echo '              <div class="tz-section-title">' . esc_html__( '物流与配送', 'tanzanite-settings' ) . '</div>';
		echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
		echo '                  <label>' . esc_html__( '配送模板', 'tanzanite-settings' ) . '<select id="tz-product-shipping-template" class="widefat" multiple size="6"></select></label>';
		echo '                  <label><input type="checkbox" id="tz-product-free-shipping" /> ' . esc_html__( '是否包邮', 'tanzanite-settings' ) . '</label>';
		echo '                  <label>' . esc_html__( '发货时效描述', 'tanzanite-settings' ) . '<input type="text" id="tz-product-shipping-time" class="regular-text" /></label>';
		echo '                  <textarea id="tz-product-logistics-tags" rows="2" class="widefat" placeholder="' . esc_attr__( '跨境 / 冷链标签等', 'tanzanite-settings' ) . '"></textarea>';
		echo '              </div>';
		echo '          </div>';

		echo '          <div class="tz-settings-section">';
		echo '              <div class="tz-section-title">' . esc_html__( '税率设置', 'tanzanite-settings' ) . '</div>';
		echo '              <p class="description">' . esc_html__( '勾选适用于此商品的税率模板，前端下单时将自动计算税费。', 'tanzanite-settings' ) . '</p>';
		echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
		echo '                  <div id="tz-product-tax-rates-list" class="tz-checkbox-list" style="display:grid;gap:8px;"></div>';
		echo '                  <p class="description" style="color:#646970;font-size:12px;">' . esc_html__( '如无可用税率模板，请先在「税率管理」页面创建。', 'tanzanite-settings' ) . '</p>';
		echo '              </div>';
		echo '          </div>';

		echo '          <div class="tz-settings-section">';
		echo '              <div class="tz-section-title">' . esc_html__( '发布与频道', 'tanzanite-settings' ) . '</div>';
		echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
		echo '                  <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select id="tz-product-status" class="widefat"><option value="draft">' . esc_html__( '草稿', 'tanzanite-settings' ) . '</option><option value="publish">' . esc_html__( '发布', 'tanzanite-settings' ) . '</option><option value="pending">' . esc_html__( '待审', 'tanzanite-settings' ) . '</option></select></label>';
		echo '                  <label>' . esc_html__( '发布时间', 'tanzanite-settings' ) . '<input type="datetime-local" id="tz-product-date" class="regular-text" /></label>';
		echo '                  <label><input type="checkbox" id="tz-product-sticky" /> ' . esc_html__( '置顶显示', 'tanzanite-settings' ) . '</label>';
		echo '                  <textarea id="tz-product-channels" rows="2" class="widefat" placeholder="' . esc_attr__( '专题频道 ID/别名（逗号或 JSON）', 'tanzanite-settings' ) . '"></textarea>';
		echo '              </div>';
		echo '          </div>';

		echo '      </section>';

		echo '      <section class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( 'SEO 优化', 'tanzanite-settings' ) . '</div>';
		echo '          <p class="description">' . esc_html__( '填写以下元信息以提升搜索引擎收录表现，保存商品后将自动同步至 SEO 插件。', 'tanzanite-settings' ) . '</p>';
		echo '          <div id="tz-product-seo-status" class="tz-seo-status" style="display:flex;align-items:center;gap:8px;"></div>';
		echo '          <div id="tz-product-seo-loading" style="display:none;">' . esc_html__( '正在加载 SEO 数据…', 'tanzanite-settings' ) . '</div>';
		echo '          <div class="tz-section-body" style="display:grid;gap:12px;">';
		echo '              <label>' . esc_html__( 'SEO 标题', 'tanzanite-settings' ) . '<input type="text" id="tz-product-seo-title" class="regular-text" placeholder="' . esc_attr__( '建议 30-60 个字符', 'tanzanite-settings' ) . '" /></label>';
		echo '              <label>' . esc_html__( 'SEO 描述', 'tanzanite-settings' ) . '<textarea id="tz-product-seo-description" rows="3" class="widefat" placeholder="' . esc_attr__( '建议 90-160 个字符，以自然语句概述商品亮点。', 'tanzanite-settings' ) . '"></textarea></label>';
		echo '              <label>' . esc_html__( 'SEO 关键字（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-product-seo-keywords" class="regular-text" placeholder="' . esc_attr__( '逗号分隔，如：戒指, 蓝宝石, 新品', 'tanzanite-settings' ) . '" /></label>';
		echo '          </div>';
		echo '          <div id="tz-product-seo-actions" style="display:none;align-items:center;gap:8px;">';
		echo '              <button type="button" class="button" id="tz-product-seo-refresh">' . esc_html__( '重新获取 SEO 数据', 'tanzanite-settings' ) . '</button>';
		echo '              <span class="description" id="tz-product-seo-hint"></span>';
		echo '          </div>';
		echo '          <input type="hidden" id="tz-product-seo-locale" value="" />';
		echo '      </section>';

		echo '      <section class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( 'API 输出预览', 'tanzanite-settings' ) . '</div>';
		echo '          <p class="description">' . esc_html__( '保存后将在此展示 REST 返回的结构，便于与 Nuxt 前端联调。', 'tanzanite-settings' ) . '</p>';
		echo '          <pre id="tz-product-preview" style="background:#f6f7f7;border:1px solid #ccd0d4;padding:12px;min-height:160px;overflow:auto;"></pre>';
		echo '      </section>';

		echo '      <div class="tz-product-editor-actions" style="display:flex;gap:12px;margin-top:24px;">';
		echo '          <button type="submit" class="button button-primary" id="tz-product-submit">' . esc_html__( '保存商品', 'tanzanite-settings' ) . '</button>';
		echo '          <button type="button" class="button" id="tz-product-save-draft">' . esc_html__( '保存为草稿', 'tanzanite-settings' ) . '</button>';
		echo '          <button type="button" class="button" id="tz-product-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';

		echo '  </form>';
		echo '</div>';

		$initial_meta = $product_id ? self::get_product_meta_payload( $product_id ) : [];
		$initial_payload = [];
		if ( $product_id ) {
			$initial_payload = self::get_initial_payload( $product_id );
		}

		$initial_shipping_template_ids = array();
		if ( $product_id ) {
			$stored_ids = get_post_meta( $product_id, '_tanz_shipping_template_ids', true );
			if ( is_array( $stored_ids ) ) {
				$initial_shipping_template_ids = array_values( array_filter( array_map( 'absint', $stored_ids ) ) );
			}

			if ( empty( $initial_shipping_template_ids ) ) {
				$single_id = (int) get_post_meta( $product_id, '_tanz_shipping_template_id', true );
				if ( $single_id > 0 ) {
					$initial_shipping_template_ids = array( $single_id );
				}
			}
		}

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.12';

		// 加载 Add Product JS
		wp_enqueue_script(
			'tz-product-editor',
			TANZANITE_PLUGIN_URL . 'assets/js/product-editor.js',
			array( 'jquery', 'wp-media', 'wp-editor', 'tz-admin-common' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-product-editor',
			'TzProductEditorConfig',
			array(
				'nonce'             => $nonce,
				'createEndpoint'    => $create_endpoint,
				'singleEndpoint'    => $single_endpoint,
				'mediaEndpoint'     => $media_endpoint,
				'detailEndpoint'    => $detail_endpoint,
				'seoEndpoint'       => $seo_endpoint,
				'productId'         => $product_id,
				'initialSkuRows'    => $initial_skus,
				'markdownRules'     => $markdown_rules,
				'markdownTemplates' => $markdown_templates,
				'initialMeta'       => $initial_meta,
				'initialPayload'    => $initial_payload,
				'initialShippingTemplateIds' => $initial_shipping_template_ids,
				'membershipTiers'   => $member_tiers,
				'taxRatesEndpoint'  => esc_url_raw( rest_url( 'tanzanite/v1/tax-rates' ) ),
				'shippingTemplatesEndpoint' => esc_url_raw( rest_url( 'tanzanite/v1/shipping-templates' ) ),
				'attributesEndpoint' => esc_url_raw( rest_url( 'tanzanite/v1/attributes' ) ),
				'strings'           => array(
					'saveSuccess'       => __( '商品已保存。', 'tanzanite-settings' ),
					'saveFailed'        => __( '保存失败，请稍后重试。', 'tanzanite-settings' ),
					'draftSuccess'      => __( '草稿已保存。', 'tanzanite-settings' ),
					'resetConfirm'      => __( '确定要重置表单吗？未保存的更改将丢失。', 'tanzanite-settings' ),
					'skuDuplicate'      => __( '检测到重复的 SKU 属性组合。', 'tanzanite-settings' ),
					'skuCodeRequired'   => __( '请填写 SKU 编码。', 'tanzanite-settings' ),
					'skuPriceRequired'  => __( '请填写 SKU 价格。', 'tanzanite-settings' ),
					'mediaSelectTitle'  => __( '选择图片', 'tanzanite-settings' ),
					'mediaSelectBtn'    => __( '使用此图片', 'tanzanite-settings' ),
					'loadingSeo'        => __( '正在加载 SEO 数据...', 'tanzanite-settings' ),
					'seoLoaded'         => __( 'SEO 数据已加载。', 'tanzanite-settings' ),
					'seoFailed'         => __( 'SEO 数据加载失败。', 'tanzanite-settings' ),
					'generatingSkus'    => __( '正在生成 SKU...', 'tanzanite-settings' ),
					'skusGenerated'     => __( '已生成 %d 个 SKU 组合。', 'tanzanite-settings' ),
					'noAttributes'      => __( '请先选择至少一个属性值。', 'tanzanite-settings' ),
				),
				'i18n'              => array(
					'saveSuccess'       => __( '商品已保存。', 'tanzanite-settings' ),
					'saveFailed'        => __( '保存失败，请稍后重试。', 'tanzanite-settings' ),
					'draftSuccess'      => __( '草稿已保存。', 'tanzanite-settings' ),
					'resetConfirm'      => __( '确定要重置表单吗？未保存的更改将丢失。', 'tanzanite-settings' ),
					'skuDuplicate'      => __( '检测到重复的 SKU 属性组合。', 'tanzanite-settings' ),
					'skuCodeRequired'   => __( '请填写 SKU 编码。', 'tanzanite-settings' ),
					'skuPriceRequired'  => __( '请填写 SKU 价格。', 'tanzanite-settings' ),
					'mediaSelectTitle'  => __( '选择图片', 'tanzanite-settings' ),
					'mediaSelectBtn'    => __( '使用此图片', 'tanzanite-settings' ),
					'loadingSeo'        => __( '正在加载 SEO 数据...', 'tanzanite-settings' ),
					'seoLoaded'         => __( 'SEO 数据已加载。', 'tanzanite-settings' ),
					'seoFailed'         => __( 'SEO 数据加载失败。', 'tanzanite-settings' ),
					'generatingSkus'    => __( '正在生成 SKU...', 'tanzanite-settings' ),
					'skusGenerated'     => __( '已生成 %d 个 SKU 组合。', 'tanzanite-settings' ),
					'noAttributes'      => __( '请先选择至少一个属性值。', 'tanzanite-settings' ),
				),
			)
		);
	}

	/**
	 * 获取会员等级
	 */
	private static function get_member_tiers(): array {
		$loyalty_settings = get_option( 'tanzanite_loyalty_config', [] );
		if ( is_string( $loyalty_settings ) ) {
			$loyalty_settings = json_decode( $loyalty_settings, true );
		}
		
		$member_tiers = [];

		if ( ! empty( $loyalty_settings['tiers'] ) && is_array( $loyalty_settings['tiers'] ) ) {
			foreach ( $loyalty_settings['tiers'] as $tier ) {
				if ( ! empty( $tier['name'] ) ) {
					$member_tiers[] = [
						'code'  => sanitize_key( $tier['name'] ),
						'label' => sanitize_text_field( $tier['label'] ?? $tier['name'] ),
					];
				}
			}
		}
		return $member_tiers;
	}

	/**
	 * 获取商品 SKU
	 */
	private static function get_product_skus( $product_id ): array {
		global $wpdb;
		$table_name = $wpdb->prefix . 'tanz_product_skus';
		
		// 检查表是否存在
		if ( $wpdb->get_var( "SHOW TABLES LIKE '$table_name'" ) !== $table_name ) {
			return [];
		}

		$results = $wpdb->get_results(
			$wpdb->prepare( "SELECT * FROM $table_name WHERE product_id = %d ORDER BY id ASC", $product_id ),
			ARRAY_A
		);

		if ( empty( $results ) ) {
			return [];
		}

		return array_map( function( $row ) {
			// 解码属性 JSON
			if ( ! empty( $row['attributes'] ) && is_string( $row['attributes'] ) ) {
				$row['attributes'] = json_decode( $row['attributes'], true );
			}
			return $row;
		}, $results );
	}

	/**
	 * 获取商品 Meta Payload
	 */
	private static function get_product_meta_payload( $product_id ): array {
		$meta = get_post_meta( $product_id, 'tanz_product_meta', true );
		if ( ! empty( $meta ) && is_string( $meta ) ) {
			return json_decode( $meta, true ) ?: [];
		}
		return is_array( $meta ) ? $meta : [];
	}

	/**
	 * 获取商品初始 Payload
	 * (用于回填表单)
	 */
	private static function get_initial_payload( $product_id ): array {
		$post = get_post( $product_id );
		if ( ! $post ) {
			return [];
		}

		$meta = self::get_product_meta_payload( $product_id );
		$skus = self::get_product_skus( $product_id );

		// 获取分类
		$categories = wp_get_post_terms( $product_id, 'product_cat', [ 'fields' => 'ids' ] );
		
		// 获取标签
		$tags = wp_get_post_terms( $product_id, 'product_tag', [ 'fields' => 'names' ] );

		// 构造 payload
		$payload = [
			'title'         => $post->post_title,
			'content'       => $post->post_content,
			'excerpt'       => $post->post_excerpt,
			'slug'          => $post->post_name,
			'status'        => $post->post_status,
			'date'          => $post->post_date,
			'price_regular' => get_post_meta( $product_id, 'price_regular', true ),
			'price_sale'    => get_post_meta( $product_id, 'price_sale', true ),
			'price_member'  => get_post_meta( $product_id, 'price_member', true ),
			'stock_qty'     => get_post_meta( $product_id, 'stock_qty', true ),
			'sku'           => get_post_meta( $product_id, 'sku', true ),
			'barcode'       => get_post_meta( $product_id, 'barcode', true ),
			'weight'        => get_post_meta( $product_id, 'weight', true ),
			'categories'    => $categories,
			'tags'          => $tags,
			'meta'          => $meta,
			'skus'          => $skus,
		];

		return $payload;
	}

	/**
	 * 获取默认 Markdown 模板
	 * (复用之前的逻辑)
	 */
	private static function get_default_markdown_templates(): array {
		return [
			'basic'      => "# 商品亮点\n- 优选材质，兼顾舒适与耐用\n- 设计贴合日常场景，易于搭配\n- 支持多种配送方式与售后保障\n\n## 详情描述\n请在此补充产品的核心卖点、使用场景与图文信息。",
			'spec'       => "## 规格参数\n| 项目 | 参数 |\n| --- | --- |\n| 材质 | 请输入 |\n| 尺寸 | 请输入 |\n| 重量 | 请输入 |\n| 颜色 | 请输入 |\n\n> 可根据实际情况补充更多行，或删除不适用的字段。",
			'after_sale' => "## 售后与保障\n1. 支持七天无理由退换货，保持商品及包装完好。\n2. 如遇质量问题，请联系客服并提供照片，我们将在 24 小时内响应。\n3. 定制/特殊商品的退换政策，请以页面说明为准。\n\n感谢您的信任与支持！",
		];
	}
}
