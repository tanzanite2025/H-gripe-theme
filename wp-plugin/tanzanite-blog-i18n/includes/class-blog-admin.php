<?php

defined('ABSPATH') || exit;

class TZ_BLOG_Admin
{
    const META_GROUP = 'tz_translation_group';

    public function __construct()
    {
        add_action('add_meta_boxes', array($this, 'register_metabox'));
        add_action('admin_post_tz_blog_create_translation', array($this, 'handle_create_translation'));
        add_action('admin_post_tz_blog_link_translation', array($this, 'handle_link_translation'));
    }

    public function register_metabox()
    {
        add_meta_box(
            'tz_blog_i18n_metabox',
            'Blog Translations',
            array($this, 'render_metabox'),
            'post',
            'side',
            'high'
        );
    }

    private function get_group_id($post_id)
    {
        $group = (string) get_post_meta($post_id, self::META_GROUP, true);
        return $group;
    }

    private function ensure_group_id($post_id)
    {
        $group = $this->get_group_id($post_id);
        if (!empty($group)) {
            return $group;
        }

        $group = wp_generate_uuid4();
        update_post_meta($post_id, self::META_GROUP, $group);
        return $group;
    }

    private function get_post_lang($post_id)
    {
        $slugs = wp_get_post_terms($post_id, TZ_BLOG_Languages::TAXONOMY, array('fields' => 'slugs'));
        if (is_wp_error($slugs) || empty($slugs)) {
            return '';
        }

        return (string) $slugs[0];
    }

    private function get_group_translations($group)
    {
        if (empty($group)) {
            return array();
        }

        $posts = get_posts(
            array(
                'post_type' => 'post',
                'post_status' => array('publish', 'draft', 'pending', 'private'),
                'numberposts' => -1,
                'meta_key' => self::META_GROUP,
                'meta_value' => $group,
                'orderby' => 'date',
                'order' => 'DESC'
            )
        );

        $map = array();
        foreach ($posts as $p) {
            if (!($p instanceof WP_Post)) {
                continue;
            }

            $lang = $this->get_post_lang($p->ID);
            if (empty($lang)) {
                continue;
            }

            $map[$lang] = $p;
        }

        return $map;
    }

    public function render_metabox($post)
    {
        if (!($post instanceof WP_Post)) {
            return;
        }

        $post_id = (int) $post->ID;
        $group = $this->get_group_id($post_id);
        $lang = $this->get_post_lang($post_id);
        $translations = $this->get_group_translations($group);

        wp_nonce_field('tz_blog_i18n_metabox', 'tz_blog_i18n_nonce');

        echo '<p><strong>Language:</strong> ' . esc_html($lang ? $lang : '-') . '</p>';
        echo '<p><strong>Group:</strong> ' . esc_html($group ? $group : '-') . '</p>';

        echo '<p><strong>Translations:</strong></p>';
        if (empty($translations)) {
            echo '<p>-</p>';
        } else {
            echo '<ul>';
            foreach ($translations as $t_lang => $t_post) {
                $edit_link = get_edit_post_link($t_post->ID);
                echo '<li>' . esc_html($t_lang) . ' → <a href="' . esc_url($edit_link) . '">#' . esc_html($t_post->ID) . '</a></li>';
            }
            echo '</ul>';
        }

        $codes = TZ_BLOG_Languages::get_locale_codes();

        echo '<hr />';
        echo '<p><strong>Create translation</strong></p>';
        echo '<form method="post" action="' . esc_url(admin_url('admin-post.php')) . '">';
        echo '<input type="hidden" name="action" value="tz_blog_create_translation" />';
        echo '<input type="hidden" name="post_id" value="' . esc_attr($post_id) . '" />';
        echo '<input type="hidden" name="nonce" value="' . esc_attr(wp_create_nonce('tz_blog_create_translation')) . '" />';

        echo '<select name="target_lang" style="width:100%">';
        foreach ($codes as $code) {
            $disabled = isset($translations[$code]) ? ' disabled' : '';
            $selected = $code === $lang ? ' selected' : '';
            echo '<option value="' . esc_attr($code) . '"' . $selected . $disabled . '>' . esc_html($code) . '</option>';
        }
        echo '</select>';

        echo '<p><button type="submit" class="button">Create</button></p>';
        echo '</form>';

        echo '<hr />';
        echo '<p><strong>Link existing translation</strong></p>';
        echo '<form method="post" action="' . esc_url(admin_url('admin-post.php')) . '">';
        echo '<input type="hidden" name="action" value="tz_blog_link_translation" />';
        echo '<input type="hidden" name="post_id" value="' . esc_attr($post_id) . '" />';
        echo '<input type="hidden" name="nonce" value="' . esc_attr(wp_create_nonce('tz_blog_link_translation')) . '" />';

        echo '<p><input type="number" name="translation_post_id" placeholder="Post ID" style="width:100%" /></p>';
        echo '<select name="translation_lang" style="width:100%">';
        foreach ($codes as $code) {
            $disabled = isset($translations[$code]) ? ' disabled' : '';
            echo '<option value="' . esc_attr($code) . '"' . $disabled . '>' . esc_html($code) . '</option>';
        }
        echo '</select>';

        echo '<p><button type="submit" class="button">Link</button></p>';
        echo '</form>';
    }

