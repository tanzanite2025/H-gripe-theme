<?php

namespace TanzanitePhotoGallery;

use WP_Comment_Query;
use WP_Error;
use WP_Query;
use WP_REST_Request;
use WP_REST_Response;
use WP_REST_Server;

if (! defined('ABSPATH')) {
    exit;
}

class TPG_REST {
    private static ?TPG_REST $instance = null;

    public static function instance(): TPG_REST {
        if (self::$instance === null) {
            self::$instance = new self();
        }
        return self::$instance;
    }

    private function __construct() {
        add_action('rest_api_init', [$this, 'register_routes']);
    }

    public function register_routes(): void {
        register_rest_route(
            'tanz-photo/v1',
            '/gallery',
            [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'handle_gallery'],
                'permission_callback' => '__return_true',
                'args'                => [
                    'type' => [
                        'type'    => 'string',
                        'default' => 'brand',
                        'enum'    => ['user', 'brand', 'all'],
                    ],
                    'status' => [
                        'type'    => 'string',
                        'default' => 'approved',
                        'enum'    => ['pending', 'approved', 'rejected', 'all'],
                    ],
                    'page' => [
                        'type'    => 'integer',
                        'default' => 1,
                        'minimum' => 1,
                    ],
                    'per_page' => [
                        'type'    => 'integer',
                        'default' => 24,
                        'minimum' => 1,
                        'maximum' => 100,
                    ],
                ],
            ]
        );

