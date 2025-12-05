<?php
/**
 * Payment Admin Page
 * 
 * 负责渲染支付方式管理页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.8
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 支付方式管理类
 */
class Tanzanite_Payment_Admin {

	/**
	 * 渲染支付方式管理页面
	 */
	public static function render_page() {
		wp_enqueue_media();
		$nonce = wp_create_nonce( 'wp_rest' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.8';

		// 加载支付方式管理 JS
		wp_enqueue_script(
			'tz-payment-methods',
			TANZANITE_PLUGIN_URL . 'assets/js/payment-methods.js',
			array( 'jquery', 'wp-media' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-payment-methods',
			'TzPaymentMethodsConfig',
			array(
				'listUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/payment-methods' ) ),
				'singleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/payment-methods/' ) ),
				'nonce'     => $nonce,
				'gatewayFields' => array(
					'paypal' => array( 'client_id', 'client_secret', 'mode', 'webhook_id' ),
					'stripe' => array( 'publishable_key', 'secret_key', 'webhook_secret', 'mode' ),
					'worldfirst' => array( 'merchant_id', 'api_key', 'api_secret', 'mode' ),
					'payoneer' => array( 'program_id', 'api_username', 'api_password', 'mode' ),
				),
			)
		);

		echo '<div class="tz-settings-wrapper tz-payments-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Payment Methods', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '配置前端可用的支付方式，包括手续费、终端可见性与会员等级限制。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-payment-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '支付方式列表', 'tanzanite-settings' ) . '</div>';
		echo '      <button class="button button-primary" id="tz-payment-create">' . esc_html__( '新增支付方式', 'tanzanite-settings' ) . '</button>';
		echo '      <div style="overflow:auto;margin-top:16px;">';
		echo '          <table class="widefat fixed striped" id="tz-payment-table" style="min-width:960px;">';
		echo '              <thead><tr>';
		foreach ( [ 'Code', 'Name', 'Fee', 'Terminals', 'Membership Levels', 'Enabled', 'Actions' ] as $column ) {
			echo '<th>' . esc_html( $column ) . '</th>';
		}
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增支付方式', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-payment-form" class="tz-payment-form">';
		echo '          <input type="hidden" id="tz-payment-id" />';
		echo '          <div class="tz-form-grid">';
		echo '              <label>' . esc_html__( '编码 (code)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-code" required /></label>';
		echo '              <label>' . esc_html__( '名称', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-name" required /></label>';
		echo '              <label>' . esc_html__( '手续费类型', 'tanzanite-settings' ) . '<select id="tz-payment-fee-type"><option value="fixed">' . esc_html__( '固定金额', 'tanzanite-settings' ) . '</option><option value="percentage">' . esc_html__( '百分比', 'tanzanite-settings' ) . '</option></select></label>';
		echo '              <label>' . esc_html__( '手续费数值', 'tanzanite-settings' ) . '<input type="number" step="0.0001" id="tz-payment-fee-value" min="0" value="0" /></label>';
		echo '              <label>' . esc_html__( '排序 (0 最靠前)', 'tanzanite-settings' ) . '<input type="number" id="tz-payment-sort" value="0" /></label>';
		echo '              <label>' . esc_html__( '启用', 'tanzanite-settings' ) . '<select id="tz-payment-enabled"><option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option><option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option></select></label>';
		echo '          </div>';
		echo '          <div style="margin-top:12px;">';
		echo '              <label>' . esc_html__( '图标 URL', 'tanzanite-settings' ) . '</label>';
		echo '              <div style="display:flex;gap:8px;align-items:center;">';
		echo '                  <input type="text" id="tz-payment-icon-url" class="regular-text" placeholder="https://example.com/icon.png" />';
		echo '                  <button type="button" class="button" id="tz-payment-icon-upload">' . esc_html__( '选择图片', 'tanzanite-settings' ) . '</button>';
		echo '              </div>';
		echo '              <div id="tz-payment-icon-preview" style="margin-top:8px;display:none;">';
		echo '                  <img src="" alt="" style="max-width:120px;max-height:60px;border:1px solid #ddd;padding:4px;background:#fff;" />';
		echo '              </div>';
		echo '          </div>';

		echo '          <label>' . esc_html__( '适用终端', 'tanzanite-settings' ) . '<div class="tz-checkbox-list" id="tz-payment-terminals"></div></label>';
		echo '          <label>' . esc_html__( '可见会员等级 (逗号分隔或逐个添加)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-levels" placeholder="gold, platinum" /></label>';
		echo '          <label>' . esc_html__( '支持的货币 (逗号分隔，如 CNY,USD,EUR)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-currencies" placeholder="CNY, USD, EUR" /></label>';
		echo '          <label>' . esc_html__( '默认货币 (必须在支持列表中)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-default-currency" placeholder="CNY" maxlength="10" /></label>';
		echo '          <label>' . esc_html__( '描述', 'tanzanite-settings' ) . '<textarea id="tz-payment-description" rows="3"></textarea></label>';
		
		echo '          <div style="margin-top:24px;padding-top:24px;border-top:1px solid #e5e7eb;">';
		echo '              <h3 style="margin:0 0 16px 0;font-size:14px;font-weight:600;">' . esc_html__( '第三方支付平台对接', 'tanzanite-settings' ) . '</h3>';
		echo '              <div style="margin-bottom:16px;">';
		echo '                  <label>' . esc_html__( '平台类型', 'tanzanite-settings' ) . '<select id="tz-payment-gateway-type" style="width:100%;">';
		echo '                      <option value="">' . esc_html__( '无 (手动处理)', 'tanzanite-settings' ) . '</option>';
		echo '                      <option value="paypal">' . esc_html__( 'PayPal', 'tanzanite-settings' ) . '</option>';
		echo '                      <option value="stripe">' . esc_html__( 'Stripe', 'tanzanite-settings' ) . '</option>';
		echo '                      <option value="worldfirst">' . esc_html__( '万里汇 (WorldFirst)', 'tanzanite-settings' ) . '</option>';
		echo '                      <option value="payoneer">' . esc_html__( '派安盈 (Payoneer)', 'tanzanite-settings' ) . '</option>';
		echo '                  </select></label>';
		echo '              </div>';
		
		// PayPal 配置
		echo '              <div id="tz-gateway-paypal" class="tz-gateway-config" style="display:none;margin-top:16px;">';
		echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">PayPal 配置</h4>';
		echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Client ID', 'tanzanite-settings' ) . '<input type="text" id="tz-paypal-client-id" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Client Secret', 'tanzanite-settings' ) . '<input type="password" id="tz-paypal-client-secret" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-paypal-mode" style="height:30px;"><option value="sandbox">Sandbox (测试)</option><option value="live">Live (生产)</option></select></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Webhook ID', 'tanzanite-settings' ) . '<input type="text" id="tz-paypal-webhook-id" class="regular-text" /></label>';
		echo '                  </div>';
		echo '              </div>';
		
		// Stripe 配置
		echo '              <div id="tz-gateway-stripe" class="tz-gateway-config" style="display:none;margin-top:16px;">';
		echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">Stripe 配置</h4>';
		echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Publishable Key', 'tanzanite-settings' ) . '<input type="text" id="tz-stripe-publishable-key" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Secret Key', 'tanzanite-settings' ) . '<input type="password" id="tz-stripe-secret-key" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Webhook Secret', 'tanzanite-settings' ) . '<input type="password" id="tz-stripe-webhook-secret" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-stripe-mode" style="height:30px;"><option value="test">Test (测试)</option><option value="live">Live (生产)</option></select></label>';
		echo '                  </div>';
		echo '              </div>';
		
		// 万里汇 配置
		echo '              <div id="tz-gateway-worldfirst" class="tz-gateway-config" style="display:none;margin-top:16px;">';
		echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">万里汇 (WorldFirst) 配置</h4>';
		echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Merchant ID', 'tanzanite-settings' ) . '<input type="text" id="tz-worldfirst-merchant-id" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Key', 'tanzanite-settings' ) . '<input type="password" id="tz-worldfirst-api-key" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Secret', 'tanzanite-settings' ) . '<input type="password" id="tz-worldfirst-api-secret" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-worldfirst-mode" style="height:30px;"><option value="sandbox">Sandbox (测试)</option><option value="production">Production (生产)</option></select></label>';
		echo '                  </div>';
		echo '              </div>';
		
		// 派安盈 配置
		echo '              <div id="tz-gateway-payoneer" class="tz-gateway-config" style="display:none;margin-top:16px;">';
		echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">派安盈 (Payoneer) 配置</h4>';
		echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Program ID', 'tanzanite-settings' ) . '<input type="text" id="tz-payoneer-program-id" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Username', 'tanzanite-settings' ) . '<input type="text" id="tz-payoneer-api-username" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Password', 'tanzanite-settings' ) . '<input type="password" id="tz-payoneer-api-password" class="regular-text" /></label>';
		echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-payoneer-mode" style="height:30px;"><option value="sandbox">Sandbox (测试)</option><option value="live">Live (生产)</option></select></label>';
		echo '                  </div>';
		echo '              </div>';
		echo '          </div>';
		
		echo '          <label>' . esc_html__( '自定义设置 (JSON)', 'tanzanite-settings' ) . '<textarea id="tz-payment-settings" rows="4" placeholder="{\"api_key\":\"...\"}"></textarea></label>';
		echo '          <label>' . esc_html__( 'Meta 信息 (JSON，可选)', 'tanzanite-settings' ) . '<textarea id="tz-payment-meta" rows="3"></textarea></label>';

		echo '          <div style="margin-top:16px;display:flex;gap:12px;">';
		echo '              <button class="button button-primary" id="tz-payment-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button" id="tz-payment-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '      </form>';
		echo '  </div>';

		// 添加独立的支付平台切换和媒体选择器脚本
		?>
		<script type="text/javascript">
		(function() {
			console.log('Payment inline script loaded');
			
			var mediaUploader = null;
			
			function toggleGatewayConfig(gatewayType) {
				console.log('Toggling gateway config for:', gatewayType);
				
				// 隐藏所有配置区域
				var configs = document.querySelectorAll('.tz-gateway-config');
				configs.forEach(function(el) {
					el.style.display = 'none';
				});
				
				// 显示选中的配置区域
				if (gatewayType) {
					var configEl = document.getElementById('tz-gateway-' + gatewayType);
					if (configEl) {
						configEl.style.display = 'block';
						console.log('Showing config for:', gatewayType);
					} else {
						console.error('Config element not found for:', gatewayType);
					}
				}
			}
			
			function openMediaUploader(e) {
				e.preventDefault();
				console.log('Opening media uploader');
				
				// 检查 wp.media 是否存在
				if (typeof wp === 'undefined' || typeof wp.media === 'undefined') {
					alert('媒体库未加载，请刷新页面重试');
					console.error('wp.media is not available');
					return;
				}
				
				if (mediaUploader) {
					mediaUploader.open();
					return;
				}
				
				try {
					mediaUploader = wp.media({
						title: '选择支付图标',
						button: { text: '使用此图片' },
						multiple: false
					});
					
					mediaUploader.on('select', function() {
						var attachment = mediaUploader.state().get('selection').first().toJSON();
						var iconUrlInput = document.getElementById('tz-payment-icon-url');
						if (iconUrlInput) {
							iconUrlInput.value = attachment.url;
							updateIconPreview();
						}
					});
					
					mediaUploader.open();
					console.log('Media uploader opened');
				} catch (error) {
					console.error('Failed to open media uploader:', error);
					alert('打开媒体库失败: ' + error.message);
				}
			}
			
			function updateIconPreview() {
				var iconUrlInput = document.getElementById('tz-payment-icon-url');
				var iconPreview = document.getElementById('tz-payment-icon-preview');
				
				if (!iconUrlInput || !iconPreview) return;
				
				var url = iconUrlInput.value;
				if (url) {
					iconPreview.innerHTML = '<img src="' + url + '" style="max-width:120px;max-height:60px;border:1px solid #ddd;padding:4px;background:#fff;" />';
					iconPreview.style.display = 'block';
				} else {
					iconPreview.innerHTML = '<span style="color:#9ca3af;">无图标</span>';
					iconPreview.style.display = 'none';
				}
			}
			
			// 等待 DOM 加载完成
			if (document.readyState === 'loading') {
				document.addEventListener('DOMContentLoaded', init);
			} else {
				init();
			}
			
			function initTerminalsCheckboxes() {
				var terminalsContainer = document.getElementById('tz-payment-terminals');
				if (!terminalsContainer) {
					console.warn('Terminals container not found');
					return;
				}
				
				var terminalOptions = [
					{ value: 'web', label: '网页端' },
					{ value: 'mobile', label: '移动端' },
					{ value: 'app', label: 'APP' },
					{ value: 'wechat', label: '微信小程序' }
				];
				
				terminalsContainer.innerHTML = '';
				
				terminalOptions.forEach(function(option) {
					var label = document.createElement('label');
					label.style.display = 'flex';
					label.style.alignItems = 'center';
					label.style.gap = '8px';
					label.style.marginBottom = '8px';
					
					var checkbox = document.createElement('input');
					checkbox.type = 'checkbox';
					checkbox.value = option.value;
					checkbox.className = 'terminal-checkbox';
					
					var span = document.createElement('span');
					span.textContent = option.label;
					
					label.appendChild(checkbox);
					label.appendChild(span);
					terminalsContainer.appendChild(label);
				});
				
				console.log('Terminals checkboxes initialized');
			}
			
			function init() {
				console.log('Initializing payment page scripts');
				
				// 初始化终端复选框
				initTerminalsCheckboxes();
				
				// 支付平台类型切换
				var gatewayTypeSelect = document.getElementById('tz-payment-gateway-type');
				if (gatewayTypeSelect) {
					console.log('Gateway type select found');
					gatewayTypeSelect.addEventListener('change', function() {
						console.log('Gateway type changed to:', this.value);
						toggleGatewayConfig(this.value);
					});
					
					if (gatewayTypeSelect.value) {
						toggleGatewayConfig(gatewayTypeSelect.value);
					}
				} else {
					console.error('Gateway type select not found!');
				}
				
				// 图标上传按钮
				var iconUploadBtn = document.getElementById('tz-payment-icon-upload');
				if (iconUploadBtn) {
					console.log('Icon upload button found');
					iconUploadBtn.addEventListener('click', openMediaUploader);
				} else {
					console.warn('Icon upload button not found');
				}
				
				// 图标 URL 输入框
				var iconUrlInput = document.getElementById('tz-payment-icon-url');
				if (iconUrlInput) {
					iconUrlInput.addEventListener('input', updateIconPreview);
					updateIconPreview(); // 初始化预览
				}
			}
		})();
		</script>
		<?php

		echo '</div>';
	}
}
