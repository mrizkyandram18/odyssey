UPDATE odyssey_tasks
SET title = REGEXP_REPLACE(
  REPLACE(REPLACE(title, 'Smoke Test ', ''), 'Smoke ', ''),
  ' [0-9]{13}$', ''
)
WHERE title ILIKE 'Smoke%';