    public function handle_create_translation()
    {
        $nonce = isset($_POST['nonce']) ? (string) $_POST['nonce'] : '';
        if (!wp_verify_nonce($nonce, 'tz_blog_create_translation')) {
            wp_die('Invalid nonce');
        }

        $post_id = isset($_POST['post_id']) ? (int) $_POST['post_id'] : 0;
        $target_lang = isset($_POST['target_lang']) ? sanitize_key((string) $_POST['target_lang']) : '';

        if ($post_id <= 0 || !TZ_BLOG_Languages::is_valid_locale($target_lang)) {
            wp_die('Invalid request');
        }

        if (!current_user_can('edit_post', $post_id)) {
            wp_die('Forbidden');
        }

        $source = get_post($post_id);
        if (!($source instanceof WP_Post) || $source->post_type !== 'post') {
            wp_die('Invalid source');
        }

        $source_lang = $this->get_post_lang($post_id);
        if (!empty($source_lang) && $source_lang === $target_lang) {
            wp_die('Target language must be different');
        }

        $group = $this->ensure_group_id($post_id);
        $translations = $this->get_group_translations($group);
        if (isset($translations[$target_lang])) {
            wp_safe_redirect(get_edit_post_link($translations[$target_lang]->ID, '')); 
            exit;
        }

        $new_post_id = wp_insert_post(
            array(
                'post_type' => 'post',
                'post_status' => 'draft',
                'post_title' => $source->post_title,
                'post_content' => $source->post_content,
                'post_excerpt' => $source->post_excerpt
            )
        );

        if (is_wp_error($new_post_id) || empty($new_post_id)) {
            wp_die('Failed to create translation');
        }

        update_post_meta($new_post_id, self::META_GROUP, $group);

        wp_set_post_terms($new_post_id, array($target_lang), TZ_BLOG_Languages::TAXONOMY, false);

        $cats = wp_get_post_terms($post_id, 'category', array('fields' => 'ids'));
        if (!is_wp_error($cats) && !empty($cats)) {
            wp_set_post_terms($new_post_id, $cats, 'category', false);
        }

        $tags = wp_get_post_terms($post_id, 'post_tag', array('fields' => 'names'));
        if (!is_wp_error($tags) && !empty($tags)) {
            wp_set_post_terms($new_post_id, $tags, 'post_tag', false);
        }

        $thumb_id = get_post_thumbnail_id($post_id);
        if (!empty($thumb_id)) {
            set_post_thumbnail($new_post_id, $thumb_id);
        }

        wp_safe_redirect(get_edit_post_link($new_post_id, ''));
        exit;
    }

    public function handle_link_translation()
    {
        $nonce = isset($_POST['nonce']) ? (string) $_POST['nonce'] : '';
        if (!wp_verify_nonce($nonce, 'tz_blog_link_translation')) {
            wp_die('Invalid nonce');
        }

        $post_id = isset($_POST['post_id']) ? (int) $_POST['post_id'] : 0;
        $translation_post_id = isset($_POST['translation_post_id']) ? (int) $_POST['translation_post_id'] : 0;
        $translation_lang = isset($_POST['translation_lang']) ? sanitize_key((string) $_POST['translation_lang']) : '';

        if ($post_id <= 0 || $translation_post_id <= 0 || !TZ_BLOG_Languages::is_valid_locale($translation_lang)) {
            wp_die('Invalid request');
        }

        if ($post_id === $translation_post_id) {
            wp_die('Cannot link post to itself');
        }

        if (!current_user_can('edit_post', $post_id) || !current_user_can('edit_post', $translation_post_id)) {
            wp_die('Forbidden');
        }

        $source = get_post($post_id);
        $target = get_post($translation_post_id);

        if (!($source instanceof WP_Post) || !($target instanceof WP_Post)) {
            wp_die('Invalid posts');
        }

        if ($source->post_type !== 'post' || $target->post_type !== 'post') {
            wp_die('Invalid post type');
        }

        $group = $this->ensure_group_id($post_id);
        $translations = $this->get_group_translations($group);
        if (isset($translations[$translation_lang]) && (int) $translations[$translation_lang]->ID !== (int) $target->ID) {
            wp_die('Language already linked in this group');
        }

        update_post_meta($target->ID, self::META_GROUP, $group);
        wp_set_post_terms($target->ID, array($translation_lang), TZ_BLOG_Languages::TAXONOMY, false);

        $cats = wp_get_post_terms($post_id, 'category', array('fields' => 'ids'));
        if (!is_wp_error($cats) && !empty($cats)) {
            wp_set_post_terms($target->ID, $cats, 'category', false);
        }

        $tags = wp_get_post_terms($post_id, 'post_tag', array('fields' => 'names'));
        if (!is_wp_error($tags) && !empty($tags)) {
            wp_set_post_terms($target->ID, $tags, 'post_tag', false);
        }

        wp_safe_redirect(get_edit_post_link($post_id, ''));
        exit;
    }
}
