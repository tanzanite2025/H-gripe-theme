<?php
/**
 * Suggestion feedback database helper.
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Tanzanite_Suggestion_Feedback_Database {
	const OPTION_KEY = 'tanz_feedback_suggestions_db_version';
	const VERSION    = '0.1.0';

	public static function ensure_tables() {
		self::create_tables();
	}

	public static function maybe_upgrade() {
		$current = get_option( self::OPTION_KEY, '0.0.0' );
		if ( version_compare( $current, self::VERSION, '<' ) ) {
			self::create_tables();
		}
	}

	private static function create_tables() {
		global $wpdb;

		$table_name = $wpdb->prefix . 'tanz_feedback_suggestions';
		$charset    = $wpdb->get_charset_collate();

		$sql = "CREATE TABLE {$table_name} (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			user_id BIGINT UNSIGNED NULL,
			full_name VARCHAR(120) NULL,
			email VARCHAR(190) NULL,
			country VARCHAR(80) NULL,
			order_number VARCHAR(80) NULL,
			product_category VARCHAR(60) NULL,
			request_type VARCHAR(60) NULL,
			message LONGTEXT NOT NULL,
			attachments LONGTEXT NULL,
			meta LONGTEXT NULL,
			status VARCHAR(25) NOT NULL DEFAULT 'new',
			member_level_required VARCHAR(60) NULL,
			member_level_met TINYINT(1) NOT NULL DEFAULT 0,
			eligibility_hash VARCHAR(190) NULL,
			reviewed_by BIGINT UNSIGNED NULL,
			reviewed_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY status_idx (status),
			KEY user_idx (user_id)
		) {$charset};";

		require_once ABSPATH . 'wp-admin/includes/upgrade.php';
		dbDelta( $sql );

		update_option( self::OPTION_KEY, self::VERSION );
	}
}
