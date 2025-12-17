<?php

defined('ABSPATH') || exit;

class TZ_BLOG_Languages
{
    const TAXONOMY = 'tz_lang';

    public function __construct()
    {
        add_action('init', array(__CLASS__, 'register_taxonomy'));
        add_action('init', array(__CLASS__, 'maybe_ensure_terms'));
    }

    public static function get_locales()
    {
        return array(
            'en' => 'English',
            'fr' => 'Français',
            'de' => 'Deutsch',
            'es' => 'Español',
            'ja' => '日本語',
            'ko' => '한국어',
            'it' => 'Italiano',
            'pt' => 'Português',
            'ru' => 'Русский',
            'ar' => 'العربية',
            'fi' => 'Suomi',
            'da' => 'Dansk',
            'th' => 'ไทย',
            'sv' => 'Svenska',
            'id' => 'Bahasa Indonesia',
            'ms' => 'Bahasa Melayu',
            'be' => 'Беларуская',
            'tr' => 'Türkçe',
            'bn' => 'বাংলা',
            'fa' => 'فارسی',
            'nl' => 'Nederlands',
            'hi' => 'हिन्दी',
            'ur' => 'اردو',
            'mr' => 'मराठी',
            'pcm' => 'Nigerian Pidgin',
            'fil' => 'Filipino',
            'te' => 'తెలుగు',
            'ha' => 'Hausa',
            'ps' => 'پښتو',
            'sw' => 'Kiswahili',
            'tl' => 'Tagalog',
            'ta' => 'தமிழ்',
            'jv' => 'Basa Jawa',
            'zh_cn' => '简体中文'
        );
    }

    public static function get_locale_codes()
    {
        return array_keys(self::get_locales());
    }

    public static function is_valid_locale($code)
    {
        return in_array($code, self::get_locale_codes(), true);
    }

    public static function register_taxonomy()
    {
        register_taxonomy(
            self::TAXONOMY,
            array('post'),
            array(
                'public' => false,
                'show_ui' => true,
                'show_admin_column' => true,
                'show_in_rest' => false,
                'hierarchical' => false,
                'labels' => array(
                    'name' => 'Languages',
                    'singular_name' => 'Language'
                )
            )
        );
    }

    public static function maybe_ensure_terms()
    {
        $ready = get_option('tz_blog_i18n_lang_terms_ready');
        if ($ready === '1') {
            return;
        }

        self::ensure_terms();
        update_option('tz_blog_i18n_lang_terms_ready', '1', false);
    }

    public static function ensure_terms()
    {
        foreach (self::get_locales() as $code => $name) {
            $existing = term_exists($code, self::TAXONOMY);
            if (is_array($existing) && isset($existing['term_id'])) {
                continue;
            }

            wp_insert_term(
                $name,
                self::TAXONOMY,
                array(
                    'slug' => $code
                )
            );
        }
    }
}
