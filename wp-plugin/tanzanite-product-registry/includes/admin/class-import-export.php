<?php
/**
 * 批量导入导出功能
 *
 * @package Tanzanite_Product_Registry
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * 导入导出管理类
 */
class Tanzanite_PR_Import_Export {

	/**
	 * 数据库实例
	 *
	 * @var Tanzanite_PR_Database
	 */
	private $db;

	/**
	 * 构造函数
	 */
	public function __construct() {
		$this->db = new Tanzanite_PR_Database();
		
		// AJAX 处理
		add_action( 'wp_ajax_tanzanite_pr_import_products', array( $this, 'ajax_import_products' ) );
		add_action( 'wp_ajax_tanzanite_pr_export_products', array( $this, 'ajax_export_products' ) );
		add_action( 'wp_ajax_tanzanite_pr_download_template', array( $this, 'ajax_download_template' ) );
	}

	/**
	 * AJAX: 导入产品
	 */
	public function ajax_import_products() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		if ( ! isset( $_FILES['file'] ) ) {
			wp_send_json_error( array( 'message' => '请选择文件' ) );
		}

		$file = $_FILES['file'];
		$ext = strtolower( pathinfo( $file['name'], PATHINFO_EXTENSION ) );

		if ( ! in_array( $ext, array( 'csv', 'xlsx', 'xls' ), true ) ) {
			wp_send_json_error( array( 'message' => '仅支持 CSV、XLSX、XLS 格式' ) );
		}

		// 读取文件内容
		$data = $this->parse_import_file( $file['tmp_name'], $ext );

		if ( is_wp_error( $data ) ) {
			wp_send_json_error( array( 'message' => $data->get_error_message() ) );
		}

		if ( empty( $data ) ) {
			wp_send_json_error( array( 'message' => '文件为空或格式错误' ) );
		}

		// 获取产品类型映射
		$types = $this->db->get_product_types();
		$type_map = array();
		foreach ( $types as $type ) {
			$type_map[ strtolower( $type['type_code'] ) ] = $type['id'];
			$type_map[ strtolower( $type['type_name'] ) ] = $type['id'];
			$type_map[ strtolower( $type['type_name_en'] ) ] = $type['id'];
		}

		$success_count = 0;
		$error_count = 0;
		$errors = array();

		foreach ( $data as $index => $row ) {
			$row_num = $index + 2; // 第一行是表头

			// 验证必填字段
			if ( empty( $row['product_code'] ) ) {
				$errors[] = "第 {$row_num} 行：产品编码不能为空";
				$error_count++;
				continue;
			}

			if ( empty( $row['product_type'] ) ) {
				$errors[] = "第 {$row_num} 行：产品类型不能为空";
				$error_count++;
				continue;
			}

			if ( empty( $row['ship_date'] ) ) {
				$errors[] = "第 {$row_num} 行：出货日期不能为空";
				$error_count++;
				continue;
			}

			// 解析产品类型
			$type_key = strtolower( trim( $row['product_type'] ) );
			if ( ! isset( $type_map[ $type_key ] ) ) {
				$errors[] = "第 {$row_num} 行：未知的产品类型 '{$row['product_type']}'";
				$error_count++;
				continue;
			}

			// 检查编码是否已存在
			$existing = $this->db->get_product_by_code( $row['product_code'] );
			if ( $existing ) {
				$errors[] = "第 {$row_num} 行：产品编码 '{$row['product_code']}' 已存在";
				$error_count++;
				continue;
			}

			// 格式化日期
			$ship_date = $this->parse_date( $row['ship_date'] );
			if ( ! $ship_date ) {
				$errors[] = "第 {$row_num} 行：日期格式错误 '{$row['ship_date']}'";
				$error_count++;
				continue;
			}

			// 准备数据
			$product_data = array(
				'product_code'    => sanitize_text_field( $row['product_code'] ),
				'product_type_id' => $type_map[ $type_key ],
				'product_name'    => sanitize_text_field( $row['product_name'] ?? '' ),
				'ship_date'       => $ship_date,
				'warranty_months' => intval( $row['warranty_months'] ?? 36 ),
				'order_id'        => sanitize_text_field( $row['order_id'] ?? '' ),
				'customer_name'   => sanitize_text_field( $row['customer_name'] ?? '' ),
				'customer_email'  => sanitize_email( $row['customer_email'] ?? '' ),
				'customer_phone'  => sanitize_text_field( $row['customer_phone'] ?? '' ),
				'notes'           => sanitize_textarea_field( $row['notes'] ?? '' ),
			);

			$result = $this->db->save_product( $product_data );

			if ( $result ) {
				$success_count++;
			} else {
				$errors[] = "第 {$row_num} 行：保存失败";
				$error_count++;
			}
		}

