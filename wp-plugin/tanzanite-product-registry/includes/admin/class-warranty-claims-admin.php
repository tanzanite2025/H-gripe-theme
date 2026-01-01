<?php
/**
 * 保修申请管理
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 保修申请后台管理类
 */
class Tanzanite_PR_Warranty_Claims_Admin {

	/**
	 * 数据库实例
	 *
	 * @var Tanzanite_PR_Database
	 */
	private $db;

	/**
	 * 构造函数
	 */
	public function __construct() {
		$this->db = new Tanzanite_PR_Database();
		add_action( 'wp_ajax_tanzanite_pr_update_claim_status', array( $this, 'ajax_update_claim_status' ) );
	}

	/**
	 * 渲染列表页面
	 */
	public function render_list_page() {
		$current_page = isset( $_GET['paged'] ) ? max( 1, intval( $_GET['paged'] ) ) : 1;
		$search       = isset( $_GET['s'] ) ? sanitize_text_field( $_GET['s'] ) : '';
		$status       = isset( $_GET['status'] ) ? sanitize_text_field( $_GET['status'] ) : '';

		$args = array(
			'page'     => $current_page,
			'per_page' => 20,
			'search'   => $search,
			'status'   => $status,
		);

		$result = $this->db->get_warranty_claims( $args );
		
		?>
		<div class="wrap tanzanite-pr-wrap">
			<h1 class="wp-heading-inline">Warranty Claims (保修申请)</h1>
			<hr class="wp-header-end">

			<div class="tanzanite-pr-filter-bar">
				<form method="get">
					<input type="hidden" name="page" value="tanzanite-pr-warranty-claims">
					
					<div class="alignleft actions">
						<select name="status">
							<option value="">所有状态</option>
							<option value="pending" <?php selected( $status, 'pending' ); ?>>待处理 (Pending)</option>
							<option value="processing" <?php selected( $status, 'processing' ); ?>>处理中 (Processing)</option>
							<option value="approved" <?php selected( $status, 'approved' ); ?>>已批准 (Approved)</option>
							<option value="rejected" <?php selected( $status, 'rejected' ); ?>>已拒绝 (Rejected)</option>
						</select>
						
						<input type="search" name="s" value="<?php echo esc_attr( $search ); ?>" placeholder="搜索订单号或邮箱...">
						<input type="submit" class="button" value="筛选">
					</div>
				</form>
			</div>

			<table class="wp-list-table widefat fixed striped">
				<thead>
					<tr>
						<th width="60">ID</th>
						<th width="120">Date</th>
						<th width="150">Order #</th>
						<th>Email</th>
						<th width="120">Status</th>
						<th width="100">Actions</th>
					</tr>
				</thead>
				<tbody>
					<?php if ( ! empty( $result['items'] ) ) : ?>
						<?php foreach ( $result['items'] as $item ) : ?>
							<tr>
								<td>#<?php echo esc_html( $item['id'] ); ?></td>
								<td><?php echo esc_html( date( 'Y-m-d', strtotime( $item['created_at'] ) ) ); ?></td>
								<td><?php echo esc_html( $item['order_number'] ); ?></td>
								<td><a href="mailto:<?php echo esc_attr( $item['email'] ); ?>"><?php echo esc_html( $item['email'] ); ?></a></td>
								<td>
									<span class="tanzanite-pr-status tanzanite-pr-status-<?php echo esc_attr( $item['status'] ); ?>">
										<?php echo esc_html( ucfirst( $item['status'] ) ); ?>
									</span>
								</td>
								<td>
									<a href="<?php echo admin_url( 'admin.php?page=tanzanite-pr-warranty-claims&action=view&id=' . $item['id'] ); ?>" class="button button-small">View</a>
								</td>
							</tr>
						<?php endforeach; ?>
					<?php else : ?>
						<tr>
							<td colspan="6">暂无数据</td>
						</tr>
					<?php endif; ?>
				</tbody>
			</table>

			<?php if ( $result['pages'] > 1 ) : ?>
				<div class="tablenav bottom">
					<div class="tablenav-pages">
						<?php
						echo paginate_links( array(
							'base'      => add_query_arg( 'paged', '%#%' ),
							'format'    => '',
							'prev_text' => '&laquo;',
							'next_text' => '&raquo;',
							'total'     => $result['pages'],
							'current'   => $current_page,
						) );
						?>
					</div>
				</div>
			<?php endif; ?>
		</div>
		<?php
	}

