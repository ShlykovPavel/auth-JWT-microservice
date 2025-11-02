INSERT INTO users (id, first_name, last_name, email, password, role, phone, created_at, updated_at)
VALUES
(1, 'John', 'Doe', 'john@example.com', '$2a$10$examplehash', 'user', '1234567890', NOW(), NOW()),
(2, 'Jane', 'Smith', 'jane@example.com', '$2a$10$examplehash', 'user', '0987654321', NOW(), NOW()),
(3, 'Alice', 'Johnson', 'alice@example.com', '$2a$10$examplehash', 'user', '1122334455', NOW(), NOW()),
(4, 'Bob', 'Brown', 'bob@example.com', '$2a$10$examplehash', 'user', '5566778899', NOW(), NOW()),
(5, 'Charlie', 'Davis', 'charlie@example.com', '$2a$10$examplehash', 'user', '6677889900', NOW(), NOW());

INSERT INTO users_outbox (user_id, send_to_kafka, event_type, attempt_count, last_attempt_at, created_at, updated_at)
VALUES (1, 'pending', 'USER_CREATED', 0, NULL, NOW(), NOW()),
       (2, 'pending', 'USER_UPDATED', 0, NULL, NOW(), NOW()),
       (3, 'failed', 'USER_DELETED', 3, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '2 hours',
        NOW() - INTERVAL '1 hour'),
       (4, 'sent', 'USER_CREATED', 1, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '1 hour',
        NOW() - INTERVAL '30 minutes'),
       (5, 'pending', 'USER_UPDATED', 0, NULL, NOW(), NOW());