		wp_send_json_success( array(
			'message'       => "导入完成：成功 {$success_count} 条，失败 {$error_count} 条",
			'success_count' => $success_count,
			'error_count'   => $error_count,
			'errors'        => array_slice( $errors, 0, 10 ), // 最多显示10条错误
		) );
	}

	/**
	 * 解析导入文件
	 *
	 * @param string $file_path 文件路径
	 * @param string $ext 文件扩展名
	 * @return array|WP_Error
	 */
	private function parse_import_file( $file_path, $ext ) {
		if ( 'csv' === $ext ) {
			return $this->parse_csv( $file_path );
		} else {
			// Excel 文件需要 PhpSpreadsheet 库
			// 如果没有安装，回退到 CSV
			if ( ! class_exists( 'PhpOffice\PhpSpreadsheet\IOFactory' ) ) {
				return new WP_Error( 'no_library', 'Excel 导入需要安装 PhpSpreadsheet 库，请使用 CSV 格式' );
			}
			return $this->parse_excel( $file_path );
		}
	}

	/**
	 * 解析 CSV 文件
	 *
	 * @param string $file_path 文件路径
	 * @return array
	 */
	private function parse_csv( $file_path ) {
		$data = array();
		$headers = array();

		// 检测编码并转换
		$content = file_get_contents( $file_path );
		$encoding = mb_detect_encoding( $content, array( 'UTF-8', 'GBK', 'GB2312', 'BIG5' ), true );
		if ( $encoding && $encoding !== 'UTF-8' ) {
			$content = mb_convert_encoding( $content, 'UTF-8', $encoding );
			file_put_contents( $file_path, $content );
		}

		if ( ( $handle = fopen( $file_path, 'r' ) ) !== false ) {
			$row_index = 0;
			while ( ( $row = fgetcsv( $handle ) ) !== false ) {
				if ( $row_index === 0 ) {
					// 第一行是表头
					$headers = $this->normalize_headers( $row );
				} else {
					$item = array();
					foreach ( $headers as $i => $header ) {
						$item[ $header ] = isset( $row[ $i ] ) ? trim( $row[ $i ] ) : '';
					}
					$data[] = $item;
				}
				$row_index++;
			}
			fclose( $handle );
		}

		return $data;
	}

	/**
	 * 解析 Excel 文件
	 *
	 * @param string $file_path 文件路径
	 * @return array|WP_Error
	 */
	private function parse_excel( $file_path ) {
		try {
			$spreadsheet = \PhpOffice\PhpSpreadsheet\IOFactory::load( $file_path );
			$sheet = $spreadsheet->getActiveSheet();
			$rows = $sheet->toArray();

			if ( empty( $rows ) ) {
				return array();
			}

			$headers = $this->normalize_headers( array_shift( $rows ) );
			$data = array();

			foreach ( $rows as $row ) {
				$item = array();
				foreach ( $headers as $i => $header ) {
					$item[ $header ] = isset( $row[ $i ] ) ? trim( $row[ $i ] ) : '';
				}
				$data[] = $item;
			}

			return $data;
		} catch ( Exception $e ) {
			return new WP_Error( 'parse_error', '解析 Excel 文件失败：' . $e->getMessage() );
		}
	}

	/**
	 * 标准化表头
	 *
	 * @param array $headers 原始表头
	 * @return array
	 */
	private function normalize_headers( $headers ) {
		$map = array(
			'产品编码' => 'product_code',
			'编码'     => 'product_code',
			'code'     => 'product_code',
			'产品类型' => 'product_type',
			'类型'     => 'product_type',
			'type'     => 'product_type',
			'产品名称' => 'product_name',
			'名称'     => 'product_name',
			'name'     => 'product_name',
			'出货日期' => 'ship_date',
			'出货年月' => 'ship_date',
			'日期'     => 'ship_date',
			'date'     => 'ship_date',
			'保修月数' => 'warranty_months',
			'保修期'   => 'warranty_months',
			'warranty' => 'warranty_months',
			'订单号'   => 'order_id',
			'订单'     => 'order_id',
			'order'    => 'order_id',
			'客户姓名' => 'customer_name',
			'客户'     => 'customer_name',
			'customer' => 'customer_name',
			'客户邮箱' => 'customer_email',
			'邮箱'     => 'customer_email',
			'email'    => 'customer_email',
			'客户电话' => 'customer_phone',
			'电话'     => 'customer_phone',
			'phone'    => 'customer_phone',
			'备注'     => 'notes',
			'notes'    => 'notes',
		);

		$normalized = array();
		foreach ( $headers as $header ) {
			$key = strtolower( trim( $header ) );
			$normalized[] = isset( $map[ $key ] ) ? $map[ $key ] : $key;
		}

		return $normalized;
	}

	/**
	 * 解析日期
	 *
	 * @param string $date_str 日期字符串
	 * @return string|false
	 */
	private function parse_date( $date_str ) {
		$date_str = trim( $date_str );

		// 尝试多种格式
		$formats = array(
			'Y-m',      // 2024-12
			'Y/m',      // 2024/12
			'Y-m-d',    // 2024-12-01
			'Y/m/d',    // 2024/12/01
			'd/m/Y',    // 01/12/2024
			'm/d/Y',    // 12/01/2024
		);

		foreach ( $formats as $format ) {
			$date = DateTime::createFromFormat( $format, $date_str );
			if ( $date ) {
				return $date->format( 'Y-m-01' );
			}
		}

		// 尝试中文格式：2024年12月
		if ( preg_match( '/(\d{4})年(\d{1,2})月/', $date_str, $matches ) ) {
			return sprintf( '%04d-%02d-01', $matches[1], $matches[2] );
		}

		return false;
	}

	/**
	 * AJAX: 导出产品
	 */
	public function ajax_export_products() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$args = array(
			'search'   => sanitize_text_field( $_POST['search'] ?? '' ),
			'type_id'  => intval( $_POST['type_id'] ?? 0 ),
			'per_page' => 10000, // 导出所有
		);

		$result = $this->db->get_products( $args );
		$products = $result['items'];

		if ( empty( $products ) ) {
			wp_send_json_error( array( 'message' => '没有数据可导出' ) );
		}

		// 生成 CSV
		$csv_data = $this->generate_csv( $products );

		wp_send_json_success( array(
			'csv'      => $csv_data,
			'filename' => 'products_' . date( 'Y-m-d_His' ) . '.csv',
		) );
	}

	/**
	 * 生成 CSV 数据
	 *
	 * @param array $products 产品列表
	 * @return string
	 */
	private function generate_csv( $products ) {
		$output = fopen( 'php://temp', 'r+' );

		// 添加 BOM 以支持 Excel 中文显示
		fwrite( $output, "\xEF\xBB\xBF" );

		// 表头
		$headers = array(
			'产品编码',
			'产品类型',
			'产品名称',
			'出货日期',
			'保修月数',
			'保修至',
			'保修状态',
			'订单号',
			'客户姓名',
			'客户邮箱',
			'客户电话',
			'备注',
		);
		fputcsv( $output, $headers );

		// 数据行
		foreach ( $products as $product ) {
			// 计算保修状态
			$ship_date = new DateTime( $product['ship_date'] );
			$warranty_months = intval( $product['warranty_months'] );
			$warranty_end = clone $ship_date;
			$warranty_end->modify( "+{$warranty_months} months" );
			$today = new DateTime();
			$status = $warranty_end > $today ? '有效' : '已过期';

			$row = array(
				$product['product_code'],
				$product['type_name'] ?? '',
				$product['product_name'] ?? '',
				substr( $product['ship_date'], 0, 7 ),
				$product['warranty_months'],
				$warranty_end->format( 'Y-m' ),
				$status,
				$product['order_id'] ?? '',
				$product['customer_name'] ?? '',
				$product['customer_email'] ?? '',
				$product['customer_phone'] ?? '',
				$product['notes'] ?? '',
			);
			fputcsv( $output, $row );
		}

		rewind( $output );
		$csv = stream_get_contents( $output );
		fclose( $output );

		return $csv;
	}

	/**
	 * AJAX: 下载导入模板
	 */
	public function ajax_download_template() {
		check_ajax_referer( 'tanzanite_pr_nonce', 'nonce' );

		$output = fopen( 'php://temp', 'r+' );

		// 添加 BOM
		fwrite( $output, "\xEF\xBB\xBF" );

		// 表头
		$headers = array(
			'产品编码',
			'产品类型',
			'产品名称',
			'出货日期',
			'保修月数',
			'订单号',
			'客户姓名',
			'客户邮箱',
			'客户电话',
			'备注',
		);
		fputcsv( $output, $headers );

		// 示例数据
		$example = array(
			'ABC123',
			'hub',
			'TZ-H01 Pro Hub',
			'2024-12',
			'36',
			'ORD-001',
			'张三',
			'zhangsan@example.com',
			'13800138000',
			'示例备注',
		);
		fputcsv( $output, $example );

		rewind( $output );
		$csv = stream_get_contents( $output );
		fclose( $output );

		wp_send_json_success( array(
			'csv'      => $csv,
			'filename' => 'import_template.csv',
		) );
	}
}
