<?php

namespace TanzaniteSubscription;

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

/**
 * Admin page for manual subscription broadcasts.
 *
 * Adds a submenu under the Tanzanite Settings menu that allows an admin to
 * choose a post or product and send a one-off email to all confirmed
 * subscribers.
 */
class TSUB_Broadcasts {
    /** @var TSUB_Broadcasts|null */
    protected static $instance = null;

    /**
     * Singleton entry point.
     *
     * @return TSUB_Broadcasts
     */
    public static function instance() {
        if ( null === self::$instance ) {
            self::$instance = new self();
        }

        return self::$instance;
    }

    /**
     * TSUB_Broadcasts constructor.
     */
    protected function __construct() {
        add_action( 'admin_menu', array( $this, 'register_menu' ) );
        add_action( 'admin_post_tsub_send_broadcast', array( $this, 'handle_send_broadcast' ) );
    }

    /**
     * Register the "Subscription Broadcasts" submenu under the main
     * Tanzanite Settings menu.
     */
    public function register_menu() {
        $parent_slug = 'tanzanite-settings';

        add_submenu_page(
            $parent_slug,
            __( 'Subscription Broadcasts', 'tanzanite-subscription' ),
            __( 'Subscription Broadcasts', 'tanzanite-subscription' ),
            'manage_options',
            'tanzanite-subscription-broadcasts',
            array( $this, 'render_page' )
        );
    }

    /**
     * Render the broadcast admin page.
     */
    public function render_page() {
        if ( ! current_user_can( 'manage_options' ) ) {
            return;
        }

        $sent  = isset( $_GET['tsub_sent'] ) ? (int) $_GET['tsub_sent'] : 0; // phpcs:ignore WordPress.Security.NonceVerification
        $error = isset( $_GET['tsub_error'] ) ? sanitize_text_field( wp_unslash( $_GET['tsub_error'] ) ) : ''; // phpcs:ignore WordPress.Security.NonceVerification

        // Fetch a small set of recent posts and products for convenience.
        $recent_posts    = get_posts(
            array(
                'post_type'      => 'post',
                'post_status'    => 'publish',
                'posts_per_page' => 20,
                'orderby'        => 'date',
                'order'          => 'DESC',
            )
        );
        $recent_products = get_posts(
            array(
                'post_type'      => 'tanz_product',
                'post_status'    => 'publish',
                'posts_per_page' => 20,
                'orderby'        => 'date',
                'order'          => 'DESC',
            )
        );
        ?>
        <div class="wrap">
            <h1><?php esc_html_e( 'Subscription Broadcasts', 'tanzanite-subscription' ); ?></h1>

            <?php if ( $sent ) : ?>
                <div class="notice notice-success is-dismissible">
                    <p><?php esc_html_e( 'Broadcast sent. It may take a short while for all emails to be delivered.', 'tanzanite-subscription' ); ?></p>
                </div>
            <?php elseif ( $error ) : ?>
                <div class="notice notice-error is-dismissible">
                    <p><?php echo esc_html( $error ); ?></p>
                </div>
            <?php endif; ?>

            <p>
                <?php esc_html_e( 'Use this tool to send a one-off email to all confirmed subscribers for a specific blog post or product. This is useful for major announcements or product launches.', 'tanzanite-subscription' ); ?>
            </p>

            <form method="post" action="<?php echo esc_url( admin_url( 'admin-post.php' ) ); ?>">
                <?php wp_nonce_field( 'tsub_send_broadcast', 'tsub_broadcast_nonce' ); ?>
                <input type="hidden" name="action" value="tsub_send_broadcast" />

                <table class="form-table" role="presentation">
                    <tr>
                        <th scope="row">
                            <label for="tsub_object_id"><?php esc_html_e( 'Target content', 'tanzanite-subscription' ); ?></label>
                        </th>
                        <td>
                            <select name="object_id" id="tsub_object_id" class="regular-text">
                                <option value=""><?php esc_html_e( 'Select a post or product…', 'tanzanite-subscription' ); ?></option>
                                <?php if ( ! empty( $recent_posts ) ) : ?>
                                    <optgroup label="<?php esc_attr_e( 'Recent blog posts', 'tanzanite-subscription' ); ?>">
                                        <?php foreach ( $recent_posts as $post ) : ?>
                                            <option value="<?php echo esc_attr( $post->ID ); ?>">
                                                <?php echo esc_html( sprintf( '[Post] %s (ID: %d)', get_the_title( $post ), $post->ID ) ); ?>
                                            </option>
                                        <?php endforeach; ?>
                                    </optgroup>
                                <?php endif; ?>

                                <?php if ( ! empty( $recent_products ) ) : ?>
                                    <optgroup label="<?php esc_attr_e( 'Recent products', 'tanzanite-subscription' ); ?>">
                                        <?php foreach ( $recent_products as $product ) : ?>
                                            <option value="<?php echo esc_attr( $product->ID ); ?>">
                                                <?php echo esc_html( sprintf( '[Product] %s (ID: %d)', get_the_title( $product ), $product->ID ) ); ?>
                                            </option>
                                        <?php endforeach; ?>
                                    </optgroup>
                                <?php endif; ?>
                            </select>
                            <p class="description">
                                <?php esc_html_e( 'If the item you need is not listed, you can still enter its ID manually below.', 'tanzanite-subscription' ); ?>
                            </p>
                            <p>
                                <label for="tsub_object_id_manual">
                                    <?php esc_html_e( 'Or enter a specific post / product ID:', 'tanzanite-subscription' ); ?>
                                </label>
                                <input type="number" name="object_id_manual" id="tsub_object_id_manual" class="small-text" />
                            </p>
                        </td>
                    </tr>

                    <tr>
                        <th scope="row">
                            <label for="tsub_subject"><?php esc_html_e( 'Email subject (optional)', 'tanzanite-subscription' ); ?></label>
                        </th>
                        <td>
                            <input type="text" name="subject" id="tsub_subject" class="regular-text" />
                            <p class="description">
                                <?php esc_html_e( 'Leave blank to use a default subject based on the selected post or product title.', 'tanzanite-subscription' ); ?>
                            </p>
                        </td>
                    </tr>

                    <tr>
                        <th scope="row">
                            <label for="tsub_message"><?php esc_html_e( 'Intro message (optional)', 'tanzanite-subscription' ); ?></label>
                        </th>
                        <td>
                            <textarea name="message" id="tsub_message" rows="6" class="large-text"></textarea>
                            <p class="description">
                                <?php esc_html_e( 'This text will appear at the top of the email. If left empty, the system will use the post excerpt or a short summary.', 'tanzanite-subscription' ); ?>
                            </p>
                        </td>
                    </tr>
                </table>

                <?php submit_button( __( 'Send Broadcast', 'tanzanite-subscription' ) ); ?>
            </form>
        </div>
        <?php
    }

