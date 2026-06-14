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
    accrual     NUMERIC(15, 2),
    user_id     bigint                      NOT NULL,
    uploaded_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT fk_orders_users
        FOREIGN KEY (user_id)
            REFERENCES users (id)
);

CREATE TABLE transactions (
    id         bigserial      PRIMARY KEY,
    user_id    bigint         NOT NULL,
    amount     NUMERIC(15, 2) NOT NULL,
    order_num  VARCHAR        NOT NULL,
    created_at TIMESTAMP      NOT NULL DEFAULT now(),
    CONSTRAINT fk_transaction_users
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_transaction_orders
        FOREIGN KEY (order_num) REFERENCES orders(number),
    CONSTRAINT uq_transaction_order UNIQUE (order_num)
);
