<?php
// ===== Locale persistence and detection =====
if (!function_exists('mytheme_vue_supported_locales')) {
    /**
     * 返回支持的语言列表 [code => name]
     * 支持 34 种语言
     */
    function mytheme_vue_supported_locales() {
        return array(
            'en' => 'English',
            'zh' => '中文',
            'fr' => 'Français',
            'de' => 'Deutsch',
            'es' => 'Español',
            'ja' => '日本語',
            'ko' => '한国어',
            'it' => 'Italiano',
            'pt' => 'Português',
            'ru' => 'Русский',
            'ar' => 'العربية',
            'fi' => 'Suomi',
            'da' => 'Dansk',
            'th' => 'ไทย',
            // 第一批新增 8 种语言
            'sv' => 'Svenska',           // 瑞典语
            'id' => 'Bahasa Indonesia',  // 印尼语
            'ms' => 'Bahasa Melayu',     // 马来语
            'nl' => 'Nederlands',        // 荷兰语
            'tr' => 'Türkçe',            // 土耳其语
            'fil' => 'Filipino',         // 菲律宾语
            'tl' => 'Tagalog',           // 他加禄语
            'jv' => 'Basa Jawa',         // 爪哇语
            // 第二批新增 12 种语言
            'hi' => 'हिन्दी',            // 印地语
            'ur' => 'اردو',              // 乌尔都语
            'mr' => 'मराठी',             // 马拉地语
            'ta' => 'தமிழ்',             // 泰米尔语
            'te' => 'తెలుగు',            // 泰卢固语
            'bn' => 'বাংলা',             // 孟加拉语
            'fa' => 'فارسی',             // 波斯语
            'ps' => 'پښتو',              // 普什图语
            'ha' => 'Hausa',             // 豪萨语
            'sw' => 'Kiswahili',         // 斯瓦希里语
            'pcm' => 'Nigerian Pidgin',  // 尼日利亚皮钦语
            'be' => 'Беларуская',        // 白俄罗斯语
        );
    }
}

/**
 * 统一 Nonce 验证辅助函数
 * 用于 REST API 端点的 nonce 验证
 */
function mytheme_vue_verify_rest_nonce($request) {
    $nonce = $request->get_header('X-WP-Nonce');
    if (empty($nonce)) {
        // 兼容：也检查 GET/POST 参数
        $nonce = $request->get_param('_wpnonce');
    }
    if (empty($nonce) || !wp_verify_nonce($nonce, 'wp_rest')) {
        return false;
    }
    return true;
}

/**
 * 统一 Permission Callback - 需要 nonce 验证（公开端点）
 */
function mytheme_vue_permission_with_nonce($request) {
    return mytheme_vue_verify_rest_nonce($request);
}

/**
 * 统一 Permission Callback - 需要登录 + nonce 验证
 */
function mytheme_vue_permission_logged_in_with_nonce($request) {
    if (!is_user_logged_in()) {
        return new WP_Error('unauthorized', __('需要登录', 'mytheme-vue'), array('status' => 401));
    }
    if (!mytheme_vue_verify_rest_nonce($request)) {
        return new WP_Error('invalid_nonce', __('无效的安全令牌', 'mytheme-vue'), array('status' => 403));
    }
    return true;
}

// ===== REST: /mytheme-vue/v1/cart-summary =====
if (!function_exists('mytheme_vue_register_rest_routes')) {
    function mytheme_vue_register_rest_routes() {
        register_rest_route('mytheme-vue/v1', '/cart-summary', array(
            'methods'  => 'GET',
            'callback' => function($request) {
                $summary = array(
                    'count' => 0,
                    'weight_grams' => 0,
                    'total' => 0.0,
                    'currency' => 'USD',
                    'symbol' => '$',
                );

                $calculation = apply_filters('tanzanite_cart_summary', null, $request);
                if (is_array($calculation)) {
                    $summary = array_merge($summary, array_filter($calculation, 'is_scalar'));
                }

                return rest_ensure_response($summary);
            },
            'permission_callback' => '__return_true',
        ));

        $namespaces = array('mytheme/v1', 'tanzanite/v1');

        foreach ($namespaces as $namespace) {
            register_rest_route($namespace, '/settings/site', array(
                'methods'  => 'GET',
                'callback' => 'mytheme_vue_rest_get_site_settings',
                'permission_callback' => '__return_true',
            ));

            register_rest_route($namespace, '/settings/quick-buy', array(
                'methods'  => 'GET',
                'callback' => 'mytheme_vue_rest_get_quickbuy_settings',
                'permission_callback' => '__return_true',
            ));
        }
    }
    add_action('rest_api_init', 'mytheme_vue_register_rest_routes');
}

function mytheme_vue_rest_get_site_settings() {
    $site_title       = get_bloginfo('name');
    $site_description = get_bloginfo('description');
    $site_logo_id     = get_theme_mod('custom_logo');
    $site_logo_src    = $site_logo_id ? wp_get_attachment_image_url($site_logo_id, 'full') : '';

    return rest_ensure_response(array(
        'siteTitle'       => $site_title,
        'siteDescription' => $site_description,
        'siteLogo'        => $site_logo_src,
    ));
}


function mytheme_vue_rest_get_quickbuy_settings() {
    $response = array(
        'steps'        => array(),
        'storeApiBase' => rest_url('tanzanite/v1'),
        'cartUrl'      => '',
        'checkoutUrl'  => '',
        'buttonText'   => get_theme_mod('quickbuy_button_text', 'Quick Buy'),
        'enabled'      => get_theme_mod('quickbuy_enabled', true) ? true : false,
    );

    // 允许自建商品系统覆盖默认数据
    $response = apply_filters('tanzanite_quickbuy_defaults', $response);

    if (empty($response['cartUrl']) && function_exists('wc_get_cart_url')) {
        $response['cartUrl'] = wc_get_cart_url();
    }

    if (empty($response['checkoutUrl']) && function_exists('wc_get_checkout_url')) {
        $response['checkoutUrl'] = wc_get_checkout_url();
    }

    // 如果自建插件没有提供步骤，则回退到主题设置（兼容旧版）
    if (empty($response['steps'])) {
        for ($i = 1; $i <= 3; $i++) {
            $cat_id   = (int) get_theme_mod('quickbuy_step' . $i . '_cat', 0);
            $footer   = get_theme_mod('quickbuy_step' . $i . '_footer', '');
            $stepName = $cat_id ? get_term_field('name', $cat_id, 'product_cat') : __('All products', 'mytheme-vue');
            $slug     = $cat_id ? get_term_field('slug', $cat_id, 'product_cat') : '';

            $response['steps'][] = array(
                'id'     => $cat_id,
                'slug'   => is_wp_error($slug) ? '' : (string) $slug,
                'name'   => is_wp_error($stepName) ? '' : (string) $stepName,
                'footer' => $footer,
            );
        }
    }

    return rest_ensure_response($response);
}

if (!function_exists('mytheme_vue_sanitize_lang')) {
    function mytheme_vue_sanitize_lang($lang) {
        $lang = strtolower(trim((string)$lang));
        $lang = preg_replace('/[^a-z\-]/', '', $lang); // 只允许字母和-
        return $lang ?: 'en';
    }
}

// 使用新版路径前缀逻辑，旧的 mytheme_vue_detect_locale() / mytheme_vue_bootstrap_locale_cookie() 已移除
 
/**
 * Tanzanite Theme Functions
 * Nuxt 3 + WordPress REST 架构主题功能文件
 *
 * @package TanzaniteTheme
 */

// 防止直接访问
if (!defined('ABSPATH')) {
    exit;
}

/**
 * 主题设置
 */
define('MYTHEME_VUE_DIR', get_template_directory());
define('MYTHEME_VUE_URI', get_template_directory_uri());
// 动态版本号：主题版本 + style.css 的修改时间，便于调试与缓存区分
$__theme_obj = function_exists('wp_get_theme') ? wp_get_theme() : null;
$__theme_ver = ($__theme_obj && $__theme_obj->get('Version')) ? $__theme_obj->get('Version') : '1.0.0';
$__style_mtime = @filemtime(MYTHEME_VUE_DIR . '/style.css');
define('MYTHEME_VUE_VERSION', $__theme_ver . '.' . ($__style_mtime ? $__style_mtime : 'dev'));

/**
 * （已移除）页脚动态组件集成
 */

/**
 * 主题初始化
 */
function mytheme_vue_setup() {
    // 添加主题支持
    add_theme_support('post-thumbnails');
    add_theme_support('title-tag');
    add_theme_support('html5', array(
        'search-form',
        'comment-form',
        'comment-list',
        'gallery',
        'caption',
    ));
    
    // 注册导航菜单
    register_nav_menus(array(
        'primary' => __('主导航菜单', 'mytheme-vue'),
        'footer' => __('页脚菜单', 'mytheme-vue'),
    ));
}
add_action('after_setup_theme', 'mytheme_vue_setup');

// ===== A1: 语言前缀路由（/fr,/de,/es,...）与服务端 locale 切换 =====
if (!function_exists('mytheme_vue_language_options')) {
    /**
     * 生成语言选项（用于后台设置、REST API）
     */
    function mytheme_vue_language_options() {
        $supported = mytheme_vue_supported_locales();
        $options = array();
        foreach ($supported as $code => $name) {
            $options[$code] = $name;
        }
        return $options;
    }
}

/**
 * 注册样式和脚本
 */
