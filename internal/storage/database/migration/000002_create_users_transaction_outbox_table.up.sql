CREATE TABLE users_outbox (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    send_to_kafka bool default false,
    payload JSONB NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    attempt_count INT DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);