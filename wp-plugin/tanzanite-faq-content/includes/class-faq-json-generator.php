<?php
/**
 * FAQ JSON Generator
 * Generates static JSON files for Nuxt frontend
 */

defined('ABSPATH') || exit;

class TZ_FAQ_Json_Generator
{
    private $upload_dir;
    
    public function __construct()
    {
        $upload = wp_upload_dir();
        $this->upload_dir = $upload['basedir'] . '/faq/';
        
        // Hook to generate JSON after FAQ save/update/delete
        add_action('tanzanite_faq_saved', [$this, 'generate_all_json']);
        add_action('tanzanite_faq_deleted', [$this, 'generate_all_json']);
    }
    
    /**
     * Get database instance
     */
    private function get_db()
    {
        return new TZ_FAQ_Database();
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
                    return $codes;
                }
            }
        }
        
        // Method 3: Try SEO plugin option
        $languages = get_option('mytheme_seo_languages', []);
        if (!empty($languages) && is_array($languages)) {
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
     * Get category name by locale
     */
    private function get_category_name($category, $locale)
    {
        $names = [
            'product' => [
                'en' => 'Product Questions',
                'zh' => '产品问题',
                'ja' => '製品に関する質問',
                'ko' => '제품 질문',
                'es' => 'Preguntas sobre productos',
                'fr' => 'Questions sur les produits',
                'de' => 'Produktfragen',
            ],
            'shipping' => [
                'en' => 'Shipping & Delivery',
                'zh' => '配送与物流',
                'ja' => '配送と配達',
                'ko' => '배송 및 배달',
                'es' => 'Envío y entrega',
                'fr' => 'Expédition et livraison',
                'de' => 'Versand und Lieferung',
            ],
            'return' => [
                'en' => 'Returns & Refunds',
                'zh' => '退货与退款',
                'ja' => '返品と返金',
                'ko' => '반품 및 환불',
                'es' => 'Devoluciones y reembolsos',
                'fr' => 'Retours et remboursements',
                'de' => 'Rücksendungen und Rückerstattungen',
            ],
            'payment' => [
                'en' => 'Payment Methods',
                'zh' => '支付方式',
                'ja' => '支払い方法',
                'ko' => '결제 방법',
                'es' => 'Métodos de pago',
                'fr' => 'Modes de paiement',
                'de' => 'Zahlungsmethoden',
            ],
        ];
        
        return $names[$category][$locale] ?? ucfirst($category);
    }
    
    /**
     * Get category icon
     */
    private function get_category_icon($category)
    {
        $icons = [
            'product' => '📦',
            'shipping' => '🚚',
            'return' => '↩️',
            'payment' => '💳',
        ];
        
        return $icons[$category] ?? '❓';
    }
    
    /**
     * Generate JSON for a single language
     */
    public function generate_json_for_language($locale)
    {
        global $wpdb;
        
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        $table_i18n = $wpdb->prefix . 'mytheme_faq_i18n';
        
        if (defined('WP_DEBUG') && WP_DEBUG) {
            error_log("TZ_FAQ: Generating JSON for locale: {$locale}");
        }
        
        // Get all FAQs for this locale
        $faqs = $wpdb->get_results($wpdb->prepare("
            SELECT 
                f.id,
                f.category,
                f.order_num,
                fi.question,
                fi.answer
            FROM {$table_faq} f
            LEFT JOIN {$table_i18n} fi 
                ON f.id = fi.faq_id AND fi.locale = %s
            WHERE fi.question IS NOT NULL AND fi.question != ''
            ORDER BY f.category, f.order_num
        ", $locale), ARRAY_A);
        
        if (defined('WP_DEBUG') && WP_DEBUG) {
            error_log("TZ_FAQ: Found " . count($faqs) . " FAQs for locale: {$locale}");
        }
        
        // Group by category
        $categories = [];
        foreach ($faqs as $faq) {
            $cat = $faq['category'];
            if (!isset($categories[$cat])) {
                $categories[$cat] = [
                    'id' => $cat,
                    'name' => $this->get_category_name($cat, $locale),
                    'icon' => $this->get_category_icon($cat),
                    'items' => []
                ];
            }
            
            $categories[$cat]['items'][] = [
                'id' => (int) $faq['id'],
                'question' => $faq['question'],
                'answer' => $faq['answer'],
                'order' => (int) $faq['order_num']
            ];
        }
        
        // Build final JSON structure
        $json_data = [
            'categories' => array_values($categories),
            'meta' => [
                'locale' => $locale,
                'last_updated' => current_time('c'),
                'total_items' => count($faqs),
                'version' => TANZANITE_FAQ_VERSION
            ]
        ];
        
        // Write to file
        $file_path = $this->upload_dir . $locale . '.json';
        $json_content = wp_json_encode($json_data, JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT);
        
        // Ensure directory exists
        if (!is_dir($this->upload_dir)) {
            wp_mkdir_p($this->upload_dir);
        }
        
        $result = file_put_contents($file_path, $json_content);
        
        if ($result === false) {
            error_log("TZ_FAQ ERROR: Failed to write JSON file: {$file_path}");
            return false;
        }
        
        chmod($file_path, 0644);
        
        if (defined('WP_DEBUG') && WP_DEBUG) {
            error_log("TZ_FAQ: Successfully generated JSON: {$file_path} (" . strlen($json_content) . " bytes)");
        }
        
        return $file_path;
    }
    
    /**
     * Generate JSON files for all languages
     */
    public function generate_all_json()
    {
        $languages = $this->get_languages();
        $generated = [];
        
        foreach ($languages as $lang) {
            try {
                $file = $this->generate_json_for_language($lang);
                $generated[] = $lang;
            } catch (Exception $e) {
                error_log('FAQ JSON generation failed for ' . $lang . ': ' . $e->getMessage());
            }
        }
        
        return [
            'success' => true,
            'message' => 'Generated ' . count($generated) . ' JSON files',
            'languages' => $generated,
            'timestamp' => current_time('c')
        ];
    }
    
    /**
     * Manual trigger for JSON generation (for admin use)
     */
    public function manual_generate()
    {
        if (!current_user_can('manage_options')) {
            return ['success' => false, 'message' => 'Unauthorized'];
        }
        
        return $this->generate_all_json();
    }
}
