-- ============================================================
-- Migration 104: Atomic Admin Income Configuration and Audit Log
--
-- 1. trg_odyssey_cap_audit_logs_immutable:
--    - Enforces append-only semantics on odyssey_cap_audit_logs (disallows UPDATE/DELETE).
--
-- 2. Authoritative odyssey_admin_update_member_income_config RPC:
--    - Single database transaction wrapping member cap/status update and mandatory audit log.
--    - Serializes via target profile SELECT FOR UPDATE.
--    - Validates target existence and MEMBER role.
--    - Validates admin actor profile and role ('ADMIN', 'GUIDE').
--    - Enforces strict family/tenant isolation between admin and member.
--    - Enforces income status rules:
--        * CONFIGURED: monthly_earning_cap > 0, clamped to system ceiling (default 10,000).
--        * NOT_CONFIGURED / DISABLED: monthly_earning_cap forced to 0 (zero earning budget).
--    - No-op detection: if neither status nor cap changed, skips mutations and creates no audit log.
--    - Fail-closed: any failure during profile update or audit log insert rolls back the entire transaction.
--    - Restricted execution: REVOKE from PUBLIC/anon/authenticated; GRANT only to service_role.
-- ============================================================

-- 1. Immutable Audit Log Trigger
CREATE OR REPLACE FUNCTION odyssey_prevent_cap_audit_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'odyssey_cap_audit_logs adalah immutable audit log; UPDATE dan DELETE dilarang' USING ERRCODE = 'P0012';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_odyssey_cap_audit_logs_immutable ON odyssey_cap_audit_logs;
CREATE TRIGGER trg_odyssey_cap_audit_logs_immutable
BEFORE UPDATE OR DELETE ON odyssey_cap_audit_logs
FOR EACH ROW EXECUTE FUNCTION odyssey_prevent_cap_audit_mutation();

-- 2. Audit Table Privilege Hardening (Append-Only Enforcement)
REVOKE ALL ON odyssey_cap_audit_logs FROM PUBLIC, anon, authenticated;
REVOKE UPDATE, DELETE ON odyssey_cap_audit_logs FROM service_role;
GRANT SELECT, INSERT ON odyssey_cap_audit_logs TO service_role;

