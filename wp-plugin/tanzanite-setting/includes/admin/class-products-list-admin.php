<?php
/**
 * Products List Admin Page
 *
 * 负责渲染商品列表页面
 * 从 legacy-pages.php 迁移重构
 *
 * @package    Tanzanite_Settings
 * @subpackage Includes/Admin
 * @since      0.2.11
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 商品列表管理类
 */
class Tanzanite_Products_List_Admin {

	/**
	 * 允许的商品状态
	 */
	const ALLOWED_PRODUCT_STATUSES = [ 'draft', 'pending', 'publish', 'private' ];

	/**
	 * 渲染商品列表页面
	 */
	public static function render_page() {
		if ( ! current_user_can( 'tanz_view_products' ) ) {
			wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
		}

		$nonce        = wp_create_nonce( 'wp_rest' );
		$list_endpoint = esc_url_raw( rest_url( 'tanzanite/v1/products' ) );
		$single_link   = esc_url( admin_url( 'admin.php?page=tanzanite-settings-add-product&product_id=' ) );
		$seo_link      = esc_url( admin_url( 'admin.php?page=tanzanite-settings-add-product&focus=seo&product_id=' ) );
		$bulk_link     = esc_url( admin_url( 'admin.php?page=tanzanite-settings-sku-importer' ) );

		$can_manage = current_user_can( 'tanz_manage_products' );
		$can_bulk   = current_user_can( 'tanz_bulk_products' );

		// 版本号
		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.2.11';

		// 加载 Products List 模块（按依赖顺序）
		// 1. 渲染模块（被 core 依赖）
		wp_enqueue_script(
			'tz-products-list-render',
			TANZANITE_PLUGIN_URL . 'assets/js/products-list-render.js',
			array(),
			$version,
			true
		);

		// 2. 筛选模块（被 core 依赖）
		wp_enqueue_script(
			'tz-products-list-filters',
			TANZANITE_PLUGIN_URL . 'assets/js/products-list-filters.js',
			array(),
			$version,
			true
		);

		// 3. 批量操作模块（被 core 依赖）
		wp_enqueue_script(
			'tz-products-list-bulk',
			TANZANITE_PLUGIN_URL . 'assets/js/products-list-bulk.js',
			array(),
			$version,
			true
		);

		// 4. 单个操作模块（被 core 依赖）
		wp_enqueue_script(
			'tz-products-list-actions',
			TANZANITE_PLUGIN_URL . 'assets/js/products-list-actions.js',
			array(),
			$version,
			true
		);

		// 5. 核心模块（依赖所有其他模块）
		wp_enqueue_script(
			'tz-products-list-core',
			TANZANITE_PLUGIN_URL . 'assets/js/products-list-core.js',
			array( 'tz-products-list-render', 'tz-products-list-filters', 'tz-products-list-bulk', 'tz-products-list-actions' ),
			$version,
			true
		);

		// 传递配置到 JS
		wp_localize_script(
			'tz-products-list-core',
			'TzProductsListConfig',
			array(
				'nonce'            => $nonce,
				'listUrl'          => $list_endpoint,
				'singleUrl'        => $list_endpoint . '/',
				'bulkUrl'          => $list_endpoint,
				'editUrl'          => $single_link,
				'seoUrl'           => $seo_link,
				'canManage'        => $can_manage,
				'canBulk'          => $can_bulk,
				'categoryEndpoint' => esc_url_raw( rest_url( 'tanzanite/v1/categories' ) ),
				'tagsEndpoint'     => esc_url_raw( rest_url( 'tanzanite/v1/tags' ) ),
				'strings'          => array(
					'noData'                  => __( '暂无数据', 'tanzanite-settings' ),
					'loadFailed'              => __( '加载失败', 'tanzanite-settings' ),
					'pageTemplate'            => __( '第 {page}/{pages} 页', 'tanzanite-settings' ),
					'expandFilters'           => __( '展开筛选', 'tanzanite-settings' ),
					'collapseFilters'         => __( '收起筛选', 'tanzanite-settings' ),
					'editLabel'               => __( '编辑', 'tanzanite-settings' ),
					'seoLabel'                => __( 'SEO', 'tanzanite-settings' ),
					'previewLabel'            => __( '预览', 'tanzanite-settings' ),
					'copyPayloadLabel'        => __( '复制', 'tanzanite-settings' ),
					'deleteLabel'             => __( '删除', 'tanzanite-settings' ),
					'stickLabel'              => __( '置顶', 'tanzanite-settings' ),
					'unstickLabel'            => __( '取消置顶', 'tanzanite-settings' ),
					'stickyBadge'             => __( '置顶', 'tanzanite-settings' ),
					'memberPriceLabel'        => __( '会员价', 'tanzanite-settings' ),
					'stockAlertLabel'         => __( '警戒', 'tanzanite-settings' ),
					'pointsLimitLabel'        => __( '限制', 'tanzanite-settings' ),
					'pointsRewardLabel'       => __( '奖励积分', 'tanzanite-settings' ),
					'priceLabel'              => __( '价格', 'tanzanite-settings' ),
					'stockLabel'              => __( '库存', 'tanzanite-settings' ),
					'categoryLabel'           => __( '分类', 'tanzanite-settings' ),
					'statusLabel'             => __( '状态', 'tanzanite-settings' ),
					'deleteConfirm'           => __( '确认删除此商品吗？', 'tanzanite-settings' ),
					'deleteFailed'            => __( '删除失败', 'tanzanite-settings' ),
					'deleteSuccess'           => __( '删除成功', 'tanzanite-settings' ),
					'stickyFailed'            => __( '置顶操作失败', 'tanzanite-settings' ),
					'stickySuccess'           => __( '置顶操作成功', 'tanzanite-settings' ),
					'copyFailed'              => __( '复制失败', 'tanzanite-settings' ),
					'copySuccess'             => __( '已复制到剪贴板', 'tanzanite-settings' ),
					'bulkNoSelection'         => __( '请先选择商品', 'tanzanite-settings' ),
					'bulkDeleteConfirm'       => __( '确认删除选中的商品吗？', 'tanzanite-settings' ),
					'bulkDeleteFailed'        => __( '批量删除失败', 'tanzanite-settings' ),
					'bulkDeleteSuccess'       => __( '批量删除成功', 'tanzanite-settings' ),
					'bulkPriceNotImplemented' => __( '批量价格调整功能开发中', 'tanzanite-settings' ),
					'taxonomyLoading'         => __( '加载中...', 'tanzanite-settings' ),
					'taxonomyHasMore'         => __( '点击加载更多', 'tanzanite-settings' ),
					'taxonomyNoMore'          => __( '已加载全部', 'tanzanite-settings' ),
					'taxonomyEmpty'           => __( '没有匹配的结果', 'tanzanite-settings' ),
					'taxonomyLoadFailed'      => __( '加载失败，请稍后重试', 'tanzanite-settings' ),
				),
			)
		);

		echo '<style>'
			. ' .tz-products-filters { display:grid; gap:16px; max-width:1400px; }'
			. ' .tz-fieldset { background:#fff; border:1px solid #dcdcde; border-radius:6px; padding:16px; }'
			. ' .tz-fieldset-title { font-weight:600; font-size:14px; margin-bottom:12px; }'
			. ' .tz-field-grid { display:grid; gap:12px; grid-template-columns:repeat(auto-fit,minmax(220px,1fr)); }'
			. ' .tz-field label { display:flex; flex-direction:column; gap:4px; font-weight:500; }'
			. ' .tz-field label input, .tz-field label select, .tz-field label textarea { font-weight:400; }'
			. ' .tz-field .description { font-size:12px; color:#646970; line-height:1.4; }'
			. ' #tz-products-bulk-panel { grid-template-columns:repeat(auto-fit,minmax(260px,1fr)); margin-bottom:16px; gap:16px; }'
			. ' .tz-bulk-card { background:#fff; border:1px solid #dcdcde; border-radius:6px; padding:16px; display:flex; flex-direction:column; gap:12px; box-shadow:0 1px 2px rgba(0,0,0,0.04); }'
			. ' .tz-bulk-card h3 { margin:0; font-size:15px; }'
			. ' .tz-bulk-card form textarea { font-family:monospace; min-height:72px; }'
			. ' .tz-bulk-grid { display:grid; gap:8px; }'
			. ' .tz-bulk-grid label { display:flex; flex-direction:column; gap:4px; font-weight:500; }'
			. ' .tz-toolbar-note { color:#646970; font-size:12px; margin-left:auto; }'
			. ' .tz-inline-actions { display:flex; flex-wrap:wrap; gap:8px; align-items:center; }'
			. ' .tz-filters-toggle { display:flex; align-items:center; gap:8px; margin-bottom:12px; }'
			. ' .tz-filters-toggle button { display:flex; align-items:center; gap:6px; }'
			. ' .tz-filters-toggle .dashicons { display:inline-block; transition:transform 0.2s ease; }'
			. ' .tz-filters-collapsed { display:none; }'
			. ' .tz-taxonomy-toolbar { display:flex; flex-wrap:wrap; gap:8px; margin-top:8px; align-items:center; }'
			. ' .tz-taxonomy-toolbar .tz-taxonomy-search { width:200px; }'
			. ' .tz-taxonomy-toolbar .button-link { padding:0; }'
			. ' .tz-taxonomy-toolbar .tz-taxonomy-status { font-size:12px; color:#646970; }'
			. ' @media (max-width: 782px) { .tz-toolbar-note { display:none; } }'
			. '</style>';

		echo '<div class="tz-settings-wrapper tz-products-list">';

		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'All Products', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( '查看、筛选并管理商品，可切换表格 / 卡片视图，并快速执行常用操作。', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div id="tz-products-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

		echo '  <div class="tz-products-summary" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:16px;margin-bottom:24px;">';
		foreach ( [
			[ 'key' => 'total_products', 'label' => __( '商品总数', 'tanzanite-settings' ) ],
			[ 'key' => 'publish', 'label' => __( '已上架', 'tanzanite-settings' ) ],
			[ 'key' => 'draft', 'label' => __( '草稿', 'tanzanite-settings' ) ],
			[ 'key' => 'pending', 'label' => __( '待审核', 'tanzanite-settings' ) ],
			[ 'key' => 'low_stock', 'label' => __( '低库存', 'tanzanite-settings' ) ],
			[ 'key' => 'pending_reviews', 'label' => __( '待审核评价', 'tanzanite-settings' ) ],
		] as $card ) {
			echo '<div class="tz-dashboard-card" data-metric="' . esc_attr( $card['key'] ) . '">';
			echo '  <div class="tz-card-value">-</div>';
			echo '  <div class="tz-card-label">' . esc_html( $card['label'] ) . '</div>';
			echo '</div>';
		}
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
		echo '      <div class="tz-filters-toggle">';
		echo '          <button type="button" class="button" id="tz-products-filters-toggle" aria-expanded="true">';
		echo '              <span class="dashicons dashicons-arrow-down"></span><span class="tz-toggle-label">' . esc_html__( '收起筛选项', 'tanzanite-settings' ) . '</span>';
		echo '          </button>';
		echo '          <span class="tz-toolbar-note">' . esc_html__( '可折叠筛选区域，节省页面空间。', 'tanzanite-settings' ) . '</span>';
		echo '      </div>';
		echo '      <form id="tz-products-filters" class="tz-products-filters" aria-hidden="false">';

		echo '          <div class="tz-fieldset">';
		echo '              <div class="tz-fieldset-title">' . esc_html__( '基础筛选', 'tanzanite-settings' ) . '</div>';
		echo '              <div class="tz-field-grid">';
		echo '                  <div class="tz-field"><label>' . esc_html__( '关键词（标题）', 'tanzanite-settings' ) . '<input type="text" name="keyword" class="widefat" placeholder="' . esc_attr__( '商品名称关键字', 'tanzanite-settings' ) . '" /></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( 'SKU 编码', 'tanzanite-settings' ) . '<input type="text" name="sku" class="widefat" placeholder="SKU-001" /><span class="description">' . esc_html__( '支持模糊匹配，自动遍历 SKU 表。', 'tanzanite-settings' ) . '</span></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select name="status" class="widefat"><option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
		foreach ( self::ALLOWED_PRODUCT_STATUSES as $status ) {
			echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
		}
		echo '                  </select></label></div>';
		echo '                  <div class="tz-field tz-taxonomy" data-taxonomy="category">';
		echo '                      <label>' . esc_html__( '分类', 'tanzanite-settings' ) . '<select name="category" class="widefat" id="tz-filter-category"><option value="">' . esc_html__( '全部分类', 'tanzanite-settings' ) . '</option></select><span class="description">' . esc_html__( '可搜索分类名称并分页加载。', 'tanzanite-settings' ) . '</span></label>';
		echo '                      <div class="tz-taxonomy-toolbar">';
		echo '                          <input type="search" class="tz-taxonomy-search regular-text" data-taxonomy="category" placeholder="' . esc_attr__( '搜索分类…', 'tanzanite-settings' ) . '" />';
		echo '                          <button type="button" class="button tz-taxonomy-search-btn" data-taxonomy="category">' . esc_html__( '搜索', 'tanzanite-settings' ) . '</button>';
		echo '                          <button type="button" class="button tz-taxonomy-load-more" data-taxonomy="category">' . esc_html__( '加载更多', 'tanzanite-settings' ) . '</button>';
		echo '                          <span class="tz-taxonomy-status" data-taxonomy="category"></span>';
		echo '                      </div>';
		echo '                  </div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '作者 ID', 'tanzanite-settings' ) . '<input type="number" name="author" class="widefat" min="0" placeholder="0" /><span class="description">' . esc_html__( '可用于筛选创建人，0 表示忽略。', 'tanzanite-settings' ) . '</span></label></div>';
		echo '              </div>';
		echo '          </div>';

		echo '          <div class="tz-fieldset">';
		echo '              <div class="tz-fieldset-title">' . esc_html__( '库存与积分', 'tanzanite-settings' ) . '</div>';
		echo '              <div class="tz-field-grid">';
		echo '                  <div class="tz-field"><label>' . esc_html__( '库存下限', 'tanzanite-settings' ) . '<input type="number" name="inventory_min" class="widefat" placeholder="0" /></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '库存上限', 'tanzanite-settings' ) . '<input type="number" name="inventory_max" class="widefat" placeholder="9999" /></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '积分下限', 'tanzanite-settings' ) . '<input type="number" name="points_min" class="widefat" placeholder="0" /></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '积分上限', 'tanzanite-settings' ) . '<input type="number" name="points_max" class="widefat" placeholder="9999" /></label></div>';
		echo '              </div>';
		echo '          </div>';

		echo '          <div class="tz-fieldset">';
		echo '              <div class="tz-fieldset-title">' . esc_html__( '高级筛选', 'tanzanite-settings' ) . '</div>';
		echo '              <div class="tz-field-grid">';
		echo '                  <div class="tz-field tz-taxonomy" data-taxonomy="tags">';
		echo '                      <label>' . esc_html__( '标签', 'tanzanite-settings' ) . '<select multiple name="tags[]" class="widefat" id="tz-filter-tags" data-placeholder="tag-a,tag-b"></select><span class="description">' . esc_html__( '支持多选，使用 Ctrl/Command 选择多个标签，可搜索加载更多。', 'tanzanite-settings' ) . '</span></label>';
		echo '                      <div class="tz-taxonomy-toolbar">';
		echo '                          <input type="search" class="tz-taxonomy-search regular-text" data-taxonomy="tags" placeholder="' . esc_attr__( '搜索标签…', 'tanzanite-settings' ) . '" />';
		echo '                          <button type="button" class="button tz-taxonomy-search-btn" data-taxonomy="tags">' . esc_html__( '搜索', 'tanzanite-settings' ) . '</button>';
		echo '                          <button type="button" class="button tz-taxonomy-load-more" data-taxonomy="tags">' . esc_html__( '加载更多', 'tanzanite-settings' ) . '</button>';
		echo '                          <span class="tz-taxonomy-status" data-taxonomy="tags"></span>';
		echo '                      </div>';
		echo '                  </div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '属性筛选', 'tanzanite-settings' ) . '<input type="text" name="attributes" class="widefat" placeholder="pa_color:red,pa_size:xl" /><span class="description">' . esc_html__( '格式为 taxonomy:term_slug，可填写多个，用逗号分隔。', 'tanzanite-settings' ) . '</span></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '排序字段', 'tanzanite-settings' ) . '<select name="sort" class="widefat">';
		foreach ( [ 'updated_at' => __( '更新时间', 'tanzanite-settings' ), 'price_regular' => __( '原价', 'tanzanite-settings' ), 'stock_qty' => __( '库存', 'tanzanite-settings' ), 'points_reward' => __( '积分奖励', 'tanzanite-settings' ) ] as $key => $label ) {
			echo '<option value="' . esc_attr( $key ) . '">' . esc_html( $label ) . '</option>';
		}
		echo '                  </select></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '排序方式', 'tanzanite-settings' ) . '<select name="order" class="widefat"><option value="DESC">DESC</option><option value="ASC">ASC</option></select></label></div>';
		echo '                  <div class="tz-field"><label>' . esc_html__( '每页数量', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select><span class="description">' . esc_html__( '默认 20 条，最大支持 200 条。', 'tanzanite-settings' ) . '</span></label></div>';
		echo '              </div>';
		echo '          </div>';

		echo '      </form>';
		echo '      <div class="tz-inline-actions" style="margin-top:4px;">';
		echo '          <button class="button button-primary" id="tz-products-filter-submit">' . esc_html__( '应用筛选', 'tanzanite-settings' ) . '</button>';
		echo '          <button class="button" id="tz-products-filter-reset">' . esc_html__( '重置条件', 'tanzanite-settings' ) . '</button>';
		if ( $can_bulk ) {
			echo '          <a class="button" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-sku-importer' ) ) . '">' . esc_html__( '前往批量工具', 'tanzanite-settings' ) . '</a>';
		}
		if ( $can_manage ) {
			echo '          <a class="button button-secondary" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-add-product' ) ) . '">' . esc_html__( '新建商品', 'tanzanite-settings' ) . '</a>';
		}
		echo '          <span class="tz-toolbar-note">' . esc_html__( '提示：更多筛选项将与 Nuxt 前端保持一致，可随时扩展。', 'tanzanite-settings' ) . '</span>';
		echo '      </div>';
		echo '  </div>';

		echo '  <div class="tz-settings-section" id="tz-products-view">';
		echo '      <div class="tz-products-toolbar" style="display:flex;flex-wrap:wrap;gap:12px;align-items:center;margin-bottom:12px;">';
		echo '          <div class="button-group" role="group">';
		echo '              <button type="button" class="button button-secondary tz-view-toggle is-active" data-view="table">' . esc_html__( '表格视图', 'tanzanite-settings' ) . '</button>';
		echo '              <button type="button" class="button button-secondary tz-view-toggle" data-view="cards">' . esc_html__( '卡片视图', 'tanzanite-settings' ) . '</button>';
		echo '          </div>';
		echo '          <button type="button" class="button" id="tz-products-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
		if ( $can_bulk ) {
			echo '          <button type="button" class="button" id="tz-products-bulk-toggle" aria-expanded="false">' . esc_html__( '批量操作', 'tanzanite-settings' ) . '</button>';
		}
		echo '      </div>';

		if ( $can_bulk ) {
			echo '      <div id="tz-products-bulk-panel" style="display:none;gap:16px;margin-bottom:16px;">';
			echo '          <div class="tz-bulk-card">';
			echo '              <h3>' . esc_html__( '批量上下架', 'tanzanite-settings' ) . '</h3>';
			echo '              <form class="tz-bulk-form" data-action="set_status">';
			echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
			echo '                  <select name="status" class="widefat">';
			foreach ( self::ALLOWED_PRODUCT_STATUSES as $status ) {
				echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
			}
			echo '                  </select>';
			echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量更新状态', 'tanzanite-settings' ) . '</button>';
			echo '              </form>';
			echo '          </div>';
			echo '          <div class="tz-bulk-card">';
			echo '              <h3>' . esc_html__( '批量调整库存', 'tanzanite-settings' ) . '</h3>';
			echo '              <form class="tz-bulk-form" data-action="adjust_stock">';
			echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
			echo '                  <input type="number" class="widefat" name="delta" placeholder="±10" />';
			echo '                  <span class="description">' . esc_html__( '正数为增加库存，负数为扣减库存。', 'tanzanite-settings' ) . '</span>';
			echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量调整库存', 'tanzanite-settings' ) . '</button>';
			echo '              </form>';
			echo '          </div>';
			echo '          <div class="tz-bulk-card">';
			echo '              <h3>' . esc_html__( '批量调价', 'tanzanite-settings' ) . '</h3>';
			echo '              <form class="tz-bulk-form" data-action="adjust_price">';
			echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
			echo '                  <div class="tz-bulk-grid">';
			echo '                      <label>' . esc_html__( '调价模式', 'tanzanite-settings' ) . '<select name="mode" class="widefat"><option value="absolute">' . esc_html__( '固定值（元）', 'tanzanite-settings' ) . '</option><option value="percent">' . esc_html__( '百分比（%）', 'tanzanite-settings' ) . '</option></select></label>';
			echo '                      <label>' . esc_html__( '调价幅度', 'tanzanite-settings' ) . '<input type="number" step="0.01" name="value" class="widefat" placeholder="10 或 5" /></label>';
			echo '                      <label>' . esc_html__( '小数位数', 'tanzanite-settings' ) . '<select name="round" class="widefat"><option value="2">2</option><option value="1">1</option><option value="0">0</option><option value="3">3</option><option value="4">4</option></select><span class="description">' . esc_html__( '调价后的数值会根据此设置四舍五入。', 'tanzanite-settings' ) . '</span></label>';
			echo '                  </div>';
			echo '                  <fieldset style="border:1px solid #dcdcde;border-radius:4px;padding:12px;">';
			echo '                      <legend>' . esc_html__( '选择需要调价的字段', 'tanzanite-settings' ) . '</legend>';
			foreach ( [
				'price_regular' => __( '原价', 'tanzanite-settings' ),
				'price_sale'    => __( '现价', 'tanzanite-settings' ),
				'price_member'  => __( '会员价', 'tanzanite-settings' ),
			] as $field => $label ) {
				echo '                      <label style="display:flex;align-items:center;gap:6px;margin-bottom:6px;"><input type="checkbox" name="fields[]" value="' . esc_attr( $field ) . '"> ' . esc_html( $label ) . '</label>';
			}
			echo '                      <span class="description">' . esc_html__( '百分比模式示例：输入 5 即在原有价格基础上 +5%。可填写负值表示降价。', 'tanzanite-settings' ) . '</span>';
			echo '                  </fieldset>';
			echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量调价', 'tanzanite-settings' ) . '</button>';
			echo '              </form>';
			echo '          </div>';
			echo '          <div class="tz-bulk-card">';
			echo '              <h3>' . esc_html__( '批量设定字段', 'tanzanite-settings' ) . '</h3>';
			echo '              <form class="tz-bulk-form" data-action="set_meta">';
			echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
			echo '                  <div class="tz-bulk-grid">';
			foreach ( [
				'price_regular' => __( '原价', 'tanzanite-settings' ),
				'price_sale'    => __( '现价', 'tanzanite-settings' ),
				'price_member'  => __( '会员价', 'tanzanite-settings' ),
				'stock_qty'     => __( '库存', 'tanzanite-settings' ),
				'points_reward' => __( '积分奖励', 'tanzanite-settings' ),
				'points_limit'  => __( '积分上限', 'tanzanite-settings' ),
			] as $field => $label ) {
				echo '<label>' . esc_html( $label ) . '<input type="number" step="0.01" name="' . esc_attr( $field ) . '" class="widefat" /></label>';
			}
			echo '                  </div>';
			echo '                  <span class="description">' . esc_html__( '直接覆盖所选字段的值，可用于一次性设定库存或积分。', 'tanzanite-settings' ) . '</span>';
			echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量更新字段', 'tanzanite-settings' ) . '</button>';
			echo '              </form>';
			echo '          </div>';
			echo '          <div class="tz-bulk-card">';
			echo '              <h3>' . esc_html__( '批量设置推荐位', 'tanzanite-settings' ) . '</h3>';
			echo '              <form class="tz-bulk-form" data-action="set_featured">';
			echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
			echo '                  <select name="enabled" class="widefat">';
			echo '                      <option value="1">' . esc_html__( '设为推荐', 'tanzanite-settings' ) . '</option>';
			echo '                      <option value="0">' . esc_html__( '取消推荐', 'tanzanite-settings' ) . '</option>';
			echo '                  </select>';
			echo '                  <input type="text" class="widefat" name="slot" placeholder="Homepage-Top" />';
			echo '                  <span class="description">' . esc_html__( '可选的推荐位标识，取消推荐时可留空。', 'tanzanite-settings' ) . '</span>';
			echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量设置推荐位', 'tanzanite-settings' ) . '</button>';
			echo '              </form>';
			echo '          </div>';
			echo '          <div class="tz-bulk-card">';
			echo '              <h3>' . esc_html__( '批量删除', 'tanzanite-settings' ) . '</h3>';
			echo '              <form class="tz-bulk-form" data-action="delete">';
			echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
			echo '                  <select name="mode" class="widefat">';
			echo '                      <option value="trash">' . esc_html__( '移动到回收站（可恢复）', 'tanzanite-settings' ) . '</option>';
			echo '                      <option value="force">' . esc_html__( '永久删除（不可恢复）', 'tanzanite-settings' ) . '</option>';
			echo '                  </select>';
			echo '                  <span class="description">' . esc_html__( '建议先移至回收站以便回滚；永久删除将同步移除 SKU 数据。', 'tanzanite-settings' ) . '</span>';
			echo '                  <button type="submit" class="button button-secondary">' . esc_html__( '批量删除', 'tanzanite-settings' ) . '</button>';
			echo '              </form>';
			echo '          </div>';
			echo '          <div class="tz-bulk-card">';
			echo '              <h3>' . esc_html__( '批量导出', 'tanzanite-settings' ) . '</h3>';
			echo '              <form class="tz-bulk-form" data-action="export">';
			echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
			echo '                  <button type="submit" class="button">' . esc_html__( '导出选中商品', 'tanzanite-settings' ) . '</button>';
			echo '              </form>';
			echo '          </div>';
			echo '      </div>';
		}

		echo '      <div class="tz-products-table-wrapper" style="overflow:auto;">';
		echo '          <table class="widefat fixed striped" id="tz-products-table" style="min-width:1200px;">';
		echo '              <thead><tr>';
		foreach ( [ __( '商品信息', 'tanzanite-settings' ), __( '价格', 'tanzanite-settings' ), __( '库存', 'tanzanite-settings' ), __( '积分', 'tanzanite-settings' ), __( '分类', 'tanzanite-settings' ), __( '状态', 'tanzanite-settings' ), __( '更新时间', 'tanzanite-settings' ), __( '操作', 'tanzanite-settings' ) ] as $column ) {
			echo '<th>' . esc_html( $column ) . '</th>';
		}
		echo '              </tr></thead><tbody></tbody>';
		echo '          </table>';
		echo '      </div>';

		echo '      <div class="tz-products-cards" id="tz-products-cards" style="display:none;gap:16px;flex-wrap:wrap;"></div>';

		echo '      <div class="tz-products-pagination" style="display:flex;align-items:center;gap:12px;margin-top:16px;">';
		echo '          <button class="button" id="tz-products-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
		echo '          <span id="tz-products-page-info">1/1</span>';
		echo '          <button class="button" id="tz-products-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
		echo '      </div>';
		echo '  </div>';

		echo '</div>';
	}
}
