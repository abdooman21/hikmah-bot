-- +goose Up
CREATE TABLE radios (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    link TEXT NOT NULL
);

-- +goose Down
DROP TABLE radios;