function mytheme_vue_scripts() {
    // CSS（为关键文件使用各自的 mtime 作为版本，避免缓存）
    $ver_style       = MYTHEME_VUE_VERSION; // 基础版本（含 style.css mtime）
    wp_enqueue_style('mytheme-vue-style', MYTHEME_VUE_URI . '/style.css', [], $ver_style);
    // Header Menu 已迁移到 Nuxt，dock-menu.js 和 popup-coordinator.js 不再需要

    // 确保 WP 自带 underscore 先加载，并保存引用，避免被其他库覆盖
    wp_enqueue_script('underscore');
    wp_add_inline_script('underscore', 'window.__wp_underscore = window._;', 'after');
    wp_add_inline_script('underscore', "(function(){var good=window.__wp_underscore||window._;function guard(){try{if(!window._||typeof window._.isEqual!=='function'||typeof window._.memoize!=='function'){window._=good;}}catch(e){window._=good;}}guard();setInterval(guard,1000);}());", 'after');

    // Vite 开发/生产资源加载
    $use_vite = defined('MYTHEME_VITE_DEV') && MYTHEME_VITE_DEV;
    $localize_handle = '';
    if ($use_vite) {
        // Dev Server：加载 @vite/client 与入口
        wp_enqueue_script('vite-client', 'http://localhost:5173/@vite/client', [], null, false);
        wp_enqueue_script('mytheme-vite-entry', 'http://localhost:5173/js/main-vite.js', ['underscore'], null, true);
        // 在我们入口执行前/后恢复 underscore
        wp_add_inline_script('mytheme-vite-entry', 'if (!window._ || typeof window._.isEqual !== "function" || typeof window._.memoize !== "function") { window._ = window.__wp_underscore || window._; }', 'before');
        wp_add_inline_script('mytheme-vite-entry', '(function(){var good=window.__wp_underscore||window._;if(!window._||typeof window._.isEqual!=="function"||typeof window._.memoize!=="function"){window._=good;}}());', 'after');
        $localize_handle = 'mytheme-vite-entry';
    } else {
        // 生产：读取 manifest.json（Vite 5 默认为 dist/.vite/manifest.json；兼容旧路径 dist/manifest.json）
        $manifest_path_v5 = MYTHEME_VUE_DIR . '/dist/.vite/manifest.json';
        $manifest_path_v4 = MYTHEME_VUE_DIR . '/dist/manifest.json';
        $manifest_path = file_exists($manifest_path_v5) ? $manifest_path_v5 : $manifest_path_v4;
        if (file_exists($manifest_path)) {
            $manifest = json_decode(file_get_contents($manifest_path), true);
            $uri_base = MYTHEME_VUE_URI . '/dist/';
            // 找到入口：优先精确匹配 'js/main-vite.js'，否则寻找 isEntry=true 的项
            $entry_key = null;
            if (isset($manifest['js/main-vite.js'])) {
                $entry_key = 'js/main-vite.js';
            } else {
                foreach ($manifest as $k => $item) {
                    if (!is_array($item)) continue;
                    $is_entry = isset($item['isEntry']) ? (bool)$item['isEntry'] : false;
                    $ends_with_main = false;
                    if (is_string($k)) {
                        $needle = 'js/main-vite.js';
                        $ends_with_main = substr($k, -strlen($needle)) === $needle;
                    }
                    if ($is_entry || $ends_with_main) { $entry_key = $k; break; }
                }
            }
            if ($entry_key && isset($manifest[$entry_key])) {
                $item = $manifest[$entry_key];
                if (!empty($item['css']) && is_array($item['css'])) {
                    foreach ($item['css'] as $cssFile) {
                        wp_enqueue_style('mytheme-vite-css', $uri_base . ltrim($cssFile, '/'), [], null);
                    }
                }
                if (!empty($item['file'])) {
                    wp_enqueue_script('mytheme-vite-js', $uri_base . ltrim($item['file'], '/'), ['underscore'], null, true);
                    // Vite 产物为 ESM，需标注为 module
                    wp_script_add_data('mytheme-vite-js', 'type', 'module');
                    // 在我们入口执行前/后恢复 underscore
                    wp_add_inline_script('mytheme-vite-js', 'if (!window._ || typeof window._.isEqual !== "function" || typeof window._.memoize !== "function") { window._ = window.__wp_underscore || window._; }', 'before');
                    wp_add_inline_script('mytheme-vite-js', '(function(){var good=window.__wp_underscore||window._;if(!window._||typeof window._.isEqual!=="function"||typeof window._.memoize!=="function"){window._=good;}}());', 'after');
                    $localize_handle = 'mytheme-vite-js';
                }
            }
            // 追加：入队 sidebar-entry（若存在）
            foreach ($manifest as $k => $it) {
                if (!is_array($it)) continue;
                $src = isset($it['src']) ? (string)$it['src'] : '';
                if ((strpos($src, 'js/sidebar-entry.js') !== false) && !empty($it['file'])) {
                    if (!empty($it['css']) && is_array($it['css'])) {
                        foreach ($it['css'] as $cssFile) {
                            wp_enqueue_style('mytheme-sidebar-css', $uri_base . ltrim($cssFile, '/'), [], null);
                        }
                    }
                    wp_enqueue_script('mytheme-sidebar-js', $uri_base . ltrim($it['file'], '/'), ['mytheme-vite-js'], null, true);
                    wp_script_add_data('mytheme-sidebar-js', 'type', 'module');
                    break;
                }
            }
        }
    }

    // 侧栏交互脚本（旧版）已停用，改由 iframe 集成最小脚本处理
    // wp_enqueue_script('mytheme-vue-sidebar-left', MYTHEME_VUE_URI . '/js/sidebar-left.js', [], MYTHEME_VUE_VERSION, true);
}
add_action('wp_enqueue_scripts', 'mytheme_vue_scripts');

/**
 * Enqueue Nuxt sidebar micro-app build artifacts (CSR widget)
 */
function mytheme_enqueue_nuxt_sidebar_widget() { /* removed: using iframe integration */ }
// 使用 iframe 集成后，不再通过 WP 队列加载 Nuxt 产物
// add_action('wp_enqueue_scripts', 'mytheme_enqueue_nuxt_sidebar_widget', 30);

// Removed: legacy script_loader_tag for 'nuxt-sidebar-script' (iframe integration)

// 让 Vite 构建产物以 ESM 模块加载（type="module"），同时兜底任何指向当前主题 dist 的脚本
add_filter('script_loader_tag', function($tag, $handle){
  try {
    $needles = array(
      'mytheme-vite-js',
      'mytheme-sidebar-js'
    );
    $must_module = in_array($handle, $needles, true);
    if (!$must_module) {
      $theme_dist = trailingslashit(get_template_directory_uri()) . 'dist/';
      if (is_string($tag) && strpos($tag, $theme_dist) !== false) {
        $must_module = true;
      }
    }
    if ($must_module) {
      if (strpos($tag, 'type="module"') === false) {
        if (strpos($tag, 'type=') !== false) {
          // 将已有的 type（如 text/javascript）替换为 module
          $tag = preg_replace('/type\s*=\s*"[^"]*"/i', 'type="module"', $tag);
        } else {
          // 无 type 时追加
          $tag = str_replace(' src="', ' type="module" src="', $tag);
        }
      }
    }
  } catch (Exception $e) { /* no-op */ }
  return $tag;
}, 999, 2);

// ===== A6: 自定义 hreflang Sitemap（/sitemap-hreflang.xml） =====
if (!function_exists('mytheme_hreflang_add_rule')) {
  function mytheme_hreflang_add_rule(){
    add_rewrite_rule('^sitemap-hreflang\.xml$', 'index.php?mytheme_hreflang=1', 'top');
  }
  add_action('init', 'mytheme_hreflang_add_rule');
}

if (!function_exists('mytheme_hreflang_query_vars')) {
  function mytheme_hreflang_query_vars($vars){
    $vars[] = 'mytheme_hreflang';
    return $vars;
  }
  add_filter('query_vars', 'mytheme_hreflang_query_vars');
}

if (!function_exists('mytheme_hreflang_sitemap_output')) {
  function mytheme_hreflang_sitemap_output(){
    if (!get_query_var('mytheme_hreflang')) return;
    // 输出 XML sitemap（仅承载 hreflang 互链）
    header('Content-Type: application/xml; charset=UTF-8');
    $supported = mytheme_vue_supported_locales();
    $codes = array_keys($supported);
    $home_default = home_url('/');
    $default_code = 'en'; // 默认语言不带前缀

    echo "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n";
    echo '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">' . "\n";

    foreach ($codes as $code){
      // 语言首页 URL：默认语言用 /，其他用 /{code}/
      $loc = ($code === $default_code) ? $home_default : home_url('/' . $code . '/');
      echo '  <url>' . "\n";
      echo '    <loc>' . esc_url($loc) . '</loc>' . "\n";
      // 为所有语言版本输出互链
      foreach ($codes as $alt){
        $href = ($alt === $default_code) ? $home_default : home_url('/' . $alt . '/');
        $hreflang = esc_attr($alt);
        echo '    <xhtml:link rel="alternate" hreflang="' . $hreflang . '" href="' . esc_url($href) . '" />' . "\n";
      }
      // x-default 指向站点默认首页
      echo '    <xhtml:link rel="alternate" hreflang="x-default" href="' . esc_url($home_default) . '" />' . "\n";
      echo '  </url>' . "\n";
    }

    echo '</urlset>';
    exit;
  }
  add_action('template_redirect', 'mytheme_hreflang_sitemap_output');
}

/**
 * Nuxt Sidebar iframe helpers: auto-height via postMessage and toggle open/close
 */
function mytheme_nuxt_sidebar_iframe_helpers() {
    if (is_admin()) return;
    ?>
    <script>
    (function(){
      // Sync handle size from Nuxt (child) -> parent, then set iframe width = 50vw + handle
      window.addEventListener('message', function(ev){
        try{
          var d = ev.data || {};
          if (d && d.type === 'SB_HANDLE_SIZE'){
            var val = String(d.value||'').trim() || '33px';
            document.documentElement.style.setProperty('--sb-btn-size', val);
            var iframe = document.getElementById('nuxt-sidebar-iframe');
            if (iframe){ iframe.style.width = 'calc(50vw + ' + val + ')'; }
          }
        }catch(_){ }
      }, false);

      // Helpers
      function setHeight(px){
        var wrap = document.getElementById('nuxt-sidebar-wrapper');
        if (!wrap) return;
        wrap.style.setProperty('--iframe-h', (Math.max(200, Math.min(1200, px|0))) + 'px');
      }

      // Auto-resize (A): postMessage fallback if the app emits height
      window.addEventListener('message', function(ev){
        try{
          var d = ev.data || {};
          if (d && d.type === 'nuxt-sidebar:height') { setHeight(d.height|0); }
        }catch(_){}
      }, false);

      // Auto-resize (B): same-origin measure using ResizeObserver + polling fallback
      function bindAutoResize(){
        var iframe = document.getElementById('nuxt-sidebar-iframe');
        if (!iframe) return;
        function measure(){
          try{
            var doc = iframe.contentDocument || iframe.contentWindow && iframe.contentWindow.document;
            if (!doc) return;
            var h = Math.max(doc.documentElement.scrollHeight||0, doc.body && doc.body.scrollHeight||0);
            if (h) setHeight(h);
          }catch(_){}
        }
        iframe.addEventListener('load', function(){
          try{
            var doc = iframe.contentDocument || (iframe.contentWindow && iframe.contentWindow.document);
            if (!doc) return;
            // Initial measure
            setTimeout(measure, 50);
            // ResizeObserver on body
            if (window.ResizeObserver && doc.body){
              var ro = new ResizeObserver(function(){ measure(); });
              ro.observe(doc.body);
            }
            // Fallback polling
            var last = 0; setInterval(function(){
              try{
                var d = iframe.contentDocument || iframe.contentWindow.document;
                if (!d) return;
                var cur = Math.max(d.documentElement.scrollHeight||0, d.body && d.body.scrollHeight||0);
                if (cur && cur !== last){ last = cur; setHeight(cur); }
              }catch(_){}
            }, 800);
          }catch(_){}
        });
      }

      // Toggle open/close
      function bindToggle(){
        // Direct binding (if present at load)
        var wrap = document.getElementById('nuxt-sidebar-wrapper');
        if (wrap){
          var btn = wrap.querySelector('.sidebar-bulge-button');
          if (btn){
            btn.addEventListener('click', function(){ wrap.classList.toggle('is-center'); });
          }
        }
        // Delegated binding: handle dynamically rendered buttons
        document.addEventListener('click', function(ev){
          var btn = ev.target.closest && ev.target.closest('.sidebar-bulge-button');
          if (!btn) return;
          var container = btn.closest && btn.closest('.sidebar-left');
          if (!container) return;
          ev.preventDefault();
          container.classList.toggle('is-center');
        }, {passive:false});
      }
      // Dropdown transforms for sidebar filters
      function bindSearchDropdowns(){
        var blocks = document.querySelectorAll('.sidebar-search-block');
        if (!blocks.length) return;
        blocks.forEach(function(block){
          var groups = block.querySelectorAll('.sb-group');
          groups.forEach(function(group){
            if (group.__sbInited) return; group.__sbInited = true;
            var titleEl = group.querySelector('.sb-title');
            var opts = group.querySelector('.sb-options');
            if (!titleEl || !opts) return;

            // Create select-like button (keep title visible, control goes AFTER title)
            var btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'sb-select';
            btn.setAttribute('aria-expanded', 'false');
            var label = titleEl.textContent.trim();
            btn.setAttribute('aria-label', label);
            btn.innerHTML = '<span class="sb-select-value">&nbsp;</span><span class="sb-select-caret">▾</span>';

            // Wrap options into panel
            var panel = document.createElement('div');
            panel.className = 'sb-panel';
            panel.appendChild(opts);

            // Insert elements (after title)
            titleEl.insertAdjacentElement('afterend', btn);
            btn.insertAdjacentElement('afterend', panel);

            function updateSummary(){
              var cbs = panel.querySelectorAll('input[type="checkbox"]');
              var allCb = panel.querySelector('input[value="all"]');
              var selected = [];
              cbs.forEach(function(cb){
                if (cb.checked && cb !== allCb) {
                  var lab = cb.closest('label');
                  if (lab) selected.push(lab.textContent.trim());
                }
              });
              if (allCb && allCb.checked) { selected = ['All']; }
              var valEl = btn.querySelector('.sb-select-value');
              valEl.textContent = selected.length ? selected.join(', ') : '';
            }

            // Behavior: toggle panel
            btn.addEventListener('click', function(e){
              e.preventDefault();
              var opened = btn.getAttribute('aria-expanded') === 'true';
              document.querySelectorAll('.sb-select[aria-expanded="true"]').forEach(function(b){
                if (b!==btn){ b.setAttribute('aria-expanded','false'); b.nextElementSibling && b.nextElementSibling.classList.remove('open'); }
              });
              btn.setAttribute('aria-expanded', opened ? 'false' : 'true');
              panel.classList.toggle('open', !opened);
            });

            // Behavior: All exclusive
            panel.addEventListener('change', function(ev){
              var t = ev.target;
              if (t && t.type === 'checkbox'){
                var allCb = panel.querySelector('input[value="all"]');
                if (allCb && t === allCb && t.checked){
                  panel.querySelectorAll('input[type="checkbox"]').forEach(function(cb){ if (cb!==allCb) cb.checked = false; });
                } else if (allCb && t !== allCb && t.checked){
                  allCb.checked = false;
                }
                updateSummary();
              }
            });

            // Close on outside click
            document.addEventListener('click', function(ev){
              if (!group.contains(ev.target)){
                btn.setAttribute('aria-expanded','false');
                panel.classList.remove('open');
              }
            });

            // Init summary
            updateSummary();
          });
        });
      }

      function init(){ bindToggle(); bindAutoResize(); bindSearchDropdowns(); }
      if (document.readyState === 'complete' || document.readyState === 'interactive') init();
      else document.addEventListener('DOMContentLoaded', init);
    })();
    </script>
    <?php
}
// 已提取为 assets/js/sidebar-helpers.js，通过 wp_enqueue_script 加载
// add_action('wp_footer', 'mytheme_nuxt_sidebar_iframe_helpers', 35);

