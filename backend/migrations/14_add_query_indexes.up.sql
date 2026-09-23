CREATE INDEX IF NOT EXISTS idx_conversation_members_member_conversation
ON conversation_members (member_id, conversation_id);

CREATE INDEX IF NOT EXISTS idx_conversation_members_conversation_member
ON conversation_members (conversation_id, member_id);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created_id
ON messages (conversation_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver_created
ON friend_requests (receiver_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_friends_conversation
ON friends (conversation_id);
