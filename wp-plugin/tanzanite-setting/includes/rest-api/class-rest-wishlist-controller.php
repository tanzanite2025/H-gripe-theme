<?php
/**
 * Wishlist REST API Controller
 *
 * 只支持已登录用户的心愿单接口
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class Tanzanite_REST_Wishlist_Controller extends Tanzanite_REST_Controller {

    /**
     * REST API 基础路径
     *
     * @var string
     */
    protected $rest_base = 'wishlist';

    /**
     * 心愿单表名
     *
     * @var string
     */
    private $table;

    public function __construct() {
        parent::__construct();
        global $wpdb;
        $this->table = $wpdb->prefix . 'tanz_wishlist_items';
    }

    /**
     * 注册路由
     */
    public function register_routes() {
        // 列表 & 创建
        register_rest_route(
            $this->namespace,
            '/' . $this->rest_base,
            array(
                array(
                    'methods'             => WP_REST_Server::READABLE,
                    'callback'            => array( $this, 'get_items' ),
                    'permission_callback' => 'is_user_logged_in',
                ),
                array(
                    'methods'             => WP_REST_Server::CREATABLE,
                    'callback'            => array( $this, 'create_item' ),
                    'permission_callback' => 'is_user_logged_in',
                ),
            )
        );

        // 删除
        register_rest_route(
            $this->namespace,
            '/' . $this->rest_base . '/(?P<id>\\d+)',
            array(
                array(
                    'methods'             => WP_REST_Server::DELETABLE,
                    'callback'            => array( $this, 'delete_item' ),
                    'permission_callback' => 'is_user_logged_in',
                ),
            )
        );
    }

    /**
     * 获取当前用户的心愿单列表
     *
     * @param WP_REST_Request $request 请求对象
     * @return WP_REST_Response
     */
    public function get_items( $request ) {
        $user_id = get_current_user_id();
        if ( ! $user_id ) {
            return $this->respond_error( 'not_logged_in', __( 'Please log in to view your wishlist.', 'tanzanite-settings' ), 401 );
        }

        global $wpdb;

        $rows = $wpdb->get_results(
            $wpdb->prepare(
                "SELECT id, product_id, created_at FROM {$this->table} WHERE user_id = %d ORDER BY created_at DESC",
                $user_id
            ),
            ARRAY_A
        );

        if ( ! is_array( $rows ) ) {
            $rows = array();
        }

        // 复用商品控制器的数据结构
        if ( ! class_exists( 'Tanzanite_REST_Products_Controller' ) ) {
            return $this->respond_error( 'products_controller_missing', __( 'Products controller not available.', 'tanzanite-settings' ), 500 );
        }

        $products_controller = new Tanzanite_REST_Products_Controller();

        $items = array();
        foreach ( $rows as $row ) {
            $product_id = (int) $row['product_id'];
            $post       = get_post( $product_id );
            if ( ! $post ) {
                continue;
            }

            $product_data = $products_controller->prepare_item_for_response( $post );

            $items[] = array(
                'id'         => (int) $row['id'],
                'product_id' => $product_id,
                'created_at' => $row['created_at'],
                'product'    => $product_data,
            );
        }

        return $this->respond_success(
            array(
                'items' => $items,
                'meta'  => array(
                    'total' => count( $items ),
                ),
            )
        );
    }

    /**
     * 将商品加入心愿单
     *
     * @param WP_REST_Request $request 请求对象
     * @return WP_REST_Response
     */
    public function create_item( $request ) {
        $user_id = get_current_user_id();
        if ( ! $user_id ) {
            return $this->respond_error( 'not_logged_in', __( 'Please log in to use wishlist.', 'tanzanite-settings' ), 401 );
        }

        $product_id = (int) $request->get_param( 'product_id' );
        if ( $product_id <= 0 ) {
            return $this->respond_error( 'invalid_product_id', __( 'Invalid product_id.', 'tanzanite-settings' ), 400 );
        }

        $post = get_post( $product_id );
        if ( ! $post || 'tanz_product' !== $post->post_type ) {
            return $this->respond_error( 'product_not_found', __( 'Product not found.', 'tanzanite-settings' ), 404 );
        }

        global $wpdb;

        // 已存在则直接返回
        $existing = $wpdb->get_row(
            $wpdb->prepare(
                "SELECT id, product_id, created_at FROM {$this->table} WHERE user_id = %d AND product_id = %d",
                $user_id,
                $product_id
            ),
            ARRAY_A
        );

        if ( $existing ) {
            $row = $existing;
        } else {
            $inserted = $wpdb->insert(
                $this->table,
                array(
                    'user_id'    => $user_id,
                    'product_id' => $product_id,
                    'created_at' => current_time( 'mysql', true ),
                ),
                array( '%d', '%d', '%s' )
            );

            if ( false === $inserted ) {
                return $this->respond_error( 'failed_add_wishlist', __( 'Failed to add product to wishlist.', 'tanzanite-settings' ), 500 );
            }

            $row = array(
                'id'         => $wpdb->insert_id,
                'product_id' => $product_id,
                'created_at' => current_time( 'mysql', true ),
            );
        }

        if ( ! class_exists( 'Tanzanite_REST_Products_Controller' ) ) {
            return $this->respond_error( 'products_controller_missing', __( 'Products controller not available.', 'tanzanite-settings' ), 500 );
        }

        $products_controller = new Tanzanite_REST_Products_Controller();
        $product_data        = $products_controller->prepare_item_for_response( $post );

        return $this->respond_success(
            array(
                'item' => array(
                    'id'         => (int) $row['id'],
                    'product_id' => (int) $row['product_id'],
                    'created_at' => $row['created_at'],
                    'product'    => $product_data,
                ),
            ),
            201
        );
    }

    /**
     * 从心愿单中移除商品
     *
     * @param WP_REST_Request $request 请求对象
     * @return WP_REST_Response
     */
    public function delete_item( $request ) {
        $user_id = get_current_user_id();
        if ( ! $user_id ) {
            return $this->respond_error( 'not_logged_in', __( 'Please log in to use wishlist.', 'tanzanite-settings' ), 401 );
        }

        $id = (int) $request['id'];
        if ( $id <= 0 ) {
            return $this->respond_error( 'invalid_id', __( 'Invalid wishlist item id.', 'tanzanite-settings' ), 400 );
        }

        global $wpdb;

        $row = $wpdb->get_row(
            $wpdb->prepare(
                "SELECT id, user_id FROM {$this->table} WHERE id = %d",
                $id
            ),
            ARRAY_A
        );

        if ( ! $row ) {
            return $this->respond_error( 'wishlist_not_found', __( 'Wishlist item not found.', 'tanzanite-settings' ), 404 );
        }

        if ( (int) $row['user_id'] !== $user_id ) {
            return $this->respond_error( 'forbidden', __( 'You cannot modify this wishlist item.', 'tanzanite-settings' ), 403 );
        }

        $deleted = $wpdb->delete(
            $this->table,
            array( 'id' => $id ),
            array( '%d' )
        );

        if ( false === $deleted ) {
            return $this->respond_error( 'failed_delete_wishlist', __( 'Failed to delete wishlist item.', 'tanzanite-settings' ), 500 );
        }

        return $this->respond_success(
            array(
                'deleted' => true,
                'id'      => $id,
            ),
            200
        );
    }
}
