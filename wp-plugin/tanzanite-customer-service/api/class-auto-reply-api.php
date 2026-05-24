<?php
/**
 * Auto reply REST API.
 *
 * @package Tanzanite_Customer_Service
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class TZ_CS_Auto_Reply_API {

	public static function register_routes(): void {
		register_rest_route(
			'tanzanite/v1',
			'/auto-reply/welcome',
			array(
				'methods'             => 'GET',
				'callback'            => array( __CLASS__, 'get_welcome_message' ),
				'permission_callback' => '__return_true',
			)
		);

		register_rest_route(
			'tanzanite/v1',
			'/auto-reply/match',
			array(
				'methods'             => 'POST',
				'callback'            => array( __CLASS__, 'match_keyword' ),
				'permission_callback' => '__return_true',
			)
		);
	}

	public static function get_welcome_message( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$conversation_id = sanitize_text_field( (string) $request->get_param( 'conversation_id' ) );
		if ( '' === $conversation_id ) {
			return new WP_REST_Response(
				array(
					'success' => false,
					'message' => 'conversation_id is required.',
				),
				400
			);
		}

		$replies_table       = $wpdb->prefix . 'tz_cs_auto_replies';
		$conversations_table = $wpdb->prefix . 'tz_cs_conversations';
		$messages_table      = $wpdb->prefix . 'tz_cs_messages';

		$rule = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT * FROM {$replies_table} WHERE type = %s AND is_active = 1 ORDER BY created_at DESC LIMIT 1",
				'welcome'
			)
		);

		if ( ! $rule ) {
			return new WP_REST_Response(
				array(
					'success' => true,
					'data'    => array(
						'message'      => '',
						'already_sent' => false,
					),
				),
				200
			);
		}

		$conversation = $wpdb->get_row(
			$wpdb->prepare(
				"SELECT id, auto_reply_count, last_welcome_sent FROM {$conversations_table} WHERE id = %s LIMIT 1",
				$conversation_id
			)
		);

		if ( $conversation && ! empty( $conversation->last_welcome_sent ) ) {
			$last_sent = strtotime( (string) $conversation->last_welcome_sent );
			if ( $last_sent && ( time() - $last_sent ) < DAY_IN_SECONDS ) {
				return new WP_REST_Response(
					array(
						'success' => true,
						'data'    => array(
							'message'      => (string) $rule->reply_message,
							'already_sent' => true,
						),
					),
					200
				);
			}
		}

		$now = current_time( 'mysql' );

		if ( $conversation ) {
			$wpdb->update(
				$conversations_table,
				array(
					'last_message'      => (string) $rule->reply_message,
					'last_message_time' => $now,
					'status'            => 'active',
					'has_auto_reply'    => 1,
					'auto_reply_count'  => (int) $conversation->auto_reply_count + 1,
					'last_welcome_sent' => $now,
					'updated_at'        => $now,
				),
				array( 'id' => $conversation_id ),
				array( '%s', '%s', '%s', '%d', '%d', '%s', '%s' ),
				array( '%s' )
			);
		} else {
			$wpdb->insert(
				$conversations_table,
				array(
					'id'                => $conversation_id,
					'last_message'      => (string) $rule->reply_message,
					'last_message_time' => $now,
					'status'            => 'active',
					'has_auto_reply'    => 1,
					'auto_reply_count'  => 1,
					'last_welcome_sent' => $now,
					'created_at'        => $now,
					'updated_at'        => $now,
				),
				array( '%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s', '%s' )
			);
		}

		self::insert_auto_reply_message( $messages_table, $conversation_id, (string) $rule->reply_message, $now );

		return new WP_REST_Response(
			array(
				'success' => true,
				'data'    => array(
					'message'      => (string) $rule->reply_message,
					'already_sent' => false,
				),
			),
			200
		);
	}

	public static function match_keyword( WP_REST_Request $request ): WP_REST_Response {
		global $wpdb;

		$message         = trim( sanitize_textarea_field( (string) $request->get_param( 'message' ) ) );
		$conversation_id = sanitize_text_field( (string) $request->get_param( 'conversation_id' ) );

		if ( '' === $message || '' === $conversation_id ) {
			return new WP_REST_Response(
				array(
					'success' => false,
					'message' => 'message and conversation_id are required.',
				),
				400
			);
		}

		$replies_table       = $wpdb->prefix . 'tz_cs_auto_replies';
		$conversations_table = $wpdb->prefix . 'tz_cs_conversations';
		$messages_table      = $wpdb->prefix . 'tz_cs_messages';

		$rules = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM {$replies_table} WHERE type = %s AND is_active = 1 ORDER BY priority DESC, created_at DESC",
				'keyword'
			)
		);

		$matched_rule = null;
		foreach ( $rules as $rule ) {
			$keyword = trim( (string) $rule->trigger_keyword );
			if ( '' === $keyword ) {
				continue;
			}

			$match_type = (string) $rule->match_type;
			if ( 'contains' === $match_type ) {
				$is_match = false !== stripos( $message, $keyword );
			} else {
				$is_match = 0 === strcasecmp( $message, $keyword );
			}

			if ( $is_match ) {
				$matched_rule = $rule;
				break;
			}
		}

		if ( ! $matched_rule ) {
			return new WP_REST_Response(
				array(
					'success' => true,
					'data'    => array(
						'reply' => '',
					),
				),
				200
			);
		}

		$now   = current_time( 'mysql' );
		$reply = (string) $matched_rule->reply_message;

		$wpdb->query(
			$wpdb->prepare(
				"UPDATE {$conversations_table}
				SET last_message = %s,
					last_message_time = %s,
					has_auto_reply = 1,
					auto_reply_count = auto_reply_count + 1,
					updated_at = %s
				WHERE id = %s",
				$reply,
				$now,
				$now,
				$conversation_id
			)
		);

		self::insert_auto_reply_message( $messages_table, $conversation_id, $reply, $now );

		return new WP_REST_Response(
			array(
				'success' => true,
				'data'    => array(
					'reply'   => $reply,
					'rule_id' => (int) $matched_rule->id,
				),
			),
			200
		);
	}

	private static function insert_auto_reply_message( string $table, string $conversation_id, string $message, string $created_at ): void {
		global $wpdb;

		$wpdb->insert(
			$table,
			array(
				'conversation_id' => $conversation_id,
				'sender_type'     => 'auto',
				'sender_id'       => 0,
				'sender_name'     => 'Auto Reply',
				'sender_email'    => '',
				'agent_id'        => '',
				'message_type'    => 'text',
				'message'         => $message,
				'is_read'         => 0,
				'created_at'      => $created_at,
			),
			array( '%s', '%s', '%d', '%s', '%s', '%s', '%s', '%s', '%d', '%s' )
		);
	}
}
