-- 033: Retire the old generic /api/v1/chat/messages fact source.
--
-- Public Chat now uses the dedicated customer-service ticket path:
--   tickets.category = 'customer_service'
--   ticket_messages
--   visitor_profiles for anonymous/customer context
--
-- The legacy chat_messages/chat_sessions tables were only tied to the removed
-- generic /api/v1/chat/messages route and would create a second message source.
-- The project is not in production, so no compatibility/data migration is kept.

DROP TABLE IF EXISTS chat_sessions CASCADE;
DROP TABLE IF EXISTS chat_messages CASCADE;
DROP FUNCTION IF EXISTS update_chat_tables_updated_at();
