CREATE TABLE users (
    id       VARCHAR(32) PRIMARY KEY NOT NULL,
    name     VARCHAR(255) UNIQUE NOT NULL,
    created  INTEGER
);

create table refresh_tokens (
    id          VARCHAR(64) PRIMARY KEY NOT NULL,
    user_id     VARCHAR(64) NOT NULL REFERENCES users(id),
    created     INTEGER,
    expires     INTEGER
);

create table sessions (
    id          VARCHAR(64) PRIMARY KEY NOT NULL,
    value       TEXT NOT NULL,
    created     INTEGER,
    expires     INTEGER
);
