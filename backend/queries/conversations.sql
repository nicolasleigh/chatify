-- name: GetConversation :many
WITH 
    clerk_users AS (
        SELECT id 
        FROM users 
        WHERE users.clerk_id = $1
    ),
    conv AS (
        SELECT id, name, is_group, last_message_id
        FROM conversations 
        WHERE conversations.id = $2
    ),
    current_user_member AS (
        SELECT last_seen_message_id
        FROM conversation_members
        WHERE member_id = (SELECT id FROM clerk_users)
          AND conversation_id = $2
    ),
    current_user_last_seen_time AS (
        SELECT created_at FROM messages
        WHERE messages.id IN (SELECT last_seen_message_id FROM current_user_member)
    ),
    current_user_unseen_count AS (
        SELECT COUNT(*) as unseen_message_count FROM messages
        WHERE created_at > (SELECT created_at FROM current_user_last_seen_time)
        AND messages.conversation_id = $2
    )
SELECT 
    (SELECT id FROM clerk_users) as current_user_id,
    users.id as other_member_id, 
    users.username as other_member_username, 
    users.email as other_member_email, 
    users.image_url as other_member_image_url, 
    member.last_seen_message_id as other_member_last_seen_message_id, 
    -- (SELECT last_seen_message_id FROM current_user_member) as current_user_last_seen_message_id,
    (SELECT unseen_message_count FROM current_user_unseen_count),
    conv.name as conversation_name, 
    conv.is_group,
    conv.id as conversation_id,
    conv.last_message_id
FROM conversation_members member
JOIN conv ON conv.id = member.conversation_id
JOIN users ON users.id = member.member_id
WHERE member.member_id != (SELECT id FROM clerk_users)
  AND member.conversation_id = $2;

-- name: GetConversationsByClerkId :many
WITH clerk_users AS (
    SELECT id
    FROM users
    WHERE users.clerk_id = $1
)
SELECT member.conversation_id FROM conversation_members member
JOIN clerk_users ON clerk_users.id = member.member_id;

-- name: GetConversationMember :one
SELECT id FROM conversation_members WHERE conversation_id = $1 AND member_id = $2;

-- name: CreateGroup :exec
WITH 
    clerk_users AS (
        SELECT id 
        FROM users 
        WHERE clerk_id = $1
    ),
    conv AS (
        INSERT INTO conversations (
            name, is_group, group_owner
        ) VALUES (
            $2, true, (SELECT id FROM clerk_users)
        )
        RETURNING id
    )
INSERT INTO conversation_members (
    conversation_id, member_id
) 
SELECT conv.id, member_id
FROM conv, unnest($3::bigint[]) as member_id
UNION
SELECT conv.id, clerk_users.id
FROM conv, clerk_users;

-- name: LeaveGroup :exec
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
DELETE FROM conversation_members 
WHERE conversation_members.member_id IN (SELECT id FROM clerk_users)
AND conversation_members.conversation_id = $2;

-- name: DeleteGroup :exec
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
DELETE FROM conversations 
WHERE group_owner IN (SELECT id FROM clerk_users)
AND conversations.id = $2;