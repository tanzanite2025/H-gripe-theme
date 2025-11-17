<?php
/**
 * FAQ Database Operations
 */

defined('ABSPATH') || exit;

class TZ_FAQ_Database
{
    public function __construct()
    {
        // Constructor
    }
    
    /**
     * Get all FAQs with i18n data
     */
    public function get_all_faqs($locale = '')
    {
        global $wpdb;
        
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        $table_i18n = $wpdb->prefix . 'mytheme_faq_i18n';
        
        if ($locale) {
            $sql = $wpdb->prepare("
                SELECT 
                    f.id,
                    f.category,
                    f.order_num,
                    f.created_at,
                    f.updated_at,
                    fi.locale,
                    fi.question,
                    fi.answer
                FROM {$table_faq} f
                LEFT JOIN {$table_i18n} fi ON f.id = fi.faq_id AND fi.locale = %s
                ORDER BY f.category, f.order_num
            ", $locale);
        } else {
            $sql = "
                SELECT 
                    f.id,
                    f.category,
                    f.order_num,
                    f.created_at,
                    f.updated_at
                FROM {$table_faq} f
                ORDER BY f.category, f.order_num
            ";
        }
        
        return $wpdb->get_results($sql, ARRAY_A);
    }
    
    /**
     * Get single FAQ by ID
     */
    public function get_faq($faq_id)
    {
        global $wpdb;
        
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        $table_i18n = $wpdb->prefix . 'mytheme_faq_i18n';
        
        $faq = $wpdb->get_row($wpdb->prepare("
            SELECT * FROM {$table_faq} WHERE id = %d
        ", $faq_id), ARRAY_A);
        
        if (!$faq) {
            return null;
        }
        
        // Get all translations
        $translations = $wpdb->get_results($wpdb->prepare("
            SELECT locale, question, answer 
            FROM {$table_i18n} 
            WHERE faq_id = %d
        ", $faq_id), ARRAY_A);
        
        $faq['translations'] = [];
        foreach ($translations as $trans) {
            $faq['translations'][$trans['locale']] = [
                'question' => $trans['question'],
                'answer' => $trans['answer']
            ];
        }
        
        return $faq;
    }
    
    /**
     * Create new FAQ
     */
    public function create_faq($category, $translations)
    {
        global $wpdb;
        
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        $table_i18n = $wpdb->prefix . 'mytheme_faq_i18n';
        
        // Insert main record
        $wpdb->insert(
            $table_faq,
            [
                'category' => sanitize_text_field($category),
                'order_num' => 0
            ],
            ['%s', '%d']
        );
        
        $faq_id = $wpdb->insert_id;
        
        // Insert translations
        foreach ($translations as $locale => $data) {
            if (empty($data['question'])) {
                continue;
            }
            
            $wpdb->insert(
                $table_i18n,
                [
                    'faq_id' => $faq_id,
                    'locale' => sanitize_key($locale),
                    'question' => sanitize_text_field($data['question']),
                    'answer' => wp_kses_post($data['answer'])
                ],
                ['%d', '%s', '%s', '%s']
            );
        }
        
        return $faq_id;
    }
    
    /**
     * Update FAQ
     */
    public function update_faq($faq_id, $category, $translations)
    {
        global $wpdb;
        
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        $table_i18n = $wpdb->prefix . 'mytheme_faq_i18n';
        
        // Update main record
        $wpdb->update(
            $table_faq,
            ['category' => sanitize_text_field($category)],
            ['id' => $faq_id],
            ['%s'],
            ['%d']
        );
        
        // Delete existing translations
        $wpdb->delete($table_i18n, ['faq_id' => $faq_id], ['%d']);
        
        // Insert new translations
        foreach ($translations as $locale => $data) {
            if (empty($data['question'])) {
                continue;
            }
            
            $wpdb->insert(
                $table_i18n,
                [
                    'faq_id' => $faq_id,
                    'locale' => sanitize_key($locale),
                    'question' => sanitize_text_field($data['question']),
                    'answer' => wp_kses_post($data['answer'])
                ],
                ['%d', '%s', '%s', '%s']
            );
        }
        
        return true;
    }
    
    /**
     * Delete FAQ
     */
    public function delete_faq($faq_id)
    {
        global $wpdb;
        
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        $table_i18n = $wpdb->prefix . 'mytheme_faq_i18n';
        
        // Delete translations first
        $wpdb->delete($table_i18n, ['faq_id' => $faq_id], ['%d']);
        
        // Delete main record
        $wpdb->delete($table_faq, ['id' => $faq_id], ['%d']);
        
        return true;
    }
    
    /**
     * Update FAQ order
     */
    public function update_order($faq_id, $order_num)
    {
        global $wpdb;
        
        $table_faq = $wpdb->prefix . 'mytheme_faq';
        
        $wpdb->update(
            $table_faq,
            ['order_num' => intval($order_num)],
            ['id' => $faq_id],
            ['%d'],
            ['%d']
        );
        
        return true;
    }
}