/**
 * Render breadcrumb <li> items (Home + hierarchical trail)
 */
function mytheme_vue_render_breadcrumbs(){
  echo '<li class="breadcrumb-item"><a class="breadcrumb-link" href="' . esc_url(home_url('/')) . '">' . esc_html__( 'Home', 'mytheme-vue' ) . '</a></li>';

  if (is_front_page()) return;

  $post_id = get_the_ID();
  if (!$post_id) { // fallback to title if no singular context
    // Archive contexts below
    if (is_search()) {
      echo '<li class="breadcrumb-item" aria-current="page">' . esc_html(sprintf(__('Search results for "%s"','mytheme-vue'), get_search_query())) . '</li>';
      return;
    }
    if (is_category()) {
      $cat = get_queried_object();
      if ($cat && !is_wp_error($cat)) {
        $parents = array_reverse(get_ancestors($cat->term_id, 'category'));
        foreach ($parents as $pid) {
          $term = get_term($pid, 'category');
          if ($term && !is_wp_error($term)) echo '<li class="breadcrumb-item"><a class="breadcrumb-link" href="' . esc_url(get_category_link($term)) . '">' . esc_html($term->name) . '</a></li>';
        }
        echo '<li class="breadcrumb-item" aria-current="page">' . esc_html($cat->name) . '</li>';
        return;
      }
    }
    if (is_tag()) {
      $tag = get_queried_object();
      if ($tag && !is_wp_error($tag)) { echo '<li class="breadcrumb-item" aria-current="page">' . esc_html($tag->name) . '</li>'; return; }
    }
    if (is_tax()) {
      $tax = get_queried_object();
      if ($tax && !is_wp_error($tax)) {
        $parents = array_reverse(get_ancestors($tax->term_id, $tax->taxonomy));
        foreach ($parents as $pid) {
          $term = get_term($pid, $tax->taxonomy);
          if ($term && !is_wp_error($term)) echo '<li class="breadcrumb-item"><a class="breadcrumb-link" href="' . esc_url(get_term_link($term)) . '">' . esc_html($term->name) . '</a></li>';
        }
        echo '<li class="breadcrumb-item" aria-current="page">' . esc_html($tax->name) . '</li>';
        return;
      }
    }
    if (function_exists('is_shop') && is_shop()) {
      echo '<li class="breadcrumb-item" aria-current="page">' . esc_html(get_the_title(wc_get_page_id('shop'))) . '</li>';
      return;
    }
    if (is_post_type_archive()) {
      $pt = get_query_var('post_type');
      $obj = get_post_type_object($pt);
      echo '<li class="breadcrumb-item" aria-current="page">' . esc_html($obj ? $obj->labels->name : ucfirst($pt)) . '</li>';
      return;
    }
    // Fallback
    echo '<li class="breadcrumb-item" aria-current="page">' . esc_html(get_bloginfo('name')) . '</li>';
    return;
  }

  // Page hierarchy
  if (is_page()) {
    $ancestors = array_reverse(get_post_ancestors($post_id));
    foreach ($ancestors as $aid) {
      echo '<li class="breadcrumb-item"><a class="breadcrumb-link" href="' . esc_url(get_permalink($aid)) . '">' . esc_html(get_the_title($aid)) . '</a></li>';
    }
    echo '<li class="breadcrumb-item" aria-current="page">' . esc_html(get_the_title($post_id)) . '</li>';
    return;
  }

  // Post hierarchy by category
  if (is_singular('post')) {
    $cats = get_the_category($post_id);
    if ($cats && !is_wp_error($cats)) {
      // choose the deepest category (with greatest depth)
      $primary = null; $depth = -1;
      foreach ($cats as $c) {
        $d = 0; $p = $c->term_id; while ($p) { $p = get_term_field('parent', $p, 'category'); $p = intval($p); if ($p>0) $d++; else break; }
        if ($d > $depth) { $depth = $d; $primary = $c; }
      }
      if ($primary) {
        $parents = array_reverse(get_ancestors($primary->term_id, 'category'));
        foreach ($parents as $pid) {
          $term = get_term($pid, 'category');
          if ($term && !is_wp_error($term)) echo '<li class="breadcrumb-item"><a class="breadcrumb-link" href="' . esc_url(get_category_link($term)) . '">' . esc_html($term->name) . '</a></li>';
        }
        echo '<li class="breadcrumb-item"><a class="breadcrumb-link" href="' . esc_url(get_category_link($primary)) . '">' . esc_html($primary->name) . '</a></li>';
      }
    }
    echo '<li class="breadcrumb-item" aria-current="page">' . esc_html(get_the_title($post_id)) . '</li>';
    return;
  }

  // Tanzanite product hierarchy
  if (is_singular('tanz_product')) {
    $ancestors = array_reverse(get_post_ancestors($post_id));
    foreach ($ancestors as $aid) {
      echo '<li class="breadcrumb-item"><a class="breadcrumb-link" href="' . esc_url(get_permalink($aid)) . '">' . esc_html(get_the_title($aid)) . '</a></li>';
    }
    echo '<li class="breadcrumb-item" aria-current="page">' . esc_html(get_the_title($post_id)) . '</li>';
    return;
  }

  // Fallback
  echo '<li class="breadcrumb-item" aria-current="page">' . esc_html(get_the_title($post_id)) . '</li>';
}

/**
 * Render sidebar-left container on all frontend pages; show breadcrumbs only for archive/search
 */
function mytheme_render_sidebar_left_for_archives_breadcrumbs(){ /* removed: iframe integration deleted */ }
function mytheme_sidebar_footer_fallback(){ /* removed: iframe integration deleted */ }

/**
 * 语言与本地化
 * 注意：mytheme_vue_supported_locales() 函数已在文件开头定义（第 3-48 行）
 */

/**
 * 短代码到完整 Locale 的映射
 * 用于路径前缀：/fr/ -> fr_FR, /zh/ -> zh_CN
 */
function mytheme_vue_short_code_to_locale($short_code) {
    $short_code = strtolower($short_code);
    $map = array(
        // 原有 14 种语言
        'en' => 'en_US',
        'zh' => 'zh_CN',
        'fr' => 'fr_FR',
        'de' => 'de_DE',
        'es' => 'es_ES',
        'ru' => 'ru_RU',
        'pt' => 'pt_BR',
        'ja' => 'ja',
        'ko' => 'ko_KR',
        'it' => 'it_IT',
        'ar' => 'ar',
        'fi' => 'fi',
        'da' => 'da_DK',
        'th' => 'th',
        // 第一批新增 8 种语言
        'sv' => 'sv_SE',
        'id' => 'id_ID',
        'ms' => 'ms_MY',
        'nl' => 'nl_NL',
        'tr' => 'tr_TR',
        'fil' => 'fil_PH',
        'tl' => 'tl_PH',
        'jv' => 'jv_ID',
        // 第二批新增 12 种语言
        'hi' => 'hi_IN',
        'ur' => 'ur_PK',
        'mr' => 'mr_IN',
        'ta' => 'ta_IN',
        'te' => 'te_IN',
        'bn' => 'bn_BD',
        'fa' => 'fa_IR',
        'ps' => 'ps_AF',
        'ha' => 'ha_NG',
        'sw' => 'sw_KE',
        'pcm' => 'pcm_NG',
        'be' => 'be_BY',
    );
    return isset($map[$short_code]) ? $map[$short_code] : 'en_US';
}

/**
 * 完整 Locale 到短代码的映射
 * 用于生成路径前缀：fr_FR -> /fr/, zh_CN -> /zh/
 */
function mytheme_vue_locale_to_short_code($locale) {
    $locale = strtolower($locale);
    $map = array(
        // 原有 14 种语言
        'en_us' => 'en',
        'zh_cn' => 'zh',
        'fr_fr' => 'fr',
        'de_de' => 'de',
        'es_es' => 'es',
        'ru_ru' => 'ru',
        'pt_br' => 'pt',
        'ja' => 'ja',
        'ko_kr' => 'ko',
        'it_it' => 'it',
        'ar' => 'ar',
        'fi' => 'fi',
        'da_dk' => 'da',
        'th' => 'th',
        // 第一批新增 8 种语言
        'sv_se' => 'sv',
        'id_id' => 'id',
        'ms_my' => 'ms',
        'nl_nl' => 'nl',
        'tr_tr' => 'tr',
        'fil_ph' => 'fil',
        'tl_ph' => 'tl',
        'jv_id' => 'jv',
        // 第二批新增 12 种语言
        'hi_in' => 'hi',
        'ur_pk' => 'ur',
        'mr_in' => 'mr',
        'ta_in' => 'ta',
        'te_in' => 'te',
        'bn_bd' => 'bn',
        'fa_ir' => 'fa',
        'ps_af' => 'ps',
        'ha_ng' => 'ha',
        'sw_ke' => 'sw',
        'pcm_ng' => 'pcm',
        'be_by' => 'be',
    );
    return isset($map[$locale]) ? $map[$locale] : 'en';
}

function mytheme_vue_map_country_to_locale($country) {
    $country = strtoupper($country);
    switch ($country) {
        case 'CN': return 'zh_CN';
        case 'HK': return 'zh_TW';
        case 'JP': return 'ja';
        case 'KR': return 'ko_KR';
        case 'DE': return 'de_DE';
        case 'FR': return 'fr_FR';
        case 'ES': return 'es_ES';
        case 'RU': return 'ru_RU';
        case 'BR': return 'pt_BR';
        case 'PT': return 'pt_PT';
        case 'US':
        case 'GB':
        default:   return 'en_US';
    }
}

