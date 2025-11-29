<?php

namespace TanzanitePhotoGallery;

if (! defined('ABSPATH')) {
    exit;
}

class TPG_Admin {
    private static ?TPG_Admin $instance = null;

    public static function instance(): TPG_Admin {
        if (self::$instance === null) {
            self::$instance = new self();
        }
        return self::$instance;
    }

    private function __construct() {
        add_action('admin_menu', [$this, 'register_menu']);
        add_filter('manage_tanz_photo_posts_columns', [$this, 'filter_columns']);
        add_action('manage_tanz_photo_posts_custom_column', [$this, 'render_column'], 10, 2);
        add_action('add_meta_boxes', [$this, 'register_meta_boxes']);
        add_action('save_post_tanz_photo', [$this, 'save_product_refs_meta']);
    }

    public function register_menu(): void {
        add_menu_page(
            __('Tanzanite Photos', 'tanzanite-photo-gallery'),
            __('Tanzanite Photos', 'tanzanite-photo-gallery'),
            'edit_posts',
            'edit.php?post_type=tanz_photo',
            '',
            'dashicons-format-gallery',
            58
        );
    }

    public function filter_columns(array $columns): array {
        // Keep the checkbox and title columns, add Type / Region / Status.
        $new = [];
        if (isset($columns['cb'])) {
            $new['cb'] = $columns['cb'];
        }

        $new['title']  = __('Title', 'tanzanite-photo-gallery');
        $new['type']   = __('Type', 'tanzanite-photo-gallery');
        $new['region'] = __('Region', 'tanzanite-photo-gallery');
        $new['status'] = __('Status', 'tanzanite-photo-gallery');

        if (isset($columns['date'])) {
            $new['date'] = $columns['date'];
        }

        return $new;
    }

    public function render_column(string $column, int $post_id): void {
        switch ($column) {
            case 'type':
                $type = get_post_meta($post_id, 'tanz_photo_type', true) ?: '-';
                echo esc_html(ucfirst($type));
                break;
            case 'region':
                $region   = get_post_meta($post_id, 'tanz_photo_region', true);
                $location = get_post_meta($post_id, 'tanz_photo_location', true);
                $parts    = array_filter([$region, $location]);
                echo esc_html($parts ? implode(' · ', $parts) : '-');
                break;
            case 'status':
                $status = get_post_meta($post_id, 'tanz_photo_status', true) ?: 'pending';
                echo esc_html(ucfirst($status));
                break;
        }
    }

    public function register_meta_boxes(): void {
        add_meta_box(
            'tpg_product_refs',
            __('Recommended build links', 'tanzanite-photo-gallery'),
            [$this, 'render_product_refs_meta_box'],
            'tanz_photo',
            'side',
            'default'
        );
    }

    public function render_product_refs_meta_box(\WP_Post $post): void {
        wp_nonce_field('tpg_save_product_refs', 'tpg_product_refs_nonce');

        $raw   = get_post_meta($post->ID, 'tanz_photo_product_refs', true);
        $data  = [];

        if (is_string($raw) && $raw !== '') {
            $decoded = json_decode($raw, true);
            if (is_array($decoded)) {
                $data = $decoded;
            }
        } elseif (is_array($raw)) {
            $data = $raw;
        }

        $rim   = isset($data['rim']) && is_string($data['rim']) ? $data['rim'] : '';
        $wheel = isset($data['wheel']) && is_string($data['wheel']) ? $data['wheel'] : '';
        $hub   = isset($data['hub']) && is_string($data['hub']) ? $data['hub'] : '';
        $tire  = isset($data['tire']) && is_string($data['tire']) ? $data['tire'] : '';

        ?>
        <p>
            <label for="tpg_product_rim"><?php esc_html_e('Rim product (slug or ID)', 'tanzanite-photo-gallery'); ?></label>
            <input
                type="text"
                id="tpg_product_rim"
                name="tpg_product_rim"
                value="<?php echo esc_attr($rim); ?>"
                class="widefat"
            />
        </p>
        <p>
            <label for="tpg_product_wheel"><?php esc_html_e('Wheel(s) product (slug or ID)', 'tanzanite-photo-gallery'); ?></label>
            <input
                type="text"
                id="tpg_product_wheel"
                name="tpg_product_wheel"
                value="<?php echo esc_attr($wheel); ?>"
                class="widefat"
            />
        </p>
        <p>
            <label for="tpg_product_hub"><?php esc_html_e('Hub product (slug or ID)', 'tanzanite-photo-gallery'); ?></label>
            <input
                type="text"
                id="tpg_product_hub"
                name="tpg_product_hub"
                value="<?php echo esc_attr($hub); ?>"
                class="widefat"
            />
        </p>
        <p>
            <label for="tpg_product_tire"><?php esc_html_e('Tire product (slug or ID)', 'tanzanite-photo-gallery'); ?></label>
            <input
                type="text"
                id="tpg_product_tire"
                name="tpg_product_tire"
                value="<?php echo esc_attr($tire); ?>"
                class="widefat"
            />
        </p>
        <?php
    }

    public function save_product_refs_meta(int $post_id): void {
        if (! isset($_POST['tpg_product_refs_nonce']) || ! wp_verify_nonce(
            sanitize_text_field(wp_unslash($_POST['tpg_product_refs_nonce'])),
            'tpg_save_product_refs'
        )) {
            return;
        }

        if (defined('DOING_AUTOSAVE') && DOING_AUTOSAVE) {
            return;
        }

        if (! current_user_can('edit_post', $post_id)) {
            return;
        }

        $rim   = isset($_POST['tpg_product_rim']) ? sanitize_text_field(wp_unslash($_POST['tpg_product_rim'])) : '';
        $wheel = isset($_POST['tpg_product_wheel']) ? sanitize_text_field(wp_unslash($_POST['tpg_product_wheel'])) : '';
        $hub   = isset($_POST['tpg_product_hub']) ? sanitize_text_field(wp_unslash($_POST['tpg_product_hub'])) : '';
        $tire  = isset($_POST['tpg_product_tire']) ? sanitize_text_field(wp_unslash($_POST['tpg_product_tire'])) : '';

        $refs = [];

        if ($rim !== '') {
            $refs['rim'] = $rim;
        }
        if ($wheel !== '') {
            $refs['wheel'] = $wheel;
        }
        if ($hub !== '') {
            $refs['hub'] = $hub;
        }
        if ($tire !== '') {
            $refs['tire'] = $tire;
        }

        if (! $refs) {
            delete_post_meta($post_id, 'tanz_photo_product_refs');
            return;
        }

        update_post_meta($post_id, 'tanz_photo_product_refs', wp_json_encode($refs));
    }
}
