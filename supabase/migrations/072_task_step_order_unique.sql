-- ============================================================
-- Migration 072: Deterministic task ordering + unique step per family/date
-- Fixes 1,1,2 bug. Preserves task IDs, submissions, ledger.
-- Preserves non-conflicting step_order values, only fixes duplicates.
-- ============================================================

-- 1. Fix duplicate active (family_id, active_date, step_order)
-- Preserve ordering step_order ASC, id ASC
-- Example 1,1,2 → 1,2,3 ; 1,3,3,5 → 1,3,4,5 (keeps 1,3,5)
DO $$
DECLARE
  grp RECORD;
  rec RECORD;
  used INT[];
  candidate INT;
  is_dup BOOLEAN;
BEGIN
  FOR grp IN
    SELECT DISTINCT family_id, active_date
    FROM odyssey_tasks
    WHERE is_active = TRUE
    GROUP BY family_id, active_date, step_order
    HAVING count(*) > 1
  LOOP
    used := ARRAY[]::INT[];
    FOR rec IN
      SELECT id, step_order
      FROM odyssey_tasks
      WHERE family_id = grp.family_id
        AND active_date = grp.active_date
        AND is_active = TRUE
      ORDER BY step_order ASC, id ASC
    LOOP
      is_dup := rec.step_order = ANY(used);
      IF is_dup THEN
        candidate := rec.step_order + 1;
        WHILE candidate = ANY(used) LOOP
          candidate := candidate + 1;
        END LOOP;
        -- Also avoid colliding with any original step that will be kept
        -- If candidate exists as original step for a later row not yet in used,
        -- it will be caught when that later row is processed (it will become dup)
        -- So just assign candidate
        UPDATE odyssey_tasks SET step_order = candidate + 1000000 WHERE id = rec.id;
        used := array_append(used, candidate);
      ELSE
        used := array_append(used, rec.step_order);
      END IF;
    END LOOP;
  END LOOP;

  -- Phase 2: bring temp values down to final (remove offset)
  UPDATE odyssey_tasks SET step_order = step_order - 1000000 WHERE step_order > 1000000 AND is_active = TRUE;
END $$;

-- 2. Create unique partial index for active tasks (Source of Truth)
CREATE UNIQUE INDEX IF NOT EXISTS uq_tasks_family_date_step_active
ON odyssey_tasks (family_id, active_date, step_order)
WHERE is_active = TRUE;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','072_task_step_order_unique')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