        register_rest_route(
            'tanz-photo/v1',
            '/upload',
            [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'handle_upload'],
                'permission_callback' => [$this, 'can_upload'],
            ]
        );

        register_rest_route(
            'tanz-photo/v1',
            '/review',
            [
                'methods'             => WP_REST_Server::EDITABLE,
                'callback'            => [$this, 'handle_review'],
                'permission_callback' => [$this, 'can_review'],
            ]
        );

        register_rest_route(
            'tanz-photo/v1',
            '/comments',
            [
                'methods'             => WP_REST_Server::READABLE,
                'callback'            => [$this, 'handle_comments_get'],
                'permission_callback' => '__return_true',
                'args'                => [
                    'photo_id' => [
                        'type'     => 'integer',
                        'required' => true,
                        'minimum'  => 1,
                    ],
                    'page' => [
                        'type'    => 'integer',
                        'default' => 1,
                        'minimum' => 1,
                    ],
                    'per_page' => [
                        'type'    => 'integer',
                        'default' => 20,
                        'minimum' => 1,
                        'maximum' => 100,
                    ],
                ],
            ]
        );

        register_rest_route(
            'tanz-photo/v1',
            '/comments',
            [
                'methods'             => WP_REST_Server::CREATABLE,
                'callback'            => [$this, 'handle_comments_create'],
                'permission_callback' => [$this, 'can_comment'],
                'args'                => [
                    'photo_id' => [
                        'type'     => 'integer',
                        'required' => true,
                        'minimum'  => 1,
                    ],
                    'content' => [
                        'type'     => 'string',
                        'required' => true,
                    ],
                    'location' => [
                        'type' => 'string',
                    ],
                ],
            ]
        );
    }

    public function handle_gallery(WP_REST_Request $request): WP_REST_Response|WP_Error {
        $type      = $request->get_param('type') ?: 'brand';
        $status    = $request->get_param('status') ?: 'approved';
        $page      = max(1, (int) ($request->get_param('page') ?: 1));
        $per_page  = (int) ($request->get_param('per_page') ?: 24);
        $per_page  = max(1, min(100, $per_page));

        $meta_query = [];

        if ($type && $type !== 'all') {
            $meta_query[] = [
                'key'   => 'tanz_photo_type',
                'value' => $type,
            ];
        }

        if ($status && $status !== 'all') {
            $meta_query[] = [
                'key'   => 'tanz_photo_status',
                'value' => $status,
            ];
        }

        $orderby  = 'date';
        $meta_key = '';

        // 如果是已审核通过的列表，优先按审核时间倒序排序（更符合前端期望）。
        if ($status === 'approved') {
            $orderby  = 'meta_value';
            $meta_key = 'tanz_photo_approved_at';
        }

        $query_args = [
            'post_type'      => 'tanz_photo',
            'post_status'    => 'publish',
            'posts_per_page' => $per_page,
            'paged'          => $page,
            'meta_query'     => $meta_query,
            'orderby'        => $orderby,
            'order'          => 'DESC',
        ];

        if ($meta_key) {
            $query_args['meta_key'] = $meta_key;
        }

        $query = new WP_Query($query_args);

        $items = [];

        foreach ($query->posts as $post) {
            $post_id  = $post->ID;
            $region   = get_post_meta($post_id, 'tanz_photo_region', true);
            $location = get_post_meta($post_id, 'tanz_photo_location', true);
            $nickname = get_post_meta($post_id, 'tanz_photo_nickname', true);

            $product_refs      = [];
            $raw_product_refs  = get_post_meta($post_id, 'tanz_photo_product_refs', true);
            $decoded_product   = null;

            if (is_string($raw_product_refs) && $raw_product_refs !== '') {
                $decoded = json_decode($raw_product_refs, true);
                if (is_array($decoded)) {
                    $decoded_product = $decoded;
                }
            } elseif (is_array($raw_product_refs)) {
                // 容错：如果历史数据已经是数组，则直接使用
                $decoded_product = $raw_product_refs;
            }

            if (is_array($decoded_product)) {
                foreach (['rim', 'wheel', 'hub', 'tire'] as $key) {
                    if (isset($decoded_product[$key]) && is_string($decoded_product[$key]) && $decoded_product[$key] !== '') {
                        $product_refs[$key] = $decoded_product[$key];
                    }
                }
            }

            $items[] = [
                'id'           => $post_id,
                'title'        => get_the_title($post_id),
                'region'       => $region ?: 'Studio',
                'location'     => $location ?: '',
                'nickname'     => $nickname ?: '',
                'type'         => get_post_meta($post_id, 'tanz_photo_type', true) ?: '',
                'status'       => get_post_meta($post_id, 'tanz_photo_status', true) ?: 'pending',
                'product_refs' => $product_refs,
            ];
        }

        $response = new WP_REST_Response($items, 200);
        $response->header('X-WP-Total', (int) $query->found_posts);
        $response->header('X-WP-TotalPages', (int) $query->max_num_pages);

        return $response;
    }

    public function handle_upload(WP_REST_Request $request): WP_REST_Response|WP_Error {
        if (! is_user_logged_in() || ! current_user_can('upload_files')) {
            return new WP_Error(
                'tpg_forbidden',
                __('You must be logged in to upload photos.', 'tanzanite-photo-gallery'),
                ['status' => 403]
            );
        }

        $user_id = get_current_user_id();

        // 简单的基础防 spam：按用户做日限额和短时间间隔限制
        $today_key   = 'tanz_photo_daily_' . gmdate('Ymd');
        $daily_count = (int) get_user_meta($user_id, $today_key, true);
        $daily_limit = 20; // 每位用户每日最多 20 次上传

        if ($daily_count >= $daily_limit) {
            return new WP_Error(
                'tpg_rate_limited',
                __('You have reached today’s upload limit.', 'tanzanite-photo-gallery'),
                ['status' => 429]
            );
        }

        $last_ts   = (int) get_user_meta($user_id, 'tanz_photo_last_upload', true);
        $min_delay = 30; // 两次上传至少间隔 30 秒
        if ($last_ts && (time() - $last_ts) < $min_delay) {
            return new WP_Error(
                'tpg_rate_limited',
                __('You are uploading too fast. Please wait a moment.', 'tanzanite-photo-gallery'),
                ['status' => 429]
            );
        }

        $files = $request->get_file_params();
        if (empty($files['file']) || empty($files['file']['name'])) {
            return new WP_Error(
                'tpg_no_file',
                __('No file was uploaded.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        $file = $files['file'];

        // 文件大小限制（约 5MB）
        $max_size = 5 * 1024 * 1024;
        if (! empty($file['size']) && (int) $file['size'] > $max_size) {
            return new WP_Error(
                'tpg_file_too_large',
                __('Uploaded file is too large.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        // 仅允许 WEBP
        $check = wp_check_filetype_and_ext($file['tmp_name'], $file['name']);
        if (empty($check['type']) || $check['type'] !== 'image/webp') {
            return new WP_Error(
                'tpg_invalid_type',
                __('Only WEBP images are allowed.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        // 需要使用媒体相关函数
        if (! function_exists('media_handle_upload')) {
            require_once ABSPATH . 'wp-admin/includes/file.php';
            require_once ABSPATH . 'wp-admin/includes/media.php';
            require_once ABSPATH . 'wp-admin/includes/image.php';
        }

        // 使用 WordPress 媒体处理上传
        $attachment_id = media_handle_upload('file', 0);

        if (is_wp_error($attachment_id)) {
            return new WP_Error(
                'tpg_upload_failed',
                __('Could not save uploaded file.', 'tanzanite-photo-gallery'),
                ['status' => 500]
            );
        }

        // 使用图像编辑器将最长边压缩到 800px 以内
        $file_path = get_attached_file($attachment_id);
        if ($file_path && file_exists($file_path)) {
            $editor = wp_get_image_editor($file_path);
            if (! is_wp_error($editor)) {
                $size = $editor->get_size();
                if (! empty($size['width']) && ! empty($size['height'])) {
                    $width    = (int) $size['width'];
                    $height   = (int) $size['height'];
                    $max_side = max($width, $height);
                    $limit    = 800;

                    if ($max_side > $limit) {
                        $scale = $limit / $max_side;
                        $new_w = (int) floor($width * $scale);
                        $new_h = (int) floor($height * $scale);

                        $editor->resize($new_w, $new_h, false);
                        $saved = $editor->save($file_path);

                        if (! is_wp_error($saved)) {
                            $meta = wp_generate_attachment_metadata($attachment_id, $file_path);
                            wp_update_attachment_metadata($attachment_id, $meta);
                        }
                    }
                }
            }
        }

        // 准备元数据
        $region   = sanitize_text_field((string) $request->get_param('region'));
        $location = sanitize_text_field((string) $request->get_param('location'));
        $nickname = sanitize_text_field((string) $request->get_param('nickname'));
        $bike     = sanitize_text_field((string) $request->get_param('bike_model'));
        $notes    = sanitize_textarea_field((string) $request->get_param('notes'));

        if ($region === '') {
            // 删除附件，避免遗留垃圾数据
            wp_delete_attachment($attachment_id, true);
            return new WP_Error(
                'tpg_missing_region',
                __('Region is required.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        $title = $region;
        if ($nickname !== '') {
            $title = $nickname . ' – ' . $region;
        }

        $post_data = [
            'post_type'   => 'tanz_photo',
            'post_status' => 'publish',
            'post_title'  => $title,
            'post_author' => $user_id,
        ];

        $post_id = wp_insert_post($post_data, true);
        if (is_wp_error($post_id)) {
            wp_delete_attachment($attachment_id, true);
            return new WP_Error(
                'tpg_create_failed',
                __('Could not create photo record.', 'tanzanite-photo-gallery'),
                ['status' => 500]
            );
        }

        $now = current_time('mysql', true);

        update_post_meta($post_id, 'tanz_photo_type', 'user');
        update_post_meta($post_id, 'tanz_photo_status', 'pending');
        update_post_meta($post_id, 'tanz_photo_image_id', $attachment_id);
        update_post_meta($post_id, 'tanz_photo_region', $region);
        update_post_meta($post_id, 'tanz_photo_location', $location);
        update_post_meta($post_id, 'tanz_photo_nickname', $nickname);
        update_post_meta($post_id, 'tanz_photo_bike_model', $bike);
        update_post_meta($post_id, 'tanz_photo_notes', $notes);
        update_post_meta($post_id, 'tanz_photo_submitted_at', $now);
        update_post_meta($post_id, 'tanz_photo_source', 'web-form');

        // 更新防 spam 计数
        update_user_meta($user_id, $today_key, $daily_count + 1);
        update_user_meta($user_id, 'tanz_photo_last_upload', time());

        $data = [
            'id'     => $post_id,
            'status' => 'pending',
        ];

        return new WP_REST_Response($data, 201);
    }

    public function handle_review(WP_REST_Request $request): WP_REST_Response|WP_Error {
        $post_id = (int) $request->get_param('id');
        if ($post_id <= 0) {
            return new WP_Error(
                'tpg_invalid_id',
                __('Invalid photo ID.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        $post = get_post($post_id);
        if (! $post || $post->post_type !== 'tanz_photo') {
            return new WP_Error(
                'tpg_not_found',
                __('Photo not found.', 'tanzanite-photo-gallery'),
                ['status' => 404]
            );
        }

        $status = (string) $request->get_param('status');
        if ($status === '') {
            $status = 'approved';
        }

        $allowed = ['approved', 'rejected', 'pending'];
        if (! in_array($status, $allowed, true)) {
            return new WP_Error(
                'tpg_invalid_status',
                __('Invalid review status.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        $now = current_time('mysql', true);

        if ($status === 'approved') {
            update_post_meta($post_id, 'tanz_photo_status', 'approved');
            update_post_meta($post_id, 'tanz_photo_approved_at', $now);
            delete_post_meta($post_id, 'tanz_photo_rejected_reason');
        } elseif ($status === 'rejected') {
            update_post_meta($post_id, 'tanz_photo_status', 'rejected');
            $reason = sanitize_text_field((string) $request->get_param('reason'));
            if ($reason !== '') {
                update_post_meta($post_id, 'tanz_photo_rejected_reason', $reason);
            }
        } else { // pending
            update_post_meta($post_id, 'tanz_photo_status', 'pending');
            // 不强制清理 approved_at/rejected_reason，让管理员自行决定是否覆盖
        }

        $data = [
            'id'          => $post_id,
            'status'      => $status,
            'approved_at' => $status === 'approved' ? $now : '',
        ];

        return new WP_REST_Response($data, 200);
    }

    public function handle_comments_get(WP_REST_Request $request): WP_REST_Response|WP_Error {
        $photo_id = (int) $request->get_param('photo_id');
        if ($photo_id <= 0) {
            return new WP_Error(
                'tpg_invalid_photo_id',
                __('Invalid photo ID.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        $post = get_post($photo_id);
        if (! $post || $post->post_type !== 'tanz_photo') {
            return new WP_Error(
                'tpg_photo_not_found',
                __('Photo not found.', 'tanzanite-photo-gallery'),
                ['status' => 404]
            );
        }

        $page     = max(1, (int) $request->get_param('page'));
        $per_page = (int) $request->get_param('per_page');
        $per_page = max(1, min(100, $per_page ?: 20));

        $query_args = [
            'post_id' => $photo_id,
            'type'    => 'tanz_photo_comment',
            'status'  => 'approve',
            'number'  => $per_page,
            'paged'   => $page,
            'orderby' => 'comment_date_gmt',
            'order'   => 'DESC',
        ];

        $query    = new WP_Comment_Query();
        $comments = $query->query($query_args);

        $items = [];

        foreach ($comments as $comment) {
            $location = get_comment_meta($comment->comment_ID, 'tanz_comment_location', true);

            $items[] = [
                'id'        => (int) $comment->comment_ID,
                'author'    => (string) $comment->comment_author,
                'content'   => (string) wp_kses_post($comment->comment_content),
                'date_gmt'  => (string) $comment->comment_date_gmt,
                'location'  => $location ? (string) $location : '',
            ];
        }

        $total       = (int) $query->found_comments;
        $total_pages = $per_page > 0 ? (int) ceil($total / $per_page) : 1;

        $response = new WP_REST_Response($items, 200);
        $response->header('X-WP-Total-Comments', $total);
        $response->header('X-WP-TotalPages', $total_pages);

        return $response;
    }

    public function handle_comments_create(WP_REST_Request $request): WP_REST_Response|WP_Error {
        if (! is_user_logged_in()) {
            return new WP_Error(
                'tpg_forbidden',
                __('You must be logged in to comment.', 'tanzanite-photo-gallery'),
                ['status' => 403]
            );
        }

        $photo_id = (int) $request->get_param('photo_id');
        if ($photo_id <= 0) {
            return new WP_Error(
                'tpg_invalid_photo_id',
                __('Invalid photo ID.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        $post = get_post($photo_id);
        if (! $post || $post->post_type !== 'tanz_photo') {
            return new WP_Error(
                'tpg_photo_not_found',
                __('Photo not found.', 'tanzanite-photo-gallery'),
                ['status' => 404]
            );
        }

        $content = trim((string) $request->get_param('content'));
        if ($content === '') {
            return new WP_Error(
                'tpg_empty_comment',
                __('Comment content cannot be empty.', 'tanzanite-photo-gallery'),
                ['status' => 400]
            );
        }

        $user = wp_get_current_user();
        if (! $user || ! $user->ID) {
            return new WP_Error(
                'tpg_forbidden',
                __('You must be logged in to comment.', 'tanzanite-photo-gallery'),
                ['status' => 403]
            );
        }

        $location = sanitize_text_field((string) $request->get_param('location'));

        $commentdata = [
            'comment_post_ID'      => $photo_id,
            'comment_content'      => $content,
            'comment_type'         => 'tanz_photo_comment',
            'comment_author'       => $user->display_name ?: $user->user_login,
            'comment_author_email' => $user->user_email,
            'user_id'              => $user->ID,
            'comment_approved'     => 0, // 默认需要审核
        ];

        $comment_id = wp_insert_comment($commentdata, true);
        if (is_wp_error($comment_id)) {
            return new WP_Error(
                'tpg_comment_failed',
                __('Could not create comment.', 'tanzanite-photo-gallery'),
                ['status' => 500]
            );
        }

        if ($location !== '') {
            update_comment_meta($comment_id, 'tanz_comment_location', $location);
        }

        $data = [
            'id'     => (int) $comment_id,
            'status' => 'pending',
        ];

        return new WP_REST_Response($data, 201);
    }

    public function can_upload(): bool {
        return current_user_can('upload_files');
    }

    public function can_review(): bool {
        return current_user_can('edit_others_posts');
    }

    public function can_comment(): bool {
        return is_user_logged_in() && current_user_can('read');
    }
}
