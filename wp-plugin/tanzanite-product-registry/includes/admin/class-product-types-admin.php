<?php
/**
 * 产品类型管理
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 产品类型后台管理类
 */
class Tanzanite_PR_Product_Types_Admin {

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
		add_action( 'wp_ajax_tanzanite_pr_save_type', array( $this, 'ajax_save_type' ) );
		add_action( 'wp_ajax_tanzanite_pr_delete_type', array( $this, 'ajax_delete_type' ) );
		add_action( 'wp_ajax_tanzanite_pr_get_types', array( $this, 'ajax_get_types' ) );
	}

	/**
	 * AJAX: 保存产品类型
	 */
	public function ajax_save_type() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$data = array(
			'id'                      => isset( $_POST['id'] ) ? intval( $_POST['id'] ) : 0,
			'type_code'               => sanitize_text_field( $_POST['type_code'] ?? '' ),
			'type_name'               => sanitize_text_field( $_POST['type_name'] ?? '' ),
			'type_name_en'            => sanitize_text_field( $_POST['type_name_en'] ?? '' ),
			'default_warranty_months' => intval( $_POST['default_warranty_months'] ?? 36 ),
			'sort_order'              => intval( $_POST['sort_order'] ?? 0 ),
			'is_active'               => isset( $_POST['is_active'] ) ? 1 : 0,
		);

		if ( empty( $data['type_code'] ) || empty( $data['type_name'] ) ) {
			wp_send_json_error( array( 'message' => '类型代码和名称不能为空' ) );
		}

		$result = $this->db->save_product_type( $data );

		if ( $result ) {
			wp_send_json_success( array( 'id' => $result, 'message' => '保存成功' ) );
		} else {
			wp_send_json_error( array( 'message' => '保存失败' ) );
		}
	}

	/**
	 * AJAX: 删除产品类型
	 */
	public function ajax_delete_type() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$id = intval( $_POST['id'] ?? 0 );
		if ( ! $id ) {
			wp_send_json_error( array( 'message' => '无效的ID' ) );
		}

		$result = $this->db->delete_product_type( $id );

		if ( $result ) {
			wp_send_json_success( array( 'message' => '删除成功' ) );
		} else {
			wp_send_json_error( array( 'message' => '删除失败' ) );
		}
	}

	/**
	 * AJAX: 获取产品类型列表
	 */
	public function ajax_get_types() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$types = $this->db->get_product_types();
		wp_send_json_success( array( 'types' => $types ) );
	}
}
