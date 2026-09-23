-- name: GetMessages :many
SELECT u.id as user_id, u.username, u.image_url, u.email, m.id as message_id, m.conversation_id as conversation_id, m.type, m.content, m.created_at FROM messages m
JOIN users u ON u.id = m.sender_id
WHERE conversation_id = $1
ORDER BY m.created_at DESC;

-- name: CreateMessage :one
WITH messages_id AS (
    INSERT INTO messages (
        sender_id, conversation_id, type, content
    ) VALUES (
        $1, $2, $3, $4
    )
    RETURNING id
)
UPDATE conversations
SET last_message_id = (SELECT id FROM messages_id)
WHERE conversations.id = $2
RETURNING (SELECT id FROM messages_id) as message_id;

-- name: GetMessageById :one
SELECT u.id as user_id, u.username, u.image_url, u.email, m.id as message_id, m.conversation_id as conversation_id, m.type, m.content, m.created_at FROM messages m
JOIN users u ON u.id = m.sender_id
WHERE m.id = $1;

-- name: MarkReadMessage :exec
UPDATE conversation_members 
SET last_seen_message_id = $3
WHERE conversation_id = $1 AND member_id = $2;

-- name: GetConversationLastMessage :one
SELECT sender_id, users.username as sender_username, users.image_url as sender_image_url, content, type 
FROM messages
JOIN users ON users.id = sender_id
JOIN conversation_members ON conversation_members.conversation_id = messages.conversation_id
JOIN users viewer ON viewer.id = conversation_members.member_id
WHERE messages.id = $1
  AND viewer.clerk_id = $2;

-- name: GetAllUnseenMessageCount :many
WITH 
    clerk_users AS (
        SELECT id 
        FROM users 
        WHERE clerk_id = $1
    ),
    conv_member AS (
        SELECT conversation_id, last_seen_message_id
        FROM conversation_members 
        WHERE member_id IN (SELECT id FROM clerk_users)
    ),
    current_user_last_seen_time AS (
        SELECT conversation_id, created_at AS last_seen_time
        FROM messages
        WHERE id IN (SELECT last_seen_message_id FROM conv_member)
    )
SELECT 
    COUNT(*) AS unseen_message_count, 
    messages.conversation_id 
FROM messages
JOIN current_user_last_seen_time
    ON messages.conversation_id = current_user_last_seen_time.conversation_id
WHERE messages.created_at > current_user_last_seen_time.last_seen_time
AND messages.conversation_id IN (SELECT conversation_id FROM conv_member)
GROUP BY messages.conversation_id;
