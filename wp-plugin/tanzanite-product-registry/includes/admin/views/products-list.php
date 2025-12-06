<?php
/**
 * 产品列表页面
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

$db = new Tanzanite_PR_Database();
$types = $db->get_product_types( true );
?>
<div class="wrap tanzanite-pr-wrap">
	<h1 class="wp-heading-inline">所有产品</h1>
	<a href="<?php echo esc_url( admin_url( 'admin.php?page=tanzanite-pr-add' ) ); ?>" class="page-title-action">添加产品</a>
	<hr class="wp-header-end">

	<!-- 筛选栏 -->
	<div class="tanzanite-pr-filters">
		<input type="text" id="pr-search" placeholder="搜索编码/名称/客户/订单号..." class="regular-text">
		<select id="pr-type-filter">
			<option value="">所有类型</option>
			<?php foreach ( $types as $type ) : ?>
				<option value="<?php echo esc_attr( $type['id'] ); ?>"><?php echo esc_html( $type['type_name'] ); ?></option>
			<?php endforeach; ?>
		</select>
		<button type="button" id="pr-search-btn" class="button">搜索</button>
		<button type="button" id="pr-export-btn" class="button">导出 Excel</button>
	</div>

	<!-- 产品列表 -->
	<table class="wp-list-table widefat fixed striped" id="pr-products-table">
		<thead>
			<tr>
				<th class="column-code" style="width: 120px;">编码</th>
				<th class="column-type" style="width: 80px;">类型</th>
				<th class="column-name">名称</th>
				<th class="column-ship" style="width: 100px;">出货日期</th>
				<th class="column-warranty" style="width: 100px;">保修至</th>
				<th class="column-status" style="width: 80px;">状态</th>
				<th class="column-customer" style="width: 100px;">客户</th>
				<th class="column-order" style="width: 100px;">订单号</th>
				<th class="column-actions" style="width: 100px;">操作</th>
			</tr>
		</thead>
		<tbody id="pr-products-body">
			<tr><td colspan="9" style="text-align: center;">加载中...</td></tr>
		</tbody>
	</table>

	<!-- 分页 -->
	<div class="tanzanite-pr-pagination" id="pr-pagination"></div>
</div>

<script>
jQuery(document).ready(function($) {
	let currentPage = 1;
	const perPage = 20;

	// 加载产品列表
	function loadProducts() {
		const search = $('#pr-search').val();
		const typeId = $('#pr-type-filter').val();

		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_get_products',
				nonce: tanzanitePR.nonce,
				search: search,
				type_id: typeId,
				page: currentPage,
				per_page: perPage
			},
			success: function(response) {
				if (response.success) {
					renderProducts(response.data.items);
					renderPagination(response.data.total, response.data.pages);
				}
			}
		});
	}

	// 渲染产品列表
	function renderProducts(items) {
		const tbody = $('#pr-products-body');
		tbody.empty();

		if (items.length === 0) {
			tbody.append('<tr><td colspan="9" style="text-align: center;">暂无数据</td></tr>');
			return;
		}

		items.forEach(function(item) {
			const statusClass = item.warranty_status === 'valid' ? 'status-valid' : 'status-expired';
			const statusText = item.warranty_status === 'valid' ? '有效' : '已过期';
			
			const row = `
				<tr data-id="${item.id}">
					<td><strong>${escapeHtml(item.product_code)}</strong></td>
					<td>${escapeHtml(item.type_name || '')}</td>
					<td>${escapeHtml(item.product_name || '-')}</td>
					<td>${item.ship_date ? item.ship_date.substring(0, 7) : '-'}</td>
					<td>${item.warranty_end || '-'}</td>
					<td><span class="pr-status ${statusClass}">${statusText}</span></td>
					<td>${escapeHtml(item.customer_name || '-')}</td>
					<td>${escapeHtml(item.order_id || '-')}</td>
					<td>
						<a href="admin.php?page=tanzanite-pr-add&id=${item.id}" class="button button-small">编辑</a>
						<button type="button" class="button button-small button-link-delete pr-delete-btn" data-id="${item.id}">删除</button>
					</td>
				</tr>
			`;
			tbody.append(row);
		});
	}

	// 渲染分页
	function renderPagination(total, pages) {
		const container = $('#pr-pagination');
		container.empty();

		if (pages <= 1) return;

		let html = '<span class="displaying-num">' + total + ' 个项目</span>';
		html += '<span class="pagination-links">';

		if (currentPage > 1) {
			html += '<a class="prev-page button" data-page="' + (currentPage - 1) + '">‹</a>';
		}

		html += '<span class="paging-input">' + currentPage + ' / ' + pages + '</span>';

		if (currentPage < pages) {
			html += '<a class="next-page button" data-page="' + (currentPage + 1) + '">›</a>';
		}

		html += '</span>';
		container.html(html);
	}

	// 转义 HTML
	function escapeHtml(text) {
		if (!text) return '';
		const div = document.createElement('div');
		div.textContent = text;
		return div.innerHTML;
	}

	// 搜索
	$('#pr-search-btn').on('click', function() {
		currentPage = 1;
		loadProducts();
	});

	$('#pr-search').on('keypress', function(e) {
		if (e.which === 13) {
			currentPage = 1;
			loadProducts();
		}
	});

	$('#pr-type-filter').on('change', function() {
		currentPage = 1;
		loadProducts();
	});

	// 分页
	$(document).on('click', '.pagination-links a', function() {
		currentPage = parseInt($(this).data('page'));
		loadProducts();
	});

	// 删除
	$(document).on('click', '.pr-delete-btn', function() {
		if (!confirm(tanzanitePR.i18n.confirmDelete)) return;

		const id = $(this).data('id');
		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_delete_product',
				nonce: tanzanitePR.nonce,
				id: id
			},
			success: function(response) {
				if (response.success) {
					loadProducts();
				} else {
					alert(response.data.message || tanzanitePR.i18n.error);
				}
			}
		});
	});

	// 初始加载
	loadProducts();
});
</script>
