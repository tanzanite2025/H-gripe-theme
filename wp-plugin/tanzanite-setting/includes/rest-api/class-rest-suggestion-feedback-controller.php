<?php
/**
 * Suggestion Feedback REST Controller
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Tanzanite_REST_Suggestion_Feedback_Controller extends Tanzanite_REST_Controller {
	protected $rest_base = 'suggestion-feedback';

	private $table;

	const DEFAULT_ATTACHMENT_LEVEL = 'silver';

	public function __construct() {
		parent::__construct();
		global $wpdb;
		$this->table = $wpdb->prefix . 'tanz_feedback_suggestions';
	}

	public function register_routes() {
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_items' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
					'args'                => $this->get_collection_params(),
				),
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'create_item' ),
					'permission_callback' => 'is_user_logged_in',
					'args'                => $this->get_create_params(),
				),
			)
		);

		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/(?P<id>\d+)/status',
			array(
				array(
					'methods'             => WP_REST_Server::EDITABLE,
					'callback'            => array( $this, 'update_status' ),
					'permission_callback' => $this->permission_callback( 'manage_options', true ),
					'args'                => array(
						'id'     => array(
							'validate_callback' => array( $this, 'validate_numeric_param' ),
						),
						'status' => array(
							'type'     => 'string',
							'required' => true,
						),
					),
				),
			)
		);

		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base . '/eligibility',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_eligibility' ),
					'permission_callback' => '__return_true',
				),
			)
		);
	}

	public function get_items( $request ) {
		global $wpdb;

		$pagination = $this->get_pagination_params( $request );
		$where      = array();
		$params     = array();

		$status = sanitize_key( (string) $request->get_param( 'status' ) );
		if ( $status ) {
			$where[]  = 'status = %s';
			$params[] = $status;
		}

		$search = $request->get_param( 'search' );
		if ( $search ) {
			$like     = '%' . $wpdb->esc_like( $search ) . '%';
			$where[]  = '(full_name LIKE %s OR email LIKE %s OR message LIKE %s OR order_number LIKE %s)';
			$params[] = $like;
			$params[] = $like;
			$params[] = $like;
			$params[] = $like;
		}

		$where_sql = $where ? 'WHERE ' . implode( ' AND ', $where ) : '';

		$query    = "SELECT * FROM {$this->table} {$where_sql} ORDER BY created_at DESC LIMIT %d OFFSET %d";
		$params[] = (int) $pagination['per_page'];
		$params[] = (int) $pagination['offset'];

		$rows = $wpdb->get_results( $wpdb->prepare( $query, $params ), ARRAY_A );
		$count_params = $where ? array_slice( $params, 0, count( $params ) - 2 ) : array();
		$count_query  = "SELECT COUNT(*) FROM {$this->table} {$where_sql}";
		$total        = $where ? (int) $wpdb->get_var( $wpdb->prepare( $count_query, $count_params ) ) : (int) $wpdb->get_var( $count_query );

		$db_error = $this->check_db_error( 'suggestion_feedback_get_items' );
		if ( $db_error ) {
			return $db_error;
		}

		$items = array_map( array( $this, 'format_row' ), $rows );

		return $this->respond_success(
			array(
				'data'       => $items,
				'pagination' => $this->build_pagination_meta( $total, $pagination['page'], $pagination['per_page'] ),
			)
		);
	}

	public function create_item( $request ) {
		global $wpdb;

		$user_id = get_current_user_id();
		if ( ! $user_id ) {
			return $this->respond_error( 'not_logged_in', __( '请先登录后再提交反馈。', 'tanzanite-settings' ), 401 );
		}

		$payload  = $this->sanitize_payload( $request );
		$required = array( 'message' );
		foreach ( $required as $field ) {
			if ( empty( $payload[ $field ] ) ) {
				return $this->respond_error( 'missing_field', sprintf( __( '缺少字段: %s', 'tanzanite-settings' ), $field ) );
			}
		}

		$member_level = $this->get_user_membership_level( $user_id );
		$required_lvl = $this->get_required_level();
		$met          = $this->member_level_meets_requirement( $member_level, $required_lvl );

		if ( ! $met ) {
			$payload['attachments'] = array();
		}

		$inserted = $wpdb->insert(
			$this->table,
			array(
				'user_id'               => $user_id,
				'full_name'             => $payload['full_name'],
				'email'                 => $payload['email'],
				'country'               => $payload['country'],
				'order_number'          => $payload['order_number'],
				'product_category'      => $payload['product_category'],
				'request_type'          => $payload['request_type'],
				'message'               => $payload['message'],
				'attachments'           => wp_json_encode( $payload['attachments'] ),
				'meta'                  => wp_json_encode( $payload['meta'] ),
				'status'                => 'new',
				'member_level_required' => $required_lvl,
				'member_level_met'      => $met ? 1 : 0,
				'eligibility_hash'      => $payload['eligibility_hash'],
				'created_at'            => current_time( 'mysql' ),
				'updated_at'            => current_time( 'mysql' ),
			),
			array( '%d', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%d', '%s', '%s', '%s' )
		);

		if ( false === $inserted ) {
			$db_error = $this->check_db_error( 'suggestion_feedback_create' );
			if ( $db_error ) {
				return $db_error;
			}
			return $this->respond_error( 'create_failed', __( '创建反馈失败，请稍后再试。', 'tanzanite-settings' ), 500 );
		}

		$insert_id = $wpdb->insert_id;

		$this->log_audit( 'create', 'suggestion_feedback', $insert_id, array( 'request_type' => $payload['request_type'] ), $request );

		return $this->respond_success(
			array(
				'id'      => $insert_id,
				'status'  => 'new',
				'message' => __( '反馈已提交，客服会尽快审核。', 'tanzanite-settings' ),
			),
			201
		);
	}

	public function update_status( $request ) {
		global $wpdb;

		$id     = (int) $request['id'];
		$status = sanitize_key( (string) $request->get_param( 'status' ) );

		$allowed = array( 'new', 'in_review', 'resolved', 'archived' );
		if ( ! in_array( $status, $allowed, true ) ) {
			return $this->respond_error( 'invalid_status', __( '无效的状态值。', 'tanzanite-settings' ), 400 );
		}

		$current_user = get_current_user_id();

		$updated = $wpdb->update(
			$this->table,
			array(
				'status'      => $status,
				'updated_at'  => current_time( 'mysql' ),
				'reviewed_by' => $current_user ?: null,
				'reviewed_at' => current_time( 'mysql' ),
			),
			array( 'id' => $id ),
			array( '%s', '%s', '%d', '%s' ),
			array( '%d' )
		);

		if ( false === $updated ) {
			$db_error = $this->check_db_error( 'suggestion_feedback_update_status' );
			if ( $db_error ) {
				return $db_error;
			}
			return $this->respond_error( 'update_failed', __( '更新失败，请稍后再试。', 'tanzanite-settings' ), 500 );
		}

		$this->log_audit( 'update_status', 'suggestion_feedback', $id, array( 'status' => $status ), $request );

		return $this->respond_success(
			array(
				'id'     => $id,
				'status' => $status,
			)
		);
	}

	public function get_eligibility( $request ) {
		$user_id  = get_current_user_id();
		$logged   = (bool) $user_id;
		$level    = $logged ? $this->get_user_membership_level( $user_id ) : '';
		$required = $this->get_required_level();
		$met      = $this->member_level_meets_requirement( $level, $required );

		return $this->respond_success(
			array(
				'loggedIn'        => $logged,
				'canAttach'       => $logged && $met,
				'requiredLevel'   => $required,
				'userLevel'       => $level,
				'reason'          => $logged ? ( $met ? null : __( '当前会员等级暂不支持上传图片。', 'tanzanite-settings' ) ) : __( '请先登录。', 'tanzanite-settings' ),
			)
		);
	}

	private function sanitize_payload( WP_REST_Request $request ) {
		return array(
			'full_name'        => sanitize_text_field( (string) $request->get_param( 'fullName' ) ),
			'email'            => sanitize_email( (string) $request->get_param( 'email' ) ),
			'country'          => sanitize_text_field( (string) $request->get_param( 'country' ) ),
			'order_number'     => sanitize_text_field( (string) $request->get_param( 'orderNumber' ) ),
			'product_category' => sanitize_text_field( (string) $request->get_param( 'productCategory' ) ),
			'request_type'     => sanitize_text_field( (string) $request->get_param( 'requestType' ) ),
			'message'          => wp_kses_post( (string) $request->get_param( 'message' ) ),
			'attachments'      => $this->sanitize_attachments( $request->get_param( 'attachments' ) ),
			'meta'             => array(
				'ip'      => $this->get_user_ip(),
				'agent'   => isset( $_SERVER['HTTP_USER_AGENT'] ) ? sanitize_text_field( wp_unslash( $_SERVER['HTTP_USER_AGENT'] ) ) : '',
				'locale'  => determine_locale(),
			),
			'eligibility_hash' => sanitize_text_field( (string) wp_hash( maybe_serialize( $_SERVER['HTTP_USER_AGENT'] ?? '' ) ) ),
		);
	}

	private function sanitize_attachments( $attachments ) {
		if ( ! is_array( $attachments ) ) {
			return array();
		}

		$clean = array();
		foreach ( $attachments as $attachment ) {
			if ( empty( $attachment['url'] ) ) {
				continue;
			}
			$clean[] = array(
				'name' => isset( $attachment['name'] ) ? sanitize_text_field( (string) $attachment['name'] ) : '',
				'url'  => esc_url_raw( $attachment['url'] ),
				'size' => isset( $attachment['size'] ) ? (int) $attachment['size'] : 0,
			);
		}

		return array_slice( $clean, 0, 3 );
	}

	private function format_row( $row ) {
		return array(
			'id'           => (int) $row['id'],
			'user_id'      => (int) $row['user_id'],
			'full_name'    => $row['full_name'],
			'email'        => $row['email'],
			'country'      => $row['country'],
			'order_number' => $row['order_number'],
			'product_category' => $row['product_category'],
			'request_type' => $row['request_type'],
			'message'      => $row['message'],
			'attachments'  => $row['attachments'] ? json_decode( $row['attachments'], true ) : array(),
			'created_at'   => $row['created_at'],
			'updated_at'   => $row['updated_at'],
			'status'       => $row['status'],
			'member_level_required' => $row['member_level_required'],
			'member_level_met'      => (bool) $row['member_level_met'],
			'meta'         => $row['meta'] ? json_decode( $row['meta'], true ) : array(),
		);
	}

	private function get_user_membership_level( $user_id ) {
		$level = get_user_meta( $user_id, 'membership_level', true );
		return sanitize_text_field( (string) $level );
	}

	private function get_required_level() {
		return sanitize_text_field( (string) apply_filters( 'tanzanite_suggestion_required_membership_level', self::DEFAULT_ATTACHMENT_LEVEL ) );
	}

	private function member_level_meets_requirement( $user_level, $required_level ) {
		if ( empty( $required_level ) ) {
			return true;
		}

		if ( empty( $user_level ) ) {
			return false;
		}

		$hierarchy = apply_filters( 'tanzanite_membership_hierarchy', array( 'bronze', 'silver', 'gold', 'platinum' ) );
		$hierarchy = array_map( 'strtolower', (array) $hierarchy );
		$user_idx  = array_search( strtolower( $user_level ), $hierarchy, true );
		$req_idx   = array_search( strtolower( $required_level ), $hierarchy, true );

		if ( false === $req_idx ) {
			return true;
		}

		if ( false === $user_idx ) {
			return false;
		}

		return $user_idx >= $req_idx;
	}

	private function get_user_ip() {
		foreach ( array( 'HTTP_CLIENT_IP', 'HTTP_X_FORWARDED_FOR', 'REMOTE_ADDR' ) as $key ) {
			if ( ! empty( $_SERVER[ $key ] ) ) {
				return sanitize_text_field( wp_unslash( $_SERVER[ $key ] ) );
			}
		}
		return '';
	}

	private function get_collection_params() {
		return array(
			'page' => array(
				'type'    => 'integer',
				'default' => 1,
			),
			'per_page' => array(
				'type'    => 'integer',
				'default' => 20,
			),
			'status' => array( 'type' => 'string' ),
			'search' => array( 'type' => 'string' ),
		);
	}

	private function get_create_params() {
		return array(
			'fullName' => array( 'type' => 'string' ),
			'email'    => array( 'type' => 'string' ),
			'country'  => array( 'type' => 'string' ),
			'orderNumber' => array( 'type' => 'string' ),
			'productCategory' => array( 'type' => 'string' ),
			'requestType'     => array( 'type' => 'string' ),
			'message'         => array( 'type' => 'string', 'required' => true ),
			'attachments'     => array( 'type' => 'array' ),
		);
	}

	protected function check_db_error( $context = 'database_error' ) {
		global $wpdb;
		if ( $wpdb->last_error ) {
			return $this->respond_error( $context, $wpdb->last_error, 500 );
		}
		return null;
	}
}