    /**
     * Handle broadcast form submission.
     */
    public function handle_send_broadcast() {
        if ( ! current_user_can( 'manage_options' ) ) {
            wp_die( esc_html__( 'You are not allowed to send broadcasts.', 'tanzanite-subscription' ) );
        }

        check_admin_referer( 'tsub_send_broadcast', 'tsub_broadcast_nonce' );

        $redirect_url = add_query_arg(
            array( 'page' => 'tanzanite-subscription-broadcasts' ),
            admin_url( 'admin.php' )
        );

        $object_id       = isset( $_POST['object_id_manual'] ) && (int) $_POST['object_id_manual'] > 0 // phpcs:ignore WordPress.Security.NonceVerification
            ? (int) $_POST['object_id_manual'] // phpcs:ignore WordPress.Security.NonceVerification
            : (int) ( $_POST['object_id'] ?? 0 ); // phpcs:ignore WordPress.Security.NonceVerification

        if ( $object_id <= 0 ) {
            $redirect_url = add_query_arg( 'tsub_error', rawurlencode( __( 'Please select a post or product, or enter a valid ID.', 'tanzanite-subscription' ) ), $redirect_url );
            wp_safe_redirect( $redirect_url );
            exit;
        }

        $post = get_post( $object_id );
        if ( ! $post || ! in_array( $post->post_type, array( 'post', 'tanz_product' ), true ) ) {
            $redirect_url = add_query_arg( 'tsub_error', rawurlencode( __( 'Selected content is not a valid post or product.', 'tanzanite-subscription' ) ), $redirect_url );
            wp_safe_redirect( $redirect_url );
            exit;
        }

        $raw_subject = isset( $_POST['subject'] ) ? wp_unslash( $_POST['subject'] ) : ''; // phpcs:ignore WordPress.Security.NonceVerification
        $subject     = trim( (string) $raw_subject );

        if ( '' === $subject ) {
            if ( 'tanz_product' === $post->post_type ) {
                $subject = sprintf(
                    /* translators: %s: product title */
                    __( 'New product on Tanzanite: %s', 'tanzanite-subscription' ),
                    get_the_title( $post )
                );
            } else {
                $subject = sprintf(
                    /* translators: %s: post title */
                    __( 'New blog post on Tanzanite: %s', 'tanzanite-subscription' ),
                    get_the_title( $post )
                );
            }
        }

        $raw_message = isset( $_POST['message'] ) ? wp_unslash( $_POST['message'] ) : ''; // phpcs:ignore WordPress.Security.NonceVerification
        $message     = trim( (string) $raw_message );

        if ( '' === $message ) {
            if ( has_excerpt( $post ) ) {
                $message = (string) $post->post_excerpt;
            } else {
                $message = wp_trim_words( wp_strip_all_tags( $post->post_content ), 40 );
            }
        }

        $link = get_permalink( $post );

        // Use the same broadcast helper as the automatic notifications.
        $notifications = TSUB_Notifications::instance();
        $notifications->broadcast_to_all_subscribers( $subject, $message, $link );

        $redirect_url = add_query_arg( 'tsub_sent', 1, $redirect_url );
        wp_safe_redirect( $redirect_url );
        exit;
    }
}
