// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: friends.sql

package store

import (
	"context"
)

const acceptRequest = `-- name: AcceptRequest :exec
WITH new_conversation AS (
  -- First, create the conversation
  INSERT INTO conversations (is_group)
  VALUES (false)
  RETURNING id AS conversation_id
),
friend_insert AS (
  -- Insert friendship with ordered user IDs
  INSERT INTO friends (user_a_id, user_b_id, conversation_id)
  SELECT 
    LEAST($1::bigint, $2::bigint),  -- Ensure user_a_id < user_b_id
    GREATEST($1::bigint, $2::bigint),
    conversation_id
  FROM new_conversation
  RETURNING conversation_id
)
INSERT INTO conversation_members (member_id, conversation_id)
SELECT user_id, conversation_id
FROM friend_insert
CROSS JOIN (VALUES (LEAST($1::bigint, $2::bigint)), (GREATEST($1::bigint, $2::bigint))) AS users(user_id)
`

type AcceptRequestParams struct {
	Column1 int64 `json:"column_1"`
	Column2 int64 `json:"column_2"`
}

// Insert both users into conversation_members
func (q *Queries) AcceptRequest(ctx context.Context, arg AcceptRequestParams) error {
	_, err := q.db.Exec(ctx, acceptRequest, arg.Column1, arg.Column2)
	return err
}

const createRequest = `-- name: CreateRequest :exec
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
INSERT INTO friend_requests (
    sender_id,
    receiver_id
)
SELECT 
    clerk_users.id,
    (SELECT id FROM users WHERE users.email = $2)
FROM clerk_users
`

type CreateRequestParams struct {
	ClerkID string `json:"clerk_id" validate:"required"`
	Email   string `json:"email" validate:"required,email,max=255"`
}

func (q *Queries) CreateRequest(ctx context.Context, arg CreateRequestParams) error {
	_, err := q.db.Exec(ctx, createRequest, arg.ClerkID, arg.Email)
	return err
}

const deleteFriend = `-- name: DeleteFriend :exec
WITH deleted_friends AS (
    DELETE FROM friends
    WHERE conversation_id = $1
    RETURNING id, user_a_id, user_b_id, created_at, conversation_id
)
DELETE FROM conversations
WHERE conversations.id = $1
`

func (q *Queries) DeleteFriend(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteFriend, id)
	return err
}

const deleteRequest = `-- name: DeleteRequest :one
DELETE FROM friend_requests 
WHERE sender_id = $1
RETURNING id, sender_id, receiver_id, created_at
`

func (q *Queries) DeleteRequest(ctx context.Context, senderID int64) (FriendRequest, error) {
	row := q.db.QueryRow(ctx, deleteRequest, senderID)
	var i FriendRequest
	err := row.Scan(
		&i.ID,
		&i.SenderID,
		&i.ReceiverID,
		&i.CreatedAt,
	)
	return i, err
}

const getFriends = `-- name: GetFriends :many
WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
SELECT users.id, users.username, users.email, users.clerk_id, users.image_url, users.created_at 
FROM users 
JOIN friends ON (
    (friends.user_a_id IN (SELECT id FROM clerk_users) AND users.id = friends.user_b_id)
    OR 
    (friends.user_b_id IN (SELECT id FROM clerk_users) AND users.id = friends.user_a_id)
)
`

func (q *Queries) GetFriends(ctx context.Context, clerkID string) ([]User, error) {
	rows, err := q.db.Query(ctx, getFriends, clerkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.ClerkID,
			&i.ImageUrl,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRequests = `-- name: GetRequests :many

WITH clerk_users AS (
    SELECT id 
    FROM users 
    WHERE users.clerk_id = $1
)
SELECT users.id, users.username, users.image_url, users.email, f.sender_id, f.receiver_id ,COUNT(*) OVER() AS request_count 
FROM friend_requests f
JOIN users ON f.sender_id = users.id
JOIN clerk_users ON f.receiver_id = clerk_users.id
`

type GetRequestsRow struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username" validate:"required,min=1,max=100"`
	ImageUrl     *string `json:"image_url" validate:"required,url"`
	Email        string  `json:"email" validate:"required,email,max=255"`
	SenderID     int64   `json:"sender_id" validate:"required"`
	ReceiverID   int64   `json:"receiver_id"`
	RequestCount int64   `json:"request_count"`
}

// Legacy:
// WITH deleted_friend AS (
//
//	DELETE FROM friends
//	WHERE user_a_id = LEAST($1::bigint, $2::bigint) AND user_b_id = GREATEST($1::bigint, $2::bigint)
//	RETURNING conversation_id
//
// )
// DELETE FROM conversations
// WHERE id = (SELECT conversation_id FROM deleted_friend);
func (q *Queries) GetRequests(ctx context.Context, clerkID string) ([]GetRequestsRow, error) {
	rows, err := q.db.Query(ctx, getRequests, clerkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRequestsRow
	for rows.Next() {
		var i GetRequestsRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.ImageUrl,
			&i.Email,
			&i.SenderID,
			&i.ReceiverID,
			&i.RequestCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
