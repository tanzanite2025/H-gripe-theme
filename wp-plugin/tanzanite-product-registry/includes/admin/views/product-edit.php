<?php
/**
 * 产品编辑页面
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

$db = new Tanzanite_PR_Database();
$types = $db->get_product_types( true );

$product_id = isset( $_GET['id'] ) ? intval( $_GET['id'] ) : 0;
$product = $product_id ? $db->get_product( $product_id ) : null;
$is_edit = ! empty( $product );

$page_title = $is_edit ? '编辑产品' : '添加产品';
?>
<div class="wrap tanzanite-pr-wrap">
	<h1><?php echo esc_html( $page_title ); ?></h1>

	<form id="pr-product-form" class="tanzanite-pr-form">
		<input type="hidden" name="id" value="<?php echo esc_attr( $product_id ); ?>">

		<table class="form-table">
			<tr>
				<th><label for="product_code">产品编码 <span class="required">*</span></label></th>
				<td>
					<input type="text" id="product_code" name="product_code" class="regular-text" 
						value="<?php echo esc_attr( $product['product_code'] ?? '' ); ?>" required>
					<p class="description">唯一标识，字母和数字组合</p>
				</td>
			</tr>
			<tr>
				<th><label for="product_type_id">产品类型 <span class="required">*</span></label></th>
				<td>
					<select id="product_type_id" name="product_type_id" required>
						<option value="">请选择</option>
						<?php foreach ( $types as $type ) : ?>
							<option value="<?php echo esc_attr( $type['id'] ); ?>" 
								data-warranty="<?php echo esc_attr( $type['default_warranty_months'] ); ?>"
								<?php selected( $product['product_type_id'] ?? '', $type['id'] ); ?>>
								<?php echo esc_html( $type['type_name'] ); ?>
							</option>
						<?php endforeach; ?>
					</select>
				</td>
			</tr>
			<tr>
				<th><label for="product_name">产品名称</label></th>
				<td>
					<input type="text" id="product_name" name="product_name" class="regular-text" 
						value="<?php echo esc_attr( $product['product_name'] ?? '' ); ?>">
					<p class="description">产品型号或名称</p>
				</td>
			</tr>
			<tr>
				<th><label for="ship_date">出货日期 <span class="required">*</span></label></th>
				<td>
					<input type="month" id="ship_date" name="ship_date" 
						value="<?php echo esc_attr( $product ? substr( $product['ship_date'], 0, 7 ) : '' ); ?>" required>
				</td>
			</tr>
			<tr>
				<th><label for="warranty_months">保修期限</label></th>
				<td>
					<input type="number" id="warranty_months" name="warranty_months" class="small-text" min="1" max="120"
						value="<?php echo esc_attr( $product['warranty_months'] ?? 36 ); ?>"> 个月
					<label style="margin-left: 15px;">
						<input type="checkbox" id="use_default_warranty"> 使用类型默认值
					</label>
				</td>
			</tr>
		</table>

		<h2>客户信息</h2>
		<table class="form-table">
			<tr>
				<th><label for="order_id">订单号</label></th>
				<td>
					<input type="text" id="order_id" name="order_id" class="regular-text" 
						value="<?php echo esc_attr( $product['order_id'] ?? '' ); ?>">
				</td>
			</tr>
			<tr>
				<th><label for="customer_name">客户姓名</label></th>
				<td>
					<input type="text" id="customer_name" name="customer_name" class="regular-text" 
						value="<?php echo esc_attr( $product['customer_name'] ?? '' ); ?>">
				</td>
			</tr>
			<tr>
				<th><label for="customer_email">客户邮箱</label></th>
				<td>
					<input type="email" id="customer_email" name="customer_email" class="regular-text" 
						value="<?php echo esc_attr( $product['customer_email'] ?? '' ); ?>">
				</td>
			</tr>
			<tr>
				<th><label for="customer_phone">客户电话</label></th>
				<td>
					<input type="text" id="customer_phone" name="customer_phone" class="regular-text" 
						value="<?php echo esc_attr( $product['customer_phone'] ?? '' ); ?>">
				</td>
			</tr>
			<tr>
				<th><label for="notes">备注</label></th>
				<td>
					<textarea id="notes" name="notes" rows="3" class="large-text"><?php echo esc_textarea( $product['notes'] ?? '' ); ?></textarea>
				</td>
			</tr>
		</table>

		<p class="submit">
			<button type="submit" class="button button-primary" id="pr-save-btn">保存</button>
			<a href="<?php echo esc_url( admin_url( 'admin.php?page=tanzanite-pr' ) ); ?>" class="button">返回列表</a>
		</p>
	</form>

	<?php if ( $is_edit ) : ?>
	<hr>
	<h2>保修记录</h2>
	<div id="pr-records-section">
		<button type="button" class="button" id="pr-add-record-btn">添加记录</button>
		<table class="wp-list-table widefat fixed striped" id="pr-records-table" style="margin-top: 10px;">
			<thead>
				<tr>
					<th style="width: 100px;">类型</th>
					<th style="width: 120px;">日期</th>
					<th style="width: 80px;">延保月数</th>
					<th>描述</th>
					<th style="width: 100px;">操作人</th>
					<th style="width: 80px;">操作</th>
				</tr>
			</thead>
			<tbody id="pr-records-body">
				<tr><td colspan="6" style="text-align: center;">加载中...</td></tr>
			</tbody>
		</table>
	</div>

	<!-- 添加记录弹窗 -->
	<div id="pr-record-modal" class="tanzanite-pr-modal" style="display: none;">
		<div class="tanzanite-pr-modal-content">
			<h3>添加保修记录</h3>
			<form id="pr-record-form">
				<input type="hidden" name="product_id" value="<?php echo esc_attr( $product_id ); ?>">
				<table class="form-table">
					<tr>
						<th><label for="record_type">记录类型</label></th>
						<td>
							<select id="record_type" name="record_type" required>
								<option value="repair">维修</option>
								<option value="extend">延保</option>
								<option value="replace">换货</option>
							</select>
						</td>
					</tr>
					<tr>
						<th><label for="record_date">记录日期</label></th>
						<td>
							<input type="date" id="record_date" name="record_date" required 
								value="<?php echo esc_attr( date( 'Y-m-d' ) ); ?>">
						</td>
					</tr>
					<tr id="extend-months-row" style="display: none;">
						<th><label for="extend_months">延保月数</label></th>
						<td>
							<input type="number" id="extend_months" name="extend_months" class="small-text" min="1" value="6"> 个月
						</td>
					</tr>
					<tr>
						<th><label for="description">描述</label></th>
						<td>
							<textarea id="description" name="description" rows="3" class="large-text"></textarea>
						</td>
					</tr>
				</table>
				<p class="submit">
					<button type="submit" class="button button-primary">保存</button>
					<button type="button" class="button" id="pr-cancel-record-btn">取消</button>
				</p>
			</form>
		</div>
	</div>
	<?php endif; ?>
</div>

<script>
jQuery(document).ready(function($) {
	// 类型切换时更新默认保修期
	$('#product_type_id').on('change', function() {
		if ($('#use_default_warranty').is(':checked')) {
			const warranty = $(this).find(':selected').data('warranty') || 36;
			$('#warranty_months').val(warranty);
		}
	});

	$('#use_default_warranty').on('change', function() {
		if ($(this).is(':checked')) {
			const warranty = $('#product_type_id').find(':selected').data('warranty') || 36;
			$('#warranty_months').val(warranty);
		}
	});

	// 保存产品
	$('#pr-product-form').on('submit', function(e) {
		e.preventDefault();

		const $btn = $('#pr-save-btn');
		$btn.prop('disabled', true).text(tanzanitePR.i18n.saving);

		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_save_product',
				nonce: tanzanitePR.nonce,
				...Object.fromEntries(new FormData(this))
			},
			success: function(response) {
				if (response.success) {
					$btn.text(tanzanitePR.i18n.saved);
					if (!$('input[name="id"]').val()) {
						// 新增成功，跳转到编辑页
						window.location.href = 'admin.php?page=tanzanite-pr-add&id=' + response.data.id;
					}
					setTimeout(() => $btn.prop('disabled', false).text('保存'), 1500);
				} else {
					alert(response.data.message || tanzanitePR.i18n.error);
					$btn.prop('disabled', false).text('保存');
				}
			},
			error: function() {
				alert(tanzanitePR.i18n.error);
				$btn.prop('disabled', false).text('保存');
			}
		});
	});

	<?php if ( $is_edit ) : ?>
	// 加载保修记录
	function loadRecords() {
		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_get_records',
				nonce: tanzanitePR.nonce,
				product_id: <?php echo esc_js( $product_id ); ?>
			},
			success: function(response) {
				if (response.success) {
					renderRecords(response.data.records);
				}
			}
		});
	}

	function renderRecords(records) {
		const tbody = $('#pr-records-body');
		tbody.empty();

		if (records.length === 0) {
			tbody.append('<tr><td colspan="6" style="text-align: center;">暂无记录</td></tr>');
			return;
		}

		records.forEach(function(record) {
			const row = `
				<tr data-id="${record.id}">
					<td>${record.record_type_name}</td>
					<td>${record.record_date}</td>
					<td>${record.record_type === 'extend' ? '+' + record.extend_months + '个月' : '-'}</td>
					<td>${record.description || '-'}</td>
					<td>${record.operator || '-'}</td>
					<td>
						<button type="button" class="button button-small button-link-delete pr-delete-record-btn" data-id="${record.id}">删除</button>
					</td>
				</tr>
			`;
			tbody.append(row);
		});
	}

	// 显示/隐藏延保月数
	$('#record_type').on('change', function() {
		$('#extend-months-row').toggle($(this).val() === 'extend');
	});

	// 添加记录弹窗
	$('#pr-add-record-btn').on('click', function() {
		$('#pr-record-modal').show();
	});

	$('#pr-cancel-record-btn').on('click', function() {
		$('#pr-record-modal').hide();
		$('#pr-record-form')[0].reset();
	});

	// 保存记录
	$('#pr-record-form').on('submit', function(e) {
		e.preventDefault();

		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_save_record',
				nonce: tanzanitePR.nonce,
				...Object.fromEntries(new FormData(this))
			},
			success: function(response) {
				if (response.success) {
					$('#pr-record-modal').hide();
					$('#pr-record-form')[0].reset();
					loadRecords();
				} else {
					alert(response.data.message || tanzanitePR.i18n.error);
				}
			}
		});
	});

	// 删除记录
	$(document).on('click', '.pr-delete-record-btn', function() {
		if (!confirm(tanzanitePR.i18n.confirmDelete)) return;

		const id = $(this).data('id');
		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_delete_record',
				nonce: tanzanitePR.nonce,
				id: id
			},
			success: function(response) {
				if (response.success) {
					loadRecords();
				} else {
					alert(response.data.message || tanzanitePR.i18n.error);
				}
			}
		});
	});

	// 初始加载
	loadRecords();
	<?php endif; ?>
});
</script>
