ALTER TABLE users
    -- password lockout
    ADD COLUMN failed_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN locked_until TIMESTAMP NULL,
    ADD COLUMN lockout_level INT NOT NULL DEFAULT 0,
    ADD COLUMN last_lockout_at TIMESTAMP NULL,

    -- totp lockout
    ADD COLUMN totp_failed_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN totp_locked_until TIMESTAMP NULL,
    ADD COLUMN totp_lockout_level INT NOT NULL DEFAULT 0,
    ADD COLUMN totp_last_lockout_at TIMESTAMP NULL;