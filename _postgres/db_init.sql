DROP TABLE IF EXISTS "users";
CREATE TABLE users (
    "id" serial PRIMARY KEY,
    "email" varchar(100) UNIQUE NOT NULL,
    "username" varchar(100) UNIQUE NOT NULL,
    "password_hashed"  bytea NOT NULL,
    "bio" text,
    "image" varchar(255),
    "created_at" timestamp, 
    "updated_at" timestamp
);

DROP TABLE IF EXISTS "articles";
CREATE TABLE articles (
    "id" serial PRIMARY KEY,
    "user_id" int UNIQUE NOT NULL,
    "body" text,
    "description" text,
    "favorited"  boolean,
    "favorites_count" int ,
    "slug" varchar(255) UNIQUE NOT NULL,
    "title" varchar(255),
    "tag_list" varchar(100)[],
    "created_at" timestamp, 
    "updated_at" timestamp
);

DROP TABLE IF EXISTS "sessions";
CREATE TABLE sessions (
    "session_key" varchar(255) UNIQUE,
    "user_id" int  
);
