<?php
/**
 * Tube specs database helper.
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Tanzanite_Tube_Specs_Database {
	const OPTION_KEY = 'tanz_tube_specs_db_version';
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

		$table_name = $wpdb->prefix . 'tanz_tube_specs';
		$charset    = $wpdb->get_charset_collate();

		$sql = "CREATE TABLE {$table_name} (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			product_id BIGINT UNSIGNED NOT NULL,
			size_label VARCHAR(80) NULL,
			etrto_range VARCHAR(80) NULL,
			valve_family VARCHAR(20) NULL,
			valve_angle_deg SMALLINT UNSIGNED NOT NULL DEFAULT 0,
			valve_length_mm SMALLINT UNSIGNED NULL,
			execution VARCHAR(40) NULL,
			segment VARCHAR(60) NULL,
			notes TEXT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY  (id),
			KEY product_idx (product_id),
			KEY execution_idx (execution),
			KEY valve_idx (valve_family, valve_angle_deg, valve_length_mm)
		) {$charset};";

		require_once ABSPATH . 'wp-admin/includes/upgrade.php';
		dbDelta( $sql );

		update_option( self::OPTION_KEY, self::VERSION );
	}
}