function mytheme_vue_detect_locale() {
    // 优先级 1：检测 URL 路径前缀（/fr/, /de/, /zh/ 等）
    $request_uri = isset($_SERVER['REQUEST_URI']) ? $_SERVER['REQUEST_URI'] : '/';
    $path = parse_url($request_uri, PHP_URL_PATH);
    $segments = array_filter(explode('/', $path));

    if (!empty($segments)) {
        $first_segment = strtolower(reset($segments));
        $supported_short_codes = array('en','fr','de','es','it','pt','ru','ja','ko','ar','fi','da','th','zh');

        if (in_array($first_segment, $supported_short_codes, true)) {
            // 设置 Cookie 为对应 locale，便于 PHP 层使用
            $locale = mytheme_vue_short_code_to_locale($first_segment);
            if (!headers_sent()) {
                setcookie('mytheme_locale', $locale, time() + (180 * DAY_IN_SECONDS), '/', '', is_ssl(), true);
            }
            return $first_segment;
        }
    }

    // 优先级 2：检测 Cookie（存储为 locale）
    if (!empty($_COOKIE['mytheme_locale'])) {
        $locale = sanitize_text_field($_COOKIE['mytheme_locale']);
        return mytheme_vue_locale_to_short_code($locale);
    }

    // 优先级 3：检测 Cloudflare 国家代码
    if (!empty($_SERVER['HTTP_CF_IPCOUNTRY'])) {
        $locale = mytheme_vue_map_country_to_locale($_SERVER['HTTP_CF_IPCOUNTRY']);
        return mytheme_vue_locale_to_short_code($locale);
    }

    // 优先级 4：检测浏览器语言
    if (!empty($_SERVER['HTTP_ACCEPT_LANGUAGE'])) {
        $al = strtolower($_SERVER['HTTP_ACCEPT_LANGUAGE']);
        if (strpos($al, 'zh-cn') !== false) return 'zh';
        if (strpos($al, 'ja') !== false) return 'ja';
        if (strpos($al, 'ko') !== false) return 'ko';
        if (strpos($al, 'fr') !== false) return 'fr';
        if (strpos($al, 'de') !== false) return 'de';
        if (strpos($al, 'es') !== false) return 'es';
        if (strpos($al, 'ru') !== false) return 'ru';
        if (strpos($al, 'pt-br') !== false) return 'pt';
        if (strpos($al, 'pt') !== false) return 'pt';
        return 'en';
    }

    // 默认：英语
    return 'en';
}

// 注意：翻译现在由前端 JavaScript 处理，不再使用 PHP 运行时翻译
// 仅保留 WordPress 默认 locale，避免服务端与 Nuxt 冲突

/**
 * 注册侧边栏
 */
function mytheme_vue_widgets_init() {
    register_sidebar(array(
        'name'          => __('页脚小工具区域', 'mytheme-vue'),
        'id'            => 'footer-widgets',
        'description'   => __('页脚容器中的小工具区域', 'mytheme-vue'),
        'before_widget' => '<div id="%1$s" class="widget %2$s">',
        'after_widget'  => '</div>',
        'before_title'  => '<h3 class="widget-title">',
        'after_title'   => '</h3>',
    ));
}
add_action('widgets_init', 'mytheme_vue_widgets_init');

/**
 * REST API 扩展
 */

// 添加自定义API端点获取菜单
function mytheme_vue_get_menu_items($request) {
    $menu_location = $request['location'];
    $locations = get_nav_menu_locations();
    
    if (!isset($locations[$menu_location])) {
        return new WP_Error('no_menu', __('菜单不存在', 'mytheme-vue'), array('status' => 404));
    }
    
    $menu_id = $locations[$menu_location];
    $menu_items = wp_get_nav_menu_items($menu_id);
    
    // 构建层级菜单结构
    $menu_tree = array();
    $menu_index = array();
    
    // 先建立索引
    foreach ($menu_items as $item) {
        $menu_index[$item->ID] = array(
            'id' => $item->ID,
            'title' => $item->title,
            'url' => $item->url,
            'parent' => $item->menu_item_parent,
            'object_id' => $item->object_id,
            'object' => $item->object,
            'type' => $item->type,
            'children' => array()
        );
    }
 


    // 构建树形结构
    foreach ($menu_index as $id => $item) {
        if ($item['parent'] == 0) {
            $menu_tree[$id] = $item;
        } elseif (isset($menu_index[$item['parent']])) {
            $menu_index[$item['parent']]['children'][] = $item;
        }
    }
    
    return array_values($menu_tree);
}

// 扩展WordPress REST API - 为文章添加 tags_names 字段
function mytheme_vue_register_api() {
    // 为页面添加 tags_names 字段
    register_rest_field('page', 'tags_names', array(
        'get_callback' => function($post) {
            $tags = get_the_tags($post['id']);
            if ($tags && !is_wp_error($tags) && is_array($tags)) {
                $tag_names = wp_list_pluck($tags, 'name');
                return is_array($tag_names) ? $tag_names : array();
            }
            return array(); // 始终返回空数组
        },
        'update_callback' => null,
        'schema' => array(
            'description' => __('页面标签名称数组', 'mytheme-vue'),
            'type' => 'array',
            'items' => array(
                'type' => 'string'
            )
        )
    ));

    // 注册 REST 路由
    register_rest_route('mytheme-vue/v1', '/menu/(?P<location>[a-zA-Z0-9_-]+)', array(
        'methods' => 'GET',
        'callback' => 'mytheme_vue_get_menu_items',
        'permission_callback' => 'mytheme_vue_permission_with_nonce' // 添加 nonce 验证
    ));

    register_rest_route('mytheme-vue/v1', '/search', array(
        'methods' => 'GET',
        'callback' => 'mytheme_vue_search_content',
        'permission_callback' => 'mytheme_vue_permission_with_nonce' // 添加 nonce 验证
    ));

    register_rest_route('mytheme-vue/v1', '/function-block', array(
        'methods' => 'GET',
        'callback' => 'mytheme_vue_get_function_block_data',
        'permission_callback' => 'mytheme_vue_permission_with_nonce' // 添加 nonce 验证
    ));

    register_rest_route('mytheme-vue/v1', '/function-block/action', array(
        'methods' => 'POST',
        'callback' => 'mytheme_vue_handle_function_block_action',
        'permission_callback' => 'mytheme_vue_permission_logged_in_with_nonce' // 需要登录 + nonce
    ));

    // ✅ 当前用户订单列表（需要 WooCommerce）
    register_rest_route('mytheme-vue/v1', '/my-orders', array(
        'methods' => 'GET',
        'callback' => function($request) {
            $orders = apply_filters('tanzanite_my_orders', null, $request);
            if (is_wp_error($orders)) {
                return $orders;
            }
            if (is_array($orders)) {
                return $orders;
            }
            return array();
        },
        'permission_callback' => 'mytheme_vue_permission_logged_in_with_nonce'
    ));
}
add_action('rest_api_init', 'mytheme_vue_register_api');

// ===== 自定义登录/登出 API =====
function mytheme_vue_register_auth_api() {
    // 登录端点
    register_rest_route('custom/v1', '/login', array(
        'methods'  => 'POST',
        'callback' => 'mytheme_vue_handle_login',
        'permission_callback' => 'mytheme_vue_permission_with_nonce', // 添加 nonce 验证
    ));

    // 登出端点
    register_rest_route('custom/v1', '/logout', array(
        'methods'  => 'POST',
        'callback' => 'mytheme_vue_handle_logout',
        'permission_callback' => 'mytheme_vue_permission_with_nonce', // 添加 nonce 验证
    ));

    // 注册端点（允许公开注册）
    register_rest_route('custom/v1', '/register', array(
        'methods'  => 'POST',
        'callback' => 'mytheme_vue_handle_register',
        'permission_callback' => 'mytheme_vue_permission_with_nonce', // 添加 nonce 验证
    ));
}
add_action('rest_api_init', 'mytheme_vue_register_auth_api');

// 处理登录
function mytheme_vue_handle_login($request) {
    $username = sanitize_text_field($request->get_param('username'));
    $password = $request->get_param('password');

    if (empty($username) || empty($password)) {
        return new WP_Error('missing_fields', 'Username and password are required', array('status' => 400));
    }

    // 尝试登录
    $user = wp_authenticate($username, $password);

    if (is_wp_error($user)) {
        return new WP_Error('login_failed', 'Invalid username or password', array('status' => 401));
    }

    // 设置登录 cookie
    wp_set_current_user($user->ID);
    wp_set_auth_cookie($user->ID, true);

    // 获取用户等级和积分
    $level = get_user_meta($user->ID, 'mytheme_level', true) ?: '';
    $points = (int) get_user_meta($user->ID, 'mytheme_points', true);

    return rest_ensure_response(array(
        'success' => true,
        'user_id' => $user->ID,
        'username' => $user->user_login,
        'email' => $user->user_email,
        'level' => $level,
        'points' => $points,
    ));
}

// 处理登出
function mytheme_vue_handle_logout() {
    wp_logout();
    return rest_ensure_response(array('success' => true));
}

// 处理注册
function mytheme_vue_handle_register($request) {
    $username = sanitize_text_field($request->get_param('username'));
    $email = sanitize_email($request->get_param('email'));
    $password = $request->get_param('password');

    if (empty($username) || empty($email) || empty($password)) {
        return new WP_Error('missing_fields', 'All fields are required', array('status' => 400));
    }

    // 检查用户名是否已存在
    if (username_exists($username)) {
        return new WP_Error('username_exists', 'Username already exists', array('status' => 400));
    }

    // 检查邮箱是否已存在
    if (email_exists($email)) {
        return new WP_Error('email_exists', 'Email already exists', array('status' => 400));
    }

    // 创建用户
    $user_id = wp_create_user($username, $password, $email);

    if (is_wp_error($user_id)) {
        return new WP_Error('registration_failed', $user_id->get_error_message(), array('status' => 500));
    }

    // 注意：积分和等级的初始化由 mytheme_vue_loyalty_on_register 钩子自动处理
    // 无需在此重复执行

    return rest_ensure_response(array(
        'success' => true,
        'user_id' => $user_id,
        'message' => 'Registration successful',
    ));
}


// 搜索内容API
function mytheme_vue_search_content($request) {
    $query = $request['q'];
    $type = isset($request['type']) ? $request['type'] : 'all';
    
    $args = array(
        's' => $query,
        'post_status' => 'publish',
        'posts_per_page' => 20
    );
    
    if ($type !== 'all') {
        $args['post_type'] = $type;
    }
    
    $search_query = new WP_Query($args);
    $results = array();
    
    if ($search_query->have_posts()) {
        while ($search_query->have_posts()) {
            $search_query->the_post();
            $results[] = array(
                'id' => get_the_ID(),
                'title' => get_the_title(),
                'excerpt' => get_the_excerpt(),
                'url' => get_permalink(),
                'type' => get_post_type(),
                'date' => get_the_date(),
                'thumbnail' => get_the_post_thumbnail_url(get_the_ID(), 'thumbnail')
            );
        }
    }
    wp_reset_postdata();
    return $results;
}

/**
 * 获取主题设置便捷函数
 */
function mytheme_vue_get_setting($setting_name) {
    return get_theme_mod($setting_name, '');
}


/**
 * 获取容器设置
 */
function mytheme_vue_get_container_settings() {
    return array(
        'min' => mytheme_vue_get_setting('min_container_content') ?: 'none',
        'min2' => mytheme_vue_get_setting('min2_container_content') ?: 'page_content',
        'min3' => mytheme_vue_get_setting('min3_container_content') ?: 'none',
        'visibility' => mytheme_vue_get_setting('container_page_visibility') ?: 'all',
    );
}

/**
 * 获取当前页面信息
 */
function mytheme_vue_get_current_page_info() {
    return array(
        'is_home' => is_home(),
        'is_front_page' => is_front_page(),
        'is_single' => is_single(),
        'is_page' => is_page(),
        'is_archive' => is_archive(),
        'post_id' => get_the_ID(),
        'post_type' => get_post_type(),
        'title' => get_the_title(),
        'url' => get_permalink(),
    );
}

