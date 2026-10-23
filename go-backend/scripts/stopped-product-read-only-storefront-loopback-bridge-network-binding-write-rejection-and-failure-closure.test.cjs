'use strict';
const test = require('node:test');
const assert = require('node:assert/strict');
const http = require('node:http');
const { createHash } = require('node:crypto');
const { startStoppedProductReadOnlyStorefrontLoopbackBridge, mapStoppedProductStorefrontGETToBoundedQuery, validateStoppedProductLoopbackBridgeConfiguration, verifyStoppedProductBridgeResultAndExtractOriginalBytes } = require('./serve-stopped-product-read-only-storefront-loopback-bridge.cjs');

const configuration = { maintenance_origin: 'http://127.0.0.1:12345', maintenance_session: 'synthetic-maintenance-session-value', storefront_origin: 'http://127.0.0.1:3000', task_id: 'a'.repeat(32), contract_sha256: 'b'.repeat(64), locale: 'en', country: 'AQ', currency: 'USD' };
const baseline = { version: 1, scope: 'product_catalog_with_images', task_id: configuration.task_id, contract_sha256: configuration.contract_sha256, window_id: 'c'.repeat(32), database_oid: 20001, detail_cache_progress_sha256: 'd'.repeat(64), application_remains_stopped: true, online_overwrite_available: false };
const originalBytes = Buffer.from('{"data":{"name":"恢复商品 <原值>","id":9007199254740993}}');

function createSyntheticBoundStoppedStorefrontResult(query, body = originalBytes) {
  return { version: 1, scope: baseline.scope, task_id: baseline.task_id, contract_sha256: baseline.contract_sha256, window_id: baseline.window_id, database_oid: baseline.database_oid, detail_cache_progress_sha256: baseline.detail_cache_progress_sha256, request: query, http_status: query.slug === 'missing' ? 404 : 200, response_bytes_base64: body.toString('base64'), response_sha256: createHash('sha256').update(body).digest('hex'), sql_read_only: true, application_remains_stopped: true, online_overwrite_available: false, verification: 'real_storefront_get_handlers_bound_read_only_sql_without_checkout_snapshots_or_network', remaining_checks: ['actual_nuxt_build_and_cdn'] };
}

