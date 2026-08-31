-- ============================================================
-- Migration 054: Server-Authoritative Device Binding (1 Account = 1 Locked Device)
-- ============================================================

-- 1. Ensure device_bound_at column exists on odyssey_user_profiles
ALTER TABLE odyssey_user_profiles ADD COLUMN IF NOT EXISTS device_bound_at TIMESTAMPTZ;

-- 2. Atomic Bind or Verify Device RPC
CREATE OR REPLACE FUNCTION odyssey_bind_or_verify_device(
    p_user_uid TEXT,
    p_device_id TEXT
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_current_device TEXT;
    v_is_active BOOLEAN;
BEGIN
    IF p_user_uid IS NULL OR trim(p_user_uid) = '' THEN
        RAISE EXCEPTION 'User UID required' USING ERRCODE = 'P0019';
    END IF;

    IF p_device_id IS NULL OR trim(p_device_id) = '' THEN
        RAISE EXCEPTION 'Device ID required' USING ERRCODE = 'P0020';
    END IF;

    -- Lock profile row for update to prevent concurrent first-login binding race
    SELECT device_id, is_active INTO v_current_device, v_is_active
    FROM odyssey_user_profiles
    WHERE uid = p_user_uid FOR UPDATE;

    IF v_is_active IS FALSE THEN
        RAISE EXCEPTION 'Akun Anda nonaktif, silakan hubungi admin' USING ERRCODE = 'P0021';
    END IF;

    IF v_current_device IS NULL OR trim(v_current_device) = '' THEN
        -- Atomic binding on first login
        UPDATE odyssey_user_profiles
        SET device_id = p_device_id,
            device_bound_at = timezone('utc'::text, now()),
            updated_at = timezone('utc'::text, now())
        WHERE uid = p_user_uid AND (device_id IS NULL OR trim(device_id) = '');

        RETURN jsonb_build_object(
            'status', 'bound',
            'bound_device_id', p_device_id,
            'is_newly_bound', true
        );
    ELSIF v_current_device = p_device_id THEN
        -- Same device login allowed
        RETURN jsonb_build_object(
            'status', 'matched',
            'bound_device_id', v_current_device,
            'is_newly_bound', false
        );
    ELSE
        -- Different device login attempt rejected
        RAISE EXCEPTION 'Akun sudah terhubung ke perangkat lain. Silakan gunakan perangkat yang sudah terdaftar atau hubungi admin.' USING ERRCODE = 'P0022';
    END IF;
END;
$$;

REVOKE ALL ON FUNCTION odyssey_bind_or_verify_device(TEXT, TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_bind_or_verify_device(TEXT, TEXT) TO service_role;

-- 3. Admin Reset Device Binding RPC
CREATE OR REPLACE FUNCTION odyssey_admin_reset_device(
    p_target_uid TEXT
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
    UPDATE odyssey_user_profiles
    SET device_id = NULL,
        device_bound_at = NULL,
        updated_at = timezone('utc'::text, now())
    WHERE uid = p_target_uid;

    RETURN jsonb_build_object(
        'status', 'reset_success',
        'target_uid', p_target_uid
    );
END;
$$;

REVOKE ALL ON FUNCTION odyssey_admin_reset_device(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_admin_reset_device(TEXT) TO service_role;

-- Register migration version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '054_device_binding')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
