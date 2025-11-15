<?php
if (!defined('ABSPATH')) { exit; }

// 不在文章编辑页显示任何 UI，所有设置集中在 URLLink 后台菜单页。

function urllink_normalize_path($raw) {
    $raw = trim((string)$raw);
    $raw = preg_replace('#^https?://[^/]+/#i', '', $raw);
    $raw = trim($raw, "/\t\n\r ");
    return $raw;
}
