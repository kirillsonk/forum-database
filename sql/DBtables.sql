DROP TABLE IF EXISTS Forums, Posts, Votes, Users, Threads CASCADE;

CREATE EXTENSION IF NOT EXISTS citext;


CREATE TABLE IF NOT EXISTS Users		-- Done(+)
(
	about citext,
	email citext UNIQUE NOT NULL,
	fullname citext NOT NULL,
	nickname citext PRIMARY KEY
);


CREATE TABLE IF NOT EXISTS Forums		-- Done (+)
(
	posts bigint DEFAULT 0,
	slug citext UNIQUE NOT NULL,
	threads bigint DEFAULT 0,
	title citext,
	author citext PRIMARY KEY REFERENCES Users(nickname)
);

CREATE TABLE IF NOT EXISTS Threads		-- Done
(
	author citext NOT NULL REFERENCES Users(nickname),
	created TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	forum citext REFERENCES Forums(slug),
	id bigserial PRIMARY KEY,
	message citext,
	slug citext UNIQUE,
	title citext NOT NULL,
	votes bigint DEFAULT 0
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
	voice int DEFAULT 0,
	thread int REFERENCES Threads(id),
	UNIQUE (nickname, thread)
);
