CREATE UNIQUE INDEX IF NOT EXISTS idx_conversation_members_unique
ON conversation_members (conversation_id, member_id);
