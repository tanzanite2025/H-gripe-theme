<?php
if (!defined('ABSPATH')) { exit; }

final class URLLink_Plugin {
    private static $instance = null;

    public static function instance() : self {
        if (!self::$instance) {
            self::$instance = new self();
        }
        return self::$instance;
    }

    private function __construct() {
        add_action('init', [$this, 'init']);
        add_action('admin_init', [$this, 'admin_init']);
        add_action('plugins_loaded', [$this, 'i18n']);
    }

    public static function activate() : void {
        // Ensure rewrite rules are rebuilt
        flush_rewrite_rules();
    }

    public static function deactivate() : void {
        flush_rewrite_rules();
    }

    public function i18n() : void {
        load_plugin_textdomain('urllink', false, dirname(plugin_basename(URLLINK_DIR . 'urllink.php')) . '/languages');
    }

    public function init() : void {
        // settings defaults
        $defaults = [
            'enable_sku_in_url' => 0,
        ];
        foreach ($defaults as $k => $v) {
            if (get_option('urllink_' . $k, null) === null) {
                add_option('urllink_' . $k, $v);
            }
        }

        // 注册层级型 taxonomy：urllink_dir，用于构建“目录树”
        $labels = [
            'name' => 'URL 目录',
            'singular_name' => 'URL 目录',
            'search_items' => '搜索目录',
            'all_items' => '所有目录',
            'edit_item' => '编辑目录',
            'update_item' => '更新目录',
            'add_new_item' => '新增目录',
            'new_item_name' => '目录名称',
            'menu_name' => 'URL 目录',
        ];
        register_taxonomy('urllink_dir', ['post','page','product'], [
            'hierarchical' => true,
            'labels' => $labels,
            'show_ui' => true,
            'show_admin_column' => true,
            'query_var' => true,
            'rewrite' => false, // 仅做分组，不参与 WP 自身 rewrite
            'show_in_rest' => true,
        ]);

        // term meta: path_slug
        register_term_meta('urllink_dir', 'path_slug', [
            'type' => 'string',
            'single' => true,
            'show_in_rest' => true,
            'sanitize_callback' => function($v){ return sanitize_title_with_dashes((string)$v); }
        ]);

        // Admin fields for term meta
        add_action('urllink_dir_add_form_fields', function(){
            echo '<div class="form-field"><label for="path_slug">路径别名 (path_slug)</label>';
            echo '<input type="text" name="path_slug" id="path_slug" value="" />';
            echo '<p class="description">将作为该目录的 URL 前缀，例如 products、guides/fr。</p></div>';
        });
        add_action('urllink_dir_edit_form_fields', function($term){
            $val = get_term_meta($term->term_id, 'path_slug', true);
            echo '<tr class="form-field"><th scope="row"><label for="path_slug">路径别名 (path_slug)</label></th>';
            echo '<td><input name="path_slug" id="path_slug" type="text" value="' . esc_attr((string)$val) . '" class="regular-text" />';
            echo '<p class="description">将作为该目录的 URL 前缀，例如 products、guides/fr。</p></td></tr>';
        });
        add_action('created_urllink_dir', function($term_id){
            if (isset($_POST['path_slug'])) { update_term_meta($term_id, 'path_slug', sanitize_title_with_dashes(wp_unslash($_POST['path_slug']))); }
        });
        add_action('edited_urllink_dir', function($term_id){
            if (isset($_POST['path_slug'])) { update_term_meta($term_id, 'path_slug', sanitize_title_with_dashes(wp_unslash($_POST['path_slug']))); }
        });
    }

    public function admin_init() : void {
        register_setting('urllink', 'urllink_enable_sku_in_url', [
            'type' => 'boolean',
            'sanitize_callback' => function($v){ return $v ? 1 : 0; },
            'default' => 0,
        ]);
    }

    /**
     * Rebuild internal maps from all posts with custom path
     */
    public static function rebuild_maps() : array {
        $path_map = [];
        $post_map = [];
        $q = new WP_Query([
            'post_type' => ['post','page','product'],
            'posts_per_page' => -1,
            'post_status' => 'any',
            'fields' => 'ids'
        ]);
        foreach ($q->posts as $pid) {
            $raw = get_post_meta($pid, '_urllink_path', true);
            if ($raw === '' || $raw === null) continue;
            $norm = urllink_normalize_path($raw);
            if ($norm === '') continue;
            $path_map[$norm] = intval($pid);
            $post_map[intval($pid)] = $norm;
        }
        update_option('urllink_path_map', $path_map, false);
        update_option('urllink_post_map', $post_map, false);
        flush_rewrite_rules(false);
        return [$path_map, $post_map];
    }
}
