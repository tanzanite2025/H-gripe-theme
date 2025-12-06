<?php
/**
 * 产品类型管理页面
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

$db = new Tanzanite_PR_Database();
$types = $db->get_product_types();
?>
<div class="wrap tanzanite-pr-wrap">
	<h1>产品类型管理</h1>

	<div class="tanzanite-pr-two-col">
		<!-- 左侧：类型列表 -->
		<div class="tanzanite-pr-col-main">
			<table class="wp-list-table widefat fixed striped" id="pr-types-table">
				<thead>
					<tr>
						<th style="width: 80px;">代码</th>
						<th style="width: 100px;">中文名称</th>
						<th style="width: 100px;">英文名称</th>
						<th style="width: 80px;">默认保修</th>
						<th style="width: 60px;">排序</th>
						<th style="width: 60px;">状态</th>
						<th style="width: 100px;">操作</th>
					</tr>
				</thead>
				<tbody id="pr-types-body">
					<?php if ( empty( $types ) ) : ?>
						<tr><td colspan="7" style="text-align: center;">暂无数据</td></tr>
					<?php else : ?>
						<?php foreach ( $types as $type ) : ?>
							<tr data-id="<?php echo esc_attr( $type['id'] ); ?>">
								<td><code><?php echo esc_html( $type['type_code'] ); ?></code></td>
								<td><?php echo esc_html( $type['type_name'] ); ?></td>
								<td><?php echo esc_html( $type['type_name_en'] ); ?></td>
								<td><?php echo esc_html( $type['default_warranty_months'] ); ?>个月</td>
								<td><?php echo esc_html( $type['sort_order'] ); ?></td>
								<td>
									<?php if ( $type['is_active'] ) : ?>
										<span class="pr-status status-valid">启用</span>
									<?php else : ?>
										<span class="pr-status status-expired">禁用</span>
									<?php endif; ?>
								</td>
								<td>
									<button type="button" class="button button-small pr-edit-type-btn" 
										data-type='<?php echo esc_attr( json_encode( $type ) ); ?>'>编辑</button>
									<button type="button" class="button button-small button-link-delete pr-delete-type-btn" 
										data-id="<?php echo esc_attr( $type['id'] ); ?>">删除</button>
								</td>
							</tr>
						<?php endforeach; ?>
					<?php endif; ?>
				</tbody>
			</table>
		</div>

		<!-- 右侧：添加/编辑表单 -->
		<div class="tanzanite-pr-col-side">
			<div class="tanzanite-pr-box">
				<h3 id="pr-type-form-title">添加类型</h3>
				<form id="pr-type-form">
					<input type="hidden" name="id" value="">
					<p>
						<label for="type_code">类型代码 <span class="required">*</span></label>
						<input type="text" id="type_code" name="type_code" class="widefat" required>
						<span class="description">英文小写，如 hub, rim</span>
					</p>
					<p>
						<label for="type_name">中文名称 <span class="required">*</span></label>
						<input type="text" id="type_name" name="type_name" class="widefat" required>
					</p>
					<p>
						<label for="type_name_en">英文名称 <span class="required">*</span></label>
						<input type="text" id="type_name_en" name="type_name_en" class="widefat" required>
					</p>
					<p>
						<label for="default_warranty_months">默认保修期（月）</label>
						<input type="number" id="default_warranty_months" name="default_warranty_months" class="widefat" value="36" min="1" max="120">
					</p>
					<p>
						<label for="sort_order">排序</label>
						<input type="number" id="sort_order" name="sort_order" class="widefat" value="0" min="0">
						<span class="description">数字越小越靠前</span>
					</p>
					<p>
						<label>
							<input type="checkbox" id="is_active" name="is_active" checked> 启用
						</label>
					</p>
					<p class="submit">
						<button type="submit" class="button button-primary">保存</button>
						<button type="button" class="button" id="pr-reset-type-btn">重置</button>
					</p>
				</form>
			</div>
		</div>
	</div>
</div>

<script>
jQuery(document).ready(function($) {
	// 编辑类型
	$(document).on('click', '.pr-edit-type-btn', function() {
		const type = $(this).data('type');
		$('#pr-type-form-title').text('编辑类型');
		$('input[name="id"]').val(type.id);
		$('#type_code').val(type.type_code);
		$('#type_name').val(type.type_name);
		$('#type_name_en').val(type.type_name_en);
		$('#default_warranty_months').val(type.default_warranty_months);
		$('#sort_order').val(type.sort_order);
		$('#is_active').prop('checked', type.is_active == 1);
	});

	// 重置表单
	$('#pr-reset-type-btn').on('click', function() {
		$('#pr-type-form-title').text('添加类型');
		$('#pr-type-form')[0].reset();
		$('input[name="id"]').val('');
	});

	// 保存类型
	$('#pr-type-form').on('submit', function(e) {
		e.preventDefault();

		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_save_type',
				nonce: tanzanitePR.nonce,
				...Object.fromEntries(new FormData(this))
			},
			success: function(response) {
				if (response.success) {
					location.reload();
				} else {
					alert(response.data.message || tanzanitePR.i18n.error);
				}
			}
		});
	});

	// 删除类型
	$(document).on('click', '.pr-delete-type-btn', function() {
		if (!confirm(tanzanitePR.i18n.confirmDelete)) return;

		const id = $(this).data('id');
		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_delete_type',
				nonce: tanzanitePR.nonce,
				id: id
			},
			success: function(response) {
				if (response.success) {
					location.reload();
				} else {
					alert(response.data.message || tanzanitePR.i18n.error);
				}
			}
		});
	});
});
</script>
