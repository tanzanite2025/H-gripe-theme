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
            <h1><?php esc_html_e( 'Tanzanite Subscription Settings', 'tanzanite-subscription' ); ?></h1>

            <form method="post" action="options.php">
                <?php settings_fields( 'tanz_sub_options' ); ?>

                <table class="form-table" role="presentation">
                    <tr>
                        <th scope="row">
                            <label for="tanz_sub_from_email"><?php esc_html_e( 'From Email', 'tanzanite-subscription' ); ?></label>
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
                                <?php esc_html_e( 'Suggested: a domain-matching address such as no-reply@your-domain.com, with SPF/DKIM configured.', 'tanzanite-subscription' ); ?>
                            </p>
                        </td>
                    </tr>

                    <tr>
                        <th scope="row">
                            <label for="tanz_sub_from_name"><?php esc_html_e( 'From Name', 'tanzanite-subscription' ); ?></label>
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
                        <th scope="row"><?php esc_html_e( 'Automatic notifications', 'tanzanite-subscription' ); ?></th>
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
                                    <?php esc_html_e( 'Send email automatically when a new blog post is published', 'tanzanite-subscription' ); ?>
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
                                    <?php esc_html_e( 'Send email automatically when a new product is published', 'tanzanite-subscription' ); ?>
                                </label>
                                <p class="description">
                                    <?php esc_html_e( 'You can disable these and use manual broadcasts only to avoid over-notifying subscribers.', 'tanzanite-subscription' ); ?>
                                </p>
                            </fieldset>
                        </td>
                    </tr>

                    <tr>
                        <th scope="row">
                            <label for="tanz_sub_footer_note"><?php esc_html_e( 'Email footer note', 'tanzanite-subscription' ); ?></label>
                        </th>
                        <td>
                            <textarea
                                name="tanz_sub_footer_note"
                                id="tanz_sub_footer_note"
                                rows="3"
                                class="large-text"
                            ><?php echo esc_textarea( $footer_note ); ?></textarea>
                            <p class="description">
                                <?php esc_html_e( 'Optional text appended to the bottom of all subscription emails, e.g. brand signature or contact info.', 'tanzanite-subscription' ); ?>
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
