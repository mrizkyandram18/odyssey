-- ============================================================
-- Migration 078: Level-based Cap Bonus + Reward Tickets + Cosmetics
-- Additive only. No existing tables/columns altered (except CREATE OR REPLACE
-- of odyssey_get_effective_earning_cap, extended with an opt-in level bonus).
-- Default bonus = 0 when unconfigured, so production behavior is unchanged
-- until level_cap_bonus is set in odyssey_system_config.
-- Vocabulary note: user-facing copy must say Tiket Hadiah / Buka Hadiah /
-- Koleksi (never gacha/pull/inventory in UI).
-- ============================================================

-- 1. Cosmetic pool (MVP = variants already rendered by Avatar.tsx)
CREATE TABLE IF NOT EXISTS cosmetic_items (
    id         TEXT PRIMARY KEY,
    slot       TEXT NOT NULL CHECK (slot IN ('frame', 'effect')),
    asset      TEXT NOT NULL,
    tier       INTEGER NOT NULL DEFAULT 1 CHECK (tier >= 1),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

INSERT INTO cosmetic_items (id, slot, asset, tier, is_active)
VALUES
    ('frame-gold', 'frame', 'gold', 1, TRUE),
    ('effect-sparkle', 'effect', 'sparkle', 1, TRUE),
    ('effect-float', 'effect', 'float', 1, TRUE),
    ('effect-trail', 'effect', 'trail', 1, TRUE)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE cosmetic_items ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on cosmetic_items" ON cosmetic_items;
CREATE POLICY "Allow service_role full access on cosmetic_items" ON cosmetic_items FOR ALL TO service_role USING (true);

-- 2. Reward tickets: SINGLE source of truth (no balance column on profiles).
-- One row per user per day. GRANTED = spendable, SPENT = consumed.
CREATE TABLE IF NOT EXISTS reward_tickets (
    user_uid            TEXT NOT NULL REFERENCES odyssey_user_profiles(uid) ON DELETE CASCADE,
    ticket_date         DATE NOT NULL,
    status              TEXT NOT NULL DEFAULT 'GRANTED' CHECK (status IN ('GRANTED', 'SPENT')),
    awarded_cosmetic_id TEXT REFERENCES cosmetic_items(id),
    created_at          TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at          TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    PRIMARY KEY (user_uid, ticket_date)
);

CREATE INDEX IF NOT EXISTS idx_reward_tickets_spendable ON reward_tickets (user_uid, ticket_date) WHERE status = 'GRANTED';

ALTER TABLE reward_tickets ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on reward_tickets" ON reward_tickets;
CREATE POLICY "Allow service_role full access on reward_tickets" ON reward_tickets FOR ALL TO service_role USING (true);

-- 3. Collection ownership: one row per owned cosmetic, tier bumps on duplicate (max 3).
CREATE TABLE IF NOT EXISTS user_cosmetics (
    user_uid    TEXT NOT NULL REFERENCES odyssey_user_profiles(uid) ON DELETE CASCADE,
    cosmetic_id TEXT NOT NULL REFERENCES cosmetic_items(id),
    tier        INTEGER NOT NULL DEFAULT 1 CHECK (tier BETWEEN 1 AND 3),
    equipped    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at  TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    PRIMARY KEY (user_uid, cosmetic_id)
);

ALTER TABLE user_cosmetics ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on user_cosmetics" ON user_cosmetics;
CREATE POLICY "Allow service_role full access on user_cosmetics" ON user_cosmetics FOR ALL TO service_role USING (true);

-- 4. Level bonus extension of the existing cap resolver.
-- effective = base + bonus(level); base 0 (unlimited) stays unlimited.
-- Config key 'level_cap_bonus' holds a JSON map {"5":300,"10":500}.
-- Missing/invalid config => bonus 0 (existing behavior preserved).
CREATE OR REPLACE FUNCTION odyssey_get_effective_earning_cap(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_cap INT; v_global INT; v_level INT; v_bonus_cfg TEXT; v_bonus INT := 0;
BEGIN
    SELECT COALESCE(NULLIF(trim(value),''), '')::INT INTO v_global FROM odyssey_system_config WHERE key='default_monthly_earning_cap';
    IF v_global IS NULL OR v_global < 0 THEN
        v_global := 3320;
    END IF;
    IF v_global > 10000 THEN v_global := 3320; END IF;
    SELECT monthly_earning_cap, COALESCE(level, 1) INTO v_cap, v_level FROM odyssey_user_profiles WHERE uid = p_user_uid;
    IF v_cap IS NULL THEN
        v_cap := v_global;
    END IF;
    IF v_cap IS NULL OR v_cap <= 0 THEN
        RETURN v_cap; -- 0 = unlimited (bonus must not cap the uncapped)
    END IF;
    BEGIN
        SELECT value INTO v_bonus_cfg FROM odyssey_system_config WHERE key='level_cap_bonus';
        SELECT (kv.value)::INT INTO v_bonus
        FROM jsonb_each_text(v_bonus_cfg::jsonb) AS kv(key, value)
        WHERE kv.key ~ '^[0-9]+$' AND kv.key::INT <= v_level
          AND kv.value ~ '^[0-9]+$' AND kv.value::INT >= 0
        ORDER BY kv.key::INT DESC
        LIMIT 1;
    EXCEPTION WHEN OTHERS THEN
        v_bonus := 0; -- malformed config must never break rewards
    END;
    IF v_bonus IS NULL OR v_bonus < 0 THEN v_bonus := 0; END IF;
    IF v_cap + v_bonus > 10000 THEN RETURN 10000; END IF;
    RETURN v_cap + v_bonus;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_get_effective_earning_cap(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_get_effective_earning_cap(TEXT) TO service_role;

-- 5. Claim daily ticket: server-side eligibility (APPROVED submission today,
-- Asia/Jakarta), idempotent via PK. Returns {granted, tickets}.
CREATE OR REPLACE FUNCTION odyssey_claim_daily_ticket(p_user_uid TEXT)
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_today DATE; v_active BOOLEAN; v_granted BOOLEAN := FALSE; v_balance INT;
BEGIN
    v_today := (timezone('Asia/Jakarta'::text, now()))::DATE;
    SELECT is_active INTO v_active FROM odyssey_user_profiles WHERE uid = p_user_uid;
    IF NOT FOUND OR v_active IS FALSE THEN
        RAISE EXCEPTION 'Akun tidak aktif' USING ERRCODE = 'P0021';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM odyssey_task_submissions s
        JOIN odyssey_tasks t ON t.id = s.task_id
        WHERE s.user_uid = p_user_uid AND s.status = 'APPROVED'
          AND (timezone('Asia/Jakarta'::text, COALESCE(s.reviewed_at, s.created_at)))::DATE = v_today
    ) THEN
        RAISE EXCEPTION 'Selesaikan minimal 1 tugas hari ini untuk mendapatkan Tiket Hadiah' USING ERRCODE = 'P0023';
    END IF;
    INSERT INTO reward_tickets (user_uid, ticket_date, status)
    VALUES (p_user_uid, v_today, 'GRANTED')
    ON CONFLICT (user_uid, ticket_date) DO NOTHING
    RETURNING TRUE INTO v_granted;
    IF v_granted IS NULL THEN v_granted := FALSE; END IF;
    SELECT COUNT(*)::INT INTO v_balance FROM reward_tickets WHERE user_uid = p_user_uid AND status = 'GRANTED';
    RETURN jsonb_build_object('granted', v_granted, 'tickets', v_balance);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_claim_daily_ticket(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_claim_daily_ticket(TEXT) TO service_role;

-- 6. Open reward box: atomically consume oldest GRANTED ticket, persist result
-- BEFORE returning (retry-safe). Duplicate cosmetic bumps tier (max 3).
CREATE OR REPLACE FUNCTION odyssey_open_reward_box(p_user_uid TEXT)
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_ticket_date DATE; v_cosmetic RECORD; v_tier INT; v_is_new BOOLEAN;
BEGIN
    SELECT ticket_date INTO v_ticket_date FROM reward_tickets
    WHERE user_uid = p_user_uid AND status = 'GRANTED'
    ORDER BY ticket_date ASC LIMIT 1 FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Tidak ada Tiket Hadiah. Selesaikan tugas hari ini untuk mendapatkannya.' USING ERRCODE = 'P0024';
    END IF;
    SELECT * INTO v_cosmetic FROM cosmetic_items WHERE is_active ORDER BY random() LIMIT 1;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Hadiah belum tersedia' USING ERRCODE = 'P0025';
    END IF;
    INSERT INTO user_cosmetics (user_uid, cosmetic_id, tier, equipped)
    VALUES (p_user_uid, v_cosmetic.id, 1, FALSE)
    ON CONFLICT (user_uid, cosmetic_id)
    DO UPDATE SET tier = LEAST(user_cosmetics.tier + 1, 3), updated_at = timezone('utc'::text, now())
    RETURNING tier, (xmax = 0) INTO v_tier, v_is_new;
    UPDATE reward_tickets SET status = 'SPENT', awarded_cosmetic_id = v_cosmetic.id,
        updated_at = timezone('utc'::text, now())
    WHERE user_uid = p_user_uid AND ticket_date = v_ticket_date;
    RETURN jsonb_build_object(
        'cosmetic_id', v_cosmetic.id, 'slot', v_cosmetic.slot, 'asset', v_cosmetic.asset,
        'tier', v_tier, 'is_new', COALESCE(v_is_new, TRUE));
END;
$$;
REVOKE ALL ON FUNCTION odyssey_open_reward_box(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_open_reward_box(TEXT) TO service_role;

-- Register migration version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '078_progression_rewards')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