/**
 * 根据菜单位置返回树状菜单结构
 */
function mytheme_vue_get_menu_location($location = 'primary') {
    $locations = get_nav_menu_locations();
    if (empty($locations[$location])) {
        return array();
    }
    $items = wp_get_nav_menu_items($locations[$location]);
    if (!$items) return array();

    $tree = array();
    $map = array();
    foreach ($items as $item) {
        $map[$item->ID] = array(
            'id' => $item->ID,
            'title' => $item->title,
            'url' => $item->url,
            'parent' => (int)$item->menu_item_parent,
            'children' => array(),
        );
    }
    foreach ($map as $id => &$node) {
        if ($node['parent'] === 0) {
            $tree[$id] = &$node;
        } elseif (isset($map[$node['parent']])) {
            $map[$node['parent']]['children'][] = &$node;
        }
    }
    return array_values($tree);
}


/**
 * 主题定制器设置（唯一定义）
 */
function mytheme_vue_customize_register($wp_customize) {
    // ================= Quick Buy 设置（使用 WooCommerce product_cat） =================
    if (taxonomy_exists('product_cat')) {
        $wp_customize->add_section('quickbuy_settings', array(
            'title' => __('Quick Buy 设置', 'mytheme-vue'),
            'priority' => 36,
        ));

        // 获取可选分类
        $cats = get_terms(array(
            'taxonomy' => 'product_cat',
            'hide_empty' => false,
            'orderby' => 'name',
            'order' => 'ASC',
            'number' => 200,
        ));
        $cat_choices = array( 0 => __('不限分类', 'mytheme-vue') );
        if (!is_wp_error($cats)) {
            foreach ($cats as $cat) { $cat_choices[$cat->term_id] = $cat->name; }
        }

        for ($i=1; $i<=3; $i++) {
            $setting_key = 'quickbuy_step'.$i.'_cat';
            $label = sprintf(__('步骤 %d 分类（单选）', 'mytheme-vue'), $i);
            $wp_customize->add_setting($setting_key, array(
                'default' => 0,
                'sanitize_callback' => 'absint',
            ));
            $wp_customize->add_control($setting_key, array(
                'label' => $label,
                'section' => 'quickbuy_settings',
                'type' => 'select',
                'choices' => $cat_choices,
            ));

            // 可选：底部提示文案
            $footer_key = 'quickbuy_step'.$i.'_footer';
            $footer_label = sprintf(__('步骤 %d 底部提示文案', 'mytheme-vue'), $i);
            $wp_customize->add_setting($footer_key, array(
                'default' => '',
                'sanitize_callback' => 'sanitize_text_field',
            ));
            $wp_customize->add_control($footer_key, array(
                'label' => $footer_label,
                'section' => 'quickbuy_settings',
                'type' => 'text',
            ));
        }
    }

    
}
add_action('customize_register', 'mytheme_vue_customize_register');

/**
 * 仅在前端禁用不需要的 REST API 端点以提升性能
 * 后台定制器仍可正常使用小工具功能
 */
function mytheme_vue_disable_rest_endpoints($endpoints) {
    // 只在前端禁用，后台和定制器不受影响
    if (!is_admin() && !is_customize_preview()) {
        // 禁用小工具类型端点（前端不需要）
        if (isset($endpoints['/wp/v2/widget-types'])) {
            unset($endpoints['/wp/v2/widget-types']);
        }
        if (isset($endpoints['/wp/v2/widget-types/(?P<id>[a-zA-Z0-9_-]+)'])) {
            unset($endpoints['/wp/v2/widget-types/(?P<id>[a-zA-Z0-9_-]+)']);
        }
        
        // 禁用小工具端点（前端不需要）
        if (isset($endpoints['/wp/v2/widgets'])) {
            unset($endpoints['/wp/v2/widgets']);
        }
        if (isset($endpoints['/wp/v2/widgets/(?P<id>[\d]+)'])) {
            unset($endpoints['/wp/v2/widgets/(?P<id>[\d]+)']);
        }
    }
    
    return $endpoints;
}
add_filter('rest_endpoints', 'mytheme_vue_disable_rest_endpoints');

/**
 * 禁用 WooCommerce 地址自动完成功能（如果文件缺失）
 */
function mytheme_vue_disable_wc_address_autocomplete() {
    wp_dequeue_script('a8c-address-autocomplete-service');
}
add_action('wp_enqueue_scripts', 'mytheme_vue_disable_wc_address_autocomplete', 100);

/**
 * 阻止前端加载 api-fetch 脚本，避免不必要的 REST API 请求
 */
function mytheme_vue_disable_api_fetch() {
    // 只在前端禁用，后台和定制器保留
    if (!is_admin() && !is_customize_preview()) {
        wp_dequeue_script('wp-api-fetch');
        wp_deregister_script('wp-api-fetch');
    }
}
add_action('wp_enqueue_scripts', 'mytheme_vue_disable_api_fetch', 100);

/**
 * 使用会员等级徽章替代 Gravatar 头像，提升加载速度并显示会员等级
 */
function mytheme_vue_replace_gravatar_with_badge($avatar, $id_or_email, $size, $default, $alt) {
    // 获取用户 ID
    $user_id = 0;
    if (is_numeric($id_or_email)) {
        $user_id = (int) $id_or_email;
    } elseif (is_object($id_or_email) && isset($id_or_email->user_id)) {
        $user_id = (int) $id_or_email->user_id;
    } elseif (is_string($id_or_email)) {
        $user = get_user_by('email', $id_or_email);
        if ($user) $user_id = $user->ID;
    }
    
    // 获取用户会员等级
    $level_name = '普通会员'; // 默认等级
    if ($user_id > 0) {
        $points = (int) get_user_meta($user_id, 'loyalty_points', true);
        $config = mytheme_vue_default_loyalty_config();
        foreach ($config['tiers'] as $tier) {
            $min = (int) $tier['min'];
            $max = (int) $tier['max'];
            if ($max === -1) {
                if ($points >= $min) {
                    $level_name = $tier['name'];
                    break;
                }
            } else {
                if ($points >= $min && $points <= $max) {
                    $level_name = $tier['name'];
                    break;
                }
            }
        }
    }
    
    // 根据等级选择徽章 emoji（使用 Unicode 字符，无需外部请求）
    $badge_map = array(
        '普通会员' => '⚪', // 白色圆圈
        '铜牌会员' => '🥉', // 铜牌
        '银牌会员' => '🥈', // 银牌
        '金牌会员' => '🥇', // 金牌
        '尊享会员' => '👑', // 皇冠
    );
    
    $badge = isset($badge_map[$level_name]) ? $badge_map[$level_name] : '⚪';
    
    // 创建徽章 SVG 图标（内联，无需外部请求）
    $badge_svg = '<svg width="' . $size . '" height="' . $size . '" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
        <circle cx="50" cy="50" r="48" fill="#f0f0f0" stroke="#ddd" stroke-width="2"/>
        <text x="50" y="50" font-size="50" text-anchor="middle" dominant-baseline="central">' . $badge . '</text>
    </svg>';
    
    // 转换为 data URI
    $badge_data_uri = 'data:image/svg+xml;base64,' . base64_encode($badge_svg);
    
    // 替换头像为徽章
    $avatar = '<img alt="' . esc_attr($alt) . '" src="' . $badge_data_uri . '" class="avatar avatar-' . $size . ' photo" height="' . $size . '" width="' . $size . '" loading="lazy" decoding="async" />';
    
    return $avatar;
}
add_filter('get_avatar', 'mytheme_vue_replace_gravatar_with_badge', 10, 5);

/**
 * 会员积分与等级系统
 */

// 获取默认会员等级配置
function mytheme_vue_default_loyalty_config() {
    return array(
        'tiers' => array(
            array('name' => '普通会员', 'min' => 0, 'max' => 500, 'discount' => 0),
            array('name' => '铜牌会员', 'min' => 501, 'max' => 2000, 'discount' => 0),
            array('name' => '银牌会员', 'min' => 2001, 'max' => 6000, 'discount' => 10),
            array('name' => '金牌会员', 'min' => 6001, 'max' => 15000, 'discount' => 15),
            array('name' => '至尊会员', 'min' => 15001, 'max' => -1, 'discount' => 20),
        ),
        'points_per_unit' => 1, // 每 1 货币单位积 1 分
        'enabled' => true,
        'apply_cart_discount' => true
    );
}

// 读取后台配置（自定义器中保存为 JSON）
function mytheme_vue_get_loyalty_config() {
    if (function_exists('mytheme_member_profiles_get_loyalty_settings')) {
        $settings = mytheme_member_profiles_get_loyalty_settings();
        if (is_array($settings) && !empty($settings)) {
            return wp_parse_args($settings, mytheme_vue_default_loyalty_config());
        }
    }

    $json = get_option('mytheme_member_loyalty_config', '');
    if (!$json) {
        return mytheme_vue_default_loyalty_config();
    }
    $data = json_decode($json, true);
    if (json_last_error() !== JSON_ERROR_NONE || !is_array($data)) {
        return mytheme_vue_default_loyalty_config();
    }
    return wp_parse_args($data, mytheme_vue_default_loyalty_config());
}

// 根据积分计算等级信息
function mytheme_vue_loyalty_calculate_level($points) {
    $conf = mytheme_vue_get_loyalty_config();
    foreach ($conf['tiers'] as $tier) {
        $min = (int)$tier['min'];
        $max = (int)$tier['max'];
        if ($max === -1) {
            if ($points >= $min) return $tier;
        } else {
            if ($points >= $min && $points <= $max) return $tier;
        }
    }
    // 默认返回最低等级
    return $conf['tiers'][0];
}

// 注册时初始化用户积分与等级
function mytheme_vue_loyalty_on_register($user_id) {
    if (function_exists('mytheme_member_profiles_initialize_loyalty')) {
        mytheme_member_profiles_initialize_loyalty($user_id);
    } else {
        add_user_meta($user_id, 'mytheme_points', 0, true);
        $tier = mytheme_vue_loyalty_calculate_level(0);
        add_user_meta($user_id, 'mytheme_level', sanitize_text_field($tier['name']), true);
    }
}
add_action('user_register', 'mytheme_vue_loyalty_on_register');


// WooCommerce: 订单完成后累计积分并更新等级
function mytheme_vue_loyalty_order_completed($order_id) {
    if (!function_exists('wc_get_order')) return;
    $conf = mytheme_vue_get_loyalty_config();
    if (empty($conf['enabled'])) return;

    $order = wc_get_order($order_id);
    if (!$order) return;
    $user_id = $order->get_user_id();
    if (!$user_id) return;

    $total = (float) $order->get_total();
    $rate = (float) ($conf['points_per_unit'] ?: 1);
    $earned = (int) floor($total * $rate);

    $current = (int) get_user_meta($user_id, 'mytheme_points', true);
    $new_points = max(0, $current + $earned);
    update_user_meta($user_id, 'mytheme_points', $new_points);

    $tier = mytheme_vue_loyalty_calculate_level($new_points);
    update_user_meta($user_id, 'mytheme_level', sanitize_text_field($tier['name']));
}
add_action('woocommerce_order_status_completed', 'mytheme_vue_loyalty_order_completed');

// WooCommerce: 购物车应用等级折扣（按配置百分比）
function mytheme_vue_loyalty_extract_ids($items) {
    if (!is_array($items)) return array();
    return array_values(array_filter(array_map(function($item) {
        if (is_array($item) && isset($item['id'])) {
            return (int) $item['id'];
        }
        return (int) $item;
    }, $items), function($id) {
        return $id > 0;
    }));
}

