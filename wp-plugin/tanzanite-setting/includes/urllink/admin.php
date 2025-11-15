<?php
if (!defined('ABSPATH')) { exit; }

// local safe-escape helpers (defensive: coerce objects before esc_*)
if (!function_exists('urllink_h')) {
  function urllink_h($v) {
    if ($v instanceof WP_Term) { $v = isset($v->name) ? $v->name : ''; }
    elseif ($v instanceof WP_Post) { $v = isset($v->post_title) ? $v->post_title : ''; }
    elseif (is_array($v)) { $v = wp_json_encode($v); }
    elseif (is_object($v) && !method_exists($v, '__toString')) { $v = wp_json_encode($v); }
    return esc_html((string)$v);
  }
}
if (!function_exists('urllink_a')) {
  function urllink_a($v) {
    if ($v instanceof WP_Term) { $v = isset($v->name) ? $v->name : ''; }
    elseif ($v instanceof WP_Post) { $v = isset($v->post_title) ? $v->post_title : ''; }
    elseif (is_array($v)) { $v = wp_json_encode($v); }
    elseif (is_object($v) && !method_exists($v, '__toString')) { $v = wp_json_encode($v); }
    return esc_attr((string)$v);
  }
}

// Admin menu: URLLink 已集成到 Tanzanite Settings，不再单独注册菜单
// 如果需要独立菜单，取消下面的注释
/*
add_action('admin_menu', function(){
    if (class_exists('MyTheme_SEO_Plugin')) { return; }
    add_menu_page(
        'URLLink',
        'URLLink',
        'manage_options',
        'urllink',
        'urllink_admin_page',
        'dashicons-admin-links',
        58
    );
});
*/

function urllink_capture_post_action(&$notice) {
    if (!current_user_can('manage_options')) return;
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') return;
    if (!isset($_POST['urllink_admin_nonce']) || !wp_verify_nonce($_POST['urllink_admin_nonce'], 'urllink_admin_save')) return;

    $post_id = isset($_POST['post_id']) ? intval($_POST['post_id']) : 0;
    $path    = isset($_POST['path']) ? urllink_normalize_path(wp_unslash($_POST['path'])) : '';
    $extra   = isset($_POST['extra']) ? trim((string)wp_unslash($_POST['extra'])) : '';
    if ($post_id > 0) {
        $old = get_post_meta($post_id, '_urllink_path', true);
        update_post_meta($post_id, '_urllink_path', $path);
        update_post_meta($post_id, '_urllink_extra_redirects', $extra);
        if ($old && $old !== $path) {
            $list = (array)get_post_meta($post_id, '_urllink_old_paths', true);
            if (!in_array($old, $list, true)) { $list[] = $old; }
            update_post_meta($post_id, '_urllink_old_paths', $list);
        }
        urllink_update_map_for_post($post_id, $path);
        $notice = __('Saved.', 'urllink');
    }
}

