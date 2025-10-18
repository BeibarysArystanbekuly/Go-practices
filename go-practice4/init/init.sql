DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name    TEXT    NOT NULL,
    email   TEXT    NOT NULL UNIQUE,
    balance NUMERIC(12,2) NOT NULL DEFAULT 0
);
INSERT INTO users (name, email, balance) VALUES
('Alice', 'alice@example.com', 1000.00),
('Bob',   'bob@example.com',     50.00),
('Eve',   'eve@example.com',    500.00);