function mytheme_vue_loyalty_tier_matches_cart_item($tier, $cart_item) {
    $product_ids = mytheme_vue_loyalty_extract_ids($tier['products'] ?? array());
    $category_ids = mytheme_vue_loyalty_extract_ids($tier['categories'] ?? array());

    $product_id = isset($cart_item['product_id']) ? (int) $cart_item['product_id'] : 0;
    $variation_id = isset($cart_item['variation_id']) ? (int) $cart_item['variation_id'] : 0;

    $matches_products = false;
    if (!empty($product_ids)) {
        if ($product_id && in_array($product_id, $product_ids, true)) {
            $matches_products = true;
        }
        if (!$matches_products && $variation_id && in_array($variation_id, $product_ids, true)) {
            $matches_products = true;
        }
    }

    $matches_categories = false;
    if (!empty($category_ids)) {
        $cat_ids = array();
        if ($variation_id) {
            $variation_product = wc_get_product($variation_id);
            if ($variation_product && method_exists($variation_product, 'get_category_ids')) {
                $cat_ids = array_map('intval', (array) $variation_product->get_category_ids());
            }
        }
        if (!$variation_id || empty($cat_ids)) {
            $product = wc_get_product($product_id);
            if ($product && method_exists($product, 'get_category_ids')) {
                $cat_ids = array_map('intval', (array) $product->get_category_ids());
            }
        }
        if (!empty($cat_ids)) {
            $matches_categories = (bool) array_intersect($cat_ids, $category_ids);
        }
    }

    if (empty($product_ids) && empty($category_ids)) {
        return true;
    }

    if (!empty($product_ids) && $matches_products) {
        return true;
    }

    if (!empty($category_ids) && $matches_categories) {
        return true;
    }

    return false;
}

function mytheme_vue_loyalty_apply_cart_discount($cart) {
    if (is_admin() && !defined('DOING_AJAX')) return;
    if (!is_user_logged_in()) return;

    $conf = mytheme_vue_get_loyalty_config();
    if (empty($conf['enabled']) || empty($conf['apply_cart_discount'])) return;

    $user_id = get_current_user_id();
    $points = (int) get_user_meta($user_id, 'mytheme_points', true);
    $tier = mytheme_vue_loyalty_calculate_level($points);
    $discount_pct = isset($tier['discount']) ? (float) $tier['discount'] : 0;
    if ($discount_pct <= 0) return;

    $eligible_total = 0.0;
    foreach ($cart->get_cart() as $cart_item) {
        if (mytheme_vue_loyalty_tier_matches_cart_item($tier, $cart_item)) {
            $line_total = isset($cart_item['line_subtotal']) ? (float) $cart_item['line_subtotal'] : 0.0;
            $eligible_total += $line_total;
        }
    }

    if ($eligible_total <= 0) return;

    $label = sprintf(__('会员等级折扣（%s - %s%%）', 'mytheme-vue'), $tier['name'], $discount_pct);
    $amount = - ($eligible_total * ($discount_pct / 100.0));

    if ($amount != 0) {
        $cart->add_fee($label, $amount, true);
    }
}
add_action('woocommerce_cart_calculate_fees', 'mytheme_vue_loyalty_apply_cart_discount', 20, 1);


// WooCommerce: 购物车自动使用最大可抵扣积分（按折扣前小计×百分比上限）
function mytheme_vue_loyalty_apply_points_redeem($cart) {
    if (is_admin() && !defined('DOING_AJAX')) return;
    if (!is_user_logged_in()) return;

    $conf = mytheme_vue_get_loyalty_config();
    if (empty($conf['enabled'])) return;

    $user_id = get_current_user_id();
    // 用户可用积分
    $user_points = (int) get_user_meta($user_id, 'loyalty_points', true);
    if ($user_points === 0) {
        $user_points = (int) get_user_meta($user_id, 'mytheme_points', true);
    }
    if ($user_points <= 0) {
        if (function_exists('WC') && WC()->session) {
            WC()->session->set('mytheme_points_redeemed', 0);
        }
        return;
    }

    // 计算当前会员等级与其抵扣设置
    $tier = mytheme_vue_loyalty_calculate_level($user_points);
    $redeem = isset($tier['redeem']) && is_array($tier['redeem']) ? $tier['redeem'] : array();
    if (empty($redeem['enabled'])) {
        if (function_exists('WC') && WC()->session) {
            WC()->session->set('mytheme_points_redeemed', 0);
        }
        return;
    }

    $percent_cap = isset($redeem['percent_of_total']) ? max(0, (float)$redeem['percent_of_total']) : 0.0;
    $value_per_point = isset($redeem['value_per_point_base']) ? max(0.0, (float)$redeem['value_per_point_base']) : 0.0;
    $min_points = isset($redeem['min_points']) ? max(0, (int)$redeem['min_points']) : 0;
    $stack_with_percent = isset($redeem['stack_with_percent']) ? (bool)$redeem['stack_with_percent'] : true;

    if ($percent_cap <= 0 || $value_per_point <= 0.0) {
        if (function_exists('WC') && WC()->session) {
            WC()->session->set('mytheme_points_redeemed', 0);
        }
        return;
    }

    // 折扣前小计：若有范围则仅统计匹配的行，否则用全购物车小计
    $pre_discount_subtotal = 0.0;
    $has_scope = !empty($tier['products']) || !empty($tier['categories']);
    if ($has_scope) {
        foreach ($cart->get_cart() as $cart_item) {
            if (mytheme_vue_loyalty_tier_matches_cart_item($tier, $cart_item)) {
                $pre_discount_subtotal += isset($cart_item['line_subtotal']) ? (float)$cart_item['line_subtotal'] : 0.0;
            }
        }
    } else {
        $pre_discount_subtotal = (float) $cart->get_subtotal();
    }

    if ($pre_discount_subtotal <= 0) {
        if (function_exists('WC') && WC()->session) {
            WC()->session->set('mytheme_points_redeemed', 0);
        }
        return;
    }

    // 上限金额 = 折扣前小计 × 百分比
    $cap_amount = $pre_discount_subtotal * ($percent_cap / 100.0);
    // 可兑换金额 = 用户可用积分 × 每积分面值（此处按 1:1 视为订单货币，若站点有多货币插件可通过钩子覆盖）
    $exchangeable_amount = $user_points * $value_per_point;

    // 实际抵扣金额
    $redeem_amount = min($cap_amount, $exchangeable_amount);

    // 计算本次需要消耗的积分（向下取整，确保不超出可用积分）
    $points_to_use = (int) floor($redeem_amount / max(0.0000001, $value_per_point));
    if ($points_to_use < $min_points) {
        if (function_exists('WC') && WC()->session) {
            WC()->session->set('mytheme_points_redeemed', 0);
        }
        return;
    }
    $points_to_use = min($points_to_use, $user_points);
    $redeem_amount = $points_to_use * $value_per_point;
    if ($redeem_amount <= 0) {
        if (function_exists('WC') && WC()->session) {
            WC()->session->set('mytheme_points_redeemed', 0);
        }
        return;
    }

    // 将抵扣金额作为负费用加入
    $label = sprintf(__('积分抵扣（最多 %s%% 上限）', 'mytheme-vue'), $percent_cap);
    $cart->add_fee($label, -$redeem_amount, true);

    // 缓存将要消耗的积分，供创建订单与状态变更时使用
    if (function_exists('WC') && WC()->session) {
        WC()->session->set('mytheme_points_redeemed', $points_to_use);
        WC()->session->set('mytheme_points_redeem_amount', $redeem_amount);
    }
}
add_action('woocommerce_cart_calculate_fees', 'mytheme_vue_loyalty_apply_points_redeem', 25, 1);


// 结账创建订单时写入本单抵扣的积分与金额
function mytheme_vue_loyalty_attach_redeem_meta($order, $data) {
    if (!function_exists('WC') || !WC()->session) return;
    $points = (int) WC()->session->get('mytheme_points_redeemed');
    $amount = (float) WC()->session->get('mytheme_points_redeem_amount');
    if ($points > 0 && $amount > 0) {
        $order->update_meta_data('_mytheme_points_redeemed', $points);
        $order->update_meta_data('_mytheme_points_redeem_amount', $amount);
        $order->save();
    }
}
add_action('woocommerce_checkout_create_order', 'mytheme_vue_loyalty_attach_redeem_meta', 20, 2);


// 订单到 processing/complete 扣减积分
function mytheme_vue_loyalty_deduct_points_on_status($order_id) {
    $order = wc_get_order($order_id);
    if (!$order) return;
    $user_id = $order->get_user_id();
    if (!$user_id) return;

    $already_deducted = (int) $order->get_meta('_mytheme_points_redeemed_deducted');
    if ($already_deducted) return;

    $points = (int) $order->get_meta('_mytheme_points_redeemed');
    if ($points <= 0) return;

    $current = (int) get_user_meta($user_id, 'loyalty_points', true);
    if ($current === 0) {
        $current = (int) get_user_meta($user_id, 'mytheme_points', true);
    }
    $new_points = max(0, $current - $points);
    update_user_meta($user_id, 'loyalty_points', $new_points);
    update_user_meta($user_id, 'mytheme_points', $new_points);

    $order->update_meta_data('_mytheme_points_redeemed_deducted', 1);
    $order->save();
}
add_action('woocommerce_order_status_processing', 'mytheme_vue_loyalty_deduct_points_on_status');
add_action('woocommerce_order_status_completed', 'mytheme_vue_loyalty_deduct_points_on_status');

// AJAX: 在弹窗内登录（不跳转）
function mytheme_vue_fb_login() {
    check_ajax_referer('mytheme_fb', 'nonce');
    if (is_user_logged_in()) {
        wp_send_json_success(array('message' => __('已登录', 'mytheme-vue')));
    }
    $username = isset($_POST['username']) ? sanitize_text_field(wp_unslash($_POST['username'])) : '';
    $password = isset($_POST['password']) ? $_POST['password'] : '';
    $remember = !empty($_POST['remember']);

    if ($username === '' || $password === '') {
        wp_send_json_error(array('message' => __('请填写用户名和密码', 'mytheme-vue')));
    }

    $creds = array(
        'user_login' => $username,
        'user_password' => $password,
        'remember' => $remember,
    );
    $user = wp_signon($creds, is_ssl());
    if (is_wp_error($user)) {
        wp_send_json_error(array('message' => $user->get_error_message()));
    }
    wp_send_json_success(array('message' => __('登录成功', 'mytheme-vue')));
}
add_action('wp_ajax_nopriv_mytheme_fb_login', 'mytheme_vue_fb_login');
add_action('wp_ajax_mytheme_fb_login', 'mytheme_vue_fb_login');

/**
 * Sidebar Filters - Dropdown Select Interaction
 * 侧边栏筛选条件下拉选择框交互逻辑
 */
