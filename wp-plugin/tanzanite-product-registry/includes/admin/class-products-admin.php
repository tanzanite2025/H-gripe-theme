<?php
/**
 * 产品管理
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 产品后台管理类
 */
class Tanzanite_PR_Products_Admin {

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
		add_action( 'wp_ajax_tanzanite_pr_save_product', array( $this, 'ajax_save_product' ) );
		add_action( 'wp_ajax_tanzanite_pr_delete_product', array( $this, 'ajax_delete_product' ) );
		add_action( 'wp_ajax_tanzanite_pr_get_products', array( $this, 'ajax_get_products' ) );
		add_action( 'wp_ajax_tanzanite_pr_get_product', array( $this, 'ajax_get_product' ) );
	}

	/**
	 * AJAX: 保存产品
	 */
	public function ajax_save_product() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$data = array(
			'id'              => isset( $_POST['id'] ) ? intval( $_POST['id'] ) : 0,
			'product_code'    => sanitize_text_field( $_POST['product_code'] ?? '' ),
			'product_type_id' => intval( $_POST['product_type_id'] ?? 0 ),
			'product_name'    => sanitize_text_field( $_POST['product_name'] ?? '' ),
			'ship_date'       => sanitize_text_field( $_POST['ship_date'] ?? '' ),
			'warranty_months' => intval( $_POST['warranty_months'] ?? 36 ),
			'order_id'        => sanitize_text_field( $_POST['order_id'] ?? '' ),
			'customer_name'   => sanitize_text_field( $_POST['customer_name'] ?? '' ),
			'customer_email'  => sanitize_email( $_POST['customer_email'] ?? '' ),
			'customer_phone'  => sanitize_text_field( $_POST['customer_phone'] ?? '' ),
			'notes'           => sanitize_textarea_field( $_POST['notes'] ?? '' ),
		);

		// 验证必填字段
		if ( empty( $data['product_code'] ) ) {
			wp_send_json_error( array( 'message' => '产品编码不能为空' ) );
		}
		if ( empty( $data['product_type_id'] ) ) {
			wp_send_json_error( array( 'message' => '请选择产品类型' ) );
		}
		if ( empty( $data['ship_date'] ) ) {
			wp_send_json_error( array( 'message' => '出货日期不能为空' ) );
		}

		// 格式化日期（确保是当月1日）
		$date_parts = explode( '-', $data['ship_date'] );
		if ( count( $date_parts ) >= 2 ) {
			$data['ship_date'] = $date_parts[0] . '-' . $date_parts[1] . '-01';
		}

		// 检查编码是否重复
		$existing = $this->db->get_product_by_code( $data['product_code'] );
		if ( $existing && $existing['id'] != $data['id'] ) {
			wp_send_json_error( array( 'message' => '产品编码已存在' ) );
		}

		$result = $this->db->save_product( $data );

		if ( $result ) {
			wp_send_json_success( array( 'id' => $result, 'message' => '保存成功' ) );
		} else {
			wp_send_json_error( array( 'message' => '保存失败' ) );
		}
	}

	/**
	 * AJAX: 删除产品
	 */
	public function ajax_delete_product() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$id = intval( $_POST['id'] ?? 0 );
		if ( ! $id ) {
			wp_send_json_error( array( 'message' => '无效的ID' ) );
		}

		$result = $this->db->delete_product( $id );

		if ( $result ) {
			wp_send_json_success( array( 'message' => '删除成功' ) );
		} else {
			wp_send_json_error( array( 'message' => '删除失败' ) );
		}
	}

	/**
	 * AJAX: 获取产品列表
	 */
	public function ajax_get_products() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$args = array(
			'search'   => sanitize_text_field( $_POST['search'] ?? '' ),
			'type_id'  => intval( $_POST['type_id'] ?? 0 ),
			'page'     => intval( $_POST['page'] ?? 1 ),
			'per_page' => intval( $_POST['per_page'] ?? 20 ),
			'orderby'  => sanitize_text_field( $_POST['orderby'] ?? 'created_at' ),
			'order'    => sanitize_text_field( $_POST['order'] ?? 'DESC' ),
		);

		$result = $this->db->get_products( $args );

		// 计算保修状态
		foreach ( $result['items'] as &$item ) {
			$item = $this->calculate_warranty_status( $item );
		}

		wp_send_json_success( $result );
	}

	/**
	 * AJAX: 获取单个产品
	 */
	public function ajax_get_product() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$id = intval( $_POST['id'] ?? 0 );
		if ( ! $id ) {
			wp_send_json_error( array( 'message' => '无效的ID' ) );
		}

		$product = $this->db->get_product( $id );
		if ( ! $product ) {
			wp_send_json_error( array( 'message' => '产品不存在' ) );
		}

		$product = $this->calculate_warranty_status( $product );
		$product['records'] = $this->db->get_warranty_records( $id );

		wp_send_json_success( array( 'product' => $product ) );
	}

	/**
	 * 计算保修状态
	 *
	 * @param array $product 产品数据
	 * @return array
	 */
	private function calculate_warranty_status( $product ) {
		$ship_date = new DateTime( $product['ship_date'] );
		$warranty_months = intval( $product['warranty_months'] );
		
		// 加上延保月数
		$extend_months = $this->db->get_total_extend_months( $product['id'] );
		$total_months = $warranty_months + $extend_months;

		$warranty_end = clone $ship_date;
		$warranty_end->modify( "+{$total_months} months" );

		$today = new DateTime();
		$diff = $today->diff( $warranty_end );

		$product['warranty_end'] = $warranty_end->format( 'Y-m' );
		$product['extend_months'] = $extend_months;
		$product['total_warranty_months'] = $total_months;

		if ( $warranty_end > $today ) {
			$product['warranty_status'] = 'valid';
			$product['remaining_months'] = ( $diff->y * 12 ) + $diff->m;
			$product['remaining_days'] = $diff->days;
		} else {
			$product['warranty_status'] = 'expired';
			$product['expired_months'] = ( $diff->y * 12 ) + $diff->m;
			$product['expired_days'] = $diff->days;
		}

		return $product;
	}
}
