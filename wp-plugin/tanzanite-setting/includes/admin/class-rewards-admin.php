<?php
/**
 * Rewards Admin Page
 * 
 * 负责渲染礼品卡、优惠券和积分兑换管理页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.5
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 奖励管理类 (Rewards)
 */
class Tanzanite_Rewards_Admin {

	/**
	 * 渲染礼品卡和优惠券管理页面
	 */
	public static function render_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		$nonce = wp_create_nonce( 'wp_rest' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.5';

		// 加载媒体上传器
		wp_enqueue_media();
		
		// 加载 Rewards JS
		wp_enqueue_script(
			'tz-rewards',
			TANZANITE_PLUGIN_URL . 'assets/js/rewards.js',
			array( 'tz-admin-common' ),
			$version . '.fixed.' . time(),
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-rewards',
			'TzRewardsConfig',
			array(
				'nonce'              => $nonce,
				'couponsListUrl'     => esc_url_raw( rest_url( 'tanzanite/v1/coupons' ) ),
				'couponsSingleUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/coupons/' ) ),
				'giftcardsListUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/giftcards' ) ),
				'giftcardsSingleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/giftcards/' ) ),
				'transactionsUrl'    => esc_url_raw( rest_url( 'tanzanite/v1/rewards-transactions' ) ),
				'i18n'               => array(
					'noData'           => __( '暂无数据', 'tanzanite-settings' ),
					'loadFailed'       => __( '加载失败', 'tanzanite-settings' ),
					'saveSuccess'      => __( '保存成功', 'tanzanite-settings' ),
					'saveFailed'       => __( '保存失败', 'tanzanite-settings' ),
					'deleteConfirm'    => __( '确定删除？', 'tanzanite-settings' ),
					'deleteSuccess'    => __( '删除成功', 'tanzanite-settings' ),
					'generateCode'     => __( '生成代码', 'tanzanite-settings' ),
					'codeGenerated'    => __( '代码已生成', 'tanzanite-settings' ),
					'invalidAmount'    => __( '金额无效', 'tanzanite-settings' ),
					'invalidPoints'    => __( '积分无效', 'tanzanite-settings' ),
				),
			)
		);

		echo '<div class="tz-settings-wrapper tz-rewards-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Gift Cards & Coupons', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '管理优惠券、礼品卡和积分兑换。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-rewards-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		// 标签页导航
		echo '  <div class="tz-tabs-nav" style="margin-bottom:20px;border-bottom:1px solid #ccc;">';
		echo '      <button class="tz-tab-btn active" data-tab="coupons">' . esc_html__( '优惠券', 'tanzanite-settings' ) . '</button>';
		echo '      <button class="tz-tab-btn" data-tab="giftcards">' . esc_html__( '礼品卡', 'tanzanite-settings' ) . '</button>';
		echo '      <button class="tz-tab-btn" data-tab="transactions">' . esc_html__( '交易记录', 'tanzanite-settings' ) . '</button>';
		echo '      <button class="tz-tab-btn" data-tab="redeem-settings">' . esc_html__( '积分兑换设置', 'tanzanite-settings' ) . '</button>';
		echo '  </div>';

		// 优惠券标签页
		echo '  <div class="tz-tab-content" id="tz-tab-coupons">';
		echo '      <div class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( '优惠券列表', 'tanzanite-settings' ) . '</div>';
		echo '          <div style="margin:16px 0;">';
		echo '              <button class="button button-primary" id="tz-coupon-add">' . esc_html__( '添加优惠券', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button" id="tz-coupon-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '          <div style="overflow:auto;">';
		echo '              <table class="widefat fixed striped" id="tz-coupon-table">';
		echo '                  <thead><tr>';
		echo '                      <th style="width:60px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:120px;">' . esc_html__( '代码', 'tanzanite-settings' ) . '</th>';
		echo '                      <th>' . esc_html__( '标题', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '折扣类型', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:80px;">' . esc_html__( '折扣', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '使用次数', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:160px;">' . esc_html__( '过期时间', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:80px;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:160px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
		echo '                  </tr></thead>';
		echo '                  <tbody></tbody>';
		echo '              </table>';
		echo '          </div>';
		echo '      </div>';
		echo '  </div>';

		// 礼品卡标签页
		echo '  <div class="tz-tab-content" id="tz-tab-giftcards" style="display:none;">';
		echo '      <div class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( '礼品卡列表', 'tanzanite-settings' ) . '</div>';
		echo '          <div style="margin:16px 0;">';
		echo '              <button class="button button-primary" id="tz-giftcard-add">' . esc_html__( '添加礼品卡', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button" id="tz-giftcard-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '          <div style="overflow:auto;">';
		echo '              <table class="widefat fixed striped" id="tz-giftcard-table">';
		echo '                  <thead><tr>';
		echo '                      <th style="width:80px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:140px;">' . esc_html__( '卡号', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '余额', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '原始金额', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '积分消耗', 'tanzanite-settings' ) . '</th>';
		echo '                      <th>' . esc_html__( '持有人', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:140px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
		echo '                  </tr></thead>';
		echo '                  <tbody></tbody>';
		echo '              </table>';
		echo '          </div>';
		echo '      </div>';
		echo '  </div>';

		// 交易记录标签页
		echo '  <div class="tz-tab-content" id="tz-tab-transactions" style="display:none;">';
		echo '      <div class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( '积分交易记录', 'tanzanite-settings' ) . '</div>';
		echo '          <div style="margin:16px 0;">';
		echo '              <button class="button" id="tz-transaction-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '          <div style="overflow:auto;">';
		echo '              <table class="widefat fixed striped" id="tz-transaction-table">';
		echo '                  <thead><tr>';
		echo '                      <th style="width:80px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:120px;">' . esc_html__( '用户', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '类型', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '动作', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '积分变化', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:100px;">' . esc_html__( '金额变化', 'tanzanite-settings' ) . '</th>';
		echo '                      <th>' . esc_html__( '备注', 'tanzanite-settings' ) . '</th>';
		echo '                      <th style="width:160px;">' . esc_html__( '时间', 'tanzanite-settings' ) . '</th>';
		echo '                  </tr></thead>';
		echo '                  <tbody></tbody>';
		echo '              </table>';
		echo '          </div>';
		echo '      </div>';
		echo '  </div>';

		// 积分兑换设置标签页
		echo '  <div class="tz-tab-content" id="tz-tab-redeem-settings" style="display:none;">';
		echo '      <div class="tz-settings-section">';
		echo '          <div class="tz-section-title">' . esc_html__( '积分兑换礼品卡设置', 'tanzanite-settings' ) . '</div>';
		echo '          <form method="post" action="' . esc_url( admin_url( 'admin-post.php' ) ) . '" id="tz-redeem-settings-form">';
		wp_nonce_field( 'tz_save_redeem_settings', 'tz_redeem_nonce' );
		echo '              <input type="hidden" name="action" value="tz_save_redeem_settings" />';
		echo '              <table class="form-table">';
		echo '                  <tr>';
		echo '                      <th scope="row"><label for="redeem_enabled">' . esc_html__( '启用积分兑换', 'tanzanite-settings' ) . '</label></th>';
		echo '                      <td>';
		echo '                          <label><input type="checkbox" name="redeem_enabled" id="redeem_enabled" value="1" ' . checked( get_option( 'tz_redeem_enabled', '1' ), '1', false ) . ' /> ' . esc_html__( '允许用户使用积分兑换礼品卡', 'tanzanite-settings' ) . '</label>';
		echo '                          <p class="description">' . esc_html__( '⚠️ 前端 Nuxt 页面需要调用 REST API 实现兑换功能', 'tanzanite-settings' ) . '</p>';
		echo '                          <p class="description" style="color:#0073aa;"><strong>💡 提示：</strong>礼品卡的面值和积分价格在"礼品卡"页面设置</p>';
		echo '                      </td>';
		echo '                  </tr>';
		echo '                  <tr>';
		echo '                      <th scope="row"><label for="redeem_card_expiry_days">' . esc_html__( '礼品卡有效期', 'tanzanite-settings' ) . '</label></th>';
		echo '                      <td>';
		echo '                          <input type="number" name="redeem_card_expiry_days" id="redeem_card_expiry_days" value="' . esc_attr( get_option( 'tz_redeem_card_expiry_days', '365' ) ) . '" class="regular-text" min="0" step="1" />';
		echo '                          <p class="description">' . esc_html__( '兑换的礼品卡有效期天数（0 表示永久有效）', 'tanzanite-settings' ) . '</p>';
		echo '                      </td>';
		echo '                  </tr>';
		echo '                  <tr>';
		echo '                      <th scope="row"><label for="giftcard_cover_design">' . esc_html__( '礼品卡封面设计', 'tanzanite-settings' ) . '</label></th>';
		echo '                      <td>';
		echo '                          <div style="margin-bottom:12px;">';
		echo '                              <label><input type="radio" name="giftcard_cover_type" value="default" ' . checked( get_option( 'tz_giftcard_cover_type', 'default' ), 'default', false ) . ' /> ' . esc_html__( '使用默认封面', 'tanzanite-settings' ) . '</label><br>';
		echo '                              <label><input type="radio" name="giftcard_cover_type" value="custom" ' . checked( get_option( 'tz_giftcard_cover_type', 'default' ), 'custom', false ) . ' /> ' . esc_html__( '自定义封面图片', 'tanzanite-settings' ) . '</label><br>';
		echo '                              <label><input type="radio" name="giftcard_cover_type" value="template" ' . checked( get_option( 'tz_giftcard_cover_type', 'default' ), 'template', false ) . ' /> ' . esc_html__( '使用封面模板', 'tanzanite-settings' ) . '</label>';
		echo '                          </div>';
		echo '                          <div id="tz-giftcard-custom-cover" style="display:' . ( get_option( 'tz_giftcard_cover_type', 'default' ) === 'custom' ? 'block' : 'none' ) . ';margin-bottom:12px;">';
		echo '                              <input type="url" name="giftcard_cover_url" id="giftcard_cover_url" value="' . esc_attr( get_option( 'tz_giftcard_cover_url', '' ) ) . '" class="regular-text" placeholder="https://example.com/giftcard-cover.jpg" />';
		echo '                              <button type="button" class="button" id="tz-upload-cover">' . esc_html__( '上传图片', 'tanzanite-settings' ) . '</button>';
		echo '                              <p class="description">' . esc_html__( '建议尺寸：400x250px，支持 JPG、PNG 格式', 'tanzanite-settings' ) . '</p>';
		echo '                          </div>';
		echo '                          <div id="tz-giftcard-template-cover" style="display:' . ( get_option( 'tz_giftcard_cover_type', 'default' ) === 'template' ? 'block' : 'none' ) . ';margin-bottom:12px;">';
		echo '                              <select name="giftcard_cover_template" id="giftcard_cover_template" class="regular-text">';
		$selected_template = get_option( 'tz_giftcard_cover_template', 'elegant' );
		$templates = array(
			'elegant' => __( '优雅风格 - 深蓝渐变', 'tanzanite-settings' ),
			'festive' => __( '节日风格 - 红金配色', 'tanzanite-settings' ),
			'modern' => __( '现代风格 - 简约灰白', 'tanzanite-settings' ),
			'luxury' => __( '奢华风格 - 黑金配色', 'tanzanite-settings' ),
			'spring' => __( '春季风格 - 清新绿色', 'tanzanite-settings' ),
		);
		foreach ( $templates as $value => $label ) {
			echo '                                  <option value="' . esc_attr( $value ) . '" ' . selected( $selected_template, $value, false ) . '>' . esc_html( $label ) . '</option>';
		}
		echo '                              </select>';
		echo '                              <p class="description">' . esc_html__( '选择预设的封面模板样式', 'tanzanite-settings' ) . '</p>';
		echo '                          </div>';
		echo '                          <div id="tz-giftcard-cover-preview" style="margin-top:12px;padding:12px;border:1px solid #ddd;border-radius:4px;background:#f9f9f9;">';
		echo '                              <strong>' . esc_html__( '封面预览：', 'tanzanite-settings' ) . '</strong>';
		echo '                              <div id="tz-cover-preview-area" style="margin-top:8px;width:200px;height:125px;border:1px solid #ccc;border-radius:8px;background:#fff;display:flex;align-items:center;justify-content:center;color:#666;font-size:12px;">';
		echo '                                  ' . esc_html__( '礼品卡封面预览', 'tanzanite-settings' );
		echo '                              </div>';
		echo '                          </div>';
		echo '                          <p class="description" style="color:#d63638;"><strong>📌 前端 Nuxt 提示：</strong>通过 REST API 获取封面配置，在礼品卡显示时使用</p>';
		echo '                      </td>';
		echo '                  </tr>';
		echo '              </table>';
		echo '              <p class="submit">';
		echo '                  <button type="submit" class="button button-primary">' . esc_html__( '保存设置', 'tanzanite-settings' ) . '</button>';
		echo '              </p>';
		echo '          </form>';
		echo '      </div>';
		
		// API 接口文档说明
		echo '      <div class="tz-settings-section" style="margin-top:20px;background:#f0f6fc;border-left:4px solid #0073aa;padding:15px;">';
		echo '          <h3 style="margin-top:0;">📖 前端 Nuxt 对接指南</h3>';
		echo '          <h4>1. 获取兑换配置</h4>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">GET /wp-json/tanzanite/v1/redeem/config</pre>';
		echo '          <p><strong>返回示例：</strong></p>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "enabled": true,
  "exchange_rate": 100,
  "min_points": 1000,
  "max_value_per_day": 500,
  "card_expiry_days": 365,
  "preset_values": [10, 50, 100, 200, 500]
}</pre>';
		echo '          <h4>2. 查询用户积分余额</h4>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">GET /wp-json/tanzanite/v1/loyalty/points
Headers: X-WP-Nonce: {nonce}</pre>';
		echo '          <p><strong>返回示例：</strong></p>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "user_id": 123,
  "points": 5000,
  "can_redeem": true,
  "max_redeemable_value": 50
}</pre>';
		echo '          <h4>3. 积分兑换礼品卡</h4>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">POST /wp-json/tanzanite/v1/giftcards/redeem
Headers: 
  Content-Type: application/json
  X-WP-Nonce: {nonce}
Body:
{
  "points": 1000,
  "giftcard_value": 10
}</pre>';
		echo '          <p><strong>成功返回：</strong></p>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "success": true,
  "giftcard_id": 456,
  "card_code": "REDEEM-ABC123XYZ",
  "balance": 10.00,
  "points_spent": 1000,
  "points_remaining": 4000,
  "expires_at": "2026-11-11 08:00:00",
  "message": "兑换成功"
}</pre>';
		echo '          <p><strong>失败返回：</strong></p>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "code": "insufficient_points",
  "message": "积分不足",
  "data": { "status": 400 }
}</pre>';
		echo '          <h4>4. 查询用户的礼品卡列表</h4>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">GET /wp-json/tanzanite/v1/giftcards/my
Headers: X-WP-Nonce: {nonce}</pre>';
		echo '          <p><strong>返回示例：</strong></p>';
		echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "items": [
    {
      "id": 456,
      "card_code": "REDEEM-ABC123XYZ",
      "balance": 10.00,
      "original_value": 10.00,
      "points_spent": 1000,
      "status": "active",
      "expires_at": "2026-11-11 08:00:00",
      "created_at": "2025-11-11 08:00:00"
    }
  ]
}</pre>';
		echo '          <hr style="margin:20px 0;" />';
		echo '          <p style="color:#d63638;"><strong>⚠️ 重要提示：</strong></p>';
		echo '          <ul style="margin-left:20px;">';
		echo '              <li>所有需要身份验证的接口都需要在请求头中包含 <code>X-WP-Nonce</code></li>';
		echo '              <li>Nonce 可通过 <code>/wp-json/</code> 端点获取，或在登录后从 cookie 中读取</li>';
		echo '              <li>兑换接口会自动扣除用户积分并创建交易记录</li>';
		echo '              <li>礼品卡卡号自动生成，格式为 <code>REDEEM-{12位随机字符}</code></li>';
		echo '              <li>建议在前端实现兑换确认弹窗，避免误操作</li>';
		echo '          </ul>';
		echo '      </div>';
		echo '  </div>';

		echo '</div>';
		
		// 添加礼品卡封面设计的 JavaScript
		?>
		<script type="text/javascript">
		document.addEventListener('DOMContentLoaded', function() {
			console.log('Gift card cover script loading...');
			
			// 封面类型切换
			const coverTypeRadios = document.querySelectorAll('input[name="giftcard_cover_type"]');
			const customCoverDiv = document.getElementById('tz-giftcard-custom-cover');
			const templateCoverDiv = document.getElementById('tz-giftcard-template-cover');
			const previewArea = document.getElementById('tz-cover-preview-area');
			
			console.log('Cover type radios found:', coverTypeRadios.length);
			console.log('Custom cover div found:', !!customCoverDiv);
			console.log('Template cover div found:', !!templateCoverDiv);
			console.log('Preview area found:', !!previewArea);
			
			function updateCoverDisplay() {
				const selectedRadio = document.querySelector('input[name="giftcard_cover_type"]:checked');
				const selectedType = selectedRadio ? selectedRadio.value : 'default';
				console.log('Selected cover type:', selectedType);
				
				if (customCoverDiv) {
					customCoverDiv.style.display = selectedType === 'custom' ? 'block' : 'none';
					console.log('Custom cover div display:', customCoverDiv.style.display);
				}
				
				if (templateCoverDiv) {
					templateCoverDiv.style.display = selectedType === 'template' ? 'block' : 'none';
					console.log('Template cover div display:', templateCoverDiv.style.display);
				}
				
				updatePreview();
			}
			
			function updatePreview() {
				if (!previewArea) return;
				
				const selectedRadio = document.querySelector('input[name="giftcard_cover_type"]:checked');
				const selectedType = selectedRadio ? selectedRadio.value : 'default';
				const customUrlInput = document.getElementById('giftcard_cover_url');
				const customUrl = customUrlInput ? customUrlInput.value : '';
				const templateSelect = document.getElementById('giftcard_cover_template');
				const selectedTemplate = templateSelect ? templateSelect.value : 'elegant';
				
				console.log('updatePreview called:');
				console.log('- selectedType:', selectedType);
				console.log('- customUrl:', customUrl);
				console.log('- selectedTemplate:', selectedTemplate);
				console.log('- templateSelect element:', templateSelect);
				
				if (selectedType === 'custom' && customUrl) {
					previewArea.style.backgroundImage = 'url(' + customUrl + ')';
					previewArea.style.backgroundSize = 'cover';
					previewArea.style.backgroundPosition = 'center';
					previewArea.textContent = '';
				} else if (selectedType === 'template') {
					const templates = {
						elegant: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
						festive: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
						modern: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
						luxury: 'linear-gradient(135deg, #434343 0%, #000000 100%)',
						spring: 'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)'
					};
					const selectedGradient = templates[selectedTemplate] || templates.elegant;
					console.log('Applying template:', selectedTemplate, 'with gradient:', selectedGradient);
					
					// 清除之前的样式
					previewArea.style.removeProperty('background');
					previewArea.style.removeProperty('background-color');
					previewArea.style.removeProperty('background-image');
					
					// 使用 setProperty 强制设置样式
					previewArea.style.setProperty('background-image', selectedGradient, 'important');
					previewArea.style.setProperty('background-color', 'transparent', 'important');
					
					// 设置文本样式
					previewArea.style.color = '#fff';
					previewArea.style.fontWeight = 'bold';
					previewArea.textContent = '礼品卡';
					
					console.log('Preview area backgroundImage set to:', previewArea.style.backgroundImage);
					console.log('Final computed style:', window.getComputedStyle(previewArea).backgroundImage);
				} else {
					previewArea.style.background = '#f0f0f0';
					previewArea.style.backgroundImage = 'none';
					previewArea.style.color = '#666';
					previewArea.style.fontWeight = 'normal';
					previewArea.textContent = '默认封面';
				}
			}
			
			// 绑定事件
			coverTypeRadios.forEach(function(radio) {
				radio.addEventListener('change', updateCoverDisplay);
			});
			
			const urlInput = document.getElementById('giftcard_cover_url');
			if (urlInput) {
				urlInput.addEventListener('input', updatePreview);
			}
			
			const templateSelect = document.getElementById('giftcard_cover_template');
			if (templateSelect) {
				templateSelect.addEventListener('change', updatePreview);
			}
			
			// 上传按钮点击事件
			const uploadBtn = document.getElementById('tz-upload-cover');
			if (uploadBtn) {
				uploadBtn.addEventListener('click', function(e) {
					e.preventDefault();
					
					// 检查 wp.media 是否存在
					if (typeof wp === 'undefined' || typeof wp.media === 'undefined') {
						alert('WordPress 媒体库未加载');
						return;
					}
					
					const imageUploader = wp.media({
						title: '选择礼品卡封面图片',
						button: {
							text: '使用此图片'
						},
						multiple: false
					});
					
					imageUploader.on('select', function() {
						const attachment = imageUploader.state().get('selection').first().toJSON();
						if (urlInput) {
							urlInput.value = attachment.url;
							updatePreview();
						}
					});
					
					imageUploader.open();
				});
			}
			
			// 初始化显示
			updateCoverDisplay();
		});
		</script>
		<?php
	}
}