function mytheme_vue_sidebar_filters_script() {
    ?>
    <script>
    (function() {
      'use strict';
      
      document.addEventListener('DOMContentLoaded', function() {
        const filterGroups = document.querySelectorAll('.sidebar-left .filter-group');
        if (!filterGroups.length) return;
        
        // 下拉展开/收起
        document.querySelectorAll('.sidebar-left .filter-select-header').forEach(function(header) {
          header.addEventListener('click', function(e) {
            e.stopPropagation();
            const group = this.closest('.filter-group');
            const options = group.querySelector('.filter-options');
            const isActive = this.classList.contains('active');
            
            // 关闭所有其他下拉框
            document.querySelectorAll('.sidebar-left .filter-select-header').forEach(function(h) {
              h.classList.remove('active');
            });
            document.querySelectorAll('.sidebar-left .filter-options').forEach(function(o) {
              o.classList.remove('show');
            });
            
            // 切换当前下拉框
            if (!isActive) {
              this.classList.add('active');
              options.classList.add('show');
            }
          });
        });
        
        // 点击页面其他地方关闭下拉框
        document.addEventListener('click', function() {
          document.querySelectorAll('.sidebar-left .filter-select-header').forEach(function(h) {
            h.classList.remove('active');
          });
          document.querySelectorAll('.sidebar-left .filter-options').forEach(function(o) {
            o.classList.remove('show');
          });
        });
        
        // 阻止点击选项时关闭下拉框
        document.querySelectorAll('.sidebar-left .filter-options').forEach(function(options) {
          options.addEventListener('click', function(e) {
            e.stopPropagation();
          });
        });
        
        // All 互斥逻辑
        filterGroups.forEach(function(group) {
          const checkboxes = group.querySelectorAll('input[type="checkbox"]');
          const allCheckbox = group.querySelector('input[value="all"]');
          const otherCheckboxes = Array.from(checkboxes).filter(function(cb) { return cb.value !== 'all'; });
          const header = group.querySelector('.filter-select-header');
          const selectedText = header.querySelector('.filter-selected-text');
          
          // 点击 All 时，取消其他选项
          if (allCheckbox) {
            allCheckbox.addEventListener('change', function() {
              if (this.checked) {
                otherCheckboxes.forEach(function(cb) { cb.checked = false; });
              }
              updateGroupDisplay(group, selectedText);
            });
          }
          
          // 点击其他选项时，取消 All
          otherCheckboxes.forEach(function(checkbox) {
            checkbox.addEventListener('change', function() {
              if (this.checked && allCheckbox) {
                allCheckbox.checked = false;
              }
              
              // 如果所有其他选项都未选中，自动勾选 All
              const anyChecked = otherCheckboxes.some(function(cb) { return cb.checked; });
              if (!anyChecked && allCheckbox) {
                allCheckbox.checked = true;
              }
              
              updateGroupDisplay(group, selectedText);
            });
          });
        });
        
        // 更新单个筛选组的显示文本
        function updateGroupDisplay(group, selectedText) {
          const checkedBoxes = group.querySelectorAll('input[type="checkbox"]:checked');
          const checkedValues = Array.from(checkedBoxes).map(function(cb) {
            return cb.closest('.filter-option').querySelector('span').textContent;
          });
          
          if (checkedValues.length === 0 || (checkedValues.length === 1 && checkedValues[0] === 'All')) {
            selectedText.textContent = 'All';
          } else {
            selectedText.textContent = checkedValues.join(', ');
          }
        }
        
        // 初始化显示
        filterGroups.forEach(function(group) {
          const selectedText = group.querySelector('.filter-selected-text');
          updateGroupDisplay(group, selectedText);
        });
      });
    })();
    </script>
    <?php
}
add_action('wp_footer', 'mytheme_vue_sidebar_filters_script');

/**
 * 添加产品筛选器的结构化数据（Schema.org）用于 SEO
 */
function mytheme_add_filter_schema() {
  if (is_singular('product') || is_post_type_archive('product') || is_tax('product_cat')) {
    ?>
    <script type="application/ld+json">
    {
      "@context": "https://schema.org",
      "@type": "ItemList",
      "name": "Product Filters",
      "description": "Filter bicycle wheels by specifications",
      "itemListElement": [
        {
          "@type": "ListItem",
          "position": 1,
          "name": "Wheel Diameter",
          "item": {
            "@type": "PropertyValue",
            "name": "wheel_diameter",
            "value": ["12 inch", "406mm", "451mm", "26 inch", "27.5 inch", "650B", "29 inch", "700C"]
          }
        },
        {
          "@type": "ListItem",
          "position": 2,
          "name": "Price Range",
          "item": {
            "@type": "PropertyValue",
            "name": "price",
            "value": ["Under 200 USD", "200-500 USD", "500 USD or More"]
          }
        },
        {
          "@type": "ListItem",
          "position": 3,
          "name": "Brand",
          "item": {
            "@type": "PropertyValue",
            "name": "brand",
            "value": ["Tanzanite", "Sapim", "DT Swiss", "Pillar"]
          }
        },
        {
          "@type": "ListItem",
          "position": 4,
          "name": "Brake Style",
          "item": {
            "@type": "PropertyValue",
            "name": "brake_style",
            "value": ["V-Brake", "Center-lock", "6-Bolt"]
          }
        }
      ]
    }
    </script>
    <?php
  }
}
add_action('wp_head', 'mytheme_add_filter_schema');

/**
 * 添加产品页面的 Meta 描述用于 SEO
 */
function mytheme_filter_meta_description() {
  if (is_post_type_archive('product')) {
    echo '<meta name="description" content="Browse our premium bicycle wheels. Filter by wheel diameter (700C, 29\", 650B, 26\"), price range, brand (Tanzanite, Sapim, DT Swiss, Pillar), and brake style (V-Brake, Center-lock, 6-Bolt). Find the perfect wheels for your bike.">';
  } elseif (is_tax('product_cat')) {
    $term = get_queried_object();
    echo '<meta name="description" content="' . esc_attr($term->name) . ' bicycle wheels. Filter by diameter, price, brand, and brake style to find your ideal wheels.">';
  }
}
add_action('wp_head', 'mytheme_filter_meta_description');

// 前台禁用小工具编辑相关脚本，避免前台触发 /wp-json/wp/v2/widget-types 请求
add_action('wp_enqueue_scripts', function(){
  if (is_admin() || is_customize_preview()) return;
  // 部分版本句柄名称可能不同，尽量多做兼容
  $handles = array(
    'customize-widgets',
    'wp-edit-widgets',
    'wp-customize-widgets',
    'wp-widgets',
    'widgets',
  );
  foreach ($handles as $h) {
    wp_dequeue_script($h);
    wp_deregister_script($h);
  }
}, 20);

/**
 * 输出多语言 hreflang 与 canonical
 * - 默认语言 en 不加前缀
 * - 其它语言加前缀（如 /fr, /de, /zh ...）
 */
function mytheme_output_hreflang_and_canonical(){
  if (is_admin()) return;
  // 与 nuxt-i18n-widget 保持一致的语言清单（可按需调整顺序/内容）
  $locales = array('en','fr','de','es','it','pt','ru','ja','ko','ar','fi','da','th','zh');
  $default = 'en';

  $scheme   = is_ssl() ? 'https://' : 'http://';
  $host     = $_SERVER['HTTP_HOST'] ?? parse_url(home_url('/'), PHP_URL_HOST);
  $req_path = $_SERVER['REQUEST_URI'] ?? '/';

  // 去除查询与 hash 仅保留路径
  $path_only = strtok($req_path, '?');
  $path_only = $path_only ?: '/';

  // 清理双斜杠并标准化
  $path_only = '/' . ltrim($path_only, '/');

  // 去除现有语言前缀
  $segments = explode('/', ltrim($path_only, '/'));
  if (!empty($segments)){
    $first = strtolower($segments[0]);
    if (in_array($first, $locales, true)) {
      array_shift($segments);
    }
  }
  $clean_path = '/' . implode('/', array_filter($segments, 'strlen'));
  if ($clean_path === '//') $clean_path = '/';

  // 生成各语言 URL
  $urls = array();
  foreach ($locales as $code){
    $code = strtolower($code);
    $url_path = ($code === $default) ? $clean_path : ('/' . $code . $clean_path);
    // 确保单根路径格式
    if ($url_path !== '/' && substr($url_path, 0, 1) !== '/') $url_path = '/' . $url_path;
    $urls[$code] = $scheme . $host . $url_path;
  }

  // 当前语言（根据路径判断）
  $current_lang = $default;
  $cur_first = strtolower(explode('/', ltrim($path_only, '/'))[0] ?? '');
  if (in_array($cur_first, $locales, true)) $current_lang = $cur_first;

  // 输出 canonical（当前语言对应 URL）
  $canonical = $urls[$current_lang] ?? ($scheme . $host . $path_only);
  echo "\n<!-- i18n SEO: hreflang/canonical -->\n";
  echo '<link rel="canonical" href="' . esc_url($canonical) . '" />' . "\n";

  // 输出 hreflang（含 x-default 映射到默认语言）
  foreach ($urls as $code => $href){
    echo '<link rel="alternate" hreflang="' . esc_attr($code) . '" href="' . esc_url($href) . '" />' . "\n";
  }
  // x-default 指向默认语言
  echo '<link rel="alternate" hreflang="x-default" href="' . esc_url($urls[$default]) . '" />' . "\n";
}
add_action('wp_head', 'mytheme_output_hreflang_and_canonical', 5);

// 语言切换器样式已统一到 header.css；不再单独队列 language-switcher.css

// ===== 聊天功能 REST API（用于移动端 APP 和 Web Sidebar）=====

/**
 * 创建聊天数据表
 * 在主题激活时自动创建
 */
function mytheme_chat_create_tables() {
    global $wpdb;
    $charset_collate = $wpdb->get_charset_collate();
    $table_name = $wpdb->prefix . 'mytheme_chat_messages';
    
    $sql = "CREATE TABLE IF NOT EXISTS $table_name (
        id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
        conversation_id VARCHAR(100) NOT NULL,
        sender_id BIGINT(20) UNSIGNED NOT NULL,
        receiver_id BIGINT(20) UNSIGNED NOT NULL,
        sender_role VARCHAR(20) NOT NULL DEFAULT 'customer',
        content TEXT NOT NULL,
        message_type VARCHAR(20) NOT NULL DEFAULT 'text',
        attachment_url VARCHAR(500) DEFAULT NULL,
        read_status TINYINT(1) NOT NULL DEFAULT 0,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (id),
        KEY conversation_id (conversation_id),
        KEY sender_id (sender_id),
        KEY receiver_id (receiver_id),
        KEY created_at (created_at),
        KEY read_status (read_status)
    ) $charset_collate;";
    
    require_once(ABSPATH . 'wp-admin/includes/upgrade.php');
    dbDelta($sql);
    
    // 创建会话表
    $conversations_table = $wpdb->prefix . 'mytheme_chat_conversations';
    $sql_conversations = "CREATE TABLE IF NOT EXISTS $conversations_table (
        id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
        conversation_id VARCHAR(100) NOT NULL UNIQUE,
        customer_id BIGINT(20) UNSIGNED NOT NULL,
        agent_id BIGINT(20) UNSIGNED DEFAULT NULL,
        status VARCHAR(20) NOT NULL DEFAULT 'active',
        last_message_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        unread_count_customer INT NOT NULL DEFAULT 0,
        unread_count_agent INT NOT NULL DEFAULT 0,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (id),
        UNIQUE KEY conversation_id (conversation_id),
        KEY customer_id (customer_id),
        KEY agent_id (agent_id),
        KEY status (status),
        KEY last_message_at (last_message_at)
    ) $charset_collate;";
    
    dbDelta($sql_conversations);
}
add_action('after_switch_theme', 'mytheme_chat_create_tables');

/**
 * 注册聊天相关的 REST API 端点
 */
