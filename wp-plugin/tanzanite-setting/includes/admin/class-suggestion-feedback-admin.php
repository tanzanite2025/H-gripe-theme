<?php
/**
 * Suggestion Feedback Admin Page
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Tanzanite_Suggestion_Feedback_Admin {
	public static function render_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( esc_html__( '无权限访问该页面。', 'tanzanite-settings' ) );
		}

		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.0.0';
		$handle  = 'tz-suggestion-feedback';

		wp_enqueue_script(
			$handle,
			TANZANITE_PLUGIN_URL . 'assets/js/suggestion-feedback.js',
			array( 'tz-admin-common' ),
			$version,
			true
		);

		wp_localize_script(
			$handle,
			'TzSuggestionFeedbackConfig',
			array(
				'listUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/suggestion-feedback' ) ),
				'eligibilityUrl' => esc_url_raw( rest_url( 'tanzanite/v1/suggestion-feedback/eligibility' ) ),
				'nonce'     => wp_create_nonce( 'wp_rest' ),
				'defaultStatus' => 'new',
				'labels'    => array(
					'new'        => __( '待审核', 'tanzanite-settings' ),
					'in_review'  => __( '处理中', 'tanzanite-settings' ),
					'resolved'   => __( '已处理', 'tanzanite-settings' ),
					'archived'   => __( '已归档', 'tanzanite-settings' ),
				),
			)
		);

		echo '<div class="tz-settings-wrapper tz-suggestion-feedback">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Suggestion Feedback', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '查看用户提交的配置反馈，支持状态流转与备注。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-suggestion-feedback-notice" class="notice" style="display:none;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
		echo '      <div class="tz-filter-row" style="display:flex;flex-wrap:wrap;gap:12px;">';
		echo '          <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select id="tz-suggestion-status" class="widefat" style="min-width:160px;">';
		echo '              <option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
		foreach ( array( 'new', 'in_review', 'resolved', 'archived' ) as $status ) {
			echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
		}
		echo '          </select></label>';
		echo '          <label style="flex:1;min-width:260px;">' . esc_html__( '关键词 (姓名/邮箱/订单/内容)', 'tanzanite-settings' ) . '<input type="text" id="tz-suggestion-search" class="widefat" /></label>';
		echo '          <button class="button button-primary" id="tz-suggestion-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '反馈列表', 'tanzanite-settings' ) . '</div>';
		echo '      <div class="tz-table-wrapper" style="overflow:auto;">';
		echo '          <table class="widefat fixed striped" id="tz-suggestion-feedback-table" style="min-width:960px;">';
		echo '              <thead><tr>';
		foreach ( array( 'ID', '用户', '分类', '内容', '时间', '状态', '操作' ) as $column ) {
			echo '<th>' . esc_html( $column ) . '</th>';
		}
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '      <div class="tz-pagination" style="display:flex;gap:12px;margin-top:12px;align-items:center;">';
		echo '          <button class="button" id="tz-suggestion-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
		echo '          <span id="tz-suggestion-page-info"></span>';
		echo '          <button class="button" id="tz-suggestion-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </div>';

		echo '</div>';
	}
}
