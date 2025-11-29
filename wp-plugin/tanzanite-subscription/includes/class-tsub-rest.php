<?php

namespace TanzaniteSubscription;

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class TSUB_REST {
    /** @var TSUB_REST|null */
    protected static $instance = null;

    /**
     * Singleton entry point.
     *
     * @return TSUB_REST
     */
    public static function instance() {
        if ( null === self::$instance ) {
            self::$instance = new self();
        }

        return self::$instance;
    }

    /**
     * TSUB_REST constructor.
     */
    protected function __construct() {
        add_action( 'rest_api_init', array( $this, 'register_routes' ) );
    }

    /**
     * Register REST API routes.
     */
    public function register_routes() {
        register_rest_route(
            'tanz/v1',
            '/subscribe',
            array(
                'methods'             => 'POST',
                'callback'            => array( $this, 'handle_subscribe' ),
                'permission_callback' => '__return_true',
            )
        );

        register_rest_route(
            'tanz/v1',
            '/subscribe/confirm',
            array(
                'methods'             => 'GET',
                'callback'            => array( $this, 'handle_confirm' ),
                'permission_callback' => '__return_true',
            )
        );

        register_rest_route(
            'tanz/v1',
            '/unsubscribe',
            array(
                'methods'             => 'GET',
                'callback'            => array( $this, 'handle_unsubscribe' ),
                'permission_callback' => '__return_true',
            )
        );
    }

    /**
     * Handle POST /subscribe.
     *
     * @param \WP_REST_Request $request Request instance.
     * @return \WP_REST_Response|\WP_Error
     */
    public function handle_subscribe( \WP_REST_Request $request ) {
        global $wpdb;

        $raw_email = $request->get_param( 'email' );
        $email     = sanitize_email( (string) $raw_email );

        if ( empty( $email ) || ! is_email( $email ) ) {
            return new \WP_Error(
                'tanz_sub_invalid_email',
                __( 'Please enter a valid email address.', 'tanzanite-subscription' ),
                array( 'status' => 400 )
            );
        }

        $table_name = $wpdb->prefix . 'tanz_subscribers';

        // Look up existing subscriber.
        $subscriber = $wpdb->get_row(
            $wpdb->prepare( "SELECT * FROM {$table_name} WHERE email = %s LIMIT 1", $email )
        );

        $now_mysql        = current_time( 'mysql' );
        $confirm_token    = wp_generate_password( 32, false );
        $unsubscribe_token = wp_generate_password( 32, false );

        if ( $subscriber ) {
            // If user is not unsubscribed, treat as already subscribed to avoid noise.
            if ( (int) $subscriber->unsubscribed === 0 ) {
                // Optionally, we could re-send confirmation if not confirmed yet.
                if ( (int) $subscriber->confirmed === 0 ) {
                    $this->send_confirmation_email( $email, $confirm_token );

                    $wpdb->update(
                        $table_name,
                        array(
                            'confirm_token' => $confirm_token,
                        ),
                        array( 'id' => (int) $subscriber->id ),
                        array( '%s' ),
                        array( '%d' )
                    );

                    return rest_ensure_response(
                        array(
                            'success' => true,
                            'message' => __( 'You are almost subscribed. Please check your email to confirm your subscription.', 'tanzanite-subscription' ),
                        )
                    );
                }

                return rest_ensure_response(
                    array(
                        'success' => true,
                        'message' => __( 'You are already subscribed.', 'tanzanite-subscription' ),
                    )
                );
            }

            // Previously unsubscribed -> treat as re-subscribe.
            $wpdb->update(
                $table_name,
                array(
                    'created_at'        => $now_mysql,
                    'confirmed'         => 0,
                    'unsubscribed'      => 0,
                    'confirm_token'     => $confirm_token,
                    // Reuse existing unsubscribe token if present; otherwise set a new one.
                    'unsubscribe_token' => $subscriber->unsubscribe_token ? $subscriber->unsubscribe_token : $unsubscribe_token,
                ),
                array( 'id' => (int) $subscriber->id ),
                array( '%s', '%d', '%d', '%s', '%s' ),
                array( '%d' )
            );
        } else {
            // New subscriber.
            $wpdb->insert(
                $table_name,
                array(
                    'email'             => $email,
                    'created_at'        => $now_mysql,
                    'confirmed'         => 0,
                    'unsubscribed'      => 0,
                    'confirm_token'     => $confirm_token,
                    'unsubscribe_token' => $unsubscribe_token,
                ),
                array( '%s', '%s', '%d', '%d', '%s', '%s' )
            );
        }

        $this->send_confirmation_email( $email, $confirm_token );

        return rest_ensure_response(
            array(
                'success' => true,
                'message' => __( 'Please check your email to confirm your subscription.', 'tanzanite-subscription' ),
            )
        );
    }

    /**
     * Handle GET /subscribe/confirm.
     *
     * @param \WP_REST_Request $request Request instance.
     * @return \WP_REST_Response
     */
    public function handle_confirm( \WP_REST_Request $request ) {
        global $wpdb;

        $token = sanitize_text_field( (string) $request->get_param( 'token' ) );

        if ( '' === $token ) {
            return rest_ensure_response(
                array(
                    'success' => false,
                    'message' => __( 'Invalid confirmation link.', 'tanzanite-subscription' ),
                )
            );
        }

        $table_name = $wpdb->prefix . 'tanz_subscribers';

        $subscriber = $wpdb->get_row(
            $wpdb->prepare( "SELECT * FROM {$table_name} WHERE confirm_token = %s AND unsubscribed = 0 LIMIT 1", $token )
        );

        if ( ! $subscriber ) {
            return rest_ensure_response(
                array(
                    'success' => false,
                    'message' => __( 'This confirmation link is invalid or has already been used.', 'tanzanite-subscription' ),
                )
            );
        }

        $wpdb->update(
            $table_name,
            array(
                'confirmed' => 1,
            ),
            array( 'id' => (int) $subscriber->id ),
            array( '%d' ),
            array( '%d' )
        );

        return rest_ensure_response(
            array(
                'success' => true,
                'message' => __( 'Thank you, your subscription has been confirmed.', 'tanzanite-subscription' ),
            )
        );
    }

    /**
     * Handle GET /unsubscribe.
     *
     * @param \WP_REST_Request $request Request instance.
     * @return \WP_REST_Response
     */
    public function handle_unsubscribe( \WP_REST_Request $request ) {
        global $wpdb;

        $token = sanitize_text_field( (string) $request->get_param( 'token' ) );

        if ( '' === $token ) {
            return rest_ensure_response(
                array(
                    'success' => false,
                    'message' => __( 'Invalid unsubscribe link.', 'tanzanite-subscription' ),
                )
            );
        }

        $table_name = $wpdb->prefix . 'tanz_subscribers';

        $subscriber = $wpdb->get_row(
            $wpdb->prepare( "SELECT * FROM {$table_name} WHERE unsubscribe_token = %s LIMIT 1", $token )
        );

        if ( ! $subscriber ) {
            return rest_ensure_response(
                array(
                    'success' => false,
                    'message' => __( 'This unsubscribe link is invalid.', 'tanzanite-subscription' ),
                )
            );
        }

        $wpdb->update(
            $table_name,
            array(
                'unsubscribed' => 1,
                'confirmed'    => 0,
            ),
            array( 'id' => (int) $subscriber->id ),
            array( '%d', '%d' ),
            array( '%d' )
        );

        return rest_ensure_response(
            array(
                'success' => true,
                'message' => __( 'You have been unsubscribed.', 'tanzanite-subscription' ),
            )
        );
    }

    /**
     * Send a confirmation email using the helper mail function.
     *
     * @param string $email          Recipient email.
     * @param string $confirm_token  Confirmation token.
     */
    protected function send_confirmation_email( $email, $confirm_token ) {
        $confirm_url = add_query_arg(
            'token',
            rawurlencode( $confirm_token ),
            rest_url( 'tanz/v1/subscribe/confirm' )
        );

        $subject = __( 'Please confirm your Tanzanite subscription', 'tanzanite-subscription' );

        $message = sprintf(
            "%s\n\n%s\n\n%s\n%s\n",
            __( 'Thank you for subscribing to Tanzanite updates.', 'tanzanite-subscription' ),
            __( 'Please click the link below to confirm your subscription:', 'tanzanite-subscription' ),
            $confirm_url,
            __( 'If you did not request this, you can safely ignore this email.', 'tanzanite-subscription' )
        );

        \tanz_sub_send_mail( $email, $subject, $message );
    }
}
