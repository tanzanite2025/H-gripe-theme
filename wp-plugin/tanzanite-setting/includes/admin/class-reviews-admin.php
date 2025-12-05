<?php
/**
 * Reviews Admin Page
 * 
 * 负责渲染评论管理页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.3
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 评论管理类
 */
class Tanzanite_Reviews_Admin {

	/**
	 * 允许的评论状态
	 */
	const ALLOWED_REVIEW_STATUSES = [ 'pending', 'approved', 'rejected', 'hidden' ];

	/**
	 * 渲染评论管理页面
	 */
	public static function render_page() {
		$nonce        = wp_create_nonce( 'wp_rest' );
		$rest_list    = esc_url_raw( rest_url( 'tanzanite/v1/reviews' ) );
		$rest_single  = esc_url_raw( rest_url( 'tanzanite/v1/reviews/' ) );
		$statuses     = wp_json_encode( self::ALLOWED_REVIEW_STATUSES );
		$can_manage   = current_user_can( 'tanz_manage_reviews' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.3';

		// 加载评价管理 JS
		wp_enqueue_script(
			'tz-reviews',
			TANZANITE_PLUGIN_URL . 'assets/js/reviews.js',
			array( 'tz-admin-common' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-reviews',
			'TzReviewsConfig',
			array(
				'listUrl'    => $rest_list,
				'singleUrl'  => $rest_single,
				'nonce'      => $nonce,
				'statuses'   => array_values( self::ALLOWED_REVIEW_STATUSES ),
				'canManage'  => $can_manage,
				'i18n'       => array(
					'noPermission'        => __( '当前账号仅具备查看权限，审核操作已禁用。', 'tanzanite-settings' ),
					'noPermissionHint'    => __( '如需执行审核或回复，请联系管理员授予"评价管理"权限。', 'tanzanite-settings' ),
					'loadFailed'          => __( '加载评价列表失败。', 'tanzanite-settings' ),
					'saveSuccess'         => __( '评价已更新。', 'tanzanite-settings' ),
					'deleteConfirm'       => __( '确定删除该评价？此操作不可撤销。', 'tanzanite-settings' ),
					'deleteSuccess'       => __( '评价已删除。', 'tanzanite-settings' ),
					'selectReview'        => __( '请先选择要操作的评价。', 'tanzanite-settings' ),
					'contentPlaceholder'  => __( '暂无内容', 'tanzanite-settings' ),
					'view'                => __( '查看', 'tanzanite-settings' ),
					'approve'             => __( '通过', 'tanzanite-settings' ),
					'reject'              => __( '拒绝', 'tanzanite-settings' ),
					'hide'                => __( '隐藏', 'tanzanite-settings' ),
					'markFeatured'        => __( '标记精华', 'tanzanite-settings' ),
					'unmarkFeatured'      => __( '取消精华', 'tanzanite-settings' ),
					'yes'                 => __( '是', 'tanzanite-settings' ),
					'no'                  => __( '否', 'tanzanite-settings' ),
					'itemsLabel'          => __( '条评价', 'tanzanite-settings' ),
				),
			)
		);

		echo '<div class="tz-settings-wrapper tz-reviews-wrapper">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Product Reviews', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '集中处理来自各渠道的商品评价，支持审核、回复与标记精华。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-review-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-review-filters" class="tz-review-filters" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px;max-width:1200px;">';
		echo '          <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select name="status" class="widefat">';
		echo '              <option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
		foreach ( self::ALLOWED_REVIEW_STATUSES as $status ) {
			echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
		}
		echo '          </select></label>';
		echo '          <label>' . esc_html__( '商品 ID', 'tanzanite-settings' ) . '<input type="number" name="product_id" class="widefat" /></label>';
		echo '          <label>' . esc_html__( '关键词（作者/内容）', 'tanzanite-settings' ) . '<input type="text" name="search" class="widefat" /></label>';
		echo '          <label>' . esc_html__( '每页条数', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select></label>';
		echo '      </form>';
		echo '      <div style="display:flex;gap:12px;margin-top:12px;">';
		echo '          <button class="button button-primary" id="tz-review-refresh">' . esc_html__( '刷新列表', 'tanzanite-settings' ) . '</button>';
		echo '          <button class="button" id="tz-review-reset">' . esc_html__( '重置筛选', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '评价列表', 'tanzanite-settings' ) . '</div>';
		echo '      <div style="overflow:auto;margin-top:12px;">';
		echo '          <table class="widefat fixed striped" id="tz-review-table" style="min-width:1100px;">';
		echo '              <thead><tr>';
		foreach ( [ 'ID', 'Product', 'User', 'Rating', 'Content', 'Status', 'Featured', 'Created', 'Actions' ] as $column ) {
			echo '<th>' . esc_html( $column ) . '</th>';
		}
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '      <div class="tz-review-pagination" style="display:flex;align-items:center;gap:12px;margin-top:12px;">';
		echo '          <button class="button" id="tz-review-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
		echo '          <span id="tz-review-page-info"></span>';
		echo '          <button class="button" id="tz-review-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '评价详情与审核', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-review-detail" style="display:grid;gap:12px;max-width:1100px;">';
		echo '          <input type="hidden" id="tz-review-id" />';
		echo '          <div class="tz-review-meta" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:12px;">';
		echo '              <label>' . esc_html__( '评价状态', 'tanzanite-settings' ) . '<select id="tz-review-status" class="widefat">';
		echo '                  <option value="">' . esc_html__( '保持不变', 'tanzanite-settings' ) . '</option>';
		foreach ( self::ALLOWED_REVIEW_STATUSES as $status ) {
			echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
		}
		echo '              </select></label>';
		echo '              <label style="display:flex;align-items:center;gap:8px;margin-top:24px;"><input type="checkbox" id="tz-review-featured" /> ' . esc_html__( '标记为精华', 'tanzanite-settings' ) . '</label>';
		echo '              <div>';
		echo '                  <div><strong>' . esc_html__( '评分', 'tanzanite-settings' ) . ':</strong> <span id="tz-review-rating">-</span></div>';
		echo '                  <div><strong>' . esc_html__( '作者', 'tanzanite-settings' ) . ':</strong> <span id="tz-review-author">-</span></div>';
		echo '                  <div><strong>' . esc_html__( '创建时间', 'tanzanite-settings' ) . ':</strong> <span id="tz-review-created">-</span></div>';
		echo '              </div>';
		echo '          </div>';
		echo '          <label>' . esc_html__( '评价内容', 'tanzanite-settings' ) . '<textarea id="tz-review-content" rows="6" class="widefat" readonly style="background:#f6f7f7;"></textarea></label>';
		echo '          <label>' . esc_html__( '附件', 'tanzanite-settings' ) . '<div id="tz-review-images" style="display:flex;gap:12px;flex-wrap:wrap;"></div></label>';
		echo '          <label>' . esc_html__( '后台回复', 'tanzanite-settings' ) . '<textarea id="tz-review-reply" rows="4" class="widefat" placeholder="' . esc_attr__( '输入回复内容，留空将清除回复。', 'tanzanite-settings' ) . '"></textarea></label>';
		echo '          <div style="display:flex;gap:12px;">';
		echo '              <button class="button button-primary" id="tz-review-save" type="submit">' . esc_html__( '保存更新', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button" id="tz-review-cancel" type="button">' . esc_html__( '取消选择', 'tanzanite-settings' ) . '</button>';
		echo '              <button class="button button-secondary" id="tz-review-delete" type="button">' . esc_html__( '删除评价', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '      </form>';
		echo '  </div>';

		echo '</div>';
	}
}
