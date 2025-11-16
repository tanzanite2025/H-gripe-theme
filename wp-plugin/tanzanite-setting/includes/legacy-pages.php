<?php
/**
 * Legacy Admin Pages
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

// 定义主插件文件路径
if ( ! defined( 'TANZANITE_LEGACY_MAIN_FILE' ) ) {
    define( 'TANZANITE_LEGACY_MAIN_FILE', dirname( __DIR__ ) . '/tanzanite-setting.php' );
}

if ( ! class_exists( 'Tanzanite_Settings_Plugin' ) ) {

    interface Tanzanite_Tracking_Provider_Interface {
        public function get_name(): string;

        /**
         * @return array|
         */
        public function fetch_tracking_events( string $tracking_number, array $args = [] );

        public function test_connection(): bool;
    }

    class Tanzanite_Tracking_Provider_17Track implements Tanzanite_Tracking_Provider_Interface {

        private string $api_key;
        private string $secret_key;
        private string $endpoint;
        private array $config;

        public function __construct( array $settings, array $config ) {
            $this->api_key    = sanitize_text_field( $settings['api_key'] ?? '' );
            $this->secret_key = sanitize_text_field( $settings['secret_key'] ?? '' );
            $this->endpoint   = $config['fields']['endpoint']['default'] ?? 'https://api.17track.net';

            if ( ! empty( $settings['endpoint'] ) ) {
                $this->endpoint = esc_url_raw( $settings['endpoint'] );
            }

            $this->config = $config;
        }

        public function get_name(): string {
            return '17track';
        }

        public function test_connection(): bool {
            return ! empty( $this->api_key );
        }

        public function fetch_tracking_events( string $tracking_number, array $args = [] ) {
            if ( empty( $tracking_number ) ) {
                return new \WP_Error( 'invalid_tracking_payload', __( '物流单号不能为空。', 'tanzanite-settings' ) );
            }

            if ( empty( $this->api_key ) ) {
                return new \WP_Error( 'invalid_tracking_payload', __( '未配置 17TRACK API Key。', 'tanzanite-settings' ) );
            }

            $body = [
                'numbers' => [ $tracking_number ],
            ];

            if ( ! empty( $args['carrier'] ) ) {
                $body['carrier_code'] = sanitize_text_field( $args['carrier'] );
            }

            $response = $this->request( '/track', $body );

            if ( is_wp_error( $response ) ) {
                return $response;
            }

            $events = [];

            if ( isset( $response['data'][0]['track_info']['tracking_details'] ) && is_array( $response['data'][0]['track_info']['tracking_details'] ) ) {
                foreach ( $response['data'][0]['track_info']['tracking_details'] as $detail ) {
                    $events[] = $this->normalize_event( (array) $detail );
                }
            } elseif ( isset( $response['data'][0]['track_info']['latest_event'] ) ) {
                $events[] = $this->normalize_event( (array) $response['data'][0]['track_info']['latest_event'] );
            }

            if ( empty( $events ) ) {
                return new \WP_Error( 'tracking_provider_response_error', __( '未能解析到有效的追踪事件。', 'tanzanite-settings' ), $response );
            }

            return $events;
        }

        private function normalize_event( array $event ): array {
            return [
                'event_code'   => sanitize_key( $event['event_code'] ?? '' ),
                'status_text'  => $event['status'] ?? ( $event['description'] ?? '' ),
                'location'     => $event['location'] ?? '',
                'event_time'   => $event['time'] ?? ( $event['event_time'] ?? '' ),
                'raw'          => $event,
            ];
        }

        private function request( string $path, array $payload ) {
            $url  = trailingslashit( rtrim( $this->endpoint, '/' ) ) . ltrim( $path, '/' );
            $body = array_filter( $payload, static function ( $value ) {
                return null !== $value && '' !== $value;
            } );

            $args = [
                'headers' => [
                    'Content-Type' => 'application/json',
                    'Accept'       => 'application/json',
                    '17token'      => $this->api_key,
                ],
                'timeout' => 20,
                'body'    => wp_json_encode( $body ),
            ];

            if ( ! empty( $this->secret_key ) ) {
                $args['headers']['17secret'] = $this->secret_key;
            }

            $response = wp_remote_post( $url, $args );

            if ( is_wp_error( $response ) ) {
                return new \WP_Error( 'tracking_provider_request_failed', __( '调用 17TRACK 接口失败。', 'tanzanite-settings' ), $response->get_error_messages() );
            }

            $code = wp_remote_retrieve_response_code( $response );
            $body = wp_remote_retrieve_body( $response );

            if ( $code >= 400 ) {
                return new \WP_Error( 'tracking_provider_http_error', sprintf( __( '17TRACK 返回错误状态码：%d', 'tanzanite-settings' ), $code ), $body );
            }

            $decoded = json_decode( $body, true );

            if ( null === $decoded ) {
                return new \WP_Error( 'tracking_provider_parse_error', __( '无法解析 17TRACK 返回的数据。', 'tanzanite-settings' ), $body );
            }

            if ( isset( $decoded['code'] ) && 0 !== (int) $decoded['code'] ) {
                $message = $decoded['message'] ?? __( '17TRACK 接口返回错误。', 'tanzanite-settings' );
                return new \WP_Error( 'tracking_provider_response_error', $message, $decoded );
            }

            return $decoded;
        }
    }

    final class Tanzanite_Settings_Plugin {

        private const VERSION             = '0.1.9';
        private const DB_VERSION          = '0.1.8';
        private const OPTION_DB_VERSION        = 'tanzanite_settings_db_version';
        private const OPTION_TRACKING_SETTINGS = 'tanz_tracking_settings';
        private const ALLOWED_PRODUCT_STATUSES = [ 'draft', 'pending', 'publish', 'private' ];
        private const ALLOWED_ORDER_STATUSES   = [ 'pending', 'paid', 'processing', 'shipped', 'completed', 'cancelled' ];
        private const ORDER_STATUS_TRANSITIONS = [
            'pending'    => [ 'pending', 'paid', 'processing', 'cancelled' ],
            'paid'       => [ 'paid', 'processing', 'shipped', 'cancelled' ],
            'processing' => [ 'processing', 'shipped', 'cancelled' ],
            'shipped'    => [ 'shipped', 'completed' ],
            'completed'  => [ 'completed' ],
            'cancelled'  => [ 'cancelled' ],
        ];
        private const ALLOWED_REVIEW_STATUSES   = [ 'pending', 'approved', 'rejected', 'hidden' ];
        private const SHIPPING_RULE_TYPES      = [ 'weight', 'quantity', 'volume', 'amount', 'items' ];
        private const BULK_PRODUCT_ACTIONS     = [ 'set_status', 'adjust_stock', 'set_meta', 'adjust_price', 'set_featured', 'delete', 'export' ];
        private const BULK_ORDER_ACTIONS       = [ 'set_status', 'export' ];
        private const PAYMENT_TERMINALS         = [ 'pc', 'h5', 'app', 'mini_program', 'kiosk' ];
        private const TRACKING_PROVIDERS        = [
            '17track' => [
                'label' => '17TRACK',
                'fields' => [
                    'api_key'    => [ 'label' => 'API Key',    'type' => 'password' ],
                    'secret_key' => [ 'label' => 'Secret Key', 'type' => 'password' ],
                    'endpoint'   => [ 'label' => 'API Endpoint', 'type' => 'text', 'default' => 'https://api.17track.net/track' ],
                ],
            ],
        ];
        private const CAPABILITIES              = [
            'tanz_view_products',
            'tanz_manage_products',
            'tanz_bulk_products',
            'tanz_view_orders',
            'tanz_manage_orders',
            'tanz_bulk_orders',
            'tanz_manage_reviews',
            'tanz_manage_payments',
            'tanz_manage_shipping',
            'tanz_view_audit_logs',
            'tanz_manage_settings',
        ];
        private const ROLE_DEFINITIONS          = [
            'tanz_product_operator' => [
                'label'        => 'Tanzanite Product Operator',
                'capabilities' => [
                    'tanz_view_products',
                    'tanz_manage_products',
                    'tanz_bulk_products',
                    'tanz_manage_reviews',
                    'tanz_view_orders',
                ],
            ],
            'tanz_order_support' => [
                'label'        => 'Tanzanite Order Support',
                'capabilities' => [
                    'tanz_view_products',
                    'tanz_view_orders',
                    'tanz_manage_orders',
                    'tanz_bulk_orders',
                    'tanz_view_audit_logs',
                ],
            ],
            'tanz_logistics_manager' => [
                'label'        => 'Tanzanite Logistics Manager',
                'capabilities' => [
                    'tanz_view_products',
                    'tanz_view_orders',
                    'tanz_manage_shipping',
                ],
            ],
            'tanz_supervisor' => [
                'label'        => 'Tanzanite Supervisor',
                'capabilities' => 'all',
            ],
        ];
        private const PRODUCT_META_MAP         = [
            'price_regular'          => [ 'key' => '_tanz_price_regular', 'type' => 'number', 'default' => 0.0 ],
            'price_sale'             => [ 'key' => '_tanz_price_sale', 'type' => 'number', 'default' => 0.0 ],
            'price_member'           => [ 'key' => '_tanz_price_member', 'type' => 'number', 'default' => 0.0 ],
            'stock_qty'              => [ 'key' => '_tanz_stock_qty', 'type' => 'integer', 'default' => 0 ],
            'stock_alert'            => [ 'key' => '_tanz_stock_alert', 'type' => 'integer', 'default' => 0 ],
            'points_reward'          => [ 'key' => '_tanz_points_reward', 'type' => 'integer', 'default' => 0 ],
            'points_limit'           => [ 'key' => '_tanz_points_limit', 'type' => 'integer', 'default' => 0 ],
            'purchase_limit'         => [ 'key' => '_tanz_purchase_limit', 'type' => 'integer', 'default' => 0 ],
            'min_purchase'           => [ 'key' => '_tanz_min_purchase', 'type' => 'integer', 'default' => 0 ],
            'content_markdown'       => [ 'key' => '_tanz_content_markdown', 'type' => 'string', 'default' => '' ],
            'tier_pricing'           => [ 'key' => '_tanz_tier_pricing', 'type' => 'array', 'default' => [], 'item_type' => 'mixed', 'sanitize_callback' => 'tanzanite_settings_sanitize_tier_pricing' ],
            'backorders_allowed'     => [ 'key' => '_tanz_backorders_allowed', 'type' => 'boolean', 'default' => false ],
            'featured_image_id'      => [ 'key' => '_tanz_featured_image_id', 'type' => 'integer', 'default' => 0 ],
            'featured_image_url'     => [ 'key' => '_tanz_featured_image_url', 'type' => 'string', 'default' => '' ],
            'gallery_media_ids'      => [ 'key' => '_tanz_gallery_media_ids', 'type' => 'array', 'default' => [], 'item_type' => 'int' ],
            'gallery_external_urls'  => [ 'key' => '_tanz_gallery_external_urls', 'type' => 'array', 'default' => [], 'item_type' => 'string' ],
            'featured_video_id'      => [ 'key' => '_tanz_featured_video_id', 'type' => 'integer', 'default' => 0 ],
            'featured_video_url'     => [ 'key' => '_tanz_featured_video_url', 'type' => 'string', 'default' => '' ],
            'membership_levels'      => [ 'key' => '_tanz_membership_levels', 'type' => 'array', 'default' => [], 'item_type' => 'string' ],
            'shipping_template_id'   => [ 'key' => '_tanz_shipping_template_id', 'type' => 'integer', 'default' => 0 ],
            'free_shipping'          => [ 'key' => '_tanz_free_shipping', 'type' => 'boolean', 'default' => false ],
            'shipping_time'          => [ 'key' => '_tanz_shipping_time', 'type' => 'string', 'default' => '' ],
            'logistics_tags'         => [ 'key' => '_tanz_logistics_tags', 'type' => 'array', 'default' => [], 'item_type' => 'string' ],
            'channels'               => [ 'key' => '_tanz_channels', 'type' => 'array', 'default' => [], 'item_type' => 'string' ],
            'is_sticky'              => [ 'key' => '_tanz_is_sticky', 'type' => 'boolean', 'default' => false ],
            'featured_flag'          => [ 'key' => '_tanz_featured_flag', 'type' => 'boolean', 'default' => false ],
            'featured_slot'          => [ 'key' => '_tanz_featured_slot', 'type' => 'string', 'default' => '' ],
            'tax_rate_ids'           => [ 'key' => '_tanz_tax_rate_ids', 'type' => 'array', 'default' => [], 'item_type' => 'int' ],
        ];
        private const ERROR_CODE_GUIDE = [
            'failed_create_product'        => '创建商品失败，请稍后重试。',
            'failed_persist_skus'          => '保存 SKU 时发生异常，请检查数据后重试。',
            'invalid_sku_payload'          => 'SKU 数据格式不正确，请检查导入的 JSON。',
            'invalid_sku_entry'            => '某个 SKU 项目格式错误，请确认字段齐全。',
            'invalid_sku_code'             => 'SKU 缺少 sku_code，请补充后重试。',
            'duplicate_sku_code'           => 'SKU 编码重复，请确保唯一。',
            'invalid_tier_price'           => '阶梯价数据格式错误，请检查层级信息。',
            'invalid_tier_qty'             => '阶梯价最小购买量必须大于 0。',
            'product_not_found'            => '指定的商品不存在或已被删除。',
            'invalid_order_items'          => '订单明细不能为空，请至少添加一个商品项。',
            'invalid_order_item'           => '订单明细格式错误，请检查字段。',
            'invalid_order_item_quantity'  => '订单明细数量必须大于 0。',
            'invalid_order_item_title'     => '订单明细缺少商品标题。',
            'invalid_order_status'         => '订单状态流转不合法，已被拦截。',
            'failed_create_order'          => '创建订单失败，请稍后重试。',
            'failed_update_order'          => '更新订单失败，请稍后重试。',
            'failed_delete_order'          => '删除订单失败，请稍后重试。',
            'order_not_found'              => '指定的订单不存在或已被删除。',
            'invalid_rule'                 => '配送规则格式错误，请检查输入。',
            'invalid_rule_type'            => '配送规则类型不合法。',
            'invalid_rule_range'           => '配送规则区间设置不正确。',
            'failed_create_shipping_template' => '创建配送模板失败，请稍后重试。',
            'failed_update_shipping_template' => '更新配送模板失败，请稍后重试。',
            'failed_delete_shipping_template' => '删除配送模板失败，请稍后重试。',
            'shipping_template_not_found'  => '指定的配送模板不存在或已被删除。',
            'failed_create_payment_method' => '创建支付方式失败，请稍后重试。',
            'failed_update_payment_method' => '更新支付方式失败，请稍后重试。',
            'failed_delete_payment_method' => '删除支付方式失败，请稍后重试。',
            'payment_method_not_found'     => '指定的支付方式不存在或已被删除。',
            'invalid_payment_payload'      => '支付方式数据格式不正确，请检查传入字段。',
            'duplicate_payment_code'       => '支付方式编码已存在，请更换。',
            'invalid_bulk_action'          => '批量操作类型不支持。',
            'invalid_bulk_payload'         => '批量操作参数缺失或格式无效。',
            'partial_bulk_failure'         => '部分项目处理失败，请查看失败详情。',
            'invalid_tracking_payload'     => '物流追踪配置不正确，请检查填写的参数。',
            'tracking_provider_not_supported' => '所选物流追踪服务当前不受支持。',
            'tracking_provider_not_configured' => '当前站点尚未配置对应的物流追踪服务。',
            'tracking_provider_request_failed' => '调用外部物流追踪服务失败。',
            'tracking_provider_http_error' => '外部物流追踪服务返回错误状态码。',
            'tracking_provider_parse_error' => '无法解析物流追踪服务返回的数据。',
            'tracking_provider_response_error' => '物流追踪服务返回异常结果。',
            'failed_sync_tracking'         => '同步物流轨迹失败，请稍后重试。',
        ];
        private const ERROR_ACTION_STEPS = [
            'failed_persist_skus' => [ '检查导入列是否齐全', '确认价格/库存为有效数字', '重新导入或手动保存' ],
            'invalid_sku_payload' => [ '确保 JSON/CSV 字段顺序为 sku_code,price_regular,price_sale,stock_qty', '必要时使用下载的模板重新填写' ],
            'duplicate_sku_code'  => [ '修改重复的 SKU 编码', '保留唯一的 SKU 记录后再次提交' ],
            'invalid_tier_price'  => [ '验证阶梯价格 JSON 格式', '确认每个阶梯包含 min_qty 与 price' ],
            'invalid_order_items' => [ '至少添加一个订单明细', '确认每个商品数量大于 0' ],
            'invalid_order_status'=> [ '刷新订单状态后重试', '仅按流程允许的状态进行切换' ],
            'failed_update_order' => [ '确认网络正常', '查看审计日志获取详情', '必要时联系运维处理' ],
            'failed_delete_order' => [ '确认订单未被其他人操作', '重试或联系运维' ],
            'failed_create_shipping_template' => [ '检查规则字段格式', '确保必填信息完整后重试' ],
            'invalid_rule_type'   => [ '仅使用支持的配送规则类型：weight/quantity/volume/amount/items' ],
            'shipping_template_not_found' => [ '刷新页面重新加载模板列表', '确认模板未被删除' ],
            'failed_create_payment_method' => [ '检查支付方式编码是否唯一', '确认手续费类型与数值有效', '重新保存' ],
            'failed_update_payment_method' => [ '确认未被其他管理员同时修改', '刷新后重试或查看审计日志' ],
            'failed_delete_payment_method' => [ '确认支付方式未被订单引用', '必要时先禁用再删除' ],
            'invalid_payment_payload'      => [ '检查终端/会员等级是否为有效值', '确认 fee_type 与 fee_value 设置正确' ],
            'duplicate_payment_code'       => [ '更换唯一的支付方式编码', '确认未与其他支付方式重复' ],
            'invalid_bulk_action'          => [ '确认批量操作类型是否填写正确', '仅使用系统支持的批量操作 key' ],
            'invalid_bulk_payload'         => [ '确认传入的 ID 列表非空', '检查额外参数是否齐全（状态/增量等）' ],
            'partial_bulk_failure'         => [ '查看 failed 列表重新处理', '检查是否存在权限或状态冲突导致失败' ],
            'invalid_tracking_payload'     => [ '确认已选择可用的追踪服务', '检查 API Key/Secret 是否填写完整', '如使用自定义 Endpoint，请确认地址可访问' ],
            'tracking_provider_not_supported' => [ '更换为受支持的物流追踪服务', '联系开发者扩展新的适配器' ],
            'tracking_provider_not_configured' => [ '在“Tracking Providers”页面选择并保存服务商凭据', '确认订单的追踪服务与站点配置一致' ],
            'tracking_provider_request_failed' => [ '检查服务器网络与防火墙设置', '确认 17TRACK 服务状态正常', '稍后重试或联系供应商' ],
            'tracking_provider_http_error' => [ '查看返回的状态码及响应体', '确认 API Key 权限是否足够', '必要时切换备用 Endpoint' ],
            'tracking_provider_parse_error' => [ '检查 17TRACK 返回值格式', '联系供应商确认是否存在接口变更' ],
            'tracking_provider_response_error' => [ '查看响应 message 字段', '根据提示修正请求参数或联系供应商' ],
            'failed_sync_tracking'         => [ '确认订单的追踪号/服务商填写正确', '稍后再次尝试同步', '必要时手动在前端刷新轨迹' ],
        ];

        private static ?Tanzanite_Settings_Plugin $instance = null;

        private string $orders_table;
        private string $order_items_table;
        private string $shipping_templates_table;
        private string $product_skus_table;
        private string $audit_log_table;
        private string $payment_methods_table;
        private string $product_reviews_table;
        private string $tracking_events_table;
        private string $tax_rates_table;
        private string $product_attributes_table;
        private string $attribute_values_table;
        private string $carriers_table;
        private array $tracking_provider_instances = [];

        /**
         * 获取插件单例。
         */
        public static function instance(): Tanzanite_Settings_Plugin {
            if ( null === self::$instance ) {
                self::$instance = new self();
            }

            return self::$instance;
        }

        private function permission_callback( string $capability, bool $require_write_context = false ): callable {
            return function ( \WP_REST_Request $request ) use ( $capability, $require_write_context ) {
                if ( ! current_user_can( $capability ) ) {
                    return $this->permission_error();
                }

                if ( ! $require_write_context ) {
                    return true;
                }

                if ( ! $this->has_valid_write_context( $request ) ) {
                    return $this->write_context_error();
                }

                return true;
            };
        }

        private function has_valid_write_context( \WP_REST_Request $request ): bool {
            if ( $this->verify_request_nonce( $request ) ) {
                return true;
            }

            $auth_header = $request->get_header( 'authorization' );
            if ( $auth_header && $this->header_starts_with( strtolower( $auth_header ), 'basic ' ) ) {
                return true;
            }

            return false;
        }

        private function permission_error(): \WP_Error {
            return new \WP_Error(
                'tanz_insufficient_permission',
                __( '您没有执行该操作的权限。', 'tanzanite-settings' ),
                [ 'status' => rest_authorization_required_code() ]
            );
        }

        private function write_context_error(): \WP_Error {
            return new \WP_Error(
                'tanz_invalid_request_context',
                __( '请求缺少有效的安全校验，请刷新后台页面后重试。', 'tanzanite-settings' ),
                [ 'status' => rest_authorization_required_code() ]
            );
        }

        private function __construct() {
            global $wpdb;

            $this->orders_table             = $wpdb->prefix . 'tanz_orders';
            $this->order_items_table        = $wpdb->prefix . 'tanz_order_items';
            $this->shipping_templates_table = $wpdb->prefix . 'tanz_shipping_templates';
            $this->product_skus_table       = $wpdb->prefix . 'tanz_product_skus';
            $this->audit_log_table          = $wpdb->prefix . 'tanz_audit_logs';
            $this->payment_methods_table    = $wpdb->prefix . 'tanz_payment_methods';
            $this->product_reviews_table    = $wpdb->prefix . 'tanz_product_reviews';
            $this->tracking_events_table    = $wpdb->prefix . 'tanz_tracking_events';
            $this->tax_rates_table          = $wpdb->prefix . 'tanz_tax_rates';
            $this->product_attributes_table = $wpdb->prefix . 'tanz_product_attributes';
            $this->attribute_values_table   = $wpdb->prefix . 'tanz_attribute_values';
            $this->carriers_table           = $wpdb->prefix . 'tanz_carriers';
            $this->member_profiles_table    = $wpdb->prefix . 'tanz_member_profiles';
            $this->coupons_table            = $wpdb->prefix . 'tanz_coupons';
            $this->giftcards_table          = $wpdb->prefix . 'tanz_giftcards';
            $this->rewards_transactions_table = $wpdb->prefix . 'tanz_rewards_transactions';

            add_action( 'init', [ $this, 'register_content_types' ] );
            add_action( 'init', [ $this, 'maybe_upgrade_schema' ], 5 );
            add_action( 'admin_menu', [ $this, 'register_admin_menu' ] );
            add_action( 'admin_enqueue_scripts', [ $this, 'enqueue_admin_assets' ] );
            add_filter( 'admin_body_class', [ $this, 'filter_admin_body_class' ] );
            add_action( 'rest_api_init', [ $this, 'register_rest_routes' ] );
            add_action( 'admin_post_tanz_save_tracking_settings', [ $this, 'handle_save_tracking_settings' ] );
            add_action( 'admin_post_tz_save_redeem_settings', [ $this, 'handle_save_redeem_settings' ] );
            add_action( 'admin_init', [ $this, 'register_loyalty_settings' ] );
            add_action( 'wp_ajax_reset_loyalty_to_english', [ $this, 'handle_reset_loyalty_to_english' ] );
            
            error_log( 'Tanzanite Settings Plugin: Hooks registered, including rest_api_init' );
        }

        /**
         * 注册商品相关的内容类型与分类。
         */
        public function register_content_types(): void {
            register_post_type(
                'tanz_product',
                [
                    'labels' => [
                        'name'          => __( 'Tanzanite Products', 'tanzanite-settings' ),
                        'singular_name' => __( 'Tanzanite Product', 'tanzanite-settings' ),
                    ],
                    'public'              => false,
                    'show_ui'             => true,
                    'show_in_menu'        => false,
                    'show_in_rest'        => true,
                    'rest_base'           => 'tanz-products',
                    'menu_position'       => 56,
                    'capability_type'     => 'product',
                    'map_meta_cap'        => true,
                    'supports'            => [ 'title', 'editor', 'excerpt', 'thumbnail', 'custom-fields', 'revisions' ],
                    'rewrite'             => false,
                    'has_archive'         => false,
                ]
            );

            register_taxonomy(
                'tanz_product_category',
                'tanz_product',
                [
                    'labels'            => [
                        'name'          => __( 'Product Categories', 'tanzanite-settings' ),
                        'singular_name' => __( 'Product Category', 'tanzanite-settings' ),
                    ],
                    'public'           => false,
                    'hierarchical'     => true,
                    'show_ui'          => true,
                    'show_in_menu'     => false,
                    'show_in_rest'     => true,
                    'rest_base'        => 'tanz-product-categories',
                    'rewrite'          => false,
                ]
            );

            register_taxonomy(
                'tanz_product_tag',
                'tanz_product',
                [
                    'labels'            => [
                        'name'          => __( 'Product Tags', 'tanzanite-settings' ),
                        'singular_name' => __( 'Product Tag', 'tanzanite-settings' ),
                    ],
                    'public'           => false,
                    'hierarchical'     => false,
                    'show_ui'          => true,
                    'show_in_menu'     => false,
                    'show_in_rest'     => true,
                    'rest_base'        => 'tanz-product-tags',
                    'rewrite'          => false,
                ]
            );

        }

        /**
         * 检查并升级自定义表结构。
         */
        public function maybe_upgrade_schema(): void {
            $stored_version = get_option( self::OPTION_DB_VERSION );
            if ( self::DB_VERSION === $stored_version ) {
                return;
            }
            $this->create_tables();
            update_option( self::OPTION_DB_VERSION, self::DB_VERSION );
        }

        /**
         * 创建自定义表结构。
         */
        private function create_tables(): void {
            global $wpdb;

            $charset = $wpdb->get_charset_collate();

            $orders_sql = "CREATE TABLE {$this->orders_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                order_number VARCHAR(64) NOT NULL,
                user_id BIGINT UNSIGNED DEFAULT 0,
                status VARCHAR(32) NOT NULL DEFAULT 'pending',
                payment_method VARCHAR(64) DEFAULT NULL,
                channel VARCHAR(32) DEFAULT NULL,
                total DECIMAL(18,2) NOT NULL DEFAULT 0,
                subtotal DECIMAL(18,2) NOT NULL DEFAULT 0,
                discount_total DECIMAL(18,2) NOT NULL DEFAULT 0,
                shipping_total DECIMAL(18,2) NOT NULL DEFAULT 0,
                coupon_code VARCHAR(64) NULL,
                coupon_discount DECIMAL(18,2) NOT NULL DEFAULT 0,
                giftcard_discount DECIMAL(18,2) NOT NULL DEFAULT 0,
                points_used INT NOT NULL DEFAULT 0,
                currency CHAR(3) NOT NULL DEFAULT 'USD',
                tracking_provider VARCHAR(60) DEFAULT NULL,
                tracking_number VARCHAR(120) DEFAULT NULL,
                tracking_synced_at DATETIME NULL,
                paid_at DATETIME NULL,
                shipped_at DATETIME NULL,
                completed_at DATETIME NULL,
                cancelled_at DATETIME NULL,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY  (id),
                UNIQUE KEY order_number (order_number),
                KEY status (status),
                KEY user_id (user_id),
                KEY coupon_code (coupon_code)
            ) {$charset};";

            $order_items_sql = "CREATE TABLE {$this->order_items_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                order_id BIGINT UNSIGNED NOT NULL,
                product_id BIGINT UNSIGNED DEFAULT 0,
                sku_id BIGINT UNSIGNED DEFAULT 0,
                product_title VARCHAR(255) NOT NULL,
                sku_code VARCHAR(120) DEFAULT NULL,
                quantity INT NOT NULL DEFAULT 1,
                price DECIMAL(18,2) NOT NULL DEFAULT 0,
                total DECIMAL(18,2) NOT NULL DEFAULT 0,
                sort_order INT NOT NULL DEFAULT 0,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY order_id (order_id),
                KEY product_id (product_id)
            ) {$charset};";

            $product_skus_sql = "CREATE TABLE {$this->product_skus_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                product_id BIGINT UNSIGNED NOT NULL,
                sku_code VARCHAR(120) NOT NULL,
                attributes LONGTEXT NULL,
                price_regular DECIMAL(18,2) NOT NULL DEFAULT 0,
                price_sale DECIMAL(18,2) NOT NULL DEFAULT 0,
                stock_qty INT NOT NULL DEFAULT 0,
                tier_prices LONGTEXT NULL,
                sort_order INT NOT NULL DEFAULT 0,
                weight DECIMAL(12,3) DEFAULT NULL,
                barcode VARCHAR(120) DEFAULT NULL,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                UNIQUE KEY product_sku_unique (product_id, sku_code),
                KEY product_id (product_id)
            ) {$charset};";

            $shipping_sql = "CREATE TABLE {$this->shipping_templates_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                template_name VARCHAR(120) NOT NULL,
                description TEXT NULL,
                rules LONGTEXT NOT NULL,
                is_active TINYINT(1) NOT NULL DEFAULT 1,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY is_active (is_active)
            ) {$charset};";

            dbDelta( $orders_sql );
            dbDelta( $order_items_sql );
            dbDelta( $product_skus_sql );
            dbDelta( $shipping_sql );

            $payments_sql = "CREATE TABLE {$this->payment_methods_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                code VARCHAR(80) NOT NULL,
                name VARCHAR(160) NOT NULL,
                description TEXT NULL,
                icon_url VARCHAR(512) NULL,
                fee_type VARCHAR(20) NOT NULL DEFAULT 'fixed',
                fee_value DECIMAL(18,4) NOT NULL DEFAULT 0,
                terminals TEXT NULL,
                membership_levels TEXT NULL,
                currencies TEXT NULL,
                default_currency VARCHAR(10) NOT NULL DEFAULT '',
                settings LONGTEXT NULL,
                is_enabled TINYINT(1) NOT NULL DEFAULT 1,
                sort_order INT NOT NULL DEFAULT 0,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                UNIQUE KEY payment_code_unique (code),
                KEY is_enabled (is_enabled)
            ) {$charset};";

            dbDelta( $payments_sql );

            $reviews_sql = "CREATE TABLE {$this->product_reviews_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                product_id BIGINT UNSIGNED NOT NULL,
                user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
                author_name VARCHAR(120) DEFAULT NULL,
                author_email VARCHAR(160) DEFAULT NULL,
                author_phone VARCHAR(40) DEFAULT NULL,
                rating TINYINT UNSIGNED NOT NULL DEFAULT 5,
                content TEXT NULL,
                images LONGTEXT NULL,
                status VARCHAR(20) NOT NULL DEFAULT 'pending',
                is_featured TINYINT(1) NOT NULL DEFAULT 0,
                reply_text TEXT NULL,
                reply_author VARCHAR(120) DEFAULT NULL,
                reply_at DATETIME NULL,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY product_id (product_id),
                KEY user_id (user_id),
                KEY status (status),
                KEY created_at (created_at)
            ) {$charset};";

            dbDelta( $reviews_sql );

            $tracking_events_sql = "CREATE TABLE {$this->tracking_events_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                order_id BIGINT UNSIGNED NOT NULL,
                provider VARCHAR(60) NOT NULL,
                tracking_number VARCHAR(120) NOT NULL,
                event_code VARCHAR(60) DEFAULT NULL,
                status_text VARCHAR(255) DEFAULT NULL,
                location VARCHAR(255) DEFAULT NULL,
                event_time DATETIME DEFAULT NULL,
                meta LONGTEXT NULL,
                raw_payload LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY order_id (order_id),
                KEY provider (provider),
                KEY tracking_number (tracking_number),
                KEY event_time (event_time)
            ) {$charset};";

            dbDelta( $tracking_events_sql );

            $tax_rates_sql = "CREATE TABLE {$this->tax_rates_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                name VARCHAR(120) NOT NULL,
                rate DECIMAL(8,4) NOT NULL DEFAULT 0,
                region VARCHAR(120) NOT NULL DEFAULT '',
                description TEXT NULL,
                is_active TINYINT(1) NOT NULL DEFAULT 1,
                sort_order INT NOT NULL DEFAULT 0,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY is_active (is_active)
            ) {$charset};";

            dbDelta( $tax_rates_sql );

            $attributes_sql = "CREATE TABLE {$this->product_attributes_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                name VARCHAR(120) NOT NULL,
                slug VARCHAR(120) NOT NULL,
                type VARCHAR(32) NOT NULL DEFAULT 'select',
                sort_order INT NOT NULL DEFAULT 0,
                is_filterable TINYINT(1) NOT NULL DEFAULT 1,
                affects_sku TINYINT(1) NOT NULL DEFAULT 1,
                affects_stock TINYINT(1) NOT NULL DEFAULT 0,
                is_enabled TINYINT(1) NOT NULL DEFAULT 1,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                UNIQUE KEY slug (slug),
                KEY sort_order (sort_order),
                KEY is_enabled (is_enabled)
            ) {$charset};";

            dbDelta( $attributes_sql );

            $attribute_values_sql = "CREATE TABLE {$this->attribute_values_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                attribute_id BIGINT UNSIGNED NOT NULL,
                name VARCHAR(120) NOT NULL,
                slug VARCHAR(120) NOT NULL,
                value VARCHAR(255) NULL,
                sort_order INT NOT NULL DEFAULT 0,
                is_enabled TINYINT(1) NOT NULL DEFAULT 1,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY attribute_id (attribute_id),
                KEY sort_order (sort_order),
                UNIQUE KEY attribute_slug (attribute_id, slug)
            ) {$charset};";

            dbDelta( $attribute_values_sql );

            $carriers_sql = "CREATE TABLE {$this->carriers_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                code VARCHAR(64) NOT NULL,
                name VARCHAR(120) NOT NULL,
                contact_person VARCHAR(120) NULL,
                contact_phone VARCHAR(32) NULL,
                tracking_url VARCHAR(512) NULL,
                service_regions LONGTEXT NULL,
                is_active TINYINT(1) NOT NULL DEFAULT 1,
                sort_order INT NOT NULL DEFAULT 0,
                meta LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                UNIQUE KEY code (code),
                KEY is_active (is_active),
                KEY sort_order (sort_order)
            ) {$charset};";

            dbDelta( $carriers_sql );

            $audit_sql = "CREATE TABLE {$this->audit_log_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                actor_id BIGINT UNSIGNED DEFAULT 0,
                actor_name VARCHAR(120) DEFAULT NULL,
                action VARCHAR(60) NOT NULL,
                target_type VARCHAR(60) NOT NULL,
                target_id BIGINT UNSIGNED DEFAULT 0,
                payload LONGTEXT NULL,
                ip_address VARCHAR(64) DEFAULT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY target_lookup (target_type, target_id)
            ) {$charset};";

            dbDelta( $audit_sql );

            $member_profiles_table = $wpdb->prefix . 'tanz_member_profiles';
            $member_profiles_sql   = "CREATE TABLE {$member_profiles_table} (
                user_id BIGINT UNSIGNED NOT NULL,
                full_name VARCHAR(190) DEFAULT NULL,
                phone VARCHAR(64) DEFAULT NULL,
                country VARCHAR(100) DEFAULT NULL,
                address VARCHAR(255) DEFAULT NULL,
                brand VARCHAR(190) DEFAULT NULL,
                points INT UNSIGNED DEFAULT 0,
                notes TEXT DEFAULT NULL,
                marketing_optin TINYINT(1) DEFAULT 0,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (user_id),
                KEY points (points)
            ) {$charset};";

            dbDelta( $member_profiles_sql );

            // 优惠券表
            $coupons_table = $wpdb->prefix . 'tanz_coupons';
            $coupons_sql   = "CREATE TABLE {$coupons_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                code VARCHAR(64) NOT NULL,
                title VARCHAR(190) NOT NULL,
                description TEXT NULL,
                reward_type VARCHAR(32) NOT NULL DEFAULT 'coupon',
                template_id BIGINT UNSIGNED NULL,
                owner_user_id BIGINT UNSIGNED NULL,
                discount_type VARCHAR(20) NOT NULL DEFAULT 'fixed_amount',
                amount DECIMAL(12,2) NOT NULL DEFAULT 0,
                points_required INT UNSIGNED NOT NULL DEFAULT 0,
                min_points INT UNSIGNED NOT NULL DEFAULT 0,
                usage_limit INT UNSIGNED NOT NULL DEFAULT 1,
                usage_count INT UNSIGNED NOT NULL DEFAULT 0,
                status VARCHAR(20) NOT NULL DEFAULT 'draft',
                metadata LONGTEXT NULL,
                expires_at DATETIME NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                UNIQUE KEY code_unique (code),
                KEY owner_user_id (owner_user_id),
                KEY template_id (template_id),
                KEY status (status)
            ) {$charset};";

            dbDelta( $coupons_sql );

            // 礼品卡表
            $giftcards_table = $wpdb->prefix . 'tanz_giftcards';
            $giftcards_sql   = "CREATE TABLE {$giftcards_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                card_code VARCHAR(64) NOT NULL,
                balance DECIMAL(12,2) NOT NULL DEFAULT 0,
                original_value DECIMAL(12,2) NOT NULL DEFAULT 0,
                currency VARCHAR(16) NOT NULL DEFAULT 'USD',
                owner_user_id BIGINT UNSIGNED NULL,
                purchaser_user_id BIGINT UNSIGNED NULL,
                points_spent INT UNSIGNED NOT NULL DEFAULT 0,
                cover_image VARCHAR(512) NULL,
                status VARCHAR(20) NOT NULL DEFAULT 'active',
                expires_at DATETIME NULL,
                metadata LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                UNIQUE KEY card_code_unique (card_code),
                KEY owner_user_id (owner_user_id),
                KEY status (status)
            ) {$charset};";

            dbDelta( $giftcards_sql );

            // 积分交易记录表
            $rewards_transactions_table = $wpdb->prefix . 'tanz_rewards_transactions';
            $rewards_transactions_sql   = "CREATE TABLE {$rewards_transactions_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                user_id BIGINT UNSIGNED NULL,
                related_type VARCHAR(32) NOT NULL,
                related_id BIGINT UNSIGNED NULL,
                action VARCHAR(32) NOT NULL,
                points_delta INT NOT NULL DEFAULT 0,
                amount_delta DECIMAL(12,2) NOT NULL DEFAULT 0,
                notes TEXT NULL,
                metadata LONGTEXT NULL,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY user_id (user_id),
                KEY related_lookup (related_type, related_id),
                KEY created_at (created_at)
            ) {$charset};";

            dbDelta( $rewards_transactions_sql );

            // 聊天会话表
            $chat_conversations_table = $wpdb->prefix . 'tanz_chat_conversations';
            $chat_conversations_sql   = "CREATE TABLE {$chat_conversations_table} (
                id VARCHAR(50) NOT NULL,
                customer_id BIGINT UNSIGNED NOT NULL,
                agent_id BIGINT UNSIGNED NOT NULL,
                status VARCHAR(20) NOT NULL DEFAULT 'active',
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY customer_id (customer_id),
                KEY agent_id (agent_id),
                KEY status (status),
                KEY updated_at (updated_at)
            ) {$charset};";

            // 聊天消息表
            $chat_messages_table = $wpdb->prefix . 'tanz_chat_messages';
            $chat_messages_sql   = "CREATE TABLE {$chat_messages_table} (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                conversation_id VARCHAR(50) NOT NULL,
                sender_id BIGINT UNSIGNED NOT NULL,
                sender_type VARCHAR(20) NOT NULL,
                message TEXT NOT NULL,
                type VARCHAR(20) NOT NULL DEFAULT 'text',
                attachment_url VARCHAR(500) NULL,
                is_read TINYINT(1) NOT NULL DEFAULT 0,
                created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                PRIMARY KEY (id),
                KEY conversation_id (conversation_id),
                KEY sender_id (sender_id),
                KEY is_read (is_read),
                KEY created_at (created_at)
            ) {$charset};";

            dbDelta( $chat_conversations_sql );
            dbDelta( $chat_messages_sql );

            update_option( self::OPTION_DB_VERSION, self::DB_VERSION, false );
            
            // 刷新 REST API 路由缓存
            delete_option( 'rewrite_rules' );
        }

        private function ensure_roles_and_caps(): void {
            $administrator = get_role( 'administrator' );

            if ( $administrator ) {
                foreach ( self::CAPABILITIES as $capability ) {
                    if ( ! $administrator->has_cap( $capability ) ) {
                        $administrator->add_cap( $capability );
                    }
                }
            }

            foreach ( self::ROLE_DEFINITIONS as $role_key => $definition ) {
                $capabilities = 'all' === $definition['capabilities'] ? self::CAPABILITIES : $definition['capabilities'];
                $role         = get_role( $role_key );

                if ( ! $role ) {
                    $caps_map = array_fill_keys( $capabilities, true );
                    add_role( $role_key, __( $definition['label'], 'tanzanite-settings' ), $caps_map );
                    continue;
                }

                foreach ( $capabilities as $capability ) {
                    if ( ! $role->has_cap( $capability ) ) {
                        $role->add_cap( $capability );
                    }
                }
            }
        }

        /**
         * 注册 REST API 占位路由。
         */
        public function register_rest_routes(): void {
            // 强制输出日志，无论是否开启 WP_DEBUG
            error_log( '=== Tanzanite Settings: register_rest_routes() called ===' );
            
            if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                error_log( 'Tanzanite Settings: Registering REST API routes' );
            }
            
            // 注册所有 REST API 控制器
            $controller_classes = array(
                'Tanzanite_REST_Orders_Controller',
                'Tanzanite_REST_Products_Controller',
                'Tanzanite_REST_Payments_Controller',
                'Tanzanite_REST_TaxRates_Controller',
                'Tanzanite_REST_Reviews_Controller',
                'Tanzanite_REST_Members_Controller',
                'Tanzanite_REST_Carriers_Controller',
                'Tanzanite_REST_Coupons_Controller',
                'Tanzanite_REST_Giftcards_Controller',
                'Tanzanite_REST_Attributes_Controller',
                'Tanzanite_REST_Audit_Controller',
                'Tanzanite_REST_ShippingTemplates_Controller',
                'Tanzanite_REST_Chat_Controller',
            );
            
            foreach ( $controller_classes as $class_name ) {
                try {
                    if ( ! class_exists( $class_name ) ) {
                        if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                            error_log( "Tanzanite Settings: Controller class not found: {$class_name}" );
                        }
                        continue;
                    }
                    
                    $controller = new $class_name();
                    $controller->register_routes();
                    
                    if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                        error_log( "Tanzanite Settings: Registered routes for {$class_name}" );
                    }
                } catch ( Exception $e ) {
                    if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                        error_log( "Tanzanite Settings: Failed to register {$class_name}: " . $e->getMessage() );
                    }
                }
            }
            
            // Loyalty Settings REST API - 公开访问会员等级配置
            register_rest_route(
                'tanzanite/v1',
                '/loyalty/settings',
                array(
                    'methods'             => \WP_REST_Server::READABLE,
                    'callback'            => array( $this, 'rest_get_loyalty_settings' ),
                    'permission_callback' => '__return_true', // 公开访问
                )
            );
            
            // Products REST API - 已迁移到 includes/rest-api/class-products-controller.php
            // 旧路由已注释，避免与新控制器冲突
            /*
            register_rest_route(
                'tanzanite/v1',
                '/products',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => 'is_user_logged_in',
                        'callback'            => [ $this, 'rest_list_products' ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::CREATABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
                        'callback'            => [ $this, 'rest_create_product' ],
                        'args'                => [
                            'title'   => [
                                'type'              => 'string',
                                'required'          => true,
                                'sanitize_callback' => 'sanitize_text_field',
                            ],
                            'content' => [
                                'type'              => 'string',
                                'sanitize_callback' => 'wp_kses_post',
                            ],
                            'status'  => [
                                'type'              => 'string',
                                'default'           => 'draft',
                                'sanitize_callback' => 'sanitize_key',
                            ],
                            'slug'    => [
                                'type'              => 'string',
                                'sanitize_callback' => 'sanitize_title',
                            ],
                            'excerpt' => [
                                'type'              => 'string',
                                'sanitize_callback' => 'sanitize_textarea_field',
                            ],
                            'date'    => [
                                'type' => 'string',
                            ],
                            'skus'    => [
                                'type'    => 'array',
                                'default' => [],
                            ],
                            'skus_bulk' => [
                                'type'              => 'string',
                                'sanitize_callback' => 'wp_kses_post',
                            ],
                            'price_regular'        => [ 'type' => 'number' ],
                            'price_sale'           => [ 'type' => 'number' ],
                            'price_member'         => [ 'type' => 'number' ],
                            'stock_qty'            => [ 'type' => 'integer' ],
                            'stock_alert'          => [ 'type' => 'integer' ],
                            'points_reward'        => [ 'type' => 'integer' ],
                            'points_limit'         => [ 'type' => 'integer' ],
                            'purchase_limit'       => [ 'type' => 'integer' ],
                            'min_purchase'         => [ 'type' => 'integer' ],
                            'tier_pricing'         => [ 'type' => 'array' ],
                            'backorders_allowed'   => [ 'type' => 'boolean' ],
                            'featured_image_id'    => [ 'type' => 'integer' ],
                            'featured_image_url'   => [ 'type' => 'string' ],
                            'gallery_media_ids'    => [ 'type' => 'array' ],
                            'gallery_external_urls'=> [ 'type' => 'array' ],
                            'featured_video_id'    => [ 'type' => 'integer' ],
                            'featured_video_url'   => [ 'type' => 'string' ],
                            'membership_levels'    => [ 'type' => 'array' ],
                            'shipping_template_id' => [ 'type' => 'integer' ],
                            'free_shipping'        => [ 'type' => 'boolean' ],
                            'shipping_time'        => [ 'type' => 'string' ],
                            'logistics_tags'       => [ 'type' => 'array' ],
                            'channels'             => [ 'type' => 'array' ],
                            'is_sticky'            => [ 'type' => 'boolean' ],
                        ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::EDITABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_bulk_products', true ),
                        'callback'            => [ $this, 'rest_bulk_products' ],
                        'args'                => [
                            'action'      => [ 'type' => 'string', 'required' => true ],
                            'ids'         => [ 'type' => 'array', 'required' => true ],
                            'payload'     => [ 'type' => 'object' ],
                        ],
                    ],
                ]
            );

            register_rest_route(
                'tanzanite/v1',
                '/products/(?P<id>\\d+)',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => 'is_user_logged_in',
                        'callback'            => [ $this, 'rest_get_product' ],
                        'args'                => [
                            'id' => [ 'validate_callback' => 'is_numeric' ],
                        ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::EDITABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
                        'callback'            => [ $this, 'rest_update_product' ],
                        'args'                => [
                            'id'      => [ 'validate_callback' => 'is_numeric' ],
                            'title'   => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'content' => [ 'type' => 'string', 'sanitize_callback' => 'wp_kses_post' ],
                            'status'  => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ],
                            'slug'    => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_title' ],
                            'excerpt' => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_textarea_field' ],
                            'date'    => [ 'type' => 'string' ],
                            'price_regular' => [ 'type' => 'number' ],
                            'price_sale'    => [ 'type' => 'number' ],
                            'price_member'  => [ 'type' => 'number' ],
                            'stock_qty'     => [ 'type' => 'integer' ],
                            'stock_alert'   => [ 'type' => 'integer' ],
                            'points_reward' => [ 'type' => 'integer' ],
                            'points_limit'  => [ 'type' => 'integer' ],
                            'purchase_limit'       => [ 'type' => 'integer' ],
                            'min_purchase'         => [ 'type' => 'integer' ],
                            'tier_pricing'         => [ 'type' => 'array' ],
                            'backorders_allowed'   => [ 'type' => 'boolean' ],
                            'featured_image_id'    => [ 'type' => 'integer' ],
                            'featured_image_url'   => [ 'type' => 'string' ],
                            'gallery_media_ids'    => [ 'type' => 'array' ],
                            'gallery_external_urls'=> [ 'type' => 'array' ],
                            'featured_video_id'    => [ 'type' => 'integer' ],
                            'featured_video_url'   => [ 'type' => 'string' ],
                            'membership_levels'    => [ 'type' => 'array' ],
                            'shipping_template_id' => [ 'type' => 'integer' ],
                            'free_shipping'        => [ 'type' => 'boolean' ],
                            'shipping_time'        => [ 'type' => 'string' ],
                            'logistics_tags'       => [ 'type' => 'array' ],
                            'channels'             => [ 'type' => 'array' ],
                            'is_sticky'            => [ 'type' => 'boolean' ],
                            'skus'          => [ 'type' => 'array' ],
                            'skus_bulk'     => [ 'type' => 'string', 'sanitize_callback' => 'wp_kses_post' ],
                        ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::DELETABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
                        'callback'            => [ $this, 'rest_delete_product' ],
                        'args'                => [
                            'id' => [ 'validate_callback' => 'is_numeric' ],
                        ],
                    ],
                ]
            );

            // Payment Methods REST API - 已迁移到 includes/rest-api/class-payments-controller.php

            // Tax Rates REST API - 已迁移到 includes/rest-api/class-taxrates-controller.php

            // ========== 属性 REST 路由 - 已迁移到 includes/rest-api/class-attributes-controller.php ==========

            // Orders REST API - 已迁移到 includes/rest-api/class-orders-controller.php
            /*
            register_rest_route(
                'tanzanite/v1',
                '/orders',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => 'is_user_logged_in',
                        'callback'            => [ $this, 'rest_list_orders' ],
                        'args'                => [
                            'page'              => [ 'type' => 'integer', 'default' => 1 ],
                            'per_page'          => [ 'type' => 'integer', 'default' => 20 ],
                            'status'            => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ],
                            'payment_method'    => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'channel'           => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'tracking_provider' => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ],
                            'customer_keyword'  => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'date_start'        => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'date_end'          => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                        ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::CREATABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_manage_orders', true ),
                        'callback'            => [ $this, 'rest_create_order' ],
                        'args'                => [
                            'channel'        => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'payment_method' => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'total'          => [ 'type' => 'number', 'default' => 0 ],
                            'subtotal'       => [ 'type' => 'number', 'default' => 0 ],
                            'discount_total' => [ 'type' => 'number', 'default' => 0 ],
                            'shipping_total' => [ 'type' => 'number', 'default' => 0 ],
                            'status'         => [ 'type' => 'string', 'default' => 'pending', 'sanitize_callback' => 'sanitize_key' ],
                            'tracking_provider' => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ],
                            'tracking_number'   => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'items'          => [ 'type' => 'array', 'required' => true ],
                        ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::EDITABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_bulk_orders', true ),
                        'callback'            => [ $this, 'rest_bulk_orders' ],
                        'args'                => [
                            'action'  => [ 'type' => 'string', 'required' => true ],
                            'ids'     => [ 'type' => 'array', 'required' => true ],
                            'payload' => [ 'type' => 'object' ],
                        ],
                    ],
                ]
            );

            register_rest_route(
                'tanzanite/v1',
                '/orders/(?P<id>\\d+)',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => 'is_user_logged_in',
                        'callback'            => [ $this, 'rest_get_order' ],
                        'args'                => [
                            'id' => [ 'validate_callback' => 'is_numeric' ],
                        ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::EDITABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_manage_orders', true ),
                        'callback'            => [ $this, 'rest_update_order' ],
                        'args'                => [
                            'id'             => [ 'validate_callback' => 'is_numeric' ],
                            'status'         => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ],
                            'payment_method' => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'channel'        => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'total'          => [ 'type' => 'number' ],
                            'subtotal'       => [ 'type' => 'number' ],
                            'discount_total' => [ 'type' => 'number' ],
                            'shipping_total' => [ 'type' => 'number' ],
                            'tracking_provider' => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_key' ],
                            'tracking_number'   => [ 'type' => 'string', 'sanitize_callback' => 'sanitize_text_field' ],
                            'items'          => [ 'type' => 'array' ],
                        ],
                    ],
                    [
                        'methods'             => \WP_REST_Server::DELETABLE,
                        'permission_callback' => $this->permission_callback( 'tanz_manage_orders', true ),
                        'callback'            => [ $this, 'rest_delete_order' ],
                        'args'                => [
                            'id' => [ 'validate_callback' => 'is_numeric' ],
                        ],
                    ],
                ]
            );

            // Shipping Templates REST API - 已迁移到 includes/rest-api/class-shipping-templates-controller.php

            register_rest_route(
                'tanzanite/v1',
                '/sku-importer/preview',
                [
                    'methods'             => \WP_REST_Server::CREATABLE,
                    'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
                    'callback'            => [ $this, 'rest_preview_sku_import' ],
                    'args'                => [
                        'product_id' => [ 'type' => 'integer', 'required' => true ],
                        'skus_bulk'  => [ 'type' => 'string', 'required' => true ],
                    ],
                ]
            );

            register_rest_route(
                'tanzanite/v1',
                '/sku-importer/apply',
                [
                    'methods'             => \WP_REST_Server::CREATABLE,
                    'permission_callback' => $this->permission_callback( 'tanz_manage_products', true ),
                    'callback'            => [ $this, 'rest_apply_sku_import' ],
                    'args'                => [
                        'product_id' => [ 'type' => 'integer', 'required' => true ],
                        'skus'       => [ 'type' => 'array', 'required' => true ],
                    ],
                ]
            );

            register_rest_route(
                'tanzanite/v1',
                '/orders/(?P<id>\\d+)/tracking',
                [
                    'methods'             => \WP_REST_Server::CREATABLE,
                    'permission_callback' => $this->permission_callback( 'tanz_manage_orders', true ),
                    'callback'            => [ $this, 'rest_sync_order_tracking' ],
                ]
            );
            */

            // Carriers REST API - 已迁移到 includes/rest-api/class-carriers-controller.php

            register_rest_route(
                'tanzanite/v1',
                '/taxonomies/(?P<taxonomy>[a-z0-9_\-]+)',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => '__return_true',
                        'callback'            => [ $this, 'rest_list_taxonomy_terms' ],
                        'args'                => [
                            'taxonomy' => [ 'required' => true ],
                            'per_page' => [ 'type' => 'integer', 'default' => 20 ],
                            'page'     => [ 'type' => 'integer', 'default' => 1 ],
                            'search'   => [ 'type' => 'string', 'default' => '' ],
                        ],
                    ],
                ]
            );

            // Categories 快捷路由
            register_rest_route(
                'tanzanite/v1',
                '/categories',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => '__return_true',
                        'callback'            => function( $request ) {
                            $request->set_param( 'taxonomy', 'tanz_product_category' );
                            return $this->rest_list_taxonomy_terms( $request );
                        },
                        'args'                => [
                            'per_page' => [ 'type' => 'integer', 'default' => 20 ],
                            'page'     => [ 'type' => 'integer', 'default' => 1 ],
                            'search'   => [ 'type' => 'string', 'default' => '' ],
                        ],
                    ],
                ]
            );

            // Tags 快捷路由
            register_rest_route(
                'tanzanite/v1',
                '/tags',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => '__return_true',
                        'callback'            => function( $request ) {
                            $request->set_param( 'taxonomy', 'tanz_product_tag' );
                            return $this->rest_list_taxonomy_terms( $request );
                        },
                        'args'                => [
                            'per_page' => [ 'type' => 'integer', 'default' => 20 ],
                            'page'     => [ 'type' => 'integer', 'default' => 1 ],
                            'search'   => [ 'type' => 'string', 'default' => '' ],
                        ],
                    ],
                ]
            );

            // Member Profiles API - 已迁移到 includes/rest-api/class-members-controller.php

            // Coupons API - 已迁移到 includes/rest-api/class-coupons-controller.php

            // Gift Cards API - 已迁移到 includes/rest-api/class-giftcards-controller.php

            // Rewards Transactions API
            register_rest_route(
                'tanzanite/v1',
                '/rewards-transactions',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => $this->permission_callback( 'manage_options' ),
                        'callback'            => [ $this, 'rest_list_transactions' ],
                    ],
                ]
            );

            // Customer Service Agents API
            register_rest_route(
                'tanzanite/v1',
                '/customer-service/agents',
                [
                    [
                        'methods'             => \WP_REST_Server::READABLE,
                        'permission_callback' => '__return_true', // 公开访问
                        'callback'            => [ $this, 'rest_get_cs_agents' ],
                    ],
                ]
            );

            // Reviews API - 已迁移到 includes/rest-api/class-reviews-controller.php
            // Audit Logs API - 已迁移到 includes/rest-api/class-audit-controller.php
        }

        public function rest_list_taxonomy_terms( \WP_REST_Request $request ): \WP_REST_Response {
            $taxonomy = sanitize_key( $request->get_param( 'taxonomy' ) );
            $per_page = (int) max( 1, min( 200, (int) ( $request->get_param( 'per_page' ) ?: 20 ) ) );
            $page     = (int) max( 1, (int) ( $request->get_param( 'page' ) ?: 1 ) );
            $search   = sanitize_text_field( (string) ( $request->get_param( 'search' ) ?: '' ) );

            if ( ! taxonomy_exists( $taxonomy ) ) {
                return new \WP_REST_Response( [ 'message' => __( '无效的分类法。', 'tanzanite-settings' ) ], 400 );
            }

            $limit = $per_page + 1;
            $args  = [
                'taxonomy'   => $taxonomy,
                'hide_empty' => false,
                'number'     => $limit,
                'offset'     => ( $page - 1 ) * $per_page,
                'orderby'    => 'name',
                'order'      => 'ASC',
            ];

            if ( '' !== $search ) {
                $args['search'] = $search;
            }

            $terms = get_terms( $args );

            if ( is_wp_error( $terms ) ) {
                return new \WP_REST_Response( [ 'message' => $terms->get_error_message() ], 500 );
            }

            $has_more = count( $terms ) > $per_page;
            if ( $has_more ) {
                $terms = array_slice( $terms, 0, $per_page );
            }

            $items = array_map(
                static function ( \WP_Term $term ): array {
                    return [
                        'term_id' => $term->term_id,
                        'slug'    => $term->slug,
                        'name'    => $term->name,
                    ];
                },
                $terms
            );

            return new \WP_REST_Response(
                [
                    'items' => $items,
                    'meta'  => [
                        'page'      => $page,
                        'per_page'  => $per_page,
                        'has_more'  => $has_more,
                        'next_page' => $has_more ? $page + 1 : null,
                        'search'    => $search,
                    ],
                ]
            );
        }

        public function render_order_detail_page(): void {
            if ( ! current_user_can( 'tanz_view_orders' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            $order_id = isset( $_GET['order_id'] ) ? (int) $_GET['order_id'] : 0; // phpcs:ignore WordPress.Security.NonceVerification.Recommended
            if ( $order_id <= 0 ) {
                wp_die( __( '缺少有效的订单 ID。', 'tanzanite-settings' ) );
            }

            $nonce        = wp_create_nonce( 'wp_rest' );
            $detail_url   = esc_url_raw( rest_url( 'tanzanite/v1/orders/' . $order_id ) );
            $sync_url     = esc_url_raw( rest_url( 'tanzanite/v1/orders/' . $order_id . '/tracking' ) );
            $can_manage   = current_user_can( 'tanz_manage_orders' );

            // 加载订单详情 JS
            wp_enqueue_script(
                'tz-order-detail',
                plugins_url( 'assets/js/order-detail.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-admin-common' ),
                self::VERSION,
                true
            );

            // 加载打印功能 JS
            wp_enqueue_script(
                'tz-order-print',
                plugins_url( 'assets/js/order-print.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-order-detail',
                'TzOrderDetailConfig',
                array(
                    'nonce'             => $nonce,
                    'detailUrl'         => $detail_url,
                    'updateUrl'         => $detail_url,
                    'syncUrl'           => $sync_url,
                    'orderId'           => $order_id,
                    'canManage'         => $can_manage,
                    'statusLabels'      => array(
                        'pending'    => __( '待支付', 'tanzanite-settings' ),
                        'paid'       => __( '已支付', 'tanzanite-settings' ),
                        'processing' => __( '处理中', 'tanzanite-settings' ),
                        'shipped'    => __( '已发货', 'tanzanite-settings' ),
                        'completed'  => __( '已完成', 'tanzanite-settings' ),
                        'cancelled'  => __( '已取消', 'tanzanite-settings' ),
                    ),
                    'statusTransitions' => self::ORDER_STATUS_TRANSITIONS,
                )
            );

            echo '<div class="tz-settings-wrapper tz-order-detail">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Order Detail', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '查看订单全量信息，支持刷新物流状态与发货处理。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-order-detail-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section tz-order-summary">';
            echo '      <div class="tz-order-summary-main">';
            echo '          <h2>' . esc_html__( '订单摘要', 'tanzanite-settings' ) . '</h2>';
            echo '          <div class="tz-order-summary-meta" id="tz-order-summary-meta"></div>';
            echo '      </div>';
            echo '      <div class="tz-order-summary-actions">';
            echo '          <a class="button" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-orders' ) ) . '">' . esc_html__( '返回订单列表', 'tanzanite-settings' ) . '</a>';
            echo '          <a class="button" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-orders-bulk' ) ) . '">' . esc_html__( '跳转批量工具', 'tanzanite-settings' ) . '</a>';
            echo '          <button class="button button-primary" id="tz-order-print-btn" onclick="if(window.currentOrderData) window.TanzaniteOrderPrint.printShippingLabel(window.currentOrderData);">' . esc_html__( '打印发货单', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section tz-order-actions">';
            echo '      <h2>' . esc_html__( '状态操作', 'tanzanite-settings' ) . '</h2>';
            echo '      <div id="tz-order-status-actions" class="tz-order-actions-buttons" style="display:flex;gap:8px;flex-wrap:wrap;"></div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section tz-order-shipping">';
            echo '      <h2>' . esc_html__( '发货处理', 'tanzanite-settings' ) . '</h2>';
            echo '      <form id="tz-order-shipping-form" class="tz-order-shipping-form" style="display:grid;gap:12px;max-width:420px;">';
            echo '          <label>' . esc_html__( '物流公司编码', 'tanzanite-settings' ) . '<input type="text" id="tz-order-shipping-provider" class="regular-text" placeholder="sfexpress" /></label>';
            echo '          <label>' . esc_html__( '运单号', 'tanzanite-settings' ) . '<input type="text" id="tz-order-shipping-number" class="regular-text" placeholder="请填写运单号" /></label>';
            echo '          <button type="submit" class="button button-primary" id="tz-order-shipping-submit">' . esc_html__( '提交发货', 'tanzanite-settings' ) . '</button>';
            echo '          <p class="description" id="tz-order-shipping-hint"></p>';
            echo '      </form>';
            echo '  </div>';

            echo '  <div class="tz-settings-section tz-order-timeline">';
            echo '      <h2>' . esc_html__( '状态时间线', 'tanzanite-settings' ) . '</h2>';
            echo '      <ul id="tz-order-timeline" class="tz-order-timeline-list"></ul>';
            echo '  </div>';

            echo '  <div class="tz-settings-grid" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(360px,1fr));gap:16px;">';
            echo '      <div class="tz-settings-section">';
            echo '          <h2>' . esc_html__( '客户信息', 'tanzanite-settings' ) . '</h2>';
            echo '          <div id="tz-order-customer"></div>';
            echo '      </div>';
            echo '      <div class="tz-settings-section">';
            echo '          <h2>' . esc_html__( '金额明细', 'tanzanite-settings' ) . '</h2>';
            echo '          <div id="tz-order-amounts"></div>';
            echo '      </div>';
            echo '      <div class="tz-settings-section">';
            echo '          <h2>' . esc_html__( '发票与备注', 'tanzanite-settings' ) . '</h2>';
            echo '          <div id="tz-order-invoice"></div>';
            echo '          <div id="tz-order-notes" style="margin-top:12px;"></div>';
            echo '      </div>';
        echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <h2>' . esc_html__( '商品明细', 'tanzanite-settings' ) . '</h2>';
            echo '      <div style="overflow:auto;">';
            echo '          <table class="widefat fixed striped" id="tz-order-items" style="min-width:1100px;">';
            echo '              <thead><tr>';
            foreach ( [ __( '商品', 'tanzanite-settings' ), __( 'SKU', 'tanzanite-settings' ), __( '单价', 'tanzanite-settings' ), __( '数量', 'tanzanite-settings' ), __( '优惠', 'tanzanite-settings' ), __( '小计', 'tanzanite-settings' ), __( '备注', 'tanzanite-settings' ) ] as $column ) {
                echo '<th>' . esc_html( $column ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <h2>' . esc_html__( '物流追踪', 'tanzanite-settings' ) . '</h2>';
            echo '      <div id="tz-order-tracking-summary" style="margin-bottom:12px;"></div>';
            echo '      <button class="button" id="tz-order-refresh-tracking" ' . ( $can_manage ? '' : 'disabled' ) . '>' . esc_html__( '刷新物流', 'tanzanite-settings' ) . '</button>';
            echo '      <div class="description" id="tz-order-tracking-hint"></div>';
            echo '      <table class="widefat striped" id="tz-order-tracking-events" style="margin-top:12px;">';
            echo '          <thead><tr><th>' . esc_html__( '时间', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '状态', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '地点', 'tanzanite-settings' ) . '</th></tr></thead><tbody></tbody>';
            echo '      </table>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <h2>' . esc_html__( '操作日志', 'tanzanite-settings' ) . '</h2>';
            echo '      <table class="widefat striped" id="tz-order-audit" style="margin-top:12px;">';
            echo '          <thead><tr><th>' . esc_html__( '时间', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '操作者', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '动作', 'tanzanite-settings' ) . '</th><th>' . esc_html__( '说明', 'tanzanite-settings' ) . '</th></tr></thead><tbody></tbody>';
            echo '      </table>';
            echo '  </div>';

            echo '</div>';

            $status_labels = [
                'pending'    => __( '待支付', 'tanzanite-settings' ),
                'paid'       => __( '已支付', 'tanzanite-settings' ),
                'processing' => __( '处理中', 'tanzanite-settings' ),
                'shipped'    => __( '已发货', 'tanzanite-settings' ),
                'completed'  => __( '已完成', 'tanzanite-settings' ),
                'cancelled'  => __( '已取消', 'tanzanite-settings' ),
            ];

            $amount_labels = [
                'total'          => __( '实付总额', 'tanzanite-settings' ),
                'subtotal'       => __( '商品小计', 'tanzanite-settings' ),
                'discount_total' => __( '优惠合计', 'tanzanite-settings' ),
                'shipping_total' => __( '运费', 'tanzanite-settings' ),
                'points_used'    => __( '使用积分', 'tanzanite-settings' ),
            ];

            $config_js = wp_json_encode(
                [
                    'nonce'      => $nonce,
                    'detailUrl'  => $detail_url,
                    'updateUrl'  => $detail_url,
                    'syncUrl'    => $sync_url,
                    'orderId'    => $order_id,
                    'canManage'  => $can_manage,
                    'statusLabels'      => $status_labels,
                    'statusTransitions' => self::ORDER_STATUS_TRANSITIONS,
                    'amountLabels'      => $amount_labels,
                    'strings'    => [
                        'loading'        => __( '加载中…', 'tanzanite-settings' ),
                        'refreshing'     => __( '正在刷新物流…', 'tanzanite-settings' ),
                        'refreshSuccess' => __( '物流状态已刷新。', 'tanzanite-settings' ),
                        'refreshFailed'  => __( '刷新物流失败。', 'tanzanite-settings' ),
                        'noData'         => __( '暂无数据', 'tanzanite-settings' ),
                        'setStatusTemplate'   => __( '切换至 %s', 'tanzanite-settings' ),
                        'statusNoActions'     => __( '暂无可执行的状态操作。', 'tanzanite-settings' ),
                        'statusNoPermission'  => __( '当前账号无权操作订单状态。', 'tanzanite-settings' ),
                        'operationSuccess'    => __( '操作成功。', 'tanzanite-settings' ),
                        'operationFailed'     => __( '操作失败。', 'tanzanite-settings' ),
                        'setStatusSuccess'    => __( '状态已更新。', 'tanzanite-settings' ),
                        'shippingSuccess'     => __( '发货信息已更新。', 'tanzanite-settings' ),
                        'shippingMissingFields'=> __( '请填写物流公司编码和运单号。', 'tanzanite-settings' ),
                        'shippingHint'        => __( '填写物流公司编码与运单号后提交即可完成发货。', 'tanzanite-settings' ),
                        'shippingNoPermission'=> __( '当前账号无权执行发货操作。', 'tanzanite-settings' ),
                        'trackingProviderLabel' => __( '服务商', 'tanzanite-settings' ),
                        'trackingNumberLabel'   => __( '运单号', 'tanzanite-settings' ),
                        'trackingSyncedLabel'   => __( '最近同步', 'tanzanite-settings' ),
                        'trackingEventStatusLabel'   => __( '状态', 'tanzanite-settings' ),
                        'trackingEventLocationLabel' => __( '地点', 'tanzanite-settings' ),
                        'trackingHintDefault'   => __( '如需刷新最新物流状态，可点击下方按钮。', 'tanzanite-settings' ),
                        'customerNoteLabel'     => __( '客户备注', 'tanzanite-settings' ),
                        'adminNoteLabel'        => __( '客服备注', 'tanzanite-settings' ),
                        'invoiceTypeLabel'      => __( '类型', 'tanzanite-settings' ),
                        'invoiceTitleLabel'     => __( '抬头', 'tanzanite-settings' ),
                        'invoiceTaxLabel'       => __( '税号', 'tanzanite-settings' ),
                        'invoiceContentLabel'   => __( '内容', 'tanzanite-settings' ),
                        'shippingButtonLabel'   => __( '提交发货', 'tanzanite-settings' ),
                        'saving'            => __( '正在保存商品…', 'tanzanite-settings' ),
                        'saveSuccess'       => __( '商品保存成功。', 'tanzanite-settings' ),
                        'saveFailed'        => __( '保存商品失败，请稍后重试。', 'tanzanite-settings' ),
                        'seoSaving'         => __( '正在同步 SEO 数据…', 'tanzanite-settings' ),
                        'seoSaveSuccess'    => __( 'SEO 数据已同步。', 'tanzanite-settings' ),
                        'seoSaveFailed'     => __( 'SEO 数据同步失败，请稍后重试。', 'tanzanite-settings' ),
                    ],
                ],
                JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES
            );


        }

        public function render_add_product(): void {
            if ( ! current_user_can( 'tanz_manage_products' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            $product_id = isset( $_GET['product_id'] ) ? absint( $_GET['product_id'] ) : 0;

            wp_enqueue_media();
            wp_enqueue_editor();

            $nonce            = wp_create_nonce( 'wp_rest' );
            $create_endpoint  = esc_url_raw( rest_url( 'tanzanite/v1/products' ) );
            $single_endpoint  = esc_url_raw( rest_url( 'tanzanite/v1/products/' ) );
            $media_endpoint   = esc_url_raw( rest_url( 'wp/v2/media' ) );
            $detail_endpoint  = $product_id ? esc_url_raw( trailingslashit( $single_endpoint ) . $product_id ) : '';
            $seo_endpoint     = esc_url_raw( rest_url( 'mytheme/v1/seo/product/' ) );
            
            // 从插件内部获取会员等级设置
            $loyalty_settings = $this->get_loyalty_settings();
            $member_tiers     = [];

            if ( ! empty( $loyalty_settings['tiers'] ) && is_array( $loyalty_settings['tiers'] ) ) {
                foreach ( $loyalty_settings['tiers'] as $tier ) {
                    if ( ! empty( $tier['name'] ) ) {
                        $member_tiers[] = [
                            'code'  => sanitize_key( $tier['name'] ),
                            'label' => sanitize_text_field( $tier['label'] ?? $tier['name'] ),
                        ];
                    }
                }
            }

            $initial_skus = [];
            if ( $product_id ) {
                $initial_skus = array_map(
                    static function ( array $sku ): array {
                        $attributes_input = '';
                        if ( isset( $sku['attributes'] ) && is_array( $sku['attributes'] ) && ! empty( $sku['attributes'] ) ) {
                            $pairs = [];
                            foreach ( $sku['attributes'] as $key => $value ) {
                                if ( is_array( $value ) ) {
                                    foreach ( $value as $single ) {
                                        $pairs[] = sprintf( '%s=%s', $key, $single );
                                    }
                                } elseif ( '' !== (string) $value ) {
                                    $pairs[] = sprintf( '%s=%s', $key, $value );
                                }
                            }
                            $attributes_input = implode( '; ', $pairs );
                        }

                        $sku['attributes_input'] = $attributes_input;

                        return $sku;
                    },
                    $this->get_product_skus( $product_id )
                );
            }

            // 从数据库读取 Markdown 模板，如果不存在则使用默认值
            $saved_templates = get_option( 'tanzanite_markdown_templates', [] );
            $default_templates = $this->get_default_markdown_templates();
            $markdown_templates = wp_parse_args( $saved_templates, $default_templates );

            $markdown_rules = [
                'requiredSections'        => [
                    [
                        'label'   => __( '商品亮点段落', 'tanzanite-settings' ),
                        'keyword' => '商品亮点',
                    ],
                    [
                        'label'   => __( '规格参数段落', 'tanzanite-settings' ),
                        'keyword' => '规格参数',
                    ],
                    [
                        'label'   => __( '售后保障段落', 'tanzanite-settings' ),
                        'keyword' => '售后',
                    ],
                ],
                'forbiddenTerms'          => [ '国家级', '顶级保证', '100%治愈', '零风险', '最高标准' ],
                'placeholderTokens'       => [ 'TODO', '待补', '待完善', '占位', '示意图', '未定稿', '[图片占位]' ],
                'requireImages'           => 1,
                'imagePlaceholderKeywords'=> [ 'placeholder', 'temp', 'todo', 'example.com', 'dummy' ],
            ];

            echo '<div class="tz-settings-wrapper tz-product-editor">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Add New Product', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '按照商品规划完善基础信息、定价、库存与物流配置，保存后即可在 Nuxt 前端展示。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-product-editor-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <form id="tz-product-editor-form" class="tz-product-editor-form" autocomplete="off">';

            echo '      <section class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( '基础信息', 'tanzanite-settings' ) . '</div>';
            echo '          <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '              <label>' . esc_html__( '商品标题', 'tanzanite-settings' ) . '<input type="text" id="tz-product-title" class="regular-text" required /></label>';
            echo '              <label>' . esc_html__( '副标题 / 摘要', 'tanzanite-settings' ) . '<textarea id="tz-product-excerpt" rows="3" class="widefat"></textarea></label>';
            echo '              <label>' . esc_html__( 'Slug（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-product-slug" class="regular-text" /></label>';
            echo '              <div>';
            echo '                  <label>' . esc_html__( 'URL 自定义路径（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-product-urllink-path" class="regular-text" placeholder="例如：products/category/product-name" /></label>';
            echo '                  <p class="description" style="margin:4px 0 0 0;">' . esc_html__( '自定义此商品的完整 URL 路径。留空则使用默认的 WordPress 固定链接。示例：products/electronics/phone', 'tanzanite-settings' ) . '</p>';
            echo '              </div>';
            echo '          </div>';
            echo '      </section>';

            echo '      <section class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( '详情内容', 'tanzanite-settings' ) . '</div>';
            echo '          <p class="description">' . esc_html__( '支持标题、段落、图片与富文本排版，内容将同步输出至 Nuxt 前端。', 'tanzanite-settings' ) . '</p>';

            ob_start();
            wp_editor(
                '',
                'tz-product-content',
                [
                    'textarea_name'  => 'tz-product-content',
                    'media_buttons'  => true,
                    'textarea_rows'  => 18,
                    'editor_height'  => 360,
                    'drag_drop_upload' => true,
                    'tinymce'        => [
                        'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
                        'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
                    ],
                    'quicktags'      => true,
                ]
            );
            $editor_html = ob_get_clean();
            echo $editor_html;

            echo '          <div class="tz-markdown-toolbar" style="display:flex;flex-wrap:wrap;gap:12px;margin-top:12px;align-items:center;background:#f9f9f9;padding:12px;border:1px solid #ddd;border-radius:4px;">';
            echo '              <label style="display:flex;align-items:center;gap:6px;font-weight:500;color:#23282d;font-size:13px;">';
            echo '                  <input type="checkbox" id="tz-product-content-toggle" style="margin:0;width:16px;height:16px;flex-shrink:0;" /> ' . esc_html__( '启用 Markdown 模式（左侧编辑、右侧预览）', 'tanzanite-settings' );
            echo '              </label>';
            echo '              <span class="description" style="color:#666;font-size:12px;">' . esc_html__( '可选用下方模板快速填充内容，Markdown 将即时同步至富文本编辑器。', 'tanzanite-settings' ) . '</span>';
            echo '          </div>';

            echo '          <div class="tz-markdown-templates" style="display:flex;flex-wrap:wrap;gap:8px;margin-top:8px;">';
            echo '              <button type="button" class="button" data-markdown-template="basic">' . esc_html__( '插入「商品亮点」模板', 'tanzanite-settings' ) . '</button>';
            echo '              <button type="button" class="button" data-markdown-template="spec">' . esc_html__( '插入「规格参数」模板', 'tanzanite-settings' ) . '</button>';
            echo '              <button type="button" class="button" data-markdown-template="after_sale">' . esc_html__( '插入「售后说明」模板', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';

            echo '          <div id="tz-product-content-markdown-wrapper" class="tz-markdown-wrapper" style="display:none;grid-template-columns:1fr 1fr;gap:16px;margin-top:16px;">';
            echo '              <div class="tz-markdown-column" style="display:flex;flex-direction:column;gap:8px;">';
            echo '                  <label for="tz-product-content-markdown" style="font-weight:600;">' . esc_html__( 'Markdown 输入', 'tanzanite-settings' ) . '</label>';
            echo '                  <textarea id="tz-product-content-markdown" class="widefat" style="min-height:360px;font-family:SFMono-Regular,Menlo,Monaco,Consolas,\'Courier New\',monospace;"></textarea>';
            echo '              </div>';
            echo '              <div class="tz-markdown-column" style="display:flex;flex-direction:column;gap:8px;">';
            echo '                  <label style="font-weight:600;">' . esc_html__( '实时预览', 'tanzanite-settings' ) . '</label>';
            echo '                  <div id="tz-product-content-preview" class="tz-markdown-preview" style="min-height:360px;border:1px solid #dcdfe5;border-radius:6px;padding:12px;background:#fff;overflow:auto;"></div>';
            echo '              </div>';
            echo '          </div>';

            echo '      </section>';

            echo '      <section class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( '媒体资源', 'tanzanite-settings' ) . '</div>';
            echo '          <div class="tz-section-body" style="display:grid;gap:20px;">';
            echo '              <div class="tz-media-block">';
            echo '                  <div class="tz-media-block__header">' . esc_html__( '主图', 'tanzanite-settings' ) . '</div>';
            echo '                  <div id="tz-product-featured-preview" class="tz-media-preview"></div>';
            echo '                  <div class="tz-media-actions" style="display:flex;gap:8px;flex-wrap:wrap;margin:12px 0;">';
            echo '                      <button type="button" class="button" id="tz-product-featured-select">' . esc_html__( '选择图片', 'tanzanite-settings' ) . '</button>';
            echo '                      <button type="button" class="button-link" id="tz-product-featured-clear">' . esc_html__( '清除', 'tanzanite-settings' ) . '</button>';
            echo '                  </div>';
            echo '                  <input type="hidden" id="tz-product-featured-id" />';
            echo '                  <label>' . esc_html__( '主图 URL（可选，覆盖上传结果）', 'tanzanite-settings' ) . '<input type="url" id="tz-product-featured-url" class="regular-text" placeholder="https://" /></label>';
            echo '              </div>';
            echo '              <div class="tz-media-block">';
            echo '                  <div class="tz-media-block__header">' . esc_html__( '图库', 'tanzanite-settings' ) . '</div>';
            echo '                  <div id="tz-product-gallery-preview" class="tz-media-gallery" style="display:flex;gap:12px;flex-wrap:wrap;"></div>';
            echo '                  <div class="tz-media-actions" style="display:flex;gap:8px;flex-wrap:wrap;margin:12px 0;">';
            echo '                      <button type="button" class="button" id="tz-product-gallery-select">' . esc_html__( '选择图片', 'tanzanite-settings' ) . '</button>';
            echo '                      <button type="button" class="button-link" id="tz-product-gallery-clear">' . esc_html__( '清空', 'tanzanite-settings' ) . '</button>';
            echo '                  </div>';
            echo '                  <input type="hidden" id="tz-product-gallery-ids" />';
            echo '                  <label>' . esc_html__( '外链图片 URL（每行一条，可选）', 'tanzanite-settings' ) . '<textarea id="tz-product-gallery-externals" rows="3" class="widefat" placeholder="https://example.com/image.jpg"></textarea></label>';
            echo '              </div>';
            echo '              <div class="tz-media-block">';
            echo '                  <div class="tz-media-block__header">' . esc_html__( '主图视频', 'tanzanite-settings' ) . '</div>';
            echo '                  <div id="tz-product-video-preview" class="tz-media-preview"></div>';
            echo '                  <div class="tz-media-actions" style="display:flex;gap:8px;flex-wrap:wrap;margin:12px 0;">';
            echo '                      <button type="button" class="button" id="tz-product-video-select">' . esc_html__( '选择视频', 'tanzanite-settings' ) . '</button>';
            echo '                      <button type="button" class="button-link" id="tz-product-video-clear">' . esc_html__( '清除', 'tanzanite-settings' ) . '</button>';
            echo '                  </div>';
            echo '                  <input type="hidden" id="tz-product-video-id" />';
            echo '                  <label>' . esc_html__( '视频 URL（可选，覆盖上传结果）', 'tanzanite-settings' ) . '<input type="url" id="tz-product-video-url" class="regular-text" placeholder="https://" /></label>';
            echo '              </div>';
            echo '          </div>';
            echo '      </section>';

            echo '      <section class="tz-settings-grid" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(340px,1fr));gap:16px;">';

            echo '          <div class="tz-settings-section">';
            echo '              <div class="tz-section-title">' . esc_html__( '价格与活动', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '                  <label>' . esc_html__( '原价', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-price-regular" class="regular-text" /></label>';
            echo '                  <label>' . esc_html__( '现价', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-price-sale" class="regular-text" /></label>';
            echo '                  <label>' . esc_html__( '会员价', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-price-member" class="regular-text" /></label>';
            echo '                  <label>' . esc_html__( '限购数量', 'tanzanite-settings' ) . '<input type="number" id="tz-product-limit" class="regular-text" /></label>';
            echo '                  <label>' . esc_html__( '起购数量', 'tanzanite-settings' ) . '<input type="number" id="tz-product-min-purchase" class="regular-text" /></label>';
            echo '                  <div class="tz-tier-pricing" style="display:flex;flex-direction:column;gap:12px;">';
            echo '                      <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:8px;">';
            echo '                          <strong>' . esc_html__( '阶梯价配置', 'tanzanite-settings' ) . '</strong>';
            echo '                          <button type="button" class="button" id="tz-tier-template">' . esc_html__( '插入示例阶梯', 'tanzanite-settings' ) . '</button>';
            echo '                      </div>';
            echo '                      <p class="description">' . esc_html__( '按购买数量区间设置折扣或特价。示例：10 件起 95 折，50 件起 9 折。可留空表示不启用阶梯价。', 'tanzanite-settings' ) . '</p>';
            echo '                      <table class="widefat striped" id="tz-tier-pricing-table" style="margin:0;">';
            echo '                          <thead><tr>';
            echo '                              <th style="width:20%;">' . esc_html__( '最小数量', 'tanzanite-settings' ) . '</th>';
            echo '                              <th style="width:20%;">' . esc_html__( '最大数量（可选）', 'tanzanite-settings' ) . '</th>';
            echo '                              <th style="width:20%;">' . esc_html__( '单价 / 折扣', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '说明', 'tanzanite-settings' ) . '</th>';
            echo '                              <th style="width:120px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
            echo '                          </tr></thead>';
            echo '                          <tbody></tbody>';
            echo '                      </table>';
            echo '                      <div id="tz-tier-empty" class="notice notice-info" style="margin:0;">' . esc_html__( '尚未配置阶梯价。', 'tanzanite-settings' ) . '</div>';
            echo '                      <input type="hidden" id="tz-product-tier-pricing" />';
            echo '                      <form id="tz-tier-form" style="display:grid;gap:12px;padding:12px;border:1px solid #dcdfe5;border-radius:8px;background:#f9fafb;">';
            echo '                          <input type="hidden" id="tz-tier-index" value="" />';
            echo '                          <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px;">';
            echo '                              <label>' . esc_html__( '最小数量', 'tanzanite-settings' ) . '<input type="number" min="1" step="1" id="tz-tier-min" class="regular-text" required /></label>';
            echo '                              <label>' . esc_html__( '最大数量（选填）', 'tanzanite-settings' ) . '<input type="number" min="1" step="1" id="tz-tier-max" class="regular-text" /></label>';
            echo '                              <label>' . esc_html__( '单价 / 折扣', 'tanzanite-settings' ) . '<input type="number" min="0" step="0.01" id="tz-tier-price" class="regular-text" required /></label>';
            echo '                              <label>' . esc_html__( '备注（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-tier-note" class="regular-text" placeholder="' . esc_attr__( '例如：95 折', 'tanzanite-settings' ) . '" /></label>';
            echo '                          </div>';
            echo '                          <div style="display:flex;gap:8px;flex-wrap:wrap;">';
            echo '                              <button type="button" class="button button-primary" id="tz-tier-submit">' . esc_html__( '保存阶梯价', 'tanzanite-settings' ) . '</button>';
            echo '                              <button type="button" class="button" id="tz-tier-reset">' . esc_html__( '清空表单', 'tanzanite-settings' ) . '</button>';
            echo '                          </div>';
            echo '                      </form>';
            echo '                  </div>';
            echo '              </div>';
            echo '          </div>';

            echo '          <div class="tz-settings-section">';
            echo '              <div class="tz-section-title">' . esc_html__( '库存与规格', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '                  <label>' . esc_html__( '总库存', 'tanzanite-settings' ) . '<input type="number" id="tz-product-stock" class="regular-text" /></label>';
            echo '                  <label>' . esc_html__( '库存预警值', 'tanzanite-settings' ) . '<input type="number" id="tz-product-stock-alert" class="regular-text" /></label>';
            echo '                  <label><input type="checkbox" id="tz-product-backorders" /> ' . esc_html__( '允许超卖 / 接受缺货订单', 'tanzanite-settings' ) . '</label>';
            
            echo '                  <div class="tz-product-attributes-selector" style="margin:16px 0;padding:16px;background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;">';
            echo '                      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">';
            echo '                          <strong>' . esc_html__( '商品属性选择', 'tanzanite-settings' ) . '</strong>';
            echo '                          <button type="button" class="button button-primary" id="tz-generate-skus" style="display:none;">' . esc_html__( '自动生成 SKU 组合', 'tanzanite-settings' ) . '</button>';
            echo '                      </div>';
            echo '                      <p class="description" style="margin:0 0 12px 0;">' . esc_html__( '从属性模板中选择影响 SKU 的属性，勾选需要的属性值，然后点击"自动生成 SKU 组合"。', 'tanzanite-settings' ) . '</p>';
            echo '                      <div id="tz-attributes-list" class="tz-attributes-list" style="display:flex;flex-direction:column;gap:16px;"></div>';
            echo '                      <div id="tz-attributes-loading" style="padding:20px;text-align:center;color:#6b7280;">' . esc_html__( '正在加载属性模板...', 'tanzanite-settings' ) . '</div>';
            echo '                      <div id="tz-attributes-empty" style="display:none;padding:20px;text-align:center;color:#6b7280;">' . esc_html__( '暂无可用属性。请先在「Attributes」页面创建属性并标记为"影响 SKU"。', 'tanzanite-settings' ) . '</div>';
            echo '                  </div>';
            
            echo '                  <div class="tz-sku-editor" id="tz-product-sku-editor" style="display:flex;flex-direction:column;gap:12px;">';
            echo '                      <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:8px;">';
            echo '                          <strong>' . esc_html__( 'SKU 组合与价格', 'tanzanite-settings' ) . '</strong>';
            echo '                          <span class="description" id="tz-product-sku-hint">' . esc_html__( '可从上方属性自动生成，或手动输入。属性格式：颜色=蓝;尺寸=16', 'tanzanite-settings' ) . '</span>';
            echo '                      </div>';
            echo '                      <table class="widefat fixed striped" id="tz-product-sku-table">';
            echo '                          <thead><tr>';
            echo '                              <th>' . esc_html__( 'SKU 编码', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '属性组合', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '原价', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '现价', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '库存', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '条码', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '重量(kg)', 'tanzanite-settings' ) . '</th>';
            echo '                              <th>' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
            echo '                          </tr></thead>';
            echo '                          <tbody></tbody>';
            echo '                      </table>';
            echo '                      <div id="tz-product-sku-empty" class="notice notice-info" style="margin:0;">';
            echo '                          <p>' . esc_html__( '尚未添加 SKU，请在下方表单中录入。', 'tanzanite-settings' ) . '</p>';
            echo '                      </div>';
            echo '                      <form id="tz-product-sku-form" class="tz-sku-form" style="display:grid;gap:12px;padding:12px;border:1px solid #dcdfe5;border-radius:8px;background:#f9fafb;">';
            echo '                          <input type="hidden" id="tz-product-sku-form-index" value="" />';
            echo '                          <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:12px;">';
            echo '                              <label>' . esc_html__( 'SKU 编码', 'tanzanite-settings' ) . '<input type="text" id="tz-product-sku-form-code" class="regular-text" required /></label>';
            echo '                              <label>' . esc_html__( '属性组合', 'tanzanite-settings' ) . '<input type="text" id="tz-product-sku-form-attrs" class="regular-text" placeholder="颜色=蓝;尺寸=16" /></label>';
            echo '                              <label>' . esc_html__( '原价', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-sku-form-price-regular" class="regular-text" /></label>';
            echo '                              <label>' . esc_html__( '现价', 'tanzanite-settings' ) . '<input type="number" step="0.01" id="tz-product-sku-form-price-sale" class="regular-text" /></label>';
            echo '                              <label>' . esc_html__( '库存', 'tanzanite-settings' ) . '<input type="number" id="tz-product-sku-form-stock" class="regular-text" /></label>';
            echo '                              <label>' . esc_html__( '条码', 'tanzanite-settings' ) . '<input type="text" id="tz-product-sku-form-barcode" class="regular-text" /></label>';
            echo '                              <label>' . esc_html__( '重量(kg)', 'tanzanite-settings' ) . '<input type="number" step="0.001" id="tz-product-sku-form-weight" class="regular-text" /></label>';
            echo '                          </div>';
            echo '                          <div style="display:flex;gap:8px;flex-wrap:wrap;">';
            echo '                              <button type="button" class="button button-primary" id="tz-product-sku-form-submit">' . esc_html__( '保存 / 新增 SKU', 'tanzanite-settings' ) . '</button>';
            echo '                              <button type="button" class="button" id="tz-product-sku-form-reset">' . esc_html__( '清空表单', 'tanzanite-settings' ) . '</button>';
            echo '                          </div>';
            echo '                      </form>';
            echo '                  </div>';
            echo '              </div>';
            echo '          </div>';
            echo '      </section>';

            echo '          <div class="tz-settings-section">';
            echo '              <div class="tz-section-title">' . esc_html__( '积分与会员', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '                  <label>' . esc_html__( '赠送积分', 'tanzanite-settings' ) . '<input type="number" id="tz-product-points-reward" class="regular-text" /></label>';
            echo '                  <label>' . esc_html__( '积分抵扣上限', 'tanzanite-settings' ) . '<input type="number" id="tz-product-points-limit" class="regular-text" /></label>';
            echo '                  <div>';
            echo '                      <div style="display:flex;align-items:center;justify-content:space-between;gap:8px;flex-wrap:wrap;">';
            echo '                          <strong>' . esc_html__( '适用会员等级', 'tanzanite-settings' ) . '</strong>';
            echo '                          <div style="display:flex;gap:6px;flex-wrap:wrap;">';
            echo '                              <button type="button" class="button" id="tz-member-select-all">' . esc_html__( '全选', 'tanzanite-settings' ) . '</button>';
            echo '                              <button type="button" class="button" id="tz-member-select-clear">' . esc_html__( '清除', 'tanzanite-settings' ) . '</button>';
            echo '                          </div>';
            echo '                      </div>';
            echo '                      <p class="description">' . esc_html__( '从会员系统同步的等级列表，可多选。未选择时表示对全部会员开放。', 'tanzanite-settings' ) . '</p>';
            echo '                      <div id="tz-product-membership-list" class="tz-membership-list" style="display:flex;flex-wrap:wrap;gap:8px;margin-top:8px;"></div>';
            echo '                      <input type="hidden" id="tz-product-membership" />';
            echo '                  </div>';
            echo '              </div>';
            echo '          </div>';

            echo '          <div class="tz-settings-section">';
            echo '              <div class="tz-section-title">' . esc_html__( '物流与配送', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '                  <label>' . esc_html__( '配送模板', 'tanzanite-settings' ) . '<select id="tz-product-shipping-template" class="widefat"><option value="">无</option></select></label>';
            echo '                  <label><input type="checkbox" id="tz-product-free-shipping" /> ' . esc_html__( '是否包邮', 'tanzanite-settings' ) . '</label>';
            echo '                  <label>' . esc_html__( '发货时效描述', 'tanzanite-settings' ) . '<input type="text" id="tz-product-shipping-time" class="regular-text" /></label>';
            echo '                  <textarea id="tz-product-logistics-tags" rows="2" class="widefat" placeholder="' . esc_attr__( '跨境 / 冷链标签等', 'tanzanite-settings' ) . '"></textarea>';
            echo '              </div>';
            echo '          </div>';

            echo '          <div class="tz-settings-section">';
            echo '              <div class="tz-section-title">' . esc_html__( '税率设置', 'tanzanite-settings' ) . '</div>';
            echo '              <p class="description">' . esc_html__( '勾选适用于此商品的税率模板，前端下单时将自动计算税费。', 'tanzanite-settings' ) . '</p>';
            echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '                  <div id="tz-product-tax-rates-list" class="tz-checkbox-list" style="display:grid;gap:8px;"></div>';
            echo '                  <p class="description" style="color:#646970;font-size:12px;">' . esc_html__( '如无可用税率模板，请先在「税率管理」页面创建。', 'tanzanite-settings' ) . '</p>';
            echo '              </div>';
            echo '          </div>';

            echo '          <div class="tz-settings-section">';
            echo '              <div class="tz-section-title">' . esc_html__( '发布与频道', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '                  <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select id="tz-product-status" class="widefat"><option value="draft">' . esc_html__( '草稿', 'tanzanite-settings' ) . '</option><option value="publish">' . esc_html__( '发布', 'tanzanite-settings' ) . '</option><option value="pending">' . esc_html__( '待审', 'tanzanite-settings' ) . '</option></select></label>';
            echo '                  <label>' . esc_html__( '发布时间', 'tanzanite-settings' ) . '<input type="datetime-local" id="tz-product-date" class="regular-text" /></label>';
            echo '                  <label><input type="checkbox" id="tz-product-sticky" /> ' . esc_html__( '置顶显示', 'tanzanite-settings' ) . '</label>';
            echo '                  <textarea id="tz-product-channels" rows="2" class="widefat" placeholder="' . esc_attr__( '专题频道 ID/别名（逗号或 JSON）', 'tanzanite-settings' ) . '"></textarea>';
            echo '              </div>';
            echo '          </div>';

            echo '      </section>';

            echo '      <section class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( 'SEO 优化', 'tanzanite-settings' ) . '</div>';
            echo '          <p class="description">' . esc_html__( '填写以下元信息以提升搜索引擎收录表现，保存商品后将自动同步至 SEO 插件。', 'tanzanite-settings' ) . '</p>';
            echo '          <div id="tz-product-seo-status" class="tz-seo-status" style="display:flex;align-items:center;gap:8px;"></div>';
            echo '          <div id="tz-product-seo-loading" style="display:none;">' . esc_html__( '正在加载 SEO 数据…', 'tanzanite-settings' ) . '</div>';
            echo '          <div class="tz-section-body" style="display:grid;gap:12px;">';
            echo '              <label>' . esc_html__( 'SEO 标题', 'tanzanite-settings' ) . '<input type="text" id="tz-product-seo-title" class="regular-text" placeholder="' . esc_attr__( '建议 30-60 个字符', 'tanzanite-settings' ) . '" /></label>';
            echo '              <label>' . esc_html__( 'SEO 描述', 'tanzanite-settings' ) . '<textarea id="tz-product-seo-description" rows="3" class="widefat" placeholder="' . esc_attr__( '建议 90-160 个字符，以自然语句概述商品亮点。', 'tanzanite-settings' ) . '"></textarea></label>';
            echo '              <label>' . esc_html__( 'SEO 关键字（可选）', 'tanzanite-settings' ) . '<input type="text" id="tz-product-seo-keywords" class="regular-text" placeholder="' . esc_attr__( '逗号分隔，如：戒指,蓝宝石,新品', 'tanzanite-settings' ) . '" /></label>';
            echo '          </div>';
            echo '          <div id="tz-product-seo-actions" style="display:none;align-items:center;gap:8px;">';
            echo '              <button type="button" class="button" id="tz-product-seo-refresh">' . esc_html__( '重新获取 SEO 数据', 'tanzanite-settings' ) . '</button>';
            echo '              <span class="description" id="tz-product-seo-hint"></span>';
            echo '          </div>';
            echo '          <input type="hidden" id="tz-product-seo-locale" value="" />';
            echo '      </section>';

            echo '      <section class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( 'API 输出预览', 'tanzanite-settings' ) . '</div>';
            echo '          <p class="description">' . esc_html__( '保存后将在此展示 REST 返回的结构，便于与 Nuxt 前端联调。', 'tanzanite-settings' ) . '</p>';
            echo '          <pre id="tz-product-preview" style="background:#f6f7f7;border:1px solid #ccd0d4;padding:12px;min-height:160px;overflow:auto;"></pre>';
            echo '      </section>';

            echo '      <div class="tz-product-editor-actions" style="display:flex;gap:12px;margin-top:24px;">';
            echo '          <button type="submit" class="button button-primary" id="tz-product-submit">' . esc_html__( '保存商品', 'tanzanite-settings' ) . '</button>';
            echo '          <button type="button" class="button" id="tz-product-save-draft">' . esc_html__( '保存为草稿', 'tanzanite-settings' ) . '</button>';
            echo '          <button type="button" class="button" id="tz-product-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';

            echo '  </form>';
            echo '</div>';

            $initial_meta = $product_id ? $this->get_product_meta_payload( $product_id ) : [];
            $initial_payload = [];
            if ( $product_id ) {
                $initial_payload = $this->build_product_response( $product_id );
            }

            $config_js_array = [
                    'nonce'              => $nonce,
                    'createEndpoint'     => $create_endpoint,
                    'singleEndpoint'     => $single_endpoint,
                    'mediaEndpoint'      => $media_endpoint,
                    'productId'          => $product_id,
                    'detailEndpoint'     => $detail_endpoint,
                    'seoEndpoint'        => $seo_endpoint,
                    'taxRatesEndpoint'   => esc_url_raw( rest_url( 'tanzanite/v1/tax-rates' ) ),
                    'attributesEndpoint' => esc_url_raw( rest_url( 'tanzanite/v1/attributes' ) ),
                    'shippingTemplatesEndpoint' => esc_url_raw( rest_url( 'tanzanite/v1/shipping-templates' ) ),
                    'initialSkuRows'     => $initial_skus,
                    'markdownTemplates'  => $markdown_templates,
                    'markdownRules'      => $markdown_rules,
                    'membershipTiers'    => $member_tiers,
                    'initialMeta'        => $initial_meta,
                    'initialPayload'     => $initial_payload,
                    'strings'            => [
                        'draftSaved'              => __( '草稿保存成功。', 'tanzanite-settings' ),
                        'resetConfirm'            => __( '确定要重置当前填写内容吗？', 'tanzanite-settings' ),
                        'loading'                 => __( '加载中…', 'tanzanite-settings' ),
                        'seoDraftHint'            => __( '保存商品后即可配置 SEO 元信息。', 'tanzanite-settings' ),
                        'seoLoadFailed'           => __( '加载 SEO 数据失败，请稍后重试。', 'tanzanite-settings' ),
                        'seoConfigured'           => __( 'SEO 数据已配置。', 'tanzanite-settings' ),
                        'seoNotConfigured'        => __( '尚未配置 SEO 数据。', 'tanzanite-settings' ),
                        'seoRefreshHint'          => __( '如需从 SEO 插件获取最新内容，请点击刷新。', 'tanzanite-settings' ),
                        'seoNotConfiguredHint'    => __( '填写下方字段并保存商品后，SEO 插件会同步这些信息。', 'tanzanite-settings' ),
                        'seoReloading'            => __( '正在重新获取…', 'tanzanite-settings' ),
                        'saving'                  => __( '正在保存商品…', 'tanzanite-settings' ),
                        'saveSuccess'             => __( '商品保存成功。', 'tanzanite-settings' ),
                        'saveFailed'              => __( '保存商品失败，请稍后重试。', 'tanzanite-settings' ),
                        'seoSaving'               => __( '正在同步 SEO 数据…', 'tanzanite-settings' ),
                        'seoSaveSuccess'          => __( 'SEO 数据已同步。', 'tanzanite-settings' ),
                        'seoSaveFailed'           => __( 'SEO 数据同步失败，请稍后重试。', 'tanzanite-settings' ),
                        'skuCreateLabel'          => __( '保存 / 新增 SKU', 'tanzanite-settings' ),
                        'skuUpdateLabel'          => __( '更新 SKU', 'tanzanite-settings' ),
                        'skuCodeRequired'         => __( '请填写 SKU 编码。', 'tanzanite-settings' ),
                        'skuDuplicate'            => __( '该 SKU 编码已存在，请修改后再保存。', 'tanzanite-settings' ),
                        'skuSaved'                => __( 'SKU 已保存。', 'tanzanite-settings' ),
                        'skuDeleteConfirm'        => __( '确定要删除该 SKU 吗？', 'tanzanite-settings' ),
                        'skuDeleted'              => __( 'SKU 已删除。', 'tanzanite-settings' ),
                        'skuEdit'                 => __( '编辑', 'tanzanite-settings' ),
                        'skuDelete'               => __( '删除', 'tanzanite-settings' ),
                        'templateMissing'         => __( '未找到对应的模板内容。', 'tanzanite-settings' ),
                        'attributeDictionaryHint' => __( '点击可快速插入属性组合。', 'tanzanite-settings' ),
                        'markdownValidationFailed'=> __( 'Markdown 内容未通过校验，请根据提示修改。', 'tanzanite-settings' ),
                        'markdownSectionMissing'  => __( '缺少以下必填段落：', 'tanzanite-settings' ),
                        'markdownSensitiveFound'  => __( '检测到需要替换的敏感描述：', 'tanzanite-settings' ),
                        'markdownPlaceholderFound'=> __( '检测到待补充的占位文案：', 'tanzanite-settings' ),
                        'markdownImageMissing'    => __( '至少需要插入一张商品图片。', 'tanzanite-settings' ),
                        'placeholderImage'        => __( '图像占位符', 'tanzanite-settings' ),
                        'markdownEmpty'           => __( '商品详情内容不能为空，请补充后再保存。', 'tanzanite-settings' ),
                        'memberSelectPlaceholder' => __( '请选择适用的会员等级（可多选）', 'tanzanite-settings' ),
                        'memberSelectAll'         => __( '全选', 'tanzanite-settings' ),
                        'memberSelectClear'       => __( '清除', 'tanzanite-settings' ),
                        'membershipInvalid'       => __( '存在无效的会员等级选择，请刷新页面或清除后再试。', 'tanzanite-settings' ),
                        'tierErrorRequired'       => __( '请填写有效的最小数量和单价。', 'tanzanite-settings' ),
                        'tierErrorRange'          => __( '最大数量必须大于或等于最小数量。', 'tanzanite-settings' ),
                        'tierErrorOverlap'        => __( '阶梯区间存在重叠或顺序错误，请检查后再保存。', 'tanzanite-settings' ),
                        'tierErrorLastUnlimited'  => __( '只有最后一个阶梯可以不设置最大数量。', 'tanzanite-settings' ),
                        'tierTemplateConfirm'     => __( '此操作将覆盖现有阶梯价配置，是否继续？', 'tanzanite-settings' ),
                        'tierTemplateApplied'     => __( '示例阶梯价已插入，请按需调整。', 'tanzanite-settings' ),
                        'tierSaved'               => __( '阶梯价配置已更新。', 'tanzanite-settings' ),
                        'mediaSelectTitle'        => __( '选择媒体资源', 'tanzanite-settings' ),
                        'mediaSelectButton'       => __( '使用所选', 'tanzanite-settings' ),
                        'mediaClearConfirm'       => __( '确定要清除该媒体吗？', 'tanzanite-settings' ),
                        'loadFailed'              => __( '加载商品数据失败，请稍后重试。', 'tanzanite-settings' ),
                    ],
                ];

            // 加载 Product Editor JS
            wp_enqueue_script(
                'tz-product-editor',
                plugins_url( 'assets/js/product-editor.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-admin-common', 'jquery', 'wp-util', 'media-upload', 'media-views' ),
                self::VERSION . '.sync.' . time(),
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-product-editor',
                'TzProductEditorConfig',
                $config_js_array
            );
        }

        private function render_tracking_settings(): void {
            if ( ! current_user_can( 'tanz_manage_shipping' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            // 加载 Tracking Settings JS
            wp_enqueue_script(
                'tz-tracking-settings',
                plugins_url( 'assets/js/tracking-settings.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            $config     = $this->get_tracking_settings();
            $provider   = $config['provider'];
            $settings   = $config['all_settings'];
            $providers  = self::TRACKING_PROVIDERS;
            $nonce      = wp_create_nonce( 'tanz_tracking_settings' );

            echo '<div class="tz-settings-wrapper tz-tracking-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Tracking Providers', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '配置物流追踪服务，可优先使用 17TRACK，并预留其他服务的适配能力。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            if ( isset( $_GET['updated'] ) ) {
                echo '<div class="notice notice-success is-dismissible"><p>' . esc_html__( '设置已保存。', 'tanzanite-settings' ) . '</p></div>';
            }

            echo '  <form method="post" action="' . esc_url( admin_url( 'admin-post.php' ) ) . '" class="tz-settings-section" style="max-width:900px;">';
            echo '      <input type="hidden" name="action" value="tanz_save_tracking_settings" />';
            echo '      <input type="hidden" name="_wpnonce" value="' . esc_attr( $nonce ) . '" />';

            echo '      <div class="tz-section-title">' . esc_html__( '服务商选择', 'tanzanite-settings' ) . '</div>';
            echo '      <label>' . esc_html__( '当前服务商', 'tanzanite-settings' ) . '<select name="provider" class="widefat">';
            foreach ( $providers as $key => $provider_config ) {
                echo '<option value="' . esc_attr( $key ) . '"' . selected( $provider, $key, false ) . '>' . esc_html( $provider_config['label'] ) . '</option>';
            }
            echo '      </select></label>';

            echo '      <div class="tz-section-title" style="margin-top:24px;">' . esc_html__( '认证信息', 'tanzanite-settings' ) . '</div>';
            echo '      <div class="tz-form-grid">';

            foreach ( $providers as $key => $provider_config ) {
                $hidden_class = $key === $provider ? '' : 'style="display:none;"';
                echo '<div class="tz-provider-fields" data-provider="' . esc_attr( $key ) . '" ' . $hidden_class . '>';

                foreach ( $provider_config['fields'] as $field_key => $field_config ) {
                    $value = $settings[ $key ][ $field_key ] ?? ( $field_config['default'] ?? '' );
                    $label = $field_config['label'] ?? ucfirst( $field_key );
                    $type  = $field_config['type'] ?? 'text';
                    echo '<label>' . esc_html( $label ) . '<input type="' . esc_attr( $type ) . '" name="settings[' . esc_attr( $key ) . '][' . esc_attr( $field_key ) . ']" value="' . esc_attr( $value ) . '" class="widefat" /></label>';
                }

                echo '</div>';
            }

            echo '      </div>';

            echo '      <p class="description">' . esc_html__( '提示：如需切换到其他服务商，请先在上方选择并填入所需凭据；后续开发可通过新增适配器拓展更多平台。', 'tanzanite-settings' ) . '</p>';

            echo '      <p><button type="submit" class="button button-primary">' . esc_html__( '保存设置', 'tanzanite-settings' ) . '</button></p>';

            echo '  </form>';

            echo '</div>';

        }

        private function get_tracking_settings(): array {
            $option = get_option( self::OPTION_TRACKING_SETTINGS, [] );

            if ( ! is_array( $option ) ) {
                $option = [];
            }

            $provider = sanitize_key( $option['provider'] ?? '17track' );
            if ( ! isset( self::TRACKING_PROVIDERS[ $provider ] ) ) {
                $provider = '17track';
            }

            $settings = $option['settings'] ?? [];
            if ( ! is_array( $settings ) ) {
                $settings = [];
            }

            return [
                'provider' => $provider,
                'settings' => $settings[ $provider ] ?? [],
                'all_settings' => $settings,
            ];
        }

        private function get_tracking_provider_config( string $provider ): array {
            return self::TRACKING_PROVIDERS[ $provider ] ?? self::TRACKING_PROVIDERS['17track'];
        }

        private function get_tracking_provider_fields( string $provider ): array {
            $config = $this->get_tracking_provider_config( $provider );

            return $config['fields'] ?? [];
        }

        private function get_tracking_provider_instance( string $provider ) {
            if ( isset( $this->tracking_provider_instances[ $provider ] ) ) {
                return $this->tracking_provider_instances[ $provider ];
            }

            if ( ! isset( self::TRACKING_PROVIDERS[ $provider ] ) ) {
                return new \WP_Error( 'tracking_provider_not_supported', __( '当前物流追踪服务暂不支持。', 'tanzanite-settings' ) );
            }

            $settings   = $this->get_tracking_settings();
            $all        = $settings['all_settings'];
            $provider_settings = $all[ $provider ] ?? [];

            switch ( $provider ) {
                case '17track':
                default:
                    $instance = new Tanzanite_Tracking_Provider_17Track( $provider_settings, $this->get_tracking_provider_config( $provider ) );
                    break;
            }

            if ( ! $instance->test_connection() ) {
                return new \WP_Error( 'tracking_provider_not_configured', __( '当前追踪服务尚未正确配置。', 'tanzanite-settings' ) );
            }

            $this->tracking_provider_instances[ $provider ] = $instance;

            return $instance;
        }

        private function normalize_datetime_value( $value ): ?string {
            if ( empty( $value ) ) {
                return null;
            }

            if ( is_numeric( $value ) ) {
                $timestamp = (int) $value;
            } else {
                $timestamp = strtotime( (string) $value );
            }

            if ( ! $timestamp ) {
                return null;
            }

            return wp_date( 'Y-m-d H:i:s', $timestamp );
        }

        private function sync_tracking_events( int $order_id, string $provider, string $tracking_number, string $context = 'manual' ) {
            global $wpdb;

            if ( '' === $provider || '' === $tracking_number ) {
                return new \WP_Error( 'invalid_tracking_payload', __( '缺少物流追踪服务或单号。', 'tanzanite-settings' ) );
            }

            $instance = $this->get_tracking_provider_instance( $provider );
            if ( is_wp_error( $instance ) ) {
                return $instance;
            }

            $result = $instance->fetch_tracking_events( $tracking_number );
            if ( is_wp_error( $result ) ) {
                return $result;
            }

            if ( ! is_array( $result ) ) {
                return new \WP_Error( 'tracking_provider_response_error', __( '物流追踪服务返回的数据格式不正确。', 'tanzanite-settings' ) );
            }

            $wpdb->delete( $this->tracking_events_table, [ 'order_id' => $order_id ], [ '%d' ] );

            $normalized_events = [];

            foreach ( $result as $event ) {
                if ( ! is_array( $event ) ) {
                    continue;
                }

                $event_code  = isset( $event['event_code'] ) ? sanitize_text_field( $event['event_code'] ) : '';
                $status_text = isset( $event['status_text'] ) ? sanitize_textarea_field( $event['status_text'] ) : '';
                $location    = isset( $event['location'] ) ? sanitize_text_field( $event['location'] ) : '';
                $event_time  = $this->normalize_datetime_value( $event['event_time'] ?? ( $event['time'] ?? '' ) );

                $meta_json = null;
                if ( isset( $event['meta'] ) && is_array( $event['meta'] ) ) {
                    $meta_json = wp_json_encode( $event['meta'], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES );
                }

                $raw_json = wp_json_encode( $event, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES );

                $wpdb->insert(
                    $this->tracking_events_table,
                    [
                        'order_id'        => $order_id,
                        'provider'        => $provider,
                        'tracking_number' => $tracking_number,
                        'event_code'      => $event_code ?: null,
                        'status_text'     => $status_text,
                        'location'        => $location ?: null,
                        'event_time'      => $event_time,
                        'meta'            => $meta_json,
                        'raw_payload'     => $raw_json,
                    ],
                    [ '%d', '%s', '%s', '%s', '%s', '%s', '%s', '%s' ]
                );

                $normalized_events[] = [
                    'event_code'  => $event_code,
                    'status_text' => $status_text,
                    'location'    => $location,
                    'event_time'  => $event_time,
                ];
            }

            $synced_at = current_time( 'mysql' );
            $wpdb->update(
                $this->orders_table,
                [ 'tracking_synced_at' => $synced_at ],
                [ 'id' => $order_id ],
                [ '%s' ],
                [ '%d' ]
            );

            do_action(
                'tanz_after_tracking_sync',
                $order_id,
                [
                    'provider'        => $provider,
                    'tracking_number' => $tracking_number,
                    'events'          => $normalized_events,
                    'synced_at'       => $synced_at,
                    'context'         => $context,
                ]
            );

            return [
                'events'    => $normalized_events,
                'synced_at' => $synced_at,
            ];
        }

        private function clear_tracking_events( int $order_id ): void {
            global $wpdb;

            $wpdb->delete( $this->tracking_events_table, [ 'order_id' => $order_id ], [ '%d' ] );
            $wpdb->query( $wpdb->prepare( "UPDATE {$this->orders_table} SET tracking_synced_at = NULL WHERE id = %d", $order_id ) );
        }

        private function maybe_sync_order_tracking( int $order_id, string $context = 'manual' ) {
            $row = $this->fetch_order_row( $order_id );

            if ( ! $row ) {
                return new \WP_Error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), [ 'status' => 404 ] );
            }

            if ( empty( $row['tracking_provider'] ) || empty( $row['tracking_number'] ) ) {
                return [];
            }

            return $this->sync_tracking_events( $order_id, $row['tracking_provider'], $row['tracking_number'], $context );
        }

        public function handle_save_tracking_settings(): void {
            if ( ! current_user_can( 'tanz_manage_shipping' ) ) {
                wp_die( __( '无权限执行此操作。', 'tanzanite-settings' ) );
            }

            check_admin_referer( 'tanz_tracking_settings' );

            $provider = sanitize_key( wp_unslash( $_POST['provider'] ?? '17track' ) );
            if ( ! isset( self::TRACKING_PROVIDERS[ $provider ] ) ) {
                $provider = '17track';
            }

            $all_settings = [];
            $raw_settings = $_POST['settings'] ?? [];
            if ( is_array( $raw_settings ) ) {
                foreach ( self::TRACKING_PROVIDERS as $key => $config ) {
                    $fields = $config['fields'] ?? [];
                    $values = $raw_settings[ $key ] ?? [];
                    $sanitized = [];

                    foreach ( $fields as $field_key => $field_config ) {
                        $value = $values[ $field_key ] ?? ( $field_config['default'] ?? '' );
                        $type  = $field_config['type'] ?? 'text';

                        switch ( $type ) {
                            case 'password':
                            case 'text':
                                $sanitized[ $field_key ] = sanitize_text_field( $value );
                                break;
                            default:
                                $sanitized[ $field_key ] = sanitize_text_field( $value );
                        }
                    }

                    $all_settings[ $key ] = $sanitized;
                }
            }

            update_option(
                self::OPTION_TRACKING_SETTINGS,
                [
                    'provider' => $provider,
                    'settings' => $all_settings,
                ],
                false
            );

            $this->tracking_provider_instances = []; // reset

            $this->log_audit(
                'update',
                'tracking_settings',
                0,
                [
                    'provider' => $provider,
                ],
                new \WP_REST_Request() // placeholder request
            );

            wp_safe_redirect( add_query_arg( [ 'page' => 'tanzanite-settings-tracking', 'updated' => '1' ], admin_url( 'admin.php' ) ) );
            exit;
        }

        /**
         * 当前用户是否可以查看后台管控数据。
         */
        public function can_view_dashboard_data( \WP_REST_Request $request ): bool {
            return current_user_can( 'tanz_view_products' ) || current_user_can( 'tanz_view_orders' );
        }

        /**
         * 当前用户是否可以写后台管控数据。
         */
        public function can_edit_dashboard_data( \WP_REST_Request $request ): bool {
            return current_user_can( 'tanz_manage_products' ) && $this->has_valid_write_context( $request );
        }

        private function verify_request_nonce( \WP_REST_Request $request ): bool {
            $nonce = $request->get_header( 'x-wp-nonce' );
            if ( $nonce && wp_verify_nonce( $nonce, 'wp_rest' ) ) {
                return true;
            }

            $nonce = $request->get_param( '_wpnonce' );

            return (bool) ( $nonce && wp_verify_nonce( $nonce, 'wp_rest' ) );
        }

        private function header_starts_with( string $haystack, string $prefix ): bool {
            return 0 === strpos( $haystack, $prefix );
        }

        private function respond_error( string $code, string $message = '', int $status = 400 ): \WP_REST_Response {
            $hint = self::ERROR_CODE_GUIDE[ $code ] ?? '';
            if ( '' === $message ) {
                $message = $hint ?: __( '请求失败，请稍后重试。', 'tanzanite-settings' );
            }

            $payload = [
                'code'    => $code,
                'message' => $message,
            ];

            if ( $hint && $hint !== $message ) {
                $payload['hint'] = $hint;
            }

            if ( isset( self::ERROR_ACTION_STEPS[ $code ] ) ) {
                $payload['actions'] = self::ERROR_ACTION_STEPS[ $code ];
            }

            return new \WP_REST_Response( $payload, $status );
        }

        private function error_to_response( \WP_Error $error, int $status = 400 ): \WP_REST_Response {
            return $this->respond_error( $error->get_error_code() ?: 'invalid_request', $error->get_error_message(), $status );
        }

        private function sanitize_skus( $skus, array $defaults = [] ) {
            if ( empty( $skus ) ) {
                return [];
            }

            if ( ! is_array( $skus ) ) {
                return new \WP_Error( 'invalid_sku_payload', __( 'SKUs must be an array.', 'tanzanite-settings' ) );
            }

            $defaults = wp_parse_args(
                $defaults,
                [
                    'price_regular' => 0.0,
                    'price_sale'    => 0.0,
                    'stock_qty'     => 0,
                ]
            );

            $sanitized = [];
            $seen_codes = [];

            foreach ( $skus as $index => $sku ) {
                if ( ! is_array( $sku ) ) {
                    return new \WP_Error( 'invalid_sku_entry', sprintf( __( 'SKU %d must be an object.', 'tanzanite-settings' ), $index + 1 ) );
                }

                $sku_code = isset( $sku['sku_code'] ) ? sanitize_text_field( $sku['sku_code'] ) : '';
                if ( '' === $sku_code ) {
                    return new \WP_Error( 'invalid_sku_code', sprintf( __( 'SKU %d requires a sku_code value.', 'tanzanite-settings' ), $index + 1 ) );
                }

                if ( isset( $seen_codes[ $sku_code ] ) ) {
                    return new \WP_Error( 'duplicate_sku_code', sprintf( __( 'Duplicate SKU code detected: %s.', 'tanzanite-settings' ), $sku_code ) );
                }

                $seen_codes[ $sku_code ] = true;

                $price_regular = isset( $sku['price_regular'] ) ? (float) $sku['price_regular'] : 0.0;
                if ( $price_regular <= 0 && $defaults['price_regular'] > 0 ) {
                    $price_regular = (float) $defaults['price_regular'];
                }

                $price_sale = isset( $sku['price_sale'] ) ? (float) $sku['price_sale'] : 0.0;
                if ( $price_sale <= 0 ) {
                    $price_sale = $defaults['price_sale'] > 0 ? (float) $defaults['price_sale'] : $price_regular;
                }

                $stock_qty = isset( $sku['stock_qty'] ) ? max( 0, (int) $sku['stock_qty'] ) : (int) $defaults['stock_qty'];
                $weight        = isset( $sku['weight'] ) && '' !== $sku['weight'] ? (float) $sku['weight'] : null;
                $barcode       = isset( $sku['barcode'] ) ? sanitize_text_field( $sku['barcode'] ) : '';

                $attributes = [];
                if ( isset( $sku['attributes'] ) && is_array( $sku['attributes'] ) ) {
                    foreach ( $sku['attributes'] as $attr_key => $attr_value ) {
                        $clean_key = sanitize_key( $attr_key );
                        if ( is_array( $attr_value ) ) {
                            $attributes[ $clean_key ] = array_map( 'sanitize_text_field', $attr_value );
                        } else {
                            $attributes[ $clean_key ] = sanitize_text_field( (string) $attr_value );
                        }
                    }
                }

                $tier_prices = [];
                if ( isset( $sku['tier_prices'] ) && is_array( $sku['tier_prices'] ) ) {
                    foreach ( $sku['tier_prices'] as $tier_index => $tier ) {
                        if ( ! is_array( $tier ) ) {
                            return new \WP_Error( 'invalid_tier_price', sprintf( __( 'Tier price %1$d in SKU %2$s must be an object.', 'tanzanite-settings' ), $tier_index + 1, $sku_code ) );
                        }

                        $min_qty = isset( $tier['min_qty'] ) ? (int) $tier['min_qty'] : 0;
                        $tier_price = isset( $tier['price'] ) ? (float) $tier['price'] : 0.0;

                        if ( $min_qty <= 0 ) {
                            return new \WP_Error( 'invalid_tier_qty', sprintf( __( 'Tier price %1$d in SKU %2$s must specify min_qty greater than zero.', 'tanzanite-settings' ), $tier_index + 1, $sku_code ) );
                        }

                        $tier_prices[] = [
                            'min_qty' => $min_qty,
                            'price'   => $tier_price,
                        ];
                    }
                }

                $sanitized[] = [
                    'sku_code'      => $sku_code,
                    'attributes'    => $attributes,
                    'price_regular' => $price_regular,
                    'price_sale'    => $price_sale,
                    'stock_qty'     => $stock_qty,
                    'tier_prices'   => $tier_prices,
                    'weight'        => $weight,
                    'barcode'       => $barcode,
                    'meta'          => isset( $sku['meta'] ) && is_array( $sku['meta'] ) ? $sku['meta'] : [],
                    'sort_order'    => isset( $sku['sort_order'] ) ? (int) $sku['sort_order'] : ( ( $index + 1 ) * 10 ),
                ];
            }

            usort(
                $sanitized,
                static function ( array $a, array $b ): int {
                    return $a['sort_order'] <=> $b['sort_order'];
                }
            );

            return $sanitized;
        }

        private function persist_product_skus( int $product_id, array $skus ): bool {
            global $wpdb;

            $wpdb->delete( $this->product_skus_table, [ 'product_id' => $product_id ], [ '%d' ] );

            foreach ( $skus as $sku ) {
                $result = $wpdb->insert(
                    $this->product_skus_table,
                    [
                        'product_id'    => $product_id,
                        'sku_code'      => $sku['sku_code'],
                        'attributes'    => wp_json_encode( $sku['attributes'] ),
                        'price_regular' => $sku['price_regular'],
                        'price_sale'    => $sku['price_sale'],
                        'stock_qty'     => $sku['stock_qty'],
                        'tier_prices'   => wp_json_encode( $sku['tier_prices'] ),
                        'sort_order'    => $sku['sort_order'],
                        'weight'        => null === $sku['weight'] ? null : (string) $sku['weight'],
                        'barcode'       => $sku['barcode'],
                        'meta'          => wp_json_encode( $sku['meta'] ),
                    ],
                    [ '%d', '%s', '%s', '%f', '%f', '%d', '%s', '%d', '%s', '%s', '%s' ]
                );

                if ( false === $result ) {
                    return false;
                }
            }

            return true;
        }

        private function delete_product_skus( int $product_id ): void {
            global $wpdb;

            $wpdb->delete( $this->product_skus_table, [ 'product_id' => $product_id ], [ '%d' ] );
        }

        private function get_product_skus( int $product_id ): array {
            global $wpdb;

            $rows = $wpdb->get_results( $wpdb->prepare( "SELECT * FROM {$this->product_skus_table} WHERE product_id = %d ORDER BY sort_order ASC, id ASC", $product_id ), ARRAY_A );

            if ( ! $rows ) {
                return [];
            }

            return array_map( [ $this, 'build_sku_response_row' ], $rows );
        }

        private function build_sku_response_row( array $row ): array {
            $attributes = $row['attributes'] ? json_decode( $row['attributes'], true ) : [];
            if ( ! is_array( $attributes ) ) {
                $attributes = [];
            }

            $tier_prices = $row['tier_prices'] ? json_decode( $row['tier_prices'], true ) : [];
            if ( ! is_array( $tier_prices ) ) {
                $tier_prices = [];
            }

            $meta = $row['meta'] ? json_decode( $row['meta'], true ) : [];
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            return [
                'id'            => (int) $row['id'],
                'sku_code'      => $row['sku_code'],
                'attributes'    => $attributes,
                'price_regular' => (float) $row['price_regular'],
                'price_sale'    => (float) $row['price_sale'],
                'stock_qty'     => (int) $row['stock_qty'],
                'tier_prices'   => $tier_prices,
                'sort_order'    => (int) $row['sort_order'],
                'weight'        => null !== $row['weight'] ? (float) $row['weight'] : null,
                'barcode'       => $row['barcode'],
                'meta'          => $meta,
            ];
        }

        private function get_product_sku_codes( int $product_id ): array {
            global $wpdb;

            $rows = $wpdb->get_col( $wpdb->prepare( "SELECT sku_code FROM {$this->product_skus_table} WHERE product_id = %d ORDER BY sku_code ASC", $product_id ) );

            if ( ! $rows ) {
                return [];
            }

            return array_map( 'strval', $rows );
        }

        private function handle_product_skus( int $product_id, \WP_REST_Request $request ) {
            $has_skus_param = $request->has_param( 'skus' );
            $bulk_payload   = $request->get_param( 'skus_bulk' );

            if ( ! $has_skus_param && null === $bulk_payload ) {
                return [ 'count' => $this->count_product_skus( $product_id ) ];
            }

            $skus = $has_skus_param ? (array) $request->get_param( 'skus' ) : [];

            if ( empty( $skus ) && null !== $bulk_payload && '' !== trim( (string) $bulk_payload ) ) {
                $parsed = $this->parse_bulk_skus( (string) $bulk_payload );
                if ( is_wp_error( $parsed ) ) {
                    return $parsed;
                }

                $skus = $parsed;
            }

            $defaults_meta = $this->get_product_meta_payload( $product_id );
            $sanitized = $this->sanitize_skus(
                $skus,
                [
                    'price_regular' => $defaults_meta['price_regular'] ?? 0,
                    'price_sale'    => $defaults_meta['price_sale'] ?? 0,
                    'stock_qty'     => $defaults_meta['stock_qty'] ?? 0,
                ]
            );

            if ( is_wp_error( $sanitized ) ) {
                return $sanitized;
            }

            if ( empty( $sanitized ) ) {
                $this->delete_product_skus( $product_id );

                return [ 'count' => 0 ];
            }

            if ( ! $this->persist_product_skus( $product_id, $sanitized ) ) {
                return new \WP_Error( 'failed_persist_skus', __( '保存 SKU 时发生异常，请稍后重试。', 'tanzanite-settings' ) );
            }

            return [ 'count' => count( $sanitized ) ];
        }

        private function parse_bulk_skus( string $bulk ) {
            $bulk = trim( $bulk );
            if ( '' === $bulk ) {
                return [];
            }

            $decoded = json_decode( $bulk, true );
            if ( is_array( $decoded ) ) {
                return $decoded;
            }

            $lines  = preg_split( '/\r?\n/', $bulk );
            $parsed = [];

            foreach ( $lines as $line ) {
                $line = trim( $line );
                if ( '' === $line ) {
                    continue;
                }

                $columns = str_getcsv( $line );
                if ( empty( $columns ) ) {
                    continue;
                }

                $sku_code = trim( (string) ( $columns[0] ?? '' ) );
                if ( '' === $sku_code ) {
                    return new \WP_Error( 'invalid_sku_code', __( '批量 SKU 中存在空的 sku_code。', 'tanzanite-settings' ) );
                }

                $entry = [
                    'sku_code'      => $sku_code,
                    'price_regular' => $columns[1] ?? null,
                    'price_sale'    => $columns[2] ?? null,
                    'stock_qty'     => $columns[3] ?? null,
                ];

                if ( isset( $columns[4] ) && '' !== trim( (string) $columns[4] ) ) {
                    $attr = json_decode( (string) $columns[4], true );
                    if ( is_array( $attr ) ) {
                        $entry['attributes'] = $attr;
                    }
                }

                $parsed[] = $entry;
            }

            if ( empty( $parsed ) ) {
                return new \WP_Error( 'invalid_sku_payload', __( '无法解析批量 SKU，请确认格式。', 'tanzanite-settings' ) );
            }

            return $parsed;
        }

        private function count_product_skus( int $product_id ): int {
            global $wpdb;

            return (int) $wpdb->get_var( $wpdb->prepare( "SELECT COUNT(*) FROM {$this->product_skus_table} WHERE product_id = %d", $product_id ) );
        }

        private function log_audit( string $action, string $target_type, int $target_id, array $payload, \WP_REST_Request $request ): void {
            global $wpdb;

            $user_id   = get_current_user_id();
            $user_name = '';
            if ( $user_id ) {
                $user = get_userdata( $user_id );
                if ( $user ) {
                    $user_name = $user->display_name ?: $user->user_login;
                }
            }

            $ip = $request->get_header( 'x-forwarded-for' );
            if ( $ip ) {
                $ip = trim( explode( ',', $ip )[0] );
            }

            if ( ! $ip ) {
                $ip = $_SERVER['REMOTE_ADDR'] ?? '';
            }

            $wpdb->insert(
                $this->audit_log_table,
                [
                    'actor_id'    => $user_id,
                    'actor_name'  => $user_name,
                    'action'      => $action,
                    'target_type' => $target_type,
                    'target_id'   => $target_id,
                    'payload'     => wp_json_encode( $payload ),
                    'ip_address'  => $ip,
                ],
                [ '%d', '%s', '%s', '%s', '%d', '%s', '%s' ]
            );
        }

        private function apply_status_timestamps( array &$data, array &$format_map, ?array $current_row, string $status ): void {
            $current_row = $current_row ?? [];
            $now         = current_time( 'mysql' );

            $assign_if_needed = function ( string $field ) use ( &$data, &$format_map, $current_row, $now ): void {
                if ( empty( $current_row[ $field ] ) && ! isset( $data[ $field ] ) ) {
                    $data[ $field ]        = $now;
                    $format_map[ $field ]  = '%s';
                }
            };

            if ( in_array( $status, [ 'paid', 'processing', 'shipped', 'completed' ], true ) ) {
                $assign_if_needed( 'paid_at' );
            }

            if ( in_array( $status, [ 'shipped', 'completed' ], true ) ) {
                $assign_if_needed( 'shipped_at' );
            }

            if ( 'completed' === $status ) {
                $assign_if_needed( 'completed_at' );
            }

            if ( 'cancelled' === $status ) {
                $assign_if_needed( 'cancelled_at' );
            }
        }

        private function sanitize_order_items( $items ) {
            if ( empty( $items ) || ! is_array( $items ) ) {
                return new \WP_Error( 'invalid_order_items', __( 'Order items must be a non-empty array.', 'tanzanite-settings' ) );
            }

            $sanitized = [];

            foreach ( $items as $index => $item ) {
                if ( ! is_array( $item ) ) {
                    return new \WP_Error( 'invalid_order_item', sprintf( __( 'Order item %d must be an object.', 'tanzanite-settings' ), $index + 1 ) );
                }

                $quantity = isset( $item['quantity'] ) ? (int) $item['quantity'] : 0;
                if ( $quantity <= 0 ) {
                    return new \WP_Error( 'invalid_order_item_quantity', sprintf( __( 'Order item %d must have quantity greater than zero.', 'tanzanite-settings' ), $index + 1 ) );
                }

                $price = isset( $item['price'] ) ? (float) $item['price'] : 0.0;
                $total = isset( $item['total'] ) ? (float) $item['total'] : $price * $quantity;

                $product_id = isset( $item['product_id'] ) ? (int) $item['product_id'] : 0;
                $sku_id     = isset( $item['sku_id'] ) ? (int) $item['sku_id'] : 0;
                $sku_code   = isset( $item['sku_code'] ) ? sanitize_text_field( $item['sku_code'] ) : '';

                $product_title = isset( $item['product_title'] ) ? sanitize_text_field( $item['product_title'] ) : '';
                if ( $product_id > 0 && '' === $product_title ) {
                    $product_title = get_the_title( $product_id );
                }

                if ( '' === $product_title ) {
                    return new \WP_Error( 'invalid_order_item_title', sprintf( __( 'Order item %d requires a product title.', 'tanzanite-settings' ), $index + 1 ) );
                }

                $meta = [];
                if ( isset( $item['meta'] ) && is_array( $item['meta'] ) ) {
                    foreach ( $item['meta'] as $meta_key => $meta_value ) {
                        $meta[ sanitize_key( $meta_key ) ] = is_scalar( $meta_value ) ? sanitize_text_field( (string) $meta_value ) : $meta_value;
                    }
                }

                $sanitized[] = [
                    'product_id'    => $product_id,
                    'sku_id'        => $sku_id,
                    'product_title' => $product_title,
                    'sku_code'      => $sku_code,
                    'quantity'      => $quantity,
                    'price'         => $price,
                    'total'         => $total,
                    'meta'          => $meta,
                ];
            }

            return $sanitized;
        }

        private function persist_order_items( int $order_id, array $items ): bool {
            global $wpdb;

            $wpdb->delete( $this->order_items_table, [ 'order_id' => $order_id ], [ '%d' ] );

            foreach ( $items as $item ) {
                $result = $wpdb->insert(
                    $this->order_items_table,
                    [
                        'order_id'      => $order_id,
                        'product_id'    => $item['product_id'],
                        'sku_id'        => $item['sku_id'],
                        'product_title' => $item['product_title'],
                        'sku_code'      => $item['sku_code'],
                        'quantity'      => $item['quantity'],
                        'price'         => $item['price'],
                        'total'         => $item['total'],
                        'meta'          => wp_json_encode( $item['meta'] ),
                    ],
                    [ '%d', '%d', '%d', '%s', '%s', '%d', '%f', '%f', '%s' ]
                );

                if ( false === $result ) {
                    return false;
                }
            }

            return true;
        }

        private function delete_order_items( int $order_id ): void {
            global $wpdb;

            $wpdb->delete( $this->order_items_table, [ 'order_id' => $order_id ], [ '%d' ] );
        }

        private function fetch_order_items( int $order_id ): array {
            global $wpdb;

            $rows = $wpdb->get_results( $wpdb->prepare( "SELECT * FROM {$this->order_items_table} WHERE order_id = %d ORDER BY id ASC", $order_id ), ARRAY_A );
            if ( ! $rows ) {
                return [];
            }

            return array_map( [ $this, 'format_order_item_row' ], $rows );
        }

        public function render_sku_importer(): void {
            $nonce      = wp_create_nonce( 'wp_rest' );
            $product_id = isset( $_GET['product_id'] ) ? (int) $_GET['product_id'] : 0; // phpcs:ignore WordPress.Security.NonceVerification.Recommended

            echo '<div class="tz-settings-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'SKU Importer', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '支持批量导入/覆盖指定商品的 SKU。您可以粘贴 CSV 行或 JSON 数组进行预检，再决定是否正式写入。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '步骤 1：选择商品', 'tanzanite-settings' ) . '</div>';
            echo '      <label>' . esc_html__( '商品 ID', 'tanzanite-settings' ) . '</label>';
            echo '      <input type="number" id="tz-sku-import-product" class="regular-text" min="1" value="' . esc_attr( $product_id ) . '" />';
            echo '      <p class="description">' . esc_html__( '可在商品列表中查看对应的 Product ID。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '步骤 2：粘贴数据', 'tanzanite-settings' ) . '</div>';
            echo '      <p>' . esc_html__( '支持以下两种格式：', 'tanzanite-settings' ) . '</p>';
            echo '      <ul class="tz-section-list">';
            echo '          <li>' . esc_html__( 'JSON 数组：[{"sku_code":"SKU-001","price_regular":199}]', 'tanzanite-settings' ) . '</li>';
            echo '          <li>' . esc_html__( 'CSV 行：sku_code,price_regular,price_sale,stock_qty,attributes(JSON)', 'tanzanite-settings' ) . '</li>';
            echo '      </ul>';
            echo '      <textarea id="tz-sku-import-text" rows="10" style="width:100%;"></textarea>';
            echo '      <div style="display:flex;gap:12px;flex-wrap:wrap;margin-top:12px;">';
            echo '          <button class="button" id="tz-sku-import-download-template">' . esc_html__( '下载 CSV 模板', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button button-primary" id="tz-sku-import-preview">' . esc_html__( '预检数据', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button button-secondary" id="tz-sku-import-apply" disabled>' . esc_html__( '确认导入', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section" id="tz-sku-import-preview-panel" style="display:none;">';
            echo '      <div class="tz-section-title">' . esc_html__( '预检结果', 'tanzanite-settings' ) . '</div>';
            echo '      <div id="tz-sku-import-summary"></div>';
            echo '      <div style="overflow:auto;">';
            echo '          <table class="widefat fixed striped" id="tz-sku-import-table" style="margin-top:16px;">';
            echo '              <thead><tr>';
            foreach ( [ 'sku_code', 'price_regular', 'price_sale', 'stock_qty', 'tier_prices', 'sort_order' ] as $column ) {
                echo '<th>' . esc_html( ucfirst( str_replace( '_', ' ', $column ) ) ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '  </div>';

            echo '</div>';

            $rest_url_preview = esc_url_raw( rest_url( 'tanzanite/v1/sku-importer/preview' ) );
            $rest_url_apply   = esc_url_raw( rest_url( 'tanzanite/v1/sku-importer/apply' ) );

            // 加载 SKU Importer JS
            wp_enqueue_script(
                'tz-sku-importer',
                plugins_url( 'assets/js/sku-importer.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-sku-importer',
                'TzSkuImporterConfig',
                array(
                    'nonce'      => $nonce,
                    'previewUrl' => $rest_url_preview,
                    'applyUrl'   => $rest_url_apply,
                    'strings'    => array(
                        'errorNoProductId'    => __( '请先输入商品 ID。', 'tanzanite-settings' ),
                        'errorPreviewFailed'  => __( '预检失败。', 'tanzanite-settings' ),
                        'errorNetworkRetry'   => __( '请检查网络后重试。', 'tanzanite-settings' ),
                        'errorImportFailed'   => __( '导入失败。', 'tanzanite-settings' ),
                        'errorRetryLater'     => __( '请稍后重试。', 'tanzanite-settings' ),
                        'successImported'     => __( '导入成功，SKU 已更新。', 'tanzanite-settings' ),
                    ),
                )
            );
        }

        public function render_audit_logs(): void {
            $nonce        = wp_create_nonce( 'wp_rest' );
            $actions      = $this->get_audit_distinct_values( 'action', 50 );
            $target_types = $this->get_audit_distinct_values( 'target_type', 50 );
            $actors       = $this->get_audit_distinct_values( 'actor_name', 50 );

            echo '<div class="tz-settings-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Audit Logs', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '查看后台关键操作记录，可按动作、目标、操作人及时间范围筛选，支持导出 CSV。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-audit-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
            echo '      <div class="tz-filter-grid" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:16px;">';

            echo '          <label>' . esc_html__( 'Action', 'tanzanite-settings' ) . '<select id="tz-audit-action"><option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
            foreach ( $actions as $action ) {
                echo '<option value="' . esc_attr( $action ) . '">' . esc_html( $action ) . '</option>';
            }
            echo '</select></label>';

            echo '          <label>' . esc_html__( 'Target Type', 'tanzanite-settings' ) . '<select id="tz-audit-target"><option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
            foreach ( $target_types as $type ) {
                echo '<option value="' . esc_attr( $type ) . '">' . esc_html( $type ) . '</option>';
            }
            echo '</select></label>';

            echo '          <label>' . esc_html__( 'Actor', 'tanzanite-settings' ) . '<select id="tz-audit-actor"><option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
            foreach ( $actors as $actor ) {
                echo '<option value="' . esc_attr( $actor ) . '">' . esc_html( $actor ) . '</option>';
            }
            echo '</select></label>';

            echo '          <label>' . esc_html__( '关键字', 'tanzanite-settings' ) . '<input type="text" id="tz-audit-search" placeholder="' . esc_attr__( '动作/目标/备注', 'tanzanite-settings' ) . '" /></label>';
            echo '          <label>' . esc_html__( '开始日期', 'tanzanite-settings' ) . '<input type="date" id="tz-audit-start" /></label>';
            echo '          <label>' . esc_html__( '结束日期', 'tanzanite-settings' ) . '<input type="date" id="tz-audit-end" /></label>';
            echo '          <label>' . esc_html__( '每页数量', 'tanzanite-settings' ) . '<select id="tz-audit-per-page"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select></label>';

            echo '      </div>';
            echo '      <div style="margin-top:16px;display:flex;gap:12px;flex-wrap:wrap;">';
            echo '          <button class="button button-primary" id="tz-audit-apply">' . esc_html__( '应用筛选', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button" id="tz-audit-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button button-secondary" id="tz-audit-export">' . esc_html__( '导出 CSV', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '日志列表', 'tanzanite-settings' ) . '</div>';
            echo '      <div id="tz-audit-summary"></div>';
            echo '      <div style="overflow:auto;">';
            echo '          <table class="widefat fixed striped" id="tz-audit-table" style="margin-top:16px;min-width:920px;">';
            echo '              <thead><tr>';
            foreach ( [ 'ID', '时间', '动作', '目标', '操作人', 'IP', '详情' ] as $column ) {
                echo '<th>' . esc_html( $column ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '      <div id="tz-audit-pagination" style="margin-top:16px;display:flex;justify-content:space-between;align-items:center;">';
            echo '          <div id="tz-audit-page-info"></div>';
            echo '          <div style="display:flex;gap:8px;">';
            echo '              <button class="button" id="tz-audit-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button" id="tz-audit-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '      </div>';
            echo '  </div>';

            echo '</div>';

            $rest_url_logs = esc_url_raw( rest_url( 'tanzanite/v1/audit-logs' ) );

            // 加载 Audit Logs JS
            wp_enqueue_script(
                'tz-audit-logs',
                plugins_url( 'assets/js/audit-logs.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-audit-logs',
                'TzAuditLogsConfig',
                array(
                    'nonce'   => $nonce,
                    'restUrl' => $rest_url_logs,
                    'strings' => array(
                        'noRecords'       => __( '暂无记录', 'tanzanite-settings' ),
                        'summaryTemplate' => __( '共 {total} 条记录，当前第 {page}/{pages} 页。', 'tanzanite-settings' ),
                        'pageTemplate'    => __( '第 {page}/{pages} 页', 'tanzanite-settings' ),
                        'loadFailed'      => __( '加载失败。', 'tanzanite-settings' ),
                        'exportStarted'   => __( '已开始导出，CSV 将在新标签页中打开。', 'tanzanite-settings' ),
                    ),
                )
            );
        }

        // get_audit_distinct_values - 已迁移到 Tanzanite_REST_Audit_Controller

        // rest_list_audit_logs - 已迁移到 Tanzanite_REST_Audit_Controller::get_items()
        // rest_export_audit_logs - 已迁移到 Tanzanite_REST_Audit_Controller::export_items()

        private function format_order_item_row( array $row ): array {
            $meta = $row['meta'] ? json_decode( $row['meta'], true ) : [];
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            return [
                'id'             => (int) $row['id'],
                'product_id'     => (int) $row['product_id'],
                'sku_id'         => (int) $row['sku_id'],
                'product_title'  => $row['product_title'],
                'sku_code'       => $row['sku_code'],
                'quantity'       => (int) $row['quantity'],
                'price'          => (float) $row['price'],
                'total'          => (float) $row['total'],
                'meta'           => $meta,
            ];
        }

        private function normalize_product_status( ?string $status ): string {
            $status = $status ? sanitize_key( $status ) : 'draft';
            if ( ! in_array( $status, self::ALLOWED_PRODUCT_STATUSES, true ) ) {
                return 'draft';
            }

            return $status;
        }

        private function ensure_order_status( ?string $status ): string {
            $status = $status ? sanitize_key( $status ) : 'pending';
            if ( ! in_array( $status, self::ALLOWED_ORDER_STATUSES, true ) ) {
                return 'pending';
            }

            return $status;
        }

        private function normalize_review_status( ?string $status ): string {
            $status = $status ? sanitize_key( $status ) : 'pending';

            return in_array( $status, self::ALLOWED_REVIEW_STATUSES, true ) ? $status : 'pending';
        }

        private function get_product_meta_payload( int $post_id ): array {
            $payload = [];

            foreach ( self::PRODUCT_META_MAP as $field => $config ) {
                $value = get_post_meta( $post_id, $config['key'], true );
                if ( '' === $value ) {
                    $value = $config['default'];
                }

                $payload[ $field ] = $this->cast_meta_value( $value, $config, false );
            }

            // 添加 URLLink 自定义路径
            $payload['urllink_custom_path'] = get_post_meta( $post_id, '_urllink_custom_path', true ) ?: '';

            return $payload;
        }

        private function update_product_meta_from_request( int $post_id, \WP_REST_Request $request ): void {
            foreach ( self::PRODUCT_META_MAP as $field => $config ) {
                if ( ! $request->has_param( $field ) ) {
                    continue;
                }

                $value = $this->cast_meta_value( $request->get_param( $field ), $config, true );
                update_post_meta( $post_id, $config['key'], $value );
            }
        }

        private function update_product_meta_fields( int $post_id, array $fields ): array {
            $updated = [];

            foreach ( self::PRODUCT_META_MAP as $field => $config ) {
                if ( ! array_key_exists( $field, $fields ) ) {
                    continue;
                }

                $value = $this->cast_meta_value( $fields[ $field ], $config, true );
                update_post_meta( $post_id, $config['key'], $value );
                $updated[ $field ] = $value;
            }

            return $updated;
        }

        private function adjust_product_stock( int $post_id, int $delta ): int {
            $config  = self::PRODUCT_META_MAP['stock_qty'];
            $current = (int) get_post_meta( $post_id, $config['key'], true );
            $new     = max( 0, $current + $delta );

            update_post_meta( $post_id, $config['key'], $new );

            return $new;
        }

        private function sanitize_id_list( $ids ): array {
            if ( ! is_array( $ids ) ) {
                return [];
            }

            $ids = array_map( 'absint', $ids );
            $ids = array_filter( $ids );

            return array_values( array_unique( $ids ) );
        }

        private function generate_csv( array $headers, array $rows ): string {
            $handle = fopen( 'php://temp', 'r+' );

            if ( ! $handle ) {
                return '';
            }

            fputcsv( $handle, $headers );

            foreach ( $rows as $row ) {
                $ordered = [];
                foreach ( $headers as $header ) {
                    $ordered[] = isset( $row[ $header ] ) ? $row[ $header ] : '';
                }

                fputcsv( $handle, $ordered );
            }

            rewind( $handle );
            $csv = stream_get_contents( $handle );
            fclose( $handle );

            return false === $csv ? '' : $csv;
        }

        private function format_review_row( array $row ): array {
            return [
                'id'          => (int) $row['id'],
                'product_id'  => (int) $row['product_id'],
                'user_id'     => (int) $row['user_id'],
                'author_name' => $row['author_name'],
                'author_email'=> $row['author_email'],
                'author_phone'=> $row['author_phone'],
                'rating'      => (int) $row['rating'],
                'content'     => $row['content'],
                'images'      => $row['images'] ? json_decode( $row['images'], true ) : [],
                'status'      => $row['status'],
                'is_featured' => (bool) $row['is_featured'],
                'reply'       => [
                    'text'   => $row['reply_text'],
                    'author' => $row['reply_author'],
                    'time'   => $row['reply_at'],
                ],
                'meta'        => $row['meta'] ? json_decode( $row['meta'], true ) : [],
                'created_at'  => $row['created_at'],
                'updated_at'  => $row['updated_at'],
            ];
        }

        private function fetch_review_row( int $id ): ?array {
            global $wpdb;

            $row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->product_reviews_table} WHERE id = %d", $id ), ARRAY_A );

            return $row ? $this->format_review_row( $row ) : null;
        }

        private function fetch_reviews( array $args = [] ): array {
            global $wpdb;

            $defaults = [
                'status'     => null,
                'product_id' => null,
                'search'     => null,
                'limit'      => 50,
                'offset'     => 0,
            ];

            $args   = wp_parse_args( $args, $defaults );
            $where  = [];
            $params = [];

            if ( $args['status'] ) {
                $where[]  = 'status = %s';
                $params[] = $this->normalize_review_status( $args['status'] );
            }

            if ( $args['product_id'] ) {
                $where[]  = 'product_id = %d';
                $params[] = (int) $args['product_id'];
            }

            if ( $args['search'] ) {
                $like     = '%' . $wpdb->esc_like( $args['search'] ) . '%';
                $where[]  = '(author_name LIKE %s OR author_email LIKE %s OR content LIKE %s)';
                $params[] = $like;
                $params[] = $like;
                $params[] = $like;
            }

            $where_sql = $where ? 'WHERE ' . implode( ' AND ', $where ) : '';

            $query = "SELECT * FROM {$this->product_reviews_table} {$where_sql} ORDER BY created_at DESC LIMIT %d OFFSET %d";
            $params[] = (int) $args['limit'];
            $params[] = (int) $args['offset'];

            $rows = $wpdb->get_results( $wpdb->prepare( $query, $params ), ARRAY_A );

            $count_query = "SELECT COUNT(*) FROM {$this->product_reviews_table} {$where_sql}";
            $total       = (int) $wpdb->get_var( $wpdb->prepare( $count_query, array_slice( $params, 0, count( $params ) - 2 ) ) );

            return [
                'items' => array_map( [ $this, 'format_review_row' ], $rows ),
                'total' => $total,
            ];
        }

        // Reviews REST API 回调函数 - 已迁移到 Tanzanite_REST_Reviews_Controller
        // rest_create_review -> Tanzanite_REST_Reviews_Controller::create_item()
        // rest_list_reviews -> Tanzanite_REST_Reviews_Controller::get_items()
        // rest_get_review -> Tanzanite_REST_Reviews_Controller::get_item()
        // rest_update_review -> Tanzanite_REST_Reviews_Controller::update_item()
        // rest_delete_review -> Tanzanite_REST_Reviews_Controller::delete_item()

        private function cast_meta_value( $value, array $config, bool $from_request = false ) {
            $type = $config['type'] ?? 'string';

            switch ( $type ) {
                case 'integer':
                    return (int) ( is_scalar( $value ) ? $value : 0 );

                case 'number':
                    return (float) ( is_scalar( $value ) ? $value : 0 );

                case 'boolean':
                    return $this->normalize_boolean_meta_value( $value );

                case 'array':
                    $normalized = $this->normalize_array_meta_value( $value, $config, $from_request );

                    if ( isset( $config['sanitize_callback'] ) && is_callable( $config['sanitize_callback'] ) ) {
                        $sanitized = call_user_func( $config['sanitize_callback'], $normalized, $from_request );

                        if ( $sanitized instanceof \WP_Error ) {
                            if ( $from_request ) {
                                throw new \RuntimeException( $sanitized->get_error_message() );
                            }

                            return [];
                        }

                        return $sanitized;
                    }

                    return $normalized;

                case 'string':
                default:
                    $string_value = is_scalar( $value ) ? (string) $value : '';
                    $string_value = trim( $string_value );

                    return $string_value;
            }
        }

        private function normalize_boolean_meta_value( $value ): bool {
            if ( is_bool( $value ) ) {
                return $value;
            }

            if ( is_numeric( $value ) ) {
                return (bool) (int) $value;
            }

            if ( is_string( $value ) ) {
                $value = strtolower( trim( $value ) );

                return in_array( $value, [ '1', 'true', 'yes', 'on' ], true );
            }

            return (bool) $value;
        }

        private function normalize_array_meta_value( $value, array $config, bool $from_request ): array {
            if ( is_string( $value ) ) {
                $decoded = json_decode( $value, true );

                if ( JSON_ERROR_NONE === json_last_error() && is_array( $decoded ) ) {
                    $value = $decoded;
                } elseif ( $from_request ) {
                    $value = preg_split( '/[\r\n,]+/', $value, -1, PREG_SPLIT_NO_EMPTY );
                }
            } elseif ( $value instanceof \Traversable ) {
                $value = iterator_to_array( $value );
            } elseif ( null === $value ) {
                $value = [];
            }

            if ( ! is_array( $value ) ) {
                $value = [];
            }

            $item_type       = $config['item_type'] ?? 'mixed';
            $item_sanitize   = $config['item_sanitize_callback'] ?? null;
            $is_mixed        = 'mixed' === $item_type;
            $normalized_list = [];

            foreach ( $value as $key => $item ) {
                $processed = $item;

                switch ( $item_type ) {
                    case 'int':
                    case 'integer':
                        $processed = (int) ( is_scalar( $item ) ? $item : 0 );
                        break;

                    case 'number':
                    case 'float':
                        $processed = (float) ( is_scalar( $item ) ? $item : 0 );
                        break;

                    case 'bool':
                    case 'boolean':
                        $processed = $this->normalize_boolean_meta_value( $item );
                        break;

                    case 'string':
                        $processed = is_scalar( $item ) ? trim( (string) $item ) : '';
                        if ( '' === $processed ) {
                            continue 2;
                        }
                        break;

                    case 'mixed':
                    default:
                        // 保留原始结构，但确保基础类型在请求阶段有序整理。
                        if ( is_object( $processed ) ) {
                            $processed = json_decode( wp_json_encode( $processed ), true );
                        }
                        break;
                }

                if ( $item_sanitize && is_callable( $item_sanitize ) && ( ! is_array( $processed ) || ! $is_mixed ) ) {
                    $processed = call_user_func( $item_sanitize, $processed );
                }

                if ( $is_mixed ) {
                    $normalized_list[ $key ] = $processed;
                } else {
                    $normalized_list[] = $processed;
                }
            }

            if ( ! $is_mixed && in_array( $item_type, [ 'int', 'integer', 'float', 'number', 'string' ], true ) ) {
                $normalized_list = array_values( array_unique( $normalized_list, SORT_REGULAR ) );
            }

            return $normalized_list;
        }

        private function build_product_response( int $post_id ): array {
            $post = get_post( $post_id );

            if ( ! $post || 'tanz_product' !== $post->post_type ) {
                return [];
            }

            $meta = $this->get_product_meta_payload( $post_id );
            $skus = $this->get_product_skus( $post_id );

            $markdown_raw = get_post_meta( $post_id, self::PRODUCT_META_MAP['content_markdown']['key'] ?? '_tanz_content_markdown', true );

            $seo_payload = $this->get_product_seo_payload( $post_id );

            $response = [
                'id'         => $post->ID,
                'title'      => get_the_title( $post ),
                'slug'       => $post->post_name,
                'status'     => $post->post_status,
                'excerpt'    => $post->post_excerpt,
                'content'    => [
                    'html'      => $post->post_content,
                    'markdown'  => is_string( $markdown_raw ) ? $markdown_raw : '',
                ],
                'dates'      => [
                    'created_gmt' => $post->post_date_gmt,
                    'modified_gmt'=> $post->post_modified_gmt,
                    'local'       => $post->post_date,
                ],
                'meta'             => $meta,
                'skus'             => $skus,
                'seo'              => $seo_payload,
                'preview'          => $this->build_product_preview_payload( $post_id, $meta, $skus ),
                'can_delete'       => current_user_can( 'delete_post', $post_id ),
                'permalink'        => get_permalink( $post_id ),
            ];

            return $response;
        }

        private function build_product_preview_payload( int $post_id, array $meta, array $skus ): array {
            $post = get_post( $post_id );

            if ( ! $post ) {
                return [];
            }

            $preview = [
                'id'       => $post_id,
                'title'    => get_the_title( $post ),
                'slug'     => $post->post_name,
                'status'   => $post->post_status,
                'excerpt'  => $post->post_excerpt,
                'prices'   => [
                    'regular' => (float) ( $meta['price_regular'] ?? 0 ),
                    'sale'    => (float) ( $meta['price_sale'] ?? 0 ),
                    'member'  => (float) ( $meta['price_member'] ?? 0 ),
                ],
                'stock'    => [
                    'quantity'           => (int) ( $meta['stock_qty'] ?? 0 ),
                    'alert'              => (int) ( $meta['stock_alert'] ?? 0 ),
                    'backorders_allowed' => (bool) ( $meta['backorders_allowed'] ?? false ),
                ],
                'points'   => [
                    'reward' => (int) ( $meta['points_reward'] ?? 0 ),
                    'limit'  => (int) ( $meta['points_limit'] ?? 0 ),
                ],
                'membership' => [
                    'levels' => array_values( $meta['membership_levels'] ?? [] ),
                ],
                'logistics'  => [
                    'template_id'   =>(int) ( $meta['shipping_template_id'] ?? 0 ),
                    'free_shipping' => (bool) ( $meta['free_shipping'] ?? false ),
                    'shipping_time' => (string) ( $meta['shipping_time'] ?? '' ),
                    'tags'          => array_values( $meta['logistics_tags'] ?? [] ),
                ],
                'media' => [
                    'featured' => [
                        'id'  => (int) ( $meta['featured_image_id'] ?? 0 ),
                        'url' => (string) ( $meta['featured_image_url'] ?? '' ),
                    ],
                    'gallery_media_ids'   => array_map( 'intval', $meta['gallery_media_ids'] ?? [] ),
                    'gallery_external_urls'=> array_map( 'esc_url_raw', $meta['gallery_external_urls'] ?? [] ),
                    'video' => [
                        'id'  => (int) ( $meta['featured_video_id'] ?? 0 ),
                        'url' => (string) ( $meta['featured_video_url'] ?? '' ),
                    ],
                ],
                'channels' => array_values( $meta['channels'] ?? [] ),
                'featured' => [
                    'enabled' => (bool) ( $meta['featured_flag'] ?? false ),
                    'slot'    => (string) ( $meta['featured_slot'] ?? '' ),
                ],
                'sticky'   => (bool) ( $meta['is_sticky'] ?? false ),
                'skus'     => $skus,
            ];

            return $preview;
        }

        private function get_product_seo_payload( int $post_id ): array {
            if ( class_exists( '\\MyTheme_SEO_Plugin' ) ) {
                $plugin = \MyTheme_SEO_Plugin::instance();

                $seo_request = new \WP_REST_Request( 'GET', '/mytheme/v1/seo/product/' . $post_id );
                $seo_request->set_param( 'id', $post_id );
                $response = $plugin->rest_get_product_seo( $seo_request );

                $payload = [];
                if ( $response instanceof \WP_REST_Response ) {
                    $data = $response->get_data();
                    if ( is_array( $data ) && isset( $data['payload'] ) && is_array( $data['payload'] ) ) {
                        $payload = $data['payload'];
                    }
                }

                $languages = [];
                if ( method_exists( $plugin, 'rest_get_languages' ) ) {
                    $languages_response = $plugin->rest_get_languages( new \WP_REST_Request( 'GET', '/mytheme/v1/seo/languages' ) );
                    if ( $languages_response instanceof \WP_REST_Response ) {
                        $lang_data = $languages_response->get_data();
                        if ( is_array( $lang_data ) && isset( $lang_data['languages'] ) && is_array( $lang_data['languages'] ) ) {
                            $languages = $lang_data['languages'];
                        }
                    }
                }

                return [
                    'available'  => true,
                    'configured' => ! empty( $payload ),
                    'payload'    => $payload,
                    'languages'  => $languages,
                ];
            }

            $stored = get_post_meta( $post_id, '_mytheme_seo_payload', true );
            $sanitized = $this->sanitize_fallback_seo_payload( $stored );

            return [
                'available'  => false,
                'configured' => ! empty( $sanitized ),
                'payload'    => $sanitized,
                'languages'  => [],
            ];
        }

        private function sync_product_seo( int $post_id, $payload ): array {
            if ( class_exists( '\\MyTheme_SEO_Plugin' ) ) {
                $plugin = \MyTheme_SEO_Plugin::instance();

                if ( is_array( $payload ) ) {
                    $update_request = new \WP_REST_Request( 'POST', '/mytheme/v1/seo/product/' . $post_id );
                    $update_request->set_param( 'id', $post_id );
                    $update_request->set_param( 'payload', $payload );
                    $plugin->rest_update_product_seo( $update_request );
                }

                return $this->get_product_seo_payload( $post_id );
            }

            if ( is_array( $payload ) ) {
                $sanitized = $this->sanitize_fallback_seo_payload( $payload );
                update_post_meta( $post_id, '_mytheme_seo_payload', $sanitized );
            }

            return $this->get_product_seo_payload( $post_id );
        }

        private function sanitize_fallback_seo_payload( $payload ): array {
            if ( ! is_array( $payload ) ) {
                return [];
            }

            $sanitized = [];
            foreach ( $payload as $locale => $data ) {
                $locale_key = sanitize_key( (string) $locale );
                if ( '' === $locale_key || ! is_array( $data ) ) {
                    continue;
                }

                $sanitized[ $locale_key ] = [
                    'title'       => isset( $data['title'] ) ? sanitize_text_field( (string) $data['title'] ) : '',
                    'description' => isset( $data['description'] ) ? sanitize_textarea_field( (string) $data['description'] ) : '',
                ];
            }

            return $sanitized;
        }

        private function prepare_content_for_save( \WP_REST_Request $request ): array {
            $content_param = $request->get_param( 'content' );

            $html      = '';
            $markdown  = '';

            if ( is_array( $content_param ) ) {
                if ( isset( $content_param['html'] ) ) {
                    $html = wp_kses_post( (string) $content_param['html'] );
                }
                if ( isset( $content_param['markdown'] ) ) {
                    $markdown = (string) $content_param['markdown'];
                }
            }

            if ( '' === $html && $request->has_param( 'content_html' ) ) {
                $html = wp_kses_post( (string) $request->get_param( 'content_html' ) );
            }

            if ( '' === $markdown && $request->has_param( 'content_markdown' ) ) {
                $markdown = (string) $request->get_param( 'content_markdown' );
            }

            return [ $html, $markdown ];
        }

        private function prepare_post_dates_from_request( $date_param ): array {
            if ( empty( $date_param ) ) {
                return [];
            }

            if ( is_numeric( $date_param ) ) {
                $timestamp = (int) $date_param;
            } else {
                $timestamp = strtotime( (string) $date_param );
            }

            if ( false === $timestamp ) {
                return [];
            }

            $gmt = gmdate( 'Y-m-d H:i:s', $timestamp );

            return [
                'post_date_gmt' => $gmt,
                'post_date'     => get_date_from_gmt( $gmt ),
            ];
        }

        private function build_product_list_item( \WP_Post $post ): array {
            $meta = $this->get_product_meta_payload( $post->ID );

            $featured_id = (int) ( $meta['featured_image_id'] ?? 0 );
            $featured_url = $meta['featured_image_url'] ?? '';
            $thumbnail = $featured_id > 0 ? wp_get_attachment_image_url( $featured_id, 'thumbnail' ) : $featured_url;

            $prices = [
                'regular' => (float) ( $meta['price_regular'] ?? 0 ),
                'sale'    => (float) ( $meta['price_sale'] ?? 0 ),
                'member'  => (float) ( $meta['price_member'] ?? 0 ),
            ];

            $stock = [
                'quantity' => (int) ( $meta['stock_qty'] ?? 0 ),
                'alert'    => (int) ( $meta['stock_alert'] ?? 0 ),
            ];

            $points = [
                'reward' => (int) ( $meta['points_reward'] ?? 0 ),
                'limit'  => (int) ( $meta['points_limit'] ?? 0 ),
            ];

            $terms = get_the_terms( $post->ID, 'category' );
            $categories = [];
            if ( is_array( $terms ) ) {
                foreach ( $terms as $term ) {
                    $categories[] = [
                        'id'   => $term->term_id,
                        'name' => $term->name,
                        'slug' => $term->slug,
                    ];
                }
            }

            return [
                'id'         => $post->ID,
                'title'      => get_the_title( $post ),
                'status'     => $post->post_status,
                'excerpt'    => $post->post_excerpt,
                'thumbnail'  => $thumbnail,
                'prices'     => $prices,
                'stock'      => $stock,
                'points'     => $points,
                'sku_count'  => $this->count_product_skus( $post->ID ),
                'categories' => $categories,
                'updated_at' => $post->post_modified_gmt,
                'created_at' => $post->post_date_gmt,
                'slug'       => $post->post_name,
                'sticky'     => (bool) get_post_meta( $post->ID, self::PRODUCT_META_MAP['is_sticky']['key'], true ),
                'preview_url'=> get_permalink( $post ),
            ];
        }

        private function get_product_statistics(): array {
            global $wpdb;

            $status_counts = wp_count_posts( 'tanz_product' );
            $status_array  = (array) $status_counts;
            $total         = array_sum( array_map( 'intval', $status_array ) );

            $low_stock_threshold = 5;
            $stock_alert_key     = self::PRODUCT_META_MAP['stock_alert']['key'];
            $stock_qty_key       = self::PRODUCT_META_MAP['stock_qty']['key'];

            $low_stock = (int) $wpdb->get_var( $wpdb->prepare(
                "SELECT COUNT(*) FROM {$wpdb->postmeta} qty
                INNER JOIN {$wpdb->posts} p ON p.ID = qty.post_id
                LEFT JOIN {$wpdb->postmeta} alert ON alert.post_id = p.ID AND alert.meta_key = %s
                WHERE qty.meta_key = %s
                  AND p.post_type = 'tanz_product'
                  AND p.post_status IN ('publish','pending','private','draft')
                  AND CAST(qty.meta_value AS SIGNED) <= COALESCE(NULLIF(alert.meta_value,''), %d)",
                $stock_alert_key,
                $stock_qty_key,
                $low_stock_threshold
            ) );

            $pending_reviews = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$this->product_reviews_table} WHERE status = 'pending'" );

            return [
                'total_products'   => (int) $total,
                'status_breakdown' => [
                    'publish' => (int) ( $status_counts->publish ?? 0 ),
                    'draft'   => (int) ( $status_counts->draft ?? 0 ),
                    'pending' => (int) ( $status_counts->pending ?? 0 ),
                    'private' => (int) ( $status_counts->private ?? 0 ),
                ],
                'low_stock'        => $low_stock,
                'pending_reviews'  => $pending_reviews,
            ];
        }

        private function build_order_response( array $row, string $context = 'summary' ): array {
            $response = [
                'id'             => (int) $row['id'],
                'order_number'   => $row['order_number'],
                'status'         => $row['status'],
                'channel'        => $row['channel'],
                'payment_method' => $row['payment_method'],
                'amounts'        => [
                    'total'          => (float) $row['total'],
                    'subtotal'       => (float) $row['subtotal'],
                    'discount_total' => (float) $row['discount_total'],
                    'shipping_total' => (float) $row['shipping_total'],
                    'points_used'    => (int) $row['points_used'],
                    'currency'       => $row['currency'],
                ],
                'customer'       => $this->get_order_customer_summary( $row ),
                'timestamps'     => array_filter(
                    [
                        'created_at'   => $row['created_at'],
                        'updated_at'   => $row['updated_at'],
                        'paid_at'      => $row['paid_at'] ?? null,
                        'shipped_at'   => $row['shipped_at'] ?? null,
                        'completed_at' => $row['completed_at'] ?? null,
                        'cancelled_at' => $row['cancelled_at'] ?? null,
                    ],
                    static fn( $value ) => ! empty( $value )
                ),
                'tracking'       => $this->get_order_tracking_payload( $row ),
                'tracking_meta'  => $this->get_order_tracking_meta( $row ),
            ];

            if ( 'detail' === $context ) {
                $response['items']           = $this->get_order_items_payload( (int) $row['id'] );
                $response['customer_detail'] = $this->get_order_customer_detail( $row );
                $response['invoice']         = $this->get_order_invoice_payload( $row );
                $response['notes']           = $this->get_order_notes_payload( $row );
                $response['timeline']        = $this->get_order_timeline( $row );
                $response['audit_log']       = $this->get_order_audit_log( (int) $row['id'] );
                $response['tracking_events'] = $this->get_order_tracking_history( (int) $row['id'] );
            }

            return $response;
        }

        private function get_order_tracking_payload( array $row ): array {
            $provider = $row['tracking_provider'] ?? null;
            $number   = $row['tracking_number'] ?? null;
            $synced   = $row['tracking_synced_at'] ?? null;

            $latest = null;
            if ( $provider && $number ) {
                $latest = $this->get_latest_tracking_event( (int) $row['id'] );
            }

            return [
                'provider'     => $provider,
                'number'       => $number,
                'synced_at'    => $synced,
                'latest_event' => $latest,
            ];
        }

        private function get_order_tracking_meta( array $row ): array {
            return array_filter(
                [
                    'synced_at' => $row['tracking_synced_at'] ?? null,
                ],
                static fn( $value ) => null !== $value && '' !== $value
            );
        }

        private function get_order_items_payload( int $order_id ): array {
            $items = $this->fetch_order_items( $order_id );

            return array_map(
                static function ( array $item ): array {
                    return [
                        'product_id'    => (int) $item['product_id'],
                        'sku_id'        => (int) $item['sku_id'],
                        'product_title' => $item['product_title'],
                        'sku_code'      => $item['sku_code'],
                        'quantity'      => (int) $item['quantity'],
                        'price'         => (float) $item['price'],
                        'total'         => (float) $item['total'],
                        'meta'          => $item['meta'],
                    ];
                },
                $items
            );
        }

        private function get_order_customer_summary( array $row ): array {
            $user_id = isset( $row['user_id'] ) ? (int) $row['user_id'] : 0;

            if ( $user_id <= 0 ) {
                return [];
            }

            $user = get_userdata( $user_id );
            if ( ! $user ) {
                return [];
            }

            $summary = [
                'id'    => $user_id,
                'name'  => $user->display_name ?: $user->user_login,
                'email' => $user->user_email,
            ];

            $phone = get_user_meta( $user_id, 'billing_phone', true ) ?: get_user_meta( $user_id, 'phone', true );
            if ( $phone ) {
                $summary['phone'] = $phone;
            }

            $level = get_user_meta( $user_id, 'membership_level', true );
            if ( $level ) {
                $summary['membership_level'] = $level;
            }

            return $summary;
        }

        private function get_order_customer_detail( array $row ): array {
            $summary  = $this->get_order_customer_summary( $row );
            $meta     = is_string( $row['meta'] ?? null ) ? json_decode( $row['meta'], true ) : ( $row['meta'] ?? [] );
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            $detail = $summary;

            $address = $meta['shipping_address'] ?? [];
            if ( is_array( $address ) && ! empty( $address ) ) {
                $detail['shipping_address'] = array_map( 'sanitize_text_field', $address );
            }

            if ( ! empty( $meta['customer_note'] ) ) {
                $detail['customer_note'] = sanitize_textarea_field( (string) $meta['customer_note'] );
            }

            if ( ! empty( $meta['member_level'] ) ) {
                $detail['membership_level'] = sanitize_text_field( (string) $meta['member_level'] );
            }

            return $detail;
        }

        private function get_order_invoice_payload( array $row ): array {
            $meta = is_string( $row['meta'] ?? null ) ? json_decode( $row['meta'], true ) : ( $row['meta'] ?? [] );
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            $invoice = $meta['invoice'] ?? [];
            if ( ! is_array( $invoice ) ) {
                return [];
            }

            return array_filter(
                [
                    'type'       => isset( $invoice['type'] ) ? sanitize_key( $invoice['type'] ) : null,
                    'title'      => isset( $invoice['title'] ) ? sanitize_text_field( $invoice['title'] ) : null,
                    'tax_number' => isset( $invoice['tax_number'] ) ? sanitize_text_field( $invoice['tax_number'] ) : null,
                    'content'    => isset( $invoice['content'] ) ? sanitize_text_field( $invoice['content'] ) : null,
                ],
                static fn( $value ) => null !== $value && '' !== $value
            );
        }

        private function get_order_notes_payload( array $row ): array {
            $meta = is_string( $row['meta'] ?? null ) ? json_decode( $row['meta'], true ) : ( $row['meta'] ?? [] );
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            return array_filter(
                [
                    'customer_note' => isset( $meta['customer_note'] ) ? sanitize_textarea_field( (string) $meta['customer_note'] ) : null,
                    'admin_note'    => isset( $meta['admin_note'] ) ? sanitize_textarea_field( (string) $meta['admin_note'] ) : null,
                ],
                static fn( $value ) => null !== $value && '' !== $value
            );
        }

        private function get_order_timeline( array $row ): array {
            $events = [];

            $map = [
                'created'   => $row['created_at'] ?? null,
                'paid'      => $row['paid_at'] ?? null,
                'shipped'   => $row['shipped_at'] ?? null,
                'completed' => $row['completed_at'] ?? null,
                'cancelled' => $row['cancelled_at'] ?? null,
            ];

            foreach ( $map as $key => $timestamp ) {
                if ( empty( $timestamp ) ) {
                    continue;
                }

                $events[] = [
                    'event' => $key,
                    'at'    => $timestamp,
                ];
            }

            return $events;
        }

        private function get_order_audit_log( int $order_id ): array {
            global $wpdb;

            $rows = $wpdb->get_results(
                $wpdb->prepare(
                    "SELECT actor_name, action, entity_type, meta, created_at FROM {$this->audit_log_table} WHERE entity_type = %s AND entity_id = %d ORDER BY id DESC LIMIT 20",
                    'order',
                    $order_id
                ),
                ARRAY_A
            );

            if ( ! $rows ) {
                return [];
            }

            return array_map(
                static function ( array $row ): array {
                    $meta = [];
                    if ( ! empty( $row['meta'] ) ) {
                        $decoded = json_decode( $row['meta'], true );
                        if ( is_array( $decoded ) ) {
                            $meta = $decoded;
                        }
                    }

                    return [
                        'actor_name' => $row['actor_name'],
                        'action'     => $row['action'],
                        'meta'       => $meta,
                        'created_at' => $row['created_at'],
                    ];
                },
                $rows
            );
        }

        private function get_latest_tracking_event( int $order_id ): ?array {
            global $wpdb;

            $row = $wpdb->get_row(
                $wpdb->prepare(
                    "SELECT event_code, status_text, location, event_time, created_at FROM {$this->tracking_events_table} WHERE order_id = %d ORDER BY COALESCE(event_time, created_at) DESC LIMIT 1",
                    $order_id
                ),
                ARRAY_A
            );

            if ( ! $row ) {
                return null;
            }

            return [
                'event_code'  => $row['event_code'],
                'status_text' => $row['status_text'],
                'location'    => $row['location'],
                'event_time'  => $row['event_time'],
                'created_at'  => $row['created_at'],
            ];
        }

        private function get_order_tracking_history( int $order_id ): array {
            global $wpdb;

            $rows = $wpdb->get_results(
                $wpdb->prepare(
                    "SELECT event_code, status_text, location, event_time, created_at FROM {$this->tracking_events_table} WHERE order_id = %d ORDER BY COALESCE(event_time, created_at) DESC",
                    $order_id
                ),
                ARRAY_A
            );

            if ( ! $rows ) {
                return [];
            }

            return array_map(
                static function ( array $row ): array {
                    return [
                        'event_code'  => $row['event_code'],
                        'status_text' => $row['status_text'],
                        'location'    => $row['location'],
                        'event_time'  => $row['event_time'],
                        'created_at'  => $row['created_at'],
                    ];
                },
                $rows
            );
        }

        private function fetch_order_row( int $id ): ?array {
            global $wpdb;

            $row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->orders_table} WHERE id = %d", $id ), ARRAY_A );

            return $row ?: null;
        }

        // Shipping Templates 辅助函数 - 已迁移到 Tanzanite_REST_ShippingTemplates_Controller
        // decode_shipping_rules -> Tanzanite_REST_ShippingTemplates_Controller::decode_shipping_rules()
        // sanitize_shipping_rules -> Tanzanite_REST_ShippingTemplates_Controller::sanitize_shipping_rules()
        // format_shipping_template_row -> Tanzanite_REST_ShippingTemplates_Controller::format_shipping_template_row()

        private function sanitize_payment_terminals( $terminals ): array {
            if ( is_string( $terminals ) ) {
                $terminals = array_map( 'trim', explode( ',', $terminals ) );
            }

            if ( ! is_array( $terminals ) ) {
                return [];
            }

            $terminals = array_map( 'sanitize_key', $terminals );
            $terminals = array_values( array_intersect( $terminals, self::PAYMENT_TERMINALS ) );

            return $terminals;
        }

        private function sanitize_membership_levels( $levels ): array {
            if ( is_string( $levels ) ) {
                $levels = array_map( 'trim', explode( ',', $levels ) );
            }

            if ( ! is_array( $levels ) ) {
                return [];
            }

            $levels = array_map( static function ( $value ) {
                return sanitize_text_field( (string) $value );
            }, $levels );

            return array_values( array_filter( $levels ) );
        }

        private function payment_method_code_exists( string $code, int $exclude_id = 0 ): bool {
            global $wpdb;

            $sql = "SELECT id FROM {$this->payment_methods_table} WHERE code = %s";
            $params = [ $code ];

            if ( $exclude_id > 0 ) {
                $sql    .= ' AND id <> %d';
                $params[] = $exclude_id;
            }

            $sql  .= ' LIMIT 1';
            $found = $wpdb->get_var( $wpdb->prepare( $sql, $params ) );

            return (bool) $found;
        }

        private function fetch_payment_method_row( int $id ): ?array {
            global $wpdb;

            $row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->payment_methods_table} WHERE id = %d", $id ), ARRAY_A );

            return $row ?: null;
        }

        private function format_payment_method_row( array $row ): array {
            $terminals = $row['terminals'] ? json_decode( $row['terminals'], true ) : [];
            if ( ! is_array( $terminals ) ) {
                $terminals = [];
            }
            $terminals = array_values( array_intersect( array_map( 'sanitize_key', $terminals ), self::PAYMENT_TERMINALS ) );

            $membership_levels = $row['membership_levels'] ? json_decode( $row['membership_levels'], true ) : [];
            if ( ! is_array( $membership_levels ) ) {
                $membership_levels = [];
            }
            $membership_levels = array_map( static fn( $level ) => sanitize_text_field( (string) $level ), $membership_levels );

            $currencies = $row['currencies'] ? json_decode( $row['currencies'], true ) : [];
            if ( ! is_array( $currencies ) ) {
                $currencies = [];
            }
            $currencies = array_values( array_map( static fn( $currency ) => strtoupper( sanitize_text_field( (string) $currency ) ), $currencies ) );

            $settings = $row['settings'] ? json_decode( $row['settings'], true ) : [];
            if ( ! is_array( $settings ) ) {
                $settings = [];
            }

            $meta = $row['meta'] ? json_decode( $row['meta'], true ) : [];
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            return [
                'id'                 => (int) $row['id'],
                'code'               => $row['code'],
                'name'               => $row['name'],
                'description'        => $row['description'],
                'icon_url'           => $row['icon_url'] ? esc_url_raw( $row['icon_url'] ) : '',
                'fee_type'           => $row['fee_type'],
                'fee_value'          => (float) $row['fee_value'],
                'terminals'          => $terminals,
                'membership_levels'  => $membership_levels,
                'currencies'         => $currencies,
                'default_currency'   => sanitize_text_field( (string) $row['default_currency'] ),
                'settings'           => $settings,
                'is_enabled'         => (bool) $row['is_enabled'],
                'sort_order'         => (int) $row['sort_order'],
                'meta'               => $meta,
                'created_at'         => $row['created_at'],
                'updated_at'         => $row['updated_at'],
            ];
        }

        private function sanitize_payment_method_request( \WP_REST_Request $request, bool $is_update = false, array $existing = [] ) {
            $data   = [];
            $format = [];

            if ( ! $is_update || $request->has_param( 'code' ) ) {
                $code = sanitize_key( (string) $request->get_param( 'code' ) );
                if ( '' === $code ) {
                    return new \WP_Error( 'invalid_payment_payload', __( '支付方式编码不能为空。', 'tanzanite-settings' ) );
                }

                $exclude = $is_update ? (int) ( $existing['id'] ?? 0 ) : 0;
                if ( $this->payment_method_code_exists( $code, $exclude ) ) {
                    return new \WP_Error( 'duplicate_payment_code', __( '支付方式编码已存在，请更换。', 'tanzanite-settings' ) );
                }

                $data['code'] = $code;
                $format[]     = '%s';
            }

            if ( ! $is_update || $request->has_param( 'name' ) ) {
                $name = sanitize_text_field( (string) $request->get_param( 'name' ) );
                if ( '' === $name ) {
                    return new \WP_Error( 'invalid_payment_payload', __( '支付方式名称不能为空。', 'tanzanite-settings' ) );
                }

                $data['name'] = $name;
                $format[]     = '%s';
            }

            if ( $request->has_param( 'description' ) ) {
                $data['description'] = sanitize_textarea_field( (string) $request->get_param( 'description' ) );
                $format[]            = '%s';
            }

            if ( $request->has_param( 'icon_url' ) ) {
                $icon_url = trim( (string) $request->get_param( 'icon_url' ) );
                if ( '' !== $icon_url ) {
                    $icon_url = esc_url_raw( $icon_url );
                }
                $data['icon_url'] = $icon_url;
                $format[]         = '%s';
            }

            if ( ! $is_update || $request->has_param( 'fee_type' ) ) {
                $fee_type = sanitize_key( (string) $request->get_param( 'fee_type' ) );
                if ( ! in_array( $fee_type, [ 'fixed', 'percentage' ], true ) ) {
                    return new \WP_Error( 'invalid_payment_payload', __( '手续费类型仅支持 fixed 或 percentage。', 'tanzanite-settings' ) );
                }

                $data['fee_type'] = $fee_type;
                $format[]         = '%s';
            }

            if ( ! $is_update || $request->has_param( 'fee_value' ) ) {
                $fee_value = (float) $request->get_param( 'fee_value' );
                if ( $fee_value < 0 ) {
                    return new \WP_Error( 'invalid_payment_payload', __( '手续费数值不能为负数。', 'tanzanite-settings' ) );
                }

                $data['fee_value'] = $fee_value;
                $format[]          = '%f';
            }

            if ( $request->has_param( 'terminals' ) ) {
                $terminals = $this->sanitize_payment_terminals( $request->get_param( 'terminals' ) );
                $data['terminals'] = wp_json_encode( $terminals );
                $format[]          = '%s';
            }

            if ( $request->has_param( 'membership_levels' ) ) {
                $levels = $this->sanitize_membership_levels( $request->get_param( 'membership_levels' ) );
                $data['membership_levels'] = wp_json_encode( $levels );
                $format[]                 = '%s';
            }

            if ( $request->has_param( 'currencies' ) ) {
                $currencies = $this->sanitize_currencies( $request->get_param( 'currencies' ) );
                $data['currencies'] = wp_json_encode( $currencies );
                $format[]          = '%s';
            }

            if ( $request->has_param( 'default_currency' ) ) {
                $default_currency = strtoupper( sanitize_text_field( (string) $request->get_param( 'default_currency' ) ) );
                $currencies       = isset( $data['currencies'] ) ? json_decode( $data['currencies'], true ) : ( $existing['currencies'] ?? [] );
                if ( ! is_array( $currencies ) ) {
                    $currencies = [];
                }
                $normalized_currencies = array_values( array_map( static fn( $currency ) => strtoupper( $currency ), $currencies ) );

                if ( '' !== $default_currency && ! in_array( $default_currency, $normalized_currencies, true ) ) {
                    return new \WP_Error( 'invalid_payment_payload', __( '默认币别必须在支持的货币列表中。', 'tanzanite-settings' ) );
                }

                $data['default_currency'] = $default_currency;
                $format[]                 = '%s';
            } elseif ( ! $is_update ) {
                $data['default_currency'] = '';
                $format[]                 = '%s';
            }

            if ( $request->has_param( 'settings' ) ) {
                $settings = $request->get_param( 'settings' );
                if ( ! is_array( $settings ) ) {
                    return new \WP_Error( 'invalid_payment_payload', __( '设置字段必须为对象。', 'tanzanite-settings' ) );
                }

                $data['settings'] = wp_json_encode( $settings );
                $format[]         = '%s';
            }

            if ( $request->has_param( 'meta' ) ) {
                $meta = $request->get_param( 'meta' );
                if ( ! is_array( $meta ) ) {
                    return new \WP_Error( 'invalid_payment_payload', __( 'Meta 字段必须为对象。', 'tanzanite-settings' ) );
                }

                $data['meta'] = wp_json_encode( $meta );
                $format[]     = '%s';
            }

            if ( $request->has_param( 'is_enabled' ) ) {
                $data['is_enabled'] = $request->get_param( 'is_enabled' ) ? 1 : 0;
                $format[]           = '%d';
            } elseif ( ! $is_update ) {
                $data['is_enabled'] = 1;
                $format[]           = '%d';
            }

            if ( $request->has_param( 'sort_order' ) ) {
                $data['sort_order'] = (int) $request->get_param( 'sort_order' );
                $format[]           = '%d';
            } elseif ( ! $is_update ) {
                $data['sort_order'] = 0;
                $format[]           = '%d';
            }

            if ( empty( $data ) ) {
                return new \WP_Error( 'invalid_payment_payload', __( '没有可更新的字段。', 'tanzanite-settings' ) );
            }

            return [
                'data'   => $data,
                'format' => $format,
            ];
        }

        private function sanitize_currencies( $currencies ): array {
            if ( is_string( $currencies ) ) {
                $currencies = preg_split( '/[\s,]+/', $currencies, -1, PREG_SPLIT_NO_EMPTY );
            }

            if ( ! is_array( $currencies ) ) {
                return [];
            }

            $currencies = array_map( static fn( $value ) => strtoupper( sanitize_text_field( (string) $value ) ), $currencies );
            $currencies = array_filter( $currencies );

            return array_values( array_unique( $currencies ) );
        }

        // Payment Methods REST API 回调函数 - 已迁移到 Tanzanite_REST_Payments_Controller
        // rest_list_payment_methods -> Tanzanite_REST_Payments_Controller::get_items()
        // rest_get_payment_method -> Tanzanite_REST_Payments_Controller::get_item()
        // rest_create_payment_method -> Tanzanite_REST_Payments_Controller::create_item()
        // rest_update_payment_method -> Tanzanite_REST_Payments_Controller::update_item()
        // rest_delete_payment_method -> Tanzanite_REST_Payments_Controller::delete_item()

        // Tax Rates REST API 回调函数 - 已迁移到 Tanzanite_REST_TaxRates_Controller
        // rest_list_tax_rates -> Tanzanite_REST_TaxRates_Controller::get_items()
        // rest_get_tax_rate -> Tanzanite_REST_TaxRates_Controller::get_item()
        // rest_create_tax_rate -> Tanzanite_REST_TaxRates_Controller::create_item()
        // rest_update_tax_rate -> Tanzanite_REST_TaxRates_Controller::update_item()
        // rest_delete_tax_rate -> Tanzanite_REST_TaxRates_Controller::delete_item()
        // fetch_tax_rate_row -> Tanzanite_REST_TaxRates_Controller::fetch_tax_rate_row()
        // format_tax_rate_row -> Tanzanite_REST_TaxRates_Controller::format_tax_rate_row()

        // ========== 属性 REST API 处理函数 - 已迁移到 Tanzanite_REST_Attributes_Controller ==========
        // rest_list_attributes -> Tanzanite_REST_Attributes_Controller::get_items()
        // rest_create_attribute -> Tanzanite_REST_Attributes_Controller::create_item()
        // rest_get_attribute -> Tanzanite_REST_Attributes_Controller::get_item()
        // rest_update_attribute -> Tanzanite_REST_Attributes_Controller::update_item()
        // rest_delete_attribute -> Tanzanite_REST_Attributes_Controller::delete_item()
        // rest_list_attribute_values -> Tanzanite_REST_Attributes_Controller::get_values()
        // rest_create_attribute_value -> Tanzanite_REST_Attributes_Controller::create_value()
        // rest_update_attribute_value -> Tanzanite_REST_Attributes_Controller::update_value()
        // rest_delete_attribute_value -> Tanzanite_REST_Attributes_Controller::delete_value()
        // fetch_attribute_row, format_attribute_row, fetch_attribute_value_row, format_attribute_value_row - 已迁移

        private function rest_list_attributes( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $page     = max( 1, (int) $request->get_param( 'page' ) );
            $per_page = max( 1, min( 100, (int) $request->get_param( 'per_page' ) ) );
            $offset   = ( $page - 1 ) * $per_page;

            $total = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$this->product_attributes_table}" );

            $rows = $wpdb->get_results(
                $wpdb->prepare(
                    "SELECT * FROM {$this->product_attributes_table} ORDER BY sort_order ASC, id ASC LIMIT %d OFFSET %d",
                    $per_page,
                    $offset
                ),
                ARRAY_A
            );

            $items = [];
            foreach ( $rows as $row ) {
                $item = $this->format_attribute_row( $row );
                
                $value_count = (int) $wpdb->get_var( $wpdb->prepare(
                    "SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d",
                    $row['id']
                ) );
                $item['values_count'] = $value_count;
                
                $items[] = $item;
            }

            return new \WP_REST_Response( [
                'items'       => $items,
                'total'       => $total,
                'page'        => $page,
                'per_page'    => $per_page,
                'total_pages' => ceil( $total / $per_page ),
            ] );
        }

        private function rest_create_attribute( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $name = sanitize_text_field( (string) $request->get_param( 'name' ) );
            if ( '' === $name ) {
                return $this->respond_error( 'invalid_attribute_payload', __( '属性名称不能为空。', 'tanzanite-settings' ) );
            }

            $slug = $request->get_param( 'slug' );
            if ( empty( $slug ) ) {
                $slug = sanitize_title( $name );
            } else {
                $slug = sanitize_title( $slug );
            }

            $exists = $wpdb->get_var( $wpdb->prepare(
                "SELECT COUNT(*) FROM {$this->product_attributes_table} WHERE slug = %s",
                $slug
            ) );

            if ( $exists > 0 ) {
                return $this->respond_error( 'duplicate_attribute_slug', __( '属性 Slug 已存在，请更换。', 'tanzanite-settings' ) );
            }

            $type = sanitize_key( (string) $request->get_param( 'type' ) );
            if ( ! in_array( $type, [ 'select', 'color', 'image' ], true ) ) {
                $type = 'select';
            }

            $meta = $request->get_param( 'meta' );
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            $data = [
                'name'          => $name,
                'slug'          => $slug,
                'type'          => $type,
                'is_filterable' => (bool) $request->get_param( 'is_filterable' ),
                'affects_sku'   => (bool) $request->get_param( 'affects_sku' ),
                'affects_stock' => (bool) $request->get_param( 'affects_stock' ),
                'is_enabled'    => (bool) $request->get_param( 'is_enabled' ),
                'sort_order'    => (int) $request->get_param( 'sort_order' ),
                'meta'          => wp_json_encode( $meta ),
            ];

            $inserted = $wpdb->insert( $this->product_attributes_table, $data, [ '%s', '%s', '%s', '%d', '%d', '%d', '%d', '%d', '%s' ] );
            if ( false === $inserted ) {
                return $this->respond_error( 'failed_create_attribute', __( '创建属性失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $id   = (int) $wpdb->insert_id;
            $row  = $this->fetch_attribute_row( $id );
            $item = $this->format_attribute_row( $row );
            $item['values_count'] = 0;

            $this->log_audit( 'create', 'product_attribute', $id, [ 'name' => $item['name'] ], $request );

            return new \WP_REST_Response( $item, 201 );
        }

        private function rest_get_attribute( \WP_REST_Request $request ): \WP_REST_Response {
            $id  = (int) $request['id'];
            $row = $this->fetch_attribute_row( $id );

            if ( ! $row ) {
                return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
            }

            $item = $this->format_attribute_row( $row );
            
            global $wpdb;
            $value_count = (int) $wpdb->get_var( $wpdb->prepare(
                "SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d",
                $id
            ) );
            $item['values_count'] = $value_count;

            return new \WP_REST_Response( $item );
        }

        private function rest_update_attribute( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $id  = (int) $request['id'];
            $row = $this->fetch_attribute_row( $id );

            if ( ! $row ) {
                return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
            }

            $data   = [];
            $format = [];

            if ( $request->has_param( 'name' ) ) {
                $name = sanitize_text_field( (string) $request->get_param( 'name' ) );
                if ( '' === $name ) {
                    return $this->respond_error( 'invalid_attribute_payload', __( '属性名称不能为空。', 'tanzanite-settings' ) );
                }
                $data['name'] = $name;
                $format[]     = '%s';
            }

            if ( $request->has_param( 'slug' ) ) {
                $slug = sanitize_title( (string) $request->get_param( 'slug' ) );
                
                $exists = $wpdb->get_var( $wpdb->prepare(
                    "SELECT COUNT(*) FROM {$this->product_attributes_table} WHERE slug = %s AND id != %d",
                    $slug,
                    $id
                ) );

                if ( $exists > 0 ) {
                    return $this->respond_error( 'duplicate_attribute_slug', __( '属性 Slug 已存在，请更换。', 'tanzanite-settings' ) );
                }

                $data['slug'] = $slug;
                $format[]     = '%s';
            }

            if ( $request->has_param( 'type' ) ) {
                $type = sanitize_key( (string) $request->get_param( 'type' ) );
                if ( ! in_array( $type, [ 'select', 'color', 'image' ], true ) ) {
                    $type = 'select';
                }
                $data['type'] = $type;
                $format[]     = '%s';
            }

            if ( $request->has_param( 'is_filterable' ) ) {
                $data['is_filterable'] = (bool) $request->get_param( 'is_filterable' );
                $format[]              = '%d';
            }

            if ( $request->has_param( 'affects_sku' ) ) {
                $data['affects_sku'] = (bool) $request->get_param( 'affects_sku' );
                $format[]            = '%d';
            }

            if ( $request->has_param( 'affects_stock' ) ) {
                $data['affects_stock'] = (bool) $request->get_param( 'affects_stock' );
                $format[]              = '%d';
            }

            if ( $request->has_param( 'is_enabled' ) ) {
                $data['is_enabled'] = (bool) $request->get_param( 'is_enabled' );
                $format[]           = '%d';
            }

            if ( $request->has_param( 'sort_order' ) ) {
                $data['sort_order'] = (int) $request->get_param( 'sort_order' );
                $format[]           = '%d';
            }

            if ( $request->has_param( 'meta' ) ) {
                $meta = $request->get_param( 'meta' );
                if ( ! is_array( $meta ) ) {
                    $meta = [];
                }
                $data['meta'] = wp_json_encode( $meta );
                $format[]     = '%s';
            }

            if ( empty( $data ) ) {
                return $this->respond_error( 'no_update_data', __( '没有需要更新的数据。', 'tanzanite-settings' ) );
            }

            $updated = $wpdb->update( $this->product_attributes_table, $data, [ 'id' => $id ], $format, [ '%d' ] );
            if ( false === $updated ) {
                return $this->respond_error( 'failed_update_attribute', __( '更新属性失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $updated_row = $this->fetch_attribute_row( $id );
            $item        = $this->format_attribute_row( $updated_row );
            
            $value_count = (int) $wpdb->get_var( $wpdb->prepare(
                "SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d",
                $id
            ) );
            $item['values_count'] = $value_count;

            $this->log_audit( 'update', 'product_attribute', $id, [ 'name' => $item['name'] ], $request );

            return new \WP_REST_Response( $item );
        }

        private function rest_delete_attribute( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $id  = (int) $request['id'];
            $row = $this->fetch_attribute_row( $id );

            if ( ! $row ) {
                return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
            }

            $wpdb->delete( $this->attribute_values_table, [ 'attribute_id' => $id ], [ '%d' ] );

            $deleted = $wpdb->delete( $this->product_attributes_table, [ 'id' => $id ], [ '%d' ] );
            if ( false === $deleted ) {
                return $this->respond_error( 'failed_delete_attribute', __( '删除属性失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $this->log_audit( 'delete', 'product_attribute', $id, [ 'name' => $row['name'] ], $request );

            return new \WP_REST_Response( [ 'deleted' => true ] );
        }

        private function fetch_attribute_row( int $id ): ?array {
            global $wpdb;

            $row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->product_attributes_table} WHERE id = %d", $id ), ARRAY_A );

            return $row ?: null;
        }

        private function format_attribute_row( array $row ): array {
            $meta = $row['meta'] ? json_decode( $row['meta'], true ) : [];
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            return [
                'id'            => (int) $row['id'],
                'name'          => $row['name'],
                'slug'          => $row['slug'],
                'type'          => $row['type'],
                'is_filterable' => (bool) $row['is_filterable'],
                'affects_sku'   => (bool) $row['affects_sku'],
                'affects_stock' => (bool) $row['affects_stock'],
                'is_enabled'    => (bool) $row['is_enabled'],
                'sort_order'    => (int) $row['sort_order'],
                'meta'          => $meta,
                'created_at'    => $row['created_at'],
                'updated_at'    => $row['updated_at'],
            ];
        }

        // ========== 属性值 REST API 处理函数 ==========

        private function rest_list_attribute_values( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $attribute_id = (int) $request['attribute_id'];

            $attribute = $this->fetch_attribute_row( $attribute_id );
            if ( ! $attribute ) {
                return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
            }

            $rows = $wpdb->get_results(
                $wpdb->prepare(
                    "SELECT * FROM {$this->attribute_values_table} WHERE attribute_id = %d ORDER BY sort_order ASC, id ASC",
                    $attribute_id
                ),
                ARRAY_A
            );

            $items = [];
            foreach ( $rows as $row ) {
                $items[] = $this->format_attribute_value_row( $row );
            }

            return new \WP_REST_Response( [
                'items' => $items,
                'total' => count( $items ),
            ] );
        }

        private function rest_create_attribute_value( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $attribute_id = (int) $request['attribute_id'];

            $attribute = $this->fetch_attribute_row( $attribute_id );
            if ( ! $attribute ) {
                return $this->respond_error( 'attribute_not_found', __( '指定的属性不存在。', 'tanzanite-settings' ), 404 );
            }

            $name = sanitize_text_field( (string) $request->get_param( 'name' ) );
            if ( '' === $name ) {
                return $this->respond_error( 'invalid_value_payload', __( '属性值名称不能为空。', 'tanzanite-settings' ) );
            }

            $slug = $request->get_param( 'slug' );
            if ( empty( $slug ) ) {
                $slug = sanitize_title( $name );
            } else {
                $slug = sanitize_title( $slug );
            }

            $exists = $wpdb->get_var( $wpdb->prepare(
                "SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d AND slug = %s",
                $attribute_id,
                $slug
            ) );

            if ( $exists > 0 ) {
                return $this->respond_error( 'duplicate_value_slug', __( '属性值 Slug 已存在，请更换。', 'tanzanite-settings' ) );
            }

            $value = sanitize_text_field( (string) $request->get_param( 'value' ) );
            $meta  = $request->get_param( 'meta' );
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            $data = [
                'attribute_id' => $attribute_id,
                'name'         => $name,
                'slug'         => $slug,
                'value'        => $value,
                'is_enabled'   => (bool) $request->get_param( 'is_enabled' ),
                'sort_order'   => (int) $request->get_param( 'sort_order' ),
                'meta'         => wp_json_encode( $meta ),
            ];

            $inserted = $wpdb->insert( $this->attribute_values_table, $data, [ '%d', '%s', '%s', '%s', '%d', '%d', '%s' ] );
            if ( false === $inserted ) {
                return $this->respond_error( 'failed_create_value', __( '创建属性值失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $id   = (int) $wpdb->insert_id;
            $row  = $this->fetch_attribute_value_row( $id );
            $item = $this->format_attribute_value_row( $row );

            $this->log_audit( 'create', 'attribute_value', $id, [ 'name' => $item['name'], 'attribute_id' => $attribute_id ], $request );

            return new \WP_REST_Response( $item, 201 );
        }

        private function rest_update_attribute_value( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $attribute_id = (int) $request['attribute_id'];
            $id           = (int) $request['id'];

            $row = $this->fetch_attribute_value_row( $id );
            if ( ! $row || (int) $row['attribute_id'] !== $attribute_id ) {
                return $this->respond_error( 'value_not_found', __( '指定的属性值不存在。', 'tanzanite-settings' ), 404 );
            }

            $data   = [];
            $format = [];

            if ( $request->has_param( 'name' ) ) {
                $name = sanitize_text_field( (string) $request->get_param( 'name' ) );
                if ( '' === $name ) {
                    return $this->respond_error( 'invalid_value_payload', __( '属性值名称不能为空。', 'tanzanite-settings' ) );
                }
                $data['name'] = $name;
                $format[]     = '%s';
            }

            if ( $request->has_param( 'slug' ) ) {
                $slug = sanitize_title( (string) $request->get_param( 'slug' ) );
                
                $exists = $wpdb->get_var( $wpdb->prepare(
                    "SELECT COUNT(*) FROM {$this->attribute_values_table} WHERE attribute_id = %d AND slug = %s AND id != %d",
                    $attribute_id,
                    $slug,
                    $id
                ) );

                if ( $exists > 0 ) {
                    return $this->respond_error( 'duplicate_value_slug', __( '属性值 Slug 已存在，请更换。', 'tanzanite-settings' ) );
                }

                $data['slug'] = $slug;
                $format[]     = '%s';
            }

            if ( $request->has_param( 'value' ) ) {
                $data['value'] = sanitize_text_field( (string) $request->get_param( 'value' ) );
                $format[]      = '%s';
            }

            if ( $request->has_param( 'is_enabled' ) ) {
                $data['is_enabled'] = (bool) $request->get_param( 'is_enabled' );
                $format[]           = '%d';
            }

            if ( $request->has_param( 'sort_order' ) ) {
                $data['sort_order'] = (int) $request->get_param( 'sort_order' );
                $format[]           = '%d';
            }

            if ( $request->has_param( 'meta' ) ) {
                $meta = $request->get_param( 'meta' );
                if ( ! is_array( $meta ) ) {
                    $meta = [];
                }
                $data['meta'] = wp_json_encode( $meta );
                $format[]     = '%s';
            }

            if ( empty( $data ) ) {
                return $this->respond_error( 'no_update_data', __( '没有需要更新的数据。', 'tanzanite-settings' ) );
            }

            $updated = $wpdb->update( $this->attribute_values_table, $data, [ 'id' => $id ], $format, [ '%d' ] );
            if ( false === $updated ) {
                return $this->respond_error( 'failed_update_value', __( '更新属性值失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $updated_row = $this->fetch_attribute_value_row( $id );
            $item        = $this->format_attribute_value_row( $updated_row );

            $this->log_audit( 'update', 'attribute_value', $id, [ 'name' => $item['name'] ], $request );

            return new \WP_REST_Response( $item );
        }

        private function rest_delete_attribute_value( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $attribute_id = (int) $request['attribute_id'];
            $id           = (int) $request['id'];

            $row = $this->fetch_attribute_value_row( $id );
            if ( ! $row || (int) $row['attribute_id'] !== $attribute_id ) {
                return $this->respond_error( 'value_not_found', __( '指定的属性值不存在。', 'tanzanite-settings' ), 404 );
            }

            $deleted = $wpdb->delete( $this->attribute_values_table, [ 'id' => $id ], [ '%d' ] );
            if ( false === $deleted ) {
                return $this->respond_error( 'failed_delete_value', __( '删除属性值失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $this->log_audit( 'delete', 'attribute_value', $id, [ 'name' => $row['name'] ], $request );

            return new \WP_REST_Response( [ 'deleted' => true ] );
        }

        private function fetch_attribute_value_row( int $id ): ?array {
            global $wpdb;

            $row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->attribute_values_table} WHERE id = %d", $id ), ARRAY_A );

            return $row ?: null;
        }

        private function format_attribute_value_row( array $row ): array {
            $meta = $row['meta'] ? json_decode( $row['meta'], true ) : [];
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            return [
                'id'           => (int) $row['id'],
                'attribute_id' => (int) $row['attribute_id'],
                'name'         => $row['name'],
                'slug'         => $row['slug'],
                'value'        => $row['value'],
                'is_enabled'   => (bool) $row['is_enabled'],
                'sort_order'   => (int) $row['sort_order'],
                'meta'         => $meta,
                'created_at'   => $row['created_at'],
                'updated_at'   => $row['updated_at'],
            ];
        }

        /**
         * REST：商品占位数据。
         */
        public function rest_list_products( \WP_REST_Request $request ): \WP_REST_Response {
            $per_page = (int) max( 1, min( 200, $request->get_param( 'per_page' ) ?: 20 ) );
            $page     = (int) max( 1, $request->get_param( 'page' ) ?: 1 );
            $keyword  = trim( (string) ( $request->get_param( 'keyword' ) ?: '' ) );
            $sku      = trim( (string) ( $request->get_param( 'sku' ) ?: '' ) );
            $status   = sanitize_key( (string) ( $request->get_param( 'status' ) ?: '' ) );
            $category = (int) ( $request->get_param( 'category' ) ?: 0 );
            $tags_raw        = $request->get_param( 'tags' );
            $attributes_param = trim( (string) ( $request->get_param( 'attributes' ) ?: '' ) );
            $author_id        = (int) ( $request->get_param( 'author' ) ?: 0 );
            $include_raw      = $request->get_param( 'include' );
            $min_inventory = $request->has_param( 'inventory_min' ) ? (int) $request->get_param( 'inventory_min' ) : null;
            $max_inventory = $request->has_param( 'inventory_max' ) ? (int) $request->get_param( 'inventory_max' ) : null;
            $points_min    = $request->has_param( 'points_min' ) ? (int) $request->get_param( 'points_min' ) : null;
            $points_max    = $request->has_param( 'points_max' ) ? (int) $request->get_param( 'points_max' ) : null;
            $sort    = sanitize_key( (string) ( $request->get_param( 'sort' ) ?: 'updated_at' ) );
            $order   = strtoupper( (string) ( $request->get_param( 'order' ) ?: 'DESC' ) );

            if ( ! in_array( $status, self::ALLOWED_PRODUCT_STATUSES, true ) ) {
                $status = '';
            }

            if ( ! in_array( $sort, [ 'updated_at', 'price_regular', 'stock_qty', 'points_reward' ], true ) ) {
                $sort = 'updated_at';
            }

            if ( ! in_array( $order, [ 'ASC', 'DESC' ], true ) ) {
                $order = 'DESC';
            }

            $meta_query = [];

            if ( null !== $min_inventory || null !== $max_inventory ) {
                if ( null !== $min_inventory && null !== $max_inventory ) {
                    $meta_query[] = [
                        'key'     => self::PRODUCT_META_MAP['stock_qty']['key'],
                        'value'   => [ $min_inventory, $max_inventory ],
                        'compare' => 'BETWEEN',
                        'type'    => 'NUMERIC',
                    ];
                } elseif ( null !== $min_inventory ) {
                    $meta_query[] = [
                        'key'     => self::PRODUCT_META_MAP['stock_qty']['key'],
                        'value'   => $min_inventory,
                        'compare' => '>=',
                        'type'    => 'NUMERIC',
                    ];
                } elseif ( null !== $max_inventory ) {
                    $meta_query[] = [
                        'key'     => self::PRODUCT_META_MAP['stock_qty']['key'],
                        'value'   => $max_inventory,
                        'compare' => '<=',
                        'type'    => 'NUMERIC',
                    ];
                }
            }

            if ( null !== $points_min || null !== $points_max ) {
                if ( null !== $points_min && null !== $points_max ) {
                    $meta_query[] = [
                        'key'     => self::PRODUCT_META_MAP['points_reward']['key'],
                        'value'   => [ $points_min, $points_max ],
                        'compare' => 'BETWEEN',
                        'type'    => 'NUMERIC',
                    ];
                } elseif ( null !== $points_min ) {
                    $meta_query[] = [
                        'key'     => self::PRODUCT_META_MAP['points_reward']['key'],
                        'value'   => $points_min,
                        'compare' => '>=',
                        'type'    => 'NUMERIC',
                    ];
                } elseif ( null !== $points_max ) {
                    $meta_query[] = [
                        'key'     => self::PRODUCT_META_MAP['points_reward']['key'],
                        'value'   => $points_max,
                        'compare' => '<=',
                        'type'    => 'NUMERIC',
                    ];
                }
            }

            $tax_query = [];
            if ( $category > 0 ) {
                $tax_query[] = [
                    'taxonomy' => 'category',
                    'field'    => 'term_id',
                    'terms'    => $category,
                ];
            }

            $tag_terms = [];
            if ( is_array( $tags_raw ) ) {
                foreach ( $tags_raw as $tag_item ) {
                    $sanitized = sanitize_title( (string) $tag_item );
                    if ( $sanitized ) {
                        $tag_terms[] = $sanitized;
                    }
                }
            } else {
                $tags_param = trim( (string) ( $tags_raw ?: '' ) );
                if ( $tags_param ) {
                    $tag_terms = array_filter(
                        array_map(
                            static function ( $tag ): string {
                                return sanitize_title( (string) $tag );
                            },
                            preg_split( '/[\s,]+/', $tags_param ) ?: []
                        )
                    );
                }
            }

            if ( $tag_terms ) {
                $tax_query[] = [
                    'taxonomy' => 'post_tag',
                    'field'    => 'slug',
                    'terms'    => $tag_terms,
                ];
            }

            if ( $attributes_param ) {
                $attribute_filters = array_filter( array_map( 'trim', explode( ',', $attributes_param ) ) );
                foreach ( $attribute_filters as $filter ) {
                    $parts = array_map( 'trim', explode( ':', $filter, 2 ) );
                    $taxonomy = $parts[0] ?? '';
                    $term     = $parts[1] ?? '';
                    if ( ! $taxonomy || ! $term ) {
                        continue;
                    }
                    $tax_query[] = [
                        'taxonomy' => $taxonomy,
                        'field'    => 'slug',
                        'terms'    => $term,
                    ];
                }
            }

            if ( null !== $min_inventory ) {
                $meta_query[] = [
                    'key'     => self::PRODUCT_META_MAP['stock_qty']['key'],
                    'value'   => $min_inventory,
                    'compare' => '>=',
                    'type'    => 'NUMERIC',
                ];
            }

            if ( null !== $max_inventory ) {
                $meta_query[] = [
                    'key'     => self::PRODUCT_META_MAP['stock_qty']['key'],
                    'value'   => $max_inventory,
                    'compare' => '<=',
                    'type'    => 'NUMERIC',
                ];
            }

            if ( $author_id > 0 ) {
                $args['author'] = $author_id;
            }

            if ( 'price_regular' === $sort ) {
                $args['meta_key'] = self::PRODUCT_META_MAP['price_regular']['key'];
                $args['orderby']  = 'meta_value_num';
            } elseif ( 'stock_qty' === $sort ) {
                $args['meta_key'] = self::PRODUCT_META_MAP['stock_qty']['key'];
                $args['orderby']  = 'meta_value_num';
            } elseif ( 'points_reward' === $sort ) {
                $args['meta_key'] = self::PRODUCT_META_MAP['points_reward']['key'];
                $args['orderby']  = 'meta_value_num';
            }

            add_filter( 'posts_where', [ $this, 'filter_products_by_sku' ], 10, 2 );
            $this->current_sku_keyword = $sku;

            $query = new \WP_Query( $args );

            remove_filter( 'posts_where', [ $this, 'filter_products_by_sku' ], 10 );
            $this->current_sku_keyword = '';

            $items = [];
            foreach ( $query->posts as $post ) {
                $items[] = $this->build_product_list_item( $post );
            }

            $stats = $this->get_product_statistics();

            return new \WP_REST_Response(
                [
                    'items'   => $items,
                    'meta'    => [
                        'page'           => $include_ids ? 1 : $page,
                        'per_page'       => $args['posts_per_page'],
                        'total_pages'    => (int) max( 1, $include_ids ? 1 : $query->max_num_pages ),
                        'total'          => (int) $query->found_posts,
                        'schema_version' => self::VERSION,
                        'filters'        => [ 'keyword', 'sku', 'status', 'category', 'tags', 'attributes', 'author', 'inventory_min', 'inventory_max', 'points_min', 'points_max' ],
                        'sorting'        => [ 'updated_at', 'price_regular', 'stock_qty', 'points_reward' ],
                        'include'        => $include_ids,
                    ],
                    'summary' => $stats,
                ]
            );
        }

        private function filter_products_by_sku( string $where, \WP_Query $query ): string {
            if ( ! empty( $this->current_sku_keyword ) && $query->get( 'post_type' ) === 'tanz_product' ) {
                global $wpdb;
                $sku_keyword = esc_sql( $wpdb->esc_like( $this->current_sku_keyword ) );
                $where      .= " AND EXISTS (SELECT 1 FROM {$this->product_skus_table} sku WHERE sku.product_id = {$wpdb->posts}.ID AND sku.sku_code LIKE '%{$sku_keyword}%')";
            }

            return $where;
        }

        /**
                    ],
                ]
            );
        }

        // Shipping Templates REST API 回调函数 - 已迁移到 Tanzanite_REST_ShippingTemplates_Controller
        // rest_list_shipping_templates -> Tanzanite_REST_ShippingTemplates_Controller::get_items()

        /**
         * REST：创建占位商品。
         */
        public function rest_create_product( \WP_REST_Request $request ): \WP_REST_Response {
            $status = $this->normalize_product_status( $request->get_param( 'status' ) );

            list( $content_html, $content_markdown ) = $this->prepare_content_for_save( $request );
            $request->set_param( 'content_markdown', $content_markdown );

            $postarr = [
                'post_type'    => 'tanz_product',
                'post_status'  => $status,
                'post_title'   => sanitize_text_field( (string) $request->get_param( 'title' ) ),
                'post_excerpt' => sanitize_textarea_field( (string) $request->get_param( 'excerpt' ) ),
                'post_content' => $content_html,
                'post_author'  => get_current_user_id(),
            ];

            $slug = $request->get_param( 'slug' );
            if ( $slug ) {
                $postarr['post_name'] = sanitize_title( (string) $slug );
            }

            $dates = $this->prepare_post_dates_from_request( $request->get_param( 'date' ) );
            if ( $dates ) {
                $postarr = array_merge( $postarr, $dates );
            }

            $post_id = wp_insert_post( $postarr, true );

            if ( is_wp_error( $post_id ) ) {
                return new \WP_REST_Response( [ 'message' => $post_id->get_error_message() ], 400 );
            }

            try {
                $this->update_product_meta_from_request( $post_id, $request );
            } catch ( \RuntimeException $exception ) {
                wp_delete_post( $post_id, true );
                return $this->respond_error( 'invalid_product_meta', $exception->getMessage(), 400 );
            }

            if ( $request->has_param( 'is_sticky' ) ) {
                $this->sync_post_sticky_status( $post_id, $request->get_param( 'is_sticky' ) );
            }

            $sku_result = $this->handle_product_skus( $post_id, $request );
            if ( is_wp_error( $sku_result ) ) {
                wp_delete_post( $post_id, true );
                return $this->error_to_response( $sku_result );
            }

            $seo_payload = null;
            if ( $request->has_param( 'seo' ) ) {
                $seo_param = $request->get_param( 'seo' );
                if ( is_array( $seo_param ) ) {
                    $seo_payload = $seo_param['payload'] ?? null;
                }
            }

            if ( null !== $seo_payload ) {
                $this->sync_product_seo( $post_id, $seo_payload );
            }

            $this->log_audit(
                'create',
                'product',
                $post_id,
                array_filter(
                    [
                        'title'      => get_the_title( $post_id ),
                        'status'     => $status,
                        'skus_count' => $sku_result['count'] ?? 0,
                    ]
                ),
                $request
            );

            return new \WP_REST_Response( $this->build_product_response( $post_id ), 201 );
        }

        public function rest_get_product( \WP_REST_Request $request ): \WP_REST_Response {
            $post_id = (int) $request['id'];
            $post    = get_post( $post_id );

            if ( ! $post || 'tanz_product' !== $post->post_type ) {
                return new \WP_REST_Response( [ 'message' => __( 'Product not found.', 'tanzanite-settings' ) ], 404 );
            }

            return new \WP_REST_Response( $this->build_product_response( $post_id ) );
        }

        public function rest_update_product( \WP_REST_Request $request ): \WP_REST_Response {
            $post_id = (int) $request['id'];
            $post    = get_post( $post_id );

            if ( ! $post || 'tanz_product' !== $post->post_type ) {
                return new \WP_REST_Response( [ 'message' => __( 'Product not found.', 'tanzanite-settings' ) ], 404 );
            }

            $data       = [ 'ID' => $post_id ];
            $has_update = false;

            if ( $request->has_param( 'title' ) ) {
                $data['post_title'] = sanitize_text_field( (string) $request->get_param( 'title' ) );
                $has_update         = true;
            }

            if ( $request->has_param( 'excerpt' ) ) {
                $data['post_excerpt'] = sanitize_textarea_field( (string) $request->get_param( 'excerpt' ) );
                $has_update           = true;
            }

            if ( $request->has_param( 'status' ) ) {
                $data['post_status'] = $this->normalize_product_status( $request->get_param( 'status' ) );
                $has_update          = true;
            }

            if ( $request->has_param( 'slug' ) ) {
                $data['post_name'] = sanitize_title( (string) $request->get_param( 'slug' ) );
                $has_update        = true;
            }

            $should_update_content = $request->has_param( 'content' ) || $request->has_param( 'content_html' ) || $request->has_param( 'content_markdown' );
            if ( $should_update_content ) {
                list( $content_html, $content_markdown ) = $this->prepare_content_for_save( $request );
                $data['post_content'] = $content_html;
                $request->set_param( 'content_markdown', $content_markdown );
                $has_update = true;
            }

            if ( $request->has_param( 'date' ) ) {
                $dates = $this->prepare_post_dates_from_request( $request->get_param( 'date' ) );
                if ( $dates ) {
                    $data       = array_merge( $data, $dates );
                    $has_update = true;
                }
            }

            if ( $has_update ) {
                $result = wp_update_post( $data, true );
                if ( is_wp_error( $result ) ) {
                    return new \WP_REST_Response( [ 'message' => $result->get_error_message() ], 400 );
                }
            }

            try {
                $this->update_product_meta_from_request( $post_id, $request );
            } catch ( \RuntimeException $exception ) {
                return $this->respond_error( 'invalid_product_meta', $exception->getMessage(), 400 );
            }

            if ( $request->has_param( 'is_sticky' ) ) {
                $this->sync_post_sticky_status( $post_id, $request->get_param( 'is_sticky' ) );
            }

            $sku_result = $this->handle_product_skus( $post_id, $request );
            if ( is_wp_error( $sku_result ) ) {
                return $this->error_to_response( $sku_result );
            }

            if ( $request->has_param( 'seo' ) ) {
                $seo_param = $request->get_param( 'seo' );
                if ( is_array( $seo_param ) ) {
                    $seo_payload = $seo_param['payload'] ?? null;
                    if ( null !== $seo_payload ) {
                        $this->sync_product_seo( $post_id, $seo_payload );
                    }
                }
            }

            $this->log_audit(
                'update',
                'product',
                $post_id,
                array_filter(
                    [
                        'title'      => get_the_title( $post_id ),
                        'status'     => get_post_status( $post_id ),
                        'skus_count' => $sku_result['count'] ?? null,
                    ]
                ),
                $request
            );

            return new \WP_REST_Response( $this->build_product_response( $post_id ) );
        }

        public function rest_delete_product( \WP_REST_Request $request ): \WP_REST_Response {
            $post_id = (int) $request['id'];
            $post    = get_post( $post_id );

            if ( ! $post || 'tanz_product' !== $post->post_type ) {
                return $this->respond_error( 'product_not_found', __( '指定的商品不存在。', 'tanzanite-settings' ), 404 );
            }

            wp_delete_post( $post_id, true );
            $this->delete_product_skus( $post_id );

            $this->log_audit(
                'delete',
                'product',
                $post_id,
                [
                    'title'  => $post->post_title,
                    'status' => $post->post_status,
                ],
                $request
            );

            return new \WP_REST_Response( [ 'deleted' => true ] );
        }

        public function rest_bulk_products( \WP_REST_Request $request ): \WP_REST_Response {
            $action = sanitize_key( $request->get_param( 'action' ) );

            if ( ! in_array( $action, self::BULK_PRODUCT_ACTIONS, true ) ) {
                return $this->respond_error( 'invalid_bulk_action', __( '当前批量操作类型不受支持。', 'tanzanite-settings' ) );
            }

            $ids = $this->sanitize_id_list( $request->get_param( 'ids' ) );

            if ( empty( $ids ) ) {
                return $this->respond_error( 'invalid_bulk_payload', __( '请选择至少一个需要处理的商品。', 'tanzanite-settings' ) );
            }

            $payload = $request->get_param( 'payload' );
            if ( ! is_array( $payload ) ) {
                $payload = [];
            }

            $summary = [
                'action'    => $action,
                'total'     => count( $ids ),
                'updated'   => 0,
                'failed'    => [],
                'details'   => [],
                'timestamp' => current_time( 'mysql' ),
            ];

            $target_status   = '';
            $stock_delta     = 0;
            $meta_fields     = [];
            $price_fields    = [];
            $price_mode      = '';
            $price_value     = 0.0;
            $price_round     = 2;
            $featured_enabled = null;
            $featured_slot    = '';
            $delete_mode      = 'trash';

            switch ( $action ) {
                case 'set_status':
                    if ( empty( $payload['status'] ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '批量修改状态需要指定目标状态。', 'tanzanite-settings' ) );
                    }

                    $target_status = sanitize_key( (string) $payload['status'] );

                    if ( ! in_array( $target_status, self::ALLOWED_PRODUCT_STATUSES, true ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '目标状态无效。', 'tanzanite-settings' ) );
                    }
                    break;

                case 'adjust_stock':
                    if ( ! isset( $payload['delta'] ) || ! is_numeric( $payload['delta'] ) || 0 === (int) $payload['delta'] ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '批量调整库存需要提供非零的增量。', 'tanzanite-settings' ) );
                    }

                    $stock_delta = (int) $payload['delta'];
                    break;

                case 'set_meta':
                    $meta_fields = array_intersect_key( $payload, self::PRODUCT_META_MAP );

                    if ( empty( $meta_fields ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '请至少提供一个可更新的商品字段。', 'tanzanite-settings' ) );
                    }
                    break;

                case 'adjust_price':
                    $price_mode = sanitize_key( (string) ( $payload['mode'] ?? '' ) );
                    if ( ! in_array( $price_mode, [ 'absolute', 'percent' ], true ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '请选择有效的调价模式。', 'tanzanite-settings' ) );
                    }

                    if ( ! isset( $payload['value'] ) || ! is_numeric( $payload['value'] ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '调价幅度必须为有效数字。', 'tanzanite-settings' ) );
                    }

                    $price_value = (float) $payload['value'];

                    $allowed_price_fields = [ 'price_regular', 'price_sale', 'price_member' ];
                    $fields_param         = $payload['fields'] ?? [];

                    if ( is_string( $fields_param ) ) {
                        $fields_param = array_filter( explode( ',', $fields_param ) );
                    }

                    if ( ! is_array( $fields_param ) ) {
                        $fields_param = [];
                    }

                    $price_fields = array_values( array_intersect( $allowed_price_fields, array_map( 'sanitize_key', $fields_param ) ) );

                    if ( empty( $price_fields ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '请选择至少一个需要调价的价格字段。', 'tanzanite-settings' ) );
                    }

                    $price_round = isset( $payload['round'] ) && is_numeric( $payload['round'] ) ? max( 0, min( 4, (int) $payload['round'] ) ) : 2;
                    break;

                case 'set_featured':
                    if ( ! array_key_exists( 'enabled', $payload ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '请指定是否设为推荐。', 'tanzanite-settings' ) );
                    }

                    $featured_enabled = (bool) filter_var( $payload['enabled'], FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE );
                    if ( null === $featured_enabled ) {
                        $featured_enabled = (bool) intval( $payload['enabled'] );
                    }

                    $featured_slot = isset( $payload['slot'] ) ? sanitize_text_field( (string) $payload['slot'] ) : '';
                    break;

                case 'delete':
                    $delete_mode = sanitize_key( (string) ( $payload['mode'] ?? 'trash' ) );
                    if ( ! in_array( $delete_mode, [ 'trash', 'force' ], true ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '删除模式无效。', 'tanzanite-settings' ) );
                    }
                    break;

                case 'export':
                    $meta_fields       = [];
                    $stock_delta       = 0;
                    $target_status     = '';
                    $price_fields      = [];
                    $price_mode        = '';
                    $featured_enabled  = null;
                    break;
            }

            $export_rows      = [];
            $export_csv_rows  = [];

            foreach ( $ids as $product_id ) {
                $post = get_post( $product_id );

                if ( ! $post || 'tanz_product' !== $post->post_type ) {
                    $summary['failed'][] = [
                        'id'     => $product_id,
                        'reason' => __( '商品不存在或类型不匹配。', 'tanzanite-settings' ),
                    ];
                    continue;
                }

                switch ( $action ) {
                    case 'set_status':
                        if ( $post->post_status === $target_status ) {
                            $summary['updated']++;
                            $summary['details'][] = [
                                'id'      => $product_id,
                                'status'  => $target_status,
                                'changed' => false,
                            ];
                            continue 2;
                        }

                        $result = wp_update_post(
                            [
                                'ID'          => $product_id,
                                'post_status' => $target_status,
                            ],
                            true
                        );

                        if ( is_wp_error( $result ) ) {
                            $summary['failed'][] = [
                                'id'     => $product_id,
                                'reason' => $result->get_error_message(),
                            ];
                            continue 2;
                        }

                        $summary['updated']++;
                        $summary['details'][] = [
                            'id'      => $product_id,
                            'status'  => $target_status,
                            'changed' => true,
                        ];

                        $this->log_audit(
                            'bulk_set_status',
                            'product',
                            $product_id,
                            [
                                'action' => 'set_status',
                                'status' => $target_status,
                            ],
                            $request
                        );
                        break;

                    case 'adjust_stock':
                        $new_stock = $this->adjust_product_stock( $product_id, $stock_delta );

                        $summary['updated']++;
                        $summary['details'][] = [
                            'id'          => $product_id,
                            'stock_delta' => $stock_delta,
                            'stock_qty'   => $new_stock,
                        ];

                        $this->log_audit(
                            'bulk_adjust_stock',
                            'product',
                            $product_id,
                            [
                                'action'      => 'adjust_stock',
                                'stock_delta' => $stock_delta,
                                'stock_qty'   => $new_stock,
                            ],
                            $request
                        );
                        break;

                    case 'set_meta':
                        $updated_fields = $this->update_product_meta_fields( $product_id, $meta_fields );

                        if ( empty( $updated_fields ) ) {
                            $summary['failed'][] = [
                                'id'     => $product_id,
                                'reason' => __( '未找到可更新的字段。', 'tanzanite-settings' ),
                            ];
                            continue 2;
                        }

                        $summary['updated']++;
                        $summary['details'][] = [
                            'id'     => $product_id,
                            'fields' => array_keys( $updated_fields ),
                        ];

                        $this->log_audit(
                            'bulk_set_meta',
                            'product',
                            $product_id,
                            [
                                'action' => 'set_meta',
                                'fields' => $updated_fields,
                            ],
                            $request
                        );
                        break;

                    case 'adjust_price':
                        $price_updates = [];

                        foreach ( $price_fields as $field ) {
                            if ( ! isset( self::PRODUCT_META_MAP[ $field ] ) ) {
                                continue;
                            }

                            $meta_key    = self::PRODUCT_META_MAP[ $field ]['key'];
                            $current_raw = get_post_meta( $product_id, $meta_key, true );
                            $current     = is_numeric( $current_raw ) ? (float) $current_raw : 0.0;

                            if ( 'percent' === $price_mode ) {
                                $new_value = $current + ( $current * $price_value / 100 );
                            } else {
                                $new_value = $current + $price_value;
                            }

                            if ( $new_value < 0 ) {
                                $new_value = 0.0;
                            }

                            $price_updates[ $field ] = round( $new_value, $price_round );
                        }

                        if ( empty( $price_updates ) ) {
                            $summary['failed'][] = [
                                'id'     => $product_id,
                                'reason' => __( '未能应用调价，请检查选定的字段。', 'tanzanite-settings' ),
                            ];
                            continue 2;
                        }

                        $updated_fields = $this->update_product_meta_fields( $product_id, $price_updates );

                        if ( empty( $updated_fields ) ) {
                            $summary['failed'][] = [
                                'id'     => $product_id,
                                'reason' => __( '调价失败，未更新任何价格字段。', 'tanzanite-settings' ),
                            ];
                            continue 2;
                        }

                        $summary['updated']++;
                        $summary['details'][] = [
                            'id'     => $product_id,
                            'mode'   => $price_mode,
                            'value'  => $price_value,
                            'fields' => $updated_fields,
                        ];

                        $this->log_audit(
                            'bulk_adjust_price',
                            'product',
                            $product_id,
                            [
                                'action' => 'adjust_price',
                                'mode'   => $price_mode,
                                'value'  => $price_value,
                                'fields' => $updated_fields,
                                'round'  => $price_round,
                            ],
                            $request
                        );
                        break;

                    case 'set_featured':
                        $featured_updates = [
                            'featured_flag' => $featured_enabled,
                            'featured_slot' => $featured_enabled ? $featured_slot : '',
                        ];

                        $updated_fields = $this->update_product_meta_fields( $product_id, $featured_updates );

                        if ( empty( $updated_fields ) ) {
                            $summary['failed'][] = [
                                'id'     => $product_id,
                                'reason' => __( '未能更新推荐状态。', 'tanzanite-settings' ),
                            ];
                            continue 2;
                        }

                        $summary['updated']++;
                        $summary['details'][] = [
                            'id'       => $product_id,
                            'featured' => [
                                'enabled' => ! empty( $updated_fields['featured_flag'] ),
                                'slot'    => $updated_fields['featured_slot'] ?? '',
                            ],
                        ];

                        $this->log_audit(
                            'bulk_set_featured',
                            'product',
                            $product_id,
                            [
                                'action'        => 'set_featured',
                                'featured_flag' => $updated_fields['featured_flag'] ?? false,
                                'featured_slot' => $updated_fields['featured_slot'] ?? '',
                            ],
                            $request
                        );
                        break;

                    case 'delete':
                        if ( 'force' === $delete_mode ) {
                            $result = wp_delete_post( $product_id, true );
                            if ( false === $result ) {
                                $summary['failed'][] = [
                                    'id'     => $product_id,
                                    'reason' => __( '永久删除失败。', 'tanzanite-settings' ),
                                ];
                                continue 2;
                            }

                            $this->delete_product_skus( $product_id );
                        } else {
                            $result = wp_trash_post( $product_id );
                            if ( false === $result ) {
                                $summary['failed'][] = [
                                    'id'     => $product_id,
                                    'reason' => __( '移动到回收站失败。', 'tanzanite-settings' ),
                                ];
                                continue 2;
                            }
                        }

                        $summary['updated']++;
                        $summary['details'][] = [
                            'id'   => $product_id,
                            'mode' => $delete_mode,
                        ];

                        $this->log_audit(
                            'bulk_delete',
                            'product',
                            $product_id,
                            [
                                'action' => 'delete',
                                'mode'   => $delete_mode,
                            ],
                            $request
                        );
                        break;

                    case 'export':
                        $product = $this->build_product_response( $product_id );

                        if ( empty( $product ) ) {
                            $summary['failed'][] = [
                                'id'     => $product_id,
                                'reason' => __( '无法加载商品数据。', 'tanzanite-settings' ),
                            ];
                            continue 2;
                        }

                        $export_rows[] = $product;
                        $meta          = $product['meta'] ?? [];

                        $export_csv_rows[] = [
                            'id'            => $product['id'],
                            'title'         => $product['title'],
                            'status'        => $product['status'],
                            'price_regular' => $meta['price_regular'] ?? '',
                            'price_sale'    => $meta['price_sale'] ?? '',
                            'price_member'  => $meta['price_member'] ?? '',
                            'stock_qty'     => $meta['stock_qty'] ?? '',
                            'stock_alert'   => $meta['stock_alert'] ?? '',
                            'points_reward' => $meta['points_reward'] ?? '',
                            'points_limit'  => $meta['points_limit'] ?? '',
                            'updated_at'    => $product['updated_at'],
                        ];
                        break;
                }
            }

            if ( 'export' === $action ) {
                $response_data = [
                    'action'    => $action,
                    'total'     => $summary['total'],
                    'exported'  => count( $export_rows ),
                    'failed'    => $summary['failed'],
                    'timestamp' => $summary['timestamp'],
                    'items'     => $export_rows,
                ];

                if ( ! empty( $export_csv_rows ) ) {
                    $headers = array_keys( $export_csv_rows[0] );
                    $csv     = $this->generate_csv( $headers, $export_csv_rows );

                    $response_data['file'] = [
                        'name'    => 'tanzanite-products-' . gmdate( 'YmdHis' ) . '.csv',
                        'mime'    => 'text/csv',
                        'content' => base64_encode( $csv ),
                    ];
                }

                $status = empty( $summary['failed'] ) ? 200 : ( count( $export_rows ) ? 207 : 400 );

                if ( 400 === $status ) {
                    $response = $this->respond_error( 'partial_bulk_failure', __( '导出失败，所有商品均未处理成功。', 'tanzanite-settings' ) );
                    $data     = $response->get_data();
                    $data['failed']    = $summary['failed'];
                    $data['action']    = $action;
                    $data['total']     = $summary['total'];
                    $data['timestamp'] = $summary['timestamp'];
                    $response->set_data( $data );

                    return $response;
                }

                return new \WP_REST_Response( $response_data, $status );
            }

            if ( 0 === $summary['updated'] ) {
                $response = $this->respond_error( 'partial_bulk_failure', __( '批量操作未成功，所有商品均处理失败。', 'tanzanite-settings' ) );
                $data     = $response->get_data();
                $data['failed']    = $summary['failed'];
                $data['action']    = $action;
                $data['total']     = $summary['total'];
                $data['timestamp'] = $summary['timestamp'];
                $response->set_data( $data );

                return $response;
            }

            $status = empty( $summary['failed'] ) ? 200 : 207;

            return new \WP_REST_Response( $summary, $status );
        }

        public function rest_preview_sku_import( \WP_REST_Request $request ): \WP_REST_Response {
            $product_id = (int) $request->get_param( 'product_id' );
            $product    = get_post( $product_id );

            if ( ! $product || 'tanz_product' !== $product->post_type ) {
                return $this->respond_error( 'product_not_found', __( '指定的商品不存在。', 'tanzanite-settings' ), 404 );
            }

            $bulk = (string) $request->get_param( 'skus_bulk' );
            $parsed = $this->parse_bulk_skus( $bulk );
            if ( is_wp_error( $parsed ) ) {
                return $this->error_to_response( $parsed );
            }

            $defaults = $this->get_product_meta_payload( $product_id );
            $sanitized = $this->sanitize_skus(
                $parsed,
                [
                    'price_regular' => $defaults['price_regular'] ?? 0,
                    'price_sale'    => $defaults['price_sale'] ?? 0,
                    'stock_qty'     => $defaults['stock_qty'] ?? 0,
                ]
            );

            if ( is_wp_error( $sanitized ) ) {
                return $this->error_to_response( $sanitized );
            }

            $existing_codes = $this->get_product_sku_codes( $product_id );
            $new_codes      = array_map( static fn( array $row ): string => (string) $row['sku_code'], $sanitized );
            $duplicates     = array_values( array_intersect( $existing_codes, $new_codes ) );

            $summary = [
                'product_title'       => get_the_title( $product_id ),
                'message'             => sprintf( __( '预检成功：共有 %d 条 SKU 将被导入。', 'tanzanite-settings' ), count( $sanitized ) ),
                'duplicates'          => $duplicates,
                'duplicates_message'  => $duplicates
                    ? __( '以下 SKU 将覆盖已有记录：', 'tanzanite-settings' )
                    : __( '未检测到已存在的 SKU 编码。', 'tanzanite-settings' ),
                'existing_total'      => count( $existing_codes ),
            ];

            return new \WP_REST_Response(
                [
                    'skus'    => $sanitized,
                    'summary' => $summary,
                ]
            );
        }

        public function rest_apply_sku_import( \WP_REST_Request $request ): \WP_REST_Response {
            $product_id = (int) $request->get_param( 'product_id' );
            $product    = get_post( $product_id );

            if ( ! $product || 'tanz_product' !== $product->post_type ) {
                return $this->respond_error( 'product_not_found', __( '指定的商品不存在。', 'tanzanite-settings' ), 404 );
            }

            $skus_payload = $request->get_param( 'skus' );
            if ( ! is_array( $skus_payload ) ) {
                return $this->respond_error( 'invalid_sku_payload', __( 'SKU 数据格式不正确。', 'tanzanite-settings' ) );
            }

            $defaults = $this->get_product_meta_payload( $product_id );
            $sanitized = $this->sanitize_skus(
                $skus_payload,
                [
                    'price_regular' => $defaults['price_regular'] ?? 0,
                    'price_sale'    => $defaults['price_sale'] ?? 0,
                    'stock_qty'     => $defaults['stock_qty'] ?? 0,
                ]
            );

            if ( is_wp_error( $sanitized ) ) {
                return $this->error_to_response( $sanitized );
            }

            if ( empty( $sanitized ) ) {
                $this->delete_product_skus( $product_id );

                $this->log_audit(
                    'import_sku',
                    'product',
                    $product_id,
                    [ 'count' => 0 ],
                    $request
                );

                return new \WP_REST_Response(
                    [
                        'updated' => 0,
                        'message' => __( '已清空该商品的所有 SKU。', 'tanzanite-settings' ),
                    ]
                );
            }

            if ( ! $this->persist_product_skus( $product_id, $sanitized ) ) {
                return $this->respond_error( 'failed_persist_skus', __( '保存 SKU 时发生异常，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $this->log_audit(
                'import_sku',
                'product',
                $product_id,
                [ 'count' => count( $sanitized ) ],
                $request
            );

            return new \WP_REST_Response(
                [
                    'updated' => count( $sanitized ),
                    'message' => sprintf( __( '已成功导入 %d 条 SKU。', 'tanzanite-settings' ), count( $sanitized ) ),
                ]
            );
        }

        /**
         * REST：订单列表。
         */
        public function rest_list_orders( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $page = max( 1, (int) $request->get_param( 'page' ) );

            $per_page = (int) $request->get_param( 'per_page' );
            if ( $per_page <= 0 ) {
                $per_page = 20;
            }
            $per_page = min( 100, $per_page );

            $offset = ( $page - 1 ) * $per_page;

            $where  = [ '1=1' ];
            $params = [];
            $joins  = '';

            $status = $request->get_param( 'status' );
            if ( $status ) {
                $status_key = sanitize_key( $status );
                if ( in_array( $status_key, self::ALLOWED_ORDER_STATUSES, true ) ) {
                    $where[]  = 'o.status = %s';
                    $params[] = $status_key;
                }
            }

            $channel = $request->get_param( 'channel' );
            if ( $channel ) {
                $where[]  = 'o.channel = %s';
                $params[] = sanitize_text_field( $channel );
            }

            $payment_method = $request->get_param( 'payment_method' );
            if ( $payment_method ) {
                $where[]  = 'o.payment_method = %s';
                $params[] = sanitize_text_field( $payment_method );
            }

            $tracking_provider = $request->get_param( 'tracking_provider' );
            if ( $tracking_provider ) {
                $where[]  = 'o.tracking_provider = %s';
                $params[] = sanitize_key( $tracking_provider );
            }

            $date_start = $request->get_param( 'date_start' );
            if ( $date_start ) {
                $start_time = strtotime( sanitize_text_field( $date_start ) );
                if ( $start_time ) {
                    $where[]  = 'o.created_at >= %s';
                    $params[] = gmdate( 'Y-m-d H:i:s', $start_time );
                }
            }

            $date_end = $request->get_param( 'date_end' );
            if ( $date_end ) {
                $end_time = strtotime( sanitize_text_field( $date_end ) );
                if ( $end_time ) {
                    $end_time = strtotime( '+1 day', $end_time ) - 1;
                    $where[]  = 'o.created_at <= %s';
                    $params[] = gmdate( 'Y-m-d H:i:s', $end_time );
                }
            }

            $customer_keyword = $request->get_param( 'customer_keyword' );
            if ( $customer_keyword ) {
                $joins   .= " LEFT JOIN {$wpdb->users} u ON u.ID = o.user_id";
                $like     = '%' . $wpdb->esc_like( $customer_keyword ) . '%';
                $where[]  = '(u.display_name LIKE %s OR u.user_email LIKE %s OR u.user_login LIKE %s)';
                $params[] = $like;
                $params[] = $like;
                $params[] = $like;
            }

            $base_sql  = "FROM {$this->orders_table} o{$joins}";
            $where_sql = $where ? 'WHERE ' . implode( ' AND ', $where ) : '';

            if ( $params ) {
                $total = (int) $wpdb->get_var( $wpdb->prepare( "SELECT COUNT(*) {$base_sql} {$where_sql}", ...$params ) );
            } else {
                $total = (int) $wpdb->get_var( "SELECT COUNT(*) {$base_sql} {$where_sql}" );
            }

            $data_sql     = "SELECT o.* {$base_sql} {$where_sql} ORDER BY o.id DESC LIMIT %d OFFSET %d";
            $data_params  = array_merge( $params, [ $per_page, $offset ] );
            $prepared_sql = $wpdb->prepare( $data_sql, ...$data_params );
            $rows         = $wpdb->get_results( $prepared_sql, ARRAY_A );

            $items = array_map( [ $this, 'build_order_response' ], $rows );

            return new \WP_REST_Response(
                [
                    'items' => $items,
                    'meta'  => [
                        'page'        => $page,
                        'per_page'    => $per_page,
                        'total'       => $total,
                        'total_pages' => $per_page ? (int) ceil( $total / $per_page ) : 0,
                    ],
                ]
            );
        }

        /**
         * REST：创建占位订单。
         */
        public function rest_create_order( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $tracking_provider_raw = $request->get_param( 'tracking_provider' );
            $tracking_number_raw   = $request->get_param( 'tracking_number' );

            $tracking_provider = $tracking_provider_raw ? sanitize_key( (string) $tracking_provider_raw ) : '';
            $tracking_number   = $tracking_number_raw ? sanitize_text_field( (string) $tracking_number_raw ) : '';

            $data = [
                'order_number'      => $this->generate_order_number(),
                'user_id'           => get_current_user_id(),
                'status'            => $this->ensure_order_status( $request->get_param( 'status' ) ),
                'payment_method'    => $request->get_param( 'payment_method' ),
                'channel'           => $request->get_param( 'channel' ),
                'total'             => (float) $request->get_param( 'total' ),
                'subtotal'          => (float) $request->get_param( 'subtotal' ),
                'discount_total'    => (float) $request->get_param( 'discount_total' ),
                'shipping_total'    => (float) $request->get_param( 'shipping_total' ),
                'points_used'       => 0,
                'currency'          => 'CNY',
                'tracking_provider' => $tracking_provider,
                'tracking_number'   => $tracking_number,
                'meta'              => wp_json_encode( [ 'stage' => 'placeholder' ], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES ),
            ];

            $format_map = [
                '%s', '%d', '%s', '%s', '%s', '%f', '%f', '%f', '%f', '%d', '%s', '%s', '%s', '%s'
            ];

            $items = $this->sanitize_order_items( $request->get_param( 'items' ) );
            if ( is_wp_error( $items ) ) {
                return $this->error_to_response( $items );
            }

            $this->apply_status_timestamps( $data, $format_map, null, $data['status'] );

            $inserted = $wpdb->insert( $this->orders_table, $data, $format_map );

            if ( false === $inserted ) {
                return $this->respond_error( 'failed_create_order', __( '创建订单失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $order_id = (int) $wpdb->insert_id;

            if ( ! $this->persist_order_items( $order_id, $items ) ) {
                $wpdb->delete( $this->orders_table, [ 'id' => $order_id ], [ '%d' ] );
                return $this->respond_error( 'failed_create_order', __( '保存订单明细失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $tracking_sync_result = null;
            if ( 'shipped' === $data['status'] && $tracking_provider && $tracking_number ) {
                $tracking_sync_result = $this->maybe_sync_order_tracking( $order_id, 'order_create' );
            }

            $row = $this->fetch_order_row( $order_id );

            if ( null !== $tracking_sync_result && ! is_wp_error( $tracking_sync_result ) ) {
                $row = $this->fetch_order_row( $order_id );
            }

            $this->log_audit(
                'create',
                'order',
                $order_id,
                [
                    'status' => $row['status'],
                    'items'  => count( $items ),
                ],
                $request
            );

            $response_payload = $this->build_order_response( $row );

            if ( null !== $tracking_sync_result ) {
                $response_payload['tracking_meta'] = [
                    'sync_status'  => is_wp_error( $tracking_sync_result ) ? 'error' : 'success',
                    'synced_at'    => is_wp_error( $tracking_sync_result ) ? null : ( $tracking_sync_result['synced_at'] ?? null ),
                    'error_code'   => is_wp_error( $tracking_sync_result ) ? $tracking_sync_result->get_error_code() : null,
                    'error_message'=> is_wp_error( $tracking_sync_result ) ? $tracking_sync_result->get_error_message() : null,
                ];
            }

            return new \WP_REST_Response( $this->build_order_response( $this->fetch_order_row( $order_id ), 'detail' ), 201 );
        }

        public function rest_get_order( \WP_REST_Request $request ): \WP_REST_Response {
            $row = $this->fetch_order_row( (int) $request['id'] );

            if ( ! $row ) {
                return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
            }

            return new \WP_REST_Response( $this->build_order_response( $row, 'detail' ) );
        }

        public function rest_update_order( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $row = $this->fetch_order_row( (int) $request['id'] );
            if ( ! $row ) {
                return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
            }

            $data  = [];
            $types = [];

            $requested_status = $row['status'];

            if ( $request->has_param( 'status' ) ) {
                $requested_status = $this->ensure_order_status( $request->get_param( 'status' ) );

                if ( ! in_array( $requested_status, self::ORDER_STATUS_TRANSITIONS[ $row['status'] ] ?? [], true ) ) {
                    return $this->respond_error( 'invalid_order_status', __( '此订单状态无法切换到目标状态。', 'tanzanite-settings' ) );
                }

                $data['status'] = $requested_status;
                $types[]        = '%s';
            }

            foreach ( [ 'payment_method', 'channel' ] as $field ) {
                if ( $request->has_param( $field ) ) {
                    $data[ $field ] = sanitize_text_field( $request->get_param( $field ) );
                    $types[]        = '%s';
                }
            }
            foreach ( [ 'total', 'subtotal', 'discount_total', 'shipping_total' ] as $field ) {
                if ( $request->has_param( $field ) ) {
                    $data[ $field ] = (float) $request->get_param( $field );
                    $types[]        = '%f';
                }
            }

            $should_sync_tracking = false;

            if ( $request->has_param( 'tracking_provider' ) ) {
                $provider              = sanitize_key( $request->get_param( 'tracking_provider' ) );
                $data['tracking_provider'] = $provider;
                $types[]               = '%s';
                $should_sync_tracking  = true;
            }

            if ( $request->has_param( 'tracking_number' ) ) {
                $tracking_number       = sanitize_text_field( $request->get_param( 'tracking_number' ) );
                $data['tracking_number'] = $tracking_number;
                $types[]               = '%s';
                $should_sync_tracking  = true;
            }

            $this->apply_status_timestamps( $data, $types, $row, $requested_status );

            if ( ! empty( $data ) ) {
                $update_result = $wpdb->update( $this->orders_table, $data, [ 'id' => (int) $request['id'] ], $types, [ '%d' ] );
                if ( false === $update_result ) {
                    return $this->respond_error( 'failed_update_order', __( '更新订单失败，请稍后重试。', 'tanzanite-settings' ), 500 );
                }
            }

            if ( $request->has_param( 'items' ) ) {
                $items = $this->sanitize_order_items( $request->get_param( 'items' ) );
                if ( is_wp_error( $items ) ) {
                    return $this->error_to_response( $items );
                }

                if ( ! $this->persist_order_items( (int) $request['id'], $items ) ) {
                    return $this->respond_error( 'failed_update_order', __( '更新订单明细失败，请稍后重试。', 'tanzanite-settings' ), 500 );
                }
            }

            $row = $this->fetch_order_row( (int) $request['id'] );

            $tracking_sync_result = null;
            if ( $should_sync_tracking ) {
                if ( empty( $row['tracking_provider'] ) || empty( $row['tracking_number'] ) ) {
                    $this->clear_tracking_events( (int) $request['id'] );
                    $tracking_sync_result = [];
                } else {
                    $tracking_sync_result = $this->maybe_sync_order_tracking( (int) $request['id'], 'order_update' );
                }
            } elseif ( 'shipped' === $row['status'] && empty( $row['tracking_synced_at'] ) && ! empty( $row['tracking_provider'] ) && ! empty( $row['tracking_number'] ) ) {
                $tracking_sync_result = $this->maybe_sync_order_tracking( (int) $request['id'], 'status_update' );
            }

            $this->log_audit(
                'update',
                'order',
                (int) $request['id'],
                [
                    'status' => $row['status'],
                    'items'  => count( $this->fetch_order_items( (int) $request['id'] ) ),
                ],
                $request
            );

            $response_payload = $this->build_order_response( $row, 'detail' );

            if ( null !== $tracking_sync_result ) {
                if ( is_wp_error( $tracking_sync_result ) ) {
                    $response_payload['tracking_meta']['sync_error'] = $tracking_sync_result->get_error_message();
                } else {
                    $response_payload['tracking_meta']['synced_at'] = $tracking_sync_result['synced_at'] ?? null;
                }
            }

            return new \WP_REST_Response( $response_payload );
        }

        public function rest_sync_order_tracking( \WP_REST_Request $request ): \WP_REST_Response {
            $order_id = (int) $request['id'];

            $row = $this->fetch_order_row( $order_id );
            if ( ! $row ) {
                return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
            }

            $result = $this->maybe_sync_order_tracking( $order_id, 'manual_rest' );

            if ( is_wp_error( $result ) ) {
                return $this->respond_error( $result->get_error_code(), $result->get_error_message() );
            }

            $row = $this->fetch_order_row( $order_id );

            $response_payload = $this->build_order_response( $row, 'detail' );
            $response_payload['tracking_meta']['synced_at'] = $result['synced_at'] ?? null;

            return new \WP_REST_Response( $response_payload );
        }

        public function rest_delete_order( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $row = $this->fetch_order_row( (int) $request['id'] );
            if ( ! $row ) {
                return $this->respond_error( 'order_not_found', __( '指定的订单不存在。', 'tanzanite-settings' ), 404 );
            }

            $deleted = $wpdb->delete( $this->orders_table, [ 'id' => (int) $request['id'] ], [ '%d' ] );
            if ( false === $deleted ) {
                return $this->respond_error( 'failed_delete_order', __( '删除订单失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $this->delete_order_items( (int) $request['id'] );

            $this->log_audit(
                'delete',
                'order',
                (int) $request['id'],
                [
                    'status' => $row['status'],
                ],
                $request
            );

            return new \WP_REST_Response( [ 'deleted' => true ] );
        }

        public function rest_bulk_orders( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $action = sanitize_key( $request->get_param( 'action' ) );

            if ( ! in_array( $action, self::BULK_ORDER_ACTIONS, true ) ) {
                return $this->respond_error( 'invalid_bulk_action', __( '当前批量操作类型不受支持。', 'tanzanite-settings' ) );
            }

            $ids = $this->sanitize_id_list( $request->get_param( 'ids' ) );

            if ( empty( $ids ) ) {
                return $this->respond_error( 'invalid_bulk_payload', __( '请选择至少一个需要处理的订单。', 'tanzanite-settings' ) );
            }

            $payload = $request->get_param( 'payload' );
            if ( ! is_array( $payload ) ) {
                $payload = [];
            }

            $summary = [
                'action'    => $action,
                'total'     => count( $ids ),
                'updated'   => 0,
                'failed'    => [],
                'details'   => [],
                'timestamp' => current_time( 'mysql' ),
            ];

            switch ( $action ) {
                case 'set_status':
                    if ( empty( $payload['status'] ) ) {
                        return $this->respond_error( 'invalid_bulk_payload', __( '批量修改状态需要指定目标状态。', 'tanzanite-settings' ) );
                    }

                    $target_status = $this->ensure_order_status( (string) $payload['status'] );
                    break;

                case 'export':
                    $target_status = '';
                    break;
            }

            $export_rows     = [];
            $export_csv_rows = [];

            foreach ( $ids as $order_id ) {
                $row = $this->fetch_order_row( $order_id );

                if ( ! $row ) {
                    $summary['failed'][] = [
                        'id'     => $order_id,
                        'reason' => __( '订单不存在。', 'tanzanite-settings' ),
                    ];
                    continue;
                }

                switch ( $action ) {
                    case 'set_status':
                        $current_status = $row['status'];

                        if ( $current_status === $target_status ) {
                            $summary['updated']++;
                            $summary['details'][] = [
                                'id'      => $order_id,
                                'status'  => $target_status,
                                'changed' => false,
                            ];
                            continue 2;
                        }

                        $allowed = self::ORDER_STATUS_TRANSITIONS[ $current_status ] ?? [];
                        if ( ! in_array( $target_status, $allowed, true ) ) {
                            $summary['failed'][] = [
                                'id'     => $order_id,
                                'reason' => __( '目标状态与当前状态不兼容。', 'tanzanite-settings' ),
                            ];
                            continue 2;
                        }

                        $data   = [ 'status' => $target_status ];
                        $types  = [ '%s' ];

                        $this->apply_status_timestamps( $data, $types, $row, $target_status );

                        $updated = $wpdb->update( $this->orders_table, $data, [ 'id' => $order_id ], $types, [ '%d' ] );

                        if ( false === $updated ) {
                            $summary['failed'][] = [
                                'id'     => $order_id,
                                'reason' => __( '数据库更新失败。', 'tanzanite-settings' ),
                            ];
                            continue 2;
                        }

                        $summary['updated']++;
                        $summary['details'][] = [
                            'id'      => $order_id,
                            'from'    => $current_status,
                            'status'  => $target_status,
                            'changed' => true,
                        ];

                        $this->log_audit(
                            'bulk_set_status',
                            'order',
                            $order_id,
                            [
                                'action' => 'set_status',
                                'from'   => $current_status,
                                'to'     => $target_status,
                            ],
                            $request
                        );
                        break;

                    case 'export':
                        $export_rows[] = $this->build_order_response( $row );

                        $export_csv_rows[] = [
                            'id'             => (int) $row['id'],
                            'order_number'   => $row['order_number'],
                            'status'         => $row['status'],
                            'total'          => (float) $row['total'],
                            'subtotal'       => (float) $row['subtotal'],
                            'discount_total' => (float) $row['discount_total'],
                            'shipping_total' => (float) $row['shipping_total'],
                            'payment_method' => $row['payment_method'],
                            'channel'        => $row['channel'],
                            'created_at'     => $row['created_at'],
                            'updated_at'     => $row['updated_at'],
                        ];
                        break;
                }
            }

            if ( 'export' === $action ) {
                $response_data = [
                    'action'    => $action,
                    'total'     => $summary['total'],
                    'exported'  => count( $export_rows ),
                    'failed'    => $summary['failed'],
                    'timestamp' => $summary['timestamp'],
                    'items'     => $export_rows,
                ];

                if ( ! empty( $export_csv_rows ) ) {
                    $headers = array_keys( $export_csv_rows[0] );
                    $csv     = $this->generate_csv( $headers, $export_csv_rows );

                    $response_data['file'] = [
                        'name'    => 'tanzanite-orders-' . gmdate( 'YmdHis' ) . '.csv',
                        'mime'    => 'text/csv',
                        'content' => base64_encode( $csv ),
                    ];
                }

                $status = empty( $summary['failed'] ) ? 200 : ( count( $export_rows ) ? 207 : 400 );

                if ( 400 === $status ) {
                    $response = $this->respond_error( 'partial_bulk_failure', __( '导出失败，所有订单均未处理成功。', 'tanzanite-settings' ) );
                    $data     = $response->get_data();
                    $data['failed']    = $summary['failed'];
                    $data['action']    = $action;
                    $data['total']     = $summary['total'];
                    $data['timestamp'] = $summary['timestamp'];
                    $response->set_data( $data );

                    return $response;
                }

                return new \WP_REST_Response( $response_data, $status );
            }

            if ( 0 === $summary['updated'] ) {
                $response = $this->respond_error( 'partial_bulk_failure', __( '批量操作未成功，所有订单均处理失败。', 'tanzanite-settings' ) );
                $data     = $response->get_data();
                $data['failed']    = $summary['failed'];
                $data['action']    = $action;
                $data['total']     = $summary['total'];
                $data['timestamp'] = $summary['timestamp'];
                $response->set_data( $data );

                return $response;
            }

            $status = empty( $summary['failed'] ) ? 200 : 207;

            return new \WP_REST_Response( $summary, $status );
        }

        // rest_create_shipping_template -> Tanzanite_REST_ShippingTemplates_Controller::create_item()
        // rest_get_shipping_template -> Tanzanite_REST_ShippingTemplates_Controller::get_item()
        // rest_update_shipping_template -> Tanzanite_REST_ShippingTemplates_Controller::update_item()
        // rest_delete_shipping_template -> Tanzanite_REST_ShippingTemplates_Controller::delete_item()

        /**
         * 生成占位订单编号。
         */
        private function generate_order_number(): string {
            return strtoupper( wp_generate_password( 4, false, false ) ) . '-' . wp_unique_id( 'TZ' );
        }

        /**
         * 注册 Tanzanite Settings 后台菜单。
         */
        public function register_admin_menu(): void {
            error_log('=== Tanzanite Settings: register_admin_menu() called ===');
            $root_capability = 'tanz_view_products';
            $root_slug       = 'tanzanite-settings';

            add_menu_page(
                __( 'Tanzanite Settings', 'tanzanite-settings' ),
                __( 'Tanzanite Settings', 'tanzanite-settings' ),
                $root_capability,
                $root_slug,
                [ $this, 'render_all_products' ],
                'dashicons-admin-generic',
                56
            );

            add_submenu_page(
                $root_slug,
                __( 'All Products', 'tanzanite-settings' ),
                __( 'All Products', 'tanzanite-settings' ),
                'tanz_view_products',
                $root_slug,
                [ $this, 'render_all_products' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Attributes', 'tanzanite-settings' ),
                __( 'Attributes', 'tanzanite-settings' ),
                'tanz_manage_products',
                'tanzanite-settings-attributes',
                [ $this, 'render_attributes' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Reviews', 'tanzanite-settings' ),
                __( 'Reviews', 'tanzanite-settings' ),
                'tanz_manage_products',
                'tanzanite-settings-reviews',
                [ $this, 'render_reviews' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Add New Product', 'tanzanite-settings' ),
                __( 'Add New Product', 'tanzanite-settings' ),
                'tanz_manage_products',
                'tanzanite-settings-add-product',
                [ $this, 'render_add_product' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Payment Method', 'tanzanite-settings' ),
                __( 'Payment Method', 'tanzanite-settings' ),
                'tanz_manage_payments',
                'tanzanite-settings-payment-method',
                [ $this, 'render_payment_method' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Tax Rates', 'tanzanite-settings' ),
                __( 'Tax Rates', 'tanzanite-settings' ),
                'tanz_manage_products',
                'tanzanite-settings-tax-rates',
                [ $this, 'render_tax_rates' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'All Orders', 'tanzanite-settings' ),
                __( 'All Orders', 'tanzanite-settings' ),
                'tanz_view_orders',
                'tanzanite-settings-orders',
                [ $this, 'render_orders_list' ]
            );

            add_submenu_page(
                null,
                __( 'Order Detail', 'tanzanite-settings' ),
                __( 'Order Detail', 'tanzanite-settings' ),
                'tanz_view_orders',
                'tanzanite-settings-order-detail',
                [ $this, 'render_order_detail_page' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Order Bulk', 'tanzanite-settings' ),
                __( 'Order Bulk', 'tanzanite-settings' ),
                'tanz_bulk_orders',
                'tanzanite-settings-orders-bulk',
                [ $this, 'render_orders_bulk' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Shipping Templates', 'tanzanite-settings' ),
                __( 'Shipping Templates', 'tanzanite-settings' ),
                'tanz_manage_shipping',
                'tanzanite-settings-shipping-templates',
                [ $this, 'render_shipping_templates' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Carriers & Tracking', 'tanzanite-settings' ),
                __( 'Carriers & Tracking', 'tanzanite-settings' ),
                'tanz_manage_shipping',
                'tanzanite-settings-carriers',
                [ $this, 'render_carriers' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Tracking Providers', 'tanzanite-settings' ),
                __( 'Tracking Providers', 'tanzanite-settings' ),
                'tanz_manage_shipping',
                'tanzanite-settings-tracking',
                [ $this, 'render_tracking_settings' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'SKU Importer', 'tanzanite-settings' ),
                __( 'SKU Importer', 'tanzanite-settings' ),
                'tanz_manage_products',
                'tanzanite-settings-sku-importer',
                [ $this, 'render_sku_importer' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Audit Logs', 'tanzanite-settings' ),
                __( 'Audit Logs', 'tanzanite-settings' ),
                'tanz_view_audit_logs',
                'tanzanite-settings-audit-logs',
                [ $this, 'render_audit_logs' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Loyalty Settings', 'tanzanite-settings' ),
                __( 'Loyalty & Points', 'tanzanite-settings' ),
                'manage_options',
                'tanzanite-settings-loyalty',
                [ $this, 'render_loyalty_settings' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Member Profiles', 'tanzanite-settings' ),
                __( 'Member Profiles', 'tanzanite-settings' ),
                'manage_options',
                'tanzanite-settings-members',
                [ $this, 'render_member_profiles' ]
            );

            add_submenu_page(
                $root_slug,
                __( 'Gift Cards & Coupons', 'tanzanite-settings' ),
                __( 'Gift Cards & Coupons', 'tanzanite-settings' ),
                'manage_options',
                'tanzanite-settings-rewards',
                [ $this, 'render_rewards' ]
            );

            // Customer Service
            add_submenu_page(
                $root_slug,
                __( 'Customer Service', 'tanzanite-settings' ),
                __( 'Customer Service', 'tanzanite-settings' ),
                'manage_options',
                'tanzanite-settings-customer-service',
                array( $this, 'render_customer_service' )
            );

            // MyTheme SEO 集成
            // 调试：确认菜单注册
            error_log('Tanzanite Settings: Registering SEO menu');
            $seo_menu_result = add_submenu_page(
                $root_slug,
                __( 'SEO Settings', 'tanzanite-settings' ),
                __( 'SEO Settings', 'tanzanite-settings' ),
                'manage_options',
                'tanzanite-settings-seo',
                [ $this, 'render_seo_page' ]
            );
            error_log('Tanzanite Settings: SEO menu registered, result = ' . var_export($seo_menu_result, true));
            error_log('Tanzanite Settings: SEO menu callback = ' . (method_exists($this, 'render_seo_page') ? 'EXISTS' : 'NOT EXISTS'));
        }

        /**
         * 按需加载后台样式表和脚本。
         */
        public function enqueue_admin_assets(): void {
            $screen = get_current_screen();
            if ( ! $screen || false === strpos( $screen->id, 'tanzanite-settings' ) ) {
                return;
            }

            // 加载样式
            // 注意：admin.css 和 admin.min.css 内容相同，保留两个文件是为了符合 WordPress 标准
            // 生产环境自动加载 .min.css，开发环境（SCRIPT_DEBUG=true）加载 .css
            $suffix = defined( 'SCRIPT_DEBUG' ) && SCRIPT_DEBUG ? '' : '.min';
            wp_enqueue_style(
                'tanzanite-settings-admin',
                plugins_url( 'assets/css/admin' . $suffix . '.css', TANZANITE_LEGACY_MAIN_FILE ),
                [],
                self::VERSION
            );

            // 注册公共 JS 库
            wp_register_script(
                'tz-admin-common',
                plugins_url( 'assets/js/admin-common.js', TANZANITE_PLUGIN_FILE ),
                array( 'jquery' ),
                self::VERSION,
                true
            );

            // 根据页面加载对应的 JS 文件
            $screen = get_current_screen();
            
            // Attributes 页面 - 脚本在 render_attributes() 中加载
            if ( $screen && strpos( $screen->id, 'tanzanite-settings-attributes' ) !== false ) {
                wp_enqueue_media();
                
                wp_enqueue_script(
                    'tz-attributes',
                    TANZANITE_PLUGIN_URL . 'assets/js/attributes.js',
                    array( 'jquery', 'wp-media' ),
                    self::VERSION,
                    true
                );
            }
        }

        /**
         * 在插件页面添加自定义 body class 以便样式隔离。
         */
        public function filter_admin_body_class( string $classes ): string {
            $screen = get_current_screen();
            if ( $screen && ( false !== strpos( $screen->id, 'tanzanite-settings' ) || false !== strpos( $screen->id, 'tanzanite-cart' ) ) ) {
                $classes .= ' tz-settings-admin';
            }

            return $classes;
        }

        /**
         * 统一的占位页面渲染逻辑。
         *
         * @param string $title       页面标题。
         * @param string $description 简要说明。
         * @param array  $checklist   后续要实现的关键点列表。
         */
        private function render_placeholder_page( string $title, string $description, array $checklist = [] ): void {
            echo '<div class="tz-settings-wrapper">';

            echo '<div class="tz-settings-header">';
            echo '<h1>' . esc_html( $title ) . '</h1>';
            if ( ! empty( $description ) ) {
                echo '<p>' . esc_html( $description ) . '</p>';
            }
            echo '</div>';

            if ( ! empty( $checklist ) ) {
                echo '<div class="tz-settings-section">';
                echo '<div class="tz-section-title">' . esc_html__( '阶段要点', 'tanzanite-settings' ) . '</div>';
                echo '<ul class="tz-section-list">';
                foreach ( $checklist as $item ) {
                    echo '<li>' . esc_html( $item ) . '</li>';
                }
                echo '</ul>';
                echo '</div>';
            }

            echo '<div class="tz-settings-placeholder">';
            echo '<span class="tz-placeholder-tag">' . esc_html__( 'Stage 1', 'tanzanite-settings' ) . '</span>';
            echo '<p>' . esc_html__( '此页面当前为布局占位，后续迭代将逐步填充功能组件。', 'tanzanite-settings' ) . '</p>';
            echo '</div>';

            echo '</div>';
        }

        public function render_all_products(): void {
            if ( ! current_user_can( 'tanz_view_products' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            $nonce        = wp_create_nonce( 'wp_rest' );
            $list_endpoint = esc_url_raw( rest_url( 'tanzanite/v1/products' ) );
            $single_link   = esc_url( admin_url( 'admin.php?page=tanzanite-settings-add-product&product_id=' ) );
            $seo_link      = esc_url( admin_url( 'admin.php?page=tanzanite-settings-add-product&focus=seo&product_id=' ) );
            $bulk_link     = esc_url( admin_url( 'admin.php?page=tanzanite-settings-sku-importer' ) );

            $can_manage = current_user_can( 'tanz_manage_products' );
            $can_bulk   = current_user_can( 'tanz_bulk_products' );

            // 加载 Products List 模块（按依赖顺序）
            // 1. 渲染模块（被 core 依赖）
            wp_enqueue_script(
                'tz-products-list-render',
                plugins_url( 'assets/js/products-list-render.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 2. 筛选模块（被 core 依赖）
            wp_enqueue_script(
                'tz-products-list-filters',
                plugins_url( 'assets/js/products-list-filters.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 3. 批量操作模块（被 core 依赖）
            wp_enqueue_script(
                'tz-products-list-bulk',
                plugins_url( 'assets/js/products-list-bulk.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 4. 单个操作模块（被 core 依赖）
            wp_enqueue_script(
                'tz-products-list-actions',
                plugins_url( 'assets/js/products-list-actions.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 5. 核心模块（依赖所有其他模块）
            wp_enqueue_script(
                'tz-products-list-core',
                plugins_url( 'assets/js/products-list-core.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-products-list-render', 'tz-products-list-filters', 'tz-products-list-bulk', 'tz-products-list-actions' ),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-products-list-core',
                'TzProductsListConfig',
                array(
                    'nonce'            => $nonce,
                    'listUrl'          => $list_endpoint,
                    'singleUrl'        => $list_endpoint . '/',
                    'bulkUrl'          => $list_endpoint,
                    'editUrl'          => $single_link,
                    'seoUrl'           => $seo_link,
                    'canManage'        => $can_manage,
                    'canBulk'          => $can_bulk,
                    'categoryEndpoint' => esc_url_raw( rest_url( 'tanzanite/v1/categories' ) ),
                    'tagsEndpoint'     => esc_url_raw( rest_url( 'tanzanite/v1/tags' ) ),
                    'strings'          => array(
                        'noData'                  => __( '暂无数据', 'tanzanite-settings' ),
                        'loadFailed'              => __( '加载失败', 'tanzanite-settings' ),
                        'pageTemplate'            => __( '第 {page}/{pages} 页', 'tanzanite-settings' ),
                        'expandFilters'           => __( '展开筛选', 'tanzanite-settings' ),
                        'collapseFilters'         => __( '收起筛选', 'tanzanite-settings' ),
                        'editLabel'               => __( '编辑', 'tanzanite-settings' ),
                        'seoLabel'                => __( 'SEO', 'tanzanite-settings' ),
                        'previewLabel'            => __( '预览', 'tanzanite-settings' ),
                        'copyPayloadLabel'        => __( '复制', 'tanzanite-settings' ),
                        'deleteLabel'             => __( '删除', 'tanzanite-settings' ),
                        'stickLabel'              => __( '置顶', 'tanzanite-settings' ),
                        'unstickLabel'            => __( '取消置顶', 'tanzanite-settings' ),
                        'stickyBadge'             => __( '置顶', 'tanzanite-settings' ),
                        'memberPriceLabel'        => __( '会员价', 'tanzanite-settings' ),
                        'stockAlertLabel'         => __( '警戒', 'tanzanite-settings' ),
                        'pointsLimitLabel'        => __( '限制', 'tanzanite-settings' ),
                        'pointsRewardLabel'       => __( '奖励积分', 'tanzanite-settings' ),
                        'priceLabel'              => __( '价格', 'tanzanite-settings' ),
                        'stockLabel'              => __( '库存', 'tanzanite-settings' ),
                        'categoryLabel'           => __( '分类', 'tanzanite-settings' ),
                        'statusLabel'             => __( '状态', 'tanzanite-settings' ),
                        'deleteConfirm'           => __( '确认删除此商品吗？', 'tanzanite-settings' ),
                        'deleteFailed'            => __( '删除失败', 'tanzanite-settings' ),
                        'deleteSuccess'           => __( '删除成功', 'tanzanite-settings' ),
                        'stickyFailed'            => __( '置顶操作失败', 'tanzanite-settings' ),
                        'stickySuccess'           => __( '置顶操作成功', 'tanzanite-settings' ),
                        'copyFailed'              => __( '复制失败', 'tanzanite-settings' ),
                        'copySuccess'             => __( '已复制到剪贴板', 'tanzanite-settings' ),
                        'bulkNoSelection'         => __( '请先选择商品', 'tanzanite-settings' ),
                        'bulkDeleteConfirm'       => __( '确认删除选中的商品吗？', 'tanzanite-settings' ),
                        'bulkDeleteFailed'        => __( '批量删除失败', 'tanzanite-settings' ),
                        'bulkDeleteSuccess'       => __( '批量删除成功', 'tanzanite-settings' ),
                        'bulkPriceNotImplemented' => __( '批量价格调整功能开发中', 'tanzanite-settings' ),
                        'taxonomyLoading'         => __( '加载中...', 'tanzanite-settings' ),
                        'taxonomyHasMore'         => __( '点击加载更多', 'tanzanite-settings' ),
                        'taxonomyNoMore'          => __( '已加载全部', 'tanzanite-settings' ),
                        'taxonomyEmpty'           => __( '没有匹配的结果', 'tanzanite-settings' ),
                        'taxonomyLoadFailed'      => __( '加载失败，请稍后重试', 'tanzanite-settings' ),
                    ),
                )
            );

            echo '<style>'
                . ' .tz-products-filters { display:grid; gap:16px; max-width:1400px; }'
                . ' .tz-fieldset { background:#fff; border:1px solid #dcdcde; border-radius:6px; padding:16px; }'
                . ' .tz-fieldset-title { font-weight:600; font-size:14px; margin-bottom:12px; }'
                . ' .tz-field-grid { display:grid; gap:12px; grid-template-columns:repeat(auto-fit,minmax(220px,1fr)); }'
                . ' .tz-field label { display:flex; flex-direction:column; gap:4px; font-weight:500; }'
                . ' .tz-field label input, .tz-field label select, .tz-field label textarea { font-weight:400; }'
                . ' .tz-field .description { font-size:12px; color:#646970; line-height:1.4; }'
                . ' #tz-products-bulk-panel { grid-template-columns:repeat(auto-fit,minmax(260px,1fr)); margin-bottom:16px; gap:16px; }'
                . ' .tz-bulk-card { background:#fff; border:1px solid #dcdcde; border-radius:6px; padding:16px; display:flex; flex-direction:column; gap:12px; box-shadow:0 1px 2px rgba(0,0,0,0.04); }'
                . ' .tz-bulk-card h3 { margin:0; font-size:15px; }'
                . ' .tz-bulk-card form textarea { font-family:monospace; min-height:72px; }'
                . ' .tz-bulk-grid { display:grid; gap:8px; }'
                . ' .tz-bulk-grid label { display:flex; flex-direction:column; gap:4px; font-weight:500; }'
                . ' .tz-toolbar-note { color:#646970; font-size:12px; margin-left:auto; }'
                . ' .tz-inline-actions { display:flex; flex-wrap:wrap; gap:8px; align-items:center; }'
                . ' .tz-filters-toggle { display:flex; align-items:center; gap:8px; margin-bottom:12px; }'
                . ' .tz-filters-toggle button { display:flex; align-items:center; gap:6px; }'
                . ' .tz-filters-toggle .dashicons { display:inline-block; transition:transform 0.2s ease; }'
                . ' .tz-filters-collapsed { display:none; }'
                . ' .tz-taxonomy-toolbar { display:flex; flex-wrap:wrap; gap:8px; margin-top:8px; align-items:center; }'
                . ' .tz-taxonomy-toolbar .tz-taxonomy-search { width:200px; }'
                . ' .tz-taxonomy-toolbar .button-link { padding:0; }'
                . ' .tz-taxonomy-toolbar .tz-taxonomy-status { font-size:12px; color:#646970; }'
                . ' @media (max-width: 782px) { .tz-toolbar-note { display:none; } }'
                . '</style>';

            echo '<div class="tz-settings-wrapper tz-products-list">';

            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'All Products', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '查看、筛选并管理商品，可切换表格 / 卡片视图，并快速执行常用操作。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-products-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-products-summary" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:16px;margin-bottom:24px;">';
            foreach ( [
                [ 'key' => 'total_products', 'label' => __( '商品总数', 'tanzanite-settings' ) ],
                [ 'key' => 'publish', 'label' => __( '已上架', 'tanzanite-settings' ) ],
                [ 'key' => 'draft', 'label' => __( '草稿', 'tanzanite-settings' ) ],
                [ 'key' => 'pending', 'label' => __( '待审核', 'tanzanite-settings' ) ],
                [ 'key' => 'low_stock', 'label' => __( '低库存', 'tanzanite-settings' ) ],
                [ 'key' => 'pending_reviews', 'label' => __( '待审核评价', 'tanzanite-settings' ) ],
            ] as $card ) {
                echo '<div class="tz-dashboard-card" data-metric="' . esc_attr( $card['key'] ) . '">';
                echo '  <div class="tz-card-value">-</div>';
                echo '  <div class="tz-card-label">' . esc_html( $card['label'] ) . '</div>';
                echo '</div>';
            }
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
            echo '      <div class="tz-filters-toggle">';
            echo '          <button type="button" class="button" id="tz-products-filters-toggle" aria-expanded="true">';
            echo '              <span class="dashicons dashicons-arrow-down"></span><span class="tz-toggle-label">' . esc_html__( '收起筛选项', 'tanzanite-settings' ) . '</span>';
            echo '          </button>';
            echo '          <span class="tz-toolbar-note">' . esc_html__( '可折叠筛选区域，节省页面空间。', 'tanzanite-settings' ) . '</span>';
            echo '      </div>';
            echo '      <form id="tz-products-filters" class="tz-products-filters" aria-hidden="false">';

            echo '          <div class="tz-fieldset">';
            echo '              <div class="tz-fieldset-title">' . esc_html__( '基础筛选', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-field-grid">';
            echo '                  <div class="tz-field"><label>' . esc_html__( '关键词（标题）', 'tanzanite-settings' ) . '<input type="text" name="keyword" class="widefat" placeholder="' . esc_attr__( '商品名称关键字', 'tanzanite-settings' ) . '" /></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( 'SKU 编码', 'tanzanite-settings' ) . '<input type="text" name="sku" class="widefat" placeholder="SKU-001" /><span class="description">' . esc_html__( '支持模糊匹配，自动遍历 SKU 表。', 'tanzanite-settings' ) . '</span></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select name="status" class="widefat"><option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
            foreach ( self::ALLOWED_PRODUCT_STATUSES as $status ) {
                echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
            }
            echo '                  </select></label></div>';
            echo '                  <div class="tz-field tz-taxonomy" data-taxonomy="category">';
            echo '                      <label>' . esc_html__( '分类', 'tanzanite-settings' ) . '<select name="category" class="widefat" id="tz-filter-category"><option value="">' . esc_html__( '全部分类', 'tanzanite-settings' ) . '</option></select><span class="description">' . esc_html__( '可搜索分类名称并分页加载。', 'tanzanite-settings' ) . '</span></label>';
            echo '                      <div class="tz-taxonomy-toolbar">';
            echo '                          <input type="search" class="tz-taxonomy-search regular-text" data-taxonomy="category" placeholder="' . esc_attr__( '搜索分类…', 'tanzanite-settings' ) . '" />';
            echo '                          <button type="button" class="button tz-taxonomy-search-btn" data-taxonomy="category">' . esc_html__( '搜索', 'tanzanite-settings' ) . '</button>';
            echo '                          <button type="button" class="button tz-taxonomy-load-more" data-taxonomy="category">' . esc_html__( '加载更多', 'tanzanite-settings' ) . '</button>';
            echo '                          <span class="tz-taxonomy-status" data-taxonomy="category"></span>';
            echo '                      </div>';
            echo '                  </div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '作者 ID', 'tanzanite-settings' ) . '<input type="number" name="author" class="widefat" min="0" placeholder="0" /><span class="description">' . esc_html__( '可用于筛选创建人，0 表示忽略。', 'tanzanite-settings' ) . '</span></label></div>';
            echo '              </div>';
            echo '          </div>';

            echo '          <div class="tz-fieldset">';
            echo '              <div class="tz-fieldset-title">' . esc_html__( '库存与积分', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-field-grid">';
            echo '                  <div class="tz-field"><label>' . esc_html__( '库存下限', 'tanzanite-settings' ) . '<input type="number" name="inventory_min" class="widefat" placeholder="0" /></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '库存上限', 'tanzanite-settings' ) . '<input type="number" name="inventory_max" class="widefat" placeholder="9999" /></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '积分下限', 'tanzanite-settings' ) . '<input type="number" name="points_min" class="widefat" placeholder="0" /></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '积分上限', 'tanzanite-settings' ) . '<input type="number" name="points_max" class="widefat" placeholder="9999" /></label></div>';
            echo '              </div>';
            echo '          </div>';

            echo '          <div class="tz-fieldset">';
            echo '              <div class="tz-fieldset-title">' . esc_html__( '高级筛选', 'tanzanite-settings' ) . '</div>';
            echo '              <div class="tz-field-grid">';
            echo '                  <div class="tz-field tz-taxonomy" data-taxonomy="tags">';
            echo '                      <label>' . esc_html__( '标签', 'tanzanite-settings' ) . '<select multiple name="tags[]" class="widefat" id="tz-filter-tags" data-placeholder="tag-a,tag-b"></select><span class="description">' . esc_html__( '支持多选，使用 Ctrl/Command 选择多个标签，可搜索加载更多。', 'tanzanite-settings' ) . '</span></label>';
            echo '                      <div class="tz-taxonomy-toolbar">';
            echo '                          <input type="search" class="tz-taxonomy-search regular-text" data-taxonomy="tags" placeholder="' . esc_attr__( '搜索标签…', 'tanzanite-settings' ) . '" />';
            echo '                          <button type="button" class="button tz-taxonomy-search-btn" data-taxonomy="tags">' . esc_html__( '搜索', 'tanzanite-settings' ) . '</button>';
            echo '                          <button type="button" class="button tz-taxonomy-load-more" data-taxonomy="tags">' . esc_html__( '加载更多', 'tanzanite-settings' ) . '</button>';
            echo '                          <span class="tz-taxonomy-status" data-taxonomy="tags"></span>';
            echo '                      </div>';
            echo '                  </div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '属性筛选', 'tanzanite-settings' ) . '<input type="text" name="attributes" class="widefat" placeholder="pa_color:red,pa_size:xl" /><span class="description">' . esc_html__( '格式为 taxonomy:term_slug，可填写多个，用逗号分隔。', 'tanzanite-settings' ) . '</span></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '排序字段', 'tanzanite-settings' ) . '<select name="sort" class="widefat">';
            foreach ( [ 'updated_at' => __( '更新时间', 'tanzanite-settings' ), 'price_regular' => __( '原价', 'tanzanite-settings' ), 'stock_qty' => __( '库存', 'tanzanite-settings' ), 'points_reward' => __( '积分奖励', 'tanzanite-settings' ) ] as $key => $label ) {
                echo '<option value="' . esc_attr( $key ) . '">' . esc_html( $label ) . '</option>';
            }
            echo '                  </select></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '排序方式', 'tanzanite-settings' ) . '<select name="order" class="widefat"><option value="DESC">DESC</option><option value="ASC">ASC</option></select></label></div>';
            echo '                  <div class="tz-field"><label>' . esc_html__( '每页数量', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select><span class="description">' . esc_html__( '默认 20 条，最大支持 200 条。', 'tanzanite-settings' ) . '</span></label></div>';
            echo '              </div>';
            echo '          </div>';

            echo '      </form>';
            echo '      <div class="tz-inline-actions" style="margin-top:4px;">';
            echo '          <button class="button button-primary" id="tz-products-filter-submit">' . esc_html__( '应用筛选', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button" id="tz-products-filter-reset">' . esc_html__( '重置条件', 'tanzanite-settings' ) . '</button>';
            if ( $can_bulk ) {
                echo '          <a class="button" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-sku-importer' ) ) . '">' . esc_html__( '前往批量工具', 'tanzanite-settings' ) . '</a>';
            }
            if ( $can_manage ) {
                echo '          <a class="button button-secondary" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-add-product' ) ) . '">' . esc_html__( '新建商品', 'tanzanite-settings' ) . '</a>';
            }
            echo '          <span class="tz-toolbar-note">' . esc_html__( '提示：更多筛选项将与 Nuxt 前端保持一致，可随时扩展。', 'tanzanite-settings' ) . '</span>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section" id="tz-products-view">';
            echo '      <div class="tz-products-toolbar" style="display:flex;flex-wrap:wrap;gap:12px;align-items:center;margin-bottom:12px;">';
            echo '          <div class="button-group" role="group">';
            echo '              <button type="button" class="button button-secondary tz-view-toggle is-active" data-view="table">' . esc_html__( '表格视图', 'tanzanite-settings' ) . '</button>';
            echo '              <button type="button" class="button button-secondary tz-view-toggle" data-view="cards">' . esc_html__( '卡片视图', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '          <button type="button" class="button" id="tz-products-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
            if ( $can_bulk ) {
                echo '          <button type="button" class="button" id="tz-products-bulk-toggle" aria-expanded="false">' . esc_html__( '批量操作', 'tanzanite-settings' ) . '</button>';
            }
            echo '      </div>';

            if ( $can_bulk ) {
                echo '      <div id="tz-products-bulk-panel" style="display:none;gap:16px;margin-bottom:16px;">';
                echo '          <div class="tz-bulk-card">';
                echo '              <h3>' . esc_html__( '批量上下架', 'tanzanite-settings' ) . '</h3>';
                echo '              <form class="tz-bulk-form" data-action="set_status">';
                echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
                echo '                  <select name="status" class="widefat">';
                foreach ( self::ALLOWED_PRODUCT_STATUSES as $status ) {
                    echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
                }
                echo '                  </select>';
                echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量更新状态', 'tanzanite-settings' ) . '</button>';
                echo '              </form>';
                echo '          </div>';
                echo '          <div class="tz-bulk-card">';
                echo '              <h3>' . esc_html__( '批量调整库存', 'tanzanite-settings' ) . '</h3>';
                echo '              <form class="tz-bulk-form" data-action="adjust_stock">';
                echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
                echo '                  <input type="number" class="widefat" name="delta" placeholder="±10" />';
                echo '                  <span class="description">' . esc_html__( '正数为增加库存，负数为扣减库存。', 'tanzanite-settings' ) . '</span>';
                echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量调整库存', 'tanzanite-settings' ) . '</button>';
                echo '              </form>';
                echo '          </div>';
                echo '          <div class="tz-bulk-card">';
                echo '              <h3>' . esc_html__( '批量调价', 'tanzanite-settings' ) . '</h3>';
                echo '              <form class="tz-bulk-form" data-action="adjust_price">';
                echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
                echo '                  <div class="tz-bulk-grid">';
                echo '                      <label>' . esc_html__( '调价模式', 'tanzanite-settings' ) . '<select name="mode" class="widefat"><option value="absolute">' . esc_html__( '固定值（元）', 'tanzanite-settings' ) . '</option><option value="percent">' . esc_html__( '百分比（%）', 'tanzanite-settings' ) . '</option></select></label>';
                echo '                      <label>' . esc_html__( '调价幅度', 'tanzanite-settings' ) . '<input type="number" step="0.01" name="value" class="widefat" placeholder="10 或 5" /></label>';
                echo '                      <label>' . esc_html__( '小数位数', 'tanzanite-settings' ) . '<select name="round" class="widefat"><option value="2">2</option><option value="1">1</option><option value="0">0</option><option value="3">3</option><option value="4">4</option></select><span class="description">' . esc_html__( '调价后的数值会根据此设置四舍五入。', 'tanzanite-settings' ) . '</span></label>';
                echo '                  </div>';
                echo '                  <fieldset style="border:1px solid #dcdcde;border-radius:4px;padding:12px;">';
                echo '                      <legend>' . esc_html__( '选择需要调价的字段', 'tanzanite-settings' ) . '</legend>';
                foreach ( [
                    'price_regular' => __( '原价', 'tanzanite-settings' ),
                    'price_sale'    => __( '现价', 'tanzanite-settings' ),
                    'price_member'  => __( '会员价', 'tanzanite-settings' ),
                ] as $field => $label ) {
                    echo '                      <label style="display:flex;align-items:center;gap:6px;margin-bottom:6px;"><input type="checkbox" name="fields[]" value="' . esc_attr( $field ) . '"> ' . esc_html( $label ) . '</label>';
                }
                echo '                      <span class="description">' . esc_html__( '百分比模式示例：输入 5 即在原有价格基础上 +5%。可填写负值表示降价。', 'tanzanite-settings' ) . '</span>';
                echo '                  </fieldset>';
                echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量调价', 'tanzanite-settings' ) . '</button>';
                echo '              </form>';
                echo '          </div>';
                echo '          <div class="tz-bulk-card">';
                echo '              <h3>' . esc_html__( '批量设定字段', 'tanzanite-settings' ) . '</h3>';
                echo '              <form class="tz-bulk-form" data-action="set_meta">';
                echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
                echo '                  <div class="tz-bulk-grid">';
                foreach ( [
                    'price_regular' => __( '原价', 'tanzanite-settings' ),
                    'price_sale'    => __( '现价', 'tanzanite-settings' ),
                    'price_member'  => __( '会员价', 'tanzanite-settings' ),
                    'stock_qty'     => __( '库存', 'tanzanite-settings' ),
                    'points_reward' => __( '积分奖励', 'tanzanite-settings' ),
                    'points_limit'  => __( '积分上限', 'tanzanite-settings' ),
                ] as $field => $label ) {
                    echo '<label>' . esc_html( $label ) . '<input type="number" step="0.01" name="' . esc_attr( $field ) . '" class="widefat" /></label>';
                }
                echo '                  </div>';
                echo '                  <span class="description">' . esc_html__( '直接覆盖所选字段的值，可用于一次性设定库存或积分。', 'tanzanite-settings' ) . '</span>';
                echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量更新字段', 'tanzanite-settings' ) . '</button>';
                echo '              </form>';
                echo '          </div>';
                echo '          <div class="tz-bulk-card">';
                echo '              <h3>' . esc_html__( '批量设置推荐位', 'tanzanite-settings' ) . '</h3>';
                echo '              <form class="tz-bulk-form" data-action="set_featured">';
                echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
                echo '                  <select name="enabled" class="widefat">';
                echo '                      <option value="1">' . esc_html__( '设为推荐', 'tanzanite-settings' ) . '</option>';
                echo '                      <option value="0">' . esc_html__( '取消推荐', 'tanzanite-settings' ) . '</option>';
                echo '                  </select>';
                echo '                  <input type="text" class="widefat" name="slot" placeholder="Homepage-Top" />';
                echo '                  <span class="description">' . esc_html__( '可选的推荐位标识，取消推荐时可留空。', 'tanzanite-settings' ) . '</span>';
                echo '                  <button type="submit" class="button button-primary">' . esc_html__( '批量设置推荐位', 'tanzanite-settings' ) . '</button>';
                echo '              </form>';
                echo '          </div>';
                echo '          <div class="tz-bulk-card">';
                echo '              <h3>' . esc_html__( '批量删除', 'tanzanite-settings' ) . '</h3>';
                echo '              <form class="tz-bulk-form" data-action="delete">';
                echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
                echo '                  <select name="mode" class="widefat">';
                echo '                      <option value="trash">' . esc_html__( '移动到回收站（可恢复）', 'tanzanite-settings' ) . '</option>';
                echo '                      <option value="force">' . esc_html__( '永久删除（不可恢复）', 'tanzanite-settings' ) . '</option>';
                echo '                  </select>';
                echo '                  <span class="description">' . esc_html__( '建议先移至回收站以便回滚；永久删除将同步移除 SKU 数据。', 'tanzanite-settings' ) . '</span>';
                echo '                  <button type="submit" class="button button-secondary">' . esc_html__( '批量删除', 'tanzanite-settings' ) . '</button>';
                echo '              </form>';
                echo '          </div>';
                echo '          <div class="tz-bulk-card">';
                echo '              <h3>' . esc_html__( '批量导出', 'tanzanite-settings' ) . '</h3>';
                echo '              <form class="tz-bulk-form" data-action="export">';
                echo '                  <textarea class="widefat" name="ids" rows="3" placeholder="1,2,3"></textarea>';
                echo '                  <button type="submit" class="button">' . esc_html__( '导出选中商品', 'tanzanite-settings' ) . '</button>';
                echo '              </form>';
                echo '          </div>';
                echo '      </div>';
            }

            echo '      <div class="tz-products-table-wrapper" style="overflow:auto;">';
            echo '          <table class="widefat fixed striped" id="tz-products-table" style="min-width:1200px;">';
            echo '              <thead><tr>';
            foreach ( [ __( '商品信息', 'tanzanite-settings' ), __( '价格', 'tanzanite-settings' ), __( '库存', 'tanzanite-settings' ), __( '积分', 'tanzanite-settings' ), __( '分类', 'tanzanite-settings' ), __( '状态', 'tanzanite-settings' ), __( '更新时间', 'tanzanite-settings' ), __( '操作', 'tanzanite-settings' ) ] as $column ) {
                echo '<th>' . esc_html( $column ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';

            echo '      <div class="tz-products-cards" id="tz-products-cards" style="display:none;gap:16px;flex-wrap:wrap;"></div>';

            echo '      <div class="tz-products-pagination" style="display:flex;align-items:center;gap:12px;margin-top:16px;">';
            echo '          <button class="button" id="tz-products-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
            echo '          <span id="tz-products-page-info">1/1</span>';
            echo '          <button class="button" id="tz-products-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '</div>';

            $config_js = wp_json_encode(
                [
                    'listUrl'    => $list_endpoint,
                    'nonce'      => $nonce,
                    'canManage'  => $can_manage,
                    'canBulk'    => $can_bulk,
                    'editUrl'    => $single_link,
                    'seoUrl'     => $seo_link,
                    'bulkUrl'    => $bulk_link,
                    'categoriesEndpoint' => esc_url_raw( rest_url( 'tanzanite/v1/taxonomies/category' ) ),
                    'tagsEndpoint'        => esc_url_raw( rest_url( 'tanzanite/v1/taxonomies/post_tag' ) ),
                ],
                JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES
            );

            $strings_js = wp_json_encode(
                [
                    'loading'              => __( '加载中…', 'tanzanite-settings' ),
                    'noData'               => __( '暂无商品记录。', 'tanzanite-settings' ),
                    'refreshSucceeded'     => __( '刷新成功。', 'tanzanite-settings' ),
                    'refreshFailed'        => __( '刷新失败，请稍后重试。', 'tanzanite-settings' ),
                    'deleteConfirm'        => __( '确定要删除该商品吗？此操作不可撤销。', 'tanzanite-settings' ),
                    'deleteDone'           => __( '商品已删除。', 'tanzanite-settings' ),
                    'deleteFailed'         => __( '删除失败。', 'tanzanite-settings' ),
                    'copySuccess'          => __( '已复制到剪贴板。', 'tanzanite-settings' ),
                    'copyFailed'           => __( '复制失败，请手动复制。', 'tanzanite-settings' ),
                    'viewTable'            => __( '表格视图', 'tanzanite-settings' ),
                    'viewCards'            => __( '卡片视图', 'tanzanite-settings' ),
                    'editLabel'            => __( '编辑', 'tanzanite-settings' ),
                    'seoLabel'             => 'SEO',
                    'copyPayloadLabel'     => __( '复制 Payload', 'tanzanite-settings' ),
                    'deleteLabel'          => __( '删除', 'tanzanite-settings' ),
                    'memberPriceLabel'     => __( '会员价', 'tanzanite-settings' ),
                    'priceLabel'           => __( '售价', 'tanzanite-settings' ),
                    'stockLabel'           => __( '库存', 'tanzanite-settings' ),
                    'stockAlertLabel'      => __( '预警', 'tanzanite-settings' ),
                    'pointsRewardLabel'    => __( '积分', 'tanzanite-settings' ),
                    'pointsLimitLabel'     => __( '上限', 'tanzanite-settings' ),
                    'categoryLabel'        => __( '分类', 'tanzanite-settings' ),
                    'statusLabel'          => __( '状态', 'tanzanite-settings' ),
                    'stickyBadge'          => __( '置顶', 'tanzanite-settings' ),
                    'refreshing'           => __( '刷新中…', 'tanzanite-settings' ),
                    'previewLabel'         => __( '前台预览', 'tanzanite-settings' ),
                    'stickLabel'           => __( '设为置顶', 'tanzanite-settings' ),
                    'unstickLabel'         => __( '取消置顶', 'tanzanite-settings' ),
                    'toggleStickyDone'     => __( '置顶状态已更新。', 'tanzanite-settings' ),
                    'toggleStickyFailed'   => __( '更新置顶状态失败。', 'tanzanite-settings' ),
                    'quickEditLabel'       => __( '快速修改', 'tanzanite-settings' ),
                    'quickEditSuccess'     => __( '商品信息已更新。', 'tanzanite-settings' ),
                    'quickEditFailed'      => __( '快速修改失败。', 'tanzanite-settings' ),
                    'promptPrice'          => __( '请输入新的现价（留空表示不修改）：', 'tanzanite-settings' ),
                    'promptStock'          => __( '请输入新的库存数量（留空不改）：', 'tanzanite-settings' ),
                    'promptPoints'         => __( '请输入新的积分奖励（留空不改）：', 'tanzanite-settings' ),
                    'bulkNeedIds'          => __( '请先填写有效的商品 ID。', 'tanzanite-settings' ),
                    'bulkProcessing'       => __( '正在处理批量请求…', 'tanzanite-settings' ),
                    'bulkDone'             => __( '批量操作完成。', 'tanzanite-settings' ),
                    'bulkFailed'           => __( '批量操作失败。', 'tanzanite-settings' ),
                    'bulkNoDelta'          => __( '库存增量不能为 0。', 'tanzanite-settings' ),
                    'bulkNeedsField'       => __( '请至少填写一个需要更新的字段。', 'tanzanite-settings' ),
                    'taxonomyLoading'      => __( '正在加载…', 'tanzanite-settings' ),
                    'taxonomyHasMore'      => __( '可继续加载更多。', 'tanzanite-settings' ),
                    'taxonomyNoMore'       => __( '已加载全部结果。', 'tanzanite-settings' ),
                    'taxonomyEmpty'        => __( '没有匹配的结果。', 'tanzanite-settings' ),
                    'taxonomyLoadFailed'   => __( '加载失败，请稍后重试。', 'tanzanite-settings' ),
                    'priceRegularLabel'    => __( '原价', 'tanzanite-settings' ),
                    'priceSaleLabel'       => __( '现价', 'tanzanite-settings' ),
                    'priceMemberLabel'     => __( '会员价', 'tanzanite-settings' ),
                    'bulkPriceModeAbsolute'=> __( '固定值', 'tanzanite-settings' ),
                    'bulkPriceModePercent' => __( '百分比', 'tanzanite-settings' ),
                    'bulkPricePreviewIntro'=> __( '即将对 %COUNT% 个商品执行批量调价（模式：%MODE%，幅度：%VALUE%，保留小数：%ROUND% 位）。', 'tanzanite-settings' ),
                    'bulkPricePreviewConfirm' => __( '以下为最多 5 条示例，确认继续执行调价？', 'tanzanite-settings' ),
                    'bulkPricePreviewNoData'  => __( '未能获取调价示例，将直接提交调价请求。确认继续吗？', 'tanzanite-settings' ),
                    'bulkPriceRollbackHint'   => __( '提示：如需回滚，请记录本次参数并以相反数或百分比再次批量调价。', 'tanzanite-settings' ),
                    'bulkDeleteConfirm'       => __( '确认要删除选中的商品吗？可选择移至回收站或永久删除，操作可能无法撤销。', 'tanzanite-settings' ),
                    'bulkDeleteModeTrash'     => __( '移动到回收站（可恢复）', 'tanzanite-settings' ),
                    'bulkDeleteModeForce'     => __( '永久删除（不可恢复）', 'tanzanite-settings' ),
                ],
                JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES
            );


        }

        public function render_orders_list(): void {
            $nonce      = wp_create_nonce( 'wp_rest' );
            $list_url   = esc_url_raw( rest_url( 'tanzanite/v1/orders' ) );
            $sync_url   = esc_url_raw( rest_url( 'tanzanite/v1/orders/' ) );
            $statuses   = array_values( self::ALLOWED_ORDER_STATUSES );
            $can_manage = current_user_can( 'tanz_manage_orders' );

            // 加载订单列表 JS
            wp_enqueue_script(
                'tz-orders-list',
                plugins_url( 'assets/js/orders-list.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-admin-common' ),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-orders-list',
                'TzOrdersListConfig',
                array(
                    'listUrl'      => $list_url,
                    'syncBase'     => $sync_url,
                    'nonce'        => $nonce,
                    'canManage'    => $can_manage,
                    'detailBase'   => esc_url_raw( admin_url( 'admin.php?page=tanzanite-settings-order-detail&order_id=' ) ),
                    'statusLabels' => array(
                        'pending'    => __( '待支付', 'tanzanite-settings' ),
                        'paid'       => __( '已支付', 'tanzanite-settings' ),
                        'processing' => __( '处理中', 'tanzanite-settings' ),
                        'shipped'    => __( '已发货', 'tanzanite-settings' ),
                        'completed'  => __( '已完成', 'tanzanite-settings' ),
                        'cancelled'  => __( '已取消', 'tanzanite-settings' ),
                    ),
                )
            );

            echo '<div class="tz-settings-wrapper tz-orders-list">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'All Orders', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '按条件筛选订单，支持刷新物流状态与跳转批量工具。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-orders-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-orders-filters" class="tz-orders-filters" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px;max-width:1400px;">';
            echo '          <label>' . esc_html__( '订单状态', 'tanzanite-settings' ) . '<select name="status" class="widefat"><option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
            foreach ( $statuses as $status ) {
                echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
            }
            echo '          </select></label>';
            echo '          <label>' . esc_html__( '渠道来源', 'tanzanite-settings' ) . '<input type="text" name="channel" class="widefat" placeholder="web/app/h5" /></label>';
            echo '          <label>' . esc_html__( '支付方式', 'tanzanite-settings' ) . '<input type="text" name="payment_method" class="widefat" placeholder="wechat_pay" /></label>';
            echo '          <label>' . esc_html__( '物流服务商', 'tanzanite-settings' ) . '<input type="text" name="tracking_provider" class="widefat" placeholder="17track" /></label>';
            echo '          <label>' . esc_html__( '客户关键词', 'tanzanite-settings' ) . '<input type="text" name="customer_keyword" class="widefat" placeholder="姓名/邮箱/账号" /></label>';
            echo '          <label>' . esc_html__( '起始时间', 'tanzanite-settings' ) . '<input type="date" name="date_start" class="widefat" /></label>';
            echo '          <label>' . esc_html__( '结束时间', 'tanzanite-settings' ) . '<input type="date" name="date_end" class="widefat" /></label>';
            echo '          <label>' . esc_html__( '每页条数', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select></label>';
            echo '      </form>';
            echo '      <div style="display:flex;gap:12px;margin-top:12px;">';
            echo '          <button class="button button-primary" id="tz-orders-filter-submit">' . esc_html__( '应用筛选', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button" id="tz-orders-filter-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
            echo '          <a class="button" href="' . esc_url( admin_url( 'admin.php?page=tanzanite-settings-orders-bulk' ) ) . '">' . esc_html__( '前往批量工具', 'tanzanite-settings' ) . '</a>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '订单列表', 'tanzanite-settings' ) . '</div>';
            echo '      <div style="overflow:auto;margin-top:12px;">';
            echo '          <table class="widefat fixed striped" id="tz-orders-table" style="min-width:1200px;">';
            echo '              <thead><tr>';
            foreach ( [ __( '订单信息', 'tanzanite-settings' ), __( '客户', 'tanzanite-settings' ), __( '金额', 'tanzanite-settings' ), __( '状态', 'tanzanite-settings' ), __( '渠道 / 支付', 'tanzanite-settings' ), __( '创建时间', 'tanzanite-settings' ), __( '物流', 'tanzanite-settings' ), __( '操作', 'tanzanite-settings' ) ] as $column ) {
                echo '<th>' . esc_html( $column ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '      <div class="tz-orders-pagination" style="display:flex;align-items:center;gap:12px;margin-top:12px;">';
            echo '          <button class="button" id="tz-orders-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
            echo '          <span id="tz-orders-page-info"></span>';
            echo '          <button class="button" id="tz-orders-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '</div>';
        }

        public function render_reviews(): void {
            $nonce        = wp_create_nonce( 'wp_rest' );
            $rest_list    = esc_url_raw( rest_url( 'tanzanite/v1/reviews' ) );
            $rest_single  = esc_url_raw( rest_url( 'tanzanite/v1/reviews/' ) );
            $statuses     = wp_json_encode( self::ALLOWED_REVIEW_STATUSES );
            $can_manage   = current_user_can( 'tanz_manage_reviews' );

            echo '<div class="tz-settings-wrapper tz-reviews-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Product Reviews', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '集中处理来自各渠道的商品评价，支持审核、回复与标记精华。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-review-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-review-filters" class="tz-review-filters" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px;max-width:1200px;">';
            echo '          <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select name="status" class="widefat">';
            echo '              <option value="">' . esc_html__( '全部', 'tanzanite-settings' ) . '</option>';
            foreach ( self::ALLOWED_REVIEW_STATUSES as $status ) {
                echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
            }
            echo '          </select></label>';
            echo '          <label>' . esc_html__( '商品 ID', 'tanzanite-settings' ) . '<input type="number" name="product_id" class="widefat" /></label>';
            echo '          <label>' . esc_html__( '关键词（作者/内容）', 'tanzanite-settings' ) . '<input type="text" name="search" class="widefat" /></label>';
            echo '          <label>' . esc_html__( '每页条数', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select></label>';
            echo '      </form>';
            echo '      <div style="display:flex;gap:12px;margin-top:12px;">';
            echo '          <button class="button button-primary" id="tz-review-refresh">' . esc_html__( '刷新列表', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button" id="tz-review-reset">' . esc_html__( '重置筛选', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '评价列表', 'tanzanite-settings' ) . '</div>';
            echo '      <div style="overflow:auto;margin-top:12px;">';
            echo '          <table class="widefat fixed striped" id="tz-review-table" style="min-width:1100px;">';
            echo '              <thead><tr>';
            foreach ( [ 'ID', 'Product', 'User', 'Rating', 'Content', 'Status', 'Featured', 'Created', 'Actions' ] as $column ) {
                echo '<th>' . esc_html( $column ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '      <div class="tz-review-pagination" style="display:flex;align-items:center;gap:12px;margin-top:12px;">';
            echo '          <button class="button" id="tz-review-prev">' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
            echo '          <span id="tz-review-page-info"></span>';
            echo '          <button class="button" id="tz-review-next">' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '评价详情与审核', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-review-detail" style="display:grid;gap:12px;max-width:1100px;">';
            echo '          <input type="hidden" id="tz-review-id" />';
            echo '          <div class="tz-review-meta" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:12px;">';
            echo '              <label>' . esc_html__( '评价状态', 'tanzanite-settings' ) . '<select id="tz-review-status" class="widefat">';
            echo '                  <option value="">' . esc_html__( '保持不变', 'tanzanite-settings' ) . '</option>';
            foreach ( self::ALLOWED_REVIEW_STATUSES as $status ) {
                echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
            }
            echo '              </select></label>';
            echo '              <label style="display:flex;align-items:center;gap:8px;margin-top:24px;"><input type="checkbox" id="tz-review-featured" /> ' . esc_html__( '标记为精华', 'tanzanite-settings' ) . '</label>';
            echo '              <div>';
            echo '                  <div><strong>' . esc_html__( '评分', 'tanzanite-settings' ) . ':</strong> <span id="tz-review-rating">-</span></div>';
            echo '                  <div><strong>' . esc_html__( '作者', 'tanzanite-settings' ) . ':</strong> <span id="tz-review-author">-</span></div>';
            echo '                  <div><strong>' . esc_html__( '创建时间', 'tanzanite-settings' ) . ':</strong> <span id="tz-review-created">-</span></div>';
            echo '              </div>';
            echo '          </div>';
            echo '          <label>' . esc_html__( '评价内容', 'tanzanite-settings' ) . '<textarea id="tz-review-content" rows="6" class="widefat" readonly style="background:#f6f7f7;"></textarea></label>';
            echo '          <label>' . esc_html__( '附件', 'tanzanite-settings' ) . '<div id="tz-review-images" style="display:flex;gap:12px;flex-wrap:wrap;"></div></label>';
            echo '          <label>' . esc_html__( '后台回复', 'tanzanite-settings' ) . '<textarea id="tz-review-reply" rows="4" class="widefat" placeholder="' . esc_attr__( '输入回复内容，留空将清除回复。', 'tanzanite-settings' ) . '"></textarea></label>';
            echo '          <div style="display:flex;gap:12px;">';
            echo '              <button class="button button-primary" id="tz-review-save" type="submit">' . esc_html__( '保存更新', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button" id="tz-review-cancel" type="button">' . esc_html__( '取消选择', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button button-secondary" id="tz-review-delete" type="button">' . esc_html__( '删除评价', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '      </form>';
            echo '  </div>';

            echo '</div>';

            // 加载评价管理 JS
            wp_enqueue_script(
                'tz-reviews',
                plugins_url( 'assets/js/reviews.js', TANZANITE_LEGACY_MAIN_FILE ),
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

        public function render_payment_method(): void {
            wp_enqueue_media();
            $nonce = wp_create_nonce( 'wp_rest' );

            // 加载支付方式管理 JS
            wp_enqueue_script(
                'tz-payment-methods',
                plugins_url( 'assets/js/payment-methods.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'jquery', 'wp-media' ),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-payment-methods',
                'TzPaymentMethodsConfig',
                array(
                    'listUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/payment-methods' ) ),
                    'singleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/payment-methods/' ) ),
                    'nonce'     => $nonce,
                    'gatewayFields' => array(
                        'paypal' => array( 'client_id', 'client_secret', 'mode', 'webhook_id' ),
                        'stripe' => array( 'publishable_key', 'secret_key', 'webhook_secret', 'mode' ),
                        'worldfirst' => array( 'merchant_id', 'api_key', 'api_secret', 'mode' ),
                        'payoneer' => array( 'program_id', 'api_username', 'api_password', 'mode' ),
                    ),
                )
            );

            echo '<div class="tz-settings-wrapper tz-payments-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Payment Methods', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '配置前端可用的支付方式，包括手续费、终端可见性与会员等级限制。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-payment-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '支付方式列表', 'tanzanite-settings' ) . '</div>';
            echo '      <button class="button button-primary" id="tz-payment-create">' . esc_html__( '新增支付方式', 'tanzanite-settings' ) . '</button>';
            echo '      <div style="overflow:auto;margin-top:16px;">';
            echo '          <table class="widefat fixed striped" id="tz-payment-table" style="min-width:960px;">';
            echo '              <thead><tr>';
            foreach ( [ 'Code', 'Name', 'Fee', 'Terminals', 'Membership Levels', 'Enabled', 'Actions' ] as $column ) {
                echo '<th>' . esc_html( $column ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增支付方式', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-payment-form" class="tz-payment-form">';
            echo '          <input type="hidden" id="tz-payment-id" />';
            echo '          <div class="tz-form-grid">';
            echo '              <label>' . esc_html__( '编码 (code)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-code" required /></label>';
            echo '              <label>' . esc_html__( '名称', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-name" required /></label>';
            echo '              <label>' . esc_html__( '手续费类型', 'tanzanite-settings' ) . '<select id="tz-payment-fee-type"><option value="fixed">' . esc_html__( '固定金额', 'tanzanite-settings' ) . '</option><option value="percentage">' . esc_html__( '百分比', 'tanzanite-settings' ) . '</option></select></label>';
            echo '              <label>' . esc_html__( '手续费数值', 'tanzanite-settings' ) . '<input type="number" step="0.0001" id="tz-payment-fee-value" min="0" value="0" /></label>';
            echo '              <label>' . esc_html__( '排序 (0 最靠前)', 'tanzanite-settings' ) . '<input type="number" id="tz-payment-sort" value="0" /></label>';
            echo '              <label>' . esc_html__( '启用', 'tanzanite-settings' ) . '<select id="tz-payment-enabled"><option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option><option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option></select></label>';
            echo '          </div>';
            echo '          <div style="margin-top:12px;">';
            echo '              <label>' . esc_html__( '图标 URL', 'tanzanite-settings' ) . '</label>';
            echo '              <div style="display:flex;gap:8px;align-items:center;">';
            echo '                  <input type="text" id="tz-payment-icon-url" class="regular-text" placeholder="https://example.com/icon.png" />';
            echo '                  <button type="button" class="button" id="tz-payment-icon-upload">' . esc_html__( '选择图片', 'tanzanite-settings' ) . '</button>';
            echo '              </div>';
            echo '              <div id="tz-payment-icon-preview" style="margin-top:8px;display:none;">';
            echo '                  <img src="" alt="" style="max-width:120px;max-height:60px;border:1px solid #ddd;padding:4px;background:#fff;" />';
            echo '              </div>';
            echo '          </div>';

            echo '          <label>' . esc_html__( '适用终端', 'tanzanite-settings' ) . '<div class="tz-checkbox-list" id="tz-payment-terminals"></div></label>';
            echo '          <label>' . esc_html__( '可见会员等级 (逗号分隔或逐个添加)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-levels" placeholder="gold, platinum" /></label>';
            echo '          <label>' . esc_html__( '支持的货币 (逗号分隔，如 CNY,USD,EUR)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-currencies" placeholder="CNY, USD, EUR" /></label>';
            echo '          <label>' . esc_html__( '默认货币 (必须在支持列表中)', 'tanzanite-settings' ) . '<input type="text" id="tz-payment-default-currency" placeholder="CNY" maxlength="10" /></label>';
            echo '          <label>' . esc_html__( '描述', 'tanzanite-settings' ) . '<textarea id="tz-payment-description" rows="3"></textarea></label>';
            
            // 第三方支付平台对接配置
            echo '          <div style="margin-top:24px;padding-top:24px;border-top:1px solid #e5e7eb;">';
            echo '              <h3 style="margin:0 0 16px 0;font-size:14px;font-weight:600;">' . esc_html__( '第三方支付平台对接', 'tanzanite-settings' ) . '</h3>';
            echo '              <div style="margin-bottom:16px;">';
            echo '                  <label>' . esc_html__( '平台类型', 'tanzanite-settings' ) . '<select id="tz-payment-gateway-type" style="width:100%;">';
            echo '                      <option value="">' . esc_html__( '无 (手动处理)', 'tanzanite-settings' ) . '</option>';
            echo '                      <option value="paypal">' . esc_html__( 'PayPal', 'tanzanite-settings' ) . '</option>';
            echo '                      <option value="stripe">' . esc_html__( 'Stripe', 'tanzanite-settings' ) . '</option>';
            echo '                      <option value="worldfirst">' . esc_html__( '万里汇 (WorldFirst)', 'tanzanite-settings' ) . '</option>';
            echo '                      <option value="payoneer">' . esc_html__( '派安盈 (Payoneer)', 'tanzanite-settings' ) . '</option>';
            echo '                  </select></label>';
            echo '              </div>';
            
            // PayPal 配置
            echo '              <div id="tz-gateway-paypal" class="tz-gateway-config" style="display:none;margin-top:16px;">';
            echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">PayPal 配置</h4>';
            echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Client ID', 'tanzanite-settings' ) . '<input type="text" id="tz-paypal-client-id" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Client Secret', 'tanzanite-settings' ) . '<input type="password" id="tz-paypal-client-secret" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-paypal-mode" style="height:30px;"><option value="sandbox">Sandbox (测试)</option><option value="live">Live (生产)</option></select></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Webhook ID', 'tanzanite-settings' ) . '<input type="text" id="tz-paypal-webhook-id" class="regular-text" /></label>';
            echo '                  </div>';
            echo '              </div>';
            
            // Stripe 配置
            echo '              <div id="tz-gateway-stripe" class="tz-gateway-config" style="display:none;margin-top:16px;">';
            echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">Stripe 配置</h4>';
            echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Publishable Key', 'tanzanite-settings' ) . '<input type="text" id="tz-stripe-publishable-key" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Secret Key', 'tanzanite-settings' ) . '<input type="password" id="tz-stripe-secret-key" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Webhook Secret', 'tanzanite-settings' ) . '<input type="password" id="tz-stripe-webhook-secret" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-stripe-mode" style="height:30px;"><option value="test">Test (测试)</option><option value="live">Live (生产)</option></select></label>';
            echo '                  </div>';
            echo '              </div>';
            
            // 万里汇 配置
            echo '              <div id="tz-gateway-worldfirst" class="tz-gateway-config" style="display:none;margin-top:16px;">';
            echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">万里汇 (WorldFirst) 配置</h4>';
            echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Merchant ID', 'tanzanite-settings' ) . '<input type="text" id="tz-worldfirst-merchant-id" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Key', 'tanzanite-settings' ) . '<input type="password" id="tz-worldfirst-api-key" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Secret', 'tanzanite-settings' ) . '<input type="password" id="tz-worldfirst-api-secret" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-worldfirst-mode" style="height:30px;"><option value="sandbox">Sandbox (测试)</option><option value="production">Production (生产)</option></select></label>';
            echo '                  </div>';
            echo '              </div>';
            
            // 派安盈 配置
            echo '              <div id="tz-gateway-payoneer" class="tz-gateway-config" style="display:none;margin-top:16px;">';
            echo '                  <h4 style="margin:12px 0 8px 0;font-size:13px;font-weight:600;color:#1f2937;">派安盈 (Payoneer) 配置</h4>';
            echo '                  <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px;">';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'Program ID', 'tanzanite-settings' ) . '<input type="text" id="tz-payoneer-program-id" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Username', 'tanzanite-settings' ) . '<input type="text" id="tz-payoneer-api-username" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( 'API Password', 'tanzanite-settings' ) . '<input type="password" id="tz-payoneer-api-password" class="regular-text" /></label>';
            echo '                      <label style="display:flex;flex-direction:column;gap:4px;">' . esc_html__( '环境模式', 'tanzanite-settings' ) . '<select id="tz-payoneer-mode" style="height:30px;"><option value="sandbox">Sandbox (测试)</option><option value="live">Live (生产)</option></select></label>';
            echo '                  </div>';
            echo '              </div>';
            echo '          </div>';
            
            echo '          <label>' . esc_html__( '自定义设置 (JSON)', 'tanzanite-settings' ) . '<textarea id="tz-payment-settings" rows="4" placeholder="{\"api_key\":\"...\"}"></textarea></label>';
            echo '          <label>' . esc_html__( 'Meta 信息 (JSON，可选)', 'tanzanite-settings' ) . '<textarea id="tz-payment-meta" rows="3"></textarea></label>';

            echo '          <div style="margin-top:16px;display:flex;gap:12px;">';
            echo '              <button class="button button-primary" id="tz-payment-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button" id="tz-payment-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '      </form>';
            echo '  </div>';

            // 添加独立的支付平台切换和媒体选择器脚本
            ?>
            <script type="text/javascript">
            (function() {
                console.log('Payment inline script loaded');
                
                var mediaUploader = null;
                
                function toggleGatewayConfig(gatewayType) {
                    console.log('Toggling gateway config for:', gatewayType);
                    
                    // 隐藏所有配置区域
                    var configs = document.querySelectorAll('.tz-gateway-config');
                    configs.forEach(function(el) {
                        el.style.display = 'none';
                    });
                    
                    // 显示选中的配置区域
                    if (gatewayType) {
                        var configEl = document.getElementById('tz-gateway-' + gatewayType);
                        if (configEl) {
                            configEl.style.display = 'block';
                            console.log('Showing config for:', gatewayType);
                        } else {
                            console.error('Config element not found for:', gatewayType);
                        }
                    }
                }
                
                function openMediaUploader(e) {
                    e.preventDefault();
                    console.log('Opening media uploader');
                    
                    // 检查 wp.media 是否存在
                    if (typeof wp === 'undefined' || typeof wp.media === 'undefined') {
                        alert('媒体库未加载，请刷新页面重试');
                        console.error('wp.media is not available');
                        return;
                    }
                    
                    if (mediaUploader) {
                        mediaUploader.open();
                        return;
                    }
                    
                    try {
                        mediaUploader = wp.media({
                            title: '选择支付图标',
                            button: { text: '使用此图片' },
                            multiple: false
                        });
                        
                        mediaUploader.on('select', function() {
                            var attachment = mediaUploader.state().get('selection').first().toJSON();
                            var iconUrlInput = document.getElementById('tz-payment-icon-url');
                            if (iconUrlInput) {
                                iconUrlInput.value = attachment.url;
                                updateIconPreview();
                            }
                        });
                        
                        mediaUploader.open();
                        console.log('Media uploader opened');
                    } catch (error) {
                        console.error('Failed to open media uploader:', error);
                        alert('打开媒体库失败: ' + error.message);
                    }
                }
                
                function updateIconPreview() {
                    var iconUrlInput = document.getElementById('tz-payment-icon-url');
                    var iconPreview = document.getElementById('tz-payment-icon-preview');
                    
                    if (!iconUrlInput || !iconPreview) return;
                    
                    var url = iconUrlInput.value;
                    if (url) {
                        iconPreview.innerHTML = '<img src="' + url + '" style="max-width:120px;max-height:60px;border:1px solid #ddd;padding:4px;background:#fff;" />';
                        iconPreview.style.display = 'block';
                    } else {
                        iconPreview.innerHTML = '<span style="color:#9ca3af;">无图标</span>';
                        iconPreview.style.display = 'none';
                    }
                }
                
                // 等待 DOM 加载完成
                if (document.readyState === 'loading') {
                    document.addEventListener('DOMContentLoaded', init);
                } else {
                    init();
                }
                
                function initTerminalsCheckboxes() {
                    var terminalsContainer = document.getElementById('tz-payment-terminals');
                    if (!terminalsContainer) {
                        console.warn('Terminals container not found');
                        return;
                    }
                    
                    var terminalOptions = [
                        { value: 'web', label: '网页端' },
                        { value: 'mobile', label: '移动端' },
                        { value: 'app', label: 'APP' },
                        { value: 'wechat', label: '微信小程序' }
                    ];
                    
                    terminalsContainer.innerHTML = '';
                    
                    terminalOptions.forEach(function(option) {
                        var label = document.createElement('label');
                        label.style.display = 'flex';
                        label.style.alignItems = 'center';
                        label.style.gap = '8px';
                        label.style.marginBottom = '8px';
                        
                        var checkbox = document.createElement('input');
                        checkbox.type = 'checkbox';
                        checkbox.value = option.value;
                        checkbox.className = 'terminal-checkbox';
                        
                        var span = document.createElement('span');
                        span.textContent = option.label;
                        
                        label.appendChild(checkbox);
                        label.appendChild(span);
                        terminalsContainer.appendChild(label);
                    });
                    
                    console.log('Terminals checkboxes initialized');
                }
                
                function init() {
                    console.log('Initializing payment page scripts');
                    
                    // 初始化终端复选框
                    initTerminalsCheckboxes();
                    
                    // 支付平台类型切换
                    var gatewayTypeSelect = document.getElementById('tz-payment-gateway-type');
                    if (gatewayTypeSelect) {
                        console.log('Gateway type select found');
                        gatewayTypeSelect.addEventListener('change', function() {
                            console.log('Gateway type changed to:', this.value);
                            toggleGatewayConfig(this.value);
                        });
                        
                        if (gatewayTypeSelect.value) {
                            toggleGatewayConfig(gatewayTypeSelect.value);
                        }
                    } else {
                        console.error('Gateway type select not found!');
                    }
                    
                    // 图标上传按钮
                    var iconUploadBtn = document.getElementById('tz-payment-icon-upload');
                    if (iconUploadBtn) {
                        console.log('Icon upload button found');
                        iconUploadBtn.addEventListener('click', openMediaUploader);
                    } else {
                        console.warn('Icon upload button not found');
                    }
                    
                    // 图标 URL 输入框
                    var iconUrlInput = document.getElementById('tz-payment-icon-url');
                    if (iconUrlInput) {
                        iconUrlInput.addEventListener('input', updateIconPreview);
                        updateIconPreview(); // 初始化预览
                    }
                }
            })();
            </script>
            <?php

            echo '</div>';
        }

        public function render_tax_rates(): void {
            $nonce = wp_create_nonce( 'wp_rest' );

            // 加载税率管理 JS
            wp_enqueue_script(
                'tz-tax-rates',
                plugins_url( 'assets/js/tax-rates.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'jquery' ),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-tax-rates',
                'TzTaxRatesConfig',
                array(
                    'listUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/tax-rates' ) ),
                    'singleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/tax-rates/' ) ),
                    'nonce'     => $nonce,
                )
            );

            echo '<div class="tz-settings-wrapper tz-tax-rates-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Tax Rates', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '配置税率模板，商品可关联一个或多个税率，前端下单时自动计算税费。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-tax-rate-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '税率模板列表', 'tanzanite-settings' ) . '</div>';
            echo '      <button class="button button-primary" id="tz-tax-rate-create">' . esc_html__( '新增税率', 'tanzanite-settings' ) . '</button>';
            echo '      <div style="overflow:auto;margin-top:16px;">';
            echo '          <table class="widefat fixed striped" id="tz-tax-rate-table" style="min-width:800px;">';
            echo '              <thead><tr>';
            foreach ( [ 'Name', 'Rate (%)', 'Region', 'Active', 'Actions' ] as $column ) {
                echo '<th>' . esc_html( $column ) . '</th>';
            }
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增税率', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-tax-rate-form" class="tz-tax-rate-form">';
            echo '          <input type="hidden" id="tz-tax-rate-id" />';
            echo '          <div class="tz-form-grid">';
            echo '              <label>' . esc_html__( '名称', 'tanzanite-settings' ) . '<input type="text" id="tz-tax-rate-name" required /></label>';
            echo '              <label>' . esc_html__( '税率 (%)', 'tanzanite-settings' ) . '<input type="number" step="0.0001" id="tz-tax-rate-rate" min="0" required /></label>';
            echo '              <label>' . esc_html__( '地区', 'tanzanite-settings' ) . '<input type="text" id="tz-tax-rate-region" /></label>';
            echo '              <label>' . esc_html__( '排序', 'tanzanite-settings' ) . '<input type="number" id="tz-tax-rate-sort" value="0" /></label>';
            echo '              <label>' . esc_html__( '启用', 'tanzanite-settings' ) . '<select id="tz-tax-rate-active"><option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option><option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option></select></label>';
            echo '          </div>';
            echo '          <label>' . esc_html__( '描述', 'tanzanite-settings' ) . '<textarea id="tz-tax-rate-description" rows="3"></textarea></label>';
            echo '          <label>' . esc_html__( 'Meta 信息 (JSON，可选)', 'tanzanite-settings' ) . '<textarea id="tz-tax-rate-meta" rows="3"></textarea></label>';
            echo '          <div style="margin-top:16px;display:flex;gap:12px;">';
            echo '              <button class="button button-primary" id="tz-tax-rate-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button" id="tz-tax-rate-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '      </form>';
            echo '  </div>';

            echo '</div>';
        }

        public function render_attributes(): void {
            // 直接输出配置和脚本到页面
            $nonce = wp_create_nonce( 'wp_rest' );
            $script_url = TANZANITE_PLUGIN_URL . 'assets/js/attributes.js?v=' . self::VERSION;
            ?>
            <script type="text/javascript">
            var TzAttributesConfig = {
                attrUrl: <?php echo wp_json_encode( rest_url( 'tanzanite/v1/attributes' ) ); ?>,
                singleUrl: <?php echo wp_json_encode( rest_url( 'tanzanite/v1/attributes/' ) ); ?>,
                nonce: <?php echo wp_json_encode( $nonce ); ?>
            };
            </script>
            <script type="text/javascript" src="<?php echo esc_url( $script_url ); ?>"></script>
            <?php
            echo '<div class="tz-settings-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>Attributes</h1>';
            echo '      <p>管理商品属性组与属性值，支持颜色、图标等多种类型</p>';
            echo '  </div>';
            echo '  <div id="tz-attr-notice" class="notice" style="display:none;"></div>';
            
            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">属性组列表</div>';
            echo '      <button type="button" class="button button-primary" id="tz-attr-create">新增属性组</button>';
            echo '      <div style="overflow:auto;margin-top:16px;"><table class="widefat fixed striped" id="tz-attr-table"><thead><tr><th>名称</th><th>类型</th><th>属性值数</th><th>筛选</th><th>SKU</th><th>状态</th><th>操作</th></tr></thead><tbody></tbody></table></div>';
            echo '  </div>';
            
            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">编辑 / 新增属性组</div>';
            echo '      <form id="tz-attr-form" onsubmit="return false;"><input type="hidden" id="tz-attr-id" />';
            echo '          <div class="tz-form-grid">';
            echo '              <label>名称<input type="text" id="tz-attr-name" required /></label>';
            echo '              <label>Slug<input type="text" id="tz-attr-slug" placeholder="自动生成" /></label>';
            echo '              <label>类型<select id="tz-attr-type"><option value="select">下拉选择</option><option value="color">色块</option><option value="image">图标</option></select></label>';
            echo '              <label>排序<input type="number" id="tz-attr-sort" value="0" /></label>';
            echo '          </div>';
            echo '          <div style="margin-top:12px;display:flex;gap:16px;flex-wrap:wrap;">';
            echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-filterable" checked /> 参与前端筛选</label>';
            echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-sku" checked /> 影响 SKU 组合</label>';
            echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-stock" /> 影响库存</label>';
            echo '              <label style="flex-direction:row;align-items:center;gap:8px;"><input type="checkbox" id="tz-attr-enabled" checked /> 启用</label>';
            echo '          </div>';
            echo '          <div style="margin-top:16px;display:flex;gap:12px;"><button class="button button-primary" id="tz-attr-save">保存</button><button class="button" id="tz-attr-reset" type="button">重置</button></div>';
            echo '      </form>';
            echo '  </div>';
            
            echo '  <div class="tz-settings-section" id="tz-attr-values-section" style="display:none;">';
            echo '      <div class="tz-section-title">属性值管理 - <span id="tz-current-attr-name"></span></div>';
            echo '      <div style="margin-bottom:16px;display:flex;gap:12px;align-items:end;">';
            echo '          <label style="flex:1;">名称<input type="text" id="tz-value-name" /></label>';
            echo '          <label style="flex:1;">Slug<input type="text" id="tz-value-slug" placeholder="自动生成" /></label>';
            echo '          <label style="flex:1;" id="tz-value-field-label">值<input type="text" id="tz-value-value" placeholder="如: #FF0000" /></label>';
            echo '          <button class="button button-primary" id="tz-value-add">添加</button>';
            echo '      </div>';
            echo '      <table class="widefat fixed striped" id="tz-values-table"><thead><tr><th>名称</th><th>Slug</th><th>预览</th><th>排序</th><th>状态</th><th>操作</th></tr></thead><tbody></tbody></table>';
            echo '  </div>';
            echo '</div>';
        }

        public function render_orders_bulk(): void {
            $nonce        = wp_create_nonce( 'wp_rest' );
            $bulk_url     = esc_url_raw( rest_url( 'tanzanite/v1/orders' ) );
            $status_list  = self::ALLOWED_ORDER_STATUSES;

            // 加载 Order Bulk JS
            wp_enqueue_script(
                'tz-order-bulk',
                plugins_url( 'assets/js/order-bulk.js', TANZANITE_LEGACY_MAIN_FILE ),
                array(),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-order-bulk',
                'TzOrderBulkConfig',
                array(
                    'nonce'   => $nonce,
                    'url'     => $bulk_url,
                    'strings' => array(
                        'invalidIds'    => __( '请输入有效的订单 ID。', 'tanzanite-settings' ),
                        'invalidStatus' => __( '请选择目标状态。', 'tanzanite-settings' ),
                        'done'          => __( '操作完成', 'tanzanite-settings' ),
                    ),
                )
            );

            echo '<div class="tz-settings-wrapper tz-orders-bulk">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Order Bulk Operations', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '这里可一次性更新多个订单状态或导出订单数据，方便客服与运营快速处理。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-order-bulk-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '批量更新订单状态', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-order-bulk-status" class="tz-bulk-form" style="display:grid;gap:12px;max-width:720px;">';
            echo '          <label>' . esc_html__( '订单 ID（逗号或换行分隔）', 'tanzanite-settings' ) . '<textarea rows="3" class="widefat" name="ids"></textarea></label>';
            echo '          <label>' . esc_html__( '目标状态', 'tanzanite-settings' ) . '<select name="status" class="widefat">';
            foreach ( $status_list as $status ) {
                echo '<option value="' . esc_attr( $status ) . '">' . esc_html( $status ) . '</option>';
            }
            echo '          </select></label>';
            echo '          <p class="description">' . esc_html__( '状态变更将按订单状态机校验，不符合流转的订单会自动跳过。', 'tanzanite-settings' ) . '</p>';
            echo '          <button class="button button-primary" type="submit">' . esc_html__( '批量修改状态', 'tanzanite-settings' ) . '</button>';
            echo '      </form>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '批量导出订单', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-order-bulk-export" class="tz-bulk-form" style="display:grid;gap:12px;max-width:720px;">';
            echo '          <label>' . esc_html__( '订单 ID（逗号或换行分隔）', 'tanzanite-settings' ) . '<textarea rows="3" class="widefat" name="ids"></textarea></label>';
            echo '          <p class="description">' . esc_html__( '导出内容包含金额、渠道、支付方式及时间戳，会提供 JSON 结果与 CSV 下载。', 'tanzanite-settings' ) . '</p>';
            echo '          <button class="button" type="submit">' . esc_html__( '开始导出', 'tanzanite-settings' ) . '</button>';
            echo '      </form>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '操作结果', 'tanzanite-settings' ) . '</div>';
            echo '      <pre id="tz-order-bulk-result" style="background:#f6f7f7;border:1px solid #ccd0d4;padding:12px;max-height:320px;overflow:auto;"></pre>';
            echo '  </div>';

            echo '</div>';

            $config_js  = wp_json_encode(
                [
                    'url'      => $bulk_url,
                    'nonce'    => $nonce,
                    'statuses' => array_values( $status_list ),
                ],
                JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES
            );

            $strings_js = wp_json_encode(
                [
                    'invalidIds'   => __( '请先填写有效的订单 ID。', 'tanzanite-settings' ),
                    'invalidStatus'=> __( '请选择目标状态。', 'tanzanite-settings' ),
                    'done'         => __( '操作完成', 'tanzanite-settings' ),
                ],
                JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES
            );

 
        }

        public function render_shipping_templates(): void {
            $nonce = wp_create_nonce( 'wp_rest' );

            // 加载配送模板管理 JS
            wp_enqueue_script(
                'tz-shipping-templates',
                plugins_url( 'assets/js/shipping-templates.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-admin-common' ),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-shipping-templates',
                'TzShippingConfig',
                array(
                    'listUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/shipping-templates' ) ),
                    'singleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/shipping-templates/' ) ),
                    'nonce'     => $nonce,
                )
            );

            echo '<div class="tz-settings-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Shipping Templates', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '定义配送规则、包邮策略与配送时效说明。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-shipping-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '配送模板列表', 'tanzanite-settings' ) . '</div>';
            echo '      <button class="button button-primary" id="tz-shipping-create">' . esc_html__( '新增模板', 'tanzanite-settings' ) . '</button>';
            echo '      <button class="button" id="tz-shipping-export" style="margin-left:8px;">' . esc_html__( '导出 JSON', 'tanzanite-settings' ) . '</button>';
            echo '      <div style="overflow:auto;margin-top:16px;">';
            echo '          <table class="widefat fixed striped" id="tz-shipping-table">';
            echo '              <thead><tr>';
            echo '                  <th style="width:25%;">' . esc_html__( '模板名称', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:35%;">' . esc_html__( '描述', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:10%;">' . esc_html__( '规则数', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:10%;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:20%;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
            echo '              </tr></thead><tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '编辑 / 新增配送模板', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-shipping-form">';
            echo '          <input type="hidden" id="tz-shipping-id" />';
            echo '          <div class="tz-form-grid">';
            echo '              <label>' . esc_html__( '模板名称', 'tanzanite-settings' ) . '<input type="text" id="tz-shipping-name" required /></label>';
            echo '              <label>' . esc_html__( '状态', 'tanzanite-settings' ) . '<select id="tz-shipping-active"><option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option><option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option></select></label>';
            echo '          </div>';
            echo '          <label>' . esc_html__( '描述', 'tanzanite-settings' ) . '<textarea id="tz-shipping-description" rows="2"></textarea></label>';
            
            echo '          <div style="margin-top:20px;">';
            echo '              <strong>' . esc_html__( '配送规则', 'tanzanite-settings' ) . '</strong>';
            echo '              <p class="description">' . esc_html__( '按重量、金额、件数等条件设置运费。支持多条规则，系统将按优先级匹配。', 'tanzanite-settings' ) . '</p>';
            echo '              <div id="tz-shipping-rules-list" style="margin-top:12px;"></div>';
            echo '              <button type="button" class="button" id="tz-shipping-add-rule" style="margin-top:12px;">' . esc_html__( '添加规则', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';

            echo '          <div style="margin-top:16px;display:flex;gap:12px;">';
            echo '              <button class="button button-primary" id="tz-shipping-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button" id="tz-shipping-reset" type="button">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '      </form>';
            echo '  </div>';

            echo '</div>';
        }

        public function render_carriers(): void {
            $active_tab = isset( $_GET['tab'] ) ? sanitize_key( $_GET['tab'] ) : 'list'; // phpcs:ignore WordPress.Security.NonceVerification.Recommended
            $nonce      = wp_create_nonce( 'wp_rest' );
            $list_url   = esc_url_raw( rest_url( 'tanzanite/v1/carriers' ) );
            $single_url = esc_url_raw( rest_url( 'tanzanite/v1/carriers/' ) );
            $can_manage = current_user_can( 'tanz_manage_shipping' );

            if ( 'list' === $active_tab ) {
                wp_enqueue_script(
                    'tz-carriers',
                    plugins_url( 'assets/js/carriers.js', TANZANITE_LEGACY_MAIN_FILE ),
                    array( 'tz-admin-common' ),
                    self::VERSION,
                    true
                );

                wp_localize_script(
                    'tz-carriers',
                    'TzCarriersConfig',
                    array(
                        'nonce'     => $nonce,
                        'listUrl'   => $list_url,
                        'singleUrl' => $single_url,
                        'canManage' => $can_manage,
                    )
                );

                echo '<div class="tz-settings-wrapper">';
                echo '  <div class="tz-settings-header">';
                echo '      <h1>' . esc_html__( 'Carriers & Tracking', 'tanzanite-settings' ) . '</h1>';
                echo '      <p>' . esc_html__( '管理物流公司信息和追踪配置。', 'tanzanite-settings' ) . '</p>';
                echo '  </div>';

                echo '  <nav class="nav-tab-wrapper">';
                echo '      <a href="?page=tanzanite-settings-carriers&tab=list" class="nav-tab nav-tab-active">' . esc_html__( '物流公司管理', 'tanzanite-settings' ) . '</a>';
                echo '      <a href="?page=tanzanite-settings-carriers&tab=config" class="nav-tab">' . esc_html__( 'API 配置', 'tanzanite-settings' ) . '</a>';
                echo '  </nav>';

                echo '  <div id="tz-carriers-notice" class="notice" style="display:none;margin-top:16px;"></div>';

                echo '  <div class="tz-settings-section" style="margin-top:24px;">';
                echo '      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;">';
                echo '          <h2>' . esc_html__( '物流公司列表', 'tanzanite-settings' ) . '</h2>';
                if ( $can_manage ) {
                    echo '          <button class="button button-primary" id="tz-carriers-create">' . esc_html__( '新建物流公司', 'tanzanite-settings' ) . '</button>';
                }
                echo '      </div>';

                echo '      <table class="widefat fixed striped" id="tz-carriers-table">';
                echo '          <thead>';
                echo '              <tr>';
                echo '                  <th style="width:120px;">' . esc_html__( '编码', 'tanzanite-settings' ) . '</th>';
                echo '                  <th>' . esc_html__( '名称', 'tanzanite-settings' ) . '</th>';
                echo '                  <th style="width:150px;">' . esc_html__( '联系人', 'tanzanite-settings' ) . '</th>';
                echo '                  <th style="width:120px;">' . esc_html__( '电话', 'tanzanite-settings' ) . '</th>';
                echo '                  <th style="width:200px;">' . esc_html__( '追踪 URL', 'tanzanite-settings' ) . '</th>';
                echo '                  <th style="width:80px;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
                echo '                  <th style="width:120px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
                echo '              </tr>';
                echo '          </thead>';
                echo '          <tbody></tbody>';
                echo '      </table>';
                echo '  </div>';

                if ( $can_manage ) {
                    echo '  <div class="tz-settings-section" style="margin-top:24px;">';
                    echo '      <h2>' . esc_html__( '新建/编辑物流公司', 'tanzanite-settings' ) . '</h2>';
                    echo '      <form id="tz-carriers-form" style="max-width:800px;">';
                    echo '          <input type="hidden" id="tz-carrier-id" />';
                    echo '          <div class="tz-form-grid">';
                    echo '              <label><strong>' . esc_html__( '编码', 'tanzanite-settings' ) . ' *</strong>';
                    echo '                  <input type="text" id="tz-carrier-code" class="regular-text" required placeholder="sf_express" />';
                    echo '              </label>';
                    echo '              <label><strong>' . esc_html__( '名称', 'tanzanite-settings' ) . ' *</strong>';
                    echo '                  <input type="text" id="tz-carrier-name" class="regular-text" required placeholder="顺丰速运" />';
                    echo '              </label>';
                    echo '              <label><strong>' . esc_html__( '联系人', 'tanzanite-settings' ) . '</strong>';
                    echo '                  <input type="text" id="tz-carrier-contact-person" class="regular-text" placeholder="张三" />';
                    echo '              </label>';
                    echo '              <label><strong>' . esc_html__( '联系电话', 'tanzanite-settings' ) . '</strong>';
                    echo '                  <input type="text" id="tz-carrier-contact-phone" class="regular-text" placeholder="400-111-1111" />';
                    echo '              </label>';
                    echo '          </div>';
                    echo '          <label style="display:block;margin-top:12px;"><strong>' . esc_html__( '追踪 URL 模板', 'tanzanite-settings' ) . '</strong>';
                    echo '              <input type="url" id="tz-carrier-tracking-url" class="regular-text" placeholder="https://www.sf-express.com/cn/sc/dynamic_function/waybill/#search_waybill={{tracking_number}}" />';
                    echo '              <p class="description">' . esc_html__( '使用 {{tracking_number}} 作为运单号占位符', 'tanzanite-settings' ) . '</p>';
                    echo '          </label>';
                    echo '          <label style="display:block;margin-top:12px;"><strong>' . esc_html__( '服务地区', 'tanzanite-settings' ) . '</strong>';
                    echo '              <input type="text" id="tz-carrier-service-regions" class="regular-text" placeholder="中国大陆, 香港, 台湾（逗号分隔）" />';
                    echo '          </label>';
                    echo '          <div class="tz-form-grid" style="margin-top:12px;">';
                    echo '              <label><strong>' . esc_html__( '状态', 'tanzanite-settings' ) . '</strong>';
                    echo '                  <select id="tz-carrier-is-active" class="regular-text">';
                    echo '                      <option value="1">' . esc_html__( '启用', 'tanzanite-settings' ) . '</option>';
                    echo '                      <option value="0">' . esc_html__( '禁用', 'tanzanite-settings' ) . '</option>';
                    echo '                  </select>';
                    echo '              </label>';
                    echo '              <label><strong>' . esc_html__( '排序', 'tanzanite-settings' ) . '</strong>';
                    echo '                  <input type="number" id="tz-carrier-sort-order" class="regular-text" value="0" />';
                    echo '              </label>';
                    echo '          </div>';
                    echo '          <label style="display:block;margin-top:12px;"><strong>' . esc_html__( 'Meta (JSON)', 'tanzanite-settings' ) . '</strong>';
                    echo '              <textarea id="tz-carrier-meta" class="large-text code" rows="4" placeholder=\'{"key": "value"}\'></textarea>';
                    echo '          </label>';
                    echo '          <div style="margin-top:16px;">';
                    echo '              <button type="submit" class="button button-primary" id="tz-carriers-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
                    echo '              <button type="button" class="button" id="tz-carriers-reset">' . esc_html__( '重置', 'tanzanite-settings' ) . '</button>';
                    echo '          </div>';
                    echo '      </form>';
                    echo '  </div>';
                }

                echo '</div>';

            } elseif ( 'config' === $active_tab ) {
                $tracking_option = get_option( 'tanzanite_tracking_settings', [] );
                $provider        = $tracking_option['provider'] ?? '17track';
                $settings        = $tracking_option['settings'][ $provider ] ?? [];

                wp_enqueue_script(
                    'tz-carriers-config',
                    plugins_url( 'assets/js/carriers-config.js', TANZANITE_LEGACY_MAIN_FILE ),
                    array( 'tz-admin-common' ),
                    self::VERSION,
                    true
                );

                wp_localize_script(
                    'tz-carriers-config',
                    'TzCarriersConfigPage',
                    array(
                        'providers'       => self::TRACKING_PROVIDERS,
                        'currentProvider' => $provider,
                        'currentSettings' => $settings,
                        'testUrl'         => admin_url( 'admin-post.php?action=tanz_test_tracking' ),
                    )
                );

                echo '<div class="tz-settings-wrapper">';
                echo '  <div class="tz-settings-header">';
                echo '      <h1>' . esc_html__( 'Carriers & Tracking', 'tanzanite-settings' ) . '</h1>';
                echo '      <p>' . esc_html__( '管理物流公司信息和追踪配置。', 'tanzanite-settings' ) . '</p>';
                echo '  </div>';

                echo '  <nav class="nav-tab-wrapper">';
                echo '      <a href="?page=tanzanite-settings-carriers&tab=list" class="nav-tab">' . esc_html__( '物流公司管理', 'tanzanite-settings' ) . '</a>';
                echo '      <a href="?page=tanzanite-settings-carriers&tab=config" class="nav-tab nav-tab-active">' . esc_html__( 'API 配置', 'tanzanite-settings' ) . '</a>';
                echo '  </nav>';

                echo '  <div id="tz-carriers-config-notice" class="notice" style="display:none;margin-top:16px;"></div>';

                echo '  <div class="tz-settings-section" style="margin-top:24px;">';
                echo '      <h2>' . esc_html__( '物流追踪 API 配置', 'tanzanite-settings' ) . '</h2>';
                echo '      <form method="post" action="' . esc_url( admin_url( 'admin-post.php' ) ) . '" id="tz-tracking-config-form" style="max-width:600px;">';
                wp_nonce_field( 'tanz_tracking_settings' );
                echo '          <input type="hidden" name="action" value="tanz_save_tracking_settings" />';
                echo '          <label style="display:block;margin-bottom:16px;"><strong>' . esc_html__( '追踪服务商', 'tanzanite-settings' ) . '</strong>';
                echo '              <select name="provider" id="tz-tracking-provider" class="regular-text" style="display:block;margin-top:4px;">';
                foreach ( self::TRACKING_PROVIDERS as $key => $config ) {
                    $selected = ( $key === $provider ) ? ' selected' : '';
                    echo '                  <option value="' . esc_attr( $key ) . '"' . $selected . '>' . esc_html( $config['label'] ) . '</option>';
                }
                echo '              </select>';
                echo '          </label>';
                echo '          <div id="tz-tracking-fields"></div>';
                echo '          <div style="margin-top:16px;">';
                echo '              <button type="submit" class="button button-primary" id="tz-tracking-save">' . esc_html__( '保存配置', 'tanzanite-settings' ) . '</button>';
                echo '              <button type="button" class="button" id="tz-tracking-test">' . esc_html__( '测试连接', 'tanzanite-settings' ) . '</button>';
                echo '          </div>';
                echo '      </form>';
                echo '  </div>';

                echo '</div>';
            }
        }

        /**
         * 渲染会员档案列表页面
         */
        public function render_member_profiles(): void {
            if ( ! current_user_can( 'manage_options' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            $nonce = wp_create_nonce( 'wp_rest' );

            // 加载 Member Profiles JS
            wp_enqueue_script(
                'tz-member-profiles',
                plugins_url( 'assets/js/member-profiles.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-admin-common' ),
                self::VERSION,
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-member-profiles',
                'TzMemberProfilesConfig',
                array(
                    'nonce'      => $nonce,
                    'listUrl'    => esc_url_raw( rest_url( 'tanzanite/v1/members' ) ),
                    'singleUrl'  => esc_url_raw( rest_url( 'tanzanite/v1/members/' ) ),
                    'exportUrl'  => esc_url_raw( admin_url( 'admin-ajax.php?action=tanz_export_members' ) ),
                    'i18n'       => array(
                        'noData'          => __( '暂无会员数据', 'tanzanite-settings' ),
                        'loadFailed'      => __( '加载会员列表失败', 'tanzanite-settings' ),
                        'saveSuccess'     => __( '会员信息已更新', 'tanzanite-settings' ),
                        'saveFailed'      => __( '保存失败', 'tanzanite-settings' ),
                        'deleteConfirm'   => __( '确定删除该会员档案？', 'tanzanite-settings' ),
                        'deleteSuccess'   => __( '会员档案已删除', 'tanzanite-settings' ),
                        'selectMember'    => __( '请先选择要操作的会员', 'tanzanite-settings' ),
                        'exportSuccess'   => __( '导出成功', 'tanzanite-settings' ),
                        'importSuccess'   => __( '导入成功', 'tanzanite-settings' ),
                    ),
                )
            );

            echo '<div class="tz-settings-wrapper tz-members-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Member Profiles', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '管理会员档案、积分余额和营销偏好。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-member-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '筛选条件', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-member-filters" class="tz-member-filters" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px;max-width:1200px;">';
            echo '          <label>' . esc_html__( '关键词（姓名/邮箱/手机）', 'tanzanite-settings' ) . '<input type="text" name="search" class="widefat" /></label>';
            echo '          <label>' . esc_html__( '最低积分', 'tanzanite-settings' ) . '<input type="number" name="min_points" class="widefat" min="0" /></label>';
            echo '          <label>' . esc_html__( '每页条数', 'tanzanite-settings' ) . '<select name="per_page" class="widefat"><option value="20">20</option><option value="50">50</option><option value="100">100</option></select></label>';
            echo '      </form>';
            echo '      <div style="display:flex;gap:12px;margin-top:12px;">';
            echo '          <button class="button button-primary" id="tz-member-refresh">' . esc_html__( '刷新列表', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button" id="tz-member-reset">' . esc_html__( '重置筛选', 'tanzanite-settings' ) . '</button>';
            echo '          <button class="button" id="tz-member-export">' . esc_html__( '导出 CSV', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section">';
            echo '      <div class="tz-section-title">' . esc_html__( '会员列表', 'tanzanite-settings' ) . '</div>';
            echo '      <div style="overflow:auto;margin-top:16px;">';
            echo '          <table class="widefat fixed striped" id="tz-member-table" style="min-width:960px;">';
            echo '              <thead><tr>';
            echo '                  <th style="width:60px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:120px;">' . esc_html__( '用户名', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:150px;">' . esc_html__( '姓名', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:180px;">' . esc_html__( '邮箱', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:120px;">' . esc_html__( '手机', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:80px;">' . esc_html__( '积分', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:100px;">' . esc_html__( '等级', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:140px;">' . esc_html__( '注册时间', 'tanzanite-settings' ) . '</th>';
            echo '                  <th style="width:120px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
            echo '              </tr></thead>';
            echo '              <tbody></tbody>';
            echo '          </table>';
            echo '      </div>';
            echo '      <div class="tablenav bottom" style="display:flex;justify-content:space-between;align-items:center;padding:12px 0;">';
            echo '          <div class="tablenav-pages">';
            echo '              <span class="displaying-num" id="tz-member-page-info"></span>';
            echo '              <span class="pagination-links">';
            echo '                  <button class="button" id="tz-member-prev" disabled>' . esc_html__( '上一页', 'tanzanite-settings' ) . '</button>';
            echo '                  <button class="button" id="tz-member-next" disabled>' . esc_html__( '下一页', 'tanzanite-settings' ) . '</button>';
            echo '              </span>';
            echo '          </div>';
            echo '      </div>';
            echo '  </div>';

            echo '  <div class="tz-settings-section" id="tz-member-detail-section" style="display:none;">';
            echo '      <div class="tz-section-title">' . esc_html__( '会员详情', 'tanzanite-settings' ) . '</div>';
            echo '      <form id="tz-member-detail" style="max-width:800px;margin-top:16px;">';
            echo '          <input type="hidden" id="tz-member-id" />';
            echo '          <table class="form-table">';
            echo '              <tr><th><label for="tz-member-username">' . esc_html__( '用户名', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-username" class="regular-text" readonly /></td></tr>';
            echo '              <tr><th><label for="tz-member-email">' . esc_html__( '邮箱', 'tanzanite-settings' ) . '</label></th><td><input type="email" id="tz-member-email" class="regular-text" readonly /></td></tr>';
            echo '              <tr><th><label for="tz-member-fullname">' . esc_html__( '姓名', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-fullname" class="regular-text" /></td></tr>';
            echo '              <tr><th><label for="tz-member-phone">' . esc_html__( '手机', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-phone" class="regular-text" /></td></tr>';
            echo '              <tr><th><label for="tz-member-country">' . esc_html__( '国家/地区', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-country" class="regular-text" /></td></tr>';
            echo '              <tr><th><label for="tz-member-address">' . esc_html__( '地址', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-address" class="regular-text" /></td></tr>';
            echo '              <tr><th><label for="tz-member-brand">' . esc_html__( '品牌', 'tanzanite-settings' ) . '</label></th><td><input type="text" id="tz-member-brand" class="regular-text" /></td></tr>';
            echo '              <tr><th><label for="tz-member-points">' . esc_html__( '积分', 'tanzanite-settings' ) . '</label></th><td><input type="number" id="tz-member-points" class="regular-text" min="0" /></td></tr>';
            echo '              <tr><th><label for="tz-member-marketing">' . esc_html__( '营销订阅', 'tanzanite-settings' ) . '</label></th><td><label><input type="checkbox" id="tz-member-marketing" /> ' . esc_html__( '接收营销信息', 'tanzanite-settings' ) . '</label></td></tr>';
            echo '              <tr><th><label for="tz-member-notes">' . esc_html__( '备注', 'tanzanite-settings' ) . '</label></th><td><textarea id="tz-member-notes" rows="5" class="large-text"></textarea></td></tr>';
            echo '          </table>';
            echo '          <div style="margin-top:16px;">';
            echo '              <button type="button" class="button button-primary" id="tz-member-save">' . esc_html__( '保存', 'tanzanite-settings' ) . '</button>';
            echo '              <button type="button" class="button" id="tz-member-cancel">' . esc_html__( '取消', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '      </form>';
            echo '  </div>';

            echo '</div>';
        }

        /**
         * 渲染礼品卡和优惠券管理页面
         */
        public function render_rewards(): void {
            if ( ! current_user_can( 'manage_options' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            $nonce = wp_create_nonce( 'wp_rest' );

            // 加载媒体上传器
            wp_enqueue_media();
            
            // 加载 Rewards JS
            wp_enqueue_script(
                'tz-rewards',
                plugins_url( 'assets/js/rewards.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-admin-common' ),
                self::VERSION . '.fixed.' . time(),
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-rewards',
                'TzRewardsConfig',
                array(
                    'nonce'              => $nonce,
                    'couponsListUrl'     => esc_url_raw( rest_url( 'tanzanite/v1/coupons' ) ),
                    'couponsSingleUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/coupons/' ) ),
                    'giftcardsListUrl'   => esc_url_raw( rest_url( 'tanzanite/v1/giftcards' ) ),
                    'giftcardsSingleUrl' => esc_url_raw( rest_url( 'tanzanite/v1/giftcards/' ) ),
                    'transactionsUrl'    => esc_url_raw( rest_url( 'tanzanite/v1/rewards-transactions' ) ),
                    'i18n'               => array(
                        'noData'           => __( '暂无数据', 'tanzanite-settings' ),
                        'loadFailed'       => __( '加载失败', 'tanzanite-settings' ),
                        'saveSuccess'      => __( '保存成功', 'tanzanite-settings' ),
                        'saveFailed'       => __( '保存失败', 'tanzanite-settings' ),
                        'deleteConfirm'    => __( '确定删除？', 'tanzanite-settings' ),
                        'deleteSuccess'    => __( '删除成功', 'tanzanite-settings' ),
                        'generateCode'     => __( '生成代码', 'tanzanite-settings' ),
                        'codeGenerated'    => __( '代码已生成', 'tanzanite-settings' ),
                        'invalidAmount'    => __( '金额无效', 'tanzanite-settings' ),
                        'invalidPoints'    => __( '积分无效', 'tanzanite-settings' ),
                    ),
                )
            );

            echo '<div class="tz-settings-wrapper tz-rewards-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Gift Cards & Coupons', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '管理优惠券、礼品卡和积分兑换。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-rewards-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            // 标签页导航
            echo '  <div class="tz-tabs-nav" style="margin-bottom:20px;border-bottom:1px solid #ccc;">';
            echo '      <button class="tz-tab-btn active" data-tab="coupons">' . esc_html__( '优惠券', 'tanzanite-settings' ) . '</button>';
            echo '      <button class="tz-tab-btn" data-tab="giftcards">' . esc_html__( '礼品卡', 'tanzanite-settings' ) . '</button>';
            echo '      <button class="tz-tab-btn" data-tab="transactions">' . esc_html__( '交易记录', 'tanzanite-settings' ) . '</button>';
            echo '      <button class="tz-tab-btn" data-tab="redeem-settings">' . esc_html__( '积分兑换设置', 'tanzanite-settings' ) . '</button>';
            echo '  </div>';

            // 优惠券标签页
            echo '  <div class="tz-tab-content" id="tz-tab-coupons">';
            echo '      <div class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( '优惠券列表', 'tanzanite-settings' ) . '</div>';
            echo '          <div style="margin:16px 0;">';
            echo '              <button class="button button-primary" id="tz-coupon-add">' . esc_html__( '添加优惠券', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button" id="tz-coupon-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '          <div style="overflow:auto;">';
            echo '              <table class="widefat fixed striped" id="tz-coupon-table">';
            echo '                  <thead><tr>';
            echo '                      <th style="width:60px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:120px;">' . esc_html__( '代码', 'tanzanite-settings' ) . '</th>';
            echo '                      <th>' . esc_html__( '标题', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '折扣类型', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:80px;">' . esc_html__( '折扣', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '使用次数', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:160px;">' . esc_html__( '过期时间', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:80px;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:160px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
            echo '                  </tr></thead>';
            echo '                  <tbody></tbody>';
            echo '              </table>';
            echo '          </div>';
            echo '      </div>';
            echo '  </div>';

            // 礼品卡标签页
            echo '  <div class="tz-tab-content" id="tz-tab-giftcards" style="display:none;">';
            echo '      <div class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( '礼品卡列表', 'tanzanite-settings' ) . '</div>';
            echo '          <div style="margin:16px 0;">';
            echo '              <button class="button button-primary" id="tz-giftcard-add">' . esc_html__( '添加礼品卡', 'tanzanite-settings' ) . '</button>';
            echo '              <button class="button" id="tz-giftcard-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '          <div style="overflow:auto;">';
            echo '              <table class="widefat fixed striped" id="tz-giftcard-table">';
            echo '                  <thead><tr>';
            echo '                      <th style="width:80px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:140px;">' . esc_html__( '卡号', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '余额', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '原始金额', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '积分消耗', 'tanzanite-settings' ) . '</th>';
            echo '                      <th>' . esc_html__( '持有人', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '状态', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:140px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
            echo '                  </tr></thead>';
            echo '                  <tbody></tbody>';
            echo '              </table>';
            echo '          </div>';
            echo '      </div>';
            echo '  </div>';

            // 交易记录标签页
            echo '  <div class="tz-tab-content" id="tz-tab-transactions" style="display:none;">';
            echo '      <div class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( '积分交易记录', 'tanzanite-settings' ) . '</div>';
            echo '          <div style="margin:16px 0;">';
            echo '              <button class="button" id="tz-transaction-refresh">' . esc_html__( '刷新', 'tanzanite-settings' ) . '</button>';
            echo '          </div>';
            echo '          <div style="overflow:auto;">';
            echo '              <table class="widefat fixed striped" id="tz-transaction-table">';
            echo '                  <thead><tr>';
            echo '                      <th style="width:80px;">' . esc_html__( 'ID', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:120px;">' . esc_html__( '用户', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '类型', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '动作', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '积分变化', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:100px;">' . esc_html__( '金额变化', 'tanzanite-settings' ) . '</th>';
            echo '                      <th>' . esc_html__( '备注', 'tanzanite-settings' ) . '</th>';
            echo '                      <th style="width:160px;">' . esc_html__( '时间', 'tanzanite-settings' ) . '</th>';
            echo '                  </tr></thead>';
            echo '                  <tbody></tbody>';
            echo '              </table>';
            echo '          </div>';
            echo '      </div>';
            echo '  </div>';

            // 积分兑换设置标签页
            echo '  <div class="tz-tab-content" id="tz-tab-redeem-settings" style="display:none;">';
            echo '      <div class="tz-settings-section">';
            echo '          <div class="tz-section-title">' . esc_html__( '积分兑换礼品卡设置', 'tanzanite-settings' ) . '</div>';
            echo '          <form method="post" action="' . esc_url( admin_url( 'admin-post.php' ) ) . '" id="tz-redeem-settings-form">';
            wp_nonce_field( 'tz_save_redeem_settings', 'tz_redeem_nonce' );
            echo '              <input type="hidden" name="action" value="tz_save_redeem_settings" />';
            echo '              <table class="form-table">';
            echo '                  <tr>';
            echo '                      <th scope="row"><label for="redeem_enabled">' . esc_html__( '启用积分兑换', 'tanzanite-settings' ) . '</label></th>';
            echo '                      <td>';
            echo '                          <label><input type="checkbox" name="redeem_enabled" id="redeem_enabled" value="1" ' . checked( get_option( 'tz_redeem_enabled', '1' ), '1', false ) . ' /> ' . esc_html__( '允许用户使用积分兑换礼品卡', 'tanzanite-settings' ) . '</label>';
            echo '                          <p class="description">' . esc_html__( '⚠️ 前端 Nuxt 页面需要调用 REST API 实现兑换功能', 'tanzanite-settings' ) . '</p>';
            echo '                          <p class="description" style="color:#0073aa;"><strong>💡 提示：</strong>礼品卡的面值和积分价格在"礼品卡"页面设置</p>';
            echo '                      </td>';
            echo '                  </tr>';
            echo '                  <tr>';
            echo '                      <th scope="row"><label for="redeem_card_expiry_days">' . esc_html__( '礼品卡有效期', 'tanzanite-settings' ) . '</label></th>';
            echo '                      <td>';
            echo '                          <input type="number" name="redeem_card_expiry_days" id="redeem_card_expiry_days" value="' . esc_attr( get_option( 'tz_redeem_card_expiry_days', '365' ) ) . '" class="regular-text" min="0" step="1" />';
            echo '                          <p class="description">' . esc_html__( '兑换的礼品卡有效期天数（0 表示永久有效）', 'tanzanite-settings' ) . '</p>';
            echo '                      </td>';
            echo '                  </tr>';
            echo '                  <tr>';
            echo '                      <th scope="row"><label for="giftcard_cover_design">' . esc_html__( '礼品卡封面设计', 'tanzanite-settings' ) . '</label></th>';
            echo '                      <td>';
            echo '                          <div style="margin-bottom:12px;">';
            echo '                              <label><input type="radio" name="giftcard_cover_type" value="default" ' . checked( get_option( 'tz_giftcard_cover_type', 'default' ), 'default', false ) . ' /> ' . esc_html__( '使用默认封面', 'tanzanite-settings' ) . '</label><br>';
            echo '                              <label><input type="radio" name="giftcard_cover_type" value="custom" ' . checked( get_option( 'tz_giftcard_cover_type', 'default' ), 'custom', false ) . ' /> ' . esc_html__( '自定义封面图片', 'tanzanite-settings' ) . '</label><br>';
            echo '                              <label><input type="radio" name="giftcard_cover_type" value="template" ' . checked( get_option( 'tz_giftcard_cover_type', 'default' ), 'template', false ) . ' /> ' . esc_html__( '使用封面模板', 'tanzanite-settings' ) . '</label>';
            echo '                          </div>';
            echo '                          <div id="tz-giftcard-custom-cover" style="display:' . ( get_option( 'tz_giftcard_cover_type', 'default' ) === 'custom' ? 'block' : 'none' ) . ';margin-bottom:12px;">';
            echo '                              <input type="url" name="giftcard_cover_url" id="giftcard_cover_url" value="' . esc_attr( get_option( 'tz_giftcard_cover_url', '' ) ) . '" class="regular-text" placeholder="https://example.com/giftcard-cover.jpg" />';
            echo '                              <button type="button" class="button" id="tz-upload-cover">' . esc_html__( '上传图片', 'tanzanite-settings' ) . '</button>';
            echo '                              <p class="description">' . esc_html__( '建议尺寸：400x250px，支持 JPG、PNG 格式', 'tanzanite-settings' ) . '</p>';
            echo '                          </div>';
            echo '                          <div id="tz-giftcard-template-cover" style="display:' . ( get_option( 'tz_giftcard_cover_type', 'default' ) === 'template' ? 'block' : 'none' ) . ';margin-bottom:12px;">';
            echo '                              <select name="giftcard_cover_template" id="giftcard_cover_template" class="regular-text">';
            $selected_template = get_option( 'tz_giftcard_cover_template', 'elegant' );
            $templates = array(
                'elegant' => __( '优雅风格 - 深蓝渐变', 'tanzanite-settings' ),
                'festive' => __( '节日风格 - 红金配色', 'tanzanite-settings' ),
                'modern' => __( '现代风格 - 简约灰白', 'tanzanite-settings' ),
                'luxury' => __( '奢华风格 - 黑金配色', 'tanzanite-settings' ),
                'spring' => __( '春季风格 - 清新绿色', 'tanzanite-settings' ),
            );
            foreach ( $templates as $value => $label ) {
                echo '                                  <option value="' . esc_attr( $value ) . '" ' . selected( $selected_template, $value, false ) . '>' . esc_html( $label ) . '</option>';
            }
            echo '                              </select>';
            echo '                              <p class="description">' . esc_html__( '选择预设的封面模板样式', 'tanzanite-settings' ) . '</p>';
            echo '                          </div>';
            echo '                          <div id="tz-giftcard-cover-preview" style="margin-top:12px;padding:12px;border:1px solid #ddd;border-radius:4px;background:#f9f9f9;">';
            echo '                              <strong>' . esc_html__( '封面预览：', 'tanzanite-settings' ) . '</strong>';
            echo '                              <div id="tz-cover-preview-area" style="margin-top:8px;width:200px;height:125px;border:1px solid #ccc;border-radius:8px;background:#fff;display:flex;align-items:center;justify-content:center;color:#666;font-size:12px;">';
            echo '                                  ' . esc_html__( '礼品卡封面预览', 'tanzanite-settings' );
            echo '                              </div>';
            echo '                          </div>';
            echo '                          <p class="description" style="color:#d63638;"><strong>📌 前端 Nuxt 提示：</strong>通过 REST API 获取封面配置，在礼品卡显示时使用</p>';
            echo '                      </td>';
            echo '                  </tr>';
            echo '              </table>';
            echo '              <p class="submit">';
            echo '                  <button type="submit" class="button button-primary">' . esc_html__( '保存设置', 'tanzanite-settings' ) . '</button>';
            echo '              </p>';
            echo '          </form>';
            echo '      </div>';
            
            // API 接口文档说明
            echo '      <div class="tz-settings-section" style="margin-top:20px;background:#f0f6fc;border-left:4px solid #0073aa;padding:15px;">';
            echo '          <h3 style="margin-top:0;">📖 前端 Nuxt 对接指南</h3>';
            echo '          <h4>1. 获取兑换配置</h4>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">GET /wp-json/tanzanite/v1/redeem/config</pre>';
            echo '          <p><strong>返回示例：</strong></p>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "enabled": true,
  "exchange_rate": 100,
  "min_points": 1000,
  "max_value_per_day": 500,
  "card_expiry_days": 365,
  "preset_values": [10, 50, 100, 200, 500]
}</pre>';
            echo '          <h4>2. 查询用户积分余额</h4>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">GET /wp-json/tanzanite/v1/loyalty/points
Headers: X-WP-Nonce: {nonce}</pre>';
            echo '          <p><strong>返回示例：</strong></p>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "user_id": 123,
  "points": 5000,
  "can_redeem": true,
  "max_redeemable_value": 50
}</pre>';
            echo '          <h4>3. 积分兑换礼品卡</h4>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">POST /wp-json/tanzanite/v1/giftcards/redeem
Headers: 
  Content-Type: application/json
  X-WP-Nonce: {nonce}
Body:
{
  "points": 1000,
  "giftcard_value": 10
}</pre>';
            echo '          <p><strong>成功返回：</strong></p>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "success": true,
  "giftcard_id": 456,
  "card_code": "REDEEM-ABC123XYZ",
  "balance": 10.00,
  "points_spent": 1000,
  "points_remaining": 4000,
  "expires_at": "2026-11-11 08:00:00",
  "message": "兑换成功"
}</pre>';
            echo '          <p><strong>失败返回：</strong></p>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "code": "insufficient_points",
  "message": "积分不足",
  "data": { "status": 400 }
}</pre>';
            echo '          <h4>4. 查询用户的礼品卡列表</h4>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">GET /wp-json/tanzanite/v1/giftcards/my
Headers: X-WP-Nonce: {nonce}</pre>';
            echo '          <p><strong>返回示例：</strong></p>';
            echo '          <pre style="background:#fff;padding:10px;border:1px solid #ddd;overflow-x:auto;">{
  "items": [
    {
      "id": 456,
      "card_code": "REDEEM-ABC123XYZ",
      "balance": 10.00,
      "original_value": 10.00,
      "points_spent": 1000,
      "status": "active",
      "expires_at": "2026-11-11 08:00:00",
      "created_at": "2025-11-11 08:00:00"
    }
  ]
}</pre>';
            echo '          <hr style="margin:20px 0;" />';
            echo '          <p style="color:#d63638;"><strong>⚠️ 重要提示：</strong></p>';
            echo '          <ul style="margin-left:20px;">';
            echo '              <li>所有需要身份验证的接口都需要在请求头中包含 <code>X-WP-Nonce</code></li>';
            echo '              <li>Nonce 可通过 <code>/wp-json/</code> 端点获取，或在登录后从 cookie 中读取</li>';
            echo '              <li>兑换接口会自动扣除用户积分并创建交易记录</li>';
            echo '              <li>礼品卡卡号自动生成，格式为 <code>REDEEM-{12位随机字符}</code></li>';
            echo '              <li>建议在前端实现兑换确认弹窗，避免误操作</li>';
            echo '          </ul>';
            echo '      </div>';
            echo '  </div>';

            echo '</div>';
            
            // 添加礼品卡封面设计的 JavaScript
            ?>
            <script type="text/javascript">
            document.addEventListener('DOMContentLoaded', function() {
                console.log('Gift card cover script loading...');
                
                // 封面类型切换
                const coverTypeRadios = document.querySelectorAll('input[name="giftcard_cover_type"]');
                const customCoverDiv = document.getElementById('tz-giftcard-custom-cover');
                const templateCoverDiv = document.getElementById('tz-giftcard-template-cover');
                const previewArea = document.getElementById('tz-cover-preview-area');
                
                console.log('Cover type radios found:', coverTypeRadios.length);
                console.log('Custom cover div found:', !!customCoverDiv);
                console.log('Template cover div found:', !!templateCoverDiv);
                console.log('Preview area found:', !!previewArea);
                
                function updateCoverDisplay() {
                    const selectedRadio = document.querySelector('input[name="giftcard_cover_type"]:checked');
                    const selectedType = selectedRadio ? selectedRadio.value : 'default';
                    console.log('Selected cover type:', selectedType);
                    
                    if (customCoverDiv) {
                        customCoverDiv.style.display = selectedType === 'custom' ? 'block' : 'none';
                        console.log('Custom cover div display:', customCoverDiv.style.display);
                    }
                    
                    if (templateCoverDiv) {
                        templateCoverDiv.style.display = selectedType === 'template' ? 'block' : 'none';
                        console.log('Template cover div display:', templateCoverDiv.style.display);
                    }
                    
                    updatePreview();
                }
                
                function updatePreview() {
                    if (!previewArea) return;
                    
                    const selectedRadio = document.querySelector('input[name="giftcard_cover_type"]:checked');
                    const selectedType = selectedRadio ? selectedRadio.value : 'default';
                    const customUrlInput = document.getElementById('giftcard_cover_url');
                    const customUrl = customUrlInput ? customUrlInput.value : '';
                    const templateSelect = document.getElementById('giftcard_cover_template');
                    const selectedTemplate = templateSelect ? templateSelect.value : 'elegant';
                    
                    console.log('updatePreview called:');
                    console.log('- selectedType:', selectedType);
                    console.log('- customUrl:', customUrl);
                    console.log('- selectedTemplate:', selectedTemplate);
                    console.log('- templateSelect element:', templateSelect);
                    
                    if (selectedType === 'custom' && customUrl) {
                        previewArea.style.backgroundImage = 'url(' + customUrl + ')';
                        previewArea.style.backgroundSize = 'cover';
                        previewArea.style.backgroundPosition = 'center';
                        previewArea.textContent = '';
                    } else if (selectedType === 'template') {
                        const templates = {
                            elegant: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                            festive: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
                            modern: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
                            luxury: 'linear-gradient(135deg, #434343 0%, #000000 100%)',
                            spring: 'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)'
                        };
                        const selectedGradient = templates[selectedTemplate] || templates.elegant;
                        console.log('Applying template:', selectedTemplate, 'with gradient:', selectedGradient);
                        
                        // 清除之前的样式
                        previewArea.style.removeProperty('background');
                        previewArea.style.removeProperty('background-color');
                        previewArea.style.removeProperty('background-image');
                        
                        // 使用 setProperty 强制设置样式
                        previewArea.style.setProperty('background-image', selectedGradient, 'important');
                        previewArea.style.setProperty('background-color', 'transparent', 'important');
                        
                        // 设置文本样式
                        previewArea.style.color = '#fff';
                        previewArea.style.fontWeight = 'bold';
                        previewArea.textContent = '礼品卡';
                        
                        console.log('Preview area backgroundImage set to:', previewArea.style.backgroundImage);
                        console.log('Final computed style:', window.getComputedStyle(previewArea).backgroundImage);
                    } else {
                        previewArea.style.background = '#f0f0f0';
                        previewArea.style.backgroundImage = 'none';
                        previewArea.style.color = '#666';
                        previewArea.style.fontWeight = 'normal';
                        previewArea.textContent = '默认封面';
                    }
                }
                
                // 绑定事件
                coverTypeRadios.forEach(function(radio) {
                    radio.addEventListener('change', updateCoverDisplay);
                });
                
                const urlInput = document.getElementById('giftcard_cover_url');
                if (urlInput) {
                    urlInput.addEventListener('input', updatePreview);
                }
                
                const templateSelect = document.getElementById('giftcard_cover_template');
                if (templateSelect) {
                    templateSelect.addEventListener('change', updatePreview);
                }
                
                // 上传图片功能
                const uploadBtn = document.getElementById('tz-upload-cover');
                if (uploadBtn) {
                    uploadBtn.addEventListener('click', function() {
                        if (typeof wp !== 'undefined' && wp.media) {
                            const mediaUploader = wp.media({
                                title: '选择礼品卡封面图片',
                                button: { text: '使用此图片' },
                                multiple: false,
                                library: { type: 'image' }
                            });
                            
                            mediaUploader.on('select', function() {
                                const attachment = mediaUploader.state().get('selection').first().toJSON();
                                const urlInput = document.getElementById('giftcard_cover_url');
                                if (urlInput) {
                                    urlInput.value = attachment.url;
                                    updatePreview();
                                }
                            });
                            
                            mediaUploader.open();
                        } else {
                            alert('媒体上传功能不可用，请手动输入图片 URL');
                        }
                    });
                }
                
                // 初始化显示
                updateCoverDisplay();
                
                console.log('Gift card cover script loaded successfully');
            });
            </script>
            <?php
        }

        /**
         * 渲染积分/等级设置页面
         */
        public function render_loyalty_settings(): void {
            if ( ! current_user_can( 'manage_options' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            $settings = $this->get_loyalty_settings();

            // 加载 WordPress 组件样式
            wp_enqueue_style( 'wp-components' );
            
            // 加载 Loyalty Settings JS
            wp_enqueue_script(
                'tz-loyalty-settings',
                plugins_url( 'assets/js/loyalty-settings.js', TANZANITE_LEGACY_MAIN_FILE ),
                array( 'tz-admin-common', 'wp-element', 'wp-i18n', 'wp-components', 'wp-url', 'wp-api-fetch' ),
                self::VERSION . '.sync.' . time(),
                true
            );

            // 传递配置到 JS
            wp_localize_script(
                'tz-loyalty-settings',
                'TzLoyaltyConfig',
                array(
                    'nonce'    => wp_create_nonce( 'wp_rest' ),
                    'settings' => $settings,
                )
            );

            echo '<div class="tz-settings-wrapper tz-loyalty-wrapper">';
            echo '  <div class="tz-settings-header">';
            echo '      <h1>' . esc_html__( 'Loyalty & Points Settings', 'tanzanite-settings' ) . '</h1>';
            echo '      <p>' . esc_html__( '配置会员等级、积分规则、推荐奖励等忠诚度系统设置。', 'tanzanite-settings' ) . '</p>';
            echo '  </div>';

            echo '  <div id="tz-loyalty-notice" class="notice" style="display:none;margin-bottom:16px;"></div>';

            echo '  <form id="tz-loyalty-form" method="post" action="options.php">';
            settings_fields( 'tanzanite_loyalty' );
            echo '      <input type="hidden" id="tz-loyalty-config" name="tanzanite_loyalty_config" value="' . esc_attr( wp_json_encode( $settings, JSON_UNESCAPED_UNICODE ) ) . '" />';
            echo '      <div id="tz-loyalty-app"></div>';
            echo '      <noscript><p class="notice notice-error">' . esc_html__( '需要启用 JavaScript 才能管理积分等级。', 'tanzanite-settings' ) . '</p></noscript>';
            echo '      <div style="display:flex;gap:12px;align-items:center;margin-top:16px;">';
            submit_button( __( '保存设置', 'tanzanite-settings' ), 'primary', 'submit', false, [ 'id' => 'tz-loyalty-submit' ] );
            echo '          <button type="button" class="button" id="tz-loyalty-reset" onclick="resetLoyaltyToEnglish()">' . esc_html__( '重置会员等级为英文', 'tanzanite-settings' ) . '</button>';
            echo '      </div>';
            echo '  </form>';
            
            // 添加重置脚本
            echo '<script type="text/javascript">';
            echo 'function resetLoyaltyToEnglish() {';
            echo '    if (!confirm("确定要重置会员等级名称为英文吗？这将覆盖当前的自定义设置。")) return;';
            echo '    ';
            echo '    fetch("' . esc_url( admin_url( 'admin-ajax.php' ) ) . '", {';
            echo '        method: "POST",';
            echo '        headers: { "Content-Type": "application/x-www-form-urlencoded" },';
            echo '        body: "action=reset_loyalty_to_english&nonce=' . wp_create_nonce( 'reset_loyalty_english' ) . '"';
            echo '    })';
            echo '    .then(response => response.json())';
            echo '    .then(data => {';
            echo '        if (data.success) {';
            echo '            alert("会员等级已重置为英文，请刷新页面查看效果。");';
            echo '            location.reload();';
            echo '        } else {';
            echo '            alert("重置失败：" + (data.data || "未知错误"));';
            echo '        }';
            echo '    })';
            echo '    .catch(error => {';
            echo '        console.error("Reset error:", error);';
            echo '        alert("重置失败，请稍后重试。");';
            echo '    });';
            echo '}';
            echo '</script>';

            echo '</div>';
        }

        /**
         * 获取积分/等级设置
         */
        private function get_loyalty_settings(): array {
            $default = $this->get_default_loyalty_config();
            $saved_json = get_option( 'tanzanite_loyalty_config', '' );

            // 数据库中存储的是 JSON 字符串，需要解码
            if ( empty( $saved_json ) ) {
                return $default;
            }

            $saved = json_decode( $saved_json, true );
            if ( ! is_array( $saved ) || empty( $saved ) ) {
                return $default;
            }

            return array_replace_recursive( $default, $saved );
        }

        /**
         * 获取默认积分/等级配置
         */
        private function get_default_loyalty_config(): array {
            return array(
                'enabled'              => true,
                'apply_cart_discount'  => true,
                'points_per_unit'      => 1,
                'daily_checkin_points' => 0,
                'referral'             => array(
                    'enabled'         => true,
                    'bonus_inviter'   => 50,
                    'bonus_invitee'   => 30,
                    'token_ttl_days'  => 7,
                    'token_max_uses'  => 50,
                ),
                'tiers'                => array(
                    'ordinary' => array(
                        'label'      => __( 'Ordinary', 'tanzanite-settings' ),
                        'name'       => __( 'Ordinary', 'tanzanite-settings' ),
                        'min'        => 0,
                        'max'        => 499,
                        'discount'   => 0,
                        'products'   => array(),
                        'categories' => array(),
                        'redeem'     => array(
                            'enabled'              => false,
                            'percent_of_total'     => 5,
                            'value_per_point_base' => 0.01,
                            'min_points'           => 0,
                            'stack_with_percent'   => true,
                        ),
                    ),
                    'bronze'   => array(
                        'label'      => __( 'Bronze', 'tanzanite-settings' ),
                        'name'       => __( 'Bronze', 'tanzanite-settings' ),
                        'min'        => 500,
                        'max'        => 1999,
                        'discount'   => 5,
                        'products'   => array(),
                        'categories' => array(),
                        'redeem'     => array(
                            'enabled'              => false,
                            'percent_of_total'     => 5,
                            'value_per_point_base' => 0.01,
                            'min_points'           => 0,
                            'stack_with_percent'   => true,
                        ),
                    ),
                    'silver'   => array(
                        'label'      => __( 'Silver', 'tanzanite-settings' ),
                        'name'       => __( 'Silver', 'tanzanite-settings' ),
                        'min'        => 2000,
                        'max'        => 4999,
                        'discount'   => 10,
                        'products'   => array(),
                        'categories' => array(),
                        'redeem'     => array(
                            'enabled'              => false,
                            'percent_of_total'     => 5,
                            'value_per_point_base' => 0.01,
                            'min_points'           => 0,
                            'stack_with_percent'   => true,
                        ),
                    ),
                    'gold'     => array(
                        'label'      => __( 'Gold', 'tanzanite-settings' ),
                        'name'       => __( 'Gold', 'tanzanite-settings' ),
                        'min'        => 5000,
                        'max'        => 9999,
                        'discount'   => 15,
                        'products'   => array(),
                        'categories' => array(),
                        'redeem'     => array(
                            'enabled'              => false,
                            'percent_of_total'     => 5,
                            'value_per_point_base' => 0.01,
                            'min_points'           => 0,
                            'stack_with_percent'   => true,
                        ),
                    ),
                    'platinum' => array(
                        'label'      => __( 'Platinum', 'tanzanite-settings' ),
                        'name'       => __( 'Platinum', 'tanzanite-settings' ),
                        'min'        => 10000,
                        'max'        => null,
                        'discount'   => 20,
                        'products'   => array(),
                        'categories' => array(),
                        'redeem'     => array(
                            'enabled'              => false,
                            'percent_of_total'     => 5,
                            'value_per_point_base' => 0.01,
                            'min_points'           => 0,
                            'stack_with_percent'   => true,
                        ),
                    ),
                ),
            );
        }

        /**
         * 注册积分/等级设置选项
         */
        public function register_loyalty_settings(): void {
            register_setting(
                'tanzanite_loyalty',
                'tanzanite_loyalty_config',
                array(
                    'type'              => 'string',
                    'sanitize_callback' => array( $this, 'sanitize_loyalty_config' ),
                    'default'           => wp_json_encode( $this->get_default_loyalty_config() ),
                )
            );
        }

        /**
         * REST API: 获取会员等级配置（公开访问）
         */
        public function rest_get_loyalty_settings( $request ) {
            $config_json = get_option( 'tanzanite_loyalty_config', '' );
            
            if ( empty( $config_json ) ) {
                $config = $this->get_default_loyalty_config();
            } else {
                $config = json_decode( $config_json, true );
                if ( ! is_array( $config ) ) {
                    $config = $this->get_default_loyalty_config();
                }
            }
            
            // 返回简化的配置，只包含前端需要的字段
            $response = array(
                'tiers' => array(),
            );
            
            if ( isset( $config['tiers'] ) && is_array( $config['tiers'] ) ) {
                foreach ( $config['tiers'] as $key => $tier ) {
                    $response['tiers'][$key] = array(
                        'name'            => $tier['name'] ?? $tier['label'] ?? ucfirst( $key ),
                        'min'             => intval( $tier['min'] ?? 0 ),
                        'max'             => $tier['max'] === null ? null : intval( $tier['max'] ),
                        'discount'        => intval( $tier['discount'] ?? 0 ),
                        'points_discount' => intval( $tier['redeem']['percent_of_total'] ?? 0 ),
                        'stackable'       => boolval( $tier['redeem']['stack_with_percent'] ?? true ),
                    );
                }
            }
            
            return rest_ensure_response( $response );
        }

        /**
         * 清理积分/等级配置数据
         */
        public function sanitize_loyalty_config( $value ) {
            if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                error_log( 'Sanitizing loyalty config. Input: ' . substr( $value, 0, 200 ) );
            }
            
            if ( empty( $value ) ) {
                if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                    error_log( 'Empty value, returning default config' );
                }
                return wp_json_encode( $this->get_default_loyalty_config() );
            }

            $decoded = json_decode( $value, true );
            if ( ! is_array( $decoded ) ) {
                if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                    error_log( 'Invalid JSON, returning default config' );
                }
                return wp_json_encode( $this->get_default_loyalty_config() );
            }
            
            if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                error_log( 'Decoded config keys: ' . implode( ', ', array_keys( $decoded ) ) );
            }

            // 基本验证
            $sanitized = array(
                'enabled'              => ! empty( $decoded['enabled'] ),
                'apply_cart_discount'  => ! empty( $decoded['apply_cart_discount'] ),
                'points_per_unit'      => max( 0, intval( $decoded['points_per_unit'] ?? 1 ) ),
                'daily_checkin_points' => max( 0, intval( $decoded['daily_checkin_points'] ?? 0 ) ),
            );

            // 推荐设置
            if ( isset( $decoded['referral'] ) && is_array( $decoded['referral'] ) ) {
                $sanitized['referral'] = array(
                    'enabled'         => ! empty( $decoded['referral']['enabled'] ),
                    'bonus_inviter'   => max( 0, intval( $decoded['referral']['bonus_inviter'] ?? 50 ) ),
                    'bonus_invitee'   => max( 0, intval( $decoded['referral']['bonus_invitee'] ?? 30 ) ),
                    'token_ttl_days'  => max( 1, intval( $decoded['referral']['token_ttl_days'] ?? 7 ) ),
                    'token_max_uses'  => max( 1, intval( $decoded['referral']['token_max_uses'] ?? 50 ) ),
                );
            }

            // 等级设置
            if ( isset( $decoded['tiers'] ) && is_array( $decoded['tiers'] ) ) {
                $sanitized['tiers'] = array();
                foreach ( $decoded['tiers'] as $key => $tier ) {
                    if ( ! is_array( $tier ) ) {
                        continue;
                    }
                    $sanitized['tiers'][ sanitize_key( $key ) ] = array(
                        'label'      => sanitize_text_field( $tier['label'] ?? '' ),
                        'name'       => sanitize_text_field( $tier['name'] ?? '' ),
                        'min'        => max( 0, intval( $tier['min'] ?? 0 ) ),
                        'max'        => isset( $tier['max'] ) && $tier['max'] !== null ? max( 0, intval( $tier['max'] ) ) : null,
                        'discount'   => max( 0, min( 100, floatval( $tier['discount'] ?? 0 ) ) ),
                        'products'   => is_array( $tier['products'] ?? null ) ? array_map( 'intval', $tier['products'] ) : array(),
                        'categories' => is_array( $tier['categories'] ?? null ) ? array_map( 'intval', $tier['categories'] ) : array(),
                        'redeem'     => array(
                            'enabled'              => ! empty( $tier['redeem']['enabled'] ),
                            'percent_of_total'     => max( 0, min( 100, floatval( $tier['redeem']['percent_of_total'] ?? 5 ) ) ),
                            'value_per_point_base' => max( 0, floatval( $tier['redeem']['value_per_point_base'] ?? 0.01 ) ),
                            'min_points'           => max( 0, intval( $tier['redeem']['min_points'] ?? 0 ) ),
                            'stack_with_percent'   => ! empty( $tier['redeem']['stack_with_percent'] ),
                        ),
                    );
                }
            }

            $result = wp_json_encode( $sanitized, JSON_UNESCAPED_UNICODE );
            
            if ( defined( 'WP_DEBUG' ) && WP_DEBUG ) {
                error_log( 'Sanitized loyalty config result: ' . substr( $result, 0, 200 ) );
                error_log( 'daily_checkin_points value: ' . $sanitized['daily_checkin_points'] );
            }
            
            return $result;
        }

        // Members REST API 回调函数 - 已迁移到 Tanzanite_REST_Members_Controller
        // rest_list_members -> Tanzanite_REST_Members_Controller::get_items()
        // rest_get_member -> Tanzanite_REST_Members_Controller::get_item()
        // rest_update_member -> Tanzanite_REST_Members_Controller::update_item()
        // rest_delete_member -> Tanzanite_REST_Members_Controller::delete_item()

        // Coupons REST API 回调函数 - 已迁移到 Tanzanite_REST_Coupons_Controller
        // rest_list_coupons -> Tanzanite_REST_Coupons_Controller::get_items()
        // rest_get_coupon -> Tanzanite_REST_Coupons_Controller::get_item()
        // rest_create_coupon -> Tanzanite_REST_Coupons_Controller::create_item()
        // rest_update_coupon -> Tanzanite_REST_Coupons_Controller::update_item()
        // rest_delete_coupon -> Tanzanite_REST_Coupons_Controller::delete_item()

        // Gift Cards REST API 回调函数 - 已迁移到 Tanzanite_REST_Giftcards_Controller
        // rest_list_giftcards -> Tanzanite_REST_Giftcards_Controller::get_items()
        // rest_get_giftcard -> Tanzanite_REST_Giftcards_Controller::get_item()
        // rest_create_giftcard -> Tanzanite_REST_Giftcards_Controller::create_item()
        // rest_update_giftcard -> Tanzanite_REST_Giftcards_Controller::update_item()
        // rest_delete_giftcard -> Tanzanite_REST_Giftcards_Controller::delete_item()

        /**
         * REST API: 获取交易记录列表
         */
        public function rest_list_transactions( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;
            $results = $wpdb->get_results( "SELECT * FROM {$this->rewards_transactions_table} ORDER BY id DESC LIMIT 100", ARRAY_A );
            return new \WP_REST_Response( [ 'items' => $results ?: [] ], 200 );
        }

        /**
         * 处理保存积分兑换设置
         */
        public function handle_save_redeem_settings(): void {
            // 验证权限
            if ( ! current_user_can( 'manage_options' ) ) {
                wp_die( __( '无权限执行此操作。', 'tanzanite-settings' ) );
            }

            // 验证 nonce
            if ( ! isset( $_POST['tz_redeem_nonce'] ) || ! wp_verify_nonce( $_POST['tz_redeem_nonce'], 'tz_save_redeem_settings' ) ) {
                wp_die( __( '安全验证失败。', 'tanzanite-settings' ) );
            }

            // 保存设置
            update_option( 'tz_redeem_enabled', isset( $_POST['redeem_enabled'] ) ? '1' : '0' );
            update_option( 'tz_redeem_card_expiry_days', absint( $_POST['redeem_card_expiry_days'] ?? 365 ) );
            
            // 保存礼品卡封面设置
            if ( isset( $_POST['giftcard_cover_type'] ) ) {
                update_option( 'tz_giftcard_cover_type', sanitize_text_field( $_POST['giftcard_cover_type'] ) );
            }
            if ( isset( $_POST['giftcard_cover_url'] ) ) {
                update_option( 'tz_giftcard_cover_url', esc_url_raw( $_POST['giftcard_cover_url'] ) );
            }
            if ( isset( $_POST['giftcard_cover_template'] ) ) {
                update_option( 'tz_giftcard_cover_template', sanitize_text_field( $_POST['giftcard_cover_template'] ) );
            }

            // 重定向回设置页面
            wp_redirect( add_query_arg(
                array(
                    'page' => 'tanzanite-settings-rewards',
                    'tab' => 'redeem-settings',
                    'saved' => '1'
                ),
                admin_url( 'admin.php' )
            ) );
            exit;
        }

        /**
         * 渲染 SEO 设置页面
         */
        public function render_seo_page(): void {
            error_log('Tanzanite Settings: render_seo_page called');
            
            if ( ! current_user_can( 'manage_options' ) ) {
                error_log('Tanzanite Settings: User does not have manage_options capability');
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            // 调用 MyTheme SEO 的渲染方法
            if ( class_exists( 'MyTheme_SEO_Plugin' ) ) {
                error_log('Tanzanite Settings: MyTheme_SEO_Plugin class exists, rendering...');
                $seo_instance = MyTheme_SEO_Plugin::instance();
                $seo_instance->render_admin_page();
            } else {
                error_log('Tanzanite Settings: MyTheme_SEO_Plugin class does NOT exist!');
                echo '<div class="wrap">';
                echo '<h1>' . esc_html__( 'SEO Settings', 'tanzanite-settings' ) . '</h1>';
                echo '<p>' . esc_html__( 'MyTheme SEO 模块未正确加载。', 'tanzanite-settings' ) . '</p>';
                echo '<p>Debug: class_exists(MyTheme_SEO_Plugin) = ' . (class_exists('MyTheme_SEO_Plugin') ? 'true' : 'false') . '</p>';
                echo '</div>';
            }
        }

        /**
         * 渲染 Markdown Templates 管理页面
         */
        public function render_markdown_templates_page(): void {
            if ( ! current_user_can( 'manage_options' ) ) {
                wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
            }

            // 处理表单提交
            if ( isset( $_POST['tz_markdown_templates_nonce'] ) && wp_verify_nonce( $_POST['tz_markdown_templates_nonce'], 'tz_save_markdown_templates' ) ) {
                $templates = [
                    'basic'      => wp_unslash( $_POST['tz_template_basic'] ?? '' ),
                    'spec'       => wp_unslash( $_POST['tz_template_spec'] ?? '' ),
                    'after_sale' => wp_unslash( $_POST['tz_template_after_sale'] ?? '' ),
                ];
                
                update_option( 'tanzanite_markdown_templates', $templates );
                
                echo '<div class="notice notice-success is-dismissible"><p>' . esc_html__( '模板已保存。', 'tanzanite-settings' ) . '</p></div>';
            }

            // 获取当前模板
            $saved_templates = get_option( 'tanzanite_markdown_templates', [] );
            $default_templates = $this->get_default_markdown_templates();
            $templates = wp_parse_args( $saved_templates, $default_templates );

            ?>
            <div class="wrap tz-settings-wrapper">
                <h1><?php esc_html_e( 'Markdown Templates', 'tanzanite-settings' ); ?></h1>
                <p class="description"><?php esc_html_e( '配置商品详情页的 Markdown 模板，这些模板可以在添加/编辑商品时快速插入。', 'tanzanite-settings' ); ?></p>

                <form method="post" action="">
                    <?php wp_nonce_field( 'tz_save_markdown_templates', 'tz_markdown_templates_nonce' ); ?>

                    <div class="tz-settings-section" style="margin-top:20px;">
                        <div class="tz-section-title"><?php esc_html_e( '商品亮点模板', 'tanzanite-settings' ); ?></div>
                        <div class="tz-section-body">
                            <p class="description"><?php esc_html_e( '用于快速插入商品的核心卖点和详情描述。支持图片、标题和富文本格式。', 'tanzanite-settings' ); ?></p>
                            <?php
                            wp_editor(
                                $templates['basic'],
                                'tz_template_basic_editor',
                                [
                                    'textarea_name'  => 'tz_template_basic',
                                    'media_buttons'  => true,
                                    'textarea_rows'  => 12,
                                    'editor_height'  => 300,
                                    'drag_drop_upload' => true,
                                    'tinymce'        => [
                                        'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
                                        'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
                                    ],
                                    'quicktags'      => true,
                                ]
                            );
                            ?>
                        </div>
                    </div>

                    <div class="tz-settings-section" style="margin-top:20px;">
                        <div class="tz-section-title"><?php esc_html_e( '规格参数模板', 'tanzanite-settings' ); ?></div>
                        <div class="tz-section-body">
                            <p class="description"><?php esc_html_e( '用于快速插入商品的规格参数表格。支持表格、图片和富文本格式。', 'tanzanite-settings' ); ?></p>
                            <?php
                            wp_editor(
                                $templates['spec'],
                                'tz_template_spec_editor',
                                [
                                    'textarea_name'  => 'tz_template_spec',
                                    'media_buttons'  => true,
                                    'textarea_rows'  => 12,
                                    'editor_height'  => 300,
                                    'drag_drop_upload' => true,
                                    'tinymce'        => [
                                        'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
                                        'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
                                    ],
                                    'quicktags'      => true,
                                ]
                            );
                            ?>
                        </div>
                    </div>

                    <div class="tz-settings-section" style="margin-top:20px;">
                        <div class="tz-section-title"><?php esc_html_e( '售后说明模板', 'tanzanite-settings' ); ?></div>
                        <div class="tz-section-body">
                            <p class="description"><?php esc_html_e( '用于快速插入售后保障和服务说明。支持图片、标题和富文本格式。', 'tanzanite-settings' ); ?></p>
                            <?php
                            wp_editor(
                                $templates['after_sale'],
                                'tz_template_after_sale_editor',
                                [
                                    'textarea_name'  => 'tz_template_after_sale',
                                    'media_buttons'  => true,
                                    'textarea_rows'  => 12,
                                    'editor_height'  => 300,
                                    'drag_drop_upload' => true,
                                    'tinymce'        => [
                                        'toolbar1' => 'formatselect,bold,italic,underline,strikethrough,alignleft,aligncenter,alignright,bullist,numlist,blockquote,link,unlink,image,undo,redo',
                                        'toolbar2' => 'styleselect,outdent,indent,table,code,removeformat',
                                    ],
                                    'quicktags'      => true,
                                ]
                            );
                            ?>
                        </div>
                    </div>

                    <p class="submit">
                        <button type="submit" class="button button-primary"><?php esc_html_e( '保存模板', 'tanzanite-settings' ); ?></button>
                        <button type="button" class="button" id="tz-restore-defaults"><?php esc_html_e( '恢复默认', 'tanzanite-settings' ); ?></button>
                    </p>
                </form>
            </div>

            <script type="text/javascript">
            document.addEventListener('DOMContentLoaded', function() {
                const restoreBtn = document.getElementById('tz-restore-defaults');
                if (restoreBtn) {
                    restoreBtn.addEventListener('click', function() {
                        if (confirm(<?php echo wp_json_encode( __( '确定要恢复默认模板吗？这将覆盖当前的自定义内容。', 'tanzanite-settings' ) ); ?>)) {
                            const defaultTemplates = <?php echo wp_json_encode( $default_templates ); ?>;
                            
                            // 更新富文本编辑器内容
                            if (typeof tinymce !== 'undefined') {
                                const basicEditor = tinymce.get('tz_template_basic_editor');
                                const specEditor = tinymce.get('tz_template_spec_editor');
                                const afterSaleEditor = tinymce.get('tz_template_after_sale_editor');
                                
                                if (basicEditor) basicEditor.setContent(defaultTemplates.basic);
                                if (specEditor) specEditor.setContent(defaultTemplates.spec);
                                if (afterSaleEditor) afterSaleEditor.setContent(defaultTemplates.after_sale);
                            }
                            
                            // 同时更新 textarea（以防编辑器未加载）
                            const basicField = document.querySelector('[name="tz_template_basic"]');
                            const specField = document.querySelector('[name="tz_template_spec"]');
                            const afterSaleField = document.querySelector('[name="tz_template_after_sale"]');
                            
                            if (basicField) basicField.value = defaultTemplates.basic;
                            if (specField) specField.value = defaultTemplates.spec;
                            if (afterSaleField) afterSaleField.value = defaultTemplates.after_sale;
                        }
                    });
                }
            });
            </script>
            <?php
        }

        /**
         * 获取默认 Markdown 模板
         */
        private function get_default_markdown_templates(): array {
            return [
                'basic'      => "# 商品亮点\n- 优选材质，兼顾舒适与耐用\n- 设计贴合日常场景，易于搭配\n- 支持多种配送方式与售后保障\n\n## 详情描述\n请在此补充产品的核心卖点、使用场景与图文信息。",
                'spec'       => "## 规格参数\n| 项目 | 参数 |\n| --- | --- |\n| 材质 | 请输入 |\n| 尺寸 | 请输入 |\n| 重量 | 请输入 |\n| 颜色 | 请输入 |\n\n> 可根据实际情况补充更多行，或删除不适用的字段。",
                'after_sale' => "## 售后与保障\n1. 支持七天无理由退换货，保持商品及包装完好。\n2. 如遇质量问题，请联系客服并提供照片，我们将在 24 小时内响应。\n3. 定制/特殊商品的退换政策，请以页面说明为准。\n\n感谢您的信任与支持！",
            ];
        }

        /**
         * 处理重置会员等级为英文的 AJAX 请求
         */
        public function handle_reset_loyalty_to_english(): void {
            // 验证 nonce
            if ( ! wp_verify_nonce( $_POST['nonce'] ?? '', 'reset_loyalty_english' ) ) {
                wp_send_json_error( '安全验证失败' );
                return;
            }

            // 检查权限
            if ( ! current_user_can( 'manage_options' ) ) {
                wp_send_json_error( '权限不足' );
                return;
            }

            try {
                // 删除现有的设置，这样会使用默认的英文配置
                delete_option( 'tanzanite_loyalty_config' );
                
                wp_send_json_success( '会员等级已重置为英文' );
            } catch ( Exception $e ) {
                wp_send_json_error( '重置失败：' . $e->getMessage() );
            }
        }
    }
}

if ( ! function_exists( 'tanzanite_settings_sanitize_tier_pricing' ) ) {
    function tanzanite_settings_sanitize_tier_pricing( $value, bool $from_request = false ) {
        if ( empty( $value ) || ! is_array( $value ) ) {
            return [];
        }

        $sanitized = [];

        foreach ( $value as $item ) {
            if ( ! is_array( $item ) ) {
                if ( is_object( $item ) ) {
                    $item = json_decode( wp_json_encode( $item ), true );
                } else {
                    continue;
                }
            }

            $min_qty = isset( $item['min_qty'] ) ? (int) $item['min_qty'] : ( isset( $item['minQty'] ) ? (int) $item['minQty'] : ( isset( $item['min'] ) ? (int) $item['min'] : 0 ) );
            $max_raw = $item['max_qty'] ?? ( $item['maxQty'] ?? ( $item['max'] ?? null ) );
            $max_qty = ( '' === $max_raw || null === $max_raw ) ? null : (int) $max_raw;
            $price_raw = $item['price'] ?? ( $item['amount'] ?? ( $item['value'] ?? null ) );
            $price = is_numeric( $price_raw ) ? (float) $price_raw : null;
            $note_raw = $item['note'] ?? ( $item['label'] ?? ( $item['desc'] ?? '' ) );
            $note = $note_raw ? sanitize_text_field( (string) $note_raw ) : '';

            if ( $min_qty <= 0 || null === $price || $price < 0 ) {
                if ( $from_request ) {
                    return new \WP_Error( 'invalid_tier_qty', __( '请填写有效的最小数量和单价。', 'tanzanite-settings' ) );
                }
                continue;
            }

            if ( null !== $max_qty && $max_qty < $min_qty ) {
                if ( $from_request ) {
                    return new \WP_Error( 'invalid_tier_range', __( '最大数量必须大于或等于最小数量。', 'tanzanite-settings' ) );
                }
                continue;
            }

            $sanitized[] = [
                'min_qty' => $min_qty,
                'max_qty' => $max_qty,
                'price'   => (float) number_format( $price, 2, '.', '' ),
                'note'    => $note,
            ];
        }

        if ( empty( $sanitized ) ) {
            return [];
        }

        usort(
            $sanitized,
            static function ( $a, $b ) {
                return $a['min_qty'] <=> $b['min_qty'];
            }
        );

        $previous_max = null;
        $previous_min = null;

        foreach ( $sanitized as $index => $row ) {
            $min = (int) $row['min_qty'];
            $max = $row['max_qty'];

            if ( 0 === $index ) {
                $previous_max = $max;
                $previous_min = $min;
                continue;
            }

            if ( null === $previous_max ) {
                $error = __( '只有最后一个阶梯可以不设置最大数量。', 'tanzanite-settings' );

                return $from_request ? new \WP_Error( 'invalid_tier_limit', $error ) : [];
            }

            if ( $min <= $previous_max || $min <= $previous_min ) {
                $error = __( '阶梯区间存在重叠或顺序错误，请检查后再保存。', 'tanzanite-settings' );

                return $from_request ? new \WP_Error( 'invalid_tier_overlap', $error ) : [];
            }

            $previous_max = $max;
            $previous_min = $min;
        }

        return $sanitized;
    }
}

add_action(
    'plugins_loaded',
    static function () {
        Tanzanite_Settings_Plugin::instance();
    }
);

register_activation_hook(
    TANZANITE_LEGACY_MAIN_FILE,
    static function (): void {
        Tanzanite_Settings_Plugin::instance()->maybe_upgrade_schema();
    }
);
