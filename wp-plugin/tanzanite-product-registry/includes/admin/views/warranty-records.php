<?php
/**
 * 保修记录页面
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}
?>
<div class="wrap tanzanite-pr-wrap">
	<h1>保修记录</h1>
	<p class="description">在产品详情页中可以添加和管理保修记录。此页面显示所有产品的保修记录汇总。</p>

	<!-- 搜索 -->
	<div class="tanzanite-pr-filters">
		<input type="text" id="pr-record-search" placeholder="搜索产品编码..." class="regular-text">
		<select id="pr-record-type-filter">
			<option value="">所有类型</option>
			<option value="repair">维修</option>
			<option value="extend">延保</option>
			<option value="replace">换货</option>
		</select>
		<button type="button" id="pr-record-search-btn" class="button">搜索</button>
	</div>

	<table class="wp-list-table widefat fixed striped" id="pr-all-records-table">
		<thead>
			<tr>
				<th style="width: 120px;">产品编码</th>
				<th style="width: 100px;">产品类型</th>
				<th style="width: 80px;">记录类型</th>
				<th style="width: 100px;">记录日期</th>
				<th style="width: 80px;">延保月数</th>
				<th>描述</th>
				<th style="width: 100px;">操作人</th>
			</tr>
		</thead>
		<tbody id="pr-all-records-body">
			<tr><td colspan="7" style="text-align: center;">请输入产品编码搜索</td></tr>
		</tbody>
	</table>
</div>

<script>
jQuery(document).ready(function($) {
	const recordTypeNames = {
		'repair': '维修',
		'extend': '延保',
		'replace': '换货'
	};

	function searchRecords() {
		const search = $('#pr-record-search').val().trim();
		if (!search) {
			$('#pr-all-records-body').html('<tr><td colspan="7" style="text-align: center;">请输入产品编码搜索</td></tr>');
			return;
		}

		// 先搜索产品
		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_get_products',
				nonce: tanzanitePR.nonce,
				search: search,
				per_page: 100
			},
			success: function(response) {
				if (response.success && response.data.items.length > 0) {
					loadRecordsForProducts(response.data.items);
				} else {
					$('#pr-all-records-body').html('<tr><td colspan="7" style="text-align: center;">未找到匹配的产品</td></tr>');
				}
			}
		});
	}

	function loadRecordsForProducts(products) {
		const tbody = $('#pr-all-records-body');
		tbody.empty();

		let totalRecords = 0;
		let loadedProducts = 0;

		products.forEach(function(product) {
			$.ajax({
				url: tanzanitePR.ajaxUrl,
				type: 'POST',
				data: {
					action: 'tanzanite_pr_get_records',
					nonce: tanzanitePR.nonce,
					product_id: product.id
				},
				success: function(response) {
					loadedProducts++;

					if (response.success && response.data.records.length > 0) {
						const typeFilter = $('#pr-record-type-filter').val();
						
						response.data.records.forEach(function(record) {
							if (typeFilter && record.record_type !== typeFilter) return;

							totalRecords++;
							const row = `
								<tr>
									<td><a href="admin.php?page=tanzanite-pr-add&id=${product.id}">${product.product_code}</a></td>
									<td>${product.type_name || '-'}</td>
									<td>${recordTypeNames[record.record_type] || record.record_type}</td>
									<td>${record.record_date}</td>
									<td>${record.record_type === 'extend' ? '+' + record.extend_months + '个月' : '-'}</td>
									<td>${record.description || '-'}</td>
									<td>${record.operator || '-'}</td>
								</tr>
							`;
							tbody.append(row);
						});
					}

					// 所有产品加载完成
					if (loadedProducts === products.length && totalRecords === 0) {
						tbody.html('<tr><td colspan="7" style="text-align: center;">暂无保修记录</td></tr>');
					}
				}
			});
		});
	}

	$('#pr-record-search-btn').on('click', searchRecords);
	$('#pr-record-search').on('keypress', function(e) {
		if (e.which === 13) searchRecords();
	});
	$('#pr-record-type-filter').on('change', searchRecords);
});
</script>
