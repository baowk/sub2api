ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS session_key VARCHAR(128);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_session_key
    ON chat_sessions (api_key_id, session_key)
    WHERE session_key IS NOT NULL;
