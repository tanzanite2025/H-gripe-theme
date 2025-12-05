<?php

namespace TanzanitePhotoGallery;

if (! defined('ABSPATH')) {
    exit;
}

class TPG_Post_Type {
    private static ?TPG_Post_Type $instance = null;

    public static function instance(): TPG_Post_Type {
        if (self::$instance === null) {
            self::$instance = new self();
        }
        return self::$instance;
    }

    private function __construct() {
        add_action('init', [$this, 'register_post_type']);
        add_action('init', [$this, 'register_meta']);
    }

    public function register_post_type(): void {
        $labels = [
            'name'               => __('Tanzanite Photos', 'tanzanite-photo-gallery'),
            'singular_name'      => __('Tanzanite Photo', 'tanzanite-photo-gallery'),
            'add_new'            => __('Add New', 'tanzanite-photo-gallery'),
            'add_new_item'       => __('Add New Photo', 'tanzanite-photo-gallery'),
            'edit_item'          => __('Edit Photo', 'tanzanite-photo-gallery'),
            'new_item'           => __('New Photo', 'tanzanite-photo-gallery'),
            'view_item'          => __('View Photo', 'tanzanite-photo-gallery'),
            'search_items'       => __('Search Photos', 'tanzanite-photo-gallery'),
            'not_found'          => __('No photos found', 'tanzanite-photo-gallery'),
            'not_found_in_trash' => __('No photos found in Trash', 'tanzanite-photo-gallery'),
            'all_items'          => __('All Photos', 'tanzanite-photo-gallery'),
        ];

        $args = [
            'labels'             => $labels,
            'public'             => false,
            'show_ui'            => true,
            'show_in_menu'       => false, // Admin class will attach it to its own menu item.
            'capability_type'    => 'post',
            'map_meta_cap'       => true,
            'supports'           => ['title', 'editor', 'thumbnail', 'author'],
            'rewrite'            => false,
        ];

        register_post_type('tanz_photo', $args);
    }

    public function register_meta(): void {
        $meta_definitions = [
            'tanz_photo_type' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_image_id' => [
                'type'         => 'integer',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_region' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_location' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_nickname' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_bike_model' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_notes' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_status' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
                'default'      => 'pending',
            ],
            'tanz_photo_submitted_at' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_approved_at' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_rejected_reason' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => false,
            ],
            'tanz_photo_source' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
            'tanz_photo_product_refs' => [
                'type'         => 'string',
                'single'       => true,
                'show_in_rest' => true,
            ],
        ];

        foreach ($meta_definitions as $key => $args) {
            register_post_meta('tanz_photo', $key, $args);
        }
    }
}
