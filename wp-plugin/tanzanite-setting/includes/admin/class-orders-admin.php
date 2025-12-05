<?php
/**
 * Orders Admin Page
 * 
 * 负责渲染订单列表和订单批量操作页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.2
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 订单管理类
 */
class Tanzanite_Orders_Admin {

	/**
	 * 允许的订单状态
	 */
	const ALLOWED_ORDER_STATUSES = [ 'pending', 'paid', 'processing', 'shipped', 'completed', 'cancelled' ];

	/**
	 * 渲染订单列表页面
	 */
	public static function render_list_page() {
		$nonce      = wp_create_nonce( 'wp_rest' );
		$list_url   = esc_url_raw( rest_url( 'tanzanite/v1/orders' ) );
		$sync_url   = esc_url_raw( rest_url( 'tanzanite/v1/orders/' ) );
		$statuses   = array_values( self::ALLOWED_ORDER_STATUSES );
		$can_manage = current_user_can( 'tanz_manage_orders' );
		
		// 版本号，尝试使用插件主类定义的版本，如果不可用则回退
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.2';

		// 加载订单列表 JS
		// 注意：这里假设 assets 目录位于插件根目录下
		wp_enqueue_script(
			'tz-orders-list',
			TANZANITE_PLUGIN_URL . 'assets/js/orders-list.js',
			array( 'tz-admin-common' ), // 确保依赖 tz-admin-common (通常在 enqueue_admin_assets 中加载)
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-orders-list',
			'TzOrdersListConfig',
			array(
				'listUrl'      => $list_url,
				'syncBase'     => $sync_url,
				'nonce'        => $nonce,
				'canManage'    => $can_manage,
				'detailBase'   => esc_url_raw( admin_url( 'admin.php?page=tanzanite-settings-order-detail&order_id=' ) ),
				'statusLabels' => array(
					'pending'    => __( '待支付', 'tanzanite-settings' ),
					'paid'       => __( '已支付', 'tanzanite-settings' ),
					'processing' => __( '处理中', 'tanzanite-settings' ),
					'shipped'    => __( '已发货', 'tanzanite-settings' ),
					'completed'  => __( '已完成', 'tanzanite-settings' ),
					'cancelled'  => __( '已取消', 'tanzanite-settings' ),
				),
			)
		);

		echo '<div class="tz-settings-wrapper tz-orders-list">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'All Orders', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '按条件筛选订单，支持刷新物流状态与跳转批量工具。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-orders-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-orders-filters" class="tz-orders-filters" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px;max-width:1400px;">';
		echo '          <label>' . esc_html__( '订单状态', 'tanzanite-settings' ) . '<select name="status" class="widefat"><option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
		foreach ( $statuses as $status ) {
			echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
		}
		echo '          </select></label>';
		echo '          <label>' . esc_html__( '渠道来源', 'tanzanite-settings' ) . '<input type="text" name="channel" class="widefat" placeholder="web/app/h5" /></label>';
		echo '          <label>' . esc_html__( '支付方式', 'tanzanite-settings' ) . '<input type="text" name="payment_method" class="widefat" placeholder="wechat_pay" /></label>';
		echo '          <label>' . esc_html__( '物流服务商', 'tanzanite-settings' ) . '<input type="text" name="tracking_provider" class="widefat" placeholder="17track" /></label>';
		echo '          <label>' . esc_html__( '客户关键词', 'tanzanite-settings' ) . '<input type="text" name="customer_keyword" class="widefat" placeholder="姓名/邮箱/账号" /></label>';
		echo '          <label>' . esc_html__( '起始时间', 'tanzanite-settings' ) . '<input type="date" name="date_start" class="widefat" /></label>';
		echo '          <label>' . esc_html__( '结束时间', 'tanzanite-settings' ) . '<input type="date" name="date_end" class="widefat" /></label>';
		echo '          <label>' . esc_html__( '每页条数', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select></label>';
		echo '      </form>';
		echo '      <div style="display:flex;gap:12px;margin-top:12px;">';
		echo '          <button class="button button-primary" id="tz-orders-filter-submit">' . esc_html__( '应用筛选', 'tanzanite-settings' ) . '</button>';
		echo '          <button class="button" id="tz-orders-filter-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
		echo '          <a class="button" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-orders-bulk' ) ) . '">' . esc_html__( '前往批量工具', 'tanzanite-settings' ) . '</a>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '订单列表', 'tanzanite-settings' ) . '</div>';
		echo '      <div style="overflow:auto;margin-top:12px;">';
		echo '          <table class="widefat fixed striped" id="tz-orders-table" style="min-width:1200px;">';
		echo '              <thead><tr>';
		foreach ( [ __( '订单信息', 'tanzanite-settings' ), __( '客户', 'tanzanite-settings' ), __( '金额', 'tanzanite-settings' ), __( '状态', 'tanzanite-settings' ), __( '渠道 / 支付', 'tanzanite-settings' ), __( '创建时间', 'tanzanite-settings' ), __( '物流', 'tanzanite-settings' ), __( '操作', 'tanzanite-settings' ) ] as $column ) {
			echo '<th>' . esc_html( $column ) . '</th>';
		}
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';
		echo '      <div class="tz-orders-pagination" style="display:flex;align-items:center;gap:12px;margin-top:12px;">';
		echo '          <button class="button" id="tz-orders-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
		echo '          <span id="tz-orders-page-info"></span>';
		echo '          <button class="button" id="tz-orders-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </div>';

