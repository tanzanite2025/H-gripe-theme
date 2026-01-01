<?php
/**
 * 保修申请 REST API 控制器
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 保修申请 REST API 控制器
 */
class Tanzanite_REST_Warranty_Claims_Controller {

	/**
	 * REST API 命名空间
	 *
	 * @var string
	 */
	protected $namespace = 'tanzanite/v1';

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'warranty/claim';

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
	}

	/**
	 * 注册路由
	 */
	public function register_routes() {
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => '__return_true', // 公开接口，或许需要 nonce 或 captcha
				),
			)
		);
	}

	/**
	 * 创建申请 (处理表单提交)
	 *
	 * @param WP_REST_Request $request 请求对象
	 * @return WP_REST_Response|WP_Error
	 */
	public function create_item( $request ) {
		// 1. 获取文本字段
		$order_number = sanitize_text_field( $request->get_param( 'order_number' ) );
		$email        = sanitize_email( $request->get_param( 'email' ) );
		$tire_pressure = sanitize_text_field( $request->get_param( 'tire_pressure' ) );
		$is_tubeless  = $request->get_param( 'is_tubeless' ) === 'yes' ? 1 : 0;
		$description  = sanitize_textarea_field( $request->get_param( 'issue_description' ) );

		if ( empty( $order_number ) || empty( $email ) ) {
			return new WP_Error( 'missing_params', 'Order Number and Email are required.', array( 'status' => 400 ) );
		}

		// 2. 处理文件上传
		require_once( ABSPATH . 'wp-admin/includes/image.php' );
		require_once( ABSPATH . 'wp-admin/includes/file.php' );
		require_once( ABSPATH . 'wp-admin/includes/media.php' );

		$uploaded_images = array();
		$video_url = '';

		// 处理图片 (images[] 是数组)
		$files = $_FILES['images'] ?? null;
		if ( $files ) {
			// 重组 $_FILES 数组以支持多文件循环处理
			// PHP 的 $_FILES['images']['name'] 是个数组
			$count = count( $files['name'] );
			for ( $i = 0; $i < $count; $i++ ) {
				if ( $files['error'][ $i ] !== 0 ) {
					continue;
				}

				// 构造单个文件的数组，模拟单个文件上传
				$file = array(
					'name'     => $files['name'][ $i ],
					'type'     => $files['type'][ $i ],
					'tmp_name' => $files['tmp_name'][ $i ],
					'error'    => $files['error'][ $i ],
					'size'     => $files['size'][ $i ],
				);
				
				// 临时覆盖 $_FILES['upload'] 供 media_handle_upload 使用 (需小心)
				// 更好的方式是使用更底层的处理，但为了利用 WP 的媒体库功能:
				$_FILES['tanz_temp_img'] = $file;
				
				$attachment_id = media_handle_upload( 'tanz_temp_img', 0 );
				
				if ( ! is_wp_error( $attachment_id ) ) {
					$uploaded_images[] = wp_get_attachment_url( $attachment_id );
				}
			}
		}

		// 处理视频
		if ( ! empty( $_FILES['video'] ) && $_FILES['video']['error'] === 0 ) {
			$attachment_id = media_handle_upload( 'video', 0 );
			if ( ! is_wp_error( $attachment_id ) ) {
				$video_url = wp_get_attachment_url( $attachment_id );
			}
		} 
		// 也许用户直接提供了视频 URL (预留)
		if ( empty( $video_url ) ) {
			$video_url = esc_url_raw( $request->get_param( 'video_url' ) );
		}

		// 3. 保存到数据库
		$data = array(
			'order_number'      => $order_number,
			'email'             => $email,
			'tire_pressure'     => $tire_pressure,
			'is_tubeless'       => $is_tubeless,
			'issue_description' => $description,
			'images'            => $uploaded_images,
			'video_url'         => $video_url,
			'status'            => 'pending',
		);

		$claim_id = $this->db->save_warranty_claim( $data );

		if ( ! $claim_id ) {
			return new WP_Error( 'db_error', 'Failed to save claim.', array( 'status' => 500 ) );
		}

		// 4. 发送邮件通知管理员 (可选)
		// wp_mail( get_option('admin_email'), 'New Warranty Claim', ... );

		return rest_ensure_response( array(
			'success' => true,
			'message' => 'Claim submitted successfully.',
			'id'      => $claim_id
		) );
	}
}
