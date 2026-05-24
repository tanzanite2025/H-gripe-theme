<?php
/**
 * Auth REST API Controller
 *
 * Keeps website session APIs out of the chat/customer-service module.
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Tanzanite_REST_Auth_Controller extends Tanzanite_REST_Controller {

	/**
	 * REST API base path.
	 *
	 * @var string
	 */
	protected $rest_base = 'auth';

	public function register_routes() {
		$this->register_auth_routes( $this->rest_base );

		// Backward compatibility for the Nuxt site while callers migrate.
		$this->register_auth_routes( 'chat' );
	}

	private function register_auth_routes( $base ) {
		$public_routes = array(
			'login'        => 'login',
			'register'     => 'register_user',
			'google-login' => 'google_login',
		);

		foreach ( $public_routes as $path => $callback ) {
			register_rest_route(
				$this->namespace,
				'/' . $base . '/' . $path,
				array(
					array(
						'methods'             => WP_REST_Server::CREATABLE,
						'callback'            => array( $this, $callback ),
						'permission_callback' => '__return_true',
					),
				)
			);
		}

		register_rest_route(
			$this->namespace,
			'/' . $base . '/logout',
			array(
				array(
					'methods'             => WP_REST_Server::CREATABLE,
					'callback'            => array( $this, 'logout' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);

		register_rest_route(
			$this->namespace,
			'/' . $base . '/me',
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_current_user' ),
					'permission_callback' => 'is_user_logged_in',
				),
			)
		);
	}

	public function login( $request ) {
		$username = sanitize_text_field( $request->get_param( 'username' ) );
		$password = (string) $request->get_param( 'password' );
		$remember = filter_var( $request->get_param( 'remember' ), FILTER_VALIDATE_BOOLEAN );

		if ( '' === $username || '' === $password ) {
			return $this->api_error( 'missing_credentials', __( 'Please enter username and password.', 'tanzanite-settings' ), 400 );
		}

		$user = wp_authenticate( $username, $password );
		if ( is_wp_error( $user ) ) {
			return $this->api_error( 'invalid_credentials', __( 'Invalid username or password.', 'tanzanite-settings' ), 401 );
		}

		wp_set_current_user( $user->ID );
		wp_set_auth_cookie( $user->ID, $remember );
		do_action( 'wp_login', $user->user_login, $user );

		update_user_meta( $user->ID, 'last_activity', time() );

		return $this->api_success(
			array(
				'user' => $this->format_user( $user ),
			)
		);
	}

	public function logout( $request ) {
		wp_logout();

		return $this->api_success(
			array(
				'success' => true,
			)
		);
	}

	public function register_user( $request ) {
		$username = sanitize_user( $request->get_param( 'username' ) );
		$email    = sanitize_email( $request->get_param( 'email' ) );
		$password = (string) $request->get_param( 'password' );
		$profile  = $request->get_param( 'profile' );

		if ( '' === $username || '' === $email || '' === $password ) {
			return $this->api_error( 'missing_fields', __( 'Please enter username, email and password.', 'tanzanite-settings' ), 400 );
		}

		if ( ! is_email( $email ) ) {
			return $this->api_error( 'invalid_email', __( 'Invalid email address.', 'tanzanite-settings' ), 400 );
		}

		if ( username_exists( $username ) ) {
			return $this->api_error( 'username_exists', __( 'Username already exists.', 'tanzanite-settings' ), 400 );
		}

		if ( email_exists( $email ) ) {
			return $this->api_error( 'email_exists', __( 'Email already exists.', 'tanzanite-settings' ), 400 );
		}

		$user_id = wp_create_user( $username, $password, $email );
		if ( is_wp_error( $user_id ) ) {
			return $this->api_error( 'registration_failed', $user_id->get_error_message(), 500 );
		}

		if ( is_array( $profile ) ) {
			$this->save_member_profile( $user_id, $profile );
		}

		if ( is_array( $profile ) && ! empty( $profile['fullName'] ) ) {
			wp_update_user(
				array(
					'ID'           => $user_id,
					'display_name' => sanitize_text_field( $profile['fullName'] ),
				)
			);
		}

		wp_set_current_user( $user_id );
		wp_set_auth_cookie( $user_id, true );
		update_user_meta( $user_id, 'registered_via', 'site' );
		update_user_meta( $user_id, 'last_activity', time() );

		$user = get_userdata( $user_id );

		return $this->api_success(
			array(
				'user' => $this->format_user( $user ),
			)
		);
	}

	public function google_login( $request ) {
		$id_token = (string) $request->get_param( 'id_token' );

		if ( '' === $id_token ) {
			return $this->api_error( 'missing_token', __( 'Missing Google ID token.', 'tanzanite-settings' ), 400 );
		}

		$google_user = $this->verify_google_id_token( $id_token );
		if ( is_wp_error( $google_user ) ) {
			return $this->api_error( $google_user->get_error_code(), $google_user->get_error_message(), 401 );
		}

		$google_email = sanitize_email( $google_user['email'] );
		$google_name  = sanitize_text_field( $google_user['name'] ?? '' );
		$google_sub   = sanitize_text_field( $google_user['sub'] );

		if ( '' === $google_email ) {
			return $this->api_error( 'invalid_email', __( 'Google account has no email address.', 'tanzanite-settings' ), 400 );
		}

		$is_new_user = false;
		$user        = get_user_by( 'email', $google_email );

		if ( ! $user ) {
			$username = $this->generate_unique_username( $google_email, $google_name );
			$password = wp_generate_password( 24, true, true );
			$user_id  = wp_create_user( $username, $password, $google_email );

			if ( is_wp_error( $user_id ) ) {
				return $this->api_error( 'user_creation_failed', $user_id->get_error_message(), 500 );
			}

			wp_update_user(
				array(
					'ID'           => $user_id,
					'display_name' => $google_name ?: $username,
				)
			);

			update_user_meta( $user_id, 'google_user_id', $google_sub );
			update_user_meta( $user_id, 'registered_via', 'google' );
			$user        = get_userdata( $user_id );
			$is_new_user = true;
		} else {
			$existing_google_id = get_user_meta( $user->ID, 'google_user_id', true );
			if ( empty( $existing_google_id ) ) {
				update_user_meta( $user->ID, 'google_user_id', $google_sub );
			}
		}

		wp_set_current_user( $user->ID );
		wp_set_auth_cookie( $user->ID, true );
		do_action( 'wp_login', $user->user_login, $user );

		update_user_meta( $user->ID, 'last_activity', time() );
		update_user_meta( $user->ID, 'last_login_via', 'google' );

		return $this->api_success(
			array(
				'user'        => $this->format_user( $user ),
				'is_new_user' => $is_new_user,
			)
		);
	}

	public function get_current_user( $request ) {
		$user = wp_get_current_user();

		if ( ! $user || ! $user->ID ) {
			return $this->api_error( 'not_logged_in', __( 'Not logged in.', 'tanzanite-settings' ), 401 );
		}

		update_user_meta( $user->ID, 'last_activity', time() );

		return $this->api_success(
			array(
				'user' => $this->format_user( $user ),
			)
		);
	}

	private function api_success( array $data, $status = 200 ) {
		return $this->respond_success(
			array(
				'ok'   => true,
				'data' => $data,
			),
			$status
		);
	}

	private function api_error( $code, $message, $status = 400 ) {
		return new WP_REST_Response(
			array(
				'ok'      => false,
				'code'    => $code,
				'message' => $message,
			),
			$status
		);
	}

	private function format_user( $user ) {
		$agent_id = $this->get_agent_id_by_user( $user->ID );

		return array(
			'id'           => (int) $user->ID,
			'username'     => $user->user_login,
			'email'        => $user->user_email,
			'display_name' => $user->display_name,
			'avatar'       => get_avatar_url( $user->ID ),
			'roles'        => array_values( (array) $user->roles ),
			'is_agent'     => ! empty( $agent_id ),
			'agent_id'     => $agent_id,
			'profile'      => $this->get_member_profile( $user->ID ),
			'loyalty'      => $this->get_loyalty_summary( $user->ID ),
		);
	}

	private function save_member_profile( $user_id, array $profile ) {
		global $wpdb;

		$table = $wpdb->prefix . 'tanz_member_profiles';
		if ( ! $this->table_exists( $table ) ) {
			return;
		}

		$data = array(
			'user_id'         => $user_id,
			'full_name'       => isset( $profile['fullName'] ) ? sanitize_text_field( $profile['fullName'] ) : '',
			'phone'           => isset( $profile['phone'] ) ? sanitize_text_field( $profile['phone'] ) : '',
			'country'         => isset( $profile['country'] ) ? sanitize_text_field( $profile['country'] ) : '',
			'brand'           => isset( $profile['company'] ) ? sanitize_text_field( $profile['company'] ) : '',
			'marketing_optin' => ! empty( $profile['marketingOptIn'] ) ? 1 : 0,
			'notes'           => isset( $profile['notes'] ) ? sanitize_textarea_field( $profile['notes'] ) : '',
			'created_at'      => current_time( 'mysql' ),
		);

		$existing_id = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT id FROM {$table} WHERE user_id = %d LIMIT 1",
				$user_id
			)
		);

		if ( $existing_id ) {
			unset( $data['created_at'] );
			$wpdb->update( $table, $data, array( 'id' => $existing_id ) );
		} else {
			$wpdb->insert( $table, $data );
		}
	}

	private function get_member_profile( $user_id ) {
		global $wpdb;

		$table = $wpdb->prefix . 'tanz_member_profiles';
		if ( ! $this->table_exists( $table ) ) {
			return null;
		}

		$profile = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT full_name, phone, country, brand, marketing_optin, notes FROM {$table} WHERE user_id = %d LIMIT 1",
				$user_id
			),
			ARRAY_A
		);

		if ( ! $profile ) {
			return null;
		}

		return array(
			'fullName'       => $profile['full_name'] ?? '',
			'phone'          => $profile['phone'] ?? '',
			'country'        => $profile['country'] ?? '',
			'company'        => $profile['brand'] ?? '',
			'marketingOptIn' => ! empty( $profile['marketing_optin'] ),
			'notes'          => $profile['notes'] ?? '',
		);
	}

	private function get_loyalty_summary( $user_id ) {
		$points = absint( get_user_meta( $user_id, 'loyalty_points', true ) );
		$tiers  = $this->get_loyalty_tiers();
		$level  = sanitize_text_field( (string) get_user_meta( $user_id, 'membership_level', true ) );

		foreach ( $tiers as $tier ) {
			$min = isset( $tier['min'] ) ? (int) $tier['min'] : 0;
			$max = isset( $tier['max'] ) ? (int) $tier['max'] : -1;

			if ( $points >= $min && ( -1 === $max || $points <= $max ) ) {
				$level = $tier['key'];
				break;
			}
		}

		return array(
			'points'         => $points,
			'level'          => $level,
			'tiers'          => $tiers,
			'top_tier_image' => '',
		);
	}

	private function get_loyalty_tiers() {
		$config = get_option( 'tanzanite_loyalty_config', '' );
		if ( is_string( $config ) ) {
			$config = json_decode( $config, true );
		}

		if ( ! is_array( $config ) || empty( $config['tiers'] ) || ! is_array( $config['tiers'] ) ) {
			return array();
		}

		$tiers = array();
		foreach ( $config['tiers'] as $key => $tier ) {
			if ( ! is_array( $tier ) ) {
				continue;
			}

			$tier_key = isset( $tier['key'] ) ? sanitize_key( $tier['key'] ) : sanitize_key( is_string( $key ) ? $key : ( $tier['name'] ?? '' ) );
			if ( '' === $tier_key ) {
				continue;
			}

			$tiers[] = array(
				'key'      => $tier_key,
				'name'     => sanitize_text_field( $tier['name'] ?? $tier['label'] ?? ucfirst( $tier_key ) ),
				'label'    => sanitize_text_field( $tier['label'] ?? $tier['name'] ?? ucfirst( $tier_key ) ),
				'min'      => (int) ( $tier['min'] ?? 0 ),
				'max'      => array_key_exists( 'max', $tier ) && null !== $tier['max'] ? (int) $tier['max'] : -1,
				'discount' => (float) ( $tier['discount'] ?? 0 ),
				'redeem'   => is_array( $tier['redeem'] ?? null ) ? $tier['redeem'] : array(),
			);
		}

		return $tiers;
	}

	private function get_agent_id_by_user( $user_id ) {
		global $wpdb;

		$table = $wpdb->prefix . 'tz_cs_agents';
		if ( ! $this->table_exists( $table ) ) {
			return null;
		}

		$agent_id = $wpdb->get_var(
			$wpdb->prepare(
				"SELECT agent_id FROM {$table} WHERE wp_user_id = %d AND status = 'active' LIMIT 1",
				$user_id
			)
		);

		return $agent_id ?: null;
	}

	private function table_exists( $table ) {
		global $wpdb;

		return $wpdb->get_var( $wpdb->prepare( 'SHOW TABLES LIKE %s', $table ) ) === $table;
	}

	private function verify_google_id_token( $id_token ) {
		$url      = 'https://oauth2.googleapis.com/tokeninfo?id_token=' . urlencode( $id_token );
		$response = wp_remote_get(
			$url,
			array(
				'timeout'   => 10,
				'sslverify' => true,
			)
		);

		if ( is_wp_error( $response ) ) {
			return new WP_Error( 'google_api_error', $response->get_error_message() );
		}

		$status_code = wp_remote_retrieve_response_code( $response );
		$body        = wp_remote_retrieve_body( $response );
		$data        = json_decode( $body, true );

		if ( 200 !== $status_code || empty( $data ) ) {
			return new WP_Error( 'invalid_token', __( 'Invalid or expired Google ID token.', 'tanzanite-settings' ) );
		}

		$expected_client_id = get_option( 'tanzanite_google_client_id', '' );
		if ( ! empty( $expected_client_id ) && isset( $data['aud'] ) && $data['aud'] !== $expected_client_id ) {
			return new WP_Error( 'invalid_audience', __( 'Google token audience does not match.', 'tanzanite-settings' ) );
		}

		if ( empty( $data['email_verified'] ) || 'true' !== (string) $data['email_verified'] ) {
			return new WP_Error( 'email_not_verified', __( 'Google account email is not verified.', 'tanzanite-settings' ) );
		}

		return array(
			'sub'     => $data['sub'] ?? '',
			'email'   => $data['email'] ?? '',
			'name'    => $data['name'] ?? '',
			'picture' => $data['picture'] ?? '',
		);
	}

	private function generate_unique_username( $email, $name = '' ) {
		$base_username = sanitize_user( explode( '@', $email )[0], true );

		if ( strlen( $base_username ) < 3 && '' !== $name ) {
			$base_username = sanitize_user( str_replace( ' ', '', $name ), true );
		}

		if ( strlen( $base_username ) < 3 ) {
			$base_username = 'user';
		}

		$username = $base_username;
		$suffix   = 1;

		while ( username_exists( $username ) ) {
			$username = $base_username . $suffix;
			$suffix++;

			if ( $suffix > 9999 ) {
				$username = 'user_' . wp_generate_password( 8, false );
				break;
			}
		}

		return $username;
	}
}
