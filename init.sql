CREATE SCHEMA IF NOT EXISTS wb;

CREATE TABLE IF NOT EXISTS wb.orders (
    id varchar(500) PRIMARY KEY,
    data jsonb
);