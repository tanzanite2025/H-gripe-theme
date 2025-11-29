<?php
/**
 * Plugin Name: Tanzanite Photo Gallery
 * Description: Manages rider and brand photos for the Picture warehouse feature.
 * Version: 0.1.0
 * Author: Tanzanite
 */

if (! defined('ABSPATH')) {
    exit;
}

define('TPG_PLUGIN_PATH', plugin_dir_path(__FILE__));
define('TPG_PLUGIN_URL', plugin_dir_url(__FILE__));
define('TPG_PLUGIN_VERSION', '0.1.0');

require_once TPG_PLUGIN_PATH . 'includes/class-tpg-post-type.php';
require_once TPG_PLUGIN_PATH . 'includes/class-tpg-rest.php';
require_once TPG_PLUGIN_PATH . 'includes/class-tpg-admin.php';

function tpg_init_plugin() {
    \TanzanitePhotoGallery\TPG_Post_Type::instance();
    \TanzanitePhotoGallery\TPG_REST::instance();
    \TanzanitePhotoGallery\TPG_Admin::instance();
}
add_action('plugins_loaded', 'tpg_init_plugin');

function tpg_activate_plugin() {
    // Ensure post type is registered before flushing rules.
    \TanzanitePhotoGallery\TPG_Post_Type::instance();
    flush_rewrite_rules();
}
register_activation_hook(__FILE__, 'tpg_activate_plugin');

function tpg_deactivate_plugin() {
    flush_rewrite_rules();
}
register_deactivation_hook(__FILE__, 'tpg_deactivate_plugin');
