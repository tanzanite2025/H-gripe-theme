<?php
/**
 * 自动重构评价模块 JS
 * 
 * 使用方法：
 * php refactor-reviews-js.php
 */

$file = __DIR__ . '/tanzanite-setting.php';

if (!file_exists($file)) {
    die("错误：找不到文件 tanzanite-setting.php\n");
}

echo "开始重构评价模块 JS...\n\n";

// 备份原文件
$backup = $file . '.backup.' . date('YmdHis');
copy($file, $backup);
echo "✓ 已创建备份：" . basename($backup) . "\n\n";

// 读取文件
$content = file_get_contents($file);
$original_lines = substr_count($content, "\n");

echo "✓ 文件总行数：{$original_lines}\n\n";

// ========================================
// 步骤 1: 找到并删除内联 JS
// ========================================
echo "步骤 1: 删除内联 JS\n";
echo "----------------------------------------\n";

// 查找内联 JS 的开始和结束
$pattern = '/\$config_js = wp_json_encode\(.*?\);.*?echo \'<script>\' \. sprintf\( \$script, \$config_js, \$strings_js \) \. \'<\/script>\';/s';

if (preg_match($pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
    $start_pos = $matches[0][1];
    $length = strlen($matches[0][0]);
    
    // 计算删除的行数
    $deleted_text = $matches[0][0];
    $deleted_lines = substr_count($deleted_text, "\n");
    
    // 删除内联 JS
    $content = substr_replace($content, '', $start_pos, $length);
    
    echo "✓ 已删除内联 JS\n";
    echo "  起始位置：{$start_pos}\n";
    echo "  删除字符数：{$length}\n";
    echo "  删除行数：{$deleted_lines}\n\n";
} else {
    echo "⚠ 未找到内联 JS，可能已经删除\n\n";
}

// ========================================
// 步骤 2: 添加 JS 加载代码
// ========================================
echo "步骤 2: 添加 JS 加载代码\n";
echo "----------------------------------------\n";

// 查找插入位置（在 echo '</div>'; 之后，render_payment_method 之前）
$search = "            echo '</div>';\n        }\n\n        public function render_payment_method";
$replacement = <<<'PHP'
            echo '</div>';

            // 加载评价管理 JS
            wp_enqueue_script(
                'tz-reviews',
                plugins_url( 'assets/js/reviews.js', __FILE__ ),
                array( 'tz-admin-common' ),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-reviews',
                'TzReviewsConfig',
                array(
                    'listUrl'    => $rest_list,
                    'singleUrl'  => $rest_single,
                    'nonce'      => $nonce,
                    'statuses'   => array_values( self::ALLOWED_REVIEW_STATUSES ),
                    'canManage'  => $can_manage,
                    'i18n'       => array(
                        'noPermission'        => __( '当前账号仅具备查看权限，审核操作已禁用。', 'tanzanite-settings' ),
                        'noPermissionHint'    => __( '如需执行审核或回复，请联系管理员授予"评价管理"权限。', 'tanzanite-settings' ),
                        'loadFailed'          => __( '加载评价列表失败。', 'tanzanite-settings' ),
                        'saveSuccess'         => __( '评价已更新。', 'tanzanite-settings' ),
                        'deleteConfirm'       => __( '确定删除该评价？此操作不可撤销。', 'tanzanite-settings' ),
                        'deleteSuccess'       => __( '评价已删除。', 'tanzanite-settings' ),
                        'selectReview'        => __( '请先选择要操作的评价。', 'tanzanite-settings' ),
                        'contentPlaceholder'  => __( '暂无内容', 'tanzanite-settings' ),
                        'view'                => __( '查看', 'tanzanite-settings' ),
                        'approve'             => __( '通过', 'tanzanite-settings' ),
                        'reject'              => __( '拒绝', 'tanzanite-settings' ),
                        'hide'                => __( '隐藏', 'tanzanite-settings' ),
                        'markFeatured'        => __( '标记精华', 'tanzanite-settings' ),
                        'unmarkFeatured'      => __( '取消精华', 'tanzanite-settings' ),
                        'yes'                 => __( '是', 'tanzanite-settings' ),
                        'no'                  => __( '否', 'tanzanite-settings' ),
                        'itemsLabel'          => __( '条评价', 'tanzanite-settings' ),
                    ),
                )
            );
        }

        public function render_payment_method
PHP;

if (strpos($content, $search) !== false) {
    $content = str_replace($search, $replacement, $content);
    echo "✓ 已添加 JS 加载代码\n\n";
} else {
    echo "⚠ 未找到插入位置，请手动添加\n\n";
}

// ========================================
// 保存文件
// ========================================
file_put_contents($file, $content);

$new_lines = substr_count($content, "\n");
$diff_lines = $original_lines - $new_lines;

echo "========================================\n";
echo "重构完成！\n";
echo "========================================\n";
echo "\n";
echo "原文件行数：{$original_lines}\n";
echo "新文件行数：{$new_lines}\n";
echo "删除行数：{$diff_lines}\n";
echo "\n";
echo "备份文件：{$backup}\n";
echo "\n";
echo "下一步：\n";
echo "1. 刷新 WordPress 后台\n";
echo "2. 访问 Products → Reviews\n";
echo "3. 测试评价管理功能\n";
echo "\n";
echo "如果有问题，可以从备份恢复：\n";
echo "cp " . basename($backup) . " tanzanite-setting.php\n";
echo "\n";
?>
