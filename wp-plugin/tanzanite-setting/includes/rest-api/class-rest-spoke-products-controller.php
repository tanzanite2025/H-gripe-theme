<?php
/**
 * Spoke Products REST API Controller
 *
 * 为辐条计算器提供简化的 Rim / Hub / Nipple 商品列表。
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class Tanzanite_REST_Spoke_Products_Controller extends Tanzanite_REST_Controller {

    protected $rest_base = 'spoke-products';

    public function register_routes() {
        register_rest_route(
            $this->namespace,
            '/' . $this->rest_base,
            [
                [
                    'methods'             => WP_REST_Server::READABLE,
                    'callback'            => [ $this, 'get_spoke_products' ],
                    'permission_callback' => '__return_true',
                ],
            ]
        );
    }

    public function get_spoke_products( WP_REST_Request $request ) {
        $rims    = $this->get_rim_options();
        $hubs    = $this->get_hub_options();
        $nipples = $this->get_nipple_options();

        return $this->respond_success(
            [
                'rims'    => $rims,
                'hubs'    => $hubs,
                'nipples' => $nipples,
            ]
        );
    }

    private function get_rim_options(): array {
        $args = [
            'post_type'      => 'tanz_product',
            'post_status'    => [ 'publish' ],
            'posts_per_page' => 200,
            'tax_query'      => [
                [
                    'taxonomy' => 'tanz_product_category',
                    'field'    => 'slug',
                    'terms'    => [ 'rim' ],
                ],
            ],
        ];

        $query = new WP_Query( $args );
        $items = [];

        foreach ( $query->posts as $post ) {
            $erd         = get_post_meta( $post->ID, '_tanz_erd', true );
            $spoke_holes = get_post_meta( $post->ID, '_tanz_spoke_holes', true );

            if ( '' === $erd || '' === $spoke_holes ) {
                continue;
            }

            $erd_val   = floatval( $erd );
            $holes_val = intval( $spoke_holes );

            if ( $erd_val <= 0 || $holes_val <= 0 ) {
                continue;
            }

            $label = get_the_title( $post );
            $diam  = get_post_meta( $post->ID, '_tanz_diameter', true );

            if ( $diam ) {
                $label .= ' (' . $diam . ')';
            }

            $items[] = [
                'id'         => (string) $post->ID,
                'label'      => $label,
                'spokeHoles' => $holes_val,
            ];
        }

        return $items;
    }

    private function get_hub_options(): array {
        $args = [
            'post_type'      => 'tanz_product',
            'post_status'    => [ 'publish' ],
            'posts_per_page' => 200,
            'tax_query'      => [
                [
                    'taxonomy' => 'tanz_product_category',
                    'field'    => 'slug',
                    'terms'    => [ 'hub' ],
                ],
            ],
        ];

        $query = new WP_Query( $args );
        $items = [];

        foreach ( $query->posts as $post ) {
            $spoke_holes = get_post_meta( $post->ID, '_tanz_spoke_holes_hub', true );
            if ( '' === $spoke_holes ) {
                continue;
            }

            $holes_val = intval( $spoke_holes );
            if ( $holes_val <= 0 ) {
                continue;
            }

            // 只要前或后任意一侧有完整几何，就认为这条商品可用于辐条计算
            $front_left_pcd   = get_post_meta( $post->ID, '_tanz_front_left_flange_pcd_mm', true );
            $front_right_pcd  = get_post_meta( $post->ID, '_tanz_front_right_flange_pcd_mm', true );
            $front_left_ctr   = get_post_meta( $post->ID, '_tanz_front_left_flange_to_center_mm', true );
            $front_right_ctr  = get_post_meta( $post->ID, '_tanz_front_right_flange_to_center_mm', true );
            $rear_left_pcd    = get_post_meta( $post->ID, '_tanz_rear_left_flange_pcd_mm', true );
            $rear_right_pcd   = get_post_meta( $post->ID, '_tanz_rear_right_flange_pcd_mm', true );
            $rear_left_ctr    = get_post_meta( $post->ID, '_tanz_rear_left_flange_to_center_mm', true );
            $rear_right_ctr   = get_post_meta( $post->ID, '_tanz_rear_right_flange_to_center_mm', true );

            $front_complete = '' !== $front_left_pcd && '' !== $front_right_pcd && '' !== $front_left_ctr && '' !== $front_right_ctr;
            $rear_complete  = '' !== $rear_left_pcd && '' !== $rear_right_pcd && '' !== $rear_left_ctr && '' !== $rear_right_ctr;

            if ( ! $front_complete && ! $rear_complete ) {
                continue;
            }

            $label    = get_the_title( $post );
            $position = 'front-rear-compatible';

            $items[] = [
                'id'         => (string) $post->ID,
                'label'      => $label,
                'position'   => $position,
                'spokeHoles' => $holes_val,
            ];
        }

        return $items;
    }

    private function get_nipple_options(): array {
        $args = [
            'post_type'      => 'tanz_product',
            'post_status'    => [ 'publish' ],
            'posts_per_page' => 200,
            'tax_query'      => [
                [
                    'taxonomy' => 'tanz_product_category',
                    'field'    => 'slug',
                    'terms'    => [ 'nipple' ],
                ],
            ],
        ];

        $query = new WP_Query( $args );
        $items = [];

        foreach ( $query->posts as $post ) {
            $items[] = [
                'id'    => (string) $post->ID,
                'label' => get_the_title( $post ),
            ];
        }

        return $items;
    }
}
