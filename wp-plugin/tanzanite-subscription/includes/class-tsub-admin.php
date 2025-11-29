<?php

namespace TanzaniteSubscription;

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class TSUB_Admin {
    /** @var TSUB_Admin|null */
    protected static $instance = null;

    /**
     * Singleton entry point.
     *
     * @return TSUB_Admin
     */
    public static function instance() {
        if ( null === self::$instance ) {
            self::$instance = new self();
        }

        return self::$instance;
    }

    /**
     * TSUB_Admin constructor.
     */
    protected function __construct() {
        add_action( 'admin_menu', array( $this, 'register_menu' ) );
        add_action( 'admin_init', array( $this, 'register_settings' ) );
    }

    /**
     * Register the Subscription settings page under the Settings menu.
     */
    public function register_menu() {
        add_options_page(
            __( 'Tanzanite Subscription', 'tanzanite-subscription' ),
            __( 'Tanzanite Subscription', 'tanzanite-subscription' ),
            'manage_options',
            'tanz-subscription-settings',
            array( $this, 'render_settings_page' )
        );
    }

    /**
     * Register options for From Email/Name and auto-notify toggles.
     */
    public function register_settings() {
        register_setting(
            'tanz_sub_options',
            'tanz_sub_from_email',
            array(
                'type'              => 'string',
                'sanitize_callback' => 'sanitize_email',
                'default'           => '',
            )
        );

        register_setting(
            'tanz_sub_options',
            'tanz_sub_from_name',
            array(
                'type'              => 'string',
                'sanitize_callback' => 'sanitize_text_field',
                'default'           => '',
            )
        );

        register_setting(
            'tanz_sub_options',
            'tanz_sub_auto_notify_posts',
            array(
                'type'              => 'integer',
                'sanitize_callback' => 'absint',
                'default'           => 0,
            )
        );

        register_setting(
            'tanz_sub_options',
            'tanz_sub_auto_notify_products',
            array(
                'type'              => 'integer',
                'sanitize_callback' => 'absint',
                'default'           => 0,
            )
        );

        register_setting(
            'tanz_sub_options',
            'tanz_sub_footer_note',
            array(
                'type'              => 'string',
                'sanitize_callback' => 'sanitize_textarea_field',
                'default'           => '',
            )
        );
    }

    /**
     * Render the settings page markup.
     */
    public function render_settings_page() {
        if ( ! current_user_can( 'manage_options' ) ) {
            return;
        }

        $from_email = get_option( 'tanz_sub_from_email', '' );
        $from_name  = get_option( 'tanz_sub_from_name', '' );

        $auto_posts    = (int) get_option( 'tanz_sub_auto_notify_posts', 0 );
        $auto_products = (int) get_option( 'tanz_sub_auto_notify_products', 0 );
        $footer_note   = get_option( 'tanz_sub_footer_note', '' );
        ?>
        <div class="wrap">
            <h1><?php esc_html_e( 'Tanzanite 订阅设置', 'tanzanite-subscription' ); ?></h1>

            <form method="post" action="options.php">
                <?php settings_fields( 'tanz_sub_options' ); ?>

                <table class="form-table" role="presentation">
                    <tr>
                        <th scope="row">
                            <label for="tanz_sub_from_email"><?php esc_html_e( '发件邮箱（From Email）', 'tanzanite-subscription' ); ?></label>
                        </th>
                        <td>
                            <input
                                name="tanz_sub_from_email"
                                id="tanz_sub_from_email"
                                type="email"
                                class="regular-text"
                                value="<?php echo esc_attr( $from_email ); ?>"
                            />
                            <p class="description">
                                <?php esc_html_e( '建议使用与你域名一致的邮箱地址，例如 no-reply@your-domain.com，并在 DNS 中配置好 SPF/DKIM。', 'tanzanite-subscription' ); ?>
                            </p>
                        </td>
                    </tr>

                    <tr>
                        <th scope="row">
                            <label for="tanz_sub_from_name"><?php esc_html_e( '发件人名称（From Name）', 'tanzanite-subscription' ); ?></label>
                        </th>
                        <td>
                            <input
                                name="tanz_sub_from_name"
                                id="tanz_sub_from_name"
                                type="text"
                                class="regular-text"
                                value="<?php echo esc_attr( $from_name ); ?>"
                            />
                        </td>
                    </tr>

                    <tr>
                        <th scope="row"><?php esc_html_e( '自动通知设置', 'tanzanite-subscription' ); ?></th>
                        <td>
                            <fieldset>
                                <label for="tanz_sub_auto_notify_posts">
                                    <input
                                        name="tanz_sub_auto_notify_posts"
                                        id="tanz_sub_auto_notify_posts"
                                        type="checkbox"
                                        value="1"
                                        <?php checked( 1, $auto_posts ); ?>
                                    />
                                    <?php esc_html_e( '当有新的博客文章发布时自动发送订阅邮件', 'tanzanite-subscription' ); ?>
                                </label>
                                <br />
                                <label for="tanz_sub_auto_notify_products">
                                    <input
                                        name="tanz_sub_auto_notify_products"
                                        id="tanz_sub_auto_notify_products"
                                        type="checkbox"
                                        value="1"
                                        <?php checked( 1, $auto_products ); ?>
                                    />
                                    <?php esc_html_e( '当有新的商品上架时自动发送订阅邮件', 'tanzanite-subscription' ); ?>
                                </label>
                                <p class="description">
                                    <?php esc_html_e( '你可以关闭以上自动通知，仅通过「订阅群发」页面手动发送，避免过于频繁打扰订阅者。', 'tanzanite-subscription' ); ?>
                                </p>
                            </fieldset>
                        </td>
                    </tr>

                    <tr>
                        <th scope="row">
                            <label for="tanz_sub_footer_note"><?php esc_html_e( '邮件尾部备注', 'tanzanite-subscription' ); ?></label>
                        </th>
                        <td>
                            <textarea
                                name="tanz_sub_footer_note"
                                id="tanz_sub_footer_note"
                                rows="3"
                                class="large-text"
                            ><?php echo esc_textarea( $footer_note ); ?></textarea>
                            <p class="description">
                                <?php esc_html_e( '可选，在所有订阅相关邮件底部自动追加的一段文字，例如品牌签名或联系方式。', 'tanzanite-subscription' ); ?>
                            </p>
                        </td>
                    </tr>
                </table>

                <?php submit_button(); ?>
            </form>
        </div>
        <?php
    }
}