	/**
	 * 渲染详情页面
	 */
	public function render_detail_page() {
		$id = isset( $_GET['id'] ) ? intval( $_GET['id'] ) : 0;
		$item = $this->db->get_warranty_claim( $id );

		if ( ! $item ) {
			echo '<div class="error"><p>Claim not found.</p></div>';
			return;
		}

		?>
		<div class="wrap tanzanite-pr-wrap">
			<h1 class="wp-heading-inline">
				Claim Details #<?php echo esc_html( $item['id'] ); ?>
				<a href="<?php echo admin_url( 'admin.php?page=tanzanite-pr-warranty-claims' ); ?>" class="page-title-action">Back to List</a>
			</h1>
			<hr class="wp-header-end">

			<div class="tanzanite-pr-claim-detail">
				<div class="postbox-container" style="display:flex; gap:20px; flex-wrap:wrap;">
					
					<!-- 左侧信息 -->
					<div class="main-content" style="flex:2; min-width:300px;">
						<div class="postbox">
							<div class="postbox-header"><h2 class="hndle">Basic Info</h2></div>
							<div class="inside">
								<table class="form-table">
									<tr>
										<th>Order Number</th>
										<td><?php echo esc_html( $item['order_number'] ); ?></td>
									</tr>
									<tr>
										<th>Email</th>
										<td><a href="mailto:<?php echo esc_attr( $item['email'] ); ?>"><?php echo esc_html( $item['email'] ); ?></a></td>
									</tr>
									<tr>
										<th>Submitted At</th>
										<td><?php echo esc_html( $item['created_at'] ); ?></td>
									</tr>
									<tr>
										<th>Tire Pressure</th>
										<td><?php echo esc_html( $item['tire_pressure'] ); ?></td>
									</tr>
									<tr>
										<th>Tubeless</th>
										<td><?php echo $item['is_tubeless'] ? 'Yes' : 'No'; ?></td>
									</tr>
									<tr>
										<th>Issue Description</th>
										<td><?php echo nl2br( esc_html( $item['issue_description'] ) ); ?></td>
									</tr>
								</table>
							</div>
						</div>

						<div class="postbox">
							<div class="postbox-header"><h2 class="hndle">Media</h2></div>
							<div class="inside">
								<?php if ( ! empty( $item['video_url'] ) ) : ?>
									<h3>Video</h3>
									<p>
										<a href="<?php echo esc_url( $item['video_url'] ); ?>" target="_blank" class="button">View Video</a>
										<code style="margin-left:10px;"><?php echo esc_html( $item['video_url'] ); ?></code>
									</p>
									<hr>
								<?php endif; ?>

								<h3>Images</h3>
								<?php if ( ! empty( $item['images'] ) && is_array( $item['images'] ) ) : ?>
									<div class="tanzanite-pr-gallery" style="display:grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap:10px;">
										<?php foreach ( $item['images'] as $img_url ) : ?>
											<a href="<?php echo esc_url( $img_url ); ?>" target="_blank" style="border:1px solid #ddd; padding:5px; display:block;">
												<img src="<?php echo esc_url( $img_url ); ?>" style="width:100%; height:150px; object-fit:cover;">
											</a>
										<?php endforeach; ?>
									</div>
								<?php else : ?>
									<p>No images uploaded.</p>
								<?php endif; ?>
							</div>
						</div>
					</div>

					<!-- 右侧操作 -->
					<div class="sidebar" style="flex:1; min-width:250px;">
						<div class="postbox">
							<div class="postbox-header"><h2 class="hndle">Action</h2></div>
							<div class="inside">
								<form id="tanzanite-pr-claim-form">
									<input type="hidden" name="id" value="<?php echo esc_attr( $item['id'] ); ?>">
									<input type="hidden" name="action" value="tanzanite_pr_update_claim_status">
									<?php wp_nonce_field( 'tanzanite_pr_nonce', 'nonce' ); ?>

									<p>
										<label><strong>Status:</strong></label><br>
										<select name="status" class="widefat">
											<option value="pending" <?php selected( $item['status'], 'pending' ); ?>>Pending</option>
											<option value="processing" <?php selected( $item['status'], 'processing' ); ?>>Processing</option>
											<option value="approved" <?php selected( $item['status'], 'approved' ); ?>>Approved</option>
											<option value="rejected" <?php selected( $item['status'], 'rejected' ); ?>>Rejected</option>
										</select>
									</p>
									<p>
										<label><strong>Admin Notes:</strong></label><br>
										<textarea name="admin_notes" rows="5" class="widefat"><?php echo esc_textarea( $item['admin_notes'] ); ?></textarea>
									</p>
									<p>
										<button type="submit" class="button button-primary large" style="width:100%;">Update Status</button>
									</p>
									<div class="tanzanite-pr-message"></div>
								</form>
							</div>
						</div>
					</div>
				</div>
			</div>
			
			<script>
			jQuery(document).ready(function($) {
				$('#tanzanite-pr-claim-form').on('submit', function(e) {
					e.preventDefault();
					var $form = $(this);
					var $btn = $form.find('button');
					var $msg = $form.find('.tanzanite-pr-message');

					$btn.prop('disabled', true).text('Updating...');
					
					$.post(ajaxurl, $form.serialize(), function(res) {
						if (res.success) {
							$msg.html('<p style="color:green;">' + res.data.message + '</p>');
						} else {
							$msg.html('<p style="color:red;">' + res.data.message + '</p>');
						}
					}).always(function() {
						$btn.prop('disabled', false).text('Update Status');
					});
				});
			});
			</script>
		</div>
		<?php
	}

	/**
	 * AJAX: 更新状态
	 */
	public function ajax_update_claim_status() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$id = intval( $_POST['id'] ?? 0 );
		$status = sanitize_text_field( $_POST['status'] ?? '' );
		$notes = sanitize_textarea_field( $_POST['admin_notes'] ?? '' );

		if ( ! $id ) {
			wp_send_json_error( array( 'message' => 'Invalid ID' ) );
		}

		$data = array(
			'id'          => $id,
			'status'      => $status,
			'admin_notes' => $notes,
		);

		$result = $this->db->save_warranty_claim( $data );

		if ( $result ) {
			wp_send_json_success( array( 'message' => 'Updated successfully.' ) );
		} else {
			wp_send_json_error( array( 'message' => 'Update failed.' ) );
		}
	}
}
