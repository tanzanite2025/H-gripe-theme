'use strict';
const test = require('node:test');
const assert = require('node:assert/strict');
const http = require('node:http');
const { createHash } = require('node:crypto');
const { startStoppedProductReadOnlyStorefrontLoopbackBridge, mapStoppedProductStorefrontGETToBoundedQuery } = require('./serve-stopped-product-read-only-storefront-loopback-bridge.cjs');
const { verifyStoppedProductPublicImageResultAndExtractOriginalBytes } = require('./stopped-product-shared-public-image-request-mapping-and-bound-response-byte-verification.cjs');

const configuration = { maintenance_origin: 'http://127.0.0.1:12345', maintenance_session: 'synthetic-site-identity-maintenance-session', storefront_origin: 'http://127.0.0.1:3000', task_id: 'a'.repeat(32), contract_sha256: 'b'.repeat(64), locale: 'en', country: 'AQ', currency: 'USD' };
const baseline = { version: 1, scope: 'product_catalog_with_images', task_id: configuration.task_id, contract_sha256: configuration.contract_sha256, window_id: 'c'.repeat(32), database_oid: 20001, detail_cache_progress_sha256: 'd'.repeat(64), application_remains_stopped: true, online_overwrite_available: false };
const pngBytes = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jfL8AAAAASUVORK5CYII=', 'base64');

function createSyntheticStoppedSiteIdentityOwnerImageResult(query, status = 200) {
  const body = status === 200 ? pngBytes : Buffer.alloc(0);
  return { ...baseline, request: query, target: { ...query, operation: 'public_upload', conversation_id: '', message_id: '', attachment_index: 0, preset: '', range: '', visitor_cookie: '', storefront_auth_cookie: '', admin_auth_cookie: '' }, visitor_credential_supplied: false, storefront_credential_supplied: false, admin_credential_supplied: false, http_status: status, object_owner: query.object_key.startsWith('site-logo/') ? 'site_logo' : 'site_favicon', headers: { 'Content-Type': 'image/png', 'Content-Length': String(body.length), 'Content-Range': '' }, response_bytes: body.length, response_bytes_base64: body.toString('base64'), response_sha256: createHash('sha256').update(body).digest('hex'), sql_read_only: true, verification: 'site_identity_owner_authorized_image_bytes_bound_read_only_sql_and_resealed_before_delivery', remaining_checks: ['other_owner_media_and_full_pages'] };
}

test('site identity image mapping retains exact keys and requires the corresponding owner verification contract', () => {
  for (const namespace of ['site-logo', 'site-favicon']) {
    const mapped = mapStoppedProductStorefrontGETToBoundedQuery({ method: 'GET', url: '/uploads/' + namespace + '/current.png', headers: {} }, configuration);
    assert.equal(mapped.publicImage, true);
    const result = createSyntheticStoppedSiteIdentityOwnerImageResult(mapped.query);
    assert.deepEqual(verifyStoppedProductPublicImageResultAndExtractOriginalBytes(result, baseline, mapped.query, configuration.task_id), pngBytes);
    for (const replacement of [{ object_owner: namespace === 'site-logo' ? 'site_favicon' : 'site_logo' }, { object_owner: undefined }, { verification: 'registered_shared_public_image_bytes_bound_read_only_sql_and_resealed_before_delivery' }]) assert.throws(() => verifyStoppedProductPublicImageResultAndExtractOriginalBytes({ ...result, ...replacement }, baseline, mapped.query, configuration.task_id));
  }
});

async function startSyntheticStoppedSiteIdentityMaintenanceUpstream(transform = result => result) {
  const calls = [];
  const server = http.createServer(async (request, response) => {
    assert.equal(request.headers.authorization, 'Bearer ' + configuration.maintenance_session);
    assert.equal(request.headers.cookie, undefined);
    const chunks = [];
    for await (const chunk of request) chunks.push(chunk);
    const query = JSON.parse(Buffer.concat(chunks)); calls.push(query);
    response.setHeader('Content-Type', 'application/json');
    if (request.url.endsWith('product-catalog-stopped-projection-inspection')) response.end(JSON.stringify({ report: baseline }));
    else response.end(JSON.stringify({ result: transform(createSyntheticStoppedSiteIdentityOwnerImageResult(query, query.object_key.includes('revoked') ? 404 : 200)) }));
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  return { configuration: { ...configuration, maintenance_origin: 'http://127.0.0.1:' + server.address().port }, calls, stop: async () => { server.closeAllConnections(); await new Promise(resolve => server.close(resolve)); } };
}

test('site logo and favicon original bytes and owner denials pass through actual loopback TCP without business cookies', async () => {
  const upstream = await startSyntheticStoppedSiteIdentityMaintenanceUpstream(); let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    for (const namespace of ['site-logo', 'site-favicon']) {
      const response = await fetch(bridge.origin + '/uploads/' + namespace + '/current.png', { headers: { Cookie: 'admin_auth_token=must-never-forward' } });
      assert.equal(response.status, 200); assert.equal(response.headers.get('content-type'), 'image/png'); assert.equal(response.headers.get('cache-control'), 'no-store');
      assert.deepEqual(Buffer.from(await response.arrayBuffer()), pngBytes);
      const revoked = await fetch(bridge.origin + '/uploads/' + namespace + '/revoked.png'); assert.equal(revoked.status, 404); await revoked.arrayBuffer();
    }
    assert.equal(upstream.calls.length, 5); assert.equal(bridge.closed(), false);
  } finally { bridge?.stop(); await upstream.stop(); }
});

test('a crossed site identity owner closes the bridge before queued or subsequent image reads', async () => {
  const upstream = await startSyntheticStoppedSiteIdentityMaintenanceUpstream(result => ({ ...result, object_owner: 'site_favicon' })); let bridge;
  try {
    bridge = await startStoppedProductReadOnlyStorefrontLoopbackBridge(upstream.configuration);
    const responses = await Promise.allSettled([fetch(bridge.origin + '/uploads/site-logo/current.png'), fetch(bridge.origin + '/uploads/site-favicon/current.png')]);
    for (const response of responses) if (response.status === 'fulfilled') { assert.equal(response.value.status, 503); await response.value.text(); }
    assert.equal(upstream.calls.length, 2); assert.equal(bridge.closed(), true);
    await assert.rejects(fetch(bridge.origin + '/uploads/site-logo/another.png'));
  } finally { bridge?.stop(); await upstream.stop(); }
});
