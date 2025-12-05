<?php
/**
 * Members Admin Page
 * 
 * 负责渲染会员档案管理页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.4
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 会员管理类
 */
class Tanzanite_Members_Admin {

	/**
	 * 渲染会员档案页面
	 */
	public static function render_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		$nonce = wp_create_nonce( 'wp_rest' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.4';

		// 加载 Member Profiles JS
		wp_enqueue_script(
			'tz-member-profiles',
			TANZANITE_PLUGIN_URL . 'assets/js/member-profiles.js',
			array( 'tz-admin-common' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-member-profiles',
			'TzMemberProfilesConfig',
			array(
				'nonce'      => $nonce,
				'listUrl'    => esc_url_raw( rest_url( 'tanzanite/v1/members' ) ),
				'singleUrl'  => esc_url_raw( rest_url( 'tanzanite/v1/members/' ) ),
				'exportUrl'  => esc_url_raw( admin_url( 'admin-ajax.php?action=tanz_export_members' ) ),
				'i18n'       => array(
					'noData'          => __( '暂无会员数据', 'tanzanite-settings' ),
					'loadFailed'      => __( '加载会员列表失败', 'tanzanite-settings' ),
					'saveSuccess'     => __( '会员信息已更新', 'tanzanite-settings' ),
					'saveFailed'      => __( '保存失败', 'tanzanite-settings' ),
					'deleteConfirm'   => __( '确定删除该会员档案？', 'tanzanite-settings' ),
					'deleteSuccess'   => __( '会员档案已删除', 'tanzanite-settings' ),
					'selectMember'    => __( '请先选择要操作的会员', 'tanzanite-settings' ),
					'exportSuccess'   => __( '导出成功', 'tanzanite-settings' ),
					'importSuccess'   => __( '导入成功', 'tanzanite-settings' ),
				),
			)
		);

		echo '<div class="tz-settings-wrapper tz-members-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Member Profiles', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '管理会员档案、积分余额和营销偏好。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-member-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-member-filters" class="tz-member-filters" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px;max-width:1200px;">';
		echo '          <label>' . esc_html__( '关键词（姓名/邮箱/手机）', 'tanzanite-settings' ) . '<input type="text" name="search" class="widefat" /></label>';
		echo '          <label>' . esc_html__( '最低积分', 'tanzanite-settings' ) . '<input type="number" name="min_points" class="widefat" min="0" /></label>';
		echo '          <label>' . esc_html__( '每页条数', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select></label>';
		echo '      </form>';
		echo '      <div style="display:flex;gap:12px;margin-top:12px;">';
		echo '          <button class="button button-primary" id="tz-member-refresh">' . esc_html__( '刷新列表', 'tanzanite-settings' ) . '</button>';
		echo '          <button class="button" id="tz-member-reset">' . esc_html__( '重置筛选', 'tanzanite-settings' ) . '</button>';
		echo '          <button class="button" id="tz-member-export">' . esc_html__( '导出 CSV', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '会员列表', 'tanzanite-settings' ) . '</div>';
		echo '      <div style="overflow:auto;margin-top:16px;">';
		echo '          <table class="widefat fixed striped" id="tz-member-table" style="min-width:960px;">';
		echo '              <thead><tr>';
		echo '                  <th style="width:60px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:120px;">' . esc_html__( '用户名', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:150px;">' . esc_html__( '姓名', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:180px;">' . esc_html__( '邮箱', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:120px;">' . esc_html__( '手机', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:80px;">' . esc_html__( '积分', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:100px;">' . esc_html__( '等级', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:140px;">' . esc_html__( '注册时间', 'tanzanite-settings' ) . '</th>';
		echo '                  <th style="width:120px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
		echo '              </tr></thead>';
		echo '              <tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '      <div class="tablenav bottom" style="display:flex;justify-content:space-between;align-items:center;padding:12px 0;">';
		echo '          <div class="tablenav-pages">';
		echo '              <span class="displaying-num" id="tz-member-page-info"></span>';
		echo '              <span class="pagination-links">';
		echo '                  <button class="button" id="tz-member-prev" disabled>' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
		echo '                  <button class="button" id="tz-member-next" disabled>' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
		echo '              </span>';
		echo '          </div>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section" id="tz-member-detail-section" style="display:none;">';
		echo '      <div class="tz-section-title">' . esc_html__( '会员详情', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-member-detail" style="max-width:800px;margin-top:16px;">';
		echo '          <input type="hidden" id="tz-member-id" />';
		echo '          <table class="form-table">';
		echo '              <tr><th><label for="tz-member-username">' . esc_html__( '用户名', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-username" class="regular-text" readonly /></td></tr>';
		echo '              <tr><th><label for="tz-member-email">' . esc_html__( '邮箱', 'tanzanite-settings' ) . '</label></th><td><input type="email" id="tz-member-email" class="regular-text" readonly /></td></tr>';
		echo '              <tr><th><label for="tz-member-fullname">' . esc_html__( '姓名', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-fullname" class="regular-text" /></td></tr>';
		echo '              <tr><th><label for="tz-member-phone">' . esc_html__( '手机', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-phone" class="regular-text" /></td></tr>';
		echo '              <tr><th><label for="tz-member-country">' . esc_html__( '国家/地区', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-country" class="regular-text" /></td></tr>';
		echo '              <tr><th><label for="tz-member-address">' . esc_html__( '地址', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-address" class="regular-text" /></td></tr>';
		echo '              <tr><th><label for="tz-member-brand">' . esc_html__( '品牌', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-brand" class="regular-text" /></td></tr>';
		echo '              <tr><th><label for="tz-member-points">' . esc_html__( '积分', 'tanzanite-settings' ) . '</label></th><td><input type="number" id="tz-member-points" class="regular-text" min="0" /></td></tr>';
		echo '              <tr><th><label for="tz-member-marketing">' . esc_html__( '营销订阅', 'tanzanite-settings' ) . '</label></th><td><label><input type="checkbox" id="tz-member-marketing" /> ' . esc_html__( '接收营销信息', 'tanzanite-settings' ) . '</label></td></tr>';
		echo '              <tr><th><label for="tz-member-notes">' . esc_html__( '备注', 'tanzanite-settings' ) . '</label></th><td><textarea id="tz-member-notes" rows="5" class="large-text"></textarea></td></tr>';
		echo '          </table>';
		echo '          <div style="margin-top:16px;">';
		echo '              <button type="button" class="button button-primary" id="tz-member-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
		echo '              <button type="button" class="button" id="tz-member-cancel">' . esc_html__( '取消', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '      </form>';
		echo '  </div>';

		echo '</div>';
	}
}
