-- name: GetMainCategories :many
SELECT
    id,
    arabicName,
    englishName
FROM MainCatagories
ORDER BY id;

-- name: GetCatagoriesTopic :many

SELECT * from Topics 
ORDER BY id;