UPDATE odyssey_tasks
SET title = REPLACE(title, 'ANTIGRAVITY: ', '')
WHERE title ILIKE 'ANTIGRAVITY:%';
