'use strict';
const { createHash } = require('node:crypto');

const publicImageVerificationName = 'registered_shared_public_image_bytes_bound_read_only_sql_and_resealed_before_delivery';
const siteIdentityImageVerificationName = 'site_identity_owner_authorized_image_bytes_bound_read_only_sql_and_resealed_before_delivery';
const publicImageMaximumBytes = 4 * 1024 * 1024;

function mapStoppedProductSharedPublicImageGETToBoundedQuery(request, target, configuration) {
  if (target.search || request.url.includes('?') || request.headers.range || request.headers['if-range'] || request.headers['if-none-match'] || request.headers['if-modified-since']) return { status: 501 };
  let objectKey;
  try { objectKey = decodeURIComponent(target.pathname.slice('/uploads/'.length)); } catch { return { status: 400 }; }
  if (!objectKey || Buffer.byteLength(objectKey, 'utf8') > 1024 || objectKey.trim() !== objectKey || /[%?#\\\x00-\x1f\x7f]/.test(objectKey) || objectKey.split('/').some(segment => !segment || segment === '.' || segment === '..')) return { status: 400 };
  if (/^(?:visual-showcase|customer-service|showcase|order-evidence|suggestion-feedback|paypal-dispute-invoices)(?:\/|$)/.test(objectKey)) return { status: 501 };
  return { query: { contract_sha256: configuration.contract_sha256, object_key: objectKey }, publicImage: true };
}

function verifyStoppedProductPublicImageResultAndExtractOriginalBytes(result, baseline, query, taskID) {
  const owner = /^site-logo(?:\/|$)/.test(query.object_key) ? 'site_logo' : /^site-favicon(?:\/|$)/.test(query.object_key) ? 'site_favicon' : undefined;
  const expectedVerification = owner ? siteIdentityImageVerificationName : publicImageVerificationName;
  if (!result || result.version !== 1 || result.scope !== 'product_catalog_with_images' || result.task_id !== taskID || result.window_id !== baseline.window_id || result.contract_sha256 !== baseline.contract_sha256 || result.database_oid !== baseline.database_oid || result.detail_cache_progress_sha256 !== baseline.detail_cache_progress_sha256 || result.sql_read_only !== true || result.application_remains_stopped !== true || result.online_overwrite_available !== false || result.verification !== expectedVerification || result.object_owner !== owner || !Array.isArray(result.remaining_checks) || !result.remaining_checks.length || !result.request || Object.keys(result.request).length !== 2 || Object.entries(query).some(([key, value]) => result.request[key] !== value) || ![200, 404].includes(result.http_status) || typeof result.response_bytes_base64 !== 'string' || result.response_bytes_base64.length > 5592408 || /[^A-Za-z0-9+/=]/.test(result.response_bytes_base64) || result.visitor_credential_supplied !== false || result.storefront_credential_supplied !== false || result.admin_credential_supplied !== false || !result.headers || typeof result.headers !== 'object' || Array.isArray(result.headers)) throw new Error('Stopped public image result binding refused');
  const expectedTarget = { contract_sha256: query.contract_sha256, operation: 'public_upload', object_key: query.object_key, conversation_id: '', message_id: '', attachment_index: 0, preset: '', range: '', visitor_cookie: '', storefront_auth_cookie: '', admin_auth_cookie: '' };
  if (!result.target || Object.keys(result.target).length !== Object.keys(expectedTarget).length || Object.entries(expectedTarget).some(([key, value]) => result.target[key] !== value)) throw new Error('Stopped public image target refused');
  const allowedHeaders = new Set(['Content-Type', 'Content-Length', 'Content-Range', 'Accept-Ranges', 'Cache-Control', 'Vary', 'X-Content-Type-Options', 'Referrer-Policy']);
  if (Object.entries(result.headers).some(([name, value]) => !allowedHeaders.has(name) || typeof value !== 'string' || /[\r\n]/.test(value)) || result.headers['Content-Range']) throw new Error('Stopped public image headers refused');
  const bytes = Buffer.from(result.response_bytes_base64, 'base64');
  if (bytes.length > publicImageMaximumBytes || !Number.isSafeInteger(result.response_bytes) || result.response_bytes !== bytes.length || bytes.toString('base64') !== result.response_bytes_base64 || createHash('sha256').update(bytes).digest('hex') !== result.response_sha256 || result.headers['Content-Length'] && result.headers['Content-Length'] !== String(bytes.length)) throw new Error('Stopped public image integrity refused');
  if (result.matches_package_file !== undefined && result.matches_package_file !== true) throw new Error('Stopped public image package identity refused');
  if (result.http_status === 200) {
    const mimeType = result.headers['Content-Type'];
    const signatureMatches = mimeType === 'image/png' && bytes.subarray(0, 8).equals(Buffer.from('89504e470d0a1a0a', 'hex')) || mimeType === 'image/jpeg' && bytes.subarray(0, 3).equals(Buffer.from('ffd8ff', 'hex')) || mimeType === 'image/gif' && ['GIF87a', 'GIF89a'].includes(bytes.subarray(0, 6).toString('ascii')) || mimeType === 'image/webp' && bytes.subarray(0, 4).toString('ascii') === 'RIFF' && bytes.subarray(8, 12).toString('ascii') === 'WEBP';
    if (!bytes.length || !signatureMatches) throw new Error('Stopped public image MIME or signature refused');
  }
  return bytes;
}

module.exports = { mapStoppedProductSharedPublicImageGETToBoundedQuery, verifyStoppedProductPublicImageResultAndExtractOriginalBytes };
