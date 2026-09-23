-- name: CreateRequest :one
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
), receiver AS (
    SELECT id
    FROM users
    WHERE users.email = $2
)
INSERT INTO friend_requests (
    sender_id,
    receiver_id
)
SELECT 
    clerk_users.id,
    receiver.id
FROM clerk_users
CROSS JOIN receiver
WHERE clerk_users.id <> receiver.id
RETURNING id;

-- name: DeleteRequest :one
DELETE FROM friend_requests request
USING users receiver
WHERE request.id = $1
  AND request.receiver_id = receiver.id
  AND receiver.clerk_id = $2
RETURNING request.id, request.sender_id, request.receiver_id, request.created_at;

-- name: AcceptRequest :one
WITH deleted_request AS (
  DELETE FROM friend_requests request
  USING users receiver
  WHERE request.id = $1
    AND request.receiver_id = receiver.id
    AND receiver.clerk_id = $2
  RETURNING request.sender_id, request.receiver_id
),
new_conversation AS (
  INSERT INTO conversations (is_group)
  SELECT false
  FROM deleted_request
  RETURNING id AS conversation_id
),
friend_insert AS (
  INSERT INTO friends (user_a_id, user_b_id, conversation_id)
  SELECT 
    LEAST(request.sender_id, request.receiver_id),
    GREATEST(request.sender_id, request.receiver_id),
    conversation.conversation_id
  FROM deleted_request request
  CROSS JOIN new_conversation conversation
  RETURNING user_a_id, user_b_id, conversation_id
),
member_insert AS (
  INSERT INTO conversation_members (member_id, conversation_id)
  SELECT member_id, conversation_id
  FROM friend_insert
  CROSS JOIN LATERAL (
    VALUES (friend_insert.user_a_id), (friend_insert.user_b_id)
  ) AS members(member_id)
  RETURNING conversation_id
)
SELECT conversation_id
FROM member_insert
LIMIT 1;

-- name: GetFriends :many
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
SELECT users.* 
FROM users 
JOIN friends ON (
    (friends.user_a_id IN (SELECT id FROM clerk_users) AND users.id = friends.user_b_id)
    OR 
    (friends.user_b_id IN (SELECT id FROM clerk_users) AND users.id = friends.user_a_id)
);

-- name: DeleteFriend :exec
WITH clerk_user AS (
    SELECT id
    FROM users
    WHERE clerk_id = $2
), deleted_friend AS (
    DELETE FROM friends
    WHERE conversation_id = $1
      AND (
          user_a_id = (SELECT id FROM clerk_user)
          OR user_b_id = (SELECT id FROM clerk_user)
      )
    RETURNING conversation_id
)
DELETE FROM conversations
WHERE conversations.id IN (SELECT conversation_id FROM deleted_friend);
-- Legacy:
-- WITH deleted_friend AS (
--   DELETE FROM friends 
--   WHERE user_a_id = LEAST($1::bigint, $2::bigint) AND user_b_id = GREATEST($1::bigint, $2::bigint)
--   RETURNING conversation_id
-- )
-- DELETE FROM conversations
-- WHERE id = (SELECT conversation_id FROM deleted_friend);

-- name: GetRequests :many
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
SELECT users.id, users.username, users.image_url, users.email, f.sender_id, f.receiver_id ,COUNT(*) OVER() AS request_count 
FROM friend_requests f
JOIN users ON f.sender_id = users.id
JOIN clerk_users ON f.receiver_id = clerk_users.id;
