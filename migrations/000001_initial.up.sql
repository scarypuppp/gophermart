CREATE TABLE users
(
    id       bigserial NOT NULL PRIMARY KEY,
    login    VARCHAR   NOT NULL,
    password VARCHAR   NOT NULL
);

CREATE UNIQUE INDEX idx_users_login ON users (login);