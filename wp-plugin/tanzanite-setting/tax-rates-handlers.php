        private function rest_list_tax_rates( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $page     = max( 1, (int) $request->get_param( 'page' ) );
            $per_page = max( 1, min( 200, (int) $request->get_param( 'per_page' ) ) );
            $offset   = ( $page - 1 ) * $per_page;

            $total = (int) $wpdb->get_var( "SELECT COUNT(*) FROM {$this->tax_rates_table}" );
            $rows  = $wpdb->get_results( $wpdb->prepare( "SELECT * FROM {$this->tax_rates_table} ORDER BY sort_order ASC, id ASC LIMIT %d OFFSET %d", $per_page, $offset ), ARRAY_A );

            $items = array_map( [ $this, 'format_tax_rate_row' ], $rows );

            return new \WP_REST_Response(
                [
                    'items' => $items,
                    'meta'  => [
                        'page'        => $page,
                        'per_page'    => $per_page,
                        'total'       => $total,
                        'total_pages' => $per_page ? (int) ceil( $total / $per_page ) : 0,
                    ],
                ]
            );
        }

        private function rest_get_tax_rate( \WP_REST_Request $request ): \WP_REST_Response {
            $row = $this->fetch_tax_rate_row( (int) $request['id'] );

            if ( ! $row ) {
                return $this->respond_error( 'tax_rate_not_found', __( '指定的税率不存在。', 'tanzanite-settings' ), 404 );
            }

            return new \WP_REST_Response( $this->format_tax_rate_row( $row ) );
        }

        private function rest_create_tax_rate( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $data = [
                'name'        => sanitize_text_field( $request->get_param( 'name' ) ),
                'rate'        => (float) $request->get_param( 'rate' ),
                'region'      => sanitize_text_field( $request->get_param( 'region' ) ?: '' ),
                'description' => sanitize_textarea_field( $request->get_param( 'description' ) ?: '' ),
                'is_active'   => (bool) $request->get_param( 'is_active' ),
                'sort_order'  => (int) $request->get_param( 'sort_order' ),
            ];

            $meta = $request->get_param( 'meta' );
            if ( is_array( $meta ) ) {
                $data['meta'] = wp_json_encode( $meta );
            } else {
                $data['meta'] = wp_json_encode( [] );
            }

            $format = [ '%s', '%f', '%s', '%s', '%d', '%d', '%s' ];

            $inserted = $wpdb->insert( $this->tax_rates_table, $data, $format );
            if ( false === $inserted ) {
                return $this->respond_error( 'failed_create_tax_rate', __( '创建税率失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $id   = (int) $wpdb->insert_id;
            $row  = $this->fetch_tax_rate_row( $id );
            $item = $this->format_tax_rate_row( $row );

            $this->log_audit( 'create', 'tax_rate', $id, [ 'name' => $item['name'] ], $request );

            return new \WP_REST_Response( $item, 201 );
        }

        private function rest_update_tax_rate( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $id   = (int) $request['id'];
            $row  = $this->fetch_tax_rate_row( $id );
            if ( ! $row ) {
                return $this->respond_error( 'tax_rate_not_found', __( '指定的税率不存在。', 'tanzanite-settings' ), 404 );
            }

            $data   = [];
            $format = [];

            if ( $request->has_param( 'name' ) ) {
                $data['name'] = sanitize_text_field( $request->get_param( 'name' ) );
                $format[]     = '%s';
            }

            if ( $request->has_param( 'rate' ) ) {
                $data['rate'] = (float) $request->get_param( 'rate' );
                $format[]     = '%f';
            }

            if ( $request->has_param( 'region' ) ) {
                $data['region'] = sanitize_text_field( $request->get_param( 'region' ) );
                $format[]       = '%s';
            }

            if ( $request->has_param( 'description' ) ) {
                $data['description'] = sanitize_textarea_field( $request->get_param( 'description' ) );
                $format[]            = '%s';
            }

            if ( $request->has_param( 'is_active' ) ) {
                $data['is_active'] = (bool) $request->get_param( 'is_active' );
                $format[]          = '%d';
            }

            if ( $request->has_param( 'sort_order' ) ) {
                $data['sort_order'] = (int) $request->get_param( 'sort_order' );
                $format[]           = '%d';
            }

            if ( $request->has_param( 'meta' ) ) {
                $meta = $request->get_param( 'meta' );
                if ( is_array( $meta ) ) {
                    $data['meta'] = wp_json_encode( $meta );
                } else {
                    $data['meta'] = wp_json_encode( [] );
                }
                $format[] = '%s';
            }

            if ( empty( $data ) ) {
                return $this->respond_error( 'invalid_tax_rate_payload', __( '没有可更新的字段。', 'tanzanite-settings' ) );
            }

            $updated = $wpdb->update( $this->tax_rates_table, $data, [ 'id' => $id ], $format, [ '%d' ] );
            if ( false === $updated ) {
                return $this->respond_error( 'failed_update_tax_rate', __( '更新税率失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $updated_row = $this->fetch_tax_rate_row( $id );
            $item        = $this->format_tax_rate_row( $updated_row );

            $this->log_audit( 'update', 'tax_rate', $id, [ 'name' => $item['name'] ], $request );

            return new \WP_REST_Response( $item );
        }

        private function rest_delete_tax_rate( \WP_REST_Request $request ): \WP_REST_Response {
            global $wpdb;

            $id  = (int) $request['id'];
            $row = $this->fetch_tax_rate_row( $id );

            if ( ! $row ) {
                return $this->respond_error( 'tax_rate_not_found', __( '指定的税率不存在。', 'tanzanite-settings' ), 404 );
            }

            $deleted = $wpdb->delete( $this->tax_rates_table, [ 'id' => $id ], [ '%d' ] );
            if ( false === $deleted ) {
                return $this->respond_error( 'failed_delete_tax_rate', __( '删除税率失败，请稍后重试。', 'tanzanite-settings' ), 500 );
            }

            $this->log_audit( 'delete', 'tax_rate', $id, [ 'name' => $row['name'] ], $request );

            return new \WP_REST_Response( [ 'deleted' => true ] );
        }

        private function fetch_tax_rate_row( int $id ): ?array {
            global $wpdb;

            $row = $wpdb->get_row( $wpdb->prepare( "SELECT * FROM {$this->tax_rates_table} WHERE id = %d", $id ), ARRAY_A );

            return $row ?: null;
        }

        private function format_tax_rate_row( array $row ): array {
            $meta = $row['meta'] ? json_decode( $row['meta'], true ) : [];
            if ( ! is_array( $meta ) ) {
                $meta = [];
            }

            return [
                'id'          => (int) $row['id'],
                'name'        => $row['name'],
                'rate'        => (float) $row['rate'],
                'region'      => $row['region'],
                'description' => $row['description'],
                'is_active'   => (bool) $row['is_active'],
                'sort_order'  => (int) $row['sort_order'],
                'meta'        => $meta,
                'created_at'  => $row['created_at'],
                'updated_at'  => $row['updated_at'],
            ];
        }
