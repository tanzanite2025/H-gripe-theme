<?php
if (!defined('ABSPATH')) { exit; }

// Register custom query var for catch-all path
add_filter('query_vars', function($vars){
    $vars[] = 'urllink_path';
    return $vars;
});

// Catch-all rule placed at top; we will early-map only if a path matches our map
add_action('init', function(){
    add_rewrite_rule('^(.+)$', 'index.php?urllink_path=$matches[1]', 'top');
}, 1);

/**
 * Build and retrieve maps
 * - path_map: [ path(without leading slash) => post_id ]
 * - post_map: [ post_id => path ]
 */
function urllink_get_maps() : array {
    $path_map = get_option('urllink_path_map', []);
    $post_map = get_option('urllink_post_map', []);
    if (!is_array($path_map)) $path_map = [];
    if (!is_array($post_map)) $post_map = [];
    return [$path_map, $post_map];
}

function urllink_update_map_for_post($post_id, $path) : void {
    $post_id = intval($post_id);
    list($path_map, $post_map) = urllink_get_maps();

    // remove old mapping
    if (isset($post_map[$post_id])) {
        $old = (string)$post_map[$post_id];
        unset($post_map[$post_id]);
        if (isset($path_map[$old]) && intval($path_map[$old]) === $post_id) {
            unset($path_map[$old]);
        }
    }

    $path = urllink_normalize_path($path);
    if ($path !== '') {
        $path_map[$path] = $post_id;
        $post_map[$post_id] = $path;
    }

    update_option('urllink_path_map', $path_map, false);
    update_option('urllink_post_map', $post_map, false);
    flush_rewrite_rules(false);
}

/**
 * Resolve placeholders within a custom path
 */
function urllink_resolve_placeholders($post, $raw_path) : string {
    $path = $raw_path;

    // {postname}
    $path = str_replace(['{postname}', '{slug}'], $post->post_name, $path);

    // {sku} for WooCommerce products
    if (strpos($path, '{sku}') !== false) {
        $sku = '';
        if (function_exists('wc_get_product')) {
            $product = wc_get_product($post->ID);
            if ($product) { $sku = (string)$product->get_sku(); }
        } else {
            $sku = (string)get_post_meta($post->ID, '_sku', true);
        }
        $path = str_replace('{sku}', sanitize_title($sku), $path);
    }

    // {field:xxx} support (ACF 或普通 meta)
    if (strpos($path, '{field:') !== false) {
        $path = preg_replace_callback('/\{field:([^}]+)\}/', function($m) use ($post){
            $key = sanitize_key($m[1]);
            $val = get_post_meta($post->ID, $key, true);
            if ($val === '' && function_exists('get_field')) { $val = get_field($key, $post->ID); }
            return sanitize_title((string)$val);
        }, $path);
    }

    // collapse duplicate slashes
    $path = preg_replace('#/{2,}#', '/', $path);
    return trim($path, '/');
}

/**
 * Make custom permalinks when available
 */
add_filter('post_type_link', function($permalink, $post){
    $custom = get_post_meta($post->ID, '_urllink_path', true);
    if (!$custom) { return $permalink; }

    $resolved = urllink_resolve_placeholders($post, (string)$custom);

    // i18n prefix compatibility (respect existing language front if present)
    $prefix = '';
    if (defined('ICL_SITEPRESS_VERSION') && function_exists('wpml_object_id_filter')) {
        // WPML: use current language code in url when needed (left to WPML filtering)
    } elseif (function_exists('pll_current_language')) {
        $lang = pll_current_language('slug');
        $def = function_exists('pll_default_language') ? pll_default_language('slug') : $lang;
        if ($lang && $def && $lang !== $def) { $prefix = '/' . $lang; }
    } else {
        // fallback: allow site admin to define prefixes
        $list = (string)get_option('urllink_lang_prefixes', ''); // e.g. fr,ja
        $langs = array_filter(array_map('trim', explode(',', $list)));
        $requri = trim(parse_url(home_url(add_query_arg([])), PHP_URL_PATH), '/');
        // keep empty
    }

    $home = rtrim(home_url(), '/');
    return $home . $prefix . '/' . $resolved . '/';
}, 10, 2);

/**
 * Intercept incoming requests and map to posts; also handle redirects from old paths or extra redirects
 */
add_action('parse_request', function(WP $wp){
    if (!isset($wp->query_vars['urllink_path'])) return;
    $path = urllink_normalize_path($wp->query_vars['urllink_path']);

    // Detect language prefix and strip for matching
    $lang_prefix = '';
    if (function_exists('pll_languages_list')) {
        $langs = (array)pll_languages_list(['fields' => 'slug']);
        foreach ($langs as $slug) {
            if ($slug && stripos($path, $slug . '/') === 0) { $lang_prefix = $slug; $path = substr($path, strlen($slug)+1); break; }
        }
    }

    list($path_map, $post_map) = urllink_get_maps();

    // Direct match
    if (isset($path_map[$path])) {
        $post_id = intval($path_map[$path]);
        if ($post_id > 0) {
            $wp->query_vars = array_merge($wp->query_vars, [
                'p' => $post_id,
                'page_id' => $post_id,
            ]);
            // Remove our var to avoid 404 template
            unset($wp->query_vars['urllink_path']);
        }
        return;
    }

    // Old paths and extra redirects
    global $wpdb;
    $maybe = $wpdb->get_col($wpdb->prepare("SELECT post_id FROM {$wpdb->postmeta} WHERE meta_key = %s", '_urllink_old_paths'));
    if ($maybe) {
        foreach ($maybe as $pid) {
            $list = (array)get_post_meta(intval($pid), '_urllink_old_paths', true);
            if (in_array($path, $list, true)) {
                $url = get_permalink(intval($pid));
                if ($url) { wp_redirect($url, 301); exit; }
            }
        }
    }

    // Extra redirects on each post
    $rows = $wpdb->get_results($wpdb->prepare("SELECT post_id, meta_value FROM {$wpdb->postmeta} WHERE meta_key=%s", '_urllink_extra_redirects'));
    foreach ((array)$rows as $row) {
        $list = preg_split('/\r?\n/', (string)$row->meta_value);
        $list = array_filter(array_map('urllink_normalize_path', $list));
        if (in_array($path, $list, true)) {
            $url = get_permalink(intval($row->post_id));
            if ($url) { wp_redirect($url, 301); exit; }
        }
    }
});
