<?php
/**
 * FAQ Editor - Admin Interface
 */

defined('ABSPATH') || exit;

class TZ_FAQ_Editor
{
    public function __construct()
    {
        add_action('admin_menu', [$this, 'add_admin_menu']);
        add_action('admin_post_save_faq', [$this, 'handle_save']);
        add_action('admin_post_delete_faq', [$this, 'handle_delete']);
        add_action('admin_post_generate_json', [$this, 'handle_generate_json']);
        add_action('admin_enqueue_scripts', [$this, 'enqueue_assets']);
    }
    
    /**
     * Get database instance
     */
    private function get_db()
    {
        return new TZ_FAQ_Database();
    }
    
    /**
     * Get generator instance
     */
    private function get_generator()
    {
        return new TZ_FAQ_Json_Generator();
    }
    
    /**
     * Add admin menu
     */
    public function add_admin_menu()
    {
        add_menu_page(
            'FAQ Content',
            'FAQ Content',
            'manage_options',
            'tanzanite-faq-content',
            [$this, 'render_list_page'],
            'dashicons-editor-help',
            30
        );
        
        add_submenu_page(
            'tanzanite-faq-content',
            'Add New FAQ',
            'Add New',
            'manage_options',
            'tanzanite-faq-add',
            [$this, 'render_edit_page']
        );
    }
    
