-- Chat session audit tables for API key scoped conversation browsing.

CREATE TABLE IF NOT EXISTS chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(64),
    user_id BIGINT NOT NULL,
    api_key_id BIGINT NOT NULL,
    account_id BIGINT,
    group_id BIGINT,
    platform VARCHAR(32) NOT NULL DEFAULT '',
    model VARCHAR(100) NOT NULL DEFAULT '',
    requested_model VARCHAR(100),
    upstream_model VARCHAR(100),
    inbound_endpoint VARCHAR(256),
    upstream_endpoint VARCHAR(256),
    request_type SMALLINT NOT NULL DEFAULT 0,
    stream BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(16) NOT NULL DEFAULT 'completed',
    http_status_code INT NOT NULL DEFAULT 200,
    user_preview VARCHAR(200),
    assistant_preview VARCHAR(200),
    message_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_request_id
    ON chat_sessions (request_id)
    WHERE request_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_chat_sessions_api_key_time
    ON chat_sessions (api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_time
    ON chat_sessions (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    seq INT NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'assistant',
    direction VARCHAR(16) NOT NULL DEFAULT 'outbound',
    content_text TEXT NOT NULL DEFAULT '',
    content_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_session_seq
    ON chat_messages (session_id, seq ASC, id ASC);