async function createSyntheticMaintenanceNetworkServerForStoppedBridge(transformResult = result => result, transformBaseline = report => report, queryResponseStatus = 200) {
  const calls = [];
  let active = 0;
  let maximumActive = 0;
  const server = http.createServer(async (request, response) => {
    assert.equal(request.method, 'POST');
    assert.equal(request.headers.authorization, 'Bearer ' + configuration.maintenance_session);
    assert.equal(request.headers.cookie, undefined);
    active++; maximumActive = Math.max(maximumActive, active);
    const chunks = [];
    for await (const chunk of request) chunks.push(chunk);
    const body = JSON.parse(Buffer.concat(chunks));
    calls.push({ url: request.url, body });
    await new Promise(resolve => setTimeout(resolve, 2));
    response.setHeader('Content-Type', 'application/json');
    if (request.url.endsWith('/product-catalog-stopped-projection-inspection')) response.end(JSON.stringify({ report: transformBaseline({ ...baseline }) }));
    else { response.statusCode = queryResponseStatus; if (queryResponseStatus === 302) response.setHeader('Location', 'http://external-redirect-must-never-be-followed.invalid'); response.end(JSON.stringify({ result: transformResult(createSyntheticBoundStoppedStorefrontResult(body)) })); }
    active--;
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  return { configuration: { ...configuration, maintenance_origin: 'http://127.0.0.1:' + server.address().port }, calls, maximumActive: () => maximumActive, stop: async () => { server.closeAllConnections(); await new Promise(resolve => server.close(resolve)); } };
}

test('loopback configuration rejects network destinations, credentials in URLs and unknown fields', () => {
  validateStoppedProductLoopbackBridgeConfiguration(configuration);
  const missingLocale={...configuration};delete missingLocale.locale;
  assert.throws(()=>validateStoppedProductLoopbackBridgeConfiguration(missingLocale));
  for (const replacement of [{ maintenance_origin: 'https://example.com' }, { maintenance_origin: 'http://localhost:12345' }, { maintenance_origin: 'http://127.0.0.1:12345/path' }, { maintenance_origin: 'http://user:pass@127.0.0.1:12345' }, { maintenance_origin: 'http://127.0.0.1:12345?token=synthetic' }, { maintenance_origin: 'http://127.0.0.1:12345/' }, { storefront_origin: configuration.maintenance_origin }, { task_id: '../wrong' }, { contract_sha256: '0'.repeat(64) }, { locale: 'en/../admin' }, { country: 'aq' }, { currency: 'usd' }, { maintenance_session: 'bad\nheader' }, { database_url: 'never-accepted' }]) assert.throws(() => validateStoppedProductLoopbackBridgeConfiguration({ ...configuration, ...replacement }));
});

test('only bounded genuine GET targets are mapped and unsupported dependencies fail explicitly', () => {
  for (const [url, operation] of [['/api/v1/storefront/context', 'context'], ['/api/v1/settings/site', 'site_settings'], ['/api/v1/products/categories', 'categories'], ['/api/v1/products', 'list'], ['/api/v1/products/恢复商品', 'detail'], ['/api/v1/customer-service/products?keyword=product', 'search']]) assert.equal(mapStoppedProductStorefrontGETToBoundedQuery({ method: 'GET', url, headers: {} }, configuration).query.operation, operation);
  for (const url of ['/api/v1/orders', '/api/v1/products?page=2', '/api/v1/products?page_size=12', '/api/v1/products?featured=true', '/api/v1/products?locale=en&locale=fr', '/api/v1/products/search', '/api/v1/storefront/context?sql=UPDATE', '/uploads/visual-showcase/test.png']) assert.equal(mapStoppedProductStorefrontGETToBoundedQuery({ method: 'GET', url, headers: {} }, configuration).status, 501);
  for (const url of ['//example.invalid/api/v1/products', '/api/v1/products/%2e%2e', '/api/v1/products/%2Fadmin', '/api/v1/products/../settings/site', '/api/v1/products/a\\b', '/api/v1/products/a#fragment']) assert.equal(mapStoppedProductStorefrontGETToBoundedQuery({ method: 'GET', url, headers: {} }, configuration).status, 400);
  const headerContext = mapStoppedProductStorefrontGETToBoundedQuery({ method: 'GET', url: '/api/v1/products?currency=GBP', headers: { 'x-display-currency': 'EUR', 'x-market-country': 'GB', 'x-locale': 'fr' } }, configuration).query;
  assert.equal(headerContext.currency, 'GBP'); assert.equal(headerContext.country, 'GB'); assert.equal(headerContext.locale, 'fr');
});

test('result digest preserves exact JSON bytes and every binding must match', () => {
  const query = mapStoppedProductStorefrontGETToBoundedQuery({ method: 'GET', url: '/api/v1/products/recovered', headers: {} }, configuration).query;
  const result = createSyntheticBoundStoppedStorefrontResult(query);
  assert.deepEqual(verifyStoppedProductBridgeResultAndExtractOriginalBytes(result, baseline, query, configuration.task_id), originalBytes);
  for (const replacement of [{ version: 0 }, { scope: 'whole_storefront' }, { task_id: 'e'.repeat(32) }, { window_id: 'e'.repeat(32) }, { contract_sha256: 'e'.repeat(64) }, { database_oid: 20002 }, { detail_cache_progress_sha256: 'e'.repeat(64) }, { sql_read_only: false }, { application_remains_stopped: false }, { online_overwrite_available: true }, { http_status: 500 }, { response_sha256: 'e'.repeat(64) }, { request: { ...query, slug: 'different' } }, { verification: 'mock' }, { remaining_checks: [] }, { response_bytes_base64: result.response_bytes_base64 + '=' }]) assert.throws(() => verifyStoppedProductBridgeResultAndExtractOriginalBytes({ ...result, ...replacement }, baseline, query, configuration.task_id));
  const largeBytes = Buffer.from(JSON.stringify({ content: 'a'.repeat(4 * 1024 * 1024 - 14) }));
  assert.deepEqual(verifyStoppedProductBridgeResultAndExtractOriginalBytes(createSyntheticBoundStoppedStorefrontResult(query, largeBytes), baseline, query, configuration.task_id), largeBytes);
});

test('real loopback HTTP serializes concurrent reads, returns exact bytes and never forwards writes/cookies', async () => {
  const upstream = await createSyntheticMaintenanceNetworkServerForStoppedBridge();
  let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    const responses = await Promise.all(Array.from({ length: 8 }, (_, index) => fetch(bridge.origin + '/api/v1/products/' + index, { headers: { Cookie: 'admin_auth_token=must-never-forward' } })));
    for (const response of responses) { assert.equal(response.status, 200); assert.equal(response.headers.get('cache-control'), 'no-store'); assert.equal(response.headers.get('access-control-allow-origin'), null); assert.deepEqual(Buffer.from(await response.arrayBuffer()), originalBytes); }
    assert.equal(upstream.maximumActive(), 1);
    assert.equal(upstream.calls.length, 9);
    for (const [path, options, status] of [['/api/v1/orders', {}, 501], ['/api/v1/products', { method: 'POST', body: '{}' }, 405], ['/api/v1/products', { method: 'DELETE' }, 405], ['/api/v1/products', { headers: { Origin: 'https://cross-site.invalid' } }, 403]]) { const response = await fetch(bridge.origin + path, options); assert.equal(response.status, status, JSON.stringify(options)); await response.text(); }
    const reboundStatus = await new Promise((resolve, reject) => { const request = http.request(bridge.origin + '/api/v1/products', { headers: { Host: 'rebound.invalid' } }, response => { response.resume(); response.once('end', () => resolve(response.statusCode)); }); request.once('error', reject); request.end(); });
    assert.equal(reboundStatus, 403);
    assert.equal(upstream.calls.length, 9);
    const missing = await fetch(bridge.origin + '/api/v1/products/missing'); assert.equal(missing.status, 404); await missing.text();
  } finally { bridge?.stop(); await upstream.stop(); }
});

for (const [name, transform, status] of [['wrong_database', result => ({ ...result, database_oid: 7 }), 200], ['wrong_digest', result => ({ ...result, response_sha256: 'f'.repeat(64) }), 200], ['missing_result', () => null, 200], ['expired_session', undefined, 401], ['redirect', undefined, 302]]) {
  test('upstream ' + name + ' closes bridge and refuses queued and future results', async () => {
    const upstream = await createSyntheticMaintenanceNetworkServerForStoppedBridge(transform, undefined, status);
    let bridge;
    try {
      bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
      const responses = await Promise.allSettled([fetch(bridge.origin + '/api/v1/products/a'), fetch(bridge.origin + '/api/v1/products/b')]);
      for (const response of responses) if (response.status === 'fulfilled') { assert.equal(response.value.status, 503); await response.value.text(); }
      assert.equal(bridge.closed(), true);
      assert.equal(upstream.calls.length, 2, 'queued query must not reach maintenance after the failure');
      await assert.rejects(fetch(bridge.origin + '/api/v1/products/c'));
    } finally { bridge?.stop(); await upstream.stop(); }
  });
}

test('stale startup baseline never opens a bridge listener', async () => {
  const upstream = await createSyntheticMaintenanceNetworkServerForStoppedBridge(undefined, report => ({ ...report, contract_sha256: 'e'.repeat(64) }));
  try { await assert.rejects(startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration)); assert.equal(upstream.calls.length, 1); } finally { await upstream.stop(); }
});

