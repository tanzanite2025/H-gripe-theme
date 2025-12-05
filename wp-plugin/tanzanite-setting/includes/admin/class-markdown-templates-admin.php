<?php
/**
 * Markdown Templates Admin Page
 * 
 * 负责渲染 Markdown 模板设置页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.9
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * Markdown 模板管理类
 */
class Tanzanite_Markdown_Templates_Admin {

	/**
	 * 渲染 Markdown Templates 管理页面
	 */
	public static function render_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		// 处理表单提交
		if ( isset( $_POST['tz_markdown_templates_nonce'] ) && wp_verify_nonce( $_POST['tz_markdown_templates_nonce'], 'tz_save_markdown_templates' ) ) {
			$templates = [
				'basic'      => wp_unslash( $_POST['tz_template_basic'] ?? '' ),
				'spec'       => wp_unslash( $_POST['tz_template_spec'] ?? '' ),
				'after_sale' => wp_unslash( $_POST['tz_template_after_sale'] ?? '' ),
			];
			
			update_option( 'tanzanite_markdown_templates', $templates );
			
			echo '<div class="notice notice-success is-dismissible"><p>' . esc_html__( '模板已保存。', 'tanzanite-settings' ) . '</p></div>';
		}

		// 获取当前模板
		$saved_templates = get_option( 'tanzanite_markdown_templates', [] );
		$default_templates = self::get_default_markdown_templates();
		$templates = wp_parse_args( $saved_templates, $default_templates );

		?>
		<div class="tz-settings-wrapper">
			<div class="tz-settings-header">
				<h1><?php esc_html_e( 'Markdown Templates', 'tanzanite-settings' ); ?></h1>
				<p><?php esc_html_e( '配置商品详情页的 Markdown 模板，这些模板可以在添加/编辑商品时快速插入。', 'tanzanite-settings' ); ?></p>
			</div>

			<form method="post" action="">
				<?php wp_nonce_field( 'tz_save_markdown_templates', 'tz_markdown_templates_nonce' ); ?>

				<div class="tz-settings-section" style="margin-top:20px;">
					<div class="tz-section-title"><?php esc_html_e( '商品亮点模板', 'tanzanite-settings' ); ?></div>
					<div class="tz-section-body">
						<p class="description"><?php esc_html_e( '用于快速插入商品的核心卖点和详情描述。支持图片、标题和富文本格式。', 'tanzanite-settings' ); ?></p>
						<?php
						wp_editor(
							$templates['basic'],
							'tz_template_basic_editor',
							[
								'textarea_name'  => 'tz_template_basic',
								'media_buttons'  => true,
								'textarea_rows'  => 12,
								'editor_height'  => 300,
								'drag_drop_upload' => true,
								'tinymce'        => [
									'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
									'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
								],
								'quicktags'      => true,
							]
						);
						?>
					</div>
				</div>

				<div class="tz-settings-section" style="margin-top:20px;">
					<div class="tz-section-title"><?php esc_html_e( '规格参数模板', 'tanzanite-settings' ); ?></div>
					<div class="tz-section-body">
						<p class="description"><?php esc_html_e( '用于快速插入商品的规格参数表格。支持表格、图片和富文本格式。', 'tanzanite-settings' ); ?></p>
						<?php
						wp_editor(
							$templates['spec'],
							'tz_template_spec_editor',
							[
								'textarea_name'  => 'tz_template_spec',
								'media_buttons'  => true,
								'textarea_rows'  => 12,
								'editor_height'  => 300,
								'drag_drop_upload' => true,
								'tinymce'        => [
									'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
									'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
								],
								'quicktags'      => true,
							]
						);
						?>
					</div>
				</div>

				<div class="tz-settings-section" style="margin-top:20px;">
					<div class="tz-section-title"><?php esc_html_e( '售后说明模板', 'tanzanite-settings' ); ?></div>
					<div class="tz-section-body">
						<p class="description"><?php esc_html_e( '用于快速插入售后保障和服务说明。支持图片、标题和富文本格式。', 'tanzanite-settings' ); ?></p>
						<?php
						wp_editor(
							$templates['after_sale'],
							'tz_template_after_sale_editor',
							[
								'textarea_name'  => 'tz_template_after_sale',
								'media_buttons'  => true,
								'textarea_rows'  => 12,
								'editor_height'  => 300,
								'drag_drop_upload' => true,
								'tinymce'        => [
									'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
									'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
								],
								'quicktags'      => true,
							]
						);
						?>
					</div>
				</div>

				<p class="submit">
					<button type="submit" class="button button-primary"><?php esc_html_e( '保存模板', 'tanzanite-settings' ); ?></button>
					<button type="button" class="button" id="tz-restore-defaults"><?php esc_html_e( '恢复默认', 'tanzanite-settings' ); ?></button>
				</p>
			</form>
		</div>

		<script type="text/javascript">
		document.addEventListener('DOMContentLoaded', function() {
			const restoreBtn = document.getElementById('tz-restore-defaults');
			if (restoreBtn) {
				restoreBtn.addEventListener('click', function() {
					if (confirm(<?php echo wp_json_encode( __( '确定要恢复默认模板吗？这将覆盖当前的自定义内容。', 'tanzanite-settings' ) ); ?>)) {
						const defaultTemplates = <?php echo wp_json_encode( $default_templates ); ?>;
						
						// 更新富文本编辑器内容
						if (typeof tinymce !== 'undefined') {
							const basicEditor = tinymce.get('tz_template_basic_editor');
							const specEditor = tinymce.get('tz_template_spec_editor');
							const afterSaleEditor = tinymce.get('tz_template_after_sale_editor');
							
							if (basicEditor) basicEditor.setContent(defaultTemplates.basic);
							if (specEditor) specEditor.setContent(defaultTemplates.spec);
							if (afterSaleEditor) afterSaleEditor.setContent(defaultTemplates.after_sale);
						}
						
						// 同时更新 textarea（以防编辑器未加载）
						const basicField = document.querySelector('[name="tz_template_basic"]');
						const specField = document.querySelector('[name="tz_template_spec"]');
						const afterSaleField = document.querySelector('[name="tz_template_after_sale"]');
						
						if (basicField) basicField.value = defaultTemplates.basic;
						if (specField) specField.value = defaultTemplates.spec;
						if (afterSaleField) afterSaleField.value = defaultTemplates.after_sale;
					}
				});
			}
		});
		</script>
		<?php
	}

	/**
	 * 获取默认 Markdown 模板
	 */
	private static function get_default_markdown_templates(): array {
		return [
			'basic'      => "# 商品亮点\n- 优选材质，兼顾舒适与耐用\n- 设计贴合日常场景，易于搭配\n- 支持多种配送方式与售后保障\n\n## 详情描述\n请在此补充产品的核心卖点、使用场景与图文信息。",
			'spec'       => "## 规格参数\n| 项目 | 参数 |\n| --- | --- |\n| 材质 | 请输入 |\n| 尺寸 | 请输入 |\n| 重量 | 请输入 |\n| 颜色 | 请输入 |\n\n> 可根据实际情况补充更多行，或删除不适用的字段。",
			'after_sale' => "## 售后与保障\n1. 支持七天无理由退换货，保持商品及包装完好。\n2. 如遇质量问题，请联系客服并提供照片，我们将在 24 小时内响应。\n3. 定制/特殊商品的退换政策，请以页面说明为准。\n\n感谢您的信任与支持！",
		];
	}
}
