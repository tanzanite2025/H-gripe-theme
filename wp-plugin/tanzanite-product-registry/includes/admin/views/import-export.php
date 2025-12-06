<?php
/**
 * 批量导入导出页面
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
	<h1>批量导入/导出</h1>

	<div class="tanzanite-pr-two-col">
		<!-- 左侧：导入 -->
		<div class="tanzanite-pr-col-main">
			<div class="tanzanite-pr-box">
				<h3>📥 批量导入</h3>
				
				<div class="tanzanite-pr-import-section">
					<p class="description">支持 CSV 格式文件，请按照模板格式准备数据。</p>
					
					<div class="tanzanite-pr-template-download">
						<button type="button" id="pr-download-template" class="button">
							📄 下载导入模板
						</button>
					</div>

					<form id="pr-import-form" enctype="multipart/form-data">
						<div class="tanzanite-pr-upload-area" id="pr-upload-area">
							<div class="tanzanite-pr-upload-icon">📁</div>
							<p>拖拽文件到此处，或点击选择文件</p>
							<input type="file" id="pr-import-file" name="file" accept=".csv,.xlsx,.xls" style="display: none;">
							<button type="button" id="pr-select-file" class="button">选择文件</button>
						</div>
						<div id="pr-file-info" style="display: none;">
							<p><strong>已选择：</strong><span id="pr-file-name"></span></p>
							<button type="button" id="pr-clear-file" class="button button-link-delete">清除</button>
						</div>
					</form>

					<div id="pr-import-progress" style="display: none;">
						<div class="tanzanite-pr-progress-bar">
							<div class="tanzanite-pr-progress-fill"></div>
						</div>
						<p id="pr-import-status">正在导入...</p>
					</div>

					<div id="pr-import-result" style="display: none;">
						<div class="tanzanite-pr-result-box">
							<h4>导入结果</h4>
							<p id="pr-import-summary"></p>
							<div id="pr-import-errors"></div>
						</div>
					</div>

					<p class="submit">
						<button type="button" id="pr-start-import" class="button button-primary" disabled>开始导入</button>
					</p>
				</div>

				<hr>

				<h4>📋 导入说明</h4>
				<table class="widefat">
					<thead>
						<tr>
							<th>列名</th>
							<th>必填</th>
							<th>说明</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td><code>产品编码</code></td>
							<td>✅ 是</td>
							<td>唯一标识，字母和数字组合</td>
						</tr>
						<tr>
							<td><code>产品类型</code></td>
							<td>✅ 是</td>
							<td>hub / rim / wheelset / spoke / other</td>
						</tr>
						<tr>
							<td><code>出货日期</code></td>
							<td>✅ 是</td>
							<td>格式：2024-12 或 2024/12 或 2024年12月</td>
						</tr>
						<tr>
							<td><code>产品名称</code></td>
							<td>否</td>
							<td>产品型号或名称</td>
						</tr>
						<tr>
							<td><code>保修月数</code></td>
							<td>否</td>
							<td>默认 36 个月</td>
						</tr>
						<tr>
							<td><code>订单号</code></td>
							<td>否</td>
							<td>关联订单号</td>
						</tr>
						<tr>
							<td><code>客户姓名</code></td>
							<td>否</td>
							<td>客户姓名</td>
						</tr>
						<tr>
							<td><code>客户邮箱</code></td>
							<td>否</td>
							<td>客户邮箱</td>
						</tr>
						<tr>
							<td><code>客户电话</code></td>
							<td>否</td>
							<td>客户电话</td>
						</tr>
						<tr>
							<td><code>备注</code></td>
							<td>否</td>
							<td>备注信息</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		<!-- 右侧：导出 -->
		<div class="tanzanite-pr-col-side">
			<div class="tanzanite-pr-box">
				<h3>📤 数据导出</h3>
				
				<form id="pr-export-form">
					<p>
						<label for="export-type">产品类型</label>
						<select id="export-type" name="type_id" class="widefat">
							<option value="">全部类型</option>
							<?php foreach ( $types as $type ) : ?>
								<option value="<?php echo esc_attr( $type['id'] ); ?>">
									<?php echo esc_html( $type['type_name'] ); ?>
								</option>
							<?php endforeach; ?>
						</select>
					</p>
					<p>
						<label for="export-search">搜索筛选</label>
						<input type="text" id="export-search" name="search" class="widefat" placeholder="编码/名称/客户...">
					</p>
					<p class="submit">
						<button type="button" id="pr-start-export" class="button button-primary">导出 CSV</button>
					</p>
				</form>

				<hr>

				<h4>📊 数据统计</h4>
				<div id="pr-stats">
					<p>加载中...</p>
				</div>
			</div>
		</div>
	</div>
</div>

<script>
jQuery(document).ready(function($) {
	const uploadArea = $('#pr-upload-area');
	const fileInput = $('#pr-import-file');
	const fileInfo = $('#pr-file-info');
	const fileName = $('#pr-file-name');
	const importBtn = $('#pr-start-import');

	// 点击选择文件
	$('#pr-select-file').on('click', function() {
		fileInput.click();
	});

	// 文件选择变化
	fileInput.on('change', function() {
		const file = this.files[0];
		if (file) {
			showFileInfo(file);
		}
	});

	// 拖拽上传
	uploadArea.on('dragover', function(e) {
		e.preventDefault();
		$(this).addClass('dragover');
	}).on('dragleave', function() {
		$(this).removeClass('dragover');
	}).on('drop', function(e) {
		e.preventDefault();
		$(this).removeClass('dragover');
		const file = e.originalEvent.dataTransfer.files[0];
		if (file) {
			fileInput[0].files = e.originalEvent.dataTransfer.files;
			showFileInfo(file);
		}
	});

	function showFileInfo(file) {
		fileName.text(file.name + ' (' + formatFileSize(file.size) + ')');
		uploadArea.hide();
		fileInfo.show();
		importBtn.prop('disabled', false);
	}

	function formatFileSize(bytes) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}

	// 清除文件
	$('#pr-clear-file').on('click', function() {
		fileInput.val('');
		fileInfo.hide();
		uploadArea.show();
		importBtn.prop('disabled', true);
		$('#pr-import-result').hide();
	});

	// 下载模板
	$('#pr-download-template').on('click', function() {
		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_download_template',
				nonce: tanzanitePR.nonce
			},
			success: function(response) {
				if (response.success) {
					downloadCSV(response.data.csv, response.data.filename);
				}
			}
		});
	});

	// 开始导入
	importBtn.on('click', function() {
		const file = fileInput[0].files[0];
		if (!file) return;

		const formData = new FormData();
		formData.append('action', 'tanzanite_pr_import_products');
		formData.append('nonce', tanzanitePR.nonce);
		formData.append('file', file);

		$('#pr-import-progress').show();
		$('#pr-import-result').hide();
		importBtn.prop('disabled', true);

		// 模拟进度
		let progress = 0;
		const progressInterval = setInterval(function() {
			progress += Math.random() * 20;
			if (progress > 90) progress = 90;
			$('.tanzanite-pr-progress-fill').css('width', progress + '%');
		}, 200);

		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: formData,
			processData: false,
			contentType: false,
			success: function(response) {
				clearInterval(progressInterval);
				$('.tanzanite-pr-progress-fill').css('width', '100%');

				setTimeout(function() {
					$('#pr-import-progress').hide();
					$('#pr-import-result').show();

					const resultBox = $('.tanzanite-pr-result-box');
					if (response.success) {
						resultBox.removeClass('error').addClass('success');
						$('#pr-import-summary').html(response.data.message);

						if (response.data.errors && response.data.errors.length > 0) {
							let errorsHtml = '<p><strong>错误详情：</strong></p>';
							response.data.errors.forEach(function(err) {
								errorsHtml += '<p>• ' + err + '</p>';
							});
							$('#pr-import-errors').html(errorsHtml);
						} else {
							$('#pr-import-errors').empty();
						}
					} else {
						resultBox.removeClass('success').addClass('error');
						$('#pr-import-summary').text(response.data.message || '导入失败');
						$('#pr-import-errors').empty();
					}

					importBtn.prop('disabled', false);
				}, 500);
			},
			error: function() {
				clearInterval(progressInterval);
				$('#pr-import-progress').hide();
				alert('导入请求失败，请重试');
				importBtn.prop('disabled', false);
			}
		});
	});

	// 导出
	$('#pr-start-export').on('click', function() {
		const $btn = $(this);
		$btn.prop('disabled', true).text('导出中...');

		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_export_products',
				nonce: tanzanitePR.nonce,
				type_id: $('#export-type').val(),
				search: $('#export-search').val()
			},
			success: function(response) {
				if (response.success) {
					downloadCSV(response.data.csv, response.data.filename);
				} else {
					alert(response.data.message || '导出失败');
				}
				$btn.prop('disabled', false).text('导出 CSV');
			},
			error: function() {
				alert('导出请求失败');
				$btn.prop('disabled', false).text('导出 CSV');
			}
		});
	});

	// 下载 CSV 文件
	function downloadCSV(csv, filename) {
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
		const link = document.createElement('a');
		link.href = URL.createObjectURL(blob);
		link.download = filename;
		link.click();
	}

	// 加载统计
	function loadStats() {
		$.ajax({
			url: tanzanitePR.ajaxUrl,
			type: 'POST',
			data: {
				action: 'tanzanite_pr_get_products',
				nonce: tanzanitePR.nonce,
				per_page: 1
			},
			success: function(response) {
				if (response.success) {
					$('#pr-stats').html('<p>共 <strong>' + response.data.total + '</strong> 个产品</p>');
				}
			}
		});
	}

	loadStats();
});
</script>
