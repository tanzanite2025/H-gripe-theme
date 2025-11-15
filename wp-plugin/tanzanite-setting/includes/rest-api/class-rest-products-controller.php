<?php
/**
 * Products REST API Controller
 *
 * 处理商品相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 商品 REST API 控制器
 *
 * 提供商品的 CRUD 操作和批量操作
 */
class Tanzanite_REST_Products_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'products';

	/**
	 * 商品自定义文章类型
	 *
	 * @var string
	 */
	private $post_type = 'tanz_product';

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 列表、创建、批量操作
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => '__return_true', // 公开访问，所有人都能查看商品列表
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'bulk_action' ),
					'permission_callback' => $this->permission_callback( 'tanz_bulk_products', true ),
				),
			)
		);

		// 获取、更新、删除单个商品
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_item' ),
					'permission_callback' => '__return_true', // 公开访问，所有人都能查看单个商品
				),
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
				array(
					'methods'             => WP_REST_Server::DELETABLE,
					'callback'            => array( $this, 'delete_item' ),
					'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
				),
			)
		);
	}

	/**
	 * 获取商品列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		// 解析查询参数
		$per_page = (int) max( 1, min( 200, $request->get_param( 'per_page' ) ?: 20 ) );
		$page     = (int) max( 1, $request->get_param( 'page' ) ?: 1 );
		$keyword  = trim( (string) ( $request->get_param( 'keyword' ) ?: '' ) );
		$status   = sanitize_key( (string) ( $request->get_param( 'status' ) ?: '' ) );
		$category = (int) ( $request->get_param( 'category' ) ?: 0 );
		$author_id = (int) ( $request->get_param( 'author' ) ?: 0 );
		$sort     = sanitize_key( (string) ( $request->get_param( 'sort' ) ?: 'updated_at' ) );
		$order    = strtoupper( (string) ( $request->get_param( 'order' ) ?: 'DESC' ) );

		// 验证状态
		$allowed_statuses = array( 'publish', 'draft', 'pending', 'private' );
		if ( ! in_array( $status, $allowed_statuses, true ) ) {
			$status = '';
		}

		// 验证排序
		if ( ! in_array( $sort, array( 'updated_at', 'price_regular', 'stock_qty', 'points_reward' ), true ) ) {
			$sort = 'updated_at';
		}

		// 验证顺序
		if ( ! in_array( $order, array( 'ASC', 'DESC' ), true ) ) {
			$order = 'DESC';
		}

		// 构建查询参数
		$args = array(
			'post_type'      => $this->post_type,
			'post_status'    => $status ? $status : 'any',
			'posts_per_page' => $per_page,
			'paged'          => $page,
			'order'          => $order,
		);

		// 关键词搜索
		if ( $keyword ) {
			$args['s'] = $keyword;
		}

		// 作者筛选
		if ( $author_id > 0 ) {
			$args['author'] = $author_id;
		}

		// 分类筛选
		$tax_query = array();
		if ( $category > 0 ) {
			$tax_query[] = array(
				'taxonomy' => 'tanz_product_category',
				'field'    => 'term_id',
				'terms'    => $category,
			);
		}

		// 标签筛选
		$tags_raw = $request->get_param( 'tags' );
		$tag_terms = array();
		if ( is_array( $tags_raw ) ) {
			foreach ( $tags_raw as $tag_item ) {
				$sanitized = sanitize_title( (string) $tag_item );
				if ( $sanitized ) {
					$tag_terms[] = $sanitized;
				}
			}
		}

		if ( $tag_terms ) {
			$tax_query[] = array(
				'taxonomy' => 'tanz_product_tag',
				'field'    => 'slug',
				'terms'    => $tag_terms,
			);
		}

		if ( ! empty( $tax_query ) ) {
			$args['tax_query'] = $tax_query;
		}

		// Meta 查询（库存、积分等）
		$meta_query = array();

		$min_inventory = $request->has_param( 'inventory_min' ) ? (int) $request->get_param( 'inventory_min' ) : null;
		$max_inventory = $request->has_param( 'inventory_max' ) ? (int) $request->get_param( 'inventory_max' ) : null;

		if ( null !== $min_inventory || null !== $max_inventory ) {
			if ( null !== $min_inventory && null !== $max_inventory ) {
				$meta_query[] = array(
					'key'     => '_tanz_stock_qty',
					'value'   => array( $min_inventory, $max_inventory ),
					'compare' => 'BETWEEN',
					'type'    => 'NUMERIC',
				);
			} elseif ( null !== $min_inventory ) {
				$meta_query[] = array(
					'key'     => '_tanz_stock_qty',
					'value'   => $min_inventory,
					'compare' => '>=',
					'type'    => 'NUMERIC',
				);
			} elseif ( null !== $max_inventory ) {
				$meta_query[] = array(
					'key'     => '_tanz_stock_qty',
					'value'   => $max_inventory,
					'compare' => '<=',
					'type'    => 'NUMERIC',
				);
			}
		}

		if ( ! empty( $meta_query ) ) {
			$args['meta_query'] = $meta_query;
		}

		// 排序
		if ( 'price_regular' === $sort ) {
			$args['meta_key'] = '_tanz_price_regular';
			$args['orderby']  = 'meta_value_num';
		} elseif ( 'stock_qty' === $sort ) {
			$args['meta_key'] = '_tanz_stock_qty';
			$args['orderby']  = 'meta_value_num';
		} elseif ( 'points_reward' === $sort ) {
			$args['meta_key'] = '_tanz_points_reward';
			$args['orderby']  = 'meta_value_num';
		} else {
			$args['orderby'] = 'modified';
		}

		// 执行查询
		$query = new WP_Query( $args );

		$items = array();
		foreach ( $query->posts as $post ) {
			$items[] = $this->prepare_item_for_response( $post );
		}

		return $this->respond_success(
			array(
				'items' => $items,
				'meta'  => array(
					'page'        => $page,
					'per_page'    => $per_page,
					'total_pages' => $query->max_num_pages,
					'total'       => $query->found_posts,
					'filters'     => array( 'keyword', 'status', 'category', 'tags', 'author', 'inventory_min', 'inventory_max' ),
					'sorting'     => array( 'updated_at', 'price_regular', 'stock_qty', 'points_reward' ),
				),
			)
		);
	}

	/**
	 * 创建商品
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function create_item( $request ) {
		// 验证状态
		$status = sanitize_key( (string) $request->get_param( 'status' ) );
		$allowed_statuses = array( 'publish', 'draft', 'pending', 'private' );
		if ( ! in_array( $status, $allowed_statuses, true ) ) {
			$status = 'draft';
		}

		// 准备文章数据
		$postarr = array(
			'post_type'    => $this->post_type,
			'post_status'  => $status,
			'post_title'   => sanitize_text_field( (string) $request->get_param( 'title' ) ),
			'post_excerpt' => sanitize_textarea_field( (string) $request->get_param( 'excerpt' ) ),
			'post_content' => wp_kses_post( (string) $request->get_param( 'content' ) ),
			'post_author'  => get_current_user_id(),
		);

		// Slug
		$slug = $request->get_param( 'slug' );
		if ( $slug ) {
			$postarr['post_name'] = sanitize_title( (string) $slug );
		}

		// 创建文章
		$post_id = wp_insert_post( $postarr, true );

		if ( is_wp_error( $post_id ) ) {
			return $this->respond_error( 'create_failed', $post_id->get_error_message(), 500 );
		}

		// 更新 Meta 数据
		$this->update_product_meta( $post_id, $request );

		// 处理分类
		if ( $request->has_param( 'category_ids' ) ) {
			$category_ids = $request->get_param( 'category_ids' );
			if ( is_array( $category_ids ) ) {
				wp_set_post_terms( $post_id, array_map( 'intval', $category_ids ), 'tanz_product_category' );
			}
		}

		// 处理标签
		if ( $request->has_param( 'tag_ids' ) ) {
			$tag_ids = $request->get_param( 'tag_ids' );
			if ( is_array( $tag_ids ) ) {
				wp_set_post_terms( $post_id, array_map( 'intval', $tag_ids ), 'tanz_product_tag' );
			}
		}

		// 处理 SKU
		$sku_count = 0;
		$sku_result = $this->handle_product_skus( $post_id, $request );
		if ( is_wp_error( $sku_result ) ) {
			wp_delete_post( $post_id, true );
			return $this->respond_error( $sku_result->get_error_code(), $sku_result->get_error_message(), 400 );
		}
		$sku_count = $sku_result['count'] ?? 0;

		// 处理 SEO 数据
		if ( $request->has_param( 'seo' ) ) {
			$seo_param = $request->get_param( 'seo' );
			if ( is_array( $seo_param ) ) {
				$seo_payload = $seo_param['payload'] ?? null;
				if ( null !== $seo_payload ) {
					$this->sync_product_seo( $post_id, $seo_payload );
				}
			}
		}

		// 审计日志
		$this->log_audit(
			'create',
			'product',
			$post_id,
			array(
				'title'      => get_the_title( $post_id ),
				'status'     => $status,
				'skus_count' => $sku_count,
			),
			$request
		);

		return $this->respond_success( $this->prepare_item_for_response( get_post( $post_id ) ), 201 );
	}

	/**
	 * 更新商品 Meta 数据
	 *
	 * @since 0.2.0
	 * @param int             $post_id 文章ID
	 * @param WP_REST_Request $request REST 请求对象
	 */
	private function update_product_meta( $post_id, $request ) {
		// Meta 字段映射
		$meta_fields = array(
			'price_regular'        => '_tanz_price_regular',
			'price_sale'           => '_tanz_price_sale',
			'price_member'         => '_tanz_price_member',
			'stock_qty'            => '_tanz_stock_qty',
			'stock_alert'          => '_tanz_stock_alert',
			'points_reward'        => '_tanz_points_reward',
			'points_limit'         => '_tanz_points_limit',
			'purchase_limit'       => '_tanz_purchase_limit',
			'min_purchase'         => '_tanz_min_purchase',
			'backorders_allowed'   => '_tanz_backorders_allowed',
			'featured_image_id'    => '_tanz_featured_image_id',
			'featured_image_url'   => '_tanz_featured_image_url',
			'featured_video_id'    => '_tanz_featured_video_id',
			'featured_video_url'   => '_tanz_featured_video_url',
			'shipping_template_id' => '_tanz_shipping_template_id',
			'free_shipping'        => '_tanz_free_shipping',
			'shipping_time'        => '_tanz_shipping_time',
			'is_sticky'            => '_tanz_is_sticky',
		);

		foreach ( $meta_fields as $param => $meta_key ) {
			if ( ! $request->has_param( $param ) ) {
				continue;
			}

			$value = $request->get_param( $param );

			// 类型转换
			if ( in_array( $param, array( 'price_regular', 'price_sale', 'price_member' ), true ) ) {
				$value = (float) $value;
			} elseif ( in_array( $param, array( 'stock_qty', 'stock_alert', 'points_reward', 'points_limit', 'purchase_limit', 'min_purchase', 'featured_image_id', 'featured_video_id', 'shipping_template_id' ), true ) ) {
				$value = (int) $value;
			} elseif ( in_array( $param, array( 'backorders_allowed', 'free_shipping', 'is_sticky' ), true ) ) {
				$value = (bool) $value;
			} else {
				$value = sanitize_text_field( (string) $value );
			}

			update_post_meta( $post_id, $meta_key, $value );
		}

		// 数组类型的 Meta
		if ( $request->has_param( 'tier_pricing' ) ) {
			$tier_pricing = $request->get_param( 'tier_pricing' );
			if ( is_array( $tier_pricing ) ) {
				update_post_meta( $post_id, '_tanz_tier_pricing', $tier_pricing );
			}
		}

		if ( $request->has_param( 'gallery_media_ids' ) ) {
			$gallery_ids = $request->get_param( 'gallery_media_ids' );
			if ( is_array( $gallery_ids ) ) {
				update_post_meta( $post_id, '_tanz_gallery_media_ids', array_map( 'intval', $gallery_ids ) );
			}
		}

		if ( $request->has_param( 'gallery_external_urls' ) ) {
			$gallery_urls = $request->get_param( 'gallery_external_urls' );
			if ( is_array( $gallery_urls ) ) {
				update_post_meta( $post_id, '_tanz_gallery_external_urls', array_map( 'esc_url_raw', $gallery_urls ) );
			}
		}

		if ( $request->has_param( 'membership_levels' ) ) {
			$levels = $request->get_param( 'membership_levels' );
			if ( is_array( $levels ) ) {
				update_post_meta( $post_id, '_tanz_membership_levels', $levels );
			}
		}

		if ( $request->has_param( 'logistics_tags' ) ) {
			$tags = $request->get_param( 'logistics_tags' );
			if ( is_array( $tags ) ) {
				update_post_meta( $post_id, '_tanz_logistics_tags', $tags );
			}
		}

		if ( $request->has_param( 'channels' ) ) {
			$channels = $request->get_param( 'channels' );
			if ( is_array( $channels ) ) {
				update_post_meta( $post_id, '_tanz_channels', $channels );
			}
		}

		// URLLink 自定义路径
		if ( $request->has_param( 'urllink_custom_path' ) ) {
			$custom_path = trim( $request->get_param( 'urllink_custom_path' ) );
			if ( ! empty( $custom_path ) ) {
				// 规范化路径：移除开头和结尾的斜杠
				$custom_path = trim( $custom_path, '/' );
				update_post_meta( $post_id, '_urllink_custom_path', sanitize_text_field( $custom_path ) );
			} else {
				// 如果为空，删除 meta
				delete_post_meta( $post_id, '_urllink_custom_path' );
			}
		}
	}

	/**
	 * 获取单个商品
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_item( $request ) {
		$post = get_post( (int) $request['id'] );

		if ( ! $post || $this->post_type !== $post->post_type ) {
			return $this->respond_error( 'product_not_found', __( '商品不存在。', 'tanzanite-settings' ), 404 );
		}

		return $this->respond_success( $this->prepare_item_for_response( $post ) );
	}

	/**
	 * 更新商品
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function update_item( $request ) {
		$post_id = (int) $request['id'];
		$post    = get_post( $post_id );

		if ( ! $post || $this->post_type !== $post->post_type ) {
			return $this->respond_error( 'product_not_found', __( '商品不存在。', 'tanzanite-settings' ), 404 );
		}

		$data       = array( 'ID' => $post_id );
		$has_update = false;

		// 更新标题
		if ( $request->has_param( 'title' ) ) {
			$data['post_title'] = sanitize_text_field( (string) $request->get_param( 'title' ) );
			$has_update         = true;
		}

		// 更新摘要
		if ( $request->has_param( 'excerpt' ) ) {
			$data['post_excerpt'] = sanitize_textarea_field( (string) $request->get_param( 'excerpt' ) );
			$has_update           = true;
		}

		// 更新状态
		if ( $request->has_param( 'status' ) ) {
			$status = sanitize_key( (string) $request->get_param( 'status' ) );
			$allowed_statuses = array( 'publish', 'draft', 'pending', 'private' );
			if ( in_array( $status, $allowed_statuses, true ) ) {
				$data['post_status'] = $status;
				$has_update          = true;
			}
		}

		// 更新 Slug
		if ( $request->has_param( 'slug' ) ) {
			$data['post_name'] = sanitize_title( (string) $request->get_param( 'slug' ) );
			$has_update        = true;
		}

		// 更新内容
		if ( $request->has_param( 'content' ) ) {
			$data['post_content'] = wp_kses_post( (string) $request->get_param( 'content' ) );
			$has_update           = true;
		}

		// 执行更新
		if ( $has_update ) {
			$result = wp_update_post( $data, true );
			if ( is_wp_error( $result ) ) {
				return $this->respond_error( 'update_failed', $result->get_error_message(), 500 );
			}
		}

		// 更新 Meta 数据
		$this->update_product_meta( $post_id, $request );

		// 更新分类
		if ( $request->has_param( 'category_ids' ) ) {
			$category_ids = $request->get_param( 'category_ids' );
			if ( is_array( $category_ids ) ) {
				wp_set_post_terms( $post_id, array_map( 'intval', $category_ids ), 'tanz_product_category' );
			}
		}

		// 更新标签
		if ( $request->has_param( 'tag_ids' ) ) {
			$tag_ids = $request->get_param( 'tag_ids' );
			if ( is_array( $tag_ids ) ) {
				wp_set_post_terms( $post_id, array_map( 'intval', $tag_ids ), 'tanz_product_tag' );
			}
		}

		// 处理 SKU
		$sku_result = $this->handle_product_skus( $post_id, $request );
		if ( is_wp_error( $sku_result ) ) {
			return $this->respond_error( $sku_result->get_error_code(), $sku_result->get_error_message(), 400 );
		}

		// 处理 SEO 数据
		if ( $request->has_param( 'seo' ) ) {
			$seo_param = $request->get_param( 'seo' );
			if ( is_array( $seo_param ) ) {
				$seo_payload = $seo_param['payload'] ?? null;
				if ( null !== $seo_payload ) {
					$this->sync_product_seo( $post_id, $seo_payload );
				}
			}
		}

		// 审计日志
		$this->log_audit( 'update', 'product', $post_id, array( 'title' => get_the_title( $post_id ) ), $request );

		return $this->respond_success( $this->prepare_item_for_response( get_post( $post_id ) ) );
	}

	/**
	 * 删除商品
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function delete_item( $request ) {
		$post_id = (int) $request['id'];
		$post    = get_post( $post_id );

		if ( ! $post || $this->post_type !== $post->post_type ) {
			return $this->respond_error( 'product_not_found', __( '商品不存在。', 'tanzanite-settings' ), 404 );
		}

		$title = get_the_title( $post_id );
		wp_delete_post( $post_id, true );

		$this->log_audit( 'delete', 'product', $post_id, array( 'title' => $title ), $request );

		return $this->respond_success( array( 'deleted' => true ) );
	}

	/**
	 * 批量操作
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function bulk_action( $request ) {
		$action = sanitize_key( $request->get_param( 'action' ) );

		// 支持的批量操作
		$allowed_actions = array( 'set_status', 'adjust_stock', 'set_meta', 'adjust_price', 'delete', 'export' );

		if ( ! in_array( $action, $allowed_actions, true ) ) {
			return $this->respond_error( 'invalid_bulk_action', __( '当前批量操作类型不受支持。', 'tanzanite-settings' ) );
		}

		// 验证 ID 列表
		$ids = $request->get_param( 'ids' );
		if ( ! is_array( $ids ) || empty( $ids ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '请选择至少一个需要处理的商品。', 'tanzanite-settings' ) );
		}

		$ids = array_map( 'absint', $ids );
		$ids = array_filter( $ids );

		if ( empty( $ids ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '无效的商品ID列表。', 'tanzanite-settings' ) );
		}

		// 获取 payload
		$payload = $request->get_param( 'payload' );
		if ( ! is_array( $payload ) ) {
			$payload = array();
		}

		// 初始化结果
		$summary = array(
			'action'    => $action,
			'total'     => count( $ids ),
			'updated'   => 0,
			'failed'    => array(),
			'details'   => array(),
			'timestamp' => current_time( 'mysql' ),
		);

		// 根据操作类型处理
		switch ( $action ) {
			case 'set_status':
				$summary = $this->bulk_set_status( $ids, $payload, $summary, $request );
				break;

			case 'adjust_stock':
				$summary = $this->bulk_adjust_stock( $ids, $payload, $summary, $request );
				break;

			case 'set_meta':
				$summary = $this->bulk_set_meta( $ids, $payload, $summary, $request );
				break;

			case 'adjust_price':
				$summary = $this->bulk_adjust_price( $ids, $payload, $summary, $request );
				break;

			case 'delete':
				$summary = $this->bulk_delete( $ids, $payload, $summary, $request );
				break;

			case 'export':
				return $this->bulk_export( $ids, $request );
		}

		return $this->respond_success( $summary );
	}

	/**
	 * 准备商品数据用于响应
	 *
	 * @since 0.2.0
	 * @param WP_Post $post 文章对象
	 * @return array
	 */
	private function prepare_item_for_response( $post ) {
		// 获取 Meta 数据
		$price_regular = (float) get_post_meta( $post->ID, '_tanz_price_regular', true );
		$price_sale    = (float) get_post_meta( $post->ID, '_tanz_price_sale', true );
		$price_member  = (float) get_post_meta( $post->ID, '_tanz_price_member', true );
		$stock_qty     = (int) get_post_meta( $post->ID, '_tanz_stock_qty', true );
		$stock_alert   = (int) get_post_meta( $post->ID, '_tanz_stock_alert', true );
		$points_reward = (int) get_post_meta( $post->ID, '_tanz_points_reward', true );
		$points_limit  = (int) get_post_meta( $post->ID, '_tanz_points_limit', true );
		$featured_image_id = (int) get_post_meta( $post->ID, '_tanz_featured_image_id', true );
		$is_sticky     = (bool) get_post_meta( $post->ID, '_tanz_is_sticky', true );

		// 获取缩略图
		$thumbnail = '';
		if ( $featured_image_id > 0 ) {
			$thumbnail = wp_get_attachment_image_url( $featured_image_id, 'thumbnail' );
		}

		// 获取分类
		$terms = get_the_terms( $post->ID, 'tanz_product_category' );
		$categories = array();
		if ( is_array( $terms ) ) {
			foreach ( $terms as $term ) {
				$categories[] = array(
					'id'   => $term->term_id,
					'name' => $term->name,
					'slug' => $term->slug,
				);
			}
		}

		return array(
			'id'         => $post->ID,
			'title'      => get_the_title( $post ),
			'status'     => $post->post_status,
			'excerpt'    => $post->post_excerpt,
			'slug'       => $post->post_name,
			'thumbnail'  => $thumbnail,
			'prices'     => array(
				'regular' => $price_regular,
				'sale'    => $price_sale,
				'member'  => $price_member,
			),
			'stock'      => array(
				'quantity' => $stock_qty,
				'alert'    => $stock_alert,
			),
			'points'     => array(
				'reward' => $points_reward,
				'limit'  => $points_limit,
			),
			'categories' => $categories,
			'sticky'     => $is_sticky,
			'updated_at' => $post->post_modified_gmt,
			'created_at' => $post->post_date_gmt,
			'preview_url' => get_permalink( $post ),
		);
	}

	/**
	 * 处理商品 SKU
	 *
	 * @since 0.2.0
	 * @param int             $product_id 商品ID
	 * @param WP_REST_Request $request REST 请求对象
	 * @return array|WP_Error
	 */
	private function handle_product_skus( $product_id, $request ) {
		$has_skus_param = $request->has_param( 'skus' );
		$bulk_payload   = $request->get_param( 'skus_bulk' );

		// 如果没有提供 SKU 参数，返回当前数量
		if ( ! $has_skus_param && null === $bulk_payload ) {
			return array( 'count' => $this->count_product_skus( $product_id ) );
		}

		$skus = $has_skus_param ? (array) $request->get_param( 'skus' ) : array();

		// 处理批量 SKU
		if ( empty( $skus ) && null !== $bulk_payload && '' !== trim( (string) $bulk_payload ) ) {
			$parsed = $this->parse_bulk_skus( (string) $bulk_payload );
			if ( is_wp_error( $parsed ) ) {
				return $parsed;
			}
			$skus = $parsed;
		}

		// 获取默认值
		$defaults = array(
			'price_regular' => (float) get_post_meta( $product_id, '_tanz_price_regular', true ),
			'price_sale'    => (float) get_post_meta( $product_id, '_tanz_price_sale', true ),
			'stock_qty'     => (int) get_post_meta( $product_id, '_tanz_stock_qty', true ),
		);

		// 清理和验证 SKU
		$sanitized = $this->sanitize_skus( $skus, $defaults );
		if ( is_wp_error( $sanitized ) ) {
			return $sanitized;
		}

		// 如果 SKU 为空，删除所有 SKU
		if ( empty( $sanitized ) ) {
			$this->delete_product_skus( $product_id );
			return array( 'count' => 0 );
		}

		// 保存 SKU
		if ( ! $this->persist_product_skus( $product_id, $sanitized ) ) {
			return new WP_Error( 'failed_persist_skus', __( '保存 SKU 时发生异常，请稍后重试。', 'tanzanite-settings' ) );
		}

		return array( 'count' => count( $sanitized ) );
	}

	/**
	 * 解析批量 SKU
	 *
	 * @since 0.2.0
	 * @param string $bulk 批量 SKU 字符串
	 * @return array|WP_Error
	 */
	private function parse_bulk_skus( $bulk ) {
		$bulk = trim( $bulk );
		if ( '' === $bulk ) {
			return array();
		}

		// 尝试 JSON 解析
		$decoded = json_decode( $bulk, true );
		if ( is_array( $decoded ) ) {
			return $decoded;
		}

		// CSV 解析
		$lines  = preg_split( '/\r?\n/', $bulk );
		$parsed = array();

		foreach ( $lines as $line ) {
			$line = trim( $line );
			if ( '' === $line ) {
				continue;
			}

			$columns = str_getcsv( $line );
			if ( empty( $columns ) ) {
				continue;
			}

			$sku_code = trim( (string) ( $columns[0] ?? '' ) );
			if ( '' === $sku_code ) {
				return new WP_Error( 'invalid_sku_code', __( '批量 SKU 中存在空的 sku_code。', 'tanzanite-settings' ) );
			}

			$entry = array(
				'sku_code'      => $sku_code,
				'price_regular' => $columns[1] ?? null,
				'price_sale'    => $columns[2] ?? null,
				'stock_qty'     => $columns[3] ?? null,
			);

			// 属性（JSON 格式）
			if ( isset( $columns[4] ) && '' !== trim( (string) $columns[4] ) ) {
				$attr = json_decode( (string) $columns[4], true );
				if ( is_array( $attr ) ) {
					$entry['attributes'] = $attr;
				}
			}

			$parsed[] = $entry;
		}

		if ( empty( $parsed ) ) {
			return new WP_Error( 'invalid_sku_payload', __( '无法解析批量 SKU，请确认格式。', 'tanzanite-settings' ) );
		}

		return $parsed;
	}

	/**
	 * 清理和验证 SKU 数据
	 *
	 * @since 0.2.0
	 * @param array $skus SKU 数组
	 * @param array $defaults 默认值
	 * @return array|WP_Error
	 */
	private function sanitize_skus( $skus, $defaults = array() ) {
		if ( empty( $skus ) ) {
			return array();
		}

		if ( ! is_array( $skus ) ) {
			return new WP_Error( 'invalid_sku_payload', __( 'SKUs must be an array.', 'tanzanite-settings' ) );
		}

		$defaults = wp_parse_args(
			$defaults,
			array(
				'price_regular' => 0.0,
				'price_sale'    => 0.0,
				'stock_qty'     => 0,
			)
		);

		$sanitized  = array();
		$seen_codes = array();

		foreach ( $skus as $index => $sku ) {
			if ( ! is_array( $sku ) ) {
				return new WP_Error( 'invalid_sku_entry', sprintf( __( 'SKU %d must be an object.', 'tanzanite-settings' ), $index + 1 ) );
			}

			$sku_code = isset( $sku['sku_code'] ) ? sanitize_text_field( $sku['sku_code'] ) : '';
			if ( '' === $sku_code ) {
				return new WP_Error( 'invalid_sku_code', sprintf( __( 'SKU %d requires a sku_code value.', 'tanzanite-settings' ), $index + 1 ) );
			}

			// 检查重复
			if ( isset( $seen_codes[ $sku_code ] ) ) {
				return new WP_Error( 'duplicate_sku_code', sprintf( __( 'Duplicate SKU code detected: %s.', 'tanzanite-settings' ), $sku_code ) );
			}
			$seen_codes[ $sku_code ] = true;

			// 价格
			$price_regular = isset( $sku['price_regular'] ) ? (float) $sku['price_regular'] : 0.0;
			if ( $price_regular <= 0 && $defaults['price_regular'] > 0 ) {
				$price_regular = (float) $defaults['price_regular'];
			}

			$price_sale = isset( $sku['price_sale'] ) ? (float) $sku['price_sale'] : 0.0;
			if ( $price_sale <= 0 ) {
				$price_sale = $defaults['price_sale'] > 0 ? (float) $defaults['price_sale'] : $price_regular;
			}

			// 库存
			$stock_qty = isset( $sku['stock_qty'] ) ? max( 0, (int) $sku['stock_qty'] ) : (int) $defaults['stock_qty'];

			// 其他字段
			$weight  = isset( $sku['weight'] ) && '' !== $sku['weight'] ? (float) $sku['weight'] : null;
			$barcode = isset( $sku['barcode'] ) ? sanitize_text_field( $sku['barcode'] ) : '';

			// 属性
			$attributes = array();
			if ( isset( $sku['attributes'] ) && is_array( $sku['attributes'] ) ) {
				foreach ( $sku['attributes'] as $attr_key => $attr_value ) {
					$clean_key = sanitize_key( $attr_key );
					if ( is_array( $attr_value ) ) {
						$attributes[ $clean_key ] = array_map( 'sanitize_text_field', $attr_value );
					} else {
						$attributes[ $clean_key ] = sanitize_text_field( (string) $attr_value );
					}
				}
			}

			$sanitized[] = array(
				'sku_code'      => $sku_code,
				'attributes'    => $attributes,
				'price_regular' => $price_regular,
				'price_sale'    => $price_sale,
				'stock_qty'     => $stock_qty,
				'weight'        => $weight,
				'barcode'       => $barcode,
				'sort_order'    => isset( $sku['sort_order'] ) ? (int) $sku['sort_order'] : ( ( $index + 1 ) * 10 ),
			);
		}

		// 按 sort_order 排序
		usort(
			$sanitized,
			function ( $a, $b ) {
				return $a['sort_order'] <=> $b['sort_order'];
			}
		);

		return $sanitized;
	}

	/**
	 * 保存商品 SKU 到数据库
	 *
	 * @since 0.2.0
	 * @param int   $product_id 商品ID
	 * @param array $skus SKU 数组
	 * @return bool
	 */
	private function persist_product_skus( $product_id, $skus ) {
		global $wpdb;
		$table = $wpdb->prefix . 'tanz_product_skus';

		// 删除旧的 SKU
		$wpdb->delete( $table, array( 'product_id' => $product_id ), array( '%d' ) );

		// 插入新的 SKU
		foreach ( $skus as $sku ) {
			$result = $wpdb->insert(
				$table,
				array(
					'product_id'    => $product_id,
					'sku_code'      => $sku['sku_code'],
					'attributes'    => wp_json_encode( $sku['attributes'] ),
					'price_regular' => $sku['price_regular'],
					'price_sale'    => $sku['price_sale'],
					'stock_qty'     => $sku['stock_qty'],
					'sort_order'    => $sku['sort_order'],
					'weight'        => null === $sku['weight'] ? null : (string) $sku['weight'],
					'barcode'       => $sku['barcode'],
				),
				array( '%d', '%s', '%s', '%f', '%f', '%d', '%d', '%s', '%s' )
			);

			if ( false === $result ) {
				return false;
			}
		}

		return true;
	}

	/**
	 * 删除商品的所有 SKU
	 *
	 * @since 0.2.0
	 * @param int $product_id 商品ID
	 */
	private function delete_product_skus( $product_id ) {
		global $wpdb;
		$table = $wpdb->prefix . 'tanz_product_skus';
		$wpdb->delete( $table, array( 'product_id' => $product_id ), array( '%d' ) );
	}

	/**
	 * 统计商品 SKU 数量
	 *
	 * @since 0.2.0
	 * @param int $product_id 商品ID
	 * @return int
	 */
	private function count_product_skus( $product_id ) {
		global $wpdb;
		$table = $wpdb->prefix . 'tanz_product_skus';
		return (int) $wpdb->get_var( $wpdb->prepare( "SELECT COUNT(*) FROM {$table} WHERE product_id = %d", $product_id ) );
	}

	/**
	 * 批量设置状态
	 *
	 * @since 0.2.0
	 */
	private function bulk_set_status( $ids, $payload, $summary, $request ) {
		if ( empty( $payload['status'] ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '批量修改状态需要指定目标状态。', 'tanzanite-settings' ) );
		}

		$target_status = sanitize_key( (string) $payload['status'] );
		$allowed_statuses = array( 'publish', 'draft', 'pending', 'private' );

		if ( ! in_array( $target_status, $allowed_statuses, true ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '目标状态无效。', 'tanzanite-settings' ) );
		}

		foreach ( $ids as $product_id ) {
			$post = get_post( $product_id );

			if ( ! $post || $this->post_type !== $post->post_type ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => __( '商品不存在或类型不匹配。', 'tanzanite-settings' ),
				);
				continue;
			}

			if ( $post->post_status === $target_status ) {
				$summary['updated']++;
				$summary['details'][] = array(
					'id'      => $product_id,
					'status'  => $target_status,
					'changed' => false,
				);
				continue;
			}

			$result = wp_update_post(
				array(
					'ID'          => $product_id,
					'post_status' => $target_status,
				),
				true
			);

			if ( is_wp_error( $result ) ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => $result->get_error_message(),
				);
				continue;
			}

			$summary['updated']++;
			$summary['details'][] = array(
				'id'      => $product_id,
				'status'  => $target_status,
				'changed' => true,
			);

			$this->log_audit(
				'bulk_set_status',
				'product',
				$product_id,
				array(
					'action' => 'set_status',
					'status' => $target_status,
				),
				$request
			);
		}

		return $summary;
	}

	/**
	 * 批量调整库存
	 *
	 * @since 0.2.0
	 */
	private function bulk_adjust_stock( $ids, $payload, $summary, $request ) {
		if ( ! isset( $payload['delta'] ) || ! is_numeric( $payload['delta'] ) || 0 === (int) $payload['delta'] ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '批量调整库存需要提供非零的增量。', 'tanzanite-settings' ) );
		}

		$stock_delta = (int) $payload['delta'];

		foreach ( $ids as $product_id ) {
			$post = get_post( $product_id );

			if ( ! $post || $this->post_type !== $post->post_type ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => __( '商品不存在或类型不匹配。', 'tanzanite-settings' ),
				);
				continue;
			}

			// 调整库存
			$current = (int) get_post_meta( $product_id, '_tanz_stock_qty', true );
			$new_stock = max( 0, $current + $stock_delta );
			update_post_meta( $product_id, '_tanz_stock_qty', $new_stock );

			$summary['updated']++;
			$summary['details'][] = array(
				'id'          => $product_id,
				'stock_delta' => $stock_delta,
				'stock_qty'   => $new_stock,
			);

			$this->log_audit(
				'bulk_adjust_stock',
				'product',
				$product_id,
				array(
					'action'      => 'adjust_stock',
					'stock_delta' => $stock_delta,
					'stock_qty'   => $new_stock,
				),
				$request
			);
		}

		return $summary;
	}

	/**
	 * 批量设置 Meta 字段
	 *
	 * @since 0.2.0
	 */
	private function bulk_set_meta( $ids, $payload, $summary, $request ) {
		// Meta 字段映射
		$meta_fields_map = array(
			'price_regular'        => '_tanz_price_regular',
			'price_sale'           => '_tanz_price_sale',
			'price_member'         => '_tanz_price_member',
			'stock_qty'            => '_tanz_stock_qty',
			'stock_alert'          => '_tanz_stock_alert',
			'points_reward'        => '_tanz_points_reward',
			'points_limit'         => '_tanz_points_limit',
			'purchase_limit'       => '_tanz_purchase_limit',
			'min_purchase'         => '_tanz_min_purchase',
			'backorders_allowed'   => '_tanz_backorders_allowed',
			'shipping_template_id' => '_tanz_shipping_template_id',
			'free_shipping'        => '_tanz_free_shipping',
			'is_sticky'            => '_tanz_is_sticky',
		);

		$meta_fields = array_intersect_key( $payload, $meta_fields_map );

		if ( empty( $meta_fields ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '请至少提供一个可更新的商品字段。', 'tanzanite-settings' ) );
		}

		foreach ( $ids as $product_id ) {
			$post = get_post( $product_id );

			if ( ! $post || $this->post_type !== $post->post_type ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => __( '商品不存在或类型不匹配。', 'tanzanite-settings' ),
				);
				continue;
			}

			$updated_fields = array();
			foreach ( $meta_fields as $field => $value ) {
				if ( ! isset( $meta_fields_map[ $field ] ) ) {
					continue;
				}

				$meta_key = $meta_fields_map[ $field ];

				// 类型转换
				if ( in_array( $field, array( 'price_regular', 'price_sale', 'price_member' ), true ) ) {
					$value = (float) $value;
				} elseif ( in_array( $field, array( 'stock_qty', 'stock_alert', 'points_reward', 'points_limit', 'purchase_limit', 'min_purchase', 'shipping_template_id' ), true ) ) {
					$value = (int) $value;
				} elseif ( in_array( $field, array( 'backorders_allowed', 'free_shipping', 'is_sticky' ), true ) ) {
					$value = (bool) $value;
				}

				update_post_meta( $product_id, $meta_key, $value );
				$updated_fields[ $field ] = $value;
			}

			if ( empty( $updated_fields ) ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => __( '未找到可更新的字段。', 'tanzanite-settings' ),
				);
				continue;
			}

			$summary['updated']++;
			$summary['details'][] = array(
				'id'     => $product_id,
				'fields' => array_keys( $updated_fields ),
			);

			$this->log_audit(
				'bulk_set_meta',
				'product',
				$product_id,
				array(
					'action' => 'set_meta',
					'fields' => $updated_fields,
				),
				$request
			);
		}

		return $summary;
	}

	/**
	 * 批量调整价格
	 *
	 * @since 0.2.0
	 */
	private function bulk_adjust_price( $ids, $payload, $summary, $request ) {
		$price_mode = sanitize_key( (string) ( $payload['mode'] ?? '' ) );
		if ( ! in_array( $price_mode, array( 'absolute', 'percent' ), true ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '请选择有效的调价模式。', 'tanzanite-settings' ) );
		}

		if ( ! isset( $payload['value'] ) || ! is_numeric( $payload['value'] ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '调价幅度必须为有效数字。', 'tanzanite-settings' ) );
		}

		$price_value = (float) $payload['value'];
		$allowed_price_fields = array( 'price_regular', 'price_sale', 'price_member' );
		$fields_param = $payload['fields'] ?? array();

		if ( is_string( $fields_param ) ) {
			$fields_param = array_filter( explode( ',', $fields_param ) );
		}

		if ( ! is_array( $fields_param ) ) {
			$fields_param = array();
		}

		$price_fields = array_values( array_intersect( $allowed_price_fields, array_map( 'sanitize_key', $fields_param ) ) );

		if ( empty( $price_fields ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '请选择至少一个需要调价的价格字段。', 'tanzanite-settings' ) );
		}

		$price_round = isset( $payload['round'] ) && is_numeric( $payload['round'] ) ? max( 0, min( 4, (int) $payload['round'] ) ) : 2;

		foreach ( $ids as $product_id ) {
			$post = get_post( $product_id );

			if ( ! $post || $this->post_type !== $post->post_type ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => __( '商品不存在或类型不匹配。', 'tanzanite-settings' ),
				);
				continue;
			}

			$price_updates = array();

			foreach ( $price_fields as $field ) {
				$meta_key = '_tanz_' . $field;
				$current = (float) get_post_meta( $product_id, $meta_key, true );

				if ( 'percent' === $price_mode ) {
					$new_value = $current + ( $current * $price_value / 100 );
				} else {
					$new_value = $current + $price_value;
				}

				if ( $new_value < 0 ) {
					$new_value = 0.0;
				}

				$new_value = round( $new_value, $price_round );
				update_post_meta( $product_id, $meta_key, $new_value );
				$price_updates[ $field ] = $new_value;
			}

			if ( empty( $price_updates ) ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => __( '未能应用调价，请检查选定的字段。', 'tanzanite-settings' ),
				);
				continue;
			}

			$summary['updated']++;
			$summary['details'][] = array(
				'id'     => $product_id,
				'mode'   => $price_mode,
				'value'  => $price_value,
				'fields' => $price_updates,
			);

			$this->log_audit(
				'bulk_adjust_price',
				'product',
				$product_id,
				array(
					'action' => 'adjust_price',
					'mode'   => $price_mode,
					'value'  => $price_value,
					'fields' => $price_updates,
					'round'  => $price_round,
				),
				$request
			);
		}

		return $summary;
	}

	/**
	 * 批量删除
	 *
	 * @since 0.2.0
	 */
	private function bulk_delete( $ids, $payload, $summary, $request ) {
		$delete_mode = sanitize_key( (string) ( $payload['mode'] ?? 'trash' ) );
		if ( ! in_array( $delete_mode, array( 'trash', 'force' ), true ) ) {
			return $this->respond_error( 'invalid_bulk_payload', __( '删除模式无效。', 'tanzanite-settings' ) );
		}

		foreach ( $ids as $product_id ) {
			$post = get_post( $product_id );

			if ( ! $post || $this->post_type !== $post->post_type ) {
				$summary['failed'][] = array(
					'id'     => $product_id,
					'reason' => __( '商品不存在或类型不匹配。', 'tanzanite-settings' ),
				);
				continue;
			}

			if ( 'force' === $delete_mode ) {
				$result = wp_delete_post( $product_id, true );
				if ( false === $result ) {
					$summary['failed'][] = array(
						'id'     => $product_id,
						'reason' => __( '永久删除失败。', 'tanzanite-settings' ),
					);
					continue;
				}

				$this->delete_product_skus( $product_id );
			} else {
				$result = wp_trash_post( $product_id );
				if ( false === $result ) {
					$summary['failed'][] = array(
						'id'     => $product_id,
						'reason' => __( '移动到回收站失败。', 'tanzanite-settings' ),
					);
					continue;
				}
			}

			$summary['updated']++;
			$summary['details'][] = array(
				'id'   => $product_id,
				'mode' => $delete_mode,
			);

			$this->log_audit(
				'bulk_delete',
				'product',
				$product_id,
				array(
					'action' => 'delete',
					'mode'   => $delete_mode,
				),
				$request
			);
		}

		return $summary;
	}

	/**
	 * 批量导出
	 *
	 * @since 0.2.0
	 */
	private function bulk_export( $ids, $request ) {
		$export_data = array();

		foreach ( $ids as $product_id ) {
			$post = get_post( $product_id );

			if ( ! $post || $this->post_type !== $post->post_type ) {
				continue;
			}

			$export_data[] = $this->prepare_item_for_response( $post );
		}

		return $this->respond_success(
			array(
				'items' => $export_data,
				'total' => count( $export_data ),
				'timestamp' => current_time( 'mysql' ),
			)
		);
	}

	/**
	 * 同步商品 SEO 数据
	 *
	 * @since 0.2.0
	 * @param int   $post_id 商品ID
	 * @param mixed $payload SEO 数据
	 * @return array
	 */
	private function sync_product_seo( $post_id, $payload ) {
		// 如果 MyTheme SEO 插件存在，使用插件的 API
		if ( class_exists( '\\MyTheme_SEO_Plugin' ) ) {
			$plugin = \MyTheme_SEO_Plugin::instance();

			if ( is_array( $payload ) ) {
				$update_request = new WP_REST_Request( 'POST', '/mytheme/v1/seo/product/' . $post_id );
				$update_request->set_param( 'id', $post_id );
				$update_request->set_param( 'payload', $payload );
				
				if ( method_exists( $plugin, 'rest_update_product_seo' ) ) {
					$plugin->rest_update_product_seo( $update_request );
				}
			}

			return $this->get_product_seo_payload( $post_id );
		}

		// 降级方案：使用 post meta 存储
		if ( is_array( $payload ) ) {
			$sanitized = $this->sanitize_fallback_seo_payload( $payload );
			update_post_meta( $post_id, '_mytheme_seo_payload', $sanitized );
		}

		return $this->get_product_seo_payload( $post_id );
	}

	/**
	 * 获取商品 SEO 数据
	 *
	 * @since 0.2.0
	 * @param int $post_id 商品ID
	 * @return array
	 */
	private function get_product_seo_payload( $post_id ) {
		// 如果 MyTheme SEO 插件存在，使用插件的 API
		if ( class_exists( '\\MyTheme_SEO_Plugin' ) ) {
			$plugin = \MyTheme_SEO_Plugin::instance();

			$seo_request = new WP_REST_Request( 'GET', '/mytheme/v1/seo/product/' . $post_id );
			$seo_request->set_param( 'id', $post_id );
			
			$payload = array();
			if ( method_exists( $plugin, 'rest_get_product_seo' ) ) {
				$response = $plugin->rest_get_product_seo( $seo_request );

				if ( $response instanceof WP_REST_Response ) {
					$data = $response->get_data();
					if ( is_array( $data ) && isset( $data['payload'] ) && is_array( $data['payload'] ) ) {
						$payload = $data['payload'];
					}
				}
			}

			// 获取可用语言
			$languages = array();
			if ( method_exists( $plugin, 'rest_get_languages' ) ) {
				$languages_response = $plugin->rest_get_languages( new WP_REST_Request( 'GET', '/mytheme/v1/seo/languages' ) );
				if ( $languages_response instanceof WP_REST_Response ) {
					$lang_data = $languages_response->get_data();
					if ( is_array( $lang_data ) && isset( $lang_data['languages'] ) && is_array( $lang_data['languages'] ) ) {
						$languages = $lang_data['languages'];
					}
				}
			}

			return array(
				'available'  => true,
				'configured' => ! empty( $payload ),
				'payload'    => $payload,
				'languages'  => $languages,
			);
		}

		// 降级方案：从 post meta 读取
		$stored = get_post_meta( $post_id, '_mytheme_seo_payload', true );
		$sanitized = $this->sanitize_fallback_seo_payload( $stored );

		return array(
			'available'  => false,
			'configured' => ! empty( $sanitized ),
			'payload'    => $sanitized,
			'languages'  => array(),
		);
	}

	/**
	 * 清理降级 SEO 数据
	 *
	 * @since 0.2.0
	 * @param mixed $payload SEO 数据
	 * @return array
	 */
	private function sanitize_fallback_seo_payload( $payload ) {
		if ( ! is_array( $payload ) ) {
			return array();
		}

		$sanitized = array();
		foreach ( $payload as $locale => $data ) {
			$locale_key = sanitize_key( (string) $locale );
			if ( '' === $locale_key || ! is_array( $data ) ) {
				continue;
			}

			$sanitized[ $locale_key ] = array(
				'title'       => isset( $data['title'] ) ? sanitize_text_field( (string) $data['title'] ) : '',
				'description' => isset( $data['description'] ) ? sanitize_textarea_field( (string) $data['description'] ) : '',
				'keywords'    => isset( $data['keywords'] ) ? sanitize_text_field( (string) $data['keywords'] ) : '',
			);
		}

		return $sanitized;
	}
}
