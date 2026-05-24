<?php
/**
 * Deprecated Chat REST API Controller
 *
 * Website authentication now lives in Tanzanite_REST_Auth_Controller.
 * Website customer-service chat now lives in Tanzanite Customer Service.
 *
 * @package    Tanzanite_Settings
 * @subpackage REST_API
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

/**
 * Deprecated no-op controller kept only to avoid fatal errors from old
 * class references. Do not register new settings/chat responsibilities here.
 */
class Tanzanite_REST_Chat_Controller extends Tanzanite_REST_Controller {

	/**
	 * Legacy base path.
	 *
	 * @var string
	 */
	protected $rest_base = 'chat';

	/**
	 * No routes are registered from settings/chat anymore.
	 */
	public function register_routes() {
		return;
	}
}