-- 3. Single-Transaction Atomic Income Configuration RPC
CREATE OR REPLACE FUNCTION odyssey_admin_update_member_income_config(
    p_target_uid TEXT,
    p_admin_uid TEXT,
    p_income_status TEXT DEFAULT NULL,
    p_monthly_cap INT DEFAULT NULL,
    p_reason TEXT DEFAULT NULL
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_admin RECORD;
    v_member RECORD;
    v_old_status odyssey_income_status;
    v_old_cap INT;
    v_new_status odyssey_income_status;
    v_new_cap INT;
    v_status_str TEXT;
    v_requested_cap INT;
    v_ceiling INT := 10000;
    v_changed BOOLEAN := false;
    v_audit_id UUID;
    v_final_reason TEXT;
BEGIN
    -- Input sanitization
    IF p_target_uid IS NULL OR trim(p_target_uid) = '' THEN
        RAISE EXCEPTION 'target_uid tidak boleh kosong' USING ERRCODE = 'P0005';
    END IF;
    IF p_admin_uid IS NULL OR trim(p_admin_uid) = '' THEN
        RAISE EXCEPTION 'admin_uid tidak boleh kosong' USING ERRCODE = 'P0005';
    END IF;

    -- 1. Verify Admin Actor Profile & Role
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid = p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN', 'GUIDE') THEN
        RAISE EXCEPTION 'Hanya admin keluarga yang dapat mengubah konfigurasi budget' USING ERRCODE = 'P0003';
    END IF;

    -- 2. Lock Target Member Profile
    SELECT * INTO v_member FROM odyssey_user_profiles WHERE uid = p_target_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Anggota tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;
    IF v_member.role != 'MEMBER' THEN
        RAISE EXCEPTION 'Target bukan merupakan MEMBER' USING ERRCODE = 'P0005';
    END IF;

    -- 3. Enforce Family / Tenant Isolation (Strict Fail-Closed)
    IF v_admin.family_id IS NULL OR v_member.family_id IS NULL OR v_admin.family_id != v_member.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: Target bukan milik anggota keluarga Anda' USING ERRCODE = 'P0003';
    END IF;

    -- 4. Resolve Target New Income Status
    IF p_income_status IS NOT NULL AND trim(p_income_status) != '' THEN
        v_status_str := upper(trim(p_income_status));
        IF v_status_str NOT IN ('NOT_CONFIGURED', 'CONFIGURED', 'DISABLED') THEN
            RAISE EXCEPTION 'income_status tidak valid (NOT_CONFIGURED|CONFIGURED|DISABLED)' USING ERRCODE = 'P0005';
        END IF;
        v_new_status := v_status_str::odyssey_income_status;
    ELSE
        v_new_status := COALESCE(v_member.income_status, 'NOT_CONFIGURED'::odyssey_income_status);
    END IF;

    -- 5. Resolve Target New Monthly Earning Cap
    IF v_new_status = 'CONFIGURED' THEN
        IF p_monthly_cap IS NOT NULL THEN
            v_requested_cap := p_monthly_cap;
        ELSE
            v_requested_cap := v_member.monthly_earning_cap;
        END IF;

        IF v_requested_cap IS NULL OR v_requested_cap <= 0 THEN
            RAISE EXCEPTION 'monthly_earning_cap harus lebih besar dari 0 untuk status CONFIGURED' USING ERRCODE = 'P0005';
        END IF;

        -- Enforce system ceiling
        SELECT COALESCE(NULLIF(trim(value), '')::INT, 10000) INTO v_ceiling
        FROM odyssey_system_config
        WHERE key = 'max_monthly_earning_cap_ceiling';

        IF v_ceiling IS NULL OR v_ceiling <= 0 THEN
            v_ceiling := 10000;
        END IF;

        IF v_requested_cap > v_ceiling THEN
            v_new_cap := v_ceiling;
        ELSE
            v_new_cap := v_requested_cap;
        END IF;
    ELSE
        -- NOT_CONFIGURED or DISABLED: cap is strictly 0
        v_new_cap := 0;
    END IF;

    -- 6. Capture Old State & Detect No-Op
    v_old_status := COALESCE(v_member.income_status, 'NOT_CONFIGURED'::odyssey_income_status);
    v_old_cap := v_member.monthly_earning_cap;

    v_changed := (v_new_status != v_old_status) OR (v_old_cap IS DISTINCT FROM v_new_cap);

    IF NOT v_changed THEN
        RETURN jsonb_build_object(
            'success', true,
            'changed', false,
            'target_uid', p_target_uid,
            'income_status', v_old_status,
            'monthly_earning_cap', v_old_cap,
            'audit_id', NULL
        );
    END IF;

    -- 7. Atomic Mutations (Profile Update + Mandatory Audit Insert)
    UPDATE odyssey_user_profiles
    SET income_status = v_new_status,
        monthly_earning_cap = v_new_cap,
        updated_at = timezone('utc'::text, now())
    WHERE uid = p_target_uid;

    v_final_reason := COALESCE(NULLIF(trim(p_reason), ''), 'Perubahan konfigurasi budget bulanan oleh admin');

    INSERT INTO odyssey_cap_audit_logs (
        user_uid,
        changed_by,
        old_income_status,
        new_income_status,
        old_cap,
        new_cap,
        reason,
        created_at
    ) VALUES (
        p_target_uid,
        p_admin_uid,
        v_old_status,
        v_new_status,
        v_old_cap,
        v_new_cap,
        v_final_reason,
        timezone('utc'::text, now())
    ) RETURNING id INTO v_audit_id;

    RETURN jsonb_build_object(
        'success', true,
        'changed', true,
        'target_uid', p_target_uid,
        'old_income_status', v_old_status,
        'new_income_status', v_new_status,
        'old_cap', v_old_cap,
        'new_cap', v_new_cap,
        'audit_id', v_audit_id
    );
END;
$$;

-- 4. Restrict Execution to service_role Only
REVOKE ALL ON FUNCTION odyssey_admin_update_member_income_config(TEXT, TEXT, TEXT, INT, TEXT) FROM PUBLIC;
REVOKE ALL ON FUNCTION odyssey_admin_update_member_income_config(TEXT, TEXT, TEXT, INT, TEXT) FROM anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_admin_update_member_income_config(TEXT, TEXT, TEXT, INT, TEXT) TO service_role;

-- 5. Record schema version
INSERT INTO odyssey_schema_version(key, value)
VALUES ('schema_version', '104_atomic_admin_income_configuration')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
