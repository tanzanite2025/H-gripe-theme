<?php

defined('ABSPATH') || exit;

class TZ_BLOG_REST
{
    const META_GROUP = 'tz_translation_group';

    public function __construct()
    {
        add_action('rest_api_init', array($this, 'register_routes'));
    }

    public function register_routes()
    {
        register_rest_route(
            'tanzanite/v1',
            '/posts',
            array(
                'methods' => 'GET',
                'callback' => array($this, 'handle_posts'),
                'permission_callback' => '__return_true'
            )
        );

        register_rest_route(
            'tanzanite/v1',
            '/post',
            array(
                'methods' => 'GET',
                'callback' => array($this, 'handle_post'),
                'permission_callback' => '__return_true'
            )
        );

        register_rest_route(
            'tanzanite/v1',
            '/translations',
            array(
                'methods' => 'GET',
                'callback' => array($this, 'handle_translations'),
                'permission_callback' => '__return_true'
            )
        );
    }

    private function build_featured_image($post_id)
    {
        $thumb_id = get_post_thumbnail_id($post_id);
        if (empty($thumb_id)) {
            return null;
        }

        $src = wp_get_attachment_image_src($thumb_id, 'full');
        if (!is_array($src) || empty($src[0])) {
            return null;
        }

        $alt = get_post_meta($thumb_id, '_wp_attachment_image_alt', true);

        return array(
            'url' => $src[0],
            'width' => isset($src[1]) ? (int) $src[1] : null,
            'height' => isset($src[2]) ? (int) $src[2] : null,
            'alt' => is_string($alt) ? $alt : ''
        );
    }

    private function get_post_lang($post_id)
    {
        $slugs = wp_get_post_terms($post_id, TZ_BLOG_Languages::TAXONOMY, array('fields' => 'slugs'));
        if (is_wp_error($slugs) || empty($slugs)) {
            return '';
        }

        return (string) $slugs[0];
    }

    private function get_translations_map($group)
    {
        if (empty($group)) {
            return array();
        }

        $cache_base = 'translations_' . md5($group);
        $cached = TZ_BLOG_Cache::get($cache_base);
        if (is_array($cached)) {
            return $cached;
        }

        $posts = get_posts(
            array(
                'post_type' => 'post',
                'post_status' => 'publish',
                'numberposts' => -1,
                'meta_key' => self::META_GROUP,
                'meta_value' => $group,
                'fields' => 'ids'
            )
        );

        $map = array();
        foreach ($posts as $id) {
            $lang = $this->get_post_lang($id);
            if (empty($lang)) {
                continue;
            }

            $map[$lang] = array(
                'id' => (int) $id,
                'slug' => (string) get_post_field('post_name', $id)
            );
        }

        TZ_BLOG_Cache::set($cache_base, $map);
        return $map;
    }

    private function format_post($post)
    {
        $group = (string) get_post_meta($post->ID, self::META_GROUP, true);

        return array(
            'id' => (int) $post->ID,
            'lang' => $this->get_post_lang($post->ID),
            'group' => $group,
            'slug' => (string) $post->post_name,
            'title' => (string) get_the_title($post),
            'excerpt' => (string) get_the_excerpt($post),
            'date' => (string) get_post_time('c', true, $post),
            'featuredImage' => $this->build_featured_image($post->ID),
            'categories' => wp_get_post_terms($post->ID, 'category', array('fields' => 'slugs')),
            'translations' => $this->get_translations_map($group)
        );
    }

    private function get_common_query_args($lang, $category)
    {
        $lang = sanitize_key((string) $lang);
        if (!TZ_BLOG_Languages::is_valid_locale($lang)) {
            return new WP_Error('tz_blog_invalid_lang', 'Invalid lang', array('status' => 400));
        }

        $tax_query = array(
            array(
                'taxonomy' => TZ_BLOG_Languages::TAXONOMY,
                'field' => 'slug',
                'terms' => array($lang)
            )
        );

        $cat_ids = TZ_BLOG_Setup::get_allowed_category_ids();
        if (!empty($category)) {
            $category = sanitize_key((string) $category);
            if (!TZ_BLOG_Setup::is_valid_category_slug($category)) {
                return new WP_Error('tz_blog_invalid_category', 'Invalid category', array('status' => 400));
            }

            return array($lang, $tax_query, array('category_name' => $category));
        }

        return array($lang, $tax_query, array('category__in' => $cat_ids));
    }

