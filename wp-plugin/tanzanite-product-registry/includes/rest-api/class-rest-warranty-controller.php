<?php
/**
 * 保修查询 REST API 控制器
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 保修查询 REST API 控制器
 * 
 * 继承 tanzanite-setting 的基类模式
 */
class Tanzanite_REST_Warranty_Controller {

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
	protected $rest_base = 'warranty';

	/**
	 * 数据库实例
	 *
	 * @var Tanzanite_PR_Database
	 */
	private $db;

	/**
	 * 限流：每用户每分钟最多查询次数
	 *
	 * @var int
	 */
	private $rate_limit = 5;

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
		// 查询保修状态
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<code>[a-zA-Z0-9\-_]+)',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_warranty_status' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => array(
						'code' => array(
							'required'          => true,
							'type'              => 'string',
							'sanitize_callback' => 'sanitize_text_field',
						),
					),
				),
			)
		);
	}

	/**
	 * 获取保修状态
	 *
	 * @param WP_REST_Request $request 请求对象
	 * @return WP_REST_Response|WP_Error
	 */
	public function get_warranty_status( $request ) {
		$code = $request->get_param( 'code' );
		$user_id = get_current_user_id();

		// 限流检查
		if ( ! $this->check_rate_limit( $user_id ) ) {
			return new WP_Error(
				'rate_limit_exceeded',
				'查询过于频繁，请稍后再试。',
				array( 'status' => 429 )
			);
		}

		// 记录查询
		$this->log_query( $user_id, $code );

		// 查询产品
		$product = $this->db->get_product_by_code( $code );

		if ( ! $product ) {
			return new WP_Error(
				'not_found',
				'Product not found.',
				array( 'status' => 404 )
			);
		}

		// 计算保修状态
		$warranty_data = $this->calculate_warranty( $product );

		// 获取保修记录
		$records = $this->db->get_warranty_records( $product['id'] );
		$formatted_records = array();
		
		$record_type_names = array(
			'repair'  => array( 'en' => 'Repair', 'zh' => '维修' ),
			'extend'  => array( 'en' => 'Warranty Extension', 'zh' => '延保' ),
			'replace' => array( 'en' => 'Replacement', 'zh' => '换货' ),
		);

		foreach ( $records as $record ) {
			$type_name = $record_type_names[ $record['record_type'] ] ?? array( 'en' => $record['record_type'], 'zh' => $record['record_type'] );
			$formatted_records[] = array(
				'type'        => $record['record_type'],
				'type_name'   => $type_name['en'],
				'type_name_zh' => $type_name['zh'],
				'date'        => $record['record_date'],
				'description' => $record['description'],
			);
		}

		return rest_ensure_response( array(
			'success' => true,
			'data'    => array(
				'product_code'  => $product['product_code'],
				'product_type'  => array(
					'code'    => $product['type_code'],
					'name'    => $product['type_name_en'],
					'name_zh' => $product['type_name'],
				),
				'product_name'  => $product['product_name'],
				'ship_date'     => substr( $product['ship_date'], 0, 7 ), // YYYY-MM
				'warranty_months' => intval( $product['warranty_months'] ),
				'warranty_end'  => $warranty_data['warranty_end'],
				'status'        => $warranty_data['status'],
				'remaining'     => $warranty_data['remaining'],
				'records'       => $formatted_records,
			),
		) );
	}

	/**
	 * 计算保修状态
	 *
	 * @param array $product 产品数据
	 * @return array
	 */
	private function calculate_warranty( $product ) {
		$ship_date = new DateTime( $product['ship_date'] );
		$warranty_months = intval( $product['warranty_months'] );
		
		// 加上延保月数
		$extend_months = $this->db->get_total_extend_months( $product['id'] );
		$total_months = $warranty_months + $extend_months;

		$warranty_end = clone $ship_date;
		$warranty_end->modify( "+{$total_months} months" );

		$today = new DateTime();
		$diff = $today->diff( $warranty_end );

		$result = array(
			'warranty_end' => $warranty_end->format( 'Y-m' ),
		);

		if ( $warranty_end > $today ) {
			$result['status'] = 'valid';
			$result['remaining'] = array(
				'months'     => ( $diff->y * 12 ) + $diff->m,
				'days'       => $diff->d,
				'total_days' => $diff->days,
			);
		} else {
			$result['status'] = 'expired';
			$result['remaining'] = array(
				'months'     => 0,
				'days'       => 0,
				'total_days' => 0,
				'expired_days' => $diff->days,
			);
		}

		return $result;
	}

	/**
	 * 检查限流
	 *
	 * @param int $user_id 用户ID
	 * @return bool
	 */
	private function check_rate_limit( $user_id ) {
		$transient_key = 'tanzanite_pr_rate_' . $user_id;
		$count = get_transient( $transient_key );

		if ( false === $count ) {
			set_transient( $transient_key, 1, 60 ); // 60秒过期
			return true;
		}

		if ( $count >= $this->rate_limit ) {
			return false;
		}

		set_transient( $transient_key, $count + 1, 60 );
		return true;
	}

	/**
	 * 记录查询日志
	 *
	 * @param int    $user_id 用户ID
	 * @param string $code    产品编码
	 */
	private function log_query( $user_id, $code ) {
		// 可以扩展为写入审计日志表
		// 目前仅记录到 WordPress 日志
		if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
			error_log( sprintf(
				'[Tanzanite PR] Warranty query: user=%d, code=%s, ip=%s',
				$user_id,
				$code,
				$_SERVER['REMOTE_ADDR'] ?? 'unknown'
			) );
		}
	}
}
