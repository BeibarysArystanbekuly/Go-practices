INSERT INTO users (email, password_hash, role)
VALUES (
  'admin@example.com',
  '$2a$10$1FZFnKbLgn02k8/vQ5RqK.1D4fl2dSZ4fpV5GYVAiLhJxXYBf8LWe',
  'admin'
)
ON CONFLICT (email) DO NOTHING;
