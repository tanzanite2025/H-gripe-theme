'use strict';
const test = require('node:test');
const assert = require('node:assert/strict');
const http = require('node:http');
const { createHash } = require('node:crypto');
const { startStoppedProductReadOnlyStorefrontLoopbackBridge, mapStoppedProductStorefrontGETToBoundedQuery } = require('./serve-stopped-product-read-only-storefront-loopback-bridge.cjs');
const { verifyStoppedProductPublicImageResultAndExtractOriginalBytes } = require('./stopped-product-shared-public-image-request-mapping-and-bound-response-byte-verification.cjs');

const configuration = { maintenance_origin: 'http://127.0.0.1:12345', maintenance_session: 'synthetic-maintenance-session-value', storefront_origin: 'http://127.0.0.1:3000', task_id: 'a'.repeat(32), contract_sha256: 'b'.repeat(64), locale: 'en', country: 'AQ', currency: 'USD' };
const baseline = { version: 1, scope: 'product_catalog_with_images', task_id: configuration.task_id, contract_sha256: configuration.contract_sha256, window_id: 'c'.repeat(32), database_oid: 20001, detail_cache_progress_sha256: 'd'.repeat(64), application_remains_stopped: true, online_overwrite_available: false };
const pngBytes = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jfL8AAAAASUVORK5CYII=', 'base64');

function createSyntheticStoppedPublicImageResult(query, bytes = pngBytes, status = 200) {
  return { ...baseline, request: query, target: { contract_sha256: query.contract_sha256, operation: 'public_upload', object_key: query.object_key, conversation_id: '', message_id: '', attachment_index: 0, preset: '', range: '', visitor_cookie: '', storefront_auth_cookie: '', admin_auth_cookie: '' }, visitor_credential_supplied: false, storefront_credential_supplied: false, admin_credential_supplied: false, http_status: status, headers: { 'Content-Type': 'image/png', 'Content-Length': String(bytes.length), 'Content-Range': '' }, response_bytes: bytes.length, response_bytes_base64: bytes.toString('base64'), response_sha256: createHash('sha256').update(bytes).digest('hex'), sql_read_only: true, verification: 'registered_shared_public_image_bytes_bound_read_only_sql_and_resealed_before_delivery', remaining_checks: ['nuxt_ipx_and_actual_browser'] };
}

