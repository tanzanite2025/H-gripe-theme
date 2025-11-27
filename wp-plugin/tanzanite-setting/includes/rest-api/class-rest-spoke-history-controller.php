<?php
/**
 * Spoke History REST API Controller
 *
 * 提供辐条长度历史数据的只读查询接口
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.1
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

/**
 * 辐条长度历史 REST API 控制器
 */
class Tanzanite_REST_Spoke_History_Controller extends Tanzanite_REST_Controller {

    /**
     * REST API 基础路径
     *
     * @var string
     */
    protected $rest_base = 'spoke-history';

    /**
     * 历史记录表名
     *
     * @var string
     */
    private $table;

    /**
     * 构造函数
     *
     * @since 0.2.1
     */
    public function __construct() {
        parent::__construct();
        global $wpdb;
        $this->table = $wpdb->prefix . 'tanz_spoke_history';
    }

    /**
     * 注册路由
     *
     * @since 0.2.1
     */
    public function register_routes() {
        register_rest_route(
            $this->namespace,
            '/' . $this->rest_base,
            array(
                array(
                    'methods'             => WP_REST_Server::READABLE,
                    'callback'            => array( $this, 'get_items' ),
                    'permission_callback' => '__return_true', // 只读公开
                    'args'                => $this->get_collection_params(),
                ),
            )
        );
    }

    /**
     * 获取历史记录列表
     *
     * @since 0.2.1
     *
     * @param WP_REST_Request $request REST 请求对象
     *
     * @return WP_REST_Response
     */
    public function get_items( $request ) {
        global $wpdb;

        $pagination = $this->get_pagination_params( $request );

        // 默认每页 5 条记录
        if ( ! $request->has_param( 'per_page' ) || ! $pagination['per_page'] ) {
            $pagination['per_page'] = 5;
            $pagination['offset']   = ( $pagination['page'] - 1 ) * $pagination['per_page'];
        }

        $search = sanitize_text_field( (string) $request->get_param( 'search' ) );

        $where  = array();
        $params = array();

        if ( '' !== $search ) {
            $like     = '%' . $wpdb->esc_like( $search ) . '%';
            $where[]  = '(hub_model LIKE %s OR hub_brand LIKE %s OR rim_model LIKE %s)';
            $params[] = $like;
            $params[] = $like;
            $params[] = $like;
        }

        $where_sql = $where ? 'WHERE ' . implode( ' AND ', $where ) : '';

        $query    = "SELECT * FROM {$this->table} {$where_sql} ORDER BY created_at DESC LIMIT %d OFFSET %d";
        $params[] = (int) $pagination['per_page'];
        $params[] = (int) $pagination['offset'];

        if ( $where_sql ) {
            $rows = $wpdb->get_results( $wpdb->prepare( $query, $params ), ARRAY_A );
        } else {
            // 无任何过滤条件时不需要 prepare
            $rows = $wpdb->get_results( $query, ARRAY_A );
        }

        $db_error = $this->check_db_error( 'spoke_history_get_items' );
        if ( $db_error ) {
            return $db_error;
        }

        // 统计总数
        $count_query = "SELECT COUNT(*) FROM {$this->table} {$where_sql}";

        if ( $where_sql ) {
            $count_params = $params;
            // 去掉 LIMIT / OFFSET
            array_pop( $count_params );
            array_pop( $count_params );
            $total = (int) $wpdb->get_var( $wpdb->prepare( $count_query, $count_params ) );
        } else {
            $total = (int) $wpdb->get_var( $count_query );
        }

        $items = array_map( array( $this, 'format_row' ), $rows );

        return $this->respond_success(
            array(
                'items' => $items,
                'meta'  => $this->build_pagination_meta( $total, $pagination['page'], $pagination['per_page'] ),
            )
        );
    }

    /**
     * 集合参数定义
     *
     * @since 0.2.1
     *
     * @return array
     */
    private function get_collection_params() {
        return array(
            'search'   => array(
                'type' => 'string',
            ),
            'page'     => array(
                'type'    => 'integer',
                'default' => 1,
            ),
            'per_page' => array(
                'type'    => 'integer',
                'default' => 5,
            ),
        );
    }

    /**
     * 格式化单条历史记录
     *
     * @since 0.2.1
     *
     * @param array $row 数据库行
     *
     * @return array
     */
    private function format_row( $row ) {
        return array(
            'id'                        => (int) $row['id'],
            'wheel_type'                => $row['wheel_type'],
            'source_type'               => $row['source_type'],
            'rim_brand'                 => $row['rim_brand'],
            'rim_model'                 => $row['rim_model'],
            'hub_brand'                 => $row['hub_brand'],
            'hub_model'                 => $row['hub_model'],
            'erd_mm'                    => null !== $row['erd_mm'] ? (float) $row['erd_mm'] : null,
            'left_flange_pcd_mm'        => null !== $row['left_flange_pcd_mm'] ? (float) $row['left_flange_pcd_mm'] : null,
            'right_flange_pcd_mm'       => null !== $row['right_flange_pcd_mm'] ? (float) $row['right_flange_pcd_mm'] : null,
            'left_flange_to_center_mm'  => null !== $row['left_flange_to_center_mm'] ? (float) $row['left_flange_to_center_mm'] : null,
            'right_flange_to_center_mm' => null !== $row['right_flange_to_center_mm'] ? (float) $row['right_flange_to_center_mm'] : null,
            'spoke_count'               => null !== $row['spoke_count'] ? (int) $row['spoke_count'] : null,
            'lacing_pattern'            => $row['lacing_pattern'],
            'nipple_type'               => $row['nipple_type'],
            'left_length_mm'            => null !== $row['left_length_mm'] ? (float) $row['left_length_mm'] : null,
            'right_length_mm'           => null !== $row['right_length_mm'] ? (float) $row['right_length_mm'] : null,
            'created_at'                => $row['created_at'],
            'updated_at'                => $row['updated_at'],
        );
    }
}
