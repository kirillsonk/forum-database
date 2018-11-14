DROP TABLE IF EXISTS Forums, Posts, Votes, Users, Threads CASCADE;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS Forums
(
	posts bigint,
	slug citext UNIQUE NOT NULL,
	threads int,
	title citext PRIMARY KEY,
	author citext UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS Posts
(
	author citext UNIQUE NOT NULL,
	created TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	forum citext,
	id bigint PRIMARY KEY,
	isEdited boolean NOT NULL,
	message citext UNIQUE NOT NULL,
	parent bigint,
	thread int
);

CREATE TABLE IF NOT EXISTS Votes (
	nickname citext PRIMARY KEY,
	voice int UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS Users (
	about citext,
	email citext UNIQUE NOT NULL,
	fullname citext NOT NULL,
	nickname citext PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS Threads (
	author citext UNIQUE NOT NULL,
	created TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	forum citext UNIQUE,
	id SERIAL PRIMARY KEY,
	message citext NOT NULL,
	slug citext UNIQUE,
	title citext UNIQUE NOT NULL,
	votes int
);