function mytheme_chat_register_rest_routes() {
    // 1. 获取会话列表
    register_rest_route('mytheme/v1', '/chat/conversations', array(
        'methods'  => 'GET',
        'callback' => 'mytheme_chat_get_conversations',
        'permission_callback' => 'mytheme_vue_permission_logged_in_with_nonce',
        'args' => array(
            'role' => array(
                'required' => false,
                'default' => 'customer',
                'validate_callback' => function($param) {
                    return in_array($param, array('customer', 'agent'));
                }
            ),
            'page' => array(
                'required' => false,
                'default' => 1,
                'validate_callback' => function($param) {
                    return is_numeric($param) && $param > 0;
                }
            ),
            'per_page' => array(
                'required' => false,
                'default' => 20,
                'validate_callback' => function($param) {
                    return is_numeric($param) && $param > 0 && $param <= 100;
                }
            )
        )
    ));
    
    // 2. 获取会话消息列表
    register_rest_route('mytheme/v1', '/chat/messages/(?P<conversation_id>[a-zA-Z0-9_-]+)', array(
        'methods'  => 'GET',
        'callback' => 'mytheme_chat_get_messages',
        'permission_callback' => 'mytheme_vue_permission_logged_in_with_nonce',
        'args' => array(
            'conversation_id' => array(
                'required' => true,
                'validate_callback' => function($param) {
                    return !empty($param);
                }
            ),
            'page' => array(
                'required' => false,
                'default' => 1
            ),
            'per_page' => array(
                'required' => false,
                'default' => 50
            )
        )
    ));
    
    // 3. 发送消息
    register_rest_route('mytheme/v1', '/chat/send', array(
        'methods'  => 'POST',
        'callback' => 'mytheme_chat_send_message',
        'permission_callback' => 'mytheme_vue_permission_logged_in_with_nonce',
        'args' => array(
            'conversation_id' => array(
                'required' => false,
                'default' => ''
            ),
            'receiver_id' => array(
                'required' => true,
                'validate_callback' => function($param) {
                    return is_numeric($param) && $param > 0;
                }
            ),
            'content' => array(
                'required' => true,
                'sanitize_callback' => 'sanitize_textarea_field'
            ),
            'message_type' => array(
                'required' => false,
                'default' => 'text',
                'validate_callback' => function($param) {
                    return in_array($param, array('text', 'image', 'file'));
                }
            ),
            'attachment_url' => array(
                'required' => false,
                'default' => '',
                'sanitize_callback' => 'esc_url_raw'
            )
        )
    ));
    
    // 4. 标记消息为已读
    register_rest_route('mytheme/v1', '/chat/mark-read/(?P<conversation_id>[a-zA-Z0-9_-]+)', array(
        'methods'  => 'POST',
        'callback' => 'mytheme_chat_mark_read',
        'permission_callback' => 'mytheme_vue_permission_logged_in_with_nonce',
        'args' => array(
            'conversation_id' => array(
                'required' => true
            )
        )
    ));
    
    // 5. 获取未读消息数
    register_rest_route('mytheme/v1', '/chat/unread-count', array(
        'methods'  => 'GET',
        'callback' => 'mytheme_chat_get_unread_count',
        'permission_callback' => 'mytheme_vue_permission_logged_in_with_nonce'
    ));
}
add_action('rest_api_init', 'mytheme_chat_register_rest_routes');

/**
 * 获取会话列表
 */
function mytheme_chat_get_conversations($request) {
    global $wpdb;
    $user_id = get_current_user_id();
    $role = $request->get_param('role');
    $page = intval($request->get_param('page'));
    $per_page = intval($request->get_param('per_page'));
    $offset = ($page - 1) * $per_page;
    
    $conversations_table = $wpdb->prefix . 'mytheme_chat_conversations';
    
    // 根据角色查询不同的会话
    if ($role === 'agent') {
        // 客服端：查询所有分配给自己的会话，或未分配的会话
        $where = $wpdb->prepare(
            "WHERE (agent_id = %d OR agent_id IS NULL) AND status = 'active'",
            $user_id
        );
    } else {
        // 访客端：查询自己发起的会话
        $where = $wpdb->prepare(
            "WHERE customer_id = %d",
            $user_id
        );
    }
    
    $conversations = $wpdb->get_results(
        "SELECT * FROM $conversations_table 
        $where 
        ORDER BY last_message_at DESC 
        LIMIT $offset, $per_page"
    );
    
    // 获取每个会话的最后一条消息
    $messages_table = $wpdb->prefix . 'mytheme_chat_messages';
    foreach ($conversations as &$conversation) {
        $last_message = $wpdb->get_row($wpdb->prepare(
            "SELECT content, message_type, created_at, sender_id 
            FROM $messages_table 
            WHERE conversation_id = %s 
            ORDER BY created_at DESC 
            LIMIT 1",
            $conversation->conversation_id
        ));
        
        $conversation->last_message = $last_message;
        
        // 获取对方用户信息
        $other_user_id = ($role === 'agent') ? $conversation->customer_id : $conversation->agent_id;
        if ($other_user_id) {
            $user_data = get_userdata($other_user_id);
            $conversation->other_user = array(
                'id' => $other_user_id,
                'name' => $user_data->display_name,
                'avatar' => get_avatar_url($other_user_id)
            );
        }
        
        // 未读数
        $conversation->unread_count = ($role === 'agent') 
            ? $conversation->unread_count_agent 
            : $conversation->unread_count_customer;
    }
    
    return rest_ensure_response(array(
        'success' => true,
        'data' => $conversations,
        'pagination' => array(
            'page' => $page,
            'per_page' => $per_page,
            'total' => $wpdb->get_var("SELECT COUNT(*) FROM $conversations_table $where")
        )
    ));
}

/**
 * 获取会话消息列表
 */
function mytheme_chat_get_messages($request) {
    global $wpdb;
    $user_id = get_current_user_id();
    $conversation_id = $request->get_param('conversation_id');
    $page = intval($request->get_param('page'));
    $per_page = intval($request->get_param('per_page'));
    $offset = ($page - 1) * $per_page;
    
    // 验证用户是否有权限访问此会话
    $conversations_table = $wpdb->prefix . 'mytheme_chat_conversations';
    $conversation = $wpdb->get_row($wpdb->prepare(
        "SELECT * FROM $conversations_table WHERE conversation_id = %s",
        $conversation_id
    ));
    
    if (!$conversation || ($conversation->customer_id != $user_id && $conversation->agent_id != $user_id)) {
        return new WP_Error('forbidden', '无权访问此会话', array('status' => 403));
    }
    
    // 获取消息列表
    $messages_table = $wpdb->prefix . 'mytheme_chat_messages';
    $messages = $wpdb->get_results($wpdb->prepare(
        "SELECT * FROM $messages_table 
        WHERE conversation_id = %s 
        ORDER BY created_at DESC 
        LIMIT %d, %d",
        $conversation_id,
        $offset,
        $per_page
    ));
    
    // 添加发送者信息
    foreach ($messages as &$message) {
        $user_data = get_userdata($message->sender_id);
        $message->sender = array(
            'id' => $message->sender_id,
            'name' => $user_data->display_name,
            'avatar' => get_avatar_url($message->sender_id),
            'role' => $message->sender_role
        );
    }
    
    return rest_ensure_response(array(
        'success' => true,
        'data' => array_reverse($messages), // 反转顺序，最旧的在前
        'pagination' => array(
            'page' => $page,
            'per_page' => $per_page,
            'total' => $wpdb->get_var($wpdb->prepare(
                "SELECT COUNT(*) FROM $messages_table WHERE conversation_id = %s",
                $conversation_id
            ))
        )
    ));
}

/**
 * 发送消息
 */
function mytheme_chat_send_message($request) {
    global $wpdb;
    $user_id = get_current_user_id();
    $conversation_id = $request->get_param('conversation_id');
    $receiver_id = intval($request->get_param('receiver_id'));
    $content = $request->get_param('content');
    $message_type = $request->get_param('message_type');
    $attachment_url = $request->get_param('attachment_url');
    
    // 确定发送者角色
    $user_meta = get_user_meta($user_id, 'chat_role', true);
    $sender_role = $user_meta ?: 'customer';
    
    // 如果没有会话 ID，创建新会话
    if (empty($conversation_id)) {
        $conversation_id = 'conv_' . $user_id . '_' . $receiver_id . '_' . time();
        
        $conversations_table = $wpdb->prefix . 'mytheme_chat_conversations';
        $wpdb->insert($conversations_table, array(
            'conversation_id' => $conversation_id,
            'customer_id' => ($sender_role === 'customer') ? $user_id : $receiver_id,
            'agent_id' => ($sender_role === 'agent') ? $user_id : $receiver_id,
            'status' => 'active',
            'last_message_at' => current_time('mysql'),
            'unread_count_customer' => ($sender_role === 'agent') ? 1 : 0,
            'unread_count_agent' => ($sender_role === 'customer') ? 1 : 0
        ));
    }
    
    // 插入消息
    $messages_table = $wpdb->prefix . 'mytheme_chat_messages';
    $result = $wpdb->insert($messages_table, array(
        'conversation_id' => $conversation_id,
        'sender_id' => $user_id,
        'receiver_id' => $receiver_id,
        'sender_role' => $sender_role,
        'content' => $content,
        'message_type' => $message_type,
        'attachment_url' => $attachment_url,
        'read_status' => 0,
        'created_at' => current_time('mysql')
    ));
    
    if ($result === false) {
        return new WP_Error('db_error', '消息发送失败', array('status' => 500));
    }
    
    $message_id = $wpdb->insert_id;
    
    // 更新会话的最后消息时间和未读数
    $conversations_table = $wpdb->prefix . 'mytheme_chat_conversations';
    $unread_field = ($sender_role === 'customer') ? 'unread_count_agent' : 'unread_count_customer';
    $wpdb->query($wpdb->prepare(
        "UPDATE $conversations_table 
        SET last_message_at = %s, 
            $unread_field = $unread_field + 1 
        WHERE conversation_id = %s",
        current_time('mysql'),
        $conversation_id
    ));
    
    // 返回新消息
    $message = $wpdb->get_row($wpdb->prepare(
        "SELECT * FROM $messages_table WHERE id = %d",
        $message_id
    ));
    
    $user_data = get_userdata($user_id);
    $message->sender = array(
        'id' => $user_id,
        'name' => $user_data->display_name,
        'avatar' => get_avatar_url($user_id),
        'role' => $sender_role
    );
    
    return rest_ensure_response(array(
        'success' => true,
        'data' => $message
    ));
}

/**
 * 标记消息为已读
 */
function mytheme_chat_mark_read($request) {
    global $wpdb;
    $user_id = get_current_user_id();
    $conversation_id = $request->get_param('conversation_id');
    
    // 验证权限
    $conversations_table = $wpdb->prefix . 'mytheme_chat_conversations';
    $conversation = $wpdb->get_row($wpdb->prepare(
        "SELECT * FROM $conversations_table WHERE conversation_id = %s",
        $conversation_id
    ));
    
    if (!$conversation || ($conversation->customer_id != $user_id && $conversation->agent_id != $user_id)) {
        return new WP_Error('forbidden', '无权访问此会话', array('status' => 403));
    }
    
    // 确定用户角色
    $user_meta = get_user_meta($user_id, 'chat_role', true);
    $user_role = $user_meta ?: 'customer';
    
    // 标记对方发送的消息为已读
    $messages_table = $wpdb->prefix . 'mytheme_chat_messages';
    $wpdb->query($wpdb->prepare(
        "UPDATE $messages_table 
        SET read_status = 1 
        WHERE conversation_id = %s 
        AND receiver_id = %d 
        AND read_status = 0",
        $conversation_id,
        $user_id
    ));
    
    // 重置未读数
    $unread_field = ($user_role === 'customer') ? 'unread_count_customer' : 'unread_count_agent';
    $wpdb->query($wpdb->prepare(
        "UPDATE $conversations_table 
        SET $unread_field = 0 
        WHERE conversation_id = %s",
        $conversation_id
    ));
    
    return rest_ensure_response(array(
        'success' => true,
        'message' => '已标记为已读'
    ));
}

/**
 * 获取未读消息数
 */
function mytheme_chat_get_unread_count($request) {
    global $wpdb;
    $user_id = get_current_user_id();
    
    // 确定用户角色
    $user_meta = get_user_meta($user_id, 'chat_role', true);
    $user_role = $user_meta ?: 'customer';
    
    $conversations_table = $wpdb->prefix . 'mytheme_chat_conversations';
    $unread_field = ($user_role === 'customer') ? 'unread_count_customer' : 'unread_count_agent';
    $user_field = ($user_role === 'customer') ? 'customer_id' : 'agent_id';
    
    $total_unread = $wpdb->get_var($wpdb->prepare(
        "SELECT SUM($unread_field) 
        FROM $conversations_table 
        WHERE $user_field = %d",
        $user_id
    ));
    
    return rest_ensure_response(array(
        'success' => true,
        'data' => array(
            'total_unread' => intval($total_unread)
        )
    ));
}