function urllink_render_block($embedded = false) {
    if (!current_user_can('manage_options')) return;

    $notice = '';
    urllink_capture_post_action($notice);

    $s = isset($_GET['s']) && (is_scalar($_GET['s']) || (is_object($_GET['s']) && method_exists($_GET['s'], '__toString')))
        ? sanitize_text_field((string) wp_unslash($_GET['s'])) : '';
    $pt = isset($_GET['post_type']) && (is_scalar($_GET['post_type']) || (is_object($_GET['post_type']) && method_exists($_GET['post_type'], '__toString')))
        ? sanitize_key((string) $_GET['post_type']) : '';
    $pp = isset($_GET['pp']) ? max(10, min(200, intval($_GET['pp']))) : 50;
    $allowed_types = ['post','page','product'];
    if (!in_array($pt, $allowed_types, true)) { $pt = ''; }

    $args = [ 'posts_per_page' => $pp, 'post_status' => 'any' ];
    if ($s) { $args['s'] = $s; }
    if ($pt) { $args['post_type'] = $pt; }
    $q = new WP_Query($args);

    if (!$embedded) echo '<div class="wrap"><h1>URLLink</h1>';
    if ($embedded)  echo '<div class="urllink-embed">';
    if ($notice) echo '<div class="updated"><p>' . esc_html($notice) . '</p></div>';

    // layout: left tree + right list
    echo '<style>
    /* Typography: unify with MyTheme SEO */
    .urllink-embed { font-size:14px; line-height:1.5; color:#111827; }
    .urllink-embed h1, .urllink-embed h2, .urllink-embed h3, .urllink-embed h4 { margin:0 0 8px; font-weight:600; }
    .urllink-embed h2 { font-size:16px; }
    .urllink-embed h3 { font-size:16px; }
    .urllink-embed, .urllink-embed input, .urllink-embed select, .urllink-embed textarea,
    .urllink-embed .button, .urllink-embed .components-button, .urllink-embed table { font-size:14px; }
    .urllink-embed .toolbar, .urllink-embed .filters { display:flex; align-items:center; gap:8px; flex-wrap:wrap; }
    .urllink-embed .toolbar { justify-content: space-between; }
    .urllink-embed .urllink-table-wrap{ overflow-x:auto; }
    .urllink-embed .urllink-table{ table-layout:fixed; width:100%; }
    .urllink-embed .urllink-table th, .urllink-embed .urllink-table td { vertical-align:top; overflow:hidden; text-overflow:ellipsis; }
    .urllink-embed .urllink-table .col-id { width:60px; }
    .urllink-embed .urllink-table .col-title { width:6%; }
    .urllink-embed .urllink-table .col-permalink { width:20%; }
    .urllink-embed .urllink-table .col-custom { width:28%; }
    .urllink-embed .urllink-table .col-extra { width:16%; }
    .urllink-embed .urllink-table .col-action { width:180px; }
    .urllink-embed .truncate { max-width:100%; word-break:break-all; overflow-wrap:anywhere; }
    .urllink-embed .urllink-table thead th { background:#f3f4f6; color:#111827; font-weight:600; font-size:14px; }
    /* main list content area: higher viewport with internal scroll; border on wrapper to visually expand */
    .urllink-embed .urllink-main-table-wrap {
      max-height: 560px; min-height: 380px; overflow-y: auto;
      border: 1px solid #e5e7eb; border-radius:8px; background:#fff;
    }
    /* avoid double borders from WP widefat */
    .urllink-embed .urllink-main-table-wrap > table.widefat { border: none; box-shadow: none; }
    /* unify with mytheme-seo card/buttons */
    .urllink-embed .button,
    .urllink-embed .button-secondary,
    .urllink-embed .button-primary {
      height:30px; border-radius:8px; padding:0 12px; border:1px solid #e5e7eb;
      background:#f8fafc; color:#111827;
    }
    /* keep components-button visually identical to WP button */
    .urllink-embed .components-button { height:30px; border-radius:8px; padding:0 12px; border:1px solid #e5e7eb; background:#f8fafc; color:#111827; box-shadow:none; }
    .urllink-embed .components-button.is-small { height:30px; padding:0 12px; }
    .urllink-embed .button-primary {
      background:#1f2937; border-color:#1f2937; color:#fff;
    }
    .urllink-embed .components-button.is-primary { background:#1f2937; border-color:#1f2937; color:#fff; }
    /* spacing for to-merge button to avoid overlap with Save */
    .urllink-embed .urllink-to-merge { margin-left:5px; cursor:pointer; }
    .urllink-embed .button:hover,
    .urllink-embed .button:focus { background:#ffffff; border-color:#cfe1f9; box-shadow:0 0 0 2px rgba(34,113,177,0.12); }
    .urllink-embed .button-primary:hover,
    .urllink-embed .button-primary:focus { background:#111827; border-color:#111827; box-shadow:none; }
    .urllink-embed .components-button:hover,
    .urllink-embed .components-button:focus { background:#ffffff; border-color:#cfe1f9; box-shadow:0 0 0 2px rgba(34,113,177,0.12); }
    .urllink-embed .components-button.is-primary:hover,
    .urllink-embed .components-button.is-primary:focus { background:#111827; border-color:#111827; box-shadow:none; }
    .urllink-embed .panel,
    .urllink-embed .urllink-table-wrap { border:1px solid #e5e7eb; border-radius:8px; background:#fff; }
    /* unify control sizes to match mytheme-seo */
    .urllink-embed input[type=text],
    .urllink-embed select,
    .urllink-embed textarea { height:30px; border-radius:8px; border:1px solid #e5e7eb; }
    .urllink-embed textarea { height:auto; min-height:30px; }
    .urllink-embed .filters input[type=text],
    .urllink-embed .filters input[type=search] { height:30px; border-radius:8px; border:1px solid #e5e7eb; }
    .urllink-embed .filters input[type=search]{ width:220px; }
    .urllink-embed .filters input[name=post_type]{ width:180px; }
    /* in-flow stack: params, then merge preview, then bottom buttons */
    .urllink-embed .urllink-card-body { --urllink-gap: 10px; }
    .urllink-embed .urllink-params-row { display:flex; gap:12px; align-items:flex-end; flex-wrap:wrap; margin-top: var(--urllink-gap); }
    .urllink-embed .urllink-merge-wrap { border:1px solid #e5e7eb; border-radius:8px; background:#fff; margin-top: var(--urllink-gap); }
    .urllink-embed .urllink-actions-bottom { display:flex; justify-content:center; align-items:center; gap:8px; margin-top: var(--urllink-gap); flex-wrap:wrap; }
    .urllink-embed .urllink-merge-title { font-weight:600; padding:8px 10px; border-bottom:1px solid #e5e7eb; }
    .urllink-embed .urllink-merge-table .merge-line td { padding:8px 10px; }
    .urllink-embed .urllink-merge-table .merge-actions td { padding:8px 10px; text-align:center; }
    .urllink-embed .urllink-merge-table .url-line { display:block; width:100%; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
    </style>';
    echo '<div style="display:flex; gap:18px; align-items:flex-start;">';

    // Left: directory tree
    echo '<div class="mytheme-seo-card" style="flex:0 0 500px;"><div class="components-card__body" style="padding:8px; min-height:600px;">';
    echo '<h3 style="margin-top:0;">URL 目录树</h3>';
    echo '<p style="margin:6px 0 8px;">可在此快速新建目录（即时保存到 WordPress），或前往“文章 → URL 目录”进行完整管理。</p>';

    // Handle create new directory
    if (isset($_POST['urllink_add_dir']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        $dir_name = isset($_POST['dir_name']) ? sanitize_text_field((string) wp_unslash($_POST['dir_name'])) : '';
        $dir_slug = isset($_POST['dir_slug']) ? sanitize_title_with_dashes((string) wp_unslash($_POST['dir_slug'])) : '';
        $dir_parent = isset($_POST['dir_parent']) ? max(0, intval($_POST['dir_parent'])) : 0;
        if ($dir_name !== '' && $dir_slug !== '') {
            $ins = wp_insert_term($dir_name, 'urllink_dir', [ 'parent' => $dir_parent ]);
            if (!is_wp_error($ins) && !empty($ins['term_id'])) {
                update_term_meta((int)$ins['term_id'], 'path_slug', $dir_slug);
                echo '<div class="notice notice-success"><p>已新建目录：' . esc_html($dir_name) . '（' . esc_html($dir_slug) . '）</p></div>';
            } else {
                $msg = is_wp_error($ins) ? $ins->get_error_message() : 'unknown error';
                echo '<div class="notice notice-error"><p>新建失败：' . esc_html($msg) . '</p></div>';
            }
        } else {
            echo '<div class="notice notice-warning"><p>请填写名称与路径别名（path_slug）。</p></div>';
        }
    }

    // Handle attach URL to directory (map post under the directory slug)
    if (isset($_POST['urllink_dir_attach']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        $term_id = isset($_POST['term_id']) ? intval($_POST['term_id']) : 0;
        $raw_url = isset($_POST['attach_url']) ? (string) wp_unslash($_POST['attach_url']) : '';
        $want_sync = !empty($_POST['attach_and_sync']);
        if ($term_id > 0 && $raw_url !== '') {
            $dir_slug = (string) get_term_meta($term_id, 'path_slug', true);
            $dir_slug = trim($dir_slug, '/');
            if ($dir_slug === '') {
                echo '<div class="notice notice-warning"><p>该目录尚未设置 path_slug，无法附加。请先为目录设置 path_slug。</p></div>';
            } else {
                // Resolve URL to post ID
                $post_id = url_to_postid($raw_url);
                if ($post_id <= 0) {
                    // try parse path and match by last segment as post_name
                    $path = $raw_url;
                    $parts = wp_parse_url($raw_url);
                    if (is_array($parts) && !empty($parts['path'])) { $path = $parts['path']; }
                    $path = trim((string)$path, '/');
                    $segments = $path !== '' ? explode('/', $path) : [];
                    $last = is_array($segments) && !empty($segments) ? sanitize_title(end($segments)) : '';
                    if ($last !== '') {
                        // search across common types
                        $candidate = get_page_by_path($last, OBJECT, ['post','page','product']);
                        if ($candidate && isset($candidate->ID)) { $post_id = intval($candidate->ID); }
                        if ($post_id <= 0) {
                            $q = get_posts([
                                'name' => $last,
                                'post_type' => ['post','page','product'],
                                'post_status' => 'any',
                                'numberposts' => 1,
                                'fields' => 'ids'
                            ]);
                            if (is_array($q) && !empty($q)) { $post_id = intval($q[0]); }
                        }
                    }
                }

                if ($post_id > 0) {
                    $post_name = (string) get_post_field('post_name', $post_id);
                    $new_path  = trim($dir_slug . '/' . ltrim($post_name, '/'), '/');
                    $old = get_post_meta($post_id, '_urllink_path', true);
                    update_post_meta($post_id, '_urllink_path', $new_path);
                    if ($old && $old !== $new_path) {
                        $list = (array)get_post_meta($post_id, '_urllink_old_paths', true);
                        if (!in_array($old, $list, true)) { $list[] = $old; }
                        update_post_meta($post_id, '_urllink_old_paths', $list);
                    }
                    if (function_exists('urllink_update_map_for_post')) { urllink_update_map_for_post($post_id, $new_path); }

                    if ($want_sync) {
                        do_action('mytheme_seo_request_sitemap_rebuild', [ 'source' => 'urllink_attach' ]);
                        do_action('mytheme_seo_refresh_after_urllink');
                    }
                    $final_url = get_permalink($post_id);
                    echo '<div class="notice notice-success"><p>已附加到目录：<code>' . esc_html($dir_slug) . '</code> / <code>' . esc_html($post_name) . '</code>。当前应用网址：<a href="' . esc_url($final_url) . '" target="_blank">' . esc_html($final_url) . '</a></p></div>';
                } else {
                    echo '<div class="notice notice-error"><p>无法从输入解析到文章，请检查网址是否正确，或确保对应内容存在。</p></div>';
                }
            }
        } else {
            echo '<div class="notice notice-warning"><p>请填写有效的网址。</p></div>';
        }
    }

    // Quick create form
    $all_for_parent = get_terms([ 'taxonomy' => 'urllink_dir', 'hide_empty' => false ]);
    echo '<form method="post" style="margin:8px 0 12px; display:grid; grid-template-columns: 1fr; gap:6px;">';
    wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
    echo '<input type="text" name="dir_name" placeholder="目录名称（必填）" />';
    echo '<input type="text" name="dir_slug" placeholder="路径别名 path_slug（必填，如 products/fr）" />';
    echo '<select name="dir_parent">';
    echo '<option value="0">置于顶层</option>';
    if (!is_wp_error($all_for_parent) && is_array($all_for_parent)) {
        foreach ($all_for_parent as $pt) {
            if (!($pt instanceof WP_Term)) { continue; }
            $name = isset($pt->name) ? (string)$pt->name : '';
            echo '<option value="' . intval($pt->term_id) . '">' . esc_html($name) . '</option>';
        }
    }
    echo '</select>';
    echo '<button class="button">新建目录</button>';
    echo '<input type="hidden" name="urllink_add_dir" value="1" />';
    echo '</form>';
    // Handle rename/delete dir
    if (isset($_POST['urllink_dir_action']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        $action = sanitize_key((string) $_POST['urllink_dir_action']);
        $term_id = isset($_POST['term_id']) ? intval($_POST['term_id']) : 0;
        if ($term_id > 0 && $action === 'rename') {
            $new_name = isset($_POST['new_name']) ? sanitize_text_field((string) wp_unslash($_POST['new_name'])) : '';
            $new_slug = isset($_POST['new_slug']) ? sanitize_title_with_dashes((string) wp_unslash($_POST['new_slug'])) : '';
            if ($new_name !== '' && $new_slug !== '') {
                $res = wp_update_term($term_id, 'urllink_dir', [ 'name' => $new_name ]);
                if (!is_wp_error($res)) {
                    update_term_meta($term_id, 'path_slug', $new_slug);
                    echo '<div class="notice notice-success"><p>目录已重命名：' . esc_html($new_name) . '（' . esc_html($new_slug) . '）</p></div>';
                } else {
                    echo '<div class="notice notice-error"><p>重命名失败：' . esc_html($res->get_error_message()) . '</p></div>';
                }
            } else {
                echo '<div class="notice notice-warning"><p>请填写名称与路径别名。</p></div>';
            }
        } elseif ($term_id > 0 && $action === 'delete') {
            // 安全起见：有子项则不允许删除
            $kids = get_terms([ 'taxonomy' => 'urllink_dir', 'hide_empty' => false, 'parent' => $term_id, 'fields' => 'ids' ]);
            if (!empty($kids)) {
                echo '<div class="notice notice-warning"><p>请先删除或移动其子目录后再删除该目录。</p></div>';
            } else {
                $res = wp_delete_term($term_id, 'urllink_dir');
                if (!is_wp_error($res)) {
                    echo '<div class="notice notice-success"><p>目录已删除。</p></div>';
                } else {
                    echo '<div class="notice notice-error"><p>删除失败：' . esc_html($res->get_error_message()) . '</p></div>';
                }
            }
        }
    }

    $terms = get_terms([ 'taxonomy' => 'urllink_dir', 'hide_empty' => false, 'parent' => 0 ]);
    $render_branch = function($parent) use (&$render_branch) {
        if (!($parent instanceof WP_Term)) { return; }
        $children = get_terms([ 'taxonomy' => 'urllink_dir', 'hide_empty' => false, 'parent' => intval($parent->term_id) ]);
        $slug = (string) get_term_meta(intval($parent->term_id), 'path_slug', true);
        $pname = isset($parent->name) ? (string)$parent->name : '';
        echo '<li>' . esc_html($pname) . ' <code>' . esc_html($slug) . '</code> ';
        // Inline actions: rename / delete
        echo '<details style="display:inline-block; margin-left:6px;"><summary style="cursor:pointer;">操作</summary>';
        echo '<div style="padding:6px 0 0 0;">';
        // rename form
        echo '<form method="post" style="display:grid; grid-template-columns: 1fr; gap:6px; margin:6px 0;">';
        wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
        echo '<input type="hidden" name="term_id" value="' . intval($parent->term_id) . '" />';
        echo '<input type="hidden" name="urllink_dir_action" value="rename" />';
        echo '<input type="text" name="new_name" placeholder="新名称" value="' . esc_attr($parent->name) . '" />';
        echo '<input type="text" name="new_slug" placeholder="新 path_slug" value="' . esc_attr((string)$slug) . '" />';
        echo '<button class="button">保存</button>';
        echo '</form>';
        // delete form
        echo '<form method="post" onsubmit="return confirm(\'确认删除该目录？若有子目录将阻止删除。\');" style="margin:0;">';
        wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
        echo '<input type="hidden" name="term_id" value="' . intval($parent->term_id) . '" />';
        echo '<input type="hidden" name="urllink_dir_action" value="delete" />';
        echo '<button class="button button-link-delete" style="color:#d63638;">删除目录</button>';
        echo '</form>';
        echo '</div></details> ';
        echo '<a href="' . esc_url(get_edit_term_link($parent->term_id, 'urllink_dir')) . '" target="_blank">在标签页中编辑</a>';
        // Inline attach URL form
        echo '<form method="post" style="margin:6px 0; display:grid; grid-template-columns: 1fr; gap:6px;">';
        wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
        echo '<input type="hidden" name="term_id" value="' . intval($parent->term_id) . '" />';
        echo '<label style="font-size:12px; color:#6b7280;">将页面附加到该目录（粘贴网址）：</label>';
        echo '<div style="display:flex; gap:8px; align-items:center;">';
        echo '<input type="url" name="attach_url" placeholder="https://example.com/path/to/page" style="flex:1; min-width:0;" />';
        echo '<button class="button button-primary" name="urllink_dir_attach" value="1" onclick="this.form.attach_and_sync.value=1;">确定并同步</button>';
        echo '<input type="hidden" name="attach_and_sync" value="0" />';
        echo '</div>';
        echo '</form>';
        if (!is_wp_error($children) && !empty($children) && is_array($children)) {
            echo '<ul style="margin:6px 0 6px 16px;">';
            foreach ($children as $c) { if ($c instanceof WP_Term) { $render_branch($c); } }
            echo '</ul>';
        }
        echo '</li>';
    };
    if (!is_wp_error($terms) && !empty($terms) && is_array($terms)) {
        echo '<ul style="margin:0;">';
        foreach ($terms as $t) { if ($t instanceof WP_Term) { $render_branch($t); } }
        echo '</ul>';
    } else {
        echo '<p>暂无目录。前往 <a href="' . esc_url(admin_url('edit-tags.php?taxonomy=urllink_dir')) . '" target="_blank">管理目录</a></p>';
    }
    echo '</div></div>';

    // Right: actions + table + bulk actions (guarded)
    echo '<div style="flex:1 1 auto;">';
    // (Removed external toolbar; actions will appear inside the card toolbar below)

    // Card wrapper for table and controls
    echo '<div class="mytheme-seo-card"><div class="components-card__body urllink-card-body">';
    echo '<div class="toolbar" style="display:none;">';
    // left group: site-level actions moved to footer
    echo '<div style="display:flex; gap:8px; align-items:center;"></div>';
    // right group: bulk actions (moved to footer); keep container for right-side filters
    echo '<div style="display:flex; gap:8px; align-items:center;">';
        // dir selector
        $all_terms_toolbar = get_terms([ 'taxonomy' => 'urllink_dir', 'hide_empty' => false ]);
        $dirSelect = '<select name="dir_id" form="urllink-bulk-form" class="components-select-control__input" style="min-width:180px;">';
        $dirSelect .= '<option value="0">— 选择目录 —</option>';
        if (!is_wp_error($all_terms_toolbar) && is_array($all_terms_toolbar)) {
            foreach ($all_terms_toolbar as $t) {
                if (!($t instanceof WP_Term)) { continue; }
                $slug = (string) get_term_meta(intval($t->term_id), 'path_slug', true);
                $tname = isset($t->name) ? (string)$t->name : '';
                $label = (string)$tname . ' (' . (string)$slug . ')';
                $dirSelect .= '<option value="' . intval($t->term_id) . '">' . urllink_h($label) . '</option>';
            }
        }
        $dirSelect .= '</select>';
        echo '<label style="display:flex; gap:6px; align-items:center;">应用到目录： ' . $dirSelect . '</label>';
        echo '<span>策略：</span>';
        echo '<label><input type="radio" name="strategy" value="replace" checked form="urllink-bulk-form" /> replace</label>';
        echo '<label><input type="radio" name="strategy" value="prefix" form="urllink-bulk-form" /> prefix</label>';
        echo '<input type="text" name="old_prefix" form="urllink-bulk-form" placeholder="旧前缀（可选）" style="width:150px;" />';
        echo '<label style="display:flex; gap:6px; align-items:center;"><input type="checkbox" name="dry_run" value="1" form="urllink-bulk-form" checked /> 干跑预览</label>';
        echo '<button class="components-button is-primary" name="urllink_bulk_apply" value="1" form="urllink-bulk-form">执行</button>';
    
    echo '</div>'; // end right group
    echo '</div>';

    if (isset($_POST['urllink_rebuild_maps'])) { URLLink_Plugin::rebuild_maps(); echo '<div class="updated"><p>映射已重建。</p></div>'; }
    if (isset($_POST['urllink_save_and_sync']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        URLLink_Plugin::rebuild_maps();
        // 通知 MyTheme SEO 执行 sitemap 刷新/提交，由其内部实现具体逻辑
        do_action('mytheme_seo_request_sitemap_rebuild', [ 'source' => 'urllink' ]);
        do_action('mytheme_seo_refresh_after_urllink');
        echo '<div class="updated"><p>已保存并请求 MyTheme SEO 刷新站点地图。</p></div>';
    }

    // Handle new bulk action buttons (based on selected ids[])
    if (isset($_POST['urllink_bulk_refresh'])) {
        echo '<div class="updated"><p>' . urllink_h(__('已刷新列表。', 'urllink')) . '</p></div>';
    }
    if (isset($_POST['urllink_bulk_duplicate']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        $ids = isset($_POST['ids']) && is_array($_POST['ids']) ? array_map('intval', $_POST['ids']) : [];
        $done = 0; $errors = 0;
        foreach ($ids as $pid) {
            $p = get_post($pid);
            if (!$p) { $errors++; continue; }
            $base_slug = sanitize_title($p->post_name . '-copy');
            if (function_exists('wp_unique_post_slug')) {
                $base_slug = wp_unique_post_slug($base_slug, 0, 'draft', $p->post_type, $p->post_parent);
            }
            $new_post = [
                'post_type' => $p->post_type,
                'post_status' => 'draft',
                'post_title' => $p->post_title . ' (Copy)',
                'post_name'  => $base_slug,
                'post_content' => $p->post_content,
                'post_excerpt' => $p->post_excerpt,
            ];
            $new_id = wp_insert_post($new_post, true);
            if (!is_wp_error($new_id) && $new_id) {
                // copy all metas
                $all_meta = get_post_meta($pid);
                if (is_array($all_meta)) {
                    $skip_keys = ['_edit_lock','_edit_last','_wp_old_slug'];
                    foreach ($all_meta as $k => $vals) {
                        if (in_array($k, $skip_keys, true)) continue;
                        foreach ((array)$vals as $v) { add_post_meta($new_id, $k, maybe_unserialize($v)); }
                    }
                }
                // ensure mark cloned & update maps if cloned has path
                update_post_meta($new_id, '_urllink_cloned', 1);
                update_post_meta($new_id, '_urllink_cloned_from', intval($pid));
                $path = get_post_meta($new_id, '_urllink_path', true);
                if ($path !== '' && $path !== null) {
                    $base = rtrim((string)$path, '/');
                    $candidate = $base . '-copy';
                    $map = get_option('urllink_path_map', []);
                    if (!is_array($map)) { $map = []; }
                    // ensure uniqueness
                    $i = 2;
                    $norm = function_exists('urllink_normalize_path') ? urllink_normalize_path($candidate) : trim($candidate, '/');
                    while (isset($map[$norm])) {
                        $candidate = $base . '-copy-' . $i;
                        $norm = function_exists('urllink_normalize_path') ? urllink_normalize_path($candidate) : trim($candidate, '/');
                        $i++;
                    }
                    update_post_meta($new_id, '_urllink_path', $candidate);
                    urllink_update_map_for_post($new_id, $candidate);
                }
                // copy taxonomies (terms)
                $taxes = get_object_taxonomies($p->post_type);
                if (is_array($taxes)) {
                    foreach ($taxes as $tax) {
                        $terms = wp_get_object_terms($pid, $tax, ['fields' => 'ids']);
                        if (!is_wp_error($terms) && !empty($terms)) { wp_set_object_terms($new_id, $terms, $tax, false); }
                    }
                }
                // copy featured image
                $thumb_id = get_post_thumbnail_id($pid);
                if ($thumb_id) { set_post_thumbnail($new_id, $thumb_id); }
                $done++;
            } else { $errors++; }
        }
        echo '<div class="updated"><p>' . urllink_h(sprintf(__('已复制 %d 条，失败 %d 条。', 'urllink'), $done, $errors)) . '</p></div>';
    }
    if (isset($_POST['urllink_bulk_delete']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        $ids = isset($_POST['ids']) && is_array($_POST['ids']) ? array_map('intval', $_POST['ids']) : [];
        $done = 0; $errors = 0;
        foreach ($ids as $pid) {
            $res = wp_trash_post($pid);
            if ($res !== false) { $done++; } else { $errors++; }
        }
        echo '<div class="updated"><p>' . urllink_h(sprintf(__('已移至回收站 %d 条，失败 %d 条。', 'urllink'), $done, $errors)) . '</p></div>';
    }
    if (isset($_POST['urllink_bulk_delete_force']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        $ids = isset($_POST['ids']) && is_array($_POST['ids']) ? array_map('intval', $_POST['ids']) : [];
        $done = 0; $errors = 0;
        foreach ($ids as $pid) {
            $res = wp_delete_post($pid, true);
            if ($res !== false) { $done++; } else { $errors++; }
        }
        echo '<div class="updated"><p>' . urllink_h(sprintf(__('已彻底删除 %d 条，失败 %d 条。', 'urllink'), $done, $errors)) . '</p></div>';
    }
    if (isset($_POST['urllink_bulk_syncwp']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        URLLink_Plugin::rebuild_maps();
        do_action('mytheme_seo_request_sitemap_rebuild', [ 'source' => 'urllink-bulk' ]);
        do_action('mytheme_seo_refresh_after_urllink');
        echo '<div class="updated"><p>' . urllink_h(__('已请求同步到 WordPress（站点地图刷新）', 'urllink')) . '</p></div>';
    }

    if ($q->have_posts()) {
        echo '<form id="urllink-bulk-form" method="post" style="margin:0;">';
        wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
        echo '<div class="urllink-table-wrap urllink-main-table-wrap">';
        echo '<table class="widefat fixed striped urllink-table"><thead><tr>';
        echo '<th style="width:28px"><input type="checkbox" id="urllink-select-all" /></th>';
        echo '<th class="col-id">ID</th><th class="col-title">Title</th><th class="col-permalink">Current Permalink</th><th class="col-custom">Custom Path</th><th class="col-extra">Extra Redirects</th><th class="col-action">Action</th>';
        echo '</tr></thead><tbody>';
        while ($q->have_posts()) { $q->the_post();
            $pid = get_the_ID();
            $custom = get_post_meta($pid, '_urllink_path', true);
            $extra  = get_post_meta($pid, '_urllink_extra_redirects', true);
            echo '<tr>';
            echo '<td><input type="checkbox" name="ids[]" value="' . intval($pid) . '" /></td>';
            echo '<td>' . intval($pid) . '</td>';
            $title = get_the_title();
            $from_id = get_post_meta($pid, '_urllink_cloned_from', true);
            $badge = get_post_meta($pid, '_urllink_cloned', true)
                ? ' <span title="cloned from #' . intval($from_id) . '" style="display:inline-block; padding:2px 6px; border-radius:6px; background:#fde68a; color:#92400e; font-size:12px;">CLONED #' . intval($from_id) . '</span>'
                : '';
            echo '<td class="truncate">' . urllink_h($title) . $badge . '</td>';
            $pl = (string) get_permalink($pid);
            echo '<td class="truncate"><a href="' . esc_url($pl) . '" title="' . esc_attr($pl) . '" target="_blank">' . esc_html($pl) . '</a></td>';
            echo '<td class="truncate">';
            $form_id = 'urllink-row-' . intval($pid);
            echo '<input type="text" name="path" form="' . urllink_a($form_id) . '" style="width:100%" placeholder="e.g. products/{sku}-{postname}" value="' . urllink_a((string)$custom) . '" />';
            echo '</td><td class="truncate">';
            echo '<textarea name="extra" form="' . urllink_a($form_id) . '" rows="2" style="width:100%" placeholder="one per line">' . esc_textarea((string)$extra) . '</textarea>';
            echo '</td><td>';
            echo '<form id="' . urllink_a($form_id) . '" method="post" style="margin:0; display:flex; flex-wrap:nowrap; gap:6px; align-items:center;">';
            wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
            echo '<input type="hidden" name="post_id" value="' . intval($pid) . '" />';
            submit_button(__('Save'), 'primary', '', false, ['class' => 'components-button is-primary is-small']);
            // to merge button
            $title_safe = esc_attr(get_the_title($pid));
            $plink_safe = esc_url(get_permalink($pid));
            echo '<button type="button" class="components-button is-primary is-small urllink-to-merge" data-merge-id="' . intval($pid) . '" data-merge-title="' . $title_safe . '" data-merge-pl="' . $plink_safe . '">to merge</button>';
            echo '</form>';
            echo '</td>';
            echo '</tr>';
        }
        echo '</tbody></table>';
        echo '</div>';
        // Bind select-all without complex inline JS to avoid layout glitches
        echo '<script>(function(){\n  var selAll=document.getElementById("urllink-select-all");\n  if(selAll){ selAll.addEventListener("change",function(){\n    var boxes=document.querySelectorAll("#urllink-bulk-form input[name=\\"ids[]\\"]");\n    boxes.forEach(function(cb){ cb.checked=selAll.checked; });\n  });}\n})();</script>';
        echo '<div style="display:none">';
        // dir selector
        $all_terms = get_terms([ 'taxonomy' => 'urllink_dir', 'hide_empty' => false ]);
        echo '<div><label>应用到目录：</label> <select name="dir_id" form="urllink-bulk-form">';
        echo '<option value="0">— 选择目录 —</option>';
        if (!is_wp_error($all_terms) && is_array($all_terms)) {
            foreach ($all_terms as $t) {
                if (!($t instanceof WP_Term)) { continue; }
                $slug = (string) get_term_meta(intval($t->term_id), 'path_slug', true);
                $tname = isset($t->name) ? (string)$t->name : '';
                $label = (string)$tname . ' (' . (string)$slug . ')';
                echo '<option value="' . intval($t->term_id) . '">' . urllink_h($label) . '</option>';
            }
        }
        echo '</select></div>';
        echo '<div><label>策略：</label> ';
        echo '<label><input type="radio" name="strategy" value="replace" checked form="urllink-bulk-form" /> replace</label> ';
        echo '<label><input type="radio" name="strategy" value="prefix" form="urllink-bulk-form" /> prefix</label>';
        echo '</div>';
        echo '<div><label>旧前缀（可选）：</label> <input type="text" name="old_prefix" placeholder="如 catalog" form="urllink-bulk-form" /></div>';
        echo '<div><label><input type="checkbox" name="dry_run" value="1" form="urllink-bulk-form" checked /> 干跑预览</label></div>';
        echo '<div><button class="components-button is-primary" name="urllink_bulk_apply" value="1" form="urllink-bulk-form">执行</button></div>';
        echo '</div>';
        echo '</form>';

        // In-flow params row (10px below table)
        echo '<div class="urllink-params-row">';
        $all_terms_under = get_terms([ 'taxonomy' => 'urllink_dir', 'hide_empty' => false ]);
        echo '<div><label>应用到目录：</label> <select name="dir_id" form="urllink-bulk-form">';
        echo '<option value="0">— 选择目录 —</option>';
        if (!is_wp_error($all_terms_under) && is_array($all_terms_under)) {
            foreach ($all_terms_under as $t) {
                if (!($t instanceof WP_Term)) { continue; }
                $slug = (string) get_term_meta(intval($t->term_id), 'path_slug', true);
                $tname = isset($t->name) ? (string)$t->name : '';
                $label = (string)$tname . ' (' . (string)$slug . ')';
                echo '<option value="' . intval($t->term_id) . '">' . urllink_h($label) . '</option>';
            }
        }
        echo '</select></div>';
        echo '<div><label>策略：</label> ';
        echo '<label><input type="radio" name="strategy" value="replace" checked form="urllink-bulk-form" /> replace</label> ';
        echo '<label><input type="radio" name="strategy" value="prefix" form="urllink-bulk-form" /> prefix</label>';
        echo '</div>';
        echo '<div><label>旧前缀（可选）：</label> <input type="text" name="old_prefix" placeholder="如 catalog" form="urllink-bulk-form" /></div>';
        echo '<div><label><input type="checkbox" name="dry_run" value="1" form="urllink-bulk-form" checked /> 干跑预览</label></div>';
        echo '<div><button class="components-button is-primary" name="urllink_bulk_apply" value="1" form="urllink-bulk-form">执行</button></div>';
        echo '</div>';

        // In-flow merge preview (10px below params)
        echo '<div class="urllink-merge-wrap">';
        echo '<div class="urllink-merge-title">合并预览</div>';
        echo '<div class="urllink-table-wrap">';
        echo '<table class="widefat fixed striped urllink-table urllink-merge-table">';
        echo '<thead><tr>';
        echo '<th style="width:36px"></th>';
        echo '<th class="col-id">ID</th><th class="col-title">Title</th><th class="col-permalink">Current Permalink</th><th class="col-custom">Custom Path</th><th class="col-extra">Extra Redirects</th><th class="col-action">Action</th>';
        echo '</tr></thead><tbody>';
        echo '<tr class="merge-line" id="urllink-merge-row1">';
        echo '<td><input type="checkbox" id="urllink-merge1-check" class="urllink-merge-checkbox" data-slot="0" disabled /></td>';
        echo '<td id="urllink-merge1-id">—</td>';
        echo '<td class="truncate" id="urllink-merge1-title">—</td>';
        echo '<td class="truncate" id="urllink-merge1-link">—</td>';
        echo '<td class="truncate" id="urllink-merge1-custom"><input type="text" id="urllink-merge1-custom-input" style="width:100%" disabled placeholder="e.g. products/{sku}-{postname}" value="" /></td>';
        echo '<td class="truncate" id="urllink-merge1-extra"><textarea id="urllink-merge1-extra-textarea" rows="2" style="width:100%" disabled placeholder="one per line"></textarea></td>';
        echo '<td>—</td>';
        echo '</tr>';
        echo '<tr class="merge-line" id="urllink-merge-row2">';
        echo '<td><input type="checkbox" id="urllink-merge2-check" class="urllink-merge-checkbox" data-slot="1" disabled /></td>';
        echo '<td id="urllink-merge2-id">—</td>';
        echo '<td class="truncate" id="urllink-merge2-title">—</td>';
        echo '<td class="truncate" id="urllink-merge2-link">—</td>';
        echo '<td class="truncate" id="urllink-merge2-custom"><input type="text" id="urllink-merge2-custom-input" style="width:100%" disabled placeholder="e.g. products/{sku}-{postname}" value="" /></td>';
        echo '<td class="truncate" id="urllink-merge2-extra"><textarea id="urllink-merge2-extra-textarea" rows="2" style="width:100%" disabled placeholder="one per line"></textarea></td>';
        echo '<td>—</td>';
        echo '</tr>';
        echo '<tr class="merge-actions"><td colspan="7">';
        echo '<button class="components-button is-primary" id="urllink-merge-confirm" disabled>确定合并</button> ';
        echo '<button class="components-button" id="urllink-merge-clear">删除</button>';
        echo '</td></tr>';
        echo '</tbody></table>';
        echo '</div>';
        echo '</div>';

        // In-flow bottom actions (10px below merge preview)
        echo '<div class="urllink-actions-bottom">';
        echo '<form method="post" style="margin:0;">';
        wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
        echo '<input type="hidden" name="post_id" value="0" />';
        echo '<button class="components-button" name="urllink_rebuild_maps" value="1">' . esc_html__('刷新映射（重建 path_map/post_map）', 'urllink') . '</button>';
        echo '</form>';
        echo '<form method="post" style="margin:0;">';
        wp_nonce_field('urllink_admin_save', 'urllink_admin_nonce');
        echo '<input type="hidden" name="post_id" value="0" />';
        echo '<button class="components-button is-primary" name="urllink_save_and_sync" value="1">' . esc_html__('保存并同步到 MyTheme SEO（触发站点地图刷新）', 'urllink') . '</button>';
        echo '</form>';
        echo '<button class="components-button" form="urllink-bulk-form" name="urllink_bulk_refresh" value="1">' . urllink_h(__('刷新', 'urllink')) . '</button>';
        echo '<button class="components-button" form="urllink-bulk-form" name="urllink_bulk_duplicate" value="1">' . urllink_h(__('复制页面', 'urllink')) . '</button>';
        echo '<button class="components-button" form="urllink-bulk-form" name="urllink_bulk_delete" value="1" onclick="return confirm(\'确认将所选页面移至回收站？\');">' . urllink_h(__('删除页面', 'urllink')) . '</button>';
        echo '<button class="components-button" form="urllink-bulk-form" name="urllink_bulk_delete_force" value="1" onclick="return confirm(\'确认彻底删除所选页面？该操作不可恢复！\');">' . urllink_h(__('彻底删除', 'urllink')) . '</button>';
        echo '<button class="components-button is-primary" form="urllink-bulk-form" name="urllink_bulk_syncwp" value="1">' . urllink_h(__('同步到 WordPress', 'urllink')) . '</button>';
        echo '</div>';

        wp_reset_postdata();

    }

    // Close card wrapper
    echo '</div></div>';

    echo '<div class="mytheme-seo-card" style="margin-top:16px;"><div class="components-card__body">';
    echo '<h2 style="margin:0 0 8px;">占位符说明</h2>';
    echo '<ul style="list-style:disc; padding-left:18px; margin:0;">';
    echo '<li>{postname}/{slug}：文章别名</li>';
    echo '<li>{sku}：WooCommerce SKU（无则读取 _sku meta）</li>';
    echo '<li>{field:meta_key}：从 ACF 或原生 post_meta 读取</li>';
    echo '</ul>';
    echo '</div></div>';

    // Handle bulk apply/preview
    if (isset($_POST['urllink_bulk_apply']) && check_admin_referer('urllink_admin_save', 'urllink_admin_nonce')) {
        $ids = isset($_POST['ids']) && is_array($_POST['ids']) ? array_map('intval', $_POST['ids']) : [];
        $dir_id = isset($_POST['dir_id']) ? intval($_POST['dir_id']) : 0;
        $strategy = isset($_POST['strategy']) && $_POST['strategy'] === 'prefix' ? 'prefix' : 'replace';
        $old_prefix = isset($_POST['old_prefix']) ? urllink_normalize_path(wp_unslash($_POST['old_prefix'])) : '';
        $dry = !empty($_POST['dry_run']);
        $dir_slug = $dir_id > 0 ? (string) get_term_meta($dir_id, 'path_slug', true) : '';
        $dir_slug = trim($dir_slug, '/');
        if (!empty($ids) && $dir_slug !== '') {
            list($path_map) = urllink_get_maps();
            echo '<div class="notice notice-info"><p>批量处理（' . esc_html($strategy) . '，目录前缀=' . esc_html($dir_slug) . '，' . ($dry ? '干跑预览' : '实际执行') . '）</p></div>';
            echo '<table class="widefat fixed striped"><thead><tr><th>ID</th><th>Title</th><th>Old</th><th>New</th><th>冲突</th><th>结果</th></tr></thead><tbody>';
            foreach ($ids as $pid) {
                $title = get_the_title($pid);
                $old_path = (string) get_post_meta($pid, '_urllink_path', true);
                $old_path = $old_path !== '' ? urllink_normalize_path($old_path) : sanitize_title(get_post_field('post_name', $pid));
                $new_path = $old_path;
                if ($strategy === 'prefix') {
                    $new_path = trim($dir_slug . '/' . ltrim($old_path, '/'), '/');
                } else {
                    // replace
                    $segments = preg_split('#/#', trim($old_path, '/'));
                    if (!is_array($segments)) { $segments = [$old_path]; }
                    if ($old_prefix !== '') {
                        $op = trim($old_prefix, '/');
                        if (stripos($old_path, $op . '/') === 0) {
                            $rest = substr($old_path, strlen($op) + 1);
                            $new_path = trim($dir_slug . '/' . ltrim($rest, '/'), '/');
                        } else {
                            // 若不匹配旧前缀，则仍然用替换第一层级
                            array_shift($segments);
                            $rest = implode('/', $segments);
                            $new_path = trim($dir_slug . '/' . ltrim($rest, '/'), '/');
                        }
                    } else {
                        array_shift($segments);
                        $rest = implode('/', $segments);
                        $new_path = trim($dir_slug . '/' . ltrim($rest, '/'), '/');
                    }
                }
                $conflict = isset($path_map[$new_path]) && intval($path_map[$new_path]) !== intval($pid);
                $result = '—';
                if (!$dry && !$conflict) {
                    $old = get_post_meta($pid, '_urllink_path', true);
                    update_post_meta($pid, '_urllink_path', $new_path);
                    if ($old && $old !== $new_path) {
                        $list = (array)get_post_meta($pid, '_urllink_old_paths', true);
                        if (!in_array($old, $list, true)) { $list[] = $old; }
                        update_post_meta($pid, '_urllink_old_paths', $list);
                    }
                    urllink_update_map_for_post($pid, $new_path);
                    $result = 'OK';
                } elseif ($dry) {
                    $result = '预览';
                } else {
                    $result = '冲突未执行';
                }
                echo '<tr><td>' . intval($pid) . '</td><td>' . esc_html($title) . '</td><td><code>' . esc_html($old_path) . '</code></td><td><code>' . esc_html($new_path) . '</code></td><td>' . ($conflict ? '<span style="color:#d63638">是</span>' : '否') . '</td><td>' . esc_html($result) . '</td></tr>';
            }
            echo '</tbody></table>';
        } else {
            echo '<div class="notice notice-warning"><p>请选择条目与目录，并确保目录具有 path_slug。</p></div>';
        }
    }

    echo '</div>'; // right panel
    echo '</div>'; // layout wrap
    echo '</div>';
    // simple client-side merge preview fill logic (echo NOWDOC to avoid PHP close/open)
    echo <<<'JS'
<script>(function(){
  var sel=[];
  function applyToRow(idx, item){
    var idEl=document.getElementById("urllink-merge"+idx+"-id");
    var tEl=document.getElementById("urllink-merge"+idx+"-title");
    var lEl=document.getElementById("urllink-merge"+idx+"-link");
    var cInput=document.getElementById("urllink-merge"+idx+"-custom-input");
    var eTa=document.getElementById("urllink-merge"+idx+"-extra-textarea");
    var chk=document.getElementById("urllink-merge"+idx+"-check");
    if(!idEl||!tEl||!lEl) return;
    if(item){
      idEl.textContent = item.id;
      tEl.textContent = item.title;
      while(lEl.firstChild){ lEl.removeChild(lEl.firstChild); }
      var a=document.createElement("a");
      a.setAttribute("href", item.pl||"");
      a.setAttribute("target","_blank");
      a.appendChild(document.createTextNode(item.pl||""));
      lEl.appendChild(a);
      if(cInput){ cInput.value = item.path || ""; }
      if(eTa){ eTa.value = item.extra || ""; }
      if(chk){ chk.disabled=false; chk.checked=false; }
    }else{
      idEl.textContent = "—"; tEl.textContent="—"; lEl.textContent="—";
      if(cInput){ cInput.value = ""; }
      if(eTa){ eTa.value = ""; }
      if(chk){ chk.disabled=true; chk.checked=false; }
    }
  }

  function render(){
    applyToRow(1, sel[0]);
    applyToRow(2, sel[1]);
    var ok=document.getElementById("urllink-merge-confirm");
    if(ok) ok.disabled = sel.length!==2;
  }

  function bindButtons(){
    var btns = document.querySelectorAll(".urllink-to-merge");
    for(var i=0;i<btns.length;i++){
      btns[i].addEventListener("click", function(){
        if (typeof window.urllinkAddToMerge === "function") { window.urllinkAddToMerge(this); }
      });
    }
  }
  bindButtons();

  document.addEventListener("click", function(e){
    var t = e.target;
    var btn = t && t.closest ? t.closest(".urllink-to-merge") : null;
    if (btn) {
      if (typeof window.urllinkAddToMerge === "function") { window.urllinkAddToMerge(btn); }
    }
  });

  window.urllinkAddToMerge=function(btn){
      var idAttr = btn.getAttribute("data-merge-id");
      var id = idAttr ? parseInt(idAttr,10) : NaN;
      var title = btn.getAttribute("data-merge-title")||"";
      var pl = btn.getAttribute("data-merge-pl")||"";
      var pathVal = "";
      var extraVal = "";
      if(!id || !title || !pl){
        var tr = btn.closest ? btn.closest("tr") : null;
        if(tr){
          var tds = tr.querySelectorAll("td");
          if(tds && tds.length>=4){
            if(!id || isNaN(id)){
              var idText = (tds[1].textContent||"").trim();
              var idNum = parseInt(idText,10);
              if(!isNaN(idNum)) id = idNum;
            }
            if(!title){ title = (tds[2].textContent||"").trim(); }
            if(!pl){
              var a = tds[3].querySelector("a");
              pl = a ? (a.getAttribute("href")||a.textContent||"") : (tds[3].textContent||"").trim();
            }
          }
        }
      }
      (function(){
        var tr = btn.closest ? btn.closest("tr") : null;
        if(tr){
          var pathInput = tr.querySelector("input[name=\"path\"]");
          var extraTa = tr.querySelector("textarea[name=\"extra\"]");
          if(pathInput){ pathVal = pathInput.value || ""; }
          if(extraTa){ extraVal = extraTa.value || ""; }
        }
      })();
      if(!id || isNaN(id)) return;
      for(var i=0;i<sel.length;i++){ if(sel[i].id===id) { render(); return; } }
      var item = {id:id, title:title||("Post #"+id), pl:pl||"", path: pathVal||"", extra: extraVal||""};
      if(sel.length<2){ sel.push(item); } else { sel[1]=item; }
      render();
      var rowEl = document.getElementById(sel.length===1? "urllink-merge-row1" : "urllink-merge-row2");
      if(rowEl){ rowEl.style.transition="background 0.6s"; rowEl.style.background="#f0f9ff"; setTimeout(function(){ rowEl.style.background=""; }, 600); }
  };

  function clearSelection(targetSlots){
    if(!Array.isArray(targetSlots)){ sel=[]; render(); return; }
    targetSlots.sort();
    for(var i=targetSlots.length-1;i>=0;i--){
      var slot=targetSlots[i];
      if(slot>=0 && slot<sel.length){ sel.splice(slot,1); }
    }
    render();
  }

  var clr=document.getElementById("urllink-merge-clear");
  if(clr){ clr.addEventListener("click", function(){
      var checks=document.querySelectorAll(".urllink-merge-checkbox:checked");
      if(!checks.length){ sel=[]; render(); return; }
      var slots=[];
      checks.forEach(function(chk){
        var slot=parseInt(chk.getAttribute("data-slot"),10);
        if(!isNaN(slot)) { slots.push(slot); }
      });
      if(!slots.length){ sel=[]; render(); return; }
      clearSelection(slots);
  }); }
  render();
})();</script>

JS;
}

function urllink_admin_page() {
    urllink_render_block(false);
}
