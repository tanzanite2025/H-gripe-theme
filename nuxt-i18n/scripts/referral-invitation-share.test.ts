import assert from 'node:assert/strict'
import { shareReferralInvitationURL } from '../app/utils/referralInvitationShare.js'
import { buildStorefrontRouteRules } from '../config/storefront/route-rules.js'

const invitationURL = 'https://shop.test/r/ABCD2345'
const calls: unknown[] = []
assert.equal(await shareReferralInvitationURL(invitationURL, {
  share: async (payload: { url: string }) => { calls.push(payload) },
  copy: async () => { throw new Error('Native sharing succeeded; clipboard must not run') },
}), 'shared')
assert.deepEqual(calls, [{ url: invitationURL }])

const copied: string[] = []
const copy = async (url: string) => { copied.push(url); return true }
assert.equal(await shareReferralInvitationURL(invitationURL, { copy }), 'copied')
assert.equal(await shareReferralInvitationURL(invitationURL, {
  share: async () => { throw Object.assign(new Error('cancelled'), { name: 'AbortError' }) }, copy,
}), 'cancelled')
assert.deepEqual(copied, [invitationURL])
assert.equal(await shareReferralInvitationURL(invitationURL, {
  share: async () => { throw new Error('not supported') }, copy,
}), 'copied')
assert.equal(await shareReferralInvitationURL(invitationURL, { copy: async () => false }), 'unavailable')
assert.equal(await shareReferralInvitationURL(invitationURL, { copy: async () => { throw new Error('denied') } }), 'unavailable')
assert.equal(await shareReferralInvitationURL('https://short.test/x/123', { copy }), 'copied')
for (const url of ['', '/products/item', 'javascript:alert(1)', 'https://user:secret@shop.test']) {
  assert.equal(await shareReferralInvitationURL(url, {
    copy: async () => { throw new Error('Invalid URLs must not reach clipboard') },
  }), 'unavailable')
}
const rules = buildStorefrontRouteRules({ internalApiOrigin: 'http://localhost:9200/', localeCodes: ['en'], defaultLocale: 'en' })
assert.equal(rules['/r/**']?.proxy, 'http://localhost:9200/r/**')
assert.equal(rules['/r/**']?.headers?.['cache-control'], 'no-store, max-age=0')
console.log('Referral invitation sharing checks passed: URL-only sharing, copy fallback, cancellation, failures, short URLs, capture proxy.')
