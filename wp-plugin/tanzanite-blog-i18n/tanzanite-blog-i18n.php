<?php
/**
 * Plugin Name: Tanzanite Blog I18N
 * Description: Minimal multilingual blog plugin (language taxonomy + translation groups + REST API for Nuxt).
 * Version: 0.1.0
 * Author: Tanzanite Team
 * License: GPL-2.0-or-later
 */

defined('ABSPATH') || exit;

define('TZ_BLOG_I18N_VERSION', '0.1.0');
define('TZ_BLOG_I18N_PLUGIN_DIR', plugin_dir_path(__FILE__));
define('TZ_BLOG_I18N_PLUGIN_URL', plugin_dir_url(__FILE__));

define('TZ_BLOG_I18N_CACHE_TTL', 900);

spl_autoload_register(function ($class) {
    $prefix = 'TZ_BLOG_';
    $base_dir = TZ_BLOG_I18N_PLUGIN_DIR . 'includes/';

    $len = strlen($prefix);
    if (strncmp($prefix, $class, $len) !== 0) {
        return;
    }

    $relative_class = substr($class, $len);
    $file = $base_dir . 'class-blog-' . str_replace('_', '-', strtolower($relative_class)) . '.php';

    if (file_exists($file)) {
        require $file;
    }
});

class TZ_BLOG_I18N_Plugin
{
    private static $instance = null;

    public static function instance()
    {
        if (self::$instance === null) {
            self::$instance = new self();
        }

        return self::$instance;
    }

    private function __construct()
    {
        $this->init();
    }

    public static function activate()
    {
        if (class_exists('TZ_BLOG_Languages')) {
            TZ_BLOG_Languages::register_taxonomy();
            TZ_BLOG_Languages::ensure_terms();
        }

        if (class_exists('TZ_BLOG_Setup')) {
            TZ_BLOG_Setup::ensure_categories();
        }

        if (class_exists('TZ_BLOG_Cache')) {
            TZ_BLOG_Cache::bump_version();
        }

        flush_rewrite_rules();
    }

    private function init()
    {
        if (class_exists('TZ_BLOG_Languages')) {
            new TZ_BLOG_Languages();
        }

        if (class_exists('TZ_BLOG_Setup')) {
            new TZ_BLOG_Setup();
        }

        if (class_exists('TZ_BLOG_Cache')) {
            new TZ_BLOG_Cache();
        }

        if (class_exists('TZ_BLOG_Admin')) {
            new TZ_BLOG_Admin();
        }

        if (class_exists('TZ_BLOG_REST')) {
            new TZ_BLOG_REST();
        }
    }
}

register_activation_hook(__FILE__, array('TZ_BLOG_I18N_Plugin', 'activate'));

add_action('plugins_loaded', function () {
    TZ_BLOG_I18N_Plugin::instance();
});
