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
    "user_id" int NOT NULL,
    "title" varchar(255) NOT NULL,
    "slug" varchar(255) NOT NULL,
    "description" text,
    "body" text,
    "tag_list" varchar(100)[],
    "created_at" timestamp, 
    "updated_at" timestamp,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS "sessions";
CREATE TABLE sessions (
    "session_key" uuid NOT NULL,
    "user_id" int,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE  
);