    public function handle_posts($request)
    {
        $lang = $request->get_param('lang');
        $category = $request->get_param('category');

        $page = (int) $request->get_param('page');
        $per_page = (int) $request->get_param('per_page');

        $page = $page > 0 ? $page : 1;
        $per_page = $per_page > 0 ? min($per_page, 50) : 10;

        $common = $this->get_common_query_args($lang, $category);
        if (is_wp_error($common)) {
            return $common;
        }

        list($lang, $tax_query, $cat_filter) = $common;

        $cache_base = 'posts_' . md5($lang . '|' . (string) $category . '|' . $page . '|' . $per_page);
        $cached = TZ_BLOG_Cache::get($cache_base);
        if (is_array($cached)) {
            return $cached;
        }

        $query = new WP_Query(
            array_merge(
                array(
                    'post_type' => 'post',
                    'post_status' => 'publish',
                    'posts_per_page' => $per_page,
                    'paged' => $page,
                    'orderby' => 'date',
                    'order' => 'DESC',
                    'tax_query' => $tax_query
                ),
                $cat_filter
            )
        );

        $items = array();
        foreach ($query->posts as $post) {
            if (!($post instanceof WP_Post)) {
                continue;
            }

            $items[] = $this->format_post($post);
        }

        $payload = array(
            'page' => $page,
            'per_page' => $per_page,
            'total' => (int) $query->found_posts,
            'items' => $items
        );

        TZ_BLOG_Cache::set($cache_base, $payload);
        return $payload;
    }

    public function handle_post($request)
    {
        $lang = $request->get_param('lang');
        $slug = $request->get_param('slug');

        $lang = sanitize_key((string) $lang);
        $slug = sanitize_title((string) $slug);

        if (!TZ_BLOG_Languages::is_valid_locale($lang)) {
            return new WP_Error('tz_blog_invalid_lang', 'Invalid lang', array('status' => 400));
        }

        if (empty($slug)) {
            return new WP_Error('tz_blog_invalid_slug', 'Invalid slug', array('status' => 400));
        }

        $cache_base = 'post_' . md5($lang . '|' . $slug);
        $cached = TZ_BLOG_Cache::get($cache_base);
        if (is_array($cached)) {
            return $cached;
        }

        $query = new WP_Query(
            array(
                'post_type' => 'post',
                'post_status' => 'publish',
                'name' => $slug,
                'posts_per_page' => 1,
                'tax_query' => array(
                    array(
                        'taxonomy' => TZ_BLOG_Languages::TAXONOMY,
                        'field' => 'slug',
                        'terms' => array($lang)
                    )
                ),
                'category__in' => TZ_BLOG_Setup::get_allowed_category_ids()
            )
        );

        if (empty($query->posts) || !($query->posts[0] instanceof WP_Post)) {
            return new WP_Error('tz_blog_not_found', 'Not found', array('status' => 404));
        }

        $post = $query->posts[0];

        $payload = $this->format_post($post);
        $payload['contentHtml'] = (string) apply_filters('the_content', $post->post_content);
        $payload['canonicalUrl'] = (string) get_permalink($post);

        TZ_BLOG_Cache::set($cache_base, $payload);
        return $payload;
    }

    public function handle_translations($request)
    {
        $group = (string) $request->get_param('group');
        $group = sanitize_text_field($group);

        if (empty($group)) {
            return new WP_Error('tz_blog_invalid_group', 'Invalid group', array('status' => 400));
        }

        return array(
            'group' => $group,
            'translations' => $this->get_translations_map($group)
        );
    }
}
