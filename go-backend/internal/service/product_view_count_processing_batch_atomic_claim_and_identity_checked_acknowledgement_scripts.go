package service

import "github.com/redis/go-redis/v9"

const productViewCountProcessingBatchIdentityRedisKey = productViewCountRedisKey + ":processing_batch_id"

const productViewCountRedisTypeAndPersistenceValidationLua = `
local function validate_key(key, expected)
    local kind = redis.call('TYPE', key).ok
    if kind ~= 'none' and kind ~= expected then error('product_view_count_unexpected_redis_key_type') end
    if kind ~= 'none' and redis.call('PTTL', key) >= 0 then error('product_view_count_expiring_state_requires_reconciliation') end
    return kind ~= 'none'
end
local active = validate_key(KEYS[1], 'hash')
local processing = validate_key(KEYS[2], 'hash')
local identity = validate_key(KEYS[3], 'string')
if processing and not identity then error('legacy_product_view_processing_batch_requires_manual_reconciliation') end
if identity and not processing then error('product_view_processing_batch_identity_without_payload_requires_reconciliation') end
`

var productViewCountClaimImmutableProcessingBatchScript = redis.NewScript(productViewCountRedisTypeAndPersistenceValidationLua + `
if not processing then
    if not active then return {'', {}} end
    redis.call('SET', KEYS[3], ARGV[1])
    redis.call('RENAME', KEYS[1], KEYS[2])
end
local entries = redis.call('HSCAN', KEYS[2], '0', 'COUNT', ARGV[2])[2]
return {redis.call('GET', KEYS[3]), entries}
`)

const productViewCountAcknowledgeCommittedBatchLua = productViewCountRedisTypeAndPersistenceValidationLua + `
if not processing or redis.call('GET', KEYS[3]) ~= ARGV[1] then return 0 end
-- Check every field before deleting any, since Redis scripts do not roll back
-- writes if a later command errors. Another flusher may have acknowledged it.
for index = 2, #ARGV, 2 do
    local actual = redis.call('HGET', KEYS[2], ARGV[index])
    if actual and actual ~= ARGV[index+1] then error('product_view_processing_payload_changed_before_acknowledgement') end
end
for index = 2, #ARGV, 2 do redis.call('HDEL', KEYS[2], ARGV[index]) end
if redis.call('HLEN', KEYS[2]) == 0 then redis.call('DEL', KEYS[2], KEYS[3]) end
return 1
`

var productViewCountAcknowledgeOnlyMatchingCommittedBatchScript = redis.NewScript(productViewCountAcknowledgeCommittedBatchLua)

var productViewCountReadOnlyRecoverySnapshotScript = redis.NewScript(productViewCountRedisTypeAndPersistenceValidationLua + `
if redis.call('HLEN', KEYS[1]) > tonumber(ARGV[1]) or redis.call('HLEN', KEYS[2]) > tonumber(ARGV[1]) then error('product_view_recovery_snapshot_exceeds_bounded_inspection_capacity') end
return {redis.call('GET', KEYS[3]) or '', redis.call('HGETALL', KEYS[1]), redis.call('HGETALL', KEYS[2])}
`)

func productViewCountActiveProcessingAndBatchIdentityRedisKeys() []string {
	return []string{productViewCountRedisKey, productViewCountProcessingRedisKey, productViewCountProcessingBatchIdentityRedisKey}
}