		echo '</div>';
	}

	/**
	 * 渲染订单批量操作页面
	 */
	public static function render_bulk_page() {
		$nonce        = wp_create_nonce( 'wp_rest' );
		$bulk_url     = esc_url_raw( rest_url( 'tanzanite/v1/orders' ) );
		$status_list  = self::ALLOWED_ORDER_STATUSES;
		
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.2';

		// 加载 Order Bulk JS
		wp_enqueue_script(
			'tz-order-bulk',
			TANZANITE_PLUGIN_URL . 'assets/js/order-bulk.js',
			array( 'jquery' ), // 确保有基本依赖
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-order-bulk',
			'TzOrderBulkConfig',
			array(
				'nonce'   => $nonce,
				'url'     => $bulk_url,
				'strings' => array(
					'invalidIds'    => __( '请输入有效的订单 ID。', 'tanzanite-settings' ),
					'invalidStatus' => __( '请选择目标状态。', 'tanzanite-settings' ),
					'done'          => __( '操作完成', 'tanzanite-settings' ),
				),
			)
		);

		echo '<div class="tz-settings-wrapper tz-orders-bulk">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Order Bulk Operations', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '这里可一次性更新多个订单状态或导出订单数据，方便客服与运营快速处理。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-order-bulk-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '批量更新订单状态', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-order-bulk-status" class="tz-bulk-form" style="display:grid;gap:12px;max-width:720px;">';
		echo '          <label>' . esc_html__( '订单 ID（逗号或换行分隔）', 'tanzanite-settings' ) . '<textarea rows="3" class="widefat" name="ids"></textarea></label>';
		echo '          <label>' . esc_html__( '目标状态', 'tanzanite-settings' ) . '<select name="status" class="widefat">';
		foreach ( $status_list as $status ) {
			echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
		}
		echo '          </select></label>';
		echo '          <p class="description">' . esc_html__( '状态变更将按订单状态机校验，不符合流转的订单会自动跳过。', 'tanzanite-settings' ) . '</p>';
		echo '          <button class="button button-primary" type="submit">' . esc_html__( '批量修改状态', 'tanzanite-settings' ) . '</button>';
		echo '      </form>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '批量导出订单', 'tanzanite-settings' ) . '</div>';
		echo '      <form id="tz-order-bulk-export" class="tz-bulk-form" style="display:grid;gap:12px;max-width:720px;">';
		echo '          <label>' . esc_html__( '订单 ID（逗号或换行分隔）', 'tanzanite-settings' ) . '<textarea rows="3" class="widefat" name="ids"></textarea></label>';
		echo '          <p class="description">' . esc_html__( '导出内容包含金额、渠道、支付方式及时间戳，会提供 JSON 结果与 CSV 下载。', 'tanzanite-settings' ) . '</p>';
		echo '          <button class="button" type="submit">' . esc_html__( '开始导出', 'tanzanite-settings' ) . '</button>';
		echo '      </form>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '操作结果', 'tanzanite-settings' ) . '</div>';
		echo '      <pre id="tz-order-bulk-result" style="background:#f6f7f7;border:1px solid #ccd0d4;padding:12px;max-height:320px;overflow:auto;"></pre>';
		echo '  </div>';

		echo '</div>';
	}
}
