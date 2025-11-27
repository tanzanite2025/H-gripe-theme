<?php
/**
 * Spoke Geometry Admin Page
 *
 * 专用后台页面：为已有商品补充 / 编辑辐条计算所需的 Rim / Hub 几何字段。
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class Tanzanite_Spoke_Geometry_Admin {

    /**
     * 渲染几何管理页面
     */
    public static function render_page() {
        if ( ! current_user_can( 'manage_options' ) ) {
            wp_die( __( '无权限访问此页面。', 'tanzanite-settings' ) );
        }

        $message   = '';
        $error_msg = '';

        // 处理保存
        if ( isset( $_POST['tanz_spoke_geometry_page_nonce'] ) && wp_verify_nonce( $_POST['tanz_spoke_geometry_page_nonce'], 'tanz_spoke_geometry_page_save' ) ) {
            $post_id = isset( $_POST['tanz_spoke_product_id'] ) ? absint( $_POST['tanz_spoke_product_id'] ) : 0;

            if ( $post_id && 'tanz_product' === get_post_type( $post_id ) ) {
                self::save_geometry_meta( $post_id );
                $message = __( '几何参数已保存。', 'tanzanite-settings' );
            } else {
                $error_msg = __( '无效的商品 ID。', 'tanzanite-settings' );
            }
        }

        // 读取筛选参数
        $search   = isset( $_GET['s'] ) ? sanitize_text_field( wp_unslash( $_GET['s'] ) ) : '';
        $category = isset( $_GET['tanz_geom_category'] ) ? sanitize_key( wp_unslash( $_GET['tanz_geom_category'] ) ) : '';
        $paged    = isset( $_GET['paged'] ) ? max( 1, absint( $_GET['paged'] ) ) : 1;
        $per_page = 20;

        $query_args = array(
            'post_type'      => 'tanz_product',
            'post_status'    => array( 'publish', 'pending', 'draft', 'private' ),
            'posts_per_page' => $per_page,
            'paged'          => $paged,
        );

        if ( $search ) {
            $query_args['s'] = $search;
        }

        if ( in_array( $category, array( 'rim', 'hub', 'nipple' ), true ) ) {
            $query_args['tax_query'] = array(
                array(
                    'taxonomy' => 'tanz_product_category',
                    'field'    => 'slug',
                    'terms'    => array( $category ),
                ),
            );
        }

        $products_query = new WP_Query( $query_args );

        $current_product_id = isset( $_GET['product_id'] ) ? absint( $_GET['product_id'] ) : 0;
        $current_product    = $current_product_id ? get_post( $current_product_id ) : null;

        if ( $current_product && 'tanz_product' !== $current_product->post_type ) {
            $current_product = null;
        }

        echo '<div class="wrap tz-settings-wrapper">';
        echo '<h1>' . esc_html__( 'Spoke Geometry / 轮组几何', 'tanzanite-settings' ) . '</h1>';
        echo '<p>' . esc_html__( '为已有商品补充轮圈 / 花鼓几何参数，用于辐条长度计算。', 'tanzanite-settings' ) . '</p>';

        if ( $message ) {
            echo '<div class="notice notice-success is-dismissible"><p>' . esc_html( $message ) . '</p></div>';
        }

        if ( $error_msg ) {
            echo '<div class="notice notice-error"><p>' . esc_html( $error_msg ) . '</p></div>';
        }

        // 当前商品的几何编辑表单
        if ( $current_product ) {
            self::render_geometry_form( $current_product );
        }

        // 筛选表单
        echo '<hr />';
        echo '<form method="get" style="margin-bottom:16px;display:flex;flex-wrap:wrap;gap:8px;align-items:center;">';
        echo '<input type="hidden" name="page" value="tanzanite-spoke-geometry" />';
        echo '<input type="search" name="s" value="' . esc_attr( $search ) . '" placeholder="' . esc_attr__( '按标题搜索…', 'tanzanite-settings' ) . '" class="regular-text" />';

        echo '<select name="tanz_geom_category">';
        echo '<option value="">' . esc_html__( '全部分类', 'tanzanite-settings' ) . '</option>';
        echo '<option value="rim" ' . selected( $category, 'rim', false ) . '>' . esc_html__( 'Rim', 'tanzanite-settings' ) . '</option>';
        echo '<option value="hub" ' . selected( $category, 'hub', false ) . '>' . esc_html__( 'Hub', 'tanzanite-settings' ) . '</option>';
        echo '<option value="nipple" ' . selected( $category, 'nipple', false ) . '>' . esc_html__( 'Nipple', 'tanzanite-settings' ) . '</option>';
        echo '</select>';

        submit_button( __( '筛选', 'tanzanite-settings' ), 'secondary', '', false );

        if ( $search || $category ) {
            $reset_url = admin_url( 'admin.php?page=tanzanite-spoke-geometry' );
            echo ' <a href="' . esc_url( $reset_url ) . '" class="button">' . esc_html__( '清除筛选', 'tanzanite-settings' ) . '</a>';
        }

        echo '</form>';

        // 商品列表
        echo '<table class="wp-list-table widefat fixed striped">';
        echo '<thead><tr>';
        echo '<th style="width:60px;">ID</th>';
        echo '<th>' . esc_html__( '标题', 'tanzanite-settings' ) . '</th>';
        echo '<th style="width:140px;">' . esc_html__( '分类', 'tanzanite-settings' ) . '</th>';
        echo '<th style="width:220px;">' . esc_html__( '几何状态', 'tanzanite-settings' ) . '</th>';
        echo '<th style="width:120px;">' . esc_html__( '操作', 'tanzanite-settings' ) . '</th>';
        echo '</tr></thead><tbody>';

        if ( ! $products_query->have_posts() ) {
            echo '<tr><td colspan="5" style="text-align:center;">' . esc_html__( '暂无商品。', 'tanzanite-settings' ) . '</td></tr>';
        } else {
            while ( $products_query->have_posts() ) {
                $products_query->the_post();
                $post_id = get_the_ID();

                $terms      = get_the_terms( $post_id, 'tanz_product_category' );
                $term_names = array();
                if ( is_array( $terms ) ) {
                    foreach ( $terms as $term ) {
                        $term_names[] = $term->name;
                    }
                }

                list( $rim_status, $hub_status ) = self::get_geometry_status( $post_id );

                $edit_url = add_query_arg(
                    array(
                        'page'       => 'tanzanite-spoke-geometry',
                        'product_id' => $post_id,
                    ),
                    admin_url( 'admin.php' )
                );

                echo '<tr>';
                echo '<td>' . esc_html( $post_id ) . '</td>';
                echo '<td>' . esc_html( get_the_title() ) . '</td>';
                echo '<td>' . esc_html( implode( ', ', $term_names ) ) . '</td>';
                echo '<td>' . esc_html( $rim_status . ' / ' . $hub_status ) . '</td>';
                echo '<td><a href="' . esc_url( $edit_url ) . '" class="button button-small">' . esc_html__( '编辑几何', 'tanzanite-settings' ) . '</a></td>';
                echo '</tr>';
            }
            wp_reset_postdata();
        }

        echo '</tbody></table>';

        // 分页
        if ( $products_query->max_num_pages > 1 ) {
            $page_links = paginate_links(
                array(
                    'base'      => add_query_arg( 'paged', '%#%' ),
                    'format'    => '',
                    'prev_text' => '&laquo;',
                    'next_text' => '&raquo;',
                    'total'     => $products_query->max_num_pages,
                    'current'   => $paged,
                )
            );

            if ( $page_links ) {
                echo '<div class="tablenav"><div class="tablenav-pages">' . wp_kses_post( $page_links ) . '</div></div>';
            }
        }

        echo '</div>';
    }

    /**
     * 保存几何 meta 字段（与 meta box 使用同一批 key）
     */
    private static function save_geometry_meta( int $post_id ) {
        $simple_number_fields = array(
            '_tanz_erd',
            '_tanz_spoke_holes',
            '_tanz_internal_width_mm',
            '_tanz_external_width_mm',
            '_tanz_spoke_holes_hub',
            '_tanz_axle_width_mm',
            '_tanz_front_left_flange_pcd_mm',
            '_tanz_front_right_flange_pcd_mm',
            '_tanz_front_left_flange_to_center_mm',
            '_tanz_front_right_flange_to_center_mm',
            '_tanz_rear_left_flange_pcd_mm',
            '_tanz_rear_right_flange_pcd_mm',
            '_tanz_rear_left_flange_to_center_mm',
            '_tanz_rear_right_flange_to_center_mm',
        );

        foreach ( $simple_number_fields as $key ) {
            if ( isset( $_POST[ $key ] ) && '' !== $_POST[ $key ] ) {
                update_post_meta( $post_id, $key, floatval( wp_unslash( $_POST[ $key ] ) ) );
            } else {
                delete_post_meta( $post_id, $key );
            }
        }

        $text_fields = array(
            '_tanz_diameter',
            '_tanz_nipple_seat_type',
            '_tanz_material',
        );

        foreach ( $text_fields as $key ) {
            if ( isset( $_POST[ $key ] ) && '' !== $_POST[ $key ] ) {
                update_post_meta( $post_id, $key, sanitize_text_field( wp_unslash( $_POST[ $key ] ) ) );
            } else {
                delete_post_meta( $post_id, $key );
            }
        }
    }

    /**
     * 计算当前商品的几何状态文案
     */
    private static function get_geometry_status( int $post_id ): array {
        $erd         = get_post_meta( $post_id, '_tanz_erd', true );
        $spoke_holes = get_post_meta( $post_id, '_tanz_spoke_holes', true );

        $rim_ok = ( '' !== $erd && '' !== $spoke_holes );

        $hub_holes = get_post_meta( $post_id, '_tanz_spoke_holes_hub', true );

        $front_left_pcd   = get_post_meta( $post_id, '_tanz_front_left_flange_pcd_mm', true );
        $front_right_pcd  = get_post_meta( $post_id, '_tanz_front_right_flange_pcd_mm', true );
        $front_left_ctr   = get_post_meta( $post_id, '_tanz_front_left_flange_to_center_mm', true );
        $front_right_ctr  = get_post_meta( $post_id, '_tanz_front_right_flange_to_center_mm', true );
        $rear_left_pcd    = get_post_meta( $post_id, '_tanz_rear_left_flange_pcd_mm', true );
        $rear_right_pcd   = get_post_meta( $post_id, '_tanz_rear_right_flange_pcd_mm', true );
        $rear_left_ctr    = get_post_meta( $post_id, '_tanz_rear_left_flange_to_center_mm', true );
        $rear_right_ctr   = get_post_meta( $post_id, '_tanz_rear_right_flange_to_center_mm', true );

        $front_complete = '' !== $front_left_pcd && '' !== $front_right_pcd && '' !== $front_left_ctr && '' !== $front_right_ctr;
        $rear_complete  = '' !== $rear_left_pcd && '' !== $rear_right_pcd && '' !== $rear_left_ctr && '' !== $rear_right_ctr;

        $hub_ok_parts = array();
        if ( '' !== $hub_holes ) {
            if ( $front_complete ) {
                $hub_ok_parts[] = __( '前花鼓: ✓', 'tanzanite-settings' );
            }
            if ( $rear_complete ) {
                $hub_ok_parts[] = __( '后花鼓: ✓', 'tanzanite-settings' );
            }
        }

        if ( empty( $hub_ok_parts ) ) {
            $hub_status = __( '无花鼓几何', 'tanzanite-settings' );
        } else {
            $hub_status = implode( ' ', $hub_ok_parts );
        }

        $rim_status = $rim_ok ? __( '轮圈: ✓', 'tanzanite-settings' ) : __( '无轮圈几何', 'tanzanite-settings' );

        return array( $rim_status, $hub_status );
    }

    /**
     * 渲染当前商品的几何编辑表单
     */
    private static function render_geometry_form( WP_Post $product ) {
        $post_id = $product->ID;

        $erd               = get_post_meta( $post_id, '_tanz_erd', true );
        $spoke_holes_rim   = get_post_meta( $post_id, '_tanz_spoke_holes', true );
        $diameter_label    = get_post_meta( $post_id, '_tanz_diameter', true );
        $internal_width_mm = get_post_meta( $post_id, '_tanz_internal_width_mm', true );
        $external_width_mm = get_post_meta( $post_id, '_tanz_external_width_mm', true );
        $nipple_seat_type  = get_post_meta( $post_id, '_tanz_nipple_seat_type', true );
        $material          = get_post_meta( $post_id, '_tanz_material', true );

        $spoke_holes_hub = get_post_meta( $post_id, '_tanz_spoke_holes_hub', true );
        $axle_width_mm   = get_post_meta( $post_id, '_tanz_axle_width_mm', true );

        $front_left_flange_pcd_mm        = get_post_meta( $post_id, '_tanz_front_left_flange_pcd_mm', true );
        $front_right_flange_pcd_mm       = get_post_meta( $post_id, '_tanz_front_right_flange_pcd_mm', true );
        $front_left_flange_to_center_mm  = get_post_meta( $post_id, '_tanz_front_left_flange_to_center_mm', true );
        $front_right_flange_to_center_mm = get_post_meta( $post_id, '_tanz_front_right_flange_to_center_mm', true );

        $rear_left_flange_pcd_mm        = get_post_meta( $post_id, '_tanz_rear_left_flange_pcd_mm', true );
        $rear_right_flange_pcd_mm       = get_post_meta( $post_id, '_tanz_rear_right_flange_pcd_mm', true );
        $rear_left_flange_to_center_mm  = get_post_meta( $post_id, '_tanz_rear_left_flange_to_center_mm', true );
        $rear_right_flange_to_center_mm = get_post_meta( $post_id, '_tanz_rear_right_flange_to_center_mm', true );

        echo '<hr />';
        echo '<h2>' . esc_html( get_the_title( $post_id ) ) . '</h2>';

        echo '<form method="post" style="margin-top:12px;">';
        wp_nonce_field( 'tanz_spoke_geometry_page_save', 'tanz_spoke_geometry_page_nonce' );
        echo '<input type="hidden" name="tanz_spoke_product_id" value="' . esc_attr( $post_id ) . '" />';

        echo '<table class="form-table" role="presentation"><tbody>';

        // Rim 字段
        echo '<tr><th scope="row"><label for="tanz_erd">ERD (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_erd" id="tanz_erd" value="' . esc_attr( $erd ) . '" class="small-text" />';
        echo '<p class="description">' . esc_html__( '有效轮径（Effective Rim Diameter），单位 mm。', 'tanzanite-settings' ) . '</p>';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_spoke_holes">Rim spoke holes</label></th><td>';
        echo '<input type="number" min="0" name="_tanz_spoke_holes" id="tanz_spoke_holes" value="' . esc_attr( $spoke_holes_rim ) . '" class="small-text" />';
        echo '<p class="description">' . esc_html__( '轮圈孔数，如 24 / 28 / 32 / 36。', 'tanzanite-settings' ) . '</p>';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_diameter">Diameter label</label></th><td>';
        echo '<input type="text" name="_tanz_diameter" id="tanz_diameter" value="' . esc_attr( $diameter_label ) . '" class="regular-text" />';
        echo '<p class="description">' . esc_html__( '显示用轮径标签，例如 700C / 29"（可选）。', 'tanzanite-settings' ) . '</p>';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_internal_width_mm">Internal width (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_internal_width_mm" id="tanz_internal_width_mm" value="' . esc_attr( $internal_width_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_external_width_mm">External width (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_external_width_mm" id="tanz_external_width_mm" value="' . esc_attr( $external_width_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_nipple_seat_type">Nipple seat type</label></th><td>';
        echo '<input type="text" name="_tanz_nipple_seat_type" id="tanz_nipple_seat_type" value="' . esc_attr( $nipple_seat_type ) . '" class="regular-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_material">Rim material</label></th><td>';
        echo '<input type="text" name="_tanz_material" id="tanz_material" value="' . esc_attr( $material ) . '" class="regular-text" />';
        echo '</td></tr>';

        // Hub
        echo '<tr><th colspan="2"><hr /><strong>' . esc_html__( 'Hub geometry (前/后花鼓成套)', 'tanzanite-settings' ) . '</strong></th></tr>';

        echo '<tr><th scope="row"><label for="tanz_spoke_holes_hub">Hub spoke holes</label></th><td>';
        echo '<input type="number" min="0" name="_tanz_spoke_holes_hub" id="tanz_spoke_holes_hub" value="' . esc_attr( $spoke_holes_hub ) . '" class="small-text" />';
        echo '<p class="description">' . esc_html__( '整套前 / 后花鼓的孔数（通常相同），如 24 / 28 / 32 / 36。', 'tanzanite-settings' ) . '</p>';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_axle_width_mm">Axle width OLD (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_axle_width_mm" id="tanz_axle_width_mm" value="' . esc_attr( $axle_width_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th colspan="2"><strong>' . esc_html__( 'Front hub geometry', 'tanzanite-settings' ) . '</strong></th></tr>';

        echo '<tr><th scope="row"><label for="tanz_front_left_flange_pcd_mm">Front left flange PCD (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_front_left_flange_pcd_mm" id="tanz_front_left_flange_pcd_mm" value="' . esc_attr( $front_left_flange_pcd_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_front_right_flange_pcd_mm">Front right flange PCD (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_front_right_flange_pcd_mm" id="tanz_front_right_flange_pcd_mm" value="' . esc_attr( $front_right_flange_pcd_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_front_left_flange_to_center_mm">Front left flange to center (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_front_left_flange_to_center_mm" id="tanz_front_left_flange_to_center_mm" value="' . esc_attr( $front_left_flange_to_center_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_front_right_flange_to_center_mm">Front right flange to center (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_front_right_flange_to_center_mm" id="tanz_front_right_flange_to_center_mm" value="' . esc_attr( $front_right_flange_to_center_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th colspan="2"><strong>' . esc_html__( 'Rear hub geometry', 'tanzanite-settings' ) . '</strong></th></tr>';

        echo '<tr><th scope="row"><label for="tanz_rear_left_flange_pcd_mm">Rear left flange PCD (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_rear_left_flange_pcd_mm" id="tanz_rear_left_flange_pcd_mm" value="' . esc_attr( $rear_left_flange_pcd_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_rear_right_flange_pcd_mm">Rear right flange PCD (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_rear_right_flange_pcd_mm" id="tanz_rear_right_flange_pcd_mm" value="' . esc_attr( $rear_right_flange_pcd_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_rear_left_flange_to_center_mm">Rear left flange to center (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_rear_left_flange_to_center_mm" id="tanz_rear_left_flange_to_center_mm" value="' . esc_attr( $rear_left_flange_to_center_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '<tr><th scope="row"><label for="tanz_rear_right_flange_to_center_mm">Rear right flange to center (mm)</label></th><td>';
        echo '<input type="number" step="0.1" min="0" name="_tanz_rear_right_flange_to_center_mm" id="tanz_rear_right_flange_to_center_mm" value="' . esc_attr( $rear_right_flange_to_center_mm ) . '" class="small-text" />';
        echo '</td></tr>';

        echo '</tbody></table>';

        submit_button( __( '保存几何参数', 'tanzanite-settings' ) );

        echo '</form>';
    }
}
