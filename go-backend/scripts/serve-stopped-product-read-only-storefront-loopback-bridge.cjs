'use strict';

const http = require('node:http');
const { createHash } = require('node:crypto');
const { mapStoppedProductSharedPublicImageGETToBoundedQuery, verifyStoppedProductPublicImageResultAndExtractOriginalBytes } = require('./stopped-product-shared-public-image-request-mapping-and-bound-response-byte-verification.cjs');

const storefrontVerificationName = 'real_storefront_get_handlers_bound_read_only_sql_without_checkout_snapshots_or_network';
const stoppedStorefrontPageShellGETPaths = Object.freeze({
  '/api/v1/storefront/redirects': 'redirects',
  '/api/v1/seo/home': 'seo_home',
  '/api/v1/settings/social': 'social_settings',
  '/api/v1/settings/website-profile': 'website_profile',
  '/api/v1/settings/website-name': 'website_name',
  '/api/v1/products/specification-templates': 'specification_templates',
  '/api/v1/settings/currency-policy': 'currency_policy',
  '/api/v1/content/faq-pages/by-route': 'product_page_faq',
});

function requireExplicitStoppedProductLoopbackOrigin(value) {
  const parsed = new URL(value);
  if (parsed.protocol !== 'http:' || parsed.hostname !== '127.0.0.1' || !parsed.port || Number(parsed.port) < 1 || parsed.username || parsed.password || parsed.pathname !== '/' || parsed.search || parsed.hash || parsed.origin !== value) {
    throw new Error('An explicit http://127.0.0.1:port origin is required');
  }
  return parsed.origin;
}

function validateStoppedProductLoopbackBridgeConfiguration(configuration) {
  const allowed = new Set(['maintenance_origin', 'maintenance_session', 'task_id', 'contract_sha256', 'storefront_origin', 'locale', 'country', 'currency']);
  if (!configuration || typeof configuration !== 'object' || Array.isArray(configuration) || Object.keys(configuration).length !== allowed.size || Object.keys(configuration).some(key => !allowed.has(key)) || [...allowed].some(key => typeof configuration[key] !== 'string')) throw new Error('Invalid bridge configuration');
  requireExplicitStoppedProductLoopbackOrigin(configuration.maintenance_origin);
  requireExplicitStoppedProductLoopbackOrigin(configuration.storefront_origin);
  if (configuration.maintenance_origin === configuration.storefront_origin || !/^[a-f0-9]{32}$/.test(configuration.task_id) || !/^[a-f0-9]{64}$/.test(configuration.contract_sha256) || /^0+$/.test(configuration.contract_sha256) || !/^[A-Za-z0-9_-]{1,32}$/.test(configuration.locale) || !/^[A-Z]{2}$/.test(configuration.country) || !/^[A-Z]{3}$/.test(configuration.currency) || !/^[A-Za-z0-9_.-]{20,16384}$/.test(configuration.maintenance_session)) throw new Error('Explicit task, contract, session and market identities are required');
  return Object.freeze({ ...configuration });
}

