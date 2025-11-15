<?php
/**
 * Audit Logs REST API Controller
 *
 * 处理审计日志相关的 REST API 请求
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 审计日志 REST API 控制器
 *
 * 提供审计日志的列表查询和 CSV 导出功能
 */
class Tanzanite_REST_Audit_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API 基础路径
	 *
	 * @var string
	 */
	protected $rest_base = 'audit-logs';

	/**
	 * 审计日志表名
	 *
	 * @var string
	 */
	private $audit_log_table;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->audit_log_table = $wpdb->prefix . 'tanz_audit_logs';
	}

	/**
	 * 注册路由
	 *
	 * @since 0.2.0
	 */
	public function register_routes() {
		// 获取审计日志列表
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => $this->permission_callback( 'tanz_view_audit_logs', true ),
					'args'                => $this->get_collection_params(),
				),
			)
		);

		// 导出审计日志为 CSV
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/export',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'export_items' ),
					'permission_callback' => $this->permission_callback( 'tanz_view_audit_logs', true ),
					'args'                => $this->get_export_params(),
				),
			)
		);
	}

	/**
	 * 获取审计日志列表
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return WP_REST_Response
	 */
	public function get_items( $request ) {
		global $wpdb;

		// 获取分页参数
		$pagination = $this->get_pagination_params( $request );

		// 构建查询条件
		$where_data = $this->build_where_clause( $request );

		// 获取总数
		$count_query = "SELECT COUNT(*) FROM {$this->audit_log_table} {$where_data['where_sql']}";
		$total       = (int) $wpdb->get_var(
			$where_data['params'] ? $wpdb->prepare( $count_query, $where_data['params'] ) : $count_query
		);

		// 获取数据
		$query                  = "SELECT * FROM {$this->audit_log_table} {$where_data['where_sql']} ORDER BY created_at DESC, id DESC LIMIT %d OFFSET %d";
		$where_data['params'][] = $pagination['per_page'];
		$where_data['params'][] = $pagination['offset'];
		$rows                   = $wpdb->get_results( $wpdb->prepare( $query, $where_data['params'] ), ARRAY_A );

		// 格式化数据
		$items = array();
		if ( is_array( $rows ) ) {
			foreach ( $rows as $row ) {
				$items[] = $this->prepare_item_for_response( $row );
			}
		}

		// 构建响应
		return $this->respond_success(
			array(
				'items' => $items,
				'meta'  => $this->build_pagination_meta( $total, $pagination['page'], $pagination['per_page'] ),
			)
		);
	}

	/**
	 * 导出审计日志为 CSV
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 */
	public function export_items( $request ) {
		global $wpdb;

		// 构建查询条件
		$where_data = $this->build_where_clause( $request );

		// 获取所有数据（限制最多 10000 条）
		$query = "SELECT * FROM {$this->audit_log_table} {$where_data['where_sql']} ORDER BY created_at DESC, id DESC LIMIT 10000";
		$rows  = $wpdb->get_results(
			$where_data['params'] ? $wpdb->prepare( $query, $where_data['params'] ) : $query,
			ARRAY_A
		);

		// 设置 CSV 响应头
		header( 'Content-Type: text/csv; charset=utf-8' );
		header( 'Content-Disposition: attachment; filename="audit-logs-' . gmdate( 'Y-m-d-His' ) . '.csv"' );
		header( 'Pragma: no-cache' );
		header( 'Expires: 0' );

		// 输出 BOM 以支持 Excel 正确显示中文
		echo "\xEF\xBB\xBF";

		// 打开输出流
		$output = fopen( 'php://output', 'w' );

		// 写入表头
		fputcsv( $output, array( 'ID', '时间', '操作人', '动作', '目标类型', '目标ID', 'IP地址', '详情' ) );

		// 写入数据
		if ( is_array( $rows ) ) {
			foreach ( $rows as $row ) {
				fputcsv(
					$output,
					array(
						$row['id'],
						$row['created_at'],
						$row['actor_name'],
						$row['action'],
						$row['target_type'],
						$row['target_id'],
						$row['ip_address'],
						$row['payload'],
					)
				);
			}
		}

		fclose( $output );
		exit;
	}

	/**
	 * 获取字段的唯一值（用于筛选下拉框）
	 *
	 * @since 0.2.0
	 * @param string $column 字段名
	 * @param int    $limit  最多返回数量
	 * @return array
	 */
	public function get_distinct_values( $column, $limit = 50 ) {
		global $wpdb;

		$allowed_columns = array( 'action', 'target_type', 'actor_name' );
		if ( ! in_array( $column, $allowed_columns, true ) ) {
			return array();
		}

		$results = $wpdb->get_col(
			$wpdb->prepare(
				"SELECT DISTINCT {$column} FROM {$this->audit_log_table} WHERE {$column} IS NOT NULL AND {$column} != '' ORDER BY {$column} ASC LIMIT %d",
				$limit
			)
		);

		return is_array( $results ) ? $results : array();
	}

	/**
	 * 构建 WHERE 子句
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return array 包含 where_sql 和 params 的数组
	 */
	private function build_where_clause( $request ) {
		global $wpdb;

		$where  = array();
		$params = array();

		// 动作筛选
		if ( $action = $request->get_param( 'action' ) ) {
			$where[]  = 'action = %s';
			$params[] = sanitize_text_field( $action );
		}

		// 目标类型筛选
		if ( $target_type = $request->get_param( 'target_type' ) ) {
			$where[]  = 'target_type = %s';
			$params[] = sanitize_text_field( $target_type );
		}

		// 操作人筛选
		if ( $actor = $request->get_param( 'actor' ) ) {
			$where[]  = 'actor_name = %s';
			$params[] = sanitize_text_field( $actor );
		}

		// 关键词搜索
		if ( $search = $request->get_param( 'search' ) ) {
			$search_term = '%' . $wpdb->esc_like( sanitize_text_field( $search ) ) . '%';
			$where[]     = '(action LIKE %s OR target_type LIKE %s OR payload LIKE %s)';
			$params[]    = $search_term;
			$params[]    = $search_term;
			$params[]    = $search_term;
		}

		// 开始日期
		if ( $start_date = $request->get_param( 'start_date' ) ) {
			$where[]  = 'created_at >= %s';
			$params[] = sanitize_text_field( $start_date ) . ' 00:00:00';
		}

		// 结束日期
		if ( $end_date = $request->get_param( 'end_date' ) ) {
			$where[]  = 'created_at <= %s';
			$params[] = sanitize_text_field( $end_date ) . ' 23:59:59';
		}

		$where_sql = $where ? 'WHERE ' . implode( ' AND ', $where ) : '';

		return array(
			'where_sql' => $where_sql,
			'params'    => $params,
		);
	}

	/**
	 * 格式化单条记录用于响应
	 *
	 * @since 0.2.0
	 * @param array $row 数据库行
	 * @return array
	 */
	private function prepare_item_for_response( $row ) {
		return array(
			'id'          => (int) $row['id'],
			'actor_id'    => (int) $row['actor_id'],
			'actor_name'  => $row['actor_name'],
			'action'      => $row['action'],
			'target_type' => $row['target_type'],
			'target_id'   => (int) $row['target_id'],
			'payload'     => json_decode( $row['payload'], true ),
			'ip_address'  => $row['ip_address'],
			'created_at'  => $row['created_at'],
		);
	}

	/**
	 * 获取集合参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_collection_params() {
		return array(
			'page'        => array(
				'type'    => 'integer',
				'default' => 1,
			),
			'per_page'    => array(
				'type'    => 'integer',
				'default' => 20,
			),
			'action'      => array(
				'type' => 'string',
			),
			'target_type' => array(
				'type' => 'string',
			),
			'actor'       => array(
				'type' => 'string',
			),
			'search'      => array(
				'type' => 'string',
			),
			'start_date'  => array(
				'type' => 'string',
			),
			'end_date'    => array(
				'type' => 'string',
			),
		);
	}

	/**
	 * 获取导出参数定义
	 *
	 * @since 0.2.0
	 * @return array
	 */
	private function get_export_params() {
		return array(
			'action'      => array(
				'type' => 'string',
			),
			'target_type' => array(
				'type' => 'string',
			),
			'actor'       => array(
				'type' => 'string',
			),
			'search'      => array(
				'type' => 'string',
			),
			'start_date'  => array(
				'type' => 'string',
			),
			'end_date'    => array(
				'type' => 'string',
			),
		);
	}
}
