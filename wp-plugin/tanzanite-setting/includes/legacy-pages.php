<?php
/**
 * Legacy Admin Pages (Deprecated)
 */
if ( ! defined( 'ABSPATH' ) ) { exit; }
// Placeholder content
if ( ! defined( 'TANZANITE_LEGACY_MAIN_FILE' ) ) {
	define( 'TANZANITE_LEGACY_MAIN_FILE', dirname( __DIR__ ) . '/tanzanite-setting.php' );
}
if ( ! class_exists( 'Tanzanite_Settings_Plugin' ) ) {
	class Tanzanite_Settings_Plugin {
		private static $instance = null;
		public static function instance() {
			if ( null === self::$instance ) { self::$instance = new self(); }
			return self::$instance;
		}
		public function __construct() {}
        public function render_all_products() {}
        public function render_add_product() {}
        public function render_orders_list() {}
        public function render_orders_bulk() {}
        public function render_attributes() {}
        public function render_reviews() {}
        public function render_payment_method() {}
        public function render_tax_rates() {}
        public function render_shipping_templates() {}
        public function render_carriers() {}
        public function render_member_profiles() {}
        public function render_rewards() {}
        public function render_loyalty_settings() {}
        public function render_audit_logs() {}
        public function render_seo_page() {}
        public function render_markdown_templates_page() {}
        public function enqueue_admin_assets() {}
        public function maybe_upgrade_schema() {}
	}
}
if ( ! function_exists( 'tanzanite_settings_sanitize_tier_pricing' ) ) {
	function tanzanite_settings_sanitize_tier_pricing( $value, bool $from_request = false ) {
		if ( class_exists( 'Tanzanite_Plugin' ) ) {
			return Tanzanite_Plugin::sanitize_tier_pricing( $value, $from_request );
		}
		return [];
	}
}
