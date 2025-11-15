<?php
/**
 * Reviews REST API Controller
 *
 * 处理商品评价相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 评价 REST API 控制器
 *
 * 提供评价的 CRUD 操作
 */
class Tanzanite_REST_Reviews_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'reviews';

	/**
	 * 评价表名
	 *
	 * @var string
	 */
	private $reviews_table;

	/**
	 * 允许的评价状态
	 *
	 * @var array
	 */
	private $allowed_statuses = array( 'pending', 'approved', 'rejected', 'hidden' );

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->reviews_table = $wpdb->prefix . 'tanz_product_reviews';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 列表和创建
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => $this->get_collection_params(),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_reviews', true ),
					'args'                => $this->get_create_params(),
				),
			)
		);

		// 获取、更新、删除单个评价
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => array(
						'id' => array(
							'validate_callback' => 'is_numeric',
						),
					),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_reviews', true ),
					'args'                => $this->get_update_params(),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_reviews', true ),
					'args'                => array(
						'id' => array(
							'validate_callback' => 'is_numeric',
						),
					),
				),
			)
		);
	}

	/**
	 * 获取评价列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		$pagination = $this->get_pagination_params( $request );

		$result = $this->fetch_reviews(
			array(
				'status'     => $request->get_param( 'status' ),
				'product_id' => $request->get_param( 'product_id' ),
				'search'     => $request->get_param( 'search' ),
				'limit'      => $pagination['per_page'],
				'offset'     => $pagination['offset'],
			)
		);

		return $this->respond_success(
			array(
				'items' => $result['items'],
				'meta'  => array_merge(
					$this->build_pagination_meta( $result['total'], $pagination['page'], $pagination['per_page'] ),
					array( 'statuses' => $this->allowed_statuses )
				),
			)
		);
	}

	/**
	 * 获取单个评价
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		$review = $this->fetch_review_row( (int) $request['id'] );

		if ( ! $review ) {
			return $this->respond_error( 'review_not_found', __( '指定的评价不存在。', 'tanzanite-settings' ), 404 );
		}

		return $this->respond_success( $review );
	}

	/**
	 * 创建评价
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		global $wpdb;

		$product_id = (int) $request->get_param( 'product_id' );
		$user_id    = $request->get_param( 'user_id' ) ? (int) $request->get_param( 'user_id' ) : get_current_user_id();
		$rating     = max( 1, min( 5, (int) $request->get_param( 'rating' ) ) );
		$content    = wp_kses_post( (string) $request->get_param( 'content' ) );
		$images     = $request->get_param( 'images' ) ?: array();
		$status     = $this->normalize_status( $request->get_param( 'status' ) );
		$featured   = (bool) $request->get_param( 'featured' );

		// 验证商品存在
		if ( ! get_post( $product_id ) ) {
			return $this->respond_error( 'invalid_product', __( '商品不存在。', 'tanzanite-settings' ), 400 );
		}

		$inserted = $wpdb->insert(
			$this->reviews_table,
			array(
				'product_id'  => $product_id,
				'user_id'     => $user_id,
				'rating'      => $rating,
				'content'     => $content,
				'images'      => wp_json_encode( $images ),
				'status'      => $status,
				'is_featured' => $featured ? 1 : 0,
				'created_at'  => current_time( 'mysql' ),
			),
			array( '%d', '%d', '%d', '%s', '%s', '%s', '%d', '%s' )
		);

		if ( false === $inserted ) {
			return $this->respond_error( 'failed_create_review', __( '创建评价失败。', 'tanzanite-settings' ), 500 );
		}

		$review_id = $wpdb->insert_id;
		$review    = $this->fetch_review_row( $review_id );

		$this->log_audit( 'create', 'review', $review_id, array( 'product_id' => $product_id ), $request );

		return $this->respond_success( $review, 201 );
	}

	/**
	 * 更新评价
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		global $wpdb;

		$review = $this->fetch_review_row( (int) $request['id'] );

		if ( ! $review ) {
			return $this->respond_error( 'review_not_found', __( '指定的评价不存在。', 'tanzanite-settings' ), 404 );
		}

		$data  = array();
		$types = array();

		// 状态更新
		if ( $request->has_param( 'status' ) ) {
			$data['status'] = $this->normalize_status( $request->get_param( 'status' ) );
			$types[]        = '%s';
		}

		// 精华标记
		if ( $request->has_param( 'is_featured' ) ) {
			$data['is_featured'] = $request->get_param( 'is_featured' ) ? 1 : 0;
			$types[]             = '%d';
		}

		// 回复处理
		if ( $request->has_param( 'reply_text' ) ) {
			$reply_text = wp_kses_post( (string) $request->get_param( 'reply_text' ) );
			$user       = wp_get_current_user();

			if ( '' === trim( $reply_text ) ) {
				$data['reply_text']   = null;
				$data['reply_author'] = null;
				$data['reply_at']     = null;
			} else {
				$data['reply_text']   = $reply_text;
				$data['reply_author'] = $user ? $user->display_name : 'system';
				$data['reply_at']     = current_time( 'mysql' );
			}

			$types[] = '%s';
			$types[] = '%s';
			$types[] = '%s';
		}

		if ( empty( $data ) ) {
			return $this->respond_error( 'invalid_review_payload', __( '没有可更新的字段。', 'tanzanite-settings' ) );
		}

		// 分离回复字段和其他字段
		$update_data  = array();
		$update_types = array();

		foreach ( $data as $key => $value ) {
			if ( in_array( $key, array( 'reply_text', 'reply_author', 'reply_at' ), true ) ) {
				continue;
			}

			$update_data[ $key ] = $value;
			$update_types[]      = array_shift( $types );
			unset( $data[ $key ] );
		}

		// 更新非回复字段
		if ( ! empty( $update_data ) ) {
			$wpdb->update( $this->reviews_table, $update_data, array( 'id' => (int) $request['id'] ), $update_types, array( '%d' ) );
		}

		// 更新回复字段
		if ( array_key_exists( 'reply_text', $data ) ) {
			$wpdb->update(
				$this->reviews_table,
				array(
					'reply_text'   => $data['reply_text'],
					'reply_author' => $data['reply_author'],
					'reply_at'     => $data['reply_at'],
				),
				array( 'id' => (int) $request['id'] ),
				array( '%s', '%s', '%s' ),
				array( '%d' )
			);
		}

		$fresh = $this->fetch_review_row( (int) $request['id'] );

		$this->log_audit(
			'update',
			'review',
			(int) $request['id'],
			array(
				'status'      => $fresh['status'],
				'is_featured' => $fresh['is_featured'],
			),
			$request
		);

		return $this->respond_success( $fresh );
	}

	/**
	 * 删除评价
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		global $wpdb;

		$review = $this->fetch_review_row( (int) $request['id'] );

		if ( ! $review ) {
			return $this->respond_error( 'review_not_found', __( '指定的评价不存在。', 'tanzanite-settings' ), 404 );
		}

		$deleted = $wpdb->delete( $this->reviews_table, array( 'id' => (int) $request['id'] ), array( '%d' ) );

		if ( false === $deleted ) {
			return $this->respond_error( 'failed_delete_review', __( '删除评价失败，请稍后重试。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit(
			'delete',
			'review',
			(int) $request['id'],
			array(
				'product_id' => $review['product_id'],
				'status'     => $review['status'],
			),
			$request
		);

		return $this->respond_success(
			array(
				'deleted' => true,
				'id'      => (int) $request['id'],
			)
		);
	}

	/**
	 * 获取评价列表（内部方法）
	 *
	 * @since 0.2.0
	 * @param array $args 查询参数
	 * @return array
	 */
	private function fetch_reviews( $args = array() ) {
		global $wpdb;

		$defaults = array(
			'status'     => null,
			'product_id' => null,
			'search'     => null,
			'limit'      => 50,
			'offset'     => 0,
		);

		$args   = wp_parse_args( $args, $defaults );
		$where  = array();
		$params = array();

		if ( $args['status'] ) {
			$where[]  = 'status = %s';
			$params[] = $this->normalize_status( $args['status'] );
		}

		if ( $args['product_id'] ) {
			$where[]  = 'product_id = %d';
			$params[] = (int) $args['product_id'];
		}

		if ( $args['search'] ) {
			$like     = '%' . $wpdb->esc_like( $args['search'] ) . '%';
			$where[]  = '(author_name LIKE %s OR author_email LIKE %s OR content LIKE %s)';
			$params[] = $like;
			$params[] = $like;
			$params[] = $like;
		}

		$where_sql = $where ? 'WHERE ' . implode( ' AND ', $where ) : '';

		$query    = "SELECT * FROM {$this->reviews_table} {$where_sql} ORDER BY created_at DESC LIMIT %d OFFSET %d";
		$params[] = (int) $args['limit'];
		$params[] = (int) $args['offset'];

		$rows = $wpdb->get_results( $wpdb->prepare( $query, $params ), ARRAY_A );

		$count_query = "SELECT COUNT(*) FROM {$this->reviews_table} {$where_sql}";
		$total       = (int) $wpdb->get_var( $wpdb->prepare( $count_query, array_slice( $params, 0, count( $params ) - 2 ) ) );

		return array(
			'items' => array_map( array( $this, 'format_review_row' ), $rows ),
			'total' => $total,
		);
	}

	/**
	 * 获取单个评价（内部方法）
	 *
	 * @since 0.2.0
	 * @param int $id 评价 ID
	 * @return array|null
	 */
	private function fetch_review_row( $id ) {
		global $wpdb;

		$row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->reviews_table} WHERE id = %d", $id ), ARRAY_A );

		return $row ? $this->format_review_row( $row ) : null;
	}

	/**
	 * 格式化评价数据
	 *
	 * @since 0.2.0
	 * @param array $row 数据库行
	 * @return array
	 */
	private function format_review_row( $row ) {
		return array(
			'id'           => (int) $row['id'],
			'product_id'   => (int) $row['product_id'],
			'user_id'      => (int) $row['user_id'],
			'author_name'  => $row['author_name'],
			'author_email' => $row['author_email'],
			'author_phone' => $row['author_phone'],
			'rating'       => (int) $row['rating'],
			'content'      => $row['content'],
			'images'       => $row['images'] ? json_decode( $row['images'], true ) : array(),
			'status'       => $row['status'],
			'is_featured'  => (bool) $row['is_featured'],
			'reply'        => array(
				'text'   => $row['reply_text'],
				'author' => $row['reply_author'],
				'time'   => $row['reply_at'],
			),
			'meta'         => $row['meta'] ? json_decode( $row['meta'], true ) : array(),
			'created_at'   => $row['created_at'],
			'updated_at'   => $row['updated_at'],
		);
	}

	/**
	 * 规范化评价状态
	 *
	 * @since 0.2.0
	 * @param string|null $status 状态
	 * @return string
	 */
	private function normalize_status( $status ) {
		$status = $status ? sanitize_key( $status ) : 'pending';

		return in_array( $status, $this->allowed_statuses, true ) ? $status : 'pending';
	}

	/**
	 * 获取集合参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_collection_params() {
		return array(
			'page'       => array(
				'type'    => 'integer',
				'default' => 1,
			),
			'per_page'   => array(
				'type'    => 'integer',
				'default' => 20,
			),
			'status'     => array(
				'type' => 'string',
			),
			'product_id' => array(
				'type' => 'integer',
			),
			'search'     => array(
				'type' => 'string',
			),
		);
	}

	/**
	 * 获取创建参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_create_params() {
		return array(
			'product_id' => array(
				'type'     => 'integer',
				'required' => true,
			),
			'user_id'    => array(
				'type' => 'integer',
			),
			'rating'     => array(
				'type'     => 'integer',
				'required' => true,
			),
			'content'    => array(
				'type' => 'string',
			),
			'images'     => array(
				'type' => 'array',
			),
			'status'     => array(
				'type'    => 'string',
				'default' => 'pending',
			),
			'featured'   => array(
				'type'    => 'boolean',
				'default' => false,
			),
		);
	}

	/**
	 * 获取更新参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_update_params() {
		return array(
			'id'         => array(
				'validate_callback' => 'is_numeric',
			),
			'status'     => array(
				'type' => 'string',
			),
			'is_featured' => array(
				'type' => 'boolean',
			),
			'reply_text' => array(
				'type' => 'string',
			),
		);
	}
}
