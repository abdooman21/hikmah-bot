-- name: GetRandomQuestionByCriteria :one
SELECT 
    q.id, 
    q.q_text, 
    q.answers, 
    q.link, 
    q.level_number,
    t.name AS topic_name, 
    c.arabicName AS category_name
FROM Questions q
JOIN Topics t ON q.topic_id = t.id
JOIN MainCatagories c ON t.category_id = c.id
WHERE c.id = $1 
  AND t.id = $2
  AND q.level_number = $3
ORDER BY RANDOM()
LIMIT 1;


-- name: GetRandomQByCatnLvl :one
SELECT 
    q.id,
    q.q_text,
    q.answers,
    q.link,
    t.name AS topic_name,
    c.englishname AS category_ename,
    c.arabicName AS category_name,
    c.icon_path
FROM questions q
JOIN topics t ON q.topic_id = t.id
JOIN maincatagories c ON t.category_id = c.id
WHERE c.id = $1 
  AND q.level_number = $2
ORDER BY RANDOM()
LIMIT 1;