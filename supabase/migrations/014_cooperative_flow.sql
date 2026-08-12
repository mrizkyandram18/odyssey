-- Epic 2: Cooperative Flow
-- Allow exercises to be assigned to a specific crew member.
-- If assigned_to is NULL, it is an open challenge anyone can take.
ALTER TABLE odyssey_exercises
ADD COLUMN assigned_to TEXT;