test('page shell dependencies use fixed genuine routes, preserve request context and reject unbound inputs over TCP', async () => {
  const upstream = await createSyntheticMaintenanceNetworkServerForStoppedBridge();
  let bridge;
  const dependencies = [
    ['/api/v1/storefront/redirects', 'redirects'], ['/api/v1/seo/home', 'seo_home'],
    ['/api/v1/settings/social', 'social_settings'], ['/api/v1/settings/website-profile', 'website_profile'],
    ['/api/v1/settings/website-name', 'website_name'], ['/api/v1/products/specification-templates', 'specification_templates'],
    ['/api/v1/settings/currency-policy', 'currency_policy'],
  ];
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    for (const [path, operation] of dependencies) {
      const response = await fetch(bridge.origin + path + '?locale=fr&country=FR&currency=EUR');
      assert.equal(response.status, 200);
      assert.deepEqual(Buffer.from(await response.arrayBuffer()), originalBytes);
      const call = upstream.calls.at(-1);
      assert.equal(call.body.operation, operation);
      assert.equal(call.body.locale, 'fr'); assert.equal(call.body.country, 'FR'); assert.equal(call.body.currency, 'EUR');
      assert.equal(call.body.slug, ''); assert.equal(call.body.keyword, '');
      for (const suffix of ['?page=2', '?keyword=private', '?slug=wrong', '?locale=en&locale=fr', '?sql=DELETE']) {
        const refused = await fetch(bridge.origin + path + suffix);
        assert.equal(refused.status, 501); await refused.text();
      }
    }
    for (const path of ['/api/v1/settings/private', '/api/v1/settings/group/website_name', '/api/v1/content/faq-pages/by-route?route_path=/company/about', '/api/v1/payment/methods', '/api/v1/reviews/my', '/api/v1/analytics']) {
      const response = await fetch(bridge.origin + path);
      assert.equal(response.status, 501); await response.text();
    }
    assert.equal(upstream.calls.length, dependencies.length + 1);
    assert.equal(upstream.maximumActive(), 1);
  } finally { bridge?.stop(); await upstream.stop(); }
});

