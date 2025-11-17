<?php
/**
 * Plugin Name: Tanzanite FAQ Content Manager
 * Description: Manage FAQ content with 34 languages support and automatic JSON generation for Nuxt frontend
 * Version: 1.0.0
 * Author: Tanzanite Theme Team
 * License: GPL-2.0-or-later
 */

defined('ABSPATH') || exit;

// Plugin constants
define('TANZANITE_FAQ_VERSION', '1.0.0');
define('TANZANITE_FAQ_PLUGIN_DIR', plugin_dir_path(__FILE__));
define('TANZANITE_FAQ_PLUGIN_URL', plugin_dir_url(__FILE__));

// Autoload classes
spl_autoload_register(function ($class) {
    $prefix = 'TZ_FAQ_';
    $base_dir = TANZANITE_FAQ_PLUGIN_DIR . 'includes/';
    
    $len = strlen($prefix);
    if (strncmp($prefix, $class, $len) !== 0) {
        return;
    }
    
    $relative_class = substr($class, $len);
    $file = $base_dir . 'class-faq-' . str_replace('_', '-', strtolower($relative_class)) . '.php';
    
    if (file_exists($file)) {
        require $file;
    } elseif (defined('WP_DEBUG') && WP_DEBUG) {
        error_log("TZ_FAQ Autoloader: File not found for class {$class} at {$file}");
    }
});

/**
 * Main plugin class
 */
class TZ_FAQ_Content_Plugin
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
        // Initialize plugin immediately
        $this->init();
    }
    
    /**
     * Plugin activation
     */
    public function activate()
    {
        $this->create_tables();
        $this->create_upload_directory();
        flush_rewrite_rules();
    }
    
    /**
     * Create database tables
     */
    private function create_tables()
    {
        global $wpdb;
        $charset_collate = $wpdb->get_charset_collate();
        
        // Main FAQ table
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        $sql_faq = "CREATE TABLE IF NOT EXISTS {$table_faq} (
            id bigint(20) NOT NULL AUTO_INCREMENT,
            category varchar(100) NOT NULL DEFAULT 'general',
            order_num int(11) NOT NULL DEFAULT 0,
            created_at datetime DEFAULT CURRENT_TIMESTAMP,
            updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
            PRIMARY KEY (id),
            KEY category (category),
            KEY order_num (order_num)
        ) {$charset_collate};";
        
        // i18n table for multilingual content
        $table_i18n = $wpdb->prefix . 'mytheme_faq_i18n';
        $sql_i18n = "CREATE TABLE IF NOT EXISTS {$table_i18n} (
            faq_id bigint(20) NOT NULL,
            locale varchar(10) NOT NULL,
            question text NOT NULL,
            answer longtext NOT NULL,
            PRIMARY KEY (faq_id, locale),
            KEY locale (locale)
        ) {$charset_collate};";
        
        require_once(ABSPATH . 'wp-admin/includes/upgrade.php');
        dbDelta($sql_faq);
        dbDelta($sql_i18n);
    }
    
    /**
     * Create upload directory for JSON files
     */
    private function create_upload_directory()
    {
        $upload_dir = wp_upload_dir();
        $faq_dir = $upload_dir['basedir'] . '/faq';
        
        if (!file_exists($faq_dir)) {
            wp_mkdir_p($faq_dir);
        }
        
        // Create .htaccess for CORS
        $htaccess_file = $faq_dir . '/.htaccess';
        if (!file_exists($htaccess_file)) {
            $htaccess_content = "<IfModule mod_headers.c>\n";
            $htaccess_content .= "    Header set Access-Control-Allow-Origin \"*\"\n";
            $htaccess_content .= "    Header set Cache-Control \"public, max-age=3600\"\n";
            $htaccess_content .= "</IfModule>\n";
            file_put_contents($htaccess_file, $htaccess_content);
        }
    }
    
    /**
     * Initialize plugin components
     */
    public function init()
    {
        if (defined('WP_DEBUG') && WP_DEBUG) {
            error_log('TZ_FAQ: init() called');
        }
        
        // Check if classes exist before instantiating
        if (class_exists('TZ_FAQ_Database')) {
            new TZ_FAQ_Database();
            if (defined('WP_DEBUG') && WP_DEBUG) {
                error_log('TZ_FAQ: Database initialized');
            }
        }
        
        if (class_exists('TZ_FAQ_Editor')) {
            new TZ_FAQ_Editor();
            if (defined('WP_DEBUG') && WP_DEBUG) {
                error_log('TZ_FAQ: Editor initialized');
            }
        }
        
        if (class_exists('TZ_FAQ_Json_Generator')) {
            new TZ_FAQ_Json_Generator();
            if (defined('WP_DEBUG') && WP_DEBUG) {
                error_log('TZ_FAQ: Json Generator initialized');
            }
        }
    }
}

// Register activation hook
register_activation_hook(__FILE__, function() {
    $plugin = TZ_FAQ_Content_Plugin::instance();
    $plugin->activate();
});

// Initialize plugin
TZ_FAQ_Content_Plugin::instance();
