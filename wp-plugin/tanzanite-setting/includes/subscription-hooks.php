<?php
/**
 * Subscription-related hooks for Tanzanite Settings.
 *
 * - Emits a custom action when a tanz_product transitions into publish status.
 * - Provides product data for the Tanzanite Subscription plugin via filter.
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * Fire custom hook when a product post transitions into publish status.
 *
 * This keeps the subscription plugin decoupled from the internal
 * implementation details of product creation / updates.
 *
 * @param string   $new_status New post status.
 * @param string   $old_status Old post status.
 * @param \WP_Post $post       Post object.
 */
function tanz_settings_handle_product_status_transition( $new_status, $old_status, $post ) {
	if ( ! $post instanceof \WP_Post ) {
		return;
	}

	if ( 'tanz_product' !== $post->post_type ) {
		return;
	}

	// Only care about transitions *into* publish.
	if ( 'publish' !== $new_status || 'publish' === $old_status ) {
		return;
	}

	/**
	 * Signal to external modules (e.g. tanzanite-subscription) that a
	 * product has just been published.
	 */
	do_action( 'tanz_new_product_published', (int) $post->ID );
}
add_action( 'transition_post_status', 'tanz_settings_handle_product_status_transition', 10, 3 );

/**
 * Provide product email data for the subscription plugin.
 *
 * This is used by Tanzanite Subscription via the `tanz_sub_product_email_data`
 * filter to build notification emails for newly published products.
 *
 * @param array $data       Existing data (usually empty).
 * @param int   $product_id Product post ID.
 *
 * @return array Filtered data with at least `title` and `url` if available.
 */
function tanz_settings_product_email_data( $data, $product_id ) {
	$product_id = (int) $product_id;
	if ( $product_id <= 0 ) {
		return $data;
	}

	$post = get_post( $product_id );
	if ( ! $post || 'tanz_product' !== $post->post_type ) {
		return $data;
	}

	$title = get_the_title( $post );
	$url   = get_permalink( $post );

	if ( ! $title || ! $url ) {
		return $data;
	}

	// Prefer explicit excerpt if available, otherwise fall back to trimmed content.
	$excerpt = '';
	if ( has_excerpt( $post ) ) {
		$excerpt = $post->post_excerpt;
	} else {
		$excerpt = wp_trim_words( wp_strip_all_tags( $post->post_content ), 40 );
	}

	return array(
		'title'   => $title,
		'url'     => $url,
		'excerpt' => $excerpt,
	);
}
add_filter( 'tanz_sub_product_email_data', 'tanz_settings_product_email_data', 10, 2 );
