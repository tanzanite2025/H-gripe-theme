<?php
/**
 * REST API Base Controller
 *
 * 所有 REST API 控制器的基类
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 * @since      0.2.0
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * REST API 基础控制器抽象类
 *
 * 提供通用的权限检查、错误响应等方法
 */
abstract class Tanzanite_REST_Controller {

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
	protected $rest_base;

	/**
	 * 数据库表前缀
	 *
	 * @var string
	 */
	protected $table_prefix;

	/**
	 * 构造函数
	 *
	 * @since 0.2.0
	 */
	public function __construct() {
		global $wpdb;
		$this->table_prefix = $wpdb->prefix . 'tanz_';
	}

	/**
	 * 注册路由（子类必须实现）
	 *
	 * @since 0.2.0
	 */
	abstract public function register_routes();

	/**
	 * 创建权限检查回调
	 *
	 * @since 0.2.0
	 * @param string $capability 需要的权限
	 * @param bool   $return_callable 是否返回可调用对象
	 * @return callable|bool
	 */
	protected function permission_callback( $capability, $return_callable = false ) {
		if ( $return_callable ) {
			return function() use ( $capability ) {
				return current_user_can( $capability );
			};
		}
		return current_user_can( $capability );
	}

	/**
	 * 返回错误响应
	 *
	 * @since 0.2.0
	 * @param string $code    错误代码
	 * @param string $message 错误消息
	 * @param int    $status  HTTP 状态码
	 * @return WP_REST_Response
	 */
	protected function respond_error( $code, $message, $status = 400 ) {
		return new WP_REST_Response(
			array(
				'code'    => $code,
				'message' => $message,
			),
			$status
		);
	}

	/**
	 * 返回成功响应
	 *
	 * @since 0.2.0
	 * @param mixed $data   响应数据
	 * @param int   $status HTTP 状态码
	 * @return WP_REST_Response
	 */
	protected function respond_success( $data, $status = 200 ) {
		return new WP_REST_Response( $data, $status );
	}

	/**
	 * 验证必需参数
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @param array           $required_params 必需参数列表
	 * @return WP_Error|bool 验证失败返回 WP_Error，成功返回 true
	 */
	protected function validate_required_params( $request, $required_params ) {
		foreach ( $required_params as $param ) {
			if ( ! $request->has_param( $param ) || empty( $request->get_param( $param ) ) ) {
				return new WP_Error(
					'missing_parameter',
					sprintf( __( '缺少必需参数: %s', 'tanzanite-settings' ), $param ),
					array( 'status' => 400 )
				);
			}
		}
		return true;
	}

	/**
	 * 获取分页参数
	 *
	 * @since 0.2.0
	 * @param WP_REST_Request $request REST 请求对象
	 * @return array 包含 page, per_page, offset 的数组
	 */
	protected function get_pagination_params( $request ) {
		$page     = max( 1, (int) $request->get_param( 'page' ) );
		$per_page = min( 100, max( 1, (int) $request->get_param( 'per_page' ) ) );
		$offset   = ( $page - 1 ) * $per_page;

		return array(
			'page'     => $page,
			'per_page' => $per_page,
			'offset'   => $offset,
		);
	}

	/**
	 * 构建分页元数据
	 *
	 * @since 0.2.0
	 * @param int $total    总记录数
	 * @param int $page     当前页码
	 * @param int $per_page 每页记录数
	 * @return array 分页元数据
	 */
	protected function build_pagination_meta( $total, $page, $per_page ) {
		$total_pages = $per_page ? (int) ceil( $total / $per_page ) : 1;

		return array(
			'total'       => $total,
			'total_pages' => $total_pages,
			'page'        => $page,
			'per_page'    => $per_page,
		);
	}

	/**
	 * 记录审计日志
	 *
	 * @since 0.2.0
	 * @param string          $action      操作类型
	 * @param string          $target_type 目标类型
	 * @param int             $target_id   目标 ID
	 * @param array           $payload     操作详情
	 * @param WP_REST_Request $request     REST 请求对象
	 */
	protected function log_audit( $action, $target_type, $target_id, $payload, $request ) {
		global $wpdb;

		$user_id   = get_current_user_id();
		$user_name = '';
		
		if ( $user_id ) {
			$user = get_userdata( $user_id );
			if ( $user ) {
				$user_name = $user->display_name ?: $user->user_login;
			}
		}

		// 获取 IP 地址
		$ip = $request->get_header( 'x-forwarded-for' );
		if ( $ip ) {
			$ip = trim( explode( ',', $ip )[0] );
		}
		if ( ! $ip ) {
			$ip = $_SERVER['REMOTE_ADDR'] ?? '';
		}

		$wpdb->insert(
			$this->table_prefix . 'audit_logs',
			array(
				'actor_id'    => $user_id,
				'actor_name'  => $user_name,
				'action'      => $action,
				'target_type' => $target_type,
				'target_id'   => $target_id,
				'payload'     => wp_json_encode( $payload ),
				'ip_address'  => $ip,
			),
			array( '%d', '%s', '%s', '%s', '%d', '%s', '%s' )
		);
	}

	/**
	 * 清理和验证数据
	 *
	 * @since 0.2.0
	 * @param array $data  要清理的数据
	 * @param array $rules 验证规则
	 * @return array 清理后的数据
	 */
	protected function sanitize_data( $data, $rules ) {
		$sanitized = array();

		foreach ( $rules as $field => $rule ) {
			if ( ! isset( $data[ $field ] ) ) {
				continue;
			}

			$value = $data[ $field ];

			switch ( $rule ) {
				case 'text':
					$sanitized[ $field ] = sanitize_text_field( $value );
					break;
				case 'textarea':
					$sanitized[ $field ] = sanitize_textarea_field( $value );
					break;
				case 'html':
					$sanitized[ $field ] = wp_kses_post( $value );
					break;
				case 'email':
					$sanitized[ $field ] = sanitize_email( $value );
					break;
				case 'url':
					$sanitized[ $field ] = esc_url_raw( $value );
					break;
				case 'int':
					$sanitized[ $field ] = (int) $value;
					break;
				case 'float':
					$sanitized[ $field ] = (float) $value;
					break;
				case 'bool':
					$sanitized[ $field ] = (bool) $value;
					break;
				case 'array':
					$sanitized[ $field ] = is_array( $value ) ? $value : array();
					break;
				case 'json':
					$sanitized[ $field ] = is_string( $value ) ? json_decode( $value, true ) : $value;
					break;
				default:
					$sanitized[ $field ] = $value;
			}
		}

		return $sanitized;
	}

	/**
	 * 检查数据库错误
	 *
	 * @since 0.2.0
	 * @param string $context 错误上下文
	 * @return WP_REST_Response|null 有错误返回响应，无错误返回 null
	 */
	protected function check_db_error( $context = 'database_error' ) {
		global $wpdb;

		if ( $wpdb->last_error ) {
			if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
				error_log( "Database error in {$context}: " . $wpdb->last_error );
			}
			return $this->respond_error(
				$context,
				__( '数据库操作失败，请稍后重试。', 'tanzanite-settings' ),
				500
			);
		}

		return null;
	}
}
