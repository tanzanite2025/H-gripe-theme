<?php

defined('ABSPATH') || exit;

class TZ_BLOG_Setup
{
    public function __construct()
    {
        add_action('init', array(__CLASS__, 'ensure_categories'));
    }

    public static function get_category_slugs()
    {
        return array('news', 'wheelsbuild');
    }

    public static function ensure_categories()
    {
        $categories = array(
            'news' => 'News',
            'wheelsbuild' => 'Wheelsbuild'
        );

        foreach ($categories as $slug => $name) {
            $existing = get_category_by_slug($slug);
            if ($existing instanceof WP_Term) {
                continue;
            }

            wp_insert_term(
                $name,
                'category',
                array(
                    'slug' => $slug
                )
            );
        }
    }

    public static function get_allowed_category_ids()
    {
        $ids = array();
        foreach (self::get_category_slugs() as $slug) {
            $term = get_category_by_slug($slug);
            if ($term instanceof WP_Term) {
                $ids[] = (int) $term->term_id;
            }
        }

        return $ids;
    }

    public static function is_valid_category_slug($slug)
    {
        return in_array($slug, self::get_category_slugs(), true);
    }
}
