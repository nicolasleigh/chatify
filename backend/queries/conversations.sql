-- name: GetConversation :many
WITH 
    clerk_users AS (
        SELECT id 
        FROM users 
        WHERE users.clerk_id = $1
    ),
    conv AS (
        SELECT id, name, is_group
        FROM conversations 
        WHERE conversations.id = $2
    )
SELECT 
    clerk_users.id as current_user_id,
    users.id as other_member_id, 
    users.username as other_member_username, 
    users.email as other_member_email, 
    users.image_url as other_member_image_url, 
    member.last_message_id as other_member_last_message_id, 
    conv.name as conversation_name, 
    conv.is_group,
    conv.id as conversation_id
FROM conversation_members member
JOIN conv ON conv.id = member.conversation_id
JOIN clerk_users ON clerk_users.id != member.member_id
JOIN users ON users.id = member.member_id;

-- name: GetConversationsByClerkId :many
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
SELECT member.conversation_id FROM conversation_members member
JOIN clerk_users ON clerk_users.id = member.member_id;

-- name: CreateGroup :exec
WITH 
    clerk_users AS (
        SELECT id 
        FROM users 
        WHERE clerk_id = $1
    ),
    conv AS (
        INSERT INTO conversations (
            name, is_group
        ) VALUES (
            $2, true
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