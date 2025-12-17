<?php

defined('ABSPATH') || exit;

class TZ_BLOG_Cache
{
    const VERSION_OPTION = 'tz_blog_i18n_cache_version';

    public function __construct()
    {
        add_action('save_post', array(__CLASS__, 'maybe_bump_on_save'), 10, 2);
        add_action('deleted_post', array(__CLASS__, 'bump_version'));
        add_action('trashed_post', array(__CLASS__, 'bump_version'));
        add_action('untrashed_post', array(__CLASS__, 'bump_version'));
    }

    public static function get_version()
    {
        $version = get_option(self::VERSION_OPTION);
        if (empty($version)) {
            $version = '1';
            update_option(self::VERSION_OPTION, $version, false);
        }

        return (string) $version;
    }

    public static function bump_version()
    {
        update_option(self::VERSION_OPTION, (string) time(), false);
    }

    public static function maybe_bump_on_save($post_id, $post)
    {
        if (!($post instanceof WP_Post)) {
            return;
        }

        if ($post->post_type !== 'post') {
            return;
        }

        if (wp_is_post_revision($post_id)) {
            return;
        }

        self::bump_version();
    }

    public static function key($base)
    {
        return 'tz_blog_i18n_' . self::get_version() . '_' . $base;
    }

    public static function get($base)
    {
        return get_transient(self::key($base));
    }

    public static function set($base, $value, $ttl = null)
    {
        $ttl = is_int($ttl) ? $ttl : (defined('TZ_BLOG_I18N_CACHE_TTL') ? TZ_BLOG_I18N_CACHE_TTL : 900);
        set_transient(self::key($base), $value, $ttl);
    }
}
