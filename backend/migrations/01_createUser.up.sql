CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
	id bigserial PRIMARY KEY,
	username TEXT NOT NULL,
	email citext UNIQUE NOT NULL,
	clerk_id TEXT UNIQUE NOT NULL,
	image_url TEXT);
