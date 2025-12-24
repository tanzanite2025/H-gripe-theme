<?php
/**
 * Tube Specs Admin Page
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Tanzanite_Tube_Specs_Admin {
	public static function render_page() {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( esc_html__( '无权限访问该页面。', 'tanzanite-settings' ) );
		}

		$version = defined( 'TANZANITE_VERSION' ) ? TANZANITE_VERSION : '0.0.0';

		echo '<div class="tz-settings-wrapper tz-tube-specs">';
		echo '  <div class="tz-settings-header">';
		echo '      <h1>' . esc_html__( 'Tube Specs', 'tanzanite-settings' ) . '</h1>';
		echo '      <p>' . esc_html__( 'Manage inner tube specifications and mapping to products (category slug: inner-tube).', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '  <div class="tz-settings-section">';
		echo '      <div class="tz-section-title">' . esc_html__( 'Configuration', 'tanzanite-settings' ) . '</div>';
		echo '      <p>' . esc_html__( 'This page will allow you to define tube specs (size, valve, execution) and map them to inner-tube products, according to the TUBE-SEARCH-INNER-TUBE.md design.', 'tanzanite-settings' ) . '</p>';
		echo '      <p>' . esc_html__( 'Initial implementation focuses on database structure and REST API integration; a richer UI can be added later.', 'tanzanite-settings' ) . '</p>';
		echo '  </div>';

		echo '</div>';
	}
}
