<?php
/**
 * Plugin Name: Tanzanite Subscription
 * Description: Simple email subscription and notification system for Tanzanite, providing opt-in, confirmation, and content update notifications.
 * Author: Tanzanite
 * Version: 0.1.0
 * Text Domain: tanzanite-subscription
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit; // Exit if accessed directly.
}

if ( ! defined( 'TANZ_SUB_PLUGIN_FILE' ) ) {
    define( 'TANZ_SUB_PLUGIN_FILE', __FILE__ );
}

if ( ! defined( 'TANZ_SUB_PLUGIN_DIR' ) ) {
    define( 'TANZ_SUB_PLUGIN_DIR', plugin_dir_path( __FILE__ ) );
}

if ( ! defined( 'TANZ_SUB_PLUGIN_VERSION' ) ) {
    define( 'TANZ_SUB_PLUGIN_VERSION', '0.1.0' );
}

/**
 * Plugin activation callback.
 *
 * Creates the subscribers table if it does not already exist.
 */
function tanz_sub_activate() {
    global $wpdb;

    $table_name      = $wpdb->prefix . 'tanz_subscribers';
    $charset_collate = $wpdb->get_charset_collate();

    require_once ABSPATH . 'wp-admin/includes/upgrade.php';

    $sql = "CREATE TABLE {$table_name} (
        id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
        email VARCHAR(255) NOT NULL,
        created_at DATETIME NOT NULL,
        confirmed TINYINT(1) NOT NULL DEFAULT 0,
        unsubscribed TINYINT(1) NOT NULL DEFAULT 0,
        confirm_token VARCHAR(64) DEFAULT NULL,
        unsubscribe_token VARCHAR(64) DEFAULT NULL,
        PRIMARY KEY  (id),
        UNIQUE KEY email (email)
    ) {$charset_collate};";

    dbDelta( $sql );
}
register_activation_hook( __FILE__, 'tanz_sub_activate' );

/**
 * Plugin deactivation callback.
 *
 * Currently does nothing; table and data are preserved.
 */
function tanz_sub_deactivate() {
    // Intentionally left blank. We keep subscriber data on deactivation.
}
register_deactivation_hook( __FILE__, 'tanz_sub_deactivate' );

/**
 * Bootstrap the plugin.
 *
 * Loads admin settings and, in future iterations, the REST API and
 * notification hooks.
 */
function tanz_sub_plugins_loaded() {
    // Admin settings and pages.
    if ( is_admin() ) {
        require_once TANZ_SUB_PLUGIN_DIR . 'includes/class-tsub-admin.php';

        if ( class_exists( '\\TanzaniteSubscription\\TSUB_Admin' ) ) {
            \TanzaniteSubscription\TSUB_Admin::instance();
        }
    }

    require_once TANZ_SUB_PLUGIN_DIR . 'includes/class-tsub-rest.php';
    require_once TANZ_SUB_PLUGIN_DIR . 'includes/class-tsub-notifications.php';
    require_once TANZ_SUB_PLUGIN_DIR . 'includes/class-tsub-broadcasts.php';

    if ( class_exists( '\\TanzaniteSubscription\\TSUB_REST' ) ) {
        \TanzaniteSubscription\TSUB_REST::instance();
    }

    if ( class_exists( '\\TanzaniteSubscription\\TSUB_Notifications' ) ) {
        \TanzaniteSubscription\TSUB_Notifications::instance();
    }

    if ( class_exists( '\\TanzaniteSubscription\\TSUB_Broadcasts' ) ) {
        \TanzaniteSubscription\TSUB_Broadcasts::instance();
    }
}
add_action( 'plugins_loaded', 'tanz_sub_plugins_loaded' );

/**
 * Helper to send email using the configured From Email/Name.
 *
 * @param string $to      Recipient email.
 * @param string $subject Subject line.
 * @param string $message Plain-text message body.
 */
function tanz_sub_send_mail( $to, $subject, $message ) {
    $from_email = get_option( 'tanz_sub_from_email', '' );
    $from_name  = get_option( 'tanz_sub_from_name', '' );
    $footer_note = get_option( 'tanz_sub_footer_note', '' );

    $headers = array();

    if ( $from_email ) {
        $from_name_safe  = $from_name ? $from_name : get_bloginfo( 'name' );
        $headers[]       = 'From: ' . sprintf( '%s <%s>', $from_name_safe, $from_email );
    }

    if ( $footer_note ) {
        $message = rtrim( (string) $message ) . "\n\n" . $footer_note;
    }

    wp_mail( $to, $subject, $message, $headers );
}
