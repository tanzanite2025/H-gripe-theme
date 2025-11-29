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
        add_action( 'admin_post_tsub_export_subscribers', array( $this, 'handle_export_subscribers' ) );
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

        global $wpdb;
        $table_name = $wpdb->prefix . 'tanz_subscribers';

        $active_count       = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$table_name} WHERE confirmed = 1 AND unsubscribed = 0" );
        $pending_count      = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$table_name} WHERE confirmed = 0 AND unsubscribed = 0" );
        $unsubscribed_count = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$table_name} WHERE unsubscribed = 1" );
        $total_count        = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$table_name}" );

        // Determine current status filter for the subscriber list.
        $status_filter = isset( $_GET['tsub_status'] ) ? sanitize_text_field( wp_unslash( $_GET['tsub_status'] ) ) : 'all'; // phpcs:ignore WordPress.Security.NonceVerification
        $valid_statuses = array( 'all', 'active', 'pending', 'unsubscribed' );
        if ( ! in_array( $status_filter, $valid_statuses, true ) ) {
            $status_filter = 'all';
        }

        $where_sql = '';
        if ( 'active' === $status_filter ) {
            $where_sql = 'WHERE confirmed = 1 AND unsubscribed = 0';
        } elseif ( 'pending' === $status_filter ) {
            $where_sql = 'WHERE confirmed = 0 AND unsubscribed = 0';
        } elseif ( 'unsubscribed' === $status_filter ) {
            $where_sql = 'WHERE unsubscribed = 1';
        }

        $subscribers = $wpdb->get_results(
            "SELECT email, created_at, confirmed, unsubscribed FROM {$table_name} {$where_sql} ORDER BY created_at DESC"
        );

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
        <div class="tz-settings-wrapper">
            <div class="tz-settings-header">
                <h1><?php esc_html_e( '订阅群发（Subscription Broadcasts）', 'tanzanite-subscription' ); ?></h1>
                <p>
                    <?php esc_html_e( '通过此工具，你可以针对某一篇博客或某个商品，向所有已确认且未退订的订阅者发送一封一次性的群发邮件，适合重要公告或新品发布。', 'tanzanite-subscription' ); ?>
                </p>
            </div>

            <?php if ( $sent ) : ?>
                <div class="notice notice-success is-dismissible">
                    <p><?php esc_html_e( '群发邮件已发送，所有订阅者收到邮件可能需要一点时间。', 'tanzanite-subscription' ); ?></p>
                </div>
            <?php elseif ( $error ) : ?>
                <div class="notice notice-error is-dismissible">
                    <p><?php echo esc_html( $error ); ?></p>
                </div>
            <?php endif; ?>

            <div class="tz-settings-section">
                <h2><?php esc_html_e( '订阅者总览', 'tanzanite-subscription' ); ?></h2>

                <p>
                    <?php esc_html_e( '本页面发送的群发邮件会发给所有「活跃订阅者」（已确认且未退订）。下面的数字可以帮助你大致了解当前触达范围。', 'tanzanite-subscription' ); ?>
                </p>
                <ul>
                    <li>
                        <strong><?php esc_html_e( '活跃订阅者（本次会收到邮件）', 'tanzanite-subscription' ); ?>:</strong>
                        <?php echo esc_html( number_format_i18n( $active_count ) ); // phpcs:ignore WordPress.Security.EscapeOutput.OutputNotEscaped ?>
                    </li>
                    <li>
                        <strong><?php esc_html_e( '待确认订阅', 'tanzanite-subscription' ); ?>:</strong>
                        <?php echo esc_html( number_format_i18n( $pending_count ) ); // phpcs:ignore WordPress.Security.EscapeOutput.OutputNotEscaped ?>
                    </li>
                    <li>
                        <strong><?php esc_html_e( '已退订', 'tanzanite-subscription' ); ?>:</strong>
                        <?php echo esc_html( number_format_i18n( $unsubscribed_count ) ); // phpcs:ignore WordPress.Security.EscapeOutput.OutputNotEscaped ?>
                    </li>
                    <li>
                        <strong><?php esc_html_e( '累计收集的邮箱总数', 'tanzanite-subscription' ); ?>:</strong>
                        <?php echo esc_html( number_format_i18n( $total_count ) ); // phpcs:ignore WordPress.Security.EscapeOutput.OutputNotEscaped ?>
                    </li>
                </ul>
            </div>

            <div class="tz-settings-section">
                <form method="post" action="<?php echo esc_url( admin_url( 'admin-post.php' ) ); ?>">
                    <?php wp_nonce_field( 'tsub_send_broadcast', 'tsub_broadcast_nonce' ); ?>
                    <?php wp_nonce_field( 'tsub_export_subscribers', 'tsub_export_nonce' ); ?>
                    <input type="hidden" name="action" value="tsub_send_broadcast" />

                    <table class="form-table" role="presentation">
                        <tr>
                            <th scope="row">
                                <label for="tsub_object_id"><?php esc_html_e( '目标内容', 'tanzanite-subscription' ); ?></label>
                            </th>
                            <td>
                                <select name="object_id" id="tsub_object_id" class="regular-text">
                                    <option value=""><?php esc_html_e( '请选择一篇文章或一个商品…', 'tanzanite-subscription' ); ?></option>
                                    <?php if ( ! empty( $recent_posts ) ) : ?>
                                        <optgroup label="<?php esc_attr_e( '最近的博客文章', 'tanzanite-subscription' ); ?>">
                                            <?php foreach ( $recent_posts as $post ) : ?>
                                                <option value="<?php echo esc_attr( $post->ID ); ?>">
                                                    <?php echo esc_html( sprintf( '[Post] %s (ID: %d)', get_the_title( $post ), $post->ID ) ); ?>
                                                </option>
                                            <?php endforeach; ?>
                                        </optgroup>
                                    <?php endif; ?>

                                    <?php if ( ! empty( $recent_products ) ) : ?>
                                        <optgroup label="<?php esc_attr_e( '最近的商品', 'tanzanite-subscription' ); ?>">
                                            <?php foreach ( $recent_products as $product ) : ?>
                                                <option value="<?php echo esc_attr( $product->ID ); ?>">
                                                    <?php echo esc_html( sprintf( '[Product] %s (ID: %d)', get_the_title( $product ), $product->ID ) ); ?>
                                                </option>
                                            <?php endforeach; ?>
                                        </optgroup>
                                    <?php endif; ?>
                                </select>
                                <p class="description">
                                    <?php esc_html_e( '如果需要的内容没有出现在列表中，你也可以在下面直接手动输入文章 / 商品 ID。', 'tanzanite-subscription' ); ?>
                                </p>
                                <p>
                                    <label for="tsub_object_id_manual">
                                        <?php esc_html_e( '或直接输入指定的文章 / 商品 ID：', 'tanzanite-subscription' ); ?>
                                    </label>
                                    <input type="number" name="object_id_manual" id="tsub_object_id_manual" class="small-text" />
                                </p>
                            </td>
                        </tr>

                        <tr>
                            <th scope="row">
                                <label for="tsub_subject"><?php esc_html_e( '邮件标题（可选）', 'tanzanite-subscription' ); ?></label>
                            </th>
                            <td>
                                <input type="text" name="subject" id="tsub_subject" class="regular-text" />
                                <p class="description">
                                    <?php esc_html_e( '如留空，将自动使用所选文章 / 商品的标题生成默认邮件标题。', 'tanzanite-subscription' ); ?>
                                </p>
                            </td>
                        </tr>

                        <tr>
                            <th scope="row">
                                <label for="tsub_message"><?php esc_html_e( '邮件开头引言（可选）', 'tanzanite-subscription' ); ?></label>
                            </th>
                            <td>
                                <textarea name="message" id="tsub_message" rows="6" class="large-text"></textarea>
                                <p class="description">
                                    <?php esc_html_e( '这段文字会出现在邮件正文的最上方；如留空，系统会使用文章摘要或自动生成的一段简短说明。', 'tanzanite-subscription' ); ?>
                                </p>
                            </td>
                        </tr>
                    </table>

                    <h2><?php esc_html_e( '订阅者列表', 'tanzanite-subscription' ); ?></h2>
                    <p>
                        <?php esc_html_e( '你可以通过下面的筛选查看不同状态的订阅者，并勾选「只发送给选中的邮箱」。如不勾选任何邮箱，则会发送给所有活跃订阅者。', 'tanzanite-subscription' ); ?>
                    </p>

                    <p>
                        <?php
                        $base_url = admin_url( 'admin.php?page=tanzanite-subscription-broadcasts' );

                        $filters = array(
                            'all'          => __( '全部', 'tanzanite-subscription' ),
                            'active'       => __( '仅活跃订阅者', 'tanzanite-subscription' ),
                            'pending'      => __( '仅待确认', 'tanzanite-subscription' ),
                            'unsubscribed' => __( '仅已退订', 'tanzanite-subscription' ),
                        );

                        foreach ( $filters as $key => $label ) {
                            $url   = add_query_arg( 'tsub_status', $key, $base_url );
                            $class = ( $key === $status_filter ) ? 'button button-primary' : 'button';
                            echo '<a href="' . esc_url( $url ) . '" class="' . esc_attr( $class ) . '" style="margin-right:4px;">' . esc_html( $label ) . '</a>'; // phpcs:ignore WordPress.Security.EscapeOutput.OutputNotEscaped
                        }

                        ?>
                        <button
                            type="submit"
                            class="button"
                            name="tsub_export_selected"
                            value="1"
                            formaction="<?php echo esc_url( admin_url( 'admin-post.php?action=tsub_export_subscribers' ) ); ?>"
                            style="margin-left:8px;"
                        >
                            <?php esc_html_e( '导出勾选邮箱为 CSV', 'tanzanite-subscription' ); ?>
                        </button>
                    </p>

                    <table class="widefat fixed striped">
                        <thead>
                            <tr>
                                <th class="check-column"><input type="checkbox" disabled="disabled" /></th>
                                <th><?php esc_html_e( '邮箱地址', 'tanzanite-subscription' ); ?></th>
                                <th><?php esc_html_e( '订阅时间', 'tanzanite-subscription' ); ?></th>
                                <th><?php esc_html_e( '状态', 'tanzanite-subscription' ); ?></th>
                            </tr>
                        </thead>
                        <tbody>
                        <?php if ( empty( $subscribers ) ) : ?>
                            <tr>
                                <td colspan="4"><?php esc_html_e( '当前没有符合该筛选条件的订阅者。', 'tanzanite-subscription' ); ?></td>
                            </tr>
                        <?php else : ?>
                            <?php foreach ( $subscribers as $row ) : ?>
                                <?php
                                $email        = isset( $row->email ) ? $row->email : '';
                                $created_at   = isset( $row->created_at ) ? $row->created_at : '';
                                $confirmed    = isset( $row->confirmed ) ? (int) $row->confirmed : 0;
                                $unsubscribed = isset( $row->unsubscribed ) ? (int) $row->unsubscribed : 0;

                                if ( $unsubscribed ) {
                                    $status_label = __( '已退订', 'tanzanite-subscription' );
                                } elseif ( $confirmed ) {
                                    $status_label = __( '已确认（活跃）', 'tanzanite-subscription' );
                                } else {
                                    $status_label = __( '待确认', 'tanzanite-subscription' );
                                }
                                ?>
                                <tr>
                                    <th scope="row" class="check-column">
                                        <?php if ( $email ) : ?>
                                            <input type="checkbox" name="tsub_selected_emails[]" value="<?php echo esc_attr( $email ); ?>" />
                                        <?php endif; ?>
                                    </th>
                                    <td><?php echo esc_html( $email ); ?></td>
                                    <td><?php echo esc_html( $created_at ); ?></td>
                                    <td><?php echo esc_html( $status_label ); ?></td>
                                </tr>
                            <?php endforeach; ?>
                        <?php endif; ?>
                        </tbody>
                    </table>

                    <?php submit_button( __( '发送群发邮件', 'tanzanite-subscription' ) ); ?>
                </form>
            </div>
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

        $selected_emails = array();
        if ( isset( $_POST['tsub_selected_emails'] ) && is_array( $_POST['tsub_selected_emails'] ) ) { // phpcs:ignore WordPress.Security.NonceVerification
            $selected_emails = array_filter( (array) $_POST['tsub_selected_emails'] ); // phpcs:ignore WordPress.Security.NonceVerification
        }

        if ( ! empty( $selected_emails ) ) {
            $notifications->broadcast_to_specific_subscribers( $subject, $message, $link, $selected_emails );
        } else {
            $notifications->broadcast_to_all_subscribers( $subject, $message, $link );
        }

        $redirect_url = add_query_arg( 'tsub_sent', 1, $redirect_url );
        wp_safe_redirect( $redirect_url );
        exit;
    }

    /**
     * Export subscribers as CSV, respecting the current status filter.
     */
    public function handle_export_subscribers() {
        if ( ! current_user_can( 'manage_options' ) ) {
            wp_die( esc_html__( 'You are not allowed to export subscribers.', 'tanzanite-subscription' ) );
        }

        check_admin_referer( 'tsub_export_subscribers', 'tsub_export_nonce' );

        global $wpdb;

        $table_name = $wpdb->prefix . 'tanz_subscribers';
        $rows = array();

        // If the export was triggered from the broadcast form with specific
        // checkboxes selected, prefer exporting only those addresses.
        $export_selected = isset( $_POST['tsub_export_selected'] ) && '1' === (string) $_POST['tsub_export_selected']; // phpcs:ignore WordPress.Security.NonceVerification

        if ( $export_selected && isset( $_POST['tsub_selected_emails'] ) && is_array( $_POST['tsub_selected_emails'] ) ) { // phpcs:ignore WordPress.Security.NonceVerification
            $emails = array_filter( (array) $_POST['tsub_selected_emails'] ); // phpcs:ignore WordPress.Security.NonceVerification
            $emails = array_unique(
                array_filter(
                    array_map( 'sanitize_email', $emails ),
                    static function ( $email ) {
                        return (bool) $email && is_email( $email );
                    }
                )
            );

            if ( ! empty( $emails ) ) {
                $placeholders = implode( ',', array_fill( 0, count( $emails ), '%s' ) );

                $query = $wpdb->prepare(
                    "SELECT email, created_at, confirmed, unsubscribed FROM {$table_name} WHERE email IN ({$placeholders}) ORDER BY created_at DESC",
                    $emails
                );

                $rows = $wpdb->get_results( $query, ARRAY_A );
            }
        }

        // If no valid selection was provided, fall back to exporting all
        // subscribers that match the current status filter (same as the list
        // view on the admin page).
        if ( empty( $rows ) ) {
            $status_filter = isset( $_GET['tsub_status'] ) ? sanitize_text_field( wp_unslash( $_GET['tsub_status'] ) ) : 'all'; // phpcs:ignore WordPress.Security.NonceVerification
            $valid_statuses = array( 'all', 'active', 'pending', 'unsubscribed' );
            if ( ! in_array( $status_filter, $valid_statuses, true ) ) {
                $status_filter = 'all';
            }

            $where_sql = '';
            if ( 'active' === $status_filter ) {
                $where_sql = 'WHERE confirmed = 1 AND unsubscribed = 0';
            } elseif ( 'pending' === $status_filter ) {
                $where_sql = 'WHERE confirmed = 0 AND unsubscribed = 0';
            } elseif ( 'unsubscribed' === $status_filter ) {
                $where_sql = 'WHERE unsubscribed = 1';
            }

            $rows = $wpdb->get_results(
                "SELECT email, created_at, confirmed, unsubscribed FROM {$table_name} {$where_sql} ORDER BY created_at DESC",
                ARRAY_A
            );
        }

        $filename = 'tanzanite-subscribers-' . gmdate( 'Ymd-His' ) . '.csv';

        header( 'Content-Type: text/csv; charset=UTF-8' );
        header( 'Content-Disposition: attachment; filename=' . $filename );
        header( 'Pragma: no-cache' );
        header( 'Expires: 0' );

        $output = fopen( 'php://output', 'w' );
        if ( ! $output ) {
            exit;
        }

        // Header row.
        fputcsv( $output, array( 'email', 'created_at', 'status' ) );

        foreach ( $rows as $row ) {
            $status_label = 'pending';
            if ( (int) $row['unsubscribed'] ) {
                $status_label = 'unsubscribed';
            } elseif ( (int) $row['confirmed'] ) {
                $status_label = 'active';
            }

            fputcsv(
                $output,
                array(
                    $row['email'],
                    $row['created_at'],
                    $status_label,
                )
            );
        }

        fclose( $output );
        exit;
    }
}
