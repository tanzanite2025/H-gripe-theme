<?php
/**
 * REST API: Spoke Database Export Controller
 *
 * Endpoint: /wp-json/tanzanite/v1/spoke-db-export
 * Purpose: Exports all spoke products (Rims & Hubs) in a structured format 
 *          that matches the frontend database.ts structure.
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Tanzanite_REST_Spoke_Export_Controller extends WP_REST_Controller {

	public function __construct() {
		$this->namespace = 'tanzanite/v1';
		$this->rest_base = 'spoke-db-export';
	}

	public function register_routes() {
		register_rest_route(
			$this->namespace,
			'/' . $this->rest_base,
			array(
				array(
					'methods'             => WP_REST_Server::READABLE,
					'callback'            => array( $this, 'get_export_data' ),
					'permission_callback' => array( $this, 'get_items_permissions_check' ),
				),
			)
		);
	}

	public function get_items_permissions_check( $request ) {
		// Allow public access for build scripts, or restrict to admin capability if needed.
		// For simplicity in this sync script context, we allow public read (read-only data).
		return true; 
	}

	/**
	 * Retrieves the full spoke database.
	 *
	 * @return WP_REST_Response
	 */
	public function get_export_data( $request ) {
		$args = array(
			'post_type'      => 'spoke_product',
			'posts_per_page' => -1,
			'post_status'    => 'publish',
		);

		$posts = get_posts( $args );

		$rim_db = array();
		$hub_db = array();

		foreach ( $posts as $post ) {
			$specs = get_post_meta( $post->ID, 'spoke_specs', true );
			$type  = get_post_meta( $post->ID, 'spoke_type', true ); // 'rim' or 'hub'
			$brand = get_post_meta( $post->ID, 'brand_name', true ); // e.g. "DT Swiss"
			
			if ( empty( $specs ) || empty( $type ) || empty( $brand ) ) {
				continue;
			}
			
			// Normalize Brand ID (lowercase, snake_case)
			$brand_id = strtolower( str_replace( ' ', '_', $brand ) );

			$item_data = array(
				'id'   => $specs['id'] ?? sanitize_title( $post->post_title ), // Fallback to slug if no ID in specs
				'name' => $post->post_title,
			);

			if ( $type === 'rim' ) {
				$item_data['erd'] = (float) ( $specs['erd'] ?? 0 );
				
				$this->add_to_db( $rim_db, $brand_id, $brand, $item_data );

			} elseif ( $type === 'hub' ) {
				// Handle Front geometry
				if ( ! empty( $specs['front'] ) ) {
					$item_data['front'] = array(
						'leftFlange'     => (float) $specs['front']['left_flange'],
						'rightFlange'    => (float) $specs['front']['right_flange'],
						'leftFlangePcd'  => (float) $specs['front']['left_pcd'],
						'rightFlangePcd' => (float) $specs['front']['right_pcd'],
					);
				}
				// Handle Rear geometry
				if ( ! empty( $specs['rear'] ) ) {
					$item_data['rear'] = array(
						'leftFlange'     => (float) $specs['rear']['left_flange'],
						'rightFlange'    => (float) $specs['rear']['right_flange'],
						'leftFlangePcd'  => (float) $specs['rear']['left_pcd'],
						'rightFlangePcd' => (float) $specs['rear']['right_pcd'],
					);
				}

				$this->add_to_db( $hub_db, $brand_id, $brand, $item_data );
			}
		}

		return new WP_REST_Response( array(
			'rims' => array_values( $rim_db ),
			'hubs' => array_values( $hub_db ),
		), 200 );
	}

	/**
	 * Helper to add item to hierarchical DB structure
	 */
	private function add_to_db( &$db, $brand_id, $brand_name, $item ) {
		if ( ! isset( $db[ $brand_id ] ) ) {
			$db[ $brand_id ] = array(
				'id'    => $brand_id,
				'name'  => $brand_name,
				'items' => array(),
			);
		}
		$db[ $brand_id ]['items'][] = $item;
	}
}
