-- name: GetRadios :many
SELECT * FROM radios ORDER BY id;

-- name: GetRadioByID :one
SELECT * FROM radios WHERE id = $1;