test('product FAQ binds its exact route template and rejects missing, duplicate and alternative targets', async () => {
  const upstream = await createSyntheticMaintenanceNetworkServerForStoppedBridge();
  let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    const path = '/api/v1/content/faq-pages/by-route';
    const response = await fetch(bridge.origin + path + '?route_path=%2Fproducts%2F%3Aslug&locale=en');
    assert.equal(response.status, 200); await response.text();
    assert.equal(upstream.calls.at(-1).body.operation, 'product_page_faq');
    assert.equal(upstream.calls.at(-1).body.slug, '');
    for (const suffix of ['', '?route_path=/products/real-product', '?route_path=/products/:slug&route_path=/products/:slug', '?route_path=https://remote.invalid', '?route_path=/products/:slug&page_id=private']) {
      const refused = await fetch(bridge.origin + path + suffix);
      assert.equal(refused.status, 501); await refused.text();
    }
    assert.equal(upstream.calls.length, 2);
  } finally { bridge?.stop(); await upstream.stop(); }
});

test('product review pages and summaries preserve numeric selectors over TCP and reject unsupported identities and filters', async () => {
  const upstream = await createSyntheticMaintenanceNetworkServerForStoppedBridge();
  let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    for (const [target,expected] of [
      ['/api/v1/reviews?product_id=42&page=2&page_size=5',{operation:'product_reviews',product_id:42,page:2,page_size:5}],
      ['/api/v1/reviews?product_id=4294967295&page=1000&page_size=20',{operation:'product_reviews',product_id:4294967295,page:1000,page_size:20}],
      ['/api/v1/reviews?product_id=42',{operation:'product_reviews',product_id:42,page:1,page_size:20}],
      ['/api/v1/reviews/summary/42',{operation:'product_review_summary',product_id:42}],
    ]) {
      const response = await fetch(bridge.origin+target,{headers:{Cookie:'must-never-forward=private'}});
      assert.equal(response.status,200);assert.deepEqual(Buffer.from(await response.arrayBuffer()),originalBytes);
      const query=upstream.calls.at(-1).body;
      for(const [key,value] of Object.entries(expected))assert.equal(query[key],value);
      assert.equal(query.slug,'');assert.equal(query.keyword,'');
      if(expected.operation==='product_review_summary'){assert.equal(query.page,undefined);assert.equal(query.page_size,undefined);}
    }
    const acceptedCalls=upstream.calls.length;
    for(const suffix of ['', '?product_id=0','?product_id=-1','?product_id=01','?product_id=1e2','?product_id=4294967296','?product_id=42&page=0','?product_id=42&page=1001','?product_id=42&page=01','?product_id=42&page_size=21','?product_id=42&page_size=0']){
      const response=await fetch(bridge.origin+'/api/v1/reviews'+suffix);assert.equal(response.status,400);await response.text();
    }
    for(const target of ['/api/v1/reviews?product_id=42&status=pending','/api/v1/reviews?product_id=42&user_id=7','/api/v1/reviews?product_id=42&product_id=43','/api/v1/reviews?product_id=42&page=1&page=2','/api/v1/reviews/summary/42?page=1','/api/v1/reviews/summary/42?product_id=43','/api/v1/reviews/summary/01','/api/v1/reviews/my','/api/v1/reviews/42','/api/v1/products?product_id=42']){
      const response=await fetch(bridge.origin+target);assert.equal(response.status,501);await response.text();
    }
    assert.equal(upstream.calls.length,acceptedCalls);
    const query=upstream.calls[1].body;
    for(const field of ['product_id','page','page_size'])assert.throws(()=>verifyStoppedProductBridgeResultAndExtractOriginalBytes(createSyntheticBoundStoppedStorefrontResult({...query,[field]:query[field]+1}),baseline,query,configuration.task_id));
  } finally {bridge?.stop();await upstream.stop();}
});

test('the ninety-seventh valid read is refused without sending another maintenance query', async () => {
  const upstream = await createSyntheticMaintenanceNetworkServerForStoppedBridge();
  let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    for (let index = 0; index < 96; index++) { const response = await fetch(bridge.origin + '/api/v1/storefront/context'); assert.equal(response.status, 200); await response.text(); }
    const refused = await fetch(bridge.origin + '/api/v1/storefront/context'); assert.equal(refused.status, 429); await refused.text();
    assert.equal(upstream.calls.length, 97);
  } finally { bridge?.stop(); await upstream.stop(); }
});
