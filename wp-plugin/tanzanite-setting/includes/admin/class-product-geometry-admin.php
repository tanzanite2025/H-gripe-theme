<?php
/**
 * Product Geometry Admin Meta Box
 *
 * 为 tanz_product 商品添加辐条计算所需的几何参数输入区域。
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class Tanzanite_Product_Geometry_Admin {

    public static function init() {
        // 在商品编辑页添加 meta box
        add_action( 'add_meta_boxes', [ __CLASS__, 'add_meta_box' ] );
        // 保存商品时写入 meta
        add_action( 'save_post_tanz_product', [ __CLASS__, 'save_meta_box' ] );
    }

    public static function add_meta_box() {
        add_meta_box(
            'tanz_spoke_geometry',
            __( 'Spoke Geometry / 轮组几何', 'tanzanite-settings' ),
            [ __CLASS__, 'render_meta_box' ],
            'tanz_product',
            'normal',
            'default'
        );
    }

    public static function render_meta_box( $post ) {
        // 安全 nonce
        wp_nonce_field( 'tanz_spoke_geometry_save', 'tanz_spoke_geometry_nonce' );

        // 轮圈字段
        $erd                 = get_post_meta( $post->ID, '_tanz_erd', true );
        $spoke_holes_rim     = get_post_meta( $post->ID, '_tanz_spoke_holes', true );
        $diameter_label      = get_post_meta( $post->ID, '_tanz_diameter', true );
        $internal_width_mm   = get_post_meta( $post->ID, '_tanz_internal_width_mm', true );
        $external_width_mm   = get_post_meta( $post->ID, '_tanz_external_width_mm', true );
        $nipple_seat_type    = get_post_meta( $post->ID, '_tanz_nipple_seat_type', true );
        $material            = get_post_meta( $post->ID, '_tanz_material', true );

        // 花鼓字段：一条商品可以同时包含前 / 后花鼓几何，方便成对售卖
        $spoke_holes_hub = get_post_meta( $post->ID, '_tanz_spoke_holes_hub', true );
        $axle_width_mm   = get_post_meta( $post->ID, '_tanz_axle_width_mm', true );

        $front_left_flange_pcd_mm        = get_post_meta( $post->ID, '_tanz_front_left_flange_pcd_mm', true );
        $front_right_flange_pcd_mm       = get_post_meta( $post->ID, '_tanz_front_right_flange_pcd_mm', true );
        $front_left_flange_to_center_mm  = get_post_meta( $post->ID, '_tanz_front_left_flange_to_center_mm', true );
        $front_right_flange_to_center_mm = get_post_meta( $post->ID, '_tanz_front_right_flange_to_center_mm', true );

        $rear_left_flange_pcd_mm        = get_post_meta( $post->ID, '_tanz_rear_left_flange_pcd_mm', true );
        $rear_right_flange_pcd_mm       = get_post_meta( $post->ID, '_tanz_rear_right_flange_pcd_mm', true );
        $rear_left_flange_to_center_mm  = get_post_meta( $post->ID, '_tanz_rear_left_flange_to_center_mm', true );
        $rear_right_flange_to_center_mm = get_post_meta( $post->ID, '_tanz_rear_right_flange_to_center_mm', true );

        ?>
        <p style="margin-bottom: 8px; color:#555;">
            <?php esc_html_e( '仅对轮圈 / 花鼓 / 辐条帽等需要参与辐条计算的商品填写。其它商品可以留空。', 'tanzanite-settings' ); ?>
        </p>

        <table class="form-table" role="presentation">
            <tbody>
            <tr>
                <th scope="row"><label for="tanz_erd">ERD (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_erd" id="tanz_erd" value="<?php echo esc_attr( $erd ); ?>" class="small-text" />
                    <p class="description">有效轮径（Effective Rim Diameter），单位 mm。</p>
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_spoke_holes">Rim spoke holes</label></th>
                <td>
                    <input type="number" min="0" name="_tanz_spoke_holes" id="tanz_spoke_holes" value="<?php echo esc_attr( $spoke_holes_rim ); ?>" class="small-text" />
                    <p class="description">轮圈孔数，如 24 / 28 / 32 / 36。</p>
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_diameter">Diameter label</label></th>
                <td>
                    <input type="text" name="_tanz_diameter" id="tanz_diameter" value="<?php echo esc_attr( $diameter_label ); ?>" class="regular-text" />
                    <p class="description">显示用轮径标签，例如 700C / 29"（可选）。</p>
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_internal_width_mm">Internal width (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_internal_width_mm" id="tanz_internal_width_mm" value="<?php echo esc_attr( $internal_width_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_external_width_mm">External width (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_external_width_mm" id="tanz_external_width_mm" value="<?php echo esc_attr( $external_width_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_nipple_seat_type">Nipple seat type</label></th>
                <td>
                    <input type="text" name="_tanz_nipple_seat_type" id="tanz_nipple_seat_type" value="<?php echo esc_attr( $nipple_seat_type ); ?>" class="regular-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_material">Rim material</label></th>
                <td>
                    <input type="text" name="_tanz_material" id="tanz_material" value="<?php echo esc_attr( $material ); ?>" class="regular-text" />
                </td>
            </tr>

            <tr>
                <th colspan="2"><hr /><strong><?php esc_html_e( 'Hub geometry (一条商品可同时包含前 / 后花鼓)', 'tanzanite-settings' ); ?></strong></th>
            </tr>

            <tr>
                <th scope="row"><label for="tanz_spoke_holes_hub">Hub spoke holes</label></th>
                <td>
                    <input type="number" min="0" name="_tanz_spoke_holes_hub" id="tanz_spoke_holes_hub" value="<?php echo esc_attr( $spoke_holes_hub ); ?>" class="small-text" />
                    <p class="description">整套前 / 后花鼓的孔数（通常相同），如 24 / 28 / 32 / 36。</p>
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_axle_width_mm">Axle width OLD (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_axle_width_mm" id="tanz_axle_width_mm" value="<?php echo esc_attr( $axle_width_mm ); ?>" class="small-text" />
                </td>
            </tr>

            <tr>
                <th colspan="2"><strong><?php esc_html_e( 'Front hub geometry', 'tanzanite-settings' ); ?></strong></th>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_front_left_flange_pcd_mm">Front left flange PCD (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_front_left_flange_pcd_mm" id="tanz_front_left_flange_pcd_mm" value="<?php echo esc_attr( $front_left_flange_pcd_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_front_right_flange_pcd_mm">Front right flange PCD (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_front_right_flange_pcd_mm" id="tanz_front_right_flange_pcd_mm" value="<?php echo esc_attr( $front_right_flange_pcd_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_front_left_flange_to_center_mm">Front left flange to center (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_front_left_flange_to_center_mm" id="tanz_front_left_flange_to_center_mm" value="<?php echo esc_attr( $front_left_flange_to_center_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_front_right_flange_to_center_mm">Front right flange to center (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_front_right_flange_to_center_mm" id="tanz_front_right_flange_to_center_mm" value="<?php echo esc_attr( $front_right_flange_to_center_mm ); ?>" class="small-text" />
                </td>
            </tr>

            <tr>
                <th colspan="2"><strong><?php esc_html_e( 'Rear hub geometry', 'tanzanite-settings' ); ?></strong></th>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_rear_left_flange_pcd_mm">Rear left flange PCD (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_rear_left_flange_pcd_mm" id="tanz_rear_left_flange_pcd_mm" value="<?php echo esc_attr( $rear_left_flange_pcd_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_rear_right_flange_pcd_mm">Rear right flange PCD (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_rear_right_flange_pcd_mm" id="tanz_rear_right_flange_pcd_mm" value="<?php echo esc_attr( $rear_right_flange_pcd_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_rear_left_flange_to_center_mm">Rear left flange to center (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_rear_left_flange_to_center_mm" id="tanz_rear_left_flange_to_center_mm" value="<?php echo esc_attr( $rear_left_flange_to_center_mm ); ?>" class="small-text" />
                </td>
            </tr>
            <tr>
                <th scope="row"><label for="tanz_rear_right_flange_to_center_mm">Rear right flange to center (mm)</label></th>
                <td>
                    <input type="number" step="0.1" min="0" name="_tanz_rear_right_flange_to_center_mm" id="tanz_rear_right_flange_to_center_mm" value="<?php echo esc_attr( $rear_right_flange_to_center_mm ); ?>" class="small-text" />
                </td>
            </tr>
            </tbody>
        </table>
        <?php
    }

    public static function save_meta_box( $post_id ) {
        if ( ! isset( $_POST['tanz_spoke_geometry_nonce'] ) || ! wp_verify_nonce( $_POST['tanz_spoke_geometry_nonce'], 'tanz_spoke_geometry_save' ) ) {
            return;
        }

        if ( defined( 'DOING_AUTOSAVE' ) && DOING_AUTOSAVE ) {
            return;
        }

        if ( ! current_user_can( 'edit_post', $post_id ) ) {
            return;
        }

        $simple_number_fields = [
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
        ];

        foreach ( $simple_number_fields as $key ) {
            if ( isset( $_POST[ $key ] ) && $_POST[ $key ] !== '' ) {
                update_post_meta( $post_id, $key, floatval( wp_unslash( $_POST[ $key ] ) ) );
            } else {
                delete_post_meta( $post_id, $key );
            }
        }

        $text_fields = [
            '_tanz_diameter',
            '_tanz_nipple_seat_type',
            '_tanz_material',
        ];

        foreach ( $text_fields as $key ) {
            if ( isset( $_POST[ $key ] ) && $_POST[ $key ] !== '' ) {
                update_post_meta( $post_id, $key, sanitize_text_field( wp_unslash( $_POST[ $key ] ) ) );
            } else {
                delete_post_meta( $post_id, $key );
            }
        }
    }
}