async function startSyntheticStoppedPublicImageMaintenanceUpstream(transform = result => result) {
  const calls = [];
  let active = 0;
  let maximumActive = 0;
  const server = http.createServer(async (request, response) => {
    assert.equal(request.method, 'POST');
    assert.equal(request.headers.authorization, 'Bearer ' + configuration.maintenance_session);
    assert.equal(request.headers.cookie, undefined);
    assert.equal(request.headers.range, undefined);
    active++; maximumActive = Math.max(maximumActive, active);
    const chunks = [];
    for await (const chunk of request) chunks.push(chunk);
    const query = JSON.parse(Buffer.concat(chunks));
    calls.push({ url: request.url, query });
    response.setHeader('Content-Type', 'application/json');
    if (request.url.endsWith('product-catalog-stopped-projection-inspection')) response.end(JSON.stringify({ report: baseline }));
    else {
      assert(request.url.endsWith('product-catalog-stopped-read-only-public-image-http-query'));
      const missing = query.object_key === 'products/missing.png';
      response.end(JSON.stringify({ result: transform(createSyntheticStoppedPublicImageResult(query, missing ? Buffer.alloc(0) : pngBytes, missing ? 404 : 200)) }));
    }
    active--;
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  return { configuration: { ...configuration, maintenance_origin: 'http://127.0.0.1:' + server.address().port }, calls, maximumActive: () => maximumActive, stop: async () => { server.closeAllConnections(); await new Promise(resolve => server.close(resolve)); } };
}

test('shared public image mapping preserves exact keys and refuses selectors, private namespaces and conditional ranges', () => {
  const mapTarget = (url, headers = {}) => mapStoppedProductStorefrontGETToBoundedQuery({ method: 'GET', url, headers }, configuration);
  assert.deepEqual(mapTarget('/uploads/products/photo.png'), { query: { contract_sha256: configuration.contract_sha256, object_key: 'products/photo.png' }, publicImage: true });
  assert.equal(mapTarget('/uploads/products/photo%20one.png').query.object_key, 'products/photo one.png');
  assert.equal(mapTarget('/uploads/products/%E5%9B%BE.png').query.object_key, 'products/图.png');
  assert.equal(mapTarget('/uploads/products/' + encodeURIComponent('图'.repeat(350)) + '.png').status, 400);
  for (const path of ['/uploads/customer-service/attachments/private.png', '/uploads/customer-service/avatars/me.png', '/uploads/showcase/approved/me.png', '/uploads/order-evidence/proof.png', '/uploads/paypal-dispute-invoices/invoice.png', '/uploads/visual-showcase/tile.png', '/uploads/products/a.png?url=https://remote.invalid', '/uploads/products/a.png?']) assert.equal(mapTarget(path).status, 501, path);
  for (const path of ['/uploads/', '/uploads/products//a.png', '/uploads/products/%252fsecret.png', '/uploads/products/%23secret.png', '/uploads/products/%0asecret.png', '/uploads/products/%2Fsecret.png', '/uploads/products/%2e%2e/secret.png', '/uploads/../secret']) assert.equal(mapTarget(path).status, 400, path);
  for (const headers of [{ range: 'bytes=0-7' }, { 'if-range': 'old' }, { 'if-none-match': 'old' }, { 'if-modified-since': 'old' }]) assert.equal(mapTarget('/uploads/products/photo.png', headers).status, 501);
});

test('public image bytes require exact binding, canonical base64, size, digest, safe MIME and credential-free target', () => {
  const query = { contract_sha256: configuration.contract_sha256, object_key: 'products/photo.png' };
  const result = createSyntheticStoppedPublicImageResult(query);
  assert.deepEqual(verifyStoppedProductPublicImageResultAndExtractOriginalBytes(result, baseline, query, configuration.task_id), pngBytes);
  const boundaryBytes = Buffer.concat([pngBytes, Buffer.alloc(4 * 1024 * 1024 - pngBytes.length)]);
  assert.deepEqual(verifyStoppedProductPublicImageResultAndExtractOriginalBytes(createSyntheticStoppedPublicImageResult(query, boundaryBytes), baseline, query, configuration.task_id), boundaryBytes);
  assert.throws(() => verifyStoppedProductPublicImageResultAndExtractOriginalBytes(createSyntheticStoppedPublicImageResult(query, Buffer.concat([boundaryBytes, Buffer.from([0])])), baseline, query, configuration.task_id));
  for (const replacement of [{ version: 2 }, { database_oid: 2 }, { task_id: 'e'.repeat(32) }, { window_id: 'e'.repeat(32) }, { contract_sha256: 'e'.repeat(64) }, { detail_cache_progress_sha256: 'e'.repeat(64) }, { online_overwrite_available: true }, { application_remains_stopped: false }, { sql_read_only: false }, { http_status: 206 }, { response_bytes: pngBytes.length + 1 }, { response_sha256: 'f'.repeat(64) }, { response_bytes_base64: result.response_bytes_base64 + '=' }, { headers: { 'Content-Type': 'image/svg+xml' } }, { headers: { 'Content-Type': 'image/png', Location: 'https://remote.invalid' } }, { headers: { 'Content-Type': 'image/png', 'Set-Cookie': 'private=value' } }, { headers: { 'Content-Type': 'image/png', 'Content-Length': '1' } }, { headers: { 'Content-Type': 'image/png\r\nLocation: bad' } }, { target: { ...result.target, admin_auth_cookie: 'must-never-exist' } }, { request: { ...query, object_key: 'different.png' } }, { visitor_credential_supplied: true }, { matches_package_file: false }, { remaining_checks: [] }]) assert.throws(() => verifyStoppedProductPublicImageResultAndExtractOriginalBytes({ ...result, ...replacement }, baseline, query, configuration.task_id));
  assert.throws(() => verifyStoppedProductPublicImageResultAndExtractOriginalBytes(createSyntheticStoppedPublicImageResult(query, Buffer.from('<html>not image</html>')), baseline, query, configuration.task_id));
});

test('real loopback TCP returns exact image bytes, safe headers and genuine 404 without forwarding business cookies or ranges', async () => {
  const upstream = await startSyntheticStoppedPublicImageMaintenanceUpstream();
  let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    const responses = await Promise.all(['original', 'small', 'unicode'].map(key => fetch(bridge.origin + '/uploads/products/' + key + '.png', { headers: { Cookie: 'admin_auth_token=must-never-forward' } })));
    for (const response of responses) {
      assert.equal(response.status, 200); assert.equal(response.headers.get('content-type'), 'image/png'); assert.equal(response.headers.get('cache-control'), 'no-store'); assert.equal(response.headers.get('x-content-type-options'), 'nosniff'); assert.equal(response.headers.get('set-cookie'), null);
      assert.deepEqual(Buffer.from(await response.arrayBuffer()), pngBytes);
    }
    const missing = await fetch(bridge.origin + '/uploads/products/missing.png'); assert.equal(missing.status, 404); assert.equal((await missing.arrayBuffer()).byteLength, 0);
    const callsBeforeRefusal = upstream.calls.length;
    for (const [path, options, status] of [['/uploads/products/original.png', { headers: { Range: 'bytes=0-7' } }, 501], ['/uploads/visual-showcase/current.png', {}, 501], ['/uploads/customer-service/attachments/private.png', {}, 501], ['/uploads/products/original.png', { method: 'POST' }, 405]]) { const response = await fetch(bridge.origin + path, options); assert.equal(response.status, status); await response.text(); }
    assert.equal(upstream.calls.length, callsBeforeRefusal);
    assert.equal(upstream.maximumActive(), 1);
    assert.equal(bridge.closed(), false);
  } finally { bridge?.stop(); await upstream.stop(); }
});

for (const [name, transform] of [
  ['wrong_database', result => ({ ...result, database_oid: 7 })],
  ['wrong_digest', result => ({ ...result, response_sha256: 'e'.repeat(64) })],
  ['private_cookie', result => ({ ...result, target: { ...result.target, visitor_cookie: 'forged' } })],
  ['html_image', result => ({ ...result, headers: { ...result.headers, 'Content-Type': 'text/html' } })],
]) test('invalid public image ' + name + ' closes the listener and rejects queued requests', async () => {
  const upstream = await startSyntheticStoppedPublicImageMaintenanceUpstream(transform);
  let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    const responses = await Promise.allSettled([fetch(bridge.origin + '/uploads/products/a.png'), fetch(bridge.origin + '/uploads/products/b.png')]);
    for (const response of responses) if (response.status === 'fulfilled') { assert.equal(response.value.status, 503); await response.value.text(); }
    assert.equal(bridge.closed(), true);
    assert.equal(upstream.calls.length, 2);
    await assert.rejects(fetch(bridge.origin + '/uploads/products/c.png'));
  } finally { bridge?.stop(); await upstream.stop(); }
});
