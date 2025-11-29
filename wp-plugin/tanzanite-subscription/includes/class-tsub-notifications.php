<?php

namespace TanzaniteSubscription;

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class TSUB_Notifications {
    /** @var TSUB_Notifications|null */
    protected static $instance = null;

    /**
     * Singleton entry point.
     *
     * @return TSUB_Notifications
     */
    public static function instance() {
        if ( null === self::$instance ) {
            self::$instance = new self();
        }

        return self::$instance;
    }

    /**
     * TSUB_Notifications constructor.
     */
    protected function __construct() {
        add_action( 'publish_post', array( $this, 'handle_new_post' ), 10, 2 );
        add_action( 'tanz_new_product_published', array( $this, 'handle_new_product' ), 10, 1 );
    }

    /**
     * Handle publish_post for blog posts.
     *
     * @param int       $post_id Post ID.
     * @param \WP_Post $post    Post object.
     */
    public function handle_new_post( $post_id, $post ) {
        // Only act on standard posts.
        if ( 'post' !== $post->post_type ) {
            return;
        }

        // Respect the auto-notify setting.
        $auto_posts = (int) get_option( 'tanz_sub_auto_notify_posts', 0 );
        if ( 1 !== $auto_posts ) {
            return;
        }

        // Basic safeguard: avoid firing on post updates that are very old imports, etc.
        // For now we always send on publish; this can be refined later if needed.

        $post_title = get_the_title( $post_id );
        $post_link  = get_permalink( $post_id );

        $raw_content = isset( $post->post_content ) ? $post->post_content : '';
        $excerpt     = wp_trim_words( wp_strip_all_tags( $raw_content ), 40 );

        $subject = sprintf(
            /* translators: %s: post title */
            __( 'New blog post on Tanzanite: %s', 'tanzanite-subscription' ),
            $post_title
        );

        $this->broadcast_to_all_subscribers( $subject, $excerpt, $post_link );
    }

    /**
     * Handle custom tanz_new_product_published action.
     *
     * @param int $product_id Product identifier from the tanzanite-setting plugin.
     */
    public function handle_new_product( $product_id ) {
        $auto_products = (int) get_option( 'tanz_sub_auto_notify_products', 0 );
        if ( 1 !== $auto_products ) {
            return;
        }

        // Ask the commerce plugin (tanzanite-setting) for email data via a filter so we stay decoupled.
        $data = apply_filters(
            'tanz_sub_product_email_data',
            array(),
            $product_id
        );

        $title   = isset( $data['title'] ) ? (string) $data['title'] : '';
        $link    = isset( $data['url'] ) ? (string) $data['url'] : '';
        $excerpt = isset( $data['excerpt'] ) ? (string) $data['excerpt'] : '';

        if ( '' === $title || '' === $link ) {
            // Without at least a title and link, we skip sending to avoid confusing emails.
            return;
        }

        $subject = sprintf(
            /* translators: %s: product title */
            __( 'New product on Tanzanite: %s', 'tanzanite-subscription' ),
            $title
        );

        $this->broadcast_to_all_subscribers( $subject, $excerpt, $link );
    }

    /**
     * Broadcast a message with optional excerpt and main link to all confirmed subscribers.
     *
     * Exposed as public so that the manual broadcast admin page can reuse the
     * same logic as the automatic notifications.
     *
     * @param string $subject Email subject.
     * @param string $excerpt Short text body/summary.
     * @param string $main_link Main URL for the content.
     */
    public function broadcast_to_all_subscribers( $subject, $excerpt, $main_link ) {
        global $wpdb;

        $table_name = $wpdb->prefix . 'tanz_subscribers';

        $rows = $wpdb->get_results(
            "SELECT email, unsubscribe_token FROM {$table_name} WHERE confirmed = 1 AND unsubscribed = 0"
        );

        if ( empty( $rows ) ) {
            return;
        }

        $main_link_line = '';
        if ( $main_link ) {
            $main_link_line = sprintf(
                "%s %s\n\n",
                __( 'Read more here:', 'tanzanite-subscription' ),
                $main_link
            );
        }

        foreach ( $rows as $row ) {
            $email = isset( $row->email ) ? $row->email : '';
            if ( ! $email || ! is_email( $email ) ) {
                continue;
            }

            $unsubscribe_url = '';
            if ( ! empty( $row->unsubscribe_token ) ) {
                $unsubscribe_url = add_query_arg(
                    'token',
                    rawurlencode( $row->unsubscribe_token ),
                    rest_url( 'tanz/v1/unsubscribe' )
                );
            }

            $body_lines = array();

            if ( $excerpt ) {
                $body_lines[] = $excerpt;
                $body_lines[] = '';
            }

            if ( $main_link_line ) {
                $body_lines[] = rtrim( $main_link_line );
            }

            if ( $unsubscribe_url ) {
                $body_lines[] = __( 'If you no longer wish to receive these emails, you can unsubscribe here:', 'tanzanite-subscription' );
                $body_lines[] = $unsubscribe_url;
            }

            $message = implode( "\n", $body_lines );

            \tanz_sub_send_mail( $email, $subject, $message );
        }
    }
}
