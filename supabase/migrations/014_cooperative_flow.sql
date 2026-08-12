-- Epic 2: Cooperative Flow
-- Allow challenges to be assigned to a specific crew member.
-- If assigned_to is NULL, it is an open challenge anyone can take.
ALTER TABLE odyssey_challenges
ADD COLUMN assigned_to TEXT;
