<?php
/**
 * 保修记录管理
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 保修记录后台管理类
 */
class Tanzanite_PR_Warranty_Records_Admin {

	/**
	 * 数据库实例
	 *
	 * @var Tanzanite_PR_Database
	 */
	private $db;

	/**
	 * 记录类型
	 *
	 * @var array
	 */
	public static $record_types = array(
		'repair'  => '维修',
		'extend'  => '延保',
		'replace' => '换货',
	);

	/**
	 * 构造函数
	 */
	public function __construct() {
		$this->db = new Tanzanite_PR_Database();
		add_action( 'wp_ajax_tanzanite_pr_save_record', array( $this, 'ajax_save_record' ) );
		add_action( 'wp_ajax_tanzanite_pr_delete_record', array( $this, 'ajax_delete_record' ) );
		add_action( 'wp_ajax_tanzanite_pr_get_records', array( $this, 'ajax_get_records' ) );
	}

	/**
	 * AJAX: 保存保修记录
	 */
	public function ajax_save_record() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$data = array(
			'id'            => isset( $_POST['id'] ) ? intval( $_POST['id'] ) : 0,
			'product_id'    => intval( $_POST['product_id'] ?? 0 ),
			'record_type'   => sanitize_text_field( $_POST['record_type'] ?? '' ),
			'record_date'   => sanitize_text_field( $_POST['record_date'] ?? '' ),
			'description'   => sanitize_textarea_field( $_POST['description'] ?? '' ),
			'extend_months' => intval( $_POST['extend_months'] ?? 0 ),
			'operator'      => sanitize_text_field( $_POST['operator'] ?? '' ),
		);

		// 验证
		if ( empty( $data['product_id'] ) ) {
			wp_send_json_error( array( 'message' => '产品ID不能为空' ) );
		}
		if ( ! in_array( $data['record_type'], array_keys( self::$record_types ), true ) ) {
			wp_send_json_error( array( 'message' => '无效的记录类型' ) );
		}
		if ( empty( $data['record_date'] ) ) {
			wp_send_json_error( array( 'message' => '记录日期不能为空' ) );
		}

		// 如果是延保，必须有延保月数
		if ( 'extend' === $data['record_type'] && $data['extend_months'] <= 0 ) {
			wp_send_json_error( array( 'message' => '延保月数必须大于0' ) );
		}

		// 如果不是延保，清空延保月数
		if ( 'extend' !== $data['record_type'] ) {
			$data['extend_months'] = 0;
		}

		// 默认操作人为当前用户
		if ( empty( $data['operator'] ) ) {
			$current_user = wp_get_current_user();
			$data['operator'] = $current_user->display_name ?: $current_user->user_login;
		}

		$result = $this->db->save_warranty_record( $data );

		if ( $result ) {
			wp_send_json_success( array( 'id' => $result, 'message' => '保存成功' ) );
		} else {
			wp_send_json_error( array( 'message' => '保存失败' ) );
		}
	}

	/**
	 * AJAX: 删除保修记录
	 */
	public function ajax_delete_record() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$id = intval( $_POST['id'] ?? 0 );
		if ( ! $id ) {
			wp_send_json_error( array( 'message' => '无效的ID' ) );
		}

		$result = $this->db->delete_warranty_record( $id );

		if ( $result ) {
			wp_send_json_success( array( 'message' => '删除成功' ) );
		} else {
			wp_send_json_error( array( 'message' => '删除失败' ) );
		}
	}

	/**
	 * AJAX: 获取产品的保修记录
	 */
	public function ajax_get_records() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$product_id = intval( $_POST['product_id'] ?? 0 );
		if ( ! $product_id ) {
			wp_send_json_error( array( 'message' => '产品ID不能为空' ) );
		}

		$records = $this->db->get_warranty_records( $product_id );

		// 添加类型名称
		foreach ( $records as &$record ) {
			$record['record_type_name'] = self::$record_types[ $record['record_type'] ] ?? $record['record_type'];
		}

		wp_send_json_success( array( 'records' => $records ) );
	}
}
