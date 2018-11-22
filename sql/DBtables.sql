DROP TABLE IF EXISTS Forums, Posts, Votes, Users, Threads CASCADE;

CREATE EXTENSION IF NOT EXISTS citext;


CREATE TABLE IF NOT EXISTS Users		-- Done(+/-)
(
	about citext,
	email citext UNIQUE NOT NULL,
	fullname citext NOT NULL,
	nickname citext PRIMARY KEY
);


CREATE TABLE IF NOT EXISTS Forums		-- Done (+/-)
(
	posts bigint,
	slug citext UNIQUE NOT NULL,
	threads int,
	title citext,
	author citext NOT NULL REFERENCES Users(nickname)
);

CREATE TABLE IF NOT EXISTS Threads		-- Done
(
	author citext NOT NULL REFERENCES Users(nickname),
	created TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	forum citext REFERENCES Forums(slug),
	id SERIAL PRIMARY KEY,
	message citext,
	slug citext UNIQUE,
	title citext UNIQUE NOT NULL,
	votes int
);

CREATE TABLE IF NOT EXISTS Posts		-- Done
(
	author citext NOT NULL REFERENCES Users(nickname),
	created TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	forum citext REFERENCES Forums(slug),
	id SERIAL PRIMARY KEY,
	isEdited boolean NOT NULL DEFAULT FALSE,
	message citext UNIQUE NOT NULL,
	parent bigint,
	thread int NOT NULL REFERENCES Threads(id)
);

CREATE TABLE IF NOT EXISTS Votes 		-- Done
(  
	nickname citext PRIMARY KEY REFERENCES Users(nickname),
	voice int UNIQUE NOT NULL
);
