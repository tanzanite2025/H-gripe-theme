<?php
/**
 * Plugin Name: MyTheme SEO Bridge
 * Description: Multilingual SEO meta management bridging WordPress content with the Tanzanite Nuxt frontend.
 * Version: 0.1.0
 * Author: Tanzanite Theme Team
 * License: GPL-2.0-or-later
 */

defined('ABSPATH') || exit;

if (!class_exists('MyTheme_SEO_Plugin')) {
    final class MyTheme_SEO_Plugin
    {
        private const OPTION_LANGUAGES = 'mytheme_seo_languages';
        private const OPTION_SETTINGS  = 'mytheme_seo_settings';
        private const OPTION_CATEGORY  = 'mytheme_seo_category_meta';
        private const OPTION_HOMEPAGE  = 'mytheme_seo_home_meta';
        private const OPTION_404_LOG   = 'mytheme_seo_404_log';
        private const META_KEY         = '_mytheme_seo_payload';
        private const DEFAULT_LANGUAGE_SOURCE = 'nuxt-i18n/public/i18n-languages.json';

        private static ?MyTheme_SEO_Plugin $instance = null;

        public static function instance(): MyTheme_SEO_Plugin
        {
            if (self::$instance === null) {
                self::$instance = new self();
            }
            return self::$instance;
        }

        public function rest_get_product_overrides(WP_REST_Request $request): WP_REST_Response
        {
            $id = (int) $request['id'];
            if ($id <= 0) { return new WP_REST_Response(['success' => false, 'message' => 'Invalid product id'], 400); }
            $p = get_post($id);
            if (!$p || $p->post_type !== 'product') { return new WP_REST_Response(['success' => false, 'message' => 'Product not found'], 404); }
            $over = get_post_meta($id, '_mytheme_seo_product_overrides', true);
            $over = is_array($over) ? $over : [];
            return new WP_REST_Response(['success' => true, 'data' => $over], 200);
        }

        public function rest_update_product_overrides(WP_REST_Request $request): WP_REST_Response
        {
            $id = (int) $request['id'];
            if ($id <= 0) { return new WP_REST_Response(['success' => false, 'message' => 'Invalid product id'], 400); }
            $p = get_post($id);
            if (!$p || $p->post_type !== 'product') { return new WP_REST_Response(['success' => false, 'message' => 'Product not found'], 404); }

            $payload = $request->get_param('overrides');
            $over = is_array($payload) ? $payload : [];
            $clean = [];
            if (isset($over['brand'])) { $clean['brand'] = sanitize_text_field((string) $over['brand']); }
            if (isset($over['priceSource'])) {
                $ps = (string) $over['priceSource'];
                $clean['priceSource'] = in_array($ps, ['sale_or_regular','regular_only'], true) ? $ps : 'sale_or_regular';
            }
            update_post_meta($id, '_mytheme_seo_product_overrides', $clean);
            return new WP_REST_Response(['success' => true, 'data' => $clean], 200);
        }

        public function rest_resolve_product_schema(WP_REST_Request $request): WP_REST_Response
        {
            $id = (int) $request->get_param('id');
            $slug = (string) ($request->get_param('slug') ?? '');
            
            $settings = $this->get_settings();
            $enabled  = !empty($settings['schema']['product']['enabled']);
            if (!$enabled) {
                return new WP_REST_Response(['success' => true, 'enabled' => false, 'schema' => null], 200);
            }
            if ($id > 0) {
                $p = get_post($id);
                if (!$p || $p->post_type !== 'tanz_product') {
                    return new WP_REST_Response(['success' => false, 'message' => 'Product not found'], 404);
                }
                $canonical = $p->post_name;
                $schema = $this->build_product_schema_array($id, $settings);
                return new WP_REST_Response([
                    'success'        => true,
                    'enabled'        => true,
                    'id'             => $id,
                    'canonical_slug' => $canonical,
                    'schema'         => $schema,
                ], 200);
            }
            if ($slug !== '') {
                $resolved = $this->resolve_product_by_slug($slug);
                if ($resolved['id'] <= 0) {
                    return new WP_REST_Response(['success' => false, 'message' => 'Product not found'], 404);
                }
                $schema = $this->build_product_schema_array($resolved['id'], $settings);
                return new WP_REST_Response([
                    'success'        => true,
                    'enabled'        => true,
                    'id'             => $resolved['id'],
                    'canonical_slug' => $resolved['slug'],
                    'schema'         => $schema,
                ], 200);
            }
            return new WP_REST_Response(['success' => false, 'message' => 'Provide id or slug'], 400);
        }

        private function __construct()
        {
            // 不再注册独立菜单，由 Tanzanite Settings 统一管理
            // add_action('admin_menu', [$this, 'register_admin_menu']);
            add_action('admin_init', [$this, 'register_settings']);
            add_action('admin_enqueue_scripts', [$this, 'enqueue_admin_assets']);
            add_action('rest_api_init', [$this, 'register_rest_routes']);
            add_action('rest_api_init', [$this, 'register_post_meta_field']);
            add_action('template_redirect', [$this, 'capture_404']);
            // Append external sitemap URLs to WP core sitemap index
            add_filter('wp_sitemaps_index', [$this, 'filter_wp_sitemaps_index']);
        }

        public function register_rest_routes(): void
        {
            // 404 logs
            register_rest_route('mytheme/v1', '/seo/404-logs', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_404_logs'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/404-logs', [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'rest_update_404_logs'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);

            // Homepage SEO
            register_rest_route('mytheme/v1', '/seo/homepage', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_homepage_meta'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/homepage', [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'rest_update_homepage_meta'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);

            // Settings
            register_rest_route('mytheme/v1', '/seo/settings', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_settings'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/settings', [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'rest_update_settings'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/settings/public', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_settings_public'],
                'permission_callback' => '__return_true'
            ]);

            // Languages
            register_rest_route('mytheme/v1', '/seo/languages', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_languages'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/languages', [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'rest_update_languages'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/languages/import', [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'rest_import_languages'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);

            // Category meta
            register_rest_route('mytheme/v1', '/seo/category/(?P<id>\d+)', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_category_meta'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/category/(?P<id>\d+)', [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'rest_update_category_meta'],
                'permission_callback' => function () { return current_user_can('manage_options'); }
            ]);

            // Product schema endpoints
            register_rest_route('mytheme/v1', '/seo/schema/product/(?P<id>\d+)', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_product_schema'],
                'permission_callback' => '__return_true'
            ]);
            register_rest_route('mytheme/v1', '/seo/schema/product/by-slug/(?P<slug>[a-z0-9\-\._]+)', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_product_schema_by_slug'],
                'permission_callback' => '__return_true'
            ]);
            register_rest_route('mytheme/v1', '/seo/schema/product/resolve', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_resolve_product_schema'],
                'permission_callback' => '__return_true'
            ]);

            // Public product basic
            register_rest_route('mytheme/v1', '/seo/product/basic/(?P<id>\d+)', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_product_basic'],
                'permission_callback' => '__return_true'
            ]);
            register_rest_route('mytheme/v1', '/seo/product/basic/by-slug/(?P<slug>[a-z0-9\-\._]+)', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_product_basic_by_slug'],
                'permission_callback' => '__return_true'
            ]);

            // Per-product overrides
            register_rest_route('mytheme/v1', '/seo/product/overrides/(?P<id>\d+)', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_product_overrides'],
                'permission_callback' => function () { return current_user_can('manage_options') || current_user_can('edit_products'); }
            ]);
            register_rest_route('mytheme/v1', '/seo/product/overrides/(?P<id>\d+)', [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'rest_update_product_overrides'],
                'permission_callback' => function (WP_REST_Request $request) { return current_user_can('manage_options') || current_user_can('edit_post', (int) $request['id']); }
            ]);

            // SEO 审核面板
            register_rest_route('mytheme/v1', '/seo/audit', [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'rest_get_seo_audit'],
                'permission_callback' => function () {
                    return current_user_can('manage_options');
                },
                'args'                => [
                    'scope' => [
                        'description' => '加载的范围：summary 或 missing。',
                        'type'        => 'string',
                        'required'    => false,
                    ],
                    'limit' => [
                        'description' => '每个集合返回的最大条目数（1-200）。',
                        'type'        => 'integer',
                        'required'    => false,
                        'minimum'     => 1,
                        'maximum'     => 200,
                    ],
                ],
            ]);
        }

        /**
         * Admin menu entry for managing languages and defaults.
         */
        public function register_admin_menu(): void
        {
            add_menu_page(
                __('MyTheme SEO', 'mytheme-seo'),
                __('MyTheme SEO', 'mytheme-seo'),
                'manage_options',
                'mytheme-seo',
                [$this, 'render_admin_page'],
                'dashicons-chart-line'
            );
        }

        public function enqueue_admin_assets(): void
        {
            $screen = function_exists('get_current_screen') ? get_current_screen() : null;
            
            // 调试：输出当前 screen ID
            if ($screen) {
                error_log('MyTheme SEO: Current screen ID = ' . $screen->id);
            }
            
            // 检查是否在 Tanzanite Settings 的 SEO 页面
            $allowed_screens = [
                'tanzanite-settings_page_tanzanite-settings-seo',  // 旧格式
                'tanzanite_page_tanzanite-settings-seo',           // 新格式（Tanzanite 子菜单）
                'toplevel_page_mytheme-seo',                       // 独立菜单
                'toplevel_page_tanzanite-settings'                 // Tanzanite 主菜单
            ];
            
            if (!$screen || !in_array((string)$screen->id, $allowed_screens, true)) { 
                error_log('MyTheme SEO: Screen ID does not match, assets not loaded');
                return; 
            }
            
            error_log('MyTheme SEO: Loading admin assets');

            // 资源路径指向 Tanzanite Settings 的 assets/seo 目录
            $plugin_dir = dirname(dirname(__FILE__)); // tanzanite-setting 目录
            $assets_dir = $plugin_dir . '/assets/seo/';
            $assets_url = plugin_dir_url($plugin_dir . '/tanzanite-setting.php') . 'assets/seo/';
            
            error_log('MyTheme SEO: Assets dir = ' . $assets_dir);
            error_log('MyTheme SEO: Assets URL = ' . $assets_url);
            error_log('MyTheme SEO: admin.css exists = ' . (file_exists($assets_dir . 'admin.css') ? 'YES' : 'NO'));
            error_log('MyTheme SEO: admin.js exists = ' . (file_exists($assets_dir . 'admin.js') ? 'YES' : 'NO'));
            
            $css_ver = file_exists($assets_dir . 'admin.css') ? (string) filemtime($assets_dir . 'admin.css') : '0.1.3';
            $js_ver  = file_exists($assets_dir . 'admin.js') ? (string) filemtime($assets_dir . 'admin.js') : '0.1.3';

            // 先加载 WordPress 核心依赖
            wp_enqueue_script('wp-element');
            wp_enqueue_script('wp-components');
            wp_enqueue_script('wp-api-fetch');
            wp_enqueue_script('wp-i18n');
            
            wp_enqueue_style('mytheme-seo-admin', $assets_url . 'admin.css', [], $css_ver);
            wp_enqueue_script('mytheme-seo-admin', $assets_url . 'admin.js', ['wp-element', 'wp-components', 'wp-api-fetch', 'wp-i18n'], $js_ver, true);
            
            error_log('MyTheme SEO: Enqueued CSS at ' . $assets_url . 'admin.css');
            error_log('MyTheme SEO: Enqueued JS at ' . $assets_url . 'admin.js');

            wp_localize_script('mytheme-seo-admin', 'MyThemeSEO', [
                'languages' => $this->get_languages(),
                'settings'  => $this->get_settings(),
            ]);
        }

        public function render_admin_page(): void
        {
            echo '<div class="wrap"><h1>MyTheme SEO</h1>';
            
            // 调试：确认页面渲染
            echo '<div class="notice notice-info"><p>SEO 页面已加载。如果下方空白，请检查浏览器控制台（F12）是否有 JavaScript 错误。</p></div>';
            
            // SEO 设置区
            echo '<div class="mytheme-seo-card"><div class="components-card__body">';
            echo '<div id="mytheme-seo-admin-app">正在加载 SEO 设置...</div>';
            echo '</div></div>';
            echo '<div id="mytheme-seo-audit-root"></div>';
            
            // 调试：添加内联脚本来测试 JavaScript 是否执行
            echo '<script>
                console.log("MyTheme SEO: Page rendered");
                console.log("MyTheme SEO: wp.element =", typeof wp !== "undefined" && wp.element ? "loaded" : "NOT loaded");
                console.log("MyTheme SEO: MyThemeSEO =", typeof MyThemeSEO !== "undefined" ? MyThemeSEO : "NOT loaded");
                console.log("MyTheme SEO: admin-app element =", document.getElementById("mytheme-seo-admin-app"));
            </script>';
            
            echo '</div>';
        }

        /**
         * Register settings used by the admin single-page app.
         */
        public function register_settings(): void
        {
            register_setting(
                'mytheme_seo_settings',
                self::OPTION_LANGUAGES,
                [
                    'type'              => 'array',
                    'sanitize_callback' => [$this, 'sanitize_languages'],
                    'default'           => []
                ]
            );

            register_setting(
                'mytheme_seo_settings',
                self::OPTION_SETTINGS,
                [
                    'type'              => 'array',
                    'sanitize_callback' => [$this, 'sanitize_settings'],
                    'default'           => []
                ]
            );
        }

        public function sanitize_languages($value): array
        {
            if (!is_array($value)) {
                $value = is_string($value) ? wp_unslash($value) : '';
                if ($value === '') {
                    return [];
                }

                $value = preg_split('/[\r\n\t,]+/', $value);
            }

            $languages = array_values(array_unique(array_filter(array_map(static function ($item) {
                $code = sanitize_key($item);
                return $code !== '' ? $code : null;
            }, $value))));

            return apply_filters('mytheme_seo_languages', $languages);
        }

        private function sanitize_string_list($value): array
        {
            // Accept array or delimiter-separated string; return unique, trimmed, non-empty strings
            if (!is_array($value)) {
                $value = is_string($value) ? preg_split('/[\r\n\t,;]+/', wp_unslash($value)) : [];
            }
            $out = [];
            foreach ((array) $value as $it) {
                if (!is_scalar($it)) { continue; }
                $s = trim((string) $it);
                if ($s === '') { continue; }
                $out[] = $s;
            }
            $out = array_values(array_unique($out));
            return $out;
        }

        public function sanitize_settings($value): array
        {
            if (!is_array($value)) {
                $value = [];
            }

            $defaults  = $this->default_settings();
            $robots    = isset($value['robots']) && is_array($value['robots']) ? $value['robots'] : [];
            $indexnow  = isset($value['indexnow']) && is_array($value['indexnow']) ? $value['indexnow'] : ($defaults['indexnow'] ?? []);
            $templates = isset($value['templates']) && is_array($value['templates']) ? $value['templates'] : ($defaults['templates'] ?? []);
            $sitemaps  = isset($value['sitemaps']) && is_array($value['sitemaps']) ? $value['sitemaps'] : ($defaults['sitemaps'] ?? []);
            $schema    = isset($value['schema']) && is_array($value['schema']) ? $value['schema'] : ($defaults['schema'] ?? []);

            $result = $defaults;
            $result['robots']['noindex_routes']     = $this->sanitize_string_list($robots['noindex_routes'] ?? []);
            $result['robots']['noindex_components'] = $this->sanitize_string_list($robots['noindex_components'] ?? []);
            // Preserve other groups largely as-is (validated during usage)
            $result['robots']['blocked_user_agents'] = $this->sanitize_string_list($robots['blocked_user_agents'] ?? []);
            $result['indexnow']  = is_array($indexnow) ? $indexnow : [];
            $result['templates'] = is_array($templates) ? $templates : [];
            $result['sitemaps']  = is_array($sitemaps) ? $sitemaps : [];
            // sanitize external URLs list
            $ext = $sitemaps['externalUrls'] ?? [];
            $ext = $this->sanitize_string_list($ext);
            // keep only http/https absolute URLs
            $ext = array_values(array_filter($ext, static function($u){
                if (!is_string($u) || $u === '') return false;
                return (stripos($u, 'http://') === 0 || stripos($u, 'https://') === 0);
            }));
            $result['sitemaps']['externalUrls'] = $ext;
            // schema settings: shallow merge with defaults
            $result['schema']    = $defaults['schema'];
            if (is_array($schema)) {
                foreach ($schema as $key => $group) {
                    if (!isset($result['schema'][$key])) { $result['schema'][$key] = []; }
                    if (is_array($group)) {
                        $result['schema'][$key] = array_merge($result['schema'][$key], $group);
                        if ($key === 'product') {
                            // sanitize brand_i18n map
                            $brandMap = [];
                            $incoming = $group['brand_i18n'] ?? [];
                            if (is_array($incoming)) {
                                foreach ($incoming as $loc => $txt) {
                                    $locKey = sanitize_key((string) $loc);
                                    if ($locKey === '') { continue; }
                                    $brandMap[$locKey] = sanitize_text_field((string) $txt);
                                }
                            }
                            $result['schema'][$key]['brand_i18n'] = $brandMap;

                            // sanitize mappings list
                            $mappings = [];
                            $incomingMaps = $group['mappings'] ?? [];
                            if (is_array($incomingMaps)) {
                                foreach ($incomingMaps as $row) {
                                    if (!is_array($row)) { continue; }
                                    $field = sanitize_key((string) ($row['field'] ?? ''));
                                    if ($field === '') { continue; }
                                    $src = $row['source'] ?? [];
                                    $type = sanitize_key((string) ($src['type'] ?? ''));
                                    $keyName = isset($src['key']) ? sanitize_text_field((string) $src['key']) : '';
                                    $constVal = isset($src['value']) ? wp_unslash((string) $src['value']) : '';
                                    $wcField = isset($src['wc']) ? sanitize_key((string) $src['wc']) : '';
                                    $perLocale = isset($src['i18n']) && is_array($src['i18n']) ? array_map('sanitize_text_field', $src['i18n']) : [];
                                    $transforms = [];
                                    if (!empty($row['transforms']) && is_array($row['transforms'])) {
                                        foreach ($row['transforms'] as $t) {
                                            $tkey = sanitize_key((string) $t);
                                            if ($tkey !== '') { $transforms[] = $tkey; }
                                        }
                                    }
                                    $mappings[] = [
                                        'field' => $field,
                                        'source' => [ 'type' => $type, 'key' => $keyName, 'value' => $constVal, 'wc' => $wcField, 'i18n' => $perLocale ],
                                        'transforms' => $transforms,
                                    ];
                                }
                            }
                            $result['schema'][$key]['mappings'] = $mappings;
                        }
                    }
                }
            }

            return $result;
        }

        private function get_settings(): array
        {
            $stored = get_option(self::OPTION_SETTINGS, []);
            $arr = is_array($stored) ? $stored : [];
            // Always sanitize+merge with defaults on read to ensure shape
            return $this->sanitize_settings($arr);
        }

        /**
         * 获取 Tanzanite 商品数据的辅助方法
         */
        private function get_tanzanite_product_data(int $product_id): ?array
        {
            global $wpdb;
            
            $post = get_post($product_id);
            if (!$post || $post->post_type !== 'tanz_product') {
                return null;
            }

            // 获取商品 meta 数据
            $price_regular = (float) get_post_meta($product_id, 'price_regular', true);
            $price_sale = (float) get_post_meta($product_id, 'price_sale', true);
            $sku = (string) get_post_meta($product_id, 'sku', true);
            $stock_qty = (int) get_post_meta($product_id, 'stock_qty', true);
            $featured_image_id = (int) get_post_thumbnail_id($product_id);
            
            // 获取图库图片
            $gallery_ids = get_post_meta($product_id, 'gallery_ids', true);
            if (!is_array($gallery_ids)) {
                $gallery_ids = [];
            }

            // 获取评分数据
            $reviews_table = $wpdb->prefix . 'tanz_product_reviews';
            $rating_stats = $wpdb->get_row($wpdb->prepare(
                "SELECT AVG(rating) as average_rating, COUNT(*) as review_count 
                FROM {$reviews_table} 
                WHERE product_id = %d AND status = 'approved'",
                $product_id
            ));

            $average_rating = $rating_stats ? (float) $rating_stats->average_rating : 0;
            $review_count = $rating_stats ? (int) $rating_stats->review_count : 0;

            return [
                'name' => $post->post_title,
                'sku' => $sku,
                'price_regular' => $price_regular,
                'price_sale' => $price_sale,
                'stock_qty' => $stock_qty,
                'in_stock' => $stock_qty > 0,
                'image_id' => $featured_image_id,
                'gallery_ids' => $gallery_ids,
                'short_description' => $post->post_excerpt,
                'description' => $post->post_content,
                'average_rating' => $average_rating,
                'review_count' => $review_count,
            ];
        }

        private function default_settings(): array
        {
            return [
                'robots' => [
                    'noindex_routes'     => [],
                    'noindex_components' => [],
                    'blocked_user_agents'=> [],
                ],
                'indexnow' => [
                    'enabled'        => false,
                    'pushAllLocales' => true,
                    'defaultNoPrefix'=> true,
                    'tplPage'        => '/{locale}/{slug}',
                    'tplPost'        => '/{locale}/{slug}',
                    'tplProduct'     => '/{locale}/product/{slug}',
                    'tplCategory'    => '/{locale}/category/{slug}',
                    'tplProductCat'  => '/{locale}/product-category/{slug}',
                ],
                'templates' => [],
                'sitemaps'  => [
                    'enabled'       => true,
                    'splitByLocale' => true,
                    'types'         => [
                        'pages'             => true,
                        'posts'             => true,
                        'products'          => true,
                        'categories'        => true,
                        'product_categories'=> true,
                        'faq'               => true,
                    ],
                    'pingOnRebuild' => true,
                    'externalUrls'  => [],
                ],
                'schema' => [
                    'product' => [
                        'enabled'      => true,
                        'brand'        => '',
                        'brand_i18n'   => [],
                        'priceSource'  => 'sale_or_regular',
                        'availabilityInStock' => 'https://schema.org/InStock',
                        'availabilityOutOfStock' => 'https://schema.org/OutOfStock',
                        'mappings'     => [],
                    ],
                    'faq' => [
                        'enabled'           => true,
                        'organizationName'  => 'Tanzanite',
                        'organizationUrl'   => '',
                    ],
                ],
            ];
        }

        private function build_product_schema_array(int $product_id, array $settings, string $locale = ''): array
        {
    // 使用 Tanzanite 商品数据
    $product_data = $this->get_tanzanite_product_data($product_id);
    if (!$product_data) { return []; }
    // Per-product overrides
    $over = get_post_meta($product_id, '_mytheme_seo_product_overrides', true);
    $over = is_array($over) ? $over : [];

    // Mapping resolver
    $mappings = isset($settings['schema']['product']['mappings']) && is_array($settings['schema']['product']['mappings']) ? $settings['schema']['product']['mappings'] : [];
    $resolver = function(string $field) use ($mappings, $product_data, $product_id, $locale) {
        // find first mapping for field
        foreach ($mappings as $row) {
            if (($row['field'] ?? '') !== $field || !is_array($row)) { continue; }
            $src = $row['source'] ?? [];
            $type = $src['type'] ?? '';
            $val = null;
            if ($type === 'core') {
                $key = (string) ($src['key'] ?? '');
                switch ($key) {
                    case 'post_title': $val = get_post_field('post_title', $product_id); break;
                    case 'post_excerpt': $val = get_post_field('post_excerpt', $product_id); break;
                    case 'post_content': $val = get_post_field('post_content', $product_id); break;
                    case 'post_author_display_name':
                        $uid = (int) get_post_field('post_author', $product_id);
                        $u = $uid ? get_userdata($uid) : null;
                        $val = $u ? $u->display_name : '';
                        break;
                }
            } elseif ($type === 'meta') {
                $key = (string) ($src['key'] ?? '');
                if ($key !== '') { $val = get_post_meta($product_id, $key, true); }
            } elseif ($type === 'wc' || $type === 'product') {
                // 兼容旧的 'wc' 类型，现在使用 Tanzanite 商品数据
                $key = (string) ($src['wc'] ?? '');
                if ($key !== '') {
                    switch ($key) {
                        case 'name': $val = $product_data['name']; break;
                        case 'sku': $val = $product_data['sku']; break;
                        case 'price': $val = $product_data['price_sale'] ?: $product_data['price_regular']; break;
                        case 'regular_price': $val = $product_data['price_regular']; break;
                        case 'sale_price': $val = $product_data['price_sale']; break;
                        case 'stock_status': $val = $product_data['in_stock'] ? 'instock' : 'outofstock'; break;
                        case 'image_id': $val = $product_data['image_id']; break;
                        case 'gallery_ids': $val = $product_data['gallery_ids']; break;
                    }
                }
            } elseif ($type === 'const') {
                // i18n constants support
                $i18n = isset($src['i18n']) && is_array($src['i18n']) ? $src['i18n'] : [];
                if ($locale && !empty($i18n[$locale])) { $val = $i18n[$locale]; }
                if ($val === null || $val === '') { $val = (string) ($src['value'] ?? ''); }
            }
            // transforms
            $trans = is_array($row['transforms'] ?? null) ? $row['transforms'] : [];
            foreach ($trans as $t) {
                switch ($t) {
                    case 'strip_tags': $val = is_string($val) ? wp_strip_all_tags($val, true) : $val; break;
                    case 'trim': $val = is_string($val) ? trim($val) : $val; break;
                    case 'to_number': $val = is_scalar($val) ? (float) $val : $val; break;
                    case 'id_to_url': $val = $val ? wp_get_attachment_url((int) $val) : $val; break;
                    case 'first_gallery_to_url':
                        if (is_array($val) && !empty($val)) { $u = wp_get_attachment_url((int) $val[0]); if ($u) { $val = $u; } }
                        break;
                }
            }
            if ($val !== null && $val !== '') { return $val; }
        }
        return null;
    };

    // Brand selection: mapping>brand_i18n[locale]>override.brand>default brand
    $brand     = '';
    $mappedBrand = $resolver('brand');
    if ($mappedBrand !== null && $mappedBrand !== '') { $brand = (string) $mappedBrand; }
    $brand_i18n = isset($settings['schema']['product']['brand_i18n']) && is_array($settings['schema']['product']['brand_i18n']) ? $settings['schema']['product']['brand_i18n'] : [];
    if ($locale !== '') {
        $locKey = sanitize_key($locale);
        if ($locKey !== '' && !empty($brand_i18n[$locKey])) {
            $brand = (string) $brand_i18n[$locKey];
        }
    }
    if ($brand === '' && !empty($over['brand'])) {
        $brand = sanitize_text_field((string) $over['brand']);
    }
    if ($brand === '') {
        $brand = (string) ($settings['schema']['product']['brand'] ?? '');
    }
    $currency  = 'USD'; // 默认货币，可以从设置中获取
    $permalink = get_permalink($product_id) ?: home_url('/');
    $name      = (string) ($resolver('name') ?? $product_data['name']);
    $sku       = (string) ($resolver('sku') ?? $product_data['sku']);
    $images    = [];
    $image_id  = $product_data['image_id'];
    $mappedImg = $resolver('image');
    if (is_string($mappedImg) && $mappedImg !== '') { $images[] = $mappedImg; }
    if ($image_id) {
        $img_url = wp_get_attachment_url($image_id);
        if ($img_url) { $images[] = $img_url; }
    }
    $gallery_ids = $product_data['gallery_ids'];
    if (is_array($gallery_ids)) {
        foreach ($gallery_ids as $aid) {
            $u = wp_get_attachment_url($aid);
            if ($u) { $images[] = $u; }
        }
    }

    $price_src = $over['priceSource'] ?? ($settings['schema']['product']['priceSource'] ?? 'sale_or_regular');
    $price = $product_data['price_regular'];
    if ($price_src === 'sale_or_regular') {
        $sale = $product_data['price_sale'];
        if ($sale > 0) { $price = $sale; }
    }
    $price = (float) $price;

    $availability = $product_data['in_stock']
        ? ($settings['schema']['product']['availabilityInStock'] ?? 'https://schema.org/InStock')
        : ($settings['schema']['product']['availabilityOutOfStock'] ?? 'https://schema.org/OutOfStock');

    // 使用真实的评分数据
    $rating_value = (float) $product_data['average_rating'];
    $rating_count = (int) $product_data['review_count'];

    $data = [
        '@context' => 'https://schema.org/',
        '@type'    => 'Product',
        'name'     => $name,
        'brand'    => $brand !== '' ? ['@type' => 'Brand', 'name' => $brand] : null,
        'sku'      => $sku ?: null,
        'image'    => !empty($images) ? $images : null,
        'description' => null,
        'offers'   => [
            '@type'         => 'Offer',
            'priceCurrency' => $currency,
            'price'         => $price,
            'availability'  => $availability,
            'url'           => $permalink,
        ],
    ];

    // mapped description, gtin, mpn
    $descMap = $resolver('description');
    if (is_string($descMap) && $descMap !== '') { $data['description'] = $descMap; }
    $gtin = $resolver('gtin'); if ($gtin !== null && $gtin !== '') { $data['gtin'] = is_scalar($gtin) ? (string) $gtin : $gtin; }
    $mpn  = $resolver('mpn');  if ($mpn  !== null && $mpn  !== '') { $data['mpn']  = is_scalar($mpn)  ? (string) $mpn  : $mpn; }

    if ($rating_count > 0 && $rating_value > 0) {
        $data['aggregateRating'] = [
            '@type'       => 'AggregateRating',
            'ratingValue' => round($rating_value, 2),
            'reviewCount' => $rating_count,
        ];
    }

    // Remove nulls
    foreach ($data as $k => $v) {
        if ($v === null || (is_array($v) && empty($v))) {
            unset($data[$k]);
        }
    }

    return $data;
}


        private function build_basic_product_payload(int $product_id): array
        {
            $product_data = $this->get_tanzanite_product_data($product_id);
            if (!$product_data) { return []; }
            
            $image = '';
            $image_id = $product_data['image_id'];
            if ($image_id) {
                $u = wp_get_attachment_url($image_id);
                if ($u) { $image = $u; }
            }
            if ($image === '') {
                $gallery_ids = $product_data['gallery_ids'];
                if (is_array($gallery_ids) && !empty($gallery_ids)) {
                    $u = wp_get_attachment_url($gallery_ids[0]);
                    if ($u) { $image = $u; }
                }
            }
            $price = (float) ($product_data['price_sale'] ?: $product_data['price_regular']);
            $currency = 'USD'; // 默认货币
            $short = wp_strip_all_tags($product_data['short_description'] ?: '', true);
            $permalink = get_permalink($product_id) ?: home_url('/');
            return [
                'name'        => $product_data['name'],
                'sku'         => $product_data['sku'],
                'price'       => $price,
                'currency'    => $currency,
                'image'       => $image,
                'short'       => $short,
                'url'         => $permalink,
            ];
        }

        public function rest_get_product_basic(WP_REST_Request $request): WP_REST_Response
        {
            $id = (int) $request['id'];
            if ($id <= 0) { return new WP_REST_Response(['success' => false, 'message' => 'Invalid product id'], 400); }
            $p = get_post($id);
            if (!$p || $p->post_type !== 'tanz_product') { return new WP_REST_Response(['success' => false, 'message' => 'Product not found'], 404); }
            $payload = $this->build_basic_product_payload($id);
            return new WP_REST_Response(['success' => true, 'data' => $payload], 200);
        }

        public function rest_get_product_basic_by_slug(WP_REST_Request $request): WP_REST_Response
        {
            $slug = (string) $request['slug'];
            if ($slug === '') { return new WP_REST_Response(['success' => false, 'message' => 'Invalid slug'], 400); }
            $resolved = $this->resolve_product_by_slug($slug);
            if ($resolved['id'] <= 0) { return new WP_REST_Response(['success' => false, 'message' => 'Product not found'], 404); }
            $payload = $this->build_basic_product_payload($resolved['id']);
            return new WP_REST_Response(['success' => true, 'id' => $resolved['id'], 'canonical_slug' => $resolved['slug'], 'data' => $payload], 200);
        }

        private function resolve_product_by_slug(string $slug): array
        {
            $args = [
                'name'           => sanitize_title($slug),
                'post_type'      => 'tanz_product',
                'post_status'    => 'publish',
                'posts_per_page' => 1,
            ];
            $posts = get_posts($args);
            if (!empty($posts)) {
                $p = $posts[0];
                return [ 'id' => $p->ID, 'slug' => $p->post_name ];
            }
            return [ 'id' => 0, 'slug' => '' ];
        }

        public function rest_get_product_schema_by_slug(WP_REST_Request $request): WP_REST_Response
        {
            $slug = (string) $request['slug'];
            if ($slug === '') {
                return new WP_REST_Response(['success' => false, 'message' => 'Invalid slug'], 400);
            }

            $settings = $this->get_settings();
            $enabled  = !empty($settings['schema']['product']['enabled']);
            if (!$enabled) {
                return new WP_REST_Response(['success' => true, 'enabled' => false, 'schema' => null], 200);
            }

            $resolved = $this->resolve_product_by_slug($slug);
            if ($resolved['id'] <= 0) {
                return new WP_REST_Response(['success' => false, 'message' => 'Product not found'], 404);
            }

            $locale = (string) ($request->get_param('locale') ?? '');
            $schema = $this->build_product_schema_array($resolved['id'], $settings, $locale);
            return new WP_REST_Response([
                'success'        => true,
                'enabled'        => true,
                'id'             => $resolved['id'],
                'canonical_slug' => $resolved['slug'],
                'schema'         => $schema,
            ], 200);
        }

        public function rest_sitemaps_rebuild(WP_REST_Request $request): WP_REST_Response
        {
            $settings = $this->get_settings();
            $sm = $settings['sitemaps'] ?? [];
            if (empty($sm['enabled'])) {
                return new WP_REST_Response([ 'success' => false, 'message' => 'Sitemaps disabled' ], 400);
            }

            $uploads = wp_get_upload_dir();
            $base_dir = trailingslashit(ABSPATH); // try site root
            $base_url = home_url('/');

            $languages = $this->get_languages();
            if (empty($languages)) { $languages = ['']; }
            $split = !empty($sm['splitByLocale']);

            $generated = [];

            $locales_sets = $split ? $languages : [''];
            foreach ($locales_sets as $locale) {
                $bundle = $this->build_sitemaps_for_locale($locale, $settings);
                foreach ($bundle as $entry) {
                    $file_path = $base_dir . ltrim($entry['filename'], '/');
                    $xml = $entry['xml'];
                    $ok = false;
                    if (wp_mkdir_p(dirname($file_path))) {
                        $ok = (bool) file_put_contents($file_path, $xml);
                    }
                    $generated[] = [
                        'path'  => $file_path,
                        'url'   => trailingslashit($base_url) . $entry['filename'],
                        'count' => $entry['count'],
                        'ok'    => $ok,
                    ];
                }
            }

            // Build/overwrite sitemap index
            $index_items = array_filter($generated, static fn($g) => str_contains($g['path'], 'sitemap-') && str_ends_with($g['path'], '.xml'));
            $index_xml = $this->render_sitemap_index(array_map(static function($g){ return $g['url']; }, $index_items));
            $index_path = $base_dir . 'sitemap.xml';
            file_put_contents($index_path, $index_xml);

            // Auto ping if enabled
            $pinged = [];
            if (!empty($sm['pingOnRebuild'])) {
                $pinged = $this->do_sitemaps_ping(home_url('/sitemap.xml'));
            }

            return new WP_REST_Response([
                'success'   => true,
                'generated' => $generated,
                'index'     => home_url('/sitemap.xml'),
                'ping'      => $pinged,
            ]);
        }

        public function rest_sitemaps_ping(WP_REST_Request $request): WP_REST_Response
        {
            $sitemap_url = trailingslashit(home_url()) . 'sitemap.xml';
            $results = $this->do_sitemaps_ping($sitemap_url, $request->get_param('engines'));
            return new WP_REST_Response([ 'success' => true, 'results' => $results ]);
        }

        private function do_sitemaps_ping(string $sitemap_url, $engines = null): array
        {
            if (!is_array($engines)) { $engines = ['google','bing']; }
            $endpoints = [
                'google' => 'https://www.google.com/ping?sitemap=' . rawurlencode($sitemap_url),
                'bing'   => 'https://www.bing.com/ping?sitemap=' . rawurlencode($sitemap_url),
            ];
            $out = [];
            foreach ($engines as $e) {
                $url = $endpoints[$e] ?? '';
                if ($url === '') { $out[$e] = [ 'status' => 0, 'error' => 'unknown engine' ]; continue; }
                $resp = wp_remote_get($url, [ 'timeout' => 8 ]);
                if (is_wp_error($resp)) {
                    $out[$e] = [ 'status' => 0, 'error' => $resp->get_error_message() ];
                } else {
                    $out[$e] = [ 'status' => wp_remote_retrieve_response_code($resp) ];
                }
            }
            return $out;
        }

        private function build_sitemaps_for_locale(string $locale, array $settings): array
        {
            $items = [];
            $types = $settings['sitemaps']['types'] ?? [];
            $indexnow = $settings['indexnow'] ?? [];

            $urls = [];
            $add = function($url) use (&$urls) { if (is_string($url) && $url !== '') $urls[] = $url; };

            // pages
            if (!empty($types['pages'])) {
                $pages = get_posts([ 'post_type' => 'page', 'post_status' => 'publish', 'numberposts' => -1 ]);
                foreach ($pages as $p) { $add($this->format_url_from_tpl('page', $p->ID, $locale, $indexnow)); }
            }
            // posts
            if (!empty($types['posts'])) {
                $posts = get_posts([ 'post_type' => 'post', 'post_status' => 'publish', 'numberposts' => -1 ]);
                foreach ($posts as $p) { $add($this->format_url_from_tpl('post', $p->ID, $locale, $indexnow)); }
            }
            // products
            if (!empty($types['products']) && post_type_exists('product')) {
                $products = get_posts([ 'post_type' => 'product', 'post_status' => 'publish', 'numberposts' => -1 ]);
                foreach ($products as $p) { $add($this->format_url_from_tpl('product', $p->ID, $locale, $indexnow)); }
            }
            // categories
            if (!empty($types['categories'])) {
                $terms = get_terms([ 'taxonomy' => 'category', 'hide_empty' => true ]);
                if (!is_wp_error($terms)) {
                    foreach ($terms as $t) { $add($this->format_url_from_tpl('taxonomy:category', (int)$t->term_id, $locale, $indexnow)); }
                }
            }
            // product categories
            if (!empty($types['product_categories']) && taxonomy_exists('product_cat')) {
                $terms = get_terms([ 'taxonomy' => 'product_cat', 'hide_empty' => true ]);
                if (!is_wp_error($terms)) {
                    foreach ($terms as $t) { $add($this->format_url_from_tpl('taxonomy:product_cat', (int)$t->term_id, $locale, $indexnow)); }
                }
            }

            // render locale sitemap file
            $filename = $locale ? ('sitemap-' . $locale . '.xml') : 'sitemap-default.xml';
            $xml = $this->render_urlset($urls);
            $items[] = [ 'filename' => $filename, 'xml' => $xml, 'count' => count($urls) ];
            return $items;
        }

        private function render_urlset(array $urls): string
        {
            $xml = '<?xml version="1.0" encoding="UTF-8"?>' . "\n";
            $xml .= '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">' . "\n";
            foreach ($urls as $u) {
                $loc = esc_url($u);
                $xml .= '  <url><loc>' . esc_html($loc) . '</loc></url>' . "\n";
            }
            $xml .= '</urlset>';
            return $xml;
        }

        private function build_seo_audit_collection(array $post_types, bool $configured, int $limit): array
        {
            $post_types = array_filter(array_map('sanitize_key', $post_types));
            if (empty($post_types)) {
                return [
                    'count' => 0,
                    'items' => [],
                ];
            }

            $query_args = [
                'post_type'      => $post_types,
                'post_status'    => 'publish',
                'posts_per_page' => $limit,
                'orderby'        => 'modified',
                'order'          => 'DESC',
                'fields'         => 'ids',
                'no_found_rows'  => false,
            ];

            if ($configured) {
                $query_args['meta_query'] = [
                    [
                        'key'     => self::META_KEY,
                        'compare' => 'EXISTS',
                    ],
                ];
            } else {
                $query_args['meta_query'] = [
                    [
                        'key'     => self::META_KEY,
                        'compare' => 'NOT EXISTS',
                    ],
                ];
            }

            $query = new WP_Query($query_args);
            $ids   = is_array($query->posts) ? $query->posts : [];
            $items = [];

            foreach ($ids as $id) {
                $id     = (int) $id;
                $items[] = [
                    'id'       => $id,
                    'title'    => get_the_title($id),
                    'type'     => get_post_type($id),
                    'edit_url' => get_edit_post_link($id, 'link') ?: '',
                    'view_url' => get_permalink($id) ?: '',
                ];
            }

            return [
                'count' => (int) $query->found_posts,
                'items' => $items,
            ];
        }

        private function render_sitemap_index(array $sitemap_urls): string
        {
            $xml = '<?xml version="1.0" encoding="UTF-8"?>' . "\n";
            $xml .= '<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">' . "\n";
            $now = date('c');
            foreach ($sitemap_urls as $u) {
                $xml .= '  <sitemap><loc>' . esc_html($u) . '</loc><lastmod>' . esc_html($now) . '</lastmod></sitemap>' . "\n";
            }
            $xml .= '</sitemapindex>';
            return $xml;
        }

        private function format_url_from_tpl(string $type, int $id, string $locale, array $idx): string
        {
            $site = trailingslashit(home_url());
            $defaultNoPrefix = !empty($idx['defaultNoPrefix']);
            $slug = '';
            $tpl = '';
            switch ($type) {
                case 'page':
                    $slug = get_post_field('post_name', $id);
                    $tpl = $idx['tplPage'] ?? '/{locale}/{slug}';
                    break;
                case 'post':
                    $slug = get_post_field('post_name', $id);
                    $tpl = $idx['tplPost'] ?? '/{locale}/{slug}';
                    break;
                case 'product':
                    $slug = get_post_field('post_name', $id);
                    $tpl = $idx['tplProduct'] ?? '/{locale}/product/{slug}';
                    break;
                case 'taxonomy:category':
                    $term = get_term($id, 'category');
                    $slug = is_object($term) ? $term->slug : '';
                    $tpl = $idx['tplCategory'] ?? '/{locale}/category/{slug}';
                    break;
                case 'taxonomy:product_cat':
                    $term = get_term($id, 'product_cat');
                    $slug = is_object($term) ? $term->slug : '';
                    $tpl = $idx['tplProductCat'] ?? '/{locale}/product-category/{slug}';
                    break;
            }
            $path = str_replace(['{slug}','{locale}'], [$slug, $locale], $tpl);
            if ($defaultNoPrefix && $locale === ($this->get_languages()[0] ?? '')) {
                $path = preg_replace('#^/\{locale\}#', '', $path);
                $path = str_replace('/' . $locale . '/', '/', $path);
            }
            $path = preg_replace('#//+#', '/', $path);
            return rtrim($site, '/') . $path;
        }

        public function rest_get_seo(WP_REST_Request $request): WP_REST_Response
        {
            $post_id = (int) $request['id'];

            $payload = $this->prepare_post_seo_payload($post_id);

            return new WP_REST_Response([
                'id'        => $post_id,
                'payload'   => $payload,
                'languages' => $this->get_languages(),
            ]);
        }

        public function rest_update_seo(WP_REST_Request $request): WP_REST_Response
        {
            $post_id = (int) $request['id'];
            $payload = $request->get_param('payload');

            if (!is_array($payload)) {
                return new WP_REST_Response([
                    'message' => __('Invalid payload structure.', 'mytheme-seo'),
                ], 400);
            }

            $sanitized = $this->sanitize_payload($payload);
            update_post_meta($post_id, self::META_KEY, $sanitized);
            $sanitized_with_fallback = $this->apply_description_fallbacks($post_id, $sanitized);
            $sanitized_with_fallback = $this->apply_media_fallbacks($post_id, $sanitized_with_fallback);

            return new WP_REST_Response([
                'id'      => $post_id,
                'payload' => $sanitized_with_fallback,
            ]);
        }

        public function rest_get_product_seo(WP_REST_Request $request): WP_REST_Response
        {
            $product_id = (int) $request['id'];

            if ($product_id <= 0) {
                return new WP_REST_Response([
                    'message' => __('Invalid product ID.', 'mytheme-seo'),
                ], 400);
            }

            $payload = $this->prepare_post_seo_payload($product_id);

            return new WP_REST_Response([
                'id'      => $product_id,
                'payload' => $payload,
            ]);
        }

        public function rest_update_product_seo(WP_REST_Request $request): WP_REST_Response
        {
            $product_id = (int) $request['id'];
            $payload    = $request->get_param('payload');

            if ($product_id <= 0) {
                return new WP_REST_Response([
                    'message' => __('Invalid product ID.', 'mytheme-seo'),
                ], 400);
            }

            if (!is_array($payload)) {
                return new WP_REST_Response([
                    'message' => __('Invalid payload structure.', 'mytheme-seo'),
                ], 400);
            }

            $sanitized = $this->sanitize_payload($payload);
            update_post_meta($product_id, self::META_KEY, $sanitized);
            $sanitized_with_fallback = $this->prepare_post_seo_payload($product_id);

            return new WP_REST_Response([
                'id'      => $product_id,
                'payload' => $sanitized_with_fallback,
            ]);
        }

        public function rest_get_seo_audit(WP_REST_Request $request): WP_REST_Response
        {
            $scope = strtolower((string) ($request->get_param('scope') ?? ''));
            $limit = (int) ($request->get_param('limit') ?? 50);
            if ($limit <= 0) {
                $limit = 50;
            }
            if ($limit > 200) {
                $limit = 200;
            }

            if ($scope === 'missing') {
                $missing = $this->build_seo_audit_collection(['tanz_product', 'post', 'page'], false, $limit);

                return new WP_REST_Response([
                    'missing' => $missing,
                ]);
            }

            $products = $this->build_seo_audit_collection(['tanz_product'], true, $limit);
            $content  = $this->build_seo_audit_collection(['post', 'page'], true, $limit);

            return new WP_REST_Response([
                'products' => $products,
                'content'  => $content,
            ]);
        }

        public function rest_get_languages(): WP_REST_Response
        {
            return new WP_REST_Response([
                'languages' => $this->get_languages(),
            ]);
        }

        public function rest_update_languages(WP_REST_Request $request): WP_REST_Response
        {
            $languages = $request->get_param('languages');
            $sanitized = $this->sanitize_languages($languages);
            update_option(self::OPTION_LANGUAGES, $sanitized);

            return new WP_REST_Response([
                'languages' => $sanitized,
            ]);
        }

        public function rest_get_settings(): WP_REST_Response
        {
            return new WP_REST_Response([
                'settings' => $this->get_settings(),
            ]);
        }

        public function rest_update_settings(WP_REST_Request $request): WP_REST_Response
        {
            $settings  = $request->get_param('settings');
            $sanitized = $this->sanitize_settings(is_array($settings) ? $settings : []);
            update_option(self::OPTION_SETTINGS, $sanitized);

            return new WP_REST_Response([
                'settings' => $sanitized,
            ]);
        }

        public function rest_get_settings_public(): WP_REST_Response
        {
            return new WP_REST_Response([
                'settings' => $this->get_settings(),
            ]);
        }

        public function rest_get_category_meta(WP_REST_Request $request): WP_REST_Response
        {
            $category_id = (int) $request['id'];
            if ($category_id <= 0) {
                return new WP_REST_Response([
                    'message' => __('Invalid category ID.', 'mytheme-seo'),
                ], 400);
            }

            return new WP_REST_Response([
                'category' => $this->get_category_meta($category_id),
            ]);
        }

        public function rest_update_category_meta(WP_REST_Request $request): WP_REST_Response
        {
            $category_id = (int) $request['id'];
            if ($category_id <= 0) {
                return new WP_REST_Response([
                    'message' => __('Invalid category ID.', 'mytheme-seo'),
                ], 400);
            }

            $meta      = $request->get_param('meta');
            $sanitized = $this->sanitize_category_meta(is_array($meta) ? $meta : []);
            $this->set_category_meta($category_id, $sanitized);

            return new WP_REST_Response([
                'category' => $sanitized,
            ]);
        }

        public function rest_get_homepage_meta(): WP_REST_Response
        {
            return new WP_REST_Response([
                'homepage' => $this->get_homepage_meta(),
            ]);
        }

        public function rest_update_homepage_meta(WP_REST_Request $request): WP_REST_Response
        {
            $meta      = $request->get_param('meta');
            $sanitized = $this->sanitize_homepage_meta(is_array($meta) ? $meta : []);
            $this->set_homepage_meta($sanitized);

            return new WP_REST_Response([
                'homepage' => $sanitized,
            ]);
        }

        private function get_homepage_meta(): array
        {
            $stored = get_option(self::OPTION_HOMEPAGE);
            if (!is_array($stored)) {
                return $this->default_homepage_meta();
            }

            $sanitized = $this->sanitize_homepage_meta($stored);

            return wp_parse_args($sanitized, $this->default_homepage_meta());
        }

        private function set_homepage_meta(array $meta): void
        {
            update_option(self::OPTION_HOMEPAGE, $this->sanitize_homepage_meta($meta));
        }

        private function sanitize_homepage_meta(array $meta): array
        {
            return [
                'title'       => sanitize_text_field($meta['title'] ?? ''),
                'description' => sanitize_textarea_field($meta['description'] ?? ''),
                'keywords'    => sanitize_text_field($meta['keywords'] ?? ''),
            ];
        }

        private function default_homepage_meta(): array
        {
            return [
                'title'       => get_bloginfo('name'),
                'description' => get_bloginfo('description'),
                'keywords'    => '',
            ];
        }

        public function rest_import_languages(WP_REST_Request $request): WP_REST_Response
        {
            $source = $request->get_param('source');
            $payload = $this->load_languages_from_source($source);

            if (is_wp_error($payload)) {
                return new WP_REST_Response([
                    'message' => $payload->get_error_message()
                ], 400);
            }

            $codes = $this->sanitize_languages(array_column($payload, 'code'));
            update_option(self::OPTION_LANGUAGES, $codes);

            return new WP_REST_Response([
                'languages' => $codes,
                'raw'       => $payload
            ]);
        }

        public function rest_get_404_logs(): WP_REST_Response
        {
            return new WP_REST_Response([
                'logs' => array_values($this->get_404_logs()),
            ]);
        }

        private function get_404_logs(): array
        {
            $stored = get_option(self::OPTION_404_LOG, []);
            return is_array($stored) ? $stored : [];
        }

        private function normalize_404_path($path): string
        {
            $p = is_string($path) ? trim($path) : '';
            if ($p === '') {
                if (isset($_SERVER['REQUEST_URI'])) { $p = (string) $_SERVER['REQUEST_URI']; }
            }
            if ($p === '') { return ''; }
            // only path+query, strip domain
            $p = preg_replace('#^[a-z]+://[^/]+#i', '', $p);
            if ($p === '') { $p = '/'; }
            return $p;
        }

        private function delete_404_entry(string $path): void
        {
            $path = $this->normalize_404_path($path);
            if ($path === '') { return; }
            $logs = $this->get_404_logs();
            unset($logs[$path]);
            update_option(self::OPTION_404_LOG, $logs);
        }

        private function set_404_resolved(string $path, bool $resolved): void
        {
            $path = $this->normalize_404_path($path);
            if ($path === '') { return; }
            $logs = $this->get_404_logs();
            if (!isset($logs[$path]) || !is_array($logs[$path])) { return; }
            $logs[$path]['resolved'] = $resolved;
            update_option(self::OPTION_404_LOG, $logs);
        }

        public function capture_404(): void
        {
            if (!is_404()) { return; }
            $path = $this->normalize_404_path($_SERVER['REQUEST_URI'] ?? '');
            if ($path === '') { return; }

            $logs = $this->get_404_logs();
            $now  = time();
            $ref  = isset($_SERVER['HTTP_REFERER']) ? (string) $_SERVER['HTTP_REFERER'] : '';
            if (!isset($logs[$path])) {
                $logs[$path] = [
                    'path'         => $path,
                    'count'        => 1,
                    'first_seen'   => $now,
                    'last_seen'    => $now,
                    'last_referrer'=> $ref,
                    'resolved'     => false,
                ];
            } else {
                $logs[$path]['count'] = (int) ($logs[$path]['count'] ?? 0) + 1;
                $logs[$path]['last_seen'] = $now;
                if ($ref) { $logs[$path]['last_referrer'] = $ref; }
            }
            update_option(self::OPTION_404_LOG, $logs);
        }

        public function rest_update_404_logs(WP_REST_Request $request): WP_REST_Response
        {
            $action = sanitize_key($request->get_param('action'));
            $path   = $this->normalize_404_path($request->get_param('path'));

            switch ($action) {
                case 'clear_all':
                    delete_option(self::OPTION_404_LOG);
                    break;
                case 'delete':
                    if ($path !== '') {
                        $this->delete_404_entry($path);
                    }
                    break;
                case 'mark_resolved':
                    if ($path !== '') {
                        $this->set_404_resolved($path, true);
                    }
                    break;
                case 'mark_unresolved':
                    if ($path !== '') {
                        $this->set_404_resolved($path, false);
                    }
                    break;
                default:
                    return new WP_REST_Response([
                        'message' => __('Invalid action.', 'mytheme-seo'),
                    ], 400);
            }

            return new WP_REST_Response([
                'logs' => array_values($this->get_404_logs()),
            ]);
        }

        /**
         * Add field to default WP REST responses so Nuxt can consume without extra calls.
         */
        public function register_post_meta_field(): void
        {
            $post_types = apply_filters('mytheme_seo_post_types', ['page', 'post', 'product']);

            register_rest_field($post_types, 'mytheme_seo', [
                'get_callback' => function (array $object) {
                    $post_id = (int) ($object['id'] ?? 0);
                    return $post_id > 0 ? self::get_post_seo($post_id) : null;
                },
                'schema'       => [
                    'description' => __('MyTheme multilingual SEO payload.', 'mytheme-seo'),
                    'type'        => 'object',
                ],
            ]);
        }

        /**
         * Retrieve stored SEO payload for given post.
         */
        public static function get_post_seo(int $post_id): array
        {
            $stored = get_post_meta($post_id, self::META_KEY, true);
            return is_array($stored) ? $stored : [];
        }

        private function sanitize_payload(array $payload): array
        {
            $languages = $this->get_languages();
            $result    = [];

            foreach ($payload as $locale => $data) {
                $locale = sanitize_key($locale);
                if (!in_array($locale, $languages, true) || !is_array($data)) {
                    continue;
                }

                $result[$locale] = [
                    'title'       => isset($data['title']) ? sanitize_text_field($data['title']) : '',
                    'description' => isset($data['description']) ? sanitize_textarea_field($data['description']) : '',
                    'focus_keyword' => isset($data['focus_keyword']) ? sanitize_text_field($data['focus_keyword']) : '',
                    'video'       => $this->sanitize_video_entries($data['video'] ?? []),
                    'images'      => $this->sanitize_image_entries($data['images'] ?? []),
                    'og'          => [
                        'title'       => isset($data['og']['title']) ? sanitize_text_field($data['og']['title']) : '',
                        'description' => isset($data['og']['description']) ? sanitize_textarea_field($data['og']['description']) : '',
                        'image'       => isset($data['og']['image']) ? esc_url_raw($data['og']['image']) : '',
                    ],
                    'twitter'     => [
                        'card'        => isset($data['twitter']['card']) ? sanitize_text_field($data['twitter']['card']) : '',
                        'title'       => isset($data['twitter']['title']) ? sanitize_text_field($data['twitter']['title']) : '',
                        'description' => isset($data['twitter']['description']) ? sanitize_textarea_field($data['twitter']['description']) : '',
                        'image'       => isset($data['twitter']['image']) ? esc_url_raw($data['twitter']['image']) : '',
                    ],
                    'jsonld'      => isset($data['jsonld']) ? $this->sanitize_jsonld($data['jsonld']) : [],
                ];
            }

            return $result;
        }

        private function sanitize_jsonld($value): array
        {
            if (is_string($value)) {
                $decoded = json_decode($value, true);
                $value   = is_array($decoded) ? $decoded : [];
            }

            return is_array($value) ? $value : [];
        }

        private function get_languages(): array
        {
            $languages = get_option(self::OPTION_LANGUAGES, []);

            if (!is_array($languages) || count($languages) === 0) {
                $languages = apply_filters('mytheme_seo_default_languages', []);
            }

            return $languages;
        }

        private function load_languages_from_source(?string $source)
        {
            $source = is_string($source) && trim($source) !== '' ? trim($source) : null;

            if ($source && filter_var($source, FILTER_VALIDATE_URL)) {
                $response = wp_remote_get($source);
                if (is_wp_error($response)) {
                    return $response;
                }
                $body = wp_remote_retrieve_body($response);
            } else {
                $path = trailingslashit(get_stylesheet_directory()) . self::DEFAULT_LANGUAGE_SOURCE;
                if (!file_exists($path)) {
                    return new WP_Error('language_source_missing', sprintf(__('Language source file not found: %s', 'mytheme-seo'), $path));
                }
                $body = file_get_contents($path);
            }

            $decoded = json_decode($body ?? '', true);
            if (!is_array($decoded)) {
                return new WP_Error('language_source_invalid', __('Language source payload is invalid JSON.', 'mytheme-seo'));
            }

            return $decoded;
        }

        /**
         * Append external sitemap URLs (from settings.sitemaps.externalUrls) to the WP sitemap index.
         * This prevents a fatal error due to a missing callback hooked in the constructor.
         *
         * @param mixed $sitemap_index The sitemap index structure produced by core (SimpleXMLElement in modern WP).
         * @return mixed The (possibly) augmented sitemap index.
         */
        public function filter_wp_sitemaps_index($sitemap_index)
        {
          // Be conservative: only attempt to append if it's an XML element
          if ($sitemap_index instanceof \SimpleXMLElement) {
            $settings = $this->get_settings();
            
            // Add FAQ pages to sitemap if enabled
            $types = isset($settings['sitemaps']['types']) ? $settings['sitemaps']['types'] : [];
            if (!empty($types['faq'])) {
              $languages = $this->get_languages();
              foreach ($languages as $lang) {
                $faq_url = home_url("/{$lang}/faq");
                $sm = $sitemap_index->addChild('sitemap');
                if ($sm) {
                  $sm->addChild('loc', esc_url($faq_url));
                  $sm->addChild('lastmod', gmdate('c'));
                }
              }
            }
            
            // Add external URLs
            $urls = isset($settings['sitemaps']['externalUrls']) && is_array($settings['sitemaps']['externalUrls'])
              ? $settings['sitemaps']['externalUrls']
              : [];
            foreach ($urls as $u) {
              $u = esc_url_raw((string) $u);
              if ($u === '') { continue; }
              $sm = $sitemap_index->addChild('sitemap');
              if ($sm) {
                $sm->addChild('loc', esc_url($u));
                $sm->addChild('lastmod', gmdate('c'));
              }
            }
          }
          return $sitemap_index;
        }
    }
}

// 不再在此处自动初始化，由 tanzanite-setting.php 统一管理
// MyTheme_SEO_Plugin::instance();