    /**
     * Enqueue admin assets
     */
    public function enqueue_assets($hook)
    {
        if (strpos($hook, 'tanzanite-faq') === false) {
            return;
        }
        
        wp_enqueue_style('wp-color-picker');
        wp_enqueue_editor();
        wp_enqueue_media();
        
        // Custom CSS - matching Customer Service style
        wp_add_inline_style('wp-admin', '
            .tz-faq-admin .tz-settings-wrapper {
                box-sizing: border-box;
                max-width: 1500px;
                margin: 40px auto;
                padding: 32px;
                background: #fff;
                border-radius: 8px;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
                display: flex;
                flex-direction: column;
                gap: 24px;
                font-size: 14px;
                color: #1f2937;
            }
            
            .tz-faq-admin .tz-settings-header {
                display: flex;
                flex-direction: column;
                gap: 12px;
                align-items: center;
                text-align: center;
            }
            
            .tz-faq-admin .tz-settings-header h1 {
                margin: 0;
                font-size: 14px;
                font-weight: 700;
                letter-spacing: 0.02em;
                text-transform: uppercase;
                color: #111827;
            }
            
            .tz-faq-admin .tz-settings-header p {
                margin: 0;
                max-width: 960px;
                color: #4b5563;
            }
            
            .tz-faq-admin .tz-settings-section {
                background: #f8fafc;
                border-radius: 8px;
                padding: 16px;
                display: flex;
                flex-direction: column;
                gap: 12px;
                border: 1px solid #e5e7eb;
            }
            
            .tz-faq-admin .wp-list-table {
                border: 1px solid #e5e7eb;
                border-radius: 8px;
                background: #fff;
            }
            
            .tz-faq-admin .wp-list-table thead th {
                background: #f3f4f6;
                color: #111827;
                font-weight: 600;
                font-size: 14px;
                border-bottom: 1px solid #e5e7eb;
            }
            
            .tz-faq-admin .wp-list-table tbody tr {
                border-bottom: 1px solid #f3f4f6;
            }
            
            .tz-faq-admin .wp-list-table tbody tr:last-child {
                border-bottom: none;
            }
            
            .tz-faq-admin .button {
                background: #1f2937;
                border-color: #1f2937;
                color: #fff;
                border-radius: 6px;
                padding: 8px 16px;
                font-weight: 500;
                transition: all 0.2s;
            }
            
            .tz-faq-admin .button:hover {
                background: #111827;
                border-color: #111827;
            }
            
            .tz-faq-admin .button-primary {
                background: #1f2937;
                border-color: #1f2937;
            }
            
            .tz-faq-admin .button-primary:hover {
                background: #111827;
                border-color: #111827;
            }
            
            .tz-faq-admin input[type="text"],
            .tz-faq-admin select,
            .tz-faq-admin textarea {
                border: 1px solid #e5e7eb;
                border-radius: 6px;
                padding: 8px 12px;
                font-size: 14px;
            }
            
            .tz-faq-admin input[type="text"]:focus,
            .tz-faq-admin select:focus,
            .tz-faq-admin textarea:focus {
                border-color: #1f2937;
                box-shadow: 0 0 0 1px #1f2937;
            }
            
            .tz-faq-admin .language-tab {
                background: #f8fafc;
                border: 1px solid #e5e7eb;
                border-radius: 8px;
                padding: 20px;
                margin-bottom: 16px;
            }
            
            .tz-faq-admin .language-tab h3 {
                margin-top: 0;
                font-size: 16px;
                font-weight: 600;
                color: #111827;
                padding-bottom: 12px;
                border-bottom: 2px solid #1f2937;
            }
            
            .tz-faq-admin .faq-form-field {
                margin-bottom: 20px;
            }
            
            .tz-faq-admin .faq-form-field label {
                display: block;
                font-weight: 600;
                margin-bottom: 8px;
                color: #111827;
            }
            
            @media (max-width: 1600px) {
                .tz-faq-admin .tz-settings-wrapper {
                    width: 90%;
                }
            }
            
            @media (max-width: 1024px) {
                .tz-faq-admin .tz-settings-wrapper {
                    width: calc(100% - 48px);
                    margin: 24px auto;
                    padding: 24px;
                }
            }
        ');
    }
    
    /**
     * Get languages from Nuxt i18n locales directory
     */
    private function get_languages()
    {
        // Method 1: Scan Nuxt i18n locales directory (most reliable)
        $possible_locale_dirs = [
            ABSPATH . '../tanzanite-theme/nuxt-i18n/i18n/locales',
            ABSPATH . '../../tanzanite-theme/nuxt-i18n/i18n/locales',
            dirname(ABSPATH) . '/tanzanite-theme/nuxt-i18n/i18n/locales',
        ];
        
        foreach ($possible_locale_dirs as $locales_dir) {
            if (is_dir($locales_dir)) {
                $files = glob($locales_dir . '/*.json');
                if (!empty($files)) {
                    $codes = [];
                    foreach ($files as $file) {
                        $filename = basename($file, '.json');
                        // Keep original filename as language code (zh_cn, not zh)
                        $codes[] = $filename;
                    }
                    
                    if (defined('WP_DEBUG') && WP_DEBUG) {
                        error_log('TZ_FAQ: Found ' . count($codes) . ' languages from Nuxt locales directory');
                    }
                    
                    return $codes;
                }
            }
        }
        
        // Method 2: Try reading i18n-languages.json from public_html
        $possible_json_paths = [
            ABSPATH . 'i18n-languages.json',
            ABSPATH . '../i18n-languages.json',
        ];
        
        foreach ($possible_json_paths as $json_file) {
            if (file_exists($json_file)) {
                $json_content = file_get_contents($json_file);
                $languages_data = json_decode($json_content, true);
                
                if (is_array($languages_data)) {
                    $codes = array_column($languages_data, 'code');
                    if (!in_array('zh', $codes)) {
                        array_unshift($codes, 'zh');
                    }
                    
                    if (defined('WP_DEBUG') && WP_DEBUG) {
                        error_log('TZ_FAQ: Found ' . count($codes) . ' languages from i18n-languages.json');
                    }
                    
                    return $codes;
                }
            }
        }
        
        // Method 3: Try SEO plugin option
        $languages = get_option('mytheme_seo_languages', []);
        if (!empty($languages) && is_array($languages)) {
            if (defined('WP_DEBUG') && WP_DEBUG) {
                error_log('TZ_FAQ: Using SEO plugin languages: ' . count($languages));
            }
            return $languages;
        }
        
        // No languages found - log error and return empty array
        error_log('TZ_FAQ ERROR: Cannot find language configuration! Please check:');
        error_log('  1. Nuxt i18n locales directory exists at: tanzanite-theme/nuxt-i18n/i18n/locales/');
        error_log('  2. i18n-languages.json exists in public_html');
        error_log('  3. mytheme_seo_languages option is set in WordPress');
        
        return [];
    }
    
    /**
     * Render FAQ list page
     */
    public function render_list_page()
    {
        $db = $this->get_db();
        $faqs = $db->get_all_faqs();
        
        ?>
        <div class="wrap tz-faq-admin">
            <div class="tz-settings-wrapper">
                <div class="tz-settings-header">
                    <h1>FAQ Content Manager</h1>
                    <p>Manage FAQ content with 34 languages support. Content will be automatically exported to JSON files for the frontend.</p>
                </div>
                
                <div style="display: flex; gap: 12px; justify-content: center;">
                    <a href="<?php echo admin_url('admin.php?page=tanzanite-faq-add'); ?>" class="button button-primary">➕ Add New FAQ</a>
                    <form method="post" action="<?php echo admin_url('admin-post.php'); ?>" style="display: inline-block;">
                        <input type="hidden" name="action" value="generate_json">
                        <?php wp_nonce_field('generate_json_nonce'); ?>
                        <button type="submit" class="button">
                            🔄 Regenerate All JSON Files
                        </button>
                    </form>
                </div>
            
            <?php if (isset($_GET['saved'])): ?>
                <div class="notice notice-success is-dismissible">
                    <p>FAQ saved successfully and JSON files generated!</p>
                </div>
            <?php endif; ?>
            
            <?php if (isset($_GET['deleted'])): ?>
                <div class="notice notice-success is-dismissible">
                    <p>FAQ deleted successfully!</p>
                </div>
            <?php endif; ?>
            
            <?php if (isset($_GET['generated'])): ?>
                <div class="notice notice-success is-dismissible">
                    <p>JSON files regenerated successfully!</p>
                </div>
            <?php endif; ?>
            
            <table class="wp-list-table widefat fixed striped faq-list-table">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Category</th>
                        <th>Question (First Language)</th>
                        <th>Order</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <?php if (empty($faqs)): ?>
                        <tr>
                            <td colspan="5" style="text-align: center; padding: 40px;">
                                No FAQs found. <a href="<?php echo admin_url('admin.php?page=tanzanite-faq-add'); ?>">Add your first FAQ</a>
                            </td>
                        </tr>
                    <?php else: ?>
                        <?php foreach ($faqs as $faq): ?>
                            <?php
                            $faq_data = $db->get_faq($faq['id']);
                            $first_question = '';
                            if (!empty($faq_data['translations'])) {
                                $first_trans = reset($faq_data['translations']);
                                $first_question = $first_trans['question'];
                            }
                            ?>
                            <tr>
                                <td><?php echo esc_html($faq['id']); ?></td>
                                <td><?php echo esc_html(ucfirst($faq['category'])); ?></td>
                                <td><?php echo esc_html($first_question); ?></td>
                                <td><?php echo esc_html($faq['order_num']); ?></td>
                                <td class="faq-actions">
                                    <a href="<?php echo admin_url('admin.php?page=tanzanite-faq-add&id=' . $faq['id']); ?>" class="button button-small">Edit</a>
                                    <form method="post" action="<?php echo admin_url('admin-post.php'); ?>" style="display: inline;">
                                        <input type="hidden" name="action" value="delete_faq">
                                        <input type="hidden" name="faq_id" value="<?php echo esc_attr($faq['id']); ?>">
                                        <?php wp_nonce_field('delete_faq_nonce'); ?>
                                        <button type="submit" class="button button-small button-link-delete" onclick="return confirm('Are you sure you want to delete this FAQ?');">Delete</button>
                                    </form>
                                </td>
                            </tr>
                        <?php endforeach; ?>
                    <?php endif; ?>
                </tbody>
            </table>
            </div>
        </div>
        <?php
    }
    
    /**
     * Render edit/add page
     */
    public function render_edit_page()
    {
        try {
            $faq_id = isset($_GET['id']) ? intval($_GET['id']) : 0;
            $db = $this->get_db();
            $faq = $faq_id ? $db->get_faq($faq_id) : null;
            $languages = $this->get_languages();
            
            if (defined('WP_DEBUG') && WP_DEBUG) {
                error_log('TZ_FAQ: render_edit_page - FAQ ID: ' . $faq_id);
                error_log('TZ_FAQ: render_edit_page - Languages count: ' . count($languages));
            }
        } catch (Exception $e) {
            if (defined('WP_DEBUG') && WP_DEBUG) {
                error_log('TZ_FAQ: render_edit_page error: ' . $e->getMessage());
            }
            wp_die('Error loading FAQ editor: ' . $e->getMessage());
        }
        
        ?>
        <div class="wrap tz-faq-admin">
            <div class="tz-settings-wrapper">
                <div class="tz-settings-header">
                    <h1><?php echo $faq_id ? 'Edit FAQ' : 'Add New FAQ'; ?></h1>
                    <p>Fill in the FAQ content for all supported languages. At least one language is required.</p>
                </div>
            
            <form method="post" action="<?php echo admin_url('admin-post.php'); ?>">
                <input type="hidden" name="action" value="save_faq">
                <?php if ($faq_id): ?>
                    <input type="hidden" name="faq_id" value="<?php echo esc_attr($faq_id); ?>">
                <?php endif; ?>
                <?php wp_nonce_field('save_faq_nonce'); ?>
                
                <div class="faq-form-field">
                    <label for="category">Category:</label>
                    <select name="category" id="category" class="regular-text" required>
                        <option value="product" <?php selected($faq['category'] ?? '', 'product'); ?>>Product Questions</option>
                        <option value="shipping" <?php selected($faq['category'] ?? '', 'shipping'); ?>>Shipping & Delivery</option>
                        <option value="return" <?php selected($faq['category'] ?? '', 'return'); ?>>Returns & Refunds</option>
                        <option value="payment" <?php selected($faq['category'] ?? '', 'payment'); ?>>Payment Methods</option>
                    </select>
                </div>
                
                <div class="faq-languages">
                    <h2>Multilingual Content (<?php echo count($languages); ?> languages)</h2>
                    <p class="description">Fill in at least one language. Empty languages will be skipped.</p>
                    
                    <?php foreach ($languages as $lang): ?>
                        <?php
                        $question = isset($faq['translations'][$lang]['question']) ? $faq['translations'][$lang]['question'] : '';
                        $answer = isset($faq['translations'][$lang]['answer']) ? $faq['translations'][$lang]['answer'] : '';
                        ?>
                        <div class="language-tab">
                            <h3><?php echo strtoupper($lang); ?> - <?php echo $this->get_language_name($lang); ?></h3>
                            
                            <div class="faq-form-field">
                                <label for="question_<?php echo $lang; ?>">Question (<?php echo $lang; ?>):</label>
                                <input type="text" 
                                       name="question[<?php echo $lang; ?>]" 
                                       id="question_<?php echo $lang; ?>"
                                       value="<?php echo esc_attr($question); ?>"
                                       class="widefat">
                            </div>
                            
                            <div class="faq-form-field">
                                <label for="answer_<?php echo $lang; ?>">Answer (<?php echo $lang; ?>):</label>
                                <?php 
                                wp_editor($answer, 'answer_' . $lang, [
                                    'textarea_name' => "answer[{$lang}]",
                                    'textarea_rows' => 8,
                                    'media_buttons' => false,
                                    'teeny' => true,
                                ]); 
                                ?>
                            </div>
                        </div>
                    <?php endforeach; ?>
                </div>
                
                <p class="submit" style="text-align: center;">
                    <button type="submit" class="button button-primary button-large">
                        <?php echo $faq_id ? '✅ Update FAQ' : '➕ Create FAQ'; ?> & Generate JSON
                    </button>
                    <a href="<?php echo admin_url('admin.php?page=tanzanite-faq-content'); ?>" class="button button-large">Cancel</a>
                </p>
            </form>
            </div>
        </div>
        <?php
    }
    
    /**
     * Get language name
     */
    private function get_language_name($code)
    {
        $names = [
            'en' => 'English', 'zh' => '中文', 'ja' => '日本語', 'ko' => '한국어',
            'es' => 'Español', 'fr' => 'Français', 'de' => 'Deutsch', 'it' => 'Italiano',
            'pt' => 'Português', 'ru' => 'Русский', 'ar' => 'العربية', 'hi' => 'हिन्दी',
        ];
        return $names[$code] ?? $code;
    }
    
    /**
     * Handle save
     */
    public function handle_save()
    {
        check_admin_referer('save_faq_nonce');
        
        if (!current_user_can('manage_options')) {
            wp_die('Unauthorized');
        }
        
        $faq_id = isset($_POST['faq_id']) ? intval($_POST['faq_id']) : 0;
        $category = sanitize_text_field($_POST['category']);
        $translations = [];
        
        foreach ($_POST['question'] as $lang => $question) {
            $answer = $_POST['answer'][$lang] ?? '';
            if (!empty($question)) {
                $translations[$lang] = [
                    'question' => $question,
                    'answer' => $answer
                ];
            }
        }
        
        $db = $this->get_db();
        
        if ($faq_id) {
            $db->update_faq($faq_id, $category, $translations);
        } else {
            $faq_id = $db->create_faq($category, $translations);
        }
        
        // Trigger JSON generation
        do_action('tanzanite_faq_saved', $faq_id);
        
        wp_redirect(admin_url('admin.php?page=tanzanite-faq-content&saved=1'));
        exit;
    }
    
    /**
     * Handle delete
     */
    public function handle_delete()
    {
        check_admin_referer('delete_faq_nonce');
        
        if (!current_user_can('manage_options')) {
            wp_die('Unauthorized');
        }
        
        $faq_id = intval($_POST['faq_id']);
        $db = $this->get_db();
        $db->delete_faq($faq_id);
        
        // Trigger JSON generation
        do_action('tanzanite_faq_deleted', $faq_id);
        
        wp_redirect(admin_url('admin.php?page=tanzanite-faq-content&deleted=1'));
        exit;
    }
    
    /**
     * Handle manual JSON generation
     */
    public function handle_generate_json()
    {
        check_admin_referer('generate_json_nonce');
        
        if (!current_user_can('manage_options')) {
            wp_die('Unauthorized');
        }
        
        $generator = $this->get_generator();
        $generator->generate_all_json();
        
        wp_redirect(admin_url('admin.php?page=tanzanite-faq-content&generated=1'));
        exit;
    }
}
