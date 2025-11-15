<?php
if (!defined('ABSPATH')) { exit; }

add_action('rest_api_init', function(){
    register_rest_route('urllink/v1', '/preview', [
        'methods' => 'POST',
        'callback' => 'urllink_rest_preview',
        'permission_callback' => 'urllink_rest_can_manage',
        'args' => [
            'post_id' => [ 'type' => 'integer', 'required' => true ],
            'template' => [ 'type' => 'string', 'required' => true ],
        ]
    ]);

    register_rest_route('urllink/v1', '/update', [
        'methods' => 'POST',
        'callback' => 'urllink_rest_update',
        'permission_callback' => 'urllink_rest_can_manage',
        'args' => [
            'post_id' => [ 'type' => 'integer', 'required' => true ],
            'path' => [ 'type' => 'string', 'required' => true ],
            'extra' => [ 'type' => 'string', 'required' => false ],
        ]
    ]);

    register_rest_route('urllink/v1', '/map', [
        'methods' => 'GET',
        'callback' => function(){
            list($path_map, $post_map) = urllink_get_maps();
            return rest_ensure_response([ 'path_map' => $path_map, 'post_map' => $post_map ]);
        },
        'permission_callback' => 'urllink_rest_can_manage'
    ]);

    register_rest_route('urllink/v1', '/bulk-apply', [
        'methods' => 'POST',
        'callback' => 'urllink_rest_bulk_apply',
        'permission_callback' => 'urllink_rest_can_manage',
        'args' => [
            'post_type' => [ 'type' => 'string', 'required' => true ],
            'template' => [ 'type' => 'string', 'required' => true ],
            'limit' => [ 'type' => 'integer', 'required' => false, 'default' => 100 ],
            'offset' => [ 'type' => 'integer', 'required' => false, 'default' => 0 ],
            'dry_run' => [ 'type' => 'boolean', 'required' => false, 'default' => true ],
        ]
    ]);
});

function urllink_rest_can_manage() : bool {
    return current_user_can('manage_options');
}

function urllink_rest_preview(WP_REST_Request $req) {
    $post_id = intval($req->get_param('post_id'));
    $tpl = (string)$req->get_param('template');
    $post = get_post($post_id);
    if (!$post) return new WP_Error('not_found', 'Post not found', [ 'status' => 404 ]);
    $resolved = urllink_resolve_placeholders($post, $tpl);
    return rest_ensure_response([ 'resolved' => $resolved, 'permalink' => home_url('/' . $resolved . '/') ]);
}

function urllink_rest_update(WP_REST_Request $req) {
    $post_id = intval($req->get_param('post_id'));
    $path = urllink_normalize_path((string)$req->get_param('path'));
    $extra = (string)$req->get_param('extra');
    if (!$post_id || $path === '') return new WP_Error('bad_request', 'Bad request', [ 'status' => 400 ]);
    $old = get_post_meta($post_id, '_urllink_path', true);
    update_post_meta($post_id, '_urllink_path', $path);
    update_post_meta($post_id, '_urllink_extra_redirects', $extra);
    if ($old && $old !== $path) {
        $list = (array)get_post_meta($post_id, '_urllink_old_paths', true);
        if (!in_array($old, $list, true)) { $list[] = $old; }
        update_post_meta($post_id, '_urllink_old_paths', $list);
    }
    urllink_update_map_for_post($post_id, $path);
    return rest_ensure_response([ 'ok' => true, 'post_id' => $post_id, 'path' => $path, 'permalink' => get_permalink($post_id) ]);
}

function urllink_rest_bulk_apply(WP_REST_Request $req) {
    $pt = sanitize_key($req->get_param('post_type'));
    $tpl = (string)$req->get_param('template');
    $limit = max(1, intval($req->get_param('limit')));
    $offset = max(0, intval($req->get_param('offset')));
    $dry = filter_var($req->get_param('dry_run'), FILTER_VALIDATE_BOOLEAN);

    $q = new WP_Query([
        'post_type' => $pt ?: 'post',
        'posts_per_page' => $limit,
        'offset' => $offset,
        'post_status' => 'any',
        'fields' => 'ids'
    ]);

    $out = [];
    foreach ($q->posts as $pid) {
        $post = get_post($pid);
        $resolved = urllink_resolve_placeholders($post, $tpl);
        $normalized = urllink_normalize_path($resolved);
        // 冲突检测
        list($path_map) = urllink_get_maps();
        $conflict = isset($path_map[$normalized]) && intval($path_map[$normalized]) !== intval($pid);
        if (!$dry && !$conflict) {
            $old = get_post_meta($pid, '_urllink_path', true);
            update_post_meta($pid, '_urllink_path', $normalized);
            if ($old && $old !== $normalized) {
                $list = (array)get_post_meta($pid, '_urllink_old_paths', true);
                if (!in_array($old, $list, true)) { $list[] = $old; }
                update_post_meta($pid, '_urllink_old_paths', $list);
            }
            urllink_update_map_for_post($pid, $normalized);
        }
        $out[] = [
            'post_id' => intval($pid),
            'preview' => $normalized,
            'conflict' => $conflict,
            'current' => (string)get_post_meta($pid, '_urllink_path', true),
        ];
    }

    return rest_ensure_response([
        'dry_run' => $dry,
        'items' => $out,
        'count' => count($out),
        'next_offset' => $offset + count($out)
    ]);
}
