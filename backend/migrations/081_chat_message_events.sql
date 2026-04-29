CREATE TABLE IF NOT EXISTS chat_message_events (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    seq INT NOT NULL,
    kind VARCHAR(32) NOT NULL DEFAULT 'aux',
    role VARCHAR(16) NOT NULL DEFAULT 'system',
    direction VARCHAR(16) NOT NULL DEFAULT 'inbound',
    content_text TEXT NOT NULL DEFAULT '',
    content_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_message_events_session_seq
    ON chat_message_events (session_id, seq ASC, id ASC);