function mapStoppedProductStorefrontGETToBoundedQuery(request, configuration) {
  if (request.method !== 'GET') return { status: 405 };
  if (request.headers.origin && request.headers.origin !== configuration.storefront_origin) return { status: 403 };
  if (request.headers['sec-fetch-site'] === 'cross-site' || request.headers['content-length'] && request.headers['content-length'] !== '0' || request.headers['transfer-encoding']) return { status: 403 };
  // Check before URL parsing, which would silently normalize dot segments.
  if (typeof request.url !== 'string' || request.url.length > 4096 || !request.url.startsWith('/') || request.url.startsWith('//') || /[\\#\x00-\x20]/.test(request.url) || /%(?:2f|5c|2e)/i.test(request.url.split('?')[0]) || /(?:^|\/)\.{1,2}(?:\/|\?|$)/.test(request.url)) return { status: 400 };
  const target = new URL(request.url, 'http://127.0.0.1');
  if (target.pathname.startsWith('/uploads/')) return mapStoppedProductSharedPublicImageGETToBoundedQuery(request, target, configuration);
  const query = { contract_sha256: configuration.contract_sha256, operation: '', locale: request.headers['x-locale'] || configuration.locale, country: request.headers['x-market-country'] || configuration.country, currency: request.headers['x-display-currency'] || configuration.currency, slug: '', keyword: '' };
  const paths = { '/api/v1/storefront/context': 'context', '/api/v1/settings/site': 'site_settings', '/api/v1/products/categories': 'categories', '/api/v1/products': 'list', '/api/v1/customer-service/products': 'search' };
  query.operation = paths[target.pathname] || stoppedStorefrontPageShellGETPaths[target.pathname] || '';
  if (target.pathname === '/api/v1/reviews') query.operation = 'product_reviews';
  const reviewSummary = target.pathname.match(/^\/api\/v1\/reviews\/summary\/([1-9][0-9]{0,9})$/);
  if (reviewSummary) {
    query.operation = 'product_review_summary';
    query.product_id = Number(reviewSummary[1]);
    if (query.product_id > 4294967295) return { status: 400 };
  }
  if (!query.operation && /^\/api\/v1\/products\/[^/]+$/.test(target.pathname)) {
    query.operation = 'detail';
    try { query.slug = decodeURIComponent(target.pathname.slice('/api/v1/products/'.length)); } catch { return { status: 400 }; }
    if (!query.slug || query.slug.length > 512 || query.slug.trim() !== query.slug || /[\x00-\x1f\x7f]/.test(query.slug) || ['specification-templates', 'schwalbe-tire-catalog', 'filterable-attributes', 'search'].includes(query.slug)) return { status: 501 };
  }
  if (!query.operation) return { status: 501 };
  const allowedParameters = new Set(['locale', 'country', 'currency']);
  if (query.operation === 'list' || query.operation === 'search') { allowedParameters.add('page'); allowedParameters.add('page_size'); }
  if (query.operation === 'product_reviews') { allowedParameters.add('product_id'); allowedParameters.add('page'); allowedParameters.add('page_size'); }
  if (query.operation === 'search') allowedParameters.add('keyword');
  if (query.operation === 'product_page_faq') allowedParameters.add('route_path');
  for (const [key, value] of target.searchParams) {
    if (!allowedParameters.has(key) || target.searchParams.getAll(key).length !== 1) return { status: 501 };
    if (query.operation !== 'product_reviews' && (key === 'page' && value !== '1' || key === 'page_size' && value !== '24')) return { status: 501 };
    if (key === 'route_path' && value !== '/products/:slug') return { status: 501 };
    if (key === 'locale' || key === 'country' || key === 'currency' || key === 'keyword') query[key] = value;
  }
  if (query.operation === 'product_page_faq' && target.searchParams.get('route_path') !== '/products/:slug') return { status: 501 };
  if (query.operation === 'product_reviews') {
    const productID = target.searchParams.get('product_id');
    const page = target.searchParams.get('page') ?? '1';
    const pageSize = target.searchParams.get('page_size') ?? '20';
    if (!/^[1-9][0-9]{0,9}$/.test(productID || '') || Number(productID) > 4294967295 || !/^[1-9][0-9]{0,3}$/.test(page) || Number(page) > 1000 || !/^[1-9][0-9]?$/.test(pageSize) || Number(pageSize) > 20) return { status: 400 };
    Object.assign(query, { product_id: Number(productID), page: Number(page), page_size: Number(pageSize) });
  }
  if (!/^[A-Za-z0-9_-]{1,32}$/.test(query.locale) || !/^[A-Z]{2}$/.test(query.country) || !/^[A-Z]{3}$/.test(query.currency) || query.operation === 'search' && (!query.keyword || Buffer.byteLength(query.keyword) > 200 || query.keyword.trim() !== query.keyword || /[\x00-\x1f\x7f]/.test(query.keyword))) return { status: 400 };
  return { query };
}

async function readBoundedStoppedProductMaintenanceJSON(response, signal) {
  if (response.statusCode !== 200 || !/^application\/json\b/i.test(response.headers['content-type'] || '') || response.headers['set-cookie']) { response.destroy(); throw new Error('Maintenance query refused'); }
  const chunks = [];
  let bytes = 0;
  for await (const chunk of response) {
    if (signal.aborted || (bytes += chunk.byteLength) > 6 * 1024 * 1024) { response.destroy(); throw new Error('Maintenance response refused'); }
    chunks.push(chunk);
  }
  return JSON.parse(Buffer.concat(chunks).toString('utf8'));
}

function verifyStoppedProductBridgeResultAndExtractOriginalBytes(result, baseline, query, taskID) {
  if (!result || result.version !== 1 || result.scope !== 'product_catalog_with_images' || result.task_id !== taskID || result.window_id !== baseline.window_id || result.contract_sha256 !== baseline.contract_sha256 || result.database_oid !== baseline.database_oid || result.detail_cache_progress_sha256 !== baseline.detail_cache_progress_sha256 || result.sql_read_only !== true || result.application_remains_stopped !== true || result.online_overwrite_available !== false || result.verification !== storefrontVerificationName || !Array.isArray(result.remaining_checks) || !result.remaining_checks.length || !result.request || Object.keys(result.request).length !== Object.keys(query).length || Object.entries(query).some(([key, value]) => result.request[key] !== value) || ![200, 404].includes(result.http_status) || typeof result.response_bytes_base64 !== 'string' || result.response_bytes_base64.length > 5592408 || /[^A-Za-z0-9+/=]/.test(result.response_bytes_base64)) throw new Error('Stopped storefront result binding refused');
  const bytes = Buffer.from(result.response_bytes_base64, 'base64');
  if (!bytes.length || bytes.length > 4 * 1024 * 1024 || bytes.toString('base64') !== result.response_bytes_base64 || createHash('sha256').update(bytes).digest('hex') !== result.response_sha256) throw new Error('Stopped storefront response integrity refused');
  JSON.parse(bytes.toString('utf8'));
  return bytes;
}

async function startStoppedProductReadOnlyStorefrontLoopbackBridge(input) {
  const configuration = validateStoppedProductLoopbackBridgeConfiguration(input);
  const controller = new AbortController();
  let closed = false;
  let pending = 0;
  let accepted = 0;
  let queue = Promise.resolve();
  const postBoundStoppedProductMaintenanceReadOnlyQuery = async (operation, body) => {
    const timeoutSignal = AbortSignal.any([controller.signal, AbortSignal.timeout(125000)]);
    // Native http.request connects directly to the literal loopback address;
    // ambient proxy variables/global fetch dispatchers must never see a session.
    const encodedBody = Buffer.from(JSON.stringify(body));
    return new Promise((resolve, reject) => {
      const request = http.request(configuration.maintenance_origin + '/api/tasks/' + configuration.task_id + '/' + operation, {
        method: 'POST', agent: false, headers: { Authorization: 'Bearer ' + configuration.maintenance_session, 'Content-Type': 'application/json', 'Content-Length': encodedBody.length }, signal: timeoutSignal,
      }, response => { readBoundedStoppedProductMaintenanceJSON(response, timeoutSignal).then(resolve, reject); });
      request.once('error', reject);
      request.end(encodedBody);
    });
  };
  const { report: baseline } = await postBoundStoppedProductMaintenanceReadOnlyQuery('product-catalog-stopped-projection-inspection', {});
  if (!baseline || baseline.version !== 1 || baseline.scope !== 'product_catalog_with_images' || baseline.task_id !== configuration.task_id || baseline.contract_sha256 !== configuration.contract_sha256 || !/^[a-f0-9]{32}$/.test(baseline.window_id) || !Number.isSafeInteger(baseline.database_oid) || baseline.database_oid <= 0 || !/^[a-f0-9]{64}$/.test(baseline.detail_cache_progress_sha256) || baseline.application_remains_stopped !== true || baseline.online_overwrite_available !== false) throw new Error('Current stopped projection baseline refused');
  let lifetimeTimer;
  const server = http.createServer((request, response) => {
    response.setHeader('Cache-Control', 'no-store');
    response.setHeader('Content-Type', 'application/json; charset=utf-8');
    response.setHeader('X-Content-Type-Options', 'nosniff');
    response.setHeader('Referrer-Policy', 'no-referrer');
    const rejectStoppedProductBridgeHTTPResponse = status => { response.writeHead(status); response.end('{"error":"stopped_read_only_verification_request_refused"}'); };
    if (closed) return rejectStoppedProductBridgeHTTPResponse(503);
    const address = server.address();
    if (request.headers.host !== '127.0.0.1:' + address.port || !['127.0.0.1', '::ffff:127.0.0.1'].includes(request.socket.remoteAddress)) return rejectStoppedProductBridgeHTTPResponse(403);
    const mapped = mapStoppedProductStorefrontGETToBoundedQuery(request, configuration);
    if (!mapped.query) return rejectStoppedProductBridgeHTTPResponse(mapped.status);
    if (pending >= 16 || accepted >= 96) return rejectStoppedProductBridgeHTTPResponse(429);
    pending++;
    accepted++;
    // One maintenance query at a time; the manager owns the original lease and
    // opens/reseals SQL separately for each bound request. No long SQL listener.
    queue = queue.then(async () => {
      if (closed) return rejectStoppedProductBridgeHTTPResponse(503);
      if (response.destroyed) return;
      const operation = mapped.publicImage ? 'product-catalog-stopped-read-only-public-image-http-query' : 'product-catalog-stopped-read-only-storefront-http-query';
      const { result } = await postBoundStoppedProductMaintenanceReadOnlyQuery(operation, mapped.query);
      const bytes = mapped.publicImage ? verifyStoppedProductPublicImageResultAndExtractOriginalBytes(result, baseline, mapped.query, configuration.task_id) : verifyStoppedProductBridgeResultAndExtractOriginalBytes(result, baseline, mapped.query, configuration.task_id);
      if (closed) return rejectStoppedProductBridgeHTTPResponse(503);
      if (mapped.publicImage) response.setHeader('Content-Type', result.http_status === 200 ? result.headers['Content-Type'] : 'text/plain; charset=utf-8');
      response.setHeader('X-Stopped-Product-Contract', configuration.contract_sha256);
      response.setHeader('Content-Length', bytes.length);
      response.writeHead(result.http_status);
      response.end(bytes);
    }).catch(() => {
      if (!response.destroyed && !response.headersSent) rejectStoppedProductBridgeHTTPResponse(503);
      stopStoppedProductReadOnlyStorefrontLoopbackBridge();
    }).finally(() => { pending--; });
  });
  server.requestTimeout = 10000;
  server.headersTimeout = 10000;
  server.keepAliveTimeout = 1000;
  server.maxHeadersCount = 32;
  function stopStoppedProductReadOnlyStorefrontLoopbackBridge() {
    if (closed) return;
    closed = true;
    controller.abort();
    clearTimeout(lifetimeTimer);
    server.close();
    server.closeIdleConnections();
  }
  server.on('error', stopStoppedProductReadOnlyStorefrontLoopbackBridge);
  await new Promise((resolve, reject) => { server.once('error', reject); server.listen(0, '127.0.0.1', resolve); });
  lifetimeTimer = setTimeout(stopStoppedProductReadOnlyStorefrontLoopbackBridge, 10 * 60 * 1000);
  return { origin: 'http://127.0.0.1:' + server.address().port, stop: stopStoppedProductReadOnlyStorefrontLoopbackBridge, closed: () => closed };
}

async function readStoppedProductBridgeConfigurationFromStandardInput() {
  let bytes = 0;
  const chunks = [];
  const timeout = setTimeout(() => { process.stderr.write('Bridge configuration input timed out\n'); process.exit(1); }, 30000);
  try {
    for await (const chunk of process.stdin) {
      if ((bytes += chunk.length) > 32768) throw new Error('Configuration input exceeds limit');
      chunks.push(chunk);
    }
    return JSON.parse(Buffer.concat(chunks).toString('utf8'));
  } finally { clearTimeout(timeout); }
}

if (require.main === module) {
  (async () => {
    const bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(await readStoppedProductBridgeConfigurationFromStandardInput());
    process.stdout.write(JSON.stringify({ read_only_api_origin: bridge.origin, lifetime_seconds: 600, maximum_queries: 96, full_page_acceptance_complete: false }) + '\n');
    process.once('SIGINT', bridge.stop);
    process.once('SIGTERM', bridge.stop);
  })().catch(() => { process.stderr.write('Stopped product read-only bridge could not start; verify the task, current contract and maintenance session\n'); process.exitCode = 1; });
}

module.exports = { startStoppedProductReadOnlyStorefrontLoopbackBridge, mapStoppedProductStorefrontGETToBoundedQuery, validateStoppedProductLoopbackBridgeConfiguration, verifyStoppedProductBridgeResultAndExtractOriginalBytes };
