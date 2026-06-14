CREATE TABLE users
(
    id       bigserial NOT NULL PRIMARY KEY,
    login    VARCHAR   NOT NULL,
    password VARCHAR   NOT NULL
);
CREATE UNIQUE INDEX idx_users_login ON users (login);

CREATE TABLE orders
(
    number      VARCHAR                     NOT NULL PRIMARY KEY,
    status      VARCHAR                     NOT NULL,
    accrual     VARCHAR,
    user_id     bigint                      NOT NULL,
    uploaded_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT fk_orders_users
        FOREIGN KEY (user_id)
            REFERENCES users (id)
);